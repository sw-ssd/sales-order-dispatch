# App (Flutter) Development Guidelines

> Flutter 客戶端(iOS/Android);一律以 `fvm` 執行。文件、註解、commit message 一律繁體中文。

## 1. 結構與慣例

- `lib/main_dev.dart` / `lib/main_prod.dart`:雙 flavor 入口;組態值集中 `lib/config.dart`,以 `--dart-define` 注入(**禁止**散寫於頁面,上線前只改 config;參照 `features/auth/auth_config.dart` 模式)。
- `lib/features/<name>/`:feature 目錄,內含 `pages/` 與該 feature 的 repository/transport/storage;跨 feature 共用才上移到 `lib/` 根層。
- `lib/router/app_router.dart`:auto_route 路由;改動路由後必跑 `task gen`(`build_runner build --delete-conflicting-outputs`)。
- `lib/gen/`:proto 生成碼(buf 產出,`task frontend:gen` 或 backend `task proto:gen`);**禁止手改**,需求變更一律改 proto。
- 後端通訊:connectrpc Dart 用戶端(`auth_transport.dart` 為 transport 集中點);token 存取走 `token_storage.dart`,不散落各處。

## 2. 平台與 flavor

- iOS scheme/組態由 `ios/setup_flavors.rb` 一次性注入(僅供重建參考,勿重跑);Flutter xcconfig 以 `#include?` 連結 Pods 設定。
- Android build variant 對齊 dev/prod flavor(`android/app/build.gradle.kts`)。
- OAuth callback scheme(`salesorder://`)異動需同步 AndroidManifest / Info.plist / `AuthConfig.callbackUrl`。

## 3. 開發流程

- `task app:run:dev` / `app:run:prod`:flavor 執行。
- `task app:analyze`:靜態分析(0 issues 才提交;deprecation 警告即修)。
- `task app:test`:全部測試;新增可觀測行為才加測試,不寫 plumbing 測試。
- 不引入新相依套件,除非現有套件明確無法滿足(pubspec.yaml 異動需在 commit message 說明理由)。
