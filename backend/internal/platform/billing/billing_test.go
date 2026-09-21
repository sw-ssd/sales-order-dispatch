// Task 4：RecordPayment 是收款的**唯一入口**（人工收款與日後金流 webhook 共用）的單元契約。
//
// 用 store.NewFakeBilling（Task 3 的記憶體 BillingStore：語意與真 store 對齊，含 WithTx 的
// 整份快照還原）驗四件事：
//
//	① 狀態機逐格：合法轉移收款成功，非法轉移（cancelled／未列舉狀態）一律 PLAT-3001；
//	② 期別與金額：金額不符 PLAT-3002、已付款同交易號重送 no-op、同交易號不同則為衝突；
//	③ 事件與稽核：復原事件只在原本非 active 時發、payload 帶 company_id、reason 必填；
//	④ 同一交易：後段寫入失敗即整份還原，不留「期別已 paid 卻沒事件與稽核」的半成品。
//
// 假實作沒有真的交易與列鎖，那條界線由 billing_integration_test.go（真 PostgreSQL）把關。
package billing_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// amountCents 為收款測試的固定期別金額：1950.00 元（月費 1500 ＋ 150 × 3 席），
// 與整合測試的 fixture 同一組數字。
const amountCents = int64(195000)

// errorInfoOf 由 connect error 取 ErrorInfo（碼與 details 到得了客戶端才有意義）。
func errorInfoOf(t *testing.T, err error) *commonv1.ErrorInfo {
	t.Helper()
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("非 connect error: %v", err)
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			t.Fatalf("detail 取值失敗: %v", derr)
		}
		if info, ok := v.(*commonv1.ErrorInfo); ok {
			return info
		}
	}
	t.Fatalf("錯誤未帶 ErrorInfo（碼到不了客戶端）: %v", err)
	return nil
}

// errorCodeOf 取對外錯誤碼（例 PLAT-3002）；裸 connect 錯誤一律紅。
func errorCodeOf(t *testing.T, err error) string { return errorInfoOf(t, err).GetCode() }

// eventTypes 回全部事件的型別（依寫入順序）。
func eventTypes(f *store.FakeBilling) []string {
	var out []string
	for _, e := range f.Events() {
		out = append(out, e.EventType)
	}
	return out
}

// eventPayload 把指定型別事件的 payload 解進 out —— consumer 就是這樣讀它的，
// 所以「事件帶得出識別欄位」是這裡要釘住的契約。
func eventPayload(t *testing.T, f *store.FakeBilling, eventType string, out any) {
	t.Helper()
	for _, e := range f.Events() {
		if e.EventType != eventType {
			continue
		}
		if err := json.Unmarshal(e.Payload, out); err != nil {
			t.Fatalf("事件 %s 的 payload 無法解析: %v (%s)", eventType, err, e.Payload)
		}
		return
	}
	t.Fatalf("找不到事件 %s，全部事件: %v", eventType, eventTypes(f))
}

// readSub 讀回訂閱現況（OpenSubscriptionTx 是介面上唯一的讀法，假實作忽略 tx）。
func readSub(t *testing.T, f *store.FakeBilling, companyID int) *store.Subscription {
	t.Helper()
	sub, err := f.OpenSubscriptionTx(context.Background(), nil, companyID)
	if err != nil {
		t.Fatalf("讀訂閱: %v", err)
	}
	if sub == nil {
		t.Fatalf("公司 %d 應有訂閱", companyID)
	}
	return sub
}

// readPeriod 讀回期別現況。
func readPeriod(t *testing.T, f *store.FakeBilling, subID int64, periodNo int) *store.Period {
	t.Helper()
	p, err := f.OpenPeriodByNoTx(context.Background(), nil, subID, periodNo)
	if err != nil {
		t.Fatalf("讀期別: %v", err)
	}
	return p
}

