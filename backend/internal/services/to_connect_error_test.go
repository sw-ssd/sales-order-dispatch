package services

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// TestToConnectErrorMapsToRegisteredCodes 釘住 toConnectError 的「輸入錯誤型別 → 註冊碼」映射。
// 意圖不是「有回錯誤」而是「回哪一個碼」:唯一真相來源是 errcode registry(碼一旦發佈不得改義),
// 故這裡同時斷言 connect 碼與註冊碼一致 —— 只改其中一邊就會紅。
func TestToConnectErrorMapsToRegisteredCodes(t *testing.T) {
	cases := []struct {
		name string
		in   error
		id   string
		cc   connect.Code
	}{
		{"找不到", &ent.NotFoundError{}, "SYS-4002", connect.CodeNotFound},
		{"驗證失敗", &ent.ValidationError{Name: "name"}, "SYS-1001", connect.CodeInvalidArgument},
		{"約束錯誤", &ent.ConstraintError{}, "SYS-3002", connect.CodeFailedPrecondition},
		{"RLS 違反(42501 不在 ent 的 constraint 判定內)", &pgconn.PgError{
			Code:    "42501",
			Message: `new row violates row-level security policy for table "customers"`,
		}, "SYS-3001", connect.CodeFailedPrecondition},
		{"多列單值", &ent.NotSingularError{}, "SYS-9000", connect.CodeInternal},
		{"未知錯誤", errors.New("boom"), "SYS-9000", connect.CodeInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := toConnectError(tc.in)

			if got := connect.CodeOf(out); got != tc.cc {
				t.Fatalf("connect 碼 = %v；want %v(%v)", got, tc.cc, out)
			}
			if got := errcodeIDOf(t, out); got != tc.id {
				t.Fatalf("錯誤碼 = %q；want %q", got, tc.id)
			}
			// 期望值必須與 registry 的定義一致(避免測試自己寫了一組不存在的碼/碼別)。
			reg, ok := errcode.Lookup(tc.id)
			if !ok {
				t.Fatalf("%s 未註冊", tc.id)
			}
			if reg.ConnectCode() != tc.cc {
				t.Fatalf("%s 的註冊 connect 碼 = %v；測試期望 %v", tc.id, reg.ConnectCode(), tc.cc)
			}
		})
	}
}

// TestToConnectErrorMasksDBDetail 釘住遮蔽面:驅動層原文(SQLSTATE／policy 文字／表名／約束名)
// 只能落 server log,不得進對外訊息,也不得掛在錯誤鏈上(用 Wrap 就會外洩)。
func TestToConnectErrorMasksDBDetail(t *testing.T) {
	cases := []struct {
		name string
		in   error
		root string // 必須出現在 server log 的根因片段
	}{
		{"RLS 違反", &pgconn.PgError{
			Code:    "42501",
			Message: `new row violates row-level security policy for table "customers"`,
		}, "42501"},
		{"約束錯誤", &ent.ConstraintError{}, "constraint failed"},
		{"未知錯誤", errors.New("boom"), "boom"},
	}
	// 對外訊息不得出現的字串:DB 原文標記、表名、約束名與根因文字。
	banned := []string{"SQLSTATE", "row-level security", "customers", "constraint", "UNIQUE",
		"duplicate key", "users_email", "pgx", "boom"}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var logged bytes.Buffer
			prev := log.Writer()
			log.SetOutput(&logged)
			t.Cleanup(func() { log.SetOutput(prev) })

			out := toConnectError(tc.in)

			ce, ok := out.(*connect.Error)
			if !ok {
				t.Fatalf("應為 *connect.Error,got %T", out)
			}
			for _, b := range banned {
				if strings.Contains(ce.Message(), b) {
					t.Fatalf("對外訊息不得含 %q(驅動層原文):%q", b, ce.Message())
				}
			}
			if !strings.Contains(logged.String(), tc.root) {
				t.Fatalf("根因 %q 必須落 server log 供維運診斷,got %q", tc.root, logged.String())
			}
			// 根因不得掛在錯誤鏈上(不用 Wrap):否則 detail 序列化/log 中介層都可能帶出去。
			if errors.Is(out, tc.in) {
				t.Fatalf("對外錯誤不得保留根因(不得用 Wrap),got %v", out)
			}
		})
	}
}

// TestToConnectErrorPassesThroughConnectError 已是 *connect.Error 者原樣回傳:
// 否則內層碼(如守衛的 internal「缺少租戶交易(context)」)會被外層蓋成 SYS-9000 而同碼異義。
func TestToConnectErrorPassesThroughConnectError(t *testing.T) {
	if got := toConnectError(nil); got != nil {
		t.Fatalf("nil 應回 nil,got %v", got)
	}
	inner := connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	if got := toConnectError(inner); got != inner {
		t.Fatalf("已是 *connect.Error 應原樣回傳,got %v", got)
	}
}

// errcodeIDOf 由 connect error 的 ErrorInfo detail 取錯誤碼(測試輔助)。
func errcodeIDOf(t *testing.T, err error) string {
	t.Helper()
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("應為 *connect.Error,got %T", err)
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			t.Fatalf("detail 取值失敗: %v", derr)
		}
		if info, ok := v.(*commonv1.ErrorInfo); ok {
			return info.GetCode()
		}
	}
	return ""
}
