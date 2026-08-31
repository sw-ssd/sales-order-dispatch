import { redirect } from "@solidjs/router";
import type { QueryClient } from "@tanstack/solid-query";
import { queryClient as defaultClient } from "~/lib/query-client";
import { abilityQueryOptions, createAppAbility } from "./service";

// makeRequireAbility 供測試注入 QueryClient;requireAbility 為正式綁定。
export function makeRequireAbility(qc: QueryClient) {
  return (action: string, subjectType: string) => async () => {
    const rules = await qc.ensureQueryData(abilityQueryOptions);
    if (!createAppAbility(rules).can(action, subjectType)) {
      throw redirect("/403");
    }
  };
}

// 用法(路由定義,Wave 3 對齊): { path: "/orders", component: OrdersPage, load: requireAbility("read", "sales_order") }
// 接線註記:frontend/src/router/index.tsx 屬 Wave 3 範圍,本檔僅提供守衛 hook;
// 權限異動後以 queryClient.invalidateQueries({ queryKey: ["ability"] }) 主動重載。
export const requireAbility = makeRequireAbility(defaultClient);
