# Backend 08 — 派車與串流看板 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-22 實作對齊。後端已落地：`DispatchService` 4 RPC（Assign／Confirm／Cancel／WatchBoard）＋派車通知（07 觸發）＋提交後發佈（程序內直投；跨 replica Valkey 訂閱層另案）；Web 看板頁＋輪詢降級待。以下保留目標架構供追溯。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D13/D14`
> **細部文件**：`docs/superpowers/plans/backend/detail/08-dispatch.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：01-auth（auth interceptor、RLS）、05-sales-orders（訂單狀態機，同套件）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | AssignRoute 樂觀鎖與順位重排（細部 5.1.1） | ✅ 完成（後端；Web 拖放頁待） |
| 2 | 車次批次 Confirm（細部 5.1.2） | ✅ 完成（後端逐筆交易＋部分失敗） |
| 3 | CancelDispatch 與派車通知觸發（細部 5.1.3–5.1.4） | ✅ 完成（後端＋重印警告＋通知） |
| 4 | WatchBoard proto 與串流 handler（細部 5.2.1） | ✅ 完成（後端串流＋部門隔離） |
| 5 | Valkey pub/sub 跨 replica 與 heartbeat（細部 5.2.2–5.2.3） | 🟡 部分（heartbeat 已落地；跨 replica 訂閱層另案，發佈器介面已預留） |

**實作範圍**：後端約 90%（跨 replica 轉發＋Web 頁待）。

---

## Global Constraints（目標，精煉自原檔）

- `DispatchService` 四個 RPC 落於 `internal/domain/salesorders`（與 05-plan 同套件，共用訂單狀態機）；三個 mutation 的 DB 交易**提交成功後**才發佈 `BoardEvent`（D14）。
- 事件經 Valkey 以部門分 channel 廣播，各 replica 僅訂閱本機有連線的部門；heartbeat 純連線維持，不寫事件、不進 Valkey。
- 認證走 01-plan auth interceptor，不實作一次性 ticket（D14，WS/ws_ticket 已廢除）；Connect server streaming + 30 秒輪詢降級。

## 待辦 Task 詳情

### Task 1: AssignRoute 樂觀鎖與順位重排（細部 5.1.1）
看板拖放派車：樂觀鎖（比對 version）+ 車次內順位重排。

### Task 2: 車次批次 Confirm（細部 5.1.2）
車次批次確認（逐筆交易、部分失敗語義）。

### Task 3: CancelDispatch 與派車通知觸發（細部 5.1.3–5.1.4）
取消派車（dept_admin + 必填原因 + 重印警告）；派車通知觸發（介面注入 `DispatchNotifier`，07 消費）。

### Task 4: WatchBoard proto 與串流 handler（細部 5.2.1）
`WatchBoard` server streaming（Connect，事件僅觸發 invalidate 全量重查）。

### Task 5: Valkey pub/sub 跨 replica 與 heartbeat（細部 5.2.2–5.2.3）
Valkey pub/sub 跨 replica 轉發 + 25 秒 heartbeat；降級 30 秒輪詢。

---

*最後更新：2026-09-22（08-dispatch 後端對齊；跨 replica＋Web 頁待）*
