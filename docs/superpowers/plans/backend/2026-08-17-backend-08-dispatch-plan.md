# Backend 08 — 派車與串流看板 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 backend 派車域全部後端功能 — AssignRoute 看板拖放(樂觀鎖 + 順位重排)、車次批次 Confirm(逐筆交易、部分失敗語義)、CancelDispatch(dept_admin + 原因 + 重印警告)、派車通知觸發(介面注入)、WatchBoard server streaming、Valkey pub/sub 跨 replica 轉發與 25 秒 heartbeat。

**Architecture:** 依 `docs/superpowers/plans/backend/detail/08-dispatch.md`(下稱「細部文件」,子功能編號 5.x.y)實作。`DispatchService` 四個 RPC 落於 `backend/internal/domain/salesorders`(與 05-plan 同套件,共用訂單狀態機);三個 mutation 的 DB 交易**提交成功後**才發佈 `BoardEvent`(D14);事件經 Valkey 以部門分 channel 廣播,各 replica 僅訂閱本機有連線的部門;heartbeat 純連線維持,不寫事件、不進 Valkey。認證走 01-plan 既有 auth interceptor,不實作一次性 ticket(D14)。

**Tech Stack:** Go 1.25、Ent(entgo.io)、Chi v5、Connect-RPC、pgx/v5、redis/go-redis/v9(Valkey)、prometheus/client_golang(看板指標,D19)、testcontainers-go(整合測試)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend/detail/08-dispatch.md`;共通規則見 `docs/superpowers/plans/backend/detail/00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- proto 產生碼:Phase 0 `buf.gen.yaml` 輸出至 `backend/gen/`;Go import 為 `protov1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"`(訊息)與 `dispatchv1connect "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1/v1connect"`(Connect 產生碼);`dispatch.proto` 須帶 `option go_package = "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1;protov1";`。每次改 proto 後跑 `cd backend && buf generate`。
- 錯誤:RPC 層統一 Connect code — `unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument` / `already_exists`(00-index §3.4)。
- 樂觀鎖:看板拖放/指派異動比對 `sales_orders.version`,衝突一律回 `failed_precondition`,後端不自動重試業務語意衝突(D14)。
- 交易與稽核(D18):狀態異動 + `sales_order_events`、取消派車 + `audit_logs` 同一 DB 交易;`BoardEvent` 於交易**提交成功後**才發佈,rollback 不產生任何事件。
- 事件語意(D14):`BoardEvent` 僅作「看板查詢失效」提示;不保證順序、不補發斷線期間事件;pub/sub 為 at-most-once,遺失由前端重連後全量重查吸收。
- 軟刪除(D10):所有查詢預設排除 `deleted_at IS NOT NULL`。
- 多租戶(D3):派車操作僅能作用於操作者 `data_scope` 範圍內部門;repository 查詢條件 + RLS(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`,由 01-plan Task 3 提供)為最後防線;看板事件以 `department_id` 分 channel,跨部門事件不得外洩。
- 認證:WatchBoard 與其他 RPC 相同,Web 走 httpOnly cookie、App 走 Bearer JWT,由 01-plan Task 11 auth middleware 注入 `rls.Identity`;**不實作一次性 ticket、不接受 URL 參數憑證**(D14)。
- Casbin 詞彙:沿用 01-plan Task 2 的 act 集合(`read|create|update|delete`);派車三 RPC 以 `obj=dispatch`、`act=update` 判定功能權限,WatchBoard 以 `obj=dispatch`、`act=read` 判定;CancelDispatch **額外**要求角色 dept_admin 以上(以 `rls.Identity.Role` 判定,細部 5.1.3 步驟 1,不新增 Casbin act 值)。Casbin 預設 policy 須含 `dispatch` 資源 — 由 01-plan Task 2 seeder 提供;若 seeder 缺 `dispatch` 規則,於該 seeder 補齊並標 `TODO(接手 Task:01-plan Task 2)`。
- `sales_order_events.event_type` 列舉含 `route_assign`(05-plan Task 1 schema 已定案八值,細部 5.1.1)。
- 狀態流轉:`pending → processing`(派車確認)與 `processing → pending`(取消派車)一律經 `CanTransition` / `Transition`(05-plan 4.1.3 提供,同套件 `salesorders`),禁止自行改 `status` 欄位;`dispatched_at` / `dispatched_by` 與 `version + 1` 由本計畫在**同一 tx** 內以條件更新完成(細部 5.1.2 步驟 3);version 遞增由觸發方做一次(同 05 細部 4.1.4 Update 由呼叫方 +1 之例)。
- 通知:派車通知以 `DispatchNotifier` 介面注入,實作由 07-plan Task 6 `triggers.NewDispatchNotifier` 提供(範本 code=`dispatch`,通道 `fcm` / `in_app`,變數 = 訂單編號/車次名稱/預計出貨日);發送失敗僅落 `notifications.status=failed`(07 既有機制),不向 Confirm 呼叫者回錯、不影響批次結果(D16)。
- heartbeat:間隔預設 25 秒(低於常見 30 秒 ingress idle timeout),可組態;heartbeat 不寫 `sales_order_events`、不發佈至 Valkey、不進跨 replica 轉發。
- 前端職責註記(後端不實作,屬原計畫 Task 5.6):前端收到任何事件(含 heartbeat)僅 invalidate 全量重查;斷線 exponential backoff 重連;連續失敗降級 30 秒輪詢。後端配合點:看板查詢 RPC(05-plan `SalesOrderService.List` 的看板視圖)與串流並存可用、heartbeat 間隔可組態。
- 測試:DB 相依測試走 testcontainers Postgres 16,共用 `testutil.NewEntClient`(01-plan Task 1 提供);Valkey 相依測試走 testcontainers `valkey/valkey:8-alpine`,`testutil.NewValkeyClient` 由本計畫 Task 5 提供;`go test ./...` 必須全綠。
- 每個 Task 結尾 commit;commit message 格式 `feat(backend): …` / `test(backend): …`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/proto/v1/dispatch.proto` | DispatchService 四 RPC 與 BoardEvent | Task 1 起隨需求增量 |
| `backend/internal/domain/salesorders/dispatch.go` | DispatchUsecase(AssignRoute / Confirm / CancelDispatch)、BoardEvent 詞彙、DispatchNotifier 介面 | Task 1-3 |
| `backend/internal/domain/salesorders/dispatch_handler.go` | Connect handler(proto ↔ usecase 轉換) | Task 1-4 |
| `backend/internal/domain/salesorders/dispatch_test.go` | AssignRoute 測試 | Task 1 |
| `backend/internal/domain/salesorders/dispatch_confirm_test.go` | Confirm 批次測試 | Task 2 |
| `backend/internal/domain/salesorders/dispatch_cancel_test.go` | CancelDispatch + 通知觸發測試 | Task 3 |
| `backend/internal/domain/salesorders/watch_board.go` | BoardHub 註冊表、Valkey relay、metrics | Task 4-5 |
| `backend/internal/domain/salesorders/watch_board_test.go` | 串流 handler 測試(in-process Connect client) | Task 4 |
| `backend/internal/testutil/valkey.go` | testcontainers Valkey 測試輔助 | Task 5 |
| `backend/internal/domain/salesorders/watch_board_valkey_test.go` | 跨 replica / rollback / heartbeat 測試 | Task 5 |
| `backend/config/api.go` | heartbeat 與 ingress timeout 組態 | Task 5 |

---

### Task 1: AssignRoute 樂觀鎖與順位重排(細部 5.1.1)

**Files:**
- Create: `backend/proto/v1/dispatch.proto`
- Create: `backend/internal/domain/salesorders/dispatch.go`
- Create: `backend/internal/domain/salesorders/dispatch_handler.go`
- Test: `backend/internal/domain/salesorders/dispatch_test.go`

**Interfaces:**
- Consumes: `rls.Identity{UserID, CompanyID uuid.UUID; DepartmentID *uuid.UUID; DataScope, Role string}` 與 `rls.FromContext(ctx)`(01-plan Task 3/11/13 提供);`ent.SalesOrder` / `ent.SalesOrderEvent` / `ent.Route` 產生碼(05-plan 4.1.1、04-plan 3.4.2 提供);`testutil.NewEntClient`(01-plan Task 1 提供)。
- Produces: proto `DispatchService.AssignRoute`;`salesorders.DispatchUsecase`(`NewDispatchUsecase(db *ent.Client, pub BoardPublisher, notifier DispatchNotifier) *DispatchUsecase`、`AssignRoute(ctx, actor rls.Identity, in AssignRouteInput) (*AssignRouteResult, error)`);看板事件詞彙 `BoardEvent` / `BoardEventKind`(`EventRouteAssign` / `EventDispatch` / `EventDispatchCancel` / `EventHeartbeat`)/ `BoardPublisher` 介面(Task 4/5 實作);`DispatchNotifier` 介面 + `noopDispatchNotifier`(Task 3 接上觸發點,實作由 07-plan Task 6 提供)。

- [ ] **Step 1: dispatch.proto 建立 AssignRoute 並產碼**

`backend/proto/v1/dispatch.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

option go_package = "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1;protov1";

// DispatchService 為派車看板後端(細部文件 08-dispatch)。
service DispatchService {
  // AssignRoute 指派車次與配送順序(拖放提交;route_id 省略 = 拖回未指派)。
  rpc AssignRoute(AssignRouteRequest) returns (AssignRouteResponse);
}

message AssignRouteRequest {
  string sales_order_id = 1;
  optional string route_id = 2;   // 未帶 = 拖回「未指派」
  int32 delivery_sequence = 3;    // route_id 有值時必須 >= 1
  int64 version = 4;              // 客戶端讀取時的 sales_orders.version,必填
  string expected_delivery_date = 5; // YYYY-MM-DD,看板當前日期,用於順位重排範圍
}
message AssignRouteResponse {
  string sales_order_id = 1;
  optional string route_id = 2;
  int32 delivery_sequence = 3;
  int64 version = 4;              // 更新後的新 version
}
```

Run: `cd backend && buf generate`
Expected: `backend/gen/proto/v1/dispatch.pb.go` 與 `backend/gen/proto/v1/v1connect/dispatch.connect.go` 產生。

- [ ] **Step 2: 寫失敗測試(樂觀鎖、重排、拖回未指派、跨部門車次、併發)**

`backend/internal/domain/salesorders/dispatch_test.go`:

```go
package salesorders_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entsalesorder "github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	entsalesorderevent "github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

var boardDate = time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)

// fakePublisher 記錄發佈的看板事件。
type fakePublisher struct {
	mu     sync.Mutex
	events []salesorders.BoardEvent
}

func (f *fakePublisher) Publish(_ context.Context, ev salesorders.BoardEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, ev)
}

func (f *fakePublisher) len() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func seedCompanyDept(t *testing.T, c *ent.Client) (*ent.Company, *ent.Department) {
	t.Helper()
	ctx := context.Background()
	co, err := c.Company.Create().SetName("甲公司").SetIdentifier("co-a").
		SetCustomerCodePrefix("AA").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	dept, err := c.Department.Create().SetCompanyID(co.ID).SetName("北區").Save(ctx)
	if err != nil {
		t.Fatalf("department: %v", err)
	}
	return co, dept
}

func seedRoute(t *testing.T, c *ent.Client, co *ent.Company, dept *ent.Department, code string) *ent.Route {
	t.Helper()
	r, err := c.Route.Create().SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetCode(code).SetName("車次" + code).SetIsActive(true).Save(context.Background())
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	return r
}

func seedOrder(t *testing.T, c *ent.Client, co *ent.Company, dept *ent.Department,
	orderNo, status string, route *ent.Route, seq int) *ent.SalesOrder {
	t.Helper()
	m := c.SalesOrder.Create().
		SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetOrderNo(orderNo).SetCustomerID(uuid.New()).
		SetSource("W").SetStatus(status).
		SetExpectedDeliveryDate(boardDate).
		SetSalesRepID(uuid.New()).SetCreatedBy(uuid.New()).
		SetVersion(1)
	if route != nil {
		m.SetRouteID(route.ID).SetDeliverySequence(seq)
	}
	od, err := m.Save(context.Background())
	if err != nil {
		t.Fatalf("order %s: %v", orderNo, err)
	}
	return od
}

func deptAdmin(co *ent.Company, dept *ent.Department) rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &dept.ID,
		DataScope: "department", Role: "dept_admin"}
}

func TestAssignRouteSuccessWithRebalance(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	rA := seedRoute(t, c, co, dept, "A")
	rB := seedRoute(t, c, co, dept, "B")
	// rB 已有 seq 1、2 兩單;把 od 插入 rB seq 2,原 seq 2 應後移為 3
	keep := seedOrder(t, c, co, dept, "W000001", "pending", rB, 1)
	shifted := seedOrder(t, c, co, dept, "W000002", "pending", rB, 2)
	od := seedOrder(t, c, co, dept, "W000003", "pending", rA, 1)

	pub := &fakePublisher{}
	uc := salesorders.NewDispatchUsecase(c, pub, nil)
	res, err := uc.AssignRoute(ctx, deptAdmin(co, dept), salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: &rB.ID, DeliverySequence: 2,
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	if res.Version != 2 || *res.RouteID != rB.ID || res.DeliverySequence != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}
	gotShifted, _ := c.SalesOrder.Get(ctx, shifted.ID)
	if *gotShifted.DeliverySequence != 3 {
		t.Fatalf("existing seq-2 order should shift to 3, got %d", *gotShifted.DeliverySequence)
	}
	gotKeep, _ := c.SalesOrder.Get(ctx, keep.ID)
	if *gotKeep.DeliverySequence != 1 {
		t.Fatalf("seq-1 order should stay, got %d", *gotKeep.DeliverySequence)
	}
	// route_assign 事件落庫(細部 5.1.1 步驟 4)
	exists, _ := c.SalesOrderEvent.Query().Where(
		entsalesorderevent.SalesOrderIDEQ(od.ID),
		entsalesorderevent.EventTypeEQ("route_assign"),
	).Exist(ctx)
	if !exists {
		t.Fatal("route_assign sales_order_event should exist")
	}
	// 提交成功後發佈 BoardEvent
	if pub.len() != 1 || pub.events[0].Kind != salesorders.EventRouteAssign ||
		pub.events[0].Version != 2 || pub.events[0].DepartmentID != dept.ID {
		t.Fatalf("board event mismatch: %+v", pub.events)
	}
}

func TestAssignRouteStaleVersionConflict(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", nil, 0)

	pub := &fakePublisher{}
	uc := salesorders.NewDispatchUsecase(c, pub, nil)
	_, err := uc.AssignRoute(ctx, deptAdmin(co, dept), salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: &r.ID, DeliverySequence: 1,
		Version: 99, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("stale version should be failed_precondition, got %v", err)
	}
	got, _ := c.SalesOrder.Get(ctx, od.ID)
	if got.RouteID != nil || got.Version != 1 {
		t.Fatal("conflict must not change any data")
	}
	if pub.len() != 0 {
		t.Fatal("conflict must not publish board event")
	}
}

func TestAssignRouteBackToUnassigned(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	od := seedOrder(t, c, co, dept, "W000002", "pending", r, 2)

	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)
	res, err := uc.AssignRoute(ctx, deptAdmin(co, dept), salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: nil, // 拖回未指派
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if err != nil {
		t.Fatalf("unassign: %v", err)
	}
	if res.RouteID != nil || res.DeliverySequence != 0 || res.Version != 2 {
		t.Fatalf("unexpected unassign result: %+v", res)
	}
	// 原車次順位空洞不回填(細部 5.1.1 步驟 4)
	other, _ := c.SalesOrder.Query().Where(entsalesorder.OrderNoEQ("W000001")).Only(ctx)
	if *other.DeliverySequence != 1 {
		t.Fatalf("remaining order seq should stay 1 (hole allowed), got %d", *other.DeliverySequence)
	}
}

func TestAssignRouteRejects(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	_, dept2 := seedCompanyDept(t, c)              // 同公司第二部門... 需同公司,改用同 co 建第二 dept:
	_ = dept2
	deptB, err := c.Department.Create().SetCompanyID(co.ID).SetName("南區").Save(ctx)
	if err != nil {
		t.Fatalf("deptB: %v", err)
	}
	rA := seedRoute(t, c, co, dept, "A")
	rB := seedRoute(t, c, co, deptB, "B") // 跨部門車次
	processing := seedOrder(t, c, co, dept, "W000001", "processing", rA, 1)
	pending := seedOrder(t, c, co, dept, "W000002", "pending", nil, 0)

	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)
	actor := deptAdmin(co, dept)

	// 非 pending → failed_precondition
	_, err = uc.AssignRoute(ctx, actor, salesorders.AssignRouteInput{
		SalesOrderID: processing.ID, RouteID: &rA.ID, DeliverySequence: 2,
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("processing order: want failed_precondition, got %v", err)
	}
	// 跨部門車次 → permission_denied
	_, err = uc.AssignRoute(ctx, actor, salesorders.AssignRouteInput{
		SalesOrderID: pending.ID, RouteID: &rB.ID, DeliverySequence: 1,
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("cross-department route: want permission_denied, got %v", err)
	}
	// 不存在車次 → not_found
	ghost := uuid.New()
	_, err = uc.AssignRoute(ctx, actor, salesorders.AssignRouteInput{
		SalesOrderID: pending.ID, RouteID: &ghost, DeliverySequence: 1,
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("ghost route: want not_found, got %v", err)
	}
	// seq < 1 / version 缺漏 → invalid_argument
	_, err = uc.AssignRoute(ctx, actor, salesorders.AssignRouteInput{
		SalesOrderID: pending.ID, RouteID: &rA.ID, DeliverySequence: 0,
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("seq 0: want invalid_argument, got %v", err)
	}
	_, err = uc.AssignRoute(ctx, actor, salesorders.AssignRouteInput{
		SalesOrderID: pending.ID, RouteID: &rA.ID, DeliverySequence: 1,
		Version: 0, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("version 0: want invalid_argument, got %v", err)
	}
}

func TestAssignRouteConcurrentOnlyOneWins(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", nil, 0)

	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)
	actor := deptAdmin(co, dept)
	in := salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: &r.ID, DeliverySequence: 1,
		Version: 1, ExpectedDeliveryDate: boardDate,
	}
	var wg sync.WaitGroup
	codes := make(chan connect.Code, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := uc.AssignRoute(ctx, actor, in)
			codes <- connect.CodeOf(err) // 成功時 CodeOf(nil) 為 CodeOK 語義外的零值,另行判斷
			_ = err
		}()
	}
	wg.Wait()
	close(codes)
	// 重新精確計算:成功數以 DB 最終 version 為準
	got, _ := c.SalesOrder.Get(ctx, od.ID)
	if got.Version != 2 {
		t.Fatalf("exactly one assign should win, version=%d", got.Version)
	}
}
```

(併發測試補強:兩 goroutine 的 error 一個為 nil、一個為 failed_precondition;上列 channel 收集後逐一檢查 `err == nil` 計數 — 實作時以 `errs := make(chan error, 2)` 收集並斷言恰好一個 nil,以此為準。)

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run TestAssignRoute -v`
Expected: FAIL — `salesorders.NewDispatchUsecase` / `AssignRouteInput` / `BoardEvent` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 DispatchUsecase.AssignRoute 與事件詞彙**

`backend/internal/domain/salesorders/dispatch.go`:

```go
package salesorders

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entprintlog "github.com/salesorder/sales-order-1.0/backend/ent/printlog"
	entroute "github.com/salesorder/sales-order-1.0/backend/ent/route"
	entsalesorder "github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// ---- 看板事件詞彙(細部 5.2.1/5.2.2;BoardHub 於 Task 4/5 實作 BoardPublisher)----

// BoardEventKind 為看板事件種類;heartbeat 僅連線維持,不進 Valkey。
type BoardEventKind string

const (
	EventRouteAssign    BoardEventKind = "route_assign"
	EventDispatch       BoardEventKind = "dispatch"
	EventDispatchCancel BoardEventKind = "dispatch_cancel"
	EventHeartbeat      BoardEventKind = "heartbeat"
)

// BoardEvent 僅作「看板查詢失效」提示(D14);不保證順序、不補發。
type BoardEvent struct {
	Kind             BoardEventKind `json:"kind"`
	SalesOrderID     uuid.UUID      `json:"sales_order_id"`
	RouteID          *uuid.UUID     `json:"route_id,omitempty"`
	DeliverySequence int32          `json:"delivery_sequence"`
	Version          int64          `json:"version"`
	DepartmentID     uuid.UUID      `json:"department_id"`
}

// BoardPublisher 由 BoardHub 實作(Task 4 本機轉發、Task 5 Valkey);mutation 於
// DB 交易提交成功後才呼叫,發佈失敗不影響 mutation 結果(細部 5.2.2 步驟 4)。
type BoardPublisher interface {
	Publish(ctx context.Context, ev BoardEvent)
}

// ---- 派車通知(細部 5.1.4;實作由 07-plan 提供)----

// DispatchNotification 為派車通知觸發載體;接收者解析、範本渲染(code=dispatch,
// 通道 fcm/in_app)與發送皆由 07-plan 的 adapter 完成。
type DispatchNotification struct {
	CompanyID            uuid.UUID
	DepartmentID         uuid.UUID
	SalesOrderID         uuid.UUID
	CustomerID           uuid.UUID
	OrderNo              string
	RouteName            string
	ExpectedDeliveryDate string // YYYY-MM-DD
}

// DispatchNotifier 為通知觸發介面;fire-and-record,錯誤不影響派車結果。
type DispatchNotifier interface {
	NotifyDispatched(ctx context.Context, n DispatchNotification) error
}

// TODO(接手:07-plan Task 6):以 triggers.NewDispatchNotifier 替換 noopDispatchNotifier
// (範本 code=dispatch,變數 order_no / route_name / expected_delivery_date)。
type noopDispatchNotifier struct{}

func (noopDispatchNotifier) NotifyDispatched(context.Context, DispatchNotification) error { return nil }

// ---- DispatchUsecase ----

// DispatchUsecase 承載派車看板三個 mutation;notifier 可為 nil(視同 noop)。
type DispatchUsecase struct {
	db       *ent.Client
	enf      *casbin.Enforcer // 可為 nil;非 nil 時以 obj=dispatch 判定功能權限
	pub      BoardPublisher
	notifier DispatchNotifier
}

func NewDispatchUsecase(db *ent.Client, pub BoardPublisher, notifier DispatchNotifier) *DispatchUsecase {
	if pub == nil {
		pub = noopBoardPublisher{}
	}
	if notifier == nil {
		notifier = noopDispatchNotifier{}
	}
	return &DispatchUsecase{db: db, pub: pub, notifier: notifier}
}

// SetEnforcer 接上 Casbin(01-plan Task 2);main.go 組裝時呼叫,測試多為 nil。
func (u *DispatchUsecase) SetEnforcer(enf *casbin.Enforcer) { u.enf = enf }

// noopBoardPublisher 為未接 BoardHub 時的佔位(main.go 組裝前)。
type noopBoardPublisher struct{}

func (noopBoardPublisher) Publish(context.Context, BoardEvent) {}

// AssignRouteInput 為拖放提交參數。
type AssignRouteInput struct {
	SalesOrderID         uuid.UUID
	RouteID              *uuid.UUID // nil = 拖回「未指派」
	DeliverySequence     int32
	Version              int64
	ExpectedDeliveryDate time.Time
}

type AssignRouteResult struct {
	SalesOrderID     uuid.UUID
	RouteID          *uuid.UUID
	DeliverySequence int32
	Version          int64
}

// authorizeDispatch 功能權限(Casbin obj=dispatch act=update)+ data_scope 部門檢查。
func (u *DispatchUsecase) authorizeDispatch(actor rls.Identity, deptID uuid.UUID) error {
	if u.enf != nil {
		ok, err := u.enf.Enforce(actor.Role, actor.CompanyID.String(), "dispatch", "update")
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		if !ok {
			return connect.NewError(connect.CodePermissionDenied, errors.New("no dispatch permission"))
		}
	}
	switch actor.DataScope {
	case "all", "company":
		return nil // 公司層範圍由 RLS 最後防線限縮
	case "department":
		if actor.DepartmentID == nil || *actor.DepartmentID != deptID {
			return connect.NewError(connect.CodePermissionDenied, errors.New("order out of data scope"))
		}
		return nil
	default:
		return connect.NewError(connect.CodePermissionDenied, errors.New("data scope not allowed"))
	}
}

// deptAdminOrAbove:CancelDispatch 角色門檻(細部 5.1.3 步驟 1)。
func deptAdminOrAbove(role string) bool {
	switch role {
	case "super", "company_admin", "dept_admin":
		return true
	}
	return false
}

// AssignRoute 指派車次與配送順序(細部 5.1.1)。
func (u *DispatchUsecase) AssignRoute(ctx context.Context, actor rls.Identity, in AssignRouteInput) (*AssignRouteResult, error) {
	// 步驟 1:輸入驗證
	if in.SalesOrderID == uuid.Nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("sales_order_id required"))
	}
	if in.RouteID != nil && in.DeliverySequence < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("delivery_sequence must be >= 1"))
	}
	if in.Version < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("version required"))
	}

	// 步驟 2:讀取訂單
	order, err := u.db.SalesOrder.Query().Where(
		entsalesorder.IDEQ(in.SalesOrderID),
		entsalesorder.DeletedAtIsNil(),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("sales order not found"))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := u.authorizeDispatch(actor, order.DepartmentID); err != nil {
		return nil, err
	}
	if order.Status != "pending" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("order not pending"))
	}

	// 步驟 3:車次驗證 — 不存在/已軟刪 → not_found;跨部門 → permission_denied
	if in.RouteID != nil {
		route, err := u.db.Route.Query().Where(
			entroute.IDEQ(*in.RouteID),
			entroute.DeletedAtIsNil(),
		).Only(ctx)
		if ent.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("route not found"))
		}
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if route.DepartmentID != order.DepartmentID {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cross-department route"))
		}
	}

	// 步驟 4:交易(邊界開始)
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	upd := tx.SalesOrder.Update().Where(
		entsalesorder.IDEQ(order.ID),
		entsalesorder.VersionEQ(int(in.Version)),
		entsalesorder.StatusEQ("pending"),
	).AddVersion(1)
	if in.RouteID != nil {
		upd.SetRouteID(*in.RouteID).SetDeliverySequence(int(in.DeliverySequence))
	} else {
		upd.ClearRouteID().ClearDeliverySequence()
	}
	n, err := upd.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if n == 0 { // 樂觀鎖衝突:整個交易 rollback(細部 5.1.1 步驟 4)
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("order changed by others"))
	}

	// 同車次順位重排:目標車次中 seq >= 新順位的其他訂單依序後移;
	// 原車次空位不回填(順位允許空洞,由看板排序吸收)。
	if in.RouteID != nil {
		_, err = tx.SalesOrder.Update().Where(
			entsalesorder.RouteIDEQ(*in.RouteID),
			entsalesorder.DepartmentIDEQ(order.DepartmentID),
			entsalesorder.ExpectedDeliveryDateEQ(in.ExpectedDeliveryDate),
			entsalesorder.StatusEQ("pending"),
			entsalesorder.DeliverySequenceGTE(int(in.DeliverySequence)),
			entsalesorder.IDNEQ(order.ID),
			entsalesorder.DeletedAtIsNil(),
		).AddDeliverySequence(1).Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// route_assign 事件同交易落庫
	if err := insertRouteAssignEvent(ctx, tx, order, actor.UserID, in); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// 步驟 5:提交成功後才發佈(細部 5.2.2 步驟 1)
	newVersion := in.Version + 1
	u.pub.Publish(ctx, BoardEvent{
		Kind:             EventRouteAssign,
		SalesOrderID:     order.ID,
		RouteID:          in.RouteID,
		DeliverySequence: in.DeliverySequence,
		Version:          newVersion,
		DepartmentID:     order.DepartmentID,
	})
	return &AssignRouteResult{
		SalesOrderID:     order.ID,
		RouteID:          in.RouteID,
		DeliverySequence: in.DeliverySequence,
		Version:          newVersion,
	}, nil
}

