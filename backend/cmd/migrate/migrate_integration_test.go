//go:build integration

// cmd/migrate 的真 PostgreSQL 整合測試。以拋棄式容器(見 internal/testsupport)驗證兩個
// 曾無自動化守門的缺陷:
//
//	F1 00005 的表層 UNIQUE 表達式:全新資料庫 `migrate up` 直接失敗(48b3d34 修復)。
//	F2 內嵌 OpenFGA 的 schema 未建立 + datastore 型別限額為 0 + 缺表時靜默降級(65826e4 修復)。
//
// 並涵蓋收尾波(P1 複審缺口):00018 冪等索引收斂、00002 去硬編庫名、server 缺 OpenFGA
// schema 的 fail-fast 與退路、OPENFGA_ENABLED=false 跳過、OPENFGA_DATABASE_URL 優先序、
// 不可連線 DSN 的 Timeout 有界性。
//
// 業務遷移走與 cmd/migrate 相同的 goose 路徑(同 dialect、同 migrations 目錄、同版本表),
// OpenFGA 遷移直接呼叫 main 所用的 migrateOpenFGA,確保測的是實際程式碼而非複製品;
// server 行為以 `go build` 出的 cmd/server 子行程驗證(log.Fatalf 的退出碼無法 in-process 觀察)。
//
// 除了「非預設庫名」與「不可連線 DSN」兩條,其餘都以「專屬容器 + 全新空庫」為前提
// (缺表/缺版號/索引不存在…),故各自先呼叫 testsupport.RequiresContainer:在
// INTEGRATION_TEST_DSN 覆寫模式(共用同一個既有的庫)下自動 skip,不會假失敗。
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

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
	testsupport.RequiresContainer(t)
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
	testsupport.RequiresContainer(t)
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
		// 不動已驗證狀態、不受子測試執行順序影響(同容器開第二個庫見 testsupport.CreateDatabase)。
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

// TestIntegrationUniqueIndexForwardMigration F1 收尾:00018 是所有環境的冪等收斂入口。
// 「既有庫」子測試模擬:庫已記錄 version 5(故 00005 永不重跑)但缺唯一索引 ——
// 這正是 00018 存在的理由;索引已在的 fresh 庫重複 up 必須是 no-op。
func TestIntegrationUniqueIndexForwardMigration(t *testing.T) {
	testsupport.RequiresContainer(t)
	t.Run("既有庫缺索引 → 00018 補回", func(t *testing.T) {
		dsn := testsupport.Postgres(t)
		db := openDB(t, dsn)
		defer func() { _ = db.Close() }()

		// 前置:套用到 17(含 00005)後手動移除索引,即「已記錄 version 5 但無索引」的既有庫。
		if err := businessUpTo(t, dsn, 17); err != nil {
			t.Fatalf("升級至 17: %v", err)
		}
		if _, err := db.Exec(`DROP INDEX IF EXISTS role_permissions_unique_idx`); err != nil {
			t.Fatalf("移除索引以模擬既有庫: %v", err)
		}
		if indexExists(t, db) {
			t.Fatal("前置條件失敗:索引應已被移除")
		}
		versions := versionRows(t, db, businessGooseTable)
		if !slices.Contains(versions, "5:true") || slices.Contains(versions, "18:true") {
			t.Fatalf("前置條件失敗:應已記錄 version 5 且尚未記錄 18,got %v", versions)
		}

		// 收斂:00018 必須把缺掉的唯一索引補回,並留下版號。
		if err := businessUpTo(t, dsn, 18); err != nil {
			t.Fatalf("00018 應為少索引的既有庫補回唯一索引: %v", err)
		}
		assertUniqueIndexShape(t, db)
		if got := versionRows(t, db, businessGooseTable); !slices.Contains(got, "18:true") {
			t.Fatalf("goose_db_version 應記錄 18,got %v", got)
		}
	})

	t.Run("fresh 庫重複 up 為 no-op", func(t *testing.T) {
		dsn := testsupport.Postgres(t)
		db := openDB(t, dsn)
		defer func() { _ = db.Close() }()

		if err := businessUp(t, dsn); err != nil {
			t.Fatalf("全新資料庫 migrate up: %v", err)
		}
		assertUniqueIndexShape(t, db)
		before := versionRows(t, db, businessGooseTable)
		if err := businessUp(t, dsn); err != nil {
			t.Fatalf("第二次 migrate up 必須為 no-op(00018 的 IF NOT EXISTS): %v", err)
		}
		assertUniqueIndexShape(t, db)
		if after := versionRows(t, db, businessGooseTable); !slices.Equal(before, after) {
			t.Fatalf("重複 migrate up 不得改動版本表:before %v after %v", before, after)
		}
	})
}

