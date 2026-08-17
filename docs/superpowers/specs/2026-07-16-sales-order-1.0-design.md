# 多公司訂出貨系統 1.0 設計規格書

> 修訂號：**v1.0.34**  
> 版本狀態：**已凍結為實作基準**（2026-08-03 起；後續需求變更須升版並補修訂記錄）  
> 本文件為 **多公司訂出貨系統 1.0** 的設計規格，以全新 monorepo 從頭開發，繼承現有 `FUNCTION_LIST.md` 功能範圍，移除外部系統整合，補全自建主檔與多公司/部門權限管理，並納入 `.txt` 的新需求。

---

## 1. 專案定位

### 1.1 目標

- 完全取代現行系統對外部系統的依賴，把客戶、商品、業務、報價等主檔改為自建管理。
- 導入「公司（Company）→ 部門（Department）」兩層租戶架構，結合 **Casbin（後端授權）+ CASL.js（前端權限）** 實現多公司權限管理。
- 強化倉儲、派車與單據列印流程，支援依車次、倉別、分切規格產生各類出貨單據。
- 以全新 monorepo 從頭開發，但沿用現有技術棧（Go + SolidJS + Flutter）。

### 1.2 不適用範圍（第一版不做）

- 供應商退貨授權（vendor return authorization）。
- 獨立報價單（Estimate）功能；由「客戶專屬商品清單」取代報價帶入。
- 簽核流程（approval workflow）。
- 進銷存庫存管理（僅記錄「庫別」與「分庫」屬性，不追蹤即時庫存）。
- 獨立會計（`acct`）角色；會計相關檢視/對帳權限由 `staff` 依所屬部門資料範圍兼任。

---

## 2. 技術架構

### 2.1 技術棧

| 層 | 技術 |
|---|---|
| 後端 | Go 1.25 + Chi + Connect-RPC（Web / App 業務 API 統一；REST 僅公開端點）+ Ent ORM + PostgreSQL + Valkey |
| 前端 | SolidJS 1.9 + TypeScript 5.9 + TanStack Query / Router + CASL.js |
| App | Flutter 3.35.2 + connectrpc + Sembast + flutter_secure_storage |
| 權限 | Casbin RBAC with domain（後端）+ CASL.js（前端） |
| 列印 | Gotenberg（HTML → PDF） |
| 派車看板推播 | Connect-RPC server streaming（輪詢降級） |
| 推播 | Firebase Cloud Messaging（FCM） |
| 部署 | Kubernetes + Traefik + GitHub Actions |
| Migration | Goose |
| Monorepo | pnpm workspace + Turborepo（全倉任務管線）+ Task（Go / Flutter 原生任務） |

### 2.2 倉庫結構

```text
sales-order-1.0/
├── backend/                 # Go 後端
│   ├── cmd/
│   ├── internal/domain/     # 各領域（auth, company, product, customer, order, dispatch, print, cms, audit）
│   ├── proto/               # .proto 檔案（後端為唯一來源）
│   ├── database/migrations/ # Goose migration（含 RLS policy）
│   ├── deploy/
│   └── Dockerfile
├── frontend/                # SolidJS 中台
│   ├── src/
│   ├── deploy/
│   └── Dockerfile
├── app/                     # Flutter App
│   ├── lib/
│   ├── proto/               # git submodule 或 copy 自 backend/proto
│   └── Dockerfile（僅供 CI 建置參考）
├── package.json             # pnpm workspace root
├── pnpm-workspace.yaml      # 定義 backend / frontend / app 等 workspace
├── turbo.json               # Turborepo 管線定義與遠端快取
├── infra/                   # k8s manifests、Helm charts、docker-compose dev
├── docs/
└── Taskfile.yml             # Go / Flutter 原生任務
```

### 2.3 API 與通訊

- **Connect-RPC**：Web / App 所有業務 API 統一使用，proto 版本 `v1` 為唯一 API 來源，走 HTTP/2 443 port，由 Traefik 依 path 轉發。
- **REST**：僅保留公開與認證回調端點（版本查詢、OAuth2 導向與 callback、QR Code token 兌換、檔案上傳/下載、`GET /api/v1/companies/public/{identifier}`），使用 `/api/v1/*`，OpenAPI/Swagger 文件。
- **派車看板即時推播**：以 Connect-RPC server streaming（`DispatchService.WatchBoard`）推送型別化看板事件；Web 端以 httpOnly cookie 認證（避免 token 出現在 URL 或伺服器 log），App 端沿用既有 access JWT header，無需額外 ticket；斷線自動重連，連續失敗降級為 30 秒輪詢。
  - 不建議將長期 access token 直接放在 URL query parameter。
- **型別同步**：Web（connect-es）與 App（Dart）型別皆由 proto 產生；REST 公開端點不產生客戶端型別。

### 2.4 API 標準

- **分頁回應格式**：沿用現有專案慣例，使用 `meta` 欄位承載分頁資訊。
  ```json
  {
    "data": [...],
    "meta": {
      "page": 1,
      "per_page": 20,
      "total": 100,
      "total_pages": 5
    }
  }
  ```
- 分頁請求參數：`?page=1&per_page=20`；`per_page` 上限 100，超過以 100 計。
- Connect-RPC 列表方法的 request / response 沿用相同分頁欄位（`page`、`per_page`、`total`、`total_pages`）。
- 若後續需要向後相容現有專案，可保留 `meta.options` 擴充欄位。

---

## 3. 權限與多租戶設計

### 3.1 租戶層級

- **Company（公司）**：最大管理單位，例如集團旗下不同法人。
- **Department（部門）**：實際業務單位，例如甲部門、乙部門。
- 資料表皆帶 `company_id` + `department_id`；`super` 可跨公司，`company_admin` 可跨部門但僅限自己公司，其他角色僅限自己部門。
- 公司停用（`status` 非 active）時，該公司所有帳號（含客戶）無法登入，已登入 session / token 於下次請求時失效。

### 3.1.1 公司識別與公開資訊

每間公司可獨立設定對外識別與公開資訊，供 Web 中台與 App 依登入使用者所屬公司動態呈現：

- **公司識別標識**：Logo 圖檔 URL、主色碼（primary color）、公司簡稱/英文名稱。
- **公開資訊**：公司電話、地址、Email、統一編號、營業時間、隱私權政策連結、服務條款連結等。
- **顯示場景**：登入頁、Web 側邊欄頂部、App 首頁與關於我們頁面、單據表頭（PDF）。
- 僅 `super` 與 `company_admin` 可編輯所屬公司識別與公開資訊；`super` 可上傳/更換 Logo 檔案。

### 3.2 角色定義

| 角色 | 權限範圍 |
|---|---|
| `super` | 建立/停用公司與部門、管理所有使用者、查看全部資料。 |
| `company_admin` | 等同該公司內的 super，可跨部門管理。 |
| `dept_admin` | 管理該部門內 `staff` 帳號；管理該部門客戶/商品/訂單/派車/單據列印。 |
| `staff` | 操作自己部門的客戶/商品/訂單，不能管理帳號；可執行會計檢視/對帳相關操作。 |
| `customer` | 外部店家；只能查看自己的訂單與使用自己的專屬商品下單。 |
| `guest` | OAuth2 首次登入後自動建立，待審核，無法操作資料。 |
| `developer` | 開發/測試用不受限制帳號；繞過 Casbin 與 RLS；僅於 `DEVELOPER_ACCOUNT_ENABLED=true` 時可登入（見 4.4），上架/上線時關閉。 |

以上七個角色為**系統預設值**，由 migration seed 建立（`is_system = true`，不可刪除、不可修改 `code` 與 `data_scope`）。`super` 可新增自訂角色，自訂角色必須指定資料範圍等級（`company` 或 `department`）；RLS 資料範圍依角色的 `data_scope` 決定，不依角色名稱。

### 3.3 後端授權（Casbin）

- Model：RBAC with domain，domain = `company_id`。
- Policy 儲存：PostgreSQL Casbin adapter。
- 範例：
  ```
  p, staff, {company}, /v1.SalesOrderService.ListSalesOrders, POST
  g, alice, staff, company_01
  ```
  Connect-RPC 的 obj 為 RPC method path、act 恆為 POST；REST 公開端點則為 URL path 與 HTTP method。
