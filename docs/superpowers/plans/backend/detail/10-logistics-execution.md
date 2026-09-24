# Backend Detail — logistics 執行層（10-logistics-execution）

> 版本：v0.1.0（2026-09-17）
> 依據：決策 **D32**、需求規格 `docs/superpowers/specs/1.0-requirements/logistics-execution/spec.md`、規劃整合 `docs/PLANNING_OVERVIEW.md`。
> 定位：本文件為 logistics 執行層的細部分解；新增到 `plans/backend/detail/00-index.md` 的地圖。授權引擎採 **OpenFGA + RLS**（D32，取代本目錄其他文件對 Casbin/CASL 的描述；凡與 D32 衝突處以 D32 為準）。
> 編號：`10.x`。對應參考計畫 **Phase 5.5（Task 5.8–5.24,1.0）**（已補入 `plans/reference/2026-07-17-sales-order-1-0-tasks.md`）。

> **執行狀態（2026-09-24,D32 授權首批落地）**：
> - **已落地（10.9 送達回寫,2026-09-24）**：`CompleteDelivery` 同交易把**該車次**(route)上 `processing` 的每筆 `sales_order` 轉 `completed` ＋蓋 `delivered_at`（`00050`）＋寫 `sales_order_events`（payload 帶 `source=delivery`，與店家手動結案可區別）＋稽核；**任一筆失敗整交易回滾**（不留「配送完成但訂單沒結」）。只撈該車次且 processing 的訂單 —— pending/cancelled 不屬這台車，撈進來會讓無關訂單卡住配送完成。狀態沿用 05 狀態機（`processing → completed`），不新增狀態值。附帶補上**終態守衛**：已完成/已取消的配送不得重指派。
> - **已落地（10.6 配送執行與簽收,2026-09-24）**：狀態機 `pending → in_progress → completed`／`pending,in_progress → cancelled`（表驅動,`internal/services/logistics_statemachine.go`）＋`logistics_delivery_events`（append-only 軌跡,只授 SELECT/INSERT）＋`logistics_proofs`（POD,photo/signature/scan,引用 `file_assets`）＋delivery 的 `started_at`/`completed_at`（00048 建表+policy、00049 ENABLE+FORCE）;三支 RPC `StartDelivery`/`CompleteDelivery`/`CancelDelivery`（version 樂觀鎖、同交易寫事件與稽核、POD 逐筆驗歸屬後寫入 —— 任一失敗整筆回滾）;**本人閘門** = `deliverySelfGate`（`logistics_drivers.user_id` 比對指派欄位＋OpenFGA instance 級 `can_write` 雙重判定,非本人 403）;POD 上傳沿用 04 檔案端點新增 `owner_type=logistics_delivery`。
> - **已落地**：`logistics_drivers`/`vehicles`/`logistics_deliveries` 三表（00046 建表+policy+app_rw、00047 ENABLE+FORCE;**無 `logistics_teams` 聚合表、無 `uuid` 欄** —— 與本文件綱要有出入,依 repo 先例 bigserial 收斂,uuid 待統一裁決）、LogisticsService 四支 RPC（CreateDriver/CreateVehicle/AssignDelivery/ListMyDeliveries;指派 version 樂觀鎖＋D18 稽核）、OpenFGA model logistics 三型別（租戶 parent 邊 + `driver#assignee` userset;見 `third_party/openfga/model.go`）、tuple 同步（**AfterCommit 直寫對帳**,非佇列 outbox —— 指派事實可由 DB 重算,失敗僅記日誌,見 `internal/services/logistics_events.go`）、10.12/10.8 閘門（ListMyDeliveries 要求 logistics_drivers 列 + 逐列 instance Check）、rolePolicy `logistics`（dept_admin 讀寫/staff 讀）＋ protectedRPC 四條。
> - **未落地**：`logistics_teams` 聚合、`service_areas`/`zones`（PostGIS）、`positions` GPS 串流（10.5、10.7、10.10 geofence）、**10.11 通知接 D16**（指派/完成推播）、10.13 停點清單、部分交付語意（1.1+）、uuid 主鍵、department#member/#admin 成員 provision（目前 instance 判決靠 assigned_by/driver 邊;成員邊產製另案）。
> - **測試**：model 契約 5 子測試（`internal/authz/openfga/logistics_test.go`）、服務 8 測試（sqlite:建檔/指派/樂觀鎖/10.12 閘門 ＋ 10.6 轉移表/狀態機/POD/本人閘門,`logistics_service_test.go`＋`logistics_delivery_test.go`）、跨部門/本人隔離整合探針（app_rw + 真 RLS + 記憶體引擎,`logistics_rls_integration_test.go`,含 10.6 狀態機寫事件表與跨部門負向控制）全綠。

