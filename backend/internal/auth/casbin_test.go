package auth

import (
	"reflect"
	"testing"
)

// TestEnforceBuiltinRoles 驗收 brief Step 5:各內建角色 allowed/denied 動作。
func TestEnforceBuiltinRoles(t *testing.T) {
	dom := "c1"

	allowed := map[string][][3]string{ // role → [][obj, act, dom]
		"super":         {{"company", "delete", dom}, {"sales_order", "cancel", dom}, {"anything", "manage", dom}},
		"developer":     {{"company", "delete", dom}, {"user", "manage", dom}},
		"company_admin": {{"company", "read", dom}, {"company", "update", dom}, {"department", "delete", dom}, {"user", "update", dom}, {"role", "update", dom}, {"sales_order", "*", dom}},
		"dept_admin":    {{"user", "read", dom}, {"customer", "update", dom}, {"sales_order", "delete", dom}, {"dispatch", "create", dom}},
		"staff":         {{"customer", "create", dom}, {"sales_order", "read", dom}, {"print", "print", dom}, {"accounting", "read", dom}},
		"customer":      {{"sales_order", "read", dom}, {"product", "read", dom}},
		"guest":         {{"user", "read", dom}},
	}
	for role, cases := range allowed {
		for _, c := range cases {
			ok, err := Enforce(role, c[0], c[1], c[2])
			if err != nil {
				t.Fatalf("Enforce(%s,%s,%s,%s): %v", role, c[0], c[1], c[2], err)
			}
			if !ok {
				t.Errorf("Enforce(%s,%s,%s,%s) = false, want true", role, c[0], c[1], c[2])
			}
		}
	}

	denied := map[string][][3]string{
		"company_admin": {{"company", "create", dom}, {"company", "delete", dom}, {"billing", "read", dom}}, // 公司 CRUD 僅 super(8.2);未知資源 deny
		"dept_admin":    {{"user", "update", dom}, {"company", "read", dom}},                                // 不可管理帳號/公司
		"staff":         {{"user", "read", dom}, {"company", "read", dom}, {"dispatch", "create", dom}},     // 不能管理帳號；dispatch 唯讀
		"customer":      {{"sales_order", "update", dom}, {"product", "update", dom}, {"user", "read", dom}},
		"guest":         {{"sales_order", "read", dom}, {"user", "update", dom}},
	}
	for role, cases := range denied {
		for _, c := range cases {
			ok, err := Enforce(role, c[0], c[1], c[2])
			if err != nil {
				t.Fatalf("Enforce(%s,%s,%s,%s): %v", role, c[0], c[1], c[2], err)
			}
			if ok {
				t.Errorf("Enforce(%s,%s,%s,%s) = true, want false", role, c[0], c[1], c[2])
			}
		}
	}

	// 未知角色一律 deny(fail-closed)。
	ok, err := Enforce("unknown_role", "sales_order", "read", dom)
	if err != nil {
		t.Fatalf("Enforce: %v", err)
	}
	if ok {
		t.Error("未知角色應 deny")
	}
}

// TestEnforceInheritance 驗收 g 繼承:子角色取得父角色權限。
func TestEnforceInheritance(t *testing.T) {
	// dept_admin 經 g 繼承 staff 的 accounting read;company_admin 繼承 dept_admin 的 dispatch create。
	ok, err := Enforce("dept_admin", "accounting", "read", "c1")
	if err != nil || !ok {
		t.Fatalf("dept_admin 應繼承 staff 的 accounting read: ok=%v err=%v", ok, err)
	}
	ok, err = Enforce("company_admin", "dispatch", "create", "c1")
	if err != nil || !ok {
		t.Fatalf("company_admin 應繼承 dept_admin 的 dispatch create: ok=%v err=%v", ok, err)
	}
}

// TestEnforceAny 多角色任一允許即通過;全拒才 false。
func TestEnforceAny(t *testing.T) {
	ok, err := EnforceAny([]string{"guest", "staff"}, "sales_order", "read", "c1")
	if err != nil || !ok {
		t.Fatalf("EnforceAny(staff) = %v, want true; err=%v", ok, err)
	}
	ok, err = EnforceAny([]string{"guest", "customer"}, "company", "delete", "c1")
	if err != nil || ok {
		t.Fatalf("EnforceAny 全拒應為 false; ok=%v err=%v", ok, err)
	}
	ok, err = EnforceAny(nil, "sales_order", "read", "c1")
	if err != nil || ok {
		t.Fatalf("EnforceAny(nil) 應為 false; ok=%v err=%v", ok, err)
	}
}

// TestRolesFor g 展開(含自身,去重保序)。
func TestRolesFor(t *testing.T) {
	got := RolesFor("company_admin")
	want := []string{"company_admin", "dept_admin", "staff", "customer"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RolesFor(company_admin) = %v, want %v", got, want)
	}
	got = RolesFor("staff")
	want = []string{"staff", "customer"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RolesFor(staff) = %v, want %v", got, want)
	}
	got = RolesFor("super")
	if !reflect.DeepEqual(got, []string{"super"}) {
		t.Fatalf("RolesFor(super) = %v, want [super]", got)
	}
	// 未知角色回傳自身(fail-closed 仍可列)。
	if got := RolesFor("custom_role"); !reflect.DeepEqual(got, []string{"custom_role"}) {
		t.Fatalf("RolesFor(custom_role) = %v, want [custom_role]", got)
	}
}
