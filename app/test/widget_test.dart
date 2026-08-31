import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/app.dart';
import 'package:sales_order_app/config.dart';

void main() {
  testWidgets('骨架 App 顯示環境標記', (tester) async {
    await tester.pumpWidget(
      const SalesOrderApp(
        config: AppConfig(env: AppEnv.dev, apiBaseUrl: 'http://localhost:3080'),
      ),
    );
    expect(find.textContaining('env: dev'), findsOneWidget);
  });
}
