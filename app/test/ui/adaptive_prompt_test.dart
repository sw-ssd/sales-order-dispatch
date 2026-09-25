import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/ui/adaptive.dart';

/// `promptAdaptive` 的 controller 生命週期回歸測試。
///
/// 實際踩到的 bug（Android 實機）：`showDialog` 的 Future 在 `pop()` **當下**就 resolve，
/// 此時對話框還在播退場動畫並持續重建。先前的實作在外層 `finally` 立刻 dispose controller
/// → 動畫期間擲 `A TextEditingController was used after being disposed`。
///
/// 關鍵在於**確認後要繼續 pump 退場動畫**：只 pump 到 Future resolve 為止重現不了，
/// 必須讓動畫真的跑完，才會在舊寫法下擲出例外。
///
/// 兩平台以 `TargetPlatformVariant` 驅動（官方做法，會正確設定／還原
/// `debugDefaultTargetPlatformOverride`；手動設定要在測試主體內還原，`addTearDown`
/// 來不及 —— 繫結在 tearDown 之前就檢查 debug 變數未被改動）。
Future<BuildContext> _openPromptHost(WidgetTester tester) async {
  late BuildContext ctx;
  await tester.pumpWidget(MaterialApp(
    home: Builder(builder: (context) {
      ctx = context;
      return const Scaffold(body: SizedBox());
    }),
  ));
  return ctx;
}

void main() {
  for (final variant in [
    TargetPlatformVariant.only(TargetPlatform.android),
    TargetPlatformVariant.only(TargetPlatform.iOS),
  ]) {
    final label =
        variant.values.first == TargetPlatform.android ? 'Material' : 'Cupertino';

    testWidgets('$label：確認後回傳輸入值，退場動畫期間不擲 dispose 例外', (tester) async {
      final ctx = await _openPromptHost(tester);
      String? value;
      var done = false;
      final future = promptAdaptive(
        ctx,
        title: '新增子帳號',
        confirmLabel: '建立',
        hint: '帳號名稱',
        requireText: true,
      ).then((v) {
        value = v;
        done = true;
        return v;
      });
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(EditableText), '二廚');
      await tester.pump();
      await tester.tap(find.text('建立'));
      await tester.pumpAndSettle(); // 退場動畫：舊寫法在此擲出 dispose 例外
      await future;

      expect(done, isTrue);
      expect(value, '二廚');
      expect(find.text('新增子帳號'), findsNothing);
      expect(tester.takeException(), isNull);
    }, variant: variant);

    testWidgets('$label：取消回 null（與「確認但留空」區分得開）', (tester) async {
      final ctx = await _openPromptHost(tester);
      String? value;
      var done = false;
      final future = promptAdaptive(ctx, title: '備註').then((v) {
        value = v;
        done = true;
        return v;
      });
      await tester.pumpAndSettle();

      await tester.tap(find.text('取消'));
      await tester.pumpAndSettle();
      await future;

      expect(done, isTrue);
      expect(value, isNull, reason: '取消回 null，才與「確認但輸入空字串」區分得開');
      expect(tester.takeException(), isNull);
    }, variant: variant);

    testWidgets('$label：requireText 且留空時不關閉對話框', (tester) async {
      final ctx = await _openPromptHost(tester);
      var resolved = false;
      final future = promptAdaptive(ctx, title: '新增子帳號', requireText: true)
          .then((_) => resolved = true);
      await tester.pumpAndSettle();

      await tester.tap(find.text('確定'));
      await tester.pumpAndSettle();

      // 仍開著 → 使用者可以繼續輸入（而不是把空字串當答案送出去）。
      expect(find.text('新增子帳號'), findsOneWidget);
      expect(resolved, isFalse, reason: 'requireText 留空時不得關閉對話框');
      expect(tester.takeException(), isNull);

      // 收尾：取消關閉，不留未完成的對話框給其他測試。
      await tester.tap(find.text('取消'));
      await tester.pumpAndSettle();
      await future;
    }, variant: variant);
  }
}
