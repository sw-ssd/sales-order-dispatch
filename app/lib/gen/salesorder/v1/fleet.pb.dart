// This is a generated file - do not edit.
//
// Generated from salesorder/v1/fleet.proto.

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

/// FleetDriver:司機單筆。
class FleetDriver extends $pb.GeneratedMessage {
  factory FleetDriver({
    $core.String? id,
    $core.String? companyId,
    $core.String? departmentId,
    $core.String? userId,
    $core.String? name,
    $core.String? phone,
    $core.String? currentStatus,
  }) {
    final result = FleetDriver._();
    if (id != null) result.id = id;
    if (companyId != null) result.companyId = companyId;
    if (departmentId != null) result.departmentId = departmentId;
    if (userId != null) result.userId = userId;
    if (name != null) result.name = name;
    if (phone != null) result.phone = phone;
    if (currentStatus != null) result.currentStatus = currentStatus;
    return result;
  }

  FleetDriver._();

  factory FleetDriver.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FleetDriver()..mergeFromBuffer(data, registry);
  factory FleetDriver.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FleetDriver()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FleetDriver',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: FleetDriver.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'companyId')
    ..aOS(3, _omitFieldNames ? '' : 'departmentId')
    ..aOS(4, _omitFieldNames ? '' : 'userId')
    ..aOS(5, _omitFieldNames ? '' : 'name')
    ..aOS(6, _omitFieldNames ? '' : 'phone')
    ..aOS(7, _omitFieldNames ? '' : 'currentStatus')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FleetDriver clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FleetDriver copyWith(void Function(FleetDriver) updates) =>
      super.copyWith((message) => updates(message as FleetDriver))
          as FleetDriver;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use FleetDriver() / FleetDriver.new instead')
  static FleetDriver create() => FleetDriver._();
  static $pb.GeneratedMessage $_createMessage() => FleetDriver._();
  @$core.override
  FleetDriver createEmptyInstance() => FleetDriver._();
  @$core.pragma('dart2js:noInline')
  static FleetDriver getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FleetDriver>(
          FleetDriver.$_createMessage);
  static FleetDriver? _defaultInstance;

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
    FleetDriver? driver,
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
    ..aOM<FleetDriver>(1, _omitFieldNames ? '' : 'driver',
        subBuilder: FleetDriver.$_createMessage)
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
  FleetDriver get driver => $_getN(0);
  @$pb.TagNumber(1)
  set driver(FleetDriver value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDriver() => $_has(0);
  @$pb.TagNumber(1)
  void clearDriver() => $_clearField(1);
  @$pb.TagNumber(1)
  FleetDriver ensureDriver() => $_ensure(0);
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

/// FleetDelivery:配送執行單元(route 粒度)。
class FleetDelivery extends $pb.GeneratedMessage {
  factory FleetDelivery({
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
  }) {
    final result = FleetDelivery._();
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
    return result;
  }

  FleetDelivery._();

  factory FleetDelivery.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FleetDelivery()..mergeFromBuffer(data, registry);
  factory FleetDelivery.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      FleetDelivery()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FleetDelivery',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'salesorder.v1'),
      createEmptyInstance: FleetDelivery.$_createMessage)
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
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FleetDelivery clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FleetDelivery copyWith(void Function(FleetDelivery) updates) =>
      super.copyWith((message) => updates(message as FleetDelivery))
          as FleetDelivery;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use FleetDelivery() / FleetDelivery.new instead')
  static FleetDelivery create() => FleetDelivery._();
  static $pb.GeneratedMessage $_createMessage() => FleetDelivery._();
  @$core.override
  FleetDelivery createEmptyInstance() => FleetDelivery._();
  @$core.pragma('dart2js:noInline')
  static FleetDelivery getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FleetDelivery>(
          FleetDelivery.$_createMessage);
  static FleetDelivery? _defaultInstance;

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
    FleetDelivery? delivery,
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
    ..aOM<FleetDelivery>(1, _omitFieldNames ? '' : 'delivery',
        subBuilder: FleetDelivery.$_createMessage)
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
  FleetDelivery get delivery => $_getN(0);
  @$pb.TagNumber(1)
  set delivery(FleetDelivery value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDelivery() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelivery() => $_clearField(1);
  @$pb.TagNumber(1)
  FleetDelivery ensureDelivery() => $_ensure(0);
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
    $core.Iterable<FleetDelivery>? deliveries,
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
    ..pPM<FleetDelivery>(1, _omitFieldNames ? '' : 'deliveries',
        subBuilder: FleetDelivery.$_createMessage)
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
  $pb.PbList<FleetDelivery> get deliveries => $_getList(0);

  @$pb.TagNumber(2)
  $core.int get total => $_getIZ(1);
  @$pb.TagNumber(2)
  set total($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => $_clearField(2);
}

/// FleetService:fleet 執行層首批(D32/10.1/10.4/10.12)。
/// 建檔(司機/車輛)與指派為後台動作(dept_admin 以上;rolePolicy fleet);
/// ListMyDeliveries 為**被指派司機本人**的任務清單(10.12:身分必須對應 fleet_drivers 列,
/// instance 級每列再經 OpenFGA Check —— 10.8:被指派司機可操作其 delivery、他人 403)。
class FleetServiceApi {
  final $pb.RpcClient _client;

  FleetServiceApi(this._client);

  /// CreateDriver:建司機(關聯既有 users;同部門 user 不重複建)。
  $async.Future<CreateDriverResponse> createDriver(
          $pb.ClientContext? ctx, CreateDriverRequest request) =>
      _client.invoke<CreateDriverResponse>(
          ctx, 'FleetService', 'CreateDriver', request, CreateDriverResponse());

  /// CreateVehicle:建車輛(plate_no 部門內唯一,重複 → already_exists)。
  $async.Future<CreateVehicleResponse> createVehicle(
          $pb.ClientContext? ctx, CreateVehicleRequest request) =>
      _client.invoke<CreateVehicleResponse>(ctx, 'FleetService',
          'CreateVehicle', request, CreateVehicleResponse());

  /// AssignDelivery:指派車次 → 司機/車輛(粒度 = route;存在即重指派,
  /// version 樂觀鎖:衝突 → failed_precondition;同交易寫稽核,tuple 經 AfterCommit 同步)。
  $async.Future<AssignDeliveryResponse> assignDelivery(
          $pb.ClientContext? ctx, AssignDeliveryRequest request) =>
      _client.invoke<AssignDeliveryResponse>(ctx, 'FleetService',
          'AssignDelivery', request, AssignDeliveryResponse());

  /// ListMyDeliveries:我(drivers.user_id = 身分)被指派的配送清單。
  $async.Future<ListMyDeliveriesResponse> listMyDeliveries(
          $pb.ClientContext? ctx, ListMyDeliveriesRequest request) =>
      _client.invoke<ListMyDeliveriesResponse>(ctx, 'FleetService',
          'ListMyDeliveries', request, ListMyDeliveriesResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
