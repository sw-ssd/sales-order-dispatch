//
//  Generated code. Do not modify.
//  source: salesorder/v1/ability.proto
//

import "package:connectrpc/connect.dart" as connect;
import "ability.pb.dart" as salesorderv1ability;
import "ability.connect.spec.dart" as specs;

/// AbilityService:能力規則下發。
extension type AbilityServiceClient (connect.Transport _transport) {
  /// GetAbility:以 ctx 身分查詢規則表,輸出 CASL JSON 規則。
  Future<salesorderv1ability.GetAbilityResponse> getAbility(
    salesorderv1ability.GetAbilityRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AbilityService.getAbility,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
