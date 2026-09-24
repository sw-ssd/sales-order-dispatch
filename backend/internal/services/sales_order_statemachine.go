// Package services 的訂單狀態機(05 計畫 Task 3, D13):`pending ⇄ processing → completed`、
// `pending → cancelled`、`completed → voided`;終態 cancelled/voided 無出口。
// 所有轉移走 Transition(條件更新 + 事件 + 必要時稽核,同一交易);各 RPC 不得自行改 status。
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// 訂單狀態(D13)。
const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusCompleted  = "completed"
	OrderStatusCancelled  = "cancelled"
	OrderStatusVoided     = "voided"
)

// 訂單事件型別(D13),與 sales_order_events.event_type 同值。
const (
	OrderEventDispatch       = "dispatch"
	OrderEventDispatchCancel = "dispatch_cancel"
	OrderEventCancel         = "cancel"
	OrderEventComplete       = "complete"
	OrderEventVoid           = "void"
)

// allowedOrderTransitions 為合法轉移表:不在表上的一律拒絕。表驅動是為了讓「非法轉移」
// 有一個地方可稽核,而不是散在各個呼叫點的 if(比照 billing.allowedTransitions)。
var allowedOrderTransitions = map[string][]string{
	OrderStatusPending:    {OrderStatusProcessing, OrderStatusCancelled},
	OrderStatusProcessing: {OrderStatusPending, OrderStatusCompleted},
	OrderStatusCompleted:  {OrderStatusVoided},
	OrderStatusCancelled:  {},
	OrderStatusVoided:     {},
}

