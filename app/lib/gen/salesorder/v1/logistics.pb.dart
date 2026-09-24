// This is a generated file - do not edit.
//
// Generated from salesorder/v1/logistics.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// LogisticsDriver:司機單筆。
class LogisticsDriver extends $pb.GeneratedMessage {
  factory LogisticsDriver({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? userId,
    $core.String? name,
    $core.String? phone,
    $core.String? currentStatus,
  }) {
    final result = LogisticsDriver._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (userId != null) result.userId = userId;
    if (name != null) result.name = name;
    if (phone != null) result.phone = phone;
    if (currentStatus != null) result.currentStatus = currentStatus;
    return result;
  }

  LogisticsDriver._();

  factory LogisticsDriver.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogisticsDriver()..mergeFromBuffer(data, registry);
  factory LogisticsDriver.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogisticsDriver()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogisticsDriver',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: LogisticsDriver.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'userId')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'phone')
    ..aOS(7, _omitFieldNames ? '' : 'currentStatus')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogisticsDriver clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogisticsDriver copyWith(void Function(LogisticsDriver) updates) =>
      super.copyWith((message) => updates(message as LogisticsDriver))
          as LogisticsDriver;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogisticsDriver() / LogisticsDriver.new instead')
  static LogisticsDriver create() => LogisticsDriver._();
  static $pb.GeneratedMessage $_createMessage() => LogisticsDriver._();
  @$core.override
  LogisticsDriver createEmptyInstance() => LogisticsDriver._();
  @$core.pragma('dart2js:noInline')
  static LogisticsDriver getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogisticsDriver>(
          LogisticsDriver.$_createMessage);
  static LogisticsDriver? _defaultInstance;

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
  $core.String get userId => $_getSZ(3);
  @$pb.TagNumber(4)
  set userId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasUserId() => $_has(3);
  @$pb.TagNumber(4)
  void clearUserId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get name => $_getSZ(4);
  @$pb.TagNumber(5)
  set name($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasName() => $_has(4);
  @$pb.TagNumber(5)
  void clearName() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get phone => $_getSZ(5);
  @$pb.TagNumber(6)
  set phone($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPhone() => $_has(5);
  @$pb.TagNumber(6)
  void clearPhone() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get currentStatus => $_getSZ(6);
  @$pb.TagNumber(7)
  set currentStatus($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCurrentStatus() => $_has(6);
  @$pb.TagNumber(7)
  void clearCurrentStatus() => $_clearField(7);
}

/// CreateDriverRequest:建司機請求。
class CreateDriverRequest extends $pb.GeneratedMessage {
  factory CreateDriverRequest({
    $core.String? userId,
    $core.String? name,
    $core.String? phone,
  }) {
    final result = CreateDriverRequest._();
    if (userId != null) result.userId = userId;
    if (name != null) result.name = name;
    if (phone != null) result.phone = phone;
    return result;
  }

  CreateDriverRequest._();

  factory CreateDriverRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDriverRequest()..mergeFromBuffer(data, registry);
  factory CreateDriverRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDriverRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDriverRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateDriverRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'phone')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDriverRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDriverRequest copyWith(void Function(CreateDriverRequest) updates) =>
      super.copyWith((message) => updates(message as CreateDriverRequest))
          as CreateDriverRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CreateDriverRequest() / CreateDriverRequest.new instead')
  static CreateDriverRequest create() => CreateDriverRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateDriverRequest._();
  @$core.override
  CreateDriverRequest createEmptyInstance() => CreateDriverRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateDriverRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDriverRequest>(
          CreateDriverRequest.$_createMessage);
  static CreateDriverRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get phone => $_getSZ(2);
  @$pb.TagNumber(3)
  set phone($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPhone() => $_has(2);
  @$pb.TagNumber(3)
  void clearPhone() => $_clearField(3);
}

/// CreateDriverResponse:建司機結果。
class CreateDriverResponse extends $pb.GeneratedMessage {
  factory CreateDriverResponse({
    LogisticsDriver? driver,
  }) {
    final result = CreateDriverResponse._();
    if (driver != null) result.driver = driver;
    return result;
  }

  CreateDriverResponse._();

  factory CreateDriverResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDriverResponse()..mergeFromBuffer(data, registry);
  factory CreateDriverResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDriverResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDriverResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateDriverResponse.$_createMessage)
    ..aOM<LogisticsDriver>(1, _omitFieldNames ? '' : 'driver',
        subBuilder: LogisticsDriver.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDriverResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDriverResponse copyWith(void Function(CreateDriverResponse) updates) =>
      super.copyWith((message) => updates(message as CreateDriverResponse))
          as CreateDriverResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDriverResponse() / CreateDriverResponse.new instead')
  static CreateDriverResponse create() => CreateDriverResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateDriverResponse._();
  @$core.override
  CreateDriverResponse createEmptyInstance() => CreateDriverResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateDriverResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDriverResponse>(
          CreateDriverResponse.$_createMessage);
  static CreateDriverResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LogisticsDriver get driver => $_getN(0);
  @$pb.TagNumber(1)
  set driver(LogisticsDriver value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDriver() => $_has(0);
  @$pb.TagNumber(1)
  void clearDriver() => $_clearField(1);
  @$pb.TagNumber(1)
  LogisticsDriver ensureDriver() => $_ensure(0);
}

/// Vehicle:車輛單筆。
class Vehicle extends $pb.GeneratedMessage {
  factory Vehicle({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? plateNo,
    $core.String? vehicleType,
    $core.String? status,
  }) {
    final result = Vehicle._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (plateNo != null) result.plateNo = plateNo;
    if (vehicleType != null) result.vehicleType = vehicleType;
    if (status != null) result.status = status;
    return result;
  }

  Vehicle._();

  factory Vehicle.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Vehicle()..mergeFromBuffer(data, registry);
  factory Vehicle.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Vehicle()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Vehicle',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: Vehicle.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'plateNo')
    ..aOS(5, _omitFieldNames ? '' : 'vehicleType')
    ..aOS(6, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Vehicle clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Vehicle copyWith(void Function(Vehicle) updates) =>
      super.copyWith((message) => updates(message as Vehicle)) as Vehicle;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Vehicle() / Vehicle.new instead')
  static Vehicle create() => Vehicle._();
  static $pb.GeneratedMessage $_createMessage() => Vehicle._();
  @$core.override
  Vehicle createEmptyInstance() => Vehicle._();
  @$core.pragma('dart2js:noInline')
  static Vehicle getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Vehicle>(Vehicle.$_createMessage);
  static Vehicle? _defaultInstance;

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
  $core.String get plateNo => $_getSZ(3);
  @$pb.TagNumber(4)
  set plateNo($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPlateNo() => $_has(3);
  @$pb.TagNumber(4)
  void clearPlateNo() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get vehicleType => $_getSZ(4);
  @$pb.TagNumber(5)
  set vehicleType($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasVehicleType() => $_has(4);
  @$pb.TagNumber(5)
  void clearVehicleType() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get status => $_getSZ(5);
  @$pb.TagNumber(6)
  set status($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);
}

/// CreateVehicleRequest:建車輛請求。
class CreateVehicleRequest extends $pb.GeneratedMessage {
  factory CreateVehicleRequest({
    $core.String? plateNo,
    $core.String? vehicleType,
  }) {
    final result = CreateVehicleRequest._();
    if (plateNo != null) result.plateNo = plateNo;
    if (vehicleType != null) result.vehicleType = vehicleType;
    return result;
  }

  CreateVehicleRequest._();

  factory CreateVehicleRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateVehicleRequest()..mergeFromBuffer(data, registry);
  factory CreateVehicleRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateVehicleRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateVehicleRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateVehicleRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'plateNo')
    ..aOS(2, _omitFieldNames ? '' : 'vehicleType')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateVehicleRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateVehicleRequest copyWith(void Function(CreateVehicleRequest) updates) =>
      super.copyWith((message) => updates(message as CreateVehicleRequest))
          as CreateVehicleRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateVehicleRequest() / CreateVehicleRequest.new instead')
  static CreateVehicleRequest create() => CreateVehicleRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateVehicleRequest._();
  @$core.override
  CreateVehicleRequest createEmptyInstance() => CreateVehicleRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateVehicleRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateVehicleRequest>(
          CreateVehicleRequest.$_createMessage);
  static CreateVehicleRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get plateNo => $_getSZ(0);
  @$pb.TagNumber(1)
  set plateNo($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlateNo() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlateNo() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get vehicleType => $_getSZ(1);
  @$pb.TagNumber(2)
  set vehicleType($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVehicleType() => $_has(1);
  @$pb.TagNumber(2)
  void clearVehicleType() => $_clearField(2);
}

/// CreateVehicleResponse:建車輛結果。
class CreateVehicleResponse extends $pb.GeneratedMessage {
  factory CreateVehicleResponse({
    Vehicle? vehicle,
  }) {
    final result = CreateVehicleResponse._();
    if (vehicle != null) result.vehicle = vehicle;
    return result;
  }

  CreateVehicleResponse._();

  factory CreateVehicleResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateVehicleResponse()..mergeFromBuffer(data, registry);
  factory CreateVehicleResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateVehicleResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateVehicleResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CreateVehicleResponse.$_createMessage)
    ..aOM<Vehicle>(1, _omitFieldNames ? '' : 'vehicle',
        subBuilder: Vehicle.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateVehicleResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateVehicleResponse copyWith(
          void Function(CreateVehicleResponse) updates) =>
      super.copyWith((message) => updates(message as CreateVehicleResponse))
          as CreateVehicleResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateVehicleResponse() / CreateVehicleResponse.new instead')
  static CreateVehicleResponse create() => CreateVehicleResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateVehicleResponse._();
  @$core.override
  CreateVehicleResponse createEmptyInstance() => CreateVehicleResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateVehicleResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateVehicleResponse>(
          CreateVehicleResponse.$_createMessage);
  static CreateVehicleResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Vehicle get vehicle => $_getN(0);
  @$pb.TagNumber(1)
  set vehicle(Vehicle value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasVehicle() => $_has(0);
  @$pb.TagNumber(1)
  void clearVehicle() => $_clearField(1);
  @$pb.TagNumber(1)
  Vehicle ensureVehicle() => $_ensure(0);
}

/// LogisticsDelivery:配送執行單元(route 粒度)。
class LogisticsDelivery extends $pb.GeneratedMessage {
  factory LogisticsDelivery({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? routeId,
    $core.String? driverId,
    $core.String? vehicleId,
    $core.String? assignedBy,
    $core.String? status,
    $core.String? version,
    $core.String? createdAt,
    $core.String? updatedAt,
    $core.String? startedAt,
    $core.String? completedAt,
  }) {
    final result = LogisticsDelivery._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (routeId != null) result.routeId = routeId;
    if (driverId != null) result.driverId = driverId;
    if (vehicleId != null) result.vehicleId = vehicleId;
    if (assignedBy != null) result.assignedBy = assignedBy;
    if (status != null) result.status = status;
    if (version != null) result.version = version;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (startedAt != null) result.startedAt = startedAt;
    if (completedAt != null) result.completedAt = completedAt;
    return result;
  }

  LogisticsDelivery._();

  factory LogisticsDelivery.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogisticsDelivery()..mergeFromBuffer(data, registry);
  factory LogisticsDelivery.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogisticsDelivery()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogisticsDelivery',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: LogisticsDelivery.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'routeId')
    ..aOS(5, _omitFieldNames ? '' : 'driverId')
    ..aOS(6, _omitFieldNames ? '' : 'vehicleId')
    ..aOS(7, _omitFieldNames ? '' : 'assignedBy')
    ..aOS(8, _omitFieldNames ? '' : 'status')
    ..aOS(9, _omitFieldNames ? '' : 'version')
    ..aOS(10, _omitFieldNames ? '' : 'createdAt')
    ..aOS(11, _omitFieldNames ? '' : 'updatedAt')
    ..aOS(12, _omitFieldNames ? '' : 'startedAt')
    ..aOS(13, _omitFieldNames ? '' : 'completedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogisticsDelivery clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogisticsDelivery copyWith(void Function(LogisticsDelivery) updates) =>
      super.copyWith((message) => updates(message as LogisticsDelivery))
          as LogisticsDelivery;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogisticsDelivery() / LogisticsDelivery.new instead')
  static LogisticsDelivery create() => LogisticsDelivery._();
  static $pb.GeneratedMessage $_createMessage() => LogisticsDelivery._();
  @$core.override
  LogisticsDelivery createEmptyInstance() => LogisticsDelivery._();
  @$core.pragma('dart2js:noInline')
  static LogisticsDelivery getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogisticsDelivery>(
          LogisticsDelivery.$_createMessage);
  static LogisticsDelivery? _defaultInstance;

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
  $core.String get routeId => $_getSZ(3);
  @$pb.TagNumber(4)
  set routeId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRouteId() => $_has(3);
  @$pb.TagNumber(4)
  void clearRouteId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get driverId => $_getSZ(4);
  @$pb.TagNumber(5)
  set driverId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDriverId() => $_has(4);
  @$pb.TagNumber(5)
  void clearDriverId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get vehicleId => $_getSZ(5);
  @$pb.TagNumber(6)
  set vehicleId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasVehicleId() => $_has(5);
  @$pb.TagNumber(6)
  void clearVehicleId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get assignedBy => $_getSZ(6);
  @$pb.TagNumber(7)
  set assignedBy($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasAssignedBy() => $_has(6);
  @$pb.TagNumber(7)
  void clearAssignedBy() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get status => $_getSZ(7);
  @$pb.TagNumber(8)
  set status($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearStatus() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get version => $_getSZ(8);
  @$pb.TagNumber(9)
  set version($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasVersion() => $_has(8);
  @$pb.TagNumber(9)
  void clearVersion() => $_clearField(9);

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
  $core.String get startedAt => $_getSZ(11);
  @$pb.TagNumber(12)
  set startedAt($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasStartedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearStartedAt() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get completedAt => $_getSZ(12);
  @$pb.TagNumber(13)
  set completedAt($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasCompletedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearCompletedAt() => $_clearField(13);
}

/// AssignDeliveryRequest:指派請求。
/// 新建時 version 帶 "0";重指派帶讀取到的 version(不符 → failed_precondition)。
class AssignDeliveryRequest extends $pb.GeneratedMessage {
  factory AssignDeliveryRequest({
    $core.String? routeId,
    $core.String? driverId,
    $core.String? vehicleId,
    $core.String? version,
  }) {
    final result = AssignDeliveryRequest._();
    if (routeId != null) result.routeId = routeId;
    if (driverId != null) result.driverId = driverId;
    if (vehicleId != null) result.vehicleId = vehicleId;
    if (version != null) result.version = version;
    return result;
  }

  AssignDeliveryRequest._();

  factory AssignDeliveryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignDeliveryRequest()..mergeFromBuffer(data, registry);
  factory AssignDeliveryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignDeliveryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignDeliveryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: AssignDeliveryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'routeId')
    ..aOS(2, _omitFieldNames ? '' : 'driverId')
    ..aOS(3, _omitFieldNames ? '' : 'vehicleId')
    ..aOS(4, _omitFieldNames ? '' : 'version')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignDeliveryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignDeliveryRequest copyWith(
          void Function(AssignDeliveryRequest) updates) =>
      super.copyWith((message) => updates(message as AssignDeliveryRequest))
          as AssignDeliveryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AssignDeliveryRequest() / AssignDeliveryRequest.new instead')
  static AssignDeliveryRequest create() => AssignDeliveryRequest._();
  static $pb.GeneratedMessage $_createMessage() => AssignDeliveryRequest._();
  @$core.override
  AssignDeliveryRequest createEmptyInstance() => AssignDeliveryRequest._();
  @$core.pragma('dart2js:noInline')
  static AssignDeliveryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignDeliveryRequest>(
          AssignDeliveryRequest.$_createMessage);
  static AssignDeliveryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get routeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set routeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRouteId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRouteId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get driverId => $_getSZ(1);
  @$pb.TagNumber(2)
  set driverId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDriverId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDriverId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get vehicleId => $_getSZ(2);
  @$pb.TagNumber(3)
  set vehicleId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVehicleId() => $_has(2);
  @$pb.TagNumber(3)
  void clearVehicleId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get version => $_getSZ(3);
  @$pb.TagNumber(4)
  set version($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => $_clearField(4);
}

/// AssignDeliveryResponse:指派結果(含遞增後的 version)。
class AssignDeliveryResponse extends $pb.GeneratedMessage {
  factory AssignDeliveryResponse({
    LogisticsDelivery? delivery,
  }) {
    final result = AssignDeliveryResponse._();
    if (delivery != null) result.delivery = delivery;
    return result;
  }

  AssignDeliveryResponse._();

  factory AssignDeliveryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignDeliveryResponse()..mergeFromBuffer(data, registry);
  factory AssignDeliveryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignDeliveryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignDeliveryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: AssignDeliveryResponse.$_createMessage)
    ..aOM<LogisticsDelivery>(1, _omitFieldNames ? '' : 'delivery',
        subBuilder: LogisticsDelivery.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignDeliveryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignDeliveryResponse copyWith(
          void Function(AssignDeliveryResponse) updates) =>
      super.copyWith((message) => updates(message as AssignDeliveryResponse))
          as AssignDeliveryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AssignDeliveryResponse() / AssignDeliveryResponse.new instead')
  static AssignDeliveryResponse create() => AssignDeliveryResponse._();
  static $pb.GeneratedMessage $_createMessage() => AssignDeliveryResponse._();
  @$core.override
  AssignDeliveryResponse createEmptyInstance() => AssignDeliveryResponse._();
  @$core.pragma('dart2js:noInline')
  static AssignDeliveryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignDeliveryResponse>(
          AssignDeliveryResponse.$_createMessage);
  static AssignDeliveryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LogisticsDelivery get delivery => $_getN(0);
  @$pb.TagNumber(1)
  set delivery(LogisticsDelivery value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDelivery() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelivery() => $_clearField(1);
  @$pb.TagNumber(1)
  LogisticsDelivery ensureDelivery() => $_ensure(0);
}

/// ListMyDeliveriesRequest:我的配送清單請求。
class ListMyDeliveriesRequest extends $pb.GeneratedMessage {
  factory ListMyDeliveriesRequest({
    $core.int? page,
    $core.int? pageSize,
  }) {
    final result = ListMyDeliveriesRequest._();
    if (page != null) result.page = page;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListMyDeliveriesRequest._();

  factory ListMyDeliveriesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMyDeliveriesRequest()..mergeFromBuffer(data, registry);
  factory ListMyDeliveriesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMyDeliveriesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMyDeliveriesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListMyDeliveriesRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'page')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMyDeliveriesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMyDeliveriesRequest copyWith(
          void Function(ListMyDeliveriesRequest) updates) =>
      super.copyWith((message) => updates(message as ListMyDeliveriesRequest))
          as ListMyDeliveriesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListMyDeliveriesRequest() / ListMyDeliveriesRequest.new instead')
  static ListMyDeliveriesRequest create() => ListMyDeliveriesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListMyDeliveriesRequest._();
  @$core.override
  ListMyDeliveriesRequest createEmptyInstance() => ListMyDeliveriesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListMyDeliveriesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMyDeliveriesRequest>(
          ListMyDeliveriesRequest.$_createMessage);
  static ListMyDeliveriesRequest? _defaultInstance;

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
}

/// ListMyDeliveriesResponse:我的配送清單結果。
class ListMyDeliveriesResponse extends $pb.GeneratedMessage {
  factory ListMyDeliveriesResponse({
    $core.Iterable<LogisticsDelivery>? deliveries,
    $core.int? total,
  }) {
    final result = ListMyDeliveriesResponse._();
    if (deliveries != null) result.deliveries.addAll(deliveries);
    if (total != null) result.total = total;
    return result;
  }

  ListMyDeliveriesResponse._();

  factory ListMyDeliveriesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMyDeliveriesResponse()..mergeFromBuffer(data, registry);
  factory ListMyDeliveriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMyDeliveriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMyDeliveriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: ListMyDeliveriesResponse.$_createMessage)
    ..pPM<LogisticsDelivery>(1, _omitFieldNames ? '' : 'deliveries',
        subBuilder: LogisticsDelivery.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMyDeliveriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMyDeliveriesResponse copyWith(
          void Function(ListMyDeliveriesResponse) updates) =>
      super.copyWith((message) => updates(message as ListMyDeliveriesResponse))
          as ListMyDeliveriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListMyDeliveriesResponse() / ListMyDeliveriesResponse.new instead')
  static ListMyDeliveriesResponse create() => ListMyDeliveriesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListMyDeliveriesResponse._();
  @$core.override
  ListMyDeliveriesResponse createEmptyInstance() =>
      ListMyDeliveriesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListMyDeliveriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMyDeliveriesResponse>(
          ListMyDeliveriesResponse.$_createMessage);
  static ListMyDeliveriesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<LogisticsDelivery> get deliveries => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get total => $_getIZ(1);
  @$pb.TagNumber(2)
  set total($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => $_clearField(2);
}

/// LogisticsProof:簽收證明 POD(D17 檔案資產)。
class LogisticsProof extends $pb.GeneratedMessage {
  factory LogisticsProof({
    $core.String? id,
    $core.String? logisticsDeliveryId,
    $core.String? proofType,
    $core.String? fileAssetId,
    $core.String? remarks,
    $core.String? capturedAt,
  }) {
    final result = LogisticsProof._();
    if (id != null) result.id = id;
    if (logisticsDeliveryId != null)
      result.logisticsDeliveryId = logisticsDeliveryId;
    if (proofType != null) result.proofType = proofType;
    if (fileAssetId != null) result.fileAssetId = fileAssetId;
    if (remarks != null) result.remarks = remarks;
    if (capturedAt != null) result.capturedAt = capturedAt;
    return result;
  }

  LogisticsProof._();

  factory LogisticsProof.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogisticsProof()..mergeFromBuffer(data, registry);
  factory LogisticsProof.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogisticsProof()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogisticsProof',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: LogisticsProof.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'logisticsDeliveryId')
    ..aOS(3, _omitFieldNames ? '' : 'proofType')
    ..aOS(4, _omitFieldNames ? '' : 'fileAssetId')
    ..aOS(5, _omitFieldNames ? '' : 'remarks')
    ..aOS(6, _omitFieldNames ? '' : 'capturedAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogisticsProof clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogisticsProof copyWith(void Function(LogisticsProof) updates) =>
      super.copyWith((message) => updates(message as LogisticsProof))
          as LogisticsProof;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogisticsProof() / LogisticsProof.new instead')
  static LogisticsProof create() => LogisticsProof._();
  static $pb.GeneratedMessage $_createMessage() => LogisticsProof._();
  @$core.override
  LogisticsProof createEmptyInstance() => LogisticsProof._();
  @$core.pragma('dart2js:noInline')
  static LogisticsProof getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogisticsProof>(
          LogisticsProof.$_createMessage);
  static LogisticsProof? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get logisticsDeliveryId => $_getSZ(1);
  @$pb.TagNumber(2)
  set logisticsDeliveryId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLogisticsDeliveryId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLogisticsDeliveryId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get proofType => $_getSZ(2);
  @$pb.TagNumber(3)
  set proofType($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProofType() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofType() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get fileAssetId => $_getSZ(3);
  @$pb.TagNumber(4)
  set fileAssetId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasFileAssetId() => $_has(3);
  @$pb.TagNumber(4)
  void clearFileAssetId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get remarks => $_getSZ(4);
  @$pb.TagNumber(5)
  set remarks($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRemarks() => $_has(4);
  @$pb.TagNumber(5)
  void clearRemarks() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get capturedAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set capturedAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCapturedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCapturedAt() => $_clearField(6);
}

/// StartDeliveryRequest:開始配送請求(id + version 樂觀鎖)。
class StartDeliveryRequest extends $pb.GeneratedMessage {
  factory StartDeliveryRequest({
    $core.String? deliveryId,
    $core.String? version,
  }) {
    final result = StartDeliveryRequest._();
    if (deliveryId != null) result.deliveryId = deliveryId;
    if (version != null) result.version = version;
    return result;
  }

  StartDeliveryRequest._();

  factory StartDeliveryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartDeliveryRequest()..mergeFromBuffer(data, registry);
  factory StartDeliveryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartDeliveryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartDeliveryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: StartDeliveryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'deliveryId')
    ..aOS(2, _omitFieldNames ? '' : 'version')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartDeliveryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartDeliveryRequest copyWith(void Function(StartDeliveryRequest) updates) =>
      super.copyWith((message) => updates(message as StartDeliveryRequest))
          as StartDeliveryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StartDeliveryRequest() / StartDeliveryRequest.new instead')
  static StartDeliveryRequest create() => StartDeliveryRequest._();
  static $pb.GeneratedMessage $_createMessage() => StartDeliveryRequest._();
  @$core.override
  StartDeliveryRequest createEmptyInstance() => StartDeliveryRequest._();
  @$core.pragma('dart2js:noInline')
  static StartDeliveryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartDeliveryRequest>(
          StartDeliveryRequest.$_createMessage);
  static StartDeliveryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get deliveryId => $_getSZ(0);
  @$pb.TagNumber(1)
  set deliveryId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeliveryId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeliveryId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get version => $_getSZ(1);
  @$pb.TagNumber(2)
  set version($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearVersion() => $_clearField(2);
}

/// StartDeliveryResponse:開始配送結果。
class StartDeliveryResponse extends $pb.GeneratedMessage {
  factory StartDeliveryResponse({
    LogisticsDelivery? delivery,
  }) {
    final result = StartDeliveryResponse._();
    if (delivery != null) result.delivery = delivery;
    return result;
  }

  StartDeliveryResponse._();

  factory StartDeliveryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartDeliveryResponse()..mergeFromBuffer(data, registry);
  factory StartDeliveryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartDeliveryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartDeliveryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: StartDeliveryResponse.$_createMessage)
    ..aOM<LogisticsDelivery>(1, _omitFieldNames ? '' : 'delivery',
        subBuilder: LogisticsDelivery.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartDeliveryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartDeliveryResponse copyWith(
          void Function(StartDeliveryResponse) updates) =>
      super.copyWith((message) => updates(message as StartDeliveryResponse))
          as StartDeliveryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StartDeliveryResponse() / StartDeliveryResponse.new instead')
  static StartDeliveryResponse create() => StartDeliveryResponse._();
  static $pb.GeneratedMessage $_createMessage() => StartDeliveryResponse._();
  @$core.override
  StartDeliveryResponse createEmptyInstance() => StartDeliveryResponse._();
  @$core.pragma('dart2js:noInline')
  static StartDeliveryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartDeliveryResponse>(
          StartDeliveryResponse.$_createMessage);
  static StartDeliveryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LogisticsDelivery get delivery => $_getN(0);
  @$pb.TagNumber(1)
  set delivery(LogisticsDelivery value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDelivery() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelivery() => $_clearField(1);
  @$pb.TagNumber(1)
  LogisticsDelivery ensureDelivery() => $_ensure(0);
}

/// CompleteDeliveryRequest:完成配送請求。proofs 可空(無簽收亦可完成);
/// 每筆 proof 的 file_asset_id 來自既有檔案上傳端點(POST /api/v1/files,owner_type=logistics_delivery)。
class CompleteDeliveryRequest extends $pb.GeneratedMessage {
  factory CompleteDeliveryRequest({
    $core.String? deliveryId,
    $core.String? version,
    $core.Iterable<CompleteProof>? proofs,
  }) {
    final result = CompleteDeliveryRequest._();
    if (deliveryId != null) result.deliveryId = deliveryId;
    if (version != null) result.version = version;
    if (proofs != null) result.proofs.addAll(proofs);
    return result;
  }

  CompleteDeliveryRequest._();

  factory CompleteDeliveryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteDeliveryRequest()..mergeFromBuffer(data, registry);
  factory CompleteDeliveryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteDeliveryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteDeliveryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CompleteDeliveryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'deliveryId')
    ..aOS(2, _omitFieldNames ? '' : 'version')
    ..pPM<CompleteProof>(3, _omitFieldNames ? '' : 'proofs',
        subBuilder: CompleteProof.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteDeliveryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteDeliveryRequest copyWith(
          void Function(CompleteDeliveryRequest) updates) =>
      super.copyWith((message) => updates(message as CompleteDeliveryRequest))
          as CompleteDeliveryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteDeliveryRequest() / CompleteDeliveryRequest.new instead')
  static CompleteDeliveryRequest create() => CompleteDeliveryRequest._();
  static $pb.GeneratedMessage $_createMessage() => CompleteDeliveryRequest._();
  @$core.override
  CompleteDeliveryRequest createEmptyInstance() => CompleteDeliveryRequest._();
  @$core.pragma('dart2js:noInline')
  static CompleteDeliveryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteDeliveryRequest>(
          CompleteDeliveryRequest.$_createMessage);
  static CompleteDeliveryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get deliveryId => $_getSZ(0);
  @$pb.TagNumber(1)
  set deliveryId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeliveryId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeliveryId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get version => $_getSZ(1);
  @$pb.TagNumber(2)
  set version($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearVersion() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<CompleteProof> get proofs => $_getList(2);
}

/// CompleteProof:一筆簽收證明(型別 + 檔案資產 id + 備註)。
class CompleteProof extends $pb.GeneratedMessage {
  factory CompleteProof({
    $core.String? proofType,
    $core.String? fileAssetId,
    $core.String? remarks,
  }) {
    final result = CompleteProof._();
    if (proofType != null) result.proofType = proofType;
    if (fileAssetId != null) result.fileAssetId = fileAssetId;
    if (remarks != null) result.remarks = remarks;
    return result;
  }

  CompleteProof._();

  factory CompleteProof.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteProof()..mergeFromBuffer(data, registry);
  factory CompleteProof.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteProof()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteProof',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CompleteProof.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'proofType')
    ..aOS(2, _omitFieldNames ? '' : 'fileAssetId')
    ..aOS(3, _omitFieldNames ? '' : 'remarks')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteProof clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteProof copyWith(void Function(CompleteProof) updates) =>
      super.copyWith((message) => updates(message as CompleteProof))
          as CompleteProof;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CompleteProof() / CompleteProof.new instead')
  static CompleteProof create() => CompleteProof._();
  static $pb.GeneratedMessage $_createMessage() => CompleteProof._();
  @$core.override
  CompleteProof createEmptyInstance() => CompleteProof._();
  @$core.pragma('dart2js:noInline')
  static CompleteProof getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompleteProof>(
          CompleteProof.$_createMessage);
  static CompleteProof? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get proofType => $_getSZ(0);
  @$pb.TagNumber(1)
  set proofType($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProofType() => $_has(0);
  @$pb.TagNumber(1)
  void clearProofType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get fileAssetId => $_getSZ(1);
  @$pb.TagNumber(2)
  set fileAssetId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFileAssetId() => $_has(1);
  @$pb.TagNumber(2)
  void clearFileAssetId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get remarks => $_getSZ(2);
  @$pb.TagNumber(3)
  set remarks($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRemarks() => $_has(2);
  @$pb.TagNumber(3)
  void clearRemarks() => $_clearField(3);
}

/// CompleteDeliveryResponse:完成配送結果(含寫入的 proofs)。
class CompleteDeliveryResponse extends $pb.GeneratedMessage {
  factory CompleteDeliveryResponse({
    LogisticsDelivery? delivery,
    $core.Iterable<LogisticsProof>? proofs,
  }) {
    final result = CompleteDeliveryResponse._();
    if (delivery != null) result.delivery = delivery;
    if (proofs != null) result.proofs.addAll(proofs);
    return result;
  }

  CompleteDeliveryResponse._();

  factory CompleteDeliveryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteDeliveryResponse()..mergeFromBuffer(data, registry);
  factory CompleteDeliveryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteDeliveryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteDeliveryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CompleteDeliveryResponse.$_createMessage)
    ..aOM<LogisticsDelivery>(1, _omitFieldNames ? '' : 'delivery',
        subBuilder: LogisticsDelivery.$_createMessage)
    ..pPM<LogisticsProof>(2, _omitFieldNames ? '' : 'proofs',
        subBuilder: LogisticsProof.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteDeliveryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteDeliveryResponse copyWith(
          void Function(CompleteDeliveryResponse) updates) =>
      super.copyWith((message) => updates(message as CompleteDeliveryResponse))
          as CompleteDeliveryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteDeliveryResponse() / CompleteDeliveryResponse.new instead')
  static CompleteDeliveryResponse create() => CompleteDeliveryResponse._();
  static $pb.GeneratedMessage $_createMessage() => CompleteDeliveryResponse._();
  @$core.override
  CompleteDeliveryResponse createEmptyInstance() =>
      CompleteDeliveryResponse._();
  @$core.pragma('dart2js:noInline')
  static CompleteDeliveryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteDeliveryResponse>(
          CompleteDeliveryResponse.$_createMessage);
  static CompleteDeliveryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LogisticsDelivery get delivery => $_getN(0);
  @$pb.TagNumber(1)
  set delivery(LogisticsDelivery value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDelivery() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelivery() => $_clearField(1);
  @$pb.TagNumber(1)
  LogisticsDelivery ensureDelivery() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<LogisticsProof> get proofs => $_getList(1);
}

/// CancelDeliveryRequest:取消配送請求(reason 必填)。
class CancelDeliveryRequest extends $pb.GeneratedMessage {
  factory CancelDeliveryRequest({
    $core.String? deliveryId,
    $core.String? version,
    $core.String? reason,
  }) {
    final result = CancelDeliveryRequest._();
    if (deliveryId != null) result.deliveryId = deliveryId;
    if (version != null) result.version = version;
    if (reason != null) result.reason = reason;
    return result;
  }

  CancelDeliveryRequest._();

  factory CancelDeliveryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDeliveryRequest()..mergeFromBuffer(data, registry);
  factory CancelDeliveryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDeliveryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelDeliveryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CancelDeliveryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'deliveryId')
    ..aOS(2, _omitFieldNames ? '' : 'version')
    ..aOS(3, _omitFieldNames ? '' : 'reason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDeliveryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDeliveryRequest copyWith(
          void Function(CancelDeliveryRequest) updates) =>
      super.copyWith((message) => updates(message as CancelDeliveryRequest))
          as CancelDeliveryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelDeliveryRequest() / CancelDeliveryRequest.new instead')
  static CancelDeliveryRequest create() => CancelDeliveryRequest._();
  static $pb.GeneratedMessage $_createMessage() => CancelDeliveryRequest._();
  @$core.override
  CancelDeliveryRequest createEmptyInstance() => CancelDeliveryRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelDeliveryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelDeliveryRequest>(
          CancelDeliveryRequest.$_createMessage);
  static CancelDeliveryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get deliveryId => $_getSZ(0);
  @$pb.TagNumber(1)
  set deliveryId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeliveryId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeliveryId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get version => $_getSZ(1);
  @$pb.TagNumber(2)
  set version($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearVersion() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get reason => $_getSZ(2);
  @$pb.TagNumber(3)
  set reason($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasReason() => $_has(2);
  @$pb.TagNumber(3)
  void clearReason() => $_clearField(3);
}

/// CancelDeliveryResponse:取消配送結果。
class CancelDeliveryResponse extends $pb.GeneratedMessage {
  factory CancelDeliveryResponse({
    LogisticsDelivery? delivery,
  }) {
    final result = CancelDeliveryResponse._();
    if (delivery != null) result.delivery = delivery;
    return result;
  }

  CancelDeliveryResponse._();

  factory CancelDeliveryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDeliveryResponse()..mergeFromBuffer(data, registry);
  factory CancelDeliveryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDeliveryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelDeliveryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: CancelDeliveryResponse.$_createMessage)
    ..aOM<LogisticsDelivery>(1, _omitFieldNames ? '' : 'delivery',
        subBuilder: LogisticsDelivery.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDeliveryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDeliveryResponse copyWith(
          void Function(CancelDeliveryResponse) updates) =>
      super.copyWith((message) => updates(message as CancelDeliveryResponse))
          as CancelDeliveryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelDeliveryResponse() / CancelDeliveryResponse.new instead')
  static CancelDeliveryResponse create() => CancelDeliveryResponse._();
  static $pb.GeneratedMessage $_createMessage() => CancelDeliveryResponse._();
  @$core.override
  CancelDeliveryResponse createEmptyInstance() => CancelDeliveryResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelDeliveryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelDeliveryResponse>(
          CancelDeliveryResponse.$_createMessage);
  static CancelDeliveryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LogisticsDelivery get delivery => $_getN(0);
  @$pb.TagNumber(1)
  set delivery(LogisticsDelivery value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDelivery() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelivery() => $_clearField(1);
  @$pb.TagNumber(1)
  LogisticsDelivery ensureDelivery() => $_ensure(0);
}

/// LogisticsService:logistics 執行層首批(D32/10.1/10.4/10.12)。
/// 建檔(司機/車輛)與指派為後台動作(dept_admin 以上;rolePolicy logistics);
/// ListMyDeliveries 為**被指派司機本人**的任務清單(10.12:身分必須對應 logistics_drivers 列,
/// instance 級每列再經 OpenFGA Check —— 10.8:被指派司機可操作其 delivery、他人 403)。
class LogisticsServiceApi {
  final $pb.RpcClient _client;

  LogisticsServiceApi(this._client);

  /// CreateDriver:建司機(關聯既有 users;同部門 user 不重複建)。
  $async.Future<CreateDriverResponse> createDriver(
          $pb.ClientContext? ctx, CreateDriverRequest request) =>
      _client.invoke<CreateDriverResponse>(ctx, 'LogisticsService',
          'CreateDriver', request, CreateDriverResponse());

  /// CreateVehicle:建車輛(plate_no 部門內唯一,重複 → already_exists)。
  $async.Future<CreateVehicleResponse> createVehicle(
          $pb.ClientContext? ctx, CreateVehicleRequest request) =>
      _client.invoke<CreateVehicleResponse>(ctx, 'LogisticsService',
          'CreateVehicle', request, CreateVehicleResponse());

  /// AssignDelivery:指派車次 → 司機/車輛(粒度 = route;存在即重指派,
  /// version 樂觀鎖:衝突 → failed_precondition;同交易寫稽核,tuple 經 AfterCommit 同步)。
  $async.Future<AssignDeliveryResponse> assignDelivery(
          $pb.ClientContext? ctx, AssignDeliveryRequest request) =>
      _client.invoke<AssignDeliveryResponse>(ctx, 'LogisticsService',
          'AssignDelivery', request, AssignDeliveryResponse());

  /// ListMyDeliveries:我(drivers.user_id = 身分)被指派的配送清單。
  $async.Future<ListMyDeliveriesResponse> listMyDeliveries(
          $pb.ClientContext? ctx, ListMyDeliveriesRequest request) =>
      _client.invoke<ListMyDeliveriesResponse>(ctx, 'LogisticsService',
          'ListMyDeliveries', request, ListMyDeliveriesResponse());

  /// StartDelivery:被指派司機開始執行(pending → in_progress;10.6)。
  $async.Future<StartDeliveryResponse> startDelivery(
          $pb.ClientContext? ctx, StartDeliveryRequest request) =>
      _client.invoke<StartDeliveryResponse>(ctx, 'LogisticsService',
          'StartDelivery', request, StartDeliveryResponse());

  /// CompleteDelivery:完成並簽收(in_progress → completed;POD 可多筆,同一交易寫事件與稽核)。
  $async.Future<CompleteDeliveryResponse> completeDelivery(
          $pb.ClientContext? ctx, CompleteDeliveryRequest request) =>
      _client.invoke<CompleteDeliveryResponse>(ctx, 'LogisticsService',
          'CompleteDelivery', request, CompleteDeliveryResponse());

  /// CancelDelivery:取消配送(pending/in_progress → cancelled;reason 必填)。
  $async.Future<CancelDeliveryResponse> cancelDelivery(
          $pb.ClientContext? ctx, CancelDeliveryRequest request) =>
      _client.invoke<CancelDeliveryResponse>(ctx, 'LogisticsService',
          'CancelDelivery', request, CancelDeliveryResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
