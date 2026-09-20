package entitlements_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

const seats = entitlements.LimitSeats

// errorInfoOf 由 connect error 取 ErrorInfo（本套件的唯一解析點；與 internal/services 的
// errorInfoOf 同構——對外碼才是前端據以導向升級方案的依據）。
func errorInfoOf(t *testing.T, err error) *commonv1.ErrorInfo {
	t.Helper()
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("非 connect error: %v", err)
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			t.Fatalf("detail 取值失敗: %v", derr)
		}
		if info, ok := v.(*commonv1.ErrorInfo); ok {
			return info
		}
	}
	t.Fatalf("錯誤未帶 ErrorInfo（碼到不了客戶端）: %v", err)
	return nil
}

func errorCodeOf(t *testing.T, err error) string { return errorInfoOf(t, err).GetCode() }

func newSvc(f *store.Fake, counts map[string]int) *entitlements.Service {
	return entitlements.New(f, counting(counts), entitlements.NewMemoryCache(), 0)
}

type counting map[string]int

func (c counting) Count(_ context.Context, _ int, feature string) (int, error) {
	return c[feature], nil
}

func TestNoSubscriptionDeniesEverything(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10))}})
	got, err := newSvc(f, nil).Allows(context.Background(), 1, seats)
	if err != nil {
		t.Fatalf("Allows: %v", err)
	}
	if got {
		t.Fatal("無訂閱時必須 fail-closed（不得因為方案有定義就放行）")
	}
}

func TestLimitBlockedReturnsFailedPrecondition(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})

	svc := newSvc(f, counting{seats: 10})
	err := svc.CheckLimit(context.Background(), 1, seats, 1)
	if err == nil {
		t.Fatal("已達上限時新增應被擋")
	}
	var cerr *connect.Error
	if !errors.As(err, &cerr) || cerr.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("額度不足必須回 FailedPrecondition（引導升級方案），got %v", err)
	}
	// 對外碼才是前端據以導向升級方案的依據（connect 碼不足以區分額度與權限）。
	if got := errorCodeOf(t, err); got != "PLAT-5001" {
		t.Fatalf("額度不足必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}

	ok := newSvc(f, counting{seats: 9})
	if err := ok.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("未達上限應放行: %v", err)
	}
}

func TestOverrideBeatsPlanAndExpiredOverrideIgnored(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})
	past := time.Now().Add(-time.Hour)
	f.PutOverride(store.Override{CompanyID: 1, FeatureCode: seats, Limit: ptr(int64(50)), ExpiresAt: &past})

	if err := newSvc(f, counting{seats: 20}).CheckLimit(context.Background(), 1, seats, 1); err == nil {
		t.Fatal("已過期的 override 不得生效（應回到方案的 10）")
	}

	f.PutOverride(store.Override{CompanyID: 1, FeatureCode: seats, Limit: ptr(int64(50))})
	if err := newSvc(f, counting{seats: 20}).CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("未過期的 override 應放行（50）: %v", err)
	}
}

func TestSuspendedSubscriptionDenied(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: entitlements.FeaturePrinting, Type: "boolean"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: entitlements.FeaturePrinting, Enabled: true}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "suspended"})

	got, err := newSvc(f, nil).Allows(context.Background(), 1, entitlements.FeaturePrinting)
	if err != nil {
		t.Fatalf("Allows: %v", err)
	}
	if got {
		t.Fatal("訂閱 suspended 時不得允許使用")
	}
}

var (
	seatsDef = store.Feature{Code: entitlements.LimitSeats, Type: "integer", Unit: "席"}
	printDef = store.Feature{Code: entitlements.FeaturePrinting, Type: "boolean"}

	stdPlan = []store.Entitlement{
		{FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(10))},
		{FeatureCode: entitlements.FeaturePrinting, Enabled: true},
	}
)

func subWithStatus(status string) *store.Subscription {
	return &store.Subscription{CompanyID: 1, PlanCode: "std", Status: status}
}

