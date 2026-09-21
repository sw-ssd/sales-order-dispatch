// Package billing 為平台帳務的狀態機與唯一收款入口：人工收款與（日後）金流 webhook 都呼叫
// RecordPayment，接上金流只需換呼叫者，收款口徑、狀態機與稽核不動。
//
// 平台域與產品域的分工（spec §2.2 規則 3）：本套件**不直接**寫產品域的資料。訂閱凍結／復原
// 是 outbox 事件驅動的（`subscription.suspended`／`reactivated`／`expired` → internal/platform/consumer
// 才去改公司狀態）；本套件的責任是把事件寫進 platform.events，payload 帶足以識別的欄位。
package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// allowedTransitions 為訂閱狀態機（spec §5.2）：不在表上的轉移一律拒絕（PLAT-3001）。收款只走
// 「→ active」，T5 的生命週期排程走其餘的列；表驅動是為了讓「非法轉移」有一個地方可稽核，
// 而不是散在各個呼叫點的 if。
//
// cancelled 沒有出口：解約後不得靠收款悄悄復活 —— 要再服務是「新合約」，不是回到 active。
var allowedTransitions = map[string][]string{
	"trialing":  {"active", "past_due", "suspended", "cancelled"},
	"active":    {"past_due", "suspended", "cancelled"},
	"past_due":  {"active", "suspended", "cancelled"},
	"suspended": {"active", "cancelled"},
	"cancelled": {},
}

// Billing 為平台帳務的狀態機：所有會改變訂閱狀態與期別的寫入都經它（唯一入口）。
// now 可注入：收款時間與測試需要固定時鐘，時間不從參數進來的地方才用它。
type Billing struct {
	st  store.BillingStore
	now func() time.Time
	// cache 為**可選**的權益快取（nil＝未接上）：訂閱狀態改變時失效該租戶的權益快照。
	// 可選的理由：排程／帳務不該因為快取沒接上（本機、CLI、單元測試）而不能跑，
	// 而沒有快取就等於沒有東西要失效。
	cache entitlements.Cache
}

// NewBilling 建立帳務狀態機。st 為平台寫入 store：正式路徑是 admin 連線（postgres.New），
// 單元測試用 store.NewFakeBilling。
func NewBilling(st store.BillingStore) *Billing { return &Billing{st: st, now: time.Now} }

// WithCache 接上權益快取（於組裝時呼叫；nil＝不失效）。
//
// 為什麼不是 NewBilling 的必填參數：這個依賴是**加速器的失效**，不是帳務語意的一部分 ——
// 讓它成為必要參數，等於每個測試、每個 CLI 都得先準備一份快取才能記一筆帳。
func (b *Billing) WithCache(c entitlements.Cache) *Billing {
	b.cache = c
	return b
}

// invalidate 在**交易提交成功之後**失效該租戶的權益快取。
//
// 為什麼一定要在提交後：交易內刪除會在回滾時白刪 —— 資料沒變、快取卻空了（下一次判定還要重建），
// 而「刪了」在 log 上與成功一模一樣。提交後刪才對得上「資料真的變成新的」這個事實。
//
// 為什麼失敗只記 log：快取的錯誤不得讓**已經落地的帳務寫入**回錯誤（錢收了卻回 500，呼叫端會
// 重試，而重試撞上的是冪等路徑 —— 症狀變成「收款成功但介面說失敗」）。代價寫在明處：這條路徑
// 失效失敗時，該租戶最長 TTL（60s）內仍以舊權益放行 —— 這是刻意的界線，不是漏掉的錯誤處理。
func (b *Billing) invalidate(ctx context.Context, companyID int) {
	if err := entitlements.Invalidate(ctx, b.cache, companyID); err != nil {
		log.Printf("billing: 權益快取失效失敗(company=%d): %v（該租戶最長 TTL 內仍讀舊權益）",
			companyID, err)
	}
}

