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

	// 平台域基礎資料(D34／G5)。連線分兩段,SeedPlatform 內註解有完整說明:
	//   - platform.*(features／方案／價目／權益／operator／settings)**不套 RLS**,走上面這條
	//     owner 連線 sqlDB;
	//   - 平台自營公司與系統使用者是業務表(companies／users,00028 FORCE RLS)→ SeedPlatform
	//     自己開系統範圍交易,故必須把 dbtenant.NewClient 建立的 client 傳進去
	//     (裸 client 包 SystemScopeTx 一樣 42501,已實測)。
	// 排在 roles／developer 之後:新庫的 developer 帳號錨點(firstCompanyID)才不會挑到平台自營公司。
	if err := SeedPlatform(ctx, sqlDB, client, cfg.Platform); err != nil {
		log.Fatalf("seed 平台域: %v", err)
	}
	log.Println("seed: 平台域已確保（7 features／3 方案與價目／首位 operator 依 env／平台自營公司與系統使用者，冪等）")
	log.Println("seed: 完成")
}
