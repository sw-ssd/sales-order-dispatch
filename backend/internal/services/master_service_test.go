package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
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

// assertMasterAudit 斷言某 resource/action 之稽核列數與租戶/資源 id(複審 Minor 5)。
func assertMasterAudit(t *testing.T, db *ent.Client, resource string, action auditlog.Action, rid string, cid, deptID, want int) {
	t.Helper()
	rows, err := db.AuditLog.Query().
		Where(auditlog.ResourceTypeEQ(resource), auditlog.ActionEQ(action)).All(context.Background())
	if err != nil {
		t.Fatalf("查稽核: %v", err)
	}
	if len(rows) != want {
		t.Fatalf("稽核 %s/%s 應 %d 筆,got %d", resource, action, want, len(rows))
	}
	for _, r := range rows {
		if r.ResourceID != rid {
			t.Fatalf("稽核 resource_id 應 %q,got %q", rid, r.ResourceID)
		}
		if r.CompanyID != cid {
			t.Fatalf("稽核 company_id 應 %d,got %d", cid, r.CompanyID)
		}
		if r.DepartmentID != deptID {
			t.Fatalf("稽核 department_id 應 %d,got %d", deptID, r.DepartmentID)
		}
	}
}

// TestMasterAuditRecorded 複審 Minor 5:每異動寫稽核(D18)、validation 失敗不留稽核(同交易回滾)。
func TestMasterAuditRecorded(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	_, wh, _, spec, _ := newMasterServer(t, deptAdminID(coID, deptID))

	resp, err := wh.CreateWarehouse(ctx, connect.NewRequest(&mastersv1.CreateWarehouseRequest{Code: "WH-A", Name: "A倉"}))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	wid := resp.Msg.GetWarehouse().GetId()
	assertMasterAudit(t, db, "warehouse", auditlog.ActionCreate, wid, coID, deptID, 1)

	if _, err := wh.UpdateWarehouse(ctx, connect.NewRequest(&mastersv1.UpdateWarehouseRequest{Id: wid, Name: strPtr("A倉2")})); err != nil {
		t.Fatalf("update: %v", err)
	}
	assertMasterAudit(t, db, "warehouse", auditlog.ActionUpdate, wid, coID, deptID, 1)

	if _, err := wh.DeleteWarehouse(ctx, connect.NewRequest(&mastersv1.DeleteWarehouseRequest{Id: wid})); err != nil {
		t.Fatalf("delete: %v", err)
	}
	assertMasterAudit(t, db, "warehouse", auditlog.ActionDelete, wid, coID, deptID, 1)

	// restore 亦記 action=update → update 累計 2 筆。
	if _, err := wh.RestoreWarehouse(ctx, connect.NewRequest(&mastersv1.RestoreWarehouseRequest{Id: wid})); err != nil {
		t.Fatalf("restore: %v", err)
	}
	assertMasterAudit(t, db, "warehouse", auditlog.ActionUpdate, wid, coID, deptID, 2)

	// validation 失敗(兩旗標皆 false)→ 不留任何 processing_spec 稽核。
	if _, err := spec.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{Code: "S-BAD", Name: "x"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid_argument,got %v", err)
	}
	if n, _ := db.AuditLog.Query().Where(auditlog.ResourceTypeEQ("processing_spec")).Count(ctx); n != 0 {
		t.Fatalf("validation 失敗不應留稽核,got %d", n)
	}
}

// TestProcessingSpecAttributesRoundTrip 複審 Minor 4:attributes 空/非空往返與清除語意。
func TestProcessingSpecAttributesRoundTrip(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptID := seedMasterDept(t, db, t.Name())
	_, _, _, spec, _ := newMasterServer(t, deptAdminID(coID, deptID))

	// 非空(含字串/陣列/數字/布林)應等值讀回。
	attrs, _ := structpb.NewStruct(map[string]any{
		"output_unit": "盒", "steps": []any{"去骨", "切0.6kg/盒"}, "portions": float64(4), "chilled": true,
	})
	r1, err := spec.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{
		Code: "S-NEST", Name: "n", AppliesToProcessing: true, Attributes: attrs,
	}))
	if err != nil {
		t.Fatalf("create nested: %v", err)
	}
	m := r1.Msg.GetProcessingSpec().GetAttributes().AsMap()
	if m["output_unit"] != "盒" || m["chilled"] != true {
		t.Fatalf("非空屬性應讀回,got %v", m)
	}
	if steps, ok := m["steps"].([]any); !ok || len(steps) != 2 {
		t.Fatalf("陣列應保留,got %v", m["steps"])
	}
	if p, ok := m["portions"].(float64); !ok || p != 4 {
		t.Fatalf("數字應保留,got %v", m["portions"])
	}

	// 顯式空 {} → 讀回空。
	empty, _ := structpb.NewStruct(map[string]any{})
	r2, err := spec.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{
		Code: "S-EMPTY", Name: "e", AppliesToPicking: true, Attributes: empty,
	}))
	if err != nil {
		t.Fatalf("create empty: %v", err)
	}
	if n := len(r2.Msg.GetProcessingSpec().GetAttributes().AsMap()); n != 0 {
		t.Fatalf("顯式空應讀回空,got %d", n)
	}

	// 未帶 attributes → 不報錯且為空。
	r3, err := spec.CreateProcessingSpec(ctx, connect.NewRequest(&mastersv1.CreateProcessingSpecRequest{
		Code: "S-ABSENT", Name: "a", AppliesToProcessing: true,
	}))
	if err != nil {
		t.Fatalf("create absent: %v", err)
	}
	if att := r3.Msg.GetProcessingSpec().GetAttributes(); att != nil && len(att.AsMap()) != 0 {
		t.Fatalf("未帶屬性應為空,got %v", att.AsMap())
	}

	// 更新非空 → 空 {} → 讀回空。
	if _, err := spec.UpdateProcessingSpec(ctx, connect.NewRequest(&mastersv1.UpdateProcessingSpecRequest{
		Id: r1.Msg.GetProcessingSpec().GetId(), Attributes: empty,
	})); err != nil {
		t.Fatalf("update to empty: %v", err)
	}
	lg, err := spec.ListProcessingSpecs(ctx, connect.NewRequest(&mastersv1.ListProcessingSpecsRequest{Keyword: "S-NEST"}))
	if err != nil || len(lg.Msg.GetProcessingSpecs()) != 1 {
		t.Fatalf("list: %v len=%d", err, len(lg.Msg.GetProcessingSpecs()))
	}
	if n := len(lg.Msg.GetProcessingSpecs()[0].GetAttributes().AsMap()); n != 0 {
		t.Fatalf("更新為空後應讀回空,got %d", n)
	}
}
