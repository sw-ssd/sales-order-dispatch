import 'dart:async';

import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/pages/qr_login_page.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.spec.dart' as specs;
import 'package:sales_order_app/gen/salesorder/v1/auth.pb.dart';

class _InMemoryTokenStorage implements TokenStorage {
  AuthTokenPair? stored;
  @override
  Future<void> save(AuthTokenPair tokens) async => stored = tokens;
  @override
  Future<AuthTokenPair?> read() async => stored;
  @override
  Future<void> clear() async => stored = null;
}

QRLoginResponse _exchange() => QRLoginResponse(
      companyId: '1',
      companyName: 'P2T4 驗收公司',
      customerCode: 'APPCUST01',
      customerName: '永和豆漿',
      accounts: [
        QRLoginResponse_Account(id: '16', accountName: '主廚'),
        QRLoginResponse_Account(id: '17', accountName: '二廚'),
      ],
    );

/// 以 fake transport 驅動 QR 兌換頁。
///
/// 回傳的 `logins` 是**可變清單**（不是單一值）：登入發生在互動之後，
/// 若在 pump 當下就取值只會拿到 null，斷言就永遠測不到東西。
Future<({List<LoginRequest> logins, _InMemoryTokenStorage storage})> _pump(
  WidgetTester tester, {
  required FutureOr<QRLoginResponse> Function(QRLoginRequest, FakeHandlerContext)
      onExchange,
}) async {
  final captured = <LoginRequest>[];
  final storage = _InMemoryTokenStorage();
  final transport = FakeTransportBuilder()
      .unary<QRLoginRequest, QRLoginResponse>(
          specs.AuthService.qRLogin, onExchange)
      .unary<LoginRequest, LoginResponse>(specs.AuthService.login, (req, _) {
    captured.add(req);
    return LoginResponse(
      accessToken: 'a',
      refreshToken: 'r',
      mustChangePassword: false,
    );
  }).build();
  final auth = AuthRepository(
    client: AuthServiceClient(transport),
    tokenStorage: storage,
  );

  await tester.pumpWidget(CacheProvider(
    cache: QueryCache(),
    child: MaterialApp(
      home: QRLoginPage(
        authRepository: auth,
        token: 'tok-1',
        onLoggedIn: (_) {},
      ),
    ),
  ));
  await tester.pumpAndSettle();
  return (logins: captured, storage: storage);
}

/// 收掉 fquery 的 GC 計時器（見 accounts_page_test.dart 同註）。
Future<void> _drain(WidgetTester tester) async {
  await tester.pumpWidget(const SizedBox());
  await tester.pump(const Duration(minutes: 6));
}

void main() {
  testWidgets('兌換後顯示公司/客戶與可選子帳號清單', (tester) async {
    QRLoginRequest? req;
    await _pump(tester, onExchange: (r, _) {
      req = r;
      return _exchange();
    });

    // token 必須原樣送進兌換 RPC（深層連結帶進來的字串）。
    expect(req?.token, 'tok-1');
    expect(find.text('P2T4 驗收公司'), findsOneWidget);
    expect(find.text('永和豆漿'), findsOneWidget);
    expect(find.text('主廚'), findsOneWidget);
    expect(find.text('二廚'), findsOneWidget);
    // 兌換本身不核發憑證：還沒選帳號也沒打密碼前，不該有任何 token 落盤。
    expect(find.text('選擇登入帳號'), findsOneWidget);

    await _drain(tester);
  });

  testWidgets('選帳號 + 密碼後以該帳號名稱登入並存放 token', (tester) async {
    final env = await _pump(tester, onExchange: (_, _) => _exchange());

    await tester.tap(find.text('主廚'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'pw-123');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle();

    // 登入用的是**選中的 account_name**（不是 customer_code）——後端 Login 以
    // users.account_name 查帳號，餵錯欄位會得到「客戶編號或密碼錯誤」。
    expect(env.logins.single.customerCode, '主廚');
    expect(env.logins.single.password, 'pw-123');
    expect(env.storage.stored?.accessToken, 'a');

    await _drain(tester);
  });

  testWidgets('未選帳號不可送出（避免送出必然失敗的請求）', (tester) async {
    final env = await _pump(tester, onExchange: (_, _) => _exchange());

    await tester.enterText(find.byType(TextField), 'pw-123');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle();

    expect(env.logins, isEmpty);

    await _drain(tester);
  });

  testWidgets('token 失效時給可讀訊息與改走帳密登入的出口', (tester) async {
    await _pump(
      tester,
      onExchange: (_, _) => throw connect.ConnectException(
        connect.Code.failedPrecondition,
        'expired',
      ),
    );

    expect(find.textContaining('已失效或已使用過'), findsOneWidget);
    expect(find.text('改用帳密登入'), findsOneWidget);
    // 不得把原始例外字串攤在使用者面前。
    expect(find.textContaining('ConnectException'), findsNothing);

    await _drain(tester);
  });
}
