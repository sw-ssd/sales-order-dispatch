// ProductService 商品主檔(04 計畫 3.3, D10/D18):商品 CRUD + 軟刪除/復原 + 分頁,
// 寫入時連同 product_units(單位換算)與 product_processing_specs(處理規格關聯)整組替換。
// 跨主檔參照(分類/倉別/處理規格)以共用 validateDeptMasterRef 驗證同部門未刪除(D33)。
package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/ent/processingspec"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productcategory"
	"github.com/salesorder/sales-order-1.0/backend/ent/productprocessingspec"
	"github.com/salesorder/sales-order-1.0/backend/ent/productunit"
	"github.com/salesorder/sales-order-1.0/backend/ent/warehouse"
	domainproducts "github.com/salesorder/sales-order-1.0/backend/internal/domain/products"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
)

// ProductService 實作 products.v1.ProductService。
type ProductService struct {
	db *ent.Client
	productsv1connect.UnimplementedProductServiceHandler
}

// NewProductService 建立 ProductService。
func NewProductService(db *ent.Client) *ProductService { return &ProductService{db: db} }

// RegisterProductService 將 ProductService 掛到 mux。
func RegisterProductService(mux *http.ServeMux, db *ent.Client) {
	path, handler := productsv1connect.NewProductServiceHandler(NewProductService(db))
	mux.Handle(path, handler)
}

// prodScopeQuery 依範圍對商品查詢加入 company/department where。
func prodScopeQuery(q *ent.ProductQuery, cid int, did *int) *ent.ProductQuery {
	if did != nil {
		return q.Where(product.CompanyIDEQ(cid), product.DepartmentIDEQ(*did))
	}
	return q.Where(product.CompanyIDEQ(cid))
}

// unitInput 為驗證後、落庫用的單位參照(rate 已解析為有理數,rateStr 為正規化十進位文字)。
type unitInput struct {
	unitCode string
	rateStr  string
	rate     *big.Rat
	isBase   bool
	sort     int32
	sizeDesc *string
}

// specInput 為驗證後、落庫用的處理規格關聯參照。
type specInput struct {
	specID     int
	attributes map[string]any
}

// ---- 單位組驗證(3.3.2 步驟 4 / 3.3.1 約定) ----
// validateUnits 驗證單位組:非空、恰一個 is_base、基本單位率 = 1、其餘率 > 0、
// unit_code 皆存在於 metadicts unit 字典(系統 + 所在部門)、unit_code 不重複。
func (s *ProductService) validateUnits(ctx context.Context, units []*productsv1.ProductUnit, did *int) ([]unitInput, error) {
	if len(units) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("商品須至少一個基本單位"))
	}
	seen := map[string]bool{}
	var baseCount int
	inputs := make([]unitInput, 0, len(units))
	for _, u := range units {
		code := strings.TrimSpace(u.GetUnitCode())
		if code == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("unit_code 必填"))
		}
		if seen[code] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("單位重複: %q", code))
		}
		seen[code] = true
		rate, err := domainproducts.ParseRate(u.GetConversionRate())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("單位 %q :%v", code, err))
		}
		if err := s.validateUnitCode(ctx, code, did); err != nil {
			return nil, err
		}
		in := unitInput{
			unitCode: code,
			rateStr:  u.GetConversionRate(),
			rate:     rate,
			isBase:   u.GetIsBase(),
			sort:     u.GetSortOrder(),
		}
		if d := strings.TrimSpace(u.GetSizeDesc()); d != "" {
			sd := d
			in.sizeDesc = &sd
		}
		if in.isBase {
			baseCount++
			if rate.Cmp(big.NewRat(1, 1)) != 0 {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("基本單位 %q 之換算率恆為 1,got %q", code, u.GetConversionRate()))
			}
		}
		inputs = append(inputs, in)
	}
	if baseCount != 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("商品須恰一個基本單位(is_base)"))
	}
	return inputs, nil
}

// validateUnitCode 驗證單位字值存在於 metadicts type=unit 字典(系統預設 + 所在部門擴充)。
func (s *ProductService) validateUnitCode(ctx context.Context, code string, did *int) error {
	q := s.db.Metadict.Query().Where(
		metadict.TypeEQ("unit"), metadict.CodeEQ(code), metadict.IsActiveEQ(true), metadict.DeletedAtIsNil(),
	)
	if did != nil {
		q = q.Where(metadict.Or(metadict.DepartmentIDIsNil(), metadict.DepartmentIDEQ(*did)))
	} else {
		q = q.Where(metadict.DepartmentIDIsNil())
	}
	ok, err := q.Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("單位字典無此值: %q", code))
	}
	return nil
}

