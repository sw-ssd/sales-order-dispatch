// Package services 提供可直接掛載到 HTTP server 的 Connect 服務實作。
package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// validUserStatuses 為 User.status 允許值(對齊 ent enum)。
var validUserStatuses = map[string]bool{
	string(user.StatusActive):   true,
	string(user.StatusInactive): true,
	string(user.StatusPending):  true,
}

// isUserManager 判斷身分是否為可管理使用者的角色(對齊 rolePolicy/細節 2.3:
// super/company_admin/dept_admin)。staff/customer/guest 不管理使用者。
func isUserManager(id authz.Identity) bool {
	for _, r := range id.Roles {
		if r == "super" || r == "company_admin" || r == "dept_admin" {
			return true
		}
	}
	return false
}

// isSuperIdentity 判斷身分是否含 super/developer(全域角色)。
func isSuperIdentity(id authz.Identity) bool {
	return slices.Contains(id.Roles, "super") || slices.Contains(id.Roles, "developer")
}

// grantableRoles 回傳操作者可授予的角色集合(授予上限;防止權限提升)。
// super/developer 可授予全部內建角色;company_admin 限 company_admin/dept_admin/staff
// (不得授予 super/developer;對齊 01-auth 1.4.4 的審核角色上限);
// dept_admin 僅能授予 staff(規格 2.3.2);其餘角色不可授予任何角色。
func grantableRoles(id authz.Identity) map[string]bool {
	if isSuperIdentity(id) {
		out := make(map[string]bool, len(auth.BuiltinRoles))
		for _, r := range auth.BuiltinRoles {
			out[r] = true
		}
		return out
	}
	if slices.Contains(id.Roles, "company_admin") {
		return map[string]bool{"company_admin": true, "dept_admin": true, "staff": true}
	}
	if slices.Contains(id.Roles, "dept_admin") {
		return map[string]bool{"staff": true}
	}
	return nil
}

// UserService 實作 salesorder.v1.UserService(使用者管理,02 計畫 Task 3)。
// 範圍控制:super 全域 / company_admin 限自己公司 / dept_admin 限自己部門 staff;
// 授權閘門由 middleware OpenFGA Check(prod)承擔,本層以 scopeForTarget 做範圍控制 + RLS 兜底。
type UserService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedUserServiceHandler
}

// NewUserService 建立 UserService。
func NewUserService(db *ent.Client) *UserService {
	return &UserService{db: db}
}

// RegisterUserServices 將 UserService 的 Connect handler 掛到 mux。
func RegisterUserServices(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewUserServiceHandler(NewUserService(db))
	mux.Handle(path, handler)
}

