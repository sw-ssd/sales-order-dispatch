//
//  Generated code. Do not modify.
//  source: salesorder/v1/role.proto
//

import "package:connectrpc/connect.dart" as connect;
import "role.pb.dart" as salesorderv1role;
import "role.connect.spec.dart" as specs;

/// RoleService:角色權限管理。
extension type RoleServiceClient (connect.Transport _transport) {
  /// ListRoles:分頁列出角色。
  Future<salesorderv1role.ListRolesResponse> listRoles(
    salesorderv1role.ListRolesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RoleService.listRoles,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetRolePermissions:取得角色功能權限。
  Future<salesorderv1role.GetRolePermissionsResponse> getRolePermissions(
    salesorderv1role.GetRolePermissionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RoleService.getRolePermissions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateRolePermissions:全量更新角色功能權限(取代式)。
  Future<salesorderv1role.UpdateRolePermissionsResponse> updateRolePermissions(
    salesorderv1role.UpdateRolePermissionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RoleService.updateRolePermissions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListConditionFields:取得資源的條件欄位白名單(條件建構器用)。
  Future<salesorderv1role.ListConditionFieldsResponse> listConditionFields(
    salesorderv1role.ListConditionFieldsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RoleService.listConditionFields,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
