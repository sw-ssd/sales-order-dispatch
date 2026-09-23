import { useRouterState } from "@tanstack/solid-router";
import { createQuery } from "@tanstack/solid-query";
import { createMemo, type ParentProps } from "solid-js";
import AppShell from "~/components/layout/AppShell";
import { abilityQueryOptions } from "~/lib/ability/service";
import { AbilityProvider } from "~/lib/ability/context";

/** 不套 shell 的路徑：登入頁與 403 各自是完整頁面。 */
const CHROMELESS_PATHS: readonly string[] = ["/login", "/403"];

/**
 * 權限集合的**唯一來源**（`Can` / `useAbility` 的 consumer 都讀這裡）。
 *
 * 掛在 `App` 而非 `main.tsx`：這是已進入 router 的最外層元件，於是同時涵蓋 shell 與
 * chromeless 兩條路徑，且與路由守衛共用 `QueryClientProvider` 的 `["ability"]` 快取
 * ——守衛的 `ensureQueryData` 會先預熱，頁面直接命中，不重複打 GetAbility。
 *
 * 未載入完成或載入失敗時傳**空集合**：`Can` 據此不渲染受控內容（fail-closed），
 * 而載入失敗不該讓整頁崩掉 —— 前端不構成授權，實際可用性一律由後端擋。
 */
const EMPTY_ABILITY = new Set<string>();

export default function App(props: ParentProps) {
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const ability = createQuery(() => abilityQueryOptions);
  const perms = createMemo(() => ability.data ?? EMPTY_ABILITY);

  return (
    <AbilityProvider ability={perms}>
      <AppShell chromeless={CHROMELESS_PATHS.includes(pathname())}>{props.children}</AppShell>
    </AbilityProvider>
  );
}