// TestIntegrationMigrateOnNonDefaultDatabaseName 00002 收尾:00002 曾以硬編名
// (`ALTER DATABASE salesorder SET row_security = on`)定址資料庫,schema 因此綁死庫名。
// 以非 salesorder 的庫驗證:遷移成功且 row_security 落在本庫;該實例內不得存在名為
// salesorder 的庫,否則硬編名會被誤改到別庫而測不出來。
func TestIntegrationMigrateOnNonDefaultDatabaseName(t *testing.T) {
	const dbName = "salesorder_alt"
	dsn := testsupport.PostgresNamed(t, dbName)
	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()

	var current string
	if err := db.QueryRow(`SELECT current_database()`).Scan(&current); err != nil {
		t.Fatalf("查 current_database(): %v", err)
	}
	if current != dbName {
		t.Fatalf("測試庫名應為 %s,got %s", dbName, current)
	}
	var defaultDBs int
	if err := db.QueryRow(`SELECT count(*) FROM pg_database WHERE datname = 'salesorder'`).Scan(&defaultDBs); err != nil {
		t.Fatalf("查 pg_database: %v", err)
	}
	if defaultDBs != 0 {
		t.Fatal("測試實例不應存在 salesorder 庫(否則 00002 的硬編名會被掩蓋)")
	}

	if err := businessUp(t, dsn); err != nil {
		t.Fatalf("非預設庫名(%s)的 migrate up 必須成功(00002 硬編名會以 SQLSTATE 3D000 database \"salesorder\" does not exist 失敗): %v", dbName, err)
	}
	assertUniqueIndexShape(t, db)
	if got := versionRows(t, db, businessGooseTable); !slices.Contains(got, "18:true") {
		t.Fatalf("goose_db_version 應記錄 18,got %v", got)
	}

	// row_security 必須設在「本庫」(current_database()),而非任何硬編名。
	var setconfig string
	if err := db.QueryRow(
		`SELECT COALESCE(array_to_string(s.setconfig, ','), '') FROM pg_db_role_setting s JOIN pg_database d ON d.oid = s.setdatabase WHERE d.datname = current_database()`,
	).Scan(&setconfig); err != nil {
		t.Fatalf("讀 pg_db_role_setting(本庫): %v", err)
	}
	if !strings.Contains(setconfig, "row_security=on") {
		t.Fatalf("%s 的 row_security 應為 on,got %q", dbName, setconfig)
	}
}

// TestIntegrationServerOpenFGAFailFast F2 收尾(P1 複審缺口):只跑業務遷移的庫(無 OpenFGA
// schema)起 cmd/server 時,OPENFGA_ENABLED=true 必須拒絕啟動並印出可行動訊息;
// OPENFGA_ENABLED=false 則必須能啟動(文件化退路)。以真子行程驗證,涵蓋 domains.go 的
// fail-fast 行為(單元測試無法觀察 log.Fatalf 的退出碼)。
func TestIntegrationServerOpenFGAFailFast(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := businessUp(t, dsn); err != nil {
		t.Fatalf("業務 migrate up(不跑 OpenFGA 遷移): %v", err)
	}
	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()
	if present := openFGATablesPresent(t, db); len(present) != 0 {
		t.Fatalf("前置條件失敗:此庫不應有 OpenFGA 表,卻有 %v", present)
	}
	bin := buildServerBinary(t)

	t.Run("OPENFGA_ENABLED=true 且缺 schema → 拒絕啟動", func(t *testing.T) {
		out, err := runServerToExit(t, bin, serverEnv(dsn, "true"), 60*time.Second)
		if err == nil {
			t.Fatalf("缺 OpenFGA schema 時 server 必須以非零碼退出(F2 靜默降級迴歸);輸出:\n%s", out)
		}
		for _, want := range []string{"cmd/migrate up", "OPENFGA_ENABLED=false"} {
			if !strings.Contains(out, want) {
				t.Fatalf("拒絕啟動的訊息應含可行動指引 %q;輸出:\n%s", want, out)
			}
		}
	})

	t.Run("OPENFGA_ENABLED=false → 可啟動", func(t *testing.T) {
		out := runServerUntilListening(t, bin, serverEnv(dsn, "false"), 60*time.Second)
		if !strings.Contains(out, "listening on") {
			t.Fatalf("OPENFGA_ENABLED=false 時 server 應開始 listening;輸出:\n%s", out)
		}
	})
}

