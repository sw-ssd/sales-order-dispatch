package config

// API 伺服器與外部服務端點設定。
type API struct {
	Env          string `envconfig:"ENV" default:"development"`
	Addr         string `envconfig:"API_ADDR" default:":3080"`
	GotenbergURL string `envconfig:"GOTENBERG_URL" default:"http://localhost:3001"`
	// CASLEnforcementEnabled 控制後端 CASL 執行層(list 過濾與實例檢查,D30-2)。
	// 三環境均預設 true;false = 降級回純 Casbin + RLS。
	CASLEnforcementEnabled bool `envconfig:"CASL_ENFORCEMENT_ENABLED" default:"true"`
}
