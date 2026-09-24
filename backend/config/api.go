package config

// API 伺服器與外部服務端點設定。
type API struct {
	Env          string `envconfig:"ENV" default:"development"`
	Addr         string `envconfig:"API_ADDR" default:":3080"`
	GotenbergURL string `envconfig:"GOTENBERG_URL" default:"http://localhost:3001"`
	// DeveloperAccountEnabled 控制 developer 角色繞過 OpenFGA 授權與 RLS 資料範圍(設計書 §4.4)。
	// development 預設 true;production 必須設為 false(Server.Init 啟動防護)。
	DeveloperAccountEnabled bool `envconfig:"DEVELOPER_ACCOUNT_ENABLED" default:"true"`
}
