package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
)

// 平台 admin 查詢的資料列型別定義在 platform/store(服務與 SQL 實作共用同一組 DTO):
// 這裡以別名沿用,讓服務層讀起來不必到處寫套件前綴。
type (
	TenantRow             = platformstore.TenantRow
	TenantOverrideRow     = platformstore.TenantOverrideRow
	PlanRow               = platformstore.PlanRow
	PlanPriceRow          = platformstore.PlanPriceRow
	FeatureRow            = platformstore.Feature
	FeatureEntitlementRow = platformstore.Entitlement
	PlatformAuditRow      = platformstore.PlatformAuditRow
	ReceivableRow         = platformstore.ReceivableRow
)

// platformStore 為本服務需要的平台存取(consumer 端定義):實作是 postgres.Admin(admin 連線),
// 單元測試用假實作。查詢 ＋ 平台稽核寫入 ＋ T9 的 console 寫入。
//
// 寫入方法的共同形狀:**接受 *sql.Tx**,由本服務的 writeTx 開交易(資料與稽核同一個 commit)。
// store 不代開交易也不代 commit —— 這是平台寫入的交易(platform schema 不套 RLS、不經租戶連線),
// 與租戶請求的 interceptor 交易是兩回事。
type platformStore interface {
	ListTenants(ctx context.Context, keyword, status string, page, pageSize int32) ([]TenantRow, int, error)
	GetTenant(ctx context.Context, companyID string) (*TenantRow, []TenantOverrideRow, error)
	ListPlans(ctx context.Context) ([]PlanRow, error)
	PlanEntitlements(ctx context.Context, planCode string) ([]FeatureEntitlementRow, []FeatureRow, error)
	ListPlatformAudit(ctx context.Context, targetType, targetID string, page, pageSize int32) ([]PlatformAuditRow, int, error)
	ListReceivables(ctx context.Context, page, pageSize int32) ([]ReceivableRow, int, error)

	// --- T9 的寫入(資料與稽核同一個交易) ---
	//
	// **沒有非交易式的稽核入口**:稽核若能在交易外單獨寫,就會出現「稽核說改了、其實沒動」與
	// 「動了、卻沒有稽核」兩種半成品(v1 曾有 recordPlatformAudit,是唯讀時期的暫置物,T9 移除)。
	WithTx(ctx context.Context, fn func(*sql.Tx) error) error
	RecordAuditTx(ctx context.Context, tx *sql.Tx, operatorID int64, action, targetType, targetID,
		reason string, before, after []byte) error
	SetTenantOverrideTx(ctx context.Context, tx *sql.Tx, in platformstore.TenantOverrideInput) (int64, error)
	RevokeTenantOverrideTx(ctx context.Context, tx *sql.Tx, overrideID int64) (platformstore.OverrideRef, error)
	UpsertPlanPriceTx(ctx context.Context, tx *sql.Tx, in platformstore.PlanPriceInput) error
	SetPlanEntitlementTx(ctx context.Context, tx *sql.Tx, planCode, featureCode string, enabled bool,
		limit *int64) error
	CreateOperatorTx(ctx context.Context, tx *sql.Tx, email, name, role string) (int64, error)
	DisableOperatorTx(ctx context.Context, tx *sql.Tx, operatorID int64) error
	Settings(ctx context.Context) (map[string]string, error)
	UpsertSettingTx(ctx context.Context, tx *sql.Tx, key, value string) error
}

// PlatformAdminService 為平台工具的唯一後端入口(D38):跨租戶視圖一律走本服務與 admin 連線
// (spec §6.4),不得以 scope=all 掃業務表。唯讀的投影(t5 支)之外,本服務同時是**平台寫入的
// 唯一入口**(T9):每個寫入 RPC 都帶 reason(平台稽核必填)、以真實 operator 為 actor、
// 資料與稽核落在同一個交易。
//
// 授權是**兩層**:掛載處的 operatorauth.Interceptor 擋一次,每個方法第一行再擋一次。
// 第二層不是重複:服務是普通的 Go 方法,可被別的掛載路徑、CLI 或未來的批次路徑直接呼叫,
// 只靠 interceptor 等於把邊界交給呼叫端決定。
//
// 依賴的邊界(刻意分開,各有各的理由):
//   - st:平台域的讀寫(方案／例外／operator／settings／稽核);
//   - billing:訂閱與帳務的**狀態機**。改訂閱狀態一律經它(收款、席位、改方案、取消),
//     服務層不得自己 UPDATE subscriptions —— 第二個入口就是第二份狀態機;
//   - cache:權益快取。寫入路徑**顯式失效**(方案／價目 → 全量,其餘 → 該租戶),
//     TTL 只是漏掉失效時的收斂上界;
//   - counters:業務域的用量(席位 = 未停用帳號數)。平台域不認得業務 schema,故由外注入。
type PlatformAdminService struct {
	st       platformStore
	billing  *billing.Billing
	cache    entitlements.Cache
	counters entitlements.Counter
	platformv1connect.UnimplementedPlatformAdminServiceHandler
}

// NewPlatformAdminService 建立 PlatformAdminService。
//
// billing 與 counters 都是**必填**:billing 是訂閱狀態的唯一入口(少了它,寫入 RPC 只能
// 繞過狀態機),counters 是「席位不得低於使用中席次」的來源(少了它,那個守衛只能是空的)。
// cache 可為 nil(沒有快取就等於沒有東西要失效,見 entitlements.Invalidate)。
func NewPlatformAdminService(st platformStore, b *billing.Billing, cache entitlements.Cache,
	counters entitlements.Counter) *PlatformAdminService {
	return &PlatformAdminService{st: st, billing: b, cache: cache, counters: counters}
}

