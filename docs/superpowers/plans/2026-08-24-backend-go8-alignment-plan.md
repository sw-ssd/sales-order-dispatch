# 後端架構 go8 對齊（文件修訂）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 將 D31 五項 go8 結構/工具慣例（集中 DI、cmd 拆分、config 逐檔 envconfig struct、third_party/、air+check+govulncheck）編碼進全部既有規劃文件，消除新路徑與既有計畫的衝突。

**Architecture:** 全部為文件修訂，無程式碼產出。核心策略 = 00-index 新增「架構慣例」章節作為唯一權威（含術語映射表），其餘文件以五條轉換規則（R1–R5）批次套用，每條規則附 grep 驗證。不逐處改寫既有程式碼片段的業務內容，只改路徑/tag/函式名/組裝術語。

**Tech Stack:** 文件編修；驗證用 grep（ripgrep 語法）。

**Spec 來源:** `docs/superpowers/specs/2026-08-24-backend-go8-structure-design.md`（下稱「設計」）；參考 https://github.com/sowiner/go8。

## Global Constraints

- 凍結決策不動：D3/D30 權限、D4 Connect-RPC、D5 session/JWT、D19 監控。客戶版規格書不升版。
- 現行計畫已定義的 `backend/internal/` 路徑一律不改；新慣例只落 `cmd/`、`config/`、`internal/server/`、`third_party/`。
- config key 歸檔映射表（R1 的依據，全文一致）：

| key 群組 | 目標檔 |
|---|---|
| OAuth client/secret/redirect/hd、JWT_SECRET、AccessTokenTTL、X-Api-Token 清單 | `backend/config/auth.go` |
| Valkey 連線、session TTL | `backend/config/cache.go` |
| ENV、API_ADDR、DEVELOPER_ACCOUNT_ENABLED、CASL_ENFORCEMENT_ENABLED、BOARD_HEARTBEAT_SECONDS、INGRESS_IDLE_TIMEOUT_SECONDS、Gotenberg URL | `backend/config/api.go` |
| DB_* | `backend/config/database.go` |
| STORAGE_ROOT | `backend/config/storage.go` |
| log 等級、Prometheus | `backend/config/observability.go` |
| 日後新 key 群組無對應檔 | 比照 go8：新建 `config/<name>.go` 並於 `config.go` 聚合 |

- 轉換規則（R2–R5 適用所有計畫文件，含程式碼片段內）：
  - **R1**：`backend/config/config.go` → 依上表分檔目標；`config.go` 僅餘聚合 struct 與 `New()`。
  - **R2**：struct tag `mapstructure:"KEY"` → `envconfig:"KEY"`；config 套件改用 `github.com/kelseyhightower/envconfig`。
  - **R3**：`config.Load()` → `config.New()`。
  - **R4**：「`main.go` 組裝點/掛載/組裝處」→「`InitDomains()` 組裝點（`internal/server/domains.go`）」；「`main.go` 啟動檢查/啟動防護/啟動序列」→「`Server.Init()`（`internal/server/server.go`）」。
  - **R5**：各計畫 File Structure 表與 Global Constraints 提及上述路徑處同步。
- 每 Task 結尾 commit：`docs: …`。
- 驗證一律用 grep（工具：`grep` 指令或等效），不接受目視。

## File Structure（修訂目標清單）

| 檔案 | 動作 | Task |
|---|---|---|
| `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md` | 追加 D31/D31-2 | 1 |
| `docs/superpowers/plans/backend-detail/00-index.md` | 新增 §3.7 架構慣例 | 2 |
| `docs/superpowers/plans/backend-detail/01-auth.md` | 五處 config/main.go 錨點 | 3 |
| `docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md` | Phase 0 骨架 + Task 1.11 config 路徑 | 4 |
| `docs/superpowers/plans/2026-08-05-sales-order-1.0-subproject-implementation-plan.md` | 目錄樹 + Phase 0 files + CI govulncheck | 5 |
| `docs/superpowers/plans/2026-08-17-backend-01-auth-plan.md` | R1–R5 批次 | 6 |
| `docs/superpowers/plans/2026-08-17-backend-0{2,3,4,7,8,9}-*-plan.md` | R1–R5 批次 | 7 |
| `docs/superpowers/plans/2026-08-17-backend-0{5,6}-*-plan.md` | grep 驗證無 config/cmd 錨點，不修訂 | 9（驗證覆蓋） |
| `docs/superpowers/plans/2026-08-24-casl-integration-plan.md` | Task 8 config 路徑 | 8 |
| 全庫 | 殘留 grep 驗證 | 9 |

