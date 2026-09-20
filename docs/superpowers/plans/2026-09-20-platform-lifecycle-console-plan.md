# 訂閱生命週期、收款與平台營運工具 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 讓平台域可以真的收錢與營運：金額計算、`RecordPayment`（人工收款，與日後金流同一入口）、訂閱狀態機與逾期凍結、事件 consumer 驅動產品域 `companies.status`、單趟可重跑的排程 binary、平台寫入 RPC 與平台稽核、以及獨立內部工具 `platform-console/` 六頁 ＋ 租戶後台唯讀權益卡片。

**Architecture:** 平台寫入集中在 `internal/platform/billing`（狀態機與帳務）與 `internal/platform/consumer`（outbox 派送）；**`platform` schema 與業務表同一個 PostgreSQL 資料庫**，因此「標記事件已派送 ＋ 改 `companies.status`」可在**單一交易**完成（原子、可重跑），outbox 僅作為解耦介面而不用於補償。排程為單趟執行（`cmd/platform-cron`），時間可覆寫以便測試。前端為獨立 package（`platform-console/`）：自有 Vite 建置、只走 `platform/v1`、以 alias 共用租戶 SPA 的 UI 元件庫（不共用路由與守衛）。

**Tech Stack:** Go 1.25、connect-go、goose、PostgreSQL 16、Valkey（go-redis v9）、SolidJS 1.9 + Vite 8 + TanStack Query/Table + Ark UI（console 與租戶 SPA 同一套）

**Spec:** `docs/superpowers/specs/2026-09-20-saas-billing-entitlements-design.md`（§2.4、§3.1、§5、§8）

**依賴**：
- **Plan A**（`2026-09-20-rls-tenant-isolation-plan.md`）：本計畫的產品域寫入（`SetStatus`）需在 Plan A 的請求層租戶交易語意下運作。若 Plan A 未完成，`SetStatus` 以 `dbtenant.SystemScopeTx` 包裝即可，不阻塞本計畫。
- **Plan B**（`2026-09-20-platform-entitlements-plan.md`）：`platform` schema、store、`entitlements.Service`、operator 認證、`PlatformAdminService` 皆為本計畫的前置。**本計畫 migration 自 `00030` 起**（Plan B 用到 `00029`）。

## Global Constraints

- **金額一律 `int64` 分（cents）**：禁止 `float32/float64` 參與任何金額運算；DB 為 `numeric(12,2)`，邊界以 `internal/platform/money` 轉換（`ParseCents`/`FormatCents`），該套件有金額路徑測試
- **平台寫入一律單一交易**：期別、訂閱狀態、事件、稽核同一 commit；跨域副作用（改 `companies.status`）因同庫而可同交易完成
- **每個平台寫入都寫 `platform.audit_logs`**，`reason` 必填（空字串即拒絕）；actor 為 `operator_id`（系統排程用 `platform.settings` 的系統 actor）
- **狀態機只有一條路徑**：人工收款與（日後）金流 webhook 都呼叫 `RecordPayment`；不得有第二個改變訂閱狀態的入口
- **排程單趟且可重跑**：`cron.RunOnce(ctx, deps, now)` 為唯一入口，`now` 由呼叫端給；重跑不得產生重複期別或重複事件（靠 unique index ＋ 狀態檢查）
- **console 只走 `platform/v1`**；不得引入租戶 SPA 的路由／守衛；租戶 SPA 也不得引入 `platform.*` 能力或平台路由（S11）
- **前端守衛不構成授權**：所有 disable 僅為 UX，後端為唯一決策者
- **通知（07-notifications）未實作**：v1 的催收以「待收款清單 ＋ 前端匯出 CSV」取代，不得假裝有自動通知
- 註解與 commit message 一律繁體中文
- 每任務結束前跑 `task check`；前端任務跑 `pnpm -C platform-console typecheck|test`；含整合測試者另跑 `task test:integration -- -run <TestName> -v`

---

## Progress

> 執行記錄與逐任務細節（每次審查的判定、裁決、修正輪、突變與回退實驗）見 `.superpowers/sdd/2026-09-20-platform-lifecycle-console-plan/progress.md`（ledger）。狀態：**15/15 任務完成（2026-09-21）**（含最終全分支審查 B-1 的補救任務 T15，以及其後的「最終修正波」一次性修正）。本計畫自 `7e31bae`（Plan B 收尾）起共 **32 個 commit**（`git log --oneline 7e31bae..HEAD | wc -l`；不含本檔案所在的最終修正波 commit）。

| # | 任務 | 狀態 | 產出（commit 範圍） | 交付後修正輪 |
|---|---|---|---|---|
| 1 | 產品域公司狀態入口 `SetCompanyStatus` | ✅ | `7e31bae`..`fadb6dc` | 無（spec ✅／quality Approved） |
| 2 | 金額套件 `internal/platform/money` | ✅ | `fadb6dc`..`d7ad29c` | 1 輪：`discountBps>10000` 靜默算出**負年費**（`(1200000,10001) = -1439`）→ 乘法前夾住兩端 |
| 3 | migration `00030` 與平台 store 交易版寫入 | ✅ | `d7ad29c`..`ee54b6f` | 1 輪（7 項）：0 列靜默成功、排程三查詢漏欄位（G1 少收 11 個月）、重播覆寫付款憑據、空 payload `''::jsonb` 22P02、`WithTx` 缺 `defer Rollback`、`void` 訊息誤導、fake 未反映唯一鍵 |
| 4 | `RecordPayment`（收款唯一入口） | ✅ | `ee54b6f`..`609e7df` | 1 輪：paid＋同交易號的重播分支**靜默丟棄 `note`**（短收／溢收唯一落點） |
| 5 | 生命週期排程邏輯（下一期／逾期／凍結／取消期末） | ✅ | `0799de4` ＋ 修正 `169321d`／`21355dc`（F-1／F-2 於 `9b7e12e`） | 1 輪（3 項）：**C-1 逾期謂詞**（已付款期別照樣轉 `past_due`）、**期別錨點永久漂移**（月底起租一年多收一期）、事件 payload 缺 `reason` |
| 6 | outbox consumer（事件驅動凍結／復原） | ✅ | `9b7e12e`..`1972efa` | 1 輪：單筆永久失敗事件**堵住後面所有租戶**的凍結（head-of-line）＋ driver 所有權註解 |
| 7 | `cmd/platform-cron` 單趟排程 | ✅ | `1972efa`..`ab52180` | 1 輪（5 項）：單租戶產生期別失敗**跳過整趟派送**、解鎖失敗鎖留池中、panic 丟摘要、`TryLock` 早於 recover、`main` 無逾時（卡住即永遠握鎖） |
| 8 | 權益快取 Valkey 實作與寫入後失效 | ✅ | `ab52180`..`67b65fb` | 無（Approved；8 條 P3 記帳，4 條帶入 T9） |
| 9 | 平台寫入 RPC 與平台稽核（含移除 `recordPlatformAudit`） | ✅ | `67b65fb`..`2d1da6e` | 1 輪：**operator 可自我提權／停用最後一位 admin**（安全）＋ `limit_value` 未擋負數 |
| 10 | `platform-console` 腳手架（operator 登入、client、守衛） | ✅ | `2d1da6e`..`f3dbc43` | 1 輪：**RPC 少 `/platform` 掛載前綴**（404＋cookie 不送出→無限導回登入）、README 跨來源指引與 CORS 相衝、登入頁套外框、探針結果永久快取 |
| 11 | 租戶列表／租戶詳情＋override／方案與價目 | ✅ | `f3dbc43`..`a765e3d` | 1 輪：權益投影宣稱「與後端同序」是錯的、對話框欄位跨次殘留（會送錯金額）、「權益下一期生效」與事實不符、`reason` 未 trim、死碼 |
| 12 | 權益矩陣（唯讀）／收款與發票／平台稽核 | ✅ | `a765e3d`..`668578c` | 1 輪：清空最後一頁→**假空狀態＋分頁消失的死路**、金額 `0` 會被當成「本期不收」而記為已收、CSV 無 BOM、query key 脫鉤、`target_type` 自由文字造成假空狀態 |
| 13 | 租戶後台唯讀權益卡片 | ✅ | `668578c`..`28b27a2` | 1 輪：`limit_set=true,limit_value=0` 被顯示成「不限」（方向性錯誤）、不可用訂閱狀態逐列寫「方案未含」、`used > limit` 仍說「接近上限」、`trialing→active` 未清試用到期 |
| 14 | CI、慣例文件與計畫文件（本任務） | ✅ | 本表所在 commit | 2 輪：`AGENTS.md` §10-6 與 §10-8 互相否定、§11-15 的 `reason` 契約與實作不符（2 Critical）＋6 項（spec 行號、數字對指令錯、`plans/README` 歷史值混淆、`test:integration` 缺 `-count=1`） |
| 15 | 最小訂閱開通介面 `CreateSubscription`（最終審查 B-1 的補救） | ✅ | `8bc8d60`..`35f78dd` | 1 輪：**`trialing` 無到期路徑＝無上界免費放行** → 新增 `ExpireTrials` 掃描（轉 `past_due`、走既有轉移表）＋ `trial_ends_at` 365 天上限；併修取價錯誤分流（缺價目 vs 查價失敗）、console 的 markdown 字面值與試用日期驗證、回滾證據補到 `OpenPeriodTx` 之後 |

每個任務都經**獨立審查**（spec 合規 ＋ quality）與必要的修正輪；修正輪一律附「修前會紅」的證據（突變、overlay、逐項回退或真容器重跑）。T5 之後凡動到 store／migration／seed 的任務，controller 另跑**完整**整合套件（聚焦 pattern 曾漏掉跨套件波及：F-1／F-2）。

**現況數字（以程式與指令重數，勿抄舊稿）**：

| 數字 | 值 | 確認指令 |
|---|---|---|
| migration 編號 | 至 **`00030`** | `ls backend/database/migrations \| tail -1` |
| `platform` schema 表數 | **11**（`00029` 的 10 ＋ `00030` 的 `settings`） | `grep -c "CREATE TABLE IF NOT EXISTS platform\." backend/database/migrations/00029_platform_schema.sql` ＝ 10；`grep -c "CREATE TABLE" backend/database/migrations/00030_platform_settings.sql` ＝ 1 |
| 平台寫入 RPC | **12**（**5 支**走 `billing`：開通／收款／席位／改方案／取消；7 支走服務層 `writeTx`：計費參數、override 設定／撤銷、方案價目、方案權益、新增／停用 operator） | `grep -c "s.writeTx(" backend/internal/services/platform_admin_service.go` ＝ 7；`grep -c "s.billing\.\(CreateSubscription\|RecordPayment\|SetSeatCount\|ChangePlan\|CancelSubscription\)" …` ＝ 5（T15 加入開通） |
| `platform/v1` RPC | **20**（T15 新增 `CreateSubscription` 前為 19） | `grep -c "^  rpc " backend/proto/platform/v1/platform.proto` |
| 平台事件型別（產生端） | **9**（`period.opened`／`period.payment_recorded`／`subscription.created`／`cancelled`／`expired`／`past_due`／`reactivated`／`suspended`／`trial_ended`）；其中排程產生 **5** 個 | `grep -rhoE '\bemit\(ctx, tx, [^,]+, "[a-z]+\.[a-z_]+"' backend/internal/platform/billing/*.go \| grep -oE '"[a-z]+\.[a-z_]+"' \| sort -u \| wc -l` ＝ 9 |
| consumer 對應的事件型別 | **3**（`subscription.suspended`／`expired` → `suspended`；`reactivated` → `active`；其餘型別記一行 log 後認領） | `sed -n '/^var actions = map/,/^}/p' backend/internal/platform/consumer/consumer.go` |
| 限 `role=admin` 的 RPC | **2**（`CreateOperator`／`DisableOperator`） | `grep -c "requireAdmin(ctx)" backend/internal/services/platform_admin_service.go` |
| 錯誤碼 | **22 碼**（SYS 7／AUTH 7／PLAT 5／CUST 3），**19 已落點** | 見 `AGENTS.md` §10-8（2026-09-21 重數） |
| errcode 基線 | **98 行／233 呼叫點**（只減不增，本計畫未動） | `wc -l < backend/internal/services/errcode_baseline.txt`；`awk -F: '{s+=$NF} END {print s}' …` |
| console 頁面 | **6**（`/tenants`、`/tenants/$tenantId`、`/plans`、`/entitlements`、`/receivables`、`/audit`；另有根路由 `"/"` 轉導與 `LOGIN_ROUTE`） | `grep -c 'path: "/' platform-console/src/router.tsx` ＝ **7**（含根路由 `"/"` 那一列，扣掉即 6 頁；`/login` 用 `LOGIN_ROUTE` 常數故不在這 7 列內） |
| console 測試 | **78**（12 檔） | `pnpm -C platform-console test` |
| CI job | **5**（go／go-integration／frontend／platform-console／flutter） | `grep -E "^  [a-z-]+:$" .github/workflows/ci.yml`（多一列是 `on:` 的 `push:`） |
| `AGENTS.md` §11 條目 | **23**（Plan B 交付時 14 條） | `sed -n '/^## 11\. 平台域/,/^## 12\./p' backend/AGENTS.md \| grep -cE "^[0-9]+\. \*\*"` |
| 本檔「未結項」列數 | **44** | `sed -n '/^### 未結項/,/^---/p' docs/superpowers/plans/2026-09-20-platform-lifecycle-console-plan.md \| grep -cE "^\| [0-9]+ \|"` |

### 執行期間的計畫更正（已寫回程式碼、AGENTS.md 或 spec）

| # | 更正 | 理由 |
|---|---|---|
| 1 | `SetCompanyStatus` **不自開交易、不 commit**（brief 樣板的 `dbclient.Tx(ctx)`＋`tx.Commit()` 皆錯） | 交易由 `dbtenant.Interceptor` 擁有；誤 commit 會讓所有經 RPC 的狀態變更二次提交而失敗。現為：讀 `dbtenant.Client(ctx, db)`、寫 `dbtenant.TxFrom(ctx)`、**不 commit** |
| 2 | 逾期掃描謂詞補 **`AND cur.status='open'`**（Task 5 的 C-1） | 舊謂詞只比 `period_end < now` → 逾期後才繳清的客戶被重新催收、寬限期重置、最後被凍結，且 `EnsureNextPeriod` 不再替他開帳（催收主線的靜默死路） |
| 3 | 期別錨點＝**該訂閱第一期起日的日號**（不用 `cur.PeriodEnd.Day()`，也不用 `time.AddDate`） | 2 月夾到 28 日後錨點永遠變 28 → 月底起租的月繳客戶 **364 天內 13 期**、帳單日永久漂移 |
| 4 | **排程不寫 `platform.audit_logs`**（核准偏離「每個平台寫入都寫稽核」） | `platform.audit_logs.operator_id` 是 `NOT NULL REFERENCES platform.operators`，而 `settings.system_actor_user_id` 是**租戶 `users.id`** → schema 上不可滿足。替代紀錄＝同交易 `platform.events` ＋ consumer 落的租戶稽核；**營運者驅動的寫入仍必須寫平台稽核**（已寫進 AGENTS.md §11-19） |
| 5 | 四個排程事件的 payload **必帶 `company_id` 與 `reason`**（`scheduled_next_period`／`period_end_passed_unpaid`／`overdue`／`cancelled_at_period_end`） | 排程不寫平台稽核，payload 是唯一的「為什麼」；亦讓 consumer 不必回頭查 DB（已寫進 spec §5.2.1 與 AGENTS.md §11-20） |
| 6 | `RecordPayment` 的 `PeriodNo=0` 改走 `CurrentPeriodTx`（不查 `PeriodsByStatus("open")`） | 樣板寫法會讓重送變成硬錯誤；「最新一期」與平台投影同義 |
| 7 | 期別已付款且交易號不同 → `PLAT-3002`（不靜默 log） | 靜默會讓第二筆匯入在帳面上消失 |
| 8 | **新增 `PLAT-3003`** 並以 `go generate` 同步三端常數與碼表 | 操作者治理的不變式需要可行動的專碼（用固定訊息的 `SYS-4001` 會讓 console 無法給指引）；errcode 的產生檔本來就是設計成由產生器同步、CI 驗證 |
| 9 | 角色矩陣**只收緊操作者管理**，完整「角色 × RPC」矩陣列為 spec 待辦 | 只有會改變權限／可回復性的操作先擋；其餘收緊需要一份明訂矩陣（已寫進 spec §2.4 待辦） |
| 10 | `cmd/platform-cron`（CLI）**不套 `errcode`** | 全域約束的價值是穩定的對外碼（跨網路契約）；CLI 錯誤只進 log，基線未加寬（已寫進 AGENTS.md §11-22） |
| 11 | 接受的範圍外編輯：`postgres/store.go` 唯讀 `Subscription` 補 `SELECT s.id`（5 行） | `Subscription` 新增 `ID` 欄位後，只讓寫入路徑填而唯讀路徑留 0 就是**靜默 0 值陷阱** |
| 12 | 計畫 Task 5 樣板碼的舊逾期 SQL 已被實作取代 | 樣板是派工用的草稿；實作加了 `cur.status='open'`，spec §5.2 已就地更正並標日期 |
| 13 | 排程事件名以 T4/T5 為準：`period.payment_recorded`（不是裁定文字裡的 `payment.recorded`） | 該事件本無跨域副作用，consumer 記一行 log 後認領，無功能影響 |
| 14 | **`task test:integration` 補上 `-count=1`**（`backend/Taskfile.yml`，T14 修） | 少了它整趟會命中 Go 測試快取並回報全綠（2026-09-21 實測：牆鐘 0.75s、全部 `(cached)`）→ **守門指令可假綠**；CI 的 `go-integration` job 本來就有 `-count=1`。**驗收指令一律要有「真的跑過」的證據** |
| 15 | **計畫漏列「開通」（`CreateSubscription`）＝計畫缺陷，由 T15 補上** | 計畫 Goal 是「平台域能真的收錢」，spec §7.4 的 DoD 第一跳就是開通，但 19 支 RPC 沒有任何建立訂閱／第一期的路徑 → 不靠手工 SQL 收不到一筆錢（收款／期別／催收／凍結全部 inert，測試都從手工種好的訂閱列起跑）。來源＝**最終全分支審查 B-1** |
| 16 | **`buf.gen.yaml` 的遠端 plugin 全面釘版**（`protocolbuffers/go:v1.36.12`／`connectrpc/go:v1.21.0`／`bufbuild/es:v2.15.0`×2／`connectrpc/dart:v2.0.0`） | 未釘版時 BSR 一上架新版，所有人的產生檔一起漂移、CI 的 proto 冪等閘門全紅（而漂移與改動無關，排查成本極高）。**釘版後以 `task proto:gen` ＋ `git diff --exit-code` 實測產物逐位元組不變**；版本本身以產生檔的表頭（`protoc-gen-go v1.36.12`／`protoc-gen-es v2.15.0`）與 `go.mod`／`pubspec.yaml` 對齊 |

### 未結項（deferred：現況、選項、歸屬）

> 來源＝`.superpowers/sdd/2026-09-20-platform-lifecycle-console-plan/progress.md` 的全部 `minor (deferred)` 與被裁定「延後／待辦」的項目（計畫已完成的範圍外事項）。多數是 P3（可接受、但別忘）；`final triage` 指的是「最終審查分診」當時未修、留待後續。

