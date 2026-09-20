// This is a generated file - do not edit.
//
// Generated from platform/v1/platform.proto.

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

@$core.Deprecated('Use tenantSummaryDescriptor instead')
const TenantSummary$json = {
  '1': 'TenantSummary',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'company_name', '3': 2, '4': 1, '5': 9, '10': 'companyName'},
    {'1': 'plan_code', '3': 3, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'plan_name', '3': 4, '4': 1, '5': 9, '10': 'planName'},
    {
      '1': 'subscription_status',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'subscriptionStatus'
    },
    {'1': 'seat_count', '3': 6, '4': 1, '5': 5, '10': 'seatCount'},
    {
      '1': 'current_period_end',
      '3': 7,
      '4': 1,
      '5': 9,
      '10': 'currentPeriodEnd'
    },
    {'1': 'overdue', '3': 8, '4': 1, '5': 8, '10': 'overdue'},
  ],
};

/// Descriptor for `TenantSummary`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List tenantSummaryDescriptor = $convert.base64Decode(
    'Cg1UZW5hbnRTdW1tYXJ5Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbnlJZBIhCgxjb21wYW'
    '55X25hbWUYAiABKAlSC2NvbXBhbnlOYW1lEhsKCXBsYW5fY29kZRgDIAEoCVIIcGxhbkNvZGUS'
    'GwoJcGxhbl9uYW1lGAQgASgJUghwbGFuTmFtZRIvChNzdWJzY3JpcHRpb25fc3RhdHVzGAUgAS'
    'gJUhJzdWJzY3JpcHRpb25TdGF0dXMSHQoKc2VhdF9jb3VudBgGIAEoBVIJc2VhdENvdW50EiwK'
    'EmN1cnJlbnRfcGVyaW9kX2VuZBgHIAEoCVIQY3VycmVudFBlcmlvZEVuZBIYCgdvdmVyZHVlGA'
    'ggASgIUgdvdmVyZHVl');

@$core.Deprecated('Use platformPaginationDescriptor instead')
const PlatformPagination$json = {
  '1': 'PlatformPagination',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'total', '3': 3, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `PlatformPagination`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List platformPaginationDescriptor = $convert.base64Decode(
    'ChJQbGF0Zm9ybVBhZ2luYXRpb24SEgoEcGFnZRgBIAEoBVIEcGFnZRIbCglwYWdlX3NpemUYAi'
    'ABKAVSCHBhZ2VTaXplEhQKBXRvdGFsGAMgASgFUgV0b3RhbA==');

@$core.Deprecated('Use listTenantsRequestDescriptor instead')
const ListTenantsRequest$json = {
  '1': 'ListTenantsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'status', '3': 4, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `ListTenantsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTenantsRequestDescriptor = $convert.base64Decode(
    'ChJMaXN0VGVuYW50c1JlcXVlc3QSEgoEcGFnZRgBIAEoBVIEcGFnZRIbCglwYWdlX3NpemUYAi'
    'ABKAVSCHBhZ2VTaXplEhgKB2tleXdvcmQYAyABKAlSB2tleXdvcmQSFgoGc3RhdHVzGAQgASgJ'
    'UgZzdGF0dXM=');

@$core.Deprecated('Use listTenantsResponseDescriptor instead')
const ListTenantsResponse$json = {
  '1': 'ListTenantsResponse',
  '2': [
    {
      '1': 'tenants',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.TenantSummary',
      '10': 'tenants'
    },
    {
      '1': 'pagination',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.platform.v1.PlatformPagination',
      '10': 'pagination'
    },
  ],
};

/// Descriptor for `ListTenantsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTenantsResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0VGVuYW50c1Jlc3BvbnNlEjQKB3RlbmFudHMYASADKAsyGi5wbGF0Zm9ybS52MS5UZW'
    '5hbnRTdW1tYXJ5Ugd0ZW5hbnRzEj8KCnBhZ2luYXRpb24YAiABKAsyHy5wbGF0Zm9ybS52MS5Q'
    'bGF0Zm9ybVBhZ2luYXRpb25SCnBhZ2luYXRpb24=');

@$core.Deprecated('Use getTenantRequestDescriptor instead')
const GetTenantRequest$json = {
  '1': 'GetTenantRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
  ],
};

/// Descriptor for `GetTenantRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTenantRequestDescriptor = $convert.base64Decode(
    'ChBHZXRUZW5hbnRSZXF1ZXN0Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbnlJZA==');

