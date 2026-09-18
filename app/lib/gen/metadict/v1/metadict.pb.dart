// This is a generated file - do not edit.
//
// Generated from metadict/v1/metadict.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../salesorder/v1/common.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Metadict:字典值。
class Metadict extends $pb.GeneratedMessage {
  factory Metadict({
    $core.String? id,
    $core.String? type,
    $core.String? code,
    $core.String? displayName,
    $core.String? departmentId,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.String? createdAt,
    $core.String? updatedAt,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (type != null) result.type = type;
    if (code != null) result.code = code;
    if (displayName != null) result.displayName = displayName;
    if (departmentId != null) result.departmentId = departmentId;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  Metadict._();

  factory Metadict.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Metadict.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Metadict',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'code')
    ..aOS(4, _omitFieldNames ? '' : 'displayName')
    ..aOS(5, _omitFieldNames ? '' : 'departmentId')
    ..a<$core.int>(6, _omitFieldNames ? '' : 'sortOrder', $pb.PbFieldType.O3)
    ..aOB(7, _omitFieldNames ? '' : 'isActive')
    ..aOS(8, _omitFieldNames ? '' : 'createdAt')
    ..aOS(9, _omitFieldNames ? '' : 'updatedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Metadict clone() => Metadict()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Metadict copyWith(void Function(Metadict) updates) =>
      super.copyWith((message) => updates(message as Metadict)) as Metadict;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Metadict create() => Metadict._();
  @$core.override
  Metadict createEmptyInstance() => create();
  static $pb.PbList<Metadict> createRepeated() => $pb.PbList<Metadict>();
  @$core.pragma('dart2js:noInline')
  static Metadict getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Metadict>(create);
  static Metadict? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get code => $_getSZ(2);
  @$pb.TagNumber(3)
  set code($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearCode() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get displayName => $_getSZ(3);
  @$pb.TagNumber(4)
  set displayName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDisplayName() => $_has(3);
  @$pb.TagNumber(4)
  void clearDisplayName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get departmentId => $_getSZ(4);
  @$pb.TagNumber(5)
  set departmentId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDepartmentId() => $_has(4);
  @$pb.TagNumber(5)
  void clearDepartmentId() => $_clearField(5);

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
}

class ListMetadictsRequest extends $pb.GeneratedMessage {
  factory ListMetadictsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? type,
    $core.bool? includeDeleted,
    $core.String? departmentId,
  }) {
    final result = create();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (type != null) result.type = type;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    if (departmentId != null) result.departmentId = departmentId;
    return result;
  }

  ListMetadictsRequest._();

  factory ListMetadictsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListMetadictsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMetadictsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'page', $pb.PbFieldType.O3)
    ..a<$core.int>(2, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOB(5, _omitFieldNames ? '' : 'includeDeleted')
    ..aOS(6, _omitFieldNames ? '' : 'departmentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMetadictsRequest clone() =>
      ListMetadictsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMetadictsRequest copyWith(void Function(ListMetadictsRequest) updates) =>
      super.copyWith((message) => updates(message as ListMetadictsRequest))
          as ListMetadictsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListMetadictsRequest create() => ListMetadictsRequest._();
  @$core.override
  ListMetadictsRequest createEmptyInstance() => create();
  static $pb.PbList<ListMetadictsRequest> createRepeated() =>
      $pb.PbList<ListMetadictsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListMetadictsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMetadictsRequest>(create);
  static ListMetadictsRequest? _defaultInstance;

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
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  @$pb.TagNumber(5)
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(5)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(5)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(5)
  void clearIncludeDeleted() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get departmentId => $_getSZ(4);
  @$pb.TagNumber(6)
  set departmentId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(6)
  $core.bool hasDepartmentId() => $_has(4);
  @$pb.TagNumber(6)
  void clearDepartmentId() => $_clearField(6);
}

class ListMetadictsResponse extends $pb.GeneratedMessage {
  factory ListMetadictsResponse({
    $core.Iterable<Metadict>? items,
    $0.Pagination? pagination,
  }) {
    final result = create();
    if (items != null) result.items.addAll(items);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListMetadictsResponse._();

  factory ListMetadictsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListMetadictsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMetadictsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..pc<Metadict>(1, _omitFieldNames ? '' : 'items', $pb.PbFieldType.PM,
        subBuilder: Metadict.create)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMetadictsResponse clone() =>
      ListMetadictsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMetadictsResponse copyWith(
          void Function(ListMetadictsResponse) updates) =>
      super.copyWith((message) => updates(message as ListMetadictsResponse))
          as ListMetadictsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListMetadictsResponse create() => ListMetadictsResponse._();
  @$core.override
  ListMetadictsResponse createEmptyInstance() => create();
  static $pb.PbList<ListMetadictsResponse> createRepeated() =>
      $pb.PbList<ListMetadictsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListMetadictsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMetadictsResponse>(create);
  static ListMetadictsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Metadict> get items => $_getList(0);

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

class GetMetadictRequest extends $pb.GeneratedMessage {
  factory GetMetadictRequest({
    $core.String? id,
  }) {
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  GetMetadictRequest._();

  factory GetMetadictRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetMetadictRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMetadictRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetadictRequest clone() => GetMetadictRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetadictRequest copyWith(void Function(GetMetadictRequest) updates) =>
      super.copyWith((message) => updates(message as GetMetadictRequest))
          as GetMetadictRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetMetadictRequest create() => GetMetadictRequest._();
  @$core.override
  GetMetadictRequest createEmptyInstance() => create();
  static $pb.PbList<GetMetadictRequest> createRepeated() =>
      $pb.PbList<GetMetadictRequest>();
  @$core.pragma('dart2js:noInline')
  static GetMetadictRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMetadictRequest>(create);
  static GetMetadictRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetMetadictResponse extends $pb.GeneratedMessage {
  factory GetMetadictResponse({
    Metadict? metadict,
  }) {
    final result = create();
    if (metadict != null) result.metadict = metadict;
    return result;
  }

  GetMetadictResponse._();

  factory GetMetadictResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetMetadictResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMetadictResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOM<Metadict>(1, _omitFieldNames ? '' : 'metadict',
        subBuilder: Metadict.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetadictResponse clone() => GetMetadictResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetadictResponse copyWith(void Function(GetMetadictResponse) updates) =>
      super.copyWith((message) => updates(message as GetMetadictResponse))
          as GetMetadictResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetMetadictResponse create() => GetMetadictResponse._();
  @$core.override
  GetMetadictResponse createEmptyInstance() => create();
  static $pb.PbList<GetMetadictResponse> createRepeated() =>
      $pb.PbList<GetMetadictResponse>();
  @$core.pragma('dart2js:noInline')
  static GetMetadictResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMetadictResponse>(create);
  static GetMetadictResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Metadict get metadict => $_getN(0);
  @$pb.TagNumber(1)
  set metadict(Metadict value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMetadict() => $_has(0);
  @$pb.TagNumber(1)
  void clearMetadict() => $_clearField(1);
  @$pb.TagNumber(1)
  Metadict ensureMetadict() => $_ensure(0);
}

class CreateMetadictRequest extends $pb.GeneratedMessage {
  factory CreateMetadictRequest({
    $core.String? type,
    $core.String? code,
    $core.String? displayName,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (code != null) result.code = code;
    if (displayName != null) result.displayName = displayName;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  CreateMetadictRequest._();

  factory CreateMetadictRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateMetadictRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateMetadictRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'sortOrder', $pb.PbFieldType.O3)
    ..aOB(5, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMetadictRequest clone() =>
      CreateMetadictRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMetadictRequest copyWith(
          void Function(CreateMetadictRequest) updates) =>
      super.copyWith((message) => updates(message as CreateMetadictRequest))
          as CreateMetadictRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMetadictRequest create() => CreateMetadictRequest._();
  @$core.override
  CreateMetadictRequest createEmptyInstance() => create();
  static $pb.PbList<CreateMetadictRequest> createRepeated() =>
      $pb.PbList<CreateMetadictRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateMetadictRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateMetadictRequest>(create);
  static CreateMetadictRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

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

class CreateMetadictResponse extends $pb.GeneratedMessage {
  factory CreateMetadictResponse({
    Metadict? metadict,
  }) {
    final result = create();
    if (metadict != null) result.metadict = metadict;
    return result;
  }

  CreateMetadictResponse._();

  factory CreateMetadictResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateMetadictResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateMetadictResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOM<Metadict>(1, _omitFieldNames ? '' : 'metadict',
        subBuilder: Metadict.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMetadictResponse clone() =>
      CreateMetadictResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMetadictResponse copyWith(
          void Function(CreateMetadictResponse) updates) =>
      super.copyWith((message) => updates(message as CreateMetadictResponse))
          as CreateMetadictResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMetadictResponse create() => CreateMetadictResponse._();
  @$core.override
  CreateMetadictResponse createEmptyInstance() => create();
  static $pb.PbList<CreateMetadictResponse> createRepeated() =>
      $pb.PbList<CreateMetadictResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateMetadictResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateMetadictResponse>(create);
  static CreateMetadictResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Metadict get metadict => $_getN(0);
  @$pb.TagNumber(1)
  set metadict(Metadict value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMetadict() => $_has(0);
  @$pb.TagNumber(1)
  void clearMetadict() => $_clearField(1);
  @$pb.TagNumber(1)
  Metadict ensureMetadict() => $_ensure(0);
}

class UpdateMetadictRequest extends $pb.GeneratedMessage {
  factory UpdateMetadictRequest({
    $core.String? id,
    $core.String? displayName,
    $core.int? sortOrder,
    $core.bool? isActive,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (displayName != null) result.displayName = displayName;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  UpdateMetadictRequest._();

  factory UpdateMetadictRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateMetadictRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateMetadictRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'sortOrder', $pb.PbFieldType.O3)
    ..aOB(4, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMetadictRequest clone() =>
      UpdateMetadictRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMetadictRequest copyWith(
          void Function(UpdateMetadictRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateMetadictRequest))
          as UpdateMetadictRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateMetadictRequest create() => UpdateMetadictRequest._();
  @$core.override
  UpdateMetadictRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateMetadictRequest> createRepeated() =>
      $pb.PbList<UpdateMetadictRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateMetadictRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateMetadictRequest>(create);
  static UpdateMetadictRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

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

class UpdateMetadictResponse extends $pb.GeneratedMessage {
  factory UpdateMetadictResponse({
    Metadict? metadict,
  }) {
    final result = create();
    if (metadict != null) result.metadict = metadict;
    return result;
  }

  UpdateMetadictResponse._();

  factory UpdateMetadictResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateMetadictResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateMetadictResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOM<Metadict>(1, _omitFieldNames ? '' : 'metadict',
        subBuilder: Metadict.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMetadictResponse clone() =>
      UpdateMetadictResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMetadictResponse copyWith(
          void Function(UpdateMetadictResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateMetadictResponse))
          as UpdateMetadictResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateMetadictResponse create() => UpdateMetadictResponse._();
  @$core.override
  UpdateMetadictResponse createEmptyInstance() => create();
  static $pb.PbList<UpdateMetadictResponse> createRepeated() =>
      $pb.PbList<UpdateMetadictResponse>();
  @$core.pragma('dart2js:noInline')
  static UpdateMetadictResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateMetadictResponse>(create);
  static UpdateMetadictResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Metadict get metadict => $_getN(0);
  @$pb.TagNumber(1)
  set metadict(Metadict value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMetadict() => $_has(0);
  @$pb.TagNumber(1)
  void clearMetadict() => $_clearField(1);
  @$pb.TagNumber(1)
  Metadict ensureMetadict() => $_ensure(0);
}

class DeleteMetadictRequest extends $pb.GeneratedMessage {
  factory DeleteMetadictRequest({
    $core.String? id,
  }) {
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  DeleteMetadictRequest._();

  factory DeleteMetadictRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteMetadictRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteMetadictRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMetadictRequest clone() =>
      DeleteMetadictRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMetadictRequest copyWith(
          void Function(DeleteMetadictRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteMetadictRequest))
          as DeleteMetadictRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteMetadictRequest create() => DeleteMetadictRequest._();
  @$core.override
  DeleteMetadictRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteMetadictRequest> createRepeated() =>
      $pb.PbList<DeleteMetadictRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteMetadictRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteMetadictRequest>(create);
  static DeleteMetadictRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteMetadictResponse extends $pb.GeneratedMessage {
  factory DeleteMetadictResponse() => create();

  DeleteMetadictResponse._();

  factory DeleteMetadictResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteMetadictResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteMetadictResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMetadictResponse clone() =>
      DeleteMetadictResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMetadictResponse copyWith(
          void Function(DeleteMetadictResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteMetadictResponse))
          as DeleteMetadictResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteMetadictResponse create() => DeleteMetadictResponse._();
  @$core.override
  DeleteMetadictResponse createEmptyInstance() => create();
  static $pb.PbList<DeleteMetadictResponse> createRepeated() =>
      $pb.PbList<DeleteMetadictResponse>();
  @$core.pragma('dart2js:noInline')
  static DeleteMetadictResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteMetadictResponse>(create);
  static DeleteMetadictResponse? _defaultInstance;
}

class ListOptionsRequest extends $pb.GeneratedMessage {
  factory ListOptionsRequest({
    $core.String? type,
    $core.String? keyword,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (keyword != null) result.keyword = keyword;
    return result;
  }

  ListOptionsRequest._();

  factory ListOptionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListOptionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOptionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'keyword')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOptionsRequest clone() => ListOptionsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOptionsRequest copyWith(void Function(ListOptionsRequest) updates) =>
      super.copyWith((message) => updates(message as ListOptionsRequest))
          as ListOptionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListOptionsRequest create() => ListOptionsRequest._();
  @$core.override
  ListOptionsRequest createEmptyInstance() => create();
  static $pb.PbList<ListOptionsRequest> createRepeated() =>
      $pb.PbList<ListOptionsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListOptionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOptionsRequest>(create);
  static ListOptionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get keyword => $_getSZ(1);
  @$pb.TagNumber(2)
  set keyword($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKeyword() => $_has(1);
  @$pb.TagNumber(2)
  void clearKeyword() => $_clearField(2);
}

class ListOptionsResponse extends $pb.GeneratedMessage {
  factory ListOptionsResponse({
    $core.Iterable<Option>? options,
  }) {
    final result = create();
    if (options != null) result.options.addAll(options);
    return result;
  }

  ListOptionsResponse._();

  factory ListOptionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListOptionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOptionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..pc<Option>(1, _omitFieldNames ? '' : 'options', $pb.PbFieldType.PM,
        subBuilder: Option.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOptionsResponse clone() => ListOptionsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOptionsResponse copyWith(void Function(ListOptionsResponse) updates) =>
      super.copyWith((message) => updates(message as ListOptionsResponse))
          as ListOptionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListOptionsResponse create() => ListOptionsResponse._();
  @$core.override
  ListOptionsResponse createEmptyInstance() => create();
  static $pb.PbList<ListOptionsResponse> createRepeated() =>
      $pb.PbList<ListOptionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListOptionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOptionsResponse>(create);
  static ListOptionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Option> get options => $_getList(0);
}

/// Option:表單下拉選項(精簡)。
class Option extends $pb.GeneratedMessage {
  factory Option({
    $core.String? code,
    $core.String? displayName,
  }) {
    final result = create();
    if (code != null) result.code = code;
    if (displayName != null) result.displayName = displayName;
    return result;
  }

  Option._();

  factory Option.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Option.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Option',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'metadict.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Option clone() => Option()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Option copyWith(void Function(Option) updates) =>
      super.copyWith((message) => updates(message as Option)) as Option;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Option create() => Option._();
  @$core.override
  Option createEmptyInstance() => create();
  static $pb.PbList<Option> createRepeated() => $pb.PbList<Option>();
  @$core.pragma('dart2js:noInline')
  static Option getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Option>(create);
  static Option? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get code => $_getSZ(0);
  @$pb.TagNumber(1)
  set code($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);
}

/// MetadictService:字典管理。
class MetadictServiceApi {
  final $pb.RpcClient _client;

  MetadictServiceApi(this._client);

  /// ListMetadicts:分頁查詢(系統預設 + 當前部門擴充合併),可依 type 篩選。
  $async.Future<ListMetadictsResponse> listMetadicts(
          $pb.ClientContext? ctx, ListMetadictsRequest request) =>
      _client.invoke<ListMetadictsResponse>(ctx, 'MetadictService',
          'ListMetadicts', request, ListMetadictsResponse());

  /// GetMetadict:取得單一字典。
  $async.Future<GetMetadictResponse> getMetadict(
          $pb.ClientContext? ctx, GetMetadictRequest request) =>
      _client.invoke<GetMetadictResponse>(ctx, 'MetadictService', 'GetMetadict',
          request, GetMetadictResponse());

  /// CreateMetadict:建立字典(super 建系統級;dept_admin/staff 自動帶當前部門,不接受請求帶 department_id)。
  $async.Future<CreateMetadictResponse> createMetadict(
          $pb.ClientContext? ctx, CreateMetadictRequest request) =>
      _client.invoke<CreateMetadictResponse>(ctx, 'MetadictService',
          'CreateMetadict', request, CreateMetadictResponse());

  /// UpdateMetadict:更新 display_name / sort_order / is_active(不含 type / code)。
  $async.Future<UpdateMetadictResponse> updateMetadict(
          $pb.ClientContext? ctx, UpdateMetadictRequest request) =>
      _client.invoke<UpdateMetadictResponse>(ctx, 'MetadictService',
          'UpdateMetadict', request, UpdateMetadictResponse());

  /// DeleteMetadict:軟刪除(order_source 不可刪)。
  $async.Future<DeleteMetadictResponse> deleteMetadict(
          $pb.ClientContext? ctx, DeleteMetadictRequest request) =>
      _client.invoke<DeleteMetadictResponse>(ctx, 'MetadictService',
          'DeleteMetadict', request, DeleteMetadictResponse());

  /// ListOptions:表單下拉選項(僅可選用啟用值;客戶端身分僅回系統預設)。
  $async.Future<ListOptionsResponse> listOptions(
          $pb.ClientContext? ctx, ListOptionsRequest request) =>
      _client.invoke<ListOptionsResponse>(ctx, 'MetadictService', 'ListOptions',
          request, ListOptionsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
