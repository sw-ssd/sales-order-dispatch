# Backend Phase 1–5 細部實作計畫 — 總覽

> 版本:v1.0.0(2026-08-17)
> 依據:執行計畫 `docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md`(v2.9.0)、規格書 v1.0.34、決策記錄 D1–D28。
> 定位:本目錄為原計畫 **backend 部分(Task 導向)的細部分解**,將每個 Task 拆成可獨立驗收的子功能,並補上實作邏輯與錯誤處理。**不取代、不修改原計畫**;原計畫仍為進度勾選基準,本目錄為實作時的邏輯依據。
> 範圍:Phase 1–5 的 backend 工作。Phase 0(基礎建設)、Phase 6(App)、Phase 7(公告/UI)、Phase 8(部署)不在本目錄;混合 Task(含前端/App)僅拆後端部分並於文中註明。

---

## 1. 文件地圖

| 文件 | 涵蓋 Task | 主題 |
|---|---|---|
| `01-auth.md` | 1.1–1.8、1.11 | 認證(Session/JWT/OAuth2/客戶帳密)、Casbin、RLS、developer 逃生門 |
| `02-tenancy-users.md` | 2.1–2.4、2.9、2.10 | 公司/部門/使用者、角色權限、Casbin policy 管理 |
| `03-metadicts-audit.md` | 2.5、2.6 | 字典檔、稽核日誌 |
| `04-master-data.md` | 3.1–3.6、3.8 | 客戶/商品/倉別/車次/分切規格/分類/專屬商品/檔案/QR |
| `05-sales-orders.md` | 4.1、4.2 | 訂單、狀態機、取號、下單邏輯 |
| `06-returns.md` | 4.7 | 退貨申請與審核(僅後端) |
| `07-notifications.md` | 4.3、4.4 | 通知範本、FCM/站內發送、裝置管理 |
| `08-dispatch.md` | 5.1、5.2 | 派車、Connect 串流看板(僅後端) |
| `09-printing.md` | 5.3–5.5 | 四種單據模板、Gotenberg、列印記錄 |

原計畫各 Phase 驗收 Task(1.12、2.12、3.9、4.8、5.7)不拆,驗收時回到原計畫勾選。

## 2. 子功能編號與模板

- 編號:`Task號.序號`(如 `1.5.1`),可雙向追溯原計畫 Task。
- 每個子功能固定六欄:**目標 / 檔案 / 介面 / 實作邏輯 / 錯誤處理 / 驗收**。
- 一個子功能 = 一個可獨立驗收的行為單元(一條 RPC、一個事務流程、一個機制)。

## 3. 共通規則(各文件只引用、不重複)

1. **交易與稽核**:取號 + 建檔、狀態異動 + 事件軌跡、業務操作 + audit log,皆同一 DB 交易,同成功同失敗(D18)。子功能「實作邏輯」欄明確標出交易邊界。
2. **軟刪除**:業務實體統一 `deleted_at` + 部分唯一索引(D10);查詢預設排除;復原 = 清欄位 + 寫稽核。特殊規則(如 `customer_products` qty=0 保留)於各子功能註明。
3. **多租戶**:每筆業務資料帶 `company_id` / `department_id`;Casbin 管功能(domain = company_id)、RLS 管資料範圍(`data_scope` 等級)、CASL 管 UI(D3)。RLS 注入(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)為最後防線。
4. **錯誤處理約定**:統一以 Connect code 表述 —
   - `unauthenticated`:登入失效、token_version 不符
   - `permission_denied`:角色/範圍不符、主帳號呼叫業務 API
   - `not_found`:資源不存在或已被軟刪除
   - `failed_precondition`:狀態機不允許、樂觀鎖衝突(前端重查)、臨時密碼過期、帳號鎖定
   - `invalid_argument`:輸入驗證失敗
   - `already_exists`:唯一約束衝突(前綴、別名、identifier)
5. **測試約定**:每個子功能「驗收」欄對應可執行測試行為;取號併發、狀態機轉移、RLS 隔離、退貨審核、送貨日順延、鎖定解除六類需整合測試(D21);三端覆蓋率 70% CI 強制。各文件結尾附「整合測試重點」。
6. **依賴標註**:子功能間先後依賴於「目標」欄標註 `相依: X.Y.Z`;全域依賴見下節。

## 4. 全域依賴順序

```mermaid
flowchart LR
    A[01-auth<br/>1.1-1.8, 1.11] --> B[02-tenancy-users<br/>2.1-2.4, 2.9-2.10]
    A --> C[03-metadicts-audit<br/>2.5-2.6]
    B --> D[04-master-data<br/>3.1-3.6, 3.8]
    C --> D
    D --> E[05-sales-orders<br/>4.1-4.2]
    C --> F[07-notifications<br/>4.3-4.4]
    E --> G[06-returns<br/>4.7]
    F --> G
    E --> H[08-dispatch<br/>5.1-5.2]
    H --> I[09-printing<br/>5.3-5.5]
```

關鍵跨檔依賴:
- `4.2.x`(下單邏輯)相依 `3.3.3`(單位換算)、`3.5.3`(別名建立)、`3.1.5`(偏好送貨日)。
- `5.1.2`(批次 Confirm)相依 `4.1.3`(狀態機);`5.1.4`(派車通知)相依 `4.4`。
- `4.7.5`(退貨推播稽核)相依 `4.3`/`4.4`、`2.6`。
- `5.4.3`(PDF 關聯)相依 `3.6`(檔案資產)。
- `2.9.4` 接管 `1.8.1`(GetAbility 由 `role_permissions` 驅動)。

## 5. 使用方式

1. 實作某 Task 前,先讀本目錄對應文件的該 Task 區段。
2. 依子功能順序實作;每完成一個子功能,其「驗收」欄即為該單元的 done 定義。
3. 原計畫 Task 全部子功能完成且整合測試通過後,回原計畫勾選該 Task。
4. 需求變更流程依 `docs/PLANNING_OVERVIEW.md` §7:先升規格書版本,再同步本目錄對應子功能。

---

*最後更新:2026-08-17*