// RegisterPlatformAdminService 把平台工具 RPC 掛到 mux(自然路徑 /platform.v1.PlatformAdminService/…)
// 並加上 operator interceptor(服務層另有第二層檢查)。
//
// **呼叫端必須把 mux 掛在字面 "/platform/" 之下**(例:s.router.Mount("/platform",
// http.StripPrefix("/platform", mux)))——operator cookie 的 Path=/platform,而 RFC 6265 的
// path-match 是逐段前綴:瀏覽器路徑若是 /platform.v1.…(未被涵蓋的第一個字元是 "."),cookie
// 根本不會送出,於是「已登入」的 operator 每個請求都拿到 AUTH-4001。反過來把 cookie 的 Path
// 放寬成 "/" 也不行:那會讓 operator cookie 跟著送往租戶 API。兩者只有一種相容的掛法,故掛載點
// 以此簽章說明(實際掛載見 server.mountPlatformAuth)。
func RegisterPlatformAdminService(mux *http.ServeMux, st platformStore, b *billing.Billing,
	cache entitlements.Cache, counters entitlements.Counter, op *operatorauth.Service) {
	path, handler := platformv1connect.NewPlatformAdminServiceHandler(
		NewPlatformAdminService(st, b, cache, counters),
		// requestid 必須**在最前面**:trace_id 由它在回應邊界補進 ErrorInfo(requestid.stampTraceID),
		// 內層 interceptor 與服務層產生的錯誤才帶得到。少了它,console 收到的 ErrorInfo 只有
		// code/message —— 客服回報時沒有任何線索能對上 server log(spec §2.2 的三種接觸面
		// 都要求含 trace_id),而這個缺失在測試只驗 connect 碼時看不出來。
		connect.WithInterceptors(requestid.Interceptor(), op.Interceptor()))
	mux.Handle(path, handler)
}

// ListTenants 分頁列出跨租戶概況(keyword 篩公司名稱／識別碼,status 篩訂閱狀態)。
func (s *PlatformAdminService) ListTenants(ctx context.Context, req *connect.Request[platformv1.ListTenantsRequest]) (*connect.Response[platformv1.ListTenantsResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	rows, total, err := s.st.ListTenants(ctx,
		strings.TrimSpace(req.Msg.GetKeyword()), strings.TrimSpace(req.Msg.GetStatus()),
		int32(page), int32(pageSize))
	if err != nil {
		return nil, platformError(err)
	}
	tenants := make([]*platformv1.TenantSummary, 0, len(rows))
	for _, r := range rows {
		tenants = append(tenants, tenantToProto(r))
	}
	return connect.NewResponse(&platformv1.ListTenantsResponse{
		Tenants:    tenants,
		Pagination: platformPagination(page, pageSize, total),
	}), nil
}

// GetTenant 取單一租戶的概況與其未撤銷的例外(簽約承諾)。
func (s *PlatformAdminService) GetTenant(ctx context.Context, req *connect.Request[platformv1.GetTenantRequest]) (*connect.Response[platformv1.GetTenantResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	rawID := strings.TrimSpace(req.Msg.GetCompanyId())
	if _, err := platformCompanyID(rawID); err != nil {
		return nil, err
	}
	row, overrides, err := s.st.GetTenant(ctx, rawID)
	if err != nil {
		return nil, platformError(err)
	}
	out := make([]*platformv1.TenantOverride, 0, len(overrides))
	for _, o := range overrides {
		out = append(out, tenantOverrideToProto(o))
	}
	return connect.NewResponse(&platformv1.GetTenantResponse{
		Tenant:    tenantToProto(*row),
		Overrides: out,
	}), nil
}

// ListPlans 回傳全部方案與各計費週期的現行價目(權益矩陣與報價單的來源)。
func (s *PlatformAdminService) ListPlans(ctx context.Context, req *connect.Request[platformv1.ListPlansRequest]) (*connect.Response[platformv1.ListPlansResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	rows, err := s.st.ListPlans(ctx)
	if err != nil {
		return nil, platformError(err)
	}
	plans := make([]*platformv1.Plan, 0, len(rows))
	for _, r := range rows {
		plans = append(plans, planToProto(r))
	}
	return connect.NewResponse(&platformv1.ListPlansResponse{Plans: plans}), nil
}

// GetPlanEntitlements 回傳某方案的權益與完整功能清單(矩陣要顯示未設定的格,故兩者都要)。
func (s *PlatformAdminService) GetPlanEntitlements(ctx context.Context, req *connect.Request[platformv1.GetPlanEntitlementsRequest]) (*connect.Response[platformv1.GetPlanEntitlementsResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	planCode := strings.TrimSpace(req.Msg.GetPlanCode())
	if planCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "plan_code"})
	}
	ents, features, err := s.st.PlanEntitlements(ctx, planCode)
	if err != nil {
		return nil, platformError(err)
	}
	entOut := make([]*platformv1.FeatureEntitlement, 0, len(ents))
	for _, e := range ents {
		entOut = append(entOut, featureEntitlementToProto(e))
	}
	featureOut := make([]*platformv1.Feature, 0, len(features))
	for _, f := range features {
		featureOut = append(featureOut, featureToProto(f))
	}
	return connect.NewResponse(&platformv1.GetPlanEntitlementsResponse{
		Entitlements: entOut,
		Features:     featureOut,
	}), nil
}

// ListPlatformAudit 分頁列出平台稽核(新到舊),可依 target_type／target_id 篩選。
func (s *PlatformAdminService) ListPlatformAudit(ctx context.Context, req *connect.Request[platformv1.ListPlatformAuditRequest]) (*connect.Response[platformv1.ListPlatformAuditResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	rows, total, err := s.st.ListPlatformAudit(ctx,
		strings.TrimSpace(req.Msg.GetTargetType()), strings.TrimSpace(req.Msg.GetTargetId()),
		int32(page), int32(pageSize))
	if err != nil {
		return nil, platformError(err)
	}
	entries := make([]*platformv1.PlatformAuditEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, platformAuditToProto(r))
	}
	return connect.NewResponse(&platformv1.ListPlatformAuditResponse{
		Entries:    entries,
		Pagination: platformPagination(page, pageSize, total),
	}), nil
}

// requireOperatorIdentity 為服務層的授權檢查:沒有 operator 身分一律 AUTH-4001。
//
// 租戶 session／JWT 與平台 token 互不通用(不同 cookie 名稱、不同 secret、不同 audience,T8),
// 故租戶身分在 ctx 裡也不會有 operatorauth.Identity —— 這裡只認 operatorauth 注入的身分,
// 不看(也不該看)租戶的 authz.Identity。
func requireOperatorIdentity(ctx context.Context) (operatorauth.Identity, error) {
	id, ok := operatorauth.IdentityFrom(ctx)
	if !ok {
		return operatorauth.Identity{}, errcode.AuthUnauthenticated.Error(nil)
	}
	return id, nil
}

// platformCompanyID 驗證 proto 的 company_id(bigint 的文字形):空字串或非數字 → SYS-1001。
//
// 不共用租戶端的 parseID:那裡回的是裸 connect 錯誤(歷史遺留),而本服務一律用註冊碼 ——
// 缺哪個欄位由 details 帶(比照 requireScope 的作法,訊息樣板不為個別呼叫點改動)。
func platformCompanyID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errcode.SysInvalidArgument.Error(map[string]string{"field": "company_id"})
	}
	return id, nil
}

