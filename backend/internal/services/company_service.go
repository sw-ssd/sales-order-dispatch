// Package services 提供可直接掛載到 HTTP server 的 Connect 服務實作。
// 各服務僅依賴 *ent.Client,註冊函式回傳掛載路徑供 server 決定前綴。
package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"

	"google.golang.org/protobuf/types/known/structpb"
)

// validCompanyStatuses 為 Company.status 允許值(對齊 ent enum)。
var validCompanyStatuses = map[string]bool{
	string(company.StatusActive):    true,
	string(company.StatusInactive):  true,
	string(company.StatusSuspended): true,
}

// CompanyService 實作 salesorder.v1.CompanyService(公司主檔 CRUD)。
// 授權比照 RoleService:每方法以 Casbin EnforceAny 檢查身分(未登入 → Unauthenticated,
// 無權 → PermissionDenied)。company 資源 read/update 限 super/company_admin(3.1.1),
// create/delete 為 super 專屬(8.2);department 資源 read 限 company_admin/dept_admin,
// write 限 company_admin(超集:super 全權)。
type CompanyService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedCompanyServiceHandler
}

// NewCompanyService 建立 CompanyService。
func NewCompanyService(db *ent.Client) *CompanyService {
	return &CompanyService{db: db}
}

// DepartmentService 實作 salesorder.v1.DepartmentService(部門主檔 CRUD)。
type DepartmentService struct {
	db *ent.Client
	salesorderv1connect.UnimplementedDepartmentServiceHandler
}

// NewDepartmentService 建立 DepartmentService。
func NewDepartmentService(db *ent.Client) *DepartmentService {
	return &DepartmentService{db: db}
}

// RegisterCompanyServices 將 CompanyService 與 DepartmentService 的 Connect handler
// 掛到 mux(以各自自然路徑,如 "/salesorder.v1.CompanyService/")。
//
// server 掛載範例(InitDomains):
//
//	mux := http.NewServeMux()
//	services.RegisterCompanyServices(mux, db)
//	s.router.Mount("/api/v1", mux) // chi Mount 會剝除 /api/v1 前綴,前端 baseUrl "/api/v1" 可直接對應
func RegisterCompanyServices(mux *http.ServeMux, db *ent.Client) {
	companyPath, companyHandler := salesorderv1connect.NewCompanyServiceHandler(NewCompanyService(db))
	mux.Handle(companyPath, companyHandler)
	departmentPath, departmentHandler := salesorderv1connect.NewDepartmentServiceHandler(NewDepartmentService(db))
	mux.Handle(departmentPath, departmentHandler)
}

// requireScope 檢查 ctx 身分具備 resource 資源的指定動作(Casbin EnforceAny,T14)。
// 未登入 → Unauthenticated;無權 → PermissionDenied。
func requireScope(ctx context.Context, resource, action string) error {
	id, err := requireAuth(ctx)
	if err != nil {
		return err
	}
	ok, err := auth.EnforceAny(id.Roles, resource, action, id.CompanyID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if !ok {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("無%s資源的%s權限", resource, action))
	}
	return nil
}

// ListCompanies 分頁列出公司,可依 status / keyword(name、identifier 模糊)篩選。
func (s *CompanyService) ListCompanies(ctx context.Context, req *connect.Request[v1.ListCompaniesRequest]) (*connect.Response[v1.ListCompaniesResponse], error) {
	if err := requireScope(ctx, "company", "read"); err != nil {
		return nil, err
	}
	q := s.db.Company.Query().Where(company.DeletedAtIsNil())
	if status := strings.TrimSpace(req.Msg.GetStatus()); status != "" {
		if !validCompanyStatuses[status] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的公司狀態 %q(允許: active / inactive / suspended)", status))
		}
		q = q.Where(company.StatusEQ(company.Status(status)))
	}
	if keyword := strings.TrimSpace(req.Msg.GetKeyword()); keyword != "" {
		q = q.Where(company.Or(company.NameContainsFold(keyword), company.IdentifierContainsFold(keyword)))
	}

	field, desc, err := companySortField(req.Msg.GetSort(), req.Msg.GetDesc())
	if err != nil {
		return nil, err
	}

	list, pg, err := pageListE(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), companyListSource{q, field, desc}, companyToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.ListCompaniesResponse{Companies: list, Pagination: pg}), nil
}

