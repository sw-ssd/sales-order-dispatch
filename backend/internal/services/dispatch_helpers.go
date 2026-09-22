// Package services 的派車輔助(08 計畫 Task 5.1):日期/順位解析、範圍比對、順位重排、
// 事件寫入、批次候選、車次範圍檢查。
package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
)

// parseBoardDate 解析看板日期(YYYY-MM-DD)。
func parseBoardDate(s string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, invalidArgField("expected_delivery_date")
	}
	return d, nil
}

// parseSeq 解析配送順位(≥1)。
func parseSeq(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 {
		return 0, invalidArgField("delivery_sequence")
	}
	return n, nil
}

// equalDept 比對部門歸屬(訂單部門可空視同不限制?不:派車要求同部門,空部門訂單僅空部門車次可指)。
func equalDept(orderDept *int, scopeDept *int) bool {
	if scopeDept == nil {
		return true
	}
	if orderDept == nil {
		return false
	}
	return *orderDept == *scopeDept
}

// shiftSequence 同車次順位重排:同日同部門 pending 訂單中 seq ≥ 新順位者(除本筆)依序後移。
// 原車次空位不回填(順位允許空洞)。
func shiftSequence(ctx context.Context, db *ent.Client, cid int, did *int, routeID int, day time.Time, seq, excludeID int) error {
	dayStart := day.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)
	q := db.SalesOrder.Query().
		Where(salesorder.CompanyIDEQ(cid), salesorder.RouteIDEQ(routeID),
			salesorder.StatusEQ(OrderStatusPending), salesorder.DeletedAtIsNil(),
			salesorder.ExpectedDeliveryDateGTE(dayStart), salesorder.ExpectedDeliveryDateLT(dayEnd),
			salesorder.DeliverySequenceGTE(seq), salesorder.IDNEQ(excludeID))
	if did != nil {
		q = q.Where(salesorder.DepartmentIDEQ(*did))
	}
	rows, err := q.Order(ent.Asc(salesorder.FieldDeliverySequence)).All(ctx)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.DeliverySequence == nil || *r.DeliverySequence < seq {
			continue
		}
		if _, err := db.SalesOrder.UpdateOneID(r.ID).
			SetDeliverySequence(*r.DeliverySequence + 1).SetVersion(r.Version + 1).Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// writeAssignEvent 寫 route_assign 事件(含新舊 route/seq)。
func writeAssignEvent(ctx context.Context, db *ent.Client, actor int, o *ent.SalesOrder, routeID *int, seq *int) error {
	b := db.SalesOrderEvent.Create().SetSalesOrderID(o.ID).SetCompanyID(o.CompanyID).
		SetEventType("route_assign").SetActorID(actor).
		SetPayload(map[string]any{
			"from_route": o.RouteID, "to_route": routeID,
			"from_seq": o.DeliverySequence, "to_seq": seq,
		})
	if _, err := b.Save(ctx); err != nil {
		return err
	}
	return nil
}

// scopedRoute 讀範圍內車次(跨範圍視同 not_found)。
func scopedRoute(ctx context.Context, db *ent.Client, cid int, did *int, rid int) (*ent.Route, error) {
	r, err := db.Route.Query().Where(route.ID(rid)).Only(ctx)
	if err != nil {
		return nil, err
	}
	if r.CompanyID != cid || !equalRouteDept(r.DepartmentID, did) {
		return nil, notFoundSentinel()
	}
	return r, nil
}

// equalRouteDept 比對車次部門(車次部門可空 = 公司共用,任何部門可派)。
func equalRouteDept(routeDept *int, scopeDept *int) bool {
	if routeDept == nil {
		return true
	}
	if scopeDept == nil {
		return true
	}
	return *routeDept == *scopeDept
}

// confirmCand 為批次候選(快照版本供條件更新)。
type confirmCand struct {
	ID       int
	RouteIDV int
	SeqV     *int
	VersionV int
	DeptV    *int
}

// confirmCandidates 撈批次候選:同車次同日同部門 pending 且 route 不為空。
func confirmCandidates(ctx context.Context, db *ent.Client, cid int, did *int, rid int, day time.Time) ([]confirmCand, error) {
	dayStart := day.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)
	q := db.SalesOrder.Query().
		Where(salesorder.CompanyIDEQ(cid), salesorder.RouteIDEQ(rid),
			salesorder.StatusEQ(OrderStatusPending), salesorder.DeletedAtIsNil(),
			salesorder.ExpectedDeliveryDateGTE(dayStart), salesorder.ExpectedDeliveryDateLT(dayEnd))
	if did != nil {
		q = q.Where(salesorder.DepartmentIDEQ(*did))
	}
	rows, err := q.Order(ent.Asc(salesorder.FieldDeliverySequence)).All(ctx)
	if err != nil {
		return nil, err
	}
	var out []confirmCand
	for _, r := range rows {
		if r.RouteID == nil {
			continue
		}
		out = append(out, confirmCand{ID: r.ID, RouteIDV: *r.RouteID,
			SeqV: r.DeliverySequence, VersionV: r.Version, DeptV: r.DepartmentID})
	}
	return out, nil
}

// failReasonOf 萃取逐筆失敗原因(對外可讀,無內部細節)。
func failReasonOf(err error) string {
	if err == nil {
		return ""
	}
	return "狀態已變更或不可派車"
}

// notFoundSentinel 回傳哨兵 not_found(由 toConnectError 映射)。
func notFoundSentinel() error {
	return &ent.NotFoundError{}
}

var _ = salesorderevent.FieldID
