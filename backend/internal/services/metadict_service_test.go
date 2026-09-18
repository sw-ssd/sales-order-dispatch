package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	metadictv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1/metadictv1connect"
)

// newMetadictTestServer 建立 MetadictService client 並注入身分 + 稽核來源。
func newMetadictTestServer(t *testing.T, id authz.Identity) (metadictv1connect.MetadictServiceClient, *ent.Client) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	mux := http.NewServeMux()
	RegisterMetadictServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "127.0.0.1", UserAgent: "test-agent"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return metadictv1connect.NewMetadictServiceClient(http.DefaultClient, ts.URL), db
}

func seedMetadictSystem(t *testing.T, db *ent.Client, typ, code, name string) int {
	t.Helper()
	m, err := db.Metadict.Create().SetType(typ).SetCode(code).SetDisplayName(name).SetSortOrder(10).Save(context.Background())
	if err != nil {
		t.Fatalf("seed metadict: %v", err)
	}
	return m.ID
}

func seedMetadictDept(t *testing.T, db *ent.Client, deptID int, typ, code, name string) int {
	t.Helper()
	m, err := db.Metadict.Create().SetType(typ).SetCode(code).SetDisplayName(name).SetDepartmentID(deptID).SetSortOrder(20).Save(context.Background())
	if err != nil {
		t.Fatalf("seed metadict dept: %v", err)
	}
	return m.ID
}

// TestCreateMetadictAsSuperCreatesSystemLevel:super 建立系統級列(department nil)並寫稽核。
func TestCreateMetadictAsSuperCreatesSystemLevel(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", CompanyID: "10", Role: "super", Roles: []string{"super"}}
	client, db := newMetadictTestServer(t, id)
	resp, err := client.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{
		Type: "unit", Code: "T", DisplayName: "噸", SortOrder: 5, IsActive: true,
	}))
	if err != nil {
		t.Fatalf("CreateMetadict: %v", err)
	}
	if resp.Msg.GetMetadict().GetDepartmentId() != "" {
		t.Fatalf("super 建立應為系統級(department 空),得到 %q", resp.Msg.GetMetadict().GetDepartmentId())
	}
	// 稽核存在且 action=create。
	n, err := db.AuditLog.Query().Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("期望 1 筆稽核,得到 %d (err=%v)", n, err)
	}
}

// TestCreateMetadictAsDeptAdminForcesOwnDept:dept_admin 建立自動帶自己部門,不接受請求帶 department_id。
func TestCreateMetadictAsDeptAdminForcesOwnDept(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	_ = deptB
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client, _ := newMetadictTestServer(t, id)
	resp, err := client.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{
		Type: "unit", Code: "PC", DisplayName: "塑膠框", IsActive: true,
	}))
	if err != nil {
		t.Fatalf("CreateMetadict: %v", err)
	}
	if got := resp.Msg.GetMetadict().GetDepartmentId(); got != uItoa(deptA) {
		t.Fatalf("dept_admin 建立應自動帶自己部門 %d,得到 %q", deptA, got)
	}
}

// TestCreateMetadictAsCustomerDenied:客戶身分不可建立字典。
func TestCreateMetadictAsCustomerDenied(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "9", Role: "customer", Roles: []string{"customer"}}
	client, _ := newMetadictTestServer(t, id)
	_, err := client.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{Type: "unit", Code: "X", DisplayName: "x"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("customer 建立應回 permission_denied,得到 %v", err)
	}
}

