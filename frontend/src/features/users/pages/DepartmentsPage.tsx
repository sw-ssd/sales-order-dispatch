import { useFieldContext } from "@ark-ui/solid/field";
import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
import { createInfiniteQuery, createQuery, useQueryClient } from "@tanstack/solid-query";
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
import { batch, createEffect, createSignal, For, Show, type Component, type JSX } from "solid-js";
import { type Company, type Department } from "~/lib/proto/salesorder/v1/company_pb";
import { appFormOptions, fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../components/ListPagination";
import {
  companyClient,
  companyDropdownQueryOptions,
  departmentClient,
  departmentsQueryOptions,
  PAGE_SIZE,
} from "../queries";
import { departmentSchema } from "../schemas";

/**
 * 新增模式的欄位預設值。`form.reset(values)` 會把傳入的 values **整份取代** `defaultValues`
 * （form-core 1.33.5 `FormApi.js`：`if (values && !opts?.keepDefaultValues) this.options = { ...this.options, defaultValues: values }`），
 * 所以每次開啟都要顯式帶完整 values（編輯模式帶入該筆部門；新增模式帶一份新複本），
 * 不能只靠 `form.reset()`，也不能只帶部分欄位（未帶到的欄位值會變 `undefined`）。
 */
const EMPTY_DEPARTMENT_VALUES = { name: "", company: "" };

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
 * 這段 `controlA11y` 與 `ui/input.tsx` 的實作逐字相同（2 份實作）：**收斂時機 = `ui/` 出現 select 元件時**
 * （依 Phase 1 規則要同步 barrel + registry + 元檔 + demo，因此不為單一呼叫點先做）。
 * 注意：**不要**改走 Ark `Field.Select`／`getSelectProps()` —— 實測它的錯誤走 `aria-errormessage`、
 * `aria-describedby` 為 `null`（Ark 5.39.2），會讓本專案「錯誤必須在 `aria-describedby` 內」的驗收失效；
 * Phase 2 的 housekeeping 已就同一問題查證並裁定維持手寫組合（見 spec §7）。
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
 *
 * 清單資料（列資料／總筆數／載入與錯誤狀態）一律來自 `../queries.ts` 的
 * `departmentsQueryOptions`＋`createQuery`；公司下拉（篩選與 modal 的 `<select>` 共用同一份選項）
 * 來自 `companyDropdownQueryOptions`＋`createInfiniteQuery`（累積式：「載入更多」＝`fetchNextPage`）。
 * 頁面持有的只有「查詢輸入」（篩選草稿、已套用篩選、頁碼、公司關鍵字），不再另存結果快取。
 */
export default function DepartmentsPage() {
  // 篩選草稿：輸入過程只動這顆 signal，不進 query key（D5：不得每按一鍵就查詢）。
  const [companyFilter, setCompanyFilter] = createSignal("");
  // 已套用的篩選＋頁碼是 query key 的來源：只有送出篩選與換頁會動它們。
  const [filter, setFilter] = createSignal("");
  const [page, setPage] = createSignal(1);
  // 公司下拉：關鍵字草稿與已套用的關鍵字各一顆（送出「搜尋公司」才進 key）。
  const [companyKeyword, setCompanyKeyword] = createSignal("");
  const [companySearch, setCompanySearch] = createSignal("");

  // 編輯時以 `getCompany` 補載的那筆公司（該部門的公司可能不在已載入分頁內）；
  // 它是查詢結果之外的補充項，故單獨持有，再由 `companies()` 併進選項。
  const [pinnedCompanies, setPinnedCompanies] = createSignal<Company[]>([]);

  const client = useQueryClient();

  // 部門清單唯一的資料來源：列資料、總筆數、載入與錯誤狀態全部由 query 狀態推導。
  const query = createQuery(() =>
    departmentsQueryOptions({
      page: page(),
      pageSize: PAGE_SIZE,
      companyId: filter() || undefined,
    })
  );

  const total = () => Number(query.data?.pagination?.total ?? 0);

  // 公司下拉（累積式）：各頁由 `fetchNextPage` 依序疊上，「載入更多」不再自己累積 signal。
  const companyOptions = createInfiniteQuery(() =>
    companyDropdownQueryOptions({ keyword: companySearch() || undefined })
  );

  const companyTotal = () =>
    Number(companyOptions.data?.pages.at(-1)?.pagination?.total ?? 0);

  /**
   * 下拉選項＝補載的公司＋各頁攤平後**去重**（伺服器端分頁的頁界可能重疊，
   * 改寫前的手寫累積版也是這樣去重）。順序即選項順序，補載的公司排在最前面。
   */
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
  const [editing, setEditing] = createSignal<Department | null>(null);
  // 伺服器回應的錯誤訊息（元件層 signal，不是驗證狀態，也不對應任何欄位）。
  const [serverError, setServerError] = createSignal<string | undefined>();
  // 刪除失敗不屬於任何 query（不是清單資料），單獨保留一顆區域 signal。
  const [deleteError, setDeleteError] = createSignal<string | null>(null);

  /** 頁面層 banner：部門清單錯誤優先，其次是公司下拉的載入錯誤，最後是刪除失敗。 */
  const bannerError = () =>
    query.error
      ? errorMessage(query.error)
      : companyOptions.error
        ? errorMessage(companyOptions.error)
        : deleteError();

  // 清單重新查詢時清掉上一次刪除留下的錯誤 banner（與改寫前 `load()` 開頭 `setError(null)` 等價）。
  createEffect(() => {
    if (query.isFetching) setDeleteError(null);
  });

  const nameValidators = fieldValidators(departmentSchema.entries.name);
  const companyValidators = fieldValidators(departmentSchema.entries.company);

  /**
   * 送出篩選：把草稿套進 query key 並回第 1 頁。兩個 signal 必須放在同一個 `batch` 內——
   * 分開寫會先以「舊頁碼＋新篩選」查一次、再以「第 1 頁＋新篩選」查一次（兩次 RPC）。
   */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter(companyFilter());
      setPage(1);
    });
  };

  const form = createForm(() => ({
    // `canSubmitWhenInvalid` 的理由（F3）見 form-helpers.ts 的 `appFormOptions`。
    ...appFormOptions,
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
        // 清單資料已變更 → 前綴失效（清單 query 立即重取，不需手動 reload）。
        await client.invalidateQueries({ queryKey: ["departments"] });
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
    // 該部門的公司可能不在已載入的分頁內,先補載再開啟（補載成功才會有對應的 <option>）
    if (!companies().some((c) => c.id === d.companyId)) {
      try {
        const res = await companyClient.getCompany({ companyId: d.companyId });
        const c = res.company;
        if (c) {
          setPinnedCompanies((prev) =>
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
    setDeleteError(null);
    try {
      await departmentClient.deleteDepartment({ departmentId: d.id });
      // 刪除後清單重取；若刪到當前頁超界，超頁退回的 effect 會把頁碼夾回合法值。
      await client.invalidateQueries({ queryKey: ["departments"] });
    } catch (err) {
      setDeleteError(errorMessage(err));
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
        <div class="flex flex-col gap-3 border-b border-border bg-muted px-3 py-3">
          <form
            class="flex flex-wrap items-end gap-3"
            onSubmit={(e) => {
              e.preventDefault();
              // 換關鍵字＝換 query key（第 1 頁起算，累積的頁面自然重來）；
              // 以 `getCompany` 補載的公司不屬於這組關鍵字的結果，一併清掉。
              setCompanySearch(companyKeyword());
              setPinnedCompanies([]);
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
            <Button type="submit" variant="outline" loading={companyOptions.isFetching}>
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
                disabled={companyOptions.isFetchingNextPage || !companyOptions.hasNextPage}
                onClick={() => void companyOptions.fetchNextPage()}
              >
                載入更多
              </Button>
            </div>
          </form>

          <form class="flex flex-wrap items-end gap-3" onSubmit={submitFilter}>
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
            <Show when={query.isPending}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={4} class="py-8 text-center text-muted-foreground">
                  載入中…
                </TableCell>
              </TableRow>
            </Show>
            <Show when={!query.isPending && (query.data?.departments.length ?? 0) === 0}>
              <TableRow class="hover:bg-transparent">
                <TableCell colspan={4} class="py-8 text-center text-muted-foreground">
                  尚無部門資料
                </TableCell>
              </TableRow>
            </Show>
            <For each={query.data?.departments ?? []}>
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
          onPageChange={setPage}
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
