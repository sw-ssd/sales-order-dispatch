// Package services 提供可直接掛載到 HTTP server 的 Connect 服務實作。
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"

	"google.golang.org/protobuf/types/known/structpb"
)

// RoleService 實作 salesorder.v1.RoleService(角色權限管理,T18)。
// 權限檢查(T14 Casbin):role 資源僅 super / company_admin(含 g 繼承)可管理;
// company_admin 限管理自訂(非 is_system)角色,且規則條件不得引用他人公司(限自己公司)。
// role_permissions 為 CASL ability 來源(前端權限矩陣與 AbilityService 共用)。
type RoleService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedRoleServiceHandler
}

// NewRoleService 建立 RoleService。
func NewRoleService(db *ent.Client) *RoleService {
	return &RoleService{db: db}
}

// RegisterRoleServices 將 RoleService 的 Connect handler 掛到 mux(以自然路徑
// "/salesorder.v1.RoleService/")。server 以 http.StripPrefix("/api/v1", mux) 掛載。
func RegisterRoleServices(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewRoleServiceHandler(NewRoleService(db))
	mux.Handle(path, handler)
}

// requireRole 檢查 ctx 身分具備 role 資源的指定動作(Casbin EnforceAny,T14)。
// 未登入 → Unauthenticated;無權 → PermissionDenied。
func requireRole(ctx context.Context, action string) error {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	ok, err := auth.EnforceAny(id.Roles, "role", action, id.CompanyID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if !ok {
		return connect.NewError(connect.CodePermissionDenied, errors.New("無角色權限管理權限"))
	}
	return nil
}

// isSuper 判斷身分是否含 super / developer(可管理內建角色)。
func isSuper(id authz.Identity) bool {
	for _, r := range id.Roles {
		if r == "super" || r == "developer" {
			return true
		}
	}
	return false
}

// ListRoles 分頁列出角色(依 id 升冪;角色權限設置頁的選單)。
func (s *RoleService) ListRoles(ctx context.Context, req *connect.Request[v1.ListRolesRequest]) (*connect.Response[v1.ListRolesResponse], error) {
	if err := requireRole(ctx, "read"); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())

	total, err := s.db.Role.Query().Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := s.db.Role.Query().
		Order(ent.Asc(role.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	roles := make([]*v1.Role, 0, len(items))
	for _, r := range items {
		roles = append(roles, &v1.Role{
			Id:        strconv.FormatInt(int64(r.ID), 10),
			Code:      r.Code,
			Name:      r.Name,
			DataScope: string(r.DataScope),
			IsSystem:  r.IsSystem,
			IsActive:  r.IsActive,
		})
	}
	return connect.NewResponse(&v1.ListRolesResponse{
		Roles:      roles,
		Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

// GetRolePermissions 取得角色功能權限(依 sort_order 升冪)。
func (s *RoleService) GetRolePermissions(ctx context.Context, req *connect.Request[v1.GetRolePermissionsRequest]) (*connect.Response[v1.GetRolePermissionsResponse], error) {
	if err := requireRole(ctx, "read"); err != nil {
		return nil, err
	}
	roleID, err := parseID(req.Msg.GetRoleId())
	if err != nil {
		return nil, err
	}
	exists, err := s.db.Role.Query().Where(role.ID(roleID)).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("角色 %d 不存在", roleID))
	}

	perms, err := s.loadPermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.GetRolePermissionsResponse{Permissions: perms}), nil
}

// UpdateRolePermissions 全量取代角色功能權限(交易內 delete + insert)。
// 限制:super / developer 可改任意角色;company_admin 僅可改自訂角色,
// 且規則條件若含 company_id 必須等於自身公司(限自己公司)。
func (s *RoleService) UpdateRolePermissions(ctx context.Context, req *connect.Request[v1.UpdateRolePermissionsRequest]) (*connect.Response[v1.UpdateRolePermissionsResponse], error) {
	id := authz.IdentityFrom(ctx)
	if err := requireRole(ctx, "update"); err != nil {
		return nil, err
	}
	roleID, err := parseID(req.Msg.GetRoleId())
	if err != nil {
		return nil, err
	}
	r, err := s.db.Role.Get(ctx, roleID)
	if err != nil {
		return nil, toConnectError(err)
	}
	if r.IsSystem && !isSuper(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("內建角色權限僅 super 可修改"))
	}
	if !isSuper(id) {
		if err := validateOwnCompany(id, req.Msg.GetPermissions()); err != nil {
			return nil, err
		}
	}
	perms, err := validatePermissions(req.Msg.GetPermissions())
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.RolePermission.Delete().Where(rolepermission.RoleID(roleID)).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	for _, p := range perms {
		build := tx.RolePermission.Create().
			SetRoleID(roleID).
			SetResource(p.resource).
			SetAction(p.action).
			SetInverted(p.inverted).
			SetSortOrder(p.sortOrder)
		if len(p.conditions) > 0 {
			build = build.SetConditions(p.conditions)
		}
		if _, err := build.Save(ctx); err != nil {
			return nil, toConnectError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}

	out, err := s.loadPermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.UpdateRolePermissionsResponse{Permissions: out}), nil
}

// loadPermissions 依 sort_order 升冪讀取角色功能權限並轉 proto。
func (s *RoleService) loadPermissions(ctx context.Context, roleID int) ([]*v1.Permission, error) {
	rows, err := s.db.RolePermission.Query().
		Where(rolepermission.RoleID(roleID)).
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	perms := make([]*v1.Permission, 0, len(rows))
	for _, rp := range rows {
		p, err := permissionToProto(rp)
		if err != nil {
			return nil, toConnectError(err)
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// permissionToProto 將 ent.RolePermission 轉為 proto Permission。
func permissionToProto(rp *ent.RolePermission) (*v1.Permission, error) {
	p := &v1.Permission{
		Resource:  rp.Resource,
		Action:    rp.Action,
		Inverted:  rp.Inverted,
		SortOrder: int32(rp.SortOrder),
	}
	if len(rp.Conditions) > 0 {
		s, err := structpb.NewStruct(rp.Conditions)
		if err != nil {
			return nil, fmt.Errorf("conditions 轉換失敗: %w", err)
		}
		p.Conditions = s
	}
	return p, nil
}

// protoConditionsToMap 將 proto Struct 轉為 conditions map(nil → nil)。
func protoConditionsToMap(s *structpb.Struct) (map[string]any, error) {
	if s == nil {
		return nil, nil
	}
	return s.AsMap(), nil
}

// conditionsKey 為重複規則偵測鍵(對齊 DB 唯一索引 role_id×resource×action×conditions)。
func conditionsKey(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("{%q}", err) // structpb 值必為 JSON 相容,理論上不會發生
	}
	return string(b)
}

// validatePermissions 結構驗證:resource/action 非空、同規則不重複。
type permission struct {
	resource   string
	action     string
	conditions map[string]any
	inverted   bool
	sortOrder  int
}

func validatePermissions(in []*v1.Permission) ([]permission, error) {
	seen := make(map[string]bool, len(in))
	out := make([]permission, 0, len(in))
	for i, p := range in {
		resource := strings.TrimSpace(p.GetResource())
		action := strings.TrimSpace(p.GetAction())
		if resource == "" || action == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("第 %d 筆權限 resource/action 不可為空", i+1))
		}
		conds, err := protoConditionsToMap(p.GetConditions())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("第 %d 筆權限 conditions 格式錯誤: %w", i+1, err))
		}
		key := resource + "\x00" + action + "\x00" + conditionsKey(conds)
		if seen[key] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("第 %d 筆權限與先前規則重複(resource=%q action=%q)", i+1, resource, action))
		}
		seen[key] = true
		out = append(out, permission{resource: resource, action: action, conditions: conds, inverted: p.GetInverted(), sortOrder: int(p.GetSortOrder())})
	}
	return out, nil
}

// validateOwnCompany 非 super 身分(company_admin)僅可寫自己公司範圍的規則:
// conditions 若含 company_id,值必須等於自身 company_id(不允許指向其他公司)。
func validateOwnCompany(id authz.Identity, in []*v1.Permission) error {
	for i, p := range in {
		conds, err := protoConditionsToMap(p.GetConditions())
		if err != nil {
			continue // 結構錯誤由 validatePermissions 處理
		}
		if v, ok := conds["company_id"]; ok {
			s, ok := v.(string)
			if !ok || s != id.CompanyID {
				return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("第 %d 筆權限引用其他公司資料範圍,僅可限自己公司", i+1))
			}
		}
	}
	return nil
}