// platformError 將 store 的錯誤映射為註冊碼:查無資料 → SYS-4002;其餘交給全服務層的
// toConnectError(未知錯誤 → SYS-9000,驅動層原文不外洩)。
func platformError(err error) error {
	if errors.Is(err, platformstore.ErrNotFound) {
		return errcode.SysNotFound.Error(nil)
	}
	return toConnectError(err)
}

// platformPagination 組出分頁結果。帶的是**正規化後**的 page／page_size(前端才算得出頁數;
// 回音未正規化的 0 會讓頁數變成 NaN)。
func platformPagination(page, pageSize, total int) *platformv1.PlatformPagination {
	return &platformv1.PlatformPagination{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Total:    int32(total),
	}
}

// formatTime 為可空的時間欄位:nil → 空字串(proto 的約定);有值一律 UTC 的 RFC3339 ——
// 同一個時間不得因伺服器時區不同而長得不一樣。
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// tenantToProto 將投影列轉為 proto(欄位一一對應;狀態 none 由 store 保證,不在此補值)。
func tenantToProto(r TenantRow) *platformv1.TenantSummary {
	return &platformv1.TenantSummary{
		CompanyId:          r.CompanyID,
		CompanyName:        r.CompanyName,
		PlanCode:           r.PlanCode,
		PlanName:           r.PlanName,
		SubscriptionStatus: r.Status,
		SeatCount:          r.SeatCount,
		CurrentPeriodEnd:   formatTime(r.CurrentPeriodEnd),
		Overdue:            r.Overdue,
	}
}

// tenantOverrideToProto 轉出例外;enabled／limit 的 NULL 以 *_set 表達(見 TenantOverrideRow)。
func tenantOverrideToProto(r TenantOverrideRow) *platformv1.TenantOverride {
	out := &platformv1.TenantOverride{
		Id:          r.ID,
		FeatureCode: r.FeatureCode,
		Reason:      r.Reason,
		Owner:       r.Owner,
		ExpiresAt:   formatTime(r.ExpiresAt),
	}
	if r.Enabled != nil {
		out.EnabledSet, out.Enabled = true, *r.Enabled
	}
	if r.Limit != nil {
		out.LimitSet, out.LimitValue = true, *r.Limit
	}
	return out
}

// planToProto 轉出方案與其現行價目。
func planToProto(r PlanRow) *platformv1.Plan {
	prices := make([]*platformv1.PlanPrice, 0, len(r.Prices))
	for _, p := range r.Prices {
		prices = append(prices, &platformv1.PlanPrice{
			BillingCycle:  p.BillingCycle,
			BasePrice:     p.BasePrice,
			SeatPrice:     p.SeatPrice,
			Currency:      p.Currency,
			EffectiveFrom: p.EffectiveFrom.UTC().Format(time.RFC3339),
		})
	}
	return &platformv1.Plan{
		Id:        r.ID,
		Code:      r.Code,
		Name:      r.Name,
		Status:    r.Status,
		SortOrder: r.SortOrder,
		Prices:    prices,
	}
}

// featureEntitlementToProto 轉出方案權益;limit 的 NULL(不限)以 limit_set=false 表達
// (若是 0,前端會顯示「上限 0」)。
func featureEntitlementToProto(e FeatureEntitlementRow) *platformv1.FeatureEntitlement {
	out := &platformv1.FeatureEntitlement{FeatureCode: e.FeatureCode, Enabled: e.Enabled}
	if e.Limit != nil {
		out.LimitSet, out.LimitValue = true, *e.Limit
	}
	return out
}

// featureToProto 轉出功能定義。
func featureToProto(f FeatureRow) *platformv1.Feature {
	return &platformv1.Feature{Code: f.Code, Type: f.Type, Unit: f.Unit, Description: f.Description}
}

// platformAuditToProto 轉出稽核列。
func platformAuditToProto(r PlatformAuditRow) *platformv1.PlatformAuditEntry {
	return &platformv1.PlatformAuditEntry{
		Id:            r.ID,
		OperatorEmail: r.OperatorEmail,
		Action:        r.Action,
		TargetType:    r.TargetType,
		TargetId:      r.TargetID,
		Reason:        r.Reason,
		CreatedAt:     r.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ---------------------------------------------------------------------------
// T9:平台寫入 RPC
//
// 每個寫入 RPC 的三條共同契約(缺一不可,單元測試以表驅動逐條釘住):
//  ① 沒有 operator 身分 → AUTH-4001(不依賴掛載處的 interceptor);
//  ② reason 全空白 → SYS-1001(平台稽核必填:動錢與權限的操作必須留下「為什麼」);
//  ③ 一次寫入 = 一個交易(資料＋恰好一筆稽核),提交後才失效權益快取。
//
// 第 ③ 條的兩個方向都要成立才叫「同一交易」:資料寫了但稽核沒寫(事後查不出是誰),
// 與稽核寫了但資料沒寫(稽核說改了、其實沒動)。故稽核一律在**同一個 WithTx 的閉包內**,
// 且是最後一步 —— 前面任一步失敗即整份回滾,不會留下半成品。
// ---------------------------------------------------------------------------

// ListReceivables 列出未付的期別(含逾期),供 console 顯示與匯出 CSV。
//
// 平台自營公司(G5)由 store 的查詢排除(它對自己開出的期別不是應收帳款);「已逾期」不是
// 另一個狀態,而是同一列上 period_end 已過的事實 —— 前端據此標記,服務層不改語意。
func (s *PlatformAdminService) ListReceivables(ctx context.Context,
	req *connect.Request[platformv1.ListReceivablesRequest]) (*connect.Response[platformv1.ListReceivablesResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	rows, total, err := s.st.ListReceivables(ctx, int32(page), int32(pageSize))
	if err != nil {
		return nil, platformError(err)
	}
	out := make([]*platformv1.Receivable, 0, len(rows))
	for _, r := range rows {
		out = append(out, &platformv1.Receivable{
			CompanyId:   r.CompanyID,
			CompanyName: r.CompanyName,
			PlanCode:    r.PlanCode,
			PeriodNo:    r.PeriodNo,
			Amount:      r.Amount,
			PeriodEnd:   r.PeriodEnd.UTC().Format(time.RFC3339),
			Status:      r.Status,
		})
	}
	return connect.NewResponse(&platformv1.ListReceivablesResponse{
		Rows:       out,
		Pagination: platformPagination(page, pageSize, total),
	}), nil
}

// RecordPayment 為人工收款的營運入口;金流 webhook(日後)呼叫同一個 billing 方法。
//
// **不得**在此另寫一條入帳路徑:狀態機、期別金額驗證、事件與稽核都在 billing.RecordPayment 內
// (同一個交易),服務層只做「身分、參數驗證、快取失效」。金額空字串 = 採用期別快照金額
// (v1 不支援部分付款);ParseCents 不接受空字串(C-06),故空字串在此先轉成 0。
func (s *PlatformAdminService) RecordPayment(ctx context.Context,
	req *connect.Request[platformv1.RecordPaymentRequest]) (*connect.Response[platformv1.RecordPaymentResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := platformCompanyID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	var cents int64
	if amount := strings.TrimSpace(req.Msg.GetAmount()); amount != "" {
		if cents, err = money.ParseCents(amount); err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "amount"})
		}
	}
	period, err := s.billing.RecordPayment(ctx, billing.RecordPaymentInput{
		CompanyID:       int(companyID),
		PeriodNo:        int(req.Msg.GetPeriodNo()),
		AmountCents:     cents,
		Provider:        req.Msg.GetProvider(),
		ExternalRef:     req.Msg.GetExternalRef(),
		InvoiceNo:       req.Msg.GetInvoiceNo(),
		InvoiceStatus:   req.Msg.GetInvoiceStatus(),
		BuyerTaxID:      req.Msg.GetBuyerTaxId(),
		Carrier:         req.Msg.GetCarrier(),
		Note:            req.Msg.GetNote(),
		ActorOperatorID: id.OperatorID,
		Reason:          req.Msg.GetReason(),
	})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, int(companyID))
	return connect.NewResponse(&platformv1.RecordPaymentResponse{
		PeriodNo: int32(period.PeriodNo),
		Status:   period.Status,
	}), nil
}

