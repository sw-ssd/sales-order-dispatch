// This is a generated file - do not edit.
//
// Generated from customers/v1/customer.proto.

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

@$core.Deprecated('Use customerDescriptor instead')
const Customer$json = {
  '1': 'Customer',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'customer_code', '3': 4, '4': 1, '5': 9, '10': 'customerCode'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'tax_id', '3': 6, '4': 1, '5': 9, '10': 'taxId'},
    {'1': 'payment_method_id', '3': 7, '4': 1, '5': 9, '10': 'paymentMethodId'},
    {
      '1': 'settlement_method_id',
      '3': 8,
      '4': 1,
      '5': 9,
      '10': 'settlementMethodId'
    },
    {'1': 'customer_type_id', '3': 9, '4': 1, '5': 9, '10': 'customerTypeId'},
    {'1': 'invoice_type_id', '3': 10, '4': 1, '5': 9, '10': 'invoiceTypeId'},
    {
      '1': 'default_sales_rep_id',
      '3': 11,
      '4': 1,
      '5': 9,
      '10': 'defaultSalesRepId'
    },
    {
      '1': 'preferred_delivery_days',
      '3': 12,
      '4': 3,
      '5': 8,
      '10': 'preferredDeliveryDays'
    },
    {'1': 'promo_tag_ids', '3': 13, '4': 3, '5': 3, '10': 'promoTagIds'},
    {'1': 'created_at', '3': 14, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 15, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 16, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `Customer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List customerDescriptor = $convert.base64Decode(
    'CghDdXN0b21lchIOCgJpZBgBIAEoCVICaWQSHQoKY29tcGFueV9pZBgCIAEoCVIJY29tcGFueU'
    'lkEiMKDWRlcGFydG1lbnRfaWQYAyABKAlSDGRlcGFydG1lbnRJZBIjCg1jdXN0b21lcl9jb2Rl'
    'GAQgASgJUgxjdXN0b21lckNvZGUSEgoEbmFtZRgFIAEoCVIEbmFtZRIVCgZ0YXhfaWQYBiABKA'
    'lSBXRheElkEioKEXBheW1lbnRfbWV0aG9kX2lkGAcgASgJUg9wYXltZW50TWV0aG9kSWQSMAoU'
    'c2V0dGxlbWVudF9tZXRob2RfaWQYCCABKAlSEnNldHRsZW1lbnRNZXRob2RJZBIoChBjdXN0b2'
    '1lcl90eXBlX2lkGAkgASgJUg5jdXN0b21lclR5cGVJZBImCg9pbnZvaWNlX3R5cGVfaWQYCiAB'
    'KAlSDWludm9pY2VUeXBlSWQSLwoUZGVmYXVsdF9zYWxlc19yZXBfaWQYCyABKAlSEWRlZmF1bH'
    'RTYWxlc1JlcElkEjYKF3ByZWZlcnJlZF9kZWxpdmVyeV9kYXlzGAwgAygIUhVwcmVmZXJyZWRE'
    'ZWxpdmVyeURheXMSIgoNcHJvbW9fdGFnX2lkcxgNIAMoA1ILcHJvbW9UYWdJZHMSHQoKY3JlYX'
    'RlZF9hdBgOIAEoCVIJY3JlYXRlZEF0Eh0KCnVwZGF0ZWRfYXQYDyABKAlSCXVwZGF0ZWRBdBId'
    'CgpkZWxldGVkX2F0GBAgASgJUglkZWxldGVkQXQ=');

@$core.Deprecated('Use listCustomersRequestDescriptor instead')
const ListCustomersRequest$json = {
  '1': 'ListCustomersRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'include_deleted', '3': 4, '4': 1, '5': 8, '10': 'includeDeleted'},
    {'1': 'sort', '3': 5, '4': 1, '5': 9, '10': 'sort'},
    {'1': 'desc', '3': 6, '4': 1, '5': 8, '10': 'desc'},
  ],
};

/// Descriptor for `ListCustomersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCustomersRequestDescriptor = $convert.base64Decode(
    'ChRMaXN0Q3VzdG9tZXJzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZR'
    'gCIAEoBVIIcGFnZVNpemUSGAoHa2V5d29yZBgDIAEoCVIHa2V5d29yZBInCg9pbmNsdWRlX2Rl'
    'bGV0ZWQYBCABKAhSDmluY2x1ZGVEZWxldGVkEhIKBHNvcnQYBSABKAlSBHNvcnQSEgoEZGVzYx'
    'gGIAEoCFIEZGVzYw==');

@$core.Deprecated('Use listCustomersResponseDescriptor instead')
const ListCustomersResponse$json = {
  '1': 'ListCustomersResponse',
  '2': [
    {
      '1': 'customers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.customers.v1.Customer',
      '10': 'customers'
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

/// Descriptor for `ListCustomersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCustomersResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0Q3VzdG9tZXJzUmVzcG9uc2USNAoJY3VzdG9tZXJzGAEgAygLMhYuY3VzdG9tZXJzLn'
    'YxLkN1c3RvbWVyUgljdXN0b21lcnMSOQoKcGFnaW5hdGlvbhgCIAEoCzIZLnNhbGVzb3JkZXIu'
    'djEuUGFnaW5hdGlvblIKcGFnaW5hdGlvbg==');

@$core.Deprecated('Use getCustomerRequestDescriptor instead')
const GetCustomerRequest$json = {
  '1': 'GetCustomerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetCustomerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCustomerRequestDescriptor =
    $convert.base64Decode('ChJHZXRDdXN0b21lclJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getCustomerResponseDescriptor instead')
const GetCustomerResponse$json = {
  '1': 'GetCustomerResponse',
  '2': [
    {
      '1': 'customer',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.Customer',
      '10': 'customer'
    },
  ],
};

/// Descriptor for `GetCustomerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCustomerResponseDescriptor = $convert.base64Decode(
    'ChNHZXRDdXN0b21lclJlc3BvbnNlEjIKCGN1c3RvbWVyGAEgASgLMhYuY3VzdG9tZXJzLnYxLk'
    'N1c3RvbWVyUghjdXN0b21lcg==');

@$core.Deprecated('Use createCustomerRequestDescriptor instead')
const CreateCustomerRequest$json = {
  '1': 'CreateCustomerRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'tax_id', '3': 2, '4': 1, '5': 9, '10': 'taxId'},
    {'1': 'payment_method_id', '3': 3, '4': 1, '5': 9, '10': 'paymentMethodId'},
    {
      '1': 'settlement_method_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'settlementMethodId'
    },
    {'1': 'customer_type_id', '3': 5, '4': 1, '5': 9, '10': 'customerTypeId'},
    {'1': 'invoice_type_id', '3': 6, '4': 1, '5': 9, '10': 'invoiceTypeId'},
    {
      '1': 'default_sales_rep_id',
      '3': 7,
      '4': 1,
      '5': 9,
      '10': 'defaultSalesRepId'
    },
    {
      '1': 'preferred_delivery_days',
      '3': 8,
      '4': 3,
      '5': 8,
      '10': 'preferredDeliveryDays'
    },
    {'1': 'promo_tag_ids', '3': 9, '4': 3, '5': 3, '10': 'promoTagIds'},
  ],
};

/// Descriptor for `CreateCustomerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCustomerRequestDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVDdXN0b21lclJlcXVlc3QSEgoEbmFtZRgBIAEoCVIEbmFtZRIVCgZ0YXhfaWQYAi'
    'ABKAlSBXRheElkEioKEXBheW1lbnRfbWV0aG9kX2lkGAMgASgJUg9wYXltZW50TWV0aG9kSWQS'
    'MAoUc2V0dGxlbWVudF9tZXRob2RfaWQYBCABKAlSEnNldHRsZW1lbnRNZXRob2RJZBIoChBjdX'
    'N0b21lcl90eXBlX2lkGAUgASgJUg5jdXN0b21lclR5cGVJZBImCg9pbnZvaWNlX3R5cGVfaWQY'
    'BiABKAlSDWludm9pY2VUeXBlSWQSLwoUZGVmYXVsdF9zYWxlc19yZXBfaWQYByABKAlSEWRlZm'
    'F1bHRTYWxlc1JlcElkEjYKF3ByZWZlcnJlZF9kZWxpdmVyeV9kYXlzGAggAygIUhVwcmVmZXJy'
    'ZWREZWxpdmVyeURheXMSIgoNcHJvbW9fdGFnX2lkcxgJIAMoA1ILcHJvbW9UYWdJZHM=');

@$core.Deprecated('Use createCustomerResponseDescriptor instead')
const CreateCustomerResponse$json = {
  '1': 'CreateCustomerResponse',
  '2': [
    {
      '1': 'customer',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.Customer',
      '10': 'customer'
    },
    {
      '1': 'primary_account_name',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'primaryAccountName'
    },
    {
      '1': 'primary_temp_password',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'primaryTempPassword'
    },
    {
      '1': 'sales_rep_account_name',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'salesRepAccountName'
    },
    {
      '1': 'sales_rep_temp_password',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'salesRepTempPassword'
    },
    {
      '1': 'account_manage_url',
      '3': 6,
      '4': 1,
      '5': 9,
      '10': 'accountManageUrl'
    },
  ],
};

/// Descriptor for `CreateCustomerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCustomerResponseDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVDdXN0b21lclJlc3BvbnNlEjIKCGN1c3RvbWVyGAEgASgLMhYuY3VzdG9tZXJzLn'
    'YxLkN1c3RvbWVyUghjdXN0b21lchIwChRwcmltYXJ5X2FjY291bnRfbmFtZRgCIAEoCVIScHJp'
    'bWFyeUFjY291bnROYW1lEjIKFXByaW1hcnlfdGVtcF9wYXNzd29yZBgDIAEoCVITcHJpbWFyeV'
    'RlbXBQYXNzd29yZBIzChZzYWxlc19yZXBfYWNjb3VudF9uYW1lGAQgASgJUhNzYWxlc1JlcEFj'
    'Y291bnROYW1lEjUKF3NhbGVzX3JlcF90ZW1wX3Bhc3N3b3JkGAUgASgJUhRzYWxlc1JlcFRlbX'
    'BQYXNzd29yZBIsChJhY2NvdW50X21hbmFnZV91cmwYBiABKAlSEGFjY291bnRNYW5hZ2VVcmw=');

