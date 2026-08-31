# 多公司訂出貨系統 1.0 — 決策記錄（D1–D28）

> 來源：原 `openspec/changes/sales-order-1-0/design.md`（OpenSpec 工作流已於 2026-08-03 停用，本檔為遷移後的**決策層 single source of truth**）。
> 細節權威：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34，18 章完整設計規格，為細節欄位與流程的唯一權威參考）；執行計畫：`docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md`（v2.9.0）。
> 本檔為決策層；資料模型逐欄位定義、頁面清單、單據排版等細節以客戶版規格書為準，本檔不重複抄錄。

## Context

現行訂出貨系統（`sales-order-backend` / `sales-order-frontend` / `sales-order-app` 三倉）高度依賴外部系統提供客戶、商品、業務、報價等主檔，無法自建管理；亦缺乏多公司（集團法人）與多部門的權限隔離，倉儲、派車與單據列印流程需要強化。

1.0 以**全新 monorepo 從頭開發**，沿用既有技術棧（Go 1.25 / SolidJS 1.9 / Flutter 3.35.2），導入 Company → Department 兩層租戶、自建全部主檔、派車看板與四種單據列印，並以 Big Bang 方式上線（無資料遷移、無並行期）。

主要關係人（角色）：`super`（集團管理）、`company_admin`、`dept_admin`、`staff`（兼會計檢視）、`customer`（外部店家）、`guest`（待審核）、`developer`（開發逃生門，環境開關控制）。

約束：
- API 協議：Connect-RPC 為業務 API 唯一來源（proto package `v1`），REST 僅保留公開端點。
- 部署：Kubernetes + Traefik，PostgreSQL / Valkey 自建 StatefulSet（非雲端託管）。
- 資料：不從舊系統匯入任何資料，主檔與訂單全部重新建檔。
- 1.0 僅繁體中文 UI；金額 TWD 整數；顯示時區 UTC+8。

## Goals / Non-Goals

**Goals:**

1. 完全取代外部系統依賴：客戶、商品、業務、報價（以客戶專屬商品清單取代）等主檔全部自建。
2. 多租戶：Company → Department 兩層架構，Casbin RBAC with domain（後端）+ PostgreSQL RLS（資料庫）+ CASL.js（前端）三重防護。
3. 強化營運流程：派車 Kanban 看板（Connect 串流即時推播）、四種單據列印（單車總表/對點單/揀貨單/加工單）、倉儲分切規格。
4. 三端一致：Go 後端 + SolidJS 中台 + Flutter App，proto 為唯一 API 來源，三端型別皆由 proto 產生。
5. 安全合規：稽核日誌同步寫入、資料保留期限、傳輸/儲存加密、備份與災難復原（RTO 4h / RPO 1h）。
6. Big Bang 上線：測試完成後全面切換。

**Non-Goals（1.0 不做）:**

- 資料遷移（不從舊系統匯入任何資料；migration script 另行評估）。
- 漸進式上線（無試點、無並行、無舊系統唯讀期）。
- 供應商退貨授權、獨立報價單（Estimate）、簽核流程。
- 進銷存庫存管理（僅記錄「庫別」「分庫」屬性，不追蹤即時庫存）。
- 獨立會計角色（`staff` 兼任）、自建 IdP（Authelia/Authentik 延後）。
- MCP / ACP / A2A 完整協定（僅預留 `capabilities` / `public_info` 欄位與公開發現端點）。
- 電子發票開立（僅保留 `customers.invoice_type_id` 欄位）、App 強制更新、通知失敗重試佇列、上傳病毒掃描、多語 UI、Cmd+K 全域搜尋。
- 1.1 AI 輔助功能（拍照建客戶、業務語音下單）為獨立迭代，見 `docs/superpowers/specs/2026-07-18-app-ai-assist-1.1-design.md`。

## Decisions

### D1：全新 monorepo 從頭開發，不遷移任何資料

- **選擇**：新倉庫（`backend/` + `frontend/` + `app/` + `infra/`，pnpm workspace + Turborepo + Task），主檔與訂單全部人工重建檔。
- **理由**：舊系統與外部系統主檔糾纏，資料品質與編碼規則（新 customer_code / order_no）不相容；重新建檔反而加快上線並讓新資料模型不受舊包袱限制。
- **已考慮 alternative**：寫 migration script 匯入舊主檔 — 拒絕，舊編碼規則不相容且匯入品質風險高；規格第 13 章明定另行評估、不納 1.0。

