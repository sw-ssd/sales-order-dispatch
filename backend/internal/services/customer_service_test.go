package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
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
	RegisterCustomerServices(mux, db, "http://localhost:3000")
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

// seedCustomerRep 建立同公司同部門的有效業務(staff),回傳其 id。
func seedCustomerRep(t *testing.T, db *ent.Client, coID, deptID int) int {
	t.Helper()
	r, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).
		SetEmail("rep-" + t.Name() + "@t.com").SetName("業務").SetRole("staff").SetPasswordHash("x").Save(context.Background())
	if err != nil {
		t.Fatalf("seed rep: %v", err)
	}
	return r.ID
}

// deptAdminID 組裝 dept_admin 身分。
func deptAdminID(coID, deptID int) authz.Identity {
	return authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptID), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
}

// TestCreateCustomerGeneratesCodeAndCounter D7+D22:dept_admin 建立客戶(帶業務)→ 取號 TY000001、counter 推進、
// 連動建主/業務子帳號、稽核共 3 筆、回應含交付欄位。
func TestCreateCustomerGeneratesCodeAndCounter(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TY", true)
	repID := seedCustomerRep(t, db, coID, deptID)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))

	resp, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王小明", DefaultSalesRepId: uItoa(repID)}))
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
	// 稽核:客戶 + 兩帳號 = 3 筆。
	if n, _ := db.AuditLog.Query().Count(ctx); n != 3 {
		t.Fatalf("建檔應寫 3 筆稽核,得到 %d", n)
	}
	assertD22Accounts(t, db, ctx, resp, "王小明")
	// 第二個客戶取號連續。
	resp2, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "李四", DefaultSalesRepId: uItoa(repID)}))
	if err != nil {
		t.Fatalf("CreateCustomer2: %v", err)
	}
	if resp2.Msg.GetCustomer().GetCustomerCode() != "TY000002" {
		t.Fatalf("第二客戶編號應為 TY000002,得到 %q", resp2.Msg.GetCustomer().GetCustomerCode())
	}
}

// TestCreateCustomerWithoutPrefixFailedPrecondition:公司未設前綴(帶合法業務)→ failed_precondition,不建檔、不建帳號。
func TestCreateCustomerWithoutPrefixFailedPrecondition(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "", true)
	repID := seedCustomerRep(t, db, coID, deptID)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王", DefaultSalesRepId: uItoa(repID)})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("無前綴建檔應 failed_precondition,got %v", err)
	}
	if n, _ := db.Customer.Query().Count(ctx); n != 0 {
		t.Fatalf("無前綴時不應建檔,得到 %d 列", n)
	}
}

// TestCreateCustomerRequiresSalesRep:D22 交付必要——缺 default_sales_rep_id → invalid_argument,不建檔不建帳號。
func TestCreateCustomerRequiresSalesRep(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TZ", true)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺 default_sales_rep_id 應 invalid_argument,got %v", err)
	}
	if n, _ := db.Customer.Query().Count(ctx); n != 0 {
		t.Fatalf("缺業務時不應建檔,得到 %d 列", n)
	}
	if n, _ := db.User.Query().Count(ctx); n != 0 {
		t.Fatalf("缺業務時不應建帳號,得到 %d 列", n)
	}
}

