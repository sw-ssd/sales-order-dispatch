// Package server 集中 DI：Server struct 持有共享依賴，
// Init() 收斂 fail-fast 啟動檢查，InitDomains() 逐 domain 組裝（D31）。
package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/salesorder/sales-order-1.0/backend/config"
)

// Version 目前版本；正式版號由建置 ldflags 注入。
var Version = "0.1.0-dev"

// Server 持有全部共享依賴與路由。
type Server struct {
	cfg    *config.Config
	router *chi.Mux
}

// New 建立 Server 並掛上全域 middleware 與基礎路由。
func New(cfg *config.Config) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s := &Server{cfg: cfg, router: r}
	r.Get("/api/v1/version", s.handleVersion)
	return s
}

// Init 執行 fail-fast 啟動檢查（DB/Valkey 連線、必要 secret 等，於各 domain 計畫補上）。
func (s *Server) Init() error {
	return nil
}

// Run 啟動 HTTP 服務。
func (s *Server) Run() error {
	return http.ListenAndServe(s.cfg.API.Addr, s.router)
}

// Handler 暴露路由供測試與掛載。
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"version": Version})
}
