//go:build integration

// seed 在 RLS 全開下的守門探針(T10):FORCE RLS 也約束 table owner,而 seed 是系統級維運
// (roles／role_permissions／developer 帳號),故必須在**明確的系統範圍交易**內執行。
//
// 本測試釘住兩件事:
//
//	① 少了系統範圍 → 業務連線(app_rw,與生產的 owner 同屬非 superuser)被 WITH CHECK 擋下
//	   (SQLSTATE 42501)—— 這正是 seed 一定要走 dbtenant.SystemScopeTx 的原因;
//	② 走 SystemScopeTx → 三個 seed 函式完成、可重複執行(冪等),且列數由 admin 真值確認。
//
// 為何用 app_rw 而不是容器的 admin:admin 是 superuser,而 PG 的 superuser 永遠繞過 RLS
// (FORCE 亦然)→ 用它跑 seed 不管有沒有修都會成功,測不出東西。
package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// seedTestMigrationsDir 相對套件目錄(go test 以套件目錄為 cwd),與 cmd/migrate 同路徑。
const seedTestMigrationsDir = "../../database/migrations"

// seedTestGooseTable 為業務遷移版本表(cmd/migrate 未改動 goose 預設值)。
const seedTestGooseTable = "goose_db_version"

// TestIntegrationSeedUnderRLS 驗「seed 在 RLS 全開後仍能完成」及「不走系統範圍就會壞」。
func TestIntegrationSeedUnderRLS(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateForSeed(t, adminDSN)

	admin := openSeedAdmin(t, adminDSN)
	// 非 superuser 的 owner 就是生產的處境:FORCE RLS 讓他一樣受 policy 約束(00025/00028)。
	app := openSeedAppRole(t, adminDSN)
	client := dbtenant.NewClient(app)
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()

	// developer 帳號的錨點(SeedDeveloper 需要既有公司;無公司則靜默略過)。
	var companyID int
	if err := admin.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ('seed 測試公司', 'SEED-T', 'active') RETURNING id`).
		Scan(&companyID); err != nil {
		t.Fatalf("建公司錨點: %v", err)
	}

	t.Run("未在系統範圍內 → 被 WITH CHECK 擋下(seed 若不修就是這個下場)", func(t *testing.T) {
		err := SeedBuiltinRoles(ctx, client)
		if !isSeedRLSViolation(err) {
			t.Fatalf("未帶 scope 寫 roles 必須是 RLS 違反(42501),got %v", err)
		}
		if n := seedCount(t, admin, `SELECT count(*) FROM roles`); n != 0 {
			t.Fatalf("被擋下的 seed 不得落地任何角色,got %d 列", n)
		}
	})

	t.Run("SystemScopeTx → 三個 seed 函式完成且冪等", func(t *testing.T) {
		run := func() error {
			return dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
				if err := SeedBuiltinRoles(ctx, tx.Client()); err != nil {
					return err
				}
				if err := SeedBuiltinRolePermissions(ctx, tx.Client()); err != nil {
					return err
				}
				return SeedDeveloper(ctx, tx.Client(), "development", firstCompanyID(ctx, tx.Client()))
			})
		}
		if err := run(); err != nil {
			t.Fatalf("系統範圍內的 seed 必須成功: %v", err)
		}
		roles := seedCount(t, admin, `SELECT count(*) FROM roles`)
		perms := seedCount(t, admin, `SELECT count(*) FROM role_permissions`)
		if roles != 7 {
			t.Fatalf("應建立 7 個內建角色,got %d", roles)
		}
		if perms == 0 {
			t.Fatal("role_permissions 應有種子規則")
		}
		if n := seedCount(t, admin, `SELECT count(*) FROM users WHERE email = 'developer@local.dev'`); n != 1 {
			t.Fatalf("development 環境應建立 developer 帳號,got %d 列", n)
		}

		// 冪等:再跑一次不得增加任何列(Taskfile 的 `task seed` 會反覆執行)。
		if err := run(); err != nil {
			t.Fatalf("第二次 seed 必須成功(冪等): %v", err)
		}
		if got := seedCount(t, admin, `SELECT count(*) FROM roles`); got != roles {
			t.Errorf("重複 seed 不得新增角色:%d → %d", roles, got)
		}
		if got := seedCount(t, admin, `SELECT count(*) FROM role_permissions`); got != perms {
			t.Errorf("重複 seed 不得新增權限:%d → %d", perms, got)
		}
	})
}

// migrateForSeed 以 cmd/migrate 相同路徑套用業務遷移(同 dialect、同目錄、同版本表)。
func migrateForSeed(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	goose.SetTableName(seedTestGooseTable)
	goose.SetBaseFS(nil)
	if err := goose.RunContext(context.Background(), "up", db, seedTestMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}

// openSeedAdmin 以 admin(owner／superuser)連線取真值;superuser 不受 RLS 影響。
func openSeedAdmin(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("admin 連線: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// openSeedAppRole 以 app_rw(00022 的 NOBYPASSRLS 業務角色)連線。
//
// **連線池不設上限 1**:seed 與未來的系統範圍路徑可能同時持有請求交易與 SystemScopeTx 兩條
// 連線,池 1 會互鎖死結(T9 前例);上限 4 是明確的「≥2」。
func openSeedAppRole(t *testing.T, adminDSN string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", testsupport.AppRoleDSN(t, adminDSN))
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	db.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// seedCount 以 admin 連線取單一純量真值。
func seedCount(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("查詢 %q: %v", query, err)
	}
	return n
}

// isSeedRLSViolation 判斷錯誤鏈中是否有 PostgreSQL 的 42501(insufficient_privilege 家族中
// RLS 違反所用的碼)。
func isSeedRLSViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42501"
}