// companyListSource 為 pageListE 的 ent 查詢橋接(排序白名單已先解析為 field/desc)。
type companyListSource struct {
	q     *ent.CompanyQuery
	field string
	desc  bool
}

func (s companyListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s companyListSource) Page(ctx context.Context, off, lim int) ([]*ent.Company, error) {
	order := ent.Asc(s.field)
	if s.desc {
		order = ent.Desc(s.field)
	}
	// F1:排序鍵非唯一時 PostgreSQL 對同值群(ties)的順序不保證一致,逐頁 LIMIT/OFFSET 會
	// 重複與遺漏資料(前端翻完只看得到部分列),故一律以 id 為次序鍵收斂成全序。
	// field 已是 id 時多一組等價鍵,id 為唯一鍵故結果不變。
	return s.q.Clone().Order(order, ent.Asc(company.FieldID)).Offset(off).Limit(lim).All(ctx)
}

// companySortField 解析排序參數,回傳 ent 欄位與是否降冪(比照 customerSortField 的白名單樣板)。
// sort 空 → 預設 id 降冪(現行行為)並忽略 desc;其餘欄位預設升冪,desc=true 轉降冪。
func companySortField(sort string, desc bool) (string, bool, error) {
	switch sort {
	case "":
		return company.FieldID, true, nil
	case "name":
		return company.FieldName, desc, nil
	case "identifier":
		return company.FieldIdentifier, desc, nil
	case "tax_id":
		return company.FieldTaxID, desc, nil
	case "status":
		return company.FieldStatus, desc, nil
	case "id":
		return company.FieldID, desc, nil
	default:
		return "", false, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("排序欄位 %q 不在白名單(name/identifier/tax_id/status/id)", sort))
	}
}

// GetCompany 取得單一公司。
func (s *CompanyService) GetCompany(ctx context.Context, req *connect.Request[v1.GetCompanyRequest]) (*connect.Response[v1.GetCompanyResponse], error) {
	if err := requireScope(ctx, "company", "read"); err != nil {
		return nil, err
	}
	id, err := parseID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	// 軟刪除(P2-A):已刪除的公司對所有查詢與異動皆不存在。
	c, err := s.db.Company.Query().Where(company.ID(id), company.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	p, err := companyToProto(c)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.GetCompanyResponse{Company: p}), nil
}

// CreateCompany 建立公司。
func (s *CompanyService) CreateCompany(ctx context.Context, req *connect.Request[v1.CreateCompanyRequest]) (*connect.Response[v1.CreateCompanyResponse], error) {
	if err := requireScope(ctx, "company", "create"); err != nil {
		return nil, err
	}
	msg := req.Msg
	if strings.TrimSpace(msg.GetName()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("公司名稱不可為空"))
	}
	if strings.TrimSpace(msg.GetIdentifier()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("識別碼(identifier)不可為空"))
	}
	if status := strings.TrimSpace(msg.GetStatus()); status != "" && !validCompanyStatuses[status] {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的公司狀態 %q(允許: active / inactive / suspended)", status))
	}

	// 識別碼僅在「未刪除」的公司之間唯一(部分唯一索引 companies_identifier_active_unique,
	// migration 00019),軟刪除後可重用。DB 約束是後盾,但自行判別才回得出語意明確的
	// AlreadyExists:約束錯誤一律映射為 FailedPrecondition 且不含 DB 原文(P2-A)。
	identifier := strings.TrimSpace(msg.GetIdentifier())
	used, err := s.db.Company.Query().Where(company.IdentifierEQ(identifier), company.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if used {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("識別碼(identifier) %q 已被使用,請換一個", identifier))
	}

	build := s.db.Company.Create().
		SetName(strings.TrimSpace(msg.GetName())).
		SetIdentifier(identifier)
	if taxID := strings.TrimSpace(msg.GetTaxId()); taxID != "" {
		build = build.SetTaxID(taxID)
	}
	if status := strings.TrimSpace(msg.GetStatus()); status != "" {
		build = build.SetStatus(company.Status(status))
	}
	// public_info / capabilities / logo_url 建立時不開放,依 ent 預設值;D31 以 UpdateBranding 補充。

	created, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	p, err := companyToProto(created)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.CreateCompanyResponse{Company: p}), nil
}

