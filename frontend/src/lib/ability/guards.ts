import { redirect } from "@tanstack/solid-router";
import type { QueryClient } from "@tanstack/solid-query";
import { queryClient as defaultClient } from "~/lib/query-client";
import { abilityQueryOptions } from "./service";
import { hasPermission } from "./permissions";

// makeRequireAbility 供測試注入 QueryClient;requireAbility 為正式綁定。
//
// TanStack Router 遷移註記:舊 @solidjs/router Route load 拋 redirect() 由路由器攔截;
// TanStack 以 beforeLoad 提供同等守衛語意——拋出 redirect({ to }) 會真正阻擋導航。
// beforeLoad args.preload 為 hover/focus 連結預熱,此時不導向。
export function makeRequireAbility(qc: QueryClient) {
  return (action: string, subjectType: string) =>
    async ({ preload }: { preload: boolean }) => {
      if (preload) return;
      const perms = await qc.ensureQueryData(abilityQueryOptions);
      if (!hasPermission(perms, subjectType, action)) {
        throw redirect({ to: "/403" });
      }
    };
}

// 用法(路由定義): createRoute({ path: "/orders", component: OrdersPage, beforeLoad: requireAbility("read", "sales_order") })
// 權限異動後以 queryClient.invalidateQueries({ queryKey: ["ability"] }) 主動重載。
export const requireAbility = makeRequireAbility(defaultClient);
