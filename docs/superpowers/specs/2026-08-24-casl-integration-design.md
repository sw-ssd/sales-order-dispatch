# CASL 整合設計：後端能力模型 + SolidJS 前端整合

> - 日期：2026-08-24
> - 狀態：已核准（brainstorming 對話逐段確認）
> - 範圍：1.0（後端 Phase 1–2、前端 WEB-INF-05 展開）
> - 修訂：D3 三層分工表述（新增第四參與者「CASL 執行層」，不改 RLS/Casbin 角色）；將產生決策記錄 **D30**
> - 參考：[CASL — Ability to database query](https://casl.js.org/v7/en/advanced/ability-to-database-query)、[stalniy/casl#8 sql & sequelize support](https://github.com/stalniy/casl/issues/8)

## 0. 背景與問題

1.0 凍結規格（D3）的權限分工為：Casbin 管後端功能授權、PostgreSQL RLS 管資料範圍、CASL 僅管前端 UI。`02-tenancy-users.md` 2.9.3 明寫「本表僅驅動前端 CASL UI 顯示，後端授權由 Casbin policy 決定」——即**後端沒有 CASL 整合**。

此分工有兩個缺口：

1. **屬性/狀態級條件無處表達**：如「staff 僅能取消 `pending` 訂單」「dept_admin 僅能審核本部門退貨」。Casbin p 規則只到 path × method；RLS 只到 company/department/self 租戶範圍。這類條件目前只能散落各 handler 手寫。
2. **前後端規則分歧風險**：前端 ability 由 `role_permissions` 產生、後端由 Casbin policy 決定，兩份來源可能不一致（前端以為可按、後端拒絕，或反之）。

另發現既有文件錯誤：`2026-08-05-subproject-implementation-plan.md` Task 4 依賴清單列了 **`@casl/solid`，此套件在 npm 不存在**（CASL 官方 monorepo 僅有 ability/angular/mongoose/prisma/react/vue）。前端整合需基於 `@casl/ability` 自寫極薄 Solid binding。

## 1. 決策摘要（將寫入決策記錄為 D30）

| # | 決策 | 一句話 |
|---|---|---|
| D30 | CASL JSON 為第二權限模型 | `role_permissions` 擴充為 CASL 規則表（conditions/inverted/sort_order），前後端同表同源；後端 Go 自研評估器執行 list 過濾與實例檢查；Casbin 與 RLS 保留不動 |
| D30-2 | env 開關 `CASL_ENFORCEMENT_ENABLED` | 全域 server 級；development/test/production 均預設 `true`；關閉 = 降級回純 Casbin+RLS（RLS 仍擋租戶越界，非安全風險） |
| D30-3 | Go 手寫評估器 + golden fixture 對賭 | 不內嵌 JS runtime；運算子白名單；CI 以真 `@casl/ability` 產生 fixture、Go 測試重放比對 |
| D30-4 | 前端用 `@casl/ability` + 自寫 binding | 無 `@casl/solid`；context + signal + `Can` 元件 + route guard；修正既有計畫文件錯誤依賴 |

已考慮 alternative：
- **goja 內嵌 JS runtime 跑真 @casl/ability**：語意 100% 一致，但每請求進 JS VM，效能/部署複雜度高 → 否決。
- **Ent interceptor 全域自動套用查詢條件**：自動化但難表達 action 語意、隱式行為除錯與測試成本高 → 否決，採 repository 顯式套用。
- **獨立 casl_rules 表與 role_permissions 並存**：兩份權限來源語意分歧，違背同源精神 → 否決。
- **欄位級遮罩（fields/rulesToFields）**：每個回應路徑需過遮罩器、與 proto 型別整合成本高 → 1.0 不做（YAGNI）。

## 2. 規則模型與 Schema

### 2.1 `role_permissions` 擴充為 CASL 規則表

| 欄位 | 型別 | 說明 |
|---|---|---|
| `id` | UUID | 沿用 |
| `role_id` | FK → roles | 沿用 |
| `action` | text | 沿用；動詞表與 Casbin `act` 同源（`create`/`read`/`update`/`delete`/`print`/`dispatch`/`manage`…）；CASL 保留字 `manage` = 任意 action |
| `resource` | text | 沿用；即 CASL `subject`（`sales_order`/`customer`/`product`/`dispatch`/`print`/`user`/`company`…）；保留字 `all` = 任意 subject |
| `conditions` | JSONB，nullable | CASL 條件物件；僅平鋪欄位 + 白名單運算子；`${user.*}` 佔位符執行期以身分注入 |
| `inverted` | bool，預設 false | true = CASL `cannot` |
| `sort_order` | int | 規則順序顯式持久化；CASL 語意由後往前比對、先命中者決定，慣例 cannot 排在 can 之後 |

唯一鍵由 `role_id + resource + action` 放寬為 `role_id + resource + action + conditions 正規化 hash`——同一 resource×action 允許多條不同條件規則（如「可讀本部門」+「可讀本人建立」）。

conditions 範例：

```json
{"status": {"$in": ["pending", "processing"]}, "department_id": "${user.department_id}"}
```

### 2.2 三層分工新表述（D3 修訂）

| 層 | 機制 | 職責 | 變更 |
|---|---|---|---|
| 進入點 | Casbin RBAC with domain | 角色 × company × RPC path × method | 不變 |
| **屬性條件** | **CASL 執行層（Go 評估器）** | **resource 級屬性/狀態條件：list 過濾 + 單筆實例檢查** | **新增** |
| 資料範圍 | PostgreSQL RLS | 租戶隔離最後防線（all/company/department/self） | 不變 |
| UI | 前端 CASL | 同一份規則 JSON 驅動路由守衛與顯示控制 | 擴充（rules 現含 conditions，前端可做更準的顯示判斷，如依訂單狀態灰化按鈕） |

`role_permissions` 不再「僅驅動前端」——它是前後端同一份權限真相來源。2.9.3 既有註記隨計畫修正移除。

## 3. Go 引擎（`backend/internal/authz/casl` 套件）

### 3.1 介面

```go
package casl

type Rule struct {
    Action     string
    Subject    string
    Conditions map[string]any // 已解析 JSONB；nil = 無條件
    Inverted   bool
}

type Evaluator struct { /* rules（已依 sort_order 排序）+ identity 展開後的快照 */ }

// Can：由後往前掃規則，先命中者決定；manage/all 保留字展開；無命中 = deny
func (e *Evaluator) Can(action, subject string, instance map[string]any) bool

// Translate：rulesToQuery 對應物。
// 直接規則條件 OR 併；inverted 規則條件 AND NOT 併；
// 無任何允許規則 → denied=true（對應 CASL 回 null：查詢短路回空集，不打 DB）
func Translate(rules []Rule, action, subject string, reg *FieldRegistry) (clause string, args []any, denied bool, err error)

// FieldRegistry：subject → 邏輯欄位 → {DB column, 允許運算子, 值型別, instance extractor}
// 三個消費者共用一份定義：SQL 翻譯、instance 值擷取、UI 條件建構器白名單 API
type FieldRegistry struct { /* ... */ }
```

### 3.2 語意規則

- 運算子白名單：`$eq` `$ne` `$in` `$nin` `$lt` `$lte` `$gt` `$gte`；裸值視為 `$eq`。
- 僅平鋪欄位，不支援 dot-notation 巢狀（1.0 無需求）。
- 未知運算子 / 未知欄位 / 值型別不符 → 該規則視為不命中並記 security log（fail-closed）。
- 佔位符 `${user.company_id}` / `${user.department_id}` / `${user.customer_id}` / `${user.id}`：規則自 DB 載入組裝 Evaluator 時以當前身分展開；展開失敗（如客戶帳號無 department_id）→ 該規則不命中。
- 規則順序：依 `sort_order` 升冪存入、比對時由後往前掃（CASL 慣例：後定義者優先，cannot 排最後）。
- 多角色：取全部角色規則聯集後排序合併，語意不變。

### 3.3 語意對賭（golden fixture）

- `backend/internal/authz/casl/testdata/gen.mjs`：Node 端以真 `@casl/ability`（含 `rulesToQuery` from `@casl/ability/extra`）對固定 cases（規則集 × action × subject × instance）產生 `cases.json`（預期 `can()` 布林與查詢結構）。
- Go 測試重放同一 `cases.json` 比對 `Can` 與 `Translate` 輸出。
- fixture 由開發者在本機重新產生並隨 PR 提交；CI 驗證 fixture 為最新（重跑 gen 後 `git diff --exit-code`）。monorepo 本有 pnpm/Node，無額外基礎設施。

### 3.4 呼叫面

- **list 過濾**：repository 層顯式套用 `authz.AccessibleFilter(ctx, "read", "sales_order")` → `(clause, args)`，以 Ent `Where(func(s *sql.Selector) { s.Where(sql.ExprP(clause, args...)) })` 注入；`denied=true` → 短路回空集。
- **實例檢查**：usecase 層取出實體後 `authz.Can(ctx, "cancel", "sales_order", order)`，FieldRegistry extractor 由實體擷取條件所需欄位值。
- 遺漏套用的風險可接受：RLS 仍擋租戶越界（不會變跨公司洩漏），僅該資源屬性條件未生效；整合測試逐端點覆蓋（§7.3）。

## 4. env 開關行為

- config 欄位 `CASLEnforcementEnabled`（env `CASL_ENFORCEMENT_ENABLED`）：**development / test / production 均預設 `true`**。回退路徑 = 設 `false` 重新部署（已接受的代價）。
- 關閉時：`authz.AccessibleFilter` 回傳空條件、`authz.Can` 直接放行 → 等效現行凍結規格（Casbin + RLS）。`GetAbility` **不受影響**，前端顯示照常；此時前後端判斷可能暫時分歧，屬文件化的降級模式。
- 無 fail-fast（與 `DEVELOPER_ACCOUNT_ENABLED` 不同：關閉 CASL 非安全風險，RLS 仍在）。啟動 log 記錄開關狀態，供部署檢查。
- 開關於 middleware 注入身分段讀取一次放入 ctx；repository / usecase 不直接讀 config。

## 5. 前端 SolidJS 整合

### 5.1 結構（`frontend/src/lib/ability/`）

| 檔案 | 職責 |
|---|---|
| `context.tsx` | `AbilityContext` + `useAbility()`；signal 持有 `AppAbility` 實例（`@casl/ability`） |
| `service.ts` | `GetAbility` Connect client 封裝；`createAppAbility(rules)` 建實例（規則型別對齊 proto `AbilityRule`） |
| `Can.tsx` | `<Can I="update" a="sales_order" fallback={...}>`；可傳 instance 做條件判斷（如依訂單狀態灰化） |
| `guards.ts` | `requireAbility(action, subject)` route guard：`@solidjs/router` `beforeLoad` 呼叫，deny → redirect `/403` |

### 5.2 生命週期

- 登入成功後取一次 `GetAbility`；TanStack Query 承載快取（`staleTime: 60_000`，對齊規格 60 秒 TTL，不重造快取輪子）；路由切換不重取。
- 權限異動後 `queryClient.invalidateQueries(['ability'])` 主動重載（權限設置頁儲存後、或收到異動通知時）。
- 更新語意 = **整個 Ability 實例替換**（signal set）→ Solid 細粒度響應自動重算所有 `can()` 引用點，無需手動訂閱。
- App（Flutter）維持簡單 role 判斷，不載入 ability（既有規格不變）。

### 5.3 既有文件修正

- `2026-08-05-subproject-implementation-plan.md` Task 4 Step 1：移除不存在的 `@casl/solid`，依賴改 `@casl/ability`。
- `2026-08-05-subproject-decomposition-design.md` WEB-INF-05 展開為上述四檔結構。

## 6. 權限設置頁：條件建構器

- 「角色權限設置」頁（既有 2.9.3 矩陣）每個 resource×action 格子可展開編輯 conditions：
  - 欄位下拉：由後端 `RoleService.ListConditionFields(resource)` 提供（FieldRegistry 白名單）。
  - 運算子下拉：依欄位型別過濾（enum/id 給 `$eq/$ne/$in/$nin`；日期/數值加 `$lt/$lte/$gt/$gte`）。
  - 值輸入：enum 欄位給選項；id 欄位給 `${user.*}` 佔位符選項 + 明示值輸入。
  - **不開放自由 JSON**。
- `inverted` 規則於 UI 以「拒絕規則」呈現；同一格可加多條規則（對應 §2.1 放寬的唯一鍵）。
- 寫入走既有 `UpdatePermissions` 全量覆寫；後端驗證未知欄位/運算子/值型別 → `invalid_argument`。
- **防鎖死延伸**：操作者自身角色的「權限管理」resource 不允許加上會排除自己的 conditions（以操作者身分代入驗證仍可 manage）；`company_admin` 的 conditions 強制限縮於自己公司（id 類欄位值須為 `${user.company_id}` 佔位符或自己公司範圍內的值）。
- 稽核快照含 conditions 前後值（D18 同事務沿用）。

## 7. 錯誤處理與測試

### 7.1 錯誤處理

- 規則解析失敗 / 未知運算子 / 未知欄位 → fail-closed（該規則不命中 + security log）。
- `Translate` denied → repository 短路回空集，不打 DB。
- Evaluator 為每請求組裝（§3.2）；引擎本體於啟動期以 fixture 全量自驗（init 自檢，失敗 fail-fast）。執行期錯誤視為 bug，不收復 panic。

### 7.2 單元測試（Go）

條件 AST 解析、佔位符展開（含客戶帳號無 department_id）、規則順序與 cannot 優先、inverted 聚合、denied 短路、`manage`/`all` 保留字、未知運算子 fail-closed。

### 7.3 整合測試（dockertest，D21 關鍵路徑）

1. env on/off 各一輪：list 過濾生效 vs 繞過、實例檢查攔截越權寫入（如 staff 取消 `processing` 訂單被拒）。
2. CASL 條件 × RLS 疊加結果正確（交集語意）。
3. 條件建構器寫入 → `UpdatePermissions` 驗證 → `GetAbility` → 規則含 conditions 全鏈。
4. 防鎖死：為自身角色權限管理加排除性 conditions 被拒。

### 7.4 前端測試（Vitest）

`Can` 顯示/隱藏與 fallback、route guard redirect、TTL 內不重取、invalidate 後重取。

## 8. 文件同步交付清單

| 文件 | 動作 |
|---|---|
| 本文件 | 新增 |
| `2026-07-19-sales-order-1.0-decisions.md` | 新增 D30/D30-2/D30-3/D30-4；D3 補註修訂（比照 D14 模式加註修訂來源） |
| `1.0-requirements/authorization/spec.md` | 「CASL 前端能力」requirement 擴充後端執行層 scenario（list 過濾、實例檢查、env 開關、同源） |
| `plans/backend/detail/01-auth.md` | 1.8.1：規則格式含 conditions/inverted/sort_order；詞彙表補 manage/all 保留字 |
| `plans/backend/detail/02-tenancy-users.md` | 2.9.1 schema 加三欄與新唯一鍵；2.9.3 移除「僅驅動前端」註記、加 conditions 驗證與防鎖死延伸；2.9.4 規則輸出含 conditions |
| `2026-08-05-subproject-implementation-plan.md` | Task 4 依賴修正（移除 `@casl/solid`） |
| `2026-08-05-subproject-decomposition-design.md` | WEB-INF-05 展開四檔結構 |
| 客戶版規格書 `2026-07-16-sales-order-1.0-design.md` | §3.4 修訂、升版 v1.0.35 並補修訂記錄 |

## 9. 非目標（1.0 不做）

- 欄位級遮罩（fields / rulesToFields）。
- dot-notation 巢狀條件、自訂運算子。
- 逐公司開關（companies.capabilities）。
- App 端 CASL（維持簡單 role 判斷）。
- 以 CASL 取代 Casbin 或 RLS。

---

*最後更新：2026-08-24*
