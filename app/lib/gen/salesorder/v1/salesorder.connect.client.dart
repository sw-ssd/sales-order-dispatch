//
//  Generated code. Do not modify.
//  source: salesorder/v1/salesorder.proto
//

import "package:connectrpc/connect.dart" as connect;
import "salesorder.pb.dart" as salesorderv1salesorder;
import "salesorder.connect.spec.dart" as specs;

/// SalesOrderService:銷售訂單管理。
extension type SalesOrderServiceClient (connect.Transport _transport) {
  /// ListOrders:分頁查詢(預設排除軟刪除;status/customer_id/source/keyword 篩選)。
  Future<salesorderv1salesorder.ListOrdersResponse> listOrders(
    salesorderv1salesorder.ListOrdersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.listOrders,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetOrder:以 id 取單筆(含明細)。
  Future<salesorderv1salesorder.GetOrderResponse> getOrder(
    salesorderv1salesorder.GetOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.getOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateOrder:建立訂單(最小可用:客戶+來源+明細;取號+建單同交易)。
  Future<salesorderv1salesorder.CreateOrderResponse> createOrder(
    salesorderv1salesorder.CreateOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.createOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateOrder:僅 pending 可編輯(攜帶 version 樂觀鎖)。
  Future<salesorderv1salesorder.UpdateOrderResponse> updateOrder(
    salesorderv1salesorder.UpdateOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.updateOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelOrder:僅 pending 可取消。
  Future<salesorderv1salesorder.CancelOrderResponse> cancelOrder(
    salesorderv1salesorder.CancelOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.cancelOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CompleteOrder:僅 processing 可完成。
  Future<salesorderv1salesorder.CompleteOrderResponse> completeOrder(
    salesorderv1salesorder.CompleteOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.completeOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// VoidOrder:僅 completed 可作廢(dept_admin 以上 + 原因)。
  Future<salesorderv1salesorder.VoidOrderResponse> voidOrder(
    salesorderv1salesorder.VoidOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.voidOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteOrder:軟刪除(僅 pending/cancelled)。
  Future<salesorderv1salesorder.DeleteOrderResponse> deleteOrder(
    salesorderv1salesorder.DeleteOrderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.deleteOrder,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListOrderEvents:異動軌跡查詢(升序)。
  Future<salesorderv1salesorder.ListOrderEventsResponse> listOrderEvents(
    salesorderv1salesorder.ListOrderEventsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SalesOrderService.listOrderEvents,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
