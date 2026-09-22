//
//  Generated code. Do not modify.
//  source: salesorder/v1/notifications.proto
//

import "package:connectrpc/connect.dart" as connect;
import "notifications.pb.dart" as salesorderv1notifications;
import "notifications.connect.spec.dart" as specs;

/// NotificationService:通知中心(07 計畫 Task 4.3.3)。
/// 僅回傳當前使用者本人的通知;記錄不可刪除(規格 §5.4)。
extension type NotificationServiceClient (connect.Transport _transport) {
  /// ListNotifications:本人通知列表(分頁 per_page ≤ 100;unread_only 篩未讀)。
  Future<salesorderv1notifications.ListNotificationsResponse> listNotifications(
    salesorderv1notifications.ListNotificationsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotificationService.listNotifications,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// MarkRead:標記已讀(僅本人;sent/pending → read;failed 不可轉;冪等)。
  Future<salesorderv1notifications.MarkReadResponse> markRead(
    salesorderv1notifications.MarkReadRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotificationService.markRead,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UnreadCount:未讀數(通知鈴角標輪詢)。
  Future<salesorderv1notifications.UnreadCountResponse> unreadCount(
    salesorderv1notifications.UnreadCountRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotificationService.unreadCount,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