// UpdateCompany 更新公司(name、tax_id、status;identifier 建立後不可修改)。
func (s *CompanyService) UpdateCompany(ctx context.Context, req *connect.Request[v1.UpdateCompanyRequest]) (*connect.Response[v1.UpdateCompanyResponse], error) {
	if err := requireScope(ctx, "company", "update"); err != nil {
		return nil, err
	}
	msg := req.Msg
	id, err := parseID(msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(msg.GetIdentifier()) != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("identifier 建立後不可修改"))
	}

	exists, err := s.db.Company.Query().Where(company.ID(id), company.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 業務異動 + 稽核(D18)同一交易,公司停用連鎖之 status 變更尤須留痕。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	build := tx.Company.UpdateOneID(id)
	changed := false
	before := map[string]any{"name": exists.Name, "tax_id": exists.TaxID, "status": string(exists.Status)}
	after := map[string]any{"name": exists.Name, "tax_id": exists.TaxID, "status": string(exists.Status)}
	if msg.Name != nil {
		if strings.TrimSpace(msg.GetName()) == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("公司名稱不可為空"))
		}
		n := strings.TrimSpace(msg.GetName())
		build = build.SetName(n)
		after["name"] = n
		changed = true
	}
	if msg.TaxId != nil {
		t := strings.TrimSpace(msg.GetTaxId())
		build = build.SetTaxID(t)
		after["tax_id"] = t
		changed = true
	}
	if msg.Status != nil {
		if !validCompanyStatuses[*msg.Status] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的公司狀態 %q(允許: active / inactive / suspended)", *msg.Status))
		}
		st := company.Status(*msg.Status)
		if st != exists.Status {
			build = build.SetStatus(st)
			after["status"] = string(st)
			changed = true
		}
	}

	updated, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if changed {
		act := authz.IdentityFrom(ctx)
		actor, _ := parseID(act.UserID)
		if err := recordAuditBA(ctx, tx, "company", "update", id, id, nil, actor, before, after); err != nil {
			return nil, toConnectError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	p, err := companyToProto(updated)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.UpdateCompanyResponse{Company: p}), nil
}

// DeleteCompany 軟刪除公司(P2-A):標記 deleted_at 而非刪列,同一交易寫 action=delete 稽核。
// 公司仍有部門/使用者時回 FailedPrecondition;已刪除或不存在 → NotFound。
//
// 為何軟刪除:audit_logs.company_id 是租戶欄(NOT NULL + FK→companies),硬刪除只要該公司
// 有任何稽核列即違反 FK(失敗又被舊映射誤報為 AlreadyExists)。列保留後稽核永遠有主可依,
// 且被刪公司的識別碼可由部分唯一索引釋出給新公司。
func (s *CompanyService) DeleteCompany(ctx context.Context, req *connect.Request[v1.DeleteCompanyRequest]) (*connect.Response[v1.DeleteCompanyResponse], error) {
	if err := requireScope(ctx, "company", "delete"); err != nil {
		return nil, err
	}
	id, err := parseID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	cur, err := s.db.Company.Query().Where(company.ID(id), company.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 已軟刪除的部門不算「仍有部門」:軟刪除部門的列會保留(00020),不排除就會讓公司永遠刪不掉。
	hasDepartments, err := s.db.Department.Query().
		Where(department.HasCompanyWith(company.ID(id)), department.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if hasDepartments {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("公司仍有部門,無法刪除"))
	}
	hasUsers, err := s.db.User.Query().Where(user.HasCompanyWith(company.ID(id))).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if hasUsers {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("公司仍有使用者,無法刪除"))
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	actor, _ := parseID(authz.IdentityFrom(ctx).UserID)
	if err := tx.Company.UpdateOneID(id).SetDeletedAt(time.Now().UTC()).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	if err := recordAuditBA(ctx, tx, "company", "delete", id, id, nil, actor,
		map[string]any{"name": cur.Name, "identifier": cur.Identifier, "status": string(cur.Status)}, nil); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.DeleteCompanyResponse{}), nil
}

