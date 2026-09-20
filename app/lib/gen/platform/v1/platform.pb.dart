// This is a generated file - do not edit.
//
// Generated from platform/v1/platform.proto.

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

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// TenantSummary:平台視角的租戶概況(跨租戶,僅 operator 可讀)。
class TenantSummary extends $pb.GeneratedMessage {
  factory TenantSummary({
    $core.String? companyId,
    $core.String? companyName,
    $core.String? planCode,
    $core.String? planName,
    $core.String? subscriptionStatus,
    $core.int? seatCount,
    $core.String? currentPeriodEnd,
    $core.bool? overdue,
  }) {
    final result = TenantSummary._();
    if (companyId != null) result.companyId = companyId;
    if (companyName != null) result.companyName = companyName;
    if (planCode != null) result.planCode = planCode;
    if (planName != null) result.planName = planName;
    if (subscriptionStatus != null)
      result.subscriptionStatus = subscriptionStatus;
    if (seatCount != null) result.seatCount = seatCount;
    if (currentPeriodEnd != null) result.currentPeriodEnd = currentPeriodEnd;
    if (overdue != null) result.overdue = overdue;
    return result;
  }

  TenantSummary._();

  factory TenantSummary.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TenantSummary()..mergeFromBuffer(data, registry);
  factory TenantSummary.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TenantSummary()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TenantSummary',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: TenantSummary.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'companyName')
    ..aOS(3, _omitFieldNames ? '' : 'planCode')
    ..aOS(4, _omitFieldNames ? '' : 'planName')
    ..aOS(5, _omitFieldNames ? '' : 'subscriptionStatus')
    ..aI(6, _omitFieldNames ? '' : 'seatCount')
    ..aOS(7, _omitFieldNames ? '' : 'currentPeriodEnd')
    ..aOB(8, _omitFieldNames ? '' : 'overdue')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TenantSummary clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TenantSummary copyWith(void Function(TenantSummary) updates) =>
      super.copyWith((message) => updates(message as TenantSummary))
          as TenantSummary;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TenantSummary() / TenantSummary.new instead')
  static TenantSummary create() => TenantSummary._();
  static $pb.GeneratedMessage $_createMessage() => TenantSummary._();
  @$core.override
  TenantSummary createEmptyInstance() => TenantSummary._();
  @$core.pragma('dart2js:noInline')
  static TenantSummary getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TenantSummary>(
          TenantSummary.$_createMessage);
  static TenantSummary? _defaultInstance;

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
  $core.String get planCode => $_getSZ(2);
  @$pb.TagNumber(3)
  set planCode($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPlanCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlanCode() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get planName => $_getSZ(3);
  @$pb.TagNumber(4)
  set planName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPlanName() => $_has(3);
  @$pb.TagNumber(4)
  void clearPlanName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get subscriptionStatus => $_getSZ(4);
  @$pb.TagNumber(5)
  set subscriptionStatus($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSubscriptionStatus() => $_has(4);
  @$pb.TagNumber(5)
  void clearSubscriptionStatus() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get seatCount => $_getIZ(5);
  @$pb.TagNumber(6)
  set seatCount($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSeatCount() => $_has(5);
  @$pb.TagNumber(6)
  void clearSeatCount() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get currentPeriodEnd => $_getSZ(6);
  @$pb.TagNumber(7)
  set currentPeriodEnd($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCurrentPeriodEnd() => $_has(6);
  @$pb.TagNumber(7)
  void clearCurrentPeriodEnd() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get overdue => $_getBF(7);
  @$pb.TagNumber(8)
  set overdue($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasOverdue() => $_has(7);
  @$pb.TagNumber(8)
  void clearOverdue() => $_clearField(8);
}

/// PlatformPagination:平台查詢統一分頁結果。
class PlatformPagination extends $pb.GeneratedMessage {
  factory PlatformPagination({
    $core.int? page,
    $core.int? pageSize,
    $core.int? total,
  }) {
    final result = PlatformPagination._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (total != null) result.total = total;
    return result;
  }

  PlatformPagination._();

  factory PlatformPagination.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PlatformPagination()..mergeFromBuffer(data, registry);
  factory PlatformPagination.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PlatformPagination()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PlatformPagination',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: PlatformPagination.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aI(3, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlatformPagination clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlatformPagination copyWith(void Function(PlatformPagination) updates) =>
      super.copyWith((message) => updates(message as PlatformPagination))
          as PlatformPagination;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PlatformPagination() / PlatformPagination.new instead')
  static PlatformPagination create() => PlatformPagination._();
  static $pb.GeneratedMessage $_createMessage() => PlatformPagination._();
  @$core.override
  PlatformPagination createEmptyInstance() => PlatformPagination._();
  @$core.pragma('dart2js:noInline')
  static PlatformPagination getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PlatformPagination>(
          PlatformPagination.$_createMessage);
  static PlatformPagination? _defaultInstance;

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
  $core.int get total => $_getIZ(2);
  @$pb.TagNumber(3)
  set total($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTotal() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotal() => $_clearField(3);
}

class ListTenantsRequest extends $pb.GeneratedMessage {
  factory ListTenantsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.String? status,
  }) {
    final result = ListTenantsRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (status != null) result.status = status;
    return result;
  }

  ListTenantsRequest._();

  factory ListTenantsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTenantsRequest()..mergeFromBuffer(data, registry);
  factory ListTenantsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTenantsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTenantsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListTenantsRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOS(4, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTenantsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTenantsRequest copyWith(void Function(ListTenantsRequest) updates) =>
      super.copyWith((message) => updates(message as ListTenantsRequest))
          as ListTenantsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListTenantsRequest() / ListTenantsRequest.new instead')
  static ListTenantsRequest create() => ListTenantsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListTenantsRequest._();
  @$core.override
  ListTenantsRequest createEmptyInstance() => ListTenantsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListTenantsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTenantsRequest>(
          ListTenantsRequest.$_createMessage);
  static ListTenantsRequest? _defaultInstance;

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
  $core.String get status => $_getSZ(3);
  @$pb.TagNumber(4)
  set status($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);
}

class ListTenantsResponse extends $pb.GeneratedMessage {
  factory ListTenantsResponse({
    $core.Iterable<TenantSummary>? tenants,
    PlatformPagination? pagination,
  }) {
    final result = ListTenantsResponse._();
    if (tenants != null) result.tenants.addAll(tenants);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListTenantsResponse._();

  factory ListTenantsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTenantsResponse()..mergeFromBuffer(data, registry);
  factory ListTenantsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTenantsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTenantsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListTenantsResponse.$_createMessage)
    ..pPM<TenantSummary>(1, _omitFieldNames ? '' : 'tenants',
        subBuilder: TenantSummary.$_createMessage)
    ..aOM<PlatformPagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: PlatformPagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTenantsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTenantsResponse copyWith(void Function(ListTenantsResponse) updates) =>
      super.copyWith((message) => updates(message as ListTenantsResponse))
          as ListTenantsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListTenantsResponse() / ListTenantsResponse.new instead')
  static ListTenantsResponse create() => ListTenantsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListTenantsResponse._();
  @$core.override
  ListTenantsResponse createEmptyInstance() => ListTenantsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListTenantsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTenantsResponse>(
          ListTenantsResponse.$_createMessage);
  static ListTenantsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<TenantSummary> get tenants => $_getList(0);

  @$pb.TagNumber(2)
  PlatformPagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination(PlatformPagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  PlatformPagination ensurePagination() => $_ensure(1);
}

class GetTenantRequest extends $pb.GeneratedMessage {
  factory GetTenantRequest({
    $core.String? companyId,
  }) {
    final result = GetTenantRequest._();
    if (companyId != null) result.companyId = companyId;
    return result;
  }

  GetTenantRequest._();

  factory GetTenantRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantRequest()..mergeFromBuffer(data, registry);
  factory GetTenantRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTenantRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetTenantRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantRequest copyWith(void Function(GetTenantRequest) updates) =>
      super.copyWith((message) => updates(message as GetTenantRequest))
          as GetTenantRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetTenantRequest() / GetTenantRequest.new instead')
  static GetTenantRequest create() => GetTenantRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetTenantRequest._();
  @$core.override
  GetTenantRequest createEmptyInstance() => GetTenantRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTenantRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTenantRequest>(
          GetTenantRequest.$_createMessage);
  static GetTenantRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);
}

class GetTenantResponse extends $pb.GeneratedMessage {
  factory GetTenantResponse({
    TenantSummary? tenant,
    $core.Iterable<TenantOverride>? overrides,
  }) {
    final result = GetTenantResponse._();
    if (tenant != null) result.tenant = tenant;
    if (overrides != null) result.overrides.addAll(overrides);
    return result;
  }

  GetTenantResponse._();

  factory GetTenantResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantResponse()..mergeFromBuffer(data, registry);
  factory GetTenantResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTenantResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetTenantResponse.$_createMessage)
    ..aOM<TenantSummary>(1, _omitFieldNames ? '' : 'tenant',
        subBuilder: TenantSummary.$_createMessage)
    ..pPM<TenantOverride>(2, _omitFieldNames ? '' : 'overrides',
        subBuilder: TenantOverride.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantResponse copyWith(void Function(GetTenantResponse) updates) =>
      super.copyWith((message) => updates(message as GetTenantResponse))
          as GetTenantResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetTenantResponse() / GetTenantResponse.new instead')
  static GetTenantResponse create() => GetTenantResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetTenantResponse._();
  @$core.override
  GetTenantResponse createEmptyInstance() => GetTenantResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTenantResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTenantResponse>(
          GetTenantResponse.$_createMessage);
  static GetTenantResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TenantSummary get tenant => $_getN(0);
  @$pb.TagNumber(1)
  set tenant(TenantSummary value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTenant() => $_has(0);
  @$pb.TagNumber(1)
  void clearTenant() => $_clearField(1);
  @$pb.TagNumber(1)
  TenantSummary ensureTenant() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<TenantOverride> get overrides => $_getList(1);
}

/// TenantOverride:平台對單一租戶的功能覆寫。
class TenantOverride extends $pb.GeneratedMessage {
  factory TenantOverride({
    $core.String? id,
    $core.String? featureCode,
    $core.bool? enabledSet,
    $core.bool? enabled,
    $core.bool? limitSet,
    $fixnum.Int64? limitValue,
    $core.String? reason,
    $core.String? owner,
    $core.String? expiresAt,
  }) {
    final result = TenantOverride._();
    if (id != null) result.id = id;
    if (featureCode != null) result.featureCode = featureCode;
    if (enabledSet != null) result.enabledSet = enabledSet;
    if (enabled != null) result.enabled = enabled;
    if (limitSet != null) result.limitSet = limitSet;
    if (limitValue != null) result.limitValue = limitValue;
    if (reason != null) result.reason = reason;
    if (owner != null) result.owner = owner;
    if (expiresAt != null) result.expiresAt = expiresAt;
    return result;
  }

  TenantOverride._();

  factory TenantOverride.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TenantOverride()..mergeFromBuffer(data, registry);
  factory TenantOverride.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TenantOverride()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TenantOverride',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: TenantOverride.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'featureCode')
    ..aOB(3, _omitFieldNames ? '' : 'enabledSet')
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..aOB(5, _omitFieldNames ? '' : 'limitSet')
    ..aInt64(6, _omitFieldNames ? '' : 'limitValue')
    ..aOS(7, _omitFieldNames ? '' : 'reason')
    ..aOS(8, _omitFieldNames ? '' : 'owner')
    ..aOS(9, _omitFieldNames ? '' : 'expiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TenantOverride clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TenantOverride copyWith(void Function(TenantOverride) updates) =>
      super.copyWith((message) => updates(message as TenantOverride))
          as TenantOverride;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TenantOverride() / TenantOverride.new instead')
  static TenantOverride create() => TenantOverride._();
  static $pb.GeneratedMessage $_createMessage() => TenantOverride._();
  @$core.override
  TenantOverride createEmptyInstance() => TenantOverride._();
  @$core.pragma('dart2js:noInline')
  static TenantOverride getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TenantOverride>(
          TenantOverride.$_createMessage);
  static TenantOverride? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get featureCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set featureCode($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFeatureCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearFeatureCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get enabledSet => $_getBF(2);
  @$pb.TagNumber(3)
  set enabledSet($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEnabledSet() => $_has(2);
  @$pb.TagNumber(3)
  void clearEnabledSet() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get enabled => $_getBF(3);
  @$pb.TagNumber(4)
  set enabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnabled() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get limitSet => $_getBF(4);
  @$pb.TagNumber(5)
  set limitSet($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLimitSet() => $_has(4);
  @$pb.TagNumber(5)
  void clearLimitSet() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get limitValue => $_getI64(5);
  @$pb.TagNumber(6)
  set limitValue($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasLimitValue() => $_has(5);
  @$pb.TagNumber(6)
  void clearLimitValue() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get reason => $_getSZ(6);
  @$pb.TagNumber(7)
  set reason($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasReason() => $_has(6);
  @$pb.TagNumber(7)
  void clearReason() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get owner => $_getSZ(7);
  @$pb.TagNumber(8)
  set owner($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasOwner() => $_has(7);
  @$pb.TagNumber(8)
  void clearOwner() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get expiresAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set expiresAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasExpiresAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearExpiresAt() => $_clearField(9);
}

class ListPlansRequest extends $pb.GeneratedMessage {
  factory ListPlansRequest() => ListPlansRequest._();

  ListPlansRequest._();

  factory ListPlansRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlansRequest()..mergeFromBuffer(data, registry);
  factory ListPlansRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlansRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListPlansRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListPlansRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlansRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlansRequest copyWith(void Function(ListPlansRequest) updates) =>
      super.copyWith((message) => updates(message as ListPlansRequest))
          as ListPlansRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListPlansRequest() / ListPlansRequest.new instead')
  static ListPlansRequest create() => ListPlansRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListPlansRequest._();
  @$core.override
  ListPlansRequest createEmptyInstance() => ListPlansRequest._();
  @$core.pragma('dart2js:noInline')
  static ListPlansRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListPlansRequest>(
          ListPlansRequest.$_createMessage);
  static ListPlansRequest? _defaultInstance;
}

class ListPlansResponse extends $pb.GeneratedMessage {
  factory ListPlansResponse({
    $core.Iterable<Plan>? plans,
  }) {
    final result = ListPlansResponse._();
    if (plans != null) result.plans.addAll(plans);
    return result;
  }

  ListPlansResponse._();

  factory ListPlansResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlansResponse()..mergeFromBuffer(data, registry);
  factory ListPlansResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlansResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListPlansResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListPlansResponse.$_createMessage)
    ..pPM<Plan>(1, _omitFieldNames ? '' : 'plans',
        subBuilder: Plan.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlansResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlansResponse copyWith(void Function(ListPlansResponse) updates) =>
      super.copyWith((message) => updates(message as ListPlansResponse))
          as ListPlansResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListPlansResponse() / ListPlansResponse.new instead')
  static ListPlansResponse create() => ListPlansResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListPlansResponse._();
  @$core.override
  ListPlansResponse createEmptyInstance() => ListPlansResponse._();
  @$core.pragma('dart2js:noInline')
  static ListPlansResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListPlansResponse>(
          ListPlansResponse.$_createMessage);
  static ListPlansResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Plan> get plans => $_getList(0);
}

/// Plan:方案(含各計費週期價格)。
class Plan extends $pb.GeneratedMessage {
  factory Plan({
    $core.String? id,
    $core.String? code,
    $core.String? name,
    $core.String? status,
    $core.int? sortOrder,
    $core.Iterable<PlanPrice>? prices,
  }) {
    final result = Plan._();
    if (id != null) result.id = id;
    if (code != null) result.code = code;
    if (name != null) result.name = name;
    if (status != null) result.status = status;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (prices != null) result.prices.addAll(prices);
    return result;
  }

  Plan._();

  factory Plan.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Plan()..mergeFromBuffer(data, registry);
  factory Plan.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Plan()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Plan',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: Plan.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'status')
    ..aI(5, _omitFieldNames ? '' : 'sortOrder')
    ..pPM<PlanPrice>(6, _omitFieldNames ? '' : 'prices',
        subBuilder: PlanPrice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Plan clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Plan copyWith(void Function(Plan) updates) =>
      super.copyWith((message) => updates(message as Plan)) as Plan;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Plan() / Plan.new instead')
  static Plan create() => Plan._();
  static $pb.GeneratedMessage $_createMessage() => Plan._();
  @$core.override
  Plan createEmptyInstance() => Plan._();
  @$core.pragma('dart2js:noInline')
  static Plan getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Plan>(Plan.$_createMessage);
  static Plan? _defaultInstance;

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
  $core.String get status => $_getSZ(3);
  @$pb.TagNumber(4)
  set status($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get sortOrder => $_getIZ(4);
  @$pb.TagNumber(5)
  set sortOrder($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSortOrder() => $_has(4);
  @$pb.TagNumber(5)
  void clearSortOrder() => $_clearField(5);

  @$pb.TagNumber(6)
  $pb.PbList<PlanPrice> get prices => $_getList(5);
}

class PlanPrice extends $pb.GeneratedMessage {
  factory PlanPrice({
    $core.String? billingCycle,
    $core.String? basePrice,
    $core.String? seatPrice,
    $core.String? currency,
    $core.String? effectiveFrom,
  }) {
    final result = PlanPrice._();
    if (billingCycle != null) result.billingCycle = billingCycle;
    if (basePrice != null) result.basePrice = basePrice;
    if (seatPrice != null) result.seatPrice = seatPrice;
    if (currency != null) result.currency = currency;
    if (effectiveFrom != null) result.effectiveFrom = effectiveFrom;
    return result;
  }

  PlanPrice._();

  factory PlanPrice.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PlanPrice()..mergeFromBuffer(data, registry);
  factory PlanPrice.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PlanPrice()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PlanPrice',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: PlanPrice.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'billingCycle')
    ..aOS(2, _omitFieldNames ? '' : 'basePrice')
    ..aOS(3, _omitFieldNames ? '' : 'seatPrice')
    ..aOS(4, _omitFieldNames ? '' : 'currency')
    ..aOS(5, _omitFieldNames ? '' : 'effectiveFrom')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlanPrice clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlanPrice copyWith(void Function(PlanPrice) updates) =>
      super.copyWith((message) => updates(message as PlanPrice)) as PlanPrice;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PlanPrice() / PlanPrice.new instead')
  static PlanPrice create() => PlanPrice._();
  static $pb.GeneratedMessage $_createMessage() => PlanPrice._();
  @$core.override
  PlanPrice createEmptyInstance() => PlanPrice._();
  @$core.pragma('dart2js:noInline')
  static PlanPrice getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PlanPrice>(PlanPrice.$_createMessage);
  static PlanPrice? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get billingCycle => $_getSZ(0);
  @$pb.TagNumber(1)
  set billingCycle($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBillingCycle() => $_has(0);
  @$pb.TagNumber(1)
  void clearBillingCycle() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get basePrice => $_getSZ(1);
  @$pb.TagNumber(2)
  set basePrice($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBasePrice() => $_has(1);
  @$pb.TagNumber(2)
  void clearBasePrice() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get seatPrice => $_getSZ(2);
  @$pb.TagNumber(3)
  set seatPrice($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSeatPrice() => $_has(2);
  @$pb.TagNumber(3)
  void clearSeatPrice() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(3);
  @$pb.TagNumber(4)
  set currency($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(3);
  @$pb.TagNumber(4)
  void clearCurrency() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get effectiveFrom => $_getSZ(4);
  @$pb.TagNumber(5)
  set effectiveFrom($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasEffectiveFrom() => $_has(4);
  @$pb.TagNumber(5)
  void clearEffectiveFrom() => $_clearField(5);
}

class GetPlanEntitlementsRequest extends $pb.GeneratedMessage {
  factory GetPlanEntitlementsRequest({
    $core.String? planCode,
  }) {
    final result = GetPlanEntitlementsRequest._();
    if (planCode != null) result.planCode = planCode;
    return result;
  }

  GetPlanEntitlementsRequest._();

  factory GetPlanEntitlementsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetPlanEntitlementsRequest()..mergeFromBuffer(data, registry);
  factory GetPlanEntitlementsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetPlanEntitlementsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetPlanEntitlementsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetPlanEntitlementsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'planCode')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPlanEntitlementsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPlanEntitlementsRequest copyWith(
          void Function(GetPlanEntitlementsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetPlanEntitlementsRequest))
          as GetPlanEntitlementsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetPlanEntitlementsRequest() / GetPlanEntitlementsRequest.new instead')
  static GetPlanEntitlementsRequest create() => GetPlanEntitlementsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetPlanEntitlementsRequest._();
  @$core.override
  GetPlanEntitlementsRequest createEmptyInstance() =>
      GetPlanEntitlementsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetPlanEntitlementsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetPlanEntitlementsRequest>(
          GetPlanEntitlementsRequest.$_createMessage);
  static GetPlanEntitlementsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get planCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set planCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlanCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlanCode() => $_clearField(1);
}

class GetPlanEntitlementsResponse extends $pb.GeneratedMessage {
  factory GetPlanEntitlementsResponse({
    $core.Iterable<FeatureEntitlement>? entitlements,
    $core.Iterable<Feature>? features,
  }) {
    final result = GetPlanEntitlementsResponse._();
    if (entitlements != null) result.entitlements.addAll(entitlements);
    if (features != null) result.features.addAll(features);
    return result;
  }

  GetPlanEntitlementsResponse._();

  factory GetPlanEntitlementsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetPlanEntitlementsResponse()..mergeFromBuffer(data, registry);
  factory GetPlanEntitlementsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetPlanEntitlementsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetPlanEntitlementsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetPlanEntitlementsResponse.$_createMessage)
    ..pPM<FeatureEntitlement>(1, _omitFieldNames ? '' : 'entitlements',
        subBuilder: FeatureEntitlement.$_createMessage)
    ..pPM<Feature>(2, _omitFieldNames ? '' : 'features',
        subBuilder: Feature.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPlanEntitlementsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPlanEntitlementsResponse copyWith(
          void Function(GetPlanEntitlementsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetPlanEntitlementsResponse))
          as GetPlanEntitlementsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetPlanEntitlementsResponse() / GetPlanEntitlementsResponse.new instead')
  static GetPlanEntitlementsResponse create() =>
      GetPlanEntitlementsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetPlanEntitlementsResponse._();
  @$core.override
  GetPlanEntitlementsResponse createEmptyInstance() =>
      GetPlanEntitlementsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetPlanEntitlementsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetPlanEntitlementsResponse>(
          GetPlanEntitlementsResponse.$_createMessage);
  static GetPlanEntitlementsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<FeatureEntitlement> get entitlements => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<Feature> get features => $_getList(1);
}

/// Feature:可授權功能定義。
class Feature extends $pb.GeneratedMessage {
  factory Feature({
    $core.String? code,
    $core.String? type,
    $core.String? unit,
    $core.String? description,
  }) {
    final result = Feature._();
    if (code != null) result.code = code;
    if (type != null) result.type = type;
    if (unit != null) result.unit = unit;
    if (description != null) result.description = description;
    return result;
  }

  Feature._();

  factory Feature.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Feature()..mergeFromBuffer(data, registry);
  factory Feature.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Feature()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Feature',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: Feature.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'unit')
    ..aOS(4, _omitFieldNames ? '' : 'description')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Feature clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Feature copyWith(void Function(Feature) updates) =>
      super.copyWith((message) => updates(message as Feature)) as Feature;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Feature() / Feature.new instead')
  static Feature create() => Feature._();
  static $pb.GeneratedMessage $_createMessage() => Feature._();
  @$core.override
  Feature createEmptyInstance() => Feature._();
  @$core.pragma('dart2js:noInline')
  static Feature getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Feature>(Feature.$_createMessage);
  static Feature? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get code => $_getSZ(0);
  @$pb.TagNumber(1)
  set code($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get unit => $_getSZ(2);
  @$pb.TagNumber(3)
  set unit($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUnit() => $_has(2);
  @$pb.TagNumber(3)
  void clearUnit() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get description => $_getSZ(3);
  @$pb.TagNumber(4)
  set description($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDescription() => $_has(3);
  @$pb.TagNumber(4)
  void clearDescription() => $_clearField(4);
}

class FeatureEntitlement extends $pb.GeneratedMessage {
  factory FeatureEntitlement({
    $core.String? featureCode,
    $core.bool? enabled,
    $core.bool? limitSet,
    $fixnum.Int64? limitValue,
  }) {
    final result = FeatureEntitlement._();
    if (featureCode != null) result.featureCode = featureCode;
    if (enabled != null) result.enabled = enabled;
    if (limitSet != null) result.limitSet = limitSet;
    if (limitValue != null) result.limitValue = limitValue;
    return result;
  }

  FeatureEntitlement._();

  factory FeatureEntitlement.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FeatureEntitlement()..mergeFromBuffer(data, registry);
  factory FeatureEntitlement.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FeatureEntitlement()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FeatureEntitlement',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: FeatureEntitlement.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'featureCode')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..aOB(3, _omitFieldNames ? '' : 'limitSet')
    ..aInt64(4, _omitFieldNames ? '' : 'limitValue')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FeatureEntitlement clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FeatureEntitlement copyWith(void Function(FeatureEntitlement) updates) =>
      super.copyWith((message) => updates(message as FeatureEntitlement))
          as FeatureEntitlement;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use FeatureEntitlement() / FeatureEntitlement.new instead')
  static FeatureEntitlement create() => FeatureEntitlement._();
  static $pb.GeneratedMessage $_createMessage() => FeatureEntitlement._();
  @$core.override
  FeatureEntitlement createEmptyInstance() => FeatureEntitlement._();
  @$core.pragma('dart2js:noInline')
  static FeatureEntitlement getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<FeatureEntitlement>(
          FeatureEntitlement.$_createMessage);
  static FeatureEntitlement? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get featureCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set featureCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFeatureCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearFeatureCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get limitSet => $_getBF(2);
  @$pb.TagNumber(3)
  set limitSet($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLimitSet() => $_has(2);
  @$pb.TagNumber(3)
  void clearLimitSet() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get limitValue => $_getI64(3);
  @$pb.TagNumber(4)
  set limitValue($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLimitValue() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimitValue() => $_clearField(4);
}

class ListPlatformAuditRequest extends $pb.GeneratedMessage {
  factory ListPlatformAuditRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? targetType,
    $core.String? targetId,
  }) {
    final result = ListPlatformAuditRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (targetType != null) result.targetType = targetType;
    if (targetId != null) result.targetId = targetId;
    return result;
  }

  ListPlatformAuditRequest._();

  factory ListPlatformAuditRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlatformAuditRequest()..mergeFromBuffer(data, registry);
  factory ListPlatformAuditRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlatformAuditRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListPlatformAuditRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListPlatformAuditRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'targetType')
    ..aOS(4, _omitFieldNames ? '' : 'targetId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlatformAuditRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlatformAuditRequest copyWith(
          void Function(ListPlatformAuditRequest) updates) =>
      super.copyWith((message) => updates(message as ListPlatformAuditRequest))
          as ListPlatformAuditRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListPlatformAuditRequest() / ListPlatformAuditRequest.new instead')
  static ListPlatformAuditRequest create() => ListPlatformAuditRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListPlatformAuditRequest._();
  @$core.override
  ListPlatformAuditRequest createEmptyInstance() =>
      ListPlatformAuditRequest._();
  @$core.pragma('dart2js:noInline')
  static ListPlatformAuditRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListPlatformAuditRequest>(
          ListPlatformAuditRequest.$_createMessage);
  static ListPlatformAuditRequest? _defaultInstance;

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
  $core.String get targetType => $_getSZ(2);
  @$pb.TagNumber(3)
  set targetType($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTargetType() => $_has(2);
  @$pb.TagNumber(3)
  void clearTargetType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get targetId => $_getSZ(3);
  @$pb.TagNumber(4)
  set targetId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTargetId() => $_has(3);
  @$pb.TagNumber(4)
  void clearTargetId() => $_clearField(4);
}

class ListPlatformAuditResponse extends $pb.GeneratedMessage {
  factory ListPlatformAuditResponse({
    $core.Iterable<PlatformAuditEntry>? entries,
    PlatformPagination? pagination,
  }) {
    final result = ListPlatformAuditResponse._();
    if (entries != null) result.entries.addAll(entries);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListPlatformAuditResponse._();

  factory ListPlatformAuditResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlatformAuditResponse()..mergeFromBuffer(data, registry);
  factory ListPlatformAuditResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListPlatformAuditResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListPlatformAuditResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListPlatformAuditResponse.$_createMessage)
    ..pPM<PlatformAuditEntry>(1, _omitFieldNames ? '' : 'entries',
        subBuilder: PlatformAuditEntry.$_createMessage)
    ..aOM<PlatformPagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: PlatformPagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlatformAuditResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListPlatformAuditResponse copyWith(
          void Function(ListPlatformAuditResponse) updates) =>
      super.copyWith((message) => updates(message as ListPlatformAuditResponse))
          as ListPlatformAuditResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListPlatformAuditResponse() / ListPlatformAuditResponse.new instead')
  static ListPlatformAuditResponse create() => ListPlatformAuditResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListPlatformAuditResponse._();
  @$core.override
  ListPlatformAuditResponse createEmptyInstance() =>
      ListPlatformAuditResponse._();
  @$core.pragma('dart2js:noInline')
  static ListPlatformAuditResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListPlatformAuditResponse>(
          ListPlatformAuditResponse.$_createMessage);
  static ListPlatformAuditResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<PlatformAuditEntry> get entries => $_getList(0);

  @$pb.TagNumber(2)
  PlatformPagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination(PlatformPagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  PlatformPagination ensurePagination() => $_ensure(1);
}

class PlatformAuditEntry extends $pb.GeneratedMessage {
  factory PlatformAuditEntry({
    $core.String? id,
    $core.String? operatorEmail,
    $core.String? action,
    $core.String? targetType,
    $core.String? targetId,
    $core.String? reason,
    $core.String? createdAt,
  }) {
    final result = PlatformAuditEntry._();
    if (id != null) result.id = id;
    if (operatorEmail != null) result.operatorEmail = operatorEmail;
    if (action != null) result.action = action;
    if (targetType != null) result.targetType = targetType;
    if (targetId != null) result.targetId = targetId;
    if (reason != null) result.reason = reason;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  PlatformAuditEntry._();

  factory PlatformAuditEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PlatformAuditEntry()..mergeFromBuffer(data, registry);
  factory PlatformAuditEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PlatformAuditEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PlatformAuditEntry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: PlatformAuditEntry.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'operatorEmail')
    ..aOS(3, _omitFieldNames ? '' : 'action')
    ..aOS(4, _omitFieldNames ? '' : 'targetType')
    ..aOS(5, _omitFieldNames ? '' : 'targetId')
    ..aOS(6, _omitFieldNames ? '' : 'reason')
    ..aOS(7, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlatformAuditEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlatformAuditEntry copyWith(void Function(PlatformAuditEntry) updates) =>
      super.copyWith((message) => updates(message as PlatformAuditEntry))
          as PlatformAuditEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PlatformAuditEntry() / PlatformAuditEntry.new instead')
  static PlatformAuditEntry create() => PlatformAuditEntry._();
  static $pb.GeneratedMessage $_createMessage() => PlatformAuditEntry._();
  @$core.override
  PlatformAuditEntry createEmptyInstance() => PlatformAuditEntry._();
  @$core.pragma('dart2js:noInline')
  static PlatformAuditEntry getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PlatformAuditEntry>(
          PlatformAuditEntry.$_createMessage);
  static PlatformAuditEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get operatorEmail => $_getSZ(1);
  @$pb.TagNumber(2)
  set operatorEmail($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasOperatorEmail() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperatorEmail() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get action => $_getSZ(2);
  @$pb.TagNumber(3)
  set action($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAction() => $_has(2);
  @$pb.TagNumber(3)
  void clearAction() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get targetType => $_getSZ(3);
  @$pb.TagNumber(4)
  set targetType($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTargetType() => $_has(3);
  @$pb.TagNumber(4)
  void clearTargetType() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get targetId => $_getSZ(4);
  @$pb.TagNumber(5)
  set targetId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTargetId() => $_has(4);
  @$pb.TagNumber(5)
  void clearTargetId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get reason => $_getSZ(5);
  @$pb.TagNumber(6)
  set reason($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasReason() => $_has(5);
  @$pb.TagNumber(6)
  void clearReason() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get createdAt => $_getSZ(6);
  @$pb.TagNumber(7)
  set createdAt($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCreatedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearCreatedAt() => $_clearField(7);
}

/// GetTenantEntitlementsRequest:租戶端唯讀投影(自己的公司;前端據此 disable 按鈕與顯示用量)。
class GetTenantEntitlementsRequest extends $pb.GeneratedMessage {
  factory GetTenantEntitlementsRequest() => GetTenantEntitlementsRequest._();

  GetTenantEntitlementsRequest._();

  factory GetTenantEntitlementsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantEntitlementsRequest()..mergeFromBuffer(data, registry);
  factory GetTenantEntitlementsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantEntitlementsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTenantEntitlementsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetTenantEntitlementsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantEntitlementsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantEntitlementsRequest copyWith(
          void Function(GetTenantEntitlementsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetTenantEntitlementsRequest))
          as GetTenantEntitlementsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTenantEntitlementsRequest() / GetTenantEntitlementsRequest.new instead')
  static GetTenantEntitlementsRequest create() =>
      GetTenantEntitlementsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTenantEntitlementsRequest._();
  @$core.override
  GetTenantEntitlementsRequest createEmptyInstance() =>
      GetTenantEntitlementsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTenantEntitlementsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTenantEntitlementsRequest>(
          GetTenantEntitlementsRequest.$_createMessage);
  static GetTenantEntitlementsRequest? _defaultInstance;
}

class GetTenantEntitlementsResponse extends $pb.GeneratedMessage {
  factory GetTenantEntitlementsResponse({
    $core.String? planCode,
    $core.String? planName,
    $core.String? status,
    $core.String? trialEndsAt,
    $core.Iterable<Usage>? usage,
  }) {
    final result = GetTenantEntitlementsResponse._();
    if (planCode != null) result.planCode = planCode;
    if (planName != null) result.planName = planName;
    if (status != null) result.status = status;
    if (trialEndsAt != null) result.trialEndsAt = trialEndsAt;
    if (usage != null) result.usage.addAll(usage);
    return result;
  }

  GetTenantEntitlementsResponse._();

  factory GetTenantEntitlementsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantEntitlementsResponse()..mergeFromBuffer(data, registry);
  factory GetTenantEntitlementsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTenantEntitlementsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTenantEntitlementsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetTenantEntitlementsResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'planCode')
    ..aOS(2, _omitFieldNames ? '' : 'planName')
    ..aOS(3, _omitFieldNames ? '' : 'status')
    ..aOS(4, _omitFieldNames ? '' : 'trialEndsAt')
    ..pPM<Usage>(5, _omitFieldNames ? '' : 'usage',
        subBuilder: Usage.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantEntitlementsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTenantEntitlementsResponse copyWith(
          void Function(GetTenantEntitlementsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetTenantEntitlementsResponse))
          as GetTenantEntitlementsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTenantEntitlementsResponse() / GetTenantEntitlementsResponse.new instead')
  static GetTenantEntitlementsResponse create() =>
      GetTenantEntitlementsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTenantEntitlementsResponse._();
  @$core.override
  GetTenantEntitlementsResponse createEmptyInstance() =>
      GetTenantEntitlementsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTenantEntitlementsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTenantEntitlementsResponse>(
          GetTenantEntitlementsResponse.$_createMessage);
  static GetTenantEntitlementsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get planCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set planCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlanCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlanCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get planName => $_getSZ(1);
  @$pb.TagNumber(2)
  set planName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPlanName() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlanName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get status => $_getSZ(2);
  @$pb.TagNumber(3)
  set status($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get trialEndsAt => $_getSZ(3);
  @$pb.TagNumber(4)
  set trialEndsAt($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTrialEndsAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearTrialEndsAt() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<Usage> get usage => $_getList(4);
}

/// Usage:單一功能的使用量投影。
class Usage extends $pb.GeneratedMessage {
  factory Usage({
    $core.String? featureCode,
    $core.bool? enabled,
    $core.bool? limitSet,
    $fixnum.Int64? limitValue,
    $fixnum.Int64? used,
  }) {
    final result = Usage._();
    if (featureCode != null) result.featureCode = featureCode;
    if (enabled != null) result.enabled = enabled;
    if (limitSet != null) result.limitSet = limitSet;
    if (limitValue != null) result.limitValue = limitValue;
    if (used != null) result.used = used;
    return result;
  }

  Usage._();

  factory Usage.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Usage()..mergeFromBuffer(data, registry);
  factory Usage.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Usage()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Usage',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: Usage.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'featureCode')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..aOB(3, _omitFieldNames ? '' : 'limitSet')
    ..aInt64(4, _omitFieldNames ? '' : 'limitValue')
    ..aInt64(5, _omitFieldNames ? '' : 'used')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Usage clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Usage copyWith(void Function(Usage) updates) =>
      super.copyWith((message) => updates(message as Usage)) as Usage;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Usage() / Usage.new instead')
  static Usage create() => Usage._();
  static $pb.GeneratedMessage $_createMessage() => Usage._();
  @$core.override
  Usage createEmptyInstance() => Usage._();
  @$core.pragma('dart2js:noInline')
  static Usage getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Usage>(Usage.$_createMessage);
  static Usage? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get featureCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set featureCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFeatureCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearFeatureCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get limitSet => $_getBF(2);
  @$pb.TagNumber(3)
  set limitSet($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLimitSet() => $_has(2);
  @$pb.TagNumber(3)
  void clearLimitSet() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get limitValue => $_getI64(3);
  @$pb.TagNumber(4)
  set limitValue($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLimitValue() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimitValue() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get used => $_getI64(4);
  @$pb.TagNumber(5)
  set used($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUsed() => $_has(4);
  @$pb.TagNumber(5)
  void clearUsed() => $_clearField(5);
}

/// PlatformAdminService:平台營運(operator session;租戶身分一律拒絕)。
class PlatformAdminServiceApi {
  final $pb.RpcClient _client;

  PlatformAdminServiceApi(this._client);

  $async.Future<ListTenantsResponse> listTenants(
          $pb.ClientContext? ctx, ListTenantsRequest request) =>
      _client.invoke<ListTenantsResponse>(ctx, 'PlatformAdminService',
          'ListTenants', request, ListTenantsResponse());
  $async.Future<GetTenantResponse> getTenant(
          $pb.ClientContext? ctx, GetTenantRequest request) =>
      _client.invoke<GetTenantResponse>(ctx, 'PlatformAdminService',
          'GetTenant', request, GetTenantResponse());
  $async.Future<ListPlansResponse> listPlans(
          $pb.ClientContext? ctx, ListPlansRequest request) =>
      _client.invoke<ListPlansResponse>(ctx, 'PlatformAdminService',
          'ListPlans', request, ListPlansResponse());
  $async.Future<GetPlanEntitlementsResponse> getPlanEntitlements(
          $pb.ClientContext? ctx, GetPlanEntitlementsRequest request) =>
      _client.invoke<GetPlanEntitlementsResponse>(ctx, 'PlatformAdminService',
          'GetPlanEntitlements', request, GetPlanEntitlementsResponse());
  $async.Future<ListPlatformAuditResponse> listPlatformAudit(
          $pb.ClientContext? ctx, ListPlatformAuditRequest request) =>
      _client.invoke<ListPlatformAuditResponse>(ctx, 'PlatformAdminService',
          'ListPlatformAudit', request, ListPlatformAuditResponse());
}

/// TenantEntitlementService:租戶端權益投影(租戶 session;唯讀)。
class TenantEntitlementServiceApi {
  final $pb.RpcClient _client;

  TenantEntitlementServiceApi(this._client);

  $async.Future<GetTenantEntitlementsResponse> getTenantEntitlements(
          $pb.ClientContext? ctx, GetTenantEntitlementsRequest request) =>
      _client.invoke<GetTenantEntitlementsResponse>(
          ctx,
          'TenantEntitlementService',
          'GetTenantEntitlements',
          request,
          GetTenantEntitlementsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
