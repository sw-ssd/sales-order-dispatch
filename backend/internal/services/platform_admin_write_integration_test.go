//go:build integration

// T9 的寫入 RPC 端到端測試:真 PostgreSQL ＋ 真 Valkey ＋ 真的 postgres.Admin／postgres.Store。
//
// 為什麼假 store 不夠(三件事在記憶體上永遠測不出來):
//   - **開票三欄真的落進 platform.subscription_periods**(T4 的缺口):欄位名、CASE 條件、
//     NULLIF 的處理都是 SQL 的事,假實作根本沒有那三欄;
//   - **同一交易**:資料寫了但稽核沒寫(事後查不出是誰)與稽核寫了但資料沒寫,兩者都只在
//     真的 BEGIN／COMMIT 上才看得出;失敗時「不留半成品」同理;
//   - **快取失效**:Valkey 上的鍵是跨行程的契約(排程與 consumer 刪的就是它),假快取只驗得到
//     「有呼叫 Delete」,驗不到「鍵真的消失」。
//
// 執行:task test:integration -- -count=1 -run TestIntegrationPlatformAdminWrite -v
package services

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/cron"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

// integrationSeatCounter 為席位用量的固定計數器:本檔要驗的是 RPC 的寫入路徑(欄位、交易、
// 稽核、快取),真實的席位計數(未停用帳號數)由 counters_integration_test.go 以真 ent client 守。
type integrationSeatCounter struct{ used int }

func (c integrationSeatCounter) Count(context.Context, int, string) (int, error) { return c.used, nil }

// writeRig 為寫入測試的夾具:真的服務 ＋ 真連線 ＋ 真快取 ＋ 種好的資料。
type writeRig struct {
	svc   *PlatformAdminService
	admin *sql.DB
	cache entitlements.Cache
	seed  platformAdminSeed
	ctx   context.Context
}

// auditCount 回某個 action 的稽核筆數(斷言「恰一筆」用差量,不用全表計數:夾具本身也有稽核)。
func (r *writeRig) auditCount(t *testing.T, action string) int {
	t.Helper()
	var n int
	if err := r.admin.QueryRowContext(t.Context(),
		`SELECT count(*) FROM platform.audit_logs WHERE action = $1`, action).Scan(&n); err != nil {
		t.Fatalf("統計稽核: %v", err)
	}
	return n
}

// lastAudit 取某個 action 最新一筆的 actor 與 reason(平台稽核要回答「誰、為什麼」)。
func (r *writeRig) lastAudit(t *testing.T, action string) (int64, string, string, string) {
	t.Helper()
	var (
		operatorID      int64
		reason, ttype   string
		targetID, email string
	)
	if err := r.admin.QueryRowContext(t.Context(), `
		SELECT a.operator_id, a.reason, a.target_type, a.target_id, o.email
		  FROM platform.audit_logs a JOIN platform.operators o ON o.id = a.operator_id
		 WHERE a.action = $1 ORDER BY a.id DESC LIMIT 1`, action).
		Scan(&operatorID, &reason, &ttype, &targetID, &email); err != nil {
		t.Fatalf("讀稽核(%s): %v", action, err)
	}
	return operatorID, reason, ttype, targetID + "|" + email
}

// newWriteRig 種出寫入測試需要的資料:一家有 active 訂閱的租戶(席位 8、一期 open)、
// 一家 G5 的平台自營公司(不得出現在待收款)、一個 operator。
func newWriteRig(t *testing.T) *writeRig {
	t.Helper()
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN) // 00029／00030(platform schema 與 settings)也走這條路徑
	admin := openRawDB(t, adminDSN)
	seed := seedPlatformAdmin(t, admin)

	// 真 Valkey:失效必須是「鍵真的消失」,不是「有呼叫 Delete」。
	client := cache.NewClient(testsupport.Valkey(t))
	t.Cleanup(func() { _ = client.Close() })
	entCache := entitlements.NewValkeyCache(client)

	book := billing.NewBilling(platformstore.New(admin)).WithCache(entCache)
	svc := NewPlatformAdminService(platformstore.NewAdmin(admin), book, entCache,
		integrationSeatCounter{used: 3})

	return &writeRig{
		svc: svc, admin: admin, cache: entCache, seed: seed,
		ctx: operatorauth.WithIdentity(t.Context(),
			operatorauth.Identity{OperatorID: seed.operatorID, Email: "ops-a@example.com", Role: "admin"}),
	}
}

// seedCachedEntitlement 在真 Valkey 上放一份該租戶的權益快取(後續斷言它被刪掉)。
func (r *writeRig) seedCachedEntitlement(t *testing.T, companyID int) {
	t.Helper()
	key := "ent:" + itoa(companyID)
	if err := r.cache.Set(t.Context(), key, []byte(`{"status":"active"}`), time.Hour); err != nil {
		t.Fatalf("seed 快取 %s: %v", key, err)
	}
	if _, ok, err := r.cache.Get(t.Context(), key); err != nil || !ok {
		t.Fatalf("seed 後必須命中: ok=%v err=%v", ok, err)
	}
}

// assertCacheInvalidated 斷言該租戶的權益快取已從真 Valkey 消失。
func (r *writeRig) assertCacheInvalidated(t *testing.T, companyID int) {
	t.Helper()
	key := "ent:" + itoa(companyID)
	if _, ok, err := r.cache.Get(t.Context(), key); err != nil || ok {
		t.Fatalf("寫入後 %s 必須失效: ok=%v err=%v", key, ok, err)
	}
}

