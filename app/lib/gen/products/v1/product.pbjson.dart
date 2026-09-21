// This is a generated file - do not edit.
//
// Generated from products/v1/product.proto.

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

import 'package:protobuf/well_known_types/google/protobuf/struct.pbjson.dart'
    as $0;

import '../../salesorder/v1/common.pbjson.dart' as $1;

@$core.Deprecated('Use productUnitDescriptor instead')
const ProductUnit$json = {
  '1': 'ProductUnit',
  '2': [
    {'1': 'unit_code', '3': 1, '4': 1, '5': 9, '10': 'unitCode'},
    {'1': 'conversion_rate', '3': 2, '4': 1, '5': 9, '10': 'conversionRate'},
    {'1': 'is_base', '3': 3, '4': 1, '5': 8, '10': 'isBase'},
    {'1': 'sort_order', '3': 4, '4': 1, '5': 5, '10': 'sortOrder'},
    {'1': 'size_desc', '3': 5, '4': 1, '5': 9, '10': 'sizeDesc'},
  ],
};

/// Descriptor for `ProductUnit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List productUnitDescriptor = $convert.base64Decode(
    'CgtQcm9kdWN0VW5pdBIbCgl1bml0X2NvZGUYASABKAlSCHVuaXRDb2RlEicKD2NvbnZlcnNpb2'
    '5fcmF0ZRgCIAEoCVIOY29udmVyc2lvblJhdGUSFwoHaXNfYmFzZRgDIAEoCFIGaXNCYXNlEh0K'
    'CnNvcnRfb3JkZXIYBCABKAVSCXNvcnRPcmRlchIbCglzaXplX2Rlc2MYBSABKAlSCHNpemVEZX'
    'Nj');

@$core.Deprecated('Use productProcessingSpecDescriptor instead')
const ProductProcessingSpec$json = {
  '1': 'ProductProcessingSpec',
  '2': [
    {
      '1': 'processing_spec_id',
      '3': 1,
      '4': 1,
      '5': 9,
      '10': 'processingSpecId'
    },
    {
      '1': 'attributes',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'attributes'
    },
  ],
};

/// Descriptor for `ProductProcessingSpec`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List productProcessingSpecDescriptor = $convert.base64Decode(
    'ChVQcm9kdWN0UHJvY2Vzc2luZ1NwZWMSLAoScHJvY2Vzc2luZ19zcGVjX2lkGAEgASgJUhBwcm'
    '9jZXNzaW5nU3BlY0lkEjcKCmF0dHJpYnV0ZXMYAiABKAsyFy5nb29nbGUucHJvdG9idWYuU3Ry'
    'dWN0UgphdHRyaWJ1dGVz');

@$core.Deprecated('Use productDescriptor instead')
const Product$json = {
  '1': 'Product',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'code', '3': 4, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'category_id', '3': 6, '4': 1, '5': 9, '10': 'categoryId'},
    {
      '1': 'inventory_warehouse_id',
      '3': 7,
      '4': 1,
      '5': 9,
      '10': 'inventoryWarehouseId'
    },
    {
      '1': 'picking_warehouse_id',
      '3': 8,
      '4': 1,
      '5': 9,
      '10': 'pickingWarehouseId'
    },
    {'1': 'description', '3': 9, '4': 1, '5': 9, '10': 'description'},
    {'1': 'is_active', '3': 10, '4': 1, '5': 8, '10': 'isActive'},
    {
      '1': 'units',
      '3': 11,
      '4': 3,
      '5': 11,
      '6': '.products.v1.ProductUnit',
      '10': 'units'
    },
    {
      '1': 'processing_specs',
      '3': 12,
      '4': 3,
      '5': 11,
      '6': '.products.v1.ProductProcessingSpec',
      '10': 'processingSpecs'
    },
    {'1': 'created_at', '3': 13, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 14, '4': 1, '5': 9, '10': 'updatedAt'},
    {'1': 'deleted_at', '3': 15, '4': 1, '5': 9, '10': 'deletedAt'},
  ],
};

