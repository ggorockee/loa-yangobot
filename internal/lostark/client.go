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

	// Lost Ark API 분당 100건 제한. 안전 마진 5건.
	apiRateLimit = 95
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

// checkAPIRateLimit은 Redis fixed-window로 전체 pod 합산 분당 API 호출 수를 체크합니다.
// Redis 장애 시 fail-open (통과)으로 동작합니다.
func (c *Client) checkAPIRateLimit(ctx context.Context) error {
	minute := time.Now().UTC().Format("200601021504")
	key := "ratelimit:lostark:" + minute
	count, err := c.cache.IncrWindow(ctx, key, time.Minute)
	if err != nil {
		// Redis 오류 시 통과 — Lost Ark API 응답에서 429로 자체 확인
		return nil
	}
	if count > apiRateLimit {
		return fmt.Errorf("lost ark API rate limit 초과 (%d/100 req/min)", count)
	}
	return nil
}

// GetCharacter는 캐릭터 이름으로 기본 정보를 조회합니다.
// Redis 캐시가 있으면 캐시된 값을 반환합니다.
func (c *Client) GetCharacter(ctx context.Context, name string) (*CharacterInfo, error) {
	siblings, err := c.GetSiblings(ctx, name)
	if err != nil {
		return nil, err
	}
	for _, s := range siblings {
		if s.CharacterName == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
}

// GetSiblings는 캐릭터 이름으로 동일 계정의 원정대 전체 캐릭터 목록을 반환합니다.
// Redis 캐시가 있으면 캐시된 값을 반환합니다 (캐시 키: siblings:<name>).
func (c *Client) GetSiblings(ctx context.Context, name string) ([]CharacterInfo, error) {
	cacheKey := "siblings:" + name

	var cached []CharacterInfo
	if err := c.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	if err := c.checkAPIRateLimit(ctx); err != nil {
		return nil, err
	}

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

	var siblings []CharacterInfo
	if err := json.NewDecoder(resp.Body).Decode(&siblings); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(siblings) == 0 {
		return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
	}

	_ = c.cache.Set(ctx, cacheKey, siblings, cacheTTL)
	return siblings, nil
}
