import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { Dialog } from "@ark-ui/solid/dialog";
import { Fieldset } from "@ark-ui/solid/fieldset";
import { createSignal, For, onMount, Show } from "solid-js";
import {
  CompanyService,
  DepartmentService,
  type Company,
  type Department,
} from "~/lib/proto/salesorder/v1/company_pb";
import { ListPagination } from "../components/ListPagination";

const PAGE_SIZE = 20;
const COMPANY_PAGE_SIZE = 50;

const departmentClient = createClient(
  DepartmentService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);
const companyClient = createClient(
  CompanyService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);

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

/** 部門主檔 CRUD 頁(/users/departments)。 */
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
  const [saving, setSaving] = createSignal(false);
  const [formError, setFormError] = createSignal<string | null>(null);

  // 表單欄位
  const [name, setName] = createSignal("");
  const [companyId, setCompanyId] = createSignal("");

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

  const openCreate = () => {
    setEditing(null);
    setName("");
    setCompanyId(companyFilter() || companies()[0]?.id || "");
    setFormError(null);
    setDialogOpen(true);
  };

  const openEdit = async (d: Department) => {
    setEditing(d);
    setName(d.name);
    setCompanyId(d.companyId);
    setFormError(null);
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
    setDialogOpen(true);
  };

  const submit = async (e: SubmitEvent) => {
    e.preventDefault();
    if (saving()) return;
    const trimmedName = name().trim();
    if (!trimmedName) {
      setFormError("請輸入部門名稱");
      return;
    }
    if (!companyId()) {
      setFormError("請選擇所屬公司");
      return;
    }
    setSaving(true);
    setFormError(null);
    try {
      const current = editing();
      if (current) {
        await departmentClient.updateDepartment({
          departmentId: current.id,
          name: trimmedName,
        });
      } else {
        await departmentClient.createDepartment({
          companyId: companyId(),
          name: trimmedName,
        });
      }
      setDialogOpen(false);
      await load();
    } catch (err) {
      setFormError(errorMessage(err));
    } finally {
      setSaving(false);
    }
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
    <main class="p-6">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">部門管理</h1>
          <p class="mt-1 text-sm text-gray-500">公司底下的業務單位(共 {total()} 筆)</p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          新增部門
        </button>
      </div>

      <form
        class="mb-4 flex flex-wrap items-end gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          loadCompanies(true);
        }}
      >
        <div>
          <label for="company-search-keyword" class="block text-sm font-medium text-gray-700">
            公司關鍵字
          </label>
          <input
            id="company-search-keyword"
            value={companyKeyword()}
            onInput={(e) => setCompanyKeyword(e.currentTarget.value)}
            placeholder="名稱 / 識別碼"
            class="mt-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          />
        </div>
        <button
          type="submit"
          disabled={companiesLoading()}
          class="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
        >
          搜尋公司
        </button>
        <div class="flex items-center gap-3 pb-1 text-sm text-gray-600">
          <span>
            已載入 {companies().length} 家,共 {companyTotal()} 家
          </span>
          <button
            type="button"
            onClick={() => loadCompanies(false)}
            disabled={companiesLoading() || companies().length >= companyTotal()}
            class="rounded-md border border-gray-300 px-3 py-1.5 text-xs text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {companiesLoading() ? "載入中…" : "載入更多"}
          </button>
        </div>
      </form>

      <form
        class="mb-4 flex flex-wrap items-end gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1); // 查詢變更時回到第一頁
          load();
        }}
      >
        <div>
          <label for="department-company-filter" class="block text-sm font-medium text-gray-700">
            所屬公司
          </label>
          <select
            id="department-company-filter"
            value={companyFilter()}
            onChange={(e) => setCompanyFilter(e.currentTarget.value)}
            class="mt-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          >
            <option value="">全部公司</option>
            <For each={companies()}>
              {(c) => <option value={c.id}>{c.name}</option>}
            </For>
          </select>
        </div>
        <button
          type="submit"
          class="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
        >
          查詢
        </button>
      </form>

      <Show when={error()}>
        <p class="mb-4 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
          {error()}
        </p>
      </Show>

      <div class="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                部門名稱
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                所屬公司
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                ID
              </th>
              <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500">
                操作
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <Show when={loading()}>
              <tr>
                <td colspan={4} class="px-4 py-8 text-center text-sm text-gray-500">
                  載入中…
                </td>
              </tr>
            </Show>
            <Show when={!loading() && departments().length === 0}>
              <tr>
                <td colspan={4} class="px-4 py-8 text-center text-sm text-gray-500">
                  尚無部門資料
                </td>
              </tr>
            </Show>
            <For each={departments()}>
              {(d) => (
                <tr class="hover:bg-gray-50">
                  <td class="px-4 py-3 text-sm font-medium text-gray-900">{d.name}</td>
                  <td class="px-4 py-3 text-sm text-gray-600">{d.companyName || "—"}</td>
                  <td class="px-4 py-3 text-sm text-gray-400">{d.id}</td>
                  <td class="px-4 py-3 text-right text-sm">
                    <button
                      type="button"
                      onClick={() => openEdit(d)}
                      class="text-blue-600 hover:text-blue-800"
                    >
                      編輯
                    </button>
                    <button
                      type="button"
                      onClick={() => remove(d)}
                      class="ml-3 text-red-600 hover:text-red-800"
                    >
                      刪除
                    </button>
                  </td>
                </tr>
              )}
            </For>
          </tbody>
        </table>
      </div>
      <ListPagination
        total={total()}
        pageSize={PAGE_SIZE}
        page={page()}
        onPageChange={goToPage}
      />

      <Dialog.Root open={dialogOpen()} onOpenChange={(e) => setDialogOpen(e.open)}>
        <Dialog.Backdrop class="fixed inset-0 z-40 bg-black/40" />
        <Dialog.Positioner class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <Dialog.Content class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <div class="flex items-start justify-between">
              <div>
                <Dialog.Title class="text-lg font-bold text-gray-900">
                  {editing() ? "編輯部門" : "新增部門"}
                </Dialog.Title>
                <Dialog.Description class="mt-1 text-sm text-gray-500">
                  {editing() ? "修改部門名稱" : "在指定公司下建立部門"}
                </Dialog.Description>
              </div>
              <Dialog.CloseTrigger
                class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                aria-label="關閉"
              >
                ×
              </Dialog.CloseTrigger>
            </div>

            <form onSubmit={submit}>
              <Fieldset.Root class="mt-4 space-y-4">
                <Fieldset.Legend class="sr-only">部門資料</Fieldset.Legend>
                <div>
                  <label
                    for="department-name"
                    class="block text-sm font-medium text-gray-700"
                  >
                    部門名稱 *
                  </label>
                  <input
                    id="department-name"
                    required
                    value={name()}
                    onInput={(e) => setName(e.currentTarget.value)}
                    class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                  />
                </div>
                <div>
                  <label
                    for="department-company"
                    class="block text-sm font-medium text-gray-700"
                  >
                    所屬公司 *
                  </label>
                  <select
                    id="department-company"
                    required
                    disabled={!!editing()}
                    value={companyId()}
                    onChange={(e) => setCompanyId(e.currentTarget.value)}
                    class="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none disabled:bg-gray-100 disabled:text-gray-400"
                  >
                    <option value="" disabled>
                      請選擇公司
                    </option>
                    <For each={companies()}>
                      {(c) => <option value={c.id}>{c.name}</option>}
                    </For>
                  </select>
                </div>

                <Show when={formError()}>
                  <p class="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
                    {formError()}
                  </p>
                </Show>

                <div class="flex justify-end gap-3 pt-2">
                  <Dialog.CloseTrigger class="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
                    取消
                  </Dialog.CloseTrigger>
                  <button
                    type="submit"
                    disabled={saving()}
                    class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                  >
                    {saving() ? "儲存中…" : "儲存"}
                  </button>
                </div>
              </Fieldset.Root>
            </form>
          </Dialog.Content>
        </Dialog.Positioner>
      </Dialog.Root>
    </main>
  );
}
