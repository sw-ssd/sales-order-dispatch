# 多公司訂出貨系統 1.0 — 規劃整合總覽

> 本文件整合所有先前規劃文件（原 `openspec/` 工作流已於 2026-08-03 停用並移除，內容遷移至 `docs/superpowers/specs/`），是後續**新增需求**與**開始實作**的單一入口。
> 整合日期：2026-08-03（最後一次規劃活動：2026-07-22）
> 核心狀態：**全新 monorepo 重建、規劃定稿、尚未實作**（0/79 Task 已開始，未部署）
> 版本凍結：規格書 **v1.0.34** 已凍結為實作基準（2026-08-03 起；需求變更須升版）

---

## 1. 文件地圖（角色與權威層級）

1.0 規劃文件皆位於 `docs/superpowers/`（客戶版規格書、執行計畫、報告、決策記錄、需求規格）。權威順序如下：

| 文件 | 角色 | 狀態 |
|---|---|---|
| `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md` | **決策記錄（D1–D28）**：每項決策的選擇 / 理由 / 已考慮 alternative / 風險 | ✅ 已定稿（2026-08-03 遷入） |
| `docs/superpowers/specs/1.0-requirements/`（12 份） | 需求層規格（Requirement + Scenario）；各領域可驗證行為基準 | ✅ 已定稿（2026-08-03 遷入） |
| `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34，52KB） | 客戶版完整規格（18 章）；**細節欄位與流程的唯一權威** | ✅ 已定稿 |
| `docs/需求備忘_2026-08-03.txt` | 2026-08-03 新增需求（不存金額 / 無 Email / 客戶多帳號 / 促銷推播 / 退貨 / 偏好送貨日 / 稽核保留） | ✅ 已對應至規格書 v1.0.34 |
| `docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md`（v2.9.0，51KB） | 執行計畫；每個 Task 的 Goal / Files / Steps / Acceptance Criteria 以此為準 | ✅ 已收斂為單一計畫 |
| `docs/superpowers/plans/archive/2026-08-03-sales-order-1.0-vibecheck-plan.md` | vibe-check 重整建置計畫（22 節：問題/目標/流程圖/功能/架構/成本/時程/檢查點/風險假設） | ✅ 已產生 |
| `docs/superpowers/reports/2026-08-03-sales-order-1.0-prd.html` | 互動式 PRD（9 分頁 + 6 張內嵌看板 + 續航快照），單一自包含檔案 | ✅ 已產生 |
| `docs/superpowers/specs/2026-07-18-app-ai-assist-1.1-design.md`（v0.2.0） | 1.1 AI 輔助備忘（拍照建客戶、語音下單），**獨立迭代不影響 1.0** | ✅ 方向定案 |
| `docs/superpowers/specs/2026-07-17-sales-order-1.0-suggestions.md` | 1.0 範圍外建議備忘（強制更新、電子發票、索引、audit 分區） | ✅ 已產生 |
| `docs/superpowers/reports/archive/2026-07-17-status-alignment.md`（v1.0.14） | 規格/計畫/狀態對齊報告（時點報告，已封存） | 📦 已封存（2026-08-03） |
| `docs/superpowers/reports/`（客戶報告 PDF / HTML ＋ 產生腳本） | 客戶端報告 v2.14（16 頁，對齊 spec v1.0.34 ＋ D29 App 技術棧定案；2026-08-04 以 Edge headless 重新產生；PPTX 已移除） | ✅ 已重新產生 |
| `docs/客戶需求.txt` | 原始需求來源（陳老闆，業務端彙整） | ✅ 已全數對應至 1.0 |
| `docs/AGENTS.md`、`docs/FUNCTION_LIST.md` | **現行三倉系統**（`sales-order-backend` / `-frontend` / `-app`）導覽 | 🟡 與 1.0 計畫不同系統 |

> 註：現行三倉系統是 1.0 的**前身**。1.0 為全新 monorepo 從頭開發，不遷移資料、不沿用舊碼；現行系統的 `AGENTS.md` 僅供背景參考。

---

## 2. 已定案決策摘要（D1–D28，詳見決策記錄）

| # | 決策 | 一句話 |
|---|---|---|
| D1 | 全新 monorepo、不遷移資料 | 新倉庫 `backend/` + `frontend/` + `app/` + `infra/`（pnpm workspace + Turborepo），主檔與訂單全部重新建檔 |
| D2 | Big Bang 上線 | 測試完成後全面切換、廢止舊系統；無試點、無並行、無唯讀期 |
| D3 | 租戶隔離三層分工 | Casbin 管功能（domain=company_id）、PostgreSQL RLS 管資料範圍（`data_scope` 等級：all/company/department/self，不依角色名稱）、CASL 管 UI |
| D4 | Connect-RPC 唯一 API 來源 | proto `v1` 產生三端型別；REST 僅留公開端點（版本、OAuth2、QR 兌換、檔案、公司公開資訊） |
| D5 | 認證雙軌 + token_version 撤銷 | Web 用 scs+Valkey session cookie；App 用 access JWT 1h + refresh 30 天旋轉；停用/強制登出/改密碼/角色變更 → `token_version`+1 全數失效 |
| D6 | 員工僅 Google Workspace OIDC | 首登未完成註冊完成頁（選公司+姓名）不建帳號；`guest` 待審核（super/company_admin 審） |
| D7 | 編號 = 前綴 + 自增、樂觀鎖取號 | `customer_code` = 公司前綴（1–4 碼全系統唯一）+ 6 位自增；`order_no` = 來源碼 + 6 位自增；取號與建檔同一交易 |
| D8 | developer 逃生門 | 第 7 內建角色，繞過 Casbin+RLS；`DEVELOPER_ACCOUNT_ENABLED`（prod 預設關、誤開 fail-fast 拒絕啟動） |
| D9 | 角色權限 seed 預設、可調、防鎖死 | 7 內建角色 + `role_permissions` seed；Web 兩頁面（角色權限、API 權限）；company_admin 限自己公司 |
| D10 | 統一軟刪除 | 業務實體 `deleted_at` + 部分唯一索引；復原=清欄位+寫稽核 |
| D11 | metadicts 單表兩層可見性 | 系統預設（`department_id IS NULL`）+ 部門擴充；倉別/車次/分切規格/商品分類為獨立實體表 |
| D12 | 全系統不儲存金額 | 訂單/明細/商品/專屬商品皆無金額欄位（2026-08-03 修訂，原金額方案廢止） |
| D13 | 訂單狀態機 + 事件軌跡 | `pending ⇄ processing → completed`、`cancelled`、作廢 `voided`（終態）；取消派車保留看板位置；所有異動寫 `sales_order_events` |
| D14 | 派車批次確認 + 樂觀鎖 + Connect 串流 | 車次批次確認；拖放比對 `version`；`WatchBoard` server streaming 事件僅觸發 invalidate 全量重查；Valkey pub/sub 跨 replica；降級 30 秒輪詢（**WS / ws_ticket 已廢除**） |
| D15 | Gotenberg HTML→PDF | 單車總表不顯示金額、對點單不顯示價格、揀貨單依車次→倉別→分類→品名、加工單分加工室揀/配送揀 + 手寫回填；重印開放 dept_admin/staff 必填原因 |
| D16 | 通知兩通道（FCM+站內）、無 Email | 業務下單推客戶子帳號；後台新增專屬商品推主責業務；促銷推播（分類標籤選群）與公告分離；失敗僅標 `failed` |
| D17 | 檔案本地儲存 | 白名單（jpeg/png/webp ≤5MB、PDF ≤10MB）+ magic bytes 雙重檢查；restic 每日備份 GCS |
| D18 | 稽核日誌同事務寫入 | 關鍵操作與業務操作同一 DB 交易，同成功同失敗 |
| D19 | 自建 StatefulSet + PITR | PostgreSQL / Valkey k8s StatefulSet；每日備份 + WAL PITR 至 GCS（RTO 4h / RPO 1h）；Prometheus + Grafana + Alertmanager；上線前首次還原演練 |
| D20 | AI 整合僅預留 | `companies.capabilities` / `public_info` + 公開發現端點；1.1 AI 輔助（拍照建客戶、語音下單）獨立迭代 |
| D21 | 三端覆蓋率 70% CI 強制 | 認證、RLS、狀態機、取號、退貨審核、送貨日順延需整合測試 |
| D22 | 客戶帳號一主多子 | 1 主帳號（僅管理、不業務登入）+ 子帳號；建立客戶自動附帶業務子帳號（專供所屬業務、店家灰化不可管理）；主帳號不可自停、停用連鎖子帳號 |
| D23 | App 推播路由 | 業務下單推客戶子帳號、後台新增專屬商品推主責業務（檢核待定）、退貨審核結果推發起帳號 |
| D24 | 促銷推播與公告分離 | `promo_tags` 分類標籤 + 客戶套用 + 依分類選群推播（FCM + 站內） |
| D25 | 退貨流程 | 客戶發起（歷史訂單 / 專屬清單並存）→ 業務審核 → 退貨證明；不建配送頁 |
| D26 | 偏好送貨日 | `preferred_delivery_days`（一~六核取）；非勾選日下單自動順延 |
| D27 | 稽核保留 3 個月可設定 | 管理頁可設 1 / 3 / 6 / 12 個月或永久 |
| D28 | 店家以主帳號自助管理子帳號 | self 範圍 + 主帳號不可自停 + 後台逃生門（含移交業務子帳號） |

其他定案重點（P0/P1，詳見決策記錄）：
- 角色：`super` / `company_admin` / `dept_admin` / `staff`（兼會計）/ `customer` / `guest` / `developer`
- 密碼政策：臨時密碼 24h 效期、最少 8 字元、5 次錯誤鎖定 30 分鐘、首登強改
- App 手動表單新增客戶（主檔 + 登入帳號）**納入 1.0**（原為 1.1 範圍，v1.0.26 提前）
- 加工單「加工後數量」列印空白、手寫回填（1.0 不回寫）

---

## 3. 1.0 範圍

### In Scope
- 自建全部主檔（客戶、商品、業務、倉別、車次、分切規格、商品分類、字典檔）
- Company → Department 兩層租戶、三重防護（Casbin + RLS + CASL）
- 銷售訂單（狀態機、編號取號、客戶專屬商品清單、手打商品別名；**不儲存金額**）、退貨申請（客戶發起/業務審核/退貨證明）
- 派車看板（Kanban 拖放、批次確認、Connect 串流推播）
- 四種單據列印（單車總表、車次對點單、揀貨單、加工單）
- 通知（FCM + 站內，無 Email；業務下單推客戶子帳號）、促銷推播（分類標籤選群）、公告 CMS、稽核日誌（保留 3 個月可設定）、檔案資產
- OAuth2 員工登入、客戶帳號一主多子（建立客戶自動附帶業務子帳號、**專供所屬業務使用**、店家管理頁灰化）/ QR Code 登入（僅店家子帳號）、店家以主帳號自助管理子帳號（主帳號僅管理不業務登入；self 範圍 + 防呆 + 後台逃生門）、developer 逃生門
- k8s 部署、備份監控、Big Bang 上線

### Out of Scope（1.0）
供應商退貨授權、獨立報價單（Estimate）、簽核流程、進銷存庫存、配送專屬頁面、獨立會計角色、自建 IdP、MCP/ACP/A2A 完整協定、Cmd+K 全域搜尋、電子發票開立（僅留 `invoice_type_id`）、App 強制更新、通知重試佇列、上傳病毒掃描、多語 UI。

---

## 4. 原始需求對應（客戶需求.txt → 1.0）

陳老闆需求已全數落入 1.0 規劃，無遺漏：

| 原始需求 | 對應 |
|---|---|
| 產品分庫、揀貨單顯示庫別 | master-data（倉別/庫別）；揀貨單依車次+倉別列印（D15） |
| 產品名稱手打 + 清單選擇 | sales-orders 手打商品存別名（綁定客戶） |
| App 業務直接刪除品項 | customer_products：qty=0 保留不顯示（偶爾不用）、刪除不再帶出（長期不用） |
| 手寫備註欄（特殊切法） | 訂單明細備註、加工單加工室揀/配送揀兩區塊 + 手寫回填 |
| 單一入口切換部門、總/部門/一般三級權限 | multi-tenancy + authorization（super / company_admin / dept_admin / staff） |
| 各部門資料庫分隔、僅總管理員調閱全部 | RLS `data_scope`（company/department 等級） |
| 各部門自建商品/客戶總表 | master-data（`department_id` 隸屬） |
| 客戶專屬商品清單（業務建立、打開客戶自動帶入） | customer_products（D10、Task 3.5） |
| 商品總表 + 客戶訂單手動改名（A 商品在不同客戶叫 BCDEF） | 別名機制（無需大總表） |
| 部門自建庫別、揀貨單依車次&倉別 | warehouses / routes 部門級實體表 + 揀貨單排序 |
| 分切規格 → 加工室揀/配送揀 | 加工單兩區塊（D15） |
| 四種單據（單車總表/對點單/揀貨單/加工單） | printing capability（Task 5.3–5.5） |
| 中台：管理員名單、客戶/產品檔案、派車、單據列印 | user management、master-data、dispatch、printing |

---

## 5. 文件間差異與已知不一致

1. **`status-alignment.md` v1.0.14（07-18）為時點報告，已於 2026-08-03 封存至 `docs/superpowers/reports/archive/`**：其中 §3.2「WebSocket 認證 / ws_ticket」、§3.4「派車看板 WebSocket」、已確認事項 #7 記錄的是該時點狀態；現行決策 = **Connect-RPC server streaming（WatchBoard）+ 輪詢降級，ws_ticket 已移除**（2026-08-03 已確認正確，見決策記錄 D14）。
2. **已修正事項（2026-08-03）**：
   - 客戶版規格書升 **v1.0.27**：§2.1 / §2.3 / §7.1 / §12.3.2 / §14.1 的 WebSocket 內容改為 Connect 串流，`ws_ticket` 端點移除；執行計畫與決策記錄引用同步更新。
   - Task 編號空洞重編：`1.12 → 1.11`（開發者帳號）、`1.13 → 1.12`（Phase 1 驗收），追蹤檔與執行計畫編號一致（76 個 Task，無缺口）。
   - **OpenSpec 工作流停用**：`openspec/` 目錄已移除；決策記錄遷至 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`、12 份需求規格遷至 `docs/superpowers/specs/1.0-requirements/`，其餘內容（範圍、任務追蹤、P0/P1 記錄）已收錄於本文件與既有文件。

