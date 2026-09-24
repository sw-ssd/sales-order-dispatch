package services

import (
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

// TestOrderScopeCustomerBranch 是 orderScope 的迴歸測試(2026-09-24 實機踩到):
//
// 「是不是客戶帳號」必須用原始角色 id.Role,不可用展開後的 hasRole(id,"customer")——
// 角色繼承是 company_admin → dept_admin → staff → customer,展開後**每個後台角色都含
// customer**,用展開集判斷會讓所有管理者誤入客戶分支(缺 CustomerID → 403 全站查不到訂單)。
// 既有整合測試手動組 Roles(不含 customer)看不到,故這裡以 RolesFor 的真實展開值驗證。
func TestOrderScopeCustomerBranch(t *testing.T) {
	const cid, did = 7, 3

	// 後台角色:展開的 Roles 實際含 customer,但仍必須走部門/公司分支。
	for _, role := range []string{"staff", "dept_admin", "company_admin"} {
		id := authz.Identity{
			UserID: "1", CompanyID: "7", DepartmentID: "3",
			Role: role, Roles: auth.RolesFor(role),
		}
		gotCID, gotDID, gotCust, err := orderScope(id)
		if err != nil {
			t.Fatalf("%s: orderScope 不應失敗(%v);展開 Roles=%v", role, err, id.Roles)
		}
		if gotCID != cid {
			t.Fatalf("%s: company 應為 %d,got %d", role, cid, gotCID)
		}
		if gotCust != nil {
			t.Fatalf("%s: 不應被當成客戶身分(custID=%v)", role, gotCust)
		}
		if role == "staff" || role == "dept_admin" {
			if gotDID == nil || *gotDID != did {
				t.Fatalf("%s: 應限本部門 %d,got %v", role, did, gotDID)
			}
		} else if gotDID != nil {
			t.Fatalf("%s: 公司層不應有部門限制,got %v", role, *gotDID)
		}
	}

	// 真客戶帳號:仍走客戶分支(僅自己客戶),且必須帶 CustomerID。
	cust := authz.Identity{
		UserID: "9", CompanyID: "7", CustomerID: "42",
		Role: "customer", Roles: auth.RolesFor("customer"),
	}
	gotCID, gotDID, gotCust, err := orderScope(cust)
	if err != nil {
		t.Fatalf("customer: orderScope 不應失敗: %v", err)
	}
	if gotCID != cid || gotDID != nil || gotCust == nil || *gotCust != 42 {
		t.Fatalf("customer 範圍不符: cid=%d did=%v cust=%v", gotCID, gotDID, gotCust)
	}

	// 客戶帳號缺 CustomerID(身分未填)→ 拒絕,不是靜默變成公司層。
	noCust := authz.Identity{UserID: "9", CompanyID: "7", Role: "customer", Roles: auth.RolesFor("customer")}
	if _, _, _, err := orderScope(noCust); err == nil {
		t.Fatal("客戶帳號缺 CustomerID 應拒絕")
	} else if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("應 permission_denied,got %v", err)
	}
}
