//go:build integration

package services

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSCustomersIsolation 以 app_rw 直連驗證客戶域隔離（00024 ENABLE + FORCE）：
// 未設 scope → 看不到任何列（fail-closed）；設 A 公司 → 只看得到 A；
// 以 A 的身分寫入 B 公司的列 → 被 WITH CHECK 擋下。
//
// 為何用 app_rw 而非 owner：owner 需 FORCE 才受 policy 約束，而容器／測試的 admin 連線是
// superuser（PG 的 superuser 永遠繞過 RLS，FORCE 亦然）→ 以 admin 連線根本測不到 fail-closed。
// app_rw 是 00022 建出的 NOBYPASSRLS 非 owner 業務角色，正是生產路徑的角色。
func TestIntegrationRLSCustomersIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("admin 連線: %v", err)
	}
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "RLS-A")
	coB := insertRLSCompany(t, admin, "B", "RLS-B")
	for _, co := range []int{coA, coB} {
		if _, err := admin.Exec(
			`INSERT INTO customers (company_id, customer_code, name, created_at, updated_at)
			 VALUES ($1, $2, '客戶', now(), now())`, co, "C-"+strconv.Itoa(co)); err != nil {
			t.Fatalf("建客戶(公司 %d): %v", co, err)
		}
	}

	app := openAppRoleDB(t, adminDSN)

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		var n int
		if err := app.QueryRow(`SELECT count(*) FROM customers`).Scan(&n); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if n != 0 {
			t.Fatalf("未設 scope 時不得看到任何客戶,got %d", n)
		}
	})

	t.Run("scope=company A → 只見 A", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScope(t, tx, coA)
		var got int
		if err := tx.QueryRow(`SELECT DISTINCT company_id FROM customers`).Scan(&got); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if got != coA {
			t.Fatalf("應只看到公司 %d,got %d", coA, got)
		}
	})

	t.Run("跨租戶 INSERT → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		setAppScope(t, tx, coA)
		_, err = tx.Exec(
			`INSERT INTO customers (company_id, customer_code, name, created_at, updated_at)
			 VALUES ($1, 'X-1', '別家客戶', now(), now())`, coB)
		if err == nil {
			t.Fatal("以 A 的身分寫入 B 公司的客戶必須被擋（WITH CHECK）")
		}
	})
}

// insertRLSCompany 以 admin 連線建一間公司並回傳 id（admin 為 superuser，不受 RLS 約束，
// 故 fixture 不必設 scope）。
func insertRLSCompany(t *testing.T, db *sql.DB, name, identifier string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ($1, $2, 'active') RETURNING id`,
		name, identifier).Scan(&id); err != nil {
		t.Fatalf("建公司 %s: %v", name, err)
	}
	return id
}

// insertRLSCustomer 以 admin 連線建一筆客戶並回傳 id。
func insertRLSCustomer(t *testing.T, db *sql.DB, companyID int, code, name string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO customers (company_id, customer_code, name) VALUES ($1, $2, $3) RETURNING id`,
		companyID, code, name).Scan(&id); err != nil {
		t.Fatalf("建客戶 %s(公司 %d): %v", code, companyID, err)
	}
	return id
}

