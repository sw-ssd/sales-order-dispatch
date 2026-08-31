# CASL 整合（後端能力模型 + SolidJS 前端）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 以 CASL JSON 為前後端同源的第二權限模型——後端 Go 自研評估器執行 list 查詢過濾與單筆實例檢查（env 開關控制），前端 SolidJS 以 `@casl/ability` 自寫 binding 驅動路由守衛與顯示控制；Casbin 與 RLS 不動。

**Architecture:** `role_permissions` 擴充為 CASL 規則表（conditions/inverted/sort_order），同時驅動 `GetAbility`（前端）與 `authz/casl` 引擎（後端）。後端引擎三件套：條件 AST parser → Evaluator（`Can`，反向掃規則）→ Translate（rules→SQL WHERE，對應 `rulesToQuery`）；FieldRegistry 一份定義供 SQL 翻譯、實例擷取、UI 條件建構器白名單三方消費。語意正確性以 golden fixture 對賭（真 `@casl/ability` 產生、Go 重放）。

**Tech Stack:** Go 1.25、Ent、pgx/v5、Connect-RPC、testcontainers-go；前端 SolidJS 1.9 + `@casl/ability` + `@tanstack/solid-query` + `@solidjs/router` + Vitest；golden 產生器 Node（frontend workspace 內）。

**Spec 來源:** `docs/superpowers/specs/2026-08-24-casl-integration-design.md`（下稱「設計」）；決策 D30 系列見 Task 1 寫入內容。

**前置相依:** `2026-08-17-backend-01-auth-plan.md`（middleware/identity ctx、Task 1.8 GetAbility）、`2026-08-17-backend-02-tenancy-users-plan.md`（roles/role_permissions schema、2.9.3 UpdatePermissions）。本計畫 Task 1 先修訂那些計畫文件使契約對齊；程式碼任務假設其產物（identity ctx、Ent client、proto 工具鏈）已存在。

## Global Constraints

- module 路徑：`github.com/salesorder/sales-order-1.0/backend`；前端路徑別名 `~/*` → `frontend/src/*`。
- 運算子白名單僅：`$eq` `$ne` `$in` `$nin` `$lt` `$lte` `$gt` `$gte`；裸值視為 `$eq`；僅平鋪欄位，無 dot-notation。
- fail-closed：未知運算子/未知欄位/值型別不符 → 該規則不命中 + security log；`Translate` 無允許規則 → `denied=true`，repository 短路回空集。
- env 開關 `CASL_ENFORCEMENT_ENABLED`：dev/test/prod 均預設 `true`；關閉時 `AccessibleFilter` 回空條件、`Can` 放行；`GetAbility` 不受影響。
- 規則順序：`sort_order` 升冪存入、比對由後往前掃，先命中者決定（cannot 慣例排最後）。
- 佔位符：`${user.id}` `${user.company_id}` `${user.department_id}` `${user.customer_id}`；展開失敗 → 規則不命中。
- 詞彙同源：action 動詞沿用 Casbin `act`（`create`/`read`/`update`/`delete`/`print`/`dispatch`/`manage`…）；subject 沿用 `role_permissions.resource`；CASL 保留字 `manage`=任意 action、`all`=任意 subject。
- SQL 片段用 `?` 佔位（Ent `sql.ExprP` 會依 dialect 重寫為 `$N`）。
- 每個 Task 結尾 commit；commit message：`feat(backend): …` / `feat(frontend): …` / `docs: …` / `test(…): …`。
- 測試：Go 走 testify + testcontainers Postgres 16；前端 Vitest + jsdom；`go test ./...` 與 `pnpm run test` 全綠才可進下一 Task。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/internal/authz/casl/condition.go` | conditions JSONB → FieldCondition AST | Task 2 |
| `backend/internal/authz/casl/evaluator.go` | Rule/Evaluator/`Can`；佔位符展開 | Task 3 |
| `frontend/scripts/casl-golden-gen.mjs` + `backend/internal/authz/casl/testdata/cases.json` | golden fixture 產生器與資料 | Task 4 |
| `backend/internal/authz/casl/translate.go` | rules→SQL WHERE（rulesToQuery 對應物） | Task 5 |
| `backend/internal/authz/casl/registry.go` | FieldRegistry（subject→欄位定義/擷取器/白名單） | Task 6 |
| `backend/ent/schema/rolepermission.go` + migration | role_permissions 擴充三欄與唯一鍵 | Task 7 |
| `backend/internal/authz/access.go` | facade：`AccessibleFilter`/`Can`；env 開關 ctx | Task 8 |
| `backend/proto/v1/ability.proto` + `backend/internal/domain/auth/ability.go` | AbilityRule 含 conditions/inverted；GetAbility 改由規則表產生 | Task 9 |
| `backend/internal/domain/roles/{usecase,handler}.go` | UpdatePermissions conditions 驗證、防鎖死延伸、ListConditionFields | Task 10 |
| `frontend/src/lib/ability/{context.tsx,service.ts}` | Ability context + GetAbility query | Task 11 |
| `frontend/src/lib/ability/Can.tsx` | `<Can>` 顯示控制元件 | Task 12 |
| `frontend/src/lib/ability/guards.ts` | `requireAbility` route guard + 快取接線 | Task 13 |

---

### Task 1: 文件同步（決策 D30、需求規格、既有計畫修訂）

**Files:**
- Modify: `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（新增 D30 區塊；D3 加修訂補註）
- Modify: `docs/superpowers/specs/1.0-requirements/authorization/spec.md`（「CASL 前端能力」requirement 擴充）
- Modify: `docs/superpowers/plans/backend/detail/01-auth.md`（1.8.1 規則格式）
- Modify: `docs/superpowers/plans/backend/detail/02-tenancy-users.md`（2.9.1/2.9.3/2.9.4）
- Modify: `docs/superpowers/plans/2026-08-05-sales-order-1.0-subproject-implementation-plan.md`（Task 4 依賴修正）
- Modify: `docs/superpowers/specs/2026-08-05-sales-order-1.0-subproject-decomposition-design.md`（WEB-INF-05 展開）
- Modify: `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（§3.4 修訂、升版 v1.0.35）

**Interfaces:**
- Produces: 下游所有 Task 的契約基準；無程式碼產物。

- [ ] **Step 1: 決策記錄新增 D30（檔尾追加）**

```markdown
### D30：CASL JSON 為第二權限模型（後端執行層 + 前端同源）

- **選擇**：`role_permissions` 擴充為 CASL 規則表（`conditions` JSONB、`inverted`、`sort_order`），前後端同表同源；後端 Go 自研評估器執行 list 查詢過濾與單筆實例 `can(action, instance)` 檢查；Casbin（RPC 進入點）與 RLS（租戶最後防線）保留不動。
- **D30-2 env 開關**：`CASL_ENFORCEMENT_ENABLED` 全域 server 級，dev/test/prod 均預設 `true`；關閉 = 降級回純 Casbin+RLS（RLS 仍擋租戶越界），`GetAbility` 不受影響。無 fail-fast；啟動 log 記錄狀態。
- **D30-3 Go 手寫評估器 + golden fixture 對賭**：不內嵌 JS runtime；運算子白名單 `$eq/$ne/$in/$nin/$lt/$lte/$gt/$gte`、平鋪欄位；CI 以真 `@casl/ability` 產生 fixture、Go 測試重放比對。
- **D30-4 前端**：`@casl/ability` + 自寫 Solid binding（`@casl/solid` 不存在於 npm）；修正 `2026-08-05-subproject-implementation-plan.md` Task 4 錯誤依賴。
- **理由**：Casbin/RLS 無法表達屬性/狀態級條件（如「staff 僅能取消 pending 訂單」）；前後端各一份權限來源有分歧風險。CASL JSON 為 isomorphic 格式，天然可跨端共享。
- **已考慮 alternative**：goja 內嵌 JS runtime（效能/部署成本）；Ent interceptor 全域自動套用（難表達 action 語意、隱式難測）；獨立 casl_rules 表（兩份來源分歧）；欄位級遮罩 fields/rulesToFields（1.0 YAGNI）。
- **修訂來源**：2026-08-24 設計文件 `docs/superpowers/specs/2026-08-24-casl-integration-design.md`。D3 三層分工隨之改為四參與者：Casbin（進入點）/ CASL 執行層（屬性條件）/ RLS（資料範圍）/ 前端 CASL（UI）。
```

並於 D3 條目尾加一行補註：

```markdown
- **修訂（2026-08-24, D30）**：新增 CASL 執行層管 resource 屬性/狀態條件；Casbin 與 RLS 職責不變。
```

- [ ] **Step 2: authorization/spec.md「CASL 前端能力」requirement 擴充**

在該 Requirement 既有文字後追加（Scenarios 之前）：

```markdown
後端 MUST 以同一份 `role_permissions` 規則（含 `conditions`、`inverted`）在 `CASL_ENFORCEMENT_ENABLED = true` 時執行：list 查詢依規則產生 SQL 過濾條件（無允許規則時回空集）；寫入/更新/取消等操作對實體做 `can(action, instance)` 條件檢查。開關關閉時後端跳過 CASL 執行層，行為等同 Casbin + RLS（RLS 仍為租戶隔離最後防線）。
```

新增 Scenarios（接在該 Requirement 既有 Scenario 之後）：

```markdown
#### Scenario: list 查詢套用規則條件

- **WHEN** `staff` 角色的 `role_permissions` 含 `read sales_order conditions={"department_id": "${user.department_id}"}`，且開關啟用
- **THEN** 該使用者 list 訂單僅回符合條件的資料，且結果仍受 RLS 限制（交集語意）

#### Scenario: 實例檢查攔截狀態越權

- **WHEN** 規則為 `cancel sales_order conditions={"status": "pending"}`，staff 嘗試取消 `processing` 訂單，且開關啟用
- **THEN** 後端回 `permission_denied`，不執行取消

#### Scenario: 開關關閉時降級

- **WHEN** `CASL_ENFORCEMENT_ENABLED = false`
- **THEN** list 查詢不附加規則條件、實例檢查放行，Casbin 與 RLS 行為不變；`GetAbility` 回傳不受影響

#### Scenario: 規則異動前後端同源生效

- **WHEN** `super` 於權限設置頁修改某角色規則的 conditions
- **THEN** 下一次 `GetAbility`（前端）與下一次 list/實例檢查（後端）皆反映新條件，不需重啟
```

- [ ] **Step 3: backend/detail/01-auth.md 修訂 1.8.1**

1.8.1「實作邏輯」第 1 點整點取代為：

```markdown
  1. 規則來源為 `role_permissions`（2.9.4 接管後）；Phase 1 無表時以硬編碼預設矩陣產生。規則格式：`{action, subject, conditions?, inverted?}`，陣列順序 = `sort_order` 升冪；conditions 帶 `${user.company_id}` 等佔位符，由前端 CASL 直接消費（佔位符於後端產生時已以身分展開為具體值）。
```

「介面」列取代為：

```markdown
- **介面**: Connect-RPC `AbilityService.GetAbility() → { rules: [{action, subject, conditions?, inverted?}] }`（CASL.js 可消費的 JSON；保留字 `manage`/`all`）。
```

- [ ] **Step 4: backend/detail/02-tenancy-users.md 修訂 2.9.x**

- 2.9.1 `role_permissions` 欄位列改為：`` `role_permissions`:id、`role_id`、`resource`、`action`、`conditions`(JSONB 可空)、`inverted`(bool 預設 false)、`sort_order`(int 預設 0)；唯一索引 `(role_id, resource, action, COALESCE(md5(conditions::text), ''))`（同 resource×action 允許多條件不同規則）。 ``
- 2.9.3 實作邏輯第 3 點整點取代為：

```markdown
  3. 內建角色的功能權允許調整(D9)。`role_permissions` 同時驅動前端 CASL 顯示與後端 CASL 執行層(D30):conditions 寫入前以 FieldRegistry 白名單驗證(未知欄位/運算子/值型別 → `invalid_argument`);防鎖死延伸 — 操作者自身角色的權限管理 resource 不得加會排除自己的 conditions(以操作者身分代入驗證仍可 manage);`company_admin` 的 id 類條件值限 `${user.company_id}` 佔位符或自己公司範圍。