@$core.Deprecated('Use updateCustomerRequestDescriptor instead')
const UpdateCustomerRequest$json = {
  '1': 'UpdateCustomerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'name', '17': true},
    {'1': 'tax_id', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'taxId', '17': true},
    {
      '1': 'payment_method_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '9': 2,
      '10': 'paymentMethodId',
      '17': true
    },
    {
      '1': 'settlement_method_id',
      '3': 5,
      '4': 1,
      '5': 9,
      '9': 3,
      '10': 'settlementMethodId',
      '17': true
    },
    {
      '1': 'customer_type_id',
      '3': 6,
      '4': 1,
      '5': 9,
      '9': 4,
      '10': 'customerTypeId',
      '17': true
    },
    {
      '1': 'invoice_type_id',
      '3': 7,
      '4': 1,
      '5': 9,
      '9': 5,
      '10': 'invoiceTypeId',
      '17': true
    },
    {
      '1': 'default_sales_rep_id',
      '3': 8,
      '4': 1,
      '5': 9,
      '9': 6,
      '10': 'defaultSalesRepId',
      '17': true
    },
    {
      '1': 'preferred_delivery_days',
      '3': 9,
      '4': 3,
      '5': 8,
      '10': 'preferredDeliveryDays'
    },
    {'1': 'promo_tag_ids', '3': 10, '4': 3, '5': 3, '10': 'promoTagIds'},
  ],
  '8': [
    {'1': '_name'},
    {'1': '_tax_id'},
    {'1': '_payment_method_id'},
    {'1': '_settlement_method_id'},
    {'1': '_customer_type_id'},
    {'1': '_invoice_type_id'},
    {'1': '_default_sales_rep_id'},
  ],
};