// 收款把 suspended 轉回 active、清寬限期，並把期別、事件與稽核全部寫下（同一交易）。
func TestRecordPaymentReactivatesSuspendedSubscription(t *testing.T) {
	grace := time.Now().Add(48 * time.Hour)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, PlanCode: "std",
		Status: "suspended", SeatCount: 3, GraceUntil: &grace})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})
	paidAt := time.Date(2026, 9, 21, 10, 30, 0, 0, time.UTC)

	got, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: paidAt, Provider: "manual", ExternalRef: "BANK-12345",
		InvoiceNo: "AB12345678", InvoiceStatus: "issued", BuyerTaxID: "12345678", Carrier: "/ABC1234",
		Note: "短收 50 元，已於 6/1 補足", ActorOperatorID: 7, Reason: "匯款入帳（台銀 12345）",
	})
	if err != nil {
		t.Fatalf("RecordPayment: %v", err)
	}
	if got == nil || got.ID != 9 {
		t.Fatalf("應回被標記的期別，got %+v", got)
	}

	// ① 期別：paid、付款憑據與人工註記落地（G8）。
	p := readPeriod(t, f, 5, 1)
	if p.Status != "paid" {
		t.Fatalf("期別應轉 paid，got %q", p.Status)
	}
	if p.PaidAt == nil || !p.PaidAt.Equal(paidAt) {
		t.Fatalf("paid_at 應為 %s，got %v", paidAt, p.PaidAt)
	}
	if p.InvoiceNo != "AB12345678" || p.PaymentProvider != "manual" || p.ExternalRef != "BANK-12345" {
		t.Fatalf("付款憑據未落地: %+v", p)
	}
	if p.Note != "短收 50 元，已於 6/1 補足" {
		t.Fatalf("人工註記未落地，got %q", p.Note)
	}

	// ② 訂閱：回 active 且清空寬限期（否則舊寬限期會再把人停用一次）。
	sub := readSub(t, f, 42)
	if sub.Status != "active" {
		t.Fatalf("訂閱應轉 active，got %q", sub.Status)
	}
	if sub.GraceUntil != nil {
		t.Fatalf("復原後應清空寬限期，got %v", sub.GraceUntil)
	}

	// ③ 事件：復原與收款各一筆，payload 帶得出下游要用的識別欄位。
	if evs := eventTypes(f); len(evs) != 2 {
		t.Fatalf("應恰好寫 2 筆事件，got %v", evs)
	}
	var reactivated struct {
		CompanyID int    `json:"company_id"`
		From      string `json:"from"`
	}
	eventPayload(t, f, "subscription.reactivated", &reactivated)
	if reactivated.CompanyID != 42 || reactivated.From != "suspended" {
		t.Fatalf("復原事件應帶 company_id=42／from=suspended，got %+v", reactivated)
	}
	var recorded struct {
		CompanyID   int    `json:"company_id"`
		PeriodNo    int    `json:"period_no"`
		AmountCents int64  `json:"amount_cents"`
		Provider    string `json:"provider"`
		ExternalRef string `json:"external_ref"`
	}
	eventPayload(t, f, "period.payment_recorded", &recorded)
	if recorded.CompanyID != 42 || recorded.PeriodNo != 1 || recorded.AmountCents != amountCents ||
		recorded.Provider != "manual" || recorded.ExternalRef != "BANK-12345" {
		t.Fatalf("收款事件的識別欄位不完整: %+v", recorded)
	}

	// ④ 稽核：一筆、actor 為 operator、reason 為人工理由、after 記下這次收款。
	audits := f.Audits()
	if len(audits) != 1 {
		t.Fatalf("應寫 1 筆平台稽核，got %v", audits)
	}
	a := audits[0]
	if a.Action != "record_payment" || a.TargetType != "subscription" || a.TargetID != "5" {
		t.Fatalf("稽核目標應為 record_payment/subscription/5，got %+v", a)
	}
	if a.OperatorID != 7 || a.Reason != "匯款入帳（台銀 12345）" {
		t.Fatalf("稽核主體與原因不對: %+v", a)
	}
	var after struct {
		PeriodNo    int    `json:"period_no"`
		AmountCents int64  `json:"amount_cents"`
		Provider    string `json:"provider"`
		ExternalRef string `json:"external_ref"`
		BuyerTaxID  string `json:"buyer_tax_id"`
	}
	if err := json.Unmarshal(a.After, &after); err != nil {
		t.Fatalf("稽核 after 不是 JSON: %v (%s)", err, a.After)
	}
	if after.PeriodNo != 1 || after.AmountCents != amountCents ||
		after.Provider != "manual" || after.ExternalRef != "BANK-12345" {
		t.Fatalf("稽核 after 未記下收款內容: %+v", after)
	}
	if after.BuyerTaxID != "12345678" {
		t.Fatalf("營運輸入的開票資訊不得靜默丟棄，稽核 after 應留 buyer_tax_id，got %+v", after)
	}
}

