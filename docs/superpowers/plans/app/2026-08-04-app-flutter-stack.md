# App Flutter 技術棧基礎（D29）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `app/` 建立並驗證 D29 App 技術棧基礎：solidart + disco + auto_route + fquery + Sembast 唯讀快取鏡像 + connectrpc 認證整合，作為 Phase 1.10 / 4.6 / 6.x 等後續 App 任務的地基。

**Architecture:** feature-first 分層（`core/` + `features/*` + `router/` + `gen/`）。server state 全在 fquery（`QueryCache` 根層提供）；solidart 只管 session 與畫面局部狀態；disco `ProviderScope` 三層（根 / auth / 路由級）做 DI；connectrpc `Transport` 注入 AuthInterceptor（Bearer 附加、401→refresh 旋轉重試、403→登出）。Sembast 為唯讀快取鏡像（write-through + 啟動 seed）。

**Tech Stack:** Flutter stable（Dart ≥3.10）、solidart 2.8.6、disco 2.0.0、auto_route 11.1.0（+generator）、fquery 3.1.0、sembast、flutter_secure_storage、flutter_web_auth、connectrpc（connect-dart）、http、path_provider、build_runner。

**來源文件：** `docs/superpowers/specs/2026-08-04-app-flutter-stack-design.md`（D29 設計，本計畫逐一對應其 §5–§11）。本計畫為 D29 追加計畫，Task 10 將內容合併回主計畫 `docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md`（升 v2.10.0），主計畫 Task 0.6（App 骨架）以本計畫 Task 1–9 取代。

## Global Constraints

- **Flutter stable、Dart SDK ≥3.10.0**（disco 2.0.0 / solidart 2.8.6 硬性要求；本機不足先 `flutter upgrade --force` 再驗證）。
- **套件鎖版**：solidart ^2.8.6、flutter_solidart ^2.8.6、disco ^2.0.0、auto_route ^11.1.0、fquery ^3.1.0、sembast ^3.7.0、sembast_io ^3.7.0、flutter_secure_storage ^9.0.0、flutter_web_auth ^0.5.0、connectrpc 最新穩定、http ^1.0.0、path_provider ^2.0.0；dev：auto_route_generator ^11.1.0、build_runner ^2.0.0、flutter_lints（預設）。
- **業務 API 僅 Connect-RPC**（D4）；REST 僅公開端點（檔案 / QR / 版本 / 公開資訊），**不進任何快取**。
- **下單需連網**（規格 §9.6）；離線僅唯讀快取顯示，不實作離線寫入。
- **認證（D5）**：access JWT 1h + refresh 30d 旋轉，存 `flutter_secure_storage`；`permissionDenied`（403，tv 不符）→ 登出 + 清空 QueryCache + 清鏡像。
- **UI 字串一律繁體中文**。
- **覆蓋率 ≥70%**（D21）：每個 Task 附測試，`flutter test` 全綠為提交門檻。
- **proto 型別**由 buf 產生至 `app/lib/gen/`（主計畫 Task 0.7）；本計畫不依賴 generated client（以 `Spec` 手建 + mock 測試），Task 6/9 的介面已為 generated client 預留。
- **commit**：monorepo git 建立後（主計畫 Task 0.1）每個 Task 結束即 commit。

---

### Task 1: App 專案骨架與依賴鎖版

**Files:**
- Create: `app/`（`flutter create` 產出）
- Modify: `app/pubspec.yaml`
- Modify: `app/analysis_options.yaml`（保留 flutter_lints 預設）

**Interfaces:**
- 產出 `app/lib/main.dart`（預設）、`app/test/widget_test.dart`（預設，Task 8 取代）。
- 依賴解析版本寫入 `app/pubspec.lock`，作為全倉鎖版基準。

- [ ] **Step 1: 驗證 Flutter / Dart 版本**

Run: `flutter --version`
Expected: `Dart 3.10.0` 或更高（Flutter stable）。低於此 → 執行 `flutter upgrade --force` 後重跑，直到滿足。

- [ ] **Step 2: 建立 App 骨架**

Run:
```bash
cd /Volumes/UTM2/Developer/sales-order-new
flutter create --platforms=android,ios --empty --org com.salesorder --project-name sales_order_app app
```
Expected: `app/` 建立完成，`app/lib/main.dart` 為最小 App，`flutter test` 可跑（空專案無失敗測試）。

- [ ] **Step 3: 加入執行期依賴並鎖版**

Run:
```bash
cd app
flutter pub add solidart:^2.8.6 flutter_solidart:^2.8.6 disco:^2.0.0 \
  auto_route:^11.1.0 fquery:^3.1.0 sembast:^3.7.0 sembast_io:^3.7.0 \
  flutter_secure_storage:^9.0.0 flutter_web_auth:^0.5.0 connectrpc http:^1.0.0 path_provider:^2.0.0
flutter pub add dev:auto_route_generator:^11.1.0 dev:build_runner:^2.0.0
```
Expected: `pubspec.yaml` 含上述依賴（connectrpc 解析為最新穩定版），`pubspec.lock` 產生。`flutter pub get` 成功（connectrpc 若要求更高 Dart SDK，回到 Step 1 升 Flutter）。

- [ ] **Step 4: 驗證骨架**

Run: `flutter analyze && flutter test`
Expected: 0 分析錯誤；`flutter test` 全綠（預設無測試亦通過）。

- [ ] **Step 5: Commit**

```bash
git add app/ && git commit -m "chore(app): scaffold flutter app with D29 pinned dependencies"
```

---

### Task 2: core/config 與 core/errors

**Files:**
- Create: `app/lib/core/config/app_config.dart`
- Create: `app/lib/core/errors/app_error.dart`
- Test: `app/test/core/errors/app_error_test.dart`

**Interfaces:**
- Consumes: 無（首批 core 檔案）。
- Produces:
  - `class AppConfig { final String apiBaseUrl; final String env; const AppConfig({required this.apiBaseUrl, required this.env}); factory AppConfig.fromEnv() }` — `apiBaseUrl` 來自 `String.fromEnvironment('API_BASE_URL', defaultValue: 'https://api.example.com')`，`env` 來自 `String.fromEnvironment('APP_ENV', defaultValue: 'dev')`。
  - `sealed class AppError` 五子類：`NetworkError`、`AuthError`、`PermissionError`、`BusinessError(String message)`、`UnexpectedError(String message)`。
  - `AppError AppError.from(Object error)`：`ConnectException` 依 `code` 對映（表見 Step 1），其餘例外 → `UnexpectedError`。
  - `String AppError.userMessage(AppError)`：繁中使用者可讀文案。

- [ ] **Step 1: 寫 failing test**

```dart
// test/core/errors/app_error_test.dart
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/errors/app_error.dart';

void main() {
  group('AppError.from', () {
    test('maps ConnectException codes', () {
      expect(AppError.from(ConnectException(Code.unauthenticated)), isA<AuthError>());
      expect(AppError.from(ConnectException(Code.permissionDenied)), isA<PermissionError>());
      expect(AppError.from(ConnectException(Code.unavailable)), isA<NetworkError>());
      expect(AppError.from(ConnectException(Code.deadlineExceeded)), isA<NetworkError>());
      expect(AppError.from(ConnectException(Code.invalidArgument, '數量不可為 0')), isA<BusinessError>());
      expect(AppError.from(StateError('boom')), isA<UnexpectedError>());
    });

    test('userMessage is Traditional Chinese', () {
      final e = AppError.from(ConnectException(Code.invalidArgument, '數量不可為 0'));
      expect(AppError.userMessage(e), contains('數量不可為 0'));
      expect(AppError.userMessage(AppError.from(StateError('x'))), '發生未預期錯誤，請稍後再試');
    });
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd app && flutter test test/core/errors/app_error_test.dart`
Expected: FAIL — `AppError` 未定義。

