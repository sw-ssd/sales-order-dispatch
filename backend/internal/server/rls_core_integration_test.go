//go:build integration

package server

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// 本檔為「未登入／系統範圍路徑」在核心五表 ENABLE + FORCE 之後的探針,釘住兩個**全站級**失效模式:
//
//  1. **登入全滅**:核心表 ENABLE 後,登入查詢若跑在「無 scope 的請求交易」上會回 0 列 →
//     所有人得到「客戶編號或密碼錯誤」。同理 authzMiddleware 的 identityFor(讀 users/roles)
//     若沒包系統範圍,已登入者的身分解析全部失敗 → 每次請求 401 並銷毀 session。兩者都不是
//     例外而是靜默的空結果 —— sqlite 單元測試與 superuser 連線都看不到。
//  2. **開機佈建靜默歸零**:domains.go 以**已裝飾的業務 client** 呼叫 authz.Provision,讀
//     role_permissions/users/roles 在無 scope 下同樣回 0 列 → OpenFGA 什麼都沒授權而**不報錯**
//     (之後所有受保護 RPC 一致 deny)。比報錯更危險:沒有 log、沒有失敗。
//
// 因此本檔以 app_rw(非 superuser,PG 的 superuser 恆繞過 RLS)＋ dbtenant.NewClient 起一個
// **與生產同構**的 HTTP server(chi + sessions.LoadAndSave + authzMiddleware + 真 AuthService/
// UserService handler),跑完登入 → 帶 session 的已驗證請求 → refresh → 改密碼全鏈,並以
// 系統範圍交易直接對照「未設 scope 讀不到」。

// coreMigrationsDir 與 cmd/migrate 同路徑(go test 以套件目錄為 cwd)。
const coreMigrationsDir = "../../database/migrations"

// migrateCoreUp 以 cmd/migrate 相同路徑套用全部遷移(含 00024~00028 的 ENABLE)。
func migrateCoreUp(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	if err := goose.RunContext(t.Context(), "up", db, coreMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}

// coreOpenDB 開一條 database/sql 連線(admin/owner:superuser 不受 RLS 約束,供夾具與真值查詢)。
func coreOpenDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// coreAppRoleClient 建立 app_rw 的業務 client(必須經 dbtenant.NewClient,RLS 裝飾器才生效)。
//
// **連線池不設上限 1**:未登入路徑的 SystemScopeTx 一定在請求交易之外再開一條交易(第二條連線);
// 池若只有 1 條會與請求交易互鎖成死結。上限 4 = 明確的「≥2」,同時讓「漏收斂的存取」直接失敗
// 而不是把測試無聲掛住。
func coreAppRoleClient(t *testing.T, adminDSN string) *ent.Client {
	t.Helper()
	return coreAppRoleClientN(t, adminDSN, 4)
}

// coreAppRoleClientN 以指定連線池上限建立 app_rw 業務 client(供「連線數即斷言」的探針使用)。
func coreAppRoleClientN(t *testing.T, adminDSN string, maxConns int) *ent.Client {
	t.Helper()
	pool := coreOpenDB(t, testsupport.AppRoleDSN(t, adminDSN))
	pool.SetMaxOpenConns(maxConns)
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestIntegrationAuthLookupNeedsSystemScope 驗證核心表 ENABLE 之後的兩件事:
// ① 無 scope 的查詢看不到任何列(fail-closed)—— 這正是「登入查詢若沒包系統範圍就 0 列」的機制;
// ② SystemScopeTx 看得到(dbtenant.SystemScopeTx 是未登入路徑的唯一合法入口)。
func TestIntegrationAuthLookupNeedsSystemScope(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateCoreUp(t, dsn)
	client := coreAppRoleClient(t, dsn)
	ctx := t.Context()

	admin := coreOpenDB(t, dsn)
	co := coreInsertCompany(t, admin, "登入測試公司", "RLS-LOGIN")
	coreInsertUser(t, admin, co, "login@example.com", "staff", 0)

	visible, err := dbtenant.Client(ctx, client).User.Query().
		Where(user.EmailEQ("login@example.com")).Exist(ctx)
	if err != nil {
		t.Fatalf("無 scope 查詢: %v", err)
	}
	if visible {
		t.Fatal("無 scope 時 users 必須不可見(fail-closed)")
	}

	found := false
	if err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		var qerr error
		found, qerr = tx.Client().User.Query().Where(user.EmailEQ("login@example.com")).Exist(ctx)
		return qerr
	}); err != nil {
		t.Fatalf("系統範圍查詢: %v", err)
	}
	if !found {
		t.Fatal("系統範圍應看得到剛建立的使用者")
	}
}

