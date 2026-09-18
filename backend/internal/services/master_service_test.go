package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
)

// newMasterServer 於單一 enttest DB 掛上四個部門級主檔 service,並以指定身分注入。
func newMasterServer(t *testing.T, id authz.Identity) (*ent.Client, mastersv1connect.WarehouseServiceClient, mastersv1connect.RouteServiceClient, mastersv1connect.ProcessingSpecServiceClient, mastersv1connect.ProductCategoryServiceClient) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	mux := http.NewServeMux()
	RegisterWarehouseService(mux, db)
	RegisterRouteService(mux, db)
	RegisterProcessingSpecService(mux, db)
	RegisterProductCategoryService(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return db,
		mastersv1connect.NewWarehouseServiceClient(http.DefaultClient, ts.URL),
		mastersv1connect.NewRouteServiceClient(http.DefaultClient, ts.URL),
		mastersv1connect.NewProcessingSpecServiceClient(http.DefaultClient, ts.URL),
		mastersv1connect.NewProductCategoryServiceClient(http.DefaultClient, ts.URL)
}

// seedMasterDept 建立公司+部門,回傳 (coID, deptID)。
func seedMasterDept(t *testing.T, db *ent.Client, ident string) (int, int) {
	t.Helper()
	ctx := context.Background()
	co, err := db.Company.Create().SetName("主檔公司").SetIdentifier("M-" + ident).SetStatus("active").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	d, err := db.Department.Create().SetCompanyID(co.ID).SetName("門市").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	return co.ID, d.ID
}

// TestMasterWarehouseCRUD 3.4.1:create/list/update/soft-delete/restore。
func TestMasterWarehouseCRUD(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	id := deptAdminID(coID, deptID)
	_, wh, _, _, _ := newMasterServer(t, id)

	// Create。
	resp, err := wh.CreateWarehouse(ctx, connect.NewRequest(&mastersv1.CreateWarehouseRequest{Code: "WH-01", Name: "冷藏倉"}))
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	if got := resp.Msg.GetWarehouse().GetDepartmentId(); got != uItoa(deptID) {
		t.Fatalf("dept_admin 建倉應帶自己部門 %d,got %q", deptID, got)
	}
	// List 命中關鍵字。
	lr, err := wh.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{Keyword: "冷藏"}))
	if err != nil || len(lr.Msg.GetWarehouses()) != 1 {
		t.Fatalf("ListWarehouses 應 1 筆,got %d err=%v", len(lr.Msg.GetWarehouses()), err)
	}
	wid := resp.Msg.GetWarehouse().GetId()
	// Update name。
	ur, err := wh.UpdateWarehouse(ctx, connect.NewRequest(&mastersv1.UpdateWarehouseRequest{Id: wid, Name: strPtr("冷藏倉二")}))
	if err != nil {
		t.Fatalf("UpdateWarehouse: %v", err)
	}
	if ur.Msg.GetWarehouse().GetName() != "冷藏倉二" {
		t.Fatalf("更新名應生效,got %q", ur.Msg.GetWarehouse().GetName())
	}
	// Soft delete → 預設列表排除。
	if _, err := wh.DeleteWarehouse(ctx, connect.NewRequest(&mastersv1.DeleteWarehouseRequest{Id: wid})); err != nil {
		t.Fatalf("DeleteWarehouse: %v", err)
	}
	lr2, _ := wh.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{}))
	if len(lr2.Msg.GetWarehouses()) != 0 {
		t.Fatalf("軟刪除後列表應為空,got %d", len(lr2.Msg.GetWarehouses()))
	}
	// include_deleted 可見。
	lr3, _ := wh.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{IncludeDeleted: true}))
	if len(lr3.Msg.GetWarehouses()) != 1 {
		t.Fatalf("include_deleted 應含 1 筆,got %d", len(lr3.Msg.GetWarehouses()))
	}
	// Restore → 回歸列表。
	if _, err := wh.RestoreWarehouse(ctx, connect.NewRequest(&mastersv1.RestoreWarehouseRequest{Id: wid})); err != nil {
		t.Fatalf("RestoreWarehouse: %v", err)
	}
	lr4, _ := wh.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{}))
	if len(lr4.Msg.GetWarehouses()) != 1 {
		t.Fatalf("復原後列表應回 1 筆,got %d", len(lr4.Msg.GetWarehouses()))
	}
}

