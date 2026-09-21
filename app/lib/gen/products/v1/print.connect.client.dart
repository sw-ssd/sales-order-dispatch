//
//  Generated code. Do not modify.
//  source: products/v1/print.proto
//

import "package:connectrpc/connect.dart" as connect;
import "print.pb.dart" as productsv1print;
import "print.connect.spec.dart" as specs;

/// PrintService:列印與預覽(09 計畫 Task 5.5.2–5.5.4)。
/// Preview 不限訂單狀態、不寫 print_logs;Print 要求範圍內訂單全 processing;
/// 重印沿用 Print(有既有記錄即重印分支,必填 reprint_reason)。
extension type PrintServiceClient (connect.Transport _transport) {
  /// Preview:預覽(任何狀態可印,寫 print_previews,不觸碰 print_logs)。
  Future<productsv1print.PreviewResponse> preview(
    productsv1print.PreviewRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PrintService.preview,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// Print:正式列印(範圍內全 processing;首印/重印推導,見 5.5.3–5.5.4)。
  Future<productsv1print.PrintResponse> print(
    productsv1print.PrintRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PrintService.print,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListLogs:列印記錄查詢(部門範圍,分頁 per_page ≤ 100)。
  Future<productsv1print.ListLogsResponse> listLogs(
    productsv1print.ListLogsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PrintService.listLogs,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
