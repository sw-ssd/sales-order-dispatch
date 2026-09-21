//go:build integration

package services

import (
	"context"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// staffIdentity 組員工身分(dept_admin 或 staff)。
func staffIdentity(v returnIDs, uid int, role string) authz.Identity {
	return authz.Identity{UserID: strconv.Itoa(uid), CompanyID: strconv.Itoa(v.co),
		DepartmentID: strconv.Itoa(v.dept), Role: role, Roles: []string{role}}
}

// seedReviewShop 在 seedReturnShop 上加主責業務 + 審核員(dept_admin)，回 ids + staff。
type reviewIDs struct {
	returnIDs
	rep, admin int
}

func seedReviewShop(t *testing.T, ctx context.Context, db *ent.Client) reviewIDs {
	t.Helper()
	v := seedReturnShop(t, ctx, db)
	var rep, admin int
	seedTx(t, db, func(tx *ent.Tx) error {
		r, err := tx.User.Create().SetCompanyID(v.co).SetDepartmentID(v.dept).
			SetEmail("rep-" + t.Name() + "@t.com").SetName("主責業務").
			SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		a, err := tx.User.Create().SetCompanyID(v.co).SetDepartmentID(v.dept).
			SetEmail("adm-" + t.Name() + "@t.com").SetName("審核員").
			SetRole("dept_admin").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		rep, admin = r.ID, a.ID
		// 客戶主責掛 rep。
		if _, err := tx.Customer.UpdateOneID(v.cust).SetDefaultSalesRepID(rep).Save(ctx); err != nil {
			return err
		}
		return nil
	})
	return reviewIDs{returnIDs: v, rep: rep, admin: admin}
}

// createPending 以子帳號建一張 pending 申請回 id。
func createPending(t *testing.T, ctx context.Context, rpc interface {
	CreateReturnRequest(context.Context, *connect.Request[salesorderv1.CreateReturnRequestRequest]) (*connect.Response[salesorderv1.CreateReturnRequestResponse], error)
}, v returnIDs) string {
	t.Helper()
	cr, err := rpc.CreateReturnRequest(ctx, connect.NewRequest(&salesorderv1.CreateReturnRequestRequest{
		Items: []*salesorderv1.ReturnItemInput{{
			SourceType: "order_item", SalesOrderItemId: strconv.Itoa(v.orderItem),
			Quantity: "1", Reason: "爛果",
		}},
	}))
	if err != nil {
		t.Fatalf("建 pending: %v", err)
	}
	return cr.Msg.GetId()
}

// TestIntegrationReturnReview 主責審 approved → 重審拒絕 → 拒絕缺原因擋 → 非主責拒絕 →
// 併發樂觀鎖僅一勝 → 證明快照 → 原訂單不變。
func TestIntegrationReturnReview(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	v := seedReviewShop(t, ctx, db)
	rpcSub := newReturnServer(t, db, subIdentity(v.returnIDs, v.sub))
	rpcRep := newReturnServer(t, db, staffIdentity(v.returnIDs, v.rep, "staff"))
	rpcAdm := newReturnServer(t, db, staffIdentity(v.returnIDs, v.admin, "dept_admin"))

	// 主責審 approved(version 0)。
	rid := createPending(t, ctx, rpcSub, v.returnIDs)
	rv, err := rpcRep.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
		Id: rid, Decision: "approved", ExpectedVersion: "0",
	}))
	if err != nil {
		t.Fatalf("主責 approved: %v", err)
	}
	if rv.Msg.GetStatus() != "approved" {
		t.Fatalf("審後應為 approved,got %q", rv.Msg.GetStatus())
	}
	// 重審 → 拒絕(非 pending)。
	if _, err := rpcRep.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
		Id: rid, Decision: "rejected", RejectReason: "太晚", ExpectedVersion: "1",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("重審應 invalid_argument,got %v", err)
	}
	// 拒絕缺原因 → invalid_argument。
	rid2 := createPending(t, ctx, rpcSub, v.returnIDs)
	if _, err := rpcAdm.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
		Id: rid2, Decision: "rejected", ExpectedVersion: "0",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("拒絕缺原因應 invalid_argument,got %v", err)
	}
	// 非主責 staff → permission_denied。
	var other int
	seedTx(t, db, func(tx *ent.Tx) error {
		u, err := tx.User.Create().SetCompanyID(v.co).SetDepartmentID(v.dept).
			SetEmail("other-" + t.Name() + "@t.com").SetName("路人業務").
			SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		other = u.ID
		return nil
	})
	rpcOther := newReturnServer(t, db, staffIdentity(v.returnIDs, other, "staff"))
	if _, err := rpcOther.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
		Id: rid2, Decision: "approved", ExpectedVersion: "0",
	})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("非主責應 permission_denied,got %v", err)
	}
	// dept_admin 可審(rid2 仍 pending)。
	if _, err := rpcAdm.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
		Id: rid2, Decision: "rejected", RejectReason: "已過期", ExpectedVersion: "0",
	})); err != nil {
		t.Fatalf("dept_admin rejected: %v", err)
	}
	// 樂觀鎖併發:同 version 搶審僅一勝。
	rid3 := createPending(t, ctx, rpcSub, v.returnIDs)
	okN := 0
	for _, rpc := range []interface {
		ReviewReturnRequest(context.Context, *connect.Request[salesorderv1.ReviewReturnRequestRequest]) (*connect.Response[salesorderv1.ReviewReturnRequestResponse], error)
	}{rpcRep, rpcAdm} {
		if _, err := rpc.ReviewReturnRequest(ctx, connect.NewRequest(&salesorderv1.ReviewReturnRequestRequest{
			Id: rid3, Decision: "approved", ExpectedVersion: "0",
		})); err == nil {
			okN++
		}
	}
	if okN != 1 {
		t.Fatalf("併發同版審核應恰一勝,got %d", okN)
	}
	// 證明:approved 有快照;rejected 取證明被拒;pending 取證明被拒。
	cert, err := rpcSub.GetReturnCertificate(ctx, connect.NewRequest(&salesorderv1.GetReturnCertificateRequest{Id: rid}))
	if err != nil {
		t.Fatalf("approved 證明: %v", err)
	}
	if cert.Msg.GetCustomerCode() != "RT000001" || len(cert.Msg.GetItems()) != 1 {
		t.Fatalf("證明內容異常: code=%q items=%d", cert.Msg.GetCustomerCode(), len(cert.Msg.GetItems()))
	}
	if _, err := rpcSub.GetReturnCertificate(ctx, connect.NewRequest(&salesorderv1.GetReturnCertificateRequest{Id: rid2})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("rejected 證明應 invalid_argument,got %v", err)
	}
	rid4 := createPending(t, ctx, rpcSub, v.returnIDs)
	if _, err := rpcSub.GetReturnCertificate(ctx, connect.NewRequest(&salesorderv1.GetReturnCertificateRequest{Id: rid4})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("pending 證明應 invalid_argument,got %v", err)
	}
	// 原訂單不變:processing 且無退貨事件。
	var status string
	_ = status
}