// TestJudgementTable 是本套件的主表：逐列固定「買了沒有／額度夠不夠」的判定語意
// （fail-closed、優先序 override > 方案 > 預設、訂閱不可用、兩個對外碼與其 details）。
func TestJudgementTable(t *testing.T) {
	expired := time.Now().Add(-time.Hour)
	tests := []struct {
		name        string
		features    []store.Feature
		ents        []store.Entitlement
		sub         *store.Subscription
		overrides   []store.Override
		counts      map[string]int
		feature     string
		delta       int
		wantAllows  bool
		wantCode    string            // 空＝CheckLimit 必須放行
		wantDetails map[string]string // 非空時逐鍵斷言 details
	}{
		{
			name:     "未定義的功能一律拒絕（fail-closed，不是放行）",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			feature:    "feature.unknown",
			wantAllows: false, wantCode: "PLAT-5002",
			wantDetails: map[string]string{"feature": "feature.unknown"},
		},
		{
			name:     "無訂閱：方案有定義也不放行",
			features: []store.Feature{seatsDef}, ents: stdPlan,
			feature:    seats,
			wantAllows: false, wantCode: "PLAT-5002",
			wantDetails: map[string]string{"feature": seats},
		},
		{
			name:     "查無方案（訂閱指向不存在的方案）",
			features: []store.Feature{seatsDef}, sub: subWithStatus("active"),
			feature:    seats,
			wantAllows: false, wantCode: "PLAT-5002",
			wantDetails: map[string]string{"feature": seats},
		},
		{
			name:     "active：方案含 boolean 功能即允許",
			features: []store.Feature{printDef}, ents: stdPlan, sub: subWithStatus("active"),
			feature:    entitlements.FeaturePrinting,
			wantAllows: true,
		},
		{
			name:     "方案把 boolean 設為 disabled 時不得使用",
			features: []store.Feature{printDef},
			ents:     []store.Entitlement{{FeatureCode: entitlements.FeaturePrinting, Enabled: false}},
			sub:      subWithStatus("active"),
			feature:  entitlements.FeaturePrinting,
			// 功能存在但方案未開 → 未含此功能
			wantAllows: false, wantCode: "PLAT-5002",
			wantDetails: map[string]string{"feature": entitlements.FeaturePrinting},
		},
		{
			name:     "額度邊界：used+delta == limit 仍放行",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			counts: map[string]int{seats: 9}, feature: seats, delta: 1,
			wantAllows: true,
		},
		{
			name:     "額度邊界：delta 0 在滿額時放行",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			counts: map[string]int{seats: 10}, feature: seats, delta: 0,
			wantAllows: true,
		},
		{
			name:     "達上限 → PLAT-5001，details 帶 feature／used／limit",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			counts: map[string]int{seats: 10}, feature: seats, delta: 1,
			// 功能本身可用（Allows 為真）；擋的是「再加一個就超額」。
			wantAllows:  true,
			wantCode:    "PLAT-5001",
			wantDetails: map[string]string{"feature": seats, "used": "10", "limit": "10"},
		},
		{
			name:     "override 只覆寫限額（Enabled=nil）不得把功能關掉",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			overrides: []store.Override{{CompanyID: 1, FeatureCode: seats, Limit: ptr(int64(50))}},
			counts:    map[string]int{seats: 20}, feature: seats, delta: 1,
			wantAllows: true,
		},
		{
			name:     "override 只覆寫 enabled：Limit=nil 是「不覆寫限額」不是「不限」",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			overrides: []store.Override{{CompanyID: 1, FeatureCode: seats, Enabled: ptr(true)}},
			counts:    map[string]int{seats: 20}, feature: seats, delta: 1,
			wantAllows: true, wantCode: "PLAT-5001",
			wantDetails: map[string]string{"feature": seats, "used": "20", "limit": "10"},
		},
		{
			name:     "override 關掉 boolean 功能",
			features: []store.Feature{printDef}, ents: stdPlan, sub: subWithStatus("active"),
			overrides:  []store.Override{{CompanyID: 1, FeatureCode: entitlements.FeaturePrinting, Enabled: ptr(false)}},
			feature:    entitlements.FeaturePrinting,
			wantAllows: false, wantCode: "PLAT-5002",
			wantDetails: map[string]string{"feature": entitlements.FeaturePrinting},
		},
		{
			name:     "已到期的 override 不生效（回到方案限額）",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("active"),
			overrides: []store.Override{{CompanyID: 1, FeatureCode: seats, Limit: ptr(int64(50)), ExpiresAt: &expired}},
			counts:    map[string]int{seats: 20}, feature: seats, delta: 1,
			wantAllows: true, wantCode: "PLAT-5001",
			wantDetails: map[string]string{"feature": seats, "used": "20", "limit": "10"},
		},
		{
			name:     "suspended：不得使用；CheckLimit 回 PLAT-3001（合約問題，非權限）",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("suspended"),
			feature: seats, delta: 1,
			wantAllows: false, wantCode: "PLAT-3001",
		},
		{
			name:     "cancelled 不會從 store 出來（Fake／SQL 都只回未取消者）→ 視同無訂閱",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("cancelled"),
			feature: seats, delta: 1,
			wantAllows: false, wantCode: "PLAT-5002",
			wantDetails: map[string]string{"feature": seats},
		},
		{
			name:     "suspended 且功能未含方案 → 仍是 PLAT-3001（合約問題優先於功能問題）",
			features: []store.Feature{seatsDef}, sub: subWithStatus("suspended"),
			feature: seats, delta: 1,
			wantAllows: false, wantCode: "PLAT-3001",
		},
		{
			name:     "past_due 仍在寬限內（契約不可用的狀態只有 suspended／cancelled）",
			features: []store.Feature{seatsDef}, ents: stdPlan, sub: subWithStatus("past_due"),
			counts: map[string]int{seats: 1}, feature: seats, delta: 1,
			wantAllows: true,
		},
		{
			name:       "trialing：方案 disabled 不擋試用",
			features:   []store.Feature{printDef},
			ents:       []store.Entitlement{{FeatureCode: entitlements.FeaturePrinting, Enabled: false}},
			sub:        subWithStatus("trialing"),
			feature:    entitlements.FeaturePrinting,
			wantAllows: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := store.NewFake()
			for _, x := range tc.features {
				f.PutFeature(x)
			}
			if tc.ents != nil {
				f.PutPlan("std", tc.ents)
			}
			if tc.sub != nil {
				f.PutSubscription(*tc.sub)
			}
			for _, o := range tc.overrides {
				f.PutOverride(o)
			}
			svc := newSvc(f, tc.counts)
			ctx := context.Background()

			got, err := svc.Allows(ctx, 1, tc.feature)
			if err != nil {
				t.Fatalf("Allows 不得因判定結果回錯（fail-closed 回 false）: %v", err)
			}
			if got != tc.wantAllows {
				t.Fatalf("Allows = %v；want %v", got, tc.wantAllows)
			}

			err = svc.CheckLimit(ctx, 1, tc.feature, tc.delta)
			if tc.wantCode == "" {
				if err != nil {
					t.Fatalf("CheckLimit 應放行: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckLimit 應回 %s", tc.wantCode)
			}
			if got, want := connect.CodeOf(err), connect.CodeFailedPrecondition; got != want {
				t.Fatalf("connect 碼 = %v；want %v（額度／功能問題不是 PermissionDenied）", got, want)
			}
			if got := errorCodeOf(t, err); got != tc.wantCode {
				t.Fatalf("ErrorInfo.code = %q；want %q", got, tc.wantCode)
			}
			details := errorInfoOf(t, err).GetDetails()
			for k, want := range tc.wantDetails {
				if details[k] != want {
					t.Fatalf("details[%q] = %q；want %q（全部 details: %v）", k, details[k], want, details)
				}
			}
		})
	}
}

