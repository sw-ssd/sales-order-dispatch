// This is a generated file - do not edit.
//
// Generated from salesorder/v1/company.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../google/protobuf/struct.pbjson.dart' as $0;
import 'common.pbjson.dart' as $1;

@$core.Deprecated('Use companyDescriptor instead')
const Company$json = {
  '1': 'Company',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'tax_id', '3': 3, '4': 1, '5': 9, '10': 'taxId'},
    {'1': 'identifier', '3': 4, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'status', '3': 5, '4': 1, '5': 9, '10': 'status'},
    {
      '1': 'public_info',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'publicInfo'
    },
    {'1': 'capabilities', '3': 7, '4': 3, '5': 9, '10': 'capabilities'},
    {'1': 'logo_url', '3': 8, '4': 1, '5': 9, '10': 'logoUrl'},
  ],
};

/// Descriptor for `Company`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List companyDescriptor = $convert.base64Decode(
    'CgdDb21wYW55Eg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEhUKBnRheF9pZB'
    'gDIAEoCVIFdGF4SWQSHgoKaWRlbnRpZmllchgEIAEoCVIKaWRlbnRpZmllchIWCgZzdGF0dXMY'
    'BSABKAlSBnN0YXR1cxI4CgtwdWJsaWNfaW5mbxgGIAEoCzIXLmdvb2dsZS5wcm90b2J1Zi5TdH'
    'J1Y3RSCnB1YmxpY0luZm8SIgoMY2FwYWJpbGl0aWVzGAcgAygJUgxjYXBhYmlsaXRpZXMSGQoI'
    'bG9nb191cmwYCCABKAlSB2xvZ29Vcmw=');

@$core.Deprecated('Use listCompaniesRequestDescriptor instead')
const ListCompaniesRequest$json = {
  '1': 'ListCompaniesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'status', '3': 3, '4': 1, '5': 9, '10': 'status'},
    {'1': 'keyword', '3': 4, '4': 1, '5': 9, '10': 'keyword'},
  ],
};

/// Descriptor for `ListCompaniesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCompaniesRequestDescriptor = $convert.base64Decode(
    'ChRMaXN0Q29tcGFuaWVzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZR'
    'gCIAEoBVIIcGFnZVNpemUSFgoGc3RhdHVzGAMgASgJUgZzdGF0dXMSGAoHa2V5d29yZBgEIAEo'
    'CVIHa2V5d29yZA==');

@$core.Deprecated('Use listCompaniesResponseDescriptor instead')
const ListCompaniesResponse$json = {
  '1': 'ListCompaniesResponse',
  '2': [
    {
      '1': 'companies',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Company',
      '10': 'companies'
    },
    {
      '1': 'pagination',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Pagination',
      '10': 'pagination'
    },
  ],
};

/// Descriptor for `ListCompaniesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCompaniesResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0Q29tcGFuaWVzUmVzcG9uc2USNAoJY29tcGFuaWVzGAEgAygLMhYuc2FsZXNvcmRlci'
    '52MS5Db21wYW55Ugljb21wYW5pZXMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIu'
    'djEuUGFnaW5hdGlvblIKcGFnaW5hdGlvbg==');

@$core.Deprecated('Use getCompanyRequestDescriptor instead')
const GetCompanyRequest$json = {
  '1': 'GetCompanyRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
  ],
};

/// Descriptor for `GetCompanyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCompanyRequestDescriptor = $convert.base64Decode(
    'ChFHZXRDb21wYW55UmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SWQ=');

@$core.Deprecated('Use getCompanyResponseDescriptor instead')
const GetCompanyResponse$json = {
  '1': 'GetCompanyResponse',
  '2': [
    {
      '1': 'company',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Company',
      '10': 'company'
    },
  ],
};

/// Descriptor for `GetCompanyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCompanyResponseDescriptor = $convert.base64Decode(
    'ChJHZXRDb21wYW55UmVzcG9uc2USMAoHY29tcGFueRgBIAEoCzIWLnNhbGVzb3JkZXIudjEuQ2'
    '9tcGFueVIHY29tcGFueQ==');

