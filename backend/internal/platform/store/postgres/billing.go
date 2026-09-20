// 平台帳務的 SQL 寫入(00029／00030)。**每個寫入方法都接受 *sql.Tx**:期別、訂閱狀態、事件、
// 稽核必須落在同一個 commit(見 store.BillingStore 的說明);store 不代呼叫端開交易或 commit,
// 唯一的例外是 WithTx —— 它是呼叫端宣告「這一段要一個交易」的地方。
//
// 這些交易是**平台寫入的交易**,不是業務表的租戶交易:本檔的每個查詢都走 admin(owner)連線,
// 且 platform schema 不套 RLS(spec §3.3),故不需要(也不得)SET LOCAL app.current_data_scope ——
// 那是業務表在租戶 interceptor 交易內的慣例。
//
// 金額一律以**分**(int64)進出,寫入時經 money.FormatCents 轉 numeric(12,2),讀取時以
// `(欄位*100)::bigint` 轉回分:中間不經 float64,帳務不接受 0.01 的誤差。
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

var _ store.BillingStore = (*Store)(nil)

// OpenSubscriptionTx 取該租戶的現行訂閱並以 FOR UPDATE 鎖住該列:併發的收款／逾期轉移必須互斥,
// 否則兩個請求可能都通過「目前是 active」的檢查而寫出互相矛盾的狀態。
//
// **不得**把 `s.status <> 'cancelled'` 寫進 WHERE(F-8):那樣「只有一筆已取消訂閱」的公司與
// 「完全沒有訂閱列」在呼叫端不可區分 —— 判定層對後者的語意是「尚未開通計費 → 不施加限制」
// (spec §4.5),等於取消流程變成送免費方案(真容器實測:10/10 席的 cancelled 公司仍可建帳號)。
// 取法與唯讀的 Store.Subscription(postgres/store.go)及平台端投影(admin.go 的 tenantJoins
// LATERAL)**逐字相同**:優先未取消,只有全是 cancelled 時才取 cancelled;partial unique index
// 保證未取消者至多一筆,故 LIMIT 1 不會少算。
func (s *Store) OpenSubscriptionTx(ctx context.Context, tx *sql.Tx, companyID int) (*store.Subscription, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT s.id, s.company_id, p.code, p.name, s.status, p.id, s.seat_count,
		       s.billing_cycle, s.trial_ends_at, s.grace_until
		  FROM platform.subscriptions s
		  JOIN platform.plans p ON p.id = s.plan_id
		 WHERE s.company_id = $1
		 ORDER BY (s.status = 'cancelled'), s.started_at DESC, s.id DESC
		 LIMIT 1
		 FOR UPDATE OF s`, companyID)
	var sub store.Subscription
	var trial, grace sql.NullTime
	err := row.Scan(&sub.ID, &sub.CompanyID, &sub.PlanCode, &sub.PlanName, &sub.Status, &sub.PlanID,
		&sub.SeatCount, &sub.BillingCycle, &trial, &grace)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if trial.Valid {
		v := trial.Time
		sub.TrialEnds = &v
	}
	if grace.Valid {
		v := grace.Time
		sub.GraceUntil = &v
	}
	return &sub, nil
}

// SetSubscriptionStatusTx 更新訂閱狀態與寬限期(cancelled 時一併記下 cancelled_at)。
func (s *Store) SetSubscriptionStatusTx(ctx context.Context, tx *sql.Tx, subID int64,
	status string, graceUntil *time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE platform.subscriptions
		   SET status = $2, grace_until = $3, updated_at = now(),
		       cancelled_at = CASE WHEN $2 = 'cancelled' THEN now() ELSE cancelled_at END
		 WHERE id = $1`, subID, status, graceUntil)
	return err
}

