// 配額守衛的掛點矩陣與語意測試（Plan D Task 5b / Plan B Task 6；spec §4.5）。
//
// 兩層證據，缺一不可：
//
//  1. 表驅動（TestGuardMatrixUsesExpectedFeature）：每個寫入 RPC 都必須以**對應的 feature**
//     呼叫 CheckLimit —— 漏掛、掛錯 feature、或一次請求檢查兩次都是紅。以記錄式假物件驗
//     「RPC → feature 的對應關係」，而不是只驗「有呼叫」。
//  2. 語意（TestCreateUserBlockedAtSeatLimit／TestCreateCustomerBlockedForDeptAdmin…）：真的接上
//     entitlements.Service（sqlite ＋ store.Fake）—— 達上限 → failed_precondition ＋
//     ErrorInfo.code=PLAT-5001、被擋後不落庫、停用釋放席位後可再建。
//
// **待補第 7 項（具名缺口）**：spec §4.5 的清單列了「部門復原」（CompanyService.CreateDepartment
// 4.5 原文為「`CompanyService.CreateDepartment` / 部門復原」），但 repo 的 DepartmentService
// 只有 List/Get/Create/Update/Delete —— 全 repo（含 generated Go）皆無 `RestoreDepartment`，
// 00020 的部門軟刪除只做了 Delete 側 → **無對應 RPC 可掛**。部門復原 RPC 落地後，必須在
// guardCases 補上 {DepartmentService.RestoreDepartment, entitlements.LimitDepartments} 並掛守衛。
// 缺口的另一半（刪除路徑不消耗配額）不需要守衛：軟刪除只減不增。
package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// guardCases 為 spec §4.5 的守衛清單：RPC → feature。新增寫入 RPC 未登記即紅
// （漏掛＝該 RPC 完全不檢查配額）。見檔頭的「待補第 7 項」。
var guardCases = []struct {
	name    string
	feature string
}{
	{"UserService.CreateUser", entitlements.LimitSeats},
	{"CustomerService.CreateCustomer", entitlements.LimitCustomers},
	{"CustomerService.RestoreCustomer", entitlements.LimitCustomers},
	{"ProductService.CreateProduct", entitlements.LimitProducts},
	{"ProductService.RestoreProduct", entitlements.LimitProducts},
	{"DepartmentService.CreateDepartment", entitlements.LimitDepartments},
}

// guardRecorder 為記錄式假物件：只實作守衛用到的那個方法（介面就這麼窄），
// 記錄被檢查的 feature 與租戶 id。
type guardRecorder struct {
	calls []string
	cids  []int
	// results 讓特定 feature 回錯（模擬「方案未含／已達上限」）；nil map 全放行。
	results map[string]error
}

func (r *guardRecorder) CheckLimit(_ context.Context, companyID int, feature string, _ int) error {
	r.calls = append(r.calls, feature)
	r.cids = append(r.cids, companyID)
	return r.results[feature]
}

// TestGuardMatrixUsesExpectedFeature 逐個寫入 RPC 斷言「恰好檢查一次、且是對應的 feature」。
// 呼叫本身必須成功（否則記錄到的呼叫可能來自失敗路徑，矩陣就變成假綠）。
func TestGuardMatrixUsesExpectedFeature(t *testing.T) {
	for _, tc := range guardCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &guardRecorder{results: map[string]error{}}
			runGuardedRPC(t, rec, tc.name)
			if len(rec.calls) != 1 || rec.calls[0] != tc.feature {
				t.Fatalf("%s 應恰好檢查一次 %s,got %v", tc.name, tc.feature, rec.calls)
			}
		})
	}
}

