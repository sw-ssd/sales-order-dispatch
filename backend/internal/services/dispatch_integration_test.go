//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// fakeBoardPublisher 計次發佈器(斷言提交後發佈)。
type fakeBoardPublisher struct {
	events []BoardEvent
}

func (f *fakeBoardPublisher) Publish(_ context.Context, _ int, ev BoardEvent) {
	f.events = append(f.events, ev)
}

// newDispatchServer 以請求交易掛 DispatchService。
func newDispatchServer(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.DispatchServiceClient {
	t.Helper()
	path, handler := salesorderv1connect.NewDispatchServiceHandler(NewDispatchService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = auth.WithRLS(ctx, auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewDispatchServiceClient(http.DefaultClient, ts.URL)
}

// seedDispatchBoard 建公司/部門/車次Ax2/操作員 + 3 張 pending(同車同日 seq 1,2,3)。
type dispatchIDs struct {
	co, dept, actor, routeA, routeB int
	orders                          []int
	version0                        int
}

func seedDispatchBoard(t *testing.T, ctx context.Context, db *ent.Client) dispatchIDs {
	t.Helper()
	var v dispatchIDs
	coID, deptID, custID, actorID, _ := seedOrderCompany(t, ctx, db)
	v.co, v.dept, v.actor = coID, deptID, actorID
	seedTx(t, db, func(tx *ent.Tx) error {
		ra, err := tx.Route.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetCode("A").SetName("甲線").Save(ctx)
		if err != nil {
			return err
		}
		rb, err := tx.Route.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetCode("B").SetName("乙線").Save(ctx)
		if err != nil {
			return err
		}
		v.routeA, v.routeB = ra.ID, rb.ID
		for i, seq := range []int{1, 2, 3} {
			o, err := tx.SalesOrder.Create().SetCompanyID(coID).SetDepartmentID(deptID).
				SetCustomerID(custID).SetOrderNo("W00000" + strconv.Itoa(i+1)).SetSource("W").
				SetRouteID(ra.ID).SetDeliverySequence(seq).
				SetExpectedDeliveryDate(dateOf(2026, 9, 22)).SetStatus("pending").
				SetCreatedBy(actorID).Save(ctx)
			if err != nil {
				return err
			}
			v.orders = append(v.orders, o.ID)
			if i == 0 {
				v.version0 = o.Version
			}
		}
		return nil
	})
	return v
}

// dispatchID 組 dept_admin 身分。
func dispatchID(v dispatchIDs) authz.Identity {
	return authz.Identity{UserID: strconv.Itoa(v.actor), CompanyID: strconv.Itoa(v.co),
		DepartmentID: strconv.Itoa(v.dept), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
}

// TestIntegrationDispatchAssign 指派 + 順位重排 + 樂觀鎖衝突 + 非 pending 拒絕。
func TestIntegrationDispatchAssign(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedDispatchBoard(t, ctx, db)
	fake := &fakeBoardPublisher{}
	SetBoardPublisher(fake)
	t.Cleanup(func() { SetBoardPublisher(noopPublisher{}) })
	rpc := newDispatchServer(t, db, dispatchID(v))

	// 拖第一筆至 seq 2:原 seq2,3 後移 → 序列應為 [_,2,3,4](首筆變 2,餘後移)。
	ar, err := rpc.AssignRoute(ctx, connect.NewRequest(&salesorderv1.AssignRouteRequest{
		SalesOrderId: strconv.Itoa(v.orders[0]), RouteId: strconv.Itoa(v.routeA),
		DeliverySequence: "2", Version: strconv.Itoa(v.version0),
		ExpectedDeliveryDate: "2026-09-22",
	}))
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if ar.Msg.GetDeliverySequence() != "2" {
		t.Fatalf("順位應為 2,got %q", ar.Msg.GetDeliverySequence())
	}
	if len(fake.events) != 1 || fake.events[0].Type != "route_assign" {
		t.Fatalf("應發佈 1 則 route_assign,got %+v", fake.events)
	}
	// 過期 version → 拒絕。
	if _, err := rpc.AssignRoute(ctx, connect.NewRequest(&salesorderv1.AssignRouteRequest{
		SalesOrderId: strconv.Itoa(v.orders[1]), RouteId: strconv.Itoa(v.routeA),
		DeliverySequence: "1", Version: "999",
		ExpectedDeliveryDate: "2026-09-22",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("過期 version 應 invalid_argument,got %v", err)
	}
	// 拖回未指派(route 清空)。
	un, err := rpc.AssignRoute(ctx, connect.NewRequest(&salesorderv1.AssignRouteRequest{
		SalesOrderId: strconv.Itoa(v.orders[2]), Version: currentVersion(t, ctx, db, v.orders[2]),
		ExpectedDeliveryDate: "2026-09-22",
	}))
	if err != nil {
		t.Fatalf("清空指派: %v", err)
	}
	if un.Msg.GetRouteId() != "" {
		t.Fatalf("清空後 route 應空,got %q", un.Msg.GetRouteId())
	}
}

// currentVersion 讀當前 version(系統範圍)。
func currentVersion(t *testing.T, ctx context.Context, db *ent.Client, oid int) string {
	t.Helper()
	var ver int
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		o, err := tx.Client().SalesOrder.Query().Where(salesorder.ID(oid)).Only(c2)
		if err != nil {
			t.Fatalf("讀訂單: %v", err)
		}
		ver = o.Version
		return nil
	})
	return strconv.Itoa(ver)
}

// TestIntegrationDispatchConfirmCancel 批次確認(部分失敗) + 取消(重印警告 + 保留位置)。
func TestIntegrationDispatchConfirmCancel(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedDispatchBoard(t, ctx, db)
	fake := &fakeBoardPublisher{}
	SetBoardPublisher(fake)
	t.Cleanup(func() { SetBoardPublisher(noopPublisher{}) })
	rpc := newDispatchServer(t, db, dispatchID(v))

	// 批次確認:3 筆全 pending → 全成功。
	cf, err := rpc.ConfirmDispatch(ctx, connect.NewRequest(&salesorderv1.ConfirmDispatchRequest{
		RouteId: strconv.Itoa(v.routeA), ExpectedDeliveryDate: "2026-09-22",
	}))
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if cf.Msg.GetSuccessCount() != 3 {
		t.Fatalf("應成功 3 筆,got %d", cf.Msg.GetSuccessCount())
	}
	// 再確認 → 無候選 → invalid_argument。
	if _, err := rpc.ConfirmDispatch(ctx, connect.NewRequest(&salesorderv1.ConfirmDispatchRequest{
		RouteId: strconv.Itoa(v.routeA), ExpectedDeliveryDate: "2026-09-22",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("無候選應 invalid_argument,got %v", err)
	}
	// 取消首筆(無列印 → 直接取消,位置保留)。
	cc, err := rpc.CancelDispatch(ctx, connect.NewRequest(&salesorderv1.CancelDispatchRequest{
		SalesOrderId: strconv.Itoa(v.orders[0]), Reason: "客戶改期",
	}))
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if cc.Msg.GetStatus() != "pending" || cc.Msg.GetRouteId() == "" {
		t.Fatalf("取消應退回 pending 且保留車次,got %q %q", cc.Msg.GetStatus(), cc.Msg.GetRouteId())
	}
	// 重印警告:手寫正式列印記錄後，未確認取消 → warning=true 且狀態不變。
	seedTx(t, db, func(tx *ent.Tx) error {
		fa, err := tx.FileAsset.Create().SetCompanyID(v.co).SetDepartmentID(v.dept).
			SetOwnerType("print_log").SetOwnerID(v.orders[1]).
			SetFilename("f.pdf").SetOriginalFilename("o.pdf").
			SetMimeType("application/pdf").SetSizeBytes(8).
			SetStoragePath("p").SetURL("u").Save(ctx)
		if err != nil {
			return err
		}
		_, err = tx.PrintLog.Create().SetCompanyID(v.co).SetDepartmentID(v.dept).
			SetDocumentType("dispatch_summary").SetRouteID(v.routeA).
			SetTargetDate(dateOf(2026, 9, 22)).SetPrintedBy(v.actor).SetFileAssetID(fa.ID).Save(ctx)
		return err
	})
	warn, err := rpc.CancelDispatch(ctx, connect.NewRequest(&salesorderv1.CancelDispatchRequest{
		SalesOrderId: strconv.Itoa(v.orders[1]), Reason: "客戶改期",
	}))
	if err != nil {
		t.Fatalf("重印檢查: %v", err)
	}
	if !warn.Msg.GetReprintWarning() {
		t.Fatal("已列印車次應回重印警告")
	}
	// 警告後未執行:仍 processing。
	if st := orderStatus(t, ctx, db, v.orders[1]); st != "processing" {
		t.Fatalf("未確認取消不得執行,got %q", st)
	}
	// 確認後取消照常執行。
	if _, err := rpc.CancelDispatch(ctx, connect.NewRequest(&salesorderv1.CancelDispatchRequest{
		SalesOrderId: strconv.Itoa(v.orders[1]), Reason: "客戶改期", AcknowledgeReprint: true,
	})); err != nil {
		t.Fatalf("確認後取消: %v", err)
	}
	if st := orderStatus(t, ctx, db, v.orders[1]); st != "pending" {
		t.Fatalf("確認後應退回 pending,got %q", st)
	}
	// staff 取消 → permission_denied。
	rpcStaff := newDispatchServer(t, db, authz.Identity{
		UserID: strconv.Itoa(v.actor), CompanyID: strconv.Itoa(v.co),
		DepartmentID: strconv.Itoa(v.dept), Role: "staff", Roles: []string{"staff"}})
	if _, err := rpcStaff.CancelDispatch(ctx, connect.NewRequest(&salesorderv1.CancelDispatchRequest{
		SalesOrderId: strconv.Itoa(v.orders[1]), Reason: "x",
	})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff 取消應 permission_denied,got %v", err)
	}
	// 無原因 → invalid_argument。
	if _, err := rpc.CancelDispatch(ctx, connect.NewRequest(&salesorderv1.CancelDispatchRequest{
		SalesOrderId: strconv.Itoa(v.orders[1]),
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("無原因應 invalid_argument,got %v", err)
	}
}

// orderStatus 讀訂單狀態(系統範圍)。
func orderStatus(t *testing.T, ctx context.Context, db *ent.Client, oid int) string {
	t.Helper()
	var st string
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		o, err := tx.Client().SalesOrder.Query().Where(salesorder.ID(oid)).Only(c2)
		if err != nil {
			t.Fatalf("讀訂單: %v", err)
		}
		st = o.Status
		return nil
	})
	return st
}
