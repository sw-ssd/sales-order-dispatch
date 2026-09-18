//
//  Generated code. Do not modify.
//  source: salesorder/v1/user.proto
//

import "package:connectrpc/connect.dart" as connect;
import "user.pb.dart" as salesorderv1user;

/// UserService:使用者管理。
abstract final class UserService {
  /// Fully-qualified name of the UserService service.
  static const name = 'salesorder.v1.UserService';

  /// ListUsers:分頁列出使用者,可依 company_id / department_id / role / status 篩選。
  static const listUsers = connect.Spec(
    '/$name/ListUsers',
    connect.StreamType.unary,
    salesorderv1user.ListUsersRequest.new,
    salesorderv1user.ListUsersResponse.new,
  );

  /// GetUser:取得單一使用者(不含 password_hash)。
  static const getUser = connect.Spec(
    '/$name/GetUser',
    connect.StreamType.unary,
    salesorderv1user.GetUserRequest.new,
    salesorderv1user.GetUserResponse.new,
  );

  /// CreateUser:建立員工帳號(super 直接建立;password_hash 依 OAuth 流程填補)。
  static const createUser = connect.Spec(
    '/$name/CreateUser',
    connect.StreamType.unary,
    salesorderv1user.CreateUserRequest.new,
    salesorderv1user.CreateUserResponse.new,
  );

  /// UpdateUser:更新使用者(name / department_id / phone / employee_no / status)。
  static const updateUser = connect.Spec(
    '/$name/UpdateUser',
    connect.StreamType.unary,
    salesorderv1user.UpdateUserRequest.new,
    salesorderv1user.UpdateUserResponse.new,
  );

  /// AssignRole:角色指派(含 guest 審核:status pending → active 並指派部門與角色)。
  static const assignRole = connect.Spec(
    '/$name/AssignRole',
    connect.StreamType.unary,
    salesorderv1user.AssignRoleRequest.new,
    salesorderv1user.AssignRoleResponse.new,
  );

  /// Deactivate:停用帳號(status → inactive, token_version+1, 刪 session)。
  static const deactivate = connect.Spec(
    '/$name/Deactivate',
    connect.StreamType.unary,
    salesorderv1user.DeactivateRequest.new,
    salesorderv1user.DeactivateResponse.new,
  );

  /// ForceLogout:強制登出(token_version+1, 使該使用者在途憑證失效)。
  static const forceLogout = connect.Spec(
    '/$name/ForceLogout',
    connect.StreamType.unary,
    salesorderv1user.ForceLogoutRequest.new,
    salesorderv1user.ForceLogoutResponse.new,
  );
}
