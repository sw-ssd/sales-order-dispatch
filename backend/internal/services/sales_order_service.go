// SalesOrderService 銷售訂單(05 計畫 Task 4, D10/D12/D13/D18)。
// CRUD + Cancel/Complete/Void + 軟刪除 + ListEvents。狀態轉移一律走 TransitionOrder,
// 建單組裝(換算/別名/守衛/順延)為 Task 5,本檔只做最小可用:客戶+來源+明細直寫。
package services

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderitem"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// SalesOrderService 實作 salesorder.v1.SalesOrderService。
type SalesOrderService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedSalesOrderServiceHandler
}

// NewSalesOrderService 建立 SalesOrderService。
func NewSalesOrderService(db *ent.Client) *SalesOrderService {
	return &SalesOrderService{db: db}
}

// RegisterSalesOrderService 掛到 /api/v1(租戶 session + RLS,比照 CustomerService)。
func RegisterSalesOrderService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewSalesOrderServiceHandler(NewSalesOrderService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// orderScopeQuery 依範圍對訂單查詢加入 company/department where。
func orderScopeQuery(q *ent.SalesOrderQuery, cid int, did *int) *ent.SalesOrderQuery {
	if did != nil {
		return q.Where(salesorder.CompanyIDEQ(cid), salesorder.DepartmentIDEQ(*did))
	}
	return q.Where(salesorder.CompanyIDEQ(cid))
}

// orderItemScopeQuery 明細同範圍(冗餘 company/department 欄)。
func orderItemScopeQuery(q *ent.SalesOrderItemQuery, cid int, did *int) *ent.SalesOrderItemQuery {
	if did != nil {
		return q.Where(salesorderitem.CompanyIDEQ(cid), salesorderitem.DepartmentIDEQ(*did))
	}
	return q.Where(salesorderitem.CompanyIDEQ(cid))
}

// ListOrders 分頁查詢,預設排除軟刪除。
func (s *SalesOrderService) ListOrders(ctx context.Context, req *connect.Request[salesorderv1.ListOrdersRequest]) (*connect.Response[salesorderv1.ListOrdersResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	q := orderScopeQuery(dbtenant.Client(ctx, s.db).SalesOrder.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(salesorder.DeletedAtIsNil())
	}
	if st := strings.TrimSpace(req.Msg.GetStatus()); st != "" {
		q = q.Where(salesorder.StatusEQ(st))
	}
	if c := strings.TrimSpace(req.Msg.GetCustomerId()); c != "" {
		custID, err := parseID(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(salesorder.CustomerIDEQ(custID))
	}
	if src := strings.TrimSpace(req.Msg.GetSource()); src != "" {
		q = q.Where(salesorder.SourceEQ(src))
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(salesorder.Or(
			salesorder.OrderNoContainsFold(kw),
			salesorder.NoteContainsFold(kw),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	orders, err := q.Order(ent.Desc(salesorder.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*salesorderv1.SalesOrder, 0, len(orders))
	for _, o := range orders {
		out = append(out, salesOrderToProto(o))
	}
	return connect.NewResponse(&salesorderv1.ListOrdersResponse{
		Orders: out, Total: int32(total),
	}), nil
}

// GetOrder 以 id 取單筆(含明細)。不可見或已軟刪除 → not_found(不洩漏存在性)。
func (s *SalesOrderService) GetOrder(ctx context.Context, req *connect.Request[salesorderv1.GetOrderRequest]) (*connect.Response[salesorderv1.GetOrderResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	oid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	o, err := orderScopeQuery(dbtenant.Client(ctx, s.db).SalesOrder.Query(), cid, did).
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := orderItemScopeQuery(dbtenant.Client(ctx, s.db).SalesOrderItem.Query(), cid, did).
		Where(salesorderitem.SalesOrderIDEQ(oid), salesorderitem.DeletedAtIsNil()).
		Order(ent.Asc(salesorderitem.FieldSortOrder), ent.Asc(salesorderitem.FieldID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*salesorderv1.SalesOrderItem, 0, len(items))
	for _, it := range items {
		out = append(out, salesOrderItemToProto(it))
	}
	return connect.NewResponse(&salesorderv1.GetOrderResponse{
		Order: salesOrderToProto(o), Items: out,
	}), nil
}

// CreateOrder 建立訂單(最小可用):驗客戶可見 → 驗來源 → 取號 → 建單+明細 → 寫 create 事件,同一交易。
// 組裝邏輯(4.2):客戶守衛 → 明細驗證(含換算) → 順延 → 取號,同一交易。
func (s *SalesOrderService) CreateOrder(ctx context.Context, req *connect.Request[salesorderv1.CreateOrderRequest]) (*connect.Response[salesorderv1.CreateOrderResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	isCustomer := isCustomerIdentity(id.Roles)
	// 客戶守衛(4.2.3):客戶帳號強制為自己;守衛在取號之前,不消耗序號。
	custID, err := orderCustomerGuard(req.Msg.GetCustomerId(), id.CustomerID, isCustomer)
	if err != nil {
		return nil, err
	}
	// 客戶存在且呼叫者可見(RLS self 為最後防線)。
	if _, err := customerScopeQuery(db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	source := strings.TrimSpace(req.Msg.GetSource())
	if err := validateOrderSource(ctx, db, source); err != nil {
		return nil, err
	}
	if len(req.Msg.GetItems()) == 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "items"})
	}
	if err := ensureOrderCounter(ctx, db, cid, source); err != nil {
		return nil, err
	}
	orderNo, err := nextOrderNo(ctx, db, cid, source)
	if err != nil {
		return nil, err
	}
	build := db.SalesOrder.Create().
		SetCompanyID(cid).SetOrderNo(orderNo).SetCustomerID(custID).
		SetSource(source).SetStatus(OrderStatusPending).SetVersion(1)
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if d := strings.TrimSpace(req.Msg.GetExpectedDeliveryDate()); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "expected_delivery_date"})
		}
		// 偏好送貨日順延(4.2.5):回應值即順延後日期(單一事實來源)。
		if days, derr := customerPreferredDays(ctx, db, cid, custID); derr != nil {
			return nil, derr
		} else {
			t = adjustDeliveryDate(days, t)
		}
		build = build.SetExpectedDeliveryDate(t)
	}
	if n := strings.TrimSpace(req.Msg.GetNote()); n != "" {
		build = build.SetNote(n)
	}
	if r := strings.TrimSpace(req.Msg.GetSalesRepId()); r != "" {
		repID, err := parseID(r)
		if err != nil {
			return nil, err
		}
		build = build.SetSalesRepID(repID)
	}
	o, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	for i, item := range req.Msg.GetItems() {
		pid, display, baseQty, err := validateOrderItem(ctx, db, cid, isCustomer, item)
		if err != nil {
			return nil, err
		}
		ib := db.SalesOrderItem.Create().
			SetSalesOrderID(o.ID).SetCompanyID(cid).
			SetDisplayName(display).
			SetQty(strings.TrimSpace(item.GetQty())).SetUnit(strings.TrimSpace(item.GetUnit())).
			SetBaseQty(baseQty).SetSortOrder(i)
		if did != nil {
			ib = ib.SetDepartmentID(*did)
		}
		if pid != 0 {
			ib = ib.SetProductID(pid)
		}
		if _, err := ib.Save(ctx); err != nil {
			return nil, toConnectError(err)
		}
		// 別名 upsert(4.2.2):save_alias 且帶 product_id → 同交易 upsert;存在改別名,
		// 不存在則建(預設值帶本次 qty 語意由清單端承接)。唯一衝突不擋單。
		if item.GetSaveAlias() && pid != 0 {
			if err := upsertCustomerAlias(ctx, db, cid, custID, pid, display); err != nil {
				return nil, err
			}
		}
		// 選用總表商品自動加入清單(4.2.2):無此商品記錄即建預設列(別名=商品名)。
		if pid != 0 {
			if err := ensureCustomerListEntry(ctx, db, cid, custID, pid); err != nil {
				return nil, err
			}
		}
	}
	if _, err := db.SalesOrderEvent.Create().
		SetSalesOrderID(o.ID).SetCompanyID(cid).
		SetEventType("create").SetActorID(actorIDOf(id)).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	created, err := db.SalesOrder.Query().Where(salesorder.ID(o.ID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.CreateOrderResponse{Order: salesOrderToProto(created)}), nil
}

// UpdateOrder 僅 pending 可編輯(攜帶 version 樂觀鎖);明細非空即整單替換。
func (s *SalesOrderService) UpdateOrder(ctx context.Context, req *connect.Request[salesorderv1.UpdateOrderRequest]) (*connect.Response[salesorderv1.UpdateOrderResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	oid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	o, err := orderScopeQuery(db.SalesOrder.Query(), cid, did).
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if o.Status != OrderStatusPending {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{
			"field": "status", "from": o.Status, "to": o.Status,
		})
	}
	if int(req.Msg.GetVersion()) != o.Version {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "version"})
	}
	upd := db.SalesOrder.UpdateOneID(oid).Where(salesorder.VersionEQ(o.Version)).
		SetVersion(o.Version + 1)
	changedDate := false
	if d := strings.TrimSpace(req.Msg.GetExpectedDeliveryDate()); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "expected_delivery_date"})
		}
		if days, derr := customerPreferredDays(ctx, db, cid, o.CustomerID); derr != nil {
			return nil, derr
		} else {
			t = adjustDeliveryDate(days, t)
		}
		upd = upd.SetExpectedDeliveryDate(t)
		changedDate = true
	}
	changedNote := false
	if req.Msg.GetNote() != "" {
		upd = upd.SetNote(strings.TrimSpace(req.Msg.GetNote()))
		changedNote = true
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "version"})
		}
		return nil, toConnectError(err)
	}
	if items := req.Msg.GetItems(); len(items) > 0 {
		if _, err := db.SalesOrderItem.Update().
			Where(salesorderitem.SalesOrderIDEQ(oid), salesorderitem.DeletedAtIsNil()).
			SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
			return nil, toConnectError(err)
		}
		isCustomer := isCustomerIdentity(id.Roles)
		for i, item := range items {
			pid, display, baseQty, err := validateOrderItem(ctx, db, cid, isCustomer, item)
			if err != nil {
				return nil, err
			}
			ib := db.SalesOrderItem.Create().
				SetSalesOrderID(oid).SetCompanyID(cid).
				SetDisplayName(display).
				SetQty(strings.TrimSpace(item.GetQty())).SetUnit(strings.TrimSpace(item.GetUnit())).
				SetBaseQty(baseQty).SetSortOrder(i)
			if did != nil {
				ib = ib.SetDepartmentID(*did)
			}
			if pid != 0 {
				ib = ib.SetProductID(pid)
			}
			if _, err := ib.Save(ctx); err != nil {
				return nil, toConnectError(err)
			}
		}
	}
	if _, err := db.SalesOrderEvent.Create().
		SetSalesOrderID(oid).SetCompanyID(cid).
		SetEventType("edit").SetActorID(actorIDOf(id)).
		SetPayload(map[string]any{"date": changedDate, "note": changedNote}).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.UpdateOrderResponse{Order: salesOrderToProto(updated)}), nil
}

