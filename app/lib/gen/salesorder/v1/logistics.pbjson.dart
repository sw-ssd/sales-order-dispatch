// This is a generated file - do not edit.
//
// Generated from salesorder/v1/logistics.proto.

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

@$core.Deprecated('Use logisticsDriverDescriptor instead')
const LogisticsDriver$json = {
  '1': 'LogisticsDriver',
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

/// Descriptor for `LogisticsDriver`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logisticsDriverDescriptor = $convert.base64Decode(
    'Cg9Mb2dpc3RpY3NEcml2ZXISDgoCaWQYASABKAlSAmlkEh0KCmNvbXBhbnlfaWQYAiABKAlSCW'
    'NvbXBhbnlJZBIjCg1kZXBhcnRtZW50X2lkGAMgASgJUgxkZXBhcnRtZW50SWQSFwoHdXNlcl9p'
    'ZBgEIAEoCVIGdXNlcklkEhIKBG5hbWUYBSABKAlSBG5hbWUSFAoFcGhvbmUYBiABKAlSBXBob2'
    '5lEiUKDmN1cnJlbnRfc3RhdHVzGAcgASgJUg1jdXJyZW50U3RhdHVz');

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
      '6': '.salesorder.v1.LogisticsDriver',
      '10': 'driver'
    },
  ],
};

/// Descriptor for `CreateDriverResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDriverResponseDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVEcml2ZXJSZXNwb25zZRI2CgZkcml2ZXIYASABKAsyHi5zYWxlc29yZGVyLnYxLk'
    'xvZ2lzdGljc0RyaXZlclIGZHJpdmVy');

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

@$core.Deprecated('Use logisticsDeliveryDescriptor instead')
const LogisticsDelivery$json = {
  '1': 'LogisticsDelivery',
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

/// Descriptor for `LogisticsDelivery`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logisticsDeliveryDescriptor = $convert.base64Decode(
    'ChFMb2dpc3RpY3NEZWxpdmVyeRIOCgJpZBgBIAEoCVICaWQSHQoKY29tcGFueV9pZBgCIAEoCV'
    'IJY29tcGFueUlkEiMKDWRlcGFydG1lbnRfaWQYAyABKAlSDGRlcGFydG1lbnRJZBIZCghyb3V0'
    'ZV9pZBgEIAEoCVIHcm91dGVJZBIbCglkcml2ZXJfaWQYBSABKAlSCGRyaXZlcklkEh0KCnZlaG'
    'ljbGVfaWQYBiABKAlSCXZlaGljbGVJZBIfCgthc3NpZ25lZF9ieRgHIAEoCVIKYXNzaWduZWRC'
    'eRIWCgZzdGF0dXMYCCABKAlSBnN0YXR1cxIYCgd2ZXJzaW9uGAkgASgJUgd2ZXJzaW9uEh0KCm'
    'NyZWF0ZWRfYXQYCiABKAlSCWNyZWF0ZWRBdBIdCgp1cGRhdGVkX2F0GAsgASgJUgl1cGRhdGVk'
    'QXQ=');

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
      '6': '.salesorder.v1.LogisticsDelivery',
      '10': 'delivery'
    },
  ],
};

/// Descriptor for `AssignDeliveryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignDeliveryResponseDescriptor =
    $convert.base64Decode(
        'ChZBc3NpZ25EZWxpdmVyeVJlc3BvbnNlEjwKCGRlbGl2ZXJ5GAEgASgLMiAuc2FsZXNvcmRlci'
        '52MS5Mb2dpc3RpY3NEZWxpdmVyeVIIZGVsaXZlcnk=');

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
      '6': '.salesorder.v1.LogisticsDelivery',
      '10': 'deliveries'
    },
    {'1': 'total', '3': 2, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListMyDeliveriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMyDeliveriesResponseDescriptor = $convert.base64Decode(
    'ChhMaXN0TXlEZWxpdmVyaWVzUmVzcG9uc2USQAoKZGVsaXZlcmllcxgBIAMoCzIgLnNhbGVzb3'
    'JkZXIudjEuTG9naXN0aWNzRGVsaXZlcnlSCmRlbGl2ZXJpZXMSFAoFdG90YWwYAiABKAVSBXRv'
    'dGFs');

const $core.Map<$core.String, $core.dynamic> LogisticsServiceBase$json = {
  '1': 'LogisticsService',
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

@$core.Deprecated('Use logisticsServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    LogisticsServiceBase$messageJson = {
  '.salesorder.v1.CreateDriverRequest': CreateDriverRequest$json,
  '.salesorder.v1.CreateDriverResponse': CreateDriverResponse$json,
  '.salesorder.v1.LogisticsDriver': LogisticsDriver$json,
  '.salesorder.v1.CreateVehicleRequest': CreateVehicleRequest$json,
  '.salesorder.v1.CreateVehicleResponse': CreateVehicleResponse$json,
  '.salesorder.v1.Vehicle': Vehicle$json,
  '.salesorder.v1.AssignDeliveryRequest': AssignDeliveryRequest$json,
  '.salesorder.v1.AssignDeliveryResponse': AssignDeliveryResponse$json,
  '.salesorder.v1.LogisticsDelivery': LogisticsDelivery$json,
  '.salesorder.v1.ListMyDeliveriesRequest': ListMyDeliveriesRequest$json,
  '.salesorder.v1.ListMyDeliveriesResponse': ListMyDeliveriesResponse$json,
};

/// Descriptor for `LogisticsService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List logisticsServiceDescriptor = $convert.base64Decode(
    'ChBMb2dpc3RpY3NTZXJ2aWNlElcKDENyZWF0ZURyaXZlchIiLnNhbGVzb3JkZXIudjEuQ3JlYX'
    'RlRHJpdmVyUmVxdWVzdBojLnNhbGVzb3JkZXIudjEuQ3JlYXRlRHJpdmVyUmVzcG9uc2USWgoN'
    'Q3JlYXRlVmVoaWNsZRIjLnNhbGVzb3JkZXIudjEuQ3JlYXRlVmVoaWNsZVJlcXVlc3QaJC5zYW'
    'xlc29yZGVyLnYxLkNyZWF0ZVZlaGljbGVSZXNwb25zZRJdCg5Bc3NpZ25EZWxpdmVyeRIkLnNh'
    'bGVzb3JkZXIudjEuQXNzaWduRGVsaXZlcnlSZXF1ZXN0GiUuc2FsZXNvcmRlci52MS5Bc3NpZ2'
    '5EZWxpdmVyeVJlc3BvbnNlEmMKEExpc3RNeURlbGl2ZXJpZXMSJi5zYWxlc29yZGVyLnYxLkxp'
    'c3RNeURlbGl2ZXJpZXNSZXF1ZXN0Gicuc2FsZXNvcmRlci52MS5MaXN0TXlEZWxpdmVyaWVzUm'
    'VzcG9uc2U=');