// 狀態機逐格：收款只把訂閱帶回 active，且只有「原本不是 active」才發復原事件；
// cancelled 是終態（解約後不得被收款悄悄復活），未列舉的狀態一律拒絕。
func TestRecordPaymentStateMachine(t *testing.T) {
	cases := []struct {
		name            string
		status          string
		wantCode        string // "" = 收款成功
		wantReactivated bool
	}{
		{"試用中轉正式", "trialing", "", true},
		{"已是 active 補繳當期", "active", "", false},
		{"逾期補繳", "past_due", "", true},
		{"停用後補繳復原", "suspended", "", true},
		{"已取消為終態", "cancelled", "PLAT-3001", false},
		{"未列舉的狀態不放行", "expired", "PLAT-3001", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := store.NewFakeBilling()
			f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: tc.status})
			f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

			got, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
				CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "補繳",
			})

			if tc.wantCode != "" {
				if got != nil {
					t.Fatalf("拒絕時不得回期別，got %+v", got)
				}
				if code := errorCodeOf(t, err); code != tc.wantCode {
					t.Fatalf("ErrorInfo.code = %q；want %q（非法轉移一律 FailedPrecondition）", code, tc.wantCode)
				}
				if connect.CodeOf(err) != connect.CodeFailedPrecondition {
					t.Fatalf("connect 碼應為 failed_precondition，got %v", connect.CodeOf(err))
				}
				// 拒絕不得留下任何一筆寫入（半成品比錯誤更難補救）。
				if evs := eventTypes(f); len(evs) != 0 {
					t.Fatalf("拒絕不得寫事件，got %v", evs)
				}
				if audits := f.Audits(); len(audits) != 0 {
					t.Fatalf("拒絕不得寫稽核，got %v", audits)
				}
				if p := readPeriod(t, f, 5, 1); p.Status != "open" || p.PaidAt != nil {
					t.Fatalf("拒絕後期別應仍是 open 且未付款，got %+v", p)
				}
				if sub := readSub(t, f, 42); sub.Status != tc.status {
					t.Fatalf("拒絕後訂閱狀態不得改變，got %q", sub.Status)
				}
				return
			}

			if err != nil {
				t.Fatalf("RecordPayment: %v", err)
			}
			if sub := readSub(t, f, 42); sub.Status != "active" {
				t.Fatalf("訂閱應轉 active，got %q", sub.Status)
			}
			evs := eventTypes(f)
			if got := slices.Contains(evs, "subscription.reactivated"); got != tc.wantReactivated {
				t.Fatalf("訂閱原本為 %q → subscription.reactivated=%v；want %v（全部：%v）",
					tc.status, got, tc.wantReactivated, evs)
			}
			if !slices.Contains(evs, "period.payment_recorded") {
				t.Fatalf("收款必寫 period.payment_recorded，got %v", evs)
			}
			if len(f.Audits()) != 1 {
				t.Fatalf("收款必寫 1 筆稽核，got %v", f.Audits())
			}
		})
	}
}