@$core.Deprecated('Use createCompanyRequestDescriptor instead')
const CreateCompanyRequest$json = {
  '1': 'CreateCompanyRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'tax_id', '3': 2, '4': 1, '5': 9, '10': 'taxId'},
    {'1': 'identifier', '3': 3, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'status', '3': 4, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `CreateCompanyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCompanyRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVDb21wYW55UmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1lEhUKBnRheF9pZBgCIA'
    'EoCVIFdGF4SWQSHgoKaWRlbnRpZmllchgDIAEoCVIKaWRlbnRpZmllchIWCgZzdGF0dXMYBCAB'
    'KAlSBnN0YXR1cw==');

@$core.Deprecated('Use createCompanyResponseDescriptor instead')
const CreateCompanyResponse$json = {
  '1': 'CreateCompanyResponse',
  '2': [
    {
      '1': 'company',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Company',
      '10': 'company'
    },
  ],
};

/// Descriptor for `CreateCompanyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCompanyResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVDb21wYW55UmVzcG9uc2USMAoHY29tcGFueRgBIAEoCzIWLnNhbGVzb3JkZXIudj'
    'EuQ29tcGFueVIHY29tcGFueQ==');

@$core.Deprecated('Use updateCompanyRequestDescriptor instead')
const UpdateCompanyRequest$json = {
  '1': 'UpdateCompanyRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'name', '17': true},
    {'1': 'tax_id', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'taxId', '17': true},
    {'1': 'status', '3': 4, '4': 1, '5': 9, '9': 2, '10': 'status', '17': true},
    {'1': 'identifier', '3': 5, '4': 1, '5': 9, '10': 'identifier'},
  ],
  '8': [
    {'1': '_name'},
    {'1': '_tax_id'},
    {'1': '_status'},
  ],
};

/// Descriptor for `UpdateCompanyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCompanyRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVDb21wYW55UmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SWQSFw'
    'oEbmFtZRgCIAEoCUgAUgRuYW1liAEBEhoKBnRheF9pZBgDIAEoCUgBUgV0YXhJZIgBARIbCgZz'
    'dGF0dXMYBCABKAlIAlIGc3RhdHVziAEBEh4KCmlkZW50aWZpZXIYBSABKAlSCmlkZW50aWZpZX'
    'JCBwoFX25hbWVCCQoHX3RheF9pZEIJCgdfc3RhdHVz');

@$core.Deprecated('Use updateCompanyResponseDescriptor instead')
const UpdateCompanyResponse$json = {
  '1': 'UpdateCompanyResponse',
  '2': [
    {
      '1': 'company',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Company',
      '10': 'company'
    },
  ],
};

/// Descriptor for `UpdateCompanyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCompanyResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVDb21wYW55UmVzcG9uc2USMAoHY29tcGFueRgBIAEoCzIWLnNhbGVzb3JkZXIudj'
    'EuQ29tcGFueVIHY29tcGFueQ==');

@$core.Deprecated('Use deleteCompanyRequestDescriptor instead')
const DeleteCompanyRequest$json = {
  '1': 'DeleteCompanyRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
  ],
};

/// Descriptor for `DeleteCompanyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCompanyRequestDescriptor = $convert.base64Decode(
    'ChREZWxldGVDb21wYW55UmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SWQ=');

@$core.Deprecated('Use deleteCompanyResponseDescriptor instead')
const DeleteCompanyResponse$json = {
  '1': 'DeleteCompanyResponse',
};

/// Descriptor for `DeleteCompanyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCompanyResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVDb21wYW55UmVzcG9uc2U=');

@$core.Deprecated('Use departmentDescriptor instead')
const Department$json = {
  '1': 'Department',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'company_name', '3': 3, '4': 1, '5': 9, '10': 'companyName'},
    {'1': 'name', '3': 4, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `Department`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List departmentDescriptor = $convert.base64Decode(
    'CgpEZXBhcnRtZW50Eg4KAmlkGAEgASgJUgJpZBIdCgpjb21wYW55X2lkGAIgASgJUgljb21wYW'
    '55SWQSIQoMY29tcGFueV9uYW1lGAMgASgJUgtjb21wYW55TmFtZRISCgRuYW1lGAQgASgJUgRu'
    'YW1l');

@$core.Deprecated('Use listDepartmentsRequestDescriptor instead')
const ListDepartmentsRequest$json = {
  '1': 'ListDepartmentsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'company_id', '3': 3, '4': 1, '5': 9, '10': 'companyId'},
  ],
};

/// Descriptor for `ListDepartmentsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDepartmentsRequestDescriptor =
    $convert.base64Decode(
        'ChZMaXN0RGVwYXJ0bWVudHNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaX'
        'plGAIgASgFUghwYWdlU2l6ZRIdCgpjb21wYW55X2lkGAMgASgJUgljb21wYW55SWQ=');