### D2：Big Bang 上線

- **選擇**：測試完成後直接全面上線並廢止舊系統，無試點、無並行、無唯讀期。
- **理由**：業主決策；多公司主檔雙寫同步成本遠高於一次性切換。
- **已考慮 alternative**：試點公司先行 / 新舊並行 — 拒絕，需維護雙系統同步與對帳，複雜度不可接受。
- **代價**：上線前的主檔人工建檔成為關鍵路徑（見 Risks R2）。

### D3：租戶隔離三層分工 — Casbin 管功能、RLS 管資料範圍、CASL 管 UI

- **選擇**：Casbin RBAC with domain（domain = `company_id`）只判「角色 × 公司 × 資源 × 動作」；部門級資料範圍不由 Casbin 承載，統一由 PostgreSQL RLS（`app.current_company_id` / `app.current_department_id` / `app.current_data_scope` / `app.current_customer_id`）+ repository 查詢條件實現；資料範圍依角色 `data_scope` 等級（`all` / `company` / `department` / `self`）決定，**不依角色名稱**。前端 CASL ability 由 `role_permissions` 表動態產生，登入後載入一次 + TTL 60 秒快取。
- **理由**：Casbin policy 若承載部門維度會爆炸（policy 數 = 角色 × 部門 × 資源）；RLS 是資料庫最後防線，即使 usecase 漏寫條件也不越界；`data_scope` 等級化讓自訂角色不必改程式即可獲得正確資料範圍。
- **已考慮 alternative**：純應用層 where 條件 — 拒絕，漏一處即越權；Casbin 承載部門 — 拒絕，policy 維護不可行。

### D4：Connect-RPC 為業務 API 唯一來源，REST 僅公開端點

- **選擇**：Web / App 所有業務 API 走 Connect-RPC（proto `v1`，每 domain 一個 Service），走 HTTP/2 443 由 Traefik 依 path 轉發；REST 僅保留：版本查詢、OAuth2 導向與 callback、QR token 兌換、檔案上傳/下載、`GET /api/v1/companies/public/{identifier}`。三端型別皆由 proto 產生（connect-go / connect-es / connect-dart）。
- **理由**：單一契約來源消除三端型別漂移；Connect 同時支援 gRPC 與 JSON-over-HTTP，除錯友善。
- **已考慮 alternative**：沿用 REST + Swagger + tygo（舊系統做法）— 拒絕，型別同步是單向產物且 Swagger 維護成本高；全 gRPC — 拒絕，瀏覽器需代理層。

### D5：認證雙軌 — Web session cookie / App JWT + refresh 旋轉，統一 `token_version` 撤銷

- **選擇**：Web 用 scs + Valkey session（httpOnly cookie + CSRF token），不發 JWT；App 用 access JWT（1 小時，`tv` claim）+ refresh token（30 天旋轉制，後端僅存雜湊）。`users.token_version` 於停用帳號、強制登出、密碼重置、角色變更時 +1，驗證 middleware 每次比對 `tv`，舊 token 立即失效。`X-Api-Token` 僅供 server-to-server，不配置於客戶端。
- **理由**：Web 用 cookie 避免 XSS 偷 token；App 無 cookie 生態用 JWT 必需可撤銷，否則 1 小時窗口太大。
- **已考慮 alternative**：Web 也用 JWT（舊系統 `X-Sowinsoft-Token`）— 拒絕，瀏覽器儲存 token 風險高且舊 token 無 exp 是已知缺陷。

### D6：員工僅 Google Workspace OIDC；新帳號走 guest 註冊完成頁 + 人工審核

- **選擇**：1.0 僅支援 Google Workspace（自建 IdP 延後）；首次 OAuth 登入無帳號時，**未完成註冊完成頁（選公司 + 姓名）前不建帳號**；完成後建立 `guest`（pending、歸屬所選公司），由 super / company_admin 審核指派部門與角色；註冊頁公司清單僅回傳 `display_name`。
- **理由**：集團已用 Google Workspace；先選公司再建帳號避免 orphan 帳號，僅暴露 display_name 避免公司名單外洩。
- **已考慮 alternative**：自動建立帳號後補審 — 拒絕，會產生無歸屬帳號；dept_admin 審核 guest — 拒絕，審核需跨部門視野，限 super/company_admin。

### D7：客戶編號 = 公司前綴 + 自增 ID；訂單編號 = 來源碼 + 自增序號；皆樂觀鎖取號同一交易

