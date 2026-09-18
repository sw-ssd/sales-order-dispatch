//
//  Generated code. Do not modify.
//  source: customers/v1/customer.proto
//

import "package:connectrpc/connect.dart" as connect;
import "customer.pb.dart" as customersv1customer;

/// CustomerService:客戶主檔管理(dept_admin/staff 限所屬部門)。
abstract final class CustomerService {
  /// Fully-qualified name of the CustomerService service.
  static const name = 'customers.v1.CustomerService';

  /// ListCustomers:分頁查詢(keyword 模糊比對 name/customer_code/tax_id;可 include_deleted)。
  static const listCustomers = connect.Spec(
    '/$name/ListCustomers',
    connect.StreamType.unary,
    customersv1customer.ListCustomersRequest.new,
    customersv1customer.ListCustomersResponse.new,
  );

  /// GetCustomer:以 id 取單筆(限可見範圍)。
  static const getCustomer = connect.Spec(
    '/$name/GetCustomer',
    connect.StreamType.unary,
    customersv1customer.GetCustomerRequest.new,
    customersv1customer.GetCustomerResponse.new,
  );

  /// CreateCustomer:建立客戶(系統取號;字典/業務 reference 驗證)。
  static const createCustomer = connect.Spec(
    '/$name/CreateCustomer',
    connect.StreamType.unary,
    customersv1customer.CreateCustomerRequest.new,
    customersv1customer.CreateCustomerResponse.new,
  );

  /// UpdateCustomer:欄位式更新(customer_code 不可改)。
  static const updateCustomer = connect.Spec(
    '/$name/UpdateCustomer',
    connect.StreamType.unary,
    customersv1customer.UpdateCustomerRequest.new,
    customersv1customer.UpdateCustomerResponse.new,
  );

  /// DeleteCustomer:軟刪除。
  static const deleteCustomer = connect.Spec(
    '/$name/DeleteCustomer',
    connect.StreamType.unary,
    customersv1customer.DeleteCustomerRequest.new,
    customersv1customer.DeleteCustomerResponse.new,
  );

  /// RestoreCustomer:復原(清 deleted_at)。
  static const restoreCustomer = connect.Spec(
    '/$name/RestoreCustomer',
    connect.StreamType.unary,
    customersv1customer.RestoreCustomerRequest.new,
    customersv1customer.RestoreCustomerResponse.new,
  );
}