// RecordPaymentInput 為一次收款的輸入。金額一律 int64 分（不得有 float）。
type RecordPaymentInput struct {
	CompanyID int
	// PeriodNo 為 0 時取「當前」期別＝該訂閱最新一期（正常情況下就是那一期 open）；已付款即視為
	// 重送（見 RecordPayment 的冪等）。
	PeriodNo int
	PaidAt   time.Time // 零值 = 現在
	// AmountCents = 0 代表「採用期別快照金額」；非 0 且與快照不符即拒絕（G4：v1 不支援部分付款，
	// 輸入 3000 而期別 1950 不得被默默丟棄 —— 短收／溢收以 Note 記錄，不改變期別金額）。
	AmountCents int64
	Provider    string // manual | ecpay | newebpay | tappay | stripe；空字串視為 manual
	ExternalRef string
	InvoiceNo   string
	// InvoiceStatus／BuyerTaxID／Carrier 為開票資訊（欄位在 platform.subscription_periods）。
	// 與其他付款憑據同一筆寫入（MarkPeriodPaidTx 的同一組 CASE）：開票是付款事件的一部分，
	// 分開寫就允許「收了錢但發票欄位沒落地」。重播（已 paid 且交易號相同）不覆寫它們。
	InvoiceStatus string
	BuyerTaxID    string
	Carrier       string
	// Note 為短收／溢收等人工註記（G8）：付款時寫入，不改變期別金額；空字串保留原值。
	Note string
	// ActorOperatorID 為平台稽核的主體（platform.operators.id，不 FK 租戶 users）。
	ActorOperatorID int64
	// Reason 必填：稽核的 reason（空字串即拒絕）。動錢的操作必須留下「為什麼」。
	Reason string
}

