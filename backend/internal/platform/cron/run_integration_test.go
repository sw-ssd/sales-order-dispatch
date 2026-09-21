//go:build integration

// Task 7 的整合驗收(真 PostgreSQL ＋ 真容器):單元測試的假 deps 驗不到三件事 ——
// (1) PostgreSQL advisory lock 真的擋得住第二條連線的單飛;(2) 四個掃描串起來的一趟,
// 真的把資料從 active 推到 past_due 再推到 suspended;(3) 凍結真的經 consumer 落到產品域
// (companies.status ＋ 租戶稽核)。
//
// 端到端一趟的租戶(now = 2026-10-01 03:00 UTC):
//
//	A:active ＋ 第 1 期 open 且期末已過  → 第 1 趟 past_due(設寬限 7 天) → 第 2 趟 suspended
//	   ＋ consumer 在同一趟內把公司轉 suspended(這就是 G7／spec §5.4 的「逾期凍結」)
//	B:active ＋ 第 1 期 open 且期末在提前窗內 → 第 1 趟開出第 2 期(period.opened 也被認領)
//
// 執行:`task test:integration -- -count=1 -run TestIntegrationRunOnceOverdueFreesTenant -v`
package cron_test

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/cron"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// cronMigrationsDir 相對套件目錄(go test 以套件目錄為 cwd),與 cmd/migrate 同路徑。
const cronMigrationsDir = "../../../database/migrations"

