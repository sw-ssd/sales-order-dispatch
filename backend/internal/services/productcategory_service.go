// ProductCategoryService 商品分類(04 計畫 3.4.4, D10/D18):部門級單層分類 CRUD + 軟刪除/復原。
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

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1/mastersv1connect"
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
	path, handler := mastersv1connect.NewProductCategoryServiceHandler(NewProductCategoryService(db), dbtenant.HandlerOption(db))
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

type catListSource struct{ q *ent.ProductCategoryQuery }

func (s catListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s catListSource) Page(ctx context.Context, off, lim int) ([]*ent.ProductCategory, error) {
	// F2(與 F1 同型):sort_order 預設 0、同值群常遠大於一頁,排序鍵非唯一時 PostgreSQL 對
	// 同值群(ties)的順序不保證一致,逐頁 LIMIT/OFFSET 會重複與遺漏資料,故以 id 為次序鍵。
	return s.q.Clone().Order(ent.Asc(productcategory.FieldSortOrder), ent.Asc(productcategory.FieldID)).Offset(off).Limit(lim).All(ctx)
}

func (s *ProductCategoryService) ListProductCategories(ctx context.Context, req *connect.Request[mastersv1.ListProductCategoriesRequest]) (*connect.Response[mastersv1.ListProductCategoriesResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	q := catScopeQuery(dbtenant.Client(ctx, s.db).ProductCategory.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(productcategory.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(productcategory.Or(productcategory.CodeContainsFold(kw), productcategory.NameContainsFold(kw)))
	}
	list, pg, err := pageList(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), catListSource{q}, catToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mastersv1.ListProductCategoriesResponse{ProductCategories: list, Pagination: pg}), nil
}

func (s *ProductCategoryService) CreateProductCategory(ctx context.Context, req *connect.Request[mastersv1.CreateProductCategoryRequest]) (*connect.Response[mastersv1.CreateProductCategoryResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	code, name, err := codeName(req.Msg.GetCode(), req.Msg.GetName())
	if err != nil {
		return nil, err
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	build := db.ProductCategory.Create().
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
	if err := recordAudit(ctx, tx, "product_category", "create", created.ID, cid, created.DepartmentID, actor, map[string]any{"code": created.Code, "name": created.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.CreateProductCategoryResponse{ProductCategory: catToProto(created)}), nil
}

func (s *ProductCategoryService) UpdateProductCategory(ctx context.Context, req *connect.Request[mastersv1.UpdateProductCategoryRequest]) (*connect.Response[mastersv1.UpdateProductCategoryResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	catID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category id 格式錯誤"))
	}
	if _, err := catScopeQuery(dbtenant.Client(ctx, s.db).ProductCategory.Query(), cid, did).Where(productcategory.ID(catID), productcategory.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	upd := db.ProductCategory.UpdateOneID(catID)
	if req.Msg.Code != nil {
		c, err := trimNonEmpty(*req.Msg.Code, "code 不可為空")
		if err != nil {
			return nil, err
		}
		upd = upd.SetCode(c)
	}
	if req.Msg.Name != nil {
		n, err := trimNonEmpty(*req.Msg.Name, "name 不可為空")
		if err != nil {
			return nil, err
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
	updated, err := upd.SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "product_category", "update", catID, cid, updated.DepartmentID, actor, map[string]any{"name": updated.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.UpdateProductCategoryResponse{ProductCategory: catToProto(updated)}), nil
}

func (s *ProductCategoryService) DeleteProductCategory(ctx context.Context, req *connect.Request[mastersv1.DeleteProductCategoryRequest]) (*connect.Response[mastersv1.DeleteProductCategoryResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	catID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category id 格式錯誤"))
	}
	cur, err := catScopeQuery(dbtenant.Client(ctx, s.db).ProductCategory.Query(), cid, did).Where(productcategory.ID(catID), productcategory.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	if err := db.ProductCategory.UpdateOneID(catID).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "product_category", "delete", catID, cid, cur.DepartmentID, actor, map[string]any{"code": cur.Code, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.DeleteProductCategoryResponse{}), nil
}

func (s *ProductCategoryService) RestoreProductCategory(ctx context.Context, req *connect.Request[mastersv1.RestoreProductCategoryRequest]) (*connect.Response[mastersv1.RestoreProductCategoryResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	catID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category id 格式錯誤"))
	}
	cur, err := catScopeQuery(dbtenant.Client(ctx, s.db).ProductCategory.Query(), cid, did).Where(productcategory.ID(catID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		return connect.NewResponse(&mastersv1.RestoreProductCategoryResponse{ProductCategory: catToProto(cur)}), nil
	}
	// 取請求交易:查詢/寫入用 db,稽核續用 tx(同一交易,D18);理由見 customer_service.CreateCustomer:
	// 主檔域 ENABLE+FORCE 後,自開交易(池化連線、未帶 scope)會被 policy 過濾成 0 列/WITH CHECK 擋下。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
	}
	db := tx.Client()
	actor, _ := parseID(id.UserID)
	restored, err := db.ProductCategory.UpdateOneID(catID).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "product_category", "update", catID, cid, restored.DepartmentID, actor, map[string]any{"restored": true, "code": restored.Code}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&mastersv1.RestoreProductCategoryResponse{ProductCategory: catToProto(restored)}), nil
}
