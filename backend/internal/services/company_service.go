// Package services 提供可直接掛載到 HTTP server 的 Connect 服務實作。
// 各服務僅依賴 *ent.Client,註冊函式回傳掛載路徑供 server 決定前綴。
package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

const (
	defaultPageSize = 20
	maxPageSize     = 100
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
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
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
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())

	q := s.db.Company.Query()
	if status := strings.TrimSpace(req.Msg.GetStatus()); status != "" {
		if !validCompanyStatuses[status] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的公司狀態 %q(允許: active / inactive / suspended)", status))
		}
		q = q.Where(company.StatusEQ(company.Status(status)))
	}
	if keyword := strings.TrimSpace(req.Msg.GetKeyword()); keyword != "" {
		q = q.Where(company.Or(company.NameContainsFold(keyword), company.IdentifierContainsFold(keyword)))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Order(ent.Desc(company.FieldID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	companies := make([]*v1.Company, 0, len(items))
	for _, c := range items {
		p, err := companyToProto(c)
		if err != nil {
			return nil, toConnectError(err)
		}
		companies = append(companies, p)
	}
	return connect.NewResponse(&v1.ListCompaniesResponse{
		Companies:  companies,
		Pagination: &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
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
	c, err := s.db.Company.Get(ctx, id)
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

	build := s.db.Company.Create().
		SetName(strings.TrimSpace(msg.GetName())).
		SetIdentifier(strings.TrimSpace(msg.GetIdentifier()))
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

	build := s.db.Company.UpdateOneID(id)
	if msg.Name != nil {
		if strings.TrimSpace(msg.GetName()) == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("公司名稱不可為空"))
		}
		build = build.SetName(strings.TrimSpace(msg.GetName()))
	}
	if msg.TaxId != nil {
		build = build.SetTaxID(strings.TrimSpace(msg.GetTaxId()))
	}
	if msg.Status != nil {
		if !validCompanyStatuses[*msg.Status] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("無效的公司狀態 %q(允許: active / inactive / suspended)", *msg.Status))
		}
		build = build.SetStatus(company.Status(*msg.Status))
	}

	updated, err := build.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	p, err := companyToProto(updated)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.UpdateCompanyResponse{Company: p}), nil
}

// DeleteCompany 刪除公司。公司仍有部門/使用者時回 FailedPrecondition。
func (s *CompanyService) DeleteCompany(ctx context.Context, req *connect.Request[v1.DeleteCompanyRequest]) (*connect.Response[v1.DeleteCompanyResponse], error) {
	if err := requireScope(ctx, "company", "delete"); err != nil {
		return nil, err
	}
	id, err := parseID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, err
	}
	exists, err := s.db.Company.Query().Where(company.ID(id)).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("公司 %d 不存在", id))
	}
	hasDepartments, err := s.db.Department.Query().Where(department.HasCompanyWith(company.ID(id))).Exist(ctx)
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
	if err := s.db.Company.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.DeleteCompanyResponse{}), nil
}

// ListDepartments 分頁列出部門,可依 company_id 篩選;一併載入所屬公司名稱供顯示。
func (s *DepartmentService) ListDepartments(ctx context.Context, req *connect.Request[v1.ListDepartmentsRequest]) (*connect.Response[v1.ListDepartmentsResponse], error) {
	if err := requireScope(ctx, "department", "read"); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())

	q := s.db.Department.Query().WithCompany()
	if companyID := strings.TrimSpace(req.Msg.GetCompanyId()); companyID != "" {
		cid, err := parseID(companyID)
		if err != nil {
			return nil, err
		}
		q = q.Where(department.HasCompanyWith(company.ID(cid)))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	items, err := q.Order(ent.Desc(department.FieldID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}

	departments := make([]*v1.Department, 0, len(items))
	for _, d := range items {
		departments = append(departments, departmentToProto(d))
	}
	return connect.NewResponse(&v1.ListDepartmentsResponse{
		Departments: departments,
		Pagination:  &v1.Pagination{Page: int32(page), PageSize: int32(pageSize), Total: int64(total)},
	}), nil
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
	d, err := s.db.Department.Query().WithCompany().Where(department.ID(id)).Only(ctx)
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
	exists, err := s.db.Company.Query().Where(company.ID(companyID)).Exist(ctx)
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

// DeleteDepartment 刪除部門。部門仍有使用者時回 FailedPrecondition。
func (s *DepartmentService) DeleteDepartment(ctx context.Context, req *connect.Request[v1.DeleteDepartmentRequest]) (*connect.Response[v1.DeleteDepartmentResponse], error) {
	if err := requireScope(ctx, "department", "delete"); err != nil {
		return nil, err
	}
	id, err := parseID(req.Msg.GetDepartmentId())
	if err != nil {
		return nil, err
	}
	exists, err := s.db.Department.Query().Where(department.ID(id)).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("部門 %d 不存在", id))
	}
	hasUsers, err := s.db.User.Query().Where(user.HasDepartmentWith(department.ID(id))).Exist(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if hasUsers {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("部門仍有使用者,無法刪除"))
	}
	if err := s.db.Department.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&v1.DeleteDepartmentResponse{}), nil
}

// normalizePage 收斂分頁參數:page ≥ 1、page_size 落在 [1, maxPageSize]。
func normalizePage(page, pageSize int32) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return int(page), int(pageSize)
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
		// 唯一值衝突(identifier 重複)或 FK 阻擋(仍有參照)。
		return connect.NewError(connect.CodeAlreadyExists, err)
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
