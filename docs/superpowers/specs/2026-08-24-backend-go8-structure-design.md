# 後端架構 go8 對齊設計：結構與工具鏈

> - 日期：2026-08-24
> - 狀態：已核准（brainstorming 對話逐段確認）
> - 範圍：1.0 後端工程慣例（Phase 0 骨架 + 全部 backend 計畫的路徑/啟動/config 約定）
> - 參考：[go8 — Go API starter kit](https://github.com/sowiner/go8)（canonical：codeberg.com/gmhafiz/go8）
> - 將產生決策記錄 **D31**；客戶版規格書**不升版**（內部工程慣例，不影響客戶可見行為）

## 0. 背景與定位

以 go8 starter kit 重新檢視後端計畫後，確認現行計畫在分層（handler→usecase→repository）、測試分層（mock 單測 + testcontainers 整合）、Taskfile、CI lint 上已與 go8 同構；缺口集中於**專案結構慣例**（DI、cmd、config、third_party）與**開發工具鏈**（hot reload、一鍵檢查、漏洞掃描）。

本設計僅做結構/工具對齊。凍結決策不動：D3/D30 權限三層+CASL 執行層、D4 Connect-RPC 唯一 API 來源、D5 scs+Valkey session / App JWT、D19 Prometheus+Grafana+Alertmanager。

## 1. 決策摘要（將寫入決策記錄為 D31）

| # | 決策 | 一句話 |
|---|---|---|
| D31 | 採納 go8 結構慣例 | 集中 DI（`internal/server`：Server struct + InitDomains）、cmd 拆分（server/migrate/seed）、config 逐檔 struct（envconfig）、`third_party/` 集中外部套件初始化 |
| D31-2 | 開發工具鏈補齊 | air hot reload（`.air.toml`）、Taskfile `check`（fmt+vet+lint+test）與 `vuln`（govulncheck）；CI Go job 加 govulncheck step |

**不採納項與理由**：

| go8 元素 | 不採納理由 |
|---|---|
| e2e 臨時容器 harness | D21 整合測試（testcontainers）+ Phase 8 上線驗收已覆蓋；Big Bang 前的端對端信心由既有兩層承擔，1.0 不再加第三層維護成本 |
| OTel traces/logs（Jaeger/Loki/otel-collector） | D19 已定 metrics-only（Prometheus/Grafana/Alertmanager）；加 tracing/日誌聚合擴大 infra 範圍與維運成本，1.0 不做 |
| protovalidate / go-playground/validator | Connect-RPC 下輸入已是 proto 強型別；業務規則驗證由 usecase 層承擔（既有計畫慣例），不引額外驗證框架 |
| sqlx | 與 Ent 職責重疊；複雜查詢以 Ent `sql.ExprP` / raw SQL escape hatch 承擔 |
| go8 REST + swaggo | 衝突 D4（Connect-RPC 唯一 API 來源；REST 僅公開端點） |
| go8 DB-backed session 表 | 衝突 D5（scs + Valkey；App 走 JWT + refresh 旋轉） |
| `cmd/route` 路由清單 | Connect-RPC 下 API 面由 proto 定義，`buf` 產物與 proto 檔即清單；無另行工具需求 |
| go8 介面分層子套件（repository/postgres） | 現行 9 份 detail 計畫已定 flat domain package（`handler.go`/`usecase.go` 同套件）；go8 風子套件對 17 份計畫文件 churn 過大，收益不成比例 |

## 2. 目錄結構

```
backend/
├── cmd/
│   ├── server/main.go      # 既有；改為極薄入口（§3）
│   ├── migrate/main.go     # 新：goose up/down/create CLI
│   └── seed/main.go        # 新：冪等 seeder 入口（roles/預設 policies/developer 帳號）
├── config/
│   ├── config.go           # Config 聚合 struct + New()
│   ├── api.go              # ENV、API_ADDR、DEVELOPER_ACCOUNT_ENABLED、CASL_ENFORCEMENT_ENABLED
│   ├── database.go         # DB_*；連線池大小註明公式 (core×2)+spindle
│   ├── cache.go            # Valkey 連線、session TTL
│   ├── auth.go             # OAuth client/hd 白名單、JWT_SECRET、X-Api-Token 清單
│   └── observability.go    # log 等級、Prometheus
├── internal/
│   ├── server/
│   │   ├── server.go       # Server struct 註冊全部依賴 + Init()
│   │   └── domains.go      # InitDomains()：逐 domain 組裝 repo→usecase→handler
│   ├── domain/…            # 不變（9 份 detail 既有路徑）
│   ├── middleware/  authz/  session/  database/  audit/   # 不變
├── third_party/
│   ├── database/ent.go     # Ent client + pgx pool 初始化
│   └── cache/valkey.go     # Valkey client 初始化
├── proto/  ent/  database/migrations/   # 不變
├── .air.toml               # 新
└── Taskfile.yml            # 補 dev/check/vuln/migrate/seed
```

原則：現行計畫已定義的 `internal/` 路徑一律不動；新慣例只落在 `cmd/`、`config/`、`internal/server/`、`third_party/` 四處。

## 3. DI 與啟動流程

- `cmd/server/main.go` 極薄：`config.New()` → `server.New(cfg)` → `s.Init()` → `s.Run()`。
- `Server` struct 持有：Config、`*ent.Client`、Valkey client、Casbin enforcer、casl `FieldRegistry`、scs manager、chi router——一眼可見伺服器全部依賴。
- `Init()` 集中全部 fail-fast 啟動檢查（現行計畫散落各處者歸位）：DB/Valkey 連線、Casbin enforcer 初始化與預設 policy seed、`JWT_SECRET` 存在、production+`DEVELOPER_ACCOUNT_ENABLED` 拒啟（1.11.1）、CASL fixture 自驗（D30-3）。
- `InitDomains()` 每 domain 一行組裝鏈（repo→usecase→handler），Connect handler 掛 chi router；新增 domain 只動此一檔。

## 4. config 逐檔 struct

- go8 同款 envconfig 風格：`github.com/kelseyhightower/envconfig`；每檔一個 struct + constructor（namespace 前綴），`config.New()` 聚合，取用有 code completion。
- 現行計畫所有 config key 按 §2 分檔歸位；`.env.example` 同步維護，`.env` 於 `.gitignore`（既有）。

## 5. 工具鏈

- Taskfile 新增：`dev`（air，`.air.toml` watch `backend/`、build `./cmd/server`）、`check`（fmt+vet+lint+test 一鍵）、`vuln`（`govulncheck ./...`）、`migrate` / `migrate:create` / `seed`。
- CI Go job（subproject plan 既有 lint+test+build）補 `govulncheck ./...` step。
- golangci-lint CI 已有；本地 `task lint` 補上。

## 6. 文件修訂範圍

| 文件 | 動作 |
|---|---|
| 本文件 | 新增 |
| `2026-07-19-sales-order-1.0-decisions.md` | 新增 D31/D31-2 |
| `backend-detail/00-index.md` | 新增「架構慣例」一節（目錄樹、DI、config、third_party、工具鏈） |
| `backend-detail/01-auth.md` | config 路徑分檔；1.11.1 啟動檢查歸 `server.Init()` |
| `backend-detail/02–09` | config 提及處批次同步（grep 驅動） |
| 8 份 `2026-08-17-backend-0X-*-plan.md` | Global Constraints/File Structure/config 與 cmd 路徑批次修訂 |
| `2026-08-24-casl-integration-plan.md` | Task 8 `config/config.go` → `config/api.go` |
| `2026-08-05-…-subproject-implementation-plan.md` | Phase 0 backend 骨架 Task 補新目錄與 Taskfile tasks；CI Go job 補 govulncheck |

驗證（docs-only）：修訂後 grep 舊路徑殘留歸零；00-index 目錄樹與各計畫 File Structure 並集一致。

## 7. 非目標

- 不改任何 `internal/domain/` 既有檔案配置與分層寫法。
- 不引入 e2e harness、OTel、protovalidate、sqlx、cmd/route（§1 不採納表）。
- 不動前端/App 計畫、不升客戶版規格書。

---

*最後更新：2026-08-24*
