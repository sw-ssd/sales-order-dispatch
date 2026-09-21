// This is a generated file - do not edit.
//
// Generated from salesorder/v1/auth.proto.

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

@$core.Deprecated('Use changePasswordRequestDescriptor instead')
const ChangePasswordRequest$json = {
  '1': 'ChangePasswordRequest',
  '2': [
    {'1': 'old_password', '3': 1, '4': 1, '5': 9, '10': 'oldPassword'},
    {'1': 'new_password', '3': 2, '4': 1, '5': 9, '10': 'newPassword'},
  ],
};

/// Descriptor for `ChangePasswordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePasswordRequestDescriptor = $convert.base64Decode(
    'ChVDaGFuZ2VQYXNzd29yZFJlcXVlc3QSIQoMb2xkX3Bhc3N3b3JkGAEgASgJUgtvbGRQYXNzd2'
    '9yZBIhCgxuZXdfcGFzc3dvcmQYAiABKAlSC25ld1Bhc3N3b3Jk');

@$core.Deprecated('Use changePasswordResponseDescriptor instead')
const ChangePasswordResponse$json = {
  '1': 'ChangePasswordResponse',
};

/// Descriptor for `ChangePasswordResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePasswordResponseDescriptor =
    $convert.base64Decode('ChZDaGFuZ2VQYXNzd29yZFJlc3BvbnNl');

@$core.Deprecated('Use resetCustomerPasswordRequestDescriptor instead')
const ResetCustomerPasswordRequest$json = {
  '1': 'ResetCustomerPasswordRequest',
  '2': [
    {'1': 'user_id', '3': 1, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `ResetCustomerPasswordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resetCustomerPasswordRequestDescriptor =
    $convert.base64Decode(
        'ChxSZXNldEN1c3RvbWVyUGFzc3dvcmRSZXF1ZXN0EhcKB3VzZXJfaWQYASABKAlSBnVzZXJJZA'
        '==');

@$core.Deprecated('Use resetCustomerPasswordResponseDescriptor instead')
const ResetCustomerPasswordResponse$json = {
  '1': 'ResetCustomerPasswordResponse',
  '2': [
    {'1': 'temp_password', '3': 1, '4': 1, '5': 9, '10': 'tempPassword'},
    {'1': 'expires_at', '3': 2, '4': 1, '5': 3, '10': 'expiresAt'},
  ],
};

/// Descriptor for `ResetCustomerPasswordResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resetCustomerPasswordResponseDescriptor =
    $convert.base64Decode(
        'Ch1SZXNldEN1c3RvbWVyUGFzc3dvcmRSZXNwb25zZRIjCg10ZW1wX3Bhc3N3b3JkGAEgASgJUg'
        'x0ZW1wUGFzc3dvcmQSHQoKZXhwaXJlc19hdBgCIAEoA1IJZXhwaXJlc0F0');

@$core.Deprecated('Use loginRequestDescriptor instead')
const LoginRequest$json = {
  '1': 'LoginRequest',
  '2': [
    {'1': 'customer_code', '3': 1, '4': 1, '5': 9, '10': 'customerCode'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '10': 'password'},
  ],
};

/// Descriptor for `LoginRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginRequestDescriptor = $convert.base64Decode(
    'CgxMb2dpblJlcXVlc3QSIwoNY3VzdG9tZXJfY29kZRgBIAEoCVIMY3VzdG9tZXJDb2RlEhoKCH'
    'Bhc3N3b3JkGAIgASgJUghwYXNzd29yZA==');

@$core.Deprecated('Use loginResponseDescriptor instead')
const LoginResponse$json = {
  '1': 'LoginResponse',
  '2': [
    {'1': 'access_token', '3': 1, '4': 1, '5': 9, '10': 'accessToken'},
    {'1': 'refresh_token', '3': 2, '4': 1, '5': 9, '10': 'refreshToken'},
    {'1': 'expires_in', '3': 3, '4': 1, '5': 3, '10': 'expiresIn'},
    {
      '1': 'must_change_password',
      '3': 4,
      '4': 1,
      '5': 8,
      '10': 'mustChangePassword'
    },
  ],
};

/// Descriptor for `LoginResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginResponseDescriptor = $convert.base64Decode(
    'Cg1Mb2dpblJlc3BvbnNlEiEKDGFjY2Vzc190b2tlbhgBIAEoCVILYWNjZXNzVG9rZW4SIwoNcm'
    'VmcmVzaF90b2tlbhgCIAEoCVIMcmVmcmVzaFRva2VuEh0KCmV4cGlyZXNfaW4YAyABKANSCWV4'
    'cGlyZXNJbhIwChRtdXN0X2NoYW5nZV9wYXNzd29yZBgEIAEoCFISbXVzdENoYW5nZVBhc3N3b3'
    'Jk');

@$core.Deprecated('Use refreshRequestDescriptor instead')
const RefreshRequest$json = {
  '1': 'RefreshRequest',
  '2': [
    {'1': 'refresh_token', '3': 1, '4': 1, '5': 9, '10': 'refreshToken'},
  ],
};

/// Descriptor for `RefreshRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List refreshRequestDescriptor = $convert.base64Decode(
    'Cg5SZWZyZXNoUmVxdWVzdBIjCg1yZWZyZXNoX3Rva2VuGAEgASgJUgxyZWZyZXNoVG9rZW4=');