/// Descriptor for `UpdateCustomerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCustomerRequestDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVDdXN0b21lclJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhcKBG5hbWUYAiABKAlIAF'
    'IEbmFtZYgBARIaCgZ0YXhfaWQYAyABKAlIAVIFdGF4SWSIAQESLwoRcGF5bWVudF9tZXRob2Rf'
    'aWQYBCABKAlIAlIPcGF5bWVudE1ldGhvZElkiAEBEjUKFHNldHRsZW1lbnRfbWV0aG9kX2lkGA'
    'UgASgJSANSEnNldHRsZW1lbnRNZXRob2RJZIgBARItChBjdXN0b21lcl90eXBlX2lkGAYgASgJ'
    'SARSDmN1c3RvbWVyVHlwZUlkiAEBEisKD2ludm9pY2VfdHlwZV9pZBgHIAEoCUgFUg1pbnZvaW'
    'NlVHlwZUlkiAEBEjQKFGRlZmF1bHRfc2FsZXNfcmVwX2lkGAggASgJSAZSEWRlZmF1bHRTYWxl'
    'c1JlcElkiAEBEjYKF3ByZWZlcnJlZF9kZWxpdmVyeV9kYXlzGAkgAygIUhVwcmVmZXJyZWREZW'
    'xpdmVyeURheXMSIgoNcHJvbW9fdGFnX2lkcxgKIAMoA1ILcHJvbW9UYWdJZHNCBwoFX25hbWVC'
    'CQoHX3RheF9pZEIUChJfcGF5bWVudF9tZXRob2RfaWRCFwoVX3NldHRsZW1lbnRfbWV0aG9kX2'
    'lkQhMKEV9jdXN0b21lcl90eXBlX2lkQhIKEF9pbnZvaWNlX3R5cGVfaWRCFwoVX2RlZmF1bHRf'
    'c2FsZXNfcmVwX2lk');

@$core.Deprecated('Use updateCustomerResponseDescriptor instead')
const UpdateCustomerResponse$json = {
  '1': 'UpdateCustomerResponse',
  '2': [
    {
      '1': 'customer',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.Customer',
      '10': 'customer'
    },
  ],
};

/// Descriptor for `UpdateCustomerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCustomerResponseDescriptor =
    $convert.base64Decode(
        'ChZVcGRhdGVDdXN0b21lclJlc3BvbnNlEjIKCGN1c3RvbWVyGAEgASgLMhYuY3VzdG9tZXJzLn'
        'YxLkN1c3RvbWVyUghjdXN0b21lcg==');

