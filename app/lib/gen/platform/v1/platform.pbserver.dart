// This is a generated file - do not edit.
//
// Generated from platform/v1/platform.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'platform.pb.dart' as $0;
import 'platform.pbjson.dart';

export 'platform.pb.dart';

abstract class PlatformAdminServiceBase extends $pb.GeneratedService {
  $async.Future<$0.ListTenantsResponse> listTenants(
      $pb.ServerContext ctx, $0.ListTenantsRequest request);
  $async.Future<$0.GetTenantResponse> getTenant(
      $pb.ServerContext ctx, $0.GetTenantRequest request);
  $async.Future<$0.ListPlansResponse> listPlans(
      $pb.ServerContext ctx, $0.ListPlansRequest request);
  $async.Future<$0.GetPlanEntitlementsResponse> getPlanEntitlements(
      $pb.ServerContext ctx, $0.GetPlanEntitlementsRequest request);
  $async.Future<$0.ListPlatformAuditResponse> listPlatformAudit(
      $pb.ServerContext ctx, $0.ListPlatformAuditRequest request);
  $async.Future<$0.ListReceivablesResponse> listReceivables(
      $pb.ServerContext ctx, $0.ListReceivablesRequest request);
  $async.Future<$0.RecordPaymentResponse> recordPayment(
      $pb.ServerContext ctx, $0.RecordPaymentRequest request);
  $async.Future<$0.CreateSubscriptionResponse> createSubscription(
      $pb.ServerContext ctx, $0.CreateSubscriptionRequest request);
  $async.Future<$0.SetSeatCountResponse> setSeatCount(
      $pb.ServerContext ctx, $0.SetSeatCountRequest request);
  $async.Future<$0.ChangePlanResponse> changePlan(
      $pb.ServerContext ctx, $0.ChangePlanRequest request);
  $async.Future<$0.CancelSubscriptionResponse> cancelSubscription(
      $pb.ServerContext ctx, $0.CancelSubscriptionRequest request);
  $async.Future<$0.GetBillingSettingsResponse> getBillingSettings(
      $pb.ServerContext ctx, $0.GetBillingSettingsRequest request);
  $async.Future<$0.UpdateBillingSettingsResponse> updateBillingSettings(
      $pb.ServerContext ctx, $0.UpdateBillingSettingsRequest request);
  $async.Future<$0.SetTenantOverrideResponse> setTenantOverride(
      $pb.ServerContext ctx, $0.SetTenantOverrideRequest request);
  $async.Future<$0.RevokeTenantOverrideResponse> revokeTenantOverride(
      $pb.ServerContext ctx, $0.RevokeTenantOverrideRequest request);
  $async.Future<$0.UpsertPlanPriceResponse> upsertPlanPrice(
      $pb.ServerContext ctx, $0.UpsertPlanPriceRequest request);
  $async.Future<$0.SetPlanEntitlementResponse> setPlanEntitlement(
      $pb.ServerContext ctx, $0.SetPlanEntitlementRequest request);
  $async.Future<$0.CreateOperatorResponse> createOperator(
      $pb.ServerContext ctx, $0.CreateOperatorRequest request);
  $async.Future<$0.DisableOperatorResponse> disableOperator(
      $pb.ServerContext ctx, $0.DisableOperatorRequest request);
  $async.Future<$0.GetOperatorSelfResponse> getOperatorSelf(
      $pb.ServerContext ctx, $0.GetOperatorSelfRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListTenants':
        return $0.ListTenantsRequest();
      case 'GetTenant':
        return $0.GetTenantRequest();
      case 'ListPlans':
        return $0.ListPlansRequest();
      case 'GetPlanEntitlements':
        return $0.GetPlanEntitlementsRequest();
      case 'ListPlatformAudit':
        return $0.ListPlatformAuditRequest();
      case 'ListReceivables':
        return $0.ListReceivablesRequest();
      case 'RecordPayment':
        return $0.RecordPaymentRequest();
      case 'CreateSubscription':
        return $0.CreateSubscriptionRequest();
      case 'SetSeatCount':
        return $0.SetSeatCountRequest();
      case 'ChangePlan':
        return $0.ChangePlanRequest();
      case 'CancelSubscription':
        return $0.CancelSubscriptionRequest();
      case 'GetBillingSettings':
        return $0.GetBillingSettingsRequest();
      case 'UpdateBillingSettings':
        return $0.UpdateBillingSettingsRequest();
      case 'SetTenantOverride':
        return $0.SetTenantOverrideRequest();
      case 'RevokeTenantOverride':
        return $0.RevokeTenantOverrideRequest();
      case 'UpsertPlanPrice':
        return $0.UpsertPlanPriceRequest();
      case 'SetPlanEntitlement':
        return $0.SetPlanEntitlementRequest();
      case 'CreateOperator':
        return $0.CreateOperatorRequest();
      case 'DisableOperator':
        return $0.DisableOperatorRequest();
      case 'GetOperatorSelf':
        return $0.GetOperatorSelfRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListTenants':
        return listTenants(ctx, request as $0.ListTenantsRequest);
      case 'GetTenant':
        return getTenant(ctx, request as $0.GetTenantRequest);
      case 'ListPlans':
        return listPlans(ctx, request as $0.ListPlansRequest);
      case 'GetPlanEntitlements':
        return getPlanEntitlements(
            ctx, request as $0.GetPlanEntitlementsRequest);
      case 'ListPlatformAudit':
        return listPlatformAudit(ctx, request as $0.ListPlatformAuditRequest);
      case 'ListReceivables':
        return listReceivables(ctx, request as $0.ListReceivablesRequest);
      case 'RecordPayment':
        return recordPayment(ctx, request as $0.RecordPaymentRequest);
      case 'CreateSubscription':
        return createSubscription(ctx, request as $0.CreateSubscriptionRequest);
      case 'SetSeatCount':
        return setSeatCount(ctx, request as $0.SetSeatCountRequest);
      case 'ChangePlan':
        return changePlan(ctx, request as $0.ChangePlanRequest);
      case 'CancelSubscription':
        return cancelSubscription(ctx, request as $0.CancelSubscriptionRequest);
      case 'GetBillingSettings':
        return getBillingSettings(ctx, request as $0.GetBillingSettingsRequest);
      case 'UpdateBillingSettings':
        return updateBillingSettings(
            ctx, request as $0.UpdateBillingSettingsRequest);
      case 'SetTenantOverride':
        return setTenantOverride(ctx, request as $0.SetTenantOverrideRequest);
      case 'RevokeTenantOverride':
        return revokeTenantOverride(
            ctx, request as $0.RevokeTenantOverrideRequest);
      case 'UpsertPlanPrice':
        return upsertPlanPrice(ctx, request as $0.UpsertPlanPriceRequest);
      case 'SetPlanEntitlement':
        return setPlanEntitlement(ctx, request as $0.SetPlanEntitlementRequest);
      case 'CreateOperator':
        return createOperator(ctx, request as $0.CreateOperatorRequest);
      case 'DisableOperator':
        return disableOperator(ctx, request as $0.DisableOperatorRequest);
      case 'GetOperatorSelf':
        return getOperatorSelf(ctx, request as $0.GetOperatorSelfRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      PlatformAdminServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => PlatformAdminServiceBase$messageJson;
}

abstract class TenantEntitlementServiceBase extends $pb.GeneratedService {
  $async.Future<$0.GetTenantEntitlementsResponse> getTenantEntitlements(
      $pb.ServerContext ctx, $0.GetTenantEntitlementsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetTenantEntitlements':
        return $0.GetTenantEntitlementsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetTenantEntitlements':
        return getTenantEntitlements(
            ctx, request as $0.GetTenantEntitlementsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      TenantEntitlementServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => TenantEntitlementServiceBase$messageJson;
}
