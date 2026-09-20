// AuditService 稽核日誌查詢(03 計畫 2.6.3, D27)。
// audit_logs 只由業務交易內同事務寫入(2.6.2, D18),本 service 僅提供查詢。
// 範圍:super/developer 全系統(可選 company_id);company_admin 僅自己公司;其餘角色不可查詢。
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	auditv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1/auditv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// validAuditActions 為 audit_logs.action 合法值(對齊 §5.2 / auditlog schema)。
var validAuditActions = map[string]bool{
	"create": true, "update": true, "delete": true, "login": true, "logout": true,
	"print": true, "force_logout": true, "role_change": true, "dispatch_cancel": true, "void": true,
}

// auditDefaultWindow 為未帶時間篩選時套用的預設保留期限(近 3 個月, D27)。
const auditDefaultWindow = 3 * 30 * 24 * time.Hour

// AuditService 實作 audit.v1.AuditService(稽核查詢)。
type AuditService struct {
	db *ent.Client
	auditv1connect.UnimplementedAuditServiceHandler
}

// NewAuditService 建立 AuditService。
func NewAuditService(db *ent.Client) *AuditService {
	return &AuditService{db: db}
}

// RegisterAuditServices 將 AuditService 的 Connect handler 掛到 mux。
func RegisterAuditServices(mux *http.ServeMux, db *ent.Client) {
	path, handler := auditv1connect.NewAuditServiceHandler(NewAuditService(db), dbtenant.HandlerOption(db))
	mux.Handle(path, handler)
}

// isCompanyAdmin 判斷身分是否為 company_admin(可查自己公司稽核)。
func isCompanyAdmin(id authz.Identity) bool {
	return slices.Contains(id.Roles, "company_admin")
}

// ListAuditLogs 分頁查詢稽核紀錄,條件皆為 AND;時間降冪(最新在前)。
func (s *AuditService) ListAuditLogs(ctx context.Context, req *connect.Request[auditv1.ListAuditLogsRequest]) (*connect.Response[auditv1.ListAuditLogsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := s.db.AuditLog.Query()

	// 範圍推導:super 全域(可選 company_id);company_admin 強制自己公司;其餘角色不可查。
	switch {
	case isSuperIdentity(id):
		// super/developer:全域;可選 company_id 篩選(非法值 → invalid_argument)。
		if raw := req.Msg.GetCompanyId(); raw != "" {
			cid, err := parseID(raw)
			if err != nil {
				return nil, err
			}
			q = q.Where(auditlog.CompanyIDEQ(cid))
		}
	case isCompanyAdmin(id):
		// fail-closed:缺公司脈絡 → 拒絕。
		if id.CompanyID == "" {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
		}
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		// 強制限自己公司,忽略請求自帶的他公司篩選。
		q = q.Where(auditlog.CompanyIDEQ(cid))
	default:
		// dept_admin / staff / customer / guest:不得查稽核。
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無稽核查詢權限"))
	}

	// 時間窗:未帶任何 from/to 時套用預設近 3 個月(D27);from/to 以 RFC3339 驗證。
	hasRange := req.Msg.GetFrom() != "" || req.Msg.GetTo() != ""
	if !hasRange {
		q = q.Where(auditlog.CreatedAtGTE(time.Now().UTC().Add(-auditDefaultWindow)))
	} else {
		if f := req.Msg.GetFrom(); f != "" {
			t, err := time.Parse(time.RFC3339, f)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 from %q", f))
			}
			q = q.Where(auditlog.CreatedAtGTE(t))
		}
		if to := req.Msg.GetTo(); to != "" {
			t, err := time.Parse(time.RFC3339, to)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 to %q", to))
			}
			q = q.Where(auditlog.CreatedAtLTE(t))
		}
	}

	// 其餘篩選條件(AND)。
	if a := req.Msg.GetAction(); a != "" {
		if !validAuditActions[a] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 action %q", a))
		}
		q = q.Where(auditlog.ActionEQ(auditlog.Action(a)))
	}
	if rt := req.Msg.GetResourceType(); rt != "" {
		q = q.Where(auditlog.ResourceTypeEQ(rt))
	}
	if rid := req.Msg.GetResourceId(); rid != "" {
		q = q.Where(auditlog.ResourceIDEQ(rid))
	}
	if uid := req.Msg.GetUserId(); uid != "" {
		u, err := parseID(uid)
		if err != nil {
			return nil, err
		}
		q = q.Where(auditlog.UserIDEQ(u))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Clone().
		// M2:created_at 由同一業務交易的多筆稽核寫入(同一交易時間),同值群真實存在且常大於
		// 一頁;排序鍵非唯一時 PostgreSQL 對 ties 的順序在 bounded top-N 與完整排序之間不保證
		// 一致,逐頁 LIMIT/OFFSET 會重複與遺漏資料,故以 id 為次序鍵(與 F1/F2 同法,方向一致)。
		Order(ent.Desc(auditlog.FieldCreatedAt), ent.Desc(auditlog.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	// 操作者顯示名稱:依結果 user_id 一次查 users(避免 N+1)。
	names := map[int]string{}
	if len(items) > 0 {
		ids := make([]int, 0, len(items))
		for _, a := range items {
			ids = append(ids, a.UserID)
		}
		us, err := s.db.User.Query().Where(user.IDIn(ids...)).All(ctx)
		if err != nil {
			return nil, toConnectError(err)
		}
		for _, u := range us {
			names[u.ID] = u.Name
		}
	}

	out := make([]*auditv1.AuditLog, 0, len(items))
	for _, a := range items {
		out = append(out, auditLogToProto(a, names[a.UserID]))
	}
	return connect.NewResponse(&auditv1.ListAuditLogsResponse{
		Items:      out,
		Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

// auditLogToProto 將 ent.AuditLog 轉為 proto AuditLog(snapshot 以 JSON 字串承載)。
func auditLogToProto(a *ent.AuditLog, userName string) *auditv1.AuditLog {
	p := &auditv1.AuditLog{
		Id:           strconv.FormatInt(int64(a.ID), 10),
		CompanyId:    strconv.FormatInt(int64(a.CompanyID), 10),
		UserId:       strconv.FormatInt(int64(a.UserID), 10),
		UserName:     userName,
		Action:       string(a.Action),
		ResourceType: a.ResourceType,
		ResourceId:   a.ResourceID,
		IpAddress:    a.IPAddress,
		UserAgent:    a.UserAgent,
	}
	if a.DepartmentID != 0 {
		p.DepartmentId = strconv.FormatInt(int64(a.DepartmentID), 10)
	}
	if a.BeforeSnapshot != nil {
		if b, err := json.Marshal(a.BeforeSnapshot); err == nil {
			p.BeforeSnapshot = string(b)
		}
	}
	if a.AfterSnapshot != nil {
		if b, err := json.Marshal(a.AfterSnapshot); err == nil {
			p.AfterSnapshot = string(b)
		}
	}
	if !a.CreatedAt.IsZero() {
		p.CreatedAt = a.CreatedAt.Format(time.RFC3339)
	}
	return p
}