// TestIntegrationLoginChainUnderAppRole 跑完整登入鏈(app_rw ＋ RLS 生效 ＋ 生產同構的 middleware):
// 登入 → 帶 session 的已驗證請求(走 authzMiddleware 的 identityFor) → 帶 Bearer 的已驗證請求
// (走 tokens.VerifyAccess ＋ identityFor) → refresh → 改密碼 → 舊 session/舊 refresh/舊 access 立即失效。
func TestIntegrationLoginChainUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateCoreUp(t, adminDSN)
	admin := coreOpenDB(t, adminDSN)
	client := coreAppRoleClient(t, adminDSN)

	coA := coreInsertCompany(t, admin, "A", "CORE-LOGIN-A")
	coB := coreInsertCompany(t, admin, "B", "CORE-LOGIN-B")
	deptA := coreInsertDepartment(t, admin, coA, "A 部門")
	deptOther := coreInsertDepartment(t, admin, coA, "A 他部門")
	deptB := coreInsertDepartment(t, admin, coB, "B 部門")
	deptAdminA := coreInsertUser(t, admin, coA, "core-login-deptadmin@example.com", "dept_admin", deptA)
	memberA := coreInsertUser(t, admin, coA, "core-login-member@example.com", "staff", deptA)
	memberOther := coreInsertUser(t, admin, coA, "core-login-other@example.com", "staff", deptOther)
	coreInsertUser(t, admin, coB, "core-login-b@example.com", "staff", deptB)
	customerCode := "CUST-LOGIN-1"
	customerID := coreInsertCustomer(t, admin, coA, deptA, customerCode, "core-login-cust@example.com", "OldPass123")

	kv := auth.NewMemoryStore()
	cfg := coreConfig()
	ts, s, sessions := newCoreHTTPServer(t, client, cfg, kv)
	authClient := salesorderv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL+"/api/v1")
	userClient := salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL+"/api/v1")

	// ① 登入:查詢必須走系統範圍(否則 0 列 → 「客戶編號或密碼錯誤」)。
	login := coreCall(t, authClient.Login, &v1.LoginRequest{CustomerCode: customerCode, Password: "OldPass123"})
	if login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatal("登入應核發 access/refresh token")
	}
	if !login.MustChangePassword {
		t.Fatal("臨時密碼帳號登入應回 must_change_password=true")
	}

	// ② 帶 session 的已驗證請求:identityFor 讀 users/companies/roles 必須走系統範圍。
	deptClient := salesorderv1connect.NewUserServiceClient(
		coreCookieHTTPClient(testSessionCookie(t, sessions, deptAdminA, "dept_admin")), ts.URL+"/api/v1")
	users := coreCall(t, deptClient.ListUsers, &v1.ListUsersRequest{}).Users
	if len(users) == 0 {
		t.Fatal("dept_admin 應看得到自己部門的成員(身分解析 + 部門範圍)")
	}
	for _, u := range users {
		if u.GetDepartmentId() != strconv.Itoa(deptA) {
			t.Fatalf("dept_admin 看到不在自己部門的成員:%+v", u)
		}
		if u.GetId() == strconv.Itoa(memberOther) {
			t.Fatal("dept_admin 看到同公司他部門的成員(範圍控制失效)")
		}
	}
	if got := coreCall(t, deptClient.GetUser, &v1.GetUserRequest{UserId: strconv.Itoa(memberA)}).User; got.GetId() != strconv.Itoa(memberA) {
		t.Fatalf("dept_admin 應讀得到自己部門的 staff:%+v", got)
	}
	if err := coreCallErr(t, deptClient.GetUser, &v1.GetUserRequest{UserId: strconv.Itoa(memberOther)}); err == nil {
		t.Fatal("dept_admin 不得讀取他部門成員")
	}

	// ③ 帶 Bearer 的已驗證請求:VerifyAccess(CurrentTokenVersion)＋ identityFor 都需系統範圍。
	access, err := s.tokens.IssueAccess(t.Context(), auth.TokenSubject{
		UserID: deptAdminA, CompanyID: coA, DepartmentID: deptA, Role: "dept_admin",
	})
	if err != nil {
		t.Fatalf("簽發 access token: %v", err)
	}
	bearerClient := salesorderv1connect.NewUserServiceClient(coreHeaderHTTPClient("Authorization", "Bearer "+access), ts.URL+"/api/v1")
	for _, u := range coreCall(t, bearerClient.ListUsers, &v1.ListUsersRequest{}).Users {
		if u.GetDepartmentId() != strconv.Itoa(deptA) {
			t.Fatalf("Bearer 路徑看到不在自己部門的成員:%+v", u)
		}
	}
	// 未帶憑證 → unauthenticated(身分缺席時不得放行)。
	if err := coreCallErr(t, userClient.ListUsers, &v1.ListUsersRequest{}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("未帶憑證的受保護 RPC 應 unauthenticated,got %v", err)
	}

	// ④ A3 受限態:must_change_password=true 時非 ChangePassword 一律 failed_precondition。
	// 這條同時是「身分已成功解析」的指紋 —— 身分解析失敗會是 unauthenticated,兩者可分辨。
	custClient := salesorderv1connect.NewUserServiceClient(
		coreCookieHTTPClient(testSessionCookie(t, sessions, customerID, "customer")), ts.URL+"/api/v1")
	err = coreCallErr(t, custClient.ListUsers, &v1.ListUsersRequest{})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("受限態應 failed_precondition(而非 unauthenticated),got %v", err)
	}

	// ⑤ refresh:Refresh → VerifyRefresh/CurrentTokenVersion 需系統範圍。
	refreshed := coreCall(t, authClient.Refresh, &v1.RefreshRequest{RefreshToken: login.RefreshToken})
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" || refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("refresh 應旋轉並核發新 token 對")
	}
	// 舊 refresh 已消耗 → 重放必拒。
	if err := coreCallErr(t, authClient.Refresh, &v1.RefreshRequest{RefreshToken: login.RefreshToken}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("已旋轉的 refresh token 重放應 unauthenticated,got %v", err)
	}

	// ⑥ 改密碼(session 身分):業務寫入 + 稽核同一請求交易,RLS 下必須成功。
	authAsCustomer := salesorderv1connect.NewAuthServiceClient(
		coreCookieHTTPClient(testSessionCookie(t, sessions, customerID, "customer")), ts.URL+"/api/v1")
	if err := coreCallErr(t, authAsCustomer.ChangePassword, &v1.ChangePasswordRequest{
		OldPassword: "OldPass123", NewPassword: "NewPass456",
	}); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	var hash string
	var mustChange bool
	var tv int
	if err := admin.QueryRow(
		`SELECT password_hash, must_change_password, token_version FROM users WHERE id = $1`, customerID).
		Scan(&hash, &mustChange, &tv); err != nil {
		t.Fatalf("讀使用者: %v", err)
	}
	if !auth.VerifyPassword(hash, "NewPass456") || mustChange || tv != 1 {
		t.Fatalf("改密碼後應為新 hash / must_change=false / tv=1,got verify=%v must_change=%v tv=%d",
			auth.VerifyPassword(hash, "NewPass456"), mustChange, tv)
	}
	if n := coreCount(t, admin,
		`SELECT count(*) FROM audit_logs WHERE resource_type = 'user' AND action = 'update' AND resource_id = $1`,
		strconv.Itoa(customerID)); n != 1 {
		t.Fatalf("改密碼應留下恰 1 列稽核(D18 同交易),got %d", n)
	}

	// ⑦ 改密碼已 bump token_version:舊 session(tv=0)、舊 refresh(tv=0)、舊 access(tv=0)一律失效。
	// 舊 session cookie 以新請求送出 → 身分解析發現 tv 不符 → 銷毀 session 並視為未登入。
	if err := coreCallErr(t, custClient.ListUsers, &v1.ListUsersRequest{}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("tv 已變更的舊 session 應 unauthenticated,got %v", err)
	}
	if err := coreCallErr(t, authClient.Refresh, &v1.RefreshRequest{RefreshToken: refreshed.RefreshToken}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("tv 已變更的 refresh token 應 unauthenticated,got %v", err)
	}
	// 舊 access token(登入時核發,tv=0)亦以 tv 為憑 → 改密碼後不得再通過(D5 立即失效)。
	authAsCustomerBearer := salesorderv1connect.NewAuthServiceClient(
		coreHeaderHTTPClient("Authorization", "Bearer "+login.AccessToken), ts.URL+"/api/v1")
	if err := coreCallErr(t, authAsCustomerBearer.ChangePassword, &v1.ChangePasswordRequest{
		OldPassword: "NewPass456", NewPassword: "ThirdPass789",
	}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("tv 已變更的舊 access token 應 unauthenticated,got %v", err)
	}
}