// TestIntegrationPlatformAdminWrite 驗每個寫入 RPC 的落地:欄位寫進正確的表、恰一筆稽核、
// 快取真的失效、失敗不留半成品。
func TestIntegrationPlatformAdminWrite(t *testing.T) {
	rig := newWriteRig(t)
	company := int(rig.seed.activeID)

	t.Run("RecordPayment 落地開票三欄與付款憑據", func(t *testing.T) {
		rig.seedCachedEntitlement(t, company)
		before := rig.auditCount(t, "record_payment")

		resp, err := rig.svc.RecordPayment(rig.ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
			CompanyId: itoa(company), PeriodNo: 2, Amount: "2600.00", Provider: "manual",
			ExternalRef: "BANK-1", InvoiceNo: "AB12345678", InvoiceStatus: "issued",
			BuyerTaxId: "12345678", Carrier: "/ABC1234", Note: "短收 100 元", Reason: "匯款入帳"}))
		if err != nil {
			t.Fatalf("RecordPayment: %v", err)
		}
		if resp.Msg.GetPeriodNo() != 2 || resp.Msg.GetStatus() != "paid" {
			t.Fatalf("回應應帶期別與**新狀態**: %v", resp.Msg)
		}

		// 期別:狀態、憑據與**開票三欄**(T4 的缺口)都要落在同一列。
		var (
			status, invoiceNo, invoiceStatus, buyerTaxID, carrier, note string
			amountCents                                                 int64
			paidAt                                                      sql.NullTime
		)
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT status, COALESCE(invoice_no,''), COALESCE(invoice_status,''),
			       COALESCE(buyer_tax_id,''), COALESCE(carrier,''), note,
			       (amount*100)::bigint, paid_at
			  FROM platform.subscription_periods
			 WHERE subscription_id = (SELECT id FROM platform.subscriptions WHERE company_id = $1)
			   AND period_no = 2`, company).
			Scan(&status, &invoiceNo, &invoiceStatus, &buyerTaxID, &carrier, &note, &amountCents, &paidAt); err != nil {
			t.Fatalf("讀期別: %v", err)
		}
		if status != "paid" || !paidAt.Valid {
			t.Fatalf("期別應轉 paid 並記下付款時間: status=%s paid_at=%v", status, paidAt)
		}
		if invoiceNo != "AB12345678" || invoiceStatus != "issued" || buyerTaxID != "12345678" ||
			carrier != "/ABC1234" {
			t.Fatalf("開票欄位必須落地: invoice_no=%q invoice_status=%q buyer_tax_id=%q carrier=%q",
				invoiceNo, invoiceStatus, buyerTaxID, carrier)
		}
		if note != "短收 100 元" {
			t.Fatalf("備註必須落地(G8): %q", note)
		}
		if amountCents != 260000 {
			t.Fatalf("期別金額不得被輸入金額改動(G4 不支援部分付款): %d", amountCents)
		}

		// 恰一筆稽核,actor 是真的 operator,原因是真的原因(store 與服務層各寫一筆就是兩份真相)。
		if got := rig.auditCount(t, "record_payment"); got != before+1 {
			t.Fatalf("收款必須恰寫一筆稽核: %d → %d", before, got)
		}
		opID, reason, ttype, extra := rig.lastAudit(t, "record_payment")
		if opID != rig.seed.operatorID || reason != "匯款入帳" {
			t.Fatalf("稽核的 actor／原因錯誤: op=%d reason=%q", opID, reason)
		}
		if ttype != "subscription" || extra != itoa(int(rig.seed.activeID))+"|ops-a@example.com" {
			t.Fatalf("稽核的目標／operator email 錯誤: %s %s", ttype, extra)
		}
		rig.assertCacheInvalidated(t, company)
	})

	t.Run("RecordPayment 重送不覆寫開票欄位也不重寫稽核", func(t *testing.T) {
		before := rig.auditCount(t, "record_payment")
		// 同一交易號的重送:no-op(不重寫事件與稽核),但**已存的憑據一個都不能動**。
		if _, err := rig.svc.RecordPayment(rig.ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
			CompanyId: itoa(company), PeriodNo: 2, Provider: "manual", ExternalRef: "BANK-1",
			Reason: "重送"})); err != nil {
			t.Fatalf("重送應為 no-op: %v", err)
		}
		if got := rig.auditCount(t, "record_payment"); got != before {
			t.Fatalf("重送不得再寫稽核: %d → %d", before, got)
		}
		var invoiceStatus, buyerTaxID string
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT COALESCE(invoice_status,''), COALESCE(buyer_tax_id,'')
			  FROM platform.subscription_periods
			 WHERE subscription_id = (SELECT id FROM platform.subscriptions WHERE company_id = $1)
			   AND period_no = 2`, company).Scan(&invoiceStatus, &buyerTaxID); err != nil {
			t.Fatalf("讀期別: %v", err)
		}
		if invoiceStatus != "issued" || buyerTaxID != "12345678" {
			t.Fatalf("重送不得把已存的開票欄位清成 NULL: %q %q", invoiceStatus, buyerTaxID)
		}
	})

	t.Run("SetSeatCount 更新席位並留下稽核", func(t *testing.T) {
		rig.seedCachedEntitlement(t, company)

		resp, err := rig.svc.SetSeatCount(rig.ctx, connect.NewRequest(&platformv1.SetSeatCountRequest{
			CompanyId: itoa(company), SeatCount: 10, Reason: "擴編"}))
		if err != nil {
			t.Fatalf("SetSeatCount: %v", err)
		}
		if resp.Msg.GetSeatCount() != 10 {
			t.Fatalf("回應應帶新席位數: %v", resp.Msg)
		}
		var seats int
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT seat_count FROM platform.subscriptions WHERE company_id = $1 AND status <> 'cancelled'`,
			company).Scan(&seats); err != nil {
			t.Fatalf("讀訂閱: %v", err)
		}
		if seats != 10 {
			t.Fatalf("席位應更新為 10,got %d", seats)
		}
		if opID, _, _, _ := rig.lastAudit(t, "subscription.set_seats"); opID != rig.seed.operatorID {
			t.Fatalf("稽核 actor 應為 operator(%d),got %d", rig.seed.operatorID, opID)
		}
		rig.assertCacheInvalidated(t, company)
	})

	t.Run("SetSeatCount 低於使用量回 PLAT-5001 且不留痕跡", func(t *testing.T) {
		before := rig.auditCount(t, "subscription.set_seats")
		_, err := rig.svc.SetSeatCount(rig.ctx, connect.NewRequest(&platformv1.SetSeatCountRequest{
			CompanyId: itoa(company), SeatCount: 2, Reason: "降席位"}))
		if connect.CodeOf(err) != connect.CodeFailedPrecondition || errorInfoOf(t, err).GetCode() != "PLAT-5001" {
			t.Fatalf("席次低於使用量應 PLAT-5001,got %v", err)
		}
		if got := rig.auditCount(t, "subscription.set_seats"); got != before {
			t.Fatalf("被拒絕的呼叫不得寫稽核: %d → %d", before, got)
		}
		var seats int
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT seat_count FROM platform.subscriptions WHERE company_id = $1 AND status <> 'cancelled'`,
			company).Scan(&seats); err != nil {
			t.Fatalf("讀訂閱: %v", err)
		}
		if seats != 10 {
			t.Fatalf("被拒絕的呼叫不得改動席位,got %d", seats)
		}
	})

	t.Run("ChangePlan 下一期生效且不動當期快照", func(t *testing.T) {
		rig.seedCachedEntitlement(t, company)
		var proID int64
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT id FROM platform.plans WHERE code = 'pro'`).Scan(&proID); err != nil {
			t.Fatalf("讀方案 pro: %v", err)
		}

		resp, err := rig.svc.ChangePlan(rig.ctx, connect.NewRequest(&platformv1.ChangePlanRequest{
			CompanyId: itoa(company), PlanCode: "pro", Reason: "升級方案"}))
		if err != nil {
			t.Fatalf("ChangePlan: %v", err)
		}
		var planID int64
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT plan_id FROM platform.subscriptions WHERE company_id = $1 AND status <> 'cancelled'`,
			company).Scan(&planID); err != nil {
			t.Fatalf("讀訂閱: %v", err)
		}
		if planID != proID {
			t.Fatalf("訂閱應改掛 pro(%d),got %d", proID, planID)
		}
		// 當期(期別 2,已付款)的計畫與金額快照不得被改動 —— 調方案不得回溯改帳。
		var (
			periodPlanID int64
			amountCents  int64
		)
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT plan_id, (amount*100)::bigint FROM platform.subscription_periods
			 WHERE subscription_id = (SELECT id FROM platform.subscriptions WHERE company_id = $1)
			   AND period_no = 2`, company).Scan(&periodPlanID, &amountCents); err != nil {
			t.Fatalf("讀期別: %v", err)
		}
		if amountCents != 260000 {
			t.Fatalf("當期期別的金額快照不得被改動: %d", amountCents)
		}
		if resp.Msg.GetPlanCode() != "pro" || resp.Msg.GetEffectiveFrom() == "" {
			t.Fatalf("回應應帶新方案與生效日: %v", resp.Msg)
		}
		rig.assertCacheInvalidated(t, company)
	})

	t.Run("ChangePlan 未知方案回 SYS-4002 且不留半成品", func(t *testing.T) {
		before := rig.auditCount(t, "subscription.change_plan")
		_, err := rig.svc.ChangePlan(rig.ctx, connect.NewRequest(&platformv1.ChangePlanRequest{
			CompanyId: itoa(company), PlanCode: "nope", Reason: "換方案"}))
		if connect.CodeOf(err) != connect.CodeNotFound || errorInfoOf(t, err).GetCode() != "SYS-4002" {
			t.Fatalf("未知方案應 SYS-4002,got %v", err)
		}
		if got := rig.auditCount(t, "subscription.change_plan"); got != before {
			t.Fatalf("失敗不得寫稽核(半成品): %d → %d", before, got)
		}
	})

	t.Run("CancelSubscription 期末終止並發事件", func(t *testing.T) {
		rig.seedCachedEntitlement(t, company)

		resp, err := rig.svc.CancelSubscription(rig.ctx, connect.NewRequest(&platformv1.CancelSubscriptionRequest{
			CompanyId: itoa(company), AtPeriodEnd: true, Reason: "客戶不續約"}))
		if err != nil {
			t.Fatalf("CancelSubscription: %v", err)
		}
		if resp.Msg.GetCancelledAt() == "" || resp.Msg.GetServiceUntil() == "" {
			t.Fatalf("回應應帶取消時間與服務到期時間: %v", resp.Msg)
		}
		var (
			status      string
			cancelledAt sql.NullTime
		)
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT status, cancelled_at FROM platform.subscriptions WHERE company_id = $1 AND status = 'cancelled'`,
			company).Scan(&status, &cancelledAt); err != nil {
			t.Fatalf("訂閱必須仍是 cancelled(期末終止、資料不刪除): %v", err)
		}
		if !cancelledAt.Valid {
			t.Fatal("取消必須記下 cancelled_at")
		}
		// 事件與狀態在同一個交易:少了它,排程的 ExpireCancelled 永遠不會把公司停掉(G7)。
		var payload string
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT payload::text FROM platform.events
			 WHERE event_type = 'subscription.cancelled'
			 ORDER BY id DESC LIMIT 1`).Scan(&payload); err != nil {
			t.Fatalf("取消必須發 subscription.cancelled: %v", err)
		}
		if !strings.Contains(payload, `"company_id": `+itoa(company)) ||
			!strings.Contains(payload, `"reason": "客戶不續約"`) {
			t.Fatalf("事件 payload 必須自帶 company_id 與 reason: %s", payload)
		}
		rig.assertCacheInvalidated(t, company)
	})

	t.Run("SetTenantOverride 與 RevokeTenantOverride", func(t *testing.T) {
		rig.seedCachedEntitlement(t, company)

		resp, err := rig.svc.SetTenantOverride(rig.ctx, connect.NewRequest(&platformv1.SetTenantOverrideRequest{
			CompanyId: itoa(company), FeatureCode: "feature.export", EnabledSet: true, Enabled: true,
			Owner: "業務C", ExpiresAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
			Reason: "簽約贈送匯出"}))
		if err != nil {
			t.Fatalf("SetTenantOverride: %v", err)
		}
		var (
			enabled   sql.NullBool
			owner     string
			createdBy int64
		)
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT enabled, owner, created_by FROM platform.tenant_overrides WHERE id = $1`,
			resp.Msg.GetId()).Scan(&enabled, &owner, &createdBy); err != nil {
			t.Fatalf("讀例外: %v", err)
		}
		if !enabled.Valid || !enabled.Bool || owner != "業務C" || createdBy != rig.seed.operatorID {
			t.Fatalf("例外必須落地且留下承諾者／建立者: enabled=%v owner=%q created_by=%d",
				enabled, owner, createdBy)
		}
		// 新增沒有「before 映像」→ 稽核的 before 必須是 SQL NULL,而不是 jsonb 的 'null'
		// (後者讀起來是「值就是 null」,與「沒有這個資訊」是兩件事)。
		var beforeNull bool
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT before IS NULL FROM platform.audit_logs
			  WHERE action = 'override.set' ORDER BY id DESC LIMIT 1`).Scan(&beforeNull); err != nil {
			t.Fatalf("讀稽核: %v", err)
		}
		if !beforeNull {
			t.Fatal("新增的稽核不得有 before 映像(應為 SQL NULL)")
		}
		rig.assertCacheInvalidated(t, company)

		// 同一 feature 再給一次 → SYS-2001(生效中的例外只能有一筆),且不留痕跡。
		before := rig.auditCount(t, "override.set")
		if _, err := rig.svc.SetTenantOverride(rig.ctx, connect.NewRequest(&platformv1.SetTenantOverrideRequest{
			CompanyId: itoa(company), FeatureCode: "feature.export", EnabledSet: true, Enabled: false,
			Owner: "業務C", Reason: "再給一次"})); connect.CodeOf(err) != connect.CodeAlreadyExists {
			t.Fatalf("重複的生效中例外應 AlreadyExists,got %v", err)
		}
		if got := rig.auditCount(t, "override.set"); got != before {
			t.Fatalf("失敗不得寫稽核: %d → %d", before, got)
		}

		// 撤銷 → revoked_at 落地,而稽核的目標是**公司**(營運查得到的那個 id)。
		rig.seedCachedEntitlement(t, company)
		if _, err := rig.svc.RevokeTenantOverride(rig.ctx, connect.NewRequest(&platformv1.RevokeTenantOverrideRequest{
			OverrideId: resp.Msg.GetId(), Reason: "合約到期"})); err != nil {
			t.Fatalf("RevokeTenantOverride: %v", err)
		}
		var revokedAt sql.NullTime
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT revoked_at FROM platform.tenant_overrides WHERE id = $1`, resp.Msg.GetId()).
			Scan(&revokedAt); err != nil {
			t.Fatalf("讀例外: %v", err)
		}
		if !revokedAt.Valid {
			t.Fatal("撤銷必須寫下 revoked_at(唯一索引只認未撤銷者)")
		}
		opID, reason, ttype, target := rig.lastAudit(t, "override.revoke")
		if opID != rig.seed.operatorID || reason != "合約到期" {
			t.Fatalf("撤銷稽核的 actor／原因錯誤: %d %q", opID, reason)
		}
		if ttype != "company" || target != itoa(company)+"|ops-a@example.com" {
			t.Fatalf("撤銷稽核的目標應為該公司: %s %s", ttype, target)
		}
		rig.assertCacheInvalidated(t, company)
	})

	t.Run("UpsertPlanPrice 與 SetPlanEntitlement 全量失效", func(t *testing.T) {
		// 兩個租戶的權益鍵都放進 Valkey:方案異動必須把**所有**租戶的鍵清掉。
		rig.seedCachedEntitlement(t, company)
		rig.seedCachedEntitlement(t, int(rig.seed.noneID))
		rig.seedCachedEntitlement(t, int(rig.seed.bothID))

		if _, err := rig.svc.UpsertPlanPrice(rig.ctx, connect.NewRequest(&platformv1.UpsertPlanPriceRequest{
			PlanCode: "std", BillingCycle: "monthly", BasePrice: "1100.00", SeatPrice: "220.00",
			Reason: "調價"})); err != nil {
			t.Fatalf("UpsertPlanPrice: %v", err)
		}
		if _, err := rig.svc.SetPlanEntitlement(rig.ctx, connect.NewRequest(&platformv1.SetPlanEntitlementRequest{
			PlanCode: "std", FeatureCode: "feature.export", Enabled: true, LimitSet: true, LimitValue: 5,
			Reason: "加值"})); err != nil {
			t.Fatalf("SetPlanEntitlement: %v", err)
		}

		// 價目是**新增一列**(價格史):現行價取 effective_from 最新者 = 1100.00。
		var base string
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT base_price::text FROM platform.plan_prices p JOIN platform.plans pl ON pl.id = p.plan_id
			 WHERE pl.code = 'std' AND p.billing_cycle = 'monthly'
			 ORDER BY p.effective_from DESC, p.id DESC LIMIT 1`).Scan(&base); err != nil {
			t.Fatalf("讀價目: %v", err)
		}
		if base != "1100.00" {
			t.Fatalf("新價目必須是最新的一列: got %s", base)
		}
		var (
			enabled bool
			limit   sql.NullInt64
		)
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT e.enabled, e.limit_value FROM platform.plan_entitlements e
			  JOIN platform.plans pl ON pl.id = e.plan_id
			 WHERE pl.code = 'std' AND e.feature_code = 'feature.export'`).Scan(&enabled, &limit); err != nil {
			t.Fatalf("讀方案權益: %v", err)
		}
		if !enabled || !limit.Valid || limit.Int64 != 5 {
			t.Fatalf("方案權益必須落地: enabled=%v limit=%v", enabled, limit)
		}

		// 全量失效:三個租戶的鍵都不見了(方案／價目的變更影響每一個用到它的租戶)。
		rig.assertCacheInvalidated(t, company)
		rig.assertCacheInvalidated(t, int(rig.seed.noneID))
		rig.assertCacheInvalidated(t, int(rig.seed.bothID))
	})

	t.Run("UpdateBillingSettings 落地且拒絕未知鍵", func(t *testing.T) {
		mustExec(t, rig.admin, `INSERT INTO platform.settings (key, value) VALUES
			('trial_days','14'),('grace_days','7'),('lead_days','14')
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`)

		if _, err := rig.svc.UpdateBillingSettings(rig.ctx,
			connect.NewRequest(&platformv1.UpdateBillingSettingsRequest{
				Settings: []*platformv1.BillingSetting{{Key: "grace_days", Value: "3"}},
				Reason:   "縮短寬限"})); err != nil {
			t.Fatalf("UpdateBillingSettings: %v", err)
		}
		var value string
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT value FROM platform.settings WHERE key = 'grace_days'`).Scan(&value); err != nil {
			t.Fatalf("讀設定: %v", err)
		}
		if value != "3" {
			t.Fatalf("設定值必須落地: %q", value)
		}
		// 未知鍵:整批拒絕(打錯的鍵會被 upsert 成一個永遠讀不到的新列,而 console 顯示「已儲存」)。
		before := rig.auditCount(t, "settings.update")
		if _, err := rig.svc.UpdateBillingSettings(rig.ctx,
			connect.NewRequest(&platformv1.UpdateBillingSettingsRequest{
				Settings: []*platformv1.BillingSetting{{Key: "nope", Value: "1"}},
				Reason:   "亂改"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("未知鍵應 InvalidArgument,got %v", err)
		}
		if got := rig.auditCount(t, "settings.update"); got != before {
			t.Fatalf("失敗不得寫稽核: %d → %d", before, got)
		}
	})

	t.Run("CreateOperator 與 DisableOperator", func(t *testing.T) {
		resp, err := rig.svc.CreateOperator(rig.ctx, connect.NewRequest(&platformv1.CreateOperatorRequest{
			Email: "new-op@example.com", Name: "新同事", Role: "operator", Reason: "到職"}))
		if err != nil {
			t.Fatalf("CreateOperator: %v", err)
		}
		if _, err := rig.svc.CreateOperator(rig.ctx, connect.NewRequest(&platformv1.CreateOperatorRequest{
			Email: "new-op@example.com", Name: "重複", Role: "operator", Reason: "再新增"})); connect.CodeOf(err) != connect.CodeAlreadyExists {
			t.Fatalf("重複 email 應 AlreadyExists,got %v", err)
		}
		if _, err := rig.svc.DisableOperator(rig.ctx, connect.NewRequest(&platformv1.DisableOperatorRequest{
			OperatorId: resp.Msg.GetId(), Reason: "離職"})); err != nil {
			t.Fatalf("DisableOperator: %v", err)
		}
		var status string
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT status FROM platform.operators WHERE id = $1`, resp.Msg.GetId()).Scan(&status); err != nil {
			t.Fatalf("讀 operator: %v", err)
		}
		if status != "disabled" {
			t.Fatalf("停用必須即時生效(status=disabled): %q", status)
		}
		for _, action := range []string{"operator.create", "operator.disable"} {
			if opID, _, _, _ := rig.lastAudit(t, action); opID != rig.seed.operatorID {
				t.Fatalf("%s 的稽核 actor 應為 operator(%d),got %d", action, rig.seed.operatorID, opID)
			}
		}
	})

	t.Run("ListReceivables 排除平台自營公司並帶公司名", func(t *testing.T) {
		// 平台自營公司的一期:待收款清單不得列出(它不是租戶,對它催收等於對平台自己催收)。
		var planID int64
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT id FROM platform.plans WHERE code = 'std'`).Scan(&planID); err != nil {
			t.Fatalf("讀方案: %v", err)
		}
		var ownSubID int64
		if err := rig.admin.QueryRowContext(t.Context(), `
			INSERT INTO platform.subscriptions (company_id, plan_id, seat_count, billing_cycle, status)
			VALUES ($1, $2, 1, 'monthly', 'active') RETURNING id`,
			rig.seed.platformOwnedID, planID).Scan(&ownSubID); err != nil {
			t.Fatalf("seed 平台自營公司的訂閱: %v", err)
		}
		mustExec(t, rig.admin, `INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, status) VALUES
			($1, 1, now() - interval '10 days', now() + interval '10 days', $2, 1000.00, 200.00, 1, 1200.00, 'open')`,
			ownSubID, planID)

		resp, err := rig.svc.ListReceivables(rig.ctx, connect.NewRequest(&platformv1.ListReceivablesRequest{
			Page: 1, PageSize: 50}))
		if err != nil {
			t.Fatalf("ListReceivables: %v", err)
		}
		byCompany := map[string]*platformv1.Receivable{}
		for _, row := range resp.Msg.GetRows() {
			byCompany[row.GetCompanyId()] = row
		}
		if _, ok := byCompany[itoa(int(rig.seed.platformOwnedID))]; ok {
			t.Fatal("待收款清單不得包含 G5 的平台自營公司(與 console 的租戶投影一致)")
		}
		row := byCompany[itoa(company)]
		if row == nil {
			t.Fatalf("有未付期別的租戶必須在清單內: %+v", resp.Msg.GetRows())
		}
		if row.GetCompanyName() != "甲公司" {
			t.Fatalf("必須帶公司名: %+v", row)
		}
		// 方案取該公司的訂閱(優先未取消者,與投影同一個取法):前面的子測試改過方案、也取消過
		// 訂閱,故不得硬編 std,也不得只認 status <> 'cancelled'(取消後那筆合約還在)。
		var planCode string
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT COALESCE(p.code,'') FROM platform.subscriptions s
			  LEFT JOIN platform.plans p ON p.id = s.plan_id
			 WHERE s.company_id = $1
			 ORDER BY (s.status <> 'cancelled') DESC, s.id DESC
			 LIMIT 1`, company).Scan(&planCode); err != nil {
			t.Fatalf("讀訂閱方案: %v", err)
		}
		if row.GetPlanCode() != planCode {
			t.Fatalf("方案必須是現行訂閱的方案: got %q want %q", row.GetPlanCode(), planCode)
		}
		// 金額一律兩位小數字串(前端據此直接匯出 CSV,不得再自行格式化)且**等於該期別的金額**。
		if want := money.FormatCents(loadPeriodCents(t, rig, company, int(row.GetPeriodNo()))); row.GetAmount() != want {
			t.Fatalf("金額格式／數值錯誤: got %q want %q", row.GetAmount(), want)
		}
		if row.GetStatus() != "open" || row.GetPeriodEnd() == "" {
			t.Fatalf("僅列出未付期別且必須帶到期日: %+v", row)
		}
	})

	t.Run("operator 管理需要 admin 角色", func(t *testing.T) {
		opCtx := operatorauth.WithIdentity(t.Context(),
			operatorauth.Identity{OperatorID: rig.seed.operatorID, Email: "ops-a@example.com", Role: "operator"})
		if _, err := rig.svc.CreateOperator(opCtx, connect.NewRequest(&platformv1.CreateOperatorRequest{
			Email: "self-promote@example.com", Role: "admin", Reason: "自我提權"})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("operator 角色建立 admin 應 PermissionDenied,got %v", err)
		}
		var n int
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT count(*) FROM platform.operators WHERE email = 'self-promote@example.com'`).Scan(&n); err != nil {
			t.Fatalf("統計 operator: %v", err)
		}
		if n != 0 {
			t.Fatal("被拒絕的提權不得留下任何白名單列")
		}
		if _, err := rig.svc.DisableOperator(opCtx, connect.NewRequest(&platformv1.DisableOperatorRequest{
			OperatorId: itoa(int(rig.seed.operatorID)), Reason: "停用自己"})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("operator 角色停用他人應 PermissionDenied,got %v", err)
		}
	})

	t.Run("不得停用自己與最後一位 admin(真 SQL)", func(t *testing.T) {
		// ① 自己:即使是 admin,停用自己等於當場把自己鎖在門外(operatorauth 每次請求都查
		// status),而「誰停的」會是那位已經進不來的人 → 服務層擋下。
		_, err := rig.svc.DisableOperator(rig.ctx, connect.NewRequest(&platformv1.DisableOperatorRequest{
			OperatorId: itoa(int(rig.seed.operatorID)), Reason: "停用自己"}))
		if info := errorInfoOf(t, err); info.GetCode() != "PLAT-3003" ||
			!strings.Contains(info.GetMessage(), "不得停用自己") {
			t.Fatalf("停用自己應 PLAT-3003 且訊息可行動,got %v", err)
		}

		// ② 最後一位 admin:夾具裡只有 ops-a 是 active admin。
		// 由**另一位 admin 身分**停用他 —— 那位的白名單列已停用(身分與列不同步:併發互停、
		// 或 token 內的角色已被改掉),這正是 SQL 條件要擋的那條路;服務層的自停用檢查管不到
		// 「別人」。
		var ghostAdminID int64
		if err := rig.admin.QueryRowContext(t.Context(), `
			INSERT INTO platform.operators (email, name, role, status)
			VALUES ('ops-ghost@example.com','已停用的管理員','admin','disabled') RETURNING id`).
			Scan(&ghostAdminID); err != nil {
			t.Fatalf("seed 已停用的 admin: %v", err)
		}
		ghostCtx := operatorauth.WithIdentity(t.Context(),
			operatorauth.Identity{OperatorID: ghostAdminID, Email: "ops-ghost@example.com", Role: "admin"})
		before := rig.auditCount(t, "operator.disable")
		_, err = rig.svc.DisableOperator(ghostCtx, connect.NewRequest(&platformv1.DisableOperatorRequest{
			OperatorId: itoa(int(rig.seed.operatorID)), Reason: "停用最後一位 admin"}))
		if info := errorInfoOf(t, err); info.GetCode() != "PLAT-3003" ||
			!strings.Contains(info.GetMessage(), "最後一位 admin") {
			t.Fatalf("停用最後一位 admin 應 PLAT-3003 且訊息可行動,got %v", err)
		}
		var status string
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT status FROM platform.operators WHERE id = $1`, rig.seed.operatorID).Scan(&status); err != nil {
			t.Fatalf("讀 operator: %v", err)
		}
		if status != "active" {
			t.Fatalf("被擋下時不得改動狀態,got %q", status)
		}
		if got := rig.auditCount(t, "operator.disable"); got != before {
			t.Fatalf("被擋下時不得寫稽核(失敗不留半成品): %d → %d", before, got)
		}

		// ③ 對照組:有**第二位 active admin** 之後,停用 ops-a 就應該成功(條件不是「一律不得
		// 停用 admin」)。第二位用自己的身分停用自己也不行,故由第三位來停 —— 這裡直接用
		// ops-a 的身分(他不能停自己,故用剛剛那位 ghost 的位置改為 active 後由它執行)。
		if _, err := rig.admin.ExecContext(t.Context(),
			`UPDATE platform.operators SET status = 'active' WHERE id = $1`, ghostAdminID); err != nil {
			t.Fatalf("還原 ghost admin: %v", err)
		}
		if _, err := rig.svc.DisableOperator(ghostCtx, connect.NewRequest(&platformv1.DisableOperatorRequest{
			OperatorId: itoa(int(rig.seed.operatorID)), Reason: "職務輪替"})); err != nil {
			t.Fatalf("還有另一位 active admin 時應可停用: %v", err)
		}
		if err := rig.admin.QueryRowContext(t.Context(),
			`SELECT status FROM platform.operators WHERE id = $1`, rig.seed.operatorID).Scan(&status); err != nil {
			t.Fatalf("讀 operator: %v", err)
		}
		if status != "disabled" {
			t.Fatalf("停用應即時生效,got %q", status)
		}
		// 收尾:把 ops-a 還原成 active(後續子測試的語意是「他就是那位 operator」)。
		// 未結項 #24：ghost 不做物理刪除 —— 它已被 ③ 改為 active 又停用了 ops-a，
		// 期間的 operator.disable 稽核 FK 指著它（audit_logs_operator_id_fkey），
		// 物理刪除會撞 23503。改回 disabled：與 ② 剛 seed 進來時的形狀一致，
		// 後續子測試照樣把它當「不存在的第二位」看待（斷言皆用差值）。
		if _, err := rig.admin.ExecContext(t.Context(),
			`UPDATE platform.operators SET status = 'active' WHERE id = $1`, rig.seed.operatorID); err != nil {
			t.Fatalf("還原 ops-a: %v", err)
		}
		if _, err := rig.admin.ExecContext(t.Context(),
			`UPDATE platform.operators SET status = 'disabled' WHERE id = $1`, ghostAdminID); err != nil {
			t.Fatalf("還原 ghost admin: %v", err)
		}
	})

	t.Run("負的限額一律拒絕(真容器)", func(t *testing.T) {
		before := rig.auditCount(t, "override.set")
		if _, err := rig.svc.SetTenantOverride(rig.ctx, connect.NewRequest(&platformv1.SetTenantOverrideRequest{
			CompanyId: itoa(company), FeatureCode: "feature.export", LimitSet: true, LimitValue: -1,
			Owner: "業務E", Reason: "誤填負值"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("負的限額應 InvalidArgument,got %v", err)
		}
		var n int
		if err := rig.admin.QueryRowContext(t.Context(), `
			SELECT count(*) FROM platform.tenant_overrides
			 WHERE company_id = $1 AND feature_code = 'feature.export' AND revoked_at IS NULL`,
			company).Scan(&n); err != nil {
			t.Fatalf("統計例外: %v", err)
		}
		if n != 0 {
			t.Fatal("負值的限額不得寫入(判定層會把它讀成「任何用量都超額」)")
		}
		if got := rig.auditCount(t, "override.set"); got != before {
			t.Fatalf("被拒絕的呼叫不得寫稽核: %d → %d", before, got)
		}
	})

	t.Run("reason 全空白一律拒絕且不留痕跡", func(t *testing.T) {
		before := rig.auditCount(t, "record_payment")
		_, err := rig.svc.RecordPayment(rig.ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
			CompanyId: itoa(company), PeriodNo: 1, Reason: "   "}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument || errorInfoOf(t, err).GetCode() != "SYS-1001" {
			t.Fatalf("空 reason 應 SYS-1001,got %v", err)
		}
		if got := rig.auditCount(t, "record_payment"); got != before {
			t.Fatalf("被拒絕的呼叫不得寫稽核: %d → %d", before, got)
		}
	})
}

// loadPeriodCents 讀某期別的金額(分);找不到即測試失敗。
func loadPeriodCents(t *testing.T, rig *writeRig, companyID, periodNo int) int64 {
	t.Helper()
	var cents int64
	if err := rig.admin.QueryRowContext(t.Context(), `
		SELECT (amount*100)::bigint FROM platform.subscription_periods
		 WHERE subscription_id = (SELECT id FROM platform.subscriptions WHERE company_id = $1)
		   AND period_no = $2`, companyID, periodNo).Scan(&cents); err != nil {
		t.Fatalf("讀期別金額(%d/%d): %v", companyID, periodNo, err)
	}
	return cents
}

// runCronOnce 跑一趟真排程(真 store ＋ 真 billing;派送器替身見 noDispatch),now 由呼叫端給
// (排程不讀時鐘 —— 試用到期／寬限的測試才能把時間往前撥)。
func runCronOnce(t *testing.T, rig *writeRig, now time.Time) cron.Summary {
	t.Helper()
	st := platformstore.New(rig.admin)
	summary, err := cron.RunOnce(t.Context(), cron.Deps{
		Billing: billing.NewBilling(st), Consumer: noDispatch{}, Store: st,
	}, now, cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: cron.EventBatch})
	if err != nil {
		t.Fatalf("跑一趟排程(now=%v): %v", now, err)
	}
	return summary
}

