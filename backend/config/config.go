// Package config 以 envconfig 逐檔 struct 聚合所有環境設定（D31）。
// 新增 key 群組無對應檔時，新建 config/<name>.go 並於此聚合。
package config

// Config 聚合全部設定分檔。
type Config struct {
	API           API
	Auth          Auth
	Cache         Cache
	Database      Database
	Storage       Storage
	Observability Observability
}

// New 由環境變數載入全部設定；解析失敗即 fail-fast。
func New() *Config {
	var c Config
	mustProcess(&c.API)
	mustProcess(&c.Auth)
	mustProcess(&c.Cache)
	mustProcess(&c.Database)
	mustProcess(&c.Storage)
	mustProcess(&c.Observability)
	return &c
}
