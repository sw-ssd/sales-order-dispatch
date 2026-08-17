# 多公司訂出貨系統 1.0 — Backend / Web / App 子專案功能拆分設計

> 日期：2026-08-05
> 狀態：**草案，待使用者審閱**
> 對應：
> - `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）
> - `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D28）
> - `docs/superpowers/specs/2026-08-04-app-flutter-stack-design.md`（D29）
> - `docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md`（v2.9.0）
>
> 本文件為 brainstorming 流程產出；實作前須經使用者審閱本文件並依 writing-plans 產生執行計畫。

---

## 1. 目的

將 1.0 核心系統拆成三個子專案 — **Backend**、**Web 中台**、**App** — 並進一步細分到「可馬上開票、指派、驗收」的粒度，使未來啟動實作時能立即安排工作。

---

## 2. 拆分原則與編號架構

### 2.1 拆分粒度

每個工作項目約為「**一位工程師 2–5 天可完成、可單獨驗收、可獨立開票**」的粒度。

### 2.2 編號規則

- `BE-<領域>-<流水號>`：Backend
- `WEB-<領域>-<流水號>`：Web 中台
- `APP-<領域>-<流水號>`：App

領域代碼：

| 代碼 | 領域 |
|---|---|
| `INF` | 基礎建設 |
| `DB` | 資料庫 |
| `PROTO` | API 定義 / proto |
| `AUTH` | 認證授權 |
| `MT` | 多租戶 |
| `USR` | 使用者管理 |
| `MD` | 主檔 |
| `ORD` | 訂單 |
| `RET` | 退貨 |
| `DSP` | 派車 |
| `PRT` | 列印 |
| `NOT` | 通知 |
| `ANN` | 公告 |
| `FIL` | 檔案資產 |
| `AUD` | 稽核 |
| `OPS` | 部署維運 |
| `LAYOUT` | 版面與導覽（僅 Web/App） |
| `SETTINGS` | 設定（僅 Web） |
| `HOME` | 首頁（僅 App） |
| `PRODUCTS` | 商品與快速下單（僅 App） |
| `CUSTOMERS` | 客戶管理（僅 App） |
| `ACCOUNT` | 店家帳號管理（僅 App） |
| `PROFILE` | 功能頁（僅 App） |
| `NOTIFICATIONS` | 通知（僅 App） |

### 2.3 每個工作項目包含

1. **名稱 / 描述**
2. **主要交付物**（檔案、API、頁面、元件）
3. **跨端介面**（proto service、頁面路由、權限點）
4. **相依項目**
5. **驗收標準**

本文件先列出 ID 與名稱；實作計畫階段再為每個 ID 補上 Files / Interfaces / Steps / Acceptance Criteria。

---

## 3. Backend 子專案功能拆分

### 3.1 INF：基礎建設

| ID | 名稱 |
|---|---|
| BE-INF-01 | monorepo 根結構（pnpm workspace + Turborepo + root Taskfile） |
| BE-INF-02 | 開發環境 docker-compose（PostgreSQL + Valkey + Gotenberg） |
| BE-INF-03 | Go 後端骨架（Chi server、config、health check） |
| BE-INF-04 | CI 基礎管線（lint / test / build for Go + frontend + Flutter） |
| BE-INF-05 | buf / proto 產生流程與三端型別同步腳本 |

### 3.2 DB：資料庫

| ID | 名稱 |
|---|---|
| BE-DB-01 | Goose migration 基礎架構與版本控制 |
| BE-DB-02 | Ent schema 基礎與產生流程 |
| BE-DB-03 | RLS 政策基礎（`data_scope` 強制、application_name 切換角色） |
| BE-DB-04 | 軟刪除統一機制與部分唯一索引 |

### 3.3 PROTO：API 定義

