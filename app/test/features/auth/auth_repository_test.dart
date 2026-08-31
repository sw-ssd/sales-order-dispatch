import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/test.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.spec.dart'
    as specs;
import 'package:sales_order_app/gen/salesorder/v1/auth.pb.dart';

/// 測試用 in-memory token 儲存(不動 Keychain method channel)。
class InMemoryTokenStorage implements TokenStorage {
  AuthTokenPair? stored;

  @override
  Future<void> save(AuthTokenPair tokens) async => stored = tokens;

  @override
  Future<AuthTokenPair?> read() async => stored;

  @override
  Future<void> clear() async => stored = null;
}

void main() {
  test('店家登入以正確 payload 呼叫 AuthService.Login 並儲存 token', () async {
    LoginRequest? captured;
    final transport = FakeTransportBuilder()
        .unary<LoginRequest, LoginResponse>(specs.AuthService.login,
            (request, context) {
          captured = request;
          return LoginResponse(
            accessToken: 'access-123',
            refreshToken: 'refresh-456',
            expiresIn: Int64(3600),
          );
        })
        .build();
    final storage = InMemoryTokenStorage();
    final repository = AuthRepository(
      client: AuthServiceClient(transport),
      tokenStorage: storage,
    );

    final tokens = await repository.loginShop(
      customerCode: 'C001',
      password: 'secret',
    );

    expect(captured, isNotNull);
    expect(captured!.customerCode, 'C001');
    expect(captured!.password, 'secret');
    expect(tokens.accessToken, 'access-123');
    expect(tokens.refreshToken, 'refresh-456');
    expect(storage.stored?.accessToken, 'access-123');
    expect(storage.stored?.refreshToken, 'refresh-456');
  });

  test('AuthService.Login 失敗(unimplemented)不儲存 token 且錯誤向上傳遞',
      () async {
    final transport = FakeTransportBuilder()
        .unary<LoginRequest, LoginResponse>(
            specs.AuthService.login,
            (request, context) =>
                throw connect.ConnectException(connect.Code.unimplemented, ''))
        .build();
    final storage = InMemoryTokenStorage();
    final repository = AuthRepository(
      client: AuthServiceClient(transport),
      tokenStorage: storage,
    );

    await expectLater(
      repository.loginShop(customerCode: 'C001', password: 'secret'),
      throwsA(isA<connect.ConnectException>()),
    );
    expect(storage.stored, isNull);
  });

  test('錯誤訊息對應:unimplemented 與連線失敗有友善提示', () {
    expect(
      authErrorMessage(connect.ConnectException(connect.Code.unimplemented, '')),
      '伺服器尚未支援登入功能,請稍後再試',
    );
    expect(
      authErrorMessage(connect.ConnectException(connect.Code.unavailable, '')),
      '無法連線至伺服器,請檢查網路後再試',
    );
    expect(
      authErrorMessage(
          connect.ConnectException(connect.Code.unauthenticated, '')),
      '客戶編號或密碼錯誤,請重新輸入',
    );
  });

  test('PKCE:S256 challenge 為 verifier 的 base64url SHA-256(無 padding)', () {
    // RFC 7636 Appendix B 官方範例。
    const verifier = 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk';
    expect(
      pkceCodeChallengeS256(verifier),
      'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM',
    );
    final generated = generatePkceCodeVerifier();
    expect(generated.length, 64);
    const allowedChars =
        'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
    expect(generated.split('').every(allowedChars.contains), isTrue);
  });
}
