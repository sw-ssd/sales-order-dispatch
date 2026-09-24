import 'package:flutter/cupertino.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:sales_order_app/app.dart';
import 'package:sales_order_app/config.dart';

/// App 業務畫面實機驗證（05/06/07）：
/// 身分選擇 → 店家登入 → 訂單/退貨/通知/我的 四分頁逐頁驗證 → 登出回登入頁。
/// 每分頁截圖（Android 跑出 Material、iOS 跑出 Cupertino,同一份頁面碼）。
/// 先決：後端 dev server 於 localhost:3080、種子帳號 APPCUST01/DevApp123456。
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  Future<void> shot(String name) async {
    try {
      await binding.takeScreenshot(name);
    } catch (e) {
      debugPrint('截圖 $name 失敗（不擋流程）: $e');
    }
  }

  testWidgets('登入→四分頁→登出', (tester) async {
    await tester.pumpWidget(SalesOrderApp(
      config: const AppConfig(
        env: AppEnv.dev,
        apiBaseUrl: 'http://localhost:3080/api/v1',
      ),
    ));
    await tester.pumpAndSettle(const Duration(seconds: 5));

    // 身分選擇頁 → 店家登入。
    expect(find.text('我是店家'), findsOneWidget);
    await shot('00_identity_select');
    await tester.tap(find.text('我是店家'));
    await tester.pumpAndSettle(const Duration(seconds: 5));

    final fields = find.byType(TextFormField);
    expect(fields, findsNWidgets(2));
    await tester.enterText(fields.at(0), 'APPCUST01');
    await tester.enterText(fields.at(1), 'DevApp123456');
    await shot('01_login_form');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle(const Duration(seconds: 10));

    // 底欄四分頁（Material 與 Cupertino 皆以文字標籤呈現）。
    for (final label in ['訂單', '退貨', '通知', '我的']) {
      expect(find.text(label), findsWidgets, reason: '底欄缺少 $label');
    }

    // 平台套件選型（objective: android=material_ui / ios=cupertino_ui）：
    // 同一份頁面碼,依平台渲染對應外殼。
    if (defaultTargetPlatform == TargetPlatform.iOS) {
      expect(find.byType(CupertinoTabScaffold), findsOneWidget,
          reason: 'iOS 應使用 CupertinoTabScaffold');
      expect(find.byType(CupertinoNavigationBar), findsWidgets,
          reason: 'iOS 頁面應使用 CupertinoNavigationBar');
      expect(find.byType(BottomNavigationBar), findsNothing);
    } else {
      expect(find.byType(BottomNavigationBar), findsOneWidget,
          reason: 'Android 應使用 BottomNavigationBar');
      expect(find.byType(AppBar), findsWidgets,
          reason: 'Android 頁面應使用 AppBar');
      expect(find.byType(CupertinoTabScaffold), findsNothing);
    }

    // 訂單分頁：種子單號 W0000xx 應可見（客戶自查範圍）。
    expect(find.textContaining('W0000'), findsWidgets,
        reason: '訂單清單未顯示自己的訂單');
    await shot('02_orders');

    // 退貨分頁：篩選 chips ＋ 清單或空態。
    await tester.tap(find.text('退貨').last);
    await tester.pumpAndSettle(const Duration(seconds: 10));
    expect(find.text('待審核'), findsOneWidget, reason: '退貨篩選缺失');
    expect(
      find.textContaining('退貨 #').evaluate().isNotEmpty ||
          find.text('沒有退貨申請').evaluate().isNotEmpty,
      isTrue,
      reason: '退貨清單既無資料亦無空態',
    );
    await shot('03_returns');

    // 通知分頁：篩選 ＋ 清單或空態；未讀篩選可切。
    await tester.tap(find.text('通知').last);
    await tester.pumpAndSettle(const Duration(seconds: 10));
    expect(find.text('未讀'), findsOneWidget, reason: '通知篩選缺失');
    expect(
      find.text('沒有通知').evaluate().isNotEmpty ||
          find.textContaining('·').evaluate().isNotEmpty,
      isTrue,
      reason: '通知清單載入失敗',
    );
    expect(find.text('載入失敗'), findsNothing);
    await shot('04_notifications');

    // 我的 → 登出 → 回登入選擇。
    await tester.tap(find.text('我的').last);
    await tester.pumpAndSettle(const Duration(seconds: 5));
    expect(find.text('登出'), findsOneWidget);
    await shot('05_profile');
    await tester.tap(find.text('登出'));
    await tester.pumpAndSettle(const Duration(seconds: 5));
    // 確認對話框 → 確定。
    final confirm = find.text('登出').last;
    await tester.tap(confirm);
    await tester.pumpAndSettle(const Duration(seconds: 10));
    expect(find.text('我是店家'), findsOneWidget,
        reason: '登出後應回到身分選擇頁');
    await shot('06_logged_out');
  });
}
