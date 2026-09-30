// goose CLI 入口：up / down / status 等子命令直接透傳（D31 cmd 拆分）。
//
// 業務遷移與平台遷移分屬兩個目錄、兩張獨立版本表（同 OpenFGA 先例）：
//   - database/migrations        → 版本表 goose_db_version（業務 schema）
//   - database/platform_migrations → 版本表 platform_goose_db_version（平台域）
//
// 兩者在同一顆 DB 上，但版本追蹤互不干擾。平台域 00054 的 backfill 會 JOIN 業務
// companies 表，故 up 必須業務先、平台後；down 反向（平台先、業務後）。
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

// businessMigrationsDir / platformMigrationsDir 為兩套遷移的檔案目錄。
const (
	businessMigrationsDir = "database/migrations"
	platformMigrationsDir = "database/platform_migrations"
	bizGooseTable         = "goose_db_version"
	platGooseTable        = "platform_goose_db_version"
)

// migrateDir 以指定版本表對單一目錄執行 goose 子命令。SetTableName 是全域狀態，
// 故每次呼叫前設定、呼叫後（defer）還原，避免汙染後續目錄或 OpenFGA。
func migrateDir(ctx context.Context, db *sql.DB, dir, table, cmd string, args ...string) error {
	prev := goose.TableName()
	goose.SetTableName(table)
	defer goose.SetTableName(prev)
	return goose.RunContext(ctx, cmd, db, dir, args...)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status|...> [args]")
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	cfg := config.New()
	db, err := sql.Open("pgx", cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("開啟資料庫: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("設定 dialect: %v", err)
	}

	// up：業務先、平台後（00054 backfill 依賴 companies）。
	// down：平台先、業務後（反向順序，避免平台表孤立引用業務表）。
	// 其餘子命令（status 等）順序無關。
	first, second := businessMigrationsDir, platformMigrationsDir
	firstTable, secondTable := bizGooseTable, platGooseTable
	if cmd == "down" || cmd == "down-to" {
		first, second = platformMigrationsDir, businessMigrationsDir
		firstTable, secondTable = platGooseTable, bizGooseTable
	}
	if err := migrateDir(context.Background(), db, first, firstTable, cmd, args...); err != nil {
		log.Fatalf("migrate %s（%s，版本表 %s）: %v", cmd, first, firstTable, err)
	}
	if err := migrateDir(context.Background(), db, second, secondTable, cmd, args...); err != nil {
		log.Fatalf("migrate %s（%s，版本表 %s）: %v", cmd, second, secondTable, err)
	}

	// OpenFGA datastore schema(F2):與業務 schema 同一顆 DB,但以自己的版本表管理;
	// 僅 `up` 執行(單一 migration 入口,避免 server 啟動時偷改 schema)。
	if cmd == "up" {
		if err := migrateOpenFGA(context.Background(), cfg); err != nil {
			log.Fatalf("migrate openfga: %v", err)
		}
	}
}
