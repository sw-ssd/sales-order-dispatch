// This is a generated file - do not edit.
//
// Generated from salesorder/v1/fleet.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use fleetDriverDescriptor instead')
const FleetDriver$json = {
  '1': 'FleetDriver',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'user_id', '3': 4, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'name', '3': 5, '4': 1, '5': 9, '10': 'name'},
    {'1': 'phone', '3': 6, '4': 1, '5': 9, '10': 'phone'},
    {'1': 'current_status', '3': 7, '4': 1, '5': 9, '10': 'currentStatus'},
  ],
};

/// Descriptor for `FleetDriver`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List fleetDriverDescriptor = $convert.base64Decode(
    'CgtGbGVldERyaXZlchIOCgJpZBgBIAEoCVICaWQSHQoKY29tcGFueV9pZBgCIAEoCVIJY29tcG'
    'FueUlkEiMKDWRlcGFydG1lbnRfaWQYAyABKAlSDGRlcGFydG1lbnRJZBIXCgd1c2VyX2lkGAQg'
    'ASgJUgZ1c2VySWQSEgoEbmFtZRgFIAEoCVIEbmFtZRIUCgVwaG9uZRgGIAEoCVIFcGhvbmUSJQ'
    'oOY3VycmVudF9zdGF0dXMYByABKAlSDWN1cnJlbnRTdGF0dXM=');

@$core.Deprecated('Use createDriverRequestDescriptor instead')
const CreateDriverRequest$json = {
  '1': 'CreateDriverRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'phone', '3': 3, '4': 1, '5': 9, '10': 'phone'},
  ],
};

/// Descriptor for `CreateDriverRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDriverRequestDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVEcml2ZXJSZXF1ZXN0EhcKB3VzZXJfaWQYASABKAlSBnVzZXJJZBISCgRuYW1lGA'
    'IgASgJUgRuYW1lEhQKBXBob25lGAMgASgJUgVwaG9uZQ==');

@$core.Deprecated('Use createDriverResponseDescriptor instead')
const CreateDriverResponse$json = {
  '1': 'CreateDriverResponse',
  '2': [
    {
      '1': 'driver',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.FleetDriver',
      '10': 'driver'
    },
  ],
};

/// Descriptor for `CreateDriverResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDriverResponseDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVEcml2ZXJSZXNwb25zZRIyCgZkcml2ZXIYASABKAsyGi5zYWxlc29yZGVyLnYxLk'
    'ZsZWV0RHJpdmVyUgZkcml2ZXI=');

@$core.Deprecated('Use vehicleDescriptor instead')
const Vehicle$json = {
  '1': 'Vehicle',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'plate_no', '3': 4, '4': 1, '5': 9, '10': 'plateNo'},
    {'1': 'vehicle_type', '3': 5, '4': 1, '5': 9, '10': 'vehicleType'},
    {'1': 'status', '3': 6, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `Vehicle`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List vehicleDescriptor = $convert.base64Decode(
    'CgdWZWhpY2xlEg4KAmlkGAEgASgJUgJpZBIdCgpjb21wYW55X2lkGAIgASgJUgljb21wYW55SW'
    'QSIwoNZGVwYXJ0bWVudF9pZBgDIAEoCVIMZGVwYXJ0bWVudElkEhkKCHBsYXRlX25vGAQgASgJ'
    'UgdwbGF0ZU5vEiEKDHZlaGljbGVfdHlwZRgFIAEoCVILdmVoaWNsZVR5cGUSFgoGc3RhdHVzGA'
    'YgASgJUgZzdGF0dXM=');

@$core.Deprecated('Use createVehicleRequestDescriptor instead')
const CreateVehicleRequest$json = {
  '1': 'CreateVehicleRequest',
  '2': [
    {'1': 'plate_no', '3': 1, '4': 1, '5': 9, '10': 'plateNo'},
    {'1': 'vehicle_type', '3': 2, '4': 1, '5': 9, '10': 'vehicleType'},
  ],
};

/// Descriptor for `CreateVehicleRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createVehicleRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVWZWhpY2xlUmVxdWVzdBIZCghwbGF0ZV9ubxgBIAEoCVIHcGxhdGVObxIhCgx2ZW'
    'hpY2xlX3R5cGUYAiABKAlSC3ZlaGljbGVUeXBl');

@$core.Deprecated('Use createVehicleResponseDescriptor instead')
const CreateVehicleResponse$json = {
  '1': 'CreateVehicleResponse',
  '2': [
    {
      '1': 'vehicle',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.Vehicle',
      '10': 'vehicle'
    },
  ],
};

