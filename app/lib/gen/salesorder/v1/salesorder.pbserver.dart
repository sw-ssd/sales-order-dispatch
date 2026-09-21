// This is a generated file - do not edit.
//
// Generated from salesorder/v1/salesorder.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'salesorder.pb.dart' as $0;
import 'salesorder.pbjson.dart';

export 'salesorder.pb.dart';

abstract class SalesOrderServiceBase extends $pb.GeneratedService {
  $async.Future<$0.ListOrdersResponse> listOrders(
      $pb.ServerContext ctx, $0.ListOrdersRequest request);
  $async.Future<$0.GetOrderResponse> getOrder(
      $pb.ServerContext ctx, $0.GetOrderRequest request);
  $async.Future<$0.CreateOrderResponse> createOrder(
      $pb.ServerContext ctx, $0.CreateOrderRequest request);
  $async.Future<$0.UpdateOrderResponse> updateOrder(
      $pb.ServerContext ctx, $0.UpdateOrderRequest request);
  $async.Future<$0.CancelOrderResponse> cancelOrder(
      $pb.ServerContext ctx, $0.CancelOrderRequest request);
  $async.Future<$0.CompleteOrderResponse> completeOrder(
      $pb.ServerContext ctx, $0.CompleteOrderRequest request);
  $async.Future<$0.VoidOrderResponse> voidOrder(
      $pb.ServerContext ctx, $0.VoidOrderRequest request);
  $async.Future<$0.DeleteOrderResponse> deleteOrder(
      $pb.ServerContext ctx, $0.DeleteOrderRequest request);
  $async.Future<$0.ListOrderEventsResponse> listOrderEvents(
      $pb.ServerContext ctx, $0.ListOrderEventsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListOrders':
        return $0.ListOrdersRequest();
      case 'GetOrder':
        return $0.GetOrderRequest();
      case 'CreateOrder':
        return $0.CreateOrderRequest();
      case 'UpdateOrder':
        return $0.UpdateOrderRequest();
      case 'CancelOrder':
        return $0.CancelOrderRequest();
      case 'CompleteOrder':
        return $0.CompleteOrderRequest();
      case 'VoidOrder':
        return $0.VoidOrderRequest();
      case 'DeleteOrder':
        return $0.DeleteOrderRequest();
      case 'ListOrderEvents':
        return $0.ListOrderEventsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListOrders':
        return listOrders(ctx, request as $0.ListOrdersRequest);
      case 'GetOrder':
        return getOrder(ctx, request as $0.GetOrderRequest);
      case 'CreateOrder':
        return createOrder(ctx, request as $0.CreateOrderRequest);
      case 'UpdateOrder':
        return updateOrder(ctx, request as $0.UpdateOrderRequest);
      case 'CancelOrder':
        return cancelOrder(ctx, request as $0.CancelOrderRequest);
      case 'CompleteOrder':
        return completeOrder(ctx, request as $0.CompleteOrderRequest);
      case 'VoidOrder':
        return voidOrder(ctx, request as $0.VoidOrderRequest);
      case 'DeleteOrder':
        return deleteOrder(ctx, request as $0.DeleteOrderRequest);
      case 'ListOrderEvents':
        return listOrderEvents(ctx, request as $0.ListOrderEventsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      SalesOrderServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => SalesOrderServiceBase$messageJson;
}