// CreateSubscription 為**開通**的營運入口:建立訂閱與第一期(同一個交易:訂閱＋期別＋事件＋
// 稽核),v1 由營運開通(spec §9 的 Out:自助註冊與試用申請流程不在本版本)。
//
// **授權層級與 ChangePlan 相同(operator,不需要 admin)**:開通是日常營運,不是操作者治理
// (只有 CreateOperator／DisableOperator 要求 admin);首行仍走 requireOperatorIdentity。
//
// 服務層只做「身分、參數驗證、快取失效」:狀態機、價格快照、期別長度與稽核都在
// billing.CreateSubscription 內(唯一入口,不得在此另寫一條開通路徑)。
func (s *PlatformAdminService) CreateSubscription(ctx context.Context,
	req *connect.Request[platformv1.CreateSubscriptionRequest]) (*connect.Response[platformv1.CreateSubscriptionResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := platformCompanyID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	in := billing.CreateSubscriptionInput{
		CompanyID:       int(companyID),
		PlanCode:        strings.TrimSpace(req.Msg.GetPlanCode()),
		BillingCycle:    strings.TrimSpace(req.Msg.GetBillingCycle()),
		SeatCount:       int(req.Msg.GetSeatCount()),
		ActorOperatorID: id.OperatorID,
		Reason:          req.Msg.GetReason(),
	}
	if raw := strings.TrimSpace(req.Msg.GetTrialEndsAt()); raw != "" {
		trialEnds, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "trial_ends_at"})
		}
		in.TrialEnds = &trialEnds
	}
	created, err := s.billing.CreateSubscription(ctx, in)
	if err != nil {
		return nil, err
	}
	// 開通是判定層改變最大的一次寫入(開通前沒有訂閱列＝不施加限制)→ 提交後失效該租戶;
	// 方案權益屬 `entitlements`／`plans` 家族,console 那側一併失效(見 TenantDetailPage)。
	s.invalidate(ctx, int(companyID))
	return connect.NewResponse(&platformv1.CreateSubscriptionResponse{
		SubscriptionId:    strconv.FormatInt(created.SubscriptionID, 10),
		Status:            created.Status,
		PlanCode:          created.PlanCode,
		BillingCycle:      created.BillingCycle,
		SeatCount:         int32(created.SeatCount),
		TrialEndsAt:       formatTime(created.TrialEnds),
		FirstPeriodNo:     int32(created.FirstPeriod.PeriodNo),
		FirstPeriodEnd:    created.FirstPeriod.PeriodEnd.UTC().Format(time.RFC3339),
		FirstPeriodAmount: money.FormatCents(created.FirstPeriod.AmountCents),
	}), nil
}

// SetSeatCount 調整席位數(下一期生效)。不得低於**目前使用中的席次**:降席位到使用量以下會讓
// 既有帳號在下次判定時超額,而超額的症狀是「使用者突然不能建單」—— 那應該由營運先停用帳號。
//
// 守衛碼用 PLAT-5001(details:feature／used／limit),與配額守衛同一個碼:前端一律導向
// 「升級方案／停用帳號」的同一個畫面,不因入口不同而分流。
func (s *PlatformAdminService) SetSeatCount(ctx context.Context,
	req *connect.Request[platformv1.SetSeatCountRequest]) (*connect.Response[platformv1.SetSeatCountResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := platformCompanyID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	seats := req.Msg.GetSeatCount()
	if seats <= 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "seat_count"})
	}
	if s.counters == nil {
		// 沒有計數器等於無法判斷「是否低於使用量」→ 必須拒絕,不得當成 0 放行。
		return nil, errcode.SysInternal.Error(map[string]string{"reason": "未注入席位計數器"})
	}
	used, err := s.counters.Count(ctx, int(companyID), entitlements.LimitSeats)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	if used > int(seats) {
		return nil, errcode.PlatformLimitExceeded.Error(map[string]string{
			"feature": entitlements.LimitSeats,
			"used":    strconv.Itoa(used),
			"limit":   strconv.Itoa(int(seats)),
		})
	}
	if _, err := s.billing.SetSeatCount(ctx, billing.SetSeatCountInput{
		CompanyID:       int(companyID),
		SeatCount:       int(seats),
		ActorOperatorID: id.OperatorID,
		Reason:          req.Msg.GetReason(),
	}); err != nil {
		return nil, err
	}
	s.invalidate(ctx, int(companyID))
	return connect.NewResponse(&platformv1.SetSeatCountResponse{SeatCount: seats}), nil
}