// ListUsers 分頁列出使用者,依操作者範圍強制注入 filter。
func (s *UserService) ListUsers(ctx context.Context, req *connect.Request[v1.ListUsersRequest]) (*connect.Response[v1.ListUsersResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	// 授權下限:非管理角色不得列舉使用者(C1;避免跨租戶使用者列舉)。
	if !isUserManager(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無使用者查詢權限"))
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())

	q := s.db.User.Query()

	// 範圍強制注入(fail-closed:忽略請求自帶的超範圍參數)。
	companyID := req.Msg.GetCompanyId()
	deptID := req.Msg.GetDepartmentId()
	switch {
	case isSuperIdentity(id):
		// super/developer:全域;可依請求 company_id / department_id 篩選。
		if companyID != "" {
			cid, err := parseID(companyID)
			if err != nil {
				return nil, err
			}
			q = q.Where(user.HasCompanyWith(company.ID(cid)))
		}
		if deptID != "" {
			did, err := parseID(deptID)
			if err != nil {
				return nil, err
			}
			q = q.Where(user.HasDepartmentWith(department.ID(did)))
		}
	case slices.Contains(id.Roles, "company_admin"):
		// company_admin:強制限自己公司(缺公司脈絡 → fail-closed)。
		if id.CompanyID == "" {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
		}
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		q = q.Where(user.HasCompanyWith(company.ID(cid)))
		// 可再依請求 department_id 縮小(限自己公司內)。
		if deptID != "" {
			did, err := parseID(deptID)
			if err != nil {
				return nil, err
			}
			q = q.Where(user.HasDepartmentWith(department.ID(did)))
		}
	default:
		// dept_admin:強制限自己部門(缺部門脈絡 → fail-closed)。
		if id.DepartmentID == "" {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少部門範圍"))
		}
		did, err := parseID(id.DepartmentID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		q = q.Where(user.HasDepartmentWith(department.ID(did)))
	}

	if role := req.Msg.GetRole(); role != "" {
		if !isValidRole(role) {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的角色 %q", role))
		}
		q = q.Where(user.Role(role))
	}
	if status := req.Msg.GetStatus(); status != "" {
		if !validUserStatuses[status] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 status %q", status))
		}
		q = q.Where(user.StatusEQ(user.Status(status)))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Clone().
		WithCompany().WithDepartment().
		Order(ent.Asc(user.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	users := make([]*v1.User, 0, len(items))
	for _, u := range items {
		users = append(users, userToProto(u))
	}
	return connect.NewResponse(&v1.ListUsersResponse{
		Users:      users,
		Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

// GetUser 取得單一使用者(不含 password_hash)。
func (s *UserService) GetUser(ctx context.Context, req *connect.Request[v1.GetUserRequest]) (*connect.Response[v1.GetUserResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	userID, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}
	u, err := s.db.User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := s.scopeForTarget(id, u); err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.GetUserResponse{User: userToProto(u)}), nil
}

// CreateUser 建立員工帳號(super/company_admin/dept_admin;password_hash 依 OAuth 首登填補)。
func (s *UserService) CreateUser(ctx context.Context, req *connect.Request[v1.CreateUserRequest]) (*connect.Response[v1.CreateUserResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	if !isUserManager(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無使用者管理權限"))
	}
	if strings.TrimSpace(req.Msg.GetName()) == "" || strings.TrimSpace(req.Msg.GetEmail()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 與 email 必填"))
	}
	cid, err := parseID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	// company_admin 只能在自己公司建帳號。
	if id.Role == "company_admin" && id.CompanyID != "" {
		if own, err := parseID(id.CompanyID); err != nil || own != cid {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("僅能建所屬公司帳號"))
		}
	}
	roleCode := req.Msg.GetRole()
	if !isValidRole(roleCode) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的角色 %q", roleCode))
	}
	// 授予上限(C2):操作者不得建立高於自身可授予範圍的角色(防止權限提升)。
	if !grantableRoles(id)[roleCode] {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("無權限授予角色 %q", roleCode))
	}

	deptRef := 0
	if req.Msg.GetDepartmentId() != "" {
		did, err := parseID(req.Msg.GetDepartmentId())
		if err != nil {
			return nil, err
		}
		// 部門必須屬於目標公司(I6;避免跨公司資料擺放)。
		if err := s.validateDepartmentInCompany(ctx, did, cid); err != nil {
			return nil, err
		}
		deptRef = did
	}

	// 建帳號為關鍵操作:業務異動 + 稽核(D18)同一交易。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	build := tx.User.Create().
		SetCompanyID(cid).
		SetEmail(req.Msg.GetEmail()).
		SetName(req.Msg.GetName()).
		SetRole(roleCode).
		SetStatus(user.StatusActive).
		// 員工帳號走 OAuth 不存密碼:password_hash 以 OIDC sentinel 佔位(規格 4.1,密碼登入必失敗)。
		SetPasswordHash(auth.OIDCPasswordSentinel)
	if deptRef > 0 {
		build = build.SetDepartmentID(deptRef)
	}
	if req.Msg.GetPhone() != "" {
		build = build.SetPhone(req.Msg.GetPhone())
	}
	if req.Msg.GetEmployeeNo() != "" {
		build = build.SetEmployeeNo(req.Msg.GetEmployeeNo())
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	actorID := 0
	if pid, perr := parseID(id.UserID); perr == nil {
		actorID = pid
	}
	var auditDept *int
	if deptRef > 0 {
		d := deptRef
		auditDept = &d
	}
	meta := audit.MetaFrom(ctx)
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "create",
		ResourceType: "user",
		ResourceID:   uItoaInt(created.ID),
		CompanyID:    cid,
		DepartmentID: auditDept,
		UserID:       actorID,
		After:        map[string]any{"name": created.Name, "email": created.Email, "role": created.Role, "status": string(created.Status)},
		IPAddress:    meta.IP,
		UserAgent:    meta.UserAgent,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("稽核寫入失敗: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}

	u, err := s.db.User.Query().WithCompany().WithDepartment().Where(user.ID(created.ID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.CreateUserResponse{User: userToProto(u)}), nil
}

// UpdateUser 更新使用者欄位(僅更新出現的 optional 欄位)。
func (s *UserService) UpdateUser(ctx context.Context, req *connect.Request[v1.UpdateUserRequest]) (*connect.Response[v1.UpdateUserResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	if !isUserManager(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無使用者管理權限"))
	}
	userID, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}
	target, err := s.loadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.scopeForTarget(id, target); err != nil {
		return nil, err
	}

	// 更新為關鍵操作:業務異動 + 稽核(D18)同一交易。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	before := map[string]any{}
	after := map[string]any{}
	update := tx.User.UpdateOneID(userID)
	if req.Msg.Name != nil {
		update = update.SetName(req.Msg.GetName())
		before["name"] = target.Name
		after["name"] = req.Msg.GetName()
	}
	if req.Msg.Phone != nil {
		update = update.SetPhone(req.Msg.GetPhone())
		before["phone"] = target.Phone
		after["phone"] = req.Msg.GetPhone()
	}
	if req.Msg.EmployeeNo != nil {
		update = update.SetEmployeeNo(req.Msg.GetEmployeeNo())
		before["employee_no"] = target.EmployeeNo
		after["employee_no"] = req.Msg.GetEmployeeNo()
	}
	if req.Msg.Status != nil {
		if !validUserStatuses[req.Msg.GetStatus()] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 status %q", req.Msg.GetStatus()))
		}
		update = update.SetStatus(user.Status(req.Msg.GetStatus()))
		before["status"] = string(target.Status)
		after["status"] = req.Msg.GetStatus()
	}
	if req.Msg.DepartmentId != nil {
		if req.Msg.GetDepartmentId() == "" {
			update = update.ClearDepartment()
		} else {
			did, err := parseID(req.Msg.GetDepartmentId())
			if err != nil {
				return nil, err
			}
			// 部門必須屬於目標使用者公司(I6)。
			companyRef := target.Edges.Company
			if companyRef == nil {
				return nil, connect.NewError(connect.CodeInternal, errors.New("目標使用者缺少公司關聯"))
			}
			if err := s.validateDepartmentInCompany(ctx, did, companyRef.ID); err != nil {
				return nil, err
			}
			update = update.SetDepartmentID(did)
			after["department_id"] = did
		}
	}
	if _, err := update.Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := s.recordAudit(ctx, tx, id, target, "update", before, after); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("稽核寫入失敗: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	u, err := s.db.User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.UpdateUserResponse{User: userToProto(u)}), nil
}

// AssignRole 指派人角色(含 guest 審核:status pending → active)。範圍內可用。
func (s *UserService) AssignRole(ctx context.Context, req *connect.Request[v1.AssignRoleRequest]) (*connect.Response[v1.AssignRoleResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	if !isUserManager(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無使用者管理權限"))
	}
	userID, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}
	target, err := s.loadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.scopeForTarget(id, target); err != nil {
		return nil, err
	}
	roleCode := req.Msg.GetRole()
	if !isValidRole(roleCode) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的角色 %q", roleCode))
	}
	if roleCode == "guest" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("不可指派為 guest(待審核狀態)"))
	}
	// 授予上限(C2):不得指派高於自身可授予範圍的角色(防止權限提升,含自我升權)。
	if !grantableRoles(id)[roleCode] {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("無權限授予角色 %q", roleCode))
	}

	// 角色指派(含 guest 審核)為關鍵操作:角色變更 + token_version+1(D5) + 稽核(D18)同一交易。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	update := tx.User.UpdateOneID(userID).
		SetRole(roleCode).
		AddTokenVersion(1) // 角色變更 → 在途舊 JWT/session 立即失效(D5)
	// guest 審核:僅 pending 轉 active;已停用帳號不被默默復活(I5,
	// 重新啟用應走 UpdateUser,規格 2.3.2)。
	if target.Status == user.StatusPending {
		update = update.SetStatus(user.StatusActive)
	}
	if req.Msg.GetDepartmentId() != "" {
		did, err := parseID(req.Msg.GetDepartmentId())
		if err != nil {
			return nil, err
		}
		companyRef := target.Edges.Company
		if companyRef == nil {
			return nil, connect.NewError(connect.CodeInternal, errors.New("目標使用者缺少公司關聯"))
		}
		if err := s.validateDepartmentInCompany(ctx, did, companyRef.ID); err != nil {
			return nil, err
		}
		update = update.SetDepartmentID(did)
	}
	if _, err := update.Save(ctx); err != nil {
		return nil, toConnectError(err)
	}

	// 稽核(同一交易,D18)。
	if err := s.recordAudit(ctx, tx, id, target, "role_change",
		map[string]any{"role": target.Role},
		map[string]any{"role": roleCode},
	); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("稽核寫入失敗: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}

	// 同步 OpenFGA assigned tuple(引擎未注入時略過,OpenFGA 停用/開發降級時由 rolePolicy 承擔)。
	// 對齊 role_service.syncRolePermissions:OpenFGA 與業務非同交易,commit 後執行。
	if err := s.syncUserRoleTuple(ctx, target, roleCode); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("OpenFGA tuple 同步失敗: %w", err))
	}

	u, err := s.db.User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.AssignRoleResponse{User: userToProto(u)}), nil
}

// Deactivate 停用帳號(status → inactive, token_version+1)。
func (s *UserService) Deactivate(ctx context.Context, req *connect.Request[v1.DeactivateRequest]) (*connect.Response[v1.DeactivateResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	if !isUserManager(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無使用者管理權限"))
	}
	userID, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}
	target, err := s.loadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.scopeForTarget(id, target); err != nil {
		return nil, err
	}
	if target.Status == user.StatusInactive {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("帳號已為 inactive"))
	}
	// 停用為關鍵操作:status 異動 + token_version+1(D5) + 稽核(D18)同一交易。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.User.UpdateOneID(userID).
		SetStatus(user.StatusInactive).
		AddTokenVersion(1).
		Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := s.recordAudit(ctx, tx, id, target, "update",
		map[string]any{"status": string(target.Status)},
		map[string]any{"status": string(user.StatusInactive)},
	); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("稽核寫入失敗: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.DeactivateResponse{}), nil
}

// ForceLogout 強制登出(token_version+1,使在途憑證失效)。不可對自己呼叫。
func (s *UserService) ForceLogout(ctx context.Context, req *connect.Request[v1.ForceLogoutRequest]) (*connect.Response[v1.ForceLogoutResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	if !isUserManager(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無使用者管理權限"))
	}
	userID, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}
	if id.UserID != "" {
		if self, perr := parseID(id.UserID); perr == nil && self == userID {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("不可對自己執行強制登出"))
		}
	}
	target, err := s.loadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.scopeForTarget(id, target); err != nil {
		return nil, err
	}
	// 強制登出為關鍵操作:token_version+1 + 稽核(D18)同一交易。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.User.UpdateOneID(userID).AddTokenVersion(1).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := s.recordAudit(ctx, tx, id, target, "force_logout", nil, nil); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("稽核寫入失敗: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.ForceLogoutResponse{}), nil
}

