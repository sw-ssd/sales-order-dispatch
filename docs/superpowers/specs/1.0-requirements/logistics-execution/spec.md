# logistics-execution 需求規格

> 來源：2026-09-17 Fleetbase 整合（決策 D32）。Fleetbase 側整合決策見 Fleetbase 倉庫 `docs/FLEETBASE_物流平台重建_PLAN.md`。本檔為 logistics 執行層的需求層規格（Requirement + Scenario），細部實作見 `docs/superpowers/plans/backend/detail/10-logistics-execution.md`。

## 概述

logistics 執行層補足 1.0「規劃/裝單」之後的「現場配送/追蹤/簽收」層：把派車看板完成的車次（`route_id`）實際綁定車輛與司機執行，並提供即時定位、路線/ETA、簽收（POD）。logistics 領域為**部門級**（`company_id` + `department_id`），資料範圍依 D3 data_scope。授權由 **OpenFGA**（資源級）+ RLS（資料庫兜底）承擔（D32），CASL 不保留。**NetSuite 不接 logistics**。

## Schema 承襲與收斂

> logistics 執行層的 schema 源自 Fleetbase 逆向（Fleetbase 倉庫 `docs/study/erd/go-*.mmd` 與 `docs/study/erd/go/schema-improvement.md`），但**已收斂至 1.0 語彙與慣例**，Fleetbase 僅為欄位級參考、非權威。

- **識別符**：採用 1.0 的 `bigserial id + uuid`（D31），**不引入** Fleetbase 的 `_key` 與雙主鍵冗餘；對外一律 `uuid`。
- **租戶**：一律 `company_id` + `department_id`（F4），資料範圍依 D3 data_scope；Fleetbase 的 `company_uuid` 單租戶欄對應為 1.0 兩層。
- **多型**：**不引入** Fleetbase 的 `*_type/*_uuid` 多型欄位；能具體化就拆明確 FK（如 `customer` 收斂至既有主檔關聯），真多型加 `CHECK`。
- **金額/度量**：一律 `BIGINT`+`currency`、`NUMERIC`+`*_unit`，**不引入** MySQL `VARCHAR` 金額/度量欄。
- **軟刪除**：業務主檔（logistics_teams/vehicles/drivers/logistics_deliveries）依 D10 軟刪除 + 部分唯一索引；軌跡表（`positions`/`logistics_delivery_events`）**append-only**、不下 `deleted_at`。
- **稽核**：指派/狀態異動/簽收與稽核同一交易（D18），共用既有 `audit_logs`。
- **欄位細目**：以 `docs/superpowers/plans/backend/detail/10-logistics-execution.md`「領域模型 schema 細目」為權威。

此收斂確保 1.0 不因「照抄 Fleetbase 表」而承接 MySQL 逆向的技術債（多型、字串數值、雙主鍵）。

## Requirements

### Requirement: logistics 主檔為部門級實體

系統 SHALL 提供 logistics 主檔 `logistics_teams` / `vehicles` / `drivers` 的建立、查詢、更新與軟刪除；三者 MUST 帶 `company_id` 與 `department_id`，資料範圍依使用者 data_scope（staff/dept_admin 限本部門、company_admin 限所屬公司全部門、super 全公司）。`vehicles` MUST 記錄 `plate_no`、`vehicle_type`、`capacity`；`drivers` MUST 關聯既有 `users`（司機為一種使用者）並記錄 `name`、`phone`、`assigned_vehicle_id`。業務唯一性（如 `plate_no`）配合軟刪除部分唯一索引（D10）。

#### Scenario: dept_admin 建立司機

- **GIVEN** dept_admin 登入於公司 A 部門甲
- **WHEN** 建立一位司機並指派到部門甲的一台車
- **THEN** 司機建立成功、`company_id`/`department_id` 為部門甲、並關聯該車
- **AND** 部門乙的使用者看不到該司機

#### Scenario: 跨部門隔離

- **WHEN** 部門乙的 staff 查詢車隊列表
- **THEN** 僅回傳部門乙的 vehicles/drivers/logistics_teams，不含部門甲任何資料

