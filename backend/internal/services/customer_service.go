// CustomerService 客戶主檔(04 計畫 3.1, D7/D10/D18):CRUD + 軟刪除/復原 + 關鍵字篩選 +
// customer_code 取號(樂觀鎖 counter) + 字典/業務 reference 驗證。
// 範圍:super 全域 / company_admin 公司 / dept_admin·staff 自己部門;customer 主帳號一律 permission_denied。
package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/customercounter"
	"github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// businessRepRoles 為可作為 default_sales_rep 的角色(業務/管理,非客戶主帳號)。
var businessRepRoles = map[string]bool{
	"super": true, "company_admin": true, "dept_admin": true, "staff": true,
}

// maxCustomerCodeRetries 取號樂觀鎖重試上限(D7, 3.1.3)。
const maxCustomerCodeRetries = 5

// CustomerService 實作 customers.v1.CustomerService。
type CustomerService struct {
	db *ent.Client
	customersv1connect.UnimplementedCustomerServiceHandler
}

// NewCustomerService 建立 CustomerService。
func NewCustomerService(db *ent.Client) *CustomerService {
	return &CustomerService{db: db}
}

// RegisterCustomerServices 將 CustomerService 掛到 mux。
func RegisterCustomerServices(mux *http.ServeMux, db *ent.Client) {
	path, handler := customersv1connect.NewCustomerServiceHandler(NewCustomerService(db))
	mux.Handle(path, handler)
}

// hasRole 判斷身分是否含指定角色。
func hasRole(id authz.Identity, role string) bool {
	return slices.Contains(id.Roles, role)
}

// customerScope 依身分推導可見/操作範圍,回傳(companyID, departmentID *int)。
// 超範圍/無權 → 回 permission_denied。
func customerScope(id authz.Identity) (int, *int, error) {
	switch {
	case isSuperIdentity(id), hasRole(id, "company_admin"):
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
		}
		return cid, nil, nil
	case hasRole(id, "dept_admin"), hasRole(id, "staff"):
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
		}
		did, err := parseID(id.DepartmentID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少部門範圍"))
		}
		return cid, &did, nil
	default:
		// customer 主帳號 / guest 一律拒絕。
		return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("無客戶主檔權限"))
	}
}

// customerScopeQuery 依範圍對查詢加入 company/department where。
func customerScopeQuery(q *ent.CustomerQuery, cid int, did *int) *ent.CustomerQuery {
	if did != nil {
		return q.Where(customer.DepartmentIDEQ(*did))
	}
	return q.Where(customer.CompanyIDEQ(cid))
}

// customerToProto 將 ent.Customer 轉為 proto Customer。
func customerToProto(c *ent.Customer) *customersv1.Customer {
	p := &customersv1.Customer{
		Id:                    strconv.FormatInt(int64(c.ID), 10),
		CompanyId:             strconv.FormatInt(int64(c.CompanyID), 10),
		CustomerCode:          c.CustomerCode,
		Name:                  c.Name,
		TaxId:                 c.TaxID,
		PreferredDeliveryDays: c.PreferredDeliveryDays,
	}
	if c.DepartmentID != nil {
		p.DepartmentId = strconv.FormatInt(int64(*c.DepartmentID), 10)
	}
	if c.PaymentMethodID != nil {
		p.PaymentMethodId = strconv.FormatInt(int64(*c.PaymentMethodID), 10)
	}
	if c.SettlementMethodID != nil {
		p.SettlementMethodId = strconv.FormatInt(int64(*c.SettlementMethodID), 10)
	}
	if c.CustomerTypeID != nil {
		p.CustomerTypeId = strconv.FormatInt(int64(*c.CustomerTypeID), 10)
	}
	if c.InvoiceTypeID != nil {
		p.InvoiceTypeId = strconv.FormatInt(int64(*c.InvoiceTypeID), 10)
	}
	if c.DefaultSalesRepID != nil {
		p.DefaultSalesRepId = strconv.FormatInt(int64(*c.DefaultSalesRepID), 10)
	}
	for _, id := range c.PromoTagIds {
		p.PromoTagIds = append(p.PromoTagIds, int64(id))
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

// validateMetadictRef 驗證字典參考:存在於 metadicts、type 相符、is_active、未刪除、操作者範圍內。
// id 為空字串時回 0,nil(可空欄位未提供)。
func (s *CustomerService) validateMetadictRef(ctx context.Context, actorID authz.Identity, id, typ string) (int, error) {
	if id == "" {
		return 0, nil
	}
	mid, err := parseID(id)
	if err != nil {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("字典欄位 %s 格式錯誤", typ))
	}
	q := s.db.Metadict.Query().Where(
		metadict.ID(mid), metadict.TypeEQ(typ), metadict.IsActiveEQ(true), metadict.DeletedAtIsNil(),
	)
	scope, err := metadictScope(actorID)
	if err != nil {
		return 0, err
	}
	ok, err := scope(q).Exist(ctx)
	if err != nil {
		return 0, toConnectError(err)
	}
	if !ok {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("字典參考無效(type %q id %d)", typ, mid))
	}
	return mid, nil
}

