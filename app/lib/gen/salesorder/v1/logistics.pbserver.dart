// This is a generated file - do not edit.
//
// Generated from salesorder/v1/logistics.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'logistics.pb.dart' as $0;
import 'logistics.pbjson.dart';

export 'logistics.pb.dart';

abstract class LogisticsServiceBase extends $pb.GeneratedService {
  $async.Future<$0.CreateDriverResponse> createDriver(
      $pb.ServerContext ctx, $0.CreateDriverRequest request);
  $async.Future<$0.CreateVehicleResponse> createVehicle(
      $pb.ServerContext ctx, $0.CreateVehicleRequest request);
  $async.Future<$0.AssignDeliveryResponse> assignDelivery(
      $pb.ServerContext ctx, $0.AssignDeliveryRequest request);
  $async.Future<$0.ListMyDeliveriesResponse> listMyDeliveries(
      $pb.ServerContext ctx, $0.ListMyDeliveriesRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateDriver':
        return $0.CreateDriverRequest();
      case 'CreateVehicle':
        return $0.CreateVehicleRequest();
      case 'AssignDelivery':
        return $0.AssignDeliveryRequest();
      case 'ListMyDeliveries':
        return $0.ListMyDeliveriesRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateDriver':
        return createDriver(ctx, request as $0.CreateDriverRequest);
      case 'CreateVehicle':
        return createVehicle(ctx, request as $0.CreateVehicleRequest);
      case 'AssignDelivery':
        return assignDelivery(ctx, request as $0.AssignDeliveryRequest);
      case 'ListMyDeliveries':
        return listMyDeliveries(ctx, request as $0.ListMyDeliveriesRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => LogisticsServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => LogisticsServiceBase$messageJson;
}
