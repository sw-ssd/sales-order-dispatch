//go:build integration

package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationCustomerProductCRUD 清單 CRUD + for_order 語意 + Ensure 冪等。
func TestIntegrationCustomerProductCRUD(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, custID, actorID, prodID := seedOrderCompany(t, ctx, db)
	_ = actorID

	newCtx := func() (context.Context, func()) {
		tx, err := db.Tx(context.Background())
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		id := authz.Identity{UserID: uItoa(actorID), CompanyID: uItoa(coID), DepartmentID: uItoa(deptID),
			Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		c := authz.WithIdentity(context.Background(), id)
		c = dbtenant.WithTenantTx(c, tx)
		return c, func() { _ = tx.Rollback() }
	}
	commit := func(c context.Context) {
		tx, _ := dbtenant.TxFrom(c)
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交: %v", err)
		}
	}
	svc := NewCustomerProductService(db)

	// Add。
	c, fin := newCtx()
	added, err := svc.AddCustomerProduct(c, connect.NewRequest(&productsv1.AddCustomerProductRequest{
		CustomerId: uItoa(custID), ProductId: uItoa(prodID), AliasName: "大蘋果",
	}))
	if err != nil {
		fin()
		t.Fatalf("Add: %v", err)
	}
	commit(c)
	fin()
	if added.Msg.GetProduct().GetAliasName() != "大蘋果" {
		t.Fatalf("別名應為大蘋果,got %q", added.Msg.GetProduct().GetAliasName())
	}

	// 重複未刪 → already_exists。
	c, fin = newCtx()
	_, err = svc.AddCustomerProduct(c, connect.NewRequest(&productsv1.AddCustomerProductRequest{
		CustomerId: uItoa(custID), ProductId: uItoa(prodID),
	}))
	fin()
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("重複應 already_exists,got %v", err)
	}

	// for_order=false 見全量;改 qty=0 後 for_order=true 不見。
	c, fin = newCtx()
	if _, err := svc.UpdateCustomerProduct(c, connect.NewRequest(&productsv1.UpdateCustomerProductRequest{
		Id: added.Msg.GetProduct().GetId(), DefaultQty: "0",
	})); err != nil {
		fin()
		t.Fatalf("Update: %v", err)
	}
	commit(c)
	fin()
	c, fin = newCtx()
	all, err := svc.ListCustomerProducts(c, connect.NewRequest(&productsv1.ListCustomerProductsRequest{
		CustomerId: uItoa(custID),
	}))
	if err != nil {
		fin()
		t.Fatalf("List: %v", err)
	}
	fin()
	if len(all.Msg.GetProducts()) != 1 {
		t.Fatalf("管理視圖應見 qty=0,got %d", len(all.Msg.GetProducts()))
	}
	c, fin = newCtx()
	forOrder, err := svc.ListCustomerProducts(c, connect.NewRequest(&productsv1.ListCustomerProductsRequest{
		CustomerId: uItoa(custID), ForOrder: true,
	}))
	fin()
	if err != nil {
		t.Fatalf("List for_order: %v", err)
	}
	if len(forOrder.Msg.GetProducts()) != 0 {
		t.Fatalf("下單視圖不應見 qty=0,got %d", len(forOrder.Msg.GetProducts()))
	}

	// Ensure 冪等:存在回既有 created=false。
	c, fin = newCtx()
	ens, err := svc.EnsureCustomerProduct(c, connect.NewRequest(&productsv1.EnsureCustomerProductRequest{
		CustomerId: uItoa(custID), ProductId: uItoa(prodID), AliasName: "新別名",
	}))
	if err != nil {
		fin()
		t.Fatalf("Ensure: %v", err)
	}
	commit(c)
	fin()
	if ens.Msg.GetCreated() {
		t.Fatal("既有應 created=false")
	}
	if ens.Msg.GetProduct().GetAliasName() != "大蘋果" {
		t.Fatalf("既有別名不更動,got %q", ens.Msg.GetProduct().GetAliasName())
	}
	if n := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.DeletedAtIsNil()).CountX(ctx); n != 1 {
		t.Fatalf("一客戶一商品僅一筆,got %d", n)
	}
}

// TestIntegrationOrderSaveAlias 下單 save_alias=true 同交易 upsert 別名;選用商品自動入清單。
func TestIntegrationOrderSaveAlias(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, custID, actorID, prodID := seedOrderCompany(t, ctx, db)

	newCtx := func() (context.Context, func()) {
		tx, err := db.Tx(context.Background())
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		id := authz.Identity{UserID: uItoa(actorID), CompanyID: uItoa(coID), DepartmentID: uItoa(deptID),
			Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		c := authz.WithIdentity(context.Background(), id)
		c = dbtenant.WithTenantTx(c, tx)
		return c, func() { _ = tx.Rollback() }
	}
	commit := func(c context.Context) {
		tx, _ := dbtenant.TxFrom(c)
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交: %v", err)
		}
	}
	osvc := NewSalesOrderService(db)

	c, fin := newCtx()
	_, err := osvc.CreateOrder(c, connect.NewRequest(&salesorderv1.CreateOrderRequest{
		CustomerId: uItoa(custID), Source: "W",
		Items: []*salesorderv1.OrderItemInput{
			{ProductId: uItoa(prodID), DisplayName: "大蘋果", Qty: "3", Unit: "斤", SaveAlias: true},
		},
	}))
	if err != nil {
		fin()
		t.Fatalf("CreateOrder: %v", err)
	}
	commit(c)
	fin()
	n := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.DeletedAtIsNil()).CountX(ctx)
	if n != 1 {
		t.Fatalf("下單應自動建別名 1 筆,got %d", n)
	}
	cp := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.DeletedAtIsNil()).OnlyX(ctx)
	if cp.AliasName != "大蘋果" {
		t.Fatalf("別名應為下單名稱,got %q", cp.AliasName)
	}
}