- Middleware 從 JWT/session 解析 `company_id`、`department_id`、`roles`，並注入 request context。
- 資料庫層防護：PostgreSQL RLS policy 使用 `SET LOCAL app.current_company_id`、`app.current_department_id`、`app.current_data_scope`、`app.current_customer_id`；資料範圍依當前角色的 `data_scope` 等級決定：`all` 繞過 RLS、`company` 限同公司、`department` 限同部門、`self` 僅本人資料（客戶帳號以 `app.current_customer_id` 比對 `customer_id`，僅見自己的訂單與主檔）。
- 職責切分：Casbin 僅負責「角色 × 公司（domain）× 資源 × 動作」的授權判斷；部門層級資料範圍（dept_admin / staff 僅限自己部門）不由 Casbin domain 承載，統一由 RLS policy 與 repository 查詢條件實現。
- 預設值：系統依 3.2 角色定義於 migration seed 預設 policy（p 規則），各角色對應的 API 權限即為預設值，可於 Web「API 權限設置」調整。
- 開發者繞過：角色為 `developer` 且 `DEVELOPER_ACCOUNT_ENABLED=true` 時，middleware 跳過 Casbin 檢查，RLS session 注入 `data_scope = all`（見 4.4）。
- 權限管理：`super` 可管理全域角色與所有 policy；`company_admin` 可管理自己公司 domain 的 policy 與自訂角色。系統需防止移除操作者自身角色的最後一個管理權限（避免鎖死），且所有權限異動寫入 `audit_logs`。

### 3.4 前端權限（CASL.js）

- Web 於登入後向後端 `AbilityService.GetAbility`（Connect-RPC）載入當前 ability JSON 並快取（短 TTL 60 秒，或權限異動後主動重新載入），不於每次路由切換請求；ability 由 `role_permissions` 表動態產生（預設值依 3.2 角色定義 seed），非前端靜態規則。
- 用於路由守衛與按鈕/選單顯示控制。
- App 因功能較少，使用簡單 role 判斷。

---

## 4. 認證機制

### 4.1 員工 / 管理員

- 採用 OAuth2 / OIDC，1.0 僅支援 Google Workspace；自建 IdP（Authelia/Authentik）延後至後續版本評估。
- Web：PKCE 流程，點選「以 Google 登入」後跳轉 provider，回調後後端建立 session 並設定 httpOnly cookie（Web 不使用 JWT，見 4.3）。
- App：使用 AppAuth / flutter_web_auth 開啟系統瀏覽器完成 OAuth2，回調後後端核發 JWT。
- 首次 OAuth2 登入若系統無此帳號，跳轉**註冊完成頁**：選擇所屬公司並輸入姓名後才算完成註冊，帳號建立為 `guest`（status = pending）並歸屬所選公司；未完成註冊完成頁前不建立帳號、不得進入系統。由 super 或該公司 company_admin 審核後指派部門與角色（dept_admin 僅能管理所屬部門內既有帳號，不審核 guest）。
- 註冊完成頁的公司清單僅回傳公開基本資訊（`display_name`），不暴露其他公司資料；App OAuth 登入新帳號同樣導向註冊完成頁（系統瀏覽器），完成後返回 App。
- 員工帳號不允許刪除，僅可停用。

### 4.2 客戶

- 使用 `customer_code` + 密碼登入。
- **客戶編號（登入帳號）編碼規則**：公司定義前綴 + 自增 ID（如 `TY000123`，序號補零至 6 位），建立時由系統依 `customer_counters` 樂觀鎖取號產生（同 6.5 訂單編號機制），公司內唯一，建立後不可修改；前綴由 `super` / `company_admin` 於公司設定維護（`companies.customer_code_prefix`，大寫英數 1–4 碼、全系統唯一，使 `customer_code` 可全域定位客戶帳號），修改前綴僅影響後續新編碼，不影響既有帳號，計數器不重置。
- 建立客戶時系統產生臨時密碼，首次登入強制修改；臨時密碼效期 **1 天（24 小時）**，過期需由 dept_admin 以上重新產生。
- 密碼強度：最少 **8 字元**。
- 登入錯誤鎖定：連續 **5 次**密碼錯誤鎖定帳號，30 分鐘後自動解除；dept_admin 以上可手動解鎖或重置密碼（重置連帶解除鎖定）。
- 忘記密碼由 dept_admin 以上重置。
- **登入帳號結構（一主多子）**：每個客戶綁定 **1 個主帳號**（建立客戶時自動產生的預設帳號，帳號名稱預設為客戶名稱，`is_primary = true`）+ 多個**子帳號**（如主廚）；各帳號獨立密碼、可獨立停用/重置。`customer_code` 為客戶身分識別，各帳號另有帳號名稱（顯示用）。任一帳號停用、登出或密碼重置不影響其他帳號，避免共用帳號離職時需要全員重設。**建立客戶時系統自動建立主帳號 + 1 個業務子帳號**（`is_primary = false`、帳號名稱預設為「客戶名稱（業務）」）；**業務子帳號專供該客戶的所屬業務（`default_sales_rep_id`）使用**，憑證交付所屬業務（業務以客戶身分登入 App，代客戶下單 / 查詢 / 退貨協助）；店家開箱僅持有主帳號（管理用），需下單時以主帳號自行建立子帳號。
- **主帳號僅供管理、不提供業務登入**：主帳號登入後僅能使用「帳號管理」功能（新增 / 停用 / 重置子帳號）；**所有業務 API**（下單、訂單歷史、退貨、專屬商品、促銷等）對主帳號 session 一律拒絕（403）。**子帳號是唯一的 App 業務登入身分**——老闆本人如需下單 / 查單，需另建一個子帳號。QR Code 登入的帳號選擇清單僅列子帳號（主帳號管理另有登入入口）。
- **店家自助管理帳號**：店家以**主帳號**於 App 自行管理自己客戶底下的帳號——新增子帳號（填寫帳號名稱，走臨時密碼流程）、停用子帳號、重置子帳號密碼。範圍僅限自己客戶（data_scope self），不得觸及其他客戶或員工帳號。**建立客戶時自動附帶的業務子帳號，於店家帳號管理清單顯示為「系統預設（業務使用）」並灰化（反白）**：店家可檢視但不可改名 / 停用 / 重置（店家無該帳號密碼，實質僅所屬業務可登入），後台可正常管理，並於客戶主責業務（`default_sales_rep_id`）變更時由後台移交新業務（重置密碼轉交）。防呆：**主帳號不可由店家停用或重置**（避免鎖死自己）；後台 dept_admin 以上仍可停用 / 重置任何店家帳號（含主帳號，重置主帳號密碼可轉交新負責人）作為逃生門；**後台停用主帳號視同店家帳號體系停用，連鎖停用全部子帳號**。帳號管理異動寫入稽核日誌。
- QR Code 深層連結含經簽章的一次性 token，token 內編碼 `company_id` + `customer_code` + 過期時間，後端驗證後帶入對應客戶身分，由使用者選擇任一**店家子帳號**完成登入（主帳號為管理用途、業務子帳號專供所屬業務使用，皆不列於 QR 清單）。
- 多公司環境下不可僅靠 `customer_code` 識別客戶，必須透過 token 中的 `company_id` 定位所屬公司。
- QR Code 深層連結採 Universal Link / App Link 形式（如 `https://<domain>/customer_account_qrcode/{token}`）：已安裝 App 直接開啟並完成登入；未安裝則導向商店下載頁。
- **帳號管理深層連結**：`https://<domain>/customer_account_manage`（Universal Link / App Link）：點擊開啟 App 並導向「帳號管理」登入流程（以主帳號帳密登入），未安裝 App 則導向商店下載頁；業務於新增客戶當場交付帳密時一併提供（可顯示為 QR Code）。連結本身不含登入憑證，登入仍須主帳號帳密。

### 4.3 Session / Token 儲存

- Web：httpOnly cookie + CSRF token；session store 使用 Valkey/Redis。
- App：`flutter_secure_storage` 存 access JWT 與 refresh token；Connect metadata 僅帶 `authorization: Bearer <jwt>`。
- JWT 效期：access token 1 小時；refresh token 30 天，採旋轉制（每次 refresh 核發新 refresh token 並作廢舊 token）。
- Token 撤銷：`users.token_version` 記錄 token 版本，JWT 與 refresh token 皆帶 `tv` claim，驗證時比對資料庫目前值；停用帳號、強制登出、密碼重置、角色變更時 `token_version + 1`，舊 token 立即失效。Web 端強制登出直接刪除 session。
- API access token：僅供 server-to-server 呼叫（內部排程、未來 agent 協定適配層），以獨立 header `X-Api-Token` 驗證，不配置於 Web / App 客戶端。
- 保留強制登出功能。

### 4.4 開發者帳號

