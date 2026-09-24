// Package authz 提供 OpenFGA 授權資料的供給(reconcile, D32)。
// 以資料庫為單一來源(role_permissions=角色→能力;users.role=使用者→角色),將授權
// 資料 translate 成 OpenFGA tuples:role→ability(can_read/can_write)與 user→role(assigned)。
// 於 server 啟用 OpenFGA 後呼叫,使 middleware OpenFGA Check 得以正常判定(修復「零 tuple
// → 全員 deny」的死局)。inverted/不支援動作略過(拒絕),與 role_service 同步語意一致。
package authz

import (
	"context"
	"fmt"
	"sort"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// permissionRelation 對映權限動作 → OpenFGA relation(can_read/can_write;其餘略過)。
// 與 role_service.permissionRelation 同源(避免雙份展開邏輯漂移)。
func permissionRelation(action string) (string, bool) {
	switch action {
	case "read":
		return "can_read", true
	case "create", "update", "delete", "write", "cancel", "manage", "approve", "reject", "export":
		return "can_write", true
	default:
		return "", false
	}
}

// primaryAccountRelation 為 model 中排除客戶主帳號的 relation 名(見 third_party/openfga modelDSL)。
const primaryAccountRelation = "primary_account"

// PrimaryAccountDeniedResources 回傳客戶主帳號必須被排除的業務資源(規格 4.2:主帳號僅供帳號管理,
// 下單/訂單歷史/退貨/專屬商品/促銷等業務 API 一律 403)。
//
// **由 customer 角色的權限集合推導**,不是手寫清單:customer 之後新增任何業務資源(例如促銷),
// 排除清單自動涵蓋;手寫清單則會靜默漏掉,變成「主帳號可用新業務 API」的破口。
// 排除了 service 層不經 ability 的資源(company/department/user/role 等主檔類,主帳號本就不具備)。
func PrimaryAccountDeniedResources() []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range auth.BuiltinRolePermissions() {
		// accountManagementResource 是主帳號**唯一被允許**的功能面(規格 4.2:主帳號登入僅供
		// 帳號管理)。若它哪天被納入 customer 角色權限,這裡必須排除 —— 否則推導會把它一起 deny,
		// 主帳號就什麼都做不了(含它唯一該能做的事)。
		if p.Role != "customer" || p.Resource == accountManagementResource || seen[p.Resource] {
			continue
		}
		seen[p.Resource] = true
		out = append(out, p.Resource)
	}
	sort.Strings(out)
	return out
}

// accountManagementResource 為店家自助帳號管理的權限資源名(CustomerAccountService)。
// 主帳號的排除清單必須永遠不含它(見 PrimaryAccountDeniedResources)。
const accountManagementResource = "customer_account"

// managedTuple 判斷 tuple 是否屬本 reconcile 管理範圍(role→ability 與 user→role),
// 避免誤刪非本模組寫入的其他 tuple。
func managedTuple(user, relation, object string) bool {
	// role→ability: user="role:<rid>#assigned"、relation=can_read/can_write、object="ability:<res>"。
	if relation == "can_read" || relation == "can_write" {
		return len(user) > 5 && user[:5] == "role:" && len(object) > 8 && object[:8] == "ability:"
	}
	// user→role assignment: user="user:<uid>"、relation="assigned"、object="role:<rid>"。
	if relation == "assigned" {
		return len(user) > 5 && user[:5] == "user:" && len(object) > 5 && object[:5] == "role:"
	}
	// 客戶主帳號排除(D22): user="user:<uid>"、relation="primary_account"、object="ability:<res>"。
	// 必須納入管理範圍,否則帳號由主改子(或反之)時舊 tuple 會殘留成永久 deny。
	if relation == primaryAccountRelation {
		return len(user) > 5 && user[:5] == "user:" && len(object) > 8 && object[:8] == "ability:"
	}
	// 客戶主帳號排除(D22): user="user:<uid>"、relation="primary_account"、object="ability:<res>"。
	// 必須納入管理範圍,否則帳號由主改子(或反之)時舊 tuple 會殘留成永久 deny。
	return false
}

