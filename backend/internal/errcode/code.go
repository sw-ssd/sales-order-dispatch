// Package errcode 為對外錯誤碼的唯一真相來源。
// 設計要點：常數即註冊（MustRegister 於套件 init 驗證格式／唯一性／區段與 connect 碼的一致性，
// 違反即 panic → 啟動就失敗，而非上線後才發現）；碼發佈後不得重用或改義，廢止只標 deprecated。
//
// 本套件是葉節點：不 import internal/obs 或任何服務層套件，任何層皆可直接引用。
package errcode

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// Domain 為碼的域前綴（碼的形態：域-4位數）。
type Domain string

const (
	DomainSys      Domain = "SYS"
	DomainAuth     Domain = "AUTH"
	DomainPlatform Domain = "PLAT"
	DomainCustomer Domain = "CUST"
)

// Code 為一筆錯誤碼定義。
//
// 欄位刻意**未匯出**（型別即約束）：合法的碼只能由 MustRegister 產生，外部套件連複合字面值都
// 建不出來，因此不可能繞過 registry 的啟動驗證（格式／區段／重複）造出未註冊的碼。
// 外部只取用匯出的存取子；Code 仍是可比對、可放 map 的值型別。
type Code struct {
	id          string
	domain      Domain
	connectCode connect.Code // 對外 connect 碼（由區段決定，MustRegister 驗證）
	message     string       // 繁中樣板，可含 {name} 參數
	deprecated  bool
}

// ID 回傳對外錯誤碼，例 "CUST-2001"。
func (c Code) ID() string { return c.id }

// Domain 回傳碼的域。
func (c Code) Domain() Domain { return c.domain }

// ConnectCode 回傳對外 connect 碼（由區段規則決定）。
func (c Code) ConnectCode() connect.Code { return c.connectCode }

// Message 回傳未渲染的繁中訊息樣板。
func (c Code) Message() string { return c.message }

// IsDeprecated 回報此碼是否已廢止（廢止只標記，不得重用或改義）。
func (c Code) IsDeprecated() bool { return c.deprecated }

var (
	idPattern = regexp.MustCompile(`^[A-Z]{2,6}-(\d{4})$`)
	// placeholderPattern 比對訊息樣板中的 {name} 佔位符。
	placeholderPattern = regexp.MustCompile(`\{(\w+)\}`)
	registry           = map[string]Code{}
)

// sectionRules 為「區段 → 允許的 connect 碼」硬規則。碼的分類必須與對外 connect 碼一致，
// 否則前端無法由區段推斷處理方式（401 重新登入、403 找管理員、422 顯示欄位錯誤…）。
// 未列出的區段一律不允許（新增區段必須同時更新此表與規格文件）。
var sectionRules = map[int][]connect.Code{
	1: {connect.CodeInvalidArgument},
	2: {connect.CodeAlreadyExists},
	3: {connect.CodeFailedPrecondition},
	4: {connect.CodePermissionDenied, connect.CodeUnauthenticated, connect.CodeNotFound},
	5: {connect.CodeFailedPrecondition},
	9: {connect.CodeInternal},
}

// MustRegister 註冊一個碼（只能在碼的宣告處呼叫）；違反下列任一即 panic
// （啟動時暴露，而非上線後才發現）：格式不符、ID 重複、缺訊息、domain 與前綴不符、
// connect 碼與區段規則不符。
func MustRegister(c Code) Code {
	if !idPattern.MatchString(c.id) {
		panic(fmt.Sprintf("errcode: 碼格式錯誤 %q（應為 域-4位數）", c.id))
	}
	if _, dup := registry[c.id]; dup {
		panic(fmt.Sprintf("errcode: 碼重複註冊 %q", c.id))
	}
	if strings.TrimSpace(c.message) == "" {
		panic(fmt.Sprintf("errcode: %q 缺訊息", c.id))
	}
	if !strings.HasPrefix(c.id, string(c.domain)+"-") {
		panic(fmt.Sprintf("errcode: %q 與 domain %q 不符", c.id, c.domain))
	}
	if !SectionAllows(c.id, c.connectCode) {
		panic(fmt.Sprintf("errcode: %q 的 connect 碼 %v 與區段 %s 的允許集合不符",
			c.id, c.connectCode, c.id[:1]))
	}
	registry[c.id] = c
	return c
}

