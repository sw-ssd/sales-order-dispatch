package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// newTestServer 以 super 身分建立測試 server(既有 CRUD 測試沿用;super 全權)。
func newTestServer(t *testing.T) (salesorderv1connect.CompanyServiceClient, salesorderv1connect.DepartmentServiceClient) {
	t.Helper()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	cc, dc, _ := newTestServerWithIdentity(t, super)
	return cc, dc
}

// newTestServerWithIdentity 建立 enttest sqlite 記憶體 client,並以 RegisterCompanyServices
// 掛載 CompanyService / DepartmentService 的 Connect handler(與 newRoleTestServer 同構:
// 注入身分 + CASL 開關 + DB),回傳兩個 client 與 db。
func newTestServerWithIdentity(t *testing.T, id authz.Identity) (salesorderv1connect.CompanyServiceClient, salesorderv1connect.DepartmentServiceClient, *ent.Client) {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:comp?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })

	mux := http.NewServeMux()
	RegisterCompanyServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL),
		salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL),
		db
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

// TestCompanyServiceAuthorization P1-2 驗收:company 資源授權門檻(比照 RoleService 的
// requireRole 模式)——read/update 限 super/company_admin,create/delete 僅 super;
// guest/staff/dept_admin 一律拒絕;未登入 → Unauthenticated。
func TestCompanyServiceAuthorization(t *testing.T) {
	ctx := context.Background()

	t.Run("未登入 Unauthenticated", func(t *testing.T) {
		cc, _, _ := newTestServerWithIdentity(t, authz.Identity{})
		_, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("未登入應 Unauthenticated,got %v", err)
		}
	})

	t.Run("guest/staff/dept_admin 一律 PermissionDenied", func(t *testing.T) {
		for _, id := range []authz.Identity{
			{UserID: "u3", CompanyID: "c1", Role: "guest", Roles: []string{"guest"}},
			{UserID: "u4", CompanyID: "c1", Role: "staff", Roles: []string{"staff"}},
			{UserID: "u5", CompanyID: "c1", Role: "dept_admin", Roles: []string{"dept_admin"}},
		} {
			cc, _, _ := newTestServerWithIdentity(t, id)
			_, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{}))
			if connect.CodeOf(err) != connect.CodePermissionDenied {
				t.Fatalf("%s ListCompanies 應 PermissionDenied,got %v", id.Role, err)
			}
		}
	})

	t.Run("company_admin 可 read/update 但不可 create/delete 公司", func(t *testing.T) {
		cc, _, db := newTestServerWithIdentity(t, authz.Identity{UserID: "u2", CompanyID: "c1", Role: "company_admin", Roles: []string{"company_admin"}})
		co := db.Company.Create().SetName("公司A").SetIdentifier("A-1").SaveX(ctx)

		list, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{}))
		if err != nil {
			t.Fatalf("company_admin ListCompanies: %v", err)
		}
		if len(list.Msg.GetCompanies()) != 1 {
			t.Fatalf("companies = %d, want 1", len(list.Msg.GetCompanies()))
		}
		if _, err := cc.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: strconvID(co.ID)})); err != nil {
			t.Fatalf("company_admin GetCompany: %v", err)
		}
		if _, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{CompanyId: strconvID(co.ID), Name: protoStr("公司A改")})); err != nil {
			t.Fatalf("company_admin UpdateCompany: %v", err)
		}
		if _, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{Name: "公司B", Identifier: "B-1"})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("company_admin CreateCompany 應 PermissionDenied,got %v", err)
		}
		if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: strconvID(co.ID)})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("company_admin DeleteCompany 應 PermissionDenied,got %v", err)
		}
	})

	t.Run("super 可完整 CRUD 公司", func(t *testing.T) {
		cc, _, db := newTestServerWithIdentity(t, authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}})
		co := db.Company.Create().SetName("公司A").SetIdentifier("A-1").SaveX(ctx)

		created, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{Name: "公司B", Identifier: "B-1"}))
		if err != nil {
			t.Fatalf("super CreateCompany: %v", err)
		}
		if created.Msg.GetCompany().GetIdentifier() != "B-1" {
			t.Fatalf("created identifier = %q, want B-1", created.Msg.GetCompany().GetIdentifier())
		}
		if _, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{CompanyId: strconvID(co.ID), Name: protoStr("公司A改")})); err != nil {
			t.Fatalf("super UpdateCompany: %v", err)
		}
		if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: created.Msg.GetCompany().GetId()})); err != nil {
			t.Fatalf("super DeleteCompany: %v", err)
		}
	})
}