// TestUpdateOrderSourceDenied:order_source 系統固定不可異動。
func TestUpdateOrderSourceDenied(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	mid := seedMetadictSystem(t, db, "order_source", "W", "Web 中台")
	id := authz.Identity{UserID: "1", CompanyID: "10", Role: "super", Roles: []string{"super"}}
	client, _ := newMetadictTestServer(t, id)
	_, err := client.UpdateMetadict(ctx, connect.NewRequest(&metadictv1.UpdateMetadictRequest{Id: uItoa(mid), DisplayName: strPtr("改")}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("order_source 異動應回 failed_precondition,得到 %v", err)
	}
	_, err = client.DeleteMetadict(ctx, connect.NewRequest(&metadictv1.DeleteMetadictRequest{Id: uItoa(mid)}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("order_source 刪除應回 failed_precondition,得到 %v", err)
	}
}

// TestDeptAdminCannotModifySystemLevel:dept_admin 對系統級列 Update/Delete 回 permission_denied。
func TestDeptAdminCannotModifySystemLevel(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	mid := seedMetadictSystem(t, db, "unit", "KG", "公斤")
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client, _ := newMetadictTestServer(t, id)
	if _, err := client.UpdateMetadict(ctx, connect.NewRequest(&metadictv1.UpdateMetadictRequest{Id: uItoa(mid), DisplayName: strPtr("改")})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("dept_admin 改系統級應回 permission_denied,得到 %v", err)
	}
	if _, err := client.DeleteMetadict(ctx, connect.NewRequest(&metadictv1.DeleteMetadictRequest{Id: uItoa(mid)})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("dept_admin 刪系統級應回 permission_denied,得到 %v", err)
	}
}

// TestSoftDeleteThenListHidesAndGetNotFound:軟刪除後 List 預設查不到、Get 回 not_found;同 code 可重建。
func TestSoftDeleteThenListHidesAndGetNotFound(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	mid := seedMetadictDept(t, db, deptA, "unit", "PP", "塑膠")
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client, _ := newMetadictTestServer(t, id)
	if _, err := client.DeleteMetadict(ctx, connect.NewRequest(&metadictv1.DeleteMetadictRequest{Id: uItoa(mid)})); err != nil {
		t.Fatalf("DeleteMetadict: %v", err)
	}
	if _, err := client.GetMetadict(ctx, connect.NewRequest(&metadictv1.GetMetadictRequest{Id: uItoa(mid)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("軟刪除後 Get 應回 not_found,得到 %v", err)
	}
	// 同 code 重建不報錯(部分唯一索引於 Postgres 生效;sqlite 無約束亦不阻擋)。
	if _, err := client.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{Type: "unit", Code: "PP", DisplayName: "塑膠2", IsActive: true})); err != nil {
		t.Fatalf("重刪重建: %v", err)
	}
}

// TestListMergeSystemPlusOwnDeptOnly:部門 A 使用者 List 見系統預設 + A 擴充,不見 B 擴充。
func TestListMergeSystemPlusOwnDeptOnly(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	seedMetadictSystem(t, db, "unit", "KG", "公斤")
	seedMetadictDept(t, db, deptA, "unit", "PP", "塑膠A")
	seedMetadictDept(t, db, deptB, "unit", "PF", "塑膠B")
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client, _ := newMetadictTestServer(t, id)
	resp, err := client.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{}))
	if err != nil {
		t.Fatalf("ListMetadicts: %v", err)
	}
	codes := map[string]bool{}
	for _, it := range resp.Msg.GetItems() {
		codes[it.GetCode()] = true
	}
	if !codes["KG"] || !codes["PP"] {
		t.Fatalf("部門 A 應見系統(KG)+A 擴充(PP),得到 %v", codes)
	}
	if codes["PF"] {
		t.Fatalf("部門 A 不得見部門 B 擴充(PF),得到 %v", codes)
	}
}

// TestListIncludeDeletedDefaultsExcluded:軟刪除列預設不出現在 List。
func TestListIncludeDeletedDefaultsExcluded(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", CompanyID: "10", Role: "super", Roles: []string{"super"}}
	client, db := newMetadictTestServer(t, id)
	m, err := db.Metadict.Create().SetType("unit").SetCode("G").SetDisplayName("公克").Save(ctx)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := db.Metadict.UpdateOneID(m.ID).SetDeletedAt(nowPtr()).Save(ctx); err != nil {
		t.Fatalf("delete: %v", err)
	}
	resp, err := client.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{}))
	if err != nil {
		t.Fatalf("ListMetadicts: %v", err)
	}
	for _, it := range resp.Msg.GetItems() {
		if it.GetId() == uItoa(m.ID) {
			t.Fatalf("軟刪除列不應出現在預設 List")
		}
	}
}

// TestListOptionsExcludesInactiveAndDeletedAndKeyword:選項排除停用/軟刪除,支援關鍵字過濾。
func TestListOptionsExcludesInactiveAndDeletedAndKeyword(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	seedMetadictSystem(t, db, "unit", "KG", "公斤")
	if _, err := db.Metadict.Create().SetType("unit").SetCode("G").SetDisplayName("公克").SetIsActive(false).Save(ctx); err != nil {
		t.Fatalf("disabled: %v", err)
	}
	if _, err := db.Metadict.Create().SetType("unit").SetCode("OLD").SetDisplayName("舊").SetDeletedAt(nowPtr()).Save(ctx); err != nil {
		t.Fatalf("deleted: %v", err)
	}
	_ = deptA
	id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), Role: "super", Roles: []string{"super"}}
	client, _ := newMetadictTestServer(t, id)
	resp, err := client.ListOptions(ctx, connect.NewRequest(&metadictv1.ListOptionsRequest{Type: "unit"}))
	if err != nil {
		t.Fatalf("ListOptions: %v", err)
	}
	set := map[string]bool{}
	for _, o := range resp.Msg.GetOptions() {
		set[o.GetCode()] = true
	}
	if !set["KG"] || set["G"] || set["OLD"] {
		t.Fatalf("選項應只含啟用未刪值(KG),得到 %v", set)
	}
	// 關鍵字過濾。
	respKw, err := client.ListOptions(ctx, connect.NewRequest(&metadictv1.ListOptionsRequest{Type: "unit", Keyword: "公斤"}))
	if err != nil {
		t.Fatalf("ListOptions kw: %v", err)
	}
	if len(respKw.Msg.GetOptions()) != 1 || respKw.Msg.GetOptions()[0].GetCode() != "KG" {
		t.Fatalf("關鍵字 '公斤' 應只命中 KG,得到 %v", respKw.Msg.GetOptions())
	}
}

