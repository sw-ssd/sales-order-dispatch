package config

// Cache Valkey 連線設定。
type Cache struct {
	ValkeyAddr string `envconfig:"VALKEY_ADDR" default:"localhost:6379"`
}
