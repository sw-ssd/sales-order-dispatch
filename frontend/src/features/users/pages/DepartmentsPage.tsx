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
import { Button, buttonVariants } from "~/components/ui/button";
import { Card } from "~/components/ui/card";
import { Field, FieldLabel } from "~/components/ui/field";
import { Input } from "~/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
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

/**
 * 原生 select 的視覺（ui/ 沒有 select 元件）。`pr-10` 與顏色覆蓋的理由同 `CompaniesPage.tsx`。
 */
const SELECT_CLASS =
  "block w-full rounded-lg border border-border bg-card py-2 pr-10 pl-3 text-sm text-foreground focus:border-primary focus:ring-3 focus:ring-primary/50 focus:outline-none disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground";

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
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">部門管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            公司底下的業務單位(共 {total()} 筆)
          </p>
        </div>
        <Button type="button" onClick={openCreate}>
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

          <form class="space-y-4" onSubmit={submit}>
            <Field>
              <FieldLabel for="department-name">部門名稱 *</FieldLabel>
              <Input
                id="department-name"
                required
                value={name()}
                onInput={(e) => setName(e.currentTarget.value)}
              />
            </Field>
            <Field>
              <FieldLabel for="department-company">所屬公司 *</FieldLabel>
              <select
                id="department-company"
                required
                disabled={!!editing()}
                value={companyId()}
                onChange={(e) => setCompanyId(e.currentTarget.value)}
                class={SELECT_CLASS}
              >
                <option value="" disabled>
                  請選擇公司
                </option>
                <For each={companies()}>
                  {(c) => <option value={c.id}>{c.name}</option>}
                </For>
              </select>
            </Field>

            <Show when={formError()}>
              <p
                class="rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
                role="alert"
              >
                {formError()}
              </p>
            </Show>

            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>
                取消
              </DialogClose>
              <Button type="submit" loading={saving()}>
                儲存
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </main>
  );
}