// ListDepartments 分頁列出部門,可依 company_id 篩選;一併載入所屬公司名稱供顯示。
func (s *DepartmentService) ListDepartments(ctx context.Context, req *connect.Request[v1.ListDepartmentsRequest]) (*connect.Response[v1.ListDepartmentsResponse], error) {
	if err := requireScope(ctx, "department", "read"); err != nil {
		return nil, err
	}
	// 軟刪除(00020):已刪除的部門對所有查詢與異動皆不存在。
	q := s.db.Department.Query().WithCompany().Where(department.DeletedAtIsNil())
	if companyID := strings.TrimSpace(req.Msg.GetCompanyId()); companyID != "" {
		cid, err := parseID(companyID)
		if err != nil {
			return nil, err
		}
		q = q.Where(department.HasCompanyWith(company.ID(cid), company.DeletedAtIsNil()))
	}

	field, desc, err := departmentSortField(req.Msg.GetSort(), req.Msg.GetDesc())
	if err != nil {
		return nil, err
	}

	list, pg, err := pageList(ctx, req.Msg.GetPage(), req.Msg.GetPageSize(), departmentListSource{q, field, desc}, departmentToProto)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.ListDepartmentsResponse{Departments: list, Pagination: pg}), nil
}

// departmentListSource 為 pageList 的 ent 查詢橋接(部門需 eager-load 公司名稱供 toProto;
// 排序白名單已先解析為 field/desc)。
type departmentListSource struct {
	q     *ent.DepartmentQuery
	field string
	desc  bool
}

func (s departmentListSource) Count(ctx context.Context) (int, error) { return s.q.Count(ctx) }
func (s departmentListSource) Page(ctx context.Context, off, lim int) ([]*ent.Department, error) {
	order := ent.Asc(s.field)
	if s.desc {
		order = ent.Desc(s.field)
	}
	// F1:排序鍵非唯一時 PostgreSQL 對同值群(ties)的順序不保證一致,逐頁 LIMIT/OFFSET 會
	// 重複與遺漏資料(前端翻完只看得到部分列),故一律以 id 為次序鍵收斂成全序。
	// field 已是 id 時多一組等價鍵,id 為唯一鍵故結果不變。
	return s.q.Clone().Order(order, ent.Asc(department.FieldID)).Offset(off).Limit(lim).All(ctx)
}

// departmentSortField 解析排序參數,回傳 ent 欄位與是否降冪(比照 companySortField 的白名單樣板)。
// sort 空 → 預設 id 降冪(現行行為)並忽略 desc;其餘欄位預設升冪,desc=true 轉降冪。
func departmentSortField(sort string, desc bool) (string, bool, error) {
	switch sort {
	case "":
		return department.FieldID, true, nil // 預設排序:此案例把 desc 吃掉(契約:sort 空忽略 desc)
	case "name":
		return department.FieldName, desc, nil
	case "id":
		return department.FieldID, desc, nil
	default:
		return "", false, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("排序欄位 %q 不在白名單(name/id)", sort))
	}
}

