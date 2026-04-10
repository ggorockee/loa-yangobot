package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/woohalabs2/sybot/internal/cache"
	"github.com/woohalabs2/sybot/internal/handler"
	"github.com/woohalabs2/sybot/internal/lopec"
	"github.com/woohalabs2/sybot/internal/lostark"
	"github.com/woohalabs2/sybot/internal/ratelimit"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/kakao", kakaoHandler.Handle)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("sybot listening on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
