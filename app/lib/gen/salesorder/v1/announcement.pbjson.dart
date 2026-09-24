// This is a generated file - do not edit.
//
// Generated from salesorder/v1/announcement.proto.

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

@$core.Deprecated('Use announcementDescriptor instead')
const Announcement$json = {
  '1': 'Announcement',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'type', '3': 4, '4': 1, '5': 9, '10': 'type'},
    {'1': 'title', '3': 5, '4': 1, '5': 9, '10': 'title'},
    {'1': 'content', '3': 6, '4': 1, '5': 9, '10': 'content'},
    {'1': 'image_url', '3': 7, '4': 1, '5': 9, '10': 'imageUrl'},
    {'1': 'link_url', '3': 8, '4': 1, '5': 9, '10': 'linkUrl'},
    {'1': 'publish_at', '3': 9, '4': 1, '5': 9, '10': 'publishAt'},
    {'1': 'unpublish_at', '3': 10, '4': 1, '5': 9, '10': 'unpublishAt'},
    {'1': 'sort_order', '3': 11, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 12, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'deploy_web', '3': 13, '4': 1, '5': 8, '10': 'deployWeb'},
    {'1': 'deploy_app', '3': 14, '4': 1, '5': 8, '10': 'deployApp'},
    {'1': 'created_at', '3': 15, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 16, '4': 1, '5': 9, '10': 'updatedAt'},
  ],
};

/// Descriptor for `Announcement`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List announcementDescriptor = $convert.base64Decode(
    'CgxBbm5vdW5jZW1lbnQSDgoCaWQYASABKAlSAmlkEh0KCmNvbXBhbnlfaWQYAiABKAlSCWNvbX'
    'BhbnlJZBIjCg1kZXBhcnRtZW50X2lkGAMgASgJUgxkZXBhcnRtZW50SWQSEgoEdHlwZRgEIAEo'
    'CVIEdHlwZRIUCgV0aXRsZRgFIAEoCVIFdGl0bGUSGAoHY29udGVudBgGIAEoCVIHY29udGVudB'
    'IbCglpbWFnZV91cmwYByABKAlSCGltYWdlVXJsEhkKCGxpbmtfdXJsGAggASgJUgdsaW5rVXJs'
    'Eh0KCnB1Ymxpc2hfYXQYCSABKAlSCXB1Ymxpc2hBdBIhCgx1bnB1Ymxpc2hfYXQYCiABKAlSC3'
    'VucHVibGlzaEF0Eh0KCnNvcnRfb3JkZXIYCyABKAVSCXNvcnRPcmRlchIbCglpc19hY3RpdmUY'
    'DCABKAhSCGlzQWN0aXZlEh0KCmRlcGxveV93ZWIYDSABKAhSCWRlcGxveVdlYhIdCgpkZXBsb3'
    'lfYXBwGA4gASgIUglkZXBsb3lBcHASHQoKY3JlYXRlZF9hdBgPIAEoCVIJY3JlYXRlZEF0Eh0K'
    'CnVwZGF0ZWRfYXQYECABKAlSCXVwZGF0ZWRBdA==');

