package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(addr string) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:         addr,
			DialTimeout:  2 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			PoolSize:     10,
		}),
	}
}

// Get은 키에 해당하는 값을 JSON 역직렬화하여 dst에 저장합니다.
func (r *Redis) Get(ctx context.Context, key string, dst any) error {
	b, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// Set은 값을 JSON 직렬화하여 TTL과 함께 저장합니다.
func (r *Redis) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

// Ping은 Redis 연결 상태를 확인합니다.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// IncrWindow는 fixed-window 카운터를 원자적으로 증가시키고 현재 카운트를 반환합니다.
// 키가 처음 생성될 때 ttl을 설정합니다 (Lua 스크립트로 INCR+EXPIRE 원자 보장).
func (r *Redis) IncrWindow(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	script := redis.NewScript(`
		local n = redis.call('INCR', KEYS[1])
		if n == 1 then
			redis.call('EXPIRE', KEYS[1], ARGV[1])
		end
		return n
	`)
	return script.Run(ctx, r.client, []string{key}, int(ttl.Seconds())).Int64()
}
