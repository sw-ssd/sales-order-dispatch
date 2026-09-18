// This is a generated file - do not edit.
//
// Generated from masters/v1/master.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'master.pb.dart' as $2;
import 'master.pbjson.dart';

export 'master.pb.dart';

abstract class WarehouseServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListWarehousesResponse> listWarehouses(
      $pb.ServerContext ctx, $2.ListWarehousesRequest request);
  $async.Future<$2.CreateWarehouseResponse> createWarehouse(
      $pb.ServerContext ctx, $2.CreateWarehouseRequest request);
  $async.Future<$2.UpdateWarehouseResponse> updateWarehouse(
      $pb.ServerContext ctx, $2.UpdateWarehouseRequest request);
  $async.Future<$2.DeleteWarehouseResponse> deleteWarehouse(
      $pb.ServerContext ctx, $2.DeleteWarehouseRequest request);
  $async.Future<$2.RestoreWarehouseResponse> restoreWarehouse(
      $pb.ServerContext ctx, $2.RestoreWarehouseRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListWarehouses':
        return $2.ListWarehousesRequest();
      case 'CreateWarehouse':
        return $2.CreateWarehouseRequest();
      case 'UpdateWarehouse':
        return $2.UpdateWarehouseRequest();
      case 'DeleteWarehouse':
        return $2.DeleteWarehouseRequest();
      case 'RestoreWarehouse':
        return $2.RestoreWarehouseRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListWarehouses':
        return listWarehouses(ctx, request as $2.ListWarehousesRequest);
      case 'CreateWarehouse':
        return createWarehouse(ctx, request as $2.CreateWarehouseRequest);
      case 'UpdateWarehouse':
        return updateWarehouse(ctx, request as $2.UpdateWarehouseRequest);
      case 'DeleteWarehouse':
        return deleteWarehouse(ctx, request as $2.DeleteWarehouseRequest);
      case 'RestoreWarehouse':
        return restoreWarehouse(ctx, request as $2.RestoreWarehouseRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => WarehouseServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => WarehouseServiceBase$messageJson;
}

abstract class RouteServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListRoutesResponse> listRoutes(
      $pb.ServerContext ctx, $2.ListRoutesRequest request);
  $async.Future<$2.CreateRouteResponse> createRoute(
      $pb.ServerContext ctx, $2.CreateRouteRequest request);
  $async.Future<$2.UpdateRouteResponse> updateRoute(
      $pb.ServerContext ctx, $2.UpdateRouteRequest request);
  $async.Future<$2.DeleteRouteResponse> deleteRoute(
      $pb.ServerContext ctx, $2.DeleteRouteRequest request);
  $async.Future<$2.RestoreRouteResponse> restoreRoute(
      $pb.ServerContext ctx, $2.RestoreRouteRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListRoutes':
        return $2.ListRoutesRequest();
      case 'CreateRoute':
        return $2.CreateRouteRequest();
      case 'UpdateRoute':
        return $2.UpdateRouteRequest();
      case 'DeleteRoute':
        return $2.DeleteRouteRequest();
      case 'RestoreRoute':
        return $2.RestoreRouteRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListRoutes':
        return listRoutes(ctx, request as $2.ListRoutesRequest);
      case 'CreateRoute':
        return createRoute(ctx, request as $2.CreateRouteRequest);
      case 'UpdateRoute':
        return updateRoute(ctx, request as $2.UpdateRouteRequest);
      case 'DeleteRoute':
        return deleteRoute(ctx, request as $2.DeleteRouteRequest);
      case 'RestoreRoute':
        return restoreRoute(ctx, request as $2.RestoreRouteRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => RouteServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => RouteServiceBase$messageJson;
}

abstract class ProcessingSpecServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListProcessingSpecsResponse> listProcessingSpecs(
      $pb.ServerContext ctx, $2.ListProcessingSpecsRequest request);
  $async.Future<$2.CreateProcessingSpecResponse> createProcessingSpec(
      $pb.ServerContext ctx, $2.CreateProcessingSpecRequest request);
  $async.Future<$2.UpdateProcessingSpecResponse> updateProcessingSpec(
      $pb.ServerContext ctx, $2.UpdateProcessingSpecRequest request);
  $async.Future<$2.DeleteProcessingSpecResponse> deleteProcessingSpec(
      $pb.ServerContext ctx, $2.DeleteProcessingSpecRequest request);
  $async.Future<$2.RestoreProcessingSpecResponse> restoreProcessingSpec(
      $pb.ServerContext ctx, $2.RestoreProcessingSpecRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListProcessingSpecs':
        return $2.ListProcessingSpecsRequest();
      case 'CreateProcessingSpec':
        return $2.CreateProcessingSpecRequest();
      case 'UpdateProcessingSpec':
        return $2.UpdateProcessingSpecRequest();
      case 'DeleteProcessingSpec':
        return $2.DeleteProcessingSpecRequest();
      case 'RestoreProcessingSpec':
        return $2.RestoreProcessingSpecRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListProcessingSpecs':
        return listProcessingSpecs(
            ctx, request as $2.ListProcessingSpecsRequest);
      case 'CreateProcessingSpec':
        return createProcessingSpec(
            ctx, request as $2.CreateProcessingSpecRequest);
      case 'UpdateProcessingSpec':
        return updateProcessingSpec(
            ctx, request as $2.UpdateProcessingSpecRequest);
      case 'DeleteProcessingSpec':
        return deleteProcessingSpec(
            ctx, request as $2.DeleteProcessingSpecRequest);
      case 'RestoreProcessingSpec':
        return restoreProcessingSpec(
            ctx, request as $2.RestoreProcessingSpecRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      ProcessingSpecServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ProcessingSpecServiceBase$messageJson;
}

abstract class ProductCategoryServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListProductCategoriesResponse> listProductCategories(
      $pb.ServerContext ctx, $2.ListProductCategoriesRequest request);
  $async.Future<$2.CreateProductCategoryResponse> createProductCategory(
      $pb.ServerContext ctx, $2.CreateProductCategoryRequest request);
  $async.Future<$2.UpdateProductCategoryResponse> updateProductCategory(
      $pb.ServerContext ctx, $2.UpdateProductCategoryRequest request);
  $async.Future<$2.DeleteProductCategoryResponse> deleteProductCategory(
      $pb.ServerContext ctx, $2.DeleteProductCategoryRequest request);
  $async.Future<$2.RestoreProductCategoryResponse> restoreProductCategory(
      $pb.ServerContext ctx, $2.RestoreProductCategoryRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListProductCategories':
        return $2.ListProductCategoriesRequest();
      case 'CreateProductCategory':
        return $2.CreateProductCategoryRequest();
      case 'UpdateProductCategory':
        return $2.UpdateProductCategoryRequest();
      case 'DeleteProductCategory':
        return $2.DeleteProductCategoryRequest();
      case 'RestoreProductCategory':
        return $2.RestoreProductCategoryRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListProductCategories':
        return listProductCategories(
            ctx, request as $2.ListProductCategoriesRequest);
      case 'CreateProductCategory':
        return createProductCategory(
            ctx, request as $2.CreateProductCategoryRequest);
      case 'UpdateProductCategory':
        return updateProductCategory(
            ctx, request as $2.UpdateProductCategoryRequest);
      case 'DeleteProductCategory':
        return deleteProductCategory(
            ctx, request as $2.DeleteProductCategoryRequest);
      case 'RestoreProductCategory':
        return restoreProductCategory(
            ctx, request as $2.RestoreProductCategoryRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      ProductCategoryServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ProductCategoryServiceBase$messageJson;
}
