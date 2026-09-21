//
//  Generated code. Do not modify.
//  source: salesorder/v1/salesorder.proto
//

import "package:connectrpc/connect.dart" as connect;
import "salesorder.pb.dart" as salesorderv1salesorder;

/// SalesOrderService:銷售訂單管理。
abstract final class SalesOrderService {
  /// Fully-qualified name of the SalesOrderService service.
  static const name = 'salesorder.v1.SalesOrderService';

  /// ListOrders:分頁查詢(預設排除軟刪除;status/customer_id/source/keyword 篩選)。
  static const listOrders = connect.Spec(
    '/$name/ListOrders',
    connect.StreamType.unary,
    salesorderv1salesorder.ListOrdersRequest.new,
    salesorderv1salesorder.ListOrdersResponse.new,
  );

  /// GetOrder:以 id 取單筆(含明細)。
  static const getOrder = connect.Spec(
    '/$name/GetOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.GetOrderRequest.new,
    salesorderv1salesorder.GetOrderResponse.new,
  );

  /// CreateOrder:建立訂單(最小可用:客戶+來源+明細;取號+建單同交易)。
  static const createOrder = connect.Spec(
    '/$name/CreateOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.CreateOrderRequest.new,
    salesorderv1salesorder.CreateOrderResponse.new,
  );

  /// UpdateOrder:僅 pending 可編輯(攜帶 version 樂觀鎖)。
  static const updateOrder = connect.Spec(
    '/$name/UpdateOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.UpdateOrderRequest.new,
    salesorderv1salesorder.UpdateOrderResponse.new,
  );

  /// CancelOrder:僅 pending 可取消。
  static const cancelOrder = connect.Spec(
    '/$name/CancelOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.CancelOrderRequest.new,
    salesorderv1salesorder.CancelOrderResponse.new,
  );

  /// CompleteOrder:僅 processing 可完成。
  static const completeOrder = connect.Spec(
    '/$name/CompleteOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.CompleteOrderRequest.new,
    salesorderv1salesorder.CompleteOrderResponse.new,
  );

  /// VoidOrder:僅 completed 可作廢(dept_admin 以上 + 原因)。
  static const voidOrder = connect.Spec(
    '/$name/VoidOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.VoidOrderRequest.new,
    salesorderv1salesorder.VoidOrderResponse.new,
  );

  /// DeleteOrder:軟刪除(僅 pending/cancelled)。
  static const deleteOrder = connect.Spec(
    '/$name/DeleteOrder',
    connect.StreamType.unary,
    salesorderv1salesorder.DeleteOrderRequest.new,
    salesorderv1salesorder.DeleteOrderResponse.new,
  );

  /// ListOrderEvents:異動軌跡查詢(升序)。
  static const listOrderEvents = connect.Spec(
    '/$name/ListOrderEvents',
    connect.StreamType.unary,
    salesorderv1salesorder.ListOrderEventsRequest.new,
    salesorderv1salesorder.ListOrderEventsResponse.new,
  );
}
