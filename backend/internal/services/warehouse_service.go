// WarehouseService 倉別(04 計畫 3.4.1, D10/D18):部門級主檔 CRUD + 軟刪除/復原 + 分頁。
// 寫入 company_id/department_id 由 session 租戶注入(不接受 Request 指定跨部門值)。
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
	"github.com/salesorder/sales-order-1.0/backend/ent/warehouse"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// WarehouseService 實作 masters.v1.WarehouseService。
type WarehouseService struct {
	db *ent.Client
	mastersv1connect.UnimplementedWarehouseServiceHandler
}

// NewWarehouseService 建立 WarehouseService。
func NewWarehouseService(db *ent.Client) *WarehouseService {
	return &WarehouseService{db: db}
}

// RegisterWarehouseService 將 WarehouseService 掛到 mux。
func RegisterWarehouseService(mux *http.ServeMux, db *ent.Client) {
	path, handler := mastersv1connect.NewWarehouseServiceHandler(NewWarehouseService(db))
	mux.Handle(path, handler)
}

// warehouseScopeQuery 依範圍對倉別查詢加入 company/department where。
func warehouseScopeQuery(q *ent.WarehouseQuery, cid int, did *int) *ent.WarehouseQuery {
	if did != nil {
		return q.Where(warehouse.CompanyIDEQ(cid), warehouse.DepartmentIDEQ(*did))
	}
	return q.Where(warehouse.CompanyIDEQ(cid))
}

// warehouseToProto 將 ent.Warehouse 轉為 proto Warehouse。
func warehouseToProto(w *ent.Warehouse) *mastersv1.Warehouse {
	p := &mastersv1.Warehouse{
		Id:        strconv.FormatInt(int64(w.ID), 10),
		CompanyId: strconv.FormatInt(int64(w.CompanyID), 10),
		Code:      w.Code,
		Name:      w.Name,
		Address:   w.Address,
		IsActive:  w.IsActive,
	}
	if w.DepartmentID != nil {
		p.DepartmentId = strconv.FormatInt(int64(*w.DepartmentID), 10)
	}
	if !w.CreatedAt.IsZero() {
		p.CreatedAt = w.CreatedAt.Format(time.RFC3339)
	}
	if !w.UpdatedAt.IsZero() {
		p.UpdatedAt = w.UpdatedAt.Format(time.RFC3339)
	}
	if w.DeletedAt != nil {
		p.DeletedAt = w.DeletedAt.Format(time.RFC3339)
	}
	return p
}

// ListWarehouses 分頁列出本部門(或公司)倉別;keyword 對 code/name 模糊比對;可 include_deleted。
func (s *WarehouseService) ListWarehouses(ctx context.Context, req *connect.Request[mastersv1.ListWarehousesRequest]) (*connect.Response[mastersv1.ListWarehousesResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := warehouseScopeQuery(s.db.Warehouse.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(warehouse.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(warehouse.Or(warehouse.CodeContainsFold(kw), warehouse.NameContainsFold(kw)))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Clone().Order(ent.Asc(warehouse.FieldCode)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*mastersv1.Warehouse, 0, len(items))
	for _, w := range items {
		out = append(out, warehouseToProto(w))
	}
	return connect.NewResponse(&mastersv1.ListWarehousesResponse{
		Warehouses: out,
		Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

// CreateWarehouse 建立倉別:租戶注入 + code 部門唯一 + 稽核(同一交易)。
func (s *WarehouseService) CreateWarehouse(ctx context.Context, req *connect.Request[mastersv1.CreateWarehouseRequest]) (*connect.Response[mastersv1.CreateWarehouseResponse], error) {
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
	build := tx.Warehouse.Create().
		SetCompanyID(cid).SetCode(code).SetName(name).
		SetIsActive(req.Msg.GetIsActive()).SetCreatedBy(actor).SetUpdatedBy(actor)
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if a := strings.TrimSpace(req.Msg.GetAddress()); a != "" {
		build = build.SetAddress(a)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "create", ResourceType: "warehouse", ResourceID: strconv.FormatInt(int64(created.ID), 10),
		CompanyID: cid, DepartmentID: created.DepartmentID, UserID: actor,
		After:     map[string]any{"code": created.Code, "name": created.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateWarehouseResponse{Warehouse: warehouseToProto(created)}), nil
}

// UpdateWarehouse 欄位式更新(code 可改,部門唯一重驗)。
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, req *connect.Request[mastersv1.UpdateWarehouseRequest]) (*connect.Response[mastersv1.UpdateWarehouseResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	wid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("warehouse id 格式錯誤"))
	}
	if _, err := warehouseScopeQuery(s.db.Warehouse.Query(), cid, did).Where(warehouse.ID(wid), warehouse.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	upd := tx.Warehouse.UpdateOneID(wid)
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
	if req.Msg.Address != nil {
		upd = upd.SetAddress(*req.Msg.Address)
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
		Action: "update", ResourceType: "warehouse", ResourceID: strconv.FormatInt(int64(wid), 10),
		CompanyID: cid, DepartmentID: updated.DepartmentID, UserID: actor,
		After:     map[string]any{"name": updated.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateWarehouseResponse{Warehouse: warehouseToProto(updated)}), nil
}

// DeleteWarehouse 軟刪除 + 稽核(同一交易)。
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, req *connect.Request[mastersv1.DeleteWarehouseRequest]) (*connect.Response[mastersv1.DeleteWarehouseResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	wid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("warehouse id 格式錯誤"))
	}
	cur, err := warehouseScopeQuery(s.db.Warehouse.Query(), cid, did).Where(warehouse.ID(wid), warehouse.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.Warehouse.UpdateOneID(wid).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "delete", ResourceType: "warehouse", ResourceID: strconv.FormatInt(int64(wid), 10),
		CompanyID: cid, DepartmentID: cur.DepartmentID, UserID: actor,
		Before:    map[string]any{"code": cur.Code, "name": cur.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteWarehouseResponse{}), nil
}

// RestoreWarehouse 復原(清 deleted_at + 稽核;已刪除才動作,冪等)。
func (s *WarehouseService) RestoreWarehouse(ctx context.Context, req *connect.Request[mastersv1.RestoreWarehouseRequest]) (*connect.Response[mastersv1.RestoreWarehouseResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	wid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("warehouse id 格式錯誤"))
	}
	cur, err := warehouseScopeQuery(s.db.Warehouse.Query(), cid, did).Where(warehouse.ID(wid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		return connect.NewResponse(&mastersv1.RestoreWarehouseResponse{Warehouse: warehouseToProto(cur)}), nil
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	restored, err := tx.Warehouse.UpdateOneID(wid).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "update", ResourceType: "warehouse", ResourceID: strconv.FormatInt(int64(wid), 10),
		CompanyID: cid, DepartmentID: restored.DepartmentID, UserID: actor,
		After:     map[string]any{"restored": true, "code": restored.Code},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreWarehouseResponse{Warehouse: warehouseToProto(restored)}), nil
}
