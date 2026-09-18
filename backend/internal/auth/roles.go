// Package auth 的內建角色 ACL(純 Go,取代 Casbin engine——D32 移除 Casbin 依賴)。
// role_permissions 表為角色/功能權限定義來源(資料驅動);此處僅提供內建角色的
// 「預設」角色→資源→動作 ACL(對齊原 rbac_policy.csv)與角色繼承展開(RolesFor),
// 供服務層 requireRole/requireScope 作為 fallback 閘門。生產授權決策由 OpenFGA Check
// (middleware 閘門)執行;RLS 承擔資料範圍隔離(company/department/self)。
package auth

// BuiltinRoles 內建 7 角色(設計書 §3.2 + §4.4),系統 seed 值。
var BuiltinRoles = []string{"super", "company_admin", "dept_admin", "staff", "customer", "guest", "developer"}

// rolePolicy 內建角色 → 資源 → 允許動作("" = 該資源全部動作;"*" 資源 = 全部資源)。
// 對齊原 rbac_policy.csv 的 7 內建角色預設 policy(設計書 §3.2 + §4.4)。
var rolePolicy = map[string]map[string][]string{
	"super":     {"*": {"*"}},
	"developer": {"*": {"*"}},
	"company_admin": {
		"company":     {"read", "update"},
		"department":  {"*"},
		"user":        {"*"},
		"role":        {"*"},
		"sales_order": {"*"},
		"customer":    {"*"},
		"product":     {"*"},
		"print":       {"*"},
		"dispatch":    {"*"},
	},
	"dept_admin": {
		"department":  {"read"},
		"user":        {"read"},
		"customer":    {"*"},
		"product":     {"*"},
		"sales_order": {"*"},
		"print":       {"*"},
		"dispatch":    {"*"},
	},
	"staff": {
		"customer":    {"*"},
		"product":     {"*"},
		"sales_order": {"*"},
		"print":       {"*"},
		"dispatch":    {"read"},
		"accounting":  {"read"},
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
