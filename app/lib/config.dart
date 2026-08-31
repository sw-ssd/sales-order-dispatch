/// App 環境（dev / prod flavor 對應）。
enum AppEnv { dev, prod }

/// App 組態：每個 flavor 由對應 main_*.dart 注入。
class AppConfig {
  const AppConfig({required this.env, required this.apiBaseUrl});

  final AppEnv env;
  final String apiBaseUrl;
}