func TestIntegrationRunOnceOverdueFreesTenant(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	dsn := testsupport.Postgres(t)
	adminDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = adminDB.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.RunContext(ctx, "up", adminDB, cronMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	now := at(2026, time.October, 1, 3)
	planID := seedCronPlan(t, ctx, adminDB)
	// 兩家租戶都是「系統排程」的排程對象:公司與其系統 actor 使用者(FORCE RLS → 必須明確
	// 系統範圍交易,與生產的 consumer 同一條路)。
	companyA, actorID := seedCronTenant(t, ctx, adminDB, "T7-OVERDUE")
	companyB, _ := seedCronTenant(t, ctx, adminDB, "T7-LEADWINDOW")
	// A:期末已過(第 1 趟逾期);B:期末在提前窗內(第 1 趟開下一期)。
	subA := seedCronSubscription(t, ctx, adminDB, planID, companyA, "active")
	seedCronPeriod(t, ctx, adminDB, planID, subA, now.Add(-time.Hour))
	subB := seedCronSubscription(t, ctx, adminDB, planID, companyB, "active")
	seedCronPeriod(t, ctx, adminDB, planID, subB, now.AddDate(0, 0, 10))
	// C:trialing 但沒有到期日（未結項 #40）—— ExpireTrials 刻意不碰，摘要必須計 1。
	// 用直接 INSERT（開通路徑要求 trial_ends_at，造不出這種列；這正是它只剩手工列的原因）。
	companyC, _ := seedCronTenant(t, ctx, adminDB, "T7-STUCKTRIAL")
	if _, err := adminDB.ExecContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle)
		VALUES ($1, $2, 'trialing', 2, 'monthly')`, companyC, planID); err != nil {
		t.Fatalf("stuck trial 訂閱: %v", err)
	}

	// 營運參數與系統 actor 一律走 platform.settings(seed 在正式環境負責;這裡逐鍵寫入,
	// 故同時驗到 LoadParams 的鍵名與真設定表一致)。
	for key, value := range map[string]string{
		"system_actor_user_id": strconv.Itoa(actorID),
		"grace_days":           "7",
		"lead_days":            "14",
	} {
		if _, err := adminDB.ExecContext(ctx, `
			INSERT INTO platform.settings (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, key, value); err != nil {
			t.Fatalf("寫 platform.settings(%s): %v", key, err)
		}
	}

	// 生產組裝:admin 連線的 store(平台表)＋ 真帳務 ＋ 真 consumer(系統範圍交易 ＋ 產品域唯一入口)
	// ＋ advisory lock 單飛鎖。cron 不碰業務連線,故沒有 dbtenant 連線池上限的問題。
	st := postgres.New(adminDB)
	deps := cron.Deps{
		Billing:  billing.NewBilling(st),
		Consumer: consumer.New(st, consumer.NewDBSystemTx(adminDB), consumer.ProductDomain{}),
		Store:    st,
		Lock:     cron.NewAdvisoryLocker(adminDB, cron.LockKey),
	}
	params, err := cron.LoadParams(ctx, st)
	if err != nil {
		t.Fatalf("讀排程參數: %v", err)
	}
	if params.GraceDays != 7 || params.LeadDays != 14 || params.EventBatch != cron.EventBatch {
		t.Fatalf("參數應由 platform.settings 來: %+v", params)
	}

	// ① 單飛鎖:另一條連線持有時,整趟跳過(不是錯誤),且**什麼都沒動**。
	held := cron.NewAdvisoryLocker(adminDB, cron.LockKey)
	release, ok, err := held.TryLock(ctx)
	if err != nil || !ok {
		t.Fatalf("取鎖: ok=%v err=%v", ok, err)
	}
	if _, ok2, err := cron.NewAdvisoryLocker(adminDB, cron.LockKey).TryLock(ctx); err != nil || ok2 {
		t.Fatalf("鎖被持有時另一條連線不得取得: ok=%v err=%v", ok2, err)
	}
	skipped, err := cron.RunGuarded(ctx, deps, now, params)
	if err != nil {
		t.Fatalf("取不到鎖不是錯誤: %v", err)
	}
	if skipped != (cron.Summary{}) {
		t.Fatalf("第二個執行不得處理任何事: %+v", skipped)
	}
	if status := cronCompanyStatus(t, ctx, adminDB, companyA); status != "active" {
		t.Fatalf("被擋下的一趟不得改變任何狀態,got %q", status)
	}
	var eventsWhileHeld int
	if err := adminDB.QueryRowContext(ctx, `SELECT count(*) FROM platform.events`).Scan(&eventsWhileHeld); err != nil {
		t.Fatalf("查事件數: %v", err)
	}
	if eventsWhileHeld != 0 {
		t.Fatalf("被擋下的一趟不得產生任何事件,got %d", eventsWhileHeld)
	}
	if err := release(ctx); err != nil {
		t.Fatalf("釋放鎖: %v", err)
	}
	// 釋放後要拿得到(鎖沒有殘留 —— 這是 advisory lock 相對「執行中旗標」的價值)。
	free, ok3, err := cron.NewAdvisoryLocker(adminDB, cron.LockKey).TryLock(ctx)
	if err != nil || !ok3 {
		t.Fatalf("釋放後應取得得到鎖: ok=%v err=%v", ok3, err)
	}
	if err := free(ctx); err != nil {
		t.Fatalf("釋放鎖: %v", err)
	}

	// ①-2 解鎖失敗時必須**丟棄**那條連線(不得還池):連線還活著就可能仍持有 session 級鎖,
	// 還回池裡會讓同一行程的下一趟擋住自己(外面看起來就是排程靜默停擺)。
	//
	// 怎麼在不靠競態的情況下造出「解鎖失敗但鎖還在」:以業務角色 app_rw 取鎖(它的 EXECUTE 可以收回),
	// 再從 owner 連線 REVOKE EXECUTE ON pg_advisory_unlock → 解鎖以 42501 失敗,而 session 級鎖仍在。
	// 探針走**另一條 session**(adminDB,superuser):鎖真的放掉了(=連線被丟棄)才算通過。
	lockDB, err := sql.Open("pgx", testsupport.AppRoleDSN(t, dsn))
	if err != nil {
		t.Fatalf("連線(鎖專用池): %v", err)
	}
	lockDB.SetMaxOpenConns(1)
	defer func() { _ = lockDB.Close() }()
	doomed := cron.NewAdvisoryLocker(lockDB, cron.LockKey)
	releaseDoomed, okDoomed, err := doomed.TryLock(ctx)
	if err != nil || !okDoomed {
		t.Fatalf("app_rw 取鎖: ok=%v err=%v", okDoomed, err)
	}
	if _, err := adminDB.ExecContext(ctx,
		`REVOKE EXECUTE ON FUNCTION pg_advisory_unlock(bigint) FROM PUBLIC`); err != nil {
		t.Fatalf("收回 pg_advisory_unlock 的執行權: %v", err)
	}
	if err := releaseDoomed(ctx); err == nil {
		t.Fatal("解鎖失敗必須回報錯誤,不得靜默宣稱已釋放")
	}
	if _, err := adminDB.ExecContext(ctx,
		`GRANT EXECUTE ON FUNCTION pg_advisory_unlock(bigint) TO PUBLIC`); err != nil {
		t.Fatalf("還原 pg_advisory_unlock 的執行權: %v", err)
	}
	if !waitLockFree(t, ctx, adminDB) {
		t.Fatal("解鎖失敗後那條連線必須被丟棄(鎖已真的放掉),否則同一行程的下一趟會擋住自己")
	}

	// ② 第 1 趟(now):A 逾期(設寬限 7 天)、B 開出第 2 期。
	first, err := cron.RunGuarded(ctx, deps, now, params)
	if err != nil {
		t.Fatalf("第 1 趟: %v", err)
	}
	if !first.Locked || first.PastDue != 1 || first.Suspended != 0 ||
		first.ExpiredCancelled != 0 || first.PeriodsOpened != 1 {
		t.Fatalf("第 1 趟計數不符: %+v", first)
	}
	// 本趟認領 2 筆:subscription.past_due(A)與 period.opened(B);派送放在產生期別之後,
	// 故 B 這一趟新開的期別事件也在同一趟被認領。
	if first.Dispatched != 2 {
		t.Fatalf("第 1 趟應認領 2 筆事件,got %d", first.Dispatched)
	}
	if first.Receivables != 1 {
		t.Fatalf("待收款應只有 A 的過期未付期別,got %d", first.Receivables)
	}
	if first.StuckTrialing != 1 {
		t.Fatalf("無到期日的試用(C)應被計入摘要，got %d", first.StuckTrialing)
	}
	if status := cronCompanyStatus(t, ctx, adminDB, companyA); status != "active" {
		t.Fatalf("逾期(past_due)仍在寬限期內,不得凍結,got %q", status)
	}
	var graceUntil time.Time
	if err := adminDB.QueryRowContext(ctx,
		`SELECT grace_until FROM platform.subscriptions WHERE id = $1`, subA).Scan(&graceUntil); err != nil {
		t.Fatalf("查寬限期: %v", err)
	}
	if want := now.AddDate(0, 0, 7); !graceUntil.Equal(want) {
		t.Fatalf("寬限期應為 %s(gnow ＋ grace_days),got %s", want.Format(time.RFC3339), graceUntil.Format(time.RFC3339))
	}

	// ③ 第 2 趟(寬限已過):A 停用,且 consumer 在同一趟內把公司轉 suspended ＋ 落租戶稽核。
	afterGrace := now.AddDate(0, 0, 7).Add(time.Hour)
	second, err := cron.RunGuarded(ctx, deps, afterGrace, params)
	if err != nil {
		t.Fatalf("第 2 趟: %v", err)
	}
	if !second.Locked || second.Suspended != 1 || second.PastDue != 0 ||
		second.ExpiredCancelled != 0 || second.PeriodsOpened != 0 || second.Dispatched != 1 {
		t.Fatalf("第 2 趟計數不符: %+v", second)
	}
	if status := cronCompanyStatus(t, ctx, adminDB, companyA); status != string(company.StatusSuspended) {
		t.Fatalf("寬限已過應凍結公司(G7),got %q", status)
	}
	var subStatus string
	var clearedGrace sql.NullTime
	if err := adminDB.QueryRowContext(ctx,
		`SELECT status, grace_until FROM platform.subscriptions WHERE id = $1`, subA).
		Scan(&subStatus, &clearedGrace); err != nil {
		t.Fatalf("查訂閱: %v", err)
	}
	if subStatus != "suspended" || clearedGrace.Valid {
		t.Fatalf("訂閱應為 suspended 且清空寬限期: status=%q grace=%v", subStatus, clearedGrace.Time)
	}
	// 凍結的稽核由 consumer 經 SetCompanyStatus 落租戶稽核(不是排程自己寫平台稽核)。
	var audits, auditUser int
	var auditStatus, auditReason string
	if err := adminDB.QueryRowContext(ctx, `
		SELECT count(*) OVER (), user_id, after_snapshot->>'status', after_snapshot->>'reason'
		  FROM audit_logs WHERE company_id = $1 AND resource_type = 'company'`, companyA).
		Scan(&audits, &auditUser, &auditStatus, &auditReason); err != nil {
		t.Fatalf("查租戶稽核: %v", err)
	}
	if audits != 1 || auditUser != actorID {
		t.Fatalf("凍結應留 1 筆租戶稽核且 actor 為系統 actor(%d): audits=%d user=%d", actorID, audits, auditUser)
	}
	if auditStatus != string(company.StatusSuspended) || auditReason == "" {
		t.Fatalf("稽核應帶新狀態與原因: status=%q reason=%q", auditStatus, auditReason)
	}
	var suspendedEventID int64
	if err := adminDB.QueryRowContext(ctx,
		`SELECT id FROM platform.events WHERE event_type = 'subscription.suspended'`).Scan(&suspendedEventID); err != nil {
		t.Fatalf("查停用事件: %v", err)
	}

	// ④ 第 3 趟(同一時間):完全可重跑 —— 不再轉移、不再派送、不留第二筆稽核。
	third, err := cron.RunGuarded(ctx, deps, afterGrace, params)
	if err != nil {
		t.Fatalf("第 3 趟: %v", err)
	}
	if third.PastDue != 0 || third.Suspended != 0 || third.ExpiredCancelled != 0 ||
		third.PeriodsOpened != 0 || third.Dispatched != 0 {
		t.Fatalf("重跑不得產生第二個副作用: %+v", third)
	}
	if third.Receivables != 1 {
		t.Fatalf("待收款是狀態不是動作,重跑不得歸零: got %d", third.Receivables)
	}
	var attempts, reAudits int
	if err := adminDB.QueryRowContext(ctx,
		`SELECT attempts FROM platform.events WHERE id = $1`, suspendedEventID).Scan(&attempts); err != nil {
		t.Fatalf("查事件 attempts: %v", err)
	}
	if err := adminDB.QueryRowContext(ctx,
		`SELECT count(*) FROM audit_logs WHERE company_id = $1`, companyA).Scan(&reAudits); err != nil {
		t.Fatalf("重跑後查稽核: %v", err)
	}
	if attempts != 1 || reAudits != audits {
		t.Fatalf("重跑不得產生第二個副作用: attempts=%d(應 1) audit=%d(原 %d)", attempts, reAudits, audits)
	}
	// 排程本身不寫 platform.audit_logs(schema 上不可滿足:operator_id NOT NULL ＋ FK platform.operators,
	// 而排程的主體是租戶 users.id)—— 排程的紀錄是事件流＋上面那一筆租戶稽核。
	var platformAudits int
	if err := adminDB.QueryRowContext(ctx, `SELECT count(*) FROM platform.audit_logs`).Scan(&platformAudits); err != nil {
		t.Fatalf("查平台稽核: %v", err)
	}
	if platformAudits != 0 {
		t.Fatalf("排程不得寫平台稽核,got %d 筆", platformAudits)
	}
}

