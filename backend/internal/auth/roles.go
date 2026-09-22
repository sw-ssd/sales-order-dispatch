// Package auth 的內建角色 ACL(純 Go,取代 Casbin engine——D32 移除 Casbin 依賴)。
// role_permissions 表為角色/功能權限定義來源(資料驅動);此處僅提供內建角色的
// 「預設」角色→資源→動作 ACL(對齊原 rbac_policy.csv)與角色繼承展開(RolesFor),
// 供服務層 requireRole/requireScope 作為 fallback 閘門。生產授權決策由 OpenFGA Check
// (middleware 閘門)執行;RLS 承擔資料範圍隔離(company/department/self)。
package auth

import (
	"cmp"
	"slices"
)

// BuiltinRoles 內建 7 角色(設計書 §3.2 + §4.4),系統 seed 值。
var BuiltinRoles = []string{"super", "company_admin", "dept_admin", "staff", "customer", "guest", "developer"}

// rolePolicy 內建角色 → 資源 → 允許動作("" = 該資源全部動作;"*" 資源 = 全部資源)。
// 對齊原 rbac_policy.csv 的 7 內建角色預設 policy(設計書 §3.2 + §4.4)。
var rolePolicy = map[string]map[string][]string{
	"super":     {"*": {"*"}},
	"developer": {"*": {"*"}},
	"company_admin": {
		"company":        {"read", "update"},
		"department":     {"*"},
		"user":           {"*"},
		"role":           {"*"},
		"sales_order":    {"*"},
		"customer":       {"*"},
		"product":        {"*"},
		"print":          {"*"},
		"dispatch":       {"*"},
		"return_request": {"*"},
		"notification":   {"*"},
		// 稽核：只有 company_admin（＋ super/developer 萬用展開）可查 —— 與
		// AuditService.ListAuditLogs 的範圍推導一致（dept_admin/staff 一律 PermissionDenied）。
		"audit_log": {"read"},
	},
	"dept_admin": {
		"department":     {"read"},
		"user":           {"read"},
		"customer":       {"*"},
		"product":        {"*"},
		"sales_order":    {"*"},
		"print":          {"*"},
		"dispatch":       {"*"},
		"return_request": {"*"},
		"notification":   {"*"},
	},
	"staff": {
		"customer":    {"*"},
		"product":     {"*"},
		"sales_order": {"*"},
		"print":       {"*"},
		"dispatch":    {"read"},
		"accounting":  {"read"},
		// 退貨:staff 可看、可審 —— 審核權再由服務層收斂到「該客戶主責業務」
		// (return_review.go canReview);此處只給類別,不給個別客戶的判斷。
		"return_request": {"read", "write"},
		// 通知中心是**本人**的通知(NotificationService 只回 user_id = 自己的列),
		// 權限資源只管「能不能開這頁」;read 讀清單、write 標記已讀。
		"notification": {"read", "write"},
	},
	"customer": {
		"sales_order": {"read"},
		"product":     {"read"},
	},
	"guest": {
		"user": {"read"},
	},
}

// roleInheritance 角色繼承(子角色取得父角色的全部權限;對齊原 rbac_policy.csv 的 g 規則)。
var roleInheritance = map[string][]string{
	"company_admin": {"dept_admin"},
	"dept_admin":    {"staff"},
	"staff":         {"customer"},
	"developer":     {"super"},
}

// effectiveRoles 回傳角色展開集:自身 + 沿继承链的間接角色(去重保序)。
func effectiveRoles(role string) []string {
	out := []string{role}
	queue := []string{role}
	seen := map[string]bool{role: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range roleInheritance[cur] {
			if !seen[child] {
				seen[child] = true
				out = append(out, child)
				queue = append(queue, child)
			}
		}
	}
	return out
}

// Enforce 判斷角色 code 對資源 obj 執行動作 act 是否允許(純 Go ACL;domain 由 RLS 承擔,忽略)。
// 任一繼承角色命中即允許。
func Enforce(role, obj, act, _ string) (bool, error) {
	for _, r := range effectiveRoles(role) {
		if enforceRole(r, obj, act) {
			return true, nil
		}
	}
	return false, nil
}

