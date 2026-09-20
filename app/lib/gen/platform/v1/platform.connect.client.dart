//
//  Generated code. Do not modify.
//  source: platform/v1/platform.proto
//

import "package:connectrpc/connect.dart" as connect;
import "platform.pb.dart" as platformv1platform;
import "platform.connect.spec.dart" as specs;

/// PlatformAdminService:平台營運(operator session;租戶身分一律拒絕)。
extension type PlatformAdminServiceClient (connect.Transport _transport) {
  Future<platformv1platform.ListTenantsResponse> listTenants(
    platformv1platform.ListTenantsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.listTenants,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.GetTenantResponse> getTenant(
    platformv1platform.GetTenantRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.getTenant,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.ListPlansResponse> listPlans(
    platformv1platform.ListPlansRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.listPlans,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.GetPlanEntitlementsResponse> getPlanEntitlements(
    platformv1platform.GetPlanEntitlementsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.getPlanEntitlements,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.ListPlatformAuditResponse> listPlatformAudit(
    platformv1platform.ListPlatformAuditRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.listPlatformAudit,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// TenantEntitlementService:租戶端權益投影(租戶 session;唯讀)。
extension type TenantEntitlementServiceClient (connect.Transport _transport) {
  Future<platformv1platform.GetTenantEntitlementsResponse> getTenantEntitlements(
    platformv1platform.GetTenantEntitlementsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TenantEntitlementService.getTenantEntitlements,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