- **選擇**：`customer_code` = `companies.customer_code_prefix`（大寫英數 1–4 碼、全系統唯一）+ 6 位補零自增（如 `TY000123`），由 `customer_counters` 樂觀鎖取號，公司內唯一、建立後不可修改；修改前綴不影響既有帳號、計數器不重置。`order_no` = 下單來源碼（系統字典，如 `W`/`A`）+ 6 位補零序號，由 `order_counters`（company_id + source）樂觀鎖取號。取號與建檔/建單同一資料庫交易，version 衝突重試。
- **理由**：前綴全系統唯一使 `customer_code` 可全域定位客戶（QR token、客服溝通）；樂觀鎖取號避免 sequence 跨公司共享或 select-for-update 鎖競爭。
- **已考慮 alternative**：PG sequence — 拒絕，無法按公司+來源分軌且有跳號語意問題；UUID — 拒絕，人員需口述/手抄編號。

### D8：開發者逃生門 — `developer` 角色繞過 Casbin + RLS，環境開關 + fail-fast 防誤開

- **選擇**：第 7 內建角色 `developer`（`is_system=true`、`data_scope=all`），通過認證後 middleware 跳過 Casbin、RLS 注入 `data_scope=all`。`DEVELOPER_ACCOUNT_ENABLED`（dev 預設 true、prod 預設 false）；`ENV=production` 且開關為 true 時後端**拒絕啟動**；開發環境 seed 本機開發者帳號；developer 操作照常寫 `audit_logs`；上架/上線檢查清單必含關閉確認。
- **理由**：開發/除錯需要跨租戶視角，但不能以犧牲生產隔離為代價；fail-fast 讓誤開在部署時爆炸而非靜默開洞。
- **已考慮 alternative**：super 兼用 — 拒絕，super 本身受 policy 限制且會汙染權限模型；feature flag 遠端開關 — 拒絕，引入外部依賴，env + 啟動防護更簡單可靠。

### D9：角色與權限為 seed 預設值，Web 兩頁面可調，防鎖死

- **選擇**：7 內建角色、功能權限（`role_permissions`：resource × action）與 Casbin policy 皆由 migration seed 建立（`is_system=true` 不可刪、不可改 `code`/`data_scope`）；super 可新增自訂角色（必須指定 `data_scope` 為 `company` 或 `department`）；Web 提供「角色權限設置」與「API 權限設置」兩頁面；`company_admin` 僅限自己公司 domain；policy 異動即時生效（`e.AddPolicy`/`RemovePolicy` 不重啟）；防止移除操作者自身角色的最後一個管理權限；所有權限異動寫稽核。
- **理由**：預設值 seed 讓新環境開箱即用且權限模型可調整而不需改碼；防鎖死避免管理員把自己鎖在門外。
- **已考慮 alternative**：權限硬編碼 — 拒絕，每家公司的職責切分不同，上線後調整需發版不可接受。

### D10：統一軟刪除 `deleted_at` + 部分唯一索引

- **選擇**：業務實體統一 `deleted_at TIMESTAMP NULL`；業務唯一性（`products.code`、`customers.customer_code`、`customer_products` 客戶+商品）一律搭配部分唯一索引（`WHERE deleted_at IS NULL`）；列表預設排除已刪除；軟刪除主檔的歷史訂單/單據仍可顯示名稱編號但不可再選用；復原 = 清 `deleted_at` + 寫稽核；硬刪除僅 super 於特定介面並強制留稽核；`notifications` / `audit_logs` / `sales_order_events` 不適用軟刪除；員工帳號僅停用不刪除。
- **理由**：全新系統無舊資料包袱，不需要舊系統的 `is_deleted` 相容欄位；部分唯一索引解決「刪除後無法重建同名資料」的經典問題。
- **已考慮 alternative**：硬刪除 — 拒絕，歷史單據需要關聯名稱；狀態欄位假裝刪除 — 拒絕，語意混雜。

### D11：字典檔單表 `metadicts` — 系統預設（`department_id IS NULL`）+ 部門擴充

- **選擇**：單一 `metadicts` 表（`type` / `code` / `display_name` / `department_id` / `sort_order` / `is_active`），查詢取「系統預設 + 當前部門」；`type + code + department_id` 唯一，以 PG15 `NULLS NOT DISTINCT` 或部分唯一索引處理 NULL；super 維護系統級、dept_admin/staff 維護自己部門；1.0 不支援部門覆寫系統預設。倉別/車次/分切規格/商品分類為部門級**獨立實體表**（非 metadicts）；訂單來源/稅率/幣別為系統級固定。
- **理由**：兩層可見性用一張表 + NULL 語意即可表達，不必兩表 JOIN；獨立實體表留給有自身欄位與關聯的主檔。
- **已考慮 alternative**：系統表 + 部門表兩張 — 拒絕，查詢 UNION 複雜且型別不一致風險。