/// Descriptor for `Product`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List productDescriptor = $convert.base64Decode(
    'CgdQcm9kdWN0Eg4KAmlkGAEgASgJUgJpZBIdCgpjb21wYW55X2lkGAIgASgJUgljb21wYW55SW'
    'QSIwoNZGVwYXJ0bWVudF9pZBgDIAEoCVIMZGVwYXJ0bWVudElkEhIKBGNvZGUYBCABKAlSBGNv'
    'ZGUSEgoEbmFtZRgFIAEoCVIEbmFtZRIfCgtjYXRlZ29yeV9pZBgGIAEoCVIKY2F0ZWdvcnlJZB'
    'I0ChZpbnZlbnRvcnlfd2FyZWhvdXNlX2lkGAcgASgJUhRpbnZlbnRvcnlXYXJlaG91c2VJZBIw'
    'ChRwaWNraW5nX3dhcmVob3VzZV9pZBgIIAEoCVIScGlja2luZ1dhcmVob3VzZUlkEiAKC2Rlc2'
    'NyaXB0aW9uGAkgASgJUgtkZXNjcmlwdGlvbhIbCglpc19hY3RpdmUYCiABKAhSCGlzQWN0aXZl'
    'Ei4KBXVuaXRzGAsgAygLMhgucHJvZHVjdHMudjEuUHJvZHVjdFVuaXRSBXVuaXRzEk0KEHByb2'
    'Nlc3Npbmdfc3BlY3MYDCADKAsyIi5wcm9kdWN0cy52MS5Qcm9kdWN0UHJvY2Vzc2luZ1NwZWNS'
    'D3Byb2Nlc3NpbmdTcGVjcxIdCgpjcmVhdGVkX2F0GA0gASgJUgljcmVhdGVkQXQSHQoKdXBkYX'
    'RlZF9hdBgOIAEoCVIJdXBkYXRlZEF0Eh0KCmRlbGV0ZWRfYXQYDyABKAlSCWRlbGV0ZWRBdA==');

@$core.Deprecated('Use listProductsRequestDescriptor instead')
const ListProductsRequest$json = {
  '1': 'ListProductsRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'keyword', '3': 3, '4': 1, '5': 9, '10': 'keyword'},
    {'1': 'category_id', '3': 4, '4': 1, '5': 9, '10': 'categoryId'},
    {'1': 'include_deleted', '3': 5, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListProductsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProductsRequestDescriptor = $convert.base64Decode(
    'ChNMaXN0UHJvZHVjdHNSZXF1ZXN0EhIKBHBhZ2UYASABKAVSBHBhZ2USGwoJcGFnZV9zaXplGA'
    'IgASgFUghwYWdlU2l6ZRIYCgdrZXl3b3JkGAMgASgJUgdrZXl3b3JkEh8KC2NhdGVnb3J5X2lk'
    'GAQgASgJUgpjYXRlZ29yeUlkEicKD2luY2x1ZGVfZGVsZXRlZBgFIAEoCFIOaW5jbHVkZURlbG'
    'V0ZWQ=');

@$core.Deprecated('Use listProductsResponseDescriptor instead')
const ListProductsResponse$json = {
  '1': 'ListProductsResponse',
  '2': [
    {
      '1': 'products',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.products.v1.Product',
      '10': 'products'
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

/// Descriptor for `ListProductsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProductsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0UHJvZHVjdHNSZXNwb25zZRIwCghwcm9kdWN0cxgBIAMoCzIULnByb2R1Y3RzLnYxLl'
    'Byb2R1Y3RSCHByb2R1Y3RzEjkKCnBhZ2luYXRpb24YAiABKAsyGS5zYWxlc29yZGVyLnYxLlBh'
    'Z2luYXRpb25SCnBhZ2luYXRpb24=');

@$core.Deprecated('Use getProductRequestDescriptor instead')
const GetProductRequest$json = {
  '1': 'GetProductRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getProductRequestDescriptor =
    $convert.base64Decode('ChFHZXRQcm9kdWN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use getProductResponseDescriptor instead')
const GetProductResponse$json = {
  '1': 'GetProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.Product',
      '10': 'product'
    },
  ],
};

/// Descriptor for `GetProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getProductResponseDescriptor = $convert.base64Decode(
    'ChJHZXRQcm9kdWN0UmVzcG9uc2USLgoHcHJvZHVjdBgBIAEoCzIULnByb2R1Y3RzLnYxLlByb2'
    'R1Y3RSB3Byb2R1Y3Q=');

@$core.Deprecated('Use createProductRequestDescriptor instead')
const CreateProductRequest$json = {
  '1': 'CreateProductRequest',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'category_id', '3': 3, '4': 1, '5': 9, '10': 'categoryId'},
    {
      '1': 'inventory_warehouse_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'inventoryWarehouseId'
    },
    {
      '1': 'picking_warehouse_id',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'pickingWarehouseId'
    },
    {'1': 'description', '3': 6, '4': 1, '5': 9, '10': 'description'},
    {'1': 'is_active', '3': 7, '4': 1, '5': 8, '10': 'isActive'},
    {
      '1': 'units',
      '3': 8,
      '4': 3,
      '5': 11,
      '6': '.products.v1.ProductUnit',
      '10': 'units'
    },
    {
      '1': 'processing_specs',
      '3': 9,
      '4': 3,
      '5': 11,
      '6': '.products.v1.ProductProcessingSpec',
      '10': 'processingSpecs'
    },
  ],
};

