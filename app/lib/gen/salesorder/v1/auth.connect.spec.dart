//
//  Generated code. Do not modify.
//  source: salesorder/v1/auth.proto
//

import "package:connectrpc/connect.dart" as connect;
import "auth.pb.dart" as salesorderv1auth;

/// AuthService:認證相關 RPC。
abstract final class AuthService {
  /// Fully-qualified name of the AuthService service.
  static const name = 'salesorder.v1.AuthService';

  /// Login:客戶帳號密碼登入。
  static const login = connect.Spec(
    '/$name/Login',
    connect.StreamType.unary,
    salesorderv1auth.LoginRequest.new,
    salesorderv1auth.LoginResponse.new,
  );

  /// Refresh:refresh token 旋轉換發。
  static const refresh = connect.Spec(
    '/$name/Refresh',
    connect.StreamType.unary,
    salesorderv1auth.RefreshRequest.new,
    salesorderv1auth.RefreshResponse.new,
  );

  /// Logout:撤銷 refresh token(冪等)。
  static const logout = connect.Spec(
    '/$name/Logout',
    connect.StreamType.unary,
    salesorderv1auth.LogoutRequest.new,
    salesorderv1auth.LogoutResponse.new,
  );

  /// RegisterComplete:guest 補完註冊資料。
  static const registerComplete = connect.Spec(
    '/$name/RegisterComplete',
    connect.StreamType.unary,
    salesorderv1auth.RegisterCompleteRequest.new,
    salesorderv1auth.RegisterCompleteResponse.new,
  );

  /// QRLogin:QR token 兌換,回公司/客戶與可選子帳號清單。
  static const qRLogin = connect.Spec(
    '/$name/QRLogin',
    connect.StreamType.unary,
    salesorderv1auth.QRLoginRequest.new,
    salesorderv1auth.QRLoginResponse.new,
  );
}
