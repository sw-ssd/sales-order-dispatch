package auth

import (
	"reflect"
	"testing"
)

// TestBuiltinRolePermissionsDeterministic 驗證種子集合決定性(排序)、去重與關鍵內容。
func TestBuiltinRolePermissionsDeterministic(t *testing.T) {
	first := BuiltinRolePermissions()
	second := BuiltinRolePermissions()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("BuiltinRolePermissions 應決定性(兩次呼叫結果相同)")
	}
	if len(first) == 0 {
		t.Fatal("種子集合不應為空")
	}
	// 排序不變式:依 (role, resource, action) 升冪。
	for i := 1; i < len(first); i++ {
		a, b := first[i-1], first[i]
		if a.Role > b.Role ||
			(a.Role == b.Role && a.Resource > b.Resource) ||
			(a.Role == b.Role && a.Resource == b.Resource && a.Action > b.Action) {
			t.Fatalf("種子未排序: %+v > %+v", a, b)
		}
	}
	// 去重:無重複 (role,resource,action)。
	seen := map[string]bool{}
	for _, s := range first {
		k := s.Role + "\x00" + s.Resource + "\x00" + s.Action
		if seen[k] {
			t.Fatalf("重複種子 %s/%s/%s", s.Role, s.Resource, s.Action)
		}
		seen[k] = true
	}
}

// TestBuiltinRolePermissionsContent 驗證各內建角色關鍵權限內容。
func TestBuiltinRolePermissionsContent(t *testing.T) {
	seeds := BuiltinRolePermissions()
	has := func(role, res, act string) bool {
		for _, s := range seeds {
			if s.Role == role && s.Resource == res && s.Action == act {
				return true
			}
		}
		return false
	}
	count := func(role string) int {
		n := 0
		for _, s := range seeds {
			if s.Role == role {
				n++
			}
		}
		return n
	}
	// super/developer = 全資源 × read/write。
	if got := count("super"); got != len(adminResources)*2 {
		t.Errorf("super 種子數 = %d, want %d", got, len(adminResources)*2)
	}
	if got := count("developer"); got != len(adminResources)*2 {
		t.Errorf("developer 種子數 = %d, want %d", got, len(adminResources)*2)
	}
	// company_admin 具名動作(company update)+ 萬用資源展開(read/write)。
	if !has("company_admin", "company", "read") || !has("company_admin", "company", "update") {
		t.Error("company_admin 缺少 company read/update")
	}
	if !has("company_admin", "user", "read") || !has("company_admin", "user", "write") {
		t.Error("company_admin 缺少 user read/write(萬用展開)")
	}
	// 繼承:company_admin 取得 staff 的 accounting read、customer 的 sales_order read。
	if !has("company_admin", "accounting", "read") {
		t.Error("company_admin 應經繼承取得 accounting read")
	}
	// guest 僅 user:read。
	if !has("guest", "user", "read") {
		t.Error("guest 缺少 user read")
	}
	if has("guest", "sales_order", "read") {
		t.Error("guest 不應有 sales_order read")
	}
	// customer 具名。
	if !has("customer", "sales_order", "read") || !has("customer", "product", "read") {
		t.Error("customer 缺少 sales_order/product read")
	}
	// 退貨頁守衛(requireAbility("read","return_request"))的受眾契約:
	// 員工看得到清單、staff 另有 write 才能審(canReview 再收斂到該客戶主責業務);
	// customer 刻意不給 —— 客戶自助走 App,Web 端維持 403 指引(D25)。
	if !has("staff", "return_request", "read") || !has("staff", "return_request", "write") {
		t.Error("staff 缺少 return_request read/write(可看可審)")
	}
	if !has("dept_admin", "return_request", "read") || !has("company_admin", "return_request", "read") {
		t.Error("dept_admin/company_admin 缺少 return_request read")
	}
	if has("customer", "return_request", "read") {
		t.Error("customer 不應有 return_request read(Web 端維持 403)")
	}
	// 通知中心(/notifications)同一批契約:員工可讀可標已讀,customer 不給(Web 403,
	// 客戶通知走 App;NotificationService 本身只回 user_id = 本人的列)。
	if !has("staff", "notification", "read") || !has("staff", "notification", "write") {
		t.Error("staff 缺少 notification read/write(可讀清單、可標已讀)")
	}
	if has("customer", "notification", "read") {
		t.Error("customer 不應有 notification read(Web 端維持 403)")
	}
}