// ChangePlan 改訂閱的方案(下一期生效,當期期別不動)。與現行方案相同時為 no-op(不寫稽核)。
func (s *PlatformAdminService) ChangePlan(ctx context.Context,
	req *connect.Request[platformv1.ChangePlanRequest]) (*connect.Response[platformv1.ChangePlanResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := platformCompanyID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	planCode := strings.TrimSpace(req.Msg.GetPlanCode())
	if planCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "plan_code"})
	}
	effectiveFrom, err := s.billing.ChangePlan(ctx, billing.ChangePlanInput{
		CompanyID:       int(companyID),
		PlanCode:        planCode,
		ActorOperatorID: id.OperatorID,
		Reason:          req.Msg.GetReason(),
	})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, int(companyID))
	return connect.NewResponse(&platformv1.ChangePlanResponse{
		PlanCode:      planCode,
		EffectiveFrom: effectiveFrom.UTC().Format(time.RFC3339),
	}), nil
}

// CancelSubscription 期末終止訂閱(v1 不支援立即終止:沒有按日比例計費就沒有「已收但不再服務」
// 的金額可言,那是退款流程)。期末前仍提供服務,期末後由排程與 consumer 接手凍結。
func (s *PlatformAdminService) CancelSubscription(ctx context.Context,
	req *connect.Request[platformv1.CancelSubscriptionRequest]) (*connect.Response[platformv1.CancelSubscriptionResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := platformCompanyID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	cancelled, err := s.billing.CancelSubscription(ctx, billing.CancelSubscriptionInput{
		CompanyID:       int(companyID),
		AtPeriodEnd:     req.Msg.GetAtPeriodEnd(),
		ActorOperatorID: id.OperatorID,
		Reason:          req.Msg.GetReason(),
	})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, int(companyID))
	return connect.NewResponse(&platformv1.CancelSubscriptionResponse{
		CancelledAt:  formatTime(&cancelled.CancelledAt), // 零值 = 本來就是 cancelled(no-op)
		ServiceUntil: cancelled.ServiceUntil.UTC().Format(time.RFC3339),
	}), nil
}

// billingSettingKeys 為可由介面調整的營運參數,順序即 console 的顯示順序。
//
// system_actor_user_id 刻意**不在此**(它由 seed 決定、是稽核主體而非營運參數):開放編輯它等於
// 讓 console 能改掉「排程與 consumer 的稽核主體」,那是權限提升的路徑。
var billingSettingKeys = []struct{ key, description string }{
	{"trial_days", "試用天數"},
	{"grace_days", "逾期寬限天數"},
	{"lead_days", "提前開立下一期的天數"},
}

// GetBillingSettings 讀出試用／寬限／提前天數(cron 讀同一張表,不硬編在程式裡)。
func (s *PlatformAdminService) GetBillingSettings(ctx context.Context,
	req *connect.Request[platformv1.GetBillingSettingsRequest]) (*connect.Response[platformv1.GetBillingSettingsResponse], error) {
	if _, err := requireOperatorIdentity(ctx); err != nil {
		return nil, err
	}
	all, err := s.st.Settings(ctx)
	if err != nil {
		return nil, platformError(err)
	}
	return connect.NewResponse(&platformv1.GetBillingSettingsResponse{
		Settings: billingSettings(all),
	}), nil
}

// UpdateBillingSettings 更新營運參數。
//
// 兩個驗證都不在 store 層:值一律非負整數(負的寬限天數等於當天就停用),鍵必須在允許清單內
// (打錯的鍵會被 upsert 成一個**永遠讀不到的新列**,而 console 顯示「已儲存」)。
//
// **不失效權益快取**:這些天數影響的是「新開的期別與新訂閱」(寬限與提前窗在建立當下就被寫進
// 訂閱／期別),判定快照裡沒有它們 —— 失效只是白打一趟 Valkey。日後若有參數進了判定,這裡必須
// 跟著改成失效(見 GetBillingSettings 的鍵清單與 entitlements.tenantState 的欄位)。
func (s *PlatformAdminService) UpdateBillingSettings(ctx context.Context,
	req *connect.Request[platformv1.UpdateBillingSettingsRequest]) (*connect.Response[platformv1.UpdateBillingSettingsResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	updates, err := validatedSettings(req.Msg.GetSettings())
	if err != nil {
		return nil, err
	}
	before, err := s.st.Settings(ctx)
	if err != nil {
		return nil, platformError(err)
	}
	err = s.writeTx(ctx, id, "settings.update", "settings", req.Msg.GetReason(),
		nil, settingsAudit(before, updates), func(tx *sql.Tx) (string, error) {
			for _, kv := range sortedSettings(updates) {
				if err := s.st.UpsertSettingTx(ctx, tx, kv.key, kv.value); err != nil {
					return "", platformWriteError(err)
				}
			}
			return "billing", nil
		})
	if err != nil {
		return nil, err
	}
	after, err := s.st.Settings(ctx)
	if err != nil {
		return nil, platformError(err)
	}
	return connect.NewResponse(&platformv1.UpdateBillingSettingsResponse{
		Settings: billingSettings(after),
	}), nil
}

// SetTenantOverride 新增一筆租戶例外(簽約承諾)。
//
// 「哪個維度要覆寫」用 *_set 表達(proto3 分不出「沒給」與「給了 false／0」):只看限額的例外
// 不得讓功能被關掉,故兩個維度各自可為 NULL。兩者都沒給的例外沒有任何效果,直接拒絕 ——
// 寫進去的話,console 上會出現一筆「有承諾但什麼都沒改」的紀錄。
func (s *PlatformAdminService) SetTenantOverride(ctx context.Context,
	req *connect.Request[platformv1.SetTenantOverrideRequest]) (*connect.Response[platformv1.SetTenantOverrideResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := platformCompanyID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	featureCode := strings.TrimSpace(req.Msg.GetFeatureCode())
	if featureCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "feature_code"})
	}
	owner := strings.TrimSpace(req.Msg.GetOwner())
	if owner == "" {
		// 例外是**有人承諾的**(表上的 owner NOT NULL):沒有承諾者就沒有可追溯的責任。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "owner"})
	}
	if !req.Msg.GetEnabledSet() && !req.Msg.GetLimitSet() {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "enabled_set"})
	}
	in := platformstore.TenantOverrideInput{
		CompanyID:   companyID,
		FeatureCode: featureCode,
		Reason:      req.Msg.GetReason(),
		Owner:       owner,
		CreatedBy:   id.OperatorID,
	}
	if req.Msg.GetEnabledSet() {
		enabled := req.Msg.GetEnabled()
		in.Enabled = &enabled
	}
	if req.Msg.GetLimitSet() {
		limit := req.Msg.GetLimitValue()
		if limit < 0 {
			// 負的上限在判定層等於「任何用量都超額」→ 該租戶的功能被永久關掉,而 console 顯示
			// 「已設定限額」。負值不是「不限」(不限用 limit_set=false 表達),一律拒絕。
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "limit_value"})
		}
		in.Limit = &limit
	}
	if raw := strings.TrimSpace(req.Msg.GetExpiresAt()); raw != "" {
		expiresAt, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "expires_at"})
		}
		in.ExpiresAt = &expiresAt
	}

	var newID int64
	after := map[string]any{
		"feature_code": featureCode,
		"enabled":      in.Enabled,
		"limit_value":  in.Limit,
		"owner":        owner,
	}
	err = s.writeTx(ctx, id, "override.set", "company",
		req.Msg.GetReason(), nil, after, func(tx *sql.Tx) (string, error) {
			var err error
			if newID, err = s.st.SetTenantOverrideTx(ctx, tx, in); err != nil {
				return "", platformWriteError(err)
			}
			return strconv.FormatInt(companyID, 10), nil
		})
	if err != nil {
		return nil, err
	}
	// 例外直接改變判定結果(方案 ⊕ override)→ 提交後失效該租戶。
	s.invalidate(ctx, int(companyID))
	return connect.NewResponse(&platformv1.SetTenantOverrideResponse{
		Id: strconv.FormatInt(newID, 10),
	}), nil
}