## 0. go8 架構嚴格遵循（D31 + 本領域）

本領域一律依 go8 / D31 結構慣例落地，不另發明目錄:
- **分層**:`backend/internal/domain/logistics/` 採 **handler → usecase → repository** 三層 + `register.go`（含 `transformation.go` DTO 轉換）;層間以 interface 解耦、經 `internal/server/domains.go` 的 `InitDomains()` 每 domain 一行組裝掛載（新增 domain 只動此檔）。
- **連線/外部服務初始化**:一律放 `backend/third_party/`（`third_party/openfga`、`third_party/osrm`、`third_party/geocode` 只做 client/連線建立）;行為邏輯(RLS hook、session 管理)留在 `internal/`。
- **組態**:逐檔 struct + `envconfig` tag(`backend/config/logistics.go` 放 `OSRM_HOST`、geocode 端點等;`config.New()` 聚合),不散落 env。
- **掛載與啟動**:Connect handler 掛 chi router 於 `InitDomains()`;fail-fast 啟動檢查集中 `Server.Init()`;`cmd/server` 極薄入口。
- **API**:一律 Connect-RPC(proto `v1`),不另走 REST/chi。

## 領域模型（部門級，Ent schema）

- `logistics_teams` / `vehicles` / `drivers`：均帶 `company_id` + `department_id` + `deleted_at`（軟刪除，D10）。`vehicles.plate_no` 部門內唯一（部分唯一索引）。`drivers.user_id` 關聯既有 `users`。
- `service_areas` / `zones`：部門級；`geometry` 用 PostGIS `geography(Polygon)`（fallback GeoJSON 文字）。
- `positions`：GPS 歷史；`driver_id`/`vehicle_id`/`company_id`/`department_id`；`location geography(Point)` + `timestamp` + `heading`/`speed`。
- `logistics_deliveries`：承接 `route_id`（車次）的執行單元；`driver_assigned_uuid`/`vehicle_assigned_uuid`/`assigned_by`/`status`（pending/dispatched/in_progress/completed/cancelled）/`started_at`/`completed_at`/`version`。
- `logistics_delivery_events`：執行軌跡（event_type + payload + created_at）。
- `proofs`：POD；`delivery_id`/`type`（photo/signature/scan）/`file_asset_id`（D17）/`captured_at`。

## 共通規則（沿用 00-index §3，另加下列）

- **RLS**：logistics 表 policy 以 `app.current_department_id` + `app.current_data_scope` 過濾（department 級;company_admin 看全部門、super 全公司）。runtime 依 Ent 官方 RLS 做法：`sql.WithStringVar(ctx, "app.current_department_id", ...)` / `WithStringVar` + `data_scope`,由 tenant middleware 注入。
- **OpenFGA**：受保護 RPC 進入點 `Check`;`list-objects` 做資源可見性;tuple 寫入走事件訂閱(與業務 handler 非同步,見 10.8)。
- **交易與稽核**：指派/狀態異動/簽收與對應稽核與事件軌跡同一 DB 交易(D18)。
- **錯誤處理**：沿用 00-index §3.4(`permission_denied` / `not_found` / `failed_precondition`(樂觀鎖) / `invalid_argument`(缺座標))。

---

## 領域模型 schema 細目（Fleetbase 逆向 → 1.0 收斂）

> 本節為 logistics 執行層各表的欄位級細目，**源自 Fleetbase 逆向 `docs/study/erd/go-logistics.mmd`（Fleetbase 倉庫）**，但已收斂為 1.0 語彙：`bigserial id + uuid`（D31）、`company_id/department_id`（F4）、D10 軟刪除（部分唯一索引）、D18 稽核。**不引入** Fleetbase 的 `_key`、多型 `*_type/*_uuid`、MySQL `VARCHAR` 金額/度量。以本節與下方各 10.x 為權威，Fleetbase 僅為欄位參考。

### logistics schemas 欄位綱要（目標 Ent）

