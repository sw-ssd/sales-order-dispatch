import { redirect } from "@tanstack/solid-router";
import type { QueryClient } from "@tanstack/solid-query";
import { queryClient as defaultClient } from "~/lib/query-client";
import { abilityQueryOptions, createAppAbility } from "./service";

// makeRequireAbility 供測試注入 QueryClient;requireAbility 為正式綁定。
//
// TanStack Router 遷移註記:舊 @solidjs/router Route load 拋 redirect() 由路由器攔截;
// TanStack 以 beforeLoad 提供同等守衛語意——拋出 redirect({ to }) 會真正阻擋導航
// (修正 Solid 2 遷移版 preload 僅資料預熱、不阻擋導航的殘留風險)。
// beforeLoad args.preload 為 hover/focus 連結預熱(舊版 intent === "preload"),
// 此時不導向,對齊舊 load 於預載時吞掉 redirect 的行為。
export function makeRequireAbility(qc: QueryClient) {
  return (action: string, subjectType: string) =>
    async ({ preload }: { preload: boolean }) => {
      if (preload) return;
      const rules = await qc.ensureQueryData(abilityQueryOptions);
      if (!createAppAbility(rules).can(action, subjectType)) {
        throw redirect({ to: "/403" });
      }
    };
}

// 用法(路由定義): createRoute({ path: "/orders", component: OrdersPage, beforeLoad: requireAbility("read", "sales_order") })
// 權限異動後以 queryClient.invalidateQueries({ queryKey: ["ability"] }) 主動重載。
export const requireAbility = makeRequireAbility(defaultClient);
