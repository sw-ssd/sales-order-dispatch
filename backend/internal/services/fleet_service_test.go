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
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// fleetFixture 為一組部門 + 司機/管理者 + 車次(sqlite;每次獨立記憶體庫)。
type fleetFixture struct {
	db        *ent.Client
	engine    *authzopenfga.Engine
	coID      int
	deptID    int
	managerID int // dept_admin(指派者)
	driver1ID int // staff,有 fleet_drivers 列
	driver2ID int
	routeID   int
}

// openFleetFixture 開獨立 db、起記憶體 OpenFGA 引擎、建基礎資料。
func openFleetFixture(t *testing.T) *fleetFixture {
	t.Helper()
	f := &fleetFixture{}
	db := enttest.Open(t, "sqlite3", "file:fleet"+fleetDBSeq()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	f.db = db

	client, err := ofga.NewMemory(context.Background(), "fleet-svc-test-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	f.engine = authzopenfga.New(client)

	ctx := context.Background()
	f.coID = db.Company.Create().SetName("Fleet 公司").SetIdentifier("FLT-1").SetStatus("active").SaveX(ctx).ID
	f.deptID = db.Department.Create().SetName("Fleet 部").SetCompanyID(f.coID).SaveX(ctx).ID
	mkUser := func(email, role string) int {
		return db.User.Create().SetEmail(email).SetName(email).SetRole(role).
			SetStatus("active").SetPasswordHash("x").
			SetCompanyID(f.coID).SetDepartmentID(f.deptID).SaveX(ctx).ID
	}
	f.managerID = mkUser("flt-manager@test", "dept_admin")
	f.driver1ID = mkUser("flt-d1@test", "staff")
	f.driver2ID = mkUser("flt-d2@test", "staff")
	f.routeID = db.Route.Create().SetCode("R1").SetName("甲路線").
		SetCompanyID(f.coID).SetDepartmentID(f.deptID).SaveX(ctx).ID
	return f
}

// fleetDBSeq 讓每案獨立記憶體庫(同 dsn 會共用)。
var fleetSeq int

func fleetDBSeq() string {
	fleetSeq++
	return strconv.Itoa(fleetSeq)
}

// newFleetServer 以指定身分 + 記憶體引擎掛真 handler(與生產同構:
// requestid/dbtenant 攔截器 + authz 身分/引擎注入)。
func (f *fleetFixture) newFleetServer(t *testing.T, id authz.Identity) salesorderv1connect.FleetServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterFleetService(mux, f.db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithEngine(ctx, f.engine)
		ctx = authz.WithDB(ctx, f.db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewFleetServiceClient(http.DefaultClient, ts.URL)
}

// fleetIdentity 造身分(Roles 走內建繼承展開;dept 帶上)。
func fleetIdentity(role string, uid, cid, did int) authz.Identity {
	return authz.Identity{
		UserID: uItoa(uid), CompanyID: uItoa(cid), DepartmentID: uItoa(did),
		Role: role, Roles: auth.RolesFor(role),
	}
}

// TestFleetCreateAndAssignTuplesSync D32 全鏈( sqlite 單測版):
// 建司機/車輛 → 指派車次 → tuple 經 AfterCommit 對帳 → 被指派司機
// 以 instance 級 Check 通過;重指派樂觀鎖 + 判決移交。
func TestFleetCreateAndAssignTuplesSync(t *testing.T) {
	ctx := context.Background()
	f := openFleetFixture(t)
	rpc := f.newFleetServer(t, fleetIdentity("dept_admin", f.managerID, f.coID, f.deptID))

	// ① 建司機(drivers.user_id 關聯既有使用者)。
	drvResp, err := rpc.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
		UserId: uItoa(f.driver1ID), Name: "司機一",
	}))
	if err != nil {
		t.Fatalf("CreateDriver: %v", err)
	}
	drvID := drvResp.Msg.GetDriver().GetId()
	// AfterCommit 已跑 → assignee tuple 在位。
	if ok, err := f.engine.Check(ctx, "user:"+uItoa(f.driver1ID), "assignee", "driver:"+drvID); err != nil || !ok {
		t.Fatalf("司機 assignee tuple 應已對帳,ok=%v err=%v", ok, err)
	}

	// ② 建車輛。
	vehResp, err := rpc.CreateVehicle(ctx, connect.NewRequest(&v1.CreateVehicleRequest{
		PlateNo: "ABC-1234",
	}))
	if err != nil {
		t.Fatalf("CreateVehicle: %v", err)
	}
	vehID := vehResp.Msg.GetVehicle().GetId()

	// ③ 指派車次(version 0 = 新建)。
	asg, err := rpc.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
		RouteId: uItoa(f.routeID), DriverId: drvID, VehicleId: vehID, Version: "0",
	}))
	if err != nil {
		t.Fatalf("AssignDelivery: %v", err)
	}
	delID := asg.Msg.GetDelivery().GetId()
	if asg.Msg.GetDelivery().GetVersion() != "1" {
		t.Fatalf("新建 version 應為 1,got %s", asg.Msg.GetDelivery().GetVersion())
	}
	// ④ instance 級:被指派司機可讀寫(10.8)。
	for _, rel := range []string{"can_read", "can_write"} {
		if ok, err := f.engine.Check(ctx, "user:"+uItoa(f.driver1ID), rel, "fleet_delivery:"+delID); err != nil || !ok {
			t.Fatalf("被指派司機 %s 應通過,ok=%v err=%v", rel, ok, err)
		}
	}
	// 無關者 deny。
	if ok, _ := f.engine.Check(ctx, "user:999", "can_read", "fleet_delivery:"+delID); ok {
		t.Fatal("無關者不得讀取")
	}

	// ⑤ 樂觀鎖:舊 version 重指派 → 拒絕。
	_, err = rpc.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
		RouteId: uItoa(f.routeID), DriverId: drvID, Version: "0",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("版本衝突應 invalid_argument(同 dispatch 慣例),got %v", err)
	}

	// ⑥ 重指派(正確 version) → 遞增,判決移交。
	drv2, err := rpc.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
		UserId: uItoa(f.driver2ID), Name: "司機二",
	}))
	if err != nil {
		t.Fatalf("CreateDriver(2): %v", err)
	}
	re, err := rpc.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
		RouteId: uItoa(f.routeID), DriverId: drv2.Msg.GetDriver().GetId(), Version: "1",
	}))
	if err != nil {
		t.Fatalf("重指派: %v", err)
	}
	if re.Msg.GetDelivery().GetVersion() != "2" {
		t.Fatalf("重指派後 version 應為 2,got %s", re.Msg.GetDelivery().GetVersion())
	}
	if ok, _ := f.engine.Check(ctx, "user:"+uItoa(f.driver1ID), "can_read", "fleet_delivery:"+delID); ok {
		t.Fatal("被替換的司機不得再讀(tuple 對帳已移交)")
	}
	if ok, _ := f.engine.Check(ctx, "user:"+uItoa(f.driver2ID), "can_read", "fleet_delivery:"+delID); !ok {
		t.Fatal("新司機應可讀")
	}
}

