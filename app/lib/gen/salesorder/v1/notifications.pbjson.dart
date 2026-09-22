// This is a generated file - do not edit.
//
// Generated from salesorder/v1/notifications.proto.

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

@$core.Deprecated('Use notificationViewDescriptor instead')
const NotificationView$json = {
  '1': 'NotificationView',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'channel', '3': 2, '4': 1, '5': 9, '10': 'channel'},
    {'1': 'title', '3': 3, '4': 1, '5': 9, '10': 'title'},
    {'1': 'content', '3': 4, '4': 1, '5': 9, '10': 'content'},
    {'1': 'payload', '3': 5, '4': 1, '5': 9, '10': 'payload'},
    {'1': 'status', '3': 6, '4': 1, '5': 9, '10': 'status'},
    {'1': 'sent_at', '3': 7, '4': 1, '5': 9, '10': 'sentAt'},
    {'1': 'read_at', '3': 8, '4': 1, '5': 9, '10': 'readAt'},
    {'1': 'created_at', '3': 9, '4': 1, '5': 9, '10': 'createdAt'},
  ],
};

/// Descriptor for `NotificationView`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List notificationViewDescriptor = $convert.base64Decode(
    'ChBOb3RpZmljYXRpb25WaWV3Eg4KAmlkGAEgASgJUgJpZBIYCgdjaGFubmVsGAIgASgJUgdjaG'
    'FubmVsEhQKBXRpdGxlGAMgASgJUgV0aXRsZRIYCgdjb250ZW50GAQgASgJUgdjb250ZW50EhgK'
    'B3BheWxvYWQYBSABKAlSB3BheWxvYWQSFgoGc3RhdHVzGAYgASgJUgZzdGF0dXMSFwoHc2VudF'
    '9hdBgHIAEoCVIGc2VudEF0EhcKB3JlYWRfYXQYCCABKAlSBnJlYWRBdBIdCgpjcmVhdGVkX2F0'
    'GAkgASgJUgljcmVhdGVkQXQ=');

@$core.Deprecated('Use listNotificationsRequestDescriptor instead')
const ListNotificationsRequest$json = {
  '1': 'ListNotificationsRequest',
  '2': [
    {'1': 'unread_only', '3': 1, '4': 1, '5': 8, '10': 'unreadOnly'},
    {'1': 'channel', '3': 2, '4': 1, '5': 9, '10': 'channel'},
    {'1': 'page', '3': 3, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 4, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListNotificationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listNotificationsRequestDescriptor = $convert.base64Decode(
    'ChhMaXN0Tm90aWZpY2F0aW9uc1JlcXVlc3QSHwoLdW5yZWFkX29ubHkYASABKAhSCnVucmVhZE'
    '9ubHkSGAoHY2hhbm5lbBgCIAEoCVIHY2hhbm5lbBISCgRwYWdlGAMgASgFUgRwYWdlEhsKCXBh'
    'Z2Vfc2l6ZRgEIAEoBVIIcGFnZVNpemU=');

@$core.Deprecated('Use listNotificationsResponseDescriptor instead')
const ListNotificationsResponse$json = {
  '1': 'ListNotificationsResponse',
  '2': [
    {
      '1': 'notifications',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.NotificationView',
      '10': 'notifications'
    },
    {'1': 'page', '3': 2, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'total', '3': 4, '4': 1, '5': 5, '10': 'total'},
    {'1': 'unread_count', '3': 5, '4': 1, '5': 5, '10': 'unreadCount'},
  ],
};

/// Descriptor for `ListNotificationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listNotificationsResponseDescriptor = $convert.base64Decode(
    'ChlMaXN0Tm90aWZpY2F0aW9uc1Jlc3BvbnNlEkUKDW5vdGlmaWNhdGlvbnMYASADKAsyHy5zYW'
    'xlc29yZGVyLnYxLk5vdGlmaWNhdGlvblZpZXdSDW5vdGlmaWNhdGlvbnMSEgoEcGFnZRgCIAEo'
    'BVIEcGFnZRIbCglwYWdlX3NpemUYAyABKAVSCHBhZ2VTaXplEhQKBXRvdGFsGAQgASgFUgV0b3'
    'RhbBIhCgx1bnJlYWRfY291bnQYBSABKAVSC3VucmVhZENvdW50');

@$core.Deprecated('Use markReadRequestDescriptor instead')
const MarkReadRequest$json = {
  '1': 'MarkReadRequest',
  '2': [
    {'1': 'notification_ids', '3': 1, '4': 3, '5': 9, '10': 'notificationIds'},
  ],
};

/// Descriptor for `MarkReadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markReadRequestDescriptor = $convert.base64Decode(
    'Cg9NYXJrUmVhZFJlcXVlc3QSKQoQbm90aWZpY2F0aW9uX2lkcxgBIAMoCVIPbm90aWZpY2F0aW'
    '9uSWRz');

@$core.Deprecated('Use markReadResponseDescriptor instead')
const MarkReadResponse$json = {
  '1': 'MarkReadResponse',
  '2': [
    {'1': 'marked_count', '3': 1, '4': 1, '5': 5, '10': 'markedCount'},
  ],
};

/// Descriptor for `MarkReadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markReadResponseDescriptor = $convert.base64Decode(
    'ChBNYXJrUmVhZFJlc3BvbnNlEiEKDG1hcmtlZF9jb3VudBgBIAEoBVILbWFya2VkQ291bnQ=');

