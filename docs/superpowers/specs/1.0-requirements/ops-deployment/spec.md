# ops-deployment 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Purpose

定義多公司訂出貨系統 1.0 的開發環境、生產部署、CI/CD、測試門檻、備份、監控告警、災難復原與 Big Bang 上線策略，確保系統可重現地建置、部署、維運與復原。

## Requirements

### Requirement: 開發環境一鍵啟動

系統 MUST 提供 docker-compose 定義，讓開發者以單一指令啟動完整的本機依賴服務，且 SHALL 至少包含 PostgreSQL、Valkey、Gotenberg 三個服務（1.0 不使用 Email，無 Mailpit）；OAuth2 使用 Google Workspace 測試 client，開發環境 MUST NOT 包含自建 IdP。

#### Scenario: 啟動本機開發依賴

- **WHEN** 開發者於全新機器 clone 倉庫後執行開發環境啟動指令
- **THEN** PostgreSQL、Valkey、Gotenberg 全部以容器啟動且可被本機後端連線
- **AND** 後端可成功執行資料庫 migration 並通過健康檢查

#### Scenario: 無自建 IdP 依賴

- **WHEN** 檢視開發環境 compose 定義
- **THEN** 不得包含 Authelia / Authentik 等自建 IdP 服務
- **AND** OAuth2 登入流程 MUST 指向 Google Workspace 測試 client

### Requirement: 統一工作入口與建置管線

倉庫 MUST 採用 pnpm workspace + Turborepo 管理全倉任務管線與快取（含前端建置、proto 型別同步、測試），並以 Task 指令作為 Go / Flutter 等原生任務的統一入口；開發者 SHALL 不需記憶各端原生命令即可執行常見任務。

#### Scenario: 透過統一入口執行測試

- **WHEN** 開發者在倉庫根目錄執行統一的測試任務指令
- **THEN** 後端、前端、App 的測試 SHALL 被依序或平行觸發
- **AND** 任一端測試失敗時整體任務 MUST 回傳非零結束碼

#### Scenario: proto 變更後型別同步

- **WHEN** 開發者修改 Connect-RPC proto 定義後執行型別同步任務
- **THEN** 後端（Go）、前端（TypeScript）、App（Dart）的產生碼 SHALL 一併更新
- **AND** Turborepo 快取 MUST 使未變更的套件跳過重複建置

### Requirement: 生產 Kubernetes 部署與流量轉發

生產環境 MUST 部署於 Kubernetes，以 Traefik 作為 ingress；PostgreSQL 與 Valkey MUST 以 StatefulSet 容器化部署於同叢集；Traefik SHALL 依 path 轉發 REST（`/api/v1/*`）與 Connect-RPC（含 server streaming，HTTP/2，走 443 port）流量至後端服務，且 MUST NOT 因 idle timeout 主動切斷已建立的 server streaming 連線。

#### Scenario: Connect-RPC 流量經 Traefik 轉發

- **WHEN** Web 或 App 客戶端向生產網域發出 Connect-RPC 請求
- **THEN** 請求 MUST 經由 Traefik 依 path 路由至後端服務
- **AND** 連線 MUST 使用 TLS 443 port 並支援 HTTP/2

#### Scenario: server streaming 長連線維持

- **WHEN** Web 客戶端對派車看板建立 Connect server streaming 連線
- **THEN** Traefik MUST 將該請求轉發至後端且不緩衝串流回應
- **AND** 連線建立後 SHALL 能持續接收串流事件而不被 ingress 因 idle timeout 中斷

#### Scenario: 資料庫以 StatefulSet 運行

- **WHEN** 檢視生產叢集的工作負載定義
- **THEN** PostgreSQL 與 Valkey MUST 以 StatefulSet 部署並掛載持久化儲存
- **AND** Pod 重建後資料 SHALL 不遺失

### Requirement: GitHub Actions CI/CD 管線

倉庫 MUST 提供 GitHub Actions CI/CD 管線，依序執行測試、建置容器 image、部署至 Kubernetes；測試未通過時 MUST 阻擋後續建置與部署。

