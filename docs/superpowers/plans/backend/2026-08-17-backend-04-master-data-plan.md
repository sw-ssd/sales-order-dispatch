# Backend 04 — 業務主檔 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：**2026-09-22 更新**（原 2026-09-18 盤點時本領域全未實作）。2026-09-18~22 已分批落地 3.1–3.6、3.8（含 3.5 客戶專屬商品後端＋Web、3.6 檔案資產後端、3.8 QR 後端）；下表為現況。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D7/D10/D12/D22`
> **細部文件**：`docs/superpowers/plans/backend/detail/04-master-data.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：01-auth（`issueTempPassword`、RLS、testutil）、02（軟刪除/範圍慣例）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | customers schema 與 CRUD（細部 3.1.1–3.1.2） | ✅ 完成（2026-09-18）|
| 2 | customer_counters 同事務取號（細部 3.1.3，D7） | ✅ 完成（2026-09-18；2026-09-20 修 id 主鍵落差，真 PG 上取號不再失敗）|
| 3 | 建檔連動主帳號 + 業務子帳號 + 偏好欄位（細部 3.1.4–3.1.5，D22） | 🟡 部分（3.1.4 建檔連動＋臨時密碼交付已落地 2026-09-18；3.1.5 `preferred_delivery_days`/`promo_tag_ids` 欄位已存在，完整驗證與促銷連動待 07-promo_tags）|
| 4 | 地址簿與聯絡人（細部 3.2.1–3.2.2） | ✅ 完成（2026-09-18；`customer_address_service.go`/`customer_contact_service.go`，00015）|
| 5 | 商品三實體與單位換算（細部 3.3.1–3.3.3） | ✅ 完成（2026-09-19；`product_service.go`＋`domain/products/conversion.go`，00017）|
| 6 | 倉別/車次/分切規格/分類 CRUD（細部 3.4.1–3.4.4） | ✅ 完成（2026-09-19；`masters.proto` 四 service 各 5 法＋restore，00016）|
| 7 | 客戶專屬商品（細部 3.5.1–3.5.3） | ✅ 完成（2026-09-22；後端 `6bebf82`＋00033／00034、Web 對話框 `6192c13`）|
| 8 | 檔案資產（細部 3.6.1–3.6.3） | 🟡 部分（後端 `1bb4c76`＋00035／00036 已落地、09-printing 已消費；Web 上傳頁與 02 Logo 消費面待）|
| 9 | QR 簽章 token 與兌換端點（細部 3.8.1–3.8.2） | ✅ 完成（2026-09-22；`67ab31d` qrcode 簽章＋`QRLogin` 兌換＋`GetCustomerQRCode`；Web/App QR 畫面另案）|

**實作範圍**：約 90%（Task 1–2、4–7、9 完成；3、8 部分——殘：3.1.5 完整驗證與促銷連動（07-promo_tags）、Logo 上傳 Web 頁與消費面）。殘項與後續見 `2026-09-19-backend-04-task34-dept-masters-plan.md` §已知缺口。

---

## Global Constraints（目標，精煉自原檔）

- 軟刪除 `deleted_at` + 部分唯一索引（D10）；取號/建檔/連動帳號皆同一 DB 交易（D18）。
- 客戶編號：`companies.customer_code_prefix`（大寫英數 1–4 碼）+ 6 位補零自增，公司內唯一不可改（D7）。
- 數量/換算率用 `shopspring/decimal`；1.0 不存任何單價/金額欄位（D12）。
- 上傳白名單：jpeg/png/webp ≤5MB、pdf ≤10MB；副檔名 + MIME + magic bytes 三重一致。
- `customer_products` 一客戶一商品一別名唯一；`default_qty = 0` 保留不顯示。
- 所有主檔表帶 `company_id` / `department_id`，沿用 RLS 注入。

## Task 詳情

### Task 1: customers schema 與 CRUD（細部 3.1.1–3.1.2）
`customers` schema（公司/部門隸屬、送貨偏好）＋ CustomerService CRUD。

### Task 2: customer_counters 同事務取號（細部 3.1.3，D7）
客戶編號樂觀鎖取號（前綴 + 6 位自增，取號與建檔同交易）。

### Task 3: 建檔連動主帳號 + 業務子帳號 + 偏好欄位（細部 3.1.4–3.1.5，D22）
建立客戶自動附帶「業務子帳號」（專供所屬業務）；主帳號管理；`preferred_delivery_days` 偏好欄位；重用 `issueTempPassword`。
> ✅ 3.1.4（D22 建檔連動主/業務子帳號＋臨時密碼交付）已落地（2026-09-18）：users 增 `customer_id/is_primary/system_generated`（migration 00014）；`CreateCustomer` 同交易建兩帳號＋24h 臨時密碼＋`must_change_password`；`CreateCustomerResponse` 回傳交付欄位與 `account_manage_url`；`default_sales_rep_id` 改必填。執行計畫見 `2026-09-18-backend-04-customers-d22-plan.md`。
> 🟡 3.1.5（偏好欄位）部分：儲存/預設（D26 全 false ×6、D24 陣列）已於核心批落地；「長度==6 拒絕」與 `promo_tag_ids` 對同部門 `promo_tags` 交叉驗證待補（依賴 07-notifications 的 promo_tags 表）。

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

*最後更新：2026-09-22（3.5/3.6/3.8 落地後對齊；殘項 3.1.5 促銷連動、Logo 上傳 Web 頁與消費面）*
