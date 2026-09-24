package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
	"google.golang.org/protobuf/proto"
)

// newProductTestServer 於單一 enttest DB 掛上 ProductService,並以指定身分注入。
func newProductTestServer(t *testing.T, id authz.Identity) (*ent.Client, productsv1connect.ProductServiceClient) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	mux := http.NewServeMux()
	RegisterProductService(mux, db, entitlements.Unlimited())
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return db, productsv1connect.NewProductServiceClient(http.DefaultClient, ts.URL)
}

// seedProductEnv 建立商品所需之參考主檔與 unit 字典,回傳 (categoryID, warehouseID, specID)。
func seedProductEnv(t *testing.T, db *ent.Client, coID, deptID int) (int, int, int) {
	t.Helper()
	ctx := context.Background()
	seedMetadictSystem(t, db, "unit", "KG", "公斤")
	seedMetadictSystem(t, db, "unit", "BOX", "盒")
	c := db.ProductCategory.Create().SetCompanyID(coID).SetDepartmentID(deptID).SetCode("CAT").SetName("分類").SaveX(ctx)
	w := db.Warehouse.Create().SetCompanyID(coID).SetDepartmentID(deptID).SetCode("WH").SetName("冷藏倉").SaveX(ctx)
	s := db.ProcessingSpec.Create().SetCompanyID(coID).SetDepartmentID(deptID).SetCode("SP").SetName("切塊").SetAppliesToProcessing(true).SetKind("cutting").SaveX(ctx)
	return c.ID, w.ID, s.ID
}

// validProductReq 組裝一組合法的 CreateProductRequest(KG 基本 + BOX 換算,一規格,一倉別/分類)。
func validProductReq(catID, whID, specID int) *productsv1.CreateProductRequest {
	return &productsv1.CreateProductRequest{
		Code: "P-001", Name: "牛肉", CategoryId: uItoa(catID),
		InventoryWarehouseId: uItoa(whID), PickingWarehouseId: uItoa(whID), IsActive: true,
		Units: []*productsv1.ProductUnit{
			{UnitCode: "KG", ConversionRate: "1", IsBase: true, SortOrder: 1},
			{UnitCode: "BOX", ConversionRate: "5", IsBase: false, SortOrder: 2, SizeDesc: "1盒=5斤"},
		},
		ProcessingSpecs: []*productsv1.ProductProcessingSpec{{ProcessingSpecId: uItoa(specID)}},
	}
}

// TestProductCreateGetUpdateDeleteRestore 3.3.2:建立(含單位/規格關聯)→ Get 讀回 → 更新整組替換 →
// 軟刪除(列表排除)→ 復原。
func TestProductCreateGetUpdateDeleteRestore(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	catID, whID, specID := seedProductEnv(t, db, coID, deptID)
	_, client := newProductTestServer(t, deptAdminID(coID, deptID))

	resp, err := client.CreateProduct(ctx, connect.NewRequest(validProductReq(catID, whID, specID)))
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	p := resp.Msg.GetProduct()
	if p.GetName() != "牛肉" || p.GetDepartmentId() != uItoa(deptID) {
		t.Fatalf("商品建立欄位不符: name=%q dept=%q", p.GetName(), p.GetDepartmentId())
	}
	if len(p.GetUnits()) != 2 {
		t.Fatalf("應有 2 單位,got %d", len(p.GetUnits()))
	}
	if len(p.GetProcessingSpecs()) != 1 {
		t.Fatalf("應有 1 規格關聯,got %d", len(p.GetProcessingSpecs()))
	}
	pid := p.GetId()

	// Get 讀回含巢狀。
	g, err := client.GetProduct(ctx, connect.NewRequest(&productsv1.GetProductRequest{Id: pid}))
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if len(g.Msg.GetProduct().GetUnits()) != 2 || !g.Msg.GetProduct().GetUnits()[0].GetIsBase() {
		t.Fatalf("GetProduct 巢狀單位不符")
	}
	if got := g.Msg.GetProduct().GetUnits()[1].GetSizeDesc(); got != "1盒=5斤" {
		t.Fatalf("box size_desc 應為 1盒=5斤,got %q", got)
	}

	// Update:改名 + units 整組替換為單一基本單位。
	ur, err := client.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{
		Id: pid, Name: strPtr("牛肉二"),
		Units: []*productsv1.ProductUnit{{UnitCode: "KG", ConversionRate: "1", IsBase: true, SortOrder: 1}},
	}))
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if ur.Msg.GetProduct().GetName() != "牛肉二" || len(ur.Msg.GetProduct().GetUnits()) != 1 {
		t.Fatalf("更新後 name/units 不符: name=%q units=%d", ur.Msg.GetProduct().GetName(), len(ur.Msg.GetProduct().GetUnits()))
	}
	// 處理規格關聯仍保留(未提供即沿用)。
	g2, _ := client.GetProduct(ctx, connect.NewRequest(&productsv1.GetProductRequest{Id: pid}))
	if len(g2.Msg.GetProduct().GetProcessingSpecs()) != 1 {
		t.Fatalf("未提供 processing_specs 應沿用既有 1 筆,got %d", len(g2.Msg.GetProduct().GetProcessingSpecs()))
	}

	// 軟刪除 → 預設列表排除。
	if _, err := client.DeleteProduct(ctx, connect.NewRequest(&productsv1.DeleteProductRequest{Id: pid})); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}
	ll, _ := client.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{}))
	if len(ll.Msg.GetProducts()) != 0 {
		t.Fatalf("軟刪除後列表應空,got %d", len(ll.Msg.GetProducts()))
	}
	li, _ := client.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{IncludeDeleted: true}))
	if len(li.Msg.GetProducts()) != 1 {
		t.Fatalf("include_deleted 應含 1 筆,got %d", len(li.Msg.GetProducts()))
	}
	// 復原。
	if _, err := client.RestoreProduct(ctx, connect.NewRequest(&productsv1.RestoreProductRequest{Id: pid})); err != nil {
		t.Fatalf("RestoreProduct: %v", err)
	}
	lr, _ := client.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{}))
	if len(lr.Msg.GetProducts()) != 1 {
		t.Fatalf("復原後列表應回 1 筆,got %d", len(lr.Msg.GetProducts()))
	}
}

