// This is a generated file - do not edit.
//
// Generated from salesorder/v1/returns.proto.

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

@$core.Deprecated('Use returnItemInputDescriptor instead')
const ReturnItemInput$json = {
  '1': 'ReturnItemInput',
  '2': [
    {'1': 'source_type', '3': 1, '4': 1, '5': 9, '10': 'sourceType'},
    {
      '1': 'sales_order_item_id',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'salesOrderItemId'
    },
    {
      '1': 'customer_product_id',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'customerProductId'
    },
    {'1': 'quantity', '3': 4, '4': 1, '5': 9, '10': 'quantity'},
    {'1': 'reason', '3': 5, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'photo_file_ids', '3': 6, '4': 3, '5': 9, '10': 'photoFileIds'},
  ],
};

/// Descriptor for `ReturnItemInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List returnItemInputDescriptor = $convert.base64Decode(
    'Cg9SZXR1cm5JdGVtSW5wdXQSHwoLc291cmNlX3R5cGUYASABKAlSCnNvdXJjZVR5cGUSLQoTc2'
    'FsZXNfb3JkZXJfaXRlbV9pZBgCIAEoCVIQc2FsZXNPcmRlckl0ZW1JZBIuChNjdXN0b21lcl9w'
    'cm9kdWN0X2lkGAMgASgJUhFjdXN0b21lclByb2R1Y3RJZBIaCghxdWFudGl0eRgEIAEoCVIIcX'
    'VhbnRpdHkSFgoGcmVhc29uGAUgASgJUgZyZWFzb24SJAoOcGhvdG9fZmlsZV9pZHMYBiADKAlS'
    'DHBob3RvRmlsZUlkcw==');

@$core.Deprecated('Use createReturnRequestRequestDescriptor instead')
const CreateReturnRequestRequest$json = {
  '1': 'CreateReturnRequestRequest',
  '2': [
    {
      '1': 'items',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.ReturnItemInput',
      '10': 'items'
    },
    {'1': 'remark', '3': 2, '4': 1, '5': 9, '10': 'remark'},
  ],
};

/// Descriptor for `CreateReturnRequestRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createReturnRequestRequestDescriptor =
    $convert.base64Decode(
        'ChpDcmVhdGVSZXR1cm5SZXF1ZXN0UmVxdWVzdBI0CgVpdGVtcxgBIAMoCzIeLnNhbGVzb3JkZX'
        'IudjEuUmV0dXJuSXRlbUlucHV0UgVpdGVtcxIWCgZyZW1hcmsYAiABKAlSBnJlbWFyaw==');

