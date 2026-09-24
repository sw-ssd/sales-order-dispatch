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
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/ent/logisticsdelivery"
	"github.com/salesorder/sales-order-1.0/backend/ent/logisticsdriver"
	"github.com/salesorder/sales-order-1.0/backend/ent/route"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
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
	if d.StartedAt != nil {
		out.StartedAt = d.StartedAt.UTC().Format(time.RFC3339)
	}
	if d.CompletedAt != nil {
		out.CompletedAt = d.CompletedAt.UTC().Format(time.RFC3339)
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
	// **部門以該使用者的所屬部門為準**,而非建立者的 scope:company_admin 的 scope 是公司層
	// (did=nil),若沿用會種出 department_id IS NULL 的司機 —— 該司機登入後 deptScope 取不到
	// 部門,ListMyDeliveries 直接 403(實機踩到)。建立者 scope 僅作為可建範圍的上界。
	drvUser, err := db.User.Query().Where(
		user.ID(uid),
		user.HasCompanyWith(company.IDEQ(cid)),
		user.StatusEQ(user.StatusActive),
	).WithDepartment().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	if did != nil {
		// 部門管理者只能建本部門的司機(上界收斂)。
		if drvUser.Edges.Department == nil || drvUser.Edges.Department.ID != *did {
			return nil, errcode.SysNotFound.Error(nil)
		}
	}
	drvDept := did
	if drvUser.Edges.Department != nil {
		d := drvUser.Edges.Department.ID
		drvDept = &d
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
	if drvDept != nil {
		build = build.SetDepartmentID(*drvDept)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	actor, _ := parseID(id.UserID)
	if err := createAudit(ctx, tx, "logistics_driver", created.ID, cid, drvDept, actor, map[string]any{
		"user_id": uid, "name": name,
	}); err != nil {
		return nil, err
	}
	if err := afterCommitLogisticsTuples(ctx, engine, driverObj(created.ID),
		driverDesiredTuples(created.ID, cid, drvDept, uid)); err != nil {
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
	routeRow, err := routeQ1(ctx, db, rid, cid, did)
	if err != nil {
		return nil, err
	}
	// 配送的部門**以車次為準**(車次本身即部門級主檔):指派者的 scope 只是可操作上界,
	// company_admin 的 did=nil 不代表執行單該無部門 —— 沿用會讓被指派司機取不到部門 scope
	// 而看不到自己的任務(與 CreateDriver 同一類 bug,2026-09-24 實機踩到)。
	deliveryDept := did
	if routeRow.DepartmentID != nil {
		d := *routeRow.DepartmentID
		deliveryDept = &d
	}

	var driverPtr, vehiclePtr *int
	var driverUserID int // 被指派司機的 users.id(assignee 邊的主體)
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
		drvRow, err := dq.Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, errcode.SysNotFound.Error(nil)
			}
			return nil, toConnectError(err)
		}
		driverPtr = &dvid
		driverUserID = drvRow.UserID
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
		if deliveryDept != nil {
			build = build.SetDepartmentID(*deliveryDept)
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
		// 終態(completed/cancelled)不得重指派 —— 10.6 的狀態機宣告終態無出口,
		// 若允許改 driver/vehicle,會出現「狀態已完成但司機換人」且 FGA 判決隨之移交的
		// 矛盾(實作狀態機後補上的守衛)。要重跑請對新車次建立新的配送單。
		if isTerminalDeliveryStatus(existing.Status) {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{
				"reason": "配送已完成或已取消，不可重指派", "status": existing.Status,
			})
		}
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
		deliveryDesiredTuples(saved.ID, cid, deliveryDept, driverPtr, actor)); err != nil {
		return nil, err
	}
	// 一併對帳被指派司機物件:tuple 寫入是外部副作用,失敗只記日誌(見 logistics_events.go),
	// 若只在 CreateDriver 寫一次,該次失敗(如 model 漂移)會在司機物件留下永久空洞 ——
	// 指派時順帶補齊,才真的讓「重指派即修復」對 driver 邊也成立(2026-09-24 實機踩到)。
	if driverPtr != nil {
		if err := afterCommitLogisticsTuples(ctx, engine, driverObj(*driverPtr),
			driverDesiredTuples(*driverPtr, cid, deliveryDept, driverUserID)); err != nil {
			return nil, err
		}
	}
	// 10.11 指派通知:推新任務給司機本人(該 driver 關聯的 users.id)。
	if driverUserID > 0 {
		if err := OnDeliveryAssigned(ctx, db, cid, deliveryDept, driverUserID,
			saved.ID, routeRow.Name, routeRow.Code); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&salesorderv1.AssignDeliveryResponse{
		Delivery: logisticsDeliveryToProto(saved),
	}), nil
}