```

- 2.9.4 實作邏輯第 1 點末尾追加：「；規則輸出含 `conditions`（佔位符已展開）與 `inverted`，依 `sort_order` 排序」。

- [ ] **Step 5: subproject-implementation-plan.md Task 4 依賴修正**

Step 1 依賴字串中的 `` `casl/ability`, `@casl/solid`, `` 取代為 `` `@casl/ability`, ``。

- [ ] **Step 6: subproject-decomposition-design.md WEB-INF-05 展開**

WEB-INF-05 列改為：

```markdown
| WEB-INF-05 | CASL ability 與 UI 權限遮蔽（`frontend/src/lib/ability/`：`context.tsx` Provider/useAbility、`service.ts` GetAbility query + createAppAbility、`Can.tsx` 顯示控制、`guards.ts` requireAbility 路由守衛；依賴 `@casl/ability`，無 `@casl/solid`） |
```

- [ ] **Step 7: 客戶版規格書 §3.4 修訂 + 升版 v1.0.35**

§3.4（前端權限 CASL 段落）追加一句：「同一份角色規則（含條件）亦由後端執行層用於列表過濾與操作檢查，可全域開關控制（預設啟用），關閉時仍以 API 權限與資料範圍隔離把關。」；§18 修訂記錄追加 v1.0.35 列：`v1.0.35 | 2026-08-24 | §3.4 CASL 規則前後端同源、後端執行層（D30）`。

- [ ] **Step 8: Commit**

```bash
git add docs/
git commit -m "docs: D30 CASL 第二權限模型；需求規格/細部計畫/子專案計畫契約對齊；規格書 v1.0.35"
```

---

### Task 2: casl 條件 AST parser

**Files:**
- Create: `backend/internal/authz/casl/condition.go`
- Test: `backend/internal/authz/casl/condition_test.go`

**Interfaces:**
- Produces: `casl.Op`（`OpEq/OpNe/OpIn/OpNin/OpLt/OpLte/OpGt/OpGte`）、`casl.FieldCondition{Field string; Op Op; Value any}`、`casl.ParseConditions(raw map[string]any) ([]FieldCondition, error)`。未知運算子回 error（呼叫面 fail-closed）。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/authz/casl/condition_test.go`:

```go
package casl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConditions(t *testing.T) {
	t.Run("裸值視為 $eq", func(t *testing.T) {
		got, err := ParseConditions(map[string]any{"status": "pending"})
		require.NoError(t, err)
		assert.Equal(t, []FieldCondition{{Field: "status", Op: OpEq, Value: "pending"}}, got)
	})
	t.Run("運算子物件展開，同欄位多運算子排序決定性", func(t *testing.T) {
		got, err := ParseConditions(map[string]any{"qty": map[string]any{"$lt": 5.0, "$gte": 1.0}})
		require.NoError(t, err)
		assert.Equal(t, []FieldCondition{
			{Field: "qty", Op: OpGte, Value: 1.0},
			{Field: "qty", Op: OpLt, Value: 5.0},
		}, got)
	})
	t.Run("$in 接受陣列", func(t *testing.T) {
		got, err := ParseConditions(map[string]any{"status": map[string]any{"$in": []any{"pending", "processing"}}})
		require.NoError(t, err)
		assert.Equal(t, OpIn, got[0].Op)
		assert.Equal(t, []any{"pending", "processing"}, got[0].Value)
	})
	t.Run("$in 非陣列回錯", func(t *testing.T) {
		_, err := ParseConditions(map[string]any{"status": map[string]any{"$in": "pending"}})
		require.Error(t, err)
	})
	t.Run("未知運算子回錯", func(t *testing.T) {
		_, err := ParseConditions(map[string]any{"x": map[string]any{"$regex": ".*"}})
		require.Error(t, err)
	})
	t.Run("nil/空 map 回 nil", func(t *testing.T) {
		got, err := ParseConditions(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/authz/casl/ -run TestParseConditions -v`
Expected: FAIL — `undefined: ParseConditions`

- [ ] **Step 3: 實作**

`backend/internal/authz/casl/condition.go`:

```go
// Package casl 為 CASL JSON 規則的 Go 評估器與 SQL 翻譯器(D30-3)。
// 語意對齊 @casl/ability,由 testdata/cases.json golden fixture 對賭。
package casl

import (
	"fmt"
	"sort"
)

// Op 為白名單運算子。
type Op string

const (
	OpEq  Op = "$eq"
	OpNe  Op = "$ne"
	OpIn  Op = "$in"
	OpNin Op = "$nin"
	OpLt  Op = "$lt"
	OpLte Op = "$lte"
	OpGt  Op = "$gt"
	OpGte Op = "$gte"
)

var validOps = map[Op]bool{
	OpEq: true, OpNe: true, OpIn: true, OpNin: true,
	OpLt: true, OpLte: true, OpGt: true, OpGte: true,
}

// FieldCondition 為單一欄位的一個條件。
type FieldCondition struct {
	Field string
	Op    Op
	Value any // $in/$nin 時為 []any
}

// ParseConditions 將已 unmarshal 的 conditions 物件解析為欄位條件切片。
// 裸值視為 $eq;輸出依 (field, op) 排序保證決定性。
func ParseConditions(raw map[string]any) ([]FieldCondition, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	fields := make([]string, 0, len(raw))
	for f := range raw {
		fields = append(fields, f)
	}
	sort.Strings(fields)

	var out []FieldCondition
	for _, f := range fields {
		switch v := raw[f].(type) {
		case map[string]any:
			ops := make([]string, 0, len(v))
			for op := range v {
				ops = append(ops, op)
			}
			sort.Strings(ops)
			for _, op := range ops {
				o := Op(op)
				if !validOps[o] {
					return nil, fmt.Errorf("casl: unknown operator %q on field %q", op, f)
				}
				val := v[op]
				if o == OpIn || o == OpNin {
					arr, ok := val.([]any)
					if !ok {
						return nil, fmt.Errorf("casl: %s on field %q requires array value", op, f)
					}
					val = arr
				}
				out = append(out, FieldCondition{Field: f, Op: o, Value: val})
			}
		default:
			out = append(out, FieldCondition{Field: f, Op: OpEq, Value: v})
		}
	}
	return out, nil
}
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/authz/casl/ -run TestParseConditions -v`
Expected: PASS（6 子測試全綠）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/authz/casl/
git commit -m "feat(backend): casl 條件 AST parser(運算子白名單)"
```

---

### Task 3: Evaluator（Can + 佔位符展開）

**Files:**
- Create: `backend/internal/authz/casl/evaluator.go`
- Test: `backend/internal/authz/casl/evaluator_test.go`

**Interfaces:**
- Consumes: Task 2 的 `FieldCondition`/`ParseConditions`。
- Produces: `casl.Rule{Action, Subject string; Conditions []FieldCondition; Inverted bool}`、`casl.Identity{UserID, CompanyID, DepartmentID, CustomerID string}`、`casl.NewEvaluator(rules []Rule, id Identity) *Evaluator`（佔位符於此展開；展開失敗規則停用）、`(e *Evaluator) Can(action, subject string, instance map[string]any) bool`。instance 為 nil 時帶條件規則不命中（對齊 CASL type-level 檢查）。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/authz/casl/evaluator_test.go`:

```go
package casl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var staffIdentity = Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1"}

func mustConds(t *testing.T, raw map[string]any) []FieldCondition {
	t.Helper()
	c, err := ParseConditions(raw)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEvaluatorCan(t *testing.T) {
	t.Run("無規則 deny", func(t *testing.T) {
		e := NewEvaluator(nil, staffIdentity)
		assert.False(t, e.Can("read", "sales_order", nil))
	})
	t.Run("無條件規則命中", func(t *testing.T) {
		e := NewEvaluator([]Rule{{Action: "read", Subject: "sales_order"}}, staffIdentity)
		assert.True(t, e.Can("read", "sales_order", nil))
		assert.False(t, e.Can("update", "sales_order", nil))
	})
	t.Run("manage/all 保留字", func(t *testing.T) {
		e := NewEvaluator([]Rule{{Action: "manage", Subject: "all"}}, staffIdentity)
		assert.True(t, e.Can("delete", "customer", map[string]any{"x": 1.0}))
	})
	t.Run("instance 條件比對", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "cancel", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": "pending"}),
		}}, staffIdentity)
		assert.True(t, e.Can("cancel", "sales_order", map[string]any{"status": "pending"}))
		assert.False(t, e.Can("cancel", "sales_order", map[string]any{"status": "processing"}))
	})
	t.Run("type-level 檢查忽略帶條件規則", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "cancel", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": "pending"}),
		}}, staffIdentity)
		assert.False(t, e.Can("cancel", "sales_order", nil))
	})
	t.Run("cannot 排後者優先", func(t *testing.T) {
		e := NewEvaluator([]Rule{
			{Action: "read", Subject: "sales_order"},
			{Action: "read", Subject: "sales_order", Inverted: true,
				Conditions: mustConds(t, map[string]any{"status": "voided"})},
		}, staffIdentity)
		assert.True(t, e.Can("read", "sales_order", map[string]any{"status": "pending"}))
		assert.False(t, e.Can("read", "sales_order", map[string]any{"status": "voided"}))
	})
	t.Run("佔位符以身分展開", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"department_id": "${user.department_id}"}),
		}}, staffIdentity)
		assert.True(t, e.Can("read", "sales_order", map[string]any{"department_id": "d1"}))
		assert.False(t, e.Can("read", "sales_order", map[string]any{"department_id": "d2"}))
	})
	t.Run("佔位符展開失敗規則停用", func(t *testing.T) {
		customer := Identity{UserID: "u9", CompanyID: "c1", CustomerID: "cu1"} // 無 department
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"department_id": "${user.department_id}"}),
		}}, customer)
		assert.False(t, e.Can("read", "sales_order", map[string]any{"department_id": "d1"}))
	})
	t.Run("instance 缺欄位不命中", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": "pending"}),
		}}, staffIdentity)
		assert.False(t, e.Can("read", "sales_order", map[string]any{"other": 1.0}))
	})
	t.Run("$in 比對", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": map[string]any{"$in": []any{"pending", "processing"}}}),
		}}, staffIdentity)
		assert.True(t, e.Can("read", "sales_order", map[string]any{"status": "processing"}))
		assert.False(t, e.Can("read", "sales_order", map[string]any{"status": "completed"}))
	})
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/authz/casl/ -run TestEvaluatorCan -v`
Expected: FAIL — `undefined: NewEvaluator`

- [ ] **Step 3: 實作**

`backend/internal/authz/casl/evaluator.go`:

