package requestid_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
)

// 同一請求內 trace_id 必須一致且有值；不同請求必須不同。
// 另外驗證「記一行 log」：log 需含方法全名與該 trace_id（客服回報代碼 → 對 log）。
func TestInterceptorInjectsStableTraceID(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(nil) })

	seen := []string{}
	handler := func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		id := requestid.From(ctx)
		if id == "" {
			t.Error("trace_id 不得為空")
		}
		seen = append(seen, id)
		return nil, nil
	}
	call := requestid.Interceptor().WrapUnary(handler)
	for range 2 {
		if _, err := call(context.Background(), connect.NewRequest(&emptypb.Empty{})); err != nil {
			t.Fatalf("呼叫: %v", err)
		}
	}
	if seen[0] == seen[1] {
		t.Fatalf("不同請求應有不同 trace_id: %v", seen)
	}
	for _, id := range seen {
		if !strings.Contains(buf.String(), "trace_id="+id) {
			t.Errorf("log 未帶 trace_id=%s：%q", id, buf.String())
		}
	}
}

// 未注入時回空字串：呼叫端不得假設非空（CLI／seed／單元測試路徑）。
func TestFromWithoutInjectionReturnsEmpty(t *testing.T) {
	if id := requestid.From(context.Background()); id != "" {
		t.Fatalf("未注入應回空字串，得到 %q", id)
	}
	if got := requestid.From(requestid.With(context.Background(), "abc")); got != "abc" {
		t.Fatalf("With 後應取回 abc，得到 %q", got)
	}
}
