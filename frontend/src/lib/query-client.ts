import { QueryClient } from "@tanstack/solid-query";

// 全站共用 QueryClient(T13;WEB-INF-04)。權限異動後主動重載慣例:
//   queryClient.invalidateQueries({ queryKey: ["ability"] })
// (權限設置頁儲存成功後呼叫,2.7/2.11 前端計畫引用)。
export const queryClient = new QueryClient();
