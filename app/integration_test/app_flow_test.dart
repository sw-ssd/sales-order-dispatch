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
/// 先決：後端 dev server 於 localhost:3080、已 seed 的店家帳號
/// （`主廚`＝子帳號、`永和豆漿`＝主帳號，密碼 DevApp123456；見 docs/README 或
/// backend 的 dev seed 說明）。主帳號在後端只有帳號管理可用，故登入後落在不同頁面。
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  /// 對話框內的單行輸入：`promptAdaptive` 依平台給 CupertinoTextField 或 TextField，
  /// 故兩者都要找（寫死其中一種會在另一平台找不到元素）。
  Finder dialogInput() => defaultTargetPlatform == TargetPlatform.iOS
      ? find.byType(CupertinoTextField)
      : find.byType(TextField);

  /// 「新增子帳號」入口：平台各自的慣用寫法（iOS 文字鈕、Android 帶 tooltip 的圖示鈕）。
  Finder addAction() => defaultTargetPlatform == TargetPlatform.iOS
      ? find.text('新增')
      : find.byTooltip('新增子帳號');

  /// 帳號管理頁的登出鈕（同上：iOS 文字、Android 圖示）。
  Finder logoutAction() => defaultTargetPlatform == TargetPlatform.iOS
      ? find.text('登出')
      : find.byTooltip('登出');

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
    await tester.enterText(fields.at(0), '主廚');
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

  // D22／規格 4.2：主帳號登入後**只有**帳號管理可用（業務能力由後端 OpenFGA 排除），
  // 因此 App 必須把它分流到帳號管理頁，而不是主殼 —— 進主殼只會看到一連串 403。
  testWidgets('主帳號登入→帳號管理頁（不進主殼）', (tester) async {
    await tester.pumpWidget(SalesOrderApp(
      config: const AppConfig(
        env: AppEnv.dev,
        apiBaseUrl: 'http://localhost:3080/api/v1',
      ),
    ));
    await tester.pumpAndSettle(const Duration(seconds: 5));

    await tester.tap(find.text('我是店家'));
    await tester.pumpAndSettle(const Duration(seconds: 5));

    final fields = find.byType(TextFormField);
    await tester.enterText(fields.at(0), '永和豆漿');
    await tester.enterText(fields.at(1), 'DevApp123456');
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle(const Duration(seconds: 15));

    // 落在帳號管理，而不是主殼的四分頁。
    expect(find.text('帳號管理'), findsWidgets, reason: '主帳號應導向帳號管理頁');
    expect(find.text('訂單'), findsNothing, reason: '主帳號不該進入業務主殼');

    // 三類帳號都列出，且只有自建子帳號可管理。
    expect(find.text('主廚'), findsWidgets);
    expect(find.textContaining('系統預設（業務使用）'), findsWidgets,
        reason: '業務子帳號必須標示為系統預設（業務使用）');
    await shot('07_primary_accounts');

    // 新增子帳號 → 顯示一次性臨時密碼。
    // 帳號名稱每次執行都不同：account_name 在客戶內唯一，重跑時同名會得到
    // conflict（而不是臨時密碼），讓測試看起來像壞了。
    final newAccountName = '子帳號${DateTime.now().millisecondsSinceEpoch % 100000}';
    await tester.tap(addAction());
    await tester.pumpAndSettle(const Duration(seconds: 3));
    await tester.enterText(dialogInput().last, newAccountName);
    await tester.tap(find.text('建立'));
    await tester.pumpAndSettle(const Duration(seconds: 10));
    expect(find.textContaining('臨時密碼'), findsOneWidget,
        reason: '新增子帳號應顯示一次性臨時密碼');
    await shot('08_temp_password');
    await tester.tap(find.text('確定'));
    await tester.pumpAndSettle(const Duration(seconds: 5));
    expect(find.text(newAccountName), findsWidgets, reason: '新帳號應出現在清單');

    // 主帳號唯一落點是這頁，必須能登出（否則無路可退）。
    await tester.tap(logoutAction());
    await tester.pumpAndSettle(const Duration(seconds: 3));
    // 確認對話框的按鈕文字兩平台皆為「登出」。
    await tester.tap(find.text('登出').last);
    await tester.pumpAndSettle(const Duration(seconds: 10));
    expect(find.text('我是店家'), findsOneWidget, reason: '登出後應回身分選擇頁');
  });
}
