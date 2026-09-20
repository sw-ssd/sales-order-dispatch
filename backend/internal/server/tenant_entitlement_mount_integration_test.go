//go:build integration

package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

const tenantEntitlementSecret = "t10b-tenant-entitlement-secret"

// TestIntegrationTenantEntitlements 驗租戶端權益投影 RPC 的**端到端掛載與身分隔離**：走
// InitDomains() 的真 router（掛載點就是 production 的掛載點）＋真 authzMiddleware ＋真 session
// ＋真計數器（app_rw／RLS）。
//
// 為什麼不能只靠單元層的 mux 斷言（`services.TestTenantEntitlementsMountedOnTenantMux` 自建 mux
// 再交給 Register…）：那種測試只證明「該函式會註冊」，刪掉 domains.go 的掛載行仍全綠，
// 也排除不了被掛到 /platform/。這個 RPC 的失敗模式是**功能消失**（租戶讀不到自己的方案）或
// **跨身分門戶**（平台工具路徑吃到租戶請求），兩者都必須由真實掛載點承擔 —— 本計畫已有兩次
// 同型教訓（T8 的 F-3：掛載測試繞過 InitDomains；B9：平台 RPC 以真 router＋真 cookie 釘住）。
func TestIntegrationTenantEntitlements(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateCoreUp(t, dsn)
	admin := coreOpenDB(t, dsn)
	valkeyAddr := testsupport.Valkey(t)

	// 夾具：A 有 2 個帳號（席位用量 2）、B 是另一個租戶（方案不同 → 用來釘「只回自己公司」）。
	coA := coreInsertCompany(t, admin, "權益投影 A", "ENT-A")
	coB := coreInsertCompany(t, admin, "權益投影 B", "ENT-B")
	userA := coreInsertUser(t, admin, coA, "ent-a@example.com", "company_admin", 0)
	coreInsertUser(t, admin, coA, "ent-a2@example.com", "staff", 0)
	coreInsertUser(t, admin, coB, "ent-b@example.com", "company_admin", 0)
	seedTenantEntitlementFixture(t, admin, coA, coB)

	var operatorID int64
	if err := admin.QueryRow(`INSERT INTO platform.operators (email, name, role, status)
		VALUES ('ops@example.com','維運','admin','active') RETURNING id`).Scan(&operatorID); err != nil {
		t.Fatalf("seed operator: %v", err)
	}

	// 平台工具也要設定齊備：本測試要拿**真的 operator token** 去撞租戶路徑（兩套身分不互通）。
	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("DATABASE_ADMIN_URL", dsn)
	t.Setenv("VALKEY_ADDR", valkeyAddr)
	t.Setenv("JWT_SECRET", tenantEntitlementSecret)
	t.Setenv("OPENFGA_ENABLED", "false")
	t.Setenv("PLATFORM_ALLOWED_EMAIL_DOMAIN", "example.com")
	t.Setenv("PLATFORM_JWT_SECRET", "t10b-operator-secret")
	t.Setenv("PLATFORM_CONSOLE_URL", "http://localhost:3000")
	// 不設 GOOGLE_CLIENT_ID：租戶 RPC 不得依賴 Google discovery（測試也不得需要外網）。
	t.Setenv("GOOGLE_CLIENT_ID", "")

	s := New(config.New())
	s.InitDomains()
	// Valkey 不可用時 mountAuth 會「log ＋ 略過掛載」→ 本測試若沒有這道檢查，會以 404 的形式
	// 偽裝成「掛載壞了」，把環境問題誤導成程式缺陷（skip 由 testsupport.Valkey 負責）。
	if s.entitlements == nil || s.operatorAuth == nil {
		t.Fatal("InitDomains 未完成組裝（entitlements／operatorAuth 為 nil）→ 掛載鏈在 Valkey 或平台設定處中斷")
	}

	// 租戶 session：以**同一顆 Valkey**（mountAuth 用的 session store）蓋一個已登入的 session cookie。
	valkeyClient := cache.NewClient(s.cfg.Cache.ValkeyAddr)
	defer func() { _ = valkeyClient.Close() }()
	sessions := auth.WebSessionManager(auth.NewSessionStore(auth.NewRedisStore(valkeyClient)),
		s.cfg.Auth.SessionLifetime, s.cfg.Auth.SessionSecure, s.cfg.Auth.SessionSameSite)
	cookieA := testSessionCookie(t, sessions, userA, "company_admin")

	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	procedure := platformv1connect.TenantEntitlementServiceGetTenantEntitlementsProcedure
	mountedPath := "/api/v1" + procedure

	clientA := platformv1connect.NewTenantEntitlementServiceClient(
		coreCookieHTTPClient(cookieA), srv.URL+"/api/v1")

	t.Run("租戶 session → 200、只回自己公司、用量與計數器一致", func(t *testing.T) {
		got := coreCall(t, clientA.GetTenantEntitlements, &platformv1.GetTenantEntitlementsRequest{})
		if got.GetPlanCode() != "std" || got.GetPlanName() != "標準" || got.GetStatus() != "active" {
			t.Fatalf("應回自己公司（A）的 std／標準／active，got %q／%q／%q",
				got.GetPlanCode(), got.GetPlanName(), got.GetStatus())
		}
		if seats := usageOfProto(got, entitlements.LimitSeats); seats == nil ||
			seats.GetUsed() != 2 || seats.GetLimitValue() != 10 {
			// 席位＝A 的兩個帳號；計數器走請求交易（app_rw ＋ RLS company scope）→ 真數字。
			t.Fatalf("席位用量應為 2/10，got %+v", seats)
		}
		if printing := usageOfProto(got, entitlements.FeaturePrinting); printing == nil ||
			!printing.GetEnabled() || printing.GetLimitSet() {
			t.Fatalf("boolean 功能應 enabled 且不帶 limit，got %+v", printing)
		}
		// 降級（controller 裁定）在本測試是**真計數器**：limit.storage_gb 有 feature 也有方案上限，
		// 但 services.countFeature 沒有它的分支 → 該筆略過（不回 0），其餘照常。
		if s := usageOfProto(got, entitlements.LimitStorageGB); s != nil {
			t.Fatalf("缺計數器的 feature 必須略過該筆（不得回 0／-1），got used=%d", s.GetUsed())
		}
		// B 的方案不得出現（投影只看身分帶的公司；本 RPC 沒有可指定他公司的參數）。
		if got.GetPlanCode() == "pro" {
			t.Fatal("回了他公司的方案（跨租戶洩漏）")
		}
	})

	t.Run("硬塞 operator cookie（真 token）→ 401：兩套身分不互通", func(t *testing.T) {
		token, err := s.operatorAuth.IssueToken(operatorauth.Identity{
			OperatorID: operatorID, Email: "ops@example.com", Role: "admin",
		})
		if err != nil {
			t.Fatalf("簽發 operator token: %v", err)
		}
		// 平台端先確認這顆 token 是有效的（否則「401」也可能只是壞 token，證明不了隔離）。
		platformPath := operatorauth.CookiePath + platformv1connect.PlatformAdminServiceListTenantsProcedure
		if code := postPlatform(t, s, platformPath, operatorauth.CookieName, token); code != http.StatusOK {
			t.Fatalf("對照組：同一顆 token 在平台路徑應為 200，got %d", code)
		}
		if code := postPlatform(t, s, mountedPath, operatorauth.CookieName, token); code != http.StatusUnauthorized {
			t.Fatalf("operator cookie 不得通過租戶路徑（應 401），got %d", code)
		}
	})

	t.Run("未加 /api/v1 前綴的 procedure 路徑不存在（不是根路徑萬用掛載）", func(t *testing.T) {
		if code := postPlatform(t, s, procedure, "", ""); code != http.StatusNotFound {
			t.Fatalf("%s 不應存在（租戶 RPC 必須掛在 /api/v1 之下），got %d", procedure, code)
		}
	})

	t.Run("掛在 /platform/ 之下也不算數（那是平台工具的路徑範圍）", func(t *testing.T) {
		if code := postPlatform(t, s, operatorauth.CookiePath+procedure, operatorauth.CookieName, ""); code != http.StatusNotFound {
			t.Fatalf("租戶 RPC 不得由 %s 提供（operator 路徑範圍），got %d", operatorauth.CookiePath, code)
		}
	})

	t.Run("未帶任何憑證 → 401（不得因未登入而回空方案）", func(t *testing.T) {
		anon := platformv1connect.NewTenantEntitlementServiceClient(http.DefaultClient, srv.URL+"/api/v1")
		_, err := anon.GetTenantEntitlements(t.Context(), connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("未登入應 unauthenticated，got %v", err)
		}
	})
}