// TestProductUnitValidation 3.3.2 步驟 4:無基本單位/多基本單位/基本率≠1/換算率非正/非法 unit_code/重複單位 → invalid_argument。
func TestProductUnitValidation(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	seedMetadictSystem(t, db, "unit", "KG", "公斤")
	seedMetadictSystem(t, db, "unit", "BOX", "盒")
	_, client := newProductTestServer(t, deptAdminID(coID, deptID))

	cases := []struct {
		name string
		req  *productsv1.CreateProductRequest
	}{
		{"無單位", &productsv1.CreateProductRequest{Code: "P", Name: "X"}},
		{"多基本單位", &productsv1.CreateProductRequest{Code: "P", Name: "X", Units: []*productsv1.ProductUnit{
			{UnitCode: "KG", ConversionRate: "1", IsBase: true}, {UnitCode: "BOX", ConversionRate: "5", IsBase: true}}}},
		{"基本率非1", &productsv1.CreateProductRequest{Code: "P", Name: "X", Units: []*productsv1.ProductUnit{
			{UnitCode: "KG", ConversionRate: "5", IsBase: true}}}},
		{"換算率非正", &productsv1.CreateProductRequest{Code: "P", Name: "X", Units: []*productsv1.ProductUnit{
			{UnitCode: "KG", ConversionRate: "1", IsBase: true}, {UnitCode: "BOX", ConversionRate: "0", IsBase: false}}}},
		{"非法unit_code", &productsv1.CreateProductRequest{Code: "P", Name: "X", Units: []*productsv1.ProductUnit{
			{UnitCode: "NOPE", ConversionRate: "1", IsBase: true}}}},
		{"重複單位", &productsv1.CreateProductRequest{Code: "P", Name: "X", Units: []*productsv1.ProductUnit{
			{UnitCode: "KG", ConversionRate: "1", IsBase: true}, {UnitCode: "KG", ConversionRate: "2", IsBase: false}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.CreateProduct(ctx, connect.NewRequest(tc.req))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("%s: 期望 invalid_argument,got %v", tc.name, err)
			}
		})
	}
}

