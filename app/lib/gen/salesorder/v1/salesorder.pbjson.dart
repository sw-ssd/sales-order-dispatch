// This is a generated file - do not edit.
//
// Generated from salesorder/v1/salesorder.proto.

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

@$core.Deprecated('Use salesOrderDescriptor instead')
const SalesOrder$json = {
  '1': 'SalesOrder',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'order_no', '3': 4, '4': 1, '5': 9, '10': 'orderNo'},
    {'1': 'customer_id', '3': 5, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'source', '3': 6, '4': 1, '5': 9, '10': 'source'},
    {'1': 'status', '3': 7, '4': 1, '5': 9, '10': 'status'},
    {
      '1': 'expected_delivery_date',
      '3': 8,
      '4': 1,
      '5': 9,
      '10': 'expectedDeliveryDate'
    },
    {'1': 'sales_rep_id', '3': 9, '4': 1, '5': 9, '10': 'salesRepId'},
    {'1': 'note', '3': 10, '4': 1, '5': 9, '10': 'note'},
    {'1': 'dispatched_at', '3': 11, '4': 1, '5': 9, '10': 'dispatchedAt'},
    {'1': 'dispatched_by', '3': 12, '4': 1, '5': 9, '10': 'dispatchedBy'},
    {'1': 'route_id', '3': 13, '4': 1, '5': 9, '10': 'routeId'},
    {
      '1': 'delivery_sequence',
      '3': 14,
      '4': 1,
      '5': 5,
      '10': 'deliverySequence'
    },
    {'1': 'version', '3': 15, '4': 1, '5': 5, '10': 'version'},
    {'1': 'created_at', '3': 16, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 17, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 18, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `SalesOrder`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List salesOrderDescriptor = $convert.base64Decode(
    'CgpTYWxlc09yZGVyEg4KAmlkGAEgASgJUgJpZBIdCgpjb21wYW55X2lkGAIgASgJUgljb21wYW'
    '55SWQSIwoNZGVwYXJ0bWVudF9pZBgDIAEoCVIMZGVwYXJ0bWVudElkEhkKCG9yZGVyX25vGAQg'
    'ASgJUgdvcmRlck5vEh8KC2N1c3RvbWVyX2lkGAUgASgJUgpjdXN0b21lcklkEhYKBnNvdXJjZR'
    'gGIAEoCVIGc291cmNlEhYKBnN0YXR1cxgHIAEoCVIGc3RhdHVzEjQKFmV4cGVjdGVkX2RlbGl2'
    'ZXJ5X2RhdGUYCCABKAlSFGV4cGVjdGVkRGVsaXZlcnlEYXRlEiAKDHNhbGVzX3JlcF9pZBgJIA'
    'EoCVIKc2FsZXNSZXBJZBISCgRub3RlGAogASgJUgRub3RlEiMKDWRpc3BhdGNoZWRfYXQYCyAB'
    'KAlSDGRpc3BhdGNoZWRBdBIjCg1kaXNwYXRjaGVkX2J5GAwgASgJUgxkaXNwYXRjaGVkQnkSGQ'
    'oIcm91dGVfaWQYDSABKAlSB3JvdXRlSWQSKwoRZGVsaXZlcnlfc2VxdWVuY2UYDiABKAVSEGRl'
    'bGl2ZXJ5U2VxdWVuY2USGAoHdmVyc2lvbhgPIAEoBVIHdmVyc2lvbhIdCgpjcmVhdGVkX2F0GB'
    'AgASgJUgljcmVhdGVkQXQSHQoKdXBkYXRlZF9hdBgRIAEoCVIJdXBkYXRlZEF0Eh0KCmRlbGV0'
    'ZWRfYXQYEiABKAlSCWRlbGV0ZWRBdA==');

@$core.Deprecated('Use salesOrderItemDescriptor instead')
const SalesOrderItem$json = {
  '1': 'SalesOrderItem',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'product_id', '3': 2, '4': 1, '5': 9, '10': 'productId'},
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'qty', '3': 4, '4': 1, '5': 9, '10': 'qty'},
    {'1': 'unit', '3': 5, '4': 1, '5': 9, '10': 'unit'},
    {'1': 'base_qty', '3': 6, '4': 1, '5': 9, '10': 'baseQty'},
    {
      '1': 'processing_spec_id',
      '3': 7,
      '4': 1,
      '5': 9,
      '10': 'processingSpecId'
    },
    {'1': 'special_cut_note', '3': 8, '4': 1, '5': 9, '10': 'specialCutNote'},
    {'1': 'warehouse_id', '3': 9, '4': 1, '5': 9, '10': 'warehouseId'},
    {'1': 'sort_order', '3': 10, '4': 1, '5': 5, '10': 'sortOrder'},
  ],
};

