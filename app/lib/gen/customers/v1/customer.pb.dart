// This is a generated file - do not edit.
//
// Generated from customers/v1/customer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

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
    final result = create();
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
      create()..mergeFromBuffer(data, registry);
  factory Customer.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Customer',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
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
  Customer clone() => Customer()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Customer copyWith(void Function(Customer) updates) =>
      super.copyWith((message) => updates(message as Customer)) as Customer;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Customer create() => Customer._();
  @$core.override
  Customer createEmptyInstance() => create();
  static $pb.PbList<Customer> createRepeated() => $pb.PbList<Customer>();
  @$core.pragma('dart2js:noInline')
  static Customer getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Customer>(create);
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
  }) {
    final result = create();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (keyword != null) result.keyword = keyword;
    if (includeDeleted != null) result.includeDeleted = includeDeleted;
    if (sort != null) result.sort = sort;
    return result;
  }

  ListCustomersRequest._();

  factory ListCustomersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListCustomersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'page', $pb.PbFieldType.O3)
    ..a<$core.int>(2, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'keyword')
    ..aOB(4, _omitFieldNames ? '' : 'includeDeleted')
    ..aOS(5, _omitFieldNames ? '' : 'sort')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersRequest clone() =>
      ListCustomersRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersRequest copyWith(void Function(ListCustomersRequest) updates) =>
      super.copyWith((message) => updates(message as ListCustomersRequest))
          as ListCustomersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListCustomersRequest create() => ListCustomersRequest._();
  @$core.override
  ListCustomersRequest createEmptyInstance() => create();
  static $pb.PbList<ListCustomersRequest> createRepeated() =>
      $pb.PbList<ListCustomersRequest>();
  @$core.pragma('dart2js:noInline')
  static ListCustomersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomersRequest>(create);
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
}

class ListCustomersResponse extends $pb.GeneratedMessage {
  factory ListCustomersResponse({
    $core.Iterable<Customer>? customers,
    $0.Pagination? pagination,
  }) {
    final result = create();
    if (customers != null) result.customers.addAll(customers);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListCustomersResponse._();

  factory ListCustomersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListCustomersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCustomersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..pc<Customer>(1, _omitFieldNames ? '' : 'customers', $pb.PbFieldType.PM,
        subBuilder: Customer.create)
    ..aOM<$0.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $0.Pagination.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersResponse clone() =>
      ListCustomersResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCustomersResponse copyWith(
          void Function(ListCustomersResponse) updates) =>
      super.copyWith((message) => updates(message as ListCustomersResponse))
          as ListCustomersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListCustomersResponse create() => ListCustomersResponse._();
  @$core.override
  ListCustomersResponse createEmptyInstance() => create();
  static $pb.PbList<ListCustomersResponse> createRepeated() =>
      $pb.PbList<ListCustomersResponse>();
  @$core.pragma('dart2js:noInline')
  static ListCustomersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCustomersResponse>(create);
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
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  GetCustomerRequest._();

  factory GetCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerRequest clone() => GetCustomerRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerRequest copyWith(void Function(GetCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as GetCustomerRequest))
          as GetCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetCustomerRequest create() => GetCustomerRequest._();
  @$core.override
  GetCustomerRequest createEmptyInstance() => create();
  static $pb.PbList<GetCustomerRequest> createRepeated() =>
      $pb.PbList<GetCustomerRequest>();
  @$core.pragma('dart2js:noInline')
  static GetCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCustomerRequest>(create);
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
    final result = create();
    if (customer != null) result.customer = customer;
    return result;
  }

  GetCustomerResponse._();