// RevokeTenantOverride 撤銷一筆例外。撤銷是**單向**的狀態轉移(revoked_at 一旦寫下不再改),
// 重複撤銷 → SYS-4002:那不是成功,它會在稽核上留下一筆「撤銷了某個不存在的承諾」。
func (s *PlatformAdminService) RevokeTenantOverride(ctx context.Context,
	req *connect.Request[platformv1.RevokeTenantOverrideRequest]) (*connect.Response[platformv1.RevokeTenantOverrideResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	overrideID, err := strconv.ParseInt(strings.TrimSpace(req.Msg.GetOverrideId()), 10, 64)
	if err != nil || overrideID <= 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "override_id"})
	}

	var ref platformstore.OverrideRef
	err = s.writeTx(ctx, id, "override.revoke", "company", req.Msg.GetReason(),
		map[string]any{"override_id": req.Msg.GetOverrideId()}, map[string]any{"revoked": true},
		func(tx *sql.Tx) (string, error) {
			var err error
			if ref, err = s.st.RevokeTenantOverrideTx(ctx, tx, overrideID); err != nil {
				return "", platformWriteError(err)
			}
			// 稽核的目標是**公司**(撤銷是對該租戶的承諾收回):override 的內部 id 在 console 的
			// 稽核篩選(依 target_type／target_id)裡沒有意義,而 company_id 是營運查得到的東西。
			return strconv.FormatInt(ref.CompanyID, 10), nil
		})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, int(ref.CompanyID))
	return connect.NewResponse(&platformv1.RevokeTenantOverrideResponse{
		CompanyId:   strconv.FormatInt(ref.CompanyID, 10),
		FeatureCode: ref.FeatureCode,
	}), nil
}

// UpsertPlanPrice 寫入一筆方案價目(調價)。價目是**新增一列**(價格史),不改既有列。
//
// 價目影響每一個用該方案的租戶(判定快照含方案的權益,期別金額快照自此而來)→ 全量失效。
func (s *PlatformAdminService) UpsertPlanPrice(ctx context.Context,
	req *connect.Request[platformv1.UpsertPlanPriceRequest]) (*connect.Response[platformv1.UpsertPlanPriceResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	planCode := strings.TrimSpace(req.Msg.GetPlanCode())
	if planCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "plan_code"})
	}
	cycle := strings.TrimSpace(req.Msg.GetBillingCycle())
	if cycle != "monthly" && cycle != "yearly" {
		// 週期決定期別「加一個月或加一年」(G1):空字串或未知值一律拒絕,不得默默當月繳。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "billing_cycle"})
	}
	baseCents, err := money.ParseCents(req.Msg.GetBasePrice())
	if err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "base_price"})
	}
	seatCents, err := money.ParseCents(req.Msg.GetSeatPrice())
	if err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "seat_price"})
	}
	in := platformstore.PlanPriceInput{
		PlanCode:     planCode,
		BillingCycle: cycle,
		BaseCents:    baseCents,
		SeatCents:    seatCents,
		Currency:     strings.TrimSpace(req.Msg.GetCurrency()),
	}
	after := map[string]any{
		"plan_code":     planCode,
		"billing_cycle": cycle,
		"base_price":    money.FormatCents(baseCents),
		"seat_price":    money.FormatCents(seatCents),
		"currency":      in.Currency,
	}
	if err := s.writeTx(ctx, id, "plan.price_upsert", "plan", req.Msg.GetReason(),
		nil, after, func(tx *sql.Tx) (string, error) {
			if err := s.st.UpsertPlanPriceTx(ctx, tx, in); err != nil {
				return "", platformWriteError(err)
			}
			return planCode, nil
		}); err != nil {
		return nil, err
	}
	s.invalidateAll(ctx)
	return connect.NewResponse(&platformv1.UpsertPlanPriceResponse{}), nil
}

// SetPlanEntitlement 設定方案的一個功能權益(新增或覆寫)。limit_set=false 代表「不限額」(NULL),
// 與 0(上限 0)不同 —— proto3 分不出兩者,故以 *_set 表達。
func (s *PlatformAdminService) SetPlanEntitlement(ctx context.Context,
	req *connect.Request[platformv1.SetPlanEntitlementRequest]) (*connect.Response[platformv1.SetPlanEntitlementResponse], error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	planCode := strings.TrimSpace(req.Msg.GetPlanCode())
	featureCode := strings.TrimSpace(req.Msg.GetFeatureCode())
	if planCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "plan_code"})
	}
	if featureCode == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "feature_code"})
	}
	var limit *int64
	if req.Msg.GetLimitSet() {
		v := req.Msg.GetLimitValue()
		if v < 0 {
			// 同 SetTenantOverride:負的上限等於把這個功能對**所有用該方案的租戶**關掉。
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "limit_value"})
		}
		limit = &v
	}
	after := map[string]any{"enabled": req.Msg.GetEnabled()}
	if limit != nil {
		after["limit_value"] = *limit
	}
	if err := s.writeTx(ctx, id, "plan.entitlement_set", "plan", req.Msg.GetReason(),
		nil, after, func(tx *sql.Tx) (string, error) {
			if err := s.st.SetPlanEntitlementTx(ctx, tx, planCode, featureCode,
				req.Msg.GetEnabled(), limit); err != nil {
				return "", platformWriteError(err)
			}
			return planCode, nil
		}); err != nil {
		return nil, err
	}
	s.invalidateAll(ctx)
	return connect.NewResponse(&platformv1.SetPlanEntitlementResponse{}), nil
}