@$core.Deprecated('Use createReturnRequestResponseDescriptor instead')
const CreateReturnRequestResponse$json = {
  '1': 'CreateReturnRequestResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'status', '3': 2, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `CreateReturnRequestResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createReturnRequestResponseDescriptor =
    $convert.base64Decode(
        'ChtDcmVhdGVSZXR1cm5SZXF1ZXN0UmVzcG9uc2USDgoCaWQYASABKAlSAmlkEhYKBnN0YXR1cx'
        'gCIAEoCVIGc3RhdHVz');

@$core.Deprecated('Use listReturnRequestsRequestDescriptor instead')
const ListReturnRequestsRequest$json = {
  '1': 'ListReturnRequestsRequest',
  '2': [
    {'1': 'status', '3': 1, '4': 1, '5': 9, '10': 'status'},
    {'1': 'page', '3': 2, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListReturnRequestsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listReturnRequestsRequestDescriptor =
    $convert.base64Decode(
        'ChlMaXN0UmV0dXJuUmVxdWVzdHNSZXF1ZXN0EhYKBnN0YXR1cxgBIAEoCVIGc3RhdHVzEhIKBH'
        'BhZ2UYAiABKAVSBHBhZ2USGwoJcGFnZV9zaXplGAMgASgFUghwYWdlU2l6ZQ==');

@$core.Deprecated('Use returnRequestEntryDescriptor instead')
const ReturnRequestEntry$json = {
  '1': 'ReturnRequestEntry',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'customer_id', '3': 2, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'status', '3': 3, '4': 1, '5': 9, '10': 'status'},
    {'1': 'remark', '3': 4, '4': 1, '5': 9, '10': 'remark'},
    {'1': 'created_at', '3': 5, '4': 1, '5': 9, '10': 'createdAt'},
  ],
};

/// Descriptor for `ReturnRequestEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List returnRequestEntryDescriptor = $convert.base64Decode(
    'ChJSZXR1cm5SZXF1ZXN0RW50cnkSDgoCaWQYASABKAlSAmlkEh8KC2N1c3RvbWVyX2lkGAIgAS'
    'gJUgpjdXN0b21lcklkEhYKBnN0YXR1cxgDIAEoCVIGc3RhdHVzEhYKBnJlbWFyaxgEIAEoCVIG'
    'cmVtYXJrEh0KCmNyZWF0ZWRfYXQYBSABKAlSCWNyZWF0ZWRBdA==');

@$core.Deprecated('Use listReturnRequestsResponseDescriptor instead')
const ListReturnRequestsResponse$json = {
  '1': 'ListReturnRequestsResponse',
  '2': [
    {
      '1': 'entries',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.ReturnRequestEntry',
      '10': 'entries'
    },
    {'1': 'page', '3': 2, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'total', '3': 4, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListReturnRequestsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listReturnRequestsResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0UmV0dXJuUmVxdWVzdHNSZXNwb25zZRI7CgdlbnRyaWVzGAEgAygLMiEuc2FsZXNvcm'
        'Rlci52MS5SZXR1cm5SZXF1ZXN0RW50cnlSB2VudHJpZXMSEgoEcGFnZRgCIAEoBVIEcGFnZRIb'
        'CglwYWdlX3NpemUYAyABKAVSCHBhZ2VTaXplEhQKBXRvdGFsGAQgASgFUgV0b3RhbA==');

@$core.Deprecated('Use returnRequestItemViewDescriptor instead')
const ReturnRequestItemView$json = {
  '1': 'ReturnRequestItemView',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'source_type', '3': 2, '4': 1, '5': 9, '10': 'sourceType'},
    {'1': 'product_name', '3': 3, '4': 1, '5': 9, '10': 'productName'},
    {'1': 'spec', '3': 4, '4': 1, '5': 9, '10': 'spec'},
    {'1': 'unit', '3': 5, '4': 1, '5': 9, '10': 'unit'},
    {'1': 'quantity', '3': 6, '4': 1, '5': 9, '10': 'quantity'},
    {'1': 'reason', '3': 7, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'photo_urls', '3': 8, '4': 3, '5': 9, '10': 'photoUrls'},
  ],
};

/// Descriptor for `ReturnRequestItemView`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List returnRequestItemViewDescriptor = $convert.base64Decode(
    'ChVSZXR1cm5SZXF1ZXN0SXRlbVZpZXcSDgoCaWQYASABKAlSAmlkEh8KC3NvdXJjZV90eXBlGA'
    'IgASgJUgpzb3VyY2VUeXBlEiEKDHByb2R1Y3RfbmFtZRgDIAEoCVILcHJvZHVjdE5hbWUSEgoE'
    'c3BlYxgEIAEoCVIEc3BlYxISCgR1bml0GAUgASgJUgR1bml0EhoKCHF1YW50aXR5GAYgASgJUg'
    'hxdWFudGl0eRIWCgZyZWFzb24YByABKAlSBnJlYXNvbhIdCgpwaG90b191cmxzGAggAygJUglw'
    'aG90b1VybHM=');

@$core.Deprecated('Use getReturnRequestRequestDescriptor instead')
const GetReturnRequestRequest$json = {
  '1': 'GetReturnRequestRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetReturnRequestRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getReturnRequestRequestDescriptor = $convert
    .base64Decode('ChdHZXRSZXR1cm5SZXF1ZXN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use getReturnRequestResponseDescriptor instead')
const GetReturnRequestResponse$json = {
  '1': 'GetReturnRequestResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'customer_id', '3': 2, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'status', '3': 3, '4': 1, '5': 9, '10': 'status'},
    {'1': 'remark', '3': 4, '4': 1, '5': 9, '10': 'remark'},
    {'1': 'reject_reason', '3': 5, '4': 1, '5': 9, '10': 'rejectReason'},
    {'1': 'created_at', '3': 6, '4': 1, '5': 9, '10': 'createdAt'},
    {
      '1': 'items',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.ReturnRequestItemView',
      '10': 'items'
    },
  ],
};

