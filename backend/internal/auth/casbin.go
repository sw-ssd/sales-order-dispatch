// Casbin RBAC 執行層(T14):角色 × 公司(domain)× 資源 × 動作 的授權判斷。
// model 與 7 內建角色 policy 位於 config/rbac_model.conf 與 config/rbac_policy.csv
// (go:embed 編譯期內嵌,測試與正式 binary 皆可直接使用)。
// 設計書 §3.3:Casbin 僅負責角色/資源/動作授權;部門資料範圍由 RLS 承擔,故 policy dom 皆為 "*"。
package auth

import (
	"fmt"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"

	"github.com/salesorder/sales-order-1.0/backend/config"
)

// BuiltinRoles 內建 7 角色(設計書 §3.2 + §4.4),系統 seed 值。
var BuiltinRoles = []string{"super", "company_admin", "dept_admin", "staff", "customer", "guest", "developer"}

var (
	enforcerOnce sync.Once
	enforcerInst *casbin.Enforcer
	enforcerErr  error
)

// casbinEnforcer 回傳進程級 enforcer(model + policy 皆來自 config go:embed)。
// production 遷移至 casbin_rules 表後改為 DB adapter(設計書 3.3 / 8.2「API 權限設置」)。
func casbinEnforcer() (*casbin.Enforcer, error) {
	enforcerOnce.Do(func() {
		m, err := model.NewModelFromString(config.RBACModel)
		if err != nil {
			enforcerErr = fmt.Errorf("casbin: model 載入失敗: %w", err)
			return
		}
		e, err := casbin.NewEnforcer(m, stringadapter.NewAdapter(config.RBACPolicy))
		if err != nil {
			enforcerErr = fmt.Errorf("casbin: enforcer 建立失敗: %w", err)
			return
		}
		enforcerInst = e
	})
	return enforcerInst, enforcerErr
}

// Enforce 判斷角色 code 於公司 dom 對資源 obj 執行動作 act 是否允許(brief 介面)。
// 任一 g 繼承鏈命中即允許;policy dom 以 "*" 表示全部公司。
func Enforce(role, obj, act, dom string) (bool, error) {
	e, err := casbinEnforcer()
	if err != nil {
		return false, err
	}
	return e.Enforce(role, obj, act, dom)
}

// EnforceAny 對角色清單逐一 Enforce,任一允許即通過(fail-closed:全拒才 false)。
// middleware 與服務層以 identity.Roles(已含 g 展開)呼叫。
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

// RolesFor 展開角色 code 的隱含角色集(依 g 繼承,含自身,去重保序)。
// 供 middleware 組 identity.Roles:CASL 依此載入各角色 role_permissions 規則。
func RolesFor(role string) []string {
	e, err := casbinEnforcer()
	if err != nil {
		return []string{role}
	}
	inherited, err := e.GetImplicitRolesForUser(role)
	if err != nil {
		return []string{role}
	}
	out := make([]string, 0, 1+len(inherited))
	seen := make(map[string]bool, 1+len(inherited))
	for _, r := range append([]string{role}, inherited...) {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}
