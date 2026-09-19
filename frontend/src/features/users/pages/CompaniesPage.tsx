import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createForm } from "@tanstack/solid-form";
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
import { createSignal, For, onMount, Show, type JSX } from "solid-js";
import {
  CompanyService,
  type Company,
} from "~/lib/proto/salesorder/v1/company_pb";
import { fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../components/ListPagination";
import { companySchema } from "../schemas";

const PAGE_SIZE = 20;

/**
 * 新增模式的欄位預設值。`form.reset(values)` 會把傳入的 values 併成新的 `defaultValues`
 * （編輯模式帶入該筆公司），所以每次開啟都要顯式帶一份新的複本，不能只靠 `form.reset()`。
 */
const EMPTY_COMPANY_VALUES = { name: "", identifier: "", taxId: "", status: "active" };

const companyClient = createClient(
  CompanyService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);

const STATUS_LABELS: Record<string, string> = {
  active: "啟用",
  inactive: "停用",
  suspended: "暫停",
};

const STATUS_VARIANTS: Record<string, "success" | "warning" | "secondary"> = {
  active: "success",
  inactive: "secondary",
  suspended: "warning",
};

/**
 * 原生 select 的視覺（ui/ 沒有 select 元件）。
 * `pr-10` 對齊 Tailkit a-c-form-elements-03：`@tailwindcss/forms` 的箭頭畫在
 * `right .5rem center` 且佔 1.5em，而外掛的 `padding-right: 2.5rem` 落在 base layer、
 * 會被頁面 utilities 蓋掉（實測只剩 12px），不補回箭頭就會疊在選項文字上。
 * 底色與 focus ring 一併覆蓋外掛在 base layer 的硬編值（`#fff` 底、1px 藍 ring）。
 */
const SELECT_CLASS =
  "block w-full rounded-lg border border-border bg-card py-2 pr-10 pl-3 text-sm text-foreground focus:border-primary focus:ring-3 focus:ring-primary/50 focus:outline-none disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground";

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.AlreadyExists:
        return "識別碼(identifier)已存在,請換一個";
      case Code.FailedPrecondition:
        return err.rawMessage || "無法刪除:仍被其他資料參照";
      case Code.NotFound:
        return "資料不存在或已被刪除";
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
 * 公司主檔 CRUD 頁(/users/companies)。
 * 版型照 Tailkit（Page Headings + In Card 表格）：標題區塊帶下框線、篩選列為卡片色帶、
 * 表格與分頁收在同一張 `Card` 內。頁面本身不帶內距——內距由 AppShell 內容區（`p-4 lg:p-6`）負責。
 *
 * modal 的欄位值、欄位驗證與提交狀態由 `createForm` 持有：驗證時機為 `onBlur` + `onSubmit`
 * （輸入過程不標紅），客戶端錯誤落在該欄下方；伺服器錯誤不對應特定欄位，由提交流程設進
 * 表單層 banner。篩選列（關鍵字／狀態）仍是獨立的查詢表單，不受 modal 的 form 管轄。
 */
export default function CompaniesPage() {
  const [companies, setCompanies] = createSignal<Company[]>([]);
  const [total, setTotal] = createSignal(0);
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal<string | null>(null);

  const [keyword, setKeyword] = createSignal("");
  const [statusFilter, setStatusFilter] = createSignal("");
  const [page, setPage] = createSignal(1);

  const goToPage = (p: number) => {
    setPage(p);
    load();
  };

  // 表單對話框狀態
  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Company | null>(null);
  // 伺服器錯誤不是驗證狀態（不對應任何欄位），由提交流程設定，落在表單層 banner。
  const [serverError, setServerError] = createSignal<string | undefined>();

  // 欄位值、欄位錯誤與提交狀態由 `createForm` 持有；驗證時機為 `onBlur` + `onSubmit`
  // （輸入過程不標紅），客戶端錯誤落在該欄下方。
  const nameValidators = fieldValidators(companySchema.entries.name);
  const identifierValidators = fieldValidators(companySchema.entries.identifier);

  const form = createForm(() => ({
    defaultValues: { ...EMPTY_COMPANY_VALUES },
    onSubmit: async ({ value }) => {
      const current = editing();
      const name = value.name.trim();
      const taxId = value.taxId.trim();
      try {
        if (current) {
          await companyClient.updateCompany({
            companyId: current.id,
            name,
            taxId,
            status: value.status,
          });
        } else {
          await companyClient.createCompany({
            name,
            taxId,
            identifier: value.identifier.trim(),
            status: value.status,
          });
        }
        setDialogOpen(false);
        await load();
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await companyClient.listCompanies({
        page: page(),
        pageSize: PAGE_SIZE,
        status: statusFilter() || undefined,
        keyword: keyword() || undefined,
      });
      setCompanies(res.companies);
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
  onMount(load);

  /**
   * 開啟 modal（新增傳 `null`）。表單欄位值、欄位錯誤與 touched 由 `form.reset(values)` 重設；
   * `serverError` 是元件層 signal、**不受 `form.reset()` 影響**，必須在這裡顯式清掉，
   * 否則重開會重現上一次的伺服器錯誤 banner（規格 R1）。
   */
  const openDialog = (company: Company | null) => {
    setEditing(company);
    setServerError(undefined);
    form.reset(
      company
        ? {
            name: company.name,
            identifier: company.identifier,
            taxId: company.taxId,
            status: company.status || "active",
          }
        : { ...EMPTY_COMPANY_VALUES }
    );
    setDialogOpen(true);
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

  const remove = async (c: Company) => {
    if (!window.confirm(`確定刪除公司「${c.name}」?`)) return;
    try {
      await companyClient.deleteCompany({ companyId: c.id });
      await load();
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">公司管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">多租戶公司主檔(共 {total()} 筆)</p>
        </div>
        <Button type="button" onClick={() => openDialog(null)}>
          新增公司
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
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={(e) => {
            e.preventDefault();
            setPage(1); // 查詢變更時回到第一頁
            load();
          }}
        >
          <Field class="w-full sm:w-64">
            <FieldLabel for="company-keyword">關鍵字</FieldLabel>
            <Input
              id="company-keyword"
              value={keyword()}
              onInput={(e) => setKeyword(e.currentTarget.value)}
              placeholder="名稱 / 識別碼"
            />
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="company-status-filter">狀態</FieldLabel>
            <select
              id="company-status-filter"
              value={statusFilter()}
              onChange={(e) => setStatusFilter(e.currentTarget.value)}
              class={SELECT_CLASS}
            >
              <option value="">全部</option>
              <option value="active">啟用</option>
              <option value="inactive">停用</option>
              <option value="suspended">暫停</option>
            </select>
          </Field>
          <Button type="submit" variant="outline">
            查詢
          </Button>
        </form>

        <Table>
          <TableHeader>
            <TableRow class="hover:bg-transparent">
              <TableHead>名稱</TableHead>
              <TableHead>識別碼</TableHead>
              <TableHead>統一編號</TableHead>
              <TableHead>狀態</TableHead>
              <TableHead>ID</TableHead>
              <TableHead class="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <Show when={loading()}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={6} class="py-8 text-center text-muted-foreground">
                  載入中…
                </TableCell>
              </TableRow>
            </Show>
            <Show when={!loading() && companies().length === 0}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={6} class="py-8 text-center text-muted-foreground">
                  尚無公司資料
                </TableCell>
              </TableRow>
            </Show>
            <For each={companies()}>
              {(c) => (
                <TableRow>
                  <TableCell class="font-medium text-foreground">{c.name}</TableCell>
                  <TableCell class="text-muted-foreground">{c.identifier}</TableCell>
                  <TableCell class="text-muted-foreground">{c.taxId || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={STATUS_VARIANTS[c.status] ?? "secondary"}>
                      {STATUS_LABELS[c.status] ?? c.status}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-muted-foreground">{c.id}</TableCell>
                  <TableCell class="text-right">
                    <button
                      type="button"
                      onClick={() => openDialog(c)}
                      class="font-medium text-primary hover:underline"
                    >
                      編輯
                    </button>
                    <button
                      type="button"
                      onClick={() => remove(c)}
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
            <DialogTitle>{editing() ? "編輯公司" : "新增公司"}</DialogTitle>
            <DialogDescription>
              {editing() ? "修改名稱、統一編號或狀態" : "建立新的公司主檔"}
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
                  <FieldLabel for="company-name">公司名稱 *</FieldLabel>
                  <Input
                    id="company-name"
                    required
                    value={field().state.value}
                    onBlur={field().handleBlur}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>

            <form.Field name="identifier" validators={identifierValidators}>
              {(field) => (
                <Field invalid={!field().state.meta.isValid}>
                  <FieldLabel for="company-identifier">識別碼(identifier) *</FieldLabel>
                  <Input
                    id="company-identifier"
                    required
                    disabled={!!editing()}
                    value={field().state.value}
                    onBlur={field().handleBlur}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                    placeholder="建立後不可修改"
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>

            <form.Field name="taxId">
              {(field) => (
                <Field>
                  <FieldLabel for="company-tax-id">統一編號</FieldLabel>
                  <Input
                    id="company-tax-id"
                    value={field().state.value}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                  />
                </Field>
              )}
            </form.Field>

            <form.Field name="status">
              {(field) => (
                <Field>
                  <FieldLabel for="company-status">狀態</FieldLabel>
                  <select
                    id="company-status"
                    value={field().state.value}
                    onChange={(e) => field().handleChange(e.currentTarget.value)}
                    class={SELECT_CLASS}
                  >
                    <option value="active">啟用</option>
                    <option value="inactive">停用</option>
                    <option value="suspended">暫停</option>
                  </select>
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