// recordAudit 於交易內依操作者身分與目標使用者寫一筆稽核(統一入口,D18)。
// before/after 為「已過濾敏感欄位」的變更前/後摘要(可空)。
func (s *UserService) recordAudit(ctx context.Context, tx *ent.Tx, actor authz.Identity, target *ent.User, action string, before, after map[string]any) error {
	companyID := 0
	if target.Edges.Company != nil {
		companyID = target.Edges.Company.ID
	}
	actorID := 0
	if pid, err := parseID(actor.UserID); err == nil {
		actorID = pid
	}
	var deptID *int
	if target.Edges.Department != nil {
		d := target.Edges.Department.ID
		deptID = &d
	}
	// 來源資訊(IP/User-Agent)由 middleware 每請求注入 ctx(I9;2.6.2)。
	meta := audit.MetaFrom(ctx)
	return audit.Record(ctx, tx, audit.Entry{
		Action:       action,
		ResourceType: "user",
		ResourceID:   uItoaInt(target.ID),
		CompanyID:    companyID,
		DepartmentID: deptID,
		UserID:       actorID,
		Before:       before,
		After:        after,
		IPAddress:    meta.IP,
		UserAgent:    meta.UserAgent,
	})
}

// syncUserRoleTuple 同步使用者的 OpenFGA assigned tuple(角色異動):
// 刪除舊 role 的 assigned、寫入新 role 的 assigned(對齊 provision 的 user→role 格式)。
// 引擎未注入(引擎 = EngineFrom ctx)或引擎/角色查詢不可用時靜默略過(OpenFGA 停用/降級)。
func (s *UserService) syncUserRoleTuple(ctx context.Context, target *ent.User, newRoleCode string) error {
	e := authz.EngineFrom(ctx)
	if e == nil {
		return nil
	}
	userObj := fmt.Sprintf("user:%d", target.ID)

	// 刪除舊角色 assigned(若存在)。
	if target.Role != "" && target.Role != newRoleCode {
		if oldRole, err := s.db.Role.Query().Where(role.CodeEQ(target.Role)).Only(ctx); err == nil {
			if derr := e.DeleteTuple(ctx, userObj, "assigned", fmt.Sprintf("role:%d", oldRole.ID)); derr != nil {
				return derr
			}
		}
	}
	// 寫入新角色 assigned(若角色存在)。
	if nr, err := s.db.Role.Query().Where(role.CodeEQ(newRoleCode)).Only(ctx); err == nil {
		if werr := e.WriteTuple(ctx, userObj, "assigned", fmt.Sprintf("role:%d", nr.ID)); werr != nil {
			return werr
		}
	}
	return nil
}

