package handler

import (
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/woohalabs2/yangobot/internal/distribute"
	"github.com/woohalabs2/yangobot/internal/lopec"
	"github.com/woohalabs2/yangobot/internal/lostark"
	"github.com/woohalabs2/yangobot/internal/ratelimit"
)

// HandleDistribute는 GET /api/v1/distribute/:n/:price 엔드포인트입니다.
func (h *APIHandler) HandleDistribute(c fiber.Ctx) error {
	ip := c.IP()
	if !h.limiter.Allow(ip) {
		return c.Status(fiber.StatusTooManyRequests).SendString("rate limit exceeded")
	}
	n, err1 := strconv.Atoi(c.Params("n"))
	price, err2 := strconv.ParseInt(strings.ReplaceAll(c.Params("price"), ",", ""), 10, 64)
	if err1 != nil || err2 != nil || (n != 4 && n != 8) || price <= 0 {
		return c.Status(fiber.StatusBadRequest).SendString("usage: /api/v1/distribute/{4|8}/{price}")
	}
	r := distribute.Result{N: n, Price: price}
	return c.JSON(apiResponse{Text: r.Format()})
}

// APIHandler는 카카오 오픈채팅 봇(메신저봇R)을 위한 JSON REST API 핸들러입니다.
//
// 엔드포인트:
//
//	GET /api/v1/character/{name}       — 캐릭터 기본 정보
//	GET /api/v1/armory/{name}          — 군장 정보 (각인·카드·보석·아크그리드 포함, lopec 병합)
//	GET /api/v1/lopec/{name}           — 로펙 스펙 점수
//	GET /api/v1/expedition/{name}      — 원정대 레이드 커트라인 카운트
//	GET /api/v1/distribute/{n}/{price} — 분배금 계산 (n=4 or 8)
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

func (h *APIHandler) Handle(c fiber.Ctx) error {
	ip := c.IP()
	if !h.limiter.Allow(ip) {
		return c.Status(fiber.StatusTooManyRequests).SendString("rate limit exceeded")
	}

	resource := c.Params("resource")
	decoded, _ := url.PathUnescape(c.Params("name"))
	name := strings.TrimSpace(decoded)
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("usage: /api/v1/{character|armory|lopec|expedition}/{name}")
	}

	ctx := c.Context()

	switch resource {
	case "character":
		info, err := h.loa.GetCharacter(ctx, name)
		if err != nil {
			log.Printf("api/character error [%s]: %v", name, err)
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}
		return c.JSON(apiResponse{Text: info.Format()})

	case "armory":
		var (
			gear      *lostark.GearData
			lopecData *lopec.SpecData
			gearErr   error
			wg        sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			gear, gearErr = h.loa.GetArmory(ctx, name)
		}()
		go func() {
			defer wg.Done()
			lopecData, _ = h.lopec.GetSpecPoint(ctx, name)
		}()
		wg.Wait()

		if gearErr != nil {
			log.Printf("api/armory error [%s]: %v", name, gearErr)
			return c.Status(fiber.StatusNotFound).SendString(gearErr.Error())
		}
		if lopecData != nil {
			gear.SecondClass = lopecData.SecondClass
			gear.LoaSpecPoint = lopecData.SpecPoint
		}
		return c.JSON(apiResponse{Text: gear.Format()})

	case "lopec":
		data, err := h.lopec.GetSpecPoint(ctx, name)
		if err != nil {
			log.Printf("api/lopec error [%s]: %v", name, err)
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}
		return c.JSON(apiResponse{Text: data.Format(name)})

	case "expedition":
		siblings, err := h.loa.GetSiblings(ctx, name)
		if err != nil {
			log.Printf("api/expedition error [%s]: %v", name, err)
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}
		return c.JSON(apiResponse{Text: lostark.FormatExpeditionRaid(name, siblings)})

	default:
		return c.Status(fiber.StatusNotFound).SendString("unknown resource: " + resource + "\nusage: /api/v1/{character|armory|lopec|expedition}/{name}")
	}
}
