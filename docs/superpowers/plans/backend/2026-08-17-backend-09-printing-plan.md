# Backend 09 — 單據列印 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-22 實作完成。Tasks 1–4 全數落地（commits ead74f0／7445e92／2ed66f1／ae0d610）：`ent/schema` 有 printlog／printpreview 實體、`internal/print`（view model＋模板＋Gotenberg client＋PDF 產線）、Preview／Print／ListLogs RPC、00037 schema＋00038 RLS ENABLE＋FORCE。以下保留目標架構供追溯。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D15`
> **細部文件**：`docs/superpowers/plans/backend/detail/09-printing.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：04-master-data（FileStore / file_assets，Task 8）、05（訂單/揀貨資料）、08（派車/車次）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | 四種單據 view model 與模板（細部 5.3.1–5.3.4） | ✅ 完成（ead74f0） |
| 2 | Gotenberg client + 資料組合 + PDF 產線（細部 5.4.1–5.4.3） | ✅ 完成（7445e92） |
| 3 | print_logs / print_previews schema + Preview API（細部 5.5.1–5.5.2） | ✅ 完成（2ed66f1，含 00037／AppRole 白名單 26） |
| 4 | Print API + 重印 + ListLogs（細部 5.5.3–5.5.4） | ✅ 完成（ae0d610，含 00038 RLS ENABLE＋FORCE） |

**實作範圍**：100%（全數完成；`task test:integration -count=1` 全套 ok=30 fail=0）。

---

## Global Constraints（目標，精煉自原檔）

- 四種 A4 單據：單車總表 / 對點單 / 揀貨單 / 加工單，**全文無金額**（D12/D15）。單車總表不顯示金額、對點單不顯示價格、揀貨單依車次→倉別→分類→品名、加工單分加工室揀/配送揀 + 手寫回填。
- 渲染管線三層：`internal/print`（view model + 模板 + Gotenberg client + PDF 產線，純內部套件）→ `internal/domain/prints`（Connect-RPC handler，交易邊界與稽核）→ `fileassets` 埠（FileStore 介面，實作由 04 Task 8 提供）。
- Gotenberg 以 `Converter` 介面抽象（fake / httptest 測試）；Gotenberg 8（Chromium）。
- 正式列印整批檢查 `status = processing`；重印必填原因且 `is_reprint = true`；PDF 經 file_assets 下載 URL 交付；Noto Sans CJK TC 字體。

## 待辦 Task 詳情

### Task 1: 四種單據 view model 與模板（細部 5.3.1–5.3.4）
單車總表 / 對點單 / 揀貨單 / 加工單 view model + `html/template`（embed）。

### Task 2: Gotenberg client + 資料組合 + PDF 產線（細部 5.4.1–5.4.3）
Gotenberg HTML→PDF client（有限重試）；資料組合與 PDF 產線（空表不產生、失敗補償刪除）。

### Task 3: print_logs / print_previews schema + Preview API（細部 5.5.1–5.5.2）
列印日誌/預覽 schema + `Preview` RPC。

### Task 4: Print API + 重印 + ListLogs（細部 5.5.3–5.5.4）
`Print` RPC（整批檢查 processing）、`ListLogs`；重印必填原因。

---

*最後更新：2026-09-22（09-printing 實作完成對齊；全整合綠 ok=30 fail=0）*
