import { useNavigate, type Navigator, type RoutePreloadFunc } from "@solidjs/router";
import type { QueryClient } from "@tanstack/solid-query";
import { queryClient as defaultClient } from "~/lib/query-client";
import { abilityQueryOptions, createAppAbility } from "./service";

// makeRequireAbility 供測試注入 QueryClient 與 navigate;requireAbility 為正式綁定。
//
// @solidjs/router 2 遷移註記:舊版 Route load 拋 redirect() 由路由器攔截;
// 2.0 的 preload 不被 await(純資料預熱),拋出的 Response 無人處理,
// 故改於 preload 內主動 navigate("/403", { replace: true })。
// preload 於路由 context 建立期間同步觸發,useNavigate 必須在 async 邊界前取得。
// intent === "preload"(連結 hover/focus 預熱)不導向,對齊舊版 load 於預載時吞掉 redirect 的行為。
export function makeRequireAbility(
  qc: QueryClient,
  getNavigate: () => Navigator = useNavigate,
) {
  return (action: string, subjectType: string): RoutePreloadFunc<Promise<void>> =>
    async (args) => {
      if (args.intent === "preload") return;
      const navigate = getNavigate();
      const rules = await qc.ensureQueryData(abilityQueryOptions);
      if (!createAppAbility(rules).can(action, subjectType)) {
        navigate("/403", { replace: true });
      }
    };
}

// 用法(路由定義): { path: "/orders", component: OrdersPage, preload: requireAbility("read", "sales_order") }
// 權限異動後以 queryClient.invalidateQueries({ queryKey: ["ability"] }) 主動重載。
export const requireAbility = makeRequireAbility(defaultClient);
