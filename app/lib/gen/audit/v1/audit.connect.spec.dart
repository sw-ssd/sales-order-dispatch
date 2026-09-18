//
//  Generated code. Do not modify.
//  source: audit/v1/audit.proto
//

import "package:connectrpc/connect.dart" as connect;
import "audit.pb.dart" as auditv1audit;

/// AuditService:稽核查詢。
abstract final class AuditService {
  /// Fully-qualified name of the AuditService service.
  static const name = 'audit.v1.AuditService';

  /// ListAuditLogs:分頁查詢,可依時間(預設近 3 個月)/action/resource/user 篩選;時間降冪。
  static const listAuditLogs = connect.Spec(
    '/$name/ListAuditLogs',
    connect.StreamType.unary,
    auditv1audit.ListAuditLogsRequest.new,
    auditv1audit.ListAuditLogsResponse.new,
  );
}
