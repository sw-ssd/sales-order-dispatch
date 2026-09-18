// This is a generated file - do not edit.
//
// Generated from masters/v1/master.proto.

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

import 'package:protobuf/well_known_types/google/protobuf/struct.pbjson.dart'
    as $1;

import '../../salesorder/v1/common.pbjson.dart' as $0;

@$core.Deprecated('Use warehouseDescriptor instead')
const Warehouse$json = {
  '1': 'Warehouse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'code', '3': 4, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'address', '3': 6, '4': 1, '5': 9, '10': 'address'},
    {'1': 'is_active', '3': 7, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'created_at', '3': 8, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 9, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 10, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `Warehouse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List warehouseDescriptor = $convert.base64Decode(
    'CglXYXJlaG91c2USDgoCaWQYASABKAlSAmlkEh0KCmNvbXBhbnlfaWQYAiABKAlSCWNvbXBhbn'
    'lJZBIjCg1kZXBhcnRtZW50X2lkGAMgASgJUgxkZXBhcnRtZW50SWQSEgoEY29kZRgEIAEoCVIE'
    'Y29kZRISCgRuYW1lGAUgASgJUgRuYW1lEhgKB2FkZHJlc3MYBiABKAlSB2FkZHJlc3MSGwoJaX'
    'NfYWN0aXZlGAcgASgIUghpc0FjdGl2ZRIdCgpjcmVhdGVkX2F0GAggASgJUgljcmVhdGVkQXQS'
    'HQoKdXBkYXRlZF9hdBgJIAEoCVIJdXBkYXRlZEF0Eh0KCmRlbGV0ZWRfYXQYCiABKAlSCWRlbG'
    'V0ZWRBdA==');

@$core.Deprecated('Use listWarehousesRequestDescriptor instead')
const ListWarehousesRequest$json = {
  '1': 'ListWarehousesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'include_deleted', '3': 4, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListWarehousesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWarehousesRequestDescriptor = $convert.base64Decode(
    'ChVMaXN0V2FyZWhvdXNlc1JlcXVlc3QSEgoEcGFnZRgBIAEoBVIEcGFnZRIbCglwYWdlX3Npem'
    'UYAiABKAVSCHBhZ2VTaXplEhgKB2tleXdvcmQYAyABKAlSB2tleXdvcmQSJwoPaW5jbHVkZV9k'
    'ZWxldGVkGAQgASgIUg5pbmNsdWRlRGVsZXRlZA==');

@$core.Deprecated('Use listWarehousesResponseDescriptor instead')
const ListWarehousesResponse$json = {
  '1': 'ListWarehousesResponse',
  '2': [
    {
      '1': 'warehouses',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.masters.v1.Warehouse',
      '10': 'warehouses'
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

/// Descriptor for `ListWarehousesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWarehousesResponseDescriptor = $convert.base64Decode(
    'ChZMaXN0V2FyZWhvdXNlc1Jlc3BvbnNlEjUKCndhcmVob3VzZXMYASADKAsyFS5tYXN0ZXJzLn'
    'YxLldhcmVob3VzZVIKd2FyZWhvdXNlcxI5CgpwYWdpbmF0aW9uGAIgASgLMhkuc2FsZXNvcmRl'
    'ci52MS5QYWdpbmF0aW9uUgpwYWdpbmF0aW9u');

@$core.Deprecated('Use createWarehouseRequestDescriptor instead')
const CreateWarehouseRequest$json = {
  '1': 'CreateWarehouseRequest',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'address', '3': 3, '4': 1, '5': 9, '10': 'address'},
    {'1': 'is_active', '3': 4, '4': 1, '5': 8, '10': 'isActive'},
  ],
};

/// Descriptor for `CreateWarehouseRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWarehouseRequestDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVXYXJlaG91c2VSZXF1ZXN0EhIKBGNvZGUYASABKAlSBGNvZGUSEgoEbmFtZRgCIA'
    'EoCVIEbmFtZRIYCgdhZGRyZXNzGAMgASgJUgdhZGRyZXNzEhsKCWlzX2FjdGl2ZRgEIAEoCFII'
    'aXNBY3RpdmU=');

@$core.Deprecated('Use createWarehouseResponseDescriptor instead')
const CreateWarehouseResponse$json = {
  '1': 'CreateWarehouseResponse',
  '2': [
    {
      '1': 'warehouse',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.Warehouse',
      '10': 'warehouse'
    },
  ],
};

/// Descriptor for `CreateWarehouseResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWarehouseResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVXYXJlaG91c2VSZXNwb25zZRIzCgl3YXJlaG91c2UYASABKAsyFS5tYXN0ZXJzLn'
        'YxLldhcmVob3VzZVIJd2FyZWhvdXNl');

@$core.Deprecated('Use updateWarehouseRequestDescriptor instead')
const UpdateWarehouseRequest$json = {
  '1': 'UpdateWarehouseRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'code', '17': true},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'name', '17': true},
    {
      '1': 'address',
      '3': 4,
      '4': 1,
      '5': 9,
      '9': 2,
      '10': 'address',
      '17': true
    },
    {
      '1': 'is_active',
      '3': 5,
      '4': 1,
      '5': 8,
      '9': 3,
      '10': 'isActive',
      '17': true
    },
  ],
  '8': [
    {'1': '_code'},
    {'1': '_name'},
    {'1': '_address'},
    {'1': '_is_active'},
  ],
};

/// Descriptor for `UpdateWarehouseRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateWarehouseRequestDescriptor = $convert.base64Decode(
    'ChZVcGRhdGVXYXJlaG91c2VSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIXCgRjb2RlGAIgASgJSA'
    'BSBGNvZGWIAQESFwoEbmFtZRgDIAEoCUgBUgRuYW1liAEBEh0KB2FkZHJlc3MYBCABKAlIAlIH'
    'YWRkcmVzc4gBARIgCglpc19hY3RpdmUYBSABKAhIA1IIaXNBY3RpdmWIAQFCBwoFX2NvZGVCBw'
    'oFX25hbWVCCgoIX2FkZHJlc3NCDAoKX2lzX2FjdGl2ZQ==');

@$core.Deprecated('Use updateWarehouseResponseDescriptor instead')
const UpdateWarehouseResponse$json = {
  '1': 'UpdateWarehouseResponse',
  '2': [
    {
      '1': 'warehouse',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.Warehouse',
      '10': 'warehouse'
    },
  ],
};

/// Descriptor for `UpdateWarehouseResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateWarehouseResponseDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVXYXJlaG91c2VSZXNwb25zZRIzCgl3YXJlaG91c2UYASABKAsyFS5tYXN0ZXJzLn'
        'YxLldhcmVob3VzZVIJd2FyZWhvdXNl');