// validateSalesRep 驗證 default_sales_rep:同公司、業務角色、active;客戶有部門時需同部門。
func (s *CustomerService) validateSalesRep(ctx context.Context, uid, cid int, did *int) (int, error) {
	u, err := s.db.User.Query().Where(user.ID(uid)).WithCompany().WithDepartment().Only(ctx)
	if err != nil {
		return 0, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 無效"))
	}
	if u.Status != user.StatusActive || !businessRepRoles[u.Role] {
		return 0, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 須為同公司有效業務帳號"))
	}
	if u.Edges.Company == nil || u.Edges.Company.ID != cid {
		return 0, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 須為同公司業務"))
	}
	if did != nil && (u.Edges.Department == nil || u.Edges.Department.ID != *did) {
		return 0, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 須為同部門業務"))
	}
	return uid, nil
}

// nextCustomerCode 於交易內以樂觀鎖 counter 取號回傳 customer_code(公司前綴 + 6 位補零)。
// version 衝突則重試更新(上限 maxCustomerCodeRetries);逾限回 failed_precondition。
func nextCustomerCode(ctx context.Context, tx *ent.Tx, cid int, prefix string) (string, error) {
	for i := 0; i < maxCustomerCodeRetries; i++ {
		c, err := tx.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(cid)).Only(ctx)
		switch {
		case ent.IsNotFound(err):
			// 無 counter:插入初始列(next_seq=1, version=0),本次序號 1 並推進。
			if _, err := tx.CustomerCounter.Create().SetCompanyID(cid).SetNextSeq(1).SetVersion(0).Save(ctx); err != nil {
				if ent.IsConstraintError(err) {
					continue // 併發插入衝突 → 重試
				}
				return "", err
			}
			n, err := tx.CustomerCounter.Update().
				Where(customercounter.CompanyIDEQ(cid), customercounter.VersionEQ(0)).
				SetNextSeq(2).SetVersion(1).Save(ctx)
			if err != nil {
				return "", err
			}
			if n == 0 {
				continue // version 衝突 → 重試
			}
			return fmt.Sprintf("%s%06d", prefix, 1), nil
		case err != nil:
			return "", err
		default:
			n, err := tx.CustomerCounter.Update().
				Where(customercounter.CompanyIDEQ(cid), customercounter.VersionEQ(c.Version)).
				SetNextSeq(c.NextSeq + 1).SetVersion(c.Version + 1).Save(ctx)
			if err != nil {
				return "", err
			}
			if n == 0 {
				continue // version 衝突 → 重試
			}
			return fmt.Sprintf("%s%06d", prefix, c.NextSeq), nil
		}
	}
	return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("客戶編號取號衝突,請稍後重試"))
}

