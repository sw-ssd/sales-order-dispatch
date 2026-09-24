//go:build integration

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// TestIntegrationLogisticsCrossDeptAndSelfIsolation D32 整合探針(真 PG + app_rw + 真 RLS +
// 記憶體 OpenFGA 引擎,生產同構接線):
//
//  1. schema 面:app_rw 對三表有完整 CRUD + 序列;00046 policy 存在、00047 ENABLE+FORCE。
//  2. 本人隔離:兩部門各建司機與指派 → 司機 A 只見自己那筆、司機 B 只見自己那筆。
//  3. 跨部門 RLS:scope=dept A 直接 SQL 讀 dept B 的 logistics_deliveries → 0 列;
//     寫 dept B → WITH CHECK 42501(負向控制,少了 RLS 接線不會紅)。
//  4. instance 級(10.8):dept B 的管理者對 dept A 的 delivery Check → deny
//     (manager 邊只存在於自己部門的 tuple)。
func TestIntegrationLogisticsCrossDeptAndSelfIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	ctx := context.Background()

	// ① schema 面。
	adminSql, adminDB := openPGEntClientFromGoose(t, adminDSN)
	for _, q := range []struct {
		name string
		sql  string
	}{
		{"app_rw CRUD", `SELECT has_table_privilege('app_rw', 'logistics_deliveries', 'SELECT')
			AND has_table_privilege('app_rw', 'logistics_deliveries', 'INSERT')
			AND has_table_privilege('app_rw', 'logistics_deliveries', 'UPDATE')
			AND has_table_privilege('app_rw', 'logistics_deliveries', 'DELETE')`},
		{"app_rw 序列", `SELECT has_sequence_privilege('app_rw', 'logistics_deliveries_id_seq', 'USAGE')
			AND has_sequence_privilege('app_rw', 'logistics_drivers_id_seq', 'USAGE')
			AND has_sequence_privilege('app_rw', 'vehicles_id_seq', 'USAGE')`},
		{"ENABLE", `SELECT relrowsecurity FROM pg_class WHERE relname = 'logistics_deliveries'`},
		{"FORCE", `SELECT relforcerowsecurity FROM pg_class WHERE relname = 'logistics_deliveries'`},
		{"policy 存在", `SELECT count(*) = 3 FROM pg_policies
			WHERE tablename IN ('logistics_drivers', 'vehicles', 'logistics_deliveries')
			AND policyname LIKE 'core_%'`},
	} {
		var ok bool
		if err := adminSql.QueryRow(q.sql).Scan(&ok); err != nil {
			t.Fatalf("schema 檢查 %s: %v", q.name, err)
		}
		if !ok {
			t.Fatalf("schema 檢查失敗: %s", q.name)
		}
	}

	// 種子:同公司兩部門 + 各自管理者/司機/車次(admin superuser 寫入)。
	co := adminDB.Company.Create().SetName("Logistics 探針公司").SetIdentifier("FLP-1").
		SetStatus("active").SaveX(ctx)
	deptA := adminDB.Department.Create().SetName("甲部").SetCompanyID(co.ID).SaveX(ctx)
	deptB := adminDB.Department.Create().SetName("乙部").SetCompanyID(co.ID).SaveX(ctx)
	mkUser := func(email, role string, did int) int {
		return adminDB.User.Create().SetEmail(email).SetName(email).SetRole(role).
			SetStatus("active").SetPasswordHash("x").
			SetCompanyID(co.ID).SetDepartmentID(did).SaveX(ctx).ID
	}
	mgrA := mkUser("flp-mgr-a@test", "dept_admin", deptA.ID)
	mgrB := mkUser("flp-mgr-b@test", "dept_admin", deptB.ID)
	drvA := mkUser("flp-drv-a@test", "staff", deptA.ID)
	drvB := mkUser("flp-drv-b@test", "staff", deptB.ID)
	routeA := adminDB.Route.Create().SetCode("RA").SetName("甲線").
		SetCompanyID(co.ID).SetDepartmentID(deptA.ID).SaveX(ctx).ID
	routeB := adminDB.Route.Create().SetCode("RB").SetName("乙線").
		SetCompanyID(co.ID).SetDepartmentID(deptB.ID).SaveX(ctx).ID

	// 生產同構:app_rw 連線 + dbtenant 攔截器 + 身分/scope/引擎注入(AGENTS §9-16)。
	appClient := openAppRoleEntClient(t, adminDSN)
	fgaClient, err := ofga.NewMemory(ctx, "logistics-rls-probe")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(func() { fgaClient.Close() })
	engine := authzopenfga.New(fgaClient)

	logisticsServer := func(id authz.Identity, scope auth.RLSScope) salesorderv1connect.LogisticsServiceClient {
		t.Helper()
		mux := http.NewServeMux()
		RegisterLogisticsService(mux, appClient)
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := authz.WithIdentity(r.Context(), id)
			c = auth.WithRLS(c, scope)
			c = authz.WithEngine(c, engine)
			c = authz.WithDB(c, appClient)
			mux.ServeHTTP(w, r.WithContext(c))
		})
		ts := httptest.NewServer(h)
		t.Cleanup(ts.Close)
		return salesorderv1connect.NewLogisticsServiceClient(http.DefaultClient, ts.URL)
	}
	idOf := func(uid, role string, did int) authz.Identity {
		return authz.Identity{
			UserID: uid, CompanyID: uItoa(co.ID), DepartmentID: uItoa(did),
			Role: role, Roles: auth.RolesFor(role),
		}
	}
	scopeDept := func(did int, uid string) auth.RLSScope {
		return auth.RLSScope{
			UserID: uid, CompanyID: uItoa(co.ID), DepartmentID: uItoa(did),
			DataScope: auth.DataScopeDepartment, CompanyActive: true,
		}
	}

	// ② 兩部門各自建檔 + 指派(真 RLS 寫入)。
	var deliveryA string
	{
		rpc := logisticsServer(idOf(uItoa(mgrA), "dept_admin", deptA.ID), scopeDept(deptA.ID, uItoa(mgrA)))
		d, err := rpc.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
			UserId: uItoa(drvA), Name: "甲司機",
		}))
		if err != nil {
			t.Fatalf("A 建司機: %v", err)
		}
		if _, err := rpc.CreateVehicle(ctx, connect.NewRequest(&v1.CreateVehicleRequest{
			PlateNo: "FLP-A1",
		})); err != nil {
			t.Fatalf("A 建車: %v", err)
		}
		asg, err := rpc.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
			RouteId: uItoa(routeA), DriverId: d.Msg.GetDriver().GetId(), Version: "0",
		}))
		if err != nil {
			t.Fatalf("A 指派: %v", err)
		}
		deliveryA = asg.Msg.GetDelivery().GetId()
	}
	{
		rpc := logisticsServer(idOf(uItoa(mgrB), "dept_admin", deptB.ID), scopeDept(deptB.ID, uItoa(mgrB)))
		d, err := rpc.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
			UserId: uItoa(drvB), Name: "乙司機",
		}))
		if err != nil {
			t.Fatalf("B 建司機: %v", err)
		}
		if _, err := rpc.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
			RouteId: uItoa(routeB), DriverId: d.Msg.GetDriver().GetId(), Version: "0",
		})); err != nil {
			t.Fatalf("B 指派: %v", err)
		}
	}

	// ②' 本人隔離:兩位司機各見自己 1 筆。
	myList := func(did int, uid int, want int) {
		t.Helper()
		rpc := logisticsServer(idOf(uItoa(uid), "staff", did), scopeDept(did, uItoa(uid)))
		resp, err := rpc.ListMyDeliveries(ctx, connect.NewRequest(&v1.ListMyDeliveriesRequest{}))
		if err != nil {
			t.Fatalf("ListMyDeliveries(uid=%d): %v", uid, err)
		}
		if resp.Msg.GetTotal() != int32(want) {
			t.Fatalf("uid=%d 應見 %d 筆,got %d", uid, want, resp.Msg.GetTotal())
		}
	}
	myList(deptA.ID, drvA, 1)
	myList(deptB.ID, drvB, 1)

	// ③ 跨部門 RLS:scope=A 讀/寫 B 的列 → 0 列 / 42501。
	rawApp := openAppRoleDB(t, adminDSN)
	// 帶 scope=A 的交易:B 部門的 delivery 不可見。
	tx, err := rawApp.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for _, stmt := range []string{
		"SET LOCAL app.current_data_scope = 'department'",
		"SET LOCAL app.current_company_id = '" + uItoa(co.ID) + "'",
		"SET LOCAL app.current_department_id = '" + uItoa(deptA.ID) + "'",
	} {
		if _, err := tx.Exec(stmt); err != nil {
			t.Fatalf("set scope: %v", err)
		}
	}
	var bRows int
	if err := tx.QueryRow(`SELECT count(*) FROM logistics_deliveries WHERE department_id = $1`,
		deptB.ID).Scan(&bRows); err != nil {
		t.Fatalf("scope A 查 B: %v", err)
	}
	if bRows != 0 {
		t.Fatalf("scope A 不得見 B 部門列,got %d", bRows)
	}
	// 負向:scope A 寫 B → WITH CHECK 42501。
	if _, err := tx.Exec(
		`INSERT INTO logistics_deliveries (company_id, department_id, route_id, assigned_by)
		 VALUES ($1, $2, $3, $4)`, co.ID, deptB.ID, routeB, mgrA,
	); err == nil {
		t.Fatal("跨部門寫入應被 WITH CHECK 擋下(42501)")
	} else if !strings.Contains(err.Error(), "42501") &&
		!strings.Contains(err.Error(), "row-level security") {
		t.Fatalf("跨部門寫入應 42501,got: %v", err)
	}
	_ = tx.Rollback()

	// ④ instance 級:dept B 管理者對 dept A 的 delivery → deny(10.8)。
	ok, err := engine.Check(ctx, "user:"+uItoa(mgrB), "can_write", logisticsObj(mustAnnID(t, deliveryA)))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if ok {
		t.Fatal("dept B 管理者不得寫 dept A 的 delivery(無 manager 邊)")
	}
}
