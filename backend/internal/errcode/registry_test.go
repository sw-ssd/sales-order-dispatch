package errcode_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

var idFormat = regexp.MustCompile(`^[A-Z]{2,6}-\d{4}$`)

// errorInfoOf 取出 connect error detail 內的 ErrorInfo（找不到即失敗）。
// connect 的 Details() 回傳 []*ErrorDetail（本身不是 proto.Message），故經 Value() 還原內層訊息。
func errorInfoOf(t *testing.T, err *connect.Error) *commonv1.ErrorInfo {
	t.Helper()
	for _, d := range err.Details() {
		m, verr := d.Value()
		if verr != nil {
			continue
		}
		if ei, ok := m.(*commonv1.ErrorInfo); ok {
			return ei
		}
	}
	t.Fatalf("錯誤未帶 ErrorInfo detail")
	return nil
}

// 契約 1：每個碼格式正確、有訊息、有 domain，且區段與 connect 碼一致。
func TestRegistryInvariants(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range errcode.All() {
		if !idFormat.MatchString(c.ID) {
			t.Fatalf("碼格式錯誤: %q", c.ID)
		}
		if seen[c.ID] {
			t.Fatalf("碼重複: %s", c.ID)
		}
		seen[c.ID] = true
		if strings.TrimSpace(c.Message) == "" {
			t.Fatalf("%s 缺訊息", c.ID)
		}
		if !strings.HasPrefix(c.ID, string(c.Domain)+"-") {
			t.Fatalf("%s 的 domain %s 與碼前綴不符", c.ID, c.Domain)
		}
		if !errcode.SectionAllows(c.ID, c.ConnectCode) {
			t.Fatalf("%s 的 connect 碼 %v 不屬於區段 %s 的允許集合", c.ID, c.ConnectCode, c.ID[:1])
		}
	}
}

// 契約 2：註冊重複 / 格式錯 / 區段不符 → panic（啟動就失敗，不等上線）。
func TestRegisterPanicsOnViolations(t *testing.T) {
	cases := []struct {
		name string
		fn   func()
	}{
		{"重複 ID", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1001", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInvalidArgument, Message: "x"})
		}},
		{"格式錯誤", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInvalidArgument, Message: "x"})
		}},
		{"區段與 connect 碼不符", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1234", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInternal, Message: "x"})
		}},
		{"缺訊息", func() {
			errcode.MustRegister(errcode.Code{ID: "SYS-1234", Domain: errcode.DomainSys,
				ConnectCode: connect.CodeInvalidArgument})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("應 panic")
				}
			}()
			tc.fn()
		})
	}
}

// 契約 3：Error() 產生帶 ErrorInfo 的 connect error（碼與 details 可被客戶端讀取）。
func TestErrorCarriesErrorInfo(t *testing.T) {
	err := errcode.PlatformLimitExceeded.Error(map[string]string{"used": "10", "limit": "10"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("connect 碼應為 FailedPrecondition，got %v", connect.CodeOf(err))
	}
	if !strings.Contains(err.Message(), "10/10") {
		t.Fatalf("訊息應已渲染參數，got %q", err.Message())
	}
	info := errorInfoOf(t, err)
	if info.GetCode() != "PLAT-5001" {
		t.Fatalf("應帶 ErrorInfo{code: PLAT-5001}，got %+v", info)
	}
	if info.GetDetails()["used"] != "10" {
		t.Fatalf("details 應帶 used=10，got %+v", info.GetDetails())
	}
}

// 契約 4：缺參數不得讓錯誤處理爆掉（回樣板原文），且 Wrap 保留底層錯誤供 log 追查。
func TestRenderMissingParamFallsBackAndWrapKeepsCause(t *testing.T) {
	msg := errcode.PlatformLimitExceeded.Render(nil)
	if !strings.Contains(msg, "{used}") {
		t.Fatalf("缺參數時應保留樣板原文，got %q", msg)
	}
	cause := errors.New("db: connection reset")
	err := errcode.SysInternal.Wrap(cause)
	if !errors.Is(err, cause) {
		t.Fatal("Wrap 應保留底層錯誤（Unwrap），否則 log 追不到原因")
	}
	if strings.Contains(err.Message(), "connection reset") {
		t.Fatal("底層錯誤細節不得進對外訊息")
	}
}

// 契約 5（trace_id）已移至 Task 3 的邊界測試：本套件不感知 ctx。