// noDispatch 為本測試的派送器替身:這一趟只驗「排程不會為剛開通的租戶再開一期」的編排路徑,
// 事件的產品域副作用(凍結公司)由 cron 自己的整合測試以真 consumer 把關。
type noDispatch struct{}

func (noDispatch) DispatchOnce(context.Context, int) (int, error) { return 0, nil }

// nextMonthSameDayIn 回「下個月的同一個日號」(該日不存在時取當月最後一日)= 月繳的期末規則。
// 刻意在測試裡再算一次:若實作改用 time.AddDate,1/31 會被正規化成 3/3(跳過整個 2 月)而在此紅。
func nextMonthSameDayIn(from time.Time) time.Time {
	lastDay := time.Date(from.Year(), from.Month()+2, 0, 0, 0, 0, 0, from.Location()).Day()
	day := from.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(from.Year(), from.Month()+1, day,
		from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
}

// TestIntegrationCreateSubscription 走完 B-1 的完整鏈(真 PostgreSQL):**開通 → 排程不重複開期
// → 收款**。這是 spec §7.4 DoD 的前兩跳,而在本任務之前全 repo 沒有任何程式路徑會建立
// platform.subscriptions 與第一期 —— 沒有它,收款／期別／催收／凍結全部 inert(測試都從手工
// 種好的訂閱列起跑)。
//
// 驗七件事:
//
//	① 訂閱列(狀態／方案／計費週期／席位／試用到期);
//	② 第一期(期別 1、金額＝基價＋席位×每席價、價格快照、期末依週期與月底錨點);
//	③ subscription.created 事件的 payload;
//	④ 恰一筆平台稽核(actor 是真的 operator、target 指向新訂閱);
//	⑤ 同一公司的第二次開通被唯一鍵擋下(已註冊的 SYS-2001)且不留痕跡;
//	⑥ 接著跑一趟 cron:不得再開第二個期別;
//	⑦ 接著 RecordPayment 成功(開通前它對這家公司回 PLAT-3001)。
//
// 執行:task test:integration -- -count=1 -run TestIntegrationCreateSubscription -v
func TestIntegrationCreateSubscription(t *testing.T) {
	rig := newWriteRig(t)
	// 完全沒有訂閱列的租戶:開通的前提。
	company := int(rig.seed.noneID)
	ctx := t.Context()

	// ⓪ 開通前:沒有合約 → 不得記帳(B-1 的症狀:不靠手工 SQL 收不到一筆錢)。
	_, err := rig.svc.RecordPayment(rig.ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
		CompanyId: itoa(company), PeriodNo: 0, Reason: "開通前的收款"}))
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-3001" {
		t.Fatalf("沒有訂閱的公司收款應 PLAT-3001,got %q (%v)", got, err)
	}

	rig.seedCachedEntitlement(t, company)
	auditsBefore := rig.auditCount(t, "subscription.create")

	before := time.Now()
	resp, err := rig.svc.CreateSubscription(rig.ctx, connect.NewRequest(&platformv1.CreateSubscriptionRequest{
		CompanyId: itoa(company), PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
		Reason: "客戶簽約開通"}))
	if err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}
	after := time.Now()
	subID := resp.Msg.GetSubscriptionId()
	if subID == "" || resp.Msg.GetStatus() != "active" || resp.Msg.GetPlanCode() != "std" ||
		resp.Msg.GetBillingCycle() != "monthly" || resp.Msg.GetSeatCount() != 3 ||
		resp.Msg.GetTrialEndsAt() != "" || resp.Msg.GetFirstPeriodNo() != 1 {
		t.Fatalf("回應不符: %+v", resp.Msg)
	}
	// 1000.00(基價)＋ 3 × 200.00(席位)= 1600.00;seed 的月繳價目有兩次調價,取現行那一筆。
	if got := resp.Msg.GetFirstPeriodAmount(); got != "1600.00" {
		t.Fatalf("第一期金額應為當期生效價的快照(1600.00),got %q", got)
	}

	// ① 訂閱列
	var (
		status, cycle, planCode string
		seats                   int
		trialEnds               sql.NullTime
	)
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT s.status, s.billing_cycle, p.code, s.seat_count, s.trial_ends_at
		  FROM platform.subscriptions s JOIN platform.plans p ON p.id = s.plan_id
		 WHERE s.id = $1`, subID).
		Scan(&status, &cycle, &planCode, &seats, &trialEnds); err != nil {
		t.Fatalf("讀訂閱: %v", err)
	}
	if status != "active" || cycle != "monthly" || planCode != "std" || seats != 3 {
		t.Fatalf("訂閱列不符: status=%q cycle=%q plan=%q seats=%d", status, cycle, planCode, seats)
	}
	if trialEnds.Valid {
		t.Fatalf("未指定試用不得留下 trial_ends_at: %v", trialEnds.Time)
	}

	// ② 第一期:金額、價格快照、期末(月底錨點)
	var (
		periodNo, periodSeats  int
		unitCents, seatCents   int64
		amountCents            int64
		periodStatus, currency string
		periodStart, periodEnd time.Time
		periodPlanID           int64
	)
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT period_no, (unit_price*100)::bigint, (seat_price*100)::bigint, seat_count,
		       (amount*100)::bigint, currency, status, period_start, period_end, plan_id
		  FROM platform.subscription_periods WHERE subscription_id = $1`, subID).
		Scan(&periodNo, &unitCents, &seatCents, &periodSeats, &amountCents, &currency,
			&periodStatus, &periodStart, &periodEnd, &periodPlanID); err != nil {
		t.Fatalf("讀第一期: %v", err)
	}
	if periodNo != 1 || periodStatus != "open" {
		t.Fatalf("應恰好一筆第 1 期且為 open: no=%d status=%q", periodNo, periodStatus)
	}
	if unitCents != 100000 || seatCents != 20000 || periodSeats != 3 || amountCents != 160000 ||
		currency != "TWD" {
		t.Fatalf("價格快照不符(1000.00／200.00／3 席／1600.00／TWD): unit=%d seat=%d seats=%d amount=%d cur=%q",
			unitCents, seatCents, periodSeats, amountCents, currency)
	}
	var planID int64
	if err := rig.admin.QueryRowContext(ctx,
		`SELECT id FROM platform.plans WHERE code = 'std'`).Scan(&planID); err != nil {
		t.Fatalf("讀方案 id: %v", err)
	}
	if periodPlanID != planID {
		t.Fatalf("期別的快照方案應為 std(%d),got %d", planID, periodPlanID)
	}
	if periodStart.Before(before) || periodStart.After(after) {
		t.Fatalf("第一期應自現在起算: %v（呼叫期間 %v..%v）", periodStart, before, after)
	}
	if want := nextMonthSameDayIn(periodStart); !periodEnd.Equal(want) {
		t.Fatalf("月繳的期末應為下月同日（月底夾擠）: got %v want %v", periodEnd, want)
	}

	// ③ 事件(payload 真的進 jsonb;consumer 不得為了補欄位再查 DB)
	var events int
	var payloadCompany, payloadSub, payloadReason string
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT count(*), COALESCE(max(payload->>'company_id'),''),
		       COALESCE(max(payload->>'subscription_id'),''), COALESCE(max(payload->>'reason'),'')
		  FROM platform.events
		 WHERE event_type = 'subscription.created' AND aggregate_id = $1`, subID).
		Scan(&events, &payloadCompany, &payloadSub, &payloadReason); err != nil {
		t.Fatalf("查事件: %v", err)
	}
	if events != 1 || payloadCompany != itoa(company) || payloadSub != subID ||
		payloadReason != "客戶簽約開通" {
		t.Fatalf("subscription.created 不符: n=%d company=%q sub=%q reason=%q",
			events, payloadCompany, payloadSub, payloadReason)
	}

	// ④ 恰一筆平台稽核(差量)
	if got := rig.auditCount(t, "subscription.create"); got != auditsBefore+1 {
		t.Fatalf("開通必須恰寫一筆稽核: %d → %d", auditsBefore, got)
	}
	opID, reason, ttype, extra := rig.lastAudit(t, "subscription.create")
	if opID != rig.seed.operatorID || reason != "客戶簽約開通" {
		t.Fatalf("稽核的 actor／原因錯誤: op=%d reason=%q", opID, reason)
	}
	if ttype != "subscription" || extra != subID+"|ops-a@example.com" {
		t.Fatalf("稽核的目標／operator email 錯誤: %s %s", ttype, extra)
	}
	rig.assertCacheInvalidated(t, company) // 開通前是「沒有訂閱列」(不施加限制)→ 必須失效

	// ⑤ 同一公司的第二次開通:唯一鍵擋下,且不得留下任何痕跡(訂閱／期別／事件／稽核都不變)。
	// 方案與週期刻意都與第一次相同:讓「被擋下的原因」只可能是唯一鍵(方案 pro 沒有年繳價目,
	// 用 pro+yearly 會先停在缺價目,那驗的是另一件事)。
	_, err = rig.svc.CreateSubscription(rig.ctx, connect.NewRequest(&platformv1.CreateSubscriptionRequest{
		CompanyId: itoa(company), PlanCode: "std", BillingCycle: "monthly", SeatCount: 9,
		Reason: "誤按第二次"}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("重複開通應 AlreadyExists,got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "SYS-2001" {
		t.Fatalf("必須是註冊碼 SYS-2001,got %q", got)
	}
	var subs, periods, audits int
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT (SELECT count(*) FROM platform.subscriptions WHERE company_id = $1),
		       (SELECT count(*) FROM platform.subscription_periods WHERE subscription_id = $2),
		       (SELECT count(*) FROM platform.audit_logs WHERE action = 'subscription.create')`,
		company, subID).Scan(&subs, &periods, &audits); err != nil {
		t.Fatalf("查計數: %v", err)
	}
	if subs != 1 || periods != 1 || audits != auditsBefore+1 {
		t.Fatalf("被拒絕的第二次開通不得留下痕跡: subs=%d periods=%d audits=%d", subs, periods, audits)
	}

	// ⑥ 排程不重複開期:剛開通的第一期期末還有一個月,不得被提前窗選中而開出第二期。
	//    (排程的掃描會處理 seed 的其他租戶,故這裡斷言的是**這家公司的期別數**與 PeriodsOpened。)
	summary := runCronOnce(t, rig, time.Now())
	if summary.PeriodsOpened != 0 {
		t.Fatalf("沒有訂閱落在提前窗內,不得開任何期別: %+v", summary)
	}
	var periodsAfterCron int
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT count(*) FROM platform.subscription_periods WHERE subscription_id = $1`, subID).
		Scan(&periodsAfterCron); err != nil {
		t.Fatalf("查期別數: %v", err)
	}
	if periodsAfterCron != 1 {
		t.Fatalf("排程不得為剛開通的租戶開第二期: %d 筆", periodsAfterCron)
	}
	if err := rig.admin.QueryRowContext(ctx,
		`SELECT status FROM platform.subscriptions WHERE id = $1`, subID).Scan(&status); err != nil {
		t.Fatalf("查訂閱狀態: %v", err)
	}
	if status != "active" {
		t.Fatalf("期末未到不得改狀態,got %q", status)
	}

	// ⑦ 收款:開通前是 PLAT-3001,現在必須成功(DoD 的第一跳 → 第二跳真的閉環)
	paid, err := rig.svc.RecordPayment(rig.ctx, connect.NewRequest(&platformv1.RecordPaymentRequest{
		CompanyId: itoa(company), PeriodNo: 1, Provider: "manual", Reason: "匯款入帳"}))
	if err != nil {
		t.Fatalf("開通後收款: %v", err)
	}
	if paid.Msg.GetPeriodNo() != 1 || paid.Msg.GetStatus() != "paid" {
		t.Fatalf("收款回應不符: %+v", paid.Msg)
	}
	if got := loadPeriodCents(t, rig, company, 1); got != 160000 {
		t.Fatalf("期別金額不得被輸入金額改動: %d", got)
	}
	if err := rig.admin.QueryRowContext(ctx,
		`SELECT status FROM platform.subscription_periods WHERE subscription_id = $1`, subID).
		Scan(&periodStatus); err != nil {
		t.Fatalf("查期別狀態: %v", err)
	}
	if periodStatus != "paid" {
		t.Fatalf("收款後期別應為 paid,got %q", periodStatus)
	}

	// ⑧ 只有一份**已取消**合約的租戶:可以再開一份新合約(00029 的部分唯一索引
	//    `WHERE status <> 'cancelled'` 不擋已取消者;要再服務是新合約,不是把舊的復活)。
	//    這一步同時釘住 ON CONFLICT 的推斷條件真的對上那個部分索引(條件寫錯會直接報
	//    「no unique or exclusion constraint matching the ON CONFLICT specification」)。
	renew := int(rig.seed.cancelledID)
	var oldSubID int64
	if err := rig.admin.QueryRowContext(ctx,
		`SELECT id FROM platform.subscriptions WHERE company_id = $1`, renew).Scan(&oldSubID); err != nil {
		t.Fatalf("讀舊合約: %v", err)
	}
	again, err := rig.svc.CreateSubscription(rig.ctx, connect.NewRequest(&platformv1.CreateSubscriptionRequest{
		CompanyId: itoa(renew), PlanCode: "std", BillingCycle: "monthly", SeatCount: 2,
		Reason: "重新簽約"}))
	if err != nil {
		t.Fatalf("已取消的合約不得擋住新合約: %v", err)
	}
	if again.Msg.GetSubscriptionId() == itoa(int(oldSubID)) {
		t.Fatalf("必須是**新**合約(不得復活舊的 %d)", oldSubID)
	}
	if got := again.Msg.GetFirstPeriodAmount(); got != "1400.00" { // 1000.00 ＋ 2 × 200.00
		t.Fatalf("新合約的第一期金額應取當期生效價(1400.00),got %q", got)
	}
	var cancelled, live int
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT count(*) FILTER (WHERE status = 'cancelled'), count(*) FILTER (WHERE status <> 'cancelled')
		  FROM platform.subscriptions WHERE company_id = $1`, renew).Scan(&cancelled, &live); err != nil {
		t.Fatalf("查合約: %v", err)
	}
	if cancelled != 1 || live != 1 {
		t.Fatalf("舊合約必須保持 cancelled 且只多出一筆新的: cancelled=%d live=%d", cancelled, live)
	}
}

