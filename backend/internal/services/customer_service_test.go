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
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
)

// newCustomerTestServer 建立 CustomerService client 並注入身分 + 稽核來源。
func newCustomerTestServer(t *testing.T, id authz.Identity) (customersv1connect.CustomerServiceClient, *ent.Client) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	mux := http.NewServeMux()
	RegisterCustomerServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL), db
}

// seedCustomerCompany 建立含前綴的 active 公司,回傳其 id 與預設部門 id(或 -1=無)。
func seedCustomerCompany(t *testing.T, db *ent.Client, prefix string, withDept bool) (int, int) {
	t.Helper()
	ctx := context.Background()
	co, err := db.Company.Create().SetName("客戶公司").SetIdentifier("C-" + t.Name()).SetStatus("active").SetCustomerCodePrefix(prefix).Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	if !withDept {
		return co.ID, -1
	}
	d, err := db.Department.Create().SetCompanyID(co.ID).SetName("門市一").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	return co.ID, d.ID
}

// deptAdminID 組裝 dept_admin 身分。
func deptAdminID(coID, deptID int) authz.Identity {
	return authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptID), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
}

// TestCreateCustomerGeneratesCodeAndCounter D7:dept_admin 建立客戶 → 取號 TY000001、counter 推進、稽核存在。
func TestCreateCustomerGeneratesCodeAndCounter(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TY", true)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))

	resp, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王小明"}))
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}
	if got := resp.Msg.GetCustomer().GetCustomerCode(); got != "TY000001" {
		t.Fatalf("customer_code 應為 TY000001,得到 %q", got)
	}
	if resp.Msg.GetCustomer().GetDepartmentId() != uItoa(deptID) {
		t.Fatalf("dept_admin 建檔應自動帶自己部門 %d,得到 %q", deptID, resp.Msg.GetCustomer().GetDepartmentId())
	}
	// counter 推進。
	cnt, err := db.CustomerCounter.Query().Only(ctx)
	if err != nil || cnt.NextSeq != 2 {
		t.Fatalf("counter next_seq 應為 2,得到 %+v (err=%v)", cnt, err)
	}
	// 稽核存在。
	if n, _ := db.AuditLog.Query().Count(ctx); n != 1 {
		t.Fatalf("建檔應寫 1 筆稽核,得到 %d", n)
	}
	// 第二個客戶取號連續。
	resp2, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "李四"}))
	if err != nil {
		t.Fatalf("CreateCustomer2: %v", err)
	}
	if resp2.Msg.GetCustomer().GetCustomerCode() != "TY000002" {
		t.Fatalf("第二客戶編號應為 TY000002,得到 %q", resp2.Msg.GetCustomer().GetCustomerCode())
	}
}

// TestCreateCustomerWithoutPrefixFailedPrecondition:公司未設前綴 → failed_precondition,不建檔。
func TestCreateCustomerWithoutPrefixFailedPrecondition(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "", true)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王"})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("無前綴建檔應 failed_precondition,got %v", err)
	}
	if n, _ := db.Customer.Query().Count(ctx); n != 0 {
		t.Fatalf("無前綴時不應建檔,得到 %d 列", n)
	}
}

// TestCreateCustomerAsCustomerDenied:客戶主帳號呼叫一律 permission_denied。
func TestCreateCustomerAsCustomerDenied(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	_ = db
	client, _ := newCustomerTestServer(t, authz.Identity{UserID: "9", CompanyID: "1", Role: "customer", Roles: []string{"customer"}})
	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王"})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("customer 建客戶應 permission_denied,got %v", err)
	}
}

