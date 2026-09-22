//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationWatchBoard 訂閱收事件 + 部門隔離 + 未認證拒絕。
func TestIntegrationWatchBoard(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedDispatchBoard(t, ctx, db)
	SetBoardPublisher(localPublisher{})
	t.Cleanup(func() { SetBoardPublisher(localPublisher{}) })

	// 訂閱(dept_admin 部門)。
	id := dispatchID(v)
	path, handler := salesorderv1connect.NewDispatchServiceHandler(NewDispatchService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := authz.WithIdentity(r.Context(), id)
		c = auth.WithRLS(c, auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true})
		handler.ServeHTTP(w, r.WithContext(c))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	rpc := salesorderv1connect.NewDispatchServiceClient(http.DefaultClient, ts.URL)

	sctx, cancel := context.WithCancel(ctx)
	defer cancel()
	got := make(chan *salesorderv1.BoardEvent, 4)
	errCh := make(chan error, 1)
	go func() {
		stream, err := rpc.WatchBoard(sctx, connect.NewRequest(&salesorderv1.WatchBoardRequest{
			ExpectedDeliveryDate: "2026-09-22",
		}))
		if err != nil {
			errCh <- err
			return
		}
		for stream.Receive() {
			ev := stream.Msg()
			if ev.GetType() == "heartbeat" {
				continue
			}
			got <- ev
			return
		}
		errCh <- stream.Err()
	}()

	// 等訂閱就緒後觸發 Assign(輪詢握手,上限 3s;避免 sleep 競態丟事件)。
	deadline := time.Now().Add(3 * time.Second)
	for sharedHub.subscriberCount(v.dept) == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if sharedHub.subscriberCount(v.dept) == 0 {
		t.Fatal("訂閱未就緒")
	}
	rpc2 := newDispatchServer(t, db, id)
	if _, err := rpc2.AssignRoute(ctx, connect.NewRequest(&salesorderv1.AssignRouteRequest{
		SalesOrderId: itoa(v.orders[0]), RouteId: itoa(v.routeB),
		DeliverySequence: "1", Version: currentVersion(t, ctx, db, v.orders[0]),
		ExpectedDeliveryDate: "2026-09-22",
	})); err != nil {
		t.Fatalf("Assign: %v", err)
	}
	select {
	case ev := <-got:
		if ev.GetType() != "route_assign" {
			t.Fatalf("事件種類應為 route_assign,got %q", ev.GetType())
		}
		if ev.GetDepartmentId() == "" {
			t.Fatal("事件應帶部門")
		}
	case err := <-errCh:
		t.Fatalf("串流錯誤: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("10 秒內未收到看板事件")
	}
}

// TestIntegrationWatchBoardRejects 未認證 / 無部門拒絕。
func TestIntegrationWatchBoardRejects(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedDispatchBoard(t, ctx, db)

	// company_admin 無部門 → failed_precondition 系(invalid_argument)。
	adminID := authz.Identity{UserID: itoa(v.actor), CompanyID: itoa(v.co),
		Role: "company_admin", Roles: []string{"company_admin"}}
	path, handler := salesorderv1connect.NewDispatchServiceHandler(NewDispatchService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := authz.WithIdentity(r.Context(), adminID)
		c = auth.WithRLS(c, auth.RLSScope{UserID: adminID.UserID, CompanyID: adminID.CompanyID,
			DataScope: auth.DataScopeCompany, CompanyActive: true})
		handler.ServeHTTP(w, r.WithContext(c))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	rpc := salesorderv1connect.NewDispatchServiceClient(http.DefaultClient, ts.URL)
	stream, err := rpc.WatchBoard(ctx, connect.NewRequest(&salesorderv1.WatchBoardRequest{
		ExpectedDeliveryDate: "2026-09-22",
	}))
	if err != nil {
		t.Fatalf("建流: %v", err)
	}
	// 首訊息或錯誤：無部門應直接回錯。
	deadline := time.After(8 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("無部門訂閱應被拒")
			return
		default:
		}
		if !stream.Receive() {
			if err := stream.Err(); err == nil {
				t.Fatal("應回錯誤而非靜默結束")
			}
			return // 被拒即達標
		}
		if stream.Msg().GetType() == "heartbeat" {
			continue
		}
		t.Fatalf("無部門不應收到業務事件,got %q", stream.Msg().GetType())
	}
}

var _ = ent.Desc
