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

// Provision 於啟動時將 DB 授權資料全量同步至 OpenFGA(D32)。
// 分兩段:role_permissions → role→ability tuples;users → user→role assignment tuples。
func Provision(ctx context.Context, e *openfga.Engine, db *ent.Client) error {
	if err := ProvisionRoleAbilities(ctx, e, db); err != nil {
		return err
	}
	return ProvisionUserRoles(ctx, e, db)
}

// ProvisionRoleAbilities 將全部 role_permissions(非 inverted)translate 成 role→ability tuples。
// 依 role 分組:先刪除既有 role→ability tuples 再寫入目前集合(reconcile,使「移除的權限」亦
// 反映到 OpenFGA)。
func ProvisionRoleAbilities(ctx context.Context, e *openfga.Engine, db *ent.Client) error {
	rows, err := db.RolePermission.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("authz: 載入 role_permissions: %w", err)
	}
	byRole := map[int][]*ent.RolePermission{}
	for _, rp := range rows {
		byRole[rp.RoleID] = append(byRole[rp.RoleID], rp)
	}
	for roleID, perms := range byRole {
		if err := deleteRoleAbilities(ctx, e, roleID); err != nil {
			return err
		}
		for _, p := range perms {
			if p.Inverted {
				continue // inverted(拒絕)不得轉為 allow tuple
			}
			rel, ok := permissionRelation(p.Action)
			if !ok {
				continue
			}
			userset := fmt.Sprintf("role:%d#assigned", roleID)
			if err := e.WriteTuple(ctx, userset, rel, "ability:"+p.Resource); err != nil {
				return fmt.Errorf("authz: 寫入 role=%d ability=%s: %w", roleID, p.Resource, err)
			}
		}
	}
	return nil
}

// ProvisionUserRoles 將全部 active 使用者的 role code translate 成 user→role assignment tuples
// (role:<rid>#assigned@user:<uid>)。model 中 role 型別的 assigned 為 user userset,故 tuple 為
// user=<uid> assigned role:<rid>。僅在 role code 對應角色存在時寫入;未知角色略過(不誤授權)。
func ProvisionUserRoles(ctx context.Context, e *openfga.Engine, db *ent.Client) error {
	users, err := db.User.Query().Where(user.StatusEQ(user.StatusActive)).All(ctx)
	if err != nil {
		return fmt.Errorf("authz: 載入 users: %w", err)
	}
	for _, u := range users {
		r, err := db.Role.Query().Where(role.CodeEQ(u.Role)).Only(ctx)
		if err != nil {
			// 角色不存在 → 略過。
			continue
		}
		uid := fmt.Sprintf("%d", u.ID)
		rid := fmt.Sprintf("%d", r.ID)
		if err := e.WriteTuple(ctx, "user:"+uid, "assigned", "role:"+rid); err != nil {
			return fmt.Errorf("authz: 指派 user=%d role=%s: %w", u.ID, rid, err)
		}
	}
	return nil
}

// deleteRoleAbilities 刪除特定 role 的既有 role→ability tuples(role:<rid>#assigned can_* ability:*)。
// 依 role userset 列舉既有 tuples 逐一刪除(reconcile 前置)。
func deleteRoleAbilities(ctx context.Context, e *openfga.Engine, roleID int) error {
	tuples, err := e.ListRoleTuples(ctx, roleID)
	if err != nil {
		return fmt.Errorf("authz: 列舉 role=%d tuples: %w", roleID, err)
	}
	for _, t := range tuples {
		if err := e.DeleteTuple(ctx, t[0], t[1], t[2]); err != nil {
			return fmt.Errorf("authz: 刪除 role=%d tuple %v: %w", roleID, t, err)
		}
	}
	return nil
}