@$core.Deprecated('Use getTenantResponseDescriptor instead')
const GetTenantResponse$json = {
  '1': 'GetTenantResponse',
  '2': [
    {
      '1': 'tenant',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.platform.v1.TenantSummary',
      '10': 'tenant'
    },
    {
      '1': 'overrides',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.TenantOverride',
      '10': 'overrides'
    },
  ],
};

/// Descriptor for `GetTenantResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTenantResponseDescriptor = $convert.base64Decode(
    'ChFHZXRUZW5hbnRSZXNwb25zZRIyCgZ0ZW5hbnQYASABKAsyGi5wbGF0Zm9ybS52MS5UZW5hbn'
    'RTdW1tYXJ5UgZ0ZW5hbnQSOQoJb3ZlcnJpZGVzGAIgAygLMhsucGxhdGZvcm0udjEuVGVuYW50'
    'T3ZlcnJpZGVSCW92ZXJyaWRlcw==');

@$core.Deprecated('Use tenantOverrideDescriptor instead')
const TenantOverride$json = {
  '1': 'TenantOverride',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'feature_code', '3': 2, '4': 1, '5': 9, '10': 'featureCode'},
    {'1': 'enabled_set', '3': 3, '4': 1, '5': 8, '10': 'enabledSet'},
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'limit_set', '3': 5, '4': 1, '5': 8, '10': 'limitSet'},
    {'1': 'limit_value', '3': 6, '4': 1, '5': 3, '10': 'limitValue'},
    {'1': 'reason', '3': 7, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'owner', '3': 8, '4': 1, '5': 9, '10': 'owner'},
    {'1': 'expires_at', '3': 9, '4': 1, '5': 9, '10': 'expiresAt'},
  ],
};

/// Descriptor for `TenantOverride`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List tenantOverrideDescriptor = $convert.base64Decode(
    'Cg5UZW5hbnRPdmVycmlkZRIOCgJpZBgBIAEoCVICaWQSIQoMZmVhdHVyZV9jb2RlGAIgASgJUg'
    'tmZWF0dXJlQ29kZRIfCgtlbmFibGVkX3NldBgDIAEoCFIKZW5hYmxlZFNldBIYCgdlbmFibGVk'
    'GAQgASgIUgdlbmFibGVkEhsKCWxpbWl0X3NldBgFIAEoCFIIbGltaXRTZXQSHwoLbGltaXRfdm'
    'FsdWUYBiABKANSCmxpbWl0VmFsdWUSFgoGcmVhc29uGAcgASgJUgZyZWFzb24SFAoFb3duZXIY'
    'CCABKAlSBW93bmVyEh0KCmV4cGlyZXNfYXQYCSABKAlSCWV4cGlyZXNBdA==');

@$core.Deprecated('Use listPlansRequestDescriptor instead')
const ListPlansRequest$json = {
  '1': 'ListPlansRequest',
};

/// Descriptor for `ListPlansRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listPlansRequestDescriptor =
    $convert.base64Decode('ChBMaXN0UGxhbnNSZXF1ZXN0');

@$core.Deprecated('Use listPlansResponseDescriptor instead')
const ListPlansResponse$json = {
  '1': 'ListPlansResponse',
  '2': [
    {
      '1': 'plans',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.Plan',
      '10': 'plans'
    },
  ],
};

/// Descriptor for `ListPlansResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listPlansResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0UGxhbnNSZXNwb25zZRInCgVwbGFucxgBIAMoCzIRLnBsYXRmb3JtLnYxLlBsYW5SBX'
    'BsYW5z');

@$core.Deprecated('Use planDescriptor instead')
const Plan$json = {
  '1': 'Plan',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'status', '3': 4, '4': 1, '5': 9, '10': 'status'},
    {'1': 'sort_order', '3': 5, '4': 1, '5': 5, '10': 'sortOrder'},
    {
      '1': 'prices',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.PlanPrice',
      '10': 'prices'
    },
  ],
};