@$core.Deprecated('Use listDepartmentsResponseDescriptor instead')
const ListDepartmentsResponse$json = {
  '1': 'ListDepartmentsResponse',
  '2': [
    {
      '1': 'departments',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Department',
      '10': 'departments'
    },
    {
      '1': 'pagination',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Pagination',
      '10': 'pagination'
    },
  ],
};

/// Descriptor for `ListDepartmentsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDepartmentsResponseDescriptor = $convert.base64Decode(
    'ChdMaXN0RGVwYXJ0bWVudHNSZXNwb25zZRI7CgtkZXBhcnRtZW50cxgBIAMoCzIZLnNhbGVzb3'
    'JkZXIudjEuRGVwYXJ0bWVudFILZGVwYXJ0bWVudHMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNh'
    'bGVzb3JkZXIudjEuUGFnaW5hdGlvblIKcGFnaW5hdGlvbg==');

@$core.Deprecated('Use getDepartmentRequestDescriptor instead')
const GetDepartmentRequest$json = {
  '1': 'GetDepartmentRequest',
  '2': [
    {'1': 'department_id', '3': 1, '4': 1, '5': 9, '10': 'departmentId'},
  ],
};

/// Descriptor for `GetDepartmentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDepartmentRequestDescriptor = $convert.base64Decode(
    'ChRHZXREZXBhcnRtZW50UmVxdWVzdBIjCg1kZXBhcnRtZW50X2lkGAEgASgJUgxkZXBhcnRtZW'
    '50SWQ=');

@$core.Deprecated('Use getDepartmentResponseDescriptor instead')
const GetDepartmentResponse$json = {
  '1': 'GetDepartmentResponse',
  '2': [
    {
      '1': 'department',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Department',
      '10': 'department'
    },
  ],
};

/// Descriptor for `GetDepartmentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDepartmentResponseDescriptor = $convert.base64Decode(
    'ChVHZXREZXBhcnRtZW50UmVzcG9uc2USOQoKZGVwYXJ0bWVudBgBIAEoCzIZLnNhbGVzb3JkZX'
    'IudjEuRGVwYXJ0bWVudFIKZGVwYXJ0bWVudA==');

@$core.Deprecated('Use createDepartmentRequestDescriptor instead')
const CreateDepartmentRequest$json = {
  '1': 'CreateDepartmentRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `CreateDepartmentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDepartmentRequestDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVEZXBhcnRtZW50UmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SW'
        'QSEgoEbmFtZRgCIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use createDepartmentResponseDescriptor instead')
const CreateDepartmentResponse$json = {
  '1': 'CreateDepartmentResponse',
  '2': [
    {
      '1': 'department',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Department',
      '10': 'department'
    },
  ],
};

/// Descriptor for `CreateDepartmentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDepartmentResponseDescriptor =
    $convert.base64Decode(
        'ChhDcmVhdGVEZXBhcnRtZW50UmVzcG9uc2USOQoKZGVwYXJ0bWVudBgBIAEoCzIZLnNhbGVzb3'
        'JkZXIudjEuRGVwYXJ0bWVudFIKZGVwYXJ0bWVudA==');

@$core.Deprecated('Use updateDepartmentRequestDescriptor instead')
const UpdateDepartmentRequest$json = {
  '1': 'UpdateDepartmentRequest',
  '2': [
    {'1': 'department_id', '3': 1, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'name', '17': true},
  ],
  '8': [
    {'1': '_name'},
  ],
};

/// Descriptor for `UpdateDepartmentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateDepartmentRequestDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVEZXBhcnRtZW50UmVxdWVzdBIjCg1kZXBhcnRtZW50X2lkGAEgASgJUgxkZXBhcn'
        'RtZW50SWQSFwoEbmFtZRgCIAEoCUgAUgRuYW1liAEBQgcKBV9uYW1l');

@$core.Deprecated('Use updateDepartmentResponseDescriptor instead')
const UpdateDepartmentResponse$json = {
  '1': 'UpdateDepartmentResponse',
  '2': [
    {
      '1': 'department',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Department',
      '10': 'department'
    },
  ],
};

/// Descriptor for `UpdateDepartmentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateDepartmentResponseDescriptor =
    $convert.base64Decode(
        'ChhVcGRhdGVEZXBhcnRtZW50UmVzcG9uc2USOQoKZGVwYXJ0bWVudBgBIAEoCzIZLnNhbGVzb3'
        'JkZXIudjEuRGVwYXJ0bWVudFIKZGVwYXJ0bWVudA==');