/// Descriptor for `SalesOrderItem`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List salesOrderItemDescriptor = $convert.base64Decode(
    'Cg5TYWxlc09yZGVySXRlbRIOCgJpZBgBIAEoCVICaWQSHQoKcHJvZHVjdF9pZBgCIAEoCVIJcH'
    'JvZHVjdElkEiEKDGRpc3BsYXlfbmFtZRgDIAEoCVILZGlzcGxheU5hbWUSEAoDcXR5GAQgASgJ'
    'UgNxdHkSEgoEdW5pdBgFIAEoCVIEdW5pdBIZCghiYXNlX3F0eRgGIAEoCVIHYmFzZVF0eRIsCh'
    'Jwcm9jZXNzaW5nX3NwZWNfaWQYByABKAlSEHByb2Nlc3NpbmdTcGVjSWQSKAoQc3BlY2lhbF9j'
    'dXRfbm90ZRgIIAEoCVIOc3BlY2lhbEN1dE5vdGUSIQoMd2FyZWhvdXNlX2lkGAkgASgJUgt3YX'
    'JlaG91c2VJZBIdCgpzb3J0X29yZGVyGAogASgFUglzb3J0T3JkZXI=');

@$core.Deprecated('Use salesOrderEventDescriptor instead')
const SalesOrderEvent$json = {
  '1': 'SalesOrderEvent',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'event_type', '3': 2, '4': 1, '5': 9, '10': 'eventType'},
    {'1': 'actor_id', '3': 3, '4': 1, '5': 9, '10': 'actorId'},
    {'1': 'reason', '3': 4, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'payload', '3': 5, '4': 1, '5': 9, '10': 'payload'},
    {'1': 'created_at', '3': 6, '4': 1, '5': 9, '10': 'createdAt'},
  ],
};

/// Descriptor for `SalesOrderEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List salesOrderEventDescriptor = $convert.base64Decode(
    'Cg9TYWxlc09yZGVyRXZlbnQSDgoCaWQYASABKAlSAmlkEh0KCmV2ZW50X3R5cGUYAiABKAlSCW'
    'V2ZW50VHlwZRIZCghhY3Rvcl9pZBgDIAEoCVIHYWN0b3JJZBIWCgZyZWFzb24YBCABKAlSBnJl'
    'YXNvbhIYCgdwYXlsb2FkGAUgASgJUgdwYXlsb2FkEh0KCmNyZWF0ZWRfYXQYBiABKAlSCWNyZW'
    'F0ZWRBdA==');

