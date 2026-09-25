// This is a generated file - do not edit.
//
// Generated from salesorder/v1/announcement.proto.

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

/// Announcement:公告單筆。
class Announcement extends $pb.GeneratedMessage {
  factory Announcement({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? type,
    $core.String? title,
    $core.String? content,
    $core.String? imageUrl,
    $core.String? linkUrl,
    $core.String? publishAt,
    $core.String? unpublishAt,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.bool? deployWeb,
    $core.bool? deployApp,
    $core.String? createdAt,
    $core.String? updatedAt,
  }) {
    final result = Announcement._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (type != null) result.type = type;
    if (title != null) result.title = title;
    if (content != null) result.content = content;
    if (imageUrl != null) result.imageUrl = imageUrl;
    if (linkUrl != null) result.linkUrl = linkUrl;
    if (publishAt != null) result.publishAt = publishAt;
    if (unpublishAt != null) result.unpublishAt = unpublishAt;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (deployWeb != null) result.deployWeb = deployWeb;
    if (deployApp != null) result.deployApp = deployApp;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  Announcement._();

  factory Announcement.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Announcement()..mergeFromBuffer(data, registry);
  factory Announcement.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Announcement()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Announcement',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: Announcement.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'type')
    ..aOS(5, _omitFieldNames ? '' : 'title')
    ..aOS(6, _omitFieldNames ? '' : 'content')
    ..aOS(7, _omitFieldNames ? '' : 'imageUrl')
    ..aOS(8, _omitFieldNames ? '' : 'linkUrl')
    ..aOS(9, _omitFieldNames ? '' : 'publishAt')
    ..aOS(10, _omitFieldNames ? '' : 'unpublishAt')
    ..aI(11, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(12, _omitFieldNames ? '' : 'isActive')
    ..aOB(13, _omitFieldNames ? '' : 'deployWeb')
    ..aOB(14, _omitFieldNames ? '' : 'deployApp')
    ..aOS(15, _omitFieldNames ? '' : 'createdAt')
    ..aOS(16, _omitFieldNames ? '' : 'updatedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Announcement clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Announcement copyWith(void Function(Announcement) updates) =>
      super.copyWith((message) => updates(message as Announcement))
          as Announcement;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Announcement() / Announcement.new instead')
  static Announcement create() => Announcement._();
  static $pb.GeneratedMessage $_createMessage() => Announcement._();
  @$core.override
  Announcement createEmptyInstance() => Announcement._();
  @$core.pragma('dart2js:noInline')
  static Announcement getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Announcement>(
          Announcement.$_createMessage);
  static Announcement? _defaultInstance;

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
  $core.String get type => $_getSZ(3);
  @$pb.TagNumber(4)
  set type($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasType() => $_has(3);
  @$pb.TagNumber(4)
  void clearType() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get title => $_getSZ(4);
  @$pb.TagNumber(5)
  set title($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTitle() => $_has(4);
  @$pb.TagNumber(5)
  void clearTitle() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get content => $_getSZ(5);
  @$pb.TagNumber(6)
  set content($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasContent() => $_has(5);
  @$pb.TagNumber(6)
  void clearContent() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get imageUrl => $_getSZ(6);
  @$pb.TagNumber(7)
  set imageUrl($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasImageUrl() => $_has(6);
  @$pb.TagNumber(7)
  void clearImageUrl() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get linkUrl => $_getSZ(7);
  @$pb.TagNumber(8)
  set linkUrl($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasLinkUrl() => $_has(7);
  @$pb.TagNumber(8)
  void clearLinkUrl() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get publishAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set publishAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasPublishAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearPublishAt() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get unpublishAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set unpublishAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasUnpublishAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearUnpublishAt() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.int get sortOrder => $_getIZ(10);
  @$pb.TagNumber(11)
  set sortOrder($core.int value) => $_setSignedInt32(10, value);
  @$pb.TagNumber(11)
  $core.bool hasSortOrder() => $_has(10);
  @$pb.TagNumber(11)
  void clearSortOrder() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get isActive => $_getBF(11);
  @$pb.TagNumber(12)
  set isActive($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasIsActive() => $_has(11);
  @$pb.TagNumber(12)
  void clearIsActive() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get deployWeb => $_getBF(12);
  @$pb.TagNumber(13)
  set deployWeb($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasDeployWeb() => $_has(12);
  @$pb.TagNumber(13)
  void clearDeployWeb() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.bool get deployApp => $_getBF(13);
  @$pb.TagNumber(14)
  set deployApp($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(14)
  $core.bool hasDeployApp() => $_has(13);
  @$pb.TagNumber(14)
  void clearDeployApp() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.String get createdAt => $_getSZ(14);
  @$pb.TagNumber(15)
  set createdAt($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasCreatedAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearCreatedAt() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.String get updatedAt => $_getSZ(15);
  @$pb.TagNumber(16)
  set updatedAt($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasUpdatedAt() => $_has(15);
  @$pb.TagNumber(16)
  void clearUpdatedAt() => $_clearField(16);
}

/// ListAnnouncementsRequest:管理列表請求。
class ListAnnouncementsRequest extends $pb.GeneratedMessage {
  factory ListAnnouncementsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? type,
    $core.bool? includeDeleted,
  }) {
    final result = ListAnnouncementsRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (type != null) result.type = type;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListAnnouncementsRequest._();

  factory ListAnnouncementsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAnnouncementsRequest()..mergeFromBuffer(data, registry);
  factory ListAnnouncementsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAnnouncementsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAnnouncementsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListAnnouncementsRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAnnouncementsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAnnouncementsRequest copyWith(
          void Function(ListAnnouncementsRequest) updates) =>
      super.copyWith((message) => updates(message as ListAnnouncementsRequest))
          as ListAnnouncementsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAnnouncementsRequest() / ListAnnouncementsRequest.new instead')
  static ListAnnouncementsRequest create() => ListAnnouncementsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListAnnouncementsRequest._();
  @$core.override
  ListAnnouncementsRequest createEmptyInstance() =>
      ListAnnouncementsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAnnouncementsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAnnouncementsRequest>(
          ListAnnouncementsRequest.$_createMessage);
  static ListAnnouncementsRequest? _defaultInstance;

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
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(4)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(4)
  void clearIncludeDeleted() => $_clearField(4);
}

/// ListAnnouncementsResponse:管理列表結果。
class ListAnnouncementsResponse extends $pb.GeneratedMessage {
  factory ListAnnouncementsResponse({
    $core.Iterable<Announcement>? announcements,
    $core.int? total,
  }) {
    final result = ListAnnouncementsResponse._();
    if (announcements != null) result.announcements.addAll(announcements);
    if (total != null) result.total = total;
    return result;
  }

  ListAnnouncementsResponse._();

  factory ListAnnouncementsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAnnouncementsResponse()..mergeFromBuffer(data, registry);
  factory ListAnnouncementsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAnnouncementsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAnnouncementsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListAnnouncementsResponse.$_createMessage)
    ..pPM<Announcement>(1, _omitFieldNames ? '' : 'announcements',
        subBuilder: Announcement.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAnnouncementsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAnnouncementsResponse copyWith(
          void Function(ListAnnouncementsResponse) updates) =>
      super.copyWith((message) => updates(message as ListAnnouncementsResponse))
          as ListAnnouncementsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAnnouncementsResponse() / ListAnnouncementsResponse.new instead')
  static ListAnnouncementsResponse create() => ListAnnouncementsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAnnouncementsResponse._();
  @$core.override
  ListAnnouncementsResponse createEmptyInstance() =>
      ListAnnouncementsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAnnouncementsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAnnouncementsResponse>(
          ListAnnouncementsResponse.$_createMessage);
  static ListAnnouncementsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Announcement> get announcements => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get total => $_getIZ(1);
  @$pb.TagNumber(2)
  set total($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => $_clearField(2);
}

/// ListActiveAnnouncementsRequest:前台列表請求。
/// platform 決定平台投放過濾(規格「平台篩選投放」):web → 只回 deploy_web;app → 只回 deploy_app。
/// 未帶/非法值 → invalid_argument(不預設平台:預設會讓一邊靜默看到不該投放的公告)。
class ListActiveAnnouncementsRequest extends $pb.GeneratedMessage {
  factory ListActiveAnnouncementsRequest({
    $core.String? platform,
  }) {
    final result = ListActiveAnnouncementsRequest._();
    if (platform != null) result.platform = platform;
    return result;
  }

  ListActiveAnnouncementsRequest._();

  factory ListActiveAnnouncementsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListActiveAnnouncementsRequest()..mergeFromBuffer(data, registry);
  factory ListActiveAnnouncementsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListActiveAnnouncementsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListActiveAnnouncementsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListActiveAnnouncementsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'platform')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListActiveAnnouncementsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListActiveAnnouncementsRequest copyWith(
          void Function(ListActiveAnnouncementsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListActiveAnnouncementsRequest))
          as ListActiveAnnouncementsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListActiveAnnouncementsRequest() / ListActiveAnnouncementsRequest.new instead')
  static ListActiveAnnouncementsRequest create() =>
      ListActiveAnnouncementsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListActiveAnnouncementsRequest._();
  @$core.override
  ListActiveAnnouncementsRequest createEmptyInstance() =>
      ListActiveAnnouncementsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListActiveAnnouncementsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListActiveAnnouncementsRequest>(
          ListActiveAnnouncementsRequest.$_createMessage);
  static ListActiveAnnouncementsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get platform => $_getSZ(0);
  @$pb.TagNumber(1)
  set platform($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlatform() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlatform() => $_clearField(1);
}

/// ListActiveAnnouncementsResponse:前台列表結果。
/// 分開回三型別,讓兩端的前台不必各自依 type 再篩一次 —— 輪播(banner)與列表(news/article)
/// 的組成是一條規格(「前台展示與排序」),收斂在後端一處。
class ListActiveAnnouncementsResponse extends $pb.GeneratedMessage {
  factory ListActiveAnnouncementsResponse({
    $core.Iterable<Announcement>? banners,
    $core.Iterable<Announcement>? news,
    $core.Iterable<Announcement>? articles,
  }) {
    final result = ListActiveAnnouncementsResponse._();
    if (banners != null) result.banners.addAll(banners);
    if (news != null) result.news.addAll(news);
    if (articles != null) result.articles.addAll(articles);
    return result;
  }

  ListActiveAnnouncementsResponse._();

  factory ListActiveAnnouncementsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListActiveAnnouncementsResponse()..mergeFromBuffer(data, registry);
  factory ListActiveAnnouncementsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListActiveAnnouncementsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListActiveAnnouncementsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListActiveAnnouncementsResponse.$_createMessage)
    ..pPM<Announcement>(1, _omitFieldNames ? '' : 'banners',
        subBuilder: Announcement.$_createMessage)
    ..pPM<Announcement>(2, _omitFieldNames ? '' : 'news',
        subBuilder: Announcement.$_createMessage)
    ..pPM<Announcement>(3, _omitFieldNames ? '' : 'articles',
        subBuilder: Announcement.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListActiveAnnouncementsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListActiveAnnouncementsResponse copyWith(
          void Function(ListActiveAnnouncementsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListActiveAnnouncementsResponse))
          as ListActiveAnnouncementsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListActiveAnnouncementsResponse() / ListActiveAnnouncementsResponse.new instead')
  static ListActiveAnnouncementsResponse create() =>
      ListActiveAnnouncementsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListActiveAnnouncementsResponse._();
  @$core.override
  ListActiveAnnouncementsResponse createEmptyInstance() =>
      ListActiveAnnouncementsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListActiveAnnouncementsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListActiveAnnouncementsResponse>(
          ListActiveAnnouncementsResponse.$_createMessage);
  static ListActiveAnnouncementsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Announcement> get banners => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<Announcement> get news => $_getList(1);

  @$pb.TagNumber(3)
  $pb.PbList<Announcement> get articles => $_getList(2);
}

/// CreateAnnouncementRequest:建立請求。**範圍欄位空值自動歸屬**(非 super):
/// company_id 空 → 自己的公司;dept_admin 的 department_id 空 → 自己的部門
/// (故 dept_admin 一律建部門層、company_admin 預設公司層,super 空 = 全系統)。
/// 覆寫時由 AnnouncementService 範圍守衛驗證(spec「管理權限依範圍分層」)。
/// publish_at 空 = 立即;unpublish_at 空 = 不自動下架。
class CreateAnnouncementRequest extends $pb.GeneratedMessage {
  factory CreateAnnouncementRequest({
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? type,
    $core.String? title,
    $core.String? content,
    $core.String? imageUrl,
    $core.String? linkUrl,
    $core.String? publishAt,
    $core.String? unpublishAt,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.bool? deployWeb,
    $core.bool? deployApp,
  }) {
    final result = CreateAnnouncementRequest._();
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (type != null) result.type = type;
    if (title != null) result.title = title;
    if (content != null) result.content = content;
    if (imageUrl != null) result.imageUrl = imageUrl;
    if (linkUrl != null) result.linkUrl = linkUrl;
    if (publishAt != null) result.publishAt = publishAt;
    if (unpublishAt != null) result.unpublishAt = unpublishAt;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (deployWeb != null) result.deployWeb = deployWeb;
    if (deployApp != null) result.deployApp = deployApp;
    return result;
  }

  CreateAnnouncementRequest._();

  factory CreateAnnouncementRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAnnouncementRequest()..mergeFromBuffer(data, registry);
  factory CreateAnnouncementRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAnnouncementRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAnnouncementRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateAnnouncementRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'departmentId')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOS(4, _omitFieldNames ? '' : 'title')
    ..aOS(5, _omitFieldNames ? '' : 'content')
    ..aOS(6, _omitFieldNames ? '' : 'imageUrl')
    ..aOS(7, _omitFieldNames ? '' : 'linkUrl')
    ..aOS(8, _omitFieldNames ? '' : 'publishAt')
    ..aOS(9, _omitFieldNames ? '' : 'unpublishAt')
    ..aI(10, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(11, _omitFieldNames ? '' : 'isActive')
    ..aOB(12, _omitFieldNames ? '' : 'deployWeb')
    ..aOB(13, _omitFieldNames ? '' : 'deployApp')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAnnouncementRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAnnouncementRequest copyWith(
          void Function(CreateAnnouncementRequest) updates) =>
      super.copyWith((message) => updates(message as CreateAnnouncementRequest))
          as CreateAnnouncementRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAnnouncementRequest() / CreateAnnouncementRequest.new instead')
  static CreateAnnouncementRequest create() => CreateAnnouncementRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateAnnouncementRequest._();
  @$core.override
  CreateAnnouncementRequest createEmptyInstance() =>
      CreateAnnouncementRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateAnnouncementRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAnnouncementRequest>(
          CreateAnnouncementRequest.$_createMessage);
  static CreateAnnouncementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get departmentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set departmentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDepartmentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDepartmentId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get title => $_getSZ(3);
  @$pb.TagNumber(4)
  set title($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTitle() => $_has(3);
  @$pb.TagNumber(4)
  void clearTitle() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get content => $_getSZ(4);
  @$pb.TagNumber(5)
  set content($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasContent() => $_has(4);
  @$pb.TagNumber(5)
  void clearContent() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get imageUrl => $_getSZ(5);
  @$pb.TagNumber(6)
  set imageUrl($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasImageUrl() => $_has(5);
  @$pb.TagNumber(6)
  void clearImageUrl() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get linkUrl => $_getSZ(6);
  @$pb.TagNumber(7)
  set linkUrl($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasLinkUrl() => $_has(6);
  @$pb.TagNumber(7)
  void clearLinkUrl() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get publishAt => $_getSZ(7);
  @$pb.TagNumber(8)
  set publishAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasPublishAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearPublishAt() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get unpublishAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set unpublishAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUnpublishAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearUnpublishAt() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.int get sortOrder => $_getIZ(9);
  @$pb.TagNumber(10)
  set sortOrder($core.int value) => $_setSignedInt32(9, value);
  @$pb.TagNumber(10)
  $core.bool hasSortOrder() => $_has(9);
  @$pb.TagNumber(10)
  void clearSortOrder() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.bool get isActive => $_getBF(10);
  @$pb.TagNumber(11)
  set isActive($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasIsActive() => $_has(10);
  @$pb.TagNumber(11)
  void clearIsActive() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get deployWeb => $_getBF(11);
  @$pb.TagNumber(12)
  set deployWeb($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasDeployWeb() => $_has(11);
  @$pb.TagNumber(12)
  void clearDeployWeb() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get deployApp => $_getBF(12);
  @$pb.TagNumber(13)
  set deployApp($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasDeployApp() => $_has(12);
  @$pb.TagNumber(13)
  void clearDeployApp() => $_clearField(13);
}

/// CreateAnnouncementResponse:建立結果。
class CreateAnnouncementResponse extends $pb.GeneratedMessage {
  factory CreateAnnouncementResponse({
    Announcement? announcement,
  }) {
    final result = CreateAnnouncementResponse._();
    if (announcement != null) result.announcement = announcement;
    return result;
  }

  CreateAnnouncementResponse._();

  factory CreateAnnouncementResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAnnouncementResponse()..mergeFromBuffer(data, registry);
  factory CreateAnnouncementResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAnnouncementResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAnnouncementResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateAnnouncementResponse.$_createMessage)
    ..aOM<Announcement>(1, _omitFieldNames ? '' : 'announcement',
        subBuilder: Announcement.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAnnouncementResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAnnouncementResponse copyWith(
          void Function(CreateAnnouncementResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateAnnouncementResponse))
          as CreateAnnouncementResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAnnouncementResponse() / CreateAnnouncementResponse.new instead')
  static CreateAnnouncementResponse create() => CreateAnnouncementResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateAnnouncementResponse._();
  @$core.override
  CreateAnnouncementResponse createEmptyInstance() =>
      CreateAnnouncementResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateAnnouncementResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAnnouncementResponse>(
          CreateAnnouncementResponse.$_createMessage);
  static CreateAnnouncementResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Announcement get announcement => $_getN(0);
  @$pb.TagNumber(1)
  set announcement(Announcement value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAnnouncement() => $_has(0);
  @$pb.TagNumber(1)
  void clearAnnouncement() => $_clearField(1);
  @$pb.TagNumber(1)
  Announcement ensureAnnouncement() => $_ensure(0);
}

/// UpdateAnnouncementRequest:全量替換請求(id 除外的內容欄位;布林無 present 語意,
/// 故不採欄位式;空字串可選欄位即清空,publish_at 空 → invalid_argument)。
/// **發佈範圍不可改**(v1:範圍於建立時決定;要改範圍 = 另建一筆並刪舊筆 ——
/// 避免 proto3 無法區分「清空範圍」與「不帶欄位」的含糊語意)。
class UpdateAnnouncementRequest extends $pb.GeneratedMessage {
  factory UpdateAnnouncementRequest({
    $core.String? id,
    $core.String? type,
    $core.String? title,
    $core.String? content,
    $core.String? imageUrl,
    $core.String? linkUrl,
    $core.String? publishAt,
    $core.String? unpublishAt,
    $core.int? sortOrder,
    $core.bool? isActive,
    $core.bool? deployWeb,
    $core.bool? deployApp,
  }) {
    final result = UpdateAnnouncementRequest._();
    if (id != null) result.id = id;
    if (type != null) result.type = type;
    if (title != null) result.title = title;
    if (content != null) result.content = content;
    if (imageUrl != null) result.imageUrl = imageUrl;
    if (linkUrl != null) result.linkUrl = linkUrl;
    if (publishAt != null) result.publishAt = publishAt;
    if (unpublishAt != null) result.unpublishAt = unpublishAt;
    if (sortOrder != null) result.sortOrder = sortOrder;
    if (isActive != null) result.isActive = isActive;
    if (deployWeb != null) result.deployWeb = deployWeb;
    if (deployApp != null) result.deployApp = deployApp;
    return result;
  }

  UpdateAnnouncementRequest._();

  factory UpdateAnnouncementRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAnnouncementRequest()..mergeFromBuffer(data, registry);
  factory UpdateAnnouncementRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAnnouncementRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAnnouncementRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateAnnouncementRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(4, _omitFieldNames ? '' : 'type')
    ..aOS(5, _omitFieldNames ? '' : 'title')
    ..aOS(6, _omitFieldNames ? '' : 'content')
    ..aOS(7, _omitFieldNames ? '' : 'imageUrl')
    ..aOS(8, _omitFieldNames ? '' : 'linkUrl')
    ..aOS(9, _omitFieldNames ? '' : 'publishAt')
    ..aOS(10, _omitFieldNames ? '' : 'unpublishAt')
    ..aI(11, _omitFieldNames ? '' : 'sortOrder')
    ..aOB(12, _omitFieldNames ? '' : 'isActive')
    ..aOB(13, _omitFieldNames ? '' : 'deployWeb')
    ..aOB(14, _omitFieldNames ? '' : 'deployApp')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAnnouncementRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAnnouncementRequest copyWith(
          void Function(UpdateAnnouncementRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateAnnouncementRequest))
          as UpdateAnnouncementRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAnnouncementRequest() / UpdateAnnouncementRequest.new instead')
  static UpdateAnnouncementRequest create() => UpdateAnnouncementRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateAnnouncementRequest._();
  @$core.override
  UpdateAnnouncementRequest createEmptyInstance() =>
      UpdateAnnouncementRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAnnouncementRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAnnouncementRequest>(
          UpdateAnnouncementRequest.$_createMessage);
  static UpdateAnnouncementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(4)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(4)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(4)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(4)
  void clearType() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(5)
  set title($core.String value) => $_setString(2, value);
  @$pb.TagNumber(5)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(5)
  void clearTitle() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get content => $_getSZ(3);
  @$pb.TagNumber(6)
  set content($core.String value) => $_setString(3, value);
  @$pb.TagNumber(6)
  $core.bool hasContent() => $_has(3);
  @$pb.TagNumber(6)
  void clearContent() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get imageUrl => $_getSZ(4);
  @$pb.TagNumber(7)
  set imageUrl($core.String value) => $_setString(4, value);
  @$pb.TagNumber(7)
  $core.bool hasImageUrl() => $_has(4);
  @$pb.TagNumber(7)
  void clearImageUrl() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get linkUrl => $_getSZ(5);
  @$pb.TagNumber(8)
  set linkUrl($core.String value) => $_setString(5, value);
  @$pb.TagNumber(8)
  $core.bool hasLinkUrl() => $_has(5);
  @$pb.TagNumber(8)
  void clearLinkUrl() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get publishAt => $_getSZ(6);
  @$pb.TagNumber(9)
  set publishAt($core.String value) => $_setString(6, value);
  @$pb.TagNumber(9)
  $core.bool hasPublishAt() => $_has(6);
  @$pb.TagNumber(9)
  void clearPublishAt() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get unpublishAt => $_getSZ(7);
  @$pb.TagNumber(10)
  set unpublishAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(10)
  $core.bool hasUnpublishAt() => $_has(7);
  @$pb.TagNumber(10)
  void clearUnpublishAt() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.int get sortOrder => $_getIZ(8);
  @$pb.TagNumber(11)
  set sortOrder($core.int value) => $_setSignedInt32(8, value);
  @$pb.TagNumber(11)
  $core.bool hasSortOrder() => $_has(8);
  @$pb.TagNumber(11)
  void clearSortOrder() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get isActive => $_getBF(9);
  @$pb.TagNumber(12)
  set isActive($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(12)
  $core.bool hasIsActive() => $_has(9);
  @$pb.TagNumber(12)
  void clearIsActive() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get deployWeb => $_getBF(10);
  @$pb.TagNumber(13)
  set deployWeb($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(13)
  $core.bool hasDeployWeb() => $_has(10);
  @$pb.TagNumber(13)
  void clearDeployWeb() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.bool get deployApp => $_getBF(11);
  @$pb.TagNumber(14)
  set deployApp($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(14)
  $core.bool hasDeployApp() => $_has(11);
  @$pb.TagNumber(14)
  void clearDeployApp() => $_clearField(14);
}

/// UpdateAnnouncementResponse:更新結果。
class UpdateAnnouncementResponse extends $pb.GeneratedMessage {
  factory UpdateAnnouncementResponse({
    Announcement? announcement,
  }) {
    final result = UpdateAnnouncementResponse._();
    if (announcement != null) result.announcement = announcement;
    return result;
  }

  UpdateAnnouncementResponse._();

  factory UpdateAnnouncementResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAnnouncementResponse()..mergeFromBuffer(data, registry);
  factory UpdateAnnouncementResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAnnouncementResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAnnouncementResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateAnnouncementResponse.$_createMessage)
    ..aOM<Announcement>(1, _omitFieldNames ? '' : 'announcement',
        subBuilder: Announcement.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAnnouncementResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAnnouncementResponse copyWith(
          void Function(UpdateAnnouncementResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateAnnouncementResponse))
          as UpdateAnnouncementResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAnnouncementResponse() / UpdateAnnouncementResponse.new instead')
  static UpdateAnnouncementResponse create() => UpdateAnnouncementResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateAnnouncementResponse._();
  @$core.override
  UpdateAnnouncementResponse createEmptyInstance() =>
      UpdateAnnouncementResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAnnouncementResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAnnouncementResponse>(
          UpdateAnnouncementResponse.$_createMessage);
  static UpdateAnnouncementResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Announcement get announcement => $_getN(0);
  @$pb.TagNumber(1)
  set announcement(Announcement value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAnnouncement() => $_has(0);
  @$pb.TagNumber(1)
  void clearAnnouncement() => $_clearField(1);
  @$pb.TagNumber(1)
  Announcement ensureAnnouncement() => $_ensure(0);
}

/// DeleteAnnouncementRequest:軟刪除請求。
class DeleteAnnouncementRequest extends $pb.GeneratedMessage {
  factory DeleteAnnouncementRequest({
    $core.String? id,
  }) {
    final result = DeleteAnnouncementRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteAnnouncementRequest._();

  factory DeleteAnnouncementRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAnnouncementRequest()..mergeFromBuffer(data, registry);
  factory DeleteAnnouncementRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAnnouncementRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAnnouncementRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteAnnouncementRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAnnouncementRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAnnouncementRequest copyWith(
          void Function(DeleteAnnouncementRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteAnnouncementRequest))
          as DeleteAnnouncementRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAnnouncementRequest() / DeleteAnnouncementRequest.new instead')
  static DeleteAnnouncementRequest create() => DeleteAnnouncementRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteAnnouncementRequest._();
  @$core.override
  DeleteAnnouncementRequest createEmptyInstance() =>
      DeleteAnnouncementRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteAnnouncementRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAnnouncementRequest>(
          DeleteAnnouncementRequest.$_createMessage);
  static DeleteAnnouncementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// DeleteAnnouncementResponse:軟刪除結果。
class DeleteAnnouncementResponse extends $pb.GeneratedMessage {
  factory DeleteAnnouncementResponse() => DeleteAnnouncementResponse._();

  DeleteAnnouncementResponse._();

  factory DeleteAnnouncementResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAnnouncementResponse()..mergeFromBuffer(data, registry);
  factory DeleteAnnouncementResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAnnouncementResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAnnouncementResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteAnnouncementResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAnnouncementResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAnnouncementResponse copyWith(
          void Function(DeleteAnnouncementResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteAnnouncementResponse))
          as DeleteAnnouncementResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAnnouncementResponse() / DeleteAnnouncementResponse.new instead')
  static DeleteAnnouncementResponse create() => DeleteAnnouncementResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteAnnouncementResponse._();
  @$core.override
  DeleteAnnouncementResponse createEmptyInstance() =>
      DeleteAnnouncementResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteAnnouncementResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAnnouncementResponse>(
          DeleteAnnouncementResponse.$_createMessage);
  static DeleteAnnouncementResponse? _defaultInstance;
}

/// AnnouncementService:公告 CMS(spec announcements)。
/// 三型別(banner/news/article)、三層發佈範圍(company/department 皆空 = 全系統)、
/// 上下架時間窗、平台投放(Web/App)。CRUD 範圍守衛:super 全範圍、
/// company_admin 所屬公司(公司層+該公司部門層)、dept_admin 本部門層。
class AnnouncementServiceApi {
  final $pb.RpcClient _client;

  AnnouncementServiceApi(this._client);

  /// ListAnnouncements:管理列表(僅列可管理範圍;含未上架/停用;預設排除已刪除)。
  $async.Future<ListAnnouncementsResponse> listAnnouncements(
          $pb.ClientContext? ctx, ListAnnouncementsRequest request) =>
      _client.invoke<ListAnnouncementsResponse>(ctx, 'AnnouncementService',
          'ListAnnouncements', request, ListAnnouncementsResponse());

  /// ListActiveAnnouncements:前台列表(規格「上下架時間窗與啟用狀態」+「平台篩選投放」)
  /// —— 只回**當下可見**的公告(is_active=true、publish_at<=now、(unpublish_at 空或 >now)、
  /// deploy_web|deploy_app 依 platform 過濾),依 type 分組供輪播(banner)與列表(news/article)。
  /// 可見範圍仍由 RLS 兜底(全系統 + 自己公司 + 自己部門);不帶 page(前台一次全取)。
  $async.Future<ListActiveAnnouncementsResponse> listActiveAnnouncements(
          $pb.ClientContext? ctx, ListActiveAnnouncementsRequest request) =>
      _client.invoke<ListActiveAnnouncementsResponse>(
          ctx,
          'AnnouncementService',
          'ListActiveAnnouncements',
          request,
          ListActiveAnnouncementsResponse());

  /// CreateAnnouncement:建立公告(範圍依身分收斂;type 非法 → invalid_argument)。
  $async.Future<CreateAnnouncementResponse> createAnnouncement(
          $pb.ClientContext? ctx, CreateAnnouncementRequest request) =>
      _client.invoke<CreateAnnouncementResponse>(ctx, 'AnnouncementService',
          'CreateAnnouncement', request, CreateAnnouncementResponse());

  /// UpdateAnnouncement:全量替換(id 除外表所有欄位;布林無 present 語意,不採欄位式)。
  /// 空字串的可選欄位(image_url/link_url/unpublish_at)即清空;publish_at 空 → invalid_argument。
  $async.Future<UpdateAnnouncementResponse> updateAnnouncement(
          $pb.ClientContext? ctx, UpdateAnnouncementRequest request) =>
      _client.invoke<UpdateAnnouncementResponse>(ctx, 'AnnouncementService',
          'UpdateAnnouncement', request, UpdateAnnouncementResponse());

  /// DeleteAnnouncement:軟刪除。
  $async.Future<DeleteAnnouncementResponse> deleteAnnouncement(
          $pb.ClientContext? ctx, DeleteAnnouncementRequest request) =>
      _client.invoke<DeleteAnnouncementResponse>(ctx, 'AnnouncementService',
          'DeleteAnnouncement', request, DeleteAnnouncementResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
