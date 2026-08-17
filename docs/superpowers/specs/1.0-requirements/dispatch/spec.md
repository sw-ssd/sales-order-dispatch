# dispatch 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Requirements

### Requirement: 派車看板檢視

系統 SHALL 提供 Web 專屬的派車看板（Kanban）：以車次（`route_id`）為欄、訂單為卡片，卡片依 `delivery_sequence` 排序顯示同車次內的配送順序；未指派車次的待派訂單 SHALL 顯示於獨立的「未指派」區域。看板 SHALL 依 `expected_delivery_date` 提供日期篩選，僅顯示所選日期預計出貨的訂單。

#### Scenario: 依日期篩選檢視看板

- **WHEN** 使用者開啟派車看板並選擇預計出貨日 `expected_delivery_date = 2026-07-20`
- **THEN** 看板僅顯示 `expected_delivery_date` 為 2026-07-20 且狀態為 `pending` 或 `processing` 的訂單
- **AND** 每張訂單卡片依其 `route_id` 歸入對應車次欄位，同欄內依 `delivery_sequence` 升序排列

#### Scenario: 訂單僅對應單一車次

- **WHEN** 一筆訂單已被指派 `route_id`
- **THEN** 該訂單 SHALL 僅出現在該車次欄位，不得同時出現在其他車次或拆分成多張卡片
- **AND** 系統 SHALL 拒絕將同一筆訂單同時指派至多個車次的請求

#### Scenario: 跨部門資料隔離

- **WHEN** 使用者檢視派車看板
- **THEN** 看板 SHALL 僅顯示該使用者權限資料範圍（`data_scope`）內所屬部門的訂單與車次，不得顯示其他部門的資料

### Requirement: 拖放指派與車內排序

系統 SHALL 支援在看板上以拖放方式調整訂單的所屬車次（`route_id`）與車內配送順序（`delivery_sequence`），包含從「未指派」拖入車次、跨車次移動、以及同車次內排序。拖放提交 SHALL 採樂觀鎖：請求攜帶讀取時的 `sales_orders.version`，後端比對一致才更新，且更新後 `version` 遞增。

#### Scenario: 拖放變更車次與順序成功

- **WHEN** 使用者將訂單 A 從車次 R1 拖放至車次 R2 的第 3 順位，且提交時訂單 A 的 `version` 與伺服器一致
- **THEN** 系統 SHALL 更新訂單 A 的 `route_id` 為 R2、`delivery_sequence` 為 3，並遞增 `version`
- **AND** 車次 R2 原有順位 ≥ 3 的訂單 SHALL 依序後移，看板顯示新的排序結果

#### Scenario: 樂觀鎖衝突拒絕提交

- **WHEN** 使用者提交拖放結果時，訂單 A 的 `version` 已因他人先行的變更而不同
- **THEN** 系統 SHALL 拒絕本次提交，不變更任何資料
- **AND** 系統 SHALL 回傳衝突錯誤，前端 SHALL 提示使用者資料已被他人變更並重新整理看板至最新狀態

### Requirement: 批次派車確認

系統 SHALL 以車次為批次單位執行派車確認：一次確認該車次於所選日期內的所有待派訂單。批次內每筆訂單 SHALL 各自記錄 `dispatched_at` 與 `dispatched_by`，狀態由 `pending` 轉為 `processing`，並寫入 `sales_order_events`（`event_type = dispatch`）。確認完成後系統 SHALL 觸發派車通知（通知之通道、範本與發送機制屬 `notifications` capability）。

#### Scenario: 車次批次確認成功

- **WHEN** 使用者對車次 R1 執行派車確認，R1 當日有 5 筆 `pending` 訂單
- **THEN** 系統 SHALL 將該 5 筆訂單狀態轉為 `processing`，並逐筆記錄相同的 `dispatched_at` 與操作者的 `dispatched_by`
- **AND** 每筆訂單 SHALL 各寫入一筆 `sales_order_events`（`event_type = dispatch`）
- **AND** 系統 SHALL 對每筆訂單觸發派車通知

#### Scenario: 批次內訂單狀態已不符

- **WHEN** 使用者執行批次確認時，批次中某筆訂單已被取消或已派車
- **THEN** 系統 SHALL 拒絕該批次的確認或明確回報該筆訂單失敗原因，不得對狀態非 `pending` 的訂單重複寫入 `dispatched_at`
- **AND** 已成功處理的訂單與失敗的訂單 SHALL 在回應中明確區分

#### Scenario: 未指派車次的訂單不納入批次

- **WHEN** 看板上存在未指派車次的 `pending` 訂單
- **THEN** 該訂單 SHALL NOT 被任何車次的批次確認納入，維持 `pending` 狀態

### Requirement: 取消派車

