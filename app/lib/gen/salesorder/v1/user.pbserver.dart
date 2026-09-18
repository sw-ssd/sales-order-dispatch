// This is a generated file - do not edit.
//
// Generated from salesorder/v1/user.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'user.pb.dart' as $1;
import 'user.pbjson.dart';

export 'user.pb.dart';

abstract class UserServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListUsersResponse> listUsers(
      $pb.ServerContext ctx, $1.ListUsersRequest request);
  $async.Future<$1.GetUserResponse> getUser(
      $pb.ServerContext ctx, $1.GetUserRequest request);
  $async.Future<$1.CreateUserResponse> createUser(
      $pb.ServerContext ctx, $1.CreateUserRequest request);
  $async.Future<$1.UpdateUserResponse> updateUser(
      $pb.ServerContext ctx, $1.UpdateUserRequest request);
  $async.Future<$1.AssignRoleResponse> assignRole(
      $pb.ServerContext ctx, $1.AssignRoleRequest request);
  $async.Future<$1.DeactivateResponse> deactivate(
      $pb.ServerContext ctx, $1.DeactivateRequest request);
  $async.Future<$1.ForceLogoutResponse> forceLogout(
      $pb.ServerContext ctx, $1.ForceLogoutRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListUsers':
        return $1.ListUsersRequest();
      case 'GetUser':
        return $1.GetUserRequest();
      case 'CreateUser':
        return $1.CreateUserRequest();
      case 'UpdateUser':
        return $1.UpdateUserRequest();
      case 'AssignRole':
        return $1.AssignRoleRequest();
      case 'Deactivate':
        return $1.DeactivateRequest();
      case 'ForceLogout':
        return $1.ForceLogoutRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListUsers':
        return listUsers(ctx, request as $1.ListUsersRequest);
      case 'GetUser':
        return getUser(ctx, request as $1.GetUserRequest);
      case 'CreateUser':
        return createUser(ctx, request as $1.CreateUserRequest);
      case 'UpdateUser':
        return updateUser(ctx, request as $1.UpdateUserRequest);
      case 'AssignRole':
        return assignRole(ctx, request as $1.AssignRoleRequest);
      case 'Deactivate':
        return deactivate(ctx, request as $1.DeactivateRequest);
      case 'ForceLogout':
        return forceLogout(ctx, request as $1.ForceLogoutRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => UserServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => UserServiceBase$messageJson;
}
