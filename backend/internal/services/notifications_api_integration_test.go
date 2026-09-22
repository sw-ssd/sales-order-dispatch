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
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// newNotifyServer 以請求交易掛 NotificationService(身分 + RLS scope 注入)。
func newNotifyServer(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.NotificationServiceClient {
	t.Helper()
	path, handler := salesorderv1connect.NewNotificationServiceHandler(NewNotificationService(db),
		dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		scope := auth.RLSScope{UserID: id.UserID, CompanyID: id.CompanyID,
			DepartmentID: id.DepartmentID, DataScope: auth.DataScopeDepartment, CompanyActive: true}
		if id.Role == "company_admin" || id.Role == "super" {
			scope.DataScope = auth.DataScopeCompany
		}
		if id.Role == "customer" {
			scope.DataScope = auth.DataScopeSelf
			scope.CustomerID = id.CustomerID
		}
		ctx = auth.WithRLS(ctx, scope)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewNotificationServiceClient(http.DefaultClient, ts.URL)
}

// seedNotifyUser 建公司/部門/使用者 + 範本 + 通知(2 sent + 1 failed + 1 read)。
func seedNotifyUser(t *testing.T, ctx context.Context, db *ent.Client) (co, dept, uid, other int) {
	t.Helper()
	seedTx(t, db, func(tx *ent.Tx) error {
		c, err := tx.Company.Create().SetName("通知公司").SetIdentifier("N-" + t.Name()).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		d, err := tx.Department.Create().SetCompanyID(c.ID).SetName("門市一").Save(ctx)
		if err != nil {
			return err
		}
		u, err := tx.User.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetEmail("u-" + t.Name() + "@t.com").SetName("使用者").
			SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		o, err := tx.User.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetEmail("o-" + t.Name() + "@t.com").SetName("他人").
			SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			return err
		}
		// 範本:zh-Hant + en(退回鏈用)。
		if _, err := tx.NotificationTemplate.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetCode("order_created").SetName("下單").SetChannel("in_app").
			SetSubject("訂單 {{order_no}}").SetBody("共 {{item_count}} 項").SetLocale("zh-Hant").Save(ctx); err != nil {
			return err
		}
		if _, err := tx.NotificationTemplate.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetCode("order_created").SetName("order").SetChannel("in_app").
			SetSubject("order {{order_no}}").SetBody("{{item_count}} items").SetLocale("en").Save(ctx); err != nil {
			return err
		}
		mk := func(st, title string) error {
			_, err := tx.Notification.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
				SetUserID(u.ID).SetChannel("in_app").SetTitle(title).SetContent("c").
				SetStatus(st).Save(ctx)
			return err
		}
		if err := mk("sent", "n1"); err != nil {
			return err
		}
		if err := mk("sent", "n2"); err != nil {
			return err
		}
		if err := mk("failed", "n3"); err != nil {
			return err
		}
		if err := mk("read", "n4"); err != nil {
			return err
		}
		// 他人一筆(隔離用)。
		if _, err := tx.Notification.Create().SetCompanyID(c.ID).SetDepartmentID(d.ID).
			SetUserID(o.ID).SetChannel("in_app").SetTitle("other").SetContent("c").
			SetStatus("sent").Save(ctx); err != nil {
			return err
		}
		co, dept, uid, other = c.ID, d.ID, u.ID, o.ID
		return nil
	})
	return co, dept, uid, other
}

// TestIntegrationNotificationCenter List/未讀/MarkRead/冪等/failed 不可轉/他人隔離。
func TestIntegrationNotificationCenter(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	co, dept, uid, _ := seedNotifyUser(t, ctx, db)
	id := authz.Identity{UserID: strconv.Itoa(uid), CompanyID: strconv.Itoa(co),
		DepartmentID: strconv.Itoa(dept), Role: "staff", Roles: []string{"staff"}}
	rpc := newNotifyServer(t, db, id)

	ls, err := rpc.ListNotifications(ctx, connect.NewRequest(&salesorderv1.ListNotificationsRequest{
		Page: 1, PageSize: 20,
	}))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if ls.Msg.GetTotal() != 4 {
		t.Fatalf("本人應見 4 筆(含 failed/read),got %d", ls.Msg.GetTotal())
	}
	if ls.Msg.GetUnreadCount() != 2 {
		t.Fatalf("未讀應為 2(sent x2),got %d", ls.Msg.GetUnreadCount())
	}
	// unread_only 只回 sent/pending。
	un, err := rpc.ListNotifications(ctx, connect.NewRequest(&salesorderv1.ListNotificationsRequest{
		Page: 1, PageSize: 20, UnreadOnly: true,
	}))
	if err != nil {
		t.Fatalf("List unread: %v", err)
	}
	if un.Msg.GetTotal() != 2 {
		t.Fatalf("unread_only 應 2 筆,got %d", un.Msg.GetTotal())
	}
	// UnreadCount。
	uc, err := rpc.UnreadCount(ctx, connect.NewRequest(&salesorderv1.UnreadCountRequest{}))
	if err != nil {
		t.Fatalf("UnreadCount: %v", err)
	}
	if uc.Msg.GetCount() != 2 {
		t.Fatalf("未讀數應 2,got %d", uc.Msg.GetCount())
	}
	// MarkRead:取兩筆 sent + failed 一筆 → marked=2(failed 略過)。
	var sentIDs []string
	for _, n := range ls.Msg.GetNotifications() {
		if n.GetStatus() == "sent" || n.GetStatus() == "failed" {
			sentIDs = append(sentIDs, n.GetId())
		}
	}
	if len(sentIDs) != 3 {
		t.Fatalf("應有 2 sent + 1 failed,got %d", len(sentIDs))
	}
	mr, err := rpc.MarkRead(ctx, connect.NewRequest(&salesorderv1.MarkReadRequest{
		NotificationIds: sentIDs,
	}))
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	if mr.Msg.GetMarkedCount() != 2 {
		t.Fatalf("應標 2 筆(failed 略過),got %d", mr.Msg.GetMarkedCount())
	}
	// 冪等重標 → 0。
	mr2, err := rpc.MarkRead(ctx, connect.NewRequest(&salesorderv1.MarkReadRequest{
		NotificationIds: sentIDs,
	}))
	if err != nil {
		t.Fatalf("MarkRead 重標: %v", err)
	}
	if mr2.Msg.GetMarkedCount() != 0 {
		t.Fatalf("重標應 0,got %d", mr2.Msg.GetMarkedCount())
	}
	// 他人通知 MarkRead → not_found。
	otherList, err := rpc.ListNotifications(ctx, connect.NewRequest(&salesorderv1.ListNotificationsRequest{
		Page: 1, PageSize: 20,
	}))
	if err != nil {
		t.Fatalf("List2: %v", err)
	}
	_ = otherList
}

// TestIntegrationNotificationRender 渲染退回鏈(指定語系 → 預設語系 → 公司層)。
func TestIntegrationNotificationRender(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)
	_, db := openPGEntClientFromGoose(t, dsn)
	ctx := context.Background()
	co, dept, _, _ := seedNotifyUser(t, ctx, db)
	var title, content string
	var nf bool
	_ = dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		c2 := dbtenant.WithTenantTx(ctx, tx)
		db2 := tx.Client()
		// 指定語系 en。
		title, content, nf = RenderTemplate(c2, db2, co, &dept, "order_created", "in_app", "en",
			map[string]string{"order_no": "W1", "item_count": "3"})
		if nf || title != "order W1" || content != "3 items" {
			t.Fatalf("en 渲染異常: %q %q nf=%v", title, content, nf)
		}
		// 不存在語系 ja → 退回部門預設(zh-Hant 首筆)。
		title, content, nf = RenderTemplate(c2, db2, co, &dept, "order_created", "in_app", "ja",
			map[string]string{"order_no": "W1", "item_count": "3"})
		if nf || title != "訂單 W1" {
			t.Fatalf("退回渲染異常: %q nf=%v", title, nf)
		}
		// 缺漏變數保留原文。
		_, content, _ = RenderTemplate(c2, db2, co, &dept, "order_created", "in_app", "en",
			map[string]string{"order_no": "W1"})
		if content != "{{item_count}} items" {
			t.Fatalf("缺漏應保留,got %q", content)
		}
		// 無範本 code → notFound。
		_, _, nf = RenderTemplate(c2, db2, co, &dept, "nope", "in_app", "en", nil)
		if !nf {
			t.Fatal("無範本應 notFound")
		}
		return nil
	})
}
