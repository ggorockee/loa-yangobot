package command

import (
	"errors"
	"strings"
)

type CmdType string

const (
	CmdCharacter   CmdType = "character"
	CmdSpec        CmdType = "spec"
	CmdGear        CmdType = "gear"
	CmdExpedition  CmdType = "expedition"
)

type Command struct {
	Type CmdType
	Args []string
}

var ErrUnknownCommand = errors.New("unknown command")

// Parse는 사용자 발화문을 파싱하여 Command를 반환합니다.
// 지원 형식:
//   - /캐릭터 <닉네임>
//   - /character <닉네임>
func Parse(utterance string) (*Command, error) {
	utterance = strings.TrimSpace(utterance)
	if !strings.HasPrefix(utterance, "/") && !strings.HasPrefix(utterance, ".") {
		return nil, ErrUnknownCommand
	}

	parts := strings.Fields(utterance)
	if len(parts) == 0 {
		return nil, ErrUnknownCommand
	}

	switch strings.ToLower(parts[0]) {
	case "/캐릭터", "/character":
		if len(parts) < 2 {
			return nil, errors.New("닉네임을 입력해주세요. 예) /캐릭터 아비투스")
		}
		return &Command{Type: CmdCharacter, Args: parts[1:]}, nil
	case "/스펙", "/spec":
		if len(parts) < 2 {
			return nil, errors.New("닉네임을 입력해주세요. 예) /스펙 아비투스")
		}
		return &Command{Type: CmdSpec, Args: parts[1:]}, nil
	case "/군장", ".군장":
		if len(parts) < 2 {
			return nil, errors.New("닉네임을 입력해주세요. 예) .군장 아비투스")
		}
		return &Command{Type: CmdGear, Args: parts[1:]}, nil
	case "/원정대", "/expedition":
		if len(parts) < 2 {
			return nil, errors.New("닉네임을 입력해주세요. 예) /원정대 아비투스")
		}
		return &Command{Type: CmdExpedition, Args: parts[1:]}, nil
	default:
		return nil, ErrUnknownCommand
	}
}
