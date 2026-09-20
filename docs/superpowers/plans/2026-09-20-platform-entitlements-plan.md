# Platform 域與權益守衛 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立平台域（`platform` schema）與權益判定：方案／權益／訂閱／override 的資料模型與 store、`entitlement.Allows/CheckLimit` 判定（fail-closed、可快取）、四個業務服務的配額守衛、平台操作者身分（OIDC ＋ 白名單）與唯讀 API（平台工具跨租戶視圖 ＋ 租戶端投影）。

**Architecture:** 平台域是獨立 package（`internal/platform/**`）＋ 獨立 proto（`platform/v1`）＋ 獨立 PG schema（`platform`），**不與業務域 JOIN**，只用 `company_id` 對照。平台表**不使用 ent**——改以 `database/sql` ＋ 明確 SQL（`platform` schema 在 sqlite 不存在，ent 的 codegen／auto-migrate 只會製造摩擦；AGENTS.md §8「migration 是 schema 唯一真相」）。權益判定以 store 介面抽象，單元測試用假實作（免 Docker），SQL 實作另以整合測試覆蓋。業務服務經建構子注入 `*entitlements.Service`（**編譯器強制每個呼叫端表態**，不會靜默漏掛）。

**Tech Stack:** Go 1.25、connect-go、goose、PostgreSQL 16、`database/sql`（pgx driver）、Valkey、testcontainers-go

**Spec:** `docs/superpowers/specs/2026-09-20-saas-billing-entitlements-design.md`（§2.2、§3、§4、§8 為權威）

**依賴**：`docs/superpowers/plans/2026-09-20-rls-tenant-isolation-plan.md`（Plan A）——**migration 編號接在其後（本計畫自 `00029` 起）**；功能上不互相依賴（entitlement 走 admin 連線，與 RLS 無關），可與 Plan A 並行，但編號順序須協調。

## Global Constraints

- **平台域不與業務域 JOIN**：`platform.*` 只以 `company_id` 對照業務資料；跨域副作用一律經 `platform.events`（outbox）
- **`app_rw` 對 `platform` schema 零權限**：平台域只用 admin 連線（`cfg.Database.AdminDSN()`）；業務連線連 `SELECT` 都不行（Plan A 已建立角色）
- **平台表不用 ent**、不得對 `platform` schema 跑 `client.Schema.Create`
- **權益判定 fail-closed**：無訂閱／查無方案 → 全部 `false` / `0`；額度不足與訂閱不可用一律回 `FailedPrecondition`（非 `PermissionDenied`——那是「缺權限」的語意），且**用註冊碼**：`PLAT-5001`／`PLAT-5002`／`PLAT-3001`（見 Task 4 開頭的錯誤碼表、Plan D Task 5b）
- **守衛掛點必須可機械驗證**：RPC → feature 對應表（見 Task 6），表驅動測試漏一項即紅
- **平台稽核不寫租戶 `audit_logs`**：`audit.Record` 要求 `company_id`／`user_id` 非零，平台操作者兩者皆無 → 一律寫 `platform.audit_logs`（S9）
- **`platform.*` 能力不得出現在租戶 `GetAbility`／角色權限矩陣**（S11）
- **平台操作者不建立、也不使用租戶 `users` 列**（S8）；租戶 session／JWT 不得通過平台 RPC 的驗證，反之亦然（不同 secret ＋ audience）
- 註解與 commit message 一律繁體中文
- 每任務結束前跑 `task check`；含整合測試者另跑 `task test:integration -- -run <TestName> -v`

---

## Progress

> 執行記錄與逐任務細節（含每次審查的判定、裁決與更正）見 `.superpowers/sdd/2026-09-20-platform-entitlements-plan/progress.md`。狀態：**12/12 任務完成**（2026-09-20）。下表依**任務編號**排列（與計畫相符）。

| # | 任務 | 狀態 | 產出（commit） | 驗證 |
|---|---|---|---|---|
| 1 | `config.Platform` 與啟動防護 | ✅ | `1c6ac7c`＋`ef3135f`／`79ce7e6` | 7 子測（全以 `config.New()`＋env 驅動）、真容器啟動防護實跑 |
| 2 | `platform` schema 與 10 張表（`00029`） | ✅ | `e3a193b`＋`ceccc8c` | 10 表＋6 索引（含 3 partial UNIQUE 的 `pg_get_expr` 謂詞）、`app_rw` 零權限雙路實證、Down 可 up→down→up |
| 3 | 平台 store（介面＋假實作＋PG） | ✅ | `edf436b`＋`a88408c`／`96a17ec`／`2b6cf1b` | 5 契約測試（免 Docker）＋真 PG 種子（NULL 限額／cancelled／已到期 override） |
| 4 | 權益判定（`entitlements.Service`） | ✅ | `875a623`／`1139e00`／`b3f685d`／`5b479c4`＋`72e7386`／`fd7ccc3`／`c66c9a4`／`8c268e1` | 17 列判定表、三碼含 details、快取 TTL 兩測試（無斷言落在緊湊時間窗） |
| 5 | 計數器與服務注入（編譯器驅動） | ✅ | `933de22`＋`d2bd170`→`193926b`→`6b2be85`→`fb41808` | app_rw 整合探針（含「無請求交易→四 feature 皆 0」對照組）；部門 scope 改公司層計數（探針留成子測） |
| 6 | 配額守衛掛點與表驅動測試 | ✅ | `cde1808`（RED）→`3bed56c`＋`7122767` | 矩陣 6/6＋語意＋真 PG；`dept_admin` 超額被擋（附突變）；`UpdateUser` 復原守衛跑完整序列 |
| 7 | `platform/v1` proto 與三端生成 | ✅ | `9f6a683` | 10 檔重生成 byte-identical、三語言執行期 smoke 全通 |
| 8 | 平台操作者認證（OIDC＋operator JWT） | ✅ | `5deaf04`…`d425818`＋`c3733c0` | 16 條（cookie 旗標／audience／`exp` 缺／HS384／空密鑰 fail-closed／state 驗完即清） |
| 9 | `PlatformAdminService` 唯讀 RPC 與平台稽核 | ✅ | `e0088d9`（RED）→`345ed2e`＋`41e620d`＋`9548c55`＋`90155a5` | `/platform/` 掛載 path-match；交易內 `SET LOCAL` 解 FORCE RLS（附突變）；LATERAL 挑 cancelled；排除平台自營公司 |
| 10 | 租戶端權益投影（`TenantEntitlementService`） | ✅ | `9270cba`（RED）→`22b00ed` | 4 條新測試；掛在租戶 `apiMux`（`/api/v1`，非 `/platform/`）；無計數器的 integer feature 降級略過＋log；守門 233/98 未加寬 |
| 11 | Seeds（features／方案／價目／權益／首位 operator） | ✅ | `6ee4f37`（RED）→`67646d3`＋`9bd96e1`＋`bfc33dc`＋`511c889`＋`0b0bc95`＋`2855964` | 真容器連跑 2 次列數不動 `7/3/6/16/1/1/1`；值未變則不推進 `settings.updated_at` |
| 12 | 慣例文件、CI 與計畫索引 | ✅ | 本表所在 commit | `git diff --stat` 僅文件與 CI；引用數字全以指令重驗（見下表） |

**現況數字（以程式與產生檔為準，勿抄舊稿）**：

| 數字 | 值 | 確認指令 |
|---|---|---|
| `platform` 表數 | **10** | `grep -c "CREATE TABLE IF NOT EXISTS platform\." backend/database/migrations/00029_platform_schema.sql` |
| seed features | **7**（`limit.storage_gb` 刻意不在內） | `grep -c '^\s*{"limit\.\|^\s*{"feature\.' backend/cmd/seed/platform.go` |
| seed 方案權益列 | **16**（free 4／std 5／pro 7） | `grep -o '"limit\.[a-z_]*": *[-0-9]*\|"feature\.[a-z_]*": *[-0-9]*' backend/cmd/seed/platform.go \| wc -l` |
| 守衛清單 | **6**（spec 的「部門復原」是缺口） | `grep -c '^\t{"' backend/internal/services/entitlement_guard_test.go` |
| errcode 基線 | **98 行／233 呼叫點**（只減不增） | `wc -l < backend/internal/services/errcode_baseline.txt`；`awk -F: '{s+=$NF} END {print s}' …` |
| 錯誤碼 | **21**（SYS 7／AUTH 7／PLAT 4／CUST 3），**17 已落點** | `grep -oE "(AUTH\|SYS\|PLAT\|CUST)-[0-9]{4}" docs/error-codes.md \| sort -u \| wc -l` |
| 權益快取 TTL | **60s**（保底；失效由寫入方 `Delete`） | `grep -n entitlementCacheTTL backend/internal/server/domains.go` |
| migration 編號 | 至 **`00029`** | `ls backend/database/migrations` |

### 執行期間的計畫更正（已寫回內文或程式碼註解）

| # | 更正 | 理由 |
|---|---|---|
| 1 | 守衛清單 **7 → 6**（`DepartmentService.RestoreDepartment` 不存在） | `proto/salesorder/v1/company.proto` 的 `DepartmentService` 只有 List/Get/Create/Update/Delete；00020 的部門軟刪除只做了 Delete 側。第 7 項記為**具名缺口**（測試檔有註解），**不**在守衛任務裡新增 RPC |
| 2 | `domains.go` 的 `mountEntitlements` **fail-fast**（計畫原為「log＋略過」） | 略過會讓四個業務 `register` 落空＝四個業務 RPC **整組不掛載**，比無守衛更糟 |
| 3 | 建構子收 **consumer 端最小介面** `entitlementChecker` | 計畫讓建構子吃具體型別 `*entitlements.Service`，T6 的記錄式假物件無法注入 |
| 4 | 錯誤一律走 **`errcode` 註冊碼**（計畫多處寫裸 `connect.NewError`） | Plan D 的守門測試會紅；啟動防護則用 `fmt.Errorf`（不是對外契約） |
| 5 | `limit.storage_gb` **自 v1 seed 移除**（features 8→7、entitlements 19→16） | 無計數器 → 種下去會讓租戶端權益投影對所有租戶失敗 |
| 6 | 席位計數在 `department`／`self` scope 改走 **`SystemScopeTx` 公司層**讀取 | T5 審查實測：請求交易內計數會被 RLS 過濾 → 部門 scope 可「每部門一份上限」繞過配額 |
| 7 | **super／developer 略過配額**（spec §4.3 有、計畫 §4.5 漏）集中在 `guardQuota`；並補 `UpdateUser` 復原守衛 | 平台方不受單一租戶合約限制（S10／R8）；`inactive→active` 與 `Restore*` 是同型漏洞的另一入口 |
| 8 | 平台 RPC 掛在**字面 `/platform/` 之下** | operator cookie `Path=/platform` 與 Connect procedure `/platform.v1.…` 不合 RFC 6265 path-match → 瀏覽器不送 cookie；放寬 cookie 成 `/` 又會送往租戶 API |
| 9 | T11（seeds）以**計畫 Task 11 段落**為權威實作 | 派工用的 `task-10-brief.md` 內容其實是計畫 Task 10（租戶端投影），兩者編號不一致（已如實記錄在該任務報告） |
| 10 | T12 追加 **proto 產生檔的 CI 漂移閘門**（`.github/workflows/ci.yml` 的「Proto generated files up to date」） | 計畫只涵蓋 errcode 產生檔（Plan D）；proto 的漂移原本要等到「編譯不過」才會被發現，而三端生成檔一旦漂移，前端／App 的型別錯誤會在更晚才爆 |
| 11 | T12 的 `backend/AGENTS.md` §11 由草案 8 條長成 **13 條** | 執行期間新增的硬約束（守衛唯一入口、check-then-act、`/platform/` 前綴、`requestid`、跨租戶 `SET LOCAL`、seed 冪等定義）都是實測換來的，不寫進慣例很快就會被下一個實作者踩回去 |
| 12 | 計畫 Step 3 的「本地 `task check` ＋ `task test:integration`」由 **CI（含新增的兩個產生檔閘門）＋全域驗證**取代 | 收尾時同一工作樹仍有其他任務在動（並行派工的已知代價），本地全套驗證會與半成品互斥 |

### 未結項（deferred：現況、選項、歸屬）

| # | 項目 | 現況 | 選項 | 歸屬 |
|---|---|---|---|---|
| 1 | **席位 vs 客戶帳號的口徑矛盾** | `CreateCustomer` 會建一列 active 的 `users`（`customer_service.go` 的 `buildCustomerAccount`，:466／:475 呼叫），而席位計數含所有非 `inactive` 帳號（`counters.go` 的 `countFeature`）→ 但該路徑只受 `LimitCustomers` 守衛 → 已達席位上限仍可藉「建客戶」超額佔席位。spec §3.2（席位＝未停用帳號）與 §4.5（只有 `CreateUser` 綁 `limit.seats`）互相打架 | (a) 席位只算**非客戶**帳號（`is_customer=false`）——SaaS 直覺，需改 spec §3.2＋計數器（**建議**）；(b) `CreateCustomer` 也檢查 `limit.seats`——與現行 spec 一致但「加一個客戶吃掉一個員工席位」 | 產品／spec 擁有者定調（會動計費語意，本計畫不自行改） |
| 2 | **spec §4.5 守衛清單缺 `UpdateUser`** | 實作已含 inactive→active 的席位守衛（`7122767`），spec 清單未列 | 回寫 spec §4.5 | spec 擁有者（T12 已在 spec §4.5 補註記） |
| 3 | **`RestoreDepartment` 不存在** | spec §4.5 列了「部門復原」，repo 無此 RPC；`guardCases` 記為具名缺口 | 補 RPC（含守衛）或從 spec 刪列 | `backend-02-tenancy-users`（T12 已在 spec §4.5 補註記） |
| 4 | **`platform.settings` 由 Plan C 的 `00030` 建立** | Plan B 的 seed 對該表採「有表才寫、跳過並印提示」；形狀已與 Plan C 對齊（`key`／`value` 皆 TEXT） | — （已寫成跨計畫硬契約） | Plan C：落地後**必須重跑 `cmd/seed`**，否則 `cmd/platform-cron` 一開跑就 Fatal |
| 5 | **seed 對 `features`／`plans`／`plan_entitlements` 是 `DO UPDATE`** | 重跑會覆寫這三張表的既有值（v1 是產品初始定案，可接受） | console 若開放營運編輯方案權益 → 改成「只補缺」（`NOT EXISTS`／`DO NOTHING`） | Plan C（console 編輯能力落地時） |
| 6 | **`limit.storage_gb` 不在 v1 seed** | 無計數器可量（`internal/services/counters.go` 四種 feature 皆無數量來源） | 檔案功能（04 §3.6，P2-1）落地時連同計數器一起加回 | P2-1（FileStore）＋ spec §4.5 |
| 7 | **v1 無登出端點** | operator token 效期 12h 內只能靠停用 `operators.status` 即時失效（已有測試） | 補 `Logout`（清 cookie／黑名單）或縮短效期 | Plan C／operator auth |
| 8 | **`platform-console/` 未落地** | 目錄不存在 → errcode 產生器對它是 no-op，CI 的 errcode 與 proto 閘門都已把路徑列入（一落地就自動納管） | — | Plan C（T1–T14） |
| 9 | **平台 admin 連線池有兩條** | `mountEntitlements` 與 `mountPlatformAuth` 各自 `database.OpenSQL(AdminDSN())`（process 級、成本可忽略；收斂成單池需改 T5 的 `mountEntitlements` 簽章） | 收斂成一池 | Plan C |

---


| 路徑 | 職責 |
|---|---|
| `config/platform.go`（新） | 平台設定：operator JWT secret、允許的 email 網域、console URL、cookie domain |
| `database/migrations/00029_platform_schema.sql`（新） | `platform` schema ＋ 10 張表 ＋ **明確不授權 `app_rw`** |
| `internal/platform/store/store.go`（新） | 平台資料的**介面**（Plans／Entitlements／Overrides／Subscriptions／Periods／Events／Operators／Audit）與 DTO |
| `internal/platform/store/fake.go`（新） | 記憶體假實作（單元測試與 CLI 使用） |
| `internal/platform/store/postgres/*.go`（新） | `database/sql` 實作（admin 連線） |
| `internal/platform/entitlements/service.go`（新） | `Allows`／`CheckLimit`／`Load`、優先序、fail-closed、快取、`Counter` 介面、`Unlimited()` |
| `internal/platform/entitlements/cache.go`（新） | 快取介面（`Cache`）＋記憶體實作；Valkey 實作由 server 注入 |
| `internal/platform/operatorauth/*.go`（新） | 操作者 OIDC 登入、operator JWT、cookie、RPC interceptor |
| `internal/services/counters.go`（新） | `Counter` 的業務域實作（席位／客戶／商品／部門） |
| `internal/services/entitlement_service.go`（新） | `TenantEntitlementService`（租戶端唯讀投影） |
| `internal/services/platform_admin_service.go`（新） | `PlatformAdminService` 唯讀 RPC（跨租戶視圖）＋ 操作者維運寫入 ＋ 平台稽核 |
| `backend/proto/platform/v1/platform.proto`（新） | 兩個服務：`PlatformAdminService`（operator）／`TenantEntitlementService`（租戶） |
| `internal/services/{user,customer,product,company}_service.go`（改） | 建構子注入 `*entitlements.Service`；7 個 RPC 掛守衛 |
| `internal/server/domains.go`（改） | 組裝：admin `*sql.DB`、store、entitlements、counters、兩個服務、operator interceptor |
| `cmd/seed/platform.go`（新） | 冪等 seed：8 個 features、3 方案 ＋ 價目 ＋ 權益、首位 operator |
| `backend/AGENTS.md`、`.env.example`（改） | 平台域慣例與設定範例 |