// loadUser 載入單一使用者(含 company/department edge,供範圍判定)。
func (s *UserService) loadUser(ctx context.Context, userID int) (*ent.User, error) {
	u, err := s.db.User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return u, nil
}

// scopeForTarget 依操作者身分判斷目標使用者是否在管理範圍內。
// super 全域;company_admin 同公司;dept_admin 同部門且目標為 staff(細節 2.3.2)。
func (s *UserService) scopeForTarget(id authz.Identity, target *ent.User) error {
	if isSuperIdentity(id) {
		return nil
	}
	// 非 super:必須具備公司脈絡,否則 fail-closed(I1)。
	if id.CompanyID == "" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
	}
	opCid, err := parseID(id.CompanyID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if target.Edges.Company == nil || target.Edges.Company.ID != opCid {
		return connect.NewError(connect.CodePermissionDenied, errors.New("目標使用者不在操作者公司範圍"))
	}
	if slices.Contains(id.Roles, "dept_admin") && !slices.Contains(id.Roles, "company_admin") {
		// dept_admin 僅能管理自己部門的 staff。
		if id.DepartmentID == "" {
			return connect.NewError(connect.CodePermissionDenied, errors.New("dept_admin 缺少部門範圍"))
		}
		opDid, err := parseID(id.DepartmentID)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		if target.Edges.Department == nil || target.Edges.Department.ID != opDid {
			return connect.NewError(connect.CodePermissionDenied, errors.New("目標使用者不在操作者部門範圍"))
		}
		if target.Role != "staff" {
			return connect.NewError(connect.CodePermissionDenied, errors.New("dept_admin 僅能管理 staff 帳號"))
		}
	}
	return nil
}

