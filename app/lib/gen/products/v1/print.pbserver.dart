// This is a generated file - do not edit.
//
// Generated from products/v1/print.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'print.pb.dart' as $0;
import 'print.pbjson.dart';

export 'print.pb.dart';

abstract class PrintServiceBase extends $pb.GeneratedService {
  $async.Future<$0.PreviewResponse> preview(
      $pb.ServerContext ctx, $0.PreviewRequest request);
  $async.Future<$0.PrintResponse> print(
      $pb.ServerContext ctx, $0.PrintRequest request);
  $async.Future<$0.ListLogsResponse> listLogs(
      $pb.ServerContext ctx, $0.ListLogsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'Preview':
        return $0.PreviewRequest();
      case 'Print':
        return $0.PrintRequest();
      case 'ListLogs':
        return $0.ListLogsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'Preview':
        return preview(ctx, request as $0.PreviewRequest);
      case 'Print':
        return print(ctx, request as $0.PrintRequest);
      case 'ListLogs':
        return listLogs(ctx, request as $0.ListLogsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => PrintServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => PrintServiceBase$messageJson;
}
