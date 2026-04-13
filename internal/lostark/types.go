package lostark

import (
	"fmt"
	"strconv"
	"strings"
)

// CharacterInfo는 로스트아크 캐릭터 기본 정보입니다.
type CharacterInfo struct {
	ServerName        string  `json:"ServerName"`
	CharacterName     string  `json:"CharacterName"`
	CharacterLevel    int     `json:"CharacterLevel"`
	CharacterClassName string  `json:"CharacterClassName"`
	ItemAvgLevel      string  `json:"ItemAvgLevel"`
	ItemMaxLevel      string  `json:"ItemMaxLevel"`
}

// Format은 카카오 챗봇 응답용 문자열을 반환합니다.
func (c *CharacterInfo) Format() string {
	return fmt.Sprintf(
		"[%s] %s\n직업: %s\n레벨: %d\n아이템 평균: %s\n아이템 최고: %s",
		c.ServerName,
		c.CharacterName,
		c.CharacterClassName,
		c.CharacterLevel,
		c.ItemAvgLevel,
		c.ItemMaxLevel,
	)
}

// ArmoryProfile은 /armories/characters/{name}/profiles 응답 구조체입니다.
type ArmoryProfile struct {
	CharacterImage    string  `json:"CharacterImage"`
	ExpeditionLevel   int     `json:"ExpeditionLevel"`
	PvpGradeName      string  `json:"PvpGradeName"`
	TownLevel         int     `json:"TownLevel"`
	TownName          string  `json:"TownName"`
	Title             string  `json:"Title"`
	GuildMemberGrade  string  `json:"GuildMemberGrade"`
	GuildName         string  `json:"GuildName"`
	Stats             []Stat  `json:"Stats"`
	CharacterInfo
}

type Stat struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

// ─── 원정대 레이드 커트라인 ─────────────────────────────────────────

type tierDef struct {
	label    string
	minLevel float64
}

type raidDef struct {
	name  string
	tiers []tierDef // 높은 레벨 순으로 정렬
}

var raidCuts = []raidDef{
	{name: "세르카", tiers: []tierDef{
		{label: "나메", minLevel: 1740},
		{label: "하드", minLevel: 1730},
		{label: "노말", minLevel: 1710},
	}},
	{name: "종막", tiers: []tierDef{
		{label: "하드", minLevel: 1730},
		{label: "노말", minLevel: 1710},
	}},
	{name: "성당", tiers: []tierDef{
		{label: "3단계", minLevel: 1750},
		{label: "2단계", minLevel: 1720},
		{label: "1단계", minLevel: 1700},
	}},
}

// parseItemLevel은 "1,710.00" 형태의 아이템 레벨 문자열을 float64로 변환합니다.
func parseItemLevel(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// FormatExpeditionRaid는 원정대 캐릭터 목록을 받아 레이드 입장 가능 수를 포맷합니다.
// queriedName: 조회에 사용한 캐릭터 이름 (헤더에 표시)
func FormatExpeditionRaid(queriedName string, chars []CharacterInfo) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("❙ [%s] 원정대\n", queriedName))

	for _, raid := range raidCuts {
		// 티어별 카운트
		counts := make([]int, len(raid.tiers))
		for _, ch := range chars {
			lvl := parseItemLevel(ch.ItemAvgLevel)
			for i, t := range raid.tiers {
				if lvl >= t.minLevel {
					counts[i]++
				}
			}
		}

		// 티어 파트 조합 (예: "나메(2) 하드(1) 노말(3)")
		parts := make([]string, len(raid.tiers))
		for i, t := range raid.tiers {
			parts[i] = fmt.Sprintf("%s(%d)", t.label, counts[i])
		}

		b.WriteString(fmt.Sprintf("%-6s%s\n", raid.name, strings.Join(parts, " ")))
	}

	return strings.TrimRight(b.String(), "\n")
}