// waitLockFree 回報單飛鎖是否已沒人持有:反覆嘗試一小段時間後才判定失敗 —— PostgreSQL 對「連線
// 斷開」的處置(連帶釋放該 session 的 advisory lock)是**非同步**的,連線剛關就問會問到還在的鎖。
func waitLockFree(t *testing.T, ctx context.Context, db *sql.DB) bool {
	t.Helper()
	for range 40 {
		release, ok, err := cron.NewAdvisoryLocker(db, cron.LockKey).TryLock(ctx)
		if err != nil {
			t.Fatalf("探針取鎖: %v", err)
		}
		if ok {
			if err := release(ctx); err != nil {
				t.Fatalf("釋放探針的鎖: %v", err)
			}
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// seedCronPlan 建立一個含月繳價目的方案(產生下一期需要當期生效價,沒有價目會大聲失敗)。
func seedCronPlan(t *testing.T, ctx context.Context, db *sql.DB) int64 {
	t.Helper()
	var planID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.plan_prices (plan_id, billing_cycle, base_price, seat_price, currency)
		VALUES ($1, 'monthly', 1500.00, 150.00, 'TWD')`, planID); err != nil {
		t.Fatalf("plan_prices: %v", err)
	}
	return planID
}

// seedCronTenant 建立一名 active 租戶與其系統 actor 使用者(users.id 是 audit_logs.user_id 的 FK),
// 回傳 (companyID, actorID)。companies／users 是 FORCE RLS 的業務表,故走明確的系統範圍交易
// —— 與生產的 consumer 同一條路(少了這一層會以 42501 失敗)。
func seedCronTenant(t *testing.T, ctx context.Context, db *sql.DB, identifier string) (companyID, actorID int) {
	t.Helper()
	admin := dbtenant.NewClient(db)
	err := dbtenant.SystemScopeTx(ctx, admin, func(tx *ent.Tx) error {
		co, err := tx.Client().Company.Create().
			SetName(identifier).SetIdentifier(identifier).SetStatus(company.StatusActive).Save(ctx)
		if err != nil {
			return err
		}
		companyID = co.ID
		u, err := tx.Client().User.Create().
			SetEmail(identifier + "@example.invalid").
			SetName("系統排程").
			SetRole("super").
			SetStatus("active").
			SetPasswordHash("!").
			SetCompanyID(companyID).
			Save(ctx)
		if err != nil {
			return err
		}
		actorID = u.ID
		return nil
	})
	if err != nil {
		t.Fatalf("建立租戶 %s: %v", identifier, err)
	}
	return companyID, actorID
}

// seedCronSubscription 建立一列訂閱(月繳,避免期別長度把月界算錯),回傳訂閱 id。
func seedCronSubscription(t *testing.T, ctx context.Context, db *sql.DB, planID int64,
	companyID int, status string) int64 {
	t.Helper()
	var subID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle)
		VALUES ($1, $2, $3, 2, 'monthly') RETURNING id`, companyID, planID, status).Scan(&subID); err != nil {
		t.Fatalf("subscription(company %d): %v", companyID, err)
	}
	return subID
}

// seedCronPeriod 建立第 1 期(open),期末為 end(起日往前一個月 —— 月底錨點由 Task 5 處理)。
func seedCronPeriod(t *testing.T, ctx context.Context, db *sql.DB, planID, subID int64, end time.Time) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, currency, status)
		VALUES ($1, 1, $2, $3, $4, 1500.00, 150.00, 2, 1800.00, 'TWD', 'open')`,
		subID, end.AddDate(0, -1, 0), end, planID); err != nil {
		t.Fatalf("period(sub %d): %v", subID, err)
	}
}

// cronCompanyStatus 讀公司目前的狀態(每次都重查,不靠上一段的區域變數)。
func cronCompanyStatus(t *testing.T, ctx context.Context, db *sql.DB, companyID int) string {
	t.Helper()
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM companies WHERE id = $1`, companyID).Scan(&status); err != nil {
		t.Fatalf("查公司狀態: %v", err)
	}
	return status
}
