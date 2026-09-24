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
	// customer 具名:App 自助三資源(訂單讀寫/退貨讀寫/本人通知讀寫),
	// 對應規格 sales-orders 4.2.3(客戶自行下單)、4.4(客戶 App 發起退貨)、
	// notifications(退貨審核結果推播客戶子帳號)。
	if !has("customer", "sales_order", "read") || !has("customer", "sales_order", "write") {
		t.Error("customer 缺少 sales_order read/write(App 查單/建單/取消)")
	}
	if !has("customer", "product", "read") {
		t.Error("customer 缺少 product read")
	}
	// 退貨頁守衛(requireAbility("read","return_request"))的受眾契約:
	// 員工看得到清單、staff 另有 write 才能審(canReview 再收斂到該客戶主責業務);
	// customer 有讀寫走 App 自助(查看自己申請+發起;Web 無建單入口,審核權仍由
	// canReview 收斂,不受此影響)。
	if !has("staff", "return_request", "read") || !has("staff", "return_request", "write") {
		t.Error("staff 缺少 return_request read/write(可看可審)")
	}
	if !has("dept_admin", "return_request", "read") || !has("company_admin", "return_request", "read") {
		t.Error("dept_admin/company_admin 缺少 return_request read")
	}
	if !has("customer", "return_request", "read") || !has("customer", "return_request", "write") {
		t.Error("customer 缺少 return_request read/write(App 查看/發起退貨)")
	}
	// 通知中心同一批契約:員工可讀可標已讀;customer 亦讀寫 —— NotificationService
	// 只回 user_id = 本人的列,App 通知中心是客戶收單建立/退貨審核通知的唯一入口。
	if !has("staff", "notification", "read") || !has("staff", "notification", "write") {
		t.Error("staff 缺少 notification read/write(可讀清單、可標已讀)")
	}
	if !has("customer", "notification", "read") || !has("customer", "notification", "write") {
		t.Error("customer 缺少 notification read/write(App 通知中心)")
	}
	// 稽核頁守衛：範圍與 AuditService.ListAuditLogs 一致 —— company_admin 可查、
	// dept_admin/staff/customer 一律不可（後端 default 分支直接 PermissionDenied）。
	if !has("company_admin", "audit_log", "read") {
		t.Error("company_admin 缺少 audit_log read")
	}
	if has("dept_admin", "audit_log", "read") || has("staff", "audit_log", "read") {
		t.Error("dept_admin/staff 不應有 audit_log read（後端拒絕非 company_admin 查詢）")
	}
}