- 內建第 7 個角色 `developer`（`is_system = true`、`data_scope = all`）：不受限制的開發/測試帳號，通過認證後繞過 Casbin 授權與 RLS 資料範圍，可跨公司/部門存取所有資料與功能。
- 環境開關：`DEVELOPER_ACCOUNT_ENABLED`（development 預設 `true`、production 預設 `false`）。開關關閉時 developer 帳號無法登入，middleware 亦不執行繞過。
- 啟動防護：後端啟動時若 `ENV=production` 且 `DEVELOPER_ACCOUNT_ENABLED=true`，拒絕啟動（fail fast）並輸出錯誤訊息；CI/CD 生產部署檢查清單必含「關閉開發者帳號」。
- 帳號來源：`super` 於管理人員名單將既有員工帳號指派 `developer` 角色；開發環境 migration 額外 seed 一組本機開發者帳號（僅 `ENV=development` 時建立）。
- 稽核：developer 帳號所有操作照常寫入 `audit_logs`，不因繞過權限而省略。
- 上架/上線：App 上架送審與 Big Bang 上線前，必須確認生產環境 `DEVELOPER_ACCOUNT_ENABLED=false`（見第 13 章）。

---

## 5. 資料模型

### 5.1 核心實體

| 實體 | 說明 | 租戶屬性 |
|---|---|---|
| `companies` | 公司主檔 | 系統級 |
| `departments` | 部門主檔，隸屬 company | `company_id` |
| `users` | 員工與客戶登入帳號（一個客戶可綁定多個） | `company_id`, `department_id` |
| `roles` | 角色定義（內建 6 角色 + 自訂角色） | 系統級 |
| `role_permissions` | 角色功能權限（resource × action），CASL ability 來源 | 關聯 `role_id` |
| `casbin_rules` | Casbin policy 儲存（p / g 規則），API 權限來源 | 系統級（domain 欄位承載公司） |
| `products` | 部門商品總表 | `department_id` |
| `product_units` | 商品單位與換算率（基本單位 + 多組換算） | 關聯 `product_id` |
| `customers` | 部門客戶總表 | `department_id` |
| `customer_addresses` | 客戶地址簿 | `customer_id`, `company_id`, `department_id` |
| `customer_contacts` | 客戶聯絡人 | `customer_id`, `company_id`, `department_id` |
| `customer_products` | 客戶專屬商品別名 | 關聯 `customer_id` |
| `promo_tags` | 促銷分類標籤（商品 / 專屬商品標記、客戶套用，見 10.1） | `department_id` |
| `return_requests` | 客戶退貨申請（客戶發起、業務確認） | `company_id`, `department_id` |
| `return_request_items` | 退貨申請品項（數量、原因、照片） | 關聯 `return_requests` |
| `sales_orders` | 銷售訂單主檔 | `department_id` |
| `sales_order_items` | 訂單明細 | 關聯 `sales_orders` |
| `sales_order_events` | 訂單狀態異動記錄（含原因） | `company_id`, `department_id`, `sales_order_id` |
| `order_counters` | 訂單編號計數器（樂觀鎖取號） | `company_id` |
| `customer_counters` | 客戶編號計數器（樂觀鎖取號） | `company_id` |
| `warehouses` | 倉別 | `department_id` |
| `routes` | 車次 | `department_id` |
| `cutting_specs` | 分切規格 + 加工/配送揀歸屬 | `department_id` |
| `product_categories` | 商品分類 | `department_id` |
| `notification_templates` | 通知範本 | `company_id`, `department_id` |
| `notifications` | 通知記錄 | `company_id`, `department_id`, `user_id` |
| `user_devices` | App 裝置與 FCM 推播 token | `user_id`, `company_id` |
| `audit_logs` | 稽核日誌 | `company_id`, `department_id`, `user_id` |
| `file_assets` | 檔案資產（Logo / 公告圖片 / 附件 / 列印 PDF） | `company_id`, `department_id` |
| `announcements` | 公告（Banner / 最新消息 / 圖文文章） | `company_id`, `department_id`（皆 NULL = 全系統公告） |
| `metadicts` | 字典檔（單位、付款方式、客戶類型等） | `department_id`（NULL = 系統預設） |
| `print_logs` | 正式單據列印記錄 | `company_id`, `department_id`, `user_id` |
| `print_previews` | 單據預覽記錄 | `company_id`, `department_id`, `user_id` |

### 5.2 關鍵欄位

- `companies`：`name`（公司名稱）、`tax_id`（統一編號）、`status`、`logo_url`、`primary_color`、`display_name`、`customer_code_prefix`（客戶編號前綴，全系統唯一，見 4.2）、`public_email`、`public_phone`、`public_address`、`terms_url`、`privacy_url`、`identifier`（穩定公司代號，全系統唯一，見第 17 章）、`capabilities`（JSON，對外協定與資源描述）、`public_info`（JSON，對外聯絡資訊）。
- `products`：`code`（部門內唯一商品編號）、`name`、`description`、`image_url`、`inventory_warehouse_id`（產品分庫）、`picking_warehouse_id`（揀貨倉別）、`sale_unit`、`max_qty`、`category_id`、`promo_tag_ids`（促銷分類標籤，JSON 陣列，見 10.1）、`is_active`；可用分切規格以關聯表 `product_cutting_specs`（`product_id` + `cutting_spec_id`）維護，不使用陣列欄位。
- `product_units`：`product_id`、`unit_code`（對應 metadicts 單位）、`is_base`（是否基本單位，每商品僅一個）、`conversion_rate`（換算為基本單位的比率，如 1 條 = 0.6 kg 則為 0.6）、`sort_order`。
- `customer_products`：`customer_id`、`product_id`、`alias_name`、`default_qty`、`default_cut_note`、`promo_tag_ids`（促銷分類標籤，JSON 陣列，見 10.1）；`customer_id + product_id` 唯一（搭配 5.4 軟刪除部分唯一索引）；不存單價。
- `customers`：`customer_code`（客戶編號：公司定義前綴 + 自增 ID，`company_id + customer_code` 唯一，建立後不可修改，編碼規則見 4.2）、`payment_method_id`、`settlement_method_id`、`customer_type_id`、`tax_id`、`invoice_type_id`、`default_sales_rep_id`（預設負責業務，通知路由用，見 6.2）、`preferred_delivery_days`（偏好送貨日：星期一到六核取，JSON 布林陣列，見 6.2）、`promo_tag_ids`（套用促銷分類，JSON 陣列，見 10.1）。
- `customer_addresses`：`customer_id`、`type`（`shipping` / `billing` / `other`）、`address`、`is_default`。
- `customer_contacts`：`customer_id`、`name`、`title`、`email`、`phone`、`is_default`。
- `sales_orders`：`order_no`（訂單編號：下單來源碼 + 自增序號，公司內唯一，見 6.5）、`customer_id`、`status`（`pending` / `processing` / `completed` / `cancelled` / `voided`）、`expected_delivery_date`（預計出貨日，派車看板日期篩選依據）、`sales_rep_id`（負責業務）、`created_by`（建立人）、`route_id`（車次）、`delivery_sequence`（同車次配送順序）、`dispatched_at`（派車確認時間）、`dispatched_by`（派車確認人）、`source`（訂單來源）、`notes`、`version`（樂觀鎖版本號，派車拖放併發控制，見 7.1）；**不含任何金額欄位**（見 6.5）。
- `order_counters`：`company_id`、`source`（訂單來源碼）、`next_seq`、`version`（樂觀鎖版本號）。
- `customer_counters`：`company_id`、`next_seq`、`version`（樂觀鎖版本號）。
- `sales_order_events`：`sales_order_id`、`event_type`（`create` / `edit` / `dispatch` / `dispatch_cancel` / `cancel` / `complete` / `void`）、`reason`、`actor_id`、`payload`、`created_at`。
- `sales_order_items`：`product_id`、`customer_product_id`、`qty`、`unit`、`base_qty`（換算為基本單位的數量，供揀貨/加工單彙總）、`cutting_spec_id`、`special_cut_note`、`warehouse_id`（無單價與金額欄位）。
- `users`：`status`（`active` / `inactive` / `pending`）、`phone`、`employee_no`、`account_name`（客戶登入帳號名稱，客戶內唯一，見 4.2）、`is_primary`（客戶主帳號標記，每客戶恰一個，部分唯一索引 `(customer_id) WHERE is_primary AND deleted_at IS NULL`；主帳號僅供管理、不提供業務登入，見 4.2）、`password_hash`（僅客戶帳號使用；員工走 OAuth2，不存密碼）、`oauth_provider` / `oauth_subject`（員工 OAuth2 識別）、`customer_id`（客戶登入帳號關聯客戶主檔，員工為 NULL；一個客戶可綁定多個帳號，見 4.2）、`token_version`（token 版本，撤銷 JWT / refresh token 用，見 4.3）、`temp_password_expires_at`（臨時密碼效期）、`failed_login_attempts` / `locked_at`（登入錯誤鎖定，見 4.2）。
- `roles`：`code`（`super` / `company_admin` / `dept_admin` / `staff` / `customer` / `guest` / 自訂）、`name`、`data_scope`（`all` / `company` / `department` / `self`）、`is_system`、`is_active`。
- `role_permissions`：`role_id`、`resource`（如 `order` / `customer` / `product` / `dispatch` / `print` / `user` / `company`）、`action`（`create` / `read` / `update` / `delete` / `manage` / `print` / `dispatch` 等）。
- `notification_templates`：`company_id`、`department_id`、`code`、`name`、`channel`（`fcm` / `in_app`，1.0 無 Email）、`subject`、`body`、`locale`、`is_active`。
- `notifications`：`company_id`、`department_id`、`user_id`、`template_id`、`channel`、`title`、`content`、`payload`、`status`（`pending` / `sent` / `failed` / `read`）、`sent_at`、`read_at`。
- `audit_logs`：`company_id`、`department_id`、`user_id`、`action`（`create` / `update` / `delete` / `login` / `logout` / `print` / `force_logout` / `role_change` / `dispatch_cancel` / `void`）、`resource_type`、`resource_id`、`before_snapshot`、`after_snapshot`、`ip_address`、`user_agent`。
- `file_assets`：`company_id`、`department_id`、`owner_type`、`owner_id`、`filename`、`original_filename`、`mime_type`、`size_bytes`、`storage_path`、`url`、`created_by`。
- `print_logs`：`company_id`、`department_id`、`document_type`（`dispatch_summary` / `delivery_note` / `picking_list` / `processing_list`）、`route_id`、`customer_id`、`warehouse_id`、`target_date`、`is_reprint`、`reprint_reason`、`printed_by`、`printed_at`、`file_asset_id`。
- `print_previews`：`company_id`、`department_id`、`document_type`、`route_id`、`customer_id`、`warehouse_id`、`target_date`、`previewed_by`、`previewed_at`、`file_asset_id`。
- `announcements`：`type`（`banner` / `news` / `article`）、`title`、`content`、`image_url`、`link_url`、`publish_at`、`unpublish_at`、`sort_order`、`is_active`。
- `user_devices`：`user_id`、`platform`（`android` / `ios`）、`fcm_token`、`device_name`、`last_seen_at`。

