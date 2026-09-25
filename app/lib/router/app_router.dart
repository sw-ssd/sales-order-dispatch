import 'package:auto_route/auto_route.dart';

import '../core/api.dart';
import '../features/accounts/accounts_page.dart';
import '../features/auth/auth_repository.dart';
import '../features/auth/pages/identity_select_page.dart';
import '../features/auth/pages/login_page.dart';
import '../features/auth/pages/qr_login_page.dart';
import '../features/shell/home_shell.dart';

/// App 根路由。
///
/// auto_route 11 支援手寫路由表(PageInfo.builder),此處直接在建構式注入
/// [AuthRepository] 與 [Api] 並以閉包傳給各頁;待 D29 後續任務(shell、guards)
/// 落地時再評估是否改用 auto_route_generator 產碼。
///
/// **深層連結**(規格 §4.2)。兩條路徑由後端組出(`config.Auth.FrontendURL` 為 base):
/// - `/customer_account_qrcode/{token}`:QR 登入。token 一次性,兌換後才進登入。
/// - `/customer_account_manage`:帳號管理入口。**連結本身不含任何憑證**(規格明訂),
///   點進來只是一般店家登入頁 —— 登入成功後才依身分決定落在帳號管理或主殼。
///
/// 為什麼這兩條**不需要** deep link transformer:auto_route 的 parser 只取 `uri.path`
/// (見 `DefaultRouteParser._normalize`),Universal Link / App Link 帶進來的 host
/// (`https://<domain>/...`)自然被丟棄,path 即為上述路由。若日後改用帶前綴的
/// filter-intent(如 `/app/customer_account_qrcode/...`),才需要 `deepLinkTransformer`
/// 去掉該前綴。
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
            builder: (context, _) => LoginPage(
              authRepository: _authRepository,
              api: _api,
            ),
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
        AutoRoute(
          path: '/account',
          page: PageInfo.builder(
            'AccountsRoute',
            builder: (context, _) => AccountsPage(
              api: _api,
              auth: _authRepository,
            ),
          ),
        ),
        // QR 登入深層連結。token 走 path 參數(規格格式),不經 query —— 連結由
        // 後端 qrDeepLink() 以「base + /customer_account_qrcode/ + token」組成。
        AutoRoute(
          path: '/customer_account_qrcode/:token',
          page: PageInfo.builder(
            'QRLoginRoute',
            builder: (context, routeData) => QRLoginPage(
              authRepository: _authRepository,
              token: routeData.params.getString('token'),
            ),
          ),
        ),
        // 帳號管理深層連結:只是登入入口(連結不含憑證),登入後由 LoginPage 依身分導向。
        AutoRoute(
          path: '/customer_account_manage',
          page: PageInfo.builder(
            'AccountManageLoginRoute',
            builder: (context, _) => LoginPage(
              authRepository: _authRepository,
              api: _api,
              title: '帳號管理登入',
            ),
          ),
        ),
      ];
}
