// This is a generated file - do not edit.
//
// Generated from salesorder/v1/role.proto.

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

@$core.Deprecated('Use roleDescriptor instead')
const Role$json = {
  '1': 'Role',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'data_scope', '3': 4, '4': 1, '5': 9, '10': 'dataScope'},
    {'1': 'is_system', '3': 5, '4': 1, '5': 8, '10': 'isSystem'},
    {'1': 'is_active', '3': 6, '4': 1, '5': 8, '10': 'isActive'},
  ],
};

/// Descriptor for `Role`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List roleDescriptor = $convert.base64Decode(
    'CgRSb2xlEg4KAmlkGAEgASgJUgJpZBISCgRjb2RlGAIgASgJUgRjb2RlEhIKBG5hbWUYAyABKA'
    'lSBG5hbWUSHQoKZGF0YV9zY29wZRgEIAEoCVIJZGF0YVNjb3BlEhsKCWlzX3N5c3RlbRgFIAEo'
    'CFIIaXNTeXN0ZW0SGwoJaXNfYWN0aXZlGAYgASgIUghpc0FjdGl2ZQ==');

@$core.Deprecated('Use permissionDescriptor instead')
const Permission$json = {
  '1': 'Permission',
  '2': [
    {'1': 'resource', '3': 1, '4': 1, '5': 9, '10': 'resource'},
    {'1': 'action', '3': 2, '4': 1, '5': 9, '10': 'action'},
    {
      '1': 'conditions',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'conditions'
    },
    {'1': 'inverted', '3': 4, '4': 1, '5': 8, '10': 'inverted'},
    {'1': 'sort_order', '3': 5, '4': 1, '5': 5, '10': 'sortOrder'},
  ],
};

/// Descriptor for `Permission`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List permissionDescriptor = $convert.base64Decode(
    'CgpQZXJtaXNzaW9uEhoKCHJlc291cmNlGAEgASgJUghyZXNvdXJjZRIWCgZhY3Rpb24YAiABKA'
    'lSBmFjdGlvbhI3Cgpjb25kaXRpb25zGAMgASgLMhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIK'
    'Y29uZGl0aW9ucxIaCghpbnZlcnRlZBgEIAEoCFIIaW52ZXJ0ZWQSHQoKc29ydF9vcmRlchgFIA'
    'EoBVIJc29ydE9yZGVy');

@$core.Deprecated('Use listRolesRequestDescriptor instead')
const ListRolesRequest$json = {
  '1': 'ListRolesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListRolesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRolesRequestDescriptor = $convert.base64Decode(
    'ChBMaXN0Um9sZXNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaXplGAIgAS'
    'gFUghwYWdlU2l6ZQ==');

