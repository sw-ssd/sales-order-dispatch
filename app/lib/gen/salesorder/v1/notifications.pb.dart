// This is a generated file - do not edit.
//
// Generated from salesorder/v1/notifications.proto.

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

/// NotificationView:通知單筆。
class NotificationView extends $pb.GeneratedMessage {
  factory NotificationView({
    $core.String? id,
    $core.String? channel,
    $core.String? title,
    $core.String? content,
    $core.String? payload,
    $core.String? status,
    $core.String? sentAt,
    $core.String? readAt,
    $core.String? createdAt,
  }) {
    final result = NotificationView._();
    if (id != null) result.id = id;
    if (channel != null) result.channel = channel;
    if (title != null) result.title = title;
    if (content != null) result.content = content;
    if (payload != null) result.payload = payload;
    if (status != null) result.status = status;
    if (sentAt != null) result.sentAt = sentAt;
    if (readAt != null) result.readAt = readAt;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  NotificationView._();

  factory NotificationView.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      NotificationView()..mergeFromBuffer(data, registry);
  factory NotificationView.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      NotificationView()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NotificationView',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: NotificationView.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'channel')
    ..aOS(3, _omitFieldNames ? '' : 'title')
    ..aOS(4, _omitFieldNames ? '' : 'content')
    ..aOS(5, _omitFieldNames ? '' : 'payload')
    ..aOS(6, _omitFieldNames ? '' : 'status')
    ..aOS(7, _omitFieldNames ? '' : 'sentAt')
    ..aOS(8, _omitFieldNames ? '' : 'readAt')
    ..aOS(9, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotificationView clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotificationView copyWith(void Function(NotificationView) updates) =>
      super.copyWith((message) => updates(message as NotificationView))
          as NotificationView;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use NotificationView() / NotificationView.new instead')
  static NotificationView create() => NotificationView._();
  static $pb.GeneratedMessage $_createMessage() => NotificationView._();
  @$core.override
  NotificationView createEmptyInstance() => NotificationView._();
  @$core.pragma('dart2js:noInline')
  static NotificationView getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NotificationView>(
          NotificationView.$_createMessage);
  static NotificationView? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get channel => $_getSZ(1);
  @$pb.TagNumber(2)
  set channel($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChannel() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannel() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(3)
  set title($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitle() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get content => $_getSZ(3);
  @$pb.TagNumber(4)
  set content($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasContent() => $_has(3);
  @$pb.TagNumber(4)
  void clearContent() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get payload => $_getSZ(4);
  @$pb.TagNumber(5)
  set payload($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPayload() => $_has(4);
  @$pb.TagNumber(5)
  void clearPayload() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get status => $_getSZ(5);
  @$pb.TagNumber(6)
  set status($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get sentAt => $_getSZ(6);
  @$pb.TagNumber(7)
  set sentAt($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasSentAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearSentAt() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get readAt => $_getSZ(7);
  @$pb.TagNumber(8)
  set readAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasReadAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearReadAt() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get createdAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set createdAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCreatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearCreatedAt() => $_clearField(9);
}

/// ListNotificationsRequest:列表請求。
class ListNotificationsRequest extends $pb.GeneratedMessage {
  factory ListNotificationsRequest({
    $core.bool? unreadOnly,
    $core.String? channel,
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListNotificationsRequest._();
    if (unreadOnly != null) result.unreadOnly = unreadOnly;
    if (channel != null) result.channel = channel;
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListNotificationsRequest._();

  factory ListNotificationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotificationsRequest()..mergeFromBuffer(data, registry);
  factory ListNotificationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotificationsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListNotificationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListNotificationsRequest.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'unreadOnly')
    ..aOS(2, _omitFieldNames ? '' : 'channel')
    ..aI(3, _omitFieldNames ? '' : 'page')
    ..aI(4, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotificationsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotificationsRequest copyWith(
          void Function(ListNotificationsRequest) updates) =>
      super.copyWith((message) => updates(message as ListNotificationsRequest))
          as ListNotificationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListNotificationsRequest() / ListNotificationsRequest.new instead')
  static ListNotificationsRequest create() => ListNotificationsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListNotificationsRequest._();
  @$core.override
  ListNotificationsRequest createEmptyInstance() =>
      ListNotificationsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListNotificationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListNotificationsRequest>(
          ListNotificationsRequest.$_createMessage);
  static ListNotificationsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get unreadOnly => $_getBF(0);
  @$pb.TagNumber(1)
  set unreadOnly($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUnreadOnly() => $_has(0);
  @$pb.TagNumber(1)
  void clearUnreadOnly() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get channel => $_getSZ(1);
  @$pb.TagNumber(2)
  set channel($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasChannel() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannel() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get page => $_getIZ(2);
  @$pb.TagNumber(3)
  set page($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPage() => $_has(2);
  @$pb.TagNumber(3)
  void clearPage() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get pageSize => $_getIZ(3);
  @$pb.TagNumber(4)
  set pageSize($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPageSize() => $_has(3);
  @$pb.TagNumber(4)
  void clearPageSize() => $_clearField(4);
}

/// ListNotificationsResponse:列表結果。
class ListNotificationsResponse extends $pb.GeneratedMessage {
  factory ListNotificationsResponse({
    $core.Iterable<NotificationView>? notifications,
    $core.int? page,
    $core.int? pageSize,
    $core.int? total,
    $core.int? unreadCount,
  }) {
    final result = ListNotificationsResponse._();
    if (notifications != null) result.notifications.addAll(notifications);
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (total != null) result.total = total;
    if (unreadCount != null) result.unreadCount = unreadCount;
    return result;
  }

  ListNotificationsResponse._();

  factory ListNotificationsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotificationsResponse()..mergeFromBuffer(data, registry);
  factory ListNotificationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotificationsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListNotificationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListNotificationsResponse.$_createMessage)
    ..pPM<NotificationView>(1, _omitFieldNames ? '' : 'notifications',
        subBuilder: NotificationView.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'page')
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..aI(4, _omitFieldNames ? '' : 'total')
    ..aI(5, _omitFieldNames ? '' : 'unreadCount')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotificationsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotificationsResponse copyWith(
          void Function(ListNotificationsResponse) updates) =>
      super.copyWith((message) => updates(message as ListNotificationsResponse))
          as ListNotificationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListNotificationsResponse() / ListNotificationsResponse.new instead')
  static ListNotificationsResponse create() => ListNotificationsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListNotificationsResponse._();
  @$core.override
  ListNotificationsResponse createEmptyInstance() =>
      ListNotificationsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListNotificationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListNotificationsResponse>(
          ListNotificationsResponse.$_createMessage);
  static ListNotificationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<NotificationView> get notifications => $_getList(0);

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

  @$pb.TagNumber(5)
  $core.int get unreadCount => $_getIZ(4);
  @$pb.TagNumber(5)
  set unreadCount($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUnreadCount() => $_has(4);
  @$pb.TagNumber(5)
  void clearUnreadCount() => $_clearField(5);
}

/// MarkReadRequest:已讀請求。
class MarkReadRequest extends $pb.GeneratedMessage {
  factory MarkReadRequest({
    $core.Iterable<$core.String>? notificationIds,
  }) {
    final result = MarkReadRequest._();
    if (notificationIds != null) result.notificationIds.addAll(notificationIds);
    return result;
  }

  MarkReadRequest._();

  factory MarkReadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkReadRequest()..mergeFromBuffer(data, registry);
  factory MarkReadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkReadRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkReadRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: MarkReadRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'notificationIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkReadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkReadRequest copyWith(void Function(MarkReadRequest) updates) =>
      super.copyWith((message) => updates(message as MarkReadRequest))
          as MarkReadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MarkReadRequest() / MarkReadRequest.new instead')
  static MarkReadRequest create() => MarkReadRequest._();
  static $pb.GeneratedMessage $_createMessage() => MarkReadRequest._();
  @$core.override
  MarkReadRequest createEmptyInstance() => MarkReadRequest._();
  @$core.pragma('dart2js:noInline')
  static MarkReadRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MarkReadRequest>(
          MarkReadRequest.$_createMessage);
  static MarkReadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get notificationIds => $_getList(0);
}

/// MarkReadResponse:已讀結果。
class MarkReadResponse extends $pb.GeneratedMessage {
  factory MarkReadResponse({
    $core.int? markedCount,
  }) {
    final result = MarkReadResponse._();
    if (markedCount != null) result.markedCount = markedCount;
    return result;
  }

  MarkReadResponse._();

  factory MarkReadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkReadResponse()..mergeFromBuffer(data, registry);
  factory MarkReadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkReadResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkReadResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: MarkReadResponse.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'markedCount')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkReadResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkReadResponse copyWith(void Function(MarkReadResponse) updates) =>
      super.copyWith((message) => updates(message as MarkReadResponse))
          as MarkReadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MarkReadResponse() / MarkReadResponse.new instead')
  static MarkReadResponse create() => MarkReadResponse._();
  static $pb.GeneratedMessage $_createMessage() => MarkReadResponse._();
  @$core.override
  MarkReadResponse createEmptyInstance() => MarkReadResponse._();
  @$core.pragma('dart2js:noInline')
  static MarkReadResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MarkReadResponse>(
          MarkReadResponse.$_createMessage);
  static MarkReadResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get markedCount => $_getIZ(0);
  @$pb.TagNumber(1)
  set markedCount($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMarkedCount() => $_has(0);
  @$pb.TagNumber(1)
  void clearMarkedCount() => $_clearField(1);
}

/// UnreadCountRequest:未讀數請求。
class UnreadCountRequest extends $pb.GeneratedMessage {
  factory UnreadCountRequest() => UnreadCountRequest._();

  UnreadCountRequest._();

  factory UnreadCountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnreadCountRequest()..mergeFromBuffer(data, registry);
  factory UnreadCountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnreadCountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UnreadCountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UnreadCountRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadCountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadCountRequest copyWith(void Function(UnreadCountRequest) updates) =>
      super.copyWith((message) => updates(message as UnreadCountRequest))
          as UnreadCountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UnreadCountRequest() / UnreadCountRequest.new instead')
  static UnreadCountRequest create() => UnreadCountRequest._();
  static $pb.GeneratedMessage $_createMessage() => UnreadCountRequest._();
  @$core.override
  UnreadCountRequest createEmptyInstance() => UnreadCountRequest._();
  @$core.pragma('dart2js:noInline')
  static UnreadCountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UnreadCountRequest>(
          UnreadCountRequest.$_createMessage);
  static UnreadCountRequest? _defaultInstance;
}

/// UnreadCountResponse:未讀數結果。
class UnreadCountResponse extends $pb.GeneratedMessage {
  factory UnreadCountResponse({
    $core.int? count,
  }) {
    final result = UnreadCountResponse._();
    if (count != null) result.count = count;
    return result;
  }

  UnreadCountResponse._();

  factory UnreadCountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnreadCountResponse()..mergeFromBuffer(data, registry);
  factory UnreadCountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UnreadCountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UnreadCountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UnreadCountResponse.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'count')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadCountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadCountResponse copyWith(void Function(UnreadCountResponse) updates) =>
      super.copyWith((message) => updates(message as UnreadCountResponse))
          as UnreadCountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use UnreadCountResponse() / UnreadCountResponse.new instead')
  static UnreadCountResponse create() => UnreadCountResponse._();
  static $pb.GeneratedMessage $_createMessage() => UnreadCountResponse._();
  @$core.override
  UnreadCountResponse createEmptyInstance() => UnreadCountResponse._();
  @$core.pragma('dart2js:noInline')
  static UnreadCountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UnreadCountResponse>(
          UnreadCountResponse.$_createMessage);
  static UnreadCountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get count => $_getIZ(0);
  @$pb.TagNumber(1)
  set count($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCount() => $_has(0);
  @$pb.TagNumber(1)
  void clearCount() => $_clearField(1);
}

/// NotificationService:通知中心(07 計畫 Task 4.3.3)。
/// 僅回傳當前使用者本人的通知;記錄不可刪除(規格 §5.4)。
class NotificationServiceApi {
  final $pb.RpcClient _client;

  NotificationServiceApi(this._client);

  /// ListNotifications:本人通知列表(分頁 per_page ≤ 100;unread_only 篩未讀)。
  $async.Future<ListNotificationsResponse> listNotifications(
          $pb.ClientContext? ctx, ListNotificationsRequest request) =>
      _client.invoke<ListNotificationsResponse>(ctx, 'NotificationService',
          'ListNotifications', request, ListNotificationsResponse());

  /// MarkRead:標記已讀(僅本人;sent/pending → read;failed 不可轉;冪等)。
  $async.Future<MarkReadResponse> markRead(
          $pb.ClientContext? ctx, MarkReadRequest request) =>
      _client.invoke<MarkReadResponse>(
          ctx, 'NotificationService', 'MarkRead', request, MarkReadResponse());

  /// UnreadCount:未讀數(通知鈴角標輪詢)。
  $async.Future<UnreadCountResponse> unreadCount(
          $pb.ClientContext? ctx, UnreadCountRequest request) =>
      _client.invoke<UnreadCountResponse>(ctx, 'NotificationService',
          'UnreadCount', request, UnreadCountResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
