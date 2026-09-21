// 訂閱生命週期的**掃描式**轉移：由 cmd/platform-cron 的單趟排程（cron.RunOnce）呼叫，
// 時間一律由呼叫端給（不在此讀 time.Now，排程才能補跑與測試）。
//
// 每個函式都必須**可重跑**：判斷只依「當前狀態 ＋ 時間」，不依賴呼叫次數 —— 排程會重複執行
// （每日一趟、手動補跑、單飛鎖失效後的第二趟），重跑不得產生第二個期別、第二個事件、第二次
// 狀態轉移。四支掃描查詢本身即帶冪等謂詞（期別期末 < now、grace_until IS NOT NULL、NOT EXISTS
// subscription.expired、trial_ends_at < now），見 store.BillingStore。
//
// 順序由呼叫端決定（cron.RunOnce）：**試用到期 → 逾期 → 停用 → 取消到期 → 產生期別 → 派送事件**。
// 先轉移狀態並派送事件（凍結／停用）再開期別，避免對剛停用的租戶開新期；試用到期排在最前面，
// 因為它是生命週期最早的階段（還在試用的租戶不該被當成逾期），而轉成 past_due 之後的催收／凍結
// 由後面幾支既有掃描接手。
//
// **不寫平台稽核**：platform.audit_logs.operator_id 是 NOT NULL 且 FK 到 platform.operators，
// 排程沒有 operator 主體（store.SystemActor 是租戶 users.id，硬寫會被 FK 擋下）。凍結／復原的
// 稽核由 consumer 經 SetCompanyStatus 落租戶稽核；排程本身的存在則由 cron 的心跳稽核記錄。
package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// EnsureNextPeriod 在到期前 leadDays 天內建立下一期（open，價格快照為當期生效價）。
// 已有下一期即 no-op（回 false）；未進入提前窗、非服務中的訂閱、還沒有任何期別也都回 false。
//
// 期別長度依訂閱的 billing_cycle（月繳 +1 月、年繳 +1 年，G1），月底由 addBillingPeriod 處理。
// companyID 為 0 或該公司沒有訂閱時只是「沒事可做」：不開期，也不回錯誤。
func (b *Billing) EnsureNextPeriod(ctx context.Context, companyID int, now time.Time, leadDays int) (bool, error) {
	created := false
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		sub, err := b.st.OpenSubscriptionTx(ctx, tx, companyID)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		// 只有服務中的訂閱開下一期：逾期／停用／取消的租戶不該被繼續計費。
		if sub == nil || (sub.Status != "active" && sub.Status != "trialing") {
			return nil
		}
		cur, err := b.st.CurrentPeriodTx(ctx, tx, sub.ID)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if cur == nil {
			// 還沒有任何期別（開通流程尚未產生第一期）：無從得知起訖日與期別號，不猜。
			// 這不是錯誤 —— 回錯誤會讓整趟排程中斷，一個資料不全的租戶會擋住其他所有租戶。
			return nil
		}
		if now.Before(cur.PeriodEnd.AddDate(0, 0, -leadDays)) {
			return nil // 尚未進入提前建立窗
		}

		nextNo := cur.PeriodNo + 1
		switch _, err := b.st.OpenPeriodByNoTx(ctx, tx, sub.ID, nextNo); {
		case err == nil:
			// 已有下一期：可重跑（也不得再發一次 period.opened）。
			return nil
		case errors.Is(err, sql.ErrNoRows):
			// 不存在 → 往下建立。store 的契約是「期別不存在回 sql.ErrNoRows」，
			// 其他錯誤（連線中斷、權限）必須中止，不得當成不存在而續開新期。
		default:
			return errcode.SysInternal.Wrap(err)
		}

		// 期別長度的**錨**是「第一期起日的日號」：月底起租的客戶永遠在月底結帳。
		// 不能用 cur.PeriodStart.Day()：2 月把 31 夾成 28 之後，錨點會永久變成 28
		// （1/31 → 2/28 → 3/28 → …）—— 月繳每期縮成 28 天，一年會開出 13 期（多收一期），
		// 帳單日也永久漂移（I-1）。
		anchorDay := cur.PeriodStart.Day()
		switch first, err := b.st.OpenPeriodByNoTx(ctx, tx, sub.ID, 1); {
		case err == nil:
			anchorDay = first.PeriodStart.Day()
		case errors.Is(err, sql.ErrNoRows):
			// 第一期不在（資料被清理或由外部寫入）：退回「當期起日的日號」—— 那是唯一還帶著
			// 客戶帳單日的線索；不猜、也不讓這一筆資料擋住整趟排程。
		default:
			return errcode.SysInternal.Wrap(err)
		}

		// 期別長度以訂閱的 billing_cycle 決定（G1）：空字串與未知值一律大聲失敗 ——
		// 默默當成月繳會讓年繳只收 1 個月（少收 11 個月）。這裡先算期末，讓週期驗證只有一處。
		nextEnd, err := addBillingPeriod(cur.PeriodEnd, anchorDay, sub.BillingCycle)
		if err != nil {
			return errcode.PlatformSubscriptionInactive.Wrap(err, map[string]string{
				"reason": fmt.Sprintf("訂閱 %d 的 billing_cycle 為 %q，無法決定期別長度",
					sub.ID, sub.BillingCycle),
			})
		}
		// 價格取「當期生效價」：期別是快照，調價後新期別用新價、舊期別不變。
		price, err := b.st.CurrentPriceTx(ctx, tx, sub.PlanID, sub.BillingCycle)
		if err != nil {
			// 「沒有價目」是資料問題（operator 要改的是價目設定），「查價失敗」是基礎設施問題
			// （連線中斷、死鎖）—— 兩者都包成 PLAT-3001 會讓整趟排程把這筆當成「單一租戶的髒
			// 資料」繼續跑，並把 operator 指向一個沒壞的價目設定。分流與 CreateSubscription 一致。
			if errors.Is(err, store.ErrNotFound) {
				// 沒有價目不得靜默用 0 元（那等於免費送方案）。
				return errcode.PlatformSubscriptionInactive.Wrap(err, map[string]string{
					"reason": fmt.Sprintf("方案 %d 沒有 %s 週期的生效價，不得開出 0 元期別",
						sub.PlanID, sub.BillingCycle),
				})
			}
			return errcode.SysInternal.Wrap(err)
		}
		amount, err := money.PeriodAmount(price.BaseCents, price.SeatCents, sub.SeatCount)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if _, err := b.st.OpenPeriodTx(ctx, tx, store.OpenPeriodInput{
			SubscriptionID: sub.ID, PeriodNo: nextNo,
			PeriodStart: cur.PeriodEnd, PeriodEnd: nextEnd,
			PlanID: sub.PlanID, UnitPriceCents: price.BaseCents, SeatPriceCents: price.SeatCents,
			SeatCount: sub.SeatCount, AmountCents: amount, Currency: price.Currency,
		}); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if err := b.emit(ctx, tx, sub.ID, "period.opened", map[string]any{
			"company_id":    companyID,
			"period_no":     nextNo,
			"amount_cents":  amount,
			"billing_cycle": sub.BillingCycle,
			// 排程不寫平台稽核（見檔頭），故每個排程事件都自帶 reason ——
			// 事件是補繳／催收／客服追查時唯一的「為什麼」。
			"reason": "scheduled_next_period",
		}); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		created = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return created, nil
}

