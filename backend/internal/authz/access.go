// Package authz 提供授權上下文與身分/規則註冊(D30-2/D30-3 → D32)。
// 授權決策已由 OpenFGA(internal/authz/openfga)提供:CASL 執行面(AccessibleFilter/Can)
// 移除,本檔保留身分注入(ctx)、資料庫注入(ctx)與條件欄位白名單 Registry(供
// role_service 條件驗證與前端條件建構器使用)。
// 開關狀態與身分由 middleware 注入 ctx(server.authzMiddleware)。
package authz

import (
	"context"
	"sync"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/casl"
)

type ctxKey int

const (
	keyEnabled ctxKey = iota
	keyRegistry
	keyIdentity
	keyDB
)

// Identity 為授權所需的身分(middleware 由 rls.Identity 轉換注入)。
type Identity struct {
	UserID       string
	CompanyID    string
	DepartmentID string
	CustomerID   string
	Role         string
	Roles        []string // 全部角色 code(依 Casbin g 規則展開)
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

// WithCASLEnabled 由 middleware 每請求呼叫一次,記錄開關狀態。
// 註:OpenFGA 落地後(CASL 執行面移除)此開關僅供向後相容記錄,授權決策不再查詢它。
func WithCASLEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, keyEnabled, enabled)
}

// Registry 回傳進程級 FieldRegistry(啟動時註冊全部 subject,供條件白名單)。
func Registry(ctx context.Context) *casl.FieldRegistry {
	if r, ok := ctx.Value(keyRegistry).(*casl.FieldRegistry); ok {
		return r
	}
	return defaultRegistry()
}

var (
	registryOnce sync.Once
	registryInst *casl.FieldRegistry
)

func defaultRegistry() *casl.FieldRegistry {
	registryOnce.Do(func() {
		registryInst = casl.NewFieldRegistry()
		casl.RegisterSalesOrder(registryInst)
	})
	return registryInst
}
