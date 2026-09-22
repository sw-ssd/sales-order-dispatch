//
//  Generated code. Do not modify.
//  source: salesorder/v1/devices.proto
//

import "package:connectrpc/connect.dart" as connect;
import "devices.pb.dart" as salesorderv1devices;
import "devices.connect.spec.dart" as specs;

/// DeviceService:裝置 FCM token 註冊/註銷(07 計畫 Task 4.3.4)。
extension type DeviceServiceClient (connect.Transport _transport) {
  /// RegisterDevice:註冊裝置(冪等;換帳轉移歸屬)。
  Future<salesorderv1devices.RegisterDeviceResponse> registerDevice(
    salesorderv1devices.RegisterDeviceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DeviceService.registerDevice,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UnregisterDevice:註銷裝置(冪等;查無不報錯)。
  Future<salesorderv1devices.UnregisterDeviceResponse> unregisterDevice(
    salesorderv1devices.UnregisterDeviceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DeviceService.unregisterDevice,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