@$core.Deprecated('Use deleteWarehouseRequestDescriptor instead')
const DeleteWarehouseRequest$json = {
  '1': 'DeleteWarehouseRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteWarehouseRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWarehouseRequestDescriptor = $convert
    .base64Decode('ChZEZWxldGVXYXJlaG91c2VSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use deleteWarehouseResponseDescriptor instead')
const DeleteWarehouseResponse$json = {
  '1': 'DeleteWarehouseResponse',
};

/// Descriptor for `DeleteWarehouseResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWarehouseResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVXYXJlaG91c2VSZXNwb25zZQ==');

@$core.Deprecated('Use restoreWarehouseRequestDescriptor instead')
const RestoreWarehouseRequest$json = {
  '1': 'RestoreWarehouseRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `RestoreWarehouseRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreWarehouseRequestDescriptor = $convert
    .base64Decode('ChdSZXN0b3JlV2FyZWhvdXNlUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use restoreWarehouseResponseDescriptor instead')
const RestoreWarehouseResponse$json = {
  '1': 'RestoreWarehouseResponse',
  '2': [
    {
      '1': 'warehouse',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.Warehouse',
      '10': 'warehouse'
    },
  ],
};

/// Descriptor for `RestoreWarehouseResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreWarehouseResponseDescriptor =
    $convert.base64Decode(
        'ChhSZXN0b3JlV2FyZWhvdXNlUmVzcG9uc2USMwoJd2FyZWhvdXNlGAEgASgLMhUubWFzdGVycy'
        '52MS5XYXJlaG91c2VSCXdhcmVob3VzZQ==');

@$core.Deprecated('Use routeDescriptor instead')
const Route$json = {
  '1': 'Route',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'code', '3': 4, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 6, '4': 1, '5': 9, '10': 'description'},
    {'1': 'sort_order', '3': 7, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 8, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'created_at', '3': 9, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 10, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 11, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `Route`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List routeDescriptor = $convert.base64Decode(
    'CgVSb3V0ZRIOCgJpZBgBIAEoCVICaWQSHQoKY29tcGFueV9pZBgCIAEoCVIJY29tcGFueUlkEi'
    'MKDWRlcGFydG1lbnRfaWQYAyABKAlSDGRlcGFydG1lbnRJZBISCgRjb2RlGAQgASgJUgRjb2Rl'
    'EhIKBG5hbWUYBSABKAlSBG5hbWUSIAoLZGVzY3JpcHRpb24YBiABKAlSC2Rlc2NyaXB0aW9uEh'
    '0KCnNvcnRfb3JkZXIYByABKAVSCXNvcnRPcmRlchIbCglpc19hY3RpdmUYCCABKAhSCGlzQWN0'
    'aXZlEh0KCmNyZWF0ZWRfYXQYCSABKAlSCWNyZWF0ZWRBdBIdCgp1cGRhdGVkX2F0GAogASgJUg'
    'l1cGRhdGVkQXQSHQoKZGVsZXRlZF9hdBgLIAEoCVIJZGVsZXRlZEF0');

@$core.Deprecated('Use listRoutesRequestDescriptor instead')
const ListRoutesRequest$json = {
  '1': 'ListRoutesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'include_deleted', '3': 4, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListRoutesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRoutesRequestDescriptor = $convert.base64Decode(
    'ChFMaXN0Um91dGVzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZRgCIA'
    'EoBVIIcGFnZVNpemUSGAoHa2V5d29yZBgDIAEoCVIHa2V5d29yZBInCg9pbmNsdWRlX2RlbGV0'
    'ZWQYBCABKAhSDmluY2x1ZGVEZWxldGVk');

@$core.Deprecated('Use listRoutesResponseDescriptor instead')
const ListRoutesResponse$json = {
  '1': 'ListRoutesResponse',
  '2': [
    {
      '1': 'routes',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.masters.v1.Route',
      '10': 'routes'
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

/// Descriptor for `ListRoutesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRoutesResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0Um91dGVzUmVzcG9uc2USKQoGcm91dGVzGAEgAygLMhEubWFzdGVycy52MS5Sb3V0ZV'
    'IGcm91dGVzEjkKCnBhZ2luYXRpb24YAiABKAsyGS5zYWxlc29yZGVyLnYxLlBhZ2luYXRpb25S'
    'CnBhZ2luYXRpb24=');

@$core.Deprecated('Use createRouteRequestDescriptor instead')
const CreateRouteRequest$json = {
  '1': 'CreateRouteRequest',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {'1': 'sort_order', '3': 4, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 5, '4': 1, '5': 8, '10': 'isActive'},
  ],
};

/// Descriptor for `CreateRouteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createRouteRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVSb3V0ZVJlcXVlc3QSEgoEY29kZRgBIAEoCVIEY29kZRISCgRuYW1lGAIgASgJUg'
    'RuYW1lEiAKC2Rlc2NyaXB0aW9uGAMgASgJUgtkZXNjcmlwdGlvbhIdCgpzb3J0X29yZGVyGAQg'
    'ASgFUglzb3J0T3JkZXISGwoJaXNfYWN0aXZlGAUgASgIUghpc0FjdGl2ZQ==');

@$core.Deprecated('Use createRouteResponseDescriptor instead')
const CreateRouteResponse$json = {
  '1': 'CreateRouteResponse',
  '2': [
    {
      '1': 'route',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.Route',
      '10': 'route'
    },
  ],
};

/// Descriptor for `CreateRouteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createRouteResponseDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVSb3V0ZVJlc3BvbnNlEicKBXJvdXRlGAEgASgLMhEubWFzdGVycy52MS5Sb3V0ZV'
    'IFcm91dGU=');

@$core.Deprecated('Use updateRouteRequestDescriptor instead')
const UpdateRouteRequest$json = {
  '1': 'UpdateRouteRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'code', '17': true},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'name', '17': true},
    {
      '1': 'description',
      '3': 4,
      '4': 1,
      '5': 9,
      '9': 2,
      '10': 'description',
      '17': true
    },
    {
      '1': 'sort_order',
      '3': 5,
      '4': 1,
      '5': 5,
      '9': 3,
      '10': 'sortOrder',
      '17': true
    },
    {
      '1': 'is_active',
      '3': 6,
      '4': 1,
      '5': 8,
      '9': 4,
      '10': 'isActive',
      '17': true
    },
  ],
  '8': [
    {'1': '_code'},
    {'1': '_name'},
    {'1': '_description'},
    {'1': '_sort_order'},
    {'1': '_is_active'},
  ],
};

/// Descriptor for `UpdateRouteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateRouteRequestDescriptor = $convert.base64Decode(
    'ChJVcGRhdGVSb3V0ZVJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhcKBGNvZGUYAiABKAlIAFIEY2'
    '9kZYgBARIXCgRuYW1lGAMgASgJSAFSBG5hbWWIAQESJQoLZGVzY3JpcHRpb24YBCABKAlIAlIL'
    'ZGVzY3JpcHRpb26IAQESIgoKc29ydF9vcmRlchgFIAEoBUgDUglzb3J0T3JkZXKIAQESIAoJaX'
    'NfYWN0aXZlGAYgASgISARSCGlzQWN0aXZliAEBQgcKBV9jb2RlQgcKBV9uYW1lQg4KDF9kZXNj'
    'cmlwdGlvbkINCgtfc29ydF9vcmRlckIMCgpfaXNfYWN0aXZl');

