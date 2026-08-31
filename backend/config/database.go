package config

// Database PostgreSQL 連線設定。
type Database struct {
	DatabaseURL string `envconfig:"DATABASE_URL" default:"postgres://postgres:postgres@localhost:5432/salesorder?sslmode=disable"`
}
