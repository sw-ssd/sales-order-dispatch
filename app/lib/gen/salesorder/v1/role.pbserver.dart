// This is a generated file - do not edit.
//
// Generated from salesorder/v1/role.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'role.pb.dart' as $2;
import 'role.pbjson.dart';

export 'role.pb.dart';

abstract class RoleServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListRolesResponse> listRoles(
      $pb.ServerContext ctx, $2.ListRolesRequest request);
  $async.Future<$2.GetRolePermissionsResponse> getRolePermissions(
      $pb.ServerContext ctx, $2.GetRolePermissionsRequest request);
  $async.Future<$2.UpdateRolePermissionsResponse> updateRolePermissions(
      $pb.ServerContext ctx, $2.UpdateRolePermissionsRequest request);
  $async.Future<$2.ListConditionFieldsResponse> listConditionFields(
      $pb.ServerContext ctx, $2.ListConditionFieldsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListRoles':
        return $2.ListRolesRequest();
      case 'GetRolePermissions':
        return $2.GetRolePermissionsRequest();
      case 'UpdateRolePermissions':
        return $2.UpdateRolePermissionsRequest();
      case 'ListConditionFields':
        return $2.ListConditionFieldsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListRoles':
        return listRoles(ctx, request as $2.ListRolesRequest);
      case 'GetRolePermissions':
        return getRolePermissions(ctx, request as $2.GetRolePermissionsRequest);
      case 'UpdateRolePermissions':
        return updateRolePermissions(
            ctx, request as $2.UpdateRolePermissionsRequest);
      case 'ListConditionFields':
        return listConditionFields(
            ctx, request as $2.ListConditionFieldsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => RoleServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => RoleServiceBase$messageJson;
}
