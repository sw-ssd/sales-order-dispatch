//go:build integration

package services

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// businessTables 為明確授權的業務表白名單（00022 的 18 張 ＋ 00031 的訂單四表 ＋
// 00033 的 customer_products ＋ 00035 的 file_assets ＋ 00037 的列印兩表 ＋
// 00039 的退貨兩表 ＋ 00041 的通知四表，共 32 張；各表授權由其 migration 明示列舉）。
// 授權必須恰好落在這些表，多一張即為權限外洩（如內嵌 OpenFGA 的授權表）。
var businessTables = map[string]bool{
	"companies": true, "departments": true, "users": true, "roles": true,
	"role_permissions": true, "audit_logs": true, "metadicts": true,
	"customers": true, "customer_counters": true, "customer_addresses": true,
	"customer_contacts": true, "warehouses": true, "routes": true,
	"processing_specs": true, "product_categories": true, "products": true,
	"product_units": true, "product_processing_specs": true,
	"sales_orders": true, "sales_order_items": true, "sales_order_events": true,
	"order_counters": true, "customer_products": true, "file_assets": true,
	"print_logs": true, "print_previews": true,
	"return_requests": true, "return_request_items": true,
	"notification_templates": true, "notifications": true,
	"user_devices": true, "promo_tags": true,
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

	// 逐表斷言授權量:業務表必須 DML 齊備（少了任一張,T3 之後的 RLS policy 會變成
	// permission denied 而非可用的隔離）;非業務表必須**零**授權（授權外溢即最小權限破口）。
	// 僅追加例外:記錄類表只授 SELECT/INSERT(見 00031/00037);UPDATE/DELETE 刻意不授。
	// 通知形狀例外:範本/通知無 DELETE(見 00041),故 want=3。
	// 以實際存在的 public 表列舉而非硬編清單 → 容器內缺 OpenFGA 的表時不會假綠。
	granted := grantedTables(t, db)
	for table := range businessTables {
		want := 4
		if table == "sales_order_events" || table == "print_logs" || table == "print_previews" {
			want = 2
		}
		if table == "notification_templates" || table == "notifications" {
			want = 3
		}
		if granted[table] != want {
			t.Errorf("業務表 %s 應有 %d 項授權,got %d", table, want, granted[table])
		}
	}
	for table, n := range granted {
		if n > 0 && !businessTables[table] {
			t.Errorf("app_rw 對非業務表 %s 有 %d 項授權（00022 只應授權業務表）", table, n)
		}
	}

	// 序列授權:每一張可 INSERT 的業務表,其 id 序列（bigserial／identity）都必須有 USAGE ——
	// 少了它,INSERT 會以 42501 `permission denied for sequence …` 失敗,而**表級授權看起來
	// 完全正常**（上一個迴圈會全綠）。00031 正是漏了 `order_counters_id_seq`,導致
	// 「該公司該來源的第一張訂單」在生產環境建不起來（counter 列不存在 → INSERT → 被拒;
	// 已有 counter 的公司走 UPDATE 不受影響,所以只有新公司會壞）。以 has_sequence_privilege
	// 逐表列舉,讓下一個漏掉的 GRANT 在這裡就紅。
	assertSequenceUsageGranted(t, db)

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

// assertSequenceUsageGranted 斷言「每張可 INSERT 的業務表,其 id 序列都有 USAGE」。
//
// 為什麼獨立成一條:表級 DML 授權與序列授權是**兩份清單**,migration 要各寫一次
// （`GRANT SELECT, INSERT … ON tbl` 與 `GRANT USAGE … ON tbl_id_seq`）。上一個斷言只看前者,
// 所以 00031 漏掉 `order_counters_id_seq` 時整支測試仍是綠的 —— 直到有人在真 PG 上跑
// 「建立該公司的第一張訂單」才以 42501 爆掉（見 00043 檔頭）。
//
// 判定方式:`pg_depend` 找出每張表所擁有的序列（`deptype` a=serial／i=identity），
// 再以 `has_sequence_privilege` 問真值 —— 不用 information_schema（只顯示當前角色有權的物件,
// 會讓「未授權」變成查不到而假綠,同 grantedTables 的理由）。
func assertSequenceUsageGranted(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`
		SELECT c.relname AS tbl, s.relname AS seq
		  FROM pg_class c
		  JOIN pg_namespace n ON n.oid = c.relnamespace
		  JOIN pg_depend d ON d.refobjid = c.oid AND d.deptype IN ('a','i')
		  JOIN pg_class s ON s.oid = d.objid AND s.relkind = 'S'
		 WHERE n.nspname = 'public' AND c.relkind = 'r'
		   AND has_table_privilege('app_rw', c.oid, 'INSERT')
		 ORDER BY c.relname`)
	if err != nil {
		t.Fatalf("列舉序列: %v", err)
	}
	defer rows.Close()

	checked := 0
	for rows.Next() {
		var tbl, seq string
		if err := rows.Scan(&tbl, &seq); err != nil {
			t.Fatalf("掃描序列: %v", err)
		}
		checked++
		var ok bool
		if err := db.QueryRow(`SELECT has_sequence_privilege('app_rw', $1, 'USAGE')`, seq).Scan(&ok); err != nil {
			t.Fatalf("查序列 %s 權限: %v", seq, err)
		}
		if !ok {
			t.Errorf("app_rw 對 %s 有 INSERT 卻對其序列 %s 無 USAGE —— INSERT 會以 42501 失敗", tbl, seq)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("走訪序列: %v", err)
	}
	// 至少要有幾條序列被檢查到,否則查詢本身寫壞（例如 deptype 條件錯）會讓本函式空轉而假綠。
	if checked < 20 {
		t.Fatalf("只檢查到 %d 條序列,預期 ≥20 —— 列舉查詢可能已失效", checked)
	}
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
