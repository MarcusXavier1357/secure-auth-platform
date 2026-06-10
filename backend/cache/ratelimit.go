package cache

import (
	"context"
	"time"
)

// RateLimiter implementa fixed window com INCR + EXPIRE na primeira tentativa.
type RateLimiter struct {
	client *Client
	limit  int64
	window time.Duration
}

func NewRateLimiter(client *Client, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, limit: limit, window: window}
}

func (r *RateLimiter) Window() time.Duration {
	return r.window
}

// Increment incrementa o contador da chave e retorna o total na janela atual.
func (r *RateLimiter) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := r.client.key(key)

	count, err := r.client.rdb.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		r.client.rdb.Expire(ctx, fullKey, r.window)
	}
	return count, nil
}

func (r *RateLimiter) Exceeded(count int64) bool {
	return count > r.limit
}

// Reset limpa o contador (ex.: após login bem-sucedido, para a chave de email).
func (r *RateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.rdb.Del(ctx, r.client.key(key)).Err()
}
