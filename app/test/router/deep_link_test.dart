import 'package:auto_route/auto_route.dart';
import 'package:connectrpc/test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/api.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';
import 'package:sales_order_app/router/app_router.dart';

class _InMemoryTokenStorage implements TokenStorage {
  AuthTokenPair? stored;
  @override
  Future<void> save(AuthTokenPair tokens) async => stored = tokens;
  @override
  Future<AuthTokenPair?> read() async => stored;
  @override
  Future<void> clear() async => stored = null;
}

/// 建立與生產同構的路由（同樣走 `AppRouter.config()` 的 parser）。
AppRouter _router() {
  final transport = FakeTransportBuilder().build();
  final auth = AuthRepository(
    client: AuthServiceClient(transport),
    tokenStorage: _InMemoryTokenStorage(),
  );
  return AppRouter(
    authRepository: auth,
    api: Api(baseUrl: 'http://test', auth: auth, transport: transport),
  );
}

/// 把一個完整 URL 餵進真實的 route information parser，回傳命中的路由鏈。
///
/// 刻意**不**直接呼叫 matcher 傳 path：真正會出錯的地方就是「Universal Link 帶了
/// host，parser 是否仍能命中」——繞過 parser 的測試測不到那件事。
Future<List<RouteMatch>> _parse(String url) async {
  final config = _router().config();
  final state = await config.routeInformationParser!
      .parseRouteInformation(RouteInformation(uri: Uri.parse(url)));
  return state.segments;
}

void main() {
  // auto_route 的 route info provider 需要 WidgetsBinding（它註冊平台 channel）。
  TestWidgetsFlutterBinding.ensureInitialized();

  // 規格 §4.2：QR 與帳號管理連結皆為 Universal Link / App Link，
  // 形式為 https://<domain>/<path>。以下用完整 URL（含 host）驗證解析。
  test('Universal Link 的 QR 連結命中 QR 路由並取出 token', () async {
    final segments = await _parse(
      'https://app.salesorder.example.com/customer_account_qrcode/abc123.def-456',
    );
    final last = segments.last;
    expect(last.name, 'QRLoginRoute');
    // token 是簽章字串（含 base64url 的 . 與 -），必須原樣取出、不得被截斷或解碼。
    expect(last.params.getString('token'), 'abc123.def-456');
  });

  test('Universal Link 的帳號管理連結命中帳號管理登入路由', () async {
    final segments = await _parse(
      'https://app.salesorder.example.com/customer_account_manage',
    );
    expect(segments.last.name, 'AccountManageLoginRoute');
  });

  test('帶 query 與尾斜線的連結仍命中（連結可能被聊天軟體加工）', () async {
    final segments = await _parse(
      'https://app.salesorder.example.com/customer_account_qrcode/tok%3D1/?utm_source=line',
    );
    expect(segments.last.name, 'QRLoginRoute');
    // 百分比編碼必須還原成原始 token，否則後端驗簽會失敗。
    expect(segments.last.params.getString('token'), 'tok=1');
  });

  test('帳號管理連結不含任何憑證欄位（規格明訂）', () async {
    // 這條是防走鐘：若日後有人為了「方便」把 token 塞進管理連結，這裡會失敗。
    final segments = await _parse(
      'https://app.salesorder.example.com/customer_account_manage?token=xyz',
    );
    expect(segments.last.name, 'AccountManageLoginRoute');
    expect(segments.last.params.get('token'), isNull);
  });
}
