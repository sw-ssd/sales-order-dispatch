// ProcessingSpecService 加工/處理規格(04 計畫 3.4.3,泛化自 cutting_specs, D10/D18):
// 商品無關、普適多商品;kind 開放集(metadicts processing_kind 背書)、多值旗標、attributes(後端不解析)。
// 共用流程見 master_crud.go。
package services

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/processingspec"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
)

// ProcessingSpecService 實作 masters.v1.ProcessingSpecService。
type ProcessingSpecService struct {
	db *ent.Client
	mastersv1connect.UnimplementedProcessingSpecServiceHandler
}

// NewProcessingSpecService 建立 ProcessingSpecService。
func NewProcessingSpecService(db *ent.Client) *ProcessingSpecService {
	return &ProcessingSpecService{db: db}
}

// RegisterProcessingSpecService 掛載。
func RegisterProcessingSpecService(mux *http.ServeMux, db *ent.Client) {
	path, handler := mastersv1connect.NewProcessingSpecServiceHandler(NewProcessingSpecService(db))
	mux.Handle(path, handler)
}

func specScopeQuery(q *ent.ProcessingSpecQuery, cid int, did *int) *ent.ProcessingSpecQuery {
	if did != nil {
		return q.Where(processingspec.CompanyIDEQ(cid), processingspec.DepartmentIDEQ(*did))
	}
	return q.Where(processingspec.CompanyIDEQ(cid))
}

func specToProto(p *ent.ProcessingSpec) *mastersv1.ProcessingSpec {
	out := &mastersv1.ProcessingSpec{
		Id: strconv.FormatInt(int64(p.ID), 10), CompanyId: strconv.FormatInt(int64(p.CompanyID), 10),
		Code: p.Code, Name: p.Name, Kind: p.Kind,
		AppliesToProcessing: p.AppliesToProcessing, AppliesToPicking: p.AppliesToPicking,
		SortOrder: int32(p.SortOrder), IsActive: p.IsActive,
	}
	if p.DepartmentID != nil {
		out.DepartmentId = strconv.FormatInt(int64(*p.DepartmentID), 10)
	}
	if len(p.Attributes) > 0 {
		if st, err := structpb.NewStruct(p.Attributes); err == nil {
			out.Attributes = st
		}
	}
	if !p.CreatedAt.IsZero() {
		out.CreatedAt = p.CreatedAt.Format(time.RFC3339)
	}
	if !p.UpdatedAt.IsZero() {
		out.UpdatedAt = p.UpdatedAt.Format(time.RFC3339)
	}
	if p.DeletedAt != nil {
		out.DeletedAt = p.DeletedAt.Format(time.RFC3339)
	}
	return out
}

// validateSpecFlags 多值旗標驗證:至少 applies_to_processing / applies_to_picking 其一為 true。
func validateSpecFlags(proc, pick bool) error {
	if !proc && !pick {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("applies_to_processing 與 applies_to_picking 至少其一為 true"))
	}
	return nil
}

type specListSource struct{ q *ent.ProcessingSpecQuery }

func (s specListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s specListSource) Page(ctx context.Context, off, lim int) ([]*ent.ProcessingSpec, error) {
	return s.q.Clone().Order(ent.Asc(processingspec.FieldSortOrder)).Offset(off).Limit(lim).All(ctx)
}