// TestIntegrationOpenFGADisabledCreatesNoSchema F2 收尾:`OPENFGA_ENABLED=false` 時
// cmd/migrate up 不得建立任何 OpenFGA 表(含專屬版本表);走 main.go 同一條呼叫路徑。
func TestIntegrationOpenFGADisabledCreatesNoSchema(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := businessUp(t, dsn); err != nil {
		t.Fatalf("業務 migrate up: %v", err)
	}
	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()

	// Enabled 零值即 OPENFGA_ENABLED=false;OPENFGA_DATABASE_URL 未設(沿用業務 DSN)。
	cfg := &config.Config{Database: config.Database{DatabaseURL: dsn}}
	if err := migrateOpenFGA(t.Context(), cfg); err != nil {
		t.Fatalf("OPENFGA_ENABLED=false 時 migrateOpenFGA 必須是 no-op: %v", err)
	}
	if present := openFGATablesPresent(t, db); len(present) != 0 {
		t.Fatalf("OPENFGA_ENABLED=false 不得建立 OpenFGA schema,卻有 %v", present)
	}
}

// TestIntegrationOpenFGADatabaseURLPrecedence F2 收尾:OPENFGA_DATABASE_URL 有值時,
// OpenFGA schema 只落在該庫,業務庫完全不動(同一顆 server 上的第二個庫,不另起容器)。
func TestIntegrationOpenFGADatabaseURLPrecedence(t *testing.T) {
	testsupport.RequiresContainer(t)
	businessDSN := testsupport.Postgres(t)
	if err := businessUp(t, businessDSN); err != nil {
		t.Fatalf("業務 migrate up: %v", err)
	}
	openfgaDSN := testsupport.CreateDatabase(t, businessDSN, "openfga_alt")
	business := openDB(t, businessDSN)
	defer func() { _ = business.Close() }()
	other := openDB(t, openfgaDSN)
	defer func() { _ = other.Close() }()

	cfg := &config.Config{
		Database: config.Database{DatabaseURL: businessDSN},
		OpenFGA:  config.OpenFGA{Enabled: true, DatabaseURL: openfgaDSN},
	}
	if err := migrateOpenFGA(t.Context(), cfg); err != nil {
		t.Fatalf("以 OPENFGA_DATABASE_URL 指向的庫跑 OpenFGA 遷移: %v", err)
	}

	if got := openFGATablesPresent(t, other); len(got) != len(openfgaSchemaTables) {
		t.Fatalf("OpenFGA schema 應完整落在 OPENFGA_DATABASE_URL 指定的庫,got %v", got)
	}
	if present := openFGATablesPresent(t, business); len(present) != 0 {
		t.Fatalf("業務庫不得被寫入 OpenFGA schema,卻有 %v", present)
	}
}

// TestIntegrationOpenFGAMigrationTimeoutBounded F2 收尾:不可連線的 DSN 必須在有界時間內回錯。
// Timeout/PingTimeout 被改回 0 時 backoff 的 MaxElapsedTime=0 表示無限重試,本測試會以
// 90 秒上限抓到該迴歸(官方常數:DefaultDatastorePingRetryMaxElapsedTime = 1m)。
func TestIntegrationOpenFGAMigrationTimeoutBounded(t *testing.T) {
	const unreachable = "postgres://postgres:postgres@127.0.0.1:1/nope?sslmode=disable"
	cfg := &config.Config{OpenFGA: config.OpenFGA{Enabled: true, DatabaseURL: unreachable}}

	start := time.Now()
	done := make(chan error, 1)
	go func() { done <- migrateOpenFGA(t.Context(), cfg) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("不可連線的 DSN 不得回 nil(遷移必須失敗)")
		}
		elapsed := time.Since(start)
		if elapsed > 90*time.Second {
			t.Fatalf("遷移未在 90 秒內放棄(實測 %v)", elapsed)
		}
		t.Logf("不可連線 DSN 於 %v 內回錯: %v", elapsed, err)
	case <-time.After(90 * time.Second):
		t.Fatal("不可連線的 DSN 超過 90 秒仍未回錯:Timeout/PingTimeout 疑似被改回 0(無限重試)")
	}
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

// businessUpTo 以 goose 版本化 API 套用業務遷移到指定版本,供「既有庫」情境模擬
// (businessUp 為全量 up;兩者共用同一 dialect/目錄/版本表)。
func businessUpTo(t *testing.T, dsn string, version int64) error {
	t.Helper()
	db := openDB(t, dsn)
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	goose.SetTableName(businessGooseTable)
	goose.SetBaseFS(nil)
	return goose.UpToContext(t.Context(), db, migrationsDir, version)
}