### D12：全系統不儲存金額欄位

- **選擇**：1.0 全系統**不儲存任何金額欄位**——`sales_orders` / `sales_order_items` 不含單價、小計、稅額、折扣；`products` 無 `default_price`、`customer_products` 無 `custom_price`；單據與報表不含金額。時間一律 `timestamptz`，顯示 UTC+8。
- **理由**：業主需求（需求備忘 2026-08-03：「整個資料庫都不需要儲存金額」）；系統只處理數量與出貨流程，金額由外部另行管理，資料模型與對帳大幅簡化。
- **已考慮 alternative**：金額 TWD 整數儲存（原 D12，v1.0.20–v1.0.27 定案）— 已修訂廢止；後續版本如需金額，另行擴充資料模型。
- 修訂來源：2026-08-03 需求備忘（客戶版規格書 v1.0.28）。

### D13：訂單狀態機含回退與終態，所有異動寫 `sales_order_events`

- **選擇**：`pending ⇄ processing → completed`；`pending` 可取消（`cancelled`）；取消派車（processing → pending）dept_admin 以上 + 原因，清 `dispatched_at`/`dispatched_by` 但**保留** `route_id`/`delivery_sequence`（訂單停留看板原位）；作廢（completed → `voided`）dept_admin 以上 + 原因，終態不可恢復，更正方式 = 作廢 + 重建（新單備註原單號）；所有異動寫 `sales_order_events`，取消派車與作廢另寫 `audit_logs`。
- **理由**：派錯車與內容錯誤是真實營運情境；「保留看板位置」讓重新派車不必重排；作廢+重建保留完整操作軌跡。
- **已考慮 alternative**：已完成可直接編輯 — 拒絕，已出貨單據與帳務會不一致。

### D14：派車批次確認 + 看板拖放樂觀鎖 + Connect 串流推播

- **選擇**：派車確認以**車次批次**執行（一次確認該車次當日全部待派訂單，逐筆記 `dispatched_at`/`dispatched_by` → processing）；拖放提交比對 `sales_orders.version`，衝突拒絕 + 重新整理看板；看板以 `DispatchService.WatchBoard` server streaming 推送型別化事件（`sales_order_id`、`route_id`、`delivery_sequence`、`version`）；事件僅觸發看板查詢 invalidate 全量重查；Valkey pub/sub 跨 replica 廣播（部門分 channel）；斷線 backoff 重連、連續失敗降級 30 秒輪詢。
- **理由**：批次確認符合實際派車作業（一次發一台車）；樂觀鎖防止同時拖放互相覆蓋；串流複用 Connect-RPC 技術棧（proto、cookie/JWT 認證、ingress、metrics），無 ws_ticket 與 upgrade 轉發等 WS 專用配套。
- **已考慮 alternative**：WebSocket — 拒絕（ws_ticket／upgrade 轉發／自訂協議成本）；SSE — 拒絕（游離 proto 型別系統外）；純輪詢 — 保留為降級模式；逐筆確認 — 拒絕，操作繁瑣。
- 修訂來源：2026-07-22 原 OpenSpec change `dispatch-board-connect-streaming`（OpenSpec 工作流已停用，決策已併入本記錄；客戶版規格書同步升 v1.0.27）。

### D15：Gotenberg HTML→PDF，四單據各有內容限制，重印需原因，PDF 全留存

- **選擇**：Gotenberg 將 HTML/CSS 模板轉 A4 PDF，字型 Noto Sans CJK TC；單車總表**不顯示金額**、對點單每店一張**不顯示價格**、揀貨單依車次→倉別→分類→品名、加工單分加工室揀/配送揀且「加工後數量」列印空白手寫回填（1.0 不回寫）；空表不印。預覽任何狀態皆可（寫 `print_previews`）；正式列印需 processing（寫 `print_logs`）；重印開放 dept_admin/staff 但必填 `reprint_reason`；每次 PDF 關聯 `file_assets` 留存。
- **理由**：單據給司機/倉庫/加工人員看，金額是機敏資訊不該出現在現場單據；PDF 留存讓「當初印了什麼」可查核。
- **已考慮 alternative**：wkhtmltopdf / 前端列印 — 拒絕，字型與中文渲染品質不穩；重印限管理員 — 拒絕，現場重印是 staff 日常，用原因+記錄控管即可。

