package services

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
)

// 本檔為 PlatformAdminService 的**服務層**契約(免容器的假 store):五個 RPC 各自的 operator
// 檢查、參數驗證與錯誤碼映射。真 SQL 的欄位對應由 platform_admin_integration_test.go 守。
//
// 為什麼要在服務層再檢查一次(interceptor 已經擋了):服務是普通的 Go 方法,可以被別的掛載
// 路徑、CLI 或未來的批次路徑直接呼叫;把唯一邊界押在 interceptor 上,等於「邊界隨掛載方式
// 改變」——而這裡漏掉的後果是跨租戶資料被非 operator 讀走。

// platformAdminCalls 為五個 RPC 的呼叫器(依固定順序)。
//
// 以清單驅動的理由:授權邊界必須**每個 RPC 都成立**。逐一複製三行走同一個斷言只會讓「新增
// RPC 時忘了加檢查」在評審眼裡看不出來——漏一個就是少一行,而這裡漏一個會少一個 case。
func platformAdminCalls() []struct {
	name string
	call func(ctx context.Context, svc *PlatformAdminService) error
} {
	return []struct {
		name string
		call func(ctx context.Context, svc *PlatformAdminService) error
	}{
		{"ListTenants", func(ctx context.Context, svc *PlatformAdminService) error {
			_, err := svc.ListTenants(ctx, connect.NewRequest(&platformv1.ListTenantsRequest{Page: 1, PageSize: 20}))
			return err
		}},
		{"GetTenant", func(ctx context.Context, svc *PlatformAdminService) error {
			_, err := svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{CompanyId: "7"}))
			return err
		}},
		{"ListPlans", func(ctx context.Context, svc *PlatformAdminService) error {
			_, err := svc.ListPlans(ctx, connect.NewRequest(&platformv1.ListPlansRequest{}))
			return err
		}},
		{"GetPlanEntitlements", func(ctx context.Context, svc *PlatformAdminService) error {
			_, err := svc.GetPlanEntitlements(ctx, connect.NewRequest(&platformv1.GetPlanEntitlementsRequest{PlanCode: "std"}))
			return err
		}},
		{"ListPlatformAudit", func(ctx context.Context, svc *PlatformAdminService) error {
			_, err := svc.ListPlatformAudit(ctx, connect.NewRequest(&platformv1.ListPlatformAuditRequest{Page: 1, PageSize: 20}))
			return err
		}},
	}
}

// platformAdminFixture 為本檔共用的假資料:一家有訂閱的租戶＋一家沒訂閱的租戶。
func platformAdminFixture() *fakePlatformStore {
	periodEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	return &fakePlatformStore{
		tenants: []TenantRow{
			{CompanyID: "7", CompanyName: "測試公司", PlanCode: "std", PlanName: "標準",
				Status: "active", SeatCount: 8, CurrentPeriodEnd: &periodEnd, Overdue: true},
			{CompanyID: "8", CompanyName: "未訂閱公司", Status: "none"},
		},
		tenant: &TenantRow{CompanyID: "7", CompanyName: "測試公司", PlanCode: "std", PlanName: "標準",
			Status: "active", SeatCount: 8, CurrentPeriodEnd: &periodEnd},
		overrides: []TenantOverrideRow{{ID: "3", FeatureCode: "feature.printing", Reason: "簽約承諾", Owner: "業務"}},
		plans: []PlanRow{{ID: "1", Code: "std", Name: "標準", Status: "active", SortOrder: 1,
			Prices: []PlanPriceRow{{BillingCycle: "monthly", BasePrice: "1000.00", SeatPrice: "200.00",
				Currency: "TWD", EffectiveFrom: periodEnd}}}},
		ents:     []FeatureEntitlementRow{{FeatureCode: "limit.seats", Enabled: true}},
		features: []FeatureRow{{Code: "limit.seats", Type: "integer", Unit: "席", Description: "席位上線"}},
		audit: []PlatformAuditRow{{ID: "5", OperatorEmail: "ops@example.com", Action: "login",
			TargetType: "operator", TargetID: "1", CreatedAt: periodEnd}},
	}
}

