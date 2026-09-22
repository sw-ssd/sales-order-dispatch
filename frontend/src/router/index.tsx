import {
  createRootRoute,
  createRoute,
  createRouter,
  lazyRouteComponent,
  Outlet,
  RouterProvider,
} from "@tanstack/solid-router";
import App from "~/App";
import AccountPage from "~/features/account/AccountPage";
import LoginPage from "~/features/auth/pages/LoginPage";
import ForbiddenPage from "~/features/auth/pages/ForbiddenPage";
import CompaniesPage from "~/features/users/pages/CompaniesPage";
import DepartmentsPage from "~/features/users/pages/DepartmentsPage";
import RolesPage from "~/features/users/pages/RolesPage";
import UsersPage from "~/features/users/pages/UsersPage";
import { requireAbility } from "~/lib/ability/guards";

function HomePage() {
  // 內距由 shell 統一提供（`AppShell` 是唯一內距所有者），頁面層不再自帶。
  return (
    <main>
      <h1 class="text-2xl font-bold">多公司訂出貨系統</h1>
      <p class="mt-2 text-muted-foreground">首頁佔位（Wave 1 骨架）</p>
    </main>
  );
}

// TanStack Router 程式化路由樹;root route component 承載 App 佈局,Outlet 渲染子路由。
const rootRoute = createRootRoute({
  component: () => (
    <App>
      <Outlet />
    </App>
  ),
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: HomePage,
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
});

const forbiddenRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/403",
  component: ForbiddenPage,
});

const companiesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/users/companies",
  component: CompaniesPage,
  beforeLoad: requireAbility("read", "company"),
});

const departmentsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/users/departments",
  component: DepartmentsPage,
  beforeLoad: requireAbility("read", "department"),
});

const rolesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/users/roles",
  component: RolesPage,
  beforeLoad: requireAbility("read", "role"),
});

const usersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/users/users",
  component: UsersPage,
  beforeLoad: requireAbility("read", "user"),
});

/**
 * 帳號／方案頁：租戶後台唯讀的權益卡片。
 *
 * 刻意的**沒有**能力守衛：`platform.*` 能力不得出現在租戶端（S11），而這張卡片誰看得到
 * 由後端決定（`GetTenantEntitlements` 只回自己的公司、`requireAuth` 擋未登入），前端不代判。
 */
const accountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/account",
  component: AccountPage,
});

// 元件庫展示頁（`/ui`）：只在開發環境註冊。正式 build 時 `import.meta.env.DEV`
// 被折成 false，整個分支——含動態 import 的 demo chunk——會被 tree-shake 掉。
const devRoutes = import.meta.env.DEV
  ? [
      createRoute({
        getParentRoute: () => rootRoute,
        path: "/ui",
        component: lazyRouteComponent(() => import("~/components/ui/demo/UiDemoPage")),
      }),
    ]
  : [];

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  forbiddenRoute,
  companiesRoute,
  departmentsRoute,
  rolesRoute,
  usersRoute,
  accountRoute,
  ...devRoutes,
]);

export const router = createRouter({ routeTree });

// 註冊 router 型別:Link/navigate 的 `to` 獲得路由字面量型別檢查。
declare module "@tanstack/solid-router" {
  interface Register {
    router: typeof router;
  }
}

export default function AppRouter() {
  return <RouterProvider router={router} />;
}