- [ ] **Step 3: 實作 `app_error.dart`**

```dart
// app/lib/core/errors/app_error.dart
import 'package:connectrpc/connect.dart';

sealed class AppError {
  const AppError(this.message);
  final String message;

  factory AppError.from(Object error) {
    if (error is ConnectException) {
      return switch (error.code) {
        Code.unauthenticated => const AuthError('登入已失效，請重新登入'),
        Code.permissionDenied => const PermissionError('您沒有權限執行此操作'),
        Code.unavailable || Code.deadlineExceeded || Code.unimplemented =>
          const NetworkError('無法連線，請檢查網路'),
        Code.invalidArgument || Code.failedPrecondition || Code.outOfRange ||
        Code.alreadyExists || Code.aborted =>
          BusinessError(error.message.isEmpty ? '請求未通過，請確認資料' : error.message),
        _ => UnexpectedError(error.message.isEmpty ? '發生未預期錯誤，請稍後再試' : error.message),
      };
    }
    return const UnexpectedError('發生未預期錯誤，請稍後再試');
  }

  static String userMessage(AppError e) => e.message;
}

final class NetworkError extends AppError { const NetworkError(super.message); }
final class AuthError extends AppError { const AuthError(super.message); }
final class PermissionError extends AppError { const PermissionError(super.message); }
final class BusinessError extends AppError { const BusinessError(super.message); }
final class UnexpectedError extends AppError { const UnexpectedError(super.message); }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd app && flutter test test/core/errors/app_error_test.dart`
Expected: PASS（2 個測試）。

- [ ] **Step 5: Commit**

```bash
git add app/lib/core app/test/core/errors && git commit -m "feat(app): core config and AppError mapping"
```

---

### Task 3: core/cache — QueryCache factory 與預設策略

**Files:**
- Create: `app/lib/core/cache/query_cache_factory.dart`
- Test: `app/test/core/cache/query_cache_factory_test.dart`

**Interfaces:**
- Consumes: 無。
- Produces:
  - `final DefaultQueryOptions kQueryCacheDefaults` — `staleDuration: 2 分鐘`、`cacheDuration: 20 分鐘`、`refetchOnMount: RefetchOnMount.stale`、`retryCount: 2`、`retryDelay: 1 秒`、`refetchInterval: null`（設計 §8.2）。
  - `class QueryCacheFactory { static QueryCache create() }` — `QueryCache(defaultQueryOptions: kQueryCacheDefaults)`。

- [ ] **Step 1: 寫 failing test**

```dart
// test/core/cache/query_cache_factory_test.dart
import 'package:fquery/fquery.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/cache/query_cache_factory.dart';

void main() {
  test('defaults follow design §8.2', () {
    expect(kQueryCacheDefaults.staleDuration, const Duration(minutes: 2));
    expect(kQueryCacheDefaults.cacheDuration, const Duration(minutes: 20));
    expect(kQueryCacheDefaults.retryCount, 2);
    expect(kQueryCacheDefaults.retryDelay, const Duration(seconds: 1));
    expect(kQueryCacheDefaults.refetchOnMount, RefetchOnMount.stale);
    expect(kQueryCacheDefaults.refetchInterval, isNull);
  });

  test('factory builds a QueryCache', () {
    expect(QueryCacheFactory.create(), isA<QueryCache>());
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd app && flutter test test/core/cache/query_cache_factory_test.dart`
Expected: FAIL — 檔案不存在 / 型別未定義。

- [ ] **Step 3: 實作 `query_cache_factory.dart`**

```dart
// app/lib/core/cache/query_cache_factory.dart
import 'package:fquery/fquery.dart';

/// 全 App 共用查詢預設（設計文件 §8.2）。
final DefaultQueryOptions kQueryCacheDefaults = DefaultQueryOptions(
  staleDuration: const Duration(minutes: 2),
  cacheDuration: const Duration(minutes: 20),
  refetchOnMount: RefetchOnMount.stale,
  retryCount: 2,
  retryDelay: const Duration(seconds: 1),
);

class QueryCacheFactory {
  const QueryCacheFactory._();

  static QueryCache create() => QueryCache(defaultQueryOptions: kQueryCacheDefaults);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd app && flutter test test/core/cache/query_cache_factory_test.dart`
Expected: PASS（2 個測試）。

- [ ] **Step 5: Commit**

```bash
git add app/lib/core/cache app/test/core/cache && git commit -m "feat(app): fquery QueryCache factory with §8.2 defaults"
```

---

### Task 4: core/cache — Sembast 唯讀快取鏡像（CacheRepository）

**Files:**
- Create: `app/lib/core/cache/cache_repository.dart`
- Test: `app/test/core/cache/cache_repository_test.dart`

**Interfaces:**
- Consumes: 無。
- Produces:
  - `abstract class CacheRepository { Future<Map<String, Object?>?> read(String key); Future<void> write(String key, Map<String, Object?> data); Future<void> remove(String key); Future<void> clear(); }`
  - `class SembastCacheRepository implements CacheRepository` — 建構式 `SembastCacheRepository({required DatabaseFactory databaseFactory, required String path})`（sembast store 名 `query_cache`，key 為 store path）。
  - Task 9 依賴此介面做 seed / write-through。

- [ ] **Step 1: 寫 failing test**

```dart
// test/core/cache/cache_repository_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sembast/sembast_memory.dart';
import 'package:sales_order_app/core/cache/cache_repository.dart';

void main() {
  late SembastCacheRepository repo;

  setUp(() async {
    // databaseFactoryMemory 為記憶體實作，測試不需真實檔案系統
    repo = SembastCacheRepository(
      databaseFactory: databaseFactoryMemory,
      path: 'test-cache.db',
    );
    await repo.clear();
  });

  test('write then read returns the same data', () async {
    await repo.write('customers', {'items': [{'code': 'TY000001'}]});
    final data = await repo.read('customers');
    expect(data, isNotNull);
    expect(data!['items'], isA<List<Object?>>());
  });

  test('read missing key returns null', () async {
    expect(await repo.read('nope'), isNull);
  });

  test('remove and clear drop data', () async {
    await repo.write('a', {'x': 1});
    await repo.remove('a');
    expect(await repo.read('a'), isNull);

    await repo.write('b', {'y': 2});
    await repo.clear();
    expect(await repo.read('b'), isNull);
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd app && flutter test test/core/cache/cache_repository_test.dart`
Expected: FAIL — 型別未定義。

- [ ] **Step 3: 實作 `cache_repository.dart`**

```dart
// app/lib/core/cache/cache_repository.dart
import 'package:sembast/sembast.dart';

abstract class CacheRepository {
  Future<Map<String, Object?>?> read(String key);
  Future<void> write(String key, Map<String, Object?> data);
  Future<void> remove(String key);
  Future<void> clear();
}

/// Sembast 鏡像：唯讀快取（設計 §8.4）。僅存 JSON 可序列化資料。
class SembastCacheRepository implements CacheRepository {
  SembastCacheRepository({required this.databaseFactory, required this.path});

  final DatabaseFactory databaseFactory;
  final String path;

  static const _storeName = 'query_cache';
  late final Future<Database> _db = databaseFactory.openDatabase(path);
  StoreRef<String, Map<String, Object?>> get _store => stringMapStoreFactory.store(_storeName);

  @override
  Future<Map<String, Object?>?> read(String key) async =>
      (await _store.record(key).get(await _db));

  @override
  Future<void> write(String key, Map<String, Object?> data) async =>
      _store.record(key).put(await _db, data);

  @override
  Future<void> remove(String key) async => _store.record(key).delete(await _db);

  @override
  Future<void> clear() async => _store.delete(await _db);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd app && flutter test test/core/cache/cache_repository_test.dart`
Expected: PASS（3 個測試）。

