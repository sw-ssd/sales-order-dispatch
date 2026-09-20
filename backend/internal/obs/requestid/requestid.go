// Package requestid 為每個 unary RPC 產生 trace_id：進 ctx（供錯誤碼與 log 使用），
// 在 log 中帶出（客戶回報代碼時可對照 server log），並在**回應邊界**補進錯誤的 ErrorInfo。
package requestid

import (
	"context"
	"errors"
	"log"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
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

// Interceptor 為每個 unary RPC 產生 trace_id、記一行 log（method + trace_id），
// 並在回應邊界把它補進錯誤的 ErrorInfo（見 stampTraceID）。
// 必須掛在租戶交易 interceptor（dbtenant）之前：先有 trace_id，交易錯誤才帶得到。
func Interceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			id := uuid.Must(uuid.NewV7()).String()
			ctx = With(ctx, id)
			log.Printf("rpc: %s trace_id=%s", req.Spec().Procedure, id)
			res, err := next(ctx, req)
			return res, stampTraceID(ctx, err)
		}
	})
}

// stampTraceID 為「已帶 ErrorInfo 且 TraceId 為空」的錯誤補上本請求的 trace_id；其餘原樣回傳
// （不覆蓋已有值、不改動不帶 ErrorInfo 或非 connect 的錯誤）。
//
// 為什麼在邊界補：錯誤碼有 176 個生產呼叫點，逐點傳 ctx 是無謂的機械改動；且「當前請求的
// trace_id」本來就只有邊界知道（errcode 因此刻意不收 ctx）。
//
// 為什麼是重建而非就地修改：connect-go 的 Details() 雖回傳同一個 *ErrorDetail，但 Value()
// 回傳的是 proto.Clone(d.pbInner)，改它不會生效；對外序列化用的又是 NewErrorDetail 當下就
// marshal 好的 Any 位元組。已以跨網路測試實證：就地修改的客戶端收到空 trace_id。
func stampTraceID(ctx context.Context, err error) error {
	id := From(ctx)
	if id == "" || err == nil {
		return err
	}
	ce, ok := err.(*connect.Error)
	if !ok {
		return err
	}
	details := ce.Details()
	stamped := make([]*connect.ErrorDetail, 0, len(details))
	found := false
	for _, d := range details {
		m, verr := d.Value()
		if verr != nil {
			// 解不出型別就無法斷定它是不是 ErrorInfo——原樣回傳，不賭。
			return err
		}
		ei, isInfo := m.(*commonv1.ErrorInfo)
		if !isInfo {
			nd, derr := connect.NewErrorDetail(m)
			if derr != nil {
				return err
			}
			stamped = append(stamped, nd)
			continue
		}
		// 已有值不得覆蓋；多個 ErrorInfo（不該發生）時不猜哪個是本請求的，一律原樣回傳。
		if ei.GetTraceId() != "" || found {
			return err
		}
		ei.TraceId = id
		nd, derr := connect.NewErrorDetail(ei)
		if derr != nil {
			return err
		}
		stamped = append(stamped, nd)
		found = true
	}
	if !found {
		return err
	}
	// 重建：只帶 connect 碼與對外訊息（底層錯誤與 DB 原文一律不進新錯誤），
	// 其他 detail 與 Meta（headers/trailers）逐鍵帶回，避免補 trace 反而丟資訊。
	ne := connect.NewError(ce.Code(), errors.New(ce.Message()))
	for _, d := range stamped {
		ne.AddDetail(d)
	}
	for k, vs := range ce.Meta() {
		for _, v := range vs {
			ne.Meta().Add(k, v)
		}
	}
	return ne
}