/// Descriptor for `CreateProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createProductRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVQcm9kdWN0UmVxdWVzdBISCgRjb2RlGAEgASgJUgRjb2RlEhIKBG5hbWUYAiABKA'
    'lSBG5hbWUSHwoLY2F0ZWdvcnlfaWQYAyABKAlSCmNhdGVnb3J5SWQSNAoWaW52ZW50b3J5X3dh'
    'cmVob3VzZV9pZBgEIAEoCVIUaW52ZW50b3J5V2FyZWhvdXNlSWQSMAoUcGlja2luZ193YXJlaG'
    '91c2VfaWQYBSABKAlSEnBpY2tpbmdXYXJlaG91c2VJZBIgCgtkZXNjcmlwdGlvbhgGIAEoCVIL'
    'ZGVzY3JpcHRpb24SGwoJaXNfYWN0aXZlGAcgASgIUghpc0FjdGl2ZRIuCgV1bml0cxgIIAMoCz'
    'IYLnByb2R1Y3RzLnYxLlByb2R1Y3RVbml0UgV1bml0cxJNChBwcm9jZXNzaW5nX3NwZWNzGAkg'
    'AygLMiIucHJvZHVjdHMudjEuUHJvZHVjdFByb2Nlc3NpbmdTcGVjUg9wcm9jZXNzaW5nU3BlY3'
    'M=');

@$core.Deprecated('Use createProductResponseDescriptor instead')
const CreateProductResponse$json = {
  '1': 'CreateProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.Product',
      '10': 'product'
    },
  ],
};

/// Descriptor for `CreateProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createProductResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVQcm9kdWN0UmVzcG9uc2USLgoHcHJvZHVjdBgBIAEoCzIULnByb2R1Y3RzLnYxLl'
    'Byb2R1Y3RSB3Byb2R1Y3Q=');

@$core.Deprecated('Use updateProductRequestDescriptor instead')
const UpdateProductRequest$json = {
  '1': 'UpdateProductRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'code', '17': true},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'name', '17': true},
    {
      '1': 'category_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '9': 2,
      '10': 'categoryId',
      '17': true
    },
    {
      '1': 'inventory_warehouse_id',
      '3': 5,
      '4': 1,
      '5': 9,
      '9': 3,
      '10': 'inventoryWarehouseId',
      '17': true
    },
    {
      '1': 'picking_warehouse_id',
      '3': 6,
      '4': 1,
      '5': 9,
      '9': 4,
      '10': 'pickingWarehouseId',
      '17': true
    },
    {
      '1': 'description',
      '3': 7,
      '4': 1,
      '5': 9,
      '9': 5,
      '10': 'description',
      '17': true
    },
    {
      '1': 'is_active',
      '3': 8,
      '4': 1,
      '5': 8,
      '9': 6,
      '10': 'isActive',
      '17': true
    },
    {
      '1': 'units',
      '3': 9,
      '4': 3,
      '5': 11,
      '6': '.products.v1.ProductUnit',
      '10': 'units'
    },
    {
      '1': 'processing_specs',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.products.v1.ProductProcessingSpec',
      '10': 'processingSpecs'
    },
  ],
  '8': [
    {'1': '_code'},
    {'1': '_name'},
    {'1': '_category_id'},
    {'1': '_inventory_warehouse_id'},
    {'1': '_picking_warehouse_id'},
    {'1': '_description'},
    {'1': '_is_active'},
  ],
};

