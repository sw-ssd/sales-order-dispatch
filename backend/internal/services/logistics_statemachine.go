// Package services 的配送執行狀態機(D32/10.6)。
// 合法轉移:`pending → in_progress → completed`、`pending/in_progress → cancelled`;
// 終態 completed/cancelled 無出口(10.6;重啟配送走新建/重指派,不回頭改已簽收的執行單)。
// 表驅動是為了讓「非法轉移」有一處可稽核,而不是散在各呼叫點的 if
// (比照 sales_order_statemachine.allowedOrderTransitions)。
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/logisticsdelivery"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// 配送執行狀態(10.6)。
const (
	DeliveryStatusPending    = "pending"
	DeliveryStatusInProgress = "in_progress"
	DeliveryStatusCompleted  = "completed"
	DeliveryStatusCancelled  = "cancelled"
)

// 配送事件型別(10.6),與 logistics_delivery_events.event_type 同值。
const (
	DeliveryEventCreated   = "created"
	DeliveryEventStarted   = "started"
	DeliveryEventCompleted = "completed"
	DeliveryEventCancelled = "cancelled"
)

// allowedDeliveryTransitions 為合法轉移表:不在表上的一律拒絕。
var allowedDeliveryTransitions = map[string][]string{
	DeliveryStatusPending:    {DeliveryStatusInProgress, DeliveryStatusCancelled},
	DeliveryStatusInProgress: {DeliveryStatusCompleted, DeliveryStatusCancelled},
	DeliveryStatusCompleted:  {},
	DeliveryStatusCancelled:  {},
}

// CanDeliveryTransition 回報 from → to 是否為合法轉移。
func CanDeliveryTransition(from, to string) bool {
	for _, next := range allowedDeliveryTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// DeliveryTransitionInput 為一次配送狀態轉移的參數。Reason 僅 cancelled 必填。
// ExpectedVersion > 0 時比對樂觀鎖(不符 → 併發改動,前端重查);0 = 不比對。
type DeliveryTransitionInput struct {
	DeliveryID      int
	To              string
	ActorID         int
	Reason          string
	ExpectedVersion int
}

// TransitionDelivery 執行一次配送狀態轉移:條件更新(WHERE status=前值 AND version=期望值)
// + 事件軌跡,同一交易。呼叫端須傳入請求交易的 ent client(db):交易邊界由呼叫端擁有
// (比照 TransitionOrder)。樂觀鎖不符(併發改動)→ failed_precondition 語意的 invalid_argument
// + reason,與 AssignDelivery 的版本衝突同語意(前端重查後重送)。
//
// 角色/本人檢查由呼叫端(授權層)執行,此處只守狀態機與原因必填。version 每次轉移 +1。
func TransitionDelivery(ctx context.Context, db *ent.Client, in DeliveryTransitionInput) (*ent.LogisticsDelivery, error) {
	to := strings.TrimSpace(in.To)
	if to == DeliveryStatusCancelled && strings.TrimSpace(in.Reason) == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "reason"})
	}
	cur, err := db.LogisticsDelivery.Query().
		Where(logisticsdelivery.ID(in.DeliveryID), logisticsdelivery.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	from := cur.Status
	if in.ExpectedVersion > 0 && in.ExpectedVersion != cur.Version {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "資料已變更，請重新載入"})
	}
	if !CanDeliveryTransition(from, to) {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{
			"field": "status", "from": from, "to": to,
		})
	}
	upd := db.LogisticsDelivery.UpdateOneID(in.DeliveryID).
		Where(logisticsdelivery.StatusEQ(from), logisticsdelivery.VersionEQ(cur.Version)).
		SetStatus(to).
		SetVersion(cur.Version + 1)
	switch to {
	case DeliveryStatusInProgress:
		upd = upd.SetStartedAt(time.Now().UTC())
	case DeliveryStatusCompleted:
		upd = upd.SetCompletedAt(time.Now().UTC())
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 條件更新 0 列 = 併發已改狀態(他人搶先)。
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "資料已變更，請重新載入"})
		}
		return nil, toConnectError(err)
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Wrap(errors.New("TransitionDelivery 需在呼叫端的交易內執行"))
	}
	ev := tx.Client().LogisticsDeliveryEvent.Create().
		SetLogisticsDeliveryID(updated.ID).
		SetCompanyID(updated.CompanyID).
		SetEventType(deliveryEventFor(from, to)).
		SetActorID(in.ActorID).
		SetPayload(map[string]any{"from": from, "to": to})
	if strings.TrimSpace(in.Reason) != "" {
		ev = ev.SetReason(strings.TrimSpace(in.Reason))
	}
	if _, err := ev.Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	return updated, nil
}

// deliveryEventFor 由狀態轉移推導事件型別。
func deliveryEventFor(from, to string) string {
	switch to {
	case DeliveryStatusInProgress:
		return DeliveryEventStarted
	case DeliveryStatusCompleted:
		return DeliveryEventCompleted
	case DeliveryStatusCancelled:
		return DeliveryEventCancelled
	default:
		return to
	}
}
