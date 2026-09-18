// Package authz 提供 OpenFGA 授權資料的供給(reconcile, D32)。
// 以資料庫為單一來源(role_permissions=角色→能力;users.role=使用者→角色),將授權
// 資料 translate 成 OpenFGA tuples:role→ability(can_read/can_write)與 user→role(assigned)。
// 於 server 啟用 OpenFGA 後呼叫,使 middleware OpenFGA Check 得以正常判定(修復「零 tuple
// → 全員 deny」的死局)。inverted/不支援動作略過(拒絕),與 role_service 同步語意一致。
package authz

import (
	"context"
	"fmt"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
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
	return false
}

// Provision 將 DB 授權資料全量 reconcile 至 OpenFGA(D32):讀取 store 現有 tuples,
// 計算期望集合(role_permissions→role ability;active users→role assigned),
// 刪除「管理範圍內但不在期望集合」的舊 tuples(使移除權限/角色變更不殘留),再補寫缺漏。
func Provision(ctx context.Context, e *openfga.Engine, db *ent.Client) error {
	desired, err := desiredTuples(ctx, db)
	if err != nil {
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
	for _, u := range users {
		r, err := db.Role.Query().Where(role.CodeEQ(u.Role)).Only(ctx)
		if err != nil {
			continue // 角色不存在 → 略過(不誤授權)
		}
		out[tupleKey(fmt.Sprintf("user:%d", u.ID), "assigned", fmt.Sprintf("role:%d", r.ID))] = true
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
