//go:build integration

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformv1connect "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationPlatformAdminMount 驗平台 RPC 的**掛載契約**：走 InitDomains() 的真 router
// （掛載點就是 production 的掛載點），每一步都是各自獨立的全滅模式。
//
// 為什麼需要這條測試（掛載位置不像「有掛就好」）：
//   - operator cookie 的 Path 是 /platform，而 RFC 6265 的 path-match 是逐段前綴比對。
//     若照 Connect 的自然路徑把 handler 掛在根（/platform.v1.PlatformAdminService/…），
//     未涵蓋的第一個字元是 "." → 瀏覽器**不會送出** cookie → 已登入的 operator 每個請求
//     都拿到 401，而且後端的 log 看起來一切正常（cookie 從沒到過伺服器）。
//     直接塞 Cookie 標頭的測試看不到這件事，故這裡同時斷言**路徑關係**本身。
//   - 反方向的錯誤（把 cookie 的 Path 放寬成 "/"）會讓 operator cookie 跟著送往租戶 API，
//     故 cookie 路徑不得改；兩者只能靠掛載前綴對齊。
//   - 平台 RPC 必須經 operatorauth interceptor：無 cookie／拿租戶密鑰簽的 token 冒充
//     operator cookie 都必須 401（租戶身分與平台 token 互不通用，T8 的另一半）。
func TestIntegrationPlatformAdminMount(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateCoreUp(t, dsn)
	admin := coreOpenDB(t, dsn)

	// operator 白名單列：interceptor 每次請求都回查白名單（停用即失效），故夾具必須真的有一列。
	var operatorID int64
	if err := admin.QueryRow(`INSERT INTO platform.operators (email, name, role, status)
		VALUES ('ops@example.com','維運','admin','active') RETURNING id`).Scan(&operatorID); err != nil {
		t.Fatalf("seed operator: %v", err)
	}
	if _, err := admin.Exec(`INSERT INTO companies (name, identifier) VALUES ('掛載測試公司','MOUNT-TENANT')`); err != nil {
		t.Fatalf("seed 公司: %v", err)
	}

	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("DATABASE_ADMIN_URL", dsn)
	t.Setenv("OPENFGA_ENABLED", "false")
	t.Setenv("JWT_SECRET", tenantSecret)
	t.Setenv("PLATFORM_ALLOWED_EMAIL_DOMAIN", "example.com")
	t.Setenv("PLATFORM_JWT_SECRET", operatorSecret)
	t.Setenv("PLATFORM_CONSOLE_URL", consoleURL)
	// 不設 GOOGLE_CLIENT_ID：平台 RPC 不得依賴 Google discovery（測試也不得需要外網）。
	t.Setenv("GOOGLE_CLIENT_ID", "")

	s := New(config.New())
	s.InitDomains()
	if s.operatorAuth == nil {
		t.Fatal("平台設定齊備時 InitDomains 必須掛載 operatorauth")
	}

	token, err := s.operatorAuth.IssueToken(operatorauth.Identity{
		OperatorID: operatorID, Email: "ops@example.com", Role: "admin",
	})
	if err != nil {
		t.Fatalf("簽發 operator token: %v", err)
	}

	procedure := platformv1connect.PlatformAdminServiceListTenantsProcedure
	mountedPath := operatorauth.CookiePath + procedure
	if !rfc6265PathMatches(operatorauth.CookiePath, mountedPath) {
		t.Fatalf("掛載路徑 %s 未被 cookie 的 Path %s 涵蓋（RFC 6265 path-match）→ 瀏覽器不會送出 operator cookie",
			mountedPath, operatorauth.CookiePath)
	}

	t.Run("未加前綴的 procedure 路徑不存在（沒掛在根）", func(t *testing.T) {
		// 掛在根（cookie 送不出去）與掛在 /platform 之下的差別，就靠這一條釘住。
		if code := postPlatform(t, s, procedure, "", ""); code != http.StatusNotFound {
			t.Fatalf("%s 不應存在（必須掛在 %s 之下），got %d", procedure, operatorauth.CookiePath, code)
		}
	})

	t.Run("帶 operator cookie → 200 且讀得到跨租戶資料", func(t *testing.T) {
		rec := postPlatformRec(t, s, mountedPath, operatorauth.CookieName, token)
		if rec.Code != http.StatusOK {
			t.Fatalf("operator cookie 應可讀取,got %d: %s", rec.Code, rec.Body.String())
		}
		var out struct {
			Tenants []struct {
				CompanyName string `json:"companyName"`
			} `json:"tenants"`
			Pagination struct {
				Total int32 `json:"total"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("回應不是合法的 ListTenantsResponse: %v（%s）", err, rec.Body.String())
		}
		if out.Pagination.Total < 1 {
			t.Fatalf("夾具有一家租戶,got total %d（0 代表投影查詢被 RLS 擋掉）", out.Pagination.Total)
		}
		if len(out.Tenants) == 0 || out.Tenants[0].CompanyName != "掛載測試公司" {
			t.Fatalf("租戶投影錯誤: %v", out.Tenants)
		}
	})

	t.Run("無 cookie → 401", func(t *testing.T) {
		if code := postPlatform(t, s, mountedPath, "", ""); code != http.StatusUnauthorized {
			t.Fatalf("未帶 operator cookie 應回 401,got %d", code)
		}
	})

	t.Run("租戶密鑰簽的 token 冒充 operator cookie → 401", func(t *testing.T) {
		// 兩個 secret 不同是硬規則(不同 secret ＋ 不同 audience):租戶 token 到不了平台工具。
		forged := tenantSignedToken(t, operatorID)
		if code := postPlatform(t, s, mountedPath, operatorauth.CookieName, forged); code != http.StatusUnauthorized {
			t.Fatalf("租戶密鑰簽的 token 不得通過平台 interceptor,got %d", code)
		}
	})

	t.Run("平台 RPC 不需要 Google discovery（登入端點 503 時仍可讀）", func(t *testing.T) {
		if code := statusOf(s, operatorauth.LoginPath); code != http.StatusServiceUnavailable {
			t.Fatalf("缺 OIDC 依賴時登入端點應為 503（路由須仍在）,got %d", code)
		}
		if code := postPlatform(t, s, mountedPath, operatorauth.CookieName, token); code != http.StatusOK {
			t.Fatalf("無 OIDC 依賴不得影響平台 RPC,got %d", code)
		}
	})
}

// postPlatform 以 Connect 協定(JSON)對平台 RPC 發一次 unary 請求,回傳狀態碼。
// cookieValue 為空 = 不帶 cookie。
func postPlatform(t *testing.T, s *Server, path, cookieName, cookieValue string) int {
	t.Helper()
	return postPlatformRec(t, s, path, cookieName, cookieValue).Code
}

// postPlatformRec 發請求並回傳整個回應。
//
// 以 Cookie 標頭直接帶 cookie:測試跑不到瀏覽器的 cookie 篩選,故「cookie 會不會被送出」由
// rfc6265PathMatches 的路徑契約斷言,這裡只驗伺服器端的接收與授權。
func postPlatformRec(t *testing.T, s *Server, path, cookieName, cookieValue string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"page":1,"pageSize":20}`))
	req.Header.Set("Content-Type", "application/json")
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: cookieValue})
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

// tenantSignedToken 以**租戶**密鑰簽一個看起來像 operator session 的 token(冒充用)。
func tenantSignedToken(t *testing.T, subject int64) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": subject, "email": "ops@example.com", "role": "admin",
		"aud": "platform", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(),
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(tenantSecret))
	if err != nil {
		t.Fatalf("簽發租戶 token: %v", err)
	}
	return raw
}

// rfc6265PathMatches 實作 RFC 6265 §5.1.4 的 path-match:cookie-path 必須是 request-path 的前綴,
// 且**(cookie-path 不以 "/" 結尾時)接續的那個字元必須是 "/"**。
//
// 這一條就是「/platform 對 /platform.v1.… 不成立」的來源:接續字元是 "."。
func rfc6265PathMatches(cookiePath, requestPath string) bool {
	if cookiePath == requestPath {
		return true
	}
	if !strings.HasPrefix(requestPath, cookiePath) {
		return false
	}
	if strings.HasSuffix(cookiePath, "/") {
		return true
	}
	return requestPath[len(cookiePath)] == '/'
}