// insertRouteAssignEvent 寫入 route_assign 事件(僅追加表,05-plan 4.1.1)。
// event_type 列舉含 route_assign(05-plan Task 1 已定案)。
func insertRouteAssignEvent(ctx context.Context, tx *ent.Tx, order *ent.SalesOrder, actorID uuid.UUID, in AssignRouteInput) error {
	payload := map[string]any{
		"new_delivery_sequence": in.DeliverySequence,
	}
	if order.RouteID != nil {
		payload["old_route_id"] = order.RouteID.String()
	}
	if order.DeliverySequence != nil {
		payload["old_delivery_sequence"] = *order.DeliverySequence
	}
	if in.RouteID != nil {
		payload["new_route_id"] = in.RouteID.String()
	}
	return tx.SalesOrderEvent.Create().
		SetSalesOrderID(order.ID).
		SetCompanyID(order.CompanyID).
		SetEventType("route_assign").
		SetActorID(actorID).
		SetPayload(payload).
		Save(ctx)
}
```

`backend/internal/domain/salesorders/dispatch_handler.go`:

```go
package salesorders

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	protov1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// DispatchHandler 為 DispatchService 的 Connect handler;僅做 proto ↔ usecase 轉換,
// 業務規則全在 DispatchUsecase。認證由 01-plan auth middleware 先行,此處再防一次。
type DispatchHandler struct {
	uc  *DispatchUsecase
	hub *BoardHub // Task 4 接上
}