@$core.Deprecated('Use deleteCustomerRequestDescriptor instead')
const DeleteCustomerRequest$json = {
  '1': 'DeleteCustomerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteCustomerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCustomerRequestDescriptor = $convert
    .base64Decode('ChVEZWxldGVDdXN0b21lclJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteCustomerResponseDescriptor instead')
const DeleteCustomerResponse$json = {
  '1': 'DeleteCustomerResponse',
};

/// Descriptor for `DeleteCustomerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCustomerResponseDescriptor =
    $convert.base64Decode('ChZEZWxldGVDdXN0b21lclJlc3BvbnNl');

@$core.Deprecated('Use restoreCustomerRequestDescriptor instead')
const RestoreCustomerRequest$json = {
  '1': 'RestoreCustomerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `RestoreCustomerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreCustomerRequestDescriptor = $convert
    .base64Decode('ChZSZXN0b3JlQ3VzdG9tZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use restoreCustomerResponseDescriptor instead')
const RestoreCustomerResponse$json = {
  '1': 'RestoreCustomerResponse',
  '2': [
    {
      '1': 'customer',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.Customer',
      '10': 'customer'
    },
  ],
};

/// Descriptor for `RestoreCustomerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreCustomerResponseDescriptor =
    $convert.base64Decode(
        'ChdSZXN0b3JlQ3VzdG9tZXJSZXNwb25zZRIyCghjdXN0b21lchgBIAEoCzIWLmN1c3RvbWVycy'
        '52MS5DdXN0b21lclIIY3VzdG9tZXI=');

@$core.Deprecated('Use customerAddressDescriptor instead')
const CustomerAddress$json = {
  '1': 'CustomerAddress',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'customer_id', '3': 2, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'type', '3': 3, '4': 1, '5': 9, '10': 'type'},
    {'1': 'recipient_name', '3': 4, '4': 1, '5': 9, '10': 'recipientName'},
    {'1': 'phone', '3': 5, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'address_line', '3': 6, '4': 1, '5': 9, '10': 'addressLine'},
    {'1': 'city', '3': 7, '4': 1, '5': 9, '10': 'city'},
    {'1': 'postal_code', '3': 8, '4': 1, '5': 9, '10': 'postalCode'},
    {'1': 'is_default', '3': 9, '4': 1, '5': 8, '10': 'isDefault'},
    {'1': 'created_at', '3': 10, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 11, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 12, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `CustomerAddress`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List customerAddressDescriptor = $convert.base64Decode(
    'Cg9DdXN0b21lckFkZHJlc3MSDgoCaWQYASABKAlSAmlkEh8KC2N1c3RvbWVyX2lkGAIgASgJUg'
    'pjdXN0b21lcklkEhIKBHR5cGUYAyABKAlSBHR5cGUSJQoOcmVjaXBpZW50X25hbWUYBCABKAlS'
    'DXJlY2lwaWVudE5hbWUSFAoFcGhvbmUYBSABKAlSBXBob25lEiEKDGFkZHJlc3NfbGluZRgGIA'
    'EoCVILYWRkcmVzc0xpbmUSEgoEY2l0eRgHIAEoCVIEY2l0eRIfCgtwb3N0YWxfY29kZRgIIAEo'
    'CVIKcG9zdGFsQ29kZRIdCgppc19kZWZhdWx0GAkgASgIUglpc0RlZmF1bHQSHQoKY3JlYXRlZF'
    '9hdBgKIAEoCVIJY3JlYXRlZEF0Eh0KCnVwZGF0ZWRfYXQYCyABKAlSCXVwZGF0ZWRBdBIdCgpk'
    'ZWxldGVkX2F0GAwgASgJUglkZWxldGVkQXQ=');

@$core.Deprecated('Use listAddressesRequestDescriptor instead')
const ListAddressesRequest$json = {
  '1': 'ListAddressesRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'include_deleted', '3': 2, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListAddressesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAddressesRequestDescriptor = $convert.base64Decode(
    'ChRMaXN0QWRkcmVzc2VzUmVxdWVzdBIfCgtjdXN0b21lcl9pZBgBIAEoCVIKY3VzdG9tZXJJZB'
    'InCg9pbmNsdWRlX2RlbGV0ZWQYAiABKAhSDmluY2x1ZGVEZWxldGVk');

@$core.Deprecated('Use listAddressesResponseDescriptor instead')
const ListAddressesResponse$json = {
  '1': 'ListAddressesResponse',
  '2': [
    {
      '1': 'addresses',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.customers.v1.CustomerAddress',
      '10': 'addresses'
    },
  ],
};

/// Descriptor for `ListAddressesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAddressesResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0QWRkcmVzc2VzUmVzcG9uc2USOwoJYWRkcmVzc2VzGAEgAygLMh0uY3VzdG9tZXJzLn'
    'YxLkN1c3RvbWVyQWRkcmVzc1IJYWRkcmVzc2Vz');

@$core.Deprecated('Use addAddressRequestDescriptor instead')
const AddAddressRequest$json = {
  '1': 'AddAddressRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'recipient_name', '3': 3, '4': 1, '5': 9, '10': 'recipientName'},
    {'1': 'phone', '3': 4, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'address_line', '3': 5, '4': 1, '5': 9, '10': 'addressLine'},
    {'1': 'city', '3': 6, '4': 1, '5': 9, '10': 'city'},
    {'1': 'postal_code', '3': 7, '4': 1, '5': 9, '10': 'postalCode'},
    {'1': 'is_default', '3': 8, '4': 1, '5': 8, '10': 'isDefault'},
  ],
};

/// Descriptor for `AddAddressRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addAddressRequestDescriptor = $convert.base64Decode(
    'ChFBZGRBZGRyZXNzUmVxdWVzdBIfCgtjdXN0b21lcl9pZBgBIAEoCVIKY3VzdG9tZXJJZBISCg'
    'R0eXBlGAIgASgJUgR0eXBlEiUKDnJlY2lwaWVudF9uYW1lGAMgASgJUg1yZWNpcGllbnROYW1l'
    'EhQKBXBob25lGAQgASgJUgVwaG9uZRIhCgxhZGRyZXNzX2xpbmUYBSABKAlSC2FkZHJlc3NMaW'
    '5lEhIKBGNpdHkYBiABKAlSBGNpdHkSHwoLcG9zdGFsX2NvZGUYByABKAlSCnBvc3RhbENvZGUS'
    'HQoKaXNfZGVmYXVsdBgIIAEoCFIJaXNEZWZhdWx0');

