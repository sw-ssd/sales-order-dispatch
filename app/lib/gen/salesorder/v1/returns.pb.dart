// This is a generated file - do not edit.
//
// Generated from salesorder/v1/returns.proto.

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

/// ReturnItemInput:單品項輸入。
class ReturnItemInput extends $pb.GeneratedMessage {
  factory ReturnItemInput({
    $core.String? sourceType,
    $core.String? salesOrderItemId,
    $core.String? customerProductId,
    $core.String? quantity,
    $core.String? reason,
    $core.Iterable<$core.String>? photoFileIds,
  }) {
    final result = ReturnItemInput._();
    if (sourceType != null) result.sourceType = sourceType;
    if (salesOrderItemId != null) result.salesOrderItemId = salesOrderItemId;
    if (customerProductId != null) result.customerProductId = customerProductId;
    if (quantity != null) result.quantity = quantity;
    if (reason != null) result.reason = reason;
    if (photoFileIds != null) result.photoFileIds.addAll(photoFileIds);
    return result;
  }

  ReturnItemInput._();

  factory ReturnItemInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReturnItemInput()..mergeFromBuffer(data, registry);
  factory ReturnItemInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReturnItemInput()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReturnItemInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ReturnItemInput.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sourceType')
    ..aOS(2, _omitFieldNames ? '' : 'salesOrderItemId')
    ..aOS(3, _omitFieldNames ? '' : 'customerProductId')
    ..aOS(4, _omitFieldNames ? '' : 'quantity')
    ..aOS(5, _omitFieldNames ? '' : 'reason')
    ..pPS(6, _omitFieldNames ? '' : 'photoFileIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReturnItemInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReturnItemInput copyWith(void Function(ReturnItemInput) updates) =>
      super.copyWith((message) => updates(message as ReturnItemInput))
          as ReturnItemInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ReturnItemInput() / ReturnItemInput.new instead')
  static ReturnItemInput create() => ReturnItemInput._();
  static $pb.GeneratedMessage $_createMessage() => ReturnItemInput._();
  @$core.override
  ReturnItemInput createEmptyInstance() => ReturnItemInput._();
  @$core.pragma('dart2js:noInline')
  static ReturnItemInput getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ReturnItemInput>(
          ReturnItemInput.$_createMessage);
  static ReturnItemInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sourceType => $_getSZ(0);
  @$pb.TagNumber(1)
  set sourceType($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSourceType() => $_has(0);
  @$pb.TagNumber(1)
  void clearSourceType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get salesOrderItemId => $_getSZ(1);
  @$pb.TagNumber(2)
  set salesOrderItemId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSalesOrderItemId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSalesOrderItemId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get customerProductId => $_getSZ(2);
  @$pb.TagNumber(3)
  set customerProductId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCustomerProductId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCustomerProductId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get quantity => $_getSZ(3);
  @$pb.TagNumber(4)
  set quantity($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasQuantity() => $_has(3);
  @$pb.TagNumber(4)
  void clearQuantity() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get reason => $_getSZ(4);
  @$pb.TagNumber(5)
  set reason($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasReason() => $_has(4);
  @$pb.TagNumber(5)
  void clearReason() => $_clearField(5);

  @$pb.TagNumber(6)
  $pb.PbList<$core.String> get photoFileIds => $_getList(5);
}

/// CreateReturnRequestRequest:發起退貨請求。
class CreateReturnRequestRequest extends $pb.GeneratedMessage {
  factory CreateReturnRequestRequest({
    $core.Iterable<ReturnItemInput>? items,
    $core.String? remark,
  }) {
    final result = CreateReturnRequestRequest._();
    if (items != null) result.items.addAll(items);
    if (remark != null) result.remark = remark;
    return result;
  }

  CreateReturnRequestRequest._();

  factory CreateReturnRequestRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateReturnRequestRequest()..mergeFromBuffer(data, registry);
  factory CreateReturnRequestRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateReturnRequestRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateReturnRequestRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateReturnRequestRequest.$_createMessage)
    ..pPM<ReturnItemInput>(1, _omitFieldNames ? '' : 'items',
        subBuilder: ReturnItemInput.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'remark')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateReturnRequestRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateReturnRequestRequest copyWith(
          void Function(CreateReturnRequestRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateReturnRequestRequest))
          as CreateReturnRequestRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateReturnRequestRequest() / CreateReturnRequestRequest.new instead')
  static CreateReturnRequestRequest create() => CreateReturnRequestRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateReturnRequestRequest._();
  @$core.override
  CreateReturnRequestRequest createEmptyInstance() =>
      CreateReturnRequestRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateReturnRequestRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateReturnRequestRequest>(
          CreateReturnRequestRequest.$_createMessage);
  static CreateReturnRequestRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ReturnItemInput> get items => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get remark => $_getSZ(1);
  @$pb.TagNumber(2)
  set remark($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRemark() => $_has(1);
  @$pb.TagNumber(2)
  void clearRemark() => $_clearField(2);
}

/// CreateReturnRequestResponse:發起結果。
class CreateReturnRequestResponse extends $pb.GeneratedMessage {
  factory CreateReturnRequestResponse({
    $core.String? id,
    $core.String? status,
  }) {
    final result = CreateReturnRequestResponse._();
    if (id != null) result.id = id;
    if (status != null) result.status = status;
    return result;
  }

  CreateReturnRequestResponse._();

  factory CreateReturnRequestResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateReturnRequestResponse()..mergeFromBuffer(data, registry);
  factory CreateReturnRequestResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateReturnRequestResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateReturnRequestResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateReturnRequestResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateReturnRequestResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateReturnRequestResponse copyWith(
          void Function(CreateReturnRequestResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateReturnRequestResponse))
          as CreateReturnRequestResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateReturnRequestResponse() / CreateReturnRequestResponse.new instead')
  static CreateReturnRequestResponse create() =>
      CreateReturnRequestResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateReturnRequestResponse._();
  @$core.override
  CreateReturnRequestResponse createEmptyInstance() =>
      CreateReturnRequestResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateReturnRequestResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateReturnRequestResponse>(
          CreateReturnRequestResponse.$_createMessage);
  static CreateReturnRequestResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get status => $_getSZ(1);
  @$pb.TagNumber(2)
  set status($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);
}

/// ListReturnRequestsRequest:列表請求。
class ListReturnRequestsRequest extends $pb.GeneratedMessage {
  factory ListReturnRequestsRequest({
    $core.String? status,
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListReturnRequestsRequest._();
    if (status != null) result.status = status;
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListReturnRequestsRequest._();

  factory ListReturnRequestsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReturnRequestsRequest()..mergeFromBuffer(data, registry);
  factory ListReturnRequestsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReturnRequestsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListReturnRequestsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListReturnRequestsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'status')
    ..aI(2, _omitFieldNames ? '' : 'page')
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReturnRequestsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReturnRequestsRequest copyWith(
          void Function(ListReturnRequestsRequest) updates) =>
      super.copyWith((message) => updates(message as ListReturnRequestsRequest))
          as ListReturnRequestsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListReturnRequestsRequest() / ListReturnRequestsRequest.new instead')
  static ListReturnRequestsRequest create() => ListReturnRequestsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListReturnRequestsRequest._();
  @$core.override
  ListReturnRequestsRequest createEmptyInstance() =>
      ListReturnRequestsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListReturnRequestsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListReturnRequestsRequest>(
          ListReturnRequestsRequest.$_createMessage);
  static ListReturnRequestsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get status => $_getSZ(0);
  @$pb.TagNumber(1)
  set status($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get page => $_getIZ(1);
  @$pb.TagNumber(2)
  set page($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPage() => $_has(1);
  @$pb.TagNumber(2)
  void clearPage() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get pageSize => $_getIZ(2);
  @$pb.TagNumber(3)
  set pageSize($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageSize() => $_clearField(3);
}

/// ReturnRequestEntry:列表單筆。
class ReturnRequestEntry extends $pb.GeneratedMessage {
  factory ReturnRequestEntry({
    $core.String? id,
    $core.String? customerId,
    $core.String? status,
    $core.String? remark,
    $core.String? createdAt,
  }) {
    final result = ReturnRequestEntry._();
    if (id != null) result.id = id;
    if (customerId != null) result.customerId = customerId;
    if (status != null) result.status = status;
    if (remark != null) result.remark = remark;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  ReturnRequestEntry._();

  factory ReturnRequestEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReturnRequestEntry()..mergeFromBuffer(data, registry);
  factory ReturnRequestEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReturnRequestEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReturnRequestEntry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ReturnRequestEntry.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'customerId')
    ..aOS(3, _omitFieldNames ? '' : 'status')
    ..aOS(4, _omitFieldNames ? '' : 'remark')
    ..aOS(5, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReturnRequestEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReturnRequestEntry copyWith(void Function(ReturnRequestEntry) updates) =>
      super.copyWith((message) => updates(message as ReturnRequestEntry))
          as ReturnRequestEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ReturnRequestEntry() / ReturnRequestEntry.new instead')
  static ReturnRequestEntry create() => ReturnRequestEntry._();
  static $pb.GeneratedMessage $_createMessage() => ReturnRequestEntry._();
  @$core.override
  ReturnRequestEntry createEmptyInstance() => ReturnRequestEntry._();
  @$core.pragma('dart2js:noInline')
  static ReturnRequestEntry getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReturnRequestEntry>(
          ReturnRequestEntry.$_createMessage);
  static ReturnRequestEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get customerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set customerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCustomerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCustomerId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get status => $_getSZ(2);
  @$pb.TagNumber(3)
  set status($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get remark => $_getSZ(3);
  @$pb.TagNumber(4)
  set remark($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRemark() => $_has(3);
  @$pb.TagNumber(4)
  void clearRemark() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get createdAt => $_getSZ(4);
  @$pb.TagNumber(5)
  set createdAt($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCreatedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreatedAt() => $_clearField(5);
}

/// ListReturnRequestsResponse:列表結果。
class ListReturnRequestsResponse extends $pb.GeneratedMessage {
  factory ListReturnRequestsResponse({
    $core.Iterable<ReturnRequestEntry>? entries,
    $core.int? page,
    $core.int? pageSize,
    $core.int? total,
  }) {
    final result = ListReturnRequestsResponse._();
    if (entries != null) result.entries.addAll(entries);
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (total != null) result.total = total;
    return result;
  }

  ListReturnRequestsResponse._();

  factory ListReturnRequestsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReturnRequestsResponse()..mergeFromBuffer(data, registry);
  factory ListReturnRequestsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListReturnRequestsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListReturnRequestsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListReturnRequestsResponse.$_createMessage)
    ..pPM<ReturnRequestEntry>(1, _omitFieldNames ? '' : 'entries',
        subBuilder: ReturnRequestEntry.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'page')
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..aI(4, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReturnRequestsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListReturnRequestsResponse copyWith(
          void Function(ListReturnRequestsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListReturnRequestsResponse))
          as ListReturnRequestsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListReturnRequestsResponse() / ListReturnRequestsResponse.new instead')
  static ListReturnRequestsResponse create() => ListReturnRequestsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListReturnRequestsResponse._();
  @$core.override
  ListReturnRequestsResponse createEmptyInstance() =>
      ListReturnRequestsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListReturnRequestsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListReturnRequestsResponse>(
          ListReturnRequestsResponse.$_createMessage);
  static ListReturnRequestsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ReturnRequestEntry> get entries => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get page => $_getIZ(1);
  @$pb.TagNumber(2)
  set page($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPage() => $_has(1);
  @$pb.TagNumber(2)
  void clearPage() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get pageSize => $_getIZ(2);
  @$pb.TagNumber(3)
  set pageSize($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageSize() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get total => $_getIZ(3);
  @$pb.TagNumber(4)
  set total($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTotal() => $_has(3);
  @$pb.TagNumber(4)
  void clearTotal() => $_clearField(4);
}

/// ReturnRequestItemView:明細單筆(含快照與照片 URL)。
class ReturnRequestItemView extends $pb.GeneratedMessage {
  factory ReturnRequestItemView({
    $core.String? id,
    $core.String? sourceType,
    $core.String? productName,
    $core.String? spec,
    $core.String? unit,
    $core.String? quantity,
    $core.String? reason,
    $core.Iterable<$core.String>? photoUrls,
  }) {
    final result = ReturnRequestItemView._();
    if (id != null) result.id = id;
    if (sourceType != null) result.sourceType = sourceType;
    if (productName != null) result.productName = productName;
    if (spec != null) result.spec = spec;
    if (unit != null) result.unit = unit;
    if (quantity != null) result.quantity = quantity;
    if (reason != null) result.reason = reason;
    if (photoUrls != null) result.photoUrls.addAll(photoUrls);
    return result;
  }

  ReturnRequestItemView._();

  factory ReturnRequestItemView.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReturnRequestItemView()..mergeFromBuffer(data, registry);
  factory ReturnRequestItemView.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReturnRequestItemView()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReturnRequestItemView',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ReturnRequestItemView.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'sourceType')
    ..aOS(3, _omitFieldNames ? '' : 'productName')
    ..aOS(4, _omitFieldNames ? '' : 'spec')
    ..aOS(5, _omitFieldNames ? '' : 'unit')
    ..aOS(6, _omitFieldNames ? '' : 'quantity')
    ..aOS(7, _omitFieldNames ? '' : 'reason')
    ..pPS(8, _omitFieldNames ? '' : 'photoUrls')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReturnRequestItemView clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReturnRequestItemView copyWith(
          void Function(ReturnRequestItemView) updates) =>
      super.copyWith((message) => updates(message as ReturnRequestItemView))
          as ReturnRequestItemView;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReturnRequestItemView() / ReturnRequestItemView.new instead')
  static ReturnRequestItemView create() => ReturnRequestItemView._();
  static $pb.GeneratedMessage $_createMessage() => ReturnRequestItemView._();
  @$core.override
  ReturnRequestItemView createEmptyInstance() => ReturnRequestItemView._();
  @$core.pragma('dart2js:noInline')
  static ReturnRequestItemView getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReturnRequestItemView>(
          ReturnRequestItemView.$_createMessage);
  static ReturnRequestItemView? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get sourceType => $_getSZ(1);
  @$pb.TagNumber(2)
  set sourceType($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSourceType() => $_has(1);
  @$pb.TagNumber(2)
  void clearSourceType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get productName => $_getSZ(2);
  @$pb.TagNumber(3)
  set productName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProductName() => $_has(2);
  @$pb.TagNumber(3)
  void clearProductName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get spec => $_getSZ(3);
  @$pb.TagNumber(4)
  set spec($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSpec() => $_has(3);
  @$pb.TagNumber(4)
  void clearSpec() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get unit => $_getSZ(4);
  @$pb.TagNumber(5)
  set unit($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUnit() => $_has(4);
  @$pb.TagNumber(5)
  void clearUnit() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get quantity => $_getSZ(5);
  @$pb.TagNumber(6)
  set quantity($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasQuantity() => $_has(5);
  @$pb.TagNumber(6)
  void clearQuantity() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get reason => $_getSZ(6);
  @$pb.TagNumber(7)
  set reason($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasReason() => $_has(6);
  @$pb.TagNumber(7)
  void clearReason() => $_clearField(7);

  @$pb.TagNumber(8)
  $pb.PbList<$core.String> get photoUrls => $_getList(7);
}

/// GetReturnRequestRequest:單筆請求。
class GetReturnRequestRequest extends $pb.GeneratedMessage {
  factory GetReturnRequestRequest({
    $core.String? id,
  }) {
    final result = GetReturnRequestRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetReturnRequestRequest._();

  factory GetReturnRequestRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetReturnRequestRequest()..mergeFromBuffer(data, registry);
  factory GetReturnRequestRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetReturnRequestRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetReturnRequestRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetReturnRequestRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetReturnRequestRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetReturnRequestRequest copyWith(
          void Function(GetReturnRequestRequest) updates) =>
      super.copyWith((message) => updates(message as GetReturnRequestRequest))
          as GetReturnRequestRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetReturnRequestRequest() / GetReturnRequestRequest.new instead')
  static GetReturnRequestRequest create() => GetReturnRequestRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetReturnRequestRequest._();
  @$core.override
  GetReturnRequestRequest createEmptyInstance() => GetReturnRequestRequest._();
  @$core.pragma('dart2js:noInline')
  static GetReturnRequestRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetReturnRequestRequest>(
          GetReturnRequestRequest.$_createMessage);
  static GetReturnRequestRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// GetReturnRequestResponse:單筆結果。
class GetReturnRequestResponse extends $pb.GeneratedMessage {
  factory GetReturnRequestResponse({
    $core.String? id,
    $core.String? customerId,
    $core.String? status,
    $core.String? remark,
    $core.String? rejectReason,
    $core.String? createdAt,
    $core.Iterable<ReturnRequestItemView>? items,
  }) {
    final result = GetReturnRequestResponse._();
    if (id != null) result.id = id;
    if (customerId != null) result.customerId = customerId;
    if (status != null) result.status = status;
    if (remark != null) result.remark = remark;
    if (rejectReason != null) result.rejectReason = rejectReason;
    if (createdAt != null) result.createdAt = createdAt;
    if (items != null) result.items.addAll(items);
    return result;
  }

  GetReturnRequestResponse._();

  factory GetReturnRequestResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetReturnRequestResponse()..mergeFromBuffer(data, registry);
  factory GetReturnRequestResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetReturnRequestResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetReturnRequestResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetReturnRequestResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'customerId')
    ..aOS(3, _omitFieldNames ? '' : 'status')
    ..aOS(4, _omitFieldNames ? '' : 'remark')
    ..aOS(5, _omitFieldNames ? '' : 'rejectReason')
    ..aOS(6, _omitFieldNames ? '' : 'createdAt')
    ..pPM<ReturnRequestItemView>(7, _omitFieldNames ? '' : 'items',
        subBuilder: ReturnRequestItemView.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetReturnRequestResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetReturnRequestResponse copyWith(
          void Function(GetReturnRequestResponse) updates) =>
      super.copyWith((message) => updates(message as GetReturnRequestResponse))
          as GetReturnRequestResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetReturnRequestResponse() / GetReturnRequestResponse.new instead')
  static GetReturnRequestResponse create() => GetReturnRequestResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetReturnRequestResponse._();
  @$core.override
  GetReturnRequestResponse createEmptyInstance() =>
      GetReturnRequestResponse._();
  @$core.pragma('dart2js:noInline')
  static GetReturnRequestResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetReturnRequestResponse>(
          GetReturnRequestResponse.$_createMessage);
  static GetReturnRequestResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get customerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set customerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCustomerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCustomerId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get status => $_getSZ(2);
  @$pb.TagNumber(3)
  set status($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get remark => $_getSZ(3);
  @$pb.TagNumber(4)
  set remark($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRemark() => $_has(3);
  @$pb.TagNumber(4)
  void clearRemark() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get rejectReason => $_getSZ(4);
  @$pb.TagNumber(5)
  set rejectReason($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRejectReason() => $_has(4);
  @$pb.TagNumber(5)
  void clearRejectReason() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get createdAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set createdAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCreatedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCreatedAt() => $_clearField(6);

  @$pb.TagNumber(7)
  $pb.PbList<ReturnRequestItemView> get items => $_getList(6);
}

/// ReturnService:退貨申請與審核(06 計畫 Task 4.7.2–4.7.4, D25)。
/// 發起僅客戶子帳號(主帳號一律拒絕);審核僅主責業務/dept_admin 以上;
/// 全程不修改原訂單(僅參照)。
class ReturnServiceApi {
  final $pb.RpcClient _client;

  ReturnServiceApi(this._client);

  /// CreateReturnRequest:發起退貨(子帳號;雙來源並存;同交易寫申請+明細+稽核)。
  $async.Future<CreateReturnRequestResponse> createReturnRequest(
          $pb.ClientContext? ctx, CreateReturnRequestRequest request) =>
      _client.invoke<CreateReturnRequestResponse>(ctx, 'ReturnService',
          'CreateReturnRequest', request, CreateReturnRequestResponse());

  /// ListReturnRequests:申請列表(客戶 self 限自己客戶;staff 依範圍;分頁 per_page ≤ 100)。
  $async.Future<ListReturnRequestsResponse> listReturnRequests(
          $pb.ClientContext? ctx, ListReturnRequestsRequest request) =>
      _client.invoke<ListReturnRequestsResponse>(ctx, 'ReturnService',
          'ListReturnRequests', request, ListReturnRequestsResponse());

  /// GetReturnRequest:單筆含品項明細(照片轉下載 URL)。
  $async.Future<GetReturnRequestResponse> getReturnRequest(
          $pb.ClientContext? ctx, GetReturnRequestRequest request) =>
      _client.invoke<GetReturnRequestResponse>(ctx, 'ReturnService',
          'GetReturnRequest', request, GetReturnRequestResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
