import { useRouterState } from "@tanstack/solid-router";
import type { ParentProps } from "solid-js";
import AppShell from "~/components/layout/AppShell";

/** 不套 shell 的路徑：登入頁與 403 各自是完整頁面。 */
const CHROMELESS_PATHS: readonly string[] = ["/login", "/403"];

export default function App(props: ParentProps) {
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  return (
    <AppShell chromeless={CHROMELESS_PATHS.includes(pathname())}>{props.children}</AppShell>
  );
}
