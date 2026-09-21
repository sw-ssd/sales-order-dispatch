// This is a generated file - do not edit.
//
// Generated from products/v1/print.proto.

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

@$core.Deprecated('Use previewRequestDescriptor instead')
const PreviewRequest$json = {
  '1': 'PreviewRequest',
  '2': [
    {'1': 'document_type', '3': 1, '4': 1, '5': 9, '10': 'documentType'},
    {'1': 'route_id', '3': 2, '4': 1, '5': 9, '10': 'routeId'},
    {'1': 'target_date', '3': 3, '4': 1, '5': 9, '10': 'targetDate'},
    {'1': 'customer_id', '3': 4, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'warehouse_id', '3': 5, '4': 1, '5': 9, '10': 'warehouseId'},
  ],
};

/// Descriptor for `PreviewRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List previewRequestDescriptor = $convert.base64Decode(
    'Cg5QcmV2aWV3UmVxdWVzdBIjCg1kb2N1bWVudF90eXBlGAEgASgJUgxkb2N1bWVudFR5cGUSGQ'
    'oIcm91dGVfaWQYAiABKAlSB3JvdXRlSWQSHwoLdGFyZ2V0X2RhdGUYAyABKAlSCnRhcmdldERh'
    'dGUSHwoLY3VzdG9tZXJfaWQYBCABKAlSCmN1c3RvbWVySWQSIQoMd2FyZWhvdXNlX2lkGAUgAS'
    'gJUgt3YXJlaG91c2VJZA==');

@$core.Deprecated('Use previewResponseDescriptor instead')
const PreviewResponse$json = {
  '1': 'PreviewResponse',
  '2': [
    {'1': 'preview_id', '3': 1, '4': 1, '5': 9, '10': 'previewId'},
    {'1': 'file_asset_id', '3': 2, '4': 1, '5': 9, '10': 'fileAssetId'},
    {'1': 'download_url', '3': 3, '4': 1, '5': 9, '10': 'downloadUrl'},
  ],
};

/// Descriptor for `PreviewResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List previewResponseDescriptor = $convert.base64Decode(
    'Cg9QcmV2aWV3UmVzcG9uc2USHQoKcHJldmlld19pZBgBIAEoCVIJcHJldmlld0lkEiIKDWZpbG'
    'VfYXNzZXRfaWQYAiABKAlSC2ZpbGVBc3NldElkEiEKDGRvd25sb2FkX3VybBgDIAEoCVILZG93'
    'bmxvYWRVcmw=');

@$core.Deprecated('Use printRequestDescriptor instead')
const PrintRequest$json = {
  '1': 'PrintRequest',
  '2': [
    {'1': 'document_type', '3': 1, '4': 1, '5': 9, '10': 'documentType'},
    {'1': 'route_id', '3': 2, '4': 1, '5': 9, '10': 'routeId'},
    {'1': 'target_date', '3': 3, '4': 1, '5': 9, '10': 'targetDate'},
    {'1': 'customer_id', '3': 4, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'warehouse_id', '3': 5, '4': 1, '5': 9, '10': 'warehouseId'},
    {'1': 'reprint_reason', '3': 6, '4': 1, '5': 9, '10': 'reprintReason'},
  ],
};

/// Descriptor for `PrintRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List printRequestDescriptor = $convert.base64Decode(
    'CgxQcmludFJlcXVlc3QSIwoNZG9jdW1lbnRfdHlwZRgBIAEoCVIMZG9jdW1lbnRUeXBlEhkKCH'
    'JvdXRlX2lkGAIgASgJUgdyb3V0ZUlkEh8KC3RhcmdldF9kYXRlGAMgASgJUgp0YXJnZXREYXRl'
    'Eh8KC2N1c3RvbWVyX2lkGAQgASgJUgpjdXN0b21lcklkEiEKDHdhcmVob3VzZV9pZBgFIAEoCV'
    'ILd2FyZWhvdXNlSWQSJQoOcmVwcmludF9yZWFzb24YBiABKAlSDXJlcHJpbnRSZWFzb24=');

@$core.Deprecated('Use printResponseDescriptor instead')
const PrintResponse$json = {
  '1': 'PrintResponse',
  '2': [
    {'1': 'print_log_id', '3': 1, '4': 1, '5': 9, '10': 'printLogId'},
    {'1': 'file_asset_id', '3': 2, '4': 1, '5': 9, '10': 'fileAssetId'},
    {'1': 'download_url', '3': 3, '4': 1, '5': 9, '10': 'downloadUrl'},
    {'1': 'is_reprint', '3': 4, '4': 1, '5': 8, '10': 'isReprint'},
  ],
};

/// Descriptor for `PrintResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List printResponseDescriptor = $convert.base64Decode(
    'Cg1QcmludFJlc3BvbnNlEiAKDHByaW50X2xvZ19pZBgBIAEoCVIKcHJpbnRMb2dJZBIiCg1maW'
    'xlX2Fzc2V0X2lkGAIgASgJUgtmaWxlQXNzZXRJZBIhCgxkb3dubG9hZF91cmwYAyABKAlSC2Rv'
    'd25sb2FkVXJsEh0KCmlzX3JlcHJpbnQYBCABKAhSCWlzUmVwcmludA==');