| ID | 名稱 |
|---|---|
| BE-PROTO-01 | `common.proto`（分頁、權重、時間區間、列印紙張） |
| BE-PROTO-02 | `auth.proto`（登入/登出/刷新/註冊完成/QR 兌換） |
| BE-PROTO-03 | `company.proto` / `department.proto` |
| BE-PROTO-04 | `user.proto` / `role.proto` |
| BE-PROTO-05 | 主檔 proto（customer / product / warehouse / route / cutting_spec / metadict） |
| BE-PROTO-06 | 訂單 proto（sales_order / order_item / return_request） |
| BE-PROTO-07 | 派車與列印 proto（dispatch / print_job） |
| BE-PROTO-08 | 通知/公告/檔案/稽核 proto |

### 3.4 AUTH：認證授權

| ID | 名稱 |
|---|---|
| BE-AUTH-01 | Google Workspace OIDC 登入與回調 |
| BE-AUTH-02 | 客戶 `customer_code + 密碼` 登入 |
| BE-AUTH-03 | JWT access/refresh 核發與 30 天旋轉 |
| BE-AUTH-04 | Web session cookie（scs + Valkey） |
| BE-AUTH-05 | `token_version` 撤銷機制（改密碼/停用/角色變更） |
| BE-AUTH-06 | 密碼政策（臨時密碼 24h、首登強改、錯誤鎖定） |
| BE-AUTH-07 | QR Code 登入兌換（店家子帳號） |
| BE-AUTH-08 | developer 逃生門（env fail-safe + 繞過 Casbin/RLS） |
| BE-AUTH-09 | Casbin policy 載入與 7 內建角色 seed |
| BE-AUTH-10 | RLS 資料範圍強制（company/department/self） |
| BE-AUTH-11 | 角色權限管理 API（company_admin 限自己公司） |
| BE-AUTH-12 | API 權限清單與 ability 元資料 |

### 3.5 MT：多租戶

| ID | 名稱 |
|---|---|
| BE-MT-01 | Company CRUD + 公開識別碼/公開資訊 |
| BE-MT-02 | Department CRUD 與隸屬關係 |
| BE-MT-03 | 公司識別呈現（logo、主題色、名稱） |
| BE-MT-04 | 部門切換 API 與 context 注入 |

### 3.6 USR：使用者管理

| ID | 名稱 |
|---|---|
| BE-USR-01 | 員工帳號 CRUD（含 guest 審核） |
| BE-USR-02 | 客戶主帳號 / 子帳號模型（D22 一主多子） |
| BE-USR-03 | 客戶帳號管理 API（主帳號自助、後台逃生門） |
| BE-USR-04 | 員工註冊完成頁 API（選公司 + 姓名） |

### 3.7 MD：主檔

| ID | 名稱 |
|---|---|
| BE-MD-01 | 客戶 CRUD + 取號（company prefix + 自增） |
| BE-MD-02 | 客戶地址簿與聯絡人 |
| BE-MD-03 | 商品 CRUD + 單位換算 |
| BE-MD-04 | 倉別 / 車次 / 分切規格 / 商品分類（部門級） |
| BE-MD-05 | metadicts 字典檔（系統預設 + 部門擴充） |
| BE-MD-06 | 客戶專屬商品（customer_products）+ 別名機制 |
| BE-MD-07 | 客戶偏好送貨日（preferred_delivery_days） |

### 3.8 ORD：訂單

| ID | 名稱 |
|---|---|
| BE-ORD-01 | 訂單 CRUD + 編號取號（來源碼 + 自增） |
| BE-ORD-02 | 訂單狀態機（pending / processing / completed / cancelled / voided） |
| BE-ORD-03 | 訂單事件軌跡（sales_order_events） |
| BE-ORD-04 | 客戶專屬商品帶入 + 手打商品別名 |
| BE-ORD-05 | 偏好送貨日順延邏輯 |
| BE-ORD-06 | 下單通知觸發（業務下單推客戶子帳號） |

### 3.9 RET：退貨

| ID | 名稱 |
|---|---|
| BE-RET-01 | 退貨申請 CRUD（客戶發起） |
| BE-RET-02 | 業務審核退貨 + 退貨證明 |
| BE-RET-03 | 退貨品項來源（歷史訂單 + 專屬商品並存） |

