// This is a generated file - do not edit.
//
// Generated from masters/v1/master.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $1;

import '../../salesorder/v1/common.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// ---- Warehouse(3.4.1) ----
class Warehouse extends $pb.GeneratedMessage {
  factory Warehouse({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? code,
    $core.String? name,
    $core.String? address,
    $core.bool? isActive,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = Warehouse._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (address != null) result.address = address;
    if (isActive != null) result.isActive = isActive;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  Warehouse._();

  factory Warehouse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Warehouse()..mergeFromBuffer(data, registry);
  factory Warehouse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Warehouse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Warehouse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: Warehouse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'code')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'address')
    ..aOB(7, _omitFieldNames ? '' : 'isActive')
    ..aOS(8, _omitFieldNames ? '' : 'createdAt')
    ..aOS(9, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(10, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Warehouse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Warehouse copyWith(void Function(Warehouse) updates) =>
      super.copyWith((message) => updates(message as Warehouse)) as Warehouse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Warehouse() / Warehouse.new instead')
  static Warehouse create() => Warehouse._();
  static $pb.GeneratedMessage $_createMessage() => Warehouse._();
  @$core.override
  Warehouse createEmptyInstance() => Warehouse._();
  @$core.pragma('dart2js:noInline')
  static Warehouse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Warehouse>(Warehouse.$_createMessage);
  static Warehouse? _defaultInstance;

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
  $core.String get address => $_getSZ(5);
  @$pb.TagNumber(6)
  set address($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAddress() => $_has(5);
  @$pb.TagNumber(6)
  void clearAddress() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get isActive => $_getBF(6);
  @$pb.TagNumber(7)
  set isActive($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasIsActive() => $_has(6);
  @$pb.TagNumber(7)
  void clearIsActive() => $_clearField(7);

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

  @$pb.TagNumber(10)
  $core.String get deletedAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set deletedAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasDeletedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearDeletedAt() => $_clearField(10);
}

class ListWarehousesRequest extends $pb.GeneratedMessage {
  factory ListWarehousesRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.bool? includeDeleted,
  }) {
    final result = ListWarehousesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListWarehousesRequest._();

  factory ListWarehousesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWarehousesRequest()..mergeFromBuffer(data, registry);
  factory ListWarehousesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWarehousesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWarehousesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListWarehousesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWarehousesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWarehousesRequest copyWith(
          void Function(ListWarehousesRequest) updates) =>
      super.copyWith((message) => updates(message as ListWarehousesRequest))
          as ListWarehousesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWarehousesRequest() / ListWarehousesRequest.new instead')
  static ListWarehousesRequest create() => ListWarehousesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListWarehousesRequest._();
  @$core.override
  ListWarehousesRequest createEmptyInstance() => ListWarehousesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListWarehousesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWarehousesRequest>(
          ListWarehousesRequest.$_createMessage);
  static ListWarehousesRequest? _defaultInstance;

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
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(4)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(4)
  void clearIncludeDeleted() => $_clearField(4);
}

class ListWarehousesResponse extends $pb.GeneratedMessage {
  factory ListWarehousesResponse({
    $core.Iterable<Warehouse>? warehouses,
    $0.Pagination? pagination,
  }) {
    final result = ListWarehousesResponse._();
    if (warehouses != null) result.warehouses.addAll(warehouses);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListWarehousesResponse._();

  factory ListWarehousesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWarehousesResponse()..mergeFromBuffer(data, registry);
  factory ListWarehousesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWarehousesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWarehousesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListWarehousesResponse.$_createMessage)
    ..pPM<Warehouse>(1, _omitFieldNames ? '' : 'warehouses',
        subBuilder: Warehouse.$_createMessage)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWarehousesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWarehousesResponse copyWith(
          void Function(ListWarehousesResponse) updates) =>
      super.copyWith((message) => updates(message as ListWarehousesResponse))
          as ListWarehousesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWarehousesResponse() / ListWarehousesResponse.new instead')
  static ListWarehousesResponse create() => ListWarehousesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListWarehousesResponse._();
  @$core.override
  ListWarehousesResponse createEmptyInstance() => ListWarehousesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListWarehousesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWarehousesResponse>(
          ListWarehousesResponse.$_createMessage);
  static ListWarehousesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Warehouse> get warehouses => $_getList(0);

  @$pb.TagNumber(2)
  $0.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($0.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Pagination ensurePagination() => $_ensure(1);
}

class CreateWarehouseRequest extends $pb.GeneratedMessage {
  factory CreateWarehouseRequest({
    $core.String? code,
    $core.String? name,
    $core.String? address,
    $core.bool? isActive,
  }) {
    final result = CreateWarehouseRequest._();
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (address != null) result.address = address;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  CreateWarehouseRequest._();

  factory CreateWarehouseRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWarehouseRequest()..mergeFromBuffer(data, registry);
  factory CreateWarehouseRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWarehouseRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWarehouseRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateWarehouseRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'address')
    ..aOB(4, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWarehouseRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWarehouseRequest copyWith(
          void Function(CreateWarehouseRequest) updates) =>
      super.copyWith((message) => updates(message as CreateWarehouseRequest))
          as CreateWarehouseRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateWarehouseRequest() / CreateWarehouseRequest.new instead')
  static CreateWarehouseRequest create() => CreateWarehouseRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateWarehouseRequest._();
  @$core.override
  CreateWarehouseRequest createEmptyInstance() => CreateWarehouseRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateWarehouseRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWarehouseRequest>(
          CreateWarehouseRequest.$_createMessage);
  static CreateWarehouseRequest? _defaultInstance;

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
  $core.String get address => $_getSZ(2);
  @$pb.TagNumber(3)
  set address($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearAddress() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get isActive => $_getBF(3);
  @$pb.TagNumber(4)
  set isActive($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsActive() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsActive() => $_clearField(4);
}

class CreateWarehouseResponse extends $pb.GeneratedMessage {
  factory CreateWarehouseResponse({
    Warehouse? warehouse,
  }) {
    final result = CreateWarehouseResponse._();
    if (warehouse != null) result.warehouse = warehouse;
    return result;
  }

  CreateWarehouseResponse._();

  factory CreateWarehouseResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWarehouseResponse()..mergeFromBuffer(data, registry);
  factory CreateWarehouseResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWarehouseResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWarehouseResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateWarehouseResponse.$_createMessage)
    ..aOM<Warehouse>(1, _omitFieldNames ? '' : 'warehouse',
        subBuilder: Warehouse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWarehouseResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWarehouseResponse copyWith(
          void Function(CreateWarehouseResponse) updates) =>
      super.copyWith((message) => updates(message as CreateWarehouseResponse))
          as CreateWarehouseResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateWarehouseResponse() / CreateWarehouseResponse.new instead')
  static CreateWarehouseResponse create() => CreateWarehouseResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateWarehouseResponse._();
  @$core.override
  CreateWarehouseResponse createEmptyInstance() => CreateWarehouseResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateWarehouseResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWarehouseResponse>(
          CreateWarehouseResponse.$_createMessage);
  static CreateWarehouseResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Warehouse get warehouse => $_getN(0);
  @$pb.TagNumber(1)
  set warehouse(Warehouse value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWarehouse() => $_has(0);
  @$pb.TagNumber(1)
  void clearWarehouse() => $_clearField(1);
  @$pb.TagNumber(1)
  Warehouse ensureWarehouse() => $_ensure(0);
}

class UpdateWarehouseRequest extends $pb.GeneratedMessage {
  factory UpdateWarehouseRequest({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.String? address,
    $core.bool? isActive,
  }) {
    final result = UpdateWarehouseRequest._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (address != null) result.address = address;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  UpdateWarehouseRequest._();

  factory UpdateWarehouseRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWarehouseRequest()..mergeFromBuffer(data, registry);
  factory UpdateWarehouseRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWarehouseRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateWarehouseRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateWarehouseRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'address')
    ..aOB(5, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWarehouseRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWarehouseRequest copyWith(
          void Function(UpdateWarehouseRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateWarehouseRequest))
          as UpdateWarehouseRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateWarehouseRequest() / UpdateWarehouseRequest.new instead')
  static UpdateWarehouseRequest create() => UpdateWarehouseRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateWarehouseRequest._();
  @$core.override
  UpdateWarehouseRequest createEmptyInstance() => UpdateWarehouseRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateWarehouseRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateWarehouseRequest>(
          UpdateWarehouseRequest.$_createMessage);
  static UpdateWarehouseRequest? _defaultInstance;

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
  $core.String get address => $_getSZ(3);
  @$pb.TagNumber(4)
  set address($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAddress() => $_has(3);
  @$pb.TagNumber(4)
  void clearAddress() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get isActive => $_getBF(4);
  @$pb.TagNumber(5)
  set isActive($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIsActive() => $_has(4);
  @$pb.TagNumber(5)
  void clearIsActive() => $_clearField(5);
}

class UpdateWarehouseResponse extends $pb.GeneratedMessage {
  factory UpdateWarehouseResponse({
    Warehouse? warehouse,
  }) {
    final result = UpdateWarehouseResponse._();
    if (warehouse != null) result.warehouse = warehouse;
    return result;
  }

  UpdateWarehouseResponse._();

  factory UpdateWarehouseResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWarehouseResponse()..mergeFromBuffer(data, registry);
  factory UpdateWarehouseResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWarehouseResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateWarehouseResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateWarehouseResponse.$_createMessage)
    ..aOM<Warehouse>(1, _omitFieldNames ? '' : 'warehouse',
        subBuilder: Warehouse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWarehouseResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWarehouseResponse copyWith(
          void Function(UpdateWarehouseResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateWarehouseResponse))
          as UpdateWarehouseResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateWarehouseResponse() / UpdateWarehouseResponse.new instead')
  static UpdateWarehouseResponse create() => UpdateWarehouseResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateWarehouseResponse._();
  @$core.override
  UpdateWarehouseResponse createEmptyInstance() => UpdateWarehouseResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateWarehouseResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateWarehouseResponse>(
          UpdateWarehouseResponse.$_createMessage);
  static UpdateWarehouseResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Warehouse get warehouse => $_getN(0);
  @$pb.TagNumber(1)
  set warehouse(Warehouse value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWarehouse() => $_has(0);
  @$pb.TagNumber(1)
  void clearWarehouse() => $_clearField(1);
  @$pb.TagNumber(1)
  Warehouse ensureWarehouse() => $_ensure(0);
}

class DeleteWarehouseRequest extends $pb.GeneratedMessage {
  factory DeleteWarehouseRequest({
    $core.String? id,
  }) {
    final result = DeleteWarehouseRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteWarehouseRequest._();

  factory DeleteWarehouseRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWarehouseRequest()..mergeFromBuffer(data, registry);
  factory DeleteWarehouseRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWarehouseRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWarehouseRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteWarehouseRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWarehouseRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWarehouseRequest copyWith(
          void Function(DeleteWarehouseRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteWarehouseRequest))
          as DeleteWarehouseRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWarehouseRequest() / DeleteWarehouseRequest.new instead')
  static DeleteWarehouseRequest create() => DeleteWarehouseRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteWarehouseRequest._();
  @$core.override
  DeleteWarehouseRequest createEmptyInstance() => DeleteWarehouseRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteWarehouseRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWarehouseRequest>(
          DeleteWarehouseRequest.$_createMessage);
  static DeleteWarehouseRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteWarehouseResponse extends $pb.GeneratedMessage {
  factory DeleteWarehouseResponse() => DeleteWarehouseResponse._();

  DeleteWarehouseResponse._();

  factory DeleteWarehouseResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWarehouseResponse()..mergeFromBuffer(data, registry);
  factory DeleteWarehouseResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWarehouseResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWarehouseResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteWarehouseResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWarehouseResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWarehouseResponse copyWith(
          void Function(DeleteWarehouseResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteWarehouseResponse))
          as DeleteWarehouseResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWarehouseResponse() / DeleteWarehouseResponse.new instead')
  static DeleteWarehouseResponse create() => DeleteWarehouseResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteWarehouseResponse._();
  @$core.override
  DeleteWarehouseResponse createEmptyInstance() => DeleteWarehouseResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteWarehouseResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWarehouseResponse>(
          DeleteWarehouseResponse.$_createMessage);
  static DeleteWarehouseResponse? _defaultInstance;
}

class RestoreWarehouseRequest extends $pb.GeneratedMessage {
  factory RestoreWarehouseRequest({
    $core.String? id,
  }) {
    final result = RestoreWarehouseRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RestoreWarehouseRequest._();

  factory RestoreWarehouseRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreWarehouseRequest()..mergeFromBuffer(data, registry);
  factory RestoreWarehouseRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreWarehouseRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreWarehouseRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreWarehouseRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreWarehouseRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreWarehouseRequest copyWith(
          void Function(RestoreWarehouseRequest) updates) =>
      super.copyWith((message) => updates(message as RestoreWarehouseRequest))
          as RestoreWarehouseRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreWarehouseRequest() / RestoreWarehouseRequest.new instead')
  static RestoreWarehouseRequest create() => RestoreWarehouseRequest._();
  static $pb.GeneratedMessage $_createMessage() => RestoreWarehouseRequest._();
  @$core.override
  RestoreWarehouseRequest createEmptyInstance() => RestoreWarehouseRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreWarehouseRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreWarehouseRequest>(
          RestoreWarehouseRequest.$_createMessage);
  static RestoreWarehouseRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RestoreWarehouseResponse extends $pb.GeneratedMessage {
  factory RestoreWarehouseResponse({
    Warehouse? warehouse,
  }) {
    final result = RestoreWarehouseResponse._();
    if (warehouse != null) result.warehouse = warehouse;
    return result;
  }

  RestoreWarehouseResponse._();

  factory RestoreWarehouseResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreWarehouseResponse()..mergeFromBuffer(data, registry);
  factory RestoreWarehouseResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreWarehouseResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreWarehouseResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreWarehouseResponse.$_createMessage)
    ..aOM<Warehouse>(1, _omitFieldNames ? '' : 'warehouse',
        subBuilder: Warehouse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreWarehouseResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreWarehouseResponse copyWith(
          void Function(RestoreWarehouseResponse) updates) =>
      super.copyWith((message) => updates(message as RestoreWarehouseResponse))
          as RestoreWarehouseResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreWarehouseResponse() / RestoreWarehouseResponse.new instead')
  static RestoreWarehouseResponse create() => RestoreWarehouseResponse._();
  static $pb.GeneratedMessage $_createMessage() => RestoreWarehouseResponse._();
  @$core.override
  RestoreWarehouseResponse createEmptyInstance() =>
      RestoreWarehouseResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreWarehouseResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreWarehouseResponse>(
          RestoreWarehouseResponse.$_createMessage);
  static RestoreWarehouseResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Warehouse get warehouse => $_getN(0);
  @$pb.TagNumber(1)
  set warehouse(Warehouse value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWarehouse() => $_has(0);
  @$pb.TagNumber(1)
  void clearWarehouse() => $_clearField(1);
  @$pb.TagNumber(1)
  Warehouse ensureWarehouse() => $_ensure(0);
}

/// ---- Route(3.4.2) ----
class Route extends $pb.GeneratedMessage {
  factory Route({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? code,
    $core.String? name,
    $core.String? description,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = Route._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  Route._();

  factory Route.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Route()..mergeFromBuffer(data, registry);
  factory Route.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Route()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Route',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: Route.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'code')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'description')
    ..aI(7, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(8, _omitFieldNames ? '' : 'isActive')
    ..aOS(9, _omitFieldNames ? '' : 'createdAt')
    ..aOS(10, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(11, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Route clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Route copyWith(void Function(Route) updates) =>
      super.copyWith((message) => updates(message as Route)) as Route;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Route() / Route.new instead')
  static Route create() => Route._();
  static $pb.GeneratedMessage $_createMessage() => Route._();
  @$core.override
  Route createEmptyInstance() => Route._();
  @$core.pragma('dart2js:noInline')
  static Route getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Route>(Route.$_createMessage);
  static Route? _defaultInstance;

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
  $core.String get description => $_getSZ(5);
  @$pb.TagNumber(6)
  set description($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasDescription() => $_has(5);
  @$pb.TagNumber(6)
  void clearDescription() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.int get sortOrder => $_getIZ(6);
  @$pb.TagNumber(7)
  set sortOrder($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasSortOrder() => $_has(6);
  @$pb.TagNumber(7)
  void clearSortOrder() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get isActive => $_getBF(7);
  @$pb.TagNumber(8)
  set isActive($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasIsActive() => $_has(7);
  @$pb.TagNumber(8)
  void clearIsActive() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get createdAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set createdAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCreatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearCreatedAt() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get updatedAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set updatedAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasUpdatedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearUpdatedAt() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get deletedAt => $_getSZ(10);
  @$pb.TagNumber(11)
  set deletedAt($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasDeletedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearDeletedAt() => $_clearField(11);
}

class ListRoutesRequest extends $pb.GeneratedMessage {
  factory ListRoutesRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.bool? includeDeleted,
  }) {
    final result = ListRoutesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListRoutesRequest._();

  factory ListRoutesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRoutesRequest()..mergeFromBuffer(data, registry);
  factory ListRoutesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRoutesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRoutesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListRoutesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRoutesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRoutesRequest copyWith(void Function(ListRoutesRequest) updates) =>
      super.copyWith((message) => updates(message as ListRoutesRequest))
          as ListRoutesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListRoutesRequest() / ListRoutesRequest.new instead')
  static ListRoutesRequest create() => ListRoutesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListRoutesRequest._();
  @$core.override
  ListRoutesRequest createEmptyInstance() => ListRoutesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListRoutesRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListRoutesRequest>(
          ListRoutesRequest.$_createMessage);
  static ListRoutesRequest? _defaultInstance;

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
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(4)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(4)
  void clearIncludeDeleted() => $_clearField(4);
}

class ListRoutesResponse extends $pb.GeneratedMessage {
  factory ListRoutesResponse({
    $core.Iterable<Route>? routes,
    $0.Pagination? pagination,
  }) {
    final result = ListRoutesResponse._();
    if (routes != null) result.routes.addAll(routes);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListRoutesResponse._();

  factory ListRoutesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRoutesResponse()..mergeFromBuffer(data, registry);
  factory ListRoutesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRoutesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRoutesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListRoutesResponse.$_createMessage)
    ..pPM<Route>(1, _omitFieldNames ? '' : 'routes',
        subBuilder: Route.$_createMessage)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRoutesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRoutesResponse copyWith(void Function(ListRoutesResponse) updates) =>
      super.copyWith((message) => updates(message as ListRoutesResponse))
          as ListRoutesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListRoutesResponse() / ListRoutesResponse.new instead')
  static ListRoutesResponse create() => ListRoutesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListRoutesResponse._();
  @$core.override
  ListRoutesResponse createEmptyInstance() => ListRoutesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListRoutesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRoutesResponse>(
          ListRoutesResponse.$_createMessage);
  static ListRoutesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Route> get routes => $_getList(0);

  @$pb.TagNumber(2)
  $0.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($0.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Pagination ensurePagination() => $_ensure(1);
}

class CreateRouteRequest extends $pb.GeneratedMessage {
  factory CreateRouteRequest({
    $core.String? code,
    $core.String? name,
    $core.String? description,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = CreateRouteRequest._();
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  CreateRouteRequest._();

  factory CreateRouteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRouteRequest()..mergeFromBuffer(data, registry);
  factory CreateRouteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRouteRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateRouteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateRouteRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..aI(4, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(5, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRouteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRouteRequest copyWith(void Function(CreateRouteRequest) updates) =>
      super.copyWith((message) => updates(message as CreateRouteRequest))
          as CreateRouteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateRouteRequest() / CreateRouteRequest.new instead')
  static CreateRouteRequest create() => CreateRouteRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateRouteRequest._();
  @$core.override
  CreateRouteRequest createEmptyInstance() => CreateRouteRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateRouteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateRouteRequest>(
          CreateRouteRequest.$_createMessage);
  static CreateRouteRequest? _defaultInstance;

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
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get sortOrder => $_getIZ(3);
  @$pb.TagNumber(4)
  set sortOrder($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSortOrder() => $_has(3);
  @$pb.TagNumber(4)
  void clearSortOrder() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get isActive => $_getBF(4);
  @$pb.TagNumber(5)
  set isActive($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIsActive() => $_has(4);
  @$pb.TagNumber(5)
  void clearIsActive() => $_clearField(5);
}

class CreateRouteResponse extends $pb.GeneratedMessage {
  factory CreateRouteResponse({
    Route? route,
  }) {
    final result = CreateRouteResponse._();
    if (route != null) result.route = route;
    return result;
  }

  CreateRouteResponse._();

  factory CreateRouteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRouteResponse()..mergeFromBuffer(data, registry);
  factory CreateRouteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRouteResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateRouteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateRouteResponse.$_createMessage)
    ..aOM<Route>(1, _omitFieldNames ? '' : 'route',
        subBuilder: Route.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRouteResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRouteResponse copyWith(void Function(CreateRouteResponse) updates) =>
      super.copyWith((message) => updates(message as CreateRouteResponse))
          as CreateRouteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CreateRouteResponse() / CreateRouteResponse.new instead')
  static CreateRouteResponse create() => CreateRouteResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateRouteResponse._();
  @$core.override
  CreateRouteResponse createEmptyInstance() => CreateRouteResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateRouteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateRouteResponse>(
          CreateRouteResponse.$_createMessage);
  static CreateRouteResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Route get route => $_getN(0);
  @$pb.TagNumber(1)
  set route(Route value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRoute() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoute() => $_clearField(1);
  @$pb.TagNumber(1)
  Route ensureRoute() => $_ensure(0);
}

class UpdateRouteRequest extends $pb.GeneratedMessage {
  factory UpdateRouteRequest({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.String? description,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = UpdateRouteRequest._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  UpdateRouteRequest._();

  factory UpdateRouteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRouteRequest()..mergeFromBuffer(data, registry);
  factory UpdateRouteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRouteRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateRouteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateRouteRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'description')
    ..aI(5, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(6, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRouteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRouteRequest copyWith(void Function(UpdateRouteRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateRouteRequest))
          as UpdateRouteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UpdateRouteRequest() / UpdateRouteRequest.new instead')
  static UpdateRouteRequest create() => UpdateRouteRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateRouteRequest._();
  @$core.override
  UpdateRouteRequest createEmptyInstance() => UpdateRouteRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateRouteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateRouteRequest>(
          UpdateRouteRequest.$_createMessage);
  static UpdateRouteRequest? _defaultInstance;

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
  $core.String get description => $_getSZ(3);
  @$pb.TagNumber(4)
  set description($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDescription() => $_has(3);
  @$pb.TagNumber(4)
  void clearDescription() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get sortOrder => $_getIZ(4);
  @$pb.TagNumber(5)
  set sortOrder($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSortOrder() => $_has(4);
  @$pb.TagNumber(5)
  void clearSortOrder() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get isActive => $_getBF(5);
  @$pb.TagNumber(6)
  set isActive($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasIsActive() => $_has(5);
  @$pb.TagNumber(6)
  void clearIsActive() => $_clearField(6);
}

class UpdateRouteResponse extends $pb.GeneratedMessage {
  factory UpdateRouteResponse({
    Route? route,
  }) {
    final result = UpdateRouteResponse._();
    if (route != null) result.route = route;
    return result;
  }

  UpdateRouteResponse._();

  factory UpdateRouteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRouteResponse()..mergeFromBuffer(data, registry);
  factory UpdateRouteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRouteResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateRouteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateRouteResponse.$_createMessage)
    ..aOM<Route>(1, _omitFieldNames ? '' : 'route',
        subBuilder: Route.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRouteResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRouteResponse copyWith(void Function(UpdateRouteResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateRouteResponse))
          as UpdateRouteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use UpdateRouteResponse() / UpdateRouteResponse.new instead')
  static UpdateRouteResponse create() => UpdateRouteResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateRouteResponse._();
  @$core.override
  UpdateRouteResponse createEmptyInstance() => UpdateRouteResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateRouteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateRouteResponse>(
          UpdateRouteResponse.$_createMessage);
  static UpdateRouteResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Route get route => $_getN(0);
  @$pb.TagNumber(1)
  set route(Route value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRoute() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoute() => $_clearField(1);
  @$pb.TagNumber(1)
  Route ensureRoute() => $_ensure(0);
}

class DeleteRouteRequest extends $pb.GeneratedMessage {
  factory DeleteRouteRequest({
    $core.String? id,
  }) {
    final result = DeleteRouteRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteRouteRequest._();

  factory DeleteRouteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRouteRequest()..mergeFromBuffer(data, registry);
  factory DeleteRouteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRouteRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteRouteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteRouteRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRouteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRouteRequest copyWith(void Function(DeleteRouteRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteRouteRequest))
          as DeleteRouteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeleteRouteRequest() / DeleteRouteRequest.new instead')
  static DeleteRouteRequest create() => DeleteRouteRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteRouteRequest._();
  @$core.override
  DeleteRouteRequest createEmptyInstance() => DeleteRouteRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteRouteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteRouteRequest>(
          DeleteRouteRequest.$_createMessage);
  static DeleteRouteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteRouteResponse extends $pb.GeneratedMessage {
  factory DeleteRouteResponse() => DeleteRouteResponse._();

  DeleteRouteResponse._();

  factory DeleteRouteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRouteResponse()..mergeFromBuffer(data, registry);
  factory DeleteRouteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRouteResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteRouteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteRouteResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRouteResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRouteResponse copyWith(void Function(DeleteRouteResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteRouteResponse))
          as DeleteRouteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DeleteRouteResponse() / DeleteRouteResponse.new instead')
  static DeleteRouteResponse create() => DeleteRouteResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteRouteResponse._();
  @$core.override
  DeleteRouteResponse createEmptyInstance() => DeleteRouteResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteRouteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteRouteResponse>(
          DeleteRouteResponse.$_createMessage);
  static DeleteRouteResponse? _defaultInstance;
}

class RestoreRouteRequest extends $pb.GeneratedMessage {
  factory RestoreRouteRequest({
    $core.String? id,
  }) {
    final result = RestoreRouteRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RestoreRouteRequest._();

  factory RestoreRouteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreRouteRequest()..mergeFromBuffer(data, registry);
  factory RestoreRouteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreRouteRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreRouteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreRouteRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreRouteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreRouteRequest copyWith(void Function(RestoreRouteRequest) updates) =>
      super.copyWith((message) => updates(message as RestoreRouteRequest))
          as RestoreRouteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use RestoreRouteRequest() / RestoreRouteRequest.new instead')
  static RestoreRouteRequest create() => RestoreRouteRequest._();
  static $pb.GeneratedMessage $_createMessage() => RestoreRouteRequest._();
  @$core.override
  RestoreRouteRequest createEmptyInstance() => RestoreRouteRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreRouteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreRouteRequest>(
          RestoreRouteRequest.$_createMessage);
  static RestoreRouteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RestoreRouteResponse extends $pb.GeneratedMessage {
  factory RestoreRouteResponse({
    Route? route,
  }) {
    final result = RestoreRouteResponse._();
    if (route != null) result.route = route;
    return result;
  }

  RestoreRouteResponse._();

  factory RestoreRouteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreRouteResponse()..mergeFromBuffer(data, registry);
  factory RestoreRouteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreRouteResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreRouteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreRouteResponse.$_createMessage)
    ..aOM<Route>(1, _omitFieldNames ? '' : 'route',
        subBuilder: Route.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreRouteResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreRouteResponse copyWith(void Function(RestoreRouteResponse) updates) =>
      super.copyWith((message) => updates(message as RestoreRouteResponse))
          as RestoreRouteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreRouteResponse() / RestoreRouteResponse.new instead')
  static RestoreRouteResponse create() => RestoreRouteResponse._();
  static $pb.GeneratedMessage $_createMessage() => RestoreRouteResponse._();
  @$core.override
  RestoreRouteResponse createEmptyInstance() => RestoreRouteResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreRouteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreRouteResponse>(
          RestoreRouteResponse.$_createMessage);
  static RestoreRouteResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Route get route => $_getN(0);
  @$pb.TagNumber(1)
  set route(Route value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRoute() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoute() => $_clearField(1);
  @$pb.TagNumber(1)
  Route ensureRoute() => $_ensure(0);
}

/// ---- ProcessingSpec(3.4.3, 泛化自 cutting_specs) ----
class ProcessingSpec extends $pb.GeneratedMessage {
  factory ProcessingSpec({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? code,
    $core.String? name,
    $core.String? kind,
    $core.bool? appliesToProcessing,
    $core.bool? appliesToPicking,
    $1.Struct? attributes,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = ProcessingSpec._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (kind != null) result.kind = kind;
    if (appliesToProcessing != null)
      result.appliesToProcessing = appliesToProcessing;
    if (appliesToPicking != null) result.appliesToPicking = appliesToPicking;
    if (attributes != null) result.attributes = attributes;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  ProcessingSpec._();

  factory ProcessingSpec.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProcessingSpec()..mergeFromBuffer(data, registry);
  factory ProcessingSpec.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProcessingSpec()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProcessingSpec',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ProcessingSpec.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'code')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'kind')
    ..aOB(7, _omitFieldNames ? '' : 'appliesToProcessing')
    ..aOB(8, _omitFieldNames ? '' : 'appliesToPicking')
    ..aOM<$1.Struct>(9, _omitFieldNames ? '' : 'attributes',
        subBuilder: $1.Struct.$_createMessage)
    ..aI(10, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(11, _omitFieldNames ? '' : 'isActive')
    ..aOS(12, _omitFieldNames ? '' : 'createdAt')
    ..aOS(13, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(14, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProcessingSpec clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProcessingSpec copyWith(void Function(ProcessingSpec) updates) =>
      super.copyWith((message) => updates(message as ProcessingSpec))
          as ProcessingSpec;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProcessingSpec() / ProcessingSpec.new instead')
  static ProcessingSpec create() => ProcessingSpec._();
  static $pb.GeneratedMessage $_createMessage() => ProcessingSpec._();
  @$core.override
  ProcessingSpec createEmptyInstance() => ProcessingSpec._();
  @$core.pragma('dart2js:noInline')
  static ProcessingSpec getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProcessingSpec>(
          ProcessingSpec.$_createMessage);
  static ProcessingSpec? _defaultInstance;

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
  $core.String get kind => $_getSZ(5);
  @$pb.TagNumber(6)
  set kind($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasKind() => $_has(5);
  @$pb.TagNumber(6)
  void clearKind() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get appliesToProcessing => $_getBF(6);
  @$pb.TagNumber(7)
  set appliesToProcessing($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasAppliesToProcessing() => $_has(6);
  @$pb.TagNumber(7)
  void clearAppliesToProcessing() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get appliesToPicking => $_getBF(7);
  @$pb.TagNumber(8)
  set appliesToPicking($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAppliesToPicking() => $_has(7);
  @$pb.TagNumber(8)
  void clearAppliesToPicking() => $_clearField(8);

  @$pb.TagNumber(9)
  $1.Struct get attributes => $_getN(8);
  @$pb.TagNumber(9)
  set attributes($1.Struct value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasAttributes() => $_has(8);
  @$pb.TagNumber(9)
  void clearAttributes() => $_clearField(9);
  @$pb.TagNumber(9)
  $1.Struct ensureAttributes() => $_ensure(8);

  @$pb.TagNumber(10)
  $core.int get sortOrder => $_getIZ(9);
  @$pb.TagNumber(10)
  set sortOrder($core.int value) => $_setSignedInt32(9, value);
  @$pb.TagNumber(10)
  $core.bool hasSortOrder() => $_has(9);
  @$pb.TagNumber(10)
  void clearSortOrder() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.bool get isActive => $_getBF(10);
  @$pb.TagNumber(11)
  set isActive($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasIsActive() => $_has(10);
  @$pb.TagNumber(11)
  void clearIsActive() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get createdAt => $_getSZ(11);
  @$pb.TagNumber(12)
  set createdAt($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasCreatedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearCreatedAt() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get updatedAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set updatedAt($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasUpdatedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearUpdatedAt() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.String get deletedAt => $_getSZ(13);
  @$pb.TagNumber(14)
  set deletedAt($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasDeletedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearDeletedAt() => $_clearField(14);
}

class ListProcessingSpecsRequest extends $pb.GeneratedMessage {
  factory ListProcessingSpecsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.bool? includeDeleted,
  }) {
    final result = ListProcessingSpecsRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListProcessingSpecsRequest._();

  factory ListProcessingSpecsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProcessingSpecsRequest()..mergeFromBuffer(data, registry);
  factory ListProcessingSpecsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProcessingSpecsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProcessingSpecsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListProcessingSpecsRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProcessingSpecsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProcessingSpecsRequest copyWith(
          void Function(ListProcessingSpecsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListProcessingSpecsRequest))
          as ListProcessingSpecsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListProcessingSpecsRequest() / ListProcessingSpecsRequest.new instead')
  static ListProcessingSpecsRequest create() => ListProcessingSpecsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListProcessingSpecsRequest._();
  @$core.override
  ListProcessingSpecsRequest createEmptyInstance() =>
      ListProcessingSpecsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListProcessingSpecsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProcessingSpecsRequest>(
          ListProcessingSpecsRequest.$_createMessage);
  static ListProcessingSpecsRequest? _defaultInstance;

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
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(4)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(4)
  void clearIncludeDeleted() => $_clearField(4);
}

class ListProcessingSpecsResponse extends $pb.GeneratedMessage {
  factory ListProcessingSpecsResponse({
    $core.Iterable<ProcessingSpec>? processingSpecs,
    $0.Pagination? pagination,
  }) {
    final result = ListProcessingSpecsResponse._();
    if (processingSpecs != null) result.processingSpecs.addAll(processingSpecs);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListProcessingSpecsResponse._();

  factory ListProcessingSpecsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProcessingSpecsResponse()..mergeFromBuffer(data, registry);
  factory ListProcessingSpecsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProcessingSpecsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProcessingSpecsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListProcessingSpecsResponse.$_createMessage)
    ..pPM<ProcessingSpec>(1, _omitFieldNames ? '' : 'processingSpecs',
        subBuilder: ProcessingSpec.$_createMessage)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProcessingSpecsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProcessingSpecsResponse copyWith(
          void Function(ListProcessingSpecsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListProcessingSpecsResponse))
          as ListProcessingSpecsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListProcessingSpecsResponse() / ListProcessingSpecsResponse.new instead')
  static ListProcessingSpecsResponse create() =>
      ListProcessingSpecsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListProcessingSpecsResponse._();
  @$core.override
  ListProcessingSpecsResponse createEmptyInstance() =>
      ListProcessingSpecsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListProcessingSpecsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProcessingSpecsResponse>(
          ListProcessingSpecsResponse.$_createMessage);
  static ListProcessingSpecsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ProcessingSpec> get processingSpecs => $_getList(0);

  @$pb.TagNumber(2)
  $0.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($0.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Pagination ensurePagination() => $_ensure(1);
}

class CreateProcessingSpecRequest extends $pb.GeneratedMessage {
  factory CreateProcessingSpecRequest({
    $core.String? code,
    $core.String? name,
    $core.String? kind,
    $core.bool? appliesToProcessing,
    $core.bool? appliesToPicking,
    $1.Struct? attributes,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = CreateProcessingSpecRequest._();
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (kind != null) result.kind = kind;
    if (appliesToProcessing != null)
      result.appliesToProcessing = appliesToProcessing;
    if (appliesToPicking != null) result.appliesToPicking = appliesToPicking;
    if (attributes != null) result.attributes = attributes;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  CreateProcessingSpecRequest._();

  factory CreateProcessingSpecRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProcessingSpecRequest()..mergeFromBuffer(data, registry);
  factory CreateProcessingSpecRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProcessingSpecRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateProcessingSpecRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateProcessingSpecRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'kind')
    ..aOB(4, _omitFieldNames ? '' : 'appliesToProcessing')
    ..aOB(5, _omitFieldNames ? '' : 'appliesToPicking')
    ..aOM<$1.Struct>(6, _omitFieldNames ? '' : 'attributes',
        subBuilder: $1.Struct.$_createMessage)
    ..aI(7, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(8, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProcessingSpecRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProcessingSpecRequest copyWith(
          void Function(CreateProcessingSpecRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateProcessingSpecRequest))
          as CreateProcessingSpecRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateProcessingSpecRequest() / CreateProcessingSpecRequest.new instead')
  static CreateProcessingSpecRequest create() =>
      CreateProcessingSpecRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateProcessingSpecRequest._();
  @$core.override
  CreateProcessingSpecRequest createEmptyInstance() =>
      CreateProcessingSpecRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateProcessingSpecRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateProcessingSpecRequest>(
          CreateProcessingSpecRequest.$_createMessage);
  static CreateProcessingSpecRequest? _defaultInstance;

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
  $core.String get kind => $_getSZ(2);
  @$pb.TagNumber(3)
  set kind($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasKind() => $_has(2);
  @$pb.TagNumber(3)
  void clearKind() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get appliesToProcessing => $_getBF(3);
  @$pb.TagNumber(4)
  set appliesToProcessing($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAppliesToProcessing() => $_has(3);
  @$pb.TagNumber(4)
  void clearAppliesToProcessing() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get appliesToPicking => $_getBF(4);
  @$pb.TagNumber(5)
  set appliesToPicking($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAppliesToPicking() => $_has(4);
  @$pb.TagNumber(5)
  void clearAppliesToPicking() => $_clearField(5);

  @$pb.TagNumber(6)
  $1.Struct get attributes => $_getN(5);
  @$pb.TagNumber(6)
  set attributes($1.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasAttributes() => $_has(5);
  @$pb.TagNumber(6)
  void clearAttributes() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Struct ensureAttributes() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.int get sortOrder => $_getIZ(6);
  @$pb.TagNumber(7)
  set sortOrder($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasSortOrder() => $_has(6);
  @$pb.TagNumber(7)
  void clearSortOrder() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get isActive => $_getBF(7);
  @$pb.TagNumber(8)
  set isActive($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasIsActive() => $_has(7);
  @$pb.TagNumber(8)
  void clearIsActive() => $_clearField(8);
}

class CreateProcessingSpecResponse extends $pb.GeneratedMessage {
  factory CreateProcessingSpecResponse({
    ProcessingSpec? processingSpec,
  }) {
    final result = CreateProcessingSpecResponse._();
    if (processingSpec != null) result.processingSpec = processingSpec;
    return result;
  }

  CreateProcessingSpecResponse._();

  factory CreateProcessingSpecResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProcessingSpecResponse()..mergeFromBuffer(data, registry);
  factory CreateProcessingSpecResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProcessingSpecResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateProcessingSpecResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateProcessingSpecResponse.$_createMessage)
    ..aOM<ProcessingSpec>(1, _omitFieldNames ? '' : 'processingSpec',
        subBuilder: ProcessingSpec.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProcessingSpecResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProcessingSpecResponse copyWith(
          void Function(CreateProcessingSpecResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateProcessingSpecResponse))
          as CreateProcessingSpecResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateProcessingSpecResponse() / CreateProcessingSpecResponse.new instead')
  static CreateProcessingSpecResponse create() =>
      CreateProcessingSpecResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateProcessingSpecResponse._();
  @$core.override
  CreateProcessingSpecResponse createEmptyInstance() =>
      CreateProcessingSpecResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateProcessingSpecResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateProcessingSpecResponse>(
          CreateProcessingSpecResponse.$_createMessage);
  static CreateProcessingSpecResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ProcessingSpec get processingSpec => $_getN(0);
  @$pb.TagNumber(1)
  set processingSpec(ProcessingSpec value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProcessingSpec() => $_has(0);
  @$pb.TagNumber(1)
  void clearProcessingSpec() => $_clearField(1);
  @$pb.TagNumber(1)
  ProcessingSpec ensureProcessingSpec() => $_ensure(0);
}

class UpdateProcessingSpecRequest extends $pb.GeneratedMessage {
  factory UpdateProcessingSpecRequest({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.String? kind,
    $core.bool? appliesToProcessing,
    $core.bool? appliesToPicking,
    $1.Struct? attributes,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = UpdateProcessingSpecRequest._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (kind != null) result.kind = kind;
    if (appliesToProcessing != null)
      result.appliesToProcessing = appliesToProcessing;
    if (appliesToPicking != null) result.appliesToPicking = appliesToPicking;
    if (attributes != null) result.attributes = attributes;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  UpdateProcessingSpecRequest._();

  factory UpdateProcessingSpecRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProcessingSpecRequest()..mergeFromBuffer(data, registry);
  factory UpdateProcessingSpecRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProcessingSpecRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProcessingSpecRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateProcessingSpecRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'kind')
    ..aOB(5, _omitFieldNames ? '' : 'appliesToProcessing')
    ..aOB(6, _omitFieldNames ? '' : 'appliesToPicking')
    ..aOM<$1.Struct>(7, _omitFieldNames ? '' : 'attributes',
        subBuilder: $1.Struct.$_createMessage)
    ..aI(8, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(9, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProcessingSpecRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProcessingSpecRequest copyWith(
          void Function(UpdateProcessingSpecRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateProcessingSpecRequest))
          as UpdateProcessingSpecRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProcessingSpecRequest() / UpdateProcessingSpecRequest.new instead')
  static UpdateProcessingSpecRequest create() =>
      UpdateProcessingSpecRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateProcessingSpecRequest._();
  @$core.override
  UpdateProcessingSpecRequest createEmptyInstance() =>
      UpdateProcessingSpecRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateProcessingSpecRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProcessingSpecRequest>(
          UpdateProcessingSpecRequest.$_createMessage);
  static UpdateProcessingSpecRequest? _defaultInstance;

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
  $core.String get kind => $_getSZ(3);
  @$pb.TagNumber(4)
  set kind($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasKind() => $_has(3);
  @$pb.TagNumber(4)
  void clearKind() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get appliesToProcessing => $_getBF(4);
  @$pb.TagNumber(5)
  set appliesToProcessing($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAppliesToProcessing() => $_has(4);
  @$pb.TagNumber(5)
  void clearAppliesToProcessing() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get appliesToPicking => $_getBF(5);
  @$pb.TagNumber(6)
  set appliesToPicking($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAppliesToPicking() => $_has(5);
  @$pb.TagNumber(6)
  void clearAppliesToPicking() => $_clearField(6);

  @$pb.TagNumber(7)
  $1.Struct get attributes => $_getN(6);
  @$pb.TagNumber(7)
  set attributes($1.Struct value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasAttributes() => $_has(6);
  @$pb.TagNumber(7)
  void clearAttributes() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Struct ensureAttributes() => $_ensure(6);

  @$pb.TagNumber(8)
  $core.int get sortOrder => $_getIZ(7);
  @$pb.TagNumber(8)
  set sortOrder($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasSortOrder() => $_has(7);
  @$pb.TagNumber(8)
  void clearSortOrder() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.bool get isActive => $_getBF(8);
  @$pb.TagNumber(9)
  set isActive($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasIsActive() => $_has(8);
  @$pb.TagNumber(9)
  void clearIsActive() => $_clearField(9);
}

class UpdateProcessingSpecResponse extends $pb.GeneratedMessage {
  factory UpdateProcessingSpecResponse({
    ProcessingSpec? processingSpec,
  }) {
    final result = UpdateProcessingSpecResponse._();
    if (processingSpec != null) result.processingSpec = processingSpec;
    return result;
  }

  UpdateProcessingSpecResponse._();

  factory UpdateProcessingSpecResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProcessingSpecResponse()..mergeFromBuffer(data, registry);
  factory UpdateProcessingSpecResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProcessingSpecResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProcessingSpecResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateProcessingSpecResponse.$_createMessage)
    ..aOM<ProcessingSpec>(1, _omitFieldNames ? '' : 'processingSpec',
        subBuilder: ProcessingSpec.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProcessingSpecResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProcessingSpecResponse copyWith(
          void Function(UpdateProcessingSpecResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateProcessingSpecResponse))
          as UpdateProcessingSpecResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProcessingSpecResponse() / UpdateProcessingSpecResponse.new instead')
  static UpdateProcessingSpecResponse create() =>
      UpdateProcessingSpecResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateProcessingSpecResponse._();
  @$core.override
  UpdateProcessingSpecResponse createEmptyInstance() =>
      UpdateProcessingSpecResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateProcessingSpecResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProcessingSpecResponse>(
          UpdateProcessingSpecResponse.$_createMessage);
  static UpdateProcessingSpecResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ProcessingSpec get processingSpec => $_getN(0);
  @$pb.TagNumber(1)
  set processingSpec(ProcessingSpec value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProcessingSpec() => $_has(0);
  @$pb.TagNumber(1)
  void clearProcessingSpec() => $_clearField(1);
  @$pb.TagNumber(1)
  ProcessingSpec ensureProcessingSpec() => $_ensure(0);
}

class DeleteProcessingSpecRequest extends $pb.GeneratedMessage {
  factory DeleteProcessingSpecRequest({
    $core.String? id,
  }) {
    final result = DeleteProcessingSpecRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteProcessingSpecRequest._();

  factory DeleteProcessingSpecRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProcessingSpecRequest()..mergeFromBuffer(data, registry);
  factory DeleteProcessingSpecRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProcessingSpecRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteProcessingSpecRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteProcessingSpecRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProcessingSpecRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProcessingSpecRequest copyWith(
          void Function(DeleteProcessingSpecRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteProcessingSpecRequest))
          as DeleteProcessingSpecRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteProcessingSpecRequest() / DeleteProcessingSpecRequest.new instead')
  static DeleteProcessingSpecRequest create() =>
      DeleteProcessingSpecRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteProcessingSpecRequest._();
  @$core.override
  DeleteProcessingSpecRequest createEmptyInstance() =>
      DeleteProcessingSpecRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteProcessingSpecRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteProcessingSpecRequest>(
          DeleteProcessingSpecRequest.$_createMessage);
  static DeleteProcessingSpecRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteProcessingSpecResponse extends $pb.GeneratedMessage {
  factory DeleteProcessingSpecResponse() => DeleteProcessingSpecResponse._();

  DeleteProcessingSpecResponse._();

  factory DeleteProcessingSpecResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProcessingSpecResponse()..mergeFromBuffer(data, registry);
  factory DeleteProcessingSpecResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProcessingSpecResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteProcessingSpecResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteProcessingSpecResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProcessingSpecResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProcessingSpecResponse copyWith(
          void Function(DeleteProcessingSpecResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteProcessingSpecResponse))
          as DeleteProcessingSpecResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteProcessingSpecResponse() / DeleteProcessingSpecResponse.new instead')
  static DeleteProcessingSpecResponse create() =>
      DeleteProcessingSpecResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteProcessingSpecResponse._();
  @$core.override
  DeleteProcessingSpecResponse createEmptyInstance() =>
      DeleteProcessingSpecResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteProcessingSpecResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteProcessingSpecResponse>(
          DeleteProcessingSpecResponse.$_createMessage);
  static DeleteProcessingSpecResponse? _defaultInstance;
}

class RestoreProcessingSpecRequest extends $pb.GeneratedMessage {
  factory RestoreProcessingSpecRequest({
    $core.String? id,
  }) {
    final result = RestoreProcessingSpecRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RestoreProcessingSpecRequest._();

  factory RestoreProcessingSpecRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProcessingSpecRequest()..mergeFromBuffer(data, registry);
  factory RestoreProcessingSpecRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProcessingSpecRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreProcessingSpecRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreProcessingSpecRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProcessingSpecRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProcessingSpecRequest copyWith(
          void Function(RestoreProcessingSpecRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RestoreProcessingSpecRequest))
          as RestoreProcessingSpecRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreProcessingSpecRequest() / RestoreProcessingSpecRequest.new instead')
  static RestoreProcessingSpecRequest create() =>
      RestoreProcessingSpecRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RestoreProcessingSpecRequest._();
  @$core.override
  RestoreProcessingSpecRequest createEmptyInstance() =>
      RestoreProcessingSpecRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreProcessingSpecRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreProcessingSpecRequest>(
          RestoreProcessingSpecRequest.$_createMessage);
  static RestoreProcessingSpecRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RestoreProcessingSpecResponse extends $pb.GeneratedMessage {
  factory RestoreProcessingSpecResponse({
    ProcessingSpec? processingSpec,
  }) {
    final result = RestoreProcessingSpecResponse._();
    if (processingSpec != null) result.processingSpec = processingSpec;
    return result;
  }

  RestoreProcessingSpecResponse._();

  factory RestoreProcessingSpecResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProcessingSpecResponse()..mergeFromBuffer(data, registry);
  factory RestoreProcessingSpecResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProcessingSpecResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreProcessingSpecResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreProcessingSpecResponse.$_createMessage)
    ..aOM<ProcessingSpec>(1, _omitFieldNames ? '' : 'processingSpec',
        subBuilder: ProcessingSpec.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProcessingSpecResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProcessingSpecResponse copyWith(
          void Function(RestoreProcessingSpecResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RestoreProcessingSpecResponse))
          as RestoreProcessingSpecResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreProcessingSpecResponse() / RestoreProcessingSpecResponse.new instead')
  static RestoreProcessingSpecResponse create() =>
      RestoreProcessingSpecResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RestoreProcessingSpecResponse._();
  @$core.override
  RestoreProcessingSpecResponse createEmptyInstance() =>
      RestoreProcessingSpecResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreProcessingSpecResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreProcessingSpecResponse>(
          RestoreProcessingSpecResponse.$_createMessage);
  static RestoreProcessingSpecResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ProcessingSpec get processingSpec => $_getN(0);
  @$pb.TagNumber(1)
  set processingSpec(ProcessingSpec value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProcessingSpec() => $_has(0);
  @$pb.TagNumber(1)
  void clearProcessingSpec() => $_clearField(1);
  @$pb.TagNumber(1)
  ProcessingSpec ensureProcessingSpec() => $_ensure(0);
}

/// ---- ProductCategory(3.4.4) ----
class ProductCategory extends $pb.GeneratedMessage {
  factory ProductCategory({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? code,
    $core.String? name,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = ProductCategory._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  ProductCategory._();

  factory ProductCategory.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProductCategory()..mergeFromBuffer(data, registry);
  factory ProductCategory.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProductCategory()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProductCategory',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ProductCategory.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'code')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aI(6, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(7, _omitFieldNames ? '' : 'isActive')
    ..aOS(8, _omitFieldNames ? '' : 'createdAt')
    ..aOS(9, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(10, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProductCategory clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProductCategory copyWith(void Function(ProductCategory) updates) =>
      super.copyWith((message) => updates(message as ProductCategory))
          as ProductCategory;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProductCategory() / ProductCategory.new instead')
  static ProductCategory create() => ProductCategory._();
  static $pb.GeneratedMessage $_createMessage() => ProductCategory._();
  @$core.override
  ProductCategory createEmptyInstance() => ProductCategory._();
  @$core.pragma('dart2js:noInline')
  static ProductCategory getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProductCategory>(
          ProductCategory.$_createMessage);
  static ProductCategory? _defaultInstance;

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
  $core.int get sortOrder => $_getIZ(5);
  @$pb.TagNumber(6)
  set sortOrder($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSortOrder() => $_has(5);
  @$pb.TagNumber(6)
  void clearSortOrder() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get isActive => $_getBF(6);
  @$pb.TagNumber(7)
  set isActive($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasIsActive() => $_has(6);
  @$pb.TagNumber(7)
  void clearIsActive() => $_clearField(7);

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

  @$pb.TagNumber(10)
  $core.String get deletedAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set deletedAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasDeletedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearDeletedAt() => $_clearField(10);
}

class ListProductCategoriesRequest extends $pb.GeneratedMessage {
  factory ListProductCategoriesRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.bool? includeDeleted,
  }) {
    final result = ListProductCategoriesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListProductCategoriesRequest._();

  factory ListProductCategoriesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductCategoriesRequest()..mergeFromBuffer(data, registry);
  factory ListProductCategoriesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductCategoriesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProductCategoriesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListProductCategoriesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductCategoriesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductCategoriesRequest copyWith(
          void Function(ListProductCategoriesRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListProductCategoriesRequest))
          as ListProductCategoriesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListProductCategoriesRequest() / ListProductCategoriesRequest.new instead')
  static ListProductCategoriesRequest create() =>
      ListProductCategoriesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListProductCategoriesRequest._();
  @$core.override
  ListProductCategoriesRequest createEmptyInstance() =>
      ListProductCategoriesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListProductCategoriesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProductCategoriesRequest>(
          ListProductCategoriesRequest.$_createMessage);
  static ListProductCategoriesRequest? _defaultInstance;

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
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(4)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(4)
  void clearIncludeDeleted() => $_clearField(4);
}

class ListProductCategoriesResponse extends $pb.GeneratedMessage {
  factory ListProductCategoriesResponse({
    $core.Iterable<ProductCategory>? productCategories,
    $0.Pagination? pagination,
  }) {
    final result = ListProductCategoriesResponse._();
    if (productCategories != null)
      result.productCategories.addAll(productCategories);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListProductCategoriesResponse._();

  factory ListProductCategoriesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductCategoriesResponse()..mergeFromBuffer(data, registry);
  factory ListProductCategoriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProductCategoriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProductCategoriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: ListProductCategoriesResponse.$_createMessage)
    ..pPM<ProductCategory>(1, _omitFieldNames ? '' : 'productCategories',
        subBuilder: ProductCategory.$_createMessage)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductCategoriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProductCategoriesResponse copyWith(
          void Function(ListProductCategoriesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListProductCategoriesResponse))
          as ListProductCategoriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListProductCategoriesResponse() / ListProductCategoriesResponse.new instead')
  static ListProductCategoriesResponse create() =>
      ListProductCategoriesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListProductCategoriesResponse._();
  @$core.override
  ListProductCategoriesResponse createEmptyInstance() =>
      ListProductCategoriesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListProductCategoriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProductCategoriesResponse>(
          ListProductCategoriesResponse.$_createMessage);
  static ListProductCategoriesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ProductCategory> get productCategories => $_getList(0);

  @$pb.TagNumber(2)
  $0.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($0.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Pagination ensurePagination() => $_ensure(1);
}

class CreateProductCategoryRequest extends $pb.GeneratedMessage {
  factory CreateProductCategoryRequest({
    $core.String? code,
    $core.String? name,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = CreateProductCategoryRequest._();
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  CreateProductCategoryRequest._();

  factory CreateProductCategoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductCategoryRequest()..mergeFromBuffer(data, registry);
  factory CreateProductCategoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductCategoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateProductCategoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateProductCategoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aI(3, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(4, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductCategoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductCategoryRequest copyWith(
          void Function(CreateProductCategoryRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateProductCategoryRequest))
          as CreateProductCategoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateProductCategoryRequest() / CreateProductCategoryRequest.new instead')
  static CreateProductCategoryRequest create() =>
      CreateProductCategoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateProductCategoryRequest._();
  @$core.override
  CreateProductCategoryRequest createEmptyInstance() =>
      CreateProductCategoryRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateProductCategoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateProductCategoryRequest>(
          CreateProductCategoryRequest.$_createMessage);
  static CreateProductCategoryRequest? _defaultInstance;

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
  $core.int get sortOrder => $_getIZ(2);
  @$pb.TagNumber(3)
  set sortOrder($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSortOrder() => $_has(2);
  @$pb.TagNumber(3)
  void clearSortOrder() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get isActive => $_getBF(3);
  @$pb.TagNumber(4)
  set isActive($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsActive() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsActive() => $_clearField(4);
}

class CreateProductCategoryResponse extends $pb.GeneratedMessage {
  factory CreateProductCategoryResponse({
    ProductCategory? productCategory,
  }) {
    final result = CreateProductCategoryResponse._();
    if (productCategory != null) result.productCategory = productCategory;
    return result;
  }

  CreateProductCategoryResponse._();

  factory CreateProductCategoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductCategoryResponse()..mergeFromBuffer(data, registry);
  factory CreateProductCategoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateProductCategoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateProductCategoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: CreateProductCategoryResponse.$_createMessage)
    ..aOM<ProductCategory>(1, _omitFieldNames ? '' : 'productCategory',
        subBuilder: ProductCategory.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductCategoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateProductCategoryResponse copyWith(
          void Function(CreateProductCategoryResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateProductCategoryResponse))
          as CreateProductCategoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateProductCategoryResponse() / CreateProductCategoryResponse.new instead')
  static CreateProductCategoryResponse create() =>
      CreateProductCategoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateProductCategoryResponse._();
  @$core.override
  CreateProductCategoryResponse createEmptyInstance() =>
      CreateProductCategoryResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateProductCategoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateProductCategoryResponse>(
          CreateProductCategoryResponse.$_createMessage);
  static CreateProductCategoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ProductCategory get productCategory => $_getN(0);
  @$pb.TagNumber(1)
  set productCategory(ProductCategory value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProductCategory() => $_has(0);
  @$pb.TagNumber(1)
  void clearProductCategory() => $_clearField(1);
  @$pb.TagNumber(1)
  ProductCategory ensureProductCategory() => $_ensure(0);
}

class UpdateProductCategoryRequest extends $pb.GeneratedMessage {
  factory UpdateProductCategoryRequest({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = UpdateProductCategoryRequest._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  UpdateProductCategoryRequest._();

  factory UpdateProductCategoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductCategoryRequest()..mergeFromBuffer(data, registry);
  factory UpdateProductCategoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductCategoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProductCategoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateProductCategoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aI(4, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(5, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductCategoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductCategoryRequest copyWith(
          void Function(UpdateProductCategoryRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateProductCategoryRequest))
          as UpdateProductCategoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProductCategoryRequest() / UpdateProductCategoryRequest.new instead')
  static UpdateProductCategoryRequest create() =>
      UpdateProductCategoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateProductCategoryRequest._();
  @$core.override
  UpdateProductCategoryRequest createEmptyInstance() =>
      UpdateProductCategoryRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateProductCategoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProductCategoryRequest>(
          UpdateProductCategoryRequest.$_createMessage);
  static UpdateProductCategoryRequest? _defaultInstance;

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
  $core.int get sortOrder => $_getIZ(3);
  @$pb.TagNumber(4)
  set sortOrder($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSortOrder() => $_has(3);
  @$pb.TagNumber(4)
  void clearSortOrder() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get isActive => $_getBF(4);
  @$pb.TagNumber(5)
  set isActive($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIsActive() => $_has(4);
  @$pb.TagNumber(5)
  void clearIsActive() => $_clearField(5);
}

class UpdateProductCategoryResponse extends $pb.GeneratedMessage {
  factory UpdateProductCategoryResponse({
    ProductCategory? productCategory,
  }) {
    final result = UpdateProductCategoryResponse._();
    if (productCategory != null) result.productCategory = productCategory;
    return result;
  }

  UpdateProductCategoryResponse._();

  factory UpdateProductCategoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductCategoryResponse()..mergeFromBuffer(data, registry);
  factory UpdateProductCategoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProductCategoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProductCategoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: UpdateProductCategoryResponse.$_createMessage)
    ..aOM<ProductCategory>(1, _omitFieldNames ? '' : 'productCategory',
        subBuilder: ProductCategory.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductCategoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProductCategoryResponse copyWith(
          void Function(UpdateProductCategoryResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateProductCategoryResponse))
          as UpdateProductCategoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProductCategoryResponse() / UpdateProductCategoryResponse.new instead')
  static UpdateProductCategoryResponse create() =>
      UpdateProductCategoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateProductCategoryResponse._();
  @$core.override
  UpdateProductCategoryResponse createEmptyInstance() =>
      UpdateProductCategoryResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateProductCategoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProductCategoryResponse>(
          UpdateProductCategoryResponse.$_createMessage);
  static UpdateProductCategoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ProductCategory get productCategory => $_getN(0);
  @$pb.TagNumber(1)
  set productCategory(ProductCategory value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProductCategory() => $_has(0);
  @$pb.TagNumber(1)
  void clearProductCategory() => $_clearField(1);
  @$pb.TagNumber(1)
  ProductCategory ensureProductCategory() => $_ensure(0);
}

class DeleteProductCategoryRequest extends $pb.GeneratedMessage {
  factory DeleteProductCategoryRequest({
    $core.String? id,
  }) {
    final result = DeleteProductCategoryRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteProductCategoryRequest._();

  factory DeleteProductCategoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductCategoryRequest()..mergeFromBuffer(data, registry);
  factory DeleteProductCategoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductCategoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteProductCategoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteProductCategoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductCategoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductCategoryRequest copyWith(
          void Function(DeleteProductCategoryRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteProductCategoryRequest))
          as DeleteProductCategoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteProductCategoryRequest() / DeleteProductCategoryRequest.new instead')
  static DeleteProductCategoryRequest create() =>
      DeleteProductCategoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteProductCategoryRequest._();
  @$core.override
  DeleteProductCategoryRequest createEmptyInstance() =>
      DeleteProductCategoryRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteProductCategoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteProductCategoryRequest>(
          DeleteProductCategoryRequest.$_createMessage);
  static DeleteProductCategoryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteProductCategoryResponse extends $pb.GeneratedMessage {
  factory DeleteProductCategoryResponse() => DeleteProductCategoryResponse._();

  DeleteProductCategoryResponse._();

  factory DeleteProductCategoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductCategoryResponse()..mergeFromBuffer(data, registry);
  factory DeleteProductCategoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteProductCategoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteProductCategoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: DeleteProductCategoryResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductCategoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteProductCategoryResponse copyWith(
          void Function(DeleteProductCategoryResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteProductCategoryResponse))
          as DeleteProductCategoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteProductCategoryResponse() / DeleteProductCategoryResponse.new instead')
  static DeleteProductCategoryResponse create() =>
      DeleteProductCategoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteProductCategoryResponse._();
  @$core.override
  DeleteProductCategoryResponse createEmptyInstance() =>
      DeleteProductCategoryResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteProductCategoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteProductCategoryResponse>(
          DeleteProductCategoryResponse.$_createMessage);
  static DeleteProductCategoryResponse? _defaultInstance;
}

class RestoreProductCategoryRequest extends $pb.GeneratedMessage {
  factory RestoreProductCategoryRequest({
    $core.String? id,
  }) {
    final result = RestoreProductCategoryRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RestoreProductCategoryRequest._();

  factory RestoreProductCategoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductCategoryRequest()..mergeFromBuffer(data, registry);
  factory RestoreProductCategoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductCategoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreProductCategoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreProductCategoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductCategoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductCategoryRequest copyWith(
          void Function(RestoreProductCategoryRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RestoreProductCategoryRequest))
          as RestoreProductCategoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreProductCategoryRequest() / RestoreProductCategoryRequest.new instead')
  static RestoreProductCategoryRequest create() =>
      RestoreProductCategoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RestoreProductCategoryRequest._();
  @$core.override
  RestoreProductCategoryRequest createEmptyInstance() =>
      RestoreProductCategoryRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreProductCategoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreProductCategoryRequest>(
          RestoreProductCategoryRequest.$_createMessage);
  static RestoreProductCategoryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RestoreProductCategoryResponse extends $pb.GeneratedMessage {
  factory RestoreProductCategoryResponse({
    ProductCategory? productCategory,
  }) {
    final result = RestoreProductCategoryResponse._();
    if (productCategory != null) result.productCategory = productCategory;
    return result;
  }

  RestoreProductCategoryResponse._();

  factory RestoreProductCategoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductCategoryResponse()..mergeFromBuffer(data, registry);
  factory RestoreProductCategoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreProductCategoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreProductCategoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'masters.v1'),
      createEmptyInstance: RestoreProductCategoryResponse.$_createMessage)
    ..aOM<ProductCategory>(1, _omitFieldNames ? '' : 'productCategory',
        subBuilder: ProductCategory.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductCategoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreProductCategoryResponse copyWith(
          void Function(RestoreProductCategoryResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RestoreProductCategoryResponse))
          as RestoreProductCategoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreProductCategoryResponse() / RestoreProductCategoryResponse.new instead')
  static RestoreProductCategoryResponse create() =>
      RestoreProductCategoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RestoreProductCategoryResponse._();
  @$core.override
  RestoreProductCategoryResponse createEmptyInstance() =>
      RestoreProductCategoryResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreProductCategoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreProductCategoryResponse>(
          RestoreProductCategoryResponse.$_createMessage);
  static RestoreProductCategoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ProductCategory get productCategory => $_getN(0);
  @$pb.TagNumber(1)
  set productCategory(ProductCategory value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProductCategory() => $_has(0);
  @$pb.TagNumber(1)
  void clearProductCategory() => $_clearField(1);
  @$pb.TagNumber(1)
  ProductCategory ensureProductCategory() => $_ensure(0);
}

class WarehouseServiceApi {
  final $pb.RpcClient _client;

  WarehouseServiceApi(this._client);

  $async.Future<ListWarehousesResponse> listWarehouses(
          $pb.ClientContext? ctx, ListWarehousesRequest request) =>
      _client.invoke<ListWarehousesResponse>(ctx, 'WarehouseService',
          'ListWarehouses', request, ListWarehousesResponse());
  $async.Future<CreateWarehouseResponse> createWarehouse(
          $pb.ClientContext? ctx, CreateWarehouseRequest request) =>
      _client.invoke<CreateWarehouseResponse>(ctx, 'WarehouseService',
          'CreateWarehouse', request, CreateWarehouseResponse());
  $async.Future<UpdateWarehouseResponse> updateWarehouse(
          $pb.ClientContext? ctx, UpdateWarehouseRequest request) =>
      _client.invoke<UpdateWarehouseResponse>(ctx, 'WarehouseService',
          'UpdateWarehouse', request, UpdateWarehouseResponse());
  $async.Future<DeleteWarehouseResponse> deleteWarehouse(
          $pb.ClientContext? ctx, DeleteWarehouseRequest request) =>
      _client.invoke<DeleteWarehouseResponse>(ctx, 'WarehouseService',
          'DeleteWarehouse', request, DeleteWarehouseResponse());
  $async.Future<RestoreWarehouseResponse> restoreWarehouse(
          $pb.ClientContext? ctx, RestoreWarehouseRequest request) =>
      _client.invoke<RestoreWarehouseResponse>(ctx, 'WarehouseService',
          'RestoreWarehouse', request, RestoreWarehouseResponse());
}

class RouteServiceApi {
  final $pb.RpcClient _client;

  RouteServiceApi(this._client);

  $async.Future<ListRoutesResponse> listRoutes(
          $pb.ClientContext? ctx, ListRoutesRequest request) =>
      _client.invoke<ListRoutesResponse>(
          ctx, 'RouteService', 'ListRoutes', request, ListRoutesResponse());
  $async.Future<CreateRouteResponse> createRoute(
          $pb.ClientContext? ctx, CreateRouteRequest request) =>
      _client.invoke<CreateRouteResponse>(
          ctx, 'RouteService', 'CreateRoute', request, CreateRouteResponse());
  $async.Future<UpdateRouteResponse> updateRoute(
          $pb.ClientContext? ctx, UpdateRouteRequest request) =>
      _client.invoke<UpdateRouteResponse>(
          ctx, 'RouteService', 'UpdateRoute', request, UpdateRouteResponse());
  $async.Future<DeleteRouteResponse> deleteRoute(
          $pb.ClientContext? ctx, DeleteRouteRequest request) =>
      _client.invoke<DeleteRouteResponse>(
          ctx, 'RouteService', 'DeleteRoute', request, DeleteRouteResponse());
  $async.Future<RestoreRouteResponse> restoreRoute(
          $pb.ClientContext? ctx, RestoreRouteRequest request) =>
      _client.invoke<RestoreRouteResponse>(
          ctx, 'RouteService', 'RestoreRoute', request, RestoreRouteResponse());
}

class ProcessingSpecServiceApi {
  final $pb.RpcClient _client;

  ProcessingSpecServiceApi(this._client);

  $async.Future<ListProcessingSpecsResponse> listProcessingSpecs(
          $pb.ClientContext? ctx, ListProcessingSpecsRequest request) =>
      _client.invoke<ListProcessingSpecsResponse>(ctx, 'ProcessingSpecService',
          'ListProcessingSpecs', request, ListProcessingSpecsResponse());
  $async.Future<CreateProcessingSpecResponse> createProcessingSpec(
          $pb.ClientContext? ctx, CreateProcessingSpecRequest request) =>
      _client.invoke<CreateProcessingSpecResponse>(ctx, 'ProcessingSpecService',
          'CreateProcessingSpec', request, CreateProcessingSpecResponse());
  $async.Future<UpdateProcessingSpecResponse> updateProcessingSpec(
          $pb.ClientContext? ctx, UpdateProcessingSpecRequest request) =>
      _client.invoke<UpdateProcessingSpecResponse>(ctx, 'ProcessingSpecService',
          'UpdateProcessingSpec', request, UpdateProcessingSpecResponse());
  $async.Future<DeleteProcessingSpecResponse> deleteProcessingSpec(
          $pb.ClientContext? ctx, DeleteProcessingSpecRequest request) =>
      _client.invoke<DeleteProcessingSpecResponse>(ctx, 'ProcessingSpecService',
          'DeleteProcessingSpec', request, DeleteProcessingSpecResponse());
  $async.Future<RestoreProcessingSpecResponse> restoreProcessingSpec(
          $pb.ClientContext? ctx, RestoreProcessingSpecRequest request) =>
      _client.invoke<RestoreProcessingSpecResponse>(
          ctx,
          'ProcessingSpecService',
          'RestoreProcessingSpec',
          request,
          RestoreProcessingSpecResponse());
}

class ProductCategoryServiceApi {
  final $pb.RpcClient _client;

  ProductCategoryServiceApi(this._client);

  $async.Future<ListProductCategoriesResponse> listProductCategories(
          $pb.ClientContext? ctx, ListProductCategoriesRequest request) =>
      _client.invoke<ListProductCategoriesResponse>(
          ctx,
          'ProductCategoryService',
          'ListProductCategories',
          request,
          ListProductCategoriesResponse());
  $async.Future<CreateProductCategoryResponse> createProductCategory(
          $pb.ClientContext? ctx, CreateProductCategoryRequest request) =>
      _client.invoke<CreateProductCategoryResponse>(
          ctx,
          'ProductCategoryService',
          'CreateProductCategory',
          request,
          CreateProductCategoryResponse());
  $async.Future<UpdateProductCategoryResponse> updateProductCategory(
          $pb.ClientContext? ctx, UpdateProductCategoryRequest request) =>
      _client.invoke<UpdateProductCategoryResponse>(
          ctx,
          'ProductCategoryService',
          'UpdateProductCategory',
          request,
          UpdateProductCategoryResponse());
  $async.Future<DeleteProductCategoryResponse> deleteProductCategory(
          $pb.ClientContext? ctx, DeleteProductCategoryRequest request) =>
      _client.invoke<DeleteProductCategoryResponse>(
          ctx,
          'ProductCategoryService',
          'DeleteProductCategory',
          request,
          DeleteProductCategoryResponse());
  $async.Future<RestoreProductCategoryResponse> restoreProductCategory(
          $pb.ClientContext? ctx, RestoreProductCategoryRequest request) =>
      _client.invoke<RestoreProductCategoryResponse>(
          ctx,
          'ProductCategoryService',
          'RestoreProductCategory',
          request,
          RestoreProductCategoryResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