/// Descriptor for `UpdateProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProductRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVQcm9kdWN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSFwoEY29kZRgCIAEoCUgAUg'
    'Rjb2RliAEBEhcKBG5hbWUYAyABKAlIAVIEbmFtZYgBARIkCgtjYXRlZ29yeV9pZBgEIAEoCUgC'
    'UgpjYXRlZ29yeUlkiAEBEjkKFmludmVudG9yeV93YXJlaG91c2VfaWQYBSABKAlIA1IUaW52ZW'
    '50b3J5V2FyZWhvdXNlSWSIAQESNQoUcGlja2luZ193YXJlaG91c2VfaWQYBiABKAlIBFIScGlj'
    'a2luZ1dhcmVob3VzZUlkiAEBEiUKC2Rlc2NyaXB0aW9uGAcgASgJSAVSC2Rlc2NyaXB0aW9uiA'
    'EBEiAKCWlzX2FjdGl2ZRgIIAEoCEgGUghpc0FjdGl2ZYgBARIuCgV1bml0cxgJIAMoCzIYLnBy'
    'b2R1Y3RzLnYxLlByb2R1Y3RVbml0UgV1bml0cxJNChBwcm9jZXNzaW5nX3NwZWNzGAogAygLMi'
    'IucHJvZHVjdHMudjEuUHJvZHVjdFByb2Nlc3NpbmdTcGVjUg9wcm9jZXNzaW5nU3BlY3NCBwoF'
    'X2NvZGVCBwoFX25hbWVCDgoMX2NhdGVnb3J5X2lkQhkKF19pbnZlbnRvcnlfd2FyZWhvdXNlX2'
    'lkQhcKFV9waWNraW5nX3dhcmVob3VzZV9pZEIOCgxfZGVzY3JpcHRpb25CDAoKX2lzX2FjdGl2'
    'ZQ==');

@$core.Deprecated('Use updateProductResponseDescriptor instead')
const UpdateProductResponse$json = {
  '1': 'UpdateProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.Product',
      '10': 'product'
    },
  ],
};

/// Descriptor for `UpdateProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProductResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVQcm9kdWN0UmVzcG9uc2USLgoHcHJvZHVjdBgBIAEoCzIULnByb2R1Y3RzLnYxLl'
    'Byb2R1Y3RSB3Byb2R1Y3Q=');

@$core.Deprecated('Use deleteProductRequestDescriptor instead')
const DeleteProductRequest$json = {
  '1': 'DeleteProductRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteProductRequestDescriptor = $convert
    .base64Decode('ChREZWxldGVQcm9kdWN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use deleteProductResponseDescriptor instead')
const DeleteProductResponse$json = {
  '1': 'DeleteProductResponse',
};

/// Descriptor for `DeleteProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteProductResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVQcm9kdWN0UmVzcG9uc2U=');

@$core.Deprecated('Use restoreProductRequestDescriptor instead')
const RestoreProductRequest$json = {
  '1': 'RestoreProductRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `RestoreProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreProductRequestDescriptor = $convert
    .base64Decode('ChVSZXN0b3JlUHJvZHVjdFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use restoreProductResponseDescriptor instead')
const RestoreProductResponse$json = {
  '1': 'RestoreProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.Product',
      '10': 'product'
    },
  ],
};

/// Descriptor for `RestoreProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreProductResponseDescriptor =
    $convert.base64Decode(
        'ChZSZXN0b3JlUHJvZHVjdFJlc3BvbnNlEi4KB3Byb2R1Y3QYASABKAsyFC5wcm9kdWN0cy52MS'
        '5Qcm9kdWN0Ugdwcm9kdWN0');

@$core.Deprecated('Use customerProductDescriptor instead')
const CustomerProduct$json = {
  '1': 'CustomerProduct',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'customer_id', '3': 2, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'product_id', '3': 3, '4': 1, '5': 9, '10': 'productId'},
    {'1': 'alias_name', '3': 4, '4': 1, '5': 9, '10': 'aliasName'},
    {'1': 'default_qty', '3': 5, '4': 1, '5': 9, '10': 'defaultQty'},
    {'1': 'cut_note', '3': 6, '4': 1, '5': 9, '10': 'cutNote'},
    {'1': 'promo_tag_ids', '3': 7, '4': 3, '5': 3, '10': 'promoTagIds'},
    {'1': 'created_at', '3': 8, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 9, '4': 1, '5': 9, '10': 'updatedAt'},
  ],
};

