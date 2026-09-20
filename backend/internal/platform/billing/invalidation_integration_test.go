//go:build integration

// Task 8 的整合驗收（真 PostgreSQL ＋ 真 Valkey）：**訂閱狀態改變後，下一個判定不得用到舊快取**。
//
// 為什麼非真容器不可：這條的失效跨兩層 —— billing 在交易提交後刪 Valkey 上的鍵，而權益判定
// 從 PostgreSQL 回源重建。假 store／假快取驗不到「同一顆 Valkey 上真的少了一個鍵」這件事，
// 而那正是三個寫入來源（API／排程／consumer）分屬不同行程時唯一的溝通管道。
//
// 兩條路徑各驗一次：
//
//	① 排程（SuspendOverdue）：past_due ＋ 寬限已過 → suspended。判定必須由「可用」變「不可用」。
//	② 平台寫入（RecordPayment）：trialing → active。判定必須由「試用期內不因方案 disabled 而擋」
//	   變回「方案說 disabled 就是 disabled」（見 entitlements.resolveFeature 的 trialing 分支）。
//
// 前置斷言（快取真的被寫入）不可省：少了它，「刪掉 Delete 呼叫」的版本仍會全綠 ——
// 沒有快取就沒有「用到舊快取」可言。
//
// 執行：task test:integration -- -count=1 -run TestIntegrationSubscriptionStateChangeInvalidatesCache -v
package billing_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

// zeroCounter 為計數器替身：本測試只判定 boolean 功能，用量不會被用到。
type zeroCounter struct{}

func (zeroCounter) Count(context.Context, int, string) (int, error) { return 0, nil }

const (
	suspendCompany = 42 // ① 排程停用
	payCompany     = 43 // ② 收款復原
)