---

### Task 1: 平台設定（`config.Platform`）與啟動防護

**Files:**
- Create: `config/platform.go`、`config/platform_test.go`
- Modify: `config/config.go`、`internal/server/server.go`（`Init()` 的 production fail-fast 區塊）

**Interfaces:**
- Produces: `config.Config.Platform`、`config.Platform.OperatorJWTSecret`、`config.Platform.AllowedEmailDomain`、`config.Platform.ConsoleURL`、`config.Platform.CookieDomain`

- [ ] **Step 1: 寫失敗測試（`config/platform_test.go`）**

```go
package config

import "testing"

func TestPlatformFromEnv(t *testing.T) {
	t.Setenv("PLATFORM_JWT_SECRET", "s3cret")
	t.Setenv("PLATFORM_ALLOWED_EMAIL_DOMAIN", "example.com")
	t.Setenv("PLATFORM_CONSOLE_URL", "https://console.example.com")
	t.Setenv("PLATFORM_COOKIE_DOMAIN", ".example.com")
	var p Platform
	mustProcess(&p)
	if p.OperatorJWTSecret != "s3cret" || p.AllowedEmailDomain != "example.com" ||
		p.ConsoleURL != "https://console.example.com" || p.CookieDomain != ".example.com" {
		t.Fatalf("平台設定未正確綁定: %+v", p)
	}
}

func TestPlatformConfigured(t *testing.T) {
	if (Platform{}).Configured() {
		t.Fatal("空設定不得視為已設定")
	}
	// CookieDomain 可為空（開發環境同源代理時用 host-only cookie）；
	// 其餘三項缺一即不掛載。
	noCookie := Platform{
		OperatorJWTSecret: "s", AllowedEmailDomain: "example.com",
		ConsoleURL: "http://localhost:5173",
	}
	if !noCookie.Configured() {
		t.Fatal("CookieDomain 可為空，三項齊備即視為已設定")
	}
	missingSecret := Platform{AllowedEmailDomain: "example.com", ConsoleURL: "http://localhost:5173"}
	if missingSecret.Configured() {
		t.Fatal("缺 OperatorJWTSecret 不得視為已設定")
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./config/ -run 'TestPlatform' -v`
Expected: FAIL —`undefined: Platform`

- [ ] **Step 3: 實作（`config/platform.go`）**

```go
package config

// Platform 為平台營運工具的設定（D38）。整組未設定時平台工具不掛載（開發環境友善），
// production 由 Server.Init() 要求必須設定。
type Platform struct {
	// OperatorJWTSecret 與租戶 JWTSecret **必須不同**：跨用等於平台工具可被租戶 token 冒充。
	OperatorJWTSecret string `envconfig:"PLATFORM_JWT_SECRET"`
	// AllowedEmailDomain 限制 OIDC 登入的 email 網域。
	// 以 email 網域比對（不依賴 Workspace 專屬的 hd claim）→ 相容 Workspace 帳號與
	// 既有 Google 帳號的已驗證別名兩種情況。
	AllowedEmailDomain string `envconfig:"PLATFORM_ALLOWED_EMAIL_DOMAIN" default:"sowinsoft.com"`
	// ConsoleURL 為 OIDC 完成後導回的 console 根網址。
	ConsoleURL string `envconfig:"PLATFORM_CONSOLE_URL"`
	// CookieDomain 為 operator session cookie 的 Domain（空 = host-only，開發環境用）。
	CookieDomain string `envconfig:"PLATFORM_COOKIE_DOMAIN"`

	// --- seed 與排程的預設值（上線前請改為真實值；**執行期以 platform.settings 為準**）---
	// 這些只是「首次建立時寫入 settings」的來源；之後由營運工具調整，重跑 seed 不覆寫。
	SeedOperatorEmail    string `envconfig:"PLATFORM_SEED_OPERATOR_EMAIL" default:"ssd@sowinsoft.com"`
	SeedOperatorName     string `envconfig:"PLATFORM_SEED_OPERATOR_NAME" default:"ssd"`
	SeedSystemActorEmail string `envconfig:"PLATFORM_SEED_SYSTEM_ACTOR_EMAIL" default:"system@sowinsoft.com"`
	DefaultTrialDays     int    `envconfig:"PLATFORM_DEFAULT_TRIAL_DAYS" default:"14"`
	DefaultGraceDays     int    `envconfig:"PLATFORM_DEFAULT_GRACE_DAYS" default:"7"`
	DefaultLeadDays      int    `envconfig:"PLATFORM_DEFAULT_LEAD_DAYS" default:"14"`

	// --- 方案價目的 seed 預設（金額字串，兩位小數）---
	// **這些是佔位數字**：首次建立 `plan_prices` 時使用，之後由營運工具（UpsertPlanPrice）維護；
	// 上線前務必改為真實定價。重跑 seed 不覆寫既有價目。
	SeedPriceFreeBase string `envconfig:"SEED_PRICE_FREE_BASE" default:"0"`
	SeedPriceFreeSeat string `envconfig:"SEED_PRICE_FREE_SEAT" default:"0"`
	SeedPriceStdBase  string `envconfig:"SEED_PRICE_STD_BASE" default:"1500"`
	SeedPriceStdSeat  string `envconfig:"SEED_PRICE_STD_SEAT" default:"150"`
	SeedPriceProBase  string `envconfig:"SEED_PRICE_PRO_BASE" default:"4500"`
	SeedPriceProSeat  string `envconfig:"SEED_PRICE_PRO_SEAT" default:"150"`
}

// Configured 表示必要設定齊備，可掛載平台工具。
// CookieDomain 可為空：同源／代理開發環境使用 host-only cookie（設 Domain=localhost 無效）。
func (p Platform) Configured() bool {
	return p.OperatorJWTSecret != "" && p.AllowedEmailDomain != "" && p.ConsoleURL != ""
}
```

- [ ] **Step 4: 聚合（`config/config.go`）**

```go
type Config struct {
	API           API
	Auth          Auth
	Cache         Cache
	Database      Database
	Platform      Platform
	Storage       Storage
	Observability Observability
	OpenFGA       OpenFGA
}

func New() *Config {
	var c Config
	mustProcess(&c.API)
	mustProcess(&c.Auth)
	mustProcess(&c.Cache)
	mustProcess(&c.Database)
	mustProcess(&c.Platform)
	mustProcess(&c.Storage)
	mustProcess(&c.Observability)
	mustProcess(&c.OpenFGA)
	return &c
}
```

- [ ] **Step 5: production fail-fast（`internal/server/server.go` 的 `Init()`，緊接既有 `JWT_SECRET` 檢查之後）**

```go
		if s.cfg.Platform.Configured() && s.cfg.Platform.OperatorJWTSecret == s.cfg.Auth.JWTSecret {
			return fmt.Errorf("config: PLATFORM_JWT_SECRET 不得與 JWT_SECRET 相同(跨用將使租戶 token 可冒充平台操作者)")
		}
```


- [ ] **Step 6: 跑測試確認通過**

Run: `cd backend && go test ./config/ -run 'TestPlatform' -v && go build ./...`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/config/platform.go backend/config/platform_test.go backend/config/config.go backend/internal/server/server.go
git commit -m "feat(backend): 平台工具設定與啟動防護（PLATFORM_JWT_SECRET 不得與租戶共用）"
```

---

### Task 2: `platform` schema 與 10 張表（migration 00029）

**Files:**
- Create: `database/migrations/00029_platform_schema.sql`
- Create: `internal/platform/store/postgres/schema_integration_test.go`

**Interfaces:**
- Produces: PG schema `platform` 與表 `plans`／`plan_prices`／`features`／`plan_entitlements`／`subscriptions`／`subscription_periods`／`tenant_overrides`／`events`／`operators`／`audit_logs`

- [ ] **Step 1: 寫失敗測試（`internal/platform/store/postgres/schema_integration_test.go`）**

```go
//go:build integration

package postgres_test

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

const platformMigrationsDir = "../../../../database/migrations"