// ---- 跨主檔參照驗證(D33,共用 validateDeptMasterRef) ----
func (s *ProductService) validateCategoryRef(ctx context.Context, idStr string, cid int, did *int) (*int, error) {
	if idStr == "" {
		return nil, nil
	}
	id, err := parseID(idStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category_id 格式錯誤"))
	}
	q := s.db.ProductCategory.Query().Where(productcategory.ID(id), productcategory.CompanyIDEQ(cid), productcategory.DeletedAtIsNil())
	if did != nil {
		q = q.Where(productcategory.DepartmentIDEQ(*did))
	}
	if err := validateDeptMasterRef(ctx, "category", q); err != nil {
		return nil, err
	}
	return &id, nil
}

func (s *ProductService) validateWarehouseRef(ctx context.Context, idStr, field string, cid int, did *int) (*int, error) {
	if idStr == "" {
		return nil, nil
	}
	id, err := parseID(idStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s 格式錯誤", field))
	}
	q := s.db.Warehouse.Query().Where(warehouse.ID(id), warehouse.CompanyIDEQ(cid), warehouse.DeletedAtIsNil())
	if did != nil {
		q = q.Where(warehouse.DepartmentIDEQ(*did))
	}
	if err := validateDeptMasterRef(ctx, field, q); err != nil {
		return nil, err
	}
	return &id, nil
}

// validateSpecRefs 驗證處理規格關聯:每規格存在、同部門、未刪除;回資化後 specInput 清單。
func (s *ProductService) validateSpecRefs(ctx context.Context, specs []*productsv1.ProductProcessingSpec, cid int, did *int) ([]specInput, error) {
	out := make([]specInput, 0, len(specs))
	seen := map[int]bool{}
	for _, sp := range specs {
		sid, err := parseID(sp.GetProcessingSpecId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("processing_spec_id 格式錯誤"))
		}
		if seen[sid] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("處理規格重複: %d", sid))
		}
		seen[sid] = true
		q := s.db.ProcessingSpec.Query().Where(processingspec.ID(sid), processingspec.CompanyIDEQ(cid), processingspec.DeletedAtIsNil())
		if did != nil {
			q = q.Where(processingspec.DepartmentIDEQ(*did))
		}
		if err := validateDeptMasterRef(ctx, "processing_spec", q); err != nil {
			return nil, err
		}
		var attrs map[string]any
		if a := sp.GetAttributes(); a != nil {
			attrs = a.AsMap()
		}
		out = append(out, specInput{specID: sid, attributes: attrs})
	}
	return out, nil
}

// ---- proto 轉換 ----
func productToProto(p *ent.Product) *productsv1.Product {
	pr := &productsv1.Product{
		Id:        strconv.FormatInt(int64(p.ID), 10),
		CompanyId: strconv.FormatInt(int64(p.CompanyID), 10),
		Code:      p.Code,
		Name:      p.Name,
		IsActive:  p.IsActive,
	}
	if p.DepartmentID != nil {
		pr.DepartmentId = strconv.FormatInt(int64(*p.DepartmentID), 10)
	}
	if p.CategoryID != nil {
		pr.CategoryId = strconv.FormatInt(int64(*p.CategoryID), 10)
	}
	if p.InventoryWarehouseID != nil {
		pr.InventoryWarehouseId = strconv.FormatInt(int64(*p.InventoryWarehouseID), 10)
	}
	if p.PickingWarehouseID != nil {
		pr.PickingWarehouseId = strconv.FormatInt(int64(*p.PickingWarehouseID), 10)
	}
	if p.Description != "" {
		pr.Description = p.Description
	}
	if !p.CreatedAt.IsZero() {
		pr.CreatedAt = p.CreatedAt.Format(time.RFC3339)
	}
	if !p.UpdatedAt.IsZero() {
		pr.UpdatedAt = p.UpdatedAt.Format(time.RFC3339)
	}
	if p.DeletedAt != nil {
		pr.DeletedAt = p.DeletedAt.Format(time.RFC3339)
	}
	return pr
}