@$core.Deprecated('Use addAddressResponseDescriptor instead')
const AddAddressResponse$json = {
  '1': 'AddAddressResponse',
  '2': [
    {
      '1': 'address',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.CustomerAddress',
      '10': 'address'
    },
  ],
};

/// Descriptor for `AddAddressResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addAddressResponseDescriptor = $convert.base64Decode(
    'ChJBZGRBZGRyZXNzUmVzcG9uc2USNwoHYWRkcmVzcxgBIAEoCzIdLmN1c3RvbWVycy52MS5DdX'
    'N0b21lckFkZHJlc3NSB2FkZHJlc3M=');

@$core.Deprecated('Use updateAddressRequestDescriptor instead')
const UpdateAddressRequest$json = {
  '1': 'UpdateAddressRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'type', '17': true},
    {
      '1': 'recipient_name',
      '3': 3,
      '4': 1,
      '5': 9,
      '9': 1,
      '10': 'recipientName',
      '17': true
    },
    {'1': 'phone', '3': 4, '4': 1, '5': 9, '9': 2, '10': 'phone', '17': true},
    {
      '1': 'address_line',
      '3': 5,
      '4': 1,
      '5': 9,
      '9': 3,
      '10': 'addressLine',
      '17': true
    },
    {'1': 'city', '3': 6, '4': 1, '5': 9, '9': 4, '10': 'city', '17': true},
    {
      '1': 'postal_code',
      '3': 7,
      '4': 1,
      '5': 9,
      '9': 5,
      '10': 'postalCode',
      '17': true
    },
    {
      '1': 'is_default',
      '3': 8,
      '4': 1,
      '5': 8,
      '9': 6,
      '10': 'isDefault',
      '17': true
    },
  ],
  '8': [
    {'1': '_type'},
    {'1': '_recipient_name'},
    {'1': '_phone'},
    {'1': '_address_line'},
    {'1': '_city'},
    {'1': '_postal_code'},
    {'1': '_is_default'},
  ],
};

/// Descriptor for `UpdateAddressRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAddressRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVBZGRyZXNzUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSFwoEdHlwZRgCIAEoCUgAUg'
    'R0eXBliAEBEioKDnJlY2lwaWVudF9uYW1lGAMgASgJSAFSDXJlY2lwaWVudE5hbWWIAQESGQoF'
    'cGhvbmUYBCABKAlIAlIFcGhvbmWIAQESJgoMYWRkcmVzc19saW5lGAUgASgJSANSC2FkZHJlc3'
    'NMaW5liAEBEhcKBGNpdHkYBiABKAlIBFIEY2l0eYgBARIkCgtwb3N0YWxfY29kZRgHIAEoCUgF'
    'Ugpwb3N0YWxDb2RliAEBEiIKCmlzX2RlZmF1bHQYCCABKAhIBlIJaXNEZWZhdWx0iAEBQgcKBV'
    '90eXBlQhEKD19yZWNpcGllbnRfbmFtZUIICgZfcGhvbmVCDwoNX2FkZHJlc3NfbGluZUIHCgVf'
    'Y2l0eUIOCgxfcG9zdGFsX2NvZGVCDQoLX2lzX2RlZmF1bHQ=');

@$core.Deprecated('Use updateAddressResponseDescriptor instead')
const UpdateAddressResponse$json = {
  '1': 'UpdateAddressResponse',
  '2': [
    {
      '1': 'address',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.CustomerAddress',
      '10': 'address'
    },
  ],
};

/// Descriptor for `UpdateAddressResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAddressResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVBZGRyZXNzUmVzcG9uc2USNwoHYWRkcmVzcxgBIAEoCzIdLmN1c3RvbWVycy52MS'
    '5DdXN0b21lckFkZHJlc3NSB2FkZHJlc3M=');

