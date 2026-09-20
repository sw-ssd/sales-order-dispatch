package services

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
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
)

// platformStore 為本服務需要的平台存取(consumer 端定義):實作是 postgres.Admin(admin 連線),
// 單元測試用假實作。只收查詢 ＋ 平台稽核寫入。
type platformStore interface {
	ListTenants(ctx context.Context, keyword, status string, page, pageSize int32) ([]TenantRow, int, error)
	GetTenant(ctx context.Context, companyID string) (*TenantRow, []TenantOverrideRow, error)
	ListPlans(ctx context.Context) ([]PlanRow, error)
	PlanEntitlements(ctx context.Context, planCode string) ([]FeatureEntitlementRow, []FeatureRow, error)
	ListPlatformAudit(ctx context.Context, targetType, targetID string, page, pageSize int32) ([]PlatformAuditRow, int, error)
	RecordAudit(ctx context.Context, operatorID int64, action, targetType, targetID, reason string,
		before, after map[string]any) error
}

// PlatformAdminService 為平台工具的唯一後端入口(D38):跨租戶視圖一律走本服務與 admin 連線
// (spec §6.4),不得以 scope=all 掃業務表。五個 RPC 全部**唯讀**——平台端的寫入(改方案、
// 停用租戶、收款)不在 v1 的 proto 裡。
//
// 授權是**兩層**:掛載處的 operatorauth.Interceptor 擋一次,每個方法第一行再擋一次。
// 第二層不是重複:服務是普通的 Go 方法,可被別的掛載路徑、CLI 或未來的批次路徑直接呼叫,
// 只靠 interceptor 等於把邊界交給呼叫端決定。
type PlatformAdminService struct {
	st platformStore
	platformv1connect.UnimplementedPlatformAdminServiceHandler
}

// NewPlatformAdminService 建立 PlatformAdminService。
func NewPlatformAdminService(st platformStore) *PlatformAdminService {
	return &PlatformAdminService{st: st}
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
func RegisterPlatformAdminService(mux *http.ServeMux, st platformStore, op *operatorauth.Service) {
	path, handler := platformv1connect.NewPlatformAdminServiceHandler(
		NewPlatformAdminService(st), connect.WithInterceptors(op.Interceptor()))
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

// recordPlatformAudit 寫入 platform.audit_logs(S9:actor 為 operator,不 FK 租戶 users)。
// 與業務稽核(audit.Record)不同表、不同交易語意:平台操作本身即為稽核對象。
//
// v1 的 proto 只有讀取 RPC,故本方法是**為後續寫入路徑(新增／停用 operator、Plan C 的訂閱
// 寫入)先備好的唯一入口**:那些路徑一律走這裡,不得各自 INSERT 稽核(否則「誰改的」會有
// 兩份實作、只會有一份寫對)。
func (s *PlatformAdminService) recordPlatformAudit(ctx context.Context, id operatorauth.Identity,
	action, targetType, targetID, reason string, before, after map[string]any) error {
	return s.st.RecordAudit(ctx, id.OperatorID, action, targetType, targetID, reason, before, after)
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
