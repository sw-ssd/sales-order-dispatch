// This is a generated file - do not edit.
//
// Generated from salesorder/v1/dispatch.proto.

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

@$core.Deprecated('Use assignRouteRequestDescriptor instead')
const AssignRouteRequest$json = {
  '1': 'AssignRouteRequest',
  '2': [
    {'1': 'sales_order_id', '3': 1, '4': 1, '5': 9, '10': 'salesOrderId'},
    {'1': 'route_id', '3': 2, '4': 1, '5': 9, '10': 'routeId'},
    {
      '1': 'delivery_sequence',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'deliverySequence'
    },
    {'1': 'version', '3': 4, '4': 1, '5': 9, '10': 'version'},
    {
      '1': 'expected_delivery_date',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'expectedDeliveryDate'
    },
  ],
};

/// Descriptor for `AssignRouteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignRouteRequestDescriptor = $convert.base64Decode(
    'ChJBc3NpZ25Sb3V0ZVJlcXVlc3QSJAoOc2FsZXNfb3JkZXJfaWQYASABKAlSDHNhbGVzT3JkZX'
    'JJZBIZCghyb3V0ZV9pZBgCIAEoCVIHcm91dGVJZBIrChFkZWxpdmVyeV9zZXF1ZW5jZRgDIAEo'
    'CVIQZGVsaXZlcnlTZXF1ZW5jZRIYCgd2ZXJzaW9uGAQgASgJUgd2ZXJzaW9uEjQKFmV4cGVjdG'
    'VkX2RlbGl2ZXJ5X2RhdGUYBSABKAlSFGV4cGVjdGVkRGVsaXZlcnlEYXRl');

@$core.Deprecated('Use assignRouteResponseDescriptor instead')
const AssignRouteResponse$json = {
  '1': 'AssignRouteResponse',
  '2': [
    {'1': 'sales_order_id', '3': 1, '4': 1, '5': 9, '10': 'salesOrderId'},
    {'1': 'route_id', '3': 2, '4': 1, '5': 9, '10': 'routeId'},
    {
      '1': 'delivery_sequence',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'deliverySequence'
    },
    {'1': 'version', '3': 4, '4': 1, '5': 9, '10': 'version'},
  ],
};

/// Descriptor for `AssignRouteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignRouteResponseDescriptor = $convert.base64Decode(
    'ChNBc3NpZ25Sb3V0ZVJlc3BvbnNlEiQKDnNhbGVzX29yZGVyX2lkGAEgASgJUgxzYWxlc09yZG'
    'VySWQSGQoIcm91dGVfaWQYAiABKAlSB3JvdXRlSWQSKwoRZGVsaXZlcnlfc2VxdWVuY2UYAyAB'
    'KAlSEGRlbGl2ZXJ5U2VxdWVuY2USGAoHdmVyc2lvbhgEIAEoCVIHdmVyc2lvbg==');

@$core.Deprecated('Use confirmDispatchRequestDescriptor instead')
const ConfirmDispatchRequest$json = {
  '1': 'ConfirmDispatchRequest',
  '2': [
    {'1': 'route_id', '3': 1, '4': 1, '5': 9, '10': 'routeId'},
    {
      '1': 'expected_delivery_date',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'expectedDeliveryDate'
    },
  ],
};

/// Descriptor for `ConfirmDispatchRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List confirmDispatchRequestDescriptor =
    $convert.base64Decode(
        'ChZDb25maXJtRGlzcGF0Y2hSZXF1ZXN0EhkKCHJvdXRlX2lkGAEgASgJUgdyb3V0ZUlkEjQKFm'
        'V4cGVjdGVkX2RlbGl2ZXJ5X2RhdGUYAiABKAlSFGV4cGVjdGVkRGVsaXZlcnlEYXRl');

@$core.Deprecated('Use dispatchItemResultDescriptor instead')
const DispatchItemResult$json = {
  '1': 'DispatchItemResult',
  '2': [
    {'1': 'sales_order_id', '3': 1, '4': 1, '5': 9, '10': 'salesOrderId'},
    {'1': 'success', '3': 2, '4': 1, '5': 8, '10': 'success'},
    {'1': 'fail_reason', '3': 3, '4': 1, '5': 9, '10': 'failReason'},
  ],
};

/// Descriptor for `DispatchItemResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List dispatchItemResultDescriptor = $convert.base64Decode(
    'ChJEaXNwYXRjaEl0ZW1SZXN1bHQSJAoOc2FsZXNfb3JkZXJfaWQYASABKAlSDHNhbGVzT3JkZX'
    'JJZBIYCgdzdWNjZXNzGAIgASgIUgdzdWNjZXNzEh8KC2ZhaWxfcmVhc29uGAMgASgJUgpmYWls'
    'UmVhc29u');

@$core.Deprecated('Use confirmDispatchResponseDescriptor instead')
const ConfirmDispatchResponse$json = {
  '1': 'ConfirmDispatchResponse',
  '2': [
    {
      '1': 'items',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.DispatchItemResult',
      '10': 'items'
    },
    {'1': 'success_count', '3': 2, '4': 1, '5': 5, '10': 'successCount'},
  ],
};

