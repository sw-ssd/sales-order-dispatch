//
//  Generated code. Do not modify.
//  source: salesorder/v1/logistics.proto
//

import "package:connectrpc/connect.dart" as connect;
import "logistics.pb.dart" as salesorderv1logistics;

/// LogisticsService:logistics 執行層首批(D32/10.1/10.4/10.12)。
/// 建檔(司機/車輛)與指派為後台動作(dept_admin 以上;rolePolicy logistics);
/// ListMyDeliveries 為**被指派司機本人**的任務清單(10.12:身分必須對應 logistics_drivers 列,
/// instance 級每列再經 OpenFGA Check —— 10.8:被指派司機可操作其 delivery、他人 403)。
abstract final class LogisticsService {
  /// Fully-qualified name of the LogisticsService service.
  static const name = 'salesorder.v1.LogisticsService';

  /// CreateDriver:建司機(關聯既有 users;同部門 user 不重複建)。
  static const createDriver = connect.Spec(
    '/$name/CreateDriver',
    connect.StreamType.unary,
    salesorderv1logistics.CreateDriverRequest.new,
    salesorderv1logistics.CreateDriverResponse.new,
  );

  /// CreateVehicle:建車輛(plate_no 部門內唯一,重複 → already_exists)。
  static const createVehicle = connect.Spec(
    '/$name/CreateVehicle',
    connect.StreamType.unary,
    salesorderv1logistics.CreateVehicleRequest.new,
    salesorderv1logistics.CreateVehicleResponse.new,
  );

  /// AssignDelivery:指派車次 → 司機/車輛(粒度 = route;存在即重指派,
  /// version 樂觀鎖:衝突 → failed_precondition;同交易寫稽核,tuple 經 AfterCommit 同步)。
  static const assignDelivery = connect.Spec(
    '/$name/AssignDelivery',
    connect.StreamType.unary,
    salesorderv1logistics.AssignDeliveryRequest.new,
    salesorderv1logistics.AssignDeliveryResponse.new,
  );

  /// ListMyDeliveries:我(drivers.user_id = 身分)被指派的配送清單。
  static const listMyDeliveries = connect.Spec(
    '/$name/ListMyDeliveries',
    connect.StreamType.unary,
    salesorderv1logistics.ListMyDeliveriesRequest.new,
    salesorderv1logistics.ListMyDeliveriesResponse.new,
  );
}