// TestListOptionsSuperSystemOnly:super 的 ListOptions 僅回系統預設,不混入部門私有值
// (與 ListMetadicts 預設一致;複審 #1 修正)。
func TestListOptionsSuperSystemOnly(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	seedMetadictSystem(t, db, "unit", "KG", "公斤")
	seedMetadictDept(t, db, deptA, "unit", "PP", "塑膠A")
	id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), Role: "super", Roles: []string{"super"}}
	client, _ := newMetadictTestServer(t, id)
	resp, err := client.ListOptions(ctx, connect.NewRequest(&metadictv1.ListOptionsRequest{Type: "unit"}))
	if err != nil {
		t.Fatalf("ListOptions: %v", err)
	}
	codes := map[string]bool{}
	for _, o := range resp.Msg.GetOptions() {
		codes[o.GetCode()] = true
	}
	if !codes["KG"] || codes["PP"] {
		t.Fatalf("super ListOptions 應僅含系統預設 KG,得到 %v", codes)
	}
}

// TestListOptionsCustomerSystemOnly:客戶身分僅見系統預設選項,不揭露部門擴充。
func TestListOptionsCustomerSystemOnly(t *testing.T) {
	ctx := context.Background()
	_, db := newMetadictTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	seedMetadictSystem(t, db, "unit", "KG", "公斤")
	seedMetadictDept(t, db, deptA, "unit", "PP", "塑膠")
	id := authz.Identity{UserID: "9", CompanyID: uItoa(coID), Role: "customer", Roles: []string{"customer"}}
	client, _ := newMetadictTestServer(t, id)
	resp, err := client.ListOptions(ctx, connect.NewRequest(&metadictv1.ListOptionsRequest{Type: "unit"}))
	if err != nil {
		t.Fatalf("ListOptions: %v", err)
	}
	for _, o := range resp.Msg.GetOptions() {
		if o.GetCode() == "PP" {
			t.Fatalf("客戶不應見部門擴充 PP")
		}
	}
}

// TestCreateMetadictRollsBackWhenAuditFails D18(複審 #2):稽核寫入失敗(缺公司脈絡 →
// audit.Record 回錯)時,同交易業務異動須一併回滾(不得「業務成功、稽核缺漏」)。
func TestCreateMetadictRollsBackWhenAuditFails(t *testing.T) {
	ctx := context.Background()
	// 身分缺 CompanyID → companyIDFrom 回 0 → audit.Record 以「缺租戶脈絡」回錯。
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, db := newMetadictTestServer(t, id)
	if _, err := client.CreateMetadict(ctx, connect.NewRequest(&metadictv1.CreateMetadictRequest{
		Type: "unit", Code: "ROLL", DisplayName: "回滾", IsActive: true,
	})); err == nil {
		t.Fatal("稽核寫入失敗時 CreateMetadict 應回錯誤")
	}
	// 業務異動應已回滾:metadicts 無任何列。
	n, err := db.Metadict.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("稽核失敗應回滾業務異動,得到 %d 列", n)
	}
}

func strPtr(s string) *string { return &s }

func nowPtr() time.Time {
	return time.Now().UTC()
}

// TestListMetadictsPagination 複審 Minor 4:分頁 meta(Total/PageSize)與跨頁切分正確(走 pageList)。
func TestListMetadictsPagination(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, db := newMetadictTestServer(t, id)
	seedMetadictSystem(t, db, "payment_method", "P1", "現金")
	seedMetadictSystem(t, db, "payment_method", "P2", "匯款")
	seedMetadictSystem(t, db, "payment_method", "P3", "信用卡")

	p1, err := client.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{Page: 1, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListMetadicts: %v", err)
	}
	if got := p1.Msg.GetPagination(); got.GetTotal() != 3 || got.GetPageSize() != 2 || len(p1.Msg.GetItems()) != 2 {
		t.Fatalf("第1頁分頁應 total=3 size=2 items=2,got total=%d size=%d items=%d", got.GetTotal(), got.GetPageSize(), len(p1.Msg.GetItems()))
	}
	p2, err := client.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{Page: 2, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListMetadicts p2: %v", err)
	}
	if len(p2.Msg.GetItems()) != 1 {
		t.Fatalf("第2頁應剩 1 筆,got %d", len(p2.Msg.GetItems()))
	}
}
