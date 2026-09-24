# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- **部門管理員 / 公司管理員 / 總管理員**（主要）：桌機後台。建立與維護各自部門的主檔（客戶、商品、業務、倉別、車次、分切規格、商品分類、字典），管理部門人員帳號與角色權限，執行派車與單據列印。公司管理員僅限自己公司，總管理員可跨部門調閱全部資料。
- **一般職員（含會計）**：桌機後台。日常開單、訂單異動、查詢與列印。
- **業務**：以 App 在店家現場作業（見 `../app/PRODUCT.md`），在後台檢視自己部門的訂單與派車狀態。
- **客戶店家**：只在 App 出現，後台沒有客戶介面。

## Product Purpose

多公司訂出貨系統 1.0：讓批發業的多個部門各自擁有客戶、商品與庫別主檔，業務在店家現場開單，中台完成派車與四種紙本單據列印，並以 FCM + 站內通知串連業務、店家與管理端。

存在的理由：取代現行三倉系統（`sales-order-backend` / `-frontend` / `-app`）。1.0 是全新 monorepo 從頭開發、**不遷移舊資料**、Big Bang 上線（測試完成後全面切換，無試點、無並行、無唯讀期）。舊系統的程式碼僅供背景參考，不沿用。

成功定義：切換後日常作業（開單 → 派車 → 揀貨 → 列印 → 送達）完全不依賴舊系統，且各部門的主檔與資料互不干擾。

## Positioning

部門自治主檔 + 雙層租戶防護：功能權限由 OpenFGA 管（userset），資料範圍由 PostgreSQL RLS 管（`data_scope` = all / company / department / self），兩者都不依角色名稱判斷。

相鄰產品做不到的機制：同一商品可在不同部門有不同名稱，又可在各部門的不同客戶各自改名綁定，且**不需要跨部門大總表**。另：全系統不儲存金額。

## Operating Context

- 派車看板是 Kanban 拖放 + 車次批次確認；拖放比對 `version` 樂觀鎖，`WatchBoard`（Connect server streaming）只觸發全量重查，跨 replica 走 Valkey pub/sub，串流不可用時降級為 30 秒輪詢。
- 四種單據是實際作業文件：單車總表、車次對點單、揀貨單（依車次 → 倉別 → 分類 → 品名排序）、加工單（加工室揀／配送揀兩區塊；加工後數量列印空白、手寫回填，1.0 不回寫）。重印開放 dept_admin / staff，必填原因。
- 稽核日誌與業務操作同一 DB 交易寫入（同成功同失敗）；保留期由管理頁設定 1 / 3 / 6 / 12 個月或永久。
- 介面語言為繁體中文；文件、註解、commit message 一律繁體中文（多語 UI 不在 1.0 範圍）。
- 既有 UI 設計稿在 Pixso（`多公司訂出貨系統` UI 稿，15 畫面 = Web 9 + App 6；來源存檔 `docs/design/2026-09-18-pixso-版面美化-進度存檔.md`）。程式實作進度落後於稿面：目前只有認證與公司／部門／角色／人員頁面。

## Capabilities and Constraints

- API：Connect-RPC 為唯一來源（proto `v1` 產生三端型別）；REST 只保留公開端點（版本、OAuth2、QR 兌換、檔案、公司公開資訊）。
- 認證：Web 用 scs + Valkey session cookie；停用／強制登出／改密碼／角色變更以 `token_version` 全數失效。員工僅能用 Google Workspace OIDC。
- 角色：`super` / `company_admin` / `dept_admin` / `staff`（兼會計）/ `customer` / `guest` / `developer`；7 個內建角色 + `role_permissions` seed，後台可調、防鎖死。developer 是逃生門，prod 預設關閉。
- 訂單狀態機：`pending ⇄ processing → completed`、`cancelled`、作廢 `voided`（終態）；所有異動寫 `sales_order_events`。
- **全系統不儲存金額**（訂單、明細、商品、客戶專屬商品皆無金額欄位）。
- 統一軟刪除（`deleted_at` + 部分唯一索引）；編號取號與建檔同一交易（樂觀鎖）。
- 檔案本地儲存，白名單（jpeg / png / webp ≤5MB、PDF ≤10MB）+ magic bytes 雙重檢查。
- 明確不在 1.0 範圍：供應商退貨授權、獨立報價單、簽核流程、進銷存庫存、配送專屬頁、獨立會計角色、自建 IdP、Cmd+K 全域搜尋、電子發票開立、App 強制更新、通知重試佇列、上傳病毒掃描、多語 UI。
- 待答：Noto Sans CJK TC 生產字體授權與安裝方式（上線前必答）。

## Evidence on Hand

- 凍結規格書 `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34，18 章）：欄位與流程的唯一權威。
- 決策記錄 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D33）。
- 需求規格 12 份 `docs/superpowers/specs/1.0-requirements/`。
- 客戶原始需求 `docs/客戶需求.txt`、`docs/需求備忘_2026-08-03.txt`。
- Pixso UI 稿與美化進度存檔 `docs/design/2026-09-18-pixso-版面美化-進度存檔.md`。

不存在、未來工作不得捏造：真實客戶名單與推薦語、金額或營收資料、品牌 logo、字體授權憑證、上線時程承諾。

## Product Principles

1. **部門是主檔的邊界**：客戶、商品、庫別、車次由各部門自建，不做跨部門大總表；總管理員的跨部門調閱是例外而非常態。
2. **權限與資料範圍分離**：功能（OpenFGA）與資料可見範圍（RLS `data_scope`）各自判斷，不綁角色名稱。
3. **現場優先**：業務在店家現場就能完成開單與改名，不要求回辦公室補資料。
4. **紙本是交付介面**：四種單據的欄位、排序與留白是作業需求，不是裝飾。
5. **不持有金額**：系統刻意不儲存價格，避免與既有帳務系統重疊。