/// Descriptor for `CustomerProduct`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List customerProductDescriptor = $convert.base64Decode(
    'Cg9DdXN0b21lclByb2R1Y3QSDgoCaWQYASABKAlSAmlkEh8KC2N1c3RvbWVyX2lkGAIgASgJUg'
    'pjdXN0b21lcklkEh0KCnByb2R1Y3RfaWQYAyABKAlSCXByb2R1Y3RJZBIdCgphbGlhc19uYW1l'
    'GAQgASgJUglhbGlhc05hbWUSHwoLZGVmYXVsdF9xdHkYBSABKAlSCmRlZmF1bHRRdHkSGQoIY3'
    'V0X25vdGUYBiABKAlSB2N1dE5vdGUSIgoNcHJvbW9fdGFnX2lkcxgHIAMoA1ILcHJvbW9UYWdJ'
    'ZHMSHQoKY3JlYXRlZF9hdBgIIAEoCVIJY3JlYXRlZEF0Eh0KCnVwZGF0ZWRfYXQYCSABKAlSCX'
    'VwZGF0ZWRBdA==');

@$core.Deprecated('Use listCustomerProductsRequestDescriptor instead')
const ListCustomerProductsRequest$json = {
  '1': 'ListCustomerProductsRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'for_order', '3': 2, '4': 1, '5': 8, '10': 'forOrder'},
    {'1': 'include_deleted', '3': 3, '4': 1, '5': 8, '10': 'includeDeleted'},
  ],
};

/// Descriptor for `ListCustomerProductsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCustomerProductsRequestDescriptor =
    $convert.base64Decode(
        'ChtMaXN0Q3VzdG9tZXJQcm9kdWN0c1JlcXVlc3QSHwoLY3VzdG9tZXJfaWQYASABKAlSCmN1c3'
        'RvbWVySWQSGwoJZm9yX29yZGVyGAIgASgIUghmb3JPcmRlchInCg9pbmNsdWRlX2RlbGV0ZWQY'
        'AyABKAhSDmluY2x1ZGVEZWxldGVk');

@$core.Deprecated('Use listCustomerProductsResponseDescriptor instead')
const ListCustomerProductsResponse$json = {
  '1': 'ListCustomerProductsResponse',
  '2': [
    {
      '1': 'products',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.products.v1.CustomerProduct',
      '10': 'products'
    },
    {'1': 'total', '3': 2, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListCustomerProductsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCustomerProductsResponseDescriptor =
    $convert.base64Decode(
        'ChxMaXN0Q3VzdG9tZXJQcm9kdWN0c1Jlc3BvbnNlEjgKCHByb2R1Y3RzGAEgAygLMhwucHJvZH'
        'VjdHMudjEuQ3VzdG9tZXJQcm9kdWN0Ughwcm9kdWN0cxIUCgV0b3RhbBgCIAEoBVIFdG90YWw=');

@$core.Deprecated('Use addCustomerProductRequestDescriptor instead')
const AddCustomerProductRequest$json = {
  '1': 'AddCustomerProductRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'product_id', '3': 2, '4': 1, '5': 9, '10': 'productId'},
    {'1': 'alias_name', '3': 3, '4': 1, '5': 9, '10': 'aliasName'},
    {'1': 'default_qty', '3': 4, '4': 1, '5': 9, '10': 'defaultQty'},
    {'1': 'cut_note', '3': 5, '4': 1, '5': 9, '10': 'cutNote'},
  ],
};

/// Descriptor for `AddCustomerProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addCustomerProductRequestDescriptor = $convert.base64Decode(
    'ChlBZGRDdXN0b21lclByb2R1Y3RSZXF1ZXN0Eh8KC2N1c3RvbWVyX2lkGAEgASgJUgpjdXN0b2'
    '1lcklkEh0KCnByb2R1Y3RfaWQYAiABKAlSCXByb2R1Y3RJZBIdCgphbGlhc19uYW1lGAMgASgJ'
    'UglhbGlhc05hbWUSHwoLZGVmYXVsdF9xdHkYBCABKAlSCmRlZmF1bHRRdHkSGQoIY3V0X25vdG'
    'UYBSABKAlSB2N1dE5vdGU=');

@$core.Deprecated('Use addCustomerProductResponseDescriptor instead')
const AddCustomerProductResponse$json = {
  '1': 'AddCustomerProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.CustomerProduct',
      '10': 'product'
    },
  ],
};