// fetchNested 讀取商品之單位與處理規格關聯,填入 proto(Get/Create/Update 回傳用)。
func (s *ProductService) fetchNested(ctx context.Context, pr *productsv1.Product, pid int) error {
	units, err := s.db.ProductUnit.Query().Where(productunit.ProductIDEQ(pid)).Order(ent.Desc(productunit.FieldIsBase), ent.Asc(productunit.FieldSortOrder), ent.Asc(productunit.FieldUnitCode)).All(ctx)
	if err != nil {
		return toConnectError(err)
	}
	for _, u := range units {
		pu := &productsv1.ProductUnit{
			UnitCode:       u.UnitCode,
			ConversionRate: u.ConversionRate,
			IsBase:         u.IsBase,
			SortOrder:      int32(u.SortOrder),
		}
		if u.SizeDesc != "" {
			pu.SizeDesc = u.SizeDesc
		}
		pr.Units = append(pr.Units, pu)
	}
	specs, err := s.db.ProductProcessingSpec.Query().Where(productprocessingspec.ProductIDEQ(pid)).Order(ent.Asc(productprocessingspec.FieldID)).All(ctx)
	if err != nil {
		return toConnectError(err)
	}
	for _, sp := range specs {
		psp := &productsv1.ProductProcessingSpec{
			ProcessingSpecId: strconv.FormatInt(int64(sp.ProcessingSpecID), 10),
		}
		if len(sp.Attributes) > 0 {
			psp.Attributes = structpbNew(sp.Attributes)
		}
		pr.ProcessingSpecs = append(pr.ProcessingSpecs, psp)
	}
	return nil
}

// productListSource 為 pageList 的 ent 查詢橋接。
type productListSource struct{ q *ent.ProductQuery }

func (s productListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s productListSource) Page(ctx context.Context, off, lim int) ([]*ent.Product, error) {
	// F2(與 F1 同型):code 的唯一性是 (department_id, code),公司層可見範圍(跨部門)內同值 ——
	// 排序鍵非唯一時 PostgreSQL 對同值群的順序不保證一致,逐頁 LIMIT/OFFSET 會重複與遺漏資料,
	// 故以 id 為次序鍵收斂成全序。
	return s.q.Clone().Order(ent.Asc(product.FieldCode), ent.Asc(product.FieldID)).Offset(off).Limit(lim).All(ctx)
}

// ---- RPC ----
// ListProducts 分頁列出;keyword 對 code/name 模糊比對;category_id 精確篩選;可 include_deleted。
func (s *ProductService) ListProducts(ctx context.Context, req *connect.Request[productsv1.ListProductsRequest]) (*connect.Response[productsv1.ListProductsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	q := prodScopeQuery(s.db.Product.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(product.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(product.Or(product.CodeContainsFold(kw), product.NameContainsFold(kw)))
	}
	if cat := req.Msg.GetCategoryId(); cat != "" {
		cid_, err := parseID(cat)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("category_id 格式錯誤"))
		}
		q = q.Where(product.CategoryIDEQ(cid_))
	}
	list, pg, err := pageList(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), productListSource{q}, productToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&productsv1.ListProductsResponse{Products: list, Pagination: pg}), nil
}

// GetProduct 取單筆商品(含單位與處理規格清單)。
func (s *ProductService) GetProduct(ctx context.Context, req *connect.Request[productsv1.GetProductRequest]) (*connect.Response[productsv1.GetProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	pid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("product id 格式錯誤"))
	}
	row, err := prodScopeQuery(s.db.Product.Query(), cid, did).Where(product.ID(pid), product.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	pr := productToProto(row)
	if err := s.fetchNested(ctx, pr, row.ID); err != nil {
		return nil, err
	}
	return connect.NewResponse(&productsv1.GetProductResponse{Product: pr}), nil
}