- **vehicles**：`id bigserial PK` + `uuid unique`;`company_id`/`department_id`（mixin_tenant）;`plate_no`（部分唯一:同部門 + deleted_at IS NULL）、`vehicle_type`、`capacity`（含 `capacity_volume`/`capacity_pallets`/`capacity_parcels` 可選）、`make`/`model`/`year`、`vin`/`engine_number`、`fuel_type`/`fuel_volume_unit`、`odometer`/`odometer_unit`、`last_position geography(Point)`（快照,見 10.5）、`status`（狀態機）、`deleted_at`/`created_at`/`updated_at`。
- **drivers**：`id uuid`;`company_id`/`department_id`;`user_id`（關聯既有 `users`）;`name`/`phone`（可自 `users` 同步）;`assigned_vehicle_id`;`current_status`（online/offline/on_task）;`skills JSONB`;`location geography(Point)`（快照）;`slug`/`deleted_at`/`created_at`/`updated_at`。
- **logistics_teams**：`id uuid`;`company_id`/`department_id`;`name`/`color`/`task`/`status`/`parent_logistics_team_id`（自引用）;`deleted_at`/`created_at`/`updated_at`;關聯表 `logistics_drivers`（`(logistics_team_id, driver_id)` PK）、`logistics_team_vehicles`（`(logistics_team_id, vehicle_id)` PK）。
- **service_areas / zones**：`id uuid`;`company_id`/`department_id`;`name`/`type`/`color`/`stroke_color`;`border geography(Polygon)`（fallback GeoJSON 文字欄位）;`trigger_on_entry`/`trigger_on_exit`/`dwell_threshold_minutes`/`speed_limit_kmh`;`parent_id`（service_areas 自引用,階層）;zones 帶 `service_area_id`;`status`/`deleted_at`/`created_at`/`updated_at`。
- **positions（GPS 歷史,append-only）**：`id bigserial`;`company_id`/`department_id`;`driver_id`/`vehicle_id`（可其一或兩者）;`location geography(Point)` + `heading`/`speed`（`numeric`）;`timestamp timestamptz`;`created_at`。**不下 `deleted_at`**,`UNIQUE(uuid)` + `(company_id, created_at)` 索引 + 依時間分區（見 10.5）。**快照層**：`drivers.location`/`vehicles.last_position` 為 denormalized 當下位置,供看板秒查。
- **logistics_deliveries**：`id uuid`;`company_id`/`department_id`;`route_id`（承接車次,1.0 sales_order route）;`driver_assigned_id`/`vehicle_assigned_id`/`assigned_by`;`status`（pending/dispatched/in_progress/completed/cancelled）;`started_at`/`completed_at`;`delivery_sequence`（停點順序）;`version`（樂觀鎖,重指派遞增）;`reassigned_from_id`/`reassigned_to_id`（重指派軌跡,可空）;`deleted_at`/`created_at`/`updated_at`。
- **logistics_delivery_events**：`id uuid`;`logistics_delivery_id`;`event_type`（`driver.location_changed`/`order.driver_assigned`/`delivery.started`/`delivery.completed`/`delivery.cancelled` 等,D16 範本鍵）;`payload JSONB`;`created_at`。**append-only**（不下 deleted_at）。
- **proofs（POD）**：`id uuid`;`logistics_delivery_id`;`type`（photo/signature/scan）;`file_asset_id`（D17 檔案資產）;`captured_at timestamptz`;`remarks`;`deleted_at`/`created_at`/`updated_at`。

> 對照清單:完整逐欄建議見 Fleetbase 倉庫 `docs/study/erd/go/schema-improvement.md` 域 D；本節僅取 1.0 承接部分並收斂。

---

### 10.1 建立 logistics 主檔 schema（logistics_teams / vehicles / drivers）

- **目標**：建立部門級 logistics 主檔的 Ent schema + RLS policy + 建檔 RPC。
- **檔案**：`ent/schema/logistics.go`、`vehicle.go`、`driver.go`;migration:`database/goose/xxx_logistics_master.sql`(含 `ENABLE ROW LEVEL SECURITY` + `CREATE POLICY logistics_dept_isolation`);`proto/v1/logistics.proto`(`LogisticsService`)。
- **介面**：`LogisticsService.CreateVehicle` / `CreateDriver` / `ListVehicles`(分頁 + filter)/ `ListDrivers` / `SoftDeleteVehicle|Driver`。
- **實作邏輯**：Ent schema 帶 `mixin_tenant`(company/department);建檔時驗證 `department_id` 在操作者 data_scope 內(OpenFGA `Check`);`plate_no` 用部分唯一索引;司機建立時關聯既有 `users`(不另建 user)。RLS policy 見共通規則。
- **錯誤處理**：`already_exists`(plate_no 重複)、`permission_denied`(跨部門)、`not_found`(department 不存在)。
- **驗收**：dept_admin 建車/建司機成功、部門乙查不到;同部門重建 deleted 車牌成功;RLS 跨部門查詢回空。