/// Descriptor for `CreateVehicleResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createVehicleResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVWZWhpY2xlUmVzcG9uc2USMAoHdmVoaWNsZRgBIAEoCzIWLnNhbGVzb3JkZXIudj'
    'EuVmVoaWNsZVIHdmVoaWNsZQ==');

@$core.Deprecated('Use fleetDeliveryDescriptor instead')
const FleetDelivery$json = {
  '1': 'FleetDelivery',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'company_id', '3': 2, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'department_id', '3': 3, '4': 1, '5': 9, '10': 'departmentId'},
    {'1': 'route_id', '3': 4, '4': 1, '5': 9, '10': 'routeId'},
    {'1': 'driver_id', '3': 5, '4': 1, '5': 9, '10': 'driverId'},
    {'1': 'vehicle_id', '3': 6, '4': 1, '5': 9, '10': 'vehicleId'},
    {'1': 'assigned_by', '3': 7, '4': 1, '5': 9, '10': 'assignedBy'},
    {'1': 'status', '3': 8, '4': 1, '5': 9, '10': 'status'},
    {'1': 'version', '3': 9, '4': 1, '5': 9, '10': 'version'},
    {'1': 'created_at', '3': 10, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'updated_at', '3': 11, '4': 1, '5': 9, '10': 'updatedAt'},
  ],
};

/// Descriptor for `FleetDelivery`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List fleetDeliveryDescriptor = $convert.base64Decode(
    'Cg1GbGVldERlbGl2ZXJ5Eg4KAmlkGAEgASgJUgJpZBIdCgpjb21wYW55X2lkGAIgASgJUgljb2'
    '1wYW55SWQSIwoNZGVwYXJ0bWVudF9pZBgDIAEoCVIMZGVwYXJ0bWVudElkEhkKCHJvdXRlX2lk'
    'GAQgASgJUgdyb3V0ZUlkEhsKCWRyaXZlcl9pZBgFIAEoCVIIZHJpdmVySWQSHQoKdmVoaWNsZV'
    '9pZBgGIAEoCVIJdmVoaWNsZUlkEh8KC2Fzc2lnbmVkX2J5GAcgASgJUgphc3NpZ25lZEJ5EhYK'
    'BnN0YXR1cxgIIAEoCVIGc3RhdHVzEhgKB3ZlcnNpb24YCSABKAlSB3ZlcnNpb24SHQoKY3JlYX'
    'RlZF9hdBgKIAEoCVIJY3JlYXRlZEF0Eh0KCnVwZGF0ZWRfYXQYCyABKAlSCXVwZGF0ZWRBdA==');

@$core.Deprecated('Use assignDeliveryRequestDescriptor instead')
const AssignDeliveryRequest$json = {
  '1': 'AssignDeliveryRequest',
  '2': [
    {'1': 'route_id', '3': 1, '4': 1, '5': 9, '10': 'routeId'},
    {'1': 'driver_id', '3': 2, '4': 1, '5': 9, '10': 'driverId'},
    {'1': 'vehicle_id', '3': 3, '4': 1, '5': 9, '10': 'vehicleId'},
    {'1': 'version', '3': 4, '4': 1, '5': 9, '10': 'version'},
  ],
};

/// Descriptor for `AssignDeliveryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignDeliveryRequestDescriptor = $convert.base64Decode(
    'ChVBc3NpZ25EZWxpdmVyeVJlcXVlc3QSGQoIcm91dGVfaWQYASABKAlSB3JvdXRlSWQSGwoJZH'
    'JpdmVyX2lkGAIgASgJUghkcml2ZXJJZBIdCgp2ZWhpY2xlX2lkGAMgASgJUgl2ZWhpY2xlSWQS'
    'GAoHdmVyc2lvbhgEIAEoCVIHdmVyc2lvbg==');

@$core.Deprecated('Use assignDeliveryResponseDescriptor instead')
const AssignDeliveryResponse$json = {
  '1': 'AssignDeliveryResponse',
  '2': [
    {
      '1': 'delivery',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.salesorder.v1.FleetDelivery',
      '10': 'delivery'
    },
  ],
};

/// Descriptor for `AssignDeliveryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignDeliveryResponseDescriptor =
    $convert.base64Decode(
        'ChZBc3NpZ25EZWxpdmVyeVJlc3BvbnNlEjgKCGRlbGl2ZXJ5GAEgASgLMhwuc2FsZXNvcmRlci'
        '52MS5GbGVldERlbGl2ZXJ5UghkZWxpdmVyeQ==');

