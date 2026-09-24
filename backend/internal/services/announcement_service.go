// AnnouncementService 公告 CMS(spec announcements):banner/news/article 三型別、
// 三層發佈範圍、上下架時間窗、平台投放。CRUD 範圍守衛依 spec「管理權限依範圍分層」:
// super 全範圍、company_admin 所屬公司(公司層+該公司部門層)、dept_admin 本部門層,
// 其餘角色一律拒絕。可見性由 RLS(00044 policy + 00045 ENABLE/FORCE)兜底。
//
// 稽核:比照 metadicts —— CMS 寫入不落 audit_logs(全系統公告 company NULL 亦過不了
// audit.Record 的 company_id≠0 約束);如需稽核另立任務。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/announcement"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// announcementTypes 為合法型別集合(spec announcements 資料模型)。
var announcementTypes = map[string]bool{"banner": true, "news": true, "article": true}

// AnnouncementService 實作 salesorder.v1.AnnouncementService。
type AnnouncementService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedAnnouncementServiceHandler
}

// NewAnnouncementService 建立 AnnouncementService。
func NewAnnouncementService(db *ent.Client) *AnnouncementService {
	return &AnnouncementService{db: db}
}

// RegisterAnnouncementService 掛到 /api/v1(租戶 session + RLS,比照 SalesOrderService)。
func RegisterAnnouncementService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewAnnouncementServiceHandler(NewAnnouncementService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// canManageAnnouncement 範圍守衛(spec「管理權限依範圍分層」):
// super 任意範圍;company_admin = 自己公司(公司層與該公司任一部門層);
// dept_admin = 自己部門層(必須帶部門);其餘角色/越權 → permission_denied。
//
// targetCompany/targetDept 為該公告(或欲建立的目標)範圍,nil = 全系統/公司層缺失。
func canManageAnnouncement(id authz.Identity, targetCompany, targetDept *int) error {
	switch {
	case isSuperIdentity(id):
		return nil
	case hasRole(id, "company_admin"):
		own, err := parseID(id.CompanyID)
		if err != nil || targetCompany == nil || *targetCompany != own {
			return errcode.SysPermissionDenied.Error(nil)
		}
		return nil
	case hasRole(id, "dept_admin"):
		ownCompany, err := parseID(id.CompanyID)
		if err != nil || targetCompany == nil || *targetCompany != ownCompany {
			return errcode.SysPermissionDenied.Error(nil)
		}
		ownDept, err := parseID(id.DepartmentID)
		if err != nil || targetDept == nil || *targetDept != ownDept {
			// dept_admin 僅本部門層:公司層/全系統(targetDept 空)一律拒絕。
			return errcode.SysPermissionDenied.Error(nil)
		}
		return nil
	default:
		return errcode.SysPermissionDenied.Error(nil)
	}
}

// validateAnnouncementDept 目標部門必須存在且屬於目標公司(防止把公告掛到別公司的部門)。
// dept 為 nil(全系統/公司層)即跳過。
func validateAnnouncementDept(ctx context.Context, db *ent.Client, companyID, deptID *int) error {
	if deptID == nil {
		return nil
	}
	if companyID == nil {
		return errcode.SysPermissionDenied.Error(nil)
	}
	ok, err := db.Department.Query().
		Where(department.ID(*deptID),
			department.HasCompanyWith(company.IDEQ(*companyID))).
		Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if !ok {
		return errcode.SysPermissionDenied.Error(nil)
	}
	return nil
}

// parseOptionalID 空字串 → nil(全系統/公司層缺省),否則解析。
func parseOptionalID(raw string) (*int, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, nil
	}
	v, err := parseID(s)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// parseAnnouncementTime 解析 RFC3339;emptyAllowed 時空字串回 (零值, nil 由呼叫端決定)。
func parseAnnouncementTime(raw string, emptyOK bool) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		if emptyOK {
			return time.Time{}, nil
		}
		return time.Time{}, invalidArgField("publish_at")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// 欄位名由呼叫端 errFieldEmptyAs 改寫(publish_at/unpublish_at 共用解析器)。
		return time.Time{}, invalidArgField("publish_at")
	}
	return t, nil
}