@$core.Deprecated('Use listRolesResponseDescriptor instead')
const ListRolesResponse$json = {
  '1': 'ListRolesResponse',
  '2': [
    {
      '1': 'roles',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Role',
      '10': 'roles'
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

/// Descriptor for `ListRolesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRolesResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0Um9sZXNSZXNwb25zZRIpCgVyb2xlcxgBIAMoCzITLnNhbGVzb3JkZXIudjEuUm9sZV'
    'IFcm9sZXMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIudjEuUGFnaW5hdGlvblIK'
    'cGFnaW5hdGlvbg==');

@$core.Deprecated('Use getRolePermissionsRequestDescriptor instead')
const GetRolePermissionsRequest$json = {
  '1': 'GetRolePermissionsRequest',
  '2': [
    {'1': 'role_id', '3': 1, '4': 1, '5': 9, '10': 'roleId'},
  ],
};

/// Descriptor for `GetRolePermissionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRolePermissionsRequestDescriptor =
    $convert.base64Decode(
        'ChlHZXRSb2xlUGVybWlzc2lvbnNSZXF1ZXN0EhcKB3JvbGVfaWQYASABKAlSBnJvbGVJZA==');

@$core.Deprecated('Use getRolePermissionsResponseDescriptor instead')
const GetRolePermissionsResponse$json = {
  '1': 'GetRolePermissionsResponse',
  '2': [
    {
      '1': 'permissions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Permission',
      '10': 'permissions'
    },
  ],
};

/// Descriptor for `GetRolePermissionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRolePermissionsResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRSb2xlUGVybWlzc2lvbnNSZXNwb25zZRI7CgtwZXJtaXNzaW9ucxgBIAMoCzIZLnNhbG'
        'Vzb3JkZXIudjEuUGVybWlzc2lvblILcGVybWlzc2lvbnM=');

@$core.Deprecated('Use updateRolePermissionsRequestDescriptor instead')
const UpdateRolePermissionsRequest$json = {
  '1': 'UpdateRolePermissionsRequest',
  '2': [
    {'1': 'role_id', '3': 1, '4': 1, '5': 9, '10': 'roleId'},
    {
      '1': 'permissions',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Permission',
      '10': 'permissions'
    },
  ],
};

/// Descriptor for `UpdateRolePermissionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateRolePermissionsRequestDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVSb2xlUGVybWlzc2lvbnNSZXF1ZXN0EhcKB3JvbGVfaWQYASABKAlSBnJvbGVJZB'
        'I7CgtwZXJtaXNzaW9ucxgCIAMoCzIZLnNhbGVzb3JkZXIudjEuUGVybWlzc2lvblILcGVybWlz'
        'c2lvbnM=');

@$core.Deprecated('Use updateRolePermissionsResponseDescriptor instead')
const UpdateRolePermissionsResponse$json = {
  '1': 'UpdateRolePermissionsResponse',
  '2': [
    {
      '1': 'permissions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Permission',
      '10': 'permissions'
    },
  ],
};

/// Descriptor for `UpdateRolePermissionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateRolePermissionsResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVSb2xlUGVybWlzc2lvbnNSZXNwb25zZRI7CgtwZXJtaXNzaW9ucxgBIAMoCzIZLn'
        'NhbGVzb3JkZXIudjEuUGVybWlzc2lvblILcGVybWlzc2lvbnM=');

@$core.Deprecated('Use listConditionFieldsRequestDescriptor instead')
const ListConditionFieldsRequest$json = {
  '1': 'ListConditionFieldsRequest',
  '2': [
    {'1': 'resource', '3': 1, '4': 1, '5': 9, '10': 'resource'},
  ],
};

/// Descriptor for `ListConditionFieldsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listConditionFieldsRequestDescriptor =
    $convert.base64Decode(
        'ChpMaXN0Q29uZGl0aW9uRmllbGRzUmVxdWVzdBIaCghyZXNvdXJjZRgBIAEoCVIIcmVzb3VyY2'
        'U=');

@$core.Deprecated('Use conditionFieldDescriptor instead')
const ConditionField$json = {
  '1': 'ConditionField',
  '2': [
    {'1': 'field', '3': 1, '4': 1, '5': 9, '10': 'field'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'ops', '3': 3, '4': 3, '5': 9, '10': 'ops'},
    {'1': 'enum', '3': 4, '4': 3, '5': 9, '10': 'enum'},
  ],
};

/// Descriptor for `ConditionField`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List conditionFieldDescriptor = $convert.base64Decode(
    'Cg5Db25kaXRpb25GaWVsZBIUCgVmaWVsZBgBIAEoCVIFZmllbGQSEgoEdHlwZRgCIAEoCVIEdH'
    'lwZRIQCgNvcHMYAyADKAlSA29wcxISCgRlbnVtGAQgAygJUgRlbnVt');

@$core.Deprecated('Use listConditionFieldsResponseDescriptor instead')
const ListConditionFieldsResponse$json = {
  '1': 'ListConditionFieldsResponse',
  '2': [
    {
      '1': 'fields',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.ConditionField',
      '10': 'fields'
    },
  ],
};

