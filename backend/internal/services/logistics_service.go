// LogisticsService logistics 執行層首批(D32/10.1/10.4/10.12):建檔(司機/車輛)、
// 車次指派(route 粒度,version 樂觀鎖)、被指派司機的任務清單。
//
// 授權三層:rolePolicy logistics(middleware ability 閘門)→ 服務層 requireScope 與
// 部門範圍(deptScope/參照比對,RLS 兜底)→ OpenFGA instance 級(10.8:ListMyDeliveries
// 逐列 Check,被指派司機沿 driver#assignee 到達;tuple 經 AfterCommit 對帳,見 logistics_events.go)。
// 稽核:建檔/指派與業務寫入同交易(recordAuditBA,D18)。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/logisticsdelivery"
	"github.com/salesorder/sales-order-1.0/backend/ent/logisticsdriver"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/ent/vehicle"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// LogisticsService 實作 salesorder.v1.LogisticsService。
type LogisticsService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedLogisticsServiceHandler
}

// NewLogisticsService 建立 LogisticsService。
func NewLogisticsService(db *ent.Client) *LogisticsService {
	return &LogisticsService{db: db}
}

// RegisterLogisticsService 掛到 /api/v1(租戶 session + RLS,比照 SalesOrderService)。
func RegisterLogisticsService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewLogisticsServiceHandler(NewLogisticsService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// logisticsEngine 取每請求注入的 OpenFGA 引擎;缺席 = 接線失敗 → fail-closed(內部錯,
// 與 authorizeRPC「啟用卻無引擎」同語意)。刻意在交易內取:tuple 同步需要它,
// 提交前就失敗好過提交後靜默漏同步。
func logisticsEngine(ctx context.Context) (*authzopenfga.Engine, error) {
	e := authz.EngineFrom(ctx)
	if e == nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	return e, nil
}

// createAudit 為建檔類寫入補一筆稽核(create,D18)。
func createAudit(ctx context.Context, tx *ent.Tx, resource string, id, cid int, did *int, actor int, after map[string]any) error {
	return recordAuditBA(ctx, tx, resource, "create", id, cid, did, actor, nil, after)
}

func driverToProto(d *ent.LogisticsDriver) *salesorderv1.LogisticsDriver {
	out := &salesorderv1.LogisticsDriver{
		Id:            strconv.Itoa(d.ID),
		CompanyId:     strconv.Itoa(d.CompanyID),
		UserId:        strconv.Itoa(d.UserID),
		Name:          d.Name,
		Phone:         d.Phone,
		CurrentStatus: d.CurrentStatus,
	}
	if d.DepartmentID != nil {
		out.DepartmentId = strconv.Itoa(*d.DepartmentID)
	}
	return out
}

func vehicleToProto(v *ent.Vehicle) *salesorderv1.Vehicle {
	out := &salesorderv1.Vehicle{
		Id:          strconv.Itoa(v.ID),
		CompanyId:   strconv.Itoa(v.CompanyID),
		PlateNo:     v.PlateNo,
		VehicleType: v.VehicleType,
		Status:      v.Status,
	}
	if v.DepartmentID != nil {
		out.DepartmentId = strconv.Itoa(*v.DepartmentID)
	}
	return out
}

func logisticsDeliveryToProto(d *ent.LogisticsDelivery) *salesorderv1.LogisticsDelivery {
	out := &salesorderv1.LogisticsDelivery{
		Id:         strconv.Itoa(d.ID),
		CompanyId:  strconv.Itoa(d.CompanyID),
		RouteId:    strconv.Itoa(d.RouteID),
		AssignedBy: strconv.Itoa(d.AssignedBy),
		Status:     d.Status,
		Version:    strconv.Itoa(d.Version),
		CreatedAt:  d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  d.UpdatedAt.Format(time.RFC3339),
	}
	if d.DepartmentID != nil {
		out.DepartmentId = strconv.Itoa(*d.DepartmentID)
	}
	if d.DriverID != nil {
		out.DriverId = strconv.Itoa(*d.DriverID)
	}
	if d.VehicleID != nil {
		out.VehicleId = strconv.Itoa(*d.VehicleID)
	}
	return out
}

// CreateDriver 建司機(10.1:關聯既有 users,不另建帳號;同公司同人不重複)。
func (s *LogisticsService) CreateDriver(ctx context.Context, req *connect.Request[salesorderv1.CreateDriverRequest]) (*connect.Response[salesorderv1.CreateDriverResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "logistics", "write"); err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, invalidArgField("name")
	}
	uid, err := parseID(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}
	engine, err := logisticsEngine(ctx)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	// 司機必須是同公司的 active 使用者(跨公司参照 → not_found,不洩漏存在性)。
	if _, err := db.User.Query().Where(
		user.ID(uid),
		user.HasCompanyWith(company.IDEQ(cid)),
		user.StatusEQ(user.StatusActive),
	).Only(ctx); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	exists, err := db.LogisticsDriver.Query().Where(
		logisticsdriver.UserIDEQ(uid),
		logisticsdriver.CompanyIDEQ(cid),
		logisticsdriver.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if exists {
		return nil, errcode.SysConflict.Error(nil)
	}

	build := db.LogisticsDriver.Create().
		SetCompanyID(cid).
		SetUserID(uid).
		SetName(name).
		SetPhone(req.Msg.GetPhone())
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	actor, _ := parseID(id.UserID)
	if err := createAudit(ctx, tx, "logistics_driver", created.ID, cid, did, actor, map[string]any{
		"user_id": uid, "name": name,
	}); err != nil {
		return nil, err
	}
	if err := afterCommitLogisticsTuples(ctx, engine, driverObj(created.ID),
		driverDesiredTuples(created.ID, cid, did, uid)); err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.CreateDriverResponse{Driver: driverToProto(created)}), nil
}

