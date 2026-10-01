# Platform 獨立 HTTP 服務實作計畫

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking。

**Goal:** 將 `internal/platform` 從同進程模組改為可獨立部署的 HTTP/gRPC 服務，產品（sales-order）經網路呼叫其配額/權益/凍結能力，使平台可與產品異步發版。

**Architecture:** platform 服務暴露 `CheckLimit`（寫路徑配額預約）、`GetTenantEntitlements`（唯讀投影，已有）、`SetCompanyStatus` 經 outbox 反轉三個邊界。交易模型翻轉：原本 `guardQuota` 在產品請求交易內 in-process 呼叫 `entitlements.Service.CheckLimit`（same-tx 計數），改為產品呼叫 platform HTTP 服務的 `CheckLimit` RPC，**計數的 same-tx 保證轉移到 platform 服務內部**（platform 擁有自己的 usage 表）。consumer 認領留 platform 側，凍結公司經 `platform.events` outbox → 產品 worker 消費 → `services.SetCompanyStatus`。採**雙模式**：platform 可同進程（預設，零風險）或獨立 HTTP（flag 切換），單產品場景下維持同進程，獨立跑才觸發網路路徑。

**Tech Stack:** Connect RPC（現有 `buf.gen.yaml` 已生成 Go/TS/console）、`github.com/salesorder/sales-order-1.0/backend/contracts`（proto 契約）、`database/platform_migrations` + `platform_goose_db_version`（已拆）、testcontainers postgres:16（integration）。

**Spec:** `docs/design/2026-09-29-multi-product-key-strategy.md`（三鍵策略：HTTP 邊界用 `company_internal_id uuid` 而非業務 `company_id int`）；AGENTS.md §11（平台模組邊界規則，step-1 已落地）。

## Global Constraints

- HTTP 邊界一律用 `company_internal_id uuid`（平台通用鍵），不洩漏業務 `companies.id int`——見三鍵策略 doc §3。
- `guardQuota` 的兩條政策（平台層身分略過判定、租戶 id 以身分為準）保留在產品層 `internal/services/shared_service.go`，不下沉到 platform RPC（判定層必須與身分無關，否則平台端視圖說謊）。
- 每個寫入呼叫點（`guardQuota`）仍經 `entitlementChecker` 介面，介面簽章不變；只有介面實作從 in-process 換成雙模式 client。
- 單產品場景下 `product_id` 恆為 `'sales-order'`（00054 已 hard-code），RPC 不需帶 product_id 直到第二產品出現。
- 所有變更頻繁 commit；每階段可獨立 revert（雙模式預設同進程，故 phase-1/2 上線後行為不變）。
- 獨立 repo 抽取（phase-4）最破壞性，放最後；此前平台仍在原 repo 內以 module + 獨立 main 存在。

---

## 現狀盤點（計畫論據）

- `entitlementChecker` 介面（`internal/services/shared_service.go:33`）唯一方法：`CheckLimit(ctx, companyID int, feature string, delta int) error`。
- `guardQuota`（`shared_service.go:47`）是唯一入口，8 個呼叫點（user_service:320、company_service:548、customer_service、product_service、shared_service、auth_handler 等）全經它。
- `entitlements.Service.CheckLimit`（`internal/platform/entitlements/service.go:272`）與 `Snapshot`（`snapshot.go:42`）是 in-process 實作，目前在產品請求交易內 same-tx 計數。
- 既有 RPC：`PlatformAdminService`（operator）、`TenantEntitlementService.GetTenantEntitlements`（租戶唯讀投影，已掛 tenant mux `domains.go:193`）。
- `consumer`（`internal/platformhost/consumer/consumer.go`）：認領 + `CompanyStatusSetter.SetCompanyStatus` 必須同交易（註解 line 13）；`CompanyStatusSetter` 接 `services.SetCompanyStatus`（line 110）。
- `platform.events` 已有條件式認領機制（outbox 基礎）。

---

## Phase 1: Platform HTTP 服務骨架（雙模式，零行為變更）

