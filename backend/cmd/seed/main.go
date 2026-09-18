// 冪等 seeder 入口（D31 cmd 拆分）：建立 7 內建角色與（development）本機開發者帳號。
package main

import (
	"context"
	"log"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
)

func main() {
	cfg := config.New()
	pool, err := database.Open(context.Background(), cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatalf("連線資料庫: %v", err)
	}
	defer pool.Close()

	// seeder 以 ent client 操作角色/使用者;與 server 共用同一初始化路徑(D31)。
	client, err := database.OpenEnt(cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatalf("開啟 ent client: %v", err)
	}
	ctx := context.Background()
	if err := SeedBuiltinRoles(ctx, client); err != nil {
		log.Fatalf("seed 角色: %v", err)
	}
	if err := SeedBuiltinRolePermissions(ctx, client); err != nil {
		log.Fatalf("seed 角色權限: %v", err)
	}
	log.Println("seed: 7 內建角色與 role_permissions 已確保（冪等）")
	// developer 帳號僅 ENV != production;需既有 dev company 錨點,無則略過提示。
	if err := SeedDeveloper(ctx, client, cfg.API.Env, firstCompanyID(ctx, client)); err != nil {
		log.Fatalf("seed developer: %v", err)
	}
	log.Println("seed: 完成")
}
