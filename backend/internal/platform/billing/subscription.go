// 訂閱生命週期的**營運寫入**(G6):**開通**、席位、改方案、取消 —— 由平台 RPC(operator)呼叫,
// 與收款同一個模式:狀態機、事件與稽核落在**同一個交易**(store.BillingStore.WithTx),
// 任一步失敗即整份回滾。
//
// 為什麼與收款一樣放在本套件而不是服務層直接 UPDATE:改訂閱狀態的入口**只能有一個**。服務層
// 直接寫 subscriptions 就會出現第二份狀態機(以及第二種「哪些狀態可以轉到哪裡」的答案),
// 而兩份實作只會有一份寫對 —— 錯的那份的症狀是「營運改了方案，帳務卻還走舊路徑」。
//
// v1 **不做按日比例計費**:席位與方案一律「下一期生效」,當期期別的金額與快照一個字都不動。
// 理由是對帳:沒有金流對帳的前提下產生半期金額,只會製造對不出來的帳(見 spec 的 G6)。
//
// 這一組方法**不寫產品域的資料**:取消只把平台域的訂閱轉 cancelled,期末之後由 ExpireCancelled
// 發事件、consumer 才去凍結公司(與 SuspendOverdue 同一條路徑)。
package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// CreateSubscriptionInput 為**開通**的輸入:建立訂閱並當場開出第一期。
//
// PlanCode 而不是 plan_id:方案對 operator 是 code(console 的方案清單就是 code),而
// 「已歸檔的方案不得再被指派」由 PlanIDByCodeTx 判定 —— 讓呼叫端先查 id 只會多一個競態。
type CreateSubscriptionInput struct {
	CompanyID    int
	PlanCode     string
	BillingCycle string
	SeatCount    int
	// TrialEnds 為試用到期(必須是未來);nil = 不試用,直接 active。
	TrialEnds       *time.Time
	ActorOperatorID int64
	Reason          string
}

// CreatedSubscription 為開通的結果(訂閱 ＋ 第一期),供呼叫端回應與測試斷言。
type CreatedSubscription struct {
	SubscriptionID int64
	Status         string // trialing | active
	PlanCode       string
	BillingCycle   string
	SeatCount      int
	TrialEnds      *time.Time
	FirstPeriod    *store.Period
}

