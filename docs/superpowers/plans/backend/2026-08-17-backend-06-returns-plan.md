# Backend 06 — 退貨申請與審核 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-22 實作對齊。後端已落地：兩表 schema（00039）＋`ReturnService` 5 RPC（Create／List／Get／Review／GetCertificate）＋00040 RLS ENABLE＋FORCE（含 self→customer_id 分支探針）；審核觸發已接 07（同交易建 pending、提交後發送；FCM 發送仍為 Fake）；Web 退貨頁 `549933c` 已落地；App 畫面待。以下保留目標架構供追溯。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D16/D18/D23/D25`
> **細部文件**：`docs/superpowers/plans/backend/detail/06-returns.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：01-auth（RLS、audit.Recorder）、04（客戶/專屬商品）、05（sales_orders 實體）、07（通知契約）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | return_requests / return_request_items schema 與 RLS（細部 4.7.1） | ✅ 完成（00039＋00040，白名單 28） |
| 2 | 發起退貨 Create 與客戶自查 List/Get（細部 4.7.2） | ✅ 完成（後端；Web `549933c`；App 畫面待） |
| 3 | 審核 API + 審核稽核與推播掛點（細部 4.7.3、4.7.5） | ✅ 完成（後端＋稽核；推播觸發已接 07） |
| 4 | 退貨證明資料輸出（細部 4.7.4） | ✅ 完成（後端快照證明；App 畫面待） |

**實作範圍**：後端 100%（推播觸發已接 07、FCM 實裝另案）＋Web 退貨頁 `549933c`（清單/明細審核/證明；`return_request` 權限資源、`Get` 回填 `version`）；App 畫面待。

---

## Global Constraints（目標，精煉自原檔）

- `return_requests` / `return_request_items` 軟刪除（D10）；客戶子帳號雙來源（歷史訂單品項 / 專屬商品）發起（細部 4.7.2）。
- 業務審核 approved/rejected、樂觀鎖、**不修改原訂單**（D25）。
- 寫入路徑單一 DB 交易 + `audit.Recorder` 同事務稽核（D18）；審核結果推播僅通知發起帳號（D23），採 07 契約（pending 同交易、提交後發送、失敗僅標 failed 不回滾 D16）。
- 系統僅記錄申請與審核狀態，無配送取貨 API（D25）。

## 待辦 Task 詳情

### Task 1: return_requests / return_request_items schema 與 RLS（細部 4.7.1）
退貨申請 + 申請明細 schema、`customer_id` 收口的 RLS（客戶帳號 `data_scope = self`）。

### Task 2: 發起退貨 Create 與客戶自查 List/Get（細部 4.7.2）
`ReturnService.Create`（歷史訂單勾選 / 專屬商品並存）＋ 客戶自查 `List`/`Get`。

### Task 3: 審核 API + 稽核與推播掛點（細部 4.7.3、4.7.5）
業務審核（approved/rejected、樂觀鎖、不修改原訂單）；審核結果同事務稽核 + 推播掛點（07 契約）。

### Task 4: 退貨證明資料輸出（細部 4.7.4）
`GetCertificate` 退貨證明資料輸出（對接 09 列印 / 檔案）。

---

*最後更新：2026-09-22（06-returns 後端＋Web 退貨頁 `549933c` 落地對齊；FCM 實裝另案、App 畫面待）*