func NewDispatchHandler(uc *DispatchUsecase) *DispatchHandler {
	return &DispatchHandler{uc: uc}
}

func identityFrom(ctx context.Context) (rls.Identity, error) {
	id, ok := rls.FromContext(ctx)
	if !ok {
		return rls.Identity{}, connect.NewError(connect.CodeUnauthenticated, errors.New("no identity"))
	}
	return id, nil
}

func parseAssignRouteRequest(m *protov1.AssignRouteRequest) (AssignRouteInput, error) {
	var in AssignRouteInput
	oid, err := uuid.Parse(m.GetSalesOrderId())
	if err != nil {
		return in, connect.NewError(connect.CodeInvalidArgument, errors.New("sales_order_id invalid"))
	}
	in.SalesOrderID = oid
	if m.RouteId != nil && *m.RouteId != "" {
		rid, err := uuid.Parse(*m.RouteId)
		if err != nil {
			return in, connect.NewError(connect.CodeInvalidArgument, errors.New("route_id invalid"))
		}
		in.RouteID = &rid
	}
	in.DeliverySequence = m.GetDeliverySequence()
	in.Version = m.GetVersion()
	if m.GetExpectedDeliveryDate() != "" {
		d, err := time.Parse("2006-01-02", m.GetExpectedDeliveryDate())
		if err != nil {
			return in, connect.NewError(connect.CodeInvalidArgument, errors.New("expected_delivery_date invalid"))
		}
		in.ExpectedDeliveryDate = d
	}
	return in, nil
}

func (h *DispatchHandler) AssignRoute(ctx context.Context, req *connect.Request[protov1.AssignRouteRequest]) (*connect.Response[protov1.AssignRouteResponse], error) {
	id, err := identityFrom(ctx)
	if err != nil {
		return nil, err
	}
	in, err := parseAssignRouteRequest(req.Msg)
	if err != nil {
		return nil, err
	}
	res, err := h.uc.AssignRoute(ctx, id, in)
	if err != nil {
		return nil, err
	}
	out := &protov1.AssignRouteResponse{
		SalesOrderId:     res.SalesOrderID.String(),
		DeliverySequence: res.DeliverySequence,
		Version:          res.Version,
	}
	if res.RouteID != nil {
		s := res.RouteID.String()
		out.RouteId = &s
	}
	return connect.NewResponse(out), nil
}
```

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go test -race ./internal/domain/salesorders/ -run TestAssignRoute -v`
Expected: PASS — 成功+重排、過期 version、拖回未指派、四種拒絕、併發僅一勝。

- [ ] **Step 6: Commit**

```bash
git add backend/proto/v1/dispatch.proto backend/gen backend/internal/domain/salesorders
git commit -m "feat(backend): AssignRoute 樂觀鎖與同車次順位重排(5.1.1)"
```

---

### Task 2: 車次批次 Confirm(細部 5.1.2)

**Files:**
- Update: `backend/proto/v1/dispatch.proto`
- Update: `backend/internal/domain/salesorders/dispatch.go`
- Update: `backend/internal/domain/salesorders/dispatch_handler.go`
- Test: `backend/internal/domain/salesorders/dispatch_confirm_test.go`

**Interfaces:**
- Consumes: Task 1 `DispatchUsecase` / 事件詞彙;`CanTransition(from, to string) bool` 與 `Transition(ctx context.Context, tx *ent.Tx, order *ent.SalesOrder, to string, actor uuid.UUID, reason string) error`(05-plan 4.1.3 提供,同套件直接呼叫;`Transition` 負責條件狀態更新 + `dispatch` 事件寫入,version 由呼叫方遞增 — 見 Global Constraints)。
- Produces: proto `DispatchService.Confirm`;`DispatchUsecase.Confirm(ctx, actor rls.Identity, routeID uuid.UUID, date time.Time) (*ConfirmResult, error)`;`ConfirmResult{Results []ConfirmItemResult; SucceededCount int}`、`ConfirmItemResult{SalesOrderID uuid.UUID; Success bool; FailureReason string}`。

- [ ] **Step 1: proto 加 Confirm 並產碼**

`dispatch.proto` 於 service 內加:

```proto
  // Confirm 以車次為批次,確認當日全部待派訂單轉 processing。
  rpc Confirm(ConfirmRequest) returns (ConfirmResponse);
```

訊息:

```proto
message ConfirmRequest {
  string route_id = 1;
  string expected_delivery_date = 2; // YYYY-MM-DD
}
message ConfirmItemResult {
  string sales_order_id = 1;
  bool success = 2;
  string failure_reason = 3; // 失敗時帶原因,成功為空
}
message ConfirmResponse {
  repeated ConfirmItemResult results = 1;
  int32 succeeded_count = 2;
}
```

Run: `cd backend && buf generate`

- [ ] **Step 2: 寫失敗測試(全成功同時戳、併發取消部分失敗、空批次、未指派不納入)**

`backend/internal/domain/salesorders/dispatch_confirm_test.go`:

```go
package salesorders_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	entsalesorder "github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	entsalesorderevent "github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestConfirmBatchAllSuccess(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	ods := []*ent_SalesOrder{} // 佔位,見下方實際宣告
	_ = ods
	o1 := seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	o2 := seedOrder(t, c, co, dept, "W000002", "pending", r, 2)
	o3 := seedOrder(t, c, co, dept, "W000003", "pending", r, 3)
	// 未指派車次的 pending 訂單不得被任何批次納入
	unassigned := seedOrder(t, c, co, dept, "W000004", "pending", nil, 0)

	pub := &fakePublisher{}
	uc := salesorders.NewDispatchUsecase(c, pub, nil)
	res, err := uc.Confirm(ctx, deptAdmin(co, dept), r.ID, boardDate)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if res.SucceededCount != 3 || len(res.Results) != 3 {
		t.Fatalf("want 3 successes, got %+v", res)
	}
	for _, it := range res.Results {
		if !it.Success {
			t.Fatalf("order %s should succeed: %s", it.SalesOrderID, it.FailureReason)
		}
	}
	// 狀態/欄位:全部 processing、dispatched_at 相同、dispatched_by = 操作者、version=2
	var dispatchedAt []*time.Time
	for _, id := range []uuid.UUID{o1.ID, o2.ID, o3.ID} {
		got, _ := c.SalesOrder.Get(ctx, id)
		if got.Status != "processing" || got.DispatchedAt == nil || got.DispatchedBy == nil || got.Version != 2 {
			t.Fatalf("order %s not dispatched correctly: %+v", id, got)
		}
		dispatchedAt = append(dispatchedAt, got.DispatchedAt)
	}
	if !dispatchedAt[0].Equal(*dispatchedAt[1]) || !dispatchedAt[1].Equal(*dispatchedAt[2]) {
		t.Fatal("batch should share one dispatched_at value")
	}
	// 未指派訂單不受影響
	gotUn, _ := c.SalesOrder.Get(ctx, unassigned.ID)
	if gotUn.Status != "pending" || gotUn.DispatchedAt != nil {
		t.Fatal("unassigned order must not be included in any batch")
	}
	// 每筆成功訂單各一筆 dispatch 事件 + 各一則 BoardEvent
	evCount, _ := c.SalesOrderEvent.Query().Where(
		entsalesorderevent.EventTypeEQ("dispatch"),
	).Count(ctx)
	if evCount != 3 {
		t.Fatalf("want 3 dispatch events, got %d", evCount)
	}
	if pub.len() != 3 {
		t.Fatalf("want 3 board events, got %d", pub.len())
	}
	for _, ev := range pub.events {
		if ev.Kind != salesorders.EventDispatch || ev.DepartmentID != dept.ID {
			t.Fatalf("board event mismatch: %+v", ev)
		}
	}
}

func TestConfirmBatchPartialFailure(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	o1 := seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	o2 := seedOrder(t, c, co, dept, "W000002", "pending", r, 2)

	// 模擬併發:候選查詢後 o2 被他人取消(測試掛勾,見實作 candidateHook)
	salesorders.SetCandidateHookForTest(func(hookCtx context.Context) {
		if _, err := c.SalesOrder.UpdateOneID(o2.ID).SetStatus("cancelled").Save(hookCtx); err != nil {
			t.Errorf("hook cancel: %v", err)
		}
	})
	defer salesorders.SetCandidateHookForTest(nil)

	pub := &fakePublisher{}
	uc := salesorders.NewDispatchUsecase(c, pub, nil)
	res, err := uc.Confirm(ctx, deptAdmin(co, dept), r.ID, boardDate)
	if err != nil {
		t.Fatalf("batch should not fail as a whole: %v", err)
	}
	if res.SucceededCount != 1 || len(res.Results) != 2 {
		t.Fatalf("want 1 success 1 failure, got %+v", res)
	}
	var okItem, failItem salesorders.ConfirmItemResult
	for _, it := range res.Results {
		if it.SalesOrderID == o1.ID {
			okItem = it
		} else {
			failItem = it
		}
	}
	if !okItem.Success {
		t.Fatal("o1 should succeed")
	}
	if failItem.Success || failItem.FailureReason == "" {
		t.Fatalf("o2 should fail with reason, got %+v", failItem)
	}
	// 失敗筆欄位不變:無 dispatched_at、狀態仍 cancelled、version 不動
	gotO2, _ := c.SalesOrder.Get(ctx, o2.ID)
	if gotO2.DispatchedAt != nil || gotO2.Status != "cancelled" || gotO2.Version != 1 {
		t.Fatalf("failed order must stay untouched: %+v", gotO2)
	}
	// 失敗筆無 dispatch 事件、無 BoardEvent
	evO2, _ := c.SalesOrderEvent.Query().Where(
		entsalesorderevent.SalesOrderIDEQ(o2.ID),
		entsalesorderevent.EventTypeEQ("dispatch"),
	).Count(ctx)
	if evO2 != 0 || pub.len() != 1 {
		t.Fatalf("failed order must produce neither event row nor board event: rows=%d pub=%d", evO2, pub.len())
	}
}

func TestConfirmEmptyBatchAndValidation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)

	// 當日該車次無 pending 待派訂單 → failed_precondition
	_, err := uc.Confirm(ctx, deptAdmin(co, dept), r.ID, boardDate)
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("empty batch: want failed_precondition, got %v", err)
	}
	// 車次不存在 → not_found
	_, err = uc.Confirm(ctx, deptAdmin(co, dept), uuid.New(), boardDate)
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("ghost route: want not_found, got %v", err)
	}
	// route_id 空 → invalid_argument
	_, err = uc.Confirm(ctx, deptAdmin(co, dept), uuid.Nil, boardDate)
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("nil route: want invalid_argument, got %v", err)
	}
	// 非本部門 staff(data_scope 不含該部門)→ permission_denied
	otherDept, _ := c.Department.Create().SetCompanyID(co.ID).SetName("南區").Save(ctx)
	staff := deptAdmin(co, otherDept)
	staff.Role = "staff"
	seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	_, err = uc.Confirm(ctx, staff, r.ID, boardDate)
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("out-of-scope staff: want permission_denied, got %v", err)
	}
}
```

