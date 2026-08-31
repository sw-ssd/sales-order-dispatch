//
//  Generated code. Do not modify.
//  source: salesorder/v1/company.proto
//

import "package:connectrpc/connect.dart" as connect;
import "company.pb.dart" as salesorderv1company;

/// CompanyService:公司主檔 CRUD。
abstract final class CompanyService {
  /// Fully-qualified name of the CompanyService service.
  static const name = 'salesorder.v1.CompanyService';

  /// ListCompanies:分頁列出公司,可依 status / keyword(name、identifier 模糊)篩選。
  static const listCompanies = connect.Spec(
    '/$name/ListCompanies',
    connect.StreamType.unary,
    salesorderv1company.ListCompaniesRequest.new,
    salesorderv1company.ListCompaniesResponse.new,
  );

  /// GetCompany:取得單一公司。
  static const getCompany = connect.Spec(
    '/$name/GetCompany',
    connect.StreamType.unary,
    salesorderv1company.GetCompanyRequest.new,
    salesorderv1company.GetCompanyResponse.new,
  );

  /// CreateCompany:建立公司。
  static const createCompany = connect.Spec(
    '/$name/CreateCompany',
    connect.StreamType.unary,
    salesorderv1company.CreateCompanyRequest.new,
    salesorderv1company.CreateCompanyResponse.new,
  );

  /// UpdateCompany:更新公司(name、tax_id、status;identifier 建立後不可修改)。
  static const updateCompany = connect.Spec(
    '/$name/UpdateCompany',
    connect.StreamType.unary,
    salesorderv1company.UpdateCompanyRequest.new,
    salesorderv1company.UpdateCompanyResponse.new,
  );

  /// DeleteCompany:刪除公司。
  static const deleteCompany = connect.Spec(
    '/$name/DeleteCompany',
    connect.StreamType.unary,
    salesorderv1company.DeleteCompanyRequest.new,
    salesorderv1company.DeleteCompanyResponse.new,
  );
}
/// DepartmentService:部門主檔 CRUD。
abstract final class DepartmentService {
  /// Fully-qualified name of the DepartmentService service.
  static const name = 'salesorder.v1.DepartmentService';

  /// ListDepartments:分頁列出部門,可依 company_id 篩選。
  static const listDepartments = connect.Spec(
    '/$name/ListDepartments',
    connect.StreamType.unary,
    salesorderv1company.ListDepartmentsRequest.new,
    salesorderv1company.ListDepartmentsResponse.new,
  );

  /// GetDepartment:取得單一部門。
  static const getDepartment = connect.Spec(
    '/$name/GetDepartment',
    connect.StreamType.unary,
    salesorderv1company.GetDepartmentRequest.new,
    salesorderv1company.GetDepartmentResponse.new,
  );

  /// CreateDepartment:建立部門(需指定所屬公司)。
  static const createDepartment = connect.Spec(
    '/$name/CreateDepartment',
    connect.StreamType.unary,
    salesorderv1company.CreateDepartmentRequest.new,
    salesorderv1company.CreateDepartmentResponse.new,
  );

  /// UpdateDepartment:更新部門名稱。
  static const updateDepartment = connect.Spec(
    '/$name/UpdateDepartment',
    connect.StreamType.unary,
    salesorderv1company.UpdateDepartmentRequest.new,
    salesorderv1company.UpdateDepartmentResponse.new,
  );

  /// DeleteDepartment:刪除部門。
  static const deleteDepartment = connect.Spec(
    '/$name/DeleteDepartment',
    connect.StreamType.unary,
    salesorderv1company.DeleteDepartmentRequest.new,
    salesorderv1company.DeleteDepartmentResponse.new,
  );
}
