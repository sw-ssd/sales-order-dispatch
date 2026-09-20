// 租戶端權益投影 RPC（GetTenantEntitlements）的契約測試：假 store ＋ 假計數器，免容器。
//
// 釘住三件事（spec §4.6）：
//  1. **租戶身分**才讀得到，且**只回自己公司**（本 RPC 沒有 company_id 參數，這就是契約）；
//  2. 若某個 integer feature **沒有對應的計數器**（或計數失敗），**略過該筆用量**並記 log ——
//     不得回 0／-1，也不得讓整筆投影失敗（controller 裁定：一個漏配計數器的新 feature 不得
//     升級成全站權益頁掛掉）；
//  3. 營運／平台層身分不得走這條（與 PlatformAdminService 相反）。
package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
)

// seatCounter 回傳固定用量（模擬「公司 42 已有 3 席」）。
type seatCounter struct{ used int }

func (c seatCounter) Count(context.Context, int, string) (int, error) { return c.used, nil }

// missingCounter 模擬「某個 integer feature 沒有對應的計數器」：比照 services.countFeature 的
// default 分支 —— 未知 feature **回錯誤，不回 0**（0 會被前端讀成「用量為零」）。
type missingCounter struct {
	missing string
	used    int
}

func (c missingCounter) Count(_ context.Context, _ int, feature string) (int, error) {
	if feature == c.missing {
		return 0, errors.New("未定義的計數 feature: " + feature)
	}
	return c.used, nil
}

// tenantFixture 組出「公司 42＝std」情境：seats（上限 10）＋ storage_gb（上限 5）＋
// printing（boolean）；**另一家公司 7 是 pro** —— 用來釘住「只回自己公司」。
func tenantFixture(counter entitlements.Counter) *TenantEntitlementService {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: entitlements.LimitSeats, Type: "integer", Unit: "席"})
	f.PutFeature(store.Feature{Code: entitlements.LimitStorageGB, Type: "integer", Unit: "GB"})
	f.PutFeature(store.Feature{Code: entitlements.FeaturePrinting, Type: "boolean"})
	f.PutPlan("std", []store.Entitlement{
		{FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(10))},
		{FeatureCode: entitlements.LimitStorageGB, Enabled: true, Limit: ptr(int64(5))},
		{FeatureCode: entitlements.FeaturePrinting, Enabled: true},
	})
	f.PutPlan("pro", []store.Entitlement{{FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(50))}})
	f.PutSubscription(store.Subscription{CompanyID: 42, PlanCode: "std", PlanName: "標準", Status: "active"})
	f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "pro", PlanName: "專業", Status: "active"})
	return NewTenantEntitlementService(entitlements.New(f, counter, entitlements.NewMemoryCache(), 0))
}

// tenantCtx 為「公司 42 的 company_admin」身分（身分即公司範圍）。
func tenantCtx() context.Context {
	return authz.WithIdentity(context.Background(), authz.Identity{
		UserID: "1", CompanyID: "42", Role: "company_admin", Roles: []string{"company_admin"},
	})
}

// usageOf 依 feature code 取用量列（找不到回 nil）。
func usageOf(resp *platformv1.GetTenantEntitlementsResponse, code string) *platformv1.Usage {
	for _, u := range resp.GetUsage() {
		if u.GetFeatureCode() == code {
			return u
		}
	}
	return nil
}

func TestTenantEntitlementsScopedToOwnCompany(t *testing.T) {
	h := tenantFixture(seatCounter{used: 3})

	// ① 未登入：AUTH-4001（Unauthenticated），不查 store。
	resp, err := h.GetTenantEntitlements(context.Background(),
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated || errorInfoOf(t, err).GetCode() != "AUTH-4001" {
		t.Fatalf("未登入必須 AUTH-4001/Unauthenticated，got %v（resp=%v）", err, resp)
	}

	// ② 已登入：只回**自己公司**的方案（公司 7 的 pro 不得出現；本 RPC 無參數可指定他公司）。
	resp, err = h.GetTenantEntitlements(tenantCtx(),
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}))
	if err != nil {
		t.Fatalf("已登入應可讀取: %v", err)
	}
	if resp.Msg.GetPlanCode() != "std" || resp.Msg.GetPlanName() != "標準" || resp.Msg.GetStatus() != "active" {
		t.Fatalf("方案應為自己公司的 std／標準／active，got %q／%q／%q",
			resp.Msg.GetPlanCode(), resp.Msg.GetPlanName(), resp.Msg.GetStatus())
	}

	// ③ 用量與計數器一致；boolean 不給用量（不帶 limit）。
	seats, printing := usageOf(resp.Msg, entitlements.LimitSeats), usageOf(resp.Msg, entitlements.FeaturePrinting)
	if seats == nil || printing == nil {
		t.Fatalf("usage 應含 limit.seats 與 feature.printing，got %d 項", len(resp.Msg.GetUsage()))
	}
	if seats.GetUsed() != 3 || seats.GetLimitValue() != 10 || !seats.GetLimitSet() {
		t.Fatalf("席位用量應為 3/10，got %d/%d（limit_set=%v）",
			seats.GetUsed(), seats.GetLimitValue(), seats.GetLimitSet())
	}
	if !printing.GetEnabled() || printing.GetLimitSet() {
		t.Fatalf("boolean 功能應 enabled 且不帶 limit，got enabled=%v limit_set=%v",
			printing.GetEnabled(), printing.GetLimitSet())
	}
}

