package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/woohalabs2/yangobot/internal/lopec"
	"github.com/woohalabs2/yangobot/internal/lostark"
	"github.com/woohalabs2/yangobot/internal/ratelimit"
)

// APIHandler는 카카오 오픈채팅 봇(메신저봇R)을 위한 JSON REST API 핸들러입니다.
//
// 엔드포인트:
//
//	GET /api/v1/character/{name}  — 캐릭터 기본 정보
//	GET /api/v1/armory/{name}     — 군장 정보 (각인·카드·보석·아크그리드 포함, lopec 병합)
//	GET /api/v1/lopec/{name}      — 로펙 스펙 점수
//
// 모든 응답은 application/json; charset=utf-8 을 반환합니다.
// {"text": "..."} 형식으로 반환하여 메신저봇R의 Utils.getWebText() HTML 래핑 후에도 줄바꿈이 유지됩니다.
type APIHandler struct {
	loa     *lostark.Client
	lopec   *lopec.Client
	limiter *ratelimit.Limiter
}

func NewAPIHandler(loa *lostark.Client, lopec *lopec.Client, limiter *ratelimit.Limiter) *APIHandler {
	return &APIHandler{loa: loa, lopec: lopec, limiter: limiter}
}

type apiResponse struct {
	Text string `json:"text"`
}

func writeAPIText(w http.ResponseWriter, text string) {
	json.NewEncoder(w).Encode(apiResponse{Text: text})
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Istio 환경에서 실제 클라이언트 IP로 rate limit
	ip := clientIP(r)
	if !h.limiter.Allow(ip) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// /api/v1/{resource}/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		http.Error(w, "usage: /api/v1/{character|armory|lopec}/{name}", http.StatusBadRequest)
		return
	}
	resource := parts[0]
	name := strings.TrimSpace(parts[1])

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch resource {
	case "character":
		h.handleCharacter(w, r, name)
	case "armory":
		h.handleArmory(w, r, name)
	case "lopec":
		h.handleLopec(w, r, name)
	default:
		http.Error(w, "unknown resource: "+resource+"\nusage: /api/v1/{character|armory|lopec}/{name}", http.StatusNotFound)
	}
}

func (h *APIHandler) handleCharacter(w http.ResponseWriter, r *http.Request, name string) {
	info, err := h.loa.GetCharacter(r.Context(), name)
	if err != nil {
		log.Printf("api/character error [%s]: %v", name, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeAPIText(w, info.Format())
}

func (h *APIHandler) handleArmory(w http.ResponseWriter, r *http.Request, name string) {
	gear, err := h.loa.GetArmory(r.Context(), name)
	if err != nil {
		log.Printf("api/armory error [%s]: %v", name, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// lopec에서 직업각인 + 스펙점수 병합 (실패해도 armory 응답은 정상 반환)
	if lopecData, err := h.lopec.GetSpecPoint(r.Context(), name); err == nil {
		gear.SecondClass = lopecData.SecondClass
		gear.LoaSpecPoint = lopecData.SpecPoint
	}

	writeAPIText(w, gear.Format())
}

func (h *APIHandler) handleLopec(w http.ResponseWriter, r *http.Request, name string) {
	data, err := h.lopec.GetSpecPoint(r.Context(), name)
	if err != nil {
		log.Printf("api/lopec error [%s]: %v", name, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeAPIText(w, data.Format(name))
}

// clientIP는 Istio/Envoy 프록시 뒤에서도 실제 클라이언트 IP를 추출합니다.
// X-Forwarded-For → X-Real-IP → RemoteAddr 순으로 시도합니다.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// "client, proxy1, proxy2" 형식에서 첫 번째 IP 사용
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