// CanOrderTransition 回報 from → to 是否為合法轉移。
func CanOrderTransition(from, to string) bool {
	for _, next := range allowedOrderTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// OrderTransitionInput 為一次狀態轉移的參數。Reason 僅 dispatch_cancel/void 必填;
// DispatchedBy/RouteID/DeliverySequence 為 dispatch 時的派車欄位(dispatch_cancel 保留
// route 看板位置,故不清 route 欄)。
//
// MarkDelivered 供 10.9 送達回寫使用:標記 delivered_at。與手動完成
// (SalesOrderService.CompleteOrder,店家自行確認完成)的差別在於**誰促成**——
// 兩者都走同一條狀態機,但只有回寫會蓋送達時點;同時事件 payload 帶 source=delivery,
// 讓事後追查分得出「司機完成配送」與「店家手動結案」。
// 只設 delivered_at 而不改狀態的呼叫(如對已完成訂單補標)不被支援 —— 狀態機是唯一入口。
type OrderTransitionInput struct {
	OrderID          int
	To               string
	ActorID          int
	Reason           string
	DispatchedBy     *int
	RouteID          *int
	DeliverySequence *int
	MarkDelivered    bool
	DeliveredAt      *time.Time
}

// TransitionOrder 執行一次訂單狀態轉移:條件更新(WHERE status=前值) + 事件 +
// 必要時稽核(dispatch_cancel/void),同一交易。呼叫端須傳入請求交易的 ent client(db):
// 交易邊界由呼叫端擁有(比照 SetCompanyStatus)。
//
// 角色檢查由呼叫端(授權層)執行,此處只守狀態機與原因必填。version 每次轉移 +1。
func TransitionOrder(ctx context.Context, db *ent.Client, in OrderTransitionInput) error {
	to := strings.TrimSpace(in.To)
	needsReason := to == OrderStatusPending || to == OrderStatusVoided
	if needsReason && strings.TrimSpace(in.Reason) == "" {
		return errcode.SysInvalidArgument.Error(map[string]string{"field": "reason"})
	}
	// 讀前值(含 company/department,事件與稽核用;軟刪除列不可轉移)。
	order, err := db.SalesOrder.Query().
		Where(salesorder.ID(in.OrderID), salesorder.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.SysNotFound.Error(nil)
		}
		return errcode.SysInternal.Wrap(err)
	}
	from := order.Status
	if !CanOrderTransition(from, to) {
		return errcode.SysInvalidArgument.Error(map[string]string{
			"field": "status", "from": from, "to": to,
		})
	}
	// 條件更新:併發已改狀態時 0 列 → failed_precondition,前端重查。
	upd := db.SalesOrder.UpdateOneID(in.OrderID).
		Where(salesorder.StatusEQ(from)).
		SetStatus(to).
		SetVersion(order.Version + 1)
	// 派車欄位:dispatch 寫入,dispatch_cancel 清 dispatched 但保留 route 看板位置(D13)。
	switch to {
	case OrderStatusProcessing:
		if in.DispatchedBy != nil {
			upd = upd.SetDispatchedBy(*in.DispatchedBy).SetDispatchedAt(time.Now().UTC())
		}
		if in.RouteID != nil {
			upd = upd.SetRouteID(*in.RouteID)
		}
		if in.DeliverySequence != nil {
			upd = upd.SetDeliverySequence(*in.DeliverySequence)
		}
	case OrderStatusPending:
		upd = upd.ClearDispatchedAt().ClearDispatchedBy()
	}
	if in.MarkDelivered {
		// 10.9:完成由配送回寫促成 → 蓋送達時點(呼叫端帶入以維持同時點,未帶則取當下)。
		at := time.Now().UTC()
		if in.DeliveredAt != nil {
			at = in.DeliveredAt.UTC()
		}
		upd = upd.SetDeliveredAt(at)
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.SysInvalidArgument.Error(map[string]string{
				"field": "status", "from": from, "to": to,
			})
		}
		return errcode.SysInternal.Wrap(err)
	}
	eventType := orderEventFor(from, to)
	payload := map[string]any{"from": from, "to": to}
	if in.MarkDelivered {
		// 10.9:標記促成來源,區別「司機完成配送」與「店家手動結案」。
		payload["source"] = "delivery"
	}
	create := db.SalesOrderEvent.Create().
		SetSalesOrderID(updated.ID).
		SetCompanyID(updated.CompanyID).
		SetEventType(eventType).
		SetActorID(in.ActorID)
	if strings.TrimSpace(in.Reason) != "" {
		create = create.SetReason(strings.TrimSpace(in.Reason))
	}
	create = create.SetPayload(payload)
	if _, err := create.Save(ctx); err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	// dispatch_cancel/void 除事件外同交易寫稽核(D13/D18);10.9 送達回寫同理
	// (稽核的 before/after 需看得出「因配送完成而結案」)。
	if eventType == OrderEventDispatchCancel || eventType == OrderEventVoid || in.MarkDelivered {
		tx, ok := dbtenant.TxFrom(ctx)
		if !ok {
			return errcode.SysInternal.Wrap(errors.New("TransitionOrder 需在呼叫端的交易內執行"))
		}
		var did *int
		if updated.DepartmentID != nil {
			d := *updated.DepartmentID
			did = &d
		}
		after := map[string]any{"status": to, "reason": strings.TrimSpace(in.Reason)}
		auditAction := eventType
		if in.MarkDelivered {
			// action 枚舉無 "complete"(§5.2 列舉:create/update/delete/login/logout/print/
			// force_logout/role_change/dispatch_cancel/void)→ 送達回寫以 update 記錄,
			// 事實由 before/after 承載(status 前後值 + source=delivery)。
			auditAction = "update"
			after["source"] = "delivery"
			if updated.DeliveredAt != nil {
				after["delivered_at"] = updated.DeliveredAt.UTC().Format(time.RFC3339)
			}
		}
		if err := recordAuditBA(ctx, tx, "sales_order", auditAction, updated.ID,
			updated.CompanyID, did, in.ActorID,
			map[string]any{"status": from}, after); err != nil {
			return errcode.SysInternal.Wrap(err)
		}
	}
	return nil
}

// orderEventFor 由狀態轉移推導事件型別。
func orderEventFor(from, to string) string {
	switch {
	case from == OrderStatusPending && to == OrderStatusProcessing:
		return OrderEventDispatch
	case from == OrderStatusProcessing && to == OrderStatusPending:
		return OrderEventDispatchCancel
	case from == OrderStatusPending && to == OrderStatusCancelled:
		return OrderEventCancel
	case from == OrderStatusProcessing && to == OrderStatusCompleted:
		return OrderEventComplete
	case from == OrderStatusCompleted && to == OrderStatusVoided:
		return OrderEventVoid
	default:
		return "edit"
	}
}