/// Descriptor for `Plan`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List planDescriptor = $convert.base64Decode(
    'CgRQbGFuEg4KAmlkGAEgASgJUgJpZBISCgRjb2RlGAIgASgJUgRjb2RlEhIKBG5hbWUYAyABKA'
    'lSBG5hbWUSFgoGc3RhdHVzGAQgASgJUgZzdGF0dXMSHQoKc29ydF9vcmRlchgFIAEoBVIJc29y'
    'dE9yZGVyEi4KBnByaWNlcxgGIAMoCzIWLnBsYXRmb3JtLnYxLlBsYW5QcmljZVIGcHJpY2Vz');

@$core.Deprecated('Use planPriceDescriptor instead')
const PlanPrice$json = {
  '1': 'PlanPrice',
  '2': [
    {'1': 'billing_cycle', '3': 1, '4': 1, '5': 9, '10': 'billingCycle'},
    {'1': 'base_price', '3': 2, '4': 1, '5': 9, '10': 'basePrice'},
    {'1': 'seat_price', '3': 3, '4': 1, '5': 9, '10': 'seatPrice'},
    {'1': 'currency', '3': 4, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'effective_from', '3': 5, '4': 1, '5': 9, '10': 'effectiveFrom'},
  ],
};

/// Descriptor for `PlanPrice`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List planPriceDescriptor = $convert.base64Decode(
    'CglQbGFuUHJpY2USIwoNYmlsbGluZ19jeWNsZRgBIAEoCVIMYmlsbGluZ0N5Y2xlEh0KCmJhc2'
    'VfcHJpY2UYAiABKAlSCWJhc2VQcmljZRIdCgpzZWF0X3ByaWNlGAMgASgJUglzZWF0UHJpY2US'
    'GgoIY3VycmVuY3kYBCABKAlSCGN1cnJlbmN5EiUKDmVmZmVjdGl2ZV9mcm9tGAUgASgJUg1lZm'
    'ZlY3RpdmVGcm9t');

@$core.Deprecated('Use getPlanEntitlementsRequestDescriptor instead')
const GetPlanEntitlementsRequest$json = {
  '1': 'GetPlanEntitlementsRequest',
  '2': [
    {'1': 'plan_code', '3': 1, '4': 1, '5': 9, '10': 'planCode'},
  ],
};

/// Descriptor for `GetPlanEntitlementsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPlanEntitlementsRequestDescriptor =
    $convert.base64Decode(
        'ChpHZXRQbGFuRW50aXRsZW1lbnRzUmVxdWVzdBIbCglwbGFuX2NvZGUYASABKAlSCHBsYW5Db2'
        'Rl');

@$core.Deprecated('Use getPlanEntitlementsResponseDescriptor instead')
const GetPlanEntitlementsResponse$json = {
  '1': 'GetPlanEntitlementsResponse',
  '2': [
    {
      '1': 'entitlements',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.FeatureEntitlement',
      '10': 'entitlements'
    },
    {
      '1': 'features',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.Feature',
      '10': 'features'
    },
  ],
};

/// Descriptor for `GetPlanEntitlementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPlanEntitlementsResponseDescriptor =
    $convert.base64Decode(
        'ChtHZXRQbGFuRW50aXRsZW1lbnRzUmVzcG9uc2USQwoMZW50aXRsZW1lbnRzGAEgAygLMh8ucG'
        'xhdGZvcm0udjEuRmVhdHVyZUVudGl0bGVtZW50UgxlbnRpdGxlbWVudHMSMAoIZmVhdHVyZXMY'
        'AiADKAsyFC5wbGF0Zm9ybS52MS5GZWF0dXJlUghmZWF0dXJlcw==');

@$core.Deprecated('Use featureDescriptor instead')
const Feature$json = {
  '1': 'Feature',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'unit', '3': 3, '4': 1, '5': 9, '10': 'unit'},
    {'1': 'description', '3': 4, '4': 1, '5': 9, '10': 'description'},
  ],
};

/// Descriptor for `Feature`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List featureDescriptor = $convert.base64Decode(
    'CgdGZWF0dXJlEhIKBGNvZGUYASABKAlSBGNvZGUSEgoEdHlwZRgCIAEoCVIEdHlwZRISCgR1bm'
    'l0GAMgASgJUgR1bml0EiAKC2Rlc2NyaXB0aW9uGAQgASgJUgtkZXNjcmlwdGlvbg==');