### 3.10 DSP：派車

| ID | 名稱 |
|---|---|
| BE-DSP-01 | 車次批次確認 API |
| BE-DSP-02 | 派車看板資料 API（依日期/部門） |
| BE-DSP-03 | `WatchBoard` Connect server streaming |
| BE-DSP-04 | Valkey pub/sub 跨 replica 通知 |
| BE-DSP-05 | 降級輪詢端點（30 秒） |

### 3.11 PRT：列印

| ID | 名稱 |
|---|---|
| BE-PRT-01 | Gotenberg HTML→PDF 整合 |
| BE-PRT-02 | 單車總表模板 |
| BE-PRT-03 | 車次對點單模板 |
| BE-PRT-04 | 揀貨單模板（車次→倉別→分類→品名） |
| BE-PRT-05 | 加工單模板（加工室揀/配送揀 + 手寫回填） |
| BE-PRT-06 | 重印記錄與原因必填 |

### 3.12 NOT：通知

| ID | 名稱 |
|---|---|
| BE-NOT-01 | FCM token 註冊/更新 |
| BE-NOT-02 | 站內通知 CRUD 與 badge |
| BE-NOT-03 | 業務下單推客戶子帳號 |
| BE-NOT-04 | 專屬商品新增推主責業務 |
| BE-NOT-05 | 退貨審核結果推發起帳號 |

### 3.13 ANN：公告

| ID | 名稱 |
|---|---|
| BE-ANN-01 | 公告 CMS CRUD |
| BE-ANN-02 | 促銷推播（分類標籤選群） |
| BE-ANN-03 | 公告輪詢/推播觸發 |

### 3.14 FIL：檔案資產

| ID | 名稱 |
|---|---|
| BE-FIL-01 | 檔案上傳（白名單 + magic bytes） |
| BE-FIL-02 | 檔案下載 / 預覽 |
| BE-FIL-03 | restic 備份 GCS |

### 3.15 AUD：稽核

| ID | 名稱 |
|---|---|
| BE-AUD-01 | audit_logs 同交易寫入 |
| BE-AUD-02 | 關鍵操作埋點（登入/建檔/下單/改權限/刪除） |
| BE-AUD-03 | 稽核保留期設定（1/3/6/12 個月或永久） |

### 3.16 OPS：部署維運

| ID | 名稱 |
|---|---|
| BE-OPS-01 | k8s manifests（PostgreSQL / Valkey StatefulSet） |
| BE-OPS-02 | Helm chart / 應用部署 |
| BE-OPS-03 | Prometheus + Grafana + Alertmanager |
| BE-OPS-04 | GCS 每日備份 + WAL PITR |
| BE-OPS-05 | 首次災難復原演練 |

---

## 4. Web 中台子專案功能拆分

採用 SolidJS + Vite + TypeScript + TanStack Query + CASL + Connect-RPC。

### 4.1 INF：前端基礎

| ID | 名稱 |
|---|---|
| WEB-INF-01 | Vite + SolidJS + TypeScript 專案骨架 |
| WEB-INF-02 | Tailwind CSS + shadcn-solid 元件庫整合 |
| WEB-INF-03 | 路由架構（@solidjs/router）與守衛（AuthGuard / RoleGuard） |
| WEB-INF-04 | TanStack Query 設定與錯誤處理 |
| WEB-INF-05 | CASL ability 與 UI 權限遮蔽 |
| WEB-INF-06 | Connect-RPC 產生 TypeScript client 整合 |

### 4.2 LAYOUT：版面與導覽

| ID | 名稱 |
|---|---|
| WEB-LAYOUT-01 | 登入後 shell（側邊欄 + header + 內容區） |
| WEB-LAYOUT-02 | 公司識別呈現（logo、主題色、名稱） |
| WEB-LAYOUT-03 | 部門切換器 |
| WEB-LAYOUT-04 | 底部/全域通知 badge |
| WEB-LAYOUT-05 | 響應式基礎（desktop 為主，tablet 相容） |

