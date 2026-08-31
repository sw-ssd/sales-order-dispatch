import 'dart:convert';
import 'dart:io';
import 'dart:math';

import 'package:connectrpc/connect.dart' as connect;
import 'package:crypto/crypto.dart';
import 'package:flutter/services.dart';
import 'package:flutter_web_auth_2/flutter_web_auth_2.dart';

import '../../gen/salesorder/v1/auth.connect.client.dart';
import '../../gen/salesorder/v1/auth.pb.dart';
import 'auth_config.dart';
import 'token_storage.dart';

/// 產生 PKCE code verifier(RFC 7636:43–128 字元,unreserved 字元集)。
String generatePkceCodeVerifier({Random? random, int length = 64}) {
  final rng = random ?? Random.secure();
  const chars =
      'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  return List.generate(length, (_) => chars[rng.nextInt(chars.length)]).join();
}

/// PKCE S256 code challenge:base64url(無 padding)的 SHA-256(verifier)。
String pkceCodeChallengeS256(String codeVerifier) {
  final digest = sha256.convert(utf8.encode(codeVerifier));
  return base64UrlEncode(digest.bytes).replaceAll('=', '');
}

/// Google 授權完成取得的授權碼與對應的 PKCE verifier。
///
/// 後端換票端點未定([AuthConfig.backendGoogleExchangePath] 為預留常數),
/// 端點上線後以此二者向後端換發本系統 token 對。
class GoogleAuthorizationGrant {
  const GoogleAuthorizationGrant({
    required this.code,
    required this.codeVerifier,
  });

  final String code;
  final String codeVerifier;
}

/// 認證流程:店家帳密登入(AuthService.Login)+ 業務 Google PKCE 登入。
class AuthRepository {
  AuthRepository({
    required AuthServiceClient client,
    required TokenStorage tokenStorage,
    Future<String> Function({required String url, required String callbackUrlScheme})?
        launchAuthSession,
  })  : _client = client,
        _tokenStorage = tokenStorage,
        _launchAuthSession = launchAuthSession ?? FlutterWebAuth2.authenticate;

  final AuthServiceClient _client;
  final TokenStorage _tokenStorage;
  final Future<String> Function({
    required String url,
    required String callbackUrlScheme,
  }) _launchAuthSession;

  /// 店家登入:customer_code + password → AuthService.Login,成功後存 token。
  Future<AuthTokenPair> loginShop({
    required String customerCode,
    required String password,
  }) async {
    final response = await _client.login(
      LoginRequest(customerCode: customerCode, password: password),
    );
    final tokens = AuthTokenPair(
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
    );
    await _tokenStorage.save(tokens);
    return tokens;
  }

  /// 業務 Google 登入:開啟系統瀏覽器走 PKCE 授權。
  ///
  /// 回傳授權碼與 verifier(待後端端點上線後換票);使用者取消時回傳 null。
  /// [AuthConfig.googleClientId] 未設定時拋 [StateError](頁面應提示尚未設定)。
  Future<GoogleAuthorizationGrant?> loginSalesWithGoogle() async {
    if (AuthConfig.googleClientId.isEmpty) {
      throw StateError('GOOGLE_CLIENT_ID 未設定,無法發起 Google 登入');
    }
    final codeVerifier = generatePkceCodeVerifier();
    final authorizationUrl = Uri.https(
      'accounts.google.com',
      '/o/oauth2/v2/auth',
      {
        'client_id': AuthConfig.googleClientId,
        'redirect_uri': AuthConfig.callbackUrl,
        'response_type': 'code',
        'scope': AuthConfig.googleScopes.join(' '),
        'code_challenge': pkceCodeChallengeS256(codeVerifier),
        'code_challenge_method': 'S256',
      },
    );
    String callback;
    try {
      callback = await _launchAuthSession(
        url: authorizationUrl.toString(),
        callbackUrlScheme: AuthConfig.callbackUrlScheme,
      );
    } on PlatformException catch (e) {
      if (e.code == 'CANCELED') return null; // 使用者關閉授權頁
      rethrow;
    }
    final code = Uri.parse(callback).queryParameters['code'];
    if (code == null || code.isEmpty) {
      throw StateError('Google 授權回調缺少 authorization code');
    }
    return GoogleAuthorizationGrant(code: code, codeVerifier: codeVerifier);
  }
}

/// 把登入相關錯誤轉為使用者可讀訊息(繁體中文)。
///
/// 後端 AuthService 尚未有 server 實作(T11–14 延後),`unimplemented`
/// 與連線失敗為目前最常見結果,皆需妥善提示。
String authErrorMessage(Object error) {
  if (error is connect.ConnectException) {
    return switch (error.code) {
      connect.Code.unimplemented => '伺服器尚未支援登入功能,請稍後再試',
      connect.Code.unavailable => '無法連線至伺服器,請檢查網路後再試',
      connect.Code.unauthenticated => '客戶編號或密碼錯誤,請重新輸入',
      connect.Code.deadlineExceeded => '連線逾時,請稍後再試',
      _ => '登入失敗(${error.code.name}),請稍後再試',
    };
  }
  if (error is SocketException) return '無法連線至伺服器,請檢查網路後再試';
  if (error is StateError) return error.message;
  return '登入失敗,請稍後再試';
}
