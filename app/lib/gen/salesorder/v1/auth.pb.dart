// This is a generated file - do not edit.
//
// Generated from salesorder/v1/auth.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// LoginRequest:客戶帳號密碼登入(Task 12;Web 店家分頁與 App 店家登入共用)。
class LoginRequest extends $pb.GeneratedMessage {
  factory LoginRequest({
    $core.String? customerCode,
    $core.String? password,
  }) {
    final result = create();
    if (customerCode != null) result.customerCode = customerCode;
    if (password != null) result.password = password;
    return result;
  }

  LoginRequest._();

  factory LoginRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'customerCode')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRequest clone() => LoginRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRequest copyWith(void Function(LoginRequest) updates) =>
      super.copyWith((message) => updates(message as LoginRequest))
          as LoginRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginRequest create() => LoginRequest._();
  @$core.override
  LoginRequest createEmptyInstance() => create();
  static $pb.PbList<LoginRequest> createRepeated() =>
      $pb.PbList<LoginRequest>();
  @$core.pragma('dart2js:noInline')
  static LoginRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginRequest>(create);
  static LoginRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

/// LoginResponse:登入成功核發的 token 對(Task 13)。
class LoginResponse extends $pb.GeneratedMessage {
  factory LoginResponse({
    $core.String? accessToken,
    $core.String? refreshToken,
    $fixnum.Int64? expiresIn,
  }) {
    final result = create();
    if (accessToken != null) result.accessToken = accessToken;
    if (refreshToken != null) result.refreshToken = refreshToken;
    if (expiresIn != null) result.expiresIn = expiresIn;
    return result;
  }

  LoginResponse._();

  factory LoginResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'accessToken')
    ..aOS(2, _omitFieldNames ? '' : 'refreshToken')
    ..aInt64(3, _omitFieldNames ? '' : 'expiresIn')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse clone() => LoginResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse copyWith(void Function(LoginResponse) updates) =>
      super.copyWith((message) => updates(message as LoginResponse))
          as LoginResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginResponse create() => LoginResponse._();
  @$core.override
  LoginResponse createEmptyInstance() => create();
  static $pb.PbList<LoginResponse> createRepeated() =>
      $pb.PbList<LoginResponse>();
  @$core.pragma('dart2js:noInline')
  static LoginResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginResponse>(create);
  static LoginResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accessToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set accessToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccessToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccessToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get refreshToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set refreshToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRefreshToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearRefreshToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get expiresIn => $_getI64(2);
  @$pb.TagNumber(3)
  set expiresIn($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpiresIn() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpiresIn() => $_clearField(3);
}

/// RefreshRequest:以 refresh token 旋轉換發新 token 對。
class RefreshRequest extends $pb.GeneratedMessage {
  factory RefreshRequest({
    $core.String? refreshToken,
  }) {
    final result = create();
    if (refreshToken != null) result.refreshToken = refreshToken;
    return result;
  }

  RefreshRequest._();

  factory RefreshRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RefreshRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RefreshRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'refreshToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RefreshRequest clone() => RefreshRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RefreshRequest copyWith(void Function(RefreshRequest) updates) =>
      super.copyWith((message) => updates(message as RefreshRequest))
          as RefreshRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RefreshRequest create() => RefreshRequest._();
  @$core.override
  RefreshRequest createEmptyInstance() => create();
  static $pb.PbList<RefreshRequest> createRepeated() =>
      $pb.PbList<RefreshRequest>();
  @$core.pragma('dart2js:noInline')
  static RefreshRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RefreshRequest>(create);
  static RefreshRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get refreshToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set refreshToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRefreshToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearRefreshToken() => $_clearField(1);
}