// indexExists 判斷 role_permissions_unique_idx 是否存在。
func indexExists(t *testing.T, db *sql.DB) bool {
	t.Helper()
	var exists bool
	if err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'role_permissions' AND indexname = 'role_permissions_unique_idx')`,
	).Scan(&exists); err != nil {
		t.Fatalf("查 role_permissions_unique_idx: %v", err)
	}
	return exists
}

// assertUniqueIndexShape 斷言 role_permissions_unique_idx 存在,且以 COALESCE(md5(...)) 表達
// 「conditions 為 NULL 視為空字串」的去重語意。
func assertUniqueIndexShape(t *testing.T, db *sql.DB) {
	t.Helper()
	var indexdef string
	if err := db.QueryRow(
		`SELECT indexdef FROM pg_indexes WHERE tablename = 'role_permissions' AND indexname = 'role_permissions_unique_idx'`,
	).Scan(&indexdef); err != nil {
		t.Fatalf("查 role_permissions_unique_idx: %v(少索引的既有庫未被收斂)", err)
	}
	for _, want := range []string{"UNIQUE", "COALESCE", "md5", "conditions"} {
		if !strings.Contains(indexdef, want) {
			t.Fatalf("role_permissions_unique_idx 定義 %q 應含 %s", indexdef, want)
		}
	}
}

// openfgaSchemaTables 為 migrateOpenFGA 應建出的完整集合:OpenFGA 核心表 + 專屬版本表。
var openfgaSchemaTables = append(strings.Split(openfgaTables, ", "), openFGAGooseTable)

// openFGATablesPresent 回傳 db 內已存在的 OpenFGA 表(含專屬版本表),供存在/不存在斷言。
func openFGATablesPresent(t *testing.T, db *sql.DB) []string {
	t.Helper()
	var present []string
	for _, table := range openfgaSchemaTables {
		var exists bool
		if err := db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`,
			table,
		).Scan(&exists); err != nil {
			t.Fatalf("查表 %s: %v", table, err)
		}
		if exists {
			present = append(present, table)
		}
	}
	return present
}

// buildServerBinary 建置 cmd/server 為暫存二進位,供子行程測試觀察啟動/退出行為
// (log.Fatalf 的退出碼與輸出無法以 in-process 測試觀察)。
func buildServerBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "server")
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", bin, "../../cmd/server")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("建置 cmd/server: %v\n%s", err, out)
	}
	return bin
}

// serverEnv 組出 server 子行程環境:DATABASE_URL 指向測試庫、API_ADDR 綁 ephemeral port
// (避免與本機 dev server 撞埠)、VALKEY_ADDR 指向不可連線位址(本測試只關心 OpenFGA 閘門,
// 其餘 domain 走既有降級路徑)。ENV 固定 development,不受外部 shell 影響。
func serverEnv(dsn, openfgaEnabled string) []string {
	return append(os.Environ(),
		"DATABASE_URL="+dsn,
		"OPENFGA_ENABLED="+openfgaEnabled,
		"API_ADDR=127.0.0.1:0",
		"VALKEY_ADDR=127.0.0.1:1",
		"ENV=development",
	)
}

// runServerToExit 起 server 並等其自行結束,回傳合併輸出與行程錯誤(exit code 0 → nil)。
func runServerToExit(t *testing.T, bin string, env []string, timeout time.Duration) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("server 未在 %v 內結束;輸出:\n%s", timeout, out)
	}
	return string(out), err
}

// runServerUntilListening 起 server,持續輪詢其輸出直到出現 listening,然後終止;
// 行程提早退出即失敗。回傳結束前收集到的輸出。
func runServerUntilListening(t *testing.T, bin string, env []string, timeout time.Duration) string {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "server.log")
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("建立 server 輸出檔: %v", err)
	}
	defer func() { _ = f.Close() }()

	cmd := exec.CommandContext(t.Context(), bin)
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = f, f
	if err := cmd.Start(); err != nil {
		t.Fatalf("啟動 server 子行程: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	defer func() {
		_ = cmd.Process.Kill()
		<-done
	}()

	for deadline := time.Now().Add(timeout); time.Now().Before(deadline); {
		out, _ := os.ReadFile(logPath)
		if strings.Contains(string(out), "listening on") {
			return string(out)
		}
		select {
		case err := <-done:
			t.Fatalf("server 未 listening 即退出(%v);輸出:\n%s", err, out)
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
	out, _ := os.ReadFile(logPath)
	t.Fatalf("server 未在 %v 內 listening;輸出:\n%s", timeout, out)
	return ""
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
