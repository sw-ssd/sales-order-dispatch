import 'package:auto_route/auto_route.dart';

import '../core/api.dart';
import '../features/auth/auth_repository.dart';
import '../features/auth/pages/identity_select_page.dart';
import '../features/auth/pages/login_page.dart';
import '../features/shell/home_shell.dart';

/// App 根路由。
///
/// auto_route 11 支援手寫路由表(PageInfo.builder),此處直接在建構式注入
/// [AuthRepository] 與 [Api] 並以閉包傳給各頁;待 D29 後續任務(shell、guards)
/// 落地時再評估是否改用 auto_route_generator 產碼。
class AppRouter extends RootStackRouter {
  AppRouter({required AuthRepository authRepository, required Api api})
      : _authRepository = authRepository,
        _api = api;

  final AuthRepository _authRepository;
  final Api _api;

  @override
  List<AutoRoute> get routes => [
        AutoRoute(
          path: '/login',
          initial: true,
          page: PageInfo.builder(
            'IdentitySelectRoute',
            builder: (context, _) =>
                IdentitySelectPage(authRepository: _authRepository),
          ),
        ),
        AutoRoute(
          path: '/login/shop',
          page: PageInfo.builder(
            'ShopLoginRoute',
            builder: (context, _) =>
                LoginPage(authRepository: _authRepository),
          ),
        ),
        AutoRoute(
          path: '/home',
          page: PageInfo.builder(
            'HomeShellRoute',
            builder: (context, _) =>
                HomeShell(api: _api, auth: _authRepository),
          ),
        ),
      ];
}
