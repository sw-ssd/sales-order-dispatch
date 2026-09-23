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
import CustomersPage from "~/features/customers/pages/CustomersPage";
import OrdersPage from "~/features/orders/pages/OrdersPage";
import ProductsPage from "~/features/products/pages/ProductsPage";
import DispatchPage from "~/features/dispatch/pages/DispatchPage";
import RoutesPage from "~/features/masters/pages/RoutesPage";
import WarehousesPage from "~/features/masters/pages/WarehousesPage";
import ProductCategoriesPage from "~/features/masters/pages/ProductCategoriesPage";
import ProcessingSpecsPage from "~/features/masters/pages/ProcessingSpecsPage";
import PrintPage from "~/features/printing/pages/PrintPage";
import ReturnsPage from "~/features/returns/pages/ReturnsPage";
import NotificationsPage from "~/features/notifications/pages/NotificationsPage";
import AuditPage from "~/features/audit/pages/AuditPage";
import { requireAbility } from "~/lib/ability/guards";
import DashboardPage from "~/features/dashboard/pages/DashboardPage";

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
  component: DashboardPage,
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

const customersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customers",
  component: CustomersPage,
  beforeLoad: requireAbility("read", "customer"),
});

const ordersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/orders",
  component: OrdersPage,
  beforeLoad: requireAbility("read", "sales_order"),
});

const productsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/products",
  component: ProductsPage,
  beforeLoad: requireAbility("read", "product"),
});

/**
 * 退貨管理：申請清單／明細審核／退貨證明（06，D25）。
 *
 * 守衛用 `read, "return_request"`：`return_request` 與本頁同批進 `rolePolicy`
 * （company_admin/dept_admin `*`、staff `read`+`write`；customer 刻意不給 —— 客戶自助走
 * App，Web 端維持 403 指引）。`ReturnService` 不在 `protectedRPC` 表內，實際授權落在
 * handler：`deptScope`（清單/明細/證明）、`canReview`（staff 收斂到該客戶主責業務）與 RLS。
 * **發起申請只有客戶子帳號能做**（`returnCustomerScope` 拒絕員工與主帳號），故本頁不提供建單。
 */
const returnsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/returns",
  component: ReturnsPage,
  beforeLoad: requireAbility("read", "return_request"),
});

const dispatchRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dispatch",
  component: DispatchPage,
  beforeLoad: requireAbility("read", "dispatch"),
});

/**
 * 通知中心：本人通知清單、未讀篩選與標記已讀（07，規格 §5.4）。
 *
 * 守衛用 `read, "notification"`：`notification` 與退貨資源同批進 `rolePolicy`
 * （company_admin/dept_admin `*`、staff `read`+`write`；customer 刻意不給 —— 客戶通知走
 * App，Web 端維持 403）。頁面資料是**本人**的（`NotificationService` 只回 `user_id` =
 * 自己的列，他人通知在 MarkRead 也視同 not_found），故這道守衛只決定「誰看得到入口」。
 */
const notificationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/notifications",
  component: NotificationsPage,
  beforeLoad: requireAbility("read", "notification"),
});

/**
 * 稽核日誌（D18/D27）。
 *
 * 守衛用 `read, "audit_log"`，受眾刻意只有 company_admin：與
 * `AuditService.ListAuditLogs` 的範圍推導一致（super/developer 萬用、company_admin
 * 強制自己公司、其餘角色直接 PermissionDenied）—— 守衛寬了會讓 dept_admin/staff
 * 看見入口卻每次查詢都 403，窄了則等於把後端的範圍判斷抄兩份。
 */
const auditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/audit",
  component: AuditPage,
  beforeLoad: requireAbility("read", "audit_log"),
});

/**
 * 車次主檔：派車看板的欄位來源。
 *
 * 守衛用 `read, dispatch` —— ability 只有 10 個資源（無 `route`/`master`），
 * 軍種介面屬派車域，故與看板同受眾；寫入授權落在 handler 的 `deptScope` 與 RLS。
 */
const routesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/masters/routes",
  component: RoutesPage,
  beforeLoad: requireAbility("read", "dispatch"),
});

/**
 * 倉別主檔（04 Task 3.4）。
 *
 * 守衛與車次同為 `read, dispatch`：ability 沒有 `warehouse`/`master` 資源，而倉別正是
 * 揀貨單（picking_list）與派車域的欄位來源 —— 能開看板的人就能看它的來源主檔。
 * 寫入授權落在 handler 的 `requireAuth`＋`deptScope` 與 RLS。
 */
const warehousesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/masters/warehouses",
  component: WarehousesPage,
  beforeLoad: requireAbility("read", "dispatch"),
});

/**
 * 商品分類主檔（04 Task 3.4）。
 *
 * 守衛用 `read, product`（不是 dispatch）：分類屬**商品域**，而 staff 對 `product` 是
 * `*`、對 `dispatch` 只有 `read` —— 掛錯資源會讓有商品權限的 staff 打不開這頁。
 */
const categoriesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/masters/categories",
  component: ProductCategoriesPage,
  beforeLoad: requireAbility("read", "product"),
});

/**
 * 分切規格主檔（04 Task 3.4；規格欄位供 05 訂單明細與 09 加工單）。
 *
 * 守衛同分類：規格是**商品域**資料（`read, product`），理由見 `categoriesRoute`。
 */
const processingSpecsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/masters/processing-specs",
  component: ProcessingSpecsPage,
  beforeLoad: requireAbility("read", "product"),
});

/**
 * 單據列印。
 *
 * 守衛用 `read, "print"`：`print` 是 ability 內建資源（`rolePolicy` 給
 * company_admin/dept_admin/staff `{"*"}`），只有有列印權限的角色看得到這頁；
 * 後端 `PrintService` 不在 `protectedRPC` 表內，寫入授權落在 handler 的
 * `requireAuth`+`deptScope` 與 RLS（與 masters 同模式）。
 */
const printRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/printing",
  component: PrintPage,
  beforeLoad: requireAbility("read", "print"),
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
  customersRoute,
  ordersRoute,
  productsRoute,
  returnsRoute,
  notificationsRoute,
  auditRoute,
  dispatchRoute,
  routesRoute,
  warehousesRoute,
  categoriesRoute,
  processingSpecsRoute,
  printRoute,
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
