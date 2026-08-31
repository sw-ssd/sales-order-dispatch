/// 業務 Google 登入(OAuth 2.0 + PKCE)組態。
///
/// 後端 OAuth 換票端點尚未定案(依 go8-alignment 後的決策延後),相關值集中於
/// 此,上線前只改這裡;勿散落各頁面。
abstract final class AuthConfig {
  /// Google 授權端點(authorization endpoint)。
  static const googleAuthorizationEndpoint =
      'https://accounts.google.com/o/oauth2/v2/auth';

  /// App 回調 scheme;需與 AndroidManifest / Info.plist 的 callback scheme 一致。
  static const callbackUrlScheme = 'salesorder';

  /// App 回調 URL(授權完成後 Google 導回此 URI)。
  static const callbackUrl = '$callbackUrlScheme://auth';

  /// Google OAuth client id。
  ///
  /// 以 `--dart-define=GOOGLE_CLIENT_ID=...` 注入;未設定時為空字串,
  /// 業務登入按鈕會提示「尚未設定」而非發起無效授權請求。
  static const googleClientId = String.fromEnvironment('GOOGLE_CLIENT_ID');

  /// 向 Google 要求的 scope。
  static const googleScopes = ['openid', 'email', 'profile'];

  /// 後端換票端點(authorization code + code_verifier 換發本系統 token)。
  ///
  /// 後端端點未定,先以常數預留;端點上線後於 AuthRepository 完成串接。
  static const backendGoogleExchangePath = '/api/v1/auth/google';
}