// TestPlatformAdminRequiresOperatorIdentity 驗五個 RPC 的 operator 邊界:無身分與**租戶身分**
// 都必須是 Unauthenticated(AUTH-4001),且擋下時**不得查詢任何平台資料**。
//
// 租戶身分那組是本服務的關鍵語意:租戶 session／JWT 與平台 token 互不通用(不同 cookie、
// 不同 secret、不同 audience)。服務層若只認「ctx 裡有身分」,租戶使用者就讀得到跨租戶視圖。
func TestPlatformAdminRequiresOperatorIdentity(t *testing.T) {
	denied := map[string]context.Context{
		"無身分": context.Background(),
		"租戶身分": authz.WithIdentity(context.Background(), authz.Identity{
			UserID: "1", CompanyID: "7", Role: "company_admin", Roles: []string{"company_admin"},
		}),
	}
	for ctxName, ctx := range denied {
		for _, c := range platformAdminCalls() {
			t.Run(ctxName+"/"+c.name, func(t *testing.T) {
				fake := platformAdminFixture()
				svc := NewPlatformAdminService(fake)

				err := c.call(ctx, svc)
				if connect.CodeOf(err) != connect.CodeUnauthenticated {
					t.Fatalf("%s 必須 Unauthenticated,got %v", ctxName, err)
				}
				if got := errorInfoOf(t, err).GetCode(); got != "AUTH-4001" {
					t.Fatalf("必須是註冊碼 AUTH-4001,got %q", got)
				}
				if fake.calls != 0 {
					t.Fatalf("身分檢查必須在查詢之前（查了 %d 次）", fake.calls)
				}
			})
		}
	}

	opCtx := operatorauth.WithIdentity(context.Background(),
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	for _, c := range platformAdminCalls() {
		t.Run("operator/"+c.name, func(t *testing.T) {
			fake := platformAdminFixture()
			svc := NewPlatformAdminService(fake)

			if err := c.call(opCtx, svc); err != nil {
				t.Fatalf("operator 身分應可讀取: %v", err)
			}
			if fake.calls != 1 {
				t.Fatalf("每個 RPC 應查詢 store 一次,got %d", fake.calls)
			}
		})
	}
}

// TestPlatformAdminListTenantsMapsSummary 驗租戶列表的投影:分頁回音、名稱／方案／狀態／席位
// 與**可空的期別到期日**(沒有期別 → 空字串,不是零值時間)、未訂閱租戶的狀態為 none。
//
// 分頁用 0（未指定）:全 domain 的慣例是 page<1→1、page_size<1→預設 20,回應帶**正規化後**的
// 值才不會讓前端算出錯誤的頁數。
func TestPlatformAdminListTenantsMapsSummary(t *testing.T) {
	svc := NewPlatformAdminService(platformAdminFixture())
	ctx := operatorauth.WithIdentity(context.Background(),
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})

	resp, err := svc.ListTenants(ctx, connect.NewRequest(&platformv1.ListTenantsRequest{}))
	if err != nil {
		t.Fatalf("operator 身分應可讀取: %v", err)
	}
	if got := resp.Msg.GetPagination().GetTotal(); got != 2 {
		t.Fatalf("fixture 有 2 家租戶,got %d", got)
	}
	if got := resp.Msg.GetPagination().GetPage(); got != 1 {
		t.Fatalf("page 未指定應正規化為 1,got %d", got)
	}
	if got := resp.Msg.GetPagination().GetPageSize(); got != int32(defaultPageSize) {
		t.Fatalf("page_size 未指定應正規化為 %d,got %d", defaultPageSize, got)
	}

	got := resp.Msg.GetTenants()
	if len(got) != 2 {
		t.Fatalf("fixture 有 2 家租戶,got %d", len(got))
	}
	first := got[0]
	if first.GetCompanyId() != "7" || first.GetCompanyName() != "測試公司" {
		t.Fatalf("租戶識別映射錯誤: %v", first)
	}
	if first.GetPlanCode() != "std" || first.GetPlanName() != "標準" ||
		first.GetSubscriptionStatus() != "active" || first.GetSeatCount() != 8 || !first.GetOverdue() {
		t.Fatalf("租戶方案／狀態／席位映射錯誤: %v", first)
	}
	if first.GetCurrentPeriodEnd() != "2026-10-31T00:00:00Z" {
		t.Fatalf("期別到期日必須是 RFC3339, got %q", first.GetCurrentPeriodEnd())
	}
	if second := got[1]; second.GetSubscriptionStatus() != "none" || second.GetCurrentPeriodEnd() != "" {
		t.Fatalf("未訂閱租戶應為 status=none 且無到期日: %v", second)
	}
}

