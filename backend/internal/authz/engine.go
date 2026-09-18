// Package authz 的 OpenFGA engine 於 ctx 的注入/讀取 helper(D32)。
// Server 於啟動時建立單一 internal/authz/openfga.Engine,每請求注入 ctx,
// 供 middleware 對受保護 RPC 做 OpenFGA Check 授權門檻。
package authz

import (
	"context"

	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
)

type engineCtxKey struct{}

// WithEngine 將 OpenFGA 授權引擎放入 ctx(middleware 每請求注入一次)。
func WithEngine(ctx context.Context, e *authzopenfga.Engine) context.Context {
	return context.WithValue(ctx, engineCtxKey{}, e)
}

// EngineFrom 由 ctx 取 OpenFGA 授權引擎;未注入時回 nil(middleware 據此決定是否跳過)。
func EngineFrom(ctx context.Context) *authzopenfga.Engine {
	e, _ := ctx.Value(engineCtxKey{}).(*authzopenfga.Engine)
	return e
}
