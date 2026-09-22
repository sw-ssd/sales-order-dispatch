//
//  Generated code. Do not modify.
//  source: salesorder/v1/notifications.proto
//

import "package:connectrpc/connect.dart" as connect;
import "notifications.pb.dart" as salesorderv1notifications;

/// NotificationService:通知中心(07 計畫 Task 4.3.3)。
/// 僅回傳當前使用者本人的通知;記錄不可刪除(規格 §5.4)。
abstract final class NotificationService {
  /// Fully-qualified name of the NotificationService service.
  static const name = 'salesorder.v1.NotificationService';

  /// ListNotifications:本人通知列表(分頁 per_page ≤ 100;unread_only 篩未讀)。
  static const listNotifications = connect.Spec(
    '/$name/ListNotifications',
    connect.StreamType.unary,
    salesorderv1notifications.ListNotificationsRequest.new,
    salesorderv1notifications.ListNotificationsResponse.new,
  );

  /// MarkRead:標記已讀(僅本人;sent/pending → read;failed 不可轉;冪等)。
  static const markRead = connect.Spec(
    '/$name/MarkRead',
    connect.StreamType.unary,
    salesorderv1notifications.MarkReadRequest.new,
    salesorderv1notifications.MarkReadResponse.new,
  );

  /// UnreadCount:未讀數(通知鈴角標輪詢)。
  static const unreadCount = connect.Spec(
    '/$name/UnreadCount',
    connect.StreamType.unary,
    salesorderv1notifications.UnreadCountRequest.new,
    salesorderv1notifications.UnreadCountResponse.new,
  );
}