@$core.Deprecated('Use deleteDepartmentRequestDescriptor instead')
const DeleteDepartmentRequest$json = {
  '1': 'DeleteDepartmentRequest',
  '2': [
    {'1': 'department_id', '3': 1, '4': 1, '5': 9, '10': 'departmentId'},
  ],
};

/// Descriptor for `DeleteDepartmentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteDepartmentRequestDescriptor =
    $convert.base64Decode(
        'ChdEZWxldGVEZXBhcnRtZW50UmVxdWVzdBIjCg1kZXBhcnRtZW50X2lkGAEgASgJUgxkZXBhcn'
        'RtZW50SWQ=');

@$core.Deprecated('Use deleteDepartmentResponseDescriptor instead')
const DeleteDepartmentResponse$json = {
  '1': 'DeleteDepartmentResponse',
};

/// Descriptor for `DeleteDepartmentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteDepartmentResponseDescriptor =
    $convert.base64Decode('ChhEZWxldGVEZXBhcnRtZW50UmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> CompanyServiceBase$json = {
  '1': 'CompanyService',
  '2': [
    {
      '1': 'ListCompanies',
      '2': '.salesorder.v1.ListCompaniesRequest',
      '3': '.salesorder.v1.ListCompaniesResponse'
    },
    {
      '1': 'GetCompany',
      '2': '.salesorder.v1.GetCompanyRequest',
      '3': '.salesorder.v1.GetCompanyResponse'
    },
    {
      '1': 'CreateCompany',
      '2': '.salesorder.v1.CreateCompanyRequest',
      '3': '.salesorder.v1.CreateCompanyResponse'
    },
    {
      '1': 'UpdateCompany',
      '2': '.salesorder.v1.UpdateCompanyRequest',
      '3': '.salesorder.v1.UpdateCompanyResponse'
    },
    {
      '1': 'DeleteCompany',
      '2': '.salesorder.v1.DeleteCompanyRequest',
      '3': '.salesorder.v1.DeleteCompanyResponse'
    },
  ],
};

@$core.Deprecated('Use companyServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    CompanyServiceBase$messageJson = {
  '.salesorder.v1.ListCompaniesRequest': ListCompaniesRequest$json,
  '.salesorder.v1.ListCompaniesResponse': ListCompaniesResponse$json,
  '.salesorder.v1.Company': Company$json,
  '.google.protobuf.Struct': $0.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $0.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $0.Value$json,
  '.google.protobuf.ListValue': $0.ListValue$json,
  '.salesorder.v1.Pagination': $1.Pagination$json,
  '.salesorder.v1.GetCompanyRequest': GetCompanyRequest$json,
  '.salesorder.v1.GetCompanyResponse': GetCompanyResponse$json,
  '.salesorder.v1.CreateCompanyRequest': CreateCompanyRequest$json,
  '.salesorder.v1.CreateCompanyResponse': CreateCompanyResponse$json,
  '.salesorder.v1.UpdateCompanyRequest': UpdateCompanyRequest$json,
  '.salesorder.v1.UpdateCompanyResponse': UpdateCompanyResponse$json,
  '.salesorder.v1.DeleteCompanyRequest': DeleteCompanyRequest$json,
  '.salesorder.v1.DeleteCompanyResponse': DeleteCompanyResponse$json,
};

