# 前端表單 Phase 2 實作計畫（精簡版）

> **For agentic workers:** REQUIRED SUB-SKILL: use superpowers:subagent-driven-development. Steps use `- [ ]`.
> **Goal:** 三個表單改用 `@tanstack/solid-form` + `valibot` 欄位級驗證，錯誤走 Ark `Field`/`FieldError`。
> **Tech Stack:** SolidJS 1.9、`@tanstack/solid-form@1.33.5`、valibot（已裝）、Ark `Field`（Phase 1）。
> **Spec:** `docs/superpowers/specs/2026-09-19-frontend-forms-tanstack-phase2-design.md`

## Global Constraints

- 新依賴**只允許** `@tanstack/solid-form`；不新增其他依賴。
- `createForm(() => ({ defaultValues, onSubmit, validators }))`（Solid 版**不是** `useForm`）；`<form.Field name children={(field) => …}>`，`field()` 為 accessor（`state.value`/`state.meta.errors`/`handleChange`/`handleBlur`）。
- valibot 以 `v.safeParse(schema, value)` 包進 validator（不依賴版本特性）。
- **規則只鏡射現況 `required`**：不新增格式驗證。
- **錯誤落點**：客戶端 → 該欄 `FieldError`（`Field invalid` 驅動）；伺服器 → 表單層 banner（`role="alert"`）；**不猜**要不要映射到欄位。
- **驗證時機**：`onBlur` + `onSubmit`（不在每次按鍵標紅）。
- 由 `createForm` 接管：欄位值、`isSubmitting`、表單層錯誤；**保留**：列表/分頁/篩選/`editing`/`dialogOpen`。
- 不改 `src/lib/**`、不改 API 契約、不動 `ui/**` 的對外 API。
- 測試禁則：不攔截 rAF、不 mock Ark 內部、不斷言 Ark 私有屬性、不留空轉斷言。
- `pnpm lint` 0 warning、四道 gate 全綠、繁中註解/commit。
- 瀏覽器用 **Microsoft Edge**（`--headless=new` + CDP）；**禁用** Chrome for Testing／`chrome-headless-shell`。收尾強制：`pkill -f "Microsoft Edge.*--headless"`、確認 0 殘留、`rm -rf /tmp/msedge-*`、repo 無暫存檔。
- Harness：shell `grep`/`rg` 被封鎖（用內建工具）；不要 `git add -A`。

---

## Task 1: 依賴 + 登入（店家分頁）表單

**Files:** `frontend/package.json`；`features/auth/pages/LoginPage.tsx`；新增 `features/auth/schemas.ts`（或同層 `login.schema.ts`）＋ `LoginPage` 的測試。

- [ ] **Step 1: 安裝** `cd frontend && pnpm add @tanstack/solid-form@1.33.5`；`pnpm ls @tanstack/solid-form solid-js` 確認 peer 無警告。
- [ ] **Step 2: schema（先寫）** `v.object({ customerCode: v.pipe(v.string(), v.nonEmpty("請輸入客戶編號")), password: … })`；同檔寫 schema 單元測試（未填→有錯、填了→無錯）。
- [ ] **Step 3: 表單測試（先紅）** jsdom：未填提交 → `getByText("請輸入客戶編號")` 出現且該 input `aria-invalid="true"`、`aria-describedby` 指向該錯誤元素；填入合法值 → 錯誤消失且以正確 payload 呼叫 `login`（spy）；伺服器錯誤 → banner `role="alert"` 出現。
- [ ] **Step 4: 實作** `createForm` 取代 `customerCode`/`password`/`submitting`/`error` 這幾個 signal；`<form.Field>` 包 `Field`/`FieldLabel`/`Input`/`FieldError`，`Field invalid` 由 `!field().state.meta.isValid` 驅動；`onBlur={field().handleBlur}` + `onInput={(e)=>field().handleChange(e.currentTarget.value)}`；提交走 `form.handleSubmit()`；伺服器錯誤設進表單層狀態並渲染 banner。
- [ ] **Step 5: 四道 gate**；**Step 6: Edge 實測**（錯誤態、成功態、店家/員工分頁互不影響）；**Step 7: commit** `refactor(frontend): 登入店家表單改用 TanStack Form + valibot 欄位驗證`。