// TestIntegrationRegisterGuestUnderAppRole guest 註冊的兩條寫入路徑在 RLS 生效下以 app_rw 走過:
//
//	① registration token 建 guest(registerWithToken:去重 + Create 同一條系統交易);
//	② 既有 guest 完成註冊(completeGuest:資料更新與 token_version+1 併成**同一個敘述**)。
//
// **連線池上限 2 + 每 RPC deadline 是本測試的第二個斷言**:這兩條路徑合法地需要 2 條連線
// (interceptor 的請求交易 + 未登入路徑的系統範圍交易),池設 2 時正確的實作會完成;若把
// token_version bump 拆成第三條交易(例如改回 Tokens.BumpTokenVersion,而外層系統交易已 UPDATE
// 同一列),PG 上會互鎖死結 → 這裡以逾時紅。上限刻意**不是 1**:1 條會讓合法的系統範圍交易也逾時,
// 分不出對錯(brief 的池限制即為此)。
func TestIntegrationRegisterGuestUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateCoreUp(t, adminDSN)
	admin := coreOpenDB(t, adminDSN)

	coA := coreInsertCompany(t, admin, "A", "CORE-GUEST-A")
	coB := coreInsertCompany(t, admin, "B", "CORE-GUEST-B")
	// 既有 guest(路徑 ②):status=active、role=guest、已歸屬 coA、tv=0。
	guestID := coreInsertUser(t, admin, coA, "core-guest-existing@example.com", "guest", 0)

	client := coreAppRoleClientN(t, adminDSN, 2)
	kv := auth.NewMemoryStore()
	ts, _, sessions := newCoreHTTPServer(t, client, coreConfig(), kv)

	// ① registration token 路徑:OneTime store 預先放入 token(email)→ RegisterComplete 建 guest。
	const regToken = "t9-registration-token"
	newGuestEmail := "core-guest-new@example.com"
	if err := kv.Set(t.Context(), auth.RegistrationKey(regToken), newGuestEmail, auth.RegistrationTokenTTL); err != nil {
		t.Fatalf("預置 registration token: %v", err)
	}
	tokenClient := salesorderv1connect.NewAuthServiceClient(
		coreHeaderHTTPClient("X-Registration-Token", regToken), ts.URL+"/api/v1")
	if err := coreCallDeadline(t, 20*time.Second, tokenClient.RegisterComplete, &v1.RegisterCompleteRequest{
		Name: "新 guest", CompanyId: strconv.Itoa(coB),
	}); err != nil {
		t.Fatalf("RegisterComplete(registration token): %v", err)
	}
	var role, status, name string
	var companyID int
	if err := admin.QueryRow(
		`SELECT role, status, name, company_users FROM users WHERE email = $1`, newGuestEmail).
		Scan(&role, &status, &name, &companyID); err != nil {
		t.Fatalf("registration token 路徑應建立 guest 帳號: %v", err)
	}
	if role != "guest" || status != "pending" || name != "新 guest" || companyID != coB {
		t.Fatalf("新建 guest 應為 role=guest/status=pending/name=新 guest/company=%d,got role=%q status=%q name=%q company=%d",
			coB, role, status, name, companyID)
	}

	// ② 既有 guest 路徑:以 session 身分完成註冊 → 更新 + token_version+1(同一敘述/同一交易)。
	guestClient := salesorderv1connect.NewAuthServiceClient(
		coreCookieHTTPClient(testSessionCookie(t, sessions, guestID, "guest")), ts.URL+"/api/v1")
	if err := coreCallDeadline(t, 20*time.Second, guestClient.RegisterComplete, &v1.RegisterCompleteRequest{
		Name: "改名後", CompanyId: strconv.Itoa(coB),
	}); err != nil {
		t.Fatalf("RegisterComplete(session guest): %v", err)
	}
	var tv int
	if err := admin.QueryRow(
		`SELECT role, status, name, company_users, token_version FROM users WHERE id = $1`, guestID).
		Scan(&role, &status, &name, &companyID, &tv); err != nil {
		t.Fatalf("讀 guest: %v", err)
	}
	if name != "改名後" || status != "pending" || companyID != coB {
		t.Fatalf("完成註冊後應 name=改名後/status=pending/company=%d,got name=%q status=%q company=%d",
			coB, name, status, companyID)
	}
	if tv != 1 {
		t.Fatalf("完成註冊須在同一交易內 bump token_version(舊憑證失效),got tv=%d", tv)
	}
}