註：`backend-detail/02–09` 經 grep 確認**無** config/cmd 路徑提及，不需修訂（設計 §6 該列落地為「驗證後無需修改」）。

---

### Task 1: 決策記錄 D31

**Files:**
- Modify: `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（檔尾追加）

- [ ] **Step 1: 追加 D31 區塊**

```markdown
### D31：後端結構慣例對齊 go8（集中 DI / cmd 拆分 / config 逐檔 / third_party）

- **選擇**：採納 go8 四項結構慣例——`internal/server`（Server struct + `Init()` + `InitDomains()` 集中 DI）、cmd 拆分（`cmd/server` 極薄入口 + `cmd/migrate` goose CLI + `cmd/seed` 冪等 seeder）、config 逐檔 struct（`kelseyhightower/envconfig`，`config.New()` 聚合）、`third_party/` 集中外部套件初始化（Ent/pgx、Valkey）。
- **D31-2 工具鏈**：air hot reload（`.air.toml`）；Taskfile 增 `dev`/`check`（fmt+vet+lint+test）/`vuln`（govulncheck）/`migrate`/`seed`；CI Go job 加 govulncheck。
- **不採納**：e2e 臨時容器 harness（D21 整合測試 + Phase 8 驗收覆蓋）；OTel traces/logs（D19 已定 metrics-only）；protovalidate/validator（proto 強型別 + usecase 驗證承擔）；sqlx（與 Ent 重疊）；go8 REST/DB-session 風格（衝突 D4/D5）；`cmd/route`（API 面由 proto 定義）；go8 介面子套件分層（對 17 份計畫 churn 過大）。
- **理由**：集中 DI 讓啟動依賴與 fail-fast 檢查一目瞭然；cmd 拆分讓 migrate/seed 不依賴 server 啟動；config 逐檔有 code completion 且新增 key 有明確流程。
- **修訂來源**：2026-08-24 設計文件 `docs/superpowers/specs/2026-08-24-backend-go8-structure-design.md`；參考 https://github.com/sowiner/go8。
```

- [ ] **Step 2: Commit**

```bash
git add docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md
git commit -m "docs: D31 後端結構慣例對齊 go8"
```

---

### Task 2: 00-index 新增「架構慣例」章節

**Files:**
- Modify: `docs/superpowers/plans/backend-detail/00-index.md`

**Interfaces:**
- Produces: 架構慣例為全部 backend 計畫的權威；R4 術語映射表讓既有 8 份 TDD 計畫的「main.go 組裝點」表述不需逐處改寫即可對應新結構。

- [ ] **Step 1: 於 §3 共通規則後插入 §3.7（或檔案慣用的下一編號）**

```markdown
## 3.7 架構慣例(D31,2026-08-24 起生效)

### 目錄結構

