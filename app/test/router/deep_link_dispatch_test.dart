import 'package:connectrpc/test.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';
import 'package:sales_order_app/core/api.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.spec.dart' as specs;
import 'package:sales_order_app/gen/salesorder/v1/auth.pb.dart';
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

/// 走**平台通道**把深層連結送進 App，而非直接呼叫 parser。
///
/// 這正是 iOS 收到 Universal Link、Android 收到 App Link 後 Flutter 引擎走的路徑：
/// 原生 → `flutter/navigation` 的 `pushRouteInformation` → WidgetsBinding → router delegate
/// → route parser → 頁面。少了這一層，測試就測不到「host 被丟棄」「座標還原」等
/// 真正會出錯的環節。
Future<void> _pushDeepLink(WidgetTester tester, String url) async {
  await tester.binding.defaultBinaryMessenger.handlePlatformMessage(
    'flutter/navigation',
    // JSONMethodCodec（不是 StandardMethodCodec）：`flutter/navigation` 是 JSON 通道，
    // 用錯編碼會得到 FormatException 而不是導航。
    const JSONMethodCodec().encodeMethodCall(
      MethodCall('pushRouteInformation', <String, Object?>{'location': url}),
    ),
    (_) {},
  );
  await tester.pumpAndSettle();
}

/// 建立 App；回傳 captured 供斷言「連結帶的 token 真的送到兌換 RPC」。
Future<List<QRLoginRequest>> _pumpApp(WidgetTester tester) async {
  final captured = <QRLoginRequest>[];
  final transport = FakeTransportBuilder()
      .unary<QRLoginRequest, QRLoginResponse>(specs.AuthService.qRLogin, (req, _) {
    captured.add(req);
    // 回一個合法兌換結果：本測試要驗導航與參數傳遞，不是錯誤態。
    return QRLoginResponse(
      companyId: '1',
      companyName: 'P2T4 驗收公司',
      customerCode: 'FAKE-000001',
      customerName: '永和豆漿',
      accounts: [QRLoginResponse_Account(id: '20', accountName: '主廚')],
    );
  })
      .build();
  final auth = AuthRepository(
    client: AuthServiceClient(transport),
    tokenStorage: _InMemoryTokenStorage(),
  );
  final router = AppRouter(
    authRepository: auth,
    api: Api(baseUrl: 'http://test', auth: auth, transport: transport),
  );
  await tester.pumpWidget(CacheProvider(
    cache: QueryCache(),
    child: MaterialApp.router(routerConfig: router.config()),
  ));
  await tester.pumpAndSettle();
  return captured;
}

void main() {
  testWidgets('Universal Link 的 QR 連結會打開 QR 兌換頁並帶入 token', (tester) async {
    TestWidgetsFlutterBinding.ensureInitialized();
    final captured = await _pumpApp(tester);

    // 起始在身分選擇頁。
    expect(find.text('我是店家'), findsOneWidget);

    await _pushDeepLink(
      tester,
      'https://app.salesorder.example.com/customer_account_qrcode/tok-from-link',
    );

    // 已離開起始頁 → 連結確實在 App 內導航了。
    expect(find.text('我是店家'), findsNothing);
    // 連結裡的路徑參數原樣送進兌換 RPC（host 已被 parser 丟棄、token 未被截斷）。
    expect(captured.single.token, 'tok-from-link');
    // 兌換成功 → 顯示公司／客戶與可選子帳號。
    expect(find.text('P2T4 驗收公司'), findsOneWidget);
    expect(find.text('選擇登入帳號'), findsOneWidget);

    // 收掉 fquery 觀察者歸零後的 GC 計時器（見 accounts_page_test.dart 同註）。
    await tester.pumpWidget(const SizedBox());
    await tester.pump(const Duration(minutes: 6));
  });

  testWidgets('Universal Link 的帳號管理連結會打開帳號管理登入頁', (tester) async {
    TestWidgetsFlutterBinding.ensureInitialized();
    await _pumpApp(tester);

    await _pushDeepLink(
      tester,
      'https://app.salesorder.example.com/customer_account_manage',
    );

    // 連結不含憑證（規格明訂）：進來的是一般登入頁，只是標題不同。
    expect(find.text('帳號管理登入'), findsOneWidget);
    expect(find.text('我是店家'), findsNothing);
    // 未登入前不得直接看到帳號管理內容。
    expect(find.text('新增'), findsNothing);
  });

  testWidgets('未知路徑不會把使用者困在錯誤頁（維持既有畫面）', (tester) async {
    TestWidgetsFlutterBinding.ensureInitialized();
    await _pumpApp(tester);

    await _pushDeepLink(
      tester,
      'https://app.salesorder.example.com/some/other/path',
    );

    // 無對應路由時 auto_route 不匹配任何 segment；行為是留在原頁（不是白畫面）。
    expect(find.text('我是店家'), findsOneWidget);
  });
}
