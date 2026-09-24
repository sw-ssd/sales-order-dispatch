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

  /// 以下為地址簿與聯絡人(3.2.1 / 3.2.2)。
  Future<customersv1customer.ListAddressesResponse> listAddresses(
    customersv1customer.ListAddressesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.listAddresses,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.AddAddressResponse> addAddress(
    customersv1customer.AddAddressRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.addAddress,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.UpdateAddressResponse> updateAddress(
    customersv1customer.UpdateAddressRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.updateAddress,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.DeleteAddressResponse> deleteAddress(
    customersv1customer.DeleteAddressRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.deleteAddress,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.ListContactsResponse> listContacts(
    customersv1customer.ListContactsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.listContacts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.AddContactResponse> addContact(
    customersv1customer.AddContactRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.addContact,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.UpdateContactResponse> updateContact(
    customersv1customer.UpdateContactRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.updateContact,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.DeleteContactResponse> deleteContact(
    customersv1customer.DeleteContactRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.deleteContact,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCustomerQRCode:為本部門客戶產生登入 QR(dept_admin/staff 限本部門)。
  Future<customersv1customer.GetCustomerQRCodeResponse> getCustomerQRCode(
    customersv1customer.GetCustomerQRCodeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerService.getCustomerQRCode,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// ---- CustomerAccountService:店家自助管理登入帳號(D22/規格 4.2,Task 6.7) ----
/// 僅**客戶主帳號**可呼叫(is_primary):主帳號是該客戶帳號體系的唯一管理者,也是它唯一被允許的
/// 功能面(業務 API 一律 403,見 server.protectedRPC 與 OpenFGA 的 primary_account 排除)。
/// 範圍僅限自己客戶,不得觸及其他客戶或員工帳號。
/// 子帳號無管理權限(本服務對非主帳號一律拒絕);建立客戶時自動附帶的**業務子帳號**
/// (system_generated=true)店家可檢視但不可改名/停用/重置 —— 它專供所屬業務使用,店家並無其密碼。
extension type CustomerAccountServiceClient (connect.Transport _transport) {
  Future<customersv1customer.ListCustomerAccountsResponse> listCustomerAccounts(
    customersv1customer.ListCustomerAccountsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerAccountService.listCustomerAccounts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.CreateCustomerAccountResponse> createCustomerAccount(
    customersv1customer.CreateCustomerAccountRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerAccountService.createCustomerAccount,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.DeactivateCustomerAccountResponse> deactivateCustomerAccount(
    customersv1customer.DeactivateCustomerAccountRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerAccountService.deactivateCustomerAccount,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<customersv1customer.ResetCustomerAccountPasswordResponse> resetCustomerAccountPassword(
    customersv1customer.ResetCustomerAccountPasswordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CustomerAccountService.resetCustomerAccountPassword,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
