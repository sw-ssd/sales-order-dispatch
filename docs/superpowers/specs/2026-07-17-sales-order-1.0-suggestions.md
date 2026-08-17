# 多公司訂出貨系統 1.0 — 建議備忘錄

> 日期：2026-07-17  
> 對應規格書：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md` v1.0.24  
> 性質：1.0 已定稿範圍外的建議事項備忘，供後續版本（1.1+）與實作期間評估；**非 1.0 承諾範圍**。

---

## 1. App 版本檢查與強制更新

- **背景**：App 上架（App Store / Google Play）後，舊版本可能持續被使用；API 若有不相容變更，無強制更新機制將導致舊版 App 錯誤或資料異常。
- **建議**：
  - 後端提供版本檢查方法（如 `VersionService.Check`，回傳最低支援版本與最新版本）。
  - App 啟動時檢查：低於最低支援版本 → 強制導向商店更新；低於最新版本 → 提示更新（可略過）。
  - Connect-RPC 以 proto package `v1` 演進，不相容變更另開 `v2`。
- **建議時機**：1.1（首次上架後即需要）。

## 2. 電子發票（統一發票開立）

- **背景**：`customers.invoice_type_id` 已在資料模型保留欄位，但 1.0 無開立發票功能；台灣 B2B 訂單幾乎必配發票，規格第 1.2 節未提及，等於默認不做。
- **建議**：
  - 明確評估 1.1 是否納入電子發票整合（加值中心 API，或人工開立後回填發票號碼）。
  - 若不整合，至少於訂單提供「發票號碼回填」欄位，供人工對帳。
  - 與「會計檢視 / 對帳」（`staff` 兼任）一併設計。
- **建議時機**：1.1 規劃時與業主確認。

## 3. 複合索引（實作細節）

- **背景**：主要查詢模式需複合索引支撐，避免全表掃描。
- **建議索引**（Phase 3–5 schema 建立時一併處理，不入規格）：

  | 資料表 | 索引 | 支撐查詢 |
  |---|---|---|
  | `sales_orders` | `(department_id, expected_delivery_date)` | 派車看板日期篩選 |
  | `sales_orders` | `(route_id, delivery_sequence)` | 車次內排序 |
  | `sales_orders` | `(customer_id, deleted_at)` | 客戶訂單歷史 |
  | `customer_products` | `(customer_id, deleted_at)` | 專屬商品清單帶出 |
  | `notifications` | `(user_id, status, created_at)` | 通知中心 |
  | `audit_logs` | `(company_id, created_at)` | 稽核查詢與保留排程 |

- **建議時機**：各 Phase Ent schema 建立時。

## 4. audit_logs 時間分區

- **背景**：規格 10.2 已註記「便於日後依時間分區或封存」；1.0 資料量小不需分區。
- **建議**：
  - 1.0 以單表 + `(company_id, created_at)` 索引即可。
  - 單月寫入超過百萬列或查詢明顯變慢時，改 PostgreSQL 宣告式分區（按月 RANGE 分區 `created_at`），保留排程以 drop partition 取代 DELETE。
- **建議時機**：上線後觀察 3–6 個月資料量再決定。

---

## 已於規格內記錄的延後項目（不重複收錄）

| 項目 | 規格位置 | 狀態 |
|---|---|---|
| 自建 IdP（Authelia/Authentik） | 4.1 | 延後至後續版本 |
| MCP / ACP / A2A AI 整合 | 第 17 章 | 1.0 僅預留欄位與公開端點 |
| Cmd+K 全域搜尋 | 8.3 | 第一版不實作 |
| 通知失敗重試佇列 | 10.1 | 1.0 不重試 |
| 上傳檔案病毒掃描 | 14.3 | 1.0 不實作 |
| 多語 UI | 10.1（模板 `locale` 欄位已留） | 1.0 繁中 |

---

*文件版本：v1.0.0*  
*日期：2026-07-17*
