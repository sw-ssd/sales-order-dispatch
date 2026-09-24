package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// TestParseAPITokens 設定解析:合法清單通過;空字串 = 未啟用;格式錯誤一律大聲失敗。
//
// 為何要嚴格:API_TOKENS 壞掉若被靜默忽略,後果是「排程的呼叫全部 401」而**沒有明顯症狀**
// (半夜無聲停擺),故寧可在啟動時 Fatal。
func TestParseAPITokens(t *testing.T) {
	// 空 = 未啟用(不報錯)。
	if toks, err := ParseAPITokens(""); err != nil || toks != nil {
		t.Fatalf("空字串應回 nil,nil;got %v,%v", toks, err)
	}
	// 合法。
	raw := `[{"name":"cron","sha256":"` + sha256Hex("s3cret") + `","user_id":7,"rpc_prefixes":["/salesorder.v1.ReportService"]}]`
	toks, err := ParseAPITokens(raw)
	if err != nil {
		t.Fatalf("合法設定不應報錯: %v", err)
	}
	if len(toks) != 1 || toks[0].Name != "cron" || toks[0].UserID != 7 {
		t.Fatalf("解析結果不符: %+v", toks)
	}

	// 以下皆必須失敗(設定錯誤必須擋在啟動)。
	for name, bad := range map[string]string{
		"非 JSON":       `cron:s3cret`,
		"缺 name":       `[{"sha256":"ab","user_id":1}]`,
		"sha256 非 hex": `[{"name":"a","sha256":"zz","user_id":1}]`,
		"user_id 為 0":  `[{"name":"a","sha256":"ab","user_id":0}]`,
		"user_id 負":    `[{"name":"a","sha256":"ab","user_id":-3}]`,
	} {
		if _, err := ParseAPITokens(bad); err == nil {
			t.Errorf("%s 應報錯", name)
		}
	}
}

// TestMatchAPIToken 以 SHA-256 常數時間比對呈現的 token 原文。
func TestMatchAPIToken(t *testing.T) {
	toks, err := ParseAPITokens(`[
		{"name":"cron","sha256":"` + sha256Hex("tok-a") + `","user_id":1,"rpc_prefixes":["/x"]},
		{"name":"agent","sha256":"` + stringToUpper(sha256Hex("tok-b")) + `","user_id":2,"rpc_prefixes":["/y"]}
	]`)
	if err != nil {
		t.Fatalf("ParseAPITokens: %v", err)
	}
	if got := MatchAPIToken(toks, "tok-a"); got == nil || got.Name != "cron" {
		t.Fatalf("tok-a 應命中 cron,got %+v", got)
	}
	// 大小寫不拘(hex 可能以任一大小寫給定)。
	if got := MatchAPIToken(toks, "tok-b"); got == nil || got.Name != "agent" {
		t.Fatalf("tok-b 應命中 agent(hex 大小寫不拘),got %+v", got)
	}
	// 未命中 / 空字串 / 無清單。
	if got := MatchAPIToken(toks, "wrong"); got != nil {
		t.Fatalf("錯的 token 不應命中,got %+v", got)
	}
	if got := MatchAPIToken(toks, ""); got != nil {
		t.Fatal("空 token 不應命中")
	}
	if got := MatchAPIToken(nil, "tok-a"); got != nil {
		t.Fatal("未設定清單時不應命中")
	}
}

// TestAllowsRPC 白名單前綴比對;空清單 fail-closed(不得變成全能 token)。
func TestAllowsRPC(t *testing.T) {
	tok := &APIToken{Name: "cron", RPCPrefixes: []string{"/salesorder.v1.ReportService"}}
	if !tok.AllowsRPC("/salesorder.v1.ReportService/ListReports") {
		t.Fatal("前綴內應允許")
	}
	if tok.AllowsRPC("/salesorder.v1.SalesOrderService/ListOrders") {
		t.Fatal("前綴外應拒絕")
	}
	// 空清單 = 什麼都不允許(設定漏寫前綴時,不得意外拿到全站權限)。
	empty := &APIToken{Name: "broken"}
	if empty.AllowsRPC("/salesorder.v1.ReportService/ListReports") {
		t.Fatal("空白名單必須 fail-closed")
	}
	// nil 受者不得 panic。
	var nilTok *APIToken
	if nilTok.AllowsRPC("/x") {
		t.Fatal("nil token 應拒絕")
	}
}

func stringToUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'f' {
			b[i] -= 32
		}
	}
	return string(b)
}
