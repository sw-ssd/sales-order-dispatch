package config

// API 伺服器與外部服務端點設定。
type API struct {
	Env          string `envconfig:"ENV" default:"development"`
	Addr         string `envconfig:"API_ADDR" default:":3080"`
	GotenbergURL string `envconfig:"GOTENBERG_URL" default:"http://localhost:3001"`
}