(測試檔 import 區的 `*ent_SalesOrder` 佔位列為誤植,實作時移除;`boardDate` / `seedCompanyDept` / `seedRoute` / `seedOrder` / `deptAdmin` / `fakePublisher` 定義於 Task 1 的 `dispatch_test.go`,同套件測試共用。)

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run TestConfirm -v`
Expected: FAIL — `DispatchUsecase.Confirm` / `SetCandidateHookForTest` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 Confirm(逐筆交易 + 提交後發佈)**

`backend/internal/domain/salesorders/dispatch.go` 加:

```go
// ---- 車次批次 Confirm(細部 5.1.2)----

// ConfirmItemResult 為批次內單筆結果;成功與失敗明確區分。
type ConfirmItemResult struct {
	SalesOrderID  uuid.UUID
	Success       bool
	FailureReason string
}

type ConfirmResult struct {
	Results         []ConfirmItemResult
	SucceededCount  int
}

// candidateHook 為測試掛勾:候選查詢後、逐筆處理前呼叫(模擬併發取消);生產恆為 nil。
var candidateHook func(ctx context.Context)

// SetCandidateHookForTest 僅供測試設定 candidateHook。
func SetCandidateHookForTest(h func(ctx context.Context)) { candidateHook = h }

// Confirm 以車次為批次確認當日全部待派訂單;每筆訂單各自一個交易,
// 單筆失敗不影響他筆(細部 5.1.2 步驟 3)。
func (u *DispatchUsecase) Confirm(ctx context.Context, actor rls.Identity, routeID uuid.UUID, date time.Time) (*ConfirmResult, error) {
	if routeID == uuid.Nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route_id required"))
	}
	route, err := u.db.Route.Query().Where(
		entroute.IDEQ(routeID),
		entroute.DeletedAtIsNil(),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("route not found"))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := u.authorizeDispatch(actor, route.DepartmentID); err != nil {
		return nil, err
	}

	// 步驟 2:批次候選 — 同車次、同日、同部門、pending、未軟刪;
	// 未指派車次的訂單(route_id 為空)天然不在任何車次篩選內。
	candidates, err := u.db.SalesOrder.Query().Where(
		entsalesorder.RouteIDEQ(routeID),
		entsalesorder.ExpectedDeliveryDateEQ(date),
		entsalesorder.DepartmentIDEQ(route.DepartmentID),
		entsalesorder.StatusEQ("pending"),
		entsalesorder.DeletedAtIsNil(),
	).Order(ent.Asc(entsalesorder.FieldDeliverySequence)).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if len(candidates) == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no pending orders for this route today"))
	}
	if candidateHook != nil {
		candidateHook(ctx)
	}

	// 步驟 3:逐筆處理,批次內各筆共用同一 dispatched_at
	batchTime := time.Now()
	res := &ConfirmResult{Results: make([]ConfirmItemResult, 0, len(candidates))}
	var succeeded []*ent.SalesOrder
	for _, od := range candidates {
		item := ConfirmItemResult{SalesOrderID: od.ID}
		if err := u.confirmOne(ctx, od, actor.UserID, batchTime); err != nil {
			item.FailureReason = err.Error()
		} else {
			item.Success = true
			res.SucceededCount++
			succeeded = append(succeeded, od)
		}
		res.Results = append(res.Results, item)
	}

	// 步驟 4:全部交易提交後 — 每筆成功訂單發佈 BoardEvent;通知觸發於 Task 3 接上。
	for _, od := range succeeded {
		u.pub.Publish(ctx, BoardEvent{
			Kind:             EventDispatch,
			SalesOrderID:     od.ID,
			RouteID:          od.RouteID,
			DeliverySequence: int32(derefInt(od.DeliverySequence)),
			Version:          int64(od.Version + 1),
			DepartmentID:     od.DepartmentID,
		})
		// TODO(接手 Task:本計畫 Task 3):u.notifyDispatched(ctx, route, od)(細部 5.1.4)
	}
	return res, nil
}

// confirmOne 處理單筆:條件更新(WHERE status='pending')設派車欄位 + version,
// 再經 05-plan 的 Transition 流轉狀態並寫 dispatch 事件;任一失敗整筆 rollback。
func (u *DispatchUsecase) confirmOne(ctx context.Context, od *ent.SalesOrder, actorID uuid.UUID, batchTime time.Time) error {
	if !CanTransition(od.Status, "processing") {
		return errors.New("status transition not allowed")
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return err
	}
	n, err := tx.SalesOrder.Update().Where(
		entsalesorder.IDEQ(od.ID),
		entsalesorder.StatusEQ("pending"),
	).SetDispatchedAt(batchTime).SetDispatchedBy(actorID).AddVersion(1).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if n == 0 { // 已被取消/已被他人派車:該筆失敗,不更新任何欄位
		_ = tx.Rollback()
		return errors.New("order status changed concurrently")
	}
	// Transition(05-plan 4.1.3):條件狀態更新 + dispatch 事件,同一 tx。
	if err := Transition(ctx, tx, od, "processing", actorID, ""); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
```

`dispatch_handler.go` 加:

```go
func (h *DispatchHandler) Confirm(ctx context.Context, req *connect.Request[protov1.ConfirmRequest]) (*connect.Response[protov1.ConfirmResponse], error) {
	id, err := identityFrom(ctx)
	if err != nil {
		return nil, err
	}
	rid, err := uuid.Parse(req.Msg.GetRouteId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route_id invalid"))
	}
	date, err := time.Parse("2006-01-02", req.Msg.GetExpectedDeliveryDate())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("expected_delivery_date invalid"))
	}
	res, err := h.uc.Confirm(ctx, id, rid, date)
	if err != nil {
		return nil, err
	}
	out := &protov1.ConfirmResponse{SucceededCount: int32(res.SucceededCount)}
	for _, it := range res.Results {
		out.Results = append(out.Results, &protov1.ConfirmItemResult{
			SalesOrderId:  it.SalesOrderID.String(),
			Success:       it.Success,
			FailureReason: it.FailureReason,
		})
	}
	return connect.NewResponse(out), nil
}
```

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go test -race ./internal/domain/salesorders/ -run 'TestConfirm|TestAssignRoute' -v`
Expected: PASS — 全成功同時戳、部分失敗語義、空批次/驗證、Task 1 測試不迴歸。

```bash
git add backend/proto/v1/dispatch.proto backend/gen backend/internal/domain/salesorders
git commit -m "feat(backend): 車次批次 Confirm 逐筆交易與部分失敗語義(5.1.2)"
```

---

### Task 3: CancelDispatch 與派車通知觸發(細部 5.1.3–5.1.4)

**Files:**
- Update: `backend/proto/v1/dispatch.proto`
- Update: `backend/internal/domain/salesorders/dispatch.go`
- Update: `backend/internal/domain/salesorders/dispatch_handler.go`
- Test: `backend/internal/domain/salesorders/dispatch_cancel_test.go`

**Interfaces:**
- Consumes: Task 1/2 全部;`Transition` 的 `processing → pending` 分支(05-plan 4.1.3 提供:寫 `dispatch_cancel` 事件 + **同交易**寫 `audit_logs`,action = dispatch_cancel);`ent.PrintLog` 產生碼(09-plan 5.5.1 提供,欄位含 `department_id` / `route_id` / `target_date`);`ent.AuditLog` 產生碼(03-plan 提供,測試斷言用);Task 1 已定義的 `DispatchNotifier` 介面(實作由 07-plan Task 6 提供)。
- Produces: proto `DispatchService.CancelDispatch`;`DispatchUsecase.CancelDispatch(ctx, actor rls.Identity, orderID uuid.UUID, reason string, ackReprint bool) (*CancelDispatchResult, error)`;`CancelDispatchResult{SalesOrderID uuid.UUID; Status string; RouteID *uuid.UUID; DeliverySequence int32; ReprintWarning bool; Cancelled bool}`;`DispatchUsecase.notifyDispatched`(private,Confirm 提交後迴圈接上 Task 2 的 TODO)。

- [ ] **Step 1: proto 加 CancelDispatch 並產碼**

`dispatch.proto` 於 service 內加:

```proto
  // CancelDispatch 取消派車(dept_admin 以上,原因必填);已列印車次回重印警告。
  rpc CancelDispatch(CancelDispatchRequest) returns (CancelDispatchResponse);
```

訊息:

```proto
message CancelDispatchRequest {
  string sales_order_id = 1;
  string reason = 2;               // 必填,空白回 invalid_argument
  bool acknowledge_reprint = 3;    // 已列印車次須前端確認後帶 true 重送
}
message CancelDispatchResponse {
  string sales_order_id = 1;
  string status = 2;               // 執行後為 pending;未執行(重印警告)時為原狀態
  optional string route_id = 3;    // 保留的看板位置
  int32 delivery_sequence = 4;
  bool reprint_warning = 5;        // 該車次當日存在正式列印記錄
  bool cancelled = 6;              // false = 因重印警告未執行
}
```

Run: `cd backend && buf generate`

- [ ] **Step 2: 寫失敗測試(回退保留位置、事務事件+稽核、原因必填、重印警告兩段、staff 拒絕、通知觸發)**

`backend/internal/domain/salesorders/dispatch_cancel_test.go`:

