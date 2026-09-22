//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// newOrderServer 以請求交易掛 SalesOrderService。
func newOrderServer(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.SalesOrderServiceClient {
	t.Helper()
	path, handler := salesorderv1connect.NewSalesOrderServiceHandler(NewSalesOrderService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		scope := auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true}
		if id.Role == "customer" {
			scope.DataScope = auth.DataScopeSelf
			scope.CustomerID = id.CustomerID
		}
		ctx = auth.WithRLS(ctx, scope)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewSalesOrderServiceClient(http.DefaultClient, ts.URL)
}

// newCustomerProductServer 以請求交易掛 CustomerProductService。
func newCustomerProductServer(t *testing.T, db *ent.Client, id authz.Identity) productsv1connect.CustomerProductServiceClient {
	t.Helper()
	path, handler := productsv1connect.NewCustomerProductServiceHandler(NewCustomerProductService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		scope := auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true}
		ctx = auth.WithRLS(ctx, scope)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return productsv1connect.NewCustomerProductServiceClient(http.DefaultClient, ts.URL)
}

// countPending 計某使用者的 pending 通知(系統範圍讀)。
func countPending(t *testing.T, ctx context.Context, db *ent.Client, uid int) int {
	t.Helper()
	var n int
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		c, err := tx.Client().Notification.Query().
			Where(notification.UserIDEQ(uid), notification.StatusEQ("pending")).Count(c2)
		if err != nil {
			t.Fatalf("計通知: %v", err)
		}
		n = c
		return nil
	})
	return n
}

// TestIntegrationOrderCreatedTrigger 業務下單推子帳號(2 通道 x N 子帳號);客戶自下不推。
func TestIntegrationOrderCreatedTrigger(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, custID, actorID, prodID := seedOrderCompany(t, ctx, db)
	// 子帳號 + 主帳號另建(seedOrderCompany 只有操作員)。
	var subID, priID int
	seedTx(t, db, func(tx *ent.Tx) error {
		pri, err := tx.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetEmail("pri-" + t.Name() + "@t.com").SetName("主").
			SetRole("customer").SetPasswordHash("x").
			SetIsCustomer(true).SetCustomerID(custID).SetIsPrimary(true).Save(ctx)
		if err != nil {
			return err
		}
		sub, err := tx.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetEmail("sub-" + t.Name() + "@t.com").SetName("子").
			SetRole("customer").SetPasswordHash("x").
			SetIsCustomer(true).SetCustomerID(custID).SetIsPrimary(false).Save(ctx)
		if err != nil {
			return err
		}
		priID, subID = pri.ID, sub.ID
		return nil
	})
	subIDv := returnIDs{co: coID, dept: deptID, cust: custID, primary: priID, sub: subID, prod: prodID}
	_ = actorID
	SetTriggerSender(&FakeSender{})
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	// 業務身分下單(dept_admin 經 SalesOrderService 全鏈)。
	adminID := authz.Identity{UserID: itoa(actorID), CompanyID: itoa(coID),
		DepartmentID: itoa(deptID), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	rpcOrder := newOrderServer(t, db, adminID)
	cr, err := rpcOrder.CreateOrder(ctx, connect.NewRequest(&salesorderv1.CreateOrderRequest{
		CustomerId: itoa(custID), Source: "W",
		ExpectedDeliveryDate: "2026-09-22",
		Items: []*salesorderv1.OrderItemInput{{
			ProductId: itoa(prodID), DisplayName: "蘋果", Qty: "2", Unit: "斤",
		}},
	}))
	if err != nil {
		t.Fatalf("業務下單: %v", err)
	}
	_ = cr
	// 子帳號應收到 in_app + fcm 各一筆。
	total := countAllFor(t, ctx, db, subID)
	if total != 2 {
		t.Fatalf("子帳號應收到 2 筆(in_app+fcm),got %d", total)
	}
	// 主帳號無通知。
	if n := countAllFor(t, ctx, db, priID); n != 0 {
		t.Fatalf("主帳號應無通知,got %d", n)
	}
	// 客戶自下：直呼 OnOrderCreated(isCustomerOrder=true) → 不建通知(路由單元斷言;
	// CreateOrder 經 deptScope 拒客戶身分,客戶自下走 App 另一通道,故不經 RPC)。
	before := countAllFor(t, ctx, db, subID)
	_ = subIDv
	var didPtr *int
	_ = didPtr
	if err := dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		d := deptID
		return OnOrderCreated(c2, tx.Client(), coID, &d, custID, true, 999, "W999999", 1)
	}); err != nil {
		t.Fatalf("客戶自下觸發: %v", err)
	}
	if after := countAllFor(t, ctx, db, subID); after != before {
		t.Fatalf("客戶自下不得新增通知,before=%d after=%d", before, after)
	}
}

// countAllFor 計某使用者全部通知。
func countAllFor(t *testing.T, ctx context.Context, db *ent.Client, uid int) int {
	t.Helper()
	var n int
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		c, err := tx.Client().Notification.Query().
			Where(notification.UserIDEQ(uid)).Count(c2)
		if err != nil {
			t.Fatalf("計通知: %v", err)
		}
		n = c
		return nil
	})
	return n
}

// TestIntegrationProductAndReviewTriggers 專屬推主責 + 退貨審核推發起。
func TestIntegrationProductAndReviewTriggers(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedReviewShop(t, ctx, db)
	SetTriggerSender(&FakeSender{})
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	// 後台新增專屬商品 → 主責業務(rep)收到 2 筆。
	rpcProd := newCustomerProductServer(t, db, staffIdentity(v.returnIDs, v.admin, "dept_admin"))
	// 需第二商品(第一已在 seed 佔用)。
	var prod2 int
	seedTx(t, db, func(tx *ent.Tx) error {
		p, err := tx.Product.Create().SetCompanyID(v.co).SetDepartmentID(v.dept).
			SetCode("P2").SetName("梨").Save(ctx)
		if err != nil {
			return err
		}
		prod2 = p.ID
		return nil
	})
	if _, err := rpcProd.AddCustomerProduct(ctx, connect.NewRequest(&productsv1.AddCustomerProductRequest{
		CustomerId: itoa(v.cust), ProductId: itoa(prod2),
	})); err != nil {
		t.Fatalf("新增專屬: %v", err)
	}
	if n := countAllFor(t, ctx, db, v.rep); n != 2 {
		t.Fatalf("主責應收到 2 筆,got %d", n)
	}
	// 退貨審核 → 發起子帳號收到 2 筆。
	rpcSub := newReturnServer(t, db, subIdentity(v.returnIDs, v.sub))
	rpcRep := newReturnServer(t, db, staffIdentity(v.returnIDs, v.rep, "staff"))
	before := countAllFor(t, ctx, db, v.sub)
	rid := createPending(t, ctx, rpcSub, v.returnIDs)
	if _, err := rpcRep.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
		Id: rid, Decision: "approved", ExpectedVersion: "0",
	})); err != nil {
		t.Fatalf("審核: %v", err)
	}
	if after := countAllFor(t, ctx, db, v.sub); after-before != 2 {
		t.Fatalf("發起帳號應新增 2 筆,before=%d after=%d", before, after)
	}
}