// ExpireTrials 掃描 **trialing 且試用已到期**的訂閱 → past_due（設 grace_until = now + graceDays），
// 回傳實際轉移的筆數。
//
// 為什麼需要（這是開通介面的閉環；沒有它，試用就是無上界的免費放行）：
//   - 判定層把 trialing 當**可用**（entitlements.usable）→ 試用到期的租戶照樣使用全部權益；
//   - EnsureNextPeriod 把 trialing 當**服務中** → 每期繼續開出 open 的未付期別（帳一直累積）；
//   - MarkPastDue 只掃 active → 試用到期後既不催收也不凍結。
//
// 轉 past_due 是**同一張 allowedTransitions** 上的既有轉移（trialing → past_due），之後
// SuspendOverdue 在寬限過後照常凍結、RecordPayment 可把它帶回 active —— 不另立第二套轉移邏輯。
//
// 事件的 payload 自帶 company_id／reason／grace_until（排程不寫平台稽核，見檔頭：事件是唯一的
// 「為什麼」）。冪等由狀態本身保證：轉過去之後就不是 trialing，查詢自然選不中 —— 重跑不重複發事件。
func (b *Billing) ExpireTrials(ctx context.Context, now time.Time, graceDays int) (int, error) {
	n := 0
	// changed 收集本趟**真的轉移**的租戶：失效只能在提交後做（見 invalidate）。
	var changed []int
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		expired, err := b.st.TrialingSubscriptionsExpiredTrial(ctx, tx, now)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		grace := now.AddDate(0, 0, graceDays)
		for _, sub := range expired {
			// 狀態機是第二道閘（查詢條件改壞時仍擋得住非法轉移）。
			if !canTransition(sub.Status, "past_due") {
				continue
			}
			// 未結項 #12：CAS —— 以查詢當下讀到的狀態為預期（見 MarkPastDue 的說明）。
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "past_due", &grace, sub.Status); err != nil {
				if errors.Is(err, store.ErrStatusChanged) {
					continue
				}
				return errcode.SysInternal.Wrap(err)
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.trial_ended", map[string]any{
				"company_id":  sub.CompanyID,
				"grace_until": grace.UTC().Format(time.RFC3339),
				"reason":      "trial_expired",
			}); err != nil {
				return errcode.SysInternal.Wrap(err)
			}
			changed = append(changed, sub.CompanyID)
			n++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	// trialing 與 past_due 都是**可用**狀態，但兩者的寬限／催收語意不同，且 trial_ends_at 仍留在
	// 投影上（租戶卡片據此顯示試用）→ 提交後逐一失效，別讓最長 TTL 內仍讀到「還在試用」。
	for _, companyID := range changed {
		b.invalidate(ctx, companyID)
	}
	return n, nil
}

