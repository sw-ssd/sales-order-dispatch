//go:build integration

package services

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// businessTables 為 00022 明確授權的業務表白名單（與 T3 的 policy 覆蓋清單一致）：
// 授權必須恰好落在這 18 張，多一張即為權限外洩（如內嵌 OpenFGA 的授權表）。
var businessTables = map[string]bool{
	"companies": true, "departments": true, "users": true, "roles": true,
	"role_permissions": true, "audit_logs": true, "metadicts": true,
	"customers": true, "customer_counters": true, "customer_addresses": true,
	"customer_contacts": true, "warehouses": true, "routes": true,
	"processing_specs": true, "product_categories": true, "products": true,
	"product_units": true, "product_processing_specs": true,
}

// TestIntegrationAppRolePrivileges 驗證業務角色是非 owner、且對業務表有 DML 權限：
// RLS 只在「非 owner」連線上才擋得住（owner 需 FORCE，而 FORCE 連帶要求回填走系統 scope）。
// 另驗證授權**不外溢**：app_rw 只碰得到業務表（最小權限）。
func TestIntegrationAppRolePrivileges(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	appDSN := testsupport.AppRoleDSN(t, adminDSN)

	db, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	defer db.Close()

	var user string
	if err := db.QueryRow(`SELECT current_user`).Scan(&user); err != nil {
		t.Fatalf("查 current_user: %v", err)
	}
	if user != "app_rw" {
		t.Fatalf("業務連線應為 app_rw,got %q", user)
	}

	var isSuper, bypass bool
	if err := db.QueryRow(
		`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`,
	).Scan(&isSuper, &bypass); err != nil {
		t.Fatalf("查角色屬性: %v", err)
	}
	if isSuper || bypass {
		t.Fatalf("app_rw 不得為 superuser 或 BYPASSRLS(rolsuper=%v rolbypassrls=%v)", isSuper, bypass)
	}

	var owners int
	if err := db.QueryRow(
		`SELECT count(*) FROM pg_class c JOIN pg_roles r ON r.oid = c.relowner
		  WHERE r.rolname = current_user AND c.relname IN ('customers','products','users')`,
	).Scan(&owners); err != nil {
		t.Fatalf("查 table owner: %v", err)
	}
	if owners != 0 {
		t.Fatalf("app_rw 不得是業務表 owner（owner 會繞過 RLS），got %d 張", owners)
	}

	// 逐表斷言授權量:業務表必須四項 DML 齊備（少了任一張,T3 之後的 RLS policy 會變成
	// permission denied 而非可用的隔離）;非業務表必須**零**授權（授權外溢即最小權限破口）。
	// 以實際存在的 public 表列舉而非硬編清單 → 容器內缺 OpenFGA 的表時不會假綠。
	granted := grantedTables(t, db)
	for table := range businessTables {
		if granted[table] != 4 {
			t.Errorf("業務表 %s 應有 SELECT/INSERT/UPDATE/DELETE 四項授權,got %d", table, granted[table])
		}
	}
	for table, n := range granted {
		if n > 0 && !businessTables[table] {
			t.Errorf("app_rw 對非業務表 %s 有 %d 項授權（00022 只應授權業務表）", table, n)
		}
	}

	// 00022 之後由 owner 在 public 建的非業務物件:內嵌 OpenFGA 的 migration 就是在同一 schema
	// 建其授權表(tuple/authorization_model/…)。故本測試自行建立同名物件 —— 不依賴 OpenFGA 的
	// 表是否存在(缺表會讓斷言假綠),才真能測出 ALL TABLES／default privileges 的外溢。
	admin := openRawDB(t, adminDSN)
	if _, err := admin.Exec(`CREATE TABLE tuple (id bigint PRIMARY KEY, value text)`); err != nil {
		t.Fatalf("owner 建非業務表: %v", err)
	}
	if _, err := admin.Exec(`CREATE SEQUENCE openfga_seq`); err != nil {
		t.Fatalf("owner 建非業務序列: %v", err)
	}
	var nonBusiness int
	if err := db.QueryRow(
		`SELECT count(*) FROM information_schema.role_table_grants
		  WHERE grantee = 'app_rw' AND table_name = 'tuple'`,
	).Scan(&nonBusiness); err != nil {
		t.Fatalf("查非業務表授權: %v", err)
	}
	if nonBusiness != 0 {
		t.Errorf("app_rw 不應對 00022 之後新建的非業務表 tuple 取得授權,got %d 項", nonBusiness)
	}
	_, err = db.Exec(`INSERT INTO tuple (id, value) VALUES (1, 'x')`)
	assertDenied(t, err, "寫入非業務表 tuple")
	var seq int64
	assertDenied(t, db.QueryRow(`SELECT nextval('openfga_seq')`).Scan(&seq), "取用非業務序列 openfga_seq")
}

// grantedTables 回傳 public 每張實際存在的表上,app_rw 取得的 DML 授權數（0 表無授權）。
// 以 pg_class + has_table_privilege 列舉,而非 information_schema:後者只顯示**當前角色有權限**
// 的物件(app_rw 無權的表直接消失),用它取差集會讓「非業務表零授權」變成結構上恆真而假綠。
func grantedTables(t *testing.T, db *sql.DB) map[string]int {
	t.Helper()
	rows, err := db.Query(
		`SELECT c.relname,
		        (has_table_privilege('app_rw', c.oid, 'SELECT')::int
		       + has_table_privilege('app_rw', c.oid, 'INSERT')::int
		       + has_table_privilege('app_rw', c.oid, 'UPDATE')::int
		       + has_table_privilege('app_rw', c.oid, 'DELETE')::int)
		   FROM pg_class c
		   JOIN pg_namespace n ON n.oid = c.relnamespace
		  WHERE n.nspname = 'public' AND c.relkind = 'r'`,
	)
	if err != nil {
		t.Fatalf("列舉授權: %v", err)
	}
	defer rows.Close()
	granted := map[string]int{}
	for rows.Next() {
		var table string
		var n int
		if err := rows.Scan(&table, &n); err != nil {
			t.Fatalf("掃描授權: %v", err)
		}
		granted[table] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("列舉授權: %v", err)
	}
	return granted
}

// assertDenied 斷言操作因權限不足被拒（SQLSTATE 42501），而非因其他原因失敗。
func assertDenied(t *testing.T, err error, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("app_rw 不應能%s（00022 不得外溢授權）", what)
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Fatalf("%s 應因權限不足被拒(42501),got %v", what, err)
	}
}