@$core.Deprecated('Use deleteAddressRequestDescriptor instead')
const DeleteAddressRequest$json = {
  '1': 'DeleteAddressRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteAddressRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAddressRequestDescriptor = $convert
    .base64Decode('ChREZWxldGVBZGRyZXNzUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use deleteAddressResponseDescriptor instead')
const DeleteAddressResponse$json = {
  '1': 'DeleteAddressResponse',
};

/// Descriptor for `DeleteAddressResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAddressResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVBZGRyZXNzUmVzcG9uc2U=');

@$core.Deprecated('Use customerContactDescriptor instead')
const CustomerContact$json = {
  '1': 'CustomerContact',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'customer_id', '3': 2, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'title', '3': 4, '4': 1, '5': 9, '10': 'title'},
    {'1': 'email', '3': 5, '4': 1, '5': 9, '10': 'email'},
    {'1': 'phone', '3': 6, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'is_default', '3': 7, '4': 1, '5': 8, '10': 'isDefault'},
    {'1': 'created_at', '3': 8, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 9, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 10, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `CustomerContact`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List customerContactDescriptor = $convert.base64Decode(
    'Cg9DdXN0b21lckNvbnRhY3QSDgoCaWQYASABKAlSAmlkEh8KC2N1c3RvbWVyX2lkGAIgASgJUg'
    'pjdXN0b21lcklkEhIKBG5hbWUYAyABKAlSBG5hbWUSFAoFdGl0bGUYBCABKAlSBXRpdGxlEhQK'
    'BWVtYWlsGAUgASgJUgVlbWFpbBIUCgVwaG9uZRgGIAEoCVIFcGhvbmUSHQoKaXNfZGVmYXVsdB'
    'gHIAEoCFIJaXNEZWZhdWx0Eh0KCmNyZWF0ZWRfYXQYCCABKAlSCWNyZWF0ZWRBdBIdCgp1cGRh'
    'dGVkX2F0GAkgASgJUgl1cGRhdGVkQXQSHQoKZGVsZXRlZF9hdBgKIAEoCVIJZGVsZXRlZEF0');

@$core.Deprecated('Use listContactsRequestDescriptor instead')
const ListContactsRequest$json = {
  '1': 'ListContactsRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'include_deleted', '3': 2, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListContactsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listContactsRequestDescriptor = $convert.base64Decode(
    'ChNMaXN0Q29udGFjdHNSZXF1ZXN0Eh8KC2N1c3RvbWVyX2lkGAEgASgJUgpjdXN0b21lcklkEi'
    'cKD2luY2x1ZGVfZGVsZXRlZBgCIAEoCFIOaW5jbHVkZURlbGV0ZWQ=');

@$core.Deprecated('Use listContactsResponseDescriptor instead')
const ListContactsResponse$json = {
  '1': 'ListContactsResponse',
  '2': [
    {
      '1': 'contacts',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.customers.v1.CustomerContact',
      '10': 'contacts'
    },
  ],
};

/// Descriptor for `ListContactsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listContactsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0Q29udGFjdHNSZXNwb25zZRI5Cghjb250YWN0cxgBIAMoCzIdLmN1c3RvbWVycy52MS'
    '5DdXN0b21lckNvbnRhY3RSCGNvbnRhY3Rz');

@$core.Deprecated('Use addContactRequestDescriptor instead')
const AddContactRequest$json = {
  '1': 'AddContactRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'title', '3': 3, '4': 1, '5': 9, '10': 'title'},
    {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
    {'1': 'phone', '3': 5, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'is_default', '3': 6, '4': 1, '5': 8, '10': 'isDefault'},
  ],
};

/// Descriptor for `AddContactRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addContactRequestDescriptor = $convert.base64Decode(
    'ChFBZGRDb250YWN0UmVxdWVzdBIfCgtjdXN0b21lcl9pZBgBIAEoCVIKY3VzdG9tZXJJZBISCg'
    'RuYW1lGAIgASgJUgRuYW1lEhQKBXRpdGxlGAMgASgJUgV0aXRsZRIUCgVlbWFpbBgEIAEoCVIF'
    'ZW1haWwSFAoFcGhvbmUYBSABKAlSBXBob25lEh0KCmlzX2RlZmF1bHQYBiABKAhSCWlzRGVmYX'
    'VsdA==');

@$core.Deprecated('Use addContactResponseDescriptor instead')
const AddContactResponse$json = {
  '1': 'AddContactResponse',
  '2': [
    {
      '1': 'contact',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.CustomerContact',
      '10': 'contact'
    },
  ],
};

/// Descriptor for `AddContactResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addContactResponseDescriptor = $convert.base64Decode(
    'ChJBZGRDb250YWN0UmVzcG9uc2USNwoHY29udGFjdBgBIAEoCzIdLmN1c3RvbWVycy52MS5DdX'
    'N0b21lckNvbnRhY3RSB2NvbnRhY3Q=');

@$core.Deprecated('Use updateContactRequestDescriptor instead')
const UpdateContactRequest$json = {
  '1': 'UpdateContactRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'name', '17': true},
    {'1': 'title', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'title', '17': true},
    {'1': 'email', '3': 4, '4': 1, '5': 9, '9': 2, '10': 'email', '17': true},
    {'1': 'phone', '3': 5, '4': 1, '5': 9, '9': 3, '10': 'phone', '17': true},
    {
      '1': 'is_default',
      '3': 6,
      '4': 1,
      '5': 8,
      '9': 4,
      '10': 'isDefault',
      '17': true
    },
  ],
  '8': [
    {'1': '_name'},
    {'1': '_title'},
    {'1': '_email'},
    {'1': '_phone'},
    {'1': '_is_default'},
  ],
};

/// Descriptor for `UpdateContactRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateContactRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVDb250YWN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSFwoEbmFtZRgCIAEoCUgAUg'
    'RuYW1liAEBEhkKBXRpdGxlGAMgASgJSAFSBXRpdGxliAEBEhkKBWVtYWlsGAQgASgJSAJSBWVt'
    'YWlsiAEBEhkKBXBob25lGAUgASgJSANSBXBob25liAEBEiIKCmlzX2RlZmF1bHQYBiABKAhIBF'
    'IJaXNEZWZhdWx0iAEBQgcKBV9uYW1lQggKBl90aXRsZUIICgZfZW1haWxCCAoGX3Bob25lQg0K'
    'C19pc19kZWZhdWx0');

