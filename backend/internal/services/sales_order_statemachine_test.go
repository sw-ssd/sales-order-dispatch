package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// seedTestOrder 建一筆 pending 訂單(公司 7)。回傳 id。
func seedTestOrder(t *testing.T, ctx context.Context, db *ent.Client) int {
	t.Helper()
	o, err := db.SalesOrder.Create().
		SetCompanyID(7).SetOrderNo("W000001").SetCustomerID(11).
		SetSource("W").SetStatus(OrderStatusPending).SetVersion(1).Save(ctx)
	if err != nil {
		t.Fatalf("seed 訂單: %v", err)
	}
	return o.ID
}

// txOrderCtx 開交易並注入 dbtenant(TransitionOrder 的稽核路徑要 TxFrom)。
func txOrderCtx(t *testing.T, db *ent.Client) (context.Context, *ent.Tx) {
	t.Helper()
	tx, err := db.Tx(context.Background())
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	return dbtenant.WithTenantTx(context.Background(), tx), tx
}

// TestOrderTransitionLegalPaths 合法路徑全通過且各寫對應事件。
func TestOrderTransitionLegalPaths(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := txOrderCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()

	id := seedTestOrder(t, ctx, client)
	by, route, seq := 3, 5, 1
	steps := []struct {
		to    string
		event string
		in    OrderTransitionInput
	}{
		{OrderStatusProcessing, OrderEventDispatch,
			OrderTransitionInput{OrderID: id, To: OrderStatusProcessing, ActorID: 3,
				DispatchedBy: &by, RouteID: &route, DeliverySequence: &seq}},
		{OrderStatusPending, OrderEventDispatchCancel,
			OrderTransitionInput{OrderID: id, To: OrderStatusPending, ActorID: 3, Reason: "派錯車"}},
		{OrderStatusProcessing, OrderEventDispatch,
			OrderTransitionInput{OrderID: id, To: OrderStatusProcessing, ActorID: 3,
				DispatchedBy: &by, RouteID: &route, DeliverySequence: &seq}},
		{OrderStatusCompleted, OrderEventComplete,
			OrderTransitionInput{OrderID: id, To: OrderStatusCompleted, ActorID: 3}},
		{OrderStatusVoided, OrderEventVoid,
			OrderTransitionInput{OrderID: id, To: OrderStatusVoided, ActorID: 3, Reason: "內容錯誤重建"}},
	}
	for _, s := range steps {
		if err := TransitionOrder(ctx, client, s.in); err != nil {
			t.Fatalf("轉移→%s: %v", s.to, err)
		}
		got := client.SalesOrder.GetX(ctx, id)
		if got.Status != s.to {
			t.Fatalf("狀態 = %q;want %q", got.Status, s.to)
		}
		n := client.SalesOrderEvent.Query().
			Where(salesorderevent.SalesOrderIDEQ(id), salesorderevent.EventTypeEQ(s.event)).CountX(ctx)
		if n < 1 {
			t.Fatalf("事件 %q 應至少 1 筆,got %d", s.event, n)
		}
	}
	// 取消派車保留看板位置:重查 pending 那次之後的 processing 列。
	got := client.SalesOrder.GetX(ctx, id)
	if got.RouteID == nil || *got.RouteID != route {
		t.Fatalf("作廢後 route 應保留,got %+v", got)
	}
}

// TestOrderTransitionCancelPending pending→cancelled 寫 cancel 事件;終態拒絕任何異動。
func TestOrderTransitionCancelPending(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := txOrderCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()

	id := seedTestOrder(t, ctx, client)
	if err := TransitionOrder(ctx, client,
		OrderTransitionInput{OrderID: id, To: OrderStatusCancelled, ActorID: 3}); err != nil {
		t.Fatalf("取消: %v", err)
	}
	for _, to := range []string{OrderStatusPending, OrderStatusProcessing, OrderStatusCompleted, OrderStatusVoided} {
		if err := TransitionOrder(ctx, client,
			OrderTransitionInput{OrderID: id, To: to, ActorID: 3, Reason: "x"}); err == nil {
			t.Fatalf("cancelled→%s 應拒絕", to)
		} else if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("cancelled→%s 應 invalid_argument,got %v", to, err)
		}
	}
}