/// Descriptor for `AddCustomerProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addCustomerProductResponseDescriptor =
    $convert.base64Decode(
        'ChpBZGRDdXN0b21lclByb2R1Y3RSZXNwb25zZRI2Cgdwcm9kdWN0GAEgASgLMhwucHJvZHVjdH'
        'MudjEuQ3VzdG9tZXJQcm9kdWN0Ugdwcm9kdWN0');

@$core.Deprecated('Use updateCustomerProductRequestDescriptor instead')
const UpdateCustomerProductRequest$json = {
  '1': 'UpdateCustomerProductRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'alias_name', '3': 2, '4': 1, '5': 9, '10': 'aliasName'},
    {'1': 'default_qty', '3': 3, '4': 1, '5': 9, '10': 'defaultQty'},
    {'1': 'cut_note', '3': 4, '4': 1, '5': 9, '10': 'cutNote'},
  ],
};

/// Descriptor for `UpdateCustomerProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCustomerProductRequestDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVDdXN0b21lclByb2R1Y3RSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIdCgphbGlhc1'
        '9uYW1lGAIgASgJUglhbGlhc05hbWUSHwoLZGVmYXVsdF9xdHkYAyABKAlSCmRlZmF1bHRRdHkS'
        'GQoIY3V0X25vdGUYBCABKAlSB2N1dE5vdGU=');

@$core.Deprecated('Use updateCustomerProductResponseDescriptor instead')
const UpdateCustomerProductResponse$json = {
  '1': 'UpdateCustomerProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.CustomerProduct',
      '10': 'product'
    },
  ],
};

/// Descriptor for `UpdateCustomerProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCustomerProductResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVDdXN0b21lclByb2R1Y3RSZXNwb25zZRI2Cgdwcm9kdWN0GAEgASgLMhwucHJvZH'
        'VjdHMudjEuQ3VzdG9tZXJQcm9kdWN0Ugdwcm9kdWN0');

@$core.Deprecated('Use deleteCustomerProductRequestDescriptor instead')
const DeleteCustomerProductRequest$json = {
  '1': 'DeleteCustomerProductRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteCustomerProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCustomerProductRequestDescriptor =
    $convert.base64Decode(
        'ChxEZWxldGVDdXN0b21lclByb2R1Y3RSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use deleteCustomerProductResponseDescriptor instead')
const DeleteCustomerProductResponse$json = {
  '1': 'DeleteCustomerProductResponse',
};

/// Descriptor for `DeleteCustomerProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCustomerProductResponseDescriptor =
    $convert.base64Decode('Ch1EZWxldGVDdXN0b21lclByb2R1Y3RSZXNwb25zZQ==');

@$core.Deprecated('Use ensureCustomerProductRequestDescriptor instead')
const EnsureCustomerProductRequest$json = {
  '1': 'EnsureCustomerProductRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
    {'1': 'product_id', '3': 2, '4': 1, '5': 9, '10': 'productId'},
    {'1': 'alias_name', '3': 3, '4': 1, '5': 9, '10': 'aliasName'},
  ],
};

/// Descriptor for `EnsureCustomerProductRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ensureCustomerProductRequestDescriptor =
    $convert.base64Decode(
        'ChxFbnN1cmVDdXN0b21lclByb2R1Y3RSZXF1ZXN0Eh8KC2N1c3RvbWVyX2lkGAEgASgJUgpjdX'
        'N0b21lcklkEh0KCnByb2R1Y3RfaWQYAiABKAlSCXByb2R1Y3RJZBIdCgphbGlhc19uYW1lGAMg'
        'ASgJUglhbGlhc05hbWU=');

@$core.Deprecated('Use ensureCustomerProductResponseDescriptor instead')
const EnsureCustomerProductResponse$json = {
  '1': 'EnsureCustomerProductResponse',
  '2': [
    {
      '1': 'product',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.products.v1.CustomerProduct',
      '10': 'product'
    },
    {'1': 'created', '3': 2, '4': 1, '5': 8, '10': 'created'},
  ],
};

/// Descriptor for `EnsureCustomerProductResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ensureCustomerProductResponseDescriptor =
    $convert.base64Decode(
        'Ch1FbnN1cmVDdXN0b21lclByb2R1Y3RSZXNwb25zZRI2Cgdwcm9kdWN0GAEgASgLMhwucHJvZH'
        'VjdHMudjEuQ3VzdG9tZXJQcm9kdWN0Ugdwcm9kdWN0EhgKB2NyZWF0ZWQYAiABKAhSB2NyZWF0'
        'ZWQ=');

