//go:build integration

// Task 4 的整合驗收（真 PostgreSQL）。單元測試的假 store 驗不到「同一交易」的真正語意
// （FOR UPDATE 的列鎖、commit 與 rollback）與欄位落地，故這裡再看一次四件事：
//
//	① 期別轉 paid，且 paid_at／invoice_no／provider／交易號／人工註記逐欄落地；
//	② 訂閱由 suspended 轉 active 且寬限期清空；
//	③ subscription.reactivated（payload 的 company_id 真的寫進 jsonb）與
//	   period.payment_recorded 各 1 筆，平台稽核 1 筆且 actor 為 operator、reason 為人工理由；
//	④ 同交易號重送 → 不再增加事件與稽核（webhook 重送安全，G3）；
//	⑤ 重送仍可補寫短收／溢收備註（G8）——note 落地、憑據不動、事件與稽核不變。
//
// 這條路徑同時是 T4 唯一的整合測試：T9 的 RPC 只做身分、參數驗證與快取失效，DB 層行為不重測。
package billing_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

func TestIntegrationRecordPayment(t *testing.T) {
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
	ctx := t.Context()

	var planID, subID, periodID, opID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	grace := time.Now().Add(48 * time.Hour)
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, grace_until)
		VALUES (42, $1, 'suspended', 3, $2) RETURNING id`, planID, grace).Scan(&subID); err != nil {
		t.Fatalf("subscription: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, currency, status)
		VALUES ($1, 1, now(), now() + interval '30 days', $2, 1500.00, 150.00, 3, 1950.00, 'TWD', 'open')
		RETURNING id`, subID, planID).Scan(&periodID); err != nil {
		t.Fatalf("period: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.operators (email, name) VALUES ('ops@example.com','Ops') RETURNING id`).
		Scan(&opID); err != nil {
		t.Fatalf("operator: %v", err)
	}

	b := billing.NewBilling(postgres.New(db))
	in := billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ExternalRef: "BANK-12345",
		InvoiceNo: "AB12345678", Note: "短收 50 元，已於 6/1 補足",
		ActorOperatorID: opID, Reason: "匯款入帳（台銀 12345）",
	}
	if _, err := b.RecordPayment(ctx, in); err != nil {
		t.Fatalf("第一次收款: %v", err)
	}

	// ① 期別：paid、付款憑據與人工註記落地
	var pStatus, provider, externalRef, invoiceNo, note string
	var paidAt sql.NullTime
	if err := db.QueryRowContext(ctx, `
		SELECT status, COALESCE(payment_provider,''), COALESCE(external_ref,''),
		       COALESCE(invoice_no,''), note, paid_at
		  FROM platform.subscription_periods WHERE id = $1`, periodID).
		Scan(&pStatus, &provider, &externalRef, &invoiceNo, &note, &paidAt); err != nil {
		t.Fatalf("查期別: %v", err)
	}
	if pStatus != "paid" || !paidAt.Valid {
		t.Fatalf("期別應為 paid 且有 paid_at，got status=%s paid_at=%v", pStatus, paidAt)
	}
	if externalRef != "BANK-12345" || provider != "manual" {
		t.Fatalf("交易號與 provider 應落地，got external_ref=%q provider=%q", externalRef, provider)
	}
	if invoiceNo != "AB12345678" || note != "短收 50 元，已於 6/1 補足" {
		t.Fatalf("發票號與人工註記應落地，got invoice_no=%q note=%q", invoiceNo, note)
	}

	// ② 訂閱：轉 active 且清空寬限期
	var subStatus string
	var grace2 sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT status, grace_until FROM platform.subscriptions WHERE id = $1`, subID).
		Scan(&subStatus, &grace2); err != nil {
		t.Fatalf("查訂閱: %v", err)
	}
	if subStatus != "active" {
		t.Fatalf("訂閱應轉 active，got %s", subStatus)
	}
	if grace2.Valid {
		t.Fatalf("復原後應清空寬限期，got %v", grace2.Time)
	}

	// ③ 事件（payload 真的進 jsonb）與稽核（actor 為 operator）
	var reactivated, recorded, audits int
	var companyID, auditTarget, auditReason string
	if err := db.QueryRowContext(ctx, `
		SELECT
		 (SELECT count(*) FROM platform.events WHERE event_type = 'subscription.reactivated'),
		 (SELECT count(*) FROM platform.events WHERE event_type = 'period.payment_recorded'),
		 (SELECT COALESCE(max(payload->>'company_id'), '') FROM platform.events
		   WHERE event_type = 'subscription.reactivated'),
		 (SELECT count(*) FROM platform.audit_logs WHERE action = 'record_payment'),
		 (SELECT COALESCE(max(target_type || '/' || target_id), '') FROM platform.audit_logs
		   WHERE action = 'record_payment'),
		 (SELECT COALESCE(max(reason), '') FROM platform.audit_logs WHERE action = 'record_payment')`).
		Scan(&reactivated, &recorded, &companyID, &audits, &auditTarget, &auditReason); err != nil {
		t.Fatalf("查事件與稽核: %v", err)
	}
	if reactivated != 1 || recorded != 1 || audits != 1 {
		t.Fatalf("事件/稽核計數不符: reactivated=%d recorded=%d audits=%d", reactivated, recorded, audits)
	}
	if companyID != "42" {
		t.Fatalf("復原事件的 payload 應帶 company_id=42（consumer 不得再查 DB），got %q", companyID)
	}
	if auditTarget != "subscription/"+strconv.FormatInt(subID, 10) || auditReason != in.Reason {
		t.Fatalf("稽核目標/原因不符: target=%q reason=%q", auditTarget, auditReason)
	}
	var auditActor int64
	if err := db.QueryRowContext(ctx,
		`SELECT operator_id FROM platform.audit_logs WHERE action = 'record_payment'`).Scan(&auditActor); err != nil {
		t.Fatalf("查稽核主體: %v", err)
	}
	if auditActor != opID {
		t.Fatalf("稽核 actor 應為操作者 %d，got %d", opID, auditActor)
	}

	// ④ 同交易號重送：no-op，不得再增加事件與稽核
	if _, err := b.RecordPayment(ctx, in); err != nil {
		t.Fatalf("重送收款: %v", err)
	}
	var recorded2, audits2 int
	if err := db.QueryRowContext(ctx, `
		SELECT
		 (SELECT count(*) FROM platform.events WHERE event_type = 'period.payment_recorded'),
		 (SELECT count(*) FROM platform.audit_logs WHERE action = 'record_payment')`).
		Scan(&recorded2, &audits2); err != nil {
		t.Fatalf("重送後查計數: %v", err)
	}
	if recorded2 != 1 || audits2 != 1 {
		t.Fatalf("重送不得重複入帳: recorded=%d audits=%d（應維持 1/1）", recorded2, audits2)
	}

	// ⑤ 重送仍可補寫短收／溢收備註（G8）：note 改變、憑據不變、事件與稽核仍不增加
	replay := in
	replay.Note = "6/1 已補足差額 50 元"
	if _, err := b.RecordPayment(ctx, replay); err != nil {
		t.Fatalf("以新備註重送: %v", err)
	}
	var note2 string
	if err := db.QueryRowContext(ctx,
		`SELECT note FROM platform.subscription_periods WHERE id = $1`, periodID).Scan(&note2); err != nil {
		t.Fatalf("查補寫後的備註: %v", err)
	}
	if note2 != "6/1 已補足差額 50 元" {
		t.Fatalf("重送應可補寫備註（短收／溢收的唯一落點），got %q", note2)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT status, COALESCE(invoice_no,''), COALESCE(external_ref,'')
		  FROM platform.subscription_periods WHERE id = $1`, periodID).
		Scan(&pStatus, &invoiceNo, &externalRef); err != nil {
		t.Fatalf("查補寫後的憑據: %v", err)
	}
	if pStatus != "paid" || invoiceNo != "AB12345678" || externalRef != "BANK-12345" {
		t.Fatalf("補寫備註不得動到付款憑據: status=%q invoice_no=%q external_ref=%q",
			pStatus, invoiceNo, externalRef)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT
		 (SELECT count(*) FROM platform.events WHERE event_type = 'period.payment_recorded'),
		 (SELECT count(*) FROM platform.audit_logs WHERE action = 'record_payment')`).
		Scan(&recorded2, &audits2); err != nil {
		t.Fatalf("補寫備註後查計數: %v", err)
	}
	if recorded2 != 1 || audits2 != 1 {
		t.Fatalf("補寫備註不得再寫事件與稽核: recorded=%d audits=%d（應維持 1/1）", recorded2, audits2)
	}
}
