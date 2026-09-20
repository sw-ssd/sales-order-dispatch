//go:build integration

// 跨租戶統一守門探針(T10):以 A 公司的身分呼叫租戶端點,斷言
//
//	①看不到 B 公司的任何列(清單集合 == 該 scope 的真值;單筆取用 B 的 id → not_found);
//	②改不到 B 公司的任何列(寫入動詞用 B 的 id 或 B 的父列 → not_found,且目標列值不變、不留稽核列)。
//
// **覆蓋範圍(精確)**:
//   - 清單:14 張租戶表的 List(companies/departments/users/roles/customers/customer_addresses/
//     customer_contacts/warehouses/routes/processing_specs/product_categories/products/metadicts/
//     audit_logs)＋ metadicts 的 ListOptions(下拉選項)。
//   - 單筆讀取:companies/departments/users/customers/products/metadicts 的 Get,以及
//     customer_addresses/customer_contacts 的父列解析(ListAddresses／ListContacts 帶 B 的客戶 id)。
//   - 寫入動詞:Update 與 Delete 各域(companies/departments/users/customers/customer_addresses/
//     customer_contacts/warehouses/routes/processing_specs/product_categories/products/metadicts);
//     Restore(customers/products/warehouses/routes/processing_specs/product_categories);
//     AddAddress／AddContact(以 B 的客戶為父列);AssignRole／ForceLogout(以 B 的使用者為目標);
//     CreateDepartment／CreateUser(帶 B 的公司 id)、CreateProduct(帶 B 的商品分類)。
//
// **不在此探針、且本身沒有跨租戶語意的端點**(不用「每個端點」含混帶過):
//   - `GetAbility`(ability.v1)、`ListConditionFields`(role.v1):回的是呼叫者自身的能力/CASL
//     欄位集,輸入裡沒有任何租戶目標,不存在「B 的資料」可供洩漏。
//   - `ListRoles`／`GetRolePermissions`／`UpdateRolePermissions`:roles／role_permissions 是**共用
//     目錄**(00025 的 policy 只要求 `scope` 非空,寫入權威在服務層 ACL)。本探針只驗「共用目錄在
//     非空 scope 下讀得到、不因 RLS 黑屏」;跨公司的寫入防線由 role_service 的 requireRole／條件
//     驗證負責(`role_service_test.go`)。
//   - `Login`／`Refresh`／`Logout`／`RegisterComplete`／`QRLogin`／`ChangePassword`:未登入／系統
//     範圍路徑(尚無身分可注入租戶 scope),由 T9 的 `internal/server` 探針與 handlers 的測試涵蓋;
//     `ResetCustomerPassword` 由 T8 的 `internal/handlers/rls_auth_password_integration_test.go` 涵蓋。
//   - `CreateCustomer`／`CreateMetadict`:`CreateCustomer` 的公司來自身分(不吃請求的公司 id),它
//     唯一的跨租戶向量是 `default_sales_rep_id` 指向 B 的使用者 —— 與 `CreateProduct` 的參照同型,
//     由本探針的 CreateProduct 代表;`CreateMetadict` 對 super 一律建系統級列(department 為 NULL),
//     沒有跨租戶向量。
//
// 為何需要這一支(其餘六組探針 T5–T9 已各自驗過所屬域):那些探針守的是「本域的路徑有沒有收斂」,
// 這支守的是**跨域的統一接線** —— 「某個端點忘了掛 dbtenant.HandlerOption / dbtenant.Client」。
// 它在單元測試(sqlite,enttest)完全看不出來:沒有 pg 的 RLS,漏掛的查詢照樣讀得到全部列。
// 這裡的每個正向斷言(「A 必須看到自家那幾筆」)與每個負向斷言(「B 的 id 一律 not_found」)
// 都只有在 RLS 這條接線存在時才成立:
//   - 漏掛租戶交易 → SET LOCAL 沒被執行 → 查詢被 policy 過濾成 0 列(正向斷言紅);
//   - 漏掛 dbtenant.Client → 同上,或跨租戶目標真的被讀出/改到(負向斷言紅)。
//
// 身分與 scope 刻意**錯開**:身分用 super(ACL 全開,requireScope/scopeForTarget 一律放行),
// scope 用 A 公司的部門層級。因此 ACL 不會代為擋下任何事,紅的一定是 RLS 這條接線。
// 部門層級(而非公司層級)是為了讓 metadicts 這張以部門為租戶鍵的表也測得動
// (其 policy 對部門擴充列要求 current_department_id)。
//
// 為何必須 app_rw:測試容器的 admin 是 superuser,而 PG 的 superuser 永遠繞過 RLS(FORCE 亦然)
// → 用 admin 建的 server 不管路徑有沒有漏掛都會全綠(T5 實測)。
package services