// TestIntegrationCustomerServiceUnderAppRole 以 app_rw + 請求層租戶交易跑**真 handler**:
// 逐條走過客戶域已遷移的路徑(27 處 dbtenant.Client、10 處請求交易),驗證它們在 RLS 生效時
// 確實在租戶範圍內讀寫(每條都斷言回來的列屬身分所屬公司,而不是 0 列或他人資料);
// 結尾的負向子測試證明「漏掛 scope 這條接線」會紅,而不只是「有接線」。
//
// 為何既有客戶域整合測試擋不住:它們的連線是容器的 superuser(PG 的 superuser 永遠繞過 RLS),
// 服務路徑漏掛 dbtenant.Client 也照樣全綠;必須換成非 superuser 的 app_rw 才測得到。
func TestIntegrationCustomerServiceUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()
	coA := insertRLSCompany(t, admin, "A", "SVC-A")
	coB := insertRLSCompany(t, admin, "B", "SVC-B")
	setCompanyCodePrefix(t, admin, coA, "SVC")
	setCompanyCodePrefix(t, admin, coB, "SVC")
	custA := insertRLSCustomer(t, admin, coA, "SVC-A-1", "A 客戶")
	custB := insertRLSCustomer(t, admin, coB, "SVC-B-1", "B 客戶")
	// 稽核列有 user_id FK(00010),操作者必須是真實使用者。
	actorA := insertRLSUser(t, admin, coA, "svc-admin-a@example.com", "company_admin")
	actorB := insertRLSUser(t, admin, coB, "svc-admin-b@example.com", "company_admin")

	// 業務 client 必須經 dbtenant.NewClient 建立(RLS 裝飾器才生效);DB 連線是 app_rw。
	client := dbtenant.NewClient(openAppRoleDB(t, adminDSN))
	t.Cleanup(func() { _ = client.Close() })

	// 受測 server:身分與 RLS scope 由 handler 注入(生產由 authzMiddleware 做)。
	svc := newCustomerAppRoleServer(t, client, actorA, coA, true)
	ctx := t.Context()

	// ── CreateCustomer(最高風險:唯一會寫 customer_counters 的路徑,該表的 policy 只認
	// company scope)────────────────────────────────────────────────────────────────
	created, err := svc.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
		Name: "新客戶", DefaultSalesRepId: itoa(actorA),
	}))
	if err != nil {
		t.Fatalf("CreateCustomer(app_rw + 租戶範圍): %v", err)
	}
	c := created.Msg.GetCustomer()
	if c.GetCompanyId() != itoa(coA) {
		t.Fatalf("CreateCustomer 回的客戶應屬公司 %d,得到 %q", coA, c.GetCompanyId())
	}
	if c.GetCustomerCode() != "SVC000001" {
		t.Fatalf("取號應為 SVC000001(counter 在本公司 scope 內建立並遞增),得到 %q", c.GetCustomerCode())
	}
	newCustID, err := strconv.Atoi(c.GetId())
	if err != nil {
		t.Fatalf("解析客戶 id %q: %v", c.GetId(), err)
	}
	// counter 列已於同一請求交易內建立(counter 表只有 company scope 的 policy)且已遞增。
	assertCounterNextSeq(t, admin, coA, 2)

	// ── 讀:自建 + fixture 兩筆,且不含 B 的客戶 ───────────────────────────────────
	list, err := svc.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{Page: 1, PageSize: 50}))
	if err != nil {
		t.Fatalf("ListCustomers: %v", err)
	}
	ids := map[string]bool{}
	for _, got := range list.Msg.GetCustomers() {
		if got.GetCompanyId() != itoa(coA) {
			t.Fatalf("ListCustomers 回傳了公司 %s 的客戶(身分屬公司 %d)", got.GetCompanyId(), coA)
		}
		ids[got.GetId()] = true
	}
	if len(ids) != 2 || !ids[c.GetId()] || !ids[itoa(custA)] {
		t.Fatalf("應只看到公司 %d 的兩筆客戶(id %s、%d),得到 %v", coA, c.GetId(), custA, ids)
	}
	if got, err := svc.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: itoa(custA)})); err != nil {
		t.Fatalf("GetCustomer(本租戶): %v", err)
	} else if got.Msg.GetCustomer().GetCompanyId() != itoa(coA) {
		t.Fatalf("GetCustomer 回的客戶應屬公司 %d,得到 %q", coA, got.Msg.GetCustomer().GetCompanyId())
	}
	// 跨租戶:B 的客戶在 A 的請求交易裡「不存在」(RLS 過濾,而非權限錯誤)。
	if _, err := svc.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: itoa(custB)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("以 A 的身分取 B 的客戶應 not_found,got %v", err)
	}

	// ── UpdateCustomer ───────────────────────────────────────────────────────────
	newName := "新客戶(改)"
	upd, err := svc.UpdateCustomer(ctx, connect.NewRequest(&customersv1.UpdateCustomerRequest{Id: c.GetId(), Name: &newName}))
	if err != nil {
		t.Fatalf("UpdateCustomer: %v", err)
	}
	if upd.Msg.GetCustomer().GetName() != newName || upd.Msg.GetCustomer().GetCompanyId() != itoa(coA) {
		t.Fatalf("UpdateCustomer 應回傳改名後、仍屬公司 %d 的客戶,得到 %+v", coA, upd.Msg.GetCustomer())
	}

	// ── DeleteCustomer → RestoreCustomer(兩者都要在租戶範圍內找到該列)─────────────
	if _, err := svc.DeleteCustomer(ctx, connect.NewRequest(&customersv1.DeleteCustomerRequest{Id: c.GetId()})); err != nil {
		t.Fatalf("DeleteCustomer: %v", err)
	}
	if _, err := svc.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: c.GetId()})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("軟刪除後 GetCustomer 應 not_found,got %v", err)
	}
	restored, err := svc.RestoreCustomer(ctx, connect.NewRequest(&customersv1.RestoreCustomerRequest{Id: c.GetId()}))
	if err != nil {
		t.Fatalf("RestoreCustomer: %v", err)
	}
	rc := restored.Msg.GetCustomer()
	if rc.GetCompanyId() != itoa(coA) || rc.GetDeletedAt() != "" {
		t.Fatalf("RestoreCustomer 應回傳已復原、屬公司 %d 的客戶,得到 %+v", coA, rc)
	}

	// ── 地址四支(requireCustomer/scopeQuery 都在請求交易內)─────────────────────
	addrRes, err := svc.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: c.GetId(), Type: "shipping", RecipientName: "收件人", AddressLine: "台北市",
	}))
	if err != nil {
		t.Fatalf("AddAddress: %v", err)
	}
	addr := addrRes.Msg.GetAddress()
	if addr.GetCustomerId() != c.GetId() {
		t.Fatalf("AddAddress 回的地址應屬客戶 %s,得到 %q", c.GetId(), addr.GetCustomerId())
	}
	if n := countRows(t, admin, `SELECT count(*) FROM customer_addresses WHERE customer_id = $1 AND company_id = $2`, newCustID, coA); n != 1 {
		t.Fatalf("地址應以公司 %d 落地並已提交,got %d 列", coA, n)
	}
	newCity := "台中市"
	updAddr, err := svc.UpdateAddress(ctx, connect.NewRequest(&customersv1.UpdateAddressRequest{Id: addr.GetId(), City: &newCity}))
	if err != nil {
		t.Fatalf("UpdateAddress: %v", err)
	}
	if updAddr.Msg.GetAddress().GetCity() != newCity || updAddr.Msg.GetAddress().GetCustomerId() != c.GetId() {
		t.Fatalf("UpdateAddress 應回傳更新後、仍屬客戶 %s 的地址,得到 %+v", c.GetId(), updAddr.Msg.GetAddress())
	}
	addrs, err := svc.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: c.GetId()}))
	if err != nil {
		t.Fatalf("ListAddresses: %v", err)
	}
	if got := addrs.Msg.GetAddresses(); len(got) != 1 || got[0].GetId() != addr.GetId() {
		t.Fatalf("應只看到該客戶的 1 筆地址(%s),得到 %+v", addr.GetId(), got)
	}
	if _, err := svc.DeleteAddress(ctx, connect.NewRequest(&customersv1.DeleteAddressRequest{Id: addr.GetId()})); err != nil {
		t.Fatalf("DeleteAddress: %v", err)
	}
	if after, err := svc.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: c.GetId()})); err != nil {
		t.Fatalf("ListAddresses(刪除後): %v", err)
	} else if n := len(after.Msg.GetAddresses()); n != 0 {
		t.Fatalf("軟刪除後地址清單應為 0 筆,得到 %d", n)
	}
	// 跨租戶:B 的客戶不在 A 的範圍 → 地址路徑的 requireCustomer 必須 not_found。
	if _, err := svc.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: itoa(custB)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("以 A 的身分查 B 客戶的地址應 not_found,got %v", err)
	}

	// ── 聯絡人四支(同型)────────────────────────────────────────────────────────
	contactRes, err := svc.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{
		CustomerId: c.GetId(), Name: "聯絡人",
	}))
	if err != nil {
		t.Fatalf("AddContact: %v", err)
	}
	contact := contactRes.Msg.GetContact()
	if contact.GetCustomerId() != c.GetId() {
		t.Fatalf("AddContact 回的聯絡人應屬客戶 %s,得到 %q", c.GetId(), contact.GetCustomerId())
	}
	newTitle := "經理"
	updContact, err := svc.UpdateContact(ctx, connect.NewRequest(&customersv1.UpdateContactRequest{Id: contact.GetId(), Title: &newTitle}))
	if err != nil {
		t.Fatalf("UpdateContact: %v", err)
	}
	if updContact.Msg.GetContact().GetTitle() != newTitle || updContact.Msg.GetContact().GetCustomerId() != c.GetId() {
		t.Fatalf("UpdateContact 應回傳更新後、仍屬客戶 %s 的聯絡人,得到 %+v", c.GetId(), updContact.Msg.GetContact())
	}
	contacts, err := svc.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: c.GetId()}))
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	if got := contacts.Msg.GetContacts(); len(got) != 1 || got[0].GetId() != contact.GetId() {
		t.Fatalf("應只看到該客戶的 1 筆聯絡人(%s),得到 %+v", contact.GetId(), got)
	}
	if _, err := svc.DeleteContact(ctx, connect.NewRequest(&customersv1.DeleteContactRequest{Id: contact.GetId()})); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}
	if after, err := svc.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: c.GetId()})); err != nil {
		t.Fatalf("ListContacts(刪除後): %v", err)
	} else if n := len(after.Msg.GetContacts()); n != 0 {
		t.Fatalf("軟刪除後聯絡人清單應為 0 筆,得到 %d", n)
	}
	if _, err := svc.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: itoa(custB)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("以 A 的身分查 B 客戶的聯絡人應 not_found,got %v", err)
	}

	// ── 負向對照:同一組 app_rw server,但**不注入 RLS scope**(模擬漏掛)───────────
	// 少了 scope 就沒有 SET LOCAL → 業務角色在 RLS 下 fail-closed:讀回 0 列、寫被 WITH CHECK
	// 擋下。少了這一段,本測試只證明「有接線」;有了它才證明「接線錯了會紅」。
	t.Run("未注入 RLS scope → fail-closed", func(t *testing.T) {
		// 刻意另開一個全新連線池,而非沿用上面的 client:同一個 pool 若先前服務過「有 scope」的
		// 請求,PG 會把 app.current_company_id 的 **session 值**留成空字串(SET LOCAL 對自訂 GUC
		// 會先在 session 層定義該變數,交易結束還原為 '' 而非 unset),policy 的 `::bigint` 便會以
		// 22P02 失敗而不是回 0 列 —— 兩者都 fail-closed,但只有全新連線是「乾淨的 NULL」語意,
		// 本子測試要釘的正是「沒有 scope 就看不到列」這個語意(實測見 report §5)。
		noScopeClient := dbtenant.NewClient(openAppRoleDB(t, adminDSN))
		t.Cleanup(func() { _ = noScopeClient.Close() })
		noScope := newCustomerAppRoleServer(t, noScopeClient, actorB, coB, false)
		res, err := noScope.ListCustomers(ctx, connect.NewRequest(&customersv1.ListCustomersRequest{Page: 1, PageSize: 50}))
		if err != nil {
			t.Fatalf("無 scope 的查詢本身應成功(只是看不到列): %v", err)
		}
		if n := len(res.Msg.GetCustomers()); n != 0 {
			t.Fatalf("無 scope 時不得看到任何客戶,得到 %d 筆", n)
		}
		if _, err := noScope.GetCustomer(ctx, connect.NewRequest(&customersv1.GetCustomerRequest{Id: itoa(custB)})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("無 scope 時連自己公司的客戶都應 not_found,got %v", err)
		}
		// 寫:B 尚無 counter 列,故無 scope 時的 INSERT 會直接被 customer_counters 的
		// WITH CHECK 擋下(company_id = current_company_id 不成立)。
		if _, err := noScope.CreateCustomer(ctx, connect.NewRequest(&customersv1.CreateCustomerRequest{
			Name: "無 scope 客戶", DefaultSalesRepId: itoa(actorB),
		})); err == nil {
			t.Fatal("無 RLS scope 時 CreateCustomer 必須失敗(寫入被 WITH CHECK 擋下)")
		}
		if n := countRows(t, admin, `SELECT count(*) FROM customers WHERE company_id = $1`, coB); n != 1 {
			t.Fatalf("被擋下的建檔不得落地:B 應仍只有 1 筆客戶,得到 %d", n)
		}
		if n := countRows(t, admin, `SELECT count(*) FROM customer_counters WHERE company_id = $1`, coB); n != 0 {
			t.Fatalf("被擋下的建檔不得建立 counter 列,得到 %d", n)
		}
	})
}