// TestCreateCustomerSoftDeletedCompany:P2-A 後續——身分所指公司在建檔前被軟刪除 → not_found,
// 不建客戶列。修復前 Company.Get 不看 deleted_at,客戶會落入已刪除的租戶。
func TestCreateCustomerSoftDeletedCompany(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TS", true)
	repID := seedCustomerRep(t, db, coID, deptID)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))

	db.Company.UpdateOneID(coID).SetDeletedAt(time.Now().UTC()).SaveX(ctx)

	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王", DefaultSalesRepId: uItoa(repID)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪除公司下建客戶應 not_found,got %v", err)
	}
	if n, _ := db.Customer.Query().Count(ctx); n != 0 {
		t.Fatalf("已刪除公司下不得建客戶,得到 %d 列", n)
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

// TestCreateCustomerAccountFailureRollsBack:D22 建帳號任一環節失敗 → 客戶主檔與計數器皆回滾(D18 同交易)。
// 手法:預先占用即將生成的主帳號 email,使主帳號插入撞唯一值而失敗。
func TestCreateCustomerAccountFailureRollsBack(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TM", true)
	repID := seedCustomerRep(t, db, coID, deptID)
	// 佔位:第一個客戶將取號 TM000001,主帳號 email 固定為 customer.TM000001@system.local。
	if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).
		SetEmail("customer.TM000001@system.local").SetName("既有").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
		t.Fatalf("seed 佔位帳號: %v", err)
	}
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))

	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王", DefaultSalesRepId: uItoa(repID)})); err == nil {
		t.Fatal("主帳號 email 衝突時 CreateCustomer 應失敗")
	}
	// 客戶不應存在、計數器不得遞增、不得產生孤兒帳號。
	if n, _ := db.Customer.Query().Count(ctx); n != 0 {
		t.Fatalf("建帳失敗時不應有客戶列,得到 %d", n)
	}
	cnt, err := db.CustomerCounter.Query().Only(ctx)
	if err != nil || cnt.NextSeq != 1 {
		t.Fatalf("建帳失敗時計數器應回滾為 1,得到 %+v (err=%v)", cnt, err)
	}
	// 基準使用者數 = 業務 rep(1) + 佔位帳號(1) = 2;失敗後不得增加。
	if n, _ := db.User.Query().Count(ctx); n != 2 {
		t.Fatalf("建帳失敗時不應產生孤兒帳號,使用者數應仍為 2,得到 %d", n)
	}
}

// assertD22Accounts 驗證 D22 建檔連動:回應交付欄位 + DB 中主/業務子帳號正確。
func assertD22Accounts(t *testing.T, db *ent.Client, ctx context.Context, resp *connect.Response[customersv1.CreateCustomerResponse], customerName string) {
	t.Helper()
	if resp.Msg.GetPrimaryAccountName() != customerName {
		t.Fatalf("主帳號名稱應=客戶名稱 %q,得到 %q", customerName, resp.Msg.GetPrimaryAccountName())
	}
	if resp.Msg.GetSalesRepAccountName() != customerName+"(業務)" {
		t.Fatalf("業務子帳號名稱應=%s(業務),得到 %q", customerName, resp.Msg.GetSalesRepAccountName())
	}
	if resp.Msg.GetPrimaryTempPassword() == "" || resp.Msg.GetSalesRepTempPassword() == "" {
		t.Fatal("兩組臨時密碼應非空")
	}
	if resp.Msg.GetPrimaryTempPassword() == resp.Msg.GetSalesRepTempPassword() {
		t.Fatal("兩組臨時密碼應不同")
	}
	if got := resp.Msg.GetAccountManageUrl(); got != "http://localhost:3000/customer_account_manage" {
		t.Fatalf("account_manage_url 不正確,得到 %q", got)
	}

	c := resp.Msg.GetCustomer()
	uid, _ := strconv.Atoi(c.GetId())
	users := db.User.Query().Where(user.CustomerIDEQ(uid)).AllX(ctx)
	if len(users) != 2 {
		t.Fatalf("該客戶應恰有 2 個連動帳號,得到 %d", len(users))
	}
	var primary, sub *ent.User
	for _, u := range users {
		if u.IsPrimary {
			primary = u
		}
		if u.SystemGenerated {
			sub = u
		}
		if u.Role != "customer" || !u.IsCustomer || !u.MustChangePassword || u.TempPasswordExpiresAt == nil {
			t.Fatalf("連動帳號欄位錯誤: role=%s is_customer=%v must_change=%v exp=%v", u.Role, u.IsCustomer, u.MustChangePassword, u.TempPasswordExpiresAt)
		}
	}
	if primary == nil || sub == nil {
		t.Fatal("應存在 1 主帳號 + 1 業務子帳號")
	}
	if primary.SystemGenerated {
		t.Fatal("主帳號 system_generated 應為 false")
	}
	if sub.IsPrimary {
		t.Fatal("業務子帳號 is_primary 應為 false")
	}
	if !primary.TempPasswordExpiresAt.After(time.Now()) || primary.TempPasswordExpiresAt.After(time.Now().Add(25*time.Hour)) {
		t.Fatalf("主帳號臨時密碼應約 24h 效期,得到 %v", primary.TempPasswordExpiresAt)
	}
	// 核心契約:交付的臨時密碼明文須能對上儲存的雜湊(可登入)。
	if !auth.VerifyPassword(primary.PasswordHash, resp.Msg.GetPrimaryTempPassword()) {
		t.Fatal("主帳號回傳之臨時密碼無法對上其雜湊(交付憑證不可用)")
	}
	if !auth.VerifyPassword(sub.PasswordHash, resp.Msg.GetSalesRepTempPassword()) {
		t.Fatal("業務子帳號回傳之臨時密碼無法對上其雜湊(交付憑證不可用)")
	}
	// 兩帳號臨時密碼各自獨立(不可互相登入)。
	if auth.VerifyPassword(sub.PasswordHash, resp.Msg.GetPrimaryTempPassword()) {
		t.Fatal("主帳號臨時密碼不應能對上業務子帳號雜湊")
	}
}

