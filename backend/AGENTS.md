# Backend Development Guidelines

> Go 後端:ent + Connect-Go + Casbin/CASL + goose + Valkey。所有文件、註解、commit message 一律繁體中文。

## 1. 架構 Template:go8

本後端架構以 **go8**(https://github.com/sowiner/go8)為 template,對齊決策見 D31(設計 `docs/superpowers/specs/2026-08-24-backend-go8-structure-design.md`;對齊計畫 `docs/superpowers/plans/backend/2026-08-24-backend-go8-alignment-plan.md`,已完成 38/38)。五項不可違反之慣例:

1. **集中 DI**:組裝點唯一,在 `internal/server/domains.go` 的 `InitDomains()`;啟動檢查/防護在 `Server.Init()`(`internal/server/server.go`)。**禁止**在 `main.go` 或各套件自行組裝。
2. **cmd 拆分**:入口於 `cmd/` 下拆分,`main.go` 僅啟動,無業務邏輯。
3. **config 逐檔 envconfig struct**:`config/` 依 key 群組分檔(`api.go`、`storage.go`、`observability.go`…),`config.go` 僅餘聚合 struct 與 `New()`;tag 一律 `envconfig:"KEY"`(github.com/kelseyhightower/envconfig);新 key 群組無對應檔時比照 go8 新建 `config/<name>.go` 並聚合。
4. **third_party/**:外部整合程式碼集中於 `third_party/`。
5. **工具鏈**:air hot reload(`task dev`)+ `task check`(fmt+vet+lint+test)+ `task vuln`(govulncheck)。

新功能選址時先對照 go8 分層;與 template 衝突的做法一律不採用。

## 2. 分層與目錄

- `proto/`:proto 定義唯一來源;改動後必跑 `task proto:gen`(buf 產生 Go/TS/Dart;TS 只產 `*_pb`,前端自 GenService 匯入,不產 `*_connect`)。`protoc-gen-dart`(於 `~/.pub-cache/bin`)為本地 plugin,**須以 fvm pin 的 Dart SDK 執行**(專案 `.fvmrc` stable = 3.13.3):`export PATH="$HOME/fvm/versions/3.47.4/bin:$HOME/.pub-cache/bin:$PATH"` 後再跑;若 `dart` 解析到錯誤版本會報 `wire unmarshal: cannot parse invalid wire-format data`。
- `internal/services/`:Connect-RPC handler;每 service 提供 `Register*Services(mux *http.ServeMux, db *ent.Client)`,由 `internal/server` 的 `mountAuth` 掛進單一 `apiMux`(**禁止** chi `Mount` 掛重複 `/api/v1` 路徑,會 panic;用 `apiMux.Handle` + `http.StripPrefix`)。
- `internal/domain/<name>/`:usecase + repository 介面;不反向依賴 services。
- `internal/auth/`:JWT/refresh(旋轉採原子消耗,Lua/鎖,禁止先讀後刪)、session、token_version(DB 欄位為準)、OIDC。
- `internal/authz/`:`authz` facade(`AccessibleFilter`/`Can`,開關 `CASL_ENFORCEMENT_ENABLED`)+ `authz/casl` 引擎(condition AST、evaluator、translate、FieldRegistry;與 @casl/ability golden 對賭,新增運算子必補 golden fixture)。
- `ent/schema/`:ent schema;改動後 `go generate ./ent` 並新增 goose migration(`database/migrations/NNNNN_name.sql`,必含 Up/Down,加欄位用 `IF NOT EXISTS` 對齊既有先例)。
- `internal/handlers/`:非 Connect 的純 HTTP handler(如 auth 回調)。

## 3. 授權與安全(不可妥協)

- 任何新 Connect 方法**必須**有授權門檻(Casbin `requireScope`/`requireRole` 模式),未登入 → `Unauthenticated`、越權 → `PermissionDenied`;前端守衛不算授權。
- 跨租戶資料存取失敗的錯誤碼分兩層:**服務層 ACL 判定的越權** → `PermissionDenied`(非 `invalid_argument`;輸入驗證失敗才回 `InvalidArgument`);**被 RLS 過濾掉的目標**(查詢根本看不到該列)→ `NotFound`,不得回 `PermissionDenied` 洩漏「該資源存在」(見 §9-8)。
- `role_permissions` 異動前必跑條件驗證 + 防鎖死(含 `all`/`*` subject);company_admin 的 id 欄位值須為自身公司或佔位符(以 `casl.ParseConditions` 展開驗證)。
- 設定密鑰(JWT_SECRET 等)production 下空值/預設值 → `Init()` fail-fast 拒絕啟動;驗證端對空密鑰 fail-closed。

## 4. 測試

- **單元測試**:只用 stdlib `testing` + `enttest`(sqlite);不引入 testify(維持 stdlib 斷言風格)。
- **整合測試(需要真 PostgreSQL / OpenFGA / Valkey 才能觀察的行為)**:**使用 testcontainers-go 起拋棄式容器**。
  - 使用者指示 2026-09-19:**啟用 testcontainers**;先前「禁止 testcontainers 等新測試依賴」的共識作廢。新測試依賴以 `github.com/testcontainers/testcontainers-go` 為限。
  - 一律以 build tag `integration` 隔離(`//go:build integration`):**預設 `go test ./...` 不得起容器、必須可離線跑綠**(CI 的預設 job 即為此)。
  - 容器啟動樣板集中在**單一** helper(位置見下),個別測試**不得**自寫 `podman`/`docker` 指令或自行拼 DSN。
  - 若環境已提供現成 DB,可用 `INTEGRATION_TEST_DSN` 覆寫、跳過容器(既有 `internal/authz/openfga/integration_test.go` 的 gating 沿用此名)。
  - 本機無 docker CLI(只有 podman machine):執行方式由 Taskfile 固化為 `task test:integration`(內部設 `DOCKER_HOST=unix://<podman machine socket>`;podman 下 ryuk 無法運作(掛載 socket volume 不支援),故關閉 `TESTCONTAINERS_RYUK_DISABLED=true`,容器清理改由 helper 的 `t.Cleanup` 負責),**不要**把這些 env 寫死在測試碼裡。
- enttest 每個 subtest 建獨立 client + cleanup;sqlite DSN 避免共享 cache(日後加 `t.Parallel` 才安全)。
- Valkey 依賴的測試需有 skip 保護(本機無 Valkey 時自動略過);整合測試同理,但界線較嚴:**容器執行環境連不上 → `t.Skip`(附可行動訊息);runtime 可用而設定/起容器流程壞掉 → 必須 fail**。理由是後者若也 skip,整包會靜默全綠卻一個測試都沒跑(且會遺留殘留容器)。無論哪種情況,`go test ./...`(預設路徑)都不得因環境而紅。
- 授權門檻測試必含矩陣:未登入 / guest / staff / dept_admin / company_admin / super。
- **分頁清單一律要有唯一次序鍵**:任何以 `ORDER BY <非唯一鍵>` 搭配 `LIMIT/OFFSET` 的清單,**必須**追加 `ent.Asc(<entity>.FieldID)` 作為 tie-break,否則同值群跨頁邊界時 PG 會讓某些列重複出現、某些列完全不出現(本專案已於公司/部門/角色/客戶/稽核與六個主檔清單實證)。
  - **這類缺陷只在真 PostgreSQL 看得到**:sqlite(enttest)對同值群給穩定次序、`LIMIT/OFFSET` 只是切片,同構探針會全綠 → 守門測試必須是 `//go:build integration` + testcontainers,並以「**逐頁掃描 == 單次全量**(筆數、id 集合、無重複)」為斷言(樣板:`internal/services/list_pagination_integration_test.go`)。
  - **固定排序的端點也要釘住方向**(`sort_order` 升冪、稽核最新在前…):只比對集合會讓 `Asc`/`Desc` 互換而測試全綠 → 需逐位比對序列或斷言預期極值。

## 5. Modern Go Guidelines(必備)

撰寫、修改、修復或重構任何 Go 程式碼前,**必須**以 Modern Go Guidelines CLI 為準則來源(go-modern-guidelines,規則可能新於模型知識 cutoff):

```bash
sh ~/.omp/plugins/node_modules/go-modern-guidelines/plugin/skills/use-modern-go/scripts/run-tool.sh list            # 依 go.mod 版本列出適用規則
sh ~/.omp/plugins/node_modules/go-modern-guidelines/plugin/skills/use-modern-go/scripts/run-tool.sh explain <id>   # 讀取單條規則全文
```

- 首次執行會自動 `go install` CLI 至快取目錄;版本自 go.mod / go.work / 本地 toolchain 解析。
- 若上述 plugin 路徑在**本機不存在**(實測:2026-09-19 本機 `~/.omp/plugins/node_modules/go-modern-guidelines` 不存在、`omp-plugins.lock.json` 無此 plugin),改用同支腳本的上游來源執行,並在該次交付說明你用的是哪個來源:
  `git clone --depth 1 https://github.com/JetBrains/go-modern-guidelines /tmp/go-modern-guidelines && sh /tmp/go-modern-guidelines/plugin/skills/use-modern-go/scripts/run-tool.sh list`
- `list` 輸出必須完整讀取,**禁止** pipe 至 head/tail/grep 截斷(新規則排在前面,截斷會漏掉重要準則)。
- 回傳的 guideline 視為本專案現代 Go 風格權威;與既有程式碼衝突時,新碼從 guideline、舊碼不主動回刷。

## 6. 驗證與工具

- 提交前必跑 `task check`(fmt + vet + lint + test);`task vuln` 掃弱點。
- migration 操作:`task migrate:up` / `migrate:status` / `migrate:down`;seed:`task seed`;本機首次啟動或刪過 DB volume 後,先 `export DATABASE_ADMIN_URL=<owner DSN>` 再跑 `task db:app-password`(`00022` 只建 `app_rw` 角色與授權、不設密碼;業務連線預設就是這個角色)。
- 開發:`task dev`(air hot reload);infra(postgres/valkey/gotenberg)由根目錄 `task infra:start` 起。

## 7. 程式碼索引與查詢（codebase-memory-mcp）

本專案**程式碼索引與查詢皆需先經過 codebase-memory-mcp**（見根目錄 `docs/AGENTS.md` §4.0）：找定義／實作／呼叫端、追蹤呼叫路徑、影響範圍分析、跨專案跳轉等一律先以 codebase-memory 知識圖譜查詢，不足處再以 grep／直接讀檔補足；改動既有程式碼前先確認目標檔的索引覆蓋狀態。

## 8. 資料庫 schema 真相、刪除語意與軟刪除（實戰教訓）

1. **`database/migrations/**` 是 schema 的唯一真相**；`ent/schema/**` 用來生成查詢碼,**不**用來改既有庫。
   - **禁止**對 goose 管理的庫執行 ent 的 `Schema.Create`／`auto-migrate`:它會把遷移建立的外鍵與索引當作多餘而 `DROP`(本專案 33 個 FK 中 ent 只宣告 4 個),且實測在既有庫上會直接以 atlas 逆向工程錯誤失敗。新庫由 `cmd/migrate up` 建;`ent` 僅在測試(enttest sqlite)與查詢碼生成中使用。
   - 改 ent schema 後必 `go generate ./ent`;**同時**新增 Goose 遷移(兩者對齊,型別/主鍵/唯一性/NULL 性都要一致)。已實證的落差類型:`id` 主鍵缺漏(`customer_counters` 曾是如此,真 PG 上 `CreateCustomer` 必然 `42703`)、`integer` vs `bigint`、DB `NOT NULL` vs ent `Optional`。
2. **刪除語意一旦改變(硬刪→軟刪),必須重掃「所有碰得到該實體的路徑」**:硬刪除的 FK 曾是**隱性的不變式保護**,軟刪後列還在、保護消失。實證漏點:`CreateUser` 沒有公司存在性檢查、`Login` 未檢查公司是否軟刪除(已刪公司的使用者仍能登入取得 token)、`UpdateDepartment` 會改到已刪部門、`DeleteCompany` 的「仍有部門」前置檢查會被軟刪部門永久擋住。
   - 軟刪除的識別碼唯一性要用 **partial unique index**(`WHERE deleted_at IS NULL`),並移除舊的表層 UNIQUE。
   - 「刪除前的前置檢查」與「掛載資料到該列」之間有競態:需**兩側對同一列取互斥鎖**(掛載端 `FOR SHARE`、刪除端先 `FOR UPDATE` 再條件式 `UPDATE`)。**單側鎖不足**:READ COMMITTED 只重評目標列,`NOT EXISTS` 子查詢仍用敘述開始的快照。方言判斷用 `sql.Selector.Dialect()`(sqlite 不支援 `FOR ...`,不可寫入鎖子句)。
3. **RLS 已全站生效**(2026-09-20:`00024`–`00028` 對 18 張業務表 `ENABLE` + `FORCE`;policy 由 `00007`/`00011`/`00023` 定義、`00025` 正規化為 `NULLIF` 形式)。RLS 是跨公司隔離的**最後一道防線**,授權仍以服務層門檻為準(§9);未帶 scope 的查詢一律 fail-closed(0 列)。
4. **`toConnectError` 之類的全域錯誤映射**:不要把 DB 原始訊息(含 `SQLSTATE`/constraint 名)回給客戶端;約束類錯誤回 `FailedPrecondition` 並落 server log,`AlreadyExists` 僅用於真正的「已存在」語意(需在建立路徑自行前置判別)。

## 9. RLS 與租戶交易（D36；2026-09-20 起全站生效）

全站 18 張業務表已 `ENABLE` + `FORCE`（`00024`–`00028`）；policy 由 `00007`/`00011`/`00023` 定義、`00025` 正規化。請求層租戶交易（每個 unary RPC 一個交易、`SET LOCAL app.*` 由 driver 裝飾器在 `Tx(ctx)` 內套用）見 `internal/dbtenant`。以下每一條都是本計畫用實測換來的，違反其中任一條都會以「黑屏」或「靜默」的形式出錯。

1. **PG 的 superuser 恆繞過 RLS（`FORCE` 亦然）** → 宣稱在驗 RLS 的測試**必須**以 `app_rw`（`00022` 的 `NOBYPASSRLS` 非 owner 角色）＋ `dbtenant.NewClient` 連線；以容器 superuser 連線的既有整合測試只能當 regression gate。
   為什麼：用 superuser 連線時，漏掛租戶交易的查詢照樣讀得到全部列 —— 測試全綠卻什麼都沒驗到（T5 實測）。
2. **`USING` 不得嚴於 `WITH CHECK`**：讀取面的收緊一律放服務層 ACL，DB 層只負責跨公司隔離；`USING` 可以比 `WITH CHECK` 寬（`core_metadicts_scope` 即如此：`USING` 允許讀系統預設列、`WITH CHECK` 不允許寫）。
   為什麼：ent 的 Create 一律 `INSERT … RETURNING id`（`sqlgraph` 的 `insert.Returning(c.ID.Column)`），而 PG 對 `RETURNING` 套的是 SELECT policy → `USING` 較嚴會讓「其實合法」的新列寫不進去。
3. **policy 取值一律 `NULLIF(current_setting('app.current_*', true), '')`**。
   為什麼：`SET LOCAL` 對自訂 GUC 會在 session 留下空字串 placeholder，交易結束後回到 `''` 而非 unset；少了 `NULLIF`，同一條池化連線之後「未帶 scope」的查詢會以 `22P02`（`''::bigint`）失敗，而不是乾淨地回 0 列（T5 實測）。
4. **每個 scope 等級都要在 `WITH CHECK` 有分支，或改用不依 scope 等級的租戶鍵**。
   為什麼：缺分支 = 該等級的寫入一律被擋。`core_audit_logs_scope` 在 `00025` 的 `WITH CHECK` 只有 `all`/`company` → department 身分連「建一筆客戶」都會整筆失敗（`00027` 才改成以 `company_id` 為鍵）。
5. **未登入／系統路徑一律走 `dbtenant.SystemScopeTx`**：現行呼叫點＝`server.identityFor`、`handlers` 的 `systemScope`（登入／註冊／OIDC／guest／refresh）、`auth.TokenManager`（自帶等價實作，見第 10 條）、`authz.Provision`、`user_service.roleIDByCode`、`cmd/seed`。
   為什麼：這些路徑沒有身分可注入 scope，不顯式宣告系統範圍就會 fail-closed，而症狀是「登入全滅」或「佈建靜默歸零」——兩者都不會報錯（T9 實測）。
6. **`SystemScopeTx` 需要 ≥2 條連線**：測試的連線池**不得**設 `SetMaxOpenConns(1)`（要釘死結請用池 2 ＋ deadline）；現況 `database.OpenSQL` 不設上限（Go 預設無限）故安全，但任何部署若要設上限必須 ≥2 且留餘裕。
   為什麼：它與請求交易並存，池只有 1 條時兩者互等 → 死結（會以逾時收斂，看起來像「服務很慢」）。
7. **同一請求內不得對同一列開第二條交易**：需要系統範圍時，併進單一的 `SystemScopeTx` 或改寫成單一敘述。
   為什麼：兩條交易會互等同一列的鎖 → 死結（T9 實測：`completeGuest` 的更新與 token_version+1 併成同一敘述）。
8. **跨租戶一律回 `NotFound`，且不得有副作用**（目標列不變、不落稽核）。
   為什麼：RLS 在讀取階段就濾掉目標，服務層沒機會看到它；回 `permission_denied` 等於洩漏「該資源存在」。與服務層 ACL 的越權（`permission_denied`）是兩種不同意義：**ACL 是「你沒有這個權」、RLS 是「這一筆不存在」**（見 §3）。守門探針：`internal/services/rls_cross_tenant_integration_test.go`。
9. **收斂掃描是路徑級且窮盡**：不限 `internal/services`，`internal/server`、`internal/handlers`、`internal/auth`、`internal/authz`、`internal/audit`、`internal/domain/**` 都要查；且要以「實際有 DB 存取」為準，不是「有沒有持有 client」（`domain/auth/ability.go` 持有 `*ent.Client` 卻完全沒用到；照持有與否改會是多餘改動）。
   為什麼：漏掉一條路徑就是該端點黑屏或靜默歸零，而下一個 RLS 波次不會回頭掃它（T9 的路徑級掃描即為此）。
10. **已知重複實作**：`internal/auth/token.go` 自帶 `systemScopeTx`（語意與 `dbtenant.SystemScopeTx` 等價），成因是 `dbtenant → auth` 的既有匯入方向會形成 import cycle。**改動 RLS 原語時兩份都要改**；後續應把 RLS 原語（`RLSScope`／`WithRLS`／`RLSStatements`／系統範圍交易）下沉到 `auth`，`dbtenant` 改為薄委派。
    為什麼：兩份實作漂移時只會壞掉其中一條路徑，而且症狀是靜默的。
11. **外部系統的副作用必須在 DB commit 之後**（`dbtenant.AfterCommit` 的 post-commit 掛鉤）：OpenFGA tuple 同步已如此，未來的 FCM 推播／派車串流一體適用。**適用範圍不限 OpenFGA**。
    為什麼：在交易內同步外部狀態，一旦該交易回滾就留下不一致，而授權類同步是 **fail-open 方向**（多授權）；掛鉤在 rollback 時整個丟棄，並讓交易的列鎖不跨越外部 I/O。
12. **`SystemScopeTx` 必須搭配 `dbtenant.NewClient` 建立的 client**：`SET LOCAL` 是 driver 裝飾器在 `Tx(ctx)` 內套的，餵裸 client（`entsql.OpenDB`／`database.OpenEnt` 直接建）時**沒有任何 scope 被設定**，寫入會以 `42501` 被 `WITH CHECK` 擋下 ——「包了 `SystemScopeTx`」不等於「有系統範圍」。同理 `dbtenant.Client(ctx, fallback)` 的 fallback 也必須是裝飾過的 client。
    為什麼：`SystemScopeTx` 只做「注入 scope 到 ctx ＋ 開交易」，真正下 `SET LOCAL` 的是裝飾器；少了裝飾器，錯誤訊息（`new row violates row-level security policy`）會誤導人以為是 policy 問題（T10 實測，`cmd/seed` 的第一版修法就是這樣錯的）。

### 9.1 已知設計缺口：客戶 App 的「只讀自己」在 RLS 下沒有 `self` 分支（T10 結論）

**事實**（2026-09-20 實查）：

- `core_customers_scope`（`00023` 建立、`00025` 以 `NULLIF` 重建）的 `USING`／`WITH CHECK` 只有 `all` 與「`company_id` 相符且（`scope=company` 或 `department_id IS NULL` 或 `= current_department_id`）」——**沒有 `self` 分支**。
- `ScopeForRole("customer") = self`（`internal/auth/rls.go`；`cmd/seed` 的內建角色 `customer` 也是 `self`），而 `RLSStatements` 對 `self` 只設 `app.current_user_id` 與 data_scope，**不設** `department_id`（客戶帳號沒有部門）。
- `RLSScope.CustomerID` 在生產程式碼**沒有任何生產者**：只有 `RLSStatements` 會把它寫進 `app.current_customer_id`，而 `server.identityFor` 只從 `users` 列組出 scope（`CustomerID` 恆為零值）。
- 客戶帳號就是 `users` 的一列（`is_customer = true`、`account_name` = `customers.customer_code`，登入見 `auth_handler.Login`）；`customers` 與該帳號之間**沒有**「這筆帳號是哪一筆客戶」的欄位（唯一連到 `users` 的 FK 是 `customers.default_sales_rep_id → users.id`，方向相反）。

**結論**：

1. 今天 customer 身分對 `customers` 的可見集合是「**同一公司所有 `department_id IS NULL` 的客戶列**」——同租戶內的公司層全體，而不是「自己那一筆」；寫入面同理。這與 `self` 的語意不符。
2. **不是跨公司洩漏**（`company_id` 條件仍在），且目前**不可達**：服務層對 customer/guest 一律 `permission_denied`（`deptScope` 的 default 分支），客戶 App 頁面亦尚未實作（見功能對照表）。
3. 要讓「只讀自己」成立，必須三者齊備：①先有「帳號 → 客戶列」的連結（新增 `users` → `customers` 的 FK，或在 `identityFor` 以 `account_name = customer_code` 解析）；②`identityFor` 據此填入 `RLSScope.CustomerID`；③在 `core_customers_scope` 補 `self` 分支（`id = NULLIF(current_setting('app.current_customer_id', true), '')::bigint`）。三者缺一不可，且屬 **policy ＋ 服務層語意變更**。
4. **歸屬**：不在 RLS 啟用計畫（T10 明文「不要改 policy」）內。此缺口應由「**客戶 App 業務頁面**」的實作計畫一併處理：先在該計畫決定連結方式（新增 FK vs 以 `account_name` 解析），再落 policy ＋ `identityFor` ＋ ACL，並附一條以 `app_rw` 跑的 self-範圍探針（樣板：`internal/services/rls_cross_tenant_integration_test.go`）。在此之前的任何「客戶 App 讀自己的資料」實作都不可依賴 RLS。