// TestIntegrationPlatformSchema 驗證：
// ① 10 張表都在 platform schema；
// ② 業務角色 app_rw 對該 schema **完全無權限**（S9／§3.3 的硬邊界）。
func TestIntegrationPlatformSchema(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer admin.Close()
	if err := goose.RunContext(t.Context(), "up", admin, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	for _, table := range []string{
		"plans", "plan_prices", "features", "plan_entitlements",
		"subscriptions", "subscription_periods", "tenant_overrides",
		"events", "operators", "audit_logs",
	} {
		var n int
		if err := admin.QueryRow(
			`SELECT count(*) FROM information_schema.tables
			  WHERE table_schema = 'platform' AND table_name = $1`, table,
		).Scan(&n); err != nil {
			t.Fatalf("查 %s: %v", table, err)
		}
		if n != 1 {
			t.Fatalf("platform.%s 不存在", table)
		}
	}

	var usage, selectPriv bool
	if err := admin.QueryRow(
		`SELECT has_schema_privilege('app_rw', 'platform', 'USAGE'),
		        has_table_privilege('app_rw', 'platform.plans', 'SELECT')`,
	).Scan(&usage, &selectPriv); err != nil {
		t.Fatalf("查 app_rw 權限: %v", err)
	}
	if usage || selectPriv {
		t.Fatalf("app_rw 不得對 platform schema 有任何權限(usage=%v select=%v)", usage, selectPriv)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationPlatformSchema -v`
Expected: FAIL —`platform.plans 不存在`（前兩項斷言先紅）

- [ ] **Step 3: 寫 migration（`database/migrations/00029_platform_schema.sql`）**

```sql
-- 平台域 schema(D34–D39)。與業務域**不 JOIN**，只以 company_id 對照。
-- 硬邊界：app_rw(業務連線)對本 schema 零權限；平台域一律走 DATABASE_ADMIN_URL。
-- +goose Up
CREATE SCHEMA IF NOT EXISTS platform;

-- 方案（價目的容器）；被引用後只歸檔不刪除，避免歷史帳斷鏈。
CREATE TABLE IF NOT EXISTS platform.plans (
    id           bigserial PRIMARY KEY,
    code         text        NOT NULL UNIQUE,
    name         text        NOT NULL,
    status       text        NOT NULL DEFAULT 'active',   -- active | archived
    sort_order   integer     NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- 價格史：同一方案可有多次調價，以 effective_from 取「當期生效價」。
CREATE TABLE IF NOT EXISTS platform.plan_prices (
    id             bigserial PRIMARY KEY,
    plan_id        bigint      NOT NULL REFERENCES platform.plans(id),
    billing_cycle  text        NOT NULL,                  -- monthly | yearly
    base_price     numeric(12,2) NOT NULL,
    seat_price     numeric(12,2) NOT NULL,
    currency       text        NOT NULL DEFAULT 'TWD',
    effective_from timestamptz NOT NULL DEFAULT now(),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS plan_prices_plan_effective_idx
    ON platform.plan_prices (plan_id, billing_cycle, effective_from DESC);

-- 可賣的功能與限額清單（**與 OpenFGA 的 resource/action 是不同軸**：這裡管「買了沒有」）。
CREATE TABLE IF NOT EXISTS platform.features (
    code        text        PRIMARY KEY,                  -- limit.seats / feature.printing …
    type        text        NOT NULL,                     -- boolean | integer
    unit        text        NOT NULL DEFAULT '',           -- 席 / 客戶 / 商品 / 部門 / GB
    description text        NOT NULL DEFAULT ''
);

-- 方案 × 功能：boolean 用 enabled，數值上限制用 limit_value（NULL = 不限）。
CREATE TABLE IF NOT EXISTS platform.plan_entitlements (
    plan_id     bigint NOT NULL REFERENCES platform.plans(id),
    feature_code text  NOT NULL REFERENCES platform.features(code),
    enabled     boolean NOT NULL DEFAULT false,
    limit_value bigint,
    PRIMARY KEY (plan_id, feature_code)
);

-- 一租戶一份合約；UNIQUE 保證同時只有一份未取消的訂閱。
CREATE TABLE IF NOT EXISTS platform.subscriptions (
    id               bigserial PRIMARY KEY,
    company_id       bigint      NOT NULL,
    plan_id          bigint      NOT NULL REFERENCES platform.plans(id),
    seat_count       integer     NOT NULL DEFAULT 1,
    -- 計費週期（G1）：期別產生必須依它決定「加一個月」或「加一年」，
    -- 否則年繳方案每次只會產生一個月期別（少收 11 個月）。
    billing_cycle    text        NOT NULL DEFAULT 'monthly', -- monthly | yearly
    status           text        NOT NULL,                -- trialing | active | past_due | suspended | cancelled
    trial_ends_at    timestamptz,
    grace_until      timestamptz,
    dunning_attempts integer     NOT NULL DEFAULT 0,
    payment_provider text        NOT NULL DEFAULT 'manual',
    invoice_provider text        NOT NULL DEFAULT 'manual',
    external_ref     text,
    started_at       timestamptz NOT NULL DEFAULT now(),
    cancelled_at     timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_active_company_unique
    ON platform.subscriptions (company_id) WHERE status <> 'cancelled';

-- 帳的單位：金額／期間／價格快照一律在此（四個「不可回填」欄位）。
CREATE TABLE IF NOT EXISTS platform.subscription_periods (
    id               bigserial PRIMARY KEY,
    subscription_id  bigint      NOT NULL REFERENCES platform.subscriptions(id),
    period_no        integer     NOT NULL,
    period_start     timestamptz NOT NULL,
    period_end       timestamptz NOT NULL,
    plan_id          bigint      NOT NULL,                -- 快照
    unit_price       numeric(12,2) NOT NULL,              -- 快照
    seat_price       numeric(12,2) NOT NULL,              -- 快照
    seat_count       integer     NOT NULL,                -- 快照
    amount           numeric(12,2) NOT NULL,
    currency         text        NOT NULL DEFAULT 'TWD',
    status           text        NOT NULL DEFAULT 'open', -- open | paid | void
    paid_at          timestamptz,
    invoice_no       text,
    invoice_status   text,
    buyer_tax_id     text,
    carrier          text,
    payment_provider text        NOT NULL DEFAULT 'manual',
    external_ref     text,
    -- 短收／溢收等人工註記（G8）：不改變期別金額，只留對帳線索。
    note             text        NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (subscription_id, period_no)
);
-- webhook 冪等：同一 provider 的交易號只能入帳一次（v1 人工時 external_ref 為 NULL）。
CREATE UNIQUE INDEX IF NOT EXISTS periods_provider_ref_unique
    ON platform.subscription_periods (payment_provider, external_ref)
    WHERE external_ref IS NOT NULL;

-- 例外（簽約承諾）；欄位強制可追溯：誰承諾、為何、何時到期。
CREATE TABLE IF NOT EXISTS platform.tenant_overrides (
    id           bigserial PRIMARY KEY,
    company_id   bigint      NOT NULL,
    feature_code text        NOT NULL REFERENCES platform.features(code),
    enabled      boolean,
    limit_value  bigint,
    reason       text        NOT NULL,
    owner        text        NOT NULL,                    -- 承諾者（平台側人員）
    expires_at   timestamptz,
    revoked_at   timestamptz,
    created_by   bigint      NOT NULL,                    -- platform.operators.id
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS tenant_overrides_active_unique
    ON platform.tenant_overrides (company_id, feature_code) WHERE revoked_at IS NULL;

-- outbox：跨域副作用由此驅動（凍結公司、通知）。
CREATE TABLE IF NOT EXISTS platform.events (
    id             bigserial PRIMARY KEY,
    aggregate_type text        NOT NULL,
    aggregate_id   bigint      NOT NULL,
    event_type     text        NOT NULL,
    payload        jsonb       NOT NULL DEFAULT '{}'::jsonb,
    dispatched_at  timestamptz,
    attempts       integer     NOT NULL DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS events_undispatched_idx ON platform.events (created_at) WHERE dispatched_at IS NULL;

-- 平台操作者白名單（S8）：不與租戶 users 有任何關聯。
CREATE TABLE IF NOT EXISTS platform.operators (
    id            bigserial PRIMARY KEY,
    email         text        NOT NULL UNIQUE,
    name          text        NOT NULL DEFAULT '',
    role          text        NOT NULL DEFAULT 'operator', -- operator | admin
    status        text        NOT NULL DEFAULT 'active',   -- active | disabled
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- 平台側稽核（S9）：actor 為 operator_id，**不 FK 租戶 users**（對方無此列）。
CREATE TABLE IF NOT EXISTS platform.audit_logs (
    id          bigserial PRIMARY KEY,
    operator_id bigint      NOT NULL REFERENCES platform.operators(id),
    action      text        NOT NULL,
    target_type text        NOT NULL,                      -- company | plan | subscription | operator
    target_id   text        NOT NULL,
    reason      text        NOT NULL DEFAULT '',
    before      jsonb,
    after       jsonb,
    ip_address  text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS platform_audit_created_idx ON platform.audit_logs (created_at DESC);

-- 硬邊界：撤銷業務角色對本 schema 的一切權限（含 plan_a 的 ALTER DEFAULT PRIVILEGES 未來影響）
REVOKE ALL ON SCHEMA platform FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA platform FROM app_rw;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA platform FROM app_rw;

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS platform CASCADE;
-- +goose StatementEnd
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && task test:integration -- -run TestIntegrationPlatformSchema -v`
Expected: PASS

- [ ] **Step 5: 驗證回滾**

Run: `cd backend && task test:integration -- -run TestIntegrationMigrateDown -v`
Expected: PASS（`DROP SCHEMA … CASCADE` 可完整回滾）

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/00029_platform_schema.sql backend/internal/platform/store/postgres/schema_integration_test.go
git commit -m "feat(backend): platform schema 與 10 張表（00029），業務角色零權限"
```

---

### Task 3: 平台 store（介面 ＋ 記憶體假實作 ＋ PostgreSQL 實作）

**Files:**
- Create: `internal/platform/store/store.go`、`internal/platform/store/fake.go`
- Create: `internal/platform/store/postgres/store.go`
- Create: `internal/platform/store/postgres/store_integration_test.go`

**Interfaces:**
- Produces:

```go
// store.go
type Feature struct {
	Code        string
	Type        string // boolean | integer
	Unit        string
	Description string
}

type Entitlement struct { // 方案預設
	FeatureCode string
	Enabled     bool
	Limit       *int64
}

type Override struct { // 租戶例外
	FeatureCode string
	Enabled     *bool
	Limit       *int64
	ExpiresAt   *time.Time
}

type Subscription struct {
	CompanyID int
	PlanCode  string
	Status    string // trialing | active | past_due | suspended | cancelled
	PlanID    int64
	SeatCount int
	// BillingCycle 決定期別產生時「加一個月」或「加一年」（G1）。
	BillingCycle string // monthly | yearly
	TrialEnds    *time.Time
	GraceUntil   *time.Time
}

type Store interface {
	// Features 回傳全部功能定義（一次載入，供型別判定與 UI 顯示）。
	Features(ctx context.Context) (map[string]Feature, error)
	// PlanEntitlements 回傳某方案的權益（方案以 code 指定）。
	PlanEntitlements(ctx context.Context, planCode string) ([]Entitlement, error)
	// Overrides 回傳某租戶「未撤銷」的例外（未過濾到期；判定層負責）。
	Overrides(ctx context.Context, companyID int) ([]Override, error)
	// Subscription 回傳某租戶未取消的訂閱；無則回 (nil, nil)。
	Subscription(ctx context.Context, companyID int) (*Subscription, error)
}
```

- [ ] **Step 1: 寫失敗測試（`internal/platform/store/store_test.go`，測假實作的行為契約）**

```go
package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// fake 必須忠實反映「未撤銷」與「未取消」的過濾語意，否則 entitlement 的單元測試會失真。
func TestFakeStoreFiltersRevokedAndCancelled(t *testing.T) {
	ctx := context.Background()
	expired := time.Now().Add(-time.Hour)
	f := store.NewFake()
	f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "std", Status: "active"})
	now := time.Now()
	f.PutOverride(store.Override{CompanyID: 7, FeatureCode: "limit.seats", Limit: ptr(int64(50))})
	f.PutOverride(store.Override{CompanyID: 7, FeatureCode: "feature.printing", Enabled: ptr(true), ExpiresAt: &expired})

	sub, err := f.Subscription(ctx, 7)
	if err != nil || sub == nil || sub.PlanCode != "std" {
		t.Fatalf("應取得訂閱,got %+v err=%v", sub, err)
	}
	ov, err := f.Overrides(ctx, 7)
	if err != nil {
		t.Fatalf("取 overrides: %v", err)
	}
	if len(ov) != 2 {
		t.Fatalf("假實作回傳未撤銷的 override（到期過濾由判定層負責）,got %d", len(ov))
	}
	if sub, err := f.Subscription(ctx, 8); err != nil || sub != nil {
		t.Fatalf("無訂閱的租戶應回 (nil, nil),got %+v err=%v", sub, err)
	}
}

func ptr[T any](v T) *T { return &v }
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/platform/store/ -v`
Expected: FAIL —`undefined: store.NewFake`

- [ ] **Step 3: 實作介面與假實作（`store.go`、`fake.go`）**

`store.go` 依 Interfaces 段落貼上（含套件宣告與 `time`／`context` import）。

`fake.go`：

```go
package store

import (
	"context"
	"sync"
)

// Fake 為記憶體實作：供 entitlement 單元測試與 CLI 使用（不進 production 路徑）。
type Fake struct {
	mu       sync.Mutex
	features map[string]Feature
	plans    map[string][]Entitlement
	subs     map[int]Subscription
	overs    map[int][]Override
}

func NewFake() *Fake {
	return &Fake{
		features: map[string]Feature{},
		plans:    map[string][]Entitlement{},
		subs:     map[int]Subscription{},
		overs:    map[int][]Override{},
	}
}

func (f *Fake) PutFeature(x Feature) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.features[x.Code] = x
}

func (f *Fake) PutPlan(planCode string, ents []Entitlement) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.plans[planCode] = ents
}

func (f *Fake) PutSubscription(s Subscription) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subs[s.CompanyID] = s
}

func (f *Fake) PutOverride(o Override) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.overs[o.CompanyID] = append(f.overs[o.CompanyID], o)
}

func (f *Fake) Features(context.Context) (map[string]Feature, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]Feature, len(f.features))
	for k, v := range f.features {
		out[k] = v
	}
	return out, nil
}

func (f *Fake) PlanEntitlements(_ context.Context, planCode string) ([]Entitlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.plans[planCode], nil
}

func (f *Fake) Overrides(_ context.Context, companyID int) ([]Override, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Override, 0, len(f.overs[companyID]))
	for _, o := range f.overs[companyID] {
		if o.ExpiresAt == nil || o.ExpiresAt.After(nowFunc()) { // 未撤銷者全回，到期由判定層處理
			out = append(out, o)
		}
	}
	return out, nil
}

func (f *Fake) Subscription(_ context.Context, companyID int) (*Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.subs[companyID]
	if !ok || s.Status == "cancelled" {
		return nil, nil
	}
	return &s, nil
}

// nowFunc 供測試覆寫時間（entitlement 的到期判定）。
var nowFunc = time.Now
```

（`fake.go` 需 import `time`。`nowFunc` 宣告於此檔供套件內共用。）

- [ ] **Step 4: 實作 PostgreSQL store（`internal/platform/store/postgres/store.go`）**

```go
// Package postgres 以 database/sql 實作平台 store（admin 連線；platform schema 不用 ent）。
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Features(ctx context.Context) (map[string]store.Feature, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT code, type, unit, description FROM platform.features`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]store.Feature{}
	for rows.Next() {
		var f store.Feature
		if err := rows.Scan(&f.Code, &f.Type, &f.Unit, &f.Description); err != nil {
			return nil, err
		}
		out[f.Code] = f
	}
	return out, rows.Err()
}

func (s *Store) PlanEntitlements(ctx context.Context, planCode string) ([]store.Entitlement, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT pe.feature_code, pe.enabled, pe.limit_value
		  FROM platform.plan_entitlements pe
		  JOIN platform.plans p ON p.id = pe.plan_id
		 WHERE p.code = $1`, planCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Entitlement
	for rows.Next() {
		var e store.Entitlement
		var limit sql.NullInt64
		if err := rows.Scan(&e.FeatureCode, &e.Enabled, &limit); err != nil {
			return nil, err
		}
		if limit.Valid {
			v := limit.Int64
			e.Limit = &v
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) Overrides(ctx context.Context, companyID int) ([]store.Override, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT feature_code, enabled, limit_value, expires_at
		  FROM platform.tenant_overrides
		 WHERE company_id = $1 AND revoked_at IS NULL
		 ORDER BY created_at DESC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Override
	for rows.Next() {
		var o store.Override
		var enabled sql.NullBool
		var limit sql.NullInt64
		var expires sql.NullTime
		if err := rows.Scan(&o.FeatureCode, &enabled, &limit, &expires); err != nil {
			return nil, err
		}
		if enabled.Valid {
			v := enabled.Bool
			o.Enabled = &v
		}
		if limit.Valid {
			v := limit.Int64
			o.Limit = &v
		}
		if expires.Valid {
			v := expires.Time
			o.ExpiresAt = &v
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) Subscription(ctx context.Context, companyID int) (*store.Subscription, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT s.company_id, p.code, s.status, p.id, s.seat_count, s.trial_ends_at
		  FROM platform.subscriptions s
		  JOIN platform.plans p ON p.id = s.plan_id
		 WHERE s.company_id = $1 AND s.status <> 'cancelled'`, companyID)
	var sub store.Subscription
	var trial sql.NullTime
	err := row.Scan(&sub.CompanyID, &sub.PlanCode, &sub.Status, &sub.PlanID, &sub.SeatCount, &trial)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if trial.Valid {
		v := trial.Time
		sub.TrialEnds = &v
	}
	return &sub, nil
}

var _ store.Store = (*Store)(nil)

```

- [ ] **Step 5: 整合測試（`store_integration_test.go`）**

```go
//go:build integration

package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationPlatformStore 以真 PostgreSQL 驗證四項查詢：方案權益（含 NULL limit）、
// 未撤銷 override、未取消訂閱、以及「無訂閱回 nil」。
func TestIntegrationPlatformStore(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer db.Close()
	if err := goose.RunContext(t.Context(), "up", db, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.features (code, type, unit, description) VALUES
		('limit.seats','integer','席','席位上線'),
		('feature.printing','boolean','','列印')`); err != nil {
		t.Fatalf("seed features: %v", err)
	}
	var planID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value) VALUES
		($1,'limit.seats',true,10), ($1,'feature.printing',true,NULL)`, planID); err != nil {
		t.Fatalf("seed entitlements: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count)
		VALUES (42, $1, 'active', 10)`, planID); err != nil {
		t.Fatalf("seed subscription: %v", err)
	}
	live := time.Now().Add(24 * time.Hour)
	dead := time.Now().Add(-24 * time.Hour)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.tenant_overrides
			(company_id, feature_code, enabled, limit_value, reason, owner, expires_at, created_by)
		VALUES (42,'limit.seats',NULL,50,'簽約承諾','sales@example.com',$1,1),
		       (42,'feature.printing',true,NULL,'短期試用','sales@example.com',$2,1)`, live, dead); err != nil {
		t.Fatalf("seed overrides: %v", err)
	}

	st := postgres.New(db)
	ents, err := st.PlanEntitlements(ctx, "std")
	if err != nil || len(ents) != 2 {
		t.Fatalf("PlanEntitlements: got %d err=%v", len(ents), err)
	}
	sub, err := st.Subscription(ctx, 42)
	if err != nil || sub == nil || sub.PlanCode != "std" || sub.Status != "active" {
		t.Fatalf("Subscription: got %+v err=%v", sub, err)
	}
	if sub, err := st.Subscription(ctx, 43); err != nil || sub != nil {
		t.Fatalf("無訂閱應回 nil,got %+v err=%v", sub, err)
	}
	ov, err := st.Overrides(ctx, 42)
	if err != nil || len(ov) != 2 {
		t.Fatalf("Overrides 應回未撤銷的兩筆（到期由判定層過濾）,got %d err=%v", len(ov), err)
	}
}
```

- [ ] **Step 6: 跑測試**

Run: `cd backend && go test ./internal/platform/store/... -v && task test:integration -- -run 'TestIntegrationPlatform' -v`
Expected: 單元 PASS；整合 PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/platform/store
git commit -m "feat(backend): 平台 store（介面＋假實作＋PostgreSQL 實作）"
```

---

### Task 4: 權益判定（`entitlements.Service`）

**錯誤碼（對外契約；由 Plan D 提供，`internal/errcode` 已可用——Plan D T1–T6 已完成，碼表見 `docs/error-codes.md`）**：本任務的 `FailedPrecondition` 一律換成註冊碼，**不得**自建 `connect.NewError(connect.CodeFailedPrecondition, …)`：

| 情境 | 碼（connect 碼不變） |
|---|---|
| `Allows` 判定「方案／override 未含此功能」 | `errcode.PlatformFeatureNotInPlan`（`PLAT-5002`，details `feature`） |
| `CheckLimit` 判定超過上限 | `errcode.PlatformLimitExceeded`（`PLAT-5001`，details `feature`／`used`／`limit`） |
| 訂閱 `suspended`／`cancelled`（合約不可用） | `errcode.PlatformSubscriptionInactive`（`PLAT-3001`） |

落地本身列在 Plan D 的 **Task 5b**（該任務負責讓四個 `PLAT-*` 在 Plan B／C 的路徑上真的被回傳並有測試斷言）——若本任務先實作，請直接照上表寫碼，避免事後再改一次。本任務下方的 `CheckLimit`／`Allows` 片段已按上表寫；若實作時仍看到任何自建的 `connect.NewError(connect.CodeFailedPrecondition, …)`，一律換成上表的碼。

**Files:**
- Create: `internal/platform/entitlements/service.go`、`internal/platform/entitlements/cache.go`
- Create: `internal/platform/entitlements/service_test.go`

**Interfaces:**
- Consumes: `store.Store`（Task 3）
- Produces:

```go
const (
	LimitSeats       = "limit.seats"
	LimitCustomers   = "limit.customers"
	LimitProducts    = "limit.products"
	LimitDepartments = "limit.departments"
	LimitStorageGB   = "limit.storage_gb"
	FeaturePrinting  = "feature.printing"
	FeatureDispatch  = "feature.dispatch"
	FeatureReturns   = "feature.returns"
)

// Counter 由業務域提供（於 server.InitDomains 注入）：判定層不認得業務 schema。
type Counter interface {
	Count(ctx context.Context, companyID int, feature string) (int, error)
}

type Service struct{ /* store, counters, cache, now */ }

func New(st store.Store, counters Counter, c Cache, ttl time.Duration) *Service
func Unlimited() *Service // 測試與 CLI 用：全部允許、不限額
func (s *Service) Allows(ctx context.Context, companyID int, feature string) (bool, error)
func (s *Service) CheckLimit(ctx context.Context, companyID int, feature string, delta int) error
func (s *Service) Snapshot(ctx context.Context, companyID int) (*Snapshot, error) // 供租戶端投影
```

- [ ] **Step 1: 寫失敗測試（`service_test.go`）— 先寫最關鍵的四條**

```go
package entitlements_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

const seats = entitlements.LimitSeats

// errorCodeOf 由 connect error 取 ErrorInfo.code（本套件的唯一解析點；與 internal/services 的
// errorInfoOf 同構——對外碼才是前端據以導向升級方案的依據）。
func errorCodeOf(t *testing.T, err error) string {
	t.Helper()
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("非 connect error: %v", err)
	}
	for _, d := range ce.Details() {
		if info, ok := d.Value().(*commonv1.ErrorInfo); ok {
			return info.GetCode()
		}
	}
	t.Fatalf("錯誤未帶 ErrorInfo（碼到不了客戶端）: %v", err)
	return ""
}

func newSvc(f *store.Fake, counts map[string]int) *entitlements.Service {
	return entitlements.New(f, counting(counts), entitlements.NewMemoryCache(), 0)
}

type counting map[string]int

func (c counting) Count(_ context.Context, _ int, feature string) (int, error) { return c[feature], nil }

func TestNoSubscriptionDeniesEverything(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10))}})
	got, err := newSvc(f, nil).Allows(context.Background(), 1, seats)
	if err != nil {
		t.Fatalf("Allows: %v", err)
	}
	if got {
		t.Fatal("無訂閱時必須 fail-closed（不得因為方案有定義就放行）")
	}
}

func TestLimitBlockedReturnsFailedPrecondition(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})

	svc := newSvc(f, counting{seats: 10})
	err := svc.CheckLimit(context.Background(), 1, seats, 1)
	if err == nil {
		t.Fatal("已達上限時新增應被擋")
	}
	var cerr *connect.Error
	if !errors.As(err, &cerr) || cerr.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("額度不足必須回 FailedPrecondition（引導升級方案），got %v", err)
	}
	// 對外碼才是前端據以導向升級方案的依據（connect 碼不足以區分額度與權限）。
	if got := errorCodeOf(t, err); got != "PLAT-5001" {
		t.Fatalf("額度不足必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}

	ok := newSvc(f, counting{seats: 9})
	if err := ok.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("未達上限應放行: %v", err)
	}
}

func TestOverrideBeatsPlanAndExpiredOverrideIgnored(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})
	past := time.Now().Add(-time.Hour)
	f.PutOverride(store.Override{CompanyID: 1, FeatureCode: seats, Limit: ptr(int64(50)), ExpiresAt: &past})

	if err := newSvc(f, counting{seats: 20}).CheckLimit(context.Background(), 1, seats, 1); err == nil {
		t.Fatal("已過期的 override 不得生效（應回到方案的 10）")
	}

	f.PutOverride(store.Override{CompanyID: 1, FeatureCode: seats, Limit: ptr(int64(50))})
	if err := newSvc(f, counting{seats: 20}).CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("未過期的 override 應放行（50）: %v", err)
	}
}

func TestSuspendedSubscriptionDenied(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: entitlements.FeaturePrinting, Type: "boolean"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: entitlements.FeaturePrinting, Enabled: true}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "suspended"})

	got, err := newSvc(f, nil).Allows(context.Background(), 1, entitlements.FeaturePrinting)
	if err != nil {
		t.Fatalf("Allows: %v", err)
	}
	if got {
		t.Fatal("訂閱 suspended 時不得允許使用")
	}
}

func ptr[T any](v T) *T { return &v }
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/platform/entitlements/ -v`
Expected: FAIL —`undefined: entitlements.New`

- [ ] **Step 3: 實作（`service.go`）——含快取**

```go
// Package entitlements 判定「這個租戶現在可以做什麼」：方案預設 ⊕ 未過期 override，
// 並在訂閱不可用時全面 fail-closed。與 OpenFGA 的角色授權是**不同軸**：
// 這裡管「買了沒有」，那裡管「誰能做」。
package entitlements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

const (
	LimitSeats       = "limit.seats"
	LimitCustomers   = "limit.customers"
	LimitProducts    = "limit.products"
	LimitDepartments = "limit.departments"
	LimitStorageGB   = "limit.storage_gb"
	FeaturePrinting  = "feature.printing"
	FeatureDispatch  = "feature.dispatch"
	FeatureReturns   = "feature.returns"
)

// Counter 由業務域提供；判定層不得直接認得業務 schema。
type Counter interface {
	Count(ctx context.Context, companyID int, feature string) (int, error)
}

// tenantState 為單一租戶的權益來源快照（快取的單位）。
type tenantState struct {
	CompanyID   int                          `json:"company_id"`
	PlanCode    string                       `json:"plan_code"`
	PlanName    string                       `json:"plan_name"`
	Status      string                       `json:"status"`
	TrialEnds   *time.Time                   `json:"trial_ends_at,omitempty"`
	Features    map[string]store.Feature     `json:"features"`
	Entitlements map[string]store.Entitlement `json:"entitlements"`
	Overrides   []store.Override             `json:"overrides"`
}

// Service 為權益判定入口。
type Service struct {
	st        store.Store
	counters  Counter
	cache     Cache
	ttl       time.Duration
	now       func() time.Time
	unlimited bool
}

// New 建立判定服務；ttl <= 0 表示不快取。
func New(st store.Store, counters Counter, c Cache, ttl time.Duration) *Service {
	return &Service{st: st, counters: counters, cache: c, ttl: ttl, now: time.Now}
}

// Unlimited 回傳「全部允許、不限額」的實例：測試與 CLI 使用，不進 production 路徑。
func Unlimited() *Service { return &Service{unlimited: true, now: time.Now} }

func cacheKey(companyID int) string { return fmt.Sprintf("ent:%d", companyID) }

// state 取得租戶權益來源：快取命中即回，未命中則從 store 組裝並寫回。
// 無訂閱（或訂閱不可用）時回傳 status 對應的狀態，由判定層決定拒絕。
func (s *Service) state(ctx context.Context, companyID int) (*tenantState, error) {
	if s.cache != nil && s.ttl > 0 {
		if raw, ok, err := s.cache.Get(ctx, cacheKey(companyID)); err != nil {
			return nil, err
		} else if ok {
			var st tenantState
			if err := json.Unmarshal(raw, &st); err == nil {
				return &st, nil
			}
			// 反序列化失敗視為未命中（不讓壞快取擋住服務）
		}
	}

	sub, err := s.st.Subscription(ctx, companyID)
	if err != nil {
		return nil, err
	}
	out := &tenantState{CompanyID: companyID, Features: map[string]store.Feature{},
		Entitlements: map[string]store.Entitlement{}}
	if sub == nil {
		out.Status = "none"
	} else {
		out.PlanCode, out.PlanName, out.Status, out.TrialEnds = sub.PlanCode, sub.PlanName, sub.Status, sub.TrialEnds
	}

	features, err := s.st.Features(ctx)
	if err != nil {
		return nil, err
	}
	out.Features = features

	if sub != nil {
		ents, err := s.st.PlanEntitlements(ctx, sub.PlanCode)
		if err != nil {
			return nil, err
		}
		for _, e := range ents {
			out.Entitlements[e.FeatureCode] = e
		}
		overs, err := s.st.Overrides(ctx, companyID)
		if err != nil {
			return nil, err
		}
		out.Overrides = overs
	}

	if s.cache != nil && s.ttl > 0 {
		if raw, err := json.Marshal(out); err == nil {
			if err := s.cache.Set(ctx, cacheKey(companyID), raw, s.ttl); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// resolved 為「不經快取」的組裝結果（供 Snapshot 使用）。
type resolved struct {
	feature store.Feature
	enabled bool
	limit   *int64
}

// resolveFeature 依 tenantState 計算單一 feature 的最終權益（純函式：可單獨測試）。
func resolveFeature(st *tenantState, feature string, now time.Time) (resolved, bool) {
	def, known := st.Features[feature]
	if !known {
		return resolved{}, false // 未定義的功能一律 denied（fail-closed）
	}
	if st.Status == "none" || st.Status == "suspended" || st.Status == "cancelled" {
		return resolved{feature: def}, true
	}

	out := resolved{feature: def}
	if e, ok := st.Entitlements[feature]; ok {
		out.enabled, out.limit = e.Enabled, e.Limit
	}
	for _, o := range st.Overrides { // 未撤銷；到期在此過濾；最特定者勝
		if o.FeatureCode != feature {
			continue
		}
		if o.ExpiresAt != nil && !o.ExpiresAt.After(now) {
			continue
		}
		if o.Enabled != nil {
			out.enabled = *o.Enabled
		}
		if o.Limit != nil {
			out.limit = o.Limit
		}
	}
	if st.Status == "trialing" {
		out.enabled = true // 試用期內不因方案 disabled 而擋（方案仍是上限來源）
	}
	return out, true
}

// Allows 判定 boolean 功能是否可用。
func (s *Service) Allows(ctx context.Context, companyID int, feature string) (bool, error) {
	if s.unlimited {
		return true, nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return false, connect.NewError(connect.CodeInternal, fmt.Errorf("讀取權益失敗: %w", err))
	}
	r, _ := resolveFeature(st, feature, s.now())
	return r.enabled, nil
}

// CheckLimit 檢查「再加 delta 是否超過上限」；額度不足回 FailedPrecondition。
func (s *Service) CheckLimit(ctx context.Context, companyID int, feature string, delta int) error {
	if s.unlimited {
		return nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("讀取權益失敗: %w", err))
	}
	r, known := resolveFeature(st, feature, s.now())
	if !known || !r.enabled {
		// PLAT-5002：方案／override 未含此功能（碼見本任務開頭的錯誤碼表）
		return errcode.PlatformFeatureNotInPlan.Error(map[string]string{"feature": feature})
	}
	if r.limit == nil {
		return nil // 不限
	}
	if s.counters == nil {
		return connect.NewError(connect.CodeInternal, errors.New("未注入計數器"))
	}
	cur, err := s.counters.Count(ctx, companyID, feature)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("計算用量失敗: %w", err))
	}
	if cur+delta > int(*r.limit) {
		// PLAT-5001：已達上限；used／limit 進 details 供前端顯示用量（碼見本任務開頭的錯誤碼表）
		return errcode.PlatformLimitExceeded.Error(map[string]string{
			"feature": feature,
			"used":    strconv.Itoa(cur),
			"limit":   strconv.FormatInt(*r.limit, 10),
		})
	}
	return nil
}
```

- [ ] **Step 4: 快取介面與記憶體實作（`cache.go`）**

```go
package entitlements

import (
	"context"
	"sync"
	"time"
)

// Cache 抽象權益快取：單元測試用記憶體、production 注入 Valkey 實作。
// 失效語意：方案／override／訂閱異動時由寫入方 Delete（不靠 TTL 正確性），TTL 僅保底。
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// MemoryCache 為程序內快取（單元測試與單 replica 開發用）。
type MemoryCache struct {
	mu   sync.Mutex
	data map[string][]byte
}

func NewMemoryCache() *MemoryCache { return &MemoryCache{data: map[string][]byte{}} }

func (m *MemoryCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, val []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
	return nil
}

func (m *MemoryCache) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}
```

- [ ] **Step 5: 補 `Snapshot`（Task 10 的租戶端投影用；`snapshot.go`）**

```go
package entitlements

import "context"

// FeatureUsage 為單一 feature 的判定結果與用量（供租戶端投影；boolean 的 Used 為 0）。
type FeatureUsage struct {
	FeatureCode string
	Enabled     bool
	Limit       *int64
	Used        int
}

// Snapshot 為租戶的權益全貌（租戶端投影；只回自己公司）。
type Snapshot struct {
	PlanCode    string
	PlanName    string
	Status      string
	TrialEndsAt *time.Time
	Usage       []FeatureUsage
}

// Snapshot 組裝租戶權益與用量；feature 清單以 store 的定義為準（未定義者不出現）。
func (s *Service) Snapshot(ctx context.Context, companyID int) (*Snapshot, error) {
	if s.unlimited {
		return &Snapshot{PlanCode: "unlimited", PlanName: "不限", Status: "active"}, nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return nil, err
	}
	out := &Snapshot{PlanCode: st.PlanCode, PlanName: st.PlanName, Status: st.Status, TrialEndsAt: st.TrialEnds}
	for code, def := range st.Features {
		r, _ := resolveFeature(st, code, s.now())
		u := FeatureUsage{FeatureCode: code, Enabled: r.enabled, Limit: r.limit}
		if def.Type == "integer" && s.counters != nil {
			n, err := s.counters.Count(ctx, companyID, code)
			if err != nil {
				return nil, err
			}
			u.Used = n
		}
		out.Usage = append(out.Usage, u)
	}
	sort.Slice(out.Usage, func(i, j int) bool { return out.Usage[i].FeatureCode < out.Usage[j].FeatureCode })
	return out, nil
}
```

（`snapshot.go` 需 import `time` 與 `sort`。）

- [ ] **Step 6: 補「快取真的生效且不改變判定」的測試（加到 `service_test.go`）**

```go
// countingStore 記錄對底層 store 的查詢次數，用來證明第二次判定走的是快取。
type countingStore struct {
	store.Store
	subCalls int
}

func (c *countingStore) Subscription(ctx context.Context, companyID int) (*store.Subscription, error) {
	c.subCalls++
	return c.Store.Subscription(ctx, companyID)
}

func TestCacheServesSecondCallWithoutStoreHit(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: seats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: seats, Enabled: true, Limit: ptr(int64(2))}})
	f.PutSubscription(store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active"})

	cs := &countingStore{Store: f}
	svc := entitlements.New(cs, counting{seats: 1}, entitlements.NewMemoryCache(), time.Minute)

	if err := svc.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("第一次（未達上限）應放行: %v", err)
	}
	first := cs.subCalls
	if err := svc.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("第二次應放行: %v", err)
	}
	if cs.subCalls != first {
		t.Fatalf("第二次判定應命中快取（store 查詢次數 %d → %d）", first, cs.subCalls)
	}

	// 快取不得讓判定失真：換一個「已達上限」的計數器，結果仍須被擋。
	if err := entitlements.New(cs, counting{seats: 2}, entitlements.NewMemoryCache(), time.Minute).
		CheckLimit(context.Background(), 1, seats, 1); err == nil {
		t.Fatal("2/2 再加 1 應被擋")
	}
}
```

- [ ] **Step 7: 跑測試確認通過**

Run: `cd backend && go test ./internal/platform/... -v`
Expected: PASS（含 Task 3 的 store 測試）

- [ ] **Step 7: Commit**

```bash
git add backend/internal/platform/entitlements
git commit -m "feat(backend): 權益判定服務（方案⊕override、fail-closed、FailedPrecondition、可快取）"
```

---

### Task 5: 計數器與服務注入（編譯器驅動）

**Files:**
- Create: `internal/services/counters.go`
- Modify: `internal/services/user_service.go`、`customer_service.go`、`product_service.go`、`company_service.go`（建構子與 register 簽名）
- Modify: `internal/server/domains.go`
- Modify: 既有 service 測試（以編譯錯誤清單逐檔補 `entitlements.Unlimited()`）

**Interfaces:**
- Consumes: `entitlements.Service`、`entitlements.Counter`
- Produces: `services.NewEntitlementCounter(db *ent.Client) entitlements.Counter`

- [ ] **Step 1: 寫失敗測試（`internal/services/counters_test.go`，sqlite 可跑）**

```go
package services

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

// 席位的計數規則（spec §3.2）：只算未軟刪除且非 inactive 的帳號 ——
// 停用可釋放席位，客戶才能自助降級。
func TestEntitlementCounterSeatsExcludesDeletedAndInactive(t *testing.T) {
	db := newTestDB(t) // 既有 sqlite 測試輔助
	ctx := context.Background()
	co := db.Company.Create().SetName("C").SetIdentifier("CNT-1").SaveX(ctx)
	db.User.Create().SetEmail("a@example.com").SetName("A").SetStatus("active").
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("b@example.com").SetName("B").SetStatus("inactive").
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("c@example.com").SetName("C").SetStatus("active").
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).
		SetDeletedAt(timeNow()).SaveX(ctx)

	got, err := NewEntitlementCounter(db).Count(ctx, co.ID, entitlements.LimitSeats)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 1 {
		t.Fatalf("席位應只算 1（active 且未刪除），got %d", got)
	}
}
```

（`newTestDB`／`timeNow` 為既有 sqlite 測試輔助；實作時沿用該檔已用的輔助名稱。）

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/services/ -run TestEntitlementCounter -v`
Expected: FAIL —`undefined: NewEntitlementCounter`

- [ ] **Step 3: 實作（`internal/services/counters.go`）**

```go
// 權益計數器：以業務 ent client 計算「有效筆數」（平台域不認得業務 schema，故由本域提供）。
package services

import (
	"context"
	"fmt"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

type entitlementCounter struct{ db *ent.Client }

// NewEntitlementCounter 建立計數器（於 server.InitDomains 注入 entitlement 服務）。
func NewEntitlementCounter(db *ent.Client) entitlements.Counter {
	return &entitlementCounter{db: db}
}

// Count 依 feature 計算有效用量：
// 席位 = 未軟刪除且非 inactive 的帳號；其餘 = 未軟刪除的列數（spec §3.2 三條規則）。
func (c *entitlementCounter) Count(ctx context.Context, companyID int, feature string) (int, error) {
	switch feature {
	case entitlements.LimitSeats:
		return c.db.User.Query().
			Where(user.CompanyIDEQ(companyID), user.DeletedAtIsNil(),
				user.StatusNEQ(user.StatusInactive)).Count(ctx)
	case entitlements.LimitCustomers:
		return c.db.Customer.Query().
			Where(customer.CompanyIDEQ(companyID), customer.DeletedAtIsNil()).Count(ctx)
	case entitlements.LimitProducts:
		return c.db.Product.Query().
			Where(product.CompanyIDEQ(companyID), product.DeletedAtIsNil()).Count(ctx)
	case entitlements.LimitDepartments:
		return c.db.Department.Query().
			Where(department.CompanyIDEQ(companyID), department.DeletedAtIsNil()).Count(ctx)
	default:
		return 0, fmt.Errorf("未定義的計數 feature: %s", feature)
	}
}
```

- [ ] **Step 4: 四個服務改為建構子注入（編譯器強制所有呼叫端表態）**

`user_service.go`：

```go
type UserService struct {
	db   *ent.Client
	ent  *entitlements.Service
	salesorderv1connect.UnimplementedUserServiceHandler
}

func NewUserService(db *ent.Client, entSvc *entitlements.Service) *UserService {
	return &UserService{db: db, ent: entSvc}
}

func RegisterUserServices(mux *http.ServeMux, db *ent.Client, entSvc *entitlements.Service) {
	path, handler := salesorderv1connect.NewUserServiceHandler(NewUserService(db, entSvc))
	mux.Handle(path, handler)
}
```

`customer_service.go`、`product_service.go`、`company_service.go`（`DepartmentService` 同檔）比照：欄位加 `ent *entitlements.Service`、建構子與 register 各加一個參數（`RegisterCustomerServices(mux, db, frontendURL, entSvc)` 順序照原簽名尾端追加）。

- [ ] **Step 5: 修正所有編譯錯誤（測試與呼叫端）**

```bash
cd backend && go build ./... 2>&1 | head -40
go vet ./... 2>&1 | head -20
```
逐檔把測試中的 `NewXService(db)` / `RegisterXServices(mux, db…)` 補上 `entitlements.Unlimited()`（測試語意是「不受方案限制，專注測該域行為」；額度行為另由 Task 6 的守衛測試覆蓋）。

- [ ] **Step 6: 組裝（`internal/server/domains.go`）**

**[controller 裁定，2026-09-20；實作採此版，勿照下方原稿]** 原稿的「admin 連線失敗 → `log`＋略過掛載」是**錯的**：那會讓四個業務 `register` 落空＝**四個業務 RPC 整組不掛載**（比「無守衛」更糟，且症狀是「功能消失」）。正解是**平台層不可用時拒絕啟動**（fail-fast，與其他 production 守護一致），並用 `internal/third_party/database` 的集中開啟（`database.OpenSQL(AdminDSN)`，會 Ping、符合 D31），而不是裸 `sql.Open`。

```go
	// 平台域（admin 連線）：store → entitlement → counters → 注入四個業務服務。
	// 契約：平台層不可用 → 拒絕啟動（不得略過掛載：那會讓四個業務 RPC 整組消失）。
	adminDB, err := database.OpenSQL(s.cfg.Database.AdminDSN())
	if err != nil {
		return fmt.Errorf("config: 無法開啟平台 admin 連線（平台權益守衛是業務寫入的前置條件）: %w", err)
	}
	platformStore := postgresstore.New(adminDB)
	entSvc := entitlements.New(platformStore, services.NewEntitlementCounter(entClient),
		entitlements.NewMemoryCache(), 60*time.Second)
	s.entitlements = entSvc
	services.RegisterUserServices(apiMux, entClient, entSvc)
	services.RegisterCustomerServices(apiMux, entClient, s.cfg.Auth.FrontendURL, entSvc)
	services.RegisterProductService(apiMux, entClient, entSvc)
```

**[controller 裁定]** 守衛要能被**記錄式假物件**注入（T6 的表驅動測試要驗「RPC → feature」對應），因此四個服務的建構子收的是 **consumer-side 最小介面**（在 `internal/services` 定義，例：`type entitlementChecker interface { CheckLimit(ctx, companyID, feature, delta) error }`），**不是**具體型別 `*entitlements.Service`。Go 慣例：接受介面、回傳結構。`domains.go` 傳真 `*entitlements.Service`（滿足該介面）即可。T10 需要的是另一組方法（`Allows`／`Load`／`Snapshot`）→ **另立窄介面**，不要塞進同一個。

（快取實作：v1 用 `MemoryCache`；Valkey 實作與失效屬 Plan C 的訂閱寫入路徑，屆時替換此處即可。）

- [ ] **Step 7: 跑測試**

Run: `cd backend && task check`
Expected: 全綠

- [ ] **Step 8: Commit**

```bash
git add backend/internal/services backend/internal/server/domains.go
git commit -m "feat(backend): 權益計數器與四個業務服務的建構子注入（entitlements.Service）"
```

---

### Task 6: 配額守衛掛點與表驅動測試

**Files:**
- Modify: `internal/services/user_service.go`（`CreateUser`）、`customer_service.go`（`CreateCustomer`、`RestoreCustomer`）、`product_service.go`（`CreateProduct`、`RestoreProduct`）、`company_service.go`（`CreateDepartment`）

**⚠️ 守衛清單是「6 項」而非 7 項（controller 更正，2026-09-20；T6 開工前實查）**：spec §4.5 列的第 7 項 `DepartmentService.RestoreDepartment` **在 repo 不存在**——`proto/salesorder/v1/company.proto` 的 `DepartmentService` 只有 `List/Get/Create/Update/Delete`（00020 的部門軟刪除**只做了 Delete 側**），全 repo（含 generated Go）grep 只命中文件。因此本任務只掛 6 個守衛，並在測試檔以**可追蹤的具名缺口註解**記錄第 7 項（部門復原 RPC 落地後必須補上 `{RestoreDepartment, LimitDepartments}` 並掛守衛）。**不得**在本任務新增該 RPC（屬新功能：proto／生成／handler／scope ability）。缺口歸屬：部門域（`backend/2026-08-17-backend-02-tenancy-users-plan.md`），並建議回寫 spec §4.5。
- Create: `internal/services/entitlement_guard_test.go`（表驅動）
- Create: `internal/services/entitlement_guard_integration_test.go`

**Interfaces:**
- Consumes: `entitlements.Service.CheckLimit`
- Produces: 受守衛保護的 7 個 RPC（清單即契約）

- [ ] **Step 1: 寫表驅動失敗測試（`entitlement_guard_test.go`）**

契約：每個寫入 RPC 都必須呼叫對應的 `CheckLimit`；漏掛即紅。以假 service（記錄呼叫）驗證**對應關係**，而非只驗「有呼叫」。

```go
package services

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

// guardCases 為 spec §4.5 的守衛清單：RPC → feature。新增寫入 RPC 未登記即紅。
var guardCases = []struct {
	name    string
	feature string
}{
	{"UserService.CreateUser", entitlements.LimitSeats},
	{"CustomerService.CreateCustomer", entitlements.LimitCustomers},
	{"CustomerService.RestoreCustomer", entitlements.LimitCustomers},
	{"ProductService.CreateProduct", entitlements.LimitProducts},
	{"ProductService.RestoreProduct", entitlements.LimitProducts},
	{"DepartmentService.CreateDepartment", entitlements.LimitDepartments},
	// 待補（第 7 項）：spec §4.5 的 DepartmentService.RestoreDepartment **在 repo 不存在**
	// （DepartmentService 只有 List/Get/Create/Update/Delete；00020 的部門軟刪除只做了 Delete 側）
	// → 部門復原 RPC 落地後，必須在此補上 {"DepartmentService.RestoreDepartment", LimitDepartments}
	//   並在 company_service.go 對應位置掛守衛。歸屬：backend-02-tenancy-users 計畫。
}

// recorder 記錄 CheckLimit 被呼叫的 feature。
type recorder struct {
	calls   []string
	results map[string]error
}

func (r *recorder) CheckLimit(_ context.Context, _ int, feature string, _ int) error {
	r.calls = append(r.calls, feature)
	return r.results[feature]
}

func (r *recorder) Allows(context.Context, int, string) (bool, error) { return true, nil }

func TestGuardMatrixUsesExpectedFeature(t *testing.T) {
	for _, tc := range guardCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &recorder{results: map[string]error{}}
			// 以 sqlite 建立最小資料，呼叫對應 RPC（呼叫方式沿用該域既有測試的建構樣板），
			// 斷言：rec.calls 恰為 []{tc.feature}（多一項或少一項都紅）。
			got := runGuardedRPC(t, rec, tc.name)
			if len(got) != 1 || got[0] != tc.feature {
				t.Fatalf("%s 應恰好檢查一次 %s,got %v", tc.name, tc.feature, got)
			}
		})
	}
}
```

`runGuardedRPC` 為本檔輔助：依 `tc.name` 建立對應服務（注入 `rec` 包裝的 entitlement）＋最小 fixture，呼叫該 RPC，回傳 `rec.calls`。**每個 RPC 一支 case**，實作時逐一補上（不得只做其中幾個）。

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/services/ -run TestGuardMatrix -v`
Expected: FAIL —7 個 case 都因「沒有呼叫 CheckLimit」而紅

- [ ] **Step 3: 掛守衛（7 處）**

以 `CreateUser` 為例（其餘同型，位置在**驗證完成、交易開始之前**）：

```go
	if err := s.ent.CheckLimit(ctx, cid, entitlements.LimitSeats, 1); err != nil {
		return nil, err
	}
```

- `CreateCustomer` / `CreateProduct` / `CreateDepartment`：同上，feature 換對應常數。
- **`Restore*` 特別注意**：復原會增加有效筆數（軟刪除設計下的專屬漏洞），守衛放在「確認該列存在且已刪除」之後、復原寫入之前。

- [ ] **Step 4: 端到端語意測試（sqlite ＋ 假 store，不需容器）**

守衛邏輯與 DB 無關，故用 sqlite ＋ `store.Fake` 就能測到完整語意（上限、釋放、被擋後不落庫）。先在 `user_service_test.go` 補一個可注入 entitlement 的建構輔助：

```go
// newUserTestServerWithEntitlement 與 newUserTestServerWithDB 同構，差別是可注入自訂權益服務
// （測試配額行為用；其他測試沿用 Unlimited()）。
func newUserTestServerWithEntitlement(t *testing.T, id authz.Identity, db *ent.Client,
	entSvc *entitlements.Service) v1connect.UserServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterUserServices(mux, db, entSvc)
	// 其餘與 newUserTestServerWithDB 相同：注入 identity、CASL 與 db 後起 httptest server。
	return startUserTestServer(t, mux, id, db)
}
```

（`startUserTestServer` 由既有 `newUserTestServerWithDB` 抽出共用段落——同一支輔助，不另寫第二套。）

```go
// TestCreateUserBlockedAtSeatLimit：方案席位上限 10、已有 10 席 → CreateUser 回 FailedPrecondition；
// 停用一席（釋放配額）後可再建。以 sqlite ＋ 假 store 覆蓋，不需 testcontainers。
func TestCreateUserBlockedAtSeatLimit(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	for i := range 10 {
		db.User.Create().SetCompanyID(coID).SetEmail(fmt.Sprintf("u%d@t.com", i)).SetName("u").
			SetRole("staff").SetStatus("active").SetPasswordHash("x").SaveX(ctx)
	}

	f := store.NewFake()
	f.PutFeature(store.Feature{Code: entitlements.LimitSeats, Type: "integer"})
	f.PutPlan("std", []store.Entitlement{
		{FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(10))}})
	f.PutSubscription(store.Subscription{CompanyID: coID, PlanCode: "std", Status: "active"})
	entSvc := entitlements.New(f, NewEntitlementCounter(db), entitlements.NewMemoryCache(), 0)

	id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), Role: "company_admin",
		Roles: []string{"company_admin"}}
	client := newUserTestServerWithEntitlement(t, id, db, entSvc)

	_, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "第 11 人", Email: "over@t.com", CompanyId: uItoa(coID), Role: "staff",
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("已達席位上限應回 failed_precondition（引導升級方案），got %v", err)
	}
	// 斷言**碼**而非只有 connect 碼：前端據 ErrorInfo.code 導向升級方案（Plan D T5b）。
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-5001" {
		t.Fatalf("席位超限必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}
	if n, _ := db.User.Query().Where(user.CompanyIDEQ(coID)).Count(ctx); n != 10 {
		t.Fatalf("被擋後不得新增帳號，got %d", n)
	}

	// 停用一席即釋放配額（spec §3.2 規則 1）→ 可再建。
	db.User.Update().Where(user.EmailEQ("u0@t.com")).SetStatus("inactive").SaveX(ctx)
	if _, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "遞補", Email: "refill@t.com", CompanyId: uItoa(coID), Role: "staff",
	})); err != nil {
		t.Fatalf("停用釋放席位後應可建立: %v", err)
	}
}

func ptr[T any](v T) *T { return &v }
```

- [ ] **Step 5: 跑測試**

Run: `cd backend && go test ./internal/services/ -run TestGuardMatrix -v && task test:integration -- -run TestIntegrationSeatGuard -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/user_service.go backend/internal/services/customer_service.go backend/internal/services/product_service.go backend/internal/services/company_service.go backend/internal/services/entitlement_guard_test.go backend/internal/services/entitlement_guard_integration_test.go
git commit -m "feat(backend): 七個寫入 RPC 掛配額守衛（含復原路徑）與表驅動守衛測試"
```

---

### Task 7: `platform/v1` proto 與三端生成

**Files:**
- Create: `backend/proto/platform/v1/platform.proto`
- Generated: `backend/internal/proto/platform/v1/*`、`frontend/src/lib/proto/platform/v1/*`、`app/lib/gen/platform/v1/*`

**Interfaces:**
- Produces: `PlatformAdminService`（operator）、`TenantEntitlementService`（租戶）

- [ ] **Step 1: 寫 proto（`backend/proto/platform/v1/platform.proto`）**

```proto
syntax = "proto3";

package platform.v1;
option go_package = "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1;platformv1";

// 平台域(v1;D34–D39):平台營運工具(operator 身分)與租戶端權益投影。
// 硬規則:platform.* 能力不得出現在租戶 GetAbility(S11);租戶身分不得呼叫 PlatformAdminService。

// TenantSummary:平台視角的租戶概況(跨租戶,僅 operator 可讀)。
message TenantSummary {
  string company_id = 1;
  string company_name = 2;
  string plan_code = 3;
  string plan_name = 4;
  string subscription_status = 5;  // trialing | active | past_due | suspended | cancelled | none
  int32  seat_count = 6;
  string current_period_end = 7;   // RFC3339,可空
  bool   overdue = 8;              // 有未付期別
}

message ListTenantsRequest {
  int32  page = 1;
  int32  page_size = 2;
  string keyword = 3;              // 公司名稱/識別碼模糊
  string status = 4;               // 訂閱狀態篩選,可空
}
message ListTenantsResponse {
  repeated TenantSummary tenants = 1;
  PlatformPagination pagination = 2;
}

message PlatformPagination {
  int32 page = 1;
  int32 page_size = 2;
  int32 total = 3;
}

message GetTenantRequest { string company_id = 1; }
message GetTenantResponse {
  TenantSummary tenant = 1;
  repeated TenantOverride overrides = 2;
}

message TenantOverride {
  string id = 1;
  string feature_code = 2;
  bool   enabled_set = 3;          // 是否指定 enabled
  bool   enabled = 4;
  bool   limit_set = 5;
  int64  limit_value = 6;
  string reason = 7;
  string owner = 8;
  string expires_at = 9;
}

message ListPlansRequest {}
message ListPlansResponse { repeated Plan plans = 1; }
message Plan {
  string id = 1;
  string code = 2;
  string name = 3;
  string status = 4;
  int32  sort_order = 5;
  repeated PlanPrice prices = 6;
}
message PlanPrice {
  string billing_cycle = 1;
  string base_price = 2;
  string seat_price = 3;
  string currency = 4;
  string effective_from = 5;
}

message GetPlanEntitlementsRequest { string plan_code = 1; }
message GetPlanEntitlementsResponse {
  repeated FeatureEntitlement entitlements = 1;
  repeated Feature features = 2;   // 完整功能清單(供矩陣顯示未設定的格)
}
message Feature {
  string code = 1;
  string type = 2;                 // boolean | integer
  string unit = 3;
  string description = 4;
}
message FeatureEntitlement {
  string feature_code = 1;
  bool   enabled = 2;
  bool   limit_set = 3;
  int64  limit_value = 4;
}

message ListPlatformAuditRequest {
  int32  page = 1;
  int32  page_size = 2;
  string target_type = 3;
  string target_id = 4;
}
message ListPlatformAuditResponse {
  repeated PlatformAuditEntry entries = 1;
  PlatformPagination pagination = 2;
}
message PlatformAuditEntry {
  string id = 1;
  string operator_email = 2;
  string action = 3;
  string target_type = 4;
  string target_id = 5;
  string reason = 6;
  string created_at = 7;
}

// PlatformAdminService:平台營運(operator session;租戶身分一律拒絕)。
service PlatformAdminService {
  rpc ListTenants(ListTenantsRequest) returns (ListTenantsResponse);
  rpc GetTenant(GetTenantRequest) returns (GetTenantResponse);
  rpc ListPlans(ListPlansRequest) returns (ListPlansResponse);
  rpc GetPlanEntitlements(GetPlanEntitlementsRequest) returns (GetPlanEntitlementsResponse);
  rpc ListPlatformAudit(ListPlatformAuditRequest) returns (ListPlatformAuditResponse);
}

// TenantEntitlement:租戶端唯讀投影(自己的公司;前端據此 disable 按鈕與顯示用量)。
message GetTenantEntitlementsRequest {}
message GetTenantEntitlementsResponse {
  string plan_code = 1;
  string plan_name = 2;
  string status = 3;
  string trial_ends_at = 4;
  repeated Usage usage = 5;
}
message Usage {
  string feature_code = 1;
  bool   enabled = 2;
  bool   limit_set = 3;
  int64  limit_value = 4;
  int64  used = 5;
}

service TenantEntitlementService {
  rpc GetTenantEntitlements(GetTenantEntitlementsRequest) returns (GetTenantEntitlementsResponse);
}
```

- [ ] **Step 2: 生成三端型別**

```bash
cd backend
# protoc-gen-dart 需以 fvm pin 的 Dart 執行（backend/AGENTS.md §2）
export PATH="$HOME/fvm/versions/stable/bin:$HOME/.pub-cache/bin:$PATH"
task proto:gen
git status --short   # 應出現 backend/internal/proto/platform、frontend/src/lib/proto/platform、app/lib/gen/platform
```

- [ ] **Step 3: 確認生成物可編譯**

Run: `cd backend && go build ./... && cd ../frontend && pnpm typecheck`
Expected: 皆通過

- [ ] **Step 4: Commit**

```bash
git add backend/proto/platform backend/internal/proto/platform frontend/src/lib/proto/platform app/lib/gen/platform
git commit -m "feat(proto): platform/v1（PlatformAdminService 與租戶端權益投影）三端生成"
```

---

### Task 8: 平台操作者認證（OIDC ＋ operator JWT ＋ cookie ＋ interceptor）

**Files:**
- Create: `internal/platform/operatorauth/service.go`、`internal/platform/operatorauth/interceptor.go`
- Create: `internal/platform/operatorauth/service_test.go`
- Modify: `internal/server/domains.go`（掛 `/platform/auth/*` 與 interceptor）、`internal/platform/store/postgres/operators.go`（`Operators` 查詢與寫入）

**Interfaces:**
- Consumes: `store`（新增 `Operators`／`Audit` 方法）、`auth.NewGoogleVerifier`、`auth.NewGoogleOAuthConfig`
- Produces:

```go
type Identity struct {
	OperatorID int64
	Email      string
	Role       string // operator | admin
}

func WithIdentity(ctx context.Context, id Identity) context.Context
func IdentityFrom(ctx context.Context) (Identity, bool)

type Service struct{ /* cfg, store, issuer, verifier, oauth */ }

// Login 導向 Google OIDC（state 存 cookie）。
func (s *Service) Login(w http.ResponseWriter, r *http.Request)
// Callback 驗 id_token → 檢查網域與白名單 → 簽 operator JWT → 設 cookie → 導回 ConsoleURL。
func (s *Service) Callback(w http.ResponseWriter, r *http.Request)
// Interceptor 驗 operator cookie；非 operator 一律 Unauthenticated。
func (s *Service) Interceptor() connect.Interceptor
```

- [ ] **Step 1: 寫失敗測試（`service_test.go`）—— 三條安全契約**

```go
package operatorauth_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
)

const testSecret = "platform-secret"

// fakeStore 以記憶體提供白名單查詢（三個安全契約都不需要真 DB）。
type fakeStore struct{ op *operatorauth.Operator }

func (f *fakeStore) OperatorByEmail(context.Context, string) (*operatorauth.Operator, error) {
	return f.op, nil
}
func (f *fakeStore) TouchOperatorLogin(context.Context, int64, time.Time) error { return nil }
func (f *fakeStore) AuditOperatorLogin(context.Context, int64, string, string, string) error {
	return nil
}

func activeStore() *fakeStore {
	return &fakeStore{op: &operatorauth.Operator{
		ID: 1, Email: "ops@example.com", Name: "Ops", Role: "admin", Status: "active"}}
}

func newTestService(t *testing.T, st *fakeStore) *operatorauth.Service {
	t.Helper()
	return operatorauth.New(operatorauth.Config{Secret: testSecret}, st)
}

// callWithCookie 以指定 cookie 值走 interceptor；handler 被呼叫即視為失敗。
func callWithCookie(t *testing.T, svc *operatorauth.Service, value string) error {
	t.Helper()
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Cookie", operatorauth.CookieName+"="+value)
	_, err := svc.Interceptor().WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("未通過驗證的請求不得進入 handler")
		return nil, nil
	})(context.Background(), req)
	return err
}

// ① 租戶 secret 簽出的 token 不得通過平台 interceptor（跨用等於可冒充平台操作者）。
func TestTenantTokenRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token := signHS256(t, "tenant-secret", jwt.MapClaims{"sub": 1, "aud": "platform"})
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("租戶 secret 簽出的 token 必須 Unauthenticated，got %v", err)
	}
}

// ② 即使 secret 相同，audience 不符（租戶 token 常無 aud=platform）也拒絕。
func TestWrongAudienceRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token := signHS256(t, testSecret, jwt.MapClaims{"sub": 1, "aud": "tenant"})
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("audience 不符必須 Unauthenticated，got %v", err)
	}
}

// ③ secret 與 audience 都對，但 operators.status = disabled → 立即失效（不需黑名單）。
func TestDisabledOperatorRejected(t *testing.T) {
	st := activeStore()
	st.op.Status = "disabled"
	svc := newTestService(t, st)
	token, err := svc.IssueToken(operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("已停用的 operator 必須 Unauthenticated，got %v", err)
	}
}

// ④ 正向：有效的 operator token 可以通過（否則前三條可能因為「什麼都拒絕」而假綠）。
func TestActiveOperatorAccepted(t *testing.T) {
	svc := newTestService(t, activeStore())
	token, err := svc.IssueToken(operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Cookie", operatorauth.CookieName+"="+token)
	reached := false
	if _, err := svc.Interceptor().WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		reached = true
		if _, ok := operatorauth.IdentityFrom(ctx); !ok {
			t.Error("handler 內必須取得 operator 身分")
		}
		return nil, nil
	})(context.Background(), req); err != nil {
		t.Fatalf("有效 token 應通過: %v", err)
	}
	if !reached {
		t.Fatal("handler 未被呼叫")
	}
}

func signHS256(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	return s
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/platform/operatorauth/ -v`
Expected: FAIL —`undefined: operatorauth.CookieName`／`newTestService`

- [ ] **Step 3: 實作（`service.go`）**

```go
// Package operatorauth 為平台工具的認證(D38/S8):OIDC 限公司網域 + platform.operators 白名單,
// 簽發獨立 secret 的 operator JWT 置於 HttpOnly cookie。租戶 session/JWT 一律不適用。
package operatorauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

// CookieName 為 operator session cookie;Path 限定 /platform,與租戶 session 不重疊。
const CookieName = "platform_session"

// Audience 為 operator token 的固定 audience;租戶 token 不得帶此值,反之亦然。
const Audience = "platform"

type Identity struct {
	OperatorID int64
	Email      string
	Role       string
}

type ctxKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// Config 為 operatorauth 的設定(由 config.Platform 對應)。
type Config struct {
	Secret         string
	CookieDomain   string
	ConsoleURL     string
	AllowedDomain  string
	TokenLifetime  time.Duration // 預設 12h
}

// Service 提供登入端點與 RPC interceptor。
type Service struct {
	cfg   Config
	store Store
}

// Store 為 operatorauth 需要的資料存取(由 platform store 的 postgres 實作提供)。
type Store interface {
	OperatorByEmail(ctx context.Context, email string) (*Operator, error)
	TouchOperatorLogin(ctx context.Context, id int64, at time.Time) error
	AuditOperatorLogin(ctx context.Context, operatorID int64, email, ip, ua string) error
}

type Operator struct {
	ID     int64
	Email  string
	Name   string
	Role   string
	Status string
}

func New(cfg Config, st Store) *Service {
	if cfg.TokenLifetime == 0 {
		cfg.TokenLifetime = 12 * time.Hour
	}
	return &Service{cfg: cfg, store: st}
}

// IssueToken 簽發 operator JWT（獨立 secret ＋ 固定 audience）。
func (s *Service) IssueToken(id Identity) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  id.OperatorID,
		"role": id.Role,
		"email": id.Email,
		"aud":  Audience,
		"iat":  now.Unix(),
		"exp":  now.Add(s.cfg.TokenLifetime).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
}

// verify 驗 token：secret、audience、exp 與 operators.status 逐一檢查。
func (s *Service) verify(ctx context.Context, raw string) (Identity, error) {
	parsed, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非預期的簽章演算法")
		}
		return []byte(s.cfg.Secret), nil
	}, jwt.WithAudience(Audience), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return Identity{}, errors.New("operator token 無效")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, errors.New("operator token 內容無效")
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return Identity{}, errors.New("operator token 缺 sub")
	}
	email, _ := claims["email"].(string)
	// 以 DB 白名單為準：停用立即失效（無需黑名單）。
	op, err := s.store.OperatorByEmail(ctx, email)
	if err != nil {
		return Identity{}, err
	}
	if op == nil || op.Status != "active" || !strings.EqualFold(op.Email, email) {
		return Identity{}, errors.New("operator 不存在或已停用")
	}
	_ = sub
	return Identity{OperatorID: op.ID, Email: op.Email, Role: op.Role}, nil
}

// Interceptor 驗 operator cookie;失敗一律 Unauthenticated。
func (s *Service) Interceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			cookie, err := cookieFrom(req)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入平台工具"))
			}
			id, err := s.verify(ctx, cookie)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			return next(WithIdentity(ctx, id), req)
		}
	})
}

// cookieFrom 由請求標頭解析 operator cookie。
func cookieFrom(req connect.AnyRequest) (string, error) {
	raw := req.Header().Get("Cookie")
	if raw == "" {
		return "", errors.New("缺少 cookie")
	}
	h := http.Header{"Cookie": []string{raw}}
	c, err := (&http.Request{Header: h}).Cookie(CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// CookiePath 為 operator cookie 的 Path:限定平台 API 與登入端點。
const CookiePath = "/platform"

// oidcStateCookie 為 OIDC CSRF state 的短期 cookie。
const oidcStateCookie = "platform_oidc_state"
```

- [ ] **Step 4: 補 `Login`／`Callback`（同檔；重用既有 OIDC 封裝 `internal/auth/oidc.go`）**

`Service` 追加三個欄位與建構參數（與既有 `handlers/auth_handler.go` 用同一組封裝，故可用 fake 測試）：

```go
type Service struct {
	cfg       Config
	store     Store
	oauth     *oauth2.Config          // auth.NewGoogleOAuthConfig(clientID, secret, redirectURL)
	exchanger auth.OAuthExchanger     // auth.NewGoogleOAuthExchanger(oauth)
	verifier  auth.OIDCVerifier       // auth.NewGoogleVerifier(ctx, clientID)
}

func New(cfg Config, st Store, oauth *oauth2.Config, ex auth.OAuthExchanger, v auth.OIDCVerifier) *Service {
	if cfg.TokenLifetime == 0 {
		cfg.TokenLifetime = 12 * time.Hour
	}
	return &Service{cfg: cfg, store: st, oauth: oauth, exchanger: ex, verifier: v}
}

// Login 導向 Google OIDC:state 存短期 cookie,完成後比對(防 CSRF)。
func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	state, err := auth.NewState()
	if err != nil {
		http.Error(w, "無法產生 OIDC state", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: oidcStateCookie, Value: state, Path: CookiePath,
		Domain: s.cfg.CookieDomain, HttpOnly: true, Secure: true,
		SameSite: http.SameSiteLaxMode, MaxAge: int(auth.StateTTL.Seconds()),
	})
	http.Redirect(w, r, s.oauth.AuthCodeURL(state), http.StatusFound)
}

// Callback 完成 OIDC:比對 state → 換 id_token → 驗網域 → 查白名單 → 簽 token → 設 cookie → 導回 console。
// 每一步失敗都不得簽發 token;錯誤訊息不含 token 內容。
func (s *Service) Callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	stateCookie, err := r.Cookie(oidcStateCookie)
	if err != nil || stateCookie.Value == "" || q.Get("state") != stateCookie.Value {
		http.Error(w, "OIDC state 不符", http.StatusBadRequest)
		return
	}
	// state 為一次性:立即清除。
	http.SetCookie(w, &http.Cookie{
		Name: oidcStateCookie, Value: "", Path: CookiePath,
		Domain: s.cfg.CookieDomain, HttpOnly: true, Secure: true, MaxAge: -1,
	})

	rawIDToken, err := s.exchanger.Exchange(r.Context(), q.Get("code"))
	if err != nil {
		http.Error(w, "授權碼交換失敗", http.StatusUnauthorized)
		return
	}
	ident, err := s.verifier.VerifyIDToken(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "ID token 驗證失敗", http.StatusUnauthorized)
		return
	}

	// 網域限制:只接受公司 Workspace 網域(不分大小寫;含 @ 之後的部分)。
	at := strings.LastIndex(ident.Email, "@")
	if at < 0 || !strings.EqualFold(ident.Email[at+1:], s.cfg.AllowedDomain) {
		http.Error(w, "此帳號不屬於允許的網域", http.StatusForbidden)
		return
	}

	op, err := s.store.OperatorByEmail(r.Context(), ident.Email)
	if err != nil {
		http.Error(w, "查詢操作者失敗", http.StatusInternalServerError)
		return
	}
	if op == nil || op.Status != "active" {
		http.Error(w, "此帳號不在平台操作者名單中", http.StatusForbidden)
		return
	}

	token, err := s.IssueToken(Identity{OperatorID: op.ID, Email: op.Email, Role: op.Role})
	if err != nil {
		http.Error(w, "簽發 token 失敗", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: token, Path: CookiePath,
		Domain: s.cfg.CookieDomain, HttpOnly: true, Secure: true,
		SameSite: http.SameSiteLaxMode, MaxAge: int(s.cfg.TokenLifetime.Seconds()),
	})

	// 登入本身也是稽核對象:成功登入要留痕(誰、何時、來自哪裡)。
	if err := s.store.TouchOperatorLogin(r.Context(), op.ID, time.Now()); err != nil {
		http.Error(w, "更新登入時間失敗", http.StatusInternalServerError)
		return
	}
	if err := s.store.AuditOperatorLogin(r.Context(), op.ID, op.Email,
		clientIP(r), r.UserAgent()); err != nil {
		http.Error(w, "寫入稽核失敗", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, s.cfg.ConsoleURL, http.StatusFound)
}

// clientIP 取真實來源 IP(X-Forwarded-For 首項,否則 RemoteAddr)。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
```

（`service.go` 追加 import：`net`、`net/url` 不需、`strings`、`golang.org/x/oauth2`、`github.com/salesorder/sales-order-1.0/backend/internal/auth`。`Store` 介面需含 `TouchOperatorLogin` 與 `AuditOperatorLogin`，兩者由 `internal/platform/store/postgres` 實作——`AuditOperatorLogin` 寫 `platform.audit_logs`，`action='login'`、`target_type='operator'`、`target_id=<operatorID>`。）

- [ ] **Step 5: 掛載（`internal/server/domains.go`）**

```go
	// 平台工具認證:僅在設定齊備時掛載(開發環境可不設)。
	if s.cfg.Platform.Configured() {
		opAuth := operatorauth.New(operatorauth.Config{
			Secret:        s.cfg.Platform.OperatorJWTSecret,
			CookieDomain:  s.cfg.Platform.CookieDomain,
			ConsoleURL:    s.cfg.Platform.ConsoleURL,
			AllowedDomain: s.cfg.Platform.AllowedEmailDomain,
		}, platformStore /* 實作 operatorauth.Store */)
		s.operatorAuth = opAuth
		s.router.Get("/platform/auth/google", opAuth.Login)
		s.router.Get("/platform/auth/google/callback", opAuth.Callback)
	}
```

- [ ] **Step 6: 跑測試**

Run: `cd backend && go test ./internal/platform/operatorauth/ -v`
Expected: 三條契約測試 PASS（`TestDisabledOperatorRejected` 的 Skip 必須移除）

- [ ] **Step 7: Commit**

```bash
git add backend/internal/platform/operatorauth backend/internal/platform/store/postgres backend/internal/server/domains.go
git commit -m "feat(backend): 平台操作者認證（OIDC 限網域＋白名單、獨立 secret 與 cookie、interceptor）"
```

---

### Task 9: `PlatformAdminService` 唯讀 RPC 與平台稽核

**Files:**
- Create: `internal/services/platform_admin_service.go`
- Create: `internal/services/platform_admin_integration_test.go`
- Modify: `internal/platform/store/postgres/*.go`（租戶列表／方案／稽核查詢）

**Interfaces:**
- Consumes: `operatorauth.IdentityFrom`、`store`
- Produces: `ListTenants`／`GetTenant`／`ListPlans`／`GetPlanEntitlements`／`ListPlatformAudit`，全部要求 operator 身分

- [ ] **Step 1: 寫授權邊界測試（`platform_admin_service_test.go`，假 store、不需容器）**

契約有兩層：**interceptor 擋一次、服務層再擋一次**（服務可能被其他路徑重用，不能只靠 interceptor）。

```go
package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
)

// TestPlatformAdminRequiresOperatorIdentity 驗證服務層的 operator 檢查（不依賴 interceptor）。
func TestPlatformAdminRequiresOperatorIdentity(t *testing.T) {
	svc := NewPlatformAdminService(&fakePlatformStore{
		tenants: []TenantRow{{CompanyID: "7", CompanyName: "測試公司", PlanCode: "std", Status: "active"}},
	})

	if _, err := svc.ListTenants(context.Background(),
		connect.NewRequest(&platformv1.ListTenantsRequest{Page: 1, PageSize: 20}),
	); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("無 operator 身分必須 Unauthenticated，got %v", err)
	}

	ctx := operatorauth.WithIdentity(context.Background(),
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	resp, err := svc.ListTenants(ctx,
		connect.NewRequest(&platformv1.ListTenantsRequest{Page: 1, PageSize: 20}))
	if err != nil {
		t.Fatalf("operator 身分應可讀取: %v", err)
	}
	if got := resp.Msg.GetPagination().GetTotal(); got != 1 {
		t.Fatalf("fixture 有 1 家租戶，got %d", got)
	}
	if got := resp.Msg.GetTenants()[0].GetCompanyName(); got != "測試公司" {
		t.Fatalf("租戶名稱映射錯誤: %q", got)
	}
}

// fakePlatformStore 只實作本測試需要的查詢；其餘回傳零值（介面契約由編譯器保證）。
type fakePlatformStore struct {
	tenants []TenantRow
}

func (f *fakePlatformStore) ListTenants(context.Context, string, string, int32, int32) ([]TenantRow, int, error) {
	return f.tenants, len(f.tenants), nil
}

func (f *fakePlatformStore) GetTenant(context.Context, string) (*TenantRow, error) { return nil, nil }
func (f *fakePlatformStore) ListPlans(context.Context) ([]PlanRow, error)          { return nil, nil }
func (f *fakePlatformStore) PlanEntitlements(context.Context, string) ([]FeatureEntitlementRow, []FeatureRow, error) {
	return nil, nil, nil
}
func (f *fakePlatformStore) ListPlatformAudit(context.Context, string, string, int32, int32) ([]PlatformAuditRow, int, error) {
	return nil, 0, nil
}
func (f *fakePlatformStore) RecordAudit(context.Context, int64, string, string, string, string, map[string]any, map[string]any) error {
	return nil
}
```

- [ ] **Step 2: 跑測試確認失敗（紅在「服務尚未存在」）**

Run: `cd backend && go build ./... 2>&1 | head`
Expected: FAIL —`undefined: RegisterPlatformAdminService`

- [ ] **Step 3: 實作服務（`platform_admin_service.go`）**

```go
// PlatformAdminService 為平台工具的唯一後端入口(D38):
// 跨租戶視圖一律走本服務與 admin 連線,不得以 scope=all 掃業務表(spec §6.4)。
package services

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
)

type PlatformAdminService struct {
	st platformStore // 介面:ListTenants/GetTenant/ListPlans/PlanEntitlements/ListPlatformAudit
	platformv1connect.UnimplementedPlatformAdminServiceHandler
}

func NewPlatformAdminService(st platformStore) *PlatformAdminService {
	return &PlatformAdminService{st: st}
}

// RegisterPlatformAdminService 掛載平台工具 RPC;呼叫端負責加 operator interceptor。
func RegisterPlatformAdminService(mux *http.ServeMux, st platformStore, op *operatorauth.Service) {
	path, handler := platformv1connect.NewPlatformAdminServiceHandler(
		NewPlatformAdminService(st), connect.WithInterceptors(op.Interceptor()))
	mux.Handle(path, handler)
}

func (s *PlatformAdminService) ListTenants(ctx context.Context, req *connect.Request[platformv1.ListTenantsRequest]) (*connect.Response[platformv1.ListTenantsResponse], error) {
	if _, ok := operatorauth.IdentityFrom(ctx); !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errNoOperator)
	}
	rows, total, err := s.st.ListTenants(ctx, req.Msg.GetKeyword(), req.Msg.GetStatus(), req.Msg.GetPage(), req.Msg.GetPageSize())
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*platformv1.TenantSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, tenantToProto(r))
	}
	return connect.NewResponse(&platformv1.ListTenantsResponse{
		Tenants: out,
		Pagination: &platformv1.PlatformPagination{
			Page: req.Msg.GetPage(), PageSize: req.Msg.GetPageSize(), Total: int32(total),
		},
	}), nil
}
```

同檔續作 `GetTenant`、`ListPlans`、`GetPlanEntitlements`、`ListPlatformAudit`，每個方法第一行都做同一個 operator 檢查（**不得只在 interceptor 檢查**：服務可被其他路徑重用，兩層都要）。

- [ ] **Step 4: 平台稽核寫入（同檔）**

```go
// recordPlatformAudit 寫入 platform.audit_logs(S9:actor 為 operator,不 FK 租戶 users)。
// 與業務稽核(audit.Record)不同表、不同交易語意:平台操作本身即為稽核對象。
func (s *PlatformAdminService) recordPlatformAudit(ctx context.Context, id operatorauth.Identity,
	action, targetType, targetID, reason string, before, after map[string]any) error {
	return s.st.RecordAudit(ctx, id.OperatorID, action, targetType, targetID, reason, before, after)
}
```

（v1 只有讀取 RPC，`recordPlatformAudit` 由 Task 9 的「新增/停用 operator」與 Plan C 的寫入路徑呼叫；本任務先提供並以單元測試覆蓋其參數映射。）

- [ ] **Step 5: 跑測試**

Run: `cd backend && task check && task test:integration -- -run TestIntegrationPlatformAdmin -v`
Expected: PASS（Skip 已移除）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/platform_admin_service.go backend/internal/services/platform_admin_integration_test.go backend/internal/platform/store/postgres
git commit -m "feat(backend): PlatformAdminService 唯讀 RPC 與平台稽核寫入（operator 雙層檢查）"
```

---

### Task 10: 租戶端權益投影（`TenantEntitlementService`）

**Files:**
- Create: `internal/services/entitlement_service.go`
- Create: `internal/services/entitlement_service_integration_test.go`

**Interfaces:**
- Consumes: `entitlements.Service.Snapshot`、`requireAuth`
- Produces: `GetTenantEntitlements`（只回自己公司）

- [ ] **Step 1: 寫契約測試（`entitlement_service_test.go`，假 store、不需容器）**

三條契約：未登入 → `Unauthenticated`；已登入 → 只回**自己公司**（本 RPC 無 `company_id` 參數即為契約）；用量數字與計數器一致。

```go
package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
)

// seatCounter 回傳固定用量（此處模擬「公司 42 已有 3 席」）。
type seatCounter struct{ used int }

func (c seatCounter) Count(context.Context, int, string) (int, error) { return c.used, nil }

func TestTenantEntitlementsScopedToOwnCompany(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(store.Feature{Code: entitlements.LimitSeats, Type: "integer", Unit: "席"})
	f.PutFeature(store.Feature{Code: entitlements.FeaturePrinting, Type: "boolean"})
	f.PutPlan("std", []store.Entitlement{
		{FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(10))},
		{FeatureCode: entitlements.FeaturePrinting, Enabled: true},
	})
	f.PutSubscription(store.Subscription{CompanyID: 42, PlanCode: "std", Status: "active"})
	svc := entitlements.New(f, seatCounter{used: 3}, entitlements.NewMemoryCache(), 0)
	h := NewTenantEntitlementService(svc)

	// ① 未登入
	if _, err := h.GetTenantEntitlements(context.Background(),
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}),
	); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("未登入必須 Unauthenticated，got %v", err)
	}

	// ② 已登入（身分即公司 42；無參數可指定他公司）
	ctx := authz.WithIdentity(context.Background(), authz.Identity{
		UserID: "1", CompanyID: "42", Role: "company_admin", Roles: []string{"company_admin"},
	})
	resp, err := h.GetTenantEntitlements(ctx,
		connect.NewRequest(&platformv1.GetTenantEntitlementsRequest{}))
	if err != nil {
		t.Fatalf("已登入應可讀取: %v", err)
	}
	if resp.Msg.GetPlanCode() != "std" {
		t.Fatalf("方案應為 std，got %q", resp.Msg.GetPlanCode())
	}

	// ③ 用量與計數器一致；boolean 不給用量
	var seats, printing *platformv1.Usage
	for _, u := range resp.Msg.GetUsage() {
		switch u.GetFeatureCode() {
		case entitlements.LimitSeats:
			seats = u
		case entitlements.FeaturePrinting:
			printing = u
		}
	}
	if seats == nil || printing == nil {
		t.Fatalf("usage 應含 limit.seats 與 feature.printing，got %d 項", len(resp.Msg.GetUsage()))
	}
	if seats.GetUsed() != 3 || seats.GetLimitValue() != 10 {
		t.Fatalf("席位用量應為 3/10，got %d/%d", seats.GetUsed(), seats.GetLimitValue())
	}
	if !printing.GetEnabled() || printing.GetLimitSet() {
		t.Fatalf("boolean 功能應 enabled 且不帶 limit，got enabled=%v limit_set=%v",
			printing.GetEnabled(), printing.GetLimitSet())
	}
}
```

（`ptr` 已於 Task 6 的同套件測試定義，此處直接沿用。）

- [ ] **Step 2: 實作（`entitlement_service.go`）**

```go
// TenantEntitlementService 為租戶端唯讀投影(spec §4.6):前端據此 disable 按鈕與顯示用量。
// **前端 disable 不構成授權**:真正的擋在服務層守衛與 RLS。
package services

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1/platformv1connect"
)

