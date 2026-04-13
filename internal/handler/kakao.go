package handler

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/woohalabs2/yangobot/internal/command"
	"github.com/woohalabs2/yangobot/internal/distribute"
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

func (h *KakaoHandler) Handle(c fiber.Ctx) error {
	var req KakaoRequest
	if err := c.Bind().JSON(&req); err != nil {
		log.Printf("decode error: %v", err)
		return c.JSON(simpleText("요청을 처리할 수 없습니다."))
	}

	userID := req.UserRequest.User.ID
	if !h.limiter.Allow(userID) {
		return c.JSON(simpleText("현재 요청이 많아 처리가 어렵습니다.\n잠시 후 다시 이용해 주시면 감사하겠습니다."))
	}

	cmd, err := command.Parse(req.UserRequest.Utterance)
	if err != nil {
		return c.JSON(simpleText("알 수 없는 명령어입니다.\n사용법: /캐릭터 <닉네임>"))
	}

	ctx := c.Context()
	var result string

	switch cmd.Type {
	case command.CmdCharacter:
		info, err := h.loa.GetCharacter(ctx, cmd.Args[0])
		if err != nil {
			log.Printf("lostark API error: %v", err)
			return c.JSON(simpleText("캐릭터 정보를 가져오지 못했습니다."))
		}
		result = info.Format()
	case command.CmdSpec:
		data, err := h.lopec.GetSpecPoint(ctx, cmd.Args[0])
		if err != nil {
			log.Printf("lopec error: %v", err)
			return c.JSON(simpleText("스펙 점수를 가져오지 못했습니다."))
		}
		result = data.Format(cmd.Args[0])
	case command.CmdGear:
		name := cmd.Args[0]
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
			log.Printf("armory error: %v", gearErr)
			return c.JSON(simpleText("군장 정보를 가져오지 못했습니다."))
		}
		if lopecData != nil {
			gear.SecondClass = lopecData.SecondClass
			gear.LoaSpecPoint = lopecData.SpecPoint
		}
		result = gear.Format()
	case command.CmdExpedition:
		name := cmd.Args[0]
		siblings, err := h.loa.GetSiblings(ctx, name)
		if err != nil {
			log.Printf("expedition error: %v", err)
			return c.JSON(simpleText("원정대 정보를 가져오지 못했습니다."))
		}
		result = lostark.FormatExpeditionRaid(name, siblings)
	case command.CmdDistribute:
		if cmd.Gold > 0 {
			// 직접 금액 입력
			r := distribute.Result{N: cmd.N, Price: cmd.Gold}
			result = r.Format()
		} else if len(cmd.Args) > 0 {
			// 각인서 이름 조회
			item, err := h.loa.GetMarketPrice(ctx, cmd.Args[0])
			if err != nil {
				log.Printf("market price error [%s]: %v", cmd.Args[0], err)
				return c.JSON(simpleText("거래소 시세를 가져오지 못했습니다."))
			}
			pricePerItem := item.PricePerItem()
			if pricePerItem <= 0 {
				return c.JSON(simpleText("거래 가능한 아이템이 없습니다."))
			}
			r := distribute.Result{
				N:     cmd.N,
				Price: pricePerItem,
				Auction: &distribute.AuctionInfo{
					Name:         item.Name,
					Grade:        item.Grade,
					CurrentPrice: pricePerItem,
					YDayPrice:    item.YDayPricePerItem(),
				},
			}
			result = r.Format()
		}
	default:
		result = "지원하지 않는 명령어입니다."
	}

	return c.JSON(simpleText(result))
}
