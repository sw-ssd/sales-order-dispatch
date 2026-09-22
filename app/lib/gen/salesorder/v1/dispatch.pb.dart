// This is a generated file - do not edit.
//
// Generated from salesorder/v1/dispatch.proto.

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

/// AssignRouteRequest:指派請求。
class AssignRouteRequest extends $pb.GeneratedMessage {
  factory AssignRouteRequest({
    $core.String? salesOrderId,
    $core.String? routeId,
    $core.String? deliverySequence,
    $core.String? version,
    $core.String? expectedDeliveryDate,
  }) {
    final result = AssignRouteRequest._();
    if (salesOrderId != null) result.salesOrderId = salesOrderId;
    if (routeId != null) result.routeId = routeId;
    if (deliverySequence != null) result.deliverySequence = deliverySequence;
    if (version != null) result.version = version;
    if (expectedDeliveryDate != null)
      result.expectedDeliveryDate = expectedDeliveryDate;
    return result;
  }

  AssignRouteRequest._();

  factory AssignRouteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignRouteRequest()..mergeFromBuffer(data, registry);
  factory AssignRouteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignRouteRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignRouteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: AssignRouteRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'salesOrderId')
    ..aOS(2, _omitFieldNames ? '' : 'routeId')
    ..aOS(3, _omitFieldNames ? '' : 'deliverySequence')
    ..aOS(4, _omitFieldNames ? '' : 'version')
    ..aOS(5, _omitFieldNames ? '' : 'expectedDeliveryDate')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRouteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRouteRequest copyWith(void Function(AssignRouteRequest) updates) =>
      super.copyWith((message) => updates(message as AssignRouteRequest))
          as AssignRouteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AssignRouteRequest() / AssignRouteRequest.new instead')
  static AssignRouteRequest create() => AssignRouteRequest._();
  static $pb.GeneratedMessage $_createMessage() => AssignRouteRequest._();
  @$core.override
  AssignRouteRequest createEmptyInstance() => AssignRouteRequest._();
  @$core.pragma('dart2js:noInline')
  static AssignRouteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignRouteRequest>(
          AssignRouteRequest.$_createMessage);
  static AssignRouteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get salesOrderId => $_getSZ(0);
  @$pb.TagNumber(1)
  set salesOrderId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSalesOrderId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSalesOrderId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get routeId => $_getSZ(1);
  @$pb.TagNumber(2)
  set routeId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRouteId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRouteId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get deliverySequence => $_getSZ(2);
  @$pb.TagNumber(3)
  set deliverySequence($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDeliverySequence() => $_has(2);
  @$pb.TagNumber(3)
  void clearDeliverySequence() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get version => $_getSZ(3);
  @$pb.TagNumber(4)
  set version($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get expectedDeliveryDate => $_getSZ(4);
  @$pb.TagNumber(5)
  set expectedDeliveryDate($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasExpectedDeliveryDate() => $_has(4);
  @$pb.TagNumber(5)
  void clearExpectedDeliveryDate() => $_clearField(5);
}

/// AssignRouteResponse:指派結果。
class AssignRouteResponse extends $pb.GeneratedMessage {
  factory AssignRouteResponse({
    $core.String? salesOrderId,
    $core.String? routeId,
    $core.String? deliverySequence,
    $core.String? version,
  }) {
    final result = AssignRouteResponse._();
    if (salesOrderId != null) result.salesOrderId = salesOrderId;
    if (routeId != null) result.routeId = routeId;
    if (deliverySequence != null) result.deliverySequence = deliverySequence;
    if (version != null) result.version = version;
    return result;
  }

  AssignRouteResponse._();

  factory AssignRouteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignRouteResponse()..mergeFromBuffer(data, registry);
  factory AssignRouteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignRouteResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignRouteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: AssignRouteResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'salesOrderId')
    ..aOS(2, _omitFieldNames ? '' : 'routeId')
    ..aOS(3, _omitFieldNames ? '' : 'deliverySequence')
    ..aOS(4, _omitFieldNames ? '' : 'version')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRouteResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignRouteResponse copyWith(void Function(AssignRouteResponse) updates) =>
      super.copyWith((message) => updates(message as AssignRouteResponse))
          as AssignRouteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use AssignRouteResponse() / AssignRouteResponse.new instead')
  static AssignRouteResponse create() => AssignRouteResponse._();
  static $pb.GeneratedMessage $_createMessage() => AssignRouteResponse._();
  @$core.override
  AssignRouteResponse createEmptyInstance() => AssignRouteResponse._();
  @$core.pragma('dart2js:noInline')
  static AssignRouteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignRouteResponse>(
          AssignRouteResponse.$_createMessage);
  static AssignRouteResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get salesOrderId => $_getSZ(0);
  @$pb.TagNumber(1)
  set salesOrderId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSalesOrderId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSalesOrderId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get routeId => $_getSZ(1);
  @$pb.TagNumber(2)
  set routeId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRouteId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRouteId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get deliverySequence => $_getSZ(2);
  @$pb.TagNumber(3)
  set deliverySequence($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDeliverySequence() => $_has(2);
  @$pb.TagNumber(3)
  void clearDeliverySequence() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get version => $_getSZ(3);
  @$pb.TagNumber(4)
  set version($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => $_clearField(4);
}

/// ConfirmDispatchRequest:批次確認請求。
class ConfirmDispatchRequest extends $pb.GeneratedMessage {
  factory ConfirmDispatchRequest({
    $core.String? routeId,
    $core.String? expectedDeliveryDate,
  }) {
    final result = ConfirmDispatchRequest._();
    if (routeId != null) result.routeId = routeId;
    if (expectedDeliveryDate != null)
      result.expectedDeliveryDate = expectedDeliveryDate;
    return result;
  }

  ConfirmDispatchRequest._();

  factory ConfirmDispatchRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmDispatchRequest()..mergeFromBuffer(data, registry);
  factory ConfirmDispatchRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmDispatchRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfirmDispatchRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ConfirmDispatchRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'routeId')
    ..aOS(2, _omitFieldNames ? '' : 'expectedDeliveryDate')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmDispatchRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmDispatchRequest copyWith(
          void Function(ConfirmDispatchRequest) updates) =>
      super.copyWith((message) => updates(message as ConfirmDispatchRequest))
          as ConfirmDispatchRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ConfirmDispatchRequest() / ConfirmDispatchRequest.new instead')
  static ConfirmDispatchRequest create() => ConfirmDispatchRequest._();
  static $pb.GeneratedMessage $_createMessage() => ConfirmDispatchRequest._();
  @$core.override
  ConfirmDispatchRequest createEmptyInstance() => ConfirmDispatchRequest._();
  @$core.pragma('dart2js:noInline')
  static ConfirmDispatchRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfirmDispatchRequest>(
          ConfirmDispatchRequest.$_createMessage);
  static ConfirmDispatchRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get routeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set routeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRouteId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRouteId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get expectedDeliveryDate => $_getSZ(1);
  @$pb.TagNumber(2)
  set expectedDeliveryDate($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasExpectedDeliveryDate() => $_has(1);
  @$pb.TagNumber(2)
  void clearExpectedDeliveryDate() => $_clearField(2);
}

/// DispatchItemResult:逐筆結果。
class DispatchItemResult extends $pb.GeneratedMessage {
  factory DispatchItemResult({
    $core.String? salesOrderId,
    $core.bool? success,
    $core.String? failReason,
  }) {
    final result = DispatchItemResult._();
    if (salesOrderId != null) result.salesOrderId = salesOrderId;
    if (success != null) result.success = success;
    if (failReason != null) result.failReason = failReason;
    return result;
  }

  DispatchItemResult._();

  factory DispatchItemResult.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DispatchItemResult()..mergeFromBuffer(data, registry);
  factory DispatchItemResult.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DispatchItemResult()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DispatchItemResult',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DispatchItemResult.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'salesOrderId')
    ..aOB(2, _omitFieldNames ? '' : 'success')
    ..aOS(3, _omitFieldNames ? '' : 'failReason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DispatchItemResult clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DispatchItemResult copyWith(void Function(DispatchItemResult) updates) =>
      super.copyWith((message) => updates(message as DispatchItemResult))
          as DispatchItemResult;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DispatchItemResult() / DispatchItemResult.new instead')
  static DispatchItemResult create() => DispatchItemResult._();
  static $pb.GeneratedMessage $_createMessage() => DispatchItemResult._();
  @$core.override
  DispatchItemResult createEmptyInstance() => DispatchItemResult._();
  @$core.pragma('dart2js:noInline')
  static DispatchItemResult getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DispatchItemResult>(
          DispatchItemResult.$_createMessage);
  static DispatchItemResult? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get salesOrderId => $_getSZ(0);
  @$pb.TagNumber(1)
  set salesOrderId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSalesOrderId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSalesOrderId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get success => $_getBF(1);
  @$pb.TagNumber(2)
  set success($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSuccess() => $_has(1);
  @$pb.TagNumber(2)
  void clearSuccess() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get failReason => $_getSZ(2);
  @$pb.TagNumber(3)
  set failReason($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFailReason() => $_has(2);
  @$pb.TagNumber(3)
  void clearFailReason() => $_clearField(3);
}

/// ConfirmDispatchResponse:批次結果。
class ConfirmDispatchResponse extends $pb.GeneratedMessage {
  factory ConfirmDispatchResponse({
    $core.Iterable<DispatchItemResult>? items,
    $core.int? successCount,
  }) {
    final result = ConfirmDispatchResponse._();
    if (items != null) result.items.addAll(items);
    if (successCount != null) result.successCount = successCount;
    return result;
  }

  ConfirmDispatchResponse._();

  factory ConfirmDispatchResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmDispatchResponse()..mergeFromBuffer(data, registry);
  factory ConfirmDispatchResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfirmDispatchResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfirmDispatchResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ConfirmDispatchResponse.$_createMessage)
    ..pPM<DispatchItemResult>(1, _omitFieldNames ? '' : 'items',
        subBuilder: DispatchItemResult.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'successCount')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmDispatchResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfirmDispatchResponse copyWith(
          void Function(ConfirmDispatchResponse) updates) =>
      super.copyWith((message) => updates(message as ConfirmDispatchResponse))
          as ConfirmDispatchResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ConfirmDispatchResponse() / ConfirmDispatchResponse.new instead')
  static ConfirmDispatchResponse create() => ConfirmDispatchResponse._();
  static $pb.GeneratedMessage $_createMessage() => ConfirmDispatchResponse._();
  @$core.override
  ConfirmDispatchResponse createEmptyInstance() => ConfirmDispatchResponse._();
  @$core.pragma('dart2js:noInline')
  static ConfirmDispatchResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfirmDispatchResponse>(
          ConfirmDispatchResponse.$_createMessage);
  static ConfirmDispatchResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<DispatchItemResult> get items => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get successCount => $_getIZ(1);
  @$pb.TagNumber(2)
  set successCount($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSuccessCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearSuccessCount() => $_clearField(2);
}

/// CancelDispatchRequest:取消派車請求。
class CancelDispatchRequest extends $pb.GeneratedMessage {
  factory CancelDispatchRequest({
    $core.String? salesOrderId,
    $core.String? reason,
    $core.bool? acknowledgeReprint,
  }) {
    final result = CancelDispatchRequest._();
    if (salesOrderId != null) result.salesOrderId = salesOrderId;
    if (reason != null) result.reason = reason;
    if (acknowledgeReprint != null)
      result.acknowledgeReprint = acknowledgeReprint;
    return result;
  }

  CancelDispatchRequest._();

  factory CancelDispatchRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDispatchRequest()..mergeFromBuffer(data, registry);
  factory CancelDispatchRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDispatchRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelDispatchRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CancelDispatchRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'salesOrderId')
    ..aOS(2, _omitFieldNames ? '' : 'reason')
    ..aOB(3, _omitFieldNames ? '' : 'acknowledgeReprint')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDispatchRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDispatchRequest copyWith(
          void Function(CancelDispatchRequest) updates) =>
      super.copyWith((message) => updates(message as CancelDispatchRequest))
          as CancelDispatchRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelDispatchRequest() / CancelDispatchRequest.new instead')
  static CancelDispatchRequest create() => CancelDispatchRequest._();
  static $pb.GeneratedMessage $_createMessage() => CancelDispatchRequest._();
  @$core.override
  CancelDispatchRequest createEmptyInstance() => CancelDispatchRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelDispatchRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelDispatchRequest>(
          CancelDispatchRequest.$_createMessage);
  static CancelDispatchRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get salesOrderId => $_getSZ(0);
  @$pb.TagNumber(1)
  set salesOrderId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSalesOrderId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSalesOrderId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get reason => $_getSZ(1);
  @$pb.TagNumber(2)
  set reason($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearReason() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get acknowledgeReprint => $_getBF(2);
  @$pb.TagNumber(3)
  set acknowledgeReprint($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAcknowledgeReprint() => $_has(2);
  @$pb.TagNumber(3)
  void clearAcknowledgeReprint() => $_clearField(3);
}

/// CancelDispatchResponse:取消結果。
class CancelDispatchResponse extends $pb.GeneratedMessage {
  factory CancelDispatchResponse({
    $core.String? salesOrderId,
    $core.String? status,
    $core.String? routeId,
    $core.String? deliverySequence,
    $core.bool? reprintWarning,
  }) {
    final result = CancelDispatchResponse._();
    if (salesOrderId != null) result.salesOrderId = salesOrderId;
    if (status != null) result.status = status;
    if (routeId != null) result.routeId = routeId;
    if (deliverySequence != null) result.deliverySequence = deliverySequence;
    if (reprintWarning != null) result.reprintWarning = reprintWarning;
    return result;
  }

  CancelDispatchResponse._();

  factory CancelDispatchResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDispatchResponse()..mergeFromBuffer(data, registry);
  factory CancelDispatchResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDispatchResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelDispatchResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CancelDispatchResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'salesOrderId')
    ..aOS(2, _omitFieldNames ? '' : 'status')
    ..aOS(3, _omitFieldNames ? '' : 'routeId')
    ..aOS(4, _omitFieldNames ? '' : 'deliverySequence')
    ..aOB(5, _omitFieldNames ? '' : 'reprintWarning')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDispatchResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDispatchResponse copyWith(
          void Function(CancelDispatchResponse) updates) =>
      super.copyWith((message) => updates(message as CancelDispatchResponse))
          as CancelDispatchResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelDispatchResponse() / CancelDispatchResponse.new instead')
  static CancelDispatchResponse create() => CancelDispatchResponse._();
  static $pb.GeneratedMessage $_createMessage() => CancelDispatchResponse._();
  @$core.override
  CancelDispatchResponse createEmptyInstance() => CancelDispatchResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelDispatchResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelDispatchResponse>(
          CancelDispatchResponse.$_createMessage);
  static CancelDispatchResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get salesOrderId => $_getSZ(0);
  @$pb.TagNumber(1)
  set salesOrderId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSalesOrderId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSalesOrderId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get status => $_getSZ(1);
  @$pb.TagNumber(2)
  set status($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get routeId => $_getSZ(2);
  @$pb.TagNumber(3)
  set routeId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRouteId() => $_has(2);
  @$pb.TagNumber(3)
  void clearRouteId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get deliverySequence => $_getSZ(3);
  @$pb.TagNumber(4)
  set deliverySequence($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDeliverySequence() => $_has(3);
  @$pb.TagNumber(4)
  void clearDeliverySequence() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get reprintWarning => $_getBF(4);
  @$pb.TagNumber(5)
  set reprintWarning($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasReprintWarning() => $_has(4);
  @$pb.TagNumber(5)
  void clearReprintWarning() => $_clearField(5);
}

/// DispatchService:派車 API(08 計畫 Task 5.1)。
/// AssignRoute 走看板拖放(樂觀鎖);Confirm 批次轉 processing(逐筆交易);
/// CancelDispatch 退回 pending(保留看板位置);WatchBoard 串流另案(5.2)。
class DispatchServiceApi {
  final $pb.RpcClient _client;

  DispatchServiceApi(this._client);

  /// AssignRoute:指派車次與配送順位(僅 pending;version 樂觀鎖)。
  $async.Future<AssignRouteResponse> assignRoute(
          $pb.ClientContext? ctx, AssignRouteRequest request) =>
      _client.invoke<AssignRouteResponse>(ctx, 'DispatchService', 'AssignRoute',
          request, AssignRouteResponse());

  /// ConfirmDispatch:車次批次確認(逐筆 pending → processing;部分失敗語義)。
  $async.Future<ConfirmDispatchResponse> confirmDispatch(
          $pb.ClientContext? ctx, ConfirmDispatchRequest request) =>
      _client.invoke<ConfirmDispatchResponse>(ctx, 'DispatchService',
          'ConfirmDispatch', request, ConfirmDispatchResponse());

  /// CancelDispatch:取消派車(僅 processing → pending;dept_admin 以上;重印警告)。
  $async.Future<CancelDispatchResponse> cancelDispatch(
          $pb.ClientContext? ctx, CancelDispatchRequest request) =>
      _client.invoke<CancelDispatchResponse>(ctx, 'DispatchService',
          'CancelDispatch', request, CancelDispatchResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
