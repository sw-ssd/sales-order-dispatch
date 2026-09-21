# Backend 05 — 銷售訂單 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-22 實作對齊。後端已落地：訂單三表 schema（00031）＋order_counters 樂觀鎖取號＋狀態機 `TransitionOrder`＋`SalesOrderService` 9 RPC（List/Get/Create/Update/Cancel/Complete/Void/Delete＋ListOrderEvents）＋00032 RLS ENABLE＋FORCE；Web/App 業務頁待。以下保留目標架構供追溯。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D7/D12/D13`
> **細部文件**：`docs/superpowers/plans/backend/detail/05-sales-orders.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：04-master-data（客戶/商品契約）、01-auth（RLS、狀態機審計）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | 訂單三表 schema（細部 4.1.1） | ✅ 完成（00031） |
| 2 | order_counters 樂觀鎖取號（細部 4.1.2，D7） | ✅ 完成（同交易取號） |
| 3 | 訂單狀態機（細部 4.1.3，D13） | ✅ 完成（`TransitionOrder` 集中模組） |
| 4 | SalesOrderService CRUD 與 ListEvents（細部 4.1.4–4.1.5） | ✅ 完成（9 RPC 後端；Web/App 頁待） |
| 5 | 下單組裝邏輯（細部 4.2.1–4.2.5） | ✅ 完成（後端組裝＋客戶專屬清單守衛） |

**實作範圍**：後端 100%（Web/App 業務頁待）。

---

## Global Constraints（目標，精煉自原檔）

- 訂單三表（sales_orders / sales_order_items / sales_order_events）＋ 狀態機集中單一模組，所有轉移走 `Transition()`（條件更新 + 事件 + 必要時稽核，同一交易）。狀態：`pending ⇄ processing → completed`、`cancelled`、終態 `voided`（D13）。
- 取號與建單同一 DB 交易（D7）；`order_no` = 來源碼 + 6 位自增。
- **不存任何金額欄位**（D12，含明細/商品）。
- 數量/換算用 `shopspring/decimal`（對齊 04 契約）；手打商品存別名（綁定客戶）、客戶專屬清單守衛、來源記錄、偏好送貨日順延。

## 待辦 Task 詳情

### Task 1: 訂單三表 schema（細部 4.1.1）
`sales_orders` / `sales_order_items` / `sales_order_events` schema 與 RLS。

### Task 2: order_counters 樂觀鎖取號（細部 4.1.2，D7）
訂單編號樂觀鎖取號（source prefix + 6 位自增，取號與建單同交易）。

### Task 3: 訂單狀態機（細部 4.1.3，D13）
集中狀態機 + `Transition()`；所有轉移寫事件軌跡 `sales_order_events`；取消派車保留看板位置。

### Task 4: SalesOrderService CRUD 與 ListEvents（細部 4.1.4–4.1.5）
訂單 CRUD（軟刪除/復原）、事件軌跡查詢、分頁。

### Task 5: 下單組裝邏輯（細部 4.2.1–4.2.5）
單位換算、手打別名、客戶專屬清單守衛、來源記錄、偏好送貨日非勾選日自動順延（D26）。

---

*最後更新：2026-09-22（05-sales-orders 後端落地對齊；Web/App 頁待）*