#### Scenario: 軟刪除後可重建同名車牌

- **WHEN** 某 `vehicles.plate_no` 已被軟刪除，同部門以相同 `plate_no` 建立新車
- **THEN** 建立成功（部分唯一索引僅約束 `deleted_at IS NULL`）

### Requirement: 服務範圍與地理區（service_areas / zones）

系統 SHALL 提供 `service_areas` 與 `zones`（部門級），記錄地理邊界（PostGIS `geography(Polygon)`；若不可用 fallback GeoJSON 文字欄位），用於後續派單與區域檢核。資料範圍與 logistics 主檔相同。

#### Scenario: 建立服務範圍

- **WHEN** dept_admin 建立一個 `service_area` 並上傳多邊形邊界
- **THEN** 邊界以 Polygon 儲存、隸屬該部門，可被區域查詢使用

### Requirement: 客戶送貨地址座標

系統 SHALL 於 `customer_addresses`（`type=shipping`）記錄座標 `location geography(Point)`，於建立/更新地址時進行地理編碼（geocode）；供 logistics 路線 / ETA（OSRM）使用。**不另建立 Fleetbase 式 places 表**；地址簿仍為客戶子資源（master-data）。地理編碼失敗時，地址 MUST 仍可建立（僅無座標），並於派單路線計算時提示缺座標。

#### Scenario: shipping 地址可取得座標

- **WHEN** 建立或更新一筆 `type=shipping` 的 `customer_addresses` 且地理編碼成功
- **THEN** 該地址取得 `location` 座標，後續可用於 OSRM 路線計算

#### Scenario: 地理編碼失敗不擋建檔

- **WHEN** 地址無法地理編碼（如地址不完整）
- **THEN** 地址照常建立、`location` 為空
- **AND** 當該地址被納入 logistics 路線計算時，系統提示缺座標（`invalid_argument` 或前端警示）

### Requirement: 車次 ↔ 車輛 / 司機指派（後台，粒度在 logistics_delivery）

系統 SHALL 提供 `AssignmentService`，由後台把一個已裝單完成的車次（`route_id`）指派實際車輛與司機，寫入 `logistics_deliveries`（每筆帶 `driver_assigned_uuid` / `vehicle_assigned_uuid`、`assigned_by`、`version`）。指派以樂觀鎖比對 `version`，並發衝突時拒絕並提示。**非固定「車次↔車」1:1**：一車一司機可服務多車次、可重指派（重指派 `version` 遞增）。自動派單（距離最近司機）排 v1 之後，v1 僅手動。

#### Scenario: 指派成功

- **WHEN** 後台對車次 R1 的 `logistics_delivery` 指派司機 D 與車輛 V，且 `version` 一致
- **THEN** `logistics_deliveries` 寫入 `driver_assigned_uuid`/`vehicle_assigned_uuid`/`assigned_by`、`version` 遞增
- **AND** 司機 App 收到該車次執行清單

#### Scenario: 並發指派樂觀鎖衝突

- **WHEN** 兩請求同時以相同 `version` 指派同一 `logistics_delivery`
- **THEN** 其中一個更新因 version 衝突被拒絕、前端提示資料已被變更並重查
- **AND** 最終只保留一次指派結果

#### Scenario: 重指派

- **WHEN** 後台已指派車輛 V1，之後改派車輛 V2
- **THEN** `vehicle_assigned_uuid` 更新為 V2、`version` 遞增，歷史指派軌跡可稽核

### Requirement: 即時定位串流

系統 SHALL 讓司機 App 上報位置（`positions` 寫入），並以 `TrackingService.SubscribePositions`（Connect server-streaming）推播 `driver.location_changed` 事件至訂閱同部門的連線。事件廣播依**部門**隔離，MUST 在任意後端 replica 數下正確送達（Valkey pub/sub 跨 replica 轉發，沿用 D14）。串流認證走與其他 RPC 相同的 cookie/token（不要求一次性 ticket）。前端收到事件後以失效提示觸發全量重查（事件不直接改快取）；串流斷線自動重連並於重連後全量重查；連續失敗降級 30 秒輪詢。