// CreateSubscription 為**開通**:建立訂閱與第一期,並在同一個交易內寫下事件與平台稽核。
//
// 為什麼第一期一定要在這裡開(B-1):`EnsureNextPeriod` 對「沒有任何期別」的訂閱是 no-op
// (沒有當前期別就不知道起訖與期別號,那不是錯誤而是資料不全)—— 少了第一期,這個租戶**永遠**
// 不會被開帳,收款也無期別可收(`RecordPayment` 回 PLAT-3001)。試用中的租戶同理:第一期照開,
// 否則試用到期時排程不會替他開帳。
//
// 期別長度依 billing_cycle(月 +1 月、年 +1 年),月底由 addBillingPeriod 的錨點規則處理;
// 金額是**當期生效價**的快照(money.PeriodAmount(方案基價, 每席價, 席位數)),沒有價目一律
// 大聲失敗 —— 開出 0 元期別等於免費送方案。
//
// 狀態由 trial_ends_at 決定(有且為未來 → trialing,否則 active):狀態機的 `trialing → active`
// 出口是收款(spec §5.2),開通只負責把合約放進正確的起點。
//
// 重複開通由 00029 的 subscriptions_active_company_unique 擋下(SYS-2001):兩份並行的合約
// 沒有「哪一份生效」的定義。已取消的合約不佔這條唯一鍵 —— 要再服務是**新合約**。
func (b *Billing) CreateSubscription(ctx context.Context, in CreateSubscriptionInput) (*CreatedSubscription, error) {
	planCode := strings.TrimSpace(in.PlanCode)
	if err := requireReason(in.Reason); err != nil {
		return nil, err
	}
	if planCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "plan_code"})
	}
	if in.SeatCount <= 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "seat_count"})
	}
	now := b.now()
	if in.TrialEnds != nil && !in.TrialEnds.After(now) {
		// 已過（或等於現在）的試用期等於「一開通就是過期」:判定層把 trialing 當可用,而排程
		// 不掃試用到期(只掃 active 的逾期),那筆訂閱會永遠停在一個不成立的事實上。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "trial_ends_at"})
	}
	// 期別長度只有 addBillingPeriod 一個真相來源(未知週期在此先擋,不讓它變成一筆開到一半的帳)。
	periodEnd, err := addBillingPeriod(now, now.Day(), in.BillingCycle)
	if err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "billing_cycle"})
	}
	status := "active"
	if in.TrialEnds != nil {
		status = "trialing"
	}

	var out CreatedSubscription
	err = b.st.WithTx(ctx, func(tx *sql.Tx) error {
		planID, err := b.st.PlanIDByCodeTx(ctx, tx, planCode)
		if errors.Is(err, sql.ErrNoRows) {
			// 「不存在」與「已歸檔」同一個碼(與 ChangePlan 一致):對 operator 而言都是
			// 「這個方案不能指派」。
			return errcode.SysNotFound.Wrap(err, map[string]string{"plan_code": planCode})
		}
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		price, err := b.st.CurrentPriceTx(ctx, tx, planID, in.BillingCycle)
		if err != nil {
			return errcode.PlatformSubscriptionInactive.Wrap(err, map[string]string{
				"reason": fmt.Sprintf("方案 %s 沒有 %s 週期的生效價，不得開出 0 元期別",
					planCode, in.BillingCycle),
			})
		}
		amount, err := money.PeriodAmount(price.BaseCents, price.SeatCents, in.SeatCount)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		subID, err := b.st.CreateSubscriptionTx(ctx, tx, store.CreateSubscriptionInput{
			CompanyID: in.CompanyID, PlanID: planID, SeatCount: in.SeatCount,
			BillingCycle: in.BillingCycle, Status: status, TrialEnds: in.TrialEnds,
		})
		if errors.Is(err, store.ErrConflict) {
			// 已註冊的「已存在」碼(SYS-2001):這是 operator 的輸入情境(這家公司已經有合約),
			// 不是 5xx —— console 要顯示得出來,也才擋得住雙擊。
			return errcode.SysConflict.Error(map[string]string{"company_id": strconv.Itoa(in.CompanyID)})
		}
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		first, err := b.st.OpenPeriodTx(ctx, tx, store.OpenPeriodInput{
			SubscriptionID: subID, PeriodNo: 1, PeriodStart: now, PeriodEnd: periodEnd,
			PlanID: planID, UnitPriceCents: price.BaseCents, SeatPriceCents: price.SeatCents,
			SeatCount: in.SeatCount, AmountCents: amount, Currency: price.Currency,
		})
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		out = CreatedSubscription{
			SubscriptionID: subID, Status: status, PlanCode: planCode,
			BillingCycle: in.BillingCycle, SeatCount: in.SeatCount,
			TrialEnds: in.TrialEnds, FirstPeriod: first,
		}
		// 事件的 payload 必須自帶足以動手的欄位(reason 亦在其中):consumer 不得為了補一個欄位
		// 再查一次 DB。`subscription.created` 目前沒有對應的產品域動作(consumer 只認
		// suspended／expired／reactivated)→ 會被認領而無副作用,安全。
		if err := b.emit(ctx, tx, subID, "subscription.created", map[string]any{
			"company_id":      in.CompanyID,
			"subscription_id": subID,
			"reason":          in.Reason,
		}); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		return b.audit(ctx, tx, in.ActorOperatorID, "subscription.create", subID, in.Reason, nil,
			map[string]any{
				"company_id":    in.CompanyID,
				"plan_code":     planCode,
				"billing_cycle": in.BillingCycle,
				"seat_count":    in.SeatCount,
				"status":        status,
				"period_no":     1,
				"amount_cents":  amount,
			})
	})
	if err != nil {
		return nil, err
	}
	// 開通前該公司通常是 status=none(**不施加任何限制**),開通後變成 active＋某方案的全部
	// 權益 —— 這是判定層改變最大的一次寫入,提交後必須失效。
	b.invalidate(ctx, in.CompanyID)
	return &out, nil
}