```go
package salesorders_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	entauditlog "github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	entsalesorderevent "github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// fakeNotifier 記錄派車通知觸發。
type fakeNotifier struct {
	mu   sync.Mutex
	nts  []salesorders.DispatchNotification
}

func (f *fakeNotifier) NotifyDispatched(_ context.Context, n salesorders.DispatchNotification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nts = append(f.nts, n)
	return nil
}

func (f *fakeNotifier) len() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.nts)
}

// dispatchOne 直接把訂單設為 processing(模擬已派車)。
func markDispatched(t *testing.T, c interface {
}, orderID uuid.UUID, by uuid.UUID) {
	t.Helper()
}

func TestCancelDispatchSuccess(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", r, 2)

	// 先經 Confirm 派車(走正式路徑,產生 dispatched_*)
	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)
	actor := deptAdmin(co, dept)
	if _, err := uc.Confirm(ctx, actor, r.ID, boardDate); err != nil {
		t.Fatalf("confirm: %v", err)
	}

	pub := &fakePublisher{}
	uc2 := salesorders.NewDispatchUsecase(c, pub, nil)
	res, err := uc2.CancelDispatch(ctx, actor, od.ID, "客戶改期", false)
	if err != nil {
		t.Fatalf("cancel dispatch: %v", err)
	}
	if !res.Cancelled || res.Status != "pending" || res.ReprintWarning {
		t.Fatalf("unexpected result: %+v", res)
	}
	// 看板位置保留
	if res.RouteID == nil || *res.RouteID != r.ID || res.DeliverySequence != 2 {
		t.Fatalf("board position should be preserved: %+v", res)
	}
	got, _ := c.SalesOrder.Get(ctx, od.ID)
	if got.Status != "pending" || got.DispatchedAt != nil || got.DispatchedBy != nil {
		t.Fatalf("dispatch fields should be cleared: %+v", got)
	}
	if got.RouteID == nil || *got.DeliverySequence != 2 || got.Version != 3 {
		t.Fatalf("position preserved & version bumped: %+v", got)
	}
	// 同事務:dispatch_cancel 事件(含原因)+ audit_logs
	evExists, _ := c.SalesOrderEvent.Query().Where(
		entsalesorderevent.SalesOrderIDEQ(od.ID),
		entsalesorderevent.EventTypeEQ("dispatch_cancel"),
	).Exist(ctx)
	if !evExists {
		t.Fatal("dispatch_cancel event should exist")
	}
	auditExists, _ := c.AuditLog.Query().Where(
		entauditlog.ActionEQ("dispatch_cancel"),
	).Exist(ctx)
	if !auditExists {
		t.Fatal("audit_logs dispatch_cancel should exist (same tx, D18)")
	}
	// 提交後 BoardEvent(dispatch_cancel,帶保留位置)
	if pub.len() != 1 || pub.events[0].Kind != salesorders.EventDispatchCancel ||
		pub.events[0].RouteID == nil || *pub.events[0].RouteID != r.ID {
		t.Fatalf("board event mismatch: %+v", pub.events)
	}
}

func TestCancelDispatchValidationAndRole(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	pendingOd := seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	dispatched := seedOrder(t, c, co, dept, "W000002", "pending", r, 2)
	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)
	actor := deptAdmin(co, dept)
	if _, err := uc.Confirm(ctx, actor, r.ID, boardDate); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	// 把 W000001 退回 pending 以供「非 processing」案例(直接改,避開正式取消)
	if _, err := c.SalesOrder.UpdateOneID(pendingOd.ID).SetStatus("pending").Save(ctx); err != nil {
		t.Fatalf("reset: %v", err)
	}

	// 原因空白 → invalid_argument,訂單完全不變
	_, err := uc.CancelDispatch(ctx, actor, dispatched.ID, "   ", false)
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("blank reason: want invalid_argument, got %v", err)
	}
	got, _ := c.SalesOrder.Get(ctx, dispatched.ID)
	if got.Status != "processing" || got.DispatchedAt == nil {
		t.Fatal("blank reason must not change the order")
	}
	// 非 processing → failed_precondition
	_, err = uc.CancelDispatch(ctx, actor, pendingOd.ID, "原因", false)
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("pending order: want failed_precondition, got %v", err)
	}
	// staff → permission_denied,訂單維持 processing
	staff := actor
	staff.Role = "staff"
	_, err = uc.CancelDispatch(ctx, staff, dispatched.ID, "原因", false)
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff: want permission_denied, got %v", err)
	}
	got2, _ := c.SalesOrder.Get(ctx, dispatched.ID)
	if got2.Status != "processing" {
		t.Fatal("staff attempt must not change the order")
	}
	// 訂單不存在 → not_found
	_, err = uc.CancelDispatch(ctx, actor, uuid.New(), "原因", false)
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("ghost order: want not_found, got %v", err)
	}
}

func TestCancelDispatchReprintWarning(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nil)
	actor := deptAdmin(co, dept)
	if _, err := uc.Confirm(ctx, actor, r.ID, boardDate); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	// 種一筆當日該車次的正式列印記錄(09-plan 5.5.1;print_previews 不算)
	if _, err := c.PrintLog.Create().
		SetCompanyID(co.ID).SetDepartmentID(dept.ID).
		SetDocumentType("delivery_note").
		SetRouteID(r.ID).SetTargetDate(boardDate).
		SetFileAssetID(uuid.New()).
		SetPrintedBy(actor.UserID).SetPrintedAt(time.Now()).
		SetIsReprint(false).
		Save(ctx); err != nil {
		t.Fatalf("print log: %v", err)
	}

	// 未確認 → 回警告、不執行
	res, err := uc.CancelDispatch(ctx, actor, od.ID, "車輛故障", false)
	if err != nil {
		t.Fatalf("first call should return warning, not error: %v", err)
	}
	if !res.ReprintWarning || res.Cancelled {
		t.Fatalf("want reprint_warning=true & cancelled=false, got %+v", res)
	}
	got, _ := c.SalesOrder.Get(ctx, od.ID)
	if got.Status != "processing" || got.DispatchedAt == nil {
		t.Fatal("warning path must not mutate the order")
	}
	// 確認後 → 照常取消,回應仍帶 reprint_warning=true(最終狀態)
	res2, err := uc.CancelDispatch(ctx, actor, od.ID, "車輛故障", true)
	if err != nil {
		t.Fatalf("acknowledged cancel: %v", err)
	}
	if !res2.Cancelled || !res2.ReprintWarning || res2.Status != "pending" {
		t.Fatalf("unexpected acknowledged result: %+v", res2)
	}
}

func TestConfirmTriggersDispatchNotification(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	seedOrder(t, c, co, dept, "W000001", "pending", r, 1)
	seedOrder(t, c, co, dept, "W000002", "pending", r, 2)

	nt := &fakeNotifier{}
	uc := salesorders.NewDispatchUsecase(c, &fakePublisher{}, nt)
	if _, err := uc.Confirm(ctx, deptAdmin(co, dept), r.ID, boardDate); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	// 每筆成功訂單各觸發一次,變數含訂單編號/車次名稱/預計出貨日(細部 5.1.4)
	if nt.len() != 2 {
		t.Fatalf("want 2 notifications, got %d", nt.len())
	}
	for _, n := range nt.nts {
		if n.RouteName != r.Name || n.ExpectedDeliveryDate != "2026-08-21" ||
			n.OrderNo == "" || n.DepartmentID != dept.ID || n.CompanyID != co.ID {
			t.Fatalf("notification payload mismatch: %+v", n)
		}
	}
}
```

(`markDispatched` 佔位函式為誤植,實作時移除 — 已派車狀態一律經 `Confirm` 建立。)

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run 'TestCancelDispatch|TestConfirmTriggers' -v`
Expected: FAIL — `DispatchUsecase.CancelDispatch` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 CancelDispatch 與通知觸發**

`backend/internal/domain/salesorders/dispatch.go` 加:

```go
// ---- CancelDispatch(細部 5.1.3)----

type CancelDispatchResult struct {
	SalesOrderID     uuid.UUID
	Status           string
	RouteID          *uuid.UUID // 保留的看板位置
	DeliverySequence int32
	ReprintWarning   bool
	Cancelled        bool // false = 因重印警告未執行
}

// CancelDispatch 將 processing 訂單退回 pending,清除派車資訊、保留看板位置;
// 已列印車次須 acknowledge_reprint=true 才執行(細部 5.1.3 步驟 4)。
func (u *DispatchUsecase) CancelDispatch(ctx context.Context, actor rls.Identity, orderID uuid.UUID, reason string, ackReprint bool) (*CancelDispatchResult, error) {
	// 步驟 1:角色門檻(dept_admin 以上;staff 或更低 → permission_denied)
	if !deptAdminOrAbove(actor.Role) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("dept_admin or above required"))
	}
	// 步驟 2:原因必填
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("reason required"))
	}
	// 步驟 3:讀取訂單與狀態檢查
	if orderID == uuid.Nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("sales_order_id required"))
	}
	order, err := u.db.SalesOrder.Query().Where(
		entsalesorder.IDEQ(orderID),
		entsalesorder.DeletedAtIsNil(),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("sales order not found"))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := u.authorizeDispatch(actor, order.DepartmentID); err != nil {
		return nil, err
	}
	if order.Status != "processing" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("order not processing"))
	}

	// 步驟 4:當日該車次正式列印記錄檢查(print_logs;預覽 print_previews 不在此表)
	res := &CancelDispatchResult{
		SalesOrderID: order.ID,
		Status:       order.Status,
		RouteID:      order.RouteID,
	}
	if order.DeliverySequence != nil {
		res.DeliverySequence = int32(*order.DeliverySequence)
	}
	printed := false
	if order.RouteID != nil {
		printed, err = u.db.PrintLog.Query().Where(
			entprintlog.DepartmentIDEQ(order.DepartmentID),
			entprintlog.RouteIDEQ(*order.RouteID),
			entprintlog.TargetDateEQ(order.ExpectedDeliveryDate),
		).Exist(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	res.ReprintWarning = printed
	if printed && !ackReprint {
		return res, nil // 不執行取消,由前端提示確認
	}

	// 步驟 5:交易(D18) — 清派車欄位、退 pending、保留 route/seq;
	// dispatch_cancel 事件 + audit_logs 由 Transition 同交易寫入(05-plan 4.1.3 步驟 4)。
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	n, err := tx.SalesOrder.Update().Where(
		entsalesorder.IDEQ(order.ID),
		entsalesorder.StatusEQ("processing"),
	).ClearDispatchedAt().ClearDispatchedBy().AddVersion(1).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if n == 0 {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("order status changed concurrently"))
	}
	if err := Transition(ctx, tx, order, "pending", actor.UserID, reason); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// 步驟 6:提交後發佈,回應帶保留位置與 reprint_warning 最終狀態
	u.pub.Publish(ctx, BoardEvent{
		Kind:             EventDispatchCancel,
		SalesOrderID:     order.ID,
		RouteID:          order.RouteID,
		DeliverySequence: res.DeliverySequence,
		Version:          int64(order.Version + 2), // Confirm +1、本次 +1
		DepartmentID:     order.DepartmentID,
	})
	res.Status = "pending"
	res.Cancelled = true
	return res, nil
}

// ---- 派車通知觸發(細部 5.1.4)----

// notifyDispatched 於批次交易提交後逐筆觸發;fire-and-record,
// 發送失敗僅由通知模組落 notifications.status=failed,不影響派車結果(D16)。
func (u *DispatchUsecase) notifyDispatched(ctx context.Context, route *ent.Route, od *ent.SalesOrder) {
	err := u.notifier.NotifyDispatched(ctx, DispatchNotification{
		CompanyID:            od.CompanyID,
		DepartmentID:         od.DepartmentID,
		SalesOrderID:         od.ID,
		CustomerID:           od.CustomerID,
		OrderNo:              od.OrderNo,
		RouteName:            route.Name,
		ExpectedDeliveryDate: od.ExpectedDeliveryDate.Format("2006-01-02"),
	})
	if err != nil {
		log.Printf("dispatch notify failed order=%s: %v", od.ID, err)
	}
}
```

Task 2 留下的 TODO 接上 — `Confirm` 的提交後迴圈改為:

```go
	for _, od := range succeeded {
		u.pub.Publish(ctx, BoardEvent{ /* …同 Task 2… */ })
		u.notifyDispatched(ctx, route, od) // 細部 5.1.4:交易提交後逐筆觸發
	}
```

(`Version: order.Version + 2` 的假設為該單先經 Confirm +1;更穩妥的作法是在條件更新後於 tx 內重讀 version 帶入事件 — 實作時改為 `CancelDispatch` 交易提交後以 `tx` 外查詢重讀 `got.Version` 作為事件版本,以此為準,測試斷言事件 version 與 DB 一致。)

`dispatch_handler.go` 加:

```go
func (h *DispatchHandler) CancelDispatch(ctx context.Context, req *connect.Request[protov1.CancelDispatchRequest]) (*connect.Response[protov1.CancelDispatchResponse], error) {
	id, err := identityFrom(ctx)
	if err != nil {
		return nil, err
	}
	oid, err := uuid.Parse(req.Msg.GetSalesOrderId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("sales_order_id invalid"))
	}
	res, err := h.uc.CancelDispatch(ctx, id, oid, req.Msg.GetReason(), req.Msg.GetAcknowledgeReprint())
	if err != nil {
		return nil, err
	}
	out := &protov1.CancelDispatchResponse{
		SalesOrderId:     res.SalesOrderID.String(),
		Status:           res.Status,
		DeliverySequence: res.DeliverySequence,
		ReprintWarning:   res.ReprintWarning,
		Cancelled:        res.Cancelled,
	}
	if res.RouteID != nil {
		s := res.RouteID.String()
		out.RouteId = &s
	}
	return connect.NewResponse(out), nil
}
```

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go test -race ./internal/domain/salesorders/ -v`
Expected: PASS — 取消回退保留位置、事件+稽核同事務、原因必填、重印警告兩段、staff 拒絕、通知兩筆觸發;Task 1/2 測試不迴歸。

