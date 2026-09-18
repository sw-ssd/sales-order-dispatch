//
//  Generated code. Do not modify.
//  source: products/v1/product.proto
//

import "package:connectrpc/connect.dart" as connect;
import "product.pb.dart" as productsv1product;

abstract final class ProductService {
  /// Fully-qualified name of the ProductService service.
  static const name = 'products.v1.ProductService';

  static const listProducts = connect.Spec(
    '/$name/ListProducts',
    connect.StreamType.unary,
    productsv1product.ListProductsRequest.new,
    productsv1product.ListProductsResponse.new,
  );

  static const getProduct = connect.Spec(
    '/$name/GetProduct',
    connect.StreamType.unary,
    productsv1product.GetProductRequest.new,
    productsv1product.GetProductResponse.new,
  );

  static const createProduct = connect.Spec(
    '/$name/CreateProduct',
    connect.StreamType.unary,
    productsv1product.CreateProductRequest.new,
    productsv1product.CreateProductResponse.new,
  );

  static const updateProduct = connect.Spec(
    '/$name/UpdateProduct',
    connect.StreamType.unary,
    productsv1product.UpdateProductRequest.new,
    productsv1product.UpdateProductResponse.new,
  );

  static const deleteProduct = connect.Spec(
    '/$name/DeleteProduct',
    connect.StreamType.unary,
    productsv1product.DeleteProductRequest.new,
    productsv1product.DeleteProductResponse.new,
  );

  static const restoreProduct = connect.Spec(
    '/$name/RestoreProduct',
    connect.StreamType.unary,
    productsv1product.RestoreProductRequest.new,
    productsv1product.RestoreProductResponse.new,
  );
}
