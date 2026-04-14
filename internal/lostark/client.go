package lostark

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/woohalabs2/yangobot/internal/cache"
)

const (
	baseURL    = "https://developer-lostark.game.onstove.com"
	cacheTTL   = 5 * time.Minute
	keyTimeout = 1 * time.Second // 키 하나당 타임아웃
)

// ErrAllKeysExhausted는 등록된 모든 API 키가 타임아웃/레이트리밋에 걸렸을 때 반환됩니다.
var ErrAllKeysExhausted = errors.New("현재 요청이 너무 많습니다. 잠시 후 다시 이용해 주세요.")

type Client struct {
	keys   []string
	keyIdx uint64 // round-robin 카운터 (atomic)
	http   *http.Client
	cache  *cache.Redis
}

// NewClient는 API 키 목록을 받아 Client를 생성합니다.
// 요청마다 round-robin으로 시작 키를 분산하고, 실패 시 다음 키로 폴백합니다.
func NewClient(keys []string, c *cache.Redis) *Client {
	return &Client{
		keys:  keys,
		http:  &http.Client{},
		cache: c,
	}
}

// doWithFallback은 round-robin으로 시작 키를 선택한 뒤 순환 폴백합니다.
// 키당 타임아웃(keyTimeout) 초과 또는 429 응답 시 다음 키로 넘어갑니다.
// 모든 키 소진 시 ErrAllKeysExhausted를 반환합니다.
func (c *Client) doWithFallback(ctx context.Context, buildReq func(ctx context.Context, key string) (*http.Request, error)) (*http.Response, error) {
	n := uint64(len(c.keys))
	start := atomic.AddUint64(&c.keyIdx, 1) % n
	for i := uint64(0); i < n; i++ {
		idx := (start + i) % n
		key := c.keys[idx]
		keyCtx, cancel := context.WithTimeout(ctx, keyTimeout)
		req, err := buildReq(keyCtx, key)
		if err != nil {
			cancel()
			return nil, err
		}
		resp, err := c.http.Do(req)
		cancel()
		if err != nil {
			log.Printf("[lostark] key[%d] failed: %v", idx+1, err)
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			log.Printf("[lostark] key[%d] rate limited (429)", idx+1)
			continue
		}
		return resp, nil
	}
	return nil, ErrAllKeysExhausted
}

// GetCharacter는 캐릭터 이름으로 기본 정보를 조회합니다.
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

// GetSiblings는 캐릭터 이름으로 원정대 전체 캐릭터 목록을 반환합니다.
// Redis 캐시가 있으면 캐시된 값을 반환합니다 (캐시 키: siblings:<name>).
func (c *Client) GetSiblings(ctx context.Context, name string) ([]CharacterInfo, error) {
	cacheKey := "siblings:" + name

	var cached []CharacterInfo
	if err := c.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	resp, err := c.doWithFallback(ctx, func(keyCtx context.Context, key string) (*http.Request, error) {
		endpoint := fmt.Sprintf("%s/characters/%s/siblings", baseURL, url.PathEscape(name))
		req, err := http.NewRequestWithContext(keyCtx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "bearer "+key)
		req.Header.Set("Accept", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, err
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
