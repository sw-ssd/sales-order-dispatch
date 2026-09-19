package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/casl"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"

	"google.golang.org/protobuf/types/known/structpb"
)

// newRoleTestServer 建立 enttest sqlite client + 注入身分的 RoleService HTTP server。
// 以測試中介層模擬 server authzMiddleware(身分 + CASL 開關 + DB 注入 ctx)。
func newRoleTestServer(t *testing.T, id authz.Identity) (salesorderv1connect.RoleServiceClient, *ent.Client) {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })

	mux := http.NewServeMux()
	RegisterRoleServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	client := salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL)
	return client, db
}

// mustStruct 建立 structpb.Struct(map 必須 JSON 相容)。
func mustStruct(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("structpb.NewStruct: %v", err)
	}
	return s
}

// staffEvaluator 依 DB 內 staff 角色規則建立 CASL evaluator(模擬 authz 規則載入)。
func staffEvaluator(t *testing.T, db *ent.Client, ctx context.Context) *casl.Evaluator {
	t.Helper()
	rows, err := db.RolePermission.Query().
		Where(rolepermission.HasRoleWith(role.CodeEQ("staff"))).
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		t.Fatalf("載入 staff 規則: %v", err)
	}
	rules := make([]casl.Rule, 0, len(rows))
	for _, rp := range rows {
		conds, err := casl.ParseConditions(rp.Conditions)
		if err != nil {
			t.Fatalf("ParseConditions(%#v): %v", rp.Conditions, err)
		}
		rules = append(rules, casl.Rule{Action: rp.Action, Subject: rp.Resource, Conditions: conds, Inverted: rp.Inverted})
	}
	return casl.NewEvaluator(rules, casl.Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1"})
}

// TestRolePermissionCRUD 驗收 brief Step 4:更新權限後新規則生效(經 CASL evaluator 強制)。
func TestRolePermissionCRUD(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db := newRoleTestServer(t, super)

	db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)

	// ListRoles:可列出角色。
	list, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{}))
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(list.Msg.GetRoles()) != 1 || list.Msg.GetRoles()[0].GetCode() != "staff" {
		t.Fatalf("ListRoles 結果不符: %#v", list.Msg.GetRoles())
	}

	// GetRolePermissions:初始為空。
	got, err := client.GetRolePermissions(ctx, connect.NewRequest(&v1.GetRolePermissionsRequest{RoleId: list.Msg.GetRoles()[0].GetId()}))
	if err != nil {
		t.Fatalf("GetRolePermissions: %v", err)
	}
	if len(got.Msg.GetPermissions()) != 0 {
		t.Fatalf("初始權限應為空,got %#v", got.Msg.GetPermissions())
	}

	// UpdateRolePermissions:寫入帶條件與無條件規則。
	upd, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId: list.Msg.GetRoles()[0].GetId(),
		Permissions: []*v1.Permission{
			{Resource: "sales_order", Action: "read", SortOrder: 1},
			{Resource: "sales_order", Action: "cancel", Conditions: mustStruct(t, map[string]any{"status": "pending"}), SortOrder: 2},
		},
	}))
	if err != nil {
		t.Fatalf("UpdateRolePermissions: %v", err)
	}
	if len(upd.Msg.GetPermissions()) != 2 {
		t.Fatalf("更新後應有 2 筆權限,got %d", len(upd.Msg.GetPermissions()))
	}

	// 回讀:依 sort_order 升冪。
	got, err = client.GetRolePermissions(ctx, connect.NewRequest(&v1.GetRolePermissionsRequest{RoleId: list.Msg.GetRoles()[0].GetId()}))
	if err != nil {
		t.Fatalf("GetRolePermissions: %v", err)
	}
	perms := got.Msg.GetPermissions()
	if len(perms) != 2 || perms[0].GetAction() != "read" || perms[1].GetAction() != "cancel" {
		t.Fatalf("權限順序/內容不符: %#v", perms)
	}
	if perms[1].GetConditions().GetFields()["status"].GetStringValue() != "pending" {
		t.Fatalf("conditions 未回讀: %#v", perms[1].GetConditions())
	}

	// 新規則強制生效(CASL evaluator 由 DB 規則載入,與 authz.loadRules 同構)。
	e := staffEvaluator(t, db, ctx)
	if !e.Can("cancel", "sales_order", map[string]any{"status": "pending"}) {
		t.Error("cancel(pending) 應允許(新規則生效)")
	}
	if e.Can("cancel", "sales_order", map[string]any{"status": "processing"}) {
		t.Error("cancel(processing) 應拒絕(條件不命中)")
	}
	if !e.Can("read", "sales_order", nil) {
		t.Error("read 無條件規則應命中")
	}
	if e.Can("delete", "sales_order", nil) {
		t.Error("未授權動作應拒絕")
	}
}

