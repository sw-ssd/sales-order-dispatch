package errcode

import (
	"testing"

	"connectrpc.com/connect"
)

// 契約 2：註冊重複 / 格式錯 / 區段與 connect 碼不符 / domain 前綴不符 / 缺訊息 → panic，
// 以及使用零值（未註冊）Code → panic（啟動就失敗，不等上線）。
//
// 為什麼這張表在**套件內**：Code 的欄位未匯出，不合法的碼只有本套件建得出來
// （型別即約束——外部套件連複合字面值都寫不出來，見 code.go 的 Code 註解）。
func TestRegisterPanicsOnViolations(t *testing.T) {
	cases := []struct {
		name string
		fn   func()
	}{
		{"重複 ID", func() {
			MustRegister(Code{id: "SYS-1001", domain: DomainSys,
				connectCode: connect.CodeInvalidArgument, message: "x"})
		}},
		{"格式錯誤", func() {
			MustRegister(Code{id: "SYS-1", domain: DomainSys,
				connectCode: connect.CodeInvalidArgument, message: "x"})
		}},
		{"區段與 connect 碼不符", func() {
			MustRegister(Code{id: "SYS-1234", domain: DomainSys,
				connectCode: connect.CodeInternal, message: "x"})
		}},
		{"缺訊息", func() {
			MustRegister(Code{id: "SYS-1234", domain: DomainSys,
				connectCode: connect.CodeInvalidArgument})
		}},
		// 唯一讓「domain 與 ID 前綴不符」這條守門失效的構造：ID 未註冊、訊息有了、
		// 且 connect 碼符合該 ID 的區段（1xxx + InvalidArgument）——於是只有前綴檢查會 panic。
		// 若前綴檢查被刪掉，這筆會註冊成功、本案例轉紅。
		{"domain 與 ID 前綴不符", func() {
			MustRegister(Code{id: "SYS-1002", domain: DomainAuth,
				connectCode: connect.CodeInvalidArgument, message: "x"})
		}},
		// 零值是外部套件唯一能造出的 Code（欄位未匯出）；使用它必須立刻爆掉，
		// 而不是回一個沒有碼、沒有訊息的錯誤回應。
		{"未註冊的零值 Code", func() {
			Code{}.Error(nil)
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