// deliverySelfGate 為 StartDelivery / CompleteDelivery / CancelDelivery 共用的前置:
// 本人閘門 + instance 級授權(狀態轉移由呼叫端執行,以便各自附加事件/稽核/POD)。
// 10.6/10.12:操作者必須是該配送被指派的司機本人(logistics_drivers.user_id = 身分);
// OpenFGA instance 級 can_write 亦須通過(10.8,tuple 缺席 → 拒絕)。
func (s *LogisticsService) deliverySelfGate(ctx context.Context, cid int,
	did *int, me, deliveryID int) (*ent.LogisticsDelivery, error) {
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	q := db.LogisticsDelivery.Query().Where(
		logisticsdelivery.ID(deliveryID),
		logisticsdelivery.CompanyIDEQ(cid),
		logisticsdelivery.DeletedAtIsNil(),
	)
	if did != nil {
		q = q.Where(logisticsdelivery.DepartmentIDEQ(*did))
	}
	d, err := q.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	// 本人:以 logistics_drivers 列取得司機 id,再比對該配送的 driver_id。
	drv, err := db.LogisticsDriver.Query().Where(
		logisticsdriver.UserIDEQ(me),
		logisticsdriver.CompanyIDEQ(cid),
		logisticsdriver.DeletedAtIsNil(),
	).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		return nil, toConnectError(err)
	}
	if d.DriverID == nil || *d.DriverID != drv.ID {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	// instance 級(10.8):非本人不得操作(沿 driver#assignee 抵達 can_write)。
	engine, err := logisticsEngine(ctx)
	if err != nil {
		return nil, err
	}
	ok, cerr := engine.Check(ctx, "user:"+strconv.Itoa(me), "can_write", logisticsObj(d.ID))
	if cerr != nil {
		return nil, errcode.SysInternal.Error(nil)
	}
	if !ok {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	return d, nil
}

// StartDelivery 開始執行(pending → in_progress;10.6)。
func (s *LogisticsService) StartDelivery(ctx context.Context, req *connect.Request[salesorderv1.StartDeliveryRequest]) (*connect.Response[salesorderv1.StartDeliveryResponse], error) {
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
	deliveryID, err := parseID(req.Msg.GetDeliveryId())
	if err != nil {
		return nil, err
	}
	me, _ := parseID(id.UserID)
	if _, err := s.deliverySelfGate(ctx, cid, did, me, deliveryID); err != nil {
		return nil, err
	}
	tx, _ := dbtenant.TxFrom(ctx)
	ver := 0
	if v := strings.TrimSpace(req.Msg.GetVersion()); v != "" {
		ver, err = strconv.Atoi(v)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "version"})
		}
	}
	saved, err := TransitionDelivery(ctx, tx.Client(), DeliveryTransitionInput{
		DeliveryID: deliveryID, To: DeliveryStatusInProgress, ActorID: me, ExpectedVersion: ver,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.StartDeliveryResponse{
		Delivery: logisticsDeliveryToProto(saved),
	}), nil
}

