// This is a generated file - do not edit.
//
// Generated from salesorder/v1/role.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $0;

import 'common.pb.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Role:角色定義(內建 7 角色 + 自訂角色;對齊 roles 表)。
class Role extends $pb.GeneratedMessage {
  factory Role({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.String? dataScope,
    $core.bool? isSystem,
    $core.bool? isActive,
  }) {
    final result = Role._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (dataScope != null) result.dataScope = dataScope;
    if (isSystem != null) result.isSystem = isSystem;
    if (isActive != null) result.isActive = isActive;
    return result;
  }

  Role._();

  factory Role.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Role()..mergeFromBuffer(data, registry);
  factory Role.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Role()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Role',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: Role.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'dataScope')
    ..aOB(5, _omitFieldNames ? '' : 'isSystem')
    ..aOB(6, _omitFieldNames ? '' : 'isActive')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Role clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Role copyWith(void Function(Role) updates) =>
      super.copyWith((message) => updates(message as Role)) as Role;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Role() / Role.new instead')
  static Role create() => Role._();
  static $pb.GeneratedMessage $_createMessage() => Role._();
  @$core.override
  Role createEmptyInstance() => Role._();
  @$core.pragma('dart2js:noInline')
  static Role getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Role>(Role.$_createMessage);
  static Role? _defaultInstance;

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
  $core.String get dataScope => $_getSZ(3);
  @$pb.TagNumber(4)
  set dataScope($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDataScope() => $_has(3);
  @$pb.TagNumber(4)
  void clearDataScope() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get isSystem => $_getBF(4);
  @$pb.TagNumber(5)
  set isSystem($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIsSystem() => $_has(4);
  @$pb.TagNumber(5)
  void clearIsSystem() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get isActive => $_getBF(5);
  @$pb.TagNumber(6)
  set isActive($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasIsActive() => $_has(5);
  @$pb.TagNumber(6)
  void clearIsActive() => $_clearField(6);
}

/// Permission:單一功能權限(resource × action;CASL ability 規則來源,對齊 role_permissions 表)。
class Permission extends $pb.GeneratedMessage {
  factory Permission({
    $core.String? resource,
    $core.String? action,
    $0.Struct? conditions,
    $core.bool? inverted,
    $core.int? sortOrder,
  }) {
    final result = Permission._();
    if (resource != null) result.resource = resource;
    if (action != null) result.action = action;
    if (conditions != null) result.conditions = conditions;
    if (inverted != null) result.inverted = inverted;
    if (sortOrder != null) result.sortOrder = sortOrder;
    return result;
  }

  Permission._();

  factory Permission.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Permission()..mergeFromBuffer(data, registry);
  factory Permission.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Permission()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Permission',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: Permission.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'resource')
    ..aOS(2, _omitFieldNames ? '' : 'action')
    ..aOM<$0.Struct>(3, _omitFieldNames ? '' : 'conditions',
        subBuilder: $0.Struct.$_createMessage)
    ..aOB(4, _omitFieldNames ? '' : 'inverted')
    ..aI(5, _omitFieldNames ? '' : 'sortOrder')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Permission clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Permission copyWith(void Function(Permission) updates) =>
      super.copyWith((message) => updates(message as Permission)) as Permission;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Permission() / Permission.new instead')
  static Permission create() => Permission._();
  static $pb.GeneratedMessage $_createMessage() => Permission._();
  @$core.override
  Permission createEmptyInstance() => Permission._();
  @$core.pragma('dart2js:noInline')
  static Permission getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Permission>(Permission.$_createMessage);
  static Permission? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get resource => $_getSZ(0);
  @$pb.TagNumber(1)
  set resource($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasResource() => $_has(0);
  @$pb.TagNumber(1)
  void clearResource() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get action => $_getSZ(1);
  @$pb.TagNumber(2)
  set action($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAction() => $_has(1);
  @$pb.TagNumber(2)
  void clearAction() => $_clearField(2);

  @$pb.TagNumber(3)
  $0.Struct get conditions => $_getN(2);
  @$pb.TagNumber(3)
  set conditions($0.Struct value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasConditions() => $_has(2);
  @$pb.TagNumber(3)
  void clearConditions() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Struct ensureConditions() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.bool get inverted => $_getBF(3);
  @$pb.TagNumber(4)
  set inverted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasInverted() => $_has(3);
  @$pb.TagNumber(4)
  void clearInverted() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get sortOrder => $_getIZ(4);
  @$pb.TagNumber(5)
  set sortOrder($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSortOrder() => $_has(4);
  @$pb.TagNumber(5)
  void clearSortOrder() => $_clearField(5);
}

/// ListRolesRequest:分頁列出角色。
class ListRolesRequest extends $pb.GeneratedMessage {
  factory ListRolesRequest({
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListRolesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListRolesRequest._();

  factory ListRolesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRolesRequest()..mergeFromBuffer(data, registry);
  factory ListRolesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRolesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRolesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListRolesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRolesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRolesRequest copyWith(void Function(ListRolesRequest) updates) =>
      super.copyWith((message) => updates(message as ListRolesRequest))
          as ListRolesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListRolesRequest() / ListRolesRequest.new instead')
  static ListRolesRequest create() => ListRolesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListRolesRequest._();
  @$core.override
  ListRolesRequest createEmptyInstance() => ListRolesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListRolesRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListRolesRequest>(
          ListRolesRequest.$_createMessage);
  static ListRolesRequest? _defaultInstance;

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
}

/// ListRolesResponse:角色清單。
class ListRolesResponse extends $pb.GeneratedMessage {
  factory ListRolesResponse({
    $core.Iterable<Role>? roles,
    $1.Pagination? pagination,
  }) {
    final result = ListRolesResponse._();
    if (roles != null) result.roles.addAll(roles);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListRolesResponse._();

  factory ListRolesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRolesResponse()..mergeFromBuffer(data, registry);
  factory ListRolesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRolesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRolesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListRolesResponse.$_createMessage)
    ..pPM<Role>(1, _omitFieldNames ? '' : 'roles',
        subBuilder: Role.$_createMessage)
    ..aOM<$1.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $1.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRolesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRolesResponse copyWith(void Function(ListRolesResponse) updates) =>
      super.copyWith((message) => updates(message as ListRolesResponse))
          as ListRolesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListRolesResponse() / ListRolesResponse.new instead')
  static ListRolesResponse create() => ListRolesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListRolesResponse._();
  @$core.override
  ListRolesResponse createEmptyInstance() => ListRolesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListRolesResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListRolesResponse>(
          ListRolesResponse.$_createMessage);
  static ListRolesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Role> get roles => $_getList(0);

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

/// GetRolePermissionsRequest:取單一角色的功能權限。
class GetRolePermissionsRequest extends $pb.GeneratedMessage {
  factory GetRolePermissionsRequest({
    $core.String? roleId,
  }) {
    final result = GetRolePermissionsRequest._();
    if (roleId != null) result.roleId = roleId;
    return result;
  }

  GetRolePermissionsRequest._();

  factory GetRolePermissionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRolePermissionsRequest()..mergeFromBuffer(data, registry);
  factory GetRolePermissionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRolePermissionsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRolePermissionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetRolePermissionsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'roleId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRolePermissionsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRolePermissionsRequest copyWith(
          void Function(GetRolePermissionsRequest) updates) =>
      super.copyWith((message) => updates(message as GetRolePermissionsRequest))
          as GetRolePermissionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRolePermissionsRequest() / GetRolePermissionsRequest.new instead')
  static GetRolePermissionsRequest create() => GetRolePermissionsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRolePermissionsRequest._();
  @$core.override
  GetRolePermissionsRequest createEmptyInstance() =>
      GetRolePermissionsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetRolePermissionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRolePermissionsRequest>(
          GetRolePermissionsRequest.$_createMessage);
  static GetRolePermissionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roleId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roleId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRoleId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoleId() => $_clearField(1);
}

/// GetRolePermissionsResponse:角色功能權限(依 sort_order 升冪)。
class GetRolePermissionsResponse extends $pb.GeneratedMessage {
  factory GetRolePermissionsResponse({
    $core.Iterable<Permission>? permissions,
  }) {
    final result = GetRolePermissionsResponse._();
    if (permissions != null) result.permissions.addAll(permissions);
    return result;
  }

  GetRolePermissionsResponse._();

  factory GetRolePermissionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRolePermissionsResponse()..mergeFromBuffer(data, registry);
  factory GetRolePermissionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRolePermissionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRolePermissionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetRolePermissionsResponse.$_createMessage)
    ..pPM<Permission>(1, _omitFieldNames ? '' : 'permissions',
        subBuilder: Permission.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRolePermissionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRolePermissionsResponse copyWith(
          void Function(GetRolePermissionsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetRolePermissionsResponse))
          as GetRolePermissionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRolePermissionsResponse() / GetRolePermissionsResponse.new instead')
  static GetRolePermissionsResponse create() => GetRolePermissionsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRolePermissionsResponse._();
  @$core.override
  GetRolePermissionsResponse createEmptyInstance() =>
      GetRolePermissionsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetRolePermissionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRolePermissionsResponse>(
          GetRolePermissionsResponse.$_createMessage);
  static GetRolePermissionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Permission> get permissions => $_getList(0);
}

/// UpdateRolePermissionsRequest:全量取代角色的功能權限(交易內 delete + insert)。
class UpdateRolePermissionsRequest extends $pb.GeneratedMessage {
  factory UpdateRolePermissionsRequest({
    $core.String? roleId,
    $core.Iterable<Permission>? permissions,
  }) {
    final result = UpdateRolePermissionsRequest._();
    if (roleId != null) result.roleId = roleId;
    if (permissions != null) result.permissions.addAll(permissions);
    return result;
  }

  UpdateRolePermissionsRequest._();

  factory UpdateRolePermissionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRolePermissionsRequest()..mergeFromBuffer(data, registry);
  factory UpdateRolePermissionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRolePermissionsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateRolePermissionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateRolePermissionsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'roleId')
    ..pPM<Permission>(2, _omitFieldNames ? '' : 'permissions',
        subBuilder: Permission.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRolePermissionsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRolePermissionsRequest copyWith(
          void Function(UpdateRolePermissionsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateRolePermissionsRequest))
          as UpdateRolePermissionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateRolePermissionsRequest() / UpdateRolePermissionsRequest.new instead')
  static UpdateRolePermissionsRequest create() =>
      UpdateRolePermissionsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateRolePermissionsRequest._();
  @$core.override
  UpdateRolePermissionsRequest createEmptyInstance() =>
      UpdateRolePermissionsRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateRolePermissionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateRolePermissionsRequest>(
          UpdateRolePermissionsRequest.$_createMessage);
  static UpdateRolePermissionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roleId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roleId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRoleId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoleId() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<Permission> get permissions => $_getList(1);
}

/// UpdateRolePermissionsResponse:更新後的完整功能權限(依 sort_order 升冪)。
class UpdateRolePermissionsResponse extends $pb.GeneratedMessage {
  factory UpdateRolePermissionsResponse({
    $core.Iterable<Permission>? permissions,
  }) {
    final result = UpdateRolePermissionsResponse._();
    if (permissions != null) result.permissions.addAll(permissions);
    return result;
  }

  UpdateRolePermissionsResponse._();

  factory UpdateRolePermissionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRolePermissionsResponse()..mergeFromBuffer(data, registry);
  factory UpdateRolePermissionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRolePermissionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateRolePermissionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateRolePermissionsResponse.$_createMessage)
    ..pPM<Permission>(1, _omitFieldNames ? '' : 'permissions',
        subBuilder: Permission.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRolePermissionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRolePermissionsResponse copyWith(
          void Function(UpdateRolePermissionsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateRolePermissionsResponse))
          as UpdateRolePermissionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateRolePermissionsResponse() / UpdateRolePermissionsResponse.new instead')
  static UpdateRolePermissionsResponse create() =>
      UpdateRolePermissionsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateRolePermissionsResponse._();
  @$core.override
  UpdateRolePermissionsResponse createEmptyInstance() =>
      UpdateRolePermissionsResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateRolePermissionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateRolePermissionsResponse>(
          UpdateRolePermissionsResponse.$_createMessage);
  static UpdateRolePermissionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Permission> get permissions => $_getList(0);
}

/// ListConditionFieldsRequest:查詢指定資源可用的條件欄位白名單(供前端條件建構器)。
class ListConditionFieldsRequest extends $pb.GeneratedMessage {
  factory ListConditionFieldsRequest({
    $core.String? resource,
  }) {
    final result = ListConditionFieldsRequest._();
    if (resource != null) result.resource = resource;
    return result;
  }

  ListConditionFieldsRequest._();

  factory ListConditionFieldsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListConditionFieldsRequest()..mergeFromBuffer(data, registry);
  factory ListConditionFieldsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListConditionFieldsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListConditionFieldsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListConditionFieldsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'resource')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConditionFieldsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConditionFieldsRequest copyWith(
          void Function(ListConditionFieldsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListConditionFieldsRequest))
          as ListConditionFieldsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListConditionFieldsRequest() / ListConditionFieldsRequest.new instead')
  static ListConditionFieldsRequest create() => ListConditionFieldsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListConditionFieldsRequest._();
  @$core.override
  ListConditionFieldsRequest createEmptyInstance() =>
      ListConditionFieldsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListConditionFieldsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListConditionFieldsRequest>(
          ListConditionFieldsRequest.$_createMessage);
  static ListConditionFieldsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get resource => $_getSZ(0);
  @$pb.TagNumber(1)
  set resource($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasResource() => $_has(0);
  @$pb.TagNumber(1)
  void clearResource() => $_clearField(1);
}

/// ConditionField:單一條件欄位描述(欄位名、值型別、允許運算子與 enum 合法值)。
class ConditionField extends $pb.GeneratedMessage {
  factory ConditionField({
    $core.String? field_1,
    $core.String? type,
    $core.Iterable<$core.String>? ops,
    $core.Iterable<$core.String>? enum_4,
  }) {
    final result = ConditionField._();
    if (field_1 != null) result.field_1 = field_1;
    if (type != null) result.type = type;
    if (ops != null) result.ops.addAll(ops);
    if (enum_4 != null) result.enum_4.addAll(enum_4);
    return result;
  }

  ConditionField._();

  factory ConditionField.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConditionField()..mergeFromBuffer(data, registry);
  factory ConditionField.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConditionField()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConditionField',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ConditionField.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'field')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..pPS(3, _omitFieldNames ? '' : 'ops')
    ..pPS(4, _omitFieldNames ? '' : 'enum')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConditionField clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConditionField copyWith(void Function(ConditionField) updates) =>
      super.copyWith((message) => updates(message as ConditionField))
          as ConditionField;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ConditionField() / ConditionField.new instead')
  static ConditionField create() => ConditionField._();
  static $pb.GeneratedMessage $_createMessage() => ConditionField._();
  @$core.override
  ConditionField createEmptyInstance() => ConditionField._();
  @$core.pragma('dart2js:noInline')
  static ConditionField getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConditionField>(
          ConditionField.$_createMessage);
  static ConditionField? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get field_1 => $_getSZ(0);
  @$pb.TagNumber(1)
  set field_1($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasField_1() => $_has(0);
  @$pb.TagNumber(1)
  void clearField_1() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get ops => $_getList(2);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get enum_4 => $_getList(3);
}

/// ListConditionFieldsResponse:資源的條件欄位白名單(依欄位名排序;未知資源為空)。
class ListConditionFieldsResponse extends $pb.GeneratedMessage {
  factory ListConditionFieldsResponse({
    $core.Iterable<ConditionField>? fields,
  }) {
    final result = ListConditionFieldsResponse._();
    if (fields != null) result.fields.addAll(fields);
    return result;
  }

  ListConditionFieldsResponse._();

  factory ListConditionFieldsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListConditionFieldsResponse()..mergeFromBuffer(data, registry);
  factory ListConditionFieldsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListConditionFieldsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListConditionFieldsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListConditionFieldsResponse.$_createMessage)
    ..pPM<ConditionField>(1, _omitFieldNames ? '' : 'fields',
        subBuilder: ConditionField.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConditionFieldsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListConditionFieldsResponse copyWith(
          void Function(ListConditionFieldsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListConditionFieldsResponse))
          as ListConditionFieldsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListConditionFieldsResponse() / ListConditionFieldsResponse.new instead')
  static ListConditionFieldsResponse create() =>
      ListConditionFieldsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListConditionFieldsResponse._();
  @$core.override
  ListConditionFieldsResponse createEmptyInstance() =>
      ListConditionFieldsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListConditionFieldsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListConditionFieldsResponse>(
          ListConditionFieldsResponse.$_createMessage);
  static ListConditionFieldsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ConditionField> get fields => $_getList(0);
}

/// RoleService:角色權限管理。
class RoleServiceApi {
  final $pb.RpcClient _client;

  RoleServiceApi(this._client);

  /// ListRoles:分頁列出角色。
  $async.Future<ListRolesResponse> listRoles(
          $pb.ClientContext? ctx, ListRolesRequest request) =>
      _client.invoke<ListRolesResponse>(
          ctx, 'RoleService', 'ListRoles', request, ListRolesResponse());

  /// GetRolePermissions:取得角色功能權限。
  $async.Future<GetRolePermissionsResponse> getRolePermissions(
          $pb.ClientContext? ctx, GetRolePermissionsRequest request) =>
      _client.invoke<GetRolePermissionsResponse>(ctx, 'RoleService',
          'GetRolePermissions', request, GetRolePermissionsResponse());

  /// UpdateRolePermissions:全量更新角色功能權限(取代式)。
  $async.Future<UpdateRolePermissionsResponse> updateRolePermissions(
          $pb.ClientContext? ctx, UpdateRolePermissionsRequest request) =>
      _client.invoke<UpdateRolePermissionsResponse>(ctx, 'RoleService',
          'UpdateRolePermissions', request, UpdateRolePermissionsResponse());

  /// ListConditionFields:取得資源的條件欄位白名單(條件建構器用)。
  $async.Future<ListConditionFieldsResponse> listConditionFields(
          $pb.ClientContext? ctx, ListConditionFieldsRequest request) =>
      _client.invoke<ListConditionFieldsResponse>(ctx, 'RoleService',
          'ListConditionFields', request, ListConditionFieldsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
