//
//  Generated code. Do not modify.
//  source: salesorder/v1/auth.proto
//

import "package:connectrpc/connect.dart" as connect;
import "auth.pb.dart" as salesorderv1auth;
import "auth.connect.spec.dart" as specs;

/// AuthService:認證相關 RPC。
extension type AuthServiceClient (connect.Transport _transport) {
  /// Login:客戶帳號密碼登入。
  Future<salesorderv1auth.LoginResponse> login(
    salesorderv1auth.LoginRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.login,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// Refresh:refresh token 旋轉換發。
  Future<salesorderv1auth.RefreshResponse> refresh(
    salesorderv1auth.RefreshRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.refresh,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// Logout:撤銷 refresh token(冪等)。
  Future<salesorderv1auth.LogoutResponse> logout(
    salesorderv1auth.LogoutRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.logout,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RegisterComplete:guest 補完註冊資料。
  Future<salesorderv1auth.RegisterCompleteResponse> registerComplete(
    salesorderv1auth.RegisterCompleteRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.registerComplete,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// QRLogin:QR token 兌換,回公司/客戶與可選子帳號清單。
  Future<salesorderv1auth.QRLoginResponse> qRLogin(
    salesorderv1auth.QRLoginRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.qRLogin,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCustomerQRCode:為本部門客戶產生登入 QR(dept_admin/staff 限本部門;回深層連結,token 另存)。
  Future<salesorderv1auth.GetCustomerQRCodeResponse> getCustomerQRCode(
    salesorderv1auth.GetCustomerQRCodeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.getCustomerQRCode,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ChangePassword:登入態修改密碼(1.5.2;must_change_password 時唯一可用)。
  Future<salesorderv1auth.ChangePasswordResponse> changePassword(
    salesorderv1auth.ChangePasswordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.changePassword,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ResetCustomerPassword:密碼重置,重新發臨時密碼(1.5.4;dept_admin 以上)。
  Future<salesorderv1auth.ResetCustomerPasswordResponse> resetCustomerPassword(
    salesorderv1auth.ResetCustomerPasswordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.resetCustomerPassword,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