### 4.3 AUTH：認證頁面

| ID | 名稱 |
|---|---|
| WEB-AUTH-01 | 員工 Google 登入導向頁 |
| WEB-AUTH-02 | 員工註冊完成頁（選公司 + 姓名） |
| WEB-AUTH-03 | 密碼變更 / 首登強改頁 |
| WEB-AUTH-04 | 登入錯誤與鎖定提示 |

### 4.4 MT：公司/部門管理

| ID | 名稱 |
|---|---|
| WEB-MT-01 | 公司管理頁（super/company_admin） |
| WEB-MT-02 | 部門管理頁（company_admin） |
| WEB-MT-03 | 公司公開資訊設定 |

### 4.5 USR：使用者管理

| ID | 名稱 |
|---|---|
| WEB-USR-01 | 員工帳號列表與 CRUD |
| WEB-USR-02 | 角色權限設定頁 |
| WEB-USR-03 | API 權限檢視頁 |
| WEB-USR-04 | guest 審核佇列 |
| WEB-USR-05 | 客戶主帳號 / 子帳號管理頁 |
| WEB-USR-06 | 客戶帳號後台逃生門（停用/重置/移交業務子帳號） |

### 4.6 MD：主檔頁面

| ID | 名稱 |
|---|---|
| WEB-MD-01 | 客戶主檔列表與 CRUD |
| WEB-MD-02 | 客戶地址簿 / 聯絡人管理 |
| WEB-MD-03 | 商品主檔列表與 CRUD |
| WEB-MD-04 | 單位換算設定 |
| WEB-MD-05 | 倉別 / 車次 / 分切規格 / 商品分類管理 |
| WEB-MD-06 | metadicts 字典檔管理 |
| WEB-MD-07 | 客戶專屬商品清單（別名、預設數量、啟用/停用） |
| WEB-MD-08 | 客戶偏好送貨日設定 |

### 4.7 ORD：訂單頁面

| ID | 名稱 |
|---|---|
| WEB-ORD-01 | 訂單列表（篩選、分頁、狀態） |
| WEB-ORD-02 | 訂單建立頁（選客戶 → 帶出專屬商品 → 手打/總表選品） |
| WEB-ORD-03 | 訂單詳情頁（狀態、事件軌跡） |
| WEB-ORD-04 | 訂單狀態操作（確認/取消/作廢） |
| WEB-ORD-05 | 訂單欄位：預計出貨日、備註、特殊切法 |

### 4.8 RET：退貨頁面

| ID | 名稱 |
|---|---|
| WEB-RET-01 | 退貨申請列表 |
| WEB-RET-02 | 業務審核退貨頁 |
| WEB-RET-03 | 退貨證明檢視/列印 |

### 4.9 DSP：派車看板

| ID | 名稱 |
|---|---|
| WEB-DSP-01 | 派車看板頁（Kanban 拖放、依日期/車次） |
| WEB-DSP-02 | 車次批次確認 UI |
| WEB-DSP-03 | Connect streaming 更新 / 輪詢降級顯示 |
| WEB-DSP-04 | 看板版本衝突提示（樂觀鎖） |

### 4.10 PRT：列印

| ID | 名稱 |
|---|---|
| WEB-PRT-01 | 列印入口與車次選擇 |
| WEB-PRT-02 | 單車總表預覽/列印 |
| WEB-PRT-03 | 車次對點單預覽/列印 |
| WEB-PRT-04 | 揀貨單預覽/列印 |
| WEB-PRT-05 | 加工單預覽/列印 |
| WEB-PRT-06 | 重印原因輸入與列印記錄 |

### 4.11 NOT：通知中心

| ID | 名稱 |
|---|---|
| WEB-NOT-01 | 站內通知列表與已讀 |
| WEB-NOT-02 | 通知 badge 與下拉 |
| WEB-NOT-03 | 通知設定（開關） |

