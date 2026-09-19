//go:build integration

// cmd/migrate 的真 PostgreSQL 整合測試。以拋棄式容器(見 internal/testsupport)驗證兩個
// 曾無自動化守門的缺陷:
//
//	F1 00005 的表層 UNIQUE 表達式:全新資料庫 `migrate up` 直接失敗(48b3d34 修復)。
//	F2 內嵌 OpenFGA 的 schema 未建立 + datastore 型別限額為 0 + 缺表時靜默降級(65826e4 修復)。
//
// 業務遷移走與 cmd/migrate 相同的 goose 路徑(同 dialect、同 migrations 目錄、同版本表),
// OpenFGA 遷移直接呼叫 main 所用的 migrateOpenFGA,確保測的是實際程式碼而非複製品。
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

const (
	// migrationsDir 相對套件目錄(go test 以套件目錄為 cwd)。
	migrationsDir = "../../database/migrations"
	// businessGooseTable 為業務遷移的版本表(goose 預設值,main.go 未改動)。
	businessGooseTable = "goose_db_version"
	// openfgaTables 為 OpenFGA postgres schema 的核心表(assets/migrations/postgres/)。
	openfgaTables = "tuple, authorization_model, store, assertion, changelog"
)

// TestIntegrationFreshDatabaseMigrateUp F1 迴歸:全新資料庫上 `migrate up` 必須成功。
// 00005 若回到表層 UNIQUE 表達式,PostgreSQL 會拒絕該 DDL,goose 直接回錯(測試紅)。
func TestIntegrationFreshDatabaseMigrateUp(t *testing.T) {
	dsn := testsupport.Postgres(t)

	if err := businessUp(t, dsn); err != nil {
		t.Fatalf("全新資料庫 migrate up 必須成功(F1:00005 的表層 UNIQUE 表達式會使 DDL 失敗): %v", err)
	}

	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()

	// 索引確實由 00005 建立,且以表達式(md5/COALESCE)表達去重語意。
	var indexdef string
	if err := db.QueryRow(
		`SELECT indexdef FROM pg_indexes WHERE tablename = 'role_permissions' AND indexname = 'role_permissions_unique_idx'`,
	).Scan(&indexdef); err != nil {
		t.Fatalf("查 role_permissions_unique_idx: %v(F1:唯一性必須由 00005 的唯一索引保證)", err)
	}
	for _, want := range []string{"UNIQUE", "COALESCE", "md5", "conditions"} {
		if !strings.Contains(indexdef, want) {
			t.Fatalf("role_permissions_unique_idx 定義 %q 應含 %s", indexdef, want)
		}
	}

	var roleID int64
	if err := db.QueryRow(`INSERT INTO roles (code, name) VALUES ('f1_regression', 'F1 迴歸') RETURNING id`).Scan(&roleID); err != nil {
		t.Fatalf("建立測試 role: %v", err)
	}
	insert := func(conditions any) error {
		_, err := db.Exec(
			`INSERT INTO role_permissions (role_id, resource, action, conditions) VALUES ($1, 'sales_order', 'read', $2)`,
			roleID, conditions,
		)
		return err
	}

	// 唯一性語意:NULL 視為 ''、jsonb 正規化後相同即重複。
	if err := insert(nil); err != nil {
		t.Fatalf("首列 conditions=NULL 應可寫入: %v", err)
	}
	if err := insert(nil); !isUniqueViolation(err) {
		t.Fatalf("同 role 第二列 conditions=NULL 必須被唯一索引擋下,got %v", err)
	}
	if err := insert(`{}`); err != nil {
		t.Fatalf("conditions='{}' 與 NULL 應視為不同(COALESCE 後 ''≠md5('{}')): %v", err)
	}
	if err := insert(`{"a": 1, "b": 2}`); err != nil {
		t.Fatalf("首列 conditions={\"a\":1,\"b\":2} 應可寫入: %v", err)
	}
	if err := insert(`{"b": 2, "a": 1}`); !isUniqueViolation(err) {
		t.Fatalf("jsonb 鍵序不同但語意相同必須視為重複,got %v", err)
	}
}

