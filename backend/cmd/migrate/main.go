// goose CLI 入口：up / down / status 等子命令直接透傳（D31 cmd 拆分）。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status|...> [args]")
		os.Exit(2)
	}
	cfg := config.New()
	db, err := sql.Open("pgx", cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatalf("開啟資料庫: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("設定 dialect: %v", err)
	}
	if err := goose.RunContext(context.Background(), os.Args[1], db, "database/migrations", os.Args[2:]...); err != nil {
		log.Fatalf("migrate %s: %v", os.Args[1], err)
	}
	// OpenFGA datastore schema(F2):與業務 schema 同一顆 DB,但以自己的版本表管理;
	// 僅 `up` 執行(單一 migration 入口,避免 server 啟動時偷改 schema)。
	if os.Args[1] == "up" {
		if err := migrateOpenFGA(context.Background(), cfg); err != nil {
			log.Fatalf("migrate openfga: %v", err)
		}
	}
}