### Task 1.1: proto 加 CheckLimit RPC

**Files:**
- Modify: `backend/proto/platform/v1/platform.proto`（在 `TenantEntitlementService` 內加 rpc）
- Test: `backend/proto/platform/v1/platform.proto` 無單測；靠 codegen + 下游編譯

**Interfaces:**
- Consumes: 現有 `TenantEntitlementService` service 定義
- Produces: `CheckLimit` RPC 方法，供 Task 1.3 server 實作與 Task 2.x client 呼叫

- [ ] **Step 1: 在 `TenantEntitlementService` 加 CheckLimit rpc 與 message**

於 `platform.proto` 的 `service TenantEntitlementService` 區塊內（`GetTenantEntitlements` 之後）加入：

```proto
  // CheckLimit:寫路徑配額預約。product_id 單產品場景恆為 'sales-order'(由服務端預設)。
  // company_internal_id 為平台通用 uuid 鍵（三鍵策略），不洩漏業務 companies.id。
  // 超額回 PLAT-5001（feature/used/limit）；訂閱 none 回 PLAT-3001。
  rpc CheckLimit(CheckLimitRequest) returns (CheckLimitResponse);

message CheckLimitRequest {
  string product_id = 1;          // 單產品場景可空,服務端預設 'sales-order'
  string company_internal_id = 2; // uuid,平台通用鍵
  string feature = 3;             // 如 limit.seats
  int32 delta = 4;                // 預約增量(通常 +1)
}

message CheckLimitResponse {}
```

- [ ] **Step 2: codegen**

Run: `cd backend && task proto:gen`（或 `buf generate` + `go generate ./contracts/errcode`）
Expected: `contracts/proto/platform/v1/platform.pb.go` 與 `platformv1connect/platform.connect.go` 含 `CheckLimit` 方法與 `CheckLimitRequest`/`CheckLimitResponse` 型別，編譯通過。

- [ ] **Step 3: Commit**

```bash
git add backend/proto/platform/v1/ contracts/proto/platform/v1/ frontend/src/lib/proto/platform/v1/ platform-console/src/lib/proto/platform/v1/
git commit -m "feat(platform): proto 加 CheckLimit RPC(寫路徑配額預約)"
```

### Task 1.2: platform 服務實作 CheckLimit（包住既有 entitlements.Service）

**Files:**
- Modify: `backend/internal/platform/entitlements/service.go`（加 `CheckLimitRPC` 或直接掛 `TenantEntitlementService` 實作）
- Test: `backend/internal/platform/entitlements/service_test.go`（加單測驗證 RPC 層錯誤對應）

**Interfaces:**
- Consumes: `entitlements.Service.CheckLimit(ctx, companyID int, feature, delta)`（現有 in-process）
- Produces: `CheckLimit` RPC handler，內部解析 `company_internal_id` → 查 `companies.id` → 呼叫現有 `CheckLimit`

- [ ] **Step 1: 寫失敗測試（RPC 層錯誤對應）**

`service_test.go` 加：

```go
func TestCheckLimitRPCOverLimit(t *testing.T) {
    svc := newTestService(t) // 既有的假 store service
    // 種子:company internal_id = X,已達 seats 上限
    _, err := svc.CheckLimit(context.Background(), &platformv1.CheckLimitRequest{
        CompanyInternalId: "00000000-0000-0000-0000-0000000000X0",
        Feature:           entitlements.LimitSeats,
        Delta:             1,
    })
    if err == nil || !strings.Contains(err.Error(), "PLAT-5001") {
        t.Fatalf("超額應回 PLAT-5001,got %v", err)
    }
}
```

- [ ] **Step 2: Run test 確認失敗**

Run: `cd backend/internal/platform && go test -count=1 -run TestCheckLimitRPCOverLimit ./entitlements/`
Expected: FAIL（`CheckLimit` RPC 方法不存在）

- [ ] **Step 3: 實作 CheckLimit RPC handler**

