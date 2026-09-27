# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

> 本檔是整個產品的共用真相，四個 surface（`frontend/`、`platform-console/`、`app/`、`backend/`）共用；`backend/` 沒有獨立檔案，繼承本檔。各 surface 另有自己的 PRODUCT.md 記錄該端特有的使用者情境與工作流；產品整體敘事以本檔為準，不重複分歧。
>
> 本產品的 platform 橫跨 web 與原生，欄位無單一值可表達：本檔記 `web`（兩個 web surface 的基準），原生端準則以 `app/PRODUCT.md` 的 `adaptive` 為準；在 `app/` 下工作時以該檔為準。

## Users

- **平台營運人員（operator）**：產品方自家的營運／支援人員，以自有白名單（`platform.operators`）＋ OIDC 登入獨立的營運主控台。**不得、也不需要**持有任一租戶公司的 `users` 列。未來可能開放給特定客戶的專責人員協助管理自己的租戶。
- **總管理員（super）**：單一公司內跨部門的最高權限，同時是平台的支援逃生門（操作落租戶稽核且須填原因）。
- **公司管理員（company_admin）**：限自己公司；公司主檔、品牌、部門、人員與角色權限。
- **部門管理員（dept_admin）**：限自己部門；部門主檔（客戶、商品、倉別、車次、分切規格、商品分類、字典）、派車與單據列印。
- **一般職員（staff，兼會計）**：日常開單、訂單異動、查詢與列印。
- **業務**：以 App 在店家現場開單、維護客戶專屬商品清單、審核退貨；在 Web 檢視自己部門的訂單與派車狀態。
- **客戶店家（customer）**：只在 App 作業。一主多子帳號：主帳號僅能自助管理子帳號（不提供業務登入），子帳號是唯一的店家業務身分；自動附帶的業務子帳號專供該客戶的主責業務登入代客操作，店家端灰化不可管理。
- **guest**：員工首次 OIDC 登入後待審核（super / company_admin 審），審核完成前不建帳號。
- **developer**：第 7 個內建角色，逃生門；production 預設關閉，誤開 fail-fast 拒絕啟動。

## Product Purpose

多公司訂出貨系統：批發業的每個部門各自擁有客戶、商品與庫別主檔，業務在店家現場開單，中台完成派車與四種紙本單據列印，並以推送（FCM）＋站內通知串連業務、店家與管理端。

商業形態是 **多租戶 SaaS**：共用部署、共用 DB，一列公司＝一租戶；方案階梯（功能開關）× 席位 × 資料筆數上限決定可用範圍；v1 為人工收款（匯款／月結）＋營運後台控管到期，金流與電子發票自動化是後續 adapter。

存在的理由：取代客戶既有的三倉系統（`sales-order-backend` / `-frontend` / `-app`）。1.0 是全新 monorepo 從頭開發、**不遷移舊資料**、以 Big Bang 完成既有客戶的切換（測試完成後全面切換，無試點、無並行、無唯讀期）；切換後平台以 SaaS 形態持續服務多個租戶公司。舊系統的程式碼僅供背景參考，不沿用。

成功定義：平台能持續服務多個租戶公司並依方案收費（訂閱、席位、筆數上限可被營運端控管），而各租戶的日常作業（開單 → 派車 → 揀貨 → 列印 → 送達）在無舊系統依賴下運轉，且彼此資料互不干擾。

## Positioning

部門自治主檔 + 雙層租戶防護：功能權限由 OpenFGA 管（資源級 userset），資料範圍由 PostgreSQL RLS 管（`data_scope` = all / company / department / self），兩者都不依角色名稱判斷。

相鄰產品做不到的機制：同一商品可在不同部門有不同名稱，又可在各部門的不同客戶各自改名綁定，且**不需要跨部門大總表**。另：**業務資料域全系統不儲存金額**。

## Operating Context

- 派車看板是 Kanban 拖放 + 車次批次確認；拖放比對 `version` 樂觀鎖，`WatchBoard`（Connect server streaming）只觸發全量重查，跨 replica 走 Valkey pub/sub，串流不可用時降級為輪詢。
- 四種單據是實際作業文件：單車總表、車次對點單、揀貨單（依車次 → 倉別 → 分類 → 品名排序）、加工單（加工室揀／配送揀兩區塊；加工後數量列印空白、手寫回填，不回寫）。重印開放 dept_admin / staff，必填原因。
- 稽核日誌與業務操作同一 DB 交易寫入（同成功同失敗）；保留期由管理頁設定 1 / 3 / 6 / 12 個月或永久。平台域另有自己的稽核（actor 為 operator，非租戶 `users`）。
- 通知兩通道（FCM + 站內，**無 Email**）；促銷推播（分類標籤選群）與公告分離。
- 平台維運一律走獨立的營運主控台（`platform-console/`，只走 `platform/v1`），不經租戶身分；租戶 SPA 不掛任何 `platform.*` 能力、也沒有平台路由。營運主控台與租戶 Web 共用同一套 UI 元件庫與設計 token。
- 訂閱與權益是營運面與租戶面共用的判準：權益決策單點、fail-closed；租戶端看到的是自己公司權益的投影，且「尚未訂閱」必須與「方案不含」可辨識。
- 介面語言為繁體中文（`zh-Hant`）；文件、註解、commit message 一律繁體中文；多語 UI 不在範圍內。
- 既有 UI 設計稿在 Pixso（`多公司訂出貨系統` UI 稿；來源存檔 `docs/design/2026-09-18-pixso-版面美化-進度存檔.md`）。

