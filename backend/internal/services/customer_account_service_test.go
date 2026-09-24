package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
)

// accountEnv 為店家自助帳號管理的測試環境。
type accountEnv struct {
	db     *ent.Client
	coID   int
	deptID int
	custID int
	priID  int // 主帳號
	sysSub int // 建檔自動附帶的業務子帳號
	client customersv1connect.CustomerAccountServiceClient
}

// accountEnvSeq 讓同一測試內可建多個獨立環境(DSN 與 identifier 都需唯一)。
var accountEnvSeq int

// newAccountEnv 建立公司/部門/客戶 + 主帳號與業務子帳號,並以指定身分掛上服務。
func newAccountEnv(t *testing.T, asPrimary bool) *accountEnv {
	t.Helper()
	accountEnvSeq++
	suffix := t.Name() + "-" + strconv.Itoa(accountEnvSeq)
	db := enttest.Open(t, "sqlite3", "file:"+suffix+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	co := db.Company.Create().SetName("店家公司").SetIdentifier("ACC-" + suffix).
		SetStatus("active").SaveX(ctx)
	dept := db.Department.Create().SetCompanyID(co.ID).SetName("門市").SaveX(ctx)
	cust := db.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCustomerCode("ACC001").SetName("連鎖店").SaveX(ctx)
	pri := db.User.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetEmail("pri-acc@t.com").SetName("主帳號").SetRole("customer").SetPasswordHash("x").
		SetAccountName("連鎖店").SetIsCustomer(true).SetCustomerID(cust.ID).
		SetIsPrimary(true).SetStatus(user.StatusActive).SaveX(ctx)
	// 建檔自動附帶的業務子帳號(system_generated=true,專供所屬業務)。
	sys := db.User.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetEmail("sys-acc@t.com").SetName("連鎖店(業務)").SetRole("customer").SetPasswordHash("x").
		SetAccountName("連鎖店(業務)").SetIsCustomer(true).SetCustomerID(cust.ID).
		SetIsPrimary(false).SetSystemGenerated(true).SetStatus(user.StatusActive).SaveX(ctx)

	uid := pri.ID
	if !asPrimary {
		uid = sys.ID // 以子帳號身分呼叫(驗子帳號無管理權限)
	}
	id := authz.Identity{
		UserID: strconv.Itoa(uid), CompanyID: strconv.Itoa(co.ID), DepartmentID: strconv.Itoa(dept.ID),
		CustomerID: strconv.Itoa(cust.ID), Role: "customer", Roles: []string{"customer"},
	}
	mux := http.NewServeMux()
	RegisterCustomerAccountService(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return &accountEnv{
		db: db, coID: co.ID, deptID: dept.ID, custID: cust.ID, priID: pri.ID, sysSub: sys.ID,
		client: customersv1connect.NewCustomerAccountServiceClient(http.DefaultClient, ts.URL),
	}
}

// identityWith 以自訂身分掛服務(驗非客戶身分/未登入的閘門)。
func (e *accountEnv) identityWith(t *testing.T, id authz.Identity) customersv1connect.CustomerAccountServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterCustomerAccountService(mux, e.db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, e.db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return customersv1connect.NewCustomerAccountServiceClient(http.DefaultClient, ts.URL)
}

// TestCustomerAccountSelfService 店家自助帳號管理(D22/規格 4.2,Task 6.7)主契約:
// 主帳號可列/新增/停用/重置**自建**子帳號;主帳號與業務子帳號皆不可管理(供 UI 灰化)。
func TestCustomerAccountSelfService(t *testing.T) {
	ctx := context.Background()
	e := newAccountEnv(t, true)

	// ① 清單:主帳號與業務子帳號都出現,且都不可管理。
	list, err := e.client.ListCustomerAccounts(ctx, connect.NewRequest(&customersv1.ListCustomerAccountsRequest{}))
	if err != nil {
		t.Fatalf("ListCustomerAccounts: %v", err)
	}
	byName := map[string]*customersv1.CustomerAccount{}
	for _, a := range list.Msg.GetAccounts() {
		byName[a.GetAccountName()] = a
	}
	if len(byName) != 2 {
		t.Fatalf("應見 2 個帳號(主+業務子),got %d", len(byName))
	}
	if a := byName["連鎖店"]; a == nil || !a.GetIsPrimary() || a.GetManageable() {
		t.Fatalf("主帳號應出現且不可管理(不可自停),got %+v", a)
	}
	if a := byName["連鎖店(業務)"]; a == nil || !a.GetSystemGenerated() || a.GetManageable() {
		t.Fatalf("業務子帳號應標系統附帶且不可管理,got %+v", a)
	}

	// ② 新增自建子帳號 → 回臨時密碼;DB 端為首登強改 + 24h 效期。
	created, err := e.client.CreateCustomerAccount(ctx, connect.NewRequest(&customersv1.CreateCustomerAccountRequest{
		AccountName: "主廚",
	}))
	if err != nil {
		t.Fatalf("CreateCustomerAccount: %v", err)
	}
	if created.Msg.GetTempPassword() == "" || created.Msg.GetTempExpiresAt() == "" {
		t.Fatal("新增子帳號應回臨時密碼與效期")
	}
	subID, _ := strconv.Atoi(created.Msg.GetAccount().GetId())
	if created.Msg.GetAccount().GetIsPrimary() || !created.Msg.GetAccount().GetManageable() {
		t.Fatalf("自建子帳號不應為主帳號且應可管理,got %+v", created.Msg.GetAccount())
	}
	sub := e.db.User.Query().Where(user.IDEQ(subID)).OnlyX(ctx)
	if !sub.MustChangePassword || sub.TempPasswordExpiresAt == nil {
		t.Fatal("新建子帳號應為 must_change_password 且有臨時效期")
	}
	if sub.CustomerID == nil || *sub.CustomerID != e.custID {
		t.Fatalf("子帳號應綁同客戶 %d,got %v", e.custID, sub.CustomerID)
	}

	// ③ 同名再建 → conflict(account_name 客戶內唯一,規格 4.2)。
	if _, err := e.client.CreateCustomerAccount(ctx, connect.NewRequest(&customersv1.CreateCustomerAccountRequest{
		AccountName: "主廚",
	})); connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("同名帳號應 conflict,got %v", err)
	}

	// ④ 重置自建子帳號 → 新臨時密碼 + 首登強改 + tv+1(在途 token 失效)。
	beforeTV := sub.TokenVersion
	reset, err := e.client.ResetCustomerAccountPassword(ctx, connect.NewRequest(&customersv1.ResetCustomerAccountPasswordRequest{
		AccountId: created.Msg.GetAccount().GetId(),
	}))
	if err != nil {
		t.Fatalf("ResetCustomerAccountPassword: %v", err)
	}
	if reset.Msg.GetTempPassword() == "" {
		t.Fatal("重置應回新臨時密碼")
	}
	afterReset := e.db.User.Query().Where(user.IDEQ(subID)).OnlyX(ctx)
	if !afterReset.MustChangePassword || afterReset.TokenVersion != beforeTV+1 {
		t.Fatalf("重置後應首登強改且 tv+1,got must_change=%v tv=%d", afterReset.MustChangePassword, afterReset.TokenVersion)
	}

	// ⑤ 業務子帳號不可停用、不可重置(店家無其密碼,專供所屬業務)。
	if _, err := e.client.DeactivateCustomerAccount(ctx, connect.NewRequest(&customersv1.DeactivateCustomerAccountRequest{
		AccountId: strconv.Itoa(e.sysSub),
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("業務子帳號不可停用,got %v", err)
	}
	if _, err := e.client.ResetCustomerAccountPassword(ctx, connect.NewRequest(&customersv1.ResetCustomerAccountPasswordRequest{
		AccountId: strconv.Itoa(e.sysSub),
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("業務子帳號不可重置,got %v", err)
	}

	// ⑥ 主帳號不可由店家停用(含自己,避免鎖死)。
	if _, err := e.client.DeactivateCustomerAccount(ctx, connect.NewRequest(&customersv1.DeactivateCustomerAccountRequest{
		AccountId: strconv.Itoa(e.priID),
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("主帳號不可自停,got %v", err)
	}

	// ⑦ 停用自建子帳號 → 成功且 tv 再 +1。
	if _, err := e.client.DeactivateCustomerAccount(ctx, connect.NewRequest(&customersv1.DeactivateCustomerAccountRequest{
		AccountId: created.Msg.GetAccount().GetId(),
	})); err != nil {
		t.Fatalf("停用自建子帳號: %v", err)
	}
	gone := e.db.User.Query().Where(user.IDEQ(subID)).OnlyX(ctx)
	if gone.Status != user.StatusInactive {
		t.Fatalf("子帳號應 inactive,got %s", gone.Status)
	}
	if gone.TokenVersion != beforeTV+2 {
		t.Fatalf("停用亦應 tv+1(重置後再+1),got %d", gone.TokenVersion)
	}
}

// TestCustomerAccountGatesNonPrimary 子帳號與員工皆無帳號管理權限(規格 4.2)。
func TestCustomerAccountGatesNonPrimary(t *testing.T) {
	ctx := context.Background()

	// ① 子帳號(以業務子帳號身分呼叫)→ permission_denied。
	e := newAccountEnv(t, false)
	if _, err := e.client.ListCustomerAccounts(ctx, connect.NewRequest(&customersv1.ListCustomerAccountsRequest{})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("子帳號不得呼叫帳號管理,got %v", err)
	}

	// ② 員工(dept_admin,展開角色含 customer)亦不得 —— 判身分必須用原始角色,不可用展開集。
	e2 := newAccountEnv(t, true)
	empClient := e2.identityWith(t, authz.Identity{
		UserID: "1", CompanyID: strconv.Itoa(e2.coID), DepartmentID: strconv.Itoa(e2.deptID),
		Role: "dept_admin", Roles: []string{"dept_admin", "staff", "customer"},
	})
	if _, err := empClient.ListCustomerAccounts(ctx, connect.NewRequest(&customersv1.ListCustomerAccountsRequest{})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("員工不得呼叫帳號管理(展開角色含 customer 不得誤放行),got %v", err)
	}

	// ③ 未登入 → unauthenticated。
	anonClient := e2.identityWith(t, authz.Identity{})
	if _, err := anonClient.ListCustomerAccounts(ctx, connect.NewRequest(&customersv1.ListCustomerAccountsRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("未登入應 unauthenticated,got %v", err)
	}
}

// TestCustomerAccountCrossCustomerIsolated 範圍限自己客戶:用別客戶的帳號 id 一律 not_found
// (不以 permission_denied 洩漏存在性)。
func TestCustomerAccountCrossCustomerIsolated(t *testing.T) {
	ctx := context.Background()
	e := newAccountEnv(t, true)
	other := e.db.Customer.Create().SetCompanyID(e.coID).SetDepartmentID(e.deptID).
		SetCustomerCode("ACC002").SetName("別店").SaveX(ctx)
	foreign := e.db.User.Create().SetCompanyID(e.coID).SetDepartmentID(e.deptID).
		SetEmail("foreign@t.com").SetName("別店子").SetRole("customer").SetPasswordHash("x").
		SetAccountName("別店子").SetIsCustomer(true).SetCustomerID(other.ID).
		SetIsPrimary(false).SetStatus(user.StatusActive).SaveX(ctx)

	for _, err := range []error{
		func() error {
			_, err := e.client.DeactivateCustomerAccount(ctx, connect.NewRequest(&customersv1.DeactivateCustomerAccountRequest{AccountId: strconv.Itoa(foreign.ID)}))
			return err
		}(),
		func() error {
			_, err := e.client.ResetCustomerAccountPassword(ctx, connect.NewRequest(&customersv1.ResetCustomerAccountPasswordRequest{AccountId: strconv.Itoa(foreign.ID)}))
			return err
		}(),
	} {
		if connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("跨客戶帳號應 not_found,got %v", err)
		}
	}
	got := e.db.User.Query().Where(user.IDEQ(foreign.ID)).OnlyX(ctx)
	if got.Status != user.StatusActive || got.TokenVersion != foreign.TokenVersion {
		t.Fatal("跨客戶操作不得改動對方帳號")
	}
}

// TestCustomerAccountAudited 帳號管理異動皆有稽核(規格 4.2)。
func TestCustomerAccountAudited(t *testing.T) {
	ctx := context.Background()
	e := newAccountEnv(t, true)
	before := e.db.AuditLog.Query().CountX(ctx)
	if _, err := e.client.CreateCustomerAccount(ctx, connect.NewRequest(&customersv1.CreateCustomerAccountRequest{AccountName: "二廚"})); err != nil {
		t.Fatalf("CreateCustomerAccount: %v", err)
	}
	if after := e.db.AuditLog.Query().CountX(ctx); after != before+1 {
		t.Fatalf("新增子帳號應寫 1 筆稽核(%d→%d)", before, after)
	}
	last := e.db.AuditLog.Query().Order(ent.Desc("id")).FirstX(ctx)
	if last.ResourceType != "user" || last.Action != "create" {
		t.Fatalf("稽核應為 user/create,got %s/%s", last.ResourceType, last.Action)
	}
	if last.AfterSnapshot["account_type"] != "customer_sub" {
		t.Fatalf("稽核應標 customer_sub,got %v", last.AfterSnapshot)
	}
}
