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
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
)

// businessRepRoles 為可作為 default_sales_rep 的角色(業務/管理,非客戶主帳號)。
var businessRepRoles = map[string]bool{
	"super": true, "company_admin": true, "dept_admin": true, "staff": true,
}

// maxCustomerCodeRetries 取號樂觀鎖重試上限(D7, 3.1.3)。
const maxCustomerCodeRetries = 5

// CustomerService 實作 customers.v1.CustomerService。
type CustomerService struct {
	db                   *ent.Client
	accountManageBaseURL string // config.Auth.FrontendURL(組 D22 帳號管理深層連結)
	customersv1connect.UnimplementedCustomerServiceHandler
}

// NewCustomerService 建立 CustomerService。accountManageBaseURL 為前端 base URL,用於組出
// D22 帳號管理深層連結(config.Auth.FrontendURL)。
func NewCustomerService(db *ent.Client, accountManageBaseURL string) *CustomerService {
	return &CustomerService{db: db, accountManageBaseURL: accountManageBaseURL}
}

// RegisterCustomerServices 將 CustomerService 掛到 mux。
func RegisterCustomerServices(mux *http.ServeMux, db *ent.Client, accountManageBaseURL string) {
	path, handler := customersv1connect.NewCustomerServiceHandler(NewCustomerService(db, accountManageBaseURL))
	mux.Handle(path, handler)
}

// hasRole 判斷身分是否含指定角色。
func hasRole(id authz.Identity, role string) bool {
	return slices.Contains(id.Roles, role)
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

// ensureCustomerCounter 確保該公司的 counter 列存在(獨立、幂等的建前步驟)。
// 為何不放在主交易內建:併發首次建立會在 Postgres 造成唯一衝突而 abort 整個交易,
// 之後的取號重試無法進行;故先以獨立小交易建好,主交易內只做樂觀更新(0 列非錯誤,交易仍健康)。
func (s *CustomerService) ensureCustomerCounter(ctx context.Context, cid int) error {
	exists, err := s.db.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(cid)).Exist(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if exists {
		return nil
	}
	if _, err := s.db.CustomerCounter.Create().SetCompanyID(cid).SetNextSeq(1).SetVersion(0).Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil // 已被併發建立,視為成功
		}
		return toConnectError(err)
	}
	return nil
}

// nextCustomerCode 於交易內以樂觀鎖 counter 取號,回傳 customer_code(公司前綴 + 6 位補零)。
// 呼叫前須先 ensureCustomerCounter。version 衝突(影響 0 列,非錯誤)則重試;逾限回 failed_precondition。
func nextCustomerCode(ctx context.Context, tx *ent.Tx, cid int, prefix string) (string, error) {
	for i := 0; i < maxCustomerCodeRetries; i++ {
		c, err := tx.CustomerCounter.Query().Where(customercounter.CompanyIDEQ(cid)).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("客戶編號計數器未初始化"))
			}
			return "", err
		}
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
	return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("客戶編號取號衝突,請稍後重試"))
}

// ListCustomers 分頁查詢,支持關鍵字(名稱/編號/統編)模糊比對與 include_deleted、排序白名單。
func (s *CustomerService) ListCustomers(ctx context.Context, req *connect.Request[customersv1.ListCustomersRequest]) (*connect.Response[customersv1.ListCustomersResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
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

	list, pg, err := pageList(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), customerListSource{q, field}, customerToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&customersv1.ListCustomersResponse{Customers: list, Pagination: pg}), nil
}

// customerListSource 為 pageList 的 ent 查詢橋接(排序白名單已先算為 field)。
type customerListSource struct {
	q     *ent.CustomerQuery
	field string
}

