// This is a generated file - do not edit.
//
// Generated from products/v1/print.proto.

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

/// PreviewRequest:預覽請求。
class PreviewRequest extends $pb.GeneratedMessage {
  factory PreviewRequest({
    $core.String? documentType,
    $core.String? routeId,
    $core.String? targetDate,
    $core.String? customerId,
    $core.String? warehouseId,
  }) {
    final result = PreviewRequest._();
    if (documentType != null) result.documentType = documentType;
    if (routeId != null) result.routeId = routeId;
    if (targetDate != null) result.targetDate = targetDate;
    if (customerId != null) result.customerId = customerId;
    if (warehouseId != null) result.warehouseId = warehouseId;
    return result;
  }

  PreviewRequest._();

  factory PreviewRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PreviewRequest()..mergeFromBuffer(data, registry);
  factory PreviewRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PreviewRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PreviewRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: PreviewRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'documentType')
    ..aOS(2, _omitFieldNames ? '' : 'routeId')
    ..aOS(3, _omitFieldNames ? '' : 'targetDate')
    ..aOS(4, _omitFieldNames ? '' : 'customerId')
    ..aOS(5, _omitFieldNames ? '' : 'warehouseId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PreviewRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PreviewRequest copyWith(void Function(PreviewRequest) updates) =>
      super.copyWith((message) => updates(message as PreviewRequest))
          as PreviewRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PreviewRequest() / PreviewRequest.new instead')
  static PreviewRequest create() => PreviewRequest._();
  static $pb.GeneratedMessage $_createMessage() => PreviewRequest._();
  @$core.override
  PreviewRequest createEmptyInstance() => PreviewRequest._();
  @$core.pragma('dart2js:noInline')
  static PreviewRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreviewRequest>(
          PreviewRequest.$_createMessage);
  static PreviewRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get documentType => $_getSZ(0);
  @$pb.TagNumber(1)
  set documentType($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDocumentType() => $_has(0);
  @$pb.TagNumber(1)
  void clearDocumentType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get routeId => $_getSZ(1);
  @$pb.TagNumber(2)
  set routeId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRouteId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRouteId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get targetDate => $_getSZ(2);
  @$pb.TagNumber(3)
  set targetDate($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTargetDate() => $_has(2);
  @$pb.TagNumber(3)
  void clearTargetDate() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get customerId => $_getSZ(3);
  @$pb.TagNumber(4)
  set customerId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCustomerId() => $_has(3);
  @$pb.TagNumber(4)
  void clearCustomerId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get warehouseId => $_getSZ(4);
  @$pb.TagNumber(5)
  set warehouseId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasWarehouseId() => $_has(4);
  @$pb.TagNumber(5)
  void clearWarehouseId() => $_clearField(5);
}

/// PreviewResponse:預覽結果。
class PreviewResponse extends $pb.GeneratedMessage {
  factory PreviewResponse({
    $core.String? previewId,
    $core.String? fileAssetId,
    $core.String? downloadUrl,
  }) {
    final result = PreviewResponse._();
    if (previewId != null) result.previewId = previewId;
    if (fileAssetId != null) result.fileAssetId = fileAssetId;
    if (downloadUrl != null) result.downloadUrl = downloadUrl;
    return result;
  }

  PreviewResponse._();

  factory PreviewResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PreviewResponse()..mergeFromBuffer(data, registry);
  factory PreviewResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PreviewResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PreviewResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: PreviewResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'previewId')
    ..aOS(2, _omitFieldNames ? '' : 'fileAssetId')
    ..aOS(3, _omitFieldNames ? '' : 'downloadUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PreviewResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PreviewResponse copyWith(void Function(PreviewResponse) updates) =>
      super.copyWith((message) => updates(message as PreviewResponse))
          as PreviewResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PreviewResponse() / PreviewResponse.new instead')
  static PreviewResponse create() => PreviewResponse._();
  static $pb.GeneratedMessage $_createMessage() => PreviewResponse._();
  @$core.override
  PreviewResponse createEmptyInstance() => PreviewResponse._();
  @$core.pragma('dart2js:noInline')
  static PreviewResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreviewResponse>(
          PreviewResponse.$_createMessage);
  static PreviewResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get previewId => $_getSZ(0);
  @$pb.TagNumber(1)
  set previewId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPreviewId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPreviewId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get fileAssetId => $_getSZ(1);
  @$pb.TagNumber(2)
  set fileAssetId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFileAssetId() => $_has(1);
  @$pb.TagNumber(2)
  void clearFileAssetId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get downloadUrl => $_getSZ(2);
  @$pb.TagNumber(3)
  set downloadUrl($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDownloadUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearDownloadUrl() => $_clearField(3);
}

/// PrintRequest:正式列印請求。
class PrintRequest extends $pb.GeneratedMessage {
  factory PrintRequest({
    $core.String? documentType,
    $core.String? routeId,
    $core.String? targetDate,
    $core.String? customerId,
    $core.String? warehouseId,
    $core.String? reprintReason,
  }) {
    final result = PrintRequest._();
    if (documentType != null) result.documentType = documentType;
    if (routeId != null) result.routeId = routeId;
    if (targetDate != null) result.targetDate = targetDate;
    if (customerId != null) result.customerId = customerId;
    if (warehouseId != null) result.warehouseId = warehouseId;
    if (reprintReason != null) result.reprintReason = reprintReason;
    return result;
  }

  PrintRequest._();

  factory PrintRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PrintRequest()..mergeFromBuffer(data, registry);
  factory PrintRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PrintRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PrintRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: PrintRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'documentType')
    ..aOS(2, _omitFieldNames ? '' : 'routeId')
    ..aOS(3, _omitFieldNames ? '' : 'targetDate')
    ..aOS(4, _omitFieldNames ? '' : 'customerId')
    ..aOS(5, _omitFieldNames ? '' : 'warehouseId')
    ..aOS(6, _omitFieldNames ? '' : 'reprintReason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PrintRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PrintRequest copyWith(void Function(PrintRequest) updates) =>
      super.copyWith((message) => updates(message as PrintRequest))
          as PrintRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PrintRequest() / PrintRequest.new instead')
  static PrintRequest create() => PrintRequest._();
  static $pb.GeneratedMessage $_createMessage() => PrintRequest._();
  @$core.override
  PrintRequest createEmptyInstance() => PrintRequest._();
  @$core.pragma('dart2js:noInline')
  static PrintRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PrintRequest>(
          PrintRequest.$_createMessage);
  static PrintRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get documentType => $_getSZ(0);
  @$pb.TagNumber(1)
  set documentType($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDocumentType() => $_has(0);
  @$pb.TagNumber(1)
  void clearDocumentType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get routeId => $_getSZ(1);
  @$pb.TagNumber(2)
  set routeId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRouteId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRouteId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get targetDate => $_getSZ(2);
  @$pb.TagNumber(3)
  set targetDate($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTargetDate() => $_has(2);
  @$pb.TagNumber(3)
  void clearTargetDate() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get customerId => $_getSZ(3);
  @$pb.TagNumber(4)
  set customerId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCustomerId() => $_has(3);
  @$pb.TagNumber(4)
  void clearCustomerId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get warehouseId => $_getSZ(4);
  @$pb.TagNumber(5)
  set warehouseId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasWarehouseId() => $_has(4);
  @$pb.TagNumber(5)
  void clearWarehouseId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get reprintReason => $_getSZ(5);
  @$pb.TagNumber(6)
  set reprintReason($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasReprintReason() => $_has(5);
  @$pb.TagNumber(6)
  void clearReprintReason() => $_clearField(6);
}

/// PrintResponse:正式列印結果。
class PrintResponse extends $pb.GeneratedMessage {
  factory PrintResponse({
    $core.String? printLogId,
    $core.String? fileAssetId,
    $core.String? downloadUrl,
    $core.bool? isReprint,
  }) {
    final result = PrintResponse._();
    if (printLogId != null) result.printLogId = printLogId;
    if (fileAssetId != null) result.fileAssetId = fileAssetId;
    if (downloadUrl != null) result.downloadUrl = downloadUrl;
    if (isReprint != null) result.isReprint = isReprint;
    return result;
  }

  PrintResponse._();

  factory PrintResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PrintResponse()..mergeFromBuffer(data, registry);
  factory PrintResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PrintResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PrintResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: PrintResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'printLogId')
    ..aOS(2, _omitFieldNames ? '' : 'fileAssetId')
    ..aOS(3, _omitFieldNames ? '' : 'downloadUrl')
    ..aOB(4, _omitFieldNames ? '' : 'isReprint')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PrintResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PrintResponse copyWith(void Function(PrintResponse) updates) =>
      super.copyWith((message) => updates(message as PrintResponse))
          as PrintResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PrintResponse() / PrintResponse.new instead')
  static PrintResponse create() => PrintResponse._();
  static $pb.GeneratedMessage $_createMessage() => PrintResponse._();
  @$core.override
  PrintResponse createEmptyInstance() => PrintResponse._();
  @$core.pragma('dart2js:noInline')
  static PrintResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PrintResponse>(
          PrintResponse.$_createMessage);
  static PrintResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get printLogId => $_getSZ(0);
  @$pb.TagNumber(1)
  set printLogId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPrintLogId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPrintLogId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get fileAssetId => $_getSZ(1);
  @$pb.TagNumber(2)
  set fileAssetId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFileAssetId() => $_has(1);
  @$pb.TagNumber(2)
  void clearFileAssetId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get downloadUrl => $_getSZ(2);
  @$pb.TagNumber(3)
  set downloadUrl($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDownloadUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearDownloadUrl() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get isReprint => $_getBF(3);
  @$pb.TagNumber(4)
  set isReprint($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsReprint() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsReprint() => $_clearField(4);
}

/// ListLogsRequest:列印記錄查詢。
class ListLogsRequest extends $pb.GeneratedMessage {
  factory ListLogsRequest({
    $core.String? dateFrom,
    $core.String? dateTo,
    $core.String? documentType,
    $core.String? routeId,
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListLogsRequest._();
    if (dateFrom != null) result.dateFrom = dateFrom;
    if (dateTo != null) result.dateTo = dateTo;
    if (documentType != null) result.documentType = documentType;
    if (routeId != null) result.routeId = routeId;
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListLogsRequest._();

  factory ListLogsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLogsRequest()..mergeFromBuffer(data, registry);
  factory ListLogsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLogsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLogsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ListLogsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'dateFrom')
    ..aOS(2, _omitFieldNames ? '' : 'dateTo')
    ..aOS(3, _omitFieldNames ? '' : 'documentType')
    ..aOS(4, _omitFieldNames ? '' : 'routeId')
    ..aI(5, _omitFieldNames ? '' : 'page')
    ..aI(6, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLogsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLogsRequest copyWith(void Function(ListLogsRequest) updates) =>
      super.copyWith((message) => updates(message as ListLogsRequest))
          as ListLogsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListLogsRequest() / ListLogsRequest.new instead')
  static ListLogsRequest create() => ListLogsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListLogsRequest._();
  @$core.override
  ListLogsRequest createEmptyInstance() => ListLogsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListLogsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListLogsRequest>(
          ListLogsRequest.$_createMessage);
  static ListLogsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get dateFrom => $_getSZ(0);
  @$pb.TagNumber(1)
  set dateFrom($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDateFrom() => $_has(0);
  @$pb.TagNumber(1)
  void clearDateFrom() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get dateTo => $_getSZ(1);
  @$pb.TagNumber(2)
  set dateTo($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDateTo() => $_has(1);
  @$pb.TagNumber(2)
  void clearDateTo() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get documentType => $_getSZ(2);
  @$pb.TagNumber(3)
  set documentType($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDocumentType() => $_has(2);
  @$pb.TagNumber(3)
  void clearDocumentType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get routeId => $_getSZ(3);
  @$pb.TagNumber(4)
  set routeId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRouteId() => $_has(3);
  @$pb.TagNumber(4)
  void clearRouteId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get page => $_getIZ(4);
  @$pb.TagNumber(5)
  set page($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPage() => $_has(4);
  @$pb.TagNumber(5)
  void clearPage() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get pageSize => $_getIZ(5);
  @$pb.TagNumber(6)
  set pageSize($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPageSize() => $_has(5);
  @$pb.TagNumber(6)
  void clearPageSize() => $_clearField(6);
}

/// LogEntry:單筆列印記錄。
class LogEntry extends $pb.GeneratedMessage {
  factory LogEntry({
    $core.String? id,
    $core.String? documentType,
    $core.String? routeId,
    $core.String? targetDate,
    $core.String? printedBy,
    $core.String? printedAt,
    $core.bool? isReprint,
    $core.String? reprintReason,
    $core.String? downloadUrl,
  }) {
    final result = LogEntry._();
    if (id != null) result.id = id;
    if (documentType != null) result.documentType = documentType;
    if (routeId != null) result.routeId = routeId;
    if (targetDate != null) result.targetDate = targetDate;
    if (printedBy != null) result.printedBy = printedBy;
    if (printedAt != null) result.printedAt = printedAt;
    if (isReprint != null) result.isReprint = isReprint;
    if (reprintReason != null) result.reprintReason = reprintReason;
    if (downloadUrl != null) result.downloadUrl = downloadUrl;
    return result;
  }

  LogEntry._();

  factory LogEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogEntry()..mergeFromBuffer(data, registry);
  factory LogEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogEntry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: LogEntry.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'documentType')
    ..aOS(3, _omitFieldNames ? '' : 'routeId')
    ..aOS(4, _omitFieldNames ? '' : 'targetDate')
    ..aOS(5, _omitFieldNames ? '' : 'printedBy')
    ..aOS(6, _omitFieldNames ? '' : 'printedAt')
    ..aOB(7, _omitFieldNames ? '' : 'isReprint')
    ..aOS(8, _omitFieldNames ? '' : 'reprintReason')
    ..aOS(9, _omitFieldNames ? '' : 'downloadUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogEntry copyWith(void Function(LogEntry) updates) =>
      super.copyWith((message) => updates(message as LogEntry)) as LogEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogEntry() / LogEntry.new instead')
  static LogEntry create() => LogEntry._();
  static $pb.GeneratedMessage $_createMessage() => LogEntry._();
  @$core.override
  LogEntry createEmptyInstance() => LogEntry._();
  @$core.pragma('dart2js:noInline')
  static LogEntry getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LogEntry>(LogEntry.$_createMessage);
  static LogEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get documentType => $_getSZ(1);
  @$pb.TagNumber(2)
  set documentType($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDocumentType() => $_has(1);
  @$pb.TagNumber(2)
  void clearDocumentType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get routeId => $_getSZ(2);
  @$pb.TagNumber(3)
  set routeId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRouteId() => $_has(2);
  @$pb.TagNumber(3)
  void clearRouteId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get targetDate => $_getSZ(3);
  @$pb.TagNumber(4)
  set targetDate($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTargetDate() => $_has(3);
  @$pb.TagNumber(4)
  void clearTargetDate() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get printedBy => $_getSZ(4);
  @$pb.TagNumber(5)
  set printedBy($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPrintedBy() => $_has(4);
  @$pb.TagNumber(5)
  void clearPrintedBy() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get printedAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set printedAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPrintedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearPrintedAt() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get isReprint => $_getBF(6);
  @$pb.TagNumber(7)
  set isReprint($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasIsReprint() => $_has(6);
  @$pb.TagNumber(7)
  void clearIsReprint() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get reprintReason => $_getSZ(7);
  @$pb.TagNumber(8)
  set reprintReason($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasReprintReason() => $_has(7);
  @$pb.TagNumber(8)
  void clearReprintReason() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get downloadUrl => $_getSZ(8);
  @$pb.TagNumber(9)
  set downloadUrl($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasDownloadUrl() => $_has(8);
  @$pb.TagNumber(9)
  void clearDownloadUrl() => $_clearField(9);
}

/// ListLogsResponse:列印記錄列表。
class ListLogsResponse extends $pb.GeneratedMessage {
  factory ListLogsResponse({
    $core.Iterable<LogEntry>? entries,
    $core.int? page,
    $core.int? pageSize,
    $core.int? total,
  }) {
    final result = ListLogsResponse._();
    if (entries != null) result.entries.addAll(entries);
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (total != null) result.total = total;
    return result;
  }

  ListLogsResponse._();

  factory ListLogsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLogsResponse()..mergeFromBuffer(data, registry);
  factory ListLogsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLogsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLogsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'products.v1'),
      createEmptyInstance: ListLogsResponse.$_createMessage)
    ..pPM<LogEntry>(1, _omitFieldNames ? '' : 'entries',
        subBuilder: LogEntry.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'page')
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..aI(4, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLogsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLogsResponse copyWith(void Function(ListLogsResponse) updates) =>
      super.copyWith((message) => updates(message as ListLogsResponse))
          as ListLogsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListLogsResponse() / ListLogsResponse.new instead')
  static ListLogsResponse create() => ListLogsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListLogsResponse._();
  @$core.override
  ListLogsResponse createEmptyInstance() => ListLogsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListLogsResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListLogsResponse>(
          ListLogsResponse.$_createMessage);
  static ListLogsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<LogEntry> get entries => $_getList(0);

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

/// PrintService:列印與預覽(09 計畫 Task 5.5.2–5.5.4)。
/// Preview 不限訂單狀態、不寫 print_logs;Print 要求範圍內訂單全 processing;
/// 重印沿用 Print(有既有記錄即重印分支,必填 reprint_reason)。
class PrintServiceApi {
  final $pb.RpcClient _client;

  PrintServiceApi(this._client);

  /// Preview:預覽(任何狀態可印,寫 print_previews,不觸碰 print_logs)。
  $async.Future<PreviewResponse> preview(
          $pb.ClientContext? ctx, PreviewRequest request) =>
      _client.invoke<PreviewResponse>(
          ctx, 'PrintService', 'Preview', request, PreviewResponse());

  /// Print:正式列印(範圍內全 processing;首印/重印推導,見 5.5.3–5.5.4)。
  $async.Future<PrintResponse> print(
          $pb.ClientContext? ctx, PrintRequest request) =>
      _client.invoke<PrintResponse>(
          ctx, 'PrintService', 'Print', request, PrintResponse());

  /// ListLogs:列印記錄查詢(部門範圍,分頁 per_page ≤ 100)。
  $async.Future<ListLogsResponse> listLogs(
          $pb.ClientContext? ctx, ListLogsRequest request) =>
      _client.invoke<ListLogsResponse>(
          ctx, 'PrintService', 'ListLogs', request, ListLogsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
