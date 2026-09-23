// CustomerProductService 客戶專屬商品清單(04 計畫 Task 3.5,不存單價 D12)。
// 一客戶一商品一筆;default_qty=0 保留不顯示(3.5.2);Ensure 供下單手打儲存(3.5.3,冪等)。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	domainproducts "github.com/salesorder/sales-order-1.0/backend/internal/domain/products"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	productsv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/products/v1/productsv1connect"
)

// CustomerProductService 實作 products.v1.CustomerProductService。
type CustomerProductService struct {
	db *ent.Client
	productsv1connect.UnimplementedCustomerProductServiceHandler
}

// NewCustomerProductService 建立 CustomerProductService。
func NewCustomerProductService(db *ent.Client) *CustomerProductService {
	return &CustomerProductService{db: db}
}

// RegisterCustomerProductService 掛到 /api/v1(租戶 session + RLS)。
func RegisterCustomerProductService(mux *http.ServeMux, db *ent.Client) {
	path, handler := productsv1connect.NewCustomerProductServiceHandler(NewCustomerProductService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// cpScopeQuery 依範圍對清單查詢加入 company/department where(department 自客戶複寫)。
func cpScopeQuery(q *ent.CustomerProductQuery, cid int, did *int) *ent.CustomerProductQuery {
	if did != nil {
		return q.Where(customerproduct.CompanyIDEQ(cid), customerproduct.DepartmentIDEQ(*did))
	}
	return q.Where(customerproduct.CompanyIDEQ(cid))
}

// ListCustomerProducts 查該客戶清單(for_order=true 排除 default_qty=0 與已刪)。
func (s *CustomerProductService) ListCustomerProducts(ctx context.Context, req *connect.Request[productsv1.ListCustomerProductsRequest]) (*connect.Response[productsv1.ListCustomerProductsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetCustomerId())
	if err != nil {
		return nil, err
	}
	if _, err := customerScopeQuery(dbtenant.Client(ctx, s.db).Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	q := cpScopeQuery(dbtenant.Client(ctx, s.db).CustomerProduct.Query(), cid, did).
		Where(customerproduct.CustomerIDEQ(custID))
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(customerproduct.DeletedAtIsNil())
	}
	all, err := q.Order(ent.Asc(customerproduct.FieldAliasName)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*productsv1.CustomerProduct, 0, len(all))
	for _, cp := range all {
		if req.Msg.GetForOrder() && strings.TrimSpace(cp.DefaultQty) == "0" {
			continue
		}
		if req.Msg.GetForOrder() && cp.DefaultQty == "" {
			continue
		}
		out = append(out, customerProductToProto(cp))
	}
	return connect.NewResponse(&productsv1.ListCustomerProductsResponse{
		Products: out, Total: int32(len(out)),
	}), nil
}

// AddCustomerProduct 新增一筆(驗客戶+商品同租戶未刪;重複未刪 → already_exists)。
func (s *CustomerProductService) AddCustomerProduct(ctx context.Context, req *connect.Request[productsv1.AddCustomerProductRequest]) (*connect.Response[productsv1.AddCustomerProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	custID, err := parseID(req.Msg.GetCustomerId())
	if err != nil {
		return nil, err
	}
	pid, err := parseID(req.Msg.GetProductId())
	if err != nil {
		return nil, err
	}
	cust, err := customerScopeQuery(db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	prod, err := prodScopeQuery(db.Product.Query(), cid, did).
		Where(product.ID(pid), product.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	alias := strings.TrimSpace(req.Msg.GetAliasName())
	if alias == "" {
		alias = prod.Name
	}
	qty := strings.TrimSpace(req.Msg.GetDefaultQty())
	if qty == "" {
		qty = "0"
	}
	if _, err := domainproducts.ParseQty(qty); err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "default_qty"})
	}
	if exists, err := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.ProductIDEQ(pid),
			customerproduct.DeletedAtIsNil()).Exist(ctx); err != nil {
		return nil, toConnectError(err)
	} else if exists {
		return nil, errcode.SysConflict.Error(map[string]string{"customer_id": req.Msg.GetCustomerId()})
	}
	build := db.CustomerProduct.Create().
		SetCompanyID(cid).SetCustomerID(custID).SetProductID(pid).
		SetAliasName(alias).SetDefaultQty(qty)
	if cust.DepartmentID != nil {
		build = build.SetDepartmentID(*cust.DepartmentID)
	}
	if cut := strings.TrimSpace(req.Msg.GetCutNote()); cut != "" {
		build = build.SetCutNote(cut)
	}
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 通知觸發(4.4.4):推主責業務(無則 dept_admin);同交易建 pending + AfterCommit 發送。
	if err := OnCustomerProductCreated(ctx, db, cid, did, custID, pid, cust.Name, alias); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productsv1.AddCustomerProductResponse{Product: customerProductToProto(created)}), nil
}