// CreateVehicle 建車輛(10.1:plate_no 部門內唯一)。
func (s *LogisticsService) CreateVehicle(ctx context.Context, req *connect.Request[salesorderv1.CreateVehicleRequest]) (*connect.Response[salesorderv1.CreateVehicleResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "logistics", "write"); err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	plate := strings.TrimSpace(req.Msg.GetPlateNo())
	if plate == "" {
		return nil, invalidArgField("plate_no")
	}
	engine, err := logisticsEngine(ctx)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	dup := db.Vehicle.Query().Where(
		vehicle.PlateNoEQ(plate),
		vehicle.CompanyIDEQ(cid),
		vehicle.DeletedAtIsNil(),
	)
	if did != nil {
		dup = dup.Where(vehicle.DepartmentIDEQ(*did))
	} else {
		dup = dup.Where(vehicle.DepartmentIDIsNil())
	}
	exists, err := dup.Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if exists {
		return nil, errcode.SysConflict.Error(nil)
	}

	build := db.Vehicle.Create().
		SetCompanyID(cid).
		SetPlateNo(plate).
		SetVehicleType(req.Msg.GetVehicleType())
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	actor, _ := parseID(id.UserID)
	if err := createAudit(ctx, tx, "vehicle", created.ID, cid, did, actor, map[string]any{
		"plate_no": plate,
	}); err != nil {
		return nil, err
	}
	if err := afterCommitLogisticsTuples(ctx, engine, vehicleObj(created.ID),
		vehicleDesiredTuples(created.ID, cid, did)); err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.CreateVehicleResponse{Vehicle: vehicleToProto(created)}), nil
}

// AssignDelivery 指派車次 → 司機/車輛(10.4:粒度 = route;version 樂觀鎖;
// 重指派 version 遞增 + 稽核;tuple 經 AfterCommit 對帳)。
// driver_id/vehicle_id 空字串 = 解除對應指派。
func (s *LogisticsService) AssignDelivery(ctx context.Context, req *connect.Request[salesorderv1.AssignDeliveryRequest]) (*connect.Response[salesorderv1.AssignDeliveryResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "logistics", "write"); err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetRouteId())
	if err != nil {
		return nil, err
	}
	engine, err := logisticsEngine(ctx)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	// 車次必須在範圍內(公司必同;有部門 scope 時部門亦須同;RLS 兜底)。
	routeQ := db.Route.Query().Where(route.ID(rid), route.CompanyIDEQ(cid), route.DeletedAtIsNil())
	if did != nil {
		routeQ = routeQ.Where(route.DepartmentIDEQ(*did))
	}
	if _, err := routeQ.Only(ctx); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}

	var driverPtr, vehiclePtr *int
	if raw := strings.TrimSpace(req.Msg.GetDriverId()); raw != "" {
		dvid, err := parseID(raw)
		if err != nil {
			return nil, err
		}
		dq := db.LogisticsDriver.Query().Where(
			logisticsdriver.ID(dvid), logisticsdriver.CompanyIDEQ(cid), logisticsdriver.DeletedAtIsNil())
		if did != nil {
			dq = dq.Where(logisticsdriver.DepartmentIDEQ(*did))
		}
		if _, err := dq.Only(ctx); err != nil {
			if ent.IsNotFound(err) {
				return nil, errcode.SysNotFound.Error(nil)
			}
			return nil, toConnectError(err)
		}
		driverPtr = &dvid
	}
	if raw := strings.TrimSpace(req.Msg.GetVehicleId()); raw != "" {
		vid, err := parseID(raw)
		if err != nil {
			return nil, err
		}
		vq := db.Vehicle.Query().Where(
			vehicle.ID(vid), vehicle.CompanyIDEQ(cid), vehicle.DeletedAtIsNil())
		if did != nil {
			vq = vq.Where(vehicle.DepartmentIDEQ(*did))
		}
		if _, err := vq.Only(ctx); err != nil {
			if ent.IsNotFound(err) {
				return nil, errcode.SysNotFound.Error(nil)
			}
			return nil, toConnectError(err)
		}
		vehiclePtr = &vid
	}

	actor, _ := parseID(id.UserID)
	rawVersion := strings.TrimSpace(req.Msg.GetVersion())

	existing, err := db.LogisticsDelivery.Query().Where(
		logisticsdelivery.RouteIDEQ(rid),
		logisticsdelivery.CompanyIDEQ(cid),
		logisticsdelivery.DeletedAtIsNil(),
	).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, toConnectError(err)
	}

	var saved *ent.LogisticsDelivery
	var auditAction string
	switch {
	case ent.IsNotFound(err):
		// 建立:version 必須是 "0"(未指派過)。
		if rawVersion != "" && rawVersion != "0" {
			return nil, errcode.SysInvalidArgument.
				Error(map[string]string{"reason": "資料已變更，請重新載入"})
		}
		build := db.LogisticsDelivery.Create().
			SetCompanyID(cid).
			SetRouteID(rid).
			SetAssignedBy(actor)
		if did != nil {
			build = build.SetDepartmentID(*did)
		}
		if driverPtr != nil {
			build = build.SetDriverID(*driverPtr)
		}
		if vehiclePtr != nil {
			build = build.SetVehicleID(*vehiclePtr)
		}
		saved, err = build.Save(ctx)
		if err != nil {
			return nil, toConnectError(err)
		}
		auditAction = "create"
	default:
		// 重指派:樂觀鎖(不符即拒絕,同 dispatch 慣例 SYS-1001 + reason)。
		expect, err := strconv.Atoi(rawVersion)
		if err != nil || expect != existing.Version {
			return nil, errcode.SysInvalidArgument.
				Error(map[string]string{"reason": "資料已變更，請重新載入"})
		}
		up := db.LogisticsDelivery.UpdateOneID(existing.ID).
			SetAssignedBy(actor).
			SetVersion(existing.Version + 1)
		if driverPtr != nil {
			up = up.SetDriverID(*driverPtr)
		} else {
			up = up.ClearDriverID()
		}
		if vehiclePtr != nil {
			up = up.SetVehicleID(*vehiclePtr)
		} else {
			up = up.ClearVehicleID()
		}
		saved, err = up.Save(ctx)
		if err != nil {
			return nil, toConnectError(err)
		}
		// action 枚舉無 "reassign"(§5.2 列舉) → 重指派以 update 記錄,
		// 新舊值由 before/after 快照承載(route/driver/vehicle/version)。
		auditAction = "update"
	}

	if err := recordAuditBA(ctx, tx, "logistics_delivery", auditAction, saved.ID, cid, did, actor,
		map[string]any{"route_id": rid, "version": saved.Version - 1},
		map[string]any{"route_id": rid, "version": saved.Version,
			"driver_id": driverPtr, "vehicle_id": vehiclePtr}); err != nil {
		return nil, err
	}
	if err := afterCommitLogisticsTuples(ctx, engine, logisticsObj(saved.ID),
		deliveryDesiredTuples(saved.ID, cid, did, driverPtr, actor)); err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.AssignDeliveryResponse{
		Delivery: logisticsDeliveryToProto(saved),
	}), nil
}