func TestIntegrationSubscriptionStateChangeInvalidatesCache(t *testing.T) {
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
	if err := goose.RunContext(t.Context(), "up", db, "../../../database/migrations"); err != nil {
		t.Fatalf("goose up: %v", err)
	}
	valkeyAddr := testsupport.Valkey(t)
	client := cache.NewClient(valkeyAddr)
	defer func() { _ = client.Close() }()
	ctx := t.Context()

	// 種子：兩個方案（std 含列印、trial 不含）＋ 兩個租戶 ＋ 一筆 open 期別 ＋ 一個 operator。
	var stdPlan, trialPlan, opID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&stdPlan); err != nil {
		t.Fatalf("plan std: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('trial','試用') RETURNING id`).Scan(&trialPlan); err != nil {
		t.Fatalf("plan trial: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.features (code, type, unit, description)
		VALUES ('feature.printing','boolean','','列印')`); err != nil {
		t.Fatalf("feature: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value) VALUES
		($1,'feature.printing',true,NULL),
		($2,'feature.printing',false,NULL)`, stdPlan, trialPlan); err != nil {
		t.Fatalf("plan_entitlements: %v", err)
	}
	// ① past_due 且寬限已過 → SuspendOverdue 的掃描對象。
	expiredGrace := time.Now().Add(-time.Hour)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle, grace_until)
		VALUES ($1, $2, 'past_due', 3, 'monthly', $3)`, suspendCompany, stdPlan, expiredGrace); err != nil {
		t.Fatalf("subscription %d: %v", suspendCompany, err)
	}
	// ② trialing＋未付期別 → RecordPayment 的對象（收款把訂閱帶回 active）。
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle, trial_ends_at)
		VALUES ($1, $2, 'trialing', 3, 'monthly', now() + interval '10 days')`,
		payCompany, trialPlan); err != nil {
		t.Fatalf("subscription %d: %v", payCompany, err)
	}
	var subID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM platform.subscriptions WHERE company_id = $1`, payCompany).Scan(&subID); err != nil {
		t.Fatalf("查訂閱 %d: %v", payCompany, err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, currency, status)
		VALUES ($1, 1, now(), now() + interval '30 days', $2, 1500.00, 150.00, 3, 1950.00, 'TWD', 'open')`,
		subID, trialPlan); err != nil {
		t.Fatalf("period: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.operators (email, name) VALUES ('ops@example.com','Ops') RETURNING id`).
		Scan(&opID); err != nil {
		t.Fatalf("operator: %v", err)
	}

	st := postgres.New(db)
	vcache := entitlements.NewValkeyCache(client)
	// TTL 刻意給長（1 小時）：這條測的是「顯式失效」，若靠 TTL 收斂，測試會因為還沒到期而失敗。
	svc := entitlements.New(st, zeroCounter{}, vcache, time.Hour)
	bcache := entitlements.NewValkeyCache(client) // 另一個實例＝另一個行程的視角（同一顆 Valkey）
	b := billing.NewBilling(st).WithCache(bcache)

	// ① 排程停用：past_due（可用）→ suspended（不可用）
	allowed, err := svc.Allows(ctx, suspendCompany, entitlements.FeaturePrinting)
	if err != nil || !allowed {
		t.Fatalf("前置：寬限內的租戶應可用，got allowed=%v err=%v", allowed, err)
	}
	if _, ok, err := vcache.Get(ctx, "ent:42"); err != nil || !ok {
		t.Fatalf("前置：判定後應在 Valkey 留下快取（否則本測試測不到舊快取），got ok=%v err=%v", ok, err)
	}

	if n, err := b.SuspendOverdue(ctx, time.Now()); err != nil || n != 1 {
		t.Fatalf("應停用一家: n=%d err=%v", n, err)
	}

	// 快取必須**已經**消失（拿掉 Delete 呼叫，這一條與下一條都會紅）。
	if _, ok, err := vcache.Get(ctx, "ent:42"); err != nil || ok {
		t.Fatalf("停用後快取必須被刪除，got ok=%v err=%v", ok, err)
	}
	// 下一個判定不得用到舊快取：DB 已是 suspended → 不可用。
	allowed, err = svc.Allows(ctx, suspendCompany, entitlements.FeaturePrinting)
	if err != nil {
		t.Fatalf("Allows: %v", err)
	}
	if allowed {
		t.Fatal("訂閱已停用，判定仍放行（＝用了失效前的舊快取）")
	}
	var status string
	if err := db.QueryRowContext(ctx,
		`SELECT status FROM platform.subscriptions WHERE company_id = $1`, suspendCompany).
		Scan(&status); err != nil || status != "suspended" {
		t.Fatalf("DB 狀態應為 suspended（證明放行與否真的來自狀態），got %q err=%v", status, err)
	}

	// ② 收款復原：trialing（試用期內不因方案 disabled 而擋）→ active（方案說 disabled 就是 disabled）
	tallowed, err := svc.Allows(ctx, payCompany, entitlements.FeaturePrinting)
	if err != nil || !tallowed {
		t.Fatalf("前置：試用期內應可用（方案未含但試用不擋），got allowed=%v err=%v", tallowed, err)
	}
	if _, ok, err := vcache.Get(ctx, "ent:43"); err != nil || !ok {
		t.Fatalf("前置：判定後應在 Valkey 留下快取，got ok=%v err=%v", ok, err)
	}

	if _, err := b.RecordPayment(ctx, billing.RecordPaymentInput{
		CompanyID: payCompany, Provider: "manual", ExternalRef: "BANK-88888",
		ActorOperatorID: opID, Reason: "匯款入帳（試用轉正式）",
	}); err != nil {
		t.Fatalf("收款: %v", err)
	}

	if _, ok, err := vcache.Get(ctx, "ent:43"); err != nil || ok {
		t.Fatalf("收款改變訂閱狀態後快取必須被刪除，got ok=%v err=%v", ok, err)
	}
	tallowed, err = svc.Allows(ctx, payCompany, entitlements.FeaturePrinting)
	if err != nil {
		t.Fatalf("Allows: %v", err)
	}
	if tallowed {
		t.Fatal("訂閱已轉 active 而方案未含列印，判定仍放行（＝用了失效前的舊快取）")
	}
}