// newCustomerAppRoleServer 以業務 client(app_rw ＋ dbtenant.NewClient)掛載真客戶域 handler:
// RegisterCustomerServices 內含 dbtenant.HandlerOption → 每個 RPC 都在請求交易內執行。
// withScope=false 模擬「漏掛 RLS scope」的情境(負向對照);生產的 scope 由 authzMiddleware
// 依身分導出(server.go 的 identityFor → auth.WithRLS)。
func newCustomerAppRoleServer(t *testing.T, client *ent.Client, actor, companyID int, withScope bool) customersv1connect.CustomerServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterCustomerServices(mux, client, "http://localhost:3000")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), authz.Identity{
			UserID: itoa(actor), CompanyID: itoa(companyID), Role: "company_admin", Roles: []string{"company_admin"},
		})
		ctx = authz.WithDB(ctx, client)
		if withScope {
			ctx = auth.WithRLS(ctx, auth.RLSScope{
				UserID: itoa(actor), CompanyID: itoa(companyID), DataScope: auth.DataScopeCompany, CompanyActive: true,
			})
		}
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL)
}

// setCompanyCodePrefix 設定公司的客戶編號前綴(CreateCustomer 取號必要;缺前綴會回 failed_precondition)。
func setCompanyCodePrefix(t *testing.T, db *sql.DB, companyID int, prefix string) {
	t.Helper()
	if _, err := db.Exec(`UPDATE companies SET customer_code_prefix = $1 WHERE id = $2`, prefix, companyID); err != nil {
		t.Fatalf("設定公司 %d 的客戶編號前綴: %v", companyID, err)
	}
}