@$core.Deprecated('Use featureEntitlementDescriptor instead')
const FeatureEntitlement$json = {
  '1': 'FeatureEntitlement',
  '2': [
    {'1': 'feature_code', '3': 1, '4': 1, '5': 9, '10': 'featureCode'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'limit_set', '3': 3, '4': 1, '5': 8, '10': 'limitSet'},
    {'1': 'limit_value', '3': 4, '4': 1, '5': 3, '10': 'limitValue'},
  ],
};

/// Descriptor for `FeatureEntitlement`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List featureEntitlementDescriptor = $convert.base64Decode(
    'ChJGZWF0dXJlRW50aXRsZW1lbnQSIQoMZmVhdHVyZV9jb2RlGAEgASgJUgtmZWF0dXJlQ29kZR'
    'IYCgdlbmFibGVkGAIgASgIUgdlbmFibGVkEhsKCWxpbWl0X3NldBgDIAEoCFIIbGltaXRTZXQS'
    'HwoLbGltaXRfdmFsdWUYBCABKANSCmxpbWl0VmFsdWU=');

@$core.Deprecated('Use listPlatformAuditRequestDescriptor instead')
const ListPlatformAuditRequest$json = {
  '1': 'ListPlatformAuditRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'target_type', '3': 3, '4': 1, '5': 9, '10': 'targetType'},
    {'1': 'target_id', '3': 4, '4': 1, '5': 9, '10': 'targetId'},
  ],
};

/// Descriptor for `ListPlatformAuditRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listPlatformAuditRequestDescriptor = $convert.base64Decode(
    'ChhMaXN0UGxhdGZvcm1BdWRpdFJlcXVlc3QSEgoEcGFnZRgBIAEoBVIEcGFnZRIbCglwYWdlX3'
    'NpemUYAiABKAVSCHBhZ2VTaXplEh8KC3RhcmdldF90eXBlGAMgASgJUgp0YXJnZXRUeXBlEhsK'
    'CXRhcmdldF9pZBgEIAEoCVIIdGFyZ2V0SWQ=');

@$core.Deprecated('Use listPlatformAuditResponseDescriptor instead')
const ListPlatformAuditResponse$json = {
  '1': 'ListPlatformAuditResponse',
  '2': [
    {
      '1': 'entries',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.PlatformAuditEntry',
      '10': 'entries'
    },
    {
      '1': 'pagination',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.platform.v1.PlatformPagination',
      '10': 'pagination'
    },
  ],
};

/// Descriptor for `ListPlatformAuditResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listPlatformAuditResponseDescriptor = $convert.base64Decode(
    'ChlMaXN0UGxhdGZvcm1BdWRpdFJlc3BvbnNlEjkKB2VudHJpZXMYASADKAsyHy5wbGF0Zm9ybS'
    '52MS5QbGF0Zm9ybUF1ZGl0RW50cnlSB2VudHJpZXMSPwoKcGFnaW5hdGlvbhgCIAEoCzIfLnBs'
    'YXRmb3JtLnYxLlBsYXRmb3JtUGFnaW5hdGlvblIKcGFnaW5hdGlvbg==');

@$core.Deprecated('Use platformAuditEntryDescriptor instead')
const PlatformAuditEntry$json = {
  '1': 'PlatformAuditEntry',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'operator_email', '3': 2, '4': 1, '5': 9, '10': 'operatorEmail'},
    {'1': 'action', '3': 3, '4': 1, '5': 9, '10': 'action'},
    {'1': 'target_type', '3': 4, '4': 1, '5': 9, '10': 'targetType'},
    {'1': 'target_id', '3': 5, '4': 1, '5': 9, '10': 'targetId'},
    {'1': 'reason', '3': 6, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'created_at', '3': 7, '4': 1, '5': 9, '10': 'createdAt'},
  ],
};

/// Descriptor for `PlatformAuditEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List platformAuditEntryDescriptor = $convert.base64Decode(
    'ChJQbGF0Zm9ybUF1ZGl0RW50cnkSDgoCaWQYASABKAlSAmlkEiUKDm9wZXJhdG9yX2VtYWlsGA'
    'IgASgJUg1vcGVyYXRvckVtYWlsEhYKBmFjdGlvbhgDIAEoCVIGYWN0aW9uEh8KC3RhcmdldF90'
    'eXBlGAQgASgJUgp0YXJnZXRUeXBlEhsKCXRhcmdldF9pZBgFIAEoCVIIdGFyZ2V0SWQSFgoGcm'
    'Vhc29uGAYgASgJUgZyZWFzb24SHQoKY3JlYXRlZF9hdBgHIAEoCVIJY3JlYXRlZEF0');

