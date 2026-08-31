// This is a generated file - do not edit.
//
// Generated from salesorder/v1/ability.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../google/protobuf/struct.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// AbilityRule:單條 CASL 能力規則(前端 @casl/ability 直接消費)。
class AbilityRule extends $pb.GeneratedMessage {
  factory AbilityRule({
    $core.String? action,
    $core.String? subject,
    $0.Struct? conditions,
    $core.bool? inverted,
  }) {
    final result = create();
    if (action != null) result.action = action;
    if (subject != null) result.subject = subject;
    if (conditions != null) result.conditions = conditions;
    if (inverted != null) result.inverted = inverted;
    return result;
  }

  AbilityRule._();

  factory AbilityRule.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AbilityRule.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AbilityRule',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'action')
    ..aOS(2, _omitFieldNames ? '' : 'subject')
    ..aOM<$0.Struct>(3, _omitFieldNames ? '' : 'conditions',
        subBuilder: $0.Struct.create)
    ..aOB(4, _omitFieldNames ? '' : 'inverted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AbilityRule clone() => AbilityRule()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AbilityRule copyWith(void Function(AbilityRule) updates) =>
      super.copyWith((message) => updates(message as AbilityRule))
          as AbilityRule;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AbilityRule create() => AbilityRule._();
  @$core.override
  AbilityRule createEmptyInstance() => create();
  static $pb.PbList<AbilityRule> createRepeated() => $pb.PbList<AbilityRule>();
  @$core.pragma('dart2js:noInline')
  static AbilityRule getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AbilityRule>(create);
  static AbilityRule? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get action => $_getSZ(0);
  @$pb.TagNumber(1)
  set action($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAction() => $_has(0);
  @$pb.TagNumber(1)
  void clearAction() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get subject => $_getSZ(1);
  @$pb.TagNumber(2)
  set subject($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSubject() => $_has(1);
  @$pb.TagNumber(2)
  void clearSubject() => $_clearField(2);

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
}

/// GetAbilityRequest:取得目前身分的能力規則。
class GetAbilityRequest extends $pb.GeneratedMessage {
  factory GetAbilityRequest() => create();

  GetAbilityRequest._();

  factory GetAbilityRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetAbilityRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAbilityRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAbilityRequest clone() => GetAbilityRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAbilityRequest copyWith(void Function(GetAbilityRequest) updates) =>
      super.copyWith((message) => updates(message as GetAbilityRequest))
          as GetAbilityRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAbilityRequest create() => GetAbilityRequest._();
  @$core.override
  GetAbilityRequest createEmptyInstance() => create();
  static $pb.PbList<GetAbilityRequest> createRepeated() =>
      $pb.PbList<GetAbilityRequest>();
  @$core.pragma('dart2js:noInline')
  static GetAbilityRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAbilityRequest>(create);
  static GetAbilityRequest? _defaultInstance;
}

/// GetAbilityResponse:能力規則清單(依 sort_order 排序,佔位符已以身分展開)。
class GetAbilityResponse extends $pb.GeneratedMessage {
  factory GetAbilityResponse({
    $core.Iterable<AbilityRule>? rules,
  }) {
    final result = create();
    if (rules != null) result.rules.addAll(rules);
    return result;
  }

  GetAbilityResponse._();

  factory GetAbilityResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetAbilityResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAbilityResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: create)
    ..pc<AbilityRule>(1, _omitFieldNames ? '' : 'rules', $pb.PbFieldType.PM,
        subBuilder: AbilityRule.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAbilityResponse clone() => GetAbilityResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAbilityResponse copyWith(void Function(GetAbilityResponse) updates) =>
      super.copyWith((message) => updates(message as GetAbilityResponse))
          as GetAbilityResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAbilityResponse create() => GetAbilityResponse._();
  @$core.override
  GetAbilityResponse createEmptyInstance() => create();
  static $pb.PbList<GetAbilityResponse> createRepeated() =>
      $pb.PbList<GetAbilityResponse>();
  @$core.pragma('dart2js:noInline')
  static GetAbilityResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAbilityResponse>(create);
  static GetAbilityResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AbilityRule> get rules => $_getList(0);
}

/// AbilityService:能力規則下發。
class AbilityServiceApi {
  final $pb.RpcClient _client;

  AbilityServiceApi(this._client);

  /// GetAbility:以 ctx 身分查詢規則表,輸出 CASL JSON 規則。
  $async.Future<GetAbilityResponse> getAbility(
          $pb.ClientContext? ctx, GetAbilityRequest request) =>
      _client.invoke<GetAbilityResponse>(
          ctx, 'AbilityService', 'GetAbility', request, GetAbilityResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
