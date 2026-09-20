//
//  Generated code. Do not modify.
//  source: platform/v1/platform.proto
//

import "package:connectrpc/connect.dart" as connect;
import "platform.pb.dart" as platformv1platform;

/// PlatformAdminService:平台營運(operator session;租戶身分一律拒絕)。
abstract final class PlatformAdminService {
  /// Fully-qualified name of the PlatformAdminService service.
  static const name = 'platform.v1.PlatformAdminService';

  static const listTenants = connect.Spec(
    '/$name/ListTenants',
    connect.StreamType.unary,
    platformv1platform.ListTenantsRequest.new,
    platformv1platform.ListTenantsResponse.new,
  );

  static const getTenant = connect.Spec(
    '/$name/GetTenant',
    connect.StreamType.unary,
    platformv1platform.GetTenantRequest.new,
    platformv1platform.GetTenantResponse.new,
  );

  static const listPlans = connect.Spec(
    '/$name/ListPlans',
    connect.StreamType.unary,
    platformv1platform.ListPlansRequest.new,
    platformv1platform.ListPlansResponse.new,
  );

  static const getPlanEntitlements = connect.Spec(
    '/$name/GetPlanEntitlements',
    connect.StreamType.unary,
    platformv1platform.GetPlanEntitlementsRequest.new,
    platformv1platform.GetPlanEntitlementsResponse.new,
  );

  static const listPlatformAudit = connect.Spec(
    '/$name/ListPlatformAudit',
    connect.StreamType.unary,
    platformv1platform.ListPlatformAuditRequest.new,
    platformv1platform.ListPlatformAuditResponse.new,
  );
}
/// TenantEntitlementService:租戶端權益投影(租戶 session;唯讀)。
abstract final class TenantEntitlementService {
  /// Fully-qualified name of the TenantEntitlementService service.
  static const name = 'platform.v1.TenantEntitlementService';

  static const getTenantEntitlements = connect.Spec(
    '/$name/GetTenantEntitlements',
    connect.StreamType.unary,
    platformv1platform.GetTenantEntitlementsRequest.new,
    platformv1platform.GetTenantEntitlementsResponse.new,
  );
}