### 4.12 ANN：公告與促銷

| ID | 名稱 |
|---|---|
| WEB-ANN-01 | 公告 CMS 編輯頁 |
| WEB-ANN-02 | 公告列表與置頂 |
| WEB-ANN-03 | 促銷推播建立頁（分類標籤選群） |

### 4.13 FIL：檔案

| ID | 名稱 |
|---|---|
| WEB-FIL-01 | 檔案上傳元件 |
| WEB-FIL-02 | 檔案列表/預覽/下載 |

### 4.14 AUD：稽核

| ID | 名稱 |
|---|---|
| WEB-AUD-01 | 稽核日誌查詢頁 |
| WEB-AUD-02 | 稽核保留期設定頁 |

### 4.15 SETTINGS：設定

| ID | 名稱 |
|---|---|
| WEB-SETTINGS-01 | 個人資料/密碼 |
| WEB-SETTINGS-02 | 公司/部門相關設定 |

---

## 5. App 子專案功能拆分

採用已定案技術棧：Flutter + auto_route + solidart + disco + fquery + Sembast 鏡像 + connectrpc。

### 5.1 INF：App 基礎

| ID | 名稱 |
|---|---|
| APP-INF-01 | Flutter 專案骨架 + flavor（dev/prod） |
| APP-INF-02 | auto_route 路由與 guards（AuthGuard / RoleGuard） |
| APP-INF-03 | disco DI 分層（根 scope / auth scope / 路由級 scope） |
| APP-INF-04 | fquery + Sembast 唯讀快取鏡像 |
| APP-INF-05 | Connect-RPC 產生 Dart client |
| APP-INF-06 | 錯誤分類與統一 UI（snackbar / 離線橫幅） |
| APP-INF-07 | FCM 推播整合與 token 上傳 |

### 5.2 LAYOUT：導覽與殼層

| ID | 名稱 |
|---|---|
| APP-LAYOUT-01 | 登入後底部 4 tab shell（首頁 / 商品 / 訂單 / 功能） |
| APP-LAYOUT-02 | 公司識別呈現（logo、名稱） |
| APP-LAYOUT-03 | 通知 badge |
| APP-LAYOUT-04 | 離線/快取資料指示器 |

### 5.3 AUTH：認證流程

| ID | 名稱 |
|---|---|
| APP-AUTH-01 | 身分選擇頁（我是店家 / 我是業務） |
| APP-AUTH-02 | 店家登入（customer_code + 密碼） |
| APP-AUTH-03 | 業務 Google OAuth PKCE 登入 |
| APP-AUTH-04 | 員工註冊完成頁（選公司 + 姓名） |
| APP-AUTH-05 | QR Code 登入（店家子帳號） |
| APP-AUTH-06 | 401 refresh + 403 token_version 撤銷處理 |
| APP-AUTH-07 | 首登強改密碼 |

### 5.4 HOME：首頁

| ID | 名稱 |
|---|---|
| APP-HOME-01 | 公告消息輪播/列表 |
| APP-HOME-02 | 快速下單入口 |
| APP-HOME-03 | 首頁資料離線鏡像與 seed |

### 5.5 PRODUCTS：商品與快速下單

| ID | 名稱 |
|---|---|
| APP-PRODUCTS-01 | 商品總表瀏覽 |
| APP-PRODUCTS-02 | 客戶專屬商品清單 |
| APP-PRODUCTS-03 | 快速下單流程（選客戶 → 選品/手打 → 數量/單位/切法/備註 → 預計出貨日 → 送出） |
| APP-PRODUCTS-04 | 手打商品別名綁定客戶 |
| APP-PRODUCTS-05 | 單位換算與預設數量 |
| APP-PRODUCTS-06 | 偏好送貨日順延提示 |

### 5.6 ORDERS：訂單與退貨