@$core.Deprecated('Use getTenantEntitlementsRequestDescriptor instead')
const GetTenantEntitlementsRequest$json = {
  '1': 'GetTenantEntitlementsRequest',
};

/// Descriptor for `GetTenantEntitlementsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTenantEntitlementsRequestDescriptor =
    $convert.base64Decode('ChxHZXRUZW5hbnRFbnRpdGxlbWVudHNSZXF1ZXN0');

@$core.Deprecated('Use getTenantEntitlementsResponseDescriptor instead')
const GetTenantEntitlementsResponse$json = {
  '1': 'GetTenantEntitlementsResponse',
  '2': [
    {'1': 'plan_code', '3': 1, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'plan_name', '3': 2, '4': 1, '5': 9, '10': 'planName'},
    {'1': 'status', '3': 3, '4': 1, '5': 9, '10': 'status'},
    {'1': 'trial_ends_at', '3': 4, '4': 1, '5': 9, '10': 'trialEndsAt'},
    {
      '1': 'usage',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.Usage',
      '10': 'usage'
    },
  ],
};

/// Descriptor for `GetTenantEntitlementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTenantEntitlementsResponseDescriptor = $convert.base64Decode(
    'Ch1HZXRUZW5hbnRFbnRpdGxlbWVudHNSZXNwb25zZRIbCglwbGFuX2NvZGUYASABKAlSCHBsYW'
    '5Db2RlEhsKCXBsYW5fbmFtZRgCIAEoCVIIcGxhbk5hbWUSFgoGc3RhdHVzGAMgASgJUgZzdGF0'
    'dXMSIgoNdHJpYWxfZW5kc19hdBgEIAEoCVILdHJpYWxFbmRzQXQSKAoFdXNhZ2UYBSADKAsyEi'
    '5wbGF0Zm9ybS52MS5Vc2FnZVIFdXNhZ2U=');

@$core.Deprecated('Use usageDescriptor instead')
const Usage$json = {
  '1': 'Usage',
  '2': [
    {'1': 'feature_code', '3': 1, '4': 1, '5': 9, '10': 'featureCode'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'limit_set', '3': 3, '4': 1, '5': 8, '10': 'limitSet'},
    {'1': 'limit_value', '3': 4, '4': 1, '5': 3, '10': 'limitValue'},
    {'1': 'used', '3': 5, '4': 1, '5': 3, '10': 'used'},
  ],
};

/// Descriptor for `Usage`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List usageDescriptor = $convert.base64Decode(
    'CgVVc2FnZRIhCgxmZWF0dXJlX2NvZGUYASABKAlSC2ZlYXR1cmVDb2RlEhgKB2VuYWJsZWQYAi'
    'ABKAhSB2VuYWJsZWQSGwoJbGltaXRfc2V0GAMgASgIUghsaW1pdFNldBIfCgtsaW1pdF92YWx1'
    'ZRgEIAEoA1IKbGltaXRWYWx1ZRISCgR1c2VkGAUgASgDUgR1c2Vk');

const $core.Map<$core.String, $core.dynamic> PlatformAdminServiceBase$json = {
  '1': 'PlatformAdminService',
  '2': [
    {
      '1': 'ListTenants',
      '2': '.platform.v1.ListTenantsRequest',
      '3': '.platform.v1.ListTenantsResponse'
    },
    {
      '1': 'GetTenant',
      '2': '.platform.v1.GetTenantRequest',
      '3': '.platform.v1.GetTenantResponse'
    },
    {
      '1': 'ListPlans',
      '2': '.platform.v1.ListPlansRequest',
      '3': '.platform.v1.ListPlansResponse'
    },
    {
      '1': 'GetPlanEntitlements',
      '2': '.platform.v1.GetPlanEntitlementsRequest',
      '3': '.platform.v1.GetPlanEntitlementsResponse'
    },
    {
      '1': 'ListPlatformAudit',
      '2': '.platform.v1.ListPlatformAuditRequest',
      '3': '.platform.v1.ListPlatformAuditResponse'
    },
  ],
};

