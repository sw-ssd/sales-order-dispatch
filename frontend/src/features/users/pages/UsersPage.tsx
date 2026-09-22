import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
import { createInfiniteQuery, createQuery, useQueryClient } from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  rowSortingFeature,
  tableFeatures,
  type PaginationState,
} from "@tanstack/solid-table";
import { batch, createEffect, createSignal, For, Show, type JSX } from "solid-js";
import type { Company } from "~/lib/proto/salesorder/v1/company_pb";
import type { Department } from "~/lib/proto/salesorder/v1/company_pb";
import type { User } from "~/lib/proto/salesorder/v1/user_pb";
import {
  Badge,
  Button,
  buttonVariants,
  Card,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Field,
  FieldError,
  FieldLabel,
  Input,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui";
import { appFormOptions, fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../components/ListPagination";
import {
  PAGE_SIZE,
  companyDropdownQueryOptions,
  departmentClient,
  userClient,
  usersQueryOptions,
} from "../queries";
import { userSchema } from "../schemas";

/**
 * 使用者表格的 table 功能集：分頁（後端無排序白名單 → 不開 sorting，見下方）。
 */
const USER_TABLE_FEATURES = tableFeatures({ rowPaginationFeature, rowSortingFeature });

const userColumnHelper = createColumnHelper<typeof USER_TABLE_FEATURES, User>();

const NO_USERS: User[] = [];

const STATUS_LABELS: Record<string, string> = {
  active: "啟用",
  inactive: "停用",
  pending: "待審核",
};

const STATUS_VARIANTS: Record<string, "success" | "warning" | "secondary"> = {
  active: "success",
  pending: "warning",
  inactive: "secondary",
};

const ROLE_OPTIONS = [
  "company_admin",
  "dept_admin",
  "staff",
  "customer",
  "guest",
];

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.PermissionDenied:
        return "沒有權限執行此操作";
      case Code.InvalidArgument:
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
      case Code.Unavailable:
        return "無法連線至伺服器,請確認後端服務已啟動";
      default:
        return err.rawMessage || "操作失敗,請稍後再試";
    }
  }
  return "無法連線至伺服器,請確認後端服務已啟動";
}

/**
 * 使用者管理頁(/users/users)。
 * 版型同 `DepartmentsPage`（篩選卡片色帶＋表格 `Card`）。差異：
 * - 後端 `ListUsers` 無 `sort`/`desc` 參數 → 不開排序（欄位 `enableSorting: false`）。
 * - 篩選：公司（下拉）／角色／狀態；送出才進 query key。
 * - 操作：停用（`deactivateUser`）／指派角色（`assignRole`，guest 審核入口）。
 */
