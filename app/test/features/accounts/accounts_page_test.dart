import 'dart:async';

import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';
import 'package:sales_order_app/core/api.dart';
import 'package:sales_order_app/features/accounts/accounts_page.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/customers/v1/customer.connect.spec.dart'
    as specs;
import 'package:sales_order_app/gen/customers/v1/customer.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';

/// 測試用 in-memory token 儲存(不動 Keychain method channel)。
class _InMemoryTokenStorage implements TokenStorage {
  AuthTokenPair? stored;
  @override
  Future<void> save(AuthTokenPair tokens) async => stored = tokens;
  @override
  Future<AuthTokenPair?> read() async => stored;
  @override
  Future<void> clear() async => stored = null;
}

/// 三類帳號各一，用來驗證 UI 是否正確區分「可管理」與「不可管理」。
final _accounts = [
  CustomerAccount(id: '1', accountName: '永和豆漿', isPrimary: true, status: 'active'),
  CustomerAccount(id: '2', accountName: '主廚', manageable: true, status: 'active'),
  CustomerAccount(
    id: '3',
    accountName: '永和豆漿(業務)',
    systemGenerated: true,
    status: 'active',
  ),
];

/// 以 fake transport 驅動帳號管理頁。
///
/// 走 `Api(transport: ...)` 注入而非 mock 整個 Api：每個 RPC 都經過真實 connect client 與
/// proto 編解碼，只有網路層被替換 —— 欄位名打錯這類錯誤仍會被抓到。
Future<void> _pumpAccountsPage(
  WidgetTester tester, {
  required FutureOr<ListCustomerAccountsResponse> Function(
          ListCustomerAccountsRequest, FakeHandlerContext)
      onList,
  FutureOr<DeactivateCustomerAccountResponse> Function(
          DeactivateCustomerAccountRequest, FakeHandlerContext)?
      onDeactivate,
}) async {
  final transport = FakeTransportBuilder()
      .unary<ListCustomerAccountsRequest, ListCustomerAccountsResponse>(
          specs.CustomerAccountService.listCustomerAccounts, onList)
      .unary<DeactivateCustomerAccountRequest, DeactivateCustomerAccountResponse>(
        specs.CustomerAccountService.deactivateCustomerAccount,
        onDeactivate ?? (req, _) => DeactivateCustomerAccountResponse(),
      )
      .build();
  final auth = AuthRepository(
    client: AuthServiceClient(transport),
    // 測試的 fake transport 同時扮演未認證與已認證兩條(ChangePassword 走後者)。
    authed: AuthServiceClient(transport),
    tokenStorage: _InMemoryTokenStorage(),
  );
  final api = Api(baseUrl: 'http://test', auth: auth, transport: transport);

  await tester.pumpWidget(CacheProvider(
    cache: QueryCache(),
    child: MaterialApp(
      home: AccountsPage(api: api, auth: auth, onLoggedOut: () {}),
    ),
  ));
  await tester.pumpAndSettle();
}

/// 收掉 fquery 在「觀察者歸零」時排的 GC 計時器。
///
/// 那是套件行為（query 無人觀察即排 `cacheDuration` 後移除），widget 樹拆掉才會排；
/// 測試若不讓它到期，`flutter_test` 會直接判 "A Timer is still pending"。先拆樹再快轉
/// 超過 `cacheDuration` 讓它自然到期（不是停掉計時器，故不會掩蓋其他計時器洩漏）。
Future<void> _drainQueryGcTimers(WidgetTester tester) async {
  await tester.pumpWidget(const SizedBox());
  await tester.pump(const Duration(minutes: 6));
}

void main() {
  testWidgets('列出三類帳號，僅自建子帳號可管理', (tester) async {
    await _pumpAccountsPage(
      tester,
      onList: (_, _) => ListCustomerAccountsResponse(accounts: _accounts),
    );

    expect(find.text('永和豆漿'), findsOneWidget);
    expect(find.text('主廚'), findsOneWidget);
    expect(find.text('永和豆漿(業務)'), findsOneWidget);

    // 主帳號要說明自己是主帳號（不可自停，否則店家會被鎖在門外）。
    expect(find.textContaining('主帳號'), findsOneWidget);
    // 業務子帳號必須說清楚是「系統預設（業務使用）」—— 使用者沒建過又不能停用的帳號，
    // 不解釋就只像系統壞了。
    expect(find.textContaining('系統預設（業務使用）'), findsOneWidget);
    expect(find.textContaining('啟用中'), findsNWidgets(3));

    // 三列之中只有可管理的那一列有動作入口。
    expect(find.byType(PopupMenuButton<String>), findsOneWidget);

    await _drainQueryGcTimers(tester);
  });

  testWidgets('不可管理的帳號沒有動作入口', (tester) async {
    await _pumpAccountsPage(
      tester,
      onList: (_, _) => ListCustomerAccountsResponse(accounts: [
        CustomerAccount(id: '9', accountName: '主廚', status: 'active'),
      ]),
    );
    expect(find.text('主廚'), findsOneWidget);
    expect(find.byType(PopupMenuButton<String>), findsNothing);

    await _drainQueryGcTimers(tester);
  });

  testWidgets('停用會先確認，並帶上該列帳號的 id', (tester) async {
    DeactivateCustomerAccountRequest? captured;
    await _pumpAccountsPage(
      tester,
      onList: (_, _) => ListCustomerAccountsResponse(accounts: _accounts),
      onDeactivate: (req, _) {
        captured = req;
        return DeactivateCustomerAccountResponse();
      },
    );

    await tester.tap(find.byType(PopupMenuButton<String>));
    await tester.pumpAndSettle();
    await tester.tap(find.text('停用帳號'));
    await tester.pumpAndSettle();
    // 停用不可逆，必須先確認；且**確認前不得**已送出請求。
    expect(captured, isNull);
    await tester.tap(find.text('停用'));
    await tester.pumpAndSettle();

    expect(captured?.accountId, '2', reason: '只可停用自建子帳號（id=2）');

    await _drainQueryGcTimers(tester);
  });

  testWidgets('主帳號的唯一落點是這頁，必須提供登出', (tester) async {
    await _pumpAccountsPage(
      tester,
      onList: (_, _) => ListCustomerAccountsResponse(accounts: _accounts),
    );
    expect(find.byTooltip('登出'), findsOneWidget);

    await _drainQueryGcTimers(tester);
  });

  testWidgets('非主帳號（permission_denied）顯示可讀訊息而非原始錯誤', (tester) async {
    await _pumpAccountsPage(
      tester,
      onList: (_, _) =>
          throw connect.ConnectException(connect.Code.permissionDenied, 'denied'),
    );

    expect(find.textContaining('僅店家主帳號可管理'), findsOneWidget);
    expect(find.textContaining('ConnectException'), findsNothing);

    await _drainQueryGcTimers(tester);
  });
}