// 缺原因即拒絕：平台稽核必填，而且這是人工動錢的操作（沒有理由的入帳無法對帳）。
func TestRecordPaymentRequiresReason(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "  ",
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺原因應回 invalid_argument，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "SYS-1001" {
		t.Fatalf("ErrorInfo.code = %q；want SYS-1001", code)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("拒絕不得寫事件，got %v", evs)
	}
	if p := readPeriod(t, f, 5, 1); p.Status != "open" {
		t.Fatalf("拒絕後期別不得變動，got %+v", p)
	}
}

// 負數金額即拒絕：金額一律 int64 分且不得為負（不讓它混進「與快照不符」的收款衝突裡，
// 那會把「打錯符號」講成「金額不符」）。
func TestRecordPaymentRejectsNegativeAmount(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", AmountCents: -amountCents,
		ActorOperatorID: 7, Reason: "打錯符號",
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("負數金額應回 invalid_argument，got %v", err)
	}
	if p := readPeriod(t, f, 5, 1); p.Status != "open" {
		t.Fatalf("拒絕後期別不得變動，got %+v", p)
	}
}

// 無訂閱即拒絕（不得為不存在的合約記帳）：錢進來了卻對不到合約，只能人工處理。
func TestRecordPaymentWithoutSubscription(t *testing.T) {
	f := store.NewFakeBilling()

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 4242, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("無訂閱應回 failed_precondition，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "PLAT-3001" {
		t.Fatalf("ErrorInfo.code = %q；want PLAT-3001", code)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("拒絕不得寫事件，got %v", evs)
	}
}

// 輸入金額與期別快照不符即衝突（G4）：v1 不支援部分付款，輸入 3000 而期別 1950 不得被
// 靜默丟棄（帳面仍會是 1950，等於輸入被吃掉）；短收／溢收一律以 note 記錄。
func TestRecordPaymentMismatchedAmountIsConflict(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", AmountCents: 300000,
		ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("金額不符應回 failed_precondition，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "PLAT-3002" {
		t.Fatalf("ErrorInfo.code = %q；want PLAT-3002", code)
	}
	if details := errorInfoOf(t, err).GetDetails(); details["reason"] == "" {
		t.Fatalf("收款衝突必須說出原因，got %v", details)
	}
	if p := readPeriod(t, f, 5, 1); p.Status != "open" {
		t.Fatalf("衝突後期別不得變動，got %+v", p)
	}
	if sub := readSub(t, f, 42); sub.Status != "suspended" {
		t.Fatalf("衝突後不得改訂閱狀態，got %q", sub.Status)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("衝突不得寫事件，got %v", evs)
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("衝突不得寫稽核，got %v", audits)
	}
}

// 金額 0 代表「採期別快照金額」：事件與稽核都要記快照金額，不得記成 0。
func TestRecordPaymentZeroAmountUsesSnapshot(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

	if _, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "匯款入帳",
	}); err != nil {
		t.Fatalf("RecordPayment: %v", err)
	}
	var recorded struct {
		AmountCents int64 `json:"amount_cents"`
	}
	eventPayload(t, f, "period.payment_recorded", &recorded)
	if recorded.AmountCents != amountCents {
		t.Fatalf("收款事件的金額應為期別快照 %d，got %d", amountCents, recorded.AmountCents)
	}
	if p := readPeriod(t, f, 5, 1); p.AmountCents != amountCents {
		t.Fatalf("期別金額不得被輸入改寫，got %d", p.AmountCents)
	}
}

// 指定期別收款要收到正確的那一期；PeriodNo = 0 取「當前 open 期別」（最新的一期 open）。
func TestRecordPaymentSelectsPeriod(t *testing.T) {
	cases := []struct {
		name     string
		periodNo int
		want     int // 應被標記的 period_no
	}{
		{"指定期別", 2, 2},
		{"未指定取最新一期", 0, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := store.NewFakeBilling()
			f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
			f.PutPeriod(store.Period{ID: 8, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: 10000})
			f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 2, Status: "open", AmountCents: amountCents})

			got, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
				CompanyID: 42, PeriodNo: tc.periodNo, PaidAt: time.Now(), Provider: "manual",
				ActorOperatorID: 7, Reason: "匯款入帳",
			})
			if err != nil {
				t.Fatalf("RecordPayment: %v", err)
			}
			if got == nil || got.PeriodNo != tc.want {
				t.Fatalf("應收回 period_no=%d，got %+v", tc.want, got)
			}
			if p := readPeriod(t, f, 5, tc.want); p.Status != "paid" {
				t.Fatalf("期別 %d 應轉 paid，got %q", tc.want, p.Status)
			}
			if p := readPeriod(t, f, 5, 1); p.Status != "open" {
				t.Fatalf("未選中的期別不得被動到，got %+v", p)
			}
		})
	}
}

