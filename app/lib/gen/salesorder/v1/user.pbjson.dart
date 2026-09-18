// This is a generated file - do not edit.
//
// Generated from salesorder/v1/user.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import 'common.pbjson.dart' as $0;

@$core.Deprecated('Use userDescriptor instead')
const User$json = {
  '1': 'User',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'phone', '3': 6, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'employee_no', '3': 7, '4': 1, '5': 9, '10': 'employeeNo'},
    {'1': 'status', '3': 8, '4': 1, '5': 9, '10': 'status'},
    {'1': 'role', '3': 9, '4': 1, '5': 9, '10': 'role'},
    {'1': 'is_customer', '3': 10, '4': 1, '5': 8, '10': 'isCustomer'},
    {'1': 'account_name', '3': 11, '4': 1, '5': 9, '10': 'accountName'},
  ],
};

/// Descriptor for `User`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userDescriptor = $convert.base64Decode(
    'CgRVc2VyEg4KAmlkGAEgASgJUgJpZBIdCgpjb21wYW55X2lkGAIgASgJUgljb21wYW55SWQSIw'
    'oNZGVwYXJ0bWVudF9pZBgDIAEoCVIMZGVwYXJ0bWVudElkEhQKBWVtYWlsGAQgASgJUgVlbWFp'
    'bBISCgRuYW1lGAUgASgJUgRuYW1lEhQKBXBob25lGAYgASgJUgVwaG9uZRIfCgtlbXBsb3llZV'
    '9ubxgHIAEoCVIKZW1wbG95ZWVObxIWCgZzdGF0dXMYCCABKAlSBnN0YXR1cxISCgRyb2xlGAkg'
    'ASgJUgRyb2xlEh8KC2lzX2N1c3RvbWVyGAogASgIUgppc0N1c3RvbWVyEiEKDGFjY291bnRfbm'
    'FtZRgLIAEoCVILYWNjb3VudE5hbWU=');

@$core.Deprecated('Use listUsersRequestDescriptor instead')
const ListUsersRequest$json = {
  '1': 'ListUsersRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'company_id', '3': 3, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 4, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'role', '3': 5, '4': 1, '5': 9, '10': 'role'},
    {'1': 'status', '3': 6, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `ListUsersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listUsersRequestDescriptor = $convert.base64Decode(
    'ChBMaXN0VXNlcnNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaXplGAIgAS'
    'gFUghwYWdlU2l6ZRIdCgpjb21wYW55X2lkGAMgASgJUgljb21wYW55SWQSIwoNZGVwYXJ0bWVu'
    'dF9pZBgEIAEoCVIMZGVwYXJ0bWVudElkEhIKBHJvbGUYBSABKAlSBHJvbGUSFgoGc3RhdHVzGA'
    'YgASgJUgZzdGF0dXM=');

@$core.Deprecated('Use listUsersResponseDescriptor instead')
const ListUsersResponse$json = {
  '1': 'ListUsersResponse',
  '2': [
    {
      '1': 'users',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.User',
      '10': 'users'
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

/// Descriptor for `ListUsersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listUsersResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0VXNlcnNSZXNwb25zZRIpCgV1c2VycxgBIAMoCzITLnNhbGVzb3JkZXIudjEuVXNlcl'
    'IFdXNlcnMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIudjEuUGFnaW5hdGlvblIK'
    'cGFnaW5hdGlvbg==');

@$core.Deprecated('Use getUserRequestDescriptor instead')
const GetUserRequest$json = {
  '1': 'GetUserRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `GetUserRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getUserRequestDescriptor = $convert
    .base64Decode('Cg5HZXRVc2VyUmVxdWVzdBIXCgd1c2VyX2lkGAEgASgJUgZ1c2VySWQ=');

@$core.Deprecated('Use getUserResponseDescriptor instead')
const GetUserResponse$json = {
  '1': 'GetUserResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `GetUserResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getUserResponseDescriptor = $convert.base64Decode(
    'Cg9HZXRVc2VyUmVzcG9uc2USJwoEdXNlchgBIAEoCzITLnNhbGVzb3JkZXIudjEuVXNlclIEdX'
    'Nlcg==');

@$core.Deprecated('Use createUserRequestDescriptor instead')
const CreateUserRequest$json = {
  '1': 'CreateUserRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'email', '3': 2, '4': 1, '5': 9, '10': 'email'},
    {'1': 'company_id', '3': 3, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 4, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'role', '3': 5, '4': 1, '5': 9, '10': 'role'},
    {'1': 'phone', '3': 6, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'employee_no', '3': 7, '4': 1, '5': 9, '10': 'employeeNo'},
  ],
};

/// Descriptor for `CreateUserRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createUserRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVVc2VyUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1lEhQKBWVtYWlsGAIgASgJUg'
    'VlbWFpbBIdCgpjb21wYW55X2lkGAMgASgJUgljb21wYW55SWQSIwoNZGVwYXJ0bWVudF9pZBgE'
    'IAEoCVIMZGVwYXJ0bWVudElkEhIKBHJvbGUYBSABKAlSBHJvbGUSFAoFcGhvbmUYBiABKAlSBX'
    'Bob25lEh8KC2VtcGxveWVlX25vGAcgASgJUgplbXBsb3llZU5v');