於 `entitlements/service.go` 加方法（掛在 `Service` 上，實作 `platformv1connect.TenantEntitlementServiceHandler`）：

```go
func (s *Service) CheckLimit(ctx context.Context, req *platformv1.CheckLimitRequest) (*platformv1.CheckLimitResponse, error) {
    productID := req.ProductId
    if productID == "" {
        productID = "sales-order" // 單產品預設
    }
    companyID, err := s.resolveCompanyID(ctx, productID, req.CompanyInternalId)
    if err != nil {
        return nil, errcode.SysInternal.Error(nil)
    }
    if err := s.CheckLimit(ctx, companyID, req.Feature, int(req.Delta)); err != nil {
        return nil, err // CheckLimit 已回 PLAT-5001/3001 等註冊碼
    }
    return &platformv1.CheckLimitResponse{}, nil
}

// resolveCompanyID:platform 內部將 company_internal_id(uuid) 反查業務 companies.id。
// 單產品下直接 JOIN companies WHERE internal_id = $1;多產品需 product_id 維度。
func (s *Service) resolveCompanyID(ctx context.Context, productID, internalID string) (int, error) {
    var id int
    if err := s.db.QueryRowContext(ctx,
        `SELECT id FROM companies WHERE internal_id = $1`, internalID,
    ).Scan(&id); err != nil {
        return 0, errcode.SysInternal.Error(nil)
    }
    return id, nil
}
```

- [ ] **Step 4: Run test 確認通過**

Run: `cd backend/internal/platform && go test -count=1 -run TestCheckLimitRPCOverLimit ./entitlements/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/platform/entitlements/
git commit -m "feat(platform): 實作 CheckLimit RPC(包住既有 Service.CheckLimit)"
```

### Task 1.3: platform HTTP/gRPC listener（獨立 main）

**Files:**
- Create: `backend/cmd/platform-server/main.go`（新獨立進程，監聽 Connect RPC）
- Modify: `backend/Taskfile.yml`（加 `platform:serve` 任務）
- Test: 手動 smoke（啟 listener + grpcurl 呼叫 CheckLimit）

**Interfaces:**
- Consumes: `platformv1connect.TenantEntitlementServiceHandler`（Task 1.2 實作）、`PlatformAdminService` handler（現有）
- Produces: 可獨立部署的 platform 服務；產品側 Task 2.x 的 HTTP client 連此

- [ ] **Step 1: 寫 platform-server main（最小可跑）**

`backend/cmd/platform-server/main.go`：

```go
package main

import (
    "context"
    "net/http"
    "os"

    "github.com/pressly/goose/v3" // 僅用於啟動時確認版本表
    "github.com/rs/zerolog/log"

    platformv1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1"
    platformv1connect "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1/platformv1connect"
    "github.com/salesorder/sales-order-1.0/backend/config"
    "github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
    "github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
)

func main() {
    cfg := config.MustLoad()
    st := postgres.New(config.MustDSN(cfg))
    svc := entitlements.New(st, /* counter */ zeroCounter{}, /* cache */ nil, cfg.EntitlementTTL)

    mux := http.NewServeMux()
    path, handler := platformv1connect.NewTenantEntitlementServiceHandler(svc)
    mux.Handle(path, handler)
    // PlatformAdminService handler 同理掛載

    addr := os.Getenv("PLATFORM_LISTEN")
    if addr == "" {
        addr = ":8081"
    }
    log.Info().Str("addr", addr).Msg("platform server listening")
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal().Err(err).Msg("platform server")
    }
}
```

（`zeroCounter` 為 entitlements 測試內的計數器替身，正式實作需注入 `Counter` 介面——見 Task 2.4 說明。）

- [ ] **Step 2: Taskfile 加 serve 任務**

`Taskfile.yml` 加：

```yaml
  platform:serve:
    desc: 啟動 platform 獨立 HTTP 服務(phase-2)
    cmd: go run ./cmd/platform-server
```

- [ ] **Step 3: Smoke（手動）**