### D16：通知兩通道（FCM + 站內）、無 Email，失敗不重試，失效 FCM token 即刪

- **選擇**：`notification_templates`（公司/部門級，`fcm`/`in_app` 兩通道，`{{變數}}`）+ `notifications` 記錄；**1.0 不使用 Email**（移除 `companies.smtp_config` 與 Email 範本）。訂單通知路由：業務下單 → 推播該客戶全部子帳號（主帳號不接收業務通知）；客戶自行下單不另行通知；後台新增客戶專屬商品 → 推播主責業務（檢核待定）；退貨審核結果 → 推播發起帳號。1.0 發送失敗**不重試**僅標 `failed` + 原因；FCM 回報 unregistered/invalid 時直接刪除 `user_devices`。
- **理由**：業主需求（「email 通知 全部不用」）；重試佇列基礎設施成本在 1.0 不成比例；失效 token 持續發送會被 FCM 懲罰。
- **已考慮 alternative**：Email 通道 + 公司級 SMTP（原 D16）— 已修訂移除；訊息佇列 + 重試 — 延後，記錄 failed 已足夠人工追蹤。
- 修訂來源：2026-08-03 需求備忘（客戶版規格書 v1.0.28）。

### D17：檔案本地儲存 + 白名單雙重檢查，備份至 GCS

- **選擇**：1.0 檔案存本地 NFS / volume，`file_assets.storage_path` 記錄路徑，後端提供下載 URL；上傳白名單（圖片 jpeg/png/webp ≤ 5 MB、PDF ≤ 10 MB），副檔名 + magic bytes 雙重檢查；不掃毒（僅內部使用者）；restic 每日備份至 GCS。
- **理由**：1.0 檔案量小（Logo、公告圖、列印 PDF），本地 + 異地備份足夠；物件儲存延後。
- **已考慮 alternative**：直接上 GCS — 拒絕，1.0 不需要，存取延遲與成本無效益。

### D18：稽核日誌同事務同步寫入

- **選擇**：關鍵操作（登入、下單、主檔異動、刪除、列印、強制登出、角色變更、取消派車、作廢）的稽核記錄與業務操作**同一資料庫交易**寫入，同成功同失敗；不採非同步佇列。
- **理由**：稽核缺漏比稽核擋下操作更糟；同步寫入保證「有操作必有稽核」。
- **已考慮 alternative**：非同步 outbox — 拒絕，1.0 量級下同步成本可忽略，一致性價值更高。

### D19：自建 PostgreSQL / Valkey StatefulSet + PITR，Prometheus 全家桶

- **選擇**：DB 與快取容器化部署於 k8s StatefulSet；每日完整備份 + WAL 歸檔（PITR）至 GCS（日備 30 天、月備 12 個月）；Valkey RDB+AOF；監控 Prometheus + Grafana + Alertmanager 同叢集；後端暴露 `/metrics`（API 延遲、錯誤率、Connect-RPC 狀態碼、業務指標）；**上線前必須完成首次 PITR 還原演練**驗證 RTO 4h / RPO 1h。
- **理由**：業主要求自持有資料；自建 StatefulSet 的代價是 DR 責任，故 PITR 演練列為上線門檻。
- **已考慮 alternative**：Cloud SQL / 託管 Redis — 拒絕（成本與資料主權考量，業主決策），但承認維運負擔轉嫁為備份/演練要求。

### D20：AI 整合僅預留欄位與公開端點；1.1 AI 輔助為獨立迭代

- **選擇**：`companies.capabilities` / `public_info` JSON 欄位 + `identifier` 穩定代號 + 免認證公開端點 `GET /api/v1/companies/public/{identifier}`；不實作完整 MCP/ACP/A2A server；未來協定適配層獨立（`internal/agent` 或獨立服務）不侵入業務 domain。拍照建客戶（LLM vision）與業務語音下單（雲端 STT + LLM）定案 1.1；App 手動表單新增客戶提前納入 1.0。
- **理由**：預留發現機制的資料結構成本極低，綁死協定成本高；AI 功能需 POC 驗證供應商，不該擋 1.0 上線。

### D21：三端覆蓋率 70% CI 強制 + 關鍵路徑整合測試

