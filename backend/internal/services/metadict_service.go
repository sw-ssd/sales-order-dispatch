// MetadictService 字典檔領域(03 計畫 2.5, D10/D11)。
// 單表兩層:department_id IS NULL = 系統預設(僅 super 可寫);非 NULL = 部門擴充。
// 寫入(create/update/delete)與 audit.Record 同一 DB 交易(D18);order_source 為系統固定不可異動。
package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	metadictv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/metadict/v1/metadictv1connect"
)

// validMetadictTypes 為字典 type 合法值(細部 2.5.4 / Global Constraints)。
var validMetadictTypes = map[string]bool{
	"unit": true, "payment_method": true, "settlement_method": true,
	"customer_type": true, "invoice_type": true, "order_source": true,
}

// systemFixedTypes 為系統級固定、API 不可異動的 type(order_source,供 05 取號)。
var systemFixedTypes = map[string]bool{"order_source": true}

// maxMetadictCodeLen 為字典 code 長度上限(細部 2.5.2:過長 → invalid_argument)。
const maxMetadictCodeLen = 64

// MetadictService 實作 metadict.v1.MetadictService(字典 CRUD + ListOptions)。
// 範圍:super/developer 全域(系統級+可按部門檢視);dept_admin/staff 僅系統預設+自己部門;customer/guest 僅系統預設(選項)。
type MetadictService struct {
	db *ent.Client
	metadictv1connect.UnimplementedMetadictServiceHandler
}

// NewMetadictService 建立 MetadictService。
func NewMetadictService(db *ent.Client) *MetadictService {
	return &MetadictService{db: db}
}

// RegisterMetadictServices 將 MetadictService 的 Connect handler 掛到 mux。
func RegisterMetadictServices(mux *http.ServeMux, db *ent.Client) {
	path, handler := metadictv1connect.NewMetadictServiceHandler(NewMetadictService(db), dbtenant.HandlerOption(db))
	mux.Handle(path, handler)
}

// metadictToProto 將 ent.Metadict 轉為 proto Metadict。
func metadictToProto(m *ent.Metadict) *metadictv1.Metadict {
	p := &metadictv1.Metadict{
		Id:          strconv.FormatInt(int64(m.ID), 10),
		Type:        m.Type,
		Code:        m.Code,
		DisplayName: m.DisplayName,
		SortOrder:   int32(m.SortOrder),
		IsActive:    m.IsActive,
	}
	if m.DepartmentID != nil {
		p.DepartmentId = strconv.FormatInt(int64(*m.DepartmentID), 10)
	}
	if !m.CreatedAt.IsZero() {
		p.CreatedAt = m.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !m.UpdatedAt.IsZero() {
		p.UpdatedAt = m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return p
}

// metadictScope 依身分回傳可見範圍的 where:系統預設 + 當前部門擴充(部門無脈絡則僅系統預設)。
func metadictScope(id authz.Identity) (predicate func(q *ent.MetadictQuery) *ent.MetadictQuery, err error) {
	if isSuperIdentity(id) {
		return func(q *ent.MetadictQuery) *ent.MetadictQuery { return q }, nil // 全域:範圍由各方法另處理
	}
	if id.DepartmentID == "" {
		// 無部門(company_admin / customer / guest):僅系統預設(細部 2.5.3 邊界)。
		return func(q *ent.MetadictQuery) *ent.MetadictQuery {
			return q.Where(metadict.DepartmentIDIsNil())
		}, nil
	}
	did, err := parseID(id.DepartmentID)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("部門脈絡異常,請重新登入"))
	}
	return func(q *ent.MetadictQuery) *ent.MetadictQuery {
		return q.Where(metadict.Or(metadict.DepartmentIDIsNil(), metadict.DepartmentIDEQ(did)))
	}, nil
}

// metadictWritable 判定操作者可否寫入目標字典(系統級僅 super;部門級需同部門)。
func metadictWritable(id authz.Identity, m *ent.Metadict) bool {
	if m.DepartmentID == nil { // 系統級
		return isSuperIdentity(id)
	}
	// 部門級:需操作者有部門且等於目標部門。
	if id.DepartmentID == "" {
		return false
	}
	did, err := parseID(id.DepartmentID)
	if err != nil {
		return false
	}
	return *m.DepartmentID == did
}

// metadictActorDeptID 回傳稽核用的部門(目標字典所屬;系統級為 nil)。
func metadictActorDeptID(m *ent.Metadict) *int {
	return m.DepartmentID
}