#### Scenario: 位置即時反映到後台地圖

- **WHEN** 司機 App 上報位置
- **THEN** 後台地圖訂閱連線收到 `driver.location_changed`（含 driver_id、location、timestamp）
- **AND** 未認證連線被拒、跨部門連線不收到他部門事件

#### Scenario: 斷線重連補齊

- **WHEN** 串流斷線、期間有位置變更，隨後重連成功
- **THEN** 前端於重連後全量重查，顯示最新位置

### Requirement: 配送執行與簽收（POD）

系統 SHALL 提供 `DeliveryService` 讓司機對指派給自己的 `logistics_delivery` 執行開始（started）、完成（completed）與取消（cancelled）；完成時可上傳簽收證明 **POD**（`proofs`，`type` 為 `photo` / `signature` / `scan`）。POD 影像走 1.0 檔案資產（D17），操作寫稽核（D18）。狀態異動寫 `logistics_delivery_events` 事件軌跡。

#### Scenario: 完成配送並上傳簽收

- **WHEN** 司機完成該 `logistics_delivery` 並上傳 photo POD
- **THEN** `logistics_deliveries.status=completed`、`proofs` 記錄型別與檔案 asset id、寫稽核與事件軌跡
- **AND** 通知/看板反映完成狀態

#### Scenario: 非本人不可操作

- **WHEN** 非被指派的司機嘗試開始/完成某 `logistics_delivery`
- **THEN** 以 `permission_denied` 拒絕（OpenFGA 判定）

### Requirement: 路線 / ETA（OSRM）

系統 SHALL 依起點（`route_id` 出發地 / 倉別）與終點（客戶 `shipping` 地址座標）呼叫 OSRM 計算路線與 ETA，供司機 App 顯示與後台參考。OSRM 以 `OSRM_HOST` 設定（預設 `https://router.project-osrm.org`；rate limit 不足時自建 `osrm/osrm-backend` container）。缺座標的地址不得參與路線計算（見「客戶送貨地址座標」）。

#### Scenario: 計算路線 ETA

- **WHEN** 司機載入車次執行清單且所有配送地址有座標
- **THEN** App 顯示 OSRM 回傳的路線與各站 ETA

### Requirement: NetSuite 不接 logistics

logistics 執行層 MUST NOT 與 NetSuite 同步或耦合；logistics 領域無任何 NetSuite 欄位或同步邏輯。

#### Scenario: logistics 領域無 NetSuite 依賴

- **WHEN** 檢查 logistics 領域 schema 與程式碼
- **THEN** 無 NetSuite 欄位、無 NetSuite client 呼叫、無同步 job

### Requirement: 每單送達狀態回寫（A,1.0）

系統 SHALL 在 `logistics_delivery` 完成時，將送達結果**回寫到該車次所載的每筆 `sales_order`**（標 `delivered`），使店家 / 客戶可看到每一單的送達狀態；回寫與 `logistics_delivery` 狀態異動、稽核同事務（D18）。未覆蓋到的訂單（部分交付語意）在 1.1+ 處理。

#### Scenario: 車次完成回寫逐單送達

- **WHEN** 司機完成該 `logistics_delivery`（含 POD）
- **THEN** 該車次所載的每筆 `sales_order` 標 `delivered`，寫 `sales_order_events` 與稽核，同一交易
- **AND** 店家 / 客戶查詢可見送達狀態

### Requirement: geofence 自動送達判定（A,1.0）

系統 SHALL 依司機即時位置與 `service_areas` / `zones` 邊界自動判定「到站」（到達 customer 的配送點 service_area/zone），供派車看板與 POD 提示，降低司機手動操作；判定結果寫事件。

#### Scenario: 進入送達範圍自動提示

- **WHEN** 司機位置進入該配送點的 service_area/zone 且尚未標記完成
- **THEN** 系統自動標記到站狀態、寫事件並提示司機可執行 POD

