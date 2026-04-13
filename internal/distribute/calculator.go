package distribute

import (
	"fmt"
	"strings"
)

// Result는 분배금 계산 결과입니다.
type Result struct {
	N     int
	Price int64
}

// DirectUse는 직접사용 시나리오 결과입니다.
//
// 공식: 입찰적정가 = floor(price × (n-1)/n)
func (r Result) DirectUse() (bidPrice, distribution int64) {
	bidPrice = r.Price * int64(r.N-1) / int64(r.N)
	distribution = r.Price - bidPrice
	return
}

// Fee는 판매 수수료(5%)를 반환합니다.
func (r Result) Fee() int64 {
	return r.Price * 5 / 100
}

// SellNet은 수수료 차감 후 실수령액입니다.
func (r Result) SellNet() int64 {
	return r.Price - r.Fee()
}

// BreakEven은 손익분기점 시나리오 결과입니다.
//
// 낙찰자 판매 순이익 = 파티원 1인 분배금이 되는 입찰가.
// 공식: bid = sell_net - floor(sell_net/n)
func (r Result) BreakEven() (bidPrice, distribution, grossProfit int64) {
	sellNet := r.SellNet()
	distribution = sellNet / int64(r.N) // floor
	bidPrice = sellNet - distribution
	grossProfit = r.Price - bidPrice
	return
}

// SellAppropriate는 판매 입찰적정가 시나리오 결과입니다.
//
// 낙찰자에게 수고비 10% 마진을 보장하는 입찰가.
// 공식: bid = ceil(손익분기점 / 1.1) = ceil(손익분기점 × 10 / 11)
func (r Result) SellAppropriate() (bidPrice, distribution, grossProfit int64) {
	breakEvenBid, _, _ := r.BreakEven()
	// ceil(breakEvenBid * 10 / 11)
	bidPrice = (breakEvenBid*10 + 10) / 11
	distribution = bidPrice / int64(r.N-1)
	grossProfit = r.Price - bidPrice
	return
}

// Format은 카카오/API 응답용 텍스트를 반환합니다.
func (r Result) Format() string {
	var b strings.Builder

	directBid, directDist := r.DirectUse()
	b.WriteString(fmt.Sprintf("| 직접사용 (%d인)\n", r.N))
	b.WriteString(fmt.Sprintf("* 입찰적정가 %s\n", fmtGold(directBid)))
	b.WriteString(fmt.Sprintf("* 분배금     %s\n", fmtGold(directDist)))

	breakBid, breakDist, breakProfit := r.BreakEven()
	b.WriteString("\n| 판매\n")
	b.WriteString(fmt.Sprintf("* 수수료       %s\n", fmtGold(r.Fee())))
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("* 손익분기점   %s\n", fmtGold(breakBid)))
	b.WriteString(fmt.Sprintf("* 분배금       %s\n", fmtGold(breakDist)))
	b.WriteString(fmt.Sprintf("* 판매차익     %s\n", fmtGold(breakProfit)))

	appBid, appDist, appProfit := r.SellAppropriate()
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("* 입찰적정가   %s\n", fmtGold(appBid)))
	b.WriteString(fmt.Sprintf("* 분배금       %s\n", fmtGold(appDist)))
	b.WriteString(fmt.Sprintf("* 판매차익     %s", fmtGold(appProfit)))

	return b.String()
}

// fmtGold는 금액을 3자리 콤마 형식으로 반환합니다. (예: 42,875)
func fmtGold(v int64) string {
	s := fmt.Sprintf("%d", v)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	rem := len(s) % 3
	for i, c := range s {
		if i > 0 && (i-rem)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}
