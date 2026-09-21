//go:build integration

package services

import (
	"strconv"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSReturnRequestsIsolation 以 app_rw 直連驗證退貨域隔離
// （00040 ENABLE + FORCE）：未設 scope → 0 列；設 A 公司 → 只見 A；跨租戶寫入擋；
// self 範圍客戶 A1 只見自己客戶的申請（同公司客戶 A2 的不可見）。
func TestIntegrationRLSReturnRequestsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	admin := openRawDB(t, dsn)

	coA := insertRLSCompanyPrint(t, admin, "A", "RLSRT-A")
	coB := insertRLSCompanyPrint(t, admin, "B", "RLSRT-B")
	custA1 := insertRLSCustomer(t, admin, coA, "RLSA1", "客戶A1")
	custA2 := insertRLSCustomer(t, admin, coA, "RLSA2", "客戶A2")
	custB := insertRLSCustomer(t, admin, coB, "RLSB1", "客戶B1")
	for _, tc := range []struct{ co, cust int }{
		{coA, custA1}, {coA, custA2}, {coB, custB},
	} {
		if _, err := admin.Exec(
			`INSERT INTO return_requests (company_id, department_id, customer_id,
			 created_by_user_id, status, version) VALUES ($1, 1, $2, 1, 'pending', 0)`,
			tc.co, tc.cust); err != nil {
			t.Fatalf("建退貨申請(公司 %d 客戶 %d): %v", tc.co, tc.cust, err)
		}
	}
	// 明細各一筆(供 items 表隔離探針)。
	var rrA1 int
	if err := admin.QueryRow(`SELECT id FROM return_requests WHERE customer_id = $1`, custA1).Scan(&rrA1); err != nil {
		t.Fatalf("查申請: %v", err)
	}
	if _, err := admin.Exec(
		`INSERT INTO return_request_items (return_request_id, company_id, department_id,
		 source_type, product_id, product_name, unit, quantity, reason)
		 VALUES ($1, $2, 1, 'customer_product', 1, '蘋果', '斤', '1', '爛果')`, rrA1, coA); err != nil {
		t.Fatalf("建明細: %v", err)
	}

	app := openAppRoleDB(t, dsn)

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		for _, tbl := range []string{"return_requests", "return_request_items"} {
			var n int
			if err := app.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n); err != nil {
				t.Fatalf("%s 查詢: %v", tbl, err)
			}
			if n != 0 {
				t.Fatalf("%s 未設 scope 不得可見,got %d", tbl, n)
			}
		}
	})

	t.Run("scope=company A → 只見 A 兩筆", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScopePrint(t, tx, coA)
		var n int
		if err := tx.QueryRow(`SELECT count(*) FROM return_requests`).Scan(&n); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if n != 2 {
			t.Fatalf("公司 A 應見 2 筆,got %d", n)
		}
	})

	t.Run("self 客戶 A1 → 只見自己 1 筆", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		for _, stmt := range []string{
			`SET LOCAL app.current_data_scope = 'self'`,
			`SET LOCAL app.current_company_id = '` + strconv.Itoa(coA) + `'`,
			`SET LOCAL app.current_customer_id = '` + strconv.Itoa(custA1) + `'`,
		} {
			if _, err := tx.Exec(stmt); err != nil {
				t.Fatalf("%s: %v", stmt, err)
			}
		}
		var n int
		if err := tx.QueryRow(`SELECT count(*) FROM return_requests`).Scan(&n); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if n != 1 {
			t.Fatalf("self A1 應只見 1 筆,got %d", n)
		}
	})

	t.Run("跨租戶 INSERT → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScopePrint(t, tx, coA)
		_, err = tx.Exec(
			`INSERT INTO return_requests (company_id, department_id, customer_id,
			 created_by_user_id, status, version) VALUES ($1, 1, $2, 1, 'pending', 0)`, coB, custB)
		if err == nil {
			t.Fatal("跨租戶寫入應被 WITH CHECK 擋下")
		}
	})
}