### Requirement: logistics 通知接 D16（A,1.0）

logistics 領域的通知（新任務指派推司機、送達推店家 / 主責業務）SHALL 走既有通知基礎設施（D16：FCM + 站內、範本/裝置管理），不另建一套通知。

#### Scenario: 指派推司機、送達推客戶

- **WHEN** 後台指派 `logistics_delivery` 或司機完成送達
- **THEN** 依 D16 範本分別推播司機（新任務）與店家 / 主責業務（送達），寫 `notifications` 記錄

### Requirement: 司機身分 / 角色（1.0,負載關鍵）

系統 SHALL 提供 **driver 角色**：`drivers` 關聯既有 `users`,並以 OpenFGA relation（`driver`）與 RLS data_scope（department/self）賦權;司機可憑既有 JWT 登入司機 App、只取自己被指派的任務。

#### Scenario: 司機登入僅見自己派單

- **WHEN** 司機以既有帳號登入司機 App
- **THEN** 取得 JWT 且 OpenFGA/RLS 僅允許其部門（data_scope=department）內、且 `assignee` 為自己的 `logistics_delivery`

### Requirement: 配送停點 / waypoint 清單（1.0 簡版）

系統 SHALL 提供「逐站執行清單」：每個 `logistics_delivery` 展開為依 `delivery_sequence` 排序的**停點（waypoint）**(客戶 / 送貨地址 / 訂單明細),供司機 App 導航與逐一執行；停點與訂單/地址座標關聯（地址座標見 master-data）。

#### Scenario: 司機取得逐站清單

- **WHEN** 司機載入被指派的 `logistics_delivery`
- **THEN** 依 `delivery_sequence` 回傳各停點(客戶、地址、訂單、座標)

### Requirement: 後台車隊管理 + 執行地圖（1.0,console）

中台 SHALL 提供車隊管理 CRUD 頁 + 即時執行地圖（Leaflet + Connect 串流訂閱 `driver.location_changed` 移動 marker/看板卡片）。

#### Scenario: 後台可見車隊與執行地圖

- **WHEN** 後台開啟車隊頁
- **THEN** 可 CRUD 車輛/司機/車隊,並在地圖即時見司機位置

### Requirement: 司機 App 模式（1.0）

App SHALL 提供「司機模式」：我的任務清單、逐一執行停點、上報位置、POD 簽收（photo/signature/scan）;沿用既有 App 基建與 JWT。

#### Scenario: 司機 App 完成配送

- **WHEN** 司機於 App 選取任務 → 依停點導覽 → 完成並簽收 POD
- **THEN** 上報位置並完成 `logistics_delivery`、回寫 `sales_order`、走 D16 通知

## 1.1+ 延伸（不在 1.0 範圍,v1 後再做）

以下 logistics 能力納入 1.1+（或另行評估），不擠入 1.0 logistics 執行層：

- 自動派單（距離最近司機、容量 / 車型匹配）
- 多點路線優化（VRP / TSP，對應 Fleetbase Vroom）
- 現場拒收 / 部分交付，並串接既有退貨流程（D25）
- ETA 異常通知（遲到 / 偏離路線）
- 客戶 / 店家端即時追蹤與 live map（customer portal）
- 車輛狀態 / 保養 / 出勤管理
- 司機 App 背景 GPS / 離線任務快取
- 每單 tracking 狀態鏈（對應 Fleetbase tracking_numbers/statuses）
- 位置軌跡保留 / 回播 / 稽核查詢（D18 延伸）
- logistics 報表與績效分析（出貨量、司機完成率）
- 多公司 logistics 視圖（super 跨公司看車隊）
- 出庫 / 裝車確認（銜接 1.0 揀貨單）
- 司機簽到 / 離班 / 交接
- 重簽 / 作廢 POD（簽收時效操作）
- 同車次多車拆分（車次單過多拆多趟）
- 送達時窗（送達時段）
- 異常通報（事故 / 緊急停單）