// CompleteDelivery 完成並簽收(in_progress → completed;POD 多筆,同交易寫入)。
// 10.6:POD 為既有檔案的引用(file_assets,owner_type=logistics_delivery),此處只驗歸屬。
func (s *LogisticsService) CompleteDelivery(ctx context.Context, req *connect.Request[salesorderv1.CompleteDeliveryRequest]) (*connect.Response[salesorderv1.CompleteDeliveryResponse], error) {
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
	deliveryID, err := parseID(req.Msg.GetDeliveryId())
	if err != nil {
		return nil, err
	}
	me, _ := parseID(id.UserID)
	if _, err := s.deliverySelfGate(ctx, cid, did, me, deliveryID); err != nil {
		return nil, err
	}
	tx, _ := dbtenant.TxFrom(ctx)
	db := tx.Client()
	ver := 0
	if v := strings.TrimSpace(req.Msg.GetVersion()); v != "" {
		ver, err = strconv.Atoi(v)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "version"})
		}
	}
	// POD:先驗每筆檔案歸屬(同公司、非軟刪除),再逐筆寫;任一失敗整交易回滾(10.6 驗收)。
	now := time.Now().UTC()
	var proofs []*ent.LogisticsProof
	for i, p := range req.Msg.GetProofs() {
		pt := strings.TrimSpace(p.GetProofType())
		if !validProofTypes[pt] {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{
				"field": "proofs[" + strconv.Itoa(i) + "].proof_type"})
		}
		faid, err := parseID(p.GetFileAssetId())
		if err != nil {
			return nil, err
		}
		fa, err := db.FileAsset.Query().Where(
			fileasset.ID(faid), fileasset.CompanyIDEQ(cid), fileasset.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, errcode.SysNotFound.Error(nil)
			}
			return nil, toConnectError(err)
		}
		// 只收本配送的 POD(owner_type/owner_id 精確對齊),避免拿別的檔案冒充簽收。
		if fa.OwnerType != logisticsDeliveryOwnerType || fa.OwnerID != deliveryID {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{
				"field": "proofs[" + strconv.Itoa(i) + "].file_asset_id"})
		}
		build := db.LogisticsProof.Create().
			SetLogisticsDeliveryID(deliveryID).
			SetCompanyID(cid).
			SetProofType(pt).
			SetFileAssetID(faid).
			SetCapturedBy(me).
			SetCapturedAt(now)
		if did != nil {
			build = build.SetDepartmentID(*did)
		}
		if r := strings.TrimSpace(p.GetRemarks()); r != "" {
			build = build.SetRemarks(r)
		}
		saved, err := build.Save(ctx)
		if err != nil {
			return nil, toConnectError(err)
		}
		proofs = append(proofs, saved)
	}

	saved, err := TransitionDelivery(ctx, tx.Client(), DeliveryTransitionInput{
		DeliveryID: deliveryID, To: DeliveryStatusCompleted, ActorID: me, ExpectedVersion: ver,
	})
	if err != nil {
		return nil, err
	}
	// 10.9 送達回寫:該車次所載每筆訂單標送達(processing → completed + delivered_at +
	// 事件 + 稽核,同一交易)。任一筆失敗 → 整交易回滾(含上面的配送完成與 POD),
	// 避免「配送完成但訂單沒結」的半套狀態。
	delivered, err := s.writebackDeliveredOrders(ctx, tx.Client(), saved, me)
	if err != nil {
		return nil, err
	}
	if err := recordAuditBA(ctx, tx, "logistics_delivery", "update", saved.ID, cid, did, me,
		map[string]any{"status": DeliveryStatusInProgress},
		map[string]any{"status": DeliveryStatusCompleted, "proofs": len(proofs),
			"delivered_orders": len(delivered)}); err != nil {
		return nil, err
	}
	// 10.11 送達通知:逐筆已送達訂單推店家(該客戶全部子帳號)/主責業務。
	// 與其他觸發一致 —— 通知列在**同一交易**建立(status=pending),FCM 於提交後發送,
	// 失敗只標 failed 不回滾(D16)。
	if err := OnDeliveryCompleted(ctx, tx.Client(), cid, saved.DepartmentID, saved.RouteID, delivered); err != nil {
		return nil, err
	}
	out := make([]*salesorderv1.LogisticsProof, 0, len(proofs))
	for _, p := range proofs {
		out = append(out, proofToProto(p))
	}
	return connect.NewResponse(&salesorderv1.CompleteDeliveryResponse{
		Delivery: logisticsDeliveryToProto(saved), Proofs: out,
	}), nil
}

// CancelDelivery 取消配送(pending/in_progress → cancelled;reason 必填;10.6)。
func (s *LogisticsService) CancelDelivery(ctx context.Context, req *connect.Request[salesorderv1.CancelDeliveryRequest]) (*connect.Response[salesorderv1.CancelDeliveryResponse], error) {
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
	deliveryID, err := parseID(req.Msg.GetDeliveryId())
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(req.Msg.GetReason())
	if reason == "" {
		return nil, invalidArgField("reason")
	}
	me, _ := parseID(id.UserID)
	if _, err := s.deliverySelfGate(ctx, cid, did, me, deliveryID); err != nil {
		return nil, err
	}
	tx, _ := dbtenant.TxFrom(ctx)
	ver := 0
	if v := strings.TrimSpace(req.Msg.GetVersion()); v != "" {
		ver, err = strconv.Atoi(v)
		if err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "version"})
		}
	}
	saved, err := TransitionDelivery(ctx, tx.Client(), DeliveryTransitionInput{
		DeliveryID: deliveryID, To: DeliveryStatusCancelled, ActorID: me,
		Reason: reason, ExpectedVersion: ver,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&salesorderv1.CancelDeliveryResponse{
		Delivery: logisticsDeliveryToProto(saved),
	}), nil
}

