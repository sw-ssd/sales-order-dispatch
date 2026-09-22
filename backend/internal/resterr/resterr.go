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

// status 對映 Connect code → HTTP 狀態（與 server.httpStatusForCode 同表）。
func status(err error) int {
	switch connect.CodeOf(err) {
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized
	case connect.CodePermissionDenied:
		return http.StatusForbidden
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest
	case connect.CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
