import { createRouter, defineRoutes, type RouteSectionProps } from "@solidjs/router";
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
      <p class="mt-2 text-gray-600">首頁佔位(Wave 1 骨架)</p>
    </main>
  );
}

// @solidjs/router 2:無 <Router>/<Route> 元件,路由樹為不可變設定物件;
// 舊 Route load={requireAbility(...)} → 路由定義 preload(守衛語意見 guards.ts)。
const routes = defineRoutes([
  { path: "/", component: HomePage },
  { path: "/login", component: LoginPage },
  { path: "/403", component: ForbiddenPage },
  { path: "/users/companies", component: CompaniesPage, preload: requireAbility("read", "company") },
  { path: "/users/departments", component: DepartmentsPage, preload: requireAbility("read", "department") },
  { path: "/users/roles", component: RolesPage, preload: requireAbility("read", "role") },
]);

// RouterInstance 本身即 provider 元件;render-prop 收到匹配內容於 props.children。
const Router = createRouter({ routes });

export default function AppRouter() {
  return <Router>{(props: RouteSectionProps) => <App>{props.children}</App>}</Router>;
}
