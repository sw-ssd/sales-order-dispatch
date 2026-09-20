package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// errorInfoOf 由 connect error 取 ErrorInfo detail（本套件共用的唯一解析點：碼／訊息／
// details／trace_id 都在 ErrorInfo 裡，逐一自行解析只會出現兩份真相）。
func errorInfoOf(t *testing.T, err error) *commonv1.ErrorInfo {
	t.Helper()
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("應為 *connect.Error,got %T", err)
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			t.Fatalf("detail 取值失敗: %v", derr)
		}
		if info, ok := v.(*commonv1.ErrorInfo); ok {
			return info
		}
	}
	t.Fatal("錯誤未帶 ErrorInfo")
	return nil
}

// TestFirstBatchCodesAreReturned 逐條斷言首批碼真的被回傳（不是只註冊了沒用）。
//
// 這裡驗的是「RPC 的錯誤碼」這條契約：只登記碼不算落地，必須是有客戶端真的收到它。
// 「配額超限 → PLAT-5001」案例屬 Task 5b（需 internal/platform/*，本計畫執行時不存在）。
func TestFirstBatchCodesAreReturned(t *testing.T) {
	cases := []struct {
		name string
		want string
		call func(t *testing.T) error
	}{
		{"未登入", "AUTH-4001", func(t *testing.T) error {
			cc, _, _ := newTestServerWithIdentity(t, authz.Identity{})
			_, err := cc.ListCompanies(context.Background(), connect.NewRequest(&v1.ListCompaniesRequest{}))
			return err
		}},
		{"角色不足", "SYS-4001", func(t *testing.T) error {
			staff := authz.Identity{UserID: "9", CompanyID: "1", Role: "staff", Roles: []string{"staff"}}
			cc, _, _ := newTestServerWithIdentity(t, staff)
			_, err := cc.DeleteCompany(context.Background(), connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: "1"}))
			return err
		}},
		{"跨租戶查詢", "SYS-4002", func(t *testing.T) error {
			admin := authz.Identity{UserID: "1", CompanyID: "1", Role: "company_admin", Roles: []string{"company_admin"}}
			cc, _, _ := newTestServerWithIdentity(t, admin)
			// 以身分所屬範圍查他公司的 id：範圍過濾後查不到 → SYS-4002（不洩漏存在性）
			_, err := cc.GetCompany(context.Background(), connect.NewRequest(&v1.GetCompanyRequest{CompanyId: "999999"}))
			return err
		}},
		{"客戶名稱空白(建檔)", "CUST-1001", func(t *testing.T) error {
			_, db := newCustomerTestServer(t, authz.Identity{})
			coID, deptID := seedCustomerCompany(t, db, "TC", true)
			client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
			_, err := client.CreateCustomer(context.Background(),
				connect.NewRequest(&customersv1.CreateCustomerRequest{Name: "   "}))
			return err
		}},
		{"客戶名稱空白(更新)", "CUST-1001", func(t *testing.T) error {
			_, db := newCustomerTestServer(t, authz.Identity{})
			coID, deptID := seedCustomerCompany(t, db, "TF", true)
			c := db.Customer.Create().SetCompanyID(coID).SetDepartmentID(deptID).
				SetCustomerCode("TF000001").SetName("原名").SaveX(context.Background())
			client, _ := newCustomerTestServer(t, deptAdminID(coID, deptID))
			_, err := client.UpdateCustomer(context.Background(), connect.NewRequest(
				&customersv1.UpdateCustomerRequest{Id: uItoa(c.ID), Name: strPtr("   ")}))
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(t)
			if err == nil {
				t.Fatal("應回錯誤")
			}
			if got := errorInfoOf(t, err).GetCode(); got != tc.want {
				t.Fatalf("錯誤碼 = %q；want %s（訊息: %s）", got, tc.want, err.Error())
			}
		})
	}
}

// TestPermissionDeniedCarriesResourceAction：SYS-4001 的訊息樣板只有「缺少權限」，
// 前端要能提示「缺 company 的 delete 權限」只能靠 details（不得改樣板文字）。
func TestPermissionDeniedCarriesResourceAction(t *testing.T) {
	staff := authz.Identity{UserID: "9", CompanyID: "1", Role: "staff", Roles: []string{"staff"}}
	cc, _, _ := newTestServerWithIdentity(t, staff)
	_, err := cc.DeleteCompany(context.Background(), connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: "1"}))

	info := errorInfoOf(t, err)
	if info.GetDetails()["resource"] != "company" || info.GetDetails()["action"] != "delete" {
		t.Fatalf("details 應帶 resource=company／action=delete,got %v", info.GetDetails())
	}
}
