package lostark

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/woohalabs2/yangobot/internal/cache"
)

const (
	baseURL        = "https://developer-lostark.game.onstove.com"
	cacheTTL       = 5 * time.Minute
	requestTimeout = 8 * time.Second
)

type Client struct {
	apiKey  string
	http    *http.Client
	cache   *cache.Redis
}

func NewClient(apiKey string, c *cache.Redis) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: requestTimeout,
		},
		cache: c,
	}
}

// GetCharacter는 캐릭터 이름으로 기본 정보를 조회합니다.
// Redis 캐시가 있으면 캐시된 값을 반환합니다.
func (c *Client) GetCharacter(ctx context.Context, name string) (*CharacterInfo, error) {
	cacheKey := "character:" + name

	// 캐시 조회
	var info CharacterInfo
	if err := c.cache.Get(ctx, cacheKey, &info); err == nil {
		return &info, nil
	}

	// API 호출
	endpoint := fmt.Sprintf("%s/characters/%s/siblings", baseURL, url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// siblings 응답은 []CharacterInfo 배열 — 요청한 캐릭터 이름과 매칭
	var siblings []CharacterInfo
	if err := json.NewDecoder(resp.Body).Decode(&siblings); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	for _, s := range siblings {
		if s.CharacterName == name {
			info = s
			_ = c.cache.Set(ctx, cacheKey, &info, cacheTTL)
			return &info, nil
		}
	}

	return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
}