// TestGuardUsesIdentityCompanyNotRequestedCompany 守衛檢查的租戶必須是**身分的公司**，
// 不得採用請求帶入的 company_id：這裡的操作者是公司 A 的 company_admin，請求指定公司 B。
// CreateDepartment 的權限閘門（company_admin 具 department:*）與「公司存在」檢查都不比對公司，
// 所以請求的 id 會一路帶到守衛 —— 若守衛採用它，就能拿 B 的額度替 A 的寫入背書。
// 斷言被檢查的租戶是 A（cids），而不是 B。
func TestGuardUsesIdentityCompanyNotRequestedCompany(t *testing.T) {
	db := guardDB(t)
	coA, _, _ := seedUserCompany(t, db)
	coB := db.Company.Create().SetName("公司B").SetIdentifier("co-b").SaveX(t.Context())
	id := authz.Identity{UserID: "1", CompanyID: uItoa(coA), Role: "company_admin", Roles: []string{"company_admin"}}
	rec := &guardRecorder{results: map[string]error{}}
	client := salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient,
		guardURL(t, db, id, companyScope(coA), rec, func(m *http.ServeMux, e entitlementChecker) {
			RegisterCompanyServices(m, db, e)
		}))
	if _, err := client.CreateDepartment(t.Context(), connect.NewRequest(&v1.CreateDepartmentRequest{
		CompanyId: uItoa(coB.ID), Name: "掛到別家",
	})); err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	if len(rec.cids) != 1 || rec.cids[0] != coA {
		t.Fatalf("守衛檢查的租戶應為身分公司 %d（不得用請求帶入的 %d）,got %v", coA, coB.ID, rec.cids)
	}
}

// runGuardedRPC 依 case 名建立對應域的最小 fixture（沿用各域既有測試的建構樣板），
// 注入 rec 為權益服務並呼叫該 RPC。
func runGuardedRPC(t *testing.T, rec *guardRecorder, name string) {
	t.Helper()
	ctx := t.Context()
	switch name {

	case "UserService.CreateUser":
		db := guardDB(t)
		co, _, _ := seedUserCompany(t, db)
		id := authz.Identity{UserID: "1", CompanyID: uItoa(co), Role: "company_admin", Roles: []string{"company_admin"}}
		client := salesorderv1connect.NewUserServiceClient(http.DefaultClient,
			guardURL(t, db, id, companyScope(co), rec, func(m *http.ServeMux, e entitlementChecker) {
				RegisterUserServices(m, db, e)
			}))
		if _, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
			Name: "守衛新人", Email: "guard-user@t.com", CompanyId: uItoa(co), Role: "staff",
		})); err != nil {
			t.Fatalf("CreateUser: %v", err)
		}

	case "CustomerService.CreateCustomer":
		db := guardDB(t)
		co, dept := seedCustomerCompany(t, db, "GD", true)
		rep := seedCustomerRep(t, db, co, dept)
		client := customersv1connect.NewCustomerServiceClient(http.DefaultClient,
			guardURL(t, db, deptAdminID(co, dept), departmentScope(co, dept), rec,
				func(m *http.ServeMux, e entitlementChecker) {
					RegisterCustomerServices(m, db, "http://localhost:3000", e)
				}))
		if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
			Name: "守衛客戶", DefaultSalesRepId: uItoa(rep),
		})); err != nil {
			t.Fatalf("CreateCustomer: %v", err)
		}

	case "CustomerService.RestoreCustomer":
		db := guardDB(t)
		co, dept := seedCustomerCompany(t, db, "GD", true)
		deleted := db.Customer.Create().SetCompanyID(co).SetDepartmentID(dept).
			SetCustomerCode("GD900001").SetName("已刪客戶").SetDeletedAt(time.Now()).SaveX(ctx)
		client := customersv1connect.NewCustomerServiceClient(http.DefaultClient,
			guardURL(t, db, deptAdminID(co, dept), departmentScope(co, dept), rec,
				func(m *http.ServeMux, e entitlementChecker) {
					RegisterCustomerServices(m, db, "http://localhost:3000", e)
				}))
		if _, err := client.RestoreCustomer(ctx, connect.NewRequest(&customersv1.RestoreCustomerRequest{
			Id: uItoa(deleted.ID),
		})); err != nil {
			t.Fatalf("RestoreCustomer: %v", err)
		}

	case "ProductService.CreateProduct":
		db := guardDB(t)
		co, dept := seedMasterDept(t, db, t.Name())
		catID, whID, specID := seedProductEnv(t, db, co, dept)
		client := productsv1connect.NewProductServiceClient(http.DefaultClient,
			guardURL(t, db, deptAdminID(co, dept), departmentScope(co, dept), rec,
				func(m *http.ServeMux, e entitlementChecker) { RegisterProductService(m, db, e) }))
		if _, err := client.CreateProduct(ctx, connect.NewRequest(validProductReq(catID, whID, specID))); err != nil {
			t.Fatalf("CreateProduct: %v", err)
		}

	case "ProductService.RestoreProduct":
		db := guardDB(t)
		co, dept := seedMasterDept(t, db, t.Name())
		deleted := db.Product.Create().SetCompanyID(co).SetDepartmentID(dept).
			SetCode("P-RST").SetName("已刪商品").SetDeletedAt(time.Now()).SaveX(ctx)
		client := productsv1connect.NewProductServiceClient(http.DefaultClient,
			guardURL(t, db, deptAdminID(co, dept), departmentScope(co, dept), rec,
				func(m *http.ServeMux, e entitlementChecker) { RegisterProductService(m, db, e) }))
		if _, err := client.RestoreProduct(ctx, connect.NewRequest(&productsv1.RestoreProductRequest{
			Id: uItoa(deleted.ID),
		})); err != nil {
			t.Fatalf("RestoreProduct: %v", err)
		}

	case "DepartmentService.CreateDepartment":
		db := guardDB(t)
		co, _, _ := seedUserCompany(t, db)
		id := authz.Identity{UserID: "1", CompanyID: uItoa(co), Role: "company_admin", Roles: []string{"company_admin"}}
		client := salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient,
			guardURL(t, db, id, companyScope(co), rec, func(m *http.ServeMux, e entitlementChecker) {
				RegisterCompanyServices(m, db, e)
			}))
		if _, err := client.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
			CompanyId: uItoa(co), Name: "守衛部門",
		})); err != nil {
			t.Fatalf("CreateDepartment: %v", err)
		}

	default:
		t.Fatalf("guardCases 有未實作的 case %q：新增 case 必須同時補上 fixture 與呼叫", name)
	}
}