const $core.Map<$core.String, $core.dynamic> ProductServiceBase$json = {
  '1': 'ProductService',
  '2': [
    {
      '1': 'ListProducts',
      '2': '.products.v1.ListProductsRequest',
      '3': '.products.v1.ListProductsResponse'
    },
    {
      '1': 'GetProduct',
      '2': '.products.v1.GetProductRequest',
      '3': '.products.v1.GetProductResponse'
    },
    {
      '1': 'CreateProduct',
      '2': '.products.v1.CreateProductRequest',
      '3': '.products.v1.CreateProductResponse'
    },
    {
      '1': 'UpdateProduct',
      '2': '.products.v1.UpdateProductRequest',
      '3': '.products.v1.UpdateProductResponse'
    },
    {
      '1': 'DeleteProduct',
      '2': '.products.v1.DeleteProductRequest',
      '3': '.products.v1.DeleteProductResponse'
    },
    {
      '1': 'RestoreProduct',
      '2': '.products.v1.RestoreProductRequest',
      '3': '.products.v1.RestoreProductResponse'
    },
  ],
};

@$core.Deprecated('Use productServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ProductServiceBase$messageJson = {
  '.products.v1.ListProductsRequest': ListProductsRequest$json,
  '.products.v1.ListProductsResponse': ListProductsResponse$json,
  '.products.v1.Product': Product$json,
  '.products.v1.ProductUnit': ProductUnit$json,
  '.products.v1.ProductProcessingSpec': ProductProcessingSpec$json,
  '.google.protobuf.Struct': $0.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $0.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $0.Value$json,
  '.google.protobuf.ListValue': $0.ListValue$json,
  '.salesorder.v1.Pagination': $1.Pagination$json,
  '.products.v1.GetProductRequest': GetProductRequest$json,
  '.products.v1.GetProductResponse': GetProductResponse$json,
  '.products.v1.CreateProductRequest': CreateProductRequest$json,
  '.products.v1.CreateProductResponse': CreateProductResponse$json,
  '.products.v1.UpdateProductRequest': UpdateProductRequest$json,
  '.products.v1.UpdateProductResponse': UpdateProductResponse$json,
  '.products.v1.DeleteProductRequest': DeleteProductRequest$json,
  '.products.v1.DeleteProductResponse': DeleteProductResponse$json,
  '.products.v1.RestoreProductRequest': RestoreProductRequest$json,
  '.products.v1.RestoreProductResponse': RestoreProductResponse$json,
};

/// Descriptor for `ProductService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List productServiceDescriptor = $convert.base64Decode(
    'Cg5Qcm9kdWN0U2VydmljZRJTCgxMaXN0UHJvZHVjdHMSIC5wcm9kdWN0cy52MS5MaXN0UHJvZH'
    'VjdHNSZXF1ZXN0GiEucHJvZHVjdHMudjEuTGlzdFByb2R1Y3RzUmVzcG9uc2USTQoKR2V0UHJv'
    'ZHVjdBIeLnByb2R1Y3RzLnYxLkdldFByb2R1Y3RSZXF1ZXN0Gh8ucHJvZHVjdHMudjEuR2V0UH'
    'JvZHVjdFJlc3BvbnNlElYKDUNyZWF0ZVByb2R1Y3QSIS5wcm9kdWN0cy52MS5DcmVhdGVQcm9k'
    'dWN0UmVxdWVzdBoiLnByb2R1Y3RzLnYxLkNyZWF0ZVByb2R1Y3RSZXNwb25zZRJWCg1VcGRhdG'
    'VQcm9kdWN0EiEucHJvZHVjdHMudjEuVXBkYXRlUHJvZHVjdFJlcXVlc3QaIi5wcm9kdWN0cy52'
    'MS5VcGRhdGVQcm9kdWN0UmVzcG9uc2USVgoNRGVsZXRlUHJvZHVjdBIhLnByb2R1Y3RzLnYxLk'
    'RlbGV0ZVByb2R1Y3RSZXF1ZXN0GiIucHJvZHVjdHMudjEuRGVsZXRlUHJvZHVjdFJlc3BvbnNl'
    'ElkKDlJlc3RvcmVQcm9kdWN0EiIucHJvZHVjdHMudjEuUmVzdG9yZVByb2R1Y3RSZXF1ZXN0Gi'
    'MucHJvZHVjdHMudjEuUmVzdG9yZVByb2R1Y3RSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> CustomerProductServiceBase$json = {
  '1': 'CustomerProductService',
  '2': [
    {
      '1': 'ListCustomerProducts',
      '2': '.products.v1.ListCustomerProductsRequest',
      '3': '.products.v1.ListCustomerProductsResponse'
    },
    {
      '1': 'AddCustomerProduct',
      '2': '.products.v1.AddCustomerProductRequest',
      '3': '.products.v1.AddCustomerProductResponse'
    },
    {
      '1': 'UpdateCustomerProduct',
      '2': '.products.v1.UpdateCustomerProductRequest',
      '3': '.products.v1.UpdateCustomerProductResponse'
    },
    {
      '1': 'DeleteCustomerProduct',
      '2': '.products.v1.DeleteCustomerProductRequest',
      '3': '.products.v1.DeleteCustomerProductResponse'
    },
    {
      '1': 'EnsureCustomerProduct',
      '2': '.products.v1.EnsureCustomerProductRequest',
      '3': '.products.v1.EnsureCustomerProductResponse'
    },
  ],
};

