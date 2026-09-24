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

  /// 以下為地址簿與聯絡人(3.2.1 / 3.2.2)。
  static const listAddresses = connect.Spec(
    '/$name/ListAddresses',
    connect.StreamType.unary,
    customersv1customer.ListAddressesRequest.new,
    customersv1customer.ListAddressesResponse.new,
  );

  static const addAddress = connect.Spec(
    '/$name/AddAddress',
    connect.StreamType.unary,
    customersv1customer.AddAddressRequest.new,
    customersv1customer.AddAddressResponse.new,
  );

  static const updateAddress = connect.Spec(
    '/$name/UpdateAddress',
    connect.StreamType.unary,
    customersv1customer.UpdateAddressRequest.new,
    customersv1customer.UpdateAddressResponse.new,
  );

  static const deleteAddress = connect.Spec(
    '/$name/DeleteAddress',
    connect.StreamType.unary,
    customersv1customer.DeleteAddressRequest.new,
    customersv1customer.DeleteAddressResponse.new,
  );

  static const listContacts = connect.Spec(
    '/$name/ListContacts',
    connect.StreamType.unary,
    customersv1customer.ListContactsRequest.new,
    customersv1customer.ListContactsResponse.new,
  );

  static const addContact = connect.Spec(
    '/$name/AddContact',
    connect.StreamType.unary,
    customersv1customer.AddContactRequest.new,
    customersv1customer.AddContactResponse.new,
  );

  static const updateContact = connect.Spec(
    '/$name/UpdateContact',
    connect.StreamType.unary,
    customersv1customer.UpdateContactRequest.new,
    customersv1customer.UpdateContactResponse.new,
  );

  static const deleteContact = connect.Spec(
    '/$name/DeleteContact',
    connect.StreamType.unary,
    customersv1customer.DeleteContactRequest.new,
    customersv1customer.DeleteContactResponse.new,
  );

  /// GetCustomerQRCode:為本部門客戶產生登入 QR(dept_admin/staff 限本部門)。
  static const getCustomerQRCode = connect.Spec(
    '/$name/GetCustomerQRCode',
    connect.StreamType.unary,
    customersv1customer.GetCustomerQRCodeRequest.new,
    customersv1customer.GetCustomerQRCodeResponse.new,
  );
}
/// ---- CustomerAccountService:店家自助管理登入帳號(D22/規格 4.2,Task 6.7) ----
/// 僅**客戶主帳號**可呼叫(is_primary):主帳號是該客戶帳號體系的唯一管理者,也是它唯一被允許的
/// 功能面(業務 API 一律 403,見 server.protectedRPC 與 OpenFGA 的 primary_account 排除)。
/// 範圍僅限自己客戶,不得觸及其他客戶或員工帳號。
/// 子帳號無管理權限(本服務對非主帳號一律拒絕);建立客戶時自動附帶的**業務子帳號**
/// (system_generated=true)店家可檢視但不可改名/停用/重置 —— 它專供所屬業務使用,店家並無其密碼。
abstract final class CustomerAccountService {
  /// Fully-qualified name of the CustomerAccountService service.
  static const name = 'customers.v1.CustomerAccountService';

  static const listCustomerAccounts = connect.Spec(
    '/$name/ListCustomerAccounts',
    connect.StreamType.unary,
    customersv1customer.ListCustomerAccountsRequest.new,
    customersv1customer.ListCustomerAccountsResponse.new,
  );

  static const createCustomerAccount = connect.Spec(
    '/$name/CreateCustomerAccount',
    connect.StreamType.unary,
    customersv1customer.CreateCustomerAccountRequest.new,
    customersv1customer.CreateCustomerAccountResponse.new,
  );

  static const deactivateCustomerAccount = connect.Spec(
    '/$name/DeactivateCustomerAccount',
    connect.StreamType.unary,
    customersv1customer.DeactivateCustomerAccountRequest.new,
    customersv1customer.DeactivateCustomerAccountResponse.new,
  );

  static const resetCustomerAccountPassword = connect.Spec(
    '/$name/ResetCustomerAccountPassword',
    connect.StreamType.unary,
    customersv1customer.ResetCustomerAccountPasswordRequest.new,
    customersv1customer.ResetCustomerAccountPasswordResponse.new,
  );
}