| # | 項目 | 現況 | 選項／修法 | 歸屬 |
|---|---|---|---|---|
| 1 | `company_status.go:42-45` 的 `reason` 檢查排在同值 no-op **之前** | 未動（consumer 一律帶 `reason`，故實務不可達） | 把 no-op 判斷提前，讓「無需動作」的變更不受 `reason` 阻擋 | `company_status` 後續維護（低） |
| 2 | `company_status_test.go:116-118`「缺交易」案例只斷言 `err != nil`（任何錯誤都通過） | 未動 | 斷言 `connect.CodeOf(err) == connect.CodeInternal`（同檔 :99-101 已示範） | final triage（測試強度） |
| 3 | `company_status.go:47`／`:58` 兩種取 client 寫法並存（`dbtenant.Client(ctx, db)` vs `tx.Client()`） | 語意正確、未動 | 綁一次變數 | final triage（風格） |
| 4 | 一次請求含 name＋status 時稽核寫 **2 列**（原 1 列），且一般更新列不再含 status | 未動；T12 的稽核頁逐列顯示，未做「一次請求一列 diff」 | console 端自行合併（本 repo 稽核無 request id），或接受 2 列分別代表兩個語意事件 | `platform-console` 稽核頁（若營運要求） |
| 5 | 測試 actor `UserID="1"` 在 sqlite 無對應 `users` 列 | 與既有 enttest 慣例一致（PG 才有 `00010` 的 FK） | — | final triage（測試） |
| 6 | `money` 套件：`num+5000` 未檢查回繞；`FormatCents` 的 `uint64` 修正／`mulCheck` 的 `MinInt64×-1`／`YearlyFromMonthly` 飽和路徑**無回歸測試**；`ParseCents("1.")` 接受但 `".5"` 被拒；未以 `numeric(12,2)` 上限約束 | 未動（刻意留給 DB 大聲失敗） | 依實際需求補測試或補上限約束 | `money` 後續維護（低） |
| 7 | `store.go:152-168` 介面 doc 未同步新語意（訂閱不存在→`sql.ErrNoRows`；憑據只在首次入帳寫；同 provider＋交易號不得入帳兩期；`void` 拒絕並說出狀態） | 未動（介面是 T4/T5/T6 作者讀的契約面） | 更新 `BillingStore` 的方法註解 | `store` 後續維護（文件） |
| 8 | 跨期衝突的**錯誤形狀** fake（格式化字串）與真 store（pgx 23505）不同；`cancelled_at` 無法由 fake 表達 | 未動（既有架構限制） | 讓 fake 回 `*pgconn.PgError` 等價物，或把取消時間語意限定在整合測試 | `platform/store` 的 fake（低） |
| 9 | `PLAT-3002` 把**基礎設施錯誤**（死鎖／序化／連線中斷）也講成收款衝突 | **T14 已在 spec §4.3 標明「不得據此放棄重試」**；結構化分流需動 proto | 在 `ErrorInfo.details` 加專屬鍵（例 `details.conflict`）；console 已有「無法分辨→通用指引」 | `platform/v1` 下一個計畫 |
| 10 | `RecordPayment` 的 no-op 分支**不比對金額**（同交易號、金額不同的第二筆匯款會被吸收） | 未動（成本極低的額外防護，非缺陷） | 在 no-op 分支比對金額並回 `PLAT-3002` | `billing` 後續維護（低） |
| 11 | 負數金額 → `SYS-1001` 是**未列於文件**的 Extra；`ActorOperatorID` 未驗證（0 會被 FK 23503 擋成 `SYS-9000`）；`now` 註解宣稱可注入但 `NewBilling` 無時鐘參數；`allowedTransitions` 其餘列待補測 | 未動 | 補文件／補驗證／補測試；webhook 落地時優先驗 actor | `billing` 後續維護（低～中） |
| 12 | `PLAT-3001` 的訊息對「資料／價目設定錯誤」語意錯配；「服務中」清單在三處各自表述；`MarkPastDue`／`SuspendOverdue` 的狀態更新**無列鎖／CAS** | 未動（靠 T7 的 advisory lock 兜住單趟；多實例時單飛鎖仍成立） | 改用 `UPDATE … WHERE status = <預期>` 的 CAS | `billing` 後續維護（中） |
| 13 | 期別錨點的 fallback（`period_no=1` 不在時退回當期起日）**靜默** | 未動（T14 只做文件化，未加檢查） | 在排程加資料完整性檢查並留痕（log／事件） | 排程告警（中） |
| 14 | 新語意殘角：當期已 `paid` 但下一期開不出來（價目缺失／`billing_cycle` 非法 →`PLAT-3001`）→ 訂閱**永遠停在 active、既不被催收也不再被開帳** | 未動 | 加告警「服務中卻沒有 open 期別」（指標或排程摘要欄位） | 排程告警（中） |
| 15 | `DispatchOnce` 回傳「本趟**認領數**」（不是成功派送筆數） | 已註明語意；若被當成「副作用發生次數」會誤報 | 重新命名或另回一個欄位 | `consumer` 後續維護（低） |
| 16 | actor 解析失敗不再中止整趟 → 每個 mapped 事件各 append 一條同根因錯誤 → `errors.Join` 訊息 O(N) 膨脹（僅 log 噪音） | 未動 | actor 失敗記一次並跳過其餘 | `consumer` 後續維護（低） |
| 17 | `reactivated` 可能解除「**管理員手動停用**」的公司（公司狀態只有一個列舉、不區分停用來源） | 裁定映射表所定（spec §5.2：`suspended → active` 由收款觸發），**非缺陷** | 產品語意定調（是否區分停用來源） | product／spec 擁有者 |
| 18 | `Params.EventBatch` 與 `const EventBatch` 兩層入口；`--date` 可任意倒回（會把 `grace_until` 設成過去）；**`PeriodsByStatus("open")` 的待收款含 G5 平台自營公司**（與 console 投影不一致）；`Summary.Locked` 是第 7 個欄位 | 未動（T9 的 RPC 待收款已排除 G5，排程這端未排） | 排程端套同一排除條件；`--date` 限制不得早於上次執行 | `platform/cron` 後續維護（低～中） |
| 19 | 整合測試的 `GRANT EXECUTE ON pg_advisory_unlock` 還原排在 `t.Fatalf` 之後；`run.go:331` 註解措辭；`run_test.go:356` 註解筆誤（43/44）；`fakeLocker` 的 `ctx.Err()` 閘門與真實 locker 行為分歧；**「一個永久壞資料的租戶會讓每趟 CronJob 以 exit 1 收場（派送其實成功）」** | 未動 | 還原改 `t.Cleanup`；日後考慮「降級成功」模式以免告警噪音 | `platform/cron` 測試與告警（低） |
| 20 | 權益快取其餘 P3：`InvalidateAll` 首個 `DEL` 失敗即 return；快取錯誤的「不支援全量失效」曾用 `errors.New`（T9 已提供 sentinel）；`openEntitlementCache` 的 `MemoryCache` 分支今日不可達；快取故障時 **log 無節流**；`TestJudgementSurvivesCacheFailure` 改動**行程級 logger**；`cmd/platform-cron` 未關 Valkey client | 未動（T9 已處理其中兩項：`ttl<=0` 語意、sentinel） | `errors.Join` 收斂部分失敗；log 節流；測試改注入 logger | `entitlements` 快取（低） |
| 21 | 註解宣稱「正確性**不依賴 TTL**」略強於實情（cache-aside 競態：提交前 miss、提交後 `Set` 可把舊快照蓋過剛做的 `Delete`） | **spec §4.4 已改寫為「失效由寫入方驅動；TTL 是最長收斂上界」並標日期**；**程式註解（`entitlements/valkey.go:20`／`cache.go:10`）未動**（本任務範圍不含程式碼） | 順手把那兩行註解改成同一措辭 | `entitlements` 後續維護（文件） |
| 22 | `SetSeatCount` 守衛的交易外讀取（TOCTOU）；`UpdateBillingSettings` 的 before 交易外讀取；`ValkeyCache.Set(ttl<=0)` 只有容器測試守護；`cron.RunOnce` 的 Receivables 未排除 G5；`ChangePlan` 的錯誤碼優先序（無訂閱＋未知方案回 `SYS-4002`）；`platform_admin_write_test.go` 的 doc comment 名稱與函式名不符 | 未動 | 把讀取搬進交易；補單元守門；統一錯誤碼優先序 | `platform_admin_service` 後續維護（低～中） |
| 23 | console 不依角色隱藏操作（`role=operator` 在操作者管理頁會吃到 `SYS-4001`） | **本任務記入 spec §2.4 待辦**（需要 `GetOperatorSelf`，會動 proto）；後端仍是唯一決策者 | 新增 `GetOperatorSelf`（或等價 RPC）＋前端依角色 disable | `platform/v1` 下一個計畫 |
| 24 | T9 修正後**整合測試夾具殘留**（把 `ops-ghost@example.com` 留在 active admin）；advisory lock 的保證**依附 isolation level** | 殘留無害（斷言用差值）；isolation 已寫進 **AGENTS.md §11-16**（現況 RC 安全，改 DSN 前須重驗） | 清理 fixture；在部署文件明示 `default_transaction_isolation` | `platform` 測試（低）／部署文件 |
| 25 | `logout()` **不取消已在飛行中的探針**（其 `.then` 仍會 `setStatus("authenticated")`）；TTL 讓「飛行中探針被替換」成為新路徑（舊解析可能覆寫較新的 status）；`/login` 精確字串對未來的 `/login/verify` 子路由會誤套外框；缺「其他頁仍有外框」的反向保護；`session.test.ts` 的 mock 有第三份字面值 | 未動（守衛仍 fail-closed，後端仍擋） | 在 `.then` 內檢查 `loggedOut`／身分，並補反向斷言 | `platform-console` `lib/session`（低） |
| 26 | `TenantDetailPage.tsx:249` 的 `now` 只在元件建立時取一次（頁面開著時到期的例外仍算生效）；`:80` 到期日檢查寬鬆（`2027/01/01` 先過再由後端擋）但訊息說「格式須為 RFC3339」 | 未動 | 用 signal 讓 `now` 隨時間前進；放寬或收緊訊息 | `platform-console`（低） |
| 27 | `TenantDetailPage` 缺「試用到期／寬限期」欄位（`TenantSummary`／`GetTenantResponse` 沒有） | 未編造（未顯示） | 在投影加欄位（動 proto） | `platform/v1` 下一個計畫 |
| 28 | **沒有「列出某訂閱的期別」RPC** → 租戶詳情無法顯示期別歷史（spec §2.4 描述有「期別」） | 未做（需新增 RPC，會動 proto） | 新增 `ListSubscriptionPeriods`（或把期別塞進 `GetTenantResponse`） | `platform/v1` 下一個計畫 |
| 29 | 權益投影的 `ambiguous` 論述**在真實部署不可達**（`00029` 的部分唯一索引 ＋ console 只取未撤銷列 ⇒ 同功能至多一筆未撤銷例外） | 防禦碼無害；論述措辭仍寫「撤銷是單向的…會有第二筆」（未改） | 改該句註解措辭；補「恰一筆 → `ambiguous=false`」邊界斷言；「立即生效」文案補「快取失效失敗時最長一個 TTL 才收斂」 | `platform-console` `lib/entitlements.ts`（低） |
| 30 | `PLAT-3002` 的**字串比對**（依 `details.reason` 的中文關鍵詞分流兩種語意）；`MarkPeriodPaidTx` 唯一鍵衝突**無** `details.reason` → 走通用指引；CSV 下載未真機驗證；**收款表單未收 `invoice_status`／`buyer_tax_id`／`carrier`**（proto 有、後端可同筆寫入）；CSV 只含目前這一頁；稽核頁與收款頁自身的查詢錯誤態無專屬測試 | 未動（受限於不得動後端／proto） | 加結構化 `details` 鍵；開票三欄收進表單；補錯誤態測試 | `platform/v1` 下一個計畫／`platform-console` |
| 31 | 清空最後一頁後 `page` 停在 2 而 `lastPage` 變 1 → 顯示「**第 2 / 1 頁（共 50 筆）**」的矛盾頁碼 | 未動（上一頁仍可按、總筆數為真） | 當頁變空時把頁碼夾回 `lastPage`，或讓 Pagination 顯示值夾住 | `platform-console`（低） |
| 32 | `PlanBanner.test.tsx` 以 `queryKey` 字串當同步閘（改 key 會假紅）；`USABLE_STATUSES` 是後端 `usable()` 的**手抄副本**；`usageValue` 的數值分支不再看 `featureCode` 前綴；提示條只有兩級色階 | 未動（漂移方向是「過度阻擋」，大聲） | 把 allow-list 寫成後端契約測試；色階分級 | `frontend` account 功能（低） |
| 33 | **Plan B 未結項 #13 仍留給後端**：投影在「無訂閱列」時仍讓每個 boolean feature 列 `enabled=false` | 前端以 `isUnprovisioned` 因應；後端未改 | 投影加「尚未開通訂閱」的專屬形狀 | `platform/entitlements` 投影（訂閱指派落地時） |
| 34 | **store 層的 `reason` 只擋空字串，不 trim**（`recordAuditTx`：`postgres/billing.go` 與 `fake_billing.go` 各一份、`admin_writes.go` 的 `RecordAuditTx` 只是轉呼叫） | 目前**不可達**：三個進入點（`RecordPayment`／`internal/platform/billing` 的 `requireReason`（四支訂閱寫入共用）／`platform_admin_service.go` 的 `platformReason`）都以 `TrimSpace` 先擋。但 `platform.audit_logs.reason` 是 NOT NULL、**空字串合法** → 未來新增寫入點若忘了 trim，`"   "` 會寫成一列看起來有值卻沒有理由的稽核（T14 審查發現，已寫進 AGENTS.md §11-15 的事實敘述） | store 端也 `TrimSpace` 後再判定（1 行，需動程式碼）；或把「進入點必須 trim」寫成 store 介面的契約註解 | `platform/store` 後續維護（低～中） |
| 35 | **`task test:integration` 曾缺 `-count=1`**（可假綠） | **T14 已修**（`backend/Taskfile.yml` 一行，見更正表第 14 列）；同類風險仍在其他「跑測試」的入口：任何只回報「全綠」而不帶 `-count=1` 的守門指令都可能命中快取 | 新增任何測試指令時一律帶 `-count=1`（或 `-count=1` 由 Taskfile 統一提供） | 全 repo 的 Taskfile／CI 慣例 |
| 36 | **開通時傳入不存在的 `company_id` → FK `23503` 被 `toConnectError` 收成 `SYS-9000`**（`CreateSubscription`；console 只能選既有租戶，故僅自製客戶端可達） | 未動 | 服務層先以租戶投影驗存在（回 `SYS-4002`），或把 23503 分流成 4xx | `platform/v1` 下一個計畫（低） |
| 37 | **價目 `base=seat=0` 仍能開出 0 元期別**（`UpdatePlanPrice` 不擋 0；開通只擋「沒有價目」） | 未動（若「免費方案」是刻意設計則可接受） | 價目寫入路徑擋 0，或為免費方案立旗標 | 產品語意定調（低） |
| 38 | **`subscription.created` 對 consumer 是未知型別**（記一行 log 後認領，無副作用；排程不會重掃） | 安全，但事件流的價值只剩審計 | 日後要在產品域對開通做事時，於 consumer 的 `actions` 明列語意 | `platform/consumer`（低） |
| 39 | **console 的開通按鈕顯示條件綁在「同一公司只能有一份未取消合約」**（只在 `none`／空／`cancelled` 顯示，與 00029 的部分唯一索引同義） | 未動 | 後端若放寬成「同公司多份合約」，前端要同步放寬（否則介面比後端保守而不自知） | 後端放寬時同步（低） |
| 40 | **`trial_ends_at IS NULL` 的 `trialing` 訂閱無告警**：`ExpireTrials` 的謂詞刻意只認 `trial_ends_at IS NOT NULL`（不讓排程猜），而這種列**永遠停在試用**（可用、不催收、不凍結） | 未動（寫入路徑已擋：開通要求未來日期；DB 層無約束） | 加檢查（指標／排程摘要欄位）或 `CHECK (status <> 'trialing' OR trial_ends_at IS NOT NULL)` | 排程告警／migration（中） |
| 41 | **試用期長於一期時，第一期會在試用期間到期且不被催收**（第一期一律自 `now` 起算，而 `MarkPastDue` 只掃 `active`） | 未動（T15 已記錄） | 「試用期不開帳」需同時改 `EnsureNextPeriod` 與催收謂詞 | 產品語意定調（中） |
| 42 | **開通只有 RPC／console 一條路徑**：seed／CLI 若要種 demo 訂閱**必須**呼叫 `billing.CreateSubscription`，不得直接 INSERT（否則第一期、事件、稽核、價格快照全缺） | 未動（已寫進 AGENTS.md §11-14） | — | 慣例（已文件化） |
| 43 | **`store.BillingStore.MarkEventDispatchedTx` 只剩測試在用**（consumer 走 `claimSQL` 的條件式認領）；移除會連帶刪掉 fake／整合測試的覆蓋 | 未動（介面 doc 已標「僅測試／歷史」） | 若 consumer 改走 store 的標記，兩者合併成一個 | `platform/store`（低） |
| 44 | **`money.YearlyFromMonthly` 沒有生產呼叫端**（proto 與 DB 都還沒有 `discount_bps`） | 未動（函式註解與 AGENTS.md §11-18 已標明；保留是為了價目加上年繳折扣那天） | 價目寫入路徑加上折扣時由它負責算，勿另寫一份 | `platform/money`（低） |

---

### 共用樣板（每個任務都適用）

**平台寫入的交易骨架**（`internal/platform/billing`；store 的寫入方法一律接受 `*sql.Tx`）：

```go
err := b.withTx(ctx, func(tx *sql.Tx) error {
    // 1) 讀取現況（必要時 SELECT … FOR UPDATE 鎖住 subscription 列）
    // 2) 寫入異動（期別／訂閱狀態）
    // 3) 寫事件（platform.events）
    // 4) 寫稽核（platform.audit_logs，reason 必填）
    return nil
})
```

**狀態機轉移一律表驅動**：合法轉移寫在 `allowedTransitions map[string][]string`，測試逐格驗證合法與非法（非法一律 `FailedPrecondition`）。

---

## File Structure

| 路徑 | 職責 |
|---|---|
| `internal/platform/money/money.go`（新） | cents ↔ numeric 字串、期別金額計算、年繳折扣 |
| `internal/services/company_status.go`（新） | `SetCompanyStatus`（產品域唯一狀態入口；`UpdateCompany` 與 consumer 共用） |
| `internal/platform/billing/billing.go`（新） | `RecordPayment`／`EnsureNextPeriod`／`MarkPastDue`／`SuspendOverdue` ＋ 狀態機表 |
| `internal/platform/consumer/consumer.go`（新） | outbox 派送：`subscription.suspended/reactivated` → 產品域 `SetStatus`（同交易標記已派送） |
| `internal/platform/cron/run.go`（新） | `RunOnce(ctx, deps, now) (Summary, error)`：逾期 → 凍結 → 產生期別 → 派送事件 |
| `cmd/platform-cron/main.go`（新） | 極薄入口：`config.New()` → admin DB → `cron.RunOnce` → log 摘要 |
| `internal/platform/store/postgres/*.go`（改） | 期別／訂閱／事件／稽核／settings／待收款的 SQL 與交易版方法 |
| `internal/platform/entitlements/valkey.go`（新） | `Cache` 的 Valkey 實作 |
| `backend/proto/platform/v1/platform.proto`（改） | 寫入 RPC：override、RecordPayment、方案／價目、operator、待收款 |
| `internal/services/platform_admin_service.go`（改） | 寫入 RPC 實作 ＋ 稽核 ＋ 快取失效 |
| `platform-console/`（新 package） | 獨立 SPA：六頁、operator 登入、API client、路由守衛、UI alias |
| `frontend/src/features/account/PlanCard.tsx`（新） | 租戶後台唯讀權益卡片 |
| `pnpm-workspace.yaml`、`Taskfile.yml`、`.github/workflows/ci.yml`、`backend/AGENTS.md`（改） | workspace／任務／CI／慣例 |

---

### Task 1: 產品域的公司狀態入口（`SetCompanyStatus`）

**Files:**
- Create: `internal/services/company_status.go`、`internal/services/company_status_test.go`
- Modify: `internal/services/company_service.go`（`UpdateCompany` 的 status 分支改走新入口）

**Interfaces:**
- Produces: `services.SetCompanyStatus(ctx context.Context, db *ent.Client, companyID int, status company.Status, reason string, actor authz.Identity) error`

**為何需要**：平台域的凍結 consumer 需要一個**不經 RPC、不經租戶身分**的狀態變更入口；目前的邏輯埋在 `UpdateCompany` 內（`company_service.go:289`），抽出來讓兩邊共用，並保證「停用連鎖」語意只有一份。

- [ ] **Step 1: 寫失敗測試（`company_status_test.go`，sqlite）**

```go
package services

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

// TestSetCompanyStatusWritesAuditAndSkipsNoop：
// ① 狀態真的變更時寫一筆 action=update 的稽核（before/after 帶 status）；
// ② 同值再設一次不寫稽核（避免排程每日重跑灌爆稽核）。
func TestSetCompanyStatusWritesAuditAndSkipsNoop(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-1").
		SetStatus(company.StatusActive).SaveX(ctx)
	actor := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}

	if err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "欠費停用", actor); err != nil {
		t.Fatalf("停用: %v", err)
	}
	got := db.Company.GetX(ctx, co.ID)
	if got.Status != company.StatusSuspended {
		t.Fatalf("狀態應為 suspended，got %s", got.Status)
	}
	n := db.AuditLog.Query().CountX(ctx)
	if n != 1 {
		t.Fatalf("應寫 1 筆稽核，got %d", n)
	}

	if err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "欠費停用", actor); err != nil {
		t.Fatalf("同值設定不應報錯: %v", err)
	}
	if n := db.AuditLog.Query().CountX(ctx); n != 1 {
		t.Fatalf("同值設定不得再寫稽核，got %d 筆", n)
	}
}

// TestSetCompanyStatusRejectsDeletedCompany：已軟刪除的公司不得被改狀態。
func TestSetCompanyStatusRejectsDeletedCompany(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	co := db.Company.Create().SetName("C").SetIdentifier("SETST-2").
		SetDeletedAt(time.Now().UTC()).SaveX(ctx)
	err := SetCompanyStatus(ctx, db, co.ID, company.StatusSuspended, "欠費停用",
		authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}})
	if err == nil {
		t.Fatal("已刪除公司不得改狀態")
	}
}
```

（`newTestDB` 為既有 sqlite 測試輔助；若名稱不同，沿用該套件既有者。）

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/services/ -run TestSetCompanyStatus -v`
Expected: FAIL —`undefined: SetCompanyStatus`

- [ ] **Step 3: 實作（`company_status.go`）**

```go
// SetCompanyStatus 為公司狀態的唯一變更入口：RPC（UpdateCompany）與平台域的
// 訂閱事件 consumer（凍結／復原）共用，確保「停用連鎖」語意只有一份。
// actor 為觸發者：平台排程觸發時帶系統 actor（見 platform.settings），仍落租戶稽核。
package services

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