@$core.Deprecated('Use listOrdersRequestDescriptor instead')
const ListOrdersRequest$json = {
  '1': 'ListOrdersRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'status', '3': 3, '4': 1, '5': 9, '10': 'status'},
    {'1': 'customer_id', '3': 4, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'source', '3': 5, '4': 1, '5': 9, '10': 'source'},
    {'1': 'keyword', '3': 6, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'include_deleted', '3': 7, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListOrdersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOrdersRequestDescriptor = $convert.base64Decode(
    'ChFMaXN0T3JkZXJzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZRgCIA'
    'EoBVIIcGFnZVNpemUSFgoGc3RhdHVzGAMgASgJUgZzdGF0dXMSHwoLY3VzdG9tZXJfaWQYBCAB'
    'KAlSCmN1c3RvbWVySWQSFgoGc291cmNlGAUgASgJUgZzb3VyY2USGAoHa2V5d29yZBgGIAEoCV'
    'IHa2V5d29yZBInCg9pbmNsdWRlX2RlbGV0ZWQYByABKAhSDmluY2x1ZGVEZWxldGVk');

@$core.Deprecated('Use listOrdersResponseDescriptor instead')
const ListOrdersResponse$json = {
  '1': 'ListOrdersResponse',
  '2': [
    {
      '1': 'orders',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'orders'
    },
    {'1': 'total', '3': 2, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListOrdersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOrdersResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0T3JkZXJzUmVzcG9uc2USMQoGb3JkZXJzGAEgAygLMhkuc2FsZXNvcmRlci52MS5TYW'
    'xlc09yZGVyUgZvcmRlcnMSFAoFdG90YWwYAiABKAVSBXRvdGFs');

@$core.Deprecated('Use getOrderRequestDescriptor instead')
const GetOrderRequest$json = {
  '1': 'GetOrderRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getOrderRequestDescriptor =
    $convert.base64Decode('Cg9HZXRPcmRlclJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getOrderResponseDescriptor instead')
const GetOrderResponse$json = {
  '1': 'GetOrderResponse',
  '2': [
    {
      '1': 'order',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'order'
    },
    {
      '1': 'items',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.SalesOrderItem',
      '10': 'items'
    },
  ],
};

/// Descriptor for `GetOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getOrderResponseDescriptor = $convert.base64Decode(
    'ChBHZXRPcmRlclJlc3BvbnNlEi8KBW9yZGVyGAEgASgLMhkuc2FsZXNvcmRlci52MS5TYWxlc0'
    '9yZGVyUgVvcmRlchIzCgVpdGVtcxgCIAMoCzIdLnNhbGVzb3JkZXIudjEuU2FsZXNPcmRlckl0'
    'ZW1SBWl0ZW1z');

@$core.Deprecated('Use orderItemInputDescriptor instead')
const OrderItemInput$json = {
  '1': 'OrderItemInput',
  '2': [
    {'1': 'product_id', '3': 1, '4': 1, '5': 9, '10': 'productId'},
    {'1': 'manual_name', '3': 2, '4': 1, '5': 9, '10': 'manualName'},
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'qty', '3': 4, '4': 1, '5': 9, '10': 'qty'},
    {'1': 'unit', '3': 5, '4': 1, '5': 9, '10': 'unit'},
    {
      '1': 'processing_spec_id',
      '3': 6,
      '4': 1,
      '5': 9,
      '10': 'processingSpecId'
    },
    {'1': 'special_cut_note', '3': 7, '4': 1, '5': 9, '10': 'specialCutNote'},
    {'1': 'warehouse_id', '3': 8, '4': 1, '5': 9, '10': 'warehouseId'},
    {'1': 'save_alias', '3': 9, '4': 1, '5': 8, '10': 'saveAlias'},
  ],
};

/// Descriptor for `OrderItemInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List orderItemInputDescriptor = $convert.base64Decode(
    'Cg5PcmRlckl0ZW1JbnB1dBIdCgpwcm9kdWN0X2lkGAEgASgJUglwcm9kdWN0SWQSHwoLbWFudW'
    'FsX25hbWUYAiABKAlSCm1hbnVhbE5hbWUSIQoMZGlzcGxheV9uYW1lGAMgASgJUgtkaXNwbGF5'
    'TmFtZRIQCgNxdHkYBCABKAlSA3F0eRISCgR1bml0GAUgASgJUgR1bml0EiwKEnByb2Nlc3Npbm'
    'dfc3BlY19pZBgGIAEoCVIQcHJvY2Vzc2luZ1NwZWNJZBIoChBzcGVjaWFsX2N1dF9ub3RlGAcg'
    'ASgJUg5zcGVjaWFsQ3V0Tm90ZRIhCgx3YXJlaG91c2VfaWQYCCABKAlSC3dhcmVob3VzZUlkEh'
    '0KCnNhdmVfYWxpYXMYCSABKAhSCXNhdmVBbGlhcw==');

@$core.Deprecated('Use createOrderRequestDescriptor instead')
const CreateOrderRequest$json = {
  '1': 'CreateOrderRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'source', '3': 2, '4': 1, '5': 9, '10': 'source'},
    {
      '1': 'expected_delivery_date',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'expectedDeliveryDate'
    },
    {'1': 'note', '3': 4, '4': 1, '5': 9, '10': 'note'},
    {'1': 'sales_rep_id', '3': 5, '4': 1, '5': 9, '10': 'salesRepId'},
    {
      '1': 'items',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.OrderItemInput',
      '10': 'items'
    },
  ],
};

/// Descriptor for `CreateOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createOrderRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVPcmRlclJlcXVlc3QSHwoLY3VzdG9tZXJfaWQYASABKAlSCmN1c3RvbWVySWQSFg'
    'oGc291cmNlGAIgASgJUgZzb3VyY2USNAoWZXhwZWN0ZWRfZGVsaXZlcnlfZGF0ZRgDIAEoCVIU'
    'ZXhwZWN0ZWREZWxpdmVyeURhdGUSEgoEbm90ZRgEIAEoCVIEbm90ZRIgCgxzYWxlc19yZXBfaW'
    'QYBSABKAlSCnNhbGVzUmVwSWQSMwoFaXRlbXMYBiADKAsyHS5zYWxlc29yZGVyLnYxLk9yZGVy'
    'SXRlbUlucHV0UgVpdGVtcw==');

