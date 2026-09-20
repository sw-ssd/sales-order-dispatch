package requestid

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// infoIn 取出錯誤裡第一個 ErrorInfo（沒有即回 nil）。
func infoIn(t *testing.T, ce *connect.Error) *commonv1.ErrorInfo {
	t.Helper()
	for _, d := range ce.Details() {
		m, err := d.Value()
		if err != nil {
			t.Fatalf("detail 取值: %v", err)
		}
		if ei, ok := m.(*commonv1.ErrorInfo); ok {
			return ei
		}
	}
	return nil
}

// infoErr 造一個「帶 ErrorInfo 的 connect error」，等同 errcode 的產物（TraceId 留空）。
func infoErr(trace string) *connect.Error {
	ce := connect.NewError(connect.CodeNotFound, errors.New("資源不存在或無權存取"))
	d, err := connect.NewErrorDetail(&commonv1.ErrorInfo{Code: "SYS-4002", TraceId: trace})
	if err != nil {
		panic(err)
	}
	ce.AddDetail(d)
	return ce
}

// 邊界補值：ctx 有 trace 且 ErrorInfo 的 TraceId 為空 → 補上本請求的 trace；
// 其餘一律原樣回傳（不得改動、不得覆蓋）。
func TestStampTraceID(t *testing.T) {
	ctx := With(context.Background(), "0192f0-trace")

	t.Run("補進本請求的 trace", func(t *testing.T) {
		ce, ok := stampTraceID(ctx, infoErr("")).(*connect.Error)
		if !ok {
			t.Fatal("應仍為 connect error")
		}
		if ce.Code() != connect.CodeNotFound || ce.Message() != "資源不存在或無權存取" {
			t.Fatalf("碼與訊息不得改變，got %v/%q", ce.Code(), ce.Message())
		}
		if ei := infoIn(t, ce); ei.GetCode() != "SYS-4002" || ei.GetTraceId() != "0192f0-trace" {
			t.Fatalf("應保留碼並補上 trace，got %+v", ei)
		}
	})

	t.Run("已有 trace 不覆蓋", func(t *testing.T) {
		ce, ok := stampTraceID(ctx, infoErr("upstream-trace")).(*connect.Error)
		if !ok {
			t.Fatal("應仍為 connect error")
		}
		if got := infoIn(t, ce).GetTraceId(); got != "upstream-trace" {
			t.Fatalf("既有 trace 不得覆蓋，got %q", got)
		}
	})

	t.Run("無 ErrorInfo 原樣", func(t *testing.T) {
		err := connect.NewError(connect.CodeInvalidArgument, errors.New("參數驗證失敗"))
		if got := stampTraceID(ctx, err); got != err {
			t.Fatalf("不帶 ErrorInfo 應原樣回傳同一錯誤，got %v", got)
		}
	})

	t.Run("非 connect error 原樣", func(t *testing.T) {
		sentinel := errors.New("plain")
		got := stampTraceID(ctx, sentinel)
		if !errors.Is(got, sentinel) {
			t.Fatalf("非 connect error 應原樣回傳，got %v", got)
		}
	})

	t.Run("err 為 nil", func(t *testing.T) {
		if got := stampTraceID(ctx, nil); got != nil {
			t.Fatalf("nil 應原樣回 nil，got %v", got)
		}
	})

	t.Run("ctx 無 trace 原樣", func(t *testing.T) {
		err := infoErr("")
		if got := stampTraceID(context.Background(), err); got != err {
			t.Fatalf("ctx 無 trace 應原樣回傳同一錯誤(不得 panic)，got %v", got)
		}
	})

	t.Run("其他 detail 與 Meta 保留", func(t *testing.T) {
		ce := infoErr("")
		other, err := connect.NewErrorDetail(&emptypb.Empty{})
		if err != nil {
			t.Fatalf("造 detail: %v", err)
		}
		ce.AddDetail(other)
		ce.Meta().Add("X-Tenant", "acme")

		got, ok := stampTraceID(ctx, ce).(*connect.Error)
		if !ok {
			t.Fatal("應仍為 connect error")
		}
		if got.Meta().Get("X-Tenant") != "acme" {
			t.Fatalf("重建後 Meta 不得遺失，got %v", got.Meta())
		}
		var infos, others int
		for _, d := range got.Details() {
			m, verr := d.Value()
			if verr != nil {
				t.Fatalf("detail 取值: %v", verr)
			}
			switch m.(type) {
			case *commonv1.ErrorInfo:
				infos++
			case *emptypb.Empty:
				others++
			default:
				t.Fatalf("不認識的 detail: %T", m)
			}
		}
		if infos != 1 || others != 1 {
			t.Fatalf("重建後應保留其他 detail（ErrorInfo=%d 其他=%d）", infos, others)
		}
	})
}

// serve 起真 HTTP server（httptest）＋掛 Interceptor 的 unary handler，用真 connect client
// 呼叫一次，回傳（客戶端收到的錯誤, handler 內看到的 trace_id）。
func serve(t *testing.T, unary func(ctx context.Context) error) (*connect.Error, string) {
	t.Helper()
	var handlerTrace string
	h := connect.NewUnaryHandler("/test.v1.Trace/Get",
		func(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			handlerTrace = From(ctx)
			return nil, unary(ctx)
		},
		connect.WithInterceptors(Interceptor()))
	mux := http.NewServeMux()
	mux.Handle("/test.v1.Trace/Get", h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := connect.NewClient[emptypb.Empty, emptypb.Empty](srv.Client(), srv.URL+"/test.v1.Trace/Get")
	_, err := client.CallUnary(context.Background(), connect.NewRequest(&emptypb.Empty{}))
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("應收到 connect error，got %v", err)
	}
	return ce, handlerTrace
}

// P0-5 的驗收：碼與 trace_id 跨網路到得了客戶端。
// 這是「trace 由邊界補」的端到端證明——errcode 不收 ctx，客戶端收到的 trace_id
// 只可能來自 interceptor 的補值。
func TestErrorInfoTraceIDReachesClient(t *testing.T) {
	ce, handlerTrace := serve(t, func(context.Context) error { return errcode.SysNotFound.Error(nil) })

	if ce.Code() != connect.CodeNotFound {
		t.Fatalf("connect 碼應保留 NotFound，got %v", ce.Code())
	}
	ei := infoIn(t, ce)
	if ei.GetCode() != "SYS-4002" {
		t.Fatalf("客戶端應收到 ErrorInfo{SYS-4002}，got %+v", ei)
	}
	if handlerTrace == "" {
		t.Fatal("handler 內應看得到 trace_id（T3 既有行為）")
	}
	if ei.GetTraceId() != handlerTrace {
		t.Fatalf("客戶端收到的 trace_id(%q) 應等於同請求 handler 內的值(%q)", ei.GetTraceId(), handlerTrace)
	}
}

// 不帶 ErrorInfo 的錯誤：回應原樣（碼與訊息不變），也不憑空多出 detail。
func TestErrorWithoutErrorInfoUnchanged(t *testing.T) {
	ce, _ := serve(t, func(context.Context) error {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("參數驗證失敗"))
	})
	if ce.Code() != connect.CodeInvalidArgument || ce.Message() != "參數驗證失敗" {
		t.Fatalf("應原樣回傳，got %v/%q", ce.Code(), ce.Message())
	}
	if n := len(ce.Details()); n != 0 {
		t.Fatalf("不應多出 detail，got %d 個: %v", n, ce.Details())
	}
}
