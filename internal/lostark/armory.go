package lostark

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ─── 각인 약칭 ───────────────────────────────────────────────────
var engravingAbbrev = map[string]string{
	"원한": "원한", "돌격대장": "돌대", "질량 증가": "질증",
	"기습의 대가": "기습", "저주받은 인형": "저받", "에테르 포식자": "에포",
	"아드레날린": "아드", "구슬동자": "구슬", "강화 무기": "강무",
	"처단자": "처단", "예리한 둔기": "예둔", "결투가의 일격": "결일",
	"바리케이드": "바리", "슈퍼 차지": "슈차", "황제의 칙령": "황칙",
	"황후의 은총": "황은", "번개의 분노": "번분", "타격의 대가": "타대",
	"피스메이커": "피스", "뇌격대장": "뇌대", "전투 태세": "전태",
	"만개": "만개", "달의 소나타": "달소", "충격 단련": "충단",
	"포격 강화": "포강", "멈추지 않는 자": "멈자", "분쇄의 주먹": "분주",
	"진화의 유산": "진유", "폭발물 전문가": "폭전",
}

func shortenEngraving(name string) string {
	if s, ok := engravingAbbrev[name]; ok {
		return s
	}
	r := []rune(name)
	if len(r) > 3 {
		return string(r[:3])
	}
	return name
}

var combatStatTypes = map[string]bool{
	"치명": true, "특화": true, "제압": true,
	"신속": true, "인내": true, "숙련": true,
}

var statAbbrevMap = map[string]string{
	"치명": "치", "특화": "특", "제압": "제",
	"신속": "신", "인내": "인", "숙련": "숙",
}

var (
	stonePosRe  = regexp.MustCompile(`FONT COLOR='#FFFFAC'>([^<]+)</FONT>\].*?Lv\.(\d+)`)
	arkSlotRe   = regexp.MustCompile(`(질서|혼돈)의\s*(해|달|별)`)
)

// ─── Raw API 타입 ─────────────────────────────────────────────────

type rawArmoryResp struct {
	ArmoryProfile   rawArmoryProfile   `json:"ArmoryProfile"`
	ArmoryEngraving rawArmoryEngraving `json:"ArmoryEngraving"`
	ArmoryEquipment []rawEquipItem     `json:"ArmoryEquipment"`
	ArmoryCard      rawArmoryCard      `json:"ArmoryCard"`
	ArkPassive      rawArkPassive      `json:"ArkPassive"`
	ArmoryGem       rawArmoryGem       `json:"ArmoryGem"`
	ArkGrid         rawArkGrid         `json:"ArkGrid"`
}

type rawArkGrid struct {
	Slots []rawArkSlot `json:"Slots"`
}

type rawArkSlot struct {
	Name    string `json:"Name"`
	Point   int    `json:"Point"`
	Tooltip string `json:"Tooltip"`
}

type rawArmoryProfile struct {
	ServerName      string    `json:"ServerName"`
	CharacterLevel  int       `json:"CharacterLevel"`
	ItemAvgLevel    string    `json:"ItemAvgLevel"`
	ExpeditionLevel int       `json:"ExpeditionLevel"`
	GuildName       string    `json:"GuildName"`
	UsingSkillPoint int       `json:"UsingSkillPoint"`
	TotalSkillPoint int       `json:"TotalSkillPoint"`
	CombatPower     string    `json:"CombatPower"`
	Stats           []rawStat `json:"Stats"`
}

type rawStat struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type rawArmoryEngraving struct {
	ArkPassiveEffects []rawEngravingEffect `json:"ArkPassiveEffects"`
}

type rawEngravingEffect struct {
	Name  string `json:"Name"`
	Level int    `json:"Level"`
}

type rawEquipItem struct {
	Type    string `json:"Type"`
	Tooltip string `json:"Tooltip"`
}

type rawArmoryCard struct {
	Effects []rawCardEffect `json:"Effects"`
}

type rawCardEffect struct {
	Items []rawCardItem `json:"Items"`
}

type rawCardItem struct {
	Name string `json:"Name"`
}

type rawArkPassive struct {
	Points []rawArkPoint `json:"Points"`
}

type rawArkPoint struct {
	Name  string `json:"Name"`
	Value int    `json:"Value"`
}

type rawArmoryGem struct {
	Gems []rawGemItem `json:"Gems"`
}

type rawGemItem struct {
	Level int `json:"Level"`
}

// ─── 결과 타입 ────────────────────────────────────────────────────