/// Descriptor for `GetReturnRequestResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getReturnRequestResponseDescriptor = $convert.base64Decode(
    'ChhHZXRSZXR1cm5SZXF1ZXN0UmVzcG9uc2USDgoCaWQYASABKAlSAmlkEh8KC2N1c3RvbWVyX2'
    'lkGAIgASgJUgpjdXN0b21lcklkEhYKBnN0YXR1cxgDIAEoCVIGc3RhdHVzEhYKBnJlbWFyaxgE'
    'IAEoCVIGcmVtYXJrEiMKDXJlamVjdF9yZWFzb24YBSABKAlSDHJlamVjdFJlYXNvbhIdCgpjcm'
    'VhdGVkX2F0GAYgASgJUgljcmVhdGVkQXQSOgoFaXRlbXMYByADKAsyJC5zYWxlc29yZGVyLnYx'
    'LlJldHVyblJlcXVlc3RJdGVtVmlld1IFaXRlbXM=');

const $core.Map<$core.String, $core.dynamic> ReturnServiceBase$json = {
  '1': 'ReturnService',
  '2': [
    {
      '1': 'CreateReturnRequest',
      '2': '.salesorder.v1.CreateReturnRequestRequest',
      '3': '.salesorder.v1.CreateReturnRequestResponse'
    },
    {
      '1': 'ListReturnRequests',
      '2': '.salesorder.v1.ListReturnRequestsRequest',
      '3': '.salesorder.v1.ListReturnRequestsResponse'
    },
    {
      '1': 'GetReturnRequest',
      '2': '.salesorder.v1.GetReturnRequestRequest',
      '3': '.salesorder.v1.GetReturnRequestResponse'
    },
  ],
};

@$core.Deprecated('Use returnServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ReturnServiceBase$messageJson = {
  '.salesorder.v1.CreateReturnRequestRequest': CreateReturnRequestRequest$json,
  '.salesorder.v1.ReturnItemInput': ReturnItemInput$json,
  '.salesorder.v1.CreateReturnRequestResponse':
      CreateReturnRequestResponse$json,
  '.salesorder.v1.ListReturnRequestsRequest': ListReturnRequestsRequest$json,
  '.salesorder.v1.ListReturnRequestsResponse': ListReturnRequestsResponse$json,
  '.salesorder.v1.ReturnRequestEntry': ReturnRequestEntry$json,
  '.salesorder.v1.GetReturnRequestRequest': GetReturnRequestRequest$json,
  '.salesorder.v1.GetReturnRequestResponse': GetReturnRequestResponse$json,
  '.salesorder.v1.ReturnRequestItemView': ReturnRequestItemView$json,
};

/// Descriptor for `ReturnService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List returnServiceDescriptor = $convert.base64Decode(
    'Cg1SZXR1cm5TZXJ2aWNlEmwKE0NyZWF0ZVJldHVyblJlcXVlc3QSKS5zYWxlc29yZGVyLnYxLk'
    'NyZWF0ZVJldHVyblJlcXVlc3RSZXF1ZXN0Giouc2FsZXNvcmRlci52MS5DcmVhdGVSZXR1cm5S'
    'ZXF1ZXN0UmVzcG9uc2USaQoSTGlzdFJldHVyblJlcXVlc3RzEiguc2FsZXNvcmRlci52MS5MaX'
    'N0UmV0dXJuUmVxdWVzdHNSZXF1ZXN0Gikuc2FsZXNvcmRlci52MS5MaXN0UmV0dXJuUmVxdWVz'
    'dHNSZXNwb25zZRJjChBHZXRSZXR1cm5SZXF1ZXN0EiYuc2FsZXNvcmRlci52MS5HZXRSZXR1cm'
    '5SZXF1ZXN0UmVxdWVzdBonLnNhbGVzb3JkZXIudjEuR2V0UmV0dXJuUmVxdWVzdFJlc3BvbnNl');
