//go:build integration

package services

import (
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSCustomerProductsIsolation 以 app_rw 直連驗證清單域隔離
// （00034 ENABLE + FORCE）：未設 scope → 0 列；設 A 公司 → 只見 A；跨租戶寫入擋。
func TestIntegrationRLSCustomerProductsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	coA := insertRLSCompany(t, admin, "A", "RLSCP-A")
	coB := insertRLSCompany(t, admin, "B", "RLSCP-B")
	for _, co := range []int{coA, coB} {
		var custID, prodID int
		if err := admin.QueryRow(
			`INSERT INTO customers (company_id, customer_code, name, created_at, updated_at)
			 VALUES ($1, $2, '客戶', now(), now()) RETURNING id`, co, "C-"+itoa(co)).Scan(&custID); err != nil {
			t.Fatalf("建客戶(公司 %d): %v", co, err)
		}
		if err := admin.QueryRow(
			`INSERT INTO products (company_id, code, name, created_at, updated_at)
			 VALUES ($1, $2, '商品', now(), now()) RETURNING id`, co, "P-"+itoa(co)).Scan(&prodID); err != nil {
			t.Fatalf("建商品(公司 %d): %v", co, err)
		}
		if _, err := admin.Exec(
			`INSERT INTO customer_products (company_id, customer_id, product_id, alias_name, default_qty, created_at, updated_at)
			 VALUES ($1, $2, $3, '別名', '1', now(), now())`, co, custID, prodID); err != nil {
			t.Fatalf("建清單(公司 %d): %v", co, err)
		}
	}

	app := openAppRoleDB(t, adminDSN)

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		var n int
		if err := app.QueryRow(`SELECT count(*) FROM customer_products`).Scan(&n); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if n != 0 {
			t.Fatalf("未設 scope 不得可見,got %d", n)
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
		if err := tx.QueryRow(`SELECT DISTINCT company_id FROM customer_products`).Scan(&got); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if got != coA {
			t.Fatalf("應只看到公司 %d,got %d", coA, got)
		}
	})

	t.Run("跨租戶 INSERT → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScope(t, tx, coA)
		var custA, prodA int
		if err := tx.QueryRow(`SELECT customer_id, product_id FROM customer_products LIMIT 1`).Scan(&custA, &prodA); err != nil {
			t.Fatalf("讀 A 家列: %v", err)
		}
		_ = custA
		_ = prodA
		var custIDB, prodIDB int
		if err := admin.QueryRow(`SELECT id FROM customers WHERE company_id = $1 LIMIT 1`, coB).Scan(&custIDB); err != nil {
			t.Fatalf("讀 B 家客戶: %v", err)
		}
		if err := admin.QueryRow(`SELECT id FROM products WHERE company_id = $1 LIMIT 1`, coB).Scan(&prodIDB); err != nil {
			t.Fatalf("讀 B 家商品: %v", err)
		}
		_, err = tx.Exec(
			`INSERT INTO customer_products (company_id, customer_id, product_id, alias_name, default_qty, created_at, updated_at)
			 VALUES ($1, $2, $3, '別家', '1', now(), now())`, coB, custIDB, prodIDB)
		if err == nil {
			t.Fatal("跨租戶寫入應被 WITH CHECK 擋下")
		}
	})
}
