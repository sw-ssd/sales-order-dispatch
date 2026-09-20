//go:build integration

package server

import (
	"net/url"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationBusinessRoleMustNotBypassRLS 釘住「整個 RLS 計畫的前提」:production 的**業務連線**
// 必須是非 superuser(且不帶 BYPASSRLS)。
//
// PostgreSQL 的 superuser 恆繞過 RLS,`FORCE ROW LEVEL SECURITY` 亦然 —— 但 `DATABASE_URL` 的預設值
// 就是 superuser(config/database.go),而 `openEntClient` 原樣使用它。部署誤設時業務路徑以 superuser
// 連線 → 00024–00028 的租戶邊界**靜默消失**(所有端點與測試照常綠),故必須在啟動時 fail-fast。
//
// 三條斷言:①容器的 superuser 連線被拒且附修復指引(指向 app_rw);②一般非 superuser 角色通過、
// 帶 BYPASSRLS 的角色被拒(避免判斷寫成「只認角色名」);③`Server.Init()` 真的呼叫了它,且只在
// production 生效(development 沿用預設 superuser DSN 不得因此被擋 —— 既有測試與本機開發不變)。
func TestIntegrationBusinessRoleMustNotBypassRLS(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	admin := coreOpenDB(t, adminDSN) // 容器的 postgres 是 superuser

	// ① 業務 DSN 就是 superuser(部署誤設的實際樣態)→ 必須拒絕,並給出可行動的修復指引。
	err := assertBusinessRoleNotSuperuser(t.Context(), adminDSN)
	if err == nil {
		t.Fatal("業務連線為 superuser 時必須拒絕啟動(否則 RLS 被繞過且無人察覺)")
	}
	if !strings.Contains(err.Error(), "app_rw") {
		t.Fatalf("拒絕訊息必須附修復指引(把 DATABASE_URL 指向 app_rw),got %q", err.Error())
	}

	// ② 同一檢查對「非 superuser」角色必須放行;對 BYPASSRLS 角色必須拒絕。
	for _, tc := range []struct {
		role   string
		attrs  string
		reject bool
	}{
		{role: "rls_probe_plain", attrs: "NOSUPERUSER NOBYPASSRLS", reject: false},
		{role: "rls_probe_bypass", attrs: "NOSUPERUSER BYPASSRLS", reject: true},
	} {
		t.Run(tc.role, func(t *testing.T) {
			if _, err := admin.Exec(`CREATE ROLE ` + tc.role + ` LOGIN PASSWORD 'probe' ` + tc.attrs); err != nil {
				t.Fatalf("建立探針角色 %s: %v", tc.role, err)
			}
			gotErr := assertBusinessRoleNotSuperuser(t.Context(), probeRoleDSN(t, adminDSN, tc.role))
			if tc.reject && gotErr == nil {
				t.Fatalf("角色 %s(%s)必須被拒", tc.role, tc.attrs)
			}
			if !tc.reject && gotErr != nil {
				t.Fatalf("角色 %s(%s)應通過,got %v", tc.role, tc.attrs, gotErr)
			}
		})
	}

	// 生產實際指向的角色:00022 建出的 app_rw(非 owner、NOBYPASSRLS)必須通過。
	t.Run("app_rw(00022 的業務角色)→ 通過", func(t *testing.T) {
		migrateCoreUp(t, adminDSN) // 需要 00022 已建立 app_rw 角色
		if err := assertBusinessRoleNotSuperuser(t.Context(), testsupport.AppRoleDSN(t, adminDSN)); err != nil {
			t.Fatalf("app_rw 作為業務連線必須通過,got %v", err)
		}
	})

	// ③ 接線:production 的 Init() 必須走到這個檢查(superuser DSN → 錯誤指向業務連線角色);
	// development 不得被這個檢查擋下(既有整合測試與本機開發仍用預設的 superuser DSN)。
	prod := New(&config.Config{
		API:      config.API{Env: "production"},
		Auth:     config.Auth{JWTSecret: "custom-secret"},
		Database: config.Database{DatabaseURL: adminDSN},
	})
	initErr := prod.Init()
	if initErr == nil {
		t.Fatal("production + superuser 業務 DSN 必須拒絕啟動")
	}
	if !strings.Contains(initErr.Error(), "業務連線角色") {
		t.Fatalf("production 的拒絕必須來自角色檢查(而非其他檢查),got %q", initErr.Error())
	}

	dev := New(&config.Config{
		API:      config.API{Env: "development"},
		Database: config.Database{DatabaseURL: adminDSN},
	})
	if err := dev.Init(); err != nil {
		t.Fatalf("development 不得因此檢查被擋(既有測試與本機開發沿用預設 superuser DSN),got %v", err)
	}
}

// probeRoleDSN 由 admin DSN 導出指定角色的連線字串(僅替換 userinfo;庫名／sslmode 沿用)。
func probeRoleDSN(t *testing.T, adminDSN, role string) string {
	t.Helper()
	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("解析 DSN: %v", err)
	}
	u.User = url.UserPassword(role, "probe")
	return u.String()
}
