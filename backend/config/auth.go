package config

// Auth 認證相關設定（OAuth、JWT 等）。
type Auth struct {
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
}