/// RefreshResponse:旋轉後的新 token 對。
class RefreshResponse extends $pb.GeneratedMessage {
  factory RefreshResponse({
    $core.String? accessToken,
    $core.String? refreshToken,
    $fixnum.Int64? expiresIn,
  }) {
    final result = create();
    if (accessToken != null) result.accessToken = accessToken;
    if (refreshToken != null) result.refreshToken = refreshToken;
    if (expiresIn != null) result.expiresIn = expiresIn;
    return result;
  }

  RefreshResponse._();

  factory RefreshResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RefreshResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RefreshResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'accessToken')
    ..aOS(2, _omitFieldNames ? '' : 'refreshToken')
    ..aInt64(3, _omitFieldNames ? '' : 'expiresIn')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RefreshResponse clone() => RefreshResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RefreshResponse copyWith(void Function(RefreshResponse) updates) =>
      super.copyWith((message) => updates(message as RefreshResponse))
          as RefreshResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RefreshResponse create() => RefreshResponse._();
  @$core.override
  RefreshResponse createEmptyInstance() => create();
  static $pb.PbList<RefreshResponse> createRepeated() =>
      $pb.PbList<RefreshResponse>();
  @$core.pragma('dart2js:noInline')
  static RefreshResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RefreshResponse>(create);
  static RefreshResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accessToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set accessToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccessToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccessToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get refreshToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set refreshToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRefreshToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearRefreshToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get expiresIn => $_getI64(2);
  @$pb.TagNumber(3)
  set expiresIn($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpiresIn() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpiresIn() => $_clearField(3);
}

/// LogoutRequest:登出並撤銷指定 refresh token。
class LogoutRequest extends $pb.GeneratedMessage {
  factory LogoutRequest({
    $core.String? refreshToken,
  }) {
    final result = create();
    if (refreshToken != null) result.refreshToken = refreshToken;
    return result;
  }

  LogoutRequest._();

  factory LogoutRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LogoutRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogoutRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'refreshToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutRequest clone() => LogoutRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutRequest copyWith(void Function(LogoutRequest) updates) =>
      super.copyWith((message) => updates(message as LogoutRequest))
          as LogoutRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LogoutRequest create() => LogoutRequest._();
  @$core.override
  LogoutRequest createEmptyInstance() => create();
  static $pb.PbList<LogoutRequest> createRepeated() =>
      $pb.PbList<LogoutRequest>();
  @$core.pragma('dart2js:noInline')
  static LogoutRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LogoutRequest>(create);
  static LogoutRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get refreshToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set refreshToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRefreshToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearRefreshToken() => $_clearField(1);
}

/// LogoutResponse:登出結果(無內容)。
class LogoutResponse extends $pb.GeneratedMessage {
  factory LogoutResponse() => create();

  LogoutResponse._();

  factory LogoutResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LogoutResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogoutResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutResponse clone() => LogoutResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutResponse copyWith(void Function(LogoutResponse) updates) =>
      super.copyWith((message) => updates(message as LogoutResponse))
          as LogoutResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LogoutResponse create() => LogoutResponse._();
  @$core.override
  LogoutResponse createEmptyInstance() => create();
  static $pb.PbList<LogoutResponse> createRepeated() =>
      $pb.PbList<LogoutResponse>();
  @$core.pragma('dart2js:noInline')
  static LogoutResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LogoutResponse>(create);
  static LogoutResponse? _defaultInstance;
}

/// RegisterCompleteRequest:guest 補完註冊資料(Task 17:選公司、填姓名,轉 pending/approved)。
class RegisterCompleteRequest extends $pb.GeneratedMessage {
  factory RegisterCompleteRequest({
    $core.String? companyId,
    $core.String? name,
  }) {
    final result = create();
    if (companyId != null) result.companyId = companyId;
    if (name != null) result.name = name;
    return result;
  }

  RegisterCompleteRequest._();

