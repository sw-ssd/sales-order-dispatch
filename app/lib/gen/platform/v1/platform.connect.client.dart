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

  /// --- 平台寫入(T9)。共同契約:每個寫入都必須帶 reason(平台稽核必填),
  /// actor 一律是 cookie 上的真實 operator;資料與稽核同一個交易(失敗不留半成品)。
  Future<platformv1platform.ListReceivablesResponse> listReceivables(
    platformv1platform.ListReceivablesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.listReceivables,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.RecordPaymentResponse> recordPayment(
    platformv1platform.RecordPaymentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.recordPayment,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.SetSeatCountResponse> setSeatCount(
    platformv1platform.SetSeatCountRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.setSeatCount,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.ChangePlanResponse> changePlan(
    platformv1platform.ChangePlanRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.changePlan,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.CancelSubscriptionResponse> cancelSubscription(
    platformv1platform.CancelSubscriptionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.cancelSubscription,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.GetBillingSettingsResponse> getBillingSettings(
    platformv1platform.GetBillingSettingsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.getBillingSettings,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.UpdateBillingSettingsResponse> updateBillingSettings(
    platformv1platform.UpdateBillingSettingsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.updateBillingSettings,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.SetTenantOverrideResponse> setTenantOverride(
    platformv1platform.SetTenantOverrideRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.setTenantOverride,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.RevokeTenantOverrideResponse> revokeTenantOverride(
    platformv1platform.RevokeTenantOverrideRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.revokeTenantOverride,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.UpsertPlanPriceResponse> upsertPlanPrice(
    platformv1platform.UpsertPlanPriceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.upsertPlanPrice,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.SetPlanEntitlementResponse> setPlanEntitlement(
    platformv1platform.SetPlanEntitlementRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.setPlanEntitlement,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.CreateOperatorResponse> createOperator(
    platformv1platform.CreateOperatorRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.createOperator,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<platformv1platform.DisableOperatorResponse> disableOperator(
    platformv1platform.DisableOperatorRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PlatformAdminService.disableOperator,
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