```bash
git add backend/proto/v1/dispatch.proto backend/gen backend/internal/domain/salesorders
git commit -m "feat(backend): CancelDispatch 重印警告與派車通知觸發介面(5.1.3-5.1.4)"
```

---

### Task 4: WatchBoard proto 與串流 handler(細部 5.2.1)

**Files:**
- Update: `backend/proto/v1/dispatch.proto`
- Create: `backend/internal/domain/salesorders/watch_board.go`
- Update: `backend/internal/domain/salesorders/dispatch_handler.go`
- Test: `backend/internal/domain/salesorders/watch_board_test.go`

**Interfaces:**
- Consumes: Task 1 `BoardEvent` / `BoardEventKind` / `BoardPublisher`;`rls.Identity`(01-plan)。
- Produces: proto `DispatchService.WatchBoard`(server streaming)與 `BoardEvent` 訊息、`BoardEventKind` enum;`salesorders.BoardHub`(`NewBoardHub() *BoardHub`、`Subscribe(deptID uuid.UUID) (<-chan BoardEvent, func())`、`Publish(ctx, ev BoardEvent)` 實作 `BoardPublisher`、`LocalSubscriberCount(deptID uuid.UUID) int` 測試用);`DispatchHandler.WatchBoard`;`NewDispatchHandler(uc *DispatchUsecase, hub *BoardHub)`(簽名自本 Task 起含 hub,Task 1-3 測試不需 handler,僅 `InitDomains()` 組裝點需更新)。

- [ ] **Step 1: proto 加 WatchBoard 與 BoardEvent 並產碼**

`dispatch.proto` 於 service 內加:

```proto
  // WatchBoard 訂閱本部門看板事件(server streaming);認證走既有 cookie/JWT。
  rpc WatchBoard(WatchBoardRequest) returns (stream BoardEvent);
```

訊息:

```proto
message WatchBoardRequest {
  string expected_delivery_date = 1; // YYYY-MM-DD;僅供前端語意,後端按部門廣播
}

enum BoardEventKind {
  BOARD_EVENT_KIND_UNSPECIFIED = 0;
  BOARD_EVENT_KIND_ROUTE_ASSIGN = 1;
  BOARD_EVENT_KIND_DISPATCH = 2;
  BOARD_EVENT_KIND_DISPATCH_CANCEL = 3;
  BOARD_EVENT_KIND_HEARTBEAT = 4; // 無業務欄位(Task 5 發送)
}

message BoardEvent {
  BoardEventKind kind = 1;
  string sales_order_id = 2;
  optional string route_id = 3;
  int32 delivery_sequence = 4;
  int64 version = 5;
  string department_id = 6;
}
```

Run: `cd backend && buf generate`

- [ ] **Step 2: 寫失敗測試(in-process Connect client 驗證事件送達、部門隔離、未認證拒絕、斷線清理)**

`backend/internal/domain/salesorders/watch_board_test.go`:

```go
package salesorders_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	protov1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	dispatchv1connect "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1/v1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
)

// withIdentity 模擬 01-plan auth middleware 注入身分;id 為 nil 時不注入(未認證)。
func withIdentity(id *rls.Identity, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id != nil {
			r = r.WithContext(rls.NewContext(r.Context(), *id))
		}
		next.ServeHTTP(w, r)
	})
}

func newBoardServer(t *testing.T, hub *salesorders.BoardHub, id *rls.Identity) (*httptest.Server, dispatchv1connect.DispatchServiceClient) {
	t.Helper()
	h := salesorders.NewDispatchHandler(nil, hub)
	mux := http.NewServeMux()
	path, handler := dispatchv1connect.NewDispatchServiceHandler(h)
	mux.Handle(path, withIdentity(id, handler))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, dispatchv1connect.NewDispatchServiceClient(srv.Client(), srv.URL)
}

func TestWatchBoardReceivesDepartmentEvents(t *testing.T) {
	coID, deptA := uuid.New(), uuid.New()
	idA := &rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &deptA,
		DataScope: "department", Role: "dept_admin"}
	hub := salesorders.NewBoardHub()
	_, client := newBoardServer(t, hub, idA)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{
		ExpectedDeliveryDate: "2026-08-21",
	}))
	if err != nil {
		t.Fatalf("watch: %v", err)
	}
	orderID := uuid.New()
	routeID := uuid.New()
	hub.Publish(ctx, salesorders.BoardEvent{
		Kind: salesorders.EventDispatch, SalesOrderID: orderID,
		RouteID: &routeID, DeliverySequence: 3, Version: 2, DepartmentID: deptA,
	})
	if !stream.Receive() {
		t.Fatalf("should receive event: %v", stream.Err())
	}
	msg := stream.Msg()
	if msg.GetKind() != protov1.BoardEventKind_BOARD_EVENT_KIND_DISPATCH ||
		msg.GetSalesOrderId() != orderID.String() ||
		msg.GetRouteId() != routeID.String() ||
		msg.GetDeliverySequence() != 3 || msg.GetVersion() != 2 ||
		msg.GetDepartmentId() != deptA.String() {
		t.Fatalf("BoardEvent fields mismatch: %+v", msg)
	}
}

func TestWatchBoardDepartmentIsolation(t *testing.T) {
	coID := uuid.New()
	deptA, deptB := uuid.New(), uuid.New()
	idA := &rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &deptA,
		DataScope: "department", Role: "staff"}
	hub := salesorders.NewBoardHub()
	_, client := newBoardServer(t, hub, idA)

	// 短時窗:部門 B 事件不應送達部門 A 連線(Task 4 尚無 heartbeat,通道應靜默)
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	stream, err := client.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err != nil {
		t.Fatalf("watch: %v", err)
	}
	hub.Publish(ctx, salesorders.BoardEvent{
		Kind: salesorders.EventDispatch, SalesOrderID: uuid.New(),
		Version: 2, DepartmentID: deptB,
	})
	if stream.Receive() {
		t.Fatalf("dept A connection must not receive dept B event: %+v", stream.Msg())
	}
	if connect.CodeOf(stream.Err()) != connect.CodeDeadlineExceeded {
		t.Fatalf("expect silence until ctx deadline, got %v", stream.Err())
	}
}

func TestWatchBoardUnauthenticatedAndNoDepartment(t *testing.T) {
	hub := salesorders.NewBoardHub()

	// 未認證(無身分注入)→ unauthenticated,且不建立訂閱
	_, anonClient := newBoardServer(t, hub, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	s1, err := anonClient.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err == nil {
		for s1.Receive() {
		}
		err = s1.Err()
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous: want unauthenticated, got %v", err)
	}
	if n := hub.LocalSubscriberCount(uuid.New()); n != 0 {
		t.Fatalf("no subscription should be registered, got %d", n)
	}

	// 無部門上下文(DepartmentID nil)→ failed_precondition
	noDept := &rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(),
		DataScope: "company", Role: "company_admin"}
	_, noDeptClient := newBoardServer(t, hub, noDept)
	s2, err := noDeptClient.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err == nil {
		for s2.Receive() {
		}
		err = s2.Err()
	}
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("no department: want failed_precondition, got %v", err)
	}
}

func TestWatchBoardDisconnectCleansRegistry(t *testing.T) {
	coID, deptA := uuid.New(), uuid.New()
	idA := &rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &deptA,
		DataScope: "department", Role: "staff"}
	hub := salesorders.NewBoardHub()
	_, client := newBoardServer(t, hub, idA)

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err != nil {
		t.Fatalf("watch: %v", err)
	}
	_ = stream
	// 等註冊生效
	deadline := time.Now().Add(2 * time.Second)
	for hub.LocalSubscriberCount(deptA) != 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.LocalSubscriberCount(deptA) != 1 {
		t.Fatal("subscription should be registered")
	}
	cancel() // 客戶端斷線
	deadline = time.Now().Add(2 * time.Second)
	for hub.LocalSubscriberCount(deptA) != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := hub.LocalSubscriberCount(deptA); n != 0 {
		t.Fatalf("registry should be empty after disconnect, got %d", n)
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run TestWatchBoard -v`
Expected: FAIL — `salesorders.NewBoardHub` / `DispatchHandler.WatchBoard` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 BoardHub(程序內註冊表)與 WatchBoard handler**

`backend/internal/domain/salesorders/watch_board.go`:

```go
package salesorders

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// BoardHub 為看板連線註冊表(department_id → 本機連線集合)。
// Task 4 為程序內轉發;Task 5 於此檔加入 Valkey relay 與 metrics。
type BoardHub struct {
	mu    sync.Mutex
	conns map[uuid.UUID]map[chan BoardEvent]struct{}
}

func NewBoardHub() *BoardHub {
	return &BoardHub{conns: make(map[uuid.UUID]map[chan BoardEvent]struct{})}
}

// Subscribe 登記部門連線;回傳事件 channel 與退訂函式(退訂必須呼叫以釋放資源,冪等)。
func (h *BoardHub) Subscribe(deptID uuid.UUID) (<-chan BoardEvent, func()) {
	ch := make(chan BoardEvent, 16)
	h.mu.Lock()
	if h.conns[deptID] == nil {
		h.conns[deptID] = make(map[chan BoardEvent]struct{})
	}
	h.conns[deptID][ch] = struct{}{}
	h.mu.Unlock()
	var once sync.Once
	return ch, func() {
		once.Do(func() { h.unsubscribe(deptID, ch) })
	}
}

func (h *BoardHub) unsubscribe(deptID uuid.UUID, ch chan BoardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[deptID], ch)
	if len(h.conns[deptID]) == 0 {
		delete(h.conns, deptID)
	}
	close(ch)
}

// Publish 實作 BoardPublisher;Task 4 僅本機轉發,Task 5 改經 Valkey(單一路徑)。
func (h *BoardHub) Publish(_ context.Context, ev BoardEvent) {
	h.publishLocal(ev)
}

// publishLocal 轉發該部門全部本機連線;慢連線(滿 buffer)丟棄,
// 由前端重連後全量重查吸收(D14)。
func (h *BoardHub) publishLocal(ev BoardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.conns[ev.DepartmentID] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// LocalSubscriberCount 供測試斷言註冊表狀態。
func (h *BoardHub) LocalSubscriberCount(deptID uuid.UUID) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns[deptID])
}
```

`dispatch_handler.go` — 建構子改含 hub,並加 WatchBoard:

```go
func NewDispatchHandler(uc *DispatchUsecase, hub *BoardHub) *DispatchHandler {
	return &DispatchHandler{uc: uc, hub: hub}
}

// WatchBoard 訂閱本部門看板事件(細部 5.2.1);事件僅轉發不過濾日期(步驟 5)。
func (h *DispatchHandler) WatchBoard(ctx context.Context, req *connect.Request[protov1.WatchBoardRequest], stream *connect.ServerStream[protov1.BoardEvent]) error {
	// 步驟 1:auth interceptor 先行,此處再防一次
	id, err := identityFrom(ctx)
	if err != nil {
		return err
	}
	// 功能權限:obj=dispatch act=read(enf 可為 nil,同 usecase 慣例)
	if h.uc != nil && h.uc.enf != nil {
		ok, err := h.uc.enf.Enforce(id.Role, id.CompanyID.String(), "dispatch", "read")
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		if !ok {
			return connect.NewError(connect.CodePermissionDenied, errors.New("no board permission"))
		}
	}
	// 日期僅供前端語意;若提供則須為合法 YYYY-MM-DD
	if d := req.Msg.GetExpectedDeliveryDate(); d != "" {
		if _, err := time.Parse("2006-01-02", d); err != nil {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("expected_delivery_date invalid"))
		}
	}
	// 步驟 2:部門上下文
	if id.DepartmentID == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("department context required"))
	}
	dept := *id.DepartmentID
	// 步驟 3:登記訂閱
	ch, unsub := h.hub.Subscribe(dept)
	defer unsub()
	// 步驟 4:迴圈轉發;客戶端斷線(ctx 取消)→ defer 清理,靜默結束
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(boardEventToProto(ev)); err != nil {
				return nil // 寫入失敗 = 客戶端已離線,清理註冊並結束
			}
		}
	}
}

func boardEventToProto(ev BoardEvent) *protov1.BoardEvent {
	out := &protov1.BoardEvent{
		Kind:             boardKindToProto(ev.Kind),
		SalesOrderId:     ev.SalesOrderID.String(),
		DeliverySequence: ev.DeliverySequence,
		Version:          ev.Version,
		DepartmentId:     ev.DepartmentID.String(),
	}
	if ev.RouteID != nil {
		s := ev.RouteID.String()
		out.RouteId = &s
	}
	return out
}

func boardKindToProto(k BoardEventKind) protov1.BoardEventKind {
	switch k {
	case EventRouteAssign:
		return protov1.BoardEventKind_BOARD_EVENT_KIND_ROUTE_ASSIGN
	case EventDispatch:
		return protov1.BoardEventKind_BOARD_EVENT_KIND_DISPATCH
	case EventDispatchCancel:
		return protov1.BoardEventKind_BOARD_EVENT_KIND_DISPATCH_CANCEL
	case EventHeartbeat:
		return protov1.BoardEventKind_BOARD_EVENT_KIND_HEARTBEAT
	}
	return protov1.BoardEventKind_BOARD_EVENT_KIND_UNSPECIFIED
}
```