- **選擇**：後端 testify + dockertest、前端 Vitest、App flutter_test + Maestro；三端單元覆蓋率 ≥ 70% 由 CI 強制；認證、授權/RLS、訂單狀態機、樂觀鎖取號、退貨審核、偏好送貨日順延需整合測試。
- **理由**：Big Bang 上線沒有並行期兜底，測試是唯一防線。
- **已考慮 alternative**：更高門檻 — 拒絕，70% 是可達成且有意義的起點，關鍵路徑用整合測試補強。

### D22：客戶帳號一主多子（1 主帳號 + 多子帳號）

- **選擇**：每個客戶綁定 **1 個主帳號**（建立客戶時自動產生，帳號名稱預設為客戶名稱，`is_primary = true`）+ 多個**子帳號**（如主廚）；各帳號獨立密碼、可獨立停用/重置；`customer_code` 為客戶身分識別（客戶編號，非登入帳號），各帳號另有帳號名稱。**主帳號僅供管理、不提供業務登入**（登入後僅能帳號管理，業務 API 一律 403）；**子帳號是唯一的 App 業務登入身分**（老闆本人需下單時另建子帳號）；QR Code 帶入客戶身分後由使用者選擇任一子帳號完成登入。店家以主帳號自助管理子帳號（見 D28）；主帳號不可由店家自行停用，後台可重置主帳號密碼轉交。
- **理由**：需求備忘：「避免有人離職時如果是共用帳號都要重設登入改密碼」——共用帳號無法區分操作者且重設成本高；主帳號提供店家一個穩定、不會被自己鎖死的管理身分；主帳號與業務操作分離（2026-08-03 定案）讓管理軌跡乾淨、子帳號才承載日常業務。
- **已考慮 alternative**：維持單一帳號 + 離職重設 — 拒絕，共用帳號無稽核歸屬且重設擾民；多個平級帳號無主從 — 已修訂為一主多子（2026-08-03 討論定案）。
- 修訂來源：2026-08-03 討論（客戶版規格書 v1.0.30）。

### D23：App 推播路由

- **選擇**：訂單建立 → 推播該客戶全部子帳號（業務下單時，主帳號不接收業務通知）；客戶專屬商品新增（後台 / Web 建立時）→ 推播主責業務（`default_sales_rep_id` → dept_admin；是否需業務檢核待定）；退貨審核結果 → 推播發起客戶帳號。
- **理由**：需求備忘指定路由；把「誰下單、誰要知道」寫成明確規則。
- **已考慮 alternative**：維持推給下單業務（原 D16 路由）— 已修訂，客戶需要即時知道自己的訂單。

### D24：促銷推播與公告分離

- **選擇**：公告（`announcements`）顯示於 App 首頁消息區、所有客戶可見，**不等同推播**；促銷推播另以分類標籤運作：`promo_tags` 部門級標籤（商品 / 客戶專屬商品標記 `promo_tag_ids`），客戶資料套用（訂閱）分類，後台依分類選擇要推播的客戶群（FCM + 站內）。
- **理由**：需求備忘：「推播與公告是 2 件事」；分類訂閱讓「推給誰」由客戶自己勾選的類別決定，避免全體轟炸，也讓業務挑選客戶群有依據。
- **已考慮 alternative**：後台直接選客戶清單 — 客戶多時維護成本高；公告兼作推播 — 語意混雜且無法分眾。

### D25：退貨流程（客戶發起、業務確認）

- **選擇**：客戶 App 發起退貨申請——品項來源從「歷史訂單勾選」或「客戶專屬商品清單選擇」擇一（1.0 兩者並存，UX 依試用回饋收斂）——填數量/原因/備註/照片（多張）→ `pending` → 主責業務審核（approved / rejected，拒絕需原因）→ 客戶端顯示退貨證明（可出示配送司機）。公司端**不建置配送專屬頁面**，取貨由業務自行通知配送。
- **理由**：需求備忘；退貨證明解決「司機如何知道這趟要收貨」的現場問題，且不把配送系統捲入 1.0 範圍。
- **已考慮 alternative**：僅歷史訂單來源 — 客戶操作複雜度待驗證，兩者並存再收斂；獨立配送頁面 — 拒絕，範圍過大。

### D26：偏好送貨日（自動順延）

- **選擇**：`customers.preferred_delivery_days` 星期一到六核取（JSON 布林陣列）；下單選擇預計出貨日時，若非勾選日則 `expected_delivery_date` 自動順延至下一個有勾選的日期（未勾選任何日則維持原選擇）。
- **理由**：需求備忘（8/3 新增需求）：客戶只收特定天數的貨，避免業務下單落在不送貨日造成派車/退貨。
- **已考慮 alternative**：下單時提示人工改日 — 容易漏，自動順延最省業務心力且規則明確。