## Capabilities and Constraints

- 技術組成：pnpm + Turborepo monorepo；`backend/` Go + Ent + Connect-RPC + PostgreSQL + Valkey + Gotenberg；`frontend/` 與 `platform-console/` SolidJS + Vite（Ark UI × Tailkit）；`app/` Flutter（fvm、雙 flavor、`--dart-define` 注入組態）。`backend/proto/` 是 API 唯一來源，產生三端型別。
- API：Connect-RPC 為唯一來源；REST 只保留公開端點（版本、OIDC、QR 兌換、檔案、公司公開資訊）。
- 認證：**OIDC（IdP 不限 Google）與帳密兩條路並存，員工與店家皆可使用**；店家子帳號另有 QR Code 登入。Web 用 scs + Valkey session cookie，營運主控台用自己的 operator session cookie；App 用 access JWT 1 小時 + refresh 30 天旋轉。停用／強制登出／改密碼／角色變更以 `token_version` 全數失效。
- 角色：`super` / `company_admin` / `dept_admin` / `staff`（兼會計）/ `customer` / `guest` / `developer`；7 個內建角色 + `role_permissions` seed，後台可調、防鎖死。`platform.*` 能力永不進租戶角色矩陣。
- 訂單狀態機：`pending ⇄ processing → completed`、`cancelled`、作廢 `voided`（終態）；所有異動寫 `sales_order_events`。
- **不儲存金額的分界**：業務資料域（訂單、明細、商品、客戶專屬商品）沒有任何金額欄位；平台計費域（方案價目、訂閱、待收款）依定義持有金額。
- 統一軟刪除（`deleted_at` + 部分唯一索引）；編號取號與建檔同一交易（樂觀鎖）。
- 檔案本地儲存，白名單（jpeg / png / webp ≤5MB、PDF ≤10MB）+ magic bytes 雙重檢查；跨域檔案存取走共用的 FileStore，各域不自建上傳／下載路徑。
- 計費與權益：方案階梯（功能開關）× 席位 × 資料筆數上限；**不按量計費**（無用量事件流、無帳單引擎）。席位值語意：0 = 不含、-1 = 不限、>0 = 上限。
- 平台與產品域的跨界副作用（含授權 tuple 同步）在交易提交後執行，不在業務交易內。
- 明確不在範圍：供應商退貨授權、獨立報價單、簽核流程、進銷存庫存、配送專屬頁、獨立會計角色、自建 IdP、Cmd+K 全域搜尋、電子發票開立、App 強制更新、通知重試佇列、上傳病毒掃描、多語 UI。
- 未定案：Noto Sans CJK TC 生產字體授權與安裝方式（上線前必答）。

## Brand Commitments

- 產品名稱「多公司訂出貨系統」。沒有品牌 logo，也沒有對外行銷素材。
- 對客戶的書面用語一律繁體中文；不使用行銷式誇詞，也不自稱未經證實的定位。
- 單據與畫面用語沿用現場慣稱（派車、揀貨、對點、分切），不為了「產品感」改寫作業詞彙。

## Evidence on Hand

- 凍結規格書 `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.35，18 章）：欄位與流程的唯一權威。
- 決策記錄 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D33）。
- 需求規格 13 份 `docs/superpowers/specs/1.0-requirements/`（announcements、audit-compliance、authorization、dispatch、file-assets、identity-access、logistics-execution、master-data、multi-tenancy、notifications、ops-deployment、printing、sales-orders）。
- SaaS 化設計 `docs/superpowers/specs/2026-09-20-saas-billing-entitlements-design.md`（決策 S1–S11）。
- 規劃總覽 `docs/PLANNING_OVERVIEW.md`（範圍 In / Out of Scope 與 D1–D33 摘要）、功能現況對照 `docs/FUNCTION_LIST.md`、錯誤碼表 `docs/error-codes.md`。
- 客戶原始需求 `docs/客戶需求.txt`、`docs/需求備忘_2026-08-03.txt`。
- Pixso UI 稿（15 畫面 = Web 9 + App 6）與 `docs/design/2026-09-18-pixso-版面美化-進度存檔.md`。

不存在、未來工作不得捏造：真實客戶名單與推薦語、營收或用戶數數據、品牌 logo、字體授權憑證、上線時程承諾、對外 SLA。

## Product Principles

1. **部門是主檔的邊界**：客戶、商品、庫別、車次由各部門自建，不做跨部門大總表；跨部門調閱是例外而非常態。
2. **權限與資料範圍分離**：功能（OpenFGA）與資料可見範圍（RLS `data_scope`）各自判斷，不綁角色名稱；客戶端守衛只影響體驗，不構成授權。
3. **現場優先**：業務在店家現場就能完成開單與改名，不要求回辦公室補資料。
4. **紙本是交付介面**：四種單據的欄位、排序與留白是作業需求，不是裝飾。
5. **營運與租戶不混用身分**：平台維運走自有 operator 身分與自有稽核，不動用租戶 `users`；計費與權益分離，權益決策單點且 fail-closed。

## Accessibility & Inclusion

- 全站繁體中文並標記 `zh-Hant`；多語 UI 不在範圍。
- 現場作業在手機上單手操作（App），桌機後台為主要工作環境；兩端都不得以顏色單獨承載狀態。
- 深色模式是既有基線：設計 token 同時提供淺／深兩套，深色由覆蓋同一組語意 token 達成，表單控件在深色下必須可辨識。
