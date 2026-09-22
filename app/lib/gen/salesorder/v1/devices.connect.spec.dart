//
//  Generated code. Do not modify.
//  source: salesorder/v1/devices.proto
//

import "package:connectrpc/connect.dart" as connect;
import "devices.pb.dart" as salesorderv1devices;

/// DeviceService:裝置 FCM token 註冊/註銷(07 計畫 Task 4.3.4)。
abstract final class DeviceService {
  /// Fully-qualified name of the DeviceService service.
  static const name = 'salesorder.v1.DeviceService';

  /// RegisterDevice:註冊裝置(冪等;換帳轉移歸屬)。
  static const registerDevice = connect.Spec(
    '/$name/RegisterDevice',
    connect.StreamType.unary,
    salesorderv1devices.RegisterDeviceRequest.new,
    salesorderv1devices.RegisterDeviceResponse.new,
  );

  /// UnregisterDevice:註銷裝置(冪等;查無不報錯)。
  static const unregisterDevice = connect.Spec(
    '/$name/UnregisterDevice',
    connect.StreamType.unary,
    salesorderv1devices.UnregisterDeviceRequest.new,
    salesorderv1devices.UnregisterDeviceResponse.new,
  );
}
