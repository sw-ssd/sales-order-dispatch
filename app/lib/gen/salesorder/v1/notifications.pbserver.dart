// This is a generated file - do not edit.
//
// Generated from salesorder/v1/notifications.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'notifications.pb.dart' as $0;
import 'notifications.pbjson.dart';

export 'notifications.pb.dart';

abstract class NotificationServiceBase extends $pb.GeneratedService {
  $async.Future<$0.ListNotificationsResponse> listNotifications(
      $pb.ServerContext ctx, $0.ListNotificationsRequest request);
  $async.Future<$0.MarkReadResponse> markRead(
      $pb.ServerContext ctx, $0.MarkReadRequest request);
  $async.Future<$0.UnreadCountResponse> unreadCount(
      $pb.ServerContext ctx, $0.UnreadCountRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListNotifications':
        return $0.ListNotificationsRequest();
      case 'MarkRead':
        return $0.MarkReadRequest();
      case 'UnreadCount':
        return $0.UnreadCountRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListNotifications':
        return listNotifications(ctx, request as $0.ListNotificationsRequest);
      case 'MarkRead':
        return markRead(ctx, request as $0.MarkReadRequest);
      case 'UnreadCount':
        return unreadCount(ctx, request as $0.UnreadCountRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      NotificationServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => NotificationServiceBase$messageJson;
}