// OpenPeriodTx 建立一期(open)。unique(subscription_id, period_no) 讓排程可重跑:
// 已存在即回既有期別,**不新增也不覆寫**(既有期別的金額是已開帳的事實,重跑不得改動它)。
func (s *Store) OpenPeriodTx(ctx context.Context, tx *sql.Tx, in store.OpenPeriodInput) (*store.Period, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, currency, status)
		VALUES ($1,$2,$3,$4,$5,$6::numeric,$7::numeric,$8,$9::numeric,$10,'open')
		ON CONFLICT (subscription_id, period_no) DO NOTHING
		RETURNING id`, in.SubscriptionID, in.PeriodNo, in.PeriodStart, in.PeriodEnd, in.PlanID,
		money.FormatCents(in.UnitPriceCents), money.FormatCents(in.SeatPriceCents), in.SeatCount,
		money.FormatCents(in.AmountCents), in.Currency)
	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) { // 已存在(DO NOTHING 沒有 RETURNING 列)
			return s.OpenPeriodByNoTx(ctx, tx, in.SubscriptionID, in.PeriodNo)
		}
		return nil, err
	}
	return s.periodByIDTx(ctx, tx, id)
}

// OpenPeriodByNoTx 以 (subscription_id, period_no) 取期別;不存在時回 sql.ErrNoRows。
func (s *Store) OpenPeriodByNoTx(ctx context.Context, tx *sql.Tx, subID int64, periodNo int) (*store.Period, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT `+periodColumns+`
		  FROM platform.subscription_periods
		 WHERE subscription_id = $1 AND period_no = $2`, subID, periodNo)
	return scanPeriod(row)
}

// CurrentPeriodTx 取訂閱最新一期(期別產生與逾期判定用);無期別回 (nil, nil) ——
// 「還沒有任何一期」不是錯誤。
func (s *Store) CurrentPeriodTx(ctx context.Context, tx *sql.Tx, subID int64) (*store.Period, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT `+periodColumns+`
		  FROM platform.subscription_periods
		 WHERE subscription_id = $1
		 ORDER BY period_no DESC
		 LIMIT 1`, subID)
	p, err := scanPeriod(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

// CurrentPriceTx 取該方案在指定計費週期的「當期生效價」(effective_from 最新且已生效者)。
// 期別金額一律用當期價快照:方案調價後新期別用新價、舊期別不變(調價不得回溯改帳)。
func (s *Store) CurrentPriceTx(ctx context.Context, tx *sql.Tx, planID int64, cycle string) (store.Price, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT (base_price*100)::bigint, (seat_price*100)::bigint, currency
		  FROM platform.plan_prices
		 WHERE plan_id = $1 AND billing_cycle = $2 AND effective_from <= now()
		 ORDER BY effective_from DESC
		 LIMIT 1`, planID, cycle)
	var p store.Price
	if err := row.Scan(&p.BaseCents, &p.SeatCents, &p.Currency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 不得靜默回 0 元:沒有價目等於免費送方案,必須讓呼叫端停下來。
			return store.Price{}, fmt.Errorf("方案 %d 沒有 %s 週期的價目", planID, cycle)
		}
		return store.Price{}, err
	}
	return p, nil
}

