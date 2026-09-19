# 前端表單 Phase 2 — TanStack Form ＋ valibot 欄位級驗證（設計）

> **狀態**：設計已定案（2026-09-19），待實作計畫
> **範圍**：`frontend/` 的三個表單
> **前置**：Phase 1（元件庫庫化 + Ark 靜態元件 + sidebar + 深色模式）已完成並通過全分支複審

---

## 0. 背景

使用者 directive 的四階段拆分（有消費者才做）：Phase 1 元件庫（**已完成**）→ **Phase 2 表單（本 spec）** → Phase 3 TanStack Table → Phase 4 Pragmatic D&D（等派車畫面）。

**現況（實查）**：三個表單都是手寫 signal + HTML `required`；錯誤只有**單一** `formError` banner（`<p role="alert">` 或 `FieldError`）；`valibot` 雖在依賴中，但 `features/**` 完全未使用（0 命中 `safeParse`）。Phase 1 已把 `Field` 改成 Ark（`invalid` prop + `FieldError` 走 Ark a11y 關聯），所以欄位級錯誤的**基礎設施已就位**。

## 1. 目標與範圍

**In**
- 三個表單改用 `@tanstack/solid-form`：**登入頁「店家」分頁**、**公司 modal**、**部門 modal**
- 以 `valibot` 定義欄位規則（**只鏡射現況的 `required`**）
- 錯誤呈現改為：客戶端驗證 → 該欄位下方（Ark `Field invalid` + `FieldError`）；伺服器錯誤 → 保留表單層 banner

**Out**
- 篩選列（公司／部門的關鍵字＋狀態查詢表單）與角色權限矩陣的儲存
- 新增後端規則、改變 API 契約、`src/lib/**`
- Phase 3/4（TanStack Table、Pragmatic D&D）
- Phase 1 遺留的 housekeeping（另立清單，見 §7）

## 2. 依賴與分工

**新依賴（1 個）**：`@tanstack/solid-form@1.33.5`（peer `solid-js >=1.9.9`；本專案 1.9.15 ✔）。

| 角色 | 負責 | 不負責 |
|---|---|---|
| **TanStack Form** | 欄位值、`touched`/`dirty`、驗證時機與執行、`handleSubmit`、`isSubmitting`/`canSubmit`、表單層錯誤 | 樣式、a11y 屬性 |
| **valibot**（已安裝） | 規則與繁中錯誤訊息 | 狀態管理 |
| **Ark `Field` / `FieldError`**（Phase 1） | `aria-invalid`、`aria-describedby` 關聯、錯誤顯示 | 驗證邏輯 |

**已查證的 API（`@tanstack/solid-form` v1，Solid 版）**：
- `createForm(() => ({ defaultValues, onSubmit, validators }))`（**不是** `useForm`）
- `<form.Field name children={(field) => …}>`，`field()` 是 accessor：`field().state.value`、`field().state.meta.errors`、`field().state.meta.errorMap`、`field().state.meta.isValid`、`field().handleChange(v)`、`field().handleBlur()`
- 欄位驗證：`validators: { onBlur, onSubmit, onChange }`，回傳錯誤字串（或物件）
- 表單層：`form.handleSubmit()`、`form.Subscribe`（selector）/`form.useSelector`、`onSubmitAsync` 可用 `fields: {}` 設定欄位錯誤

**valibot 接入方式**：以 `v.safeParse(schema, value)` 包在 validator 內（明確、不依賴版本特性）。若實作時發現安裝版本支援 Standard Schema 而可直接傳 schema，可用之，但**必須**在報告說明並以測試守住錯誤訊息來源。

## 3. 三個表單的設計

| 表單 | 欄位 | 規則（**只鏡射現況 `required`**） | 錯誤落點 |
|---|---|---|---|
| 登入頁（店家分頁） | `customerCode`、`password` | 兩者必填、非空字串 | 各欄下方 |
| 公司 modal | `name`、`identifier`、`taxId`、`status` | `name`/`identifier` 必填；`taxId`/`status` 選填（`status` 有預設值） | 必填兩欄下方 |
| 部門 modal | `name`、`company`（select） | 兩者必填 | 各欄下方 |

**刻意不做的事**：不新增格式驗證（例如編號格式、統編格式）——後端才有真正的規則，客戶端自行發明會擋下後端接受的值。

**modal 的 reset/dirty**：現況是「每次開啟都顯式重設 signal」；改用 TanStack Form 後需在開啟時 `form.reset()`（或重建 form）並確認：關閉再開不會殘留上次的錯誤訊息與 touched 狀態。**這是本階段最容易出錯處，需逐一比對。**