@$core.Deprecated('Use unreadCountRequestDescriptor instead')
const UnreadCountRequest$json = {
  '1': 'UnreadCountRequest',
};

/// Descriptor for `UnreadCountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List unreadCountRequestDescriptor =
    $convert.base64Decode('ChJVbnJlYWRDb3VudFJlcXVlc3Q=');

@$core.Deprecated('Use unreadCountResponseDescriptor instead')
const UnreadCountResponse$json = {
  '1': 'UnreadCountResponse',
  '2': [
    {'1': 'count', '3': 1, '4': 1, '5': 5, '10': 'count'},
  ],
};

/// Descriptor for `UnreadCountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List unreadCountResponseDescriptor =
    $convert.base64Decode(
        'ChNVbnJlYWRDb3VudFJlc3BvbnNlEhQKBWNvdW50GAEgASgFUgVjb3VudA==');

const $core.Map<$core.String, $core.dynamic> NotificationServiceBase$json = {
  '1': 'NotificationService',
  '2': [
    {
      '1': 'ListNotifications',
      '2': '.salesorder.v1.ListNotificationsRequest',
      '3': '.salesorder.v1.ListNotificationsResponse'
    },
    {
      '1': 'MarkRead',
      '2': '.salesorder.v1.MarkReadRequest',
      '3': '.salesorder.v1.MarkReadResponse'
    },
    {
      '1': 'UnreadCount',
      '2': '.salesorder.v1.UnreadCountRequest',
      '3': '.salesorder.v1.UnreadCountResponse'
    },
  ],
};

@$core.Deprecated('Use notificationServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    NotificationServiceBase$messageJson = {
  '.salesorder.v1.ListNotificationsRequest': ListNotificationsRequest$json,
  '.salesorder.v1.ListNotificationsResponse': ListNotificationsResponse$json,
  '.salesorder.v1.NotificationView': NotificationView$json,
  '.salesorder.v1.MarkReadRequest': MarkReadRequest$json,
  '.salesorder.v1.MarkReadResponse': MarkReadResponse$json,
  '.salesorder.v1.UnreadCountRequest': UnreadCountRequest$json,
  '.salesorder.v1.UnreadCountResponse': UnreadCountResponse$json,
};

/// Descriptor for `NotificationService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List notificationServiceDescriptor = $convert.base64Decode(
    'ChNOb3RpZmljYXRpb25TZXJ2aWNlEmYKEUxpc3ROb3RpZmljYXRpb25zEicuc2FsZXNvcmRlci'
    '52MS5MaXN0Tm90aWZpY2F0aW9uc1JlcXVlc3QaKC5zYWxlc29yZGVyLnYxLkxpc3ROb3RpZmlj'
    'YXRpb25zUmVzcG9uc2USSwoITWFya1JlYWQSHi5zYWxlc29yZGVyLnYxLk1hcmtSZWFkUmVxdW'
    'VzdBofLnNhbGVzb3JkZXIudjEuTWFya1JlYWRSZXNwb25zZRJUCgtVbnJlYWRDb3VudBIhLnNh'
    'bGVzb3JkZXIudjEuVW5yZWFkQ291bnRSZXF1ZXN0GiIuc2FsZXNvcmRlci52MS5VbnJlYWRDb3'
    'VudFJlc3BvbnNl');
