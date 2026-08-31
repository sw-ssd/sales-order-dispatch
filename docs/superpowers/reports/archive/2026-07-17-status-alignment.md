# 多公司訂出貨系統 1.0 — 規格 / 計畫 / 狀態對齊報告

> 日期：2026-07-18  
> 狀態：**規格書 v1.0.26 已定稿，實作計畫已收斂為單一文件，尚未開始實作**

---

## 1. 文件對齊總覽

| 文件類型 | 檔案路徑 | 版本 / 日期 | 狀態 |
|---|---|---|---|
| 設計規格 | `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md` | v1.0.26 | ✅ 已定稿 |
| 實作計畫 | `docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md` | v2.8.0 | ✅ 已收斂為單一計畫 |
| 建議備忘 | `docs/superpowers/specs/2026-07-17-sales-order-1.0-suggestions.md` | v1.0.0 | ✅ 已產生（1.0 範圍外建議） |
| 客戶報告 PDF | `docs/superpowers/reports/多公司訂出貨系統_1.0_客戶報告.pdf`（來源：`customer_report.html`） | v2.1.0 | ✅ 已重新產生（對齊 spec v1.0.26，15 頁附圖） |
| 1.1 AI 輔助功能設計備忘 | `docs/superpowers/specs/2026-07-18-app-ai-assist-1.1-design.md` | v0.2.0 | ✅ 已定案方向（拍照建客戶、業務語音下單，1.1 獨立迭代） |
| 狀態報告 | `docs/superpowers/reports/2026-07-17-status-alignment.md` | v1.0.14 | ✅ 本文件 |

> 計畫文件已於 2026-07-17 收斂為單一實作計畫；舊版四份計畫（`2026-07-16-sales-order-2.0-phase1.md`、`2026-07-17-multi-company-sales-order-1-0.md`、`2026-07-17-multi-company-sales-order-1-0-companies.md`、`2026-07-17-sales-order-1-0-milestones.md`）已刪除，里程碑總覽與跨 Phase 注意事項併入實作計畫開頭。

---

## 2. 產品定位對齊

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 產品名稱 | 多公司訂出貨系統 1.0 | 多公司訂出貨系統 1.0 | ✅ 一致 |
| 核心目標 | 取代外部系統依賴；Company → Department 兩層租戶；強化派車/列印；公司識別與公開資訊 | 9 個 Phase 覆蓋基礎建設、權限、主檔、訂單、派車/列印、App、公告/UI、部署維運 | ✅ 一致 |
| 技術棧 | Go 1.25 + SolidJS 1.9 + Flutter 3.35.2 + pnpm + Turborepo | 相同 | ✅ 一致 |
| 開發策略 | 全新 monorepo 從頭開發 | Phase 0 建立 monorepo 骨架 | ✅ 一致 |
| 上線策略 | Big Bang：測試完成後直接全面上線，無並行/試點 | Phase 8 執行 Big Bang 上線與舊系統廢止 | ✅ 一致 |

---

## 3. 關鍵設計對齊

### 3.1 權限與多租戶

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 租戶層級 | Company → Department | Phase 2 公司/部門管理；所有業務表帶 `company_id` / `department_id` | ✅ 一致 |
| 角色 | super、company_admin、dept_admin、staff（兼會計）、customer、guest | Phase 1–2 實作角色與範圍 | ✅ 一致 |
| 後端授權 | Casbin RBAC with domain | Phase 1 Task 1.2 | ✅ 一致 |
| 前端授權 | CASL.js | Phase 1 Task 1.9 | ✅ 一致 |
| 資料庫隔離 | PostgreSQL RLS + session variables | Phase 1 Task 1.3 | ✅ 一致 |