type TenantEntitlementService struct {
	ent *entitlements.Service
	platformv1connect.UnimplementedTenantEntitlementServiceHandler
}

func NewTenantEntitlementService(ent *entitlements.Service) *TenantEntitlementService {
	return &TenantEntitlementService{ent: ent}
}

func RegisterTenantEntitlementService(mux *http.ServeMux, ent *entitlements.Service) {
	path, handler := platformv1connect.NewTenantEntitlementServiceHandler(
		NewTenantEntitlementService(ent))
	mux.Handle(path, handler)
}

// GetTenantEntitlements 回傳呼叫者所屬公司的方案、試用到期與用量;無參數可指定他公司(契約)。
func (s *TenantEntitlementService) GetTenantEntitlements(ctx context.Context, _ *connect.Request[platformv1.GetTenantEntitlementsRequest]) (*connect.Response[platformv1.GetTenantEntitlementsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, err := parseID(id.CompanyID)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	snap, err := s.ent.Snapshot(ctx, cid)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(snapshotToProto(snap)), nil
}
```

`Snapshot` 於 `entitlements.Service` 補上（同一個任務內）：回傳方案 code/name/status/trial_ends 與每個已知 feature 的 `enabled`／`limit`／`used`（used 由 Counter 取得；boolean 不給 used）。

- [ ] **Step 3: 掛載並跑測試**

```go
		services.RegisterTenantEntitlementService(apiMux, entSvc)
```

Run: `cd backend && task check && task test:integration -- -run TestIntegrationTenantEntitlements -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/entitlement_service.go backend/internal/services/entitlement_service_integration_test.go backend/internal/server/domains.go backend/internal/platform/entitlements
git commit -m "feat(backend): 租戶端權益投影 RPC（唯讀、只回自己公司）"
```

---

### Task 11: Seeds（features／方案／價目／權益／首位 operator）

**Files:**
- Create: `cmd/seed/platform.go`
- Modify: `cmd/seed/main.go`（呼叫新 seeder）、`.env.example`

**Interfaces:**
- Produces: `SeedPlatform(ctx, db *sql.DB, operatorEmail string) error`（冪等）

- [ ] **Step 1: 寫失敗測試（`cmd/seed/platform_test.go`，整合測試）**

```go
//go:build integration

package main

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/pgx"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationSeedPlatformIdempotent 驗證:連跑兩次結果相同(筆數不變),且
// 8 個 features、3 個方案、每方案每月/每年價目、首位 operator 都在。
func TestIntegrationSeedPlatformIdempotent(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer db.Close()
	if err := goose.RunContext(t.Context(), "up", db, "../../database/migrations"); err != nil {
		t.Fatalf("goose up: %v", err)
	}
	for i := range 2 {
		if err := SeedPlatform(t.Context(), db, "ops@example.com"); err != nil {
			t.Fatalf("第 %d 次 seed: %v", i+1, err)
		}
	}
	var features, plans, prices, operators int
	for _, q := range []struct {
		sql string
		dst *int
	}{
		{`SELECT count(*) FROM platform.features`, &features},
		{`SELECT count(*) FROM platform.plans`, &plans},
		{`SELECT count(*) FROM platform.plan_prices`, &prices},
		{`SELECT count(*) FROM platform.operators`, &operators},
	} {
		if err := db.QueryRow(q.sql).Scan(q.dst); err != nil {
			t.Fatalf("%s: %v", q.sql, err)
		}
	}
	if features != 8 || plans != 3 || prices != 6 || operators != 1 {
		t.Fatalf("seed 結果不符（features=%d plans=%d prices=%d operators=%d）", features, plans, prices, operators)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationSeedPlatformIdempotent -v`
Expected: FAIL —`undefined: SeedPlatform`

- [ ] **Step 3: 實作（`cmd/seed/platform.go`）**

```go
// 平台域 seeder(D34):8 個 features、3 個方案(免費/標準/專業)與價目、方案權益、首位 operator。
// 冪等:一律以 ON CONFLICT DO UPDATE 對齊,不刪除既有資料。
package main

import (
	"context"
	"database/sql"
	"fmt"
)

// features 為 v1 定案清單(spec §4.5):5 個數值上限 + 3 個 boolean 功能。
var platformFeatures = []struct {
	Code, Type, Unit, Desc string
}{
	{"limit.seats", "integer", "席", "帳號席位上線"},
	{"limit.customers", "integer", "客戶", "客戶筆數上線"},
	{"limit.products", "integer", "商品", "商品筆數上線"},
	{"limit.departments", "integer", "部門", "部門數上線"},
	{"limit.storage_gb", "integer", "GB", "檔案空間上線"},
	{"feature.printing", "boolean", "", "單據列印"},
	{"feature.dispatch", "boolean", "", "派車看板"},
	{"feature.returns", "boolean", "", "退貨申請與審核"},
}

// platformPlans 為 v1 起始方案;價格為預設值,實際調價由平台工具維護。
var platformPlans = []struct {
	Code, Name string
	SortOrder  int
	BasePrice  string // 月費
	SeatPrice  string // 每席月費
	Entitle    map[string]int64 // feature → limit(-1 表示 enabled 但不限)
}{
	{"free", "免費", 1, "0", "0", map[string]int64{
		"limit.seats": 3, "limit.customers": 50, "limit.products": 100,
		"limit.departments": 1, "limit.storage_gb": 1,
	}},
	{"std", "標準", 2, "1500", "150", map[string]int64{
		"limit.seats": 10, "limit.customers": 500, "limit.products": 2000,
		"limit.departments": 5, "limit.storage_gb": 20,
		"feature.printing": -1,
	}},
	{"pro", "專業", 3, "4500", "150", map[string]int64{
		"limit.seats": 50, "limit.customers": -1, "limit.products": -1,
		"limit.departments": 20, "limit.storage_gb": 200,
		"feature.printing": -1, "feature.dispatch": -1, "feature.returns": -1,
	}},
}

// SeedPlatform 冪等建立平台基礎資料。**設定一律由 config.Platform 帶入（env 驅動）**，
// 使「上線前改預設值」不必改程式碼：
//   - SeedOperatorEmail／SeedSystemActorEmail：首位操作者與系統 actor（G5）
//   - DefaultTrialDays／DefaultGraceDays／DefaultLeadDays：寫入 platform.settings（可再於營運工具調整）
//   - SeedPrice{Free,Std,Pro}{Base,Seat}：方案價目的首次預設值（佔位數字，上線前務必改）
func SeedPlatform(ctx context.Context, db *sql.DB, cfg config.Platform) error {
	for _, f := range platformFeatures {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.features (code, type, unit, description)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (code) DO UPDATE SET type = EXCLUDED.type, unit = EXCLUDED.unit,
				description = EXCLUDED.description`, f.Code, f.Type, f.Unit, f.Desc); err != nil {
			return fmt.Errorf("seed feature %s: %w", f.Code, err)
		}
	}

	for _, p := range platformPlans {
		var planID int64
		if err := db.QueryRowContext(ctx, `
			INSERT INTO platform.plans (code, name, sort_order)
			VALUES ($1,$2,$3)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, sort_order = EXCLUDED.sort_order
			RETURNING id`, p.Code, p.Name, p.SortOrder).Scan(&planID); err != nil {
			return fmt.Errorf("seed plan %s: %w", p.Code, err)
		}
		for _, cycle := range []string{"monthly", "yearly"} {
			base, seat := p.BasePrice, p.SeatPrice
			// 價目預設值由 env 覆寫（SEED_PRICE_*）。下方 INSERT 帶 NOT EXISTS，
			// 因此重跑 seed **不會覆寫營運已調整的價目**（只補缺的週期）。
			switch p.Code {
			case "free":
				base, seat = cfg.SeedPriceFreeBase, cfg.SeedPriceFreeSeat
			case "std":
				base, seat = cfg.SeedPriceStdBase, cfg.SeedPriceStdSeat
			case "pro":
				base, seat = cfg.SeedPriceProBase, cfg.SeedPriceProSeat
			}
			if cycle == "yearly" { // 年繳 = 月費 × 12 × 0.9（取整到元）
				base = fmt.Sprintf("%.0f", mustFloat(base)*12*0.9)
				seat = fmt.Sprintf("%.0f", mustFloat(seat)*12*0.9)
			}
			if _, err := db.ExecContext(ctx, `
				INSERT INTO platform.plan_prices (plan_id, billing_cycle, base_price, seat_price, currency)
				SELECT $1,$2,$3::numeric,$4::numeric,'TWD'
				 WHERE NOT EXISTS (
					SELECT 1 FROM platform.plan_prices WHERE plan_id = $1 AND billing_cycle = $2
				 )`, planID, cycle, base, seat); err != nil {
				return fmt.Errorf("seed price %s/%s: %w", p.Code, cycle, err)
			}
		}
		for code, limit := range p.Entitle {
			enabled := limit != 0
			var limitValue any
			if limit > 0 {
				limitValue = limit
			}
			if _, err := db.ExecContext(ctx, `
				INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value)
				VALUES ($1,$2,$3,$4)
				ON CONFLICT (plan_id, feature_code)
				DO UPDATE SET enabled = EXCLUDED.enabled, limit_value = EXCLUDED.limit_value`,
				planID, code, enabled, limitValue); err != nil {
				return fmt.Errorf("seed entitlement %s/%s: %w", p.Code, code, err)
			}
		}
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.operators (email, name, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (email) DO NOTHING`, cfg.SeedOperatorEmail, cfg.SeedOperatorName); err != nil {
		return fmt.Errorf("seed operator: %w", err)
	}

	// G5：系統 actor。排程改動租戶狀態時，租戶稽核（audit_logs）需要真實且存在的
	// user_id 與 company_id（兩者都有 FK），故建立「平台自營公司 ＋ 系統使用者」，
	// 並把其 user id 記入 platform.settings 供 cmd/platform-cron 與 consumer 取用。
	//
	// 為何用 WHERE NOT EXISTS 而非 ON CONFLICT：companies.identifier 與 users.email 的
	// 唯一性由「部分唯一索引」（WHERE deleted_at IS NULL）表達，部分索引不能當衝突目標。
	//
	// systemActorEmail 由呼叫端提供（建議 system@<AllowedEmailDomain>）；password_hash
	// 填 '!' 是刻意的不可登入值 —— 此帳號只作為系統 actor，不得有任何人以它登入。
	var platformCompanyID int64
	if err := db.QueryRowContext(ctx, `
		WITH ins AS (
			INSERT INTO companies (name, identifier, status, created_at, updated_at)
			SELECT '平台營運', 'platform', 'active', now(), now()
			 WHERE NOT EXISTS (
				SELECT 1 FROM companies WHERE identifier = 'platform' AND deleted_at IS NULL
			 )
			RETURNING id
		)
		SELECT id FROM ins
		UNION ALL
		SELECT id FROM companies WHERE identifier = 'platform' AND deleted_at IS NULL
		LIMIT 1`).Scan(&platformCompanyID); err != nil {
		return fmt.Errorf("seed 平台自營公司: %w", err)
	}

	var systemUserID int64
	if err := db.QueryRowContext(ctx, `
		WITH ins AS (
			INSERT INTO users (email, name, status, role, password_hash, company_users, created_at, updated_at)
			SELECT $1, '系統排程', 'active', 'super', '!', $2, now(), now()
			 WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)
			RETURNING id
		)
		SELECT id FROM ins
		UNION ALL
		SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1`, cfg.SeedSystemActorEmail, platformCompanyID).Scan(&systemUserID); err != nil {
		return fmt.Errorf("seed 系統使用者: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES ('system_actor_user_id', $1)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		strconv.FormatInt(systemUserID, 10)); err != nil {
		return fmt.Errorf("seed system_actor_user_id: %w", err)
	}

	// 營運參數（試用／寬限／提前天數）：**只補缺、不覆寫** —— 營運調過的值不能被重跑 seed 蓋回去。
	// 之後由 UpdateBillingSettings（Plan C Task 9）與 console 維護。
	for _, kv := range []struct {
		key   string
		value int
	}{
		{"trial_days", cfg.DefaultTrialDays},
		{"grace_days", cfg.DefaultGraceDays},
		{"lead_days", cfg.DefaultLeadDays},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.settings (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO NOTHING`, kv.key, strconv.Itoa(kv.value)); err != nil {
			return fmt.Errorf("seed setting %s: %w", kv.key, err)
		}
	}
	return nil
}

func mustFloat(s string) float64 {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0
	}
	return f
}
```

**seeds 的預設值由 env 提供、UI 可再變更**（G5 的系統 actor 一併在此定案）：

| env | 預設 | 用途 |
|---|---|---|
| `PLATFORM_ALLOWED_EMAIL_DOMAIN` | `sowinsoft.com` | 平台工具登入網域（email 網域比對，相容 Workspace 帳號與已驗證的 Google 別名） |
| `PLATFORM_SEED_OPERATOR_EMAIL` | `ssd@sowinsoft.com` | 首位平台操作者 |
| `PLATFORM_SEED_OPERATOR_NAME` | `ssd` | 顯示名稱 |
| `PLATFORM_SEED_SYSTEM_ACTOR_EMAIL` | `system@sowinsoft.com` | 系統排程 actor（`password_hash='!'`，不可登入） |
| `PLATFORM_DEFAULT_TRIAL_DAYS` | `14` | 開通時的試用天數 |
| `PLATFORM_DEFAULT_GRACE_DAYS` | `7` | 逾期寬限天數 |
| `PLATFORM_DEFAULT_LEAD_DAYS` | `14` | 提前產生下一期的天數 |
| `SEED_PRICE_{FREE,STD,PRO}_{BASE,SEAT}` | 見下方佔位值 | 首次建立方案價目時的預設 |

**「可變更」要三個環節一起做到，否則只是 env 換個地方硬編**：

1. `SeedPlatform` 的價目改讀 `SEED_PRICE_*`，且**只作為首次建立的預設**——`plan_prices` 已有值即不覆寫（避免重跑 seed 把營運調過的價格蓋回去）。
2. 試用／寬限／提前天數寫入 `platform.settings`（key：`trial_days`／`grace_days`／`lead_days`）；**cron 讀 settings 而非硬編**（Plan C Task 7 Step 3）。
3. 營運工具提供編輯入口：Plan C Task 9 的 `UpdateBillingSettings` RPC ＋ console「方案與價目」頁的設定區塊。

**注意**：`SEED_PRICE_*` 的數字（免費 0／標準 1500＋150／專業 4500＋150）是為了讓計畫可執行而填的**佔位值**，上線前必須換成真實定價；`.env.example` 需標明這點。

- [ ] **Step 4: 接上 `cmd/seed/main.go` 與 `.env.example`**

```go
	// 平台域基礎資料(需 admin 連線:platform schema 為 owner 專屬)
	adminDB, err := sql.Open("pgx", cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("連線 admin 資料庫: %v", err)
	}
	defer adminDB.Close()
	if err := SeedPlatform(ctx, adminDB, cfg.Platform); err != nil {
		log.Fatalf("seed 平台域: %v", err)
	}
```

`.env.example` 追加：

```dotenv
# --- config/platform.go (Platform) ---
# 平台工具未設定則不掛載（開發環境可不設）
PLATFORM_JWT_SECRET=（獨立於 JWT_SECRET，兩者相同則拒絕啟動）
PLATFORM_ALLOWED_EMAIL_DOMAIN=sowinsoft.com
PLATFORM_CONSOLE_URL=http://localhost:5173          # 開發；正式為 console 的網址
# PLATFORM_COOKIE_DOMAIN=.sowinsoft.com             # 正式設；開發留空（host-only cookie）

# --- seed 預設值（上線前請改為真實值；執行期以 platform.settings 為準）---
PLATFORM_SEED_OPERATOR_EMAIL=ssd@sowinsoft.com
PLATFORM_SEED_OPERATOR_NAME=ssd
PLATFORM_SEED_SYSTEM_ACTOR_EMAIL=system@sowinsoft.com
PLATFORM_DEFAULT_TRIAL_DAYS=14
PLATFORM_DEFAULT_GRACE_DAYS=7
PLATFORM_DEFAULT_LEAD_DAYS=14

# --- 方案價目 seed 預設（**佔位數字，上線前務必改為真實定價**）---
SEED_PRICE_FREE_BASE=0
SEED_PRICE_FREE_SEAT=0
SEED_PRICE_STD_BASE=1500
SEED_PRICE_STD_SEAT=150
SEED_PRICE_PRO_BASE=4500
SEED_PRICE_PRO_SEAT=150
```

（`SeedOperatorEmail` 需加進 `config.Platform` 並在 Task 1 的測試補一行綁定斷言。）

- [ ] **Step 5: 跑測試**

Run: `cd backend && task test:integration -- -run TestIntegrationSeedPlatformIdempotent -v`
Expected: PASS（連跑兩次筆數不變）

- [ ] **Step 6: Commit**

```bash
git add backend/cmd/seed backend/config/platform.go backend/config/platform_test.go backend/.env.example
git commit -m "feat(backend): 平台域 seeder（8 features、3 方案與價目、首 operator，冪等）"
```

---

### Task 12: 慣例文件、CI 與計畫索引

**Files:**
- Modify: `backend/AGENTS.md`、`docs/superpowers/plans/README.md`

- [ ] **Step 1: `backend/AGENTS.md` 新增平台域小節**

（**編號**：`backend/AGENTS.md` §10 已由 Plan D 的「錯誤碼」佔用，故本節為 **§11**。**定稿為 13 條**——下面區塊是起草時期的 8 條草案，實作期間另外長出「守衛唯一入口與平台層略過」「配額是 check-then-act」「`/platform/` 前綴與 `requestid` interceptor」「跨租戶讀業務表的 `SET LOCAL`」「seed 的冪等定義」等條目；**唯一權威是 `backend/AGENTS.md` §11**，本區塊僅留為歷史記錄。）

```markdown
## 11. 平台域（SaaS 訂閱與權益，D34–D39）

1. **`platform` schema 與業務域不 JOIN**：只以 `company_id` 對照；跨域副作用一律經 `platform.events`。
2. **`app_rw` 對 `platform` schema 零權限**：平台域只走 `cfg.Database.AdminDSN()`。新增平台表時**不得**授權給 `app_rw`。
3. **平台表不用 ent**：`database/sql` ＋ 明確 SQL（`platform` schema 在 sqlite 不存在，ent codegen／auto-migrate 只會製造摩擦）；migration 是唯一 schema 真相。
4. **權益判定 fail-closed**：無訂閱／未定義 feature → 拒絕；額度不足與訂閱不可用一律 `FailedPrecondition`（`PermissionDenied` 保留給「缺權限」）。
5. **守衛必須可機械驗證**：新增寫入 RPC 要在 `internal/services/entitlement_guard_test.go` 的 `guardCases` 登記（RPC → feature），漏登記即測試紅。
6. **平台稽核不寫租戶 `audit_logs`**：`audit.Record` 要求 `company_id`／`user_id` 非零，平台操作者兩者皆無 → 一律 `platform.audit_logs`。
7. **能力命名空間**：`platform.*` 不得出現在租戶 `GetAbility`／角色權限矩陣（S11）。
8. **操作者與租戶身分不互通**：不同 JWT secret ＋ 不同 audience ＋ 不同 cookie path；跨用測試必須存在。
```

- [ ] **Step 2: 更新計畫索引（`docs/superpowers/plans/README.md`）**

把「SaaS 化 ①」列改為三列（① RLS 已備計畫／② 本計畫／③ 待寫），② 的狀態為 `🟡 計畫已備、未開工`，說明填「12 tasks：platform schema、store、entitlements、守衛、operator 認證、唯讀 RPC、投影、seeds」。

- [ ] **Step 3: 跑 CI 等價指令**

Run: `cd backend && task check && task test:integration`
Expected: 全綠

- [ ] **Step 4: Commit**

```bash
git add backend/AGENTS.md docs/superpowers/plans/README.md
git commit -m "docs(backend): 平台域慣例（schema 邊界／fail-closed／守衛登記／稽核獨立）與計畫索引更新"
```

---

## 驗收對照（spec）

| spec 要求 | 對應任務 |
|---|---|
| `platform` schema 與 10 張表、`app_rw` 零權限（§3、§3.3） | Task 2 |
| 方案／權益／override／訂閱的資料存取（§3.1） | Task 3 |
| `Allows`/`CheckLimit`、優先序、fail-closed、`FailedPrecondition`＋碼（`PLAT-5001`／`PLAT-5002`／`PLAT-3001`）（§4.1–4.3） | Task 4 |
| `Counter` 由業務域提供並注入（§4.2） | Task 5 |
| 守衛掛點清單與表驅動驗證（§4.5，含復原路徑） | Task 6 |
| `platform/v1`（平台工具與租戶端投影）（§2.2、§4.6） | Task 7、9、10 |
| 操作者 OIDC ＋ 白名單 ＋ 獨立 secret／cookie（S8） | Task 8 |
| 平台稽核獨立（S9） | Task 9 |
| 租戶端投影唯讀且只回自己公司（§4.6） | Task 10 |
| features／方案／價目／權益 seeds（§3.2、§4.5 清單） | Task 11 |
| 慣例成文、`platform.*` 不出現在租戶能力（S11） | Task 12 |

## 風險與對策

| # | 風險 | 對策 |
|---|---|---|
| B1 | 業務服務建構子改簽名造成大面積編譯錯誤 | 這是刻意的（編譯器強制每個呼叫端表態、不漏掛守衛）；以 `go build ./...` 的錯誤清單逐檔補 `entitlements.Unlimited()` |
| B2 | 平台 store 用 `database/sql`，SQL 錯誤只會在真 PG 出現 | store 的每個查詢都有整合測試（Task 3 Step 5）；單元測試走假實作，不假裝測到 SQL |
| B3 | 快取與權益異動競態（改了方案卻仍讀舊值） | 快取只在**讀取**路徑；Plan C 的寫入路徑負責 `Delete`。本計畫的 TTL 60s 為保底，且有「快取不改變判定結果」測試 |
| B4 | 操作者 cookie 與租戶 session 混用 | 不同 secret ＋ audience ＋ cookie name/path（Task 8 有三條安全契約測試） |
| B5 | 平台工具與租戶 SPA 共用 proto 時誤掛 `platform.*` 能力 | Task 12 把「命名空間隔離」寫成慣例；Plan C 的前端守衛測試需斷言租戶 ability 不含 `platform.*` |

---

*建立：2026-09-20（SaaS 化 spec 的第二份實作計畫：platform 域與權益守衛）*
