// 冪等 seeder 入口（D31 cmd 拆分）。
// roles / 預設 policies / developer 帳號的 seed 於 02-tenancy-users 計畫落地；
// 現階段僅做連線檢查，保證入口可用。
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
	log.Println("seed: 資料庫連線正常；目前無 seed 項目（roles seed 於 02 計畫落地）")
}