\```
backend/
├── cmd/
│   ├── server/main.go      # 極薄入口:config.New() → server.New(cfg) → s.Init() → s.Run()
│   ├── migrate/main.go     # goose up/down/create CLI
│   └── seed/main.go        # 冪等 seeder 入口(roles/預設 policies/developer 帳號)
├── config/                 # 逐檔 struct + constructor(envconfig),config.go 僅聚合
│   ├── config.go  api.go  auth.go  cache.go  database.go  storage.go  observability.go
├── internal/
│   ├── server/server.go    # Server struct 註冊全部依賴 + Init()
│   ├── server/domains.go   # InitDomains():逐 domain 組裝 repo→usecase→handler
│   └── (domain/ middleware/ authz/ session/ database/ audit/ 既有路徑不變)
├── third_party/
│   ├── database/ent.go     # Ent client + pgx pool 初始化
│   └── cache/valkey.go     # Valkey client 初始化
├── proto/  ent/  database/migrations/  .air.toml  Taskfile.yml
\```

### DI 與啟動

- `Server` struct 持有全部共享依賴(Config、*ent.Client、Valkey、Casbin enforcer、casl FieldRegistry、scs manager、chi router)。
- 所有 fail-fast 啟動檢查集中於 `Server.Init()`:DB/Valkey 連線、Casbin enforcer 與預設 policy、`JWT_SECRET` 存在、production+developer 開關拒啟(1.11.1)、CASL fixture 自驗(D30-3)。
- `InitDomains()` 每 domain 一行組裝鏈;Connect handler 掛 chi router;新增 domain 只動此一檔。

### config 規則

- 套件改用 `github.com/kelseyhightower/envconfig`;struct tag 一律 `envconfig:"KEY"`;聚合入口 `config.New()`。
- key 歸檔依映射表(auth/cache/api/database/storage/observability);新群組無對應檔 → 新建 `config/<name>.go` 並聚合。
- `.env.example` 同步維護;`.env` 不入版控。

### third_party 規則

- 外部套件僅做「初始化/連線建立」者放 `third_party/`;行為邏輯(RLS hook、session 管理)留在 `internal/`。`third_party/database/ent.go` 建立 client 時掛 `internal/database/rls.go` 的 driver hook。

### 工具鏈

- Taskfile:`dev`(air)、`check`(fmt+vet+lint+test)、`vuln`(govulncheck)、`migrate`/`migrate:create`/`seed`;CI Go job 含 govulncheck。

### 術語映射(既有計畫文件適用)

| 既有計畫表述 | 實際落點 |
|---|---|
| 「main.go 組裝點 / 掛載 / 組裝處」 | `InitDomains()`(`internal/server/domains.go`) |
| 「main.go 啟動檢查 / 啟動防護 / 啟動序列」 | `Server.Init()`(`internal/server/server.go`) |
| 「config/config.go 加欄位 X」 | 依 key 歸檔映射表的分檔 |
```

（實際插入時移除 `\``` 的跳脫；章節編號依 00-index 現行結構接續。）

- [ ] **Step 2: 驗證**

Run: `grep -n "3.7 架構慣例" docs/superpowers/plans/backend-detail/00-index.md`
Expected: 1 hit

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/plans/backend-detail/00-index.md
git commit -m "docs: 00-index 新增架構慣例章節(D31)"
```

---

### Task 3: backend-detail/01-auth.md 錨點修訂

**Files:**
- Modify: `docs/superpowers/plans/backend-detail/01-auth.md`

- [ ] **Step 1: 五處修訂**

1. 子功能 1.4.1 檔案列 `Update backend/config/config.go(OAuth client id/secret/redirect/hd 限制)` → `Update backend/config/auth.go(OAuth client id/secret/redirect/hd 限制)`。
2. 子功能 1.6.1 檔案列 `Update backend/config/config.go(Valkey 連線、session TTL)` → `Update backend/config/cache.go(Valkey 連線、session TTL)`。
3. 子功能 1.6.6 檔案列 `Update backend/config/config.go(token 清單與對應身分/範圍)` → `Update backend/config/auth.go(token 清單與對應身分/範圍)`。
4. 子功能 1.11.1 檔案列：
   - `Update backend/config/config.go` → `Update backend/config/api.go`
   - `Update backend/cmd/server/main.go(啟動檢查)` → `Update backend/internal/server/server.go(Server.Init() 啟動檢查)`
5. 1.11.1 實作邏輯第 1 點「啟動時檢查:`ENV = production`…」維持行為描述不變（落點已由檔案列改明）。

- [ ] **Step 2: 驗證**

Run: `grep -n "config/config.go\|cmd/server/main.go" docs/superpowers/plans/backend-detail/01-auth.md`
Expected: 0 hits

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/plans/backend-detail/01-auth.md
git commit -m "docs: 01-auth 細部 config 分檔與啟動檢查落點(D31)"
```

---

### Task 4: 原計畫（2026-07-17）Phase 0 骨架與 Task 1.11

**Files:**
- Modify: `docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md`

- [ ] **Step 1: Phase 0 backend 骨架 Files 列修訂**

Task 0.x（Go 骨架）Files 列由：

```markdown
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/config/config.go`
- Create: `backend/internal/server/server.go`
- Create: `backend/Taskfile.yml`
```