#### Scenario: 測試失敗阻擋部署

- **WHEN** 任一端的測試或覆蓋率門檻於 CI 中失敗
- **THEN** 管線 MUST 中止
- **AND** 不得產生可部署的 image，也不得觸發 k8s 部署

#### Scenario: 主線合入後自動部署

- **WHEN** 程式碼合入主線且測試全部通過
- **THEN** CI SHALL 建置後端與前端容器 image 並推送至 image registry
- **AND** 管線 SHALL 將新版本部署至 Kubernetes 並完成滾動更新

### Requirement: 測試覆蓋率與關鍵路徑測試門檻

後端（Go testify）、前端（Vitest）、App（flutter_test）三端的單元測試覆蓋率門檻 MUST 皆為 70%，由 CI 強制檢查；關鍵路徑（認證、授權 / RLS、訂單狀態機、樂觀鎖取號、退貨審核、偏好送貨日順延）MUST 具備整合測試；App MUST 以 Maestro 執行整合測試。

#### Scenario: 覆蓋率低於門檻時 CI 失敗

- **WHEN** 任一端的單元測試覆蓋率低於 70%
- **THEN** CI 管線 MUST 判定失敗
- **AND** 該次變更不得合入主線或進入部署階段

#### Scenario: 關鍵路徑整合測試存在且執行

- **WHEN** CI 執行後端整合測試
- **THEN** 認證、授權 / RLS、訂單狀態機、樂觀鎖取號、退貨審核、偏好送貨日順延六條關鍵路徑 SHALL 各有至少一個整合測試被執行
- **AND** 整合測試 MUST 以真實 PostgreSQL（如 dockertest）執行，不得僅以 mock 取代資料層

#### Scenario: App 端對端流程測試

- **WHEN** CI 執行 App 的 Maestro 測試
- **THEN** 登入與快速下單等核心流程 SHALL 於模擬器／真機環境自動化跑完
- **AND** 任一流程步驟失敗時測試 MUST 回報失敗

### Requirement: 備份策略與保留期限

系統 MUST 依下表執行備份，所有備份 SHALL 存放於 Google Cloud Storage（k8s 設定另存 Git）：

| 資料 | 備份方式 | 頻率 | 保留期限 |
|---|---|---|---|
| PostgreSQL | 每日完整備份 + WAL 歸檔（PITR） | 每日 | 每日備份 30 天，每月備份 12 個月 |
| Valkey | RDB 快照 + AOF | 每小時 RDB，持續 AOF | 7 天 |
| 本地檔案儲存（NFS / volume） | restic | 每日 | 30 天 |
| Kubernetes manifests / 設定 | Git 版控 + 加密 secrets 匯出 | 每次變更 | 永久 |

#### Scenario: PostgreSQL 每日備份與 WAL 歸檔

- **WHEN** 到達每日排程時間
- **THEN** 系統 MUST 產生 PostgreSQL 完整備份並上傳 GCS
- **AND** WAL 歸檔 SHALL 持續寫入，使資料庫可還原至保留期間內任意時間點（PITR）

#### Scenario: 備份保留期限自動清理

- **WHEN** PostgreSQL 每日備份超過 30 天、Valkey 備份超過 7 天、restic 檔案備份超過 30 天
- **THEN** 過期備份 MUST 被自動清理
- **AND** PostgreSQL 每月備份 SHALL 例外保留 12 個月

#### Scenario: k8s 設定變更可追溯

- **WHEN** Kubernetes manifests 或叢集設定變更
- **THEN** 變更 MUST 以 Git 版控提交
- **AND** secrets SHALL 以加密形式匯出備份，不得以明文入庫

### Requirement: 監控指標、日誌與告警

系統 MUST 於同叢集部署 Prometheus + Grafana + Alertmanager；後端 MUST 暴露 `/metrics` 端點，涵蓋基礎設施指標（CPU、記憶體、磁碟、網路、Pod 重啟次數）、應用程式指標（API 延遲、錯誤率、請求量、Connect-RPC 狀態碼）與業務指標（登入失敗次數、訂單建立量、列印次數、稽核日誌異常查詢）；日誌 MUST 以結構化 JSON 集中收集並保留 30 天；告警 MUST 支援 Email / Slack / Webhook 通道。

