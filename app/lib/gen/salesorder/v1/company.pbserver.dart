// This is a generated file - do not edit.
//
// Generated from salesorder/v1/company.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'company.pb.dart' as $2;
import 'company.pbjson.dart';

export 'company.pb.dart';

abstract class CompanyServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListCompaniesResponse> listCompanies(
      $pb.ServerContext ctx, $2.ListCompaniesRequest request);
  $async.Future<$2.GetCompanyResponse> getCompany(
      $pb.ServerContext ctx, $2.GetCompanyRequest request);
  $async.Future<$2.CreateCompanyResponse> createCompany(
      $pb.ServerContext ctx, $2.CreateCompanyRequest request);
  $async.Future<$2.UpdateCompanyResponse> updateCompany(
      $pb.ServerContext ctx, $2.UpdateCompanyRequest request);
  $async.Future<$2.DeleteCompanyResponse> deleteCompany(
      $pb.ServerContext ctx, $2.DeleteCompanyRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListCompanies':
        return $2.ListCompaniesRequest();
      case 'GetCompany':
        return $2.GetCompanyRequest();
      case 'CreateCompany':
        return $2.CreateCompanyRequest();
      case 'UpdateCompany':
        return $2.UpdateCompanyRequest();
      case 'DeleteCompany':
        return $2.DeleteCompanyRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListCompanies':
        return listCompanies(ctx, request as $2.ListCompaniesRequest);
      case 'GetCompany':
        return getCompany(ctx, request as $2.GetCompanyRequest);
      case 'CreateCompany':
        return createCompany(ctx, request as $2.CreateCompanyRequest);
      case 'UpdateCompany':
        return updateCompany(ctx, request as $2.UpdateCompanyRequest);
      case 'DeleteCompany':
        return deleteCompany(ctx, request as $2.DeleteCompanyRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => CompanyServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => CompanyServiceBase$messageJson;
}

abstract class DepartmentServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListDepartmentsResponse> listDepartments(
      $pb.ServerContext ctx, $2.ListDepartmentsRequest request);
  $async.Future<$2.GetDepartmentResponse> getDepartment(
      $pb.ServerContext ctx, $2.GetDepartmentRequest request);
  $async.Future<$2.CreateDepartmentResponse> createDepartment(
      $pb.ServerContext ctx, $2.CreateDepartmentRequest request);
  $async.Future<$2.UpdateDepartmentResponse> updateDepartment(
      $pb.ServerContext ctx, $2.UpdateDepartmentRequest request);
  $async.Future<$2.DeleteDepartmentResponse> deleteDepartment(
      $pb.ServerContext ctx, $2.DeleteDepartmentRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListDepartments':
        return $2.ListDepartmentsRequest();
      case 'GetDepartment':
        return $2.GetDepartmentRequest();
      case 'CreateDepartment':
        return $2.CreateDepartmentRequest();
      case 'UpdateDepartment':
        return $2.UpdateDepartmentRequest();
      case 'DeleteDepartment':
        return $2.DeleteDepartmentRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListDepartments':
        return listDepartments(ctx, request as $2.ListDepartmentsRequest);
      case 'GetDepartment':
        return getDepartment(ctx, request as $2.GetDepartmentRequest);
      case 'CreateDepartment':
        return createDepartment(ctx, request as $2.CreateDepartmentRequest);
      case 'UpdateDepartment':
        return updateDepartment(ctx, request as $2.UpdateDepartmentRequest);
      case 'DeleteDepartment':
        return deleteDepartment(ctx, request as $2.DeleteDepartmentRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      DepartmentServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => DepartmentServiceBase$messageJson;
}