```go
package casl

import (
	"reflect"
	"strings"
)

// Rule 為一條已解析的 CASL 規則。
type Rule struct {
	Action     string
	Subject    string
	Conditions []FieldCondition // nil = 無條件
	Inverted   bool            // true = cannot
}

// Identity 為佔位符展開用的當前身分。
type Identity struct {
	UserID       string
	CompanyID    string
	DepartmentID string
	CustomerID   string
}

const (
	actionManage = "manage"
	subjectAll   = "all"
)

var placeholders = map[string]func(Identity) string{
	"${user.id}":            func(i Identity) string { return i.UserID },
	"${user.company_id}":    func(i Identity) string { return i.CompanyID },
	"${user.department_id}": func(i Identity) string { return i.DepartmentID },
	"${user.customer_id}":   func(i Identity) string { return i.CustomerID },
}

// Evaluator 持有以身分展開後的規則快照;規則須依 sort_order 升冪傳入。
type Evaluator struct {
	rules []Rule // disabled 規則以 disabledRule 標記,永不命中
}

// NewEvaluator 組裝 Evaluator;佔位符展開失敗(身分缺對應值)的規則停用。
func NewEvaluator(rules []Rule, id Identity) *Evaluator {
	out := make([]Rule, len(rules))
	for i, r := range rules {
		out[i] = Rule{Action: r.Action, Subject: r.Subject, Inverted: r.Inverted}
		disabled := false
		for _, c := range r.Conditions {
			nc := c
			if s, ok := c.Value.(string); ok {
				if fn, isPH := placeholders[s]; isPH {
					v := fn(id)
					if v == "" {
						disabled = true
						break
					}
					nc.Value = v
				}
			}
			out[i].Conditions = append(out[i].Conditions, nc)
		}
		if disabled {
			out[i].Subject = disabledRule
		}
	}
	return &Evaluator{rules: out}
}

const disabledRule = "\x00disabled"

// Can 由後往前掃規則,先命中者決定;無命中 = deny。
// instance 為 nil(type-level 檢查)時帶條件規則不命中。
func (e *Evaluator) Can(action, subject string, instance map[string]any) bool {
	for i := len(e.rules) - 1; i >= 0; i-- {
		r := e.rules[i]
		if r.Action != actionManage && r.Action != action {
			continue
		}
		if r.Subject != subjectAll && r.Subject != subject {
			continue
		}
		if len(r.Conditions) > 0 {
			if instance == nil || !conditionsMatch(r.Conditions, instance) {
				continue
			}
		}
		return !r.Inverted
	}
	return false
}

func conditionsMatch(conds []FieldCondition, instance map[string]any) bool {
	for _, c := range conds {
		v, ok := instance[c.Field]
		if !ok || !matchValue(c, v) {
			return false
		}
	}
	return true
}

func matchValue(c FieldCondition, v any) bool {
	switch c.Op {
	case OpEq:
		return reflect.DeepEqual(normalize(v), normalize(c.Value))
	case OpNe:
		return !reflect.DeepEqual(normalize(v), normalize(c.Value))
	case OpIn, OpNin:
		arr, _ := c.Value.([]any)
		found := false
		for _, item := range arr {
			if reflect.DeepEqual(normalize(v), normalize(item)) {
				found = true
				break
			}
		}
		if c.Op == OpIn {
			return found
		}
		return !found
	default: // $lt/$lte/$gt/$gte:數值或字串序比較
		cmp, ok := compareOrdered(v, c.Value)
		if !ok {
			return false
		}
		switch c.Op {
		case OpLt:
			return cmp < 0
		case OpLte:
			return cmp <= 0
		case OpGt:
			return cmp > 0
		default: // OpGte
			return cmp >= 0
		}
	}
}

// normalize 把 Go 常見數值型別統一為 float64,讓 JSON 數字與 int 可比。
func normalize(v any) any {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	default:
		return v
	}
}

func compareOrdered(a, b any) (int, bool) {
	an, aok := normalize(a).(float64)
	bn, bok := normalize(b).(float64)
	if aok && bok {
		switch {
		case an < bn:
			return -1, true
		case an > bn:
			return 1, true
		default:
			return 0, true
		}
	}
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return strings.Compare(as, bs), true
	}
	return 0, false
}
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/authz/casl/ -run TestEvaluatorCan -v`
Expected: PASS（10 子測試全綠）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/authz/casl/evaluator.go backend/internal/authz/casl/evaluator_test.go
git commit -m "feat(backend): casl Evaluator(反向掃規則/保留字/佔位符展開)"
```

---

### Task 4: golden fixture 對賭

**Files:**
- Create: `frontend/scripts/casl-golden-gen.mjs`
- Create: `backend/internal/authz/casl/testdata/cases.json`（由產生器輸出）
- Test: `backend/internal/authz/casl/golden_test.go`

**Interfaces:**
- Consumes: Task 2/3 全部。
- Produces: `cases.json` 契約 — `{ "cases": [{ "name", "rules": [{action, subject, conditions?, inverted?}], "identity": {user_id?, company_id?, department_id?, customer_id?}, "checks": [{ "action", "subject", "instance"?, "expect": bool, "query": {"denied": bool, "or": int, "and": int} }] }] }`。`query` 由 JS 端 `rulesToQuery` 輸出結構推得（`$or`/`$and` 陣列長度、null→denied），供 Task 5 比對 Translate 聚合形狀。

- [ ] **Step 1: 寫產生器**

`frontend/scripts/casl-golden-gen.mjs`（於 frontend workspace 執行，`@casl/ability` 由其 package.json 提供——Task 11 的依賴先行加入，或此處暫以 `pnpm add -D @casl/ability` 裝於 frontend）:

```js
// 產生 Go casl 套件的 golden fixture。用法:pnpm --filter frontend exec node scripts/casl-golden-gen.mjs
import { createMongoAbility, subject as wrapSubject } from "@casl/ability";
import { rulesToQuery } from "@casl/ability/extra";
import { writeFileSync } from "node:fs";

const U = (id) => Object.fromEntries(
  Object.entries({ user_id: id.user, company_id: id.company, department_id: id.department, customer_id: id.customer })
    .filter(([, v]) => v !== undefined),
);

// 與 Go 端相同的佔位符展開(產生 fixture 時展開,讓 JS 端見到具體值)
function expand(rules, id) {
  const map = {
    "${user.id}": id.user, "${user.company_id}": id.company,
    "${user.department_id}": id.department, "${user.customer_id}": id.customer,
  };
  return rules.map((r) => {
    if (!r.conditions) return { action: r.action, subject: r.subject, ...(r.inverted ? { inverted: true } : {}) };
    const conds = {};
    for (const [f, v] of Object.entries(r.conditions)) {
      conds[f] = typeof v === "string" && map[v] !== undefined ? map[v] : v;
    }
    return { action: r.action, subject: r.subject, conditions: conds, ...(r.inverted ? { inverted: true } : {}) };
  });
}

const staff = { user: "u1", company: "c1", department: "d1" };

const cases = [
  {
    name: "無條件 + cannot 狀態排除",
    identity: U(staff),
    rules: [
      { action: "read", subject: "sales_order" },
      { action: "read", subject: "sales_order", inverted: true, conditions: { status: "voided" } },
    ],
    checks: [
      { action: "read", subject: "sales_order", instance: { status: "pending" }, expect: true },
      { action: "read", subject: "sales_order", instance: { status: "voided" }, expect: false },
    ],
  },
  {
    name: "佔位符 + $in",
    identity: U(staff),
    rules: [
      { action: "read", subject: "sales_order", conditions: { department_id: "${user.department_id}" } },
      { action: "cancel", subject: "sales_order", conditions: { status: { $in: ["pending"] } } },
    ],
    checks: [
      { action: "read", subject: "sales_order", instance: { department_id: "d1" }, expect: true },
      { action: "read", subject: "sales_order", instance: { department_id: "d2" }, expect: false },
      { action: "cancel", subject: "sales_order", instance: { status: "pending" }, expect: true },
      { action: "cancel", subject: "sales_order", instance: { status: "processing" }, expect: false },
      { action: "cancel", subject: "sales_order", expect: false }, // type-level:帶條件規則不命中
    ],
  },
  {
    name: "manage all 與無規則 deny",
    identity: U(staff),
    rules: [{ action: "manage", subject: "all" }],
    checks: [
      { action: "delete", subject: "customer", instance: { x: 1 }, expect: true },
    ],
  },
  {
    name: "多允許規則 OR 聚合(query 形狀)",
    identity: U(staff),
    rules: [
      { action: "read", subject: "sales_order", conditions: { department_id: "${user.department_id}" } },
      { action: "read", subject: "sales_order", conditions: { created_by: "${user.id}" } },
      { action: "read", subject: "sales_order", inverted: true, conditions: { status: "voided" } },
    ],
    checks: [
      { action: "read", subject: "sales_order", instance: { department_id: "d1" }, expect: true },
      { action: "read", subject: "sales_order", instance: { department_id: "d2", created_by: "u1" }, expect: true },
      { action: "read", subject: "sales_order", instance: { department_id: "d1", status: "voided" }, expect: false },
      { action: "read", subject: "customer", instance: {}, expect: false }, // 無規則 subject → query denied
    ],
  },
];

for (const c of cases) {
  const expanded = expand(c.rules, JSON.parse(JSON.stringify(c.identity)));
  const ability = createMongoAbility(expanded);
  for (const chk of c.checks) {
    const inst = chk.instance ? wrapSubject(chk.subject, { ...chk.instance }) : undefined;
    const got = ability.can(chk.action, inst ?? chk.subject);
    if (got !== chk.expect) throw new Error(`${c.name}: can(${chk.action},${chk.subject}) = ${got}, expect ${chk.expect}`);
    const q = rulesToQuery(ability, chk.action, chk.subject, (rule) =>
      rule.inverted ? { not: rule.conditions ?? {} } : (rule.conditions ?? {}),
    );
    chk.query = q === null
      ? { denied: true, or: 0, and: 0 }
      : { denied: false, or: (q.$or ?? []).length, and: (q.$and ?? []).length };
  }
}

const out = new URL("../../backend/internal/authz/casl/testdata/cases.json", import.meta.url);
writeFileSync(out, JSON.stringify({ cases }, null, 2) + "\n");
console.log(`wrote ${out.pathname}`);
```

- [ ] **Step 2: 執行產生器並檢查輸出**

Run: `pnpm --filter frontend exec node scripts/casl-golden-gen.mjs`
Expected: 印出 `wrote …/cases.json`，無例外（產生器內建 JS 端斷言，expect 與真 CASL 不一致會直接 throw）。

- [ ] **Step 3: 寫 Go 重放測試**

`backend/internal/authz/casl/golden_test.go`:

```go
package casl

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type goldenFile struct {
	Cases []struct {
		Name     string            `json:"name"`
		Rules    []json.RawMessage `json:"rules"`
		Identity struct {
			UserID       string `json:"user_id"`
			CompanyID    string `json:"company_id"`
			DepartmentID string `json:"department_id"`
			CustomerID   string `json:"customer_id"`
		} `json:"identity"`
		Checks []struct {
			Action   string         `json:"action"`
			Subject  string         `json:"subject"`
			Instance map[string]any `json:"instance"`
			Expect   bool           `json:"expect"`
			Query    struct {
				Denied bool `json:"denied"`
				Or     int  `json:"or"`
				And    int  `json:"and"`
			} `json:"query"`
		} `json:"checks"`
	} `json:"cases"`
}

func loadGolden(t *testing.T) goldenFile {
	t.Helper()
	raw, err := os.ReadFile("testdata/cases.json")
	require.NoError(t, err)
	var gf goldenFile
	require.NoError(t, json.Unmarshal(raw, &gf))
	require.NotEmpty(t, gf.Cases)
	return gf
}

func TestGoldenCan(t *testing.T) {
	for _, c := range loadGolden(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var rules []Rule
			for _, rr := range c.Rules {
				var jr struct {
					Action     string         `json:"action"`
					Subject    string         `json:"subject"`
					Conditions map[string]any `json:"conditions"`
					Inverted   bool           `json:"inverted"`
				}
				require.NoError(t, json.Unmarshal(rr, &jr))
				conds, err := ParseConditions(jr.Conditions)
				require.NoError(t, err)
				rules = append(rules, Rule{Action: jr.Action, Subject: jr.Subject, Conditions: conds, Inverted: jr.Inverted})
			}
			id := Identity{c.Identity.UserID, c.Identity.CompanyID, c.Identity.DepartmentID, c.Identity.CustomerID}
			e := NewEvaluator(rules, id)
			for _, chk := range c.Checks {
				assert.Equal(t, chk.Expect, e.Can(chk.Action, chk.Subject, chk.Instance),
					"can(%s, %s, %v)", chk.Action, chk.Subject, chk.Instance)
			}
		})
	}
}
```

註：golden 的 `query` 聚合形狀比對在 Task 5 的 `TestGoldenTranslate` 實作（需 Translate 與 registry 存在）；本 Task 的 golden_test.go 只驗 `Can`。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/authz/casl/ -run TestGoldenCan -v`
Expected: PASS（4 cases 全綠）

