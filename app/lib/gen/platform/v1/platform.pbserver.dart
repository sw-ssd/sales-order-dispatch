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
