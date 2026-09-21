// This is a generated file - do not edit.
//
// Generated from salesorder/v1/returns.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'returns.pb.dart' as $0;
import 'returns.pbjson.dart';

export 'returns.pb.dart';

abstract class ReturnServiceBase extends $pb.GeneratedService {
  $async.Future<$0.CreateReturnRequestResponse> createReturnRequest(
      $pb.ServerContext ctx, $0.CreateReturnRequestRequest request);
  $async.Future<$0.ListReturnRequestsResponse> listReturnRequests(
      $pb.ServerContext ctx, $0.ListReturnRequestsRequest request);
  $async.Future<$0.GetReturnRequestResponse> getReturnRequest(
      $pb.ServerContext ctx, $0.GetReturnRequestRequest request);
  $async.Future<$0.ReviewReturnRequestResponse> reviewReturnRequest(
      $pb.ServerContext ctx, $0.ReviewReturnRequestRequest request);
  $async.Future<$0.GetReturnCertificateResponse> getReturnCertificate(
      $pb.ServerContext ctx, $0.GetReturnCertificateRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateReturnRequest':
        return $0.CreateReturnRequestRequest();
      case 'ListReturnRequests':
        return $0.ListReturnRequestsRequest();
      case 'GetReturnRequest':
        return $0.GetReturnRequestRequest();
      case 'ReviewReturnRequest':
        return $0.ReviewReturnRequestRequest();
      case 'GetReturnCertificate':
        return $0.GetReturnCertificateRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateReturnRequest':
        return createReturnRequest(
            ctx, request as $0.CreateReturnRequestRequest);
      case 'ListReturnRequests':
        return listReturnRequests(ctx, request as $0.ListReturnRequestsRequest);
      case 'GetReturnRequest':
        return getReturnRequest(ctx, request as $0.GetReturnRequestRequest);
      case 'ReviewReturnRequest':
        return reviewReturnRequest(
            ctx, request as $0.ReviewReturnRequestRequest);
      case 'GetReturnCertificate':
        return getReturnCertificate(
            ctx, request as $0.GetReturnCertificateRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ReturnServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ReturnServiceBase$messageJson;
}
