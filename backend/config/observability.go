package config

// Observability 日誌與指標設定。
type Observability struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