改為：

```markdown
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`(極薄入口)
- Create: `backend/cmd/migrate/main.go`、`backend/cmd/seed/main.go`
- Create: `backend/config/{config,api,auth,cache,database,storage,observability}.go`(envconfig 逐檔 struct)
- Create: `backend/internal/server/{server,domains}.go`
- Create: `backend/third_party/{database/ent.go,cache/valkey.go}`
- Create: `backend/.air.toml`、`backend/Taskfile.yml`(含 dev/check/vuln/migrate/seed)
```

- [ ] **Step 2: Task 1.11 config 路徑**

`Modify: backend/config/config.go` → `Modify: backend/config/api.go`；Step 1 文字「config 新增 `DEVELOPER_ACCOUNT_ENABLED`…」中 config 改指 `config/api.go`。

- [ ] **Step 3: 驗證**

Run: `grep -n "config/config.go" docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md`
Expected: 0 hits

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md
git commit -m "docs: 原計畫 Phase 0 骨架與 1.11 對齊 D31 結構"
```

---

### Task 5: subproject-implementation-plan 目錄樹 + CI

**Files:**
- Modify: `docs/superpowers/plans/2026-08-05-sales-order-1.0-subproject-implementation-plan.md`

- [ ] **Step 1: 目錄樹修訂（backend 區塊）**

樹中：

```markdown
│   ├── Taskfile.yml
│   ├── cmd/server/main.go
│   ├── config/config.go
│   ├── internal/
│   │   ├── server/server.go
```

改為：

```markdown
│   ├── Taskfile.yml                    # 含 dev(air)/check/vuln/migrate/seed
│   ├── .air.toml
│   ├── cmd/{server,migrate,seed}/main.go
│   ├── config/                         # 逐檔 envconfig struct;config.go 聚合
│   │   ├── config.go  api.go  auth.go  cache.go  database.go  storage.go  observability.go
│   ├── internal/
│   │   ├── server/{server,domains}.go  # 集中 DI(D31)
│   ├── third_party/{database/ent.go,cache/valkey.go}
```

- [ ] **Step 2: Phase 0 backend 骨架 Task 的 Files 列同步**

同 Task 4 Step 1 的新 Files 列內容（七行 Create）取代該計畫對應四行。

- [ ] **Step 3: CI Go job 補 govulncheck**

CI Task Step 2「Lint (`golangci-lint`), test (`go test ./...`), build (`go build ./cmd/server`)」→「Lint (`golangci-lint`), vuln (`govulncheck ./...`), test (`go test ./...`), build (`go build ./cmd/server`)」。

- [ ] **Step 4: 驗證**

