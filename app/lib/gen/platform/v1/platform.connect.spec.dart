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

  /// --- 平台寫入(T9)。共同契約:每個寫入都必須帶 reason(平台稽核必填),
  /// actor 一律是 cookie 上的真實 operator;資料與稽核同一個交易(失敗不留半成品)。
  static const listReceivables = connect.Spec(
    '/$name/ListReceivables',
    connect.StreamType.unary,
    platformv1platform.ListReceivablesRequest.new,
    platformv1platform.ListReceivablesResponse.new,
  );

  static const recordPayment = connect.Spec(
    '/$name/RecordPayment',
    connect.StreamType.unary,
    platformv1platform.RecordPaymentRequest.new,
    platformv1platform.RecordPaymentResponse.new,
  );

  static const createSubscription = connect.Spec(
    '/$name/CreateSubscription',
    connect.StreamType.unary,
    platformv1platform.CreateSubscriptionRequest.new,
    platformv1platform.CreateSubscriptionResponse.new,
  );

  static const setSeatCount = connect.Spec(
    '/$name/SetSeatCount',
    connect.StreamType.unary,
    platformv1platform.SetSeatCountRequest.new,
    platformv1platform.SetSeatCountResponse.new,
  );

  static const changePlan = connect.Spec(
    '/$name/ChangePlan',
    connect.StreamType.unary,
    platformv1platform.ChangePlanRequest.new,
    platformv1platform.ChangePlanResponse.new,
  );

  static const cancelSubscription = connect.Spec(
    '/$name/CancelSubscription',
    connect.StreamType.unary,
    platformv1platform.CancelSubscriptionRequest.new,
    platformv1platform.CancelSubscriptionResponse.new,
  );

  static const getBillingSettings = connect.Spec(
    '/$name/GetBillingSettings',
    connect.StreamType.unary,
    platformv1platform.GetBillingSettingsRequest.new,
    platformv1platform.GetBillingSettingsResponse.new,
  );

  static const updateBillingSettings = connect.Spec(
    '/$name/UpdateBillingSettings',
    connect.StreamType.unary,
    platformv1platform.UpdateBillingSettingsRequest.new,
    platformv1platform.UpdateBillingSettingsResponse.new,
  );

  static const setTenantOverride = connect.Spec(
    '/$name/SetTenantOverride',
    connect.StreamType.unary,
    platformv1platform.SetTenantOverrideRequest.new,
    platformv1platform.SetTenantOverrideResponse.new,
  );

  static const revokeTenantOverride = connect.Spec(
    '/$name/RevokeTenantOverride',
    connect.StreamType.unary,
    platformv1platform.RevokeTenantOverrideRequest.new,
    platformv1platform.RevokeTenantOverrideResponse.new,
  );

  static const upsertPlanPrice = connect.Spec(
    '/$name/UpsertPlanPrice',
    connect.StreamType.unary,
    platformv1platform.UpsertPlanPriceRequest.new,
    platformv1platform.UpsertPlanPriceResponse.new,
  );

  static const setPlanEntitlement = connect.Spec(
    '/$name/SetPlanEntitlement',
    connect.StreamType.unary,
    platformv1platform.SetPlanEntitlementRequest.new,
    platformv1platform.SetPlanEntitlementResponse.new,
  );

  static const createOperator = connect.Spec(
    '/$name/CreateOperator',
    connect.StreamType.unary,
    platformv1platform.CreateOperatorRequest.new,
    platformv1platform.CreateOperatorResponse.new,
  );

  static const disableOperator = connect.Spec(
    '/$name/DisableOperator',
    connect.StreamType.unary,
    platformv1platform.DisableOperatorRequest.new,
    platformv1platform.DisableOperatorResponse.new,
  );

  /// 未結項 #23：自己的身分（console 依角色隱藏操作）。後端仍是唯一決策者；
  /// 前端只據此 disable 按鈕，不做授權判斷。
  static const getOperatorSelf = connect.Spec(
    '/$name/GetOperatorSelf',
    connect.StreamType.unary,
    platformv1platform.GetOperatorSelfRequest.new,
    platformv1platform.GetOperatorSelfResponse.new,
  );

  /// 未結項 #28：某訂閱的期別歷史（租戶詳情的「期別」段；spec §2.4）。
  static const listSubscriptionPeriods = connect.Spec(
    '/$name/ListSubscriptionPeriods',
    connect.StreamType.unary,
    platformv1platform.ListSubscriptionPeriodsRequest.new,
    platformv1platform.ListSubscriptionPeriodsResponse.new,
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
