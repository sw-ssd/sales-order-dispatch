package config

// OpenFGA 內嵌授權引擎設定(D32)。內嵌於後端程序,datastore 共用既有 PostgreSQL(單一 store)。
type OpenFGA struct {
	// Enabled 控制 OpenFGA 授權是否啟用;關閉 = 回退(僅 RLS 兜底)。
	Enabled bool `envconfig:"OPENFGA_ENABLED" default:"true"`
	// DatabaseURL 為 OpenFGA datastore 連線(與業務同庫;空則沿用 Database.DatabaseURL)。
	DatabaseURL string `envconfig:"OPENFGA_DATABASE_URL"`
	// StoreName 為 OpenFGA store 名稱。
	StoreName string `envconfig:"OPENFGA_STORE_NAME" default:"sales_order"`
}
