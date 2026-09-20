//go:build integration

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationPlatformAuthMount 驗平台認證的**掛載契約**（需要真 admin 連線：掛載會開 owner 池）。
//
// 兩個失效模式是這條測試存在的理由：
//   - 設定齊備卻沒掛：T9 的 PlatformAdminService 拿不到 interceptor（`s.operatorAuth` 為 nil），
//     平台 RPC 只能整組不掛——「設定好了卻什麼都沒有」在 log 上只會看到一行 skip。
//   - 缺 OIDC 依賴時路由必須仍在（回 503）：回 404 會被當成掛載壞掉，維運查錯方向。
//     503 也讓「平台工具未設定」與「設定齊備但登入暫不可用」兩件事在 log／回應上分得開。
func TestIntegrationPlatformAuthMount(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateCoreUp(t, dsn)

	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("DATABASE_ADMIN_URL", dsn)
	t.Setenv("OPENFGA_ENABLED", "false")
	t.Setenv("PLATFORM_ALLOWED_EMAIL_DOMAIN", "example.com")
	// 不設 GOOGLE_CLIENT_ID：掛載不依賴 Google discovery（測試不得需要外網）。
	t.Setenv("GOOGLE_CLIENT_ID", "")

	t.Run("設定齊備 → 掛載 service 與登入端點", func(t *testing.T) {
		t.Setenv("PLATFORM_JWT_SECRET", operatorSecret)
		t.Setenv("PLATFORM_CONSOLE_URL", consoleURL)

		s := New(config.New())
		s.mountPlatformAuth()

		if s.operatorAuth == nil {
			t.Fatal("平台設定齊備時必須掛載 operatorauth（T9 的 interceptor 由此取用）")
		}
		// 缺 OIDC 依賴：端點在，回 503。
		for _, path := range []string{operatorauth.LoginPath, operatorauth.CallbackPath} {
			if code := statusOf(s, path); code != http.StatusServiceUnavailable {
				t.Fatalf("%s 應回 503（路由存在但登入未設定），got %d", path, code)
			}
		}
	})

	t.Run("整組未設 → 不掛載也不留路由", func(t *testing.T) {
		t.Setenv("PLATFORM_JWT_SECRET", "")
		t.Setenv("PLATFORM_CONSOLE_URL", "")

		s := New(config.New())
		s.mountPlatformAuth()

		if s.operatorAuth != nil {
			t.Fatal("平台設定未設齊時不得掛載 operatorauth（開發環境不設平台工具）")
		}
		if code := statusOf(s, operatorauth.LoginPath); code != http.StatusNotFound {
			t.Fatalf("未掛載時 %s 應回 404，got %d", operatorauth.LoginPath, code)
		}
	})
}

// statusOf 以 chi router 回應一個 GET 請求並取回狀態碼。
func statusOf(s *Server, path string) int {
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code
}
