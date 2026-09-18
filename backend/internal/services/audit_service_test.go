package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	auditv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/audit/v1/auditv1connect"
)

// newAuditTestServer 建立 AuditService client 並注入身分。
func newAuditTestServer(t *testing.T, id authz.Identity) (auditv1connect.AuditServiceClient, *ent.Client) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	mux := http.NewServeMux()
	RegisterAuditServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "127.0.0.1", UserAgent: "test-agent"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return auditv1connect.NewAuditServiceClient(http.DefaultClient, ts.URL), db
}

// seedAuditCompany 建立公司並回傳其 ID。
func seedAuditCompany(t *testing.T, db *ent.Client, ident string) int {
	t.Helper()
	co, err := db.Company.Create().SetName("公司-" + ident).SetIdentifier(ident).Save(context.Background())
	if err != nil {
		t.Fatalf("seed company: %v", err)
	}
	return co.ID
}

// seedAuditUser 建立使用者並回傳其 ID / 名稱。
func seedAuditUser(t *testing.T, db *ent.Client, companyID int, email, name string) (int, string) {
	t.Helper()
	u, err := db.User.Create().SetCompanyID(companyID).SetEmail(email).SetName(name).SetRole("staff").SetPasswordHash("x").Save(context.Background())
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u.ID, name
}

func seedAuditLog(t *testing.T, db *ent.Client, companyID, userID int, action auditlog.Action, resourceType, resourceID string, createdAt time.Time) int {
	t.Helper()
	a, err := db.AuditLog.Create().
		SetCompanyID(companyID).SetUserID(userID).
		SetAction(action).SetResourceType(resourceType).
		SetResourceID(resourceID).
		SetIPAddress("1.2.3.4").SetUserAgent("ua").
		SetBeforeSnapshot(map[string]any{"k": "v"}).
		SetCreatedAt(createdAt).
		Save(context.Background())
	if err != nil {
		t.Fatalf("seed auditlog: %v", err)
	}
	return a.ID
}

// TestAuditListSuperSeesAllAndCompanyFilter:super 可見全系統並可按 company 篩選。
func TestAuditListSuperSeesAllAndCompanyFilter(t *testing.T) {
	ctx := context.Background()
	_, db := newAuditTestServer(t, authz.Identity{})
	coA := seedAuditCompany(t, db, "co-a")
	coB := seedAuditCompany(t, db, "co-b")
	u1, _ := seedAuditUser(t, db, coA, "a@t.com", "甲")
	seedAuditLog(t, db, coA, u1, auditlog.ActionCreate, "customer", "1", time.Now())
	seedAuditLog(t, db, coB, u1, auditlog.ActionCreate, "customer", "2", time.Now())
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, _ := newAuditTestServer(t, id)
	resp, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{}))
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if resp.Msg.GetPagination().GetTotal() != 2 {
		t.Fatalf("super 應見全部 2 筆,得到 %d", resp.Msg.GetPagination().GetTotal())
	}
	// 以 company_id 篩選。
	respC, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{CompanyId: uItoa(coA)}))
	if err != nil {
		t.Fatalf("ListAuditLogs company: %v", err)
	}
	if respC.Msg.GetPagination().GetTotal() != 1 {
		t.Fatalf("filter company 應 1 筆,得到 %d", respC.Msg.GetPagination().GetTotal())
	}
}

// TestAuditListCompanyAdminOwnCompanyOnly:company_admin 僅見自己公司(忽略請求帶的他公司)。
func TestAuditListCompanyAdminOwnCompanyOnly(t *testing.T) {
	ctx := context.Background()
	_, db := newAuditTestServer(t, authz.Identity{})
	coA := seedAuditCompany(t, db, "co-a")
	coB := seedAuditCompany(t, db, "co-b")
	u1, _ := seedAuditUser(t, db, coA, "a@t.com", "甲")
	seedAuditLog(t, db, coA, u1, auditlog.ActionCreate, "customer", "1", time.Now())
	seedAuditLog(t, db, coB, u1, auditlog.ActionCreate, "customer", "2", time.Now())
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coA), Role: "company_admin", Roles: []string{"company_admin"}}
	client, _ := newAuditTestServer(t, id)
	// 試圖查他公司:應被強制限自己公司。
	resp, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{CompanyId: uItoa(coB)}))
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if resp.Msg.GetPagination().GetTotal() != 1 || resp.Msg.GetItems()[0].GetCompanyId() != uItoa(coA) {
		t.Fatalf("company_admin 應僅見自己公司 1 筆,得到 %d 筆 / %v", resp.Msg.GetPagination().GetTotal(), resp.Msg.GetItems()[0].GetCompanyId())
	}
}

// TestAuditListRoleDenied:staff 不可查稽核。
func TestAuditListRoleDenied(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "9", CompanyID: "10", Role: "staff", Roles: []string{"staff"}}
	client, _ := newAuditTestServer(t, id)
	_, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff 查稽核應回 permission_denied,得到 %v", err)
	}
}

