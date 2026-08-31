import 'package:auto_route/auto_route.dart';

import '../features/auth/auth_repository.dart';
import '../features/auth/pages/identity_select_page.dart';
import '../features/auth/pages/login_page.dart';

/// App 根路由。
///
/// auto_route 11 支援手寫路由表(PageInfo.builder),此處直接在建構式注入
/// [AuthRepository] 並以閉包傳給各頁;待 D29 後續任務(shell、guards)
/// 落地時再評估是否改用 auto_route_generator 產碼。
class AppRouter extends RootStackRouter {
  AppRouter({required AuthRepository authRepository})
      : _authRepository = authRepository;

  final AuthRepository _authRepository;

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
      ];
}