// TestMasterCrossDeptIsolated 3.4 共通:dept_admin 不可見他部門資源(列表不含、操作 not_found)。
func TestMasterCrossDeptIsolated(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptA := seedMasterDept(t, db, "A")
	_, deptB := seedMasterDept2(t, db, coID, "B")
	// 在 deptB 建立倉/車次。
	idB := deptAdminID(coID, deptB)
	_, whB, rtB, _, _ := newMasterServer(t, idB)
	wb, err := whB.CreateWarehouse(ctx, connect.NewRequest(&mastersv1.CreateWarehouseRequest{Code: "WB", Name: "B 倉"}))
	if err != nil {
		t.Fatalf("B 建倉: %v", err)
	}
	rb, err := rtB.CreateRoute(ctx, connect.NewRequest(&mastersv1.CreateRouteRequest{Code: "RB", Name: "B 車"}))
	if err != nil {
		t.Fatalf("B 建車: %v", err)
	}
	// deptA 身分看不到 deptB 資源。
	idA := deptAdminID(coID, deptA)
	_, whA, rtA, _, _ := newMasterServer(t, idA)
	lw, _ := whA.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{}))
	if len(lw.Msg.GetWarehouses()) != 0 {
		t.Fatalf("deptA 不應見 deptB 倉,got %d", len(lw.Msg.GetWarehouses()))
	}
	if _, err := whA.UpdateWarehouse(ctx, connect.NewRequest(&mastersv1.UpdateWarehouseRequest{Id: wb.Msg.GetWarehouse().GetId(), Name: strPtr("X")})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("跨部門 UpdateWarehouse 應 not_found,got %v", err)
	}
	if _, err := rtA.DeleteRoute(ctx, connect.NewRequest(&mastersv1.DeleteRouteRequest{Id: rb.Msg.GetRoute().GetId()})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("跨部門 DeleteRoute 應 not_found,got %v", err)
	}
}

// TestMasterProcessingSpecFlagsAndAttributes 3.4.3:多值旗標至少其一 true + attributes 儲存/讀回。
func TestMasterProcessingSpecFlagsAndAttributes(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	_, _, _, spec, _ := newMasterServer(t, deptAdminID(coID, deptID))

	// 兩旗標皆 false → invalid_argument。
	if _, err := spec.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{
		Code: "SP1", Name: "切半斤",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("兩旗標皆 false 應 invalid_argument,got %v", err)
	}
	// attributes 結構化 + 多值旗標(同時 true)。
	attrs, _ := structpb.NewStruct(map[string]any{"output_unit": "盒", "output_qty_per_portion": "0.6"})
	resp, err := spec.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{
		Code: "SP1", Name: "切半斤", Kind: "cutting",
		AppliesToProcessing: true, AppliesToPicking: true,
		Attributes: attrs,
	}))
	if err != nil {
		t.Fatalf("CreateProcessingSpec: %v", err)
	}
	got := resp.Msg.GetProcessingSpec()
	if !got.GetAppliesToProcessing() || !got.GetAppliesToPicking() {
		t.Fatal("兩旗標應皆為 true")
	}
	if got.GetAttributes().AsMap()["output_unit"] != "盒" {
		t.Fatalf("attributes 應讀回,got %v", got.GetAttributes().AsMap())
	}
	// Update:僅把 processing 改 false(picking 仍 true)→ 合併現值合法,應成功。
	sid := got.GetId()
	if _, err := spec.UpdateProcessingSpec(ctx, connect.NewRequest(&mastersv1.UpdateProcessingSpecRequest{
		Id: sid, AppliesToProcessing: boolPtr(false),
	})); err != nil {
		t.Fatalf("僅改 processing=false 且 picking 仍 true 應成功,got %v", err)
	}
	// 兩旗標皆 false → invalid_argument。
	if _, err := spec.UpdateProcessingSpec(ctx, connect.NewRequest(&mastersv1.UpdateProcessingSpecRequest{
		Id: sid, AppliesToProcessing: boolPtr(false), AppliesToPicking: boolPtr(false),
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("兩旗標皆 false 更新應 invalid_argument,got %v", err)
	}
}

// TestMasterProductCategoryBasic 3.4.4:create/list/restore 基本收斂。
func TestMasterProductCategoryBasic(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	_, _, _, _, cat := newMasterServer(t, deptAdminID(coID, deptID))
	resp, err := cat.CreateProductCategory(ctx, connect.NewRequest(&mastersv1.CreateProductCategoryRequest{Code: "BEEF", Name: "牛肉"}))
	if err != nil {
		t.Fatalf("CreateProductCategory: %v", err)
	}
	lr, _ := cat.ListProductCategories(ctx, connect.NewRequest(&mastersv1.ListProductCategoriesRequest{Keyword: "牛"}))
	if len(lr.Msg.GetProductCategories()) != 1 {
		t.Fatalf("List 應 1 筆,got %d", len(lr.Msg.GetProductCategories()))
	}
	if _, err := cat.RestoreProductCategory(ctx, connect.NewRequest(&mastersv1.RestoreProductCategoryRequest{Id: resp.Msg.GetProductCategory().GetId()})); err != nil {
		t.Fatalf("Restore(未刪除冪等)應成功,got %v", err)
	}
}

// seedMasterDept2 於既有公司再建一個部門(供跨部門測試)。
func seedMasterDept2(t *testing.T, db *ent.Client, coID int, ident string) (int, int) {
	t.Helper()
	d, err := db.Department.Create().SetCompanyID(coID).SetName("部門" + ident).Save(context.Background())
	if err != nil {
		t.Fatalf("dept2: %v", err)
	}
	return coID, d.ID
}