@$core.Deprecated('Use createUserResponseDescriptor instead')
const CreateUserResponse$json = {
  '1': 'CreateUserResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `CreateUserResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createUserResponseDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVVc2VyUmVzcG9uc2USJwoEdXNlchgBIAEoCzITLnNhbGVzb3JkZXIudjEuVXNlcl'
    'IEdXNlcg==');

@$core.Deprecated('Use updateUserRequestDescriptor instead')
const UpdateUserRequest$json = {
  '1': 'UpdateUserRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'name', '17': true},
    {
      '1': 'department_id',
      '3': 3,
      '4': 1,
      '5': 9,
      '9': 1,
      '10': 'departmentId',
      '17': true
    },
    {'1': 'phone', '3': 4, '4': 1, '5': 9, '9': 2, '10': 'phone', '17': true},
    {
      '1': 'employee_no',
      '3': 5,
      '4': 1,
      '5': 9,
      '9': 3,
      '10': 'employeeNo',
      '17': true
    },
    {'1': 'status', '3': 6, '4': 1, '5': 9, '9': 4, '10': 'status', '17': true},
  ],
  '8': [
    {'1': '_name'},
    {'1': '_department_id'},
    {'1': '_phone'},
    {'1': '_employee_no'},
    {'1': '_status'},
  ],
};

/// Descriptor for `UpdateUserRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateUserRequestDescriptor = $convert.base64Decode(
    'ChFVcGRhdGVVc2VyUmVxdWVzdBIXCgd1c2VyX2lkGAEgASgJUgZ1c2VySWQSFwoEbmFtZRgCIA'
    'EoCUgAUgRuYW1liAEBEigKDWRlcGFydG1lbnRfaWQYAyABKAlIAVIMZGVwYXJ0bWVudElkiAEB'
    'EhkKBXBob25lGAQgASgJSAJSBXBob25liAEBEiQKC2VtcGxveWVlX25vGAUgASgJSANSCmVtcG'
    'xveWVlTm+IAQESGwoGc3RhdHVzGAYgASgJSARSBnN0YXR1c4gBAUIHCgVfbmFtZUIQCg5fZGVw'
    'YXJ0bWVudF9pZEIICgZfcGhvbmVCDgoMX2VtcGxveWVlX25vQgkKB19zdGF0dXM=');

@$core.Deprecated('Use updateUserResponseDescriptor instead')
const UpdateUserResponse$json = {
  '1': 'UpdateUserResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `UpdateUserResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateUserResponseDescriptor = $convert.base64Decode(
    'ChJVcGRhdGVVc2VyUmVzcG9uc2USJwoEdXNlchgBIAEoCzITLnNhbGVzb3JkZXIudjEuVXNlcl'
    'IEdXNlcg==');

@$core.Deprecated('Use assignRoleRequestDescriptor instead')
const AssignRoleRequest$json = {
  '1': 'AssignRoleRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'role', '3': 2, '4': 1, '5': 9, '10': 'role'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
  ],
};

/// Descriptor for `AssignRoleRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignRoleRequestDescriptor = $convert.base64Decode(
    'ChFBc3NpZ25Sb2xlUmVxdWVzdBIXCgd1c2VyX2lkGAEgASgJUgZ1c2VySWQSEgoEcm9sZRgCIA'
    'EoCVIEcm9sZRIjCg1kZXBhcnRtZW50X2lkGAMgASgJUgxkZXBhcnRtZW50SWQ=');

@$core.Deprecated('Use assignRoleResponseDescriptor instead')
const AssignRoleResponse$json = {
  '1': 'AssignRoleResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `AssignRoleResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignRoleResponseDescriptor = $convert.base64Decode(
    'ChJBc3NpZ25Sb2xlUmVzcG9uc2USJwoEdXNlchgBIAEoCzITLnNhbGVzb3JkZXIudjEuVXNlcl'
    'IEdXNlcg==');

@$core.Deprecated('Use deactivateRequestDescriptor instead')
const DeactivateRequest$json = {
  '1': 'DeactivateRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `DeactivateRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deactivateRequestDescriptor = $convert.base64Decode(
    'ChFEZWFjdGl2YXRlUmVxdWVzdBIXCgd1c2VyX2lkGAEgASgJUgZ1c2VySWQ=');

@$core.Deprecated('Use deactivateResponseDescriptor instead')
const DeactivateResponse$json = {
  '1': 'DeactivateResponse',
};

/// Descriptor for `DeactivateResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deactivateResponseDescriptor =
    $convert.base64Decode('ChJEZWFjdGl2YXRlUmVzcG9uc2U=');