// TestListKeywordAndIncludeDeleted:關鍵字模糊比對(名稱/編號/統編)+ include_deleted。
func TestListKeywordAndIncludeDeleted(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TK", true)
	// 直接 seed 一筆客户(跳過 RPC,便於控制 deleted)。
	c1, err := db.Customer.Create().SetCompanyID(coID).SetDepartmentID(deptID).
		SetCustomerCode("TK000001").SetName("王小明").SetTaxID("X").Save(ctx)
	if err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))

	// 關鍵字命中名稱。
	resp, err := client.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{Keyword: "王小明"}))
	if err != nil {
		t.Fatalf("ListCustomers: %v", err)
	}
	if resp.Msg.GetPagination().GetTotal() != 1 || resp.Msg.GetCustomers()[0].GetCustomerCode() != "TK000001" {
		t.Fatalf("關鍵字應命中 1 筆,得到 %d", resp.Msg.GetPagination().GetTotal())
	}
	// 軟刪除後:預設排除(include_deleted=false)。
	ts := time.Now().UTC()
	db.Customer.UpdateOneID(c1.ID).SetDeletedAt(ts).SaveX(ctx)
	respD, err := client.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{}))
	if err != nil {
		t.Fatalf("ListCustomers(deleted): %v", err)
	}
	if respD.Msg.GetPagination().GetTotal() != 0 {
		t.Fatalf("軟刪除後預設應排除,得到 %d", respD.Msg.GetPagination().GetTotal())
	}
	// include_deleted=true → 可見。
	respI, err := client.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{IncludeDeleted: true}))
	if err != nil {
		t.Fatalf("ListCustomers(include_deleted): %v", err)
	}
	if respI.Msg.GetPagination().GetTotal() != 1 {
		t.Fatalf("include_deleted 應回 1 筆,得到 %d", respI.Msg.GetPagination().GetTotal())
	}
}

// TestUpdateKeepsCodeAndDeleteRestore:更新 name 不改 code;Delete 軟刪後 Restore 復原。
func TestUpdateKeepsCodeAndDeleteRestore(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TU", true)
	c, err := db.Customer.Create().SetCompanyID(coID).SetDepartmentID(deptID).
		SetCustomerCode("TU000001").SetName("原名").Save(ctx)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))

	// Update name,code 不變。
	u, err := client.UpdateCustomer(ctx, connect.NewRequest(&customersv1.UpdateCustomerRequest{Id: uItoa(c.ID), Name: strPtr("新名")}))
	if err != nil {
		t.Fatalf("UpdateCustomer: %v", err)
	}
	if u.Msg.GetCustomer().GetName() != "新名" || u.Msg.GetCustomer().GetCustomerCode() != "TU000001" {
		t.Fatalf("update 應改 name 且不改 code,得到 %s/%s", u.Msg.GetCustomer().GetName(), u.Msg.GetCustomer().GetCustomerCode())
	}
	// Delete 軟刪除。
	if _, err := client.DeleteCustomer(ctx, connect.NewRequest(&customersv1.DeleteCustomerRequest{Id: uItoa(c.ID)})); err != nil {
		t.Fatalf("DeleteCustomer: %v", err)
	}
	fresh := db.Customer.GetX(ctx, c.ID)
	if fresh.DeletedAt == nil {
		t.Fatal("Delete 應設 deleted_at")
	}
	// Get(不含刪除) → not_found。
	if _, err := client.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: uItoa(c.ID)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪客戶 Get 應 not_found,got %v", err)
	}
	// Restore 復原。
	if _, err := client.RestoreCustomer(ctx, connect.NewRequest(&customersv1.RestoreCustomerRequest{Id: uItoa(c.ID)})); err != nil {
		t.Fatalf("RestoreCustomer: %v", err)
	}
	if got := db.Customer.GetX(ctx, c.ID); got.DeletedAt != nil {
		t.Fatal("Restore 應清 deleted_at")
	}
}

// TestCustomerScopeCrossDept:dept_admin 查他部門客戶 → not_found(範圍隔離)。
func TestCustomerScopeCrossDept(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	_ = deptB
	other, err := db.Customer.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetCustomerCode("X000001").SetName("他部門").Save(ctx)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptA))
	if _, err := client.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: uItoa(other.ID)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("跨部門 Get 應 not_found,got %v", err)
	}
}
