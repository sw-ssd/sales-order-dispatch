// This is a generated file - do not edit.
//
// Generated from metadict/v1/metadict.proto.

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

@$core.Deprecated('Use metadictDescriptor instead')
const Metadict$json = {
  '1': 'Metadict',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'code', '3': 3, '4': 1, '5': 9, '10': 'code'},
    {'1': 'display_name', '3': 4, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'department_id', '3': 5, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'sort_order', '3': 6, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 7, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'created_at', '3': 8, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 9, '4': 1, '5': 9, '10': 'updatedAt'},
  ],
};

/// Descriptor for `Metadict`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List metadictDescriptor = $convert.base64Decode(
    'CghNZXRhZGljdBIOCgJpZBgBIAEoCVICaWQSEgoEdHlwZRgCIAEoCVIEdHlwZRISCgRjb2RlGA'
    'MgASgJUgRjb2RlEiEKDGRpc3BsYXlfbmFtZRgEIAEoCVILZGlzcGxheU5hbWUSIwoNZGVwYXJ0'
    'bWVudF9pZBgFIAEoCVIMZGVwYXJ0bWVudElkEh0KCnNvcnRfb3JkZXIYBiABKAVSCXNvcnRPcm'
    'RlchIbCglpc19hY3RpdmUYByABKAhSCGlzQWN0aXZlEh0KCmNyZWF0ZWRfYXQYCCABKAlSCWNy'
    'ZWF0ZWRBdBIdCgp1cGRhdGVkX2F0GAkgASgJUgl1cGRhdGVkQXQ=');

@$core.Deprecated('Use listMetadictsRequestDescriptor instead')
const ListMetadictsRequest$json = {
  '1': 'ListMetadictsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'type', '3': 3, '4': 1, '5': 9, '10': 'type'},
    {'1': 'include_deleted', '3': 5, '4': 1, '5': 8, '10': 'includeDeleted'},
    {'1': 'department_id', '3': 6, '4': 1, '5': 9, '10': 'departmentId'},
  ],
  '9': [
    {'1': 4, '2': 5},
  ],
};

/// Descriptor for `ListMetadictsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMetadictsRequestDescriptor = $convert.base64Decode(
    'ChRMaXN0TWV0YWRpY3RzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZR'
    'gCIAEoBVIIcGFnZVNpemUSEgoEdHlwZRgDIAEoCVIEdHlwZRInCg9pbmNsdWRlX2RlbGV0ZWQY'
    'BSABKAhSDmluY2x1ZGVEZWxldGVkEiMKDWRlcGFydG1lbnRfaWQYBiABKAlSDGRlcGFydG1lbn'
    'RJZEoECAQQBQ==');

@$core.Deprecated('Use listMetadictsResponseDescriptor instead')
const ListMetadictsResponse$json = {
  '1': 'ListMetadictsResponse',
  '2': [
    {
      '1': 'items',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.metadict.v1.Metadict',
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

/// Descriptor for `ListMetadictsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMetadictsResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0TWV0YWRpY3RzUmVzcG9uc2USKwoFaXRlbXMYASADKAsyFS5tZXRhZGljdC52MS5NZX'
    'RhZGljdFIFaXRlbXMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIudjEuUGFnaW5h'
    'dGlvblIKcGFnaW5hdGlvbg==');

@$core.Deprecated('Use getMetadictRequestDescriptor instead')
const GetMetadictRequest$json = {
  '1': 'GetMetadictRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetMetadictRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMetadictRequestDescriptor =
    $convert.base64Decode('ChJHZXRNZXRhZGljdFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getMetadictResponseDescriptor instead')
const GetMetadictResponse$json = {
  '1': 'GetMetadictResponse',
  '2': [
    {
      '1': 'metadict',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.metadict.v1.Metadict',
      '10': 'metadict'
    },
  ],
};

/// Descriptor for `GetMetadictResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMetadictResponseDescriptor = $convert.base64Decode(
    'ChNHZXRNZXRhZGljdFJlc3BvbnNlEjEKCG1ldGFkaWN0GAEgASgLMhUubWV0YWRpY3QudjEuTW'
    'V0YWRpY3RSCG1ldGFkaWN0');

@$core.Deprecated('Use createMetadictRequestDescriptor instead')
const CreateMetadictRequest$json = {
  '1': 'CreateMetadictRequest',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'sort_order', '3': 4, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 5, '4': 1, '5': 8, '10': 'isActive'},
  ],
};

/// Descriptor for `CreateMetadictRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMetadictRequestDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVNZXRhZGljdFJlcXVlc3QSEgoEdHlwZRgBIAEoCVIEdHlwZRISCgRjb2RlGAIgAS'
    'gJUgRjb2RlEiEKDGRpc3BsYXlfbmFtZRgDIAEoCVILZGlzcGxheU5hbWUSHQoKc29ydF9vcmRl'
    'chgEIAEoBVIJc29ydE9yZGVyEhsKCWlzX2FjdGl2ZRgFIAEoCFIIaXNBY3RpdmU=');