系統 SHALL 允許 `dept_admin` 以上角色將 `processing` 狀態的訂單取消派車，且 MUST 填寫原因。取消派車 SHALL：清除 `dispatched_at` 與 `dispatched_by`、狀態退回 `pending`、保留原 `route_id` 與 `delivery_sequence`（訂單停留看板原車次原位待重新派車）；同時寫入 `sales_order_events`（`event_type = dispatch_cancel`）與 `audit_logs`（`action = dispatch_cancel`）。

#### Scenario: 取消派車成功並保留看板位置

- **WHEN** `dept_admin` 對一筆 `processing` 訂單執行取消派車並填寫原因「派錯車次」
- **THEN** 系統 SHALL 清除該訂單的 `dispatched_at` / `dispatched_by` 並將狀態退回 `pending`
- **AND** 系統 SHALL 保留原 `route_id` 與 `delivery_sequence`，訂單在看板上停留原車次原順位
- **AND** 系統 SHALL 寫入一筆 `sales_order_events`（`event_type = dispatch_cancel`，含原因）與一筆 `audit_logs`（`action = dispatch_cancel`）

#### Scenario: 未填原因拒絕取消

- **WHEN** 使用者執行取消派車但未填寫原因
- **THEN** 系統 SHALL 拒絕操作並提示原因為必填，訂單狀態與欄位維持不變

#### Scenario: 車次已有正式列印記錄時提示重印

- **WHEN** 使用者取消派車的訂單所屬車次已存在該日期的正式列印記錄（`print_logs`）
- **THEN** 系統 SHALL 在確認前提示該車次已列印、取消後需重新列印相關單據
- **AND** 使用者確認後取消派車照常執行

#### Scenario: 權限不足拒絕取消

- **WHEN** `staff` 或更低權限角色嘗試取消派車
- **THEN** 系統 SHALL 拒絕操作並回傳權限不足錯誤，訂單維持 `processing` 狀態

### Requirement: 看板即時推播與輪詢降級

系統 SHALL 以 Connect-RPC server streaming 提供派車看板的即時更新：後端 MUST 提供看板訂閱的 streaming RPC（屬 `DispatchService`），於拖放指派、批次派車確認、取消派車提交後，向訂閱同部門看板的連線推送型別化事件（proto message，至少含 `sales_order_id`、`route_id`、`delivery_sequence`、`version`）。串流連線 SHALL 走與其他 RPC 相同的認證（Web 以 httpOnly cookie、App 以 token header），MUST NOT 要求一次性 ticket 或 URL 參數憑證。事件廣播 SHALL 依部門隔離，且 MUST 在後端任意 replica 數下正確送達（以 Valkey pub/sub 跨 replica 轉發）。前端收到事件後 SHALL 使看板查詢失效並全量重查（事件僅作失效提示，MUST NOT 直接修改快取資料）。串流斷線時前端 SHALL 自動重連並於重連後全量重查；串流連續建立失敗時 SHALL 降級為定時輪詢（預設 30 秒，視窗隱藏時暫停、聚焦時立即重查），並於串流可建立時恢復。

#### Scenario: 他人操作經串流即時反映

- **WHEN** 使用者甲完成一筆拖放指派，同部門的使用者乙正開啟同日期看板且串流連線正常
- **THEN** 乙的連線 SHALL 收到該訂單變更事件（含最新 `route_id`、`delivery_sequence` 與 `version`）
- **AND** 乙的前端 SHALL 使看板查詢失效並全量重查，顯示伺服器最新狀態

#### Scenario: 未認證串流連線被拒絕

- **WHEN** 客戶端未攜帶有效 cookie 或 token 建立看板串流連線
- **THEN** 系統 SHALL 拒絕該連線，不建立看板訂閱

#### Scenario: 跨部門連線不接收事件

- **WHEN** 部門 A 發生派車變更，而部門 B 的使用者持有有效串流連線
- **THEN** 系統 SHALL NOT 向部門 B 的連線發送部門 A 的派車事件

#### Scenario: 跨 replica 事件送達

- **WHEN** 派車變更在後端 replica 1 提交，而看板串流連線建立在 replica 2
- **THEN** 事件 SHALL 經 Valkey pub/sub 轉發至 replica 2 並送達該連線

#### Scenario: 斷線重連補齊最新狀態

- **WHEN** 看板串流斷線，期間發生派車變更，隨後前端重連成功
- **THEN** 前端 SHALL 於重連後立即全量重查看板，顯示包含斷線期間變更的最新狀態

#### Scenario: 串流不可用時降級輪詢

- **WHEN** 前端連續建立串流失敗達降級閾值
- **THEN** 前端 SHALL 改以定時輪詢更新看板（視窗隱藏時暫停、聚焦時立即重查）
- **AND** 串流可再次建立時 SHALL 恢復串流推播