@$core.Deprecated('Use updateContactResponseDescriptor instead')
const UpdateContactResponse$json = {
  '1': 'UpdateContactResponse',
  '2': [
    {
      '1': 'contact',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.customers.v1.CustomerContact',
      '10': 'contact'
    },
  ],
};

/// Descriptor for `UpdateContactResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateContactResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVDb250YWN0UmVzcG9uc2USNwoHY29udGFjdBgBIAEoCzIdLmN1c3RvbWVycy52MS'
    '5DdXN0b21lckNvbnRhY3RSB2NvbnRhY3Q=');

@$core.Deprecated('Use deleteContactRequestDescriptor instead')
const DeleteContactRequest$json = {
  '1': 'DeleteContactRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteContactRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteContactRequestDescriptor = $convert
    .base64Decode('ChREZWxldGVDb250YWN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use deleteContactResponseDescriptor instead')
const DeleteContactResponse$json = {
  '1': 'DeleteContactResponse',
};

/// Descriptor for `DeleteContactResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteContactResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVDb250YWN0UmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> CustomerServiceBase$json = {
  '1': 'CustomerService',
  '2': [
    {
      '1': 'ListCustomers',
      '2': '.customers.v1.ListCustomersRequest',
      '3': '.customers.v1.ListCustomersResponse'
    },
    {
      '1': 'GetCustomer',
      '2': '.customers.v1.GetCustomerRequest',
      '3': '.customers.v1.GetCustomerResponse'
    },
    {
      '1': 'CreateCustomer',
      '2': '.customers.v1.CreateCustomerRequest',
      '3': '.customers.v1.CreateCustomerResponse'
    },
    {
      '1': 'UpdateCustomer',
      '2': '.customers.v1.UpdateCustomerRequest',
      '3': '.customers.v1.UpdateCustomerResponse'
    },
    {
      '1': 'DeleteCustomer',
      '2': '.customers.v1.DeleteCustomerRequest',
      '3': '.customers.v1.DeleteCustomerResponse'
    },
    {
      '1': 'RestoreCustomer',
      '2': '.customers.v1.RestoreCustomerRequest',
      '3': '.customers.v1.RestoreCustomerResponse'
    },
    {
      '1': 'ListAddresses',
      '2': '.customers.v1.ListAddressesRequest',
      '3': '.customers.v1.ListAddressesResponse'
    },
    {
      '1': 'AddAddress',
      '2': '.customers.v1.AddAddressRequest',
      '3': '.customers.v1.AddAddressResponse'
    },
    {
      '1': 'UpdateAddress',
      '2': '.customers.v1.UpdateAddressRequest',
      '3': '.customers.v1.UpdateAddressResponse'
    },
    {
      '1': 'DeleteAddress',
      '2': '.customers.v1.DeleteAddressRequest',
      '3': '.customers.v1.DeleteAddressResponse'
    },
    {
      '1': 'ListContacts',
      '2': '.customers.v1.ListContactsRequest',
      '3': '.customers.v1.ListContactsResponse'
    },
    {
      '1': 'AddContact',
      '2': '.customers.v1.AddContactRequest',
      '3': '.customers.v1.AddContactResponse'
    },
    {
      '1': 'UpdateContact',
      '2': '.customers.v1.UpdateContactRequest',
      '3': '.customers.v1.UpdateContactResponse'
    },
    {
      '1': 'DeleteContact',
      '2': '.customers.v1.DeleteContactRequest',
      '3': '.customers.v1.DeleteContactResponse'
    },
  ],
};

@$core.Deprecated('Use customerServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    CustomerServiceBase$messageJson = {
  '.customers.v1.ListCustomersRequest': ListCustomersRequest$json,
  '.customers.v1.ListCustomersResponse': ListCustomersResponse$json,
  '.customers.v1.Customer': Customer$json,
  '.salesorder.v1.Pagination': $0.Pagination$json,
  '.customers.v1.GetCustomerRequest': GetCustomerRequest$json,
  '.customers.v1.GetCustomerResponse': GetCustomerResponse$json,
  '.customers.v1.CreateCustomerRequest': CreateCustomerRequest$json,
  '.customers.v1.CreateCustomerResponse': CreateCustomerResponse$json,
  '.customers.v1.UpdateCustomerRequest': UpdateCustomerRequest$json,
  '.customers.v1.UpdateCustomerResponse': UpdateCustomerResponse$json,
  '.customers.v1.DeleteCustomerRequest': DeleteCustomerRequest$json,
  '.customers.v1.DeleteCustomerResponse': DeleteCustomerResponse$json,
  '.customers.v1.RestoreCustomerRequest': RestoreCustomerRequest$json,
  '.customers.v1.RestoreCustomerResponse': RestoreCustomerResponse$json,
  '.customers.v1.ListAddressesRequest': ListAddressesRequest$json,
  '.customers.v1.ListAddressesResponse': ListAddressesResponse$json,
  '.customers.v1.CustomerAddress': CustomerAddress$json,
  '.customers.v1.AddAddressRequest': AddAddressRequest$json,
  '.customers.v1.AddAddressResponse': AddAddressResponse$json,
  '.customers.v1.UpdateAddressRequest': UpdateAddressRequest$json,
  '.customers.v1.UpdateAddressResponse': UpdateAddressResponse$json,
  '.customers.v1.DeleteAddressRequest': DeleteAddressRequest$json,
  '.customers.v1.DeleteAddressResponse': DeleteAddressResponse$json,
  '.customers.v1.ListContactsRequest': ListContactsRequest$json,
  '.customers.v1.ListContactsResponse': ListContactsResponse$json,
  '.customers.v1.CustomerContact': CustomerContact$json,
  '.customers.v1.AddContactRequest': AddContactRequest$json,
  '.customers.v1.AddContactResponse': AddContactResponse$json,
  '.customers.v1.UpdateContactRequest': UpdateContactRequest$json,
  '.customers.v1.UpdateContactResponse': UpdateContactResponse$json,
  '.customers.v1.DeleteContactRequest': DeleteContactRequest$json,
  '.customers.v1.DeleteContactResponse': DeleteContactResponse$json,
};

