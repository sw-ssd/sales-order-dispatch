// This is a generated file - do not edit.
//
// Generated from products/v1/product.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $0;

import '../../salesorder/v1/common.pb.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// ---- 子訊息 ----
class ProductUnit extends $pb.GeneratedMessage {
  factory ProductUnit({
    $core.String? unitCode,
    $core.String? conversionRate,
    $core.bool? isBase,
    $core.int? sortOrder,
    $core.String? sizeDesc,
  }) {
    final result = ProductUnit._();
    if (unitCode != null) result.unitCode = unitCode;
    if (conversionRate != null) result.conversionRate = conversionRate;
    if (isBase != null) result.isBase = isBase;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (sizeDesc != null) result.sizeDesc = sizeDesc;
    return result;
  }

  ProductUnit._();

  factory ProductUnit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProductUnit()..mergeFromBuffer(data, registry);
  factory ProductUnit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProductUnit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProductUnit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ProductUnit.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'unitCode')
    ..aOS(2, _omitFieldNames ? '' : 'conversionRate')
    ..aOB(3, _omitFieldNames ? '' : 'isBase')
    ..aI(4, _omitFieldNames ? '' : 'sortOrder')
    ..aOS(5, _omitFieldNames ? '' : 'sizeDesc')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProductUnit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProductUnit copyWith(void Function(ProductUnit) updates) =>
      super.copyWith((message) => updates(message as ProductUnit))
          as ProductUnit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProductUnit() / ProductUnit.new instead')
  static ProductUnit create() => ProductUnit._();
  static $pb.GeneratedMessage $_createMessage() => ProductUnit._();
  @$core.override
  ProductUnit createEmptyInstance() => ProductUnit._();
  @$core.pragma('dart2js:noInline')
  static ProductUnit getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProductUnit>(
          ProductUnit.$_createMessage);
  static ProductUnit? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get unitCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set unitCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUnitCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearUnitCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get conversionRate => $_getSZ(1);
  @$pb.TagNumber(2)
  set conversionRate($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasConversionRate() => $_has(1);
  @$pb.TagNumber(2)
  void clearConversionRate() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get isBase => $_getBF(2);
  @$pb.TagNumber(3)
  set isBase($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasIsBase() => $_has(2);
  @$pb.TagNumber(3)
  void clearIsBase() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get sortOrder => $_getIZ(3);
  @$pb.TagNumber(4)
  set sortOrder($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSortOrder() => $_has(3);
  @$pb.TagNumber(4)
  void clearSortOrder() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get sizeDesc => $_getSZ(4);
  @$pb.TagNumber(5)
  set sizeDesc($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSizeDesc() => $_has(4);
  @$pb.TagNumber(5)
  void clearSizeDesc() => $_clearField(5);
}

class ProductProcessingSpec extends $pb.GeneratedMessage {
  factory ProductProcessingSpec({
    $core.String? processingSpecId,
    $0.Struct? attributes,
  }) {
    final result = ProductProcessingSpec._();
    if (processingSpecId != null) result.processingSpecId = processingSpecId;
    if (attributes != null) result.attributes = attributes;
    return result;
  }

  ProductProcessingSpec._();

  factory ProductProcessingSpec.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProductProcessingSpec()..mergeFromBuffer(data, registry);
  factory ProductProcessingSpec.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProductProcessingSpec()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProductProcessingSpec',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ProductProcessingSpec.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'processingSpecId')
    ..aOM<$0.Struct>(2, _omitFieldNames ? '' : 'attributes',
        subBuilder: $0.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProductProcessingSpec clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProductProcessingSpec copyWith(
          void Function(ProductProcessingSpec) updates) =>
      super.copyWith((message) => updates(message as ProductProcessingSpec))
          as ProductProcessingSpec;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ProductProcessingSpec() / ProductProcessingSpec.new instead')
  static ProductProcessingSpec create() => ProductProcessingSpec._();
  static $pb.GeneratedMessage $_createMessage() => ProductProcessingSpec._();
  @$core.override
  ProductProcessingSpec createEmptyInstance() => ProductProcessingSpec._();
  @$core.pragma('dart2js:noInline')
  static ProductProcessingSpec getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProductProcessingSpec>(
          ProductProcessingSpec.$_createMessage);
  static ProductProcessingSpec? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get processingSpecId => $_getSZ(0);
  @$pb.TagNumber(1)
  set processingSpecId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProcessingSpecId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProcessingSpecId() => $_clearField(1);

  @$pb.TagNumber(2)
  $0.Struct get attributes => $_getN(1);
  @$pb.TagNumber(2)
  set attributes($0.Struct value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasAttributes() => $_has(1);
  @$pb.TagNumber(2)
  void clearAttributes() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Struct ensureAttributes() => $_ensure(1);
}

/// ---- Product ----
class Product extends $pb.GeneratedMessage {
  factory Product({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? code,
    $core.String? name,
    $core.String? categoryId,
    $core.String? inventoryWarehouseId,
    $core.String? pickingWarehouseId,
    $core.String? description,
    $core.bool? isActive,
    $core.Iterable<ProductUnit>? units,
    $core.Iterable<ProductProcessingSpec>? processingSpecs,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = Product._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (categoryId != null) result.categoryId = categoryId;
    if (inventoryWarehouseId != null)
      result.inventoryWarehouseId = inventoryWarehouseId;
    if (pickingWarehouseId != null)
      result.pickingWarehouseId = pickingWarehouseId;
    if (description != null) result.description = description;
    if (isActive != null) result.isActive = isActive;
    if (units != null) result.units.addAll(units);
    if (processingSpecs != null) result.processingSpecs.addAll(processingSpecs);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  Product._();

  factory Product.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Product()..mergeFromBuffer(data, registry);
  factory Product.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Product()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Product',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: Product.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'code')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'categoryId')
    ..aOS(7, _omitFieldNames ? '' : 'inventoryWarehouseId')
    ..aOS(8, _omitFieldNames ? '' : 'pickingWarehouseId')
    ..aOS(9, _omitFieldNames ? '' : 'description')
    ..aOB(10, _omitFieldNames ? '' : 'isActive')
    ..pPM<ProductUnit>(11, _omitFieldNames ? '' : 'units',
        subBuilder: ProductUnit.$_createMessage)
    ..pPM<ProductProcessingSpec>(12, _omitFieldNames ? '' : 'processingSpecs',
        subBuilder: ProductProcessingSpec.$_createMessage)
    ..aOS(13, _omitFieldNames ? '' : 'createdAt')
    ..aOS(14, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(15, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Product clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Product copyWith(void Function(Product) updates) =>
      super.copyWith((message) => updates(message as Product)) as Product;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Product() / Product.new instead')
  static Product create() => Product._();
  static $pb.GeneratedMessage $_createMessage() => Product._();
  @$core.override
  Product createEmptyInstance() => Product._();
  @$core.pragma('dart2js:noInline')
  static Product getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Product>(Product.$_createMessage);
  static Product? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get companyId => $_getSZ(1);
  @$pb.TagNumber(2)
  set companyId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCompanyId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompanyId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get departmentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set departmentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDepartmentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearDepartmentId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get code => $_getSZ(3);
  @$pb.TagNumber(4)
  set code($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearCode() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get name => $_getSZ(4);
  @$pb.TagNumber(5)
  set name($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasName() => $_has(4);
  @$pb.TagNumber(5)
  void clearName() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get categoryId => $_getSZ(5);
  @$pb.TagNumber(6)
  set categoryId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCategoryId() => $_has(5);
  @$pb.TagNumber(6)
  void clearCategoryId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get inventoryWarehouseId => $_getSZ(6);
  @$pb.TagNumber(7)
  set inventoryWarehouseId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInventoryWarehouseId() => $_has(6);
  @$pb.TagNumber(7)
  void clearInventoryWarehouseId() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get pickingWarehouseId => $_getSZ(7);
  @$pb.TagNumber(8)
  set pickingWarehouseId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasPickingWarehouseId() => $_has(7);
  @$pb.TagNumber(8)
  void clearPickingWarehouseId() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get description => $_getSZ(8);
  @$pb.TagNumber(9)
  set description($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasDescription() => $_has(8);
  @$pb.TagNumber(9)
  void clearDescription() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.bool get isActive => $_getBF(9);
  @$pb.TagNumber(10)
  set isActive($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasIsActive() => $_has(9);
  @$pb.TagNumber(10)
  void clearIsActive() => $_clearField(10);

  @$pb.TagNumber(11)
  $pb.PbList<ProductUnit> get units => $_getList(10);

  @$pb.TagNumber(12)
  $pb.PbList<ProductProcessingSpec> get processingSpecs => $_getList(11);

  @$pb.TagNumber(13)
  $core.String get createdAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set createdAt($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasCreatedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearCreatedAt() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.String get updatedAt => $_getSZ(13);
  @$pb.TagNumber(14)
  set updatedAt($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasUpdatedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearUpdatedAt() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.String get deletedAt => $_getSZ(14);
  @$pb.TagNumber(15)
  set deletedAt($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasDeletedAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearDeletedAt() => $_clearField(15);
}

class ListProductsRequest extends $pb.GeneratedMessage {
  factory ListProductsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.String? categoryId,
    $core.bool? includeDeleted,
  }) {
    final result = ListProductsRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (categoryId != null) result.categoryId = categoryId;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListProductsRequest._();

  factory ListProductsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductsRequest()..mergeFromBuffer(data, registry);
  factory ListProductsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProductsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ListProductsRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOS(4, _omitFieldNames ? '' : 'categoryId')
    ..aOB(5, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductsRequest copyWith(void Function(ListProductsRequest) updates) =>
      super.copyWith((message) => updates(message as ListProductsRequest))
          as ListProductsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListProductsRequest() / ListProductsRequest.new instead')
  static ListProductsRequest create() => ListProductsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListProductsRequest._();
  @$core.override
  ListProductsRequest createEmptyInstance() => ListProductsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListProductsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProductsRequest>(
          ListProductsRequest.$_createMessage);
  static ListProductsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get page => $_getIZ(0);
  @$pb.TagNumber(1)
  set page($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPage() => $_has(0);
  @$pb.TagNumber(1)
  void clearPage() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get pageSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set pageSize($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageSize() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get keyword => $_getSZ(2);
  @$pb.TagNumber(3)
  set keyword($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasKeyword() => $_has(2);
  @$pb.TagNumber(3)
  void clearKeyword() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get categoryId => $_getSZ(3);
  @$pb.TagNumber(4)
  set categoryId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCategoryId() => $_has(3);
  @$pb.TagNumber(4)
  void clearCategoryId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get includeDeleted => $_getBF(4);
  @$pb.TagNumber(5)
  set includeDeleted($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIncludeDeleted() => $_has(4);
  @$pb.TagNumber(5)
  void clearIncludeDeleted() => $_clearField(5);
}

class ListProductsResponse extends $pb.GeneratedMessage {
  factory ListProductsResponse({
    $core.Iterable<Product>? products,
    $1.Pagination? pagination,
  }) {
    final result = ListProductsResponse._();
    if (products != null) result.products.addAll(products);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListProductsResponse._();

  factory ListProductsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductsResponse()..mergeFromBuffer(data, registry);
  factory ListProductsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProductsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ListProductsResponse.$_createMessage)
    ..pPM<Product>(1, _omitFieldNames ? '' : 'products',
        subBuilder: Product.$_createMessage)
    ..aOM<$1.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $1.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductsResponse copyWith(void Function(ListProductsResponse) updates) =>
      super.copyWith((message) => updates(message as ListProductsResponse))
          as ListProductsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListProductsResponse() / ListProductsResponse.new instead')
  static ListProductsResponse create() => ListProductsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListProductsResponse._();
  @$core.override
  ListProductsResponse createEmptyInstance() => ListProductsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListProductsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProductsResponse>(
          ListProductsResponse.$_createMessage);
  static ListProductsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Product> get products => $_getList(0);

  @$pb.TagNumber(2)
  $1.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($1.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $1.Pagination ensurePagination() => $_ensure(1);
}

class GetProductRequest extends $pb.GeneratedMessage {
  factory GetProductRequest({
    $core.String? id,
  }) {
    final result = GetProductRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetProductRequest._();

  factory GetProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetProductRequest()..mergeFromBuffer(data, registry);
  factory GetProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: GetProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProductRequest copyWith(void Function(GetProductRequest) updates) =>
      super.copyWith((message) => updates(message as GetProductRequest))
          as GetProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetProductRequest() / GetProductRequest.new instead')
  static GetProductRequest create() => GetProductRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetProductRequest._();
  @$core.override
  GetProductRequest createEmptyInstance() => GetProductRequest._();
  @$core.pragma('dart2js:noInline')
  static GetProductRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetProductRequest>(
          GetProductRequest.$_createMessage);
  static GetProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetProductResponse extends $pb.GeneratedMessage {
  factory GetProductResponse({
    Product? product,
  }) {
    final result = GetProductResponse._();
    if (product != null) result.product = product;
    return result;
  }

  GetProductResponse._();

  factory GetProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetProductResponse()..mergeFromBuffer(data, registry);
  factory GetProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: GetProductResponse.$_createMessage)
    ..aOM<Product>(1, _omitFieldNames ? '' : 'product',
        subBuilder: Product.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProductResponse copyWith(void Function(GetProductResponse) updates) =>
      super.copyWith((message) => updates(message as GetProductResponse))
          as GetProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetProductResponse() / GetProductResponse.new instead')
  static GetProductResponse create() => GetProductResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetProductResponse._();
  @$core.override
  GetProductResponse createEmptyInstance() => GetProductResponse._();
  @$core.pragma('dart2js:noInline')
  static GetProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetProductResponse>(
          GetProductResponse.$_createMessage);
  static GetProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Product get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(Product value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  Product ensureProduct() => $_ensure(0);
}

class CreateProductRequest extends $pb.GeneratedMessage {
  factory CreateProductRequest({
    $core.String? code,
    $core.String? name,
    $core.String? categoryId,
    $core.String? inventoryWarehouseId,
    $core.String? pickingWarehouseId,
    $core.String? description,
    $core.bool? isActive,
    $core.Iterable<ProductUnit>? units,
    $core.Iterable<ProductProcessingSpec>? processingSpecs,
  }) {
    final result = CreateProductRequest._();
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (categoryId != null) result.categoryId = categoryId;
    if (inventoryWarehouseId != null)
      result.inventoryWarehouseId = inventoryWarehouseId;
    if (pickingWarehouseId != null)
      result.pickingWarehouseId = pickingWarehouseId;
    if (description != null) result.description = description;
    if (isActive != null) result.isActive = isActive;
    if (units != null) result.units.addAll(units);
    if (processingSpecs != null) result.processingSpecs.addAll(processingSpecs);
    return result;
  }

  CreateProductRequest._();

  factory CreateProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductRequest()..mergeFromBuffer(data, registry);
  factory CreateProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: CreateProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'categoryId')
    ..aOS(4, _omitFieldNames ? '' : 'inventoryWarehouseId')
    ..aOS(5, _omitFieldNames ? '' : 'pickingWarehouseId')
    ..aOS(6, _omitFieldNames ? '' : 'description')
    ..aOB(7, _omitFieldNames ? '' : 'isActive')
    ..pPM<ProductUnit>(8, _omitFieldNames ? '' : 'units',
        subBuilder: ProductUnit.$_createMessage)
    ..pPM<ProductProcessingSpec>(9, _omitFieldNames ? '' : 'processingSpecs',
        subBuilder: ProductProcessingSpec.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductRequest copyWith(void Function(CreateProductRequest) updates) =>
      super.copyWith((message) => updates(message as CreateProductRequest))
          as CreateProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateProductRequest() / CreateProductRequest.new instead')
  static CreateProductRequest create() => CreateProductRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateProductRequest._();
  @$core.override
  CreateProductRequest createEmptyInstance() => CreateProductRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateProductRequest>(
          CreateProductRequest.$_createMessage);
  static CreateProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get code => $_getSZ(0);
  @$pb.TagNumber(1)
  set code($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get categoryId => $_getSZ(2);
  @$pb.TagNumber(3)
  set categoryId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCategoryId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCategoryId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get inventoryWarehouseId => $_getSZ(3);
  @$pb.TagNumber(4)
  set inventoryWarehouseId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasInventoryWarehouseId() => $_has(3);
  @$pb.TagNumber(4)
  void clearInventoryWarehouseId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get pickingWarehouseId => $_getSZ(4);
  @$pb.TagNumber(5)
  set pickingWarehouseId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPickingWarehouseId() => $_has(4);
  @$pb.TagNumber(5)
  void clearPickingWarehouseId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get description => $_getSZ(5);
  @$pb.TagNumber(6)
  set description($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasDescription() => $_has(5);
  @$pb.TagNumber(6)
  void clearDescription() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get isActive => $_getBF(6);
  @$pb.TagNumber(7)
  set isActive($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasIsActive() => $_has(6);
  @$pb.TagNumber(7)
  void clearIsActive() => $_clearField(7);

  @$pb.TagNumber(8)
  $pb.PbList<ProductUnit> get units => $_getList(7);

  @$pb.TagNumber(9)
  $pb.PbList<ProductProcessingSpec> get processingSpecs => $_getList(8);
}

class CreateProductResponse extends $pb.GeneratedMessage {
  factory CreateProductResponse({
    Product? product,
  }) {
    final result = CreateProductResponse._();
    if (product != null) result.product = product;
    return result;
  }

  CreateProductResponse._();

  factory CreateProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductResponse()..mergeFromBuffer(data, registry);
  factory CreateProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: CreateProductResponse.$_createMessage)
    ..aOM<Product>(1, _omitFieldNames ? '' : 'product',
        subBuilder: Product.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductResponse copyWith(
          void Function(CreateProductResponse) updates) =>
      super.copyWith((message) => updates(message as CreateProductResponse))
          as CreateProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateProductResponse() / CreateProductResponse.new instead')
  static CreateProductResponse create() => CreateProductResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateProductResponse._();
  @$core.override
  CreateProductResponse createEmptyInstance() => CreateProductResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateProductResponse>(
          CreateProductResponse.$_createMessage);
  static CreateProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Product get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(Product value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  Product ensureProduct() => $_ensure(0);
}

class UpdateProductRequest extends $pb.GeneratedMessage {
  factory UpdateProductRequest({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.String? categoryId,
    $core.String? inventoryWarehouseId,
    $core.String? pickingWarehouseId,
    $core.String? description,
    $core.bool? isActive,
    $core.Iterable<ProductUnit>? units,
    $core.Iterable<ProductProcessingSpec>? processingSpecs,
  }) {
    final result = UpdateProductRequest._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (categoryId != null) result.categoryId = categoryId;
    if (inventoryWarehouseId != null)
      result.inventoryWarehouseId = inventoryWarehouseId;
    if (pickingWarehouseId != null)
      result.pickingWarehouseId = pickingWarehouseId;
    if (description != null) result.description = description;
    if (isActive != null) result.isActive = isActive;
    if (units != null) result.units.addAll(units);
    if (processingSpecs != null) result.processingSpecs.addAll(processingSpecs);
    return result;
  }

  UpdateProductRequest._();

  factory UpdateProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductRequest()..mergeFromBuffer(data, registry);
  factory UpdateProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: UpdateProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'categoryId')
    ..aOS(5, _omitFieldNames ? '' : 'inventoryWarehouseId')
    ..aOS(6, _omitFieldNames ? '' : 'pickingWarehouseId')
    ..aOS(7, _omitFieldNames ? '' : 'description')
    ..aOB(8, _omitFieldNames ? '' : 'isActive')
    ..pPM<ProductUnit>(9, _omitFieldNames ? '' : 'units',
        subBuilder: ProductUnit.$_createMessage)
    ..pPM<ProductProcessingSpec>(10, _omitFieldNames ? '' : 'processingSpecs',
        subBuilder: ProductProcessingSpec.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductRequest copyWith(void Function(UpdateProductRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateProductRequest))
          as UpdateProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProductRequest() / UpdateProductRequest.new instead')
  static UpdateProductRequest create() => UpdateProductRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateProductRequest._();
  @$core.override
  UpdateProductRequest createEmptyInstance() => UpdateProductRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProductRequest>(
          UpdateProductRequest.$_createMessage);
  static UpdateProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get categoryId => $_getSZ(3);
  @$pb.TagNumber(4)
  set categoryId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCategoryId() => $_has(3);
  @$pb.TagNumber(4)
  void clearCategoryId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get inventoryWarehouseId => $_getSZ(4);
  @$pb.TagNumber(5)
  set inventoryWarehouseId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInventoryWarehouseId() => $_has(4);
  @$pb.TagNumber(5)
  void clearInventoryWarehouseId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get pickingWarehouseId => $_getSZ(5);
  @$pb.TagNumber(6)
  set pickingWarehouseId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPickingWarehouseId() => $_has(5);
  @$pb.TagNumber(6)
  void clearPickingWarehouseId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get description => $_getSZ(6);
  @$pb.TagNumber(7)
  set description($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasDescription() => $_has(6);
  @$pb.TagNumber(7)
  void clearDescription() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get isActive => $_getBF(7);
  @$pb.TagNumber(8)
  set isActive($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasIsActive() => $_has(7);
  @$pb.TagNumber(8)
  void clearIsActive() => $_clearField(8);

  @$pb.TagNumber(9)
  $pb.PbList<ProductUnit> get units => $_getList(8);

  @$pb.TagNumber(10)
  $pb.PbList<ProductProcessingSpec> get processingSpecs => $_getList(9);
}

class UpdateProductResponse extends $pb.GeneratedMessage {
  factory UpdateProductResponse({
    Product? product,
  }) {
    final result = UpdateProductResponse._();
    if (product != null) result.product = product;
    return result;
  }

  UpdateProductResponse._();

  factory UpdateProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductResponse()..mergeFromBuffer(data, registry);
  factory UpdateProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: UpdateProductResponse.$_createMessage)
    ..aOM<Product>(1, _omitFieldNames ? '' : 'product',
        subBuilder: Product.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductResponse copyWith(
          void Function(UpdateProductResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateProductResponse))
          as UpdateProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProductResponse() / UpdateProductResponse.new instead')
  static UpdateProductResponse create() => UpdateProductResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateProductResponse._();
  @$core.override
  UpdateProductResponse createEmptyInstance() => UpdateProductResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProductResponse>(
          UpdateProductResponse.$_createMessage);
  static UpdateProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Product get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(Product value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  Product ensureProduct() => $_ensure(0);
}

class DeleteProductRequest extends $pb.GeneratedMessage {
  factory DeleteProductRequest({
    $core.String? id,
  }) {
    final result = DeleteProductRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteProductRequest._();

  factory DeleteProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductRequest()..mergeFromBuffer(data, registry);
  factory DeleteProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: DeleteProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductRequest copyWith(void Function(DeleteProductRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteProductRequest))
          as DeleteProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteProductRequest() / DeleteProductRequest.new instead')
  static DeleteProductRequest create() => DeleteProductRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteProductRequest._();
  @$core.override
  DeleteProductRequest createEmptyInstance() => DeleteProductRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteProductRequest>(
          DeleteProductRequest.$_createMessage);
  static DeleteProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteProductResponse extends $pb.GeneratedMessage {
  factory DeleteProductResponse() => DeleteProductResponse._();

  DeleteProductResponse._();

  factory DeleteProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductResponse()..mergeFromBuffer(data, registry);
  factory DeleteProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: DeleteProductResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductResponse copyWith(
          void Function(DeleteProductResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteProductResponse))
          as DeleteProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteProductResponse() / DeleteProductResponse.new instead')
  static DeleteProductResponse create() => DeleteProductResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteProductResponse._();
  @$core.override
  DeleteProductResponse createEmptyInstance() => DeleteProductResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteProductResponse>(
          DeleteProductResponse.$_createMessage);
  static DeleteProductResponse? _defaultInstance;
}

class RestoreProductRequest extends $pb.GeneratedMessage {
  factory RestoreProductRequest({
    $core.String? id,
  }) {
    final result = RestoreProductRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RestoreProductRequest._();

  factory RestoreProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductRequest()..mergeFromBuffer(data, registry);
  factory RestoreProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: RestoreProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductRequest copyWith(
          void Function(RestoreProductRequest) updates) =>
      super.copyWith((message) => updates(message as RestoreProductRequest))
          as RestoreProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreProductRequest() / RestoreProductRequest.new instead')
  static RestoreProductRequest create() => RestoreProductRequest._();
  static $pb.GeneratedMessage $_createMessage() => RestoreProductRequest._();
  @$core.override
  RestoreProductRequest createEmptyInstance() => RestoreProductRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreProductRequest>(
          RestoreProductRequest.$_createMessage);
  static RestoreProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RestoreProductResponse extends $pb.GeneratedMessage {
  factory RestoreProductResponse({
    Product? product,
  }) {
    final result = RestoreProductResponse._();
    if (product != null) result.product = product;
    return result;
  }

  RestoreProductResponse._();

  factory RestoreProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductResponse()..mergeFromBuffer(data, registry);
  factory RestoreProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: RestoreProductResponse.$_createMessage)
    ..aOM<Product>(1, _omitFieldNames ? '' : 'product',
        subBuilder: Product.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductResponse copyWith(
          void Function(RestoreProductResponse) updates) =>
      super.copyWith((message) => updates(message as RestoreProductResponse))
          as RestoreProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreProductResponse() / RestoreProductResponse.new instead')
  static RestoreProductResponse create() => RestoreProductResponse._();
  static $pb.GeneratedMessage $_createMessage() => RestoreProductResponse._();
  @$core.override
  RestoreProductResponse createEmptyInstance() => RestoreProductResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreProductResponse>(
          RestoreProductResponse.$_createMessage);
  static RestoreProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Product get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(Product value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  Product ensureProduct() => $_ensure(0);
}

/// CustomerProduct:客戶專屬商品清單(04 Task 3.5,不存單價)。
class CustomerProduct extends $pb.GeneratedMessage {
  factory CustomerProduct({
    $core.String? id,
    $core.String? customerId,
    $core.String? productId,
    $core.String? aliasName,
    $core.String? defaultQty,
    $core.String? cutNote,
    $core.Iterable<$fixnum.Int64>? promoTagIds,
    $core.String? createdAt,
    $core.String? updatedAt,
  }) {
    final result = CustomerProduct._();
    if (id != null) result.id = id;
    if (customerId != null) result.customerId = customerId;
    if (productId != null) result.productId = productId;
    if (aliasName != null) result.aliasName = aliasName;
    if (defaultQty != null) result.defaultQty = defaultQty;
    if (cutNote != null) result.cutNote = cutNote;
    if (promoTagIds != null) result.promoTagIds.addAll(promoTagIds);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  CustomerProduct._();

  factory CustomerProduct.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerProduct()..mergeFromBuffer(data, registry);
  factory CustomerProduct.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerProduct()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CustomerProduct',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: CustomerProduct.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'customerId')
    ..aOS(3, _omitFieldNames ? '' : 'productId')
    ..aOS(4, _omitFieldNames ? '' : 'aliasName')
    ..aOS(5, _omitFieldNames ? '' : 'defaultQty')
    ..aOS(6, _omitFieldNames ? '' : 'cutNote')
    ..p<$fixnum.Int64>(
        7, _omitFieldNames ? '' : 'promoTagIds', $pb.PbFieldType.K6)
    ..aOS(8, _omitFieldNames ? '' : 'createdAt')
    ..aOS(9, _omitFieldNames ? '' : 'updatedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerProduct clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerProduct copyWith(void Function(CustomerProduct) updates) =>
      super.copyWith((message) => updates(message as CustomerProduct))
          as CustomerProduct;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CustomerProduct() / CustomerProduct.new instead')
  static CustomerProduct create() => CustomerProduct._();
  static $pb.GeneratedMessage $_createMessage() => CustomerProduct._();
  @$core.override
  CustomerProduct createEmptyInstance() => CustomerProduct._();
  @$core.pragma('dart2js:noInline')
  static CustomerProduct getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CustomerProduct>(
          CustomerProduct.$_createMessage);
  static CustomerProduct? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get customerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set customerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCustomerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCustomerId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get productId => $_getSZ(2);
  @$pb.TagNumber(3)
  set productId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProductId() => $_has(2);
  @$pb.TagNumber(3)
  void clearProductId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get aliasName => $_getSZ(3);
  @$pb.TagNumber(4)
  set aliasName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAliasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearAliasName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get defaultQty => $_getSZ(4);
  @$pb.TagNumber(5)
  set defaultQty($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDefaultQty() => $_has(4);
  @$pb.TagNumber(5)
  void clearDefaultQty() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get cutNote => $_getSZ(5);
  @$pb.TagNumber(6)
  set cutNote($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCutNote() => $_has(5);
  @$pb.TagNumber(6)
  void clearCutNote() => $_clearField(6);

  @$pb.TagNumber(7)
  $pb.PbList<$fixnum.Int64> get promoTagIds => $_getList(6);

  @$pb.TagNumber(8)
  $core.String get createdAt => $_getSZ(7);
  @$pb.TagNumber(8)
  set createdAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasCreatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCreatedAt() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get updatedAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set updatedAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUpdatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpdatedAt() => $_clearField(9);
}

class ListCustomerProductsRequest extends $pb.GeneratedMessage {
  factory ListCustomerProductsRequest({
    $core.String? customerId,
    $core.bool? forOrder,
    $core.bool? includeDeleted,
  }) {
    final result = ListCustomerProductsRequest._();
    if (customerId != null) result.customerId = customerId;
    if (forOrder != null) result.forOrder = forOrder;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListCustomerProductsRequest._();

  factory ListCustomerProductsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerProductsRequest()..mergeFromBuffer(data, registry);
  factory ListCustomerProductsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerProductsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomerProductsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ListCustomerProductsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOB(2, _omitFieldNames ? '' : 'forOrder')
    ..aOB(3, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerProductsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerProductsRequest copyWith(
          void Function(ListCustomerProductsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListCustomerProductsRequest))
          as ListCustomerProductsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCustomerProductsRequest() / ListCustomerProductsRequest.new instead')
  static ListCustomerProductsRequest create() =>
      ListCustomerProductsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListCustomerProductsRequest._();
  @$core.override
  ListCustomerProductsRequest createEmptyInstance() =>
      ListCustomerProductsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListCustomerProductsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomerProductsRequest>(
          ListCustomerProductsRequest.$_createMessage);
  static ListCustomerProductsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get forOrder => $_getBF(1);
  @$pb.TagNumber(2)
  set forOrder($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasForOrder() => $_has(1);
  @$pb.TagNumber(2)
  void clearForOrder() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get includeDeleted => $_getBF(2);
  @$pb.TagNumber(3)
  set includeDeleted($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasIncludeDeleted() => $_has(2);
  @$pb.TagNumber(3)
  void clearIncludeDeleted() => $_clearField(3);
}

class ListCustomerProductsResponse extends $pb.GeneratedMessage {
  factory ListCustomerProductsResponse({
    $core.Iterable<CustomerProduct>? products,
    $core.int? total,
  }) {
    final result = ListCustomerProductsResponse._();
    if (products != null) result.products.addAll(products);
    if (total != null) result.total = total;
    return result;
  }

  ListCustomerProductsResponse._();

  factory ListCustomerProductsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerProductsResponse()..mergeFromBuffer(data, registry);
  factory ListCustomerProductsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerProductsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomerProductsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ListCustomerProductsResponse.$_createMessage)
    ..pPM<CustomerProduct>(1, _omitFieldNames ? '' : 'products',
        subBuilder: CustomerProduct.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerProductsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerProductsResponse copyWith(
          void Function(ListCustomerProductsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListCustomerProductsResponse))
          as ListCustomerProductsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCustomerProductsResponse() / ListCustomerProductsResponse.new instead')
  static ListCustomerProductsResponse create() =>
      ListCustomerProductsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListCustomerProductsResponse._();
  @$core.override
  ListCustomerProductsResponse createEmptyInstance() =>
      ListCustomerProductsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListCustomerProductsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomerProductsResponse>(
          ListCustomerProductsResponse.$_createMessage);
  static ListCustomerProductsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CustomerProduct> get products => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get total => $_getIZ(1);
  @$pb.TagNumber(2)
  set total($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => $_clearField(2);
}

class AddCustomerProductRequest extends $pb.GeneratedMessage {
  factory AddCustomerProductRequest({
    $core.String? customerId,
    $core.String? productId,
    $core.String? aliasName,
    $core.String? defaultQty,
    $core.String? cutNote,
  }) {
    final result = AddCustomerProductRequest._();
    if (customerId != null) result.customerId = customerId;
    if (productId != null) result.productId = productId;
    if (aliasName != null) result.aliasName = aliasName;
    if (defaultQty != null) result.defaultQty = defaultQty;
    if (cutNote != null) result.cutNote = cutNote;
    return result;
  }

  AddCustomerProductRequest._();

  factory AddCustomerProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddCustomerProductRequest()..mergeFromBuffer(data, registry);
  factory AddCustomerProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddCustomerProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddCustomerProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: AddCustomerProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOS(2, _omitFieldNames ? '' : 'productId')
    ..aOS(3, _omitFieldNames ? '' : 'aliasName')
    ..aOS(4, _omitFieldNames ? '' : 'defaultQty')
    ..aOS(5, _omitFieldNames ? '' : 'cutNote')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddCustomerProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddCustomerProductRequest copyWith(
          void Function(AddCustomerProductRequest) updates) =>
      super.copyWith((message) => updates(message as AddCustomerProductRequest))
          as AddCustomerProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AddCustomerProductRequest() / AddCustomerProductRequest.new instead')
  static AddCustomerProductRequest create() => AddCustomerProductRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      AddCustomerProductRequest._();
  @$core.override
  AddCustomerProductRequest createEmptyInstance() =>
      AddCustomerProductRequest._();
  @$core.pragma('dart2js:noInline')
  static AddCustomerProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AddCustomerProductRequest>(
          AddCustomerProductRequest.$_createMessage);
  static AddCustomerProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get productId => $_getSZ(1);
  @$pb.TagNumber(2)
  set productId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProductId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProductId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get aliasName => $_getSZ(2);
  @$pb.TagNumber(3)
  set aliasName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAliasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAliasName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get defaultQty => $_getSZ(3);
  @$pb.TagNumber(4)
  set defaultQty($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDefaultQty() => $_has(3);
  @$pb.TagNumber(4)
  void clearDefaultQty() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get cutNote => $_getSZ(4);
  @$pb.TagNumber(5)
  set cutNote($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCutNote() => $_has(4);
  @$pb.TagNumber(5)
  void clearCutNote() => $_clearField(5);
}

class AddCustomerProductResponse extends $pb.GeneratedMessage {
  factory AddCustomerProductResponse({
    CustomerProduct? product,
  }) {
    final result = AddCustomerProductResponse._();
    if (product != null) result.product = product;
    return result;
  }

  AddCustomerProductResponse._();

  factory AddCustomerProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddCustomerProductResponse()..mergeFromBuffer(data, registry);
  factory AddCustomerProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddCustomerProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddCustomerProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: AddCustomerProductResponse.$_createMessage)
    ..aOM<CustomerProduct>(1, _omitFieldNames ? '' : 'product',
        subBuilder: CustomerProduct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddCustomerProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddCustomerProductResponse copyWith(
          void Function(AddCustomerProductResponse) updates) =>
      super.copyWith(
              (message) => updates(message as AddCustomerProductResponse))
          as AddCustomerProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AddCustomerProductResponse() / AddCustomerProductResponse.new instead')
  static AddCustomerProductResponse create() => AddCustomerProductResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      AddCustomerProductResponse._();
  @$core.override
  AddCustomerProductResponse createEmptyInstance() =>
      AddCustomerProductResponse._();
  @$core.pragma('dart2js:noInline')
  static AddCustomerProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AddCustomerProductResponse>(
          AddCustomerProductResponse.$_createMessage);
  static AddCustomerProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerProduct get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(CustomerProduct value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerProduct ensureProduct() => $_ensure(0);
}

class UpdateCustomerProductRequest extends $pb.GeneratedMessage {
  factory UpdateCustomerProductRequest({
    $core.String? id,
    $core.String? aliasName,
    $core.String? defaultQty,
    $core.String? cutNote,
  }) {
    final result = UpdateCustomerProductRequest._();
    if (id != null) result.id = id;
    if (aliasName != null) result.aliasName = aliasName;
    if (defaultQty != null) result.defaultQty = defaultQty;
    if (cutNote != null) result.cutNote = cutNote;
    return result;
  }

  UpdateCustomerProductRequest._();

  factory UpdateCustomerProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerProductRequest()..mergeFromBuffer(data, registry);
  factory UpdateCustomerProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCustomerProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: UpdateCustomerProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'aliasName')
    ..aOS(3, _omitFieldNames ? '' : 'defaultQty')
    ..aOS(4, _omitFieldNames ? '' : 'cutNote')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerProductRequest copyWith(
          void Function(UpdateCustomerProductRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateCustomerProductRequest))
          as UpdateCustomerProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCustomerProductRequest() / UpdateCustomerProductRequest.new instead')
  static UpdateCustomerProductRequest create() =>
      UpdateCustomerProductRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateCustomerProductRequest._();
  @$core.override
  UpdateCustomerProductRequest createEmptyInstance() =>
      UpdateCustomerProductRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateCustomerProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCustomerProductRequest>(
          UpdateCustomerProductRequest.$_createMessage);
  static UpdateCustomerProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get aliasName => $_getSZ(1);
  @$pb.TagNumber(2)
  set aliasName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAliasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAliasName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get defaultQty => $_getSZ(2);
  @$pb.TagNumber(3)
  set defaultQty($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDefaultQty() => $_has(2);
  @$pb.TagNumber(3)
  void clearDefaultQty() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get cutNote => $_getSZ(3);
  @$pb.TagNumber(4)
  set cutNote($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCutNote() => $_has(3);
  @$pb.TagNumber(4)
  void clearCutNote() => $_clearField(4);
}

class UpdateCustomerProductResponse extends $pb.GeneratedMessage {
  factory UpdateCustomerProductResponse({
    CustomerProduct? product,
  }) {
    final result = UpdateCustomerProductResponse._();
    if (product != null) result.product = product;
    return result;
  }

  UpdateCustomerProductResponse._();

  factory UpdateCustomerProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerProductResponse()..mergeFromBuffer(data, registry);
  factory UpdateCustomerProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCustomerProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: UpdateCustomerProductResponse.$_createMessage)
    ..aOM<CustomerProduct>(1, _omitFieldNames ? '' : 'product',
        subBuilder: CustomerProduct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerProductResponse copyWith(
          void Function(UpdateCustomerProductResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateCustomerProductResponse))
          as UpdateCustomerProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCustomerProductResponse() / UpdateCustomerProductResponse.new instead')
  static UpdateCustomerProductResponse create() =>
      UpdateCustomerProductResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateCustomerProductResponse._();
  @$core.override
  UpdateCustomerProductResponse createEmptyInstance() =>
      UpdateCustomerProductResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateCustomerProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCustomerProductResponse>(
          UpdateCustomerProductResponse.$_createMessage);
  static UpdateCustomerProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerProduct get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(CustomerProduct value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerProduct ensureProduct() => $_ensure(0);
}

class DeleteCustomerProductRequest extends $pb.GeneratedMessage {
  factory DeleteCustomerProductRequest({
    $core.String? id,
  }) {
    final result = DeleteCustomerProductRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteCustomerProductRequest._();

  factory DeleteCustomerProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerProductRequest()..mergeFromBuffer(data, registry);
  factory DeleteCustomerProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCustomerProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: DeleteCustomerProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerProductRequest copyWith(
          void Function(DeleteCustomerProductRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteCustomerProductRequest))
          as DeleteCustomerProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCustomerProductRequest() / DeleteCustomerProductRequest.new instead')
  static DeleteCustomerProductRequest create() =>
      DeleteCustomerProductRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteCustomerProductRequest._();
  @$core.override
  DeleteCustomerProductRequest createEmptyInstance() =>
      DeleteCustomerProductRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteCustomerProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCustomerProductRequest>(
          DeleteCustomerProductRequest.$_createMessage);
  static DeleteCustomerProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteCustomerProductResponse extends $pb.GeneratedMessage {
  factory DeleteCustomerProductResponse() => DeleteCustomerProductResponse._();

  DeleteCustomerProductResponse._();

  factory DeleteCustomerProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerProductResponse()..mergeFromBuffer(data, registry);
  factory DeleteCustomerProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCustomerProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: DeleteCustomerProductResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerProductResponse copyWith(
          void Function(DeleteCustomerProductResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteCustomerProductResponse))
          as DeleteCustomerProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCustomerProductResponse() / DeleteCustomerProductResponse.new instead')
  static DeleteCustomerProductResponse create() =>
      DeleteCustomerProductResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteCustomerProductResponse._();
  @$core.override
  DeleteCustomerProductResponse createEmptyInstance() =>
      DeleteCustomerProductResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteCustomerProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCustomerProductResponse>(
          DeleteCustomerProductResponse.$_createMessage);
  static DeleteCustomerProductResponse? _defaultInstance;
}

class EnsureCustomerProductRequest extends $pb.GeneratedMessage {
  factory EnsureCustomerProductRequest({
    $core.String? customerId,
    $core.String? productId,
    $core.String? aliasName,
  }) {
    final result = EnsureCustomerProductRequest._();
    if (customerId != null) result.customerId = customerId;
    if (productId != null) result.productId = productId;
    if (aliasName != null) result.aliasName = aliasName;
    return result;
  }

  EnsureCustomerProductRequest._();

  factory EnsureCustomerProductRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EnsureCustomerProductRequest()..mergeFromBuffer(data, registry);
  factory EnsureCustomerProductRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EnsureCustomerProductRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EnsureCustomerProductRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: EnsureCustomerProductRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOS(2, _omitFieldNames ? '' : 'productId')
    ..aOS(3, _omitFieldNames ? '' : 'aliasName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnsureCustomerProductRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnsureCustomerProductRequest copyWith(
          void Function(EnsureCustomerProductRequest) updates) =>
      super.copyWith(
              (message) => updates(message as EnsureCustomerProductRequest))
          as EnsureCustomerProductRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use EnsureCustomerProductRequest() / EnsureCustomerProductRequest.new instead')
  static EnsureCustomerProductRequest create() =>
      EnsureCustomerProductRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      EnsureCustomerProductRequest._();
  @$core.override
  EnsureCustomerProductRequest createEmptyInstance() =>
      EnsureCustomerProductRequest._();
  @$core.pragma('dart2js:noInline')
  static EnsureCustomerProductRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EnsureCustomerProductRequest>(
          EnsureCustomerProductRequest.$_createMessage);
  static EnsureCustomerProductRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get productId => $_getSZ(1);
  @$pb.TagNumber(2)
  set productId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProductId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProductId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get aliasName => $_getSZ(2);
  @$pb.TagNumber(3)
  set aliasName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAliasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAliasName() => $_clearField(3);
}

class EnsureCustomerProductResponse extends $pb.GeneratedMessage {
  factory EnsureCustomerProductResponse({
    CustomerProduct? product,
    $core.bool? created,
  }) {
    final result = EnsureCustomerProductResponse._();
    if (product != null) result.product = product;
    if (created != null) result.created = created;
    return result;
  }

  EnsureCustomerProductResponse._();

  factory EnsureCustomerProductResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EnsureCustomerProductResponse()..mergeFromBuffer(data, registry);
  factory EnsureCustomerProductResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EnsureCustomerProductResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EnsureCustomerProductResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: EnsureCustomerProductResponse.$_createMessage)
    ..aOM<CustomerProduct>(1, _omitFieldNames ? '' : 'product',
        subBuilder: CustomerProduct.$_createMessage)
    ..aOB(2, _omitFieldNames ? '' : 'created')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnsureCustomerProductResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnsureCustomerProductResponse copyWith(
          void Function(EnsureCustomerProductResponse) updates) =>
      super.copyWith(
              (message) => updates(message as EnsureCustomerProductResponse))
          as EnsureCustomerProductResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use EnsureCustomerProductResponse() / EnsureCustomerProductResponse.new instead')
  static EnsureCustomerProductResponse create() =>
      EnsureCustomerProductResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      EnsureCustomerProductResponse._();
  @$core.override
  EnsureCustomerProductResponse createEmptyInstance() =>
      EnsureCustomerProductResponse._();
  @$core.pragma('dart2js:noInline')
  static EnsureCustomerProductResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EnsureCustomerProductResponse>(
          EnsureCustomerProductResponse.$_createMessage);
  static EnsureCustomerProductResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerProduct get product => $_getN(0);
  @$pb.TagNumber(1)
  set product(CustomerProduct value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProduct() => $_has(0);
  @$pb.TagNumber(1)
  void clearProduct() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerProduct ensureProduct() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.bool get created => $_getBF(1);
  @$pb.TagNumber(2)
  set created($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCreated() => $_has(1);
  @$pb.TagNumber(2)
  void clearCreated() => $_clearField(2);
}

class ProductServiceApi {
  final $pb.RpcClient _client;

  ProductServiceApi(this._client);

  $async.Future<ListProductsResponse> listProducts(
          $pb.ClientContext? ctx, ListProductsRequest request) =>
      _client.invoke<ListProductsResponse>(ctx, 'ProductService',
          'ListProducts', request, ListProductsResponse());
  $async.Future<GetProductResponse> getProduct(
          $pb.ClientContext? ctx, GetProductRequest request) =>
      _client.invoke<GetProductResponse>(
          ctx, 'ProductService', 'GetProduct', request, GetProductResponse());
  $async.Future<CreateProductResponse> createProduct(
          $pb.ClientContext? ctx, CreateProductRequest request) =>
      _client.invoke<CreateProductResponse>(ctx, 'ProductService',
          'CreateProduct', request, CreateProductResponse());
  $async.Future<UpdateProductResponse> updateProduct(
          $pb.ClientContext? ctx, UpdateProductRequest request) =>
      _client.invoke<UpdateProductResponse>(ctx, 'ProductService',
          'UpdateProduct', request, UpdateProductResponse());
  $async.Future<DeleteProductResponse> deleteProduct(
          $pb.ClientContext? ctx, DeleteProductRequest request) =>
      _client.invoke<DeleteProductResponse>(ctx, 'ProductService',
          'DeleteProduct', request, DeleteProductResponse());
  $async.Future<RestoreProductResponse> restoreProduct(
          $pb.ClientContext? ctx, RestoreProductRequest request) =>
      _client.invoke<RestoreProductResponse>(ctx, 'ProductService',
          'RestoreProduct', request, RestoreProductResponse());
}

/// CustomerProductService:客戶專屬清單(dept_admin/staff 限部門;客戶僅 for_order 語意由 List 旗標表達)。
class CustomerProductServiceApi {
  final $pb.RpcClient _client;

  CustomerProductServiceApi(this._client);

  /// ListCustomerProducts:查該客戶清單(for_order=true 排除 default_qty=0 與已刪)。
  $async.Future<ListCustomerProductsResponse> listCustomerProducts(
          $pb.ClientContext? ctx, ListCustomerProductsRequest request) =>
      _client.invoke<ListCustomerProductsResponse>(
          ctx,
          'CustomerProductService',
          'ListCustomerProducts',
          request,
          ListCustomerProductsResponse());

  /// AddCustomerProduct:新增一筆(一客戶一商品;重複未刪 → already_exists)。
  $async.Future<AddCustomerProductResponse> addCustomerProduct(
          $pb.ClientContext? ctx, AddCustomerProductRequest request) =>
      _client.invoke<AddCustomerProductResponse>(ctx, 'CustomerProductService',
          'AddCustomerProduct', request, AddCustomerProductResponse());

  /// UpdateCustomerProduct:改 alias/default_qty/cut_note(不可改 customer/product)。
  $async.Future<UpdateCustomerProductResponse> updateCustomerProduct(
          $pb.ClientContext? ctx, UpdateCustomerProductRequest request) =>
      _client.invoke<UpdateCustomerProductResponse>(
          ctx,
          'CustomerProductService',
          'UpdateCustomerProduct',
          request,
          UpdateCustomerProductResponse());

  /// DeleteCustomerProduct:軟刪除 + 稽核。
  $async.Future<DeleteCustomerProductResponse> deleteCustomerProduct(
          $pb.ClientContext? ctx, DeleteCustomerProductRequest request) =>
      _client.invoke<DeleteCustomerProductResponse>(
          ctx,
          'CustomerProductService',
          'DeleteCustomerProduct',
          request,
          DeleteCustomerProductResponse());

  /// EnsureCustomerProduct:下單手打確認儲存後呼叫(冪等:存在回既有 created=false;唯一衝突吸收)。
  $async.Future<EnsureCustomerProductResponse> ensureCustomerProduct(
          $pb.ClientContext? ctx, EnsureCustomerProductRequest request) =>
      _client.invoke<EnsureCustomerProductResponse>(
          ctx,
          'CustomerProductService',
          'EnsureCustomerProduct',
          request,
          EnsureCustomerProductResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
