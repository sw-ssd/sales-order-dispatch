// This is a generated file - do not edit.
//
// Generated from salesorder/v1/devices.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// RegisterDeviceRequest:註冊請求。
class RegisterDeviceRequest extends $pb.GeneratedMessage {
  factory RegisterDeviceRequest({
    $core.String? platform,
    $core.String? fcmToken,
    $core.String? deviceName,
  }) {
    final result = RegisterDeviceRequest._();
    if (platform != null) result.platform = platform;
    if (fcmToken != null) result.fcmToken = fcmToken;
    if (deviceName != null) result.deviceName = deviceName;
    return result;
  }

  RegisterDeviceRequest._();

  factory RegisterDeviceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterDeviceRequest()..mergeFromBuffer(data, registry);
  factory RegisterDeviceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterDeviceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterDeviceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: RegisterDeviceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'platform')
    ..aOS(2, _omitFieldNames ? '' : 'fcmToken')
    ..aOS(3, _omitFieldNames ? '' : 'deviceName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterDeviceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterDeviceRequest copyWith(
          void Function(RegisterDeviceRequest) updates) =>
      super.copyWith((message) => updates(message as RegisterDeviceRequest))
          as RegisterDeviceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RegisterDeviceRequest() / RegisterDeviceRequest.new instead')
  static RegisterDeviceRequest create() => RegisterDeviceRequest._();
  static $pb.GeneratedMessage $_createMessage() => RegisterDeviceRequest._();
  @$core.override
  RegisterDeviceRequest createEmptyInstance() => RegisterDeviceRequest._();
  @$core.pragma('dart2js:noInline')
  static RegisterDeviceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterDeviceRequest>(
          RegisterDeviceRequest.$_createMessage);
  static RegisterDeviceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get platform => $_getSZ(0);
  @$pb.TagNumber(1)
  set platform($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlatform() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlatform() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get fcmToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set fcmToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFcmToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearFcmToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get deviceName => $_getSZ(2);
  @$pb.TagNumber(3)
  set deviceName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDeviceName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDeviceName() => $_clearField(3);
}

/// RegisterDeviceResponse:註冊結果。
class RegisterDeviceResponse extends $pb.GeneratedMessage {
  factory RegisterDeviceResponse({
    $core.String? deviceId,
  }) {
    final result = RegisterDeviceResponse._();
    if (deviceId != null) result.deviceId = deviceId;
    return result;
  }

  RegisterDeviceResponse._();

  factory RegisterDeviceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterDeviceResponse()..mergeFromBuffer(data, registry);
  factory RegisterDeviceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterDeviceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterDeviceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: RegisterDeviceResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'deviceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterDeviceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterDeviceResponse copyWith(
          void Function(RegisterDeviceResponse) updates) =>
      super.copyWith((message) => updates(message as RegisterDeviceResponse))
          as RegisterDeviceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RegisterDeviceResponse() / RegisterDeviceResponse.new instead')
  static RegisterDeviceResponse create() => RegisterDeviceResponse._();
  static $pb.GeneratedMessage $_createMessage() => RegisterDeviceResponse._();
  @$core.override
  RegisterDeviceResponse createEmptyInstance() => RegisterDeviceResponse._();
  @$core.pragma('dart2js:noInline')
  static RegisterDeviceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterDeviceResponse>(
          RegisterDeviceResponse.$_createMessage);
  static RegisterDeviceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get deviceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set deviceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeviceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeviceId() => $_clearField(1);
}

/// UnregisterDeviceRequest:註銷請求。
class UnregisterDeviceRequest extends $pb.GeneratedMessage {
  factory UnregisterDeviceRequest({
    $core.String? fcmToken,
  }) {
    final result = UnregisterDeviceRequest._();
    if (fcmToken != null) result.fcmToken = fcmToken;
    return result;
  }

  UnregisterDeviceRequest._();

  factory UnregisterDeviceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnregisterDeviceRequest()..mergeFromBuffer(data, registry);
  factory UnregisterDeviceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnregisterDeviceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UnregisterDeviceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UnregisterDeviceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'fcmToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnregisterDeviceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnregisterDeviceRequest copyWith(
          void Function(UnregisterDeviceRequest) updates) =>
      super.copyWith((message) => updates(message as UnregisterDeviceRequest))
          as UnregisterDeviceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UnregisterDeviceRequest() / UnregisterDeviceRequest.new instead')
  static UnregisterDeviceRequest create() => UnregisterDeviceRequest._();
  static $pb.GeneratedMessage $_createMessage() => UnregisterDeviceRequest._();
  @$core.override
  UnregisterDeviceRequest createEmptyInstance() => UnregisterDeviceRequest._();
  @$core.pragma('dart2js:noInline')
  static UnregisterDeviceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UnregisterDeviceRequest>(
          UnregisterDeviceRequest.$_createMessage);
  static UnregisterDeviceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get fcmToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set fcmToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFcmToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearFcmToken() => $_clearField(1);
}

/// UnregisterDeviceResponse:註銷結果。
class UnregisterDeviceResponse extends $pb.GeneratedMessage {
  factory UnregisterDeviceResponse() => UnregisterDeviceResponse._();

  UnregisterDeviceResponse._();

  factory UnregisterDeviceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnregisterDeviceResponse()..mergeFromBuffer(data, registry);
  factory UnregisterDeviceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnregisterDeviceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UnregisterDeviceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UnregisterDeviceResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnregisterDeviceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnregisterDeviceResponse copyWith(
          void Function(UnregisterDeviceResponse) updates) =>
      super.copyWith((message) => updates(message as UnregisterDeviceResponse))
          as UnregisterDeviceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UnregisterDeviceResponse() / UnregisterDeviceResponse.new instead')
  static UnregisterDeviceResponse create() => UnregisterDeviceResponse._();
  static $pb.GeneratedMessage $_createMessage() => UnregisterDeviceResponse._();
  @$core.override
  UnregisterDeviceResponse createEmptyInstance() =>
      UnregisterDeviceResponse._();
  @$core.pragma('dart2js:noInline')
  static UnregisterDeviceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UnregisterDeviceResponse>(
          UnregisterDeviceResponse.$_createMessage);
  static UnregisterDeviceResponse? _defaultInstance;
}

/// DeviceService:裝置 FCM token 註冊/註銷(07 計畫 Task 4.3.4)。
class DeviceServiceApi {
  final $pb.RpcClient _client;

  DeviceServiceApi(this._client);

  /// RegisterDevice:註冊裝置(冪等;換帳轉移歸屬)。
  $async.Future<RegisterDeviceResponse> registerDevice(
          $pb.ClientContext? ctx, RegisterDeviceRequest request) =>
      _client.invoke<RegisterDeviceResponse>(ctx, 'DeviceService',
          'RegisterDevice', request, RegisterDeviceResponse());

  /// UnregisterDevice:註銷裝置(冪等;查無不報錯)。
  $async.Future<UnregisterDeviceResponse> unregisterDevice(
          $pb.ClientContext? ctx, UnregisterDeviceRequest request) =>
      _client.invoke<UnregisterDeviceResponse>(ctx, 'DeviceService',
          'UnregisterDevice', request, UnregisterDeviceResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