// ListMyDeliveries 我被指派的配送(10.12 閘門①:身分必須對應 logistics_drivers 列;
// 10.8 閘門②:逐列 OpenFGA Check —— tuple 缺席的列被過濾,不做 400/403 驚嚇)。
func (s *LogisticsService) ListMyDeliveries(ctx context.Context, req *connect.Request[salesorderv1.ListMyDeliveriesRequest]) (*connect.Response[salesorderv1.ListMyDeliveriesResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireScope(ctx, "logistics", "read"); err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	engine, err := logisticsEngine(ctx)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()

	me, _ := parseID(id.UserID)
	drvQ := db.LogisticsDriver.Query().Where(
		logisticsdriver.UserIDEQ(me),
		logisticsdriver.CompanyIDEQ(cid),
		logisticsdriver.DeletedAtIsNil(),
	)
	if did != nil {
		drvQ = drvQ.Where(logisticsdriver.DepartmentIDEQ(*did))
	}
	drv, err := drvQ.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 10.12:非司機身分不得使用本端點。
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		return nil, toConnectError(err)
	}

	rows, err := db.LogisticsDelivery.Query().Where(
		logisticsdelivery.DriverIDEQ(drv.ID),
		logisticsdelivery.DeletedAtIsNil(),
	).Order(ent.Desc(logisticsdelivery.FieldID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// instance 級過濾(10.8):Check 不過的列視同不可見(與 list-objects 語意一致)。
	visible := make([]*ent.LogisticsDelivery, 0, len(rows))
	subject := "user:" + strconv.Itoa(me)
	for _, d := range rows {
		ok, cerr := engine.Check(ctx, subject, "can_read", logisticsObj(d.ID))
		if cerr != nil {
			return nil, errcode.SysInternal.Error(nil)
		}
		if ok {
			visible = append(visible, d)
		}
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	start := (page - 1) * pageSize
	if start > len(visible) {
		start = len(visible)
	}
	end := start + pageSize
	if end > len(visible) {
		end = len(visible)
	}
	out := make([]*salesorderv1.LogisticsDelivery, 0, end-start)
	for _, d := range visible[start:end] {
		out = append(out, logisticsDeliveryToProto(d))
	}
	return connect.NewResponse(&salesorderv1.ListMyDeliveriesResponse{
		Deliveries: out, Total: int32(len(visible)),
	}), nil
}
