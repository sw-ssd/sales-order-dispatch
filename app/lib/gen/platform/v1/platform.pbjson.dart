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
    {'1': 'trial_ends_at', '3': 10, '4': 1, '5': 9, '10': 'trialEndsAt'},
    {'1': 'grace_until', '3': 11, '4': 1, '5': 9, '10': 'graceUntil'},
  ],
};

/// Descriptor for `TenantSummary`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List tenantSummaryDescriptor = $convert.base64Decode(
    'Cg1UZW5hbnRTdW1tYXJ5Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbnlJZBIhCgxjb21wYW'
    '55X25hbWUYAiABKAlSC2NvbXBhbnlOYW1lEhsKCXBsYW5fY29kZRgDIAEoCVIIcGxhbkNvZGUS'
    'GwoJcGxhbl9uYW1lGAQgASgJUghwbGFuTmFtZRIvChNzdWJzY3JpcHRpb25fc3RhdHVzGAUgAS'
    'gJUhJzdWJzY3JpcHRpb25TdGF0dXMSHQoKc2VhdF9jb3VudBgGIAEoBVIJc2VhdENvdW50EiwK'
    'EmN1cnJlbnRfcGVyaW9kX2VuZBgHIAEoCVIQY3VycmVudFBlcmlvZEVuZBIYCgdvdmVyZHVlGA'
    'ggASgIUgdvdmVyZHVlEiIKDXRyaWFsX2VuZHNfYXQYCiABKAlSC3RyaWFsRW5kc0F0Eh8KC2dy'
    'YWNlX3VudGlsGAsgASgJUgpncmFjZVVudGls');

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