@$core.Deprecated('Use createOrderResponseDescriptor instead')
const CreateOrderResponse$json = {
  '1': 'CreateOrderResponse',
  '2': [
    {
      '1': 'order',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'order'
    },
  ],
};

/// Descriptor for `CreateOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createOrderResponseDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVPcmRlclJlc3BvbnNlEi8KBW9yZGVyGAEgASgLMhkuc2FsZXNvcmRlci52MS5TYW'
    'xlc09yZGVyUgVvcmRlcg==');

@$core.Deprecated('Use updateOrderRequestDescriptor instead')
const UpdateOrderRequest$json = {
  '1': 'UpdateOrderRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'version', '3': 2, '4': 1, '5': 5, '10': 'version'},
    {
      '1': 'expected_delivery_date',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'expectedDeliveryDate'
    },
    {'1': 'note', '3': 4, '4': 1, '5': 9, '10': 'note'},
    {
      '1': 'items',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.OrderItemInput',
      '10': 'items'
    },
  ],
};

/// Descriptor for `UpdateOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateOrderRequestDescriptor = $convert.base64Decode(
    'ChJVcGRhdGVPcmRlclJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhgKB3ZlcnNpb24YAiABKAVSB3'
    'ZlcnNpb24SNAoWZXhwZWN0ZWRfZGVsaXZlcnlfZGF0ZRgDIAEoCVIUZXhwZWN0ZWREZWxpdmVy'
    'eURhdGUSEgoEbm90ZRgEIAEoCVIEbm90ZRIzCgVpdGVtcxgFIAMoCzIdLnNhbGVzb3JkZXIudj'
    'EuT3JkZXJJdGVtSW5wdXRSBWl0ZW1z');

@$core.Deprecated('Use updateOrderResponseDescriptor instead')
const UpdateOrderResponse$json = {
  '1': 'UpdateOrderResponse',
  '2': [
    {
      '1': 'order',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'order'
    },
  ],
};

