// This is a generated file - do not edit.
//
// Generated from salesorder/v1/common.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use printPaperSizeDescriptor instead')
const PrintPaperSize$json = {
  '1': 'PrintPaperSize',
  '2': [
    {'1': 'PRINT_PAPER_SIZE_UNSPECIFIED', '2': 0},
    {'1': 'PRINT_PAPER_SIZE_A4', '2': 1},
    {'1': 'PRINT_PAPER_SIZE_A5', '2': 2},
    {'1': 'PRINT_PAPER_SIZE_80MM', '2': 3},
  ],
};

/// Descriptor for `PrintPaperSize`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List printPaperSizeDescriptor = $convert.base64Decode(
    'Cg5QcmludFBhcGVyU2l6ZRIgChxQUklOVF9QQVBFUl9TSVpFX1VOU1BFQ0lGSUVEEAASFwoTUF'
    'JJTlRfUEFQRVJfU0laRV9BNBABEhcKE1BSSU5UX1BBUEVSX1NJWkVfQTUQAhIZChVQUklOVF9Q'
    'QVBFUl9TSVpFXzgwTU0QAw==');

@$core.Deprecated('Use paginationDescriptor instead')
const Pagination$json = {
  '1': 'Pagination',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'total', '3': 3, '4': 1, '5': 3, '10': 'total'},
  ],
};

/// Descriptor for `Pagination`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paginationDescriptor = $convert.base64Decode(
    'CgpQYWdpbmF0aW9uEhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaXplGAIgASgFUghwYW'
    'dlU2l6ZRIUCgV0b3RhbBgDIAEoA1IFdG90YWw=');

@$core.Deprecated('Use timestampRangeDescriptor instead')
const TimestampRange$json = {
  '1': 'TimestampRange',
  '2': [
    {'1': 'start_unix', '3': 1, '4': 1, '5': 3, '10': 'startUnix'},
    {'1': 'end_unix', '3': 2, '4': 1, '5': 3, '10': 'endUnix'},
  ],
};

/// Descriptor for `TimestampRange`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List timestampRangeDescriptor = $convert.base64Decode(
    'Cg5UaW1lc3RhbXBSYW5nZRIdCgpzdGFydF91bml4GAEgASgDUglzdGFydFVuaXgSGQoIZW5kX3'
    'VuaXgYAiABKANSB2VuZFVuaXg=');

@$core.Deprecated('Use moneyDescriptor instead')
const Money$json = {
  '1': 'Money',
  '2': [
    {'1': 'amount_minor', '3': 1, '4': 1, '5': 3, '10': 'amountMinor'},
    {'1': 'currency', '3': 2, '4': 1, '5': 9, '10': 'currency'},
  ],
  '7': {'3': true},
};

/// Descriptor for `Money`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List moneyDescriptor = $convert.base64Decode(
    'CgVNb25leRIhCgxhbW91bnRfbWlub3IYASABKANSC2Ftb3VudE1pbm9yEhoKCGN1cnJlbmN5GA'
    'IgASgJUghjdXJyZW5jeToCGAE=');
