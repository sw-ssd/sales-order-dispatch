package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
)

// newAddressContactClient 於同一個 enttest DB 建立公司+部門+rep+客戶,並以該 DB
// 掛上 dept_admin 身分的 CustomerService server,回傳 (client, coID, deptID, customerID)。
func newAddressContactClient(t *testing.T) (customersv1connect.CustomerServiceClient, int, int, int) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	co, err := db.Company.Create().SetName("客戶公司").SetIdentifier("AC-" + t.Name()).SetStatus("active").SetCustomerCodePrefix("AC").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	d, err := db.Department.Create().SetCompanyID(co.ID).SetName("門市一").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	if _, err := db.User.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).
		SetEmail("ac-rep-" + t.Name() + "@t.com").SetName("業務").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
		t.Fatalf("rep: %v", err)
	}
	c, err := db.Customer.Create().SetCompanyID(co.ID).SetDepartmentID(d.ID).SetCustomerCode("AC000001").SetName("王小明").Save(ctx)
	if err != nil {
		t.Fatalf("customer: %v", err)
	}
	id := deptAdminID(co.ID, d.ID)
	mux := http.NewServeMux()
	RegisterCustomerServices(mux, db, "http://localhost:3000")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cctx := authz.WithIdentity(r.Context(), id)
		cctx = authz.WithDB(cctx, db)
		cctx = audit.WithMeta(cctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		mux.ServeHTTP(w, r.WithContext(cctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL), co.ID, d.ID, c.ID
}

// TestAddressFirstAutoDefault 3.2.1:同類型首筆自動為預設;shipping 與 billing 各自獨立。
func TestAddressFirstAutoDefault(t *testing.T) {
	ctx := context.Background()
	client, _, _, custID := newAddressContactClient(t)

	// 首筆 shipping 未指定 default → 自動為預設。
	resp, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: uItoa(custID), Type: "shipping", RecipientName: "王小明", AddressLine: "台北市大安區",
	}))
	if err != nil {
		t.Fatalf("AddAddress: %v", err)
	}
	if !resp.Msg.GetAddress().GetIsDefault() {
		t.Fatal("同類型首筆地址應自動為預設")
	}
	// 第二筆 shipping(未指定)→ 不為預設。
	resp2, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: uItoa(custID), Type: "shipping", RecipientName: "王小明", AddressLine: "台中市西屯區",
	}))
	if err != nil {
		t.Fatalf("AddAddress2: %v", err)
	}
	if resp2.Msg.GetAddress().GetIsDefault() {
		t.Fatal("第二筆 shipping 不應自動為預設")
	}
	// 新增 billing 首筆 → 自動為預設(非 shipping)。
	resp3, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: uItoa(custID), Type: "billing", RecipientName: "王小明", AddressLine: "桃園市中壢區",
	}))
	if err != nil {
		t.Fatalf("AddAddressBilling: %v", err)
	}
	if !resp3.Msg.GetAddress().GetIsDefault() {
		t.Fatal("billing 首筆應自動為預設")
	}
	// 列表:shipping 恰一筆預設、billing 恰一筆預設。
	list, err := client.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: uItoa(custID)}))
	if err != nil {
		t.Fatalf("ListAddresses: %v", err)
	}
	var shipDefault, billDefault int
	for _, a := range list.Msg.GetAddresses() {
		if a.GetIsDefault() && a.GetType() == "shipping" {
			shipDefault++
		}
		if a.GetIsDefault() && a.GetType() == "billing" {
			billDefault++
		}
	}
	if shipDefault != 1 || billDefault != 1 {
		t.Fatalf("shipping/billing 應各恰一筆預設,got ship=%d bill=%d", shipDefault, billDefault)
	}
}

// TestAddressSetDefaultClearsOthers 3.2.1:設定某筆為預設後,同類型其餘自動失去預設(先清後設)。
func TestAddressSetDefaultClearsOthers(t *testing.T) {
	ctx := context.Background()
	client, _, _, custID := newAddressContactClient(t)
	mustAdd := func(line string) string {
		r, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
			CustomerId: uItoa(custID), Type: "shipping", RecipientName: "王", AddressLine: line}))
		if err != nil {
			t.Fatalf("AddAddress(%s): %v", line, err)
		}
		return r.Msg.GetAddress().GetId()
	}
	first := mustAdd("地址A") // 自動預設
	second := mustAdd("地址B")
	_ = first
	// 設定第二筆為預設 → 第一筆失去預設。
	if _, err := client.UpdateAddress(ctx, connect.NewRequest(&customersv1.UpdateAddressRequest{
		Id: second, IsDefault: boolPtr(true)})); err != nil {
		t.Fatalf("UpdateAddress set default: %v", err)
	}
	list, err := client.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: uItoa(custID)}))
	if err != nil {
		t.Fatalf("ListAddresses: %v", err)
	}
	defaultCount := 0
	defaultID := ""
	for _, a := range list.Msg.GetAddresses() {
		if a.GetType() == "shipping" && a.GetIsDefault() {
			defaultCount++
			defaultID = a.GetId()
		}
	}
	if defaultCount != 1 || defaultID != second {
		t.Fatalf("設定預設後應恰一筆 shipping 預設且為第二筆,got count=%d id=%s(期望 %s)", defaultCount, defaultID, second)
	}
}