// SetSeatCountInput 為改席位的輸入。SeatCount 必須 > 0(0 席的訂閱等於沒有使用者,那不是
// 營運調整而是停用,該走取消)。「不得小於目前使用中的席次」由服務層以計數器判定
// (PLAT-5001):那是**業務域的用量**,平台域的狀態機看不到。
type SetSeatCountInput struct {
	CompanyID       int
	SeatCount       int
	ActorOperatorID int64
	Reason          string
}

// SetSeatCount 更新訂閱的席位數(下一次產期即用新席位數計價)。
//
// 已取消的訂閱不得改席位:狀態機對 cancelled 沒有出口,能改席位的話 console 會顯示一個
// 「可以調整、但怎麼調都不會有下一期」的合約 —— 要再服務是**新合約**。
func (b *Billing) SetSeatCount(ctx context.Context, in SetSeatCountInput) (int, error) {
	if err := requireReason(in.Reason); err != nil {
		return 0, err
	}
	if in.SeatCount <= 0 {
		return 0, errcode.SysInvalidArgument.Error(map[string]string{"field": "seat_count"})
	}
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		sub, err := serviceableSubscription(ctx, b.st, tx, in.CompanyID)
		if err != nil {
			return err
		}
		if err := b.st.SetSeatCountTx(ctx, tx, sub.ID, in.SeatCount); err != nil {
			return subscriptionWriteError(err)
		}
		return b.audit(ctx, tx, in.ActorOperatorID, "subscription.set_seats", sub.ID, in.Reason,
			map[string]any{"seat_count": sub.SeatCount},
			map[string]any{"seat_count": in.SeatCount})
	})
	if err != nil {
		return 0, err
	}
	// 席位數進了權益快照(租戶端投影要顯示用量／上限),故與收款一樣在提交後失效。
	b.invalidate(ctx, in.CompanyID)
	return in.SeatCount, nil
}

// ChangePlanInput 為改方案的輸入。
type ChangePlanInput struct {
	CompanyID       int
	PlanCode        string
	ActorOperatorID int64
	Reason          string
}

// ChangePlan 只改訂閱的 plan_id:**當期期別不動**(價格快照已寫死),下一期起用新方案與新價。
//
// 與現行方案相同 → no-op(**不寫稽核**):那不是一次變更。若照樣寫,稽核上會出現「改了某方案」
// 但 before 與 after 一模一樣的紀錄,讓「誰在什麼時候真的改了方案」需要逐筆比對才看得出來。
//
// 回傳「下一期起日」= 當前期的期末(沒有期別時為現在):console 要用它顯示「新方案自 X 生效」。
func (b *Billing) ChangePlan(ctx context.Context, in ChangePlanInput) (time.Time, error) {
	planCode := strings.TrimSpace(in.PlanCode)
	if err := requireReason(in.Reason); err != nil {
		return time.Time{}, err
	}
	var effectiveFrom time.Time
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		planID, err := b.st.PlanIDByCodeTx(ctx, tx, planCode)
		if errors.Is(err, sql.ErrNoRows) {
			// 「不存在」與「已歸檔」同一個碼(SYS-4002):對 operator 而言都是「這個方案不能指派」,
			// 分開只會讓 console 多一個要翻譯的狀態。
			return errcode.SysNotFound.Wrap(err, map[string]string{"plan_code": planCode})
		}
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		sub, err := serviceableSubscription(ctx, b.st, tx, in.CompanyID)
		if err != nil {
			return err
		}
		from, err := b.serviceUntil(ctx, tx, sub, b.now())
		if err != nil {
			return err
		}
		effectiveFrom = from
		if sub.PlanID == planID {
			return nil // 同方案:no-op(不寫稽核)
		}
		if err := b.st.SetSubscriptionPlanTx(ctx, tx, sub.ID, planID); err != nil {
			return subscriptionWriteError(err)
		}
		return b.audit(ctx, tx, in.ActorOperatorID, "subscription.change_plan", sub.ID, in.Reason,
			map[string]any{"plan_id": sub.PlanID},
			map[string]any{"plan_id": planID, "plan_code": planCode})
	})
	if err != nil {
		return time.Time{}, err
	}
	// 方案換了 → 權益(方案的 entitlements)整組可能不同,必須失效該租戶的快取。
	b.invalidate(ctx, in.CompanyID)
	return effectiveFrom, nil
}