// TestPlatformAdminRejectsInvalidArguments 驗服務層的參數驗證:SYS-1001 且**不查詢** store。
// company_id 是 proto 的 string（bigint 的文字形）,非數字不得落到 SQL 當成 0 或報 500。
func TestPlatformAdminRejectsInvalidArguments(t *testing.T) {
	ctx := operatorauth.WithIdentity(context.Background(),
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})

	tests := map[string]func(svc *PlatformAdminService) error{
		"company_id 非數字": func(svc *PlatformAdminService) error {
			_, err := svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{CompanyId: "abc"}))
			return err
		},
		"company_id 為空": func(svc *PlatformAdminService) error {
			_, err := svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{}))
			return err
		},
		"plan_code 為空": func(svc *PlatformAdminService) error {
			_, err := svc.GetPlanEntitlements(ctx, connect.NewRequest(&platformv1.GetPlanEntitlementsRequest{}))
			return err
		},
	}
	for name, call := range tests {
		t.Run(name, func(t *testing.T) {
			fake := platformAdminFixture()
			svc := NewPlatformAdminService(fake)

			err := call(svc)
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("應回 InvalidArgument,got %v", err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "SYS-1001" {
				t.Fatalf("必須是註冊碼 SYS-1001,got %q", got)
			}
			if fake.calls != 0 {
				t.Fatalf("參數不合法不得查詢 store（查了 %d 次）", fake.calls)
			}
		})
	}
}

// TestPlatformAdminMapsNotFound 驗查無租戶／方案 → SYS-4002(NotFound):operator 之外沒有別人
// 進得來,故「不存在」不必與「無權」合併,前端才能顯示「查無此租戶」而不是「系統忙碌」。
func TestPlatformAdminMapsNotFound(t *testing.T) {
	ctx := operatorauth.WithIdentity(context.Background(),
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})

	calls := map[string]func(svc *PlatformAdminService) error{
		"GetTenant": func(svc *PlatformAdminService) error {
			_, err := svc.GetTenant(ctx, connect.NewRequest(&platformv1.GetTenantRequest{CompanyId: "7"}))
			return err
		},
		"GetPlanEntitlements": func(svc *PlatformAdminService) error {
			_, err := svc.GetPlanEntitlements(ctx, connect.NewRequest(&platformv1.GetPlanEntitlementsRequest{PlanCode: "nope"}))
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			fake := platformAdminFixture()
			fake.err = platformstore.ErrNotFound
			svc := NewPlatformAdminService(fake)

			err := call(svc)
			if connect.CodeOf(err) != connect.CodeNotFound {
				t.Fatalf("查無資料應回 NotFound,got %v", err)
			}
			if got := errorInfoOf(t, err).GetCode(); got != "SYS-4002" {
				t.Fatalf("必須是註冊碼 SYS-4002,got %q", got)
			}
		})
	}
}

