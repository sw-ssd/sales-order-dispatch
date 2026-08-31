// 極薄入口：config.New() → server.New(cfg) → Init() → InitDomains() → Run()（D31）。
package main

import (
	"log"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/server"
)

func main() {
	cfg := config.New()
	s := server.New(cfg)
	if err := s.Init(); err != nil {
		log.Fatalf("啟動檢查失敗: %v", err)
	}
	s.InitDomains()
	log.Printf("listening on %s (env=%s)", cfg.API.Addr, cfg.API.Env)
	log.Fatal(s.Run())
}
