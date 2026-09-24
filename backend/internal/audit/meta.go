package audit

import "context"

// Meta 為請求層的稽核來源資訊(IP / User-Agent),由 middleware 每請求注入 ctx。
// 來源資訊為 03 計畫 2.6.2 要求(供稽核調查);缺值時存空,不視為錯誤。
type Meta struct {
	IP        string
	UserAgent string
	// ActorKind 標示操作主體種類:空 = 人為操作(session/JWT);"api-token:<name>" = 機器代打
	// (01 1.6.6 的 server-to-server token)。機器身分沒有自己的 users 列,稽核的 user_id
	// 是 token 綁定的**真實使用者**,故另以此欄位區分「這個人是親自操作還是被排程代打」。
	ActorKind string
}

type metaCtxKey struct{}

// WithMeta 將請求層稽核來源資訊放入 ctx(middleware 每請求注入一次)。
func WithMeta(ctx context.Context, m Meta) context.Context {
	return context.WithValue(ctx, metaCtxKey{}, m)
}

// MetaFrom 由 ctx 取稽核來源資訊;未注入時回零值。
func MetaFrom(ctx context.Context) Meta {
	m, _ := ctx.Value(metaCtxKey{}).(Meta)
	return m
}
