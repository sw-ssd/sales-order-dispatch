package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// 假實作是判定層單元測試的地基:它的過濾語意一旦與 SQL 實作不同,判定層的測試就會失真。
// 例如假實作若「順手」把過期的 override 也濾掉,判定層的到期測試就會假綠 —— 真環境裡的
// 到期判斷是判定層的職責,store 只負責剔除已撤銷者。以下逐一釘住四方法的行為契約。

// 無未取消的訂閱時必須回 (nil, nil):判定層靠它區分「沒訂閱」與「查詢失敗」,
// 回錯誤會讓「公司尚未訂閱」變成 500。
func TestFakeStoreSubscriptionMissingOrCancelledIsNil(t *testing.T) {
	ctx := context.Background()
	f := store.NewFake()

	sub, err := f.Subscription(ctx, 7)
	if err != nil || sub != nil {
		t.Fatalf("無訂閱的租戶應回 (nil, nil),got %+v err=%v", sub, err)
	}

	f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "std", Status: "cancelled"})
	if sub, err := f.Subscription(ctx, 7); err != nil || sub != nil {
		t.Fatalf("已取消的訂閱應視同無訂閱,got %+v err=%v", sub, err)
	}
}

// 未取消的訂閱要帶出判定層真正會用到的欄位:BillingCycle 決定期別 +1 月還是 +1 年(G1),
// Status 決定是否放行,PlanCode 是取權益的鍵。
func TestFakeStoreSubscriptionCarriesBillingCycle(t *testing.T) {
	ctx := context.Background()
	trial := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	f := store.NewFake()
	f.PutSubscription(store.Subscription{
		CompanyID: 7, PlanID: 3, PlanCode: "std", Status: "trialing",
		SeatCount: 10, BillingCycle: "yearly", TrialEnds: &trial,
	})

	sub, err := f.Subscription(ctx, 7)
	if err != nil || sub == nil {
		t.Fatalf("應取得訂閱,got %+v err=%v", sub, err)
	}
	if sub.PlanCode != "std" || sub.Status != "trialing" || sub.BillingCycle != "yearly" {
		t.Fatalf("訂閱的 code／status／billing_cycle 未完整帶出,got %+v", *sub)
	}
	if sub.TrialEnds == nil || !sub.TrialEnds.Equal(trial) {
		t.Fatalf("試用到期日未帶出,got %+v", sub.TrialEnds)
	}
}

// 到期的 override 仍必須回傳(判定層自己判斷到期),且 null 欄位要維持 nil:
// 一筆只寫限額的例外若讓 Enabled 變成 false,就會把功能整個關掉。
func TestFakeStoreOverridesKeepExpiredRows(t *testing.T) {
	ctx := context.Background()
	expired := time.Now().Add(-time.Hour)
	f := store.NewFake()
	f.PutOverride(store.Override{CompanyID: 7, FeatureCode: "limit.seats", Limit: ptr(int64(50))})
	f.PutOverride(store.Override{
		CompanyID: 7, FeatureCode: "feature.printing", Enabled: ptr(true), ExpiresAt: &expired,
	})

	ov, err := f.Overrides(ctx, 7)
	if err != nil {
		t.Fatalf("取 overrides: %v", err)
	}
	if len(ov) != 2 {
		t.Fatalf("到期過濾是判定層的職責,假實作應回全部未撤銷的 override,got %d 筆: %+v", len(ov), ov)
	}
	if ov[0].Enabled != nil || ov[0].Limit == nil || *ov[0].Limit != 50 {
		t.Fatalf("只寫限額的例外:enabled 必須保持 nil、limit 必須帶出,got %+v", ov[0])
	}
	if ov[1].Enabled == nil || !*ov[1].Enabled || ov[1].ExpiresAt == nil || !ov[1].ExpiresAt.Equal(expired) {
		t.Fatalf("enabled／到期日未帶出,got %+v", ov[1])
	}

	if other, err := f.Overrides(ctx, 8); err != nil || len(other) != 0 {
		t.Fatalf("無例外的租戶應回空,got %+v err=%v", other, err)
	}
}

// 方案權益以**方案 code** 取(不是 plan_id),Features 以 code 為鍵;兩者都回副本 ——
// 呼叫端改動不得汙染後續讀取(同一份假實作會被同一套測試的多個案例共用)。
func TestFakeStoreFeaturesAndPlanEntitlementsByCode(t *testing.T) {
	ctx := context.Background()
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: "limit.seats", Type: "integer", Unit: "席", Description: "席位上線"})
	f.PutPlan("std", []store.Entitlement{
		{FeatureCode: "limit.seats", Enabled: true, Limit: ptr(int64(10))},
		{FeatureCode: "feature.printing", Enabled: true, Limit: nil},
	})

	feats, err := f.Features(ctx)
	if err != nil {
		t.Fatalf("取 features: %v", err)
	}
	if feats["limit.seats"].Type != "integer" || feats["limit.seats"].Unit != "席" {
		t.Fatalf("功能定義未完整帶出,got %+v", feats["limit.seats"])
	}

	ents, err := f.PlanEntitlements(ctx, "std")
	if err != nil {
		t.Fatalf("取方案權益: %v", err)
	}
	if len(ents) != 2 {
		t.Fatalf("方案權益應回 2 筆,got %d: %+v", len(ents), ents)
	}
	if ents[0].Limit == nil || *ents[0].Limit != 10 {
		t.Fatalf("權益限額未帶出,got %+v", ents[0])
	}
	if ents[1].Limit != nil {
		t.Fatalf("NULL 限額(不限)必須維持 nil,不是 0,got %+v", ents[1])
	}
	if unknown, err := f.PlanEntitlements(ctx, "no_such_plan"); err != nil || len(unknown) != 0 {
		t.Fatalf("未知方案應回空,got %+v err=%v", unknown, err)
	}

	feats["limit.seats"] = store.Feature{Code: "limit.seats", Type: "汙染"}
	if again, err := f.Features(ctx); err != nil || again["limit.seats"].Type != "integer" {
		t.Fatalf("Features 必須回副本:呼叫端的改動不得影響後續讀取,got %+v err=%v", again, err)
	}
}

func ptr[T any](v T) *T { return &v }
