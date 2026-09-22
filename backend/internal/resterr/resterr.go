// Package resterr 將錯誤以 Connect 錯誤協定寫成 JSON（REST 端點共用；Connect RPC 走
// server.writeConnectError，兩者同形）。trace_id 經 requestid.Stamp 附成 ErrorInfo detail，
// 供前端與 log 對照——REST 不經 interceptor，只能在寫出邊界補。
package resterr

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
)

// Write 以 Connect 錯誤協定寫出（code/message/details；trace_id 由 requestid.Stamp 補）。
func Write(w http.ResponseWriter, r *http.Request, err error) {
	ctx, _ := requestid.Ensure(r.Context(), r.URL.Path)
	err = requestid.Stamp(ctx, err)
	body := map[string]any{"code": connect.CodeOf(err).String(), "message": err.Error()}
	if ce, ok := err.(*connect.Error); ok {
		body["message"] = ce.Message()
		var details []map[string]string
		for _, d := range ce.Details() {
			details = append(details, map[string]string{
				"type":  d.Type(),
				"value": base64.RawStdEncoding.EncodeToString(d.Bytes()),
			})
		}
		if details != nil {
			body["details"] = details
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status(err))
	_ = json.NewEncoder(w).Encode(body)
}

// JSON 寫一般 JSON 回應。
func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// status 對映 Connect code → HTTP 狀態。
//
// **全站唯一一份**(含 server.writeConnectError 的 middleware 閘門與所有 REST 端點):兩邊各留
// 一張表會分岐 —— 原本 server 那份沒有 CodeNotFound 分支,同一個 NotFound 走 REST 是 404、
// 走 middleware 是 500。新增 code 只改這裡。
func status(err error) int {
	switch connect.CodeOf(err) {
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized // 401
	case connect.CodePermissionDenied:
		return http.StatusForbidden // 403
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest // 400
	case connect.CodeNotFound:
		return http.StatusNotFound // 404
	case connect.CodeInternal:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 其餘(含 failed_precondition)一律 500
	}
}
