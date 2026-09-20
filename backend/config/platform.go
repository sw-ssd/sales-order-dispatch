package config

// Platform 為平台營運工具的設定（D38）。整組未設定時平台工具不掛載（開發環境友善），
// production 由 Server.Init() 要求必須設定。
type Platform struct {
	// OperatorJWTSecret 與租戶 JWTSecret **必須不同**：跨用等於平台工具可被租戶 token 冒充。
	OperatorJWTSecret string `envconfig:"PLATFORM_JWT_SECRET"`
	// AllowedEmailDomain 限制 OIDC 登入的 email 網域。
	// 以 email 網域比對（不依賴 Workspace 專屬的 hd claim）→ 相容 Workspace 帳號與
	// 既有 Google 帳號的已驗證別名兩種情況。
	AllowedEmailDomain string `envconfig:"PLATFORM_ALLOWED_EMAIL_DOMAIN" default:"sowinsoft.com"`
	// ConsoleURL 為 OIDC 完成後導回的 console 根網址。
	ConsoleURL string `envconfig:"PLATFORM_CONSOLE_URL"`
	// CookieDomain 為 operator session cookie 的 Domain（空 = host-only，開發環境用）。
	CookieDomain string `envconfig:"PLATFORM_COOKIE_DOMAIN"`

	// --- seed 與排程的預設值（上線前請改為真實值；**執行期以 platform.settings 為準**）---
	// 這些只是「首次建立時寫入 settings」的來源；之後由營運工具調整，重跑 seed 不覆寫。
	SeedOperatorEmail    string `envconfig:"PLATFORM_SEED_OPERATOR_EMAIL" default:"ssd@sowinsoft.com"`
	SeedOperatorName     string `envconfig:"PLATFORM_SEED_OPERATOR_NAME" default:"ssd"`
	SeedSystemActorEmail string `envconfig:"PLATFORM_SEED_SYSTEM_ACTOR_EMAIL" default:"system@sowinsoft.com"`
	DefaultTrialDays     int    `envconfig:"PLATFORM_DEFAULT_TRIAL_DAYS" default:"14"`
	DefaultGraceDays     int    `envconfig:"PLATFORM_DEFAULT_GRACE_DAYS" default:"7"`
	DefaultLeadDays      int    `envconfig:"PLATFORM_DEFAULT_LEAD_DAYS" default:"14"`

	// --- 方案價目的 seed 預設（金額字串，兩位小數）---
	// **這些是佔位數字**：首次建立 `plan_prices` 時使用，之後由營運工具（UpsertPlanPrice）維護；
	// 上線前務必改為真實定價。重跑 seed 不覆寫既有價目。
	SeedPriceFreeBase string `envconfig:"SEED_PRICE_FREE_BASE" default:"0"`
	SeedPriceFreeSeat string `envconfig:"SEED_PRICE_FREE_SEAT" default:"0"`
	SeedPriceStdBase  string `envconfig:"SEED_PRICE_STD_BASE" default:"1500"`
	SeedPriceStdSeat  string `envconfig:"SEED_PRICE_STD_SEAT" default:"150"`
	SeedPriceProBase  string `envconfig:"SEED_PRICE_PRO_BASE" default:"4500"`
	SeedPriceProSeat  string `envconfig:"SEED_PRICE_PRO_SEAT" default:"150"`
}

// Configured 表示必要設定齊備，可掛載平台工具。
// CookieDomain 可為空：同源／代理開發環境使用 host-only cookie（設 Domain=localhost 無效）。
func (p Platform) Configured() bool {
	return p.OperatorJWTSecret != "" && p.AllowedEmailDomain != "" && p.ConsoleURL != ""
}
