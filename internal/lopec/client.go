package lopec

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/woohalabs2/yangobot/internal/cache"
)

const (
	lopecBaseURL = "https://lopec.kr/character/specPoint"
	cacheTTL     = 5 * time.Minute
	timeout      = 8 * time.Second
)

var (
	specPointRe  = regexp.MustCompile(`"specPoint":([0-9.]+)`)
	itemLevelRe  = regexp.MustCompile(`"itemLevel":([0-9.]+)`)
	secondClassRe = regexp.MustCompile(`"secondClass":"([^"]+)"`)
)

var tiers = []struct {
	name   string
	cutoff float64
}{
	{"에스더", 6200},
	{"그랜드 마스터", 5600},
	{"마스터", 4900},
	{"다이아", 4000},
	{"골드", 3000},
	{"실버", 2000},
	{"브론즈", 0},
}

type SpecData struct {
	SpecPoint   float64
	ItemLevel   float64
	SecondClass string
}

func (s *SpecData) TierName() string {
	for _, t := range tiers {
		if s.SpecPoint >= t.cutoff {
			return t.name
		}
	}
	return "브론즈"
}

func (s *SpecData) Format(name string) string {
	return fmt.Sprintf(
		"[%s] 로펙 스펙점수\n점수: %.2f (%s)\n아이템 레벨: %.2f",
		name,
		s.SpecPoint,
		s.TierName(),
		s.ItemLevel,
	)
}

type Client struct {
	http  *http.Client
	cache *cache.Redis
}

func NewClient(c *cache.Redis) *Client {
	return &Client{
		http:  &http.Client{Timeout: timeout},
		cache: c,
	}
}

func (c *Client) GetSpecPoint(ctx context.Context, name string) (*SpecData, error) {
	cacheKey := "lopec:spec:" + name

	var cached SpecData
	if err := c.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	endpoint := fmt.Sprintf("%s/%s", lopecBaseURL, url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("RSC", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; yangobot/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lopec 응답 오류: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	specMatch := specPointRe.FindSubmatch(body)
	if specMatch == nil {
		return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다: %s", name)
	}
	specPoint, _ := strconv.ParseFloat(string(specMatch[1]), 64)

	itemMatch := itemLevelRe.FindSubmatch(body)
	var itemLevel float64
	if itemMatch != nil {
		itemLevel, _ = strconv.ParseFloat(string(itemMatch[1]), 64)
	}

	var secondClass string
	if m := secondClassRe.FindSubmatch(body); m != nil {
		secondClass = string(m[1])
	}

	data := &SpecData{SpecPoint: specPoint, ItemLevel: itemLevel, SecondClass: secondClass}
	_ = c.cache.Set(ctx, cacheKey, data, cacheTTL)
	return data, nil
}