// EnforceAny 對角色清單逐一 Enforce,任一允許即通過(fail-closed:全拒才 false)。
func EnforceAny(roles []string, obj, act, dom string) (bool, error) {
	for _, r := range roles {
		ok, err := Enforce(r, obj, act, dom)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// RolesFor 展開角色 code 的隱含角色集(依繼承鏈,含自身,去重保序)。
// 供 middleware 組 identity.Roles:載入各角色 role_permissions 規則與 fallback ACL。
func RolesFor(role string) []string {
	return effectiveRoles(role)
}

// enforceRole 判斷單一角色(不含繼承)是否允許 obj+act。
func enforceRole(role, obj, act string) bool {
	perms, ok := rolePolicy[role]
	if !ok {
		return false
	}
	if acts, ok := perms["*"]; ok {
		return actionAllowed(acts, act)
	}
	acts, ok := perms[obj]
	if !ok {
		return false
	}
	return actionAllowed(acts, act)
}

// actionAllowed 判斷 act 是否在允許列表(含 "*" 全動作)。
func actionAllowed(acts []string, act string) bool {
	for _, a := range acts {
		if a == "*" || a == act {
			return true
		}
	}
	return false
}

// adminResources 為「全資源」內建角色( super / developer )萬用展開的具名業務資源集合。
// OpenFGA ability:<res> 為資料驅動、無全域萬用(見 third_party/openfga/model.go),
// 故以固定業務資源集種出 read/write,涵蓋受保護 RPC 與服務層檢查之資源。
var adminResources = []string{
	"company", "department", "user", "role",
	"sales_order", "customer", "product", "print", "dispatch", "accounting",
	"return_request", "notification", "audit_log",
}

// PermissionSeed 描述單一 role_permissions 種子列( role code → resource → action )。
// Resource/Action 直接寫入 role_permissions(resource,action)欄;Action 用 read/write 兩類
// 覆蓋 OpenFGA 的 can_read/can_write( permissionRelation 對映),具名動作(如 update)原樣保留。
type PermissionSeed struct {
	Role     string
	Resource string
	Action   string
}

// BuiltinRolePermissions 由單一來源 rolePolicy+roleInheritance(本檔)展開 7 內建角色的
// 「有效(繼承後)」權限種子;供 cmd/seed 冪等寫入 role_permissions,使 DB 成為唯一持久來源、
// OpenFGA provision 得以產出 role→ability tuples( D32 修訂,消除雙來源漂移 )。
// "*"(全資源/全動作)展開為固定業務資源 × read/write。回傳以 (role,resource,action) 排序,
// 決定性次序( map 迭代不保證順序,測試與寫入需穩定排序)。
func BuiltinRolePermissions() []PermissionSeed {
	seen := map[string]bool{} // (role,resource,action) 去重
	var seeds []PermissionSeed
	add := func(role, res, act string) {
		k := role + "\x00" + res + "\x00" + act
		if !seen[k] {
			seen[k] = true
			seeds = append(seeds, PermissionSeed{Role: role, Resource: res, Action: act})
		}
	}
	for _, role := range BuiltinRoles {
		// 展開繼承鏈(子角色取得父角色全部權限,對齊 effectiveRoles)。
		for _, r := range effectiveRoles(role) {
			perms := rolePolicy[r]
			if allRes := perms["*"]; len(allRes) > 0 && allRes[0] == "*" {
				// 全資源全動作(super/developer)→ 具名資源 × read/write。
				for _, res := range adminResources {
					add(role, res, "read")
					add(role, res, "write")
				}
				continue
			}
			for res, acts := range perms {
				if res == "*" {
					continue
				}
				for _, act := range acts {
					if act == "*" {
						add(role, res, "read")
						add(role, res, "write")
					} else {
						add(role, res, act)
					}
				}
			}
		}
	}
	// 決定性排序。
	slices.SortFunc(seeds, func(a, b PermissionSeed) int {
		if a.Role != b.Role {
			return cmp.Compare(a.Role, b.Role)
		}
		if a.Resource != b.Resource {
			return cmp.Compare(a.Resource, b.Resource)
		}
		return cmp.Compare(a.Action, b.Action)
	})
	return seeds
}