// UpdateCustomerProduct 改 alias/default_qty/cut_note(不可改 customer/product)。
func (s *CustomerProductService) UpdateCustomerProduct(ctx context.Context, req *connect.Request[productsv1.UpdateCustomerProductRequest]) (*connect.Response[productsv1.UpdateCustomerProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	cpID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	cp, err := cpScopeQuery(db.CustomerProduct.Query(), cid, did).
		Where(customerproduct.ID(cpID), customerproduct.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	upd := db.CustomerProduct.UpdateOneID(cp.ID)
	if a := strings.TrimSpace(req.Msg.GetAliasName()); a != "" {
		upd = upd.SetAliasName(a)
	}
	if q := strings.TrimSpace(req.Msg.GetDefaultQty()); q != "" {
		if _, err := domainproducts.ParseQty(q); err != nil {
			return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "default_qty"})
		}
		upd = upd.SetDefaultQty(q)
	}
	if c := strings.TrimSpace(req.Msg.GetCutNote()); c != "" {
		upd = upd.SetCutNote(c)
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productsv1.UpdateCustomerProductResponse{Product: customerProductToProto(updated)}), nil
}

// DeleteCustomerProduct 軟刪除 + 同交易稽核。
func (s *CustomerProductService) DeleteCustomerProduct(ctx context.Context, req *connect.Request[productsv1.DeleteCustomerProductRequest]) (*connect.Response[productsv1.DeleteCustomerProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	cpID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	cp, err := cpScopeQuery(db.CustomerProduct.Query(), cid, did).
		Where(customerproduct.ID(cpID), customerproduct.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if _, err := db.CustomerProduct.UpdateOneID(cp.ID).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAuditBA(ctx, tx, "customer_product", "delete", cp.ID, cid, did, actorIDOf(id),
		map[string]any{"alias": cp.AliasName}, map[string]any{"deleted": true}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productsv1.DeleteCustomerProductResponse{}), nil
}

// EnsureCustomerProduct 下單手打確認儲存後呼叫(3.5.3,冪等):存在回既有 created=false;
// 唯一衝突吸收改讀既有;軟刪除過的不復活(新建一筆);商品已刪則拒絕。
func (s *CustomerProductService) EnsureCustomerProduct(ctx context.Context, req *connect.Request[productsv1.EnsureCustomerProductRequest]) (*connect.Response[productsv1.EnsureCustomerProductResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	custID, err := parseID(req.Msg.GetCustomerId())
	if err != nil {
		return nil, err
	}
	pid, err := parseID(req.Msg.GetProductId())
	if err != nil {
		return nil, err
	}
	alias := strings.TrimSpace(req.Msg.GetAliasName())
	if alias == "" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "alias_name"})
	}
	cust, err := customerScopeQuery(db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if _, err := prodScopeQuery(db.Product.Query(), cid, did).
		Where(product.ID(pid), product.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if existing, err := db.CustomerProduct.Query().
		Where(customerproduct.CustomerIDEQ(custID), customerproduct.ProductIDEQ(pid),
			customerproduct.DeletedAtIsNil()).Only(ctx); err == nil {
		return connect.NewResponse(&productsv1.EnsureCustomerProductResponse{
			Product: customerProductToProto(existing), Created: false,
		}), nil
	} else if !ent.IsNotFound(err) {
		return nil, toConnectError(err)
	}
	build := db.CustomerProduct.Create().
		SetCompanyID(cid).SetCustomerID(custID).SetProductID(pid).
		SetAliasName(alias).SetDefaultQty("0")
	if cust.DepartmentID != nil {
		build = build.SetDepartmentID(*cust.DepartmentID)
	}
	created, err := build.Save(ctx)
	if err != nil {
		// 併發唯一衝突 → 吸收,改讀既有(冪等收斂)。
		if ent.IsConstraintError(err) {
			if existing, rerr := db.CustomerProduct.Query().
				Where(customerproduct.CustomerIDEQ(custID), customerproduct.ProductIDEQ(pid),
					customerproduct.DeletedAtIsNil()).Only(ctx); rerr == nil {
				return connect.NewResponse(&productsv1.EnsureCustomerProductResponse{
					Product: customerProductToProto(existing), Created: false,
				}), nil
			}
		}
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productsv1.EnsureCustomerProductResponse{
		Product: customerProductToProto(created), Created: true,
	}), nil
}

// customerProductToProto 轉 proto(id 全轉字串)。
func customerProductToProto(cp *ent.CustomerProduct) *productsv1.CustomerProduct {
	p := &productsv1.CustomerProduct{
		Id:         strconv.FormatInt(int64(cp.ID), 10),
		CustomerId: strconv.FormatInt(int64(cp.CustomerID), 10),
		ProductId:  strconv.FormatInt(int64(cp.ProductID), 10),
		AliasName:  cp.AliasName,
		DefaultQty: cp.DefaultQty,
		CreatedAt:  cp.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  cp.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if cp.CutNote != "" {
		p.CutNote = cp.CutNote
	}
	for _, t := range cp.PromoTagIds {
		p.PromoTagIds = append(p.PromoTagIds, int64(t))
	}
	return p
}
