// 平台域 RPC 伺服器(phase-2 獨立 HTTP 服務)。
//
// 與 platform-cron 共用同一份 entitlements.Service 組裝(owner 連線 + Valkey 快取);
// 本行程只負責把服務掛上 HTTP mux 並 listen —— 組裝細節不應在 main、Taskfile 與 cron 各寫一遍。
//
// 配額計數器(Counter)本階段用 Unlimited(無限制)占位;platform.usage_counters 落地後
// (Task 2.4)此處注入真實計數器。
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/contracts/cache"
	"github.com/sw-ssd/platform/entitlements"
	"github.com/sw-ssd/platform/server"
	"github.com/sw-ssd/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
)

func main() {
	addr := flag.String("addr", ":8082", "HTTP 監聽位址")
	flag.Parse()

	cfg := config.New()
	db, err := database.OpenSQL(cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("admin 連線: %v", err)
	}
	defer func() { _ = db.Close() }()

	st := postgres.New(db)

	// 權益快取:Valkey 不可用不讓服務失敗(最長 TTL 內讀舊權益,見 entitlements.ValkeyCache 註解)。
	var entCache entitlements.Cache
	valkeyClient := cache.NewClient(cfg.Cache.ValkeyAddr)
	defer func() { _ = valkeyClient.Close() }()
	if pingErr := cache.Ping(context.Background(), valkeyClient); pingErr != nil {
		log.Printf("platform-server: Valkey 不可用(%v) → 權益快取失效停用(最長 TTL 內仍讀舊權益)", pingErr)
	} else {
		entCache = entitlements.NewValkeyCache(valkeyClient)
	}

	// 本階段 Counter 不需注入:CheckLimit 由 product 側帶入 current_used(見 platformquota.Client),
	// platform 不再數業務表;Snapshot 暫未經此服務暴露,nil 無副作用。
	ent := entitlements.New(st, nil, entCache, 0)

	mux := http.NewServeMux()
	server.Register(mux, ent)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("platform-server 監聽 %s", *addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("platform-server: %v", err)
	}
}