### 5.3 字典檔策略

| 字典 | 層級 |
|---|---|
| 倉別、車次、分切規格、商品分類 | 部門級（獨立實體表，見 5.1，非 metadicts） |
| 單位、付款方式、結帳方式、客戶類型、發票類型、規格、代切實重肉片 | 系統級預設 + 部門可擴充 |
| 訂單來源 | 系統級固定 |

「系統級預設 + 部門可擴充」採用單一 `metadicts` 表實作：

- `metadicts` 欄位：`type`（字典類型）、`code`（程式識別）、`display_name`（顯示名稱）、`department_id`（NULL = 系統預設）、`sort_order`、`is_active`、`deleted_at`。
- 系統預設：`department_id IS NULL`，所有部門皆可見。
- 部門擴充：`department_id = <部門 ID>`，僅該部門可見。
- 查詢邏輯：取 `department_id IS NULL OR department_id = 當前部門` 的資料，並排除 `deleted_at IS NOT NULL`。
- 唯一性：`type + code + department_id` 唯一；系統預設與部門擴充可共用 `code`。PostgreSQL 實作需以 `NULLS NOT DISTINCT`（PG15+）或部分唯一索引處理 `department_id IS NULL` 的系統預設列。
- 權限：`super` 可維護系統級預設；`dept_admin` / `staff` 可維護自己部門的擴充值。1.0 不支援部門覆寫或停用系統預設。

### 5.4 軟刪除慣例

- 業務實體統一使用 `deleted_at TIMESTAMP NULL` 作為軟刪除標記；`NULL` 表示未刪除，非 `NULL` 表示刪除時間。
- 適用軟刪除的實體：`companies`、`departments`、`users`、`products`、`customers`、`customer_addresses`、`customer_contacts`、`customer_products`、`sales_orders`、`sales_order_items`、`warehouses`、`routes`、`cutting_specs`、`product_categories`、`notification_templates`、`file_assets`、`metadicts`、`announcements`、`roles`。
- `roles` 中內建角色（`is_system = true`）不可刪除；軟刪除僅適用自訂角色。
- `users` 中的員工帳號不刪除、僅停用（見 4.1）；軟刪除僅適用客戶帳號。
- `customer_products` 直接使用 `deleted_at`；1.0 為全新系統且不匯入舊資料（見第 13 章），不保留舊系統的 `is_deleted` 相容欄位。
- 業務唯一性限制（如 `products.code`、`customers.customer_code`、`customer_products` 的客戶+商品）一律搭配部分唯一索引（`WHERE deleted_at IS NULL`）實作，避免軟刪除後無法重建同名資料。
- 不適用軟刪除：`notifications`（通知記錄不可刪除）、`audit_logs`（稽核日誌不可刪除）、`sales_order_events`（訂單異動記錄不可刪除）。
- 列表查詢預設排除 `deleted_at IS NOT NULL` 的資料；管理員介面可提供「顯示已刪除」開關。
- 軟刪除實體如需復原，僅需清空 `deleted_at` 並寫入稽核日誌；硬刪除僅限 super 在特定管理介面執行，並強制留存稽核紀錄。
- 主檔軟刪除後，歷史訂單與單據仍可正常顯示其關聯資料（名稱、編號），僅不可再選用於新單據。

---

## 6. 訂單流程

### 6.1 訂單狀態

```text
待處理 ⇄ 處理中 → 已完成
  ↓                 ↓
已取消           已作廢（終態）
```

- 待處理：可編輯、可取消。
- 處理中：已派車確認，不可編輯。此階段涵蓋「已派車」、「已列印」等中間進度，實際細節由 `dispatched_at`、列印記錄等欄位追蹤。
- **取消派車（處理中 → 待處理）**：派錯車時 `dept_admin` 以上可退回待處理，需填寫原因；清除 `dispatched_at` / `dispatched_by`，保留 `route_id` / `delivery_sequence`（訂單停留看板原位，待重新派車）。若該車次已有正式列印記錄，操作時須提示需重新列印。
- 已完成：管理人員標記出貨完成，不可編輯。
- **作廢（已完成 → 已作廢）**：內容錯誤時，`dept_admin` 以上可作廢，需填寫原因；`voided` 為終態不可恢復。1.0 不提供已完成訂單的直接編輯，更正方式為作廢後重新建立訂單（新單備註原單號）。
- 所有狀態異動（建立、編輯、派車、取消派車、取消、完成、作廢）皆寫入 `sales_order_events`（含操作人、時間、原因）；取消派車與作廢另寫入 `audit_logs`。

### 6.2 下單流程

1. 選擇客戶（業務）或自動為自己（客戶）。
2. 帶出 `customer_products`（客戶只能選此清單；業務可選商品總表或手打）。
3. 業務手打商品名稱時，詢問是否儲存為該客戶別名。
4. 輸入數量、單位、分切規格、特殊切法備註（系統不儲存金額與單價）。
5. 選擇/建議預計出貨日；客戶設有偏好送貨日（`preferred_delivery_days`，見 5.2）時，若選擇日非勾選日則自動順延至下一個有勾選的日期。
6. 提交後發送訂單建立通知（FCM + 站內，1.0 無 Email）：業務下單時推播給該客戶的全部子帳號（主帳號為管理用途、不接收業務通知）；客戶自行下單不另行通知。

### 6.3 客戶專屬商品清單

- 首次由業務在客戶檔案頁手動建立，或下單時自動加入。
- 可手動刪除（軟刪除）。
- 數量設為 0 時保留在清單，但後續單據不顯示。
- 一個客戶對一個總表商品只能有一個別名。
- `customer_products` 可設定預設數量、預設切法備註、促銷分類標籤（`promo_tag_ids`，見 10.1）；不存單價。

### 6.4 單位換算

- 商品以 `product_units` 設定「基本單位」（`is_base = true`）與多組換算率（如 1 條 = 0.6 kg）。
- 下單時選擇單位後自動換算數量，並於 `sales_order_items.base_qty` 保存換算為基本單位的數量，供揀貨單與加工單彙總（系統不計算金額）。

### 6.5 訂單編號

- **訂單編號**：`下單來源碼 + 自增序號`（如 `W000123`，序號補零至 6 位），公司內唯一。來源碼對應系統級字典「訂單來源」（如 `W` = Web 中台、`A` = App）。
- **併發取號**：`order_counters` 以樂觀鎖更新（`UPDATE ... SET next_seq = next_seq + 1, version = version + 1 WHERE company_id = ? AND source = ? AND version = ?`，衝突時重試），取號與訂單建立於同一資料庫交易。
- **金額**：1.0 全系統**不儲存任何金額欄位**（訂單、明細、商品、專屬商品皆無單價 / 小計 / 稅額 / 折扣；`sales_orders` 與 `sales_order_items` 不含金額欄位），單據與報表不含金額。後續版本如需金額，另行擴充資料模型。
- **型別與時區**：時間欄位一律 `timestamptz`，顯示時區 UTC+8。

