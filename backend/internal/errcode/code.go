// Package errcode 為對外錯誤碼的唯一真相來源。
// 設計要點：常數即註冊（MustRegister 於套件 init 驗證格式／唯一性／區段與 connect 碼的一致性，
// 違反即 panic → 啟動就失敗，而非上線後才發現）；碼發佈後不得重用或改義，廢止只標 Deprecated。
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
type Code struct {
	ID          string
	Domain      Domain
	ConnectCode connect.Code // 對外 connect 碼（由區段決定，MustRegister 驗證）
	Message     string       // 繁中樣板，可含 {name} 參數
	Deprecated  bool
}

var (
	idPattern = regexp.MustCompile(`^[A-Z]{2,6}-(\d{4})$`)
	registry  = map[string]Code{}
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

// MustRegister 註冊一個碼；違反下列任一即 panic（啟動時暴露，而非上線後才發現）：
// 格式不符、ID 重複、缺訊息、domain 與前綴不符、connect 碼與區段規則不符。
func MustRegister(c Code) Code {
	if !idPattern.MatchString(c.ID) {
		panic(fmt.Sprintf("errcode: 碼格式錯誤 %q（應為 域-4位數）", c.ID))
	}
	if _, dup := registry[c.ID]; dup {
		panic(fmt.Sprintf("errcode: 碼重複註冊 %q", c.ID))
	}
	if strings.TrimSpace(c.Message) == "" {
		panic(fmt.Sprintf("errcode: %q 缺訊息", c.ID))
	}
	if !strings.HasPrefix(c.ID, string(c.Domain)+"-") {
		panic(fmt.Sprintf("errcode: %q 與 domain %q 不符", c.ID, c.Domain))
	}
	if !SectionAllows(c.ID, c.ConnectCode) {
		panic(fmt.Sprintf("errcode: %q 的 connect 碼 %v 與區段 %s 的允許集合不符",
			c.ID, c.ConnectCode, c.ID[:1]))
	}
	registry[c.ID] = c
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
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Render 以參數渲染訊息樣板；缺參數時回樣板原文並記 log（不得讓錯誤處理本身爆掉）。
func (c Code) Render(params map[string]string) string {
	msg := c.Message
	for k, v := range params {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	if strings.Contains(msg, "{") {
		log.Printf("errcode: %s 訊息缺參數(已回樣板原文): %q params=%v", c.ID, c.Message, params)
	}
	return msg
}

// Error 產生帶 ErrorInfo detail 的 connect error（碼與 details 可被客戶端讀取）。
//
// 不收 ctx：trace_id 由 requestid interceptor 在**回應邊界**補進 ErrorInfo（見 Task 3），
// 故本套件不依賴 internal/obs；「當前請求的 trace_id」本來就只有邊界知道。
func (c Code) Error(params map[string]string) *connect.Error {
	msg := c.Render(params)
	return c.attachInfo(connect.NewError(c.ConnectCode, errors.New(msg)), msg, params)
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
	err := connect.NewError(c.ConnectCode, &wrapped{msg: msg, cause: cause})
	return c.attachInfo(err, msg, p)
}

// attachInfo 附掛 ErrorInfo detail。附掛失敗不得讓錯誤處理本身失效（記 log，仍回 connect 碼與訊息）——
// 錯誤路徑上的二次失敗最難追。
func (c Code) attachInfo(err *connect.Error, msg string, params map[string]string) *connect.Error {
	info := &commonv1.ErrorInfo{Code: c.ID, Message: msg}
	if len(params) > 0 {
		info.Details = params
	}
	detail, derr := connect.NewErrorDetail(info)
	if derr != nil {
		log.Printf("errcode: %s 附掛 ErrorInfo 失敗（回應仍帶 connect 碼）: %v", c.ID, derr)
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
