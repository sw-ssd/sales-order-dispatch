package services

import (
	"context"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// mkOrder 在 fixture 公司/部門下建一筆訂單(status/route 由參數決定;客戶沿用預設佔位)。
func mkOrder(t *testing.T, f *logisticsFixture, status string, routeID *int) int {
	return mkOrderFor(t, f, status, routeID, 11)
}

// mkOrderFor 同 mkOrder,但指定客戶(10.11 送達通知需要真 customer 列才能解析收件者)。
func mkOrderFor(t *testing.T, f *logisticsFixture, status string, routeID *int, custID int) int {
	t.Helper()
	build := f.db.SalesOrder.Create().
		SetCompanyID(f.coID).SetDepartmentID(f.deptID).
		SetOrderNo("W-" + status).SetCustomerID(custID).
		SetSource("W").SetStatus(status).SetVersion(1)
	if routeID != nil {
		build = build.SetRouteID(*routeID)
	}
	return build.SaveX(context.Background()).ID
}

// orderByID 讀回訂單(測試斷言用)。
func orderByID(t *testing.T, f *logisticsFixture, id int) *ent.SalesOrder {
	t.Helper()
	o, err := f.db.SalesOrder.Query().Where(salesorder.ID(id)).Only(context.Background())
	if err != nil {
		t.Fatalf("讀訂單 %d: %v", id, err)
	}
	return o
}

// mkCustomerWithSub 建一客戶 + 一子帳號（10.11 送達通知的收件者）。
func mkCustomerWithSub(t *testing.T, f *logisticsFixture, code string) (custID, subID int) {
	t.Helper()
	ctx := context.Background()
	c := f.db.Customer.Create().SetCompanyID(f.coID).SetDepartmentID(f.deptID).
		SetCustomerCode(code).SetName("客戶" + code).SaveX(ctx)
	u := f.db.User.Create().SetEmail("sub-" + code + "@t.com").SetName("子").
		SetRole("customer").SetPasswordHash("x").SetCompanyID(f.coID).
		SetDepartmentID(f.deptID).SetIsCustomer(true).SetCustomerID(c.ID).
		SetIsPrimary(false).SetStatus("active").SaveX(ctx)
	return c.ID, u.ID
}

// startAndComplete 走一次完整的「開始 → 完成」(交付單 id 固定為 1,由 setupAssignedDelivery 建立)。
func startAndComplete(t *testing.T, rpc salesorderv1connect.LogisticsServiceClient) error {
	t.Helper()
	ctx := context.Background()
	if _, err := rpc.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: "1", Version: "1",
	})); err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}
	_, err := rpc.CompleteDelivery(ctx, connect.NewRequest(&v1.CompleteDeliveryRequest{
		DeliveryId: "1", Version: "2",
	}))
	return err
}

// TestDeliveryCompleteWritesBackOrders 10.9 核心驗收:車次完成 → 該車次所載每筆
// processing 訂單轉 completed、蓋 delivered_at、寫 complete 事件(標記來源為配送);
// 他車次與 pending 訂單不受影響。
func TestDeliveryCompleteWritesBackOrders(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, delID := setupAssignedDelivery(t)
	_ = delID
	rpc := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	custID, _ := mkCustomerWithSub(t, f, "WB1")
	o1 := mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custID)
	o2 := mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custID)
	otherRoute := f.db.Route.Create().SetCode("R9").SetName("別線").
		SetCompanyID(f.coID).SetDepartmentID(f.deptID).SaveX(ctx).ID
	oOther := mkOrderFor(t, f, OrderStatusProcessing, &otherRoute, custID)
	oPending := mkOrderFor(t, f, OrderStatusPending, &f.routeID, custID)

	if err := startAndComplete(t, rpc); err != nil {
		t.Fatalf("完成配送: %v", err)
	}

	for _, id := range []int{o1, o2} {
		o := orderByID(t, f, id)
		if o.Status != OrderStatusCompleted || o.DeliveredAt == nil {
			t.Fatalf("訂單 %d 應 completed 且有 delivered_at,got status=%s delivered_at=%v",
				id, o.Status, o.DeliveredAt)
		}
	}
	// 事件:complete 且 payload 標 source=delivery(與店家手動結案可區別)。
	events := f.db.SalesOrderEvent.Query().AllX(ctx)
	byOrder := map[int]*ent.SalesOrderEvent{}
	for _, e := range events {
		if e.EventType == OrderEventComplete {
			byOrder[e.SalesOrderID] = e
		}
	}
	for _, id := range []int{o1, o2} {
		e, ok := byOrder[id]
		if !ok {
			t.Fatalf("訂單 %d 應有 complete 事件", id)
		}
		if src, _ := e.Payload["source"].(string); src != "delivery" {
			t.Fatalf("訂單 %d 事件應標 source=delivery,got %+v", id, e.Payload)
		}
	}
	// 稽核:每筆回寫恰一列。
	audits := f.db.AuditLog.Query().AllX(ctx)
	for _, id := range []int{o1, o2} {
		n := 0
		for _, a := range audits {
			if a.ResourceType == "sales_order" && a.ResourceID == strconv.Itoa(id) &&
				a.Action == "update" && a.AfterSnapshot["source"] == "delivery" {
				n++
			}
		}
		if n != 1 {
			t.Fatalf("訂單 %d 應有 1 筆回寫稽核,got %d", id, n)
		}
	}
	// Negative control:他車次與 pending 不動。
	if o := orderByID(t, f, oOther); o.Status != OrderStatusProcessing {
		t.Fatalf("他車次訂單不應被回寫,got %s", o.Status)
	}
	if o := orderByID(t, f, oPending); o.Status != OrderStatusPending || o.DeliveredAt != nil {
		t.Fatalf("pending 訂單不應被回寫,got status=%s", o.Status)
	}
}