  factory GetCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerResponse clone() => GetCustomerResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCustomerResponse copyWith(void Function(GetCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as GetCustomerResponse))
          as GetCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetCustomerResponse create() => GetCustomerResponse._();
  @$core.override
  GetCustomerResponse createEmptyInstance() => create();
  static $pb.PbList<GetCustomerResponse> createRepeated() =>
      $pb.PbList<GetCustomerResponse>();
  @$core.pragma('dart2js:noInline')
  static GetCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCustomerResponse>(create);
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
    final result = create();
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
      create()..mergeFromBuffer(data, registry);
  factory CreateCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
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
  CreateCustomerRequest clone() =>
      CreateCustomerRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerRequest copyWith(
          void Function(CreateCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as CreateCustomerRequest))
          as CreateCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateCustomerRequest create() => CreateCustomerRequest._();
  @$core.override
  CreateCustomerRequest createEmptyInstance() => create();
  static $pb.PbList<CreateCustomerRequest> createRepeated() =>
      $pb.PbList<CreateCustomerRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCustomerRequest>(create);
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
    final result = create();
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
      create()..mergeFromBuffer(data, registry);
  factory CreateCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.create)
    ..aOS(2, _omitFieldNames ? '' : 'primaryAccountName')
    ..aOS(3, _omitFieldNames ? '' : 'primaryTempPassword')
    ..aOS(4, _omitFieldNames ? '' : 'salesRepAccountName')
    ..aOS(5, _omitFieldNames ? '' : 'salesRepTempPassword')
    ..aOS(6, _omitFieldNames ? '' : 'accountManageUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerResponse clone() =>
      CreateCustomerResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCustomerResponse copyWith(
          void Function(CreateCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as CreateCustomerResponse))
          as CreateCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateCustomerResponse create() => CreateCustomerResponse._();
  @$core.override
  CreateCustomerResponse createEmptyInstance() => create();
  static $pb.PbList<CreateCustomerResponse> createRepeated() =>
      $pb.PbList<CreateCustomerResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCustomerResponse>(create);
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
    final result = create();
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
      create()..mergeFromBuffer(data, registry);
  factory UpdateCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
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
  UpdateCustomerRequest clone() =>
      UpdateCustomerRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerRequest copyWith(
          void Function(UpdateCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateCustomerRequest))
          as UpdateCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateCustomerRequest create() => UpdateCustomerRequest._();
  @$core.override
  UpdateCustomerRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateCustomerRequest> createRepeated() =>
      $pb.PbList<UpdateCustomerRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCustomerRequest>(create);
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
    final result = create();
    if (customer != null) result.customer = customer;
    return result;
  }

  UpdateCustomerResponse._();

  factory UpdateCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerResponse clone() =>
      UpdateCustomerResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCustomerResponse copyWith(
          void Function(UpdateCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateCustomerResponse))
          as UpdateCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateCustomerResponse create() => UpdateCustomerResponse._();
  @$core.override
  UpdateCustomerResponse createEmptyInstance() => create();
  static $pb.PbList<UpdateCustomerResponse> createRepeated() =>
      $pb.PbList<UpdateCustomerResponse>();
  @$core.pragma('dart2js:noInline')
  static UpdateCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCustomerResponse>(create);
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
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  DeleteCustomerRequest._();

  factory DeleteCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerRequest clone() =>
      DeleteCustomerRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerRequest copyWith(
          void Function(DeleteCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteCustomerRequest))
          as DeleteCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteCustomerRequest create() => DeleteCustomerRequest._();
  @$core.override
  DeleteCustomerRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteCustomerRequest> createRepeated() =>
      $pb.PbList<DeleteCustomerRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCustomerRequest>(create);
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
  factory DeleteCustomerResponse() => create();

  DeleteCustomerResponse._();

  factory DeleteCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerResponse clone() =>
      DeleteCustomerResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCustomerResponse copyWith(
          void Function(DeleteCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteCustomerResponse))
          as DeleteCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteCustomerResponse create() => DeleteCustomerResponse._();
  @$core.override
  DeleteCustomerResponse createEmptyInstance() => create();
  static $pb.PbList<DeleteCustomerResponse> createRepeated() =>
      $pb.PbList<DeleteCustomerResponse>();
  @$core.pragma('dart2js:noInline')
  static DeleteCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCustomerResponse>(create);
  static DeleteCustomerResponse? _defaultInstance;
}

class RestoreCustomerRequest extends $pb.GeneratedMessage {
  factory RestoreCustomerRequest({
    $core.String? id,
  }) {
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  RestoreCustomerRequest._();

  factory RestoreCustomerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RestoreCustomerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreCustomerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerRequest clone() =>
      RestoreCustomerRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerRequest copyWith(
          void Function(RestoreCustomerRequest) updates) =>
      super.copyWith((message) => updates(message as RestoreCustomerRequest))
          as RestoreCustomerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RestoreCustomerRequest create() => RestoreCustomerRequest._();
  @$core.override
  RestoreCustomerRequest createEmptyInstance() => create();
  static $pb.PbList<RestoreCustomerRequest> createRepeated() =>
      $pb.PbList<RestoreCustomerRequest>();
  @$core.pragma('dart2js:noInline')
  static RestoreCustomerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreCustomerRequest>(create);
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
    final result = create();
    if (customer != null) result.customer = customer;
    return result;
  }

  RestoreCustomerResponse._();

  factory RestoreCustomerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RestoreCustomerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreCustomerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'customers.v1'),
      createEmptyInstance: create)
    ..aOM<Customer>(1, _omitFieldNames ? '' : 'customer',
        subBuilder: Customer.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerResponse clone() =>
      RestoreCustomerResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreCustomerResponse copyWith(
          void Function(RestoreCustomerResponse) updates) =>
      super.copyWith((message) => updates(message as RestoreCustomerResponse))
          as RestoreCustomerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RestoreCustomerResponse create() => RestoreCustomerResponse._();
  @$core.override
  RestoreCustomerResponse createEmptyInstance() => create();
  static $pb.PbList<RestoreCustomerResponse> createRepeated() =>
      $pb.PbList<RestoreCustomerResponse>();
  @$core.pragma('dart2js:noInline')
  static RestoreCustomerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreCustomerResponse>(create);
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
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
