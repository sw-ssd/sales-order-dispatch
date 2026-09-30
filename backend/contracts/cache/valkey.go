// Package cache 集中 Valkey client 初始化（D31 third_party 規則）。
package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// NewClient 建立 Valkey（Redis 相容）client。
func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

// Ping 確認連線可用。
func Ping(ctx context.Context, c *redis.Client) error {
	return c.Ping(ctx).Err()
}