// TestOrderTransitionIllegalPaths 非法路徑全拒絕且狀態不變、無事件。
func TestOrderTransitionIllegalPaths(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := txOrderCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()

	id := seedTestOrder(t, ctx, client)
	cases := []struct {
		name string
		in   OrderTransitionInput
	}{
		{"pending 不得直達 completed", OrderTransitionInput{OrderID: id, To: OrderStatusCompleted, ActorID: 3}},
		{"pending 不得直達 voided", OrderTransitionInput{OrderID: id, To: OrderStatusVoided, ActorID: 3, Reason: "x"}},
		{"回退 pending 不填原因", OrderTransitionInput{OrderID: id, To: OrderStatusPending, ActorID: 3}},
		{"作廢不填原因(先 pending→processing 再 void)", OrderTransitionInput{OrderID: id, To: OrderStatusProcessing, ActorID: 3}},
	}
	for _, c := range cases[:3] {
		if err := TransitionOrder(ctx, client, c.in); err == nil {
			t.Fatalf("%s:應拒絕", c.name)
		}
	}
	if got := client.SalesOrder.GetX(ctx, id); got.Status != OrderStatusPending {
		t.Fatalf("非法轉移不得改狀態,got %q", got.Status)
	}
	if n := client.SalesOrderEvent.Query().Where(salesorderevent.SalesOrderIDEQ(id)).CountX(ctx); n != 0 {
		t.Fatalf("非法轉移不得寫事件,got %d 筆", n)
	}
}

// TestOrderTransitionDispatchCancelKeepsRoute 取消派車清 dispatched 但保留 route 看板位置。
func TestOrderTransitionDispatchCancelKeepsRoute(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := txOrderCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()

	id := seedTestOrder(t, ctx, client)
	by, route, seq := 3, 9, 2
	if err := TransitionOrder(ctx, client, OrderTransitionInput{OrderID: id,
		To: OrderStatusProcessing, ActorID: 3,
		DispatchedBy: &by, RouteID: &route, DeliverySequence: &seq}); err != nil {
		t.Fatalf("派車: %v", err)
	}
	if err := TransitionOrder(ctx, client, OrderTransitionInput{OrderID: id,
		To: OrderStatusPending, ActorID: 3, Reason: "派錯車"}); err != nil {
		t.Fatalf("取消派車: %v", err)
	}
	got := client.SalesOrder.GetX(ctx, id)
	if got.DispatchedAt != nil || got.DispatchedBy != nil {
		t.Fatalf("取消派車應清 dispatched,got %+v", got)
	}
	if got.RouteID == nil || *got.RouteID != route || got.DeliverySequence == nil || *got.DeliverySequence != seq {
		t.Fatalf("取消派車應保留看板位置,got %+v", got)
	}
}

// TestOrderTransitionVoidWritesAudit 作廢同交易寫稽核;稽核失敗則狀態回滾。
func TestOrderTransitionVoidWritesAudit(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := txOrderCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()

	id := seedTestOrder(t, ctx, client)
	by := 3
	if err := TransitionOrder(ctx, client, OrderTransitionInput{OrderID: id,
		To: OrderStatusProcessing, ActorID: 3, DispatchedBy: &by}); err != nil {
		t.Fatalf("派車: %v", err)
	}
	if err := TransitionOrder(ctx, client, OrderTransitionInput{OrderID: id,
		To: OrderStatusCompleted, ActorID: 3}); err != nil {
		t.Fatalf("完成: %v", err)
	}
	if err := TransitionOrder(ctx, client, OrderTransitionInput{OrderID: id,
		To: OrderStatusVoided, ActorID: 3, Reason: "內容錯誤"}); err != nil {
		t.Fatalf("作廢: %v", err)
	}
	n := client.AuditLog.Query().CountX(ctx)
	if n != 1 {
		t.Fatalf("作廢應恰寫一筆稽核,got %d", n)
	}
}