// MarkPeriodPaidTx 標記期別為已付款。三種情境各自有明確結果:
//   - status='open' → 入帳;
//   - 已是 paid 且**交易號相同** → 視為重複入帳(webhook 重播),no-op 且不報錯;
//   - 已是 paid 但交易號不同 → 錯誤(不得覆蓋別筆收款)。
//
// note(G8)為短收／溢收的人工註記:未提供(空字串)時**保留原值** —— 用空字串清掉既有註記
// 等於抹掉對帳線索。
func (s *Store) MarkPeriodPaidTx(ctx context.Context, tx *sql.Tx, id int64, paidAt time.Time,
	invoiceNo, provider, externalRef, note string) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE platform.subscription_periods
		   SET status = 'paid', paid_at = $2, invoice_no = NULLIF($3,''),
		       payment_provider = $4, external_ref = NULLIF($5,''),
		       note = COALESCE(NULLIF($6,''), note)
		 WHERE id = $1 AND (
		    status = 'open'
		    OR (status = 'paid' AND COALESCE(external_ref,'') = COALESCE(NULLIF($5,''),''))
		 )`, id, paidAt, invoiceNo, provider, externalRef, note)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// 沒有任何列被改到:分開「期別不存在」與「交易號衝突」,呼叫端才知道該顯示什麼。
		var status string
		var ref sql.NullString
		err := tx.QueryRowContext(ctx,
			`SELECT status, external_ref FROM platform.subscription_periods WHERE id = $1`, id).
			Scan(&status, &ref)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("期別 %d 不存在: %w", id, sql.ErrNoRows)
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("期別 %d 已付款（交易號 %q）且交易號不同（%q），拒絕覆蓋",
			id, ref.String, externalRef)
	}
	return nil
}

// PeriodsByStatus 取指定狀態的期別(收款清單與對帳用),依 (subscription_id, period_no) 排序。
func (s *Store) PeriodsByStatus(ctx context.Context, status string) ([]store.Period, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+periodColumns+`
		  FROM platform.subscription_periods
		 WHERE status = $1
		 ORDER BY subscription_id, period_no`, status)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []store.Period
	for rows.Next() {
		p, err := scanPeriod(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// ActiveSubscriptionsWithDueOpenPeriod 回 active 且最新一期已過期末者(排程轉 past_due)。
func (s *Store) ActiveSubscriptionsWithDueOpenPeriod(ctx context.Context, tx *sql.Tx, now time.Time) ([]store.Subscription, error) {
	return s.scanSubscriptions(ctx, tx, `
		SELECT s.id, s.company_id, s.status
		  FROM platform.subscriptions s
		  JOIN LATERAL (
			SELECT period_end FROM platform.subscription_periods p
			 WHERE p.subscription_id = s.id
			 ORDER BY p.period_no DESC
			 LIMIT 1
		  ) cur ON true
		 WHERE s.status = 'active' AND cur.period_end < $1`, now)
}

// PastDueSubscriptionsExpiredGrace 回 past_due 且寬限期已過者(排程轉 suspended)。
// grace_until IS NULL 不算「已過」:沒有設定寬限期不等於寬限期已到期。
func (s *Store) PastDueSubscriptionsExpiredGrace(ctx context.Context, tx *sql.Tx, now time.Time) ([]store.Subscription, error) {
	return s.scanSubscriptions(ctx, tx, `
		SELECT id, company_id, status
		  FROM platform.subscriptions
		 WHERE status = 'past_due' AND grace_until IS NOT NULL AND grace_until < $1`, now)
}

// CancelledSubscriptionsPastPeriodEnd 回 cancelled 且最新一期已過期末者(G7)。
// EXISTS 排除「已發過 subscription.expired」者 → 排程可重跑且不重複發事件。
func (s *Store) CancelledSubscriptionsPastPeriodEnd(ctx context.Context, tx *sql.Tx, now time.Time) ([]store.Subscription, error) {
	return s.scanSubscriptions(ctx, tx, `
		SELECT s.id, s.company_id, s.status
		  FROM platform.subscriptions s
		  JOIN LATERAL (
			SELECT period_end FROM platform.subscription_periods p
			 WHERE p.subscription_id = s.id
			 ORDER BY p.period_no DESC
			 LIMIT 1
		  ) cur ON true
		 WHERE s.status = 'cancelled' AND cur.period_end < $1
		   AND NOT EXISTS (
			SELECT 1 FROM platform.events e
			 WHERE e.aggregate_type = 'subscription' AND e.aggregate_id = s.id
			   AND e.event_type = 'subscription.expired'
		   )`, now)
}

// ActiveOrTrialingSubscriptions 回仍在服務中的訂閱(排程逐租戶產生下一期用):
// billing_cycle 必須帶出,期別產生靠它決定 +1 月或 +1 年(G1)。
func (s *Store) ActiveOrTrialingSubscriptions(ctx context.Context) ([]store.Subscription, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company_id, status, plan_id, seat_count, billing_cycle
		  FROM platform.subscriptions
		 WHERE status IN ('active','trialing')
		 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []store.Subscription
	for rows.Next() {
		var sub store.Subscription
		if err := rows.Scan(&sub.ID, &sub.CompanyID, &sub.Status,
			&sub.PlanID, &sub.SeatCount, &sub.BillingCycle); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

// scanSubscriptions 為三個排程查詢的共用列掃描(欄位形狀相同:id, company_id, status)。
func (s *Store) scanSubscriptions(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]store.Subscription, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []store.Subscription
	for rows.Next() {
		var sub store.Subscription
		if err := rows.Scan(&sub.ID, &sub.CompanyID, &sub.Status); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

// EmitEventTx 寫入 outbox 事件(必須與期別／狀態的變更同一個交易,否則會出現
// 「改了狀態卻沒有事件」或反之的孤兒)。
func (s *Store) EmitEventTx(ctx context.Context, tx *sql.Tx, aggregateType string,
	aggregateID int64, eventType string, payload []byte) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.events (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1,$2,$3,$4::jsonb)`, aggregateType, aggregateID, eventType, string(payload))
	return err
}

