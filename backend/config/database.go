package config

// Database PostgreSQL 連線設定。
type Database struct {
	DatabaseURL string `envconfig:"DATABASE_URL" default:"postgres://postgres:postgres@localhost:5432/salesorder?sslmode=disable"`
	// AdminURL 為 owner 連線(goose 遷移、seed、OpenFGA、平台域)；業務連線改走非 owner 的 app_rw，
	// 使 RLS 對業務生效。空則沿用 DatabaseURL(單一角色的開發/測試環境)。
	AdminURL string `envconfig:"DATABASE_ADMIN_URL"`
}

// AdminDSN 回傳 owner 連線字串；未設定 DATABASE_ADMIN_URL 時沿用業務 DSN。
func (d Database) AdminDSN() string {
	if d.AdminURL != "" {
		return d.AdminURL
	}
	return d.DatabaseURL
}
