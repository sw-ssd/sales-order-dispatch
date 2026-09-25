import 'package:flutter/cupertino.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:sales_order_app/app.dart';
import 'package:sales_order_app/config.dart';

/// A3 1.5.2／規格「臨時密碼首次登入強制修改」實機驗證：
/// 臨時密碼登入 → 落在改密碼頁（不進主殼）→ 前端門檻 → 設定新密碼 → 以新密碼登入可用業務功能。
///
/// **先決**：`主廚` 處於臨時密碼態。與 `app_flow_test.dart` 分開成獨立檔，因為那個檔
/// 的其他案例需要 `主廚` 的密碼是固定的 `DevApp123456`——把該帳號改成臨時密碼態會讓它們
/// 全紅。準備與還原由外部指令負責：
///
/// ```
/// # 準備（把 主廚 設為臨時密碼態，密碼 = TempPass123456）
/// go run ./tmp/mktemp "<admin dsn>" 主廚 TempPass123456
/// # 跑本檔
/// fvm flutter test integration_test/change_password_test.dart -d <device> --flavor dev
/// # 還原（把 主廚 設回固定密碼 DevApp123456、非受限態）
/// go run ./tmp/setpw "<admin dsn>" 主廚 DevApp123456
/// ```
///
/// 不把帳號狀態當測試的一部分，是因為「測完自己改回」需要在 UI 上再走一次改密碼頁
/// （帳號數量與名稱還會累積），失敗時也無法保證還原 —— 交給外部的冪等指令更單純。
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  Future<void> shot(String name) async {
    try {
      await binding.takeScreenshot(name);
    } catch (e) {
      debugPrint('截圖 $name 失敗（不擋流程）: $e');
    }
  }

  Future<void> pumpApp(WidgetTester tester) async {
    await tester.pumpWidget(SalesOrderApp(
      config: const AppConfig(
        env: AppEnv.dev,
        apiBaseUrl: 'http://localhost:3080/api/v1',
      ),
    ));
    await tester.pumpAndSettle(const Duration(seconds: 5));
  }

  testWidgets('臨時密碼首登→強制改密碼→以新密碼登入', (tester) async {
    const tempPw = 'TempPass123456';
    final newPw = 'NewPass${DateTime.now().millisecondsSinceEpoch % 100000}';

    await pumpApp(tester);

    await tester.tap(find.text('我是店家'));
    await tester.pumpAndSettle(const Duration(seconds: 5));

    var fields = find.byType(TextFormField);
    await tester.enterText(fields.at(0), '主廚');
    await tester.enterText(fields.at(1), tempPw);
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle(const Duration(seconds: 15));

    // 落在改密碼頁而不是主殼。
    expect(
      find.text('首次登入請先設定新密碼'),
      findsOneWidget,
      reason: '臨時密碼登入應被導向改密碼頁'
          '（若這裡失敗：先確認 主廚 是否處於臨時密碼態，見檔頭先決）',
    );
    expect(find.text('訂單'), findsNothing, reason: '受限態不得進主殼');
    // 平台套件選型仍然成立（改密碼頁也走自適應外殼）。
    if (defaultTargetPlatform == TargetPlatform.iOS) {
      expect(find.byType(CupertinoNavigationBar), findsWidgets);
    } else {
      expect(find.byType(AppBar), findsWidgets);
    }
    await shot('01_force_change_password');

    // 前端鏡射後端門檻：< 8 字元不送出。
    fields = find.byType(TextFormField);
    expect(fields, findsNWidgets(3), reason: '改密碼頁應有舊/新/確認三欄');
    await tester.enterText(fields.at(0), tempPw);
    await tester.enterText(fields.at(1), 'short');
    await tester.enterText(fields.at(2), 'short');
    await tester.tap(find.text('確認修改'));
    await tester.pumpAndSettle(const Duration(seconds: 3));
    expect(find.text('新密碼至少 8 個字元'), findsOneWidget,
        reason: '前端應鏡射後端 8 字元門檻');

    // 兩次不一致也不送出。
    await tester.enterText(fields.at(1), newPw);
    await tester.enterText(fields.at(2), '${newPw}x');
    await tester.tap(find.text('確認修改'));
    await tester.pumpAndSettle(const Duration(seconds: 3));
    expect(find.text('兩次輸入的新密碼不一致'), findsOneWidget);

    // 舊密碼錯：留在頁面且顯示後端訊息（AUTH-4003「帳號或密碼錯誤」）。
    await tester.enterText(fields.at(0), 'WrongOld123');
    await tester.enterText(fields.at(2), newPw);
    await tester.tap(find.text('確認修改'));
    await tester.pumpAndSettle(const Duration(seconds: 10));
    expect(find.text('首次登入請先設定新密碼'), findsOneWidget,
        reason: '舊密碼錯誤應留在改密碼頁');
    // iOS 的 showFeedback 是 modal alert（Android 才是 SnackBar）：不關掉它，
    // 後續的 tap 會被它吃掉（實測 log:「derived an Offset that would not hit test」），
    // 然後停在改密碼頁讓最後一個斷言失敗 —— 看起來像「改完沒回登入頁」。
    if (find.text('確定').evaluate().isNotEmpty) {
      await tester.tap(find.text('確定'));
      await tester.pumpAndSettle(const Duration(seconds: 3));
    }

    // 正確舊密碼 → 完成修改 → 回登入頁。
    await tester.enterText(fields.at(0), tempPw);
    await tester.tap(find.text('確認修改'));
    await tester.pumpAndSettle(const Duration(seconds: 15));
    // 成功訊息同樣會跳 alert（iOS），先關掉再斷言身分選擇頁。
    if (find.text('確定').evaluate().isNotEmpty) {
      await tester.tap(find.text('確定'));
      await tester.pumpAndSettle(const Duration(seconds: 3));
    }
    expect(find.text('我是店家'), findsOneWidget, reason: '改完密碼應回身分選擇頁');
    await shot('02_after_change');

    // 以新密碼登入 → 不再受限，落在主殼。
    await tester.tap(find.text('我是店家'));
    await tester.pumpAndSettle(const Duration(seconds: 5));
    fields = find.byType(TextFormField);
    await tester.enterText(fields.at(0), '主廚');
    await tester.enterText(fields.at(1), newPw);
    await tester.tap(find.text('登入'));
    await tester.pumpAndSettle(const Duration(seconds: 15));

    expect(find.text('首次登入請先設定新密碼'), findsNothing,
        reason: '改完密碼後不該再被判為受限態');
    for (final label in ['訂單', '退貨', '通知', '我的']) {
      expect(find.text(label), findsWidgets, reason: '新密碼登入後底欄缺少 $label');
    }
    await shot('03_new_password_home');
  });
}
