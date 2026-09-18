# Backend 04 — 業務主檔 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-18 盤點。**本計畫對應領域（客戶/商品/倉別/車次/分切規格/分類/檔案/QR）實際尚未實作**——`ent/schema` 僅 company/department/role/rolepermission/user 五檔、migrations 僅 4 個、無 customers/products/warehouses 等實體。以下為保留的目標計畫架構，全數 ⬜ 未開始。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D7/D10/D12/D22`
> **細部文件**：`docs/superpowers/plans/backend/detail/04-master-data.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：01-auth（`issueTempPassword`、RLS、testutil）、02（軟刪除/範圍慣例）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | customers schema 與 CRUD（細部 3.1.1–3.1.2） | ⬜ 未開始 |
| 2 | customer_counters 同事務取號（細部 3.1.3，D7） | ⬜ 未開始 |
| 3 | 建檔連動主帳號 + 業務子帳號 + 偏好欄位（細部 3.1.4–3.1.5，D22） | ⬜ 未開始 |
| 4 | 地址簿與聯絡人（細部 3.2.1–3.2.2） | ⬜ 未開始 |
| 5 | 商品三實體與單位換算（細部 3.3.1–3.3.3） | ⬜ 未開始 |
| 6 | 倉別/車次/分切規格/分類 CRUD（細部 3.4.1–3.4.4） | ⬜ 未開始 |
| 7 | 客戶專屬商品（細部 3.5.1–3.5.3） | ⬜ 未開始 |
| 8 | 檔案資產（細部 3.6.1–3.6.3） | ⬜ 未開始 |
| 9 | QR 簽章 token 與兌換端點（細部 3.8.1–3.8.2） | ⬜ 未開始 |

**實作範圍**：0%（全部待辦）。本計畫為 04 領域（業務主檔）目標藍圖，尚未落地。

---

## Global Constraints（目標，精煉自原檔）

- 軟刪除 `deleted_at` + 部分唯一索引（D10）；取號/建檔/連動帳號皆同一 DB 交易（D18）。
- 客戶編號：`companies.customer_code_prefix`（大寫英數 1–4 碼）+ 6 位補零自增，公司內唯一不可改（D7）。
- 數量/換算率用 `shopspring/decimal`；1.0 不存任何單價/金額欄位（D12）。
- 上傳白名單：jpeg/png/webp ≤5MB、pdf ≤10MB；副檔名 + MIME + magic bytes 三重一致。
- `customer_products` 一客戶一商品一別名唯一；`default_qty = 0` 保留不顯示。
- 所有主檔表帶 `company_id` / `department_id`，沿用 RLS 注入。

## 待辦 Task 詳情

### Task 1: customers schema 與 CRUD（細部 3.1.1–3.1.2）
`customers` schema（公司/部門隸屬、送貨偏好）＋ CustomerService CRUD。

### Task 2: customer_counters 同事務取號（細部 3.1.3，D7）
客戶編號樂觀鎖取號（前綴 + 6 位自增，取號與建檔同交易）。

### Task 3: 建檔連動主帳號 + 業務子帳號 + 偏好欄位（細部 3.1.4–3.1.5，D22）
建立客戶自動附帶「業務子帳號」（專供所屬業務）；主帳號管理；`preferred_delivery_days` 偏好欄位；重用 `issueTempPassword`。

### Task 4: 地址簿與聯絡人（細部 3.2.1–3.2.2）
地址簿（多筆、預設）、聯絡人（稱呼/職稱/Email/電話）CRUD。

### Task 5: 商品三實體與單位換算（細部 3.3.1–3.3.3）
商品主檔 + 單位換算（decimal，如 1 條=0.6kg）、分切規格關聯。

### Task 6: 倉別/車次/分切規格/分類 CRUD（細部 3.4.1–3.4.4）
部門級實體表（warehouses/routes/cut_specs/product_categories）CRUD。

### Task 7: 客戶專屬商品（細部 3.5.1–3.5.3）
`customer_products`（業務建立、打開客戶自動帶入；別名機制；qty=0 處理）。

### Task 8: 檔案資產（細部 3.6.1–3.6.3）
FileStore 本地儲存（白名單 + magic bytes）；跨 domain 共用（02 Logo、06 退貨照片、09 PDF）。

### Task 9: QR 簽章 token 與兌換端點（細部 3.8.1–3.8.2）
客戶 QR 登入簽章 token 產生 + `/api/v1/...qrcode` 兌換端點（串 01-auth）。

---

*最後更新：2026-09-18（04-master-data 現況對齊重建；領域未實作）*