// 指定的期別不存在即拒絕：呼叫端（console／webhook）拿到 not_found 才能說「這期別沒有」，
// 而不是回一個模糊的狀態錯誤。
func TestRecordPaymentUnknownPeriod(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PeriodNo: 9, PaidAt: time.Now(), Provider: "manual",
		ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("期別不存在應回 not_found，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "SYS-4002" {
		t.Fatalf("ErrorInfo.code = %q；want SYS-4002", code)
	}
	if p := readPeriod(t, f, 5, 1); p.Status != "open" {
		t.Fatalf("拒絕後期別不得變動，got %+v", p)
	}
}

// 完全沒有期別即拒絕（先產生期別再記帳）：否則會把錢記到一期不存在的帳上。
func TestRecordPaymentWithoutPeriod(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("沒有期別應回 failed_precondition，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "PLAT-3001" {
		t.Fatalf("ErrorInfo.code = %q；want PLAT-3001", code)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("拒絕不得寫事件，got %v", evs)
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("拒絕不得寫稽核，got %v", audits)
	}
}

// 當前（最新）期別已付款且交易號相同 → 重送（G3）：即使已經沒有 open 期別也不得報錯，
// 且不得再寫事件與稽核 —— 金流 webhook 重送與人工重按都走這條。
func TestRecordPaymentReplayOfLatestPaidPeriod(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "paid", AmountCents: amountCents})

	got, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if err != nil {
		t.Fatalf("重送應為 no-op，got %v", err)
	}
	if got == nil || got.ID != 9 {
		t.Fatalf("重送應回當前那一期，got %+v", got)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("重送不得寫事件，got %v", evs)
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("重送不得寫稽核，got %v", audits)
	}
}

// void（作廢）的期別不得入帳，且錯誤要說出真正的狀態（不得講成「已付款」）。
func TestRecordPaymentRejectsVoidPeriod(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "void", AmountCents: amountCents})

	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PeriodNo: 1, PaidAt: time.Now(), Provider: "manual",
		ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("作廢期別應回 failed_precondition，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "PLAT-3001" {
		t.Fatalf("ErrorInfo.code = %q；want PLAT-3001", code)
	}
	if p := readPeriod(t, f, 5, 1); p.Status != "void" {
		t.Fatalf("作廢期別不得被改動，got %+v", p)
	}
	if sub := readSub(t, f, 42); sub.Status != "suspended" {
		t.Fatalf("訂閱不得被復原，got %q", sub.Status)
	}
}

