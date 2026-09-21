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
/// CustomerProductService:客戶專屬清單(dept_admin/staff 限部門;客戶僅 for_order 語意由 List 旗標表達)。
abstract final class CustomerProductService {
  /// Fully-qualified name of the CustomerProductService service.
  static const name = 'products.v1.CustomerProductService';

  /// ListCustomerProducts:查該客戶清單(for_order=true 排除 default_qty=0 與已刪)。
  static const listCustomerProducts = connect.Spec(
    '/$name/ListCustomerProducts',
    connect.StreamType.unary,
    productsv1product.ListCustomerProductsRequest.new,
    productsv1product.ListCustomerProductsResponse.new,
  );

  /// AddCustomerProduct:新增一筆(一客戶一商品;重複未刪 → already_exists)。
  static const addCustomerProduct = connect.Spec(
    '/$name/AddCustomerProduct',
    connect.StreamType.unary,
    productsv1product.AddCustomerProductRequest.new,
    productsv1product.AddCustomerProductResponse.new,
  );

  /// UpdateCustomerProduct:改 alias/default_qty/cut_note(不可改 customer/product)。
  static const updateCustomerProduct = connect.Spec(
    '/$name/UpdateCustomerProduct',
    connect.StreamType.unary,
    productsv1product.UpdateCustomerProductRequest.new,
    productsv1product.UpdateCustomerProductResponse.new,
  );

  /// DeleteCustomerProduct:軟刪除 + 稽核。
  static const deleteCustomerProduct = connect.Spec(
    '/$name/DeleteCustomerProduct',
    connect.StreamType.unary,
    productsv1product.DeleteCustomerProductRequest.new,
    productsv1product.DeleteCustomerProductResponse.new,
  );

  /// EnsureCustomerProduct:下單手打確認儲存後呼叫(冪等:存在回既有 created=false;唯一衝突吸收)。
  static const ensureCustomerProduct = connect.Spec(
    '/$name/EnsureCustomerProduct',
    connect.StreamType.unary,
    productsv1product.EnsureCustomerProductRequest.new,
    productsv1product.EnsureCustomerProductResponse.new,
  );
}