// GetDepartment 取得單一部門(含所屬公司名稱)。
func (s *DepartmentService) GetDepartment(ctx context.Context, req *connect.Request[v1.GetDepartmentRequest]) (*connect.Response[v1.GetDepartmentResponse], error) {
	if err := requireScope(ctx, "department", "read"); err != nil {
		return nil, err
	}
	id, err := parseID(req.Msg.GetDepartmentId())
	if err != nil {
		return nil, err
	}
	d, err := s.db.Department.Query().WithCompany().Where(department.ID(id), department.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.GetDepartmentResponse{Department: departmentToProto(d)}), nil
}

// CreateDepartment 建立部門(需指定所屬公司)。
func (s *DepartmentService) CreateDepartment(ctx context.Context, req *connect.Request[v1.CreateDepartmentRequest]) (*connect.Response[v1.CreateDepartmentResponse], error) {
	if err := requireScope(ctx, "department", "create"); err != nil {
		return nil, err
	}
	msg := req.Msg
	companyID, err := parseID(msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(msg.GetName()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("部門名稱不可為空"))
	}
	exists, err := s.db.Company.Query().Where(company.ID(companyID), company.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("公司 %d 不存在", companyID))
	}

	created, err := s.db.Department.Create().
		SetCompanyID(companyID).
		SetName(strings.TrimSpace(msg.GetName())).
		Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	// 回傳時補載公司名稱。
	created, err = s.db.Department.Query().WithCompany().Where(department.ID(created.ID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.CreateDepartmentResponse{Department: departmentToProto(created)}), nil
}

// UpdateDepartment 更新部門名稱。
func (s *DepartmentService) UpdateDepartment(ctx context.Context, req *connect.Request[v1.UpdateDepartmentRequest]) (*connect.Response[v1.UpdateDepartmentResponse], error) {
	if err := requireScope(ctx, "department", "update"); err != nil {
		return nil, err
	}
	msg := req.Msg
	id, err := parseID(msg.GetDepartmentId())
	if err != nil {
		return nil, err
	}
	// 軟刪除(00020):已刪除的部門視同不存在,不得再更名(UpdateOneID 不看 deleted_at,故須自行擋)。
	if _, err := s.db.Department.Query().Where(department.ID(id), department.DeletedAtIsNil()).Only(ctx); err != nil {
		return nil, toConnectError(err)
	}
	build := s.db.Department.UpdateOneID(id)
	if msg.Name != nil {
		if strings.TrimSpace(msg.GetName()) == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("部門名稱不可為空"))
		}
		build = build.SetName(strings.TrimSpace(msg.GetName()))
	}

	updated, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	updated, err = s.db.Department.Query().WithCompany().Where(department.ID(updated.ID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.UpdateDepartmentResponse{Department: departmentToProto(updated)}), nil
}