// countRows 以 admin 連線取單一純量真值(superuser 不受 RLS 影響,故看得到「實際落地」的列)。
func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("查詢 %q: %v", query, err)
	}
	return n
}

// assertCounterNextSeq 斷言某公司 counter 的下一個序號(CreateCustomer 已在請求交易內建列並遞增)。
func assertCounterNextSeq(t *testing.T, db *sql.DB, companyID, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT next_seq FROM customer_counters WHERE company_id = $1`, companyID).Scan(&got); err != nil {
		t.Fatalf("查公司 %d 的 counter: %v", companyID, err)
	}
	if got != want {
		t.Fatalf("公司 %d 的 counter next_seq 應為 %d,得到 %d", companyID, want, got)
	}
}

// insertRLSUser 以 admin 連線建一位使用者(供身分與稽核 FK)並回傳 id。
func insertRLSUser(t *testing.T, db *sql.DB, companyID int, email, role string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, password_hash, company_users)
		 VALUES ($1, '公司管理員', $2, 'x', $3) RETURNING id`, email, role, companyID).Scan(&id); err != nil {
		t.Fatalf("建使用者 %s: %v", email, err)
	}
	return id
}

// TestIntegrationRLSCustomersEnableMigrationDown 00024 的 Down 必須對稱還原(先 NO FORCE 再
// DISABLE),且可重複套用。少寫 `NO FORCE` 會留下 FORCE 旗標(owner 仍被擋)、少寫 `DISABLE`
// 則 RLS 仍生效 —— 兩者都讓回退後的環境與 00023 的狀態不一致,而 DISABLE 之後的行為差異
// 只有旗標看得出來(資料在兩種情況下都可見),故直接斷言 pg_class 的兩個旗標。
func TestIntegrationRLSCustomersEnableMigrationDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	admin := openRawDB(t, dsn)
	defer func() { _ = admin.Close() }()
	assertCustomerRLSFlags(t, admin, true, true)

	migrateBusinessDownTo(t, dsn, "23") // 僅回退 00024(00023 保留)
	assertCustomerRLSFlags(t, admin, false, false)

	migrateBusinessUp(t, dsn) // 回退後必須能重新套用(Up 冪等)
	assertCustomerRLSFlags(t, admin, true, true)
}