// SetCompanyStatus 設定公司狀態並在狀態真的改變時寫稽核（同一交易）。
// 同值為 no-op（不寫稽核）：排程每日重跑不得灌爆稽核。
func SetCompanyStatus(ctx context.Context, db *ent.Client, companyID int,
	status company.Status, reason string, actor authz.Identity) error {
	cur, err := dbtenant.Client(ctx, db).Company.Query().
		Where(company.ID(companyID), company.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if cur.Status == status {
		return nil
	}
	if strings.TrimSpace(reason) == "" {
		return connect.NewError(connect.CodeInvalidArgument,
			errors.New("狀態變更必須提供原因"))
	}

	tx, err := dbclient.Tx(ctx) // 依 Global Constraints：Plan A 完成後為 dbtenant.Client(ctx, db)
	if err != nil {
		return toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	updated, err := tx.Company.UpdateOneID(companyID).SetStatus(status).Save(ctx)
	if err != nil {
		return toConnectError(err)
	}
	id, _ := parseID(actor.UserID)
	before := map[string]any{"status": string(cur.Status)}
	after := map[string]any{"status": string(updated.Status), "reason": reason}
	if err := recordAuditBA(ctx, tx, "company", "status", companyID, companyID, nil, id,
		before, after); err != nil {
		return toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return toConnectError(err)
	}
	return nil
}
```

（實作時以該檔既有的 client 取得方式為準：Plan A 完成後為 `dbtenant.Client(ctx, db)`，未完成則為 `db`；交易同理。**不要**同時出現兩種寫法。）

- [ ] **Step 4: `UpdateCompany` 改走新入口**

把 `UpdateCompany` 內 `msg.Status != nil` 的分支改成：

```go
	if msg.Status != nil {
		if !validCompanyStatuses[*msg.Status] {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				fmt.Errorf("無效的公司狀態 %q(允許: active / inactive / suspended)", *msg.Status))
		}
		st := company.Status(*msg.Status)
		if st != exists.Status {
			// 狀態變更走共用入口（與平台域的凍結／復原同一份語意與稽核格式）。
			if err := SetCompanyStatus(ctx, s.db, id, st, "管理員手動變更", authz.IdentityFrom(ctx)); err != nil {
				return nil, err
			}
			changed = true
		}
	}
```

（`build` 鏈上的 `SetStatus` 一併移除，避免同一次更新寫兩次狀態與兩筆稽核；`before/after` 的 status 欄位由 `SetCompanyStatus` 負責。）

- [ ] **Step 5: 跑測試**

Run: `cd backend && task check`
Expected: 全綠（既有 `UpdateCompany` 測試必須全過——若因稽核筆數改變而紅，代表狀態變更被寫了兩次，回頭修 Step 4）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/company_status.go backend/internal/services/company_status_test.go backend/internal/services/company_service.go
git commit -m "refactor(backend): 抽出 SetCompanyStatus 為公司狀態唯一入口（RPC 與平台 consumer 共用）"
```

---

### Task 2: 金額套件（`internal/platform/money`）

**Files:**
- Create: `internal/platform/money/money.go`、`internal/platform/money/money_test.go`

**Interfaces:**
- Produces:

```go
func ParseCents(s string) (int64, error)   // "1500.00" → 150000；>2 位小數即錯
func FormatCents(c int64) string           // 150000 → "1500.00"
func PeriodAmount(baseCents, seatCents int64, seats int) (int64, error)
func YearlyFromMonthly(monthlyCents int64, discountBps int) int64
```

- [ ] **Step 1: 寫失敗測試（`money_test.go`）**

```go
package money_test

import (
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
)

func TestParseCentsRejectsAmbiguity(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"0", 0, false},
		{"1500", 150000, false},
		{"1500.5", 150050, false},
		{"1500.05", 150005, false},
		{"1500.005", 0, true}, // 超過兩位小數：拒絕，不得默默四捨五入（帳務不得失真）
		{"", 0, true},
		{"abc", 0, true},
		{"-10.00", 0, true}, // 金額不得為負（退款是獨立流程，不是負應收）
	}
	for _, tc := range cases {
		got, err := money.ParseCents(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseCents(%q) 應報錯，got %d", tc.in, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Fatalf("ParseCents(%q) = %d, err=%v；want %d", tc.in, got, err, tc.want)
		}
	}
}

func TestFormatCentsRoundTrip(t *testing.T) {
	for _, in := range []string{"0.00", "1500.00", "1500.05", "12.30"} {
		c, err := money.ParseCents(in)
		if err != nil {
			t.Fatalf("ParseCents(%q): %v", in, err)
		}
		if got := money.FormatCents(c); got != in {
			t.Fatalf("FormatCents(%d) = %q；want %q", c, got, in)
		}
	}
}

// 期別金額 = 月費（或年費）＋ 席位單價 × 席位數；整數運算，不得溢位。
func TestPeriodAmount(t *testing.T) {
	got, err := money.PeriodAmount(150000, 15000, 10) // 1500 + 150×10 = 3000
	if err != nil || got != 300000 {
		t.Fatalf("PeriodAmount = %d, err=%v；want 300000", got, err)
	}
	if _, err := money.PeriodAmount(100, 100, -1); err == nil {
		t.Fatal("負席位數應報錯")
	}
	if _, err := money.PeriodAmount(math.MaxInt64, math.MaxInt64, 2); err == nil {
		t.Fatal("溢位應報錯，不得默默回繞")
	}
}

// 年繳 = 月費 × 12 × (1 - 折扣基點/10000)，四捨五入到分。
func TestYearlyFromMonthly(t *testing.T) {
	if got := money.YearlyFromMonthly(150000, 1000); got != 1620000 { // 1500×12×0.9
		t.Fatalf("年繳 = %d；want 1620000", got)
	}
	if got := money.YearlyFromMonthly(100000, 0); got != 1200000 {
		t.Fatalf("無折扣年繳 = %d；want 1200000", got)
	}
	if got := money.YearlyFromMonthly(12345, 1000); got != 133326 { // 123.45×12×0.9 = 1333.26
		t.Fatalf("四捨五入到分 = %d；want 133326", got)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/platform/money/ -v`
Expected: FAIL —`undefined: money.ParseCents`

- [ ] **Step 3: 實作（`money.go`）**

```go
// Package money 以「分」為單位的整數運算處理金額：不用 float（二進位浮點無法精確表示
// 十進位金額），DB 邊界以字串轉換 numeric(12,2)。
package money

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseCents 解析 "1500.00" 形式的金額字串為分；超過兩位小數或負值一律拒絕
// （帳務不得四捨五入掉使用者看不到的尾差）。
func ParseCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("金額為空")
	}
	if strings.HasPrefix(s, "-") {
		return 0, errors.New("金額不得為負")
	}
	intPart, fracPart, hasFrac := strings.Cut(s, ".")
	if intPart == "" {
		return 0, fmt.Errorf("金額格式錯誤: %q", s)
	}
	for _, r := range intPart + fracPart {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("金額格式錯誤: %q", s)
		}
	}
	if hasFrac {
		if len(fracPart) > 2 {
			return 0, fmt.Errorf("金額最多兩位小數: %q", s)
		}
		for len(fracPart) < 2 {
			fracPart += "0"
		}
	} else {
		fracPart = "00"
	}
	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("金額過大: %q", s)
	}
	if whole > math.MaxInt64/100 {
		return 0, fmt.Errorf("金額過大: %q", s)
	}
	cents, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("金額格式錯誤: %q", s)
	}
	return whole*100 + cents, nil
}

// FormatCents 將分格式化為兩位小數字串（DB numeric(12,2) 的寫入形式）。
func FormatCents(c int64) string {
	neg := c < 0
	if neg {
		c = -c
	}
	s := fmt.Sprintf("%d.%02d", c/100, c%100)
	if neg {
		return "-" + s
	}
	return s
}

// PeriodAmount 計算單期金額 = baseCents ＋ seatCents × seats（整數運算，溢位即錯）。
func PeriodAmount(baseCents, seatCents int64, seats int) (int64, error) {
	if seats < 0 {
		return 0, errors.New("席位數不得為負")
	}
	seatTotal, err := mulCheck(seatCents, int64(seats))
	if err != nil {
		return 0, err
	}
	return addCheck(baseCents, seatTotal)
}

// YearlyFromMonthly 由月費推導年費：月費 × 12 後套用折扣基點（1000 = 9 折），
// 結果四捨五入到分。
func YearlyFromMonthly(monthlyCents int64, discountBps int) int64 {
	yearly, err := mulCheck(monthlyCents, 12)
	if err != nil {
		return monthlyCents * 12 // 極端值：回退為無折扣（呼叫端金額由 seed/工具設定，非使用者輸入）
	}
	if discountBps <= 0 {
		return yearly
	}
	// x * (10000 - bps) / 10000，四捨五入：先乘後除以避免精度損失。
	num, err := mulCheck(yearly, int64(10000-discountBps))
	if err != nil {
		return yearly
	}
	return (num + 5000) / 10000
}

func addCheck(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, errors.New("金額溢位")
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, errors.New("金額溢位")
	}
	return a + b, nil
}

func mulCheck(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	r := a * b
	if r/b != a {
		return 0, errors.New("金額溢位")
	}
	return r, nil
}
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/platform/money/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/platform/money
git commit -m "feat(backend): 金額套件（int64 分、禁 float、期別金額與年繳折扣）"
```

---

### Task 3: `migration 00030`（settings）與 store 的交易版寫入

**Files:**
- Create: `database/migrations/00030_platform_settings.sql`

**[controller 加的硬契約（來自 Plan B 的 seed 實作，2026-09-20）]**
1. **鍵名必須與 Plan B 的 seeder 對齊**：`cmd/seed/platform.go`（Plan B Task 11）在「`platform.settings` 存在時」會寫入 `system_actor_user_id`／`trial_days`／`grace_days`／`lead_days`（這四個 key 同時是 **G5**（系統 actor）與本計畫 cron 的參數來源）。本 migration 的欄位／鍵名稱與型別必須與它一致，否則 seed 寫得進去但讀不到（或反之）。**落地後必須重跑 `cmd/seed`**（或由本計畫的 seed 補寫）——否則 `cmd/platform-cron` 一開跑就 Fatal，即 **G5 的原始缺陷原樣回來**。
2. **與 seed 的寫入策略互動**：Plan B 的 seeder 對 `features`／`plans`／`plan_entitlements` 是 **`DO UPDATE`**（價目已改成「只補缺」）。若本計畫的 console 之後允許營運編輯方案權益，**重跑 seed 會把營運調整蓋回去** → 屆時必須把那三張表也改成「只補缺」或在 seed 中排除已有列（本任務不需修，但需知道）。
- Create: `internal/platform/store/postgres/billing.go`
- Create: `internal/platform/store/postgres/billing_integration_test.go`
- Modify: `internal/platform/store/store.go`（新增期別／訂閱／事件／稽核／settings 的介面方法）

**Interfaces:**
- Produces（`store.go` 追加）：

```go
type Period struct {
	ID             int64
	SubscriptionID int64
	PeriodNo       int
	PeriodStart    time.Time
	PeriodEnd      time.Time
	PlanID         int64
	UnitPriceCents int64
	SeatPriceCents int64
	SeatCount      int
	AmountCents    int64
	Currency       string
	Status         string // open | paid | void
	PaidAt         *time.Time
	InvoiceNo      string
	PaymentProvider string
	ExternalRef    string
	Note           string // 短收／溢收等人工註記（G8；付款時寫入）
}

// OpenPeriodInput 為建立一期所需的快照欄位（不含 note：註記屬付款事件，見 MarkPeriodPaidTx）。
type OpenPeriodInput struct {
	SubscriptionID int64
	PeriodNo       int
	PeriodStart    time.Time
	PeriodEnd      time.Time
	PlanID         int64
	UnitPriceCents int64
	SeatPriceCents int64
	SeatCount      int
	AmountCents    int64
	Currency       string
}

// Price 為「當期生效價」（GB1：以 billing_cycle 對應的價目）。
type Price struct {
	BaseCents int64
	SeatCents int64
	Currency  string
}

// Event 為 outbox 事件（consumer 派送用）。
type Event struct {
	ID            int64
	AggregateType string
	AggregateID   int64
	EventType     string
	Payload       []byte
}

// BillingStore 由 internal/platform/billing 使用；所有寫入方法接受 *sql.Tx 以保證單一交易。
type BillingStore interface {
	OpenSubscriptionTx(ctx context.Context, tx *sql.Tx, companyID int) (*Subscription, error) // SELECT … FOR UPDATE
	SetSubscriptionStatusTx(ctx context.Context, tx *sql.Tx, subID int64, status string, graceUntil *time.Time) error
	OpenPeriodTx(ctx context.Context, tx *sql.Tx, in OpenPeriodInput) (*Period, error)
	OpenPeriodByNoTx(ctx context.Context, tx *sql.Tx, subID int64, periodNo int) (*Period, error)
	CurrentPeriodTx(ctx context.Context, tx *sql.Tx, subID int64) (*Period, error)
	// CurrentPriceTx 取該方案在指定計費週期的當期生效價（cycle: monthly | yearly，G1）。
	CurrentPriceTx(ctx context.Context, tx *sql.Tx, planID int64, cycle string) (Price, error)
	// MarkPeriodPaidTx 標記期別已付款（note 為短收／溢收註記，空字串保留原值，G8）。
	MarkPeriodPaidTx(ctx context.Context, tx *sql.Tx, id int64, paidAt time.Time,
		invoiceNo, provider, externalRef, note string) error
	PeriodsByStatus(ctx context.Context, status string) ([]Period, error)
	ActiveSubscriptionsWithDueOpenPeriod(ctx context.Context, tx *sql.Tx, now time.Time) ([]Subscription, error)
	PastDueSubscriptionsExpiredGrace(ctx context.Context, tx *sql.Tx, now time.Time) ([]Subscription, error)
	CancelledSubscriptionsPastPeriodEnd(ctx context.Context, tx *sql.Tx, now time.Time) ([]Subscription, error)
	ActiveOrTrialingSubscriptions(ctx context.Context) ([]Subscription, error)
	EmitEventTx(ctx context.Context, tx *sql.Tx, aggregateType string, aggregateID int64, eventType string, payload []byte) error
	UndispatchedEvents(ctx context.Context, limit int) ([]Event, error)
	MarkEventDispatchedTx(ctx context.Context, tx *sql.Tx, id int64) error
	RecordAuditTx(ctx context.Context, tx *sql.Tx, operatorID int64, action, targetType, targetID, reason string, before, after []byte) error
	WithTx(ctx context.Context, fn func(*sql.Tx) error) error
	SystemActor(ctx context.Context) (int64, error) // 讀 platform.settings.system_actor_user_id
	Setting(ctx context.Context, key string) (string, error)
	UpsertSettingTx(ctx context.Context, tx *sql.Tx, key, value string) error
}
```

- [ ] **Step 1: 寫 migration（`00030_platform_settings.sql`）**

```sql
-- 平台 KV 設定（系統 actor、寬限天數、催收提前天數…）：避免把營運參數寫死在程式碼。
-- +goose Up
CREATE TABLE IF NOT EXISTS platform.settings (
    key        text PRIMARY KEY,
    value      text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
REVOKE ALL ON platform.settings FROM app_rw;

-- +goose Down
DROP TABLE IF EXISTS platform.settings;
```

- [ ] **Step 2: 寫整合測試（`billing_integration_test.go`）**

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

// TestIntegrationBillingStoreTx 驗證寫入路徑的交易語意：
// ① 期別寫入與事件、稽核同一交易；回滾後三者皆不存在（不得留下孤兒事件）；
// ② OpenSubscriptionTx 會鎖住列（兩個並行交易不得同時改同一訂閱）；
// ③ settings 的讀寫與系統 actor。
func TestIntegrationBillingStoreTx(t *testing.T) {
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

	// fixture：方案 + 訂閱 + 系統 actor 設定 + operator（稽核 FK）
	var planID, subID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count)
		VALUES (42, $1, 'active', 3) RETURNING id`, planID).Scan(&subID); err != nil {
		t.Fatalf("subscription: %v", err)
	}
	var opID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.operators (email, name) VALUES ('ops@example.com','Ops') RETURNING id`).
		Scan(&opID); err != nil {
		t.Fatalf("operator: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES ('system_actor_user_id','7')`); err != nil {
		t.Fatalf("settings: %v", err)
	}

	st := postgres.New(db)

	if actor, err := st.SystemActor(ctx); err != nil || actor != 7 {
		t.Fatalf("SystemActor = %d, err=%v；want 7", actor, err)
	}

	// ① 回滾：期別／事件／稽核全部不得留下
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	sub, err := st.OpenSubscriptionTx(ctx, tx, 42)
	if err != nil {
		t.Fatalf("OpenSubscriptionTx: %v", err)
	}
	if _, err := st.OpenPeriodTx(ctx, tx, postgres.OpenPeriodInput{
		SubscriptionID: sub.ID, PeriodNo: 1,
		PeriodStart: time.Now(), PeriodEnd: time.Now().Add(720 * time.Hour),
		PlanID: planID, UnitPriceCents: 150000, SeatPriceCents: 15000, SeatCount: 3,
		AmountCents: 195000, Currency: "TWD",
	}); err != nil {
		t.Fatalf("OpenPeriodTx: %v", err)
	}
	if err := st.EmitEventTx(ctx, tx, "subscription", sub.ID, "period.opened", []byte(`{}`)); err != nil {
		t.Fatalf("EmitEventTx: %v", err)
	}
	if err := st.RecordAuditTx(ctx, tx, opID, "open_period", "subscription", "42", "測試", nil, nil); err != nil {
		t.Fatalf("RecordAuditTx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	for _, q := range []string{
		`SELECT count(*) FROM platform.subscription_periods`,
		`SELECT count(*) FROM platform.events`,
		`SELECT count(*) FROM platform.audit_logs`,
	} {
		var n int
		if err := db.QueryRow(q).Scan(&n); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		if n != 0 {
			t.Fatalf("%s 應為 0（回滾後不得留下痕跡），got %d", q, n)
		}
	}

	// ② 列鎖：另一交易改同一訂閱狀態應被擋到第一交易結束
	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	defer func() { _ = tx1.Rollback() }()
	if _, err := st.OpenSubscriptionTx(ctx, tx1, 42); err != nil {
		t.Fatalf("tx1 鎖定: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		tx2, err := db.BeginTx(ctx, nil)
		if err != nil {
			done <- err
			return
		}
		defer func() { _ = tx2.Rollback() }()
		if _, err := st.OpenSubscriptionTx(ctx, tx2, 42); err != nil {
			done <- err
			return
		}
		done <- nil
	}()
	select {
	case err := <-done:
		t.Fatalf("第二個交易不應在 tx1 持有鎖時取得列（got err=%v）", err)
	case <-time.After(300 * time.Millisecond):
		// 預期：仍被鎖住
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationBillingStoreTx -v`
Expected: FAIL —`undefined: postgres.OpenPeriodInput`／`st.OpenSubscriptionTx`

- [ ] **Step 4: 實作（`store/postgres/billing.go`）**

```go
// 平台帳務的 SQL 寫入（全部接受 *sql.Tx：期別、訂閱狀態、事件、稽核必須同一交易）。
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

type OpenPeriodInput struct {
	SubscriptionID  int64
	PeriodNo        int
	PeriodStart     time.Time
	PeriodEnd       time.Time
	PlanID          int64
	UnitPriceCents  int64
	SeatPriceCents  int64
	SeatCount       int
	AmountCents     int64
	Currency        string
	Note            string // 短收／溢收等人工註記（G8）
}

// OpenSubscriptionTx 以 FOR UPDATE 鎖住該租戶未取消的訂閱：併發的收款／逾期轉移必須互斥，
// 否則兩個請求可能都通過「目前是 active」的檢查而寫出不一致的狀態。
func (s *Store) OpenSubscriptionTx(ctx context.Context, tx *sql.Tx, companyID int) (*store.Subscription, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT s.id, s.company_id, p.code, s.status, p.id, s.seat_count, s.trial_ends_at, s.grace_until
		  FROM platform.subscriptions s
		  JOIN platform.plans p ON p.id = s.plan_id
		 WHERE s.company_id = $1 AND s.status <> 'cancelled'
		 FOR UPDATE OF s`, companyID)
	var sub store.Subscription
	var trial, grace sql.NullTime
	if err := row.Scan(&sub.ID, &sub.CompanyID, &sub.PlanCode, &sub.Status, &sub.PlanID,
		&sub.SeatCount, &trial, &grace); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if trial.Valid {
		v := trial.Time
		sub.TrialEnds = &v
	}
	if grace.Valid {
		v := grace.Time
		sub.GraceUntil = &v
	}
	return &sub, nil
}

// OpenPeriodTx 建立一期（open）。unique(subscription_id, period_no) 讓重跑安全：
// 已存在即回既有期別，不新增。
func (s *Store) OpenPeriodTx(ctx context.Context, tx *sql.Tx, in OpenPeriodInput) (*store.Period, error) {
	row := tx.QueryRowContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, currency, status)
		VALUES ($1,$2,$3,$4,$5,$6::numeric,$7::numeric,$8,$9::numeric,$10,'open')
		ON CONFLICT (subscription_id, period_no) DO NOTHING
		RETURNING id`, in.SubscriptionID, in.PeriodNo, in.PeriodStart, in.PeriodEnd, in.PlanID,
		centsToNumeric(in.UnitPriceCents), centsToNumeric(in.SeatPriceCents), in.SeatCount,
		centsToNumeric(in.AmountCents), in.Currency)
	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) { // 已存在
			return s.OpenPeriodByNoTx(ctx, tx, in.SubscriptionID, in.PeriodNo)
		}
		return nil, err
	}
	return s.periodByIDTx(ctx, tx, id)
}

func (s *Store) OpenPeriodByNoTx(ctx context.Context, tx *sql.Tx, subID int64, periodNo int) (*store.Period, error) {
	row := tx.QueryRowContext(ctx, `SELECT id FROM platform.subscription_periods
		WHERE subscription_id = $1 AND period_no = $2`, subID, periodNo)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return s.periodByIDTx(ctx, tx, id)
}

// rowScanner 抽象 *sql.Row 與 *sql.Rows（兩者 Scan 簽章相同），讓期別掃描只有一份。
type rowScanner interface{ Scan(dest ...any) error }

// periodColumns 為期別欄位順序的單一來源：多處 SELECT 若有順序漂移，
// 掃描會靜默取到錯欄位（note 的加入即為此類風險）。
const periodColumns = `id, subscription_id, period_no, period_start, period_end, plan_id,
	(unit_price*100)::bigint, (seat_price*100)::bigint, seat_count,
	(amount*100)::bigint, currency, status, paid_at, COALESCE(invoice_no,''),
	payment_provider, COALESCE(external_ref,''), note`

// scanPeriod 掃描一列期別（欄位順序見 periodColumns）。
func scanPeriod(sc rowScanner) (*store.Period, error) {
	var p store.Period
	var paid sql.NullTime
	if err := sc.Scan(&p.ID, &p.SubscriptionID, &p.PeriodNo, &p.PeriodStart, &p.PeriodEnd,
		&p.PlanID, &p.UnitPriceCents, &p.SeatPriceCents, &p.SeatCount, &p.AmountCents,
		&p.Currency, &p.Status, &paid, &p.InvoiceNo, &p.PaymentProvider, &p.ExternalRef,
		&p.Note); err != nil {
		return nil, err
	}
	if paid.Valid {
		v := paid.Time
		p.PaidAt = &v
	}
	return &p, nil
}

func (s *Store) periodByIDTx(ctx context.Context, tx *sql.Tx, id int64) (*store.Period, error) {
	return scanPeriod(tx.QueryRowContext(ctx,
		`SELECT `+periodColumns+` FROM platform.subscription_periods WHERE id = $1`, id))
}

// MarkPeriodPaidTx 標記期別為已付款；已是 paid 且交易號相同 → 視為重複入帳（no-op，回 nil）；
// 已 paid 但交易號不同 → 錯誤（避免覆蓋別筆收款）。
// note（G8）為短收／溢收的人工註記：未提供時**保留原值**（不得用空字串清掉既有註記）。
func (s *Store) MarkPeriodPaidTx(ctx context.Context, tx *sql.Tx, id int64, paidAt time.Time,
	invoiceNo, provider, externalRef, note string) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE platform.subscription_periods
		   SET status = 'paid', paid_at = $2, invoice_no = NULLIF($3,''),
		       payment_provider = $4, external_ref = NULLIF($5,''),
		       note = COALESCE(NULLIF($6,''), note)
		 WHERE id = $1 AND (
		    status = 'open'
		    OR (status = 'paid' AND COALESCE(external_ref,'') = COALESCE(NULLIF($5,''),''))
		 )`, id, paidAt, invoiceNo, provider, externalRef, note)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("期別 %d 已付款且交易號不同，拒絕覆蓋", id)
	}
	return nil
}

// SetSubscriptionStatusTx 更新訂閱狀態與寬限期。
func (s *Store) SetSubscriptionStatusTx(ctx context.Context, tx *sql.Tx, subID int64,
	status string, graceUntil *time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE platform.subscriptions
		   SET status = $2, grace_until = $3, updated_at = now(),
		       cancelled_at = CASE WHEN $2 = 'cancelled' THEN now() ELSE cancelled_at END
		 WHERE id = $1`, subID, status, graceUntil)
	return err
}

// EmitEventTx 寫入 outbox 事件。
func (s *Store) EmitEventTx(ctx context.Context, tx *sql.Tx, aggregateType string,
	aggregateID int64, eventType string, payload []byte) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.events (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1,$2,$3,$4::jsonb)`, aggregateType, aggregateID, eventType, string(payload))
	return err
}

// RecordAuditTx 寫入平台稽核（S9：actor 為 operator，不 FK 租戶 users）。
func (s *Store) RecordAuditTx(ctx context.Context, tx *sql.Tx, operatorID int64,
	action, targetType, targetID, reason string, before, after []byte) error {
	if reason == "" {
		return errors.New("平台稽核必須提供原因")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.audit_logs
			(operator_id, action, target_type, target_id, reason, before, after)
		VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::jsonb,NULLIF($7,'')::jsonb)`,
		operatorID, action, targetType, targetID, reason, stringOrEmpty(before), stringOrEmpty(after))
	return err
}

// SystemActor 讀取系統 actor 的租戶 user id（排程改公司狀態時的稽核主體）。
func (s *Store) SystemActor(ctx context.Context) (int64, error) {
	var v string
	if err := s.db.QueryRowContext(ctx,
		`SELECT value FROM platform.settings WHERE key = 'system_actor_user_id'`).Scan(&v); err != nil {
		return 0, err
	}
	var id int64
	if _, err := fmt.Sscanf(v, "%d", &id); err != nil {
		return 0, fmt.Errorf("system_actor_user_id 格式錯誤: %q", v)
	}
	return id, nil
}

func (s *Store) Setting(ctx context.Context, key string) (string, error) {
	var v string
	if err := s.db.QueryRowContext(ctx,
		`SELECT value FROM platform.settings WHERE key = $1`, key).Scan(&v); err != nil {
		return "", err
	}
	return v, nil
}

func (s *Store) UpsertSettingTx(ctx context.Context, tx *sql.Tx, key, value string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES ($1,$2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`, key, value)
	return err
}

// CurrentPriceTx 取該方案在指定計費週期的「當期生效價」（effective_from 最新者）。
// 期別金額一律用當期價快照：方案調價後新期別用新價、舊期別不變（G1 一併校正週期）。
func (s *Store) CurrentPriceTx(ctx context.Context, tx *sql.Tx, planID int64, cycle string) (store.Price, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT (base_price*100)::bigint, (seat_price*100)::bigint, currency
		  FROM platform.plan_prices
		 WHERE plan_id = $1 AND billing_cycle = $2 AND effective_from <= now()
		 ORDER BY effective_from DESC LIMIT 1`, planID, cycle)
	var p store.Price
	if err := row.Scan(&p.BaseCents, &p.SeatCents, &p.Currency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Price{}, fmt.Errorf("方案 %d 沒有 %s 週期的價目", planID, cycle)
		}
		return store.Price{}, err
	}
	return p, nil
}

