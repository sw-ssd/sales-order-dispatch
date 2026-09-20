//go:build integration

// platform 域(00029)的遷移驗收。三個可觀測契約:
//
//	① 10 張表都在 platform schema(平台域的介面本體,後續 store 全數依賴);
//	② 6 個具名索引都存在、落在正確的表上,且 partial／unique 語意正確;
//	③ 業務連線角色 app_rw 對該 schema **零權限**(S9／§3.3 的硬邊界)—— 平台域的存取
//	   一律走 DATABASE_ADMIN_URL,業務路徑不得順手讀到帳務資料。
//
// ③以兩種彼此獨立的方式驗:系統目錄函式(has_schema_privilege/has_table_privilege/
// has_sequence_privilege)斷言「沒有任何權限」,再以 app_rw 真連線實測一次被拒(SQLSTATE
// 42501)。只驗前者會漏掉「函式說沒有、實際卻進得去」的落差;只驗後者則可能在查詢本身
// 打錯字時以「被拒」假綠,故另附對照組(同一條連線讀 public 業務表必須成功)。
package postgres_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// platformMigrationsDir 為業務遷移目錄(與 cmd/migrate 同一份檔案、同一張版本表)。
const platformMigrationsDir = "../../../../database/migrations"

// platformTables 為 00029 應建出的完整表集(平台域介面,缺一即後續 store 全滅)。
var platformTables = []string{
	"plans", "plan_prices", "features", "plan_entitlements",
	"subscriptions", "subscription_periods", "tenant_overrides",
	"events", "operators", "audit_logs",
}

// TestIntegrationPlatformSchema 驗證 platform schema 的 10 張表齊備,且業務角色 app_rw
// 對其零權限(實證)。
func TestIntegrationPlatformSchema(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = admin.Close() }()
	if err := goose.RunContext(t.Context(), "up", admin, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	assertPlatformTables(t, admin)
	assertPlatformIndexes(t, admin)

	// ②a 系統目錄函式:app_rw 對 schema／每一張表／每一個序列都不得有任何權限。
	// 表與序列以 pg_class 列舉,而非 information_schema:後者只顯示「當前角色有權限」的物件,
	// 一旦換成 app_rw 查就會讓「零授權」變成結構上恆真(見 rls_app_role_integration_test.go)。
	var schemaPriv bool
	if err := admin.QueryRow(
		`SELECT has_schema_privilege('app_rw', 'platform', 'USAGE,CREATE')`,
	).Scan(&schemaPriv); err != nil {
		t.Fatalf("查 app_rw 的 schema 權限: %v", err)
	}
	if schemaPriv {
		t.Fatal("app_rw 不得對 platform schema 有任何權限(USAGE 或 CREATE)")
	}

	// 表與序列分開列舉:`has_table_privilege` 對序列、`has_sequence_privilege` 對表都會直接
	// 回 SQLSTATE 42809(「不是該類物件」),故同一條 SELECT 不能對每列都呼叫兩個函式。
	seen := map[string]bool{}
	for _, kind := range []struct {
		relkind, fn, privs, what string
	}{
		{"r", "has_table_privilege", "SELECT,INSERT,UPDATE,DELETE,TRUNCATE,REFERENCES,TRIGGER", "表"},
		{"S", "has_sequence_privilege", "USAGE,SELECT,UPDATE", "序列"},
	} {
		rows, err := admin.Query(
			`SELECT c.relname, `+kind.fn+`('app_rw', c.oid, $1)
			   FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
			  WHERE n.nspname = 'platform' AND c.relkind = $2`, kind.privs, kind.relkind,
		)
		if err != nil {
			t.Fatalf("列舉 platform %s: %v", kind.what, err)
		}
		for rows.Next() {
			var name string
			var priv bool
			if err := rows.Scan(&name, &priv); err != nil {
				rows.Close()
				t.Fatalf("掃描 platform %s: %v", kind.what, err)
			}
			if kind.relkind == "r" {
				seen[name] = true
			}
			if priv {
				rows.Close()
				t.Fatalf("app_rw 不得對 platform.%s 這個%s有任何權限", name, kind.what)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			t.Fatalf("列舉 platform %s: %v", kind.what, err)
		}
		rows.Close()
	}
	for _, table := range platformTables {
		if !seen[table] {
			t.Fatalf("pg_class 列舉應涵蓋 platform.%s,got %v", table, seen)
		}
	}

	// ②b 實測被拒:以業務角色真連線,讀 platform 表必須 permission denied。
	app, err := sql.Open("pgx", testsupport.AppRoleDSN(t, dsn))
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	defer func() { _ = app.Close() }()
	var n int
	assertDenied(t, app.QueryRow(`SELECT count(*) FROM platform.plans`).Scan(&n), "app_rw 讀 platform.plans")
	// 對照組:同一條連線讀業務表必須成功(00022 有授權)—— 否則上面的「被拒」可能只是連線壞掉。
	if err := app.QueryRow(`SELECT count(*) FROM companies`).Scan(&n); err != nil {
		t.Fatalf("對照組失敗:app_rw 應可讀 public 業務表 companies,got %v", err)
	}
}

// TestIntegrationPlatformSchemaDown 驗證 00029 的回滾完整:`down-to 0` 之後 platform schema
// 必須整個消失(schema 內 10 張表與其索引、序列一併帶走,不留孤兒物件),
// 且重新 up 能再建回 —— 回滾鏈不得只把版本列往回寫而把物件留在庫上。
func TestIntegrationPlatformSchemaDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = admin.Close() }()
	if err := goose.RunContext(t.Context(), "up", admin, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
	assertPlatformTables(t, admin)

	if err := goose.RunContext(t.Context(), "down-to", admin, platformMigrationsDir, "0"); err != nil {
		t.Fatalf("goose down-to 0: %v", err)
	}
	var objects int
	if err := admin.QueryRow(
		`SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		  WHERE n.nspname = 'platform'`,
	).Scan(&objects); err != nil {
		t.Fatalf("查回滾後的 platform 物件: %v", err)
	}
	if objects != 0 {
		t.Fatalf("down-to 0 後 platform schema 內的物件應全數消失(含序列與索引),仍剩 %d 個", objects)
	}
	var schemas int
	if err := admin.QueryRow(`SELECT count(*) FROM pg_namespace WHERE nspname = 'platform'`).Scan(&schemas); err != nil {
		t.Fatalf("查回滾後的 platform schema: %v", err)
	}
	if schemas != 0 {
		t.Fatal("down-to 0 後 platform schema 本身也應消失")
	}

	// 回滾鏈可重複:回滾後重新 up 必須完整回來。
	if err := goose.RunContext(t.Context(), "up", admin, platformMigrationsDir); err != nil {
		t.Fatalf("回滾後重新 up: %v", err)
	}
	assertPlatformTables(t, admin)
}