### 6.6 退貨流程（客戶發起）

- 客戶從 App 發起退貨申請，品項來源擇一（1.0 兩者並存，實際 UX 依試用回饋調整）：
  - 從**歷史訂單**勾選品項。
  - 從**客戶專屬商品清單**選擇品項。
- 填寫數量、原因、備註，並上傳照片（可多張，經 `file_assets` 留存）。
- 送出後狀態 `pending`，由該客戶 `default_sales_rep_id`（未設定則部門 `dept_admin`）於 App / Web 審核：通過（`approved`）或拒絕（`rejected`，需填原因）。
- 客戶端顯示審核結果；通過時顯示**退貨證明畫面**，可出示給配送司機。
- 公司端**不建置配送專屬頁面**：退貨取貨由業務自行通知配送，系統僅記錄申請與審核狀態。
- 退貨不修改原訂單狀態；申請與審核異動寫入稽核日誌，審核結果推播客戶帳號（見 10.1）。

---

## 7. 派車與單據列印

### 7.1 派車看板

- Web 專屬頁面，Kanban 看板：車次為欄位，訂單為卡片。
- 每筆訂單透過 `sales_orders.route_id` 關聯車次；`sales_orders.delivery_sequence` 決定同車次內的配送順序。
- 支援拖放調整訂單所屬車次與車內順序。拖放提交採**樂觀鎖**：比對 `sales_orders.version`，衝突則拒絕並重新整理看板（配合 Connect 串流即時推播他人變更）。
- 即時更新：Connect-RPC server streaming（`DispatchService.WatchBoard`）推送型別化事件；斷線自動重連，連續失敗降級為 30 秒輪詢。
- 日期篩選：依預計出貨日查看。
- 派車確認以**車次批次**執行：一次確認該車次當日所有待派訂單，批次內每筆訂單各自記錄 `dispatched_at` / `dispatched_by`，訂單狀態變為「處理中」。
- 取消派車：`dept_admin` 以上可將處理中訂單退回待處理（需填寫原因，見 6.1）；若該車次已有正式列印記錄，操作時提示需重新列印。
- 1.0 一筆訂單僅對應一個車次，不支援拆單至多車次。

### 7.2 單據列印

在派車規劃頁面統一觸發，隨時可預覽，正式列印需確認派車。

| 單據 | 內容 | 排序/分組 |
|---|---|---|
| 單車總表 | 各車次的所有店家，含各店家明細摘要（系統不存金額，單據無金額欄位） | 依車次 |
| 對點單 | 每個店家一張 A4，列出該店家訂單明細（品名、數量、單位） | 依車次 → 店家 |
| 揀貨單 | 依車次與倉別印製 | 車次 → 倉別 → 商品分類 → 商品名稱 |
| 加工單 | 要切的東西，分為「加工室揀」與「配送揀」兩區塊；顯示原始數量，「加工後數量」欄位列印為空白，由加工人員加工時手寫回填（1.0 不回寫系統） | 先加工室揀，再配送揀 |

### 列印流程

1. **預覽**：任何狀態皆可預覽，產生 PDF 並寫入 `print_previews`，不影響正式列印記錄。
2. **正式列印**：訂單狀態需為「處理中」（已派車確認）。產生 PDF 並寫入 `print_logs`，`is_reprint = false`。
3. **重印**：已有正式列印記錄後，`dept_admin` / `staff` 皆可重印，但需填寫 `reprint_reason`。產生新 PDF 並寫入 `print_logs`，`is_reprint = true`。
4. **PDF 留存**：每次正式列印與預覽產生的 PDF 皆關聯 `file_assets`，保留當時列印內容供日後查核。

- 技術：Gotenberg 將 HTML/CSS 模板轉為 PDF，字型使用 Noto Sans CJK TC。
- 紙張：A4。
- 空表不印出。
- 列印與重印皆記錄時間、操作人與原因。

---

## 8. Web 中台功能

### 8.1 導覽

- 登入後選擇部門（super/company_admin）。
- 側邊欄依角色動態群組，並顯示當前公司 Logo、主色與公開資訊。
- 部門切換後重新整理當前頁面資料，不跳轉。

### 8.2 頁面清單

| 頁面 | 說明 | 適用角色 |
|---|---|---|
| Dashboard | 今日待出貨、待處理訂單數量、快速連結 | 全部 |
| 公司管理 | CRUD 公司 | super |
| 公司識別設定 | 上傳 Logo、設定主色、公開資訊 | super / company_admin |
| 部門管理 | CRUD 部門 | super |
| 管理人員名單 | 使用者 CRUD、角色指派、停用、強制登出 | super / company_admin / dept_admin（依範圍） |
| 角色權限設置 | 角色 CRUD（自訂角色）、功能權限（resource × action）編輯；內建角色與權限為預設值 | super / company_admin（限自己公司） |
| API 權限設置 | Casbin policy CRUD（p 規則：角色 × domain × path × method）；預設值依 3.2 seed | super / company_admin（限自己公司） |
| 客戶總表 | 部門客戶 CRUD、地址簿、聯絡人、QR Code | dept_admin / staff |
| 商品總表 | 部門商品 CRUD、單位換算、庫別、分切規格 | dept_admin / staff |
| 訂單管理 | 訂單列表、新增/編輯/檢視 | dept_admin / staff |
| 派車規劃 | Kanban 看板、拖放、列印 | dept_admin / staff |
| 字典檔管理 | 部門級與系統級字典維護 | super / dept_admin（依字典） |
| 公告管理 | Banner / 最新消息 / 圖文文章 | super / company_admin / dept_admin |
| 通知中心 | 系統通知列表 | 全部 |
| 系統設定 | 主題、版本資訊 | 全部 |

### 8.3 互動元件

- DataTable：分頁、排序、篩選、欄位隱藏、批次操作。
- Sheet 表單：側邊滑出新增/編輯。
- Cmd+K 搜尋：預留，第一版不實作。
- 通知鈴：實作通知中心。
- 主題切換：Light / Dark / System。

---

## 9. App 功能

### 9.1 登入與首頁

- 身分選擇：「我是店家」或「我是業務」。
- 業務登入後固定所屬部門，不選擇部門。
- 首頁：公告輪播 / 最新消息（含促銷訊息）為主，並依所屬公司顯示對應 Logo 與主色；快速下單入口於底部導覽「商品」頁，首頁不另放下單入口。

### 9.2 底部導覽

- 首頁 / 商品（快速下單） / 訂單歷史 / 功能

### 9.3 功能頁

- 客戶快速查詢、QR Code、關於我們（顯示公司公開資訊）、隱私權政策、設定、登出；店家另有「帳號管理」頁（新增 / 停用 / 重置名下登入帳號，見 4.2）。

### 9.4 業務流程

- 從底部導覽「商品」頁進入快速下單流程。
- 選擇客戶 → 帶出 `customer_products` → 可選總表商品或手打 → 輸入數量/單位/規格/備註 → 提交訂單。
- 可新增客戶（手動表單）：填寫客戶主檔、地址與聯絡人，送出後系統自動產生**主帳號與業務子帳號**（各含帳號名稱與臨時密碼，24 小時效期，見 4.2；`customer_code` = 公司定義前綴 + 自增 ID），成功頁顯示主帳號帳密（供業務轉交店家）、業務子帳號帳密（供所屬業務留存使用）與**管理網址鏈結**（店家帳號管理深層連結，可顯示為 QR Code，見 4.2）。
- 可刪除訂單明細品項。
- 可管理客戶專屬商品清單（新增/刪除/編輯別名）。

### 9.5 客戶流程

- 從底部導覽「商品」頁進入快速下單流程。
- 只能從自己的 `customer_products` 下單。
- 查看自己的訂單歷史。

### 9.6 離線與快取

- 快取登入資訊、客戶列表、商品列表、訂單歷史。
- 下單仍需連網。
- 下拉重新整理：首頁、訂單歷史、客戶列表、商品列表。

---

## 10. 通知與日誌

### 10.1 通知

通知系統由 `notification_templates` 與 `notifications` 兩個實體支撐：

- `notification_templates`：依公司 / 部門維護通知範本，支援 `fcm`、`in_app` 兩種通道（**1.0 不使用 Email**），內文可使用 `{{變數}}` 模板語法。
- `notifications`：實際發送記錄，包含標題、內容、payload、發送狀態與已讀時間。

