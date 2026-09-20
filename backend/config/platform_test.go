package config

import (
	"os"
	"testing"
)

func TestPlatformFromEnv(t *testing.T) {
	t.Setenv("PLATFORM_JWT_SECRET", "s3cret")
	t.Setenv("PLATFORM_ALLOWED_EMAIL_DOMAIN", "example.com")
	t.Setenv("PLATFORM_CONSOLE_URL", "https://console.example.com")
	t.Setenv("PLATFORM_COOKIE_DOMAIN", ".example.com")
	var p Platform
	mustProcess(&p)
	if p.OperatorJWTSecret != "s3cret" || p.AllowedEmailDomain != "example.com" ||
		p.ConsoleURL != "https://console.example.com" || p.CookieDomain != ".example.com" {
		t.Fatalf("平台設定未正確綁定: %+v", p)
	}
}

// TestNewBindsPlatform 確認 Config.Platform 真的由 New() 載入：
// 漏了 mustProcess(&c.Platform) 時 Platform 恆為零值 → Server.Init() 的平台守護**整段靜默跳過**
// (空 secret 不會等於 JWT_SECRET)。
func TestNewBindsPlatform(t *testing.T) {
	t.Setenv("PLATFORM_JWT_SECRET", "s3cret")
	t.Setenv("PLATFORM_CONSOLE_URL", "https://console.example.com")
	c := New()
	if c.Platform.OperatorJWTSecret != "s3cret" || c.Platform.ConsoleURL != "https://console.example.com" {
		t.Fatalf("Config.Platform 未由 config.New() 載入: %+v", c.Platform)
	}
}

// TestPlatformSeedOperatorEmailHasNoDefault 釘住「首位 operator 的 email／name 不得有預設值」：
// 它是 platform.operators 白名單（＝能登入平台工具的人），有預設值就等於把版控裡的真實 email
// 種成可登入帳號（controller 的硬要求）。把 `default:"…"` 加回 struct tag 必須讓本測試轉紅。
//
// envconfig **只在環境變數不存在時**才套用 default（明確設成空字串會覆蓋預設值），
// 故這裡必須真的 unset —— t.Setenv 只能設值。
func TestPlatformSeedOperatorEmailHasNoDefault(t *testing.T) {
	for _, key := range []string{"PLATFORM_SEED_OPERATOR_EMAIL", "PLATFORM_SEED_OPERATOR_NAME"} {
		prev, had := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if had {
				_ = os.Setenv(key, prev)
			}
		})
	}
	p := New().Platform
	if p.SeedOperatorEmail != "" || p.SeedOperatorName != "" {
		t.Fatalf("首位 operator 的 email／name 不得有預設值（未設時 seed 必須跳過），got %q／%q",
			p.SeedOperatorEmail, p.SeedOperatorName)
	}
}

func TestPlatformConfigured(t *testing.T) {
	if (Platform{}).Configured() {
		t.Fatal("空設定不得視為已設定")
	}
	// CookieDomain 可為空（開發環境同源代理時用 host-only cookie）；
	// 其餘三項缺一即不掛載。
	noCookie := Platform{
		OperatorJWTSecret: "s", AllowedEmailDomain: "example.com",
		ConsoleURL: "http://localhost:5173",
	}
	if !noCookie.Configured() {
		t.Fatal("CookieDomain 可為空，三項齊備即視為已設定")
	}
	missingSecret := Platform{AllowedEmailDomain: "example.com", ConsoleURL: "http://localhost:5173"}
	if missingSecret.Configured() {
		t.Fatal("缺 OperatorJWTSecret 不得視為已設定")
	}
}