- [ ] **Step 5: Commit**

```bash
git add frontend/scripts/casl-golden-gen.mjs backend/internal/authz/casl/testdata/cases.json backend/internal/authz/casl/golden_test.go
git commit -m "test(backend): casl golden fixture 對賭(真 @casl/ability 產生)"
```

---

### Task 5: Translate（rules→SQL WHERE）

**Files:**
- Create: `backend/internal/authz/casl/translate.go`
- Test: `backend/internal/authz/casl/translate_test.go`（含 `TestGoldenTranslate`）

**Interfaces:**
- Consumes: Task 2/3；Task 6 的 `FieldRegistry`（本 Task 先以測試用 registry 定義介面需求，Task 6 補正式實作——為免循環，**本 Task 一併建立 `registry.go` 的型別定義**，Task 6 填 sales_order 註冊與白名單 API）。
- Produces: `casl.Translate(rules []Rule, action, subject string, reg *FieldRegistry) (clause string, args []any, denied bool, err error)`；聚合語意：直接規則條件 OR 併、inverted 規則 `NOT(...)` AND 併、無直接規則 → `denied=true`。

- [ ] **Step 1: 建立 FieldRegistry 型別骨架**

`backend/internal/authz/casl/registry.go`:

```go
package casl

import "fmt"

// FieldType 為條件欄位值型別,供驗證與 UI 建構器。
type FieldType string

const (
	TypeString FieldType = "string"
	TypeEnum   FieldType = "enum"
	TypeID     FieldType = "id"
	TypeNumber FieldType = "number"
)

// FieldDef 為單一邏輯欄位的定義。
type FieldDef struct {
	Column string    // DB column 名
	Ops    []Op      // 允許運算子
	Type   FieldType // 值型別
	Enum   []string  // TypeEnum 的合法值
}

// SubjectDef 為單一 subject 的欄位集與實例擷取器。
type SubjectDef struct {
	Fields map[string]FieldDef
	// ToInstance 由 Ent 實體擷取條件所需欄位值(供 Evaluator.Can)。
	ToInstance func(entity any) map[string]any
}

// FieldRegistry 為 subject → 定義的註冊表;SQL 翻譯、實例擷取、UI 白名單三方共用。
type FieldRegistry struct {
	subjects map[string]SubjectDef
}

func NewFieldRegistry() *FieldRegistry {
	return &FieldRegistry{subjects: map[string]SubjectDef{}}
}

func (r *FieldRegistry) Register(subject string, def SubjectDef) {
	r.subjects[subject] = def
}

// Field 查詢 subject 的欄位定義;未知 subject/欄位回 error(fail-closed 由呼叫面處理)。
func (r *FieldRegistry) Field(subject, field string) (FieldDef, error) {
	sd, ok := r.subjects[subject]
	if !ok {
		return FieldDef{}, fmt.Errorf("casl: unknown subject %q", subject)
	}
	fd, ok := sd.Fields[field]
	if !ok {
		return FieldDef{}, fmt.Errorf("casl: unknown field %q on subject %q", field, subject)
	}
	return fd, nil
}

// Instance 由實體擷取 instance map;subject 未註冊回 nil。
func (r *FieldRegistry) Instance(subject string, entity any) map[string]any {
	sd, ok := r.subjects[subject]
	if !ok || sd.ToInstance == nil {
		return nil
	}
	return sd.ToInstance(entity)
}
```

- [ ] **Step 2: 寫失敗測試**

`backend/internal/authz/casl/translate_test.go`:

```go
package casl

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func orderRegistry() *FieldRegistry {
	r := NewFieldRegistry()
	r.Register("sales_order", SubjectDef{Fields: map[string]FieldDef{
		"status":        {Column: "status", Ops: []Op{OpEq, OpNe, OpIn, OpNin}, Type: TypeEnum, Enum: []string{"pending", "processing", "completed", "cancelled", "voided"}},
		"department_id": {Column: "department_id", Ops: []Op{OpEq, OpNe, OpIn, OpNin}, Type: TypeID},
		"created_by":    {Column: "created_by", Ops: []Op{OpEq, OpNe}, Type: TypeID},
	}})
	return r
}

func TestTranslate(t *testing.T) {
	reg := orderRegistry()
	read := func(raw map[string]any, inverted bool) Rule {
		c, err := ParseConditions(raw)
		require.NoError(t, err)
		return Rule{Action: "read", Subject: "sales_order", Conditions: c, Inverted: inverted}
	}
	t.Run("單一直接規則", func(t *testing.T) {
		clause, args, denied, err := Translate([]Rule{read(map[string]any{"status": "pending"}, false)}, "read", "sales_order", reg)
		require.NoError(t, err)
		assert.False(t, denied)
		assert.Equal(t, `("status" = ?)`, clause)
		assert.Equal(t, []any{"pending"}, args)
	})
	t.Run("多直接規則 OR 併 + inverted AND NOT 併", func(t *testing.T) {
		clause, args, denied, err := Translate([]Rule{
			read(map[string]any{"status": map[string]any{"$in": []any{"pending", "processing"}}}, false),
			read(map[string]any{"created_by": "u1"}, false),
			read(map[string]any{"status": "voided"}, true),
		}, "read", "sales_order", reg)
		require.NoError(t, err)
		assert.False(t, denied)
		assert.Equal(t, `(("status" IN (?, ?)) OR ("created_by" = ?)) AND (NOT ("status" = ?))`, clause)
		assert.Equal(t, []any{"pending", "processing", "u1", "voided"}, args)
	})
	t.Run("action/subject 不符的規則略過", func(t *testing.T) {
		_, _, denied, err := Translate([]Rule{
			{Action: "update", Subject: "sales_order"},
		}, "read", "sales_order", reg)
		require.NoError(t, err)
		assert.True(t, denied)
	})
	t.Run("無條件直接規則 = 全放行(true)", func(t *testing.T) {
		clause, args, denied, err := Translate([]Rule{{Action: "read", Subject: "sales_order"}}, "read", "sales_order", reg)
		require.NoError(t, err)
		assert.False(t, denied)
		assert.Equal(t, "(TRUE)", clause)
		assert.Empty(t, args)
	})
	t.Run("未知欄位規則 fail-closed 略過", func(t *testing.T) {
		_, _, denied, err := Translate([]Rule{
			read(map[string]any{"nonexistent": "x"}, false),
		}, "read", "sales_order", reg)
		require.NoError(t, err)
		assert.True(t, denied) // 唯一規則被略過 → 無允許規則
	})
	t.Run("manage/all 保留字參與聚合", func(t *testing.T) {
		clause, _, denied, err := Translate([]Rule{{Action: "manage", Subject: "all"}}, "read", "sales_order", reg)
		require.NoError(t, err)
		assert.False(t, denied)
		assert.Equal(t, "(TRUE)", clause)
	})
}

// TestGoldenTranslate 比對 cases.json 的 query 聚合形狀(denied / $or 數 / $and 數)。
func TestGoldenTranslate(t *testing.T) {
	reg := orderRegistry()
	for _, c := range loadGolden(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var rules []Rule
			for _, rr := range c.Rules {
				var jr struct {
					Action     string         `json:"action"`
					Subject    string         `json:"subject"`
					Conditions map[string]any `json:"conditions"`
					Inverted   bool           `json:"inverted"`
				}
				require.NoError(t, json.Unmarshal(rr, &jr))
				conds, err := ParseConditions(jr.Conditions)
				require.NoError(t, err)
				rules = append(rules, Rule{Action: jr.Action, Subject: jr.Subject, Conditions: conds, Inverted: jr.Inverted})
			}
			id := Identity{c.Identity.UserID, c.Identity.CompanyID, c.Identity.DepartmentID, c.Identity.CustomerID}
			rules = NewEvaluator(rules, id).rules // 佔位符展開後再翻譯
			for _, chk := range c.Checks {
				_, _, denied, err := Translate(rules, chk.Action, chk.Subject, reg)
				require.NoError(t, err)
				assert.Equal(t, chk.Query.Denied, denied, "denied for %s/%s", chk.Action, chk.Subject)
			}
		})
	}
}
```

註：golden 的 `created_by` 欄位需存在於 `orderRegistry()`（已含）；`customer` subject 未註冊 → Translate 對其回 `denied`（對齊 JS `rulesToQuery` 對無規則 subject 回 null）。

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/authz/casl/ -run 'TestTranslate|TestGoldenTranslate' -v`
Expected: FAIL — `undefined: Translate`

- [ ] **Step 4: 實作**

`backend/internal/authz/casl/translate.go`:

```go
package casl

import (
	"fmt"
	"strings"
)

// Translate 將規則翻譯為參數化 SQL WHERE 片段(佔位符 ?,Ent sql.ExprP 依 dialect 重寫)。
// 對應 @casl/ability 的 rulesToQuery:直接規則條件 OR 併、inverted 規則 NOT(...) AND 併;
// 無任何直接規則命中 → denied=true(對應 rulesToQuery 回 null,呼叫面短路回空集)。
// 未知 subject/欄位的規則略過(fail-closed,等同不命中),由呼叫面記 security log。
func Translate(rules []Rule, action, subject string, reg *FieldRegistry) (clause string, args []any, denied bool, err error) {
	var ors, ands []string
	for _, r := range rules {
		if r.Action != actionManage && r.Action != action {
			continue
		}
		if r.Subject != subjectAll && r.Subject != subject {
			continue
		}
		frag, a, ferr := conditionsSQL(subject, r.Conditions, reg)
		if ferr != nil {
			continue // fail-closed:視為不命中
		}
		if r.Inverted {
			ands = append(ands, "NOT ("+frag+")")
		} else {
			ors = append(ors, frag)
			args = append(args, a...)
		}
	}
	if len(ors) == 0 {
		return "", nil, true, nil
	}
	clause = "(" + strings.Join(ors, " OR ") + ")"
	if len(ands) > 0 {
		clause += " AND (" + strings.Join(ands, " AND ") + ")"
	}
	return clause, args, false, nil
}

