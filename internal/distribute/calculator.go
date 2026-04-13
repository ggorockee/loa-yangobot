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
// 공식:
//   - 분배금 = floor(price / n)
//   - 입찰적정가 = price - 분배금
func (r Result) DirectUse() (bidPrice, distribution int64) {
	distribution = r.Price / int64(r.N)
	bidPrice = r.Price - distribution
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

// BreakEven는 손익분기점 시나리오 결과입니다.
//
// 손익분기점: 판매 순이익 = 타인 분배금
//
//	sell_net - bid = floor(bid / (n-1))
//
// 공식:
//   - 분배금 = floor(sell_net / n)
//   - bid    = sell_net - 분배금
func (r Result) BreakEven() (bidPrice, distribution, grossProfit int64) {
	sellNet := r.SellNet()
	distribution = sellNet / int64(r.N)
	bidPrice = sellNet - distribution
	grossProfit = r.Price - bidPrice
	return
}

// SellAppropriate는 판매 입찰적정가 시나리오 결과입니다.
//
// TODO: 정확한 공식 확인 필요.
// n=8, price=49000 기준 예상값 37,030 vs 현 공식 산출값 37,057 (27 차이).
//
// 현 공식: bid = floor(직접사용_입찰적정가 - sell_net / n)
func (r Result) SellAppropriate() (bidPrice, distribution, grossProfit int64) {
	sellNet := r.SellNet()
	directBid, _ := r.DirectUse()
	bidPrice = directBid - sellNet/int64(r.N)
	if bidPrice <= 0 {
		return 0, 0, 0
	}
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

	b.WriteString("\n| 판매\n")
	b.WriteString(fmt.Sprintf("* 수수료     %s\n", fmtGold(r.Fee())))

	breakBid, breakDist, breakProfit := r.BreakEven()
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("* 손익분기점   %s\n", fmtGold(breakBid)))
	b.WriteString(fmt.Sprintf("* 분배금       %s\n", fmtGold(breakDist)))
	b.WriteString(fmt.Sprintf("* 판매차익     %s\n", fmtGold(breakProfit)))

	appBid, appDist, appProfit := r.SellAppropriate()
	if appBid > 0 {
		b.WriteString("---\n")
		b.WriteString(fmt.Sprintf("* 입찰적정가    %s\n", fmtGold(appBid)))
		b.WriteString(fmt.Sprintf("* 분배금        %s\n", fmtGold(appDist)))
		b.WriteString(fmt.Sprintf("* 판매차익      %s", fmtGold(appProfit)))
	}

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
