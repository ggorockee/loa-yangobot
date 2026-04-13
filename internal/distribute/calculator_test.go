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

// TestSellBid: 판매 입찰적정가 = floor(price × 0.95 × (n-1)/n)
func TestSellBid(t *testing.T) {
	// 8인, 49000: floor(49000×0.95×7/8) = floor(46550×7/8) = floor(40731.25) = 40731
	r := Result{N: 8, Price: 49000}

	if got := r.Fee(); got != 2450 {
		t.Errorf("Fee: got %d, want 2450", got)
	}

	bid, dist, profit := r.SellBid()
	// floor(46550 × 7/8) = floor(40731.25) = 40731
	if bid != 40731 {
		t.Errorf("SellBid bid: got %d, want 40731", bid)
	}
	// floor(40731/7) = 5818
	if dist != 5818 {
		t.Errorf("SellBid dist: got %d, want 5818", dist)
	}
	// 49000 - 40731 = 8269
	if profit != 8269 {
		t.Errorf("SellBid profit: got %d, want 8269", profit)
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