// CreateOperator 新增一個 operator(白名單)。email 已存在 → SYS-2001(不是 5xx:那是輸入問題)。
//
// role 只允許 operator／admin(表上的註解即這兩個值):打錯的 role 會寫進一個沒有人認得的白名單列。
// 不失效權益快取:白名單是**平台端**的身分,不進租戶的判定快照。
func (s *PlatformAdminService) CreateOperator(ctx context.Context,
	req *connect.Request[platformv1.CreateOperatorRequest]) (*connect.Response[platformv1.CreateOperatorResponse], error) {
	id, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	email := strings.TrimSpace(req.Msg.GetEmail())
	if !strings.Contains(email, "@") {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "email"})
	}
	role := strings.TrimSpace(req.Msg.GetRole())
	if role != "" && role != "operator" && role != "admin" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "role"})
	}
	name := strings.TrimSpace(req.Msg.GetName())

	var newID int64
	err = s.writeTx(ctx, id, "operator.create", "operator",
		req.Msg.GetReason(), nil, map[string]any{"email": email, "name": name, "role": role},
		func(tx *sql.Tx) (string, error) {
			var err error
			if newID, err = s.st.CreateOperatorTx(ctx, tx, email, name, role); err != nil {
				return "", platformWriteError(err)
			}
			return strconv.FormatInt(newID, 10), nil
		})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&platformv1.CreateOperatorResponse{Id: strconv.FormatInt(newID, 10)}), nil
}

// DisableOperator 停用一個 operator。停用即時生效(operatorauth 每次請求都查 status),
// 故不需要撤銷已簽發的 token;已停用者重複呼叫 → SYS-4002(不是成功)。
func (s *PlatformAdminService) DisableOperator(ctx context.Context,
	req *connect.Request[platformv1.DisableOperatorRequest]) (*connect.Response[platformv1.DisableOperatorResponse], error) {
	id, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := platformReason(req.Msg.GetReason()); err != nil {
		return nil, err
	}
	operatorID, err := strconv.ParseInt(strings.TrimSpace(req.Msg.GetOperatorId()), 10, 64)
	if err != nil || operatorID <= 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "operator_id"})
	}
	// 不得停用自己:停用是即時的(operatorauth 每次請求都查 status),自己停自己等於當場把自己
	// 鎖在門外,而「誰停的」會是那位已經進不來的人。要離職請由另一位 admin 操作。
	if operatorID == id.OperatorID {
		return nil, errcode.PlatformOperatorGovernance.Error(map[string]string{
			"reason": "不得停用自己（請由另一位 admin 操作）",
		})
	}
	err = s.writeTx(ctx, id, "operator.disable", "operator",
		req.Msg.GetReason(), map[string]any{"status": "active"}, map[string]any{"status": "disabled"},
		func(tx *sql.Tx) (string, error) {
			if err := s.st.DisableOperatorTx(ctx, tx, operatorID); err != nil {
				return "", platformWriteError(err)
			}
			return req.Msg.GetOperatorId(), nil
		})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&platformv1.DisableOperatorResponse{}), nil
}

// writeTx 執行一次平台寫入:**資料與稽核在同一個交易**,成功提交後才失效快取。
//
// 稽核的 before／after 在交易外先序列化:json.Marshal 失敗(SQL 參數不可能,但 map 值若夾帶
// 不可序列化的型別就會)必須在動任何資料**之前**就失敗,否則會出現「資料改了、稽核沒寫」。
//
// apply **回傳稽核的目標 id**:有些寫入的目標只有寫完才知道(新增 override 的新 id、撤銷時反查
// 到的公司、新增 operator 的 id)。回傳值而不是在閉包裡自己寫稽核,是為了讓「恰一筆稽核」與
// 「稽核在資料之後」這兩件事留在同一個地方 —— 由每個呼叫點自己寫就會有第二種順序。
func (s *PlatformAdminService) writeTx(ctx context.Context, id operatorauth.Identity, action,
	targetType, reason string, before, after map[string]any,
	apply func(*sql.Tx) (targetID string, err error)) error {
	// nil 的 before／after **不序列化**:Marshal(nil map) 會得到 JSON 的 "null"(4 個位元,
	// 一列非 NULL 的 jsonb),而稽核上「沒有這個資訊」與「值就是 null」是兩件事(見
	// store.BillingStore.RecordAuditTx)。空切片才會被 store 寫成 SQL NULL。
	rawBefore, err := marshalAuditImage(before)
	if err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	rawAfter, err := marshalAuditImage(after)
	if err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	return s.st.WithTx(ctx, func(tx *sql.Tx) error {
		targetID, err := apply(tx)
		if err != nil {
			return err
		}
		if err := s.st.RecordAuditTx(ctx, tx, id.OperatorID, action, targetType,
			targetID, reason, rawBefore, rawAfter); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
		return nil
	})
}

// marshalAuditImage 把稽核的 before／after 映像轉為可寫入的位元:nil → nil(＝SQL NULL)。
func marshalAuditImage(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// requireAdmin 為**操作者管理** RPC 的角色檢查(先驗身分再驗角色),不足即 SYS-4001。
//
// 為什麼只有這兩支 RPC 限 admin(v1 的角色矩陣,判斷見報告):`operator | admin` 兩個角色
// 在 spec 只有「白名單」的語意,而**建立／停用 operator 就是治理動作本身** —— 讓 operator
// 能做,等於任何一個 operator 都能憑空替自己或別人加一個 admin(提權),或停用其他 admin
// (破壞可管理性)。其餘寫入(收款、方案價目、權益、參數、override)是日常營運,維持 operator
// 即可;要收緊必須先有一份明訂的權限矩陣,那不是本任務能自行擴大的範圍。
//
// 角色來自 operatorauth 從**資料庫白名單列**讀出的身分(token 只是載體;每次請求都查 status
// 與 role),故改了 role 立刻生效。
func requireAdmin(ctx context.Context) (operatorauth.Identity, error) {
	id, err := requireOperatorIdentity(ctx)
	if err != nil {
		return operatorauth.Identity{}, err
	}
	if id.Role != "admin" {
		return operatorauth.Identity{}, errcode.SysPermissionDenied.Error(
			map[string]string{"required_role": "admin"})
	}
	return id, nil
}

// platformReason 為所有平台寫入的共同必填檢查:reason 全空白即拒絕(SYS-1001)。
func platformReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errcode.SysInvalidArgument.Error(map[string]string{"field": "reason"})
	}
	return nil
}