@$core.Deprecated('Use customerProductServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    CustomerProductServiceBase$messageJson = {
  '.products.v1.ListCustomerProductsRequest': ListCustomerProductsRequest$json,
  '.products.v1.ListCustomerProductsResponse':
      ListCustomerProductsResponse$json,
  '.products.v1.CustomerProduct': CustomerProduct$json,
  '.products.v1.AddCustomerProductRequest': AddCustomerProductRequest$json,
  '.products.v1.AddCustomerProductResponse': AddCustomerProductResponse$json,
  '.products.v1.UpdateCustomerProductRequest':
      UpdateCustomerProductRequest$json,
  '.products.v1.UpdateCustomerProductResponse':
      UpdateCustomerProductResponse$json,
  '.products.v1.DeleteCustomerProductRequest':
      DeleteCustomerProductRequest$json,
  '.products.v1.DeleteCustomerProductResponse':
      DeleteCustomerProductResponse$json,
  '.products.v1.EnsureCustomerProductRequest':
      EnsureCustomerProductRequest$json,
  '.products.v1.EnsureCustomerProductResponse':
      EnsureCustomerProductResponse$json,
};

/// Descriptor for `CustomerProductService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List customerProductServiceDescriptor = $convert.base64Decode(
    'ChZDdXN0b21lclByb2R1Y3RTZXJ2aWNlEmsKFExpc3RDdXN0b21lclByb2R1Y3RzEigucHJvZH'
    'VjdHMudjEuTGlzdEN1c3RvbWVyUHJvZHVjdHNSZXF1ZXN0GikucHJvZHVjdHMudjEuTGlzdEN1'
    'c3RvbWVyUHJvZHVjdHNSZXNwb25zZRJlChJBZGRDdXN0b21lclByb2R1Y3QSJi5wcm9kdWN0cy'
    '52MS5BZGRDdXN0b21lclByb2R1Y3RSZXF1ZXN0GicucHJvZHVjdHMudjEuQWRkQ3VzdG9tZXJQ'
    'cm9kdWN0UmVzcG9uc2USbgoVVXBkYXRlQ3VzdG9tZXJQcm9kdWN0EikucHJvZHVjdHMudjEuVX'
    'BkYXRlQ3VzdG9tZXJQcm9kdWN0UmVxdWVzdBoqLnByb2R1Y3RzLnYxLlVwZGF0ZUN1c3RvbWVy'
    'UHJvZHVjdFJlc3BvbnNlEm4KFURlbGV0ZUN1c3RvbWVyUHJvZHVjdBIpLnByb2R1Y3RzLnYxLk'
    'RlbGV0ZUN1c3RvbWVyUHJvZHVjdFJlcXVlc3QaKi5wcm9kdWN0cy52MS5EZWxldGVDdXN0b21l'
    'clByb2R1Y3RSZXNwb25zZRJuChVFbnN1cmVDdXN0b21lclByb2R1Y3QSKS5wcm9kdWN0cy52MS'
    '5FbnN1cmVDdXN0b21lclByb2R1Y3RSZXF1ZXN0GioucHJvZHVjdHMudjEuRW5zdXJlQ3VzdG9t'
    'ZXJQcm9kdWN0UmVzcG9uc2U=');
