//
//  Generated code. Do not modify.
//  source: products/v1/print.proto
//

import "package:connectrpc/connect.dart" as connect;
import "print.pb.dart" as productsv1print;

/// PrintService:列印與預覽(09 計畫 Task 5.5.2–5.5.4)。
/// Preview 不限訂單狀態、不寫 print_logs;Print 要求範圍內訂單全 processing;
/// 重印沿用 Print(有既有記錄即重印分支,必填 reprint_reason)。
abstract final class PrintService {
  /// Fully-qualified name of the PrintService service.
  static const name = 'products.v1.PrintService';

  /// Preview:預覽(任何狀態可印,寫 print_previews,不觸碰 print_logs)。
  static const preview = connect.Spec(
    '/$name/Preview',
    connect.StreamType.unary,
    productsv1print.PreviewRequest.new,
    productsv1print.PreviewResponse.new,
  );

  /// Print:正式列印(範圍內全 processing;首印/重印推導,見 5.5.3–5.5.4)。
  static const print = connect.Spec(
    '/$name/Print',
    connect.StreamType.unary,
    productsv1print.PrintRequest.new,
    productsv1print.PrintResponse.new,
  );

  /// ListLogs:列印記錄查詢(部門範圍,分頁 per_page ≤ 100)。
  static const listLogs = connect.Spec(
    '/$name/ListLogs',
    connect.StreamType.unary,
    productsv1print.ListLogsRequest.new,
    productsv1print.ListLogsResponse.new,
  );
}
