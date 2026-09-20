// Package requestid 為每個 unary RPC 產生 trace_id：進 ctx（供錯誤碼與 log 使用），
// 並在 log 中帶出（客戶回報代碼時可對照 server log）。
package requestid

import (
	"context"
	"log"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

type ctxKey struct{}

// With 將 trace_id 放入 ctx（測試與跨服務傳遞用）。
func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// From 取 trace_id；未注入時回空字串（呼叫端不得假設非空）。
func From(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// Interceptor 為每個 unary RPC 產生 trace_id 並記一行 log（method + trace_id）。
// 必須掛在租戶交易 interceptor（dbtenant）之前：先有 trace_id，交易錯誤才帶得到。
func Interceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			id := uuid.Must(uuid.NewV7()).String()
			ctx = With(ctx, id)
			log.Printf("rpc: %s trace_id=%s", req.Spec().Procedure, id)
			return next(ctx, req)
		}
	})
}