/// Descriptor for `ConfirmDispatchResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List confirmDispatchResponseDescriptor = $convert.base64Decode(
    'ChdDb25maXJtRGlzcGF0Y2hSZXNwb25zZRI3CgVpdGVtcxgBIAMoCzIhLnNhbGVzb3JkZXIudj'
    'EuRGlzcGF0Y2hJdGVtUmVzdWx0UgVpdGVtcxIjCg1zdWNjZXNzX2NvdW50GAIgASgFUgxzdWNj'
    'ZXNzQ291bnQ=');

@$core.Deprecated('Use cancelDispatchRequestDescriptor instead')
const CancelDispatchRequest$json = {
  '1': 'CancelDispatchRequest',
  '2': [
    {'1': 'sales_order_id', '3': 1, '4': 1, '5': 9, '10': 'salesOrderId'},
    {'1': 'reason', '3': 2, '4': 1, '5': 9, '10': 'reason'},
    {
      '1': 'acknowledge_reprint',
      '3': 3,
      '4': 1,
      '5': 8,
      '10': 'acknowledgeReprint'
    },
  ],
};

/// Descriptor for `CancelDispatchRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelDispatchRequestDescriptor = $convert.base64Decode(
    'ChVDYW5jZWxEaXNwYXRjaFJlcXVlc3QSJAoOc2FsZXNfb3JkZXJfaWQYASABKAlSDHNhbGVzT3'
    'JkZXJJZBIWCgZyZWFzb24YAiABKAlSBnJlYXNvbhIvChNhY2tub3dsZWRnZV9yZXByaW50GAMg'
    'ASgIUhJhY2tub3dsZWRnZVJlcHJpbnQ=');

@$core.Deprecated('Use cancelDispatchResponseDescriptor instead')
const CancelDispatchResponse$json = {
  '1': 'CancelDispatchResponse',
  '2': [
    {'1': 'sales_order_id', '3': 1, '4': 1, '5': 9, '10': 'salesOrderId'},
    {'1': 'status', '3': 2, '4': 1, '5': 9, '10': 'status'},
    {'1': 'route_id', '3': 3, '4': 1, '5': 9, '10': 'routeId'},
    {
      '1': 'delivery_sequence',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'deliverySequence'
    },
    {'1': 'reprint_warning', '3': 5, '4': 1, '5': 8, '10': 'reprintWarning'},
  ],
};

/// Descriptor for `CancelDispatchResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelDispatchResponseDescriptor = $convert.base64Decode(
    'ChZDYW5jZWxEaXNwYXRjaFJlc3BvbnNlEiQKDnNhbGVzX29yZGVyX2lkGAEgASgJUgxzYWxlc0'
    '9yZGVySWQSFgoGc3RhdHVzGAIgASgJUgZzdGF0dXMSGQoIcm91dGVfaWQYAyABKAlSB3JvdXRl'
    'SWQSKwoRZGVsaXZlcnlfc2VxdWVuY2UYBCABKAlSEGRlbGl2ZXJ5U2VxdWVuY2USJwoPcmVwcm'
    'ludF93YXJuaW5nGAUgASgIUg5yZXByaW50V2FybmluZw==');

const $core.Map<$core.String, $core.dynamic> DispatchServiceBase$json = {
  '1': 'DispatchService',
  '2': [
    {
      '1': 'AssignRoute',
      '2': '.salesorder.v1.AssignRouteRequest',
      '3': '.salesorder.v1.AssignRouteResponse'
    },
    {
      '1': 'ConfirmDispatch',
      '2': '.salesorder.v1.ConfirmDispatchRequest',
      '3': '.salesorder.v1.ConfirmDispatchResponse'
    },
    {
      '1': 'CancelDispatch',
      '2': '.salesorder.v1.CancelDispatchRequest',
      '3': '.salesorder.v1.CancelDispatchResponse'
    },
  ],
};

@$core.Deprecated('Use dispatchServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DispatchServiceBase$messageJson = {
  '.salesorder.v1.AssignRouteRequest': AssignRouteRequest$json,
  '.salesorder.v1.AssignRouteResponse': AssignRouteResponse$json,
  '.salesorder.v1.ConfirmDispatchRequest': ConfirmDispatchRequest$json,
  '.salesorder.v1.ConfirmDispatchResponse': ConfirmDispatchResponse$json,
  '.salesorder.v1.DispatchItemResult': DispatchItemResult$json,
  '.salesorder.v1.CancelDispatchRequest': CancelDispatchRequest$json,
  '.salesorder.v1.CancelDispatchResponse': CancelDispatchResponse$json,
};

/// Descriptor for `DispatchService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List dispatchServiceDescriptor = $convert.base64Decode(
    'Cg9EaXNwYXRjaFNlcnZpY2USVAoLQXNzaWduUm91dGUSIS5zYWxlc29yZGVyLnYxLkFzc2lnbl'
    'JvdXRlUmVxdWVzdBoiLnNhbGVzb3JkZXIudjEuQXNzaWduUm91dGVSZXNwb25zZRJgCg9Db25m'
    'aXJtRGlzcGF0Y2gSJS5zYWxlc29yZGVyLnYxLkNvbmZpcm1EaXNwYXRjaFJlcXVlc3QaJi5zYW'
    'xlc29yZGVyLnYxLkNvbmZpcm1EaXNwYXRjaFJlc3BvbnNlEl0KDkNhbmNlbERpc3BhdGNoEiQu'
    'c2FsZXNvcmRlci52MS5DYW5jZWxEaXNwYXRjaFJlcXVlc3QaJS5zYWxlc29yZGVyLnYxLkNhbm'
    'NlbERpc3BhdGNoUmVzcG9uc2U=');
