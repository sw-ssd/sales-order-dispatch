import { QueryClientProvider } from "@tanstack/solid-query";
import { render } from "solid-js/web";
import AppRouter from "~/router";
import { queryClient } from "~/lib/query-client";
import "~/index.css";

const root = document.getElementById("root");
if (!root) throw new Error("找不到 #root 掛載點");

// 全站共用的 QueryClient（`lib/query-client.ts`，含共用 retry 謂詞）：
// 頁面的 `createQuery` 與路由守衛的 `ensureQueryData` 走同一個實例，
// 權限異動後的 invalidateQueries(["ability"]) 才會同時對兩邊生效。
//
// `AbilityProvider` 自 `App`（router 之內）而非此處：它需要與 router 守衛**同一個**
// QueryClient 讀 `["ability"]`（守衛走 `ensureQueryData` 預熱，頁面直接命中快取），
// 在 `App` 內即可同時滿足這一點並涵蓋 shell 與 chromeless 兩條路徑。
render(
  () => (
    <QueryClientProvider client={queryClient}>
      <AppRouter />
    </QueryClientProvider>
  ),
  root
);
