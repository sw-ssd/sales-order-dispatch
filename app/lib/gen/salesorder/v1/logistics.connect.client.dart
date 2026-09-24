//
//  Generated code. Do not modify.
//  source: salesorder/v1/logistics.proto
//

import "package:connectrpc/connect.dart" as connect;
import "logistics.pb.dart" as salesorderv1logistics;
import "logistics.connect.spec.dart" as specs;

/// LogisticsService:logistics 執行層首批(D32/10.1/10.4/10.12)。
/// 建檔(司機/車輛)與指派為後台動作(dept_admin 以上;rolePolicy logistics);
/// ListMyDeliveries 為**被指派司機本人**的任務清單(10.12:身分必須對應 logistics_drivers 列,
/// instance 級每列再經 OpenFGA Check —— 10.8:被指派司機可操作其 delivery、他人 403)。
extension type LogisticsServiceClient (connect.Transport _transport) {
  /// CreateDriver:建司機(關聯既有 users;同部門 user 不重複建)。
  Future<salesorderv1logistics.CreateDriverResponse> createDriver(
    salesorderv1logistics.CreateDriverRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.createDriver,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateVehicle:建車輛(plate_no 部門內唯一,重複 → already_exists)。
  Future<salesorderv1logistics.CreateVehicleResponse> createVehicle(
    salesorderv1logistics.CreateVehicleRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.createVehicle,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AssignDelivery:指派車次 → 司機/車輛(粒度 = route;存在即重指派,
  /// version 樂觀鎖:衝突 → failed_precondition;同交易寫稽核,tuple 經 AfterCommit 同步)。
  Future<salesorderv1logistics.AssignDeliveryResponse> assignDelivery(
    salesorderv1logistics.AssignDeliveryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.assignDelivery,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListMyDeliveries:我(drivers.user_id = 身分)被指派的配送清單。
  Future<salesorderv1logistics.ListMyDeliveriesResponse> listMyDeliveries(
    salesorderv1logistics.ListMyDeliveriesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.listMyDeliveries,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// StartDelivery:被指派司機開始執行(pending → in_progress;10.6)。
  Future<salesorderv1logistics.StartDeliveryResponse> startDelivery(
    salesorderv1logistics.StartDeliveryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.startDelivery,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CompleteDelivery:完成並簽收(in_progress → completed;POD 可多筆,同一交易寫事件與稽核)。
  Future<salesorderv1logistics.CompleteDeliveryResponse> completeDelivery(
    salesorderv1logistics.CompleteDeliveryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.completeDelivery,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelDelivery:取消配送(pending/in_progress → cancelled;reason 必填)。
  Future<salesorderv1logistics.CancelDeliveryResponse> cancelDelivery(
    salesorderv1logistics.CancelDeliveryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LogisticsService.cancelDelivery,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