func (s customerListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s customerListSource) Page(ctx context.Context, off, lim int) ([]*ent.Customer, error) {
	// P1-A(與 F1 同型):排序鍵非唯一時 PostgreSQL 對同值群(ties)的順序不保證一致,逐頁
	// LIMIT/OFFSET 會重複與遺漏資料,故一律以 id 為次序鍵收斂成全序(name/customer_code/
	// created_at 三者中 name 與 created_at 皆非唯一;field 為唯一鍵時多一組等價鍵,結果不變)。
	return s.q.Clone().Order(ent.Asc(s.field), ent.Asc(customer.FieldID)).Offset(off).Limit(lim).All(ctx)
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
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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
	// D22 交付流程必要:業務子帳號憑證交予 default_sales_rep(3.1.4 錯誤處理:未提供/非法 → invalid_argument)。
	if req.Msg.GetDefaultSalesRepId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 必填(建檔連動業務子帳號交付)"))
	}
	ruid, err := parseID(req.Msg.GetDefaultSalesRepId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("default_sales_rep_id 格式錯誤"))
	}
	repID, err = s.validateSalesRep(ctx, ruid, cid, did)
	if err != nil {
		return nil, err
	}
	// 公司前綴(customer_code 取號必要)。軟刪除的公司(P2-A)視同不存在:不得在其下建客戶。
	co, err := s.db.Company.Query().Where(company.ID(cid), company.DeletedAtIsNil()).Only(ctx)
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

	// 先確保 counter 列存在(獨立幂等步驟,避免主交易被唯一衝突 abort;見 ensureCustomerCounter)。
	if err := s.ensureCustomerCounter(ctx, cid); err != nil {
		return nil, err
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
	// 偏好送貨日未帶時套預設(一~六全 false, 長度 6, D26)。
	days := req.Msg.GetPreferredDeliveryDays()
	if len(days) == 0 {
		days = []bool{false, false, false, false, false, false}
	}
	build := tx.Customer.Create().
		SetCompanyID(cid).
		SetCustomerCode(code).
		SetName(name).
		SetPreferredDeliveryDays(days).
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
	if err := recordAudit(ctx, tx, "customer", "create", created.ID, cid, created.DepartmentID, actor, map[string]any{"customer_code": created.CustomerCode, "name": created.Name}); err != nil {
		return nil, toConnectError(err)
	}

	// D22(3.1.4)同期建主帳號 + 業務子帳號:兩帳號各自 24h 隨機臨時密碼、must_change_password=true;
	// 主帳號 system_generated=false、業務子帳號 system_generated=true(灰化標記)。
	// email 以全域唯一之 customer_code 生成佔位(user.email 全域唯一,避免撞既有帳號)。
	primaryEmail := fmt.Sprintf("customer.%s@system.local", created.CustomerCode)
	subEmail := fmt.Sprintf("salesrep.%s@system.local", created.CustomerCode)
	subName := created.Name + "(業務)"
	primaryTemp, err := auth.GenerateTempPassword()
	if err != nil {
		return nil, toConnectError(err)
	}
	primaryHash, err := auth.HashPassword(primaryTemp)
	if err != nil {
		return nil, toConnectError(err)
	}
	subTemp, err := auth.GenerateTempPassword()
	if err != nil {
		return nil, toConnectError(err)
	}
	subHash, err := auth.HashPassword(subTemp)
	if err != nil {
		return nil, toConnectError(err)
	}
	exp := time.Now().UTC().Add(customerTempPasswordTTL)

	primaryUser, err := buildCustomerAccount(ctx, tx, accountSpec{
		CompanyID: cid, DepartmentID: did, CustomerID: created.ID,
		Email: primaryEmail, Name: created.Name, AccountName: created.Name,
		IsPrimary: true, SystemGenerated: false,
		PasswordHash: primaryHash, MustChange: true, TempExpiresAt: exp,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	subUser, err := buildCustomerAccount(ctx, tx, accountSpec{
		CompanyID: cid, DepartmentID: did, CustomerID: created.ID,
		Email: subEmail, Name: subName, AccountName: subName,
		IsPrimary: false, SystemGenerated: true,
		PasswordHash: subHash, MustChange: true, TempExpiresAt: exp,
	})
	if err != nil {
		return nil, toConnectError(err)
	}

	// 兩帳號建檔稽核(與客戶建檔同交易,D18)。
	if err := auditUserCreate(ctx, tx, actor, cid, did, primaryUser.ID, primaryEmail, created.Name); err != nil {
		return nil, toConnectError(err)
	}
	if err := auditUserCreate(ctx, tx, actor, cid, did, subUser.ID, subEmail, subName); err != nil {
		return nil, toConnectError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.CreateCustomerResponse{
		Customer:             customerToProto(created),
		PrimaryAccountName:   primaryUser.Name,
		PrimaryTempPassword:  primaryTemp,
		SalesRepAccountName:  subUser.Name,
		SalesRepTempPassword: subTemp,
		AccountManageUrl:     s.accountManageURL(),
	}), nil
}

// UpdateCustomer 欄位式更新(customer_code 不可改;update 請求無 code 欄位,天然拒絕)。
func (s *CustomerService) UpdateCustomer(ctx context.Context, req *connect.Request[customersv1.UpdateCustomerRequest]) (*connect.Response[customersv1.UpdateCustomerResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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
	if err := recordAudit(ctx, tx, "customer", "update", custID, cid, cur.DepartmentID, actor, map[string]any{"name": updated.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.UpdateCustomerResponse{Customer: customerToProto(updated)}), nil
}

// DeleteCustomer 軟刪除(設定 deleted_at + 稽核,同一交易)。
func (s *CustomerService) DeleteCustomer(ctx context.Context, req *connect.Request[customersv1.DeleteCustomerRequest]) (*connect.Response[customersv1.DeleteCustomerResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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
	if err := recordAudit(ctx, tx, "customer", "delete", custID, cid, cur.DepartmentID, parseActor(id), map[string]any{"customer_code": cur.CustomerCode, "name": cur.Name}); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&customersv1.DeleteCustomerResponse{}), nil
}

// RestoreCustomer 復原(清 deleted_at + 稽核;若未刪除則冪等回傳)。
func (s *CustomerService) RestoreCustomer(ctx context.Context, req *connect.Request[customersv1.RestoreCustomerRequest]) (*connect.Response[customersv1.RestoreCustomerResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	cid, did, err := deptScope(id)
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
	if err := recordAudit(ctx, tx, "customer", "update", custID, cid, cur.DepartmentID, parseActor(id), map[string]any{"restored": true, "customer_code": restored.CustomerCode}); err != nil {
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

// customerTempPasswordTTL 為建檔連動帳號臨時密碼效期(now+24h, D22/01-auth 1.5.2)。
const customerTempPasswordTTL = 24 * time.Hour

// accountSpec 描述 D22 建檔連動要建立的客戶帳號。
type accountSpec struct {
	CompanyID       int
	DepartmentID    *int
	CustomerID      int
	Email           string
	Name            string
	AccountName     string
	IsPrimary       bool
	SystemGenerated bool
	PasswordHash    string
	MustChange      bool
	TempExpiresAt   time.Time
}

// buildCustomerAccount 於交易內建立客戶帳號(角色 customer、is_customer=true)。
func buildCustomerAccount(ctx context.Context, tx *ent.Tx, s accountSpec) (*ent.User, error) {
	b := tx.User.Create().
		SetCompanyID(s.CompanyID).
		SetEmail(s.Email).
		SetName(s.Name).
		SetRole("customer").
		SetStatus(user.StatusActive).
		SetAccountName(s.AccountName).
		SetIsCustomer(true).
		SetCustomerID(s.CustomerID).
		SetIsPrimary(s.IsPrimary).
		SetSystemGenerated(s.SystemGenerated).
		SetPasswordHash(s.PasswordHash).
		SetMustChangePassword(s.MustChange).
		SetTempPasswordExpiresAt(s.TempExpiresAt)
	if s.DepartmentID != nil {
		b = b.SetDepartmentID(*s.DepartmentID)
	}
	return b.Save(ctx)
}

// auditUserCreate 在交易內寫一筆「客戶帳號建檔」稽核(與客戶主檔建檔同回滾,D18)。
func auditUserCreate(ctx context.Context, tx *ent.Tx, actor, cid int, did *int, uid int, email, name string) error {
	return recordAudit(ctx, tx, "user", "create", uid, cid, did, actor, map[string]any{"email": email, "name": name, "account_type": "customer"})
}

// accountManageURL 組出 D22 帳號管理深層連結 https://<domain>/customer_account_manage(規格 §9.4)。
func (s *CustomerService) accountManageURL() string {
	base := strings.TrimRight(s.accountManageBaseURL, "/")
	if base == "" {
		return "/customer_account_manage"
	}
	return base + "/customer_account_manage"
}