@$core.Deprecated('Use refreshResponseDescriptor instead')
const RefreshResponse$json = {
  '1': 'RefreshResponse',
  '2': [
    {'1': 'access_token', '3': 1, '4': 1, '5': 9, '10': 'accessToken'},
    {'1': 'refresh_token', '3': 2, '4': 1, '5': 9, '10': 'refreshToken'},
    {'1': 'expires_in', '3': 3, '4': 1, '5': 3, '10': 'expiresIn'},
  ],
};

/// Descriptor for `RefreshResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List refreshResponseDescriptor = $convert.base64Decode(
    'Cg9SZWZyZXNoUmVzcG9uc2USIQoMYWNjZXNzX3Rva2VuGAEgASgJUgthY2Nlc3NUb2tlbhIjCg'
    '1yZWZyZXNoX3Rva2VuGAIgASgJUgxyZWZyZXNoVG9rZW4SHQoKZXhwaXJlc19pbhgDIAEoA1IJ'
    'ZXhwaXJlc0lu');

@$core.Deprecated('Use logoutRequestDescriptor instead')
const LogoutRequest$json = {
  '1': 'LogoutRequest',
  '2': [
    {'1': 'refresh_token', '3': 1, '4': 1, '5': 9, '10': 'refreshToken'},
  ],
};

/// Descriptor for `LogoutRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logoutRequestDescriptor = $convert.base64Decode(
    'Cg1Mb2dvdXRSZXF1ZXN0EiMKDXJlZnJlc2hfdG9rZW4YASABKAlSDHJlZnJlc2hUb2tlbg==');

@$core.Deprecated('Use logoutResponseDescriptor instead')
const LogoutResponse$json = {
  '1': 'LogoutResponse',
};

/// Descriptor for `LogoutResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logoutResponseDescriptor =
    $convert.base64Decode('Cg5Mb2dvdXRSZXNwb25zZQ==');

@$core.Deprecated('Use registerCompleteRequestDescriptor instead')
const RegisterCompleteRequest$json = {
  '1': 'RegisterCompleteRequest',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `RegisterCompleteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerCompleteRequestDescriptor =
    $convert.base64Decode(
        'ChdSZWdpc3RlckNvbXBsZXRlUmVxdWVzdBIdCgpjb21wYW55X2lkGAEgASgJUgljb21wYW55SW'
        'QSEgoEbmFtZRgCIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use registerCompleteResponseDescriptor instead')
const RegisterCompleteResponse$json = {
  '1': 'RegisterCompleteResponse',
};

/// Descriptor for `RegisterCompleteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerCompleteResponseDescriptor =
    $convert.base64Decode('ChhSZWdpc3RlckNvbXBsZXRlUmVzcG9uc2U=');

@$core.Deprecated('Use qRLoginRequestDescriptor instead')
const QRLoginRequest$json = {
  '1': 'QRLoginRequest',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
  ],
};

/// Descriptor for `QRLoginRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List qRLoginRequestDescriptor = $convert
    .base64Decode('Cg5RUkxvZ2luUmVxdWVzdBIUCgV0b2tlbhgBIAEoCVIFdG9rZW4=');