// ListMetadicts 分頁查詢(系統預設 + 當前部門擴充合併;super 可指定部門檢視)。
func (s *MetadictService) ListMetadicts(ctx context.Context, req *connect.Request[metadictv1.ListMetadictsRequest]) (*connect.Response[metadictv1.ListMetadictsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	q := s.db.Metadict.Query()

	switch {
	case isSuperIdentity(id):
		// super/developer:全域。可依請求 department_id 檢視指定部門擴充;空 = 僅系統預設(管理用途)。
		if req.Msg.GetDepartmentId() != "" {
			did, err := parseID(req.Msg.GetDepartmentId())
			if err != nil {
				return nil, err
			}
			q = q.Where(metadict.Or(metadict.DepartmentIDIsNil(), metadict.DepartmentIDEQ(did)))
		} else {
			q = q.Where(metadict.DepartmentIDIsNil())
		}
	default:
		scope, err := metadictScope(id)
		if err != nil {
			return nil, err
		}
		q = scope(q)
	}

	// 已刪除:預設排除;include_deleted 僅 super 管理介面可用(非 super 一律排除)。
	if !req.Msg.GetIncludeDeleted() || !isSuperIdentity(id) {
		q = q.Where(metadict.DeletedAtIsNil())
	}
	if t := req.Msg.GetType(); t != "" {
		if !validMetadictTypes[t] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 type %q", t))
		}
		q = q.Where(metadict.TypeEQ(t))
	}

	list, pg, err := pageList(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), metadictListSource{q}, metadictToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&metadictv1.ListMetadictsResponse{Items: list, Pagination: pg}), nil
}

// metadictListSource 為 pageList 的 ent 查詢橋接。
type metadictListSource struct{ q *ent.MetadictQuery }

func (s metadictListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s metadictListSource) Page(ctx context.Context, off, lim int) ([]*ent.Metadict, error) {
	// F2(與 F1 同型):code 的唯一性只有 (type, 部門),可見集合(系統預設 + 當前部門)內
	// (sort_order, code) 並不唯一 —— 同值群被頁邊界切開時 PostgreSQL 的 ties 順序又不保證
	// 一致,故以 id 為次序鍵收斂成全序。
	return s.q.Clone().Order(ent.Asc(metadict.FieldSortOrder), ent.Asc(metadict.FieldCode), ent.Asc(metadict.FieldID)).Offset(off).Limit(lim).All(ctx)
}

// GetMetadict 取得單一字典(限可見範圍)。
func (s *MetadictService) GetMetadict(ctx context.Context, req *connect.Request[metadictv1.GetMetadictRequest]) (*connect.Response[metadictv1.GetMetadictResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	mid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	// 可見範圍條件直接併入查詢(單次往返);範圍外一律 not_found(不區分不存在/越權)。
	scope, err := metadictScope(id)
	if err != nil {
		return nil, err
	}
	m, err := scope(s.db.Metadict.Query().Where(metadict.ID(mid), metadict.DeletedAtIsNil())).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&metadictv1.GetMetadictResponse{Metadict: metadictToProto(m)}), nil
}

// CreateMetadict 建立字典:super 建系統級(department NULL);dept_admin/staff 自動帶當前部門,不接受請求帶 department_id。
func (s *MetadictService) CreateMetadict(ctx context.Context, req *connect.Request[metadictv1.CreateMetadictRequest]) (*connect.Response[metadictv1.CreateMetadictResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	typ := strings.TrimSpace(req.Msg.GetType())
	code := strings.TrimSpace(req.Msg.GetCode())
	name := strings.TrimSpace(req.Msg.GetDisplayName())
	if !validMetadictTypes[typ] {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 type %q", typ))
	}
	if code == "" || name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code 與 display_name 必填"))
	}
	if len(code) > maxMetadictCodeLen {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("code 過長(上限 %d 字元)", maxMetadictCodeLen))
	}
	if systemFixedTypes[typ] {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("%s 為系統固定值,不可異動", typ))
	}

	// 範圍推導:super → 系統級;dept_admin/staff → 強制當前部門。
	var deptID *int
	switch {
	case isSuperIdentity(id):
		deptID = nil // 系統級
	case id.DepartmentID != "":
		did, err := parseID(id.DepartmentID)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("部門脈絡異常,請重新登入"))
		}
		deptID = &did
	default:
		// company_admin / customer / guest:不可建立(無系統級寫權、無部門脈絡)。
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無權限建立字典"))
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	build := tx.Metadict.Create().
		SetType(typ).
		SetCode(code).
		SetDisplayName(name).
		SetSortOrder(int(req.Msg.GetSortOrder())).
		SetIsActive(req.Msg.GetIsActive())
	if deptID != nil {
		build = build.SetDepartmentID(*deptID)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	actorID := actorIDFrom(id)
	cid := companyIDFrom(id)
	if err := recordAudit(ctx, tx, "metadict", "create", created.ID, cid, created.DepartmentID, actorID, map[string]any{"type": created.Type, "code": created.Code, "display_name": created.DisplayName, "department_id": created.DepartmentID, "is_active": created.IsActive}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&metadictv1.CreateMetadictResponse{Metadict: metadictToProto(created)}), nil
}