Run: `grep -n "config/config.go" docs/superpowers/plans/2026-08-05-sales-order-1.0-subproject-implementation-plan.md`
Expected: 0 hits；且 `grep -c "govulncheck" <同檔>` ≥ 1

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/plans/2026-08-05-sales-order-1.0-subproject-implementation-plan.md
git commit -m "docs: 子專案計畫目錄樹/Phase 0/CI 對齊 D31"
```

---

### Task 6: 01-auth TDD 計畫批次修訂

**Files:**
- Modify: `docs/superpowers/plans/2026-08-17-backend-01-auth-plan.md`

- [ ] **Step 1: 套用 R1（config 分檔）**

逐處修訂（錨點行號為撰寫時參考，以內容比對為準）：

| 位置 | 現行 | 改為 |
|---|---|---|
| Task OAuth（~L922/931/1286） | `backend/config/config.go`（OAuth 欄位） | `backend/config/auth.go` |
| Task session（~L2415/2630） | `config/config.go` 加 `JWTSecret`/`AccessTokenTTL`、Valkey | JWT 欄位 → `config/auth.go`；Valkey/session TTL → `config/cache.go` |
| Task middleware（~L3019/3028/3317） | `config/config.go`（ApiTokens） | `backend/config/auth.go` |
| Task developer（~L3787/4032） | `config/config.go` | `backend/config/api.go` |
| Global Constraints / File Structure 表 | `config/config.go` | `config/`（逐檔） |

- [ ] **Step 2: 套用 R2（tag）**

全部 struct tag `mapstructure:"…"` → `envconfig:"…"`（含 `JWTSecret`/`ApiTokens` 片段）；片段中 config 套件 import 註明 `github.com/kelseyhightower/envconfig`。

- [ ] **Step 3: 套用 R3（Load→New）**

`config.Load()` → `config.New()`（含 Task 14 main.go 啟動序列片段「在 `config.Load()` 之後」→「在 `config.New()` 之後」）。

- [ ] **Step 4: 套用 R4（組裝/啟動術語）**

- 「`main.go` 組裝時注入」「`main.go` 組裝點」→「`InitDomains()` 組裝點」。
- Task 14 啟動序列片段標題「`backend/cmd/server/main.go` 啟動序列插入（在 `config.New()` 之後）」→「`backend/internal/server/server.go` `Init()` 插入（依賴初始化之後）」；片段內容（developer 防護邏輯）不變。
- Task 14 Files 列 `Update: backend/cmd/server/main.go` → `Update: backend/internal/server/server.go`。
- L4016 備註「main.go 於 `cfg.Env == "development"` 時呼叫」→「`cmd/seed/main.go`（或 development 啟動時 `Server.Init()`）呼叫」。
- 手動驗證指令 `go run ./cmd/server` 不變（行為驗證仍經 server 入口）。

- [ ] **Step 5: 驗證**

Run: `grep -n "config/config.go\|mapstructure\|config\.Load()\|main\.go 組裝\|main\.go 啟動" docs/superpowers/plans/2026-08-17-backend-01-auth-plan.md`
Expected: 0 hits

- [ ] **Step 6: Commit**

```bash
git add docs/superpowers/plans/2026-08-17-backend-01-auth-plan.md
git commit -m "docs: 01-auth TDD 計畫套用 D31 轉換規則(R1-R4)"
```

---

### Task 7: 02/03/04/07/08/09 TDD 計畫批次修訂

（05-sales-orders、06-returns 兩份經錨點調查確認無 config/cmd 路徑提及，不修訂；由 Task 9 全庫驗證覆蓋。）

**Files:**
- Modify: `docs/superpowers/plans/2026-08-17-backend-02-tenancy-users-plan.md`
- Modify: `docs/superpowers/plans/2026-08-17-backend-03-metadicts-audit-plan.md`
- Modify: `docs/superpowers/plans/2026-08-17-backend-04-master-data-plan.md`
- Modify: `docs/superpowers/plans/2026-08-17-backend-07-notifications-plan.md`
- Modify: `docs/superpowers/plans/2026-08-17-backend-08-dispatch-plan.md`
- Modify: `docs/superpowers/plans/2026-08-17-backend-09-printing-plan.md`

- [ ] **Step 1: 02-tenancy-users**

- File Structure 表「`backend/config/config.go` | `StorageRoot`」→「`backend/config/storage.go` | `StorageRoot`」。
- Task 2.4（~L2130/2361/2708）：`config/config.go(StorageRoot)` → `config/storage.go`；StorageRoot 片段 tag 若為 `mapstructure` → `envconfig`。
- Task 2.10（~L3950/4326/4877/4900）：「`main.go` 呼叫點/組裝增量/seeder 呼叫點」→「`InitDomains()` 組裝點」；Files 列 `backend/cmd/server/main.go` → `backend/internal/server/domains.go`。
- 計畫尾「類型一致」段提及「main.go 同步」處 → 「`InitDomains()` 同步」。

- [ ] **Step 2: 03-metadicts-audit**

Task 2（~L1177/1360/1370）：Files 列與片段「`main.go` 組裝處」→「`InitDomains()` 組裝點（`internal/server/domains.go`）」；Files 列路徑同步。

- [ ] **Step 3: 04-master-data**

~L1803 commit 片段 `git add … backend/config/config.go` → 對應分檔（QR 簽章密鑰 → `config/auth.go`）。

- [ ] **Step 4: 07-notifications**

~L3390「main.go 組裝點」→「`InitDomains()` 組裝點」。

- [ ] **Step 5: 08-dispatch**

- File Structure 表（~L47）「`backend/config/config.go` | heartbeat…」→「`backend/config/api.go`」。
- Task 5（~L2048/2625/2647/2663）：`config/config.go` → `config/api.go`；`BoardHeartbeatSeconds` 片段 tag `mapstructure` → `envconfig`；「`main.go` 組裝點更新」→「`InitDomains()` 組裝點更新」。
- ~L487/2027 註解「main.go 組裝時呼叫」→「`InitDomains()` 組裝時呼叫」。
- 計畫尾「類型一致」段「main.go 組裝點同步」→「`InitDomains()` 組裝點同步」。

- [ ] **Step 6: 09-printing**

- File Structure 表（~L56）與 Task 4（~L3316/3994/4122/4145）：`backend/cmd/server/main.go`（掛載 PrintService）→「`InitDomains()`（`internal/server/domains.go`）掛載」；「Gotenberg base URL / 儲存根目錄自 config」→「自 `config/api.go` / `config/storage.go`」。

- [ ] **Step 7: 驗證（六檔）**

Run: `grep -n "config/config.go\|mapstructure\|main\.go 組裝\|main\.go 掛載\|main\.go 呼叫" docs/superpowers/plans/2026-08-17-backend-0{2,3,4,7,8,9}-*.md`
Expected: 0 hits

- [ ] **Step 8: Commit**

```bash
git add docs/superpowers/plans/2026-08-17-backend-0{2,3,4,7,8,9}-*.md
git commit -m "docs: 02-09 TDD 計畫套用 D31 轉換規則(R1-R4)"
```

---

### Task 8: CASL 計畫 Task 8 config 路徑

**Files:**
- Modify: `docs/superpowers/plans/2026-08-24-casl-integration-plan.md`

- [ ] **Step 1: 三處修訂**

1. Task 8 Files 列「Modify: `backend/config/config.go`（`CASLEnforcementEnabled bool`…）」→「Modify: `backend/config/api.go`（`CASLEnforcementEnabled bool`，env `CASL_ENFORCEMENT_ENABLED`，三環境預設 true）」。
2. Task 8 Step 3 配套修改第 2 點「`config.go` 加欄位與預設」→「`config/api.go` 加欄位與預設（tag `envconfig:"CASL_ENFORCEMENT_ENABLED"`）」。
3. Task 8 Step 5 commit 片段 `git add … backend/config/config.go …` → `git add … backend/config/api.go …`。

- [ ] **Step 2: 驗證**

Run: `grep -n "config/config.go" docs/superpowers/plans/2026-08-24-casl-integration-plan.md`
Expected: 0 hits

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/plans/2026-08-24-casl-integration-plan.md
git commit -m "docs: CASL 計畫 config 路徑對齊 D31(config/api.go)"
```

