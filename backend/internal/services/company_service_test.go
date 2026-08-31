package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// newTestServer 建立 enttest sqlite 記憶體 client,並以 RegisterCompanyServices 掛載
// CompanyService / DepartmentService 的 Connect handler,回傳兩個 client。
func newTestServer(t *testing.T) (salesorderv1connect.CompanyServiceClient, salesorderv1connect.DepartmentServiceClient) {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })

	mux := http.NewServeMux()
	RegisterCompanyServices(mux, db)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	companyClient := salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL)
	departmentClient := salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL)
	return companyClient, departmentClient
}

func TestCompanyCRUD(t *testing.T) {
	ctx := context.Background()
	cc, _ := newTestServer(t)

	// Create:預設 status 為 active。
	created, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
		Name:       "甲股份有限公司",
		TaxId:      "12345678",
		Identifier: "co-a",
	}))
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	co := created.Msg.GetCompany()
	if co.GetId() == "" {
		t.Fatal("CreateCompany 應回傳 id")
	}
	if co.GetName() != "甲股份有限公司" || co.GetTaxId() != "12345678" || co.GetIdentifier() != "co-a" {
		t.Fatalf("CreateCompany 欄位不符: %+v", co)
	}
	if co.GetStatus() != "active" {
		t.Fatalf("預設 status 應為 active,got %q", co.GetStatus())
	}
	companyID := co.GetId()

	// Create:缺少 identifier → InvalidArgument。
	_, err = cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{Name: "乙"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺 identifier 應回 InvalidArgument,got %v", err)
	}

	// Create:identifier 重複 → AlreadyExists(ent unique constraint)。
	_, err = cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{Name: "重複", Identifier: "co-a"}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("identifier 重複應回 AlreadyExists,got %v", err)
	}

	// Get:取得單筆。
	got, err := cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: companyID}))
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if got.Msg.GetCompany().GetName() != "甲股份有限公司" {
		t.Fatalf("GetCompany 名稱不符: %+v", got.Msg.GetCompany())
	}

	// Get:不存在 → NotFound。
	_, err = cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: "999999"}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("不存在的公司應回 NotFound,got %v", err)
	}

	// Update:改 name / status。
	name := "甲股份(更名)"
	status := "suspended"
	upd, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{
		CompanyId: companyID,
		Name:      &name,
		Status:    &status,
	}))
	if err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}
	if upd.Msg.GetCompany().GetName() != name || upd.Msg.GetCompany().GetStatus() != status {
		t.Fatalf("UpdateCompany 未生效: %+v", upd.Msg.GetCompany())
	}

	// Update:identifier 出現即 InvalidArgument(建立後不可修改)。
	identifier := "co-x"
	_, err = cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{
		CompanyId:  companyID,
		Identifier: identifier,
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("改 identifier 應回 InvalidArgument,got %v", err)
	}

	// Update:非法 status → InvalidArgument。
	bad := "bogus"
	_, err = cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{
		CompanyId: companyID,
		Status:    &bad,
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法 status 應回 InvalidArgument,got %v", err)
	}

	// List:分頁 + keyword 篩選。
	list, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{
		Page:     1,
		PageSize: 10,
		Keyword:  "更名",
	}))
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if list.Msg.GetPagination().GetTotal() != 1 || len(list.Msg.GetCompanies()) != 1 {
		t.Fatalf("keyword 篩選結果不符: %+v", list.Msg)
	}
	if list.Msg.GetCompanies()[0].GetId() != companyID {
		t.Fatalf("keyword 篩選應命中公司 %s", companyID)
	}

	// List:status 篩選(inactive 應為 0 筆)。
	list, err = cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{Status: "inactive"}))
	if err != nil {
		t.Fatalf("ListCompanies(status): %v", err)
	}
	if list.Msg.GetPagination().GetTotal() != 0 {
		t.Fatalf("status 篩選應為 0 筆,got %d", list.Msg.GetPagination().GetTotal())
	}

	// Delete:刪除後 Get → NotFound。
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: companyID})); err != nil {
		t.Fatalf("DeleteCompany: %v", err)
	}
	_, err = cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: companyID}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("刪除後應回 NotFound,got %v", err)
	}
}

