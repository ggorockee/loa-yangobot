package lostark

import (
	"fmt"
	"sort"
	"strings"
)

// ─── 직업 약칭 ──────────────────────────────────────────────────────

var classAbbrev = map[string]string{
	"디스트로이어": "디트",
	"워로드":      "워로드",
	"버서커":      "버서커",
	"홀리나이트":   "홀나",
	"슬레이어":    "슬레",
	"발키리":      "발키리",
	"배틀마스터":   "배마",
	"인파이터":    "인파",
	"기공사":      "기공사",
	"창술사":      "창술사",
	"스트라이커":   "스커",
	"브레이커":     "브커",
	"데빌헌터":     "데헌",
	"블래스터":     "블래",
	"호크아이":     "호크",
	"스카우터":     "스카",
	"건슬링어":     "건슬",
	"바드":        "바드",
	"서머너":       "서머너",
	"아르카나":     "알카",
	"소서리스":    "소서",
	"블레이드":     "블레",
	"데모닉":       "데모닉",
	"리퍼":         "리퍼",
	"소울이터":      "소울",
	"도화가":        "도화가",
	"기상술사":      "기상",
	"환수사":        "환수사",
	"가디언나이트":   "가나",
}

// padClass는 약칭을 전각 3자 폭으로 맞춥니다.
// 2글자이면 중간에 전각 공백(U+3000)을 삽입합니다.
func padClass(abbrev string) string {
	runes := []rune(abbrev)
	if len(runes) == 2 {
		return string(runes[0]) + "　" + string(runes[1])
	}
	return abbrev
}

func classLabel(className string) string {
	abbrev, ok := classAbbrev[className]
	if !ok {
		r := []rune(className)
		if len(r) >= 3 {
			abbrev = string(r[:3])
		} else {
			abbrev = className
		}
	}
	return padClass(abbrev)
}

// ─── 레이드 골드 테이블 ─────────────────────────────────────────────

type raidEntry struct {
	group     string
	label     string
	minLevel  float64
	clearGold int64
	moreGold  int64
}

// raidGroups는 레이드 그룹 순서대로, 각 그룹 내 티어는 높은 컷라인 순으로 정렬합니다.
// 한 그룹에서 캐릭터가 입장 가능한 최상위 티어 하나만 골드 후보로 취합니다.
var altRaidGroups = []struct {
	name    string
	entries []raidEntry
}{
	{"세르카", []raidEntry{
		{"세르카", "나메", 1740, 54000, 17280},
		{"세르카", "하드", 1730, 44000, 14080},
		{"세르카", "노말", 1710, 35000, 11200},
	}},
	{"종막", []raidEntry{
		{"종막", "하드", 1730, 52000, 16640},
		{"종막", "노말", 1710, 40000, 12800},
	}},
	{"4막", []raidEntry{
		{"4막", "하드", 1720, 42000, 13440},
		{"4막", "노말", 1700, 33000, 10560},
	}},
	{"3막", []raidEntry{
		{"3막", "하드", 1700, 27000, 8350},
		{"3막", "노말", 1680, 21000, 7010},
	}},
	{"2막", []raidEntry{
		{"2막", "하드", 1690, 23000, 7500},
		{"2막", "노말", 1670, 16500, 5540},
	}},
	{"1막", []raidEntry{
		{"1막", "하드", 1680, 18000, 5970},
		{"1막", "노말", 1660, 11500, 2530},
	}},
	{"서막", []raidEntry{
		{"서막", "하드", 1640, 7200, 2350},
	}},
	{"베히모스", []raidEntry{
		{"베히모스", "", 1600, 7200, 2350},
	}},
}

// availableRaidsForChar는 캐릭터 아이템 레벨 기준으로 각 레이드 그룹에서
// 입장 가능한 최상위 티어를 모두 반환합니다. 제한 없이 전체 레이드 포함.
func availableRaidsForChar(itemLevel float64) []raidEntry {
	var available []raidEntry
	for _, group := range altRaidGroups {
		for _, e := range group.entries {
			if itemLevel >= e.minLevel {
				available = append(available, e)
				break
			}
		}
	}
	return available
}

// ─── 포맷 ──────────────────────────────────────────────────────────

type charGoldResult struct {
	char      CharacterInfo
	withMore  int64
	clearOnly int64
}

// FormatAlts는 원정대 부캐 골드 계산 결과를 포맷합니다.
// queriedName: 조회에 사용한 캐릭터 이름 (헤더에 표시)
func FormatAlts(queriedName string, siblings []CharacterInfo) string {
	// 서버별 그룹핑 (조회 결과 순서 유지)
	serverMap := make(map[string][]CharacterInfo)
	var serverOrder []string
	seen := make(map[string]bool)
	for _, s := range siblings {
		if !seen[s.ServerName] {
			serverOrder = append(serverOrder, s.ServerName)
			seen[s.ServerName] = true
		}
		serverMap[s.ServerName] = append(serverMap[s.ServerName], s)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s님의 부캐 정보\n", queriedName))

	for _, server := range serverOrder {
		chars := serverMap[server]

		// 아이템 레벨 내림차순 정렬
		sort.Slice(chars, func(i, j int) bool {
			return parseItemLevel(chars[i].ItemAvgLevel) > parseItemLevel(chars[j].ItemAvgLevel)
		})

		// 골드 계산
		// clearGold: 클리어 시 수령 골드 (gross)
		// netGold:   클리어 골드 - 더보기 비용 = 순이익 (더보기 제외)
		var results []charGoldResult
		var totalClear int64
		for _, ch := range chars {
			lvl := parseItemLevel(ch.ItemAvgLevel)
			raids := availableRaidsForChar(lvl)
			var clearGold, netGold int64
			for _, r := range raids {
				clearGold += r.clearGold
				netGold += r.clearGold - r.moreGold
			}
			results = append(results, charGoldResult{char: ch, withMore: clearGold, clearOnly: netGold})
			totalClear += clearGold
		}

		// 상위 6캐릭 합계
		var top6Clear, top6Net int64
		for i, cg := range results {
			if i >= 6 {
				break
			}
			top6Clear += cg.withMore
			top6Net += cg.clearOnly
		}

		b.WriteString(fmt.Sprintf("\n✤ %s 서버\n", server))
		for _, cg := range results {
			label := classLabel(cg.char.CharacterClassName)
			b.WriteString(fmt.Sprintf("[%s] %s (%s)\n", label, cg.char.CharacterName, cg.char.ItemAvgLevel))
		}
		b.WriteString(fmt.Sprintf("\n• 6캐릭 합계: %s 골드\n", formatGold(top6Clear)))
		b.WriteString(fmt.Sprintf("• 전체 합계: %s 골드\n", formatGold(totalClear)))
		b.WriteString(fmt.Sprintf("• 더보기 제외: %s 골드", formatGold(top6Net)))
	}

	return b.String()
}

func formatGold(g int64) string {
	if g == 0 {
		return "0"
	}
	s := fmt.Sprintf("%d", g)
	n := len(s)
	var result []byte
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, s[i])
	}
	return string(result)
}
