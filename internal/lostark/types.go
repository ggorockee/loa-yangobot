package lostark

import "fmt"

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