/// Descriptor for `ListConditionFieldsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listConditionFieldsResponseDescriptor =
    $convert.base64Decode(
        'ChtMaXN0Q29uZGl0aW9uRmllbGRzUmVzcG9uc2USNQoGZmllbGRzGAEgAygLMh0uc2FsZXNvcm'
        'Rlci52MS5Db25kaXRpb25GaWVsZFIGZmllbGRz');

const $core.Map<$core.String, $core.dynamic> RoleServiceBase$json = {
  '1': 'RoleService',
  '2': [
    {
      '1': 'ListRoles',
      '2': '.salesorder.v1.ListRolesRequest',
      '3': '.salesorder.v1.ListRolesResponse'
    },
    {
      '1': 'GetRolePermissions',
      '2': '.salesorder.v1.GetRolePermissionsRequest',
      '3': '.salesorder.v1.GetRolePermissionsResponse'
    },
    {
      '1': 'UpdateRolePermissions',
      '2': '.salesorder.v1.UpdateRolePermissionsRequest',
      '3': '.salesorder.v1.UpdateRolePermissionsResponse'
    },
    {
      '1': 'ListConditionFields',
      '2': '.salesorder.v1.ListConditionFieldsRequest',
      '3': '.salesorder.v1.ListConditionFieldsResponse'
    },
  ],
};

@$core.Deprecated('Use roleServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    RoleServiceBase$messageJson = {
  '.salesorder.v1.ListRolesRequest': ListRolesRequest$json,
  '.salesorder.v1.ListRolesResponse': ListRolesResponse$json,
  '.salesorder.v1.Role': Role$json,
  '.salesorder.v1.Pagination': $1.Pagination$json,
  '.salesorder.v1.GetRolePermissionsRequest': GetRolePermissionsRequest$json,
  '.salesorder.v1.GetRolePermissionsResponse': GetRolePermissionsResponse$json,
  '.salesorder.v1.Permission': Permission$json,
  '.google.protobuf.Struct': $0.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $0.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $0.Value$json,
  '.google.protobuf.ListValue': $0.ListValue$json,
  '.salesorder.v1.UpdateRolePermissionsRequest':
      UpdateRolePermissionsRequest$json,
  '.salesorder.v1.UpdateRolePermissionsResponse':
      UpdateRolePermissionsResponse$json,
  '.salesorder.v1.ListConditionFieldsRequest': ListConditionFieldsRequest$json,
  '.salesorder.v1.ListConditionFieldsResponse':
      ListConditionFieldsResponse$json,
  '.salesorder.v1.ConditionField': ConditionField$json,
};

/// Descriptor for `RoleService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List roleServiceDescriptor = $convert.base64Decode(
    'CgtSb2xlU2VydmljZRJOCglMaXN0Um9sZXMSHy5zYWxlc29yZGVyLnYxLkxpc3RSb2xlc1JlcX'
    'Vlc3QaIC5zYWxlc29yZGVyLnYxLkxpc3RSb2xlc1Jlc3BvbnNlEmkKEkdldFJvbGVQZXJtaXNz'
    'aW9ucxIoLnNhbGVzb3JkZXIudjEuR2V0Um9sZVBlcm1pc3Npb25zUmVxdWVzdBopLnNhbGVzb3'
    'JkZXIudjEuR2V0Um9sZVBlcm1pc3Npb25zUmVzcG9uc2UScgoVVXBkYXRlUm9sZVBlcm1pc3Np'
    'b25zEisuc2FsZXNvcmRlci52MS5VcGRhdGVSb2xlUGVybWlzc2lvbnNSZXF1ZXN0Giwuc2FsZX'
    'NvcmRlci52MS5VcGRhdGVSb2xlUGVybWlzc2lvbnNSZXNwb25zZRJsChNMaXN0Q29uZGl0aW9u'
    'RmllbGRzEikuc2FsZXNvcmRlci52MS5MaXN0Q29uZGl0aW9uRmllbGRzUmVxdWVzdBoqLnNhbG'
    'Vzb3JkZXIudjEuTGlzdENvbmRpdGlvbkZpZWxkc1Jlc3BvbnNl');
