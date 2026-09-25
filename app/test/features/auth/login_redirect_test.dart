import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/test.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';
import 'package:sales_order_app/core/api.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.spec.dart'
    as specs;
import 'package:sales_order_app/gen/salesorder/v1/auth.pb.dart';
import 'package:sales_order_app/gen/customers/v1/customer.connect.spec.dart'
    as customer_specs;
import 'package:sales_order_app/gen/customers/v1/customer.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/announcement.connect.spec.dart'
    as announcement_specs;
import 'package:sales_order_app/gen/salesorder/v1/announcement.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/notifications.connect.spec.dart'
    as notification_specs;
import 'package:sales_order_app/gen/salesorder/v1/notifications.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/salesorder.connect.spec.dart'
    as order_specs;
import 'package:sales_order_app/gen/salesorder/v1/returns.connect.spec.dart'
    as return_specs;
import 'package:sales_order_app/gen/salesorder/v1/returns.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/salesorder.pb.dart';
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

/// 掛真 router，起點在店家登入頁 `/login/shop`。
///
/// 用真 router 而非單獨 pump [LoginPage]：本測試要驗的正是**登入後被導到哪條路由**
/// （`/change-password` 是否真的註冊、render 得出來），單頁測試測不到路由表。
Future<void> _pumpShopLogin(
  WidgetTester tester, {
  required bool mustChangePassword,
  bool accountsAllowed = false,
}) async {
  final transport = FakeTransportBuilder()
      .unary<LoginRequest, LoginResponse>(specs.AuthService.login,
          (req, _) => LoginResponse(
                accessToken: 'a',
                refreshToken: 'r',
                expiresIn: Int64(3600),
                mustChangePassword: mustChangePassword,
              ))
      .unary<ListCustomerAccountsRequest, ListCustomerAccountsResponse>(
          customer_specs.CustomerAccountService.listCustomerAccounts,
          (req, _) {
        // 主/子分流靠「試呼帳號管理 API」判斷：成功＝主帳號。拒絕才是子帳號
        // （這就是 `accountsAllowed` 在控制的東西）。
        if (!accountsAllowed) {
          throw connect.ConnectException(connect.Code.permissionDenied, 'denied');
        }
        return ListCustomerAccountsResponse();
      })
      // 主殼（子帳號落點）會掛通知角標與訂單清單——不註冊這兩個 handler 的話，
      // fake transport 會以 UnimplementedError 讓測試炸在與本測試無關的地方。
      .unary<UnreadCountRequest, UnreadCountResponse>(
          notification_specs.NotificationService.unreadCount,
          (req, _) => UnreadCountResponse(count: 0))
      .unary<ListOrdersRequest, ListOrdersResponse>(
          order_specs.SalesOrderService.listOrders,
          (req, _) => ListOrdersResponse())
      .unary<ListReturnRequestsRequest, ListReturnRequestsResponse>(
          return_specs.ReturnService.listReturnRequests,
          (req, _) => ListReturnRequestsResponse())
      .unary<ListActiveAnnouncementsRequest, ListActiveAnnouncementsResponse>(
          announcement_specs.AnnouncementService.listActiveAnnouncements,
          (req, _) => ListActiveAnnouncementsResponse())
      .unary<ListNotificationsRequest, ListNotificationsResponse>(
          notification_specs.NotificationService.listNotifications,
          (req, _) => ListNotificationsResponse())
      .build();
  final auth = AuthRepository(
    client: AuthServiceClient(transport),
    authed: AuthServiceClient(transport),
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

  // 從身分選擇頁進店家表單（`/login` 是 initial route）。
  await tester.tap(find.text('我是店家'));
  await tester.pumpAndSettle();
}

Future<void> _drain(WidgetTester tester) async {
  await tester.pumpWidget(const SizedBox());
  await tester.pump(const Duration(minutes: 6));
}

void main() {
  testWidgets('臨時密碼首登 → 落在改密碼頁，且不進主殼', (tester) async {
    await _pumpShopLogin(tester, mustChangePassword: true);

    await tester.enterText(find.byType(TextFormField).at(0), 'C001');
    await tester.enterText(find.byType(TextFormField).at(1), 'temp-1234');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle();

    // 首登受限態：後端只放行 ChangePassword，進主殼只會是一連串 AUTH-3004。
    expect(find.text('首次登入請先設定新密碼'), findsOneWidget);
    expect(find.text('臨時密碼'), findsOneWidget);
    // 主殼的底欄不該出現。
    expect(find.text('訂單'), findsNothing);
    expect(find.text('退貨'), findsNothing);

    await _drain(tester);
  });

  testWidgets('受限態壓過主帳號落點：即使探測成功也不進帳號管理頁', (tester) async {
    // `accountsAllowed: true` 讓分流探測「成功」（＝會被判成主帳號）。若少了受限態那一段，
    // 這個登入會落在 /account；有它就必須落在改密碼頁。這才是可觀察的差異 ——
    // 單純調換兩個 if 的順序測不出來（分流結果與落點都不變）。
    await _pumpShopLogin(
      tester,
      mustChangePassword: true,
      accountsAllowed: true,
    );

    await tester.enterText(find.byType(TextFormField).at(0), 'C001');
    await tester.enterText(find.byType(TextFormField).at(1), 'temp-1234');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle();

    expect(find.text('首次登入請先設定新密碼'), findsOneWidget);
    expect(find.text('新增'), findsNothing, reason: '不得落在帳號管理頁');

    await _drain(tester);
  });

  testWidgets('一般登入（非受限態）仍進主殼，不誤導向改密碼頁', (tester) async {
    await _pumpShopLogin(tester, mustChangePassword: false);

    await tester.enterText(find.byType(TextFormField).at(0), 'C001');
    await tester.enterText(find.byType(TextFormField).at(1), 'pw-12345');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle();

    expect(find.text('首次登入請先設定新密碼'), findsNothing);
    // 子帳號 → 主殼（底欄四項）。
    expect(find.text('訂單'), findsWidgets);

    await _drain(tester);
  });
}
