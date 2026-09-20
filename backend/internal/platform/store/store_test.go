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

// 無訂閱列時必須回 (nil, nil):判定層靠它區分「沒訂閱」與「查詢失敗」,
// 回錯誤會讓「公司尚未訂閱」變成 500。
//
// **已取消的訂閱必須照樣回傳**（F-8）：store 不得預先濾掉 cancelled —— 否則「有取消的合約」與
// 「完全沒有合約」在判定層不可區分，而判定層對後者是「尚未開通計費 → 不施加限制」，等於讓取消
// 流程變成送免費方案。判定 cancelled 是判定層的職責（allow-list → PLAT-3001），與 override 的
// 「到期由判定層判斷」同一原則。
func TestFakeStoreSubscriptionMissingIsNilAndCancelledIsReturned(t *testing.T) {
	ctx := context.Background()
	f := store.NewFake()

	sub, err := f.Subscription(ctx, 7)
	if err != nil || sub != nil {
		t.Fatalf("無訂閱列的租戶應回 (nil, nil),got %+v err=%v", sub, err)
	}

	f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "std", Status: "cancelled"})
	sub, err = f.Subscription(ctx, 7)
	if err != nil || sub == nil {
		t.Fatalf("已取消的訂閱必須回傳(不得當成沒有訂閱),got %+v err=%v", sub, err)
	}
	if sub.Status != "cancelled" || sub.PlanCode != "std" {
		t.Fatalf("取消的訂閱必須帶出 status／方案(判定層據以回 PLAT-3001),got %+v", *sub)
	}
}

// 未取消的訂閱要帶出判定層真正會用到的欄位:BillingCycle 決定期別 +1 月還是 +1 年(G1),
// Status 決定是否放行,PlanCode 是取權益的鍵,PlanName 是租戶端投影要顯示的名字。
func TestFakeStoreSubscriptionCarriesBillingCycle(t *testing.T) {
	ctx := context.Background()
	trial := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	f := store.NewFake()
	f.PutSubscription(store.Subscription{
		CompanyID: 7, PlanID: 3, PlanCode: "std", PlanName: "標準", Status: "trialing",
		SeatCount: 10, BillingCycle: "yearly", TrialEnds: &trial,
	})

	sub, err := f.Subscription(ctx, 7)
	if err != nil || sub == nil {
		t.Fatalf("應取得訂閱,got %+v err=%v", sub, err)
	}
	if sub.PlanCode != "std" || sub.Status != "trialing" || sub.BillingCycle != "yearly" {
		t.Fatalf("訂閱的 code／status／billing_cycle 未完整帶出,got %+v", *sub)
	}
	if sub.PlanName != "標準" {
		t.Fatalf("訂閱的方案名未帶出(租戶端投影要顯示它),got %q", sub.PlanName)
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
	// 以功能 code 定址,不靠切片位置:介面從未承諾順序(真 PG 實作反而 ORDER BY created_at
	// DESC),位置斷言會把「假實作的插入順序」變成事實上的契約。
	byFeature := map[string]store.Override{}
	for _, o := range ov {
		byFeature[o.FeatureCode] = o
	}
	if o, ok := byFeature["limit.seats"]; !ok || o.Enabled != nil || o.Limit == nil || *o.Limit != 50 {
		t.Fatalf("只寫限額的例外:enabled 必須保持 nil、limit 必須帶出,got %+v(存在=%v)", o, ok)
	}
	if o, ok := byFeature["feature.printing"]; !ok || o.Enabled == nil || !*o.Enabled ||
		o.ExpiresAt == nil || !o.ExpiresAt.Equal(expired) {
		t.Fatalf("enabled／到期日未帶出,got %+v(存在=%v)", o, ok)
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
	byCode := map[string]store.Entitlement{}
	for _, e := range ents {
		byCode[e.FeatureCode] = e
	}
	if e, ok := byCode["limit.seats"]; !ok || e.Limit == nil || *e.Limit != 10 {
		t.Fatalf("權益限額未帶出,got %+v(存在=%v)", e, ok)
	}
	if e, ok := byCode["feature.printing"]; !ok || e.Limit != nil {
		t.Fatalf("NULL 限額(不限)必須維持 nil,不是 0,got %+v(存在=%v)", e, ok)
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

// 假實作要與 PG 實作一樣「回傳的資料與內部狀態無關」:PG 每列掃描都配置新的指標,呼叫端
// 改動不到 store。若假實作共用指標,`*ov[0].Limit = 999` 會靜默改到之後所有讀取
// (跨測試／跨案例汙染,而判定層的測試正是共用同一份假實作)。Put 與 getter 兩端的承諾
// 必須一致:都不別名呼叫端的指標。
func TestFakeStoreDoesNotAliasPointerFields(t *testing.T) {
	ctx := context.Background()
	limit, enabled := int64(50), true
	expires := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	trial := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)

	f := store.NewFake()
	f.PutOverride(store.Override{
		CompanyID: 7, FeatureCode: "limit.seats", Enabled: &enabled, Limit: &limit, ExpiresAt: &expires,
	})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: "limit.seats", Enabled: true, Limit: &limit}})
	f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "std", Status: "active", TrialEnds: &trial})

	// 放進去之後才改動那組指標(呼叫端擁有的記憶體)。
	limit, enabled = 1, false
	expires, trial = time.Time{}, time.Time{}

	// 讀出來的那組指標也改動一次。
	ov, err := f.Overrides(ctx, 7)
	if err != nil || len(ov) != 1 {
		t.Fatalf("取 overrides: got %+v err=%v", ov, err)
	}
	*ov[0].Limit, *ov[0].Enabled, *ov[0].ExpiresAt = 999, false, time.Time{}
	ents, err := f.PlanEntitlements(ctx, "std")
	if err != nil || len(ents) != 1 {
		t.Fatalf("取方案權益: got %+v err=%v", ents, err)
	}
	*ents[0].Limit = 999
	sub, err := f.Subscription(ctx, 7)
	if err != nil || sub == nil {
		t.Fatalf("取訂閱: got %+v err=%v", sub, err)
	}
	*sub.TrialEnds = time.Time{}

	// 再讀一次:兩端都必須維持原值。
	ov, err = f.Overrides(ctx, 7)
	if err != nil || len(ov) != 1 {
		t.Fatalf("重取 overrides: got %+v err=%v", ov, err)
	}
	if ov[0].Limit == nil || *ov[0].Limit != 50 || ov[0].Enabled == nil || !*ov[0].Enabled ||
		ov[0].ExpiresAt == nil || !ov[0].ExpiresAt.Equal(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("override 的指標欄位被呼叫端改到了,got %+v", ov[0])
	}
	ents, err = f.PlanEntitlements(ctx, "std")
	if err != nil || len(ents) != 1 || ents[0].Limit == nil || *ents[0].Limit != 50 {
		t.Fatalf("權益的限額被呼叫端改到了,got %+v err=%v", ents, err)
	}
	sub, err = f.Subscription(ctx, 7)
	if err != nil || sub == nil || sub.TrialEnds == nil ||
		!sub.TrialEnds.Equal(time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("訂閱的試用到期日被呼叫端改到了,got %+v err=%v", sub, err)
	}
}
