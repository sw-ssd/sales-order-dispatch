//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// newOrderServerReal 以**生產同構**的攔截器鏈（requestid + dbtenant 請求交易）掛
// SalesOrderService —— 通知的提交後發送只有在真交易（commit/rollback 語意）下才走得對,
// 單元測試的 sqlite/enttest 看不到這條路徑的失敗。
func newOrderServerReal(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.SalesOrderServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterSalesOrderService(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = auth.WithRLS(ctx, auth.RLSScope{
			UserID: id.UserID, CompanyID: id.CompanyID, DepartmentID: id.DepartmentID,
			DataScope: auth.DataScopeDepartment, CompanyActive: true,
		})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewSalesOrderServiceClient(http.DefaultClient, ts.URL)
}

// TestIntegrationNotificationSendPersistsStatus 迴歸測試(2026-09-24,由 10.11 通知工作發現):
// 通知在**同一交易**建列(status=pending),發送排在提交後執行 —— 提交後的寫入必須用
// 「能寫的 client」,而不是已提交的請求交易 client(對它寫入會回 ErrTxDone,
// 而發送端忽略錯誤 → 通知永遠停在 pending,使用者看不到「已發送」也無法標已讀為 sent)。
//
// 這條用真 PG + 真攔截器鏈,因為 sqlite/enttest 的交易語意寬鬆、不會如實暴露。
func TestIntegrationNotificationSendPersistsStatus(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	coID, deptID, custID, actorID, prodID := seedOrderCompany(t, ctx, db)

	// 子帳號:通知的收件者。
	var subID int
	seedTx(t, db, func(tx *ent.Tx) error {
		sub, err := tx.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).
			SetEmail("send-" + t.Name() + "@t.com").SetName("子").
			SetRole("customer").SetPasswordHash("x").
			SetIsCustomer(true).SetCustomerID(custID).SetIsPrimary(false).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		subID = sub.ID
		return nil
	})

	SetTriggerSender(&FakeSender{})
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	adminID := authz.Identity{UserID: uItoa(actorID), CompanyID: uItoa(coID),
		DepartmentID: uItoa(deptID), Role: "dept_admin",
		Roles: auth.RolesFor("dept_admin")}
	rpc := newOrderServerReal(t, db, adminID)
	if _, err := rpc.CreateOrder(ctx, connect.NewRequest(&salesorderv1.CreateOrderRequest{
		CustomerId: uItoa(custID), Source: "W", ExpectedDeliveryDate: "2026-09-22",
		Items: []*salesorderv1.OrderItemInput{{
			ProductId: uItoa(prodID), DisplayName: "蘋果", Qty: "2", Unit: "斤",
		}},
	})); err != nil {
		t.Fatalf("業務下單: %v", err)
	}

	// 通知已建(pending 起手),且**提交後的發送確實落庫為 sent**。
	var total, sent int
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c := dbtenant.WithTenantTx(ctx, tx)
		rows, err := tx.Client().Notification.Query().Where().All(c)
		if err != nil {
			return err
		}
		for _, n := range rows {
			if n.UserID != subID {
				continue
			}
			total++
			if n.Status == "sent" {
				sent++
			}
		}
		return nil
	})
	if total == 0 {
		t.Fatalf("下單應為子帳號建立通知(in_app+fcm)")
	}
	if sent != total {
		t.Fatalf("提交後發送應把全部通知標為 sent: total=%d sent=%d(停在 pending = 提交後寫入用了已提交的交易 client)", total, sent)
	}
}
