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
  ],
};

/// Descriptor for `ListCustomersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCustomersRequestDescriptor = $convert.base64Decode(
    'ChRMaXN0Q3VzdG9tZXJzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2l6ZR'
    'gCIAEoBVIIcGFnZVNpemUSGAoHa2V5d29yZBgDIAEoCVIHa2V5d29yZBInCg9pbmNsdWRlX2Rl'
    'bGV0ZWQYBCABKAhSDmluY2x1ZGVEZWxldGVkEhIKBHNvcnQYBSABKAlSBHNvcnQ=');

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
    'bWVyUmVzcG9uc2U=');
