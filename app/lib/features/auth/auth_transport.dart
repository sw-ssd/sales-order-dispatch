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

/// 建立「帶身分」的 Connect transport：
/// 每請求附加 `Authorization: Bearer <access>`；收到 unauthenticated 時
/// 單飛 refresh 一次並重試該請求（refresh 仍失敗 → 原樣拋出，由頁面導回登入）。
///
/// [accessToken] 回目前 access token；[refresh] 為單飛的 refresh（見
/// AuthRepository.refreshTokens）。refresh 自身走未認證 transport，不經過本
/// 攔截器，避免自我遞迴。
connect.Transport createAuthedTransport({
  required String baseUrl,
  required Future<String?> Function() accessToken,
  required Future<bool> Function() refresh,
}) {
  return connect_protocol.Transport(
    baseUrl: baseUrl,
    codec: const ProtoCodec(),
    httpClient: connect_io.createHttpClient(io.HttpClient()),
    interceptors: [
      bearerInterceptor(accessToken: accessToken, refresh: refresh),
    ],
  );
}

/// Bearer 攔截器：逐請求附加 access token；收到 unauthenticated 時
/// 以 [refresh] 單飛換票成功後重試該請求一次（refresh 失敗 → 原樣拋出）。
connect.Interceptor bearerInterceptor({
  required Future<String?> Function() accessToken,
  required Future<bool> Function() refresh,
}) =>
    <I extends Object, O extends Object>(connect.AnyFn<I, O> chain) {
      return (connect.Request<I, O> request) async {
        Future<void> applyToken() async {
          final token = await accessToken();
          if (token != null && token.isNotEmpty) {
            request.headers.set('authorization', ['Bearer $token']);
          }
        }

        await applyToken();
        try {
          return await chain(request);
        } on connect.ConnectException catch (e) {
          if (e.code != connect.Code.unauthenticated || !await refresh()) {
            rethrow;
          }
          await applyToken();
          return await chain(request);
        }
      };
    };
