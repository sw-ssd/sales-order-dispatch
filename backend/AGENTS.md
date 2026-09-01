# Backend Development Guidelines

> Go 後端:ent + Connect-Go + Casbin/CASL + goose + Valkey。所有文件、註解、commit message 一律繁體中文。

## 1. 分層與目錄

- `proto/`:proto 定義唯一來源;改動後必跑 `task proto:gen`(buf 產生 Go/TS/Dart;TS 只產 `*_pb`,前端自 GenService 匯入,不產 `*_connect`)。
- `internal/services/`:Connect-RPC handler;每 service 提供 `Register*Services(mux *http.ServeMux, db *ent.Client)`,由 `internal/server` 的 `mountAuth` 掛進單一 `apiMux`(**禁止** chi `Mount` 掛重複 `/api/v1` 路徑,會 panic;用 `apiMux.Handle` + `http.StripPrefix`)。
- `internal/domain/<name>/`:usecase + repository 介面;不反向依賴 services。
- `internal/auth/`:JWT/refresh(旋轉採原子消耗,Lua/鎖,禁止先讀後刪)、session、token_version(DB 欄位為準)、OIDC。
- `internal/authz/`:`authz` facade(`AccessibleFilter`/`Can`,開關 `CASL_ENFORCEMENT_ENABLED`)+ `authz/casl` 引擎(condition AST、evaluator、translate、FieldRegistry;與 @casl/ability golden 對賭,新增運算子必補 golden fixture)。
- `ent/schema/`:ent schema;改動後 `go generate ./ent` 並新增 goose migration(`database/migrations/NNNNN_name.sql`,必含 Up/Down,加欄位用 `IF NOT EXISTS` 對齊既有先例)。
- `internal/handlers/`:非 Connect 的純 HTTP handler(如 auth 回調)。

## 2. 授權與安全(不可妥協)

- 任何新 Connect 方法**必須**有授權門檻(Casbin `requireScope`/`requireRole` 模式),未登入 → `Unauthenticated`、越權 → `PermissionDenied`;前端守衛不算授權。
- 跨公司/跨租戶資料存取失敗回 `PermissionDenied`(非 `invalid_argument`);輸入驗證失敗才回 `InvalidArgument`。
- `role_permissions` 異動前必跑條件驗證 + 防鎖死(含 `all`/`*` subject);company_admin 的 id 欄位值須為自身公司或佔位符(以 `casl.ParseConditions` 展開驗證)。
- 設定密鑰(JWT_SECRET 等)production 下空值/預設值 → `Init()` fail-fast 拒絕啟動;驗證端對空密鑰 fail-closed。

## 3. 測試

- **只用 stdlib `testing` + `enttest`(sqlite)**;禁止引入 testify/testcontainers 等新測試依賴(go.mod 所有權共識)。
- enttest 每個 subtest 建獨立 client + cleanup;sqlite DSN 避免共享 cache(日後加 `t.Parallel` 才安全)。
- Valkey 依賴的測試需有 skip 保護(本機無 Valkey 時自動略過)。
- 授權門檻測試必含矩陣:未登入 / guest / staff / dept_admin / company_admin / super。

## 4. Modern Go Guidelines(必備)

撰寫、修改、修復或重構任何 Go 程式碼前,**必須**以 Modern Go Guidelines CLI 為準則來源(go-modern-guidelines,規則可能新於模型知識 cutoff):

```bash
sh ~/.omp/plugins/node_modules/go-modern-guidelines/plugin/skills/use-modern-go/scripts/run-tool.sh list            # 依 go.mod 版本列出適用規則
sh ~/.omp/plugins/node_modules/go-modern-guidelines/plugin/skills/use-modern-go/scripts/run-tool.sh explain <id>   # 讀取單條規則全文
```

- 首次執行會自動 `go install` CLI 至快取目錄;版本自 go.mod / go.work / 本地 toolchain 解析。
- `list` 輸出必須完整讀取,**禁止** pipe 至 head/tail/grep 截斷(新規則排在前面,截斷會漏掉重要準則)。
- 回傳的 guideline 視為本專案現代 Go 風格權威;與既有程式碼衝突時,新碼從 guideline、舊碼不主動回刷。

## 5. 驗證與工具

- 提交前必跑 `task check`(fmt + vet + lint + test);`task vuln` 掃弱點。
- migration 操作:`task migrate:up` / `migrate:status` / `migrate:down`;seed:`task seed`。
- 開發:`task dev`(air hot reload);infra(postgres/valkey/gotenberg)由根目錄 `task infra:start` 起。
