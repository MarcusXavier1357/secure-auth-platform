package cache

import (
	"context"
	"fmt"
	"time"
)

// PermissionCache guarda os codes de permissão por usuário em um SET Redis
// (cache-aside). Um resultado vazio é tratado como miss — usuários sem
// permissões sempre consultam o PostgreSQL, conforme o plano.
type PermissionCache struct {
	client *Client
	ttl    time.Duration
}

func NewPermissionCache(client *Client, ttl time.Duration) *PermissionCache {
	return &PermissionCache{client: client, ttl: ttl}
}

func (p *PermissionCache) key(userID int64) string {
	return p.client.key(fmt.Sprintf("permissions:user:%d", userID))
}

// Get retorna os codes cacheados. found=false indica cache miss.
func (p *PermissionCache) Get(ctx context.Context, userID int64) (codes []string, found bool, err error) {
	codes, err = p.client.rdb.SMembers(ctx, p.key(userID)).Result()
	if err != nil {
		return nil, false, err
	}
	return codes, len(codes) > 0, nil
}

func (p *PermissionCache) Set(ctx context.Context, userID int64, codes []string) error {
	key := p.key(userID)

	pipe := p.client.rdb.Pipeline()
	pipe.Del(ctx, key)
	if len(codes) > 0 {
		members := make([]any, len(codes))
		for i, c := range codes {
			members[i] = c
		}
		pipe.SAdd(ctx, key, members...)
	}
	pipe.Expire(ctx, key, p.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (p *PermissionCache) Invalidate(ctx context.Context, userID int64) error {
	return p.client.rdb.Del(ctx, p.key(userID)).Err()
}