type GearData struct {
	ServerName      string
	GuildName       string
	CharacterLevel  int
	ItemAvgLevel    string
	ExpeditionLevel int
	UsingSkillPoint int
	TotalSkillPoint int
	CombatPower     string
	TopStats        []string // e.g. ["특:1839", "신:560"]

	// 어빌리티 스톤: positive 레벨, 각 풀네임
	StoneLevels    []int
	StoneFullNames []string

	// 각인 (ArkPassive 기준)
	Engravings []engravingEntry

	TopCardEffect string
	ArkFull       bool
	ArkPoints     int

	GemAvg   float64
	GemCount int

	ArkGridLines []string // ["질서: 해 유(19) | 달 유(20) | 별 고(20)", "혼돈: ..."]

	// lopec에서 채워짐
	SecondClass  string
	LoaSpecPoint float64
}

type engravingEntry struct {
	FullName   string
	Level      int
	StoneLevel int // 0 = 스톤 기여 없음
}

// Format은 챗봇 응답 문자열을 생성합니다.
func (g *GearData) Format() string {
	var b strings.Builder

	// 💠 어빌리티 스톤
	if len(g.StoneLevels) > 0 {
		lvls := make([]string, len(g.StoneLevels))
		abbrs := make([]string, len(g.StoneLevels))
		for i, l := range g.StoneLevels {
			lvls[i] = strconv.Itoa(l)
			abbrs[i] = shortenEngraving(g.StoneFullNames[i])
		}
		b.WriteString(fmt.Sprintf("💠 %s돌 %s\n", strings.Join(lvls, ","), strings.Join(abbrs, "/")))
	}

	// 각인
	var engParts []string
	for _, e := range g.Engravings {
		short := shortenEngraving(e.FullName)
		if e.StoneLevel > 0 {
			engParts = append(engParts, fmt.Sprintf("%s(%d,%d)", short, e.Level, e.StoneLevel))
		} else {
			engParts = append(engParts, fmt.Sprintf("%s(%d)", short, e.Level))
		}
	}
	if len(engParts) > 0 {
		b.WriteString(strings.Join(engParts, " ") + "\n")
	}

	// 카드
	if g.TopCardEffect != "" {
		b.WriteString(g.TopCardEffect + "\n")
	}

	b.WriteString("\n")

	// 템/전/원
	b.WriteString(fmt.Sprintf("템/전/원\t%s/%d/%d\n", g.ItemAvgLevel, g.CharacterLevel, g.ExpeditionLevel))

	// 아크패시브
	if g.ArkFull {
		b.WriteString("\n아크패시브\tfull\n")
	} else {
		b.WriteString("\n아크패시브\tloss\n")
	}

	// 직업각인
	if g.SecondClass != "" {
		b.WriteString(fmt.Sprintf("\n직업각인\t%s\n", g.SecondClass))
	}

	// 전투력/로펙
	combatPower := g.CombatPower
	if combatPower == "" {
		combatPower = "-"
	}
	if g.LoaSpecPoint > 0 {
		b.WriteString(fmt.Sprintf("\n전투력/로펙\t%s/%.2f\n", combatPower, g.LoaSpecPoint))
	}

	// 서버/길드
	guild := g.GuildName
	if guild == "" {
		guild = "-"
	}
	b.WriteString(fmt.Sprintf("\n서버/길드\t%s/%s\n", g.ServerName, guild))

	// 전투특성
	if len(g.TopStats) > 0 {
		b.WriteString("\n전투특성\t" + strings.Join(g.TopStats, " ") + "\n")
	}

	// 스킬포인트
	b.WriteString(fmt.Sprintf("\n스킬포인트\t%d/%d\n", g.UsingSkillPoint, g.TotalSkillPoint))

	// 보석평균
	if g.GemCount > 0 {
		b.WriteString(fmt.Sprintf("\n보석평균\t%.1f (총 %d개)\n", g.GemAvg, g.GemCount))
	}

	// 아크그리드
	if len(g.ArkGridLines) > 0 {
		b.WriteString("\n")
		for _, line := range g.ArkGridLines {
			b.WriteString(line + "\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

// ─── API 호출 ─────────────────────────────────────────────────────

const armoryCacheTTL = 5 * time.Minute

func (c *Client) GetArmory(ctx context.Context, name string) (*GearData, error) {
	cacheKey := "armory:" + name

	var cached GearData
	if err := c.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	endpoint := fmt.Sprintf("%s/armories/characters/%s", baseURL, url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var raw rawArmoryResp
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	gear := buildGearData(&raw)
	_ = c.cache.Set(ctx, cacheKey, gear, armoryCacheTTL)
	return gear, nil
}

func buildGearData(raw *rawArmoryResp) *GearData {
	g := &GearData{}
	p := raw.ArmoryProfile

	// 프로필 기본
	g.ServerName = p.ServerName
	g.GuildName = p.GuildName
	g.CharacterLevel = p.CharacterLevel
	g.ItemAvgLevel = p.ItemAvgLevel
	g.ExpeditionLevel = p.ExpeditionLevel
	g.UsingSkillPoint = p.UsingSkillPoint
	g.TotalSkillPoint = p.TotalSkillPoint
	g.CombatPower = p.CombatPower

	// 전투특성 상위 2개
	type statVal struct {
		typ string
		val int
	}
	var cs []statVal
	for _, s := range p.Stats {
		if !combatStatTypes[s.Type] {
			continue
		}
		v, _ := strconv.Atoi(s.Value)
		cs = append(cs, statVal{s.Type, v})
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i].val > cs[j].val })
	for i := 0; i < 2 && i < len(cs); i++ {
		abbr := statAbbrevMap[cs[i].typ]
		if abbr == "" {
			abbr = cs[i].typ
		}
		g.TopStats = append(g.TopStats, fmt.Sprintf("%s:%d", abbr, cs[i].val))
	}

	// 어빌리티 스톤 파싱
	stoneMap := parseStoneEquip(raw.ArmoryEquipment, g)

	// 각인 (스톤 cross-ref)
	for _, e := range raw.ArmoryEngraving.ArkPassiveEffects {
		entry := engravingEntry{FullName: e.Name, Level: e.Level}
		if sl, ok := stoneMap[e.Name]; ok {
			entry.StoneLevel = sl
		}
		g.Engravings = append(g.Engravings, entry)
	}

	// 카드 최고 효과
	for _, eff := range raw.ArmoryCard.Effects {
		for _, item := range eff.Items {
			if item.Name != "" {
				g.TopCardEffect = item.Name
			}
		}
	}

	// 아크패시브
	total := 0
	for _, pt := range raw.ArkPassive.Points {
		total += pt.Value
	}
	g.ArkPoints = total
	g.ArkFull = total >= 300

	// 보석 평균
	var gemSum int
	for _, gem := range raw.ArmoryGem.Gems {
		gemSum += gem.Level
		g.GemCount++
	}
	if g.GemCount > 0 {
		g.GemAvg = float64(gemSum) / float64(g.GemCount)
	}

	// 아크그리드
	g.ArkGridLines = parseArkGrid(raw.ArkGrid.Slots)

	return g
}

func arkGradeAbbrev(tooltip string) string {
	switch {
	case strings.Contains(tooltip, "고대 아크"):
		return "고"
	case strings.Contains(tooltip, "전설 아크"):
		return "전"
	case strings.Contains(tooltip, "유물 아크"):
		return "유"
	default:
		return "?"
	}
}

// parseArkGrid: 슬롯 목록에서 질서/혼돈 × 해/달/별 파싱 후 포맷 라인 반환
func parseArkGrid(slots []rawArkSlot) []string {
	// type → slot → (grade, point)
	type slotInfo struct {
		grade string
		point int
	}
	grid := map[string]map[string]slotInfo{
		"질서": {},
		"혼돈": {},
	}
	slotOrder := []string{"해", "달", "별"}

	for _, s := range slots {
		m := arkSlotRe.FindStringSubmatch(s.Name)
		if m == nil {
			continue
		}
		typ, slot := m[1], m[2]
		grid[typ][slot] = slotInfo{
			grade: arkGradeAbbrev(s.Tooltip),
			point: s.Point,
		}
	}

	var lines []string
	for _, typ := range []string{"질서", "혼돈"} {
		if len(grid[typ]) == 0 {
			continue
		}
		var parts []string
		for _, slot := range slotOrder {
			if info, ok := grid[typ][slot]; ok {
				parts = append(parts, fmt.Sprintf("%s %s(%d)", slot, info.grade, info.point))
			}
		}
		if len(parts) > 0 {
			lines = append(lines, fmt.Sprintf("%s: %s", typ, strings.Join(parts, " | ")))
		}
	}
	return lines
}

// parseStoneEquip: 어빌리티 스톤에서 positive 각인 파싱, stoneMap(풀네임→레벨) 반환
func parseStoneEquip(equips []rawEquipItem, g *GearData) map[string]int {
	stoneMap := make(map[string]int)
	for _, e := range equips {
		if e.Type != "어빌리티 스톤" {
			continue
		}
		type stoneLine struct {
			name  string
			level int
		}
		var positives []stoneLine

		matches := stonePosRe.FindAllStringSubmatch(e.Tooltip, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			name := m[1]
			level, _ := strconv.Atoi(m[2])
			if level > 0 {
				positives = append(positives, stoneLine{name, level})
				stoneMap[name] = level
			}
		}
		// 레벨 내림차순 정렬
		sort.Slice(positives, func(i, j int) bool {
			return positives[i].level > positives[j].level
		})
		for _, sl := range positives {
			g.StoneLevels = append(g.StoneLevels, sl.level)
			g.StoneFullNames = append(g.StoneFullNames, sl.name)
		}
		break
	}
	return stoneMap
}