// TestAddressInvalidTypeAndDelete 3.2.1:非法 type → invalid_argument;刪除後預設列表不再出現。
func TestAddressInvalidTypeAndDelete(t *testing.T) {
	ctx := context.Background()
	client, _, _, custID := newAddressContactClient(t)

	if _, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: uItoa(custID), Type: "carrier", RecipientName: "王", AddressLine: "x"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法 type 應 invalid_argument,got %v", err)
	}
	if _, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: uItoa(custID), Type: "shipping"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺必填應 invalid_argument,got %v", err)
	}
	r, err := client.AddAddress(ctx, connect.NewRequest(&customersv1.AddAddressRequest{
		CustomerId: uItoa(custID), Type: "shipping", RecipientName: "王", AddressLine: "地址C"}))
	if err != nil {
		t.Fatalf("AddAddress: %v", err)
	}
	if _, err := client.DeleteAddress(ctx, connect.NewRequest(&customersv1.DeleteAddressRequest{Id: r.Msg.GetAddress().GetId()})); err != nil {
		t.Fatalf("DeleteAddress: %v", err)
	}
	list, err := client.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: uItoa(custID)}))
	if err != nil {
		t.Fatalf("ListAddresses: %v", err)
	}
	if len(list.Msg.GetAddresses()) != 0 {
		t.Fatalf("刪除後列表應為空,got %d", len(list.Msg.GetAddresses()))
	}
}

// TestContactFirstAutoDefaultAndSetDefault 3.2.2:首筆自動預設;設第二筆為預設清第一筆;email 驗證。
func TestContactFirstAutoDefaultAndSetDefault(t *testing.T) {
	ctx := context.Background()
	client, _, _, custID := newAddressContactClient(t)

	r1, err := client.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{CustomerId: uItoa(custID), Name: "王小明"}))
	if err != nil {
		t.Fatalf("AddContact: %v", err)
	}
	if !r1.Msg.GetContact().GetIsDefault() {
		t.Fatal("首筆聯絡人應自動為預設")
	}
	r2, err := client.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{CustomerId: uItoa(custID), Name: "李四"}))
	if err != nil {
		t.Fatalf("AddContact2: %v", err)
	}
	if r2.Msg.GetContact().GetIsDefault() {
		t.Fatal("第二筆聯絡人不應自動為預設")
	}
	// 設第二筆為預設 → 第一筆失去。
	if _, err := client.UpdateContact(ctx, connect.NewRequest(&customersv1.UpdateContactRequest{Id: r2.Msg.GetContact().GetId(), IsDefault: boolPtr(true)})); err != nil {
		t.Fatalf("UpdateContact: %v", err)
	}
	list, err := client.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: uItoa(custID)}))
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	var defID string
	for _, c := range list.Msg.GetContacts() {
		if c.GetIsDefault() {
			defID = c.GetId()
		}
	}
	if defID != r2.Msg.GetContact().GetId() {
		t.Fatalf("預設聯絡人應為第二筆,got %s(期望 %s)", defID, r2.Msg.GetContact().GetId())
	}

	// email 格式驗證。
	if _, err := client.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{CustomerId: uItoa(custID), Name: "王", Email: "bad-email"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法 email 應 invalid_argument,got %v", err)
	}
	// name 必填。
	if _, err := client.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{CustomerId: uItoa(custID)})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺 name 應 invalid_argument,got %v", err)
	}
}

// TestContactDelete 3.2.2:軟刪除聯絡人不再出現於列表。
func TestContactDelete(t *testing.T) {
	ctx := context.Background()
	client, _, _, custID := newAddressContactClient(t)
	r, err := client.AddContact(ctx, connect.NewRequest(&customersv1.AddContactRequest{CustomerId: uItoa(custID), Name: "王小明"}))
	if err != nil {
		t.Fatalf("AddContact: %v", err)
	}
	if _, err := client.DeleteContact(ctx, connect.NewRequest(&customersv1.DeleteContactRequest{Id: r.Msg.GetContact().GetId()})); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}
	list, err := client.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: uItoa(custID)}))
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	if len(list.Msg.GetContacts()) != 0 {
		t.Fatalf("刪除後聯絡人列表應為空,got %d", len(list.Msg.GetContacts()))
	}
}

// TestAddressContactUnauthenticated 未登入 → unauthenticated。
func TestAddressContactUnauthenticated(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	mux := http.NewServeMux()
	RegisterCustomerServices(mux, db, "http://localhost:3000")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cctx := authz.WithIdentity(r.Context(), authz.Identity{})
		cctx = authz.WithDB(cctx, db)
		mux.ServeHTTP(w, r.WithContext(cctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	client := customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL)
	if _, err := client.ListAddresses(ctx, connect.NewRequest(&customersv1.ListAddressesRequest{CustomerId: "1"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("未登入 ListAddresses 應 unauthenticated,got %v", err)
	}
	if _, err := client.ListContacts(ctx, connect.NewRequest(&customersv1.ListContactsRequest{CustomerId: "1"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("未登入 ListContacts 應 unauthenticated,got %v", err)
	}
}

// boolPtr 回傳 bool 指標(proto optional bool)。
func boolPtr(b bool) *bool { return &b }
