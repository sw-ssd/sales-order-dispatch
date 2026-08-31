import 'config.dart';
import 'main.dart';

/// prod flavor 入口:`flutter run --flavor prod -t lib/main_prod.dart`。
void main() {
  bootstrap(const AppConfig(env: AppEnv.prod, apiBaseUrl: 'https://api.example.com'));
}
