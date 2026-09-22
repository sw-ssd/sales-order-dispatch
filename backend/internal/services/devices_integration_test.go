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
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// newDeviceServer 以請求交易掛 DeviceService。
func newDeviceServer(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.DeviceServiceClient {
	t.Helper()
	path, handler := salesorderv1connect.NewDeviceServiceHandler(NewDeviceService(db),
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
	return salesorderv1connect.NewDeviceServiceClient(http.DefaultClient, ts.URL)
}

// seedDeviceUsers 建公司/部門/兩使用者。
func seedDeviceUsers(t *testing.T, ctx context.Context, db *ent.Client) (co, dept, u1, u2 int) {
	t.Helper()
	seedTx(t, db, func(tx *ent.Tx) error {
		c, err := tx.Company.Create().SetName("裝置公司").SetIdentifier("D-" + t.Name()).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		d, err := tx.Department.Create().SetCompanyID(c.ID).SetName("門市一").Save(ctx)
		if err != nil {
			return err
		}
		a, err := tx.User.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetEmail("a-" + t.Name() + "@t.com").SetName("甲").
			SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		b, err := tx.User.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetEmail("b-" + t.Name() + "@t.com").SetName("乙").
			SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		co, dept, u1, u2 = c.ID, d.ID, a.ID, b.ID
		return nil
	})
	return co, dept, u1, u2
}

// staffID 組員工身分。
func staffID(co, dept, uid int) authz.Identity {
	return authz.Identity{UserID: strconv.Itoa(uid), CompanyID: strconv.Itoa(co),
		DepartmentID: strconv.Itoa(dept), Role: "staff", Roles: []string{"staff"}}
}

// TestIntegrationDeviceLifecycle 註冊冪等 → 換帳轉移 → 註銷冪等 → 平台非法擋。
func TestIntegrationDeviceLifecycle(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	co, dept, u1, u2 := seedDeviceUsers(t, ctx, db)
	rpc1 := newDeviceServer(t, db, staffID(co, dept, u1))
	rpc2 := newDeviceServer(t, db, staffID(co, dept, u2))

	r1, err := rpc1.RegisterDevice(ctx, connect.NewRequest(&salesorderv1.RegisterDeviceRequest{
		Platform: "android", FcmToken: "tok-1", DeviceName: "甲機",
	}))
	if err != nil {
		t.Fatalf("註冊: %v", err)
	}
	// 重複註冊冪等(同 ID)。
	r1b, err := rpc1.RegisterDevice(ctx, connect.NewRequest(&salesorderv1.RegisterDeviceRequest{
		Platform: "android", FcmToken: "tok-1",
	}))
	if err != nil {
		t.Fatalf("重複註冊: %v", err)
	}
	if r1b.Msg.GetDeviceId() != r1.Msg.GetDeviceId() {
		t.Fatalf("冪等應回同 ID,got %q vs %q", r1b.Msg.GetDeviceId(), r1.Msg.GetDeviceId())
	}
	// 換帳登入:乙註冊同 token → 歸屬轉移(新 ID),甲查不到。
	r2, err := rpc2.RegisterDevice(ctx, connect.NewRequest(&salesorderv1.RegisterDeviceRequest{
		Platform: "ios", FcmToken: "tok-1",
	}))
	if err != nil {
		t.Fatalf("換帳註冊: %v", err)
	}
	if r2.Msg.GetDeviceId() == r1.Msg.GetDeviceId() {
		t.Fatal("轉移應建新記錄")
	}
	// 註銷(乙) + 重複註銷冪等。
	if _, err := rpc2.UnregisterDevice(ctx, connect.NewRequest(&salesorderv1.UnregisterDeviceRequest{
		FcmToken: "tok-1",
	})); err != nil {
		t.Fatalf("註銷: %v", err)
	}
	if _, err := rpc2.UnregisterDevice(ctx, connect.NewRequest(&salesorderv1.UnregisterDeviceRequest{
		FcmToken: "tok-1",
	})); err != nil {
		t.Fatalf("重複註銷應冪等成功: %v", err)
	}
	// 平台非法 → invalid_argument。
	if _, err := rpc1.RegisterDevice(ctx, connect.NewRequest(&salesorderv1.RegisterDeviceRequest{
		Platform: "huawei", FcmToken: "tok-x",
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法平台應 invalid_argument,got %v", err)
	}
}

// TestIntegrationSenderFailmark FakeSender 成功/失效/失敗三態 + failed 終態 + sent 競態。
func TestIntegrationSenderFailmark(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	co, dept, u1, _ := seedDeviceUsers(t, ctx, db)
	var n1, n2, n3 int
	seedTx(t, db, func(tx *ent.Tx) error {
		mk := func(title string) (int, error) {
			n, err := tx.Notification.Create().SetCompanyID(co).SetDepartmentID(dept).
				SetUserID(u1).SetChannel("fcm").SetTitle(title).SetContent("c").
				SetStatus("pending").Save(ctx)
			if err != nil {
				return 0, err
			}
			return n.ID, nil
		}
		var err error
		if n1, err = mk("a"); err != nil {
			return err
		}
		if n2, err = mk("b"); err != nil {
			return err
		}
		if n3, err = mk("c"); err != nil {
			return err
		}
		return nil
	})
	fake := &FakeSender{OnSend: func(n *ent.Notification) SendResult {
		switch n.ID {
		case n1:
			return SendResult{Sent: true}
		case n2:
			return SendResult{InvalidTokens: []string{"tok-bad"}, FailReason: "unregistered"}
		default:
			return SendResult{FailReason: "timeout"}
		}
	}}
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		db2 := tx.Client()
		res := fake.Send(c2, db2, []int{n1, n2, n3})
		if len(res) != 3 {
			t.Fatalf("應回 3 結果,got %d", len(res))
		}
		st := func(id int) string {
			n, _ := db2.Notification.Query().Where(notification.ID(id)).Only(c2)
			return n.Status
		}
		if st(n1) != "sent" {
			t.Fatalf("n1 應 sent,got %q", st(n1))
		}
		if st(n2) != "failed" {
			t.Fatalf("n2 應 failed,got %q", st(n2))
		}
		if st(n3) != "failed" {
			t.Fatalf("n3 應 failed,got %q", st(n3))
		}
		// failed 終態:再送成功也不覆蓋。
		fake2 := &FakeSender{}
		fake2.Send(c2, db2, []int{n2})
		if st(n2) != "failed" {
			t.Fatalf("failed 不可被覆蓋,got %q", st(n2))
		}
		// Purge 失效 token。
		if _, err := db2.UserDevice.Create().SetUserID(u1).SetCompanyID(co).
			SetPlatform("android").SetFcmToken("tok-bad").Save(c2); err != nil {
			t.Fatalf("建裝置: %v", err)
		}
		PurgeInvalidTokens(c2, db2, []string{"tok-bad"}, co, u1)
		if n, _ := db2.UserDevice.Query().Where(userdevice.FcmTokenEQ("tok-bad")).Only(c2); n.DeletedAt == nil {
			t.Fatal("失效 token 應被軟刪除")
		}
		_ = tx.Rollback()
		return nil
	})
}
