import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// 登入核發的 token 對(D5:access JWT 1h + refresh 30d 旋轉)。
class AuthTokenPair {
  const AuthTokenPair({required this.accessToken, required this.refreshToken});

  final String accessToken;
  final String refreshToken;
}

/// Token 持久化介面;正式環境用 [SecureTokenStorage],測試用 in-memory 實作。
abstract interface class TokenStorage {
  Future<void> save(AuthTokenPair tokens);
  Future<AuthTokenPair?> read();
  Future<void> clear();
}

/// 以 flutter_secure_storage 存放 token(Keychain / EncryptedSharedPreferences)。
class SecureTokenStorage implements TokenStorage {
  SecureTokenStorage([FlutterSecureStorage? storage])
      : _storage = storage ?? const FlutterSecureStorage();

  static const _accessTokenKey = 'auth.access_token';
  static const _refreshTokenKey = 'auth.refresh_token';

  final FlutterSecureStorage _storage;

  @override
  Future<void> save(AuthTokenPair tokens) async {
    await _storage.write(key: _accessTokenKey, value: tokens.accessToken);
    await _storage.write(key: _refreshTokenKey, value: tokens.refreshToken);
  }

  @override
  Future<AuthTokenPair?> read() async {
    final accessToken = await _storage.read(key: _accessTokenKey);
    final refreshToken = await _storage.read(key: _refreshTokenKey);
    if (accessToken == null || refreshToken == null) return null;
    return AuthTokenPair(accessToken: accessToken, refreshToken: refreshToken);
  }

  @override
  Future<void> clear() async {
    await _storage.delete(key: _accessTokenKey);
    await _storage.delete(key: _refreshTokenKey);
  }
}