// TestRolePermissionCompanyScope 驗收「company_admin 限自己公司」:
// 不可改內建角色;conditions.company_id 僅可為自身公司。
func TestRolePermissionCompanyScope(t *testing.T) {
	ctx := context.Background()
	admin := authz.Identity{UserID: "u2", CompanyID: "c1", Role: "company_admin", Roles: []string{"company_admin"}}

	t.Run("company_admin 不可修改內建角色", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		sys := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId:      strconvID(sys.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read"}},
		}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("company_admin 改內建角色應 PermissionDenied,got %v", err)
		}
	})

	t.Run("company_admin 條件不得引用其他公司", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(custom.ID),
			Permissions: []*v1.Permission{
				{Resource: "sales_order", Action: "read", Conditions: mustStruct(t, map[string]any{"company_id": "c2"})},
			},
		}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("引用其他公司應 PermissionDenied,got %v", err)
		}
	})

	t.Run("company_admin 可管理自訂角色且限自己公司", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(custom.ID),
			Permissions: []*v1.Permission{
				{Resource: "sales_order", Action: "read", Conditions: mustStruct(t, map[string]any{"company_id": "c1"}), SortOrder: 1},
				{Resource: "customer", Action: "read", SortOrder: 2},
			},
		}))
		if err != nil {
			t.Fatalf("company_admin 改自訂角色應允許,got %v", err)
		}
	})
}

// TestRolePermissionDenied 驗收權限門檻:無身分 / 無管理權角色一律拒絕。
func TestRolePermissionDenied(t *testing.T) {
	ctx := context.Background()

	t.Run("未登入 Unauthenticated", func(t *testing.T) {
		client, _ := newRoleTestServer(t, authz.Identity{})
		_, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("未登入應 Unauthenticated,got %v", err)
		}
	})

	for _, id := range []authz.Identity{
		{UserID: "u3", CompanyID: "c1", Role: "staff", Roles: []string{"staff"}},
		{UserID: "u4", CompanyID: "c1", Role: "guest", Roles: []string{"guest"}},
		{UserID: "u5", CompanyID: "c1", Role: "dept_admin", Roles: []string{"dept_admin"}},
	} {
		client, _ := newRoleTestServer(t, id)
		_, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("%s ListRoles 應 PermissionDenied,got %v", id.Role, err)
		}
	}
}

// TestUpdateRolePermissionsValidation 輸入驗證:空欄位 / 重複規則 / 不存在角色。
func TestUpdateRolePermissionsValidation(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db := newRoleTestServer(t, super)
	custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SaveX(ctx)
	roleID := strconvID(custom.ID)

	_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId:      roleID,
		Permissions: []*v1.Permission{{Resource: " ", Action: "read"}},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("空 resource 應 InvalidArgument,got %v", err)
	}

	_, err = client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId: roleID,
		Permissions: []*v1.Permission{
			{Resource: "sales_order", Action: "read"},
			{Resource: "sales_order", Action: "read"},
		},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("重複規則應 InvalidArgument,got %v", err)
	}

	_, err = client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId:      "99999",
		Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read"}},
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("不存在角色應 NotFound,got %v", err)
	}
}

