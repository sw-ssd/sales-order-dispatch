// Package handlers 提供認證 HTTP 端點與 RoleService 的 Connect handler 組裝。
package handlers

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
)

// RoleHandler 實作 salesorder.v1.RoleService(Connect 傳輸層;業務邏輯於 services.RoleService)。
// 身分/權限檢查在 service 內讀取 ctx(server authzMiddleware 注入),handler 不重複檢查。
type RoleHandler struct {
	salesorderv1connect.UnimplementedRoleServiceHandler
	svc *services.RoleService
}

// NewRoleHandler 建立 RoleHandler。
func NewRoleHandler(svc *services.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// ListRoles 分頁列出角色。
func (h *RoleHandler) ListRoles(ctx context.Context, req *connect.Request[v1.ListRolesRequest]) (*connect.Response[v1.ListRolesResponse], error) {
	return h.svc.ListRoles(ctx, req)
}

// GetRolePermissions 取得角色功能權限。
func (h *RoleHandler) GetRolePermissions(ctx context.Context, req *connect.Request[v1.GetRolePermissionsRequest]) (*connect.Response[v1.GetRolePermissionsResponse], error) {
	return h.svc.GetRolePermissions(ctx, req)
}

// UpdateRolePermissions 全量更新角色功能權限。
func (h *RoleHandler) UpdateRolePermissions(ctx context.Context, req *connect.Request[v1.UpdateRolePermissionsRequest]) (*connect.Response[v1.UpdateRolePermissionsResponse], error) {
	return h.svc.UpdateRolePermissions(ctx, req)
}

// RegisterRoleHandler 組裝 RoleService 的 Connect handler 並掛到 mux(T18;server 掛載點)。
// server 以 http.StripPrefix("/api/v1", mux) 掛載,與 RegisterCompanyServices 慣例一致。
func RegisterRoleHandler(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewRoleServiceHandler(NewRoleHandler(services.NewRoleService(db)))
	mux.Handle(path, handler)
}