  factory RegisterCompleteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RegisterCompleteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterCompleteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterCompleteRequest clone() =>
      RegisterCompleteRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterCompleteRequest copyWith(
          void Function(RegisterCompleteRequest) updates) =>
      super.copyWith((message) => updates(message as RegisterCompleteRequest))
          as RegisterCompleteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RegisterCompleteRequest create() => RegisterCompleteRequest._();
  @$core.override
  RegisterCompleteRequest createEmptyInstance() => create();
  static $pb.PbList<RegisterCompleteRequest> createRepeated() =>
      $pb.PbList<RegisterCompleteRequest>();
  @$core.pragma('dart2js:noInline')
  static RegisterCompleteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterCompleteRequest>(create);
  static RegisterCompleteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

/// RegisterCompleteResponse:註冊完成結果(無內容)。
class RegisterCompleteResponse extends $pb.GeneratedMessage {
  factory RegisterCompleteResponse() => create();

  RegisterCompleteResponse._();

  factory RegisterCompleteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RegisterCompleteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterCompleteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterCompleteResponse clone() =>
      RegisterCompleteResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterCompleteResponse copyWith(
          void Function(RegisterCompleteResponse) updates) =>
      super.copyWith((message) => updates(message as RegisterCompleteResponse))
          as RegisterCompleteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RegisterCompleteResponse create() => RegisterCompleteResponse._();
  @$core.override
  RegisterCompleteResponse createEmptyInstance() => create();
  static $pb.PbList<RegisterCompleteResponse> createRepeated() =>
      $pb.PbList<RegisterCompleteResponse>();
  @$core.pragma('dart2js:noInline')
  static RegisterCompleteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterCompleteResponse>(create);
  static RegisterCompleteResponse? _defaultInstance;
}

/// QRLoginRequest:App 掃描客戶 QR 後兌換(前段:帶出身分,後續以子帳號 + 密碼完成登入)。
class QRLoginRequest extends $pb.GeneratedMessage {
  factory QRLoginRequest({
    $core.String? token,
  }) {
    final result = create();
    if (token != null) result.token = token;
    return result;
  }

  QRLoginRequest._();

  factory QRLoginRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory QRLoginRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'QRLoginRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QRLoginRequest clone() => QRLoginRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QRLoginRequest copyWith(void Function(QRLoginRequest) updates) =>
      super.copyWith((message) => updates(message as QRLoginRequest))
          as QRLoginRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static QRLoginRequest create() => QRLoginRequest._();
  @$core.override
  QRLoginRequest createEmptyInstance() => create();
  static $pb.PbList<QRLoginRequest> createRepeated() =>
      $pb.PbList<QRLoginRequest>();
  @$core.pragma('dart2js:noInline')
  static QRLoginRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<QRLoginRequest>(create);
  static QRLoginRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);
}

/// Account:可選店家子帳號(僅店家子帳號;不含主帳號與業務子帳號)。
class QRLoginResponse_Account extends $pb.GeneratedMessage {
  factory QRLoginResponse_Account({
    $core.String? id,
    $core.String? accountName,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (accountName != null) result.accountName = accountName;
    return result;
  }

  QRLoginResponse_Account._();

  factory QRLoginResponse_Account.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory QRLoginResponse_Account.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'QRLoginResponse.Account',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'accountName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QRLoginResponse_Account clone() =>
      QRLoginResponse_Account()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QRLoginResponse_Account copyWith(
          void Function(QRLoginResponse_Account) updates) =>
      super.copyWith((message) => updates(message as QRLoginResponse_Account))
          as QRLoginResponse_Account;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static QRLoginResponse_Account create() => QRLoginResponse_Account._();
  @$core.override
  QRLoginResponse_Account createEmptyInstance() => create();
  static $pb.PbList<QRLoginResponse_Account> createRepeated() =>
      $pb.PbList<QRLoginResponse_Account>();
  @$core.pragma('dart2js:noInline')
  static QRLoginResponse_Account getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<QRLoginResponse_Account>(create);
  static QRLoginResponse_Account? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get accountName => $_getSZ(1);
  @$pb.TagNumber(2)
  set accountName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAccountName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAccountName() => $_clearField(2);
}

/// QRLoginResponse:兌換結果 — 公司/客戶識別資訊與可選店家子帳號清單。
class QRLoginResponse extends $pb.GeneratedMessage {
  factory QRLoginResponse({
    $core.String? companyId,
    $core.String? companyName,
    $core.String? customerCode,
    $core.String? customerName,
    $core.Iterable<QRLoginResponse_Account>? accounts,
  }) {
    final result = create();
    if (companyId != null) result.companyId = companyId;
    if (companyName != null) result.companyName = companyName;
    if (customerCode != null) result.customerCode = customerCode;
    if (customerName != null) result.customerName = customerName;
    if (accounts != null) result.accounts.addAll(accounts);
    return result;
  }