- [ ] **Step 5: Commit**

```bash
git add app/lib/core/cache app/test/core/cache && git commit -m "feat(app): sembast mirror CacheRepository (write-through/seed ready)"
```

---

### Task 5: core/auth — TokenStore、AuthSession、單飛 refresh

**Files:**
- Create: `app/lib/core/auth/token_store.dart`
- Create: `app/lib/core/auth/auth_session.dart`
- Create: `app/lib/core/auth/auth_refresh_controller.dart`
- Test: `app/test/core/auth/token_store_test.dart`、`app/test/core/auth/auth_session_test.dart`、`app/test/core/auth/auth_refresh_controller_test.dart`

**Interfaces:**
- Consumes: 無。
- Produces:
  - `abstract class TokenStore { Future<String?> readAccessToken(); Future<String?> readRefreshToken(); Future<void> writeTokens({required String access, required String refresh}); Future<void> clear(); }`
  - `class SecureTokenStore implements TokenStore` — 建構式 `SecureTokenStore({FlutterSecureStorage? storage})`，key：`access_token` / `refresh_token`。
  - `class SessionUser { final String userId; final String roleCode; final bool isPrimary; final String? companyId; final String? departmentId; const SessionUser({...}); Map<String, Object?> toJson(); factory SessionUser.fromJson(Map<String, Object?>); }`
  - `class AuthSession { AuthSession({required this.tokenStore}); final TokenStore tokenStore; final Signal<SessionUser?> user; final Signal<String?> accessToken; final Signal<bool> expired; Computed<bool> get signedIn; Future<void> restore(); Future<void> signIn(SessionUser u, {required String access, required String refresh}); Future<void> signOut(); void markExpired(); }`
  - `class AuthRefreshController { AuthRefreshController({required Future<bool> Function() doRefresh, required VoidCallback onRotated}); Future<bool> refresh(); }` — **單飛鎖**：並發呼叫共享同一 in-flight future；成功時觸發 `onRotated`。

- [ ] **Step 1: 寫 failing test（token_store + session + refresh controller）**

```dart
// test/core/auth/token_store_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/auth/token_store.dart';

class FakeSecureStorage implements SecureStorage {
  final store = <String, String>{};
  @override Future<String?> read({required String key}) async => store[key];
  @override Future<void> write({required String key, required String value}) async =>
      store[key] = value;
  @override Future<void> delete({required String key}) async => store.remove(key);
  @override Future<void> deleteAll() async => store.clear();
}

void main() {
  test('SecureTokenStore round-trips and clears', () async {
    final fake = FakeSecureStorage();
    final store = SecureTokenStore(storage: fake);
    await store.writeTokens(access: 'a1', refresh: 'r1');
    expect(await store.readAccessToken(), 'a1');
    expect(await store.readRefreshToken(), 'r1');
    await store.clear();
    expect(await store.readAccessToken(), isNull);
  });
}
```

```dart
// test/core/auth/auth_session_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/auth/auth_session.dart';
import 'package:sales_order_app/core/auth/token_store.dart';

class InMemoryTokenStore implements TokenStore {
  String? access; String? refresh;
  @override Future<String?> readAccessToken() async => access;
  @override Future<String?> readRefreshToken() async => refresh;
  @override Future<void> writeTokens({required String access, required String refresh}) async { this.access = access; this.refresh = refresh; }
  @override Future<void> clear() async { access = null; refresh = null; }
}

void main() {
  test('signedIn computed reacts to signIn/signOut', () async {
    final session = AuthSession(tokenStore: InMemoryTokenStore());
    expect(session.signedIn.value, isFalse);

    await session.signIn(
      const SessionUser(userId: 'u1', roleCode: 'customer', isPrimary: false),
      access: 'a', refresh: 'r',
    );
    expect(session.signedIn.value, isTrue);
    expect(session.user.value!.isPrimary, isFalse);

    await session.signOut();
    expect(session.signedIn.value, isFalse);
  });
}
```

```dart
// test/core/auth/auth_refresh_controller_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/auth/auth_refresh_controller.dart';

void main() {
  test('concurrent refresh() calls share one doRefresh (single-flight)', () async {
    var calls = 0;
    var rotated = 0;
    final c = AuthRefreshController(
      doRefresh: () async { calls++; await Future<void>.delayed(const Duration(milliseconds: 20)); return true; },
      onRotated: () => rotated++,
    );
    final results = await Future.wait([c.refresh(), c.refresh(), c.refresh()]);
    expect(results, everyElement(isTrue));
    expect(calls, 1);
    expect(rotated, 1);
  });

  test('failure returns false and allows retry later', () async {
    var calls = 0;
    final c = AuthRefreshController(
      doRefresh: () async { calls++; return calls > 1; },
      onRotated: () {},
    );
    expect(await c.refresh(), isFalse);
    expect(await c.refresh(), isTrue);
  });
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd app && flutter test test/core/auth`
Expected: FAIL — 型別未定義。

- [ ] **Step 3: 實作三檔**

```dart
// app/lib/core/auth/token_store.dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// secure storage 抽象：讓測試以記憶體 fake 注入（正式實作包 FlutterSecureStorage）。
abstract class SecureStorage {
  Future<String?> read({required String key});
  Future<void> write({required String key, required String value});
  Future<void> delete({required String key});
  Future<void> deleteAll();
}

class FlutterSecureStorageAdapter implements SecureStorage {
  const FlutterSecureStorageAdapter({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  @override Future<String?> read({required String key}) => _storage.read(key: key);
  @override Future<void> write({required String key, required String value}) =>
      _storage.write(key: key, value: value);
  @override Future<void> delete({required String key}) => _storage.delete(key: key);
  @override Future<void> deleteAll() => _storage.deleteAll();
}

abstract class TokenStore {
  Future<String?> readAccessToken();
  Future<String?> readRefreshToken();
  Future<void> writeTokens({required String access, required String refresh});
  Future<void> clear();
}

class SecureTokenStore implements TokenStore {
  SecureTokenStore({SecureStorage? storage}) : _storage = storage ?? const FlutterSecureStorageAdapter();

  final SecureStorage _storage;
  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';

  @override Future<String?> readAccessToken() => _storage.read(key: _accessKey);
  @override Future<String?> readRefreshToken() => _storage.read(key: _refreshKey);
  @override Future<void> writeTokens({required String access, required String refresh}) async {
    await _storage.write(key: _accessKey, value: access);
    await _storage.write(key: _refreshKey, value: refresh);
  }
  @override Future<void> clear() => _storage.deleteAll();
}
```

```dart
// app/lib/core/auth/auth_session.dart
import 'package:solidart/solidart.dart';

class SessionUser {
  const SessionUser({
    required this.userId,
    required this.roleCode,
    required this.isPrimary,
    this.companyId,
    this.departmentId,
  });

  final String userId;
  final String roleCode;
  final bool isPrimary;
  final String? companyId;
  final String? departmentId;

  Map<String, Object?> toJson() => {
        'userId': userId,
        'roleCode': roleCode,
        'isPrimary': isPrimary,
        'companyId': companyId,
        'departmentId': departmentId,
      };

  factory SessionUser.fromJson(Map<String, Object?> json) => SessionUser(
        userId: json['userId']! as String,
        roleCode: json['roleCode']! as String,
        isPrimary: json['isPrimary']! as bool,
        companyId: json['companyId'] as String?,
        departmentId: json['departmentId'] as String?,
      );
}

class AuthSession {
  AuthSession({required this.tokenStore});

  final TokenStore tokenStore;

  final user = Signal<SessionUser?>(null);
  final accessToken = Signal<String?>(null);
  final expired = Signal(false);

  late final Computed<bool> signedIn = Computed(() => user.value != null);

  /// App 啟動時讀取持久化 token 與使用者（D5）。
  Future<void> restore() async {
    accessToken.value = await tokenStore.readAccessToken();
  }

  Future<void> signIn(SessionUser u, {required String access, required String refresh}) async {
    await tokenStore.writeTokens(access: access, refresh: refresh);
    accessToken.value = access;
    user.value = u;
    expired.value = false;
  }

  Future<void> signOut() async {
    await tokenStore.clear();
    user.value = null;
    accessToken.value = null;
  }

  /// 403（tv 不符，D5）→ 通知根層登出。
  void markExpired() => expired.value = true;
}
```

