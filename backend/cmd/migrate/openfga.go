package main

import (
	"context"
	"fmt"
	"log"

	serverconfig "github.com/openfga/openfga/pkg/server/config"
	ofgamigrate "github.com/openfga/openfga/pkg/storage/migrate"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/config"
)

// openFGAGooseTable 為 OpenFGA schema 專屬的 goose 版本表名。
// OpenFGA 與業務 schema 共用同一顆 DB,若共用預設 `goose_db_version`(業務已到 17),
// 官方遷移會判定「found N missing migrations before current version」而拒絕(F2)。
const openFGAGooseTable = "openfga_goose_db_version"

// migrateOpenFGA 以 OpenFGA 官方公開遷移入口(pkg/storage/migrate.RunMigrations,
// doc 明示支援嵌入方自管 schema)建立/升級 OpenFGA datastore 表。
// 版本表為 goose 全域狀態,故設定後還原,避免汙染業務遷移;
// dsn 與 server 端一致:OpenFGA.DatabaseURL 空則沿用 owner DSN(AdminDSN)。
// OPENFGA_ENABLED=false 時跳過(config.OpenFGA.Enabled 語意:關閉 = 回退)。
func migrateOpenFGA(ctx context.Context, cfg *config.Config) error {
	if !cfg.OpenFGA.Enabled {
		log.Println("migrate: OPENFGA_ENABLED=false,略過 OpenFGA schema 遷移")
		return nil
	}
	dsn := cfg.OpenFGA.DatabaseURL
	if dsn == "" {
		dsn = cfg.Database.AdminDSN()
	}
	prevTable := goose.TableName()
	goose.SetTableName(openFGAGooseTable)
	defer func() {
		goose.SetTableName(prevTable)
		// RunMigrations 會把 base FS 設為 OpenFGA 內嵌 assets;還原為 os 檔案系統。
		goose.SetBaseFS(nil)
	}()

	if err := ofgamigrate.RunMigrations(ofgamigrate.MigrationConfig{
		Engine:      "postgres",
		URI:         dsn,
		Timeout:     serverconfig.DefaultDatastorePingRetryMaxElapsedTime,
		PingTimeout: serverconfig.DefaultDatastorePingTimeout,
	}); err != nil {
		return fmt.Errorf("OpenFGA schema 遷移失敗(版本表 %s): %w", openFGAGooseTable, err)
	}
	log.Printf("migrate: OpenFGA schema 已就緒(版本表 %s)", openFGAGooseTable)
	return nil
}
