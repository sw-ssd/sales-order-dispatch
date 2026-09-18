// RouteService 車次(04 計畫 3.4.2, D10/D18):部門級主檔 CRUD + 軟刪除/復原 + 分頁。
package services

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// RouteService 實作 masters.v1.RouteService。
type RouteService struct {
	db *ent.Client
	mastersv1connect.UnimplementedRouteServiceHandler
}

// NewRouteService 建立 RouteService。
func NewRouteService(db *ent.Client) *RouteService {
	return &RouteService{db: db}
}

// RegisterRouteService 將 RouteService 掛到 mux。
func RegisterRouteService(mux *http.ServeMux, db *ent.Client) {
	path, handler := mastersv1connect.NewRouteServiceHandler(NewRouteService(db))
	mux.Handle(path, handler)
}

func routeScopeQuery(q *ent.RouteQuery, cid int, did *int) *ent.RouteQuery {
	if did != nil {
		return q.Where(route.CompanyIDEQ(cid), route.DepartmentIDEQ(*did))
	}
	return q.Where(route.CompanyIDEQ(cid))
}

func routeToProto(r *ent.Route) *mastersv1.Route {
	p := &mastersv1.Route{
		Id: strconv.FormatInt(int64(r.ID), 10), CompanyId: strconv.FormatInt(int64(r.CompanyID), 10),
		Code: r.Code, Name: r.Name, Description: r.Description,
		SortOrder: int32(r.SortOrder), IsActive: r.IsActive,
	}
	if r.DepartmentID != nil {
		p.DepartmentId = strconv.FormatInt(int64(*r.DepartmentID), 10)
	}
	if !r.CreatedAt.IsZero() {
		p.CreatedAt = r.CreatedAt.Format(time.RFC3339)
	}
	if !r.UpdatedAt.IsZero() {
		p.UpdatedAt = r.UpdatedAt.Format(time.RFC3339)
	}
	if r.DeletedAt != nil {
		p.DeletedAt = r.DeletedAt.Format(time.RFC3339)
	}
	return p
}

func (s *RouteService) ListRoutes(ctx context.Context, req *connect.Request[mastersv1.ListRoutesRequest]) (*connect.Response[mastersv1.ListRoutesResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := routeScopeQuery(s.db.Route.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(route.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(route.Or(route.CodeContainsFold(kw), route.NameContainsFold(kw)))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Clone().Order(ent.Asc(route.FieldSortOrder)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*mastersv1.Route, 0, len(items))
	for _, r := range items {
		out = append(out, routeToProto(r))
	}
	return connect.NewResponse(&mastersv1.ListRoutesResponse{
		Routes: out, Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

func (s *RouteService) CreateRoute(ctx context.Context, req *connect.Request[mastersv1.CreateRouteRequest]) (*connect.Response[mastersv1.CreateRouteResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.Msg.GetCode())
	name := strings.TrimSpace(req.Msg.GetName())
	if code == "" || name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code 與 name 必填"))
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	build := tx.Route.Create().
		SetCompanyID(cid).SetCode(code).SetName(name).
		SetSortOrder(int(req.Msg.GetSortOrder())).SetIsActive(req.Msg.GetIsActive()).
		SetCreatedBy(actor).SetUpdatedBy(actor)
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if d := strings.TrimSpace(req.Msg.GetDescription()); d != "" {
		build = build.SetDescription(d)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "create", ResourceType: "route", ResourceID: strconv.FormatInt(int64(created.ID), 10),
		CompanyID: cid, DepartmentID: created.DepartmentID, UserID: actor,
		After:     map[string]any{"code": created.Code, "name": created.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateRouteResponse{Route: routeToProto(created)}), nil
}

func (s *RouteService) UpdateRoute(ctx context.Context, req *connect.Request[mastersv1.UpdateRouteRequest]) (*connect.Response[mastersv1.UpdateRouteResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route id 格式錯誤"))
	}
	if _, err := routeScopeQuery(s.db.Route.Query(), cid, did).Where(route.ID(rid), route.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	upd := tx.Route.UpdateOneID(rid)
	if req.Msg.Code != nil {
		c := strings.TrimSpace(*req.Msg.Code)
		if c == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code 不可為空"))
		}
		upd = upd.SetCode(c)
	}
	if req.Msg.Name != nil {
		n := strings.TrimSpace(*req.Msg.Name)
		if n == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 不可為空"))
		}
		upd = upd.SetName(n)
	}
	if req.Msg.Description != nil {
		upd = upd.SetDescription(*req.Msg.Description)
	}
	if req.Msg.SortOrder != nil {
		upd = upd.SetSortOrder(int(*req.Msg.SortOrder))
	}
	if req.Msg.IsActive != nil {
		upd = upd.SetIsActive(*req.Msg.IsActive)
	}
	actor, _ := parseID(id.UserID)
	upd = upd.SetUpdatedBy(actor)
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "update", ResourceType: "route", ResourceID: strconv.FormatInt(int64(rid), 10),
		CompanyID: cid, DepartmentID: updated.DepartmentID, UserID: actor,
		After:     map[string]any{"name": updated.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateRouteResponse{Route: routeToProto(updated)}), nil
}

func (s *RouteService) DeleteRoute(ctx context.Context, req *connect.Request[mastersv1.DeleteRouteRequest]) (*connect.Response[mastersv1.DeleteRouteResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route id 格式錯誤"))
	}
	cur, err := routeScopeQuery(s.db.Route.Query(), cid, did).Where(route.ID(rid), route.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.Route.UpdateOneID(rid).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "delete", ResourceType: "route", ResourceID: strconv.FormatInt(int64(rid), 10),
		CompanyID: cid, DepartmentID: cur.DepartmentID, UserID: actor,
		Before:    map[string]any{"code": cur.Code, "name": cur.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteRouteResponse{}), nil
}

func (s *RouteService) RestoreRoute(ctx context.Context, req *connect.Request[mastersv1.RestoreRouteRequest]) (*connect.Response[mastersv1.RestoreRouteResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route id 格式錯誤"))
	}
	cur, err := routeScopeQuery(s.db.Route.Query(), cid, did).Where(route.ID(rid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		return connect.NewResponse(&mastersv1.RestoreRouteResponse{Route: routeToProto(cur)}), nil
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	restored, err := tx.Route.UpdateOneID(rid).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "update", ResourceType: "route", ResourceID: strconv.FormatInt(int64(rid), 10),
		CompanyID: cid, DepartmentID: restored.DepartmentID, UserID: actor,
		After:     map[string]any{"restored": true, "code": restored.Code},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreRouteResponse{Route: routeToProto(restored)}), nil
}
