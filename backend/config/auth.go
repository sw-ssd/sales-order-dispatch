package config

import "time"

// DefaultJWTSecret 為 development 預設 JWT 密鑰(與 struct tag default 同值)。
// production 啟動防護(Server.Init)以它比對,正式環境必須以 JWT_SECRET 覆寫。
const DefaultJWTSecret = "dev-only-jwt-secret-change-me"

// Auth 認證相關設定（OAuth、JWT、Web session）。
type Auth struct {
	GoogleClientID     string        `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string        `envconfig:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string        `envconfig:"GOOGLE_REDIRECT_URL" default:"http://localhost:3080/api/v1/auth/google/callback"`
	GoogleHostedDomain string        `envconfig:"GOOGLE_HOSTED_DOMAIN"` // 限定 Workspace 網域（可留空 = 不限制）
	JWTSecret          string        `envconfig:"JWT_SECRET" default:"dev-only-jwt-secret-change-me"`
	FrontendURL        string        `envconfig:"FRONTEND_URL" default:"http://localhost:3000"`
	SessionLifetime    time.Duration `envconfig:"SESSION_LIFETIME" default:"720h"` // Web session cookie 效期（30 天）
	SessionSecure      bool          `envconfig:"SESSION_SECURE" default:"false"`
	SessionSameSite    string        `envconfig:"SESSION_SAME_SITE" default:"lax"` // lax | strict | none

	// APITokens 為 server-to-server 的靜態 token 清單(01 1.6.6),JSON 陣列:
	//
	//	[{"name":"cron","sha256":"<hex>","user_id":42,"rpc_prefixes":["/salesorder.v1.ReportService"]}]
	//
	// 只存**雜湊**(不存原文,避免設定檔／環境變數成為可直接使用的憑證);`user_id` 必須是
	// 真實使用者(稽核 audit_logs.user_id 有 FK 到 users,機器身分不可為幽靈列);
	// `rpc_prefixes` 為允許的 RPC path 前綴(空 = 不允許任何 RPC,避免設定疏漏變成全能 token)。
	// 留空 = 不啟用本認證路徑。
	APITokens string `envconfig:"API_TOKENS"`
}