@$core.Deprecated('Use updateRouteResponseDescriptor instead')
const UpdateRouteResponse$json = {
  '1': 'UpdateRouteResponse',
  '2': [
    {
      '1': 'route',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.Route',
      '10': 'route'
    },
  ],
};

/// Descriptor for `UpdateRouteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateRouteResponseDescriptor = $convert.base64Decode(
    'ChNVcGRhdGVSb3V0ZVJlc3BvbnNlEicKBXJvdXRlGAEgASgLMhEubWFzdGVycy52MS5Sb3V0ZV'
    'IFcm91dGU=');

@$core.Deprecated('Use deleteRouteRequestDescriptor instead')
const DeleteRouteRequest$json = {
  '1': 'DeleteRouteRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteRouteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteRouteRequestDescriptor =
    $convert.base64Decode('ChJEZWxldGVSb3V0ZVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteRouteResponseDescriptor instead')
const DeleteRouteResponse$json = {
  '1': 'DeleteRouteResponse',
};

/// Descriptor for `DeleteRouteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteRouteResponseDescriptor =
    $convert.base64Decode('ChNEZWxldGVSb3V0ZVJlc3BvbnNl');

@$core.Deprecated('Use restoreRouteRequestDescriptor instead')
const RestoreRouteRequest$json = {
  '1': 'RestoreRouteRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `RestoreRouteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreRouteRequestDescriptor = $convert
    .base64Decode('ChNSZXN0b3JlUm91dGVSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use restoreRouteResponseDescriptor instead')
const RestoreRouteResponse$json = {
  '1': 'RestoreRouteResponse',
  '2': [
    {
      '1': 'route',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.Route',
      '10': 'route'
    },
  ],
};

/// Descriptor for `RestoreRouteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreRouteResponseDescriptor = $convert.base64Decode(
    'ChRSZXN0b3JlUm91dGVSZXNwb25zZRInCgVyb3V0ZRgBIAEoCzIRLm1hc3RlcnMudjEuUm91dG'
    'VSBXJvdXRl');

@$core.Deprecated('Use processingSpecDescriptor instead')
const ProcessingSpec$json = {
  '1': 'ProcessingSpec',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'code', '3': 4, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'kind', '3': 6, '4': 1, '5': 9, '10': 'kind'},
    {
      '1': 'applies_to_processing',
      '3': 7,
      '4': 1,
      '5': 8,
      '10': 'appliesToProcessing'
    },
    {
      '1': 'applies_to_picking',
      '3': 8,
      '4': 1,
      '5': 8,
      '10': 'appliesToPicking'
    },
    {
      '1': 'attributes',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'attributes'
    },
    {'1': 'sort_order', '3': 10, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 11, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'created_at', '3': 12, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 13, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 14, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `ProcessingSpec`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List processingSpecDescriptor = $convert.base64Decode(
    'Cg5Qcm9jZXNzaW5nU3BlYxIOCgJpZBgBIAEoCVICaWQSHQoKY29tcGFueV9pZBgCIAEoCVIJY2'
    '9tcGFueUlkEiMKDWRlcGFydG1lbnRfaWQYAyABKAlSDGRlcGFydG1lbnRJZBISCgRjb2RlGAQg'
    'ASgJUgRjb2RlEhIKBG5hbWUYBSABKAlSBG5hbWUSEgoEa2luZBgGIAEoCVIEa2luZBIyChVhcH'
    'BsaWVzX3RvX3Byb2Nlc3NpbmcYByABKAhSE2FwcGxpZXNUb1Byb2Nlc3NpbmcSLAoSYXBwbGll'
    'c190b19waWNraW5nGAggASgIUhBhcHBsaWVzVG9QaWNraW5nEjcKCmF0dHJpYnV0ZXMYCSABKA'
    'syFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgphdHRyaWJ1dGVzEh0KCnNvcnRfb3JkZXIYCiAB'
    'KAVSCXNvcnRPcmRlchIbCglpc19hY3RpdmUYCyABKAhSCGlzQWN0aXZlEh0KCmNyZWF0ZWRfYX'
    'QYDCABKAlSCWNyZWF0ZWRBdBIdCgp1cGRhdGVkX2F0GA0gASgJUgl1cGRhdGVkQXQSHQoKZGVs'
    'ZXRlZF9hdBgOIAEoCVIJZGVsZXRlZEF0');

@$core.Deprecated('Use listProcessingSpecsRequestDescriptor instead')
const ListProcessingSpecsRequest$json = {
  '1': 'ListProcessingSpecsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'include_deleted', '3': 4, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListProcessingSpecsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProcessingSpecsRequestDescriptor =
    $convert.base64Decode(
        'ChpMaXN0UHJvY2Vzc2luZ1NwZWNzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2'
        'Vfc2l6ZRgCIAEoBVIIcGFnZVNpemUSGAoHa2V5d29yZBgDIAEoCVIHa2V5d29yZBInCg9pbmNs'
        'dWRlX2RlbGV0ZWQYBCABKAhSDmluY2x1ZGVEZWxldGVk');

@$core.Deprecated('Use listProcessingSpecsResponseDescriptor instead')
const ListProcessingSpecsResponse$json = {
  '1': 'ListProcessingSpecsResponse',
  '2': [
    {
      '1': 'processing_specs',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.masters.v1.ProcessingSpec',
      '10': 'processingSpecs'
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

/// Descriptor for `ListProcessingSpecsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProcessingSpecsResponseDescriptor =
    $convert.base64Decode(
        'ChtMaXN0UHJvY2Vzc2luZ1NwZWNzUmVzcG9uc2USRQoQcHJvY2Vzc2luZ19zcGVjcxgBIAMoCz'
        'IaLm1hc3RlcnMudjEuUHJvY2Vzc2luZ1NwZWNSD3Byb2Nlc3NpbmdTcGVjcxI5CgpwYWdpbmF0'
        'aW9uGAIgASgLMhkuc2FsZXNvcmRlci52MS5QYWdpbmF0aW9uUgpwYWdpbmF0aW9u');

@$core.Deprecated('Use createProcessingSpecRequestDescriptor instead')
const CreateProcessingSpecRequest$json = {
  '1': 'CreateProcessingSpecRequest',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'kind', '3': 3, '4': 1, '5': 9, '10': 'kind'},
    {
      '1': 'applies_to_processing',
      '3': 4,
      '4': 1,
      '5': 8,
      '10': 'appliesToProcessing'
    },
    {
      '1': 'applies_to_picking',
      '3': 5,
      '4': 1,
      '5': 8,
      '10': 'appliesToPicking'
    },
    {
      '1': 'attributes',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'attributes'
    },
    {'1': 'sort_order', '3': 7, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 8, '4': 1, '5': 8, '10': 'isActive'},
  ],
};

/// Descriptor for `CreateProcessingSpecRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createProcessingSpecRequestDescriptor = $convert.base64Decode(
    'ChtDcmVhdGVQcm9jZXNzaW5nU3BlY1JlcXVlc3QSEgoEY29kZRgBIAEoCVIEY29kZRISCgRuYW'
    '1lGAIgASgJUgRuYW1lEhIKBGtpbmQYAyABKAlSBGtpbmQSMgoVYXBwbGllc190b19wcm9jZXNz'
    'aW5nGAQgASgIUhNhcHBsaWVzVG9Qcm9jZXNzaW5nEiwKEmFwcGxpZXNfdG9fcGlja2luZxgFIA'
    'EoCFIQYXBwbGllc1RvUGlja2luZxI3CgphdHRyaWJ1dGVzGAYgASgLMhcuZ29vZ2xlLnByb3Rv'
    'YnVmLlN0cnVjdFIKYXR0cmlidXRlcxIdCgpzb3J0X29yZGVyGAcgASgFUglzb3J0T3JkZXISGw'
    'oJaXNfYWN0aXZlGAggASgIUghpc0FjdGl2ZQ==');

@$core.Deprecated('Use createProcessingSpecResponseDescriptor instead')
const CreateProcessingSpecResponse$json = {
  '1': 'CreateProcessingSpecResponse',
  '2': [
    {
      '1': 'processing_spec',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.ProcessingSpec',
      '10': 'processingSpec'
    },
  ],
};

/// Descriptor for `CreateProcessingSpecResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createProcessingSpecResponseDescriptor =
    $convert.base64Decode(
        'ChxDcmVhdGVQcm9jZXNzaW5nU3BlY1Jlc3BvbnNlEkMKD3Byb2Nlc3Npbmdfc3BlYxgBIAEoCz'
        'IaLm1hc3RlcnMudjEuUHJvY2Vzc2luZ1NwZWNSDnByb2Nlc3NpbmdTcGVj');

@$core.Deprecated('Use updateProcessingSpecRequestDescriptor instead')
const UpdateProcessingSpecRequest$json = {
  '1': 'UpdateProcessingSpecRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'code', '17': true},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'name', '17': true},
    {'1': 'kind', '3': 4, '4': 1, '5': 9, '9': 2, '10': 'kind', '17': true},
    {
      '1': 'applies_to_processing',
      '3': 5,
      '4': 1,
      '5': 8,
      '9': 3,
      '10': 'appliesToProcessing',
      '17': true
    },
    {
      '1': 'applies_to_picking',
      '3': 6,
      '4': 1,
      '5': 8,
      '9': 4,
      '10': 'appliesToPicking',
      '17': true
    },
    {
      '1': 'attributes',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'attributes'
    },
    {
      '1': 'sort_order',
      '3': 8,
      '4': 1,
      '5': 5,
      '9': 5,
      '10': 'sortOrder',
      '17': true
    },
    {
      '1': 'is_active',
      '3': 9,
      '4': 1,
      '5': 8,
      '9': 6,
      '10': 'isActive',
      '17': true
    },
  ],
  '8': [
    {'1': '_code'},
    {'1': '_name'},
    {'1': '_kind'},
    {'1': '_applies_to_processing'},
    {'1': '_applies_to_picking'},
    {'1': '_sort_order'},
    {'1': '_is_active'},
  ],
};

/// Descriptor for `UpdateProcessingSpecRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProcessingSpecRequestDescriptor = $convert.base64Decode(
    'ChtVcGRhdGVQcm9jZXNzaW5nU3BlY1JlcXVlc3QSDgoCaWQYASABKAlSAmlkEhcKBGNvZGUYAi'
    'ABKAlIAFIEY29kZYgBARIXCgRuYW1lGAMgASgJSAFSBG5hbWWIAQESFwoEa2luZBgEIAEoCUgC'
    'UgRraW5kiAEBEjcKFWFwcGxpZXNfdG9fcHJvY2Vzc2luZxgFIAEoCEgDUhNhcHBsaWVzVG9Qcm'
    '9jZXNzaW5niAEBEjEKEmFwcGxpZXNfdG9fcGlja2luZxgGIAEoCEgEUhBhcHBsaWVzVG9QaWNr'
    'aW5niAEBEjcKCmF0dHJpYnV0ZXMYByABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0UgphdH'
    'RyaWJ1dGVzEiIKCnNvcnRfb3JkZXIYCCABKAVIBVIJc29ydE9yZGVyiAEBEiAKCWlzX2FjdGl2'
    'ZRgJIAEoCEgGUghpc0FjdGl2ZYgBAUIHCgVfY29kZUIHCgVfbmFtZUIHCgVfa2luZEIYChZfYX'
    'BwbGllc190b19wcm9jZXNzaW5nQhUKE19hcHBsaWVzX3RvX3BpY2tpbmdCDQoLX3NvcnRfb3Jk'
    'ZXJCDAoKX2lzX2FjdGl2ZQ==');

@$core.Deprecated('Use updateProcessingSpecResponseDescriptor instead')
const UpdateProcessingSpecResponse$json = {
  '1': 'UpdateProcessingSpecResponse',
  '2': [
    {
      '1': 'processing_spec',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.ProcessingSpec',
      '10': 'processingSpec'
    },
  ],
};

/// Descriptor for `UpdateProcessingSpecResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProcessingSpecResponseDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVQcm9jZXNzaW5nU3BlY1Jlc3BvbnNlEkMKD3Byb2Nlc3Npbmdfc3BlYxgBIAEoCz'
        'IaLm1hc3RlcnMudjEuUHJvY2Vzc2luZ1NwZWNSDnByb2Nlc3NpbmdTcGVj');