// proofToProto 轉換 POD 列。
func proofToProto(p *ent.LogisticsProof) *salesorderv1.LogisticsProof {
	return &salesorderv1.LogisticsProof{
		Id:                  strconv.Itoa(p.ID),
		LogisticsDeliveryId: strconv.Itoa(p.LogisticsDeliveryID),
		ProofType:           p.ProofType,
		FileAssetId:         strconv.Itoa(p.FileAssetID),
		Remarks:             p.Remarks,
		CapturedAt:          p.CapturedAt.Format(time.RFC3339),
	}
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

// logisticsDeliveryOwnerType 為 POD 檔案的 owner_type(04 檔案資產的 owner 分類之一);
// 只收 owner 指向本配送的檔案,避免簽收引用無關檔案。
const logisticsDeliveryOwnerType = "logistics_delivery"

// validProofTypes 為 POD 型別白名單(10.6:photo / signature / scan)。
var validProofTypes = map[string]bool{"photo": true, "signature": true, "scan": true}

// routeQ1 取一筆在範圍內的車次(公司必同;呼叫者有部門 scope 時部門亦須同;RLS 兜底)。
// 找不到 → not_found(不洩漏存在性);其他錯誤轉 connect 錯誤。
func routeQ1(ctx context.Context, db *ent.Client, rid, cid int, did *int) (*ent.Route, error) {
	q := db.Route.Query().Where(route.ID(rid), route.CompanyIDEQ(cid), route.DeletedAtIsNil())
	if did != nil {
		q = q.Where(route.DepartmentIDEQ(*did))
	}
	row, err := q.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.SysNotFound.Error(nil)
		}
		return nil, toConnectError(err)
	}
	return row, nil
}

// writebackDeliveredOrders 執行 10.9 送達回寫:把該配送所屬車次上「運送中」的每筆訂單
// 標為完成並蓋送達時點,回傳成功回寫的訂單數。
//
// 篩選條件是 RouteID + status=processing:派車確認(08 ConfirmDispatch)會把訂單轉
// processing 並寫 route_id,所以「車次所載」精確等於「該 route 的 processing 訂單」。
// 不撈 pending/cancelled/completed —— 那些不是這台車載的貨;pulling them in would
// make an unrelated pending order block delivery completion (狀態機不允許 pending→completed)。
//
// 回寫本身走既有狀態機 TransitionOrder(不自行改 status):同一交易內逐筆條件更新,
// 任一笔失敗即回傳錯誤 → 呼叫端交易回滾(10.9 錯誤處理:不留部分送達)。
// 車次上沒有 processing 訂單(例如先建配送、後派車)不算錯誤:回寫 0 筆、配送照常完成。
func (s *LogisticsService) writebackDeliveredOrders(ctx context.Context, db *ent.Client,
	d *ent.LogisticsDelivery, actor int) ([]*ent.SalesOrder, error) {
	if d.RouteID == 0 {
		return nil, nil
	}
	orders, err := db.SalesOrder.Query().Where(
		salesorder.RouteIDEQ(d.RouteID),
		salesorder.CompanyIDEQ(d.CompanyID),
		salesorder.StatusEQ(OrderStatusProcessing),
		salesorder.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := markOrdersDelivered(ctx, db, orders, actor); err != nil {
		return nil, err
	}
	return orders, nil
}

// markOrdersDelivered 逐筆標送達(失敗即返回,由呼叫端的交易回滾)。
// 與查詢分開是為了讓「批次中任一筆失敗即整體回滾」可被直接測試(見
// logistics_delivered_writeback_test.go:不需要併發插隊就能驗證 fail-fast)。
func markOrdersDelivered(ctx context.Context, db *ent.Client, orders []*ent.SalesOrder, actor int) error {
	now := time.Now().UTC()
	for _, o := range orders {
		if err := TransitionOrder(ctx, db, OrderTransitionInput{
			OrderID: o.ID, To: OrderStatusCompleted, ActorID: actor,
			MarkDelivered: true, DeliveredAt: &now,
		}); err != nil {
			return err
		}
	}
	return nil
}
