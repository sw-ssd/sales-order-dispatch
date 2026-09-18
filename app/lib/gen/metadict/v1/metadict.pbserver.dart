// This is a generated file - do not edit.
//
// Generated from metadict/v1/metadict.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'metadict.pb.dart' as $1;
import 'metadict.pbjson.dart';

export 'metadict.pb.dart';

abstract class MetadictServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListMetadictsResponse> listMetadicts(
      $pb.ServerContext ctx, $1.ListMetadictsRequest request);
  $async.Future<$1.GetMetadictResponse> getMetadict(
      $pb.ServerContext ctx, $1.GetMetadictRequest request);
  $async.Future<$1.CreateMetadictResponse> createMetadict(
      $pb.ServerContext ctx, $1.CreateMetadictRequest request);
  $async.Future<$1.UpdateMetadictResponse> updateMetadict(
      $pb.ServerContext ctx, $1.UpdateMetadictRequest request);
  $async.Future<$1.DeleteMetadictResponse> deleteMetadict(
      $pb.ServerContext ctx, $1.DeleteMetadictRequest request);
  $async.Future<$1.ListOptionsResponse> listOptions(
      $pb.ServerContext ctx, $1.ListOptionsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListMetadicts':
        return $1.ListMetadictsRequest();
      case 'GetMetadict':
        return $1.GetMetadictRequest();
      case 'CreateMetadict':
        return $1.CreateMetadictRequest();
      case 'UpdateMetadict':
        return $1.UpdateMetadictRequest();
      case 'DeleteMetadict':
        return $1.DeleteMetadictRequest();
      case 'ListOptions':
        return $1.ListOptionsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListMetadicts':
        return listMetadicts(ctx, request as $1.ListMetadictsRequest);
      case 'GetMetadict':
        return getMetadict(ctx, request as $1.GetMetadictRequest);
      case 'CreateMetadict':
        return createMetadict(ctx, request as $1.CreateMetadictRequest);
      case 'UpdateMetadict':
        return updateMetadict(ctx, request as $1.UpdateMetadictRequest);
      case 'DeleteMetadict':
        return deleteMetadict(ctx, request as $1.DeleteMetadictRequest);
      case 'ListOptions':
        return listOptions(ctx, request as $1.ListOptionsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => MetadictServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => MetadictServiceBase$messageJson;
}
