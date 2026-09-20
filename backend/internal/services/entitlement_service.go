// 租戶端權益投影（spec §4.6）:租戶後台／App 看見「我的方案與用量」的唯一入口。
//
// **唯讀**（proto 只有一個 Get），且**只回自己公司**：請求沒有 company_id 參數——身分即範圍，
// 前端 disable 按鈕與顯示用量都只是展示，**不構成授權**（真正的擋在服務層守衛與 RLS）。
//
// 與 PlatformAdminService 剛好相反：那裡要求 operator 身分、拒絕租戶；這裡只認租戶身分，
// 營運身分（operator token 或無租戶範圍的平台層身分）一律讀不到任何租戶的權益。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
)

// TenantEntitlementService 為租戶端權益投影。
type TenantEntitlementService struct {
	// 直接吃判定層的具體型別：本服務要的是 Snapshot 的完整投影（含各 feature 用量），
	// 單元測試以假 store ＋ 假計數器組出真的 entitlements.Service，不需要記錄式假物件。
	ent *entitlements.Service
	platformv1connect.UnimplementedTenantEntitlementServiceHandler
}

// NewTenantEntitlementService 建立 TenantEntitlementService。
func NewTenantEntitlementService(ent *entitlements.Service) *TenantEntitlementService {
	return &TenantEntitlementService{ent: ent}
}

// RegisterTenantEntitlementService 把租戶端權益投影掛到**租戶 mux**（server 把它掛在
// "/api/v1" 之下），不是平台工具的 /platform/：那裡的 operator cookie 與 interceptor 是另一
// 組（T8／T9），租戶 session 到不了。
//
// dbtenant.Interceptor 不可省：計數器（services.entitlementCounter）在**請求交易**內計數，
// 少了它就只能退回 fallback client 而看不到 RLS scope，用量會算錯。
func RegisterTenantEntitlementService(mux *http.ServeMux, db *ent.Client, ent *entitlements.Service) {
	path, handler := platformv1connect.NewTenantEntitlementServiceHandler(
		NewTenantEntitlementService(ent),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// GetTenantEntitlements 回傳**呼叫者所屬公司**的方案、狀態、試用到期與各 feature 用量。
//
// 公司 id 一律取自身分：請求沒有可指定他公司的參數，平台層身分（無租戶範圍）也讀不到。
func (s *TenantEntitlementService) GetTenantEntitlements(ctx context.Context, _ *connect.Request[platformv1.GetTenantEntitlementsRequest]) (*connect.Response[platformv1.GetTenantEntitlementsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, err := strconv.Atoi(strings.TrimSpace(id.CompanyID))
	if err != nil || cid <= 0 {
		// 平台層身分（super／developer）與 operator 都沒有租戶範圍：fail-closed 拒絕，
		// 不得退化成「回全部」或「回空方案」——空 plan_code 的 200 會被前端讀成「未訂閱」。
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	snap, err := s.ent.Snapshot(ctx, cid)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(snapshotToProto(snap)), nil
}

// snapshotToProto 將判定層的投影轉為 proto。
//
// 「未帶計數器的 integer feature」不會走到這裡：那筆已由 entitlements.Snapshot 略過（見該處
// 的降級說明），故前端不會看到假的 0 用量；未帶 limit 的 feature（boolean／不限額）
// limit_set=false，前端才分得出「不限」與「上限 0」。
func snapshotToProto(s *entitlements.Snapshot) *platformv1.GetTenantEntitlementsResponse {
	resp := &platformv1.GetTenantEntitlementsResponse{
		PlanCode:    s.PlanCode,
		PlanName:    s.PlanName,
		Status:      s.Status,
		TrialEndsAt: formatTime(s.TrialEndsAt),
	}
	for _, u := range s.Usage {
		row := &platformv1.Usage{
			FeatureCode: u.FeatureCode,
			Enabled:     u.Enabled,
			Used:        int64(u.Used),
		}
		if u.Limit != nil {
			row.LimitSet = true
			row.LimitValue = *u.Limit
		}
		resp.Usage = append(resp.Usage, row)
	}
	return resp
}
