//
//  Generated code. Do not modify.
//  source: salesorder/v1/role.proto
//

import "package:connectrpc/connect.dart" as connect;
import "role.pb.dart" as salesorderv1role;

/// RoleService:角色權限管理。
abstract final class RoleService {
  /// Fully-qualified name of the RoleService service.
  static const name = 'salesorder.v1.RoleService';

  /// ListRoles:分頁列出角色。
  static const listRoles = connect.Spec(
    '/$name/ListRoles',
    connect.StreamType.unary,
    salesorderv1role.ListRolesRequest.new,
    salesorderv1role.ListRolesResponse.new,
  );

  /// GetRolePermissions:取得角色功能權限。
  static const getRolePermissions = connect.Spec(
    '/$name/GetRolePermissions',
    connect.StreamType.unary,
    salesorderv1role.GetRolePermissionsRequest.new,
    salesorderv1role.GetRolePermissionsResponse.new,
  );

  /// UpdateRolePermissions:全量更新角色功能權限(取代式)。
  static const updateRolePermissions = connect.Spec(
    '/$name/UpdateRolePermissions',
    connect.StreamType.unary,
    salesorderv1role.UpdateRolePermissionsRequest.new,
    salesorderv1role.UpdateRolePermissionsResponse.new,
  );

  /// ListConditionFields:取得資源的條件欄位白名單(條件建構器用)。
  static const listConditionFields = connect.Spec(
    '/$name/ListConditionFields',
    connect.StreamType.unary,
    salesorderv1role.ListConditionFieldsRequest.new,
    salesorderv1role.ListConditionFieldsResponse.new,
  );
}