// SectionAllows 斷言「此碼的 connect 碼落在其區段允許集合內」；格式不符或區段未定義即 false。
// 4xxx 有三個允許值（PermissionDenied／Unauthenticated／NotFound），故不能只比對單一值。
// 與 MustRegister 共用同一份規則：測試與產生器斷言的正是註冊時驗證過的條件。
func SectionAllows(id string, cc connect.Code) bool {
	m := idPattern.FindStringSubmatch(id)
	if m == nil {
		return false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return false
	}
	return slices.Contains(sectionRules[n/1000], cc)
}

// Lookup 以 ID 取碼。
func Lookup(id string) (Code, bool) {
	c, ok := registry[id]
	return c, ok
}

// All 回傳全部碼（依 ID 排序，供產生器與文件）。
func All() []Code {
	out := make([]Code, 0, len(registry))
	for _, c := range registry {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out
}

// Render 以參數渲染訊息樣板；缺參數時該佔位符保留樣板原文並記 log warn
// （不得讓錯誤處理本身爆掉）。
//
// 缺參數的判定以**樣板中的佔位符**為準（渲染前抽一次），不看渲染結果——否則參數值本身
// 帶大括號（例：{code} 帶入 "{x}"）會被誤判成缺參數；且逐佔位符替換不會二次替換剛帶入的值。
func (c Code) Render(params map[string]string) string {
	// 零值是外部唯一能造出的 Code（欄位未匯出）；濫用是程式錯誤，立刻爆掉而不是回一個
	// 沒有碼、沒有訊息的錯誤回應（「未註冊的碼不可用」要能成立，不能只是不建議）。
	if c.id == "" {
		panic("errcode: 使用了未經 MustRegister 的零值 Code（未註冊的碼不得建構錯誤）")
	}
	msg := c.Message()
	var missing []string
	for _, m := range placeholderPattern.FindAllStringSubmatch(c.Message(), -1) {
		v, ok := params[m[1]]
		if !ok {
			missing = append(missing, m[0])
			continue
		}
		msg = strings.ReplaceAll(msg, m[0], v)
	}
	if len(missing) > 0 {
		log.Printf("errcode: %s 訊息缺參數(已回樣板原文): %q 缺=%v params=%v",
			c.id, c.Message(), missing, params)
	}
	return msg
}

// Error 產生帶 ErrorInfo detail 的 connect error（碼與 details 可被客戶端讀取）。
//
// 不收 ctx：trace_id 由 requestid interceptor 在**回應邊界**補進 ErrorInfo（見 Task 3），
// 故本套件不依賴 internal/obs；「當前請求的 trace_id」本來就只有邊界知道。
func (c Code) Error(params map[string]string) *connect.Error {
	msg := c.Render(params)
	return c.attachInfo(connect.NewError(c.connectCode, errors.New(msg)), msg, params)
}

// Wrap 同 Error，但保留底層錯誤（Unwrap）供 log 追查；**cause 的文字不進對外 message**。
// 為什麼不用 fmt.Errorf("%s: %w", msg, cause)：那會把根因文字寫進 message，而 connect-go
// 對任何 code 都逐字轉送 Message()（客戶端會收到 SQLSTATE／policy 名／SQL 片段）。
func (c Code) Wrap(cause error, params ...map[string]string) *connect.Error {
	var p map[string]string
	if len(params) > 0 {
		p = params[0]
	}
	msg := c.Render(p)
	err := connect.NewError(c.connectCode, &wrapped{msg: msg, cause: cause})
	return c.attachInfo(err, msg, p)
}

// attachInfo 附掛 ErrorInfo detail。附掛失敗不得讓錯誤處理本身失效（記 log，仍回 connect 碼與訊息）——
// 錯誤路徑上的二次失敗最難追。
func (c Code) attachInfo(err *connect.Error, msg string, params map[string]string) *connect.Error {
	info := &commonv1.ErrorInfo{Code: c.id, Message: msg}
	if len(params) > 0 {
		info.Details = params
	}
	detail, derr := connect.NewErrorDetail(info)
	if derr != nil {
		log.Printf("errcode: %s 附掛 ErrorInfo 失敗（回應仍帶 connect 碼）: %v", c.id, derr)
		return err
	}
	err.AddDetail(detail)
	return err
}

// wrapped 讓對外訊息與根因分離：Error() 只回對外訊息，Unwrap() 保留根因供 log 追查。
type wrapped struct {
	msg   string
	cause error
}

func (e *wrapped) Error() string { return e.msg }
func (e *wrapped) Unwrap() error { return e.cause }