export default function UsersPage() {
  const [companyFilter, setCompanyFilter] = createSignal("");
  const [roleFilter, setRoleFilter] = createSignal("");
  const [statusFilter, setStatusFilter] = createSignal("");
  const [filter, setFilter] = createSignal({ company: "", role: "", status: "" });
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: PAGE_SIZE,
  });
  const [companyKeyword, setCompanyKeyword] = createSignal("");
  const [companySearch, setCompanySearch] = createSignal("");
  const [pinnedCompanies, setPinnedCompanies] = createSignal<Company[]>([]);
  const [departments, setDepartments] = createSignal<Department[]>([]);

  const client = useQueryClient();
  const sortableOff = (_c: unknown, label: string) => label;

  const columns = userColumnHelper.columns([
    userColumnHelper.accessor("name", {
      enableSorting: false,
      header: (ctx) => sortableOff(ctx.column, "姓名"),
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    userColumnHelper.accessor("email", {
      enableSorting: false,
      header: (ctx) => sortableOff(ctx.column, "Email"),
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    userColumnHelper.accessor("role", {
      enableSorting: false,
      header: (ctx) => sortableOff(ctx.column, "角色"),
      cell: (info) => <span>{info.getValue()}</span>,
    }),
    userColumnHelper.accessor("status", {
      enableSorting: false,
      header: (ctx) => sortableOff(ctx.column, "狀態"),
      cell: (info) => {
        const s = info.getValue();
        return (
          <Badge variant={STATUS_VARIANTS[s] ?? "secondary"}>{STATUS_LABELS[s] ?? s}</Badge>
        );
      },
    }),
    userColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="text-right">
          <button
            type="button"
            onClick={() => void openAssign(info.row.original)}
            class="font-medium text-primary hover:underline"
          >
            指派角色
          </button>
          <button
            type="button"
            onClick={() => void deactivate(info.row.original)}
            class="ml-3 font-medium text-destructive hover:underline"
          >
            停用
          </button>
        </div>
      ),
    }),
  ]);

  const query = createQuery(() =>
    usersQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      companyId: filter().company || undefined,
      role: filter().role || undefined,
      status: filter().status || undefined,
    })
  );

  const total = () => Number(query.data?.pagination?.total ?? 0);

  const table = createTable({
    features: USER_TABLE_FEATURES,
    columns,
    get data() {
      return query.data?.users ?? NO_USERS;
    },
    get rowCount() {
      return total();
    },
    manualPagination: true,
    get state() {
      return { pagination: pagination() };
    },
    onPaginationChange: setPagination,
  });

  const companyOptions = createInfiniteQuery(() =>
    companyDropdownQueryOptions({ keyword: companySearch() || undefined })
  );

  const companies = (): Company[] => {
    const seen = new Set<string>();
    const collect = (list: Company[]) =>
      list.filter((c) => {
        if (seen.has(c.id)) return false;
        seen.add(c.id);
        return true;
      });
    return [
      ...collect(pinnedCompanies()),
      ...collect((companyOptions.data?.pages ?? []).flatMap((p) => p.companies)),
    ];
  };

  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / pagination().pageSize));
    if (pagination().pageIndex + 1 > maxPage) table.setPageIndex(maxPage - 1);
  });

  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [assignOpen, setAssignOpen] = createSignal(false);
  const [assignTarget, setAssignTarget] = createSignal<User | null>(null);
  const [assignRole, setAssignRole] = createSignal("");
  const [assignDept, setAssignDept] = createSignal("");
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [deleteError, setDeleteError] = createSignal<string | null>(null);

  const bannerError = () =>
    query.error
      ? errorMessage(query.error)
      : companyOptions.error
        ? errorMessage(companyOptions.error)
        : deleteError();

  createEffect(() => {
    if (query.isFetching) setDeleteError(null);
  });

  const nameValidators = fieldValidators(userSchema.entries.name);
  const emailValidators = fieldValidators(userSchema.entries.email);
  const companyValidators = fieldValidators(userSchema.entries.company);
  const roleValidators = fieldValidators(userSchema.entries.role);

  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter({ company: companyFilter(), role: roleFilter(), status: statusFilter() });
      table.setPageIndex(0);
    });
  };

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { name: "", email: "", company: "", department: "", role: "", phone: "", employeeNo: "" },
    onSubmit: async ({ value }) => {
      try {
        await userClient.createUser({
          name: value.name.trim(),
          email: value.email.trim(),
          companyId: value.company,
          departmentId: value.department || "",
          role: value.role,
          phone: value.phone,
          employeeNo: value.employeeNo,
        });
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["users"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const openDialog = () => {
    setServerError(undefined);
    form.reset({
      name: "",
      email: "",
      company: companyFilter() || companies()[0]?.id || "",
      department: "",
      role: "",
      phone: "",
      employeeNo: "",
    });
    setDialogOpen(true);
  };

  /** modal 公司切換時重載該公司部門（指派角色的部門下拉共用）。 */
  const loadDepartments = async (companyId: string) => {
    if (!companyId) {
      setDepartments([]);
      return;
    }
    try {
      const res = await departmentClient.listDepartments({ page: 1, pageSize: 100, companyId });
      setDepartments(res.departments);
    } catch {
      setDepartments([]);
    }
  };

  const openAssign = async (u: User) => {
    setAssignTarget(u);
    setAssignRole("");
    setAssignDept("");
    setServerError(undefined);
    await loadDepartments(u.companyId);
    setAssignOpen(true);
  };

  const submitAssign = async () => {
    const target = assignTarget();
    if (!target || !assignRole()) return;
    try {
      await userClient.assignRole({
        userId: target.id,
        role: assignRole(),
        departmentId: assignDept(),
      });
      setAssignOpen(false);
      await client.invalidateQueries({ queryKey: ["users"] });
    } catch (err) {
      setServerError(errorMessage(err));
    }
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const deactivate = async (u: User) => {
    if (!window.confirm(`確定停用使用者「${u.name}」?`)) return;
    setDeleteError(null);
    try {
      await userClient.deactivate({ userId: u.id });
      await client.invalidateQueries({ queryKey: ["users"] });
    } catch (err) {
      setDeleteError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">使用者管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">員工與客戶帳號(共 {total()} 筆)</p>
        </div>
        <Button type="button" onClick={openDialog}>
          新增使用者
        </Button>
      </header>

      <Show when={bannerError()}>
        {(message) => (
          <p
            class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
            role="alert"
          >
            {message()}
          </p>
        )}
      </Show>

      <Card>
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={(e) => {
            e.preventDefault();
            setCompanySearch(companyKeyword());
            setPinnedCompanies([]);
          }}
        >
          <Field class="w-full sm:w-64">
            <FieldLabel for="user-company-search">公司關鍵字</FieldLabel>
            <Input
              id="user-company-search"
              value={companyKeyword()}
              onInput={(e) => setCompanyKeyword(e.currentTarget.value)}
              placeholder="名稱 / 識別碼"
            />
          </Field>
          <Button type="submit" variant="outline" loading={companyOptions.isFetching}>
            搜尋公司
          </Button>
        </form>
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={submitFilter}
        >
          <Field class="w-full sm:w-48">
            <FieldLabel for="user-company-filter">所屬公司</FieldLabel>
            <select
              id="user-company-filter"
              value={companyFilter()}
              onChange={(e) => setCompanyFilter(e.currentTarget.value)}
            >
              <option value="">全部公司</option>
              <For each={companies()}>{(c) => <option value={c.id}>{c.name}</option>}</For>
            </select>
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="user-role-filter">角色</FieldLabel>
            <select
              id="user-role-filter"
              value={roleFilter()}
              onChange={(e) => setRoleFilter(e.currentTarget.value)}
            >
              <option value="">全部角色</option>
              <For each={ROLE_OPTIONS}>{(r) => <option value={r}>{r}</option>}</For>
            </select>
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="user-status-filter">狀態</FieldLabel>
            <select
              id="user-status-filter"
              value={statusFilter()}
              onChange={(e) => setStatusFilter(e.currentTarget.value)}
            >
              <option value="">全部狀態</option>
              <option value="active">啟用</option>
              <option value="pending">待審核</option>
              <option value="inactive">停用</option>
            </select>
          </Field>
          <Button type="submit" variant="outline">
            篩選
          </Button>
        </form>

        <Table>
          <TableHeader>
            <For each={table.getHeaderGroups()}>
              {(hg) => (
                <TableRow>
                  <For each={hg.headers}>
                    {(h) => (
                      <TableHead>{flexRender(h.column.columnDef.header, h.getContext())}</TableHead>
                    )}
                  </For>
                </TableRow>
              )}
            </For>
          </TableHeader>
          <TableBody>
            <Show
              when={query.data?.users?.length}
              fallback={
                <TableRow>
                  <TableCell colSpan={5}>
                    {query.isFetching ? "載入中…" : "尚無使用者"}
                  </TableCell>
                </TableRow>
              }
            >
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow>
                    <For each={row.getAllCells()}>
                      {(cell) => (
                        <TableCell>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                      )}
                    </For>
                  </TableRow>
                )}
              </For>
            </Show>
          </TableBody>
        </Table>

        <ListPagination
          total={table.getRowCount()}
          pageSize={table.atoms.pagination.get().pageSize}
          page={table.atoms.pagination.get().pageIndex + 1}
          onPageChange={(page) => table.setPageIndex(page - 1)}
        />
      </Card>

      <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新增使用者</DialogTitle>
            <DialogDescription>建立員工帳號（客戶帳號由客戶主檔建立）</DialogDescription>
          </DialogHeader>
          <Show when={serverError()}>
            {(message) => (
              <p class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive" role="alert">
                {message()}
              </p>
            )}
          </Show>
          <form onSubmit={submit}>
            <form.Field name="name" validators={nameValidators}>
              {(field) => (
                <Field>
                  <FieldLabel for="user-name">姓名</FieldLabel>
                  <Input
                    id="user-name"
                    value={field().state.value}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                    placeholder="王小明"
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>
            <form.Field name="email" validators={emailValidators}>
              {(field) => (
                <Field>
                  <FieldLabel for="user-email">Email</FieldLabel>
                  <Input
                    id="user-email"
                    value={field().state.value}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                    placeholder="user@example.com"
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>
            <form.Field name="company" validators={companyValidators}>
              {(field) => (
                <Field>
                  <FieldLabel for="user-company">所屬公司</FieldLabel>
                  <select
                    id="user-company"
                    value={field().state.value}
                    onChange={(e) => {
                      field().handleChange(e.currentTarget.value);
                      void loadDepartments(e.currentTarget.value);
                    }}
                  >
                    <option value="">請選擇</option>
                    <For each={companies()}>{(c) => <option value={c.id}>{c.name}</option>}</For>
                  </select>
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>
            <form.Field name="department">
              {(field) => (
                <Field>
                  <FieldLabel for="user-department">所屬部門（選填）</FieldLabel>
                  <select
                    id="user-department"
                    value={field().state.value}
                    onChange={(e) => field().handleChange(e.currentTarget.value)}
                  >
                    <option value="">無</option>
                    <For each={departments()}>{(d) => <option value={d.id}>{d.name}</option>}</For>
                  </select>
                </Field>
              )}
            </form.Field>
            <form.Field name="role" validators={roleValidators}>
              {(field) => (
                <Field>
                  <FieldLabel for="user-role">角色</FieldLabel>
                  <select
                    id="user-role"
                    value={field().state.value}
                    onChange={(e) => field().handleChange(e.currentTarget.value)}
                  >
                    <option value="">請選擇</option>
                    <For each={ROLE_OPTIONS}>{(r) => <option value={r}>{r}</option>}</For>
                  </select>
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>
            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>
                取消
              </DialogClose>
              <Button type="submit" loading={isSubmitting()}>
                建立
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={assignOpen()} onOpenChange={setAssignOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>指派角色</DialogTitle>
            <DialogDescription>
              {assignTarget()?.name}（目前：{assignTarget()?.role}）
            </DialogDescription>
          </DialogHeader>
          <Show when={serverError()}>
            {(message) => (
              <p class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive" role="alert">
                {message()}
              </p>
            )}
          </Show>
          <Field>
            <FieldLabel for="assign-role">新角色</FieldLabel>
            <select
              id="assign-role"
              value={assignRole()}
              onChange={(e) => setAssignRole(e.currentTarget.value)}
            >
              <option value="">請選擇</option>
              <For each={ROLE_OPTIONS}>{(r) => <option value={r}>{r}</option>}</For>
            </select>
          </Field>
          <Field>
            <FieldLabel for="assign-dept">部門（dept_admin/staff 必填）</FieldLabel>
            <select
              id="assign-dept"
              value={assignDept()}
              onChange={(e) => setAssignDept(e.currentTarget.value)}
            >
              <option value="">無</option>
              <For each={departments()}>{(d) => <option value={d.id}>{d.name}</option>}</For>
            </select>
          </Field>
          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline" })}>
              取消
            </DialogClose>
            <Button type="button" onClick={() => void submitAssign()} disabled={!assignRole()}>
              確認指派
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  );
}