### 3.2 認證

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 員工登入 | OAuth2 / OIDC（Google Workspace；自建 IdP 延後） | Phase 1 Task 1.4 | ✅ 一致 |
| 客戶登入 | customer_code + 密碼；customer_code = 公司定義前綴 + 自增 ID（v1.0.26） | Phase 1 Task 1.5；Phase 3 Task 3.1 取號 | ✅ 一致 |
| QR Code | 簽章 token 含 `company_id` + `customer_code` | Phase 1 Task 1.10 | ✅ 一致 |
| WebSocket 認證 | cookie / 一次性 `ws_ticket` | Phase 1 Task 1.11、Phase 5 Task 5.2 | ✅ 一致 |
| 強制登出 | 記錄並清除 session | Phase 1 Task 1.7 | ✅ 一致 |

### 3.3 資料模型

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 公司/部門/使用者 | §5.1 核心實體 | Phase 2 Task 2.1–2.3 | ✅ 一致 |
| 客戶地址簿/聯絡人 | `customer_addresses` / `customer_contacts` | Phase 3 Task 3.2 | ✅ 一致 |
| 商品與單位換算 | `products` 含 `code`、unit conversion | Phase 3 Task 3.3 | ✅ 一致 |
| 客戶專屬商品 | `customer_products` | Phase 3 Task 3.5 | ✅ 一致 |
| 字典檔 | `metadicts`：系統預設 + 部門擴充 | Phase 2 Task 2.5 | ✅ 一致 |
| 通知系統 | `notification_templates` / `notifications` | Phase 4 Task 4.3 | ✅ 一致 |
| 稽核日誌 | `audit_logs` 獨立表 | Phase 2 Task 2.6 | ✅ 一致 |
| 檔案資產 | `file_assets` 本地儲存 | Phase 3 Task 3.6 | ✅ 一致 |
| 列印記錄 | `print_logs` / `print_previews` | Phase 5 Task 5.5 | ✅ 一致 |
| 軟刪除 | 統一 `deleted_at` | 第 5.4 節；各 Phase schema 設計 | ✅ 一致 |

### 3.4 訂單、派車與列印

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 訂單狀態 | pending → processing → completed；cancelled | Phase 4 Task 4.1 | ✅ 一致 |
| 金額計算 | subtotal / tax / discount / total | Phase 4 Task 4.1 | ✅ 一致 |
| 派車看板 | Kanban、拖放、WebSocket | Phase 5 Task 5.1–5.2 | ✅ 一致 |
| 四種單據 | 單車總表、對點單、揀貨單、加工單 | Phase 5 Task 5.3–5.4 | ✅ 一致 |
| 列印權限 | 重印開放 staff，需填原因 | Phase 5 Task 5.5 | ✅ 一致 |

### 3.5 App

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 底部導覽 | 首頁 / 商品（快速下單） / 訂單歷史 / 功能 | Phase 6 Task 6.1–6.3 | ✅ 一致 |
| 業務下單 | 選客戶 → 帶出 customer_products → 提交 | Phase 4 Task 4.6、Phase 6 | ✅ 一致 |
| 客戶下單 | 只能從自己的 customer_products 下單 | Phase 4 Task 4.6 | ✅ 一致 |
| 離線快取 | Sembast 快取主檔與歷史 | Phase 6 Task 6.4 | ✅ 一致 |
| App 新增客戶 | 手動表單建立主檔 + 登入帳號（§9.4，v1.0.26 納入 1.0） | Phase 6 Task 6.6 | ✅ 一致 |
| FCM 推播 | 訂單狀態與派車通知 | Phase 4 Task 4.4、Phase 6 Task 6.5 | ✅ 一致 |

### 3.6 安全、合規與維運

| 項目 | 規格書 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| 安全與合規 | 第 14 章：資料保護、傳輸/儲存安全、應用程式安全、稽核監控 | Phase 8 Task 8.5 | ✅ 一致 |
| 備份策略 | PostgreSQL + Valkey + 檔案備份至 GCS | Phase 8 Task 8.3 | ✅ 一致 |
| 災難復原 | RTO 4 小時 / RPO 1 小時 | Phase 8 Task 8.6 | ✅ 一致 |
| 監控告警 | 基礎設施/應用/業務指標；工具待選 | Phase 8 Task 8.4 | ✅ 一致 |
| 上線過渡 | Big Bang，無試點/並行/唯讀期 | Phase 8 Task 8.7 | ✅ 一致 |