@$core.Deprecated('Use createMetadictResponseDescriptor instead')
const CreateMetadictResponse$json = {
  '1': 'CreateMetadictResponse',
  '2': [
    {
      '1': 'metadict',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.metadict.v1.Metadict',
      '10': 'metadict'
    },
  ],
};

/// Descriptor for `CreateMetadictResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMetadictResponseDescriptor =
    $convert.base64Decode(
        'ChZDcmVhdGVNZXRhZGljdFJlc3BvbnNlEjEKCG1ldGFkaWN0GAEgASgLMhUubWV0YWRpY3Qudj'
        'EuTWV0YWRpY3RSCG1ldGFkaWN0');

@$core.Deprecated('Use updateMetadictRequestDescriptor instead')
const UpdateMetadictRequest$json = {
  '1': 'UpdateMetadictRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'display_name',
      '3': 2,
      '4': 1,
      '5': 9,
      '9': 0,
      '10': 'displayName',
      '17': true
    },
    {
      '1': 'sort_order',
      '3': 3,
      '4': 1,
      '5': 5,
      '9': 1,
      '10': 'sortOrder',
      '17': true
    },
    {
      '1': 'is_active',
      '3': 4,
      '4': 1,
      '5': 8,
      '9': 2,
      '10': 'isActive',
      '17': true
    },
  ],
  '8': [
    {'1': '_display_name'},
    {'1': '_sort_order'},
    {'1': '_is_active'},
  ],
};

/// Descriptor for `UpdateMetadictRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateMetadictRequestDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVNZXRhZGljdFJlcXVlc3QSDgoCaWQYASABKAlSAmlkEiYKDGRpc3BsYXlfbmFtZR'
    'gCIAEoCUgAUgtkaXNwbGF5TmFtZYgBARIiCgpzb3J0X29yZGVyGAMgASgFSAFSCXNvcnRPcmRl'
    'cogBARIgCglpc19hY3RpdmUYBCABKAhIAlIIaXNBY3RpdmWIAQFCDwoNX2Rpc3BsYXlfbmFtZU'
    'INCgtfc29ydF9vcmRlckIMCgpfaXNfYWN0aXZl');

@$core.Deprecated('Use updateMetadictResponseDescriptor instead')
const UpdateMetadictResponse$json = {
  '1': 'UpdateMetadictResponse',
  '2': [
    {
      '1': 'metadict',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.metadict.v1.Metadict',
      '10': 'metadict'
    },
  ],
};

/// Descriptor for `UpdateMetadictResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateMetadictResponseDescriptor =
    $convert.base64Decode(
        'ChZVcGRhdGVNZXRhZGljdFJlc3BvbnNlEjEKCG1ldGFkaWN0GAEgASgLMhUubWV0YWRpY3Qudj'
        'EuTWV0YWRpY3RSCG1ldGFkaWN0');

@$core.Deprecated('Use deleteMetadictRequestDescriptor instead')
const DeleteMetadictRequest$json = {
  '1': 'DeleteMetadictRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteMetadictRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteMetadictRequestDescriptor = $convert
    .base64Decode('ChVEZWxldGVNZXRhZGljdFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteMetadictResponseDescriptor instead')
const DeleteMetadictResponse$json = {
  '1': 'DeleteMetadictResponse',
};

/// Descriptor for `DeleteMetadictResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteMetadictResponseDescriptor =
    $convert.base64Decode('ChZEZWxldGVNZXRhZGljdFJlc3BvbnNl');

@$core.Deprecated('Use listOptionsRequestDescriptor instead')
const ListOptionsRequest$json = {
  '1': 'ListOptionsRequest',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    {'1': 'keyword', '3': 2, '4': 1, '5': 9, '10': 'keyword'},
  ],
};

/// Descriptor for `ListOptionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOptionsRequestDescriptor = $convert.base64Decode(
    'ChJMaXN0T3B0aW9uc1JlcXVlc3QSEgoEdHlwZRgBIAEoCVIEdHlwZRIYCgdrZXl3b3JkGAIgAS'
    'gJUgdrZXl3b3Jk');

@$core.Deprecated('Use listOptionsResponseDescriptor instead')
const ListOptionsResponse$json = {
  '1': 'ListOptionsResponse',
  '2': [
    {
      '1': 'options',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.metadict.v1.Option',
      '10': 'options'
    },
  ],
};

/// Descriptor for `ListOptionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOptionsResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0T3B0aW9uc1Jlc3BvbnNlEi0KB29wdGlvbnMYASADKAsyEy5tZXRhZGljdC52MS5PcH'
    'Rpb25SB29wdGlvbnM=');