---

### Task 9: 全庫殘留驗證

**Files:** 無修改；純驗證。

- [ ] **Step 1: 舊路徑殘留歸零**

Run:

```bash
grep -rn "backend/config/config\.go" docs/superpowers/plans/ ; \
grep -rn "mapstructure" docs/superpowers/plans/ ; \
grep -rn "config\.Load()" docs/superpowers/plans/
```

Expected: 三條皆 0 hits（00-index 術語映射表若提及舊術語作為「既有表述」欄，屬白名單，不在排除範圍）。

- [ ] **Step 2: 組裝術語殘留檢查**

Run: `grep -rn "main\.go 組裝\|main\.go 掛載\|main\.go 啟動\|main\.go 呼叫" docs/superpowers/plans/`
Expected: 僅 00-index 術語映射表（白名單）；其餘 0 hits。

- [ ] **Step 3: 交叉一致檢查**

Run: `grep -rln "InitDomains\|envconfig\|third_party" docs/superpowers/plans/`
Expected: 命中清單至少含 00-index.md、原計畫、subproject-implementation-plan、01/02/03/04/08/09 TDD 計畫、CASL 計畫；逐檔確認新術語與 00-index §3.7 用字一致。

- [ ] **Step 4: 無 commit（無變更）**；若任一步驟有殘留 → 回到對應 Task 補修後重跑本 Task。

---

*計畫版本：v1.0.0（2026-08-24）；對應設計 `2026-08-24-backend-go8-structure-design.md`、決策 D31/D31-2。*