// TestDepartmentServiceAuthorization P1-2 驗收:department 資源授權門檻——read 限
// company_admin/dept_admin,write 限 company_admin;staff/guest 拒絕;未登入 Unauthenticated。
func TestDepartmentServiceAuthorization(t *testing.T) {
	ctx := context.Background()

	t.Run("未登入 Unauthenticated", func(t *testing.T) {
		_, dc, _ := newTestServerWithIdentity(t, authz.Identity{})
		_, err := dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("未登入應 Unauthenticated,got %v", err)
		}
	})

	t.Run("staff/guest 一律 PermissionDenied", func(t *testing.T) {
		for _, id := range []authz.Identity{
			{UserID: "u3", CompanyID: "c1", Role: "staff", Roles: []string{"staff"}},
			{UserID: "u4", CompanyID: "c1", Role: "guest", Roles: []string{"guest"}},
		} {
			_, dc, _ := newTestServerWithIdentity(t, id)
			_, err := dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{}))
			if connect.CodeOf(err) != connect.CodePermissionDenied {
				t.Fatalf("%s ListDepartments 應 PermissionDenied,got %v", id.Role, err)
			}
		}
	})

	t.Run("dept_admin 可 read 但不可 write 部門", func(t *testing.T) {
		_, dc, db := newTestServerWithIdentity(t, authz.Identity{UserID: "u5", CompanyID: "c1", Role: "dept_admin", Roles: []string{"dept_admin"}})
		co := db.Company.Create().SetName("公司A").SetIdentifier("A-1").SaveX(ctx)
		dep := db.Department.Create().SetName("門市一").SetCompanyID(co.ID).SaveX(ctx)

		list, err := dc.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{}))
		if err != nil {
			t.Fatalf("dept_admin ListDepartments: %v", err)
		}
		if len(list.Msg.GetDepartments()) != 1 {
			t.Fatalf("departments = %d, want 1", len(list.Msg.GetDepartments()))
		}
		if _, err := dc.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: strconvID(dep.ID)})); err != nil {
			t.Fatalf("dept_admin GetDepartment: %v", err)
		}
		if _, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{CompanyId: strconvID(co.ID), Name: "門市二"})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("dept_admin CreateDepartment 應 PermissionDenied,got %v", err)
		}
		if _, err := dc.UpdateDepartment(ctx, connect.NewRequest(&v1.UpdateDepartmentRequest{DepartmentId: strconvID(dep.ID), Name: protoStr("門市一改")})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("dept_admin UpdateDepartment 應 PermissionDenied,got %v", err)
		}
		if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: strconvID(dep.ID)})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("dept_admin DeleteDepartment 應 PermissionDenied,got %v", err)
		}
	})

	t.Run("company_admin 可 read/write 部門", func(t *testing.T) {
		_, dc, db := newTestServerWithIdentity(t, authz.Identity{UserID: "u2", CompanyID: "c1", Role: "company_admin", Roles: []string{"company_admin"}})
		co := db.Company.Create().SetName("公司A").SetIdentifier("A-1").SaveX(ctx)
		dep := db.Department.Create().SetName("門市一").SetCompanyID(co.ID).SaveX(ctx)

		if _, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{CompanyId: strconvID(co.ID), Name: "門市二"})); err != nil {
			t.Fatalf("company_admin CreateDepartment: %v", err)
		}
		if _, err := dc.UpdateDepartment(ctx, connect.NewRequest(&v1.UpdateDepartmentRequest{DepartmentId: strconvID(dep.ID), Name: protoStr("門市一改")})); err != nil {
			t.Fatalf("company_admin UpdateDepartment: %v", err)
		}
		if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: strconvID(dep.ID)})); err != nil {
			t.Fatalf("company_admin DeleteDepartment: %v", err)
		}
	})
}

func protoStr(s string) *string { return &s }