func (s *ProcessingSpecService) ListProcessingSpecs(ctx context.Context, req *connect.Request[mastersv1.ListProcessingSpecsRequest]) (*connect.Response[mastersv1.ListProcessingSpecsResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	q := specScopeQuery(s.db.ProcessingSpec.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(processingspec.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(processingspec.Or(processingspec.CodeContainsFold(kw), processingspec.NameContainsFold(kw)))
	}
	list, pg, err := masterPage(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), specListSource{q}, specToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mastersv1.ListProcessingSpecsResponse{ProcessingSpecs: list, Pagination: pg}), nil
}

func (s *ProcessingSpecService) CreateProcessingSpec(ctx context.Context, req *connect.Request[mastersv1.CreateProcessingSpecRequest]) (*connect.Response[mastersv1.CreateProcessingSpecResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	code, name, err := masterCodeName(req.Msg.GetCode(), req.Msg.GetName())
	if err != nil {
		return nil, err
	}
	if err := validateSpecFlags(req.Msg.GetAppliesToProcessing(), req.Msg.GetAppliesToPicking()); err != nil {
		return nil, err
	}
	kind := strings.TrimSpace(req.Msg.GetKind())
	if kind == "" {
		kind = "other"
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	build := tx.ProcessingSpec.Create().
		SetCompanyID(cid).SetCode(code).SetName(name).SetKind(kind).
		SetAppliesToProcessing(req.Msg.GetAppliesToProcessing()).
		SetAppliesToPicking(req.Msg.GetAppliesToPicking()).
		SetSortOrder(int(req.Msg.GetSortOrder())).SetIsActive(req.Msg.GetIsActive()).
		SetCreatedBy(actor).SetUpdatedBy(actor)
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if attrs := req.Msg.GetAttributes(); attrs != nil {
		build = build.SetAttributes(attrs.AsMap())
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordMasterAudit(ctx, tx, "processing_spec", "create", created.ID, cid, created.DepartmentID, actor, map[string]any{"code": created.Code, "name": created.Name, "kind": created.Kind}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateProcessingSpecResponse{ProcessingSpec: specToProto(created)}), nil
}

func (s *ProcessingSpecService) UpdateProcessingSpec(ctx context.Context, req *connect.Request[mastersv1.UpdateProcessingSpecRequest]) (*connect.Response[mastersv1.UpdateProcessingSpecResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	sid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("processing_spec id 格式錯誤"))
	}
	cur, err := specScopeQuery(s.db.ProcessingSpec.Query(), cid, did).Where(processingspec.ID(sid), processingspec.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	upd := tx.ProcessingSpec.UpdateOneID(sid)
	if req.Msg.Code != nil {
		c, err := masterTrimNonEmpty(*req.Msg.Code, "code 不可為空")
		if err != nil {
			return nil, err
		}
		upd = upd.SetCode(c)
	}
	if req.Msg.Name != nil {
		n, err := masterTrimNonEmpty(*req.Msg.Name, "name 不可為空")
		if err != nil {
			return nil, err
		}
		upd = upd.SetName(n)
	}
	if req.Msg.Kind != nil {
		k := strings.TrimSpace(*req.Msg.Kind)
		if k == "" {
			k = "other"
		}
		upd = upd.SetKind(k)
	}
	// 多值旗標:任一提供即套用;以最終值合併現值驗證至少其一 true。
	if req.Msg.AppliesToProcessing != nil || req.Msg.AppliesToPicking != nil {
		proc := cur.AppliesToProcessing
		pick := cur.AppliesToPicking
		if req.Msg.AppliesToProcessing != nil {
			proc = *req.Msg.AppliesToProcessing
		}
		if req.Msg.AppliesToPicking != nil {
			pick = *req.Msg.AppliesToPicking
		}
		if err := validateSpecFlags(proc, pick); err != nil {
			return nil, err
		}
		upd = upd.SetAppliesToProcessing(proc).SetAppliesToPicking(pick)
	}
	if req.Msg.Attributes != nil {
		upd = upd.SetAttributes(req.Msg.GetAttributes().AsMap())
	}
	if req.Msg.SortOrder != nil {
		upd = upd.SetSortOrder(int(*req.Msg.SortOrder))
	}
	if req.Msg.IsActive != nil {
		upd = upd.SetIsActive(*req.Msg.IsActive)
	}
	actor, _ := parseID(id.UserID)
	updated, err := upd.SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordMasterAudit(ctx, tx, "processing_spec", "update", sid, cid, updated.DepartmentID, actor, map[string]any{"name": updated.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateProcessingSpecResponse{ProcessingSpec: specToProto(updated)}), nil
}

func (s *ProcessingSpecService) DeleteProcessingSpec(ctx context.Context, req *connect.Request[mastersv1.DeleteProcessingSpecRequest]) (*connect.Response[mastersv1.DeleteProcessingSpecResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	sid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("processing_spec id 格式錯誤"))
	}
	cur, err := specScopeQuery(s.db.ProcessingSpec.Query(), cid, did).Where(processingspec.ID(sid), processingspec.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.ProcessingSpec.UpdateOneID(sid).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordMasterAudit(ctx, tx, "processing_spec", "delete", sid, cid, cur.DepartmentID, actor, map[string]any{"code": cur.Code, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteProcessingSpecResponse{}), nil
}

func (s *ProcessingSpecService) RestoreProcessingSpec(ctx context.Context, req *connect.Request[mastersv1.RestoreProcessingSpecRequest]) (*connect.Response[mastersv1.RestoreProcessingSpecResponse], error) {
	id, err := masterRequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	sid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("processing_spec id 格式錯誤"))
	}
	cur, err := specScopeQuery(s.db.ProcessingSpec.Query(), cid, did).Where(processingspec.ID(sid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		return connect.NewResponse(&mastersv1.RestoreProcessingSpecResponse{ProcessingSpec: specToProto(cur)}), nil
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	restored, err := tx.ProcessingSpec.UpdateOneID(sid).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordMasterAudit(ctx, tx, "processing_spec", "update", sid, cid, restored.DepartmentID, actor, map[string]any{"restored": true, "code": restored.Code}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreProcessingSpecResponse{ProcessingSpec: specToProto(restored)}), nil
}
