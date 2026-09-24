//
//  Generated code. Do not modify.
//  source: salesorder/v1/fleet.proto
//

import "package:connectrpc/connect.dart" as connect;
import "fleet.pb.dart" as salesorderv1fleet;

/// FleetService:fleet 執行層首批(D32/10.1/10.4/10.12)。
/// 建檔(司機/車輛)與指派為後台動作(dept_admin 以上;rolePolicy fleet);
/// ListMyDeliveries 為**被指派司機本人**的任務清單(10.12:身分必須對應 fleet_drivers 列,
/// instance 級每列再經 OpenFGA Check —— 10.8:被指派司機可操作其 delivery、他人 403)。
abstract final class FleetService {
  /// Fully-qualified name of the FleetService service.
  static const name = 'salesorder.v1.FleetService';

  /// CreateDriver:建司機(關聯既有 users;同部門 user 不重複建)。
  static const createDriver = connect.Spec(
    '/$name/CreateDriver',
    connect.StreamType.unary,
    salesorderv1fleet.CreateDriverRequest.new,
    salesorderv1fleet.CreateDriverResponse.new,
  );

  /// CreateVehicle:建車輛(plate_no 部門內唯一,重複 → already_exists)。
  static const createVehicle = connect.Spec(
    '/$name/CreateVehicle',
    connect.StreamType.unary,
    salesorderv1fleet.CreateVehicleRequest.new,
    salesorderv1fleet.CreateVehicleResponse.new,
  );

  /// AssignDelivery:指派車次 → 司機/車輛(粒度 = route;存在即重指派,
  /// version 樂觀鎖:衝突 → failed_precondition;同交易寫稽核,tuple 經 AfterCommit 同步)。
  static const assignDelivery = connect.Spec(
    '/$name/AssignDelivery',
    connect.StreamType.unary,
    salesorderv1fleet.AssignDeliveryRequest.new,
    salesorderv1fleet.AssignDeliveryResponse.new,
  );

  /// ListMyDeliveries:我(drivers.user_id = 身分)被指派的配送清單。
  static const listMyDeliveries = connect.Spec(
    '/$name/ListMyDeliveries',
    connect.StreamType.unary,
    salesorderv1fleet.ListMyDeliveriesRequest.new,
    salesorderv1fleet.ListMyDeliveriesResponse.new,
  );
}
