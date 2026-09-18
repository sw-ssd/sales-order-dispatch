// WarehouseService 倉別(04 計畫 3.4.1, D10/D18):部門級主檔 CRUD + 軟刪除/復原 + 分頁。
// 共用流程(身分/驗證/分頁/稽核)見 master_crud.go;本檔僅列 per-entity 橋接與 builder。
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
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
)

// WarehouseService 實作 masters.v1.WarehouseService。
type WarehouseService struct {
	db *ent.Client
	mastersv1connect.UnimplementedWarehouseServiceHandler
}

// NewWarehouseService 建立 WarehouseService。
func NewWarehouseService(db *ent.Client) *WarehouseService { return &WarehouseService{db: db} }

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

// warehouseListSource 為 masterPage 的 ent 查詢橋接。
type warehouseListSource struct{ q *ent.WarehouseQuery }

func (s warehouseListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s warehouseListSource) Page(ctx context.Context, off, lim int) ([]*ent.Warehouse, error) {
	return s.q.Clone().Order(ent.Asc(warehouse.FieldCode)).Offset(off).Limit(lim).All(ctx)
}

// ListWarehouses 分頁列出本部門(或公司)倉別;keyword 對 code/name 模糊比對;可 include_deleted。
func (s *WarehouseService) ListWarehouses(ctx context.Context, req *connect.Request[mastersv1.ListWarehousesRequest]) (*connect.Response[mastersv1.ListWarehousesResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	q := warehouseScopeQuery(s.db.Warehouse.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(warehouse.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(warehouse.Or(warehouse.CodeContainsFold(kw), warehouse.NameContainsFold(kw)))
	}
	list, pg, err := masterPage(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), warehouseListSource{q}, warehouseToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mastersv1.ListWarehousesResponse{Warehouses: list, Pagination: pg}), nil
}

// CreateWarehouse 建立倉別:租戶注入 + code 部門唯一 + 稽核(同一交易)。
func (s *WarehouseService) CreateWarehouse(ctx context.Context, req *connect.Request[mastersv1.CreateWarehouseRequest]) (*connect.Response[mastersv1.CreateWarehouseResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	code, name, err := masterCodeName(req.Msg.GetCode(), req.Msg.GetName())
	if err != nil {
		return nil, err
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
	if err := recordMasterAudit(ctx, tx, "warehouse", "create", created.ID, cid, created.DepartmentID, actor, map[string]any{"code": created.Code, "name": created.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateWarehouseResponse{Warehouse: warehouseToProto(created)}), nil
}

// UpdateWarehouse 欄位式更新(code 可改,部門唯一重驗)。
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, req *connect.Request[mastersv1.UpdateWarehouseRequest]) (*connect.Response[mastersv1.UpdateWarehouseResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
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
		c, err := masterTrimNonEmpty(*req.Msg.Code, "code 不可為空")
		if err != nil {
			return nil, err
		}
		upd = upd.SetCode(c)
	}
	if req.Msg.Name != nil {
		n, err := masterTrimNonEmpty(*req.Msg.Name, "name 不可為空")
		if err != nil {
			return nil, err
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
	updated, err := upd.SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordMasterAudit(ctx, tx, "warehouse", "update", wid, cid, updated.DepartmentID, actor, map[string]any{"name": updated.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateWarehouseResponse{Warehouse: warehouseToProto(updated)}), nil
}

// DeleteWarehouse 軟刪除 + 稽核(同一交易)。
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, req *connect.Request[mastersv1.DeleteWarehouseRequest]) (*connect.Response[mastersv1.DeleteWarehouseResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
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
	if err := recordMasterAudit(ctx, tx, "warehouse", "delete", wid, cid, cur.DepartmentID, actor, map[string]any{"code": cur.Code, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteWarehouseResponse{}), nil
}

// RestoreWarehouse 復原(清 deleted_at + 稽核;已刪除才動作,冪等)。
func (s *WarehouseService) RestoreWarehouse(ctx context.Context, req *connect.Request[mastersv1.RestoreWarehouseRequest]) (*connect.Response[mastersv1.RestoreWarehouseResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
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
	if err := recordMasterAudit(ctx, tx, "warehouse", "update", wid, cid, restored.DepartmentID, actor, map[string]any{"restored": true, "code": restored.Code}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreWarehouseResponse{Warehouse: warehouseToProto(restored)}), nil
}
