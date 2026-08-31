// This is a generated file - do not edit.
//
// Generated from salesorder/v1/ability.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../google/protobuf/struct.pbjson.dart' as $0;

@$core.Deprecated('Use abilityRuleDescriptor instead')
const AbilityRule$json = {
  '1': 'AbilityRule',
  '2': [
    {'1': 'action', '3': 1, '4': 1, '5': 9, '10': 'action'},
    {'1': 'subject', '3': 2, '4': 1, '5': 9, '10': 'subject'},
    {
      '1': 'conditions',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'conditions'
    },
    {'1': 'inverted', '3': 4, '4': 1, '5': 8, '10': 'inverted'},
  ],
};

/// Descriptor for `AbilityRule`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List abilityRuleDescriptor = $convert.base64Decode(
    'CgtBYmlsaXR5UnVsZRIWCgZhY3Rpb24YASABKAlSBmFjdGlvbhIYCgdzdWJqZWN0GAIgASgJUg'
    'dzdWJqZWN0EjcKCmNvbmRpdGlvbnMYAyABKAsyFy5nb29nbGUucHJvdG9idWYuU3RydWN0Ugpj'
    'b25kaXRpb25zEhoKCGludmVydGVkGAQgASgIUghpbnZlcnRlZA==');

@$core.Deprecated('Use getAbilityRequestDescriptor instead')
const GetAbilityRequest$json = {
  '1': 'GetAbilityRequest',
};

/// Descriptor for `GetAbilityRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAbilityRequestDescriptor =
    $convert.base64Decode('ChFHZXRBYmlsaXR5UmVxdWVzdA==');

@$core.Deprecated('Use getAbilityResponseDescriptor instead')
const GetAbilityResponse$json = {
  '1': 'GetAbilityResponse',
  '2': [
    {
      '1': 'rules',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.AbilityRule',
      '10': 'rules'
    },
  ],
};

/// Descriptor for `GetAbilityResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAbilityResponseDescriptor = $convert.base64Decode(
    'ChJHZXRBYmlsaXR5UmVzcG9uc2USMAoFcnVsZXMYASADKAsyGi5zYWxlc29yZGVyLnYxLkFiaW'
    'xpdHlSdWxlUgVydWxlcw==');

const $core.Map<$core.String, $core.dynamic> AbilityServiceBase$json = {
  '1': 'AbilityService',
  '2': [
    {
      '1': 'GetAbility',
      '2': '.salesorder.v1.GetAbilityRequest',
      '3': '.salesorder.v1.GetAbilityResponse'
    },
  ],
};

@$core.Deprecated('Use abilityServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AbilityServiceBase$messageJson = {
  '.salesorder.v1.GetAbilityRequest': GetAbilityRequest$json,
  '.salesorder.v1.GetAbilityResponse': GetAbilityResponse$json,
  '.salesorder.v1.AbilityRule': AbilityRule$json,
  '.google.protobuf.Struct': $0.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $0.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $0.Value$json,
  '.google.protobuf.ListValue': $0.ListValue$json,
};

/// Descriptor for `AbilityService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List abilityServiceDescriptor = $convert.base64Decode(
    'Cg5BYmlsaXR5U2VydmljZRJRCgpHZXRBYmlsaXR5EiAuc2FsZXNvcmRlci52MS5HZXRBYmlsaX'
    'R5UmVxdWVzdBohLnNhbGVzb3JkZXIudjEuR2V0QWJpbGl0eVJlc3BvbnNl');
