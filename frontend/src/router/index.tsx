import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from "@tanstack/solid-router";
import App from "~/App";
import LoginPage from "~/features/auth/pages/LoginPage";
import ForbiddenPage from "~/features/auth/pages/ForbiddenPage";
import CompaniesPage from "~/features/users/pages/CompaniesPage";
import DepartmentsPage from "~/features/users/pages/DepartmentsPage";
import RolesPage from "~/features/users/pages/RolesPage";
import { requireAbility } from "~/lib/ability/guards";

function HomePage() {
  return (
    <main class="p-8">
      <h1 class="text-2xl font-bold">多公司訂出貨系統</h1>
      <p class="mt-2 text-gray-600">首頁佔位（Wave 1 骨架）</p>
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

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  forbiddenRoute,
  companiesRoute,
  departmentsRoute,
  rolesRoute,
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
