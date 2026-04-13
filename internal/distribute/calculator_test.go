package distribute

import (
	"testing"
)

func TestResult_8persons_49000(t *testing.T) {
	r := Result{N: 8, Price: 49000}

	t.Run("직접사용", func(t *testing.T) {
		bid, dist := r.DirectUse()
		if bid != 42875 {
			t.Errorf("입찰적정가: got %d, want 42875", bid)
		}
		if dist != 6125 {
			t.Errorf("분배금: got %d, want 6125", dist)
		}
	})

	t.Run("수수료", func(t *testing.T) {
		if got := r.Fee(); got != 2450 {
			t.Errorf("수수료: got %d, want 2450", got)
		}
	})

	t.Run("손익분기점", func(t *testing.T) {
		bid, dist, profit := r.BreakEven()
		if bid != 40732 {
			t.Errorf("손익분기점 bid: got %d, want 40732", bid)
		}
		if dist != 5818 {
			t.Errorf("손익분기점 분배금: got %d, want 5818", dist)
		}
		if profit != 8268 {
			t.Errorf("손익분기점 판매차익: got %d, want 8268", profit)
		}
	})
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

func TestResult_4persons(t *testing.T) {
	r := Result{N: 4, Price: 40000}

	bid, dist := r.DirectUse()
	if dist != 10000 {
		t.Errorf("4인 분배금: got %d, want 10000", dist)
	}
	if bid != 30000 {
		t.Errorf("4인 입찰적정가: got %d, want 30000", bid)
	}
}