| ID | 名稱 |
|---|---|
| APP-ORDERS-01 | 訂單歷史列表（篩選、分頁） |
| APP-ORDERS-02 | 訂單詳情頁 |
| APP-ORDERS-03 | 退貨申請（歷史訂單勾選 + 專屬商品清單並存） |
| APP-ORDERS-04 | 退貨結果通知與狀態追蹤 |

### 5.7 CUSTOMERS：客戶管理

| ID | 名稱 |
|---|---|
| APP-CUSTOMERS-01 | 客戶列表 |
| APP-CUSTOMERS-02 | 客戶搜尋 |
| APP-CUSTOMERS-03 | 手動表單新增客戶（主檔 + 登入帳號） |
| APP-CUSTOMERS-04 | 客戶詳情與編輯 |

### 5.8 ACCOUNT：店家帳號管理（主帳號）

| ID | 名稱 |
|---|---|
| APP-ACCOUNT-01 | 子帳號列表 |
| APP-ACCOUNT-02 | 新增子帳號 |
| APP-ACCOUNT-03 | 停用/啟用子帳號 |
| APP-ACCOUNT-04 | 重置子帳號密碼 |
| APP-ACCOUNT-05 | 主帳號路由守衛（禁止進入業務流程） |

### 5.9 PROFILE：功能頁

| ID | 名稱 |
|---|---|
| APP-PROFILE-01 | 個人資料 |
| APP-PROFILE-02 | QR Code 入口 |
| APP-PROFILE-03 | 設定（通知開關） |
| APP-PROFILE-04 | 登出與清空鏡像 |

### 5.10 NOTIFICATIONS：通知

| ID | 名稱 |
|---|---|
| APP-NOTIFICATIONS-01 | 站內通知列表 |
| APP-NOTIFICATIONS-02 | FCM 前景/背景處理 |
| APP-NOTIFICATIONS-03 | 通知點擊導航 |

---

## 6. 跨子專案介面對照與實作順序

### 6.1 Proto Service ↔ 子專案對照

| Proto Service | Backend 負責 | Web 消費 | App 消費 |
|---|---|---|---|
| `AuthService` | 登入/登出/刷新/QR 兌換 | ✅ | ✅ |
| `CompanyService` / `DepartmentService` | 公司/部門 CRUD | ✅ | 僅讀取 |
| `UserService` / `RoleService` | 使用者/角色權限 | ✅ | 僅讀取 session |
| `CustomerService` | 客戶主檔 | ✅ | ✅ |
| `ProductService` | 商品主檔 | ✅ | ✅（唯讀為主） |
| `CustomerProductService` | 客戶專屬商品 | ✅ | ✅ |
| `WarehouseService` / `RouteService` / `CuttingSpecService` | 倉別/車次/分切 | ✅ | 僅讀取 |
| `SalesOrderService` | 訂單 CRUD + 狀態機 | ✅ | ✅ |
| `ReturnRequestService` | 退貨申請 | ✅ | ✅ |
| `DispatchService` / `WatchBoard` | 派車與串流 | ✅ | ❌ |
| `PrintService` | 列印 | ✅ | ❌ |
| `NotificationService` | 通知 | ✅ | ✅ |
| `AnnouncementService` | 公告/促銷 | ✅ | ✅ |
| `FileAssetService` | 檔案 | ✅ | ✅ |
| `AuditLogService` | 稽核 | ✅ | ❌ |

### 6.2 關鍵權限點（Casbin 最小集合）

| action | resource | 適用角色範例 |
|---|---|---|
| `create/read/update/delete` | `company` | super |
| `create/read/update/delete` | `department` | super, company_admin |
| `create/read/update/delete` | `user` | super, company_admin, dept_admin（限 scope） |
| `manage` | `role_permissions` | company_admin（限自己公司） |
| `create/read/update/delete` | `customer` | staff, dept_admin（限 department/self） |
| `create/read/update/delete` | `product` | staff, dept_admin |
| `create/read/update/delete` | `sales_order` | staff, dept_admin, customer（self） |
| `approve` | `return_request` | staff, dept_admin |
| `operate` | `dispatch_board` | staff, dept_admin |
| `print` | `print_job` | staff, dept_admin |
| `manage` | `announcement` | dept_admin, company_admin |
| `read` | `audit_log` | super, company_admin, dept_admin（依 scope） |

