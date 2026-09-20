import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  redirect,
  RouterProvider,
  useRouterState,
  type RouterHistory,
} from "@tanstack/solid-router";
import { Show } from "solid-js";
import App from "./App";
import { requireOperator } from "./lib/guard";
import AuditPage from "./pages/AuditPage";
import EntitlementsPage from "./pages/EntitlementsPage";
import LoginPage from "./pages/LoginPage";
import PlansPage from "./pages/PlansPage";
import ReceivablesPage from "./pages/ReceivablesPage";
import TenantDetailPage from "./pages/TenantDetailPage";
import TenantsPage from "./pages/TenantsPage";

/** 唯一不需要 operator 的路徑；也是唯一不套外框的路徑。 */
const LOGIN_ROUTE = "/login";

/**
 * console 自己的路由樹（**不共用**租戶 SPA 的 router 或守衛）。
 *
 * root 只負責「套不套外框」：`/login` 不套（否則會出現按了又被守衛彈回的登出鈕與整排導覽，
 * 且 `<main>` 會嵌套；與租戶 SPA 的 chromeless 慣例同型），其餘頁面套外框（導覽＋登出）。
 * 六頁平台功能全部掛 `beforeLoad: requireOperator()`；只有 `/login` 不需要 operator。
 *
 * 守衛是 UX，不是授權：後端 operatorauth.Interceptor 才是決策者（見 lib/guard.ts）。
 */
const rootRoute = createRootRoute({
  component: () => {
    const pathname = useRouterState({ select: (state) => state.location.pathname });
    return (
      <Show when={pathname() !== LOGIN_ROUTE} fallback={<Outlet />}>
        <App>
          <Outlet />
        </App>
      </Show>
    );
  },
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: LOGIN_ROUTE,
  component: LoginPage,
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  beforeLoad: () => {
    throw redirect({ to: "/tenants" });
  },
});

const tenantsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/tenants",
  component: TenantsPage,
  beforeLoad: requireOperator(),
});

const tenantDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/tenants/$tenantId",
  component: TenantDetailPage,
  beforeLoad: requireOperator(),
});

const plansRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/plans",
  component: PlansPage,
  beforeLoad: requireOperator(),
});

const entitlementsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/entitlements",
  component: EntitlementsPage,
  beforeLoad: requireOperator(),
});

const receivablesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/receivables",
  component: ReceivablesPage,
  beforeLoad: requireOperator(),
});

const auditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/audit",
  component: AuditPage,
  beforeLoad: requireOperator(),
});

export const routeTree = rootRoute.addChildren([
  loginRoute,
  indexRoute,
  tenantsRoute,
  tenantDetailRoute,
  plansRoute,
  entitlementsRoute,
  receivablesRoute,
  auditRoute,
]);

/** history 可注入：測試以 createMemoryHistory 驅動，正式用瀏覽器 history。 */
export function createAppRouter(history?: RouterHistory) {
  return createRouter({ routeTree, history });
}

export const router = createAppRouter();

// 註冊 router 型別：Link／navigate 的 `to` 取得路由字面量型別檢查。
declare module "@tanstack/solid-router" {
  interface Register {
    router: typeof router;
  }
}

export function AppRouter() {
  return <RouterProvider router={router} />;
}
