//
//  Generated code. Do not modify.
//  source: masters/v1/master.proto
//

import "package:connectrpc/connect.dart" as connect;
import "master.pb.dart" as mastersv1master;

abstract final class WarehouseService {
  /// Fully-qualified name of the WarehouseService service.
  static const name = 'masters.v1.WarehouseService';

  static const listWarehouses = connect.Spec(
    '/$name/ListWarehouses',
    connect.StreamType.unary,
    mastersv1master.ListWarehousesRequest.new,
    mastersv1master.ListWarehousesResponse.new,
  );

  static const createWarehouse = connect.Spec(
    '/$name/CreateWarehouse',
    connect.StreamType.unary,
    mastersv1master.CreateWarehouseRequest.new,
    mastersv1master.CreateWarehouseResponse.new,
  );

  static const updateWarehouse = connect.Spec(
    '/$name/UpdateWarehouse',
    connect.StreamType.unary,
    mastersv1master.UpdateWarehouseRequest.new,
    mastersv1master.UpdateWarehouseResponse.new,
  );

  static const deleteWarehouse = connect.Spec(
    '/$name/DeleteWarehouse',
    connect.StreamType.unary,
    mastersv1master.DeleteWarehouseRequest.new,
    mastersv1master.DeleteWarehouseResponse.new,
  );

  static const restoreWarehouse = connect.Spec(
    '/$name/RestoreWarehouse',
    connect.StreamType.unary,
    mastersv1master.RestoreWarehouseRequest.new,
    mastersv1master.RestoreWarehouseResponse.new,
  );
}
abstract final class RouteService {
  /// Fully-qualified name of the RouteService service.
  static const name = 'masters.v1.RouteService';

  static const listRoutes = connect.Spec(
    '/$name/ListRoutes',
    connect.StreamType.unary,
    mastersv1master.ListRoutesRequest.new,
    mastersv1master.ListRoutesResponse.new,
  );

  static const createRoute = connect.Spec(
    '/$name/CreateRoute',
    connect.StreamType.unary,
    mastersv1master.CreateRouteRequest.new,
    mastersv1master.CreateRouteResponse.new,
  );

  static const updateRoute = connect.Spec(
    '/$name/UpdateRoute',
    connect.StreamType.unary,
    mastersv1master.UpdateRouteRequest.new,
    mastersv1master.UpdateRouteResponse.new,
  );

  static const deleteRoute = connect.Spec(
    '/$name/DeleteRoute',
    connect.StreamType.unary,
    mastersv1master.DeleteRouteRequest.new,
    mastersv1master.DeleteRouteResponse.new,
  );

  static const restoreRoute = connect.Spec(
    '/$name/RestoreRoute',
    connect.StreamType.unary,
    mastersv1master.RestoreRouteRequest.new,
    mastersv1master.RestoreRouteResponse.new,
  );
}
abstract final class ProcessingSpecService {
  /// Fully-qualified name of the ProcessingSpecService service.
  static const name = 'masters.v1.ProcessingSpecService';

  static const listProcessingSpecs = connect.Spec(
    '/$name/ListProcessingSpecs',
    connect.StreamType.unary,
    mastersv1master.ListProcessingSpecsRequest.new,
    mastersv1master.ListProcessingSpecsResponse.new,
  );

  static const createProcessingSpec = connect.Spec(
    '/$name/CreateProcessingSpec',
    connect.StreamType.unary,
    mastersv1master.CreateProcessingSpecRequest.new,
    mastersv1master.CreateProcessingSpecResponse.new,
  );

  static const updateProcessingSpec = connect.Spec(
    '/$name/UpdateProcessingSpec',
    connect.StreamType.unary,
    mastersv1master.UpdateProcessingSpecRequest.new,
    mastersv1master.UpdateProcessingSpecResponse.new,
  );

  static const deleteProcessingSpec = connect.Spec(
    '/$name/DeleteProcessingSpec',
    connect.StreamType.unary,
    mastersv1master.DeleteProcessingSpecRequest.new,
    mastersv1master.DeleteProcessingSpecResponse.new,
  );

  static const restoreProcessingSpec = connect.Spec(
    '/$name/RestoreProcessingSpec',
    connect.StreamType.unary,
    mastersv1master.RestoreProcessingSpecRequest.new,
    mastersv1master.RestoreProcessingSpecResponse.new,
  );
}
abstract final class ProductCategoryService {
  /// Fully-qualified name of the ProductCategoryService service.
  static const name = 'masters.v1.ProductCategoryService';

  static const listProductCategories = connect.Spec(
    '/$name/ListProductCategories',
    connect.StreamType.unary,
    mastersv1master.ListProductCategoriesRequest.new,
    mastersv1master.ListProductCategoriesResponse.new,
  );

  static const createProductCategory = connect.Spec(
    '/$name/CreateProductCategory',
    connect.StreamType.unary,
    mastersv1master.CreateProductCategoryRequest.new,
    mastersv1master.CreateProductCategoryResponse.new,
  );

  static const updateProductCategory = connect.Spec(
    '/$name/UpdateProductCategory',
    connect.StreamType.unary,
    mastersv1master.UpdateProductCategoryRequest.new,
    mastersv1master.UpdateProductCategoryResponse.new,
  );

  static const deleteProductCategory = connect.Spec(
    '/$name/DeleteProductCategory',
    connect.StreamType.unary,
    mastersv1master.DeleteProductCategoryRequest.new,
    mastersv1master.DeleteProductCategoryResponse.new,
  );

  static const restoreProductCategory = connect.Spec(
    '/$name/RestoreProductCategory',
    connect.StreamType.unary,
    mastersv1master.RestoreProductCategoryRequest.new,
    mastersv1master.RestoreProductCategoryResponse.new,
  );
}