// RecordPayment 為人工收款與（日後）金流 webhook 的共同入口：鎖訂閱 → 標期別已付 → 訂閱轉
// active（清寬限期）→ 寫事件與稽核，四件事都在**同一個交易**內（store.BillingStore.WithTx），
// 任一步失敗即整份回滾 —— 收到款卻沒寫事件、或寫了事件卻沒改期別，都是無法回補的帳務事實。
//
// 冪等：期別已付款且交易號相同時為 no-op（webhook 重送安全，不重複寫事件與稽核）。
func (b *Billing) RecordPayment(ctx context.Context, in RecordPaymentInput) (*store.Period, error) {
	if err := requireActor(in.ActorOperatorID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Reason) == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "reason"})
	}
	if in.AmountCents < 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "amount_cents"})
	}
	if in.Provider == "" {
		in.Provider = "manual"
	}
	if in.PaidAt.IsZero() {
		in.PaidAt = b.now()
	}

	var out *store.Period
	err := b.st.WithTx(ctx, func(tx *sql.Tx) error {
		sub, err := b.st.OpenSubscriptionTx(ctx, tx, in.CompanyID)
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if sub == nil {
			// 沒有合約不得記帳：錢進來了卻沒有可對應的期別，只能人工處理。
			return errcode.PlatformSubscriptionInactive.Error(
				map[string]string{"company_id": strconv.Itoa(in.CompanyID)})
		}
		// 狀態機：收款只把訂閱帶回 active。這道閘必須在期別判定**之前**，否則「已付款期別的重送」
		// 會繞過它 —— 已取消的合約不得因為重送而看起來正常。
		if sub.Status != "active" && !canTransition(sub.Status, "active") {
			return errcode.PlatformSubscriptionInactive.Error(map[string]string{"status": sub.Status})
		}

		period, err := b.periodForPayment(ctx, tx, sub.ID, in.PeriodNo)
		if err != nil {
			return err
		}

		// 冪等（G3）：期別已 paid → 不重複寫事件與稽核。這道判斷**不看交易號是否為空**——人工收款
		// 多數沒有交易號，若要求交易號非空才 no-op，兩邊都空的重送會再寫一次事件與稽核（帳面與
		// 事件流失真）。交易號**不同**（含「已存為空、本次帶號」）才是另一件事：那是重複收款，
		// 在下一段分流處理。
		if period.Status == "paid" {
			// 但交易號不同就不是重送，而是**同一期收到第二筆不同的錢**（重複收款／溢收）：
			// 這是收款衝突，必須讓人看到；靜默 no-op 會讓第二筆匯入在帳面上消失。
			if in.ExternalRef != period.ExternalRef {
				return errcode.PlatformPaymentConflict.Error(map[string]string{
					// 未結項 #30：結構化 kind 鍵 —— console 原依 reason 中文關鍵詞
					// （「已付款」／「不符」）分流兩種語意，脆弱。新舊並存：
					// 有 kind 走 kind，無 kind 回退關鍵詞（舊版 console 相容）。
					"kind": "ref_mismatch",
					"reason": fmt.Sprintf("期別 %d 已付款（交易號 %q），本次交易號 %q 不同；"+
						"請確認是否重複收款，或改用人工對帳處理溢收",
						period.PeriodNo, period.ExternalRef, in.ExternalRef),
				})
			}
			// 未結項 #10:同交易號但金額不同也不是重送 —— webhook 重送帶著錯誤金額重試時，
			// no-op 直接回成功會讓這筆差異在帳面上永遠消失。金額 0 ＝「未填，採快照」，
			// 與收款主路徑的 G4 語意一致，故只在「填了且不符」時衝突。
			if in.AmountCents != 0 && in.AmountCents != period.AmountCents {
				return errcode.PlatformPaymentConflict.Error(map[string]string{
					"kind": "amount_mismatch",
					"reason": fmt.Sprintf("期別 %d 已付款 %s，本次重送金額 %s 不符；"+
						"請確認是否重複收款，或改用人工對帳處理差異",
						period.PeriodNo, money.FormatCents(period.AmountCents),
						money.FormatCents(in.AmountCents)),
				})
			}
			// 交易號相同才是重送：不重寫事件與稽核，但**備註可以補寫** —— note 是短收／溢收的
			// 唯一落點（G8），而 store 對已經 paid 的列只寫 note（其餘憑據由 `status='open'` 的
			// CASE 保留原值）。少了這一段，console 對已入帳期別補記差異會「回成功但什麼都沒寫」。
			if in.Note != "" {
				// 重播只補寫 note:其餘憑據(含開票三欄)由 store 的 `status='open'` CASE 保留原值。
				if err := b.st.MarkPeriodPaidTx(ctx, tx, period.ID, in.PaidAt, in.InvoiceNo,
					in.InvoiceStatus, in.BuyerTaxID, in.Carrier,
					in.Provider, in.ExternalRef, in.Note); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return errcode.SysNotFound.Wrap(err)
					}
					return errcode.SysInternal.Wrap(err)
				}
				period.Note = in.Note
			}
			out = period
			return nil
		}
		// void 等狀態不得入帳：說出真正的狀態（store 也會擋，這裡先給出可讀的錯誤碼與狀態）。
		if period.Status != "open" {
			return errcode.PlatformSubscriptionInactive.Error(
				map[string]string{"period_status": period.Status})
		}

		// 金額驗證（G4）：未填 → 採期別快照；填了但與快照不符 → 拒絕，且說出差異。
		if in.AmountCents != 0 && in.AmountCents != period.AmountCents {
			return errcode.PlatformPaymentConflict.Error(map[string]string{
				"kind": "amount_mismatch",
				"reason": fmt.Sprintf("輸入金額 %s 與期別金額 %s 不符（不支援部分付款；差異請記於備註）",
					money.FormatCents(in.AmountCents), money.FormatCents(period.AmountCents)),
			})
		}

		if err := b.st.MarkPeriodPaidTx(ctx, tx, period.ID, in.PaidAt, in.InvoiceNo,
			in.InvoiceStatus, in.BuyerTaxID, in.Carrier, in.Provider, in.ExternalRef, in.Note); err != nil {
			// 走到這裡還能失敗的只有：期別消失、或同一 provider＋交易號已入帳另一期
			// （00029 的 periods_provider_ref_unique）。狀態問題已在上方擋掉。
			if errors.Is(err, sql.ErrNoRows) {
				return errcode.SysNotFound.Wrap(err)
			}
			return errcode.PlatformPaymentConflict.Wrap(err, map[string]string{"kind": "cross_period"})
		}

		// 入帳成功 → 回傳的期別必須是**新狀態**：period 是 MarkPeriodPaidTx 之前讀出來的列，
		// 帶著它回上層等於回應說「已付款期別仍是 open」（console 顯示未付、operator 再按一次）。
		period.Status = "paid"

		// 復原：清寬限期（補繳後不得再因舊寬限期被停用）。原本已是 active（補繳當期）時不寫事件
		// —— 沒有「從哪裡復原」可言，寫了反而讓 consumer 對已啟用的公司再做一次復原。
		if sub.Status != "active" {
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "active", nil); err != nil {
				return errcode.SysInternal.Wrap(err)
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.reactivated", map[string]any{
				"company_id": in.CompanyID,
				"from":       sub.Status,
			}); err != nil {
				return errcode.SysInternal.Wrap(err)
			}
		}
		if err := b.emit(ctx, tx, sub.ID, "period.payment_recorded", map[string]any{
			"company_id":   in.CompanyID,
			"period_no":    period.PeriodNo,
			"amount_cents": period.AmountCents,
			"provider":     in.Provider,
			"external_ref": in.ExternalRef,
			"paid_at":      in.PaidAt.UTC().Format(time.RFC3339),
		}); err != nil {
			return errcode.SysInternal.Wrap(err)
		}

		after, err := json.Marshal(map[string]any{
			"period_no":    period.PeriodNo,
			"paid_at":      in.PaidAt.UTC().Format(time.RFC3339),
			"provider":     in.Provider,
			"external_ref": in.ExternalRef,
			"amount_cents": period.AmountCents,
			"invoice_no":   in.InvoiceNo,
			// 開票資訊同時落地到期別欄位（見 MarkPeriodPaidTx）與稽核：欄位是帳務的事實，
			// 稽核是「這次入帳帶了什麼」的紀錄，兩者回答的問題不同。
			"invoice_status": in.InvoiceStatus,
			"buyer_tax_id":   in.BuyerTaxID,
			"carrier":        in.Carrier,
		})
		if err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		if err := b.st.RecordAuditTx(ctx, tx, in.ActorOperatorID, "record_payment",
			"subscription", strconv.FormatInt(sub.ID, 10), in.Reason, nil, after); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		out = period
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 收款可能把訂閱由 trialing／past_due／suspended 帶回 active（權益因此改變）→ 提交後失效。
	//
	// 重送（期別已付款、交易號相同的 no-op）也會刪一次：DEL 是冪等的，而且分辨「這次到底改了什麼」
	// 需要在交易內多帶一個旗標出來，換來的只是一次 Valkey 往返 —— 不值得。
	b.invalidate(ctx, in.CompanyID)
	return out, nil
}