@$core.Deprecated('Use platformAdminServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    PlatformAdminServiceBase$messageJson = {
  '.platform.v1.ListTenantsRequest': ListTenantsRequest$json,
  '.platform.v1.ListTenantsResponse': ListTenantsResponse$json,
  '.platform.v1.TenantSummary': TenantSummary$json,
  '.platform.v1.PlatformPagination': PlatformPagination$json,
  '.platform.v1.GetTenantRequest': GetTenantRequest$json,
  '.platform.v1.GetTenantResponse': GetTenantResponse$json,
  '.platform.v1.TenantOverride': TenantOverride$json,
  '.platform.v1.ListPlansRequest': ListPlansRequest$json,
  '.platform.v1.ListPlansResponse': ListPlansResponse$json,
  '.platform.v1.Plan': Plan$json,
  '.platform.v1.PlanPrice': PlanPrice$json,
  '.platform.v1.GetPlanEntitlementsRequest': GetPlanEntitlementsRequest$json,
  '.platform.v1.GetPlanEntitlementsResponse': GetPlanEntitlementsResponse$json,
  '.platform.v1.FeatureEntitlement': FeatureEntitlement$json,
  '.platform.v1.Feature': Feature$json,
  '.platform.v1.ListPlatformAuditRequest': ListPlatformAuditRequest$json,
  '.platform.v1.ListPlatformAuditResponse': ListPlatformAuditResponse$json,
  '.platform.v1.PlatformAuditEntry': PlatformAuditEntry$json,
};

/// Descriptor for `PlatformAdminService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List platformAdminServiceDescriptor = $convert.base64Decode(
    'ChRQbGF0Zm9ybUFkbWluU2VydmljZRJQCgtMaXN0VGVuYW50cxIfLnBsYXRmb3JtLnYxLkxpc3'
    'RUZW5hbnRzUmVxdWVzdBogLnBsYXRmb3JtLnYxLkxpc3RUZW5hbnRzUmVzcG9uc2USSgoJR2V0'
    'VGVuYW50Eh0ucGxhdGZvcm0udjEuR2V0VGVuYW50UmVxdWVzdBoeLnBsYXRmb3JtLnYxLkdldF'
    'RlbmFudFJlc3BvbnNlEkoKCUxpc3RQbGFucxIdLnBsYXRmb3JtLnYxLkxpc3RQbGFuc1JlcXVl'
    'c3QaHi5wbGF0Zm9ybS52MS5MaXN0UGxhbnNSZXNwb25zZRJoChNHZXRQbGFuRW50aXRsZW1lbn'
    'RzEicucGxhdGZvcm0udjEuR2V0UGxhbkVudGl0bGVtZW50c1JlcXVlc3QaKC5wbGF0Zm9ybS52'
    'MS5HZXRQbGFuRW50aXRsZW1lbnRzUmVzcG9uc2USYgoRTGlzdFBsYXRmb3JtQXVkaXQSJS5wbG'
    'F0Zm9ybS52MS5MaXN0UGxhdGZvcm1BdWRpdFJlcXVlc3QaJi5wbGF0Zm9ybS52MS5MaXN0UGxh'
    'dGZvcm1BdWRpdFJlc3BvbnNl');

const $core.Map<$core.String, $core.dynamic> TenantEntitlementServiceBase$json =
    {
  '1': 'TenantEntitlementService',
  '2': [
    {
      '1': 'GetTenantEntitlements',
      '2': '.platform.v1.GetTenantEntitlementsRequest',
      '3': '.platform.v1.GetTenantEntitlementsResponse'
    },
  ],
};

@$core.Deprecated('Use tenantEntitlementServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TenantEntitlementServiceBase$messageJson = {
  '.platform.v1.GetTenantEntitlementsRequest':
      GetTenantEntitlementsRequest$json,
  '.platform.v1.GetTenantEntitlementsResponse':
      GetTenantEntitlementsResponse$json,
  '.platform.v1.Usage': Usage$json,
};

/// Descriptor for `TenantEntitlementService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List tenantEntitlementServiceDescriptor = $convert.base64Decode(
    'ChhUZW5hbnRFbnRpdGxlbWVudFNlcnZpY2USbgoVR2V0VGVuYW50RW50aXRsZW1lbnRzEikucG'
    'xhdGZvcm0udjEuR2V0VGVuYW50RW50aXRsZW1lbnRzUmVxdWVzdBoqLnBsYXRmb3JtLnYxLkdl'
    'dFRlbmFudEVudGl0bGVtZW50c1Jlc3BvbnNl');
