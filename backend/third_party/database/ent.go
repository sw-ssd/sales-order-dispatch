// Package database 集中 Ent client / pgx pool 初始化（D31 third_party 規則）。
// Ent client 掛接（含 RLS driver hook）於 01-auth 計畫 Task 1 schema 就緒後加入。
package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open 建立 pgx pool 並 Ping 確認連線。
func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
