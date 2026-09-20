package errcode_test

import (
	"bytes"
	"errors"
	"log"
	"os"
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

// 契約 1：每個碼格式正確、有訊息、有 domain，區段與 connect 碼一致，以 ID 查得回同一個碼，
// 且 All() 依 ID 遞增排序（產生器靠這個順序讓 `git diff --exit-code` 穩定）。
// （「同 ID 出現兩次」不可能發生——All() 由 map 產生，故不寫那種永不失敗的斷言。）
func TestRegistryInvariants(t *testing.T) {
	prev := ""
	for _, c := range errcode.All() {
		if !idFormat.MatchString(c.ID()) {
			t.Fatalf("碼格式錯誤: %q", c.ID())
		}
		if prev >= c.ID() {
			t.Fatalf("All() 未依 ID 遞增排序: %q 出現在 %q 之後（或重複）", c.ID(), prev)
		}
		prev = c.ID()
		if strings.TrimSpace(c.Message()) == "" {
			t.Fatalf("%s 缺訊息", c.ID())
		}
		if !strings.HasPrefix(c.ID(), string(c.Domain())+"-") {
			t.Fatalf("%s 的 domain %s 與碼前綴不符", c.ID(), c.Domain())
		}
		if !errcode.SectionAllows(c.ID(), c.ConnectCode()) {
			t.Fatalf("%s 的 connect 碼 %v 不屬於區段 %s 的允許集合", c.ID(), c.ConnectCode(), c.ID()[:1])
		}
		if got, ok := errcode.Lookup(c.ID()); !ok || got != c {
			t.Fatalf("%s 以 ID 查不回同一個碼: ok=%v got=%+v", c.ID(), ok, got)
		}
	}
}

// 契約 2（註冊違規即 panic）在 code_internal_test.go：Code 欄位未匯出，
// 不合法的碼只有套件內建得出來（型別即約束）。

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

// 契約 6：缺參數的判定以**樣板中的佔位符**為準，不看渲染結果——
// 參數值本身帶大括號（例：{code} 帶入 "{x}"）不得被誤判成缺參數而記 warn。
func TestRenderDoesNotWarnWhenParamValueHasBraces(t *testing.T) {
	var logged bytes.Buffer
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	got := errcode.CustomerCodeExists.Render(map[string]string{"code": "{x}"})
	if want := "客戶編號 {x} 已存在"; got != want {
		t.Fatalf("Render = %q；want %q", got, want)
	}
	if strings.Contains(logged.String(), "缺參數") {
		t.Fatalf("參數值含大括號不應記缺參數 warn：%q", logged.String())
	}
}
