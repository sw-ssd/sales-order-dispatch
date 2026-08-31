//
//  Generated code. Do not modify.
//  source: salesorder/v1/ability.proto
//

import "package:connectrpc/connect.dart" as connect;
import "ability.pb.dart" as salesorderv1ability;

/// AbilityService:能力規則下發。
abstract final class AbilityService {
  /// Fully-qualified name of the AbilityService service.
  static const name = 'salesorder.v1.AbilityService';

  /// GetAbility:以 ctx 身分查詢規則表,輸出 CASL JSON 規則。
  static const getAbility = connect.Spec(
    '/$name/GetAbility',
    connect.StreamType.unary,
    salesorderv1ability.GetAbilityRequest.new,
    salesorderv1ability.GetAbilityResponse.new,
  );
}
