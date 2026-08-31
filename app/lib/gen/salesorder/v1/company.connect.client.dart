//
//  Generated code. Do not modify.
//  source: salesorder/v1/company.proto
//

import "package:connectrpc/connect.dart" as connect;
import "company.pb.dart" as salesorderv1company;
import "company.connect.spec.dart" as specs;

/// CompanyService:公司主檔 CRUD。
extension type CompanyServiceClient (connect.Transport _transport) {
  /// ListCompanies:分頁列出公司,可依 status / keyword(name、identifier 模糊)篩選。
  Future<salesorderv1company.ListCompaniesResponse> listCompanies(
    salesorderv1company.ListCompaniesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CompanyService.listCompanies,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCompany:取得單一公司。
  Future<salesorderv1company.GetCompanyResponse> getCompany(
    salesorderv1company.GetCompanyRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CompanyService.getCompany,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateCompany:建立公司。
  Future<salesorderv1company.CreateCompanyResponse> createCompany(
    salesorderv1company.CreateCompanyRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CompanyService.createCompany,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateCompany:更新公司(name、tax_id、status;identifier 建立後不可修改)。
  Future<salesorderv1company.UpdateCompanyResponse> updateCompany(
    salesorderv1company.UpdateCompanyRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CompanyService.updateCompany,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteCompany:刪除公司。
  Future<salesorderv1company.DeleteCompanyResponse> deleteCompany(
    salesorderv1company.DeleteCompanyRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CompanyService.deleteCompany,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// DepartmentService:部門主檔 CRUD。
extension type DepartmentServiceClient (connect.Transport _transport) {
  /// ListDepartments:分頁列出部門,可依 company_id 篩選。
  Future<salesorderv1company.ListDepartmentsResponse> listDepartments(
    salesorderv1company.ListDepartmentsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DepartmentService.listDepartments,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetDepartment:取得單一部門。
  Future<salesorderv1company.GetDepartmentResponse> getDepartment(
    salesorderv1company.GetDepartmentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DepartmentService.getDepartment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateDepartment:建立部門(需指定所屬公司)。
  Future<salesorderv1company.CreateDepartmentResponse> createDepartment(
    salesorderv1company.CreateDepartmentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DepartmentService.createDepartment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateDepartment:更新部門名稱。
  Future<salesorderv1company.UpdateDepartmentResponse> updateDepartment(
    salesorderv1company.UpdateDepartmentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DepartmentService.updateDepartment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteDepartment:刪除部門。
  Future<salesorderv1company.DeleteDepartmentResponse> deleteDepartment(
    salesorderv1company.DeleteDepartmentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DepartmentService.deleteDepartment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