**哪些 signal 由表單接管（避免拆錯邊界）**：
- **由 `createForm` 接管並移除**：表單欄位值（`name`/`identifier`/`taxId`/`status`、`customerCode`/`password`、部門的 `name`/`company`）、`saving`/`submitting`（改由 `isSubmitting`）、`formError`（改由表單層驗證 + banner 由提交流程設定）。
- **必須保留**：列表資料、分頁（`page`/`total`）、篩選條件（關鍵字／狀態）、`editing`/`dialogOpen` 這類 UI 狀態 —— 它們不是表單欄位（篩選列本身在本階段範圍外）。

## 4. 驗證時機與錯誤呈現

- **時機**：`onBlur` + `onSubmit`（不在每次按鍵就標紅）；提交失敗後該欄有錯即顯示。
- **客戶端錯誤** → `Field invalid` 為真 + `FieldError` 顯示該欄第一條錯誤（`field().state.meta.errors`）。
- **伺服器錯誤** → 表單層 banner（沿用現有 `role="alert"` 樣式），**不**猜測要掛到哪一欄。
- 唯一的例外：若伺服器錯誤**明確可對應**到某欄（例如識別碼重複），才額外顯示在該欄；**本階段不新增 proto 欄位或字串比對**，判斷不出來就維持 banner。

## 5. 測試與驗收

**測試（Vitest + jsdom + testing-library）**
- schema 單元測試：每個表單的「必填未填 → 有錯誤」「填了 → 無錯誤」
- 元件測試（每個表單至少）：
  1. 未填直接提交 → 對應欄出現錯誤、`aria-invalid="true"`、`aria-describedby` **指向該錯誤元素**（沿用 Phase 1 的斷言風格）
  2. 填入合法值 → 錯誤消失
  3. 成功提交 → 以正確的 payload 呼叫對應 API（可用 spy 驗證）
  4. 伺服器錯誤 → banner 出現且 `role="alert"`
  5. modal：關閉再開 → 錯誤與 touched 狀態已重設
- 禁則沿用前兩階段：不攔截 rAF、不 mock Ark 內部、不斷言 Ark 私有屬性、不留空轉斷言

**驗收**
1. `pnpm typecheck && pnpm lint && pnpm test && pnpm build` 全綠、**lint 0 warning**
2. Edge 實測：三個表單的錯誤態與成功態；`/login` 兩分頁；modal 開關（含重開後的乾淨狀態）；**深色以真實切換器切換**
3. 既有的登入／CRUD 行為不變（伺服器錯誤仍以 banner 呈現）
4. `impeccable detect` 掃改動面
5. 收尾衛生（Edge 程序 0、暫存檔不入 repo）

## 6. 風險

- **R1（最可能出錯）**：modal 表單的 reset/dirty 語意與現行「顯式重設 signal」不同 → 以測試守住「關閉再開不殘留」。
- **R2**：`@tanstack/solid-form` 首次引入，bundle 會再增；沿用 Phase 1 的裁定（接受並記錄，不拆 chunk）。
- **R3**：valibot 規則若與後端不一致，會出現「客戶端擋下、後端接受」→ 因此**只鏡射現有 `required`**。
- **R4**：登入頁的 submit 路徑曾在 Phase 1 驗收時疑似 promise 不 settle（最終複審已以 HEAD 重跑結案為非回歸）→ 改寫時**必須重驗該路徑**。
- **R5**：`Field` 目前未開放 `required`（型別擋下，只能下在控件上）→ 必填標示仍以「標籤文字 + 屬性的既有做法」呈現，不擴大元件 API（該項列 Phase 1 housekeeping）。

## 7. 後續

- **Phase 2 開頭的 housekeeping（來自 Phase 1 最終複審，皆 1–3 行）**：`PaginationProps` 收斂 `defaultPage`/`defaultPageSize`；`FieldDescription` 補 `Field` 外 fallback（目前會 TypeError）；`input.tsx` 的 `aria-describedby` 改走 Ark `getInputProps()`（可能指向不存在的元素）；`ui/**` 統一 `cn` alias（`theme.tsx`/`parts.tsx` 用 `~/lib/cn`，其餘 11 檔用 `@/lib/cn`）。
- **Phase 3**：TanStack Table 重寫三張表（`Table` 元件的 markup 保留、排序/篩選/分頁狀態交 TanStack）。
- **Phase 4**：Pragmatic drag and drop 接入派車看板（待該畫面落地）。

---

*建立：2026-09-19*