// TestDeliveryCompleteNoOrdersOnRoute 車次上沒有可回寫訂單 → 回寫 0 筆、配送照常完成。
func TestDeliveryCompleteNoOrdersOnRoute(t *testing.T) {
	f, driverUID, _, _ := setupAssignedDelivery(t)
	rpc := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))
	if err := startAndComplete(t, rpc); err != nil {
		t.Fatalf("無訂單車次完成應成功: %v", err)
	}
}

// TestWritebackRollsBackWholeTransaction 10.9 錯誤處理:回寫途中任一筆失敗 →
// 整交易回滾,不留「配送完成但訂單沒結」或「部分訂單已送達」的半套。
//
// 造法(不靠併發插隊):直接把「查詢得到的訂單切片」餵給 markOrdersDelivered,
// 其中一筆在寫入前已被改成終態 —— 狀態機的條件更新落空 → 回寫失敗 → 交易回滾。
// 這正是真實競態的等價物:查詢與更新之間狀態被別人改掉。
func TestWritebackRollsBackWholeTransaction(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, _ := setupAssignedDelivery(t)
	rpc := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	o1 := mkOrder(t, f, OrderStatusProcessing, &f.routeID)
	o2 := mkOrder(t, f, OrderStatusProcessing, &f.routeID)

	if _, err := rpc.StartDelivery(ctx, connect.NewRequest(&v1.StartDeliveryRequest{
		DeliveryId: "1", Version: "1",
	})); err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}
	// 查詢當下取到的切片(兩筆都是 processing),接著把第二筆改成終態模擬競態。
	orders, err := f.db.SalesOrder.Query().Where(salesorder.IDEQ(o2)).All(ctx)
	if err != nil {
		t.Fatalf("讀訂單: %v", err)
	}
	stale := []*ent.SalesOrder{orderByID(t, f, o1), orders[0]}
	if err := f.db.SalesOrder.UpdateOneID(o2).SetStatus(OrderStatusVoided).Exec(ctx); err != nil {
		t.Fatalf("前置改狀態: %v", err)
	}

	// 交易語意:在此交易內回寫 → 失敗 → 整交易回滾。
	tx, err := f.db.Tx(ctx)
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	tctx := dbtenant.WithTenantTx(context.Background(), tx)
	if err := markOrdersDelivered(tctx, tx.Client(), stale, 1); err == nil {
		t.Fatal("回寫遇非法狀態應失敗(不得靜默跳過)")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	// 回滾後:第一筆仍是 processing、未蓋送達時點。
	if o := orderByID(t, f, o1); o.Status != OrderStatusProcessing || o.DeliveredAt != nil {
		t.Fatalf("o1 應回滾為 processing,got status=%s delivered_at=%v", o.Status, o.DeliveredAt)
	}
}

// TestWritebackFailFastStopsAtFirstFailure 逐筆處理、遇錯即停(不繼續污染後續訂單)。
func TestWritebackFailFastStopsAtFirstFailure(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, _ := setupAssignedDelivery(t)
	_ = driverUID

	o1 := mkOrder(t, f, OrderStatusProcessing, &f.routeID)
	o2 := mkOrder(t, f, OrderStatusProcessing, &f.routeID)
	// 第一筆先改成終態,查詢仍取兩筆(模擬競態)後逐筆回寫。
	snap1, snap2 := orderByID(t, f, o1), orderByID(t, f, o2)
	if err := f.db.SalesOrder.UpdateOneID(o1).SetStatus(OrderStatusVoided).Exec(ctx); err != nil {
		t.Fatalf("前置: %v", err)
	}
	tx, err := f.db.Tx(ctx)
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	tctx := dbtenant.WithTenantTx(context.Background(), tx)
	if err := markOrdersDelivered(tctx, tx.Client(), []*ent.SalesOrder{snap1, snap2}, 1); err == nil {
		t.Fatal("應失敗")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	// 遇錯即停:第二筆未被處理(交易已回滾,狀態仍為 processing)。
	if o := orderByID(t, f, o2); o.Status != OrderStatusProcessing {
		t.Fatalf("第二筆不應被處理(遇錯即停),got %s", o.Status)
	}
}
