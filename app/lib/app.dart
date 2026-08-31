import 'package:flutter/material.dart';

import 'config.dart';
import 'features/auth/auth_repository.dart';
import 'features/auth/auth_transport.dart';
import 'features/auth/token_storage.dart';
import 'router/app_router.dart';

/// 根 Widget。D29:後續在此掛 CacheProvider(fquery)+
/// 根 ProviderScope(disco)。
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
  late final AppRouter _router = AppRouter(authRepository: _authRepository);

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: '多公司訂出貨系統',
      routerConfig: _router.config(),
    );
  }
}
