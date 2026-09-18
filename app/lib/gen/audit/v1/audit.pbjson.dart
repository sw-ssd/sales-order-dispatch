// This is a generated file - do not edit.
//
// Generated from audit/v1/audit.proto.

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

import '../../salesorder/v1/common.pbjson.dart' as $0;

@$core.Deprecated('Use auditLogDescriptor instead')
const AuditLog$json = {
  '1': 'AuditLog',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'user_id', '3': 4, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'user_name', '3': 5, '4': 1, '5': 9, '10': 'userName'},
    {'1': 'action', '3': 6, '4': 1, '5': 9, '10': 'action'},
    {'1': 'resource_type', '3': 7, '4': 1, '5': 9, '10': 'resourceType'},
    {'1': 'resource_id', '3': 8, '4': 1, '5': 9, '10': 'resourceId'},
    {'1': 'before_snapshot', '3': 9, '4': 1, '5': 9, '10': 'beforeSnapshot'},
    {'1': 'after_snapshot', '3': 10, '4': 1, '5': 9, '10': 'afterSnapshot'},
    {'1': 'ip_address', '3': 11, '4': 1, '5': 9, '10': 'ipAddress'},
    {'1': 'user_agent', '3': 12, '4': 1, '5': 9, '10': 'userAgent'},
    {'1': 'created_at', '3': 13, '4': 1, '5': 9, '10': 'createdAt'},
  ],
};

/// Descriptor for `AuditLog`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List auditLogDescriptor = $convert.base64Decode(
    'CghBdWRpdExvZxIOCgJpZBgBIAEoCVICaWQSHQoKY29tcGFueV9pZBgCIAEoCVIJY29tcGFueU'
    'lkEiMKDWRlcGFydG1lbnRfaWQYAyABKAlSDGRlcGFydG1lbnRJZBIXCgd1c2VyX2lkGAQgASgJ'
    'UgZ1c2VySWQSGwoJdXNlcl9uYW1lGAUgASgJUgh1c2VyTmFtZRIWCgZhY3Rpb24YBiABKAlSBm'
    'FjdGlvbhIjCg1yZXNvdXJjZV90eXBlGAcgASgJUgxyZXNvdXJjZVR5cGUSHwoLcmVzb3VyY2Vf'
    'aWQYCCABKAlSCnJlc291cmNlSWQSJwoPYmVmb3JlX3NuYXBzaG90GAkgASgJUg5iZWZvcmVTbm'
    'Fwc2hvdBIlCg5hZnRlcl9zbmFwc2hvdBgKIAEoCVINYWZ0ZXJTbmFwc2hvdBIdCgppcF9hZGRy'
    'ZXNzGAsgASgJUglpcEFkZHJlc3MSHQoKdXNlcl9hZ2VudBgMIAEoCVIJdXNlckFnZW50Eh0KCm'
    'NyZWF0ZWRfYXQYDSABKAlSCWNyZWF0ZWRBdA==');

@$core.Deprecated('Use listAuditLogsRequestDescriptor instead')
const ListAuditLogsRequest$json = {
  '1': 'ListAuditLogsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'from', '3': 3, '4': 1, '5': 9, '10': 'from'},
    {'1': 'to', '3': 4, '4': 1, '5': 9, '10': 'to'},
    {'1': 'company_id', '3': 5, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'action', '3': 6, '4': 1, '5': 9, '10': 'action'},
    {'1': 'resource_type', '3': 7, '4': 1, '5': 9, '10': 'resourceType'},
    {'1': 'resource_id', '3': 8, '4': 1, '5': 9, '10': 'resourceId'},
    {'1': 'user_id', '3': 9, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `ListAuditLogsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAuditLogsRequestDescriptor = $convert.base64Decode(
    'ChRMaXN0QXVkaXRMb2dzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZR'
    'gCIAEoBVIIcGFnZVNpemUSEgoEZnJvbRgDIAEoCVIEZnJvbRIOCgJ0bxgEIAEoCVICdG8SHQoK'
    'Y29tcGFueV9pZBgFIAEoCVIJY29tcGFueUlkEhYKBmFjdGlvbhgGIAEoCVIGYWN0aW9uEiMKDX'
    'Jlc291cmNlX3R5cGUYByABKAlSDHJlc291cmNlVHlwZRIfCgtyZXNvdXJjZV9pZBgIIAEoCVIK'
    'cmVzb3VyY2VJZBIXCgd1c2VyX2lkGAkgASgJUgZ1c2VySWQ=');

@$core.Deprecated('Use listAuditLogsResponseDescriptor instead')
const ListAuditLogsResponse$json = {
  '1': 'ListAuditLogsResponse',
  '2': [
    {
      '1': 'items',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.audit.v1.AuditLog',
      '10': 'items'
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

/// Descriptor for `ListAuditLogsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAuditLogsResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0QXVkaXRMb2dzUmVzcG9uc2USKAoFaXRlbXMYASADKAsyEi5hdWRpdC52MS5BdWRpdE'
    'xvZ1IFaXRlbXMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIudjEuUGFnaW5hdGlv'
    'blIKcGFnaW5hdGlvbg==');

const $core.Map<$core.String, $core.dynamic> AuditServiceBase$json = {
  '1': 'AuditService',
  '2': [
    {
      '1': 'ListAuditLogs',
      '2': '.audit.v1.ListAuditLogsRequest',
      '3': '.audit.v1.ListAuditLogsResponse'
    },
  ],
};

@$core.Deprecated('Use auditServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AuditServiceBase$messageJson = {
  '.audit.v1.ListAuditLogsRequest': ListAuditLogsRequest$json,
  '.audit.v1.ListAuditLogsResponse': ListAuditLogsResponse$json,
  '.audit.v1.AuditLog': AuditLog$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
};

/// Descriptor for `AuditService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List auditServiceDescriptor = $convert.base64Decode(
    'CgxBdWRpdFNlcnZpY2USUAoNTGlzdEF1ZGl0TG9ncxIeLmF1ZGl0LnYxLkxpc3RBdWRpdExvZ3'
    'NSZXF1ZXN0Gh8uYXVkaXQudjEuTGlzdEF1ZGl0TG9nc1Jlc3BvbnNl');
