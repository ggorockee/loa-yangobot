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

// SellBid는 판매 시나리오 입찰적정가 결과입니다.
//
// 공식: 입찰적정가 = floor(price × 0.95 × (n-1)/n)
//
// 거래소에서 판매할 때 수수료(5%) 제외 실수령액을 전체 인원이 공평하게 나눌 수 있는 입찰가.
func (r Result) SellBid() (bidPrice, distribution, grossProfit int64) {
	sellNet := r.SellNet()
	bidPrice = sellNet * int64(r.N-1) / int64(r.N)
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

	sellBid, sellDist, sellProfit := r.SellBid()
	b.WriteString("\n| 판매\n")
	b.WriteString(fmt.Sprintf("* 수수료     %s\n", fmtGold(r.Fee())))
	b.WriteString(fmt.Sprintf("* 입찰적정가 %s\n", fmtGold(sellBid)))
	b.WriteString(fmt.Sprintf("* 분배금     %s\n", fmtGold(sellDist)))
	b.WriteString(fmt.Sprintf("* 판매차익   %s", fmtGold(sellProfit)))

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
