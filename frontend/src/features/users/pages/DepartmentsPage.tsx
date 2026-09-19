import { useFieldContext } from "@ark-ui/solid/field";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createForm } from "@tanstack/solid-form";
import {
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
import { createSignal, For, onMount, Show, type Component, type JSX } from "solid-js";
import {
  CompanyService,
  DepartmentService,
  type Company,
  type Department,
} from "~/lib/proto/salesorder/v1/company_pb";
import { fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../components/ListPagination";
import { departmentSchema } from "../schemas";

const PAGE_SIZE = 20;
const COMPANY_PAGE_SIZE = 50;

/**
 * 新增模式的欄位預設值。`form.reset(values)` 會把傳入的 values **整份取代** `defaultValues`
 * （form-core 1.33.5 `FormApi.js`：`if (values && !opts?.keepDefaultValues) this.options = { ...this.options, defaultValues: values }`），
 * 所以每次開啟都要顯式帶完整 values（編輯模式帶入該筆部門；新增模式帶一份新複本），
 * 不能只靠 `form.reset()`，也不能只帶部分欄位（未帶到的欄位值會變 `undefined`）。
 */
const EMPTY_DEPARTMENT_VALUES = { name: "", company: "" };

const departmentClient = createClient(
  DepartmentService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);
const companyClient = createClient(
  CompanyService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);

/**
 * 原生 select 的視覺（ui/ 沒有 select 元件）。`pr-10` 與顏色覆蓋的理由同 `CompaniesPage.tsx`。
 * invalid 變體與 `ui/input.tsx` 一致：`Field` 標紅時，同一個 modal 裡的 Input 與 select 都是紅框。
 */
const SELECT_CLASS =
  "block w-full rounded-lg border border-border bg-card py-2 pr-10 pl-3 text-sm text-foreground focus:border-primary focus:ring-3 focus:ring-primary/50 focus:outline-none aria-invalid:border-destructive aria-invalid:focus-visible:ring-destructive/50 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground";

/**
 * 原生 select（`ui/` 沒有 select 元件，用 `SELECT_CLASS` 手寫視覺）。
 * `Input` 接手的 Ark `Field` 關聯（`aria-invalid`／`aria-describedby`）只服務 `<input>`，
 * 所以這裡比照它的做法補上同一組屬性；`Field` 之外使用時完全不影響。
 *
 * 這段 `controlA11y` 與 `ui/input.tsx` 的實作逐字相同 —— 複製是**有期限的**：`ui/**` 在本階段凍結，
 * 且 Ark `Field.Select` 的錯誤走 `aria-errormessage`（`aria-describedby` 為 null），與本專案
 * `aria-describedby` 的既有語意不一致。收斂時機 = Phase 1 遺留的同一批 housekeeping
 * （`ui/input.tsx` 的 `aria-describedby` 改走 Ark `getInputProps()`，見
 * `.superpowers/sdd/2026-09-19-frontend-ui-library-phase1-plan/deferred-minors.md`）：
 * 該批處理時，這裡要一併改為共用的 `ui` 層 select（或同步改採 `aria-errormessage`）。
 */
const SelectControl: Component<JSX.SelectHTMLAttributes<HTMLSelectElement>> = (props) => {
  const field = useFieldContext();

  const controlA11y = (): JSX.SelectHTMLAttributes<HTMLSelectElement> => {
    const api = field?.();
    if (!api) return {};
    const describedBy = [
      api.invalid ? api.ids.errorText : undefined,
      api.ariaDescribedby,
    ]
      .filter(Boolean)
      .join(" ");
    return {
      "aria-invalid": api.invalid ? "true" : undefined,
      "aria-describedby": describedBy || undefined,
    };
  };

  return <select class={SELECT_CLASS} {...controlA11y()} {...props} />;
};

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.FailedPrecondition:
        return err.rawMessage || "無法刪除:仍被其他資料參照";
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
 * 部門主檔 CRUD 頁(/users/departments)。
 * 版型同 `CompaniesPage`（Page Headings + In Card）：兩組篩選表單收在同一條卡片色帶內，
 * 表格與分頁收在同一張 `Card`。內距由 AppShell 內容區負責。
 */
export default function DepartmentsPage() {
  const [departments, setDepartments] = createSignal<Department[]>([]);
  const [companies, setCompanies] = createSignal<Company[]>([]);
  const [total, setTotal] = createSignal(0);
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal<string | null>(null);

  const [companyFilter, setCompanyFilter] = createSignal("");
  // 公司下拉:可搜尋(keyword)+ 分頁載入
  const [companyKeyword, setCompanyKeyword] = createSignal("");
  const [companyPage, setCompanyPage] = createSignal(1);
  const [companyTotal, setCompanyTotal] = createSignal(0);
  const [companiesLoading, setCompaniesLoading] = createSignal(false);
  const [page, setPage] = createSignal(1);

  const goToPage = (p: number) => {
    setPage(p);
    load();
  };

  // 表單對話框狀態
  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Department | null>(null);
  // 伺服器回應的錯誤訊息（元件層 signal，不是驗證狀態，也不對應任何欄位）。
  const [serverError, setServerError] = createSignal<string | undefined>();

  const nameValidators = fieldValidators(departmentSchema.entries.name);
  const companyValidators = fieldValidators(departmentSchema.entries.company);

  const loadCompanies = async (reset: boolean) => {
    const target = reset ? 1 : companyPage();
    const keyword = companyKeyword() || undefined;
    setCompaniesLoading(true);
    setError(null);
    try {
      const res = await companyClient.listCompanies({
        page: target,
        pageSize: COMPANY_PAGE_SIZE,
        keyword,
      });
      setCompanies((prev) => {
        if (reset) return res.companies;
        const seen = new Set(prev.map((c) => c.id));
        return [...prev, ...res.companies.filter((c) => !seen.has(c.id))];
      });
      setCompanyTotal(Number(res.pagination?.total ?? 0));
      setCompanyPage(target + 1);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setCompaniesLoading(false);
    }
  };

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await departmentClient.listDepartments({
        page: page(),
        pageSize: PAGE_SIZE,
        companyId: companyFilter() || undefined,
      });
      setDepartments(res.departments);
      const t = Number(res.pagination?.total ?? 0);
      setTotal(t);
      // 刪除/篩選後若目前頁碼超出總頁數,退回最後一頁並重新載入
      const maxPage = Math.max(1, Math.ceil(t / PAGE_SIZE));
      if (page() > maxPage) {
        setPage(maxPage);
        return load();
      }
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoading(false);
    }
  };

  onMount(async () => {
    await loadCompanies(true);
    await load();
  });

  const form = createForm(() => ({
    defaultValues: { ...EMPTY_DEPARTMENT_VALUES },
    onSubmit: async ({ value }) => {
      const current = editing();
      // 名稱照舊 trim 後才送出；`company` 是公司 id，更新時不帶（改寫前就沒有更新公司）。
      const name = value.name.trim();
      try {
        if (current) {
          await departmentClient.updateDepartment({ departmentId: current.id, name });
        } else {
          await departmentClient.createDepartment({ companyId: value.company, name });
        }
        setDialogOpen(false);
        await load();
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  /**
   * 開啟 modal（新增傳 `null`）。欄位值、欄位錯誤與 touched 由 `form.reset(values)` 重設；
   * `serverError` 是元件層 signal、**不受 `form.reset()` 影響**，必須在這裡顯式清掉，
   * 否則重開會重現上一次的伺服器錯誤 banner（規格 R1）。
   */
  const openDialog = (department: Department | null) => {
    setEditing(department);
    setServerError(undefined);
    form.reset(
      department
        ? { name: department.name, company: department.companyId }
        : {
            ...EMPTY_DEPARTMENT_VALUES,
            // 所屬公司沿用改寫前的挑選邏輯：目前篩選的公司，否則公司清單第一筆
            company: companyFilter() || companies()[0]?.id || "",
          }
    );
    setDialogOpen(true);
  };

  const openEdit = async (d: Department) => {
    // 該部門的公司可能不在已載入的分頁內,先補載再開啟
    if (!companies().some((c) => c.id === d.companyId)) {
      try {
        const res = await companyClient.getCompany({ companyId: d.companyId });
        const c = res.company;
        if (c) {
          setCompanies((prev) =>
            prev.some((x) => x.id === c.id) ? prev : [c, ...prev],
          );
        }
      } catch {
        // 找不到時仍可從下拉搜尋補上
      }
    }
    openDialog(d);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    // 提交中不重送（改寫前是 `if (saving()) return`）：form-core 的 `handleSubmit`
    // 只在第一次嘗試（`submissionAttempts <= 1`）擋下，進行中的第二次提交仍會送 API。
    if (isSubmitting()) return;
    // 客戶端驗證失敗時 `onSubmit` 不會被呼叫，舊的伺服器錯誤 banner 必須在這裡先清掉。
    setServerError(undefined);
    void form.handleSubmit();
  };

  const remove = async (d: Department) => {
    if (!window.confirm(`確定刪除部門「${d.name}」?`)) return;
    try {
      await departmentClient.deleteDepartment({ departmentId: d.id });
      await load();
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">部門管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            公司底下的業務單位(共 {total()} 筆)
          </p>
        </div>
        <Button type="button" onClick={() => openDialog(null)}>
          新增部門
        </Button>
      </header>

      <Show when={error()}>
        <p
          class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
          role="alert"
        >
          {error()}
        </p>
      </Show>

      <Card>
        <div class="flex flex-col gap-3 border-b border-border bg-muted px-3 py-3">
          <form
            class="flex flex-wrap items-end gap-3"
            onSubmit={(e) => {
              e.preventDefault();
              loadCompanies(true);
            }}
          >
            <Field class="w-full sm:w-64">
              <FieldLabel for="company-search-keyword">公司關鍵字</FieldLabel>
              <Input
                id="company-search-keyword"
                value={companyKeyword()}
                onInput={(e) => setCompanyKeyword(e.currentTarget.value)}
                placeholder="名稱 / 識別碼"
              />
            </Field>
            <Button type="submit" variant="outline" loading={companiesLoading()}>
              搜尋公司
            </Button>
            <div class="flex h-9 items-center gap-3 text-sm text-muted-foreground">
              <span>
                已載入 {companies().length} 家,共 {companyTotal()} 家
              </span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={companiesLoading() || companies().length >= companyTotal()}
                onClick={() => loadCompanies(false)}
              >
                載入更多
              </Button>
            </div>
          </form>

          <form
            class="flex flex-wrap items-end gap-3"
            onSubmit={(e) => {
              e.preventDefault();
              setPage(1); // 查詢變更時回到第一頁
              load();
            }}
          >
            <Field class="w-full sm:w-64">
              <FieldLabel for="department-company-filter">所屬公司</FieldLabel>
              <select
                id="department-company-filter"
                value={companyFilter()}
                onChange={(e) => setCompanyFilter(e.currentTarget.value)}
                class={SELECT_CLASS}
              >
                <option value="">全部公司</option>
                <For each={companies()}>
                  {(c) => <option value={c.id}>{c.name}</option>}
                </For>
              </select>
            </Field>
            <Button type="submit" variant="outline">
              查詢
            </Button>
          </form>
        </div>

        <Table>
          <TableHeader>
            <TableRow class="hover:bg-transparent">
              <TableHead>部門名稱</TableHead>
              <TableHead>所屬公司</TableHead>
              <TableHead>ID</TableHead>
              <TableHead class="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <Show when={loading()}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={4} class="py-8 text-center text-muted-foreground">
                  載入中…
                </TableCell>
              </TableRow>
            </Show>
            <Show when={!loading() && departments().length === 0}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={4} class="py-8 text-center text-muted-foreground">
                  尚無部門資料
                </TableCell>
              </TableRow>
            </Show>
            <For each={departments()}>
              {(d) => (
                <TableRow>
                  <TableCell class="font-medium text-foreground">{d.name}</TableCell>
                  <TableCell class="text-muted-foreground">{d.companyName || "—"}</TableCell>
                  <TableCell class="text-muted-foreground">{d.id}</TableCell>
                  <TableCell class="text-right">
                    <button
                      type="button"
                      onClick={() => openEdit(d)}
                      class="font-medium text-primary hover:underline"
                    >
                      編輯
                    </button>
                    <button
                      type="button"
                      onClick={() => remove(d)}
                      class="ml-3 font-medium text-destructive hover:underline"
                    >
                      刪除
                    </button>
                  </TableCell>
                </TableRow>
              )}
            </For>
          </TableBody>
        </Table>

        <ListPagination
          total={total()}
          pageSize={PAGE_SIZE}
          page={page()}
          onPageChange={goToPage}
        />
      </Card>

      <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing() ? "編輯部門" : "新增部門"}</DialogTitle>
            <DialogDescription>
              {editing() ? "修改部門名稱" : "在指定公司下建立部門"}
            </DialogDescription>
          </DialogHeader>

          {/*
            `novalidate`：必填規則已由 valibot 鏡射，原生驗證會用瀏覽器泡泡擋下 submit
            並讓自訂的繁中欄位錯誤沒有機會顯示；`required` 保留作為必填的語意標記。
          */}
          <form class="space-y-4" novalidate onSubmit={submit}>
            <form.Field name="name" validators={nameValidators}>
              {(field) => (
                <Field invalid={!field().state.meta.isValid}>
                  <FieldLabel for="department-name">部門名稱 *</FieldLabel>
                  <Input
                    id="department-name"
                    required
                    value={field().state.value}
                    onBlur={field().handleBlur}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>

            <form.Field name="company" validators={companyValidators}>
              {(field) => (
                <Field invalid={!field().state.meta.isValid}>
                  <FieldLabel for="department-company">所屬公司 *</FieldLabel>
                  <SelectControl
                    id="department-company"
                    required
                    disabled={!!editing()}
                    value={field().state.value}
                    onBlur={field().handleBlur}
                    onChange={(e) => field().handleChange(e.currentTarget.value)}
                  >
                    <option value="" disabled>
                      請選擇公司
                    </option>
                    <For each={companies()}>
                      {(c) => <option value={c.id}>{c.name}</option>}
                    </For>
                  </SelectControl>
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>

            <Show when={serverError()}>
              {(message) => (
                <p
                  class="rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
                  role="alert"
                >
                  {message()}
                </p>
              )}
            </Show>

            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>
                取消
              </DialogClose>
              <Button type="submit" loading={isSubmitting()}>
                儲存
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </main>
  );
}