// seedTenantEntitlementFixture 種下平台域夾具（admin／owner 連線）：3 個 feature、2 個方案、
// 方案權益與兩家公司的 active 訂閱。
//
// `limit.storage_gb` 刻意**種進去但沒有計數器**（`services.countFeature` 沒有它的分支）：
// 它是「新 feature 忘了配計數器」的真實長相（T11 就是為此把它從 seed 清單移除）。
func seedTenantEntitlementFixture(t *testing.T, db *sql.DB, coA, coB int) {
	t.Helper()
	for _, f := range []struct{ code, typ, unit string }{
		{entitlements.LimitSeats, "integer", "席"},
		{entitlements.LimitStorageGB, "integer", "GB"},
		{entitlements.FeaturePrinting, "boolean", ""},
	} {
		if _, err := db.Exec(`INSERT INTO platform.features (code, type, unit) VALUES ($1,$2,$3)`,
			f.code, f.typ, f.unit); err != nil {
			t.Fatalf("seed feature %s: %v", f.code, err)
		}
	}
	planID := func(code, name string, sortOrder int) int64 {
		var id int64
		if err := db.QueryRow(`INSERT INTO platform.plans (code, name, sort_order)
			VALUES ($1,$2,$3) RETURNING id`, code, name, sortOrder).Scan(&id); err != nil {
			t.Fatalf("seed 方案 %s: %v", code, err)
		}
		return id
	}
	stdID := planID("std", "標準", 1)
	proID := planID("pro", "專業", 2)

	entitlement := func(planID int64, featureCode string, enabled bool, limit *int64) {
		if _, err := db.Exec(`INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value)
			VALUES ($1,$2,$3,$4)`, planID, featureCode, enabled, limit); err != nil {
			t.Fatalf("seed 權益 %s: %v", featureCode, err)
		}
	}
	entitlement(stdID, entitlements.LimitSeats, true, ptr64(10))
	entitlement(stdID, entitlements.LimitStorageGB, true, ptr64(5))
	entitlement(stdID, entitlements.FeaturePrinting, true, nil)
	entitlement(proID, entitlements.LimitSeats, true, ptr64(50))

	subscribe := func(companyID int, planID int64) {
		if _, err := db.Exec(`INSERT INTO platform.subscriptions (company_id, plan_id, status)
			VALUES ($1,$2,'active')`, companyID, planID); err != nil {
			t.Fatalf("seed 訂閱（公司 %d）: %v", companyID, err)
		}
	}
	subscribe(coA, stdID)
	subscribe(coB, proID)
}

// ptr64 取 *int64（平台域的可空限額欄位：nil = 不限，與 0 不同）。
func ptr64(v int64) *int64 { return &v }

// usageOfProto 依 feature code 取用量列（找不到回 nil）。
func usageOfProto(resp *platformv1.GetTenantEntitlementsResponse, code string) *platformv1.Usage {
	for _, u := range resp.GetUsage() {
		if u.GetFeatureCode() == code {
			return u
		}
	}
	return nil
}