// platformWriteError 把平台寫入的錯誤映射為註冊碼:
//   - 已存在(唯一鍵)→ SYS-2001(輸入問題,不是系統故障);
//   - 不存在 → SYS-4002;
//   - 其餘 → toConnectError(未知錯誤 → SYS-9000,驅動層原文不外洩)。
//
// **不得**把這些失敗講成 PLAT-3002(收款衝突):那一碼描述的是帳務衝突(重複收款、金額不符),
// 而這裡可能是死鎖、序化失敗或連線中斷 —— 講成「收款衝突」會讓 operator 以為重試沒有意義而放棄;
// 反之,T9 的服務層也**不得**把 PLAT-3002 當成「不可重試」的依據(它由 billing 產生,涵蓋的
// 情境比「重複收款」更廣,見 billing.RecordPayment 的錯誤映射)。
func platformWriteError(err error) error {
	switch {
	case errors.Is(err, platformstore.ErrConflict):
		return errcode.SysConflict.Error(nil)
	case errors.Is(err, platformstore.ErrLastAdmin):
		// 治理不變式:要說出「該怎麼做」,不是「查無此人」。
		return errcode.PlatformOperatorGovernance.Error(map[string]string{
			"reason": "不得停用最後一位 admin（請先新增另一位 admin）",
		})
	case errors.Is(err, platformstore.ErrNotFound), errors.Is(err, sql.ErrNoRows):
		return errcode.SysNotFound.Error(nil)
	default:
		return toConnectError(err)
	}
}

// invalidate 失效該租戶的權益快取;失敗只記 log(TTL 保底,不讓**已成功的寫入**回錯誤)。
//
// 為什麼不讓 RPC 失敗:寫入已經提交,回錯誤會讓 console 顯示失敗、operator 再按一次 ——
// 而重試撞上的是冪等路徑,症狀變成「明明成功卻說失敗」。代價寫在明處:失效失敗時該租戶最長
// TTL(entitlementCacheTTL)內仍以舊權益放行。
func (s *PlatformAdminService) invalidate(ctx context.Context, companyID int) {
	if err := entitlements.Invalidate(ctx, s.cache, companyID); err != nil {
		log.Printf("platform: 權益快取失效失敗(company=%d): %v（該租戶最長 TTL 內仍讀舊權益）",
			companyID, err)
	}
}

// invalidateAll 失效**所有**租戶的權益快取(方案／價目的變更會影響每一個用到該方案的租戶)。
//
// 兩種失敗分開對待(見 entitlements.ErrScanUnsupported):
//   - **不支援**(行程內 MemoryCache,即 Valkey 缺席的降級部署):這是已知的部署形態,不是故障。
//     log 一行說明「最長 TTL 內仍讀舊權益」—— console 顯示的舊值會在一個 TTL 內收斂;
//   - **掃描／刪除失敗**(Valkey 故障):同樣只記 log(RPC 不失敗,理由同上),但字樣必須不同 ——
//     把故障說成「不支援」會讓真正該處理的事被當成設定問題。
func (s *PlatformAdminService) invalidateAll(ctx context.Context) {
	err := entitlements.InvalidateAll(ctx, s.cache)
	switch {
	case err == nil:
	case errors.Is(err, entitlements.ErrScanUnsupported):
		log.Printf("platform: 此快取不支援全量失效（方案／價目已改）: %v"+
			"（該租戶最長一個快取 TTL 內仍讀舊權益）", err)
	default:
		log.Printf("platform: 權益快取全量失效失敗（方案／價目已改）: %v", err)
	}
}

// billingSettings 依固定順序組出設定清單;缺鍵照樣列出(值為空字串)——
// 「這一項還沒設定」與「這一項不存在」在 console 上是兩件事(前者要顯示提示)。
func billingSettings(all map[string]string) []*platformv1.BillingSetting {
	out := make([]*platformv1.BillingSetting, 0, len(billingSettingKeys))
	for _, def := range billingSettingKeys {
		out = append(out, &platformv1.BillingSetting{
			Key:         def.key,
			Value:       all[def.key],
			Description: def.description,
		})
	}
	return out
}

// settingPair 為一組正規化後的參數(鍵＋值),排序後才寫入:同一次請求的寫入順序不該取決於
// map 的走訪順序(否則鎖的取得順序會在同一份資料上漂移)。
type settingPair struct{ key, value string }

// validatedSettings 驗證並正規化更新內容:鍵必須在允許清單內、值必須是非負整數。
//
// 值是 text 欄位,故驗證只能在寫入前做:「-1 天的寬限期」與「abc 天」在 DB 都存得進去,
// 而讀取端(cron 的 LoadParams)會把它當成設定錯誤讓整趟排程失敗。
func validatedSettings(in []*platformv1.BillingSetting) (map[string]string, error) {
	allowed := make(map[string]bool, len(billingSettingKeys))
	for _, def := range billingSettingKeys {
		allowed[def.key] = true
	}
	if len(in) == 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "settings"})
	}
	out := make(map[string]string, len(in))
	for _, kv := range in {
		key := strings.TrimSpace(kv.GetKey())
		if !allowed[key] {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "settings.key"})
		}
		value := strings.TrimSpace(kv.GetValue())
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "settings.value"})
		}
		out[key] = value
	}
	return out, nil
}

// sortedSettings 把更新內容排成穩定的順序(鍵遞增)。
func sortedSettings(updates map[string]string) []settingPair {
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]settingPair, 0, len(keys))
	for _, k := range keys {
		out = append(out, settingPair{key: k, value: updates[k]})
	}
	return out
}

// settingsAudit 組出「改了哪幾鍵、從什麼變成什麼」的稽核內容(只帶被改到的鍵:整份 settings
// 進稽核會讓每一筆紀錄都夾帶無關的參數,而稽核的價值在於一眼看出差異)。
func settingsAudit(before, updates map[string]string) map[string]any {
	out := make(map[string]any, len(updates))
	for k, v := range updates {
		out[k] = map[string]string{"before": before[k], "after": v}
	}
	return out
}
