// This is a generated file - do not edit.
//
// Generated from salesorder/v1/devices.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'devices.pb.dart' as $0;
import 'devices.pbjson.dart';

export 'devices.pb.dart';

abstract class DeviceServiceBase extends $pb.GeneratedService {
  $async.Future<$0.RegisterDeviceResponse> registerDevice(
      $pb.ServerContext ctx, $0.RegisterDeviceRequest request);
  $async.Future<$0.UnregisterDeviceResponse> unregisterDevice(
      $pb.ServerContext ctx, $0.UnregisterDeviceRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'RegisterDevice':
        return $0.RegisterDeviceRequest();
      case 'UnregisterDevice':
        return $0.UnregisterDeviceRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'RegisterDevice':
        return registerDevice(ctx, request as $0.RegisterDeviceRequest);
      case 'UnregisterDevice':
        return unregisterDevice(ctx, request as $0.UnregisterDeviceRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => DeviceServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => DeviceServiceBase$messageJson;
}
