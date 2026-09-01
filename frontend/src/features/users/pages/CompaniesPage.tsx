import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Field, FieldLabel } from "~/components/ui/field";
import { createSignal, For, onMount, Show } from "solid-js";
import {
  CompanyService,
  type Company,
} from "~/lib/proto/salesorder/v1/company_pb";
import { cn } from "~/lib/cn";
import { ListPagination } from "../components/ListPagination";

const PAGE_SIZE = 20;

const companyClient = createClient(
  CompanyService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);

const STATUS_LABELS: Record<string, string> = {
  active: "啟用",
  inactive: "停用",
  suspended: "暫停",
};

function statusBadgeClass(status: string): string {
  switch (status) {
    case "active":
      return "bg-green-100 text-green-700";
    case "inactive":
      return "bg-gray-100 text-gray-600";
    case "suspended":
      return "bg-amber-100 text-amber-700";
    default:
      return "bg-gray-100 text-gray-600";
  }
}

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

/** 公司主檔 CRUD 頁(/users/companies)。 */
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
  const [saving, setSaving] = createSignal(false);
  const [formError, setFormError] = createSignal<string | null>(null);

  // 表單欄位
  const [name, setName] = createSignal("");
  const [taxId, setTaxId] = createSignal("");
  const [identifier, setIdentifier] = createSignal("");
  const [status, setStatus] = createSignal("active");

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

  const openCreate = () => {
    setEditing(null);
    setName("");
    setTaxId("");
    setIdentifier("");
    setStatus("active");
    setFormError(null);
    setDialogOpen(true);
  };

  const openEdit = (c: Company) => {
    setEditing(c);
    setName(c.name);
    setTaxId(c.taxId);
    setIdentifier(c.identifier);
    setStatus(c.status || "active");
    setFormError(null);
    setDialogOpen(true);
  };

  const submit = async (e: SubmitEvent) => {
    e.preventDefault();
    if (saving()) return;
    const trimmedName = name().trim();
    const trimmedIdentifier = identifier().trim();
    if (!trimmedName) {
      setFormError("請輸入公司名稱");
      return;
    }
    const current = editing();
    if (!current && !trimmedIdentifier) {
      setFormError("請輸入識別碼(identifier)");
      return;
    }
    setSaving(true);
    setFormError(null);
    try {
      if (current) {
        await companyClient.updateCompany({
          companyId: current.id,
          name: trimmedName,
          taxId: taxId().trim(),
          status: status(),
        });
      } else {
        await companyClient.createCompany({
          name: trimmedName,
          taxId: taxId().trim(),
          identifier: trimmedIdentifier,
          status: status(),
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
    <main class="p-6">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">公司管理</h1>
          <p class="mt-1 text-sm text-gray-500">多租戶公司主檔(共 {total()} 筆)</p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          新增公司
        </button>
      </div>

      <form
        class="mb-4 flex flex-wrap items-end gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1); // 查詢變更時回到第一頁
          load();
        }}
      >
        <div>
          <label for="company-keyword" class="block text-sm font-medium text-gray-700">
            關鍵字
          </label>
          <input
            id="company-keyword"
            value={keyword()}
            onInput={(e) => setKeyword(e.currentTarget.value)}
            placeholder="名稱 / 識別碼"
            class="mt-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          />
        </div>
        <div>
          <label for="company-status-filter" class="block text-sm font-medium text-gray-700">
            狀態
          </label>
          <select
            id="company-status-filter"
            value={statusFilter()}
            onChange={(e) => setStatusFilter(e.currentTarget.value)}
            class="mt-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          >
            <option value="">全部</option>
            <option value="active">啟用</option>
            <option value="inactive">停用</option>
            <option value="suspended">暫停</option>
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
                名稱
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                識別碼
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                統一編號
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                狀態
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
                <td colspan={6} class="px-4 py-8 text-center text-sm text-gray-500">
                  載入中…
                </td>
              </tr>
            </Show>
            <Show when={!loading() && companies().length === 0}>
              <tr>
                <td colspan={6} class="px-4 py-8 text-center text-sm text-gray-500">
                  尚無公司資料
                </td>
              </tr>
            </Show>
            <For each={companies()}>
              {(c) => (
                <tr class="hover:bg-gray-50">
                  <td class="px-4 py-3 text-sm font-medium text-gray-900">{c.name}</td>
                  <td class="px-4 py-3 text-sm text-gray-600">{c.identifier}</td>
                  <td class="px-4 py-3 text-sm text-gray-600">{c.taxId || "—"}</td>
                  <td class="px-4 py-3 text-sm">
                    <span
                      class={cn(
                        "rounded-full px-2 py-0.5 text-xs font-medium",
                        statusBadgeClass(c.status),
                      )}
                    >
                      {STATUS_LABELS[c.status] ?? c.status}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-sm text-gray-400">{c.id}</td>
                  <td class="px-4 py-3 text-right text-sm">
                    <button
                      type="button"
                      onClick={() => openEdit(c)}
                      class="text-blue-600 hover:text-blue-800"
                    >
                      編輯
                    </button>
                    <button
                      type="button"
                      onClick={() => remove(c)}
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

      <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing() ? "編輯公司" : "新增公司"}</DialogTitle>
            <DialogDescription>
              {editing() ? "修改名稱、統一編號或狀態" : "建立新的公司主檔"}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={submit}>
            <div class="mt-4 space-y-4">
              <Field>
                <FieldLabel for="company-name">公司名稱 *</FieldLabel>
                <input
                  id="company-name"
                  required
                  value={name()}
                  onInput={(e) => setName(e.currentTarget.value)}
                  class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                />
              </Field>
              <Field>
                <FieldLabel for="company-identifier">識別碼(identifier) *</FieldLabel>
                <input
                  id="company-identifier"
                  required
                  disabled={!!editing()}
                  value={identifier()}
                  onInput={(e) => setIdentifier(e.currentTarget.value)}
                  placeholder="建立後不可修改"
                  class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none disabled:bg-gray-100 disabled:text-gray-400"
                />
              </Field>
              <Field>
                <FieldLabel for="company-tax-id">統一編號</FieldLabel>
                <input
                  id="company-tax-id"
                  value={taxId()}
                  onInput={(e) => setTaxId(e.currentTarget.value)}
                  class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                />
              </Field>
              <Field>
                <FieldLabel for="company-status">狀態</FieldLabel>
                <select
                  id="company-status"
                  value={status()}
                  onChange={(e) => setStatus(e.currentTarget.value)}
                  class="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                >
                  <option value="active">啟用</option>
                  <option value="inactive">停用</option>
                  <option value="suspended">暫停</option>
                </select>
              </Field>

              <Show when={formError()}>
                <p class="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
                  {formError()}
                </p>
              </Show>

              <DialogFooter>
                <DialogClose class="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
                  取消
                </DialogClose>
                <button
                  type="submit"
                  disabled={saving()}
                  class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  {saving() ? "儲存中…" : "儲存"}
                </button>
              </DialogFooter>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </main>
  );
}
