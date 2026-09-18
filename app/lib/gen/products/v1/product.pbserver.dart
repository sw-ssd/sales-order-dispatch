// This is a generated file - do not edit.
//
// Generated from products/v1/product.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'product.pb.dart' as $2;
import 'product.pbjson.dart';

export 'product.pb.dart';

abstract class ProductServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListProductsResponse> listProducts(
      $pb.ServerContext ctx, $2.ListProductsRequest request);
  $async.Future<$2.GetProductResponse> getProduct(
      $pb.ServerContext ctx, $2.GetProductRequest request);
  $async.Future<$2.CreateProductResponse> createProduct(
      $pb.ServerContext ctx, $2.CreateProductRequest request);
  $async.Future<$2.UpdateProductResponse> updateProduct(
      $pb.ServerContext ctx, $2.UpdateProductRequest request);
  $async.Future<$2.DeleteProductResponse> deleteProduct(
      $pb.ServerContext ctx, $2.DeleteProductRequest request);
  $async.Future<$2.RestoreProductResponse> restoreProduct(
      $pb.ServerContext ctx, $2.RestoreProductRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListProducts':
        return $2.ListProductsRequest();
      case 'GetProduct':
        return $2.GetProductRequest();
      case 'CreateProduct':
        return $2.CreateProductRequest();
      case 'UpdateProduct':
        return $2.UpdateProductRequest();
      case 'DeleteProduct':
        return $2.DeleteProductRequest();
      case 'RestoreProduct':
        return $2.RestoreProductRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListProducts':
        return listProducts(ctx, request as $2.ListProductsRequest);
      case 'GetProduct':
        return getProduct(ctx, request as $2.GetProductRequest);
      case 'CreateProduct':
        return createProduct(ctx, request as $2.CreateProductRequest);
      case 'UpdateProduct':
        return updateProduct(ctx, request as $2.UpdateProductRequest);
      case 'DeleteProduct':
        return deleteProduct(ctx, request as $2.DeleteProductRequest);
      case 'RestoreProduct':
        return restoreProduct(ctx, request as $2.RestoreProductRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ProductServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ProductServiceBase$messageJson;
}