// CancelOrder 僅 pending 可取消(走狀態機)。
func (s *SalesOrderService) CancelOrder(ctx context.Context, req *connect.Request[salesorderv1.CancelOrderRequest]) (*connect.Response[salesorderv1.CancelOrderResponse], error) {
	return s.transition(ctx, req.Msg.GetId(), OrderStatusCancelled, "")
}

// CompleteOrder 僅 processing 可完成(走狀態機)。
func (s *SalesOrderService) CompleteOrder(ctx context.Context, req *connect.Request[salesorderv1.CompleteOrderRequest]) (*connect.Response[salesorderv1.CompleteOrderResponse], error) {
	o, err := s.transitionOrder(ctx, req.Msg.GetId(), OrderStatusCompleted, "")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.CompleteOrderResponse{Order: o}), nil
}

// VoidOrder 僅 completed 可作廢(dept_admin 以上 + 原因,走狀態機)。
func (s *SalesOrderService) VoidOrder(ctx context.Context, req *connect.Request[salesorderv1.VoidOrderRequest]) (*connect.Response[salesorderv1.VoidOrderResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !hasRole(id, "dept_admin") && !hasRole(id, "company_admin") && !isSuperIdentity(id) {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	o, err := s.transitionOrder(ctx, req.Msg.GetId(), OrderStatusVoided, req.Msg.GetReason())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.VoidOrderResponse{Order: o}), nil
}

// DeleteOrder 軟刪除(僅 pending/cancelled;連同明細軟刪除 + 寫稽核,同交易)。
func (s *SalesOrderService) DeleteOrder(ctx context.Context, req *connect.Request[salesorderv1.DeleteOrderRequest]) (*connect.Response[salesorderv1.DeleteOrderResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	oid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	o, err := orderScopeQuery(db.SalesOrder.Query(), cid, did).
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if o.Status != OrderStatusPending && o.Status != OrderStatusCancelled {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{
			"field": "status", "from": o.Status,
		})
	}
	now := time.Now().UTC()
	if _, err := db.SalesOrder.UpdateOneID(oid).SetDeletedAt(now).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if _, err := db.SalesOrderItem.Update().
		Where(salesorderitem.SalesOrderIDEQ(oid), salesorderitem.DeletedAtIsNil()).
		SetDeletedAt(now).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAuditBA(ctx, tx, "sales_order", "delete", oid, cid, did, actorIDOf(id),
		map[string]any{"status": o.Status}, map[string]any{"deleted": true}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.DeleteOrderResponse{}), nil
}

// ListOrderEvents 異動軌跡查詢(升序)。訂單不可見 → not_found(不洩漏存在性)。
func (s *SalesOrderService) ListOrderEvents(ctx context.Context, req *connect.Request[salesorderv1.ListOrderEventsRequest]) (*connect.Response[salesorderv1.ListOrderEventsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	oid, err := parseID(req.Msg.GetSalesOrderId())
	if err != nil {
		return nil, err
	}
	if _, err := orderScopeQuery(dbtenant.Client(ctx, s.db).SalesOrder.Query(), cid, did).
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	q := dbtenant.Client(ctx, s.db).SalesOrderEvent.Query().
		Where(salesorderevent.SalesOrderIDEQ(oid))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	events, err := q.Order(ent.Asc(salesorderevent.FieldCreatedAt), ent.Asc(salesorderevent.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*salesorderv1.SalesOrderEvent, 0, len(events))
	for _, e := range events {
		out = append(out, salesOrderEventToProto(e))
	}
	return connect.NewResponse(&salesorderv1.ListOrderEventsResponse{
		Events: out, Total: int32(total),
	}), nil
}

// transition 為 Cancel 的路徑:範圍內載入 → TransitionOrder → 重查回傳。
func (s *SalesOrderService) transition(ctx context.Context, rawID, to, reason string) (*connect.Response[salesorderv1.CancelOrderResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	oid, err := parseID(rawID)
	if err != nil {
		return nil, err
	}
	if _, err := orderScopeQuery(db.SalesOrder.Query(), cid, did).
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := TransitionOrder(ctx, db, OrderTransitionInput{
		OrderID: oid, To: to, ActorID: actorIDOf(id), Reason: reason,
	}); err != nil {
		return nil, err
	}
	o, err := db.SalesOrder.Query().Where(salesorder.ID(oid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.CancelOrderResponse{Order: salesOrderToProto(o)}), nil
}

// transitionOrder 為 Cancel/Complete/Void 的共用核心:範圍內載入 → TransitionOrder → 重查回 proto。
func (s *SalesOrderService) transitionOrder(ctx context.Context, rawID, to, reason string) (*salesorderv1.SalesOrder, error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	oid, err := parseID(rawID)
	if err != nil {
		return nil, err
	}
	if _, err := orderScopeQuery(db.SalesOrder.Query(), cid, did).
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := TransitionOrder(ctx, db, OrderTransitionInput{
		OrderID: oid, To: to, ActorID: actorIDOf(id), Reason: reason,
	}); err != nil {
		return nil, err
	}
	o, err := db.SalesOrder.Query().Where(salesorder.ID(oid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return salesOrderToProto(o), nil
}

// actorIDOf 由身分取數字 actor id(解析失敗回 0;稽核層會擋 0,此處不靜默)。
func actorIDOf(id authz.Identity) int {
	n, err := parseID(id.UserID)
	if err != nil {
		return 0
	}
	return n
}

func salesOrderToProto(o *ent.SalesOrder) *salesorderv1.SalesOrder {
	p := &salesorderv1.SalesOrder{
		Id:         strconv.FormatInt(int64(o.ID), 10),
		CompanyId:  strconv.FormatInt(int64(o.CompanyID), 10),
		OrderNo:    o.OrderNo,
		CustomerId: strconv.FormatInt(int64(o.CustomerID), 10),
		Source:     o.Source,
		Status:     o.Status,
		Version:    int32(o.Version),
		CreatedAt:  o.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  o.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if o.DepartmentID != nil {
		p.DepartmentId = strconv.FormatInt(int64(*o.DepartmentID), 10)
	}
	if o.ExpectedDeliveryDate != nil {
		p.ExpectedDeliveryDate = o.ExpectedDeliveryDate.Format("2006-01-02")
	}
	if o.SalesRepID != nil {
		p.SalesRepId = strconv.FormatInt(int64(*o.SalesRepID), 10)
	}
	if o.Note != "" {
		p.Note = o.Note
	}
	if o.DispatchedAt != nil {
		p.DispatchedAt = o.DispatchedAt.UTC().Format(time.RFC3339)
	}
	if o.DispatchedBy != nil {
		p.DispatchedBy = strconv.FormatInt(int64(*o.DispatchedBy), 10)
	}
	if o.RouteID != nil {
		p.RouteId = strconv.FormatInt(int64(*o.RouteID), 10)
	}
	if o.DeliverySequence != nil {
		p.DeliverySequence = int32(*o.DeliverySequence)
	}
	if o.DeletedAt != nil {
		p.DeletedAt = o.DeletedAt.UTC().Format(time.RFC3339)
	}
	return p
}

// salesOrderItemToProto 將 ent.SalesOrderItem 轉為 proto。
func salesOrderItemToProto(it *ent.SalesOrderItem) *salesorderv1.SalesOrderItem {
	p := &salesorderv1.SalesOrderItem{
		Id:          strconv.FormatInt(int64(it.ID), 10),
		DisplayName: it.DisplayName,
		Qty:         it.Qty,
		Unit:        it.Unit,
		BaseQty:     it.BaseQty,
		SortOrder:   int32(it.SortOrder),
	}
	if it.ProductID != nil {
		p.ProductId = strconv.FormatInt(int64(*it.ProductID), 10)
	}
	if it.ProcessingSpecID != nil {
		p.ProcessingSpecId = strconv.FormatInt(int64(*it.ProcessingSpecID), 10)
	}
	if it.SpecialCutNote != "" {
		p.SpecialCutNote = it.SpecialCutNote
	}
	if it.WarehouseID != nil {
		p.WarehouseId = strconv.FormatInt(int64(*it.WarehouseID), 10)
	}
	return p
}

// salesOrderEventToProto 將 ent.SalesOrderEvent 轉為 proto(payload 轉 JSON 字串)。
func salesOrderEventToProto(e *ent.SalesOrderEvent) *salesorderv1.SalesOrderEvent {
	p := &salesorderv1.SalesOrderEvent{
		Id:        strconv.FormatInt(int64(e.ID), 10),
		EventType: e.EventType,
		ActorId:   strconv.FormatInt(int64(e.ActorID), 10),
		CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
	}
	if e.Reason != "" {
		p.Reason = e.Reason
	}
	if len(e.Payload) > 0 {
		if raw, err := json.Marshal(e.Payload); err == nil {
			p.Payload = string(raw)
		}
	}
	return p
}
