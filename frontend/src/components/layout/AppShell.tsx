import { useRouterState } from "@tanstack/solid-router";
import { createEffect, createSignal, Show, type ParentProps } from "solid-js";
import { cn } from "~/lib/cn";
import Sidebar from "./Sidebar";
import Topbar from "./Topbar";

export type AppShellProps = ParentProps<{
  /** true = 只渲染頁面內容（登入頁、403 這類自己就是完整版面的頁面）。 */
  chromeless: boolean;
}>;

/**
 * app shell：Tailkit Light Sidebar 版型（固定側邊欄 + 64px 頂欄 + 內容區）。
 * 行動寬度下側邊欄收合成全寬抽屜，由頂欄的收合鈕開、抽屜的關閉鈕關。
 */
export default function AppShell(props: AppShellProps) {
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const [sidebarOpen, setSidebarOpen] = createSignal(false);

  // 換頁後收起行動版抽屜，避免它蓋住剛進去的頁面。
  createEffect(() => {
    pathname();
    setSidebarOpen(false);
  });

  return (
    <div
      class={cn(
        "mx-auto flex min-h-dvh w-full min-w-80 flex-col",
        !props.chromeless && "bg-background text-foreground lg:pl-64",
      )}
    >
      <Show when={!props.chromeless}>
        <Sidebar pathname={pathname()} open={sidebarOpen()} onOpenChange={setSidebarOpen} />
        <Topbar pathname={pathname()} onMenuClick={() => setSidebarOpen(true)} />
      </Show>
      <div class={cn("grow", !props.chromeless && "p-4 lg:p-6")}>{props.children}</div>
    </div>
  );
}