// TestCreateUserBlockedAtSeatLimit 方案席位上限 10、已有 10 席 → CreateUser 回 FailedPrecondition
// ＋ ErrorInfo.code=PLAT-5001，且**不落庫**；停用一席（釋放配額，spec §3.2 規則 1）後可再建。
// 以 sqlite ＋ store.Fake 覆蓋完整語意，不需 testcontainers。
func TestCreateUserBlockedAtSeatLimit(t *testing.T) {
	ctx := t.Context()
	db := guardDB(t)
	co, _, _ := seedUserCompany(t, db)
	for i := range 10 {
		db.User.Create().SetCompanyID(co).SetEmail(fmt.Sprintf("seat%d@t.com", i)).SetName("佔位").
			SetRole("staff").SetStatus(user.StatusActive).SetPasswordHash("x").SaveX(ctx)
	}
	entSvc := guardEntitlements(t, db, co, entitlements.LimitSeats, 10)

	id := authz.Identity{UserID: "1", CompanyID: uItoa(co), Role: "company_admin", Roles: []string{"company_admin"}}
	client := salesorderv1connect.NewUserServiceClient(http.DefaultClient,
		guardURL(t, db, id, companyScope(co), entSvc, func(m *http.ServeMux, e entitlementChecker) {
			RegisterUserServices(m, db, e)
		}))

	_, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "第 11 人", Email: "over@t.com", CompanyId: uItoa(co), Role: "staff",
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("已達席位上限應回 failed_precondition（引導升級方案），got %v", err)
	}
	// 斷言**碼**而非只有 connect 碼：前端據 ErrorInfo.code 導向升級方案（Plan D T5b）。
	info := errorInfoOf(t, err)
	if got := info.GetCode(); got != "PLAT-5001" {
		t.Fatalf("席位超限必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}
	if info.GetDetails()["used"] != "10" || info.GetDetails()["limit"] != "10" {
		t.Fatalf("用量細節應為 used=10/limit=10，got %v", info.GetDetails())
	}
	if n, _ := db.User.Query().Where(user.HasCompanyWith(company.ID(co))).Count(ctx); n != 10 {
		t.Fatalf("被擋後不得新增帳號，got %d", n)
	}

	// 停用一席即釋放配額 → 可再建。
	db.User.Update().Where(user.EmailEQ("seat0@t.com")).SetStatus(user.StatusInactive).SaveX(ctx)
	if _, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "遞補", Email: "refill@t.com", CompanyId: uItoa(co), Role: "staff",
	})); err != nil {
		t.Fatalf("停用釋放席位後應可建立: %v", err)
	}
}