@$core.Deprecated('Use listReceivablesRequestDescriptor instead')
const ListReceivablesRequest$json = {
  '1': 'ListReceivablesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListReceivablesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listReceivablesRequestDescriptor =
    $convert.base64Decode(
        'ChZMaXN0UmVjZWl2YWJsZXNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaX'
        'plGAIgASgFUghwYWdlU2l6ZQ==');

@$core.Deprecated('Use listReceivablesResponseDescriptor instead')
const ListReceivablesResponse$json = {
  '1': 'ListReceivablesResponse',
  '2': [
    {
      '1': 'rows',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.Receivable',
      '10': 'rows'
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

/// Descriptor for `ListReceivablesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listReceivablesResponseDescriptor = $convert.base64Decode(
    'ChdMaXN0UmVjZWl2YWJsZXNSZXNwb25zZRIrCgRyb3dzGAEgAygLMhcucGxhdGZvcm0udjEuUm'
    'VjZWl2YWJsZVIEcm93cxI/CgpwYWdpbmF0aW9uGAIgASgLMh8ucGxhdGZvcm0udjEuUGxhdGZv'
    'cm1QYWdpbmF0aW9uUgpwYWdpbmF0aW9u');

@$core.Deprecated('Use receivableDescriptor instead')
const Receivable$json = {
  '1': 'Receivable',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'company_name', '3': 2, '4': 1, '5': 9, '10': 'companyName'},
    {'1': 'plan_code', '3': 3, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'period_no', '3': 4, '4': 1, '5': 5, '10': 'periodNo'},
    {'1': 'amount', '3': 5, '4': 1, '5': 9, '10': 'amount'},
    {'1': 'period_end', '3': 6, '4': 1, '5': 9, '10': 'periodEnd'},
    {'1': 'status', '3': 7, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `Receivable`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List receivableDescriptor = $convert.base64Decode(
    'CgpSZWNlaXZhYmxlEh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbnlJZBIhCgxjb21wYW55X2'
    '5hbWUYAiABKAlSC2NvbXBhbnlOYW1lEhsKCXBsYW5fY29kZRgDIAEoCVIIcGxhbkNvZGUSGwoJ'
    'cGVyaW9kX25vGAQgASgFUghwZXJpb2RObxIWCgZhbW91bnQYBSABKAlSBmFtb3VudBIdCgpwZX'
    'Jpb2RfZW5kGAYgASgJUglwZXJpb2RFbmQSFgoGc3RhdHVzGAcgASgJUgZzdGF0dXM=');

@$core.Deprecated('Use recordPaymentRequestDescriptor instead')
const RecordPaymentRequest$json = {
  '1': 'RecordPaymentRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'period_no', '3': 2, '4': 1, '5': 5, '10': 'periodNo'},
    {'1': 'amount', '3': 3, '4': 1, '5': 9, '10': 'amount'},
    {'1': 'provider', '3': 4, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'external_ref', '3': 5, '4': 1, '5': 9, '10': 'externalRef'},
    {'1': 'invoice_no', '3': 6, '4': 1, '5': 9, '10': 'invoiceNo'},
    {'1': 'invoice_status', '3': 7, '4': 1, '5': 9, '10': 'invoiceStatus'},
    {'1': 'buyer_tax_id', '3': 8, '4': 1, '5': 9, '10': 'buyerTaxId'},
    {'1': 'carrier', '3': 9, '4': 1, '5': 9, '10': 'carrier'},
    {'1': 'note', '3': 10, '4': 1, '5': 9, '10': 'note'},
    {'1': 'reason', '3': 11, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `RecordPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recordPaymentRequestDescriptor = $convert.base64Decode(
    'ChRSZWNvcmRQYXltZW50UmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SWQSGw'
    'oJcGVyaW9kX25vGAIgASgFUghwZXJpb2RObxIWCgZhbW91bnQYAyABKAlSBmFtb3VudBIaCghw'
    'cm92aWRlchgEIAEoCVIIcHJvdmlkZXISIQoMZXh0ZXJuYWxfcmVmGAUgASgJUgtleHRlcm5hbF'
    'JlZhIdCgppbnZvaWNlX25vGAYgASgJUglpbnZvaWNlTm8SJQoOaW52b2ljZV9zdGF0dXMYByAB'
    'KAlSDWludm9pY2VTdGF0dXMSIAoMYnV5ZXJfdGF4X2lkGAggASgJUgpidXllclRheElkEhgKB2'
    'NhcnJpZXIYCSABKAlSB2NhcnJpZXISEgoEbm90ZRgKIAEoCVIEbm90ZRIWCgZyZWFzb24YCyAB'
    'KAlSBnJlYXNvbg==');

@$core.Deprecated('Use recordPaymentResponseDescriptor instead')
const RecordPaymentResponse$json = {
  '1': 'RecordPaymentResponse',
  '2': [
    {'1': 'period_no', '3': 1, '4': 1, '5': 5, '10': 'periodNo'},
    {'1': 'status', '3': 2, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `RecordPaymentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recordPaymentResponseDescriptor = $convert.base64Decode(
    'ChVSZWNvcmRQYXltZW50UmVzcG9uc2USGwoJcGVyaW9kX25vGAEgASgFUghwZXJpb2RObxIWCg'
    'ZzdGF0dXMYAiABKAlSBnN0YXR1cw==');

@$core.Deprecated('Use createSubscriptionRequestDescriptor instead')
const CreateSubscriptionRequest$json = {
  '1': 'CreateSubscriptionRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'plan_code', '3': 2, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'billing_cycle', '3': 3, '4': 1, '5': 9, '10': 'billingCycle'},
    {'1': 'seat_count', '3': 4, '4': 1, '5': 5, '10': 'seatCount'},
    {'1': 'trial_ends_at', '3': 5, '4': 1, '5': 9, '10': 'trialEndsAt'},
    {'1': 'reason', '3': 6, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `CreateSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSubscriptionRequestDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVTdWJzY3JpcHRpb25SZXF1ZXN0Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbn'
    'lJZBIbCglwbGFuX2NvZGUYAiABKAlSCHBsYW5Db2RlEiMKDWJpbGxpbmdfY3ljbGUYAyABKAlS'
    'DGJpbGxpbmdDeWNsZRIdCgpzZWF0X2NvdW50GAQgASgFUglzZWF0Q291bnQSIgoNdHJpYWxfZW'
    '5kc19hdBgFIAEoCVILdHJpYWxFbmRzQXQSFgoGcmVhc29uGAYgASgJUgZyZWFzb24=');

@$core.Deprecated('Use createSubscriptionResponseDescriptor instead')
const CreateSubscriptionResponse$json = {
  '1': 'CreateSubscriptionResponse',
  '2': [
    {'1': 'subscription_id', '3': 1, '4': 1, '5': 9, '10': 'subscriptionId'},
    {'1': 'status', '3': 2, '4': 1, '5': 9, '10': 'status'},
    {'1': 'plan_code', '3': 3, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'billing_cycle', '3': 4, '4': 1, '5': 9, '10': 'billingCycle'},
    {'1': 'seat_count', '3': 5, '4': 1, '5': 5, '10': 'seatCount'},
    {'1': 'trial_ends_at', '3': 6, '4': 1, '5': 9, '10': 'trialEndsAt'},
    {'1': 'first_period_no', '3': 7, '4': 1, '5': 5, '10': 'firstPeriodNo'},
    {'1': 'first_period_end', '3': 8, '4': 1, '5': 9, '10': 'firstPeriodEnd'},
    {
      '1': 'first_period_amount',
      '3': 9,
      '4': 1,
      '5': 9,
      '10': 'firstPeriodAmount'
    },
  ],
};

/// Descriptor for `CreateSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSubscriptionResponseDescriptor = $convert.base64Decode(
    'ChpDcmVhdGVTdWJzY3JpcHRpb25SZXNwb25zZRInCg9zdWJzY3JpcHRpb25faWQYASABKAlSDn'
    'N1YnNjcmlwdGlvbklkEhYKBnN0YXR1cxgCIAEoCVIGc3RhdHVzEhsKCXBsYW5fY29kZRgDIAEo'
    'CVIIcGxhbkNvZGUSIwoNYmlsbGluZ19jeWNsZRgEIAEoCVIMYmlsbGluZ0N5Y2xlEh0KCnNlYX'
    'RfY291bnQYBSABKAVSCXNlYXRDb3VudBIiCg10cmlhbF9lbmRzX2F0GAYgASgJUgt0cmlhbEVu'
    'ZHNBdBImCg9maXJzdF9wZXJpb2Rfbm8YByABKAVSDWZpcnN0UGVyaW9kTm8SKAoQZmlyc3RfcG'
    'VyaW9kX2VuZBgIIAEoCVIOZmlyc3RQZXJpb2RFbmQSLgoTZmlyc3RfcGVyaW9kX2Ftb3VudBgJ'
    'IAEoCVIRZmlyc3RQZXJpb2RBbW91bnQ=');

@$core.Deprecated('Use setSeatCountRequestDescriptor instead')
const SetSeatCountRequest$json = {
  '1': 'SetSeatCountRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'seat_count', '3': 2, '4': 1, '5': 5, '10': 'seatCount'},
    {'1': 'reason', '3': 3, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `SetSeatCountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSeatCountRequestDescriptor = $convert.base64Decode(
    'ChNTZXRTZWF0Q291bnRSZXF1ZXN0Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbnlJZBIdCg'
    'pzZWF0X2NvdW50GAIgASgFUglzZWF0Q291bnQSFgoGcmVhc29uGAMgASgJUgZyZWFzb24=');

@$core.Deprecated('Use setSeatCountResponseDescriptor instead')
const SetSeatCountResponse$json = {
  '1': 'SetSeatCountResponse',
  '2': [
    {'1': 'seat_count', '3': 1, '4': 1, '5': 5, '10': 'seatCount'},
  ],
};

/// Descriptor for `SetSeatCountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSeatCountResponseDescriptor = $convert.base64Decode(
    'ChRTZXRTZWF0Q291bnRSZXNwb25zZRIdCgpzZWF0X2NvdW50GAEgASgFUglzZWF0Q291bnQ=');

@$core.Deprecated('Use changePlanRequestDescriptor instead')
const ChangePlanRequest$json = {
  '1': 'ChangePlanRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'plan_code', '3': 2, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'reason', '3': 3, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `ChangePlanRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePlanRequestDescriptor = $convert.base64Decode(
    'ChFDaGFuZ2VQbGFuUmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SWQSGwoJcG'
    'xhbl9jb2RlGAIgASgJUghwbGFuQ29kZRIWCgZyZWFzb24YAyABKAlSBnJlYXNvbg==');

@$core.Deprecated('Use changePlanResponseDescriptor instead')
const ChangePlanResponse$json = {
  '1': 'ChangePlanResponse',
  '2': [
    {'1': 'plan_code', '3': 1, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'effective_from', '3': 2, '4': 1, '5': 9, '10': 'effectiveFrom'},
  ],
};

/// Descriptor for `ChangePlanResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePlanResponseDescriptor = $convert.base64Decode(
    'ChJDaGFuZ2VQbGFuUmVzcG9uc2USGwoJcGxhbl9jb2RlGAEgASgJUghwbGFuQ29kZRIlCg5lZm'
    'ZlY3RpdmVfZnJvbRgCIAEoCVINZWZmZWN0aXZlRnJvbQ==');

@$core.Deprecated('Use cancelSubscriptionRequestDescriptor instead')
const CancelSubscriptionRequest$json = {
  '1': 'CancelSubscriptionRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'at_period_end', '3': 2, '4': 1, '5': 8, '10': 'atPeriodEnd'},
    {'1': 'reason', '3': 3, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `CancelSubscriptionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelSubscriptionRequestDescriptor = $convert.base64Decode(
    'ChlDYW5jZWxTdWJzY3JpcHRpb25SZXF1ZXN0Eh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbXBhbn'
    'lJZBIiCg1hdF9wZXJpb2RfZW5kGAIgASgIUgthdFBlcmlvZEVuZBIWCgZyZWFzb24YAyABKAlS'
    'BnJlYXNvbg==');

@$core.Deprecated('Use cancelSubscriptionResponseDescriptor instead')
const CancelSubscriptionResponse$json = {
  '1': 'CancelSubscriptionResponse',
  '2': [
    {'1': 'cancelled_at', '3': 1, '4': 1, '5': 9, '10': 'cancelledAt'},
    {'1': 'service_until', '3': 2, '4': 1, '5': 9, '10': 'serviceUntil'},
  ],
};

/// Descriptor for `CancelSubscriptionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelSubscriptionResponseDescriptor =
    $convert.base64Decode(
        'ChpDYW5jZWxTdWJzY3JpcHRpb25SZXNwb25zZRIhCgxjYW5jZWxsZWRfYXQYASABKAlSC2Nhbm'
        'NlbGxlZEF0EiMKDXNlcnZpY2VfdW50aWwYAiABKAlSDHNlcnZpY2VVbnRpbA==');

@$core.Deprecated('Use getBillingSettingsRequestDescriptor instead')
const GetBillingSettingsRequest$json = {
  '1': 'GetBillingSettingsRequest',
};

/// Descriptor for `GetBillingSettingsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBillingSettingsRequestDescriptor =
    $convert.base64Decode('ChlHZXRCaWxsaW5nU2V0dGluZ3NSZXF1ZXN0');

@$core.Deprecated('Use getBillingSettingsResponseDescriptor instead')
const GetBillingSettingsResponse$json = {
  '1': 'GetBillingSettingsResponse',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.BillingSetting',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `GetBillingSettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBillingSettingsResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRCaWxsaW5nU2V0dGluZ3NSZXNwb25zZRI3CghzZXR0aW5ncxgBIAMoCzIbLnBsYXRmb3'
        'JtLnYxLkJpbGxpbmdTZXR0aW5nUghzZXR0aW5ncw==');

@$core.Deprecated('Use billingSettingDescriptor instead')
const BillingSetting$json = {
  '1': 'BillingSetting',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
  ],
};

/// Descriptor for `BillingSetting`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List billingSettingDescriptor = $convert.base64Decode(
    'Cg5CaWxsaW5nU2V0dGluZxIQCgNrZXkYASABKAlSA2tleRIUCgV2YWx1ZRgCIAEoCVIFdmFsdW'
    'USIAoLZGVzY3JpcHRpb24YAyABKAlSC2Rlc2NyaXB0aW9u');

@$core.Deprecated('Use updateBillingSettingsRequestDescriptor instead')
const UpdateBillingSettingsRequest$json = {
  '1': 'UpdateBillingSettingsRequest',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.BillingSetting',
      '10': 'settings'
    },
    {'1': 'reason', '3': 2, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `UpdateBillingSettingsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateBillingSettingsRequestDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVCaWxsaW5nU2V0dGluZ3NSZXF1ZXN0EjcKCHNldHRpbmdzGAEgAygLMhsucGxhdG'
        'Zvcm0udjEuQmlsbGluZ1NldHRpbmdSCHNldHRpbmdzEhYKBnJlYXNvbhgCIAEoCVIGcmVhc29u');

@$core.Deprecated('Use updateBillingSettingsResponseDescriptor instead')
const UpdateBillingSettingsResponse$json = {
  '1': 'UpdateBillingSettingsResponse',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.platform.v1.BillingSetting',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `UpdateBillingSettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateBillingSettingsResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVCaWxsaW5nU2V0dGluZ3NSZXNwb25zZRI3CghzZXR0aW5ncxgBIAMoCzIbLnBsYX'
        'Rmb3JtLnYxLkJpbGxpbmdTZXR0aW5nUghzZXR0aW5ncw==');

@$core.Deprecated('Use setTenantOverrideRequestDescriptor instead')
const SetTenantOverrideRequest$json = {
  '1': 'SetTenantOverrideRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'feature_code', '3': 2, '4': 1, '5': 9, '10': 'featureCode'},
    {'1': 'enabled_set', '3': 3, '4': 1, '5': 8, '10': 'enabledSet'},
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'limit_set', '3': 5, '4': 1, '5': 8, '10': 'limitSet'},
    {'1': 'limit_value', '3': 6, '4': 1, '5': 3, '10': 'limitValue'},
    {'1': 'owner', '3': 7, '4': 1, '5': 9, '10': 'owner'},
    {'1': 'expires_at', '3': 8, '4': 1, '5': 9, '10': 'expiresAt'},
    {'1': 'reason', '3': 9, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `SetTenantOverrideRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setTenantOverrideRequestDescriptor = $convert.base64Decode(
    'ChhTZXRUZW5hbnRPdmVycmlkZVJlcXVlc3QSHQoKY29tcGFueV9pZBgBIAEoCVIJY29tcGFueU'
    'lkEiEKDGZlYXR1cmVfY29kZRgCIAEoCVILZmVhdHVyZUNvZGUSHwoLZW5hYmxlZF9zZXQYAyAB'
    'KAhSCmVuYWJsZWRTZXQSGAoHZW5hYmxlZBgEIAEoCFIHZW5hYmxlZBIbCglsaW1pdF9zZXQYBS'
    'ABKAhSCGxpbWl0U2V0Eh8KC2xpbWl0X3ZhbHVlGAYgASgDUgpsaW1pdFZhbHVlEhQKBW93bmVy'
    'GAcgASgJUgVvd25lchIdCgpleHBpcmVzX2F0GAggASgJUglleHBpcmVzQXQSFgoGcmVhc29uGA'
    'kgASgJUgZyZWFzb24=');

@$core.Deprecated('Use setTenantOverrideResponseDescriptor instead')
const SetTenantOverrideResponse$json = {
  '1': 'SetTenantOverrideResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `SetTenantOverrideResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setTenantOverrideResponseDescriptor =
    $convert.base64Decode(
        'ChlTZXRUZW5hbnRPdmVycmlkZVJlc3BvbnNlEg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use revokeTenantOverrideRequestDescriptor instead')
const RevokeTenantOverrideRequest$json = {
  '1': 'RevokeTenantOverrideRequest',
  '2': [
    {'1': 'override_id', '3': 1, '4': 1, '5': 9, '10': 'overrideId'},
    {'1': 'reason', '3': 2, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `RevokeTenantOverrideRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revokeTenantOverrideRequestDescriptor =
    $convert.base64Decode(
        'ChtSZXZva2VUZW5hbnRPdmVycmlkZVJlcXVlc3QSHwoLb3ZlcnJpZGVfaWQYASABKAlSCm92ZX'
        'JyaWRlSWQSFgoGcmVhc29uGAIgASgJUgZyZWFzb24=');

@$core.Deprecated('Use revokeTenantOverrideResponseDescriptor instead')
const RevokeTenantOverrideResponse$json = {
  '1': 'RevokeTenantOverrideResponse',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'feature_code', '3': 2, '4': 1, '5': 9, '10': 'featureCode'},
  ],
};

/// Descriptor for `RevokeTenantOverrideResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revokeTenantOverrideResponseDescriptor =
    $convert.base64Decode(
        'ChxSZXZva2VUZW5hbnRPdmVycmlkZVJlc3BvbnNlEh0KCmNvbXBhbnlfaWQYASABKAlSCWNvbX'
        'BhbnlJZBIhCgxmZWF0dXJlX2NvZGUYAiABKAlSC2ZlYXR1cmVDb2Rl');

@$core.Deprecated('Use upsertPlanPriceRequestDescriptor instead')
const UpsertPlanPriceRequest$json = {
  '1': 'UpsertPlanPriceRequest',
  '2': [
    {'1': 'plan_code', '3': 1, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'billing_cycle', '3': 2, '4': 1, '5': 9, '10': 'billingCycle'},
    {'1': 'base_price', '3': 3, '4': 1, '5': 9, '10': 'basePrice'},
    {'1': 'seat_price', '3': 4, '4': 1, '5': 9, '10': 'seatPrice'},
    {'1': 'currency', '3': 5, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'reason', '3': 6, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `UpsertPlanPriceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List upsertPlanPriceRequestDescriptor = $convert.base64Decode(
    'ChZVcHNlcnRQbGFuUHJpY2VSZXF1ZXN0EhsKCXBsYW5fY29kZRgBIAEoCVIIcGxhbkNvZGUSIw'
    'oNYmlsbGluZ19jeWNsZRgCIAEoCVIMYmlsbGluZ0N5Y2xlEh0KCmJhc2VfcHJpY2UYAyABKAlS'
    'CWJhc2VQcmljZRIdCgpzZWF0X3ByaWNlGAQgASgJUglzZWF0UHJpY2USGgoIY3VycmVuY3kYBS'
    'ABKAlSCGN1cnJlbmN5EhYKBnJlYXNvbhgGIAEoCVIGcmVhc29u');

@$core.Deprecated('Use upsertPlanPriceResponseDescriptor instead')
const UpsertPlanPriceResponse$json = {
  '1': 'UpsertPlanPriceResponse',
};

/// Descriptor for `UpsertPlanPriceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List upsertPlanPriceResponseDescriptor =
    $convert.base64Decode('ChdVcHNlcnRQbGFuUHJpY2VSZXNwb25zZQ==');

@$core.Deprecated('Use setPlanEntitlementRequestDescriptor instead')
const SetPlanEntitlementRequest$json = {
  '1': 'SetPlanEntitlementRequest',
  '2': [
    {'1': 'plan_code', '3': 1, '4': 1, '5': 9, '10': 'planCode'},
    {'1': 'feature_code', '3': 2, '4': 1, '5': 9, '10': 'featureCode'},
    {'1': 'enabled', '3': 3, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'limit_set', '3': 4, '4': 1, '5': 8, '10': 'limitSet'},
    {'1': 'limit_value', '3': 5, '4': 1, '5': 3, '10': 'limitValue'},
    {'1': 'reason', '3': 6, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `SetPlanEntitlementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setPlanEntitlementRequestDescriptor = $convert.base64Decode(
    'ChlTZXRQbGFuRW50aXRsZW1lbnRSZXF1ZXN0EhsKCXBsYW5fY29kZRgBIAEoCVIIcGxhbkNvZG'
    'USIQoMZmVhdHVyZV9jb2RlGAIgASgJUgtmZWF0dXJlQ29kZRIYCgdlbmFibGVkGAMgASgIUgdl'
    'bmFibGVkEhsKCWxpbWl0X3NldBgEIAEoCFIIbGltaXRTZXQSHwoLbGltaXRfdmFsdWUYBSABKA'
    'NSCmxpbWl0VmFsdWUSFgoGcmVhc29uGAYgASgJUgZyZWFzb24=');

@$core.Deprecated('Use setPlanEntitlementResponseDescriptor instead')
const SetPlanEntitlementResponse$json = {
  '1': 'SetPlanEntitlementResponse',
};

/// Descriptor for `SetPlanEntitlementResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setPlanEntitlementResponseDescriptor =
    $convert.base64Decode('ChpTZXRQbGFuRW50aXRsZW1lbnRSZXNwb25zZQ==');

@$core.Deprecated('Use createOperatorRequestDescriptor instead')
const CreateOperatorRequest$json = {
  '1': 'CreateOperatorRequest',
  '2': [
    {'1': 'email', '3': 1, '4': 1, '5': 9, '10': 'email'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'role', '3': 3, '4': 1, '5': 9, '10': 'role'},
    {'1': 'reason', '3': 4, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `CreateOperatorRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createOperatorRequestDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVPcGVyYXRvclJlcXVlc3QSFAoFZW1haWwYASABKAlSBWVtYWlsEhIKBG5hbWUYAi'
    'ABKAlSBG5hbWUSEgoEcm9sZRgDIAEoCVIEcm9sZRIWCgZyZWFzb24YBCABKAlSBnJlYXNvbg==');

@$core.Deprecated('Use createOperatorResponseDescriptor instead')
const CreateOperatorResponse$json = {
  '1': 'CreateOperatorResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `CreateOperatorResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createOperatorResponseDescriptor = $convert
    .base64Decode('ChZDcmVhdGVPcGVyYXRvclJlc3BvbnNlEg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use disableOperatorRequestDescriptor instead')
const DisableOperatorRequest$json = {
  '1': 'DisableOperatorRequest',
  '2': [
    {'1': 'operator_id', '3': 1, '4': 1, '5': 9, '10': 'operatorId'},
    {'1': 'reason', '3': 2, '4': 1, '5': 9, '10': 'reason'},
  ],
};

/// Descriptor for `DisableOperatorRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List disableOperatorRequestDescriptor =
    $convert.base64Decode(
        'ChZEaXNhYmxlT3BlcmF0b3JSZXF1ZXN0Eh8KC29wZXJhdG9yX2lkGAEgASgJUgpvcGVyYXRvck'
        'lkEhYKBnJlYXNvbhgCIAEoCVIGcmVhc29u');

@$core.Deprecated('Use disableOperatorResponseDescriptor instead')
const DisableOperatorResponse$json = {
  '1': 'DisableOperatorResponse',
};

/// Descriptor for `DisableOperatorResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List disableOperatorResponseDescriptor =
    $convert.base64Decode('ChdEaXNhYmxlT3BlcmF0b3JSZXNwb25zZQ==');

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
    {
      '1': 'ListReceivables',
      '2': '.platform.v1.ListReceivablesRequest',
      '3': '.platform.v1.ListReceivablesResponse'
    },
    {
      '1': 'RecordPayment',
      '2': '.platform.v1.RecordPaymentRequest',
      '3': '.platform.v1.RecordPaymentResponse'
    },
    {
      '1': 'CreateSubscription',
      '2': '.platform.v1.CreateSubscriptionRequest',
      '3': '.platform.v1.CreateSubscriptionResponse'
    },
    {
      '1': 'SetSeatCount',
      '2': '.platform.v1.SetSeatCountRequest',
      '3': '.platform.v1.SetSeatCountResponse'
    },
    {
      '1': 'ChangePlan',
      '2': '.platform.v1.ChangePlanRequest',
      '3': '.platform.v1.ChangePlanResponse'
    },
    {
      '1': 'CancelSubscription',
      '2': '.platform.v1.CancelSubscriptionRequest',
      '3': '.platform.v1.CancelSubscriptionResponse'
    },
    {
      '1': 'GetBillingSettings',
      '2': '.platform.v1.GetBillingSettingsRequest',
      '3': '.platform.v1.GetBillingSettingsResponse'
    },
    {
      '1': 'UpdateBillingSettings',
      '2': '.platform.v1.UpdateBillingSettingsRequest',
      '3': '.platform.v1.UpdateBillingSettingsResponse'
    },
    {
      '1': 'SetTenantOverride',
      '2': '.platform.v1.SetTenantOverrideRequest',
      '3': '.platform.v1.SetTenantOverrideResponse'
    },
    {
      '1': 'RevokeTenantOverride',
      '2': '.platform.v1.RevokeTenantOverrideRequest',
      '3': '.platform.v1.RevokeTenantOverrideResponse'
    },
    {
      '1': 'UpsertPlanPrice',
      '2': '.platform.v1.UpsertPlanPriceRequest',
      '3': '.platform.v1.UpsertPlanPriceResponse'
    },
    {
      '1': 'SetPlanEntitlement',
      '2': '.platform.v1.SetPlanEntitlementRequest',
      '3': '.platform.v1.SetPlanEntitlementResponse'
    },
    {
      '1': 'CreateOperator',
      '2': '.platform.v1.CreateOperatorRequest',
      '3': '.platform.v1.CreateOperatorResponse'
    },
    {
      '1': 'DisableOperator',
      '2': '.platform.v1.DisableOperatorRequest',
      '3': '.platform.v1.DisableOperatorResponse'
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
  '.platform.v1.ListReceivablesRequest': ListReceivablesRequest$json,
  '.platform.v1.ListReceivablesResponse': ListReceivablesResponse$json,
  '.platform.v1.Receivable': Receivable$json,
  '.platform.v1.RecordPaymentRequest': RecordPaymentRequest$json,
  '.platform.v1.RecordPaymentResponse': RecordPaymentResponse$json,
  '.platform.v1.CreateSubscriptionRequest': CreateSubscriptionRequest$json,
  '.platform.v1.CreateSubscriptionResponse': CreateSubscriptionResponse$json,
  '.platform.v1.SetSeatCountRequest': SetSeatCountRequest$json,
  '.platform.v1.SetSeatCountResponse': SetSeatCountResponse$json,
  '.platform.v1.ChangePlanRequest': ChangePlanRequest$json,
  '.platform.v1.ChangePlanResponse': ChangePlanResponse$json,
  '.platform.v1.CancelSubscriptionRequest': CancelSubscriptionRequest$json,
  '.platform.v1.CancelSubscriptionResponse': CancelSubscriptionResponse$json,
  '.platform.v1.GetBillingSettingsRequest': GetBillingSettingsRequest$json,
  '.platform.v1.GetBillingSettingsResponse': GetBillingSettingsResponse$json,
  '.platform.v1.BillingSetting': BillingSetting$json,
  '.platform.v1.UpdateBillingSettingsRequest':
      UpdateBillingSettingsRequest$json,
  '.platform.v1.UpdateBillingSettingsResponse':
      UpdateBillingSettingsResponse$json,
  '.platform.v1.SetTenantOverrideRequest': SetTenantOverrideRequest$json,
  '.platform.v1.SetTenantOverrideResponse': SetTenantOverrideResponse$json,
  '.platform.v1.RevokeTenantOverrideRequest': RevokeTenantOverrideRequest$json,
  '.platform.v1.RevokeTenantOverrideResponse':
      RevokeTenantOverrideResponse$json,
  '.platform.v1.UpsertPlanPriceRequest': UpsertPlanPriceRequest$json,
  '.platform.v1.UpsertPlanPriceResponse': UpsertPlanPriceResponse$json,
  '.platform.v1.SetPlanEntitlementRequest': SetPlanEntitlementRequest$json,
  '.platform.v1.SetPlanEntitlementResponse': SetPlanEntitlementResponse$json,
  '.platform.v1.CreateOperatorRequest': CreateOperatorRequest$json,
  '.platform.v1.CreateOperatorResponse': CreateOperatorResponse$json,
  '.platform.v1.DisableOperatorRequest': DisableOperatorRequest$json,
  '.platform.v1.DisableOperatorResponse': DisableOperatorResponse$json,
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
    'dGZvcm1BdWRpdFJlc3BvbnNlElwKD0xpc3RSZWNlaXZhYmxlcxIjLnBsYXRmb3JtLnYxLkxpc3'
    'RSZWNlaXZhYmxlc1JlcXVlc3QaJC5wbGF0Zm9ybS52MS5MaXN0UmVjZWl2YWJsZXNSZXNwb25z'
    'ZRJWCg1SZWNvcmRQYXltZW50EiEucGxhdGZvcm0udjEuUmVjb3JkUGF5bWVudFJlcXVlc3QaIi'
    '5wbGF0Zm9ybS52MS5SZWNvcmRQYXltZW50UmVzcG9uc2USZQoSQ3JlYXRlU3Vic2NyaXB0aW9u'
    'EiYucGxhdGZvcm0udjEuQ3JlYXRlU3Vic2NyaXB0aW9uUmVxdWVzdBonLnBsYXRmb3JtLnYxLk'
    'NyZWF0ZVN1YnNjcmlwdGlvblJlc3BvbnNlElMKDFNldFNlYXRDb3VudBIgLnBsYXRmb3JtLnYx'
    'LlNldFNlYXRDb3VudFJlcXVlc3QaIS5wbGF0Zm9ybS52MS5TZXRTZWF0Q291bnRSZXNwb25zZR'
    'JNCgpDaGFuZ2VQbGFuEh4ucGxhdGZvcm0udjEuQ2hhbmdlUGxhblJlcXVlc3QaHy5wbGF0Zm9y'
    'bS52MS5DaGFuZ2VQbGFuUmVzcG9uc2USZQoSQ2FuY2VsU3Vic2NyaXB0aW9uEiYucGxhdGZvcm'
    '0udjEuQ2FuY2VsU3Vic2NyaXB0aW9uUmVxdWVzdBonLnBsYXRmb3JtLnYxLkNhbmNlbFN1YnNj'
    'cmlwdGlvblJlc3BvbnNlEmUKEkdldEJpbGxpbmdTZXR0aW5ncxImLnBsYXRmb3JtLnYxLkdldE'
    'JpbGxpbmdTZXR0aW5nc1JlcXVlc3QaJy5wbGF0Zm9ybS52MS5HZXRCaWxsaW5nU2V0dGluZ3NS'
    'ZXNwb25zZRJuChVVcGRhdGVCaWxsaW5nU2V0dGluZ3MSKS5wbGF0Zm9ybS52MS5VcGRhdGVCaW'
    'xsaW5nU2V0dGluZ3NSZXF1ZXN0GioucGxhdGZvcm0udjEuVXBkYXRlQmlsbGluZ1NldHRpbmdz'
    'UmVzcG9uc2USYgoRU2V0VGVuYW50T3ZlcnJpZGUSJS5wbGF0Zm9ybS52MS5TZXRUZW5hbnRPdm'
    'VycmlkZVJlcXVlc3QaJi5wbGF0Zm9ybS52MS5TZXRUZW5hbnRPdmVycmlkZVJlc3BvbnNlEmsK'
    'FFJldm9rZVRlbmFudE92ZXJyaWRlEigucGxhdGZvcm0udjEuUmV2b2tlVGVuYW50T3ZlcnJpZG'
    'VSZXF1ZXN0GikucGxhdGZvcm0udjEuUmV2b2tlVGVuYW50T3ZlcnJpZGVSZXNwb25zZRJcCg9V'
    'cHNlcnRQbGFuUHJpY2USIy5wbGF0Zm9ybS52MS5VcHNlcnRQbGFuUHJpY2VSZXF1ZXN0GiQucG'
    'xhdGZvcm0udjEuVXBzZXJ0UGxhblByaWNlUmVzcG9uc2USZQoSU2V0UGxhbkVudGl0bGVtZW50'
    'EiYucGxhdGZvcm0udjEuU2V0UGxhbkVudGl0bGVtZW50UmVxdWVzdBonLnBsYXRmb3JtLnYxLl'
    'NldFBsYW5FbnRpdGxlbWVudFJlc3BvbnNlElkKDkNyZWF0ZU9wZXJhdG9yEiIucGxhdGZvcm0u'
    'djEuQ3JlYXRlT3BlcmF0b3JSZXF1ZXN0GiMucGxhdGZvcm0udjEuQ3JlYXRlT3BlcmF0b3JSZX'
    'Nwb25zZRJcCg9EaXNhYmxlT3BlcmF0b3ISIy5wbGF0Zm9ybS52MS5EaXNhYmxlT3BlcmF0b3JS'
    'ZXF1ZXN0GiQucGxhdGZvcm0udjEuRGlzYWJsZU9wZXJhdG9yUmVzcG9uc2U=');

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