`InitDomains()` 組裝點更新:`NewDispatchHandler(uc, hub)`;`uc` 以 `NewDispatchUsecase(db, hub, notifier)` 建立(hub 同時是 BoardPublisher)。

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go test -race ./internal/domain/salesorders/ -v`
Expected: PASS — 事件送達且欄位齊全、部門隔離、未認證/無部門拒絕、斷線清理;Task 1-3 測試不迴歸。

```bash
git add backend/proto/v1/dispatch.proto backend/gen backend/internal/domain/salesorders backend/internal/server
git commit -m "feat(backend): WatchBoard 串流 handler 與程序內 BoardHub(5.2.1)"
```

---

### Task 5: Valkey pub/sub 跨 replica 與 heartbeat(細部 5.2.2–5.2.3)

**Files:**
- Create: `backend/internal/testutil/valkey.go`
- Update: `backend/internal/domain/salesorders/watch_board.go`
- Update: `backend/internal/domain/salesorders/dispatch_handler.go`
- Update: `backend/config/api.go`
- Update: `backend/internal/server/domains.go`
- Test: `backend/internal/domain/salesorders/watch_board_valkey_test.go`

**Interfaces:**
- Consumes: Task 4 `BoardHub` / WatchBoard;`redis/go-redis/v9`(01-plan Task 4 已引入);`testutil.NewEntClient`。
- Produces: `testutil.NewValkeyClient(t *testing.T) *redis.Client`;`BoardHub` 改建構 `NewBoardHub(rdb *redis.Client, logger *log.Logger) *BoardHub`(Task 4 測試同步更新建構引數)、`ActiveRelayCount() int`(測試用);channel 命名 `board:<department_id>`;config `BoardHeartbeatSeconds`(預設 25)/ `IngressIdleTimeoutSeconds`(預設 30);`DispatchHandler` 加 `heartbeat time.Duration` 欄位,`NewDispatchHandler(uc, hub, heartbeat)`(`InitDomains()` 組裝點同步);metrics `board_events_published_total` / `board_events_relayed_total` / `board_connections`。

- [ ] **Step 1: testutil.NewValkeyClient**

`backend/internal/testutil/valkey.go`:

```go
package testutil

import (
	"context"
	"net"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewValkeyClient 啟動一個拋棄式 Valkey container,回傳可用於測試的 redis 客戶端。
// 測試結束自動銷毀。
func NewValkeyClient(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "valkey/valkey:8-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start valkey container: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })

	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("container host: %v", err)
	}
	port, err := c.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("container port: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, port.Port())})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping valkey: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}
```

Run: `cd backend && go get github.com/redis/go-redis/v9 github.com/prometheus/client_golang`

- [ ] **Step 2: 寫失敗測試(跨 replica 送達、rollback 不發事件、Valkey 故障不影響 mutation、heartbeat、退訂停 relay)**

`backend/internal/domain/salesorders/watch_board_valkey_test.go`:

```go
package salesorders_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	protov1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	dispatchv1connect "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1/v1connect"
	entsalesorderevent "github.com/salesorder/sales-order-1.0/backend/ent/salesorderevent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func testLogger() *log.Logger { return log.New(os.Stderr, "boardtest ", 0) }

// rawSubscribe 直接訂閱 Valkey channel,用於斷言「有/無事件發佈」。
func rawSubscribe(t *testing.T, rdb *redis.Client, channel string) <-chan *redis.Message {
	t.Helper()
	ps := rdb.Subscribe(context.Background(), channel)
	t.Cleanup(func() { _ = ps.Close() })
	return ps.Channel()
}

func drain(ch <-chan *redis.Message, within time.Duration) int {
	n := 0
	timer := time.NewTimer(within)
	defer timer.Stop()
	for {
		select {
		case <-ch:
			n++
		case <-timer.C:
			return n
		}
	}
}

func TestCrossReplicaEventDelivery(t *testing.T) {
	c := testutil.NewEntClient(t)
	rdb := testutil.NewValkeyClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	otherDept, err := c.Department.Create().SetCompanyID(co.ID).SetName("南區").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", nil, 0)

	// replica 1:mutation 端;replica 2:看板連線端(兩個 BoardHub 共享同一 Valkey)
	hubMutation := salesorders.NewBoardHub(rdb, testLogger())
	hubBoard := salesorders.NewBoardHub(rdb, testLogger())
	uc := salesorders.NewDispatchUsecase(c, hubMutation, nil)
	actor := deptAdmin(co, dept)

	// replica 2 上的 WatchBoard handler + in-process client
	idA := &rls.Identity{UserID: actor.UserID, CompanyID: co.ID, DepartmentID: &dept.ID,
		DataScope: "department", Role: "dept_admin"}
	h := salesorders.NewDispatchHandler(uc, hubBoard, time.Hour) // 關掉 heartbeat 干擾
	mux := http.NewServeMux()
	path, handler := dispatchv1connect.NewDispatchServiceHandler(h)
	mux.Handle(path, withIdentity(idA, handler))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := dispatchv1connect.NewDispatchServiceClient(srv.Client(), srv.URL)

	watchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	stream, err := client.WatchBoard(watchCtx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err != nil {
		t.Fatalf("watch: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for hubBoard.LocalSubscriberCount(dept.ID) != 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	// replica 1 執行三種 mutation 之一:AssignRoute
	if _, err := uc.AssignRoute(ctx, actor, salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: &r.ID, DeliverySequence: 1,
		Version: 1, ExpectedDeliveryDate: boardDate,
	}); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if !stream.Receive() {
		t.Fatalf("replica 2 should receive event via Valkey: %v", stream.Err())
	}
	msg := stream.Msg()
	if msg.GetKind() != protov1.BoardEventKind_BOARD_EVENT_KIND_ROUTE_ASSIGN ||
		msg.GetSalesOrderId() != od.ID.String() || msg.GetVersion() != 2 {
		t.Fatalf("event mismatch: %+v", msg)
	}

	// 部門隔離:otherDept 連線收不到 dept 事件
	idB := &rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &otherDept.ID,
		DataScope: "department", Role: "staff"}
	mux2 := http.NewServeMux()
	path2, handler2 := dispatchv1connect.NewDispatchServiceHandler(h)
	_ = path2
	mux2.Handle(path, withIdentity(idB, handler2))
	srv2 := httptest.NewServer(mux2)
	t.Cleanup(srv2.Close)
	clientB := dispatchv1connect.NewDispatchServiceClient(srv2.Client(), srv2.URL)
	shortCtx, shortCancel := context.WithTimeout(ctx, 400*time.Millisecond)
	defer shortCancel()
	streamB, err := clientB.WatchBoard(shortCtx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err != nil {
		t.Fatalf("watch B: %v", err)
	}
	// 再觸發一筆 dept 事件(Confirm)
	seedOrder(t, c, co, dept, "W000002", "pending", r, 2)
	if _, err := uc.Confirm(ctx, actor, r.ID, boardDate); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if streamB.Receive() {
		t.Fatalf("dept B connection must not receive dept A events: %+v", streamB.Msg())
	}
}

func TestRollbackPublishesNoEvent(t *testing.T) {
	c := testutil.NewEntClient(t)
	rdb := testutil.NewValkeyClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", nil, 0)

	hub := salesorders.NewBoardHub(rdb, testLogger())
	uc := salesorders.NewDispatchUsecase(c, hub, nil)
	msgs := rawSubscribe(t, rdb, "board:"+dept.ID.String())

	// version 衝突 → 交易 rollback → 不得有任何事件發佈
	_, err := uc.AssignRoute(ctx, deptAdmin(co, dept), salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: &r.ID, DeliverySequence: 1,
		Version: 99, ExpectedDeliveryDate: boardDate,
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("want failed_precondition, got %v", err)
	}
	if n := drain(msgs, 300*time.Millisecond); n != 0 {
		t.Fatalf("rollback must publish nothing, got %d messages", n)
	}
	// 空批次 Confirm(無交易建立)同樣不發事件
	_, err = uc.Confirm(ctx, deptAdmin(co, dept), r.ID, boardDate)
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("empty batch: want failed_precondition, got %v", err)
	}
	if n := drain(msgs, 300*time.Millisecond); n != 0 {
		t.Fatalf("empty batch must publish nothing, got %d messages", n)
	}
}

func TestPublishFailureDoesNotFailMutation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dept := seedCompanyDept(t, c)
	r := seedRoute(t, c, co, dept, "A")
	od := seedOrder(t, c, co, dept, "W000001", "pending", nil, 0)

	// 指向不存在 Valkey 的 client:發佈必失敗
	dead := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 100 * time.Millisecond})
	t.Cleanup(func() { _ = dead.Close() })
	hub := salesorders.NewBoardHub(dead, testLogger())
	uc := salesorders.NewDispatchUsecase(c, hub, nil)

	res, err := uc.AssignRoute(ctx, deptAdmin(co, dept), salesorders.AssignRouteInput{
		SalesOrderID: od.ID, RouteID: &r.ID, DeliverySequence: 1,
		Version: 1, ExpectedDeliveryDate: boardDate,
	})
	if err != nil {
		t.Fatalf("Valkey outage must not fail mutation: %v", err)
	}
	if res.Version != 2 {
		t.Fatalf("mutation should succeed with version bump, got %+v", res)
	}
}

func TestHeartbeatKeepsIdleStreamAlive(t *testing.T) {
	c := testutil.NewEntClient(t)
	rdb := testutil.NewValkeyClient(t)
	co, dept := seedCompanyDept(t, c)
	idA := &rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &dept.ID,
		DataScope: "department", Role: "staff"}
	hub := salesorders.NewBoardHub(rdb, testLogger())
	h := salesorders.NewDispatchHandler(nil, hub, 50*time.Millisecond) // 測試用短間隔
	mux := http.NewServeMux()
	path, handler := dispatchv1connect.NewDispatchServiceHandler(h)
	mux.Handle(path, withIdentity(idA, handler))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := dispatchv1connect.NewDispatchServiceClient(srv.Client(), srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{}))
	if err != nil {
		t.Fatalf("watch: %v", err)
	}
	msgs := rawSubscribe(t, rdb, "board:"+dept.ID.String())
	evBefore, _ := c.SalesOrderEvent.Query().Count(ctx)

	// 無業務事件:間隔內持續收到 heartbeat
	beats := 0
	for beats < 2 {
		if !stream.Receive() {
			t.Fatalf("stream closed while waiting heartbeat: %v", stream.Err())
		}
		if stream.Msg().GetKind() == protov1.BoardEventKind_BOARD_EVENT_KIND_HEARTBEAT {
			beats++
		}
	}
	// heartbeat 不進 Valkey、不寫業務事件表
	if n := drain(msgs, 200*time.Millisecond); n != 0 {
		t.Fatalf("heartbeat must not enter Valkey channel, got %d", n)
	}
	evAfter, _ := c.SalesOrderEvent.Query().Where(
		entsalesorderevent.EventTypeEQ("heartbeat"),
	).Count(ctx)
	if evAfter != evBefore && evAfter != 0 {
		t.Fatalf("heartbeat must not write sales_order_events: before=%d after=%d", evBefore, evAfter)
	}
}