// ListCustomers 分頁查詢,支持關鍵字(名稱/編號/統編)模糊比對與 include_deleted、排序白名單。
func (s *CustomerService) ListCustomers(ctx context.Context, req *connect.Request[customersv1.ListCustomersRequest]) (*connect.Response[customersv1.ListCustomersResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := customerScopeQuery(s.db.Customer.Query(), cid, did)
	if !req.Msg.GetIncludeDeleted() {
		q = q.Where(customer.DeletedAtIsNil())
	}
	if kw := strings.TrimSpace(req.Msg.GetKeyword()); kw != "" {
		q = q.Where(customer.Or(
			customer.NameContainsFold(kw),
			customer.CustomerCodeContainsFold(kw),
			customer.TaxIDContainsFold(kw),
		))
	}
	field, err := customerSortField(req.Msg.GetSort())
	if err != nil {
		return nil, err
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Clone().Order(ent.Asc(field)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	out := make([]*customersv1.Customer, 0, len(items))
	for _, c := range items {
		out = append(out, customerToProto(c))
	}
	return connect.NewResponse(&customersv1.ListCustomersResponse{
		Customers:  out,
		Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
}

// customerSortField 將排序參數對應到白名單欄位;預設 name。
func customerSortField(sort string) (string, error) {
	switch sort {
	case "", "name":
		return customer.FieldName, nil
	case "customer_code":
		return customer.FieldCustomerCode, nil
	case "created_at":
		return customer.FieldCreatedAt, nil
	default:
		return "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("排序欄位 %q 不在白名單(name/customer_code/created_at)", sort))
	}
}

// GetCustomer 以 id 取單筆(限可見範圍;已刪除/不存在 → not_found)。
func (s *CustomerService) GetCustomer(ctx context.Context, req *connect.Request[customersv1.GetCustomerRequest]) (*connect.Response[customersv1.GetCustomerResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	c, err := customerScopeQuery(s.db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.GetCustomerResponse{Customer: customerToProto(c)}), nil
}

// CreateCustomer 建立客戶:取號 + 驗證字典/業務 + 建檔 + 稽核同一交易(D18)。
func (s *CustomerService) CreateCustomer(ctx context.Context, req *connect.Request[customersv1.CreateCustomerRequest]) (*connect.Response[customersv1.CreateCustomerResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 必填"))
	}
	// 字典/業務驗證(交易外先驗,減少 tx 內失敗)。
	payID, err := s.validateMetadictRef(ctx, id, req.Msg.GetPaymentMethodId(), "payment_method")
	if err != nil {
		return nil, err
	}
	settleID, err := s.validateMetadictRef(ctx, id, req.Msg.GetSettlementMethodId(), "settlement_method")
	if err != nil {
		return nil, err
	}
	custTypeID, err := s.validateMetadictRef(ctx, id, req.Msg.GetCustomerTypeId(), "customer_type")
	if err != nil {
		return nil, err
	}
	invTypeID, err := s.validateMetadictRef(ctx, id, req.Msg.GetInvoiceTypeId(), "invoice_type")
	if err != nil {
		return nil, err
	}
	var repID int
	if req.Msg.GetDefaultSalesRepId() != "" {
		ruid, err := parseID(req.Msg.GetDefaultSalesRepId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 格式錯誤"))
		}
		repID, err = s.validateSalesRep(ctx, ruid, cid, did)
		if err != nil {
			return nil, err
		}
	}
	// 公司前綴(customer_code 取號必要)。
	co, err := s.db.Company.Get(ctx, cid)
	if err != nil {
		return nil, toConnectError(err)
	}
	if co.Status != company.StatusActive {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("公司已停用,無法建檔"))
	}
	prefix := co.CustomerCodePrefix
	if prefix == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("公司未設定客戶編號前綴"))
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := nextCustomerCode(ctx, tx, cid, prefix)
	if err != nil {
		return nil, toConnectError(err)
	}
	build := tx.Customer.Create().
		SetCompanyID(cid).
		SetCustomerCode(code).
		SetName(name).
		SetPreferredDeliveryDays(req.Msg.GetPreferredDeliveryDays()).
		SetPromoTagIds(intSlice(req.Msg.GetPromoTagIds()))
	if did != nil {
		build = build.SetDepartmentID(*did)
	}
	if t := req.Msg.GetTaxId(); t != "" {
		build = build.SetTaxID(t)
	}
	if payID != 0 {
		build = build.SetPaymentMethodID(payID)
	}
	if settleID != 0 {
		build = build.SetSettlementMethodID(settleID)
	}
	if custTypeID != 0 {
		build = build.SetCustomerTypeID(custTypeID)
	}
	if invTypeID != 0 {
		build = build.SetInvoiceTypeID(invTypeID)
	}
	if repID != 0 {
		build = build.SetDefaultSalesRepID(repID)
	}
	actor, _ := parseID(id.UserID)
	build = build.SetCreatedBy(actor).SetUpdatedBy(actor)
	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "create",
		ResourceType: "customer",
		ResourceID:   strconv.FormatInt(int64(created.ID), 10),
		CompanyID:    cid,
		DepartmentID: created.DepartmentID,
		UserID:       actor,
		After: map[string]any{
			"customer_code": created.CustomerCode, "name": created.Name,
		},
		IPAddress: audit.MetaFrom(ctx).IP,
		UserAgent: audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.CreateCustomerResponse{Customer: customerToProto(created)}), nil
}

// UpdateCustomer 欄位式更新(customer_code 不可改;update 請求無 code 欄位,天然拒絕)。
func (s *CustomerService) UpdateCustomer(ctx context.Context, req *connect.Request[customersv1.UpdateCustomerRequest]) (*connect.Response[customersv1.UpdateCustomerResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	cur, err := customerScopeQuery(s.db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	upd := tx.Customer.UpdateOneID(custID)
	if req.Msg.Name != nil {
		n := strings.TrimSpace(*req.Msg.Name)
		if n == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name 不可為空"))
		}
		upd = upd.SetName(n)
	}
	if req.Msg.TaxId != nil {
		upd = upd.SetTaxID(*req.Msg.TaxId)
	}
	if req.Msg.PaymentMethodId != nil {
		if *req.Msg.PaymentMethodId == "" {
			upd = upd.ClearPaymentMethodID()
		} else {
			v, err := s.validateMetadictRef(ctx, id, *req.Msg.PaymentMethodId, "payment_method")
			if err != nil {
				return nil, err
			}
			upd = upd.SetPaymentMethodID(v)
		}
	}
	if req.Msg.SettlementMethodId != nil {
		if *req.Msg.SettlementMethodId == "" {
			upd = upd.ClearSettlementMethodID()
		} else {
			v, err := s.validateMetadictRef(ctx, id, *req.Msg.SettlementMethodId, "settlement_method")
			if err != nil {
				return nil, err
			}
			upd = upd.SetSettlementMethodID(v)
		}
	}
	if req.Msg.CustomerTypeId != nil {
		if *req.Msg.CustomerTypeId == "" {
			upd = upd.ClearCustomerTypeID()
		} else {
			v, err := s.validateMetadictRef(ctx, id, *req.Msg.CustomerTypeId, "customer_type")
			if err != nil {
				return nil, err
			}
			upd = upd.SetCustomerTypeID(v)
		}
	}
	if req.Msg.InvoiceTypeId != nil {
		if *req.Msg.InvoiceTypeId == "" {
			upd = upd.ClearInvoiceTypeID()
		} else {
			v, err := s.validateMetadictRef(ctx, id, *req.Msg.InvoiceTypeId, "invoice_type")
			if err != nil {
				return nil, err
			}
			upd = upd.SetInvoiceTypeID(v)
		}
	}
	if req.Msg.DefaultSalesRepId != nil {
		if *req.Msg.DefaultSalesRepId == "" {
			upd = upd.ClearDefaultSalesRepID()
		} else {
			ruid, err := parseID(*req.Msg.DefaultSalesRepId)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 格式錯誤"))
			}
			v, err := s.validateSalesRep(ctx, ruid, cid, did)
			if err != nil {
				return nil, err
			}
			upd = upd.SetDefaultSalesRepID(v)
		}
	}
	if req.Msg.PreferredDeliveryDays != nil {
		upd = upd.SetPreferredDeliveryDays(req.Msg.PreferredDeliveryDays)
	}
	if req.Msg.PromoTagIds != nil {
		upd = upd.SetPromoTagIds(intSlice(req.Msg.PromoTagIds))
	}
	actor, _ := parseID(id.UserID)
	upd = upd.SetUpdatedBy(actor)

	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	rec := audit.Entry{
		Action:       "update",
		ResourceType: "customer",
		ResourceID:   strconv.FormatInt(int64(custID), 10),
		CompanyID:    cid,
		DepartmentID: cur.DepartmentID,
		UserID:       actor,
		After:        map[string]any{"name": updated.Name},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}
	if err := audit.Record(ctx, tx, rec); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.UpdateCustomerResponse{Customer: customerToProto(updated)}), nil
}

// DeleteCustomer 軟刪除(設定 deleted_at + 稽核,同一交易)。
func (s *CustomerService) DeleteCustomer(ctx context.Context, req *connect.Request[customersv1.DeleteCustomerRequest]) (*connect.Response[customersv1.DeleteCustomerResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	cur, err := customerScopeQuery(s.db.Customer.Query(), cid, did).
		Where(customer.ID(custID), customer.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := tx.Customer.UpdateOneID(custID).SetDeletedAt(time.Now().UTC()).SetUpdatedBy(parseActor(id)).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "delete",
		ResourceType: "customer",
		ResourceID:   strconv.FormatInt(int64(custID), 10),
		CompanyID:    cid,
		DepartmentID: cur.DepartmentID,
		UserID:       parseActor(id),
		Before:       map[string]any{"customer_code": cur.CustomerCode, "name": cur.Name},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.DeleteCustomerResponse{}), nil
}

// RestoreCustomer 復原(清 deleted_at + 稽核;若未刪除則冪等回傳)。
func (s *CustomerService) RestoreCustomer(ctx context.Context, req *connect.Request[customersv1.RestoreCustomerRequest]) (*connect.Response[customersv1.RestoreCustomerResponse], error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	cid, did, err := customerScope(id)
	if err != nil {
		return nil, err
	}
	custID, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	// 復原需能找到已刪除列:不加 DeletedAtIsNil,以範圍 + id 查詢。
	cur, err := customerScopeQuery(s.db.Customer.Query(), cid, did).Where(customer.ID(custID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if cur.DeletedAt == nil {
		// 已是 active,冪等回傳。
		return connect.NewResponse(&customersv1.RestoreCustomerResponse{Customer: customerToProto(cur)}), nil
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()
	restored, err := tx.Customer.UpdateOneID(custID).
		ClearDeletedAt().SetUpdatedBy(parseActor(id)).Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := audit.Record(ctx, tx, audit.Entry{
		Action:       "update",
		ResourceType: "customer",
		ResourceID:   strconv.FormatInt(int64(custID), 10),
		CompanyID:    cid,
		DepartmentID: cur.DepartmentID,
		UserID:       parseActor(id),
		After:        map[string]any{"restored": true, "customer_code": restored.CustomerCode},
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.RestoreCustomerResponse{Customer: customerToProto(restored)}), nil
}

// parseActor 由身分取操作者 int id(0 時由 audit.Record 拒寫,符合 D18)。
func parseActor(id authz.Identity) int {
	if pid, err := parseID(id.UserID); err == nil {
		return pid
	}
	return 0
}

// intSlice 將 proto int64 slice 轉為 []int(ent PromoTagIDs 為 []int)。
func intSlice(in []int64) []int {
	if len(in) == 0 {
		return []int{}
	}
	out := make([]int, 0, len(in))
	for _, v := range in {
		out = append(out, int(v))
	}
	return out
}