@$core.Deprecated('Use qRLoginResponseDescriptor instead')
const QRLoginResponse$json = {
  '1': 'QRLoginResponse',
  '2': [
    {'1': 'company_id', '3': 1, '4': 1, '5': 9, '10': 'companyId'},
    {'1': 'company_name', '3': 2, '4': 1, '5': 9, '10': 'companyName'},
    {'1': 'customer_code', '3': 3, '4': 1, '5': 9, '10': 'customerCode'},
    {'1': 'customer_name', '3': 4, '4': 1, '5': 9, '10': 'customerName'},
    {
      '1': 'accounts',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.salesorder.v1.QRLoginResponse.Account',
      '10': 'accounts'
    },
  ],
  '3': [QRLoginResponse_Account$json],
};

@$core.Deprecated('Use qRLoginResponseDescriptor instead')
const QRLoginResponse_Account$json = {
  '1': 'Account',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'account_name', '3': 2, '4': 1, '5': 9, '10': 'accountName'},
  ],
};

/// Descriptor for `QRLoginResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List qRLoginResponseDescriptor = $convert.base64Decode(
    'Cg9RUkxvZ2luUmVzcG9uc2USHQoKY29tcGFueV9pZBgBIAEoCVIJY29tcGFueUlkEiEKDGNvbX'
    'BhbnlfbmFtZRgCIAEoCVILY29tcGFueU5hbWUSIwoNY3VzdG9tZXJfY29kZRgDIAEoCVIMY3Vz'
    'dG9tZXJDb2RlEiMKDWN1c3RvbWVyX25hbWUYBCABKAlSDGN1c3RvbWVyTmFtZRJCCghhY2NvdW'
    '50cxgFIAMoCzImLnNhbGVzb3JkZXIudjEuUVJMb2dpblJlc3BvbnNlLkFjY291bnRSCGFjY291'
    'bnRzGjwKB0FjY291bnQSDgoCaWQYASABKAlSAmlkEiEKDGFjY291bnRfbmFtZRgCIAEoCVILYW'
    'Njb3VudE5hbWU=');

@$core.Deprecated('Use getCustomerQRCodeRequestDescriptor instead')
const GetCustomerQRCodeRequest$json = {
  '1': 'GetCustomerQRCodeRequest',
  '2': [
    {'1': 'customer_id', '3': 1, '4': 1, '5': 9, '10': 'customerId'},
  ],
};

/// Descriptor for `GetCustomerQRCodeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCustomerQRCodeRequestDescriptor =
    $convert.base64Decode(
        'ChhHZXRDdXN0b21lclFSQ29kZVJlcXVlc3QSHwoLY3VzdG9tZXJfaWQYASABKAlSCmN1c3RvbW'
        'VySWQ=');

@$core.Deprecated('Use getCustomerQRCodeResponseDescriptor instead')
const GetCustomerQRCodeResponse$json = {
  '1': 'GetCustomerQRCodeResponse',
  '2': [
    {'1': 'qr_url', '3': 1, '4': 1, '5': 9, '10': 'qrUrl'},
  ],
};

/// Descriptor for `GetCustomerQRCodeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCustomerQRCodeResponseDescriptor =
    $convert.base64Decode(
        'ChlHZXRDdXN0b21lclFSQ29kZVJlc3BvbnNlEhUKBnFyX3VybBgBIAEoCVIFcXJVcmw=');

const $core.Map<$core.String, $core.dynamic> AuthServiceBase$json = {
  '1': 'AuthService',
  '2': [
    {
      '1': 'Login',
      '2': '.salesorder.v1.LoginRequest',
      '3': '.salesorder.v1.LoginResponse'
    },
    {
      '1': 'Refresh',
      '2': '.salesorder.v1.RefreshRequest',
      '3': '.salesorder.v1.RefreshResponse'
    },
    {
      '1': 'Logout',
      '2': '.salesorder.v1.LogoutRequest',
      '3': '.salesorder.v1.LogoutResponse'
    },
    {
      '1': 'RegisterComplete',
      '2': '.salesorder.v1.RegisterCompleteRequest',
      '3': '.salesorder.v1.RegisterCompleteResponse'
    },
    {
      '1': 'QRLogin',
      '2': '.salesorder.v1.QRLoginRequest',
      '3': '.salesorder.v1.QRLoginResponse'
    },
    {
      '1': 'GetCustomerQRCode',
      '2': '.salesorder.v1.GetCustomerQRCodeRequest',
      '3': '.salesorder.v1.GetCustomerQRCodeResponse'
    },
    {
      '1': 'ChangePassword',
      '2': '.salesorder.v1.ChangePasswordRequest',
      '3': '.salesorder.v1.ChangePasswordResponse'
    },
    {
      '1': 'ResetCustomerPassword',
      '2': '.salesorder.v1.ResetCustomerPasswordRequest',
      '3': '.salesorder.v1.ResetCustomerPasswordResponse'
    },
  ],
};