// UpdateMetadict 更新 display_name / sort_order / is_active(不含 type / code;order_source 不可異動)。
func (s *MetadictService) UpdateMetadict(ctx context.Context, req *connect.Request[metadictv1.UpdateMetadictRequest]) (*connect.Response[metadictv1.UpdateMetadictResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	mid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := s.db.Metadict.Query().Where(metadict.ID(mid), metadict.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if systemFixedTypes[m.Type] {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("%s 為系統固定值,不可異動", m.Type))
	}
	if !metadictWritable(id, m) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無權限修改此字典"))
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	upd := tx.Metadict.UpdateOneID(mid)
	if req.Msg.DisplayName != nil {
		upd = upd.SetDisplayName(*req.Msg.DisplayName)
	}
	if req.Msg.SortOrder != nil {
		upd = upd.SetSortOrder(int(*req.Msg.SortOrder))
	}
	if req.Msg.IsActive != nil {
		upd = upd.SetIsActive(*req.Msg.IsActive)
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	if err := recordAuditBA(ctx, tx, "metadict", "update", mid, companyIDFrom(id), metadictActorDeptID(m), actorIDFrom(id), map[string]any{"display_name": m.DisplayName, "sort_order": m.SortOrder, "is_active": m.IsActive}, map[string]any{"display_name": updated.DisplayName, "sort_order": updated.SortOrder, "is_active": updated.IsActive}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&metadictv1.UpdateMetadictResponse{Metadict: metadictToProto(updated)}), nil
}

// DeleteMetadict 軟刪除(設定 deleted_at;order_source 不可刪;復原不在 1.0 API 範圍)。
func (s *MetadictService) DeleteMetadict(ctx context.Context, req *connect.Request[metadictv1.DeleteMetadictRequest]) (*connect.Response[metadictv1.DeleteMetadictResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	mid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := s.db.Metadict.Query().Where(metadict.ID(mid), metadict.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if systemFixedTypes[m.Type] {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("%s 為系統固定值,不可異動", m.Type))
	}
	if !metadictWritable(id, m) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("無權限刪除此字典"))
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.Metadict.UpdateOneID(mid).SetDeletedAt(time.Now().UTC()).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAuditBA(ctx, tx, "metadict", "delete", mid, companyIDFrom(id), metadictActorDeptID(m), actorIDFrom(id), map[string]any{"type": m.Type, "code": m.Code, "display_name": m.DisplayName}, nil); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&metadictv1.DeleteMetadictResponse{}), nil
}

// ListOptions 表單下拉選項:僅可選用啟用值(系統預設 + 當前部門擴充;客戶端僅系統預設)。
func (s *MetadictService) ListOptions(ctx context.Context, req *connect.Request[metadictv1.ListOptionsRequest]) (*connect.Response[metadictv1.ListOptionsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	typ := strings.TrimSpace(req.Msg.GetType())
	if !validMetadictTypes[typ] {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("type 必填且必須為合法值,got %q", typ))
	}
	q := s.db.Metadict.Query().
		Where(metadict.TypeEQ(typ), metadict.IsActiveEQ(true), metadict.DeletedAtIsNil())
	// 可見範圍:super 僅系統預設(與 ListMetadicts 預設一致,避免下拉混入各部門私有值);
	// 其餘依身分(系統+當前部門 / 僅系統)。
	if isSuperIdentity(id) {
		q = q.Where(metadict.DepartmentIDIsNil())
	} else {
		scope, err := metadictScope(id)
		if err != nil {
			return nil, err
		}
		q = scope(q)
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(metadict.Or(
			metadict.DisplayNameContains(kw),
			metadict.CodeContains(kw),
		))
	}
	items, err := q.Order(ent.Asc(metadict.FieldSortOrder), ent.Asc(metadict.FieldCode)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	opts := make([]*metadictv1.Option, 0, len(items))
	for _, m := range items {
		opts = append(opts, &metadictv1.Option{Code: m.Code, DisplayName: m.DisplayName})
	}
	return connect.NewResponse(&metadictv1.ListOptionsResponse{Options: opts}), nil
}

// actorIDFrom 由身分取出操作者 int id;缺則 0(audit.Record 會以缺操作者拒寫,符合 D18)。
func actorIDFrom(id authz.Identity) int {
	if pid, err := parseID(id.UserID); err == nil {
		return pid
	}
	return 0
}

// companyIDFrom 由身分取出公司 int id;缺則 0(audit.Record 會以缺租戶拒寫,符合 D18)。
func companyIDFrom(id authz.Identity) int {
	if cid, err := parseID(id.CompanyID); err == nil {
		return cid
	}
	return 0
}
