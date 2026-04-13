package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type CmdType string

const (
	CmdCharacter   CmdType = "character"
	CmdSpec        CmdType = "spec"
	CmdGear        CmdType = "gear"
	CmdExpedition  CmdType = "expedition"
	CmdDistribute  CmdType = "distribute"
)

type Command struct {
	Type CmdType
	Args []string
	N    int   // CmdDistribute: 인원수 (4 or 8)
	Gold int64 // CmdDistribute: 입찰 금액
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
	case ".ㄱㅁ8", ".ㄱㅁ4":
		n := 8
		if parts[0] == ".ㄱㅁ4" {
			n = 4
		}
		if len(parts) < 2 {
			return nil, fmt.Errorf("금액 또는 각인서 이름을 입력해주세요. 예) %s 49000 또는 %s 아드레날린", parts[0], parts[0])
		}
		raw := strings.ReplaceAll(parts[1], ",", "")
		gold, err := strconv.ParseInt(raw, 10, 64)
		if err == nil && gold > 0 {
			// 숫자 → 직접 금액 입력
			return &Command{Type: CmdDistribute, N: n, Gold: gold}, nil
		}
		// 텍스트 → 각인서 이름 조회
		name := strings.Join(parts[1:], " ")
		return &Command{Type: CmdDistribute, N: n, Args: []string{name}}, nil
	default:
		return nil, ErrUnknownCommand
	}
}
