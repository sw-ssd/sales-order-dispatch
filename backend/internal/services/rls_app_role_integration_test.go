//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationAppRolePrivileges 驗證業務角色是非 owner、且對業務表有 DML 權限：
// RLS 只在「非 owner」連線上才擋得住（owner 需 FORCE，而 FORCE 連帶要求回填走系統 scope）。
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

	var granted int
	if err := db.QueryRow(
		`SELECT count(*) FROM information_schema.role_table_grants
		  WHERE grantee = 'app_rw' AND table_name = 'customers'
		    AND privilege_type IN ('SELECT','INSERT','UPDATE','DELETE')`,
	).Scan(&granted); err != nil {
		t.Fatalf("查授權: %v", err)
	}
	if granted != 4 {
		t.Fatalf("app_rw 對 customers 應有 SELECT/INSERT/UPDATE/DELETE 四項授權,got %d", granted)
	}
}
