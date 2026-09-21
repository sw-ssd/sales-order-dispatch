//
//  Generated code. Do not modify.
//  source: products/v1/product.proto
//

import "package:connectrpc/connect.dart" as connect;
import "product.pb.dart" as productsv1product;
import "product.connect.spec.dart" as specs;

extension type ProductServiceClient (connect.Transport _transport) {
  Future<productsv1product.ListProductsResponse> listProducts(
    productsv1product.ListProductsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ProductService.listProducts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<productsv1product.GetProductResponse> getProduct(
    productsv1product.GetProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ProductService.getProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<productsv1product.CreateProductResponse> createProduct(
    productsv1product.CreateProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ProductService.createProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<productsv1product.UpdateProductResponse> updateProduct(
    productsv1product.UpdateProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ProductService.updateProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<productsv1product.DeleteProductResponse> deleteProduct(
    productsv1product.DeleteProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ProductService.deleteProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<productsv1product.RestoreProductResponse> restoreProduct(
    productsv1product.RestoreProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ProductService.restoreProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// CustomerProductService:客戶專屬清單(dept_admin/staff 限部門;客戶僅 for_order 語意由 List 旗標表達)。
extension type CustomerProductServiceClient (connect.Transport _transport) {
  /// ListCustomerProducts:查該客戶清單(for_order=true 排除 default_qty=0 與已刪)。
  Future<productsv1product.ListCustomerProductsResponse> listCustomerProducts(
    productsv1product.ListCustomerProductsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerProductService.listCustomerProducts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AddCustomerProduct:新增一筆(一客戶一商品;重複未刪 → already_exists)。
  Future<productsv1product.AddCustomerProductResponse> addCustomerProduct(
    productsv1product.AddCustomerProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerProductService.addCustomerProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateCustomerProduct:改 alias/default_qty/cut_note(不可改 customer/product)。
  Future<productsv1product.UpdateCustomerProductResponse> updateCustomerProduct(
    productsv1product.UpdateCustomerProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerProductService.updateCustomerProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteCustomerProduct:軟刪除 + 稽核。
  Future<productsv1product.DeleteCustomerProductResponse> deleteCustomerProduct(
    productsv1product.DeleteCustomerProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerProductService.deleteCustomerProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// EnsureCustomerProduct:下單手打確認儲存後呼叫(冪等:存在回既有 created=false;唯一衝突吸收)。
  Future<productsv1product.EnsureCustomerProductResponse> ensureCustomerProduct(
    productsv1product.EnsureCustomerProductRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerProductService.ensureCustomerProduct,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