通知通道：

- App 推播：FCM，用於訂單建立、審核結果、促銷推播。
- Web / App 通知中心（站內）：顯示系統通知。

自動通知路由（1.0）：

- **訂單建立**：業務下單 → 推播該客戶的全部**子帳號**（主帳號為管理用途、不接收業務通知）；客戶自行下單不另行通知。
- **客戶專屬商品新增**（後台 / Web 建立時）→ 推播該客戶 `default_sales_rep_id` 主責業務（未設定則部門 `dept_admin`；是否需要業務檢核，待定）。
- **退貨審核結果** → 推播發起退貨的客戶帳號。

促銷推播（與公告分離，是兩件事）：

- 公告（`announcements`）顯示於 App 首頁消息區，所有客戶可見，不等同推播。
- 促銷推播：商品 / 客戶專屬商品可標記促銷分類標籤（`promo_tags`）；客戶資料可套用（訂閱）分類（`promo_tag_ids`）；後台依分類選擇要推播的客戶群，推播至該客戶群的**子帳號**，經 FCM + 站內發送。

發送失敗處理：1.0 **不重試**，僅將 `notifications.status` 標記為 `failed` 並記錄原因；重試佇列列後續版本評估。FCM 回報裝置 token 失效（unregistered / invalid）時，後端直接刪除對應 `user_devices` 記錄，避免持續發送無效推播。

### 10.2 稽核日誌

- 記錄關鍵操作：登入、下單、修改主檔、刪除、列印、強制登出、角色變更。
- 記錄操作人、時間、操作類型、異動前後摘要。
- 稽核日誌存放於獨立的 `audit_logs` 表，與業務資料實體分開儲存，便於日後依時間分區或封存。
- 寫入方式：與業務操作同一資料庫交易**同步寫入**（操作與稽核同成功或同失敗），不採非同步佇列。
- **保留期限**：`audit_logs` 預設保留 **3 個月**，由 `super` / `company_admin` 於管理頁面設定（1 / 3 / 6 / 12 個月或永久）；到期由保留排程刪除。

---

## 11. 測試策略

- 後端：Go testify + dockertest，單元測試與整合測試。
- 前端：Vitest + jsdom + @solidjs/testing-library。
- App：flutter_test。
- 三端單元測試覆蓋率門檻 **70%**，CI 強制檢查；關鍵路徑（認證、授權 / RLS、訂單狀態機、樂觀鎖取號、退貨審核、偏好送貨日順延）需有整合測試。
- Maestro 整合測試用於 App。

---

## 12. 部署與維運

### 12.1 開發環境

- docker-compose 啟動 PostgreSQL + Valkey + Gotenberg + Mailpit（1.0 不含自建 IdP；OAuth2 使用 Google Workspace 測試 client）。
- 檔案儲存：1.0 採用本地掛載（NFS / 容器 volume），透過 `file_assets.storage_path` 記錄實際路徑，並由後端提供下載 URL。
- Task 指令管理 Go / Flutter 原生任務。
- pnpm workspace + Turborepo 管理全倉任務管線與快取（含前端建置、proto 型別同步、測試）。

### 12.2 生產環境

- Kubernetes + Traefik ingress。
- PostgreSQL 與 Valkey 皆容器化部署於 k8s（StatefulSet）。
- GitHub Actions CI/CD：測試、建置 image、部署至 k8s。

### 12.3 備份、監控與災難復原

#### 12.3.1 備份策略

| 資料 | 備份方式 | 頻率 | 保留期限 | 儲存位置 |
|---|---|---|---|---|
| PostgreSQL | 每日完整備份 + WAL 歸檔（PITR） | 每日 | 每日備份 30 天，每月備份 12 個月 | Google Cloud Storage |
| Valkey / Redis | RDB 快照 + AOF | 每小時 RDB，持續 AOF | 7 天 | Google Cloud Storage |
| 本地檔案儲存（NFS / volume） | restic 到異地備份空間 | 每日 | 30 天 | Google Cloud Storage |
| Kubernetes manifests / 設定 | Git 版控 + 加密 secrets 匯出 | 每次變更 | 永久 | Git + Google Cloud Storage |

#### 12.3.2 監控與告警

- **基礎設施監控**：CPU、記憶體、磁碟、網路、Pod 重啟次數。
- **應用程式指標**：API 延遲、錯誤率、請求量、Connect-RPC 狀態碼。
- **業務指標**：登入失敗次數、訂單建立量、列印次數、稽核日誌異常查詢。
- **日誌收集**：結構化 JSON 日誌集中收集，保留 30 天。
- **告警通道**：Email / Slack / Webhook，告警條件包括服務不可用、錯誤率驟升、備份失敗、磁碟空間不足。
- **監控工具**：1.0 採用 Prometheus + Grafana + Alertmanager，與應用程式同叢集部署。

#### 12.3.3 災難復原

- **RTO（復原時間目標）**：4 小時。
- **RPO（復原點目標）**：1 小時（依賴 WAL 歸檔）。
- **復原流程**：
  1. 確認災難範圍與備份可用性。
  2. 從 Google Cloud Storage 最新備份還原 PostgreSQL 與 Valkey。
  3. 重新部署 k8s 服務與 ingress。
  4. 驗證應用程式健康狀態與關鍵業務流程。
  5. 切換 DNS 或調整負載均衡器。
- **演練**：每半年執行一次災難復原演練，並更新復原手冊；PostgreSQL 與 Valkey 採自建 StatefulSet，**上線前必須完成首次 PITR（WAL 歸檔）還原演練**，驗證 RTO 4 小時 / RPO 1 小時可達成。

---

## 13. 資料遷移與上線過渡

- 第一版不從現有 sales-order 系統匯入任何資料，客戶、商品等主檔與訂單皆於新系統重新建檔。
- 無試點、無新舊系統並行運作、無舊系統唯讀維護期。
- 新系統完成開發與測試後，直接全面上線並廢止舊系統。
- 上線檢查清單必含：確認生產環境 `DEVELOPER_ACCOUNT_ENABLED=false`（開發者帳號關閉，見 4.4）。
- 後續如需要從舊系統匯出主檔的 migration script，將另行評估，不納入 1.0 範圍。

---

## 14. 安全與合規

### 14.1 資料保護

- **個資範圍**：客戶名稱、電話、地址、統一編號、聯絡人資料、員工帳號與 OAuth2 識別資訊皆視為個人資料。
- **最小權限**：資料存取嚴格依角色與部門隔離；`super` 跨公司查詢需記錄於 `audit_logs`。
- **資料保留**：客戶與訂單資料保留期限依公司內規；軟刪除後保留 30 天，之後由 super 執行硬刪除或封存。記錄類資料保留期限：

  | 資料 | 保留期限 | 到期處理 |
  |---|---|---|
  | `notifications` | 180 天 | 排程刪除 |
  | `audit_logs` | 預設 3 個月（管理頁可設定 1 / 3 / 6 / 12 個月或永久） | 到期排程刪除 |
  | `print_logs`（含關聯 PDF） | 2 年 | 排程刪除 |
  | `print_previews`（含關聯 PDF） | 90 天 | 排程刪除 |
- **當事人權利**：1.0 提供客戶資料查詢、更正、刪除機制，由 `super` / `company_admin` 於中台操作。

### 14.2 傳輸與儲存安全

- **傳輸加密**：所有 API、Connect-RPC（含 server streaming）皆強制 TLS 1.2+；生產環境不使用 `InsecureSkipVerify`。
- **靜態加密**：PostgreSQL 與檔案儲存使用磁區加密；App JWT 儲存於 `flutter_secure_storage`。
- **密碼儲存**：客戶密碼使用 Argon2id 雜湊；員工採用 OAuth2，後端不儲存密碼。
- **Session 安全**：httpOnly cookie + CSRF token；Valkey session store 啟用 TTL。

### 14.3 應用程式安全

- **注入防護**：使用 Ent ORM 與參數化查詢，避免 SQL 注入。
- **XSS 防護**：前端輸出跳脫；Rich Text 內容需經過消毒（sanitize）。
- **CSRF 防護**：Web 中台非讀取操作需帶入 CSRF token。
- **Rate Limiting**：登入、QR Code 兌換、密碼重置等端點實作 rate limit。
- **檔案上傳**：MIME 白名單（圖片 `image/jpeg` / `image/png` / `image/webp` ≤ 5 MB；文件 `application/pdf` ≤ 10 MB），副檔名與 magic bytes 雙重檢查，白名單以外拒絕。
- **病毒掃描**：1.0 不實作上傳檔案病毒掃描（僅內部使用者上傳），後續版本評估。
- **安全標頭**：生產環境啟用 HSTS、CSP、X-Frame-Options 等標頭。
- **依賴安全**：CI 中執行 `govulncheck`、`npm audit`、Flutter 依賴掃描或 Dependabot。

