package distribute

import (
	"testing"
)

// TestDirectUse: 직접사용 입찰적정가 = floor(price × (n-1)/n)
func TestDirectUse(t *testing.T) {
	cases := []struct {
		n, price  int
		wantBid   int64
		wantDist  int64
	}{
		// 8인, 49000: floor(49000×7/8) = floor(42875) = 42875
		{8, 49000, 42875, 6125},
		// 4인, 40000: floor(40000×3/4) = 30000
		{4, 40000, 30000, 10000},
		// 4인, 49000: floor(49000×3/4) = floor(36750) = 36750
		{4, 49000, 36750, 12250},
	}
	for _, c := range cases {
		r := Result{N: c.n, Price: int64(c.price)}
		bid, dist := r.DirectUse()
		if bid != c.wantBid {
			t.Errorf("DirectUse(%d인, %d) bid: got %d, want %d", c.n, c.price, bid, c.wantBid)
		}
		if dist != c.wantDist {
			t.Errorf("DirectUse(%d인, %d) dist: got %d, want %d", c.n, c.price, dist, c.wantDist)
		}
	}
}

// TestBreakEven: 손익분기점 = sell_net - floor(sell_net/n)
func TestBreakEven(t *testing.T) {
	// 8인, 49000: sell_net=46550, floor(46550/8)=5818, bid=40732
	r := Result{N: 8, Price: 49000}

	if got := r.Fee(); got != 2450 {
		t.Errorf("Fee: got %d, want 2450", got)
	}

	bid, dist, profit := r.BreakEven()
	if bid != 40732 {
		t.Errorf("BreakEven bid: got %d, want 40732", bid)
	}
	if dist != 5818 {
		t.Errorf("BreakEven dist: got %d, want 5818", dist)
	}
	if profit != 8268 {
		t.Errorf("BreakEven profit: got %d, want 8268", profit)
	}
}

// TestSellAppropriate: 판매 입찰적정가 = ceil(손익분기점 / 1.1)
func TestSellAppropriate(t *testing.T) {
	cases := []struct {
		n, price  int
		wantBid   int64
		wantDist  int64
		wantProfit int64
	}{
		// 8인, 49000: ceil(40732/1.1) = ceil(37029.09) = 37030
		{8, 49000, 37030, 5290, 11970},
		// 8인, 53000: 손익분기점=44057, ceil(44057/1.1)=ceil(40051.8)=40052
		{8, 53000, 40052, 5721, 12948},
	}
	for _, c := range cases {
		r := Result{N: c.n, Price: int64(c.price)}
		bid, dist, profit := r.SellAppropriate()
		if bid != c.wantBid {
			t.Errorf("SellAppropriate(%d인, %d) bid: got %d, want %d", c.n, c.price, bid, c.wantBid)
		}
		if dist != c.wantDist {
			t.Errorf("SellAppropriate(%d인, %d) dist: got %d, want %d", c.n, c.price, dist, c.wantDist)
		}
		if profit != c.wantProfit {
			t.Errorf("SellAppropriate(%d인, %d) profit: got %d, want %d", c.n, c.price, profit, c.wantProfit)
		}
	}
}

func TestFmtGold(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1,000"},
		{42875, "42,875"},
		{49000, "49,000"},
		{1000000, "1,000,000"},
	}
	for _, c := range cases {
		if got := fmtGold(c.in); got != c.want {
			t.Errorf("fmtGold(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