#### Scenario: 後端暴露應用程式指標

- **WHEN** Prometheus 抓取後端 `/metrics` 端點
- **THEN** 回應 MUST 包含 API 延遲、錯誤率、請求量與 Connect-RPC 狀態碼指標
- **AND** 業務指標 SHALL 包含登入失敗次數、訂單建立量與列印次數

#### Scenario: 異常狀況觸發告警

- **WHEN** 發生服務不可用、錯誤率驟升、備份失敗或磁碟空間不足
- **THEN** Alertmanager MUST 發出告警
- **AND** 告警 SHALL 可經 Email、Slack、Webhook 任一已設定通道送達維運人員

#### Scenario: 結構化日誌集中收集

- **WHEN** 後端或基礎設施元件輸出日誌
- **THEN** 日誌 MUST 為結構化 JSON 並集中收集
- **AND** 超過 30 天的日誌 SHALL 被自動清除

### Requirement: 災難復原目標與演練

系統災難復原 MUST 達成 RTO 4 小時、RPO 1 小時（依賴 WAL 歸檔）；復原流程 SHALL 依序為：確認災難範圍與備份可用性 → 從 GCS 最新備份還原 PostgreSQL 與 Valkey → 重新部署 k8s 服務與 ingress → 驗證應用程式健康狀態與關鍵業務流程 → 切換 DNS 或調整負載均衡器；MUST 每半年執行一次災難復原演練並更新復原手冊，且上線前 MUST 完成首次 PITR 還原演練並留存紀錄。

#### Scenario: 依五步流程完成復原

- **WHEN** 發生需整體復原的災難事件
- **THEN** 維運團隊 MUST 依確認範圍 → GCS 還原 → 重新部署 → 驗證 → 切換 DNS 五步驟執行
- **AND** 自復原啟動至服務恢復 SHALL 不超過 4 小時
- **AND** 還原後資料遺失 SHALL 不超過 1 小時

#### Scenario: 上線前首次 PITR 演練

- **WHEN** 系統尚未正式上線且準備進入上線階段
- **THEN** 維運團隊 MUST 已使用 WAL 歸檔完成至少一次 PITR 還原演練，驗證 RTO 4 小時 / RPO 1 小時可達成
- **AND** 演練過程與結果 SHALL 留存書面紀錄；未完成演練不得上線

#### Scenario: 半年度災難復原演練

- **WHEN** 距前一次災難復原演練達六個月
- **THEN** 維運團隊 MUST 執行一次完整 DR 演練
- **AND** 演練後 SHALL 依演練發現更新復原手冊

### Requirement: Big Bang 上線策略

系統 MUST 採 Big Bang 上線：MUST NOT 從舊系統匯入任何資料（客戶、商品等主檔與訂單皆於新系統重新建檔）；MUST NOT 有試點、新舊系統並行運作或舊系統唯讀維護期；測試完成後 SHALL 直接全面上線並廢止舊系統；上線檢查清單 MUST 包含確認生產環境 `DEVELOPER_ACCOUNT_ENABLED=false`。

#### Scenario: 上線不匯入舊系統資料

- **WHEN** 新系統正式上線
- **THEN** 新系統資料庫 MUST 僅含 migration seed 與使用者新建資料
- **AND** 不得存在任何自舊系統匯入的主檔或訂單資料

#### Scenario: 無並行與唯讀期

- **WHEN** 新系統完成測試並切換上線
- **THEN** 舊系統 MUST 同時廢止
- **AND** 不得設置試點期、雙系統並行期或舊系統唯讀維護期

#### Scenario: 上線檢查清單含開發者帳號關閉

- **WHEN** 執行上線前檢查
- **THEN** 檢查清單 MUST 包含確認生產環境 `DEVELOPER_ACCOUNT_ENABLED=false`
- **AND** 該項目未確認通過時 SHALL 不得放行上線
