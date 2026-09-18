// This is a generated file - do not edit.
//
// Generated from salesorder/v1/company.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $0;

import 'common.pb.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Company:公司(租戶)主檔。
class Company extends $pb.GeneratedMessage {
  factory Company({
    $core.String? id,
    $core.String? name,
    $core.String? taxId,
    $core.String? identifier,
    $core.String? status,
    $0.Struct? publicInfo,
    $core.Iterable<$core.String>? capabilities,
    $core.String? logoUrl,
  }) {
    final result = Company._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (taxId != null) result.taxId = taxId;
    if (identifier != null) result.identifier = identifier;
    if (status != null) result.status = status;
    if (publicInfo != null) result.publicInfo = publicInfo;
    if (capabilities != null) result.capabilities.addAll(capabilities);
    if (logoUrl != null) result.logoUrl = logoUrl;
    return result;
  }

  Company._();

  factory Company.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Company()..mergeFromBuffer(data, registry);
  factory Company.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Company()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Company',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: Company.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'taxId')
    ..aOS(4, _omitFieldNames ? '' : 'identifier')
    ..aOS(5, _omitFieldNames ? '' : 'status')
    ..aOM<$0.Struct>(6, _omitFieldNames ? '' : 'publicInfo',
        subBuilder: $0.Struct.$_createMessage)
    ..pPS(7, _omitFieldNames ? '' : 'capabilities')
    ..aOS(8, _omitFieldNames ? '' : 'logoUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Company clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Company copyWith(void Function(Company) updates) =>
      super.copyWith((message) => updates(message as Company)) as Company;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Company() / Company.new instead')
  static Company create() => Company._();
  static $pb.GeneratedMessage $_createMessage() => Company._();
  @$core.override
  Company createEmptyInstance() => Company._();
  @$core.pragma('dart2js:noInline')
  static Company getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Company>(Company.$_createMessage);
  static Company? _defaultInstance;

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
  $core.String get identifier => $_getSZ(3);
  @$pb.TagNumber(4)
  set identifier($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIdentifier() => $_has(3);
  @$pb.TagNumber(4)
  void clearIdentifier() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get status => $_getSZ(4);
  @$pb.TagNumber(5)
  set status($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasStatus() => $_has(4);
  @$pb.TagNumber(5)
  void clearStatus() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Struct get publicInfo => $_getN(5);
  @$pb.TagNumber(6)
  set publicInfo($0.Struct value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPublicInfo() => $_has(5);
  @$pb.TagNumber(6)
  void clearPublicInfo() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Struct ensurePublicInfo() => $_ensure(5);

  @$pb.TagNumber(7)
  $pb.PbList<$core.String> get capabilities => $_getList(6);

  @$pb.TagNumber(8)
  $core.String get logoUrl => $_getSZ(7);
  @$pb.TagNumber(8)
  set logoUrl($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasLogoUrl() => $_has(7);
  @$pb.TagNumber(8)
  void clearLogoUrl() => $_clearField(8);
}

class ListCompaniesRequest extends $pb.GeneratedMessage {
  factory ListCompaniesRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? status,
    $core.String? keyword,
  }) {
    final result = ListCompaniesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (status != null) result.status = status;
    if (keyword != null) result.keyword = keyword;
    return result;
  }

  ListCompaniesRequest._();

  factory ListCompaniesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCompaniesRequest()..mergeFromBuffer(data, registry);
  factory ListCompaniesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCompaniesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCompaniesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListCompaniesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'status')
    ..aOS(4, _omitFieldNames ? '' : 'keyword')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCompaniesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCompaniesRequest copyWith(void Function(ListCompaniesRequest) updates) =>
      super.copyWith((message) => updates(message as ListCompaniesRequest))
          as ListCompaniesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCompaniesRequest() / ListCompaniesRequest.new instead')
  static ListCompaniesRequest create() => ListCompaniesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListCompaniesRequest._();
  @$core.override
  ListCompaniesRequest createEmptyInstance() => ListCompaniesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListCompaniesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCompaniesRequest>(
          ListCompaniesRequest.$_createMessage);
  static ListCompaniesRequest? _defaultInstance;

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
  $core.String get status => $_getSZ(2);
  @$pb.TagNumber(3)
  set status($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get keyword => $_getSZ(3);
  @$pb.TagNumber(4)
  set keyword($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasKeyword() => $_has(3);
  @$pb.TagNumber(4)
  void clearKeyword() => $_clearField(4);
}

class ListCompaniesResponse extends $pb.GeneratedMessage {
  factory ListCompaniesResponse({
    $core.Iterable<Company>? companies,
    $1.Pagination? pagination,
  }) {
    final result = ListCompaniesResponse._();
    if (companies != null) result.companies.addAll(companies);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListCompaniesResponse._();

  factory ListCompaniesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCompaniesResponse()..mergeFromBuffer(data, registry);
  factory ListCompaniesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCompaniesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCompaniesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListCompaniesResponse.$_createMessage)
    ..pPM<Company>(1, _omitFieldNames ? '' : 'companies',
        subBuilder: Company.$_createMessage)
    ..aOM<$1.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $1.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCompaniesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCompaniesResponse copyWith(
          void Function(ListCompaniesResponse) updates) =>
      super.copyWith((message) => updates(message as ListCompaniesResponse))
          as ListCompaniesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCompaniesResponse() / ListCompaniesResponse.new instead')
  static ListCompaniesResponse create() => ListCompaniesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListCompaniesResponse._();
  @$core.override
  ListCompaniesResponse createEmptyInstance() => ListCompaniesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListCompaniesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCompaniesResponse>(
          ListCompaniesResponse.$_createMessage);
  static ListCompaniesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Company> get companies => $_getList(0);

  @$pb.TagNumber(2)
  $1.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($1.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $1.Pagination ensurePagination() => $_ensure(1);
}

class GetCompanyRequest extends $pb.GeneratedMessage {
  factory GetCompanyRequest({
    $core.String? companyId,
  }) {
    final result = GetCompanyRequest._();
    if (companyId != null) result.companyId = companyId;
    return result;
  }

  GetCompanyRequest._();

  factory GetCompanyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCompanyRequest()..mergeFromBuffer(data, registry);
  factory GetCompanyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCompanyRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCompanyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetCompanyRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCompanyRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCompanyRequest copyWith(void Function(GetCompanyRequest) updates) =>
      super.copyWith((message) => updates(message as GetCompanyRequest))
          as GetCompanyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetCompanyRequest() / GetCompanyRequest.new instead')
  static GetCompanyRequest create() => GetCompanyRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetCompanyRequest._();
  @$core.override
  GetCompanyRequest createEmptyInstance() => GetCompanyRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCompanyRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetCompanyRequest>(
          GetCompanyRequest.$_createMessage);
  static GetCompanyRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);
}

class GetCompanyResponse extends $pb.GeneratedMessage {
  factory GetCompanyResponse({
    Company? company,
  }) {
    final result = GetCompanyResponse._();
    if (company != null) result.company = company;
    return result;
  }

  GetCompanyResponse._();

  factory GetCompanyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCompanyResponse()..mergeFromBuffer(data, registry);
  factory GetCompanyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCompanyResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCompanyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetCompanyResponse.$_createMessage)
    ..aOM<Company>(1, _omitFieldNames ? '' : 'company',
        subBuilder: Company.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCompanyResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCompanyResponse copyWith(void Function(GetCompanyResponse) updates) =>
      super.copyWith((message) => updates(message as GetCompanyResponse))
          as GetCompanyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetCompanyResponse() / GetCompanyResponse.new instead')
  static GetCompanyResponse create() => GetCompanyResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetCompanyResponse._();
  @$core.override
  GetCompanyResponse createEmptyInstance() => GetCompanyResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCompanyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCompanyResponse>(
          GetCompanyResponse.$_createMessage);
  static GetCompanyResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Company get company => $_getN(0);
  @$pb.TagNumber(1)
  set company(Company value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCompany() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompany() => $_clearField(1);
  @$pb.TagNumber(1)
  Company ensureCompany() => $_ensure(0);
}

class CreateCompanyRequest extends $pb.GeneratedMessage {
  factory CreateCompanyRequest({
    $core.String? name,
    $core.String? taxId,
    $core.String? identifier,
    $core.String? status,
  }) {
    final result = CreateCompanyRequest._();
    if (name != null) result.name = name;
    if (taxId != null) result.taxId = taxId;
    if (identifier != null) result.identifier = identifier;
    if (status != null) result.status = status;
    return result;
  }

  CreateCompanyRequest._();

  factory CreateCompanyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCompanyRequest()..mergeFromBuffer(data, registry);
  factory CreateCompanyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCompanyRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCompanyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateCompanyRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'taxId')
    ..aOS(3, _omitFieldNames ? '' : 'identifier')
    ..aOS(4, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCompanyRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCompanyRequest copyWith(void Function(CreateCompanyRequest) updates) =>
      super.copyWith((message) => updates(message as CreateCompanyRequest))
          as CreateCompanyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCompanyRequest() / CreateCompanyRequest.new instead')
  static CreateCompanyRequest create() => CreateCompanyRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateCompanyRequest._();
  @$core.override
  CreateCompanyRequest createEmptyInstance() => CreateCompanyRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCompanyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCompanyRequest>(
          CreateCompanyRequest.$_createMessage);
  static CreateCompanyRequest? _defaultInstance;

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
  $core.String get identifier => $_getSZ(2);
  @$pb.TagNumber(3)
  set identifier($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasIdentifier() => $_has(2);
  @$pb.TagNumber(3)
  void clearIdentifier() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get status => $_getSZ(3);
  @$pb.TagNumber(4)
  set status($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);
}

class CreateCompanyResponse extends $pb.GeneratedMessage {
  factory CreateCompanyResponse({
    Company? company,
  }) {
    final result = CreateCompanyResponse._();
    if (company != null) result.company = company;
    return result;
  }

  CreateCompanyResponse._();

  factory CreateCompanyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCompanyResponse()..mergeFromBuffer(data, registry);
  factory CreateCompanyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCompanyResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCompanyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateCompanyResponse.$_createMessage)
    ..aOM<Company>(1, _omitFieldNames ? '' : 'company',
        subBuilder: Company.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCompanyResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCompanyResponse copyWith(
          void Function(CreateCompanyResponse) updates) =>
      super.copyWith((message) => updates(message as CreateCompanyResponse))
          as CreateCompanyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCompanyResponse() / CreateCompanyResponse.new instead')
  static CreateCompanyResponse create() => CreateCompanyResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateCompanyResponse._();
  @$core.override
  CreateCompanyResponse createEmptyInstance() => CreateCompanyResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCompanyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCompanyResponse>(
          CreateCompanyResponse.$_createMessage);
  static CreateCompanyResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Company get company => $_getN(0);
  @$pb.TagNumber(1)
  set company(Company value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCompany() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompany() => $_clearField(1);
  @$pb.TagNumber(1)
  Company ensureCompany() => $_ensure(0);
}

class UpdateCompanyRequest extends $pb.GeneratedMessage {
  factory UpdateCompanyRequest({
    $core.String? companyId,
    $core.String? name,
    $core.String? taxId,
    $core.String? status,
    $core.String? identifier,
  }) {
    final result = UpdateCompanyRequest._();
    if (companyId != null) result.companyId = companyId;
    if (name != null) result.name = name;
    if (taxId != null) result.taxId = taxId;
    if (status != null) result.status = status;
    if (identifier != null) result.identifier = identifier;
    return result;
  }

  UpdateCompanyRequest._();

  factory UpdateCompanyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCompanyRequest()..mergeFromBuffer(data, registry);
  factory UpdateCompanyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCompanyRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCompanyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateCompanyRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'taxId')
    ..aOS(4, _omitFieldNames ? '' : 'status')
    ..aOS(5, _omitFieldNames ? '' : 'identifier')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCompanyRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCompanyRequest copyWith(void Function(UpdateCompanyRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateCompanyRequest))
          as UpdateCompanyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCompanyRequest() / UpdateCompanyRequest.new instead')
  static UpdateCompanyRequest create() => UpdateCompanyRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateCompanyRequest._();
  @$core.override
  UpdateCompanyRequest createEmptyInstance() => UpdateCompanyRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateCompanyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCompanyRequest>(
          UpdateCompanyRequest.$_createMessage);
  static UpdateCompanyRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

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
  $core.String get status => $_getSZ(3);
  @$pb.TagNumber(4)
  set status($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  /// identifier 建立後不可修改;於請求中出現即回 invalid_argument。
  @$pb.TagNumber(5)
  $core.String get identifier => $_getSZ(4);
  @$pb.TagNumber(5)
  set identifier($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasIdentifier() => $_has(4);
  @$pb.TagNumber(5)
  void clearIdentifier() => $_clearField(5);
}

class UpdateCompanyResponse extends $pb.GeneratedMessage {
  factory UpdateCompanyResponse({
    Company? company,
  }) {
    final result = UpdateCompanyResponse._();
    if (company != null) result.company = company;
    return result;
  }

  UpdateCompanyResponse._();

  factory UpdateCompanyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCompanyResponse()..mergeFromBuffer(data, registry);
  factory UpdateCompanyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCompanyResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCompanyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateCompanyResponse.$_createMessage)
    ..aOM<Company>(1, _omitFieldNames ? '' : 'company',
        subBuilder: Company.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCompanyResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCompanyResponse copyWith(
          void Function(UpdateCompanyResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateCompanyResponse))
          as UpdateCompanyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCompanyResponse() / UpdateCompanyResponse.new instead')
  static UpdateCompanyResponse create() => UpdateCompanyResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateCompanyResponse._();
  @$core.override
  UpdateCompanyResponse createEmptyInstance() => UpdateCompanyResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateCompanyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCompanyResponse>(
          UpdateCompanyResponse.$_createMessage);
  static UpdateCompanyResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Company get company => $_getN(0);
  @$pb.TagNumber(1)
  set company(Company value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCompany() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompany() => $_clearField(1);
  @$pb.TagNumber(1)
  Company ensureCompany() => $_ensure(0);
}

class DeleteCompanyRequest extends $pb.GeneratedMessage {
  factory DeleteCompanyRequest({
    $core.String? companyId,
  }) {
    final result = DeleteCompanyRequest._();
    if (companyId != null) result.companyId = companyId;
    return result;
  }

  DeleteCompanyRequest._();

  factory DeleteCompanyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCompanyRequest()..mergeFromBuffer(data, registry);
  factory DeleteCompanyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCompanyRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCompanyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteCompanyRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCompanyRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCompanyRequest copyWith(void Function(DeleteCompanyRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteCompanyRequest))
          as DeleteCompanyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCompanyRequest() / DeleteCompanyRequest.new instead')
  static DeleteCompanyRequest create() => DeleteCompanyRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteCompanyRequest._();
  @$core.override
  DeleteCompanyRequest createEmptyInstance() => DeleteCompanyRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteCompanyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCompanyRequest>(
          DeleteCompanyRequest.$_createMessage);
  static DeleteCompanyRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);
}

class DeleteCompanyResponse extends $pb.GeneratedMessage {
  factory DeleteCompanyResponse() => DeleteCompanyResponse._();

  DeleteCompanyResponse._();

  factory DeleteCompanyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCompanyResponse()..mergeFromBuffer(data, registry);
  factory DeleteCompanyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCompanyResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCompanyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteCompanyResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCompanyResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCompanyResponse copyWith(
          void Function(DeleteCompanyResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteCompanyResponse))
          as DeleteCompanyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCompanyResponse() / DeleteCompanyResponse.new instead')
  static DeleteCompanyResponse create() => DeleteCompanyResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteCompanyResponse._();
  @$core.override
  DeleteCompanyResponse createEmptyInstance() => DeleteCompanyResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteCompanyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCompanyResponse>(
          DeleteCompanyResponse.$_createMessage);
  static DeleteCompanyResponse? _defaultInstance;
}

/// Department:部門主檔(屬於單一公司)。
class Department extends $pb.GeneratedMessage {
  factory Department({
    $core.String? id,
    $core.String? companyId,
    $core.String? companyName,
    $core.String? name,
  }) {
    final result = Department._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (companyName != null) result.companyName = companyName;
    if (name != null) result.name = name;
    return result;
  }

  Department._();

  factory Department.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Department()..mergeFromBuffer(data, registry);
  factory Department.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Department()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Department',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: Department.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'companyName')
    ..aOS(4, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Department clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Department copyWith(void Function(Department) updates) =>
      super.copyWith((message) => updates(message as Department)) as Department;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Department() / Department.new instead')
  static Department create() => Department._();
  static $pb.GeneratedMessage $_createMessage() => Department._();
  @$core.override
  Department createEmptyInstance() => Department._();
  @$core.pragma('dart2js:noInline')
  static Department getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Department>(Department.$_createMessage);
  static Department? _defaultInstance;

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
  $core.String get companyName => $_getSZ(2);
  @$pb.TagNumber(3)
  set companyName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCompanyName() => $_has(2);
  @$pb.TagNumber(3)
  void clearCompanyName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get name => $_getSZ(3);
  @$pb.TagNumber(4)
  set name($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasName() => $_has(3);
  @$pb.TagNumber(4)
  void clearName() => $_clearField(4);
}

class ListDepartmentsRequest extends $pb.GeneratedMessage {
  factory ListDepartmentsRequest({
    $core.int? page,
    $core.int? pageSize,
    $core.String? companyId,
  }) {
    final result = ListDepartmentsRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    if (companyId != null) result.companyId = companyId;
    return result;
  }

  ListDepartmentsRequest._();

  factory ListDepartmentsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDepartmentsRequest()..mergeFromBuffer(data, registry);
  factory ListDepartmentsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDepartmentsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDepartmentsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListDepartmentsRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'companyId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDepartmentsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDepartmentsRequest copyWith(
          void Function(ListDepartmentsRequest) updates) =>
      super.copyWith((message) => updates(message as ListDepartmentsRequest))
          as ListDepartmentsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListDepartmentsRequest() / ListDepartmentsRequest.new instead')
  static ListDepartmentsRequest create() => ListDepartmentsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListDepartmentsRequest._();
  @$core.override
  ListDepartmentsRequest createEmptyInstance() => ListDepartmentsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListDepartmentsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDepartmentsRequest>(
          ListDepartmentsRequest.$_createMessage);
  static ListDepartmentsRequest? _defaultInstance;

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
  $core.String get companyId => $_getSZ(2);
  @$pb.TagNumber(3)
  set companyId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCompanyId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCompanyId() => $_clearField(3);
}

class ListDepartmentsResponse extends $pb.GeneratedMessage {
  factory ListDepartmentsResponse({
    $core.Iterable<Department>? departments,
    $1.Pagination? pagination,
  }) {
    final result = ListDepartmentsResponse._();
    if (departments != null) result.departments.addAll(departments);
    if (pagination != null) result.pagination = pagination;
    return result;
  }

  ListDepartmentsResponse._();

  factory ListDepartmentsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDepartmentsResponse()..mergeFromBuffer(data, registry);
  factory ListDepartmentsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDepartmentsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDepartmentsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListDepartmentsResponse.$_createMessage)
    ..pPM<Department>(1, _omitFieldNames ? '' : 'departments',
        subBuilder: Department.$_createMessage)
    ..aOM<$1.Pagination>(2, _omitFieldNames ? '' : 'pagination',
        subBuilder: $1.Pagination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDepartmentsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDepartmentsResponse copyWith(
          void Function(ListDepartmentsResponse) updates) =>
      super.copyWith((message) => updates(message as ListDepartmentsResponse))
          as ListDepartmentsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListDepartmentsResponse() / ListDepartmentsResponse.new instead')
  static ListDepartmentsResponse create() => ListDepartmentsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListDepartmentsResponse._();
  @$core.override
  ListDepartmentsResponse createEmptyInstance() => ListDepartmentsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListDepartmentsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDepartmentsResponse>(
          ListDepartmentsResponse.$_createMessage);
  static ListDepartmentsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Department> get departments => $_getList(0);

  @$pb.TagNumber(2)
  $1.Pagination get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($1.Pagination value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => $_clearField(2);
  @$pb.TagNumber(2)
  $1.Pagination ensurePagination() => $_ensure(1);
}

class GetDepartmentRequest extends $pb.GeneratedMessage {
  factory GetDepartmentRequest({
    $core.String? departmentId,
  }) {
    final result = GetDepartmentRequest._();
    if (departmentId != null) result.departmentId = departmentId;
    return result;
  }

  GetDepartmentRequest._();

  factory GetDepartmentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDepartmentRequest()..mergeFromBuffer(data, registry);
  factory GetDepartmentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDepartmentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDepartmentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetDepartmentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'departmentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDepartmentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDepartmentRequest copyWith(void Function(GetDepartmentRequest) updates) =>
      super.copyWith((message) => updates(message as GetDepartmentRequest))
          as GetDepartmentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetDepartmentRequest() / GetDepartmentRequest.new instead')
  static GetDepartmentRequest create() => GetDepartmentRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetDepartmentRequest._();
  @$core.override
  GetDepartmentRequest createEmptyInstance() => GetDepartmentRequest._();
  @$core.pragma('dart2js:noInline')
  static GetDepartmentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetDepartmentRequest>(
          GetDepartmentRequest.$_createMessage);
  static GetDepartmentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get departmentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set departmentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDepartmentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDepartmentId() => $_clearField(1);
}

class GetDepartmentResponse extends $pb.GeneratedMessage {
  factory GetDepartmentResponse({
    Department? department,
  }) {
    final result = GetDepartmentResponse._();
    if (department != null) result.department = department;
    return result;
  }

  GetDepartmentResponse._();

  factory GetDepartmentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDepartmentResponse()..mergeFromBuffer(data, registry);
  factory GetDepartmentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDepartmentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDepartmentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: GetDepartmentResponse.$_createMessage)
    ..aOM<Department>(1, _omitFieldNames ? '' : 'department',
        subBuilder: Department.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDepartmentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDepartmentResponse copyWith(
          void Function(GetDepartmentResponse) updates) =>
      super.copyWith((message) => updates(message as GetDepartmentResponse))
          as GetDepartmentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetDepartmentResponse() / GetDepartmentResponse.new instead')
  static GetDepartmentResponse create() => GetDepartmentResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetDepartmentResponse._();
  @$core.override
  GetDepartmentResponse createEmptyInstance() => GetDepartmentResponse._();
  @$core.pragma('dart2js:noInline')
  static GetDepartmentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetDepartmentResponse>(
          GetDepartmentResponse.$_createMessage);
  static GetDepartmentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Department get department => $_getN(0);
  @$pb.TagNumber(1)
  set department(Department value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDepartment() => $_has(0);
  @$pb.TagNumber(1)
  void clearDepartment() => $_clearField(1);
  @$pb.TagNumber(1)
  Department ensureDepartment() => $_ensure(0);
}

class CreateDepartmentRequest extends $pb.GeneratedMessage {
  factory CreateDepartmentRequest({
    $core.String? companyId,
    $core.String? name,
  }) {
    final result = CreateDepartmentRequest._();
    if (companyId != null) result.companyId = companyId;
    if (name != null) result.name = name;
    return result;
  }

  CreateDepartmentRequest._();

  factory CreateDepartmentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDepartmentRequest()..mergeFromBuffer(data, registry);
  factory CreateDepartmentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDepartmentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDepartmentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateDepartmentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'companyId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDepartmentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDepartmentRequest copyWith(
          void Function(CreateDepartmentRequest) updates) =>
      super.copyWith((message) => updates(message as CreateDepartmentRequest))
          as CreateDepartmentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDepartmentRequest() / CreateDepartmentRequest.new instead')
  static CreateDepartmentRequest create() => CreateDepartmentRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateDepartmentRequest._();
  @$core.override
  CreateDepartmentRequest createEmptyInstance() => CreateDepartmentRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateDepartmentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDepartmentRequest>(
          CreateDepartmentRequest.$_createMessage);
  static CreateDepartmentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get companyId => $_getSZ(0);
  @$pb.TagNumber(1)
  set companyId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCompanyId() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompanyId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

class CreateDepartmentResponse extends $pb.GeneratedMessage {
  factory CreateDepartmentResponse({
    Department? department,
  }) {
    final result = CreateDepartmentResponse._();
    if (department != null) result.department = department;
    return result;
  }

  CreateDepartmentResponse._();

  factory CreateDepartmentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDepartmentResponse()..mergeFromBuffer(data, registry);
  factory CreateDepartmentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDepartmentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDepartmentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateDepartmentResponse.$_createMessage)
    ..aOM<Department>(1, _omitFieldNames ? '' : 'department',
        subBuilder: Department.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDepartmentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDepartmentResponse copyWith(
          void Function(CreateDepartmentResponse) updates) =>
      super.copyWith((message) => updates(message as CreateDepartmentResponse))
          as CreateDepartmentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDepartmentResponse() / CreateDepartmentResponse.new instead')
  static CreateDepartmentResponse create() => CreateDepartmentResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateDepartmentResponse._();
  @$core.override
  CreateDepartmentResponse createEmptyInstance() =>
      CreateDepartmentResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateDepartmentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDepartmentResponse>(
          CreateDepartmentResponse.$_createMessage);
  static CreateDepartmentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Department get department => $_getN(0);
  @$pb.TagNumber(1)
  set department(Department value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDepartment() => $_has(0);
  @$pb.TagNumber(1)
  void clearDepartment() => $_clearField(1);
  @$pb.TagNumber(1)
  Department ensureDepartment() => $_ensure(0);
}

class UpdateDepartmentRequest extends $pb.GeneratedMessage {
  factory UpdateDepartmentRequest({
    $core.String? departmentId,
    $core.String? name,
  }) {
    final result = UpdateDepartmentRequest._();
    if (departmentId != null) result.departmentId = departmentId;
    if (name != null) result.name = name;
    return result;
  }

  UpdateDepartmentRequest._();

  factory UpdateDepartmentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDepartmentRequest()..mergeFromBuffer(data, registry);
  factory UpdateDepartmentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDepartmentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateDepartmentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateDepartmentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'departmentId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDepartmentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDepartmentRequest copyWith(
          void Function(UpdateDepartmentRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateDepartmentRequest))
          as UpdateDepartmentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateDepartmentRequest() / UpdateDepartmentRequest.new instead')
  static UpdateDepartmentRequest create() => UpdateDepartmentRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateDepartmentRequest._();
  @$core.override
  UpdateDepartmentRequest createEmptyInstance() => UpdateDepartmentRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateDepartmentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateDepartmentRequest>(
          UpdateDepartmentRequest.$_createMessage);
  static UpdateDepartmentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get departmentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set departmentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDepartmentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDepartmentId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

class UpdateDepartmentResponse extends $pb.GeneratedMessage {
  factory UpdateDepartmentResponse({
    Department? department,
  }) {
    final result = UpdateDepartmentResponse._();
    if (department != null) result.department = department;
    return result;
  }

  UpdateDepartmentResponse._();

  factory UpdateDepartmentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDepartmentResponse()..mergeFromBuffer(data, registry);
  factory UpdateDepartmentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDepartmentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateDepartmentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: UpdateDepartmentResponse.$_createMessage)
    ..aOM<Department>(1, _omitFieldNames ? '' : 'department',
        subBuilder: Department.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDepartmentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDepartmentResponse copyWith(
          void Function(UpdateDepartmentResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateDepartmentResponse))
          as UpdateDepartmentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateDepartmentResponse() / UpdateDepartmentResponse.new instead')
  static UpdateDepartmentResponse create() => UpdateDepartmentResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateDepartmentResponse._();
  @$core.override
  UpdateDepartmentResponse createEmptyInstance() =>
      UpdateDepartmentResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateDepartmentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateDepartmentResponse>(
          UpdateDepartmentResponse.$_createMessage);
  static UpdateDepartmentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Department get department => $_getN(0);
  @$pb.TagNumber(1)
  set department(Department value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDepartment() => $_has(0);
  @$pb.TagNumber(1)
  void clearDepartment() => $_clearField(1);
  @$pb.TagNumber(1)
  Department ensureDepartment() => $_ensure(0);
}

class DeleteDepartmentRequest extends $pb.GeneratedMessage {
  factory DeleteDepartmentRequest({
    $core.String? departmentId,
  }) {
    final result = DeleteDepartmentRequest._();
    if (departmentId != null) result.departmentId = departmentId;
    return result;
  }

  DeleteDepartmentRequest._();

  factory DeleteDepartmentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDepartmentRequest()..mergeFromBuffer(data, registry);
  factory DeleteDepartmentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDepartmentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteDepartmentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteDepartmentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'departmentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDepartmentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDepartmentRequest copyWith(
          void Function(DeleteDepartmentRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteDepartmentRequest))
          as DeleteDepartmentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteDepartmentRequest() / DeleteDepartmentRequest.new instead')
  static DeleteDepartmentRequest create() => DeleteDepartmentRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteDepartmentRequest._();
  @$core.override
  DeleteDepartmentRequest createEmptyInstance() => DeleteDepartmentRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteDepartmentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteDepartmentRequest>(
          DeleteDepartmentRequest.$_createMessage);
  static DeleteDepartmentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get departmentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set departmentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDepartmentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDepartmentId() => $_clearField(1);
}

class DeleteDepartmentResponse extends $pb.GeneratedMessage {
  factory DeleteDepartmentResponse() => DeleteDepartmentResponse._();

  DeleteDepartmentResponse._();

  factory DeleteDepartmentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDepartmentResponse()..mergeFromBuffer(data, registry);
  factory DeleteDepartmentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDepartmentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteDepartmentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: DeleteDepartmentResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDepartmentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDepartmentResponse copyWith(
          void Function(DeleteDepartmentResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteDepartmentResponse))
          as DeleteDepartmentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteDepartmentResponse() / DeleteDepartmentResponse.new instead')
  static DeleteDepartmentResponse create() => DeleteDepartmentResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteDepartmentResponse._();
  @$core.override
  DeleteDepartmentResponse createEmptyInstance() =>
      DeleteDepartmentResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteDepartmentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteDepartmentResponse>(
          DeleteDepartmentResponse.$_createMessage);
  static DeleteDepartmentResponse? _defaultInstance;
}

/// CompanyService:公司主檔 CRUD。
class CompanyServiceApi {
  final $pb.RpcClient _client;

  CompanyServiceApi(this._client);

  /// ListCompanies:分頁列出公司,可依 status / keyword(name、identifier 模糊)篩選。
  $async.Future<ListCompaniesResponse> listCompanies(
          $pb.ClientContext? ctx, ListCompaniesRequest request) =>
      _client.invoke<ListCompaniesResponse>(ctx, 'CompanyService',
          'ListCompanies', request, ListCompaniesResponse());

  /// GetCompany:取得單一公司。
  $async.Future<GetCompanyResponse> getCompany(
          $pb.ClientContext? ctx, GetCompanyRequest request) =>
      _client.invoke<GetCompanyResponse>(
          ctx, 'CompanyService', 'GetCompany', request, GetCompanyResponse());

  /// CreateCompany:建立公司。
  $async.Future<CreateCompanyResponse> createCompany(
          $pb.ClientContext? ctx, CreateCompanyRequest request) =>
      _client.invoke<CreateCompanyResponse>(ctx, 'CompanyService',
          'CreateCompany', request, CreateCompanyResponse());

  /// UpdateCompany:更新公司(name、tax_id、status;identifier 建立後不可修改)。
  $async.Future<UpdateCompanyResponse> updateCompany(
          $pb.ClientContext? ctx, UpdateCompanyRequest request) =>
      _client.invoke<UpdateCompanyResponse>(ctx, 'CompanyService',
          'UpdateCompany', request, UpdateCompanyResponse());

  /// DeleteCompany:刪除公司。
  $async.Future<DeleteCompanyResponse> deleteCompany(
          $pb.ClientContext? ctx, DeleteCompanyRequest request) =>
      _client.invoke<DeleteCompanyResponse>(ctx, 'CompanyService',
          'DeleteCompany', request, DeleteCompanyResponse());
}

/// DepartmentService:部門主檔 CRUD。
class DepartmentServiceApi {
  final $pb.RpcClient _client;

  DepartmentServiceApi(this._client);

  /// ListDepartments:分頁列出部門,可依 company_id 篩選。
  $async.Future<ListDepartmentsResponse> listDepartments(
          $pb.ClientContext? ctx, ListDepartmentsRequest request) =>
      _client.invoke<ListDepartmentsResponse>(ctx, 'DepartmentService',
          'ListDepartments', request, ListDepartmentsResponse());

  /// GetDepartment:取得單一部門。
  $async.Future<GetDepartmentResponse> getDepartment(
          $pb.ClientContext? ctx, GetDepartmentRequest request) =>
      _client.invoke<GetDepartmentResponse>(ctx, 'DepartmentService',
          'GetDepartment', request, GetDepartmentResponse());

  /// CreateDepartment:建立部門(需指定所屬公司)。
  $async.Future<CreateDepartmentResponse> createDepartment(
          $pb.ClientContext? ctx, CreateDepartmentRequest request) =>
      _client.invoke<CreateDepartmentResponse>(ctx, 'DepartmentService',
          'CreateDepartment', request, CreateDepartmentResponse());

  /// UpdateDepartment:更新部門名稱。
  $async.Future<UpdateDepartmentResponse> updateDepartment(
          $pb.ClientContext? ctx, UpdateDepartmentRequest request) =>
      _client.invoke<UpdateDepartmentResponse>(ctx, 'DepartmentService',
          'UpdateDepartment', request, UpdateDepartmentResponse());

  /// DeleteDepartment:刪除部門。
  $async.Future<DeleteDepartmentResponse> deleteDepartment(
          $pb.ClientContext? ctx, DeleteDepartmentRequest request) =>
      _client.invoke<DeleteDepartmentResponse>(ctx, 'DepartmentService',
          'DeleteDepartment', request, DeleteDepartmentResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
