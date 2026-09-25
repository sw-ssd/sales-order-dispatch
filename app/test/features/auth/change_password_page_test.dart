import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/error_info.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/pages/change_password_page.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.spec.dart'
    as specs;
import 'package:sales_order_app/gen/salesorder/v1/auth.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/common.pb.dart';

class _InMemoryTokenStorage implements TokenStorage {
  AuthTokenPair? stored = const AuthTokenPair(
    accessToken: 'access-1',
    refreshToken: 'refresh-1',
  );

  @override
  Future<void> save(AuthTokenPair tokens) async => stored = tokens;

  @override
  Future<AuthTokenPair?> read() async => stored;

  @override
  Future<void> clear() async => stored = null;
}

/// 後端 `errcode.Code.Error` 的形狀：connect 錯誤 ＋ 一筆 ErrorInfo detail。
connect.ConnectException _coded(
  String code,
  connect.Code connectCode,
) {
  final info = ErrorInfo(code: code, message: code);
  return connect.ConnectException(
    connectCode,
    code,
    details: [
          connect.ErrorDetail('salesorder.v1.ErrorInfo', info.writeToBuffer()),
        ],
  );
}

Widget _wrap(Widget child) => MaterialApp(home: child);

void main() {
  group('errorCodeOf／localizedErrorMessage（ErrorInfo detail 解碼）', () {
    test('由 detail 取出錯誤碼（connect 碼不足以區分 AUTH-3002/3003/3004）', () {
      final err = _coded('AUTH-3004', connect.Code.failedPrecondition);
      expect(errorCodeOf(err), 'AUTH-3004');
      // 三碼共用同一個 connect 碼：只比對 connect 碼的話三者互換也測不出來。
      expect(errorCodeOf(_coded('AUTH-3002', connect.Code.failedPrecondition)),
          'AUTH-3002');
      expect(errorCodeOf(_coded('AUTH-3003', connect.Code.failedPrecondition)),
          'AUTH-3003');
    });

    test('訊息優先用後端碼表，不以 connect 碼代替', () {
      expect(
        localizedErrorMessage(
            _coded('AUTH-3004', connect.Code.failedPrecondition)),
        '首次登入須先修改密碼',
      );
      expect(
        localizedErrorMessage(
            _coded('AUTH-3002', connect.Code.failedPrecondition)),
        '臨時密碼已過期，請聯繫管理員重置',
      );
    });

    test('碼表的佔位符以 detail 參數渲染（AUTH-3003 的 {until}）', () {
      final info = ErrorInfo(code: 'AUTH-3003', message: 'x')
        ..details['until'] = '14:30';
      final err = connect.ConnectException(
        connect.Code.failedPrecondition,
        'x',
        details: [
          connect.ErrorDetail('salesorder.v1.ErrorInfo', info.writeToBuffer()),
        ],
      );
      expect(localizedErrorMessage(err), '帳號已鎖定，請於 14:30 後再試');
    });

    test('無 ErrorInfo detail 時退回連線層訊息，不炸', () {
      expect(errorCodeOf(connect.ConnectException(connect.Code.unavailable, '')),
          isNull);
      expect(
        localizedErrorMessage(
            connect.ConnectException(connect.Code.unavailable, '')),
        '無法連線至伺服器,請檢查網路後再試',
      );
    });
  });

  group('ChangePasswordPage', () {
    testWidgets('舊密碼錯誤時留在頁面並顯示後端碼表訊息', (tester) async {
      final transport = FakeTransportBuilder()
          .unary<ChangePasswordRequest, ChangePasswordResponse>(
              specs.AuthService.changePassword,
              (request, context) => throw _coded(
                  'AUTH-4003', connect.Code.unauthenticated))
          .build();
      final storage = _InMemoryTokenStorage();
      final repo = AuthRepository(
        client: AuthServiceClient(transport),
        authed: AuthServiceClient(transport),
        tokenStorage: storage,
      );

      await tester.pumpWidget(_wrap(
        ChangePasswordPage(authRepository: repo, mustChange: true),
      ));
      await tester.enterText(find.byType(TextFormField).at(0), 'temp-1234');
      await tester.enterText(find.byType(TextFormField).at(1), 'newpass-1234');
      await tester.enterText(find.byType(TextFormField).at(2), 'newpass-1234');
      await tester.tap(find.text('確認修改'));
      await tester.pumpAndSettle();

      // 失敗時不得清除本機 token（否則使用者被登出，還得重打一次舊密碼）。
      expect(storage.stored, isNotNull,
          reason: '改密碼失敗不該清掉 token');
    });

    testWidgets('成功後清除本機 token —— 後端 tv+1 已使這對憑證失效', (tester) async {
      final transport = FakeTransportBuilder()
          .unary<ChangePasswordRequest, ChangePasswordResponse>(
              specs.AuthService.changePassword,
              (request, context) => ChangePasswordResponse())
          .build();
      final storage = _InMemoryTokenStorage();
      final repo = AuthRepository(
        client: AuthServiceClient(transport),
        authed: AuthServiceClient(transport),
        tokenStorage: storage,
      );
      expect(storage.stored, isNotNull);

      await tester.pumpWidget(_wrap(
        ChangePasswordPage(authRepository: repo, mustChange: true),
      ));
      await tester.enterText(find.byType(TextFormField).at(0), 'temp-1234');
      await tester.enterText(find.byType(TextFormField).at(1), 'newpass-1234');
      await tester.enterText(find.byType(TextFormField).at(2), 'newpass-1234');
      await tester.tap(find.text('確認修改'));
      await tester.pumpAndSettle();

      expect(storage.stored, isNull,
          reason: '改完密碼後本機 token 已失效，必須清掉才不會停在 401 的殼裡');
    });

    testWidgets('前端鏡射後端門檻：新密碼 < 8 字元不送出', (tester) async {
      var calls = 0;
      final transport = FakeTransportBuilder()
          .unary<ChangePasswordRequest, ChangePasswordResponse>(
              specs.AuthService.changePassword, (request, context) {
        calls++;
        return ChangePasswordResponse();
      }).build();
      final repo = AuthRepository(
        client: AuthServiceClient(transport),
        authed: AuthServiceClient(transport),
        tokenStorage: _InMemoryTokenStorage(),
      );

      await tester.pumpWidget(_wrap(
        ChangePasswordPage(authRepository: repo, mustChange: true),
      ));
      await tester.enterText(find.byType(TextFormField).at(0), 'temp-1234');
      await tester.enterText(find.byType(TextFormField).at(1), 'short');
      await tester.enterText(find.byType(TextFormField).at(2), 'short');
      await tester.tap(find.text('確認修改'));
      await tester.pumpAndSettle();

      expect(find.text('新密碼至少 8 個字元'), findsOneWidget);
      expect(calls, 0, reason: '前端驗證不過就不該白送一次往返');
    });

    testWidgets('兩次新密碼不一致時不送出', (tester) async {
      var calls = 0;
      final transport = FakeTransportBuilder()
          .unary<ChangePasswordRequest, ChangePasswordResponse>(
              specs.AuthService.changePassword, (request, context) {
        calls++;
        return ChangePasswordResponse();
      }).build();
      final repo = AuthRepository(
        client: AuthServiceClient(transport),
        authed: AuthServiceClient(transport),
        tokenStorage: _InMemoryTokenStorage(),
      );

      await tester.pumpWidget(_wrap(
        ChangePasswordPage(authRepository: repo, mustChange: true),
      ));
      await tester.enterText(find.byType(TextFormField).at(0), 'temp-1234');
      await tester.enterText(find.byType(TextFormField).at(1), 'newpass-1234');
      await tester.enterText(find.byType(TextFormField).at(2), 'different-1234');
      await tester.tap(find.text('確認修改'));
      await tester.pumpAndSettle();

      expect(find.text('兩次輸入的新密碼不一致'), findsOneWidget);
      expect(calls, 0);
    });

    testWidgets('受限態顯示「臨時密碼」標籤；主動更改顯示「目前密碼」', (tester) async {
      final transport = FakeTransportBuilder().build();
      final repo = AuthRepository(
        client: AuthServiceClient(transport),
        authed: AuthServiceClient(transport),
        tokenStorage: _InMemoryTokenStorage(),
      );

      await tester.pumpWidget(_wrap(
        ChangePasswordPage(authRepository: repo, mustChange: true),
      ));
      expect(find.text('臨時密碼'), findsOneWidget);
      expect(find.text('首次登入請先設定新密碼'), findsOneWidget);

      await tester.pumpWidget(_wrap(
        ChangePasswordPage(authRepository: repo, mustChange: false),
      ));
      await tester.pumpAndSettle();
      expect(find.text('目前密碼'), findsOneWidget);
      expect(find.text('請輸入目前的密碼與新密碼'), findsOneWidget);
    });
  });

  group('session 生命週期掛鉤（推播註冊/註銷）', _pushHookTests);

  group('loginShop 回傳受限態旗標', () {
    test('must_change_password=true 時回報必須改密碼', () async {
      final transport = FakeTransportBuilder()
          .unary<LoginRequest, LoginResponse>(specs.AuthService.login,
              (request, context) => LoginResponse(
                    accessToken: 'a',
                    refreshToken: 'r',
                    mustChangePassword: true,
                  ))
          .build();
      final repo = AuthRepository(
        client: AuthServiceClient(transport),
        authed: AuthServiceClient(transport),
        tokenStorage: _InMemoryTokenStorage(),
      );

      final result =
          await repo.loginShop(customerCode: 'C1', password: 'p');
      expect(result.mustChangePassword, isTrue,
          reason: 'true 必須被傳出去，否則首登會被導進主殼並收到一連串 AUTH-3004');
    });
  });
}