### D27：稽核日誌保留 3 個月、管理頁可設定

- **選擇**：`audit_logs` 預設保留 **3 個月**，`super` / `company_admin` 於管理頁面設定 1 / 3 / 6 / 12 個月或永久；保留排程到期刪除。
- **理由**：需求備忘：「稽核日誌保留時間 3 個月就好（或者可以在管理頁面自行選擇要存多久）」；彈性設定取代固定值，兼顧法務與儲存成本。
- **已考慮 alternative**：固定 2 年封存（原 §14.1）— 已修訂；固定 3 個月不可調 — 無法因應各公司法務需求。

### D28：店家以主帳號自助管理子帳號

- **選擇**：店家以**主帳號**於 App 管理自己客戶底下的帳號：新增子帳號（填寫帳號名稱，走臨時密碼流程）、停用子帳號、重置子帳號密碼；**子帳號僅能登入使用，無帳號管理權限**。**主帳號登入僅提供帳號管理功能**（業務 API 一律 403）；**建立客戶時自動附帶 1 個業務子帳號**（取代首登引導），**專供該客戶的所屬業務（`default_sales_rep_id`）使用**：憑證交付所屬業務（業務以客戶身分登入、代客操作），店家開箱僅持主帳號（管理用），需下單時以主帳號自行建立子帳號。該帳號於店家帳號管理清單**灰化（反白）顯示為「系統預設（業務使用）」、店家不可改名 / 停用 / 重置**（店家無密碼即無法登入），後台可正常管理並於主責業務變更時移交（重置密碼轉交）。範圍僅限自己客戶（data_scope self）。防呆：**主帳號不可由店家停用或重置**；後台 dept_admin 以上仍可停用 / 重置任何店家帳號（含主帳號，重置主帳號密碼可轉交新負責人）作為逃生門；**後台停用主帳號連鎖停用全部子帳號**。帳號管理異動寫入稽核日誌。
- **理由**：「有人離職不用全店重設」的動機要閉環——若只有後台能管理，離職時店家仍需聯絡業務等待處理；店家自助讓老闆當場停用離職者帳號，後台負擔同步減輕。主帳號固定不可自停，確保店家永遠有一個能登入的管理身分；RLS self 範圍 + 防呆 + 後台逃生門把誤操作風險壓低。
- **已考慮 alternative**：僅後台管理（dept_admin）— 行政負擔與等待成本仍在，離職場景未閉環；店家可停用任何帳號（無防呆）— 拒絕，可能把自己鎖死，必須有逃生門；任一帳號皆可管理其他帳號 — 拒絕，權限擴散，管理軌跡不清。
- 修訂來源：2026-08-03 討論（客戶版規格書 v1.0.30）。

### D31：後端結構慣例對齊 go8（集中 DI / cmd 拆分 / config 逐檔 / third_party）

- **選擇**：採納 go8 四項結構慣例——`internal/server`（Server struct + `Init()` + `InitDomains()` 集中 DI）、cmd 拆分（`cmd/server` 極薄入口 + `cmd/migrate` goose CLI + `cmd/seed` 冪等 seeder）、config 逐檔 struct（`kelseyhightower/envconfig`，tag 一律 `envconfig:"KEY"`，`config.New()` 聚合；取代原規劃的 viper/mapstructure 單檔 `config.Load()`）、`third_party/` 集中外部套件初始化（Ent/pgx、Valkey）。
- **D31-2 工具鏈**：air hot reload（`.air.toml`）；Taskfile 增 `dev`/`check`（fmt+vet+lint+test）/`vuln`（govulncheck）/`migrate`/`seed`；CI Go job 加 govulncheck。
- **不採納**：e2e 臨時容器 harness（D21 整合測試 + Phase 8 驗收已覆蓋）；OTel traces/logs（D19 已定 metrics-only）；protovalidate/validator（proto 強型別 + usecase 驗證承擔）；sqlx（與 Ent 重疊）；go8 REST/DB-session 風格（衝突 D4/D5）；`cmd/route`（API 面由 proto 定義）；go8 介面子套件分層（對既有計畫 churn 過大）。
- **理由**：集中 DI 讓啟動依賴與 fail-fast 檢查一目瞭然；cmd 拆分讓 migrate/seed 不依賴 server 啟動；config 逐檔有 code completion 且新增 key 有明確歸檔流程。
- 修訂來源：2026-08-24 設計文件 `docs/superpowers/specs/2026-08-24-backend-go8-structure-design.md`；參考 https://github.com/sowiner/go8。（插入位置依編號排序；D29/D30 條目由其各自計畫執行時補入本節。）