### 3.7 AI 整合預留

| 項目 | 規格書 §17 | 實作計畫 | 對齊狀態 |
|---|---|---|---|
| `capabilities` / `public_info` | `companies` 結構化 JSON 欄位 | Phase 2 Task 2.4 | ✅ 一致 |
| 公開發現端點 | `GET /api/v1/companies/public/{identifier}` | Phase 2 Task 2.4 | ✅ 一致 |
| MCP/ACP/A2A | 第一版僅預留，不實作完整協定 | Phase 2 僅建立欄位與端點 | ✅ 一致 |

---

## 4. 已確認事項

1. **會計角色由 `staff` 兼任**：規格書第 1.2 / 3.2 節已明確，不單獨設立 `acct`。
2. **monorepo 工具鏈**：pnpm workspace + Turborepo，Task 管理 Go/Flutter 原生任務。
3. **檔案儲存**：1.0 採用本地 NFS / volume，`file_assets` 記錄路徑；備份存放 GCS。
4. **資料遷移**：不從舊系統匯入任何資料，主檔與訂單皆重新建檔。
5. **上線方式**：Big Bang，無並行、無試點、無舊系統唯讀期。
6. **監控工具**：定案 Prometheus + Grafana + Alertmanager（規格書 v1.0.16）。
9. **角色與 API 權限設置**：Web 新增兩頁面；內建 6 角色、功能權限與 Casbin policy 皆為 migration seed 預設值，super 可新增自訂角色（規格書 v1.0.17）。
10. **開發者帳號**：第 7 角色 `developer` 不受限制（繞過 Casbin 與 RLS），以 `DEVELOPER_ACCOUNT_ENABLED` 控制，production 預設關閉、誤開拒絕啟動；上架/上線前必關（規格書 v1.0.18）。
7. **WebSocket 認證**：Web 使用 cookie；App / 無法帶 cookie 情境使用一次性 `ws_ticket`。
8. **列印權限**：重印開放給 `staff`，需填寫原因，記錄於 `print_logs`。
11. **App 新增客戶與客戶編號**：App 手動表單新增客戶（主檔 + 登入帳號）納入 1.0；`customer_code` = 公司定義前綴 + 自增 ID，前綴全系統唯一，`customer_counters` 樂觀鎖取號（規格書 v1.0.26）。
12. **AI 輔助功能定案 1.1**：拍照建客戶（LLM vision）與業務語音下單（雲端 STT + LLM 解析）為 1.1 獨立迭代，不影響 1.0 範圍，見 `2026-07-18-app-ai-assist-1.1-design.md`。

---

## 5. 待調整/封存事項

| 項目 | 說明 | 優先級 |
|---|---|---|
| 字體授權 | 單據使用 Noto Sans CJK TC，需確認生產環境字體授權與安裝方式。 | 低 |

---

## 6. 下一步建議

1. **確認本報告**：確認規格書 v1.0.16、實作計畫、客戶報告 PPT 內容無誤。
2. **開始實作**：依 `2026-07-17-sales-order-1-0-tasks.md` 從 Phase 0 開始執行。
3. **定期對齊**：每個 Phase 結束後更新本報告，確保規格、計畫、實作三者一致。

---

## 7. 修訂記錄

