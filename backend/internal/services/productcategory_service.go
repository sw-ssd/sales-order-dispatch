// ProductCategoryService 商品分類(04 計畫 3.4.4, D10/D18):部門級單層分類 CRUD + 軟刪除/復原。
package services

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// ProductCategoryService 實作 masters.v1.ProductCategoryService。
type ProductCategoryService struct {
	db *ent.Client
	mastersv1connect.UnimplementedProductCategoryServiceHandler
}

// NewProductCategoryService 建立 ProductCategoryService。
func NewProductCategoryService(db *ent.Client) *ProductCategoryService {
	return &ProductCategoryService{db: db}
}

// RegisterProductCategoryService 掛載。
func RegisterProductCategoryService(mux *http.ServeMux, db *ent.Client) {
	path, handler := mastersv1connect.NewProductCategoryServiceHandler(NewProductCategoryService(db))
	mux.Handle(path, handler)
}

func catScopeQuery(q *ent.ProductCategoryQuery, cid int, did *int) *ent.ProductCategoryQuery {
	if did != nil {
		return q.Where(productcategory.CompanyIDEQ(cid), productcategory.DepartmentIDEQ(*did))
	}
	return q.Where(productcategory.CompanyIDEQ(cid))
}

func catToProto(c *ent.ProductCategory) *mastersv1.ProductCategory {
	p := &mastersv1.ProductCategory{
		Id: strconv.FormatInt(int64(c.ID), 10), CompanyId: strconv.FormatInt(int64(c.CompanyID), 10),
		Code: c.Code, Name: c.Name, SortOrder: int32(c.SortOrder), IsActive: c.IsActive,
	}
	if c.DepartmentID != nil {
		p.DepartmentId = strconv.FormatInt(int64(*c.DepartmentID), 10)
	}
	if !c.CreatedAt.IsZero() {
		p.CreatedAt = c.CreatedAt.Format(time.RFC3339)
	}
	if !c.UpdatedAt.IsZero() {
		p.UpdatedAt = c.UpdatedAt.Format(time.RFC3339)
	}
	if c.DeletedAt != nil {
		p.DeletedAt = c.DeletedAt.Format(time.RFC3339)
	}
	return p
}

func (s *ProductCategoryService) ListProductCategories(ctx context.Context, req *connect.Request[mastersv1.ListProductCategoriesRequest]) (*connect.Response[mastersv1.ListProductCategoriesResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := catScopeQuery(s.db.ProductCategory.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(productcategory.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(productcategory.Or(productcategory.CodeContainsFold(kw), productcategory.NameContainsFold(kw)))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Clone().Order(ent.Asc(productcategory.FieldSortOrder)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*mastersv1.ProductCategory, 0, len(items))
	for _, c := range items {
		out = append(out, catToProto(c))
	}
	return connect.NewResponse(&mastersv1.ListProductCategoriesResponse{
		ProductCategories: out, Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

func (s *ProductCategoryService) CreateProductCategory(ctx context.Context, req *connect.Request[mastersv1.CreateProductCategoryRequest]) (*connect.Response[mastersv1.CreateProductCategoryResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.Msg.GetCode())
	name := strings.TrimSpace(req.Msg.GetName())
	if code == "" || name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code 與 name 必填"))
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	build := tx.ProductCategory.Create().
		SetCompanyID(cid).SetCode(code).SetName(name).
		SetSortOrder(int(req.Msg.GetSortOrder())).SetIsActive(req.Msg.GetIsActive()).
		SetCreatedBy(actor).SetUpdatedBy(actor)
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "create", ResourceType: "product_category", ResourceID: strconv.FormatInt(int64(created.ID), 10),
		CompanyID: cid, DepartmentID: created.DepartmentID, UserID: actor,
		After:     map[string]any{"code": created.Code, "name": created.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateProductCategoryResponse{ProductCategory: catToProto(created)}), nil
}

func (s *ProductCategoryService) UpdateProductCategory(ctx context.Context, req *connect.Request[mastersv1.UpdateProductCategoryRequest]) (*connect.Response[mastersv1.UpdateProductCategoryResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	catID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category id 格式錯誤"))
	}
	if _, err := catScopeQuery(s.db.ProductCategory.Query(), cid, did).Where(productcategory.ID(catID), productcategory.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	upd := tx.ProductCategory.UpdateOneID(catID)
	if req.Msg.Code != nil {
		c := strings.TrimSpace(*req.Msg.Code)
		if c == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code 不可為空"))
		}
		upd = upd.SetCode(c)
	}
	if req.Msg.Name != nil {
		n := strings.TrimSpace(*req.Msg.Name)
		if n == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 不可為空"))
		}
		upd = upd.SetName(n)
	}
	if req.Msg.SortOrder != nil {
		upd = upd.SetSortOrder(int(*req.Msg.SortOrder))
	}
	if req.Msg.IsActive != nil {
		upd = upd.SetIsActive(*req.Msg.IsActive)
	}
	actor, _ := parseID(id.UserID)
	upd = upd.SetUpdatedBy(actor)
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "update", ResourceType: "product_category", ResourceID: strconv.FormatInt(int64(catID), 10),
		CompanyID: cid, DepartmentID: updated.DepartmentID, UserID: actor,
		After:     map[string]any{"name": updated.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateProductCategoryResponse{ProductCategory: catToProto(updated)}), nil
}

func (s *ProductCategoryService) DeleteProductCategory(ctx context.Context, req *connect.Request[mastersv1.DeleteProductCategoryRequest]) (*connect.Response[mastersv1.DeleteProductCategoryResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	catID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category id 格式錯誤"))
	}
	cur, err := catScopeQuery(s.db.ProductCategory.Query(), cid, did).Where(productcategory.ID(catID), productcategory.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.ProductCategory.UpdateOneID(catID).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "delete", ResourceType: "product_category", ResourceID: strconv.FormatInt(int64(catID), 10),
		CompanyID: cid, DepartmentID: cur.DepartmentID, UserID: actor,
		Before:    map[string]any{"code": cur.Code, "name": cur.Name},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteProductCategoryResponse{}), nil
}

func (s *ProductCategoryService) RestoreProductCategory(ctx context.Context, req *connect.Request[mastersv1.RestoreProductCategoryRequest]) (*connect.Response[mastersv1.RestoreProductCategoryResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := masterScope(id)
	if err != nil {
		return nil, err
	}
	catID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category id 格式錯誤"))
	}
	cur, err := catScopeQuery(s.db.ProductCategory.Query(), cid, did).Where(productcategory.ID(catID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		return connect.NewResponse(&mastersv1.RestoreProductCategoryResponse{ProductCategory: catToProto(cur)}), nil
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	restored, err := tx.ProductCategory.UpdateOneID(catID).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action: "update", ResourceType: "product_category", ResourceID: strconv.FormatInt(int64(catID), 10),
		CompanyID: cid, DepartmentID: restored.DepartmentID, UserID: actor,
		After:     map[string]any{"restored": true, "code": restored.Code},
		IPAddress: audit.MetaFrom(ctx).IP, UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreProductCategoryResponse{ProductCategory: catToProto(restored)}), nil
}