@$core.Deprecated('Use listAnnouncementsRequestDescriptor instead')
const ListAnnouncementsRequest$json = {
  '1': 'ListAnnouncementsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'type', '3': 3, '4': 1, '5': 9, '10': 'type'},
    {'1': 'include_deleted', '3': 4, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListAnnouncementsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAnnouncementsRequestDescriptor = $convert.base64Decode(
    'ChhMaXN0QW5ub3VuY2VtZW50c1JlcXVlc3QSEgoEcGFnZRgBIAEoBVIEcGFnZRIbCglwYWdlX3'
    'NpemUYAiABKAVSCHBhZ2VTaXplEhIKBHR5cGUYAyABKAlSBHR5cGUSJwoPaW5jbHVkZV9kZWxl'
    'dGVkGAQgASgIUg5pbmNsdWRlRGVsZXRlZA==');

@$core.Deprecated('Use listAnnouncementsResponseDescriptor instead')
const ListAnnouncementsResponse$json = {
  '1': 'ListAnnouncementsResponse',
  '2': [
    {
      '1': 'announcements',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.Announcement',
      '10': 'announcements'
    },
    {'1': 'total', '3': 2, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListAnnouncementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAnnouncementsResponseDescriptor = $convert.base64Decode(
    'ChlMaXN0QW5ub3VuY2VtZW50c1Jlc3BvbnNlEkEKDWFubm91bmNlbWVudHMYASADKAsyGy5zYW'
    'xlc29yZGVyLnYxLkFubm91bmNlbWVudFINYW5ub3VuY2VtZW50cxIUCgV0b3RhbBgCIAEoBVIF'
    'dG90YWw=');

@$core.Deprecated('Use createAnnouncementRequestDescriptor instead')
const CreateAnnouncementRequest$json = {
  '1': 'CreateAnnouncementRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 2, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'type', '3': 3, '4': 1, '5': 9, '10': 'type'},
    {'1': 'title', '3': 4, '4': 1, '5': 9, '10': 'title'},
    {'1': 'content', '3': 5, '4': 1, '5': 9, '10': 'content'},
    {'1': 'image_url', '3': 6, '4': 1, '5': 9, '10': 'imageUrl'},
    {'1': 'link_url', '3': 7, '4': 1, '5': 9, '10': 'linkUrl'},
    {'1': 'publish_at', '3': 8, '4': 1, '5': 9, '10': 'publishAt'},
    {'1': 'unpublish_at', '3': 9, '4': 1, '5': 9, '10': 'unpublishAt'},
    {'1': 'sort_order', '3': 10, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 11, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'deploy_web', '3': 12, '4': 1, '5': 8, '10': 'deployWeb'},
    {'1': 'deploy_app', '3': 13, '4': 1, '5': 8, '10': 'deployApp'},
  ],
};

/// Descriptor for `CreateAnnouncementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAnnouncementRequestDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVBbm5vdW5jZW1lbnRSZXF1ZXN0Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbn'
    'lJZBIjCg1kZXBhcnRtZW50X2lkGAIgASgJUgxkZXBhcnRtZW50SWQSEgoEdHlwZRgDIAEoCVIE'
    'dHlwZRIUCgV0aXRsZRgEIAEoCVIFdGl0bGUSGAoHY29udGVudBgFIAEoCVIHY29udGVudBIbCg'
    'lpbWFnZV91cmwYBiABKAlSCGltYWdlVXJsEhkKCGxpbmtfdXJsGAcgASgJUgdsaW5rVXJsEh0K'
    'CnB1Ymxpc2hfYXQYCCABKAlSCXB1Ymxpc2hBdBIhCgx1bnB1Ymxpc2hfYXQYCSABKAlSC3VucH'
    'VibGlzaEF0Eh0KCnNvcnRfb3JkZXIYCiABKAVSCXNvcnRPcmRlchIbCglpc19hY3RpdmUYCyAB'
    'KAhSCGlzQWN0aXZlEh0KCmRlcGxveV93ZWIYDCABKAhSCWRlcGxveVdlYhIdCgpkZXBsb3lfYX'
    'BwGA0gASgIUglkZXBsb3lBcHA=');

@$core.Deprecated('Use createAnnouncementResponseDescriptor instead')
const CreateAnnouncementResponse$json = {
  '1': 'CreateAnnouncementResponse',
  '2': [
    {
      '1': 'announcement',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Announcement',
      '10': 'announcement'
    },
  ],
};

/// Descriptor for `CreateAnnouncementResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAnnouncementResponseDescriptor =
    $convert.base64Decode(
        'ChpDcmVhdGVBbm5vdW5jZW1lbnRSZXNwb25zZRI/Cgxhbm5vdW5jZW1lbnQYASABKAsyGy5zYW'
        'xlc29yZGVyLnYxLkFubm91bmNlbWVudFIMYW5ub3VuY2VtZW50');

@$core.Deprecated('Use updateAnnouncementRequestDescriptor instead')
const UpdateAnnouncementRequest$json = {
  '1': 'UpdateAnnouncementRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'type', '3': 4, '4': 1, '5': 9, '10': 'type'},
    {'1': 'title', '3': 5, '4': 1, '5': 9, '10': 'title'},
    {'1': 'content', '3': 6, '4': 1, '5': 9, '10': 'content'},
    {'1': 'image_url', '3': 7, '4': 1, '5': 9, '10': 'imageUrl'},
    {'1': 'link_url', '3': 8, '4': 1, '5': 9, '10': 'linkUrl'},
    {'1': 'publish_at', '3': 9, '4': 1, '5': 9, '10': 'publishAt'},
    {'1': 'unpublish_at', '3': 10, '4': 1, '5': 9, '10': 'unpublishAt'},
    {'1': 'sort_order', '3': 11, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'is_active', '3': 12, '4': 1, '5': 8, '10': 'isActive'},
    {'1': 'deploy_web', '3': 13, '4': 1, '5': 8, '10': 'deployWeb'},
    {'1': 'deploy_app', '3': 14, '4': 1, '5': 8, '10': 'deployApp'},
  ],
};

/// Descriptor for `UpdateAnnouncementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAnnouncementRequestDescriptor = $convert.base64Decode(
    'ChlVcGRhdGVBbm5vdW5jZW1lbnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBISCgR0eXBlGAQgAS'
    'gJUgR0eXBlEhQKBXRpdGxlGAUgASgJUgV0aXRsZRIYCgdjb250ZW50GAYgASgJUgdjb250ZW50'
    'EhsKCWltYWdlX3VybBgHIAEoCVIIaW1hZ2VVcmwSGQoIbGlua191cmwYCCABKAlSB2xpbmtVcm'
    'wSHQoKcHVibGlzaF9hdBgJIAEoCVIJcHVibGlzaEF0EiEKDHVucHVibGlzaF9hdBgKIAEoCVIL'
    'dW5wdWJsaXNoQXQSHQoKc29ydF9vcmRlchgLIAEoBVIJc29ydE9yZGVyEhsKCWlzX2FjdGl2ZR'
    'gMIAEoCFIIaXNBY3RpdmUSHQoKZGVwbG95X3dlYhgNIAEoCFIJZGVwbG95V2ViEh0KCmRlcGxv'
    'eV9hcHAYDiABKAhSCWRlcGxveUFwcA==');

@$core.Deprecated('Use updateAnnouncementResponseDescriptor instead')
const UpdateAnnouncementResponse$json = {
  '1': 'UpdateAnnouncementResponse',
  '2': [
    {
      '1': 'announcement',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Announcement',
      '10': 'announcement'
    },
  ],
};

/// Descriptor for `UpdateAnnouncementResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAnnouncementResponseDescriptor =
    $convert.base64Decode(
        'ChpVcGRhdGVBbm5vdW5jZW1lbnRSZXNwb25zZRI/Cgxhbm5vdW5jZW1lbnQYASABKAsyGy5zYW'
        'xlc29yZGVyLnYxLkFubm91bmNlbWVudFIMYW5ub3VuY2VtZW50');

@$core.Deprecated('Use deleteAnnouncementRequestDescriptor instead')
const DeleteAnnouncementRequest$json = {
  '1': 'DeleteAnnouncementRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteAnnouncementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAnnouncementRequestDescriptor =
    $convert.base64Decode(
        'ChlEZWxldGVBbm5vdW5jZW1lbnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use deleteAnnouncementResponseDescriptor instead')
const DeleteAnnouncementResponse$json = {
  '1': 'DeleteAnnouncementResponse',
};

/// Descriptor for `DeleteAnnouncementResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAnnouncementResponseDescriptor =
    $convert.base64Decode('ChpEZWxldGVBbm5vdW5jZW1lbnRSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> AnnouncementServiceBase$json = {
  '1': 'AnnouncementService',
  '2': [
    {
      '1': 'ListAnnouncements',
      '2': '.salesorder.v1.ListAnnouncementsRequest',
      '3': '.salesorder.v1.ListAnnouncementsResponse'
    },
    {
      '1': 'CreateAnnouncement',
      '2': '.salesorder.v1.CreateAnnouncementRequest',
      '3': '.salesorder.v1.CreateAnnouncementResponse'
    },
    {
      '1': 'UpdateAnnouncement',
      '2': '.salesorder.v1.UpdateAnnouncementRequest',
      '3': '.salesorder.v1.UpdateAnnouncementResponse'
    },
    {
      '1': 'DeleteAnnouncement',
      '2': '.salesorder.v1.DeleteAnnouncementRequest',
      '3': '.salesorder.v1.DeleteAnnouncementResponse'
    },
  ],
};

@$core.Deprecated('Use announcementServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AnnouncementServiceBase$messageJson = {
  '.salesorder.v1.ListAnnouncementsRequest': ListAnnouncementsRequest$json,
  '.salesorder.v1.ListAnnouncementsResponse': ListAnnouncementsResponse$json,
  '.salesorder.v1.Announcement': Announcement$json,
  '.salesorder.v1.CreateAnnouncementRequest': CreateAnnouncementRequest$json,
  '.salesorder.v1.CreateAnnouncementResponse': CreateAnnouncementResponse$json,
  '.salesorder.v1.UpdateAnnouncementRequest': UpdateAnnouncementRequest$json,
  '.salesorder.v1.UpdateAnnouncementResponse': UpdateAnnouncementResponse$json,
  '.salesorder.v1.DeleteAnnouncementRequest': DeleteAnnouncementRequest$json,
  '.salesorder.v1.DeleteAnnouncementResponse': DeleteAnnouncementResponse$json,
};

/// Descriptor for `AnnouncementService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List announcementServiceDescriptor = $convert.base64Decode(
    'ChNBbm5vdW5jZW1lbnRTZXJ2aWNlEmYKEUxpc3RBbm5vdW5jZW1lbnRzEicuc2FsZXNvcmRlci'
    '52MS5MaXN0QW5ub3VuY2VtZW50c1JlcXVlc3QaKC5zYWxlc29yZGVyLnYxLkxpc3RBbm5vdW5j'
    'ZW1lbnRzUmVzcG9uc2USaQoSQ3JlYXRlQW5ub3VuY2VtZW50Eiguc2FsZXNvcmRlci52MS5Dcm'
    'VhdGVBbm5vdW5jZW1lbnRSZXF1ZXN0Gikuc2FsZXNvcmRlci52MS5DcmVhdGVBbm5vdW5jZW1l'
    'bnRSZXNwb25zZRJpChJVcGRhdGVBbm5vdW5jZW1lbnQSKC5zYWxlc29yZGVyLnYxLlVwZGF0ZU'
    'Fubm91bmNlbWVudFJlcXVlc3QaKS5zYWxlc29yZGVyLnYxLlVwZGF0ZUFubm91bmNlbWVudFJl'
    'c3BvbnNlEmkKEkRlbGV0ZUFubm91bmNlbWVudBIoLnNhbGVzb3JkZXIudjEuRGVsZXRlQW5ub3'
    'VuY2VtZW50UmVxdWVzdBopLnNhbGVzb3JkZXIudjEuRGVsZXRlQW5ub3VuY2VtZW50UmVzcG9u'
    'c2U=');