// platformIndexes 為 00029 的索引契約:名稱、所在表、以及 partial／unique 語意。
// partial index 少了 WHERE 就不是同一個索引 —— 例如 subscriptions 少了 `status <> 'cancelled'`,
// 同一租戶只要有一筆取消過的訂閱就再也簽不了新約(而錯誤要到上線後才出現),故一併釘住。
var platformIndexes = []struct {
	name, table     string
	partial, unique bool
}{
	{"plan_prices_plan_effective_idx", "plan_prices", false, false},
	{"subscriptions_active_company_unique", "subscriptions", true, true},
	{"periods_provider_ref_unique", "subscription_periods", true, true},
	{"tenant_overrides_active_unique", "tenant_overrides", true, true},
	{"events_undispatched_idx", "events", true, false},
	{"platform_audit_created_idx", "audit_logs", false, false},
}

// assertPlatformIndexes 斷言每個索引都落在正確的表上,且 partial／unique 語意正確。
func assertPlatformIndexes(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, idx := range platformIndexes {
		var table string
		var partial, unique bool
		err := db.QueryRow(
			`SELECT t.relname, pg_get_expr(i.indpred, i.indrelid) IS NOT NULL, i.indisunique
			   FROM pg_index i
			   JOIN pg_class ix ON ix.oid = i.indexrelid
			   JOIN pg_class t  ON t.oid  = i.indrelid
			   JOIN pg_namespace n ON n.oid = t.relnamespace
			  WHERE n.nspname = 'platform' AND ix.relname = $1`, idx.name,
		).Scan(&table, &partial, &unique)
		if err != nil {
			t.Fatalf("索引 %s 不存在(或查詢失敗): %v", idx.name, err)
		}
		if table != idx.table {
			t.Fatalf("索引 %s 應在 platform.%s,got platform.%s", idx.name, idx.table, table)
		}
		if partial != idx.partial {
			t.Fatalf("索引 %s 的 partial 語意應為 %v,got %v", idx.name, idx.partial, partial)
		}
		if unique != idx.unique {
			t.Fatalf("索引 %s 的 unique 語意應為 %v,got %v", idx.name, idx.unique, unique)
		}
	}
}

// assertPlatformTables 斷言 10 張表都在 platform schema。
func assertPlatformTables(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, table := range platformTables {
		var n int
		if err := db.QueryRow(
			`SELECT count(*) FROM information_schema.tables
			  WHERE table_schema = 'platform' AND table_name = $1`, table,
		).Scan(&n); err != nil {
			t.Fatalf("查 %s: %v", table, err)
		}
		if n != 1 {
			t.Fatalf("platform.%s 不存在", table)
		}
	}
}

// assertDenied 斷言操作因權限不足被拒(SQLSTATE 42501),而非因其他原因失敗。
func assertDenied(t *testing.T, err error, what string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Fatalf("%s 必須以權限不足(42501)被拒,got %v", what, err)
	}
}
