//
//  Generated code. Do not modify.
//  source: salesorder/v1/dispatch.proto
//

import "package:connectrpc/connect.dart" as connect;
import "dispatch.pb.dart" as salesorderv1dispatch;

/// DispatchService:派車 API(08 計畫 Task 5.1)。
/// AssignRoute 走看板拖放(樂觀鎖);Confirm 批次轉 processing(逐筆交易);
/// CancelDispatch 退回 pending(保留看板位置);WatchBoard 串流另案(5.2)。
abstract final class DispatchService {
  /// Fully-qualified name of the DispatchService service.
  static const name = 'salesorder.v1.DispatchService';

  /// AssignRoute:指派車次與配送順位(僅 pending;version 樂觀鎖)。
  static const assignRoute = connect.Spec(
    '/$name/AssignRoute',
    connect.StreamType.unary,
    salesorderv1dispatch.AssignRouteRequest.new,
    salesorderv1dispatch.AssignRouteResponse.new,
  );

  /// ConfirmDispatch:車次批次確認(逐筆 pending → processing;部分失敗語義)。
  static const confirmDispatch = connect.Spec(
    '/$name/ConfirmDispatch',
    connect.StreamType.unary,
    salesorderv1dispatch.ConfirmDispatchRequest.new,
    salesorderv1dispatch.ConfirmDispatchResponse.new,
  );

  /// CancelDispatch:取消派車(僅 processing → pending;dept_admin 以上;重印警告)。
  static const cancelDispatch = connect.Spec(
    '/$name/CancelDispatch',
    connect.StreamType.unary,
    salesorderv1dispatch.CancelDispatchRequest.new,
    salesorderv1dispatch.CancelDispatchResponse.new,
  );
}
