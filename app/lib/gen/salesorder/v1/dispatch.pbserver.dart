// This is a generated file - do not edit.
//
// Generated from salesorder/v1/dispatch.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'dispatch.pb.dart' as $0;
import 'dispatch.pbjson.dart';

export 'dispatch.pb.dart';

abstract class DispatchServiceBase extends $pb.GeneratedService {
  $async.Future<$0.AssignRouteResponse> assignRoute(
      $pb.ServerContext ctx, $0.AssignRouteRequest request);
  $async.Future<$0.ConfirmDispatchResponse> confirmDispatch(
      $pb.ServerContext ctx, $0.ConfirmDispatchRequest request);
  $async.Future<$0.CancelDispatchResponse> cancelDispatch(
      $pb.ServerContext ctx, $0.CancelDispatchRequest request);
  $async.Future<$0.BoardEvent> watchBoard(
      $pb.ServerContext ctx, $0.WatchBoardRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'AssignRoute':
        return $0.AssignRouteRequest();
      case 'ConfirmDispatch':
        return $0.ConfirmDispatchRequest();
      case 'CancelDispatch':
        return $0.CancelDispatchRequest();
      case 'WatchBoard':
        return $0.WatchBoardRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'AssignRoute':
        return assignRoute(ctx, request as $0.AssignRouteRequest);
      case 'ConfirmDispatch':
        return confirmDispatch(ctx, request as $0.ConfirmDispatchRequest);
      case 'CancelDispatch':
        return cancelDispatch(ctx, request as $0.CancelDispatchRequest);
      case 'WatchBoard':
        return watchBoard(ctx, request as $0.WatchBoardRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => DispatchServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => DispatchServiceBase$messageJson;
}