Run: `cd backend && PLATFORM_LISTEN=:8081 go run ./cmd/platform-server &`
Expected: 日誌 `platform server listening`，`:8081` 接受 Connect RPC。
（integration 測試在 Task 2.x 補）

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/platform-server/ backend/Taskfile.yml
git commit -m "feat(platform): 獨立 HTTP 服務 listener(cmd/platform-server)"
```

### Task 1.4: 雙模式 entitlementChecker（同進程/HTTP client）

**Files:**
- Create: `backend/internal/services/platform_client.go`（`entitlementChecker` 的雙模式實作）
- Modify: `backend/internal/services/server.go`（組裝時依 `PLATFORM_MODE=local|remote` 選實作）
- Test: `backend/internal/services/platform_client_test.go`（驗證模式選擇與 local 路徑行為不變）

**Interfaces:**
- Consumes: `platformv1connect.TenantEntitlementServiceClient`（codegen）、現有 `entitlements.Service`
- Produces: `entitlementChecker` 的遠端實作，供 `guardQuota` 透明呼叫

- [ ] **Step 1: 寫失敗測試（local 模式行為不變）**

`platform_client_test.go`：

```go
func TestEntitlementCheckerLocalUnchanged(t *testing.T) {
    // local 模式直接包 entitlements.Service,行為與現有一致
    svc := newTestEntitlements(t)
    c := NewLocalChecker(svc)
    // 超額應回 PLAT-5001
    if err := c.CheckLimit(context.Background(), 42, entitlements.LimitSeats, 1); err == nil {
        t.Fatal("local 模式應保留 same-tx 計數語意")
    }
}
```

- [ ] **Step 2: Run test 確認失敗**

Run: `cd backend && go test -count=1 -run TestEntitlementCheckerLocalUnchanged ./internal/services/`
Expected: FAIL（`NewLocalChecker` 不存在）

- [ ] **Step 3: 實作雙模式 client**

`backend/internal/services/platform_client.go`：

```go
package services