// periodForPayment 取本次收款的期別：periodNo > 0 取指定期別（不存在 → SYS-4002），
// periodNo = 0 取該訂閱的「當前」期別（最新一期，狀態不限）。
//
// 為什麼 0 不限定 open：金流 webhook 重送時只帶得回「這一期」而不是狀態 —— 取最新一期才能讓
// 「已付款的重送」走到 no-op（G3）。若反過來要求先找到一期 open，重送會變成「找不到可收款期別」
// 的硬錯誤，而真正的重複收款（交易號不同）也會被誤報成狀態問題。
func (b *Billing) periodForPayment(ctx context.Context, tx *sql.Tx, subID int64, periodNo int) (*store.Period, error) {
	if periodNo > 0 {
		period, err := b.st.OpenPeriodByNoTx(ctx, tx, subID, periodNo)
		if errors.Is(err, sql.ErrNoRows) {
			// 指定的期別不存在（或屬於別的訂閱）：對呼叫端而言就是「這個資源沒有」。
			return nil, errcode.SysNotFound.Wrap(err, map[string]string{"period_no": strconv.Itoa(periodNo)})
		}
		if err != nil {
			return nil, errcode.SysInternal.Wrap(err)
		}
		return period, nil
	}

	period, err := b.st.CurrentPeriodTx(ctx, tx, subID)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	if period == nil {
		return nil, errcode.PlatformSubscriptionInactive.Error(
			map[string]string{"reason": "找不到可收款的期別（請先產生期別）"})
	}
	return period, nil
}

// emit 把 outbox 事件寫在同一交易內。payload 必須帶足以讓 consumer 動手的欄位（例如
// company_id）：consumer 不得為了補一個欄位再查一次 DB —— 事件與查詢之間狀態可能已經變了。
func (b *Billing) emit(ctx context.Context, tx *sql.Tx, subID int64, eventType string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return b.st.EmitEventTx(ctx, tx, "subscription", subID, eventType, raw)
}

// canTransition 回報 from → to 是否為合法轉移。
func canTransition(from, to string) bool {
	return slices.Contains(allowedTransitions[from], to)
}
