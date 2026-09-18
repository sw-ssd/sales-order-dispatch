//
//  Generated code. Do not modify.
//  source: metadict/v1/metadict.proto
//

import "package:connectrpc/connect.dart" as connect;
import "metadict.pb.dart" as metadictv1metadict;
import "metadict.connect.spec.dart" as specs;

/// MetadictService:字典管理。
extension type MetadictServiceClient (connect.Transport _transport) {
  /// ListMetadicts:分頁查詢(系統預設 + 當前部門擴充合併),可依 type 篩選。
  Future<metadictv1metadict.ListMetadictsResponse> listMetadicts(
    metadictv1metadict.ListMetadictsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MetadictService.listMetadicts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetMetadict:取得單一字典。
  Future<metadictv1metadict.GetMetadictResponse> getMetadict(
    metadictv1metadict.GetMetadictRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MetadictService.getMetadict,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateMetadict:建立字典(super 建系統級;dept_admin/staff 自動帶當前部門,不接受請求帶 department_id)。
  Future<metadictv1metadict.CreateMetadictResponse> createMetadict(
    metadictv1metadict.CreateMetadictRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MetadictService.createMetadict,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateMetadict:更新 display_name / sort_order / is_active(不含 type / code)。
  Future<metadictv1metadict.UpdateMetadictResponse> updateMetadict(
    metadictv1metadict.UpdateMetadictRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MetadictService.updateMetadict,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteMetadict:軟刪除(order_source 不可刪)。
  Future<metadictv1metadict.DeleteMetadictResponse> deleteMetadict(
    metadictv1metadict.DeleteMetadictRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MetadictService.deleteMetadict,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListOptions:表單下拉選項(僅可選用啟用值;客戶端身分僅回系統預設)。
  Future<metadictv1metadict.ListOptionsResponse> listOptions(
    metadictv1metadict.ListOptionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MetadictService.listOptions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