| 修訂號 | 日期 | 修訂內容 |
|---|---|---|
| v1.0.0 | 2026-07-17 | 初版；對齊 spec v1.0.2 與舊 plan。 |
| v1.0.1 | 2026-07-17 | 確認 1.0 採用全新 monorepo；更新 spec 為 v1.0.3。 |
| v1.0.2 | 2026-07-17 | 全面更新以對齊 spec v1.0.14、新里程碑與詳細任務清單；更新客戶報告 PPT 至 v1.0.2。 |
| v1.0.3 | 2026-07-17 | 對齊 spec v1.0.16（P0 決策定案）；計畫收斂為單一文件 `2026-07-17-sales-order-1-0-tasks.md`，舊版四份計畫刪除；監控工具定案 Prometheus + Grafana + Alertmanager。 |
| v1.0.4 | 2026-07-17 | 對齊 spec v1.0.17（Web 角色權限設置與 API 權限設置，原規劃權限為預設值）；計畫升版 v2.1.0（新增 Task 2.9–2.11）。 |
| v1.0.5 | 2026-07-17 | 對齊 spec v1.0.18（開發者帳號與環境開關）；計畫升版 v2.2.0（新增 Task 1.12，上線檢查清單加入關閉開發者帳號）。 |
| v1.0.6 | 2026-07-17 | 對齊 spec v1.0.19；計畫升版 v2.3.0：全部業務 API 介面改為 Connect-RPC（REST 僅剩公開端點）；Task 1.6 補 refresh token 旋轉與 `token_version` 驗證 middleware；Task 1.7 強制登出連動 `token_version + 1`；Task 8.4 展開 Prometheus / Grafana / Alertmanager 步驟；Task 8.6 補 PITR runbook 與上線前首次還原演練。 |
| v1.0.7 | 2026-07-17 | 對齊 spec v1.0.20（第 6.5 節訂單編號與金額計算：來源碼+自增、樂觀鎖取號、稅外加預設、整數四捨五入、1.0 無折扣）；計畫升版 v2.4.0（Task 4.1 / 4.2 補取號與金額規則）。 |
| v1.0.8 | 2026-07-17 | 對齊 spec v1.0.21（狀態機補取消派車回退與作廢終態、新增 `sales_order_events`）；計畫升版 v2.5.0（Task 4.1 / 5.1 補 Void 與 CancelDispatch）。 |
| v1.0.9 | 2026-07-17 | 對齊 spec v1.0.22（P1 決策：批次派車確認、看板樂觀鎖、通知對象下單業務、公司級 SMTP、失敗不重試、guest 註冊完成頁、加工單手寫回填、保留期限 / 上傳白名單 / 覆蓋率 70%）；計畫升版 v2.6.0。 |
| v1.0.10 | 2026-07-17 | 對齊 spec v1.0.23（單車總表不顯示金額；密碼政策：臨時密碼 1 天、最少 8 字元、5 次錯誤鎖定）；計畫升版 v2.7.0。全部 P1 待決項目定案。 |
| v1.0.11 | 2026-07-17 | 對齊 spec v1.0.24（全面複查：RLS customer_id、ability 快取、公司停用連鎖、稽核同步寫入、分頁上限等 9 項修補）；計畫升版 v2.7.1（onboarding 頁、公司停用阻擋、DeviceService）。 |
| v1.0.12 | 2026-07-17 | 剩餘建議（App 強制更新、電子發票、複合索引、audit 分區）收錄至建議備忘文件 `2026-07-17-sales-order-1.0-suggestions.md`。 |
| v1.0.13 | 2026-07-18 | 客戶報告重新產生為 PDF v2.0.0（15 頁、9 張 SVG 圖解、白話文案，對齊 spec v1.0.24）；來源改為 `customer_report.html`，舊 PPTX v1.0.2 停止維護。 |
| v1.0.14 | 2026-07-18 | 對齊 spec v1.0.26（App 手動新增客戶納入 1.0；customer_code = 公司定義前綴 + 自增 ID、新增 `customer_counters`）；計畫升版 v2.8.0（Task 2.1 / 2.7 補前綴維護、Task 3.1 補取號與帳號建立、新增 Task 6.6 App 新增客戶、原 6.6 驗收改 6.7）；客戶報告升版 v2.1.0；新增 1.1 AI 輔助功能設計備忘（拍照建客戶、業務語音下單，定案 1.1 獨立迭代）。 |