func TestLastDisconnectStopsValkeyRelay(t *testing.T) {
	rdb := testutil.NewValkeyClient(t)
	deptA := uuid.New()
	idA := &rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DepartmentID: &deptA,
		DataScope: "department", Role: "staff"}
	hub := salesorders.NewBoardHub(rdb, testLogger())
	h := salesorders.NewDispatchHandler(nil, hub, time.Hour)
	mux := http.NewServeMux()
	path, handler := dispatchv1connect.NewDispatchServiceHandler(h)
	mux.Handle(path, withIdentity(idA, handler))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := dispatchv1connect.NewDispatchServiceClient(srv.Client(), srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	if _, err := client.WatchBoard(ctx, connect.NewRequest(&protov1.WatchBoardRequest{})); err != nil {
		t.Fatalf("watch: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for hub.ActiveRelayCount() != 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.ActiveRelayCount() != 1 {
		t.Fatal("relay should start for subscribed department")
	}
	cancel()
	deadline = time.Now().Add(2 * time.Second)
	for (hub.LocalSubscriberCount(deptA) != 0 || hub.ActiveRelayCount() != 0) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.ActiveRelayCount() != 0 {
		t.Fatal("relay should stop when no local connection remains for the department")
	}
}
```

Task 4 測試同步更新:`NewBoardHub()` → `NewBoardHub(testutil.NewValkeyClient(t), testLogger())`;`NewDispatchHandler(nil, hub)` → `NewDispatchHandler(nil, hub, time.Hour)`(不需 heartbeat 的測試用極長間隔)。`watch_board_test.go` 需 import `testutil` 與 `log/os`(或抽出共用 `testLogger`)。

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run 'TestCrossReplica|TestRollback|TestPublishFailure|TestHeartbeat|TestLastDisconnect' -v`
Expected: FAIL — `NewBoardHub` 簽名不符 / `ActiveRelayCount` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 Valkey relay、heartbeat 與 config**

`backend/internal/domain/salesorders/watch_board.go` 全檔改寫為:

```go
package salesorders

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

// 看板指標(D19;細部 5.2.2 步驟 5)。
var (
	boardPublishedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "board_events_published_total", Help: "Board events published to Valkey pub/sub",
	})
	boardRelayedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "board_events_relayed_total", Help: "Board events relayed from Valkey to local connections",
	})
	boardConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "board_connections", Help: "Active WatchBoard connections on this replica",
	})
)

// boardChannel 一部門一 channel(細部 5.2.2 介面)。
func boardChannel(dept uuid.UUID) string { return "board:" + dept.String() }

// BoardHub 為看板連線註冊表 + Valkey 訂閱轉發層:
// 發佈一律經 Valkey(單一路徑);每 replica 僅對「本機有連線的部門」啟 relay goroutine,
// 最後一條連線離開即退訂並停止 relay。
type BoardHub struct {
	mu     sync.Mutex
	conns  map[uuid.UUID]map[chan BoardEvent]struct{}
	relays map[uuid.UUID]context.CancelFunc
	rdb    *redis.Client
	logger *log.Logger
}

func NewBoardHub(rdb *redis.Client, logger *log.Logger) *BoardHub {
	if logger == nil {
		logger = log.Default()
	}
	return &BoardHub{
		conns:  make(map[uuid.UUID]map[chan BoardEvent]struct{}),
		relays: make(map[uuid.UUID]context.CancelFunc),
		rdb:    rdb,
		logger: logger,
	}
}

// Publish 實作 BoardPublisher:DB 交易提交成功後由 mutation 呼叫。
// 發佈失敗僅記錄,不影響 mutation 結果、不重試(at-most-once,D14;細部 5.2.2 步驟 4)。
func (h *BoardHub) Publish(ctx context.Context, ev BoardEvent) {
	if ev.Kind == EventHeartbeat {
		return // heartbeat 純連線維持,不進 Valkey(細部 5.2.3 步驟 3)
	}
	b, err := json.Marshal(ev)
	if err != nil {
		h.logger.Printf("board event marshal failed: %v", err)
		return
	}
	if err := h.rdb.Publish(ctx, boardChannel(ev.DepartmentID), b).Err(); err != nil {
		h.logger.Printf("board publish failed dept=%s kind=%s: %v", ev.DepartmentID, ev.Kind, err)
		return
	}
	boardPublishedTotal.Inc()
}

// Subscribe 登記部門連線;該部門首條連線建立時啟動 Valkey relay。
func (h *BoardHub) Subscribe(deptID uuid.UUID) (<-chan BoardEvent, func()) {
	ch := make(chan BoardEvent, 16)
	h.mu.Lock()
	if h.conns[deptID] == nil {
		h.conns[deptID] = make(map[chan BoardEvent]struct{})
	}
	h.conns[deptID][ch] = struct{}{}
	if _, ok := h.relays[deptID]; !ok {
		relayCtx, cancel := context.WithCancel(context.Background())
		h.relays[deptID] = cancel
		go h.relay(relayCtx, deptID)
	}
	h.mu.Unlock()
	boardConnections.Inc()
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			boardConnections.Dec()
			h.unsubscribe(deptID, ch)
		})
	}
}

func (h *BoardHub) unsubscribe(deptID uuid.UUID, ch chan BoardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[deptID], ch)
	close(ch)
	if len(h.conns[deptID]) == 0 {
		delete(h.conns, deptID)
		if cancel, ok := h.relays[deptID]; ok { // 本 replica 已無該部門連線 → 退訂 channel
			cancel()
			delete(h.relays, deptID)
		}
	}
}

// relay 訂閱部門 channel 並轉發本機連線;斷線自動重連,
// 重連期間事件遺失由前端重連後全量重查吸收,後端不補發(D14;細部 5.2.2 錯誤處理)。
func (h *BoardHub) relay(ctx context.Context, dept uuid.UUID) {
	for {
		ps := h.rdb.Subscribe(ctx, boardChannel(dept))
		err := h.relayOnce(ctx, ps)
		_ = ps.Close()
		if err == nil || ctx.Err() != nil {
			return
		}
		h.logger.Printf("board relay reconnect dept=%s: %v", dept, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (h *BoardHub) relayOnce(ctx context.Context, ps *redis.PubSub) error {
	ch := ps.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return errors.New("valkey subscription channel closed")
			}
			var ev BoardEvent
			if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
				h.logger.Printf("board event decode failed, dropped: %v", err) // 不影響其他連線
				continue
			}
			boardRelayedTotal.Inc()
			h.publishLocal(ev)
		}
	}
}

// publishLocal 轉發該部門全部本機連線;慢連線(滿 buffer)丟棄(D14)。
func (h *BoardHub) publishLocal(ev BoardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.conns[ev.DepartmentID] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// LocalSubscriberCount / ActiveRelayCount 供測試斷言註冊表與 relay 狀態。
func (h *BoardHub) LocalSubscriberCount(deptID uuid.UUID) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns[deptID])
}

func (h *BoardHub) ActiveRelayCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.relays)
}
```

`dispatch_handler.go` — 建構子加 heartbeat 間隔,WatchBoard 加 ticker(細部 5.2.3 步驟 1-2):

```go
type DispatchHandler struct {
	uc        *DispatchUsecase
	hub       *BoardHub
	heartbeat time.Duration // WatchBoard heartbeat 間隔;<=0 時停用(測試用)
}

func NewDispatchHandler(uc *DispatchUsecase, hub *BoardHub, heartbeat time.Duration) *DispatchHandler {
	return &DispatchHandler{uc: uc, hub: hub, heartbeat: heartbeat}
}
```

`WatchBoard` 迴圈改為:

```go
	ch, unsub := h.hub.Subscribe(dept)
	defer unsub()
	var ticker *time.Ticker
	var tickC <-chan time.Time
	if h.heartbeat > 0 {
		ticker = time.NewTicker(h.heartbeat)
		tickC = ticker.C
		defer ticker.Stop()
	}
	lastSend := time.Now()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(boardEventToProto(ev)); err != nil {
				return nil
			}
			lastSend = time.Now()
		case <-tickC:
			// 期間無業務事件送出才發 heartbeat(細部 5.2.3 步驟 1)
			if time.Since(lastSend) < h.heartbeat {
				continue
			}
			err := stream.Send(&protov1.BoardEvent{Kind: protov1.BoardEventKind_BOARD_EVENT_KIND_HEARTBEAT})
			if err != nil {
				return nil // heartbeat 寫入失敗視同客戶端斷線,走清理流程(步驟 2)
			}
			lastSend = time.Now()
		}
	}
```

`backend/config/api.go` 加欄位與預設:

```go
BoardHeartbeatSeconds     int `envconfig:"BOARD_HEARTBEAT_SECONDS"`      // 預設 25(細部 5.2.3)
IngressIdleTimeoutSeconds int `envconfig:"INGRESS_IDLE_TIMEOUT_SECONDS"` // 預設 30,僅供啟動檢查
```

`Load()` 內補預設與防誤配警告(細部 5.2.3 錯誤處理):

```go
	if cfg.BoardHeartbeatSeconds == 0 {
		cfg.BoardHeartbeatSeconds = 25
	}
	if cfg.IngressIdleTimeoutSeconds == 0 {
		cfg.IngressIdleTimeoutSeconds = 30
	}
	if cfg.BoardHeartbeatSeconds >= cfg.IngressIdleTimeoutSeconds {
		log.Printf("WARNING: BOARD_HEARTBEAT_SECONDS(%d) >= INGRESS_IDLE_TIMEOUT_SECONDS(%d), WatchBoard 串流可能被 ingress idle timeout 切斷",
			cfg.BoardHeartbeatSeconds, cfg.IngressIdleTimeoutSeconds)
	}
```

`InitDomains()` 組裝點更新:

```go
	boardHub := salesorders.NewBoardHub(rdb, log.Default())
	dispatchUC := salesorders.NewDispatchUsecase(db, boardHub, dispatchNotifier) // notifier 由 07-plan adapter 提供,未接前傳 nil
	dispatchUC.SetEnforcer(enf)
	dispatchHandler := salesorders.NewDispatchHandler(dispatchUC, boardHub,
		time.Duration(cfg.BoardHeartbeatSeconds)*time.Second)
```

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go test -race ./internal/domain/salesorders/ ./internal/testutil/ -v`
Expected: PASS — 跨 replica 送達、部門隔離、rollback/空批次無事件、Valkey 故障 mutation 照常、閒置串流收 heartbeat 且 heartbeat 不進 Valkey/事件表、退訂停 relay;Task 1-4 全部測試不迴歸。

```bash
git add backend/internal/testutil/valkey.go backend/internal/domain/salesorders backend/config/api.go backend/internal/server
git commit -m "feat(backend): Valkey pub/sub 跨 replica 轉發與串流 heartbeat(5.2.2-5.2.3)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:細部文件 7 子功能 → Task 對應:5.1.1→T1(樂觀鎖、重排、拖回未指派、跨部門車次、併發);5.1.2→T2(批次候選、逐筆交易、共用 dispatched_at、部分失敗、空批次);5.1.3→T3(角色門檻、原因必填、清派車欄位退 pending、保留看板位置、重印警告兩段、事件+稽核同事務);5.1.4→T3(提交後逐筆觸發、DispatchNotifier 介面注入、fire-and-record);5.2.1→T4(串流 handler、部門訂閱、既有 auth、無 ticket、斷線清理);5.2.2→T5(提交後發佈、部門 channel、跨 replica、rollback 無事件、Valkey 故障降級、metrics);5.2.3→T5(25s heartbeat 可組態、不進 Valkey/事件表、ingress 誤配警告、前端降級僅註記)。細部「整合測試重點」六條全數落於 T1(樂觀鎖併發)、T2(批次部分失敗)、T3(取消事務)、T4/T5(串流生命週期)、T5(跨 replica、rollback 不發事件)。無缺漏。
- **已知佔位(皆標 TODO + 接手方)**:`noopDispatchNotifier`(→ 07-plan Task 6 `triggers.NewDispatchNotifier` 提供實作);Casbin `dispatch` 資源預設 policy(→ 01-plan Task 2 seeder);看板查詢 RPC(降級輪詢用,→ 05-plan `SalesOrderService.List`);前端重連/輪詢降級(→ 原計畫 Task 5.6,後端不實作);`audit_logs` 斷言依賴 03-plan 的 `AuditLog` 實體、`print_logs` 依賴 09-plan 5.5.1。這些皆為跨 domain 依賴,非本計畫範圍缺口。
- **類型一致**:`BoardEvent` / `BoardEventKind` / `BoardPublisher` / `DispatchNotifier` / `DispatchNotification`(T1 定義,T3/T4/T5 引用,名稱逐字一致);`NewDispatchUsecase(db, pub, notifier)` 三參數自 T1 固定;`NewBoardHub` 於 T4→T5 簽名變更一次(T5 Step 2 明列 T4 測試同步更新);`NewDispatchHandler` 於 T4/T5 各擴充一次(`InitDomains()` 組裝點同步);`CanTransition` / `Transition` 簽名逐字引用 05 細部文件 4.1.3;`rls.Identity` / `rls.FromContext` / `rls.NewContext` 引用 01-plan。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-17-backend-08-dispatch-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `detail/08-dispatch.md`、原計畫 v2.9.0、規格書 v1.0.34。*
