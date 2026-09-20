import { Link, useNavigate } from "@tanstack/solid-router";
import { For, type ParentProps } from "solid-js";
import { Button } from "@ui/button";
import { logout } from "./lib/session";

/** 導覽：console 的功能只有這五頁（登入頁不在此列）。 */
const NAV = [
  { to: "/tenants", label: "租戶" },
  { to: "/plans", label: "方案" },
  { to: "/entitlements", label: "權益" },
  { to: "/receivables", label: "待收款" },
  { to: "/audit", label: "稽核" },
] as const;

/**
 * console 的外框：標題、導覽、登出。
 *
 * 登出是**純前端**動作（v1 無登出端點、cookie 是 HttpOnly）：只清本地狀態後回到登入頁，
 * cookie 仍有效至效期結束。要即時撤銷 operator 存取，得在後端停用該 operator 白名單列。
 */
export default function App(props: ParentProps) {
  const navigate = useNavigate();
  return (
    <div class="min-h-screen">
      <header class="border-b border-border bg-card">
        <div class="mx-auto flex max-w-6xl items-center gap-4 px-4 py-3">
          <span class="font-semibold">平台營運主控台</span>
          <nav class="flex items-center gap-1" aria-label="主導覽">
            <For each={NAV}>
              {(item) => (
                <Link
                  to={item.to}
                  class="rounded-lg px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                >
                  {item.label}
                </Link>
              )}
            </For>
          </nav>
          <Button
            variant="outline"
            size="sm"
            class="ml-auto"
            onClick={() => {
              logout();
              void navigate({ to: "/login" });
            }}
          >
            登出
          </Button>
        </div>
      </header>
      <main class="mx-auto max-w-6xl px-4 py-6">{props.children}</main>
    </div>
  );
}