```dart
// app/lib/core/auth/auth_refresh_controller.dart
import 'dart:async';

/// 401 refresh 單飛鎖：並發請求共享一次 refresh（設計 §9）。
class AuthRefreshController {
  AuthRefreshController({required this.doRefresh, required this.onRotated});

  final Future<bool> Function() doRefresh;
  final void Function() onRotated;

  Future<bool>? _inFlight;

  Future<bool> refresh() => _inFlight ??= _run();

  Future<bool> _run() async {
    try {
      final ok = await doRefresh();
      if (ok) onRotated();
      return ok;
    } finally {
      _inFlight = null;
    }
  }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd app && flutter test test/core/auth`
Expected: PASS（4 個測試：token round-trip、session 計算、單飛鎖、失敗重試）。

- [ ] **Step 5: Commit**

```bash
git add app/lib/core/auth app/test/core/auth && git commit -m "feat(app): token store, auth session, single-flight refresh"
```

---

### Task 6: core/network — AuthInterceptor 與 Transport factory

**Files:**
- Create: `app/lib/core/network/auth_interceptor.dart`
- Create: `app/lib/core/network/app_transport_factory.dart`
- Test: `app/test/core/network/auth_interceptor_test.dart`、`app/test/core/network/app_transport_factory_test.dart`

**Interfaces:**
- Consumes: Task 5 的 `AuthSession`（`accessToken` signal、`markExpired`）、`AuthRefreshController.refresh()`、Task 2 的 `AppConfig.apiBaseUrl`。
- Produces:
  - `class AuthInterceptor { AuthInterceptor({required this.tokenProvider, required this.refresh, required this.onExpired}); final Future<String?> Function() tokenProvider; final Future<bool> Function() refresh; final void Function() onExpired; Interceptor call(); }` — 每個請求附加 `authorization: Bearer <token>`；`ConnectException(Code.unauthenticated)` 且有 token 時 refresh 後重試一次；`unauthenticated` 無 token 或 refresh 失敗、及 `permissionDenied` → `onExpired()` 並 rethrow。
  - `class AppTransportFactory { AppTransportFactory({required this.apiBaseUrl, required this.authInterceptor, http.Client? httpClient}); Transport createAuthenticated(); Transport createUnauthenticated(); }` — 用 `package:connectrpc/protocol/connect.dart` 的 `Transport` + `ProtoCodec` + `createHttpClient()`（`package:connectrpc/io.dart`）；unauth 版本（登入 / refresh / 註冊完成用）**不含** AuthInterceptor（避免 refresh 自我遞迴）。
- 後續 generated clients 以 `createAuthenticated()` 建構（Phase 1.10+）。

- [ ] **Step 1: 寫 failing test**

```dart
// test/core/network/auth_interceptor_test.dart
import 'package:connectrpc/connect.dart' as connect;
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/network/auth_interceptor.dart';

const spec = connect.Spec<int, String>(
  '/test/TestService/Echo',
  connect.StreamType.unary,
  () => 0,
  () => '',
);

void main() {
  test('attaches Bearer header', () async {
    connect.Headers? seen;
    final interceptor = AuthInterceptor(
      tokenProvider: () async => 'tok1',
      refresh: () async => true,
      onExpired: () {},
    ).call();
    final res = await interceptor((req) async {
      seen = (req as connect.UnaryRequest<int, String>).headers;
      return connect.UnaryResponse<int, String>(spec, connect.Headers(), 'ok', connect.Headers());
    })(connect.UnaryRequest<int, String>(spec, 1, Uri.parse('https://test.local'), headers: connect.Headers()));
    expect(seen!['authorization'], 'Bearer tok1');
    expect(res.message, 'ok');
  });

  test('401 triggers single refresh then retries once', () async {
    var calls = 0;
    var refreshCalls = 0;
    final interceptor = AuthInterceptor(
      tokenProvider: () async => calls == 0 ? 'expired' : 'fresh',
      refresh: () async { refreshCalls++; return true; },
      onExpired: () {},
    ).call();
    final res = await interceptor((req) async {
      calls++;
      if (calls == 1) throw connect.ConnectException(connect.Code.unauthenticated, 'expired');
      return connect.UnaryResponse<int, String>(spec, connect.Headers(), 'ok', connect.Headers());
    })(connect.UnaryRequest<int, String>(spec, 1, Uri.parse('https://test.local'), headers: connect.Headers()));
    expect(calls, 2);
    expect(refreshCalls, 1);
    expect(res.message, 'ok');
  });

  test('401 without token calls onExpired and rethrows', () async {
    var expired = 0;
    final interceptor = AuthInterceptor(
      tokenProvider: () async => null,
      refresh: () async => true,
      onExpired: () => expired++,
    ).call();
    await expectLater(
      interceptor((req) async => throw connect.ConnectException(connect.Code.unauthenticated, 'nope'))(
        connect.UnaryRequest<int, String>(spec, 1, Uri.parse('https://test.local'), headers: connect.Headers()),
      ),
      throwsA(isA<connect.ConnectException>()),
    );
    expect(expired, 1);
  });

  test('permissionDenied calls onExpired', () async {
    var expired = 0;
    final interceptor = AuthInterceptor(
      tokenProvider: () async => 'tok',
      refresh: () async => true,
      onExpired: () => expired++,
    ).call();
    await expectLater(
      interceptor((req) async => throw connect.ConnectException(connect.Code.permissionDenied, 'tv mismatch'))(
        connect.UnaryRequest<int, String>(spec, 1, Uri.parse('https://test.local'), headers: connect.Headers()),
      ),
      throwsA(isA<connect.ConnectException>()),
    );
    expect(expired, 1);
  });
}
```

```dart
// test/core/network/app_transport_factory_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/network/app_transport_factory.dart';
import 'package:sales_order_app/core/network/auth_interceptor.dart';

void main() {
  test('factory builds authenticated and unauthenticated transports', () {
    final f = AppTransportFactory(
      apiBaseUrl: 'https://api.example.com',
      authInterceptor: AuthInterceptor(
        tokenProvider: () async => null,
        refresh: () async => false,
        onExpired: () {},
      ),
    );
    expect(f.createAuthenticated(), isNotNull);
    expect(f.createUnauthenticated(), isNotNull);
  });
}
```

> 註：`UnaryRequest` 建構式參數順序以 IDE 簽章提示為準（欄位：`spec`、`message`、`url`、`headers`、`signal`）；`UnaryResponse` 為 `(spec, headers, message, trailers)`。

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd app && flutter test test/core/network`
Expected: FAIL — 型別未定義。

- [ ] **Step 3: 實作兩檔**

```dart
// app/lib/core/network/auth_interceptor.dart
import 'package:connectrpc/connect.dart';

/// connect interceptor：Bearer 附加 + 401 refresh 重試 + 403 登出（D5）。
class AuthInterceptor {
  AuthInterceptor({
    required this.tokenProvider,
    required this.refresh,
    required this.onExpired,
  });

  final Future<String?> Function() tokenProvider;
  final Future<bool> Function() refresh;
  final void Function() onExpired;