## Task 2: 公司 modal 表單

**Files:** `features/users/pages/CompaniesPage.tsx`；新增 `features/users/schemas.ts`（或 `company.schema.ts`）＋ 測試。

- [ ] **Step 1: schema** `name`/`identifier` 必填（繁中訊息）、`taxId`/`status` 選填；單元測試。
- [ ] **Step 2: 測試（先紅）** ①未填提交 → 兩欄各自出現錯誤且 aria 關聯正確 ②填合法 → 錯誤消失、payload 正確 ③伺服器錯誤 → banner ④**關閉再開 → 錯誤與 touched 已重設**（R1）。
- [ ] **Step 3: 實作** 以 `createForm` 取代 `name`/`identifier`/`taxId`/`status`/`saving`/`formError`；開啟 modal 時 `form.reset()`（或重建）並帶入編輯值；`submit` 走 `form.handleSubmit()`；錯誤落點照 §4。
- [ ] **Step 4: 四道 gate**；**Step 5: Edge 實測**（新增/編輯、關閉再開、深色以真實切換器）；**Step 6: commit**。

## Task 3: 部門 modal 表單

**Files:** `features/users/pages/DepartmentsPage.tsx`；`features/users/schemas.ts`（與 Task 2 同檔則沿用，否則自建）＋ `DepartmentsPage` 的測試。

- [ ] **Step 1: schema** `name`（必填，繁中訊息）、`company`（必填，select 的值為公司 id）＋單元測試（未填→有錯、填了→無錯）。
- [ ] **Step 2: 測試（先紅）** ①未填提交 → `name` 與 `company` 各自出現錯誤且 `aria-invalid`/`aria-describedby` 正確 ②填合法 → 錯誤消失、payload 正確（含 `companyId`）③伺服器錯誤 → banner `role="alert"` ④**關閉再開 → 錯誤與 touched 已重設**（R1）。
- [ ] **Step 3: 實作** 以 `createForm` 取代 `name`/`company`/`saving`/`formError`；`company` 是原生 `<select>` → 以 `onInput={(e)=>field().handleChange(e.currentTarget.value)}`（或 `onChange`）接線，`Field invalid` 由該欄有效性驅動；開啟 modal 時 `form.reset()`（或重建）並帶入編輯值；提交走 `form.handleSubmit()`；錯誤落點照 §4。
- [ ] **Step 4: 四道 gate**；**Step 5: Edge 實測**（新增/編輯、關閉再開、深色以真實切換器）；**Step 6: commit** `refactor(frontend): 部門 modal 表單改用 TanStack Form + valibot 欄位驗證`。

## Task 4: 端到端驗收

- [ ] **Step 1** 四道 gate（`typecheck && lint && test && build`，lint 0 warning）→ 貼實際輸出。
- [ ] **Step 2** Edge 實測清單：三表單的錯誤/成功態、`/login` 兩分頁、modal 開關與重開後乾淨、CRUD 成功後列表更新、**深色以真實切換器**；未驗項註明理由（後端未啟動時以 `/tmp` 一次性 stub，不得入 repo）。
- [ ] **Step 3** `impeccable detect`（掃改動面）→ 有發現即判定修/不修並附理由。
- [ ] **Step 4** 收尾衛生並以指令輸出證明（Edge 0 殘留、`/tmp/msedge-*` 已刪、repo 無暫存檔）。
- [ ] **Step 5** 回報（含未驗項）。

---

## 驗收對照（spec §5）

| spec 驗收 | 步驟 |
|---|---|
| 1. 四道 gate + lint 0 warning | Task 4 Step 1 |
| 2. Edge 實測（三表單、兩分頁、modal 重開、深色真實切換） | Task 4 Step 2 |
| 3. 既有行為不變（伺服器錯誤仍 banner） | Task 1-3 的測試案例 + Task 4 Step 2 |
| 4. impeccable detect | Task 4 Step 3 |
| 5. 收尾衛生 | Task 4 Step 4 |

*建立：2026-09-19*
