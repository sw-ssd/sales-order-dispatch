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
render(
  () => (
    <QueryClientProvider client={queryClient}>
      <AppRouter />
    </QueryClientProvider>
  ),
  root
);