// TestIntegrationOpenFGAMigration F2 迴歸:OpenFGA schema 遷移、版本表獨立、可重複執行、
// 可 bootstrap,且缺表時必須失敗(不得靜默降級)。
func TestIntegrationOpenFGAMigration(t *testing.T) {
	ctx := context.Background()

	// 完整遷移庫:業務 schema + OpenFGA schema(順序同 cmd/migrate main)。
	dsn := testsupport.Postgres(t)
	if err := businessUp(t, dsn); err != nil {
		t.Fatalf("業務 migrate up: %v", err)
	}
	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()
	businessVersions := versionRows(t, db, businessGooseTable)

	if err := migrateOpenFGA(ctx, openfgaConfig(dsn)); err != nil {
		t.Fatalf("migrateOpenFGA 必須成功(F2:內嵌 OpenFGA 的 schema 曾從未建立): %v", err)
	}

	t.Run("OpenFGA 表齊全", func(t *testing.T) {
		var missing []string
		for _, table := range strings.Split(openfgaTables, ", ") {
			var exists bool
			if err := db.QueryRow(
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`,
				table,
			).Scan(&exists); err != nil {
				t.Fatalf("查表 %s: %v", table, err)
			}
			if !exists {
				missing = append(missing, table)
			}
		}
		if len(missing) > 0 {
			t.Fatalf("遷移後缺少 OpenFGA 表: %v(應由 migrateOpenFGA 建立)", missing)
		}
	})

	t.Run("兩版本表獨立且重複執行為 no-op", func(t *testing.T) {
		if !slices.Equal(businessVersions, versionRows(t, db, businessGooseTable)) {
			t.Fatalf("OpenFGA 遷移不得改動業務版本表 %s", businessGooseTable)
		}
		openfgaVersions := versionRows(t, db, openFGAGooseTable)
		if slices.Equal(businessVersions, openfgaVersions) {
			t.Fatalf("兩版本表內容不得相同(OpenFGA 必須有自己的版本表 %s)", openFGAGooseTable)
		}

		// 冪等:重複 `migrate up`(業務 + OpenFGA)不得報錯、不得新增版本列。
		if err := businessUp(t, dsn); err != nil {
			t.Fatalf("重複業務 migrate up 必須為 no-op(F2:共用版本表會因 OpenFGA 版本而拒絕): %v", err)
		}
		if err := migrateOpenFGA(ctx, openfgaConfig(dsn)); err != nil {
			t.Fatalf("重複 migrateOpenFGA 必須為 no-op: %v", err)
		}
		if !slices.Equal(businessVersions, versionRows(t, db, businessGooseTable)) {
			t.Fatalf("重複業務 migrate up 不得新增 %s 版本列", businessGooseTable)
		}
		if !slices.Equal(openfgaVersions, versionRows(t, db, openFGAGooseTable)) {
			t.Fatalf("重複 migrateOpenFGA 不得新增 %s 版本列", openFGAGooseTable)
		}
	})

	t.Run("NewPostgres bootstrap 成功", func(t *testing.T) {
		// NewPostgres 內含 EnsureDefaultModel → WriteAuthorizationModel;datastore 型別限額為 0
		// 時官方 server 會以 "exceeds the allowed limit of 0" 拒絕(F2 的另一半),故必然踩到。
		client, err := ofga.NewPostgres(ctx, dsn, "f2-integration-store")
		if err != nil {
			t.Fatalf("遷移後的資料庫必須能 bootstrap OpenFGA(F2:datastore 限額為 0 會拒絕寫 model): %v", err)
		}
		t.Cleanup(client.Close)
		if client.StoreID == "" || client.ModelID == "" {
			t.Fatalf("bootstrap 後應有 store/model,got store=%q model=%q", client.StoreID, client.ModelID)
		}
		if err := client.WriteTuple(ctx, "user:f2", "can_read", "ability:sales_order"); err != nil {
			t.Fatalf("寫入 tuple: %v", err)
		}
		if ok, err := client.Check(ctx, "user:f2", "can_read", "ability:sales_order"); err != nil || !ok {
			t.Fatalf("tuple 讀取應通過,got ok=%t err=%v", ok, err)
		}
	})

	t.Run("缺 OpenFGA 表時 bootstrap 必須失敗", func(t *testing.T) {
		// 另起一顆容器(其 salesorder 為全新空庫)只跑業務遷移:比同一庫 drop 表乾淨,
		// 不動已驗證狀態、不受子測試執行順序影響;遷移 00002 以名定址,故不能同庫開第二個庫名。
		dsnOnly := testsupport.Postgres(t)
		if err := businessUp(t, dsnOnly); err != nil {
			t.Fatalf("業務 migrate up: %v", err)
		}
		client, err := ofga.NewPostgres(ctx, dsnOnly, "f2-business-only-store")
		if err == nil {
			client.Close()
			t.Fatal("缺 OpenFGA datastore 表時 NewPostgres 必須失敗(否則即為 F2 的靜默降級)")
		}
		t.Logf("預期失敗(缺表): %v", err)
	})
}

// openfgaConfig 組出 cmd/migrate 會看到的設定(Enabled=true + 明確 DSN)。
func openfgaConfig(dsn string) *config.Config {
	return &config.Config{OpenFGA: config.OpenFGA{Enabled: true, DatabaseURL: dsn}}
}

// businessUp 以 cmd/migrate 相同路徑套用業務遷移:同 dialect、同 migrations 目錄、同版本表。
func businessUp(t *testing.T, dsn string) error {
	t.Helper()
	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	goose.SetTableName(businessGooseTable)
	goose.SetBaseFS(nil)
	return goose.RunContext(context.Background(), "up", db, migrationsDir)
}

// versionRows 讀版本表內容為排序後清單(格式 version:applied),供獨立性/冪等比對。
func versionRows(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`SELECT version_id, is_applied FROM %s ORDER BY version_id, is_applied`, table))
	if err != nil {
		t.Fatalf("讀版本表 %s: %v", table, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var version int64
		var applied bool
		if err := rows.Scan(&version, &applied); err != nil {
			t.Fatalf("讀版本表 %s: %v", table, err)
		}
		out = append(out, fmt.Sprintf("%d:%t", version, applied))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("讀版本表 %s: %v", table, err)
	}
	if len(out) == 0 {
		t.Fatalf("版本表 %s 不應為空", table)
	}
	return out
}

// openDB 開啟測試連線(呼叫端負責 Close)。
func openDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("開啟資料庫: %v", err)
	}
	return db
}

// isUniqueViolation 判斷是否為 PostgreSQL 唯一性違反(23505)。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