/// Descriptor for `CompanyService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List companyServiceDescriptor = $convert.base64Decode(
    'Cg5Db21wYW55U2VydmljZRJaCg1MaXN0Q29tcGFuaWVzEiMuc2FsZXNvcmRlci52MS5MaXN0Q2'
    '9tcGFuaWVzUmVxdWVzdBokLnNhbGVzb3JkZXIudjEuTGlzdENvbXBhbmllc1Jlc3BvbnNlElEK'
    'CkdldENvbXBhbnkSIC5zYWxlc29yZGVyLnYxLkdldENvbXBhbnlSZXF1ZXN0GiEuc2FsZXNvcm'
    'Rlci52MS5HZXRDb21wYW55UmVzcG9uc2USWgoNQ3JlYXRlQ29tcGFueRIjLnNhbGVzb3JkZXIu'
    'djEuQ3JlYXRlQ29tcGFueVJlcXVlc3QaJC5zYWxlc29yZGVyLnYxLkNyZWF0ZUNvbXBhbnlSZX'
    'Nwb25zZRJaCg1VcGRhdGVDb21wYW55EiMuc2FsZXNvcmRlci52MS5VcGRhdGVDb21wYW55UmVx'
    'dWVzdBokLnNhbGVzb3JkZXIudjEuVXBkYXRlQ29tcGFueVJlc3BvbnNlEloKDURlbGV0ZUNvbX'
    'BhbnkSIy5zYWxlc29yZGVyLnYxLkRlbGV0ZUNvbXBhbnlSZXF1ZXN0GiQuc2FsZXNvcmRlci52'
    'MS5EZWxldGVDb21wYW55UmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> DepartmentServiceBase$json = {
  '1': 'DepartmentService',
  '2': [
    {
      '1': 'ListDepartments',
      '2': '.salesorder.v1.ListDepartmentsRequest',
      '3': '.salesorder.v1.ListDepartmentsResponse'
    },
    {
      '1': 'GetDepartment',
      '2': '.salesorder.v1.GetDepartmentRequest',
      '3': '.salesorder.v1.GetDepartmentResponse'
    },
    {
      '1': 'CreateDepartment',
      '2': '.salesorder.v1.CreateDepartmentRequest',
      '3': '.salesorder.v1.CreateDepartmentResponse'
    },
    {
      '1': 'UpdateDepartment',
      '2': '.salesorder.v1.UpdateDepartmentRequest',
      '3': '.salesorder.v1.UpdateDepartmentResponse'
    },
    {
      '1': 'DeleteDepartment',
      '2': '.salesorder.v1.DeleteDepartmentRequest',
      '3': '.salesorder.v1.DeleteDepartmentResponse'
    },
  ],
};

@$core.Deprecated('Use departmentServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DepartmentServiceBase$messageJson = {
  '.salesorder.v1.ListDepartmentsRequest': ListDepartmentsRequest$json,
  '.salesorder.v1.ListDepartmentsResponse': ListDepartmentsResponse$json,
  '.salesorder.v1.Department': Department$json,
  '.salesorder.v1.Pagination': $1.Pagination$json,
  '.salesorder.v1.GetDepartmentRequest': GetDepartmentRequest$json,
  '.salesorder.v1.GetDepartmentResponse': GetDepartmentResponse$json,
  '.salesorder.v1.CreateDepartmentRequest': CreateDepartmentRequest$json,
  '.salesorder.v1.CreateDepartmentResponse': CreateDepartmentResponse$json,
  '.salesorder.v1.UpdateDepartmentRequest': UpdateDepartmentRequest$json,
  '.salesorder.v1.UpdateDepartmentResponse': UpdateDepartmentResponse$json,
  '.salesorder.v1.DeleteDepartmentRequest': DeleteDepartmentRequest$json,
  '.salesorder.v1.DeleteDepartmentResponse': DeleteDepartmentResponse$json,
};

/// Descriptor for `DepartmentService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List departmentServiceDescriptor = $convert.base64Decode(
    'ChFEZXBhcnRtZW50U2VydmljZRJgCg9MaXN0RGVwYXJ0bWVudHMSJS5zYWxlc29yZGVyLnYxLk'
    'xpc3REZXBhcnRtZW50c1JlcXVlc3QaJi5zYWxlc29yZGVyLnYxLkxpc3REZXBhcnRtZW50c1Jl'
    'c3BvbnNlEloKDUdldERlcGFydG1lbnQSIy5zYWxlc29yZGVyLnYxLkdldERlcGFydG1lbnRSZX'
    'F1ZXN0GiQuc2FsZXNvcmRlci52MS5HZXREZXBhcnRtZW50UmVzcG9uc2USYwoQQ3JlYXRlRGVw'
    'YXJ0bWVudBImLnNhbGVzb3JkZXIudjEuQ3JlYXRlRGVwYXJ0bWVudFJlcXVlc3QaJy5zYWxlc2'
    '9yZGVyLnYxLkNyZWF0ZURlcGFydG1lbnRSZXNwb25zZRJjChBVcGRhdGVEZXBhcnRtZW50EiYu'
    'c2FsZXNvcmRlci52MS5VcGRhdGVEZXBhcnRtZW50UmVxdWVzdBonLnNhbGVzb3JkZXIudjEuVX'
    'BkYXRlRGVwYXJ0bWVudFJlc3BvbnNlEmMKEERlbGV0ZURlcGFydG1lbnQSJi5zYWxlc29yZGVy'
    'LnYxLkRlbGV0ZURlcGFydG1lbnRSZXF1ZXN0Gicuc2FsZXNvcmRlci52MS5EZWxldGVEZXBhcn'
    'RtZW50UmVzcG9uc2U=');