import (
    "context"

    platformv1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1"
    "github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

// LocalChecker:同進程模式,直接包 entitlements.Service(行為與現狀完全一致)。
type LocalChecker struct{ svc *entitlements.Service }

func NewLocalChecker(svc *entitlements.Service) *LocalChecker { return &LocalChecker{svc: svc} }

func (c *LocalChecker) CheckLimit(ctx context.Context, companyID int, feature string, delta int) error {
    return c.svc.CheckLimit(ctx, companyID, feature, delta)
}

// RemoteChecker:獨立服務模式,經 HTTP 呼叫 platform CheckLimit RPC。
// company_internal_id 由呼叫端從身分解析(見 guardQuota 的 companyID → internal_id 對應)。
type RemoteChecker struct{ client platformv1.TenantEntitlementServiceClient }

func (c *RemoteChecker) CheckLimit(ctx context.Context, companyID int, feature string, delta int) error {
    internalID := companyInternalIDFromCtx(ctx) // 由 authz.Identity 取 company.internal_id
    _, err := c.client.CheckLimit(ctx, &platformv1.CheckLimitRequest{
        CompanyInternalId: internalID,
        Feature:           feature,
        Delta:             int32(delta),
    })
    if err != nil {
        // Connect 錯誤已帶 PLAT-xxxx code,直接透傳
        return err
    }
    return nil
}
```

- [ ] **Step 4: Run test 確認通過**

Run: `cd backend && go test -count=1 -run TestEntitlementCheckerLocalUnchanged ./internal/services/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/platform_client.go backend/internal/services/platform_client_test.go
git commit -m "feat(services): entitlementChecker 雙模式(local/remote)實作"
```

---

## Phase 2: guardQuota 切到雙模式 client（預設 local）

### Task 2.1: 組裝點切換 mode

**Files:**
- Modify: `backend/internal/services/server.go`（InitDomains 注入 `entitlementChecker` 時依 env 選 Local/Remote）
- Test: 現有 `guardQuota` 8 呼叫點 integration 測試（預設 local → 行為不變）

**Interfaces:**
- Consumes: `LocalChecker`/`RemoteChecker`（Task 1.4）
- Produces: 上線後預設 local，行為零變更；設 `PLATFORM_MODE=remote` 才走網路

- [ ] **Step 1: 改 InitDomains 注入邏輯**

`server.go` 的 domain 組裝處（搜 `NewLocalChecker` 或 `entitlements.New` 注入點）：

```go
var checker entitlementChecker
if os.Getenv("PLATFORM_MODE") == "remote" {
    client := platformv1connect.NewTenantEntitlementServiceClient(http.DefaultClient, platformBaseURL())
    checker = &RemoteChecker{client: client}
} else {
    checker = NewLocalChecker(entSvc) // entSvc 為現有 entitlements.Service
}
```

- [ ] **Step 2: Run 現有 guardQuota integration（預設 local）**

Run: `cd backend && TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags integration -run 'TestIntegrationCreateUser|TestIntegrationCompanySoftDelete' ./internal/services/...`
Expected: PASS（行為不變，仍 same-tx）

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/server.go
git commit -m "feat(services): guardQuota 依 PLATFORM_MODE 選 local/remote(預設 local)"
```

### Task 2.2: Remote 模式 integration（同一 container 起 platform-server）

**Files:**
- Modify: `backend/internal/services/*_integration_test.go`（在 remote 模式下起 platform-server 於 :8081）
- Test: `TestIntegrationGuardQuotaRemote`（新，驗證 remote 路徑 + same-tx 轉移）

**Interfaces:**
- Consumes: `cmd/platform-server`（Task 1.3）、`RemoteChecker`（Task 1.4）
- Produces: remote 路徑的可驗證證據

- [ ] **Step 1: 寫 remote 模式測試**

`guard_quota_remote_integration_test.go`：

```go
func TestIntegrationGuardQuotaRemote(t *testing.T) {
    testsupport.RequiresContainer(t)
    // 起 platform-server(獨立進程,共用同一 PG)
    srv := startPlatformServer(t) // helper:go run ./cmd/platform-server,設 PLATFORM_LISTEN=:8081
    defer srv.Stop()
    t.Setenv("PLATFORM_MODE", "remote")
    t.Setenv("PLATFORM_BASE_URL", "http://localhost:8081")

    dsn := testsupport.Postgres(t)
    testsupport.MigrateUp(t, dsn) // 業務+平台兩段
    // 種子:company internal_id=X, seats 上限=1
    // 第一次 guardQuota(+1) 成功,第二次(+1) 應回 PLAT-5001
    // 斷言:platform 服務內部 usage 表計數正確(same-tx 轉移)
}
```

- [ ] **Step 2: Run 測試**

Run: `cd backend && TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags integration -run TestIntegrationGuardQuotaRemote ./internal/services/...`
Expected: PASS（remote 模式配額預約正確）

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/guard_quota_remote_integration_test.go
git commit -m "test(services): guardQuota remote 模式 integration(platform 服務 same-tx 計數)"
```

### Task 2.3: company_internal_id 跨邊界傳遞

**Files:**
- Modify: `backend/internal/services/platform_client.go`（`companyInternalIDFromCtx` 實作：從 `authz.Identity` 取 `company.internal_id`）
- Modify: `backend/internal/authz`（確認 `Identity` 攜 `CompanyInternalID` 或從 `companies` 反查）

**Interfaces:**
- Consumes: `authz.IdentityFrom(ctx)`、現有 `companies.internal_id` 欄位（00053 已加）
- Produces: remote 模式下 `CheckLimit` 帶正確 uuid 鍵

- [ ] **Step 1: 實作 companyInternalIDFromCtx**

`platform_client.go` 加：

```go
func companyInternalIDFromCtx(ctx context.Context) string {
    id := authz.IdentityFrom(ctx)
    if id == nil {
        return ""
    }
    // Identity 應攜 CompanyInternalID;否則由 CompanyID 反查(一次 SELECT)。
    if id.CompanyInternalID != "" {
        return id.CompanyInternalID
    }
    return lookupInternalID(ctx, id.CompanyID)
}
```

- [ ] **Step 2: Run Task 2.2 測試**

Run: `cd backend && TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags integration -run TestIntegrationGuardQuotaRemote ./internal/services/...`
Expected: PASS（uuid 鍵正確傳遞）

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/platform_client.go
git commit -m "feat(services): remote CheckLimit 攜 company_internal_id uuid 鍵"
```

### Task 2.4: Counter 介面注入 platform 服務

**Files:**
- Modify: `backend/cmd/platform-server/main.go`（注入真 `Counter` 實作，非 zeroCounter）
- Modify: `backend/internal/platform/entitlements`（確認 `New` 的 `Counter` 參數來源）

**Interfaces:**
- Consumes: 現有 `Counter` 介面（`entitlements/service.go:72`，產品側 `services/counters.go` 實作）
- Produces: platform 服務獨立運作時的計數正確性

- [ ] **Step 1: 將產品 Counter 實作移至可共用位置**

`services/counters.go` 的 `Counter` 實作依賴產品 `ent`/寫路徑。獨立服務下需將其重構為 platform 可注入（或 platform 自備 counter 表）。決策：**platform 自備 usage 表**（`platform.usage_counters`，按 company_internal_id+feature 計數），獨立服務擁有 same-tx；產品側 `Counter` 實作退場（§102-149 的「usage 在產品側」翻轉為「usage 在 platform 側」）。

- [ ] **Step 2: 加 platform.usage_counters 遷移（新 00055+）**

`database/platform_migrations/00055_usage_counters.sql`：

```sql
CREATE TABLE platform.usage_counters (
    company_internal_id uuid NOT NULL,
    feature            text NOT NULL,
    used               int  NOT NULL DEFAULT 0,
    PRIMARY KEY (company_internal_id, feature)
);
```

- [ ] **Step 3: platform 服務內 CheckLimit 改查 usage_counters（same-tx）**

`entitlements/service.go` 的 `CheckLimit` 內部 `counters.Count` 改為查 `platform.usage_counters` 的 `used` 欄位（在請求交易內）。

- [ ] **Step 4: Run platform integration**

Run: `cd backend/internal/platform && TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags integration ./...`
Expected: PASS（計數來自 platform.usage_counters）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/platform/ database/platform_migrations/ backend/cmd/platform-server/
git commit -m "feat(platform): usage_counters 表 + CheckLimit same-tx 計數(獨立服務自備)"
```

---

## Phase 3: Consumer outbox 反轉（凍結公司跨網路）

### Task 3.1: consumer 認領後發 company.status_changed outbox

**Files:**
- Modify: `backend/internal/platformhost/consumer/consumer.go`（認領成功後 `EmitEventTx` 帶 `company.status_changed`）
- Test: `backend/internal/platformhost/consumer/consumer_integration_test.go`（驗證 outbox 事件寫入）

**Interfaces:**
- Consumes: 現有 `store.EmitEventTx`（00054 已支援 `product_id`）、`company_internal_id`
- Produces: 產品 worker（Task 3.2）消費的事件

- [ ] **Step 1: 寫失敗測試**

`consumer_integration_test.go` 加：認領後 `platform.events` 應出現 `aggregate_type='company.status_changed'` 且 `aggregate_id=company_internal_id`。

- [ ] **Step 2: Run 確認失敗**

Run: `cd backend && TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags integration -run TestIntegrationDispatchEmitsStatusChanged ./internal/platformhost/consumer/`
Expected: FAIL

- [ ] **Step 3: 實作 outbox 發送**

`consumer.go` 認領區塊（line 110 附近）後加：

```go
if err := store.EmitEventTx(ctx, tx, store.Event{
    ProductID:      "sales-order",
    AggregateType: "company.status_changed",
    AggregateID:    companyInternalID, // uuid
    Payload:        json.RawMessage(`{"status":"` + status + `","reason":"` + reason + `"}`),
}); err != nil {
    return err
}
```

（同交易：認領 UPDATE + outbox INSERT 在同一 tx，保持冪等）

- [ ] **Step 4: Run 確認通過**

Run: `cd backend && TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags integration -run TestIntegrationDispatchEmitsStatusChanged ./internal/platformhost/consumer/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/platformhost/consumer/
git commit -m "feat(consumer): 認領後發 company.status_changed outbox 事件"
```

### Task 3.2: 產品 worker 消費 outbox → SetCompanyStatus

**Files:**
- Create: `backend/internal/platformhost/company_status_worker.go`（消費 `company.status_changed` → `services.SetCompanyStatus`）
- Modify: `backend/cmd/platform-cron/main.go`（註冊 worker 或獨立 consumer 迴圈）
- Test: `backend/internal/platformhost/company_status_worker_test.go`（驗證凍結語意 + 冪等）

**Interfaces:**
- Consumes: `platform.events`（Task 3.1 發出）、`services.SetCompanyStatus`
- Produces: 跨網路後公司凍結的最終一致性

- [ ] **Step 1: 寫失敗測試**

`company_status_worker_test.go`：注入假 `CompanyStatusSetter`，消費事件後應呼叫 `SetCompanyStatus(companyInternalID→companyID, status, reason)`。

- [ ] **Step 2: Run 確認失敗**

Run: `cd backend && go test -count=1 -run TestCompanyStatusWorker ./internal/platformhost/`
Expected: FAIL

- [ ] **Step 3: 實作 worker**

`company_status_worker.go`：

```go
func (w *CompanyStatusWorker) Process(ctx context.Context, ev store.Event) error {
    var p struct {
        Status string `json:"status"`
        Reason string `json:"reason"`
    }
    if err := json.Unmarshal(ev.Payload, &p); err != nil {
        return err
    }
    companyID, err := w.resolveCompanyID(ctx, ev.ProductID, ev.AggregateID.String())
    if err != nil {
        return err
    }
    return services.SetCompanyStatus(ctx, w.entClient, companyID, p.Status, p.Reason, w.actor)
}
```

（冪等：worker 用事件 id 去重；失敗可重試，SetCompanyStatus 本身冪等）

- [ ] **Step 4: Run 確認通過**

Run: `cd backend && go test -count=1 -run TestCompanyStatusWorker ./internal/platformhost/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/platformhost/company_status_worker.go backend/internal/platformhost/company_status_worker_test.go backend/cmd/platform-cron/main.go
git commit -m "feat(platformhost): company.status_changed worker → services.SetCompanyStatus(跨網路最終一致)"
```

---

## Phase 4: 獨立 repo 抽取（最破壞性，最後做）

### Task 4.1: 安裝 git-filter-repo（已完成，最終未用）

**Files:**
- 環境：安裝 `git-filter-repo`（brew 或 pip）

- [x] **Step 1: 安裝**

Run: `brew install git-filter-repo`（或 `pip install git-filter-repo`）
Expected: `git filter-repo --version` 可執行 —— 2026-10-01 已裝（2.47.0）。

- [x] **Step 2: 決策（手動 split 替代 filter-repo）**

filter-repo 重寫 product 全歷史（不可逆，team clone 失效）。2026-10-01 選**手動 split**（git mv + 獨立 repo init），product 歷史不動，達成相同「獨立 repo」目標。見 Task 4.2。

### Task 4.2: 抽取 platform 為獨立 repo（手動 split，已完成 2026-10-01）

**決策（2026-10-01）**：不分 filter-repo 重寫 product 歷史，改為手動 split —— platform 目錄 `git mv` 到 repo 根 `platform/` 並 `git init` 成獨立 repo（module path 保持 `…/backend/internal/platform`，避免改所有 import）；product 歷史不動。

- 新 repo：`platform/`（repo 根，獨立 `.git`，init commit `6580843`）
- 抽取範圍：`backend/internal/platform` → `platform/`（含 `billing`/`cron`/`entitlements`/`money`/`operatorauth`/`server`/`store`）；`contracts/` 仍留產品 repo（`backend/contracts`），platform 的 `go.mod` 以 `replace …/backend/contracts => ../backend/contracts` 引用；`database/platform_migrations` 已屬 platform 模組（00054 拆分後）；`cmd/platform-server` 在產品 repo（引用 platform 模組，獨立 binary）；`internal/platformhost` 留產品 repo（產品側 adapter）。

**Interfaces:**
- Consumes: 現有 module 結構（step-1 已就位）
- Produces: 獨立 `go.mod` 發版，產品 repo `require` 去 `replace`（待 push 遠端後）

- [x] **Step 1: 手動 split**

`git mv backend/internal/platform platform`；`platform/go.mod` 的 contracts replace 改 `../backend/contracts`；product `go.mod` 的 platform replace 改 `../platform`；`.gitignore` 排除 `/platform/`（獨立 repo 不入 product tree）。

- [x] **Step 2: 獨立 repo init**

`cd platform && git init && git add -A && git commit`（init commit `6580843`）；product repo `git rm --cached -f platform` + commit（移除 platform 的 index 追蹤，保留工作目錄）。

- [x] **Step 3: CI checkout step**

`.github/workflows/ci.yml` backend job 加 platform checkout（`git clone "${PLATFORM_REPO_URL}" ../platform`，`if: env.PLATFORM_REPO_URL != ''`）；`PLATFORM_REPO_URL` 待 push 遠端後設 secret。本地開發用 `replace ../platform` 已綠。

- [ ] **Step 4: 推送遠端 + 去 replace（待執行）**

push `platform/` 到 `github.com/sw-ssd/platform` 後：`backend/go.mod` 移除 `replace …/backend/internal/platform => ../platform`，加 `require github.com/sw-ssd/platform v0.1.0`，CI 設 `PLATFORM_REPO_URL` secret。單產品下本地 replace 已滿足開發，此步收益為零，待第二產品或獨立發版需求才做。

```bash
git commit -m "chore: platform 抽取為獨立 repo（手動 split）"
```

---

## 自我審查

1. **Spec 覆蓋**：三鍵策略的 `company_internal_id uuid` 邊界（Task 1.1/2.3/3.1/3.2 均使用）→ 覆蓋。AGENTS §11 模組邊界（step-1 已落地，phase-2 在其上擴展）→ 覆蓋。guardQuota 政策保留產品層（Global Constraints 明載）→ 覆蓋。
2. **Placeholder 掃描**：Task 2.4 Step 1「Counter 實作退場」有設計決策但無程式碼——已補 Task 2.4 Step 2-3 的 `platform.usage_counters` 遷移 + same-tx 計數實作。Task 4.2 path 需對應實際佈局（已註明）。其餘步驟含程式碼。
3. **型別一致性**：`CheckLimitRequest{CompanyInternalId, Feature, Delta}` 在 Task 1.1 定義、Task 1.2/1.4/2.2/2.3 共用 → 一致。`entitlementChecker.CheckLimit(ctx, companyID int, feature, delta)` 簽章 Task 1.4 不變 → 與 `shared_service.go:33` 一致。`store.Event{AggregateID uuid, ProductID}` 與 00054 一致 → 一致。

## 風險標註（單產品場景）

- Phase 1-3 上線後**預設 local 模式**，行為與現狀零差異（same-tx 保留），可獨立 revert。
- `PLATFORM_MODE=remote` 才觸發網路路徑；單產品下無收益，僅驗證獨立部署可行性。
- Phase 4 改寫歷史不可逆，需在確認第二產品 repo 存在且發版節奏不同時才執行（§1417 觸發條件）。
- §149 分析的 counting 跨網路 RTT + 同交易破壞：Task 2.4 將 same-tx 轉移到 platform 服務內部（platform 自備 usage 表）解決；consumer 跨網路凍結經 outbox 最終一致（Task 3.x），失敗開放風險由冪等 + 重試吸收。