// TestCreateCustomerWithValidDictRefAndRep:合法字典參考與業務 → 成功且欄位落庫。
func TestCreateCustomerWithValidDictRefAndRep(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TR", true)
	// 系統級 payment_method 字典。
	md, err := db.Metadict.Create().SetType("payment_method").SetCode("CASH").SetDisplayName("現金").SetIsActive(true).Save(ctx)
	if err != nil {
		t.Fatalf("metadict: %v", err)
	}
	// 同公司同部門業務。
	rep, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).SetEmail("rep@t.com").SetName("業務").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("rep: %v", err)
	}
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
	resp, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
		Name: "王小明", PaymentMethodId: uItoa(md.ID), DefaultSalesRepId: uItoa(rep.ID),
	}))
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}
	if resp.Msg.GetCustomer().GetPaymentMethodId() != uItoa(md.ID) || resp.Msg.GetCustomer().GetDefaultSalesRepId() != uItoa(rep.ID) {
		t.Fatalf("字典/業務參考應落庫,得到 %+v", resp.Msg.GetCustomer())
	}
	// 偏好送貨日預設為長度 6 全 false。
	if len(resp.Msg.GetCustomer().GetPreferredDeliveryDays()) != 6 {
		t.Fatalf("偏好送貨日預設應為長度 6,得到 %d", len(resp.Msg.GetCustomer().GetPreferredDeliveryDays()))
	}
}

// TestCreateCustomerInvalidDictRef:字典參考不存在 → invalid_argument,不建檔。
func TestCreateCustomerInvalidDictRef(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TV", true)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王", PaymentMethodId: "99999"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法字典參考應 invalid_argument,got %v", err)
	}
	if n, _ := db.Customer.Query().Count(ctx); n != 0 {
		t.Fatalf("驗證失敗不應建檔,得到 %d", n)
	}
}