@$core.Deprecated('Use forceLogoutRequestDescriptor instead')
const ForceLogoutRequest$json = {
  '1': 'ForceLogoutRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `ForceLogoutRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List forceLogoutRequestDescriptor =
    $convert.base64Decode(
        'ChJGb3JjZUxvZ291dFJlcXVlc3QSFwoHdXNlcl9pZBgBIAEoCVIGdXNlcklk');

@$core.Deprecated('Use forceLogoutResponseDescriptor instead')
const ForceLogoutResponse$json = {
  '1': 'ForceLogoutResponse',
};

/// Descriptor for `ForceLogoutResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List forceLogoutResponseDescriptor =
    $convert.base64Decode('ChNGb3JjZUxvZ291dFJlc3BvbnNl');

const $core.Map<$core.String, $core.dynamic> UserServiceBase$json = {
  '1': 'UserService',
  '2': [
    {
      '1': 'ListUsers',
      '2': '.salesorder.v1.ListUsersRequest',
      '3': '.salesorder.v1.ListUsersResponse'
    },
    {
      '1': 'GetUser',
      '2': '.salesorder.v1.GetUserRequest',
      '3': '.salesorder.v1.GetUserResponse'
    },
    {
      '1': 'CreateUser',
      '2': '.salesorder.v1.CreateUserRequest',
      '3': '.salesorder.v1.CreateUserResponse'
    },
    {
      '1': 'UpdateUser',
      '2': '.salesorder.v1.UpdateUserRequest',
      '3': '.salesorder.v1.UpdateUserResponse'
    },
    {
      '1': 'AssignRole',
      '2': '.salesorder.v1.AssignRoleRequest',
      '3': '.salesorder.v1.AssignRoleResponse'
    },
    {
      '1': 'Deactivate',
      '2': '.salesorder.v1.DeactivateRequest',
      '3': '.salesorder.v1.DeactivateResponse'
    },
    {
      '1': 'ForceLogout',
      '2': '.salesorder.v1.ForceLogoutRequest',
      '3': '.salesorder.v1.ForceLogoutResponse'
    },
  ],
};

@$core.Deprecated('Use userServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    UserServiceBase$messageJson = {
  '.salesorder.v1.ListUsersRequest': ListUsersRequest$json,
  '.salesorder.v1.ListUsersResponse': ListUsersResponse$json,
  '.salesorder.v1.User': User$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.salesorder.v1.GetUserRequest': GetUserRequest$json,
  '.salesorder.v1.GetUserResponse': GetUserResponse$json,
  '.salesorder.v1.CreateUserRequest': CreateUserRequest$json,
  '.salesorder.v1.CreateUserResponse': CreateUserResponse$json,
  '.salesorder.v1.UpdateUserRequest': UpdateUserRequest$json,
  '.salesorder.v1.UpdateUserResponse': UpdateUserResponse$json,
  '.salesorder.v1.AssignRoleRequest': AssignRoleRequest$json,
  '.salesorder.v1.AssignRoleResponse': AssignRoleResponse$json,
  '.salesorder.v1.DeactivateRequest': DeactivateRequest$json,
  '.salesorder.v1.DeactivateResponse': DeactivateResponse$json,
  '.salesorder.v1.ForceLogoutRequest': ForceLogoutRequest$json,
  '.salesorder.v1.ForceLogoutResponse': ForceLogoutResponse$json,
};

/// Descriptor for `UserService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List userServiceDescriptor = $convert.base64Decode(
    'CgtVc2VyU2VydmljZRJOCglMaXN0VXNlcnMSHy5zYWxlc29yZGVyLnYxLkxpc3RVc2Vyc1JlcX'
    'Vlc3QaIC5zYWxlc29yZGVyLnYxLkxpc3RVc2Vyc1Jlc3BvbnNlEkgKB0dldFVzZXISHS5zYWxl'
    'c29yZGVyLnYxLkdldFVzZXJSZXF1ZXN0Gh4uc2FsZXNvcmRlci52MS5HZXRVc2VyUmVzcG9uc2'
    'USUQoKQ3JlYXRlVXNlchIgLnNhbGVzb3JkZXIudjEuQ3JlYXRlVXNlclJlcXVlc3QaIS5zYWxl'
    'c29yZGVyLnYxLkNyZWF0ZVVzZXJSZXNwb25zZRJRCgpVcGRhdGVVc2VyEiAuc2FsZXNvcmRlci'
    '52MS5VcGRhdGVVc2VyUmVxdWVzdBohLnNhbGVzb3JkZXIudjEuVXBkYXRlVXNlclJlc3BvbnNl'
    'ElEKCkFzc2lnblJvbGUSIC5zYWxlc29yZGVyLnYxLkFzc2lnblJvbGVSZXF1ZXN0GiEuc2FsZX'
    'NvcmRlci52MS5Bc3NpZ25Sb2xlUmVzcG9uc2USUQoKRGVhY3RpdmF0ZRIgLnNhbGVzb3JkZXIu'
    'djEuRGVhY3RpdmF0ZVJlcXVlc3QaIS5zYWxlc29yZGVyLnYxLkRlYWN0aXZhdGVSZXNwb25zZR'
    'JUCgtGb3JjZUxvZ291dBIhLnNhbGVzb3JkZXIudjEuRm9yY2VMb2dvdXRSZXF1ZXN0GiIuc2Fs'
    'ZXNvcmRlci52MS5Gb3JjZUxvZ291dFJlc3BvbnNl');
