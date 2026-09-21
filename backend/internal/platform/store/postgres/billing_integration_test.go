//go:build integration

// 平台帳務寫入路徑（00030／billing.go）的整合驗收。四個可觀測契約：
//
//	① 00030 的遷移可完整回滾，且 settings 的欄位形狀讓 cmd/seed 的
//	   `INSERT INTO platform.settings (key, value) VALUES (…)` 直接成功（seed 不帶 updated_at，
//	   並以 key 當 ON CONFLICT 的衝突目標）；
//	② 所有寫入共用同一個 *sql.Tx：期別／事件／稽核在同一交易內寫入後可讀，回滾後三者皆不存在
//	   （不得留下孤兒事件或半筆期別），且 OpenSubscriptionTx 的 FOR UPDATE 真的鎖住列；
//	③ 取列語意與平台投影同源：**不得**預先濾掉 cancelled（F-8／C-02），只有全是 cancelled 時才取它；
//	④ 期別冪等鍵與排程三個查詢（逾期未付／寬限到期／已取消期末）的集合邊界。
//
// 這些在記憶體假實作上永遠測不出來：表不存在、欄位錯位、unique 索引不存在、FOR UPDATE 沒鎖到列、
// cancelled 被過濾 —— 全都只有真 PostgreSQL 說得出來。
package postgres_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationBillingMigrationSettings 驗證 00030 的兩件事：回滾完整（只帶走 settings）、
// 以及 settings 的欄位形狀（key 為主鍵、value 為文字、updated_at 有預設值）真的能承接
// cmd/seed/platform.go 的兩句寫入 —— 少了 updated_at 的預設值，seed 的第一步就 23502 失敗，
// 而排程讀不到 system_actor_user_id 就 Fatal（G5 的原始缺陷原樣回來）。
func TestIntegrationBillingMigrationSettings(t *testing.T) {
	testsupport.RequiresContainer(t)
	db, err := sql.Open("pgx", testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.RunContext(t.Context(), "up", db, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
	ctx := t.Context()

	// settings 的形狀：key 是唯一衝突目標（seed 用 ON CONFLICT (key)）、value 為非空文字、
	// updated_at 有預設值（seed 不帶它）。
	var dataType, isNullable string
	if err := db.QueryRowContext(ctx, `
		SELECT data_type, is_nullable FROM information_schema.columns
		 WHERE table_schema = 'platform' AND table_name = 'settings' AND column_name = 'value'`,
	).Scan(&dataType, &isNullable); err != nil {
		t.Fatalf("查 settings.value 的型別: %v", err)
	}
	if dataType != "text" || isNullable != "NO" {
		t.Fatalf("settings.value 應為 text NOT NULL（seeder 寫的是文字），got %s nullable=%s", dataType, isNullable)
	}
	var updatedDefault sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT column_default FROM information_schema.columns
		 WHERE table_schema = 'platform' AND table_name = 'settings' AND column_name = 'updated_at'`,
	).Scan(&updatedDefault); err != nil {
		t.Fatalf("查 settings.updated_at: %v", err)
	}
	if !updatedDefault.Valid || updatedDefault.String == "" {
		t.Fatal("settings.updated_at 必須有預設值：seed 的 INSERT 不帶它，沒有預設值即 23502 失敗")
	}

	// 逐字重現 cmd/seed/platform.go 的兩句寫入（不帶 updated_at、以 key 為衝突目標）。
	seedActor := func(v string) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO platform.settings (key, value) VALUES ('system_actor_user_id', $1)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
			 WHERE platform.settings.value IS DISTINCT FROM EXCLUDED.value`, v)
		return err
	}
	if err := seedActor("7"); err != nil {
		t.Fatalf("seed 的 system_actor_user_id 寫入必須成功（否則 seed 中止）: %v", err)
	}
	var firstUpdated time.Time
	if err := db.QueryRowContext(ctx,
		`SELECT updated_at FROM platform.settings WHERE key = 'system_actor_user_id'`).Scan(&firstUpdated); err != nil {
		t.Fatalf("讀回 updated_at: %v", err)
	}
	// 重跑同值：seed 的 `WHERE value IS DISTINCT FROM EXCLUDED.value` 應讓整句不寫入
	//（updated_at 是 console 的「最後修改時間」，被每次重跑推進會誤導營運）。
	if err := seedActor("7"); err != nil {
		t.Fatalf("seed 重跑同值: %v", err)
	}
	var again time.Time
	if err := db.QueryRowContext(ctx,
		`SELECT updated_at FROM platform.settings WHERE key = 'system_actor_user_id'`).Scan(&again); err != nil {
		t.Fatalf("重跑後讀回 updated_at: %v", err)
	}
	if !again.Equal(firstUpdated) {
		t.Fatalf("同值重跑不得推進 updated_at，got %v → %v", firstUpdated, again)
	}
	// 值真的不同時要寫入（自癒：值被改壞時重跑 seed 修正）。
	time.Sleep(2 * time.Millisecond)
	if err := seedActor("9"); err != nil {
		t.Fatalf("seed 改值: %v", err)
	}
	var changed time.Time
	if err := db.QueryRowContext(ctx,
		`SELECT updated_at FROM platform.settings WHERE key = 'system_actor_user_id'`).Scan(&changed); err != nil {
		t.Fatalf("改值後讀回 updated_at: %v", err)
	}
	if !changed.After(firstUpdated) {
		t.Fatalf("值不同時必須寫入並推進 updated_at，got %v → %v", firstUpdated, changed)
	}
	// 營運參數走「只補缺」：既有值不得被重跑蓋回去。
	for _, v := range []string{"14", "30"} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.settings (key, value) VALUES ('trial_days', $1)
			ON CONFLICT (key) DO NOTHING`, v); err != nil {
			t.Fatalf("seed 的 trial_days 寫入(%s): %v", v, err)
		}
	}
	var trialDays string
	if err := db.QueryRowContext(ctx,
		`SELECT value FROM platform.settings WHERE key = 'trial_days'`).Scan(&trialDays); err != nil {
		t.Fatalf("讀回 trial_days: %v", err)
	}
	if trialDays != "14" {
		t.Fatalf("重跑 seed 不得覆寫營運調整過的天數，got %q", trialDays)
	}

	// 回滾：00030 的 down 只帶走 settings，00029 的其他表必須留著；回滾後重新 up 要能再建回
	//（回滾鏈可重複：migrate down 不得只把版本列往回寫而把物件留在庫上）。
	// down-to 29（29 保留、30＋ 回滾）：goose 的 down-to 語意是「回滾到該版本」，
	// 不是「回滾該版本」—— 要執行 00030 的 Down 必須停在 29。00031＋ 的訂單／清單表
	// 在本測試的庫上游（up 跑全量目錄），故一併被帶走；本測試只斷言 00030/00029 的語意，
	// 不斷言訂單表去留（訂單表的 Down 由 sales_order_schema_test 覆蓋）。
	if err := goose.RunContext(ctx, "down-to", db, platformMigrationsDir, "29"); err != nil {
		t.Fatalf("goose down-to 29: %v", err)
	}
	var settings, subscriptions sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT to_regclass('platform.settings')::text, to_regclass('platform.subscriptions')::text`,
	).Scan(&settings, &subscriptions); err != nil {
		t.Fatalf("查回滾後的表: %v", err)
	}
	if settings.Valid {
		t.Fatal("00030 回滾後 platform.settings 必須消失（DROP TABLE，不得只寫版本列）")
	}
	if !subscriptions.Valid {
		t.Fatal("00030 回滾不得動到 00029 的表（platform.subscriptions 應留著）")
	}
	if err := goose.RunContext(ctx, "up", db, platformMigrationsDir); err != nil {
		t.Fatalf("回滾後重新 up: %v", err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM platform.settings`).Scan(&n); err != nil {
		t.Fatalf("重新 up 後查 settings: %v", err)
	}
	if n != 0 {
		t.Fatalf("重新 up 建出的是空表，got %d 列", n)
	}
}

// TestIntegrationPlatformBillingStoreTx 驗證寫入路徑的交易語意、期別冪等鍵與取列邊界。
func TestIntegrationPlatformBillingStoreTx(t *testing.T) {
	testsupport.RequiresContainer(t)
	db, err := sql.Open("pgx", testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.RunContext(t.Context(), "up", db, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// fixture：方案 ＋ 月繳價目 ＋ 訂閱 ＋ operator（稽核 FK）＋ settings 四鍵。
	var planID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.plan_prices (plan_id, billing_cycle, base_price, seat_price, currency)
		VALUES ($1,'monthly',1500.00,150.00,'TWD')`, planID); err != nil {
		t.Fatalf("plan_prices: %v", err)
	}
	trialEnds := now.Add(48 * time.Hour)
	var subID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, trial_ends_at)
		VALUES (42, $1, 'active', 3, $2) RETURNING id`, planID, trialEnds).Scan(&subID); err != nil {
		t.Fatalf("subscription: %v", err)
	}
	// 43 的唯一一列就是 cancelled（F-8 的形狀）；42 另有一列已取消的舊約（partial unique index
	// 只管未取消者，故同一公司可以並存）；44 完全沒有列。
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle) VALUES
		(43, $1, 'cancelled', 99, 'yearly'),
		(42, $1, 'cancelled', 77, 'monthly')`, planID); err != nil {
		t.Fatalf("subscription 43／42-cancelled: %v", err)
	}
	var opID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.operators (email, name) VALUES ('ops@example.com','Ops') RETURNING id`).
		Scan(&opID); err != nil {
		t.Fatalf("operator: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES
		('system_actor_user_id','7'), ('trial_days','14'), ('grace_days','7'), ('lead_days','14')`); err != nil {
		t.Fatalf("settings: %v", err)
	}

	st := postgres.New(db)

	// 設定與系統 actor（排程的稽核主體，G5）。
	if actor, err := st.SystemActor(ctx); err != nil || actor != 7 {
		t.Fatalf("SystemActor = %d, err=%v；want 7", actor, err)
	}
	if v, err := st.Setting(ctx, "lead_days"); err != nil || v != "14" {
		t.Fatalf("Setting(lead_days) = %q, err=%v；want 14", v, err)
	}
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.UpsertSettingTx(ctx, tx, "lead_days", "20")
	}); err != nil {
		t.Fatalf("UpsertSettingTx: %v", err)
	}
	if v, err := st.Setting(ctx, "lead_days"); err != nil || v != "20" {
		t.Fatalf("UpsertSettingTx 後 Setting(lead_days) = %q, err=%v；want 20", v, err)
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE platform.settings SET value = 'not-a-number' WHERE key = 'system_actor_user_id'`); err != nil {
		t.Fatalf("破壞 system_actor_user_id: %v", err)
	}
	if actor, err := st.SystemActor(ctx); err == nil {
		t.Fatalf("system_actor_user_id 不是數字時必須回錯誤（不得靜默當 0），got %d", actor)
	}

	// ① 同一交易內寫入期別／事件／稽核，回滾後三者皆不得留下。
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	sub, err := st.OpenSubscriptionTx(ctx, tx, 42)
	if err != nil || sub == nil {
		t.Fatalf("OpenSubscriptionTx(42): got %+v err=%v", sub, err)
	}
	// 同一公司另有已取消的舊約：必須挑未取消那一列，且欄位要真的從資料帶出（含 ID／週期／試用到期）。
	if sub.ID != subID || sub.Status != "active" || sub.SeatCount != 3 || sub.PlanCode != "std" ||
		sub.PlanName != "標準" || sub.PlanID != planID || sub.BillingCycle != "monthly" {
		t.Fatalf("42 的訂閱欄位不對（同一公司另有已取消的列，不得回錯那一列），got %+v", *sub)
	}
	if sub.TrialEnds == nil || !sub.TrialEnds.Equal(trialEnds) || sub.GraceUntil != nil {
		t.Fatalf("試用到期日／寬限期未正確帶出，got %+v", *sub)
	}
	periodStart, periodEnd := now, now.Add(720*time.Hour)
	period, err := st.OpenPeriodTx(ctx, tx, store.OpenPeriodInput{
		SubscriptionID: sub.ID, PeriodNo: 1,
		PeriodStart: periodStart, PeriodEnd: periodEnd,
		PlanID: planID, UnitPriceCents: 150000, SeatPriceCents: 15000, SeatCount: 3,
		AmountCents: 195000, Currency: "TWD",
	})
	if err != nil || period == nil {
		t.Fatalf("OpenPeriodTx: got %+v err=%v", period, err)
	}
	// 分 → numeric(12,2)：寫進 DB 的是 1500.00／150.00／1950.00，讀回轉分必須逐分吻合；
	// 期間與快照欄位也要對（欄位順序漂移會靜默取到錯欄位）。
	if period.AmountCents != 195000 || period.UnitPriceCents != 150000 || period.SeatPriceCents != 15000 ||
		period.Status != "open" || period.Currency != "TWD" || period.PeriodNo != 1 ||
		period.SubscriptionID != subID || period.PlanID != planID || period.SeatCount != 3 ||
		period.Note != "" || period.PaidAt != nil ||
		!period.PeriodStart.Equal(periodStart) || !period.PeriodEnd.Equal(periodEnd) {
		t.Fatalf("期別的金額／期間／欄位快照不對，got %+v", *period)
	}
	if err := st.EmitEventTx(ctx, tx, "subscription", sub.ID, "period.opened", []byte(`{"amount":"1950"}`)); err != nil {
		t.Fatalf("EmitEventTx: %v", err)
	}
	if err := st.RecordAuditTx(ctx, tx, opID, "open_period", "subscription", "42", "測試", nil, nil); err != nil {
		t.Fatalf("RecordAuditTx: %v", err)
	}
	// 平台稽核的原因必填（S9）：空字串即拒絕，且不得留下半筆。
	if err := st.RecordAuditTx(ctx, tx, opID, "open_period", "subscription", "42", "", nil, nil); err == nil {
		t.Fatal("RecordAuditTx 的原因為空字串時必須拒絕")
	}
	// 同一交易內可讀：三者確實落在 fn 拿到的那個 tx 上。
	var inTx int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.events WHERE aggregate_id = $1`, sub.ID).Scan(&inTx); err != nil || inTx != 1 {
		t.Fatalf("同一交易內應讀到剛寫的事件，got %d err=%v", inTx, err)
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.audit_logs WHERE operator_id = $1`, opID).Scan(&inTx); err != nil || inTx != 1 {
		t.Fatalf("同一交易內應讀到剛寫的稽核（原因為空者不得寫入），got %d err=%v", inTx, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	for _, q := range []string{
		`SELECT count(*) FROM platform.subscription_periods`,
		`SELECT count(*) FROM platform.events`,
		`SELECT count(*) FROM platform.audit_logs`,
	} {
		var n int
		if err := db.QueryRowContext(ctx, q).Scan(&n); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		if n != 0 {
			t.Fatalf("%s 應為 0（回滾後不得留下痕跡），got %d", q, n)
		}
	}

	// WithTx 的錯誤路徑同樣不留半成品，且**回傳 fn 的錯誤**（呼叫端要能分辨原因）。
	sentinel := errors.New("收款的業務規則拒絕")
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		if err := st.EmitEventTx(ctx, tx, "subscription", subID, "period.opened", []byte(`{}`)); err != nil {
			return err
		}
		return sentinel
	}); !errors.Is(err, sentinel) {
		t.Fatalf("WithTx 應回傳 fn 的錯誤，got %v", err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM platform.events`).Scan(&n); err != nil {
		t.Fatalf("數事件: %v", err)
	}
	if n != 0 {
		t.Fatalf("WithTx 內 fn 回錯誤時不得提交（事件應為 0），got %d", n)
	}

	// ② 列鎖：另一交易取同一訂閱應被擋到第一交易結束。
	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	defer func() { _ = tx1.Rollback() }()
	if _, err := st.OpenSubscriptionTx(ctx, tx1, 42); err != nil {
		t.Fatalf("tx1 鎖定: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		tx2, err := db.BeginTx(ctx, nil)
		if err != nil {
			done <- err
			return
		}
		defer func() { _ = tx2.Rollback() }()
		if _, err := st.OpenSubscriptionTx(ctx, tx2, 42); err != nil {
			done <- err
			return
		}
		done <- nil
	}()
	select {
	case err := <-done:
		t.Fatalf("第二個交易不應在 tx1 持有鎖時取得列（got err=%v）", err)
	case <-time.After(300 * time.Millisecond):
		// 預期：仍被鎖住。
	}
	// 放掉鎖之後第二個交易必須真的完成 —— 否則上面的「被擋住」可能只是它本來就失敗。
	if err := tx1.Rollback(); err != nil {
		t.Fatalf("tx1 rollback: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("tx1 結束後第二個交易應取得列，got err=%v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("tx1 結束後第二個交易仍未取得列（鎖未釋放）")
	}

	// ③ 只有一筆 cancelled 的公司必須回傳它（F-8／C-02）：預先濾掉會讓「已取消」與「從未訂閱」
	// 在判定層不可區分，等於取消即送免費方案。
	if got := openSub(t, db, st, 43); got.Status != "cancelled" || got.SeatCount != 99 || got.BillingCycle != "yearly" {
		t.Fatalf("回傳的必須是那一列已取消的訂閱（狀態／席次／週期皆要帶出），got %+v", *got)
	}
	if got := withTx(t, db, func(tx *sql.Tx) (*store.Subscription, error) {
		return st.OpenSubscriptionTx(ctx, tx, 44)
	}); got != nil {
		t.Fatalf("完全沒有訂閱列應回 (nil, nil)，got %+v", *got)
	}

	// ④ 期別冪等鍵：(subscription_id, period_no) 已存在即回既有期別，不新增。
	openOne := func() *store.Period {
		return withTx(t, db, func(tx *sql.Tx) (*store.Period, error) {
			return st.OpenPeriodTx(ctx, tx, store.OpenPeriodInput{
				SubscriptionID: subID, PeriodNo: 1,
				PeriodStart: periodStart, PeriodEnd: periodEnd,
				PlanID: planID, UnitPriceCents: 150000, SeatPriceCents: 15000, SeatCount: 3,
				AmountCents: 195000, Currency: "TWD",
			})
		})
	}
	opened := openOne()
	if again := openOne(); again.ID != opened.ID {
		t.Fatalf("同一期別重複開啟必須回同一列，got %d → %d", opened.ID, again.ID)
	}
	var periodCount int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.subscription_periods WHERE subscription_id = $1`, subID).Scan(&periodCount); err != nil {
		t.Fatalf("數期別: %v", err)
	}
	if periodCount != 1 {
		t.Fatalf("排程重跑不得產生重複期別，got %d 期", periodCount)
	}
	// 直接插入同一 (subscription_id, period_no) 必須被唯一約束擋下（23505）—— 這條 UNIQUE 是
	// 「重跑不產生重複期別」唯一的地基，少了它 ON CONFLICT 的冪等也無從成立。
	_, err = db.ExecContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount)
		VALUES ($1, 1, now(), now(), $2, 1500.00, 150.00, 3, 1950.00)`, subID, planID)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("重複的 (subscription_id, period_no) 必須以唯一性違反(23505)失敗，got %v", err)
	}
	if got := withTx(t, db, func(tx *sql.Tx) (*store.Period, error) {
		return st.OpenPeriodByNoTx(ctx, tx, subID, 1)
	}); got.ID != opened.ID {
		t.Fatalf("OpenPeriodByNoTx 應回同一期別，got %+v", *got)
	}
	// 不存在的期別：只認 sql.ErrNoRows（C-19：其他錯誤不得被當成「不存在」）。
	noSuchPeriod := func() error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		_, err = st.OpenPeriodByNoTx(ctx, tx, subID, 99)
		return err
	}
	if err := noSuchPeriod(); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("不存在的期別應回 sql.ErrNoRows，got %v", err)
	}

	// ⑤ MarkPeriodPaidTx：標記付款、同交易號冪等、note 不得被空字串清掉。
	paidAt := now
	stamp := func(externalRef, note string) error {
		return st.WithTx(ctx, func(tx *sql.Tx) error {
			return st.MarkPeriodPaidTx(ctx, tx, opened.ID, paidAt, "INV-2026-001", "", "", "", "manual", externalRef, note)
		})
	}
	if err := stamp("REF-1", "短收 100 元"); err != nil {
		t.Fatalf("MarkPeriodPaidTx: %v", err)
	}
	paid := currentPeriod(t, db, st, subID)
	if paid.Status != "paid" || paid.InvoiceNo != "INV-2026-001" || paid.PaymentProvider != "manual" ||
		paid.ExternalRef != "REF-1" || paid.Note != "短收 100 元" {
		t.Fatalf("付款欄位未寫入，got %+v", *paid)
	}
	if paid.PaidAt == nil || !paid.PaidAt.Equal(paidAt) {
		t.Fatalf("paid_at 未寫入，got %v", paid.PaidAt)
	}
	// 同一筆交易號重送（webhook 重播）：no-op，不得報錯。
	if err := stamp("REF-1", ""); err != nil {
		t.Fatalf("同交易號重送應視為重複入帳（no-op），got %v", err)
	}
	if again := currentPeriod(t, db, st, subID); again.Note != "短收 100 元" {
		t.Fatalf("note 為空字串時必須保留原值，got %q", again.Note)
	}
	// 不同交易號 → 拒絕覆蓋別筆收款。
	if err := stamp("REF-2", ""); err == nil {
		t.Fatal("已付款且交易號不同時必須拒絕覆蓋")
	}
	if again := currentPeriod(t, db, st, subID); again.ExternalRef != "REF-1" {
		t.Fatalf("被拒絕的收款不得改動原交易號，got %q", again.ExternalRef)
	}
	// 重播（webhook 重送／T4 重試）是 **no-op**：同交易號、但沒重帶發票號與 provider、時間點也不同，
	// 已入帳的憑據一個都不得被改掉 —— 發票號與 provider 是稅務與對帳憑據，被空字串清成 NULL
	// 等於把帳抹掉（G8）。
	later := paidAt.Add(time.Hour)
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.MarkPeriodPaidTx(ctx, tx, opened.ID, later, "", "", "", "", "", "REF-1", "")
	}); err != nil {
		t.Fatalf("同交易號重播應為 no-op，got %v", err)
	}
	if again := currentPeriod(t, db, st, subID); again.InvoiceNo != "INV-2026-001" ||
		again.PaymentProvider != "manual" || again.ExternalRef != "REF-1" ||
		again.PaidAt == nil || !again.PaidAt.Equal(paidAt) || again.Note != "短收 100 元" {
		t.Fatalf("重播不得改動已入帳的憑據（發票號／provider／paid_at／交易號／note），got %+v", *again)
	}
	if n := countPaid(t, db, subID); n != 1 {
		t.Fatalf("重播後仍應只有 1 筆已付款期別，got %d", n)
	}
	// 同一筆交易號不得入帳**兩期**（00029 的 periods_provider_ref_unique）：webhook 對錯期別時
	// 必須失敗，而不是把同一筆錢記在兩期上。
	//
	// fixture 用自己的第二個訂閱（狀態 suspended，故不會出現在任何排程查詢的集合裡）：subID(42)
	// 只能有那一期，否則下面的期別清單與「最新一期」斷言會被這裡的 fixture 汙染。
	var otherSubID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count)
		VALUES (45, $1, 'suspended', 3) RETURNING id`, planID).Scan(&otherSubID); err != nil {
		t.Fatalf("第二個訂閱: %v", err)
	}
	second := withTx(t, db, func(tx *sql.Tx) (*store.Period, error) {
		return st.OpenPeriodTx(ctx, tx, store.OpenPeriodInput{
			SubscriptionID: otherSubID, PeriodNo: 1,
			PeriodStart: now, PeriodEnd: periodEnd,
			PlanID: planID, UnitPriceCents: 150000, SeatPriceCents: 15000, SeatCount: 3,
			AmountCents: 195000, Currency: "TWD",
		})
	})
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.MarkPeriodPaidTx(ctx, tx, second.ID, paidAt, "INV-2026-002", "", "", "", "manual", "REF-1", "")
	}); err == nil {
		t.Fatal("同一 provider＋交易號不得入帳兩期（唯一索引）")
	}
	if n := countPaid(t, db, subID); n != 1 {
		t.Fatalf("被拒絕的跨期入帳不得改動任何一期的狀態，42 的已付款期數應仍為 1，got %d", n)
	}
	if n := countPaid(t, db, otherSubID); n != 0 {
		t.Fatalf("被拒絕的跨期入帳不得把第二期記成已付款，got %d", n)
	}
	// 作廢（void）的期別不得入帳，且錯誤要說出真正的原因（不是「已付款」）。
	if _, err := db.ExecContext(ctx,
		`UPDATE platform.subscription_periods SET status = 'void' WHERE id = $1`, second.ID); err != nil {
		t.Fatalf("作廢期別: %v", err)
	}
	err = st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.MarkPeriodPaidTx(ctx, tx, second.ID, paidAt, "INV-2026-002", "", "", "", "manual", "", "")
	})
	if err == nil {
		t.Fatal("作廢期別不得入帳")
	}
	if !strings.Contains(err.Error(), "void") {
		t.Fatalf("作廢期別的錯誤訊息必須說出狀態（不得講成「已付款」），got %v", err)
	}
	var voidStatus string
	if err := db.QueryRowContext(ctx,
		`SELECT status FROM platform.subscription_periods WHERE id = $1`, second.ID).Scan(&voidStatus); err != nil {
		t.Fatalf("查作廢期別狀態: %v", err)
	}
	if voidStatus != "void" {
		t.Fatalf("被拒絕的入帳不得改動期別狀態，got %q", voidStatus)
	}

	// ⑥ 訂閱狀態變更（收款復原／逾期轉移用）。
	grace := now.Add(48 * time.Hour)
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, subID, "past_due", &grace)
	}); err != nil {
		t.Fatalf("SetSubscriptionStatusTx: %v", err)
	}
	if got := openSub(t, db, st, 42); got.Status != "past_due" || got.GraceUntil == nil || !got.GraceUntil.Equal(grace) {
		t.Fatalf("狀態／寬限期未寫入，got %+v", *got)
	}
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, subID, "active", nil)
	}); err != nil {
		t.Fatalf("SetSubscriptionStatusTx(復原): %v", err)
	}
	if got := openSub(t, db, st, 42); got.Status != "active" || got.GraceUntil != nil {
		t.Fatalf("復原後應清空寬限期，got %+v", *got)
	}
	// 轉 cancelled 時一併記下 cancelled_at（狀態轉移的時間點是事後查帳的線索；
	// 直接 INSERT status='cancelled' 的 fixture 不會有它，故這裡一定要由 store 寫出來）。
	sub43 := openSub(t, db, st, 43)
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, sub43.ID, "cancelled", nil)
	}); err != nil {
		t.Fatalf("SetSubscriptionStatusTx(cancelled): %v", err)
	}
	var cancelledAt sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT cancelled_at FROM platform.subscriptions WHERE id = $1`, sub43.ID).Scan(&cancelledAt); err != nil {
		t.Fatalf("查 cancelled_at: %v", err)
	}
	if !cancelledAt.Valid {
		t.Fatal("轉為 cancelled 時必須寫入 cancelled_at")
	}
	// 重複取消（排程重跑／重送的取消請求）不得推進 cancelled_at：它記的是**取消發生的時間**，
	// 被每次重跑推進去就不再是事實。
	firstCancelledAt := cancelledAt.Time
	time.Sleep(2 * time.Millisecond)
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, sub43.ID, "cancelled", nil)
	}); err != nil {
		t.Fatalf("重複取消: %v", err)
	}
	var againAt sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT cancelled_at FROM platform.subscriptions WHERE id = $1`, sub43.ID).Scan(&againAt); err != nil {
		t.Fatalf("重複取消後查 cancelled_at: %v", err)
	}
	if !againAt.Valid || !againAt.Time.Equal(firstCancelledAt) {
		t.Fatalf("重複取消不得推進 cancelled_at，got %v → %v", firstCancelledAt, againAt.Time)
	}
	// 訂閱不存在時**不得靜默成功**：0 列被改到卻回 nil，呼叫端（T4 復原、T5 逾期／取消）會以為
	// 狀態已改，並在同一交易內照樣 commit 事件與稽核 → 留下「稽核說 suspended、DB 仍是 active」。
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, subID+1000, "suspended", nil)
	}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("對不存在的訂閱改狀態應回 sql.ErrNoRows，got %v", err)
	}
	// 未結項 #12（RED）：CAS —— 預期狀態不符時 0 列，必須回 ErrStatusChanged
	// （不是 ErrNoRows），且不得改動狀態。否則併發寫入者的狀態會被靜默蓋掉。
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, subID, "suspended", nil, "past_due")
	}); !errors.Is(err, store.ErrStatusChanged) {
		t.Fatalf("預期狀態不符應回 ErrStatusChanged，got %v", err)
	}
	if got := openSub(t, db, st, 42); got.Status != "active" {
		t.Fatalf("CAS 未命中不得改動狀態，got %q", got.Status)
	}
	// 對照組：預期相符時照常寫入（CAS 不是一律拒絕）。
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, subID, "suspended", nil, "active")
	}); err != nil {
		t.Fatalf("預期相符應寫入，got %v", err)
	}
	if got := openSub(t, db, st, 42); got.Status != "suspended" {
		t.Fatalf("CAS 命中應寫入，got %q", got.Status)
	}
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.SetSubscriptionStatusTx(ctx, tx, subID, "active", nil)
	}); err != nil {
		t.Fatalf("SetSubscriptionStatusTx(復原): %v", err)
	}
	// 50 逾期未付、51 期末未到、52 寬限已過、53 寬限未到、54 past_due 無寬限期、
	// 55 已取消且期末已過（未發過 expired）、56 已取消但期末未到、57 試用中年繳。
	seedSub := func(company int, status, billingCycle string, graceUntil *time.Time) int64 {
		var id int64
		if err := db.QueryRowContext(ctx, `
			INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle, grace_until)
			VALUES ($1,$2,$3,2,$4,$5) RETURNING id`, company, planID, status, billingCycle, graceUntil).Scan(&id); err != nil {
			t.Fatalf("subscription(company %d): %v", company, err)
		}
		return id
	}
	seedPeriod := func(id int64, no int, end time.Time, status ...string) {
		// 期別狀態預設 open；已付款／作廢的期別要能種得出來（C-1：當期已付款不算逾期）。
		periodStatus := "open"
		if len(status) > 0 {
			periodStatus = status[0]
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.subscription_periods
				(subscription_id, period_no, period_start, period_end, plan_id,
				 unit_price, seat_price, seat_count, amount, status)
			VALUES ($1,$2,$3,$4,$5,1500.00,150.00,2,1800.00,$6)`,
			id, no, end.Add(-720*time.Hour), end, planID, periodStatus); err != nil {
			t.Fatalf("period(sub %d): %v", id, err)
		}
	}
	pastGrace := now.Add(-24 * time.Hour)
	// 排程 fixture 的訂閱 id 逐一留存：PeriodsByStatus 只按狀態過濾（無租戶條件），
	// 這裡的斷言一律針對自己的列，不用全表計數 —— 下一次新增 fixture 才不會又紅（F-1）。
	dueSub := seedSub(50, "active", "monthly", nil)
	seedPeriod(dueSub, 1, now.Add(-24*time.Hour))
	futureSub := seedSub(51, "active", "monthly", nil)
	seedPeriod(futureSub, 3, now.Add(24*time.Hour))
	gracePastSub := seedSub(52, "past_due", "yearly", &pastGrace)
	seedPeriod(gracePastSub, 1, now.Add(-24*time.Hour))
	graceFuture := now.Add(24 * time.Hour)
	graceFutureSub := seedSub(53, "past_due", "yearly", &graceFuture)
	seedPeriod(graceFutureSub, 1, now.Add(-24*time.Hour))
	noGraceSub := seedSub(54, "past_due", "monthly", nil)
	seedPeriod(noGraceSub, 1, now.Add(-24*time.Hour))
	cancelledSub := seedSub(55, "cancelled", "yearly", nil)
	seedPeriod(cancelledSub, 1, now.Add(-24*time.Hour))
	cancelledFutureSub := seedSub(56, "cancelled", "monthly", nil)
	seedPeriod(cancelledFutureSub, 1, now.Add(24*time.Hour))
	trialingSub := seedSub(57, "trialing", "yearly", nil)
	// 58／59：期末已過但當期已付款／已作廢 —— 逾期後才繳清的客戶不得被再次催收（C-1）。
	seedPeriod(seedSub(58, "active", "monthly", nil), 1, now.Add(-24*time.Hour), "paid")
	seedPeriod(seedSub(59, "active", "monthly", nil), 1, now.Add(-24*time.Hour), "void")

	due := withTx(t, db, func(tx *sql.Tx) ([]store.Subscription, error) {
		return st.ActiveSubscriptionsWithDueOpenPeriod(ctx, tx, now)
	})
	if !containsCompany(due, 50) || containsCompany(due, 42) || containsCompany(due, 51) ||
		containsCompany(due, 52) || containsCompany(due, 58) || containsCompany(due, 59) {
		t.Fatalf("逾期未付應只含 50（active ＋ 當期仍 open；42／51 期末未到、52 非 active、"+
			"58 已付款、59 已作廢），got %v", companyIDs(due))
	}
	// 排程拿到的訂閱必須帶得動期別產生的欄位：漏帶 billing_cycle 就是 G1（年繳只加一個月、
	// 少收 11 個月）。三條查詢都逐一驗，因為它們各自是不同的一段 SQL。
	if got := subscription(t, due, 50); got.BillingCycle != "monthly" || got.SeatCount != 2 || got.PlanID != planID {
		t.Fatalf("逾期未付的列必須帶出週期／席次／方案，got %+v", got)
	}
	graceExpired := withTx(t, db, func(tx *sql.Tx) ([]store.Subscription, error) {
		return st.PastDueSubscriptionsExpiredGrace(ctx, tx, now)
	})
	// 54 的 grace_until 是 NULL：沒設定寬限期不等於「已到期」，不得誤凍結。
	if !containsCompany(graceExpired, 52) || containsCompany(graceExpired, 53) || containsCompany(graceExpired, 54) {
		t.Fatalf("寬限已過應只含 52（53 未到、54 無寬限期），got %v", companyIDs(graceExpired))
	}
	if got := subscription(t, graceExpired, 52); got.BillingCycle != "yearly" || got.SeatCount != 2 || got.PlanID != planID {
		t.Fatalf("寬限已過的列必須帶出 billing_cycle=yearly（G1），got %+v", got)
	}
	expired := withTx(t, db, func(tx *sql.Tx) ([]store.Subscription, error) {
		return st.CancelledSubscriptionsPastPeriodEnd(ctx, tx, now)
	})
	if !containsCompany(expired, 55) || containsCompany(expired, 56) {
		t.Fatalf("已取消且期末已過應只含 55（56 期末未到），got %v", companyIDs(expired))
	}
	if got := subscription(t, expired, 55); got.BillingCycle != "yearly" || got.SeatCount != 2 || got.PlanID != planID {
		t.Fatalf("已取消期末已過的列必須帶出 billing_cycle=yearly（G1），got %+v", got)
	}
	// 排程可重跑：已發過 subscription.expired 者不得再被選中。
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.EmitEventTx(ctx, tx, "subscription", cancelledSub, "subscription.expired", []byte(`{}`))
	}); err != nil {
		t.Fatalf("EmitEventTx(expired): %v", err)
	}
	// 不帶資料的事件（nil／空 payload，例：subscription.suspended）必須寫得進去：送 ''::jsonb 會
	// 22P02，而表的 DEFAULT '{}' 永遠不會生效 —— 單元測試（假實作）擋不住這一條，只有真 store 會炸。
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		return st.EmitEventTx(ctx, tx, "subscription", cancelledSub, "subscription.suspended", nil)
	}); err != nil {
		t.Fatalf("無 payload 的事件必須寫得進去（payload 為 '{}'），got %v", err)
	}
	expiredAgain := withTx(t, db, func(tx *sql.Tx) ([]store.Subscription, error) {
		return st.CancelledSubscriptionsPastPeriodEnd(ctx, tx, now)
	})
	if containsCompany(expiredAgain, 55) {
		t.Fatalf("已發過 subscription.expired 的訂閱不得再被選中（重跑不得重複發事件），got %v", companyIDs(expiredAgain))
	}
	serving, err := st.ActiveOrTrialingSubscriptions(ctx)
	if err != nil {
		t.Fatalf("ActiveOrTrialingSubscriptions: %v", err)
	}
	if !containsCompany(serving, 42) || !containsCompany(serving, 50) || !containsCompany(serving, 57) ||
		containsCompany(serving, 52) || containsCompany(serving, 55) {
		t.Fatalf("仍在服務中的只有 active／trialing（42／50／57），got %v", companyIDs(serving))
	}
	// 年繳方案必須帶出 billing_cycle：期別產生靠它決定 +1 月或 +1 年（G1）。
	for _, s := range serving {
		if s.CompanyID == 57 && (s.BillingCycle != "yearly" || s.Status != "trialing" || s.SeatCount != 2 || s.PlanID != planID) {
			t.Fatalf("57 的欄位未帶出（年繳週期／狀態／席次／方案），got %+v", s)
		}
	}

	// ⑧ 期別與價目的讀取路徑。
	if cur := currentPeriod(t, db, st, subID); cur.ID != opened.ID {
		t.Fatalf("CurrentPeriodTx 應回最新一期，got %+v", *cur)
	}
	if cur := currentPeriod(t, db, st, trialingSub); cur != nil {
		t.Fatalf("沒有期別的訂閱應回 (nil, nil)（不是錯誤），got %+v", *cur)
	}
	// PeriodsByStatus 只按狀態過濾（沒有租戶／訂閱條件），而 fixture 會隨時間增加 ——
	// 斷言必須只針對**本測試自己的**訂閱，不得假設全表只有一筆（F-1）。
	paidList, err := st.PeriodsByStatus(ctx, "paid")
	if err != nil {
		t.Fatalf("PeriodsByStatus(paid): %v", err)
	}
	var mine []store.Period
	for _, p := range paidList {
		if p.SubscriptionID == subID {
			mine = append(mine, p)
		}
	}
	if len(mine) != 1 || mine[0].ID != opened.ID || mine[0].Note != "短收 100 元" {
		t.Fatalf("本訂閱（%d）已付款的期別應只有那一期（含 note），got %+v（全表 %d 筆）",
			subID, mine, len(paidList))
	}
	// open 清單同樣只按狀態過濾：只斷言「排程 fixture 的那 7 期仍在」＋「42 已付款的那一期不在」
	// —— 不用全表計數，新增 fixture 不會讓它失真（F-1）。
	openList, err := st.PeriodsByStatus(ctx, "open")
	if err != nil {
		t.Fatalf("PeriodsByStatus(open): %v", err)
	}
	open := make(map[int64]bool, len(openList))
	for _, p := range openList {
		open[p.SubscriptionID] = true
	}
	for _, id := range []int64{dueSub, futureSub, gracePastSub, graceFutureSub,
		noGraceSub, cancelledSub, cancelledFutureSub} {
		if !open[id] {
			t.Fatalf("排程 fixture 的訂閱 %d 應仍有一期 open，got %v", id, open)
		}
	}
	if open[subID] {
		t.Fatalf("42 已付款的那一期不得留在 open 清單，got %v", open)
	}
	// OverdueReceivablePeriods：open 且期末已過，且排除 G5 平台自營公司
	// （與 console 的 ListReceivables 同一謂詞 `identifier <> 'platform'`）。
	// 用真實 companies 列驗 JOIN：60 是一般租戶（identifier 非 platform），
	// 'platform' 是自營公司 —— 期別同樣逾期未付，只有前者算待收款。
	var platformCompanyID, normalCompanyID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO companies (name, identifier, status) VALUES ('一般租戶','OVERDUE-60','active') RETURNING id`).Scan(&normalCompanyID); err != nil {
		t.Fatalf("一般租戶公司: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`INSERT INTO companies (name, identifier, status) VALUES ('平台自營','platform','active') RETURNING id`).Scan(&platformCompanyID); err != nil {
		t.Fatalf("平台自營公司: %v", err)
	}
	var normalSubID, platformSubID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle)
		VALUES ($1,$2,'active',1,'monthly') RETURNING id`, normalCompanyID, planID).Scan(&normalSubID); err != nil {
		t.Fatalf("一般租戶訂閱: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, billing_cycle)
		VALUES ($1,$2,'active',1,'monthly') RETURNING id`, platformCompanyID, planID).Scan(&platformSubID); err != nil {
		t.Fatalf("自營公司訂閱: %v", err)
	}
	for _, id := range []int64{normalSubID, platformSubID} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.subscription_periods
				(subscription_id, period_no, period_start, period_end, plan_id,
				 unit_price, seat_price, seat_count, amount, status)
			VALUES ($1,1,$2,$3,$4,1500.00,150.00,1,1650.00,'open')`,
			id, now.Add(-48*time.Hour), now.Add(-24*time.Hour), planID); err != nil {
			t.Fatalf("逾期期別(sub %d): %v", id, err)
		}
	}
	overdue, err := st.OverdueReceivablePeriods(ctx, now)
	if err != nil {
		t.Fatalf("OverdueReceivablePeriods: %v", err)
	}
	found := map[int64]bool{}
	for _, p := range overdue {
		found[p.SubscriptionID] = true
	}
	if !found[normalSubID] {
		t.Fatalf("一般租戶的逾期期別必須是待收款，got %d 筆", len(overdue))
	}
	if found[platformSubID] {
		t.Fatalf("平台自營公司的期別不得是待收款（與 ListReceivables 一致）")
	}
	price := withTx(t, db, func(tx *sql.Tx) (store.Price, error) {
		return st.CurrentPriceTx(ctx, tx, planID, "monthly")
	})
	if price.BaseCents != 150000 || price.SeatCents != 15000 || price.Currency != "TWD" {
		t.Fatalf("CurrentPriceTx(monthly) = %+v；want 150000／15000／TWD", price)
	}
	// 沒有該週期的價目必須是錯誤（不得靜默用 0 → 免費送方案）。
	if err := st.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := st.CurrentPriceTx(ctx, tx, planID, "yearly")
		return err
	}); err == nil {
		t.Fatal("方案沒有該週期價目時必須回錯誤，不得靜默回 0 元")
	}

	// ⑨ outbox：未派送事件依 id 取出，標記後不再出現（consumer 的重試語意）。
	events, err := st.UndispatchedEvents(ctx, 10)
	if err != nil {
		t.Fatalf("UndispatchedEvents: %v", err)
	}
	if len(events) != 2 || events[0].AggregateType != "subscription" || events[0].AggregateID != cancelledSub ||
		events[0].EventType != "subscription.expired" || events[1].EventType != "subscription.suspended" {
		t.Fatalf("未派送事件應照 id 序為 expired→suspended（WithTx 回錯誤者已回滾），got %+v", events)
	}
	// 無 payload 的事件讀回來是 '{}'（不是 NULL、也不是空字串）——consumer 解 payload 時不必分兩條路。
	if string(events[0].Payload) != "{}" || string(events[1].Payload) != "{}" {
		t.Fatalf("事件 payload 應為 '{}'（含無 payload 者），got %q／%q", events[0].Payload, events[1].Payload)
	}
	for _, e := range events {
		if err := st.WithTx(ctx, func(tx *sql.Tx) error {
			return st.MarkEventDispatchedTx(ctx, tx, e.ID)
		}); err != nil {
			t.Fatalf("MarkEventDispatchedTx(%d): %v", e.ID, err)
		}
	}
	if left, err := st.UndispatchedEvents(ctx, 10); err != nil || len(left) != 0 {
		t.Fatalf("已派送的事件不得再出現，got %+v err=%v", left, err)
	}

	// WithTx 的 panic 路徑必須放掉連線（fn 內 panic 時也要 Rollback）：少了它，交易會懸著佔住
	// 連線（production 是排程 goroutine 的 panic 被 RunGuarded 復原，連線會一路漏到 GC）。
	inUseBefore := db.Stats().InUse
	func() {
		defer func() {
			if recover() == nil {
				t.Error("WithTx 應讓 fn 的 panic 繼續往外傳（由呼叫端復原）")
			}
		}()
		_ = st.WithTx(ctx, func(*sql.Tx) error { panic("排程 panic（處理程序復原）") })
	}()
	if inUse := db.Stats().InUse; inUse > inUseBefore {
		t.Fatalf("WithTx 的 fn panic 後連線未放回池中（InUse %d → %d）：交易必須 defer Rollback",
			inUseBefore, inUse)
	}
}

