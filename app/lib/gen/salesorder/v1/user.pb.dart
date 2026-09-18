// This is a generated file - do not edit.
//
// Generated from salesorder/v1/user.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'common.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// User:使用者帳號。
class User extends $pb.GeneratedMessage {
  factory User({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? email,
    $core.String? name,
    $core.String? phone,
    $core.String? employeeNo,
    $core.String? status,
    $core.String? role,
    $core.bool? isCustomer,
    $core.String? accountName,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (email != null) result.email = email;
    if (name != null) result.name = name;
    if (phone != null) result.phone = phone;
    if (employeeNo != null) result.employeeNo = employeeNo;
    if (status != null) result.status = status;
    if (role != null) result.role = role;
    if (isCustomer != null) result.isCustomer = isCustomer;
    if (accountName != null) result.accountName = accountName;
    return result;
  }

  User._();

  factory User.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory User.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'User',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'email')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'phone')
    ..aOS(7, _omitFieldNames ? '' : 'employeeNo')
    ..aOS(8, _omitFieldNames ? '' : 'status')
    ..aOS(9, _omitFieldNames ? '' : 'role')
    ..aOB(10, _omitFieldNames ? '' : 'isCustomer')
    ..aOS(11, _omitFieldNames ? '' : 'accountName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User clone() => User()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User copyWith(void Function(User) updates) =>
      super.copyWith((message) => updates(message as User)) as User;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static User create() => User._();
  @$core.override
  User createEmptyInstance() => create();
  static $pb.PbList<User> createRepeated() => $pb.PbList<User>();
  @$core.pragma('dart2js:noInline')
  static User getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<User>(create);
  static User? _defaultInstance;

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
  $core.String get email => $_getSZ(3);
  @$pb.TagNumber(4)
  set email($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEmail() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmail() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get name => $_getSZ(4);
  @$pb.TagNumber(5)
  set name($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasName() => $_has(4);
  @$pb.TagNumber(5)
  void clearName() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get phone => $_getSZ(5);
  @$pb.TagNumber(6)
  set phone($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPhone() => $_has(5);
  @$pb.TagNumber(6)
  void clearPhone() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get employeeNo => $_getSZ(6);
  @$pb.TagNumber(7)
  set employeeNo($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasEmployeeNo() => $_has(6);
  @$pb.TagNumber(7)
  void clearEmployeeNo() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get status => $_getSZ(7);
  @$pb.TagNumber(8)
  set status($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearStatus() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get role => $_getSZ(8);
  @$pb.TagNumber(9)
  set role($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasRole() => $_has(8);
  @$pb.TagNumber(9)
  void clearRole() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.bool get isCustomer => $_getBF(9);
  @$pb.TagNumber(10)
  set isCustomer($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasIsCustomer() => $_has(9);
  @$pb.TagNumber(10)
  void clearIsCustomer() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get accountName => $_getSZ(10);
  @$pb.TagNumber(11)
  set accountName($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasAccountName() => $_has(10);
  @$pb.TagNumber(11)
  void clearAccountName() => $_clearField(11);
}

class ListUsersRequest extends $pb.GeneratedMessage {
  factory ListUsersRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? role,
    $core.String? status,
  }) {
    final result = create();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (role != null) result.role = role;
    if (status != null) result.status = status;
    return result;
  }

  ListUsersRequest._();

  factory ListUsersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListUsersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListUsersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'page', $pb.PbFieldType.O3)
    ..a<$core.int>(2, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'companyId')
    ..aOS(4, _omitFieldNames ? '' : 'departmentId')
    ..aOS(5, _omitFieldNames ? '' : 'role')
    ..aOS(6, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersRequest clone() => ListUsersRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersRequest copyWith(void Function(ListUsersRequest) updates) =>
      super.copyWith((message) => updates(message as ListUsersRequest))
          as ListUsersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListUsersRequest create() => ListUsersRequest._();
  @$core.override
  ListUsersRequest createEmptyInstance() => create();
  static $pb.PbList<ListUsersRequest> createRepeated() =>
      $pb.PbList<ListUsersRequest>();
  @$core.pragma('dart2js:noInline')
  static ListUsersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListUsersRequest>(create);
  static ListUsersRequest? _defaultInstance;

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
  $core.String get companyId => $_getSZ(2);
  @$pb.TagNumber(3)
  set companyId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCompanyId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCompanyId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get departmentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set departmentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDepartmentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearDepartmentId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get role => $_getSZ(4);
  @$pb.TagNumber(5)
  set role($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRole() => $_has(4);
  @$pb.TagNumber(5)
  void clearRole() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get status => $_getSZ(5);
  @$pb.TagNumber(6)
  set status($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);
}

class ListUsersResponse extends $pb.GeneratedMessage {
  factory ListUsersResponse({
    $core.Iterable<User>? users,
    $0.Pagination? pagination,
  }) {
    final result = create();
    if (users != null) result.users.addAll(users);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListUsersResponse._();

  factory ListUsersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListUsersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListUsersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..pc<User>(1, _omitFieldNames ? '' : 'users', $pb.PbFieldType.PM,
        subBuilder: User.create)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersResponse clone() => ListUsersResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersResponse copyWith(void Function(ListUsersResponse) updates) =>
      super.copyWith((message) => updates(message as ListUsersResponse))
          as ListUsersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListUsersResponse create() => ListUsersResponse._();
  @$core.override
  ListUsersResponse createEmptyInstance() => create();
  static $pb.PbList<ListUsersResponse> createRepeated() =>
      $pb.PbList<ListUsersResponse>();
  @$core.pragma('dart2js:noInline')
  static ListUsersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListUsersResponse>(create);
  static ListUsersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<User> get users => $_getList(0);

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

class GetUserRequest extends $pb.GeneratedMessage {
  factory GetUserRequest({
    $core.String? userId,
  }) {
    final result = create();
    if (userId != null) result.userId = userId;
    return result;
  }

  GetUserRequest._();

  factory GetUserRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetUserRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetUserRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetUserRequest clone() => GetUserRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetUserRequest copyWith(void Function(GetUserRequest) updates) =>
      super.copyWith((message) => updates(message as GetUserRequest))
          as GetUserRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetUserRequest create() => GetUserRequest._();
  @$core.override
  GetUserRequest createEmptyInstance() => create();
  static $pb.PbList<GetUserRequest> createRepeated() =>
      $pb.PbList<GetUserRequest>();
  @$core.pragma('dart2js:noInline')
  static GetUserRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetUserRequest>(create);
  static GetUserRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);
}

class GetUserResponse extends $pb.GeneratedMessage {
  factory GetUserResponse({
    User? user,
  }) {
    final result = create();
    if (user != null) result.user = user;
    return result;
  }

  GetUserResponse._();

  factory GetUserResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetUserResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetUserResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user', subBuilder: User.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetUserResponse clone() => GetUserResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetUserResponse copyWith(void Function(GetUserResponse) updates) =>
      super.copyWith((message) => updates(message as GetUserResponse))
          as GetUserResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetUserResponse create() => GetUserResponse._();
  @$core.override
  GetUserResponse createEmptyInstance() => create();
  static $pb.PbList<GetUserResponse> createRepeated() =>
      $pb.PbList<GetUserResponse>();
  @$core.pragma('dart2js:noInline')
  static GetUserResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetUserResponse>(create);
  static GetUserResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class CreateUserRequest extends $pb.GeneratedMessage {
  factory CreateUserRequest({
    $core.String? name,
    $core.String? email,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? role,
    $core.String? phone,
    $core.String? employeeNo,
  }) {
    final result = create();
    if (name != null) result.name = name;
    if (email != null) result.email = email;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (role != null) result.role = role;
    if (phone != null) result.phone = phone;
    if (employeeNo != null) result.employeeNo = employeeNo;
    return result;
  }

  CreateUserRequest._();

  factory CreateUserRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateUserRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateUserRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'email')
    ..aOS(3, _omitFieldNames ? '' : 'companyId')
    ..aOS(4, _omitFieldNames ? '' : 'departmentId')
    ..aOS(5, _omitFieldNames ? '' : 'role')
    ..aOS(6, _omitFieldNames ? '' : 'phone')
    ..aOS(7, _omitFieldNames ? '' : 'employeeNo')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserRequest clone() => CreateUserRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserRequest copyWith(void Function(CreateUserRequest) updates) =>
      super.copyWith((message) => updates(message as CreateUserRequest))
          as CreateUserRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateUserRequest create() => CreateUserRequest._();
  @$core.override
  CreateUserRequest createEmptyInstance() => create();
  static $pb.PbList<CreateUserRequest> createRepeated() =>
      $pb.PbList<CreateUserRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateUserRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateUserRequest>(create);
  static CreateUserRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get email => $_getSZ(1);
  @$pb.TagNumber(2)
  set email($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEmail() => $_has(1);
  @$pb.TagNumber(2)
  void clearEmail() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get companyId => $_getSZ(2);
  @$pb.TagNumber(3)
  set companyId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCompanyId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCompanyId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get departmentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set departmentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDepartmentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearDepartmentId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get role => $_getSZ(4);
  @$pb.TagNumber(5)
  set role($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRole() => $_has(4);
  @$pb.TagNumber(5)
  void clearRole() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get phone => $_getSZ(5);
  @$pb.TagNumber(6)
  set phone($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPhone() => $_has(5);
  @$pb.TagNumber(6)
  void clearPhone() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get employeeNo => $_getSZ(6);
  @$pb.TagNumber(7)
  set employeeNo($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasEmployeeNo() => $_has(6);
  @$pb.TagNumber(7)
  void clearEmployeeNo() => $_clearField(7);
}

class CreateUserResponse extends $pb.GeneratedMessage {
  factory CreateUserResponse({
    User? user,
  }) {
    final result = create();
    if (user != null) result.user = user;
    return result;
  }

  CreateUserResponse._();

  factory CreateUserResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateUserResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateUserResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user', subBuilder: User.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserResponse clone() => CreateUserResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserResponse copyWith(void Function(CreateUserResponse) updates) =>
      super.copyWith((message) => updates(message as CreateUserResponse))
          as CreateUserResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateUserResponse create() => CreateUserResponse._();
  @$core.override
  CreateUserResponse createEmptyInstance() => create();
  static $pb.PbList<CreateUserResponse> createRepeated() =>
      $pb.PbList<CreateUserResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateUserResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateUserResponse>(create);
  static CreateUserResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class UpdateUserRequest extends $pb.GeneratedMessage {
  factory UpdateUserRequest({
    $core.String? userId,
    $core.String? name,
    $core.String? departmentId,
    $core.String? phone,
    $core.String? employeeNo,
    $core.String? status,
  }) {
    final result = create();
    if (userId != null) result.userId = userId;
    if (name != null) result.name = name;
    if (departmentId != null) result.departmentId = departmentId;
    if (phone != null) result.phone = phone;
    if (employeeNo != null) result.employeeNo = employeeNo;
    if (status != null) result.status = status;
    return result;
  }

  UpdateUserRequest._();

  factory UpdateUserRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateUserRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateUserRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'phone')
    ..aOS(5, _omitFieldNames ? '' : 'employeeNo')
    ..aOS(6, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserRequest clone() => UpdateUserRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserRequest copyWith(void Function(UpdateUserRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateUserRequest))
          as UpdateUserRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateUserRequest create() => UpdateUserRequest._();
  @$core.override
  UpdateUserRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateUserRequest> createRepeated() =>
      $pb.PbList<UpdateUserRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateUserRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateUserRequest>(create);
  static UpdateUserRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get departmentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set departmentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDepartmentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearDepartmentId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get phone => $_getSZ(3);
  @$pb.TagNumber(4)
  set phone($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPhone() => $_has(3);
  @$pb.TagNumber(4)
  void clearPhone() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get employeeNo => $_getSZ(4);
  @$pb.TagNumber(5)
  set employeeNo($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasEmployeeNo() => $_has(4);
  @$pb.TagNumber(5)
  void clearEmployeeNo() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get status => $_getSZ(5);
  @$pb.TagNumber(6)
  set status($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);
}

class UpdateUserResponse extends $pb.GeneratedMessage {
  factory UpdateUserResponse({
    User? user,
  }) {
    final result = create();
    if (user != null) result.user = user;
    return result;
  }

  UpdateUserResponse._();

  factory UpdateUserResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateUserResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateUserResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user', subBuilder: User.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserResponse clone() => UpdateUserResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserResponse copyWith(void Function(UpdateUserResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateUserResponse))
          as UpdateUserResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateUserResponse create() => UpdateUserResponse._();
  @$core.override
  UpdateUserResponse createEmptyInstance() => create();
  static $pb.PbList<UpdateUserResponse> createRepeated() =>
      $pb.PbList<UpdateUserResponse>();
  @$core.pragma('dart2js:noInline')
  static UpdateUserResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateUserResponse>(create);
  static UpdateUserResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class AssignRoleRequest extends $pb.GeneratedMessage {
  factory AssignRoleRequest({
    $core.String? userId,
    $core.String? role,
    $core.String? departmentId,
  }) {
    final result = create();
    if (userId != null) result.userId = userId;
    if (role != null) result.role = role;
    if (departmentId != null) result.departmentId = departmentId;
    return result;
  }

  AssignRoleRequest._();

  factory AssignRoleRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AssignRoleRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignRoleRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..aOS(2, _omitFieldNames ? '' : 'role')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRoleRequest clone() => AssignRoleRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRoleRequest copyWith(void Function(AssignRoleRequest) updates) =>
      super.copyWith((message) => updates(message as AssignRoleRequest))
          as AssignRoleRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AssignRoleRequest create() => AssignRoleRequest._();
  @$core.override
  AssignRoleRequest createEmptyInstance() => create();
  static $pb.PbList<AssignRoleRequest> createRepeated() =>
      $pb.PbList<AssignRoleRequest>();
  @$core.pragma('dart2js:noInline')
  static AssignRoleRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignRoleRequest>(create);
  static AssignRoleRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get role => $_getSZ(1);
  @$pb.TagNumber(2)
  set role($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRole() => $_has(1);
  @$pb.TagNumber(2)
  void clearRole() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get departmentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set departmentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDepartmentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearDepartmentId() => $_clearField(3);
}

class AssignRoleResponse extends $pb.GeneratedMessage {
  factory AssignRoleResponse({
    User? user,
  }) {
    final result = create();
    if (user != null) result.user = user;
    return result;
  }

  AssignRoleResponse._();

  factory AssignRoleResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AssignRoleResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignRoleResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user', subBuilder: User.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRoleResponse clone() => AssignRoleResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRoleResponse copyWith(void Function(AssignRoleResponse) updates) =>
      super.copyWith((message) => updates(message as AssignRoleResponse))
          as AssignRoleResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AssignRoleResponse create() => AssignRoleResponse._();
  @$core.override
  AssignRoleResponse createEmptyInstance() => create();
  static $pb.PbList<AssignRoleResponse> createRepeated() =>
      $pb.PbList<AssignRoleResponse>();
  @$core.pragma('dart2js:noInline')
  static AssignRoleResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignRoleResponse>(create);
  static AssignRoleResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class DeactivateRequest extends $pb.GeneratedMessage {
  factory DeactivateRequest({
    $core.String? userId,
  }) {
    final result = create();
    if (userId != null) result.userId = userId;
    return result;
  }

  DeactivateRequest._();

  factory DeactivateRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeactivateRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeactivateRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateRequest clone() => DeactivateRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateRequest copyWith(void Function(DeactivateRequest) updates) =>
      super.copyWith((message) => updates(message as DeactivateRequest))
          as DeactivateRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeactivateRequest create() => DeactivateRequest._();
  @$core.override
  DeactivateRequest createEmptyInstance() => create();
  static $pb.PbList<DeactivateRequest> createRepeated() =>
      $pb.PbList<DeactivateRequest>();
  @$core.pragma('dart2js:noInline')
  static DeactivateRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeactivateRequest>(create);
  static DeactivateRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);
}

class DeactivateResponse extends $pb.GeneratedMessage {
  factory DeactivateResponse() => create();

  DeactivateResponse._();

  factory DeactivateResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeactivateResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeactivateResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateResponse clone() => DeactivateResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateResponse copyWith(void Function(DeactivateResponse) updates) =>
      super.copyWith((message) => updates(message as DeactivateResponse))
          as DeactivateResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeactivateResponse create() => DeactivateResponse._();
  @$core.override
  DeactivateResponse createEmptyInstance() => create();
  static $pb.PbList<DeactivateResponse> createRepeated() =>
      $pb.PbList<DeactivateResponse>();
  @$core.pragma('dart2js:noInline')
  static DeactivateResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeactivateResponse>(create);
  static DeactivateResponse? _defaultInstance;
}

class ForceLogoutRequest extends $pb.GeneratedMessage {
  factory ForceLogoutRequest({
    $core.String? userId,
  }) {
    final result = create();
    if (userId != null) result.userId = userId;
    return result;
  }

  ForceLogoutRequest._();

  factory ForceLogoutRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ForceLogoutRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ForceLogoutRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForceLogoutRequest clone() => ForceLogoutRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForceLogoutRequest copyWith(void Function(ForceLogoutRequest) updates) =>
      super.copyWith((message) => updates(message as ForceLogoutRequest))
          as ForceLogoutRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ForceLogoutRequest create() => ForceLogoutRequest._();
  @$core.override
  ForceLogoutRequest createEmptyInstance() => create();
  static $pb.PbList<ForceLogoutRequest> createRepeated() =>
      $pb.PbList<ForceLogoutRequest>();
  @$core.pragma('dart2js:noInline')
  static ForceLogoutRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ForceLogoutRequest>(create);
  static ForceLogoutRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);
}

class ForceLogoutResponse extends $pb.GeneratedMessage {
  factory ForceLogoutResponse() => create();

  ForceLogoutResponse._();

  factory ForceLogoutResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ForceLogoutResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ForceLogoutResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForceLogoutResponse clone() => ForceLogoutResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ForceLogoutResponse copyWith(void Function(ForceLogoutResponse) updates) =>
      super.copyWith((message) => updates(message as ForceLogoutResponse))
          as ForceLogoutResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ForceLogoutResponse create() => ForceLogoutResponse._();
  @$core.override
  ForceLogoutResponse createEmptyInstance() => create();
  static $pb.PbList<ForceLogoutResponse> createRepeated() =>
      $pb.PbList<ForceLogoutResponse>();
  @$core.pragma('dart2js:noInline')
  static ForceLogoutResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ForceLogoutResponse>(create);
  static ForceLogoutResponse? _defaultInstance;
}

/// UserService:使用者管理。
class UserServiceApi {
  final $pb.RpcClient _client;

  UserServiceApi(this._client);

  /// ListUsers:分頁列出使用者,可依 company_id / department_id / role / status 篩選。
  $async.Future<ListUsersResponse> listUsers(
          $pb.ClientContext? ctx, ListUsersRequest request) =>
      _client.invoke<ListUsersResponse>(
          ctx, 'UserService', 'ListUsers', request, ListUsersResponse());

  /// GetUser:取得單一使用者(不含 password_hash)。
  $async.Future<GetUserResponse> getUser(
          $pb.ClientContext? ctx, GetUserRequest request) =>
      _client.invoke<GetUserResponse>(
          ctx, 'UserService', 'GetUser', request, GetUserResponse());

  /// CreateUser:建立員工帳號(super 直接建立;password_hash 依 OAuth 流程填補)。
  $async.Future<CreateUserResponse> createUser(
          $pb.ClientContext? ctx, CreateUserRequest request) =>
      _client.invoke<CreateUserResponse>(
          ctx, 'UserService', 'CreateUser', request, CreateUserResponse());

  /// UpdateUser:更新使用者(name / department_id / phone / employee_no / status)。
  $async.Future<UpdateUserResponse> updateUser(
          $pb.ClientContext? ctx, UpdateUserRequest request) =>
      _client.invoke<UpdateUserResponse>(
          ctx, 'UserService', 'UpdateUser', request, UpdateUserResponse());

  /// AssignRole:角色指派(含 guest 審核:status pending → active 並指派部門與角色)。
  $async.Future<AssignRoleResponse> assignRole(
          $pb.ClientContext? ctx, AssignRoleRequest request) =>
      _client.invoke<AssignRoleResponse>(
          ctx, 'UserService', 'AssignRole', request, AssignRoleResponse());

  /// Deactivate:停用帳號(status → inactive, token_version+1, 刪 session)。
  $async.Future<DeactivateResponse> deactivate(
          $pb.ClientContext? ctx, DeactivateRequest request) =>
      _client.invoke<DeactivateResponse>(
          ctx, 'UserService', 'Deactivate', request, DeactivateResponse());

  /// ForceLogout:強制登出(token_version+1, 使該使用者在途憑證失效)。
  $async.Future<ForceLogoutResponse> forceLogout(
          $pb.ClientContext? ctx, ForceLogoutRequest request) =>
      _client.invoke<ForceLogoutResponse>(
          ctx, 'UserService', 'ForceLogout', request, ForceLogoutResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