@$core.Deprecated('Use optionDescriptor instead')
const Option$json = {
  '1': 'Option',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
  ],
};

/// Descriptor for `Option`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List optionDescriptor = $convert.base64Decode(
    'CgZPcHRpb24SEgoEY29kZRgBIAEoCVIEY29kZRIhCgxkaXNwbGF5X25hbWUYAiABKAlSC2Rpc3'
    'BsYXlOYW1l');

const $core.Map<$core.String, $core.dynamic> MetadictServiceBase$json = {
  '1': 'MetadictService',
  '2': [
    {
      '1': 'ListMetadicts',
      '2': '.metadict.v1.ListMetadictsRequest',
      '3': '.metadict.v1.ListMetadictsResponse'
    },
    {
      '1': 'GetMetadict',
      '2': '.metadict.v1.GetMetadictRequest',
      '3': '.metadict.v1.GetMetadictResponse'
    },
    {
      '1': 'CreateMetadict',
      '2': '.metadict.v1.CreateMetadictRequest',
      '3': '.metadict.v1.CreateMetadictResponse'
    },
    {
      '1': 'UpdateMetadict',
      '2': '.metadict.v1.UpdateMetadictRequest',
      '3': '.metadict.v1.UpdateMetadictResponse'
    },
    {
      '1': 'DeleteMetadict',
      '2': '.metadict.v1.DeleteMetadictRequest',
      '3': '.metadict.v1.DeleteMetadictResponse'
    },
    {
      '1': 'ListOptions',
      '2': '.metadict.v1.ListOptionsRequest',
      '3': '.metadict.v1.ListOptionsResponse'
    },
  ],
};

@$core.Deprecated('Use metadictServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    MetadictServiceBase$messageJson = {
  '.metadict.v1.ListMetadictsRequest': ListMetadictsRequest$json,
  '.metadict.v1.ListMetadictsResponse': ListMetadictsResponse$json,
  '.metadict.v1.Metadict': Metadict$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.metadict.v1.GetMetadictRequest': GetMetadictRequest$json,
  '.metadict.v1.GetMetadictResponse': GetMetadictResponse$json,
  '.metadict.v1.CreateMetadictRequest': CreateMetadictRequest$json,
  '.metadict.v1.CreateMetadictResponse': CreateMetadictResponse$json,
  '.metadict.v1.UpdateMetadictRequest': UpdateMetadictRequest$json,
  '.metadict.v1.UpdateMetadictResponse': UpdateMetadictResponse$json,
  '.metadict.v1.DeleteMetadictRequest': DeleteMetadictRequest$json,
  '.metadict.v1.DeleteMetadictResponse': DeleteMetadictResponse$json,
  '.metadict.v1.ListOptionsRequest': ListOptionsRequest$json,
  '.metadict.v1.ListOptionsResponse': ListOptionsResponse$json,
  '.metadict.v1.Option': Option$json,
};

/// Descriptor for `MetadictService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List metadictServiceDescriptor = $convert.base64Decode(
    'Cg9NZXRhZGljdFNlcnZpY2USVgoNTGlzdE1ldGFkaWN0cxIhLm1ldGFkaWN0LnYxLkxpc3RNZX'
    'RhZGljdHNSZXF1ZXN0GiIubWV0YWRpY3QudjEuTGlzdE1ldGFkaWN0c1Jlc3BvbnNlElAKC0dl'
    'dE1ldGFkaWN0Eh8ubWV0YWRpY3QudjEuR2V0TWV0YWRpY3RSZXF1ZXN0GiAubWV0YWRpY3Qudj'
    'EuR2V0TWV0YWRpY3RSZXNwb25zZRJZCg5DcmVhdGVNZXRhZGljdBIiLm1ldGFkaWN0LnYxLkNy'
    'ZWF0ZU1ldGFkaWN0UmVxdWVzdBojLm1ldGFkaWN0LnYxLkNyZWF0ZU1ldGFkaWN0UmVzcG9uc2'
    'USWQoOVXBkYXRlTWV0YWRpY3QSIi5tZXRhZGljdC52MS5VcGRhdGVNZXRhZGljdFJlcXVlc3Qa'
    'Iy5tZXRhZGljdC52MS5VcGRhdGVNZXRhZGljdFJlc3BvbnNlElkKDkRlbGV0ZU1ldGFkaWN0Ei'
    'IubWV0YWRpY3QudjEuRGVsZXRlTWV0YWRpY3RSZXF1ZXN0GiMubWV0YWRpY3QudjEuRGVsZXRl'
    'TWV0YWRpY3RSZXNwb25zZRJQCgtMaXN0T3B0aW9ucxIfLm1ldGFkaWN0LnYxLkxpc3RPcHRpb2'
    '5zUmVxdWVzdBogLm1ldGFkaWN0LnYxLkxpc3RPcHRpb25zUmVzcG9uc2U=');