// TestPlatformAdminRecordAuditKeepsOperatorActor 驗平台稽核的參數映射:actor 是** operator id**
// (v1 只有讀取 RPC,寫入路徑由後續任務接上;映射若把租戶 user id 當 actor,平台稽核就失去意義)。
func TestPlatformAdminRecordAuditKeepsOperatorActor(t *testing.T) {
	fake := platformAdminFixture()
	svc := NewPlatformAdminService(fake)
	identity := operatorauth.Identity{OperatorID: 42, Email: "ops@example.com", Role: "admin"}

	before := map[string]any{"role": "operator"}
	after := map[string]any{"role": "admin"}
	if err := svc.recordPlatformAudit(context.Background(), identity,
		"operator.update", "operator", "9", "升為管理員", before, after); err != nil {
		t.Fatalf("寫入稽核: %v", err)
	}

	if len(fake.recorded) != 1 {
		t.Fatalf("應寫入 1 筆稽核,got %d", len(fake.recorded))
	}
	got := fake.recorded[0]
	if got.operatorID != 42 {
		t.Fatalf("actor 必須是 operator id（42）,got %d", got.operatorID)
	}
	if got.action != "operator.update" || got.targetType != "operator" || got.targetID != "9" ||
		got.reason != "升為管理員" {
		t.Fatalf("稽核欄位映射錯誤: %+v", got)
	}
	if got.before["role"] != "operator" || got.after["role"] != "admin" {
		t.Fatalf("before／after 未原樣帶入: %+v / %+v", got.before, got.after)
	}
}

// fakePlatformStore 只實作本測試需要的查詢;其餘回傳零值(介面契約由編譯器保證)。
//
// err 為一體適用的錯誤注入:本檔驗的是服務層的錯誤映射,不是各方法的個別錯誤路徑(SQL 的
// 真實行為由整合測試守),故不需要每個方法一份錯誤欄位。
type fakePlatformStore struct {
	tenants   []TenantRow
	tenant    *TenantRow
	overrides []TenantOverrideRow
	plans     []PlanRow
	ents      []FeatureEntitlementRow
	features  []FeatureRow
	audit     []PlatformAuditRow
	err       error

	// calls 為 store 被查詢的次數:「擋下來了」不能只是回應長得像——資料庫早已被讀過一遍
	// 也算越界,故授權／參數檢查必須在查詢之前。
	calls    int
	recorded []recordedAudit
}

type recordedAudit struct {
	operatorID                           int64
	action, targetType, targetID, reason string
	before, after                        map[string]any
}

func (f *fakePlatformStore) ListTenants(context.Context, string, string, int32, int32) ([]TenantRow, int, error) {
	f.calls++
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.tenants, len(f.tenants), nil
}

func (f *fakePlatformStore) GetTenant(context.Context, string) (*TenantRow, []TenantOverrideRow, error) {
	f.calls++
	if f.err != nil {
		return nil, nil, f.err
	}
	return f.tenant, f.overrides, nil
}

func (f *fakePlatformStore) ListPlans(context.Context) ([]PlanRow, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.plans, nil
}

func (f *fakePlatformStore) PlanEntitlements(context.Context, string) ([]FeatureEntitlementRow, []FeatureRow, error) {
	f.calls++
	if f.err != nil {
		return nil, nil, f.err
	}
	return f.ents, f.features, nil
}

func (f *fakePlatformStore) ListPlatformAudit(context.Context, string, string, int32, int32) ([]PlatformAuditRow, int, error) {
	f.calls++
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.audit, len(f.audit), nil
}

func (f *fakePlatformStore) RecordAudit(_ context.Context, operatorID int64,
	action, targetType, targetID, reason string, before, after map[string]any) error {
	f.recorded = append(f.recorded, recordedAudit{
		operatorID: operatorID, action: action, targetType: targetType, targetID: targetID,
		reason: reason, before: before, after: after,
	})
	return f.err
}
