//
//  Generated code. Do not modify.
//  source: salesorder/v1/returns.proto
//

import "package:connectrpc/connect.dart" as connect;
import "returns.pb.dart" as salesorderv1returns;
import "returns.connect.spec.dart" as specs;

/// ReturnService:退貨申請與審核(06 計畫 Task 4.7.2–4.7.4, D25)。
/// 發起僅客戶子帳號(主帳號一律拒絕);審核僅主責業務/dept_admin 以上;
/// 全程不修改原訂單(僅參照)。
extension type ReturnServiceClient (connect.Transport _transport) {
  /// CreateReturnRequest:發起退貨(子帳號;雙來源並存;同交易寫申請+明細+稽核)。
  Future<salesorderv1returns.CreateReturnRequestResponse> createReturnRequest(
    salesorderv1returns.CreateReturnRequestRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ReturnService.createReturnRequest,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListReturnRequests:申請列表(客戶 self 限自己客戶;staff 依範圍;分頁 per_page ≤ 100)。
  Future<salesorderv1returns.ListReturnRequestsResponse> listReturnRequests(
    salesorderv1returns.ListReturnRequestsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ReturnService.listReturnRequests,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetReturnRequest:單筆含品項明細(照片轉下載 URL)。
  Future<salesorderv1returns.GetReturnRequestResponse> getReturnRequest(
    salesorderv1returns.GetReturnRequestRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ReturnService.getReturnRequest,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ReviewReturnRequest:審核(approved/rejected + 樂觀鎖;同交易寫稽核;不碰原訂單)。
  Future<salesorderv1returns.ReviewReturnRequestResponse> reviewReturnRequest(
    salesorderv1returns.ReviewReturnRequestRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ReturnService.reviewReturnRequest,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetReturnCertificate:退貨證明(僅 approved;快照內容;唯讀不寫稽核)。
  Future<salesorderv1returns.GetReturnCertificateResponse> getReturnCertificate(
    salesorderv1returns.GetReturnCertificateRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ReturnService.getReturnCertificate,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
