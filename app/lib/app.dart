import 'package:flutter/material.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import 'config.dart';
import 'core/api.dart';
import 'features/auth/auth_repository.dart';
import 'features/auth/auth_transport.dart';
import 'features/auth/token_storage.dart';
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
  late final AuthRepository _authRepository = AuthRepository(
    client: createAuthServiceClient(widget.config.apiBaseUrl),
    tokenStorage: SecureTokenStorage(),
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