// TestTenantEntitlementsSkipsIntegerFeatureWithoutCounter 是 controller 裁定的那條：
// **有 feature 但沒有計數器**（未來任何新 feature 忘了配計數器）時，只略過該筆用量並記 log ——
// 不得回 0（會被讀成「用量為零」）、不得回 -1、不得讓整筆 Snapshot 變成 SysInternal。
func TestTenantEntitlementsSkipsIntegerFeatureWithoutCounter(t *testing.T) {
	h := tenantFixture(missingCounter{missing: entitlements.LimitStorageGB, used: 3})

	resp, err := h.GetTenantEntitlements(tenantCtx(),
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}))
	if err != nil {
		t.Fatalf("缺計數器不得讓整筆投影失敗（否則一個漏配的 feature 就讓所有租戶的權益頁掛掉）: %v", err)
	}
	if resp.Msg.GetPlanCode() != "std" || resp.Msg.GetStatus() != "active" {
		t.Fatalf("方案／狀態仍必須投影出來，got %q／%q", resp.Msg.GetPlanCode(), resp.Msg.GetStatus())
	}
	if got := usageOf(resp.Msg, entitlements.LimitStorageGB); got != nil {
		t.Fatalf("缺計數器的 feature 必須略過該筆用量（不得回 0／-1），got used=%d limit_set=%v",
			got.GetUsed(), got.GetLimitSet())
	}
	seats, printing := usageOf(resp.Msg, entitlements.LimitSeats), usageOf(resp.Msg, entitlements.FeaturePrinting)
	if seats == nil || printing == nil || len(resp.Msg.GetUsage()) != 2 {
		t.Fatalf("其餘 feature 必須照常投影（seats／printing 各一筆），got %+v", resp.Msg.GetUsage())
	}
	if seats.GetUsed() != 3 || seats.GetLimitValue() != 10 {
		t.Fatalf("席位用量應為 3/10，got %d/%d", seats.GetUsed(), seats.GetLimitValue())
	}
	if !printing.GetEnabled() || printing.GetLimitSet() {
		t.Fatalf("boolean 功能應 enabled 且不帶 limit，got enabled=%v limit_set=%v",
			printing.GetEnabled(), printing.GetLimitSet())
	}
}

// TestTenantEntitlementsRejectsPlatformIdentity 釘住硬要求①的相反面：營運工具的身分（operator
// token）與平台層身分（super／developer，無租戶範圍）都不得經這條讀到租戶權益。
// 與 PlatformAdminService 剛好相反：那裡租戶身分一律拒絕。
func TestTenantEntitlementsRejectsPlatformIdentity(t *testing.T) {
	h := tenantFixture(seatCounter{used: 3})

	// operator（平台工具身分）在租戶路徑上沒有 authz 身分 → 未登入（AUTH-4001）。
	opCtx := operatorauth.WithIdentity(context.Background(),
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "operator"})
	if _, err := h.GetTenantEntitlements(opCtx,
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("operator 身分不得讀租戶權益，want Unauthenticated，got %v", err)
	}

	// platform 層身分（super，無租戶範圍）→ fail-closed 拒絕，且不得回任何 200 內容（含空 plan）。
	superCtx := authz.WithIdentity(context.Background(), authz.Identity{UserID: "9", Roles: []string{"super"}})
	resp, err := h.GetTenantEntitlements(superCtx,
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}))
	if connect.CodeOf(err) != connect.CodePermissionDenied || errorInfoOf(t, err).GetCode() != "SYS-4001" {
		t.Fatalf("無租戶範圍的平台層身分必須 fail-closed（SYS-4001/PermissionDenied），got %v（resp=%v）", err, resp)
	}
}

// TestTenantEntitlementsMountedOnTenantMux 為掛載契約：本 RPC 的 procedure 由**租戶 mux**
// （server 掛在 /api/v1 之下）承接，不是平台工具的 /platform/（那裡是 operator cookie 與
// operator interceptor 的範圍，租戶 session 到不了）。
func TestTenantEntitlementsMountedOnTenantMux(t *testing.T) {
	mux := http.NewServeMux()
	RegisterTenantEntitlementService(mux, nil, entitlements.Unlimited())

	req := httptest.NewRequest(http.MethodPost,
		platformv1connect.TenantEntitlementServiceGetTenantEntitlementsProcedure, nil)
	if _, pattern := mux.Handler(req); pattern == "" {
		t.Fatalf("租戶 mux 未註冊 %s（服務有實作但沒掛＝功能消失）",
			platformv1connect.TenantEntitlementServiceGetTenantEntitlementsProcedure)
	}
}