// 重送（同期別、同交易號）是 no-op（G3）：金流 webhook 重送靠這條，不得再寫一次事件與稽核，
// 也不得動已落地的憑據。
// 未結項 #10:no-op 分支不比對金額 —— 同交易號、金額不同的第二筆匯款會被靜默吸收。
// webhook 重送帶著錯誤金額重試時，no-op 直接回成功，帳面上永遠看不到這筆差異。
func TestRecordPaymentReplayComparesAmount(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})
	b := billing.NewBilling(f)
	paidAt := time.Date(2026, 9, 21, 10, 30, 0, 0, time.UTC)
	in := billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: paidAt, Provider: "manual", ExternalRef: "BANK-12345",
		ActorOperatorID: 7, Reason: "匯款入帳",
	}
	if _, err := b.RecordPayment(context.Background(), in); err != nil {
		t.Fatalf("第一次收款: %v", err)
	}
	// 同交易號、金額不同 → 必須是收款衝突(PLAT-3002)，不得靜默 no-op。
	in.AmountCents = amountCents - 5000
	_, err := b.RecordPayment(context.Background(), in)
	if err == nil {
		t.Fatal("同交易號但金額不同的重送不得靜默成功")
	}
	if got := errorCodeOf(t, err); got != "PLAT-3002" {
		t.Fatalf("必須是收款衝突 PLAT-3002,got %q (%v)", got, err)
	}
	if got := len(f.Events()); got != 1 {
		t.Fatalf("衝突不得再寫事件，got %d 筆", got)
	}
	if got := len(f.Audits()); got != 1 {
		t.Fatalf("衝突不得再寫稽核，got %d 筆", got)
	}
}
func TestRecordPaymentReplayIsNoop(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})
	b := billing.NewBilling(f)
	paidAt := time.Date(2026, 9, 21, 10, 30, 0, 0, time.UTC)
	in := billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: paidAt, Provider: "manual", ExternalRef: "BANK-12345",
		InvoiceNo: "AB12345678", ActorOperatorID: 7, Reason: "匯款入帳（台銀 12345）",
	}
	if _, err := b.RecordPayment(context.Background(), in); err != nil {
		t.Fatalf("第一次收款: %v", err)
	}

	got, err := b.RecordPayment(context.Background(), in)
	if err != nil {
		t.Fatalf("重送應為 no-op，got %v", err)
	}
	if got == nil || got.ID != 9 {
		t.Fatalf("重送應回同一期別，got %+v", got)
	}
	if evs := eventTypes(f); len(evs) != 2 {
		t.Fatalf("重送不得再寫事件，got %v", evs)
	}
	if audits := f.Audits(); len(audits) != 1 {
		t.Fatalf("重送不得再寫稽核，got %v", audits)
	}
	p := readPeriod(t, f, 5, 1)
	if p.PaidAt == nil || !p.PaidAt.Equal(paidAt) || p.InvoiceNo != "AB12345678" || p.ExternalRef != "BANK-12345" {
		t.Fatalf("重送不得改動已落地的憑據，got %+v", p)
	}
}

// 重播（同期別、同交易號）可以**補寫**短收／溢收的備註，但事件與稽核仍不得重寫 ——
// store 的契約是「note 入帳與重播都可補寫、空字串保留原值」，而 note 是短收／溢收的唯一落點
// （G8）；少了這條，console 對已入帳期別補記差異會回成功但什麼都沒寫。
func TestRecordPaymentReplayPatchesNote(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})
	b := billing.NewBilling(f)
	paidAt := time.Date(2026, 9, 21, 10, 30, 0, 0, time.UTC)
	in := billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: paidAt, Provider: "manual", ExternalRef: "BANK-12345",
		InvoiceNo: "AB12345678", Note: "匯入 1900（短收 50）", ActorOperatorID: 7, Reason: "匯款入帳",
	}
	if _, err := b.RecordPayment(context.Background(), in); err != nil {
		t.Fatalf("第一次收款: %v", err)
	}
	eventsBefore, auditsBefore := len(f.Events()), len(f.Audits())

	// 空備註的重播：保留原註記（不得被空字串清掉）。
	in.Note = ""
	if _, err := b.RecordPayment(context.Background(), in); err != nil {
		t.Fatalf("空備註重播: %v", err)
	}
	if got := readPeriod(t, f, 5, 1).Note; got != "匯入 1900（短收 50）" {
		t.Fatalf("空備註不得清掉原註記，got %q", got)
	}

	// 非空備註的重播：補寫差異（短收／溢收的唯一落點，G8）。
	in.Note = "6/1 已補足差額 50 元"
	if _, err := b.RecordPayment(context.Background(), in); err != nil {
		t.Fatalf("補寫備註: %v", err)
	}
	p := readPeriod(t, f, 5, 1)
	if p.Note != "6/1 已補足差額 50 元" {
		t.Fatalf("重播應可補寫備註，got %q", p.Note)
	}
	if p.Status != "paid" || p.InvoiceNo != "AB12345678" || p.ExternalRef != "BANK-12345" ||
		p.PaidAt == nil || !p.PaidAt.Equal(paidAt) {
		t.Fatalf("補寫備註不得動到付款憑據，got %+v", p)
	}
	if got := len(f.Events()); got != eventsBefore {
		t.Fatalf("重播不得再寫事件，got %d 筆（原 %d 筆）", got, eventsBefore)
	}
	if got := len(f.Audits()); got != auditsBefore {
		t.Fatalf("重播不得再寫稽核，got %d 筆（原 %d 筆）", got, auditsBefore)
	}
}