import (
	"context"
	"database/sql"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	auditv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	metadictv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// crossTenantFixture 是兩家公司的等價資料:每個租戶端點都必須能在 A 看到自家那一筆、
// 且對 B 的那一筆完全無感。
type crossTenantFixture struct {
	coA, coB             int
	deptA, deptB         int
	actorA, actorB       int
	custA, custB         int
	addrA, addrB         int
	contactA, contactB   int
	productA, productB   int
	metadictA, metadictB int
	auditA, auditB       int
	// masters 是四張部門級主檔的 A/B 列 id(table → 兩家公司的 id)。
	masters map[string][2]string
}

// TestIntegrationRLSCrossTenantEndpoints 是 T10 的統一守門:逐端點驗「A 看不到/改不到 B」。
//
// 斷言一律是 `not_found`(不是 permission_denied):RLS 在讀取階段就把 B 的列濾掉,
// 服務層的 ACL 根本沒機會看到目標 → 回應不得洩漏「該資源存在」。
func TestIntegrationRLSCrossTenantEndpoints(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	fx := seedCrossTenant(t, admin)

	// 受測 server **必須**以 app_rw(app_rw 是 00022 的 NOBYPASSRLS 業務角色)＋
	// dbtenant.NewClient 建立:SET LOCAL 是 driver 裝飾器在 Tx(ctx) 內套的。
	// 連線池上限 4(不是 1):請求交易之外的 SystemScopeTx 需要第二條連線,池 1 會互鎖死結
	// (T9 的 TestIntegrationRegisterGuestUnderAppRole 前例);4 同時讓漏收斂的查詢不至於吃滿池。
	appSQL := openAppRoleDB(t, adminDSN)
	appSQL.SetMaxOpenConns(4)
	appClient := openScopedClient(t, appSQL)
	cl := newListScanServerWithScope(t, appClient, auth.RLSScope{
		UserID: itoa(fx.actorA), CompanyID: itoa(fx.coA), DepartmentID: itoa(fx.deptA),
		DataScope: auth.DataScopeDepartment, CompanyActive: true,
	})
	ctx := t.Context()

	// 正向對照(唯一一次成功的寫入):A 改自己的客戶必須成功並落稽核。沒有這一段,
	// 下面那一堆 not_found 可能只是「整個端點壞掉」的假證據。
	t.Run("正向對照:A 改自己的客戶成功", func(t *testing.T) {
		name := "A 客戶(改)"
		got := crossCall(t, "UpdateCustomer(自家)", func(ctx context.Context) (*connect.Response[customersv1.UpdateCustomerResponse], error) {
			return cl.customers.UpdateCustomer(ctx, connect.NewRequest(&customersv1.UpdateCustomerRequest{
				Id: itoa(fx.custA), Name: &name,
			}))
		})
		if got.Msg.GetCustomer().GetName() != name {
			t.Fatalf("A 應改得動自己的客戶,got %+v", got.Msg.GetCustomer())
		}
		assertColumnValue(t, admin, "A 的客戶名稱", "customers", "name", fx.custA, name)
		assertAuditRow(t, admin, "customer", "update", itoa(fx.custA), fx.coA)
	})

	t.Run("companies", func(t *testing.T) {
		res := crossCall(t, "ListCompanies", func(ctx context.Context) (*connect.Response[v1.ListCompaniesResponse], error) {
			return cl.companies.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{Page: 1, PageSize: 50}))
		})
		ids := make([]string, 0, len(res.Msg.GetCompanies()))
		for _, c := range res.Msg.GetCompanies() {
			ids = append(ids, c.GetId())
		}
		// 公司清單在服務層**沒有**依身分過濾(僅 RLS)→ 這條是純 RLS 偵測器。
		assertSetEquals(t, "ListCompanies", ids, adminIDs(t, admin, `SELECT id FROM companies WHERE id = $1`, fx.coA)...)

		crossCallErr(t, "GetCompany(他公司)", func(ctx context.Context) error {
			_, err := cl.companies.GetCompany(ctx, connect.NewRequest(&v1.GetCompanyRequest{CompanyId: itoa(fx.coB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateCompany(他公司)", func() error {
			n := "改到 B 公司"
			_, err := cl.companies.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{
				CompanyId: itoa(fx.coB), Name: &n,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteCompany(他公司)", func() error {
			_, err := cl.companies.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: itoa(fx.coB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的公司名稱", "companies", "name", fx.coB, "跨租戶 B")
		assertNullColumn(t, admin, "B 的公司未被軟刪除", "companies", "deleted_at", fx.coB)
	})

	t.Run("departments", func(t *testing.T) {
		res := crossCall(t, "ListDepartments", func(ctx context.Context) (*connect.Response[v1.ListDepartmentsResponse], error) {
			return cl.departments.ListDepartments(ctx, connect.NewRequest(&v1.ListDepartmentsRequest{Page: 1, PageSize: 50}))
		})
		ids := make([]string, 0, len(res.Msg.GetDepartments()))
		for _, d := range res.Msg.GetDepartments() {
			ids = append(ids, d.GetId())
		}
		assertSetEquals(t, "ListDepartments", ids, itoa(fx.deptA))

		crossCallErr(t, "GetDepartment(他公司)", func(ctx context.Context) error {
			_, err := cl.departments.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: itoa(fx.deptB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateDepartment(他公司)", func() error {
			n := "改到 B 部門"
			_, err := cl.departments.UpdateDepartment(ctx, connect.NewRequest(&v1.UpdateDepartmentRequest{
				DepartmentId: itoa(fx.deptB), Name: &n,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteDepartment(他公司)", func() error {
			_, err := cl.departments.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: itoa(fx.deptB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的部門名稱", "departments", "name", fx.deptB, "跨租戶 B 部門")
		assertNullColumn(t, admin, "B 的部門未被軟刪除", "departments", "deleted_at", fx.deptB)
	})

	t.Run("users", func(t *testing.T) {
		// 不帶 company_id:super 身分在服務層沒有範圍過濾 → 可見集合完全由 RLS 決定。
		res := crossCall(t, "ListUsers", func(ctx context.Context) (*connect.Response[v1.ListUsersResponse], error) {
			return cl.users.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{Page: 1, PageSize: 50}))
		})
		ids := make([]string, 0, len(res.Msg.GetUsers()))
		for _, u := range res.Msg.GetUsers() {
			ids = append(ids, u.GetId())
		}
		assertSetEquals(t, "ListUsers", ids, adminIDs(t, admin, `SELECT id FROM users WHERE department_users = $1`, fx.deptA)...)

		crossCallErr(t, "GetUser(他公司)", func(ctx context.Context) error {
			_, err := cl.users.GetUser(ctx, connect.NewRequest(&v1.GetUserRequest{UserId: itoa(fx.actorB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateUser(他公司)", func() error {
			n := "改到 B 使用者"
			_, err := cl.users.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{
				UserId: itoa(fx.actorB), Name: &n,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "Deactivate(他公司)", func() error {
			_, err := cl.users.Deactivate(ctx, connect.NewRequest(&v1.DeactivateRequest{UserId: itoa(fx.actorB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的使用者名稱", "users", "name", fx.actorB, "跨租戶 B 管理員")
		assertColumnValue(t, admin, "B 的使用者狀態", "users", "status", fx.actorB, "active")
		// 新帳號的 token_version 起始為 0(欄位無 DB 預設);Deactivate 若得逞會 +1(D5)
		// → 這個斷言是「停用沒有生效」的直接證據。
		assertColumnValue(t, admin, "B 的 token_version", "users", "token_version", fx.actorB, "0")
	})

	t.Run("roles", func(t *testing.T) {
		// roles／role_permissions 是**共用目錄**,不是租戶表:00025 的 policy 只要求「scope 非空」
		// (寫入權威在服務層 ACL)。因此「看不到 B 的角色」在語意上不成立 —— 這條端點要驗的是
		// 「共用目錄在任何非空 scope 下讀得到,不會因 RLS 而黑屏」,跨公司的寫入防線由
		// role_service 的 requireRole／條件驗證負責(見 role_service_test.go）。
		sharedRoleID := insertRLSCoreRole(t, admin, "XT-SHARED")
		res := crossCall(t, "ListRoles", func(ctx context.Context) (*connect.Response[v1.ListRolesResponse], error) {
			return cl.roles.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{Page: 1, PageSize: 50}))
		})
		if !containsRoleCode(res.Msg.GetRoles(), "XT-SHARED") {
			t.Fatalf("共用角色目錄應在任何非空 scope 下讀得到(XT-SHARED 缺席)")
		}
		perms := crossCall(t, "GetRolePermissions", func(ctx context.Context) (*connect.Response[v1.GetRolePermissionsResponse], error) {
			return cl.roles.GetRolePermissions(ctx, connect.NewRequest(&v1.GetRolePermissionsRequest{RoleId: itoa(sharedRoleID)}))
		})
		if n := len(perms.Msg.GetPermissions()); n != 0 {
			t.Fatalf("自訂角色尚無權限,應回 0 筆,got %d", n)
		}
	})

	t.Run("customers", func(t *testing.T) {
		res := crossCall(t, "ListCustomers", func(ctx context.Context) (*connect.Response[customersv1.ListCustomersResponse], error) {
			return cl.customers.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{Page: 1, PageSize: 50}))
		})
		ids := make([]string, 0, len(res.Msg.GetCustomers()))
		for _, c := range res.Msg.GetCustomers() {
			ids = append(ids, c.GetId())
		}
		assertSetEquals(t, "ListCustomers", ids, tenantScopedIDs(t, admin, "customers", fx.coA, fx.deptA)...)

		crossCallErr(t, "GetCustomer(他公司)", func(ctx context.Context) error {
			_, err := cl.customers.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: itoa(fx.custB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateCustomer(他公司)", func() error {
			n := "改到 B 客戶"
			_, err := cl.customers.UpdateCustomer(ctx, connect.NewRequest(&customersv1.UpdateCustomerRequest{
				Id: itoa(fx.custB), Name: &n,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteCustomer(他公司)", func() error {
			_, err := cl.customers.DeleteCustomer(ctx, connect.NewRequest(&customersv1.DeleteCustomerRequest{Id: itoa(fx.custB)}))
			return err
		})
		assertCrossTenantDenied(t, admin, "RestoreCustomer(他公司)", func() error {
			_, err := cl.customers.RestoreCustomer(ctx, connect.NewRequest(&customersv1.RestoreCustomerRequest{Id: itoa(fx.custB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的客戶名稱", "customers", "name", fx.custB, "B 客戶")
		assertNullColumn(t, admin, "B 的客戶未被軟刪除", "customers", "deleted_at", fx.custB)
	})

	t.Run("customer_addresses", func(t *testing.T) {
		res := crossCall(t, "ListAddresses(自家客戶)", func(ctx context.Context) (*connect.Response[customersv1.ListAddressesResponse], error) {
			return cl.customers.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: itoa(fx.custA)}))
		})
		ids := make([]string, 0, len(res.Msg.GetAddresses()))
		for _, a := range res.Msg.GetAddresses() {
			ids = append(ids, a.GetId())
		}
		assertSetEquals(t, "ListAddresses(自家客戶)", ids, itoa(fx.addrA))

		// 用 B 的客戶 id 查地址:子表的路徑必須先在租戶範圍內找到父列(requireCustomer)。
		crossCallErr(t, "ListAddresses(他公司客戶)", func(ctx context.Context) error {
			_, err := cl.customers.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: itoa(fx.custB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateAddress(他公司)", func() error {
			city := "改到 B 市"
			_, err := cl.customers.UpdateAddress(ctx, connect.NewRequest(&customersv1.UpdateAddressRequest{
				Id: itoa(fx.addrB), City: &city,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteAddress(他公司)", func() error {
			_, err := cl.customers.DeleteAddress(ctx, connect.NewRequest(&customersv1.DeleteAddressRequest{Id: itoa(fx.addrB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的地址城市", "customer_addresses", "city", fx.addrB, "B 市")
		assertNullColumn(t, admin, "B 的地址未被軟刪除", "customer_addresses", "deleted_at", fx.addrB)
	})

	t.Run("customer_contacts", func(t *testing.T) {
		res := crossCall(t, "ListContacts(自家客戶)", func(ctx context.Context) (*connect.Response[customersv1.ListContactsResponse], error) {
			return cl.customers.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: itoa(fx.custA)}))
		})
		ids := make([]string, 0, len(res.Msg.GetContacts()))
		for _, c := range res.Msg.GetContacts() {
			ids = append(ids, c.GetId())
		}
		assertSetEquals(t, "ListContacts(自家客戶)", ids, itoa(fx.contactA))

		crossCallErr(t, "ListContacts(他公司客戶)", func(ctx context.Context) error {
			_, err := cl.customers.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: itoa(fx.custB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateContact(他公司)", func() error {
			title := "改到 B 職稱"
			_, err := cl.customers.UpdateContact(ctx, connect.NewRequest(&customersv1.UpdateContactRequest{
				Id: itoa(fx.contactB), Title: &title,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteContact(他公司)", func() error {
			_, err := cl.customers.DeleteContact(ctx, connect.NewRequest(&customersv1.DeleteContactRequest{Id: itoa(fx.contactB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的聯絡人職稱", "customer_contacts", "title", fx.contactB, "B 職稱")
		assertNullColumn(t, admin, "B 的聯絡人未被軟刪除", "customer_contacts", "deleted_at", fx.contactB)
	})

	t.Run("cross_tenant_writes_with_B_parent", func(t *testing.T) {
		// 這一組的目標不是「B 的某一列」而是 **B 的父列／目標使用者**:子表新增、以公司 id 建部門、
		// 以使用者 id 指派角色／強制登出。RLS 讓 B 的父列在 A 的請求交易裡不存在,故寫入一律不得成立。
		assertCrossTenantDenied(t, admin, "AddAddress(B 的客戶)", func() error {
			_, err := cl.customers.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
				CustomerId: itoa(fx.custB), Type: "shipping", RecipientName: "偷渡收件人", AddressLine: "偷渡地址",
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "AddContact(B 的客戶)", func() error {
			_, err := cl.customers.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{
				CustomerId: itoa(fx.custB), Name: "偷渡聯絡人",
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "AssignRole(B 的使用者)", func() error {
			_, err := cl.users.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{
				UserId: itoa(fx.actorB), Role: "staff",
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "ForceLogout(B 的使用者)", func() error {
			_, err := cl.users.ForceLogout(ctx, connect.NewRequest(&v1.ForceLogoutRequest{UserId: itoa(fx.actorB)}))
			return err
		})
		assertCrossTenantDenied(t, admin, "CreateDepartment(B 的公司)", func() error {
			_, err := cl.departments.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
				CompanyId: itoa(fx.coB), Name: "偷渡部門",
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "CreateUser(B 的公司)", func() error {
			_, err := cl.users.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
				Name: "偷渡帳號", Email: "smuggled@example.com", CompanyId: itoa(fx.coB), Role: "staff",
			}))
			return err
		})

		// 不得落地:全表計數(A 自己也不得多出列 —— 探針只建了兩家公司的等價 fixture)。
		for _, tbl := range []string{"departments", "users", "products", "customer_addresses", "customer_contacts"} {
			if n := countRows(t, admin, `SELECT count(*) FROM `+tbl); n != 2 {
				t.Errorf("被拒的跨租戶寫入不得新增 %s 的列:應維持 2 列(兩家公司各一),got %d", tbl, n)
			}
		}
		if n := countRows(t, admin, `SELECT count(*) FROM users WHERE email = 'smuggled@example.com'`); n != 0 {
			t.Errorf("被拒的 CreateUser 不得落地任何帳號,got %d 列", n)
		}

		// CreateProduct 帶 B 的商品分類:服務層的參照解析以「**範圍內**恰一列存在」為準
		// (validateCategoryRef → validateDeptMasterRef),B 的分類在 A 的請求交易裡不存在 →
		// invalid_argument(引用非法,非權限錯誤;這是既有且刻意的契約)。本斷言釘的是
		// 「B 的列不會被當成合法參照」與「不得落地」,錯誤碼不是 not_found。
		beforeAudit := countRows(t, admin, `SELECT count(*) FROM audit_logs`)
		productErr := func() error {
			_, err := cl.products.CreateProduct(ctx, connect.NewRequest(&productsv1.CreateProductRequest{
				Code:       "XT-SMUGGLE",
				Name:       "偷渡商品",
				CategoryId: fx.masters["product_categories"][1],
				Units:      []*productsv1.ProductUnit{{UnitCode: "PCS", ConversionRate: "1", IsBase: true}},
			}))
			return err
		}()
		if connect.CodeOf(productErr) != connect.CodeInvalidArgument {
			t.Errorf("CreateProduct 帶 B 的商品分類應 invalid_argument(參照在範圍外),got %v", productErr)
		}
		if after := countRows(t, admin, `SELECT count(*) FROM audit_logs`); after != beforeAudit {
			t.Errorf("被拒的 CreateProduct 不得留下稽核列(%d → %d)", beforeAudit, after)
		}
		// 正向對照:同一個請求形狀、只把分類換成自家的 → 必須成功。
		// 沒有這一段,上面那個 invalid_argument 可能只是「單位規格寫錯」之類的假證據。
		created := crossCall(t, "CreateProduct(自家分類)", func(ctx context.Context) (*connect.Response[productsv1.CreateProductResponse], error) {
			return cl.products.CreateProduct(ctx, connect.NewRequest(&productsv1.CreateProductRequest{
				Code:       "XT-OWN",
				Name:       "自家商品",
				CategoryId: fx.masters["product_categories"][0],
				Units:      []*productsv1.ProductUnit{{UnitCode: "PCS", ConversionRate: "1", IsBase: true}},
			}))
		})
		if created.Msg.GetProduct().GetCompanyId() != itoa(fx.coA) {
			t.Fatalf("自家分類建商品應落在公司 %d,got %+v", fx.coA, created.Msg.GetProduct())
		}
		assertColumnValue(t, admin, "自家新建商品的公司", "products", "company_id", mustItoa(t, created.Msg.GetProduct().GetId()), itoa(fx.coA))

		// B 的使用者:角色／token_version／狀態皆不得被動到(AssignRole 會改角色、ForceLogout 會 +1)。
		assertColumnValue(t, admin, "B 的使用者角色", "users", "role", fx.actorB, "company_admin")
		assertColumnValue(t, admin, "B 的 token_version", "users", "token_version", fx.actorB, "0")
		assertColumnValue(t, admin, "B 的使用者狀態", "users", "status", fx.actorB, "active")
	})

	t.Run("masters", func(t *testing.T) {
		masters := []struct {
			table                 string
			list                  func(ctx context.Context) ([]string, error)
			update, drop, restore func(ctx context.Context) error
		}{
			{"warehouses",
				func(ctx context.Context) ([]string, error) {
					res, err := cl.warehouses.ListWarehouses(ctx, connect.NewRequest(&mastersv1.ListWarehousesRequest{Page: 1, PageSize: 50}))
					if err != nil {
						return nil, err
					}
					ids := make([]string, 0, len(res.Msg.GetWarehouses()))
					for _, w := range res.Msg.GetWarehouses() {
						ids = append(ids, w.GetId())
					}
					return ids, nil
				},
				func(ctx context.Context) error {
					n := "改到 B 倉"
					_, err := cl.warehouses.UpdateWarehouse(ctx, connect.NewRequest(&mastersv1.UpdateWarehouseRequest{Id: fx.masters["warehouses"][1], Name: &n}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.warehouses.DeleteWarehouse(ctx, connect.NewRequest(&mastersv1.DeleteWarehouseRequest{Id: fx.masters["warehouses"][1]}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.warehouses.RestoreWarehouse(ctx, connect.NewRequest(&mastersv1.RestoreWarehouseRequest{Id: fx.masters["warehouses"][1]}))
					return err
				}},
			{"routes",
				func(ctx context.Context) ([]string, error) {
					res, err := cl.routes.ListRoutes(ctx, connect.NewRequest(&mastersv1.ListRoutesRequest{Page: 1, PageSize: 50}))
					if err != nil {
						return nil, err
					}
					ids := make([]string, 0, len(res.Msg.GetRoutes()))
					for _, r := range res.Msg.GetRoutes() {
						ids = append(ids, r.GetId())
					}
					return ids, nil
				},
				func(ctx context.Context) error {
					n := "改到 B 車次"
					_, err := cl.routes.UpdateRoute(ctx, connect.NewRequest(&mastersv1.UpdateRouteRequest{Id: fx.masters["routes"][1], Name: &n}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.routes.DeleteRoute(ctx, connect.NewRequest(&mastersv1.DeleteRouteRequest{Id: fx.masters["routes"][1]}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.routes.RestoreRoute(ctx, connect.NewRequest(&mastersv1.RestoreRouteRequest{Id: fx.masters["routes"][1]}))
					return err
				}},
			{"processing_specs",
				func(ctx context.Context) ([]string, error) {
					res, err := cl.specs.ListProcessingSpecs(ctx, connect.NewRequest(&mastersv1.ListProcessingSpecsRequest{Page: 1, PageSize: 50}))
					if err != nil {
						return nil, err
					}
					ids := make([]string, 0, len(res.Msg.GetProcessingSpecs()))
					for _, s := range res.Msg.GetProcessingSpecs() {
						ids = append(ids, s.GetId())
					}
					return ids, nil
				},
				func(ctx context.Context) error {
					n := "改到 B 加工規格"
					_, err := cl.specs.UpdateProcessingSpec(ctx, connect.NewRequest(&mastersv1.UpdateProcessingSpecRequest{Id: fx.masters["processing_specs"][1], Name: &n}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.specs.DeleteProcessingSpec(ctx, connect.NewRequest(&mastersv1.DeleteProcessingSpecRequest{Id: fx.masters["processing_specs"][1]}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.specs.RestoreProcessingSpec(ctx, connect.NewRequest(&mastersv1.RestoreProcessingSpecRequest{Id: fx.masters["processing_specs"][1]}))
					return err
				}},
			{"product_categories",
				func(ctx context.Context) ([]string, error) {
					res, err := cl.cats.ListProductCategories(ctx, connect.NewRequest(&mastersv1.ListProductCategoriesRequest{Page: 1, PageSize: 50}))
					if err != nil {
						return nil, err
					}
					ids := make([]string, 0, len(res.Msg.GetProductCategories()))
					for _, c := range res.Msg.GetProductCategories() {
						ids = append(ids, c.GetId())
					}
					return ids, nil
				},
				func(ctx context.Context) error {
					n := "改到 B 分類"
					_, err := cl.cats.UpdateProductCategory(ctx, connect.NewRequest(&mastersv1.UpdateProductCategoryRequest{Id: fx.masters["product_categories"][1], Name: &n}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.cats.DeleteProductCategory(ctx, connect.NewRequest(&mastersv1.DeleteProductCategoryRequest{Id: fx.masters["product_categories"][1]}))
					return err
				},
				func(ctx context.Context) error {
					_, err := cl.cats.RestoreProductCategory(ctx, connect.NewRequest(&mastersv1.RestoreProductCategoryRequest{Id: fx.masters["product_categories"][1]}))
					return err
				}},
		}
		for _, m := range masters {
			t.Run(m.table, func(t *testing.T) {
				ids, err := m.list(ctx)
				if err != nil {
					t.Fatalf("List %s: %v", m.table, err)
				}
				assertSetEquals(t, "List "+m.table, ids, fx.masters[m.table][0])
				assertCrossTenantDenied(t, admin, "Update "+m.table, func() error { return m.update(ctx) })
				assertCrossTenantDenied(t, admin, "Delete "+m.table, func() error { return m.drop(ctx) })
				// Restore 與 Delete 同型:軟刪除路徑不得成為「以 id 復原他家的列」的後門。
				assertCrossTenantDenied(t, admin, "Restore "+m.table, func() error { return m.restore(ctx) })
				assertColumnValue(t, admin, "B 的 "+m.table+" 名稱", m.table, "name", mustItoa(t, fx.masters[m.table][1]), "B 列")
				assertNullColumn(t, admin, "B 的 "+m.table+" 未被軟刪除", m.table, "deleted_at", mustItoa(t, fx.masters[m.table][1]))
			})
		}
	})

	t.Run("products", func(t *testing.T) {
		res := crossCall(t, "ListProducts", func(ctx context.Context) (*connect.Response[productsv1.ListProductsResponse], error) {
			return cl.products.ListProducts(ctx, connect.NewRequest(&productsv1.ListProductsRequest{Page: 1, PageSize: 50}))
		})
		ids := make([]string, 0, len(res.Msg.GetProducts()))
		for _, p := range res.Msg.GetProducts() {
			ids = append(ids, p.GetId())
		}
		assertSetEquals(t, "ListProducts", ids, tenantScopedIDs(t, admin, "products", fx.coA, fx.deptA)...)
		if !containsValue(ids, itoa(fx.productA)) {
			t.Fatalf("A 自家的商品必須看得到(id=%d),got %v", fx.productA, ids)
		}

		crossCallErr(t, "GetProduct(他公司)", func(ctx context.Context) error {
			_, err := cl.products.GetProduct(ctx, connect.NewRequest(&productsv1.GetProductRequest{Id: itoa(fx.productB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateProduct(他公司)", func() error {
			n := "改到 B 商品"
			_, err := cl.products.UpdateProduct(ctx, connect.NewRequest(&productsv1.UpdateProductRequest{
				Id: itoa(fx.productB), Name: &n,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteProduct(他公司)", func() error {
			_, err := cl.products.DeleteProduct(ctx, connect.NewRequest(&productsv1.DeleteProductRequest{Id: itoa(fx.productB)}))
			return err
		})
		assertCrossTenantDenied(t, admin, "RestoreProduct(他公司)", func() error {
			_, err := cl.products.RestoreProduct(ctx, connect.NewRequest(&productsv1.RestoreProductRequest{Id: itoa(fx.productB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的商品名稱", "products", "name", fx.productB, "商品")
		assertNullColumn(t, admin, "B 的商品未被軟刪除", "products", "deleted_at", fx.productB)
	})

	t.Run("metadicts", func(t *testing.T) {
		// 指定 department_id=deptA:可見集合 = 系統預設列 + deptA 的擴充列(metadicts 以部門為
		// 租戶鍵,沒有 company_id)。以 admin 真值取同一條件的集合,多的少的都要紅。
		res := crossCall(t, "ListMetadicts", func(ctx context.Context) (*connect.Response[metadictv1.ListMetadictsResponse], error) {
			return cl.metadicts.ListMetadicts(ctx, connect.NewRequest(&metadictv1.ListMetadictsRequest{
				Page: 1, PageSize: 50, Type: "unit", DepartmentId: itoa(fx.deptA),
			}))
		})
		ids := make([]string, 0, len(res.Msg.GetItems()))
		for _, m := range res.Msg.GetItems() {
			ids = append(ids, m.GetId())
		}
		assertSetEquals(t, "ListMetadicts", ids, adminIDs(t, admin,
			`SELECT id FROM metadicts WHERE type = 'unit' AND deleted_at IS NULL
			   AND (department_id IS NULL OR department_id = $1)`, fx.deptA)...)
		if !containsValue(ids, itoa(fx.metadictA)) {
			t.Fatalf("A 部門自家的字典擴充列必須看得到(id=%d),got %v", fx.metadictA, ids)
		}

		crossCallErr(t, "GetMetadict(他部門)", func(ctx context.Context) error {
			_, err := cl.metadicts.GetMetadict(ctx, connect.NewRequest(&metadictv1.GetMetadictRequest{Id: itoa(fx.metadictB)}))
			return err
		}, connect.CodeNotFound)

		assertCrossTenantDenied(t, admin, "UpdateMetadict(他部門)", func() error {
			n := "改到 B 字典"
			_, err := cl.metadicts.UpdateMetadict(ctx, connect.NewRequest(&metadictv1.UpdateMetadictRequest{
				Id: itoa(fx.metadictB), DisplayName: &n,
			}))
			return err
		})
		assertCrossTenantDenied(t, admin, "DeleteMetadict(他部門)", func() error {
			_, err := cl.metadicts.DeleteMetadict(ctx, connect.NewRequest(&metadictv1.DeleteMetadictRequest{Id: itoa(fx.metadictB)}))
			return err
		})
		assertColumnValue(t, admin, "B 的字典名稱", "metadicts", "display_name", fx.metadictB, "B 單位")
		assertNullColumn(t, admin, "B 的字典未被軟刪除", "metadicts", "deleted_at", fx.metadictB)
	})

	t.Run("metadict_list_options", func(t *testing.T) {
		// 表單下拉選項:super 身分只回系統預設列(服務層已限縮,RLS 再擋一次);斷言集合恰為
		// 真值,且不得出現 B 部門的私有值。
		res := crossCall(t, "ListOptions", func(ctx context.Context) (*connect.Response[metadictv1.ListOptionsResponse], error) {
			return cl.metadicts.ListOptions(ctx, connect.NewRequest(&metadictv1.ListOptionsRequest{Type: "unit"}))
		})
		codes := make([]string, 0, len(res.Msg.GetOptions()))
		for _, o := range res.Msg.GetOptions() {
			codes = append(codes, o.GetCode())
		}
		assertSetEquals(t, "ListOptions", codes, adminTexts(t, admin,
			`SELECT code FROM metadicts WHERE type = 'unit' AND is_active AND deleted_at IS NULL
			   AND department_id IS NULL`)...)
		if containsValue(codes, "XT-B") {
			t.Fatalf("下拉選項洩漏了 B 部門的字典值(XT-B)")
		}
	})

	t.Run("audits", func(t *testing.T) {
		// 稽核清單在 super 身分下**沒有**服務層的 company 過濾(company_admin 才有)→ 純 RLS 偵測器。
		res := crossCall(t, "ListAuditLogs", func(ctx context.Context) (*connect.Response[auditv1.ListAuditLogsResponse], error) {
			return cl.audits.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Page: 1, PageSize: 50}))
		})
		ids := make([]string, 0, len(res.Msg.GetItems()))
		for _, a := range res.Msg.GetItems() {
			ids = append(ids, a.GetId())
		}
		// 真值包含本探針正向對照寫下的稽核(A 客戶的 update)。
		assertSetEquals(t, "ListAuditLogs", ids, adminIDs(t, admin, `SELECT id FROM audit_logs WHERE company_id = $1`, fx.coA)...)
		if !containsValue(ids, itoa(fx.auditA)) {
			t.Fatalf("A 自家的稽核列必須看得到(id=%d),got %v", fx.auditA, ids)
		}
		if containsValue(ids, itoa(fx.auditB)) {
			t.Fatalf("稽核清單洩漏了 B 公司的稽核列(id=%d)", fx.auditB)
		}
	})
}

// seedCrossTenant 以 admin(superuser,恆繞過 RLS)建立兩家公司的等價資料。
func seedCrossTenant(t *testing.T, admin *sql.DB) crossTenantFixture {
	t.Helper()
	fx := crossTenantFixture{}
	fx.coA = insertRLSCompany(t, admin, "跨租戶 A", "XT-A")
	fx.coB = insertRLSCompany(t, admin, "跨租戶 B", "XT-B")
	fx.deptA = insertRLSDepartment(t, admin, fx.coA, "跨租戶 A 部門")
	fx.deptB = insertRLSDepartment(t, admin, fx.coB, "跨租戶 B 部門")
	fx.actorA = insertRLSUser(t, admin, fx.coA, "xt-a@example.com", "company_admin")
	fx.actorB = insertRLSUser(t, admin, fx.coB, "xt-b@example.com", "company_admin")
	if _, err := admin.Exec(`UPDATE users SET name = '跨租戶 B 管理員' WHERE id = $1`, fx.actorB); err != nil {
		t.Fatalf("命名 B 管理員: %v", err)
	}
	setRLSUserDepartment(t, admin, fx.actorA, fx.deptA)
	setRLSUserDepartment(t, admin, fx.actorB, fx.deptB)

	fx.custA = insertRLSCustomer(t, admin, fx.coA, "XT-C-A", "A 客戶")
	fx.custB = insertRLSCustomer(t, admin, fx.coB, "XT-C-B", "B 客戶")
	fx.addrA = insertCrossTenantAddress(t, admin, fx.coA, fx.custA, "A 市")
	fx.addrB = insertCrossTenantAddress(t, admin, fx.coB, fx.custB, "B 市")
	fx.contactA = insertCrossTenantContact(t, admin, fx.coA, fx.custA, "A 職稱")
	fx.contactB = insertCrossTenantContact(t, admin, fx.coB, fx.custB, "B 職稱")

	fx.masters = map[string][2]string{}
	for _, tbl := range mastersRLSTables {
		fx.masters[tbl.table] = [2]string{
			insertMasterRow(t, admin, tbl.insert, fx.coA, "XT-A", "A 列"),
			insertMasterRow(t, admin, tbl.insert, fx.coB, "XT-B", "B 列"),
		}
	}
	fx.productA = insertRLSProduct(t, admin, fx.coA, "XT-PD-A")
	fx.productB = insertRLSProduct(t, admin, fx.coB, "XT-PD-B")
	fx.metadictA = insertRLSMetadict(t, admin, "unit", "XT-A", "A 單位", &fx.deptA)
	fx.metadictB = insertRLSMetadict(t, admin, "unit", "XT-B", "B 單位", &fx.deptB)
	fx.auditA = insertRLSAuditRow(t, admin, fx.coA, fx.actorA, "create")
	fx.auditB = insertRLSAuditRow(t, admin, fx.coB, fx.actorB, "create")
	return fx
}

// insertCrossTenantAddress 以 admin 連線建一筆地址(公司層:department_id NULL,與服務層的
// 複寫語意一致)。
func insertCrossTenantAddress(t *testing.T, db *sql.DB, companyID, customerID int, city string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO customer_addresses (company_id, customer_id, type, recipient_name, address_line, city)
		 VALUES ($1, $2, 'shipping', '收件人', '地址', $3) RETURNING id`,
		companyID, customerID, city).Scan(&id); err != nil {
		t.Fatalf("建地址(公司 %d): %v", companyID, err)
	}
	return id
}

// insertCrossTenantContact 以 admin 連線建一筆聯絡人。
func insertCrossTenantContact(t *testing.T, db *sql.DB, companyID, customerID int, title string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO customer_contacts (company_id, customer_id, name, title)
		 VALUES ($1, $2, '聯絡人', $3) RETURNING id`,
		companyID, customerID, title).Scan(&id); err != nil {
		t.Fatalf("建聯絡人(公司 %d): %v", companyID, err)
	}
	return id
}

// openScopedClient 以既有的 app_rw 連線池建立業務 client(**必須**經 dbtenant.NewClient)。
func openScopedClient(t *testing.T, pool *sql.DB) *ent.Client {
	t.Helper()
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// crossCall 執行一次「在這個 scope 下應該成功」的呼叫;失敗即測試失敗(多半是漏掛租戶交易
// 導致 RLS 把自家的列也濾掉)。
func crossCall[T any](t *testing.T, endpoint string, fn func(context.Context) (T, error)) T {
	t.Helper()
	v, err := fn(t.Context())
	if err != nil {
		t.Fatalf("%s:以 A 的身分呼叫自家端點不該失敗: %v", endpoint, err)
	}
	return v
}

// crossCallErr 執行一次應該失敗的呼叫,並斷言錯誤碼。
func crossCallErr(t *testing.T, endpoint string, fn func(context.Context) error, want connect.Code) {
	t.Helper()
	if err := fn(t.Context()); connect.CodeOf(err) != want {
		t.Fatalf("%s:應 %v,got %v", endpoint, want, err)
	}
}

// assertCrossTenantDenied 執行一次跨租戶寫入嘗試,並斷言兩件事:
//
//	① 回應是 not_found(RLS 先濾掉目標 → 服務層的 ACL 沒機會把「目標存在」洩漏出去);
//	② 這次嘗試沒有留下任何稽核列(以全表計數為準 —— 不同的服務對稽核的 company_id 取值不同:
//	   主檔用目標所屬公司、字典用操作者公司,故逐表比對會漏)。
func assertCrossTenantDenied(t *testing.T, admin *sql.DB, endpoint string, attempt func() error) {
	t.Helper()
	before := countRows(t, admin, `SELECT count(*) FROM audit_logs`)
	err := attempt()
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("%s:以 A 的身分對 B 公司的列應 not_found,got %v", endpoint, err)
	}
	if after := countRows(t, admin, `SELECT count(*) FROM audit_logs`); after != before {
		t.Errorf("%s:被拒的跨租戶寫入不得留下稽核列(audit_logs %d → %d)", endpoint, before, after)
	}
}

// assertSetEquals 斷言端點回傳的識別值集合與期望集合完全相同(少了自家的列或多出他家的列都要紅)。
func assertSetEquals(t *testing.T, endpoint string, got []string, want ...string) {
	t.Helper()
	if len(want) == 0 {
		// 期望集合算成空集合時,任何「剛好也回空」的實作都會假通過 → 這種斷言沒有偵測力,直接紅。
		t.Fatalf("%s:真值集合為空,本斷言不成立(修正真值查詢或 fixture)", endpoint)
	}
	gotSet := map[string]bool{}
	for _, id := range got {
		if gotSet[id] {
			t.Errorf("%s:重複的值 %s", endpoint, id)
		}
		gotSet[id] = true
	}
	for _, id := range want {
		if !gotSet[id] {
			t.Errorf("%s:缺少期望值 %s(實際 %v)", endpoint, id, got)
		}
	}
	if len(gotSet) != len(want) {
		t.Errorf("%s:應為 %d 筆 %v,得到 %d 筆 %v(多出的列可能是別家公司的資料洩漏)", endpoint, len(want), want, len(gotSet), got)
	}
}

// adminIDs 以 admin(superuser)取真值 id 集合:superuser 不受 RLS 影響,故它看到的才是
// 「該條件下真正存在的列」;端點回傳的必須恰好是這個集合。
func adminIDs(t *testing.T, admin *sql.DB, query string, args ...any) []string {
	t.Helper()
	rows, err := admin.Query(query, args...)
	if err != nil {
		t.Fatalf("取真值(%s): %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("掃描真值: %v", err)
		}
		ids = append(ids, itoa(id))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("讀真值: %v", err)
	}
	return ids
}

// adminTexts 以 admin(superuser)取單一文字欄的真值集合(供非 id 欄位:字典 code、角色 code 等)。
func adminTexts(t *testing.T, admin *sql.DB, query string, args ...any) []string {
	t.Helper()
	rows, err := admin.Query(query, args...)
	if err != nil {
		t.Fatalf("取文字真值(%s): %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("掃描文字真值: %v", err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("讀文字真值: %v", err)
	}
	return out
}

// tenantScopedIDs 取「本部門 scope 可見」的租戶表列 id:公司相符且(公司層列或本部門列)。
// 與 policy 的條件同構,但以 admin 執行 → 這是真值,不依賴 RLS 本身。
func tenantScopedIDs(t *testing.T, admin *sql.DB, table string, companyID, departmentID int) []string {
	t.Helper()
	return adminIDs(t, admin,
		`SELECT id FROM `+table+` WHERE company_id = $1 AND deleted_at IS NULL
		   AND (department_id IS NULL OR department_id = $2)`, companyID, departmentID)
}

// assertColumnValue 以 admin 真值斷言某列某欄仍是原值(跨租戶寫入沒有副作用)。
func assertColumnValue(t *testing.T, admin *sql.DB, what, table, column string, id int, want string) {
	t.Helper()
	var got string
	if err := admin.QueryRow(`SELECT `+column+`::text FROM `+table+` WHERE id = $1`, id).Scan(&got); err != nil {
		t.Fatalf("查 %s(%s.id=%d): %v", what, table, id, err)
	}
	if got != want {
		t.Errorf("%s 被改動了:應為 %q,got %q", what, want, got)
	}
}

// assertNullColumn 以 admin 真值斷言某列某欄仍為 NULL(跨租戶刪除沒有副作用)。
func assertNullColumn(t *testing.T, admin *sql.DB, what, table, column string, id int) {
	t.Helper()
	var got sql.NullString
	if err := admin.QueryRow(`SELECT `+column+`::text FROM `+table+` WHERE id = $1`, id).Scan(&got); err != nil {
		t.Fatalf("查 %s(%s.id=%d): %v", what, table, id, err)
	}
	if got.Valid {
		t.Errorf("%s 被改動了:應為 NULL,got %q", what, got.String)
	}
}

// containsValue 判斷值是否在集合中(供 id／code 等識別值共用)。
func containsValue(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

// containsRoleCode 判斷角色清單是否含指定 code。
func containsRoleCode(roles []*v1.Role, want string) bool {
	for _, r := range roles {
		if r.GetCode() == want {
			return true
		}
	}
	return false
}
