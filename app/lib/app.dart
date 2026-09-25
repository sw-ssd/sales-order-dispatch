import 'package:flutter/material.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import 'config.dart';
import 'core/api.dart';
import 'features/auth/auth_repository.dart';
import 'features/auth/auth_transport.dart';
import 'features/auth/token_storage.dart';
import 'gen/salesorder/v1/auth.connect.client.dart';
import 'router/app_router.dart';

/// 根 Widget。D29 根佈線：CacheProvider(fquery) 已掛；
/// 根 ProviderScope(disco) 仍屬後續任務。
class SalesOrderApp extends StatefulWidget {
  const SalesOrderApp({super.key, required this.config});

  final AppConfig config;

  @override
  State<SalesOrderApp> createState() => _SalesOrderAppState();
}

class _SalesOrderAppState extends State<SalesOrderApp> {
  final _tokenStorage = SecureTokenStorage();
  late final AuthRepository _authRepository = AuthRepository(
    client: createAuthServiceClient(widget.config.apiBaseUrl),
    // 帶身分的 AuthService：ChangePassword 是登入態 RPC，走未認證 transport 會被
    // middleware 判成未登入（AUTH-4001）。
    //
    // `refresh` 以閉包回指 `_authRepository`：`late final` 在第一次被讀取時才初始化，
    // 屆時本欄位已賦值（與 `Api(auth: _authRepository)` 同一手法，兩者都不會在初始化
    // 期間呼叫到它）。改寫成先建 repository 再補 transport 會需要第二個可變欄位。
    authed: AuthServiceClient(createAuthedTransport(
      baseUrl: widget.config.apiBaseUrl,
      accessToken: () async => (await _tokenStorage.read())?.accessToken,
      refresh: () => _authRepository.refreshTokens(),
    )),
    tokenStorage: _tokenStorage,
  );
  late final Api _api = Api(
    baseUrl: widget.config.apiBaseUrl,
    auth: _authRepository,
  );
  late final AppRouter _router =
      AppRouter(authRepository: _authRepository, api: _api);
  final QueryCache _queryCache = QueryCache();

  @override
  Widget build(BuildContext context) {
    return CacheProvider(
      cache: _queryCache,
      child: MaterialApp.router(
        title: '多公司訂出貨系統',
        routerConfig: _router.config(),
      ),
    );
  }
}