// conditionsSQL 將單條規則的欄位條件 AND 併為 SQL 片段。
func conditionsSQL(subject string, conds []FieldCondition, reg *FieldRegistry) (string, []any, error) {
	if len(conds) == 0 {
		return "TRUE", nil, nil
	}
	var parts []string
	var args []any
	for _, c := range conds {
		fd, err := reg.Field(subject, c.Field)
		if err != nil {
			return "", nil, err
		}
		if !opAllowed(fd, c.Op) {
			return "", nil, fmt.Errorf("casl: op %s not allowed on %s.%s", c.Op, subject, c.Field)
		}
		col := `"` + fd.Column + `"`
		switch c.Op {
		case OpEq:
			parts = append(parts, col+" = ?")
			args = append(args, c.Value)
		case OpNe:
			parts = append(parts, col+" <> ?")
			args = append(args, c.Value)
		case OpIn, OpNin:
			arr, _ := c.Value.([]any)
			if len(arr) == 0 {
				return "", nil, fmt.Errorf("casl: %s on %s.%s requires non-empty array", c.Op, subject, c.Field)
			}
			ph := strings.TrimSuffix(strings.Repeat("?, ", len(arr)), ", ")
			kw := "IN"
			if c.Op == OpNin {
				kw = "NOT IN"
			}
			parts = append(parts, col+" "+kw+" ("+ph+")")
			args = append(args, arr...)
		default: // $lt/$lte/$gt/$gte
			sym := map[Op]string{OpLt: "<", OpLte: "<=", OpGt: ">", OpGte: ">="}[c.Op]
			parts = append(parts, col+" "+sym+" ?")
			args = append(args, c.Value)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}

func opAllowed(fd FieldDef, op Op) bool {
	for _, o := range fd.Ops {
		if o == op {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go test ./internal/authz/casl/ -v`
Expected: PASS（含 TestGoldenTranslate；注意 `Evaluator.rules` 為未匯出欄位，golden 測試與套件同 package 可直接取用）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/authz/casl/{registry,translate}.go backend/internal/authz/casl/translate_test.go
git commit -m "feat(backend): casl Translate(rules→SQL,rulesToQuery 對應物) + FieldRegistry 骨架"
```

---

### Task 6: FieldRegistry 正式註冊與白名單 API

**Files:**
- Modify: `backend/internal/authz/casl/registry.go`（新增 `RegisterSalesOrder`、`ConditionFields`、`ValidateRuleConditions`）
- Test: `backend/internal/authz/casl/registry_test.go`

**Interfaces:**
- Consumes: Task 5 `FieldRegistry`。
- Produces:
  - `casl.RegisterSalesOrder(reg *FieldRegistry)` — 註冊 `sales_order` 全部條件欄位（`status`/`department_id`/`company_id`/`customer_id`/`created_by`/`expected_delivery_date`）與 `ToInstance`（消費 `*ent.SalesOrder`；Ent 產生碼未就緒前以介面 `interface{ … }` 延後——**改以 map 輸入**：`ToInstance func(entity any) map[string]any` 的實作於 05-sales-orders 計畫落地，本 Task 先以 `nil` 註冊並於該計畫補實作，doc-sync 已於 Task 1 處理計畫間引用）。
  - `(r *FieldRegistry) ConditionFields(subject string) []ConditionFieldInfo`（`{Field, Type, Ops, Enum}`，供 Task 10 RPC）。
  - `(r *FieldRegistry) ValidateRuleConditions(subject string, conds []FieldCondition) error`（未知欄位/運算子/enum 值 → error，供 Task 10 寫入驗證）。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/authz/casl/registry_test.go`:

```go
package casl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConditionFields(t *testing.T) {
	reg := NewFieldRegistry()
	RegisterSalesOrder(reg)
	fields := reg.ConditionFields("sales_order")
	require.NotEmpty(t, fields)
	byName := map[string]ConditionFieldInfo{}
	for _, f := range fields {
		byName[f.Field] = f
	}
	assert.Equal(t, TypeEnum, byName["status"].Type)
	assert.Contains(t, byName["status"].Enum, "pending")
	assert.Equal(t, TypeID, byName["department_id"].Type)
	assert.Empty(t, reg.ConditionFields("nonexistent"))
}

func TestValidateRuleConditions(t *testing.T) {
	reg := NewFieldRegistry()
	RegisterSalesOrder(reg)
	t.Run("合法", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"status": map[string]any{"$in": []any{"pending"}}})
		assert.NoError(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("未知欄位", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"nope": 1.0})
		assert.Error(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("運算子不在該欄位白名單", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"status": map[string]any{"$lt": "x"}})
		assert.Error(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("enum 值非法", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"status": "bogus"})
		assert.Error(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("無條件恆合法", func(t *testing.T) {
		assert.NoError(t, reg.ValidateRuleConditions("sales_order", nil))
	})
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/authz/casl/ -run 'TestConditionFields|TestValidateRuleConditions' -v`
Expected: FAIL — `undefined: RegisterSalesOrder`

- [ ] **Step 3: 實作（registry.go 追加）**

```go
// ConditionFieldInfo 為 UI 條件建構器的欄位描述。
type ConditionFieldInfo struct {
	Field string
	Type  FieldType
	Ops   []Op
	Enum  []string
}

// ConditionFields 回傳 subject 的條件欄位白名單(依欄位名排序);未知 subject 回空。
func (r *FieldRegistry) ConditionFields(subject string) []ConditionFieldInfo {
	sd, ok := r.subjects[subject]
	if !ok {
		return nil
	}
	names := make([]string, 0, len(sd.Fields))
	for n := range sd.Fields {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]ConditionFieldInfo, 0, len(names))
	for _, n := range names {
		fd := sd.Fields[n]
		out = append(out, ConditionFieldInfo{Field: n, Type: fd.Type, Ops: fd.Ops, Enum: fd.Enum})
	}
	return out
}

// ValidateRuleConditions 驗證寫入的規則條件:欄位/運算子白名單 + enum 值合法。
func (r *FieldRegistry) ValidateRuleConditions(subject string, conds []FieldCondition) error {
	for _, c := range conds {
		fd, err := r.Field(subject, c.Field)
		if err != nil {
			return err
		}
		if !opAllowed(fd, c.Op) {
			return fmt.Errorf("casl: op %s not allowed on %s.%s", c.Op, subject, c.Field)
		}
		if fd.Type == TypeEnum {
			vals, _ := c.Value.([]any)
			if c.Op != OpIn && c.Op != OpNin {
				vals = []any{c.Value}
			}
			for _, v := range vals {
				s, ok := v.(string)
				if !ok || !slices.Contains(fd.Enum, s) {
					return fmt.Errorf("casl: invalid enum value %v on %s.%s", v, subject, c.Field)
				}
			}
		}
	}
	return nil
}

// RegisterSalesOrder 註冊 sales_order subject 的條件欄位。
// ToInstance 由 05-sales-orders 計畫於 Ent 產生碼就緒後補實作(消費 *ent.SalesOrder)。
func RegisterSalesOrder(reg *FieldRegistry) {
	idOps := []Op{OpEq, OpNe, OpIn, OpNin}
	reg.Register("sales_order", SubjectDef{
		Fields: map[string]FieldDef{
			"status":                 {Column: "status", Ops: idOps, Type: TypeEnum, Enum: []string{"pending", "processing", "completed", "cancelled", "voided"}},
			"department_id":          {Column: "department_id", Ops: idOps, Type: TypeID},
			"company_id":             {Column: "company_id", Ops: idOps, Type: TypeID},
			"customer_id":            {Column: "customer_id", Ops: idOps, Type: TypeID},
			"created_by":             {Column: "created_by", Ops: []Op{OpEq, OpNe}, Type: TypeID},
			"expected_delivery_date": {Column: "expected_delivery_date", Ops: []Op{OpEq, OpNe, OpLt, OpLte, OpGt, OpGte}, Type: TypeString},
		},
	})
}
```

（檔頭 import 加 `"sort"` 與 `"slices"`。）

- [ ] **Step 4: 跑測試確認通過 + 全套件回歸**

Run: `cd backend && go test ./internal/authz/casl/ -v`
Expected: PASS（全部）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/authz/casl/registry.go backend/internal/authz/casl/registry_test.go
git commit -m "feat(backend): casl FieldRegistry 註冊與白名單驗證(sales_order)"
```

---

### Task 7: role_permissions schema 擴充

**Files:**
- Modify: `backend/ent/schema/rolepermission.go`（若 02 計畫尚未實作則 Create，以其 schema 為基底加三欄）
- Create: `backend/database/migrations/000XX_role_permissions_casl.sql`
- Test: `backend/internal/domain/roles/schema_casl_test.go`

**Interfaces:**
- Consumes: 02 計畫 2.9.1 的 roles/role_permissions 基底（Task 1 Step 4 已修訂其定義，以此為準）。
- Produces: `role_permissions` 加 `conditions jsonb NULL`、`inverted bool NOT NULL DEFAULT false`、`sort_order int NOT NULL DEFAULT 0`；唯一索引 `(role_id, resource, action, COALESCE(md5(conditions::text), ''))`；Ent 實體 `ent.RolePermission` 含三新欄位。

- [ ] **Step 1: 寫失敗測試（DB 整合）**

`backend/internal/domain/roles/schema_casl_test.go`:

```go
package roles_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestRolePermissionCASLColumns(t *testing.T) {
	client := testutil.NewEntClient(t)
	ctx := context.Background()

	role := client.Role.Create().SetCode("staff").SetName("門市人員").
		SetDataScope("department").SetIsSystem(true).SaveX(ctx)

	t.Run("conditions/inverted/sort_order 寫入讀回", func(t *testing.T) {
		rp := client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("sales_order").SetAction("cancel").
			SetConditions(map[string]any{"status": "pending"}).
			SetInverted(false).SetSortOrder(10).
			SaveX(ctx)
		got := client.RolePermission.GetX(ctx, rp.ID)
		assert.Equal(t, map[string]any{"status": "pending"}, got.Conditions)
		assert.False(t, got.Inverted)
		assert.Equal(t, 10, got.SortOrder)
	})

	t.Run("同 resource×action 不同 conditions 可並存", func(t *testing.T) {
		client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
			SetConditions(map[string]any{"department_id": "${user.department_id}"}).SetSortOrder(1).
			SaveX(ctx)
		client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
			SetConditions(map[string]any{"created_by": "${user.id}"}).SetSortOrder(2).
			SaveX(ctx)
		n := client.RolePermission.Query().
			Where(/* rolepermission.RoleID(role.ID), Resource("sales_order"), Action("read") */).
			CountX(ctx)
		assert.Equal(t, 2, n)
	})

	t.Run("同 resource×action 相同 conditions 被唯一索引拒絕", func(t *testing.T) {
		client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("customer").SetAction("read").
			SetConditions(map[string]any{"company_id": "c1"}).
			SaveX(ctx)
		_, err := client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("customer").SetAction("read").
			SetConditions(map[string]any{"company_id": "c1"}).
			Save(ctx)
		require.Error(t, err)
	})
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/roles/ -run TestRolePermissionCASLColumns -v`
Expected: FAIL — schema 無三欄（編譯錯或欄位不存在）

- [ ] **Step 3: Ent schema 三欄 + migration**

`backend/ent/schema/rolepermission.go` 的 Fields 追加：

```go
field.JSON("conditions", map[string]any{}).Optional(),
field.Bool("inverted").Default(false),
field.Int("sort_order").Default(0),
```

migration `backend/database/migrations/000XX_role_permissions_casl.sql`（goose 格式，XX 接續既有編號）:

```sql
-- +goose Up
ALTER TABLE role_permissions
  ADD COLUMN conditions jsonb NULL,
  ADD COLUMN inverted boolean NOT NULL DEFAULT false,
  ADD COLUMN sort_order integer NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS role_permissions_role_id_resource_action_key;
CREATE UNIQUE INDEX role_permissions_rule_key
  ON role_permissions (role_id, resource, action, COALESCE(md5(conditions::text), ''));

-- +goose Down
DROP INDEX IF EXISTS role_permissions_rule_key;
ALTER TABLE role_permissions
  DROP COLUMN conditions,
  DROP COLUMN inverted,
  DROP COLUMN sort_order;
```

（若 02 計畫未實作、role_permissions 尚未建表：改為於建表 migration 直接含三欄與唯一索引，本 alter migration 省略。）

- [ ] **Step 4: Ent code gen 後跑測試確認通過**

Run: `cd backend && task backend:ent:gen && go test ./internal/domain/roles/ -run TestRolePermissionCASLColumns -v`
Expected: PASS（3 子測試全綠）

- [ ] **Step 5: Commit**

```bash
git add backend/ent/schema/rolepermission.go backend/database/migrations/ backend/ent/ backend/internal/domain/roles/schema_casl_test.go
git commit -m "feat(backend): role_permissions 擴充 CASL 三欄與條件唯一鍵"
```

---

### Task 8: authz facade 與 env 開關

**Files:**
- Create: `backend/internal/authz/access.go`
- Modify: `backend/config/api.go`（`CASLEnforcementEnabled bool`，env `CASL_ENFORCEMENT_ENABLED`，三環境預設 true）
- Modify: `backend/internal/middleware/auth.go`（01 計畫產物；身分注入後呼叫 `authz.WithCASLEnabled(ctx, cfg.CASLEnforcementEnabled)`）
- Test: `backend/internal/authz/access_test.go`

**Interfaces:**
- Consumes: Task 2–6 全部；01 計畫 middleware 的 identity ctx（`middleware.IdentityFrom(ctx)` 回 `{UserID, CompanyID, DepartmentID, CustomerID string; Role string; Roles []string}`——01 計畫若無 `Roles` 複數欄位，於 middleware 依 Casbin g 規則補齊，Task 9/10 依賴此形）。
- Produces:
  - `authz.WithCASLEnabled(ctx, enabled bool) context.Context`；middleware 每請求呼叫一次。
  - `authz.AccessibleFilter(ctx, action, subject string) (clause string, args []any, denied bool)` — 關閉時 `("", nil, false)`；開啟時自 ctx 取身分→查角色規則→Evaluator 展開→Translate。
  - `authz.Can(ctx, action, subject string, entity any) bool` — 關閉時 true；開啟時 registry.Instance + Evaluator.Can。
  - 規則載入：`authz.loadRules(ctx, db, roles []string) ([]casl.Rule, error)` — 由 `role_permissions` JOIN `roles` 依 `sort_order` 升冪取出；同請求快取於 ctx。

- [ ] **Step 1: 寫失敗測試（DB 整合）**

`backend/internal/authz/access_test.go`:

```go
package authz_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestAccessibleFilter(t *testing.T) {
	client := testutil.NewEntClient(t)
	ctx := context.Background()
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetConditions(map[string]any{"department_id": "${user.department_id}"}).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetInverted(true).SetConditions(map[string]any{"status": "voided"}).SetSortOrder(1).SaveX(ctx)

	id := middleware.Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1", Roles: []string{"staff"}}

	t.Run("開啟時回條件片段", func(t *testing.T) {
		c := authz.WithCASLEnabled(middleware.WithIdentity(testutil.CtxWithDB(ctx, client), id), true)
		clause, args, denied := authz.AccessibleFilter(c, "read", "sales_order")
		assert.False(t, denied)
		assert.Equal(t, `("department_id" = ?) AND (NOT ("status" = ?))`, clause)
		assert.Equal(t, []any{"d1", "voided"}, args)
	})
	t.Run("關閉時回空條件", func(t *testing.T) {
		c := authz.WithCASLEnabled(middleware.WithIdentity(testutil.CtxWithDB(ctx, client), id), false)
		clause, args, denied := authz.AccessibleFilter(c, "read", "sales_order")
		assert.False(t, denied)
		assert.Empty(t, clause)
		assert.Nil(t, args)
	})
	t.Run("無規則 resource denied", func(t *testing.T) {
		c := authz.WithCASLEnabled(middleware.WithIdentity(testutil.CtxWithDB(ctx, client), id), true)
		_, _, denied := authz.AccessibleFilter(c, "read", "customer")
		assert.True(t, denied)
	})
}
```

（`testutil.CtxWithDB` / `middleware.WithIdentity` 若 01 計畫命名不同，以 01 計畫產物為準調整 import 與呼叫；斷言不變。）

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/authz/ -run TestAccessibleFilter -v`
Expected: FAIL — `undefined: authz.WithCASLEnabled`

- [ ] **Step 3: 實作**

`backend/internal/authz/access.go`:

```go
package authz

import (
	"context"
	"sync"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/casl"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
)

type ctxKey int

const (
	keyEnabled ctxKey = iota
	keyRegistry
)

// WithCASLEnabled 由 middleware 每請求呼叫一次,記錄開關狀態。
func WithCASLEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, keyEnabled, enabled)
}

func caslEnabled(ctx context.Context) bool {
	v, _ := ctx.Value(keyEnabled).(bool)
	return v
}

// Registry 回傳進程級 FieldRegistry(啟動時註冊全部 subject)。
func Registry(ctx context.Context) *casl.FieldRegistry {
	if r, ok := ctx.Value(keyRegistry).(*casl.FieldRegistry); ok {
		return r
	}
	return defaultRegistry()
}

var (
	registryOnce sync.Once
	registryInst *casl.FieldRegistry
)

func defaultRegistry() *casl.FieldRegistry {
	registryOnce.Do(func() {
		registryInst = casl.NewFieldRegistry()
		casl.RegisterSalesOrder(registryInst)
	})
	return registryInst
}

// loadRules 取當前身分全部角色的規則,SQL ORDER BY sort_order 升冪。
// edge 與 predicate 名依 02 計畫 role_permissions schema 的 Ent 產生碼為準。
func loadRules(ctx context.Context, db *ent.Client, roles []string) ([]casl.Rule, error) {
	rows, err := db.RolePermission.Query().
		Where(rolepermission.HasRoleWith(role.CodeIn(roles...))).
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]casl.Rule, 0, len(rows))
	for _, rp := range rows {
		conds, err := casl.ParseConditions(rp.Conditions)
		if err != nil {
			// fail-closed:壞規則視為不存在,記 security log(接入 03 計畫的 logger 後補)
			continue
		}
		rules = append(rules, casl.Rule{
			Action: rp.Action, Subject: rp.Resource, Conditions: conds, Inverted: rp.Inverted,
		})
	}
	return rules, nil
}

func evaluator(ctx context.Context, db *ent.Client) (*casl.Evaluator, error) {
	id := middleware.IdentityFrom(ctx)
	rules, err := loadRules(ctx, db, id.Roles)
	if err != nil {
		return nil, err
	}
	return casl.NewEvaluator(rules, casl.Identity{
		UserID: id.UserID, CompanyID: id.CompanyID,
		DepartmentID: id.DepartmentID, CustomerID: id.CustomerID,
	}), nil
}

// AccessibleFilter 供 repository list 查詢套用。
// 開關關閉 → 空條件;無允許規則 → denied=true(短路回空集)。
func AccessibleFilter(ctx context.Context, db *ent.Client, action, subject string) (string, []any, bool) {
	if !caslEnabled(ctx) {
		return "", nil, false
	}
	e, err := evaluator(ctx, db)
	if err != nil {
		return "", nil, true // 規則載入失敗 fail-closed
	}
	clause, args, denied, err := casl.Translate(e.Rules(), action, subject, Registry(ctx))
	if err != nil {
		return "", nil, true
	}
	return clause, args, denied
}

// Can 供 usecase 對單筆實體做條件檢查。開關關閉 → true。
func Can(ctx context.Context, db *ent.Client, action, subject string, entity any) bool {
	if !caslEnabled(ctx) {
		return true
	}
	e, err := evaluator(ctx, db)
	if err != nil {
		return false
	}
	return e.Can(action, subject, Registry(ctx).Instance(subject, entity))
}
```

配套修改：
1. `evaluator.go` 加匯出存取器：`func (e *Evaluator) Rules() []Rule { return e.rules }`（Translate 需要展開後的規則）。
2. `config.go` 加欄位與預設：`CASLEnforcementEnabled bool`，env 解析 `CASL_ENFORCEMENT_ENABLED`，未設定時 `true`（三環境同）；啟動 log 印出狀態。
3. `middleware/auth.go` 在 identity 注入後加一行：`ctx = authz.WithCASLEnabled(ctx, cfg.CASLEnforcementEnabled)`。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/authz/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/authz/ backend/config/api.go backend/internal/middleware/auth.go
git commit -m "feat(backend): authz facade(AccessibleFilter/Can) + CASL_ENFORCEMENT_ENABLED 開關"
```

---

### Task 9: GetAbility 輸出 conditions/inverted

**Files:**
- Modify: `backend/proto/v1/ability.proto`
- Modify: `backend/internal/domain/auth/ability.go`（01 計畫 Task 1.8 產物）
- Test: `backend/internal/domain/auth/ability_test.go`

**Interfaces:**
- Consumes: Task 7 表、Task 8 規則載入邏輯（重用 `authz` 內部載入或於 ability.go 內聯同構查詢）。
- Produces: proto `AbilityRule{string action=1; string subject=2; optional google.protobuf.Struct conditions=3; bool inverted=4;}`；`GetAbilityResponse{repeated AbilityRule rules=1;}`。規則依 `sort_order` 排序輸出；佔位符**已以身分展開為具體值**（前端不處理佔位符）。`developer`（開關啟用）回 `[{action:"manage", subject:"all"}]`；`guest` 回空。

- [ ] **Step 1: proto 更新並產生碼**

`backend/proto/v1/ability.proto`:

```proto
syntax = "proto3";
package v1;

import "google/protobuf/struct.proto";

message AbilityRule {
  string action = 1;
  string subject = 2;
  google.protobuf.Struct conditions = 3; // 無條件時為 null
  bool inverted = 4;
}

message GetAbilityRequest {}

message GetAbilityResponse {
  repeated AbilityRule rules = 1;
}

service AbilityService {
  rpc GetAbility(GetAbilityRequest) returns (GetAbilityResponse);
}
```

Run: `cd backend && task proto:gen`
Expected: Go 與 TS 產生碼更新無錯。

- [ ] **Step 2: 寫失敗測試**

`backend/internal/domain/auth/ability_test.go`:

```go
package auth_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestGetAbilityWithConditions(t *testing.T) {
	client := testutil.NewEntClient(t)
	ctx := context.Background()
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("cancel").
		SetConditions(map[string]any{"status": "pending"}).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetConditions(map[string]any{"department_id": "${user.department_id}"}).SetSortOrder(1).SaveX(ctx)

	h := NewAbilityHandler(client) // 01 計畫既有 handler 建構式,依其實際簽名調整
	req := connect.NewRequest(&v1.GetAbilityRequest{})
	resp, err := h.GetAbility(
		middleware.WithIdentity(ctx, middleware.Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1", Roles: []string{"staff"}}),
		req,
	)
	require.NoError(t, err)
	rules := resp.Msg.GetRules()
	require.Len(t, rules, 2)

	assert.Equal(t, "cancel", rules[0].GetAction())
	assert.Equal(t, "pending", rules[0].GetConditions().GetFields()["status"].GetStringValue())

	assert.Equal(t, "read", rules[1].GetAction())
	// 佔位符已展開
	assert.Equal(t, "d1", rules[1].GetConditions().GetFields()["department_id"].GetStringValue())
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -run TestGetAbilityWithConditions -v`
Expected: FAIL — 現行實作無 conditions 輸出（或建構式簽名不符）

- [ ] **Step 4: 實作**

`ability.go` 核心（替換 01 計畫 1.8.1 的規則產生段）:

```go
func (h *AbilityHandler) GetAbility(ctx context.Context, req *connect.Request[v1.GetAbilityRequest]) (*connect.Response[v1.GetAbilityResponse], error) {
	id := middleware.IdentityFrom(ctx)
	if id.Role == "developer" && h.cfg.DeveloperAccountEnabled {
		return connect.NewResponse(&v1.GetAbilityResponse{Rules: []*v1.AbilityRule{{Action: "manage", Subject: "all"}}}), nil
	}
	rows, err := h.db.RolePermission.Query().
		Where(rolepermission.HasRoleWith(role.CodeIn(id.Roles...))). // edge/predicate 名依產生碼
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rules := make([]*v1.AbilityRule, 0, len(rows))
	for _, rp := range rows {
		conds, err := casl.ParseConditions(rp.Conditions)
		if err != nil {
			continue // fail-closed
		}
		expanded, ok := casl.ExpandForIdentity(conds, casl.Identity{
			UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, CustomerID: id.CustomerID,
		})
		if !ok {
			continue // 佔位符展開失敗 → 規則不下發(前端亦無從命中)
		}
		rule := &v1.AbilityRule{Action: rp.Action, Subject: rp.Resource, Inverted: rp.Inverted}
		if len(expanded) > 0 {
			s, err := structpb.NewStruct(condsToMap(expanded))
			if err != nil {
				continue
			}
			rule.Conditions = s
		}
		rules = append(rules, rule)
	}
	return connect.NewResponse(&v1.GetAbilityResponse{Rules: rules}), nil
}
```

配套：`evaluator.go` 抽出匯出函式（NewEvaluator 內部改呼叫它）:

```go
// ExpandForIdentity 展開佔位符;任一佔位符對應身分值為空 → ok=false(規則停用)。
func ExpandForIdentity(conds []FieldCondition, id Identity) (out []FieldCondition, ok bool) {
	for _, c := range conds {
		if s, isStr := c.Value.(string); isStr {
			if fn, isPH := placeholders[s]; isPH {
				v := fn(id)
				if v == "" {
					return nil, false
				}
				c.Value = v
			}
		}
		out = append(out, c)
	}
	return out, true
}

func condsToMap(conds []FieldCondition) map[string]any { /* Field→(op→value 或裸值) 還原 CASL JSON 形 */ }
```

`condsToMap` 需還原 CASL JSON 形：`$eq` 輸出裸值、其餘運算子輸出 `{op: value}`、同欄位多運算子合併一個物件。實作：

```go
func condsToMap(conds []FieldCondition) map[string]any {
	out := map[string]any{}
	for _, c := range conds {
		if c.Op == OpEq {
			out[c.Field] = c.Value
			continue
		}
		m, _ := out[c.Field].(map[string]any)
		if m == nil {
			m = map[string]any{}
			out[c.Field] = m
		}
		m[string(c.Op)] = c.Value
	}
	return out
}
```

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go test ./internal/domain/auth/ -run TestGetAbilityWithConditions -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/proto/v1/ability.proto backend/gen/ backend/internal/domain/auth/ability.go backend/internal/authz/casl/evaluator.go
git commit -m "feat(backend): GetAbility 輸出 conditions/inverted,佔位符後端展開"
```

---

### Task 10: UpdatePermissions conditions 驗證 + 防鎖死延伸 + ListConditionFields

**Files:**
- Modify: `backend/proto/v1/roles.proto`（`PermissionRule` 加 `conditions`/`inverted`/`sort_order`；新增 `ListConditionFields` RPC）
- Modify: `backend/internal/domain/roles/usecase.go`、`handler.go`
- Test: `backend/internal/domain/roles/usecase_casl_test.go`

**Interfaces:**
- Consumes: Task 6 `ValidateRuleConditions`/`ConditionFields`；Task 8 `authz`。
- Produces:
  - `RoleService.ListConditionFields(resource) → { fields: [{field, type, ops[], enum[]}] }`（供前端條件建構器）。
  - `UpdatePermissions` 寫入前：每條規則 `ValidateRuleConditions`；未知值 → `invalid_argument`。
  - 防鎖死：目標角色 ∈ 操作者自身角色 且 resource 為權限管理類（`role`/`policy`）時，以操作者身分代入驗證更新後仍可 `manage`/`read`/`update` 該 resource，否則 `failed_precondition`。
  - `company_admin`：id 類欄位值須為 `${user.company_id}` 佔位符或屬自己公司的值，否則 `invalid_argument`。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/roles/usecase_casl_test.go`:

```go
package roles_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
)

func TestUpdatePermissionsConditionValidation(t *testing.T) {
	uc, ctx := newTestUsecase(t) // 02 計畫既有測試輔助,注入 super 身分

	t.Run("未知欄位 invalid_argument", func(t *testing.T) {
		_, err := uc.UpdatePermissions(ctx, connect.NewRequest(&v1.UpdatePermissionsRequest{
			RoleId: testRoleID(t, ctx, "staff"),
			Rules: []*v1.PermissionRule{{Action: "read", Resource: "sales_order",
				Conditions: structOf(t, map[string]any{"nope": 1})}},
		}))
		require.Error(t, err)
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	t.Run("enum 值非法 invalid_argument", func(t *testing.T) {
		_, err := uc.UpdatePermissions(ctx, connect.NewRequest(&v1.UpdatePermissionsRequest{
			RoleId: testRoleID(t, ctx, "staff"),
			Rules: []*v1.PermissionRule{{Action: "read", Resource: "sales_order",
				Conditions: structOf(t, map[string]any{"status": "bogus"})}},
		}))
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	t.Run("防鎖死:排除自身的權限管理條件被拒", func(t *testing.T) {
		_, err := uc.UpdatePermissions(ctx, connect.NewRequest(&v1.UpdatePermissionsRequest{
			RoleId: testRoleID(t, ctx, "super"),
			Rules: []*v1.PermissionRule{{Action: "manage", Resource: "role",
				Conditions: structOf(t, map[string]any{"company_id": "nonexistent-company"})}},
		}))
		assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	})

	t.Run("合法 conditions 寫入並可讀回", func(t *testing.T) {
		_, err := uc.UpdatePermissions(ctx, connect.NewRequest(&v1.UpdatePermissionsRequest{
			RoleId: testRoleID(t, ctx, "staff"),
			Rules: []*v1.PermissionRule{{Action: "cancel", Resource: "sales_order",
				Conditions: structOf(t, map[string]any{"status": "pending"})}},
		}))
		require.NoError(t, err)
	})
}

func TestListConditionFields(t *testing.T) {
	uc, ctx := newTestUsecase(t)
	resp, err := uc.ListConditionFields(ctx, connect.NewRequest(&v1.ListConditionFieldsRequest{Resource: "sales_order"}))
	require.NoError(t, err)
	names := []string{}
	for _, f := range resp.Msg.GetFields() {
		names = append(names, f.GetField())
	}
	assert.Contains(t, names, "status")
	assert.Contains(t, names, "department_id")
}
```

測試輔助（同檔定義；`roles.NewUsecase` 建構式若與 02 計畫產物不同，以 02 產物為準調整，斷言不變）：

```go
func structOf(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	require.NoError(t, err)
	return s
}

// newTestUsecase 建測試用 usecase 並回傳帶 super 身分的 ctx。
func newTestUsecase(t *testing.T) (*roles.Usecase, context.Context) {
	t.Helper()
	client := testutil.NewEntClient(t)
	testutil.SeedBuiltinRoles(t, client) // 02 計畫的 seeder;無此輔助時於此展開七角色 seed
	uc := roles.NewUsecase(client, authz.Registry(context.Background()))
	ctx := middleware.WithIdentity(context.Background(),
		middleware.Identity{UserID: "super-1", Roles: []string{"super"}})
	return uc, ctx
}

func testRoleID(t *testing.T, ctx context.Context, code string) string {
	t.Helper()
	// 由 ctx 中的 testutil client 查 roles.code;實際取用依 02 計畫的測試輔助調整
	return code // 若 Usecase 以 code 解析則直接可用;否則於此查表回 UUID
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/roles/ -run 'TestUpdatePermissionsConditionValidation|TestListConditionFields' -v`
Expected: FAIL — 驗證與 RPC 未實作

- [ ] **Step 3: 實作**

proto 增量（`roles.proto`）:

```proto
message PermissionRule {
  string action = 1;
  string resource = 2;
  optional google.protobuf.Struct conditions = 3;
  bool inverted = 4;
  int32 sort_order = 5;
}

message ListConditionFieldsRequest { string resource = 1; }
message ConditionField {
  string field = 1;
  string type = 2;           // string | enum | id | number
  repeated string ops = 3;   // "$eq" ...
  repeated string enum = 4;
}
message ListConditionFieldsResponse { repeated ConditionField fields = 1; }

// 於 RoleService 內追加:
// rpc ListConditionFields(ListConditionFieldsRequest) returns (ListConditionFieldsResponse);
```

usecase 驗證段（UpdatePermissions 寫入前呼叫）。檔頭定義錯誤值：

```go
var (
	errCompanyScope = errors.New("company_admin 的 company_id 條件值限 ${user.company_id} 佔位符或自己公司")
	errLockout      = errors.New("此異動會排除操作者自身的權限管理能力,已拒絕(防鎖死)")
)
```

```go
func (u *Usecase) validateRules(ctx context.Context, actor middleware.Identity, targetRole *ent.Role, rules []*v1.PermissionRule) error {

```go
func (u *Usecase) validateRules(ctx context.Context, actor middleware.Identity, targetRole *ent.Role, rules []*v1.PermissionRule) error {
	reg := authz.Registry(ctx)
	for _, r := range rules {
		var raw map[string]any
		if r.Conditions != nil {
			raw = r.Conditions.AsMap()
		}
		conds, err := casl.ParseConditions(raw)
		if err != nil {
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		if err := reg.ValidateRuleConditions(r.Resource, conds); err != nil {
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		// company_admin:id 類條件值須為佔位符或自己公司範圍
		if actor.Role == "company_admin" {
			for _, c := range conds {
				fd, ferr := reg.Field(r.Resource, c.Field)
				if ferr != nil {
					return connect.NewError(connect.CodeInvalidArgument, ferr)
				}
				if fd.Type == casl.TypeID && c.Field == "company_id" {
					if s, ok := c.Value.(string); ok && s != "${user.company_id}" && s != actor.CompanyID {
						return connect.NewError(connect.CodeInvalidArgument, errCompanyScope)
					}
				}
			}
		}
	}
	// 防鎖死:目標角色含操作者自身角色且動到權限管理 resource
	if slices.Contains(actor.Roles, targetRole.Code) {
		for _, r := range rules {
			if r.Resource != "role" && r.Resource != "policy" {
				continue
			}
			if ruleExcludesActor(r, actor) {
				return connect.NewError(connect.CodeFailedPrecondition, errLockout)
			}
		}
	}
	return nil
}

// ruleExcludesActor:帶條件的權限管理規則,以操作者身分代入後不命中 → 會排除自己
func ruleExcludesActor(r *v1.PermissionRule, actor middleware.Identity) bool {
	if r.Conditions == nil {
		return false
	}
	conds, err := casl.ParseConditions(r.Conditions.AsMap())
	if err != nil {
		return true // 解析失敗視為排除(fail-closed)
	}
	expanded, ok := casl.ExpandForIdentity(conds, casl.Identity{
		UserID: actor.UserID, CompanyID: actor.CompanyID,
		DepartmentID: actor.DepartmentID, CustomerID: actor.CustomerID,
	})
	if !ok {
		return true
	}
	e := casl.NewEvaluator([]casl.Rule{{Action: r.Action, Subject: r.Resource, Conditions: expanded, Inverted: r.Inverted}},
		casl.Identity{})
	inst := map[string]any{"company_id": actor.CompanyID, "department_id": actor.DepartmentID}
	return !e.Can(r.Action, r.Resource, inst)
}
```

`ListConditionFields` handler：直接映射 `reg.ConditionFields(resource)` 至 proto；未知 resource 回空陣列。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && task proto:gen && go test ./internal/domain/roles/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/proto/v1/roles.proto backend/gen/ backend/internal/domain/roles/
git commit -m "feat(backend): UpdatePermissions conditions 驗證 + 防鎖死延伸 + ListConditionFields"
```

---

### Task 11: 前端 ability service + context

**Files:**
- Modify: `frontend/package.json`（dependencies 加 `@casl/ability`，**不加 `@casl/solid`**）
- Create: `frontend/src/lib/ability/service.ts`
- Create: `frontend/src/lib/ability/context.tsx`
- Test: `frontend/src/lib/ability/service.test.ts`

**Interfaces:**
- Consumes: Task 9 proto 產生的 TS 型別（`AbilityRule`）；Connect client（WEB-INF-06 產物，`~/lib/transport`）。
- Produces:
  - `abilityQueryOptions`（`queryKey: ["ability"]`、`staleTime: 60_000`、`queryFn` 取 GetAbility）。
  - `createAppAbility(rules: AbilityRule[]): AppAbility`。
  - `<AbilityProvider>` / `useAbility(): Accessor<AppAbility>`；`AppAbility = MongoAbility`。

- [ ] **Step 1: 安裝依賴**

Run: `cd frontend && pnpm add @casl/ability`
Expected: package.json dependencies 出現 `@casl/ability`（7.x）。

- [ ] **Step 2: 寫失敗測試**

`frontend/src/lib/ability/service.test.ts`:

```ts
import { subject } from "@casl/ability";
import { describe, expect, it } from "vitest";
import { createAppAbility } from "./service";

describe("createAppAbility", () => {
  it("由 proto 規則建 ability,支援 conditions 與 inverted", () => {
    const ability = createAppAbility([
      { action: "read", subject: "sales_order", conditions: { status: "pending" } },
      { action: "cancel", subject: "sales_order", inverted: true },
    ] as any);
    expect(ability.can("read", "sales_order")).toBe(true);
    expect(ability.can("cancel", "sales_order")).toBe(false);
    expect(ability.can("read", "customer")).toBe(false);
  });
  it("conditions 參與 instance 判斷", () => {
    const ability = createAppAbility([
      { action: "cancel", subject: "sales_order", conditions: { status: "pending" } },
    ] as any);
    expect(ability.can("cancel", subject("sales_order", { status: "pending" }))).toBe(true);
    expect(ability.can("cancel", subject("sales_order", { status: "processing" }))).toBe(false);
  });
  it("空規則 fail-closed", () => {
    const ability = createAppAbility([]);
    expect(ability.can("read", "sales_order")).toBe(false);
  });
});
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd frontend && pnpm run test -- src/lib/ability/service.test.ts`
Expected: FAIL — module not found

- [ ] **Step 4: 實作**

`frontend/src/lib/ability/service.ts`:

```ts
import { createMongoAbility, type MongoAbility, type RawRuleOf } from "@casl/ability";
import { queryOptions } from "@tanstack/solid-query";
import { createClient } from "@connectrpc/connect";
import { AbilityService } from "~/gen/proto/v1/ability_pb";
import { transport } from "~/lib/transport";

export type AppAbility = MongoAbility;

const client = createClient(AbilityService, transport);

export const abilityQueryOptions = queryOptions({
  queryKey: ["ability"],
  staleTime: 60_000, // 規格 60 秒 TTL;路由切換不重取
  queryFn: async () => (await client.getAbility({})).rules,
});

export function createAppAbility(rules: ReadonlyArray<{
  action: string;
  subject: string;
  conditions?: Record<string, unknown>;
  inverted?: boolean;
}>): AppAbility {
  const raw: RawRuleOf<AppAbility>[] = rules.map((r) => ({
    action: r.action,
    subject: r.subject,
    ...(r.conditions ? { conditions: r.conditions } : {}),
    ...(r.inverted ? { inverted: true as const } : {}),
  }));
  return createMongoAbility(raw);
}
```

`frontend/src/lib/ability/context.tsx`:

```tsx
import { createContext, useContext, type Accessor, type ParentComponent } from "solid-js";
import type { AppAbility } from "./service";

const AbilityContext = createContext<Accessor<AppAbility>>();

export const AbilityProvider: ParentComponent<{ ability: Accessor<AppAbility> }> = (props) => (
  <AbilityContext.Provider value={props.ability}>{props.children}</AbilityContext.Provider>
);

// 回傳 accessor 而非實例:ability 更新=整個實例替換,JSX 內引用點隨 signal 重算
export function useAbility(): Accessor<AppAbility> {
  const acc = useContext(AbilityContext);
  if (!acc) throw new Error("useAbility must be used within <AbilityProvider>");
  return acc;
}
```

- [ ] **Step 5: 跑測試確認通過**

Run: `cd frontend && pnpm run test -- src/lib/ability/service.test.ts`
Expected: PASS（3 測試全綠）

- [ ] **Step 6: Commit**

```bash
git add frontend/package.json frontend/pnpm-lock.yaml frontend/src/lib/ability/
git commit -m "feat(frontend): ability service + context(@casl/ability)"
```

---

### Task 12: `<Can>` 顯示控制元件

**Files:**
- Create: `frontend/src/lib/ability/Can.tsx`
- Test: `frontend/src/lib/ability/Can.test.tsx`

**Interfaces:**
- Consumes: Task 11 `useAbility`/`AppAbility`。
- Produces: `<Can I="cancel" a="sales_order" | {subject("sales_order", order)} fallback={...}>`；instance 以 CASL `subject(type, obj)` 包裝傳入（前端唯一 instance 包裝慣例，寫入元件 doc comment）。

- [ ] **Step 1: 寫失敗測試**

`frontend/src/lib/ability/Can.test.tsx`:

```tsx
import { describe, expect, it } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { createSignal } from "solid-js";
import { subject } from "@casl/ability";
import { AbilityProvider } from "./context";
import { createAppAbility } from "./service";
import { Can } from "./Can";

function renderWithAbility(rules: any[], ui: () => any) {
  const [ability] = createSignal(createAppAbility(rules));
  return render(() => <AbilityProvider ability={ability}>{ui()}</AbilityProvider>);
}

describe("<Can>", () => {
  it("允許時顯示 children", () => {
    renderWithAbility([{ action: "read", subject: "sales_order" }], () => (
      <Can I="read" a="sales_order"><button>列表</button></Can>
    ));
    expect(screen.getByText("列表")).toBeTruthy();
  });
  it("拒絕時顯示 fallback", () => {
    renderWithAbility([], () => (
      <Can I="read" a="sales_order" fallback={<span>無權限</span>}><button>列表</button></Can>
    ));
    expect(screen.queryByText("列表")).toBeNull();
    expect(screen.getByText("無權限")).toBeTruthy();
  });
  it("instance 條件判斷(按鈕依狀態灰化)", () => {
    const order = subject("sales_order", { status: "processing" });
    renderWithAbility([{ action: "cancel", subject: "sales_order", conditions: { status: "pending" } }], () => (
      <Can I="cancel" a={order} fallback={<button disabled>取消</button>}><button>取消</button></Can>
    ));
    expect((screen.getByText("取消") as HTMLButtonElement).disabled).toBe(true);
  });
});
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd frontend && pnpm run test -- src/lib/ability/Can.test.tsx`
Expected: FAIL — module not found

- [ ] **Step 3: 實作**

`frontend/src/lib/ability/Can.tsx`:

```tsx
import { Show, type JSX, type ParentComponent } from "solid-js";
import { useAbility } from "./context";

/**
 * 顯示控制元件。instance 判斷時以 CASL subject() 包裝傳入:
 *   <Can I="cancel" a={subject("sales_order", order)} fallback={<Disabled/>}>
 */
export const Can: ParentComponent<{
  I: string;
  a: string | object;
  fallback?: JSX.Element;
}> = (props) => {
  const ability = useAbility();
  const allowed = () => ability().can(props.I, props.a as never);
  return <Show when={allowed()} fallback={props.fallback}>{props.children}</Show>;
};
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd frontend && pnpm run test -- src/lib/ability/Can.test.tsx`
Expected: PASS（3 測試全綠）

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/ability/Can.tsx frontend/src/lib/ability/Can.test.tsx
git commit -m "feat(frontend): <Can> 顯示控制元件"
```

---

### Task 13: route guard 與快取接線

**Files:**
- Create: `frontend/src/lib/ability/guards.ts`
- Modify: `frontend/src/lib/query-client.ts`（WEB-INF-04 產物；若不存在則 Create，內容如下）
- Test: `frontend/src/lib/ability/guards.test.ts`

**Interfaces:**
- Consumes: Task 11 `abilityQueryOptions`/`createAppAbility`；`@solidjs/router` `redirect`。
- Produces: `requireAbility(action, subjectType)` — 用於路由定義 `load`/`beforeLoad`；deny → `throw redirect("/403")`。權限異動後主動重載慣例：`queryClient.invalidateQueries({ queryKey: ["ability"] })`（權限設置頁儲存成功後呼叫，2.7/2.11 前端計畫引用）。

- [ ] **Step 1: 寫失敗測試**

`frontend/src/lib/ability/guards.test.ts`:

```ts
import { describe, expect, it, vi, beforeEach } from "vitest";
import { QueryClient } from "@tanstack/solid-query";

const mockGetAbility = vi.fn();
vi.mock("./service", async (orig) => {
  const mod = await orig<typeof import("./service")>();
  return {
    ...mod,
    abilityQueryOptions: {
      ...mod.abilityQueryOptions,
      queryFn: async () => (await mockGetAbility()).rules,
    },
  };
});

import { makeRequireAbility } from "./guards";

describe("requireAbility", () => {
  beforeEach(() => mockGetAbility.mockReset());

  it("有權限時放行", async () => {
    mockGetAbility.mockResolvedValue({ rules: [{ action: "read", subject: "sales_order" }] });
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    await expect(guard()).resolves.toBeUndefined();
  });
  it("無權限時 redirect /403", async () => {
    mockGetAbility.mockResolvedValue({ rules: [] });
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    // @solidjs/router redirect 拋出物件的形狀隨版本不同(href/url/Response headers),取任一判斷
    try {
      await guard();
      expect.unreachable("應拋出 redirect");
    } catch (e: any) {
      const href = e?.href ?? e?.url ?? e?.headers?.get?.("Location");
      expect(href).toContain("/403");
    }
  });
});
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd frontend && pnpm run test -- src/lib/ability/guards.test.ts`
Expected: FAIL — module not found

- [ ] **Step 3: 實作**

`frontend/src/lib/ability/guards.ts`:

```ts
import { redirect } from "@solidjs/router";
import type { QueryClient } from "@tanstack/solid-query";
import { queryClient as defaultClient } from "~/lib/query-client";
import { abilityQueryOptions, createAppAbility } from "./service";

// makeRequireAbility 供測試注入 QueryClient;requireAbility 為正式綁定
export function makeRequireAbility(qc: QueryClient) {
  return (action: string, subjectType: string) => async () => {
    const rules = await qc.ensureQueryData(abilityQueryOptions);
    if (!createAppAbility(rules).can(action, subjectType)) {
      throw redirect("/403");
    }
  };
}

// 用法(路由定義): { path: "/orders", component: OrdersPage, load: requireAbility("read", "sales_order") }
export const requireAbility = makeRequireAbility(defaultClient);
```

`frontend/src/lib/query-client.ts`（若 WEB-INF-04 已建則僅確認匯出名 `queryClient`，不覆寫）:

```ts
import { QueryClient } from "@tanstack/solid-query";

export const queryClient = new QueryClient();
```

- [ ] **Step 4: 跑測試確認通過 + 前端全測回歸**

Run: `cd frontend && pnpm run test`
Expected: PASS（ability 四檔測試全綠，既有測試不破壞）

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/ability/guards.ts frontend/src/lib/ability/guards.test.ts frontend/src/lib/query-client.ts
git commit -m "feat(frontend): requireAbility 路由守衛與 ability 快取接線"
```

---

## 整合測試重點（D21，隨各 Task 測試落地，上線前於 05-sales-orders 計畫補端對端）

0. **範圍界線**：權限設置頁的條件建構器 UI 元件屬前端權限頁計畫（原計畫 Task 2.7/2.11 範圍），本計畫提供其全部後端契約（Task 10 的 `ListConditionFields` 白名單與 `UpdatePermissions` 驗證）與前端 ability 消費層（Task 11–13）。
1. **env on/off**：開啟時 list 過濾與實例檢查生效；關閉時行為 = 純 Casbin+RLS（Task 8 測試覆蓋）。
2. **CASL × RLS 交集**：規則條件與 RLS 結果取交集，互不放大（Task 8 整合測試 + 05 計畫端對端）。
3. **同源生效鏈**：UpdatePermissions（含 conditions）→ GetAbility 輸出 → 前端 ability 更新（Task 9/10/11 測試串聯語意）。
4. **防鎖死**：自身角色權限管理排除性條件被 `failed_precondition` 擋下（Task 10）。
5. **golden 對賭**：CI 重跑 `casl-golden-gen.mjs` 後 `git diff --exit-code` 確保 fixture 最新（Phase 0 CI 管線加入此步）。

---

*計畫版本：v1.0.0（2026-08-24）；對應設計 `2026-08-24-casl-integration-design.md`、決策 D30 系列、規格書 v1.0.35。*