/// Descriptor for `CustomerService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List customerServiceDescriptor = $convert.base64Decode(
    'Cg9DdXN0b21lclNlcnZpY2USWAoNTGlzdEN1c3RvbWVycxIiLmN1c3RvbWVycy52MS5MaXN0Q3'
    'VzdG9tZXJzUmVxdWVzdBojLmN1c3RvbWVycy52MS5MaXN0Q3VzdG9tZXJzUmVzcG9uc2USUgoL'
    'R2V0Q3VzdG9tZXISIC5jdXN0b21lcnMudjEuR2V0Q3VzdG9tZXJSZXF1ZXN0GiEuY3VzdG9tZX'
    'JzLnYxLkdldEN1c3RvbWVyUmVzcG9uc2USWwoOQ3JlYXRlQ3VzdG9tZXISIy5jdXN0b21lcnMu'
    'djEuQ3JlYXRlQ3VzdG9tZXJSZXF1ZXN0GiQuY3VzdG9tZXJzLnYxLkNyZWF0ZUN1c3RvbWVyUm'
    'VzcG9uc2USWwoOVXBkYXRlQ3VzdG9tZXISIy5jdXN0b21lcnMudjEuVXBkYXRlQ3VzdG9tZXJS'
    'ZXF1ZXN0GiQuY3VzdG9tZXJzLnYxLlVwZGF0ZUN1c3RvbWVyUmVzcG9uc2USWwoORGVsZXRlQ3'
    'VzdG9tZXISIy5jdXN0b21lcnMudjEuRGVsZXRlQ3VzdG9tZXJSZXF1ZXN0GiQuY3VzdG9tZXJz'
    'LnYxLkRlbGV0ZUN1c3RvbWVyUmVzcG9uc2USXgoPUmVzdG9yZUN1c3RvbWVyEiQuY3VzdG9tZX'
    'JzLnYxLlJlc3RvcmVDdXN0b21lclJlcXVlc3QaJS5jdXN0b21lcnMudjEuUmVzdG9yZUN1c3Rv'
    'bWVyUmVzcG9uc2USWAoNTGlzdEFkZHJlc3NlcxIiLmN1c3RvbWVycy52MS5MaXN0QWRkcmVzc2'
    'VzUmVxdWVzdBojLmN1c3RvbWVycy52MS5MaXN0QWRkcmVzc2VzUmVzcG9uc2USTwoKQWRkQWRk'
    'cmVzcxIfLmN1c3RvbWVycy52MS5BZGRBZGRyZXNzUmVxdWVzdBogLmN1c3RvbWVycy52MS5BZG'
    'RBZGRyZXNzUmVzcG9uc2USWAoNVXBkYXRlQWRkcmVzcxIiLmN1c3RvbWVycy52MS5VcGRhdGVB'
    'ZGRyZXNzUmVxdWVzdBojLmN1c3RvbWVycy52MS5VcGRhdGVBZGRyZXNzUmVzcG9uc2USWAoNRG'
    'VsZXRlQWRkcmVzcxIiLmN1c3RvbWVycy52MS5EZWxldGVBZGRyZXNzUmVxdWVzdBojLmN1c3Rv'
    'bWVycy52MS5EZWxldGVBZGRyZXNzUmVzcG9uc2USVQoMTGlzdENvbnRhY3RzEiEuY3VzdG9tZX'
    'JzLnYxLkxpc3RDb250YWN0c1JlcXVlc3QaIi5jdXN0b21lcnMudjEuTGlzdENvbnRhY3RzUmVz'
    'cG9uc2USTwoKQWRkQ29udGFjdBIfLmN1c3RvbWVycy52MS5BZGRDb250YWN0UmVxdWVzdBogLm'
    'N1c3RvbWVycy52MS5BZGRDb250YWN0UmVzcG9uc2USWAoNVXBkYXRlQ29udGFjdBIiLmN1c3Rv'
    'bWVycy52MS5VcGRhdGVDb250YWN0UmVxdWVzdBojLmN1c3RvbWVycy52MS5VcGRhdGVDb250YW'
    'N0UmVzcG9uc2USWAoNRGVsZXRlQ29udGFjdBIiLmN1c3RvbWVycy52MS5EZWxldGVDb250YWN0'
    'UmVxdWVzdBojLmN1c3RvbWVycy52MS5EZWxldGVDb250YWN0UmVzcG9uc2U=');