// MarkPastDue 掃描 active 且期末已過的訂閱 → past_due，並設 grace_until = now + graceDays；
// 回傳實際轉移的筆數。
func (b *Billing) MarkPastDue(ctx context.Context, now time.Time, graceDays int) (int, error) {
	n := 0
	// changed 收集本趟**真的轉移**的租戶：失效只能在提交後做（見 invalidate），故先記下來。
	var changed []int
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		overdue, err := b.st.ActiveSubscriptionsWithDueOpenPeriod(ctx, tx, now)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		grace := now.AddDate(0, 0, graceDays)
		for _, sub := range overdue {
			// 狀態機是這條路徑的第二道閘：查詢條件改壞時，這裡仍擋得住非法轉移。
			if !canTransition(sub.Status, "past_due") {
				continue
			}
			// **第二層防線**（第一層是查詢的 `cur.status = 'open'`）：最新一期已付款或作廢不算逾期。
			// 少了這一層，逾期後才繳清的客戶會被重新催收、寬限期重置，最後被停用凍結，
			// 而 EnsureNextPeriod 只認服務中的訂閱又不會替他開下一期 → 客戶從此停止被開帳（C-1）。
			cur, err := b.st.CurrentPeriodTx(ctx, tx, sub.ID)
			if err != nil {
				return errcode.SysInternal.Wrap(err)
			}
			if cur == nil || cur.Status != "open" {
				continue
			}
			// 未結項 #12：CAS —— 以查詢當下讀到的狀態為預期。同交易內查詢與寫入之間
			// 若有別的寫入者把狀態改走，本次寫入 0 列（ErrStatusChanged）→ 跳過、
			// 不發事件、不計數，而不是把別人的狀態蓋掉。
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "past_due", &grace, sub.Status); err != nil {
				if errors.Is(err, store.ErrStatusChanged) {
					continue
				}
				return errcode.SysInternal.Wrap(err)
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.past_due", map[string]any{
				"company_id":  sub.CompanyID,
				"grace_until": grace.UTC().Format(time.RFC3339),
				"reason":      "period_end_passed_unpaid",
			}); err != nil {
				return errcode.SysInternal.Wrap(err)
			}
			changed = append(changed, sub.CompanyID)
			n++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	for _, companyID := range changed {
		b.invalidate(ctx, companyID)
	}
	return n, nil
}

