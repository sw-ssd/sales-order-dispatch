import type { CreateQueryResult } from "@tanstack/solid-query";
import { Match, Switch, type JSX } from "solid-js";
import { Spinner } from "@ui/spinner";
import { describeError } from "../lib/errors";

/** 頁面外框：console 的每個頁面都是「標題 + 說明 + 內容」，統一在這裡給。 */
export function PageShell(props: {
  title: string;
  description?: string;
  children: JSX.Element;
}): JSX.Element {
  return (
    <section class="space-y-4">
      <header>
        <h1 class="text-2xl font-bold">{props.title}</h1>
        <p class="mt-1 text-sm text-muted-foreground">{props.description}</p>
      </header>
      {props.children}
    </section>
  );
}

/**
 * 查詢三態（載入／錯誤／成功）的共用呈現；成功時把資料交給呼叫端的 render。
 * 錯誤一律經 describeError（碼表投影），不硬編訊息字串。
 */
export function queryBoundary<T>(
  query: CreateQueryResult<T>,
  render: (data: T) => JSX.Element,
): JSX.Element {
  return (
    <Switch>
      <Match when={query.isPending}>
        <div class="flex items-center gap-2 text-sm text-muted-foreground">
          <Spinner label="載入中" />
          <span>載入中…</span>
        </div>
      </Match>
      <Match when={query.isError}>
        <p role="alert" class="text-sm text-destructive">
          {describeError(query.error)}
        </p>
      </Match>
      <Match when={query.data !== undefined}>
        {render(query.data as T)}
      </Match>
    </Switch>
  );
}

/** 空清單的統一說法（避免每頁各寫一種空狀態文案）。 */
export function EmptyState(props: { children: JSX.Element }): JSX.Element {
  return <p class="text-sm text-muted-foreground">{props.children}</p>;
}