// stubStore 遮蔽 Fake 的「不吐 cancelled」語意：store 契約上 cancelled 不會出現，
// 但判定層仍須對它回 PLAT-3001（換一個 store 實作就可能看到）。
type stubStore struct {
	store.Store
	sub *store.Subscription
}

func (s stubStore) Subscription(context.Context, int) (*store.Subscription, error) {
	return s.sub, nil
}

func TestCancelledSubscriptionIsContractInactive(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(seatsDef)
	f.PutPlan("std", stdPlan)

	sub := store.Subscription{CompanyID: 1, PlanCode: "std", Status: "cancelled"}
	svc := entitlements.New(stubStore{Store: f, sub: &sub}, counting{seats: 1}, entitlements.NewMemoryCache(), 0)
	ctx := context.Background()

	if ok, err := svc.Allows(ctx, 1, seats); err != nil || ok {
		t.Fatalf("Allows = (%v, %v)；want (false, nil)", ok, err)
	}
	err := svc.CheckLimit(ctx, 1, seats, 1)
	if err == nil {
		t.Fatal("cancelled 訂閱不得放行")
	}
	if got := errorCodeOf(t, err); got != "PLAT-3001" {
		t.Fatalf("ErrorInfo.code = %q；want %q（合約不可用不是「功能未含方案」）", got, "PLAT-3001")
	}
}

