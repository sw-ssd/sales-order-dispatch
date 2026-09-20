//go:build integration

package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationPlatformStore 以真 PostgreSQL 驗證 store 的四項唯讀查詢與 00029 的
// 表名／欄位逐字對應 —— 欄位錯位(把 unit 掃進 description)、表名寫錯、少帶一欄,在記憶體
// 假實作上永遠測不出來。
//
// 種子刻意涵蓋四個掃描容易寫錯的地方:
//   - plan_entitlements.limit_value 的 NULL(不限)必須掃成 nil,不是 0;
//   - tenant_overrides.enabled 的 NULL(只覆寫限額)必須掃成 nil,不是 false —— 否則一筆
//     只寫限額的例外會把功能整個關掉;
//   - subscriptions.billing_cycle／trial_ends_at／grace_until 必須真的從資料帶出
//     (年繳方案的期別要 +1 年,G1);
//   - 已撤銷(revoked_at)的例外不得回傳;已到期的例外**必須**回傳(到期由判定層處理);
//   - 已取消(status = 'cancelled')的訂閱不得當成現行訂閱(partial unique index 只管未
//     取消者,故同一公司可同時有 active 與 cancelled 列)。
func TestIntegrationPlatformStore(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.RunContext(t.Context(), "up", db, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	ctx := t.Context()
	trialEnds := time.Now().Add(7 * 24 * time.Hour).Truncate(time.Millisecond)
	graceUntil := time.Now().Add(14 * 24 * time.Hour).Truncate(time.Millisecond)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.features (code, type, unit, description) VALUES
		('limit.seats','integer','席','席位上線'),
		('feature.printing','boolean','','列印')`); err != nil {
		t.Fatalf("seed features: %v", err)
	}
	var planID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value) VALUES
		($1,'limit.seats',true,10), ($1,'feature.printing',true,NULL)`, planID); err != nil {
		t.Fatalf("seed entitlements: %v", err)
	}
	// 42 為月繳(用預設值)、無試用／寬限;44 為年繳且兩者有值 —— 兩者相反,才驗得出欄位確實
	//「從資料來」而不是常數或零值。
	//
	// 43／45 與 42 的第二列專為「未取消」這條過濾而設:partial unique index 只管未取消者,
	// 故同一公司可以同時有 active 與 cancelled 列。43 是**唯一一列就是 cancelled** 的租戶 ——
	// 少了 `status <> 'cancelled'` 就一定會回那一列(不依賴列的實體順序,是這條過濾的守門);
	// 42 的 cancelled 列則驗「兩列並存時挑的是 active 那一列」;45 完全沒有列,走 ErrNoRows。
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count) VALUES
		(42, $1, 'active', 10),
		(44, $1, 'trialing', 3),
		(43, $1, 'cancelled', 99),
		(42, $1, 'cancelled', 77)`, planID); err != nil {
		t.Fatalf("seed subscription: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE platform.subscriptions
		   SET billing_cycle = 'yearly', trial_ends_at = $1, grace_until = $2
		 WHERE company_id = 44`, trialEnds, graceUntil); err != nil {
		t.Fatalf("seed subscription 44: %v", err)
	}

	live := time.Now().Add(24 * time.Hour)
	dead := time.Now().Add(-24 * time.Hour)
	// 三筆 override,其中一筆是「同一個功能被撤銷的舊承諾」:少了 revoked_at IS NULL 的過濾
	// 就會多回一筆(限額 999),而只看筆數的測試才驗得出這件事。
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.tenant_overrides
			(company_id, feature_code, enabled, limit_value, reason, owner, expires_at, revoked_at, created_by)
		VALUES (42,'limit.seats',NULL,50,'簽約承諾','sales@example.com',$1,NULL,1),
		       (42,'feature.printing',true,NULL,'短期試用','sales@example.com',$2,NULL,1),
		       (42,'limit.seats',NULL,999,'作廢的舊承諾','sales@example.com',$1,now(),1)`,
		live, dead); err != nil {
		t.Fatalf("seed overrides: %v", err)
	}

	st := postgres.New(db)

	feats, err := st.Features(ctx)
	if err != nil {
		t.Fatalf("Features: %v", err)
	}
	if len(feats) != 2 {
		t.Fatalf("Features 應回 2 筆,got %d: %+v", len(feats), feats)
	}
	// 逐欄比對非空值:欄位順序寫錯(例如 unit/description 互換)在這裡就會露出來。
	if got := feats["limit.seats"]; got.Type != "integer" || got.Unit != "席" || got.Description != "席位上線" {
		t.Fatalf("limit.seats 的功能定義錯位,got %+v", got)
	}
	if got := feats["feature.printing"]; got.Type != "boolean" || got.Unit != "" || got.Description != "列印" {
		t.Fatalf("feature.printing 的功能定義錯位,got %+v", got)
	}

	ents, err := st.PlanEntitlements(ctx, "std")
	if err != nil {
		t.Fatalf("PlanEntitlements: %v", err)
	}
	byCode := map[string]store.Entitlement{}
	for _, e := range ents {
		byCode[e.FeatureCode] = e
	}
	if len(ents) != 2 {
		t.Fatalf("PlanEntitlements 應回 2 筆,got %d: %+v", len(ents), ents)
	}
	if e := byCode["limit.seats"]; !e.Enabled || e.Limit == nil || *e.Limit != 10 {
		t.Fatalf("limit.seats 的限額應為 10,got %+v", e)
	}
	if e, ok := byCode["feature.printing"]; !ok || !e.Enabled || e.Limit != nil {
		t.Fatalf("NULL 限額(不限)必須掃成 nil 而非 0,got %+v(存在=%v)", e, ok)
	}
	if unknown, err := st.PlanEntitlements(ctx, "no_such_plan"); err != nil || len(unknown) != 0 {
		t.Fatalf("未知方案應回空而非錯誤,got %+v err=%v", unknown, err)
	}

	// 42 另有一列已取消的舊訂閱(seat_count 77):未取消的過濾必須挑出 active 那一列。
	sub, err := st.Subscription(ctx, 42)
	if err != nil || sub == nil {
		t.Fatalf("Subscription(42): got %+v err=%v", sub, err)
	}
	if sub.CompanyID != 42 || sub.PlanCode != "std" || sub.PlanID != planID ||
		sub.Status != "active" || sub.SeatCount != 10 || sub.BillingCycle != "monthly" {
		t.Fatalf("42 的訂閱欄位不對(同一公司另有已取消的列,不得回錯那一列),got %+v", *sub)
	}
	if sub.TrialEnds != nil || sub.GraceUntil != nil {
		t.Fatalf("42 的試用／寬限為 NULL,應掃成 nil,got %+v", *sub)
	}
	if sub, err := st.Subscription(ctx, 44); err != nil || sub == nil {
		t.Fatalf("Subscription(44): got %+v err=%v", sub, err)
	} else if sub.BillingCycle != "yearly" || sub.Status != "trialing" || sub.SeatCount != 3 {
		t.Fatalf("年繳方案必須帶出 billing_cycle=yearly(期別要 +1 年,G1),got %+v", *sub)
	} else if sub.TrialEnds == nil || !sub.TrialEnds.Equal(trialEnds) ||
		sub.GraceUntil == nil || !sub.GraceUntil.Equal(graceUntil) {
		t.Fatalf("試用到期／寬限日未帶出,got %+v", *sub)
	}
	// 43 的唯一一列是已取消:拿掉 `status <> 'cancelled'` 必定回那一列,故此斷言是該條件的守門
	// (把已取消的舊約當成現行訂閱,會讓停用／退款的公司繼續享有權益)。
	if sub, err := st.Subscription(ctx, 43); err != nil || sub != nil {
		t.Fatalf("只有已取消訂閱的租戶應回 (nil, nil),got %+v err=%v", sub, err)
	}
	// 45 完全沒有列:走的是 ErrNoRows → (nil, nil) 的路徑,與上面那條不同。
	if sub, err := st.Subscription(ctx, 45); err != nil || sub != nil {
		t.Fatalf("無訂閱應回 (nil, nil),got %+v err=%v", sub, err)
	}

	ov, err := st.Overrides(ctx, 42)
	if err != nil {
		t.Fatalf("Overrides: %v", err)
	}
	if len(ov) != 2 {
		t.Fatalf("Overrides 應回未撤銷的兩筆(撤銷者剔除、到期者保留),got %d: %+v", len(ov), ov)
	}
	byFeature := map[string]store.Override{}
	for _, o := range ov {
		if o.CompanyID != 42 {
			t.Fatalf("override 未帶出 company_id,got %+v", o)
		}
		byFeature[o.FeatureCode] = o
	}
	if o := byFeature["limit.seats"]; o.Enabled != nil || o.Limit == nil || *o.Limit != 50 || o.ExpiresAt == nil {
		t.Fatalf("只覆寫限額的例外:enabled 必須為 nil、限額 50、未到期,got %+v", o)
	}
	// 已到期者仍要回傳:判定層才看得到它並據以判定失效。
	if o := byFeature["feature.printing"]; o.Enabled == nil || !*o.Enabled || o.Limit != nil ||
		o.ExpiresAt == nil || !o.ExpiresAt.Before(time.Now()) {
		t.Fatalf("已到期的例外仍應回傳且 enabled=true、limit=nil、到期日在過去,got %+v", o)
	}
	if other, err := st.Overrides(ctx, 43); err != nil || len(other) != 0 {
		t.Fatalf("無例外的租戶應回空,got %+v err=%v", other, err)
	}
}