### 6.3 共享頁面/流程對照

| 業務流程 | Web 頁面 | App 流程 |
|---|---|---|
| 登入 | `/login`、Google callback | `/login`、身分選擇 |
| 選部門/註冊完成 | `/register-complete` | `/register-complete` |
| 客戶列表/CRUD | `/customers/*` | 客戶 tab |
| 快速下單 | `/orders/new` | `/products/quick-order` |
| 訂單歷史 | `/orders` | `/orders` |
| 退貨申請 | `/returns/*` | `/orders/:id/return` |
| 派車看板 | `/dispatch` | 無 |
| 列印 | `/prints/*` | 無 |
| 公告 | `/announcements` | 首頁輪播 |
| 通知 | 通知下拉/列表 | 通知列表 |
| 店家帳號管理 | `/users/customer-accounts` | `/account/sub-accounts` |

### 6.4 實作順序與阻塞關係

**Wave 1：必須先完成**
- BE-INF-*、BE-DB-*、BE-PROTO-*、BE-AUTH-*、BE-MT-*、BE-USR-*
- WEB-INF-*、WEB-LAYOUT-*、WEB-AUTH-*
- APP-INF-*、APP-LAYOUT-*、APP-AUTH-*

**Wave 2：主檔**
- BE-MD-*、WEB-MD-*、APP-CUSTOMERS-*、APP-PRODUCTS-*

**Wave 3：訂單與通知**
- BE-ORD-*、BE-NOT-*、WEB-ORD-*、APP-ORDERS-*、APP-PRODUCTS-03

**Wave 4：派車與列印**
- BE-DSP-*、BE-PRT-*、WEB-DSP-*、WEB-PRT-*

**Wave 5：公告、檔案、稽核、部署**
- BE-ANN-*、BE-FIL-*、BE-AUD-*、WEB-ANN-*、WEB-FIL-*、WEB-AUD-*、APP-NOTIFICATIONS-*、BE-OPS-*

**原則**：每一 Wave 內三端可並行；跨 Wave 的後端 API 必須先於前端/App 頁面。

### 6.5 風險與相依注意

- `BE-AUTH-*` 全部完成前，任何需要登入的頁面都無法整合測試。
- `BE-MD-*` 完成前，`BE-ORD-*` 無法驗收（下單需要客戶與商品）。
- `BE-DSP-*` 的 `WatchBoard` 串流需等 `BE-ORD-*` 穩定後再對接。
- `BE-PRT-*` 需等 `BE-DSP-*` 的車次資料穩定。

---

## 7. 文件回寫（審閱通過後執行）

1. 本設計文件保留於 `docs/superpowers/specs/`。
2. 依本文件產生 `docs/superpowers/plans/YYYY-MM-DD-sales-order-1.0-subproject-implementation-plan.md`，每個工作項目補上 Files / Interfaces / Steps / Acceptance Criteria。
3. 若 Wave 順序影響現有執行計畫 `2026-07-17-sales-order-1-0-tasks.md`，則升版對齊。

---

## 8. Open Questions

| 問題 | 影響 | 狀態 |
|---|---|---|
| 每個工作項目的確切工時與負責人 | 排程 | 於實作計畫階段確認 |
| 是否需要為每個 BE-* 項目補上 Ent schema / migration / proto / service 四層拆分 | 粒度 | 於實作計畫階段決定 |
| Web/App 共用元件庫是否抽成獨立 workspace package | 架構 | 於實作計畫階段決定 |

---

## 9. 修訂記錄

| 修訂號 | 日期 | 修訂內容 | 修訂者 |
|---|---|---|---|
| v0.1.0 | 2026-08-05 | 初版：Backend / Web / App 三子專案功能拆分、跨端介面對照、實作 Wave 順序 | 開發團隊 |
