package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Client encapsula o go-redis aplicando o prefixo de chaves configurável
// (REDIS_KEY_PREFIX) para ambientes compartilhados.
type Client struct {
	rdb    *redis.Client
	prefix string
}

func NewClient(url, prefix string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &Client{rdb: redis.NewClient(opts), prefix: prefix}, nil
}

func (c *Client) key(parts string) string {
	return c.prefix + parts
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