// withTx 在測試自建的交易內跑 fn、成功即提交並回傳結果（BillingStore 的方法一律要求 *sql.Tx）。
func withTx[T any](t *testing.T, db *sql.DB, fn func(*sql.Tx) (T, error)) T {
	t.Helper()
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	v, err := fn(tx)
	if err != nil {
		t.Fatalf("交易內操作: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return v
}

// openSub 取某公司的現行訂閱（走與排程同一條取列路徑；無列即測試失敗）。
func openSub(t *testing.T, db *sql.DB, st *postgres.Store, company int) *store.Subscription {
	t.Helper()
	sub := withTx(t, db, func(tx *sql.Tx) (*store.Subscription, error) {
		return st.OpenSubscriptionTx(t.Context(), tx, company)
	})
	if sub == nil {
		t.Fatalf("company %d 應有訂閱列", company)
	}
	return sub
}

// currentPeriod 取某訂閱最新一期（無期別回 nil）。
func currentPeriod(t *testing.T, db *sql.DB, st *postgres.Store, subID int64) *store.Period {
	t.Helper()
	return withTx(t, db, func(tx *sql.Tx) (*store.Period, error) {
		return st.CurrentPeriodTx(t.Context(), tx, subID)
	})
}

// companyIDs 抽出查詢回傳的 company_id（失敗訊息用）。
func companyIDs(subs []store.Subscription) []int {
	out := make([]int, 0, len(subs))
	for _, s := range subs {
		out = append(out, s.CompanyID)
	}
	return out
}

// subscription 取集合中該公司的那一列（找不到即測試失敗）——排程查詢回傳的欄位要逐欄驗，
// 只驗集合成員會漏掉「欄位沒帶出來」（billing_cycle 空字串就是少收 11 個月的 G1）。
func subscription(t *testing.T, subs []store.Subscription, company int) store.Subscription {
	t.Helper()
	for _, s := range subs {
		if s.CompanyID == company {
			return s
		}
	}
	t.Fatalf("集合中應有 company %d，got %v", company, companyIDs(subs))
	return store.Subscription{}
}

// countPaid 數該訂閱的已付款期別（交易號冪等與跨期拒絕都要確認「沒有多記一筆帳」）。
func countPaid(t *testing.T, db *sql.DB, subID int64) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM platform.subscription_periods WHERE subscription_id = $1 AND status = 'paid'`,
		subID).Scan(&n); err != nil {
		t.Fatalf("數已付款期別: %v", err)
	}
	return n
}

// containsCompany 回報集合內是否有該公司。
func containsCompany(subs []store.Subscription, company int) bool {
	for _, s := range subs {
		if s.CompanyID == company {
			return true
		}
	}
	return false
}