// TestIntegrationProvisionUnderAppRole 釘住「開機佈建靜默歸零」:authz.Provision 讀
// role_permissions／users／roles 時必須跑在系統範圍,否則在核心表 ENABLE 後回 0 列 →
// OpenFGA 一個 tuple 都不寫而**不報錯**(之後所有受保護 RPC 一致 deny)。
//
// 斷言方式是「有沒有 tuple」而非「有沒有錯誤」:沉默的 0 筆正是本缺陷的形狀。
func TestIntegrationProvisionUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateCoreUp(t, dsn)
	client := coreAppRoleClient(t, dsn)
	ctx := t.Context()

	admin := coreOpenDB(t, dsn)
	co := coreInsertCompany(t, admin, "佈建公司", "RLS-PROV")
	roleID := coreInsertRole(t, admin, "core_prov_role")
	coreInsertRolePermission(t, admin, roleID, "customer", "read")
	coreInsertRolePermission(t, admin, roleID, "customer", "update")
	uid := coreInsertUser(t, admin, co, "core-prov@example.com", "core_prov_role", 0)

	fgaClient, err := ofga.NewMemory(ctx, "t9-provision")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(fgaClient.Close)
	engine := authzopenfga.New(fgaClient)

	if err := authz.Provision(ctx, engine, client); err != nil {
		t.Fatalf("Provision 不得因 RLS 而報錯(它是靜默歸零,不是失敗): %v", err)
	}

	userset := "role:" + strconv.Itoa(roleID) + "#assigned"
	if ok, err := engine.Check(ctx, userset, "can_read", "ability:customer"); err != nil || !ok {
		t.Fatalf("role→ability(can_read/customer) 應被佈建,got ok=%v err=%v", ok, err)
	}
	if ok, err := engine.Check(ctx, userset, "can_write", "ability:customer"); err != nil || !ok {
		t.Fatalf("role→ability(can_write/customer) 應被佈建,got ok=%v err=%v", ok, err)
	}
	if ok, err := engine.Check(ctx, "user:"+strconv.Itoa(uid), "assigned", "role:"+strconv.Itoa(roleID)); err != nil || !ok {
		t.Fatalf("user→role assigned 應被佈建,got ok=%v err=%v", ok, err)
	}
	if tuples, err := engine.ListTuples(ctx); err != nil || len(tuples) < 3 {
		t.Fatalf("佈建後應有 ≥3 個 tuple(0 筆即為靜默歸零),got %d err=%v", len(tuples), err)
	}
}