/// Descriptor for `UpdateOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateOrderResponseDescriptor = $convert.base64Decode(
    'ChNVcGRhdGVPcmRlclJlc3BvbnNlEi8KBW9yZGVyGAEgASgLMhkuc2FsZXNvcmRlci52MS5TYW'
    'xlc09yZGVyUgVvcmRlcg==');

@$core.Deprecated('Use cancelOrderRequestDescriptor instead')
const CancelOrderRequest$json = {
  '1': 'CancelOrderRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `CancelOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelOrderRequestDescriptor =
    $convert.base64Decode('ChJDYW5jZWxPcmRlclJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use cancelOrderResponseDescriptor instead')
const CancelOrderResponse$json = {
  '1': 'CancelOrderResponse',
  '2': [
    {
      '1': 'order',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'order'
    },
  ],
};

/// Descriptor for `CancelOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelOrderResponseDescriptor = $convert.base64Decode(
    'ChNDYW5jZWxPcmRlclJlc3BvbnNlEi8KBW9yZGVyGAEgASgLMhkuc2FsZXNvcmRlci52MS5TYW'
    'xlc09yZGVyUgVvcmRlcg==');

@$core.Deprecated('Use completeOrderRequestDescriptor instead')
const CompleteOrderRequest$json = {
  '1': 'CompleteOrderRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `CompleteOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeOrderRequestDescriptor = $convert
    .base64Decode('ChRDb21wbGV0ZU9yZGVyUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use completeOrderResponseDescriptor instead')
const CompleteOrderResponse$json = {
  '1': 'CompleteOrderResponse',
  '2': [
    {
      '1': 'order',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'order'
    },
  ],
};

/// Descriptor for `CompleteOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeOrderResponseDescriptor = $convert.base64Decode(
    'ChVDb21wbGV0ZU9yZGVyUmVzcG9uc2USLwoFb3JkZXIYASABKAsyGS5zYWxlc29yZGVyLnYxLl'
    'NhbGVzT3JkZXJSBW9yZGVy');

@$core.Deprecated('Use voidOrderRequestDescriptor instead')
const VoidOrderRequest$json = {
  '1': 'VoidOrderRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'reason', '3': 2, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `VoidOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidOrderRequestDescriptor = $convert.base64Decode(
    'ChBWb2lkT3JkZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIWCgZyZWFzb24YAiABKAlSBnJlYX'
    'Nvbg==');

@$core.Deprecated('Use voidOrderResponseDescriptor instead')
const VoidOrderResponse$json = {
  '1': 'VoidOrderResponse',
  '2': [
    {
      '1': 'order',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.SalesOrder',
      '10': 'order'
    },
  ],
};

/// Descriptor for `VoidOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidOrderResponseDescriptor = $convert.base64Decode(
    'ChFWb2lkT3JkZXJSZXNwb25zZRIvCgVvcmRlchgBIAEoCzIZLnNhbGVzb3JkZXIudjEuU2FsZX'
    'NPcmRlclIFb3JkZXI=');