func TestCompanyDeleteBlockedByDepartment(t *testing.T) {
	ctx := context.Background()
	cc, dc := newTestServer(t)

	created, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
		Name:       "丙有限公司",
		Identifier: "co-c",
	}))
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	companyID := created.Msg.GetCompany().GetId()

	if _, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
		CompanyId: companyID,
		Name:      "業務部",
	})); err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	// 仍有部門的公司不可刪除(FK 阻擋)→ FailedPrecondition。
	_, err = cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: companyID}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("有部門的公司刪除應回 FailedPrecondition,got %v", err)
	}
}

func TestDepartmentCRUD(t *testing.T) {
	ctx := context.Background()
	cc, dc := newTestServer(t)

	coA, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
		Name:       "丁有限公司",
		Identifier: "co-d",
	}))
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	coAID := coA.Msg.GetCompany().GetId()

	coB, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
		Name:       "戊有限公司",
		Identifier: "co-e",
	}))
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	coBID := coB.Msg.GetCompany().GetId()

	// CreateDepartment:不存在的公司 → NotFound。
	_, err = dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
		CompanyId: "999999",
		Name:      "孤兒部門",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("不存在的公司應回 NotFound,got %v", err)
	}

	// CreateDepartment:缺名稱 → InvalidArgument。
	_, err = dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{CompanyId: coAID}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺名稱應回 InvalidArgument,got %v", err)
	}

	depA, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
		CompanyId: coAID,
		Name:      "業務部",
	}))
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	depAID := depA.Msg.GetDepartment().GetId()
	if depA.Msg.GetDepartment().GetCompanyId() != coAID || depA.Msg.GetDepartment().GetCompanyName() != "丁有限公司" {
		t.Fatalf("CreateDepartment 回傳應含公司資訊: %+v", depA.Msg.GetDepartment())
	}

	if _, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
		CompanyId: coBID,
		Name:      "採購部",
	})); err != nil {
		t.Fatalf("CreateDepartment(coB): %v", err)
	}

	// ListDepartments:company_id 篩選僅回 coA 的部門。
	list, err := dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{
		Page:      1,
		PageSize:  10,
		CompanyId: coAID,
	}))
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	if list.Msg.GetPagination().GetTotal() != 1 || len(list.Msg.GetDepartments()) != 1 {
		t.Fatalf("company_id 篩選結果不符: %+v", list.Msg)
	}
	if list.Msg.GetDepartments()[0].GetCompanyName() != "丁有限公司" {
		t.Fatalf("部門應帶公司名稱: %+v", list.Msg.GetDepartments()[0])
	}

	// ListDepartments:全部(2 筆)。
	list, err = dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{Page: 1, PageSize: 10}))
	if err != nil {
		t.Fatalf("ListDepartments(all): %v", err)
	}
	if list.Msg.GetPagination().GetTotal() != 2 {
		t.Fatalf("全部部門應為 2 筆,got %d", list.Msg.GetPagination().GetTotal())
	}

	// GetDepartment。
	got, err := dc.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: depAID}))
	if err != nil {
		t.Fatalf("GetDepartment: %v", err)
	}
	if got.Msg.GetDepartment().GetName() != "業務部" {
		t.Fatalf("GetDepartment 名稱不符: %+v", got.Msg.GetDepartment())
	}

	// UpdateDepartment:改名。
	newName := "業務一課"
	upd, err := dc.UpdateDepartment(ctx, connect.NewRequest(&v1.UpdateDepartmentRequest{
		DepartmentId: depAID,
		Name:         &newName,
	}))
	if err != nil {
		t.Fatalf("UpdateDepartment: %v", err)
	}
	if upd.Msg.GetDepartment().GetName() != newName {
		t.Fatalf("UpdateDepartment 未生效: %+v", upd.Msg.GetDepartment())
	}

	// UpdateDepartment:清空名稱 → InvalidArgument。
	empty := ""
	_, err = dc.UpdateDepartment(ctx, connect.NewRequest(&v1.UpdateDepartmentRequest{
		DepartmentId: depAID,
		Name:         &empty,
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("空名稱應回 InvalidArgument,got %v", err)
	}

	// DeleteDepartment。
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: depAID})); err != nil {
		t.Fatalf("DeleteDepartment: %v", err)
	}
	_, err = dc.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: depAID}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("刪除後應回 NotFound,got %v", err)
	}
}
