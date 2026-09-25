import 'dart:convert';
import 'dart:math';

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

/// 店家登入結果：token 對 ＋ 是否處於「首登強改」受限態。
///
/// `mustChangePassword` 不是 token 的一部分（是帳號狀態），故不塞進 [AuthTokenPair]
/// （那是持久化型別，多一個不落盤的欄位只會誤導）。登入回應本來就帶這個欄位
/// （`LoginResponse.must_change_password`），丟掉它就只能靠試呼 API 猜。
class ShopLoginResult {
  const ShopLoginResult({
    required this.tokens,
    required this.mustChangePassword,
  });

  final AuthTokenPair tokens;
  final bool mustChangePassword;
}

/// 認證流程:店家帳密登入(AuthService.Login)+ 業務 Google PKCE 登入。
class AuthRepository {
  AuthRepository({
    required AuthServiceClient client,
    required AuthServiceClient authed,
    required TokenStorage tokenStorage,
    Future<String> Function({required String url, required String callbackUrlScheme})?
        launchAuthSession,
  })  : _client = client,
        _authed = authed,
        _tokenStorage = tokenStorage,
        _launchAuthSession = launchAuthSession ?? FlutterWebAuth2.authenticate;

  final AuthServiceClient _client;

  /// 帶身分的 AuthService 客戶端（Bearer + 401 單飛 refresh）。
  /// 目前只有 ChangePassword 用；其餘 AuthService RPC 都是登入前的公開端點。
  final AuthServiceClient _authed;
  final TokenStorage _tokenStorage;

  /// 登入成功後 / 登出前的掛鉤（由 `Api` 設為推播註冊/註銷）。
  ///
  /// 掛在這裡而不是各頁的登入/登出呼叫點：本檔有**兩個**登入呼叫端（一般登入、QR 登入）
  /// 與**三個**登出呼叫端（主殼、帳號管理、改密碼頁）；逐點加會讓下一個新增的呼叫端
  /// 靜默漏掉推播註冊。掛鉤為 nil 時完全不做事（測試與未接推播的環境）。
  Future<void> Function()? onSessionEstablished;
  Future<void> Function()? onSessionCleared;
  final Future<String> Function({
    required String url,
    required String callbackUrlScheme,
  }) _launchAuthSession;

  /// 店家登入:customer_code + password → AuthService.Login,成功後存 token。
  ///
  /// 回傳 `mustChangePassword`（A3 1.5.2）：true 代表這是臨時密碼首登，後端只放行
  /// ChangePassword。呼叫端據此導向改密碼頁 —— **不要靠試呼 API 探測**，登入回應
  /// 本來就帶這個欄位。
  Future<ShopLoginResult> loginShop({
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
    // 推播註冊：**不 await 失敗**（見 PushRegistration 的說明：缺 Firebase 原生設定檔時
    // 靜默跳過），也刻意不等它完成才回傳 —— 登入不該被推播拖慢。
    _runHook(onSessionEstablished);
    return ShopLoginResult(
      tokens: tokens,
      mustChangePassword: response.mustChangePassword,
    );
  }

  /// 執行掛鉤並吞掉例外：推播屬附加能力，任何失敗都不得阻斷登入/登出。
  void _runHook(Future<void> Function()? hook) {
    if (hook == null) return;
    hook().catchError((Object _) {});
  }

  /// 目前的 access token（無 token 回 null）；供已認證 transport 逐請求附加。
  Future<String?> currentAccessToken() async =>
      (await _tokenStorage.read())?.accessToken;

  Future<bool>? _refreshing;

  /// 以 refresh token 換新 token 對並落盤（單飛：併發呼叫共用同一次 refresh）。
  ///
  /// refresh 失敗（過期/已旋轉）→ 清除本機 token 回 false，交由呼叫端導回登入。
  Future<bool> refreshTokens() =>
      _refreshing ??= _doRefresh().whenComplete(() => _refreshing = null);

  Future<bool> _doRefresh() async {
    final pair = await _tokenStorage.read();
    if (pair == null) return false;
    try {
      final response =
          await _client.refresh(RefreshRequest(refreshToken: pair.refreshToken));
      await _tokenStorage.save(AuthTokenPair(
        accessToken: response.accessToken,
        refreshToken: response.refreshToken,
      ));
      return true;
    } catch (_) {
      await _tokenStorage.clear();
      return false;
    }
  }

  /// 登出：撤銷 refresh token（伺服器端，盡力而為）＋清除本機 token。
  Future<void> logout() async {
    // 先註銷裝置再清 token：註銷是帶身分的 RPC，token 清掉後就沒有憑證可用了。
    await onSessionCleared?.call();
    final pair = await _tokenStorage.read();
    if (pair != null) {
      try {
        await _client.logout(LogoutRequest(refreshToken: pair.refreshToken));
      } catch (_) {
        // 撤銷失敗不擋登出：本機 token 仍會清除，服務端到期自失效。
      }
    }
    await _tokenStorage.clear();
  }

  /// 修改密碼（A3 1.5.2）：`must_change_password=true` 時**唯一**可用的 RPC。
  ///
  /// 必須帶身分呼叫，故走已認證 transport（`_authed`），不是這個物件的未認證
  /// `_client` —— 後者是登入前的公開端點用（login/refresh/QR 兌換）。
  ///
  /// 成功後後端已 `token_version + 1`，本機這對 token 立即失效：這裡**主動清除**
  /// 本機憑證（否則使用者停在一個所有請求都會 401 的殼裡，還得自己找到登出）。
  /// 呼叫端負責導回登入頁，讓使用者以新密碼重登。
  Future<void> changePassword({
    required String oldPassword,
    required String newPassword,
  }) async {
    await _authed.changePassword(ChangePasswordRequest(
      oldPassword: oldPassword,
      newPassword: newPassword,
    ));
    // 改完密碼回到登入頁＝這台裝置的 session 結束：裝置歸屬也該一併解除，
    // 否則下一個用這台手機登入的人會收到前一位的推播。
    await onSessionCleared?.call();
    await _tokenStorage.clear();
  }

  /// QR token 兌換（規格 §4.2：公開端點，不需登入）。
  ///
  /// 回傳公司／客戶識別資訊與**可選的店家子帳號清單**（主帳號與業務子帳號由後端排除）。
  /// 登入仍須子帳號帳密 —— 兌換本身不核發任何憑證（規格明訂）。
  Future<QRLoginResponse> exchangeQR(String token) =>
      _client.qRLogin(QRLoginRequest(token: token));

  /// 業務 Google 登入:開啟系統瀏覽器走 PKCE 授權。
  ///
  /// 回傳授權碼與 verifier(待後端端點上線後換票);使用者取消時回傳 null。
  /// [AuthConfig.googleClientId] 未設定時拋 [StateError](頁面應提示尚未設定)。
  Future<GoogleAuthorizationGrant?> loginSalesWithGoogle() async {
    if (AuthConfig.googleClientId.isEmpty) {
      throw StateError('GOOGLE_CLIENT_ID 未設定,無法發起 Google 登入');
    }
    final codeVerifier = generatePkceCodeVerifier();
    final authorizationUrl = Uri.parse(AuthConfig.googleAuthorizationEndpoint).replace(
      queryParameters: {
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

/// 把登入相關錯誤轉為使用者可讀訊息：見 `core/error_info.dart` 的
/// [localizedErrorMessage]（優先後端 ErrorInfo 碼表）與 [authErrorMessage]（連線層）。
///
/// 這兩個函式**不在此檔**：它們是全 App 共用的錯誤投影，放在 auth feature 會讓
/// 改密碼頁、公告頁都得反向依賴 auth feature。