// assertCustomerRLSFlags 斷言客戶域四表的 ENABLE/FORCE 旗標。
func assertCustomerRLSFlags(t *testing.T, db *sql.DB, enabled, forced bool) {
	t.Helper()
	for _, table := range []string{"customers", "customer_counters", "customer_addresses", "customer_contacts"} {
		var e, f bool
		if err := db.QueryRow(
			`SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = $1::regclass`, table).Scan(&e, &f); err != nil {
			t.Fatalf("查 %s 的 RLS 旗標: %v", table, err)
		}
		if e != enabled || f != forced {
			t.Fatalf("%s 的 RLS 旗標應為 ENABLE=%v FORCE=%v,got ENABLE=%v FORCE=%v", table, enabled, forced, e, f)
		}
	}
}

// openAppRoleDB 以 app_rw（00022 的非 owner 業務角色）連線，回傳連線池。
func openAppRoleDB(t *testing.T, adminDSN string) *sql.DB {
	t.Helper()
	app, err := sql.Open("pgx", testsupport.AppRoleDSN(t, adminDSN))
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })
	return app
}

// setAppScope 於交易內套用公司範圍（SET LOCAL 不吃參數佔位，故以字串拼接；
// 值來自測試 fixture 的整數，非使用者輸入）。
func setAppScope(t *testing.T, tx *sql.Tx, companyID int) {
	t.Helper()
	for _, stmt := range []string{
		`SET LOCAL app.current_data_scope = 'company'`,
		`SET LOCAL app.current_company_id = '` + itoa(companyID) + `'`,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
}

// itoa 為本檔（同套件其他 RLS 整合測試共用，T6–T9 的探針亦沿用）的整數字串轉換輔助：
// SET LOCAL 的參數無法以 $1 綁定，故以字串拼接（值來自測試 fixture 的整數，非使用者輸入）。
func itoa(i int) string { return strconv.Itoa(i) }