// 期別已付款但交易號不同 = 同一期收到第二筆不同的錢（溢收／重複收款）：這是收款衝突，
// 必須回 PLAT-3002 讓人看到；靜默 no-op 會讓第二筆匯入在帳面上消失。
func TestRecordPaymentConflictsOnDifferentExternalRef(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})
	b := billing.NewBilling(f)
	if _, err := b.RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ExternalRef: "BANK-12345",
		ActorOperatorID: 7, Reason: "匯款入帳（台銀 12345）",
	}); err != nil {
		t.Fatalf("第一次收款: %v", err)
	}

	eventsBefore, auditsBefore := len(f.Events()), len(f.Audits())
	_, err := b.RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ExternalRef: "BANK-99999",
		ActorOperatorID: 7, Reason: "第二筆匯款",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("交易號不同應回 failed_precondition，got %v", err)
	}
	if code := errorCodeOf(t, err); code != "PLAT-3002" {
		t.Fatalf("ErrorInfo.code = %q；want PLAT-3002", code)
	}
	if got := len(f.Events()); got != eventsBefore {
		t.Fatalf("衝突不得寫事件，got %d 筆（原 %d 筆）", got, eventsBefore)
	}
	if got := len(f.Audits()); got != auditsBefore {
		t.Fatalf("衝突不得寫稽核，got %d 筆（原 %d 筆）", got, auditsBefore)
	}
	if p := readPeriod(t, f, 5, 1); p.ExternalRef != "BANK-12345" {
		t.Fatalf("已入帳的交易號不得被覆蓋，got %q", p.ExternalRef)
	}
}

// failEvents 讓事件寫入失敗：WithTx 的整份還原（＝真 PG 的 rollback）就是「同一交易」的驗收點
// —— 期別已 paid、訂閱已 active 卻沒有事件與稽核，帳就對不起來。
type failEvents struct{ *store.FakeBilling }

func (failEvents) EmitEventTx(context.Context, *sql.Tx, string, int64, string, []byte) error {
	return errors.New("模擬事件寫入失敗")
}

func TestRecordPaymentRollsBackOnWriteFailure(t *testing.T) {
	grace := time.Now().Add(48 * time.Hour)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended", GraceUntil: &grace})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: amountCents})

	_, err := billing.NewBilling(failEvents{f}).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 7, Reason: "匯款入帳",
	})
	if err == nil {
		t.Fatal("事件寫入失敗時不得回成功")
	}
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("寫入失敗應回 internal（基礎設施問題不是客戶端問題），got %v", err)
	}
	if code := errorCodeOf(t, err); code != "SYS-9000" {
		t.Fatalf("ErrorInfo.code = %q；want SYS-9000", code)
	}
	if p := readPeriod(t, f, 5, 1); p.Status != "open" || p.PaidAt != nil {
		t.Fatalf("回滾後期別應仍是 open，got %+v", p)
	}
	sub := readSub(t, f, 42)
	if sub.Status != "suspended" || sub.GraceUntil == nil {
		t.Fatalf("回滾後訂閱應仍是 suspended 且保留寬限期，got %+v", sub)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("回滾後不得留下事件，got %v", evs)
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("回滾後不得留下稽核，got %v", audits)
	}
}
