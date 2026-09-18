// This is a generated file - do not edit.
//
// Generated from audit/v1/audit.proto.

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

/// AuditLog:一筆稽核紀錄。
class AuditLog extends $pb.GeneratedMessage {
  factory AuditLog({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? userId,
    $core.String? userName,
    $core.String? action,
    $core.String? resourceType,
    $core.String? resourceId,
    $core.String? beforeSnapshot,
    $core.String? afterSnapshot,
    $core.String? ipAddress,
    $core.String? userAgent,
    $core.String? createdAt,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (userId != null) result.userId = userId;
    if (userName != null) result.userName = userName;
    if (action != null) result.action = action;
    if (resourceType != null) result.resourceType = resourceType;
    if (resourceId != null) result.resourceId = resourceId;
    if (beforeSnapshot != null) result.beforeSnapshot = beforeSnapshot;
    if (afterSnapshot != null) result.afterSnapshot = afterSnapshot;
    if (ipAddress != null) result.ipAddress = ipAddress;
    if (userAgent != null) result.userAgent = userAgent;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  AuditLog._();

  factory AuditLog.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AuditLog.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AuditLog',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'audit.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'userId')
    ..aOS(5, _omitFieldNames ? '' : 'userName')
    ..aOS(6, _omitFieldNames ? '' : 'action')
    ..aOS(7, _omitFieldNames ? '' : 'resourceType')
    ..aOS(8, _omitFieldNames ? '' : 'resourceId')
    ..aOS(9, _omitFieldNames ? '' : 'beforeSnapshot')
    ..aOS(10, _omitFieldNames ? '' : 'afterSnapshot')
    ..aOS(11, _omitFieldNames ? '' : 'ipAddress')
    ..aOS(12, _omitFieldNames ? '' : 'userAgent')
    ..aOS(13, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuditLog clone() => AuditLog()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AuditLog copyWith(void Function(AuditLog) updates) =>
      super.copyWith((message) => updates(message as AuditLog)) as AuditLog;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AuditLog create() => AuditLog._();
  @$core.override
  AuditLog createEmptyInstance() => create();
  static $pb.PbList<AuditLog> createRepeated() => $pb.PbList<AuditLog>();
  @$core.pragma('dart2js:noInline')
  static AuditLog getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AuditLog>(create);
  static AuditLog? _defaultInstance;

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
  $core.String get userId => $_getSZ(3);
  @$pb.TagNumber(4)
  set userId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasUserId() => $_has(3);
  @$pb.TagNumber(4)
  void clearUserId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get userName => $_getSZ(4);
  @$pb.TagNumber(5)
  set userName($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUserName() => $_has(4);
  @$pb.TagNumber(5)
  void clearUserName() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get action => $_getSZ(5);
  @$pb.TagNumber(6)
  set action($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAction() => $_has(5);
  @$pb.TagNumber(6)
  void clearAction() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get resourceType => $_getSZ(6);
  @$pb.TagNumber(7)
  set resourceType($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasResourceType() => $_has(6);
  @$pb.TagNumber(7)
  void clearResourceType() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get resourceId => $_getSZ(7);
  @$pb.TagNumber(8)
  set resourceId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasResourceId() => $_has(7);
  @$pb.TagNumber(8)
  void clearResourceId() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get beforeSnapshot => $_getSZ(8);
  @$pb.TagNumber(9)
  set beforeSnapshot($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasBeforeSnapshot() => $_has(8);
  @$pb.TagNumber(9)
  void clearBeforeSnapshot() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get afterSnapshot => $_getSZ(9);
  @$pb.TagNumber(10)
  set afterSnapshot($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasAfterSnapshot() => $_has(9);
  @$pb.TagNumber(10)
  void clearAfterSnapshot() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get ipAddress => $_getSZ(10);
  @$pb.TagNumber(11)
  set ipAddress($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasIpAddress() => $_has(10);
  @$pb.TagNumber(11)
  void clearIpAddress() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get userAgent => $_getSZ(11);
  @$pb.TagNumber(12)
  set userAgent($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasUserAgent() => $_has(11);
  @$pb.TagNumber(12)
  void clearUserAgent() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get createdAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set createdAt($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasCreatedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearCreatedAt() => $_clearField(13);
}

class ListAuditLogsRequest extends $pb.GeneratedMessage {
  factory ListAuditLogsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? from,
    $core.String? to,
    $core.String? companyId,
    $core.String? action,
    $core.String? resourceType,
    $core.String? resourceId,
    $core.String? userId,
  }) {
    final result = create();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (from != null) result.from = from;
    if (to != null) result.to = to;
    if (companyId != null) result.companyId = companyId;
    if (action != null) result.action = action;
    if (resourceType != null) result.resourceType = resourceType;
    if (resourceId != null) result.resourceId = resourceId;
    if (userId != null) result.userId = userId;
    return result;
  }

  ListAuditLogsRequest._();

  factory ListAuditLogsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListAuditLogsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAuditLogsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'audit.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'page', $pb.PbFieldType.O3)
    ..a<$core.int>(2, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'from')
    ..aOS(4, _omitFieldNames ? '' : 'to')
    ..aOS(5, _omitFieldNames ? '' : 'companyId')
    ..aOS(6, _omitFieldNames ? '' : 'action')
    ..aOS(7, _omitFieldNames ? '' : 'resourceType')
    ..aOS(8, _omitFieldNames ? '' : 'resourceId')
    ..aOS(9, _omitFieldNames ? '' : 'userId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAuditLogsRequest clone() =>
      ListAuditLogsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAuditLogsRequest copyWith(void Function(ListAuditLogsRequest) updates) =>
      super.copyWith((message) => updates(message as ListAuditLogsRequest))
          as ListAuditLogsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListAuditLogsRequest create() => ListAuditLogsRequest._();
  @$core.override
  ListAuditLogsRequest createEmptyInstance() => create();
  static $pb.PbList<ListAuditLogsRequest> createRepeated() =>
      $pb.PbList<ListAuditLogsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListAuditLogsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAuditLogsRequest>(create);
  static ListAuditLogsRequest? _defaultInstance;

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
  $core.String get from => $_getSZ(2);
  @$pb.TagNumber(3)
  set from($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFrom() => $_has(2);
  @$pb.TagNumber(3)
  void clearFrom() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get to => $_getSZ(3);
  @$pb.TagNumber(4)
  set to($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTo() => $_has(3);
  @$pb.TagNumber(4)
  void clearTo() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get companyId => $_getSZ(4);
  @$pb.TagNumber(5)
  set companyId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCompanyId() => $_has(4);
  @$pb.TagNumber(5)
  void clearCompanyId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get action => $_getSZ(5);
  @$pb.TagNumber(6)
  set action($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAction() => $_has(5);
  @$pb.TagNumber(6)
  void clearAction() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get resourceType => $_getSZ(6);
  @$pb.TagNumber(7)
  set resourceType($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasResourceType() => $_has(6);
  @$pb.TagNumber(7)
  void clearResourceType() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get resourceId => $_getSZ(7);
  @$pb.TagNumber(8)
  set resourceId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasResourceId() => $_has(7);
  @$pb.TagNumber(8)
  void clearResourceId() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get userId => $_getSZ(8);
  @$pb.TagNumber(9)
  set userId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUserId() => $_has(8);
  @$pb.TagNumber(9)
  void clearUserId() => $_clearField(9);
}

class ListAuditLogsResponse extends $pb.GeneratedMessage {
  factory ListAuditLogsResponse({
    $core.Iterable<AuditLog>? items,
    $0.Pagination? pagination,
  }) {
    final result = create();
    if (items != null) result.items.addAll(items);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListAuditLogsResponse._();

  factory ListAuditLogsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListAuditLogsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAuditLogsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'audit.v1'),
      createEmptyInstance: create)
    ..pc<AuditLog>(1, _omitFieldNames ? '' : 'items', $pb.PbFieldType.PM,
        subBuilder: AuditLog.create)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAuditLogsResponse clone() =>
      ListAuditLogsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAuditLogsResponse copyWith(
          void Function(ListAuditLogsResponse) updates) =>
      super.copyWith((message) => updates(message as ListAuditLogsResponse))
          as ListAuditLogsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListAuditLogsResponse create() => ListAuditLogsResponse._();
  @$core.override
  ListAuditLogsResponse createEmptyInstance() => create();
  static $pb.PbList<ListAuditLogsResponse> createRepeated() =>
      $pb.PbList<ListAuditLogsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListAuditLogsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAuditLogsResponse>(create);
  static ListAuditLogsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AuditLog> get items => $_getList(0);

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

/// AuditService:稽核查詢。
class AuditServiceApi {
  final $pb.RpcClient _client;

  AuditServiceApi(this._client);

  /// ListAuditLogs:分頁查詢,可依時間(預設近 3 個月)/action/resource/user 篩選;時間降冪。
  $async.Future<ListAuditLogsResponse> listAuditLogs(
          $pb.ClientContext? ctx, ListAuditLogsRequest request) =>
      _client.invoke<ListAuditLogsResponse>(ctx, 'AuditService',
          'ListAuditLogs', request, ListAuditLogsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
