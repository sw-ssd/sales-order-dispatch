//go:build integration

package services

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSPrintRecordsIsolation 以 app_rw 直連驗證列印域隔離
// （00038 ENABLE + FORCE）：未設 scope → 0 列；設 A 公司 → 只見 A；跨租戶寫入擋。
// print_logs/print_previews 僅追加(SELECT/INSERT),故只探 SELECT 可見性 + INSERT 擋。
func TestIntegrationRLSPrintRecordsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	admin := openRawDB(t, dsn)

	coA := insertRLSCompanyPrint(t, admin, "A", "RLSPR-A")
	coB := insertRLSCompanyPrint(t, admin, "B", "RLSPR-B")
	faA := insertRLSFileAsset(t, admin, coA)
	faB := insertRLSFileAsset(t, admin, coB)
	for _, tc := range []struct {
		tbl string
		co  int
		fa  int
	}{
		{"print_logs", coA, faA}, {"print_logs", coB, faB},
		{"print_previews", coA, faA}, {"print_previews", coB, faB},
	} {
		by := "printed_by"
		if tc.tbl == "print_previews" {
			by = "previewed_by"
		}
		at := "printed_at"
		if tc.tbl == "print_previews" {
			at = "previewed_at"
		}
		if _, err := admin.Exec(
			`INSERT INTO `+tc.tbl+` (company_id, department_id, document_type, route_id,
			 target_date, `+by+`, `+at+`, file_asset_id)
			 VALUES ($1, 1, 'dispatch_summary', 1, '2026-09-22', 1, now(), $2)`,
			tc.co, tc.fa); err != nil {
			t.Fatalf("建列印記錄 %s(公司 %d): %v", tc.tbl, tc.co, err)
		}
	}

	app := openAppRoleDB(t, dsn)

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		for _, tbl := range []string{"print_logs", "print_previews"} {
			var n int
			if err := app.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n); err != nil {
				t.Fatalf("%s 查詢: %v", tbl, err)
			}
			if n != 0 {
				t.Fatalf("%s 未設 scope 不得可見,got %d", tbl, n)
			}
		}
	})

	t.Run("scope=company A → 只見 A", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScopePrint(t, tx, coA)
		for _, tbl := range []string{"print_logs", "print_previews"} {
			var got int
			if err := tx.QueryRow(`SELECT DISTINCT company_id FROM ` + tbl).Scan(&got); err != nil {
				t.Fatalf("%s 查詢: %v", tbl, err)
			}
			if got != coA {
				t.Fatalf("%s 應只看到公司 %d,got %d", tbl, coA, got)
			}
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
			`INSERT INTO print_logs (company_id, department_id, document_type, route_id,
			 target_date, printed_by, printed_at, file_asset_id)
			 VALUES ($1, 1, 'dispatch_summary', 1, '2026-09-22', 1, now(), $2)`, coB, faB)
		if err == nil {
			t.Fatal("跨租戶寫入應被 WITH CHECK 擋下")
		}
	})
}

// insertRLSCompanyPrint 建公司回 id。
func insertRLSCompanyPrint(t *testing.T, db *sql.DB, name, identifier string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ($1, $2, 'active') RETURNING id`,
		name, identifier).Scan(&id); err != nil {
		t.Fatalf("建公司 %s: %v", name, err)
	}
	return id
}

// insertRLSFileAsset 建 file_asset 回 id(print FK 用)。
func insertRLSFileAsset(t *testing.T, db *sql.DB, coID int) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO file_assets (company_id, owner_type, owner_id, filename, original_filename,
		 mime_type, size_bytes, storage_path, url, created_at)
		 VALUES ($1, 'company', $2, 'f.pdf', 'o.pdf', 'application/pdf', 8, 'p', 'u', now()) RETURNING id`,
		coID, coID).Scan(&id); err != nil {
		t.Fatalf("建 file_asset(公司 %d): %v", coID, err)
	}
	return id
}

// setAppScopePrint 於交易內套用公司範圍。
func setAppScopePrint(t *testing.T, tx *sql.Tx, companyID int) {
	t.Helper()
	for _, stmt := range []string{
		`SET LOCAL app.current_data_scope = 'company'`,
		`SET LOCAL app.current_company_id = '` + strconv.Itoa(companyID) + `'`,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
}
