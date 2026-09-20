// Package database 集中 Ent client / pgx pool 初始化（D31 third_party 規則）。
// Server 的 ent client 與 seed 的 pgx pool 皆由此處建立,避免各套件自行 sql.Open。
package database

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// Open 建立 pgx pool 並 Ping 確認連線（seed 等需要原生 pool 的用途）。
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

// OpenEnt 建立 PostgreSQL ent client（pgx driver）並 Ping 確認連線。
// Server 於 Init()/mountDomain 使用;統一由此處初始化,不再於各套件自行 sql.Open + ent.NewClient。
func OpenEnt(dsn string) (*ent.Client, error) {
	db, err := OpenSQL(dsn)
	if err != nil {
		return nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), nil
}

// OpenSQL 建立 pgx 連線（*sql.DB）並 Ping 確認。供需要自行決定 ent driver 的呼叫點使用
// —— 業務 client 必須由 dbtenant.NewClient 包上 RLS 裝飾器（請求交易內 SET LOCAL）,
// 故 Server 取的是原始 *sql.DB;sql.Open 仍集中於此,不散落各套件（D31）。
func OpenSQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Probe 僅做連線可用性檢查（production fail-fast 用）,用完立即關閉,不留閒置連線。
func Probe(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return db.PingContext(ctx)
}