@$core.Deprecated('Use deleteProcessingSpecRequestDescriptor instead')
const DeleteProcessingSpecRequest$json = {
  '1': 'DeleteProcessingSpecRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteProcessingSpecRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteProcessingSpecRequestDescriptor =
    $convert.base64Decode(
        'ChtEZWxldGVQcm9jZXNzaW5nU3BlY1JlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteProcessingSpecResponseDescriptor instead')
const DeleteProcessingSpecResponse$json = {
  '1': 'DeleteProcessingSpecResponse',
};

/// Descriptor for `DeleteProcessingSpecResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteProcessingSpecResponseDescriptor =
    $convert.base64Decode('ChxEZWxldGVQcm9jZXNzaW5nU3BlY1Jlc3BvbnNl');

@$core.Deprecated('Use restoreProcessingSpecRequestDescriptor instead')
const RestoreProcessingSpecRequest$json = {
  '1': 'RestoreProcessingSpecRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `RestoreProcessingSpecRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreProcessingSpecRequestDescriptor =
    $convert.base64Decode(
        'ChxSZXN0b3JlUHJvY2Vzc2luZ1NwZWNSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use restoreProcessingSpecResponseDescriptor instead')
const RestoreProcessingSpecResponse$json = {
  '1': 'RestoreProcessingSpecResponse',
  '2': [
    {
      '1': 'processing_spec',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.ProcessingSpec',
      '10': 'processingSpec'
    },
  ],
};

/// Descriptor for `RestoreProcessingSpecResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreProcessingSpecResponseDescriptor =
    $convert.base64Decode(
        'Ch1SZXN0b3JlUHJvY2Vzc2luZ1NwZWNSZXNwb25zZRJDCg9wcm9jZXNzaW5nX3NwZWMYASABKA'
        'syGi5tYXN0ZXJzLnYxLlByb2Nlc3NpbmdTcGVjUg5wcm9jZXNzaW5nU3BlYw==');

@$core.Deprecated('Use productCategoryDescriptor instead')
const ProductCategory$json = {
  '1': 'ProductCategory',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'code', '3': 4, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'sort_order', '3': 6, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 7, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'created_at', '3': 8, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 9, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 10, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `ProductCategory`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List productCategoryDescriptor = $convert.base64Decode(
    'Cg9Qcm9kdWN0Q2F0ZWdvcnkSDgoCaWQYASABKAlSAmlkEh0KCmNvbXBhbnlfaWQYAiABKAlSCW'
    'NvbXBhbnlJZBIjCg1kZXBhcnRtZW50X2lkGAMgASgJUgxkZXBhcnRtZW50SWQSEgoEY29kZRgE'
    'IAEoCVIEY29kZRISCgRuYW1lGAUgASgJUgRuYW1lEh0KCnNvcnRfb3JkZXIYBiABKAVSCXNvcn'
    'RPcmRlchIbCglpc19hY3RpdmUYByABKAhSCGlzQWN0aXZlEh0KCmNyZWF0ZWRfYXQYCCABKAlS'
    'CWNyZWF0ZWRBdBIdCgp1cGRhdGVkX2F0GAkgASgJUgl1cGRhdGVkQXQSHQoKZGVsZXRlZF9hdB'
    'gKIAEoCVIJZGVsZXRlZEF0');

@$core.Deprecated('Use listProductCategoriesRequestDescriptor instead')
const ListProductCategoriesRequest$json = {
  '1': 'ListProductCategoriesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'include_deleted', '3': 4, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListProductCategoriesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProductCategoriesRequestDescriptor =
    $convert.base64Decode(
        'ChxMaXN0UHJvZHVjdENhdGVnb3JpZXNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcG'
        'FnZV9zaXplGAIgASgFUghwYWdlU2l6ZRIYCgdrZXl3b3JkGAMgASgJUgdrZXl3b3JkEicKD2lu'
        'Y2x1ZGVfZGVsZXRlZBgEIAEoCFIOaW5jbHVkZURlbGV0ZWQ=');

@$core.Deprecated('Use listProductCategoriesResponseDescriptor instead')
const ListProductCategoriesResponse$json = {
  '1': 'ListProductCategoriesResponse',
  '2': [
    {
      '1': 'product_categories',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.masters.v1.ProductCategory',
      '10': 'productCategories'
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

/// Descriptor for `ListProductCategoriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProductCategoriesResponseDescriptor = $convert.base64Decode(
    'Ch1MaXN0UHJvZHVjdENhdGVnb3JpZXNSZXNwb25zZRJKChJwcm9kdWN0X2NhdGVnb3JpZXMYAS'
    'ADKAsyGy5tYXN0ZXJzLnYxLlByb2R1Y3RDYXRlZ29yeVIRcHJvZHVjdENhdGVnb3JpZXMSOQoK'
    'cGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIudjEuUGFnaW5hdGlvblIKcGFnaW5hdGlvbg'
    '==');

@$core.Deprecated('Use createProductCategoryRequestDescriptor instead')
const CreateProductCategoryRequest$json = {
  '1': 'CreateProductCategoryRequest',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'sort_order', '3': 3, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 4, '4': 1, '5': 8, '10': 'isActive'},
  ],
};

/// Descriptor for `CreateProductCategoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createProductCategoryRequestDescriptor =
    $convert.base64Decode(
        'ChxDcmVhdGVQcm9kdWN0Q2F0ZWdvcnlSZXF1ZXN0EhIKBGNvZGUYASABKAlSBGNvZGUSEgoEbm'
        'FtZRgCIAEoCVIEbmFtZRIdCgpzb3J0X29yZGVyGAMgASgFUglzb3J0T3JkZXISGwoJaXNfYWN0'
        'aXZlGAQgASgIUghpc0FjdGl2ZQ==');

@$core.Deprecated('Use createProductCategoryResponseDescriptor instead')
const CreateProductCategoryResponse$json = {
  '1': 'CreateProductCategoryResponse',
  '2': [
    {
      '1': 'product_category',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.ProductCategory',
      '10': 'productCategory'
    },
  ],
};

/// Descriptor for `CreateProductCategoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createProductCategoryResponseDescriptor =
    $convert.base64Decode(
        'Ch1DcmVhdGVQcm9kdWN0Q2F0ZWdvcnlSZXNwb25zZRJGChBwcm9kdWN0X2NhdGVnb3J5GAEgAS'
        'gLMhsubWFzdGVycy52MS5Qcm9kdWN0Q2F0ZWdvcnlSD3Byb2R1Y3RDYXRlZ29yeQ==');

@$core.Deprecated('Use updateProductCategoryRequestDescriptor instead')
const UpdateProductCategoryRequest$json = {
  '1': 'UpdateProductCategoryRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'code', '17': true},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'name', '17': true},
    {
      '1': 'sort_order',
      '3': 4,
      '4': 1,
      '5': 5,
      '9': 2,
      '10': 'sortOrder',
      '17': true
    },
    {
      '1': 'is_active',
      '3': 5,
      '4': 1,
      '5': 8,
      '9': 3,
      '10': 'isActive',
      '17': true
    },
  ],
  '8': [
    {'1': '_code'},
    {'1': '_name'},
    {'1': '_sort_order'},
    {'1': '_is_active'},
  ],
};

/// Descriptor for `UpdateProductCategoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProductCategoryRequestDescriptor = $convert.base64Decode(
    'ChxVcGRhdGVQcm9kdWN0Q2F0ZWdvcnlSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIXCgRjb2RlGA'
    'IgASgJSABSBGNvZGWIAQESFwoEbmFtZRgDIAEoCUgBUgRuYW1liAEBEiIKCnNvcnRfb3JkZXIY'
    'BCABKAVIAlIJc29ydE9yZGVyiAEBEiAKCWlzX2FjdGl2ZRgFIAEoCEgDUghpc0FjdGl2ZYgBAU'
    'IHCgVfY29kZUIHCgVfbmFtZUINCgtfc29ydF9vcmRlckIMCgpfaXNfYWN0aXZl');

@$core.Deprecated('Use updateProductCategoryResponseDescriptor instead')
const UpdateProductCategoryResponse$json = {
  '1': 'UpdateProductCategoryResponse',
  '2': [
    {
      '1': 'product_category',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.ProductCategory',
      '10': 'productCategory'
    },
  ],
};

/// Descriptor for `UpdateProductCategoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProductCategoryResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVQcm9kdWN0Q2F0ZWdvcnlSZXNwb25zZRJGChBwcm9kdWN0X2NhdGVnb3J5GAEgAS'
        'gLMhsubWFzdGVycy52MS5Qcm9kdWN0Q2F0ZWdvcnlSD3Byb2R1Y3RDYXRlZ29yeQ==');

@$core.Deprecated('Use deleteProductCategoryRequestDescriptor instead')
const DeleteProductCategoryRequest$json = {
  '1': 'DeleteProductCategoryRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteProductCategoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteProductCategoryRequestDescriptor =
    $convert.base64Decode(
        'ChxEZWxldGVQcm9kdWN0Q2F0ZWdvcnlSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use deleteProductCategoryResponseDescriptor instead')
const DeleteProductCategoryResponse$json = {
  '1': 'DeleteProductCategoryResponse',
};

/// Descriptor for `DeleteProductCategoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteProductCategoryResponseDescriptor =
    $convert.base64Decode('Ch1EZWxldGVQcm9kdWN0Q2F0ZWdvcnlSZXNwb25zZQ==');

@$core.Deprecated('Use restoreProductCategoryRequestDescriptor instead')
const RestoreProductCategoryRequest$json = {
  '1': 'RestoreProductCategoryRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `RestoreProductCategoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreProductCategoryRequestDescriptor =
    $convert.base64Decode(
        'Ch1SZXN0b3JlUHJvZHVjdENhdGVnb3J5UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use restoreProductCategoryResponseDescriptor instead')
const RestoreProductCategoryResponse$json = {
  '1': 'RestoreProductCategoryResponse',
  '2': [
    {
      '1': 'product_category',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.masters.v1.ProductCategory',
      '10': 'productCategory'
    },
  ],
};

/// Descriptor for `RestoreProductCategoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreProductCategoryResponseDescriptor =
    $convert.base64Decode(
        'Ch5SZXN0b3JlUHJvZHVjdENhdGVnb3J5UmVzcG9uc2USRgoQcHJvZHVjdF9jYXRlZ29yeRgBIA'
        'EoCzIbLm1hc3RlcnMudjEuUHJvZHVjdENhdGVnb3J5Ug9wcm9kdWN0Q2F0ZWdvcnk=');

const $core.Map<$core.String, $core.dynamic> WarehouseServiceBase$json = {
  '1': 'WarehouseService',
  '2': [
    {
      '1': 'ListWarehouses',
      '2': '.masters.v1.ListWarehousesRequest',
      '3': '.masters.v1.ListWarehousesResponse'
    },
    {
      '1': 'CreateWarehouse',
      '2': '.masters.v1.CreateWarehouseRequest',
      '3': '.masters.v1.CreateWarehouseResponse'
    },
    {
      '1': 'UpdateWarehouse',
      '2': '.masters.v1.UpdateWarehouseRequest',
      '3': '.masters.v1.UpdateWarehouseResponse'
    },
    {
      '1': 'DeleteWarehouse',
      '2': '.masters.v1.DeleteWarehouseRequest',
      '3': '.masters.v1.DeleteWarehouseResponse'
    },
    {
      '1': 'RestoreWarehouse',
      '2': '.masters.v1.RestoreWarehouseRequest',
      '3': '.masters.v1.RestoreWarehouseResponse'
    },
  ],
};

@$core.Deprecated('Use warehouseServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    WarehouseServiceBase$messageJson = {
  '.masters.v1.ListWarehousesRequest': ListWarehousesRequest$json,
  '.masters.v1.ListWarehousesResponse': ListWarehousesResponse$json,
  '.masters.v1.Warehouse': Warehouse$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.masters.v1.CreateWarehouseRequest': CreateWarehouseRequest$json,
  '.masters.v1.CreateWarehouseResponse': CreateWarehouseResponse$json,
  '.masters.v1.UpdateWarehouseRequest': UpdateWarehouseRequest$json,
  '.masters.v1.UpdateWarehouseResponse': UpdateWarehouseResponse$json,
  '.masters.v1.DeleteWarehouseRequest': DeleteWarehouseRequest$json,
  '.masters.v1.DeleteWarehouseResponse': DeleteWarehouseResponse$json,
  '.masters.v1.RestoreWarehouseRequest': RestoreWarehouseRequest$json,
  '.masters.v1.RestoreWarehouseResponse': RestoreWarehouseResponse$json,
};

/// Descriptor for `WarehouseService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List warehouseServiceDescriptor = $convert.base64Decode(
    'ChBXYXJlaG91c2VTZXJ2aWNlElcKDkxpc3RXYXJlaG91c2VzEiEubWFzdGVycy52MS5MaXN0V2'
    'FyZWhvdXNlc1JlcXVlc3QaIi5tYXN0ZXJzLnYxLkxpc3RXYXJlaG91c2VzUmVzcG9uc2USWgoP'
    'Q3JlYXRlV2FyZWhvdXNlEiIubWFzdGVycy52MS5DcmVhdGVXYXJlaG91c2VSZXF1ZXN0GiMubW'
    'FzdGVycy52MS5DcmVhdGVXYXJlaG91c2VSZXNwb25zZRJaCg9VcGRhdGVXYXJlaG91c2USIi5t'
    'YXN0ZXJzLnYxLlVwZGF0ZVdhcmVob3VzZVJlcXVlc3QaIy5tYXN0ZXJzLnYxLlVwZGF0ZVdhcm'
    'Vob3VzZVJlc3BvbnNlEloKD0RlbGV0ZVdhcmVob3VzZRIiLm1hc3RlcnMudjEuRGVsZXRlV2Fy'
    'ZWhvdXNlUmVxdWVzdBojLm1hc3RlcnMudjEuRGVsZXRlV2FyZWhvdXNlUmVzcG9uc2USXQoQUm'
    'VzdG9yZVdhcmVob3VzZRIjLm1hc3RlcnMudjEuUmVzdG9yZVdhcmVob3VzZVJlcXVlc3QaJC5t'
    'YXN0ZXJzLnYxLlJlc3RvcmVXYXJlaG91c2VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> RouteServiceBase$json = {
  '1': 'RouteService',
  '2': [
    {
      '1': 'ListRoutes',
      '2': '.masters.v1.ListRoutesRequest',
      '3': '.masters.v1.ListRoutesResponse'
    },
    {
      '1': 'CreateRoute',
      '2': '.masters.v1.CreateRouteRequest',
      '3': '.masters.v1.CreateRouteResponse'
    },
    {
      '1': 'UpdateRoute',
      '2': '.masters.v1.UpdateRouteRequest',
      '3': '.masters.v1.UpdateRouteResponse'
    },
    {
      '1': 'DeleteRoute',
      '2': '.masters.v1.DeleteRouteRequest',
      '3': '.masters.v1.DeleteRouteResponse'
    },
    {
      '1': 'RestoreRoute',
      '2': '.masters.v1.RestoreRouteRequest',
      '3': '.masters.v1.RestoreRouteResponse'
    },
  ],
};

@$core.Deprecated('Use routeServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    RouteServiceBase$messageJson = {
  '.masters.v1.ListRoutesRequest': ListRoutesRequest$json,
  '.masters.v1.ListRoutesResponse': ListRoutesResponse$json,
  '.masters.v1.Route': Route$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.masters.v1.CreateRouteRequest': CreateRouteRequest$json,
  '.masters.v1.CreateRouteResponse': CreateRouteResponse$json,
  '.masters.v1.UpdateRouteRequest': UpdateRouteRequest$json,
  '.masters.v1.UpdateRouteResponse': UpdateRouteResponse$json,
  '.masters.v1.DeleteRouteRequest': DeleteRouteRequest$json,
  '.masters.v1.DeleteRouteResponse': DeleteRouteResponse$json,
  '.masters.v1.RestoreRouteRequest': RestoreRouteRequest$json,
  '.masters.v1.RestoreRouteResponse': RestoreRouteResponse$json,
};

/// Descriptor for `RouteService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List routeServiceDescriptor = $convert.base64Decode(
    'CgxSb3V0ZVNlcnZpY2USSwoKTGlzdFJvdXRlcxIdLm1hc3RlcnMudjEuTGlzdFJvdXRlc1JlcX'
    'Vlc3QaHi5tYXN0ZXJzLnYxLkxpc3RSb3V0ZXNSZXNwb25zZRJOCgtDcmVhdGVSb3V0ZRIeLm1h'
    'c3RlcnMudjEuQ3JlYXRlUm91dGVSZXF1ZXN0Gh8ubWFzdGVycy52MS5DcmVhdGVSb3V0ZVJlc3'
    'BvbnNlEk4KC1VwZGF0ZVJvdXRlEh4ubWFzdGVycy52MS5VcGRhdGVSb3V0ZVJlcXVlc3QaHy5t'
    'YXN0ZXJzLnYxLlVwZGF0ZVJvdXRlUmVzcG9uc2USTgoLRGVsZXRlUm91dGUSHi5tYXN0ZXJzLn'
    'YxLkRlbGV0ZVJvdXRlUmVxdWVzdBofLm1hc3RlcnMudjEuRGVsZXRlUm91dGVSZXNwb25zZRJR'
    'CgxSZXN0b3JlUm91dGUSHy5tYXN0ZXJzLnYxLlJlc3RvcmVSb3V0ZVJlcXVlc3QaIC5tYXN0ZX'
    'JzLnYxLlJlc3RvcmVSb3V0ZVJlc3BvbnNl');

const $core.Map<$core.String, $core.dynamic> ProcessingSpecServiceBase$json = {
  '1': 'ProcessingSpecService',
  '2': [
    {
      '1': 'ListProcessingSpecs',
      '2': '.masters.v1.ListProcessingSpecsRequest',
      '3': '.masters.v1.ListProcessingSpecsResponse'
    },
    {
      '1': 'CreateProcessingSpec',
      '2': '.masters.v1.CreateProcessingSpecRequest',
      '3': '.masters.v1.CreateProcessingSpecResponse'
    },
    {
      '1': 'UpdateProcessingSpec',
      '2': '.masters.v1.UpdateProcessingSpecRequest',
      '3': '.masters.v1.UpdateProcessingSpecResponse'
    },
    {
      '1': 'DeleteProcessingSpec',
      '2': '.masters.v1.DeleteProcessingSpecRequest',
      '3': '.masters.v1.DeleteProcessingSpecResponse'
    },
    {
      '1': 'RestoreProcessingSpec',
      '2': '.masters.v1.RestoreProcessingSpecRequest',
      '3': '.masters.v1.RestoreProcessingSpecResponse'
    },
  ],
};

@$core.Deprecated('Use processingSpecServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ProcessingSpecServiceBase$messageJson = {
  '.masters.v1.ListProcessingSpecsRequest': ListProcessingSpecsRequest$json,
  '.masters.v1.ListProcessingSpecsResponse': ListProcessingSpecsResponse$json,
  '.masters.v1.ProcessingSpec': ProcessingSpec$json,
  '.google.protobuf.Struct': $1.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $1.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $1.Value$json,
  '.google.protobuf.ListValue': $1.ListValue$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.masters.v1.CreateProcessingSpecRequest': CreateProcessingSpecRequest$json,
  '.masters.v1.CreateProcessingSpecResponse': CreateProcessingSpecResponse$json,
  '.masters.v1.UpdateProcessingSpecRequest': UpdateProcessingSpecRequest$json,
  '.masters.v1.UpdateProcessingSpecResponse': UpdateProcessingSpecResponse$json,
  '.masters.v1.DeleteProcessingSpecRequest': DeleteProcessingSpecRequest$json,
  '.masters.v1.DeleteProcessingSpecResponse': DeleteProcessingSpecResponse$json,
  '.masters.v1.RestoreProcessingSpecRequest': RestoreProcessingSpecRequest$json,
  '.masters.v1.RestoreProcessingSpecResponse':
      RestoreProcessingSpecResponse$json,
};

/// Descriptor for `ProcessingSpecService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List processingSpecServiceDescriptor = $convert.base64Decode(
    'ChVQcm9jZXNzaW5nU3BlY1NlcnZpY2USZgoTTGlzdFByb2Nlc3NpbmdTcGVjcxImLm1hc3Rlcn'
    'MudjEuTGlzdFByb2Nlc3NpbmdTcGVjc1JlcXVlc3QaJy5tYXN0ZXJzLnYxLkxpc3RQcm9jZXNz'
    'aW5nU3BlY3NSZXNwb25zZRJpChRDcmVhdGVQcm9jZXNzaW5nU3BlYxInLm1hc3RlcnMudjEuQ3'
    'JlYXRlUHJvY2Vzc2luZ1NwZWNSZXF1ZXN0GigubWFzdGVycy52MS5DcmVhdGVQcm9jZXNzaW5n'
    'U3BlY1Jlc3BvbnNlEmkKFFVwZGF0ZVByb2Nlc3NpbmdTcGVjEicubWFzdGVycy52MS5VcGRhdG'
    'VQcm9jZXNzaW5nU3BlY1JlcXVlc3QaKC5tYXN0ZXJzLnYxLlVwZGF0ZVByb2Nlc3NpbmdTcGVj'
    'UmVzcG9uc2USaQoURGVsZXRlUHJvY2Vzc2luZ1NwZWMSJy5tYXN0ZXJzLnYxLkRlbGV0ZVByb2'
    'Nlc3NpbmdTcGVjUmVxdWVzdBooLm1hc3RlcnMudjEuRGVsZXRlUHJvY2Vzc2luZ1NwZWNSZXNw'
    'b25zZRJsChVSZXN0b3JlUHJvY2Vzc2luZ1NwZWMSKC5tYXN0ZXJzLnYxLlJlc3RvcmVQcm9jZX'
    'NzaW5nU3BlY1JlcXVlc3QaKS5tYXN0ZXJzLnYxLlJlc3RvcmVQcm9jZXNzaW5nU3BlY1Jlc3Bv'
    'bnNl');

