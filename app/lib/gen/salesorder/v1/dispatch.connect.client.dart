//
//  Generated code. Do not modify.
//  source: salesorder/v1/dispatch.proto
//

import "package:connectrpc/connect.dart" as connect;
import "dispatch.pb.dart" as salesorderv1dispatch;
import "dispatch.connect.spec.dart" as specs;

/// DispatchService:派車 API(08 計畫 Task 5.1)。
/// AssignRoute 走看板拖放(樂觀鎖);Confirm 批次轉 processing(逐筆交易);
/// CancelDispatch 退回 pending(保留看板位置);WatchBoard 串流另案(5.2)。
extension type DispatchServiceClient (connect.Transport _transport) {
  /// AssignRoute:指派車次與配送順位(僅 pending;version 樂觀鎖)。
  Future<salesorderv1dispatch.AssignRouteResponse> assignRoute(
    salesorderv1dispatch.AssignRouteRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DispatchService.assignRoute,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ConfirmDispatch:車次批次確認(逐筆 pending → processing;部分失敗語義)。
  Future<salesorderv1dispatch.ConfirmDispatchResponse> confirmDispatch(
    salesorderv1dispatch.ConfirmDispatchRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DispatchService.confirmDispatch,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelDispatch:取消派車(僅 processing → pending;dept_admin 以上;重印警告)。
  Future<salesorderv1dispatch.CancelDispatchResponse> cancelDispatch(
    salesorderv1dispatch.CancelDispatchRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DispatchService.cancelDispatch,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// WatchBoard:看板訂閱(server streaming;部門隔離;heartbeat 保活)。
  Stream<salesorderv1dispatch.BoardEvent> watchBoard(
    salesorderv1dispatch.WatchBoardRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).server(
      specs.DispatchService.watchBoard,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
