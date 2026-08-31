package config

import "time"

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
}