// SuspendOverdue 掃描 past_due 且 grace_until < now 的訂閱 → suspended（清空寬限期）並發
// subscription.suspended —— 凍結由該事件驅動（consumer 才去改產品域的公司狀態）；
// 回傳實際停用的筆數。
func (b *Billing) SuspendOverdue(ctx context.Context, now time.Time) (int, error) {
	n := 0
	var changed []int // 本趟真的被停用的租戶（提交後逐一失效，見 invalidate）
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		due, err := b.st.PastDueSubscriptionsExpiredGrace(ctx, tx, now)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		for _, sub := range due {
			if !canTransition(sub.Status, "suspended") {
				continue
			}
			// 寬限期在此清空：停用後不該再留著一個已過期的寬限日，
			// 否則下一次掃描與 console 顯示都會說「還在寬限中」。
			// 未結項 #12：CAS —— 以查詢當下讀到的狀態為預期（見 MarkPastDue 的說明）。
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "suspended", nil, sub.Status); err != nil {
				if errors.Is(err, store.ErrStatusChanged) {
					continue
				}
				return errcode.SysInternal.Wrap(err)
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.suspended", map[string]any{
				"company_id": sub.CompanyID,
				"reason":     "overdue",
			}); err != nil {
				return errcode.SysInternal.Wrap(err)
			}
			changed = append(changed, sub.CompanyID)
			n++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	// 停用的意義就是「這個租戶不能再用了」，而判定層看到的是快取裡的舊狀態（usable 仍為 true）
	// → 這是最不能漏失效的一條路徑（漏了就是停用後還照常寫資料，最長 60s）。
	for _, companyID := range changed {
		b.invalidate(ctx, companyID)
	}
	return n, nil
}

// ExpireCancelled 掃描 cancelled 且期末已過的訂閱 → 發 subscription.expired，回傳筆數。
//
// 為什麼需要（spec §5.6 的 G7）：MarkPastDue 只掃 active，已取消的租戶期滿後會一直可用。
// 取消是「期末終止」：期末前仍提供服務，期末後才停止。
//
// **不改變訂閱狀態**（cancelled 是終態，狀態一改，帳與稽核就無法重現）：只發事件，由 consumer
// 把公司轉為 suspended（資料保留、登入被擋）。冪等由查詢的 NOT EXISTS 謂詞保證 ——
// 排程可重跑且不重複發事件。
//
// 因此這裡**不做**權益快取失效：訂閱狀態沒變，判定層的答案就不會變（cancelled 本來就不可用，
// 見 entitlements.usable），真正的狀態變更是 consumer 端把公司凍結那一步（那條路徑自己失效）。
func (b *Billing) ExpireCancelled(ctx context.Context, now time.Time) (int, error) {
	n := 0
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		expired, err := b.st.CancelledSubscriptionsPastPeriodEnd(ctx, tx, now)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		for _, sub := range expired {
			// payload 帶足識別欄位：consumer 不得為了補一個欄位再查一次 DB
			// （事件與查詢之間狀態可能已經變了）。
			if err := b.emit(ctx, tx, sub.ID, "subscription.expired", map[string]any{
				"company_id":      sub.CompanyID,
				"subscription_id": sub.ID,
				"reason":          "cancelled_at_period_end",
			}); err != nil {
				return errcode.SysInternal.Wrap(err)
			}
			n++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

// addBillingPeriod 由期別期末起算下一個期末：月繳取「下月同日」、年繳取「明年同日」；
// 該日不存在（1/31、3/31、閏年的 2/29）時取當月最後一日（G2）。
//
// anchorDay 是**該訂閱第一期起日的日號**（不是 from 的日號）：被 2 月夾擠一次之後，
// 錨點若跟著變成 28，帳單日就永久漂移（1/31 → 2/28 → 3/28 → 4/28…），月繳每期縮成 28 天、
// 一年開出 13 期 —— 月底起租的客戶每年多收一期（I-1）。
//
// 為何不用 time.AddDate：它會正規化（1/31 + 1 月 = 3/3），帳期會跳過整個 2 月 ——
// 客戶被少算一個月的服務，而帳上卻看不出來。未知／空字串週期一律報錯，不得默默當成月繳（G1）。
func addBillingPeriod(from time.Time, anchorDay int, cycle string) (time.Time, error) {
	switch cycle {
	case "monthly":
		return dayOfMonthOrLast(from, anchorDay, from.Year(), int(from.Month())+1), nil
	case "yearly":
		return dayOfMonthOrLast(from, anchorDay, from.Year()+1, int(from.Month())), nil
	default:
		return time.Time{}, fmt.Errorf("未知的計費週期 %q（允許 monthly / yearly）", cycle)
	}
}

// dayOfMonthOrLast 回傳「year-month 的 anchorDay」（時分秒與時區沿用 from）；
// 該日不存在時回該月最後一天。
func dayOfMonthOrLast(from time.Time, anchorDay, year, month int) time.Time {
	if month > 12 {
		year, month = year+1, month-12
	}
	loc := from.Location()
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, loc).Day()
	day := anchorDay
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, time.Month(month), day,
		from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), loc)
}
