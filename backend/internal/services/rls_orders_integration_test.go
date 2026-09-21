//go:build integration

package services

import (
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSOrdersIsolation 以 app_rw 直連驗證訂單域隔離（00032 ENABLE + FORCE）：
// 未設 scope → 四表皆 0 列（fail-closed）；設 A 公司 → 訂單／明細／counters 只見 A；
// 以 A 的身分寫入 B 公司的訂單 → 被 WITH CHECK 擋下；events 無 UPDATE/DELETE 權限。
//
// 為何用 app_rw 而非 owner：容器 admin 是 superuser（永遠繞過 RLS，FORCE 亦然），
// 只有 00022 的 NOBYPASSRLS 業務角色能測到真正的 fail-closed（比照客戶域探針）。
func TestIntegrationRLSOrdersIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	coA := insertRLSCompany(t, admin, "A", "RLSORD-A")
	coB := insertRLSCompany(t, admin, "B", "RLSORD-B")
	for _, co := range []int{coA, coB} {
		var custID int
		if err := admin.QueryRow(
			`INSERT INTO customers (company_id, customer_code, name, created_at, updated_at)
			 VALUES ($1, $2, '客戶', now(), now()) RETURNING id`, co, "C-"+itoa(co)).Scan(&custID); err != nil {
			t.Fatalf("建客戶(公司 %d): %v", co, err)
		}
		if _, err := admin.Exec(
			`INSERT INTO sales_orders (company_id, order_no, customer_id, source, status, version, created_at, updated_at)
			 VALUES ($1, $2, $3, 'W', 'pending', 1, now(), now())`, co, "W000001", custID); err != nil {
			t.Fatalf("建訂單(公司 %d): %v", co, err)
		}
		if _, err := admin.Exec(
			`INSERT INTO order_counters (company_id, source, next_seq, version) VALUES ($1, 'W', 2, 1)`, co); err != nil {
			t.Fatalf("建 counter(公司 %d): %v", co, err)
		}
	}

	app := openAppRoleDB(t, adminDSN)

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		for _, tb := range []string{"sales_orders", "sales_order_items", "sales_order_events", "order_counters"} {
			var n int
			if err := app.QueryRow(`SELECT count(*) FROM ` + tb).Scan(&n); err != nil {
				t.Fatalf("查 %s: %v", tb, err)
			}
			if n != 0 {
				t.Fatalf("未設 scope 時 %s 不得可見,got %d", tb, n)
			}
		}
	})

	t.Run("scope=company A → 只見 A", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScope(t, tx, coA)
		var got int
		if err := tx.QueryRow(`SELECT DISTINCT company_id FROM sales_orders`).Scan(&got); err != nil {
			t.Fatalf("查訂單: %v", err)
		}
		if got != coA {
			t.Fatalf("應只看到公司 %d,got %d", coA, got)
		}
		var cnt int
		if err := tx.QueryRow(`SELECT DISTINCT company_id FROM order_counters`).Scan(&cnt); err != nil {
			t.Fatalf("查 counter: %v", err)
		}
		if cnt != coA {
			t.Fatalf("counter 應只看到公司 %d,got %d", coA, cnt)
		}
	})

	t.Run("跨租戶 INSERT → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScope(t, tx, coA)
		var custB int
		if err := tx.QueryRow(`SELECT id FROM customers LIMIT 1`).Scan(&custB); err != nil {
			t.Fatalf("讀客戶(應為 A 家): %v", err)
		}
		_ = custB
		// 直接以 B 的 company_id 寫訂單 → WITH CHECK 擋下（42501）。
		var custIDB int
		if err := admin.QueryRow(`SELECT id FROM customers WHERE company_id = $1 LIMIT 1`, coB).Scan(&custIDB); err != nil {
			t.Fatalf("讀 B 家客戶: %v", err)
		}
		_, err = tx.Exec(
			`INSERT INTO sales_orders (company_id, order_no, customer_id, source, status, version, created_at, updated_at)
			 VALUES ($1, 'W000002', $2, 'W', 'pending', 1, now(), now())`, coB, custIDB)
		if err == nil {
			t.Fatal("跨租戶寫訂單應被 WITH CHECK 擋下")
		}
	})

	t.Run("events 僅追加", func(t *testing.T) {
		var hasUpdate, hasDelete bool
		if err := app.QueryRow(`SELECT has_table_privilege('app_rw', 'sales_order_events', 'UPDATE')`).Scan(&hasUpdate); err != nil {
			t.Fatalf("查權限: %v", err)
		}
		if err := app.QueryRow(`SELECT has_table_privilege('app_rw', 'sales_order_events', 'DELETE')`).Scan(&hasDelete); err != nil {
			t.Fatalf("查權限: %v", err)
		}
		if hasUpdate || hasDelete {
			t.Fatal("sales_order_events 必須僅追加(app_rw 不得 UPDATE/DELETE)")
		}
	})
}