// CurrentPeriodTx 取訂閱最新一期（期別產生與逾期判定使用）。
func (s *Store) CurrentPeriodTx(ctx context.Context, tx *sql.Tx, subID int64) (*store.Period, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM platform.subscription_periods
		 WHERE subscription_id = $1 ORDER BY period_no DESC LIMIT 1`, subID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s.periodByIDTx(ctx, tx, id)
}

// ActiveSubscriptionsWithDueOpenPeriod：active 且最新期別已過期末（MarkPastDue）。
func (s *Store) ActiveSubscriptionsWithDueOpenPeriod(ctx context.Context, tx *sql.Tx, now time.Time) ([]store.Subscription, error) {
	return s.scanSubscriptions(ctx, tx, `
		SELECT s.id, s.company_id, s.status
		  FROM platform.subscriptions s
		  JOIN LATERAL (
			SELECT period_end FROM platform.subscription_periods p
			 WHERE p.subscription_id = s.id ORDER BY p.period_no DESC LIMIT 1
		  ) cur ON true
		 WHERE s.status = 'active' AND cur.period_end < $1`, now)
}

// PastDueSubscriptionsExpiredGrace：past_due 且寬限已過（SuspendOverdue）。
func (s *Store) PastDueSubscriptionsExpiredGrace(ctx context.Context, tx *sql.Tx, now time.Time) ([]store.Subscription, error) {
	return s.scanSubscriptions(ctx, tx, `
		SELECT id, company_id, status FROM platform.subscriptions
		 WHERE status = 'past_due' AND grace_until IS NOT NULL AND grace_until < $1`, now)
}

// CancelledSubscriptionsPastPeriodEnd：cancelled 且最新期別已過期末（G7）。
// EXISTS 排除「已發過 subscription.expired」者 → 排程可重跑且不重複發事件。
func (s *Store) CancelledSubscriptionsPastPeriodEnd(ctx context.Context, tx *sql.Tx, now time.Time) ([]store.Subscription, error) {
	return s.scanSubscriptions(ctx, tx, `
		SELECT s.id, s.company_id, s.status
		  FROM platform.subscriptions s
		  JOIN LATERAL (
			SELECT period_end FROM platform.subscription_periods p
			 WHERE p.subscription_id = s.id ORDER BY p.period_no DESC LIMIT 1
		  ) cur ON true
		 WHERE s.status = 'cancelled' AND cur.period_end < $1
		   AND NOT EXISTS (
			SELECT 1 FROM platform.events e
			 WHERE e.aggregate_type = 'subscription' AND e.aggregate_id = s.id
			   AND e.event_type = 'subscription.expired'
		   )`, now)
}

// scanSubscriptions 為三個掃描查詢的共用列掃描（欄位形狀相同）。
func (s *Store) scanSubscriptions(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]store.Subscription, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Subscription
	for rows.Next() {
		var sub store.Subscription
		if err := rows.Scan(&sub.ID, &sub.CompanyID, &sub.Status); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

// PeriodsByStatus 取指定狀態的期別（收款找當期 open 期別；含 note）。
func (s *Store) PeriodsByStatus(ctx context.Context, status string) ([]store.Period, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, subscription_id, period_no, period_start, period_end, plan_id,
		       (unit_price*100)::bigint, (seat_price*100)::bigint, seat_count,
		       (amount*100)::bigint, currency, status, paid_at, COALESCE(invoice_no,''),
		       payment_provider, COALESCE(external_ref,''), note
		  FROM platform.subscription_periods WHERE status = $1 ORDER BY subscription_id, period_no`,
		status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Period
	for rows.Next() {
		p, err := scanPeriod(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// WithTx 開交易並把 *sql.Tx 交給 fn（平台寫入一律單一交易，見 Global Constraints）。
func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// ActiveOrTrialingSubscriptions 供排程逐租戶產生下一期（只有這兩種狀態仍在服務中）。
func (s *Store) ActiveOrTrialingSubscriptions(ctx context.Context) ([]store.Subscription, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company_id, status, plan_id, seat_count, billing_cycle
		  FROM platform.subscriptions
		 WHERE status IN ('active','trialing')
		 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Subscription
	for rows.Next() {
		var sub store.Subscription
		if err := rows.Scan(&sub.ID, &sub.CompanyID, &sub.Status,
			&sub.PlanID, &sub.SeatCount, &sub.BillingCycle); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

// UndispatchedEvents 取未派送事件（consumer 用）。
func (s *Store) UndispatchedEvents(ctx context.Context, limit int) ([]store.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload
		  FROM platform.events WHERE dispatched_at IS NULL ORDER BY id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []store.Event
	for rows.Next() {
		var e store.Event
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// MarkEventDispatchedTx 標記事件已派送；attempts 為可觀測性計數（重試次數）。
func (s *Store) MarkEventDispatchedTx(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE platform.events SET dispatched_at = now(), attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

// centsToNumeric 將分轉為 numeric(12,2) 的字串形式（金額一律經 money 套件產生，非使用者輸入）。
func centsToNumeric(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func stringOrEmpty(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return string(b)
}
```

（`UndispatchedEvents`／`MarkEventDispatchedTx`／`PeriodsByStatus`／`ListTenants` 等查詢依 Task 5／9 需要補在同檔，介面已於 Interfaces 段定義。）

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && task test:integration -- -run TestIntegrationBillingStoreTx -v`
Expected: PASS（含回滾無痕與列鎖兩條）

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/00030_platform_settings.sql backend/internal/platform/store
git commit -m "feat(backend): 平台帳務的交易版 store 寫入與 settings（00030）"
```

---

### Task 4: `RecordPayment`（收款的唯一入口）

**錯誤碼（Plan D 提供；`internal/errcode` 已可用，碼表見 `docs/error-codes.md`）**：本任務的 `FailedPrecondition` 一律換成註冊碼，落地歸屬 Plan D **Task 5b**：

| 情境 | 碼（connect 碼不變：`failed_precondition`） |
|---|---|
| 狀態機拒絕（`cancelled` 終態、無有效訂閱、找不到可收款期別、不可轉 `active`） | `errcode.PlatformSubscriptionInactive`（`PLAT-3001`） |
| 收款衝突（輸入金額與期別快照不符；其他無法 no-op 的衝突） | `errcode.PlatformPaymentConflict`（`PLAT-3002`，details `reason`） |
| 同一期別、同一交易號的重送 | **no-op（不是錯誤）** —— G3 的冪等設計，金流 webhook 重送依賴它，**不得**改成錯誤 |

**Files:**
- Create: `internal/platform/billing/billing.go`、`internal/platform/billing/billing_test.go`
- Create: `internal/platform/billing/billing_integration_test.go`

**Interfaces:**
- Consumes: `store.BillingStore`（Task 3）、`money`（Task 2）
- Produces:

```go
type Billing struct{ /* st, now */ }

func NewBilling(st store.BillingStore) *Billing

type RecordPaymentInput struct {
	CompanyID     int
	PeriodNo      int    // 0 = 當前 open 期別
	PaidAt        time.Time
	AmountCents   int64  // 0 = 採用期別快照金額
	Provider      string // manual | ecpay | newebpay | tappay | stripe
	ExternalRef   string
	InvoiceNo     string
	InvoiceStatus string
	BuyerTaxID    string
	Carrier       string
	ActorOperatorID int64
	Reason        string
}

// RecordPayment 為人工收款與（日後）金流 webhook 的共同入口：
// 標記期別已付 → 訂閱轉 active（清寬限期）→ 寫事件與稽核（同一交易）。
func (b *Billing) RecordPayment(ctx context.Context, in RecordPaymentInput) (*store.Period, error)
```

- [ ] **Step 1: 寫失敗測試（`billing_test.go`，假 store、不需容器）**

```go
package billing_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// fakeBilling 為記憶體 BillingStore：記錄交易內的每一次寫入，驗證「同一交易」的契約。
type fakeBilling struct {
	sub       *store.Subscription
	period    *store.Period
	status    string
	grace     *time.Time
	events    []string
	audits    []string
	rollback  bool
}

func (f *fakeBilling) OpenSubscriptionTx(context.Context, *sql.Tx, int) (*store.Subscription, error) {
	return f.sub, nil
}
func (f *fakeBilling) SetSubscriptionStatusTx(_ context.Context, _ *sql.Tx, _ int64, status string, grace *time.Time) error {
	f.status, f.grace = status, grace
	if f.rollback {
		return errors.New("模擬交易失敗")
	}
	return nil
}
func (f *fakeBilling) OpenPeriodTx(context.Context, *sql.Tx, store.OpenPeriodInput) (*store.Period, error) {
	return f.period, nil
}
func (f *fakeBilling) OpenPeriodByNoTx(context.Context, *sql.Tx, int64, int) (*store.Period, error) {
	return f.period, nil
}
func (f *fakeBilling) MarkPeriodPaidTx(context.Context, *sql.Tx, int64, time.Time, string, string, string) error {
	f.events = append(f.events, "period_paid")
	return nil
}
func (f *fakeBilling) PeriodsByStatus(context.Context, string) ([]store.Period, error) { return nil, nil }
func (f *fakeBilling) EmitEventTx(_ context.Context, _ *sql.Tx, _ string, _ int64, eventType string, _ []byte) error {
	f.events = append(f.events, eventType)
	return nil
}
func (f *fakeBilling) UndispatchedEvents(context.Context, int) ([]store.Event, error) { return nil, nil }
func (f *fakeBilling) MarkEventDispatchedTx(context.Context, *sql.Tx, int64) error    { return nil }
func (f *fakeBilling) RecordAuditTx(_ context.Context, _ *sql.Tx, _ int64, action, _, _, _ string, _, _ []byte) error {
	f.audits = append(f.audits, action)
	return nil
}
func (f *fakeBilling) SystemActor(context.Context) (int64, error)      { return 7, nil }
func (f *fakeBilling) Setting(context.Context, string) (string, error) { return "7", nil }
func (f *fakeBilling) UpsertSettingTx(context.Context, *sql.Tx, string, string) error {
	return nil
}

// 收款把 suspended 轉回 active、清寬限期、寫事件與稽核。
func TestRecordPaymentReactivatesSuspendedSubscription(t *testing.T) {
	grace := time.Now().Add(48 * time.Hour)
	f := &fakeBilling{
		sub:    &store.Subscription{ID: 5, CompanyID: 42, PlanCode: "std", Status: "suspended", SeatCount: 3, GraceUntil: &grace},
		period: &store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open", AmountCents: 195000},
	}
	got, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual",
		ActorOperatorID: 1, Reason: "匯款入帳（台銀 12345）",
	})
	if err != nil {
		t.Fatalf("RecordPayment: %v", err)
	}
	if got == nil || got.ID != 9 {
		t.Fatalf("應回被標記的期別,got %+v", got)
	}
	if f.status != "active" {
		t.Fatalf("訂閱應轉 active，got %q", f.status)
	}
	if f.grace != nil {
		t.Fatalf("復原後應清空寬限期，got %v", f.grace)
	}
	if !slices.Contains(f.events, "period.payment_recorded") || !slices.Contains(f.events, "subscription.reactivated") {
		t.Fatalf("應寫入收款與復原事件，got %v", f.events)
	}
	if len(f.audits) != 1 {
		t.Fatalf("應寫 1 筆平台稽核，got %v", f.audits)
	}
}

// 缺原因即拒絕：平台稽核必填，且這是人工動錢的操作。
func TestRecordPaymentRequiresReason(t *testing.T) {
	f := &fakeBilling{sub: &store.Subscription{ID: 5, CompanyID: 42, Status: "active"},
		period: &store.Period{ID: 9, Status: "open"}}
	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 1, Reason: "  ",
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("缺原因應回 invalid_argument，got %v", err)
	}
}

// 無訂閱即拒絕（不得為不存在的合約記帳）。
func TestRecordPaymentWithoutSubscription(t *testing.T) {
	f := &fakeBilling{}
	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 4242, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 1, Reason: "x",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("無訂閱應回 failed_precondition，got %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/platform/billing/ -v`
Expected: FAIL —`undefined: billing.NewBilling`

- [ ] **Step 3: 實作（`billing.go`）**

```go
// Package billing 為平台帳務的狀態機與唯一收款入口：
// 人工收款與（日後）金流 webhook 都呼叫 RecordPayment，升級只需換呼叫者。
package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// allowedTransitions 為訂閱狀態機（spec §5.2）：非法轉移一律 FailedPrecondition。
var allowedTransitions = map[string][]string{
	"trialing":  {"active", "past_due", "suspended", "cancelled"},
	"active":    {"past_due", "suspended", "cancelled"},
	"past_due":  {"active", "suspended", "cancelled"},
	"suspended": {"active", "cancelled"},
	"cancelled": {},
}

type Billing struct {
	st  store.BillingStore
	now func() time.Time
}

func NewBilling(st store.BillingStore) *Billing { return &Billing{st: st, now: time.Now} }

type RecordPaymentInput struct {
	CompanyID       int
	PeriodNo        int
	PaidAt          time.Time
	AmountCents     int64 // 0 = 採用期別快照金額；非 0 且與快照不符 → 拒絕（G4）
	Provider        string
	ExternalRef     string
	InvoiceNo       string
	InvoiceStatus   string
	BuyerTaxID      string
	Carrier         string
	Note            string // 短收／溢收等人工註記（G8）
	ActorOperatorID int64
	Reason          string
}

// RecordPayment 在單一交易內完成：鎖訂閱 → 標期別已付 → 訂閱轉 active（清寬限期）
// → 寫事件與稽核。重複入帳（同期別、同交易號）為 no-op（webhook 重送安全）。
func (b *Billing) RecordPayment(ctx context.Context, in RecordPaymentInput) (*store.Period, error) {
	if strings.TrimSpace(in.Reason) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("收款必須填寫原因/來源"))
	}
	if in.Provider == "" {
		in.Provider = "manual"
	}
	if in.PaidAt.IsZero() {
		in.PaidAt = b.now()
	}

	var out *store.Period
	err := withTx(ctx, b.st, func(tx *sql.Tx) error {
		sub, err := b.st.OpenSubscriptionTx(ctx, tx, in.CompanyID)
		if err != nil {
			return err
		}
		if sub == nil {
			return connect.NewError(connect.CodeFailedPrecondition,
				errors.New("此租戶沒有有效訂閱，無法記帳"))
		}

		var period *store.Period
		if in.PeriodNo > 0 {
			period, err = b.st.OpenPeriodByNoTx(ctx, tx, sub.ID, in.PeriodNo)
		} else {
			period, err = b.latestOpenPeriod(ctx, sub.ID)
		}
		if err != nil {
			return err
		}
		if period == nil {
			return connect.NewError(connect.CodeFailedPrecondition,
				errors.New("找不到可收款的期別（請先產生期別）"))
		}
		// 重複入帳判定（G3）：期別已 paid → 一律 no-op。
		// 為什麼不看交易號：人工收款多數沒有交易號，若要求 external_ref 非空才 no-op，
		// 空交易號的重送會再寫一次 period.payment_recorded 事件與稽核（帳面與事件流失真）。
		// 注意：**這裡刻意不回 PLAT-3002** —— 重送是冪等（不是衝突），webhook 重送依賴它。
		if period.Status == "paid" {
			if in.ExternalRef != period.ExternalRef {
				log.Printf("platform payment: 期別 %d 已付款，本次交易號 %q 與原 %q 不同（可能為溢收，請人工確認）",
					period.ID, in.ExternalRef, period.ExternalRef)
			}
			out = period
			return nil
		}

		// 金額驗證（G4）：未填 → 採期別快照；填了但與快照不符 → 拒絕。
		// v1 不支援部分付款：短收／溢收以 note 記錄，不改變期別金額。
		if in.AmountCents != 0 && in.AmountCents != period.AmountCents {
			return errcode.PlatformPaymentConflict.Error(map[string]string{
				"reason": fmt.Sprintf("輸入金額 %s 與期別金額 %s 不符（不支援部分付款；差異請記於備註）",
					money.FormatCents(in.AmountCents), money.FormatCents(period.AmountCents)),
			})
		}

		if err := b.st.MarkPeriodPaidTx(ctx, tx, period.ID, in.PaidAt,
			in.InvoiceNo, in.Provider, in.ExternalRef, in.Note); err != nil {
			return connect.NewError(connect.CodeFailedPrecondition, err)
		}

		// 訂閱復原：清寬限期、重設催收次數。
		if sub.Status != "active" {
			if !canTransition(sub.Status, "active") {
				return connect.NewError(connect.CodeFailedPrecondition,
					fmt.Errorf("訂閱狀態 %s 不可轉為 active", sub.Status))
			}
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "active", nil); err != nil {
				return err
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.reactivated", map[string]any{
				"company_id": in.CompanyID, "from": sub.Status,
			}); err != nil {
				return err
			}
		}
		if err := b.emit(ctx, tx, sub.ID, "period.payment_recorded", map[string]any{
			"company_id": in.CompanyID, "period_no": period.PeriodNo,
			"amount_cents": period.AmountCents, "provider": in.Provider,
		}); err != nil {
			return err
		}

		after, _ := json.Marshal(map[string]any{
			"period_no": period.PeriodNo, "paid_at": in.PaidAt.UTC().Format(time.RFC3339),
			"provider": in.Provider, "external_ref": in.ExternalRef, "amount": period.AmountCents,
		})
		if err := b.st.RecordAuditTx(ctx, tx, in.ActorOperatorID, "record_payment",
			"subscription", fmt.Sprintf("%d", sub.ID), in.Reason, nil, after); err != nil {
			return err
		}
		out = period
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// latestOpenPeriod 取最新一期 open；無則回 (nil, nil)。
func (b *Billing) latestOpenPeriod(ctx context.Context, subID int64) (*store.Period, error) {
	rows, err := b.st.PeriodsByStatus(ctx, "open")
	if err != nil {
		return nil, err
	}
	var best *store.Period
	for i := range rows {
		if rows[i].SubscriptionID != subID {
			continue
		}
		if best == nil || rows[i].PeriodNo > best.PeriodNo {
			best = &rows[i]
		}
	}
	return best, nil
}

func (b *Billing) emit(ctx context.Context, tx *sql.Tx, subID int64, eventType string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return b.st.EmitEventTx(ctx, tx, "subscription", subID, eventType, raw)
}

func canTransition(from, to string) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// withTx 開交易並把 *sql.Tx 交給 fn（平台 store 的寫入介面一律接受 tx）。
func withTx(ctx context.Context, st store.BillingStore, fn func(*sql.Tx) error) error {
	return st.WithTx(ctx, fn)
}
```

（`store.BillingStore` 需追加 `WithTx(ctx, fn) error`：以 `s.db.BeginTx` 實作；假實作直接呼叫 `fn(nil)` 並可用旗標模擬失敗。）

- [ ] **Step 4: 補狀態機的非法轉移測試（同檔）**

```go
// 非法轉移一律 FailedPrecondition：cancelled 是終態（不得被收款偷偷復活）。
func TestRecordPaymentRejectsCancelledSubscription(t *testing.T) {
	f := &fakeBilling{
		sub:    &store.Subscription{ID: 5, CompanyID: 42, Status: "cancelled"},
		period: &store.Period{ID: 9, Status: "paid", ExternalRef: ""},
	}
	_, err := billing.NewBilling(f).RecordPayment(context.Background(), billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ActorOperatorID: 1, Reason: "誤匯款",
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("cancelled 不得復原，got %v", err)
	}
}
```

（`fakeBilling` 的 `MarkPeriodPaidTx` 需回傳可設定錯誤以便此案例成立；實作時讓 fake 有 `markPaidErr error` 欄位。`OpenSubscriptionTx` 對 `cancelled` 訂閱回 `nil`（SQL 條件已排除），故本案例的錯誤來自「無訂閱」路徑——測試斷言碼相同，但請在註解寫明是哪一條路徑，避免誤解。）

- [ ] **Step 5: 整合測試（真 PG：金額寫入、狀態轉移、事件與稽核同一交易）**

```go
//go:build integration

package billing_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRecordPayment 在真 PostgreSQL 上完成一次入帳並驗證四件事：
// ① 期別轉 paid 且 paid_at／交易號落地；
// ② 訂閱由 suspended 轉 active、寬限期清空；
// ③ 事件（subscription.reactivated＋period.payment_recorded）與 1 筆平台稽核都寫入；
// ④ 同交易號重送 → 不再增加事件與稽核（webhook 重送安全）。
func TestIntegrationRecordPayment(t *testing.T) {
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
	ctx := t.Context()

	var planID, subID, periodID, opID int64
	if err := db.QueryRowContext(ctx,
		`INSERT INTO platform.plans (code, name) VALUES ('std','標準') RETURNING id`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	grace := time.Now().Add(48 * time.Hour)
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscriptions (company_id, plan_id, status, seat_count, grace_until)
		VALUES (42, $1, 'suspended', 3, $2) RETURNING id`, planID, grace).Scan(&subID); err != nil {
		t.Fatalf("subscription: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.subscription_periods
			(subscription_id, period_no, period_start, period_end, plan_id,
			 unit_price, seat_price, seat_count, amount, currency, status)
		VALUES ($1, 1, now(), now() + interval '30 days', $2, 1500.00, 150.00, 3, 1950.00, 'TWD', 'open')
		RETURNING id`, subID, planID).Scan(&periodID); err != nil {
		t.Fatalf("period: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.operators (email, name) VALUES ('ops@example.com','Ops') RETURNING id`).
		Scan(&opID); err != nil {
		t.Fatalf("operator: %v", err)
	}
	b := billing.NewBilling(postgres.New(db))
	in := billing.RecordPaymentInput{
		CompanyID: 42, PaidAt: time.Now(), Provider: "manual", ExternalRef: "BANK-12345",
		InvoiceNo: "AB12345678", ActorOperatorID: opID, Reason: "匯款入帳（台銀 12345）",
	}

	if _, err := b.RecordPayment(ctx, in); err != nil {
		t.Fatalf("第一次收款: %v", err)
	}

	// ② 訂閱轉 active 且清空寬限期
	var subStatus string
	var grace2 sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT status, grace_until FROM platform.subscriptions WHERE id = $1`, subID).
		Scan(&subStatus, &grace2); err != nil {
		t.Fatalf("查訂閱: %v", err)
	}
	if subStatus != "active" {
		t.Fatalf("訂閱應轉 active，got %s", subStatus)
	}
	if grace2.Valid {
		t.Fatalf("復原後應清空寬限期，got %v", grace2.Time)
	}

	// ① 期別轉 paid 且 paid_at／交易號落地
	var pStatus, externalRef string
	var paidAt sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT status, paid_at, COALESCE(external_ref,'') FROM platform.subscription_periods WHERE id = $1`,
		periodID).Scan(&pStatus, &paidAt, &externalRef); err != nil {
		t.Fatalf("查期別: %v", err)
	}
	if pStatus != "paid" || !paidAt.Valid {
		t.Fatalf("期別應為 paid 且有 paid_at，got status=%s paid_at=%v", pStatus, paidAt)
	}
	if externalRef != "BANK-12345" {
		t.Fatalf("交易號應落地為 BANK-12345，got %q", externalRef)
	}

	// ③ 事件與稽核
	var reactivated, recorded, audits int
	if err := db.QueryRowContext(ctx, `
		SELECT
		 (SELECT count(*) FROM platform.events WHERE event_type = 'subscription.reactivated'),
		 (SELECT count(*) FROM platform.events WHERE event_type = 'period.payment_recorded'),
		 (SELECT count(*) FROM platform.audit_logs WHERE action = 'record_payment')`).
		Scan(&reactivated, &recorded, &audits); err != nil {
		t.Fatalf("查事件與稽核: %v", err)
	}
	if reactivated != 1 || recorded != 1 || audits != 1 {
		t.Fatalf("事件/稽核計數不符: reactivated=%d recorded=%d audits=%d", reactivated, recorded, audits)
	}

	// ④ 同交易號重送：期別已 paid 且 external_ref 相同 → no-op
	if _, err := b.RecordPayment(ctx, in); err != nil {
		t.Fatalf("重送收款: %v", err)
	}
	var recorded2, audits2 int
	if err := db.QueryRowContext(ctx, `
		SELECT
		 (SELECT count(*) FROM platform.events WHERE event_type = 'period.payment_recorded'),
		 (SELECT count(*) FROM platform.audit_logs WHERE action = 'record_payment')`).
		Scan(&recorded2, &audits2); err != nil {
		t.Fatalf("重送後查計數: %v", err)
	}
	if recorded2 != 1 || audits2 != 1 {
		t.Fatalf("重送不得重複入帳: recorded=%d audits=%d（應維持 1/1）", recorded2, audits2)
	}
}
```

- [ ] **Step 6: 跑測試**

Run: `cd backend && go test ./internal/platform/billing/ -v && task test:integration -- -run TestIntegrationRecordPayment -v`
Expected: 全 PASS（Skip 已移除）

- [ ] **Step 7: Commit**

```bash
git add backend/internal/platform/billing
git commit -m "feat(backend): RecordPayment 為唯一收款入口（狀態機、事件、稽核同一交易、冪等）"
```

---

### Task 5: 生命週期排程邏輯（`EnsureNextPeriod`／`MarkPastDue`／`SuspendOverdue`）

**錯誤碼**：本任務的非法狀態轉移一律 `PLAT-3001` `PlatformSubscriptionInactive`（承 Task 4 開頭的錯誤碼表；不得自建裸 `FailedPrecondition`）。

**Files:**
- Create: `internal/platform/billing/lifecycle.go`、`internal/platform/billing/lifecycle_test.go`

**Interfaces:**
- Produces:

```go
// EnsureNextPeriod 於到期前 leadDays 天內建立下一期（open，價格快照為當期生效價）。
// 已有下一期即 no-op（回 false）。
func (b *Billing) EnsureNextPeriod(ctx context.Context, companyID int, now time.Time, leadDays int) (bool, error)

// MarkPastDue 掃描 active 且已過 period_end 的訂閱 → past_due，並設 grace_until = now + graceDays。
func (b *Billing) MarkPastDue(ctx context.Context, now time.Time, graceDays int) (int, error)

// SuspendOverdue 掃描 past_due 且 grace_until < now 的訂閱 → suspended（發事件驅動凍結）。
func (b *Billing) SuspendOverdue(ctx context.Context, now time.Time) (int, error)
```

- [ ] **Step 1: 擴充假 store 並寫失敗測試（`lifecycle_test.go`）**

先在同一測試套件的 `fakeBilling`（`billing_test.go`）追加生命週期所需的欄位與方法：

```go
// Task 5 生命週期測試用欄位（追加到 fakeBilling）
type fakeBilling struct {
	// … Task 4 既有欄位 …
	due        []store.Subscription // MarkPastDue 的掃描結果
	expired    []store.Subscription // SuspendOverdue 的掃描結果
	curPeriod  *store.Period
	nextExists bool  // 是否已有下一期（驗證冪等）
	price      store.Price
	opened     []store.OpenPeriodInput
}

func (f *fakeBilling) ActiveSubscriptionsWithDueOpenPeriod(context.Context, *sql.Tx, time.Time) ([]store.Subscription, error) {
	return f.due, nil
}

func (f *fakeBilling) PastDueSubscriptionsExpiredGrace(context.Context, *sql.Tx, time.Time) ([]store.Subscription, error) {
	return f.expired, nil
}

func (f *fakeBilling) CurrentPeriodTx(context.Context, *sql.Tx, int64) (*store.Period, error) {
	return f.curPeriod, nil
}

func (f *fakeBilling) CurrentPriceTx(context.Context, *sql.Tx, int64) (store.Price, error) {
	return f.price, nil
}

// OpenPeriodByNoTx 覆寫 Task 4 版本以支援冪等測試：已存在回既有期別、否則回 sql.ErrNoRows。
func (f *fakeBilling) OpenPeriodByNoTxTx(ctx context.Context, tx *sql.Tx, subID int64, no int) (*store.Period, error) {
	if f.nextExists {
		return &store.Period{SubscriptionID: subID, PeriodNo: no}, nil
	}
	return nil, sql.ErrNoRows
}
```

（實作時把 `OpenPeriodByNoTxTx` 命名與介面統一為 `OpenPeriodByNoTx` 並覆蓋 Task 4 的版本，避免兩個方法並存。）

測試本體：

```go
package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// 逾期轉移：active 且期末已過 → past_due ＋ 設寬限期；重跑（掃描已無此列）不重複發事件。
func TestMarkPastDueSetsGraceAndEmitsOnce(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	f := &fakeBilling{due: []store.Subscription{
		{ID: 5, CompanyID: 42, Status: "active"},
	}}
	b := billing.NewBilling(f)

	n, err := b.MarkPastDue(context.Background(), now, 7)
	if err != nil || n != 1 {
		t.Fatalf("應轉移 1 筆: n=%d err=%v", n, err)
	}
	if f.status != "past_due" {
		t.Fatalf("狀態應為 past_due，got %q", f.status)
	}
	want := now.AddDate(0, 0, 7)
	if f.grace == nil || !f.grace.Equal(want) {
		t.Fatalf("寬限期應為 %s，got %v", want, f.grace)
	}
	if !slices.Contains(f.events, "subscription.past_due") {
		t.Fatalf("應發 subscription.past_due，got %v", f.events)
	}

	// 重跑：SQL 條件已排除 past_due，掃描結果為空 → 不得再發事件
	eventsBefore := len(f.events)
	f.due = nil
	n2, err := b.MarkPastDue(context.Background(), now, 7)
	if err != nil || n2 != 0 {
		t.Fatalf("重跑不應轉移: n=%d err=%v", n2, err)
	}
	if len(f.events) != eventsBefore {
		t.Fatalf("重跑不得重複發事件，got %v", f.events)
	}
}

// 寬限已過 → suspended 並發事件（此事件驅動產品域凍結）；不得殘留寬限期。
func TestSuspendOverdueEmitsEvent(t *testing.T) {
	now := time.Date(2026, 10, 10, 3, 0, 0, 0, time.UTC)
	grace := now.Add(-time.Hour)
	f := &fakeBilling{
		expired: []store.Subscription{{ID: 5, CompanyID: 42, Status: "past_due", GraceUntil: &grace}},
	}
	b := billing.NewBilling(f)

	n, err := b.SuspendOverdue(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應停用 1 筆: n=%d err=%v", n, err)
	}
	if f.status != "suspended" || f.grace != nil {
		t.Fatalf("狀態應為 suspended 且清空寬限期，got status=%q grace=%v", f.status, f.grace)
	}
	if !slices.Contains(f.events, "subscription.suspended") {
		t.Fatalf("應發 subscription.suspended，got %v", f.events)
	}
}

// 產生期別：已有下一期即 no-op（回 false），且不得再寫入期別。
func TestEnsureNextPeriodIsIdempotent(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	f := &fakeBilling{
		sub:       &store.Subscription{ID: 5, CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3},
		curPeriod: &store.Period{SubscriptionID: 5, PeriodNo: 1, PeriodEnd: now.Add(48 * time.Hour)},
		nextExists: true,
	}
	b := billing.NewBilling(f)

	created, err := b.EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil {
		t.Fatalf("EnsureNextPeriod: %v", err)
	}
	if created {
		t.Fatal("已有下一期時應回 false")
	}
	if len(f.opened) != 0 {
		t.Fatalf("不得再寫入期別，got %d 筆", len(f.opened))
	}
}

// 價格快照：新期別取「當期生效價」，不得沿用舊期別金額。
func TestEnsureNextPeriodSnapshotUsesCurrentPrice(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	f := &fakeBilling{
		sub:       &store.Subscription{ID: 5, CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3},
		curPeriod: &store.Period{SubscriptionID: 5, PeriodNo: 1, PeriodEnd: now.Add(48 * time.Hour)},
		price:     store.Price{BaseCents: 180000, SeatCents: 20000, Currency: "TWD"}, // 調價後
	}
	b := billing.NewBilling(f)

	created, err := b.EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	if len(f.opened) != 1 {
		t.Fatalf("應寫入 1 筆期別，got %d", len(f.opened))
	}
	got := f.opened[0]
	if got.AmountCents != 240000 { // 1800 + 200×3
		t.Fatalf("新期別金額應為 240000（當期生效價快照），got %d", got.AmountCents)
	}
	if got.UnitPriceCents != 180000 || got.SeatPriceCents != 20000 {
		t.Fatalf("價格快照未更新: %+v", got)
	}
	if got.PeriodNo != 2 {
		t.Fatalf("期別號應遞增為 2，got %d", got.PeriodNo)
	}
}

// G1：年繳訂閱產生下一期時，期末必須加一年（不是一個月）。
func TestEnsureNextPeriodUsesBillingCycle(t *testing.T) {
	end := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 10, 25, 3, 0, 0, 0, time.UTC) // 在 leadDays=14 的提前窗內
	f := &fakeBilling{
		sub: &store.Subscription{ID: 5, CompanyID: 42, Status: "active",
			PlanID: 1, SeatCount: 3, BillingCycle: "yearly"},
		curPeriod: &store.Period{SubscriptionID: 5, PeriodNo: 1, PeriodEnd: end},
		price:     store.Price{BaseCents: 1620000, SeatCents: 162000, Currency: "TWD"},
	}
	created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	wantEnd := time.Date(2027, 10, 31, 0, 0, 0, 0, time.UTC)
	if got := f.opened[0].PeriodEnd; !got.Equal(wantEnd) {
		t.Fatalf("年繳期末應為 %s（+1 年），got %s", wantEnd, got)
	}
	if got := f.opened[0].PeriodStart; !got.Equal(end) {
		t.Fatalf("新期別起日應接續前期期末 %s，got %s", end, got)
	}
}
```

**以下兩條置於獨立檔案 `internal/platform/billing/lifecycle_internal_test.go`（`package billing`；`addBillingPeriod` 未匯出，外部測試套件取用不到）**

```go
// G2：月底不能靠 time.AddDate（1/31 + 1 月會正規化成 3/3，帳期跳過整個 2 月）。
func TestAddBillingPeriodHandlesMonthEnd(t *testing.T) {
	cases := []struct {
		from, cycle, want string
	}{
		{"2026-01-31T00:00:00Z", "monthly", "2026-02-28T00:00:00Z"},
		{"2026-03-31T00:00:00Z", "monthly", "2026-04-30T00:00:00Z"},
		{"2028-01-31T00:00:00Z", "monthly", "2028-02-29T00:00:00Z"}, // 閏年
		{"2026-01-15T00:00:00Z", "monthly", "2026-02-15T00:00:00Z"},
		{"2026-12-15T00:00:00Z", "monthly", "2027-01-15T00:00:00Z"}, // 跨年
		{"2026-02-28T00:00:00Z", "yearly", "2027-02-28T00:00:00Z"},
		{"2028-02-29T00:00:00Z", "yearly", "2029-02-28T00:00:00Z"}, // 閏日遇平年
	}
	for _, tc := range cases {
		from, err := time.Parse(time.RFC3339, tc.from)
		if err != nil {
			t.Fatalf("解析 %s: %v", tc.from, err)
		}
		got, err := addBillingPeriod(from, tc.cycle)
		if err != nil {
			t.Fatalf("addBillingPeriod(%s, %s): %v", tc.from, tc.cycle, err)
		}
		want, _ := time.Parse(time.RFC3339, tc.want)
		if !got.Equal(want) {
			t.Fatalf("addBillingPeriod(%s, %s) = %s；want %s",
				tc.from, tc.cycle, got.Format(time.RFC3339), tc.want)
		}
	}
	if _, err := addBillingPeriod(time.Now(), "weekly"); err == nil {
		t.Fatal("未知計費週期應報錯（不得默默當成月繳）")
	}
}
```

**回到 `lifecycle_test.go`（`package billing_test`）**

```go
// G7：cancelled 且期末已過 → 發 subscription.expired（由 consumer 把公司轉 suspended）；
// **不得改變訂閱狀態**（cancelled 是終態，改狀態會破壞帳與稽核的可重現性）。
func TestExpireCancelledEmitsEventWithoutChangingStatus(t *testing.T) {
	now := time.Date(2026, 11, 1, 3, 0, 0, 0, time.UTC)
	f := &fakeBilling{
		expiredCancelled: []store.Subscription{{ID: 5, CompanyID: 42, Status: "cancelled"}},
	}
	n, err := billing.NewBilling(f).ExpireCancelled(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應處理 1 筆: n=%d err=%v", n, err)
	}
	if !slices.Contains(f.events, "subscription.expired") {
		t.Fatalf("應發 subscription.expired，got %v", f.events)
	}
	if f.status != "" {
		t.Fatalf("不得改變訂閱狀態，got %q", f.status)
	}

	// 重跑：SQL 條件已排除「已發過 expired 事件」者 → 計數 0
	f.expiredCancelled = nil
	if n2, err := billing.NewBilling(f).ExpireCancelled(context.Background(), now); err != nil || n2 != 0 {
		t.Fatalf("重跑不應再處理: n=%d err=%v", n2, err)
	}
}
```

（`fakeBilling` 追加 `expiredCancelled []store.Subscription` 欄位與 `CancelledSubscriptionsPastPeriodEnd` 方法（回傳它），以及 `CurrentPriceTx(ctx, tx, planID int64, cycle string)` 的簽章對齊。）

- [ ] **Step 2: 實作（`lifecycle.go`）**

```go
// 訂閱生命週期的掃描式轉移（由 cmd/platform-cron 單趟呼叫）。
// 每個函式都必須可重跑：以「當前狀態 ＋ 時間」判斷，不依賴呼叫次數。
package billing

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"errors"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// EnsureNextPeriod 在到期前 leadDays 天建立下一期（open）；已存在即回 false。
// 期別長度依訂閱的 billing_cycle（月繳 +1 月、年繳 +1 年），且以 addBillingPeriod
// 處理月底（Go 的 AddDate 會把 1/31 正規化成 3/3，帳期會跳過整個 2 月 —— G1/G2）。
func (b *Billing) EnsureNextPeriod(ctx context.Context, companyID int, now time.Time, leadDays int) (bool, error) {
	created := false
	err := withTx(ctx, b.st, func(tx *sql.Tx) error {
		sub, err := b.st.OpenSubscriptionTx(ctx, tx, companyID)
		if err != nil || sub == nil {
			return err
		}
		if sub.Status != "active" && sub.Status != "trialing" {
			return nil
		}
		cur, err := b.currentPeriod(ctx, tx, sub.ID)
		if err != nil || cur == nil {
			return err
		}
		if now.Before(cur.PeriodEnd.AddDate(0, 0, -leadDays)) {
			return nil // 尚未進入提前建立窗
		}
		if _, err := b.st.OpenPeriodByNoTx(ctx, tx, sub.ID, cur.PeriodNo+1); err == nil {
			return nil // 已有下一期
		}
		price, err := b.st.CurrentPriceTx(ctx, tx, sub.PlanID, sub.BillingCycle)
		if err != nil {
			return connect.NewError(connect.CodeFailedPrecondition, err)
		}
		amount, err := money.PeriodAmount(price.BaseCents, price.SeatCents, sub.SeatCount)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		nextEnd, err := addBillingPeriod(cur.PeriodEnd, sub.BillingCycle)
		if err != nil {
			return connect.NewError(connect.CodeFailedPrecondition, err)
		}
		if _, err := b.st.OpenPeriodTx(ctx, tx, store.OpenPeriodInput{
			SubscriptionID: sub.ID, PeriodNo: cur.PeriodNo + 1,
			PeriodStart: cur.PeriodEnd, PeriodEnd: nextEnd,
			PlanID: sub.PlanID, UnitPriceCents: price.BaseCents, SeatPriceCents: price.SeatCents,
			SeatCount: sub.SeatCount, AmountCents: amount, Currency: price.Currency,
		}); err != nil {
			return err
		}
		if err := b.emit(ctx, tx, sub.ID, "period.opened", map[string]any{
			"company_id": companyID, "period_no": cur.PeriodNo + 1,
			"amount_cents": amount, "billing_cycle": sub.BillingCycle,
		}); err != nil {
			return err
		}
		created = true
		return nil
	})
	return created, err
}

// addBillingPeriod 由期別起算下一個期末：月繳取「下月同日」、年繳取「明年同日」；
// 該日不存在（1/31、3/31、閏年 2/29）時取當月最後一日。
// 為何不用 time.AddDate：它會正規化（1/31 + 1 月 = 3/3），帳期會跳過整個 2 月。
func addBillingPeriod(from time.Time, cycle string) (time.Time, error) {
	switch cycle {
	case "monthly":
		return dayOfMonthOrLast(from, from.Year(), int(from.Month())+1)
	case "yearly":
		return dayOfMonthOrLast(from, from.Year()+1, int(from.Month()))
	default:
		return time.Time{}, fmt.Errorf("未知的計費週期 %q（允許 monthly / yearly）", cycle)
	}
}

// dayOfMonthOrLast 回傳「year-month 的同一天」；該日不存在時回該月最後一天。
func dayOfMonthOrLast(from time.Time, year, month int) (time.Time, error) {
	if month > 12 {
		year, month = year+1, month-12
	}
	loc := from.Location()
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, loc).Day()
	day := from.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, time.Month(month), day,
		from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), loc), nil
}

// MarkPastDue 掃描 active 且期末已過的訂閱 → past_due ＋ 寬限期；回傳受影響筆數。
func (b *Billing) MarkPastDue(ctx context.Context, now time.Time, graceDays int) (int, error) {
	n := 0
	err := withTx(ctx, b.st, func(tx *sql.Tx) error {
		overdue, err := b.st.ActiveSubscriptionsWithDueOpenPeriod(ctx, tx, now)
		if err != nil {
			return err
		}
		grace := now.AddDate(0, 0, graceDays)
		for _, sub := range overdue {
			if !canTransition(sub.Status, "past_due") {
				continue
			}
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "past_due", &grace); err != nil {
				return err
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.past_due", map[string]any{
				"company_id": sub.CompanyID, "grace_until": grace.UTC().Format(time.RFC3339),
			}); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	return n, err
}

// SuspendOverdue 掃描 past_due 且寬限已過的訂閱 → suspended ＋ 發事件（驅動產品域凍結）。
func (b *Billing) SuspendOverdue(ctx context.Context, now time.Time) (int, error) {
	n := 0
	err := withTx(ctx, b.st, func(tx *sql.Tx) error {
		due, err := b.st.PastDueSubscriptionsExpiredGrace(ctx, tx, now)
		if err != nil {
			return err
		}
		for _, sub := range due {
			if !canTransition(sub.Status, "suspended") {
				continue
			}
			if err := b.st.SetSubscriptionStatusTx(ctx, tx, sub.ID, "suspended", nil); err != nil {
				return err
			}
			if err := b.emit(ctx, tx, sub.ID, "subscription.suspended", map[string]any{
				"company_id": sub.CompanyID, "reason": "overdue",
			}); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	return n, err
}

// ExpireCancelled 掃描 cancelled 且期末已過的訂閱 → 發 subscription.expired（G7）。
// 為何需要：MarkPastDue 只掃 active，已取消的租戶期滿後會一直可用；
// 取消是「期末終止」，故期末前仍提供服務、期末後才停止。
// 不改變訂閱狀態（cancelled 是終態）：只發事件，由 consumer 把公司轉為 suspended。
func (b *Billing) ExpireCancelled(ctx context.Context, now time.Time) (int, error) {
	n := 0
	err := withTx(ctx, b.st, func(tx *sql.Tx) error {
		expired, err := b.st.CancelledSubscriptionsPastPeriodEnd(ctx, tx, now)
		if err != nil {
			return err
		}
		for _, sub := range expired {
			if err := b.emit(ctx, tx, sub.ID, "subscription.expired", map[string]any{
				"company_id": sub.CompanyID, "reason": "cancelled_at_period_end",
			}); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	return n, err
}
```

（`store.BillingStore` 追加四個查詢：`CurrentPeriodTx`、`CurrentPriceTx(ctx, tx, planID, cycle)`、`ActiveSubscriptionsWithDueOpenPeriod`、`PastDueSubscriptionsExpiredGrace`、`CancelledSubscriptionsPastPeriodEnd`；`errors` import 若未用到即移除。）

- [ ] **Step 3: 補齊測試至全綠並可重跑**

Run: `cd backend && go test ./internal/platform/billing/ -v`
Expected: PASS（含「重跑不重複發事件」：第二次呼叫因狀態已非 active 而不進入轉移）

- [ ] **Step 4: Commit**

```bash
git add backend/internal/platform/billing
git commit -m "feat(backend): 訂閱生命週期轉移（產生期別、逾期、寬限後停用，皆可重跑）"
```

---

### Task 6: outbox consumer（事件驅動的租戶凍結／復原）

**Files:**
- Create: `internal/platform/consumer/consumer.go`、`internal/platform/consumer/consumer_test.go`

**Interfaces:**
- Consumes: 產品域 `SetCompanyStatus`；**窄介面**（consumer 不該依賴整個 `BillingStore`）
- Produces:

```go
// Store 為 consumer 實際需要的三個能力（比 store.BillingStore 窄，假實作因此只要三支方法）。
type Store interface {
	UndispatchedEvents(ctx context.Context, limit int) ([]store.Event, error)
	MarkEventDispatchedTx(ctx context.Context, tx *sql.Tx, eventID int64) error
	WithTx(ctx context.Context, fn func(*sql.Tx) error) error
}

// CompanyStatusSetter 由產品域提供（於 server 組裝與 cron 注入）。
type CompanyStatusSetter interface {
	SetStatus(ctx context.Context, companyID int, status string, reason string, actor authz.Identity) error
}

type Consumer struct{ /* st, setter, systemActor */ }

// DispatchOnce 派送未處理事件；每筆與其副作用同一交易，成功即標記已派送（可重跑）。
func (c *Consumer) DispatchOnce(ctx context.Context, limit int) (int, error)
```

**為何同一交易可行**：`platform` schema 與業務表在**同一個 PostgreSQL 資料庫**，admin 連線可同時寫兩者；因此「改 `companies.status` ＋ 標記事件已派送」是原子操作，不需要補償。

- [ ] **Step 1: 寫失敗測試（`consumer_test.go`，三支方法的假 store）**

```go
package consumer_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// fakeDispatchStore 只實作 consumer 需要的三支方法。
type fakeDispatchStore struct {
	events  []store.Event
	marked  map[int64]bool
	txError error // 模擬產品域寫入失敗
}

func (f *fakeDispatchStore) UndispatchedEvents(context.Context, int) ([]store.Event, error) {
	return f.events, nil
}

func (f *fakeDispatchStore) MarkEventDispatchedTx(_ context.Context, _ *sql.Tx, id int64) error {
	f.marked[id] = true
	return nil
}

func (f *fakeDispatchStore) WithTx(_ context.Context, fn func(*sql.Tx) error) error {
	if f.txError != nil {
		return f.txError
	}
	return fn(nil)
}

// fakeSetter 記錄收到的狀態變更呼叫。
type fakeSetter struct {
	calls []struct {
		CompanyID int
		Status    string
		Reason    string
	}
	err error
}

func (f *fakeSetter) SetStatus(_ context.Context, companyID int, status, reason string, _ authz.Identity) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, struct {
		CompanyID int
		Status    string
		Reason    string
	}{companyID, status, reason})
	return nil
}

func newConsumer(st *fakeDispatchStore, setter *fakeSetter) *consumer.Consumer {
	return consumer.New(st, setter, authz.Identity{UserID: "7", Role: "super", Roles: []string{"super"}})
}

// ① suspended → 凍結該公司，並標記事件已派送。
func TestDispatchSuspendedFreezesCompany(t *testing.T) {
	st := &fakeDispatchStore{
		events: []store.Event{{ID: 1, EventType: "subscription.suspended",
			Payload: []byte(`{"company_id":42}`)}},
		marked: map[int64]bool{},
	}
	setter := &fakeSetter{}
	n, err := newConsumer(st, setter).DispatchOnce(context.Background(), 100)
	if err != nil || n != 1 {
		t.Fatalf("應派送 1 筆: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 {
		t.Fatalf("應呼叫 setter 一次，got %d", len(setter.calls))
	}
	got := setter.calls[0]
	if got.CompanyID != 42 || got.Status != "suspended" {
		t.Fatalf("凍結呼叫參數錯誤: %+v", got)
	}
	if !strings.Contains(got.Reason, "欠費") {
		t.Fatalf("凍結原因應說明欠費，got %q", got.Reason)
	}
	if !st.marked[1] {
		t.Fatal("事件 1 應標記為已派送")
	}
}

// ② reactivated → 復原為 active。
func TestDispatchReactivatedUnfreezesCompany(t *testing.T) {
	st := &fakeDispatchStore{
		events: []store.Event{{ID: 2, EventType: "subscription.reactivated",
			Payload: []byte(`{"company_id":42}`)}},
		marked: map[int64]bool{},
	}
	setter := &fakeSetter{}
	if _, err := newConsumer(st, setter).DispatchOnce(context.Background(), 100); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(setter.calls) != 1 || setter.calls[0].Status != "active" {
		t.Fatalf("應以 active 復原，got %+v", setter.calls)
	}
	if !st.marked[2] {
		t.Fatal("事件 2 應標記為已派送")
	}
}

// ③ 產品域寫入失敗 → 不標記已派送（下趟重跑會再試），錯誤往外傳。
func TestDispatchFailureKeepsEventUndispatched(t *testing.T) {
	st := &fakeDispatchStore{
		events: []store.Event{{ID: 3, EventType: "subscription.suspended",
			Payload: []byte(`{"company_id":42}`)}},
		marked:  map[int64]bool{},
		txError: errors.New("模擬交易失敗"),
	}
	setter := &fakeSetter{}
	if _, err := newConsumer(st, setter).DispatchOnce(context.Background(), 100); err == nil {
		t.Fatal("失敗應往外傳")
	}
	if st.marked[3] {
		t.Fatal("失敗的事件不得標記為已派送")
	}
}

// ④ 未知事件型別 → 標記已派送（避免無窮重試）且不呼叫 setter。
func TestDispatchUnknownEventTypeMarksDispatchedWithoutSideEffect(t *testing.T) {
	st := &fakeDispatchStore{
		events: []store.Event{{ID: 4, EventType: "period.opened", Payload: []byte(`{}`)}},
		marked: map[int64]bool{},
	}
	setter := &fakeSetter{}
	if _, err := newConsumer(st, setter).DispatchOnce(context.Background(), 100); err != nil {
		t.Fatalf("未知型別不應報錯（僅標記）: %v", err)
	}
	if !st.marked[4] {
		t.Fatal("未知型別應標記為已派送，避免每趟重試")
	}
	if len(setter.calls) != 0 {
		t.Fatalf("未知型別不得有副作用，got %+v", setter.calls)
	}
}
```

- [ ] **Step 2: 實作（`consumer.go`）**

```go
// Package consumer 派送 platform.events 的跨域副作用。
// 凍結／復原租戶是唯一會改動產品域資料的事件；其餘型別（期別、收款）目前僅供未來通知使用。
package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

type CompanyStatusSetter interface {
	SetStatus(ctx context.Context, companyID int, status string, reason string, actor authz.Identity) error
}

type Consumer struct {
	st          Store
	setter      CompanyStatusSetter
	systemActor authz.Identity // 排程的稽核主體（tenant audit 需要 user_id）
	statusFor   map[string]string
}

func New(st store.BillingStore, setter CompanyStatusSetter, systemActor authz.Identity) *Consumer {
	return &Consumer{
		st: st, setter: setter, systemActor: systemActor,
		// 事件 → 公司狀態：suspended 凍結、reactivated 復原。
		statusFor: map[string]string{
			"subscription.suspended":   "suspended",
			"subscription.reactivated": "active",
		},
	}
}

// DispatchOnce 依序處理未派送事件；回傳成功派送筆數。
// 每筆事件的副作用與「標記已派送」在同一交易；失敗則整筆回滾（下趟重試）。
func (c *Consumer) DispatchOnce(ctx context.Context, limit int) (int, error) {
	events, err := c.st.UndispatchedEvents(ctx, limit)
	if err != nil {
		return 0, err
	}
	done := 0
	for _, ev := range events {
		status, known := c.statusFor[ev.EventType]
		if !known {
			// 未知型別：標記已派送但留下 log，避免每趟重試同一筆（可觀測性優先於靜默）。
			log.Printf("platform consumer: 未知事件型別 %q（event=%d），標記為已派送", ev.EventType, ev.ID)
			if err := c.st.WithTx(ctx, func(tx *sql.Tx) error {
				return c.st.MarkEventDispatchedTx(ctx, tx, ev.ID)
			}); err != nil {
				return done, err
			}
			continue
		}
		var payload struct {
			CompanyID int `json:"company_id"`
		}
		if err := json.Unmarshal(ev.Payload, &payload); err != nil || payload.CompanyID == 0 {
			return done, fmt.Errorf("事件 %d 的 payload 無法解析: %w", ev.ID, err)
		}
		reason := "訂閱欠費停用（系統自動）"
		if status == "active" {
			reason = "訂閱復原（系統自動）"
		}
		if err := c.st.WithTx(ctx, func(tx *sql.Tx) error {
			if err := c.setter.SetStatus(ctx, payload.CompanyID, status, reason, c.systemActor); err != nil {
				return err
			}
			return c.st.MarkEventDispatchedTx(ctx, tx, ev.ID)
		}); err != nil {
			return done, err
		}
		done++
	}
	return done, nil
}
```

（`Consumer` 只依賴 Task 6 Interfaces 段的窄介面 `Store`（三支方法），因此假實作只需三支；`markEventDispatchedTx` 的 SQL 由 `store/postgres` 實作，內容為 `UPDATE platform.events SET dispatched_at = now(), attempts = attempts + 1 WHERE id = $1`。）

- [ ] **Step 3: 跑測試**

Run: `cd backend && go test ./internal/platform/consumer/ -v`
Expected: PASS（含失敗重試與未知型別兩條）

- [ ] **Step 4: Commit**

```bash
git add backend/internal/platform/consumer
git commit -m "feat(backend): outbox consumer 以事件驅動租戶凍結/復原（同交易、可重跑）"
```

---

### Task 7: `cmd/platform-cron`（單趟排程）

**觸發器設計（controller 裁定，2026-09-20）：支援兩條路徑，但只有一份工作單元**

`RunOnce` 仍是唯一工作單元；觸發方式只是**薄適配器**。落地順序：**(a) 本任務先做 `RunGuarded` ＋ 心跳落 DB ＋ binary**；**(b) HTTP 端點是後加項**（約 40 行；環境只能走 HTTP 時再加）。因為 (a) 的不變式已就位，(b) 是**純新增**，不需回頭改任何東西。

**七條不變式（兩條路徑都必須滿足）**：
1. 都呼叫**同一個** `RunGuarded(ctx, deps, now)`（＝`RunOnce` ＋ 單飛鎖），不得各寫一點邏輯。
2. 都走**系統範圍**（`dbtenant.NewClient` ＋ `SystemScopeTx`，scope=all）——它要改 `companies.status` 與寫租戶稽核（RLS 已 `ENABLE`+`FORCE`）；**裸 client 包 `SystemScopeTx` 一樣 42501**（Plan A T10 實測）。
3. `now` 由呼叫端給；`RunOnce` 內**不得** `time.Now()`。
4. **共用同一顆單飛鎖**（`pg_try_advisory_lock(常數 key)`，不新增表）：鎖拿不到 → 記 log 回「跳過」（HTTP 路徑回 409），**不是錯誤**。
5. **同一稽核主體**（`platform.settings` 的系統 actor，G5）——否則同一件事在稽核裡會有兩種 actor。
6. **同一份 `Summary` struct**（log 行與 JSON 兩種呈現），不得各自定義。
7. **心跳落 DB**（`platform.audit_logs` 一筆 `cron_run`，或 `platform.settings` 的 `cron_last_summary`）：讓「有沒有跑」是**資料**問題、與觸發路徑無關 → P1-6 只需盯一個地方。

**HTTP 端點（後加項）的硬要求**：token 由環境變數提供（不入版控）＋ **constant-time 比對**；限制來源網段；請求逾時 ≥ 一趟正常執行時間；那筆心跳要記**觸發者**（token 身分／IP）——該端點能停掉租戶，必須留痕。**驗收**：一條測試用同一組 fake deps 分別走 binary 與端點 handler，斷言兩者 `Summary` 完全相同（這是「不得漂移」的守門）。

**Files:**
- Create: `internal/platform/cron/run.go`、`internal/platform/cron/run_test.go`
- Create: `cmd/platform-cron/main.go`
- Modify: `Taskfile.yml`（`platform:cron`）

**Interfaces:**
- Produces:

```go
type Deps struct {
	Billing  *billing.Billing
	Consumer *consumer.Consumer
	Store    store.BillingStore
}

type Summary struct {
	PastDue       int
	Suspended     int
	PeriodsOpened int
	Dispatched    int
	Receivables   int
}

// RunOnce 執行一趟完整排程；now 由呼叫端提供（可覆寫 → 可測）。
func RunOnce(ctx context.Context, deps Deps, now time.Time, cfg Params) (Summary, error)

type Params struct {
	GraceDays      int
	LeadDays       int
	EventBatch     int
}
```

- [ ] **Step 1: 寫失敗測試（`run_test.go`）**

```go
// 一趟排程的順序與冪等：逾期末付 → past_due；寬限已過 → suspended；進入提前窗 → 開下一期；
// 事件派送（含凍結）；重跑第二趟時所有計數應為 0（除待收款清單）。
func TestRunOnceIsIdempotent(t *testing.T) {
	deps := newFakeDeps(t) // 假 store ＋ 假 setter
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)

	first, err := RunOnce(context.Background(), deps, now, Params{GraceDays: 7, LeadDays: 14, EventBatch: 100})
	if err != nil {
		t.Fatalf("第一趟: %v", err)
	}
	// fixture 需含：1 家逾期末付（→past_due）、1 家寬限已過（→suspended）、
	// 1 家 cancelled 且期末已過（→expired，G7）、1 家進入提前窗（→開下一期）。
	if first.PastDue != 1 || first.Suspended != 1 || first.ExpiredCancelled != 1 ||
		first.PeriodsOpened != 1 || first.Dispatched < 1 {
		t.Fatalf("第一趟計數不符: %+v", first)
	}
	second, err := RunOnce(context.Background(), deps, now, Params{GraceDays: 7, LeadDays: 14, EventBatch: 100})
	if err != nil {
		t.Fatalf("第二趟: %v", err)
	}
	if second.PastDue != 0 || second.Suspended != 0 || second.ExpiredCancelled != 0 ||
		second.PeriodsOpened != 0 || second.Dispatched != 0 {
		t.Fatalf("重跑不得重複轉移或重複派送: %+v", second)
	}
}
```

- [ ] **Step 2: 實作（`run.go`）**

```go
// Package cron 為平台排程的單趟執行（不內建迴圈）：可由 k8s CronJob、Taskfile 或測試呼叫，
// 時間由呼叫端給，故完全可測、可重跑。
package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

type Deps struct {
	Billing  *billing.Billing
	Consumer *consumer.Consumer
	Store    store.BillingStore
}

type Params struct {
	GraceDays  int
	LeadDays   int
	EventBatch int
}

// LoadParams 由 platform.settings 讀取排程參數（key：grace_days／lead_days）。
// 缺席或值不合理一律回錯誤 —— **不得默默用預設值**：那會讓「忘記 seed」變成
// 無聲的錯誤寬限期（帳務參數的預設值不能是猜的）。
func LoadParams(ctx context.Context, st store.BillingStore) (Params, error) {
	parse := func(key string) (int, error) {
		raw, err := st.Setting(ctx, key)
		if err != nil {
			return 0, fmt.Errorf("缺少設定 %s: %w", key, err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return 0, fmt.Errorf("設定 %s 不是整數: %q", key, raw)
		}
		if n < 0 || n > 365 {
			return 0, fmt.Errorf("設定 %s 的值不合理（應為 0..365 天）: %d", key, n)
		}
		return n, nil
	}
	var p Params
	var err error
	if p.GraceDays, err = parse("grace_days"); err != nil {
		return Params{}, err
	}
	if p.LeadDays, err = parse("lead_days"); err != nil {
		return Params{}, err
	}
	p.EventBatch = 200 // 每趟派送上限；非營運參數，留在程式碼
	return p, nil
}

type Summary struct {
	PastDue          int `json:"past_due"`
	Suspended        int `json:"suspended"`
	ExpiredCancelled int `json:"expired_cancelled"`
	PeriodsOpened    int `json:"periods_opened"`
	Dispatched       int `json:"dispatched"`
}

// RunOnce 依序執行：逾期轉移 → 寬限停用 → 取消到期 → 事件派送（凍結/復原）→ 產生下一期。
// 順序有依賴：先轉移狀態並派送事件（凍結），再產期別 —— 避免對剛停用的租戶開新期。
// 待收款清單不在排程內（它是唯讀視圖，由 PlatformAdminService.ListReceivables 提供）。
func RunOnce(ctx context.Context, deps Deps, now time.Time, p Params) (Summary, error) {
	var s Summary
	var err error

	if s.PastDue, err = deps.Billing.MarkPastDue(ctx, now, p.GraceDays); err != nil {
		return s, fmt.Errorf("標記逾期: %w", err)
	}
	if s.Suspended, err = deps.Billing.SuspendOverdue(ctx, now); err != nil {
		return s, fmt.Errorf("停用欠費: %w", err)
	}
	// G7：已取消且期末已過 → 發 subscription.expired（下一段的 DispatchOnce 會派送成凍結）。
	if s.ExpiredCancelled, err = deps.Billing.ExpireCancelled(ctx, now); err != nil {
		return s, fmt.Errorf("取消到期: %w", err)
	}
	if s.Dispatched, err = deps.Consumer.DispatchOnce(ctx, p.EventBatch); err != nil {
		return s, fmt.Errorf("派送事件: %w", err)
	}

	subs, err := deps.Store.ActiveOrTrialingSubscriptions(ctx)
	if err != nil {
		return s, fmt.Errorf("列出有效訂閱: %w", err)
	}
	for _, sub := range subs {
		created, err := deps.Billing.EnsureNextPeriod(ctx, sub.CompanyID, now, p.LeadDays)
		if err != nil {
			return s, fmt.Errorf("產生期別(company=%d): %w", sub.CompanyID, err)
		}
		if created {
			s.PeriodsOpened++
		}
	}
	return s, nil
}
```

- [ ] **Step 3: 入口（`cmd/platform-cron/main.go`）**

```go
// 平台排程入口（單趟）：config → admin DB → store/billing/consumer → RunOnce → 輸出摘要。
// 以 --date 覆寫「現在」以便手動補跑或測試；不內建迴圈（交由 k8s CronJob／Taskfile）。
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	platformcron "github.com/salesorder/sales-order-1.0/backend/internal/platform/cron"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
)

func main() {
	dateFlag := flag.String("date", "", "覆寫執行時間（RFC3339；預設為現在）")
	flag.Parse()

	now := time.Now().UTC()
	if *dateFlag != "" {
		parsed, err := time.Parse(time.RFC3339, *dateFlag)
		if err != nil {
			log.Fatalf("--date 格式錯誤: %v", err)
		}
		now = parsed
	}

	cfg := config.New()
	db, err := sql.Open("pgx", cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("admin 連線: %v", err)
	}
	defer db.Close()

	st := postgres.New(db)
	actorID, err := st.SystemActor(context.Background())
	if err != nil {
		log.Fatalf("讀取系統 actor 失敗: %v\n"+
			"修復：先執行 `task seed`（或 `go run ./cmd/seed`）建立平台自營公司與系統使用者，"+
			"它會寫入 platform.settings.system_actor_user_id（G5）。", err)
	}
	systemActor := authz.Identity{UserID: strconv.FormatInt(actorID, 10), Role: "super", Roles: []string{"super"}}

	// 產品域狀態入口：以業務連線執行（與 web 路徑同一份語意）。
	bizClient, err := database.OpenEntForCron(cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatalf("業務連線: %v", err)
	}
	defer bizClient.Close()
	setter := services.NewCompanyStatusSetter(bizClient)

	deps := platformcron.Deps{
		Billing:  billing.NewBilling(st),
		Consumer: consumer.New(st, setter, systemActor),
		Store:    st,
	}
	// 參數一律來自 platform.settings（可於營運工具調整；首次由 seed 以 env 預設值寫入）。
	// 不硬編在程式碼裡：寬限天數與提前天數是營運參數，會隨客戶與季節調整。
	params, err := platformcron.LoadParams(context.Background(), st)
	if err != nil {
		log.Fatalf("讀取排程參數失敗: %v\n"+
			"修復：執行 `task seed` 建立預設值（grace_days／lead_days／trial_days）。", err)
	}

	summary, err := platformcron.RunOnce(context.Background(), deps, now, params)
	if err != nil {
		log.Fatalf("排程執行失敗: %v", err)
	}
	// 摘要帶 finished_at：缺席偵測以「最後一次摘要的時間」判斷排程是否還在跑（見 Step 4）。
	raw, _ := json.Marshal(struct {
		platformcron.Summary
		FinishedAt string `json:"finished_at"`
	}{Summary: summary, FinishedAt: time.Now().UTC().Format(time.RFC3339)})
	log.Printf("platform-cron 完成（now=%s）: %s", now.Format(time.RFC3339), raw)
}
```

（`services.NewCompanyStatusSetter(db)` 為 Task 1 `SetCompanyStatus` 的介面轉接（實作 `consumer.CompanyStatusSetter`）；`database.OpenEntForCron` 沿用既有 `database.OpenEnt`，不需新函式——實作時直接呼叫 `database.OpenEnt`。）

- [ ] **Step 4: 執行環境（compose cron）＋ Taskfile ＋最小告警**

```yaml
  platform:cron:
    desc: 執行一趟平台排程（產生期別、逾期轉移、事件派送；--date 可覆寫時間）
    dir: backend
    cmds:
      - go run ./cmd/platform-cron {{.CLI_ARGS}}
```

**排程不能只靠人工執行**（排序文件 P1-5：沒有部署就沒有排程，逾期凍結等於不會發生）。`docker-compose.dev.yml` 追加一個獨立的 cron 容器：

```yaml
  # 平台排程：每日執行一趟 platform-cron（k8s CronJob 於 D19 接手）。
  # 以 profiles 隔離：不隨 `task infra:start` 起（排程是常駐服務，不該混進每次開發的基礎設施）。
  platform-cron:
    profiles: ["cron"]
    build:
      context: ./backend
      dockerfile: Dockerfile          # 沿用既有 backend image；若無則以 golang image 編譯
    command: ["sh", "-c", "while true; do /app/platform-cron; sleep 86400; done"]
    environment:
      DATABASE_URL: postgres://app_rw:app_rw@postgres:5432/salesorder?sslmode=disable
      DATABASE_ADMIN_URL: postgres://postgres:postgres@postgres:5432/salesorder?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy
```

```yaml
  platform:cron:up:
    desc: 啟動排程容器（profiles=cron）
    cmd: DOCKER_HOST="unix://{{.PODMAN_SOCK}}" docker-compose -f docker-compose.dev.yml --profile cron up -d platform-cron

  platform:cron:check:
    desc: 缺席偵測——檢查排程是否在容許間隔內執行過（正式環境由告警系統接手）
    cmds:
      - ./scripts/check_cron_freshness.sh 26
```

**最小告警（P1-6）**：無聲故障是最貴的故障——排程沒跑、凍結沒生效，帳就直接漏。兩層：

1. **每次執行留摘要**：`cmd/platform-cron` 的摘要 JSON 追加 `finished_at`（於 `main` 印出時填入 `time.Now().UTC()`；**不進 `RunOnce` 回傳值**，避免排程邏輯依賴時鐘），並把 stdout 導到固定 log。
2. **缺席偵測**：以「最近一次摘要的時間」判斷，而不是看 exit code（沒跑與跑失敗是兩件事）。

```bash
#!/usr/bin/env bash
# check_cron_freshness.sh：檢查 platform-cron 是否在容許間隔內執行過。
# 判斷依據是 log 中最後一筆摘要的 finished_at，而非 exit code —— 排程最常見的失效是
# 「根本沒被執行」（容器沒起、cron 沒設），不是「執行失敗」。
set -euo pipefail
max_hours="${1:-26}"
log="${PLATFORM_CRON_LOG:-/var/log/platform-cron.log}"

if [[ ! -f "$log" ]]; then
  echo "ALERT: 找不到排程 log（$log）——排程可能從未執行" >&2
  exit 1
fi
last=$(grep -o '"finished_at":"[^"]*"' "$log" | tail -1 | cut -d'"' -f4 || true)
if [[ -z "$last" ]]; then
  echo "ALERT: 排程 log 沒有摘要紀錄" >&2
  exit 1
fi
age_h=$(( ( $(date -u +%s) - $(date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$last" +%s 2>/dev/null || date -u -d "$last" +%s) ) / 3600 ))
if (( age_h > max_hours )); then
  echo "ALERT: 排程已 ${age_h} 小時未執行（上限 ${max_hours}）" >&2
  exit 1
fi
echo "OK: 排程最後執行於 ${last}"
```

（GNU/BSD `date` 的 `-d` 與 `-j -f` 寫法不同：上面的 `||` 已同時涵蓋 macOS 與 Linux；實作時若嫌醜，改用 `python3 -c` 或讓 `platform-cron` 另外寫一個 epoch 秒數的檔——**重點是判斷依據必須是時間，不是 exit code**。）

- [ ] **Step 5: 跑測試**

Run: `cd backend && go test ./internal/platform/cron/ -v && go build ./cmd/platform-cron`
Expected: PASS、可建置

- [ ] **Step 6: Commit**

```bash
git add backend/internal/platform/cron backend/cmd/platform-cron backend/Taskfile.yml
git commit -m "feat(backend): platform-cron 單趟排程（逾期→停用→事件派送→產生期別→待收款）"
```

---

### Task 8: Valkey 快取實作與失效

**Files:**
- Create: `internal/platform/entitlements/valkey.go`、`internal/platform/entitlements/valkey_test.go`（整合）
- Modify: `internal/server/domains.go`（以 Valkey 取代 `MemoryCache`）、`internal/services/platform_admin_service.go`（寫入後失效）

**Interfaces:**
- Produces: `entitlements.NewValkeyCache(client *redis.Client) entitlements.Cache`；`entitlements.Invalidate(ctx, cache, companyID) error`

- [ ] **Step 1: 寫整合測試（連 Valkey；無 Valkey 即 skip）**

```go
//go:build integration

// TestIntegrationValkeyCacheRoundTrip 驗證：
// ① Set 後 Get 命中；② Delete 後不再命中；③ Invalidate 會清掉該租戶的鍵（寫入路徑靠它）。
func TestIntegrationValkeyCacheRoundTrip(t *testing.T) {
	addr := os.Getenv("VALKEY_ADDR")
	if addr == "" {
		t.Skip("未設定 VALKEY_ADDR（本機請 task infra:start），略過 Valkey 快取整合測試")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })
	c := entitlements.NewValkeyCache(client)
	ctx := t.Context()

	if err := c.Set(ctx, "ent:42", []byte(`{"ok":true}`), time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, ok, err := c.Get(ctx, "ent:42"); err != nil || !ok {
		t.Fatalf("Get 應命中: ok=%v err=%v", ok, err)
	}
	if err := entitlements.Invalidate(ctx, c, 42); err != nil {
		t.Fatalf("Invalidate: %v", err)
	}
	if _, ok, _ := c.Get(ctx, "ent:42"); ok {
		t.Fatal("Invalidate 後不得命中")
	}
}
```

- [ ] **Step 2: 實作（`valkey.go`）**

```go
package entitlements

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ValkeyCache 為跨 replica 共用的權益快取。失效由寫入路徑顯式 Delete（方案、override、
// 訂閱狀態異動），TTL 僅為保底：ttl 到期前的正確性不依賴 TTL。
type ValkeyCache struct{ client *redis.Client }

func NewValkeyCache(client *redis.Client) *ValkeyCache { return &ValkeyCache{client: client} }

func (v *ValkeyCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	raw, err := v.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

func (v *ValkeyCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return v.client.Set(ctx, key, val, ttl).Err()
}

func (v *ValkeyCache) Delete(ctx context.Context, key string) error {
	return v.client.Del(ctx, key).Err()
}

// Invalidate 清除某租戶的權益快取（寫入路徑的唯一失效入口）。
func Invalidate(ctx context.Context, c Cache, companyID int) error {
	if c == nil {
		return nil
	}
	if err := c.Delete(ctx, cacheKey(companyID)); err != nil {
		return fmt.Errorf("失效權益快取(company=%d): %w", companyID, err)
	}
	return nil
}
```

- [ ] **Step 3: 寫入路徑接上失效並測試**

在 `PlatformAdminService` 的每個寫入方法（override、RecordPayment、方案／價目）「提交成功後」呼叫：

```go
	if err := entitlements.Invalidate(ctx, s.cache, companyID); err != nil {
		// 失效失敗不得讓已成功的寫入回錯誤：記 log，讓 TTL 保底（60s 內自然收斂）。
		log.Printf("platform: 權益快取失效失敗(company=%d): %v", companyID, err)
	}
```

並補單元測試（兩條）：

```go
// ① 方案／價目變更 → 失效所有租戶快取（否則已開頁面的租戶仍讀舊權益）。
//    先補實作：entitlements.InvalidateAll 以 Keys() 掃 ent:* 逐鍵刪（v1 的量體很小）。
type Scanner interface {
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// InvalidateAll 清除所有租戶的權益快取（方案／價目異動時使用）。
// ponytail: 逐鍵 SCAN；方案改動頻率極低，量體成長或延遲可感時再改標記版本號。
func InvalidateAll(ctx context.Context, c Cache) error {
	sc, ok := c.(Scanner)
	if !ok {
		return errors.New("此快取實作不支援全量失效（缺少 Keys）")
	}
	keys, err := sc.Keys(ctx, "ent:*")
	if err != nil {
		return err
	}
	for _, k := range keys {
		if err := c.Delete(ctx, k); err != nil {
			return err
		}
	}
	return nil
}
```

```go
// scanCache 為支援 Keys 的測試替身。
type scanCache struct{ keys []string; deleted []string }

func (c *scanCache) Get(context.Context, string) ([]byte, bool, error) { return nil, false, nil }
func (c *scanCache) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (c *scanCache) Delete(_ context.Context, k string) error {
	c.deleted = append(c.deleted, k)
	return nil
}
func (c *scanCache) Keys(_ context.Context, pattern string) ([]string, error) {
	// 測試替身只需支援 "ent:*" 前綴（實作的 Valkey 版以 SCAN MATCH 實作）。
	var out []string
	for _, k := range c.keys {
		if strings.HasPrefix(k, "ent:") {
			out = append(out, k)
		}
	}
	return out, nil
}

func TestInvalidateAllRemovesEveryTenantKey(t *testing.T) {
	c := &scanCache{keys: []string{"ent:1", "ent:2", "other:9"}}
	if err := entitlements.InvalidateAll(context.Background(), c); err != nil {
		t.Fatalf("InvalidateAll: %v", err)
	}
	if len(c.deleted) != 2 {
		t.Fatalf("應刪除 2 個承租戶鍵，got %v", c.deleted)
	}
	for _, k := range c.deleted {
		if k == "other:9" {
			t.Fatal("不得刪除不相關的鍵")
		}
	}
}

// ② 寫入失敗時**不得**失效快取：失敗卻清快取會造成無謂重載，且掩蓋真實錯誤來源。
func TestInvalidationSkippedWhenWriteFails(t *testing.T) {
	cache := &recordingCache{}
	f := &fakeBilling{
		sub:         &store.Subscription{ID: 5, CompanyID: 42, Status: "active"},
		period:      &store.Period{ID: 9, SubscriptionID: 5, Status: "open"},
		markPaidErr: errors.New("模擬寫入失敗"),
	}
	h := NewPlatformAdminService(&fakePlatformStore{writes: &fakeWrites{}},
		billing.NewBilling(f), cache)

	if _, err := h.RecordPayment(withOperator(context.Background()),
		connect.NewRequest(&platformv1.RecordPaymentRequest{CompanyId: "42", Reason: "匯款"})); err == nil {
		t.Fatal("寫入失敗應回錯誤")
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("寫入失敗不得失效快取，got %v", cache.deleted)
	}
}
```

（`fakeBilling` 需有 `markPaidErr error` 欄位並在 `MarkPeriodPaidTx` 回傳它。`InvalidateAll` 的 Valkey 實作在 `valkey.go` 補 `Keys(ctx, pattern)`：`SCAN MATCH pattern COUNT 100` 逐批收集。）

- [ ] **Step 4: `domains.go` 換成 Valkey 實作**

```go
		entSvc := entitlements.New(platformStore, services.NewEntitlementCounter(entClient),
			entitlements.NewValkeyCache(valkeyClient), 60*time.Second)
```

- [ ] **Step 5: 跑測試**

Run: `cd backend && task check && task test:integration -- -run TestIntegrationValkeyCache -v`
Expected: 全綠（無 Valkey 時該測試 skip，其餘全跑）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/platform/entitlements backend/internal/server/domains.go backend/internal/services/platform_admin_service.go
git commit -m "feat(backend): 權益快取改 Valkey 並在寫入路徑顯式失效（TTL 僅保底）"
```

---

### Task 9: 平台寫入 RPC（override／收款／方案價目／待收款）與稽核

**Files:**
- Modify: `backend/proto/platform/v1/platform.proto`（新增寫入 RPC）、生成三端
- Modify: `internal/services/platform_admin_service.go`
- Create: `internal/services/platform_admin_write_test.go`、`platform_admin_write_integration_test.go`

**Interfaces:**
- Produces（proto 摘要）：

```proto
service PlatformAdminService {
  // 既有唯讀 …
  rpc ListReceivables(ListReceivablesRequest) returns (ListReceivablesResponse);
  rpc RecordPayment(RecordPaymentRequest) returns (RecordPaymentResponse);
  // 訂閱生命週期（G6）：席位、改方案、取消 —— 三者都必須填 reason（平台稽核必填）。
  rpc SetSeatCount(SetSeatCountRequest) returns (SetSeatCountResponse);
  rpc ChangePlan(ChangePlanRequest) returns (ChangePlanResponse);
  rpc CancelSubscription(CancelSubscriptionRequest) returns (CancelSubscriptionResponse);
  // 營運參數（G5/5-6）：試用／寬限／提前天數可由介面調整，cron 讀 settings 而非硬編。
  rpc GetBillingSettings(GetBillingSettingsRequest) returns (GetBillingSettingsResponse);
  rpc UpdateBillingSettings(UpdateBillingSettingsRequest) returns (UpdateBillingSettingsResponse);
  rpc SetTenantOverride(SetTenantOverrideRequest) returns (SetTenantOverrideResponse);
  rpc RevokeTenantOverride(RevokeTenantOverrideRequest) returns (RevokeTenantOverrideResponse);
  rpc UpsertPlanPrice(UpsertPlanPriceRequest) returns (UpsertPlanPriceResponse);
  rpc SetPlanEntitlement(SetPlanEntitlementRequest) returns (SetPlanEntitlementResponse);
  rpc CreateOperator(CreateOperatorRequest) returns (CreateOperatorResponse);
  rpc DisableOperator(DisableOperatorRequest) returns (DisableOperatorResponse);
}
```

三支生命週期 RPC 的訊息與語意（v1 刻意**不做按日比例計費**：變更一律「下一期生效」，避免在沒有金流對帳的前提下產生半期金額）：

```proto
message SetSeatCountRequest {
  string company_id = 1;
  int32  seat_count = 2;   // 新席位數；不得小於目前使用中的席次
  string reason = 3;       // 必填
}
message SetSeatCountResponse { int32 seat_count = 1; }

message ChangePlanRequest {
  string company_id = 1;
  string plan_code = 2;    // 新方案；下一期生效，當期不動
  string reason = 3;       // 必填
}
message ChangePlanResponse { string plan_code = 1; string effective_from = 2; }

message CancelSubscriptionRequest {
  string company_id = 1;
  bool   at_period_end = 2; // v1 僅支援 true（期末終止）；false 屬特殊處理，回 PLAT-3001
  string reason = 3;        // 必填
}
message CancelSubscriptionResponse { string cancelled_at = 1; string service_until = 2; }

message ListReceivablesRequest  { int32 page = 1; int32 page_size = 2; }
message ListReceivablesResponse {
  repeated Receivable rows = 1;
  PlatformPagination pagination = 2;
}
message Receivable {
  string company_id = 1;
  string company_name = 2;
  string plan_code = 3;
  int32  period_no = 4;
  string amount = 5;        // 兩位小數字串（"1500.00"）
  string period_end = 6;
  string status = 7;        // open | paid（僅列出 open 與已逾期）
}
```

**語意與邊界（實作時照此實作，不得自行放寬）**——拒絕條件一律回**註冊碼**（Plan D；`internal/errcode` 已可用，落地歸屬 Task 5b）：

| RPC | 行為 | 拒絕條件（碼） |
|---|---|---|
| `SetSeatCount` | 更新 `subscriptions.seat_count`（下一次產期即用新席位數計價） | `seat_count < 使用中席次`（以計數器查）→ `PLAT-5001` `PlatformLimitExceeded`（details `feature=limit.seats`／`used`／`limit`）；`<= 0` → `SYS-1001` `SysInvalidArgument` |
| `ChangePlan` | 只改 `subscriptions.plan_id`；**當期期別不動**（價格快照已寫死），下一期起用新方案與新價 | 目標方案不存在／已歸檔 → `SYS-4002` `SysNotFound`（不存在語意，不另立碼）；與現行方案相同 → no-op（不寫稽核） |
| `CancelSubscription` | `status → cancelled`（`allowedTransitions` 檢查）＋發 `subscription.cancelled`；期末後由 `ExpireCancelled` 轉 `suspended` | 已 `cancelled` → no-op；`at_period_end=false` → `PLAT-3001` `PlatformSubscriptionInactive`（v1 不支援立即終止）；狀態機拒絕 → 同碼 |
| `ListReceivables` | 列出所有 `open` 期別（含已逾期者）＋公司名；供 console 匯出 CSV | 無（唯讀） |

- [ ] **Step 1: proto 與生成**

依上列服務定義補齊訊息（`reason` 為所有寫入 RPC 的共同必填欄位；`RecordPaymentRequest` 欄位對齊 `billing.RecordPaymentInput`），然後：

```bash
cd backend && export PATH="$HOME/fvm/versions/stable/bin:$HOME/.pub-cache/bin:$PATH" && task proto:gen
cd .. && pnpm -C platform-console typecheck   # Task 11 建好後才有意義；此步驗證生成物可編譯
```

- [ ] **Step 2: 寫測試（`platform_admin_write_test.go`，假 store ＋ 記錄版 cache）**

先讓 `NewPlatformAdminService` 取得帳務與快取依賴（**Plan B 的 Task 9 只有 `st`，此處擴充簽名**）：

```go
func NewPlatformAdminService(st platformStore, b *billing.Billing, cache entitlements.Cache) *PlatformAdminService {
	return &PlatformAdminService{st: st, billing: b, cache: cache, now: time.Now}
}
```

```go
package services

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/platform/v1"
)

// recordingCache 記錄失效呼叫；其餘 Cache 方法為 no-op。
type recordingCache struct{ deleted []string }

func (c *recordingCache) Get(context.Context, string) ([]byte, bool, error) { return nil, false, nil }
func (c *recordingCache) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}
func (c *recordingCache) Delete(_ context.Context, key string) error {
	c.deleted = append(c.deleted, key)
	return nil
}

func withOperator(ctx context.Context) context.Context {
	return operatorauth.WithIdentity(ctx,
		operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
}

// TestWriteRPCsShareThreeContracts 以表驅動驗證所有寫入 RPC 的三條共同契約：
// ① 無 operator 身分 → Unauthenticated（不依賴 interceptor）；
// ② reason 全空白 → InvalidArgument（平台稽核必填）；
// ③ 成功 → 該租戶的權益快取被失效一次（否則改了方案卻仍讀舊值）。
func TestWriteRPCsShareThreeContracts(t *testing.T) {
	newHarness := func() (*PlatformAdminService, *recordingCache) {
		cache := &recordingCache{}
		h := NewPlatformAdminService(
			&fakePlatformStore{writes: &fakeWrites{}},
			billing.NewBilling(&fakeBilling{
				sub:    &store.Subscription{ID: 5, CompanyID: 42, PlanCode: "std", Status: "active"},
				period: &store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open"},
			}),
			cache,
		)
		return h, cache
	}

	cases := []struct {
		name string
		call func(h *PlatformAdminService, ctx context.Context, reason string) error
	}{
		{"RecordPayment", func(h *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := h.RecordPayment(ctx, connect.NewRequest(
				&platformv1.RecordPaymentRequest{CompanyId: "42", PeriodNo: 1, Reason: reason}))
			return err
		}},
		{"SetSeatCount", func(h *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := h.SetSeatCount(ctx, connect.NewRequest(
				&platformv1.SetSeatCountRequest{CompanyId: "42", SeatCount: 5, Reason: reason}))
			return err
		}},
		{"ChangePlan", func(h *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := h.ChangePlan(ctx, connect.NewRequest(
				&platformv1.ChangePlanRequest{CompanyId: "42", PlanCode: "pro", Reason: reason}))
			return err
		}},
		{"CancelSubscription", func(h *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := h.CancelSubscription(ctx, connect.NewRequest(
				&platformv1.CancelSubscriptionRequest{
					CompanyId: "42", AtPeriodEnd: true, Reason: reason}))
			return err
		}},
		{"SetTenantOverride", func(h *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := h.SetTenantOverride(ctx, connect.NewRequest(
				&platformv1.SetTenantOverrideRequest{
					CompanyId: "42", FeatureCode: entitlements.LimitSeats,
					LimitSet: true, LimitValue: 50,
					Owner: "sales@example.com", Reason: reason,
				}))
			return err
		}},
		{"RevokeTenantOverride", func(h *PlatformAdminService, ctx context.Context, reason string) error {
			_, err := h.RevokeTenantOverride(ctx, connect.NewRequest(
				&platformv1.RevokeTenantOverrideRequest{OverrideId: "1", Reason: reason}))
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, cache := newWriteHarness()

			if err := tc.call(h, context.Background(), "匯款入帳"); connect.CodeOf(err) != connect.CodeUnauthenticated {
				t.Fatalf("無 operator 身分應 Unauthenticated，got %v", err)
			}
			if err := tc.call(h, withOperator(context.Background()), "   "); connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("reason 全空白應 InvalidArgument（平台稽核必填），got %v", err)
			}
			if err := tc.call(h, withOperator(context.Background()), "合法原因"); err != nil {
				t.Fatalf("合法呼叫應成功: %v", err)
			}
			if len(cache.deleted) == 0 {
				t.Fatal("寫入成功後必須失效權益快取（否則方案改了仍讀舊值）")
			}
			// RevokeTenantOverride 由 overrideId 反查租戶，鍵值在該路徑無法從參數斷言，
			// 故只要求「有失效」；另兩條路徑斷言鍵為 ent:42。
			if tc.name != "RevokeTenantOverride" && cache.deleted[len(cache.deleted)-1] != "ent:42" {
				t.Fatalf("應失效 ent:42，got %v", cache.deleted)
			}
		})
	}
}

// newWriteHarness 建構一組乾淨的服務 ＋ 記錄版快取（每個 case 各自一份，避免互相污染）。
func newWriteHarness() (*PlatformAdminService, *recordingCache) {
	cache := &recordingCache{}
	h := NewPlatformAdminService(
		&fakePlatformStore{writes: &fakeWrites{}},
		billing.NewBilling(&fakeBilling{
			sub:    &store.Subscription{ID: 5, CompanyID: 42, PlanCode: "std", Status: "active"},
			period: &store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open"},
		}),
		cache,
	)
	return h, cache
}
```

（`fakePlatformStore` 需追加 Task 9 新增的寫入方法（`SetTenantOverride`／`RevokeTenantOverride`／`UpsertPlanPrice`／`SetPlanEntitlement`／`CreateOperator`／`DisableOperator`）並以 `fakeWrites` 記錄呼叫；`fakeBilling` 沿用 Task 4 的假實作。`newHarness` 已由 `newWriteHarness` 取代，不得並存。）

- [ ] **Step 3: 實作寫入 RPC（`platform_admin_service.go`）**

```go
// RecordPayment 為人工收款的營運入口；金流 webhook（日後）呼叫同一個 billing 方法。
func (s *PlatformAdminService) RecordPayment(ctx context.Context,
	req *connect.Request[platformv1.RecordPaymentRequest]) (*connect.Response[platformv1.RecordPaymentResponse], error) {
	id, err := s.requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	companyID, err := parseID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	amount, err := money.ParseCents(req.Msg.GetAmount()) // 空字串 → 0 = 採用期別快照金額
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	period, err := s.billing.RecordPayment(ctx, billing.RecordPaymentInput{
		CompanyID: companyID, PeriodNo: int(req.Msg.GetPeriodNo()), PaidAt: s.now(),
		AmountCents: amount, Provider: req.Msg.GetProvider(), ExternalRef: req.Msg.GetExternalRef(),
		InvoiceNo: req.Msg.GetInvoiceNo(), InvoiceStatus: req.Msg.GetInvoiceStatus(),
		BuyerTaxID: req.Msg.GetBuyerTaxId(), Carrier: req.Msg.GetCarrier(),
		ActorOperatorID: id.OperatorID, Reason: req.Msg.GetReason(),
	})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, companyID)
	return connect.NewResponse(&platformv1.RecordPaymentResponse{
		PeriodNo: int32(period.PeriodNo), Status: period.Status,
	}), nil
}

// requireOperator 為所有寫入 RPC 的共同前置：身分 ＋ 服務層再檢查（不依賴 interceptor）。
func (s *PlatformAdminService) requireOperator(ctx context.Context) (operatorauth.Identity, error) {
	id, ok := operatorauth.IdentityFrom(ctx)
	if !ok {
		return operatorauth.Identity{}, connect.NewError(connect.CodeUnauthenticated,
			errors.New("平台操作需要 operator 身分"))
	}
	return id, nil
}

// invalidate 失效該租戶權益快取；失敗只記 log（TTL 保底，不讓已成功的寫入回錯）。
func (s *PlatformAdminService) invalidate(ctx context.Context, companyID int) {
	if err := entitlements.Invalidate(ctx, s.cache, companyID); err != nil {
		log.Printf("platform: 權益快取失效失敗(company=%d): %v", companyID, err)
	}
}
```

`SetTenantOverride`（寫 `tenant_overrides` ＋ 稽核 ＋ 失效）、`RevokeTenantOverride`（設 `revoked_at`）、`UpsertPlanPrice`／`SetPlanEntitlement`（改方案後失效**所有**租戶快取：以 `InvalidateAll` 掃 `ent:*`——v1 用 `SCAN` 逐鍵刪，並記 `ponytail:` 註記（方案改動頻率低））、`CreateOperator`／`DisableOperator`（白名單維護，落稽核）比照同一骨架實作。

- [ ] **Step 4: 不重複寫 DB 層整合測試（刻意省略，附理由）**

DB 層行為（期別轉 paid、訂閱復原、事件與稽核同交易、重送 no-op）已由 **Task 4 Step 5 的 `TestIntegrationRecordPayment`** 覆蓋，而 RPC 層的差異只在「身分、參數驗證、快取失效」——那些由 Step 2 的單元測試覆蓋。**刻意不新增第二條整合測試**：同一條入帳路徑寫兩次整合測試會讓 fixture 維護成本翻倍，卻測同一件事。

因此本步驟的驗收是「跑 Step 5 的既有測試集合」，不產出新檔案。

- [ ] **Step 5: 跑測試**

Run: `cd backend && task check && task test:integration -- -run 'TestIntegrationRecordPayment|TestIntegrationPlatform' -v`
Expected: 全綠

- [ ] **Step 6: Commit**

```bash
git add backend/proto/platform backend/internal/proto/platform backend/internal/services platform-console/src/lib/proto
git commit -m "feat(backend): 平台寫入 RPC（收款、override、方案價目、operator）與稽核"
```

---

### Task 10: `platform-console/` 腳手架（獨立 SPA、operator 登入、API client、守衛）

**Files:**
- Create: `platform-console/package.json`、`vite.config.ts`、`tsconfig.json`、`index.html`、`src/main.tsx`、`src/App.tsx`、`src/router.tsx`、`src/lib/api.ts`、`src/lib/auth.tsx`、`src/lib/auth.test.tsx`、`vitest.config.ts`、`Taskfile.yml`、`README.md`
- Modify: `pnpm-workspace.yaml`（加 `platform-console`）、`Taskfile.yml`（root includes）、`buf.gen.yaml`（TS 輸出指向 console）

**Interfaces:**
- Produces: 可跑的 console 骨架（登入 → 六頁路由 → 未登入導向登入）
- **前置缺口（Plan B Task 8 遺留，2026-09-20 記）**：operator 認證 **v1 沒有登出端點** —— `platform_session` 效期 12h 內只能靠停用 `operators.status` 即時失效（已有測試）。選項：(a) console 只做「清 cookie」的登出（純前端即可，不動後端）；(b) 後端 `Logout` RPC ＋ Valkey 黑名單（需要一個可撤銷清單）。本頁要一併決定，否則使用者按不到登出。

- [ ] **Step 1: workspace 與 proto 輸出**

```yaml
# pnpm-workspace.yaml
packages:
  - backend
  - frontend
  - app
  - platform-console
  - infra
```

```yaml
# buf.gen.yaml：在既有 bufbuild/es 之後追加第二個 TS 輸出（console 專用）
  - remote: buf.build/bufbuild/es
    out: ../platform-console/src/lib/proto
```

- [ ] **Step 2: 寫 package 骨架**

```json
// platform-console/package.json
{
  "name": "platform-console",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc --noEmit && vite build",
    "typecheck": "tsc --noEmit",
    "lint": "eslint .",
    "test": "vitest run"
  },
  "dependencies": {
    "@bufbuild/protobuf": "^2.15.0",
    "@connectrpc/connect": "^2.2.0",
    "@connectrpc/connect-web": "^2.2.0",
    "@tanstack/solid-query": "^5.102.8",
    "@tanstack/solid-router": "^1.170.30",
    "@tanstack/solid-table": "^9.2.4",
    "solid-js": "^1.9.15"
  },
  "devDependencies": {
    "@solidjs/testing-library": "^0.8.10",
    "jsdom": "^26.1.0",
    "typescript": "^5.9.0",
    "vite": "^8.3.0",
    "vite-plugin-solid": "^2.11.14",
    "vitest": "^4.1.11"
  }
}
```

```ts
// platform-console/vite.config.ts
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import path from "node:path";

// UI 元件庫以 alias 共用租戶 SPA 的既有實作（同一份 Ark UI × Tailkit 元件）：
// 不另抽共用 package，等第三個 app 出現再抽（ponytail: 目前兩個 app，alias 足夠）。
export default defineConfig({
  plugins: [solid()],
  resolve: {
    alias: {
      "~": path.resolve(import.meta.dirname, "src"),
      "@ui": path.resolve(import.meta.dirname, "../frontend/src/components/ui"),
      "@lib": path.resolve(import.meta.dirname, "../frontend/src/lib"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // console 只打平台 API 與平台登入端點；租戶 API 一律不代理（隔離）
      "/platform": { target: "http://localhost:3080", changeOrigin: true },
    },
  },
});
```

- [ ] **Step 3: API client 與登入（`src/lib/api.ts`、`src/lib/auth.tsx`）**

```ts
// api.ts：所有平台呼叫都帶 cookie（operator session 為 HttpOnly cookie）。
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PlatformAdminService } from "./proto/platform/v1/platform_pb";

const transport = createConnectTransport({
  baseUrl: "/",
  fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
});

export const platform = createClient(PlatformAdminService, transport);

// 平台登入入口：後端 OIDC 完成後會帶 cookie 導回 console 根路徑。
export const loginUrl = "/platform/auth/google";
```

```tsx
// auth.tsx：以「能否列出 0 筆租戶」判斷 session 是否有效（單一探針，不新增端點）。
import { createContext, useContext, type JSX } from "solid-js";
import { createQuery } from "@tanstack/solid-query";
import { platform } from "./api";

type AuthState = { ready: boolean; authenticated: boolean };
const Ctx = createContext<AuthState>({ ready: false, authenticated: false });

export function AuthProvider(props: { children: JSX.Element }) {
  const probe = createQuery(() => ({
    queryKey: ["auth-probe"],
    queryFn: async () => {
      try {
        await platform.listTenants({ page: 1, pageSize: 1 });
        return true;
      } catch {
        return false; // 未登入或無白名單：一律視為未登入（fail-closed）
      }
    },
    retry: false,
  }));
  const state = () => ({
    ready: !probe.isLoading,
    authenticated: probe.data === true,
  });
  return <Ctx.Provider value={state()}>{props.children}</Ctx.Provider>;
}

export const useAuth = () => useContext(Ctx);
```

- [ ] **Step 4: 路由與守衛（`src/router.tsx`、`src/App.tsx`）**

```tsx
// router.tsx：未登入一律導向登入頁（外部跳轉到 /platform/auth/google）。
import { Navigate, Route, Router } from "@tanstack/solid-router";
import { useAuth } from "./lib/auth";
import TenantsPage from "./pages/TenantsPage";
import TenantDetailPage from "./pages/TenantDetailPage";
import PlansPage from "./pages/PlansPage";
import EntitlementsPage from "./pages/EntitlementsPage";
import ReceivablesPage from "./pages/ReceivablesPage";
import AuditPage from "./pages/AuditPage";

export function AppRouter() {
  const auth = useAuth();
  return (
    <Router>
      <Route path="/" component={() => <Navigate to="/tenants" />} />
      <Route
        path="/tenants"
        component={() => (auth.ready && !auth.authenticated ? <Login /> : <TenantsPage />)}
      />
      {/* 其餘五頁同型：/tenants/$id、/plans、/entitlements、/receivables、/audit */}
    </Router>
  );
}
```

- [ ] **Step 5: 守衛測試（`src/lib/auth.test.tsx`）**

```tsx
// ① 探針失敗（未登入）→ authenticated=false；② 探針成功 → true。
// 以假 transport（攔截 listTenants）驗證，不連真後端。
it("未登入時 authenticated 必為 false（fail-closed）", async () => {
  // 假 transport 擲出 Unauthenticated；斷言畫面顯示登入引導且不渲染任何租戶資料。
});

it("探針成功時 authenticated 為 true", async () => {
  // 假 transport 回空清單；斷言 authenticated=true。
});
```

- [ ] **Step 6: 建置與測試**

Run: `cd platform-console && pnpm install && pnpm typecheck && pnpm test && pnpm build`
Expected: 全綠

- [ ] **Step 7: Commit**

```bash
git add platform-console pnpm-workspace.yaml Taskfile.yml buf.gen.yaml pnpm-lock.yaml
git commit -m "feat(console): 平台營運工具腳手架（獨立 SPA、operator 登入、路由守衛、UI alias）"
```

---

### Task 11: console 頁面（租戶列表／租戶詳情＋override／方案與價目）

**Files:**
- Create: `platform-console/src/pages/TenantsPage.tsx`、`TenantDetailPage.tsx`、`PlansPage.tsx`
- Create: 對應測試 `TenantsPage.test.tsx` 等

**Interfaces:**
- Consumes: `platform` client（Task 10）、`@ui/*` 元件（租戶 SPA 的 `table`／`button`／`dialog`／`field`）

- [ ] **Step 1: 寫測試（先寫契約）**

```tsx
import { render, screen, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { platform } from "../lib/api";
import TenantsPage from "./TenantsPage";

// 以模組替身取代 API：只驗 UI 契約，不連後端（console 的測試與租戶 SPA 同策略）。
vi.mock("../lib/api", () => ({ platform: { listTenants: vi.fn() } }));

const rows = [
  { companyId: "1", companyName: "甲公司", planCode: "std", planName: "標準",
    subscriptionStatus: "active", seatCount: 8, currentPeriodEnd: "2026-10-31T00:00:00Z", overdue: false },
  { companyId: "2", companyName: "乙公司", planCode: "pro", planName: "專業",
    subscriptionStatus: "past_due", seatCount: 12, currentPeriodEnd: "2026-09-15T00:00:00Z", overdue: true },
  { companyId: "3", companyName: "丙公司", planCode: "free", planName: "免費",
    subscriptionStatus: "none", seatCount: 0, currentPeriodEnd: "", overdue: false },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(() => (
    <QueryClientProvider client={client}>
      <TenantsPage />
    </QueryClientProvider>
  ));
}

describe("TenantsPage", () => {
  beforeEach(() => {
    vi.mocked(platform.listTenants).mockResolvedValue({
      tenants: rows, pagination: { page: 1, pageSize: 20, total: 3 },
    } as never);
  });

  it("列出三家租戶，且僅逾期者顯示標記", async () => {
    renderPage();
    // 結構斷言：表頭 + 3 列（不得以「某段文字不存在」代替狀態判定）
    await waitFor(() => expect(screen.getAllByRole("row")).toHaveLength(4));
    expect(screen.getByText("乙公司")).toBeTruthy();
    expect(screen.getAllByText("逾期")).toHaveLength(1);
  });

  it("空清單顯示空狀態而非空白表格", async () => {
    vi.mocked(platform.listTenants).mockResolvedValue({
      tenants: [], pagination: { page: 1, pageSize: 20, total: 0 },
    } as never);
    renderPage();
    await waitFor(() => expect(screen.getByText("目前沒有租戶")).toBeTruthy());
    expect(screen.queryAllByRole("row")).toHaveLength(0);
  });
});
```

（`vi.mocked(...).mockResolvedValue(... as never)` 的 `as never` 是因為 connect 產生的回應型別帶 message wrapper；若實作時採用輕量 client 包裝（建議），改用該包裝的型別即可去掉 cast。）

- [ ] **Step 2: 實作三頁**

```tsx
// TenantsPage：分頁表格（TanStack Table manual）＋關鍵字與狀態篩選。
// 欄位：公司、方案、訂閱狀態、席位用量、當期到期日、待收款。
// 點列 → /tenants/$id。
```

```tsx
// TenantDetailPage：三塊
// ① 訂閱卡（方案、狀態、席位、試用到期、寬限期）＋「記收款」對話框
//    （欄位：期別、金額、方式、交易號、發票號、統編、載具、**原因必填**）
// ② override 清單（feature、值、負責人、到期、原因）＋新增／撤銷
// ③ 該租戶的權益現況（唯讀，顯示 effective 值：方案 ⊕ override）
```

```tsx
// PlansPage：方案清單（分頁）＋價目編輯（月/年、base/seat）＋權益矩陣入口。
// 價目以字串輸入（"1500.00"），前端僅做格式檢查（金額真偽由後端 money.ParseCents 決定）。
```

- [ ] **Step 3: 測試與建置**

Run: `cd platform-console && pnpm test && pnpm typecheck`
Expected: 全綠

- [ ] **Step 4: Commit**

```bash
git add platform-console/src/pages platform-console/src/pages/*.test.tsx
git commit -m "feat(console): 租戶列表、租戶詳情（收款與 override）、方案與價目三頁"
```

---

### Task 12: console 頁面（權益矩陣／收款與發票／平台稽核）

**Files:**
- Create: `platform-console/src/pages/EntitlementsPage.tsx`、`ReceivablesPage.tsx`、`AuditPage.tsx` ＋ 測試

- [ ] **Step 1: 寫測試**

```tsx
import { render, screen, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { platform } from "../lib/api";
import AuditPage from "./AuditPage";
import EntitlementsPage from "./EntitlementsPage";
import ReceivablesPage from "./ReceivablesPage";

vi.mock("../lib/api", () => ({
  platform: { getPlanEntitlements: vi.fn(), listReceivables: vi.fn(), listPlatformAudit: vi.fn() },
}));

function renderWithQuery(ui: () => unknown) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(() => <QueryClientProvider client={client}>{ui()}</QueryClientProvider>);
}

// 矩陣形狀斷言：3 方案 × 8 features → 9 列（含表頭）、每列 4 格（功能名 + 3 方案）。
it("權益矩陣為 8 列 × 3 方案，且未設定的格顯示未開通", async () => {
  vi.mocked(platform.getPlanEntitlements).mockResolvedValue({
    features: [
      { code: "limit.seats", type: "integer", unit: "席", description: "席位" },
      { code: "limit.customers", type: "integer", unit: "客戶", description: "客戶" },
      { code: "limit.products", type: "integer", unit: "商品", description: "商品" },
      { code: "limit.departments", type: "integer", unit: "部門", description: "部門" },
      { code: "limit.storage_gb", type: "integer", unit: "GB", description: "空間" },
      { code: "feature.printing", type: "boolean", unit: "", description: "列印" },
      { code: "feature.dispatch", type: "boolean", unit: "", description: "派車" },
      { code: "feature.returns", type: "boolean", unit: "", description: "退貨" },
    ],
    entitlements: [
      { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: 10 },
      { featureCode: "feature.printing", enabled: false, limitSet: false, limitValue: 0 },
    ],
  } as never);

  renderWithQuery(() => <EntitlementsPage planCodes={["free", "std", "pro"]} />);
  await waitFor(() => expect(screen.getAllByRole("row")).toHaveLength(9));
  for (const row of screen.getAllByRole("row")) {
    expect(row.querySelectorAll("td,th")).toHaveLength(4);
  }
});

// CSV 匯出：斷言表頭與金額格式（金額一律兩位小數字串，不接受 "1500" 這種形式）。
it("匯出 CSV 含表頭與兩位小數金額", async () => {
  vi.mocked(platform.listReceivables).mockResolvedValue({
    rows: [{ companyId: "42", companyName: "甲公司", planCode: "std", periodNo: 2,
             amount: "1500.00", periodEnd: "2026-10-31T00:00:00Z", status: "open" }],
    pagination: { page: 1, pageSize: 50, total: 1 },
  } as never);

  renderWithQuery(() => <ReceivablesPage />);
  const exported: string[] = [];
  // 以 spy 攔下 Blob 內容：實作時 exportCsv 以 Blob([csv]) 觸發下載，此處讀其文字。
  vi.spyOn(URL, "createObjectURL").mockImplementation((blob: Blob) => {
    void blob.text().then((t) => exported.push(t));
    return "blob:mock";
  });

  await waitFor(() => expect(screen.getByText("甲公司")).toBeTruthy());
  screen.getByRole("button", { name: "匯出 CSV" }).click();
  await waitFor(() => expect(exported).toHaveLength(1));
  const lines = exported[0].split("\n");
  expect(lines[0]).toBe("公司,方案,期別,金額,到期日,狀態");
  expect(lines[1]).toContain("1500.00");
});

// 平台稽核：列出 operator、action、目標、原因，並可依目標篩選。
it("平台稽核列出紀錄並可依目標篩選", async () => {
  vi.mocked(platform.listPlatformAudit).mockResolvedValue({
    entries: [
      { id: "1", operatorEmail: "ops@example.com", action: "record_payment",
        targetType: "subscription", targetId: "5", reason: "匯款入帳",
        createdAt: "2026-09-20T10:00:00Z" },
    ],
    pagination: { page: 1, pageSize: 20, total: 1 },
  } as never);

  renderWithQuery(() => <AuditPage />);
  await waitFor(() => expect(screen.getByText("ops@example.com")).toBeTruthy());
  expect(screen.getByText("匯款入帳")).toBeTruthy();
  expect(screen.getAllByRole("row")).toHaveLength(2); // 表頭 + 1 列
});
```

- [ ] **Step 2: 實作三頁**

```tsx
// EntitlementsPage：features（列）× plans（欄）矩陣；格內為限制值或開關；
// 編輯走 SetPlanEntitlement（型別 boolean → 開關；integer → 數字輸入，空 = 不限）。
```

```tsx
// ReceivablesPage：待收款表格 ＋「匯出 CSV」按鈕（前端組 CSV，不需後端端點）。
// 匯出欄位：公司、方案、期別、金額、到期日、狀態。
```

```tsx
// AuditPage：平台稽核清單（分頁）＋ target_type／target_id 篩選。
```

- [ ] **Step 3: 測試與建置**

Run: `cd platform-console && pnpm test && pnpm typecheck && pnpm build`
Expected: 全綠

- [ ] **Step 4: Commit**

```bash
git add platform-console/src/pages
git commit -m "feat(console): 權益矩陣、待收款與 CSV 匯出、平台稽核三頁"
```

---

### Task 13: 租戶後台唯讀權益卡片

**Files:**
- Create: `frontend/src/features/account/PlanCard.tsx`、`PlanCard.test.tsx`、`frontend/src/features/account/queries.ts`
- Modify: `frontend/src/components/layout/AppShell.tsx`（掛載卡片入口）、`frontend/src/lib/proto`（Task 9 已生成）

**Interfaces:**
- Consumes: `TenantEntitlementService.GetTenantEntitlements`（Plan B Task 10）

- [ ] **Step 1: 寫測試**

```tsx
import { render, screen, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { loadTenantEntitlements } from "./queries";
import PlanCard from "./PlanCard";

vi.mock("./queries", () => ({ loadTenantEntitlements: vi.fn() }));

function renderCard() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(() => (
    <QueryClientProvider client={client}>
      <PlanCard />
    </QueryClientProvider>
  ));
}

describe("PlanCard", () => {
  beforeEach(() => {
    vi.mocked(loadTenantEntitlements).mockReset();
  });

  it("用量達 80% 時顯示升級提示", async () => {
    vi.mocked(loadTenantEntitlements).mockResolvedValue({
      planCode: "std", planName: "標準", status: "active", trialEndsAt: "",
      usage: [{ featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: 10, used: 8 }],
    } as never);

    renderCard();
    await waitFor(() => expect(screen.getByText(/8\/10/)).toBeTruthy());
    expect(screen.getByText(/接近上限/)).toBeTruthy();
  });

  it("用量未達 80% 時不顯示提示（卡片不吵）", async () => {
    vi.mocked(loadTenantEntitlements).mockResolvedValue({
      planCode: "std", planName: "標準", status: "active", trialEndsAt: "",
      usage: [{ featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: 10, used: 7 }],
    } as never);

    renderCard();
    await waitFor(() => expect(screen.getByText(/7\/10/)).toBeTruthy());
    expect(screen.queryByText(/接近上限/)).toBeNull();
  });

  it("載入失敗時靜默：不顯示任何限制，也不阻擋操作", async () => {
    vi.mocked(loadTenantEntitlements).mockRejectedValue(new Error("boom"));

    renderCard();
    // 失敗後不得出現誤導性的額度資訊（前端守衛不構成授權：可用性交給後端擋）
    await waitFor(() => expect(screen.queryByText(/\/10/)).toBeNull());
    expect(screen.getByText(/方案資訊暫時無法取得/)).toBeTruthy();
  });
});
```

- [ ] **Step 2: 實作**

```tsx
// PlanCard：顯示方案名稱、狀態、席位用量 bar、試用到期；未達門檻時僅為一行摘要（不吵）。
// 權益載入失敗一律靜默（不擋操作、不顯示誤導性限制）。
```

- [ ] **Step 3: 測試與建置**

Run: `cd frontend && pnpm test && pnpm typecheck && pnpm lint`
Expected: 全綠（**既有 174 條測試不得因新增卡片而紅**）

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/account frontend/src/components/layout/AppShell.tsx
git commit -m "feat(frontend): 租戶後台唯讀權益卡片（用量提示，載入失敗不擋操作）"
```

---

### Task 14: 部署、CI 與慣例文件

**Files:**
- Modify: `.github/workflows/ci.yml`（新增 `platform-console` job）、`backend/AGENTS.md`、`docs/superpowers/plans/README.md`、`.env.example`

- [ ] **Step 1: CI job**

```yaml
  platform-console:
    name: Platform console(SolidJS)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm -C platform-console typecheck
      - run: pnpm -C platform-console lint
      - run: pnpm -C platform-console test
      - run: pnpm -C platform-console build
```

- [ ] **Step 2: 慣例文件（`backend/AGENTS.md` §11 追加）**

（Plan B 的 Task 12 已把平台域寫成 §11；§10 是 Plan D 的「錯誤碼」。）

```markdown
9. **金額一律 `int64` 分**：禁止 float 參與金額運算；DB 邊界用 `internal/platform/money`；金額路徑必附測試。
10. **平台寫入單一交易**：期別／訂閱狀態／事件／稽核同一 commit；因 `platform` 與業務表同庫，跨域副作用（`companies.status`）可同交易完成，**不使用補償式設計**。
11. **收款只有一個入口**：`billing.RecordPayment`。新增金流商＝新增 adapter 呼叫它，不得新增改變訂閱狀態的路徑。
12. **排程單趟可重跑**：`cron.RunOnce(ctx, deps, now)`；`now` 由呼叫端給（`cmd/platform-cron --date`）。重跑不得產生重複期別或事件。
13. **平台操作者與租戶身分不互通**（承 §11-8）：console 只走 `platform/v1`；租戶 SPA 不得掛平台路由或 `platform.*` 能力。
14. **console 不手抄錯誤碼**：`platform-console/src/lib/errcode.ts` 由 `go generate ./internal/errcode` 產生（Plan D T6 的產生器在該目錄存在時即輸出，CI 的「Error codes up to date」已把此路徑納管）。顯示錯誤一律依 `ErrorInfo.code`（`docs/error-codes.md` 為碼表來源），不得硬編訊息字串。
```

- [ ] **Step 3: 更新計畫索引與 `.env.example`**

`plans/README.md` 新增 ③ 一列（狀態 `🟡 計畫已備、未開工`）；`.env.example` 補 `PLATFORM_*` 與 `PLATFORM_SEED_OPERATOR_EMAIL`（已於 Plan B Task 11 加入者不重複）。

- [ ] **Step 4: 跑全部守門**

Run: `cd backend && task check && task test:integration`；`cd .. && pnpm -C frontend test && pnpm -C platform-console test && pnpm -C platform-console build`
Expected: 全綠

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/ci.yml backend/AGENTS.md docs/superpowers/plans/README.md backend/.env.example
git commit -m "ci(console): 平台工具 CI job 與平台域慣例（金額、單一交易、收款入口、排程可重跑）"
```

---

## 驗收對照（spec）

| spec 要求 | 對應任務 |
|---|---|
| 訂閱狀態機與轉移（§5.2） | Task 4（收款復原）、Task 5（逾期／停用） |
| `RecordPayment` 為人工與金流共同入口（§5.2、§8.2） | Task 4、Task 9 |
| 凍結／復原由事件驅動、不直寫產品域（§5.3） | Task 1（狀態入口）、Task 6（consumer） |
| 排程可重跑、期別產生與催收（§5.4、§5.5） | Task 5、Task 7（v1 以清單取代通知） |
| 平台稽核獨立且 `reason` 必填（S9、§3.3） | Task 3、Task 4、Task 9 |
| 權益快取失效不靠 TTL（§4.4） | Task 8 |
| 平台工具六頁與 operator 認證（S7、§2.4） | Task 10–12 |
| 租戶後台唯讀權益卡片（§2.4） | Task 13 |
| 金額不以 float 表示（§3.1 帳務） | Task 2（`money`） |
| `platform.*` 不出現在租戶端（S11） | Task 10（只代理 `/platform`）、Task 13（卡片不引入平台能力） |

## 風險與對策

| # | 風險 | 對策 |
|---|---|---|
| C1 | 排程與人工操作併發（同時收款與停用） | 所有平台寫入以 `OpenSubscriptionTx`（`FOR UPDATE`）鎖訂閱列；Task 3 有列鎖測試 |
| C2 | 事件派送與狀態變更不一致 | 同一個 PostgreSQL 資料庫 → 同交易完成；consumer 失敗則整筆回滾，不標記已派送（Task 6 測試） |
| C3 | 金額運算尾差或溢位 | 一律 int64 分；`money` 有拒絕 >2 位小數、負值與溢位的測試 |
| C4 | console 與租戶 SPA 的 UI 元件庫以 alias 共用，未來改動互相影響 | 目前兩個 app 共用是刻意的（單一視覺權威）；出現第三個 app 或改動開始互相牽制時抽成 `packages/ui`（`ponytail:` 留痕） |
| C5 | 催收無自動通知（07 未實作） | v1 以待收款清單＋前端 CSV 匯出來源；列為已知依賴，不假裝有通知 |
| C6 | `--date` 補跑可能造成重複期別 | `UNIQUE (subscription_id, period_no)` ＋ `ON CONFLICT DO NOTHING`；`EnsureNextPeriod` 先查後建 |

---

*建立：2026-09-20（SaaS 化 spec 的第三份實作計畫：訂閱生命週期、收款與平台營運工具）*