### 10.2 service_areas / zones

- **目標**：地理服務範圍與區域(部門級)CRUD。
- **檔案**：`ent/schema/service_area.go`、`zone.go`;migration(PostGIS Polygon);`proto/v1/logistics.proto`(service_areas/zones 於 `LogisticsService`)。
- **介面**：`LogisticsService.CreateServiceArea` / `ListZones` / `ZoneBelongs`(檢查點是否落在 zone)。
- **實作邏輯**：Polygon 以 `geography` 儲存;建立時若 PostGIS 不可用 fallback GeoJSON 文字欄位;`ZoneBelongs` 用空間函數,座標缺失回 `invalid_argument`。
- **錯誤處理**：`invalid_argument`(座標/邊界格式錯)、`permission_denied`(跨部門)。
- **驗收**：建 service_area + 加入 zone;點落在 zone 內回 true、跨部門查不到。

### 10.3 客戶送貨地址座標（master-data 擴充）

- **目標**：`customer_addresses`(`type=shipping`)取得座標供 OSRM。
- **檔案**：`ent/schema/customer_address.go`(加 `location geography(Point)` + fallback `latitude/longitude`);migration;geocode 服務於 `third_party/geocode`(抽象 adapter);master-data 建檔 handler。
- **介面**：沿用既有 `customer_address` 建/改 RPC;新增欄位 `location`。
- **實作邏輯**：建立/更新 `type=shipping` 地址時呼叫 geocode;成功寫座標、失敗仍建檔(`location` 空);非 shipping 不強制。
- **錯誤處理**：geocode 失敗不回丟建檔錯誤;於 logistics 路線計算時才提示缺座標。
- **驗收**：shipping 地址 geocode 成功寫座標;geocode 失敗仍建檔且 `location` 空。
- **相依**：`10.3` 提供 logistics 路線(10.7)所需的終點座標。

### 10.4 車次 ↔ 車輛 / 司機指派（AssignmentService）

- **目標**：後台把車次指派實際車輛/司機,粒度在 `logistics_delivery`。
- **檔案**：`ent/schema/logisticsdelivery.go`;`internal/domain/logistics`(handler/usecase/repository/register);`proto/v1/logistics.proto`(`AssignmentService.AssignRouteDelivery` / `ReassignVehicle`)。
- **介面**：`AssignmentService.AssignRouteDelivery(route_id, driver_uuid, vehicle_uuid, version)`。
- **實作邏輯**：由 1.0 已裝單完成的 `route_id` 建立/更新 `logistics_delivery`;寫 `driver_assigned_uuid`/`vehicle_assigned_uuid`/`assigned_by`;樂觀鎖比對 `version`(衝突重試/拒絕);完成後發 `order.driver_assigned`(Valkey pub/sub 通知司機 App)。**非車次↔車 1:1**,可重指派(version 遞增、寫稽核)。
- **錯誤處理**：`failed_precondition`(version 衝突)、`permission_denied`(非後台授權角色)、`not_found`(車次/司機/車不存在)。
- **驗收**：指派成功司機 App 收到;並發指派衝突拒絕;重指派 version 遞增 + 稽核。

### 10.5 即時定位（TrackingService + positions）

- **目標**：司機上報位置並以 Connect 串流 + Valkey pub/sub 廣播 `driver.location_changed`。
- **檔案**：`ent/schema/position.go`;`internal/domain/logistics/tracking.go`;`proto/v1/tracking.proto`(`TrackingService.SubmitPosition` / `SubscribePositions`)。
- **介面**：`SubmitPosition(location, heading, speed)`;`SubscribePositions()`(server-streaming)。
- **實作邏輯**：`SubmitPosition` 寫 `positions`(RLS 部門級);同一交易內同時更新**快照層** `drivers.location`(或 `vehicles.last_position`)為當下位置(供看板秒查,不做歷史累加);`SubscribePositions` 訂閱部門 channel(Valkey pub/sub 跨 replica,D14);事件僅作失效提示;斷線重連全量重查、連續失敗降級 30 秒輪詢。認證同其他 RPC(cookie/token),無一次性 ticket。
- **positions 高頻表設計**：`positions` 為 **append-only**(不下 `deleted_at`,D10 不適用)、`UNIQUE(uuid)`+`(company_id, created_at)` 索引、**依 `created_at` 時間分區**(如按月);RLS policy 與分區鍵一起存在(避免跨租戶掃全表);歷史保留策略於 Phase 8 維運(D27 同理)。快照(當下位置)由 `drivers.location`/`vehicles.last_position` 承載,`positions` 僅存軌跡供稽核/回播(1.1+ 回播功能)。從 Fleetbase 逆向:定位欄位 `geography(Point)` + `heading`/`speed` 為 `numeric`(不存字串)。
- **錯誤處理**：`unauthenticated`(無憑證)、`permission_denied`(跨部門)、`invalid_argument`(座標缺失)。
- **驗收**：司機上報 → 後台連線收到 `driver.location_changed`;跨部門連線不收;未認證拒連;斷線重連補齊。

