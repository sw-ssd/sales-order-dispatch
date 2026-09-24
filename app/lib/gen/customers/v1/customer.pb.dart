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

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../salesorder/v1/common.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Customer:客戶主檔。
class Customer extends $pb.GeneratedMessage {
  factory Customer({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? customerCode,
    $core.String? name,
    $core.String? taxId,
    $core.String? paymentMethodId,
    $core.String? settlementMethodId,
    $core.String? customerTypeId,
    $core.String? invoiceTypeId,
    $core.String? defaultSalesRepId,
    $core.Iterable<$core.bool>? preferredDeliveryDays,
    $core.Iterable<$fixnum.Int64>? promoTagIds,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = Customer._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (customerCode != null) result.customerCode = customerCode;
    if (name != null) result.name = name;
    if (taxId != null) result.taxId = taxId;
    if (paymentMethodId != null) result.paymentMethodId = paymentMethodId;
    if (settlementMethodId != null)
      result.settlementMethodId = settlementMethodId;
    if (customerTypeId != null) result.customerTypeId = customerTypeId;
    if (invoiceTypeId != null) result.invoiceTypeId = invoiceTypeId;
    if (defaultSalesRepId != null) result.defaultSalesRepId = defaultSalesRepId;
    if (preferredDeliveryDays != null)
      result.preferredDeliveryDays.addAll(preferredDeliveryDays);
    if (promoTagIds != null) result.promoTagIds.addAll(promoTagIds);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  Customer._();

  factory Customer.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Customer()..mergeFromBuffer(data, registry);
  factory Customer.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Customer()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Customer',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: Customer.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'customerCode')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'taxId')
    ..aOS(7, _omitFieldNames ? '' : 'paymentMethodId')
    ..aOS(8, _omitFieldNames ? '' : 'settlementMethodId')
    ..aOS(9, _omitFieldNames ? '' : 'customerTypeId')
    ..aOS(10, _omitFieldNames ? '' : 'invoiceTypeId')
    ..aOS(11, _omitFieldNames ? '' : 'defaultSalesRepId')
    ..p<$core.bool>(
        12, _omitFieldNames ? '' : 'preferredDeliveryDays', $pb.PbFieldType.KB)
    ..p<$fixnum.Int64>(
        13, _omitFieldNames ? '' : 'promoTagIds', $pb.PbFieldType.K6)
    ..aOS(14, _omitFieldNames ? '' : 'createdAt')
    ..aOS(15, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(16, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Customer clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Customer copyWith(void Function(Customer) updates) =>
      super.copyWith((message) => updates(message as Customer)) as Customer;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Customer() / Customer.new instead')
  static Customer create() => Customer._();
  static $pb.GeneratedMessage $_createMessage() => Customer._();
  @$core.override
  Customer createEmptyInstance() => Customer._();
  @$core.pragma('dart2js:noInline')
  static Customer getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Customer>(Customer.$_createMessage);
  static Customer? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get companyId => $_getSZ(1);
  @$pb.TagNumber(2)
  set companyId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCompanyId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompanyId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get departmentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set departmentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDepartmentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearDepartmentId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get customerCode => $_getSZ(3);
  @$pb.TagNumber(4)
  set customerCode($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCustomerCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearCustomerCode() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get name => $_getSZ(4);
  @$pb.TagNumber(5)
  set name($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasName() => $_has(4);
  @$pb.TagNumber(5)
  void clearName() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get taxId => $_getSZ(5);
  @$pb.TagNumber(6)
  set taxId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTaxId() => $_has(5);
  @$pb.TagNumber(6)
  void clearTaxId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get paymentMethodId => $_getSZ(6);
  @$pb.TagNumber(7)
  set paymentMethodId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPaymentMethodId() => $_has(6);
  @$pb.TagNumber(7)
  void clearPaymentMethodId() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get settlementMethodId => $_getSZ(7);
  @$pb.TagNumber(8)
  set settlementMethodId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasSettlementMethodId() => $_has(7);
  @$pb.TagNumber(8)
  void clearSettlementMethodId() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get customerTypeId => $_getSZ(8);
  @$pb.TagNumber(9)
  set customerTypeId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCustomerTypeId() => $_has(8);
  @$pb.TagNumber(9)
  void clearCustomerTypeId() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get invoiceTypeId => $_getSZ(9);
  @$pb.TagNumber(10)
  set invoiceTypeId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasInvoiceTypeId() => $_has(9);
  @$pb.TagNumber(10)
  void clearInvoiceTypeId() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get defaultSalesRepId => $_getSZ(10);
  @$pb.TagNumber(11)
  set defaultSalesRepId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasDefaultSalesRepId() => $_has(10);
  @$pb.TagNumber(11)
  void clearDefaultSalesRepId() => $_clearField(11);

  @$pb.TagNumber(12)
  $pb.PbList<$core.bool> get preferredDeliveryDays => $_getList(11);

  @$pb.TagNumber(13)
  $pb.PbList<$fixnum.Int64> get promoTagIds => $_getList(12);

  @$pb.TagNumber(14)
  $core.String get createdAt => $_getSZ(13);
  @$pb.TagNumber(14)
  set createdAt($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasCreatedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearCreatedAt() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.String get updatedAt => $_getSZ(14);
  @$pb.TagNumber(15)
  set updatedAt($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasUpdatedAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearUpdatedAt() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.String get deletedAt => $_getSZ(15);
  @$pb.TagNumber(16)
  set deletedAt($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasDeletedAt() => $_has(15);
  @$pb.TagNumber(16)
  void clearDeletedAt() => $_clearField(16);
}

class ListCustomersRequest extends $pb.GeneratedMessage {
  factory ListCustomersRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? keyword,
    $core.bool? includeDeleted,
    $core.String? sort,
    $core.bool? desc,
  }) {
    final result = ListCustomersRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    if (sort != null) result.sort = sort;
    if (desc != null) result.desc = desc;
    return result;
  }

  ListCustomersRequest._();

  factory ListCustomersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomersRequest()..mergeFromBuffer(data, registry);
  factory ListCustomersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListCustomersRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..aOS(5, _omitFieldNames ? '' : 'sort')
    ..aOB(6, _omitFieldNames ? '' : 'desc')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersRequest copyWith(void Function(ListCustomersRequest) updates) =>
      super.copyWith((message) => updates(message as ListCustomersRequest))
          as ListCustomersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCustomersRequest() / ListCustomersRequest.new instead')
  static ListCustomersRequest create() => ListCustomersRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListCustomersRequest._();
  @$core.override
  ListCustomersRequest createEmptyInstance() => ListCustomersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListCustomersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomersRequest>(
          ListCustomersRequest.$_createMessage);
  static ListCustomersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get page => $_getIZ(0);
  @$pb.TagNumber(1)
  set page($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPage() => $_has(0);
  @$pb.TagNumber(1)
  void clearPage() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get pageSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set pageSize($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageSize() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get keyword => $_getSZ(2);
  @$pb.TagNumber(3)
  set keyword($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasKeyword() => $_has(2);
  @$pb.TagNumber(3)
  void clearKeyword() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get includeDeleted => $_getBF(3);
  @$pb.TagNumber(4)
  set includeDeleted($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIncludeDeleted() => $_has(3);
  @$pb.TagNumber(4)
  void clearIncludeDeleted() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get sort => $_getSZ(4);
  @$pb.TagNumber(5)
  set sort($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSort() => $_has(4);
  @$pb.TagNumber(5)
  void clearSort() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get desc => $_getBF(5);
  @$pb.TagNumber(6)
  set desc($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasDesc() => $_has(5);
  @$pb.TagNumber(6)
  void clearDesc() => $_clearField(6);
}

class ListCustomersResponse extends $pb.GeneratedMessage {
  factory ListCustomersResponse({
    $core.Iterable<Customer>? customers,
    $0.Pagination? pagination,
  }) {
    final result = ListCustomersResponse._();
    if (customers != null) result.customers.addAll(customers);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListCustomersResponse._();

  factory ListCustomersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomersResponse()..mergeFromBuffer(data, registry);
  factory ListCustomersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListCustomersResponse.$_createMessage)
    ..pPM<Customer>(1, _omitFieldNames ? '' : 'customers',
        subBuilder: Customer.$_createMessage)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersResponse copyWith(
          void Function(ListCustomersResponse) updates) =>
      super.copyWith((message) => updates(message as ListCustomersResponse))
          as ListCustomersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCustomersResponse() / ListCustomersResponse.new instead')
  static ListCustomersResponse create() => ListCustomersResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListCustomersResponse._();
  @$core.override
  ListCustomersResponse createEmptyInstance() => ListCustomersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListCustomersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomersResponse>(
          ListCustomersResponse.$_createMessage);
  static ListCustomersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Customer> get customers => $_getList(0);

  @$pb.TagNumber(2)
  $0.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($0.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Pagination ensurePagination() => $_ensure(1);
}

class GetCustomerRequest extends $pb.GeneratedMessage {
  factory GetCustomerRequest({
    $core.String? id,
  }) {
    final result = GetCustomerRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetCustomerRequest._();

  factory GetCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerRequest()..mergeFromBuffer(data, registry);
  factory GetCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: GetCustomerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerRequest copyWith(void Function(GetCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as GetCustomerRequest))
          as GetCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetCustomerRequest() / GetCustomerRequest.new instead')
  static GetCustomerRequest create() => GetCustomerRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetCustomerRequest._();
  @$core.override
  GetCustomerRequest createEmptyInstance() => GetCustomerRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCustomerRequest>(
          GetCustomerRequest.$_createMessage);
  static GetCustomerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetCustomerResponse extends $pb.GeneratedMessage {
  factory GetCustomerResponse({
    Customer? customer,
  }) {
    final result = GetCustomerResponse._();
    if (customer != null) result.customer = customer;
    return result;
  }

  GetCustomerResponse._();

  factory GetCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerResponse()..mergeFromBuffer(data, registry);
  factory GetCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: GetCustomerResponse.$_createMessage)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerResponse copyWith(void Function(GetCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as GetCustomerResponse))
          as GetCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetCustomerResponse() / GetCustomerResponse.new instead')
  static GetCustomerResponse create() => GetCustomerResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetCustomerResponse._();
  @$core.override
  GetCustomerResponse createEmptyInstance() => GetCustomerResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCustomerResponse>(
          GetCustomerResponse.$_createMessage);
  static GetCustomerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Customer get customer => $_getN(0);
  @$pb.TagNumber(1)
  set customer(Customer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomer() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomer() => $_clearField(1);
  @$pb.TagNumber(1)
  Customer ensureCustomer() => $_ensure(0);
}

class CreateCustomerRequest extends $pb.GeneratedMessage {
  factory CreateCustomerRequest({
    $core.String? name,
    $core.String? taxId,
    $core.String? paymentMethodId,
    $core.String? settlementMethodId,
    $core.String? customerTypeId,
    $core.String? invoiceTypeId,
    $core.String? defaultSalesRepId,
    $core.Iterable<$core.bool>? preferredDeliveryDays,
    $core.Iterable<$fixnum.Int64>? promoTagIds,
  }) {
    final result = CreateCustomerRequest._();
    if (name != null) result.name = name;
    if (taxId != null) result.taxId = taxId;
    if (paymentMethodId != null) result.paymentMethodId = paymentMethodId;
    if (settlementMethodId != null)
      result.settlementMethodId = settlementMethodId;
    if (customerTypeId != null) result.customerTypeId = customerTypeId;
    if (invoiceTypeId != null) result.invoiceTypeId = invoiceTypeId;
    if (defaultSalesRepId != null) result.defaultSalesRepId = defaultSalesRepId;
    if (preferredDeliveryDays != null)
      result.preferredDeliveryDays.addAll(preferredDeliveryDays);
    if (promoTagIds != null) result.promoTagIds.addAll(promoTagIds);
    return result;
  }

  CreateCustomerRequest._();

  factory CreateCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerRequest()..mergeFromBuffer(data, registry);
  factory CreateCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CreateCustomerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'taxId')
    ..aOS(3, _omitFieldNames ? '' : 'paymentMethodId')
    ..aOS(4, _omitFieldNames ? '' : 'settlementMethodId')
    ..aOS(5, _omitFieldNames ? '' : 'customerTypeId')
    ..aOS(6, _omitFieldNames ? '' : 'invoiceTypeId')
    ..aOS(7, _omitFieldNames ? '' : 'defaultSalesRepId')
    ..p<$core.bool>(
        8, _omitFieldNames ? '' : 'preferredDeliveryDays', $pb.PbFieldType.KB)
    ..p<$fixnum.Int64>(
        9, _omitFieldNames ? '' : 'promoTagIds', $pb.PbFieldType.K6)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerRequest copyWith(
          void Function(CreateCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as CreateCustomerRequest))
          as CreateCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCustomerRequest() / CreateCustomerRequest.new instead')
  static CreateCustomerRequest create() => CreateCustomerRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateCustomerRequest._();
  @$core.override
  CreateCustomerRequest createEmptyInstance() => CreateCustomerRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCustomerRequest>(
          CreateCustomerRequest.$_createMessage);
  static CreateCustomerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get taxId => $_getSZ(1);
  @$pb.TagNumber(2)
  set taxId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTaxId() => $_has(1);
  @$pb.TagNumber(2)
  void clearTaxId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get paymentMethodId => $_getSZ(2);
  @$pb.TagNumber(3)
  set paymentMethodId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPaymentMethodId() => $_has(2);
  @$pb.TagNumber(3)
  void clearPaymentMethodId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get settlementMethodId => $_getSZ(3);
  @$pb.TagNumber(4)
  set settlementMethodId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSettlementMethodId() => $_has(3);
  @$pb.TagNumber(4)
  void clearSettlementMethodId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get customerTypeId => $_getSZ(4);
  @$pb.TagNumber(5)
  set customerTypeId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCustomerTypeId() => $_has(4);
  @$pb.TagNumber(5)
  void clearCustomerTypeId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get invoiceTypeId => $_getSZ(5);
  @$pb.TagNumber(6)
  set invoiceTypeId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasInvoiceTypeId() => $_has(5);
  @$pb.TagNumber(6)
  void clearInvoiceTypeId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get defaultSalesRepId => $_getSZ(6);
  @$pb.TagNumber(7)
  set defaultSalesRepId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasDefaultSalesRepId() => $_has(6);
  @$pb.TagNumber(7)
  void clearDefaultSalesRepId() => $_clearField(7);

  @$pb.TagNumber(8)
  $pb.PbList<$core.bool> get preferredDeliveryDays => $_getList(7);

  @$pb.TagNumber(9)
  $pb.PbList<$fixnum.Int64> get promoTagIds => $_getList(8);
}

class CreateCustomerResponse extends $pb.GeneratedMessage {
  factory CreateCustomerResponse({
    Customer? customer,
    $core.String? primaryAccountName,
    $core.String? primaryTempPassword,
    $core.String? salesRepAccountName,
    $core.String? salesRepTempPassword,
    $core.String? accountManageUrl,
  }) {
    final result = CreateCustomerResponse._();
    if (customer != null) result.customer = customer;
    if (primaryAccountName != null)
      result.primaryAccountName = primaryAccountName;
    if (primaryTempPassword != null)
      result.primaryTempPassword = primaryTempPassword;
    if (salesRepAccountName != null)
      result.salesRepAccountName = salesRepAccountName;
    if (salesRepTempPassword != null)
      result.salesRepTempPassword = salesRepTempPassword;
    if (accountManageUrl != null) result.accountManageUrl = accountManageUrl;
    return result;
  }

  CreateCustomerResponse._();

  factory CreateCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerResponse()..mergeFromBuffer(data, registry);
  factory CreateCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CreateCustomerResponse.$_createMessage)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'primaryAccountName')
    ..aOS(3, _omitFieldNames ? '' : 'primaryTempPassword')
    ..aOS(4, _omitFieldNames ? '' : 'salesRepAccountName')
    ..aOS(5, _omitFieldNames ? '' : 'salesRepTempPassword')
    ..aOS(6, _omitFieldNames ? '' : 'accountManageUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerResponse copyWith(
          void Function(CreateCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as CreateCustomerResponse))
          as CreateCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCustomerResponse() / CreateCustomerResponse.new instead')
  static CreateCustomerResponse create() => CreateCustomerResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateCustomerResponse._();
  @$core.override
  CreateCustomerResponse createEmptyInstance() => CreateCustomerResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCustomerResponse>(
          CreateCustomerResponse.$_createMessage);
  static CreateCustomerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Customer get customer => $_getN(0);
  @$pb.TagNumber(1)
  set customer(Customer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomer() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomer() => $_clearField(1);
  @$pb.TagNumber(1)
  Customer ensureCustomer() => $_ensure(0);

  /// D22 建檔連動帳號交付(規格 §4.2/§9.4):臨時密碼僅本次回應出現,系統不留明文。
  @$pb.TagNumber(2)
  $core.String get primaryAccountName => $_getSZ(1);
  @$pb.TagNumber(2)
  set primaryAccountName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPrimaryAccountName() => $_has(1);
  @$pb.TagNumber(2)
  void clearPrimaryAccountName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get primaryTempPassword => $_getSZ(2);
  @$pb.TagNumber(3)
  set primaryTempPassword($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPrimaryTempPassword() => $_has(2);
  @$pb.TagNumber(3)
  void clearPrimaryTempPassword() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get salesRepAccountName => $_getSZ(3);
  @$pb.TagNumber(4)
  set salesRepAccountName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSalesRepAccountName() => $_has(3);
  @$pb.TagNumber(4)
  void clearSalesRepAccountName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get salesRepTempPassword => $_getSZ(4);
  @$pb.TagNumber(5)
  set salesRepTempPassword($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSalesRepTempPassword() => $_has(4);
  @$pb.TagNumber(5)
  void clearSalesRepTempPassword() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get accountManageUrl => $_getSZ(5);
  @$pb.TagNumber(6)
  set accountManageUrl($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAccountManageUrl() => $_has(5);
  @$pb.TagNumber(6)
  void clearAccountManageUrl() => $_clearField(6);
}

class UpdateCustomerRequest extends $pb.GeneratedMessage {
  factory UpdateCustomerRequest({
    $core.String? id,
    $core.String? name,
    $core.String? taxId,
    $core.String? paymentMethodId,
    $core.String? settlementMethodId,
    $core.String? customerTypeId,
    $core.String? invoiceTypeId,
    $core.String? defaultSalesRepId,
    $core.Iterable<$core.bool>? preferredDeliveryDays,
    $core.Iterable<$fixnum.Int64>? promoTagIds,
  }) {
    final result = UpdateCustomerRequest._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (taxId != null) result.taxId = taxId;
    if (paymentMethodId != null) result.paymentMethodId = paymentMethodId;
    if (settlementMethodId != null)
      result.settlementMethodId = settlementMethodId;
    if (customerTypeId != null) result.customerTypeId = customerTypeId;
    if (invoiceTypeId != null) result.invoiceTypeId = invoiceTypeId;
    if (defaultSalesRepId != null) result.defaultSalesRepId = defaultSalesRepId;
    if (preferredDeliveryDays != null)
      result.preferredDeliveryDays.addAll(preferredDeliveryDays);
    if (promoTagIds != null) result.promoTagIds.addAll(promoTagIds);
    return result;
  }

  UpdateCustomerRequest._();

  factory UpdateCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerRequest()..mergeFromBuffer(data, registry);
  factory UpdateCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: UpdateCustomerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'taxId')
    ..aOS(4, _omitFieldNames ? '' : 'paymentMethodId')
    ..aOS(5, _omitFieldNames ? '' : 'settlementMethodId')
    ..aOS(6, _omitFieldNames ? '' : 'customerTypeId')
    ..aOS(7, _omitFieldNames ? '' : 'invoiceTypeId')
    ..aOS(8, _omitFieldNames ? '' : 'defaultSalesRepId')
    ..p<$core.bool>(
        9, _omitFieldNames ? '' : 'preferredDeliveryDays', $pb.PbFieldType.KB)
    ..p<$fixnum.Int64>(
        10, _omitFieldNames ? '' : 'promoTagIds', $pb.PbFieldType.K6)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerRequest copyWith(
          void Function(UpdateCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateCustomerRequest))
          as UpdateCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCustomerRequest() / UpdateCustomerRequest.new instead')
  static UpdateCustomerRequest create() => UpdateCustomerRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateCustomerRequest._();
  @$core.override
  UpdateCustomerRequest createEmptyInstance() => UpdateCustomerRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCustomerRequest>(
          UpdateCustomerRequest.$_createMessage);
  static UpdateCustomerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get taxId => $_getSZ(2);
  @$pb.TagNumber(3)
  set taxId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTaxId() => $_has(2);
  @$pb.TagNumber(3)
  void clearTaxId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get paymentMethodId => $_getSZ(3);
  @$pb.TagNumber(4)
  set paymentMethodId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPaymentMethodId() => $_has(3);
  @$pb.TagNumber(4)
  void clearPaymentMethodId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get settlementMethodId => $_getSZ(4);
  @$pb.TagNumber(5)
  set settlementMethodId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSettlementMethodId() => $_has(4);
  @$pb.TagNumber(5)
  void clearSettlementMethodId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get customerTypeId => $_getSZ(5);
  @$pb.TagNumber(6)
  set customerTypeId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCustomerTypeId() => $_has(5);
  @$pb.TagNumber(6)
  void clearCustomerTypeId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get invoiceTypeId => $_getSZ(6);
  @$pb.TagNumber(7)
  set invoiceTypeId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInvoiceTypeId() => $_has(6);
  @$pb.TagNumber(7)
  void clearInvoiceTypeId() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get defaultSalesRepId => $_getSZ(7);
  @$pb.TagNumber(8)
  set defaultSalesRepId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasDefaultSalesRepId() => $_has(7);
  @$pb.TagNumber(8)
  void clearDefaultSalesRepId() => $_clearField(8);

  @$pb.TagNumber(9)
  $pb.PbList<$core.bool> get preferredDeliveryDays => $_getList(8);

  @$pb.TagNumber(10)
  $pb.PbList<$fixnum.Int64> get promoTagIds => $_getList(9);
}

class UpdateCustomerResponse extends $pb.GeneratedMessage {
  factory UpdateCustomerResponse({
    Customer? customer,
  }) {
    final result = UpdateCustomerResponse._();
    if (customer != null) result.customer = customer;
    return result;
  }

  UpdateCustomerResponse._();

  factory UpdateCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerResponse()..mergeFromBuffer(data, registry);
  factory UpdateCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCustomerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: UpdateCustomerResponse.$_createMessage)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerResponse copyWith(
          void Function(UpdateCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateCustomerResponse))
          as UpdateCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCustomerResponse() / UpdateCustomerResponse.new instead')
  static UpdateCustomerResponse create() => UpdateCustomerResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateCustomerResponse._();
  @$core.override
  UpdateCustomerResponse createEmptyInstance() => UpdateCustomerResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCustomerResponse>(
          UpdateCustomerResponse.$_createMessage);
  static UpdateCustomerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Customer get customer => $_getN(0);
  @$pb.TagNumber(1)
  set customer(Customer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomer() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomer() => $_clearField(1);
  @$pb.TagNumber(1)
  Customer ensureCustomer() => $_ensure(0);
}

class DeleteCustomerRequest extends $pb.GeneratedMessage {
  factory DeleteCustomerRequest({
    $core.String? id,
  }) {
    final result = DeleteCustomerRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteCustomerRequest._();

  factory DeleteCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerRequest()..mergeFromBuffer(data, registry);
  factory DeleteCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeleteCustomerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerRequest copyWith(
          void Function(DeleteCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteCustomerRequest))
          as DeleteCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCustomerRequest() / DeleteCustomerRequest.new instead')
  static DeleteCustomerRequest create() => DeleteCustomerRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteCustomerRequest._();
  @$core.override
  DeleteCustomerRequest createEmptyInstance() => DeleteCustomerRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCustomerRequest>(
          DeleteCustomerRequest.$_createMessage);
  static DeleteCustomerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteCustomerResponse extends $pb.GeneratedMessage {
  factory DeleteCustomerResponse() => DeleteCustomerResponse._();

  DeleteCustomerResponse._();

  factory DeleteCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerResponse()..mergeFromBuffer(data, registry);
  factory DeleteCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCustomerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeleteCustomerResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerResponse copyWith(
          void Function(DeleteCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteCustomerResponse))
          as DeleteCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCustomerResponse() / DeleteCustomerResponse.new instead')
  static DeleteCustomerResponse create() => DeleteCustomerResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteCustomerResponse._();
  @$core.override
  DeleteCustomerResponse createEmptyInstance() => DeleteCustomerResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCustomerResponse>(
          DeleteCustomerResponse.$_createMessage);
  static DeleteCustomerResponse? _defaultInstance;
}

class RestoreCustomerRequest extends $pb.GeneratedMessage {
  factory RestoreCustomerRequest({
    $core.String? id,
  }) {
    final result = RestoreCustomerRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RestoreCustomerRequest._();

  factory RestoreCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreCustomerRequest()..mergeFromBuffer(data, registry);
  factory RestoreCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreCustomerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: RestoreCustomerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerRequest copyWith(
          void Function(RestoreCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as RestoreCustomerRequest))
          as RestoreCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreCustomerRequest() / RestoreCustomerRequest.new instead')
  static RestoreCustomerRequest create() => RestoreCustomerRequest._();
  static $pb.GeneratedMessage $_createMessage() => RestoreCustomerRequest._();
  @$core.override
  RestoreCustomerRequest createEmptyInstance() => RestoreCustomerRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreCustomerRequest>(
          RestoreCustomerRequest.$_createMessage);
  static RestoreCustomerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RestoreCustomerResponse extends $pb.GeneratedMessage {
  factory RestoreCustomerResponse({
    Customer? customer,
  }) {
    final result = RestoreCustomerResponse._();
    if (customer != null) result.customer = customer;
    return result;
  }

  RestoreCustomerResponse._();

  factory RestoreCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreCustomerResponse()..mergeFromBuffer(data, registry);
  factory RestoreCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreCustomerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: RestoreCustomerResponse.$_createMessage)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerResponse copyWith(
          void Function(RestoreCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as RestoreCustomerResponse))
          as RestoreCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreCustomerResponse() / RestoreCustomerResponse.new instead')
  static RestoreCustomerResponse create() => RestoreCustomerResponse._();
  static $pb.GeneratedMessage $_createMessage() => RestoreCustomerResponse._();
  @$core.override
  RestoreCustomerResponse createEmptyInstance() => RestoreCustomerResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreCustomerResponse>(
          RestoreCustomerResponse.$_createMessage);
  static RestoreCustomerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Customer get customer => $_getN(0);
  @$pb.TagNumber(1)
  set customer(Customer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomer() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomer() => $_clearField(1);
  @$pb.TagNumber(1)
  Customer ensureCustomer() => $_ensure(0);
}

/// CustomerAddress:客戶地址(3.2.1)。
class CustomerAddress extends $pb.GeneratedMessage {
  factory CustomerAddress({
    $core.String? id,
    $core.String? customerId,
    $core.String? type,
    $core.String? recipientName,
    $core.String? phone,
    $core.String? addressLine,
    $core.String? city,
    $core.String? postalCode,
    $core.bool? isDefault,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = CustomerAddress._();
    if (id != null) result.id = id;
    if (customerId != null) result.customerId = customerId;
    if (type != null) result.type = type;
    if (recipientName != null) result.recipientName = recipientName;
    if (phone != null) result.phone = phone;
    if (addressLine != null) result.addressLine = addressLine;
    if (city != null) result.city = city;
    if (postalCode != null) result.postalCode = postalCode;
    if (isDefault != null) result.isDefault = isDefault;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  CustomerAddress._();

  factory CustomerAddress.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerAddress()..mergeFromBuffer(data, registry);
  factory CustomerAddress.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerAddress()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CustomerAddress',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CustomerAddress.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'customerId')
    ..aOS(3, _omitFieldNames ? '' : 'type')
    ..aOS(4, _omitFieldNames ? '' : 'recipientName')
    ..aOS(5, _omitFieldNames ? '' : 'phone')
    ..aOS(6, _omitFieldNames ? '' : 'addressLine')
    ..aOS(7, _omitFieldNames ? '' : 'city')
    ..aOS(8, _omitFieldNames ? '' : 'postalCode')
    ..aOB(9, _omitFieldNames ? '' : 'isDefault')
    ..aOS(10, _omitFieldNames ? '' : 'createdAt')
    ..aOS(11, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(12, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerAddress clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerAddress copyWith(void Function(CustomerAddress) updates) =>
      super.copyWith((message) => updates(message as CustomerAddress))
          as CustomerAddress;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CustomerAddress() / CustomerAddress.new instead')
  static CustomerAddress create() => CustomerAddress._();
  static $pb.GeneratedMessage $_createMessage() => CustomerAddress._();
  @$core.override
  CustomerAddress createEmptyInstance() => CustomerAddress._();
  @$core.pragma('dart2js:noInline')
  static CustomerAddress getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CustomerAddress>(
          CustomerAddress.$_createMessage);
  static CustomerAddress? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get customerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set customerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCustomerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCustomerId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get type => $_getSZ(2);
  @$pb.TagNumber(3)
  set type($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get recipientName => $_getSZ(3);
  @$pb.TagNumber(4)
  set recipientName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRecipientName() => $_has(3);
  @$pb.TagNumber(4)
  void clearRecipientName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get phone => $_getSZ(4);
  @$pb.TagNumber(5)
  set phone($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPhone() => $_has(4);
  @$pb.TagNumber(5)
  void clearPhone() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get addressLine => $_getSZ(5);
  @$pb.TagNumber(6)
  set addressLine($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAddressLine() => $_has(5);
  @$pb.TagNumber(6)
  void clearAddressLine() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get city => $_getSZ(6);
  @$pb.TagNumber(7)
  set city($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCity() => $_has(6);
  @$pb.TagNumber(7)
  void clearCity() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get postalCode => $_getSZ(7);
  @$pb.TagNumber(8)
  set postalCode($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasPostalCode() => $_has(7);
  @$pb.TagNumber(8)
  void clearPostalCode() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.bool get isDefault => $_getBF(8);
  @$pb.TagNumber(9)
  set isDefault($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasIsDefault() => $_has(8);
  @$pb.TagNumber(9)
  void clearIsDefault() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get createdAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set createdAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasCreatedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearCreatedAt() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get updatedAt => $_getSZ(10);
  @$pb.TagNumber(11)
  set updatedAt($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasUpdatedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearUpdatedAt() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.String get deletedAt => $_getSZ(11);
  @$pb.TagNumber(12)
  set deletedAt($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasDeletedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearDeletedAt() => $_clearField(12);
}

class ListAddressesRequest extends $pb.GeneratedMessage {
  factory ListAddressesRequest({
    $core.String? customerId,
    $core.bool? includeDeleted,
  }) {
    final result = ListAddressesRequest._();
    if (customerId != null) result.customerId = customerId;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListAddressesRequest._();

  factory ListAddressesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAddressesRequest()..mergeFromBuffer(data, registry);
  factory ListAddressesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAddressesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAddressesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListAddressesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOB(2, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAddressesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAddressesRequest copyWith(void Function(ListAddressesRequest) updates) =>
      super.copyWith((message) => updates(message as ListAddressesRequest))
          as ListAddressesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAddressesRequest() / ListAddressesRequest.new instead')
  static ListAddressesRequest create() => ListAddressesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListAddressesRequest._();
  @$core.override
  ListAddressesRequest createEmptyInstance() => ListAddressesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAddressesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAddressesRequest>(
          ListAddressesRequest.$_createMessage);
  static ListAddressesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get includeDeleted => $_getBF(1);
  @$pb.TagNumber(2)
  set includeDeleted($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasIncludeDeleted() => $_has(1);
  @$pb.TagNumber(2)
  void clearIncludeDeleted() => $_clearField(2);
}

class ListAddressesResponse extends $pb.GeneratedMessage {
  factory ListAddressesResponse({
    $core.Iterable<CustomerAddress>? addresses,
  }) {
    final result = ListAddressesResponse._();
    if (addresses != null) result.addresses.addAll(addresses);
    return result;
  }

  ListAddressesResponse._();

  factory ListAddressesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAddressesResponse()..mergeFromBuffer(data, registry);
  factory ListAddressesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAddressesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAddressesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListAddressesResponse.$_createMessage)
    ..pPM<CustomerAddress>(1, _omitFieldNames ? '' : 'addresses',
        subBuilder: CustomerAddress.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAddressesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAddressesResponse copyWith(
          void Function(ListAddressesResponse) updates) =>
      super.copyWith((message) => updates(message as ListAddressesResponse))
          as ListAddressesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAddressesResponse() / ListAddressesResponse.new instead')
  static ListAddressesResponse create() => ListAddressesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListAddressesResponse._();
  @$core.override
  ListAddressesResponse createEmptyInstance() => ListAddressesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAddressesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAddressesResponse>(
          ListAddressesResponse.$_createMessage);
  static ListAddressesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CustomerAddress> get addresses => $_getList(0);
}

class AddAddressRequest extends $pb.GeneratedMessage {
  factory AddAddressRequest({
    $core.String? customerId,
    $core.String? type,
    $core.String? recipientName,
    $core.String? phone,
    $core.String? addressLine,
    $core.String? city,
    $core.String? postalCode,
    $core.bool? isDefault,
  }) {
    final result = AddAddressRequest._();
    if (customerId != null) result.customerId = customerId;
    if (type != null) result.type = type;
    if (recipientName != null) result.recipientName = recipientName;
    if (phone != null) result.phone = phone;
    if (addressLine != null) result.addressLine = addressLine;
    if (city != null) result.city = city;
    if (postalCode != null) result.postalCode = postalCode;
    if (isDefault != null) result.isDefault = isDefault;
    return result;
  }

  AddAddressRequest._();

  factory AddAddressRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddAddressRequest()..mergeFromBuffer(data, registry);
  factory AddAddressRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddAddressRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddAddressRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: AddAddressRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'recipientName')
    ..aOS(4, _omitFieldNames ? '' : 'phone')
    ..aOS(5, _omitFieldNames ? '' : 'addressLine')
    ..aOS(6, _omitFieldNames ? '' : 'city')
    ..aOS(7, _omitFieldNames ? '' : 'postalCode')
    ..aOB(8, _omitFieldNames ? '' : 'isDefault')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddAddressRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddAddressRequest copyWith(void Function(AddAddressRequest) updates) =>
      super.copyWith((message) => updates(message as AddAddressRequest))
          as AddAddressRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AddAddressRequest() / AddAddressRequest.new instead')
  static AddAddressRequest create() => AddAddressRequest._();
  static $pb.GeneratedMessage $_createMessage() => AddAddressRequest._();
  @$core.override
  AddAddressRequest createEmptyInstance() => AddAddressRequest._();
  @$core.pragma('dart2js:noInline')
  static AddAddressRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AddAddressRequest>(
          AddAddressRequest.$_createMessage);
  static AddAddressRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get recipientName => $_getSZ(2);
  @$pb.TagNumber(3)
  set recipientName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRecipientName() => $_has(2);
  @$pb.TagNumber(3)
  void clearRecipientName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get phone => $_getSZ(3);
  @$pb.TagNumber(4)
  set phone($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPhone() => $_has(3);
  @$pb.TagNumber(4)
  void clearPhone() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get addressLine => $_getSZ(4);
  @$pb.TagNumber(5)
  set addressLine($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAddressLine() => $_has(4);
  @$pb.TagNumber(5)
  void clearAddressLine() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get city => $_getSZ(5);
  @$pb.TagNumber(6)
  set city($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCity() => $_has(5);
  @$pb.TagNumber(6)
  void clearCity() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get postalCode => $_getSZ(6);
  @$pb.TagNumber(7)
  set postalCode($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPostalCode() => $_has(6);
  @$pb.TagNumber(7)
  void clearPostalCode() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get isDefault => $_getBF(7);
  @$pb.TagNumber(8)
  set isDefault($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasIsDefault() => $_has(7);
  @$pb.TagNumber(8)
  void clearIsDefault() => $_clearField(8);
}

class AddAddressResponse extends $pb.GeneratedMessage {
  factory AddAddressResponse({
    CustomerAddress? address,
  }) {
    final result = AddAddressResponse._();
    if (address != null) result.address = address;
    return result;
  }

  AddAddressResponse._();

  factory AddAddressResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddAddressResponse()..mergeFromBuffer(data, registry);
  factory AddAddressResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddAddressResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddAddressResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: AddAddressResponse.$_createMessage)
    ..aOM<CustomerAddress>(1, _omitFieldNames ? '' : 'address',
        subBuilder: CustomerAddress.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddAddressResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddAddressResponse copyWith(void Function(AddAddressResponse) updates) =>
      super.copyWith((message) => updates(message as AddAddressResponse))
          as AddAddressResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AddAddressResponse() / AddAddressResponse.new instead')
  static AddAddressResponse create() => AddAddressResponse._();
  static $pb.GeneratedMessage $_createMessage() => AddAddressResponse._();
  @$core.override
  AddAddressResponse createEmptyInstance() => AddAddressResponse._();
  @$core.pragma('dart2js:noInline')
  static AddAddressResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AddAddressResponse>(
          AddAddressResponse.$_createMessage);
  static AddAddressResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerAddress get address => $_getN(0);
  @$pb.TagNumber(1)
  set address(CustomerAddress value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerAddress ensureAddress() => $_ensure(0);
}

class UpdateAddressRequest extends $pb.GeneratedMessage {
  factory UpdateAddressRequest({
    $core.String? id,
    $core.String? type,
    $core.String? recipientName,
    $core.String? phone,
    $core.String? addressLine,
    $core.String? city,
    $core.String? postalCode,
    $core.bool? isDefault,
  }) {
    final result = UpdateAddressRequest._();
    if (id != null) result.id = id;
    if (type != null) result.type = type;
    if (recipientName != null) result.recipientName = recipientName;
    if (phone != null) result.phone = phone;
    if (addressLine != null) result.addressLine = addressLine;
    if (city != null) result.city = city;
    if (postalCode != null) result.postalCode = postalCode;
    if (isDefault != null) result.isDefault = isDefault;
    return result;
  }

  UpdateAddressRequest._();

  factory UpdateAddressRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAddressRequest()..mergeFromBuffer(data, registry);
  factory UpdateAddressRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAddressRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAddressRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: UpdateAddressRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'recipientName')
    ..aOS(4, _omitFieldNames ? '' : 'phone')
    ..aOS(5, _omitFieldNames ? '' : 'addressLine')
    ..aOS(6, _omitFieldNames ? '' : 'city')
    ..aOS(7, _omitFieldNames ? '' : 'postalCode')
    ..aOB(8, _omitFieldNames ? '' : 'isDefault')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAddressRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAddressRequest copyWith(void Function(UpdateAddressRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateAddressRequest))
          as UpdateAddressRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAddressRequest() / UpdateAddressRequest.new instead')
  static UpdateAddressRequest create() => UpdateAddressRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAddressRequest._();
  @$core.override
  UpdateAddressRequest createEmptyInstance() => UpdateAddressRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAddressRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAddressRequest>(
          UpdateAddressRequest.$_createMessage);
  static UpdateAddressRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get recipientName => $_getSZ(2);
  @$pb.TagNumber(3)
  set recipientName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRecipientName() => $_has(2);
  @$pb.TagNumber(3)
  void clearRecipientName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get phone => $_getSZ(3);
  @$pb.TagNumber(4)
  set phone($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPhone() => $_has(3);
  @$pb.TagNumber(4)
  void clearPhone() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get addressLine => $_getSZ(4);
  @$pb.TagNumber(5)
  set addressLine($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAddressLine() => $_has(4);
  @$pb.TagNumber(5)
  void clearAddressLine() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get city => $_getSZ(5);
  @$pb.TagNumber(6)
  set city($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCity() => $_has(5);
  @$pb.TagNumber(6)
  void clearCity() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get postalCode => $_getSZ(6);
  @$pb.TagNumber(7)
  set postalCode($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPostalCode() => $_has(6);
  @$pb.TagNumber(7)
  void clearPostalCode() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get isDefault => $_getBF(7);
  @$pb.TagNumber(8)
  set isDefault($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasIsDefault() => $_has(7);
  @$pb.TagNumber(8)
  void clearIsDefault() => $_clearField(8);
}

class UpdateAddressResponse extends $pb.GeneratedMessage {
  factory UpdateAddressResponse({
    CustomerAddress? address,
  }) {
    final result = UpdateAddressResponse._();
    if (address != null) result.address = address;
    return result;
  }

  UpdateAddressResponse._();

  factory UpdateAddressResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAddressResponse()..mergeFromBuffer(data, registry);
  factory UpdateAddressResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAddressResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAddressResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: UpdateAddressResponse.$_createMessage)
    ..aOM<CustomerAddress>(1, _omitFieldNames ? '' : 'address',
        subBuilder: CustomerAddress.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAddressResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAddressResponse copyWith(
          void Function(UpdateAddressResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateAddressResponse))
          as UpdateAddressResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAddressResponse() / UpdateAddressResponse.new instead')
  static UpdateAddressResponse create() => UpdateAddressResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAddressResponse._();
  @$core.override
  UpdateAddressResponse createEmptyInstance() => UpdateAddressResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAddressResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAddressResponse>(
          UpdateAddressResponse.$_createMessage);
  static UpdateAddressResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerAddress get address => $_getN(0);
  @$pb.TagNumber(1)
  set address(CustomerAddress value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerAddress ensureAddress() => $_ensure(0);
}

class DeleteAddressRequest extends $pb.GeneratedMessage {
  factory DeleteAddressRequest({
    $core.String? id,
  }) {
    final result = DeleteAddressRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteAddressRequest._();

  factory DeleteAddressRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAddressRequest()..mergeFromBuffer(data, registry);
  factory DeleteAddressRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAddressRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAddressRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeleteAddressRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAddressRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAddressRequest copyWith(void Function(DeleteAddressRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteAddressRequest))
          as DeleteAddressRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAddressRequest() / DeleteAddressRequest.new instead')
  static DeleteAddressRequest create() => DeleteAddressRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAddressRequest._();
  @$core.override
  DeleteAddressRequest createEmptyInstance() => DeleteAddressRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteAddressRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAddressRequest>(
          DeleteAddressRequest.$_createMessage);
  static DeleteAddressRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteAddressResponse extends $pb.GeneratedMessage {
  factory DeleteAddressResponse() => DeleteAddressResponse._();

  DeleteAddressResponse._();

  factory DeleteAddressResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAddressResponse()..mergeFromBuffer(data, registry);
  factory DeleteAddressResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAddressResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAddressResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeleteAddressResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAddressResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAddressResponse copyWith(
          void Function(DeleteAddressResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteAddressResponse))
          as DeleteAddressResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAddressResponse() / DeleteAddressResponse.new instead')
  static DeleteAddressResponse create() => DeleteAddressResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAddressResponse._();
  @$core.override
  DeleteAddressResponse createEmptyInstance() => DeleteAddressResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteAddressResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAddressResponse>(
          DeleteAddressResponse.$_createMessage);
  static DeleteAddressResponse? _defaultInstance;
}

/// CustomerContact:客戶聯絡人(3.2.2)。
class CustomerContact extends $pb.GeneratedMessage {
  factory CustomerContact({
    $core.String? id,
    $core.String? customerId,
    $core.String? name,
    $core.String? title,
    $core.String? email,
    $core.String? phone,
    $core.bool? isDefault,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? deletedAt,
  }) {
    final result = CustomerContact._();
    if (id != null) result.id = id;
    if (customerId != null) result.customerId = customerId;
    if (name != null) result.name = name;
    if (title != null) result.title = title;
    if (email != null) result.email = email;
    if (phone != null) result.phone = phone;
    if (isDefault != null) result.isDefault = isDefault;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    return result;
  }

  CustomerContact._();

  factory CustomerContact.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerContact()..mergeFromBuffer(data, registry);
  factory CustomerContact.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerContact()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CustomerContact',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CustomerContact.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'customerId')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'title')
    ..aOS(5, _omitFieldNames ? '' : 'email')
    ..aOS(6, _omitFieldNames ? '' : 'phone')
    ..aOB(7, _omitFieldNames ? '' : 'isDefault')
    ..aOS(8, _omitFieldNames ? '' : 'createdAt')
    ..aOS(9, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(10, _omitFieldNames ? '' : 'deletedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerContact clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerContact copyWith(void Function(CustomerContact) updates) =>
      super.copyWith((message) => updates(message as CustomerContact))
          as CustomerContact;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CustomerContact() / CustomerContact.new instead')
  static CustomerContact create() => CustomerContact._();
  static $pb.GeneratedMessage $_createMessage() => CustomerContact._();
  @$core.override
  CustomerContact createEmptyInstance() => CustomerContact._();
  @$core.pragma('dart2js:noInline')
  static CustomerContact getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CustomerContact>(
          CustomerContact.$_createMessage);
  static CustomerContact? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get customerId => $_getSZ(1);
  @$pb.TagNumber(2)
  set customerId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCustomerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearCustomerId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get title => $_getSZ(3);
  @$pb.TagNumber(4)
  set title($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTitle() => $_has(3);
  @$pb.TagNumber(4)
  void clearTitle() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get email => $_getSZ(4);
  @$pb.TagNumber(5)
  set email($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasEmail() => $_has(4);
  @$pb.TagNumber(5)
  void clearEmail() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get phone => $_getSZ(5);
  @$pb.TagNumber(6)
  set phone($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPhone() => $_has(5);
  @$pb.TagNumber(6)
  void clearPhone() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get isDefault => $_getBF(6);
  @$pb.TagNumber(7)
  set isDefault($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasIsDefault() => $_has(6);
  @$pb.TagNumber(7)
  void clearIsDefault() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get createdAt => $_getSZ(7);
  @$pb.TagNumber(8)
  set createdAt($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasCreatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCreatedAt() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get updatedAt => $_getSZ(8);
  @$pb.TagNumber(9)
  set updatedAt($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUpdatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpdatedAt() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get deletedAt => $_getSZ(9);
  @$pb.TagNumber(10)
  set deletedAt($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasDeletedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearDeletedAt() => $_clearField(10);
}

class ListContactsRequest extends $pb.GeneratedMessage {
  factory ListContactsRequest({
    $core.String? customerId,
    $core.bool? includeDeleted,
  }) {
    final result = ListContactsRequest._();
    if (customerId != null) result.customerId = customerId;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    return result;
  }

  ListContactsRequest._();

  factory ListContactsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListContactsRequest()..mergeFromBuffer(data, registry);
  factory ListContactsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListContactsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListContactsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListContactsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOB(2, _omitFieldNames ? '' : 'includeDeleted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListContactsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListContactsRequest copyWith(void Function(ListContactsRequest) updates) =>
      super.copyWith((message) => updates(message as ListContactsRequest))
          as ListContactsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListContactsRequest() / ListContactsRequest.new instead')
  static ListContactsRequest create() => ListContactsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListContactsRequest._();
  @$core.override
  ListContactsRequest createEmptyInstance() => ListContactsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListContactsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListContactsRequest>(
          ListContactsRequest.$_createMessage);
  static ListContactsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get includeDeleted => $_getBF(1);
  @$pb.TagNumber(2)
  set includeDeleted($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasIncludeDeleted() => $_has(1);
  @$pb.TagNumber(2)
  void clearIncludeDeleted() => $_clearField(2);
}

class ListContactsResponse extends $pb.GeneratedMessage {
  factory ListContactsResponse({
    $core.Iterable<CustomerContact>? contacts,
  }) {
    final result = ListContactsResponse._();
    if (contacts != null) result.contacts.addAll(contacts);
    return result;
  }

  ListContactsResponse._();

  factory ListContactsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListContactsResponse()..mergeFromBuffer(data, registry);
  factory ListContactsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListContactsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListContactsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListContactsResponse.$_createMessage)
    ..pPM<CustomerContact>(1, _omitFieldNames ? '' : 'contacts',
        subBuilder: CustomerContact.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListContactsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListContactsResponse copyWith(void Function(ListContactsResponse) updates) =>
      super.copyWith((message) => updates(message as ListContactsResponse))
          as ListContactsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListContactsResponse() / ListContactsResponse.new instead')
  static ListContactsResponse create() => ListContactsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListContactsResponse._();
  @$core.override
  ListContactsResponse createEmptyInstance() => ListContactsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListContactsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListContactsResponse>(
          ListContactsResponse.$_createMessage);
  static ListContactsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CustomerContact> get contacts => $_getList(0);
}

class AddContactRequest extends $pb.GeneratedMessage {
  factory AddContactRequest({
    $core.String? customerId,
    $core.String? name,
    $core.String? title,
    $core.String? email,
    $core.String? phone,
    $core.bool? isDefault,
  }) {
    final result = AddContactRequest._();
    if (customerId != null) result.customerId = customerId;
    if (name != null) result.name = name;
    if (title != null) result.title = title;
    if (email != null) result.email = email;
    if (phone != null) result.phone = phone;
    if (isDefault != null) result.isDefault = isDefault;
    return result;
  }

  AddContactRequest._();

  factory AddContactRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddContactRequest()..mergeFromBuffer(data, registry);
  factory AddContactRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddContactRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddContactRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: AddContactRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'title')
    ..aOS(4, _omitFieldNames ? '' : 'email')
    ..aOS(5, _omitFieldNames ? '' : 'phone')
    ..aOB(6, _omitFieldNames ? '' : 'isDefault')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddContactRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddContactRequest copyWith(void Function(AddContactRequest) updates) =>
      super.copyWith((message) => updates(message as AddContactRequest))
          as AddContactRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AddContactRequest() / AddContactRequest.new instead')
  static AddContactRequest create() => AddContactRequest._();
  static $pb.GeneratedMessage $_createMessage() => AddContactRequest._();
  @$core.override
  AddContactRequest createEmptyInstance() => AddContactRequest._();
  @$core.pragma('dart2js:noInline')
  static AddContactRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AddContactRequest>(
          AddContactRequest.$_createMessage);
  static AddContactRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(3)
  set title($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitle() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get email => $_getSZ(3);
  @$pb.TagNumber(4)
  set email($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEmail() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmail() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get phone => $_getSZ(4);
  @$pb.TagNumber(5)
  set phone($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPhone() => $_has(4);
  @$pb.TagNumber(5)
  void clearPhone() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get isDefault => $_getBF(5);
  @$pb.TagNumber(6)
  set isDefault($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasIsDefault() => $_has(5);
  @$pb.TagNumber(6)
  void clearIsDefault() => $_clearField(6);
}

class AddContactResponse extends $pb.GeneratedMessage {
  factory AddContactResponse({
    CustomerContact? contact,
  }) {
    final result = AddContactResponse._();
    if (contact != null) result.contact = contact;
    return result;
  }

  AddContactResponse._();

  factory AddContactResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddContactResponse()..mergeFromBuffer(data, registry);
  factory AddContactResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddContactResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddContactResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: AddContactResponse.$_createMessage)
    ..aOM<CustomerContact>(1, _omitFieldNames ? '' : 'contact',
        subBuilder: CustomerContact.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddContactResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddContactResponse copyWith(void Function(AddContactResponse) updates) =>
      super.copyWith((message) => updates(message as AddContactResponse))
          as AddContactResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AddContactResponse() / AddContactResponse.new instead')
  static AddContactResponse create() => AddContactResponse._();
  static $pb.GeneratedMessage $_createMessage() => AddContactResponse._();
  @$core.override
  AddContactResponse createEmptyInstance() => AddContactResponse._();
  @$core.pragma('dart2js:noInline')
  static AddContactResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AddContactResponse>(
          AddContactResponse.$_createMessage);
  static AddContactResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerContact get contact => $_getN(0);
  @$pb.TagNumber(1)
  set contact(CustomerContact value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasContact() => $_has(0);
  @$pb.TagNumber(1)
  void clearContact() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerContact ensureContact() => $_ensure(0);
}

class UpdateContactRequest extends $pb.GeneratedMessage {
  factory UpdateContactRequest({
    $core.String? id,
    $core.String? name,
    $core.String? title,
    $core.String? email,
    $core.String? phone,
    $core.bool? isDefault,
  }) {
    final result = UpdateContactRequest._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (title != null) result.title = title;
    if (email != null) result.email = email;
    if (phone != null) result.phone = phone;
    if (isDefault != null) result.isDefault = isDefault;
    return result;
  }

  UpdateContactRequest._();

  factory UpdateContactRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateContactRequest()..mergeFromBuffer(data, registry);
  factory UpdateContactRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateContactRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateContactRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: UpdateContactRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'title')
    ..aOS(4, _omitFieldNames ? '' : 'email')
    ..aOS(5, _omitFieldNames ? '' : 'phone')
    ..aOB(6, _omitFieldNames ? '' : 'isDefault')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateContactRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateContactRequest copyWith(void Function(UpdateContactRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateContactRequest))
          as UpdateContactRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateContactRequest() / UpdateContactRequest.new instead')
  static UpdateContactRequest create() => UpdateContactRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateContactRequest._();
  @$core.override
  UpdateContactRequest createEmptyInstance() => UpdateContactRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateContactRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateContactRequest>(
          UpdateContactRequest.$_createMessage);
  static UpdateContactRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(3)
  set title($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitle() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get email => $_getSZ(3);
  @$pb.TagNumber(4)
  set email($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEmail() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmail() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get phone => $_getSZ(4);
  @$pb.TagNumber(5)
  set phone($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPhone() => $_has(4);
  @$pb.TagNumber(5)
  void clearPhone() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get isDefault => $_getBF(5);
  @$pb.TagNumber(6)
  set isDefault($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasIsDefault() => $_has(5);
  @$pb.TagNumber(6)
  void clearIsDefault() => $_clearField(6);
}

class UpdateContactResponse extends $pb.GeneratedMessage {
  factory UpdateContactResponse({
    CustomerContact? contact,
  }) {
    final result = UpdateContactResponse._();
    if (contact != null) result.contact = contact;
    return result;
  }

  UpdateContactResponse._();

  factory UpdateContactResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateContactResponse()..mergeFromBuffer(data, registry);
  factory UpdateContactResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateContactResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateContactResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: UpdateContactResponse.$_createMessage)
    ..aOM<CustomerContact>(1, _omitFieldNames ? '' : 'contact',
        subBuilder: CustomerContact.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateContactResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateContactResponse copyWith(
          void Function(UpdateContactResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateContactResponse))
          as UpdateContactResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateContactResponse() / UpdateContactResponse.new instead')
  static UpdateContactResponse create() => UpdateContactResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateContactResponse._();
  @$core.override
  UpdateContactResponse createEmptyInstance() => UpdateContactResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateContactResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateContactResponse>(
          UpdateContactResponse.$_createMessage);
  static UpdateContactResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerContact get contact => $_getN(0);
  @$pb.TagNumber(1)
  set contact(CustomerContact value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasContact() => $_has(0);
  @$pb.TagNumber(1)
  void clearContact() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerContact ensureContact() => $_ensure(0);
}

class DeleteContactRequest extends $pb.GeneratedMessage {
  factory DeleteContactRequest({
    $core.String? id,
  }) {
    final result = DeleteContactRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteContactRequest._();

  factory DeleteContactRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteContactRequest()..mergeFromBuffer(data, registry);
  factory DeleteContactRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteContactRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteContactRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeleteContactRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteContactRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteContactRequest copyWith(void Function(DeleteContactRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteContactRequest))
          as DeleteContactRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteContactRequest() / DeleteContactRequest.new instead')
  static DeleteContactRequest create() => DeleteContactRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteContactRequest._();
  @$core.override
  DeleteContactRequest createEmptyInstance() => DeleteContactRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteContactRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteContactRequest>(
          DeleteContactRequest.$_createMessage);
  static DeleteContactRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteContactResponse extends $pb.GeneratedMessage {
  factory DeleteContactResponse() => DeleteContactResponse._();

  DeleteContactResponse._();

  factory DeleteContactResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteContactResponse()..mergeFromBuffer(data, registry);
  factory DeleteContactResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteContactResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteContactResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeleteContactResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteContactResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteContactResponse copyWith(
          void Function(DeleteContactResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteContactResponse))
          as DeleteContactResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteContactResponse() / DeleteContactResponse.new instead')
  static DeleteContactResponse create() => DeleteContactResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteContactResponse._();
  @$core.override
  DeleteContactResponse createEmptyInstance() => DeleteContactResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteContactResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteContactResponse>(
          DeleteContactResponse.$_createMessage);
  static DeleteContactResponse? _defaultInstance;
}

/// GetCustomerQRCodeRequest:為指定客戶產生登入 QR(3.8.2 產生端)。
class GetCustomerQRCodeRequest extends $pb.GeneratedMessage {
  factory GetCustomerQRCodeRequest({
    $core.String? customerId,
  }) {
    final result = GetCustomerQRCodeRequest._();
    if (customerId != null) result.customerId = customerId;
    return result;
  }

  GetCustomerQRCodeRequest._();

  factory GetCustomerQRCodeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerQRCodeRequest()..mergeFromBuffer(data, registry);
  factory GetCustomerQRCodeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerQRCodeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCustomerQRCodeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: GetCustomerQRCodeRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'customerId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerQRCodeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerQRCodeRequest copyWith(
          void Function(GetCustomerQRCodeRequest) updates) =>
      super.copyWith((message) => updates(message as GetCustomerQRCodeRequest))
          as GetCustomerQRCodeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCustomerQRCodeRequest() / GetCustomerQRCodeRequest.new instead')
  static GetCustomerQRCodeRequest create() => GetCustomerQRCodeRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetCustomerQRCodeRequest._();
  @$core.override
  GetCustomerQRCodeRequest createEmptyInstance() =>
      GetCustomerQRCodeRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCustomerQRCodeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCustomerQRCodeRequest>(
          GetCustomerQRCodeRequest.$_createMessage);
  static GetCustomerQRCodeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get customerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set customerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCustomerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCustomerId() => $_clearField(1);
}

/// GetCustomerQRCodeResponse:深層連結(App 未裝導商店、已裝直開)。
class GetCustomerQRCodeResponse extends $pb.GeneratedMessage {
  factory GetCustomerQRCodeResponse({
    $core.String? qrUrl,
  }) {
    final result = GetCustomerQRCodeResponse._();
    if (qrUrl != null) result.qrUrl = qrUrl;
    return result;
  }

  GetCustomerQRCodeResponse._();

  factory GetCustomerQRCodeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerQRCodeResponse()..mergeFromBuffer(data, registry);
  factory GetCustomerQRCodeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCustomerQRCodeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCustomerQRCodeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: GetCustomerQRCodeResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'qrUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerQRCodeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerQRCodeResponse copyWith(
          void Function(GetCustomerQRCodeResponse) updates) =>
      super.copyWith((message) => updates(message as GetCustomerQRCodeResponse))
          as GetCustomerQRCodeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCustomerQRCodeResponse() / GetCustomerQRCodeResponse.new instead')
  static GetCustomerQRCodeResponse create() => GetCustomerQRCodeResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetCustomerQRCodeResponse._();
  @$core.override
  GetCustomerQRCodeResponse createEmptyInstance() =>
      GetCustomerQRCodeResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCustomerQRCodeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCustomerQRCodeResponse>(
          GetCustomerQRCodeResponse.$_createMessage);
  static GetCustomerQRCodeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get qrUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set qrUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasQrUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearQrUrl() => $_clearField(1);
}

/// CustomerAccount:登入帳號(不含密碼欄位)。
class CustomerAccount extends $pb.GeneratedMessage {
  factory CustomerAccount({
    $core.String? id,
    $core.String? accountName,
    $core.bool? isPrimary,
    $core.bool? systemGenerated,
    $core.bool? manageable,
    $core.String? status,
    $core.String? createdAt,
  }) {
    final result = CustomerAccount._();
    if (id != null) result.id = id;
    if (accountName != null) result.accountName = accountName;
    if (isPrimary != null) result.isPrimary = isPrimary;
    if (systemGenerated != null) result.systemGenerated = systemGenerated;
    if (manageable != null) result.manageable = manageable;
    if (status != null) result.status = status;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  CustomerAccount._();

  factory CustomerAccount.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerAccount()..mergeFromBuffer(data, registry);
  factory CustomerAccount.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CustomerAccount()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CustomerAccount',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CustomerAccount.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'accountName')
    ..aOB(3, _omitFieldNames ? '' : 'isPrimary')
    ..aOB(4, _omitFieldNames ? '' : 'systemGenerated')
    ..aOB(5, _omitFieldNames ? '' : 'manageable')
    ..aOS(6, _omitFieldNames ? '' : 'status')
    ..aOS(7, _omitFieldNames ? '' : 'createdAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerAccount clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CustomerAccount copyWith(void Function(CustomerAccount) updates) =>
      super.copyWith((message) => updates(message as CustomerAccount))
          as CustomerAccount;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CustomerAccount() / CustomerAccount.new instead')
  static CustomerAccount create() => CustomerAccount._();
  static $pb.GeneratedMessage $_createMessage() => CustomerAccount._();
  @$core.override
  CustomerAccount createEmptyInstance() => CustomerAccount._();
  @$core.pragma('dart2js:noInline')
  static CustomerAccount getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CustomerAccount>(
          CustomerAccount.$_createMessage);
  static CustomerAccount? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get accountName => $_getSZ(1);
  @$pb.TagNumber(2)
  set accountName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAccountName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAccountName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get isPrimary => $_getBF(2);
  @$pb.TagNumber(3)
  set isPrimary($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasIsPrimary() => $_has(2);
  @$pb.TagNumber(3)
  void clearIsPrimary() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get systemGenerated => $_getBF(3);
  @$pb.TagNumber(4)
  set systemGenerated($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSystemGenerated() => $_has(3);
  @$pb.TagNumber(4)
  void clearSystemGenerated() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get manageable => $_getBF(4);
  @$pb.TagNumber(5)
  set manageable($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasManageable() => $_has(4);
  @$pb.TagNumber(5)
  void clearManageable() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get status => $_getSZ(5);
  @$pb.TagNumber(6)
  set status($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get createdAt => $_getSZ(6);
  @$pb.TagNumber(7)
  set createdAt($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCreatedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearCreatedAt() => $_clearField(7);
}

class ListCustomerAccountsRequest extends $pb.GeneratedMessage {
  factory ListCustomerAccountsRequest() => ListCustomerAccountsRequest._();

  ListCustomerAccountsRequest._();

  factory ListCustomerAccountsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerAccountsRequest()..mergeFromBuffer(data, registry);
  factory ListCustomerAccountsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerAccountsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomerAccountsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListCustomerAccountsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerAccountsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerAccountsRequest copyWith(
          void Function(ListCustomerAccountsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListCustomerAccountsRequest))
          as ListCustomerAccountsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCustomerAccountsRequest() / ListCustomerAccountsRequest.new instead')
  static ListCustomerAccountsRequest create() =>
      ListCustomerAccountsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListCustomerAccountsRequest._();
  @$core.override
  ListCustomerAccountsRequest createEmptyInstance() =>
      ListCustomerAccountsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListCustomerAccountsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomerAccountsRequest>(
          ListCustomerAccountsRequest.$_createMessage);
  static ListCustomerAccountsRequest? _defaultInstance;
}

class ListCustomerAccountsResponse extends $pb.GeneratedMessage {
  factory ListCustomerAccountsResponse({
    $core.Iterable<CustomerAccount>? accounts,
  }) {
    final result = ListCustomerAccountsResponse._();
    if (accounts != null) result.accounts.addAll(accounts);
    return result;
  }

  ListCustomerAccountsResponse._();

  factory ListCustomerAccountsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerAccountsResponse()..mergeFromBuffer(data, registry);
  factory ListCustomerAccountsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCustomerAccountsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomerAccountsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ListCustomerAccountsResponse.$_createMessage)
    ..pPM<CustomerAccount>(1, _omitFieldNames ? '' : 'accounts',
        subBuilder: CustomerAccount.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerAccountsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomerAccountsResponse copyWith(
          void Function(ListCustomerAccountsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListCustomerAccountsResponse))
          as ListCustomerAccountsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCustomerAccountsResponse() / ListCustomerAccountsResponse.new instead')
  static ListCustomerAccountsResponse create() =>
      ListCustomerAccountsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListCustomerAccountsResponse._();
  @$core.override
  ListCustomerAccountsResponse createEmptyInstance() =>
      ListCustomerAccountsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListCustomerAccountsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomerAccountsResponse>(
          ListCustomerAccountsResponse.$_createMessage);
  static ListCustomerAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CustomerAccount> get accounts => $_getList(0);
}

class CreateCustomerAccountRequest extends $pb.GeneratedMessage {
  factory CreateCustomerAccountRequest({
    $core.String? accountName,
  }) {
    final result = CreateCustomerAccountRequest._();
    if (accountName != null) result.accountName = accountName;
    return result;
  }

  CreateCustomerAccountRequest._();

  factory CreateCustomerAccountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerAccountRequest()..mergeFromBuffer(data, registry);
  factory CreateCustomerAccountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerAccountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCustomerAccountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CreateCustomerAccountRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerAccountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerAccountRequest copyWith(
          void Function(CreateCustomerAccountRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCustomerAccountRequest))
          as CreateCustomerAccountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCustomerAccountRequest() / CreateCustomerAccountRequest.new instead')
  static CreateCustomerAccountRequest create() =>
      CreateCustomerAccountRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCustomerAccountRequest._();
  @$core.override
  CreateCustomerAccountRequest createEmptyInstance() =>
      CreateCustomerAccountRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCustomerAccountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCustomerAccountRequest>(
          CreateCustomerAccountRequest.$_createMessage);
  static CreateCustomerAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountName => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountName() => $_clearField(1);
}

class CreateCustomerAccountResponse extends $pb.GeneratedMessage {
  factory CreateCustomerAccountResponse({
    CustomerAccount? account,
    $core.String? tempPassword,
    $core.String? tempExpiresAt,
  }) {
    final result = CreateCustomerAccountResponse._();
    if (account != null) result.account = account;
    if (tempPassword != null) result.tempPassword = tempPassword;
    if (tempExpiresAt != null) result.tempExpiresAt = tempExpiresAt;
    return result;
  }

  CreateCustomerAccountResponse._();

  factory CreateCustomerAccountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerAccountResponse()..mergeFromBuffer(data, registry);
  factory CreateCustomerAccountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCustomerAccountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCustomerAccountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: CreateCustomerAccountResponse.$_createMessage)
    ..aOM<CustomerAccount>(1, _omitFieldNames ? '' : 'account',
        subBuilder: CustomerAccount.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'tempPassword')
    ..aOS(3, _omitFieldNames ? '' : 'tempExpiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerAccountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerAccountResponse copyWith(
          void Function(CreateCustomerAccountResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateCustomerAccountResponse))
          as CreateCustomerAccountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCustomerAccountResponse() / CreateCustomerAccountResponse.new instead')
  static CreateCustomerAccountResponse create() =>
      CreateCustomerAccountResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateCustomerAccountResponse._();
  @$core.override
  CreateCustomerAccountResponse createEmptyInstance() =>
      CreateCustomerAccountResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCustomerAccountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCustomerAccountResponse>(
          CreateCustomerAccountResponse.$_createMessage);
  static CreateCustomerAccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CustomerAccount get account => $_getN(0);
  @$pb.TagNumber(1)
  set account(CustomerAccount value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccount() => $_clearField(1);
  @$pb.TagNumber(1)
  CustomerAccount ensureAccount() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get tempPassword => $_getSZ(1);
  @$pb.TagNumber(2)
  set tempPassword($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTempPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearTempPassword() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get tempExpiresAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set tempExpiresAt($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTempExpiresAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearTempExpiresAt() => $_clearField(3);
}

class DeactivateCustomerAccountRequest extends $pb.GeneratedMessage {
  factory DeactivateCustomerAccountRequest({
    $core.String? accountId,
  }) {
    final result = DeactivateCustomerAccountRequest._();
    if (accountId != null) result.accountId = accountId;
    return result;
  }

  DeactivateCustomerAccountRequest._();

  factory DeactivateCustomerAccountRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeactivateCustomerAccountRequest()..mergeFromBuffer(data, registry);
  factory DeactivateCustomerAccountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeactivateCustomerAccountRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeactivateCustomerAccountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeactivateCustomerAccountRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateCustomerAccountRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateCustomerAccountRequest copyWith(
          void Function(DeactivateCustomerAccountRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeactivateCustomerAccountRequest))
          as DeactivateCustomerAccountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeactivateCustomerAccountRequest() / DeactivateCustomerAccountRequest.new instead')
  static DeactivateCustomerAccountRequest create() =>
      DeactivateCustomerAccountRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeactivateCustomerAccountRequest._();
  @$core.override
  DeactivateCustomerAccountRequest createEmptyInstance() =>
      DeactivateCustomerAccountRequest._();
  @$core.pragma('dart2js:noInline')
  static DeactivateCustomerAccountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeactivateCustomerAccountRequest>(
          DeactivateCustomerAccountRequest.$_createMessage);
  static DeactivateCustomerAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);
}

class DeactivateCustomerAccountResponse extends $pb.GeneratedMessage {
  factory DeactivateCustomerAccountResponse() =>
      DeactivateCustomerAccountResponse._();

  DeactivateCustomerAccountResponse._();

  factory DeactivateCustomerAccountResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeactivateCustomerAccountResponse()..mergeFromBuffer(data, registry);
  factory DeactivateCustomerAccountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeactivateCustomerAccountResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeactivateCustomerAccountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: DeactivateCustomerAccountResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateCustomerAccountResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeactivateCustomerAccountResponse copyWith(
          void Function(DeactivateCustomerAccountResponse) updates) =>
      super.copyWith((message) =>
              updates(message as DeactivateCustomerAccountResponse))
          as DeactivateCustomerAccountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeactivateCustomerAccountResponse() / DeactivateCustomerAccountResponse.new instead')
  static DeactivateCustomerAccountResponse create() =>
      DeactivateCustomerAccountResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeactivateCustomerAccountResponse._();
  @$core.override
  DeactivateCustomerAccountResponse createEmptyInstance() =>
      DeactivateCustomerAccountResponse._();
  @$core.pragma('dart2js:noInline')
  static DeactivateCustomerAccountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeactivateCustomerAccountResponse>(
          DeactivateCustomerAccountResponse.$_createMessage);
  static DeactivateCustomerAccountResponse? _defaultInstance;
}

class ResetCustomerAccountPasswordRequest extends $pb.GeneratedMessage {
  factory ResetCustomerAccountPasswordRequest({
    $core.String? accountId,
  }) {
    final result = ResetCustomerAccountPasswordRequest._();
    if (accountId != null) result.accountId = accountId;
    return result;
  }

  ResetCustomerAccountPasswordRequest._();

  factory ResetCustomerAccountPasswordRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResetCustomerAccountPasswordRequest()..mergeFromBuffer(data, registry);
  factory ResetCustomerAccountPasswordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResetCustomerAccountPasswordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResetCustomerAccountPasswordRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ResetCustomerAccountPasswordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'accountId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResetCustomerAccountPasswordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResetCustomerAccountPasswordRequest copyWith(
          void Function(ResetCustomerAccountPasswordRequest) updates) =>
      super.copyWith((message) =>
              updates(message as ResetCustomerAccountPasswordRequest))
          as ResetCustomerAccountPasswordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResetCustomerAccountPasswordRequest() / ResetCustomerAccountPasswordRequest.new instead')
  static ResetCustomerAccountPasswordRequest create() =>
      ResetCustomerAccountPasswordRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ResetCustomerAccountPasswordRequest._();
  @$core.override
  ResetCustomerAccountPasswordRequest createEmptyInstance() =>
      ResetCustomerAccountPasswordRequest._();
  @$core.pragma('dart2js:noInline')
  static ResetCustomerAccountPasswordRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ResetCustomerAccountPasswordRequest>(
          ResetCustomerAccountPasswordRequest.$_createMessage);
  static ResetCustomerAccountPasswordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountId() => $_clearField(1);
}

class ResetCustomerAccountPasswordResponse extends $pb.GeneratedMessage {
  factory ResetCustomerAccountPasswordResponse({
    $core.String? tempPassword,
    $core.String? tempExpiresAt,
  }) {
    final result = ResetCustomerAccountPasswordResponse._();
    if (tempPassword != null) result.tempPassword = tempPassword;
    if (tempExpiresAt != null) result.tempExpiresAt = tempExpiresAt;
    return result;
  }

  ResetCustomerAccountPasswordResponse._();

  factory ResetCustomerAccountPasswordResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResetCustomerAccountPasswordResponse()..mergeFromBuffer(data, registry);
  factory ResetCustomerAccountPasswordResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResetCustomerAccountPasswordResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResetCustomerAccountPasswordResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: ResetCustomerAccountPasswordResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'tempPassword')
    ..aOS(2, _omitFieldNames ? '' : 'tempExpiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResetCustomerAccountPasswordResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResetCustomerAccountPasswordResponse copyWith(
          void Function(ResetCustomerAccountPasswordResponse) updates) =>
      super.copyWith((message) =>
              updates(message as ResetCustomerAccountPasswordResponse))
          as ResetCustomerAccountPasswordResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResetCustomerAccountPasswordResponse() / ResetCustomerAccountPasswordResponse.new instead')
  static ResetCustomerAccountPasswordResponse create() =>
      ResetCustomerAccountPasswordResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ResetCustomerAccountPasswordResponse._();
  @$core.override
  ResetCustomerAccountPasswordResponse createEmptyInstance() =>
      ResetCustomerAccountPasswordResponse._();
  @$core.pragma('dart2js:noInline')
  static ResetCustomerAccountPasswordResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ResetCustomerAccountPasswordResponse>(
          ResetCustomerAccountPasswordResponse.$_createMessage);
  static ResetCustomerAccountPasswordResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get tempPassword => $_getSZ(0);
  @$pb.TagNumber(1)
  set tempPassword($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTempPassword() => $_has(0);
  @$pb.TagNumber(1)
  void clearTempPassword() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get tempExpiresAt => $_getSZ(1);
  @$pb.TagNumber(2)
  set tempExpiresAt($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTempExpiresAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearTempExpiresAt() => $_clearField(2);
}

/// CustomerService:客戶主檔管理(dept_admin/staff 限所屬部門)。
class CustomerServiceApi {
  final $pb.RpcClient _client;

  CustomerServiceApi(this._client);

  /// ListCustomers:分頁查詢(keyword 模糊比對 name/customer_code/tax_id;可 include_deleted)。
  $async.Future<ListCustomersResponse> listCustomers(
          $pb.ClientContext? ctx, ListCustomersRequest request) =>
      _client.invoke<ListCustomersResponse>(ctx, 'CustomerService',
          'ListCustomers', request, ListCustomersResponse());

  /// GetCustomer:以 id 取單筆(限可見範圍)。
  $async.Future<GetCustomerResponse> getCustomer(
          $pb.ClientContext? ctx, GetCustomerRequest request) =>
      _client.invoke<GetCustomerResponse>(ctx, 'CustomerService', 'GetCustomer',
          request, GetCustomerResponse());

  /// CreateCustomer:建立客戶(系統取號;字典/業務 reference 驗證)。
  $async.Future<CreateCustomerResponse> createCustomer(
          $pb.ClientContext? ctx, CreateCustomerRequest request) =>
      _client.invoke<CreateCustomerResponse>(ctx, 'CustomerService',
          'CreateCustomer', request, CreateCustomerResponse());

  /// UpdateCustomer:欄位式更新(customer_code 不可改)。
  $async.Future<UpdateCustomerResponse> updateCustomer(
          $pb.ClientContext? ctx, UpdateCustomerRequest request) =>
      _client.invoke<UpdateCustomerResponse>(ctx, 'CustomerService',
          'UpdateCustomer', request, UpdateCustomerResponse());

  /// DeleteCustomer:軟刪除。
  $async.Future<DeleteCustomerResponse> deleteCustomer(
          $pb.ClientContext? ctx, DeleteCustomerRequest request) =>
      _client.invoke<DeleteCustomerResponse>(ctx, 'CustomerService',
          'DeleteCustomer', request, DeleteCustomerResponse());

  /// RestoreCustomer:復原(清 deleted_at)。
  $async.Future<RestoreCustomerResponse> restoreCustomer(
          $pb.ClientContext? ctx, RestoreCustomerRequest request) =>
      _client.invoke<RestoreCustomerResponse>(ctx, 'CustomerService',
          'RestoreCustomer', request, RestoreCustomerResponse());

  /// 以下為地址簿與聯絡人(3.2.1 / 3.2.2)。
  $async.Future<ListAddressesResponse> listAddresses(
          $pb.ClientContext? ctx, ListAddressesRequest request) =>
      _client.invoke<ListAddressesResponse>(ctx, 'CustomerService',
          'ListAddresses', request, ListAddressesResponse());
  $async.Future<AddAddressResponse> addAddress(
          $pb.ClientContext? ctx, AddAddressRequest request) =>
      _client.invoke<AddAddressResponse>(
          ctx, 'CustomerService', 'AddAddress', request, AddAddressResponse());
  $async.Future<UpdateAddressResponse> updateAddress(
          $pb.ClientContext? ctx, UpdateAddressRequest request) =>
      _client.invoke<UpdateAddressResponse>(ctx, 'CustomerService',
          'UpdateAddress', request, UpdateAddressResponse());
  $async.Future<DeleteAddressResponse> deleteAddress(
          $pb.ClientContext? ctx, DeleteAddressRequest request) =>
      _client.invoke<DeleteAddressResponse>(ctx, 'CustomerService',
          'DeleteAddress', request, DeleteAddressResponse());
  $async.Future<ListContactsResponse> listContacts(
          $pb.ClientContext? ctx, ListContactsRequest request) =>
      _client.invoke<ListContactsResponse>(ctx, 'CustomerService',
          'ListContacts', request, ListContactsResponse());
  $async.Future<AddContactResponse> addContact(
          $pb.ClientContext? ctx, AddContactRequest request) =>
      _client.invoke<AddContactResponse>(
          ctx, 'CustomerService', 'AddContact', request, AddContactResponse());
  $async.Future<UpdateContactResponse> updateContact(
          $pb.ClientContext? ctx, UpdateContactRequest request) =>
      _client.invoke<UpdateContactResponse>(ctx, 'CustomerService',
          'UpdateContact', request, UpdateContactResponse());
  $async.Future<DeleteContactResponse> deleteContact(
          $pb.ClientContext? ctx, DeleteContactRequest request) =>
      _client.invoke<DeleteContactResponse>(ctx, 'CustomerService',
          'DeleteContact', request, DeleteContactResponse());

  /// GetCustomerQRCode:為本部門客戶產生登入 QR(dept_admin/staff 限本部門)。
  $async.Future<GetCustomerQRCodeResponse> getCustomerQRCode(
          $pb.ClientContext? ctx, GetCustomerQRCodeRequest request) =>
      _client.invoke<GetCustomerQRCodeResponse>(ctx, 'CustomerService',
          'GetCustomerQRCode', request, GetCustomerQRCodeResponse());
}

/// ---- CustomerAccountService:店家自助管理登入帳號(D22/規格 4.2,Task 6.7) ----
///
/// 僅**客戶主帳號**可呼叫(is_primary):主帳號是該客戶帳號體系的唯一管理者,也是它唯一被允許的
/// 功能面(業務 API 一律 403,見 server.protectedRPC 與 OpenFGA 的 primary_account 排除)。
/// 範圍僅限自己客戶,不得觸及其他客戶或員工帳號。
///
/// 子帳號無管理權限(本服務對非主帳號一律拒絕);建立客戶時自動附帶的**業務子帳號**
/// (system_generated=true)店家可檢視但不可改名/停用/重置 —— 它專供所屬業務使用,店家並無其密碼。
class CustomerAccountServiceApi {
  final $pb.RpcClient _client;

  CustomerAccountServiceApi(this._client);

  $async.Future<ListCustomerAccountsResponse> listCustomerAccounts(
          $pb.ClientContext? ctx, ListCustomerAccountsRequest request) =>
      _client.invoke<ListCustomerAccountsResponse>(
          ctx,
          'CustomerAccountService',
          'ListCustomerAccounts',
          request,
          ListCustomerAccountsResponse());
  $async.Future<CreateCustomerAccountResponse> createCustomerAccount(
          $pb.ClientContext? ctx, CreateCustomerAccountRequest request) =>
      _client.invoke<CreateCustomerAccountResponse>(
          ctx,
          'CustomerAccountService',
          'CreateCustomerAccount',
          request,
          CreateCustomerAccountResponse());
  $async.Future<DeactivateCustomerAccountResponse> deactivateCustomerAccount(
          $pb.ClientContext? ctx, DeactivateCustomerAccountRequest request) =>
      _client.invoke<DeactivateCustomerAccountResponse>(
          ctx,
          'CustomerAccountService',
          'DeactivateCustomerAccount',
          request,
          DeactivateCustomerAccountResponse());
  $async.Future<ResetCustomerAccountPasswordResponse>
      resetCustomerAccountPassword($pb.ClientContext? ctx,
              ResetCustomerAccountPasswordRequest request) =>
          _client.invoke<ResetCustomerAccountPasswordResponse>(
              ctx,
              'CustomerAccountService',
              'ResetCustomerAccountPassword',
              request,
              ResetCustomerAccountPasswordResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