// TestFleetCreateGuards 建檔守衛:車牌重複 already_exists、跨公司參照 not_found、
// 非管理角色(staff 無 fleet write) permission_denied。
func TestFleetCreateGuards(t *testing.T) {
	ctx := context.Background()
	f := openFleetFixture(t)

	deptAdmin := f.newFleetServer(t, fleetIdentity("dept_admin", f.managerID, f.coID, f.deptID))
	if _, err := deptAdmin.CreateVehicle(ctx, connect.NewRequest(&v1.CreateVehicleRequest{
		PlateNo: "XYZ-0001",
	})); err != nil {
		t.Fatalf("首輛車: %v", err)
	}
	_, err := deptAdmin.CreateVehicle(ctx, connect.NewRequest(&v1.CreateVehicleRequest{
		PlateNo: "XYZ-0001",
	}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("重複車牌應 already_exists,got %v", err)
	}

	// 跨公司使用者參照 → not_found(不洩漏存在性)。
	otherCo := f.db.Company.Create().SetName("別家").SetIdentifier("FLT-2").SetStatus("active").SaveX(ctx)
	foriegnUser := f.db.User.Create().SetEmail("flt-foreign@test").SetName("外人").
		SetRole("staff").SetStatus("active").SetPasswordHash("x").
		SetCompanyID(otherCo.ID).SaveX(ctx)
	_, err = deptAdmin.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
		UserId: uItoa(foriegnUser.ID), Name: "外人",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("跨公司參照應 not_found,got %v", err)
	}

	// staff 無 fleet write(rolePolicy) → middleware 之外的 requireScope 拒絕。
	staff := f.newFleetServer(t, fleetIdentity("staff", f.driver1ID, f.coID, f.deptID))
	_, err = staff.CreateVehicle(ctx, connect.NewRequest(&v1.CreateVehicleRequest{PlateNo: "S-1"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff 不得建車,got %v", err)
	}
}

// TestFleetListMyDeliveriesGates 10.12/10.8 閘門:
// 非司機身分拒絕;司機只見自己被指派的;tuple 缺席的列被 instance 級過濾。
func TestFleetListMyDeliveriesGates(t *testing.T) {
	ctx := context.Background()
	f := openFleetFixture(t)
	deptAdmin := f.newFleetServer(t, fleetIdentity("dept_admin", f.managerID, f.coID, f.deptID))

	// 兩位司機、兩條車次、各指派一筆。
	route2 := f.db.Route.Create().SetCode("R2").SetName("乙路線").
		SetCompanyID(f.coID).SetDepartmentID(f.deptID).SaveX(ctx).ID
	var driverIDs []string
	for _, uid := range []int{f.driver1ID, f.driver2ID} {
		r, err := deptAdmin.CreateDriver(ctx, connect.NewRequest(&v1.CreateDriverRequest{
			UserId: uItoa(uid), Name: "司機",
		}))
		if err != nil {
			t.Fatalf("CreateDriver: %v", err)
		}
		driverIDs = append(driverIDs, r.Msg.GetDriver().GetId())
	}
	for i, rid := range []int{f.routeID, route2} {
		if _, err := deptAdmin.AssignDelivery(ctx, connect.NewRequest(&v1.AssignDeliveryRequest{
			RouteId: uItoa(rid), DriverId: driverIDs[i], Version: "0",
		})); err != nil {
			t.Fatalf("AssignDelivery: %v", err)
		}
	}

	// ① 非司機身分(未建 fleet_drivers 列) → permission_denied。
	stranger := f.newFleetServer(t, fleetIdentity("staff", f.managerID, f.coID, f.deptID))
	if _, err := stranger.ListMyDeliveries(ctx, connect.NewRequest(&v1.ListMyDeliveriesRequest{})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("非司機應 permission_denied,got %v", err)
	}

	// ② 司機一:只見自己那筆(共 1)。
	d1 := f.newFleetServer(t, fleetIdentity("staff", f.driver1ID, f.coID, f.deptID))
	list, err := d1.ListMyDeliveries(ctx, connect.NewRequest(&v1.ListMyDeliveriesRequest{}))
	if err != nil {
		t.Fatalf("ListMyDeliveries: %v", err)
	}
	if list.Msg.GetTotal() != 1 {
		t.Fatalf("司機一只應見 1 筆,got %d", list.Msg.GetTotal())
	}

	// ③ instance 級過濾:把司機一那筆的指派 tuple 抽掉 → 視同不可見(10.8 第二道閘門)。
	target := list.Msg.GetDeliveries()[0].GetId()
	drv := f.db.FleetDriver.Query().AllX(ctx)
	var driverOfTarget string
	for _, d := range drv {
		if d.UserID == f.driver1ID {
			driverOfTarget = driverObj(d.ID)
		}
	}
	if driverOfTarget == "" {
		t.Fatal("找不到司機一的 fleet_drivers 列")
	}
	if err := f.engine.DeleteTuple(ctx, driverOfTarget+"#assignee", "driver", "fleet_delivery:"+target); err != nil {
		t.Fatalf("DeleteTuple(指派): %v", err)
	}
	again, err := d1.ListMyDeliveries(ctx, connect.NewRequest(&v1.ListMyDeliveriesRequest{}))
	if err != nil {
		t.Fatalf("ListMyDeliveries(2): %v", err)
	}
	if again.Msg.GetTotal() != 0 {
		t.Fatalf("tuple 缺席的列應被過濾,got %d", again.Msg.GetTotal())
	}
}
