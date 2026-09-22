// This is a generated file - do not edit.
//
// Generated from salesorder/v1/devices.proto.

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

@$core.Deprecated('Use registerDeviceRequestDescriptor instead')
const RegisterDeviceRequest$json = {
  '1': 'RegisterDeviceRequest',
  '2': [
    {'1': 'platform', '3': 1, '4': 1, '5': 9, '10': 'platform'},
    {'1': 'fcm_token', '3': 2, '4': 1, '5': 9, '10': 'fcmToken'},
    {'1': 'device_name', '3': 3, '4': 1, '5': 9, '10': 'deviceName'},
  ],
};

/// Descriptor for `RegisterDeviceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerDeviceRequestDescriptor = $convert.base64Decode(
    'ChVSZWdpc3RlckRldmljZVJlcXVlc3QSGgoIcGxhdGZvcm0YASABKAlSCHBsYXRmb3JtEhsKCW'
    'ZjbV90b2tlbhgCIAEoCVIIZmNtVG9rZW4SHwoLZGV2aWNlX25hbWUYAyABKAlSCmRldmljZU5h'
    'bWU=');

@$core.Deprecated('Use registerDeviceResponseDescriptor instead')
const RegisterDeviceResponse$json = {
  '1': 'RegisterDeviceResponse',
  '2': [
    {'1': 'device_id', '3': 1, '4': 1, '5': 9, '10': 'deviceId'},
  ],
};

/// Descriptor for `RegisterDeviceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerDeviceResponseDescriptor =
    $convert.base64Decode(
        'ChZSZWdpc3RlckRldmljZVJlc3BvbnNlEhsKCWRldmljZV9pZBgBIAEoCVIIZGV2aWNlSWQ=');

@$core.Deprecated('Use unregisterDeviceRequestDescriptor instead')
const UnregisterDeviceRequest$json = {
  '1': 'UnregisterDeviceRequest',
  '2': [
    {'1': 'fcm_token', '3': 1, '4': 1, '5': 9, '10': 'fcmToken'},
  ],
};

/// Descriptor for `UnregisterDeviceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List unregisterDeviceRequestDescriptor =
    $convert.base64Decode(
        'ChdVbnJlZ2lzdGVyRGV2aWNlUmVxdWVzdBIbCglmY21fdG9rZW4YASABKAlSCGZjbVRva2Vu');

@$core.Deprecated('Use unregisterDeviceResponseDescriptor instead')
const UnregisterDeviceResponse$json = {
  '1': 'UnregisterDeviceResponse',
};

/// Descriptor for `UnregisterDeviceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List unregisterDeviceResponseDescriptor =
    $convert.base64Decode('ChhVbnJlZ2lzdGVyRGV2aWNlUmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> DeviceServiceBase$json = {
  '1': 'DeviceService',
  '2': [
    {
      '1': 'RegisterDevice',
      '2': '.salesorder.v1.RegisterDeviceRequest',
      '3': '.salesorder.v1.RegisterDeviceResponse'
    },
    {
      '1': 'UnregisterDevice',
      '2': '.salesorder.v1.UnregisterDeviceRequest',
      '3': '.salesorder.v1.UnregisterDeviceResponse'
    },
  ],
};

@$core.Deprecated('Use deviceServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DeviceServiceBase$messageJson = {
  '.salesorder.v1.RegisterDeviceRequest': RegisterDeviceRequest$json,
  '.salesorder.v1.RegisterDeviceResponse': RegisterDeviceResponse$json,
  '.salesorder.v1.UnregisterDeviceRequest': UnregisterDeviceRequest$json,
  '.salesorder.v1.UnregisterDeviceResponse': UnregisterDeviceResponse$json,
};

/// Descriptor for `DeviceService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List deviceServiceDescriptor = $convert.base64Decode(
    'Cg1EZXZpY2VTZXJ2aWNlEl0KDlJlZ2lzdGVyRGV2aWNlEiQuc2FsZXNvcmRlci52MS5SZWdpc3'
    'RlckRldmljZVJlcXVlc3QaJS5zYWxlc29yZGVyLnYxLlJlZ2lzdGVyRGV2aWNlUmVzcG9uc2US'
    'YwoQVW5yZWdpc3RlckRldmljZRImLnNhbGVzb3JkZXIudjEuVW5yZWdpc3RlckRldmljZVJlcX'
    'Vlc3QaJy5zYWxlc29yZGVyLnYxLlVucmVnaXN0ZXJEZXZpY2VSZXNwb25zZQ==');
