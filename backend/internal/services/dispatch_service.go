// DispatchService 派車 API(08 計畫 Task 5.1)。
// AssignRoute 看板拖放(樂觀鎖 version + 順位重排);Confirm 批次轉 processing(逐筆交易,
// 部分失敗語義);CancelDispatch 退回 pending(保留看板位置 + 重印警告)。
// 事件經 publishAfterCommit 於提交後發佈;派車通知經 OnDispatchConfirmed 於提交後發送。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// DispatchService 為派車服務。
type DispatchService struct {
	db *ent.Client
}

// NewDispatchService 建立 DispatchService。
func NewDispatchService(db *ent.Client) *DispatchService {
	return &DispatchService{db: db}
}

// RegisterDispatchService 掛到 /api/v1(租戶 session + RLS)。
func RegisterDispatchService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewDispatchServiceHandler(NewDispatchService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// AssignRoute 指派車次與配送順位(僅 pending;version 樂觀鎖 + 同車次順位重排)。
func (s *DispatchService) AssignRoute(ctx context.Context, req *connect.Request[salesorderv1.AssignRouteRequest]) (*connect.Response[salesorderv1.AssignRouteResponse], error) {
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
		return nil, invalidArgField("sales_order_id")
	}
	expVer, err := parseVersion(req.Msg.GetVersion())
	if err != nil {
		return nil, invalidArgField("version")
	}
	day, err := parseBoardDate(req.Msg.GetExpectedDeliveryDate())
	if err != nil {
		return nil, err
	}
	var routeID *int
	var seq *int
	if rid := strings.TrimSpace(req.Msg.GetRouteId()); rid != "" {
		n, err := parseID(rid)
		if err != nil {
			return nil, invalidArgField("route_id")
		}
		routeID = &n
		s, err := parseSeq(req.Msg.GetDeliverySequence())
		if err != nil {
			return nil, err
		}
		seq = &s
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	o, err := db.SalesOrder.Query().
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if o.CompanyID != cid || (did != nil && !equalDept(o.DepartmentID, did)) {
		return nil, errcode.SysNotFound.Error(nil)
	}
	if o.Status != OrderStatusPending {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "僅待派訂單可指派"})
	}
	if o.Version != expVer {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "資料已變更，請重新載入"})
	}
	if routeID != nil {
		r, err := db.Route.Query().Where(route.ID(*routeID)).Only(ctx)
		if err != nil {
			return nil, toConnectError(err)
		}
		if r.CompanyID != cid || !equalDept(r.DepartmentID, did) {
			return nil, errcode.SysNotFound.Error(nil)
		}
	}
	// 同車次順位重排:目標車次同日同部門 pending 訂單中 seq ≥ 新順位者依序後移。
	if routeID != nil && seq != nil {
		if err := shiftSequence(ctx, db, cid, did, *routeID, day, *seq, oid); err != nil {
			return nil, toConnectError(err)
		}
	}
	upd := db.SalesOrder.UpdateOneID(o.ID).Where(salesorder.VersionEQ(expVer)).
		SetVersion(o.Version + 1)
	if routeID == nil {
		upd = upd.ClearRouteID().ClearDeliverySequence()
	} else {
		upd = upd.SetRouteID(*routeID).SetDeliverySequence(*seq)
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if updated == nil {
		// 條件更新 0 列 → 樂觀鎖衝突(併發已改)。
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "資料已變更，請重新載入"})
	}
	// 事件:route_assign(含新舊 route/seq)。
	if err := writeAssignEvent(ctx, db, actorIDOf(id), o, routeID, seq); err != nil {
		return nil, toConnectError(err)
	}
	var deptID int
	if updated.DepartmentID != nil {
		deptID = *updated.DepartmentID
	}
	if err := publishAfterCommit(ctx, deptID, BoardEvent{Type: "route_assign",
		SalesOrderID: updated.ID, RouteID: updated.RouteID,
		DeliverySequence: updated.DeliverySequence, Version: updated.Version,
		DepartmentID: deptID}); err != nil {
		return nil, toConnectError(err)
	}
	resp := &salesorderv1.AssignRouteResponse{
		SalesOrderId: strconv.Itoa(updated.ID), Version: strconv.Itoa(updated.Version),
	}
	if updated.RouteID != nil {
		resp.RouteId = strconv.Itoa(*updated.RouteID)
	}
	if updated.DeliverySequence != nil {
		resp.DeliverySequence = strconv.Itoa(*updated.DeliverySequence)
	}
	return connect.NewResponse(resp), nil
}

