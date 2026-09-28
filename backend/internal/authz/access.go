// Package authz 提供授權上下文與身分/規則註冊(D30-2/D30-3 → D32)。
// 授權決策已由 OpenFGA(internal/authz/openfga)提供:middleware 對受保護 RPC 做 Check
// (資源級),服務層以 PermissionGrantedTx(DB)＋內建 ACL 後備做逐動作二次確認(見下)。
// 本檔保留身分注入(ctx)、資料庫注入(ctx)與條件欄位白名單 Registry(供
// role_service 條件驗證與前端條件建構器使用)。
// 身分由 middleware 注入 ctx(server.authzMiddleware)。
package authz

import (
	"context"
	"fmt"
	"sync"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/scopecond"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

type ctxKey int

const (
	keyRegistry ctxKey = iota
	keyIdentity
	keyDB
)

// PermissionGrantedTx 以 role_permissions 表判定角色是否具 resource×action 權限(D33)。
// 回傳 true = DB 明確認可。DB 未注入 ctx 或查無該列 → 回傳 false(「無意見」),
// 由呼叫端決定後備(見 requireScope:內建 ACL 後備),不可逕自視為拒絕。
//
// 為何需要服務層第二次檢查:middleware 的 OpenFGA 閘門只判 can_read/can_write(二元),
// 同一資源的不同動作可能由不同角色持有 —— company_admin 有 company/update 卻無
// company/create|delete,二元關係會讓 update 授權放行 create。故逐動作檢查必須保留,
// 但角色來源改以 DB 為準(自訂角色只存在於 role_permissions,硬編碼表看不懂它們)。
//
// 安全性:本函式只會「額外允許」role_permissions 明確授予的組合;呼叫端的後備是內建 ACL,
// 內建角色的答案與 DB 種子同源(BuiltinRolePermissions),故不引入新的放行面。
func PermissionGrantedTx(ctx context.Context, roles []string, resource, action string) (bool, error) {
	db := DBFrom(ctx)
	if db == nil {
		return false, nil // 無 DB → 無意見,交由呼叫端後備
	}
	// 若已有請求交易,必須用交易內的 client:role_permissions 自 00025 起受 RLS 約束
	// (policy 要求 app.current_data_scope 非空),而 SET LOCAL 只作用於當前交易 ——
	// 走交易外的 client 會查不到任何列,自訂角色因此靜默失去全部權限。
	if tx, ok := dbtenant.TxFrom(ctx); ok {
		db = tx.Client()
	}
	if len(roles) == 0 {
		return false, nil
	}
	ok, err := db.RolePermission.Query().
		Where(
			rolepermission.HasRoleWith(role.CodeIn(roles...)),
			rolepermission.ResourceEQ(resource),
			rolepermission.ActionEQ(action),
			rolepermission.InvertedEQ(false),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("authz: 查詢角色權限: %w", err)
	}
	return ok, nil
}

// Identity 為授權所需的身分(middleware 由 rls.Identity 轉換注入)。
type Identity struct {
	UserID       string
	CompanyID    string
	DepartmentID string
	CustomerID   string
	Role         string
	Roles        []string // 全部角色 code(依內建角色繼承展開(RolesFor))

	// MustChangePassword 為首登/臨時密碼態(A3 1.5.2):true 時 middleware 僅放行
	// ChangePassword,其餘受保護 RPC 回 failed_precondition。
	MustChangePassword bool
}

// WithIdentity 將身分放入 ctx(測試與 middleware 使用)。
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, keyIdentity, id)
}

// IdentityFrom 由 ctx 取身分;未注入時回零值。
func IdentityFrom(ctx context.Context) Identity {
	id, _ := ctx.Value(keyIdentity).(Identity)
	return id
}

// WithDB 將 ent client 放入 ctx(測試與 server 組裝使用)。
func WithDB(ctx context.Context, db *ent.Client) context.Context {
	return context.WithValue(ctx, keyDB, db)
}

// Registry 回傳進程級 FieldRegistry(啟動時註冊全部 subject,供條件白名單)。
func Registry(ctx context.Context) *scopecond.FieldRegistry {
	if r, ok := ctx.Value(keyRegistry).(*scopecond.FieldRegistry); ok {
		return r
	}
	return defaultRegistry()
}

var (
	registryOnce sync.Once
	registryInst *scopecond.FieldRegistry
)

func defaultRegistry() *scopecond.FieldRegistry {
	registryOnce.Do(func() {
		registryInst = scopecond.NewFieldRegistry()
		scopecond.RegisterSalesOrder(registryInst)
	})
	return registryInst
}

// DBFrom 由 ctx 取 ent client(server 與測試以 WithDB 注入);未注入時回 nil。
func DBFrom(ctx context.Context) *ent.Client {
	db, _ := ctx.Value(keyDB).(*ent.Client)
	return db
}
