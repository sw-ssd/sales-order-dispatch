//go:build integration

package fileassets_test

import (
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSFileAssetsIsolation 以 app_rw 直連驗證檔案域隔離
// （00036 ENABLE + FORCE）：未設 scope → 0 列；設 A 公司 → 只見 A；跨租戶寫入擋。
func TestIntegrationRLSFileAssetsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUpFile(t, dsn)
	admin := openRawDBFile(t, dsn)

	coA := insertRLSCompanyFile(t, admin, "A", "RLSFA-A")
	coB := insertRLSCompanyFile(t, admin, "B", "RLSFA-B")
	for _, co := range []int{coA, coB} {
		if _, err := admin.Exec(
			`INSERT INTO file_assets (company_id, owner_type, owner_id, filename, original_filename,
			 mime_type, size_bytes, storage_path, url, created_at)
			 VALUES ($1, 'company', $2, 'f.bin', 'o.bin', 'image/png', 8, 'p', 'u', now())`,
			co, co); err != nil {
			t.Fatalf("建檔案(公司 %d): %v", co, err)
		}
	}

	app := openAppRoleDBFile(t, dsn)

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		var n int
		if err := app.QueryRow(`SELECT count(*) FROM file_assets`).Scan(&n); err != nil {
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
		setAppScopeFile(t, tx, coA)
		var got int
		if err := tx.QueryRow(`SELECT DISTINCT company_id FROM file_assets`).Scan(&got); err != nil {
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
		setAppScopeFile(t, tx, coA)
		_, err = tx.Exec(
			`INSERT INTO file_assets (company_id, owner_type, owner_id, filename, original_filename,
			 mime_type, size_bytes, storage_path, url, created_at)
			 VALUES ($1, 'company', $2, 'x.bin', 'y.bin', 'image/png', 8, 'p', 'u')`, coB, coB)
		if err == nil {
			t.Fatal("跨租戶寫入應被 WITH CHECK 擋下")
		}
	})
}
