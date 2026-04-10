package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// Limiter는 사용자별 토큰 버킷 Rate Limiter입니다.
type Limiter struct {
	mu      sync.Mutex
	users   map[string]*rate.Limiter
	r       rate.Limit // 초당 허용 요청 수
	b       int        // 버스트 크기
}

func New(rps float64, burst int) *Limiter {
	return &Limiter{
		users: make(map[string]*rate.Limiter),
		r:     rate.Limit(rps),
		b:     burst,
	}
}

// Allow는 해당 userID의 요청을 허용할지 여부를 반환합니다.
func (l *Limiter) Allow(userID string) bool {
	return l.getLimiter(userID).Allow()
}

func (l *Limiter) getLimiter(userID string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.users[userID]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.b)
	l.users[userID] = lim
	return lim
}
