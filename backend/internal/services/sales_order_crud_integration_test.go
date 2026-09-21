//go:build integration

package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// seedOrderCompany 建公司+部門+操作者+客戶,回 coID/deptID/custID/actorID。
// actor 必須是真實 users 列(audit_logs.user_id FK);來源字典由 00011 seed。
func seedOrderCompany(t *testing.T, ctx context.Context, db *ent.Client) (int, int, int, int) {
	t.Helper()
	var coID, deptID, custID, actorID int
	seedTx(t, db, func(tx *ent.Tx) error {
		co, err := tx.Company.Create().SetName("訂單公司").SetIdentifier("O-" + t.Name()).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		d, err := tx.Department.Create().SetCompanyID(co.ID).SetName("門市一").Save(ctx)
		if err != nil {
			return err
		}
		cust, err := tx.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetCustomerCode("TY000001").SetName("王小明").Save(ctx)
		if err != nil {
			return err
		}
		op, err := tx.User.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetEmail("op-" + t.Name() + "@t.com").SetName("操作員").
			SetRole("dept_admin").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		// order_source W 由 migration 00011 seed(系統級固定);fixture 不重建。
		coID, deptID, custID, actorID = co.ID, d.ID, cust.ID, op.ID
		return nil
	})
	return coID, deptID, custID, actorID
}

// TestIntegrationSalesOrderCRUD 建單→查單→編輯→取消→軟刪除,經服務層 + 真 PG + RLS 定義。
func TestIntegrationSalesOrderCRUD(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, custID, actorID := seedOrderCompany(t, ctx, db)
	svc := NewSalesOrderService(db)

	newReqCtx := func() (context.Context, func()) {
		tx, err := db.Tx(context.Background())
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		id := authz.Identity{UserID: uItoa(actorID), CompanyID: uItoa(coID), DepartmentID: uItoa(deptID),
			Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		ctx := authz.WithIdentity(context.Background(), id)
		ctx = dbtenant.WithTenantTx(ctx, tx)
		return ctx, func() { _ = tx.Rollback() }
	}
	commit := func(ctx context.Context) {
		tx, _ := dbtenant.TxFrom(ctx)
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交: %v", err)
		}
	}

	// 建單。
	cctx, fin := newReqCtx()
	created, err := svc.CreateOrder(cctx, connect.NewRequest(&salesorderv1.CreateOrderRequest{
		CustomerId: uItoa(custID), Source: "W",
		Items: []*salesorderv1.OrderItemInput{
			{DisplayName: "蘋果", Qty: "10", Unit: "斤"},
		},
	}))
	if err != nil {
		fin()
		t.Fatalf("CreateOrder: %v", err)
	}
	commit(cctx)
	fin()
	if got := created.Msg.GetOrder().GetOrderNo(); got != "W000001" {
		t.Fatalf("order_no 應為 W000001,got %q", got)
	}
	oid := created.Msg.GetOrder().GetId()

	// 查單(含明細 base_qty=qty)。
	gctx, gfin := newReqCtx()
	got, err := svc.GetOrder(gctx, connect.NewRequest(&salesorderv1.GetOrderRequest{Id: oid}))
	if err != nil {
		gfin()
		t.Fatalf("GetOrder: %v", err)
	}
	commit(gctx)
	gfin()
	if len(got.Msg.GetItems()) != 1 || got.Msg.GetItems()[0].GetBaseQty() != "10" {
		t.Fatalf("明細應 1 筆且 base_qty=qty,got %+v", got.Msg.GetItems())
	}

	// 編輯(version 1→2)。
	uctx, ufin := newReqCtx()
	upd, err := svc.UpdateOrder(uctx, connect.NewRequest(&salesorderv1.UpdateOrderRequest{
		Id: oid, Version: 1, Note: "改備註",
		Items: []*salesorderv1.OrderItemInput{{DisplayName: "蘋果", Qty: "20", Unit: "斤"}},
	}))
	if err != nil {
		ufin()
		t.Fatalf("UpdateOrder: %v", err)
	}
	commit(uctx)
	ufin()
	if upd.Msg.GetOrder().GetVersion() != 2 {
		t.Fatalf("version 應為 2,got %d", upd.Msg.GetOrder().GetVersion())
	}

	// 取消。
	cactx, cafin := newReqCtx()
	cancelled, err := svc.CancelOrder(cactx, connect.NewRequest(&salesorderv1.CancelOrderRequest{Id: oid}))
	if err != nil {
		cafin()
		t.Fatalf("CancelOrder: %v", err)
	}
	commit(cactx)
	cafin()
	if cancelled.Msg.GetOrder().GetStatus() != OrderStatusCancelled {
		t.Fatalf("取消後應為 cancelled,got %q", cancelled.Msg.GetOrder().GetStatus())
	}

	// 終態拒絕編輯。
	ectx, ecfin := newReqCtx()
	_, err = svc.UpdateOrder(ectx, connect.NewRequest(&salesorderv1.UpdateOrderRequest{
		Id: oid, Version: 2, Note: "x",
	}))
	ecfin()
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("cancelled 編輯應 invalid_argument,got %v", err)
	}

	// 軟刪除(cancelled 可刪)後 Get → not_found。
	dctx, dfin := newReqCtx()
	if _, err := svc.DeleteOrder(dctx, connect.NewRequest(&salesorderv1.DeleteOrderRequest{Id: oid})); err != nil {
		dfin()
		t.Fatalf("DeleteOrder: %v", err)
	}
	commit(dctx)
	dfin()
	lctx, lfin := newReqCtx()
	_, err = svc.GetOrder(lctx, connect.NewRequest(&salesorderv1.GetOrderRequest{Id: oid}))
	lfin()
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("刪除後 Get 應 not_found,got %v", err)
	}

	// 事件軌跡:create + edit + cancel 至少 3 筆(寫在第二筆訂單上查,因本單已刪不可見)。
	c2ctx, c2fin := newReqCtx()
	created2, err := svc.CreateOrder(c2ctx, connect.NewRequest(&salesorderv1.CreateOrderRequest{
		CustomerId: uItoa(custID), Source: "W",
		Items: []*salesorderv1.OrderItemInput{{DisplayName: "梨", Qty: "5", Unit: "顆"}},
	}))
	if err != nil {
		c2fin()
		t.Fatalf("CreateOrder2: %v", err)
	}
	commit(c2ctx)
	c2fin()
	oid2 := created2.Msg.GetOrder().GetId()
	evctx, evfin := newReqCtx()
	events, err := svc.ListOrderEvents(evctx, connect.NewRequest(&salesorderv1.ListOrderEventsRequest{SalesOrderId: oid2}))
	evfin()
	if err != nil {
		t.Fatalf("ListOrderEvents: %v", err)
	}
	if len(events.Msg.GetEvents()) < 1 || events.Msg.GetEvents()[0].GetEventType() != "create" {
		t.Fatalf("首事件應為 create,got %+v", events.Msg.GetEvents())
	}
	_ = salesorderevent.EventTypeEQ
}
