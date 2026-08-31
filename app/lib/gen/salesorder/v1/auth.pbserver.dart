// This is a generated file - do not edit.
//
// Generated from salesorder/v1/auth.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'auth.pb.dart' as $0;
import 'auth.pbjson.dart';

export 'auth.pb.dart';

abstract class AuthServiceBase extends $pb.GeneratedService {
  $async.Future<$0.LoginResponse> login(
      $pb.ServerContext ctx, $0.LoginRequest request);
  $async.Future<$0.RefreshResponse> refresh(
      $pb.ServerContext ctx, $0.RefreshRequest request);
  $async.Future<$0.LogoutResponse> logout(
      $pb.ServerContext ctx, $0.LogoutRequest request);
  $async.Future<$0.RegisterCompleteResponse> registerComplete(
      $pb.ServerContext ctx, $0.RegisterCompleteRequest request);
  $async.Future<$0.QRLoginResponse> qRLogin(
      $pb.ServerContext ctx, $0.QRLoginRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'Login':
        return $0.LoginRequest();
      case 'Refresh':
        return $0.RefreshRequest();
      case 'Logout':
        return $0.LogoutRequest();
      case 'RegisterComplete':
        return $0.RegisterCompleteRequest();
      case 'QRLogin':
        return $0.QRLoginRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'Login':
        return login(ctx, request as $0.LoginRequest);
      case 'Refresh':
        return refresh(ctx, request as $0.RefreshRequest);
      case 'Logout':
        return logout(ctx, request as $0.LogoutRequest);
      case 'RegisterComplete':
        return registerComplete(ctx, request as $0.RegisterCompleteRequest);
      case 'QRLogin':
        return qRLogin(ctx, request as $0.QRLoginRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => AuthServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => AuthServiceBase$messageJson;
}