@$core.Deprecated('Use listLogsRequestDescriptor instead')
const ListLogsRequest$json = {
  '1': 'ListLogsRequest',
  '2': [
    {'1': 'date_from', '3': 1, '4': 1, '5': 9, '10': 'dateFrom'},
    {'1': 'date_to', '3': 2, '4': 1, '5': 9, '10': 'dateTo'},
    {'1': 'document_type', '3': 3, '4': 1, '5': 9, '10': 'documentType'},
    {'1': 'route_id', '3': 4, '4': 1, '5': 9, '10': 'routeId'},
    {'1': 'page', '3': 5, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 6, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListLogsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLogsRequestDescriptor = $convert.base64Decode(
    'Cg9MaXN0TG9nc1JlcXVlc3QSGwoJZGF0ZV9mcm9tGAEgASgJUghkYXRlRnJvbRIXCgdkYXRlX3'
    'RvGAIgASgJUgZkYXRlVG8SIwoNZG9jdW1lbnRfdHlwZRgDIAEoCVIMZG9jdW1lbnRUeXBlEhkK'
    'CHJvdXRlX2lkGAQgASgJUgdyb3V0ZUlkEhIKBHBhZ2UYBSABKAVSBHBhZ2USGwoJcGFnZV9zaX'
    'plGAYgASgFUghwYWdlU2l6ZQ==');

@$core.Deprecated('Use logEntryDescriptor instead')
const LogEntry$json = {
  '1': 'LogEntry',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'document_type', '3': 2, '4': 1, '5': 9, '10': 'documentType'},
    {'1': 'route_id', '3': 3, '4': 1, '5': 9, '10': 'routeId'},
    {'1': 'target_date', '3': 4, '4': 1, '5': 9, '10': 'targetDate'},
    {'1': 'printed_by', '3': 5, '4': 1, '5': 9, '10': 'printedBy'},
    {'1': 'printed_at', '3': 6, '4': 1, '5': 9, '10': 'printedAt'},
    {'1': 'is_reprint', '3': 7, '4': 1, '5': 8, '10': 'isReprint'},
    {'1': 'reprint_reason', '3': 8, '4': 1, '5': 9, '10': 'reprintReason'},
    {'1': 'download_url', '3': 9, '4': 1, '5': 9, '10': 'downloadUrl'},
  ],
};

/// Descriptor for `LogEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logEntryDescriptor = $convert.base64Decode(
    'CghMb2dFbnRyeRIOCgJpZBgBIAEoCVICaWQSIwoNZG9jdW1lbnRfdHlwZRgCIAEoCVIMZG9jdW'
    '1lbnRUeXBlEhkKCHJvdXRlX2lkGAMgASgJUgdyb3V0ZUlkEh8KC3RhcmdldF9kYXRlGAQgASgJ'
    'Ugp0YXJnZXREYXRlEh0KCnByaW50ZWRfYnkYBSABKAlSCXByaW50ZWRCeRIdCgpwcmludGVkX2'
    'F0GAYgASgJUglwcmludGVkQXQSHQoKaXNfcmVwcmludBgHIAEoCFIJaXNSZXByaW50EiUKDnJl'
    'cHJpbnRfcmVhc29uGAggASgJUg1yZXByaW50UmVhc29uEiEKDGRvd25sb2FkX3VybBgJIAEoCV'
    'ILZG93bmxvYWRVcmw=');

@$core.Deprecated('Use listLogsResponseDescriptor instead')
const ListLogsResponse$json = {
  '1': 'ListLogsResponse',
  '2': [
    {
      '1': 'entries',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.products.v1.LogEntry',
      '10': 'entries'
    },
    {'1': 'page', '3': 2, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'total', '3': 4, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListLogsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLogsResponseDescriptor = $convert.base64Decode(
    'ChBMaXN0TG9nc1Jlc3BvbnNlEi8KB2VudHJpZXMYASADKAsyFS5wcm9kdWN0cy52MS5Mb2dFbn'
    'RyeVIHZW50cmllcxISCgRwYWdlGAIgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZRgDIAEoBVIIcGFn'
    'ZVNpemUSFAoFdG90YWwYBCABKAVSBXRvdGFs');

const $core.Map<$core.String, $core.dynamic> PrintServiceBase$json = {
  '1': 'PrintService',
  '2': [
    {
      '1': 'Preview',
      '2': '.products.v1.PreviewRequest',
      '3': '.products.v1.PreviewResponse'
    },
    {
      '1': 'Print',
      '2': '.products.v1.PrintRequest',
      '3': '.products.v1.PrintResponse'
    },
    {
      '1': 'ListLogs',
      '2': '.products.v1.ListLogsRequest',
      '3': '.products.v1.ListLogsResponse'
    },
  ],
};

@$core.Deprecated('Use printServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    PrintServiceBase$messageJson = {
  '.products.v1.PreviewRequest': PreviewRequest$json,
  '.products.v1.PreviewResponse': PreviewResponse$json,
  '.products.v1.PrintRequest': PrintRequest$json,
  '.products.v1.PrintResponse': PrintResponse$json,
  '.products.v1.ListLogsRequest': ListLogsRequest$json,
  '.products.v1.ListLogsResponse': ListLogsResponse$json,
  '.products.v1.LogEntry': LogEntry$json,
};

/// Descriptor for `PrintService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List printServiceDescriptor = $convert.base64Decode(
    'CgxQcmludFNlcnZpY2USRAoHUHJldmlldxIbLnByb2R1Y3RzLnYxLlByZXZpZXdSZXF1ZXN0Gh'
    'wucHJvZHVjdHMudjEuUHJldmlld1Jlc3BvbnNlEj4KBVByaW50EhkucHJvZHVjdHMudjEuUHJp'
    'bnRSZXF1ZXN0GhoucHJvZHVjdHMudjEuUHJpbnRSZXNwb25zZRJHCghMaXN0TG9ncxIcLnByb2'
    'R1Y3RzLnYxLkxpc3RMb2dzUmVxdWVzdBodLnByb2R1Y3RzLnYxLkxpc3RMb2dzUmVzcG9uc2U=');