// coreConfig 建立測試用 config:OpenFGA 關閉 → authorizeRPC 走「刻意停用 → 回退放行」語意,
// 故本檔量到的是「身分是否解析成功」而不是 OpenFGA 授權決策(授權由各服務層 ACL 承擔)。
func coreConfig() *config.Config {
	return &config.Config{
		API:     config.API{Env: "test", DeveloperAccountEnabled: false},
		Auth:    config.Auth{JWTSecret: "t9-core-probe-secret", SessionLifetime: 24 * time.Hour, FrontendURL: "http://localhost:3000"},
		OpenFGA: config.OpenFGA{Enabled: false},
	}
}

// newCoreHTTPServer 以業務 client 建出**與生產同構**的 HTTP server:
// chi + sessions.LoadAndSave + s.authzMiddleware(真 server.go 的身分解析)＋ 真 AuthService／
// UserService(經 dbtenant.HandlerOption 取得請求交易)。刻意不呼叫 InitDomains:它需要 Valkey
// (Redis)與 OIDC discovery,測試改以 scs memstore + 記憶體 KV,其餘組裝與 domains.go 一致。
func newCoreHTTPServer(t *testing.T, client *ent.Client, cfg *config.Config, kv auth.KVStore) (*httptest.Server, *Server, *scs.SessionManager) {
	t.Helper()
	sessions := auth.WebSessionManager(memstore.New(), cfg.Auth.SessionLifetime, false, "lax")
	s := &Server{cfg: cfg, router: chi.NewRouter()}
	s.tokens = auth.NewTokenManager(cfg.Auth.JWTSecret, kv, client)

	authHandler := handlers.NewAuthHandler(handlers.AuthDeps{
		Cfg:      cfg,
		DB:       client,
		Tokens:   s.tokens,
		Lockout:  auth.NewLoginLock(kv),
		OneTime:  auth.NewOneTimeStore(kv),
		Sessions: sessions,
		// 席位守衛（本檔驗 RLS 與身分路徑，不驗配額）：Unlimited 等同不受配額限制，
		// 與注入前語意相同（nil 是 fail-closed，會擋住 OIDC／註冊的建帳號）。
		Entitlements: entitlements.Unlimited(),
	})
	apiMux := http.NewServeMux()
	authPath, authConnectHandler := salesorderv1connect.NewAuthServiceHandler(authHandler, dbtenant.HandlerOption(client))
	apiMux.Handle(authPath, authConnectHandler)
	services.RegisterUserServices(apiMux, client, entitlements.Unlimited())
	s.router.Mount("/api/v1", http.StripPrefix("/api/v1", sessions.LoadAndSave(s.authzMiddleware(client, sessions, apiMux))))

	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts, s, sessions
}

