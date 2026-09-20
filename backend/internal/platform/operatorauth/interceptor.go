package operatorauth

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// Interceptor 驗 operator cookie;失敗一律 Unauthenticated(AUTH-4001)。
// 租戶 session/JWT 走的是不同 cookie 名稱與不同 secret,故不會、也不能通過這裡。
func (s *Service) Interceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			raw, err := cookieFrom(req)
			if err != nil {
				return nil, errcode.AuthUnauthenticated.Error(nil)
			}
			id, err := s.verify(ctx, raw)
			if err != nil {
				// 根因(簽章／audience／白名單)只供 log 與 Unwrap,不進對外訊息。
				return nil, errcode.AuthUnauthenticated.Wrap(err)
			}
			return next(WithIdentity(ctx, id), req)
		}
	})
}

// cookieFrom 由請求標頭解析 operator cookie。
func cookieFrom(req connect.AnyRequest) (string, error) {
	c, err := (&http.Request{Header: req.Header()}).Cookie(CookieName)
	if err != nil {
		return "", errors.New("缺少 operator cookie")
	}
	return c.Value, nil
}
