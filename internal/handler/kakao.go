package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/woohalabs2/yangobot/internal/command"
	"github.com/woohalabs2/yangobot/internal/lopec"
	"github.com/woohalabs2/yangobot/internal/lostark"
	"github.com/woohalabs2/yangobot/internal/ratelimit"
)

// KakaoRequest는 카카오 챗봇 서버로부터 전달되는 요청 구조체입니다.
type KakaoRequest struct {
	UserRequest struct {
		Utterance string `json:"utterance"`
		User      struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"userRequest"`
}

// KakaoResponse는 카카오 챗봇 서버에 반환하는 응답 구조체입니다.
type KakaoResponse struct {
	Version  string `json:"version"`
	Template struct {
		Outputs []map[string]any `json:"outputs"`
	} `json:"template"`
}

func simpleText(text string) KakaoResponse {
	resp := KakaoResponse{Version: "2.0"}
	resp.Template.Outputs = []map[string]any{
		{"simpleText": map[string]string{"text": text}},
	}
	return resp
}

type KakaoHandler struct {
	loa     *lostark.Client
	lopec   *lopec.Client
	limiter *ratelimit.Limiter
}

func NewKakaoHandler(loa *lostark.Client, lopec *lopec.Client, limiter *ratelimit.Limiter) *KakaoHandler {
	return &KakaoHandler{loa: loa, lopec: lopec, limiter: limiter}
}

func (h *KakaoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req KakaoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("decode error: %v", err)
		h.writeJSON(w, simpleText("요청을 처리할 수 없습니다."))
		return
	}

	userID := req.UserRequest.User.ID
	if !h.limiter.Allow(userID) {
		h.writeJSON(w, simpleText("현재 요청이 많아 처리가 어렵습니다.\n잠시 후 다시 이용해 주시면 감사하겠습니다."))
		return
	}

	cmd, err := command.Parse(req.UserRequest.Utterance)
	if err != nil {
		h.writeJSON(w, simpleText("알 수 없는 명령어입니다.\n사용법: /캐릭터 <닉네임>"))
		return
	}

	var result string
	switch cmd.Type {
	case command.CmdCharacter:
		info, err := h.loa.GetCharacter(r.Context(), cmd.Args[0])
		if err != nil {
			log.Printf("lostark API error: %v", err)
			h.writeJSON(w, simpleText("캐릭터 정보를 가져오지 못했습니다."))
			return
		}
		result = info.Format()
	case command.CmdSpec:
		data, err := h.lopec.GetSpecPoint(r.Context(), cmd.Args[0])
		if err != nil {
			log.Printf("lopec error: %v", err)
			h.writeJSON(w, simpleText("스펙 점수를 가져오지 못했습니다."))
			return
		}
		result = data.Format(cmd.Args[0])
	case command.CmdGear:
		name := cmd.Args[0]
		gear, err := h.loa.GetArmory(r.Context(), name)
		if err != nil {
			log.Printf("armory error: %v", err)
			h.writeJSON(w, simpleText("군장 정보를 가져오지 못했습니다."))
			return
		}
		// lopec에서 직업각인 + 스펙점수 병합
		if lopecData, err := h.lopec.GetSpecPoint(r.Context(), name); err == nil {
			gear.SecondClass = lopecData.SecondClass
			gear.LoaSpecPoint = lopecData.SpecPoint
		}
		result = gear.Format()
	default:
		result = "지원하지 않는 명령어입니다."
	}

	h.writeJSON(w, simpleText(result))
}

func (h *KakaoHandler) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode error: %v", err)
	}
}
