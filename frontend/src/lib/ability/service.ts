import { createMongoAbility, type MongoAbility, type RawRuleOf } from "@casl/ability";
import { queryOptions } from "@tanstack/solid-query";
import { createClient } from "@connectrpc/connect";
import { AbilityService } from "~/lib/proto/salesorder/v1/ability_pb";
import { transport } from "~/lib/transport";

export type AppAbility = MongoAbility;

const client = createClient(AbilityService, transport);

// ability 查詢選項:queryKey ["ability"] 供 invalidateQueries 精準失效;
// staleTime 60 秒 = 規格 TTL,路由切換不重取。
export const abilityQueryOptions = queryOptions({
  queryKey: ["ability"],
  staleTime: 60_000,
  queryFn: async () => (await client.getAbility({})).rules,
});

// 由 proto GetAbility 規則建立 CASL ability;空規則 fail-closed(全部 deny)。
export function createAppAbility(rules: ReadonlyArray<{
  action: string;
  subject: string;
  conditions?: Record<string, unknown>;
  inverted?: boolean;
}>): AppAbility {
  const raw: RawRuleOf<AppAbility>[] = rules.map((r) => ({
    action: r.action,
    subject: r.subject,
    ...(r.conditions ? { conditions: r.conditions } : {}),
    ...(r.inverted ? { inverted: true as const } : {}),
  }));
  return createMongoAbility(raw);
}