// TestListRolesPagination 複審 Minor 4:分頁 meta(Total/PageSize)與跨頁切分正確(走 pageList)。
func TestListRolesPagination(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db := newRoleTestServer(t, super)

	for i, n := range []string{"門市人員", "區域經理", "總部稽核"} {
		db.Role.Create().SetCode("role" + strconvID(i+1)).SetName(n).SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
	}

	p1, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{Page: 1, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if pg := p1.Msg.GetPagination(); pg.GetTotal() != 3 || pg.GetPageSize() != 2 || len(p1.Msg.GetRoles()) != 2 {
		t.Fatalf("第1頁分頁應 total=3 size=2 roles=2,got total=%d size=%d roles=%d", pg.GetTotal(), pg.GetPageSize(), len(p1.Msg.GetRoles()))
	}
	p2, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{Page: 2, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListRoles p2: %v", err)
	}
	if len(p2.Msg.GetRoles()) != 1 {
		t.Fatalf("第2頁應剩 1 筆,got %d", len(p2.Msg.GetRoles()))
	}
}

// seedRoleSortFixtures 建立三個角色(插入順序 丙→甲→乙,id 依序 1/2/3):
// 名稱碼位升冪(丙乙甲)、code 升冪(甲乙丙)與 id 升冪(丙甲乙)互不相同,
// 確保各欄位排序確實生效(值全同則無法區分)。
func seedRoleSortFixtures(t *testing.T, ctx context.Context, db *ent.Client) {
	t.Helper()
	for _, f := range []struct{ code, name string }{
		{"role-c", "丙角色"},
		{"role-a", "甲角色"},
		{"role-b", "乙角色"},
	} {
		db.Role.Create().SetCode(f.code).SetName(f.name).SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
	}
}

// listRoleNames 以指定排序取回角色名稱序列(sort 空 = 服務預設排序)。
func listRoleNames(t *testing.T, ctx context.Context, client salesorderv1connect.RoleServiceClient, sort string, desc bool) []string {
	t.Helper()
	res, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{
		Page: 1, PageSize: 10, Sort: sort, Desc: desc,
	}))
	if err != nil {
		t.Fatalf("ListRoles(sort=%q desc=%v): %v", sort, desc, err)
	}
	names := make([]string, 0, len(res.Msg.GetRoles()))
	for _, r := range res.Msg.GetRoles() {
		names = append(names, r.GetName())
	}
	return names
}

// TestListRolesSort D1/P3:角色清單 sort 白名單(code/name/id)與 desc 方向。
// **預設(sort 空)為 id 升冪**——與公司/部門的 id 降冪不同,故以兩個案例釘住
// (空、空+desc=true)且期望值與 id 升冪一致;其餘欄位預設升冪,desc=true 轉降冪;
// 非法值 → InvalidArgument 並列出白名單。
func TestListRolesSort(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db := newRoleTestServer(t, super)
	seedRoleSortFixtures(t, ctx, db)

	for _, tc := range []struct {
		name string
		sort string
		desc bool
		want []string
	}{
		{"sort 空 → 預設 id 升冪(現行行為)", "", false, []string{"丙角色", "甲角色", "乙角色"}},
		{"sort 空且 desc=true → 仍為預設 id 升冪(忽略 desc)", "", true, []string{"丙角色", "甲角色", "乙角色"}},
		{"sort=code → code 升冪", "code", false, []string{"甲角色", "乙角色", "丙角色"}},
		{"sort=code+desc → code 降冪", "code", true, []string{"丙角色", "乙角色", "甲角色"}},
		{"sort=name → 名稱升冪(sqlite 依碼位:丙 U+4E19 < 乙 U+4E59 < 甲 U+7532)", "name", false, []string{"丙角色", "乙角色", "甲角色"}},
		{"sort=name+desc → 名稱降冪", "name", true, []string{"甲角色", "乙角色", "丙角色"}},
		{"sort=id → id 升冪", "id", false, []string{"丙角色", "甲角色", "乙角色"}},
		{"sort=id+desc → id 降冪", "id", true, []string{"乙角色", "甲角色", "丙角色"}},
	} {
		if got := listRoleNames(t, ctx, client, tc.sort, tc.desc); !slices.Equal(got, tc.want) {
			t.Errorf("%s:got %v,want %v", tc.name, got, tc.want)
		}
	}

	// 非法值 → InvalidArgument,且訊息須列出白名單欄位。
	_, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{Sort: "bogus"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("非法 sort 應回 InvalidArgument,got %v", err)
	}
	for _, w := range []string{"code", "name", "id"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("錯誤訊息 %q 應列出白名單欄位 %q", err.Error(), w)
		}
	}
}

func strconvID(id int) string {
	return fmt.Sprintf("%d", id)
}
