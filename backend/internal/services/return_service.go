// ReturnService 退貨申請與審核(06 計畫 Task 4.7.2–4.7.3, D25)。
// 發起僅客戶子帳號(主帳號一律拒絕);審核僅主責業務/dept_admin 以上;
// 全程不修改原訂單(僅參照);寫入同交易 + audit.Recorder(D18)。
package services

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/productunit"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequestitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/ent/salesorderitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/products"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// ReturnService 為退貨服務。
type ReturnService struct {
	db *ent.Client
}

// NewReturnService 建立 ReturnService。
func NewReturnService(db *ent.Client) *ReturnService {
	return &ReturnService{db: db}
}

// RegisterReturnService 掛到 /api/v1(租戶 session + RLS)。
func RegisterReturnService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewReturnServiceHandler(NewReturnService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// returnCustomerScope 解析客戶子帳號範圍(回 company/department/customer/user)。
// 主帳號/員工/訪客一律拒絕(Create/List/Get 皆同)。
func returnCustomerScope(id authz.Identity) (cid, did, custID, actor int, err error) {
	if id.Role != "customer" || strings.TrimSpace(id.CustomerID) == "" {
		return 0, 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	cid, err = parseID(id.CompanyID)
	if err != nil {
		return 0, 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	did, err = parseID(id.DepartmentID)
	if err != nil {
		return 0, 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	custID, err = parseID(id.CustomerID)
	if err != nil {
		return 0, 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	actor, err = parseID(id.UserID)
	if err != nil {
		return 0, 0, 0, 0, errcode.SysPermissionDenied.Error(nil)
	}
	// 主帳號拒絕:users.is_primary 由服務層核實(測試身分無該欄時以 DB 為準)。
	return cid, did, custID, actor, nil
}

// resolvedItem 為核實後的品項(來源 + 快照)。
type resolvedItem struct {
	sourceType      string
	salesOrderID    *int
	salesOrderItem  *int
	customerProduct *int
	productID       int
	productName     string
	spec            string
	unit            string
	qty             string
	reason          string
	photoIDs        []string
}

// CreateReturnRequest 發起退貨(子帳號;雙來源並存;同交易寫申請+明細+稽核)。
func (s *ReturnService) CreateReturnRequest(ctx context.Context, req *connect.Request[salesorderv1.CreateReturnRequestRequest]) (*connect.Response[salesorderv1.CreateReturnRequestResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, custID, actor, err := returnCustomerScope(id)
	if err != nil {
		return nil, err
	}
	if len(req.Msg.GetItems()) == 0 {
		return nil, invalidArgField("items")
	}
	// 主帳號拒絕(DB 核實 is_primary)。
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	u, err := db.User.Query().Where(user.IDEQ(actor)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if u.IsPrimary {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	var resolved []resolvedItem
	for _, it := range req.Msg.GetItems() {
		r, err := s.resolveItem(ctx, db, cid, did, custID, it)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, r)
	}
	rr, err := db.ReturnRequest.Create().
		SetCompanyID(cid).SetDepartmentID(did).SetCustomerID(custID).
		SetCreatedByUserID(actor).SetStatus("pending").Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if rm := strings.TrimSpace(req.Msg.GetRemark()); rm != "" {
		if _, err := db.ReturnRequest.UpdateOneID(rr.ID).SetRemark(rm).Save(ctx); err != nil {
			return nil, toConnectError(err)
		}
	}
	for _, r := range resolved {
		b := db.ReturnRequestItem.Create().
			SetReturnRequestID(rr.ID).SetCompanyID(cid).SetDepartmentID(did).
			SetSourceType(r.sourceType).SetProductID(r.productID).SetProductName(r.productName).
			SetUnit(r.unit).SetQuantity(r.qty).SetReason(r.reason)
		if r.salesOrderID != nil {
			b = b.SetSalesOrderID(*r.salesOrderID)
		}
		if r.salesOrderItem != nil {
			b = b.SetSalesOrderItemID(*r.salesOrderItem)
		}
		if r.customerProduct != nil {
			b = b.SetCustomerProductID(*r.customerProduct)
		}
		if r.spec != "" {
			b = b.SetSpec(r.spec)
		}
		if len(r.photoIDs) > 0 {
			b = b.SetPhotoFileIds(r.photoIDs)
		}
		if _, err := b.Save(ctx); err != nil {
			return nil, toConnectError(err)
		}
	}
	if err := recordAudit(ctx, tx, "return_request", "create", rr.ID, cid, &did, actor,
		map[string]any{"customer_id": custID, "items": len(resolved)}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.CreateReturnRequestResponse{
		Id: strconv.Itoa(rr.ID), Status: "pending",
	}), nil
}

// resolveItem 核實單品項(來源配對 + 歸屬 + 數量 + 照片)。
func (s *ReturnService) resolveItem(ctx context.Context, db *ent.Client, cid, did, custID int, it *salesorderv1.ReturnItemInput) (resolvedItem, error) {
	var r resolvedItem
	st := strings.TrimSpace(it.GetSourceType())
	if st != "order_item" && st != "customer_product" {
		return r, invalidArgField("source_type")
	}
	r.sourceType = st
	q, err := products.ParseQty(it.GetQuantity())
	if err != nil || q.Sign() <= 0 {
		return r, invalidArgField("quantity")
	}
	r.qty = it.GetQuantity()
	if strings.TrimSpace(it.GetReason()) == "" {
		return r, invalidArgField("reason")
	}
	r.reason = strings.TrimSpace(it.GetReason())
	if len(it.GetPhotoFileIds()) > 5 {
		return r, invalidArgField("photo_file_ids")
	}
	for _, fid := range it.GetPhotoFileIds() {
		fidN, err := parseID(fid)
		if err != nil {
			return r, invalidArgField("photo_file_ids")
		}
		fa, err := db.FileAsset.Query().Where(fileasset.IDEQ(fidN)).Only(ctx)
		if err != nil || fa.CompanyID != cid || fa.DeletedAt != nil {
			return r, errcode.SysNotFound.Error(nil)
		}
		r.photoIDs = append(r.photoIDs, fid)
	}
	switch st {
	case "order_item":
		oiN, err := parseID(it.GetSalesOrderItemId())
		if err != nil {
			return r, invalidArgField("sales_order_item_id")
		}
		oi, err := db.SalesOrderItem.Query().
			Where(salesorderitem.ID(oiN), salesorderitem.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			return r, errcode.SysNotFound.Error(nil)
		}
		o, err := db.SalesOrder.Query().Where(salesorder.IDEQ(oi.SalesOrderID)).Only(ctx)
		if err != nil || o.CustomerID != custID || o.CompanyID != cid || o.DeletedAt != nil {
			return r, errcode.SysNotFound.Error(nil)
		}
		r.salesOrderID = &oi.SalesOrderID
		r.salesOrderItem = &oiN
		if oi.ProductID != nil {
			r.productID = *oi.ProductID
		}
		r.productName = oi.DisplayName
		r.unit = oi.Unit
	case "customer_product":
		cpN, err := parseID(it.GetCustomerProductId())
		if err != nil {
			return r, invalidArgField("customer_product_id")
		}
		cp, err := db.CustomerProduct.Query().
			Where(customerproduct.ID(cpN), customerproduct.DeletedAtIsNil()).Only(ctx)
		if err != nil || cp.CustomerID != custID || cp.CompanyID != cid {
			return r, errcode.SysNotFound.Error(nil)
		}
		r.customerProduct = &cpN
		r.productID = cp.ProductID
		r.productName = strings.TrimSpace(cp.AliasName)
		if r.productName == "" {
			if p, err := db.Product.Query().Where(product.IDEQ(cp.ProductID)).Only(ctx); err == nil {
				r.productName = p.Name
			} else {
				r.productName = "（已刪除商品）"
			}
		}
		r.unit = baseUnitOf(ctx, db, cp.ProductID)
	}
	return r, nil
}

// ListReturnRequests 申請列表(客戶 self 限自己客戶;staff 依範圍)。
func (s *ReturnService) ListReturnRequests(ctx context.Context, req *connect.Request[salesorderv1.ListReturnRequestsRequest]) (*connect.Response[salesorderv1.ListReturnRequestsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := db.ReturnRequest.Query().Where(returnrequest.DeletedAtIsNil())
	if id.Role == "customer" {
		if strings.TrimSpace(id.CustomerID) == "" {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		custID, err := parseID(id.CustomerID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		q = q.Where(returnrequest.CompanyIDEQ(cid), returnrequest.CustomerIDEQ(custID))
	} else {
		cid, did, err := deptScope(id)
		if err != nil {
			return nil, err
		}
		q = q.Where(returnrequest.CompanyIDEQ(cid))
		if did != nil {
			q = q.Where(returnrequest.DepartmentIDEQ(*did))
		}
	}
	if st := strings.TrimSpace(req.Msg.GetStatus()); st != "" {
		if st != "pending" && st != "approved" && st != "rejected" {
			return nil, invalidArgField("status")
		}
		q = q.Where(returnrequest.StatusEQ(st))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	rows, err := q.Order(ent.Desc(returnrequest.FieldCreatedAt), ent.Desc(returnrequest.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	resp := &salesorderv1.ListReturnRequestsResponse{
		Page: int32(page), PageSize: int32(pageSize), Total: int32(total),
	}
	for _, r := range rows {
		resp.Entries = append(resp.Entries, &salesorderv1.ReturnRequestEntry{
			Id: strconv.Itoa(r.ID), CustomerId: strconv.Itoa(r.CustomerID),
			Status: r.Status, Remark: r.Remark,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return connect.NewResponse(resp), nil
}

// GetReturnRequest 單筆含品項明細(照片轉下載 URL;跨範圍視同 not_found)。
func (s *ReturnService) GetReturnRequest(ctx context.Context, req *connect.Request[salesorderv1.GetReturnRequestRequest]) (*connect.Response[salesorderv1.GetReturnRequestResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, invalidArgField("id")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	q := db.ReturnRequest.Query().Where(returnrequest.ID(rid), returnrequest.DeletedAtIsNil())
	if id.Role == "customer" {
		if strings.TrimSpace(id.CustomerID) == "" {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		custID, err := parseID(id.CustomerID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		q = q.Where(returnrequest.CompanyIDEQ(cid), returnrequest.CustomerIDEQ(custID))
	} else {
		cid, did, err := deptScope(id)
		if err != nil {
			return nil, err
		}
		q = q.Where(returnrequest.CompanyIDEQ(cid))
		if did != nil {
			q = q.Where(returnrequest.DepartmentIDEQ(*did))
		}
	}
	rr, err := q.Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := db.ReturnRequestItem.Query().
		Where(returnrequestitem.ReturnRequestIDEQ(rr.ID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	resp := &salesorderv1.GetReturnRequestResponse{
		Id: strconv.Itoa(rr.ID), CustomerId: strconv.Itoa(rr.CustomerID),
		Status: rr.Status, Remark: rr.Remark, RejectReason: rr.RejectReason,
		CreatedAt: rr.CreatedAt.Format(time.RFC3339),
		// 樂觀鎖:ReviewReturnRequest 必帶 expected_version,讀取端須能拿到當前版本。
		Version: strconv.Itoa(rr.Version),
	}
	for _, it := range items {
		v := &salesorderv1.ReturnRequestItemView{
			Id: strconv.Itoa(it.ID), SourceType: it.SourceType,
			ProductName: it.ProductName, Spec: it.Spec, Unit: it.Unit,
			Quantity: it.Quantity, Reason: it.Reason,
		}
		for _, fid := range it.PhotoFileIds {
			fidN, err := parseID(fid)
			if err != nil {
				continue
			}
			if fa, err := db.FileAsset.Query().Where(fileasset.IDEQ(fidN)).Only(ctx); err == nil {
				v.PhotoUrls = append(v.PhotoUrls, fa.URL)
			}
		}
		resp.Items = append(resp.Items, v)
	}
	return connect.NewResponse(resp), nil
}

// invalidArgField 回 invalid_argument(欄位級)。
func invalidArgField(f string) error {
	return errcode.SysInvalidArgument.Error(map[string]string{"field": f})
}

// baseUnitOf 取商品 base 單位(查無回空,不擋退貨)。
func baseUnitOf(ctx context.Context, db *ent.Client, pid int) string {
	u, err := db.ProductUnit.Query().
		Where(productunit.ProductIDEQ(pid), productunit.IsBaseEQ(true)).Only(ctx)
	if err != nil {
		return ""
	}
	return u.UnitCode
}
