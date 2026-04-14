package lostark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const marketCacheTTL = 3 * time.Minute

// MarketItem은 거래소 아이템 시세 정보입니다.
type MarketItem struct {
	Id              int     `json:"Id"`
	Name            string  `json:"Name"`
	Grade           string  `json:"Grade"`
	BundleCount     int     `json:"BundleCount"`
	CurrentMinPrice int64   `json:"CurrentMinPrice"`
	YDayAvgPrice    float64 `json:"YDayAvgPrice"` // API가 소수점 반환
}

// PricePerItem은 BundleCount로 나눈 개당 현재 최저가를 반환합니다.
func (m *MarketItem) PricePerItem() int64 {
	b := int64(m.BundleCount)
	if b <= 0 {
		b = 1
	}
	return m.CurrentMinPrice / b
}

// YDayPricePerItem은 BundleCount로 나눈 개당 전일 평균가를 반환합니다 (반올림).
func (m *MarketItem) YDayPricePerItem() int64 {
	b := float64(m.BundleCount)
	if b <= 0 {
		b = 1
	}
	return int64(m.YDayAvgPrice/b + 0.5)
}

type marketRequest struct {
	Sort          string `json:"Sort"`
	SortCondition string `json:"SortCondition"`
	CategoryCode  int    `json:"CategoryCode,omitempty"`
	ItemGrade     string `json:"ItemGrade,omitempty"`
	ItemName      string `json:"ItemName"`
	PageNo        int    `json:"PageNo"`
}

type marketResponse struct {
	TotalCount int          `json:"TotalCount"`
	Items      []MarketItem `json:"Items"`
}

// GetMarketPrice는 아이템 이름으로 거래소 최저가를 조회합니다.
// 현재 최저가 기준 오름차순 정렬 후 첫 번째 결과를 반환합니다.
// Redis에 3분 캐시됩니다.
func (c *Client) GetMarketPrice(ctx context.Context, itemName string) (*MarketItem, error) {
	cacheKey := "market:" + itemName
	var cached MarketItem
	if err := c.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	reqBody := marketRequest{
		Sort:          "CURRENT_MIN_PRICE",
		SortCondition: "ASC",
		CategoryCode:  40000, // 각인서
		ItemGrade:     "유물",
		ItemName:      itemName,
		PageNo:        1,
	}

	resp, err := c.doWithFallback(ctx, func(keyCtx context.Context, key string) (*http.Request, error) {
		// POST body는 시도마다 새로 생성 (Reader는 소모성)
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		req, err := http.NewRequestWithContext(keyCtx, http.MethodPost,
			baseURL+"/markets/items", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "bearer "+key)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("아이템을 찾을 수 없습니다: %s", itemName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result marketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("아이템을 찾을 수 없습니다: %s", itemName)
	}

	item := result.Items[0]
	_ = c.cache.Set(ctx, cacheKey, item, marketCacheTTL)
	return &item, nil
}