// TestAuditListFiltersAndUserName:action/resource/user 篩選與操作者名稱 join。
func TestAuditListFiltersAndUserName(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, db := newAuditTestServer(t, id)
	coA := seedAuditCompany(t, db, "co-a")
	u1, n1 := seedAuditUser(t, db, coA, "a@t.com", "甲")
	seedAuditLog(t, db, coA, u1, auditlog.ActionCreate, "customer", "5", time.Now())
	seedAuditLog(t, db, coA, u1, auditlog.ActionUpdate, "customer", "5", time.Now())
	// action 篩選。
	resp, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Action: "update"}))
	if err != nil {
		t.Fatalf("ListAuditLogs action: %v", err)
	}
	if resp.Msg.GetPagination().GetTotal() != 1 || resp.Msg.GetItems()[0].GetAction() != "update" {
		t.Fatalf("action=update 應 1 筆,得到 %d", resp.Msg.GetPagination().GetTotal())
	}
	// resource_type + resource_id 定位單一資源軌跡。
	respR, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{ResourceType: "customer", ResourceId: "5"}))
	if err != nil {
		t.Fatalf("ListAuditLogs resource: %v", err)
	}
	if respR.Msg.GetPagination().GetTotal() != 2 {
		t.Fatalf("resource customer/5 應 2 筆,得到 %d", respR.Msg.GetPagination().GetTotal())
	}
	// 操作者名稱 join。
	if respR.Msg.GetItems()[0].GetUserName() != n1 {
		t.Fatalf("操作者名稱應為 %q,得到 %q", n1, respR.Msg.GetItems()[0].GetUserName())
	}
	// user 篩選。
	respU, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{UserId: uItoa(u1)}))
	if err != nil {
		t.Fatalf("ListAuditLogs user: %v", err)
	}
	if respU.Msg.GetPagination().GetTotal() != 2 {
		t.Fatalf("user 篩選應 2 筆,得到 %d", respU.Msg.GetPagination().GetTotal())
	}
}

// TestAuditListDefaultWindowExcludesOld:未帶時間篩選時套用近 3 個月,逾期的列不顯示。
func TestAuditListDefaultWindowExcludesOld(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, db := newAuditTestServer(t, id)
	coA := seedAuditCompany(t, db, "co-a")
	u1, _ := seedAuditUser(t, db, coA, "a@t.com", "甲")
	seedAuditLog(t, db, coA, u1, auditlog.ActionCreate, "customer", "recent", time.Now())
	seedAuditLog(t, db, coA, u1, auditlog.ActionCreate, "customer", "old", time.Now().AddDate(0, -4, 0)) // 4 個月前
	resp, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{}))
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if resp.Msg.GetPagination().GetTotal() != 1 {
		t.Fatalf("預設時間窗應僅含近 3 個月 1 筆,得到 %d", resp.Msg.GetPagination().GetTotal())
	}
	// 帶 from/to 顯示 4 個月前的列。
	respRange, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{
		From: time.Now().AddDate(0, -5, 0).Format(time.RFC3339),
		To:   time.Now().AddDate(0, -3, 0).Format(time.RFC3339),
	}))
	if err != nil {
		t.Fatalf("ListAuditLogs range: %v", err)
	}
	if respRange.Msg.GetPagination().GetTotal() != 1 {
		t.Fatalf("from/to 時間窗應含 1 筆,得到 %d", respRange.Msg.GetPagination().GetTotal())
	}
}

// TestAuditListInvalidFilterRejected:無效時間 / action / from>to 時序由應用層驗證。
func TestAuditListInvalidFilterRejected(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, _ := newAuditTestServer(t, id)
	if _, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{From: "not-a-time"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("from 非 RFC3339 應回 invalid_argument,得到 %v", err)
	}
	if _, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Action: "nope"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("action 非法應回 invalid_argument,得到 %v", err)
	}
}

// TestAuditListPagination:分頁正確。
func TestAuditListPagination(t *testing.T) {
	ctx := context.Background()
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client, db := newAuditTestServer(t, id)
	coA := seedAuditCompany(t, db, "co-a")
	u1, _ := seedAuditUser(t, db, coA, "a@t.com", "甲")
	for i := 0; i < 3; i++ {
		seedAuditLog(t, db, coA, u1, auditlog.ActionCreate, "customer", fmt.Sprintf("%d", i), time.Now())
	}
	resp, err := client.ListAuditLogs(ctx, connect.NewRequest(&auditv1.ListAuditLogsRequest{Page: 1, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if resp.Msg.GetPagination().GetTotal() != 3 || len(resp.Msg.GetItems()) != 2 {
		t.Fatalf("分頁應 total=3 items=2,得到 total=%d items=%d", resp.Msg.GetPagination().GetTotal(), len(resp.Msg.GetItems()))
	}
}
