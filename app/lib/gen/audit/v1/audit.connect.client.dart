//
//  Generated code. Do not modify.
//  source: audit/v1/audit.proto
//

import "package:connectrpc/connect.dart" as connect;
import "audit.pb.dart" as auditv1audit;
import "audit.connect.spec.dart" as specs;

/// AuditService:稽核查詢。
extension type AuditServiceClient (connect.Transport _transport) {
  /// ListAuditLogs:分頁查詢,可依時間(預設近 3 個月)/action/resource/user 篩選;時間降冪。
  Future<auditv1audit.ListAuditLogsResponse> listAuditLogs(
    auditv1audit.ListAuditLogsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuditService.listAuditLogs,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
