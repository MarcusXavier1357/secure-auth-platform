package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LastLoginInfo guarda o último login bem-sucedido para detecção de viagem impossível.
type LastLoginInfo struct {
	CountryCode string    `json:"country"`
	IP          string    `json:"ip"`
	At          time.Time `json:"at"`
}

type LastLoginStore struct {
	client *Client
	ttl    time.Duration
}

func NewLastLoginStore(client *Client, ttl time.Duration) *LastLoginStore {
	if ttl == 0 {
		ttl = 30 * 24 * time.Hour
	}
	return &LastLoginStore{client: client, ttl: ttl}
}

func lastLoginKey(userID int64) string {
	return fmt.Sprintf("last_login:%d", userID)
}

func (s *LastLoginStore) Get(ctx context.Context, userID int64) (*LastLoginInfo, error) {
	raw, err := s.client.rdb.Get(ctx, s.client.key(lastLoginKey(userID))).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var info LastLoginInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *LastLoginStore) Set(ctx context.Context, userID int64, info LastLoginInfo) error {
	raw, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return s.client.rdb.Set(ctx, s.client.key(lastLoginKey(userID)), raw, s.ttl).Err()
}