  Interceptor call() => <I extends Object, O extends Object>(next) {
        return (req) async {
          final token = await tokenProvider();
          final authed = _withAuth(req, token);
          try {
            return await next(authed);
          } on ConnectException catch (e) {
            if (e.code == Code.unauthenticated) {
              if (token != null && await refresh()) {
                final fresh = await tokenProvider();
                return next(_withAuth(req, fresh));
              }
              onExpired();
            } else if (e.code == Code.permissionDenied) {
              onExpired();
            }
            rethrow;
          }
        };
      };

  Request<I, O> _withAuth<I, O>(Request<I, O> req, String? token) {
    if (token == null) return req;
    if (req case UnaryRequest<I, O>(:final spec, :final message, :final url, :final headers, :final signal)) {
      return UnaryRequest<I, O>(spec, message, url,
          headers: Headers.from(headers)..['authorization'] = 'Bearer $token',
          signal: signal);
    }
    if (req case StreamRequest<I, O>(:final spec, :final message, :final url, :final headers, :final signal)) {
      return StreamRequest<I, O>(spec, message, url,
          headers: Headers.from(headers)..['authorization'] = 'Bearer $token',
          signal: signal);
    }
    return req;
  }
}
```

```dart
// app/lib/core/network/app_transport_factory.dart
import 'package:connectrpc/connect.dart';
import 'package:connectrpc/io.dart' show createHttpClient;
import 'package:connectrpc/protobuf.dart' show ProtoCodec;
import 'package:connectrpc/protocol/connect.dart' as protocol;
import 'package:http/http.dart' as http;
import 'auth_interceptor.dart';

/// 業務 Transport factory（設計 §9）：認證版含 AuthInterceptor；
/// 免認證版供登入 / refresh / 註冊完成使用（避免 refresh 經 AuthInterceptor 自我遞迴）。
class AppTransportFactory {
  AppTransportFactory({
    required this.apiBaseUrl,
    required this.authInterceptor,
    http.Client? httpClient,
  }) : _httpClient = httpClient;

  final String apiBaseUrl;
  final AuthInterceptor authInterceptor;
  final http.Client? _httpClient;

  Transport createAuthenticated() => protocol.Transport(
        baseUrl: apiBaseUrl,
        codec: const ProtoCodec(),
        httpClient: _httpClient ?? createHttpClient(),
        interceptors: [authInterceptor.call()],
      );