## Risks / Trade-offs

- [Risk] RLS 與 Ent 整合複雜（connection hook 注入 session variables，每個查詢路徑都要正確設定） → Mitigation: 統一 connection hook + 整合測試驗證各 `data_scope` 等級（執行計畫 Task 1.3 驗收即含跨角色隔離測試）。
- [Risk] Big Bang 上線前的主檔人工建檔（客戶、商品、車次、專屬商品清單）成為隱形關鍵路徑，目前無建檔計畫 → Mitigation: 上線檢查清單納入主檔建檔完成確認；App/Web 建檔介面（含 App 新增客戶）提早於 Phase 3/6 交付，讓建檔可提前開始。
- [Risk] `DEVELOPER_ACCOUNT_ENABLED` 生產誤開 → Mitigation: fail-fast 拒絕啟動 + CI/CD 部署檢查清單 + 上線檢查清單三重防護；但 env 被直接修改仍是殘餘風險，靠稽核監控（developer 帳號登入告警）補。
- [Risk] OAuth2 首登待審核流程在上線初期可能造成員工無法登入 → Mitigation: 上線前由管理員預先建立員工帳號。
- [Risk] 單據模板需求變動頻繁 → Mitigation: HTML/CSS 模板 + 列印服務獨立，調整不改業務邏輯。
- [Risk] 自建 PostgreSQL / Valkey（StatefulSet）維運與 DR 複雜 → Mitigation: 每日備份 + WAL PITR + 上線前首次還原演練 + 每半年 DR 演練 + Prometheus 告警。
- [Trade-off] 通知失敗不重試 → 接受理由：1.0 量級下失敗率低，`failed` 記錄 + 告警足夠人工補救；重試佇列留 1.1+。
- [Trade-off] 檔案本地儲存非物件儲存 → 接受理由：1.0 檔案量小，restic 異地備份補足；搬 GCS 的遷移成本留給需要的時候。
- [Trade-off] 稽核同步寫入增加每筆寫入延遲 → 接受理由：換取「有操作必有稽核」的強一致保證，量級下成本可忽略。

## Migration Plan

無資料遷移（D1）。部署與切換順序：

1. Phase 0–7 完成開發與測試（三端覆蓋率 ≥ 70%、關鍵路徑整合測試通過）。
2. Phase 8：k8s 部署 staging → 生產 manifests、CI/CD、備份排程、監控告警就緒。
3. 上線前檢查：`DEVELOPER_ACCOUNT_ENABLED=false`（fail-fast 會擋誤開）、首次 PITR 還原演練完成並留存紀錄、最終壓力測試、備份還原至測試環境驗證。
4. 主檔人工建檔（客戶/商品/倉/車次/分切/專屬商品/員工帳號預建）。
5. 全面切換，廢止舊系統。
6. Rollback 策略：無並行期，事實上不可回滾至舊系統 → 以「上線前完整演練 + 上線後快速修復」為策略；資料安全靠 PITR（RPO 1h）。

## Open Questions

| 問題 | 影響 | 狀態 |
|---|---|---|
| Noto Sans CJK TC 生產環境字體授權與安裝方式（Gotenberg 容器） | 單據列印 | 待確認（優先級低，上線前必答） |
| `audit_logs` 時間分區啟動時機 | 維運 | 觀察指標已定（單月百萬列或查詢變慢），上線後 3–6 個月評估 |
| 複合索引細節 | 效能 | 各 Phase Ent schema 建立時依 `docs/superpowers/specs/2026-07-17-sales-order-1.0-suggestions.md` §3 一併處理 |
| 1.1 AI 供應商選型 POC（名片 20 張、語音 20 句） | 1.1 | 1.1 啟動前完成，不擋 1.0 |
| 退貨品項來源 UX（歷史訂單 vs 專屬商品清單） | 退貨 | 1.0 兩者並存，試用回饋後收斂 |
| 專屬商品新增推播是否需業務檢核 | 通知 | 待定（v1.0.28） |

---

*遷移自原 OpenSpec change `sales-order-1-0`（2026-08-03，OpenSpec 工作流停用）*
