import { queryOptions } from "@tanstack/solid-query";
import { createClient } from "@connectrpc/connect";
import { AbilityService } from "~/lib/proto/salesorder/v1/ability_pb";
import { transport } from "~/lib/transport";
import { createPermissions } from "./permissions";

export type { Permission };
export type AbilitySet = ReadonlySet<string>;

const client = createClient(AbilityService, transport);

// 權限查詢選項:queryKey ["ability"] 供 invalidateQueries 精準失效;
// staleTime 60 秒 = 規格 TTL。queryFn 回傳權限集合(取代 CASL ability)。
export const abilityQueryOptions = queryOptions({
  queryKey: ["ability"],
  staleTime: 60_000,
  queryFn: loadPermissions,
});

// loadPermissions 呼叫後端 GetAbility(OpenFGA proxy),將規則轉為 (resource:action) 集合。
// 後端 AbilityRule 的 subject 欄位即資源 resource。空規則 fail-closed(空集合)。
export async function loadPermissions(): Promise<AbilitySet> {
  const rules = (await client.getAbility({})).rules;
  // 後端 AbilityRule.subject 欄位即資源 resource。
  return createPermissions(rules.map((r) => ({ action: r.action, resource: r.subject })));
}

export { createPermissions, hasPermission, permKey } from "./permissions";
