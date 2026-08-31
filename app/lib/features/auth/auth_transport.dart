import 'dart:io' as io;

import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/io.dart' as connect_io;
import 'package:connectrpc/protobuf.dart';
import 'package:connectrpc/protocol/connect.dart' as connect_protocol;

import '../../gen/salesorder/v1/auth.connect.client.dart';

/// 建立「未帶身分」的 Connect transport:登入 / refresh / 註冊完成等公開 RPC 用,
/// 不附加 Bearer(D29:避免 refresh 自我遞迴;AuthInterceptor 屬後續任務)。
connect.Transport createUnauthenticatedTransport(String baseUrl) {
  return connect_protocol.Transport(
    baseUrl: baseUrl,
    codec: const ProtoCodec(),
    httpClient: connect_io.createHttpClient(io.HttpClient()),
  );
}

/// 以 AppConfig.apiBaseUrl 建立 AuthService 用戶端。
AuthServiceClient createAuthServiceClient(String baseUrl) {
  return AuthServiceClient(createUnauthenticatedTransport(baseUrl));
}
