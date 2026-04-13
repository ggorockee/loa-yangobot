package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/woohalabs2/yangobot/internal/cache"
	"github.com/woohalabs2/yangobot/internal/handler"
	"github.com/woohalabs2/yangobot/internal/lopec"
	"github.com/woohalabs2/yangobot/internal/lostark"
	"github.com/woohalabs2/yangobot/internal/ratelimit"
)

func main() {
	loaAPIKey := os.Getenv("LOSTARK_API_KEY")
	if loaAPIKey == "" {
		log.Fatal("LOSTARK_API_KEY is required")
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisCache := cache.New(redisAddr)
	loaClient := lostark.NewClient(loaAPIKey, redisCache)
	lopecClient := lopec.NewClient(redisCache)
	limiter := ratelimit.New(10, 100) // 10 req/s per user, burst 100

	kakaoHandler := handler.NewKakaoHandler(loaClient, lopecClient, limiter)
	apiHandler := handler.NewAPIHandler(loaClient, lopecClient, limiter)

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Post("/webhook/kakao", kakaoHandler.Handle)
	app.Get("/api/v1/distribute/:n/:price", apiHandler.HandleDistribute)
	app.Get("/api/v1/:resource/:name", apiHandler.Handle)
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	log.Println("yangobot listening on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
