//
//  Generated code. Do not modify.
//  source: customers/v1/customer.proto
//

import "package:connectrpc/connect.dart" as connect;
import "customer.pb.dart" as customersv1customer;
import "customer.connect.spec.dart" as specs;

/// CustomerService:客戶主檔管理(dept_admin/staff 限所屬部門)。
extension type CustomerServiceClient (connect.Transport _transport) {
  /// ListCustomers:分頁查詢(keyword 模糊比對 name/customer_code/tax_id;可 include_deleted)。
  Future<customersv1customer.ListCustomersResponse> listCustomers(
    customersv1customer.ListCustomersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.listCustomers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCustomer:以 id 取單筆(限可見範圍)。
  Future<customersv1customer.GetCustomerResponse> getCustomer(
    customersv1customer.GetCustomerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.getCustomer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateCustomer:建立客戶(系統取號;字典/業務 reference 驗證)。
  Future<customersv1customer.CreateCustomerResponse> createCustomer(
    customersv1customer.CreateCustomerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.createCustomer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateCustomer:欄位式更新(customer_code 不可改)。
  Future<customersv1customer.UpdateCustomerResponse> updateCustomer(
    customersv1customer.UpdateCustomerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.updateCustomer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteCustomer:軟刪除。
  Future<customersv1customer.DeleteCustomerResponse> deleteCustomer(
    customersv1customer.DeleteCustomerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.deleteCustomer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RestoreCustomer:復原(清 deleted_at)。
  Future<customersv1customer.RestoreCustomerResponse> restoreCustomer(
    customersv1customer.RestoreCustomerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.restoreCustomer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