// CancelSubscriptionInput 為取消訂閱的輸入。
type CancelSubscriptionInput struct {
	CompanyID       int
	AtPeriodEnd     bool
	ActorOperatorID int64
	Reason          string
}

// Cancellation 為取消的結果。CancelledAt 為零值代表這次是 **no-op**(訂閱本來就是 cancelled)。
type Cancellation struct {
	CancelledAt  time.Time
	ServiceUntil time.Time
}

// CancelSubscription 把訂閱轉為 cancelled(期末終止):期末前仍提供服務,期末後由
// ExpireCancelled 發 subscription.expired、consumer 才把公司凍結(資料保留、登入被擋)。
//
// 兩條界線:
//   - **at_period_end=false 一律拒絕**(PLAT-3001):v1 沒有按日比例計費,立即終止會產生一筆
//     「已收但不再服務」的期別 —— 那是退款流程,不是取消流程。
//   - 已是 cancelled → no-op(不寫稽核、不重發事件):重複取消不得推進 cancelled_at,也不得
//     再發一次 subscription.expired(那會讓 consumer 重複處理同一件事)。
func (b *Billing) CancelSubscription(ctx context.Context, in CancelSubscriptionInput) (Cancellation, error) {
	if err := requireReason(in.Reason); err != nil {
		return Cancellation{}, err
	}
	if !in.AtPeriodEnd {
		return Cancellation{}, errcode.PlatformSubscriptionInactive.Error(map[string]string{
			"reason": "v1 僅支援期末終止（按日比例計費與退款不在本版本範圍）",
		})
	}
	var out Cancellation
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		sub, err := b.st.OpenSubscriptionTx(ctx, tx, in.CompanyID)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if sub == nil {
			return errcode.PlatformSubscriptionInactive.Error(
				map[string]string{"company_id": strconv.Itoa(in.CompanyID)})
		}
		if out.ServiceUntil, err = b.serviceUntil(ctx, tx, sub, b.now()); err != nil {
			return err
		}
		if sub.Status == "cancelled" {
			return nil // no-op
		}
		if !canTransition(sub.Status, "cancelled") {
			return errcode.PlatformSubscriptionInactive.Error(map[string]string{"status": sub.Status})
		}
		if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "cancelled", nil); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		out.CancelledAt = b.now()
		// 事件的 payload 必須自帶足以動手的欄位(reason 亦在其中):consumer 不得為了補一個欄位
		// 再查一次 DB —— 事件與查詢之間狀態可能已經變了。
		if err := b.emit(ctx, tx, sub.ID, "subscription.cancelled", map[string]any{
			"company_id":    in.CompanyID,
			"from":          sub.Status,
			"service_until": out.ServiceUntil.UTC().Format(time.RFC3339),
			"reason":        in.Reason,
		}); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		return b.audit(ctx, tx, in.ActorOperatorID, "subscription.cancel", sub.ID, in.Reason,
			map[string]any{"status": sub.Status},
			map[string]any{
				"status":        "cancelled",
				"service_until": out.ServiceUntil.UTC().Format(time.RFC3339),
			})
	})
	if err != nil {
		return Cancellation{}, err
	}
	// cancelled 在判定層是**不可用**狀態(entitlements.usable 的 allow-list),故提交後必須失效:
	// 少了這一步,客戶在解約後最長 TTL(60s)內仍以原方案的權益放行。
	b.invalidate(ctx, in.CompanyID)
	return out, nil
}

