//go:build integration

package fileassets_test

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// openRawDBFile 開 admin 直連(RLS 探針的 fixture 寫入用)。
func openRawDBFile(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("開 admin 連線: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// insertRLSCompanyFile 建公司回 id。
func insertRLSCompanyFile(t *testing.T, db *sql.DB, name, identifier string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ($1, $2, 'active') RETURNING id`,
		name, identifier).Scan(&id); err != nil {
		t.Fatalf("建公司 %s: %v", name, err)
	}
	return id
}

// openAppRoleDBFile 開 app_rw 直連(生產路徑角色)。
func openAppRoleDBFile(t *testing.T, adminDSN string) *sql.DB {
	t.Helper()
	app, err := sql.Open("pgx", testsupport.AppRoleDSN(t, adminDSN))
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	return app
}

// setAppScopeFile 於交易內套用公司範圍(值來自 fixture 整數,非使用者輸入)。
func setAppScopeFile(t *testing.T, tx *sql.Tx, companyID int) {
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