// CreateProduct 建立商品 + 單位 + 處理規格關聯 + 稽核(同一交易,D18)。
func (s *ProductService) CreateProduct(ctx context.Context, req *connect.Request[productsv1.CreateProductRequest]) (*connect.Response[productsv1.CreateProductResponse], error) {
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
	units, err := s.validateUnits(ctx, req.Msg.GetUnits(), did)
	if err != nil {
		return nil, err
	}
	specs, err := s.validateSpecRefs(ctx, req.Msg.GetProcessingSpecs(), cid, did)
	if err != nil {
		return nil, err
	}
	catID, err := s.validateCategoryRef(ctx, req.Msg.GetCategoryId(), cid, did)
	if err != nil {
		return nil, err
	}
	invWH, err := s.validateWarehouseRef(ctx, req.Msg.GetInventoryWarehouseId(), "inventory_warehouse_id", cid, did)
	if err != nil {
		return nil, err
	}
	pickWH, err := s.validateWarehouseRef(ctx, req.Msg.GetPickingWarehouseId(), "picking_warehouse_id", cid, did)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	build := tx.Product.Create().
		SetCompanyID(cid).SetCode(code).SetName(name).
		SetIsActive(req.Msg.GetIsActive()).SetCreatedBy(actor).SetUpdatedBy(actor)
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if catID != nil {
		build = build.SetCategoryID(*catID)
	}
	if invWH != nil {
		build = build.SetInventoryWarehouseID(*invWH)
	}
	if pickWH != nil {
		build = build.SetPickingWarehouseID(*pickWH)
	}
	if d := strings.TrimSpace(req.Msg.GetDescription()); d != "" {
		build = build.SetDescription(d)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := createUnits(ctx, tx, created.ID, units, actor); err != nil {
		return nil, err
	}
	if err := createSpecs(ctx, tx, created.ID, specs, actor); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, tx, "product", "create", created.ID, cid, created.DepartmentID, actor, map[string]any{"code": created.Code, "name": created.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	pr := productToProto(created)
	if err := s.fetchNested(ctx, pr, created.ID); err != nil {
		return nil, err
	}
	return connect.NewResponse(&productsv1.CreateProductResponse{Product: pr}), nil
}

// UpdateProduct 欄位式更新 + units/processing_specs 整組替換(提供即替換)。
func (s *ProductService) UpdateProduct(ctx context.Context, req *connect.Request[productsv1.UpdateProductRequest]) (*connect.Response[productsv1.UpdateProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	pid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("product id 格式錯誤"))
	}
	if _, err := prodScopeQuery(s.db.Product.Query(), cid, did).Where(product.ID(pid), product.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	// 有異動欄位才重驗參照;未提供欄位沿用現值。
	var units []unitInput
	if req.Msg.Units != nil {
		units, err = s.validateUnits(ctx, req.Msg.GetUnits(), did)
		if err != nil {
			return nil, err
		}
	}
	var specs []specInput
	if req.Msg.ProcessingSpecs != nil {
		specs, err = s.validateSpecRefs(ctx, req.Msg.GetProcessingSpecs(), cid, did)
		if err != nil {
			return nil, err
		}
	}
	var catID, invWH, pickWH *int
	if req.Msg.CategoryId != nil {
		catID, err = s.validateCategoryRef(ctx, *req.Msg.CategoryId, cid, did)
		if err != nil {
			return nil, err
		}
	}
	if req.Msg.InventoryWarehouseId != nil {
		invWH, err = s.validateWarehouseRef(ctx, *req.Msg.InventoryWarehouseId, "inventory_warehouse_id", cid, did)
		if err != nil {
			return nil, err
		}
	}
	if req.Msg.PickingWarehouseId != nil {
		pickWH, err = s.validateWarehouseRef(ctx, *req.Msg.PickingWarehouseId, "picking_warehouse_id", cid, did)
		if err != nil {
			return nil, err
		}
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	upd := tx.Product.UpdateOneID(pid)
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
	if req.Msg.Description != nil {
		upd = upd.SetDescription(*req.Msg.Description)
	}
	if req.Msg.IsActive != nil {
		upd = upd.SetIsActive(*req.Msg.IsActive)
	}
	if req.Msg.CategoryId != nil {
		if catID != nil {
			upd = upd.SetCategoryID(*catID)
		} else {
			upd = upd.ClearCategoryID()
		}
	}
	if req.Msg.InventoryWarehouseId != nil {
		if invWH != nil {
			upd = upd.SetInventoryWarehouseID(*invWH)
		} else {
			upd = upd.ClearInventoryWarehouseID()
		}
	}
	if req.Msg.PickingWarehouseId != nil {
		if pickWH != nil {
			upd = upd.SetPickingWarehouseID(*pickWH)
		} else {
			upd = upd.ClearPickingWarehouseID()
		}
	}
	updated, err := upd.SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if req.Msg.Units != nil {
		if err := replaceUnits(ctx, tx, pid, units, actor); err != nil {
			return nil, err
		}
	}
	if req.Msg.ProcessingSpecs != nil {
		if err := replaceSpecs(ctx, tx, pid, specs, actor); err != nil {
			return nil, err
		}
	}
	if err := recordAudit(ctx, tx, "product", "update", pid, cid, updated.DepartmentID, actor, map[string]any{"code": updated.Code, "name": updated.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	pr := productToProto(updated)
	if err := s.fetchNested(ctx, pr, pid); err != nil {
		return nil, err
	}
	return connect.NewResponse(&productsv1.UpdateProductResponse{Product: pr}), nil
}

// DeleteProduct 軟刪除 + 稽核(保留單位與關聯列供歷史查詢)。
func (s *ProductService) DeleteProduct(ctx context.Context, req *connect.Request[productsv1.DeleteProductRequest]) (*connect.Response[productsv1.DeleteProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	pid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("product id 格式錯誤"))
	}
	cur, err := prodScopeQuery(s.db.Product.Query(), cid, did).Where(product.ID(pid), product.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	if err := tx.Product.UpdateOneID(pid).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(actor).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "product", "delete", pid, cid, cur.DepartmentID, actor, map[string]any{"code": cur.Code, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productsv1.DeleteProductResponse{}), nil
}

// RestoreProduct 復原(清 deleted_at + 稽核;已刪除才動作,冪等)。
func (s *ProductService) RestoreProduct(ctx context.Context, req *connect.Request[productsv1.RestoreProductRequest]) (*connect.Response[productsv1.RestoreProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	pid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("product id 格式錯誤"))
	}
	cur, err := prodScopeQuery(s.db.Product.Query(), cid, did).Where(product.ID(pid)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		pr := productToProto(cur)
		if err := s.fetchNested(ctx, pr, pid); err != nil {
			return nil, err
		}
		return connect.NewResponse(&productsv1.RestoreProductResponse{Product: pr}), nil
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	actor, _ := parseID(id.UserID)
	restored, err := tx.Product.UpdateOneID(pid).ClearDeletedAt().SetUpdatedBy(actor).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAudit(ctx, tx, "product", "update", pid, cid, restored.DepartmentID, actor, map[string]any{"restored": true, "code": restored.Code}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	pr := productToProto(restored)
	if err := s.fetchNested(ctx, pr, pid); err != nil {
		return nil, err
	}
	return connect.NewResponse(&productsv1.RestoreProductResponse{Product: pr}), nil
}

// ---- 交易內落庫輔助函式(單位/關聯列整組) ----
// createUnits 於交易內建立商品之單位列(不含稽核;由呼叫端統一寫一筆商品稽核)。
func createUnits(ctx context.Context, tx *ent.Tx, pid int, units []unitInput, _ int) error {
	for _, u := range units {
		b := tx.ProductUnit.Create().
			SetProductID(pid).SetUnitCode(u.unitCode).SetConversionRate(u.rateStr).
			SetIsBase(u.isBase).SetSortOrder(int(u.sort))
		if u.sizeDesc != nil {
			b = b.SetSizeDesc(*u.sizeDesc)
		}
		if _, err := b.Save(ctx); err != nil {
			return toConnectError(err)
		}
	}
	return nil
}

// createSpecs 於交易內建立商品之處理規格關聯列。
func createSpecs(ctx context.Context, tx *ent.Tx, pid int, specs []specInput, _ int) error {
	for _, sp := range specs {
		b := tx.ProductProcessingSpec.Create().SetProductID(pid).SetProcessingSpecID(sp.specID)
		if sp.attributes != nil {
			b = b.SetAttributes(sp.attributes)
		}
		if _, err := b.Save(ctx); err != nil {
			return toConnectError(err)
		}
	}
	return nil
}

// replaceUnits 整組替換:先刪除既有單位列,再重建(3.3.2 步驟 2 整組替換語意)。
func replaceUnits(ctx context.Context, tx *ent.Tx, pid int, units []unitInput, actor int) error {
	if _, err := tx.ProductUnit.Delete().Where(productunit.ProductIDEQ(pid)).Exec(ctx); err != nil {
		return toConnectError(err)
	}
	return createUnits(ctx, tx, pid, units, actor)
}

// replaceSpecs 整組替換處理規格關聯列。
func replaceSpecs(ctx context.Context, tx *ent.Tx, pid int, specs []specInput, actor int) error {
	if _, err := tx.ProductProcessingSpec.Delete().Where(productprocessingspec.ProductIDEQ(pid)).Exec(ctx); err != nil {
		return toConnectError(err)
	}
	return createSpecs(ctx, tx, pid, specs, actor)
}

// structpbNew 將 map 轉為 protobuf Struct(用於 attributes 回應)。
func structpbNew(m map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}