// coreCookieHTTPClient 回傳每個請求都帶上指定 cookie 的 http.Client
// (scs session 走 cookie,connect 的 CallOption 帶不了它,故以 transport 附加)。
func coreCookieHTTPClient(cookie *http.Cookie) *http.Client {
	return &http.Client{Transport: &coreCookieTransport{cookie: cookie}}
}

// coreHeaderHTTPClient 回傳每個請求都帶上指定標頭的 http.Client(Bearer／registration token 共用)。
func coreHeaderHTTPClient(name, value string) *http.Client {
	return &http.Client{Transport: &coreHeaderTransport{name: name, value: value}}
}

type coreCookieTransport struct{ cookie *http.Cookie }

func (t *coreCookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.AddCookie(t.cookie)
	return http.DefaultTransport.RoundTrip(clone)
}

type coreHeaderTransport struct{ name, value string }

func (t *coreHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set(t.name, t.value)
	return http.DefaultTransport.RoundTrip(clone)
}

// coreCall 執行單一 RPC 並回傳回應訊息(錯誤即測試失敗)。
func coreCall[M, T any](t *testing.T, call func(context.Context, *connect.Request[M]) (*connect.Response[T], error), msg *M) *T {
	t.Helper()
	resp, err := call(t.Context(), connect.NewRequest(msg))
	if err != nil {
		t.Fatalf("RPC 失敗: %v", err)
	}
	return resp.Msg
}