### 14.4 稽核與監控

- 關鍵操作皆寫入 `audit_logs`（登入、下單、修改主檔、刪除、列印、強制登出、角色變更）。
- 生產環境整合日誌收集與異常告警（如大量登入失敗、非預期跨部門查詢）。

---

## 15. 風險與假設

| 風險 | 因應 |
|---|---|
| Connect-RPC + REST 雙協定增加維護成本 | 業務 API 統一 Connect-RPC（Web / App），REST 僅保留公開端點；proto 為唯一 API 來源。 |
| RLS 與 Ent 整合複雜 | 使用 connection hook 統一注入 session variables，並寫整合測試驗證。 |
| OAuth2 首次登入待審核可能影響上線 | 上線前由管理員預先匯入/建立員工帳號。 |
| 單據模板需求變動頻繁 | HTML/CSS 模板易於調整，列印服務獨立。 |
| 自建 PostgreSQL / Valkey（k8s StatefulSet）維運與災難復原複雜 | 每日備份 + WAL 歸檔；上線前完成 PITR 還原演練；Prometheus 監控告警；每半年 DR 演練。 |

---

## 16. 後續步驟

1. 確認本設計規格。
2. 使用 `writing-plans` skill 產生詳細實作計畫與里程碑。
3. 依計畫分階段實作：權限/認證 → 主檔 → 訂單 → 派車/列印 → App → 部署。
4. 1.1 獨立迭代：App AI 輔助功能（拍照建客戶、業務語音下訂單），見 `2026-07-18-app-ai-assist-1.1-design.md`。

---

## 17. AI 整合備忘錄（MCP-SERVER / ACP / A2A）

> 本節為未來導入 AI Agent 整合的設計預留與備忘，第一版不強制實作，但核心資料結構與 API 需預留擴充點。

### 17.1 預留目標

- 讓外部 AI Agent 能透過標準協定發現（discover）各公司的公開資訊與可用能力。
- 讓本系統未來可同時作為 **MCP Server**、**ACP Server** 或 **A2A Agent** 的服務端/對等端。
- 不預先綁定單一協定，以 `capabilities` 與 `public_info` 欄位描述能力，由協定適配層轉譯。

### 17.2 協定簡述

| 協定 | 定位 | 與本系統的關係 |
|---|---|---|
| **MCP-SERVER**（Model Context Protocol） | Anthropic 主導的「工具/資源/提示」標準，讓 LLM 透過統一介面呼叫外部能力。 | 本系統的 `companies/{identifier}/public` 與各業務 API 可包裝為 MCP resources/tools。 |
| **ACP**（Agent Communication Protocol / Agent Connect Protocol） | Agent 之間的連線與能力交換協定，強調安全發現與會話。 | 公司主檔的 `capabilities` 可作為 ACP 能力清單（capability manifest）的資料來源。 |
| **A2A**（Agent-to-Agent，Google） | Agent 之間協作協定，聚焦任務委派與狀態追蹤。 | 訂單、派車等長時間任務未來可透過 A2A 任務物件暴露給外部 Agent。 |

### 17.3 預留設計

1. **公司層級能力描述**
   - `companies.capabilities` 記錄該公司對外提供的協定清單（`protocols`）、API 基底網址（`api_base_url`）、資源清單（`resources`）。
   - `companies.public_info` 記錄對外聯絡資訊，供 Agent 在無認證情境下識別服務主體。

2. **穩定資源識別**
   - 以 `companies.identifier` 作為跨協定的穩定公司代號。
   - 各業務資源（客戶、商品、訂單）以 `company_id` + `department_id` + 自身 ID 組成全域唯一識別。

3. **公開發現端點**
   - 第一版預留 `GET /api/v1/companies/public/{identifier}`，無需認證，回傳該公司公開資訊與 capabilities。
   - 未來可擴充為 `/api/v1/.well-known/agent-manifest` 或各協定專屬 discovery endpoint。

4. **協定適配層（未來）**
   - 不將 MCP/ACP/A2A 邏輯直接寫入業務 domain，而是新增 `internal/agent` 或獨立 `mcp-server` 服務。
   - 適配層透過 Connect-RPC/REST 呼叫現有 domain usecase，避免業務邏輯與協定耦合。

### 17.4 第一版限制

- 第一版僅建立 `companies.capabilities` / `public_info` 欄位與公開發現端點，不實作完整 MCP/ACP/A2A server。
- 不對外公開需認證的業務資源；Agent 整合需在後續版本補上 OAuth2 / API token + 權限控管後才開放。

---

## 18. 修訂記錄

