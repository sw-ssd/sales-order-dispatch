// This is a generated file - do not edit.
//
// Generated from customers/v1/customer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'customer.pb.dart' as $1;
import 'customer.pbjson.dart';

export 'customer.pb.dart';

abstract class CustomerServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListCustomersResponse> listCustomers(
      $pb.ServerContext ctx, $1.ListCustomersRequest request);
  $async.Future<$1.GetCustomerResponse> getCustomer(
      $pb.ServerContext ctx, $1.GetCustomerRequest request);
  $async.Future<$1.CreateCustomerResponse> createCustomer(
      $pb.ServerContext ctx, $1.CreateCustomerRequest request);
  $async.Future<$1.UpdateCustomerResponse> updateCustomer(
      $pb.ServerContext ctx, $1.UpdateCustomerRequest request);
  $async.Future<$1.DeleteCustomerResponse> deleteCustomer(
      $pb.ServerContext ctx, $1.DeleteCustomerRequest request);
  $async.Future<$1.RestoreCustomerResponse> restoreCustomer(
      $pb.ServerContext ctx, $1.RestoreCustomerRequest request);
  $async.Future<$1.ListAddressesResponse> listAddresses(
      $pb.ServerContext ctx, $1.ListAddressesRequest request);
  $async.Future<$1.AddAddressResponse> addAddress(
      $pb.ServerContext ctx, $1.AddAddressRequest request);
  $async.Future<$1.UpdateAddressResponse> updateAddress(
      $pb.ServerContext ctx, $1.UpdateAddressRequest request);
  $async.Future<$1.DeleteAddressResponse> deleteAddress(
      $pb.ServerContext ctx, $1.DeleteAddressRequest request);
  $async.Future<$1.ListContactsResponse> listContacts(
      $pb.ServerContext ctx, $1.ListContactsRequest request);
  $async.Future<$1.AddContactResponse> addContact(
      $pb.ServerContext ctx, $1.AddContactRequest request);
  $async.Future<$1.UpdateContactResponse> updateContact(
      $pb.ServerContext ctx, $1.UpdateContactRequest request);
  $async.Future<$1.DeleteContactResponse> deleteContact(
      $pb.ServerContext ctx, $1.DeleteContactRequest request);
  $async.Future<$1.GetCustomerQRCodeResponse> getCustomerQRCode(
      $pb.ServerContext ctx, $1.GetCustomerQRCodeRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListCustomers':
        return $1.ListCustomersRequest();
      case 'GetCustomer':
        return $1.GetCustomerRequest();
      case 'CreateCustomer':
        return $1.CreateCustomerRequest();
      case 'UpdateCustomer':
        return $1.UpdateCustomerRequest();
      case 'DeleteCustomer':
        return $1.DeleteCustomerRequest();
      case 'RestoreCustomer':
        return $1.RestoreCustomerRequest();
      case 'ListAddresses':
        return $1.ListAddressesRequest();
      case 'AddAddress':
        return $1.AddAddressRequest();
      case 'UpdateAddress':
        return $1.UpdateAddressRequest();
      case 'DeleteAddress':
        return $1.DeleteAddressRequest();
      case 'ListContacts':
        return $1.ListContactsRequest();
      case 'AddContact':
        return $1.AddContactRequest();
      case 'UpdateContact':
        return $1.UpdateContactRequest();
      case 'DeleteContact':
        return $1.DeleteContactRequest();
      case 'GetCustomerQRCode':
        return $1.GetCustomerQRCodeRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListCustomers':
        return listCustomers(ctx, request as $1.ListCustomersRequest);
      case 'GetCustomer':
        return getCustomer(ctx, request as $1.GetCustomerRequest);
      case 'CreateCustomer':
        return createCustomer(ctx, request as $1.CreateCustomerRequest);
      case 'UpdateCustomer':
        return updateCustomer(ctx, request as $1.UpdateCustomerRequest);
      case 'DeleteCustomer':
        return deleteCustomer(ctx, request as $1.DeleteCustomerRequest);
      case 'RestoreCustomer':
        return restoreCustomer(ctx, request as $1.RestoreCustomerRequest);
      case 'ListAddresses':
        return listAddresses(ctx, request as $1.ListAddressesRequest);
      case 'AddAddress':
        return addAddress(ctx, request as $1.AddAddressRequest);
      case 'UpdateAddress':
        return updateAddress(ctx, request as $1.UpdateAddressRequest);
      case 'DeleteAddress':
        return deleteAddress(ctx, request as $1.DeleteAddressRequest);
      case 'ListContacts':
        return listContacts(ctx, request as $1.ListContactsRequest);
      case 'AddContact':
        return addContact(ctx, request as $1.AddContactRequest);
      case 'UpdateContact':
        return updateContact(ctx, request as $1.UpdateContactRequest);
      case 'DeleteContact':
        return deleteContact(ctx, request as $1.DeleteContactRequest);
      case 'GetCustomerQRCode':
        return getCustomerQRCode(ctx, request as $1.GetCustomerQRCodeRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => CustomerServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => CustomerServiceBase$messageJson;
}