  Transport createUnauthenticated() => protocol.Transport(
        baseUrl: apiBaseUrl,
        codec: const ProtoCodec(),
        httpClient: _httpClient ?? createHttpClient(),
      );
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd app && flutter test test/core/network`
Expected: PASS（5 個測試）。若 `Headers.from` / 建構式簽章有出入，以套件原始碼為準修正本步驟（欄位不變，僅順序/工廠方法）。

- [ ] **Step 5: Commit**

```bash
git add app/lib/core/network app/test/core/network && git commit -m "feat(app): auth interceptor and transport factory"
```

---

### Task 7: router 骨架（auto_route + guards）

**Files:**
- Create: `app/lib/router/route_access.dart`
- Create: `app/lib/router/guards.dart`
- Create: `app/lib/router/app_router.dart`
- Create: `app/lib/features/auth/presentation/login_page.dart`
- Create: `app/lib/features/shell/presentation/shell_page.dart`
- Create: `app/lib/features/home/presentation/home_page.dart`、`app/lib/features/products/presentation/products_page.dart`、`app/lib/features/orders/presentation/orders_page.dart`、`app/lib/features/profile/presentation/profile_page.dart`（4 tab 佔位頁）
- Test: `app/test/router/route_access_test.dart`

**Interfaces:**
- Consumes: Task 5 `AuthSession`（`signedIn`、`user.value!.isPrimary`）。
- Produces:
  - `bool canAccessRoute({required bool signedIn, required bool isPrimaryAccount, required bool isBusinessRoute})`（純函式，route 守衛決策）。
  - `class AuthGuard implements AutoRouteGuard { AuthGuard(this.session); final AuthSession session; }` — 未登入 → redirect `/login`。
  - `class RoleGuard implements AutoRouteGuard { RoleGuard(this.session); final AuthSession session; }` — 主帳號進業務路由 → redirect `/profile`（D22/D28）。
  - `class AppRouter extends RootStackRouter`（`@AutoRouterConfig`，建構式注入 `AuthSession`）— routes：`LoginRoute('/login')`、`ShellRoute('/shell', guards: [AuthGuard], children: 4 tabs：Home/Products/Orders/Profile)`、業務 tab（Products/Orders）額外掛 `RoleGuard`、`ProfileRoute('/profile')`。
  - 佔位頁皆 `@RoutePage()`、繁中標題。

- [ ] **Step 1: 寫 failing test（純決策函式）**

```dart
// test/router/route_access_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/router/route_access.dart';

void main() {
  test('未登入不可進任何受保護路由', () {
    expect(canAccessRoute(signedIn: false, isPrimaryAccount: false, isBusinessRoute: false), isFalse);
    expect(canAccessRoute(signedIn: false, isPrimaryAccount: false, isBusinessRoute: true), isFalse);
  });

  test('一般帳號可進業務路由', () {
    expect(canAccessRoute(signedIn: true, isPrimaryAccount: false, isBusinessRoute: true), isTrue);
  });

  test('主帳號僅管理、不可進業務路由（D22/D28）', () {
    expect(canAccessRoute(signedIn: true, isPrimaryAccount: true, isBusinessRoute: true), isFalse);
    expect(canAccessRoute(signedIn: true, isPrimaryAccount: true, isBusinessRoute: false), isTrue);
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd app && flutter test test/router/route_access_test.dart`
Expected: FAIL — 檔案不存在。

- [ ] **Step 3: 實作 route_access.dart、guards.dart、app_router.dart 與佔位頁**

```dart
// app/lib/router/route_access.dart
/// 路由守衛純決策（可獨立測試；guards 只是薄包裝）。
bool canAccessRoute({
  required bool signedIn,
  required bool isPrimaryAccount,
  required bool isBusinessRoute,
}) {
  if (!signedIn) return false;
  if (isPrimaryAccount && isBusinessRoute) return false;
  return true;
}
```

```dart
// app/lib/router/guards.dart
import 'package:auto_route/auto_route.dart';
import '../core/auth/auth_session.dart';
import 'app_router.dart';
import 'route_access.dart';

/// 未登入一律回登入頁。
class AuthGuard extends AutoRouteGuard {
  AuthGuard(this.session);
  final AuthSession session;

  @override
  Future<void> onNavigate(NavigationResolver resolver, StackRouter router) async {
    if (session.signedIn.value) {
      resolver.next();
    } else {
      resolver.redirect(const LoginRoute());
    }
  }
}

/// 主帳號僅可進管理/功能頁，業務路由擋下（D22/D28）。
class RoleGuard extends AutoRouteGuard {
  RoleGuard(this.session);
  final AuthSession session;

  @override
  Future<void> onNavigate(NavigationResolver resolver, StackRouter router) async {
    final allowed = canAccessRoute(
      signedIn: session.signedIn.value,
      isPrimaryAccount: session.user.value?.isPrimary ?? false,
      isBusinessRoute: true,
    );
    if (allowed) {
      resolver.next();
    } else {
      resolver.redirect(const ProfileRoute());
    }
  }
}
```

```dart
// app/lib/router/app_router.dart
import 'package:auto_route/auto_route.dart';
import '../core/auth/auth_session.dart';
import '../features/auth/presentation/login_page.dart';
import '../features/home/presentation/home_page.dart';
import '../features/orders/presentation/orders_page.dart';
import '../features/products/presentation/products_page.dart';
import '../features/profile/presentation/profile_page.dart';
import '../features/shell/presentation/shell_page.dart';
import 'guards.dart';

part 'app_router.gr.dart';

@AutoRouterConfig()
class AppRouter extends RootStackRouter {
  AppRouter({required this.session});

  final AuthSession session;

  @override
  List<AutoRoute> get routes => [
        AutoRoute(page: LoginRoute.page, initial: true),
        AutoRoute(
          page: ShellRoute.page,
          guards: [AuthGuard(session)],
          children: [
            AutoRoute(page: HomeRoute.page, path: 'home'),
            AutoRoute(
              page: ProductsRoute.page,
              path: 'products',
              guards: [RoleGuard(session)],
            ),
            AutoRoute(
              page: OrdersRoute.page,
              path: 'orders',
              guards: [RoleGuard(session)],
            ),
            AutoRoute(page: ProfileRoute.page, path: 'profile'),
          ],
        ),
      ];
}
```

```dart
// app/lib/features/auth/presentation/login_page.dart
import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

@RoutePage()
class LoginPage extends StatelessWidget {
  const LoginPage({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('登入')),
        body: const Center(child: Text('我是店家 / 我是業務（身分選擇）')),
      );
}
```

```dart
// app/lib/features/shell/presentation/shell_page.dart
import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

@RoutePage()
class ShellPage extends StatelessWidget {
  const ShellPage({super.key});

  @override
  Widget build(BuildContext context) => AutoTabsScaffold(
        routes: const [
          HomeRoute(),
          ProductsRoute(),
          OrdersRoute(),
          ProfileRoute(),
        ],
        bottomNavigationBuilder: (_, tabsRouter) => NavigationBar(
          selectedIndex: tabsRouter.activeIndex,
          onDestinationSelected: tabsRouter.setActiveIndex,
          destinations: const [
            NavigationDestination(icon: Icon(Icons.home), label: '首頁'),
            NavigationDestination(icon: Icon(Icons.shopping_cart), label: '商品'),
            NavigationDestination(icon: Icon(Icons.receipt_long), label: '訂單'),
            NavigationDestination(icon: Icon(Icons.menu), label: '功能'),
          ],
        ),
      );
}
```

佔位頁（Home / Products / Orders / Profile）同一模式：
```dart
// app/lib/features/home/presentation/home_page.dart
import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

@RoutePage()
class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('首頁')),
        body: const Center(child: Text('公告消息（Phase 6.1 實作）')),
      );
}
```
（Products / Orders / Profile 頁面同構，標題分別為「商品」「訂單」「功能」，內容佔位文案「快速下單（Phase 4.6 實作）」「訂單歷史（Phase 6.2 實作）」「客戶查詢 / 設定 / 登出（Phase 6.3 實作）」。）

- [ ] **Step 4: 產生路由碼**

Run: `cd app && dart run build_runner build --delete-conflicting-outputs`
Expected: 產生 `app/lib/router/app_router.gr.dart`，無錯誤。

- [ ] **Step 5: Run test to verify it passes + 靜態分析**

Run: `cd app && flutter test test/router/route_access_test.dart && flutter analyze`
Expected: PASS（3 個測試）+ 0 分析錯誤。

- [ ] **Step 6: Commit**

```bash
git add app/lib/router app/lib/features && git commit -m "feat(app): auto_route skeleton with auth and role guards"
```

---

### Task 8: 根佈線（CacheProvider + disco ProviderScope + app.dart/main.dart）

**Files:**
- Create: `app/lib/app.dart`
- Modify: `app/lib/main.dart`
- Test: `app/test/app_test.dart`

**Interfaces:**
- Consumes: Task 3 `QueryCacheFactory`、Task 4 `SembastCacheRepository`、Task 5 `AuthSession`、Task 7 `AppRouter`。
- Produces:
  - `class SalesOrderApp extends StatelessWidget { SalesOrderApp({required this.session, required this.queryCache, required this.router}); ... }` — 樹：`CacheProvider(cache:)` > disco `ProviderScope`（根 providers）> `ProviderPortal` > `MaterialApp.router(router:)`。
  - disco 根 providers 定義於 `app/lib/app.dart`：`final appConfigProvider = Provider((_) => AppConfig.fromEnv());`。**session 相依的 providers（TokenStore / AppTransportFactory / AuthInterceptor / AuthRefreshController）於 Phase 1.10 建 auth scope 時加入**——本 Task 以建構式注入單一實例，避免跨 provider 以 context 閉包互相引用（disco 反模式）。
  - `main.dart`：`WidgetsFlutterBinding.ensureInitialized()` → 建 `AuthSession`、`restore()`、建 `QueryCache`、建 `SembastCacheRepository`、建 `AppRouter` → `runApp(RootExpiryListener(...))`。
  - 根層以 solidart `Effect` 監聽 `session.expired`：觸發 → `queryCache.removeQueries(QueryKey([]))`（全清）+ 鏡像 `clear()` + `session.signOut()` + `router.replaceAll([LoginRoute()])`（設計 §8.4 / §9）。Effect 於 `StatefulWidget` state 中建立並在 dispose 時取消。

- [ ] **Step 1: 寫 failing widget test**

```dart
// test/app_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/app.dart';
import 'package:sales_order_app/core/auth/auth_session.dart';
import 'package:sales_order_app/core/auth/token_store.dart';
import 'package:sales_order_app/core/cache/query_cache_factory.dart';
import 'package:sales_order_app/router/app_router.dart';

class InMemoryTokenStore implements TokenStore {
  String? access; String? refresh;
  @override Future<String?> readAccessToken() async => access;
  @override Future<String?> readRefreshToken() async => refresh;
  @override Future<void> writeTokens({required String access, required String refresh}) async { this.access = access; this.refresh = refresh; }
  @override Future<void> clear() async { access = null; refresh = null; }
}

void main() {
  testWidgets('未登入顯示登入頁', (tester) async {
    final session = AuthSession(tokenStore: InMemoryTokenStore());
    final router = AppRouter(session: session);
    await tester.pumpWidget(SalesOrderApp(
      session: session,
      queryCache: QueryCacheFactory.create(),
      router: router,
    ));
    await tester.pumpAndSettle();
    expect(find.text('我是店家 / 我是業務（身分選擇）'), findsOneWidget);
  });

  testWidgets('登入後可進入 shell 首頁 tab', (tester) async {
    final session = AuthSession(tokenStore: InMemoryTokenStore());
    await session.signIn(
      const SessionUser(userId: 'u1', roleCode: 'customer', isPrimary: false),
      access: 'a', refresh: 'r',
    );
    final router = AppRouter(session: session);
    await tester.pumpWidget(SalesOrderApp(
      session: session,
      queryCache: QueryCacheFactory.create(),
      router: router,
    ));
    await tester.pumpAndSettle();
    // 登入頁完成登入後導向 shell（真實流程於 Phase 1.10 的 LoginPage 執行）
    router.replaceAll(const [ShellRoute()]);
    await tester.pumpAndSettle();
    expect(find.text('公告消息（Phase 6.1 實作）'), findsOneWidget);
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd app && flutter test test/app_test.dart`
Expected: FAIL — `SalesOrderApp` 未定義。

- [ ] **Step 3: 實作 app.dart 與 main.dart**

```dart
// app/lib/app.dart
import 'package:auto_route/auto_route.dart';
import 'package:disco/disco.dart';
import 'package:fquery/fquery.dart';
import 'package:flutter/material.dart';
import 'core/auth/auth_session.dart';
import 'core/config/app_config.dart';
import 'router/app_router.dart';

/// 根 providers（disco，設計 §6）。session 相依的 providers（TokenStore /
/// TransportFactory / AuthInterceptor / AuthRefreshController）於 Phase 1.10
/// 建 auth scope 時加入；此處僅放可自足建構者，避免跨 provider 以 context
/// 閉包相互引用。
final appConfigProvider = Provider((_) => AppConfig.fromEnv());

class SalesOrderApp extends StatelessWidget {
  const SalesOrderApp({
    super.key,
    required this.session,
    required this.queryCache,
    required this.router,
  });

  final AuthSession session;
  final QueryCache queryCache;
  final AppRouter router;

  @override
  Widget build(BuildContext context) {
    return CacheProvider(
      cache: queryCache,
      child: ProviderScope(
        providers: [appConfigProvider],
        child: ProviderPortal(
          child: MaterialApp.router(
            title: '訂出貨系統',
            theme: ThemeData(colorSchemeSeed: Colors.indigo, useMaterial3: true),
            routerConfig: router.config(),
          ),
        ),
      ),
    );
  }
}
```

```dart
// app/lib/main.dart
import 'package:flutter/material.dart';
import 'package:solidart/solidart.dart';
import 'app.dart';
import 'core/auth/auth_session.dart';
import 'core/auth/token_store.dart';
import 'core/cache/cache_repository.dart';
import 'core/cache/query_cache_factory.dart';
import 'router/app_router.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final session = AuthSession(tokenStore: SecureTokenStore());
  await session.restore(); // D5：啟動讀回 token

  final queryCache = QueryCacheFactory.create();
  final cacheRepository = SembastCacheRepository(
    databaseFactory: sembastIoDatabaseFactory,
    path: 'query_cache.db',
  );
  final router = AppRouter(session: session);

  runApp(RootExpiryListener(
    session: session,
    queryCache: queryCache,
    cacheRepository: cacheRepository,
    router: router,
    child: SalesOrderApp(session: session, queryCache: queryCache, router: router),
  ));
}

/// 根層監聽 session.expired（403 tv 不符）→ 清快取 + 回登入頁（設計 §8.4/§9）。
class RootExpiryListener extends StatefulWidget {
  const RootExpiryListener({
    super.key,
    required this.session,
    required this.queryCache,
    required this.cacheRepository,
    required this.router,
    required this.child,
  });

  final AuthSession session;
  final QueryCache queryCache;
  final CacheRepository cacheRepository;
  final AppRouter router;
  final Widget child;

  @override
  State<RootExpiryListener> createState() => _RootExpiryListenerState();
}

class _RootExpiryListenerState extends State<RootExpiryListener> {
  Effect? _effect;

  @override
  void initState() {
    super.initState();
    _effect = Effect(() {
      if (widget.session.expired.value) {
        widget.queryCache.removeQueries(QueryKey([]));
        widget.cacheRepository.clear();
        widget.session.signOut();
        widget.router.replaceAll([const LoginRoute()]);
      }
    });
  }

  @override
  void dispose() {
    _effect?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
```

> 註：`sembastIoDatabaseFactory` 自 `package:sembast_io/sembast_io.dart` 匯出；`QueryKey([])` 為 fquery 全清慣例（前綴匹配空 key）。`LoginRoute` 由 Task 7 codegen 產生（`app_router.gr.dart`）。

- [ ] **Step 4: Run test to verify it passes**

Run: `cd app && flutter test test/app_test.dart && flutter analyze`
Expected: PASS（2 個 widget 測試）+ 0 分析錯誤。若 disco Provider 建構需要非同步（本設計 providers 皆同步），保持同步。

- [ ] **Step 5: Commit**

```bash
git add app/lib/app.dart app/lib/main.dart app/test/app_test.dart && git commit -m "feat(app): root wiring with CacheProvider, disco scopes, expiry listener"
```

---

### Task 9: fquery 資料層慣例（QueryKeys + seed + 範例 repository）

**Files:**
- Create: `app/lib/core/cache/query_keys.dart`
- Create: `app/lib/core/cache/cache_bootstrap.dart`
- Create: `app/lib/features/home/data/announcements_repository.dart`
- Test: `app/test/core/cache/cache_bootstrap_test.dart`、`app/test/features/home/data/announcements_repository_test.dart`

**Interfaces:**
- Consumes: Task 3 `QueryCache`、Task 4 `CacheRepository`。
- Produces:
  - `class QueryKeys { static const customers = QueryKey(['customers']); ... }`（設計 §8.1 全表）+ `static const mirrorable = <QueryKey>[...]`（customers / customer-products / products / orders / announcements）。
  - `String queryKeyId(QueryKey key)` — 鏡像 key 的字串化（`key.toString()`；若 QueryKey 無規範 toString，以 `key.keys.join('/')` 取代，兩處使用點一致即可）。
  - `Future<void> seedQueryCacheFromMirror({required QueryCache cache, required CacheRepository mirror, required Iterable<QueryKey> keys})` — 逐 key 讀鏡像，有資料即以 `cache.setQueryData` 回填（設計 §8.4 seed）。
  - `class AnnouncementsRepository { AnnouncementsRepository({required this.fetcher, required this.cacheRepository}); final Future<List<Map<String, Object?>>> Function() fetcher; final CacheRepository cacheRepository; Future<List<Map<String, Object?>>> fetchAnnouncements(QueryCache cache); }` — queryFn 慣例：fetcher 成功 → write-through 鏡像；拋出離線 fallback 由 UI 以 seed 資料顯示（§8.4）。
- 後續 feature repositories（Phase 4.6 / 6.x）依此模式：queryFn 包 generated client + write-through；mutation 後 `invalidateQueries`（設計 §8.3 對照表）。

- [ ] **Step 1: 寫 failing test**

```dart
// test/core/cache/cache_bootstrap_test.dart
import 'package:fquery/fquery.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sembast/sembast_memory.dart';
import 'package:sales_order_app/core/cache/cache_bootstrap.dart';
import 'package:sales_order_app/core/cache/cache_repository.dart';
import 'package:sales_order_app/core/cache/query_cache_factory.dart';
import 'package:sales_order_app/core/cache/query_keys.dart';

void main() {
  testWidgets('seed 後 QueryBuilder 直接顯示鏡像資料、不呼叫 queryFn', (tester) async {
    final repo = SembastCacheRepository(databaseFactory: databaseFactoryMemory, path: 't.db');
    await repo.clear();
    await repo.write(queryKeyId(QueryKeys.announcements), {
      'items': [
        {'title': '中秋公休'},
      ],
    });

    final cache = QueryCacheFactory.create();
    await seedQueryCacheFromMirror(cache: cache, mirror: repo, keys: QueryKeys.mirrorable);

    var fetchCalls = 0;
    await tester.pumpWidget(CacheProvider(
      cache: cache,
      child: MaterialApp(
        home: QueryBuilder<List<Map<String, Object?>>, Exception>(
          options: QueryOptions(
            queryKey: QueryKeys.announcements,
            queryFn: () async {
              fetchCalls++;
              return [];
            },
            refetchOnMount: RefetchOnMount.never,
          ),
          builder: (context, state) => Text(state.data?[0]['title']?.toString() ?? 'empty'),
        ),
      ),
    ));
    await tester.pumpAndSettle();

    expect(fetchCalls, 0); // 未重抓
    expect(find.text('中秋公休'), findsOneWidget); // 顯示鏡像 seed 資料
  });
}
```

```dart
// test/features/home/data/announcements_repository_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:sembast/sembast_memory.dart';
import 'package:sales_order_app/core/cache/cache_repository.dart';
import 'package:sales_order_app/core/cache/query_keys.dart';
import 'package:sales_order_app/features/home/data/announcements_repository.dart';

void main() {
  test('fetchAnnouncements writes through to the mirror', () async {
    final repo = SembastCacheRepository(databaseFactory: databaseFactoryMemory, path: 't.db');
    await repo.clear();
    final repository = AnnouncementsRepository(
      fetcher: () async => [
        {'title': '中秋公休'},
      ],
      cacheRepository: repo,
    );

    final data = await repository.fetchAnnouncements();
    expect(data, hasLength(1));

    final mirrored = await repo.read(queryKeyId(QueryKeys.announcements));
    expect(mirrored, isNotNull);
    expect((mirrored!['items'] as List), hasLength(1));
  });
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd app && flutter test test/core/cache/cache_bootstrap_test.dart test/features/home/data/announcements_repository_test.dart`
Expected: FAIL — 型別未定義。

- [ ] **Step 3: 實作三檔**

```dart
// app/lib/core/cache/query_keys.dart
import 'package:fquery/fquery.dart';

/// 全 App QueryKey 目錄（設計 §8.1）。
class QueryKeys {
  QueryKeys._();

  static const customers = QueryKey(['customers']);
  static const customerProducts = QueryKey(['customer-products']);
  static const products = QueryKey(['products']);
  static const orders = QueryKey(['orders']);
  static const announcements = QueryKey(['announcements']);
  static const notifications = QueryKey(['notifications']);
  static const metadicts = QueryKey(['metadicts']);
  static const routes = QueryKey(['routes']);
  static const warehouses = QueryKey(['warehouses']);
  static const returns = QueryKey(['returns']);

  /// 可離線鏡像清單（設計 §8.1 標 ✅）。
  static const mirrorable = <QueryKey>[
    customers,
    customerProducts,
    products,
    orders,
    announcements,
  ];
}

/// 鏡像 key 字串化（CacheRepository 使用）。
String queryKeyId(QueryKey key) => key.toString();
```

```dart
// app/lib/core/cache/cache_bootstrap.dart
import 'package:fquery/fquery.dart';
import 'cache_repository.dart';
import 'query_keys.dart';

/// App 啟動 seed：有鏡像資料即以 setQueryData 回填（設計 §8.4），UI 立即顯示並標 stale。
Future<void> seedQueryCacheFromMirror({
  required QueryCache cache,
  required CacheRepository mirror,
  required Iterable<QueryKey> keys,
}) async {
  for (final key in keys) {
    final data = await mirror.read(queryKeyId(key));
    if (data != null) {
      cache.setQueryData(key, (_) => data);
    }
  }
}
```

```dart
// app/lib/features/home/data/announcements_repository.dart
import 'package:fquery/fquery.dart';
import '../../../core/cache/cache_repository.dart';
import '../../../core/cache/query_keys.dart';

/// 範例 repository：queryFn 慣例 = fetcher + write-through 鏡像（設計 §8.4）。
/// 後續 feature repos（Phase 4.6 / 6.x）同一模式，fetcher 換成 generated client 呼叫。
class AnnouncementsRepository {
  AnnouncementsRepository({required this.fetcher, required this.cacheRepository});

  final Future<List<Map<String, Object?>>> Function() fetcher;
  final CacheRepository cacheRepository;

  /// 給 QueryBuilder 的 queryFn：成功即寫入鏡像。
  QueryOptions<List<Map<String, Object?>>, Exception> queryOptions() => QueryOptions(
        queryKey: QueryKeys.announcements,
        queryFn: fetchAnnouncements,
      );

  Future<List<Map<String, Object?>>> fetchAnnouncements() async {
    final data = await fetcher();
    await cacheRepository.write(queryKeyId(QueryKeys.announcements), {'items': data});
    return data;
  }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd app && flutter test test/core/cache/cache_bootstrap_test.dart test/features/home/data/announcements_repository_test.dart && flutter analyze`
Expected: PASS（2 個測試）+ 0 分析錯誤。

- [ ] **Step 5: Commit**

```bash
git add app/lib/core/cache/query_keys.dart app/lib/core/cache/cache_bootstrap.dart app/lib/features/home/data app/test && git commit -m "feat(app): fquery query keys, mirror seed, and repository pattern"
```

---

### Task 10: 文件回寫（D29 / 規格書 v1.0.35 / 主計畫 v2.10.0）

**Files:**
- Modify: `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（新增 D29 條目，內容取自設計文件 §2）
- Modify: `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（升版 v1.0.35：§2.1 App 技術列、§9.6 離線快取實作輪廓、§16 後續步驟、§18 修訂記錄）
- Modify: `docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md`（升版 v2.10.0：Task 0.6 以本計畫 Task 1–9 取代；Task 1.10 / 4.6 / 6.4 之 App 棧引用更新；加註本計畫為來源）

**Interfaces:**
- 產出：決策記錄 D29（含 D29-2 持久化唯讀快取、D29-3 Flutter 升版）；規格書 v1.0.35 修訂記錄；主計畫 v2.10.0 版本行。

- [ ] **Step 1: 決策記錄新增 D29**

於 `2026-07-19-sales-order-1.0-decisions.md` 追加：
```markdown
### D29：App 技術棧定案（solidart / disco / auto_route / fquery / Sembast 鏡像）

- **選擇**：App 狀態管理 solidart（+flutter_solidart）、DI disco scoped providers、路由 auto_route 11、資料層 fquery 3（server state）+ Sembast 唯讀快取鏡像；**不採用 dio_cache_interceptor**（Connect-RPC 全 POST 常數 URL 與其 GET 預設快取語意不合，且需自養 Dio→http.Client 橋接層、無 query/mutation state）。Flutter 升 stable（Dart ≥3.10，disco 2.0.0 要求）。
- **理由**：fquery 為 transport-agnostic，直接呼叫 connect-dart generated client，零協議摩擦；mutation/invalidation 語意與下單、退貨、帳號管理流程對應；Sembast write-through 滿足 §9.6 持久化唯讀快取。
- **已考慮 alternative**：dio_cache_interceptor（協議硬傷，見設計文件 §3 比較表）；solidart Resource 手刻（無跨頁面 cache/dedup/invalidate）。
- 修訂來源：2026-08-04 設計文件 `docs/superpowers/specs/2026-08-04-app-flutter-stack-design.md`。
```

- [ ] **Step 2: 規格書升版 v1.0.35**

§2.1 App 列改為：`Flutter stable（Dart ≥3.10）+ solidart + disco + auto_route + fquery + Sembast + flutter_secure_storage + connectrpc`；§9.6 補一句：快取以 fquery QueryCache + Sembast 鏡像實作（write-through、啟動 seed、登出清空）；§16 後續步驟指向 D29 設計文件；§18 補 v1.0.35 列。

- [ ] **Step 3: 主計畫升版 v2.10.0**

版本行加註；Task 0.6 內容替換為引用本計畫 Task 1–9；Task 1.10 / 4.6 / 6.1–6.4 的 App 棧表述（若有寫死 Sembast 用法或棧）更新為 D29。

- [ ] **Step 4: 驗收（文件一致性）**

Run: 以 grep 確認三份文件皆出現「D29」且規格書出現「v1.0.35」、主計畫出現「v2.10.0」。
Expected: 各命中 ≥1 處。

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers && git commit -m "docs: record D29 app stack decision, spec v1.0.35, plan v2.10.0"
```

---

## Self-Review 註記（寫作時已執行）

- **Spec coverage（設計文件 §5–§11）**：§5 分層 → Task 1/2/7/8；§6 狀態與 DI → Task 5/8；§7 路由 → Task 7；§8 資料層 → Task 3/4/9；§9 認證 → Task 5/6/8；§10 錯誤 → Task 2；§11 測試 → 各 Task 測試；§12 文件回寫 → Task 10。
- **Placeholder scan**：無 TBD/TODO；兩處 API 簽章（`UnaryRequest` 建構式、`QueryKey.toString`）標註以套件原始碼為準並給定替代方案，非留白。
- **Type consistency**：`AuthSession.signedIn`（Computed）於 Task 7/8 使用一致；`CacheRepository` 介面於 Task 4/8/9 一致；`queryKeyId` 於 Task 9 兩處使用一致；`AppTransportFactory.createAuthenticated()/createUnauthenticated()` 於 Task 6 定義，消費點為 Phase 1.10（auth scope 佈線時以 `createUnauthenticated()` 實作 refresh 呼叫，避免 AuthInterceptor 自我遞迴）。