| 修訂號 | 日期 | 修訂內容 | 修訂者 |
|---|---|---|---|
| v1.0.0 | 2026-07-17 | 初版發布；產品名稱定為「多公司訂出貨系統 1.0」；新增公司識別與公開資訊需求。 | — |
| v1.0.1 | 2026-07-17 | 移除設計規格中的外部系統相關用語，統一以「外部系統」描述。 | — |
| v1.0.2 | 2026-07-17 | 新增第 17 節「AI 整合備忘錄」，納入 MCP-SERVER、ACP、A2A 的未來整合預留。 | — |
| v1.0.3 | 2026-07-17 | 確認 1.0 採用全新 monorepo 開發策略，統一規格書與實作計畫的專案結構。 | — |
| v1.0.4 | 2026-07-17 | 補充資料模型：客戶地址簿/聯絡人、通知系統、稽核日誌、檔案資產；擴充訂單與商品欄位；確認本地檔案儲存與稽核日誌獨立儲存。 | — |
| v1.0.5 | 2026-07-17 | 會計角色由 `staff` 兼任；訂單狀態機維持三態並以欄位追蹤派車/列印細節；新增 API 分頁格式標準（`meta`）。 | — |
| v1.0.6 | 2026-07-17 | QR Code 深層連結改以簽章 token 內含 `company_id` + `customer_code`；新增第 5.4 節統一軟刪除慣例（`deleted_at`）。 | — |
| v1.0.7 | 2026-07-17 | 新增 `metadicts` 核心實體；補充系統級字典「部門可擴充」的具體機制（單一表 + `department_id` NULL = 系統預設）。 | — |
| v1.0.8 | 2026-07-17 | 新增 `print_logs` / `print_previews` 實體；補充單據列印流程、重印權限（開放 `staff`）、PDF 留存與預覽記錄機制。 | — |
| v1.0.9 | 2026-07-17 | 明確 App 底部導覽「商品」頁為快速下單入口，並同步更新業務/客戶流程說明。 | — |
| v1.0.10 | 2026-07-17 | 第 13 章改為「資料遷移與上線過渡」，明確採用 Big Bang 切換：無試點、無並行、無唯讀期，測試完成後直接全面上線。 | — |
| v1.0.11 | 2026-07-17 | 確定 monorepo 工具鏈為 pnpm workspace + Turborepo；更新第 2.1 / 2.2 / 12.1 節相關描述。 | — |
| v1.0.12 | 2026-07-17 | 新增第 14 章「安全與合規」；後續章節（風險與假設、後續步驟、AI 整合備忘錄）編號順延並更新內部引用。 | — |
| v1.0.13 | 2026-07-17 | 更新第 2.3 節 WebSocket 認證機制：Web 使用 cookie，App / 無法帶 cookie 情境使用一次性 `ws_ticket`，避免長期 token 置於 URL。 | — |
| v1.0.14 | 2026-07-17 | 新增第 12.3 節「備份、監控與災難復原」；確定 RTO 4 小時 / RPO 1 小時、備份存放 Google Cloud Storage、監控工具待選。 | — |
| v1.0.15 | 2026-07-17 | 規格審查修正：修正標頭修訂號（原誤植 v1.0.4）與章節編號（修訂記錄改為第 18 章）；companies 補 `identifier` / `capabilities` / `public_info`；users 補登入相關欄位；sales_orders 補 `order_no` / `customer_id` / `status` / `expected_delivery_date`；新增 `product_units` / `announcements` / `user_devices` 實體；products 分切規格改關聯表 `product_cutting_specs`；customer_products 移除 `is_deleted`；補充 metadicts NULL 唯一索引與軟刪除部分唯一索引慣例；明確 Web 用 session、App 用 JWT；QR Code 採 Universal Link。 | — |
| v1.0.16 | 2026-07-17 | P0 決策定案：業務 API 統一 Connect-RPC，REST 僅保留公開端點；1.0 僅支援 Google Workspace（自建 IdP 延後）；新增 `token_version` 撤銷機制與 JWT 效期 / refresh 規則；確認自建 PostgreSQL / Valkey，要求上線前完成 PITR 還原演練；監控定案 Prometheus + Grafana + Alertmanager；API access token 限定 server-to-server（`X-Api-Token`）。 | — |
| v1.0.17 | 2026-07-17 | Web 新增角色權限設置與 API 權限設置：新增 `roles` / `role_permissions` / `casbin_rules` 實體；內建 6 角色與其功能權限、Casbin policy 改為 migration seed 預設值；`super` 可新增自訂角色（須指定 `data_scope`）；RLS 改依 `data_scope` 等級判斷；ability 改由 `role_permissions` 動態產生；權限異動需防鎖死並寫入稽核。 | — |
| v1.0.18 | 2026-07-17 | 新增第 4.4 節「開發者帳號」：內建第 7 角色 `developer`（繞過 Casbin 與 RLS），以 `DEVELOPER_ACCOUNT_ENABLED` 環境開關控制，production 預設關閉且啟動防護拒絕誤開；操作照常寫入稽核；上架/上線檢查清單必含關閉確認。 | — |
| v1.0.19 | 2026-07-17 | 對齊 Connect-RPC 統一後的端點描述：REST 公開端點清單補版本查詢與檔案上傳；Casbin 範例改為 RPC method path；ability 端點改為 `AbilityService.GetAbility`。 | — |
| v1.0.20 | 2026-07-17 | 新增第 6.5 節「訂單編號與金額計算」：編號 = 下單來源碼 + 自增序號、`order_counters` 樂觀鎖取號；`tax_inclusive` 下單決定預設外加；金額整數（小數 0 位）逐項四捨五入；1.0 不實作折扣；金額與時間型別慣例。 | — |
| v1.0.21 | 2026-07-17 | 狀態機補回退路徑：新增「取消派車」（處理中 → 待處理，dept_admin 以上 + 原因）與「作廢」（已完成 → `voided` 終態，更正以作廢 + 重建替代）；新增 `sales_order_events` 異動記錄實體；audit action 補 `dispatch_cancel` / `void`。 | — |
| v1.0.22 | 2026-07-17 | P1 決策定案：派車確認改車次批次、看板拖放樂觀鎖（`sales_orders.version`）；訂單通知對象為下單業務（fallback 預設負責業務 / dept_admin）；SMTP 公司級（`companies.smtp_config`）；通知失敗不重試；guest 改註冊完成頁（選公司 + 姓名）；加工單「加工後數量」列印空白手寫回填；新增記錄類資料保留期限表、檔案上傳白名單、覆蓋率門檻 70%。 | — |
| v1.0.23 | 2026-07-17 | 單車總表不顯示金額；客戶密碼政策定案：臨時密碼效期 1 天、最少 8 字元、連續 5 次錯誤鎖定（30 分鐘自動解除）；`users` 補 `temp_password_expires_at` / `failed_login_attempts` / `locked_at`。 | — |
| v1.0.24 | 2026-07-17 | 全面複查修補：RLS 補 `app.current_customer_id`（self 範圍）；ability 改登入載入 + 短 TTL 快取；註冊完成頁公司清單僅公開 display_name；公司停用連鎖登入失效；軟刪除主檔歷史關聯可顯示；FCM 失效 token 自動清除；稽核同步寫入；分頁 per_page 上限 100 且 Connect 列表沿用 meta；註記 1.0 不做病毒掃描。 | — |
| v1.0.25 | 2026-07-18 | 後續步驟補 1.1 規劃指標：App AI 輔助功能（拍照建客戶、業務語音下訂單）經討論定案為 1.1 獨立迭代，不影響 1.0 範圍，設計細節另見 `2026-07-18-app-ai-assist-1.1-design.md`。 | — |
| v1.0.26 | 2026-07-18 | App 手動表單新增客戶（主檔 + 登入帳號）納入 1.0（§9.4）；客戶登入帳號編碼定案為公司定義前綴 + 自增 ID：`companies` 補 `customer_code_prefix`（全系統唯一）、新增 `customer_counters` 樂觀鎖取號實體、§4.2 補編碼規則；1.1 備忘錄之 App 新增客戶前置相應提前，1.1 僅保留拍照 AI 預填與語音下單。 | — |
| v1.0.27 | 2026-08-03 | 派車看板即時機制由 WebSocket 改為 Connect-RPC server streaming（`DispatchService.WatchBoard`，事件僅觸發看板重查、斷線重連、連續失敗降級 30 秒輪詢；對齊 change `dispatch-board-connect-streaming`）：更新 §2.1 技術棧、§2.3 API 與通訊、§7.1 派車看板、§12.3.2 監控指標、§14.1 傳輸加密；移除 `ws_ticket` 端點。 | — |
| v1.0.28 | 2026-08-03 | 需求更新（需求備忘）：①全系統**不儲存任何金額欄位**（§5.2 / §6.5 / §7.2 / §11）；②通知移除 Email / SMTP（§10.1 改 FCM + 站內兩通道）；③一個客戶可綁定**多個登入帳號**（§4.2 / `users.customer_id`）；④通知路由：業務下單推客戶帳號、後台新增專屬商品推主責業務（檢核待定）、退貨審核結果推客戶；⑤**促銷推播**與公告分離（`promo_tags` / `product_promo_tags` 標籤、客戶套用、依分類選客戶群，§5.1 / §5.2 / §10.1）；⑥新增**退貨流程**（客戶發起、業務審核、退貨證明、不建配送頁，§6.6）；⑦客戶主檔新增**偏好送貨日**，非勾選日下單自動順延（§5.2 / §6.2）；⑧`audit_logs` 保留預設 3 個月、管理頁可設定（§10.2 / §14.1）；⑨App 首頁以公告為主、快速下單入口移至底部導覽（§9.1）。 | — |
| v1.0.29 | 2026-08-03 | **店家自助管理帳號**：店家可於 App / Web 自行管理名下登入帳號（新增子帳號 / 停用 / 重置密碼），範圍僅限自己客戶（data_scope self）；防呆（不可停用最後一個 active 帳號、不可停用當前登入帳號）+ 後台 dept_admin 逃生門；帳號管理寫稽核（§4.2 / §9.3）。 | — |
| v1.0.30 | 2026-08-03 | 帳號結構定案為**一主多子**：每客戶 1 個主帳號（建立客戶時產生、`users.is_primary`，帳號名稱預設客戶名稱）+ 多個子帳號；店家以主帳號自助管理（新增 / 停用 / 重置子帳號），子帳號無管理權限；主帳號不可由店家停用 / 重置，後台可重置主帳號密碼轉交（§4.2 / §5.2 / §9.3）。 | — |
| v1.0.31 | 2026-08-03 | **主帳號僅供管理、不提供業務登入**：主帳號登入僅能帳號管理，業務 API（下單 / 訂單 / 退貨 / 專屬商品 / 促銷）一律 403；子帳號為唯一 App 業務登入身分（老闆需下單則另建子帳號）；QR 登入清單僅列子帳號；通知（訂單 / 促銷 / 退貨）僅推子帳號；主帳號首登引導建立首個子帳號；後台停用主帳號連鎖停用子帳號（§4.2 / §5.2 / §10.1）。 | — |
| v1.0.32 | 2026-08-03 | **建立客戶自動附帶業務子帳號**：取代首登引導——建立客戶時同步自動建立主帳號 + 1 個業務子帳號（`is_primary = false`、名稱預設「客戶名稱（業務）」），兩組帳密一次交付（§4.2 / §9.4）；自動附帶的子帳號於店家帳號管理清單**灰化（反白）顯示為「系統預設」**，店家不可改名 / 停用 / 重置，後台可正常管理（§4.2）。 | — |
| v1.0.33 | 2026-08-03 | **業務子帳號專供所屬業務使用**：業務子帳號憑證交付該客戶所屬業務（`default_sales_rep_id`，業務以客戶身分登入代客操作）；店家開箱僅持主帳號，需下單以主帳號自建子帳號；店家帳號管理清單顯示「系統預設（業務使用）」灰化、不可管理（無密碼即無法登入）；主責業務變更時後台移交（重置密碼轉交）；QR 登入清單排除主帳號與業務子帳號（§4.2 / §9.4）。 | — |
| v1.0.34 | 2026-08-03 | **App 新增客戶當場交付管理網址鏈結**：成功頁除主帳號帳密（轉交店家）與業務子帳號帳密（業務留存）外，新增**帳號管理深層連結**（`https://<domain>/customer_account_manage`，Universal Link / App Link，開啟 App 導向帳號管理登入；未安裝導向商店；不含登入憑證，可顯示為 QR Code）一併交付（§4.2 / §9.4）。 | — |

---

*設計規格版本：v1.0.34*
*日期：2026-08-03*
