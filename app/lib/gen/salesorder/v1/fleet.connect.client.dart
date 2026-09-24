//
//  Generated code. Do not modify.
//  source: salesorder/v1/fleet.proto
//

import "package:connectrpc/connect.dart" as connect;
import "fleet.pb.dart" as salesorderv1fleet;
import "fleet.connect.spec.dart" as specs;

/// FleetService:fleet 執行層首批(D32/10.1/10.4/10.12)。
/// 建檔(司機/車輛)與指派為後台動作(dept_admin 以上;rolePolicy fleet);
/// ListMyDeliveries 為**被指派司機本人**的任務清單(10.12:身分必須對應 fleet_drivers 列,
/// instance 級每列再經 OpenFGA Check —— 10.8:被指派司機可操作其 delivery、他人 403)。
extension type FleetServiceClient (connect.Transport _transport) {
  /// CreateDriver:建司機(關聯既有 users;同部門 user 不重複建)。
  Future<salesorderv1fleet.CreateDriverResponse> createDriver(
    salesorderv1fleet.CreateDriverRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.FleetService.createDriver,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateVehicle:建車輛(plate_no 部門內唯一,重複 → already_exists)。
  Future<salesorderv1fleet.CreateVehicleResponse> createVehicle(
    salesorderv1fleet.CreateVehicleRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.FleetService.createVehicle,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AssignDelivery:指派車次 → 司機/車輛(粒度 = route;存在即重指派,
  /// version 樂觀鎖:衝突 → failed_precondition;同交易寫稽核,tuple 經 AfterCommit 同步)。
  Future<salesorderv1fleet.AssignDeliveryResponse> assignDelivery(
    salesorderv1fleet.AssignDeliveryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.FleetService.assignDelivery,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListMyDeliveries:我(drivers.user_id = 身分)被指派的配送清單。
  Future<salesorderv1fleet.ListMyDeliveriesResponse> listMyDeliveries(
    salesorderv1fleet.ListMyDeliveriesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.FleetService.listMyDeliveries,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
