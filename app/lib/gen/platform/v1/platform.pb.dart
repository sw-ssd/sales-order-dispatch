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
    $core.String? trialEndsAt,
    $core.String? graceUntil,
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
    if (trialEndsAt != null) result.trialEndsAt = trialEndsAt;
    if (graceUntil != null) result.graceUntil = graceUntil;
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
    ..aOS(10, _omitFieldNames ? '' : 'trialEndsAt')
    ..aOS(11, _omitFieldNames ? '' : 'graceUntil')
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

  @$pb.TagNumber(10)
  $core.String get trialEndsAt => $_getSZ(8);
  @$pb.TagNumber(10)
  set trialEndsAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(10)
  $core.bool hasTrialEndsAt() => $_has(8);
  @$pb.TagNumber(10)
  void clearTrialEndsAt() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get graceUntil => $_getSZ(9);
  @$pb.TagNumber(11)
  set graceUntil($core.String value) => $_setString(9, value);
  @$pb.TagNumber(11)
  $core.bool hasGraceUntil() => $_has(9);
  @$pb.TagNumber(11)
  void clearGraceUntil() => $_clearField(11);
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

class ListReceivablesRequest extends $pb.GeneratedMessage {
  factory ListReceivablesRequest({
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListReceivablesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListReceivablesRequest._();

  factory ListReceivablesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReceivablesRequest()..mergeFromBuffer(data, registry);
  factory ListReceivablesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReceivablesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListReceivablesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListReceivablesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReceivablesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReceivablesRequest copyWith(
          void Function(ListReceivablesRequest) updates) =>
      super.copyWith((message) => updates(message as ListReceivablesRequest))
          as ListReceivablesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListReceivablesRequest() / ListReceivablesRequest.new instead')
  static ListReceivablesRequest create() => ListReceivablesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListReceivablesRequest._();
  @$core.override
  ListReceivablesRequest createEmptyInstance() => ListReceivablesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListReceivablesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListReceivablesRequest>(
          ListReceivablesRequest.$_createMessage);
  static ListReceivablesRequest? _defaultInstance;

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

class ListReceivablesResponse extends $pb.GeneratedMessage {
  factory ListReceivablesResponse({
    $core.Iterable<Receivable>? rows,
    PlatformPagination? pagination,
  }) {
    final result = ListReceivablesResponse._();
    if (rows != null) result.rows.addAll(rows);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListReceivablesResponse._();

  factory ListReceivablesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReceivablesResponse()..mergeFromBuffer(data, registry);
  factory ListReceivablesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReceivablesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListReceivablesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListReceivablesResponse.$_createMessage)
    ..pPM<Receivable>(1, _omitFieldNames ? '' : 'rows',
        subBuilder: Receivable.$_createMessage)
    ..aOM<PlatformPagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: PlatformPagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReceivablesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReceivablesResponse copyWith(
          void Function(ListReceivablesResponse) updates) =>
      super.copyWith((message) => updates(message as ListReceivablesResponse))
          as ListReceivablesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListReceivablesResponse() / ListReceivablesResponse.new instead')
  static ListReceivablesResponse create() => ListReceivablesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListReceivablesResponse._();
  @$core.override
  ListReceivablesResponse createEmptyInstance() => ListReceivablesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListReceivablesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListReceivablesResponse>(
          ListReceivablesResponse.$_createMessage);
  static ListReceivablesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Receivable> get rows => $_getList(0);

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

/// Receivable:一期未付的帳(供 console 顯示與匯出 CSV)。G5 的平台自營公司不算租戶,不列入。
class Receivable extends $pb.GeneratedMessage {
  factory Receivable({
    $core.String? companyId,
    $core.String? companyName,
    $core.String? planCode,
    $core.int? periodNo,
    $core.String? amount,
    $core.String? periodEnd,
    $core.String? status,
  }) {
    final result = Receivable._();
    if (companyId != null) result.companyId = companyId;
    if (companyName != null) result.companyName = companyName;
    if (planCode != null) result.planCode = planCode;
    if (periodNo != null) result.periodNo = periodNo;
    if (amount != null) result.amount = amount;
    if (periodEnd != null) result.periodEnd = periodEnd;
    if (status != null) result.status = status;
    return result;
  }

  Receivable._();

  factory Receivable.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Receivable()..mergeFromBuffer(data, registry);
  factory Receivable.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Receivable()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Receivable',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: Receivable.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'companyName')
    ..aOS(3, _omitFieldNames ? '' : 'planCode')
    ..aI(4, _omitFieldNames ? '' : 'periodNo')
    ..aOS(5, _omitFieldNames ? '' : 'amount')
    ..aOS(6, _omitFieldNames ? '' : 'periodEnd')
    ..aOS(7, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Receivable clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Receivable copyWith(void Function(Receivable) updates) =>
      super.copyWith((message) => updates(message as Receivable)) as Receivable;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Receivable() / Receivable.new instead')
  static Receivable create() => Receivable._();
  static $pb.GeneratedMessage $_createMessage() => Receivable._();
  @$core.override
  Receivable createEmptyInstance() => Receivable._();
  @$core.pragma('dart2js:noInline')
  static Receivable getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Receivable>(Receivable.$_createMessage);
  static Receivable? _defaultInstance;

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
  $core.int get periodNo => $_getIZ(3);
  @$pb.TagNumber(4)
  set periodNo($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPeriodNo() => $_has(3);
  @$pb.TagNumber(4)
  void clearPeriodNo() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get amount => $_getSZ(4);
  @$pb.TagNumber(5)
  set amount($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmount() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get periodEnd => $_getSZ(5);
  @$pb.TagNumber(6)
  set periodEnd($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPeriodEnd() => $_has(5);
  @$pb.TagNumber(6)
  void clearPeriodEnd() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get status => $_getSZ(6);
  @$pb.TagNumber(7)
  set status($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasStatus() => $_has(6);
  @$pb.TagNumber(7)
  void clearStatus() => $_clearField(7);
}

class RecordPaymentRequest extends $pb.GeneratedMessage {
  factory RecordPaymentRequest({
    $core.String? companyId,
    $core.int? periodNo,
    $core.String? amount,
    $core.String? provider,
    $core.String? externalRef,
    $core.String? invoiceNo,
    $core.String? invoiceStatus,
    $core.String? buyerTaxId,
    $core.String? carrier,
    $core.String? note,
    $core.String? reason,
  }) {
    final result = RecordPaymentRequest._();
    if (companyId != null) result.companyId = companyId;
    if (periodNo != null) result.periodNo = periodNo;
    if (amount != null) result.amount = amount;
    if (provider != null) result.provider = provider;
    if (externalRef != null) result.externalRef = externalRef;
    if (invoiceNo != null) result.invoiceNo = invoiceNo;
    if (invoiceStatus != null) result.invoiceStatus = invoiceStatus;
    if (buyerTaxId != null) result.buyerTaxId = buyerTaxId;
    if (carrier != null) result.carrier = carrier;
    if (note != null) result.note = note;
    if (reason != null) result.reason = reason;
    return result;
  }

  RecordPaymentRequest._();

  factory RecordPaymentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentRequest()..mergeFromBuffer(data, registry);
  factory RecordPaymentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecordPaymentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: RecordPaymentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aI(2, _omitFieldNames ? '' : 'periodNo')
    ..aOS(3, _omitFieldNames ? '' : 'amount')
    ..aOS(4, _omitFieldNames ? '' : 'provider')
    ..aOS(5, _omitFieldNames ? '' : 'externalRef')
    ..aOS(6, _omitFieldNames ? '' : 'invoiceNo')
    ..aOS(7, _omitFieldNames ? '' : 'invoiceStatus')
    ..aOS(8, _omitFieldNames ? '' : 'buyerTaxId')
    ..aOS(9, _omitFieldNames ? '' : 'carrier')
    ..aOS(10, _omitFieldNames ? '' : 'note')
    ..aOS(11, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentRequest copyWith(void Function(RecordPaymentRequest) updates) =>
      super.copyWith((message) => updates(message as RecordPaymentRequest))
          as RecordPaymentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RecordPaymentRequest() / RecordPaymentRequest.new instead')
  static RecordPaymentRequest create() => RecordPaymentRequest._();
  static $pb.GeneratedMessage $_createMessage() => RecordPaymentRequest._();
  @$core.override
  RecordPaymentRequest createEmptyInstance() => RecordPaymentRequest._();
  @$core.pragma('dart2js:noInline')
  static RecordPaymentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecordPaymentRequest>(
          RecordPaymentRequest.$_createMessage);
  static RecordPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get periodNo => $_getIZ(1);
  @$pb.TagNumber(2)
  set periodNo($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPeriodNo() => $_has(1);
  @$pb.TagNumber(2)
  void clearPeriodNo() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get amount => $_getSZ(2);
  @$pb.TagNumber(3)
  set amount($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAmount() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get provider => $_getSZ(3);
  @$pb.TagNumber(4)
  set provider($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasProvider() => $_has(3);
  @$pb.TagNumber(4)
  void clearProvider() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get externalRef => $_getSZ(4);
  @$pb.TagNumber(5)
  set externalRef($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasExternalRef() => $_has(4);
  @$pb.TagNumber(5)
  void clearExternalRef() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get invoiceNo => $_getSZ(5);
  @$pb.TagNumber(6)
  set invoiceNo($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasInvoiceNo() => $_has(5);
  @$pb.TagNumber(6)
  void clearInvoiceNo() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get invoiceStatus => $_getSZ(6);
  @$pb.TagNumber(7)
  set invoiceStatus($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInvoiceStatus() => $_has(6);
  @$pb.TagNumber(7)
  void clearInvoiceStatus() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get buyerTaxId => $_getSZ(7);
  @$pb.TagNumber(8)
  set buyerTaxId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasBuyerTaxId() => $_has(7);
  @$pb.TagNumber(8)
  void clearBuyerTaxId() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get carrier => $_getSZ(8);
  @$pb.TagNumber(9)
  set carrier($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCarrier() => $_has(8);
  @$pb.TagNumber(9)
  void clearCarrier() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get note => $_getSZ(9);
  @$pb.TagNumber(10)
  set note($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasNote() => $_has(9);
  @$pb.TagNumber(10)
  void clearNote() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get reason => $_getSZ(10);
  @$pb.TagNumber(11)
  set reason($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasReason() => $_has(10);
  @$pb.TagNumber(11)
  void clearReason() => $_clearField(11);
}

class RecordPaymentResponse extends $pb.GeneratedMessage {
  factory RecordPaymentResponse({
    $core.int? periodNo,
    $core.String? status,
  }) {
    final result = RecordPaymentResponse._();
    if (periodNo != null) result.periodNo = periodNo;
    if (status != null) result.status = status;
    return result;
  }

  RecordPaymentResponse._();

  factory RecordPaymentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentResponse()..mergeFromBuffer(data, registry);
  factory RecordPaymentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RecordPaymentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecordPaymentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: RecordPaymentResponse.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'periodNo')
    ..aOS(2, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecordPaymentResponse copyWith(
          void Function(RecordPaymentResponse) updates) =>
      super.copyWith((message) => updates(message as RecordPaymentResponse))
          as RecordPaymentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RecordPaymentResponse() / RecordPaymentResponse.new instead')
  static RecordPaymentResponse create() => RecordPaymentResponse._();
  static $pb.GeneratedMessage $_createMessage() => RecordPaymentResponse._();
  @$core.override
  RecordPaymentResponse createEmptyInstance() => RecordPaymentResponse._();
  @$core.pragma('dart2js:noInline')
  static RecordPaymentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecordPaymentResponse>(
          RecordPaymentResponse.$_createMessage);
  static RecordPaymentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get periodNo => $_getIZ(0);
  @$pb.TagNumber(1)
  set periodNo($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPeriodNo() => $_has(0);
  @$pb.TagNumber(1)
  void clearPeriodNo() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get status => $_getSZ(1);
  @$pb.TagNumber(2)
  set status($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);
}

class CreateSubscriptionRequest extends $pb.GeneratedMessage {
  factory CreateSubscriptionRequest({
    $core.String? companyId,
    $core.String? planCode,
    $core.String? billingCycle,
    $core.int? seatCount,
    $core.String? trialEndsAt,
    $core.String? reason,
  }) {
    final result = CreateSubscriptionRequest._();
    if (companyId != null) result.companyId = companyId;
    if (planCode != null) result.planCode = planCode;
    if (billingCycle != null) result.billingCycle = billingCycle;
    if (seatCount != null) result.seatCount = seatCount;
    if (trialEndsAt != null) result.trialEndsAt = trialEndsAt;
    if (reason != null) result.reason = reason;
    return result;
  }

  CreateSubscriptionRequest._();

  factory CreateSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory CreateSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: CreateSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'planCode')
    ..aOS(3, _omitFieldNames ? '' : 'billingCycle')
    ..aI(4, _omitFieldNames ? '' : 'seatCount')
    ..aOS(5, _omitFieldNames ? '' : 'trialEndsAt')
    ..aOS(6, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionRequest copyWith(
          void Function(CreateSubscriptionRequest) updates) =>
      super.copyWith((message) => updates(message as CreateSubscriptionRequest))
          as CreateSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSubscriptionRequest() / CreateSubscriptionRequest.new instead')
  static CreateSubscriptionRequest create() => CreateSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateSubscriptionRequest._();
  @$core.override
  CreateSubscriptionRequest createEmptyInstance() =>
      CreateSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSubscriptionRequest>(
          CreateSubscriptionRequest.$_createMessage);
  static CreateSubscriptionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get planCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set planCode($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPlanCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlanCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get billingCycle => $_getSZ(2);
  @$pb.TagNumber(3)
  set billingCycle($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBillingCycle() => $_has(2);
  @$pb.TagNumber(3)
  void clearBillingCycle() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get seatCount => $_getIZ(3);
  @$pb.TagNumber(4)
  set seatCount($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSeatCount() => $_has(3);
  @$pb.TagNumber(4)
  void clearSeatCount() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get trialEndsAt => $_getSZ(4);
  @$pb.TagNumber(5)
  set trialEndsAt($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTrialEndsAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearTrialEndsAt() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get reason => $_getSZ(5);
  @$pb.TagNumber(6)
  set reason($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasReason() => $_has(5);
  @$pb.TagNumber(6)
  void clearReason() => $_clearField(6);
}

class CreateSubscriptionResponse extends $pb.GeneratedMessage {
  factory CreateSubscriptionResponse({
    $core.String? subscriptionId,
    $core.String? status,
    $core.String? planCode,
    $core.String? billingCycle,
    $core.int? seatCount,
    $core.String? trialEndsAt,
    $core.int? firstPeriodNo,
    $core.String? firstPeriodEnd,
    $core.String? firstPeriodAmount,
  }) {
    final result = CreateSubscriptionResponse._();
    if (subscriptionId != null) result.subscriptionId = subscriptionId;
    if (status != null) result.status = status;
    if (planCode != null) result.planCode = planCode;
    if (billingCycle != null) result.billingCycle = billingCycle;
    if (seatCount != null) result.seatCount = seatCount;
    if (trialEndsAt != null) result.trialEndsAt = trialEndsAt;
    if (firstPeriodNo != null) result.firstPeriodNo = firstPeriodNo;
    if (firstPeriodEnd != null) result.firstPeriodEnd = firstPeriodEnd;
    if (firstPeriodAmount != null) result.firstPeriodAmount = firstPeriodAmount;
    return result;
  }

  CreateSubscriptionResponse._();

  factory CreateSubscriptionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory CreateSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: CreateSubscriptionResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'subscriptionId')
    ..aOS(2, _omitFieldNames ? '' : 'status')
    ..aOS(3, _omitFieldNames ? '' : 'planCode')
    ..aOS(4, _omitFieldNames ? '' : 'billingCycle')
    ..aI(5, _omitFieldNames ? '' : 'seatCount')
    ..aOS(6, _omitFieldNames ? '' : 'trialEndsAt')
    ..aI(7, _omitFieldNames ? '' : 'firstPeriodNo')
    ..aOS(8, _omitFieldNames ? '' : 'firstPeriodEnd')
    ..aOS(9, _omitFieldNames ? '' : 'firstPeriodAmount')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSubscriptionResponse copyWith(
          void Function(CreateSubscriptionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateSubscriptionResponse))
          as CreateSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSubscriptionResponse() / CreateSubscriptionResponse.new instead')
  static CreateSubscriptionResponse create() => CreateSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateSubscriptionResponse._();
  @$core.override
  CreateSubscriptionResponse createEmptyInstance() =>
      CreateSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSubscriptionResponse>(
          CreateSubscriptionResponse.$_createMessage);
  static CreateSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get subscriptionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set subscriptionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSubscriptionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubscriptionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get status => $_getSZ(1);
  @$pb.TagNumber(2)
  set status($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get planCode => $_getSZ(2);
  @$pb.TagNumber(3)
  set planCode($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPlanCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlanCode() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get billingCycle => $_getSZ(3);
  @$pb.TagNumber(4)
  set billingCycle($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBillingCycle() => $_has(3);
  @$pb.TagNumber(4)
  void clearBillingCycle() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get seatCount => $_getIZ(4);
  @$pb.TagNumber(5)
  set seatCount($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSeatCount() => $_has(4);
  @$pb.TagNumber(5)
  void clearSeatCount() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get trialEndsAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set trialEndsAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTrialEndsAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearTrialEndsAt() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.int get firstPeriodNo => $_getIZ(6);
  @$pb.TagNumber(7)
  set firstPeriodNo($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasFirstPeriodNo() => $_has(6);
  @$pb.TagNumber(7)
  void clearFirstPeriodNo() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get firstPeriodEnd => $_getSZ(7);
  @$pb.TagNumber(8)
  set firstPeriodEnd($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasFirstPeriodEnd() => $_has(7);
  @$pb.TagNumber(8)
  void clearFirstPeriodEnd() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get firstPeriodAmount => $_getSZ(8);
  @$pb.TagNumber(9)
  set firstPeriodAmount($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasFirstPeriodAmount() => $_has(8);
  @$pb.TagNumber(9)
  void clearFirstPeriodAmount() => $_clearField(9);
}

class SetSeatCountRequest extends $pb.GeneratedMessage {
  factory SetSeatCountRequest({
    $core.String? companyId,
    $core.int? seatCount,
    $core.String? reason,
  }) {
    final result = SetSeatCountRequest._();
    if (companyId != null) result.companyId = companyId;
    if (seatCount != null) result.seatCount = seatCount;
    if (reason != null) result.reason = reason;
    return result;
  }

  SetSeatCountRequest._();

  factory SetSeatCountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetSeatCountRequest()..mergeFromBuffer(data, registry);
  factory SetSeatCountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetSeatCountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetSeatCountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SetSeatCountRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aI(2, _omitFieldNames ? '' : 'seatCount')
    ..aOS(3, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetSeatCountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetSeatCountRequest copyWith(void Function(SetSeatCountRequest) updates) =>
      super.copyWith((message) => updates(message as SetSeatCountRequest))
          as SetSeatCountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use SetSeatCountRequest() / SetSeatCountRequest.new instead')
  static SetSeatCountRequest create() => SetSeatCountRequest._();
  static $pb.GeneratedMessage $_createMessage() => SetSeatCountRequest._();
  @$core.override
  SetSeatCountRequest createEmptyInstance() => SetSeatCountRequest._();
  @$core.pragma('dart2js:noInline')
  static SetSeatCountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetSeatCountRequest>(
          SetSeatCountRequest.$_createMessage);
  static SetSeatCountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get seatCount => $_getIZ(1);
  @$pb.TagNumber(2)
  set seatCount($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSeatCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearSeatCount() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get reason => $_getSZ(2);
  @$pb.TagNumber(3)
  set reason($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasReason() => $_has(2);
  @$pb.TagNumber(3)
  void clearReason() => $_clearField(3);
}

class SetSeatCountResponse extends $pb.GeneratedMessage {
  factory SetSeatCountResponse({
    $core.int? seatCount,
  }) {
    final result = SetSeatCountResponse._();
    if (seatCount != null) result.seatCount = seatCount;
    return result;
  }

  SetSeatCountResponse._();

  factory SetSeatCountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetSeatCountResponse()..mergeFromBuffer(data, registry);
  factory SetSeatCountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetSeatCountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetSeatCountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SetSeatCountResponse.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'seatCount')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetSeatCountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetSeatCountResponse copyWith(void Function(SetSeatCountResponse) updates) =>
      super.copyWith((message) => updates(message as SetSeatCountResponse))
          as SetSeatCountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetSeatCountResponse() / SetSeatCountResponse.new instead')
  static SetSeatCountResponse create() => SetSeatCountResponse._();
  static $pb.GeneratedMessage $_createMessage() => SetSeatCountResponse._();
  @$core.override
  SetSeatCountResponse createEmptyInstance() => SetSeatCountResponse._();
  @$core.pragma('dart2js:noInline')
  static SetSeatCountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetSeatCountResponse>(
          SetSeatCountResponse.$_createMessage);
  static SetSeatCountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get seatCount => $_getIZ(0);
  @$pb.TagNumber(1)
  set seatCount($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSeatCount() => $_has(0);
  @$pb.TagNumber(1)
  void clearSeatCount() => $_clearField(1);
}

class ChangePlanRequest extends $pb.GeneratedMessage {
  factory ChangePlanRequest({
    $core.String? companyId,
    $core.String? planCode,
    $core.String? reason,
  }) {
    final result = ChangePlanRequest._();
    if (companyId != null) result.companyId = companyId;
    if (planCode != null) result.planCode = planCode;
    if (reason != null) result.reason = reason;
    return result;
  }

  ChangePlanRequest._();

  factory ChangePlanRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePlanRequest()..mergeFromBuffer(data, registry);
  factory ChangePlanRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePlanRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangePlanRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ChangePlanRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'planCode')
    ..aOS(3, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePlanRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePlanRequest copyWith(void Function(ChangePlanRequest) updates) =>
      super.copyWith((message) => updates(message as ChangePlanRequest))
          as ChangePlanRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChangePlanRequest() / ChangePlanRequest.new instead')
  static ChangePlanRequest create() => ChangePlanRequest._();
  static $pb.GeneratedMessage $_createMessage() => ChangePlanRequest._();
  @$core.override
  ChangePlanRequest createEmptyInstance() => ChangePlanRequest._();
  @$core.pragma('dart2js:noInline')
  static ChangePlanRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ChangePlanRequest>(
          ChangePlanRequest.$_createMessage);
  static ChangePlanRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get planCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set planCode($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPlanCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlanCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get reason => $_getSZ(2);
  @$pb.TagNumber(3)
  set reason($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasReason() => $_has(2);
  @$pb.TagNumber(3)
  void clearReason() => $_clearField(3);
}

class ChangePlanResponse extends $pb.GeneratedMessage {
  factory ChangePlanResponse({
    $core.String? planCode,
    $core.String? effectiveFrom,
  }) {
    final result = ChangePlanResponse._();
    if (planCode != null) result.planCode = planCode;
    if (effectiveFrom != null) result.effectiveFrom = effectiveFrom;
    return result;
  }

  ChangePlanResponse._();

  factory ChangePlanResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePlanResponse()..mergeFromBuffer(data, registry);
  factory ChangePlanResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePlanResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangePlanResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ChangePlanResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'planCode')
    ..aOS(2, _omitFieldNames ? '' : 'effectiveFrom')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePlanResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePlanResponse copyWith(void Function(ChangePlanResponse) updates) =>
      super.copyWith((message) => updates(message as ChangePlanResponse))
          as ChangePlanResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChangePlanResponse() / ChangePlanResponse.new instead')
  static ChangePlanResponse create() => ChangePlanResponse._();
  static $pb.GeneratedMessage $_createMessage() => ChangePlanResponse._();
  @$core.override
  ChangePlanResponse createEmptyInstance() => ChangePlanResponse._();
  @$core.pragma('dart2js:noInline')
  static ChangePlanResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangePlanResponse>(
          ChangePlanResponse.$_createMessage);
  static ChangePlanResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get planCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set planCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlanCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlanCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get effectiveFrom => $_getSZ(1);
  @$pb.TagNumber(2)
  set effectiveFrom($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEffectiveFrom() => $_has(1);
  @$pb.TagNumber(2)
  void clearEffectiveFrom() => $_clearField(2);
}

class CancelSubscriptionRequest extends $pb.GeneratedMessage {
  factory CancelSubscriptionRequest({
    $core.String? companyId,
    $core.bool? atPeriodEnd,
    $core.String? reason,
  }) {
    final result = CancelSubscriptionRequest._();
    if (companyId != null) result.companyId = companyId;
    if (atPeriodEnd != null) result.atPeriodEnd = atPeriodEnd;
    if (reason != null) result.reason = reason;
    return result;
  }

  CancelSubscriptionRequest._();

  factory CancelSubscriptionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionRequest()..mergeFromBuffer(data, registry);
  factory CancelSubscriptionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelSubscriptionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: CancelSubscriptionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOB(2, _omitFieldNames ? '' : 'atPeriodEnd')
    ..aOS(3, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionRequest copyWith(
          void Function(CancelSubscriptionRequest) updates) =>
      super.copyWith((message) => updates(message as CancelSubscriptionRequest))
          as CancelSubscriptionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelSubscriptionRequest() / CancelSubscriptionRequest.new instead')
  static CancelSubscriptionRequest create() => CancelSubscriptionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CancelSubscriptionRequest._();
  @$core.override
  CancelSubscriptionRequest createEmptyInstance() =>
      CancelSubscriptionRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelSubscriptionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelSubscriptionRequest>(
          CancelSubscriptionRequest.$_createMessage);
  static CancelSubscriptionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get atPeriodEnd => $_getBF(1);
  @$pb.TagNumber(2)
  set atPeriodEnd($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAtPeriodEnd() => $_has(1);
  @$pb.TagNumber(2)
  void clearAtPeriodEnd() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get reason => $_getSZ(2);
  @$pb.TagNumber(3)
  set reason($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasReason() => $_has(2);
  @$pb.TagNumber(3)
  void clearReason() => $_clearField(3);
}

class CancelSubscriptionResponse extends $pb.GeneratedMessage {
  factory CancelSubscriptionResponse({
    $core.String? cancelledAt,
    $core.String? serviceUntil,
  }) {
    final result = CancelSubscriptionResponse._();
    if (cancelledAt != null) result.cancelledAt = cancelledAt;
    if (serviceUntil != null) result.serviceUntil = serviceUntil;
    return result;
  }

  CancelSubscriptionResponse._();

  factory CancelSubscriptionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionResponse()..mergeFromBuffer(data, registry);
  factory CancelSubscriptionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelSubscriptionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelSubscriptionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: CancelSubscriptionResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'cancelledAt')
    ..aOS(2, _omitFieldNames ? '' : 'serviceUntil')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelSubscriptionResponse copyWith(
          void Function(CancelSubscriptionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CancelSubscriptionResponse))
          as CancelSubscriptionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelSubscriptionResponse() / CancelSubscriptionResponse.new instead')
  static CancelSubscriptionResponse create() => CancelSubscriptionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CancelSubscriptionResponse._();
  @$core.override
  CancelSubscriptionResponse createEmptyInstance() =>
      CancelSubscriptionResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelSubscriptionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelSubscriptionResponse>(
          CancelSubscriptionResponse.$_createMessage);
  static CancelSubscriptionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get cancelledAt => $_getSZ(0);
  @$pb.TagNumber(1)
  set cancelledAt($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCancelledAt() => $_has(0);
  @$pb.TagNumber(1)
  void clearCancelledAt() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get serviceUntil => $_getSZ(1);
  @$pb.TagNumber(2)
  set serviceUntil($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasServiceUntil() => $_has(1);
  @$pb.TagNumber(2)
  void clearServiceUntil() => $_clearField(2);
}

class GetBillingSettingsRequest extends $pb.GeneratedMessage {
  factory GetBillingSettingsRequest() => GetBillingSettingsRequest._();

  GetBillingSettingsRequest._();

  factory GetBillingSettingsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingSettingsRequest()..mergeFromBuffer(data, registry);
  factory GetBillingSettingsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingSettingsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBillingSettingsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetBillingSettingsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingSettingsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingSettingsRequest copyWith(
          void Function(GetBillingSettingsRequest) updates) =>
      super.copyWith((message) => updates(message as GetBillingSettingsRequest))
          as GetBillingSettingsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetBillingSettingsRequest() / GetBillingSettingsRequest.new instead')
  static GetBillingSettingsRequest create() => GetBillingSettingsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetBillingSettingsRequest._();
  @$core.override
  GetBillingSettingsRequest createEmptyInstance() =>
      GetBillingSettingsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetBillingSettingsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetBillingSettingsRequest>(
          GetBillingSettingsRequest.$_createMessage);
  static GetBillingSettingsRequest? _defaultInstance;
}

class GetBillingSettingsResponse extends $pb.GeneratedMessage {
  factory GetBillingSettingsResponse({
    $core.Iterable<BillingSetting>? settings,
  }) {
    final result = GetBillingSettingsResponse._();
    if (settings != null) result.settings.addAll(settings);
    return result;
  }

  GetBillingSettingsResponse._();

  factory GetBillingSettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingSettingsResponse()..mergeFromBuffer(data, registry);
  factory GetBillingSettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBillingSettingsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBillingSettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetBillingSettingsResponse.$_createMessage)
    ..pPM<BillingSetting>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: BillingSetting.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingSettingsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBillingSettingsResponse copyWith(
          void Function(GetBillingSettingsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetBillingSettingsResponse))
          as GetBillingSettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetBillingSettingsResponse() / GetBillingSettingsResponse.new instead')
  static GetBillingSettingsResponse create() => GetBillingSettingsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetBillingSettingsResponse._();
  @$core.override
  GetBillingSettingsResponse createEmptyInstance() =>
      GetBillingSettingsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetBillingSettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetBillingSettingsResponse>(
          GetBillingSettingsResponse.$_createMessage);
  static GetBillingSettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<BillingSetting> get settings => $_getList(0);
}

/// BillingSetting:一項可由介面調整的營運參數(trial_days／grace_days／lead_days)。
class BillingSetting extends $pb.GeneratedMessage {
  factory BillingSetting({
    $core.String? key,
    $core.String? value,
    $core.String? description,
  }) {
    final result = BillingSetting._();
    if (key != null) result.key = key;
    if (value != null) result.value = value;
    if (description != null) result.description = description;
    return result;
  }

  BillingSetting._();

  factory BillingSetting.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BillingSetting()..mergeFromBuffer(data, registry);
  factory BillingSetting.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BillingSetting()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BillingSetting',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: BillingSetting.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'key')
    ..aOS(2, _omitFieldNames ? '' : 'value')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BillingSetting clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BillingSetting copyWith(void Function(BillingSetting) updates) =>
      super.copyWith((message) => updates(message as BillingSetting))
          as BillingSetting;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BillingSetting() / BillingSetting.new instead')
  static BillingSetting create() => BillingSetting._();
  static $pb.GeneratedMessage $_createMessage() => BillingSetting._();
  @$core.override
  BillingSetting createEmptyInstance() => BillingSetting._();
  @$core.pragma('dart2js:noInline')
  static BillingSetting getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BillingSetting>(
          BillingSetting.$_createMessage);
  static BillingSetting? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get key => $_getSZ(0);
  @$pb.TagNumber(1)
  set key($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearKey() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get value => $_getSZ(1);
  @$pb.TagNumber(2)
  set value($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);
}

class UpdateBillingSettingsRequest extends $pb.GeneratedMessage {
  factory UpdateBillingSettingsRequest({
    $core.Iterable<BillingSetting>? settings,
    $core.String? reason,
  }) {
    final result = UpdateBillingSettingsRequest._();
    if (settings != null) result.settings.addAll(settings);
    if (reason != null) result.reason = reason;
    return result;
  }

  UpdateBillingSettingsRequest._();

  factory UpdateBillingSettingsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateBillingSettingsRequest()..mergeFromBuffer(data, registry);
  factory UpdateBillingSettingsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateBillingSettingsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateBillingSettingsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: UpdateBillingSettingsRequest.$_createMessage)
    ..pPM<BillingSetting>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: BillingSetting.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateBillingSettingsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateBillingSettingsRequest copyWith(
          void Function(UpdateBillingSettingsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateBillingSettingsRequest))
          as UpdateBillingSettingsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateBillingSettingsRequest() / UpdateBillingSettingsRequest.new instead')
  static UpdateBillingSettingsRequest create() =>
      UpdateBillingSettingsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateBillingSettingsRequest._();
  @$core.override
  UpdateBillingSettingsRequest createEmptyInstance() =>
      UpdateBillingSettingsRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateBillingSettingsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateBillingSettingsRequest>(
          UpdateBillingSettingsRequest.$_createMessage);
  static UpdateBillingSettingsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<BillingSetting> get settings => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get reason => $_getSZ(1);
  @$pb.TagNumber(2)
  set reason($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearReason() => $_clearField(2);
}

class UpdateBillingSettingsResponse extends $pb.GeneratedMessage {
  factory UpdateBillingSettingsResponse({
    $core.Iterable<BillingSetting>? settings,
  }) {
    final result = UpdateBillingSettingsResponse._();
    if (settings != null) result.settings.addAll(settings);
    return result;
  }

  UpdateBillingSettingsResponse._();

  factory UpdateBillingSettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateBillingSettingsResponse()..mergeFromBuffer(data, registry);
  factory UpdateBillingSettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateBillingSettingsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateBillingSettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: UpdateBillingSettingsResponse.$_createMessage)
    ..pPM<BillingSetting>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: BillingSetting.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateBillingSettingsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateBillingSettingsResponse copyWith(
          void Function(UpdateBillingSettingsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateBillingSettingsResponse))
          as UpdateBillingSettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateBillingSettingsResponse() / UpdateBillingSettingsResponse.new instead')
  static UpdateBillingSettingsResponse create() =>
      UpdateBillingSettingsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateBillingSettingsResponse._();
  @$core.override
  UpdateBillingSettingsResponse createEmptyInstance() =>
      UpdateBillingSettingsResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateBillingSettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateBillingSettingsResponse>(
          UpdateBillingSettingsResponse.$_createMessage);
  static UpdateBillingSettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<BillingSetting> get settings => $_getList(0);
}

class SetTenantOverrideRequest extends $pb.GeneratedMessage {
  factory SetTenantOverrideRequest({
    $core.String? companyId,
    $core.String? featureCode,
    $core.bool? enabledSet,
    $core.bool? enabled,
    $core.bool? limitSet,
    $fixnum.Int64? limitValue,
    $core.String? owner,
    $core.String? expiresAt,
    $core.String? reason,
  }) {
    final result = SetTenantOverrideRequest._();
    if (companyId != null) result.companyId = companyId;
    if (featureCode != null) result.featureCode = featureCode;
    if (enabledSet != null) result.enabledSet = enabledSet;
    if (enabled != null) result.enabled = enabled;
    if (limitSet != null) result.limitSet = limitSet;
    if (limitValue != null) result.limitValue = limitValue;
    if (owner != null) result.owner = owner;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (reason != null) result.reason = reason;
    return result;
  }

  SetTenantOverrideRequest._();

  factory SetTenantOverrideRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTenantOverrideRequest()..mergeFromBuffer(data, registry);
  factory SetTenantOverrideRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTenantOverrideRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetTenantOverrideRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SetTenantOverrideRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'featureCode')
    ..aOB(3, _omitFieldNames ? '' : 'enabledSet')
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..aOB(5, _omitFieldNames ? '' : 'limitSet')
    ..aInt64(6, _omitFieldNames ? '' : 'limitValue')
    ..aOS(7, _omitFieldNames ? '' : 'owner')
    ..aOS(8, _omitFieldNames ? '' : 'expiresAt')
    ..aOS(9, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTenantOverrideRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTenantOverrideRequest copyWith(
          void Function(SetTenantOverrideRequest) updates) =>
      super.copyWith((message) => updates(message as SetTenantOverrideRequest))
          as SetTenantOverrideRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetTenantOverrideRequest() / SetTenantOverrideRequest.new instead')
  static SetTenantOverrideRequest create() => SetTenantOverrideRequest._();
  static $pb.GeneratedMessage $_createMessage() => SetTenantOverrideRequest._();
  @$core.override
  SetTenantOverrideRequest createEmptyInstance() =>
      SetTenantOverrideRequest._();
  @$core.pragma('dart2js:noInline')
  static SetTenantOverrideRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetTenantOverrideRequest>(
          SetTenantOverrideRequest.$_createMessage);
  static SetTenantOverrideRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

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
  $core.String get owner => $_getSZ(6);
  @$pb.TagNumber(7)
  set owner($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasOwner() => $_has(6);
  @$pb.TagNumber(7)
  void clearOwner() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get expiresAt => $_getSZ(7);
  @$pb.TagNumber(8)
  set expiresAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasExpiresAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearExpiresAt() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get reason => $_getSZ(8);
  @$pb.TagNumber(9)
  set reason($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasReason() => $_has(8);
  @$pb.TagNumber(9)
  void clearReason() => $_clearField(9);
}

class SetTenantOverrideResponse extends $pb.GeneratedMessage {
  factory SetTenantOverrideResponse({
    $core.String? id,
  }) {
    final result = SetTenantOverrideResponse._();
    if (id != null) result.id = id;
    return result;
  }

  SetTenantOverrideResponse._();

  factory SetTenantOverrideResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTenantOverrideResponse()..mergeFromBuffer(data, registry);
  factory SetTenantOverrideResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTenantOverrideResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetTenantOverrideResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SetTenantOverrideResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTenantOverrideResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTenantOverrideResponse copyWith(
          void Function(SetTenantOverrideResponse) updates) =>
      super.copyWith((message) => updates(message as SetTenantOverrideResponse))
          as SetTenantOverrideResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetTenantOverrideResponse() / SetTenantOverrideResponse.new instead')
  static SetTenantOverrideResponse create() => SetTenantOverrideResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetTenantOverrideResponse._();
  @$core.override
  SetTenantOverrideResponse createEmptyInstance() =>
      SetTenantOverrideResponse._();
  @$core.pragma('dart2js:noInline')
  static SetTenantOverrideResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetTenantOverrideResponse>(
          SetTenantOverrideResponse.$_createMessage);
  static SetTenantOverrideResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RevokeTenantOverrideRequest extends $pb.GeneratedMessage {
  factory RevokeTenantOverrideRequest({
    $core.String? overrideId,
    $core.String? reason,
  }) {
    final result = RevokeTenantOverrideRequest._();
    if (overrideId != null) result.overrideId = overrideId;
    if (reason != null) result.reason = reason;
    return result;
  }

  RevokeTenantOverrideRequest._();

  factory RevokeTenantOverrideRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeTenantOverrideRequest()..mergeFromBuffer(data, registry);
  factory RevokeTenantOverrideRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeTenantOverrideRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RevokeTenantOverrideRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: RevokeTenantOverrideRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'overrideId')
    ..aOS(2, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeTenantOverrideRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeTenantOverrideRequest copyWith(
          void Function(RevokeTenantOverrideRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RevokeTenantOverrideRequest))
          as RevokeTenantOverrideRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RevokeTenantOverrideRequest() / RevokeTenantOverrideRequest.new instead')
  static RevokeTenantOverrideRequest create() =>
      RevokeTenantOverrideRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RevokeTenantOverrideRequest._();
  @$core.override
  RevokeTenantOverrideRequest createEmptyInstance() =>
      RevokeTenantOverrideRequest._();
  @$core.pragma('dart2js:noInline')
  static RevokeTenantOverrideRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RevokeTenantOverrideRequest>(
          RevokeTenantOverrideRequest.$_createMessage);
  static RevokeTenantOverrideRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get overrideId => $_getSZ(0);
  @$pb.TagNumber(1)
  set overrideId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOverrideId() => $_has(0);
  @$pb.TagNumber(1)
  void clearOverrideId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get reason => $_getSZ(1);
  @$pb.TagNumber(2)
  set reason($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearReason() => $_clearField(2);
}

class RevokeTenantOverrideResponse extends $pb.GeneratedMessage {
  factory RevokeTenantOverrideResponse({
    $core.String? companyId,
    $core.String? featureCode,
  }) {
    final result = RevokeTenantOverrideResponse._();
    if (companyId != null) result.companyId = companyId;
    if (featureCode != null) result.featureCode = featureCode;
    return result;
  }

  RevokeTenantOverrideResponse._();

  factory RevokeTenantOverrideResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeTenantOverrideResponse()..mergeFromBuffer(data, registry);
  factory RevokeTenantOverrideResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeTenantOverrideResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RevokeTenantOverrideResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: RevokeTenantOverrideResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'featureCode')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeTenantOverrideResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeTenantOverrideResponse copyWith(
          void Function(RevokeTenantOverrideResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RevokeTenantOverrideResponse))
          as RevokeTenantOverrideResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RevokeTenantOverrideResponse() / RevokeTenantOverrideResponse.new instead')
  static RevokeTenantOverrideResponse create() =>
      RevokeTenantOverrideResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RevokeTenantOverrideResponse._();
  @$core.override
  RevokeTenantOverrideResponse createEmptyInstance() =>
      RevokeTenantOverrideResponse._();
  @$core.pragma('dart2js:noInline')
  static RevokeTenantOverrideResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RevokeTenantOverrideResponse>(
          RevokeTenantOverrideResponse.$_createMessage);
  static RevokeTenantOverrideResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get featureCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set featureCode($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFeatureCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearFeatureCode() => $_clearField(2);
}

class UpsertPlanPriceRequest extends $pb.GeneratedMessage {
  factory UpsertPlanPriceRequest({
    $core.String? planCode,
    $core.String? billingCycle,
    $core.String? basePrice,
    $core.String? seatPrice,
    $core.String? currency,
    $core.String? reason,
  }) {
    final result = UpsertPlanPriceRequest._();
    if (planCode != null) result.planCode = planCode;
    if (billingCycle != null) result.billingCycle = billingCycle;
    if (basePrice != null) result.basePrice = basePrice;
    if (seatPrice != null) result.seatPrice = seatPrice;
    if (currency != null) result.currency = currency;
    if (reason != null) result.reason = reason;
    return result;
  }

  UpsertPlanPriceRequest._();

  factory UpsertPlanPriceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpsertPlanPriceRequest()..mergeFromBuffer(data, registry);
  factory UpsertPlanPriceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpsertPlanPriceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpsertPlanPriceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: UpsertPlanPriceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'planCode')
    ..aOS(2, _omitFieldNames ? '' : 'billingCycle')
    ..aOS(3, _omitFieldNames ? '' : 'basePrice')
    ..aOS(4, _omitFieldNames ? '' : 'seatPrice')
    ..aOS(5, _omitFieldNames ? '' : 'currency')
    ..aOS(6, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpsertPlanPriceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpsertPlanPriceRequest copyWith(
          void Function(UpsertPlanPriceRequest) updates) =>
      super.copyWith((message) => updates(message as UpsertPlanPriceRequest))
          as UpsertPlanPriceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpsertPlanPriceRequest() / UpsertPlanPriceRequest.new instead')
  static UpsertPlanPriceRequest create() => UpsertPlanPriceRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpsertPlanPriceRequest._();
  @$core.override
  UpsertPlanPriceRequest createEmptyInstance() => UpsertPlanPriceRequest._();
  @$core.pragma('dart2js:noInline')
  static UpsertPlanPriceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpsertPlanPriceRequest>(
          UpsertPlanPriceRequest.$_createMessage);
  static UpsertPlanPriceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get planCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set planCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlanCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlanCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get billingCycle => $_getSZ(1);
  @$pb.TagNumber(2)
  set billingCycle($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBillingCycle() => $_has(1);
  @$pb.TagNumber(2)
  void clearBillingCycle() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get basePrice => $_getSZ(2);
  @$pb.TagNumber(3)
  set basePrice($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBasePrice() => $_has(2);
  @$pb.TagNumber(3)
  void clearBasePrice() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get seatPrice => $_getSZ(3);
  @$pb.TagNumber(4)
  set seatPrice($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSeatPrice() => $_has(3);
  @$pb.TagNumber(4)
  void clearSeatPrice() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get currency => $_getSZ(4);
  @$pb.TagNumber(5)
  set currency($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCurrency() => $_has(4);
  @$pb.TagNumber(5)
  void clearCurrency() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get reason => $_getSZ(5);
  @$pb.TagNumber(6)
  set reason($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasReason() => $_has(5);
  @$pb.TagNumber(6)
  void clearReason() => $_clearField(6);
}

class UpsertPlanPriceResponse extends $pb.GeneratedMessage {
  factory UpsertPlanPriceResponse() => UpsertPlanPriceResponse._();

  UpsertPlanPriceResponse._();

  factory UpsertPlanPriceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpsertPlanPriceResponse()..mergeFromBuffer(data, registry);
  factory UpsertPlanPriceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpsertPlanPriceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpsertPlanPriceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: UpsertPlanPriceResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpsertPlanPriceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpsertPlanPriceResponse copyWith(
          void Function(UpsertPlanPriceResponse) updates) =>
      super.copyWith((message) => updates(message as UpsertPlanPriceResponse))
          as UpsertPlanPriceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpsertPlanPriceResponse() / UpsertPlanPriceResponse.new instead')
  static UpsertPlanPriceResponse create() => UpsertPlanPriceResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpsertPlanPriceResponse._();
  @$core.override
  UpsertPlanPriceResponse createEmptyInstance() => UpsertPlanPriceResponse._();
  @$core.pragma('dart2js:noInline')
  static UpsertPlanPriceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpsertPlanPriceResponse>(
          UpsertPlanPriceResponse.$_createMessage);
  static UpsertPlanPriceResponse? _defaultInstance;
}

class SetPlanEntitlementRequest extends $pb.GeneratedMessage {
  factory SetPlanEntitlementRequest({
    $core.String? planCode,
    $core.String? featureCode,
    $core.bool? enabled,
    $core.bool? limitSet,
    $fixnum.Int64? limitValue,
    $core.String? reason,
  }) {
    final result = SetPlanEntitlementRequest._();
    if (planCode != null) result.planCode = planCode;
    if (featureCode != null) result.featureCode = featureCode;
    if (enabled != null) result.enabled = enabled;
    if (limitSet != null) result.limitSet = limitSet;
    if (limitValue != null) result.limitValue = limitValue;
    if (reason != null) result.reason = reason;
    return result;
  }

  SetPlanEntitlementRequest._();

  factory SetPlanEntitlementRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetPlanEntitlementRequest()..mergeFromBuffer(data, registry);
  factory SetPlanEntitlementRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetPlanEntitlementRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetPlanEntitlementRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SetPlanEntitlementRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'planCode')
    ..aOS(2, _omitFieldNames ? '' : 'featureCode')
    ..aOB(3, _omitFieldNames ? '' : 'enabled')
    ..aOB(4, _omitFieldNames ? '' : 'limitSet')
    ..aInt64(5, _omitFieldNames ? '' : 'limitValue')
    ..aOS(6, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetPlanEntitlementRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetPlanEntitlementRequest copyWith(
          void Function(SetPlanEntitlementRequest) updates) =>
      super.copyWith((message) => updates(message as SetPlanEntitlementRequest))
          as SetPlanEntitlementRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetPlanEntitlementRequest() / SetPlanEntitlementRequest.new instead')
  static SetPlanEntitlementRequest create() => SetPlanEntitlementRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetPlanEntitlementRequest._();
  @$core.override
  SetPlanEntitlementRequest createEmptyInstance() =>
      SetPlanEntitlementRequest._();
  @$core.pragma('dart2js:noInline')
  static SetPlanEntitlementRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetPlanEntitlementRequest>(
          SetPlanEntitlementRequest.$_createMessage);
  static SetPlanEntitlementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get planCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set planCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlanCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlanCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get featureCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set featureCode($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFeatureCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearFeatureCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get enabled => $_getBF(2);
  @$pb.TagNumber(3)
  set enabled($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearEnabled() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get limitSet => $_getBF(3);
  @$pb.TagNumber(4)
  set limitSet($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLimitSet() => $_has(3);
  @$pb.TagNumber(4)
  void clearLimitSet() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get limitValue => $_getI64(4);
  @$pb.TagNumber(5)
  set limitValue($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLimitValue() => $_has(4);
  @$pb.TagNumber(5)
  void clearLimitValue() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get reason => $_getSZ(5);
  @$pb.TagNumber(6)
  set reason($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasReason() => $_has(5);
  @$pb.TagNumber(6)
  void clearReason() => $_clearField(6);
}

class SetPlanEntitlementResponse extends $pb.GeneratedMessage {
  factory SetPlanEntitlementResponse() => SetPlanEntitlementResponse._();

  SetPlanEntitlementResponse._();

  factory SetPlanEntitlementResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetPlanEntitlementResponse()..mergeFromBuffer(data, registry);
  factory SetPlanEntitlementResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetPlanEntitlementResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetPlanEntitlementResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SetPlanEntitlementResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetPlanEntitlementResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetPlanEntitlementResponse copyWith(
          void Function(SetPlanEntitlementResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SetPlanEntitlementResponse))
          as SetPlanEntitlementResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetPlanEntitlementResponse() / SetPlanEntitlementResponse.new instead')
  static SetPlanEntitlementResponse create() => SetPlanEntitlementResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetPlanEntitlementResponse._();
  @$core.override
  SetPlanEntitlementResponse createEmptyInstance() =>
      SetPlanEntitlementResponse._();
  @$core.pragma('dart2js:noInline')
  static SetPlanEntitlementResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetPlanEntitlementResponse>(
          SetPlanEntitlementResponse.$_createMessage);
  static SetPlanEntitlementResponse? _defaultInstance;
}

class CreateOperatorRequest extends $pb.GeneratedMessage {
  factory CreateOperatorRequest({
    $core.String? email,
    $core.String? name,
    $core.String? role,
    $core.String? reason,
  }) {
    final result = CreateOperatorRequest._();
    if (email != null) result.email = email;
    if (name != null) result.name = name;
    if (role != null) result.role = role;
    if (reason != null) result.reason = reason;
    return result;
  }

  CreateOperatorRequest._();

  factory CreateOperatorRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOperatorRequest()..mergeFromBuffer(data, registry);
  factory CreateOperatorRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOperatorRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateOperatorRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: CreateOperatorRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'email')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'role')
    ..aOS(4, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOperatorRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOperatorRequest copyWith(
          void Function(CreateOperatorRequest) updates) =>
      super.copyWith((message) => updates(message as CreateOperatorRequest))
          as CreateOperatorRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateOperatorRequest() / CreateOperatorRequest.new instead')
  static CreateOperatorRequest create() => CreateOperatorRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateOperatorRequest._();
  @$core.override
  CreateOperatorRequest createEmptyInstance() => CreateOperatorRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateOperatorRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateOperatorRequest>(
          CreateOperatorRequest.$_createMessage);
  static CreateOperatorRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get email => $_getSZ(0);
  @$pb.TagNumber(1)
  set email($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEmail() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmail() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get role => $_getSZ(2);
  @$pb.TagNumber(3)
  set role($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRole() => $_has(2);
  @$pb.TagNumber(3)
  void clearRole() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get reason => $_getSZ(3);
  @$pb.TagNumber(4)
  set reason($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasReason() => $_has(3);
  @$pb.TagNumber(4)
  void clearReason() => $_clearField(4);
}

class CreateOperatorResponse extends $pb.GeneratedMessage {
  factory CreateOperatorResponse({
    $core.String? id,
  }) {
    final result = CreateOperatorResponse._();
    if (id != null) result.id = id;
    return result;
  }

  CreateOperatorResponse._();

  factory CreateOperatorResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOperatorResponse()..mergeFromBuffer(data, registry);
  factory CreateOperatorResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOperatorResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateOperatorResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: CreateOperatorResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOperatorResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOperatorResponse copyWith(
          void Function(CreateOperatorResponse) updates) =>
      super.copyWith((message) => updates(message as CreateOperatorResponse))
          as CreateOperatorResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateOperatorResponse() / CreateOperatorResponse.new instead')
  static CreateOperatorResponse create() => CreateOperatorResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateOperatorResponse._();
  @$core.override
  CreateOperatorResponse createEmptyInstance() => CreateOperatorResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateOperatorResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateOperatorResponse>(
          CreateOperatorResponse.$_createMessage);
  static CreateOperatorResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DisableOperatorRequest extends $pb.GeneratedMessage {
  factory DisableOperatorRequest({
    $core.String? operatorId,
    $core.String? reason,
  }) {
    final result = DisableOperatorRequest._();
    if (operatorId != null) result.operatorId = operatorId;
    if (reason != null) result.reason = reason;
    return result;
  }

  DisableOperatorRequest._();

  factory DisableOperatorRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisableOperatorRequest()..mergeFromBuffer(data, registry);
  factory DisableOperatorRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisableOperatorRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisableOperatorRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: DisableOperatorRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'operatorId')
    ..aOS(2, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableOperatorRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableOperatorRequest copyWith(
          void Function(DisableOperatorRequest) updates) =>
      super.copyWith((message) => updates(message as DisableOperatorRequest))
          as DisableOperatorRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DisableOperatorRequest() / DisableOperatorRequest.new instead')
  static DisableOperatorRequest create() => DisableOperatorRequest._();
  static $pb.GeneratedMessage $_createMessage() => DisableOperatorRequest._();
  @$core.override
  DisableOperatorRequest createEmptyInstance() => DisableOperatorRequest._();
  @$core.pragma('dart2js:noInline')
  static DisableOperatorRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisableOperatorRequest>(
          DisableOperatorRequest.$_createMessage);
  static DisableOperatorRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get operatorId => $_getSZ(0);
  @$pb.TagNumber(1)
  set operatorId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOperatorId() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperatorId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get reason => $_getSZ(1);
  @$pb.TagNumber(2)
  set reason($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearReason() => $_clearField(2);
}

class DisableOperatorResponse extends $pb.GeneratedMessage {
  factory DisableOperatorResponse() => DisableOperatorResponse._();

  DisableOperatorResponse._();

  factory DisableOperatorResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisableOperatorResponse()..mergeFromBuffer(data, registry);
  factory DisableOperatorResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisableOperatorResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisableOperatorResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: DisableOperatorResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableOperatorResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableOperatorResponse copyWith(
          void Function(DisableOperatorResponse) updates) =>
      super.copyWith((message) => updates(message as DisableOperatorResponse))
          as DisableOperatorResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DisableOperatorResponse() / DisableOperatorResponse.new instead')
  static DisableOperatorResponse create() => DisableOperatorResponse._();
  static $pb.GeneratedMessage $_createMessage() => DisableOperatorResponse._();
  @$core.override
  DisableOperatorResponse createEmptyInstance() => DisableOperatorResponse._();
  @$core.pragma('dart2js:noInline')
  static DisableOperatorResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisableOperatorResponse>(
          DisableOperatorResponse.$_createMessage);
  static DisableOperatorResponse? _defaultInstance;
}

class GetOperatorSelfRequest extends $pb.GeneratedMessage {
  factory GetOperatorSelfRequest() => GetOperatorSelfRequest._();

  GetOperatorSelfRequest._();

  factory GetOperatorSelfRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOperatorSelfRequest()..mergeFromBuffer(data, registry);
  factory GetOperatorSelfRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOperatorSelfRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetOperatorSelfRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetOperatorSelfRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOperatorSelfRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOperatorSelfRequest copyWith(
          void Function(GetOperatorSelfRequest) updates) =>
      super.copyWith((message) => updates(message as GetOperatorSelfRequest))
          as GetOperatorSelfRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetOperatorSelfRequest() / GetOperatorSelfRequest.new instead')
  static GetOperatorSelfRequest create() => GetOperatorSelfRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetOperatorSelfRequest._();
  @$core.override
  GetOperatorSelfRequest createEmptyInstance() => GetOperatorSelfRequest._();
  @$core.pragma('dart2js:noInline')
  static GetOperatorSelfRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetOperatorSelfRequest>(
          GetOperatorSelfRequest.$_createMessage);
  static GetOperatorSelfRequest? _defaultInstance;
}

class GetOperatorSelfResponse extends $pb.GeneratedMessage {
  factory GetOperatorSelfResponse({
    $core.String? operatorId,
    $core.String? email,
    $core.String? role,
  }) {
    final result = GetOperatorSelfResponse._();
    if (operatorId != null) result.operatorId = operatorId;
    if (email != null) result.email = email;
    if (role != null) result.role = role;
    return result;
  }

  GetOperatorSelfResponse._();

  factory GetOperatorSelfResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOperatorSelfResponse()..mergeFromBuffer(data, registry);
  factory GetOperatorSelfResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOperatorSelfResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetOperatorSelfResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: GetOperatorSelfResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'operatorId')
    ..aOS(2, _omitFieldNames ? '' : 'email')
    ..aOS(3, _omitFieldNames ? '' : 'role')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOperatorSelfResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOperatorSelfResponse copyWith(
          void Function(GetOperatorSelfResponse) updates) =>
      super.copyWith((message) => updates(message as GetOperatorSelfResponse))
          as GetOperatorSelfResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetOperatorSelfResponse() / GetOperatorSelfResponse.new instead')
  static GetOperatorSelfResponse create() => GetOperatorSelfResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetOperatorSelfResponse._();
  @$core.override
  GetOperatorSelfResponse createEmptyInstance() => GetOperatorSelfResponse._();
  @$core.pragma('dart2js:noInline')
  static GetOperatorSelfResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetOperatorSelfResponse>(
          GetOperatorSelfResponse.$_createMessage);
  static GetOperatorSelfResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get operatorId => $_getSZ(0);
  @$pb.TagNumber(1)
  set operatorId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOperatorId() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperatorId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get email => $_getSZ(1);
  @$pb.TagNumber(2)
  set email($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEmail() => $_has(1);
  @$pb.TagNumber(2)
  void clearEmail() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get role => $_getSZ(2);
  @$pb.TagNumber(3)
  set role($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRole() => $_has(2);
  @$pb.TagNumber(3)
  void clearRole() => $_clearField(3);
}

class ListSubscriptionPeriodsRequest extends $pb.GeneratedMessage {
  factory ListSubscriptionPeriodsRequest({
    $core.String? companyId,
  }) {
    final result = ListSubscriptionPeriodsRequest._();
    if (companyId != null) result.companyId = companyId;
    return result;
  }

  ListSubscriptionPeriodsRequest._();

  factory ListSubscriptionPeriodsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionPeriodsRequest()..mergeFromBuffer(data, registry);
  factory ListSubscriptionPeriodsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionPeriodsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSubscriptionPeriodsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListSubscriptionPeriodsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionPeriodsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionPeriodsRequest copyWith(
          void Function(ListSubscriptionPeriodsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListSubscriptionPeriodsRequest))
          as ListSubscriptionPeriodsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSubscriptionPeriodsRequest() / ListSubscriptionPeriodsRequest.new instead')
  static ListSubscriptionPeriodsRequest create() =>
      ListSubscriptionPeriodsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListSubscriptionPeriodsRequest._();
  @$core.override
  ListSubscriptionPeriodsRequest createEmptyInstance() =>
      ListSubscriptionPeriodsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListSubscriptionPeriodsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSubscriptionPeriodsRequest>(
          ListSubscriptionPeriodsRequest.$_createMessage);
  static ListSubscriptionPeriodsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);
}

class SubscriptionPeriod extends $pb.GeneratedMessage {
  factory SubscriptionPeriod({
    $core.int? periodNo,
    $core.String? periodStart,
    $core.String? periodEnd,
    $core.String? status,
    $core.String? amount,
    $core.String? paidAt,
    $core.String? invoiceNo,
    $core.String? externalRef,
    $core.String? note,
  }) {
    final result = SubscriptionPeriod._();
    if (periodNo != null) result.periodNo = periodNo;
    if (periodStart != null) result.periodStart = periodStart;
    if (periodEnd != null) result.periodEnd = periodEnd;
    if (status != null) result.status = status;
    if (amount != null) result.amount = amount;
    if (paidAt != null) result.paidAt = paidAt;
    if (invoiceNo != null) result.invoiceNo = invoiceNo;
    if (externalRef != null) result.externalRef = externalRef;
    if (note != null) result.note = note;
    return result;
  }

  SubscriptionPeriod._();

  factory SubscriptionPeriod.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SubscriptionPeriod()..mergeFromBuffer(data, registry);
  factory SubscriptionPeriod.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SubscriptionPeriod()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SubscriptionPeriod',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: SubscriptionPeriod.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'periodNo')
    ..aOS(2, _omitFieldNames ? '' : 'periodStart')
    ..aOS(3, _omitFieldNames ? '' : 'periodEnd')
    ..aOS(4, _omitFieldNames ? '' : 'status')
    ..aOS(5, _omitFieldNames ? '' : 'amount')
    ..aOS(6, _omitFieldNames ? '' : 'paidAt')
    ..aOS(7, _omitFieldNames ? '' : 'invoiceNo')
    ..aOS(8, _omitFieldNames ? '' : 'externalRef')
    ..aOS(9, _omitFieldNames ? '' : 'note')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SubscriptionPeriod clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SubscriptionPeriod copyWith(void Function(SubscriptionPeriod) updates) =>
      super.copyWith((message) => updates(message as SubscriptionPeriod))
          as SubscriptionPeriod;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SubscriptionPeriod() / SubscriptionPeriod.new instead')
  static SubscriptionPeriod create() => SubscriptionPeriod._();
  static $pb.GeneratedMessage $_createMessage() => SubscriptionPeriod._();
  @$core.override
  SubscriptionPeriod createEmptyInstance() => SubscriptionPeriod._();
  @$core.pragma('dart2js:noInline')
  static SubscriptionPeriod getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SubscriptionPeriod>(
          SubscriptionPeriod.$_createMessage);
  static SubscriptionPeriod? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get periodNo => $_getIZ(0);
  @$pb.TagNumber(1)
  set periodNo($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPeriodNo() => $_has(0);
  @$pb.TagNumber(1)
  void clearPeriodNo() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get periodStart => $_getSZ(1);
  @$pb.TagNumber(2)
  set periodStart($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPeriodStart() => $_has(1);
  @$pb.TagNumber(2)
  void clearPeriodStart() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get periodEnd => $_getSZ(2);
  @$pb.TagNumber(3)
  set periodEnd($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPeriodEnd() => $_has(2);
  @$pb.TagNumber(3)
  void clearPeriodEnd() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get status => $_getSZ(3);
  @$pb.TagNumber(4)
  set status($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get amount => $_getSZ(4);
  @$pb.TagNumber(5)
  set amount($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmount() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get paidAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set paidAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPaidAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearPaidAt() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get invoiceNo => $_getSZ(6);
  @$pb.TagNumber(7)
  set invoiceNo($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInvoiceNo() => $_has(6);
  @$pb.TagNumber(7)
  void clearInvoiceNo() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get externalRef => $_getSZ(7);
  @$pb.TagNumber(8)
  set externalRef($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasExternalRef() => $_has(7);
  @$pb.TagNumber(8)
  void clearExternalRef() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get note => $_getSZ(8);
  @$pb.TagNumber(9)
  set note($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasNote() => $_has(8);
  @$pb.TagNumber(9)
  void clearNote() => $_clearField(9);
}

class ListSubscriptionPeriodsResponse extends $pb.GeneratedMessage {
  factory ListSubscriptionPeriodsResponse({
    $core.Iterable<SubscriptionPeriod>? periods,
  }) {
    final result = ListSubscriptionPeriodsResponse._();
    if (periods != null) result.periods.addAll(periods);
    return result;
  }

  ListSubscriptionPeriodsResponse._();

  factory ListSubscriptionPeriodsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionPeriodsResponse()..mergeFromBuffer(data, registry);
  factory ListSubscriptionPeriodsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSubscriptionPeriodsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSubscriptionPeriodsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'platform.v1'),
      createEmptyInstance: ListSubscriptionPeriodsResponse.$_createMessage)
    ..pPM<SubscriptionPeriod>(1, _omitFieldNames ? '' : 'periods',
        subBuilder: SubscriptionPeriod.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionPeriodsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSubscriptionPeriodsResponse copyWith(
          void Function(ListSubscriptionPeriodsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListSubscriptionPeriodsResponse))
          as ListSubscriptionPeriodsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSubscriptionPeriodsResponse() / ListSubscriptionPeriodsResponse.new instead')
  static ListSubscriptionPeriodsResponse create() =>
      ListSubscriptionPeriodsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListSubscriptionPeriodsResponse._();
  @$core.override
  ListSubscriptionPeriodsResponse createEmptyInstance() =>
      ListSubscriptionPeriodsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListSubscriptionPeriodsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSubscriptionPeriodsResponse>(
          ListSubscriptionPeriodsResponse.$_createMessage);
  static ListSubscriptionPeriodsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<SubscriptionPeriod> get periods => $_getList(0);
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

  /// --- 平台寫入(T9)。共同契約:每個寫入都必須帶 reason(平台稽核必填),
  /// actor 一律是 cookie 上的真實 operator;資料與稽核同一個交易(失敗不留半成品)。
  $async.Future<ListReceivablesResponse> listReceivables(
          $pb.ClientContext? ctx, ListReceivablesRequest request) =>
      _client.invoke<ListReceivablesResponse>(ctx, 'PlatformAdminService',
          'ListReceivables', request, ListReceivablesResponse());
  $async.Future<RecordPaymentResponse> recordPayment(
          $pb.ClientContext? ctx, RecordPaymentRequest request) =>
      _client.invoke<RecordPaymentResponse>(ctx, 'PlatformAdminService',
          'RecordPayment', request, RecordPaymentResponse());
  $async.Future<CreateSubscriptionResponse> createSubscription(
          $pb.ClientContext? ctx, CreateSubscriptionRequest request) =>
      _client.invoke<CreateSubscriptionResponse>(ctx, 'PlatformAdminService',
          'CreateSubscription', request, CreateSubscriptionResponse());
  $async.Future<SetSeatCountResponse> setSeatCount(
          $pb.ClientContext? ctx, SetSeatCountRequest request) =>
      _client.invoke<SetSeatCountResponse>(ctx, 'PlatformAdminService',
          'SetSeatCount', request, SetSeatCountResponse());
  $async.Future<ChangePlanResponse> changePlan(
          $pb.ClientContext? ctx, ChangePlanRequest request) =>
      _client.invoke<ChangePlanResponse>(ctx, 'PlatformAdminService',
          'ChangePlan', request, ChangePlanResponse());
  $async.Future<CancelSubscriptionResponse> cancelSubscription(
          $pb.ClientContext? ctx, CancelSubscriptionRequest request) =>
      _client.invoke<CancelSubscriptionResponse>(ctx, 'PlatformAdminService',
          'CancelSubscription', request, CancelSubscriptionResponse());
  $async.Future<GetBillingSettingsResponse> getBillingSettings(
          $pb.ClientContext? ctx, GetBillingSettingsRequest request) =>
      _client.invoke<GetBillingSettingsResponse>(ctx, 'PlatformAdminService',
          'GetBillingSettings', request, GetBillingSettingsResponse());
  $async.Future<UpdateBillingSettingsResponse> updateBillingSettings(
          $pb.ClientContext? ctx, UpdateBillingSettingsRequest request) =>
      _client.invoke<UpdateBillingSettingsResponse>(ctx, 'PlatformAdminService',
          'UpdateBillingSettings', request, UpdateBillingSettingsResponse());
  $async.Future<SetTenantOverrideResponse> setTenantOverride(
          $pb.ClientContext? ctx, SetTenantOverrideRequest request) =>
      _client.invoke<SetTenantOverrideResponse>(ctx, 'PlatformAdminService',
          'SetTenantOverride', request, SetTenantOverrideResponse());
  $async.Future<RevokeTenantOverrideResponse> revokeTenantOverride(
          $pb.ClientContext? ctx, RevokeTenantOverrideRequest request) =>
      _client.invoke<RevokeTenantOverrideResponse>(ctx, 'PlatformAdminService',
          'RevokeTenantOverride', request, RevokeTenantOverrideResponse());
  $async.Future<UpsertPlanPriceResponse> upsertPlanPrice(
          $pb.ClientContext? ctx, UpsertPlanPriceRequest request) =>
      _client.invoke<UpsertPlanPriceResponse>(ctx, 'PlatformAdminService',
          'UpsertPlanPrice', request, UpsertPlanPriceResponse());
  $async.Future<SetPlanEntitlementResponse> setPlanEntitlement(
          $pb.ClientContext? ctx, SetPlanEntitlementRequest request) =>
      _client.invoke<SetPlanEntitlementResponse>(ctx, 'PlatformAdminService',
          'SetPlanEntitlement', request, SetPlanEntitlementResponse());
  $async.Future<CreateOperatorResponse> createOperator(
          $pb.ClientContext? ctx, CreateOperatorRequest request) =>
      _client.invoke<CreateOperatorResponse>(ctx, 'PlatformAdminService',
          'CreateOperator', request, CreateOperatorResponse());
  $async.Future<DisableOperatorResponse> disableOperator(
          $pb.ClientContext? ctx, DisableOperatorRequest request) =>
      _client.invoke<DisableOperatorResponse>(ctx, 'PlatformAdminService',
          'DisableOperator', request, DisableOperatorResponse());

  /// 未結項 #23：自己的身分（console 依角色隱藏操作）。後端仍是唯一決策者；
  /// 前端只據此 disable 按鈕，不做授權判斷。
  $async.Future<GetOperatorSelfResponse> getOperatorSelf(
          $pb.ClientContext? ctx, GetOperatorSelfRequest request) =>
      _client.invoke<GetOperatorSelfResponse>(ctx, 'PlatformAdminService',
          'GetOperatorSelf', request, GetOperatorSelfResponse());

  /// 未結項 #28：某訂閱的期別歷史（租戶詳情的「期別」段；spec §2.4）。
  $async.Future<ListSubscriptionPeriodsResponse> listSubscriptionPeriods(
          $pb.ClientContext? ctx, ListSubscriptionPeriodsRequest request) =>
      _client.invoke<ListSubscriptionPeriodsResponse>(
          ctx,
          'PlatformAdminService',
          'ListSubscriptionPeriods',
          request,
          ListSubscriptionPeriodsResponse());
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
