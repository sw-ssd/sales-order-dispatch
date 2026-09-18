package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// newUserTestServer 建立自建 db(empty) 的 UserService server 並注入身分。
func newUserTestServer(t *testing.T, id authz.Identity) (salesorderv1connect.UserServiceClient, *ent.Client) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	return newUserTestServerWithDB(t, id, db), db
}

// newUserTestServerWithDB 以已建立的 db 建立 UserService server 並注入身分。
func newUserTestServerWithDB(t *testing.T, id authz.Identity, db *ent.Client) salesorderv1connect.UserServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterUserServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL)
}

// seedUserCompany 建立公司與兩部門,回傳 (companyID, deptA, deptB)。
func seedUserCompany(t *testing.T, db *ent.Client) (int, int, int) {
	t.Helper()
	ctx := context.Background()
	co, err := db.Company.Create().SetName("公司A").SetIdentifier("co-a").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	da, err := db.Department.Create().SetCompanyID(co.ID).SetName("部門甲").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	dbb, err := db.Department.Create().SetCompanyID(co.ID).SetName("部門乙").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	return co.ID, da.ID, dbb.ID
}

func uItoa(i int) string { return strconv.Itoa(i) }

func TestAssignRoleBumpsTokenVersionAndAudit(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	// 目標使用者(guest, pending)。
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("g@t.com").SetName("guest").SetRole("guest").SetStatus("pending").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// 操作者 company_admin(同公司)。
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	resp, err := client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{
		UserId: uItoa(target.ID), Role: "staff", DepartmentId: uItoa(deptA),
	}))
	if err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
	if resp.Msg.GetUser().GetStatus() != "active" {
		t.Errorf("期望 active,得到 %s", resp.Msg.GetUser().GetStatus())
	}
	if resp.Msg.GetUser().GetRole() != "staff" {
		t.Errorf("期望 role staff,得到 %s", resp.Msg.GetUser().GetRole())
	}

	// token_version 遞增(由 0 → 1)。
	fresh, err := db.User.Get(ctx, target.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if fresh.TokenVersion != 1 {
		t.Errorf("期望 token_version=1,得到 %d", fresh.TokenVersion)
	}

	// 稽核存在(action=role_change)。
	audits, err := db.AuditLog.Query().All(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	found := false
	for _, a := range audits {
		if a.Action == "role_change" && a.ResourceID == uItoa(target.ID) {
			found = true
		}
	}
	if !found {
		t.Errorf("期望存在 role_change 稽核紀錄,得到 %d 筆", len(audits))
	}
}

func TestAssignRoleGuestDeniedNoManager(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("g2@t.com").SetName("g").SetRole("guest").SetStatus("pending").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// staff 非管理者 → permission_denied。
	id := authz.Identity{UserID: "3", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "staff", Roles: []string{"staff", "customer"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "staff"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("期望 permission_denied,得到 %v", err)
	}
}

func TestDeactivateBumpsTokenVersionAndAudit(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetEmail("d@t.com").SetName("d").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	if _, err := client.Deactivate(ctx, connect.NewRequest(&v1.DeactivateRequest{UserId: uItoa(target.ID)})); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	fresh, err := db.User.Get(ctx, target.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if fresh.Status != "inactive" {
		t.Errorf("期望 inactive,得到 %s", fresh.Status)
	}
	if fresh.TokenVersion != 1 {
		t.Errorf("期望 token_version=1,得到 %d", fresh.TokenVersion)
	}
	// 稽核存在。
	n, err := db.AuditLog.Query().Count(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if n < 1 {
		t.Errorf("期望稽核紀錄,得到 %d", n)
	}
}

func TestForceLogoutBumpsAndBlocksSelf(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetEmail("f@t.com").SetName("f").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	if _, err := client.ForceLogout(ctx, connect.NewRequest(&v1.ForceLogoutRequest{UserId: uItoa(target.ID)})); err != nil {
		t.Fatalf("ForceLogout: %v", err)
	}
	fresh, _ := db.User.Get(ctx, target.ID)
	if fresh.TokenVersion != 1 {
		t.Errorf("期望 token_version=1,得到 %d", fresh.TokenVersion)
	}

	// 對自己 → invalid_argument。
	// 操作者 id=2,目標 id=2(無此列但身份成立)。
	_, err = client.ForceLogout(ctx, connect.NewRequest(&v1.ForceLogoutRequest{UserId: "2"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("對自己期望 invalid_argument,得到 %v", err)
	}
}

func TestDeactivateOutOfScopeDenied(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	// 部門乙的 staff。
	other, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("ob@t.com").SetName("ob").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// dept_admin(部門甲)想停用部門乙帳號。
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.Deactivate(ctx, connect.NewRequest(&v1.DeactivateRequest{UserId: uItoa(other.ID)}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("期望 permission_denied,得到 %v", err)
	}
}

func TestListUsersScope(t *testing.T) {
	ctx := context.Background()

	t.Run("未登入回 Unauthenticated", func(t *testing.T) {
		client, _ := newUserTestServer(t, authz.Identity{})
		_, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("期望 unauthenticated,得到 %v", err)
		}
	})

	t.Run("dept_admin 僅見自己部門", func(t *testing.T) {
		_, db := newUserTestServer(t, authz.Identity{})
		coID, deptA, deptB := seedUserCompany(t, db)
		// 部門甲一個 staff、部門乙一個 staff。
		if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("a1@t.com").SetName("甲一").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
			t.Fatalf("user: %v", err)
		}
		if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("b1@t.com").SetName("乙一").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
			t.Fatalf("user: %v", err)
		}
		// 以 dept_admin 身分重建 client(注入身分)。
		id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		client := newUserTestServerWithDB(t, id, db)

		resp, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{}))
		if err != nil {
			t.Fatalf("ListUsers: %v", err)
		}
		if len(resp.Msg.GetUsers()) != 1 {
			t.Fatalf("期望 dept_admin 見 1 人,得到 %d", len(resp.Msg.GetUsers()))
		}
		if got := resp.Msg.GetUsers()[0].GetDepartmentId(); got != uItoa(deptA) {
			t.Fatalf("dept_admin 看到跨部門使用者 dept=%s", got)
		}
	})
}

func TestUpdateUserScope(t *testing.T) {
	ctx := context.Background()
	client, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	staff, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("s@t.com").SetName("staff").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}

	t.Run("dept_admin 更新自己部門 staff 成功", func(t *testing.T) {
		id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		c := newUserTestServerWithDB(t, id, db)
		name := "改名"
		if _, err := c.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uItoa(staff.ID), Name: &name})); err != nil {
			t.Fatalf("UpdateUser: %v", err)
		}
	})

	t.Run("dept_admin 更新他部門使用者被拒", func(t *testing.T) {
		other, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("o@t.com").SetName("o").SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			t.Fatalf("user: %v", err)
		}
		id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		c := newUserTestServerWithDB(t, id, db)
		name := "x"
		_, err = c.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uItoa(other.ID), Name: &name}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("期望 permission_denied,得到 %v", err)
		}
	})

	t.Run("未登入更新被拒", func(t *testing.T) {
		name := "x"
		_, err := client.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uItoa(staff.ID), Name: &name}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("期望 unauthenticated,得到 %v", err)
		}
	})
}