### 10.6 配送執行與簽收（DeliveryService + POD）

- **目標**：司機執行配送與簽收。
- **檔案**：`ent/schema/proof.go`、`logistics_delivery.go` 狀態欄位;`proto/v1/logistics.proto`(`DeliveryService.Start / Complete / Cancel`);POD 走檔案資產(D17)。
- **介面**：`DeliveryService.Start(delivery_id)` / `Complete(delivery_id, proof)` / `Cancel(delivery_id, reason)`。
- **實作邏輯**：狀態機 `pending→in_progress→completed`,`cancelled`(含原因);完成時寫 `proofs`(photo/signature/scan + file_asset_id);狀態異動 + 稽核 + `logistics_delivery_events` 同一交易;OpenFGA 判定操作者為被指派司機。
- **錯誤處理**：`permission_denied`(非本人)、`failed_precondition`(狀態機不允許)、`invalid_argument`(缺 POD 或原因)。
- **驗收**：完成 + 上傳 photo POD 成功、寫稽核與事件;非被指派司機操作被拒。

### 10.7 路線 / ETA（OSRM）

- **目標**：依車次與客戶 shipping 地址座標算路線/ETA。
- **檔案**：`third_party/osrm`(client,`OSRM_HOST` 設定);`internal/domain/logistics/routing.go`;`proto/v1/logistics.proto`(`RoutingService.GetRoute`)。
- **介面**：`RoutingService.GetRoute(delivery_id)` → 起終點座標 + 路線 + ETA。
- **實作邏輯**：起點取車次/倉別、終點取 `customer_addresses(type=shipping)` 座標;缺座標回 `invalid_argument`;呼叫 OSRM `/route/v1/driving/...`。
- **錯誤處理**：`invalid_argument`(缺座標)、`unavailable`(OSRM 不可用/failover 自建 container)。
- **驗收**：全地址有座標 → 回路線 + ETA;某地址缺座標 → 提示缺座標。

### 10.8 OpenFGA 授權接入（logistics 資源）

- **目標**：logistics 資源納入 OpenFGA 決策。
- **檔案**：`third_party/openfga`(client + store 初始化);`internal/authz`(Check/list-objects middleware);tuple 寫入 subscriber(事件 → outbox → 寫 OpenFGA,不於 handler 同步)。
- **介面**：middleware 對受保護 logistics RPC `Check`;list 用 `list-objects`。
- **實作邏輯**：型別 = `company`/`department` 租戶 + `role`/`group`/`system`;資源 `driver`/`vehicle`/`logistics_delivery` 以租戶 parent 邊 + userset rewrite;指派/建檔時寫 tuple(經事件流)。
- **錯誤處理**：`permission_denied`(Check false)。
- **驗收**：被指派司機可操作其 delivery、他人 403;`list-objects` 對所屬部門回正確集合。
- **相依**：本子功能為 logistics 全部 RPC 的授權前置(bootstrap 於 `Server.Init()`)。

---

### 10.9 每單送達狀態回寫（A,1.0）

- **目標**：`logistics_delivery` 完成時把送達回寫到該車次所載每筆 `sales_order`（標 `delivered`）。
- **檔案**：`internal/domain/logistics/delivery.go`（完成流程內）；update `ent`（`sales_order.status` 沿用既有狀態機）；`proto/v1/sales_order.proto`（查詢帶 `delivered_date`）。
- **介面**：於 `DeliveryService.Complete` 完成後同交易逐筆回寫車次所載 `sales_order`。
- **實作邏輯**：完成時依 `route_id` 取該車次 `sales_order` 清單 → 逐筆標 `delivered` + 寫 `sales_order_events` + 稽核（同交易,D18）。部分交付語意排 1.1+。
- **錯誤處理**：任一訂單回寫失敗 → 整交易回滾，避免部分送達半套。
- **驗收**：完成車次後,店家/客戶查得到每單 `delivered` 與 `sales_order_events`。

