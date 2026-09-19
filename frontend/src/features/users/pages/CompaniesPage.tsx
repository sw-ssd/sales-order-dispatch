import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
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
import { batch, createEffect, createSignal, For, Show, type JSX } from "solid-js";
import type { Company } from "~/lib/proto/salesorder/v1/company_pb";
import { appFormOptions, fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../components/ListPagination";
import { companiesQueryOptions, companyClient, PAGE_SIZE } from "../queries";
import { companySchema } from "../schemas";

/**
 * 新增模式的欄位預設值。`form.reset(values)` 會把傳入的 values **整份取代** `defaultValues`
 * （form-core 1.33.5 `FormApi.js`：`if (values && !opts?.keepDefaultValues) this.options = { ...this.options, defaultValues: values }`），
 * 所以每次開啟都要顯式帶完整 values（編輯模式帶入該筆公司；新增模式帶一份新複本），
 * 不能只靠 `form.reset()`，也不能只帶部分欄位（未帶到的欄位值會變 `undefined`）。
 */
const EMPTY_COMPANY_VALUES = { name: "", identifier: "", taxId: "", status: "active" };

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
 * 清單資料（列資料／總筆數／載入與錯誤狀態）一律來自 `../queries.ts` 的
 * `companiesQueryOptions`＋`createQuery`：頁面持有的只有「查詢輸入」（篩選草稿、已套用篩選、
 * 頁碼），不再另存一份結果快取。mutation 成功後以 `invalidateQueries(["companies"])` 前綴失效。
 *
 * modal 的欄位值、欄位驗證與提交狀態由 `createForm` 持有：驗證時機為 `onBlur` + `onSubmit`
 * （輸入過程不標紅），客戶端錯誤落在該欄下方；伺服器錯誤不對應特定欄位，由提交流程設進
 * 表單層 banner。篩選列（關鍵字／狀態）仍是獨立的查詢表單，不受 modal 的 form 管轄。
 */
export default function CompaniesPage() {
  // 篩選草稿：輸入過程只動這兩個 signal，不進 query key（D5：不得每按一鍵就查詢）。
  const [keyword, setKeyword] = createSignal("");
  const [statusFilter, setStatusFilter] = createSignal("");
  // 已套用的篩選＋頁碼是 query key 的來源：只有送出篩選與換頁會動它們。
  const [filter, setFilter] = createSignal({ keyword: "", status: "" });
  const [page, setPage] = createSignal(1);

  const client = useQueryClient();

  // 清單唯一的資料來源：列資料、總筆數、載入與錯誤狀態全部由 query 狀態推導，
  // 頁面不再另存 companies/total/loading/error signal（避免兩份真相）。
  const query = createQuery(() =>
    companiesQueryOptions({
      page: page(),
      pageSize: PAGE_SIZE,
      keyword: filter().keyword || undefined,
      status: filter().status || undefined,
    })
  );

  const total = () => Number(query.data?.pagination?.total ?? 0);

  /**
   * 超頁退回：回傳的 total 讓目前頁碼超界時（例：該頁資料被刪光），把頁碼夾到合法值。
   * `page` 是 query key 的一部分 → `setPage` 自己就會觸發重取，不必也不能再手動重載；
   * 夾到的頁碼必定 ≤ maxPage < 原頁碼（嚴格遞減、下界 1），所以重取次數有界。
   *
   * `isPlaceholderData` 期間的 `data` 屬於前一個 key（placeholderData 保留的舊結果），
   * 據以退回會把剛切過去的頁碼彈回來，故必須排除；這不會漏掉退回——新資料一到，
   * `data` 與 `isPlaceholderData` 都變動，這個 effect 會再跑一次。
   */
  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / PAGE_SIZE));
    if (page() > maxPage) setPage(maxPage);
  });

  // 表單對話框狀態
  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Company | null>(null);
  // 伺服器錯誤不是驗證狀態（不對應任何欄位），由提交流程設定，落在表單層 banner。
  const [serverError, setServerError] = createSignal<string | undefined>();
  // 刪除失敗不屬於任何 query（不是清單資料），單獨保留一顆區域 signal；
  // 清單載入失敗一律由 `query.error` 驅動（banner 顯示處直接組兩者）。
  const [deleteError, setDeleteError] = createSignal<string | null>(null);

  // 清單重新查詢時清掉上一次刪除留下的錯誤 banner（與改寫前 `load()` 開頭 `setError(null)` 等價）。
  createEffect(() => {
    if (query.isFetching) setDeleteError(null);
  });

  // 欄位值、欄位錯誤與提交狀態由 `createForm` 持有；驗證時機為 `onBlur` + `onSubmit`
  // （輸入過程不標紅），客戶端錯誤落在該欄下方。
  const nameValidators = fieldValidators(companySchema.entries.name);
  const identifierValidators = fieldValidators(companySchema.entries.identifier);

  const form = createForm(() => ({
    // `canSubmitWhenInvalid` 的理由（F3）見 form-helpers.ts 的 `appFormOptions`。
    ...appFormOptions,
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
        // 清單資料已變更 → 前綴失效（清單 query 立即重取，不需手動 reload）。
        await client.invalidateQueries({ queryKey: ["companies"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  /**
   * 送出篩選：把草稿套進 query key 並回第 1 頁。兩個 signal 必須放在同一個 `batch` 內——
   * 分開寫會先以「舊頁碼＋新篩選」查一次、再以「第 1 頁＋新篩選」查一次（兩次 RPC）。
   */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter({ keyword: keyword(), status: statusFilter() });
      setPage(1);
    });
  };

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
    setDeleteError(null);
    try {
      await companyClient.deleteCompany({ companyId: c.id });
      // 刪除後清單重取；若刪到當前頁超界，超頁退回的 effect 會把頁碼夾回合法值。
      await client.invalidateQueries({ queryKey: ["companies"] });
    } catch (err) {
      setDeleteError(errorMessage(err));
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

      <Show when={query.error ? errorMessage(query.error) : deleteError()}>
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
          onSubmit={submitFilter}
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
            <Show when={query.isPending}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={6} class="py-8 text-center text-muted-foreground">
                  載入中…
                </TableCell>
              </TableRow>
            </Show>
            <Show when={!query.isPending && (query.data?.companies.length ?? 0) === 0}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={6} class="py-8 text-center text-muted-foreground">
                  尚無公司資料
                </TableCell>
              </TableRow>
            </Show>
            <For each={query.data?.companies ?? []}>
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
          onPageChange={setPage}
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

            {/* status 無 validators → `meta.isValid` 恆真、永不 invalid，故此欄不需要
                `aria-invalid`／`aria-describedby` 接線（與 DepartmentsPage 的 select 情境不同，不是缺漏）。 */}
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
