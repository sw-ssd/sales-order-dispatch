import 'config.dart';
import 'main.dart';

/// dev flavor 入口:`flutter run --flavor dev -t lib/main_dev.dart`。
void main() {
  bootstrap(const AppConfig(env: AppEnv.dev, apiBaseUrl: 'http://localhost:3080'));
}