### 10.10 geofence 自動送達判定（A,1.0）

- **目標**：依司機位置 + `service_area`/`zone` 邊界自動判定到站。
- **檔案**：`internal/domain/logistics/geofence.go`；`third_party/geocode`（已含空間計算）；事件寫 `logistics_delivery_events`。
- **介面**：`TrackingService.SubmitPosition` 內部判定（或獨立 `GeofenceService.CheckArrival`）。
- **實作邏輯**：位置寫入時以空間函數比對該配送點所在 zone；進入且未完成 → 標到站、寫事件、提示司機可 POD。
- **錯誤處理**：缺座標 → 不做自動判定（回退人工標記），不擋主流程。
- **驗收**：司機進入配送點 zone 自動標到站並寫事件。

### 10.11 logistics 通知接 D16（A,1.0）

- **目標**：reuse 既有通知（D16：FCM+站內、範本/裝置管理），不另建一套。
- **檔案**：`internal/domain/logistics/notify.go`（訂閱 logistics 事件 → 叫 `notifications` domain）；沿用 D16 範本。
- **介面**：指派完成發司機新任務通知;送達回寫後推店家/主責業務（`default_sales_rep_id`）。
- **實作邏輯**：於指派與完成事件觸發 D16 通知路由;失敗僅標 `failed`（沿用 D16,不重試）。
- **驗收**：指派推司機、送達推店家/業務,寫 `notifications` 記錄。

### 10.12 司機身分 / 角色（1.0,負載關鍵）

- **目標**：`drivers` 關聯 `users`,以 OpenFGA `driver` relation + RLS data_scope(department/self)賦權;司機以既有 JWT 登入只取自己被指派的任務。
- **檔案**：`internal/domain/logistics/driver.go`;`internal/authz`(openfga model 增 `driver` relation);`config/logistics.go`。
- **介面**：於認證 middleware 後以 OpenFGA 判定 `assignee`;`ListMyDeliveries`。
- **實作邏輯**：建立 `drivers` 時以事件流寫 OpenFGA `driver`/`assignee` tuple;App 登入 JWT 帶身分 → list-objects 限定自己部門+自己指派。
- **錯誤處理**：非 driver 角色嘗試司機端點 → `permission_denied`。
- **驗收**：司機登入僅見被指派的任務;它人 403。

### 10.13 配送停點 / waypoint 清單（1.0 簡版）

- **目標**：`logistics_delivery` 展開為依 `delivery_sequence` 排序的停點(客戶/地址/訂單/座標),供司機 App 逐站執行。
- **檔案**：`ent`(或由 `logistics_deliveries` + `sales_order` join 產生,不加持久表,簡版以查詢產出)。
- **介面**：`DeliveryService.GetStops(delivery_id)` → 停點清單。
- **實作邏輯**：依 `route_id` + `delivery_sequence` 取該車次訂單與 `customer_addresses` 座標,組停點序列。
- **錯誤處理**：缺座標停點標示缺座標不擋清單。
- **驗收**：司機取得依序停點清單(客戶/地址/訂單/座標)。

> 司機 App 模式（#2）與後台車隊/執行地圖（#3）為「跨端」項目：App 模式屬 Phase 6（App 功能）司機端任務,後台頁屬 Phase 5.5 的 console 任務（見 reference Task 5.21–5.24）。

## 1.1+ 延伸（v1 後,不擠入 1.0）

自動派單、VRP 多車次優化、現場拒收/部分交付(串 D25)、ETA 異常通知、customer 端 live map、車輛狀態/保養、司機 App 背景 GPS/離線、每單 tracking 狀態鏈、位置軌跡回播/稽核、logistics 報表/績效、多公司 logistics 視圖、出庫/裝車確認、司機簽到離班、重簽/作廢 POD、同車次多車拆分、送達時窗、異常通報——均列 `logistics-execution/spec.md` §1.1+ 延伸,1.0 不做。

## 整合測試重點（D21）

RLS 部門隔離(logistics 表)、OpenFGA Check/list-objects、指派樂觀鎖並發、狀態機轉移、POD 上傳 + 稽核、地理編碼失敗不擋建檔、Valkey 串流跨 replica 送達。
