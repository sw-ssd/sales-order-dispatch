//
//  Generated code. Do not modify.
//  source: salesorder/v1/user.proto
//

import "package:connectrpc/connect.dart" as connect;
import "user.pb.dart" as salesorderv1user;
import "user.connect.spec.dart" as specs;

/// UserService:使用者管理。
extension type UserServiceClient (connect.Transport _transport) {
  /// ListUsers:分頁列出使用者,可依 company_id / department_id / role / status 篩選。
  Future<salesorderv1user.ListUsersResponse> listUsers(
    salesorderv1user.ListUsersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.listUsers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetUser:取得單一使用者(不含 password_hash)。
  Future<salesorderv1user.GetUserResponse> getUser(
    salesorderv1user.GetUserRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.getUser,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateUser:建立員工帳號(super 直接建立;password_hash 依 OAuth 流程填補)。
  Future<salesorderv1user.CreateUserResponse> createUser(
    salesorderv1user.CreateUserRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.createUser,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateUser:更新使用者(name / department_id / phone / employee_no / status)。
  Future<salesorderv1user.UpdateUserResponse> updateUser(
    salesorderv1user.UpdateUserRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.updateUser,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AssignRole:角色指派(含 guest 審核:status pending → active 並指派部門與角色)。
  Future<salesorderv1user.AssignRoleResponse> assignRole(
    salesorderv1user.AssignRoleRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.assignRole,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// Deactivate:停用帳號(status → inactive, token_version+1, 刪 session)。
  Future<salesorderv1user.DeactivateResponse> deactivate(
    salesorderv1user.DeactivateRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.deactivate,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ForceLogout:強制登出(token_version+1, 使該使用者在途憑證失效)。
  Future<salesorderv1user.ForceLogoutResponse> forceLogout(
    salesorderv1user.ForceLogoutRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.UserService.forceLogout,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
