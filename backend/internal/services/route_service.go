// RouteService 車次(04 計畫 3.4.2, D10/D18):部門級主檔 CRUD + 軟刪除/復原 + 分頁。
// 共用流程見 master_crud.go。
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
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
)

// RouteService 實作 masters.v1.RouteService。
type RouteService struct {
	db *ent.Client
	mastersv1connect.UnimplementedRouteServiceHandler
}

// NewRouteService 建立 RouteService。
func NewRouteService(db *ent.Client) *RouteService { return &RouteService{db: db} }

// RegisterRouteService 將 RouteService 掛到 mux。
func RegisterRouteService(mux *http.ServeMux, db *ent.Client) {
	path, handler := mastersv1connect.NewRouteServiceHandler(NewRouteService(db), connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
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

type routeListSource struct{ q *ent.RouteQuery }

func (s routeListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s routeListSource) Page(ctx context.Context, off, lim int) ([]*ent.Route, error) {
	// F2(與 F1 同型):sort_order 預設 0、同值群常遠大於一頁,排序鍵非唯一時 PostgreSQL 對
	// 同值群(ties)的順序不保證一致,逐頁 LIMIT/OFFSET 會重複與遺漏資料,故以 id 為次序鍵。
	return s.q.Clone().Order(ent.Asc(route.FieldSortOrder), ent.Asc(route.FieldID)).Offset(off).Limit(lim).All(ctx)
}

func (s *RouteService) ListRoutes(ctx context.Context, req *connect.Request[mastersv1.ListRoutesRequest]) (*connect.Response[mastersv1.ListRoutesResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	q := routeScopeQuery(dbtenant.Client(ctx, s.db).Route.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(route.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(route.Or(route.CodeContainsFold(kw), route.NameContainsFold(kw)))
	}
	list, pg, err := pageList(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), routeListSource{q}, routeToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mastersv1.ListRoutesResponse{Routes: list, Pagination: pg}), nil
}

func (s *RouteService) CreateRoute(ctx context.Context, req *connect.Request[mastersv1.CreateRouteRequest]) (*connect.Response[mastersv1.CreateRouteResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	code, name, err := codeName(req.Msg.GetCode(), req.Msg.GetName())
	if err != nil {
		return nil, err
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	build := db.Route.Create().
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
	if err := recordAudit(ctx, tx, "route", "create", created.ID, cid, created.DepartmentID, actor, map[string]any{"code": created.Code, "name": created.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateRouteResponse{Route: routeToProto(created)}), nil
}

func (s *RouteService) UpdateRoute(ctx context.Context, req *connect.Request[mastersv1.UpdateRouteRequest]) (*connect.Response[mastersv1.UpdateRouteResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route id 格式錯誤"))
	}
	if _, err := routeScopeQuery(dbtenant.Client(ctx, s.db).Route.Query(), cid, did).Where(route.ID(rid), route.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	upd := db.Route.UpdateOneID(rid)
	if req.Msg.Code != nil {
		c, err := trimNonEmpty(*req.Msg.Code, "code 不可為空")
		if err != nil {
			return nil, err
		}
		upd = upd.SetCode(c)
	}
	if req.Msg.Name != nil {
		n, err := trimNonEmpty(*req.Msg.Name, "name 不可為空")
		if err != nil {
			return nil, err
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
	updated, err := upd.SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "route", "update", rid, cid, updated.DepartmentID, actor, map[string]any{"name": updated.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateRouteResponse{Route: routeToProto(updated)}), nil
}

func (s *RouteService) DeleteRoute(ctx context.Context, req *connect.Request[mastersv1.DeleteRouteRequest]) (*connect.Response[mastersv1.DeleteRouteResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route id 格式錯誤"))
	}
	cur, err := routeScopeQuery(dbtenant.Client(ctx, s.db).Route.Query(), cid, did).Where(route.ID(rid), route.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	if err := db.Route.UpdateOneID(rid).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "route", "delete", rid, cid, cur.DepartmentID, actor, map[string]any{"code": cur.Code, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteRouteResponse{}), nil
}

func (s *RouteService) RestoreRoute(ctx context.Context, req *connect.Request[mastersv1.RestoreRouteRequest]) (*connect.Response[mastersv1.RestoreRouteResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route id 格式錯誤"))
	}
	cur, err := routeScopeQuery(dbtenant.Client(ctx, s.db).Route.Query(), cid, did).Where(route.ID(rid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		return connect.NewResponse(&mastersv1.RestoreRouteResponse{Route: routeToProto(cur)}), nil
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	restored, err := db.Route.UpdateOneID(rid).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "route", "update", rid, cid, restored.DepartmentID, actor, map[string]any{"restored": true, "code": restored.Code}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreRouteResponse{Route: routeToProto(restored)}), nil
}