func TestUnlimitedAllowsEverything(t *testing.T) {
	svc := entitlements.Unlimited()
	ctx := context.Background()
	if ok, err := svc.Allows(ctx, 1, seats); err != nil || !ok {
		t.Fatalf("Unlimited 的 Allows = (%v, %v)；want (true, nil)", ok, err)
	}
	// 不注入 store／counters 也不得 panic。
	if err := svc.CheckLimit(ctx, 1, seats, 1_000_000); err != nil {
		t.Fatalf("Unlimited 的 CheckLimit 應放行: %v", err)
	}
}

// countingStore 記錄對底層 store 的查詢次數，用來證明第二次判定走的是快取。
type countingStore struct {
	store.Store
	subCalls int
}

func (c *countingStore) Subscription(ctx context.Context, companyID int) (*store.Subscription, error) {
	c.subCalls++
	return c.Store.Subscription(ctx, companyID)
}

func TestCacheServesSecondCallWithoutStoreHit(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(2))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})

	cs := &countingStore{Store: f}
	svc := entitlements.New(cs, counting{seats: 1}, entitlements.NewMemoryCache(), time.Minute)

	if err := svc.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("第一次（未達上限）應放行: %v", err)
	}
	first := cs.subCalls
	if err := svc.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("第二次應放行: %v", err)
	}
	if cs.subCalls != first {
		t.Fatalf("第二次判定應命中快取（store 查詢次數 %d → %d）", first, cs.subCalls)
	}

	// 快取不得讓判定失真：換一個「已達上限」的計數器，結果仍須被擋。
	if err := entitlements.New(cs, counting{seats: 2}, entitlements.NewMemoryCache(), time.Minute).
		CheckLimit(context.Background(), 1, seats, 1); err == nil {
		t.Fatal("2/2 再加 1 應被擋")
	}
}

// ttl <= 0 時不快取：store 每次都被查（避免「快取關不掉」的意外）。
func TestZeroTTLDisablesCache(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(2))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})

	cs := &countingStore{Store: f}
	svc := entitlements.New(cs, counting{seats: 1}, entitlements.NewMemoryCache(), 0)
	for i := 0; i < 2; i++ {
		if err := svc.CheckLimit(context.Background(), 1, seats, 1); err != nil {
			t.Fatalf("第 %d 次應放行: %v", i+1, err)
		}
	}
	if cs.subCalls != 2 {
		t.Fatalf("ttl<=0 時不得快取（store 查詢次數 %d；want 2）", cs.subCalls)
	}
}

func TestMemoryCacheRoundTrip(t *testing.T) {
	c := entitlements.NewMemoryCache()
	ctx := context.Background()

	if _, ok, err := c.Get(ctx, "ent:1"); err != nil || ok {
		t.Fatalf("未寫入即命中: ok=%v err=%v", ok, err)
	}
	if err := c.Set(ctx, "ent:1", []byte(`{"x":1}`), time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok, err := c.Get(ctx, "ent:1")
	if err != nil || !ok {
		t.Fatalf("寫入後應命中: ok=%v err=%v", ok, err)
	}
	if string(got) != `{"x":1}` {
		t.Fatalf("Get = %q；want %q", got, `{"x":1}`)
	}
	if err := c.Delete(ctx, "ent:1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok, _ := c.Get(ctx, "ent:1"); ok {
		t.Fatal("Delete 後不得命中（Plan C 的失效機制靠它）")
	}
}

func ptr[T any](v T) *T { return &v }