/// 推播掛鉤（07 Task 4.3.4）：session 生命週期必須驅動裝置註冊/註銷。
///
/// 這條測的是**接線**而不是 Firebase 本身（那需要原生設定檔）：掛鉤有沒有在對的時機被叫。
/// 掛鉤掛在 AuthRepository 而非各頁呼叫點，正是為了讓「新增登入/登出呼叫端」不會漏掉它
/// —— 所以驗證點放在 repository 的生命週期事件上。
void _pushHookTests() {
  test('登入成功後觸發 session established 掛鉤', () async {
    final transport = FakeTransportBuilder()
        .unary<LoginRequest, LoginResponse>(specs.AuthService.login,
            (req, _) => LoginResponse(accessToken: 'a', refreshToken: 'r'))
        .build();
    final repo = AuthRepository(
      client: AuthServiceClient(transport),
      authed: AuthServiceClient(transport),
      tokenStorage: _InMemoryTokenStorage(),
    );
    var established = 0;
    var cleared = 0;
    repo.onSessionEstablished = () async => established++;
    repo.onSessionCleared = () async => cleared++;

    await repo.loginShop(customerCode: 'C1', password: 'p');
    // 掛鉤是 fire-and-forget（不拖慢登入），故讓 microtask 跑完再斷言。
    await Future<void>.delayed(Duration.zero);

    expect(established, 1, reason: '登入成功應註冊裝置（否則永遠收不到推播）');
    expect(cleared, 0);
  });

  test('登出先註銷裝置再清 token（註銷是帶身分的 RPC）', () async {
    final transport = FakeTransportBuilder()
        .unary<LoginRequest, LoginResponse>(specs.AuthService.login,
            (req, _) => LoginResponse(accessToken: 'a', refreshToken: 'r'))
        .unary<LogoutRequest, LogoutResponse>(
            specs.AuthService.logout, (req, _) => LogoutResponse())
        .build();
    final storage = _InMemoryTokenStorage();
    final repo = AuthRepository(
      client: AuthServiceClient(transport),
      authed: AuthServiceClient(transport),
      tokenStorage: storage,
    );
    var tokenAtUnregister = 'unset';
    repo.onSessionCleared = () async {
      // 註銷當下 token 必須還在，否則 RPC 沒有憑證可用。
      tokenAtUnregister = (await storage.read())?.accessToken ?? 'gone';
    };

    await repo.logout();

    expect(tokenAtUnregister, 'access-1',
        reason: '註銷裝置必須在清 token 之前（否則 RPC 無憑證）');
    expect(storage.stored, isNull);
  });

  test('改密碼成功後也解除裝置歸屬（下一個登入者不該收到前一位的推播）', () async {
    final transport = FakeTransportBuilder()
        .unary<ChangePasswordRequest, ChangePasswordResponse>(
            specs.AuthService.changePassword,
            (req, _) => ChangePasswordResponse())
        .build();
    final storage = _InMemoryTokenStorage();
    final repo = AuthRepository(
      client: AuthServiceClient(transport),
      authed: AuthServiceClient(transport),
      tokenStorage: storage,
    );
    var cleared = 0;
    repo.onSessionCleared = () async => cleared++;

    await repo.changePassword(oldPassword: 'old-1234', newPassword: 'new-1234');

    expect(cleared, 1);
    expect(storage.stored, isNull);
  });

  test('掛鉤擲錯不影響登入（推播是附加能力）', () async {
    final transport = FakeTransportBuilder()
        .unary<LoginRequest, LoginResponse>(specs.AuthService.login,
            (req, _) => LoginResponse(accessToken: 'a', refreshToken: 'r'))
        .build();
    final repo = AuthRepository(
      client: AuthServiceClient(transport),
      authed: AuthServiceClient(transport),
      tokenStorage: _InMemoryTokenStorage(),
    );
    repo.onSessionEstablished = () async => throw StateError('push boom');

    // 不應拋出。
    final result = await repo.loginShop(customerCode: 'C1', password: 'p');
    expect(result.tokens.accessToken, 'a');
  });
}