// UndispatchedEvents 取未派送事件(依 id 排序:先寫先派送);limit 由呼叫端給(consumer 的批次大小)。
func (s *Store) UndispatchedEvents(ctx context.Context, limit int) ([]store.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload
		  FROM platform.events
		 WHERE dispatched_at IS NULL
		 ORDER BY id
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []store.Event
	for rows.Next() {
		var e store.Event
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// MarkEventDispatchedTx 標記事件已派送;attempts 為可觀測性計數(含重試)。
func (s *Store) MarkEventDispatchedTx(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE platform.events SET dispatched_at = now(), attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

// RecordAuditTx 寫入平台稽核(S9:actor 為 operator_id,不 FK 租戶 users —— 對方沒有這列)。
// reason 必填:動到錢與權限的操作必須留下「為什麼」,空字串即拒絕(不寫半筆)。
func (s *Store) RecordAuditTx(ctx context.Context, tx *sql.Tx, operatorID int64,
	action, targetType, targetID, reason string, before, after []byte) error {
	if reason == "" {
		return errors.New("平台稽核必須提供原因")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.audit_logs
			(operator_id, action, target_type, target_id, reason, before, after)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb)`,
		operatorID, action, targetType, targetID, reason, nullJSON(before), nullJSON(after))
	return err
}

// WithTx 在**admin 連線**上開一個交易並把 *sql.Tx 交給 fn;fn 回錯誤即回滾,否則提交。
//
// 這裡的交易是平台寫入的交易(platform schema 不套 RLS、不經租戶連線),與業務表的請求交易
// (dbtenant 的 interceptor 交易、帶 app.current_data_scope)是兩回事,不可互相混用。
func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// SystemActor 讀系統 actor 的租戶 user id(排程／consumer 改公司狀態時的稽核主體,G5)。
// 值不是整數即錯誤:稽核的 actor 不得靜默變成 0(那會寫出無主的稽核)。
func (s *Store) SystemActor(ctx context.Context) (int64, error) {
	v, err := s.Setting(ctx, "system_actor_user_id")
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("system_actor_user_id 格式錯誤: %q", v)
	}
	return id, nil
}

// Setting 讀平台營運參數(試用／寬限／提前天數)。查無此鍵即 sql.ErrNoRows。
func (s *Store) Setting(ctx context.Context, key string) (string, error) {
	var v string
	if err := s.db.QueryRowContext(ctx,
		`SELECT value FROM platform.settings WHERE key = $1`, key).Scan(&v); err != nil {
		return "", err
	}
	return v, nil
}

// UpsertSettingTx 寫入營運參數(console 維護用;seed 亦寫同一張表,故用 upsert 保持冪等)。
func (s *Store) UpsertSettingTx(ctx context.Context, tx *sql.Tx, key, value string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES ($1,$2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`, key, value)
	return err
}

// periodColumns 為期別欄位順序的**單一來源**:多處 SELECT 若順序漂移,掃描會靜默取到錯欄位
// (金額與期間互換在帳上不會報錯,只會算錯)。金額以 *100 轉回分,與寫入端的 money.FormatCents 對稱。
const periodColumns = `id, subscription_id, period_no, period_start, period_end, plan_id,
	(unit_price*100)::bigint, (seat_price*100)::bigint, seat_count,
	(amount*100)::bigint, currency, status, paid_at, COALESCE(invoice_no,''),
	payment_provider, COALESCE(external_ref,''), note`

// scanPeriod 掃描一列期別(欄位順序見 periodColumns)。
func scanPeriod(sc rowScanner) (*store.Period, error) {
	var p store.Period
	var paid sql.NullTime
	if err := sc.Scan(&p.ID, &p.SubscriptionID, &p.PeriodNo, &p.PeriodStart, &p.PeriodEnd,
		&p.PlanID, &p.UnitPriceCents, &p.SeatPriceCents, &p.SeatCount, &p.AmountCents,
		&p.Currency, &p.Status, &paid, &p.InvoiceNo, &p.PaymentProvider, &p.ExternalRef,
		&p.Note); err != nil {
		return nil, err
	}
	if paid.Valid {
		v := paid.Time
		p.PaidAt = &v
	}
	return &p, nil
}

// periodByIDTx 以主鍵取期別(OpenPeriodTx 寫入後回讀同一列)。
func (s *Store) periodByIDTx(ctx context.Context, tx *sql.Tx, id int64) (*store.Period, error) {
	return scanPeriod(tx.QueryRowContext(ctx,
		`SELECT `+periodColumns+` FROM platform.subscription_periods WHERE id = $1`, id))
}

// nullJSON 把空的 []byte 轉成 SQL NULL(jsonb 的 before／after 沒提供時是 NULL,不是 '{}')。
func nullJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}