// TestCreateCustomerInvalidRep:業務參考不存在 → invalid_argument。
func TestCreateCustomerInvalidRep(t *testing.T) {
	ctx := context.Background()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TW", true)
	client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
	if _, err := client.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "王", DefaultSalesRepId: "88888"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法業務參考應 invalid_argument,got %v", err)
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

// TestListCustomersSortWhitelist:客戶清單排序白名單的映射與預設(sort 空 → name 升冪)
// 與升/降冪方向(desc)。
//
// sqlite 層的回歸鎖:跨頁重複/遺漏是同值群次序不穩定的**執行計畫層級**缺陷,sqlite 上看不到
// (見 list_pagination_integration_test.go 的說明),故本測試只釘住「每個白名單值用到正確的
// 欄位」「sort 空 = name 且忽略 desc」與「desc 真的反轉」。fixture 三筆的 id / name(碼位) /
// customer_code / created_at 升冪序列兩兩互異 → 服務若忽略 sort、把某欄映射到別的欄位、
// 或忽略/誤用 desc,期望序列立即不符。
func TestListCustomersSortWhitelist(t *testing.T) {
	ctx := t.Context()
	_, db := newCustomerTestServer(t, authz.Identity{})
	coID, deptID := seedCustomerCompany(t, db, "TS", true)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// 插入序 甲→乙→丙(id 升冪 = 甲乙丙);name 碼位序 丙 < 乙 < 甲;sorted by code = 丙甲乙;
	// created_at 升冪 = 乙丙甲。
	for i, s := range []struct {
		name string
		code string
		day  int
	}{
		{"甲客戶", "B-01", 3},
		{"乙客戶", "C-01", 1},
		{"丙客戶", "A-01", 2},
	} {
		if _, err := db.Customer.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetName(s.name).SetCustomerCode(s.code).
			SetCreatedAt(base.AddDate(0, 0, s.day-1)).Save(ctx); err != nil {
			t.Fatalf("建立第 %d 筆客戶 fixture: %v", i, err)
		}
	}
	client, _ := newCustomerTestServer(t, authz.Identity{UserID: "1", CompanyID: uItoa(coID), Role: "super", Roles: []string{"super"}})

	for _, tc := range []struct {
		sort string
		desc bool
		want []string
	}{
		{"", false, []string{"丙客戶", "乙客戶", "甲客戶"}},  // 預設排序 == name 升冪
		{"", true, []string{"丙客戶", "乙客戶", "甲客戶"}},   // 契約:sort 空時忽略 desc
		{"  ", true, []string{"丙客戶", "乙客戶", "甲客戶"}}, // 全空白 sort == 空(sort 前後空白須 trim)
		{"name", false, []string{"丙客戶", "乙客戶", "甲客戶"}},
		{"name", true, []string{"甲客戶", "乙客戶", "丙客戶"}},
		{"customer_code", false, []string{"丙客戶", "甲客戶", "乙客戶"}},
		{"customer_code", true, []string{"乙客戶", "甲客戶", "丙客戶"}},
		{"created_at", false, []string{"乙客戶", "丙客戶", "甲客戶"}},
		{"created_at", true, []string{"甲客戶", "丙客戶", "乙客戶"}},
		{"  name  ", false, []string{"丙客戶", "乙客戶", "甲客戶"}}, // 前後空白可解析
		{"  name  ", true, []string{"甲客戶", "乙客戶", "丙客戶"}},
	} {
		res, err := client.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{Sort: tc.sort, Desc: tc.desc}))
		if err != nil {
			t.Errorf("sort=%q desc=%v: %v", tc.sort, tc.desc, err)
			continue
		}
		got := make([]string, 0, len(res.Msg.GetCustomers()))
		for _, c := range res.Msg.GetCustomers() {
			got = append(got, c.GetName())
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("sort=%q desc=%v:got %v,want %v", tc.sort, tc.desc, got, tc.want)
		}
	}

	// 空白包住的亂值仍在 trim 後落回白名單外 → 必須仍是 InvalidArgument(而非被當成預設排序)。
	for _, sort := range []string{"bogus", "  bogus  ", "name,desc"} {
		_, err := client.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{Sort: sort}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("非法 sort=%q 應 InvalidArgument,got %v", sort, err)
		}
		if msg := err.Error(); !strings.Contains(msg, "name/customer_code/created_at") {
			t.Errorf("錯誤訊息應逐字列出白名單(sort=%q),got %q", sort, msg)
		}
	}
}
