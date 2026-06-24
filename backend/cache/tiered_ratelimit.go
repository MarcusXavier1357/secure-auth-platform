package cache

import (
	"context"
	"fmt"
	"time"
)

// RateTier define um limiar de tentativas e a duração do bloqueio ao atingi-lo.
type RateTier struct {
	Threshold int64
	Block     time.Duration
}

// TieredRateLimiter aplica bloqueio escalonado conforme o total de tentativas.
type TieredRateLimiter struct {
	client *Client
	tiers  []RateTier
	// TTL do contador de tentativas — deve cobrir a janela mais longa.
	counterTTL time.Duration
}

func NewTieredRateLimiter(client *Client, tiers []RateTier, counterTTL time.Duration) *TieredRateLimiter {
	return &TieredRateLimiter{client: client, tiers: tiers, counterTTL: counterTTL}
}

// MaxTierThreshold retorna o maior limiar configurado.
func (r *TieredRateLimiter) MaxTierThreshold() int64 {
	var maxThreshold int64
	for _, tier := range r.tiers {
		if tier.Threshold > maxThreshold {
			maxThreshold = tier.Threshold
		}
	}
	return maxThreshold
}

// Check incrementa o contador e retorna se a chave está bloqueada.
// retryAfter indica quanto tempo aguardar quando bloqueado; count é o total de tentativas.
func (r *TieredRateLimiter) Check(ctx context.Context, key string) (blocked bool, retryAfter time.Duration, count int64, err error) {
	fullKey := r.client.key(key)
	blockKey := r.client.key(key + ":block")

	ttl, err := r.client.rdb.TTL(ctx, blockKey).Result()
	if err != nil {
		return false, 0, 0, err
	}
	if ttl > 0 {
		stored, _ := r.client.rdb.Get(ctx, fullKey).Int64()
		return true, ttl, stored, nil
	}

	count, err = r.client.rdb.Incr(ctx, fullKey).Result()
	if err != nil {
		return false, 0, 0, err
	}
	if count == 1 {
		r.client.rdb.Expire(ctx, fullKey, r.counterTTL)
	}

	block := r.blockDuration(count)
	if block > 0 {
		if err := r.client.rdb.Set(ctx, blockKey, "1", block).Err(); err != nil {
			return false, 0, 0, err
		}
		return true, block, count, nil
	}
	return false, 0, count, nil
}

func (r *TieredRateLimiter) blockDuration(count int64) time.Duration {
	var block time.Duration
	for _, tier := range r.tiers {
		if count >= tier.Threshold && tier.Block > block {
			block = tier.Block
		}
	}
	return block
}

// Reset limpa contador e bloqueio (ex.: após login bem-sucedido).
func (r *TieredRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := r.client.key(key)
	blockKey := r.client.key(key + ":block")
	return r.client.rdb.Del(ctx, fullKey, blockKey).Err()
}

func LoginIPKey(ip string) string {
	return fmt.Sprintf("login:ip:%s", ip)
}

func LoginEmailKey(email string) string {
	return fmt.Sprintf("login:email:%s", email)
}