// TestIntegrationExpireTrial 串起**試用到期**（I-1）：開通（trialing）→ 排程把試用已到期的訂閱
// 轉 past_due → 寬限過後由既有的 SuspendOverdue 接手。沒有這一步，開通就等於無上界的免費放行：
// 判定層把 trialing 當可用、EnsureNextPeriod 每期照開未付期別、而 MarkPastDue 只掃 active。
//
// 三件事在假 store 上驗不到，故用真容器:①部分唯一索引與真 SQL 的掃描謂詞一起動;
// ②「試用未到 → 不動」與「到期 → 轉一次」的分界是**時間比較**（`trial_ends_at < $1`）;
// ③寬限期的日曆運算（now + 7 天）真的落地。
//
// 執行:task test:integration -- -count=1 -run TestIntegrationExpireTrial -v
func TestIntegrationExpireTrial(t *testing.T) {
	rig := newWriteRig(t)
	company := int(rig.seed.noneID)
	ctx := t.Context()

	// trial_ends_at 必須是未來（開通端擋過去），故取 +45 秒，再把排程的 now 撥到它前後。
	// 窗要夠寬：這一段是**唯一**看牆鐘的地方（`CreateSubscription` 要求 trial_ends_at > now），
	// 2 秒的窗在 CI 的冷啟動與負載下會偶發「一開通就過期」（SYS-1001）；排程判定不受影響 ——
	// 它的 now 一律由呼叫端顯式傳入（trialEnds ± 1 秒），與真實時間無關。
	trialEnds := time.Now().UTC().Add(45 * time.Second).Truncate(time.Second)
	resp, err := rig.svc.CreateSubscription(rig.ctx, connect.NewRequest(&platformv1.CreateSubscriptionRequest{
		CompanyId: itoa(company), PlanCode: "std", BillingCycle: "monthly", SeatCount: 2,
		TrialEndsAt: trialEnds.Format(time.RFC3339), Reason: "POC 試用"}))
	if err != nil {
		t.Fatalf("開通試用: %v", err)
	}
	if resp.Msg.GetStatus() != "trialing" || resp.Msg.GetTrialEndsAt() != trialEnds.Format(time.RFC3339) {
		t.Fatalf("有未來試用應為 trialing，got %+v", resp.Msg)
	}
	subID := resp.Msg.GetSubscriptionId()

	// ① 試用還差一秒 → 排程一個字都不動。
	if got := runCronOnce(t, rig, trialEnds.Add(-time.Second)); got.TrialsExpired != 0 {
		t.Fatalf("試用未到不得轉移，got %+v", got)
	}
	if status := subscriptionStatus(t, rig, subID); status != "trialing" {
		t.Fatalf("試用未到的訂閱不得被動到，got %q", status)
	}

	// ② 試用到期 → past_due ＋ 寬限期（now + 7 天）＋ 恰一筆 subscription.trial_ended。
	after := trialEnds.Add(time.Second)
	if got := runCronOnce(t, rig, after); got.TrialsExpired != 1 {
		t.Fatalf("試用到期應轉 1 筆，got %+v", got)
	}
	var (
		status     string
		graceUntil sql.NullTime
		trialKept  sql.NullTime
	)
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT status, grace_until, trial_ends_at FROM platform.subscriptions WHERE id = $1`, subID).
		Scan(&status, &graceUntil, &trialKept); err != nil {
		t.Fatalf("讀訂閱: %v", err)
	}
	if status != "past_due" {
		t.Fatalf("試用到期應為 past_due，got %q", status)
	}
	if !graceUntil.Valid || !graceUntil.Time.Equal(after.AddDate(0, 0, 7)) {
		t.Fatalf("寬限期應為 now+7 天(%v)，got %v", after.AddDate(0, 0, 7), graceUntil.Time)
	}
	if !trialKept.Valid || !trialKept.Time.Equal(trialEnds) {
		t.Fatalf("試用到期日必須保留(trial_ends_at 是唯一的事實): %v", trialKept.Time)
	}
	var trialEvents int
	var payloadCompany, payloadReason string
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT count(*), COALESCE(max(payload->>'company_id'),''), COALESCE(max(payload->>'reason'),'')
		  FROM platform.events WHERE event_type = 'subscription.trial_ended' AND aggregate_id = $1`,
		subID).Scan(&trialEvents, &payloadCompany, &payloadReason); err != nil {
		t.Fatalf("查事件: %v", err)
	}
	if trialEvents != 1 || payloadCompany != itoa(company) || payloadReason != "trial_expired" {
		t.Fatalf("subscription.trial_ended 不符: n=%d company=%q reason=%q",
			trialEvents, payloadCompany, payloadReason)
	}
	// 轉走之後就不在「服務中」→ 不得被開新的一期。
	if n := countPeriods(t, rig, subID); n != 1 {
		t.Fatalf("試用到期不得被開新期，got %d 期", n)
	}

	// ③ 再跑一趟同一時間 → 不轉也不再發事件（掃描謂詞自己冪等）。
	if got := runCronOnce(t, rig, after); got.TrialsExpired != 0 {
		t.Fatalf("重跑不得再轉移，got %+v", got)
	}
	if err := rig.admin.QueryRowContext(ctx, `
		SELECT count(*) FROM platform.events
		 WHERE event_type = 'subscription.trial_ended' AND aggregate_id = $1`, subID).
		Scan(&trialEvents); err != nil {
		t.Fatalf("重跑後查事件: %v", err)
	}
	if trialEvents != 1 {
		t.Fatalf("重跑不得重複發事件，got %d 筆", trialEvents)
	}

	// ④ 寬限過後 → 既有的 SuspendOverdue 接手（同一張狀態機，不是第二套邏輯）。
	//    產品的凍結（companies.status）由 consumer 負責，真 consumer 的那條路徑由 cron 的整合
	//    測試把關；這裡驗平台域的轉移真的發生。
	runCronOnce(t, rig, after.AddDate(0, 0, 8))
	if status := subscriptionStatus(t, rig, subID); status != "suspended" {
		t.Fatalf("寬限過後應由 SuspendOverdue 接手轉 suspended，got %q", status)
	}
}

// subscriptionStatus 讀某訂閱的狀態；countPeriods 數它的期別數。
func subscriptionStatus(t *testing.T, rig *writeRig, subID string) string {
	t.Helper()
	var status string
	if err := rig.admin.QueryRowContext(t.Context(),
		`SELECT status FROM platform.subscriptions WHERE id = $1`, subID).Scan(&status); err != nil {
		t.Fatalf("讀訂閱狀態(%s): %v", subID, err)
	}
	return status
}

func countPeriods(t *testing.T, rig *writeRig, subID string) int {
	t.Helper()
	var n int
	if err := rig.admin.QueryRowContext(t.Context(),
		`SELECT count(*) FROM platform.subscription_periods WHERE subscription_id = $1`, subID).Scan(&n); err != nil {
		t.Fatalf("數期別(%s): %v", subID, err)
	}
	return n
}