// isValidRole 判斷角色 code 是否為已知角色(目前僅內建角色;自訂角色見 2.9.2 待接入)。
func isValidRole(role string) bool {
	return slices.Contains(auth.BuiltinRoles, role)
}

// validateDepartmentInCompany 驗證 department 存在且屬於指定公司(I6)。
// 不符 → InvalidArgument(輸入驗證失敗),不允許跨公司資料擺放。
func (s *UserService) validateDepartmentInCompany(ctx context.Context, deptID, companyID int) error {
	ok, err := s.db.Department.Query().
		Where(department.ID(deptID), department.HasCompanyWith(company.ID(companyID))).
		Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("部門 %d 不屬於公司 %d", deptID, companyID))
	}
	return nil
}

// uItoaInt 將 int 轉字串(稽核 ResourceID 用)。
func uItoaInt(i int) string {
	return strconv.Itoa(i)
}

// userToProto 將 ent.User 轉為 proto User(不含 password_hash)。
func userToProto(u *ent.User) *v1.User {
	p := &v1.User{
		Id:          strconv.FormatInt(int64(u.ID), 10),
		Name:        u.Name,
		Email:       u.Email,
		Phone:       u.Phone,
		EmployeeNo:  u.EmployeeNo,
		Status:      string(u.Status),
		Role:        u.Role,
		IsCustomer:  u.IsCustomer,
		AccountName: u.AccountName,
	}
	if u.Edges.Company != nil {
		p.CompanyId = strconv.FormatInt(int64(u.Edges.Company.ID), 10)
	}
	if u.Edges.Department != nil {
		p.DepartmentId = strconv.FormatInt(int64(u.Edges.Department.ID), 10)
	}
	return p
}