// TestProductCrossDeptRef 3.3.2:引用跨部門倉別 → invalid_argument。
func TestProductCrossDeptRef(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptA := seedMasterDept(t, db, "A")
	_, deptB := seedMasterDept2(t, db, coID, "B")
	catA, _, specA := seedProductEnv(t, db, coID, deptA)
	// 在 deptB 建立倉別。
	wB := db.Warehouse.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetCode("WB").SetName("B倉").SaveX(ctx)
	_, client := newProductTestServer(t, deptAdminID(coID, deptA))
	req := &productsv1.CreateProductRequest{
		Code: "P-002", Name: "雞肉", CategoryId: uItoa(catA),
		InventoryWarehouseId: uItoa(wB.ID), PickingWarehouseId: uItoa(wB.ID), IsActive: true,
		Units:           []*productsv1.ProductUnit{{UnitCode: "KG", ConversionRate: "1", IsBase: true, SortOrder: 1}},
		ProcessingSpecs: []*productsv1.ProductProcessingSpec{{ProcessingSpecId: uItoa(specA)}},
	}
	_, err := client.CreateProduct(ctx, connect.NewRequest(req))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("跨部門倉別應 invalid_argument,got %v", err)
	}
}

// TestProductCrossDeptIsolated 3.3.2:dept_admin 不可見/不可取他部門商品。
func TestProductCrossDeptIsolated(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptA := seedMasterDept(t, db, "A")
	_, deptB := seedMasterDept2(t, db, coID, "B")
	// deptB 建商品。
	catB, whB, specB := seedProductEnv(t, db, coID, deptB)
	_, clientB := newProductTestServer(t, deptAdminID(coID, deptB))
	rb, err := clientB.CreateProduct(ctx, connect.NewRequest(validProductReq(catB, whB, specB)))
	if err != nil {
		t.Fatalf("B 建商品: %v", err)
	}
	pidB := rb.Msg.GetProduct().GetId()
	// deptA 身分列表不含、Get not_found。
	_, clientA := newProductTestServer(t, deptAdminID(coID, deptA))
	ll, _ := clientA.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{}))
	if len(ll.Msg.GetProducts()) != 0 {
		t.Fatalf("A 列表應不含 B 商品,got %d", len(ll.Msg.GetProducts()))
	}
	_, err = clientA.GetProduct(ctx, connect.NewRequest(&productsv1.GetProductRequest{Id: pidB}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("A 取 B 商品應 not_found,got %v", err)
	}
}

// TestUpdateProductClearFlags 「取消全部關聯」的契約(2026-09-24)。
//
// 為何需要旗標:proto3 的 repeated **無 presence** —— 送 `[]` 在線上與「未提供」無法區分
// (實測 protojson.Unmarshal(`{"processingSpecs":[]}`) 得到 nil slice),故空陣列清不掉既有關聯。
// repeated 也不能加 `optional`(buf 拒收:multiple modifiers),因此以 clear_* 旗標表達清空。
//
// 同時帶旗標與非空陣列 = 自相矛盾 → invalid_argument,不默默擇一。
func TestUpdateProductClearFlags(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	catID, whID, specID := seedProductEnv(t, db, coID, deptID)
	_, client := newProductTestServer(t, deptAdminID(coID, deptID))

	created, err := client.CreateProduct(ctx, connect.NewRequest(validProductReq(catID, whID, specID)))
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	id := created.Msg.GetProduct().GetId()
	if n := len(created.Msg.GetProduct().GetProcessingSpecs()); n != 1 {
		t.Fatalf("建立後應有 1 筆規格關聯,got %d", n)
	}

	// 送空陣列(無旗標)= 線上等於未提供 → 維持現值(proto3 的既有限制,不是 bug)。
	keep, err := client.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{
		Id: id, ProcessingSpecs: []*productsv1.ProductProcessingSpec{},
	}))
	if err != nil {
		t.Fatalf("UpdateProduct(空陣列): %v", err)
	}
	if n := len(keep.Msg.GetProduct().GetProcessingSpecs()); n != 1 {
		t.Fatalf("空陣列不應清空(線上等於未提供),應仍為 1 筆,got %d", n)
	}

	// 帶旗標 → 整組清空(這正是 Web 取消全部勾選時走的路徑)。
	cleared, err := client.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{
		Id: id, ClearProcessingSpecs: proto.Bool(true),
	}))
	if err != nil {
		t.Fatalf("UpdateProduct(clear): %v", err)
	}
	if n := len(cleared.Msg.GetProduct().GetProcessingSpecs()); n != 0 {
		t.Fatalf("clear_processing_specs 應清空全部關聯,got %d", n)
	}

	// 自相矛盾:旗標 + 非空陣列 → invalid_argument。
	_, err = client.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{
		Id: id, ClearProcessingSpecs: proto.Bool(true),
		ProcessingSpecs: []*productsv1.ProductProcessingSpec{{ProcessingSpecId: uItoa(specID)}},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("同時帶 clear 旗標與非空陣列應為 invalid_argument,got %v", err)
	}
}