func announcementToProto(a *ent.Announcement) *salesorderv1.Announcement {
	out := &salesorderv1.Announcement{
		Id:        strconv.Itoa(a.ID),
		Type:      a.Type,
		Title:     a.Title,
		Content:   a.Content,
		ImageUrl:  a.ImageURL,
		LinkUrl:   a.LinkURL,
		SortOrder: int32(a.SortOrder),
		IsActive:  a.IsActive,
		DeployWeb: a.DeployWeb,
		DeployApp: a.DeployApp,
		PublishAt: a.PublishAt.Format(time.RFC3339),
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
	if a.CompanyID != nil {
		out.CompanyId = strconv.Itoa(*a.CompanyID)
	}
	if a.DepartmentID != nil {
		out.DepartmentId = strconv.Itoa(*a.DepartmentID)
	}
	if a.UnpublishAt != nil {
		out.UnpublishAt = a.UnpublishAt.Format(time.RFC3339)
	}
	return out
}

// ListAnnouncements 管理列表:只列**可管理範圍**(super 全部;company_admin 限自己公司
// —— 系統公告對其為唯讀可見(RLS),但 CMS 不列,由 spec「僅可管理」收斂;dept_admin 限本部門)。
func (s *AnnouncementService) ListAnnouncements(ctx context.Context, req *connect.Request[salesorderv1.ListAnnouncementsRequest]) (*connect.Response[salesorderv1.ListAnnouncementsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "announcement", "read"); err != nil {
		return nil, err
	}
	q := dbtenant.Client(ctx, s.db).Announcement.Query()
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(announcement.DeletedAtIsNil())
	}
	switch {
	case isSuperIdentity(id):
		// 全範圍(含全系統公告)。
	case hasRole(id, "company_admin"):
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		q = q.Where(announcement.CompanyIDEQ(cid))
	case hasRole(id, "dept_admin"):
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		did, err := parseID(id.DepartmentID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		q = q.Where(announcement.CompanyIDEQ(cid), announcement.DepartmentIDEQ(did))
	default:
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	if t := strings.TrimSpace(req.Msg.GetType()); t != "" {
		if !announcementTypes[t] {
			return nil, invalidArgField("type")
		}
		q = q.Where(announcement.TypeEQ(t))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	rows, err := q.Order(ent.Desc(announcement.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*salesorderv1.Announcement, 0, len(rows))
	for _, a := range rows {
		out = append(out, announcementToProto(a))
	}
	return connect.NewResponse(&salesorderv1.ListAnnouncementsResponse{
		Announcements: out, Total: int32(total),
	}), nil
}

// CreateAnnouncement 建立:範圍守衛 + 部門歸屬 + 型別/標題驗證,同一交易寫入。
func (s *AnnouncementService) CreateAnnouncement(ctx context.Context, req *connect.Request[salesorderv1.CreateAnnouncementRequest]) (*connect.Response[salesorderv1.CreateAnnouncementResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "announcement", "write"); err != nil {
		return nil, err
	}
	if !announcementTypes[strings.TrimSpace(req.Msg.GetType())] {
		return nil, invalidArgField("type")
	}
	if strings.TrimSpace(req.Msg.GetTitle()) == "" {
		return nil, invalidArgField("title")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	targetCompany, err := parseOptionalID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	targetDept, err := parseOptionalID(req.Msg.GetDepartmentId())
	if err != nil {
		return nil, err
	}
	// 範圍空值自動歸屬(非 super,見 proto 註):company_admin 預設公司層;
	// dept_admin(不含其上位 company_admin)補自己部門 → 恆為部門層。
	if targetCompany == nil && !isSuperIdentity(id) {
		own, err := parseID(id.CompanyID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		targetCompany = &own
	}
	if targetDept == nil && hasRole(id, "dept_admin") && !hasRole(id, "company_admin") {
		own, err := parseID(id.DepartmentID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		targetDept = &own
	}
	if err := canManageAnnouncement(id, targetCompany, targetDept); err != nil {
		return nil, err
	}
	if err := validateAnnouncementDept(ctx, db, targetCompany, targetDept); err != nil {
		return nil, err
	}
	publishAt, err := parseAnnouncementTime(req.Msg.GetPublishAt(), true)
	if err != nil {
		return nil, err
	}
	if publishAt.IsZero() {
		publishAt = time.Now()
	}
	unpublishAt, err := parseAnnouncementTime(req.Msg.GetUnpublishAt(), true)
	if err != nil {
		return nil, errFieldEmptyAs("unpublish_at", err)
	}
	actor, _ := parseID(id.UserID)

	build := db.Announcement.Create().
		SetType(strings.TrimSpace(req.Msg.GetType())).
		SetTitle(strings.TrimSpace(req.Msg.GetTitle())).
		SetContent(req.Msg.GetContent()).
		SetImageURL(req.Msg.GetImageUrl()).
		SetLinkURL(req.Msg.GetLinkUrl()).
		SetPublishAt(publishAt).
		SetSortOrder(int(req.Msg.GetSortOrder())).
		SetIsActive(req.Msg.GetIsActive()).
		SetDeployWeb(req.Msg.GetDeployWeb()).
		SetDeployApp(req.Msg.GetDeployApp()).
		SetCreatedBy(actor)
	if targetCompany != nil {
		build = build.SetCompanyID(*targetCompany)
	}
	if targetDept != nil {
		build = build.SetDepartmentID(*targetDept)
	}
	if !unpublishAt.IsZero() {
		build = build.SetUnpublishAt(unpublishAt)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.CreateAnnouncementResponse{
		Announcement: announcementToProto(created),
	}), nil
}

// errfieldEmptyAs 把 parseAnnouncementTime 對欄位的誤判改為指定欄位名
// (publish_at/unpublish_at 兩欄共用解析器,空值/格式錯都要回到自己的欄位名)。
func errFieldEmptyAs(field string, err error) error {
	if connect.CodeOf(err) == connect.CodeInvalidArgument {
		return invalidArgField(field)
	}
	return err
}

// UpdateAnnouncement 全量替換:先在 RLS 範圍內載入(範圍外 = not_found),
// 再對**舊範圍與新範圍**都做範圍守衛(防止把別處公告改掛到自己名下,或動系統公告)。
func (s *AnnouncementService) UpdateAnnouncement(ctx context.Context, req *connect.Request[salesorderv1.UpdateAnnouncementRequest]) (*connect.Response[salesorderv1.UpdateAnnouncementResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "announcement", "write"); err != nil {
		return nil, err
	}
	if !announcementTypes[strings.TrimSpace(req.Msg.GetType())] {
		return nil, invalidArgField("type")
	}
	if strings.TrimSpace(req.Msg.GetTitle()) == "" {
		return nil, invalidArgField("title")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	aid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	existing, err := db.Announcement.Query().
		Where(announcement.ID(aid), announcement.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 舊範圍守衛:全系統公告(company NULL)僅 super 可動;他公司/他部門列於
	// RLS 外的第二道防線(sqlite 單測無 RLS,靠這條擋跨範圍編輯)。
	if err := canManageAnnouncement(id, existing.CompanyID, existing.DepartmentID); err != nil {
		return nil, err
	}
	// 範圍不可改(proto 註:v1 於建立時決定) → 本請求不含範圍欄位,列範圍維持原樣。
	publishAt, err := parseAnnouncementTime(req.Msg.GetPublishAt(), false)
	if err != nil {
		return nil, err
	}
	unpublishAt, err := parseAnnouncementTime(req.Msg.GetUnpublishAt(), true)
	if err != nil {
		return nil, errFieldEmptyAs("unpublish_at", err)
	}

	up := db.Announcement.UpdateOneID(aid).
		SetType(strings.TrimSpace(req.Msg.GetType())).
		SetTitle(strings.TrimSpace(req.Msg.GetTitle())).
		SetContent(req.Msg.GetContent()).
		SetImageURL(req.Msg.GetImageUrl()).
		SetLinkURL(req.Msg.GetLinkUrl()).
		SetPublishAt(publishAt).
		SetSortOrder(int(req.Msg.GetSortOrder())).
		SetIsActive(req.Msg.GetIsActive()).
		SetDeployWeb(req.Msg.GetDeployWeb()).
		SetDeployApp(req.Msg.GetDeployApp())
	if !unpublishAt.IsZero() {
		up = up.SetUnpublishAt(unpublishAt)
	} else {
		up = up.ClearUnpublishAt()
	}
	updated, err := up.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.UpdateAnnouncementResponse{
		Announcement: announcementToProto(updated),
	}), nil
}

// DeleteAnnouncement 軟刪除(舊範圍守衛同 Update)。
func (s *AnnouncementService) DeleteAnnouncement(ctx context.Context, req *connect.Request[salesorderv1.DeleteAnnouncementRequest]) (*connect.Response[salesorderv1.DeleteAnnouncementResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "announcement", "write"); err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	aid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	existing, err := db.Announcement.Query().
		Where(announcement.ID(aid), announcement.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := canManageAnnouncement(id, existing.CompanyID, existing.DepartmentID); err != nil {
		return nil, err
	}
	if _, err := db.Announcement.UpdateOneID(aid).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.DeleteAnnouncementResponse{}), nil
}