// DeleteDepartment 軟刪除部門(00020):標記 deleted_at 而非刪列,同一交易寫 action=delete 稽核。
// 部門仍有使用者時回 FailedPrecondition;已刪除或不存在 → NotFound。
//
// 為何軟刪除:audit_logs.department_id 是 FK(00010),而以「目標使用者部門」寫入的稽核
// (recordUserAudit)以及倉別/路線/加工規格/產品分類/客戶的 department_id,在成員被調離後
// 仍會指向該部門 —— 硬刪除必被 FK 擋下,且錯誤被映射成與原因無關的通用訊息。列保留後所有
// FK 永遠有主可依。軟刪除同時拿走了硬刪除的隱性保護,故所有部門查詢與掛載路徑皆已排除已刪除列
// (見 migration 00020 檔頭影響清單)。
func (s *DepartmentService) DeleteDepartment(ctx context.Context, req *connect.Request[v1.DeleteDepartmentRequest]) (*connect.Response[v1.DeleteDepartmentResponse], error) {
	if err := requireScope(ctx, "department", "delete"); err != nil {
		return nil, err
	}
	id, err := parseID(req.Msg.GetDepartmentId())
	if err != nil {
		return nil, err
	}
	cur, err := s.db.Department.Query().WithCompany().
		Where(department.ID(id), department.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	hasUsers, err := s.db.User.Query().Where(user.HasDepartmentWith(department.ID(id))).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if hasUsers {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("部門仍有使用者,無法刪除"))
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.Department.UpdateOneID(id).SetDeletedAt(time.Now().UTC()).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	companyID := 0
	if cur.Edges.Company != nil {
		companyID = cur.Edges.Company.ID
	}
	// 稽核的 department_id 指向被刪部門本身:列保留(軟刪除)故 FK 成立。
	actor, _ := parseID(authz.IdentityFrom(ctx).UserID)
	if err := recordAuditBA(ctx, tx, "department", "delete", id, companyID, &id, actor,
		map[string]any{"name": cur.Name, "company_id": companyID}, nil); err != nil {
		return nil, toConnectError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.DeleteDepartmentResponse{}), nil
}

// parseID 將字串 ID 轉為 ent 自增 int64 ID;格式錯誤回 InvalidArgument。
func parseID(s string) (int, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || id <= 0 {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的 ID %q", s))
	}
	return int(id), nil
}

// toConnectError 將 ent 錯誤映射為 Connect 錯誤碼。
//
// 約束錯誤(ent.IsConstraintError)涵蓋唯一鍵衝突與 FK 阻擋,其 Error() 挾帶驅動層原文
// (SQLSTATE、約束名、欄位名…),一律不外洩給客戶端:改以 FailedPrecondition + 繁中可行動
// 訊息表示「資料現況不允許此操作」(P2-A 的原始缺陷正是 FK 阻擋被當成識別碼重複回
// AlreadyExists 並附上整句 SQL)。需要 AlreadyExists 語意者由呼叫端自行判別(如
// CreateCompany 以 DeletedAtIsNil 前置查詢判斷識別碼重複),語意明確且訊息不含 DB 細節。
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case ent.IsNotFound(err):
		return connect.NewError(connect.CodeNotFound, err)
	case ent.IsValidationError(err):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case ent.IsConstraintError(err):
		// 原始錯誤(含 SQLSTATE/約束名/欄位名)仍必須落 server log 才能追查;只對客戶端隱藏。
		log.Printf("services: 資料庫約束錯誤(已映射為 failed_precondition,不對外揭露細節): %v", err)
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("資料違反資料庫約束,無法完成此操作(請確認識別碼是否已被使用、參照對象是否仍存在)"))
	case ent.IsNotSingular(err):
		return connect.NewError(connect.CodeInternal, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

// companyToProto 將 ent.Company 轉為 proto Company。
func companyToProto(c *ent.Company) (*v1.Company, error) {
	var publicInfo *structpb.Struct
	if c.PublicInfo != nil {
		s, err := structpb.NewStruct(c.PublicInfo)
		if err != nil {
			return nil, fmt.Errorf("public_info 轉換失敗: %w", err)
		}
		publicInfo = s
	}
	return &v1.Company{
		Id:           strconv.FormatInt(int64(c.ID), 10),
		Name:         c.Name,
		TaxId:        c.TaxID,
		Identifier:   c.Identifier,
		Status:       string(c.Status),
		PublicInfo:   publicInfo,
		Capabilities: c.Capabilities,
		LogoUrl:      c.LogoURL,
	}, nil
}

// departmentToProto 將 ent.Department 轉為 proto Department(公司名稱需已 eager-load)。
func departmentToProto(d *ent.Department) *v1.Department {
	p := &v1.Department{
		Id:   strconv.FormatInt(int64(d.ID), 10),
		Name: d.Name,
	}
	if d.Edges.Company != nil {
		p.CompanyId = strconv.FormatInt(int64(d.Edges.Company.ID), 10)
		p.CompanyName = d.Edges.Company.Name
	}
	return p
}