@$core.Deprecated('Use deleteOrderRequestDescriptor instead')
const DeleteOrderRequest$json = {
  '1': 'DeleteOrderRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteOrderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteOrderRequestDescriptor =
    $convert.base64Decode('ChJEZWxldGVPcmRlclJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteOrderResponseDescriptor instead')
const DeleteOrderResponse$json = {
  '1': 'DeleteOrderResponse',
};

/// Descriptor for `DeleteOrderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteOrderResponseDescriptor =
    $convert.base64Decode('ChNEZWxldGVPcmRlclJlc3BvbnNl');

@$core.Deprecated('Use listOrderEventsRequestDescriptor instead')
const ListOrderEventsRequest$json = {
  '1': 'ListOrderEventsRequest',
  '2': [
    {'1': 'sales_order_id', '3': 1, '4': 1, '5': 9, '10': 'salesOrderId'},
    {'1': 'page', '3': 2, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListOrderEventsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOrderEventsRequestDescriptor = $convert.base64Decode(
    'ChZMaXN0T3JkZXJFdmVudHNSZXF1ZXN0EiQKDnNhbGVzX29yZGVyX2lkGAEgASgJUgxzYWxlc0'
    '9yZGVySWQSEgoEcGFnZRgCIAEoBVIEcGFnZRIbCglwYWdlX3NpemUYAyABKAVSCHBhZ2VTaXpl');

@$core.Deprecated('Use listOrderEventsResponseDescriptor instead')
const ListOrderEventsResponse$json = {
  '1': 'ListOrderEventsResponse',
  '2': [
    {
      '1': 'events',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.SalesOrderEvent',
      '10': 'events'
    },
    {'1': 'total', '3': 2, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListOrderEventsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOrderEventsResponseDescriptor =
    $convert.base64Decode(
        'ChdMaXN0T3JkZXJFdmVudHNSZXNwb25zZRI2CgZldmVudHMYASADKAsyHi5zYWxlc29yZGVyLn'
        'YxLlNhbGVzT3JkZXJFdmVudFIGZXZlbnRzEhQKBXRvdGFsGAIgASgFUgV0b3RhbA==');

const $core.Map<$core.String, $core.dynamic> SalesOrderServiceBase$json = {
  '1': 'SalesOrderService',
  '2': [
    {
      '1': 'ListOrders',
      '2': '.salesorder.v1.ListOrdersRequest',
      '3': '.salesorder.v1.ListOrdersResponse'
    },
    {
      '1': 'GetOrder',
      '2': '.salesorder.v1.GetOrderRequest',
      '3': '.salesorder.v1.GetOrderResponse'
    },
    {
      '1': 'CreateOrder',
      '2': '.salesorder.v1.CreateOrderRequest',
      '3': '.salesorder.v1.CreateOrderResponse'
    },
    {
      '1': 'UpdateOrder',
      '2': '.salesorder.v1.UpdateOrderRequest',
      '3': '.salesorder.v1.UpdateOrderResponse'
    },
    {
      '1': 'CancelOrder',
      '2': '.salesorder.v1.CancelOrderRequest',
      '3': '.salesorder.v1.CancelOrderResponse'
    },
    {
      '1': 'CompleteOrder',
      '2': '.salesorder.v1.CompleteOrderRequest',
      '3': '.salesorder.v1.CompleteOrderResponse'
    },
    {
      '1': 'VoidOrder',
      '2': '.salesorder.v1.VoidOrderRequest',
      '3': '.salesorder.v1.VoidOrderResponse'
    },
    {
      '1': 'DeleteOrder',
      '2': '.salesorder.v1.DeleteOrderRequest',
      '3': '.salesorder.v1.DeleteOrderResponse'
    },
    {
      '1': 'ListOrderEvents',
      '2': '.salesorder.v1.ListOrderEventsRequest',
      '3': '.salesorder.v1.ListOrderEventsResponse'
    },
  ],
};

@$core.Deprecated('Use salesOrderServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    SalesOrderServiceBase$messageJson = {
  '.salesorder.v1.ListOrdersRequest': ListOrdersRequest$json,
  '.salesorder.v1.ListOrdersResponse': ListOrdersResponse$json,
  '.salesorder.v1.SalesOrder': SalesOrder$json,
  '.salesorder.v1.GetOrderRequest': GetOrderRequest$json,
  '.salesorder.v1.GetOrderResponse': GetOrderResponse$json,
  '.salesorder.v1.SalesOrderItem': SalesOrderItem$json,
  '.salesorder.v1.CreateOrderRequest': CreateOrderRequest$json,
  '.salesorder.v1.OrderItemInput': OrderItemInput$json,
  '.salesorder.v1.CreateOrderResponse': CreateOrderResponse$json,
  '.salesorder.v1.UpdateOrderRequest': UpdateOrderRequest$json,
  '.salesorder.v1.UpdateOrderResponse': UpdateOrderResponse$json,
  '.salesorder.v1.CancelOrderRequest': CancelOrderRequest$json,
  '.salesorder.v1.CancelOrderResponse': CancelOrderResponse$json,
  '.salesorder.v1.CompleteOrderRequest': CompleteOrderRequest$json,
  '.salesorder.v1.CompleteOrderResponse': CompleteOrderResponse$json,
  '.salesorder.v1.VoidOrderRequest': VoidOrderRequest$json,
  '.salesorder.v1.VoidOrderResponse': VoidOrderResponse$json,
  '.salesorder.v1.DeleteOrderRequest': DeleteOrderRequest$json,
  '.salesorder.v1.DeleteOrderResponse': DeleteOrderResponse$json,
  '.salesorder.v1.ListOrderEventsRequest': ListOrderEventsRequest$json,
  '.salesorder.v1.ListOrderEventsResponse': ListOrderEventsResponse$json,
  '.salesorder.v1.SalesOrderEvent': SalesOrderEvent$json,
};

/// Descriptor for `SalesOrderService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List salesOrderServiceDescriptor = $convert.base64Decode(
    'ChFTYWxlc09yZGVyU2VydmljZRJRCgpMaXN0T3JkZXJzEiAuc2FsZXNvcmRlci52MS5MaXN0T3'
    'JkZXJzUmVxdWVzdBohLnNhbGVzb3JkZXIudjEuTGlzdE9yZGVyc1Jlc3BvbnNlEksKCEdldE9y'
    'ZGVyEh4uc2FsZXNvcmRlci52MS5HZXRPcmRlclJlcXVlc3QaHy5zYWxlc29yZGVyLnYxLkdldE'
    '9yZGVyUmVzcG9uc2USVAoLQ3JlYXRlT3JkZXISIS5zYWxlc29yZGVyLnYxLkNyZWF0ZU9yZGVy'
    'UmVxdWVzdBoiLnNhbGVzb3JkZXIudjEuQ3JlYXRlT3JkZXJSZXNwb25zZRJUCgtVcGRhdGVPcm'
    'RlchIhLnNhbGVzb3JkZXIudjEuVXBkYXRlT3JkZXJSZXF1ZXN0GiIuc2FsZXNvcmRlci52MS5V'
    'cGRhdGVPcmRlclJlc3BvbnNlElQKC0NhbmNlbE9yZGVyEiEuc2FsZXNvcmRlci52MS5DYW5jZW'
    'xPcmRlclJlcXVlc3QaIi5zYWxlc29yZGVyLnYxLkNhbmNlbE9yZGVyUmVzcG9uc2USWgoNQ29t'
    'cGxldGVPcmRlchIjLnNhbGVzb3JkZXIudjEuQ29tcGxldGVPcmRlclJlcXVlc3QaJC5zYWxlc2'
    '9yZGVyLnYxLkNvbXBsZXRlT3JkZXJSZXNwb25zZRJOCglWb2lkT3JkZXISHy5zYWxlc29yZGVy'
    'LnYxLlZvaWRPcmRlclJlcXVlc3QaIC5zYWxlc29yZGVyLnYxLlZvaWRPcmRlclJlc3BvbnNlEl'
    'QKC0RlbGV0ZU9yZGVyEiEuc2FsZXNvcmRlci52MS5EZWxldGVPcmRlclJlcXVlc3QaIi5zYWxl'
    'c29yZGVyLnYxLkRlbGV0ZU9yZGVyUmVzcG9uc2USYAoPTGlzdE9yZGVyRXZlbnRzEiUuc2FsZX'
    'NvcmRlci52MS5MaXN0T3JkZXJFdmVudHNSZXF1ZXN0GiYuc2FsZXNvcmRlci52MS5MaXN0T3Jk'
    'ZXJFdmVudHNSZXNwb25zZQ==');