@$core.Deprecated('Use authServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AuthServiceBase$messageJson = {
  '.salesorder.v1.LoginRequest': LoginRequest$json,
  '.salesorder.v1.LoginResponse': LoginResponse$json,
  '.salesorder.v1.RefreshRequest': RefreshRequest$json,
  '.salesorder.v1.RefreshResponse': RefreshResponse$json,
  '.salesorder.v1.LogoutRequest': LogoutRequest$json,
  '.salesorder.v1.LogoutResponse': LogoutResponse$json,
  '.salesorder.v1.RegisterCompleteRequest': RegisterCompleteRequest$json,
  '.salesorder.v1.RegisterCompleteResponse': RegisterCompleteResponse$json,
  '.salesorder.v1.QRLoginRequest': QRLoginRequest$json,
  '.salesorder.v1.QRLoginResponse': QRLoginResponse$json,
  '.salesorder.v1.QRLoginResponse.Account': QRLoginResponse_Account$json,
  '.salesorder.v1.GetCustomerQRCodeRequest': GetCustomerQRCodeRequest$json,
  '.salesorder.v1.GetCustomerQRCodeResponse': GetCustomerQRCodeResponse$json,
  '.salesorder.v1.ChangePasswordRequest': ChangePasswordRequest$json,
  '.salesorder.v1.ChangePasswordResponse': ChangePasswordResponse$json,
  '.salesorder.v1.ResetCustomerPasswordRequest':
      ResetCustomerPasswordRequest$json,
  '.salesorder.v1.ResetCustomerPasswordResponse':
      ResetCustomerPasswordResponse$json,
};

/// Descriptor for `AuthService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List authServiceDescriptor = $convert.base64Decode(
    'CgtBdXRoU2VydmljZRJCCgVMb2dpbhIbLnNhbGVzb3JkZXIudjEuTG9naW5SZXF1ZXN0Ghwuc2'
    'FsZXNvcmRlci52MS5Mb2dpblJlc3BvbnNlEkgKB1JlZnJlc2gSHS5zYWxlc29yZGVyLnYxLlJl'
    'ZnJlc2hSZXF1ZXN0Gh4uc2FsZXNvcmRlci52MS5SZWZyZXNoUmVzcG9uc2USRQoGTG9nb3V0Eh'
    'wuc2FsZXNvcmRlci52MS5Mb2dvdXRSZXF1ZXN0Gh0uc2FsZXNvcmRlci52MS5Mb2dvdXRSZXNw'
    'b25zZRJjChBSZWdpc3RlckNvbXBsZXRlEiYuc2FsZXNvcmRlci52MS5SZWdpc3RlckNvbXBsZX'
    'RlUmVxdWVzdBonLnNhbGVzb3JkZXIudjEuUmVnaXN0ZXJDb21wbGV0ZVJlc3BvbnNlEkgKB1FS'
    'TG9naW4SHS5zYWxlc29yZGVyLnYxLlFSTG9naW5SZXF1ZXN0Gh4uc2FsZXNvcmRlci52MS5RUk'
    'xvZ2luUmVzcG9uc2USZgoRR2V0Q3VzdG9tZXJRUkNvZGUSJy5zYWxlc29yZGVyLnYxLkdldEN1'
    'c3RvbWVyUVJDb2RlUmVxdWVzdBooLnNhbGVzb3JkZXIudjEuR2V0Q3VzdG9tZXJRUkNvZGVSZX'
    'Nwb25zZRJdCg5DaGFuZ2VQYXNzd29yZBIkLnNhbGVzb3JkZXIudjEuQ2hhbmdlUGFzc3dvcmRS'
    'ZXF1ZXN0GiUuc2FsZXNvcmRlci52MS5DaGFuZ2VQYXNzd29yZFJlc3BvbnNlEnIKFVJlc2V0Q3'
    'VzdG9tZXJQYXNzd29yZBIrLnNhbGVzb3JkZXIudjEuUmVzZXRDdXN0b21lclBhc3N3b3JkUmVx'
    'dWVzdBosLnNhbGVzb3JkZXIudjEuUmVzZXRDdXN0b21lclBhc3N3b3JkUmVzcG9uc2U=');