// coreCallDeadline 以 ctx deadline 執行單一 RPC:供「連線池上限受限」的探針把「卡在第三條交易」
// 收斂成可讀的逾時(逾時本身就是死結的證據)。
func coreCallDeadline[M, T any](t *testing.T, d time.Duration, call func(context.Context, *connect.Request[M]) (*connect.Response[T], error), msg *M) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), d)
	defer cancel()
	_, err := call(ctx, connect.NewRequest(msg))
	return err
}

// coreCallErr 執行單一 RPC,錯誤交由呼叫端斷言。
func coreCallErr[M, T any](t *testing.T, call func(context.Context, *connect.Request[M]) (*connect.Response[T], error), msg *M) error {
	t.Helper()
	_, err := call(t.Context(), connect.NewRequest(msg))
	return err
}

// coreInsertCompany 以 admin 連線建一間公司(companies 無 created_at/updated_at)。
func coreInsertCompany(t *testing.T, db *sql.DB, name, identifier string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ($1, $2, 'active') RETURNING id`,
		name, identifier).Scan(&id); err != nil {
		t.Fatalf("建公司 %s: %v", name, err)
	}
	return id
}

// coreInsertDepartment 以 admin 連線建一間部門。
func coreInsertDepartment(t *testing.T, db *sql.DB, companyID int, name string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO departments (name, company_departments) VALUES ($1, $2) RETURNING id`, name, companyID).Scan(&id); err != nil {
		t.Fatalf("建部門 %s: %v", name, err)
	}
	return id
}

// coreInsertUser 以 admin 連線建一般(非客戶)帳號;departmentID=0 表不掛部門。
func coreInsertUser(t *testing.T, db *sql.DB, companyID int, email, role string, departmentID int) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, password_hash, company_users, department_users)
		 VALUES ($1, '公司管理員', $2, 'x', $3, NULLIF($4, 0)::bigint) RETURNING id`,
		email, role, companyID, departmentID).Scan(&id); err != nil {
		t.Fatalf("建使用者 %s: %v", email, err)
	}
	return id
}

// coreInsertCustomer 以 admin 連線建一個「臨時密碼待改」的客戶帳號(account_name = 客戶編號)。
func coreInsertCustomer(t *testing.T, db *sql.DB, companyID, departmentID int, accountName, email, password string) int {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("雜湊密碼: %v", err)
	}
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, is_customer, account_name, password_hash,
		                    company_users, department_users, must_change_password, temp_password_expires_at)
		 VALUES ($1, '客戶', 'customer', true, $2, $3, $4, $5, true, now() + interval '1 hour') RETURNING id`,
		email, accountName, hash, companyID, departmentID).Scan(&id); err != nil {
		t.Fatalf("建客戶帳號 %s: %v", email, err)
	}
	return id
}

// coreInsertRole 以 admin 連線建一列角色。
func coreInsertRole(t *testing.T, db *sql.DB, code string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO roles (code, name, data_scope, is_system, is_active)
		 VALUES ($1, '佈建探針角色', 'company', false, true) RETURNING id`, code).Scan(&id); err != nil {
		t.Fatalf("建角色 %s: %v", code, err)
	}
	return id
}

// coreInsertRolePermission 以 admin 連線建一筆角色權限(供 Provision 轉 tuple)。
func coreInsertRolePermission(t *testing.T, db *sql.DB, roleID int, resource, action string) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO role_permissions (role_id, resource, action, sort_order) VALUES ($1, $2, $3, 0)`,
		roleID, resource, action); err != nil {
		t.Fatalf("建角色權限 %s/%s: %v", resource, action, err)
	}
}

// coreCount 以 admin 連線取單一純量真值。
func coreCount(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("查詢 %q: %v", query, err)
	}
	return n
}