  QRLoginResponse._();

  factory QRLoginResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory QRLoginResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'QRLoginResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'companyName')
    ..aOS(3, _omitFieldNames ? '' : 'customerCode')
    ..aOS(4, _omitFieldNames ? '' : 'customerName')
    ..pc<QRLoginResponse_Account>(
        5, _omitFieldNames ? '' : 'accounts', $pb.PbFieldType.PM,
        subBuilder: QRLoginResponse_Account.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QRLoginResponse clone() => QRLoginResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QRLoginResponse copyWith(void Function(QRLoginResponse) updates) =>
      super.copyWith((message) => updates(message as QRLoginResponse))
          as QRLoginResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static QRLoginResponse create() => QRLoginResponse._();
  @$core.override
  QRLoginResponse createEmptyInstance() => create();
  static $pb.PbList<QRLoginResponse> createRepeated() =>
      $pb.PbList<QRLoginResponse>();
  @$core.pragma('dart2js:noInline')
  static QRLoginResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<QRLoginResponse>(create);
  static QRLoginResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get companyName => $_getSZ(1);
  @$pb.TagNumber(2)
  set companyName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCompanyName() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompanyName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get customerCode => $_getSZ(2);
  @$pb.TagNumber(3)
  set customerCode($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCustomerCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearCustomerCode() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get customerName => $_getSZ(3);
  @$pb.TagNumber(4)
  set customerName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCustomerName() => $_has(3);
  @$pb.TagNumber(4)
  void clearCustomerName() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<QRLoginResponse_Account> get accounts => $_getList(4);
}

/// AuthService:認證相關 RPC。
class AuthServiceApi {
  final $pb.RpcClient _client;

  AuthServiceApi(this._client);

  /// Login:客戶帳號密碼登入。
  $async.Future<LoginResponse> login(
          $pb.ClientContext? ctx, LoginRequest request) =>
      _client.invoke<LoginResponse>(
          ctx, 'AuthService', 'Login', request, LoginResponse());

  /// Refresh:refresh token 旋轉換發。
  $async.Future<RefreshResponse> refresh(
          $pb.ClientContext? ctx, RefreshRequest request) =>
      _client.invoke<RefreshResponse>(
          ctx, 'AuthService', 'Refresh', request, RefreshResponse());

  /// Logout:撤銷 refresh token(冪等)。
  $async.Future<LogoutResponse> logout(
          $pb.ClientContext? ctx, LogoutRequest request) =>
      _client.invoke<LogoutResponse>(
          ctx, 'AuthService', 'Logout', request, LogoutResponse());

  /// RegisterComplete:guest 補完註冊資料。
  $async.Future<RegisterCompleteResponse> registerComplete(
          $pb.ClientContext? ctx, RegisterCompleteRequest request) =>
      _client.invoke<RegisterCompleteResponse>(ctx, 'AuthService',
          'RegisterComplete', request, RegisterCompleteResponse());

  /// QRLogin:QR token 兌換,回公司/客戶與可選子帳號清單。
  $async.Future<QRLoginResponse> qRLogin(
          $pb.ClientContext? ctx, QRLoginRequest request) =>
      _client.invoke<QRLoginResponse>(
          ctx, 'AuthService', 'QRLogin', request, QRLoginResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
