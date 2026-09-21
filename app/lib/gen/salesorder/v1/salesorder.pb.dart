// This is a generated file - do not edit.
//
// Generated from salesorder/v1/salesorder.proto.

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

/// SalesOrder:訂單主檔。
class SalesOrder extends $pb.GeneratedMessage {
  factory SalesOrder({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? orderNo,
    $core.String? customerId,
    $core.String? source,
    $core.String? status,
    $core.String? expectedDeliveryDate,
    $core.String? salesRepId,
    $core.String? note,
    $core.String? dispatchedAt,
    $core.String? dispatchedBy,
    $core.String? routeId,
    $core.int? deliverySequence,
    $core.int? version,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = SalesOrder._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (orderNo != null) result.orderNo = orderNo;
    if (customerId != null) result.customerId = customerId;
    if (source != null) result.source = source;
    if (status != null) result.status = status;
    if (expectedDeliveryDate != null)
      result.expectedDeliveryDate = expectedDeliveryDate;
    if (salesRepId != null) result.salesRepId = salesRepId;
    if (note != null) result.note = note;
    if (dispatchedAt != null) result.dispatchedAt = dispatchedAt;
    if (dispatchedBy != null) result.dispatchedBy = dispatchedBy;
    if (routeId != null) result.routeId = routeId;
    if (deliverySequence != null) result.deliverySequence = deliverySequence;
    if (version != null) result.version = version;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  SalesOrder._();

  factory SalesOrder.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SalesOrder()..mergeFromBuffer(data, registry);
  factory SalesOrder.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SalesOrder()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SalesOrder',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: SalesOrder.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'orderNo')
    ..aOS(5, _omitFieldNames ? '' : 'customerId')
    ..aOS(6, _omitFieldNames ? '' : 'source')
    ..aOS(7, _omitFieldNames ? '' : 'status')
    ..aOS(8, _omitFieldNames ? '' : 'expectedDeliveryDate')
    ..aOS(9, _omitFieldNames ? '' : 'salesRepId')
    ..aOS(10, _omitFieldNames ? '' : 'note')
    ..aOS(11, _omitFieldNames ? '' : 'dispatchedAt')
    ..aOS(12, _omitFieldNames ? '' : 'dispatchedBy')
    ..aOS(13, _omitFieldNames ? '' : 'routeId')
    ..aI(14, _omitFieldNames ? '' : 'deliverySequence')
    ..aI(15, _omitFieldNames ? '' : 'version')
    ..aOS(16, _omitFieldNames ? '' : 'createdAt')
    ..aOS(17, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(18, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SalesOrder clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SalesOrder copyWith(void Function(SalesOrder) updates) =>
      super.copyWith((message) => updates(message as SalesOrder)) as SalesOrder;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SalesOrder() / SalesOrder.new instead')
  static SalesOrder create() => SalesOrder._();
  static $pb.GeneratedMessage $_createMessage() => SalesOrder._();
  @$core.override
  SalesOrder createEmptyInstance() => SalesOrder._();
  @$core.pragma('dart2js:noInline')
  static SalesOrder getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SalesOrder>(SalesOrder.$_createMessage);
  static SalesOrder? _defaultInstance;

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
  $core.String get orderNo => $_getSZ(3);
  @$pb.TagNumber(4)
  set orderNo($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasOrderNo() => $_has(3);
  @$pb.TagNumber(4)
  void clearOrderNo() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get customerId => $_getSZ(4);
  @$pb.TagNumber(5)
  set customerId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCustomerId() => $_has(4);
  @$pb.TagNumber(5)
  void clearCustomerId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get source => $_getSZ(5);
  @$pb.TagNumber(6)
  set source($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSource() => $_has(5);
  @$pb.TagNumber(6)
  void clearSource() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get status => $_getSZ(6);
  @$pb.TagNumber(7)
  set status($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasStatus() => $_has(6);
  @$pb.TagNumber(7)
  void clearStatus() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get expectedDeliveryDate => $_getSZ(7);
  @$pb.TagNumber(8)
  set expectedDeliveryDate($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasExpectedDeliveryDate() => $_has(7);
  @$pb.TagNumber(8)
  void clearExpectedDeliveryDate() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get salesRepId => $_getSZ(8);
  @$pb.TagNumber(9)
  set salesRepId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasSalesRepId() => $_has(8);
  @$pb.TagNumber(9)
  void clearSalesRepId() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get note => $_getSZ(9);
  @$pb.TagNumber(10)
  set note($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasNote() => $_has(9);
  @$pb.TagNumber(10)
  void clearNote() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get dispatchedAt => $_getSZ(10);
  @$pb.TagNumber(11)
  set dispatchedAt($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasDispatchedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearDispatchedAt() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get dispatchedBy => $_getSZ(11);
  @$pb.TagNumber(12)
  set dispatchedBy($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasDispatchedBy() => $_has(11);
  @$pb.TagNumber(12)
  void clearDispatchedBy() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get routeId => $_getSZ(12);
  @$pb.TagNumber(13)
  set routeId($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasRouteId() => $_has(12);
  @$pb.TagNumber(13)
  void clearRouteId() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.int get deliverySequence => $_getIZ(13);
  @$pb.TagNumber(14)
  set deliverySequence($core.int value) => $_setSignedInt32(13, value);
  @$pb.TagNumber(14)
  $core.bool hasDeliverySequence() => $_has(13);
  @$pb.TagNumber(14)
  void clearDeliverySequence() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.int get version => $_getIZ(14);
  @$pb.TagNumber(15)
  set version($core.int value) => $_setSignedInt32(14, value);
  @$pb.TagNumber(15)
  $core.bool hasVersion() => $_has(14);
  @$pb.TagNumber(15)
  void clearVersion() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.String get createdAt => $_getSZ(15);
  @$pb.TagNumber(16)
  set createdAt($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasCreatedAt() => $_has(15);
  @$pb.TagNumber(16)
  void clearCreatedAt() => $_clearField(16);

  @$pb.TagNumber(17)
  $core.String get updatedAt => $_getSZ(16);
  @$pb.TagNumber(17)
  set updatedAt($core.String value) => $_setString(16, value);
  @$pb.TagNumber(17)
  $core.bool hasUpdatedAt() => $_has(16);
  @$pb.TagNumber(17)
  void clearUpdatedAt() => $_clearField(17);

  @$pb.TagNumber(18)
  $core.String get deletedAt => $_getSZ(17);
  @$pb.TagNumber(18)
  set deletedAt($core.String value) => $_setString(17, value);
  @$pb.TagNumber(18)
  $core.bool hasDeletedAt() => $_has(17);
  @$pb.TagNumber(18)
  void clearDeletedAt() => $_clearField(18);
}

/// SalesOrderItem:訂單明細(不含金額)。
class SalesOrderItem extends $pb.GeneratedMessage {
  factory SalesOrderItem({
    $core.String? id,
    $core.String? productId,
    $core.String? displayName,
    $core.String? qty,
    $core.String? unit,
    $core.String? baseQty,
    $core.String? processingSpecId,
    $core.String? specialCutNote,
    $core.String? warehouseId,
    $core.int? sortOrder,
  }) {
    final result = SalesOrderItem._();
    if (id != null) result.id = id;
    if (productId != null) result.productId = productId;
    if (displayName != null) result.displayName = displayName;
    if (qty != null) result.qty = qty;
    if (unit != null) result.unit = unit;
    if (baseQty != null) result.baseQty = baseQty;
    if (processingSpecId != null) result.processingSpecId = processingSpecId;
    if (specialCutNote != null) result.specialCutNote = specialCutNote;
    if (warehouseId != null) result.warehouseId = warehouseId;
    if (sortOrder != null) result.sortOrder = sortOrder;
    return result;
  }

  SalesOrderItem._();

  factory SalesOrderItem.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SalesOrderItem()..mergeFromBuffer(data, registry);
  factory SalesOrderItem.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SalesOrderItem()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SalesOrderItem',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: SalesOrderItem.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'productId')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'qty')
    ..aOS(5, _omitFieldNames ? '' : 'unit')
    ..aOS(6, _omitFieldNames ? '' : 'baseQty')
    ..aOS(7, _omitFieldNames ? '' : 'processingSpecId')
    ..aOS(8, _omitFieldNames ? '' : 'specialCutNote')
    ..aOS(9, _omitFieldNames ? '' : 'warehouseId')
    ..aI(10, _omitFieldNames ? '' : 'sortOrder')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SalesOrderItem clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SalesOrderItem copyWith(void Function(SalesOrderItem) updates) =>
      super.copyWith((message) => updates(message as SalesOrderItem))
          as SalesOrderItem;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SalesOrderItem() / SalesOrderItem.new instead')
  static SalesOrderItem create() => SalesOrderItem._();
  static $pb.GeneratedMessage $_createMessage() => SalesOrderItem._();
  @$core.override
  SalesOrderItem createEmptyInstance() => SalesOrderItem._();
  @$core.pragma('dart2js:noInline')
  static SalesOrderItem getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SalesOrderItem>(
          SalesOrderItem.$_createMessage);
  static SalesOrderItem? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get productId => $_getSZ(1);
  @$pb.TagNumber(2)
  set productId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProductId() => $_has(1);
  @$pb.TagNumber(2)
  void clearProductId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get qty => $_getSZ(3);
  @$pb.TagNumber(4)
  set qty($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasQty() => $_has(3);
  @$pb.TagNumber(4)
  void clearQty() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get unit => $_getSZ(4);
  @$pb.TagNumber(5)
  set unit($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUnit() => $_has(4);
  @$pb.TagNumber(5)
  void clearUnit() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get baseQty => $_getSZ(5);
  @$pb.TagNumber(6)
  set baseQty($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasBaseQty() => $_has(5);
  @$pb.TagNumber(6)
  void clearBaseQty() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get processingSpecId => $_getSZ(6);
  @$pb.TagNumber(7)
  set processingSpecId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasProcessingSpecId() => $_has(6);
  @$pb.TagNumber(7)
  void clearProcessingSpecId() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get specialCutNote => $_getSZ(7);
  @$pb.TagNumber(8)
  set specialCutNote($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasSpecialCutNote() => $_has(7);
  @$pb.TagNumber(8)
  void clearSpecialCutNote() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get warehouseId => $_getSZ(8);
  @$pb.TagNumber(9)
  set warehouseId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasWarehouseId() => $_has(8);
  @$pb.TagNumber(9)
  void clearWarehouseId() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.int get sortOrder => $_getIZ(9);
  @$pb.TagNumber(10)
  set sortOrder($core.int value) => $_setSignedInt32(9, value);
  @$pb.TagNumber(10)
  $core.bool hasSortOrder() => $_has(9);
  @$pb.TagNumber(10)
  void clearSortOrder() => $_clearField(10);
}

/// SalesOrderEvent:訂單異動事件(僅追加)。
class SalesOrderEvent extends $pb.GeneratedMessage {
  factory SalesOrderEvent({
    $core.String? id,
    $core.String? eventType,
    $core.String? actorId,
    $core.String? reason,
    $core.String? payload,
    $core.String? createdAt,
  }) {
    final result = SalesOrderEvent._();
    if (id != null) result.id = id;
    if (eventType != null) result.eventType = eventType;
    if (actorId != null) result.actorId = actorId;
    if (reason != null) result.reason = reason;
    if (payload != null) result.payload = payload;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  SalesOrderEvent._();

  factory SalesOrderEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SalesOrderEvent()..mergeFromBuffer(data, registry);
  factory SalesOrderEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SalesOrderEvent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SalesOrderEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: SalesOrderEvent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'eventType')
    ..aOS(3, _omitFieldNames ? '' : 'actorId')
    ..aOS(4, _omitFieldNames ? '' : 'reason')
    ..aOS(5, _omitFieldNames ? '' : 'payload')
    ..aOS(6, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SalesOrderEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SalesOrderEvent copyWith(void Function(SalesOrderEvent) updates) =>
      super.copyWith((message) => updates(message as SalesOrderEvent))
          as SalesOrderEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SalesOrderEvent() / SalesOrderEvent.new instead')
  static SalesOrderEvent create() => SalesOrderEvent._();
  static $pb.GeneratedMessage $_createMessage() => SalesOrderEvent._();
  @$core.override
  SalesOrderEvent createEmptyInstance() => SalesOrderEvent._();
  @$core.pragma('dart2js:noInline')
  static SalesOrderEvent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SalesOrderEvent>(
          SalesOrderEvent.$_createMessage);
  static SalesOrderEvent? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get eventType => $_getSZ(1);
  @$pb.TagNumber(2)
  set eventType($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEventType() => $_has(1);
  @$pb.TagNumber(2)
  void clearEventType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get actorId => $_getSZ(2);
  @$pb.TagNumber(3)
  set actorId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasActorId() => $_has(2);
  @$pb.TagNumber(3)
  void clearActorId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get reason => $_getSZ(3);
  @$pb.TagNumber(4)
  set reason($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasReason() => $_has(3);
  @$pb.TagNumber(4)
  void clearReason() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get payload => $_getSZ(4);
  @$pb.TagNumber(5)
  set payload($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPayload() => $_has(4);
  @$pb.TagNumber(5)
  void clearPayload() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get createdAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set createdAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCreatedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCreatedAt() => $_clearField(6);
}

class ListOrdersRequest extends $pb.GeneratedMessage {
  factory ListOrdersRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? status,
    $core.String? customerId,
    $core.String? source,
    $core.String? keyword,
    $core.bool? includeDeleted,
  }) {
    final result = ListOrdersRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (status != null) result.status = status;
    if (customerId != null) result.customerId = customerId;
    if (source != null) result.source = source;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListOrdersRequest._();

  factory ListOrdersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrdersRequest()..mergeFromBuffer(data, registry);
  factory ListOrdersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrdersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOrdersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListOrdersRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'status')
    ..aOS(4, _omitFieldNames ? '' : 'customerId')
    ..aOS(5, _omitFieldNames ? '' : 'source')
    ..aOS(6, _omitFieldNames ? '' : 'keyword')
    ..aOB(7, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrdersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrdersRequest copyWith(void Function(ListOrdersRequest) updates) =>
      super.copyWith((message) => updates(message as ListOrdersRequest))
          as ListOrdersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListOrdersRequest() / ListOrdersRequest.new instead')
  static ListOrdersRequest create() => ListOrdersRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListOrdersRequest._();
  @$core.override
  ListOrdersRequest createEmptyInstance() => ListOrdersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListOrdersRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListOrdersRequest>(
          ListOrdersRequest.$_createMessage);
  static ListOrdersRequest? _defaultInstance;

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
  $core.String get status => $_getSZ(2);
  @$pb.TagNumber(3)
  set status($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get customerId => $_getSZ(3);
  @$pb.TagNumber(4)
  set customerId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCustomerId() => $_has(3);
  @$pb.TagNumber(4)
  void clearCustomerId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get source => $_getSZ(4);
  @$pb.TagNumber(5)
  set source($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSource() => $_has(4);
  @$pb.TagNumber(5)
  void clearSource() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get keyword => $_getSZ(5);
  @$pb.TagNumber(6)
  set keyword($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasKeyword() => $_has(5);
  @$pb.TagNumber(6)
  void clearKeyword() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get includeDeleted => $_getBF(6);
  @$pb.TagNumber(7)
  set includeDeleted($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasIncludeDeleted() => $_has(6);
  @$pb.TagNumber(7)
  void clearIncludeDeleted() => $_clearField(7);
}

class ListOrdersResponse extends $pb.GeneratedMessage {
  factory ListOrdersResponse({
    $core.Iterable<SalesOrder>? orders,
    $core.int? total,
  }) {
    final result = ListOrdersResponse._();
    if (orders != null) result.orders.addAll(orders);
    if (total != null) result.total = total;
    return result;
  }

  ListOrdersResponse._();

  factory ListOrdersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrdersResponse()..mergeFromBuffer(data, registry);
  factory ListOrdersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrdersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOrdersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListOrdersResponse.$_createMessage)
    ..pPM<SalesOrder>(1, _omitFieldNames ? '' : 'orders',
        subBuilder: SalesOrder.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrdersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrdersResponse copyWith(void Function(ListOrdersResponse) updates) =>
      super.copyWith((message) => updates(message as ListOrdersResponse))
          as ListOrdersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListOrdersResponse() / ListOrdersResponse.new instead')
  static ListOrdersResponse create() => ListOrdersResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListOrdersResponse._();
  @$core.override
  ListOrdersResponse createEmptyInstance() => ListOrdersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListOrdersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOrdersResponse>(
          ListOrdersResponse.$_createMessage);
  static ListOrdersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<SalesOrder> get orders => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get total => $_getIZ(1);
  @$pb.TagNumber(2)
  set total($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => $_clearField(2);
}

class GetOrderRequest extends $pb.GeneratedMessage {
  factory GetOrderRequest({
    $core.String? id,
  }) {
    final result = GetOrderRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetOrderRequest._();

  factory GetOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOrderRequest()..mergeFromBuffer(data, registry);
  factory GetOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOrderRequest copyWith(void Function(GetOrderRequest) updates) =>
      super.copyWith((message) => updates(message as GetOrderRequest))
          as GetOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetOrderRequest() / GetOrderRequest.new instead')
  static GetOrderRequest create() => GetOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetOrderRequest._();
  @$core.override
  GetOrderRequest createEmptyInstance() => GetOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static GetOrderRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetOrderRequest>(
          GetOrderRequest.$_createMessage);
  static GetOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetOrderResponse extends $pb.GeneratedMessage {
  factory GetOrderResponse({
    SalesOrder? order,
    $core.Iterable<SalesOrderItem>? items,
  }) {
    final result = GetOrderResponse._();
    if (order != null) result.order = order;
    if (items != null) result.items.addAll(items);
    return result;
  }

  GetOrderResponse._();

  factory GetOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOrderResponse()..mergeFromBuffer(data, registry);
  factory GetOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetOrderResponse.$_createMessage)
    ..aOM<SalesOrder>(1, _omitFieldNames ? '' : 'order',
        subBuilder: SalesOrder.$_createMessage)
    ..pPM<SalesOrderItem>(2, _omitFieldNames ? '' : 'items',
        subBuilder: SalesOrderItem.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOrderResponse copyWith(void Function(GetOrderResponse) updates) =>
      super.copyWith((message) => updates(message as GetOrderResponse))
          as GetOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetOrderResponse() / GetOrderResponse.new instead')
  static GetOrderResponse create() => GetOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetOrderResponse._();
  @$core.override
  GetOrderResponse createEmptyInstance() => GetOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static GetOrderResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetOrderResponse>(
          GetOrderResponse.$_createMessage);
  static GetOrderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SalesOrder get order => $_getN(0);
  @$pb.TagNumber(1)
  set order(SalesOrder value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOrder() => $_has(0);
  @$pb.TagNumber(1)
  void clearOrder() => $_clearField(1);
  @$pb.TagNumber(1)
  SalesOrder ensureOrder() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<SalesOrderItem> get items => $_getList(1);
}

class OrderItemInput extends $pb.GeneratedMessage {
  factory OrderItemInput({
    $core.String? productId,
    $core.String? manualName,
    $core.String? displayName,
    $core.String? qty,
    $core.String? unit,
    $core.String? processingSpecId,
    $core.String? specialCutNote,
    $core.String? warehouseId,
    $core.bool? saveAlias,
  }) {
    final result = OrderItemInput._();
    if (productId != null) result.productId = productId;
    if (manualName != null) result.manualName = manualName;
    if (displayName != null) result.displayName = displayName;
    if (qty != null) result.qty = qty;
    if (unit != null) result.unit = unit;
    if (processingSpecId != null) result.processingSpecId = processingSpecId;
    if (specialCutNote != null) result.specialCutNote = specialCutNote;
    if (warehouseId != null) result.warehouseId = warehouseId;
    if (saveAlias != null) result.saveAlias = saveAlias;
    return result;
  }

  OrderItemInput._();

  factory OrderItemInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OrderItemInput()..mergeFromBuffer(data, registry);
  factory OrderItemInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OrderItemInput()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OrderItemInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: OrderItemInput.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'productId')
    ..aOS(2, _omitFieldNames ? '' : 'manualName')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'qty')
    ..aOS(5, _omitFieldNames ? '' : 'unit')
    ..aOS(6, _omitFieldNames ? '' : 'processingSpecId')
    ..aOS(7, _omitFieldNames ? '' : 'specialCutNote')
    ..aOS(8, _omitFieldNames ? '' : 'warehouseId')
    ..aOB(9, _omitFieldNames ? '' : 'saveAlias')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OrderItemInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OrderItemInput copyWith(void Function(OrderItemInput) updates) =>
      super.copyWith((message) => updates(message as OrderItemInput))
          as OrderItemInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use OrderItemInput() / OrderItemInput.new instead')
  static OrderItemInput create() => OrderItemInput._();
  static $pb.GeneratedMessage $_createMessage() => OrderItemInput._();
  @$core.override
  OrderItemInput createEmptyInstance() => OrderItemInput._();
  @$core.pragma('dart2js:noInline')
  static OrderItemInput getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<OrderItemInput>(
          OrderItemInput.$_createMessage);
  static OrderItemInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get productId => $_getSZ(0);
  @$pb.TagNumber(1)
  set productId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProductId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProductId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get manualName => $_getSZ(1);
  @$pb.TagNumber(2)
  set manualName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasManualName() => $_has(1);
  @$pb.TagNumber(2)
  void clearManualName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get qty => $_getSZ(3);
  @$pb.TagNumber(4)
  set qty($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasQty() => $_has(3);
  @$pb.TagNumber(4)
  void clearQty() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get unit => $_getSZ(4);
  @$pb.TagNumber(5)
  set unit($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUnit() => $_has(4);
  @$pb.TagNumber(5)
  void clearUnit() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get processingSpecId => $_getSZ(5);
  @$pb.TagNumber(6)
  set processingSpecId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasProcessingSpecId() => $_has(5);
  @$pb.TagNumber(6)
  void clearProcessingSpecId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get specialCutNote => $_getSZ(6);
  @$pb.TagNumber(7)
  set specialCutNote($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasSpecialCutNote() => $_has(6);
  @$pb.TagNumber(7)
  void clearSpecialCutNote() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get warehouseId => $_getSZ(7);
  @$pb.TagNumber(8)
  set warehouseId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasWarehouseId() => $_has(7);
  @$pb.TagNumber(8)
  void clearWarehouseId() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.bool get saveAlias => $_getBF(8);
  @$pb.TagNumber(9)
  set saveAlias($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasSaveAlias() => $_has(8);
  @$pb.TagNumber(9)
  void clearSaveAlias() => $_clearField(9);
}

class CreateOrderRequest extends $pb.GeneratedMessage {
  factory CreateOrderRequest({
    $core.String? customerId,
    $core.String? source,
    $core.String? expectedDeliveryDate,
    $core.String? note,
    $core.String? salesRepId,
    $core.Iterable<OrderItemInput>? items,
  }) {
    final result = CreateOrderRequest._();
    if (customerId != null) result.customerId = customerId;
    if (source != null) result.source = source;
    if (expectedDeliveryDate != null)
      result.expectedDeliveryDate = expectedDeliveryDate;
    if (note != null) result.note = note;
    if (salesRepId != null) result.salesRepId = salesRepId;
    if (items != null) result.items.addAll(items);
    return result;
  }

  CreateOrderRequest._();

  factory CreateOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOrderRequest()..mergeFromBuffer(data, registry);
  factory CreateOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOS(2, _omitFieldNames ? '' : 'source')
    ..aOS(3, _omitFieldNames ? '' : 'expectedDeliveryDate')
    ..aOS(4, _omitFieldNames ? '' : 'note')
    ..aOS(5, _omitFieldNames ? '' : 'salesRepId')
    ..pPM<OrderItemInput>(6, _omitFieldNames ? '' : 'items',
        subBuilder: OrderItemInput.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOrderRequest copyWith(void Function(CreateOrderRequest) updates) =>
      super.copyWith((message) => updates(message as CreateOrderRequest))
          as CreateOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateOrderRequest() / CreateOrderRequest.new instead')
  static CreateOrderRequest create() => CreateOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateOrderRequest._();
  @$core.override
  CreateOrderRequest createEmptyInstance() => CreateOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateOrderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateOrderRequest>(
          CreateOrderRequest.$_createMessage);
  static CreateOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get source => $_getSZ(1);
  @$pb.TagNumber(2)
  set source($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSource() => $_has(1);
  @$pb.TagNumber(2)
  void clearSource() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get expectedDeliveryDate => $_getSZ(2);
  @$pb.TagNumber(3)
  set expectedDeliveryDate($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpectedDeliveryDate() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpectedDeliveryDate() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get note => $_getSZ(3);
  @$pb.TagNumber(4)
  set note($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasNote() => $_has(3);
  @$pb.TagNumber(4)
  void clearNote() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get salesRepId => $_getSZ(4);
  @$pb.TagNumber(5)
  set salesRepId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSalesRepId() => $_has(4);
  @$pb.TagNumber(5)
  void clearSalesRepId() => $_clearField(5);

  @$pb.TagNumber(6)
  $pb.PbList<OrderItemInput> get items => $_getList(5);
}

class CreateOrderResponse extends $pb.GeneratedMessage {
  factory CreateOrderResponse({
    SalesOrder? order,
  }) {
    final result = CreateOrderResponse._();
    if (order != null) result.order = order;
    return result;
  }

  CreateOrderResponse._();

  factory CreateOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOrderResponse()..mergeFromBuffer(data, registry);
  factory CreateOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateOrderResponse.$_createMessage)
    ..aOM<SalesOrder>(1, _omitFieldNames ? '' : 'order',
        subBuilder: SalesOrder.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateOrderResponse copyWith(void Function(CreateOrderResponse) updates) =>
      super.copyWith((message) => updates(message as CreateOrderResponse))
          as CreateOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CreateOrderResponse() / CreateOrderResponse.new instead')
  static CreateOrderResponse create() => CreateOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateOrderResponse._();
  @$core.override
  CreateOrderResponse createEmptyInstance() => CreateOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateOrderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateOrderResponse>(
          CreateOrderResponse.$_createMessage);
  static CreateOrderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SalesOrder get order => $_getN(0);
  @$pb.TagNumber(1)
  set order(SalesOrder value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOrder() => $_has(0);
  @$pb.TagNumber(1)
  void clearOrder() => $_clearField(1);
  @$pb.TagNumber(1)
  SalesOrder ensureOrder() => $_ensure(0);
}

class UpdateOrderRequest extends $pb.GeneratedMessage {
  factory UpdateOrderRequest({
    $core.String? id,
    $core.int? version,
    $core.String? expectedDeliveryDate,
    $core.String? note,
    $core.Iterable<OrderItemInput>? items,
  }) {
    final result = UpdateOrderRequest._();
    if (id != null) result.id = id;
    if (version != null) result.version = version;
    if (expectedDeliveryDate != null)
      result.expectedDeliveryDate = expectedDeliveryDate;
    if (note != null) result.note = note;
    if (items != null) result.items.addAll(items);
    return result;
  }

  UpdateOrderRequest._();

  factory UpdateOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateOrderRequest()..mergeFromBuffer(data, registry);
  factory UpdateOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aI(2, _omitFieldNames ? '' : 'version')
    ..aOS(3, _omitFieldNames ? '' : 'expectedDeliveryDate')
    ..aOS(4, _omitFieldNames ? '' : 'note')
    ..pPM<OrderItemInput>(5, _omitFieldNames ? '' : 'items',
        subBuilder: OrderItemInput.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateOrderRequest copyWith(void Function(UpdateOrderRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateOrderRequest))
          as UpdateOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UpdateOrderRequest() / UpdateOrderRequest.new instead')
  static UpdateOrderRequest create() => UpdateOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateOrderRequest._();
  @$core.override
  UpdateOrderRequest createEmptyInstance() => UpdateOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateOrderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateOrderRequest>(
          UpdateOrderRequest.$_createMessage);
  static UpdateOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get version => $_getIZ(1);
  @$pb.TagNumber(2)
  set version($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearVersion() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get expectedDeliveryDate => $_getSZ(2);
  @$pb.TagNumber(3)
  set expectedDeliveryDate($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpectedDeliveryDate() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpectedDeliveryDate() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get note => $_getSZ(3);
  @$pb.TagNumber(4)
  set note($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasNote() => $_has(3);
  @$pb.TagNumber(4)
  void clearNote() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<OrderItemInput> get items => $_getList(4);
}

class UpdateOrderResponse extends $pb.GeneratedMessage {
  factory UpdateOrderResponse({
    SalesOrder? order,
  }) {
    final result = UpdateOrderResponse._();
    if (order != null) result.order = order;
    return result;
  }

  UpdateOrderResponse._();

  factory UpdateOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateOrderResponse()..mergeFromBuffer(data, registry);
  factory UpdateOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateOrderResponse.$_createMessage)
    ..aOM<SalesOrder>(1, _omitFieldNames ? '' : 'order',
        subBuilder: SalesOrder.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateOrderResponse copyWith(void Function(UpdateOrderResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateOrderResponse))
          as UpdateOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use UpdateOrderResponse() / UpdateOrderResponse.new instead')
  static UpdateOrderResponse create() => UpdateOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateOrderResponse._();
  @$core.override
  UpdateOrderResponse createEmptyInstance() => UpdateOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateOrderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateOrderResponse>(
          UpdateOrderResponse.$_createMessage);
  static UpdateOrderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SalesOrder get order => $_getN(0);
  @$pb.TagNumber(1)
  set order(SalesOrder value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOrder() => $_has(0);
  @$pb.TagNumber(1)
  void clearOrder() => $_clearField(1);
  @$pb.TagNumber(1)
  SalesOrder ensureOrder() => $_ensure(0);
}

class CancelOrderRequest extends $pb.GeneratedMessage {
  factory CancelOrderRequest({
    $core.String? id,
  }) {
    final result = CancelOrderRequest._();
    if (id != null) result.id = id;
    return result;
  }

  CancelOrderRequest._();

  factory CancelOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelOrderRequest()..mergeFromBuffer(data, registry);
  factory CancelOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CancelOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelOrderRequest copyWith(void Function(CancelOrderRequest) updates) =>
      super.copyWith((message) => updates(message as CancelOrderRequest))
          as CancelOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CancelOrderRequest() / CancelOrderRequest.new instead')
  static CancelOrderRequest create() => CancelOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => CancelOrderRequest._();
  @$core.override
  CancelOrderRequest createEmptyInstance() => CancelOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelOrderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelOrderRequest>(
          CancelOrderRequest.$_createMessage);
  static CancelOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class CancelOrderResponse extends $pb.GeneratedMessage {
  factory CancelOrderResponse({
    SalesOrder? order,
  }) {
    final result = CancelOrderResponse._();
    if (order != null) result.order = order;
    return result;
  }

  CancelOrderResponse._();

  factory CancelOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelOrderResponse()..mergeFromBuffer(data, registry);
  factory CancelOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CancelOrderResponse.$_createMessage)
    ..aOM<SalesOrder>(1, _omitFieldNames ? '' : 'order',
        subBuilder: SalesOrder.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelOrderResponse copyWith(void Function(CancelOrderResponse) updates) =>
      super.copyWith((message) => updates(message as CancelOrderResponse))
          as CancelOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CancelOrderResponse() / CancelOrderResponse.new instead')
  static CancelOrderResponse create() => CancelOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => CancelOrderResponse._();
  @$core.override
  CancelOrderResponse createEmptyInstance() => CancelOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelOrderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelOrderResponse>(
          CancelOrderResponse.$_createMessage);
  static CancelOrderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SalesOrder get order => $_getN(0);
  @$pb.TagNumber(1)
  set order(SalesOrder value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOrder() => $_has(0);
  @$pb.TagNumber(1)
  void clearOrder() => $_clearField(1);
  @$pb.TagNumber(1)
  SalesOrder ensureOrder() => $_ensure(0);
}

class CompleteOrderRequest extends $pb.GeneratedMessage {
  factory CompleteOrderRequest({
    $core.String? id,
  }) {
    final result = CompleteOrderRequest._();
    if (id != null) result.id = id;
    return result;
  }

  CompleteOrderRequest._();

  factory CompleteOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOrderRequest()..mergeFromBuffer(data, registry);
  factory CompleteOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CompleteOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOrderRequest copyWith(void Function(CompleteOrderRequest) updates) =>
      super.copyWith((message) => updates(message as CompleteOrderRequest))
          as CompleteOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteOrderRequest() / CompleteOrderRequest.new instead')
  static CompleteOrderRequest create() => CompleteOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => CompleteOrderRequest._();
  @$core.override
  CompleteOrderRequest createEmptyInstance() => CompleteOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static CompleteOrderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteOrderRequest>(
          CompleteOrderRequest.$_createMessage);
  static CompleteOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class CompleteOrderResponse extends $pb.GeneratedMessage {
  factory CompleteOrderResponse({
    SalesOrder? order,
  }) {
    final result = CompleteOrderResponse._();
    if (order != null) result.order = order;
    return result;
  }

  CompleteOrderResponse._();

  factory CompleteOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOrderResponse()..mergeFromBuffer(data, registry);
  factory CompleteOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CompleteOrderResponse.$_createMessage)
    ..aOM<SalesOrder>(1, _omitFieldNames ? '' : 'order',
        subBuilder: SalesOrder.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOrderResponse copyWith(
          void Function(CompleteOrderResponse) updates) =>
      super.copyWith((message) => updates(message as CompleteOrderResponse))
          as CompleteOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteOrderResponse() / CompleteOrderResponse.new instead')
  static CompleteOrderResponse create() => CompleteOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => CompleteOrderResponse._();
  @$core.override
  CompleteOrderResponse createEmptyInstance() => CompleteOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static CompleteOrderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteOrderResponse>(
          CompleteOrderResponse.$_createMessage);
  static CompleteOrderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SalesOrder get order => $_getN(0);
  @$pb.TagNumber(1)
  set order(SalesOrder value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOrder() => $_has(0);
  @$pb.TagNumber(1)
  void clearOrder() => $_clearField(1);
  @$pb.TagNumber(1)
  SalesOrder ensureOrder() => $_ensure(0);
}

class VoidOrderRequest extends $pb.GeneratedMessage {
  factory VoidOrderRequest({
    $core.String? id,
    $core.String? reason,
  }) {
    final result = VoidOrderRequest._();
    if (id != null) result.id = id;
    if (reason != null) result.reason = reason;
    return result;
  }

  VoidOrderRequest._();

  factory VoidOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidOrderRequest()..mergeFromBuffer(data, registry);
  factory VoidOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VoidOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: VoidOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidOrderRequest copyWith(void Function(VoidOrderRequest) updates) =>
      super.copyWith((message) => updates(message as VoidOrderRequest))
          as VoidOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use VoidOrderRequest() / VoidOrderRequest.new instead')
  static VoidOrderRequest create() => VoidOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => VoidOrderRequest._();
  @$core.override
  VoidOrderRequest createEmptyInstance() => VoidOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static VoidOrderRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoidOrderRequest>(
          VoidOrderRequest.$_createMessage);
  static VoidOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get reason => $_getSZ(1);
  @$pb.TagNumber(2)
  set reason($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearReason() => $_clearField(2);
}

class VoidOrderResponse extends $pb.GeneratedMessage {
  factory VoidOrderResponse({
    SalesOrder? order,
  }) {
    final result = VoidOrderResponse._();
    if (order != null) result.order = order;
    return result;
  }

  VoidOrderResponse._();

  factory VoidOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidOrderResponse()..mergeFromBuffer(data, registry);
  factory VoidOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VoidOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VoidOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: VoidOrderResponse.$_createMessage)
    ..aOM<SalesOrder>(1, _omitFieldNames ? '' : 'order',
        subBuilder: SalesOrder.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VoidOrderResponse copyWith(void Function(VoidOrderResponse) updates) =>
      super.copyWith((message) => updates(message as VoidOrderResponse))
          as VoidOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use VoidOrderResponse() / VoidOrderResponse.new instead')
  static VoidOrderResponse create() => VoidOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => VoidOrderResponse._();
  @$core.override
  VoidOrderResponse createEmptyInstance() => VoidOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static VoidOrderResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoidOrderResponse>(
          VoidOrderResponse.$_createMessage);
  static VoidOrderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SalesOrder get order => $_getN(0);
  @$pb.TagNumber(1)
  set order(SalesOrder value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOrder() => $_has(0);
  @$pb.TagNumber(1)
  void clearOrder() => $_clearField(1);
  @$pb.TagNumber(1)
  SalesOrder ensureOrder() => $_ensure(0);
}

class DeleteOrderRequest extends $pb.GeneratedMessage {
  factory DeleteOrderRequest({
    $core.String? id,
  }) {
    final result = DeleteOrderRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteOrderRequest._();

  factory DeleteOrderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteOrderRequest()..mergeFromBuffer(data, registry);
  factory DeleteOrderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteOrderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteOrderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteOrderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteOrderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteOrderRequest copyWith(void Function(DeleteOrderRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteOrderRequest))
          as DeleteOrderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeleteOrderRequest() / DeleteOrderRequest.new instead')
  static DeleteOrderRequest create() => DeleteOrderRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteOrderRequest._();
  @$core.override
  DeleteOrderRequest createEmptyInstance() => DeleteOrderRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteOrderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteOrderRequest>(
          DeleteOrderRequest.$_createMessage);
  static DeleteOrderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteOrderResponse extends $pb.GeneratedMessage {
  factory DeleteOrderResponse() => DeleteOrderResponse._();

  DeleteOrderResponse._();

  factory DeleteOrderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteOrderResponse()..mergeFromBuffer(data, registry);
  factory DeleteOrderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteOrderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteOrderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteOrderResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteOrderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteOrderResponse copyWith(void Function(DeleteOrderResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteOrderResponse))
          as DeleteOrderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DeleteOrderResponse() / DeleteOrderResponse.new instead')
  static DeleteOrderResponse create() => DeleteOrderResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteOrderResponse._();
  @$core.override
  DeleteOrderResponse createEmptyInstance() => DeleteOrderResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteOrderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteOrderResponse>(
          DeleteOrderResponse.$_createMessage);
  static DeleteOrderResponse? _defaultInstance;
}

class ListOrderEventsRequest extends $pb.GeneratedMessage {
  factory ListOrderEventsRequest({
    $core.String? salesOrderId,
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListOrderEventsRequest._();
    if (salesOrderId != null) result.salesOrderId = salesOrderId;
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListOrderEventsRequest._();

  factory ListOrderEventsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrderEventsRequest()..mergeFromBuffer(data, registry);
  factory ListOrderEventsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrderEventsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOrderEventsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListOrderEventsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'salesOrderId')
    ..aI(2, _omitFieldNames ? '' : 'page')
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrderEventsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrderEventsRequest copyWith(
          void Function(ListOrderEventsRequest) updates) =>
      super.copyWith((message) => updates(message as ListOrderEventsRequest))
          as ListOrderEventsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListOrderEventsRequest() / ListOrderEventsRequest.new instead')
  static ListOrderEventsRequest create() => ListOrderEventsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListOrderEventsRequest._();
  @$core.override
  ListOrderEventsRequest createEmptyInstance() => ListOrderEventsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListOrderEventsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOrderEventsRequest>(
          ListOrderEventsRequest.$_createMessage);
  static ListOrderEventsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get salesOrderId => $_getSZ(0);
  @$pb.TagNumber(1)
  set salesOrderId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSalesOrderId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSalesOrderId() => $_clearField(1);

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

class ListOrderEventsResponse extends $pb.GeneratedMessage {
  factory ListOrderEventsResponse({
    $core.Iterable<SalesOrderEvent>? events,
    $core.int? total,
  }) {
    final result = ListOrderEventsResponse._();
    if (events != null) result.events.addAll(events);
    if (total != null) result.total = total;
    return result;
  }

  ListOrderEventsResponse._();

  factory ListOrderEventsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrderEventsResponse()..mergeFromBuffer(data, registry);
  factory ListOrderEventsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOrderEventsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOrderEventsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListOrderEventsResponse.$_createMessage)
    ..pPM<SalesOrderEvent>(1, _omitFieldNames ? '' : 'events',
        subBuilder: SalesOrderEvent.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrderEventsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOrderEventsResponse copyWith(
          void Function(ListOrderEventsResponse) updates) =>
      super.copyWith((message) => updates(message as ListOrderEventsResponse))
          as ListOrderEventsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListOrderEventsResponse() / ListOrderEventsResponse.new instead')
  static ListOrderEventsResponse create() => ListOrderEventsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListOrderEventsResponse._();
  @$core.override
  ListOrderEventsResponse createEmptyInstance() => ListOrderEventsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListOrderEventsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOrderEventsResponse>(
          ListOrderEventsResponse.$_createMessage);
  static ListOrderEventsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<SalesOrderEvent> get events => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get total => $_getIZ(1);
  @$pb.TagNumber(2)
  set total($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => $_clearField(2);
}

/// SalesOrderService:銷售訂單管理。
class SalesOrderServiceApi {
  final $pb.RpcClient _client;

  SalesOrderServiceApi(this._client);

  /// ListOrders:分頁查詢(預設排除軟刪除;status/customer_id/source/keyword 篩選)。
  $async.Future<ListOrdersResponse> listOrders(
          $pb.ClientContext? ctx, ListOrdersRequest request) =>
      _client.invoke<ListOrdersResponse>(ctx, 'SalesOrderService', 'ListOrders',
          request, ListOrdersResponse());

  /// GetOrder:以 id 取單筆(含明細)。
  $async.Future<GetOrderResponse> getOrder(
          $pb.ClientContext? ctx, GetOrderRequest request) =>
      _client.invoke<GetOrderResponse>(
          ctx, 'SalesOrderService', 'GetOrder', request, GetOrderResponse());

  /// CreateOrder:建立訂單(最小可用:客戶+來源+明細;取號+建單同交易)。
  $async.Future<CreateOrderResponse> createOrder(
          $pb.ClientContext? ctx, CreateOrderRequest request) =>
      _client.invoke<CreateOrderResponse>(ctx, 'SalesOrderService',
          'CreateOrder', request, CreateOrderResponse());

  /// UpdateOrder:僅 pending 可編輯(攜帶 version 樂觀鎖)。
  $async.Future<UpdateOrderResponse> updateOrder(
          $pb.ClientContext? ctx, UpdateOrderRequest request) =>
      _client.invoke<UpdateOrderResponse>(ctx, 'SalesOrderService',
          'UpdateOrder', request, UpdateOrderResponse());

  /// CancelOrder:僅 pending 可取消。
  $async.Future<CancelOrderResponse> cancelOrder(
          $pb.ClientContext? ctx, CancelOrderRequest request) =>
      _client.invoke<CancelOrderResponse>(ctx, 'SalesOrderService',
          'CancelOrder', request, CancelOrderResponse());

  /// CompleteOrder:僅 processing 可完成。
  $async.Future<CompleteOrderResponse> completeOrder(
          $pb.ClientContext? ctx, CompleteOrderRequest request) =>
      _client.invoke<CompleteOrderResponse>(ctx, 'SalesOrderService',
          'CompleteOrder', request, CompleteOrderResponse());

  /// VoidOrder:僅 completed 可作廢(dept_admin 以上 + 原因)。
  $async.Future<VoidOrderResponse> voidOrder(
          $pb.ClientContext? ctx, VoidOrderRequest request) =>
      _client.invoke<VoidOrderResponse>(
          ctx, 'SalesOrderService', 'VoidOrder', request, VoidOrderResponse());

  /// DeleteOrder:軟刪除(僅 pending/cancelled)。
  $async.Future<DeleteOrderResponse> deleteOrder(
          $pb.ClientContext? ctx, DeleteOrderRequest request) =>
      _client.invoke<DeleteOrderResponse>(ctx, 'SalesOrderService',
          'DeleteOrder', request, DeleteOrderResponse());

  /// ListOrderEvents:異動軌跡查詢(升序)。
  $async.Future<ListOrderEventsResponse> listOrderEvents(
          $pb.ClientContext? ctx, ListOrderEventsRequest request) =>
      _client.invoke<ListOrderEventsResponse>(ctx, 'SalesOrderService',
          'ListOrderEvents', request, ListOrderEventsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