// serviceUntil 回傳「服務提供到什麼時候」= 當前期別的期末(沒有期別時為 now)。
// 取消與改方案共用:兩者都是「下一期生效」,生效點就是同一個時間。
func (b *Billing) serviceUntil(ctx context.Context, tx *sql.Tx, sub *store.Subscription, now time.Time) (time.Time, error) {
	cur, err := b.st.CurrentPeriodTx(ctx, tx, sub.ID)
	if err != nil {
		return time.Time{}, errcode.SysInternal.Wrap(err)
	}
	if cur == nil {
		return now, nil
	}
	return cur.PeriodEnd, nil
}

// serviceableSubscription 取現行訂閱並擋掉不可變更的狀態(沒有合約、已取消)。
//
// 先 FOR UPDATE 鎖列(OpenSubscriptionTx)再判定:不鎖的話兩個併發的營運操作(改席位與取消)
// 會各自讀到舊狀態,一個寫進 cancelled 的合約、一個把它又改回 active。
func serviceableSubscription(ctx context.Context, st store.BillingStore, tx *sql.Tx,
	companyID int) (*store.Subscription, error) {
	sub, err := st.OpenSubscriptionTx(ctx, tx, companyID)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	if sub == nil {
		return nil, errcode.PlatformSubscriptionInactive.Error(
			map[string]string{"company_id": strconv.Itoa(companyID)})
	}
	if sub.Status == "cancelled" {
		// cancelled 是終態:要再服務是「新合約」,不是在舊合約上調整(與收款不得使它復活同一立場)。
		return nil, errcode.PlatformSubscriptionInactive.Error(map[string]string{"status": sub.Status})
	}
	return sub, nil
}

// audit 寫入平台稽核(與資料同一個交易)。before／after 為 nil 的欄位序列化後照樣帶上鍵:
// 稽核要能回答「從什麼變成什麼」,少一個鍵就得回頭讀程式碼才知道當時有沒有這欄。
func (b *Billing) audit(ctx context.Context, tx *sql.Tx, operatorID int64, action string,
	subID int64, reason string, before, after map[string]any) error {
	rawBefore, err := json.Marshal(before)
	if err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	rawAfter, err := json.Marshal(after)
	if err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	if err := b.st.RecordAuditTx(ctx, tx, operatorID, action, "subscription",
		strconv.FormatInt(subID, 10), reason, rawBefore, rawAfter); err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	return nil
}

// requireReason 是所有平台寫入的共同必填檢查(空字串即拒絕):沒有原因的調整在事後無法複查 ——
// 「為什麼這個租戶的席位突然變成 100」是稽核唯一能回答的問題。
func requireReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errcode.SysInvalidArgument.Error(map[string]string{"field": "reason"})
	}
	return nil
}

// subscriptionWriteError 把訂閱寫入的錯誤收斂:列不存在 → SYS-4002(資源消失),其餘 → SYS-9000。
//
// **不得**映射成 PLAT-3002(收款衝突):那一碼描述的是帳務衝突(重複收款、金額不符),而這裡
// 的失敗可能是死鎖、序化失敗或連線中斷 —— 講成「收款衝突」會讓 operator 以為重試沒用而放棄。
func subscriptionWriteError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errcode.SysNotFound.Wrap(err)
	}
	return errcode.SysInternal.Wrap(err)
}