// Provision 將 DB 授權資料全量 reconcile 至 OpenFGA(D32):讀取 store 現有 tuples,
// 計算期望集合(role_permissions→role ability;active users→role assigned),
// 刪除「管理範圍內但不在期望集合」的舊 tuples(使移除權限/角色變更不殘留),再補寫缺漏。
//
// 讀取必須跑在**系統範圍**:本函式於開機時由已裝飾的業務 client 呼叫(server.mountOpenFGA),
// 此刻沒有任何請求 scope,而核心三表(role_permissions/users/roles)在 00028 之後受 RLS 約束 ——
// 未包系統範圍會靜默讀到 0 列,於是「一個 tuple 都不寫」而**不報錯**(之後所有受保護 RPC 一致
// deny,比報錯更難診斷)。
func Provision(ctx context.Context, e *openfga.Engine, db *ent.Client) error {
	var desired map[string]bool
	if err := dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		var err error
		desired, err = desiredTuples(ctx, tx.Client())
		return err
	}); err != nil {
		return err
	}
	existing, err := e.ListTuples(ctx)
	if err != nil {
		return fmt.Errorf("authz: 列舉既有 tuples: %w", err)
	}
	// 刪除管理範圍內、但已不在期望集合的 tuples(避免殘留授權)。
	for _, t := range existing {
		if !managedTuple(t[0], t[1], t[2]) {
			continue
		}
		key := tupleKey(t[0], t[1], t[2])
		if desired[key] {
			continue
		}
		if err := e.DeleteTuple(ctx, t[0], t[1], t[2]); err != nil {
			return fmt.Errorf("authz: 刪除殘留 tuple %v: %w", t, err)
		}
	}
	// 補寫缺漏(期望集合中尚不存在的 tuples,以現存集合去重)。
	have := map[string]bool{}
	for _, t := range existing {
		have[tupleKey(t[0], t[1], t[2])] = true
	}
	for key := range desired {
		if have[key] {
			continue
		}
		u, rel, obj := splitTupleKey(key)
		if err := e.WriteTuple(ctx, u, rel, obj); err != nil {
			return fmt.Errorf("authz: 寫入 tuple %v: %w", key, err)
		}
	}
	return nil
}

// desiredTuples 由 DB 計算期望的 OpenFGA tuple 集合(key=user\x00relation\x00object)。
// db **必須是已套用系統範圍的 client**(唯一呼叫端 Provision 傳入 tx.Client()):這三張表在
// 00028 之後受 RLS 約束,未帶 scope 的查詢會回 0 列 → 靜默佈建 0 筆而不報錯。
func desiredTuples(ctx context.Context, db *ent.Client) (map[string]bool, error) {
	out := map[string]bool{}

	// role_permissions(非 inverted、支援動作)→ role→ability。
	perms, err := db.RolePermission.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("authz: 載入 role_permissions: %w", err)
	}
	for _, p := range perms {
		if p.Inverted {
			continue // inverted(拒絕)不得轉為 allow tuple
		}
		rel, ok := permissionRelation(p.Action)
		if !ok {
			continue
		}
		userset := fmt.Sprintf("role:%d#assigned", p.RoleID)
		out[tupleKey(userset, rel, "ability:"+p.Resource)] = true
	}

	// active users(role code 對應既有角色)→ user→role assignment。
	users, err := db.User.Query().Where(user.StatusEQ(user.StatusActive)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("authz: 載入 users: %w", err)
	}
	// 客戶主帳號(D22/規格 4.2:僅供帳號管理,業務 API 一律 403)→ ability primary_account 排除 tuple。
	// 每個業務資源各一條:model 的 can_read/can_write 以 `but not primary_account` 排除持有者。
	// 為何不寫「role→ability 的否定」:規格語意是「這個**人**不得用業務 API」,不是角色被收回能力
	// ——主帳號與子帳號同角色,能力集合必須對子帳號照常生效,故排除掛在 user 層。
	//
	// 條件必須是 is_primary **且** is_customer:users.is_primary 是通用欄位,員工帳號被誤設
	// (或日後其他子系統沿用同欄位)時,單看 is_primary 會讓該員工被 deny 掉全部業務能力 ——
	// 「主帳號」這個概念只存在於客戶帳號體系(D22),故以 is_customer 收斂語意。
	for _, u := range users {
		r, err := db.Role.Query().Where(role.CodeEQ(u.Role)).Only(ctx)
		if err != nil {
			continue // 角色不存在 → 略過(不誤授權)
		}
		out[tupleKey(fmt.Sprintf("user:%d", u.ID), "assigned", fmt.Sprintf("role:%d", r.ID))] = true
		if u.IsPrimary && u.IsCustomer {
			userObj := fmt.Sprintf("user:%d", u.ID)
			for _, res := range PrimaryAccountDeniedResources() {
				out[tupleKey(userObj, primaryAccountRelation, "ability:"+res)] = true
			}
		}
	}
	return out, nil
}

// tupleKey 以 \x00 串接 (user, relation, object) 作為集合鍵。
func tupleKey(user, relation, object string) string {
	return user + "\x00" + relation + "\x00" + object
}

// splitTupleKey 還原 tupleKey 的三元組。
func splitTupleKey(k string) (user, relation, object string) {
	parts := []string{}
	cur := ""
	for i := 0; i < len(k); i++ {
		if k[i] == 0 {
			parts = append(parts, cur)
			cur = ""
			continue
		}
		cur += string(k[i])
	}
	parts = append(parts, cur)
	if len(parts) != 3 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[2]
}