@$core.Deprecated('Use listMyDeliveriesRequestDescriptor instead')
const ListMyDeliveriesRequest$json = {
  '1': 'ListMyDeliveriesRequest',
  '2': [
    {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListMyDeliveriesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMyDeliveriesRequestDescriptor =
    $convert.base64Decode(
        'ChdMaXN0TXlEZWxpdmVyaWVzUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhsKCXBhZ2Vfc2'
        'l6ZRgCIAEoBVIIcGFnZVNpemU=');

@$core.Deprecated('Use listMyDeliveriesResponseDescriptor instead')
const ListMyDeliveriesResponse$json = {
  '1': 'ListMyDeliveriesResponse',
  '2': [
    {
      '1': 'deliveries',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.FleetDelivery',
      '10': 'deliveries'
    },
    {'1': 'total', '3': 2, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListMyDeliveriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMyDeliveriesResponseDescriptor = $convert.base64Decode(
    'ChhMaXN0TXlEZWxpdmVyaWVzUmVzcG9uc2USPAoKZGVsaXZlcmllcxgBIAMoCzIcLnNhbGVzb3'
    'JkZXIudjEuRmxlZXREZWxpdmVyeVIKZGVsaXZlcmllcxIUCgV0b3RhbBgCIAEoBVIFdG90YWw=');

const $core.Map<$core.String, $core.dynamic> FleetServiceBase$json = {
  '1': 'FleetService',
  '2': [
    {
      '1': 'CreateDriver',
      '2': '.salesorder.v1.CreateDriverRequest',
      '3': '.salesorder.v1.CreateDriverResponse'
    },
    {
      '1': 'CreateVehicle',
      '2': '.salesorder.v1.CreateVehicleRequest',
      '3': '.salesorder.v1.CreateVehicleResponse'
    },
    {
      '1': 'AssignDelivery',
      '2': '.salesorder.v1.AssignDeliveryRequest',
      '3': '.salesorder.v1.AssignDeliveryResponse'
    },
    {
      '1': 'ListMyDeliveries',
      '2': '.salesorder.v1.ListMyDeliveriesRequest',
      '3': '.salesorder.v1.ListMyDeliveriesResponse'
    },
  ],
};

@$core.Deprecated('Use fleetServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    FleetServiceBase$messageJson = {
  '.salesorder.v1.CreateDriverRequest': CreateDriverRequest$json,
  '.salesorder.v1.CreateDriverResponse': CreateDriverResponse$json,
  '.salesorder.v1.FleetDriver': FleetDriver$json,
  '.salesorder.v1.CreateVehicleRequest': CreateVehicleRequest$json,
  '.salesorder.v1.CreateVehicleResponse': CreateVehicleResponse$json,
  '.salesorder.v1.Vehicle': Vehicle$json,
  '.salesorder.v1.AssignDeliveryRequest': AssignDeliveryRequest$json,
  '.salesorder.v1.AssignDeliveryResponse': AssignDeliveryResponse$json,
  '.salesorder.v1.FleetDelivery': FleetDelivery$json,
  '.salesorder.v1.ListMyDeliveriesRequest': ListMyDeliveriesRequest$json,
  '.salesorder.v1.ListMyDeliveriesResponse': ListMyDeliveriesResponse$json,
};

/// Descriptor for `FleetService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List fleetServiceDescriptor = $convert.base64Decode(
    'CgxGbGVldFNlcnZpY2USVwoMQ3JlYXRlRHJpdmVyEiIuc2FsZXNvcmRlci52MS5DcmVhdGVEcm'
    'l2ZXJSZXF1ZXN0GiMuc2FsZXNvcmRlci52MS5DcmVhdGVEcml2ZXJSZXNwb25zZRJaCg1DcmVh'
    'dGVWZWhpY2xlEiMuc2FsZXNvcmRlci52MS5DcmVhdGVWZWhpY2xlUmVxdWVzdBokLnNhbGVzb3'
    'JkZXIudjEuQ3JlYXRlVmVoaWNsZVJlc3BvbnNlEl0KDkFzc2lnbkRlbGl2ZXJ5EiQuc2FsZXNv'
    'cmRlci52MS5Bc3NpZ25EZWxpdmVyeVJlcXVlc3QaJS5zYWxlc29yZGVyLnYxLkFzc2lnbkRlbG'
    'l2ZXJ5UmVzcG9uc2USYwoQTGlzdE15RGVsaXZlcmllcxImLnNhbGVzb3JkZXIudjEuTGlzdE15'
    'RGVsaXZlcmllc1JlcXVlc3QaJy5zYWxlc29yZGVyLnYxLkxpc3RNeURlbGl2ZXJpZXNSZXNwb2'
    '5zZQ==');