---

## 6. 待辦與 Open Questions

### 待定討論項目（2026-08-03 需求備忘，尚未定案）
- 退貨品項來源 UX：歷史訂單勾選 vs 專屬商品清單選擇（1.0 兩者並存，試用回饋後收斂）
- 後台新增客戶專屬商品推播主責業務時，是否需業務檢核
- 促銷推播的分類標籤機制（初步構想，實作前確認）

### 1.1+（不擋 1.0）
- AI 供應商選型 POC（名片 20 張、語音 20 句）+ 名片輸出 schema 定案 + 語音 token 成本估算（1.1 啟動前）
- App 強制更新、通知重試佇列、上傳病毒掃描、多語 UI、自建 IdP、Cmd+K、audit_logs 時間分區（單月百萬列或查詢變慢再評估）

### 實作期間處理
- Noto Sans CJK TC 生產字體授權與安裝方式（上線前必答）
- 複合索引（`sales_orders(department_id, expected_delivery_date)` 等 6 組，Phase 3–5 schema 建立時）

---

## 7. 下一 session 接續指引（新增需求流程）

新增需求直接更新 `docs/superpowers/` 文件，原則如下：

1. **需求變更** → 客戶版規格書 `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（升版並補修訂記錄）+ 對應需求規格 `docs/superpowers/specs/1.0-requirements/<capability>/spec.md`（屬既有能力時）。
2. **決策變更** → 決策記錄 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（新增 D# 條目並加註修訂來源，比照 D14 模式）。
3. **任務變更** → 執行計畫 `docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md`（更新勾選進度；增刪 Task 時維持連續編號）。
4. **歷史文件不直接改**：`docs/superpowers/reports/` 為時點報告；客戶版規格書每次修改升版。
5. **1.1 事項**（AI 輔助）以 `2026-07-18-app-ai-assist-1.1-design.md` 為準，勿混入 1.0 範圍。

---

*最後更新：2026-08-03*
