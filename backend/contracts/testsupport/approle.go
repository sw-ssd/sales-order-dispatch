//go:build integration

package testsupport

import (
	"database/sql"
	"net/url"
	"testing"
)

// AppRoleDSN 由 owner DSN 導出 app_rw 連線字串：先把密碼設成固定值（容器為拋棄式，
// 且本檔是唯一設定點），再替換 userinfo。非容器模式（INTEGRATION_TEST_DSN）同樣可行，
// 但會改動該 DSN 指向的庫之角色密碼 → 需專用拋棄式資料庫。
func AppRoleDSN(t *testing.T, adminDSN string) string {
	t.Helper()
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("開 admin 連線: %v", err)
	}
	defer func() { _ = admin.Close() }()
	if _, err := admin.Exec(`ALTER ROLE app_rw WITH PASSWORD 'app_rw'`); err != nil {
		t.Fatalf("設定 app_rw 密碼: %v（migration 00022 是否已套用？）", err)
	}

	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("解析 DSN: %v", err)
	}
	u.User = url.UserPassword("app_rw", "app_rw")
	return u.String()
}