const $core.Map<$core.String, $core.dynamic> ProductCategoryServiceBase$json = {
  '1': 'ProductCategoryService',
  '2': [
    {
      '1': 'ListProductCategories',
      '2': '.masters.v1.ListProductCategoriesRequest',
      '3': '.masters.v1.ListProductCategoriesResponse'
    },
    {
      '1': 'CreateProductCategory',
      '2': '.masters.v1.CreateProductCategoryRequest',
      '3': '.masters.v1.CreateProductCategoryResponse'
    },
    {
      '1': 'UpdateProductCategory',
      '2': '.masters.v1.UpdateProductCategoryRequest',
      '3': '.masters.v1.UpdateProductCategoryResponse'
    },
    {
      '1': 'DeleteProductCategory',
      '2': '.masters.v1.DeleteProductCategoryRequest',
      '3': '.masters.v1.DeleteProductCategoryResponse'
    },
    {
      '1': 'RestoreProductCategory',
      '2': '.masters.v1.RestoreProductCategoryRequest',
      '3': '.masters.v1.RestoreProductCategoryResponse'
    },
  ],
};

@$core.Deprecated('Use productCategoryServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ProductCategoryServiceBase$messageJson = {
  '.masters.v1.ListProductCategoriesRequest': ListProductCategoriesRequest$json,
  '.masters.v1.ListProductCategoriesResponse':
      ListProductCategoriesResponse$json,
  '.masters.v1.ProductCategory': ProductCategory$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.masters.v1.CreateProductCategoryRequest': CreateProductCategoryRequest$json,
  '.masters.v1.CreateProductCategoryResponse':
      CreateProductCategoryResponse$json,
  '.masters.v1.UpdateProductCategoryRequest': UpdateProductCategoryRequest$json,
  '.masters.v1.UpdateProductCategoryResponse':
      UpdateProductCategoryResponse$json,
  '.masters.v1.DeleteProductCategoryRequest': DeleteProductCategoryRequest$json,
  '.masters.v1.DeleteProductCategoryResponse':
      DeleteProductCategoryResponse$json,
  '.masters.v1.RestoreProductCategoryRequest':
      RestoreProductCategoryRequest$json,
  '.masters.v1.RestoreProductCategoryResponse':
      RestoreProductCategoryResponse$json,
};

