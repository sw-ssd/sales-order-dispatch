// 冪等 seeder 入口（D31 cmd 拆分）：建立 7 內建角色與（development）本機開發者帳號。
package main

import (
	"context"
	"log"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
)

func main() {
	cfg := config.New()
	// 連線仍委派 third_party/database(D31),但**必須**經 dbtenant.NewClient 建立:
	// 系統範圍的 SET LOCAL 是 RLS driver 裝飾器在 Tx(ctx) 內套的,未包裝的 ent client
	// (database.OpenEnt)就算包進 dbtenant.SystemScopeTx 也一樣被 WITH CHECK 擋下。
	sqlDB, err := database.OpenSQL(cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("連線資料庫: %v", err)
	}
	client := dbtenant.NewClient(sqlDB)
	defer func() { _ = client.Close() }()

	// seeder 與 server 共用同一初始化函式(D31),連線改用 owner DSN(AdminDSN):
	// seed 與遷移/OpenFGA 同屬平台側寫入。
	ctx := context.Background()
	// seed 一律在**明確的系統範圍**交易內執行(scope=all):roles/role_permissions/users 已
	// ENABLE+FORCE RLS,而 FORCE 讓 table owner 也受 policy 約束(00025/00028 檔頭),故
	// 連 owner 連線都必須先 SET LOCAL app.current_data_scope = 'all' 才寫得進去。生產的 owner
	// 不是 superuser,少了這一層 seed 會以 42501 失敗(cmd/seed 的整合探針釘住此行為)。
	if err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		if err := SeedBuiltinRoles(ctx, tx.Client()); err != nil {
			return err
		}
		if err := SeedBuiltinRolePermissions(ctx, tx.Client()); err != nil {
			return err
		}
		// developer 帳號僅 ENV != production;需既有 dev company 錨點,無則略過提示。
		return SeedDeveloper(ctx, tx.Client(), cfg.API.Env, firstCompanyID(ctx, tx.Client()))
	}); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("seed: 7 內建角色與 role_permissions 已確保（冪等）")
	log.Println("seed: 完成")
}
