// Package auth 的 server-to-server 靜態 API token 認證(01 計畫 Task 11 / 細部 1.6.6)。
//
// 定位:供**內部排程與未來 agent 協定適配層**以 HTTP 呼叫後端時使用(M2M),不開放給
// Web/App 客戶端 —— 客戶端一律走 session(cookie)或 JWT(Bearer)。設計書 §3.3 明訂
// 「API access token 僅供 server-to-server 呼叫...不得配置於 Web / App 客戶端」。
//
// 三個刻意的設計選擇:
//
//  1. **只存雜湊**:設定值是 token 的 SHA-256(hex),原文不落設定檔／環境變數 ——
//     環境變數常被 dump 進日誌或 CI 記錄,存原文等於存明文憑證。
//  2. **綁真實使用者**:token 映射到既有 users.id,不建「幽靈機器帳號」。原因有二 ——
//     audit_logs.user_id 有 FK 到 users(id)(00010),幽靈 id 寫稽核會被 FK 擋;
//     而重用既有使用者讓權限判定(OpenFGA/RLS)自動與該使用者的角色一致,不必另立一套。
//  3. **RPC 前綴白名單**:每個 token 只能叫自己那組 RPC。空清單 = 不允許任何 RPC
//     (fail-closed):設定漏寫前綴時寧可什麼都不能做,也不要意外拿到全站權限。
package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// APIToken 為單一 server-to-server token 的設定(雜湊與授權範圍)。
type APIToken struct {
	// Name 為 token 的識別名(稽核與日誌用;不出現在對外錯誤訊息)。
	Name string `json:"name"`
	// SHA256 為 token 原文的 SHA-256(hex,大小寫不拘)。
	SHA256 string `json:"sha256"`
	// UserID 為該 token 對應的使用者(身分、權限、稽核歸屬皆以此人為準)。
	UserID int `json:"user_id"`
	// RPCPrefixes 為允許呼叫的 Connect-RPC path 前綴(剝除 /api/v1 後的路徑)。
	// 空 = 不允許任何 RPC(fail-closed)。
	RPCPrefixes []string `json:"rpc_prefixes"`
}

// ParseAPITokens 解析設定中的 JSON 陣列。空字串 = 未啟用(回 nil,不報錯)。
// 格式錯誤或欄位不合法一律回錯 —— 設定壞掉必須在啟動時大聲失敗,不可靜默降級成
// 「token 全部無效」(那會讓排程在半夜無聲停擺)。
func ParseAPITokens(raw string) ([]APIToken, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	var toks []APIToken
	if err := json.Unmarshal([]byte(trimmed), &toks); err != nil {
		return nil, fmt.Errorf("API_TOKENS 不是合法 JSON 陣列: %w", err)
	}
	for i := range toks {
		t := &toks[i]
		if strings.TrimSpace(t.Name) == "" {
			return nil, fmt.Errorf("API_TOKENS[%d] 缺少 name", i)
		}
		if _, err := hex.DecodeString(strings.TrimSpace(t.SHA256)); err != nil {
			return nil, fmt.Errorf("API_TOKENS[%d](%s) 的 sha256 不是合法 hex", i, t.Name)
		}
		if t.UserID <= 0 {
			return nil, fmt.Errorf("API_TOKENS[%d](%s) 的 user_id 必須為正整數(需對應真實使用者)", i, t.Name)
		}
	}
	return toks, nil
}

// MatchAPIToken 比對呈現的 token 原文,回傳命中的設定(未命中回 nil)。
//
// 以 SHA-256 + crypto/subtle 常數時間比對:避免以回應時間逐字元推測雜湊。
// 全部候選都比對(不提早 return),讓命中位置不影響耗時。
func MatchAPIToken(tokens []APIToken, presented string) *APIToken {
	if presented == "" || len(tokens) == 0 {
		return nil
	}
	sum := sha256.Sum256([]byte(presented))
	got := hex.EncodeToString(sum[:])
	var hit *APIToken
	for i := range tokens {
		want := strings.ToLower(strings.TrimSpace(tokens[i].SHA256))
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 && hit == nil {
			hit = &tokens[i]
		}
	}
	return hit
}

// AllowsRPC 判斷該 token 是否允許呼叫指定 RPC path(前綴比對)。
// 前綴清單為空 → 一律 false(fail-closed,見檔頭第 3 點)。
func (t *APIToken) AllowsRPC(path string) bool {
	if t == nil {
		return false
	}
	for _, p := range t.RPCPrefixes {
		p = strings.TrimSpace(p)
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