// TestCreateCustomerBlockedForDeptAdminAtCompanyLimit 配額是**公司層**的（spec §3.2 的
// WHERE company_id=?），不是每部門一份：dept_admin 的請求 scope 是 department，若守衛就著請求
// 交易數，只會數到本部門 → 低報 → 超額放行（fail-open，T5 修的就是這個）。
//
// 這裡以 department scope 的請求（DataScopeDepartment → 計數走獨立的系統範圍交易）驗證
// 「公司共 10 位客戶（6 在本部門、4 在另一部門）、上限 10 → 建第 11 位必須被擋」。
// 真 RLS 下的完整證明在 entitlement_guard_integration_test.go（sqlite 沒有 RLS，看不出低報）。
func TestCreateCustomerBlockedForDeptAdminAtCompanyLimit(t *testing.T) {
	ctx := t.Context()
	db := guardDB(t)
	co, deptA := seedCustomerCompany(t, db, "GD", true)
	deptB := db.Department.Create().SetCompanyID(co).SetName("門市二").SaveX(ctx).ID
	rep := seedCustomerRep(t, db, co, deptA)
	// 本部門 6 位 + 另一部門 4 位 = 公司 10 位（若只數本部門會是 6 < 10 → 放行）。
	for i := range 6 {
		db.Customer.Create().SetCompanyID(co).SetDepartmentID(deptA).
			SetCustomerCode(fmt.Sprintf("GD10%04d", i)).SetName("本部門客戶").SaveX(ctx)
	}
	for i := range 4 {
		db.Customer.Create().SetCompanyID(co).SetDepartmentID(deptB).
			SetCustomerCode(fmt.Sprintf("GD20%04d", i)).SetName("他部門客戶").SaveX(ctx)
	}
	entSvc := guardEntitlements(t, db, co, entitlements.LimitCustomers, 10)

	client := customersv1connect.NewCustomerServiceClient(http.DefaultClient,
		guardURL(t, db, deptAdminID(co, deptA), departmentScope(co, deptA), entSvc,
			func(m *http.ServeMux, e entitlementChecker) {
				RegisterCustomerServices(m, db, "http://localhost:3000", e)
			}))

	_, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
		Name: "超額客戶", DefaultSalesRepId: uItoa(rep),
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("公司層已達客戶上限時 dept_admin 不得放行（fail-open 回歸），got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-5001" {
		t.Fatalf("客戶超限必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}
	if n, _ := db.Customer.Query().Where(customer.CompanyIDEQ(co)).Count(ctx); n != 10 {
		t.Fatalf("被擋後不得新增客戶，got %d", n)
	}
}

// guardDB 建立 sqlite 記憶體 client（各域既有測試同款）。
func guardDB(t *testing.T) *ent.Client {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// guardURL 以 entSvc 掛載單一域 handler（mount 由呼叫端給，等同各域既有 newXTestServer），
// 注入身分／DB／稽核來源與 RLS scope —— 生產的 scope 由 authzMiddleware 依身分導出，
// 假 store 的守衛測試必須照樣注入，才走到與生產相同的計數分支。
func guardURL(t *testing.T, db *ent.Client, id authz.Identity, scope auth.RLSScope,
	entSvc entitlementChecker, mount func(*http.ServeMux, entitlementChecker)) string {
	t.Helper()
	mux := http.NewServeMux()
	mount(mux, entSvc)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = auth.WithRLS(ctx, scope)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "127.0.0.1", UserAgent: "guard-test"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return ts.URL
}

// guardEntitlements 建一個「方案 std 對 feature 限 limit、訂閱 active」的判定服務（ttl=0 不快取，
// 免測試間互相汙染）。計數器用真品（NewEntitlementCounter）—— 守衛的低報 bug 就在它裡面。
func guardEntitlements(t *testing.T, db *ent.Client, companyID int, feature string, limit int64) *entitlements.Service {
	t.Helper()
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: feature, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: feature, Enabled: true, Limit: ptr(limit)}})
	f.PutSubscription(store.Subscription{CompanyID: companyID, PlanCode: "std", Status: "active"})
	return entitlements.New(f, NewEntitlementCounter(db), entitlements.NewMemoryCache(), 0)
}

// companyScope／departmentScope 組出與生產 authzMiddleware 相同的 RLS scope。
func companyScope(companyID int) auth.RLSScope {
	return auth.RLSScope{DataScope: auth.DataScopeCompany, CompanyID: uItoa(companyID), CompanyActive: true}
}

func departmentScope(companyID, departmentID int) auth.RLSScope {
	return auth.RLSScope{
		DataScope:     auth.DataScopeDepartment,
		CompanyID:     uItoa(companyID),
		DepartmentID:  uItoa(departmentID),
		CompanyActive: true,
	}
}

// ptr 為本檔的指標輔助（internal/services 原本沒有；entitlements 套件的同名 helper 屬另一 package，
// 不會衝突）。
func ptr[T any](v T) *T { return &v }