/// Descriptor for `ProductCategoryService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List productCategoryServiceDescriptor = $convert.base64Decode(
    'ChZQcm9kdWN0Q2F0ZWdvcnlTZXJ2aWNlEmwKFUxpc3RQcm9kdWN0Q2F0ZWdvcmllcxIoLm1hc3'
    'RlcnMudjEuTGlzdFByb2R1Y3RDYXRlZ29yaWVzUmVxdWVzdBopLm1hc3RlcnMudjEuTGlzdFBy'
    'b2R1Y3RDYXRlZ29yaWVzUmVzcG9uc2USbAoVQ3JlYXRlUHJvZHVjdENhdGVnb3J5EigubWFzdG'
    'Vycy52MS5DcmVhdGVQcm9kdWN0Q2F0ZWdvcnlSZXF1ZXN0GikubWFzdGVycy52MS5DcmVhdGVQ'
    'cm9kdWN0Q2F0ZWdvcnlSZXNwb25zZRJsChVVcGRhdGVQcm9kdWN0Q2F0ZWdvcnkSKC5tYXN0ZX'
    'JzLnYxLlVwZGF0ZVByb2R1Y3RDYXRlZ29yeVJlcXVlc3QaKS5tYXN0ZXJzLnYxLlVwZGF0ZVBy'
    'b2R1Y3RDYXRlZ29yeVJlc3BvbnNlEmwKFURlbGV0ZVByb2R1Y3RDYXRlZ29yeRIoLm1hc3Rlcn'
    'MudjEuRGVsZXRlUHJvZHVjdENhdGVnb3J5UmVxdWVzdBopLm1hc3RlcnMudjEuRGVsZXRlUHJv'
    'ZHVjdENhdGVnb3J5UmVzcG9uc2USbwoWUmVzdG9yZVByb2R1Y3RDYXRlZ29yeRIpLm1hc3Rlcn'
    'MudjEuUmVzdG9yZVByb2R1Y3RDYXRlZ29yeVJlcXVlc3QaKi5tYXN0ZXJzLnYxLlJlc3RvcmVQ'
    'cm9kdWN0Q2F0ZWdvcnlSZXNwb25zZQ==');
