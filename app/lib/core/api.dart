import 'package:connectrpc/connect.dart' as connect;

import '../features/auth/auth_repository.dart';
import '../features/auth/auth_transport.dart';
import '../features/notifications/push_registration.dart';
import '../gen/customers/v1/customer.connect.client.dart';
import '../gen/salesorder/v1/devices.connect.client.dart';
import '../gen/salesorder/v1/announcement.connect.client.dart';
import '../gen/salesorder/v1/notifications.connect.client.dart';
import '../gen/salesorder/v1/returns.connect.client.dart';
import '../gen/salesorder/v1/salesorder.connect.client.dart';

/// 已認證 API 入口：單一 transport（Bearer + 401 單飛 refresh 重試）＋
/// 各業務服務客戶端（延遲建立，共用同一 transport）。
///
/// 由 app.dart 建立、經 AppRouter 以建構式注入各頁（與 AuthRepository 同模式）。
class Api {
  Api({
    required String baseUrl,
    required AuthRepository auth,
    connect.Transport? transport,
  }) : _transport = transport ??
            createAuthedTransport(
              baseUrl: baseUrl,
              accessToken: auth.currentAccessToken,
              refresh: auth.refreshTokens,
            ) {
    // 推播註冊/註銷掛在 AuthRepository 的 session 生命週期上（見該檔欄位說明）:
    // 登入成功即註冊、登出/改密碼即註銷。缺 Firebase 原生設定檔時整段靜默跳過。
    final push = PushRegistration(
      api: this,
      token: auth.currentAccessToken,
    );
    auth.onSessionEstablished = push.register;
    auth.onSessionCleared = push.unregister;
  }

  final connect.Transport _transport;

  late final SalesOrderServiceClient orders =
      SalesOrderServiceClient(_transport);
  late final ReturnServiceClient returns =
      ReturnServiceClient(_transport);
  late final NotificationServiceClient notifications =
      NotificationServiceClient(_transport);
  /// 公告前台（spec announcements「前台展示與排序」）：App 首頁讀當下可見且投放 App 的公告。
  late final AnnouncementServiceClient announcements =
      AnnouncementServiceClient(_transport);
  /// 裝置 FCM token 註冊/註銷(07 Task 4.3.4):App 推播的 App 端入口。
  late final DeviceServiceClient devices = DeviceServiceClient(_transport);
  /// 店家自助帳號管理(D22/規格 4.2):僅客戶**主帳號**可呼叫;子帳號與員工一律被拒。
  late final CustomerAccountServiceClient accounts =
      CustomerAccountServiceClient(_transport);
}
