//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// newReturnServer 以請求交易掛 ReturnService(身分 + RLS scope 注入)。
// 客戶子帳號 → self scope + customer_id;員工 → department/company。
func newReturnServer(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.ReturnServiceClient {
	t.Helper()
	path, handler := salesorderv1connect.NewReturnServiceHandler(NewReturnService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		scope := auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true}
		if id.Role == "company_admin" || id.Role == "super" {
			scope.DataScope = auth.DataScopeCompany
		}
		if id.Role == "customer" {
			scope.DataScope = auth.DataScopeSelf
			scope.CustomerID = id.CustomerID
		}
		ctx = auth.WithRLS(ctx, scope)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewReturnServiceClient(http.DefaultClient, ts.URL)
}

// seedReturnShop 建公司/部門/客戶/主帳號/子帳號/商品/專屬/訂單品項，回 ids。
// 子帳號 users.is_primary=false + customer_id；訂單 processing 含一明細。
type returnIDs struct {
	co, dept, cust, primary, sub, prod, cp, orderItem int
}

func seedReturnShop(t *testing.T, ctx context.Context, db *ent.Client) returnIDs {
	t.Helper()
	var v returnIDs
	seedTx(t, db, func(tx *ent.Tx) error {
		co, err := tx.Company.Create().SetName("退貨公司").SetIdentifier("R-" + t.Name()).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		d, err := tx.Department.Create().SetCompanyID(co.ID).SetName("門市一").Save(ctx)
		if err != nil {
			return err
		}
		cust, err := tx.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetCustomerCode("RT000001").SetName("王小明").Save(ctx)
		if err != nil {
			return err
		}
		pri, err := tx.User.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetEmail("pri-" + t.Name() + "@t.com").SetName("主帳號").
			SetRole("customer").SetPasswordHash("x").
			SetIsCustomer(true).SetCustomerID(cust.ID).SetIsPrimary(true).Save(ctx)
		if err != nil {
			return err
		}
		sub, err := tx.User.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetEmail("sub-" + t.Name() + "@t.com").SetName("子帳號").
			SetRole("customer").SetPasswordHash("x").
			SetIsCustomer(true).SetCustomerID(cust.ID).SetIsPrimary(false).Save(ctx)
		if err != nil {
			return err
		}
		prod, err := tx.Product.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetCode("P1").SetName("蘋果").Save(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.ProductUnit.Create().SetProductID(prod.ID).
			SetUnitCode("斤").SetConversionRate("1").SetIsBase(true).Save(ctx); err != nil {
			return err
		}
		cp, err := tx.CustomerProduct.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetCustomerID(cust.ID).SetProductID(prod.ID).SetAliasName("紅蘋果").Save(ctx)
		if err != nil {
			return err
		}
		o, err := tx.SalesOrder.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
			SetCustomerID(cust.ID).SetOrderNo("W000001").SetSource("W").
			SetExpectedDeliveryDate(dateOf(2026, 9, 22)).SetStatus("processing").
			SetCreatedBy(pri.ID).Save(ctx)
		if err != nil {
			return err
		}
		oi, err := tx.SalesOrderItem.Create().SetSalesOrderID(o.ID).
			SetCompanyID(co.ID).SetDepartmentID(d.ID).SetProductID(prod.ID).
			SetDisplayName("蘋果").SetQty("10").SetUnit("斤").SetBaseQty("10").Save(ctx)
		if err != nil {
			return err
		}
		v = returnIDs{co: co.ID, dept: d.ID, cust: cust.ID, primary: pri.ID,
			sub: sub.ID, prod: prod.ID, cp: cp.ID, orderItem: oi.ID}
		return nil
	})
	return v
}

// subIdentity 組子帳號身分。
func subIdentity(v returnIDs, uid int) authz.Identity {
	return authz.Identity{UserID: strconv.Itoa(uid), CompanyID: strconv.Itoa(v.co),
		DepartmentID: strconv.Itoa(v.dept), CustomerID: strconv.Itoa(v.cust),
		Role: "customer", Roles: []string{"customer"}}
}

// TestIntegrationReturnCreateListGet 雙來源建單 → List 自查 → Get 明細快照。
func TestIntegrationReturnCreateListGet(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedReturnShop(t, ctx, db)
	rpc := newReturnServer(t, db, subIdentity(v, v.sub))

	// order_item 來源建單。
	cr, err := rpc.CreateReturnRequest(ctx, connect.NewRequest(&salesorderv1.CreateReturnRequestRequest{
		Items: []*salesorderv1.ReturnItemInput{{
			SourceType: "order_item", SalesOrderItemId: strconv.Itoa(v.orderItem),
			Quantity: "2", Reason: "爛果",
		}},
		Remark: "首單退貨",
	}))
	if err != nil {
		t.Fatalf("Create(order_item): %v", err)
	}
	if cr.Msg.GetStatus() != "pending" {
		t.Fatalf("新建應為 pending,got %q", cr.Msg.GetStatus())
	}
	// customer_product 來源建單。
	if _, err := rpc.CreateReturnRequest(ctx, connect.NewRequest(&salesorderv1.CreateReturnRequestRequest{
		Items: []*salesorderv1.ReturnItemInput{{
			SourceType: "customer_product", CustomerProductId: strconv.Itoa(v.cp),
			Quantity: "1", Reason: "太酸",
		}},
	})); err != nil {
		t.Fatalf("Create(customer_product): %v", err)
	}
	// List 自查應見 2 筆。
	ls, err := rpc.ListReturnRequests(ctx, connect.NewRequest(&salesorderv1.ListReturnRequestsRequest{
		Page: 1, PageSize: 20,
	}))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if ls.Msg.GetTotal() != 2 {
		t.Fatalf("List 應見 2 筆,got %d", ls.Msg.GetTotal())
	}
	// Get 明細:order_item 快照品名 + 單位。
	gt, err := rpc.GetReturnRequest(ctx, connect.NewRequest(&salesorderv1.GetReturnRequestRequest{
		Id: cr.Msg.GetId(),
	}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(gt.Msg.GetItems()) != 1 {
		t.Fatalf("明細應 1 筆,got %d", len(gt.Msg.GetItems()))
	}
	if gt.Msg.GetItems()[0].GetProductName() != "蘋果" {
		t.Fatalf("快照品名應為 蘋果,got %q", gt.Msg.GetItems()[0].GetProductName())
	}
	// 樂觀鎖版本須由讀取端帶回(新申請 version=0;審核必帶 expected_version)。
	if gt.Msg.GetVersion() != "0" {
		t.Fatalf("新申請 version 應為 0,got %q", gt.Msg.GetVersion())
	}
	// 主帳號 Create → permission_denied。
	rpcPri := newReturnServer(t, db, subIdentity(v, v.primary))
	if _, err := rpcPri.CreateReturnRequest(ctx, connect.NewRequest(&salesorderv1.CreateReturnRequestRequest{
		Items: []*salesorderv1.ReturnItemInput{{
			SourceType: "order_item", SalesOrderItemId: strconv.Itoa(v.orderItem),
			Quantity: "1", Reason: "x",
		}},
	})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("主帳號應 permission_denied,got %v", err)
	}
	// 空品項 → invalid_argument。
	if _, err := rpc.CreateReturnRequest(ctx, connect.NewRequest(&salesorderv1.CreateReturnRequestRequest{})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("空品項應 invalid_argument,got %v", err)
	}
}
