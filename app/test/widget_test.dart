import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/app.dart';
import 'package:sales_order_app/config.dart';

void main() {
  testWidgets('App 啟動後顯示身分選擇頁(/login)', (tester) async {
    await tester.pumpWidget(
      const SalesOrderApp(
        config: AppConfig(env: AppEnv.dev, apiBaseUrl: 'http://localhost:3080'),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('我是店家'), findsOneWidget);
    expect(find.text('我是業務(Google 登入)'), findsOneWidget);
  });
}