// ConfirmDispatch 車次批次確認(逐筆 pending → processing;部分失敗語義)。
func (s *DispatchService) ConfirmDispatch(ctx context.Context, req *connect.Request[salesorderv1.ConfirmDispatchRequest]) (*connect.Response[salesorderv1.ConfirmDispatchResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetRouteId())
	if err != nil {
		return nil, invalidArgField("route_id")
	}
	day, err := parseBoardDate(req.Msg.GetExpectedDeliveryDate())
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	if _, err := scopedRoute(ctx, db, cid, did, rid); err != nil {
		return nil, err
	}
	cands, err := confirmCandidates(ctx, db, cid, did, rid, day)
	if err != nil {
		return nil, toConnectError(err)
	}
	if len(cands) == 0 {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "該車次當日無待派訂單"})
	}
	actor, _ := parseID(id.UserID)
	shared := time.Now().UTC()
	resp := &salesorderv1.ConfirmDispatchResponse{}
	for _, c := range cands {
		// 逐筆獨立交易语意:本筆條件更新(pending → processing) + 事件;失敗记筆不挡他筆。
		// 同一請求交易內逐筆條件更新(WHERE status=pending),天然部分失敗。
		err := TransitionOrder(ctx, db, OrderTransitionInput{
			OrderID: c.ID, To: OrderStatusProcessing, ActorID: actor,
			DispatchedBy: &actor, RouteID: &c.RouteIDV, DeliverySequence: c.SeqV,
		})
		item := &salesorderv1.DispatchItemResult{SalesOrderId: strconv.Itoa(c.ID)}
		if err != nil {
			item.Success = false
			item.FailReason = failReasonOf(err)
		} else {
			item.Success = true
			resp.SuccessCount++
			var deptID int
			if c.DeptV != nil {
				deptID = *c.DeptV
			}
			_ = publishAfterCommit(ctx, deptID, BoardEvent{Type: "dispatch",
				SalesOrderID: c.ID, RouteID: &c.RouteIDV,
				DeliverySequence: c.SeqV, Version: c.VersionV + 1, DepartmentID: deptID})
			_ = shared
		}
		resp.Items = append(resp.Items, item)
	}
	// 派車通知:逐筆成功訂單於同交易建 pending + 提交後發送(07 觸發 5.1.4;fire-and-record)。
	routeName := ""
	if r, err := db.Route.Query().Where(route.ID(rid)).Only(ctx); err == nil {
		routeName = r.Name
	}
	for _, it := range resp.Items {
		if !it.Success {
			continue
		}
		oid, _ := parseID(it.GetSalesOrderId())
		o, err := db.SalesOrder.Query().Where(salesorder.ID(oid)).Only(ctx)
		if err != nil {
			continue
		}
		date := ""
		if o.ExpectedDeliveryDate != nil {
			date = o.ExpectedDeliveryDate.Format("2006-01-02")
		}
		if err := OnDispatchConfirmed(ctx, db, cid, did, o.CustomerID, o.ID, o.OrderNo, routeName, date); err != nil {
			continue // fire-and-record:通知失敗不影響批次結果
		}
	}
	return connect.NewResponse(resp), nil
}

// CancelDispatch 取消派車(僅 processing → pending;dept_admin 以上;重印警告)。
func (s *DispatchService) CancelDispatch(ctx context.Context, req *connect.Request[salesorderv1.CancelDispatchRequest]) (*connect.Response[salesorderv1.CancelDispatchResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !hasRole(id, "dept_admin") && !hasRole(id, "company_admin") && !isSuperIdentity(id) {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	oid, err := parseID(req.Msg.GetSalesOrderId())
	if err != nil {
		return nil, invalidArgField("sales_order_id")
	}
	reason := strings.TrimSpace(req.Msg.GetReason())
	if reason == "" {
		return nil, invalidArgField("reason")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	o, err := db.SalesOrder.Query().
		Where(salesorder.ID(oid), salesorder.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if o.CompanyID != cid || (did != nil && !equalDept(o.DepartmentID, did)) {
		return nil, errcode.SysNotFound.Error(nil)
	}
	if o.Status != OrderStatusProcessing {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "僅處理中可取消派車"})
	}
	// 重印檢查:該車次當日正式列印記錄(print_logs)。
	warn := false
	if o.RouteID != nil && o.ExpectedDeliveryDate != nil {
		exists, err := db.PrintLog.Query().
			Where(printlog.RouteIDEQ(*o.RouteID),
				printlog.TargetDateEQ(o.ExpectedDeliveryDate.Truncate(24*time.Hour))).Exist(ctx)
		if err != nil {
			return nil, toConnectError(err)
		}
		warn = exists
		if warn && !req.Msg.GetAcknowledgeReprint() {
			resp := &salesorderv1.CancelDispatchResponse{
				SalesOrderId: strconv.Itoa(o.ID), Status: o.Status, ReprintWarning: true,
			}
			if o.RouteID != nil {
				resp.RouteId = strconv.Itoa(*o.RouteID)
			}
			if o.DeliverySequence != nil {
				resp.DeliverySequence = strconv.Itoa(*o.DeliverySequence)
			}
			return connect.NewResponse(resp), nil
		}
	}
	actor, _ := parseID(id.UserID)
	if err := TransitionOrder(ctx, db, OrderTransitionInput{
		OrderID: o.ID, To: OrderStatusPending, ActorID: actor, Reason: reason,
	}); err != nil {
		return nil, toConnectError(err)
	}
	updated, err := db.SalesOrder.Query().Where(salesorder.ID(o.ID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	var deptID int
	if updated.DepartmentID != nil {
		deptID = *updated.DepartmentID
	}
	if err := publishAfterCommit(ctx, deptID, BoardEvent{Type: "dispatch_cancel",
		SalesOrderID: updated.ID, RouteID: updated.RouteID,
		DeliverySequence: updated.DeliverySequence, Version: updated.Version,
		DepartmentID: deptID}); err != nil {
		return nil, toConnectError(err)
	}
	resp := &salesorderv1.CancelDispatchResponse{
		SalesOrderId: strconv.Itoa(updated.ID), Status: updated.Status, ReprintWarning: warn,
	}
	if updated.RouteID != nil {
		resp.RouteId = strconv.Itoa(*updated.RouteID)
	}
	if updated.DeliverySequence != nil {
		resp.DeliverySequence = strconv.Itoa(*updated.DeliverySequence)
	}
	return connect.NewResponse(resp), nil
}
