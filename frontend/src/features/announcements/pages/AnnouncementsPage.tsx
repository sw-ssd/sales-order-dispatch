import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  tableFeatures,
  type PaginationState,
} from "@tanstack/solid-table";
import { createEffect, createSignal, For, Show, type JSX } from "solid-js";
import type { Announcement } from "~/lib/proto/salesorder/v1/announcement_pb";
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
import { toSortOrder } from "../../masters/schemas";
import { ListPagination } from "../../users/components/ListPagination";
import {
  ANNOUNCEMENTS_PAGE_SIZE,
  announcementClient,
  announcementsQueryOptions,
} from "../queries";
import {
  ANNOUNCEMENT_SCOPES,
  ANNOUNCEMENT_TYPES,
  ANNOUNCEMENT_TYPE_LABELS,
  announcementSchema,
  fromRFC3339,
  scopeTargets,
  toRFC3339,
} from "../schemas";
import { queryData } from "~/lib/query-data";

/** 公告表格的 table 功能集：**只有分頁**（`ListAnnouncementsRequest` 沒有 sort/desc）。 */
const ANNOUNCEMENT_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const announcementColumnHelper =
  createColumnHelper<typeof ANNOUNCEMENT_TABLE_FEATURES, Announcement>();

/** 沒有資料時的穩定空陣列（避免 row model 反覆重算）。 */
const NO_ANNOUNCEMENTS: Announcement[] = [];

/** 表單值型別（明確傳給 createForm,避免推斷在巢狀 Field 區塊退化成 unknown）。 */
type AnnouncementFormValues = {
  scope: (typeof ANNOUNCEMENT_SCOPES)[number];
  companyId: string;
  departmentId: string;
  type: (typeof ANNOUNCEMENT_TYPES)[number];
  title: string;
  content: string;
  imageUrl: string;
  linkUrl: string;
  publishAt: string;
  unpublishAt: string;
  sortOrder: string;
  isActive: boolean;
  deployWeb: boolean;
  deployApp: boolean;
};

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values,理由見 CompaniesPage 檔頭）。 */
const EMPTY_ANNOUNCEMENT_VALUES: AnnouncementFormValues = {
  scope: "auto" as (typeof ANNOUNCEMENT_SCOPES)[number],
  companyId: "",
  departmentId: "",
  type: "news" as (typeof ANNOUNCEMENT_TYPES)[number],
  title: "",
  content: "",
  imageUrl: "",
  linkUrl: "",
  publishAt: "", // 開對話框時由 openCreate 補「現在」
  unpublishAt: "",
  sortOrder: "0",
  isActive: true,
  deployWeb: true,
  deployApp: true,
};

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.InvalidArgument:
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
      case Code.PermissionDenied:
        return "沒有權限執行此操作（公告範圍須在你的管理範圍內）";
      case Code.Unauthenticated:
        return "請先登入";
      case Code.Unavailable:
        return "無法連線至伺服器,請確認後端服務已啟動";
      default:
        return err.rawMessage || "操作失敗,請稍後再試";
    }
  }
  return "無法連線至伺服器,請確認後端服務已啟動";
}

/**
 * 公告管理頁(/announcements)。
 *
 * 版型同 `WarehousesPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 三處契約：
 * - **總數在 `total`**（`ListAnnouncementsResponse.total`,int32）。
 * - **無排序**（`ListAnnouncementsRequest` 沒有 `sort`/`desc` → 表頭一律非可排序）。
 * - **範圍於建立時決定**（Update 不可改範圍,見 proto 註）:建立時以「自動歸屬/全系統/
 *   公司/部門」選擇,非 super 的「自動」由後端補自己的公司/部門;編輯時範圍唯讀。
 * - 軟刪除**無復原**（spec 只訂軟刪除,後端無 Restore）→ 不提供「含已刪除」開關。
 */
export default function AnnouncementsPage() {
  const [typeFilter, setTypeFilter] = createSignal("");
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: ANNOUNCEMENTS_PAGE_SIZE,
  });

  const client = useQueryClient();

  const scopeText = (a: Announcement) => {
    if (!a.companyId) return "全系統";
    return a.departmentId ? `部門 #${a.departmentId}` : `公司 #${a.companyId}`;
  };

  const columns = announcementColumnHelper.columns([
    announcementColumnHelper.accessor("title", {
      header: "標題",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    announcementColumnHelper.accessor("type", {
      header: "類型",
      cell: (info) => {
        const v = info.getValue() as (typeof ANNOUNCEMENT_TYPES)[number];
        return <Badge variant="outline">{ANNOUNCEMENT_TYPE_LABELS[v] ?? v}</Badge>;
      },
    }),
    announcementColumnHelper.display({
      id: "scope",
      header: "範圍",
      cell: (info) => <span class="text-muted-foreground">{scopeText(info.row.original)}</span>,
    }),
    announcementColumnHelper.display({
      id: "deploy",
      header: "投放",
      cell: (info) => {
        const a = info.row.original;
        return (
          <span class="flex gap-1">
            {a.deployWeb ? <Badge variant="secondary">Web</Badge> : null}
            {a.deployApp ? <Badge variant="secondary">App</Badge> : null}
            {!a.deployWeb && !a.deployApp ? (
              <span class="text-muted-foreground">—</span>
            ) : null}
          </span>
        );
      },
    }),
    announcementColumnHelper.accessor("publishAt", {
      header: "上架",
      cell: (info) => (
        <span class="text-muted-foreground">
          {info.getValue() ? new Date(info.getValue()).toLocaleString() : "—"}
        </span>
      ),
    }),
    announcementColumnHelper.accessor("unpublishAt", {
      header: "下架",
      cell: (info) => (
        <span class="text-muted-foreground">
          {info.getValue() ? new Date(info.getValue()).toLocaleString() : "手動"}
        </span>
      ),
    }),
    announcementColumnHelper.accessor("isActive", {
      header: "啟用",
      cell: (info) =>
        info.getValue() ? (
          <Badge variant="success">啟用</Badge>
        ) : (
          <Badge variant="secondary">停用</Badge>
        ),
    }),
    announcementColumnHelper.accessor("sortOrder", {
      header: "排序",
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    announcementColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => {
        const a = info.row.original;
        return (
          <div class="flex justify-end gap-3 whitespace-nowrap">
            <button
              type="button"
              onClick={() => openEdit(a)}
              class="font-medium text-primary hover:underline"
            >
              編輯
            </button>
            <button
              type="button"
              onClick={() => void remove(a)}
              class="font-medium text-destructive hover:underline"
            >
              刪除
            </button>
          </div>
        );
      },
    }),
  ]);

  const query = createQuery(() =>
    announcementsQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      type: typeFilter() || undefined,
    })
  );

  const total = () => Number(queryData(query, (d) => d?.total) ?? 0);

  const table = createTable({
    features: ANNOUNCEMENT_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.announcements) ?? NO_ANNOUNCEMENTS;
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

  /** 超頁退回：total 讓目前頁碼超界時把頁碼夾到合法值（`page` 在 key 內 → 夾了就重取）。 */
  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / pagination().pageSize));
    if (pagination().pageIndex + 1 > maxPage) table.setPageIndex(maxPage - 1);
  });

  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Announcement | null>(null);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const titleValidators = fieldValidators(announcementSchema.entries.title);

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_ANNOUNCEMENT_VALUES },
    onSubmit: async ({ value }) => {
      try {
        const current = editing();
        if (current) {
          // 範圍不可改（proto 註）→ Update 不帶範圍欄位。
          await announcementClient.updateAnnouncement({
            id: current.id,
            type: value.type,
            title: value.title.trim(),
            content: value.content,
            imageUrl: value.imageUrl.trim(),
            linkUrl: value.linkUrl.trim(),
            publishAt: toRFC3339(value.publishAt),
            unpublishAt: toRFC3339(value.unpublishAt),
            sortOrder: toSortOrder(value.sortOrder),
            isActive: value.isActive,
            deployWeb: value.deployWeb,
            deployApp: value.deployApp,
          });
        } else {
          const target = scopeTargets(value.scope, value.companyId, value.departmentId);
          await announcementClient.createAnnouncement({
            companyId: target.companyId,
            departmentId: target.departmentId,
            type: value.type,
            title: value.title.trim(),
            content: value.content,
            imageUrl: value.imageUrl.trim(),
            linkUrl: value.linkUrl.trim(),
            publishAt: toRFC3339(value.publishAt),
            unpublishAt: toRFC3339(value.unpublishAt),
            sortOrder: toSortOrder(value.sortOrder),
            isActive: value.isActive,
            deployWeb: value.deployWeb,
            deployApp: value.deployApp,
          });
        }
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["announcements"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);
  /** 目前表單的範圍選擇（決定公司/部門 ID 欄是否呈現;僅建立模式可見）。 */
  const scopeValue = form.useSelector((state) => state.values.scope);

  const openCreate = () => {
    setEditing(null);
    setServerError(undefined);
    // 上架時間預設「現在」（EMPTY 的值在模組載入時算好,開對話框時重取）。
    form.reset({
      ...EMPTY_ANNOUNCEMENT_VALUES,
      publishAt: fromRFC3339(new Date().toISOString()),
    });
    setDialogOpen(true);
  };

  /** 編輯：清單已帶全部欄位,直接以該列值開表單（範圍唯讀,見檔頭契約）。 */
  const openEdit = (a: Announcement) => {
    setEditing(a);
    setServerError(undefined);
    form.reset({
      scope: "auto",
      companyId: a.companyId,
      departmentId: a.departmentId,
      type: ANNOUNCEMENT_TYPES.includes(a.type as (typeof ANNOUNCEMENT_TYPES)[number])
        ? (a.type as (typeof ANNOUNCEMENT_TYPES)[number])
        : "news",
      title: a.title,
      content: a.content,
      imageUrl: a.imageUrl,
      linkUrl: a.linkUrl,
      publishAt: fromRFC3339(a.publishAt),
      unpublishAt: fromRFC3339(a.unpublishAt),
      sortOrder: String(a.sortOrder),
      isActive: a.isActive,
      deployWeb: a.deployWeb,
      deployApp: a.deployApp,
    });
    setDialogOpen(true);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const remove = async (a: Announcement) => {
    if (!window.confirm(`確定刪除公告「${a.title}」？（軟刪除,前台將不再顯示）`)) return;
    setActionError(null);
    try {
      await announcementClient.deleteAnnouncement({ id: a.id });
      await client.invalidateQueries({ queryKey: ["announcements"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">公告管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            輪播／最新消息／圖文文章,投放 Web 中台與 App 首頁（共 {total()} 筆）
          </p>
        </div>
        <Button type="button" onClick={openCreate}>
          新增公告
        </Button>
      </header>

      <Show when={query.error ? errorMessage(query.error) : actionError()}>
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
        <div class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3">
          <Field class="w-full sm:w-48">
            <FieldLabel for="announcement-type">類型</FieldLabel>
            <select
              id="announcement-type"
              class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
              value={typeFilter()}
              onChange={(e) => {
                setTypeFilter(e.currentTarget.value);
                table.setPageIndex(0);
              }}
            >
              <option value="">全部</option>
              <For each={[...ANNOUNCEMENT_TYPES]}>
                {(t) => <option value={t}>{ANNOUNCEMENT_TYPE_LABELS[t]}</option>}
              </For>
            </select>
          </Field>
        </div>

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
              when={queryData(query, (d) => d?.announcements?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={9}>
                    {query.isFetching ? "載入中…" : "尚無公告"}
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
            <DialogTitle>{editing() ? "編輯公告" : "新增公告"}</DialogTitle>
            <DialogDescription>
              {editing()
                ? "內容可改；發佈範圍於建立時決定,不可在此變更"
                : "選擇型別與投放平台；上下架時間窗決定前台可見期間"}
            </DialogDescription>
          </DialogHeader>

          <Show when={serverError()}>
            {(message) => (
              <p
                class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive"
                role="alert"
              >
                {message()}
              </p>
            )}
          </Show>

          <form class="space-y-4" onSubmit={submit}>
            <Show
              when={!editing()}
              fallback={
                <Field>
                  <FieldLabel>發佈範圍</FieldLabel>
                  <p class="text-sm text-muted-foreground">
                    {editing()
                      ? scopeText(editing() as Announcement)
                      : "—"}
                  </p>
                </Field>
              }
            >
              <div class="grid gap-4 sm:grid-cols-3">
                <form.Field name="scope">
                  {(field) => (
                    <Field>
                      <FieldLabel for="announcement-form-scope">範圍</FieldLabel>
                      <select
                        id="announcement-form-scope"
                        class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                        value={field().state.value}
                        onChange={(e) =>
                          field().handleChange(
                            e.currentTarget.value as (typeof ANNOUNCEMENT_SCOPES)[number]
                          )
                        }
                      >
                        <option value="auto">自動歸屬（依身分）</option>
                        <option value="system">全系統</option>
                        <option value="company">公司層</option>
                        <option value="department">部門層</option>
                      </select>
                    </Field>
                  )}
                </form.Field>
                {scopeValue() === "company" || scopeValue() === "department" ? (
                  <>
                    <form.Field name="companyId">
                      {(field) => (
                        <Field>
                          <FieldLabel for="announcement-form-company">公司 ID</FieldLabel>
                          <Input
                            id="announcement-form-company"
                            value={field().state.value}
                            onInput={(e) => field().handleChange(e.currentTarget.value)}
                            placeholder="super 指定公司層/部門層時填"
                          />
                        </Field>
                      )}
                    </form.Field>
                    <form.Field name="departmentId">
                      {(field) => (
                        <Field>
                          <FieldLabel for="announcement-form-dept">部門 ID</FieldLabel>
                          <Input
                            id="announcement-form-dept"
                            value={field().state.value}
                            onInput={(e) => field().handleChange(e.currentTarget.value)}
                            placeholder="部門層時填"
                          />
                        </Field>
                      )}
                    </form.Field>
                  </>
                ) : null}
              </div>
            </Show>

            <div class="grid gap-4 sm:grid-cols-2">
              <form.Field name="type">
                {(field) => (
                  <Field>
                    <FieldLabel for="announcement-form-type">型別</FieldLabel>
                    <select
                      id="announcement-form-type"
                      class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                      value={field().state.value}
                      onChange={(e) =>
                        field().handleChange(
                          e.currentTarget.value as (typeof ANNOUNCEMENT_TYPES)[number]
                        )
                      }
                    >
                      <For each={[...ANNOUNCEMENT_TYPES]}>
                        {(t) => <option value={t}>{ANNOUNCEMENT_TYPE_LABELS[t]}</option>}
                      </For>
                    </select>
                  </Field>
                )}
              </form.Field>

              <form.Field name="title" validators={titleValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="announcement-form-title">標題 *</FieldLabel>
                    <Input
                      id="announcement-form-title"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="週年慶活動開跑"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="publishAt">
                {(field) => (
                  <Field>
                    <FieldLabel for="announcement-form-publish">上架時間</FieldLabel>
                    <Input
                      id="announcement-form-publish"
                      type="datetime-local"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                    <p class="text-xs text-muted-foreground">留空 = 立即上架</p>
                  </Field>
                )}
              </form.Field>

              <form.Field name="unpublishAt">
                {(field) => (
                  <Field>
                    <FieldLabel for="announcement-form-unpublish">下架時間</FieldLabel>
                    <Input
                      id="announcement-form-unpublish"
                      type="datetime-local"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                    <p class="text-xs text-muted-foreground">留空 = 不自動下架</p>
                  </Field>
                )}
              </form.Field>

              <form.Field name="imageUrl">
                {(field) => (
                  <Field>
                    <FieldLabel for="announcement-form-image">圖片 URL（選填）</FieldLabel>
                    <Input
                      id="announcement-form-image"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="/uploads/…（banner 建議圖）"
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="linkUrl">
                {(field) => (
                  <Field>
                    <FieldLabel for="announcement-form-link">連結 URL（選填）</FieldLabel>
                    <Input
                      id="announcement-form-link"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="點擊 banner 的導向"
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="sortOrder">
                {(field) => (
                  <Field>
                    <FieldLabel for="announcement-form-sort">排序</FieldLabel>
                    <Input
                      id="announcement-form-sort"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="0"
                    />
                    <p class="text-xs text-muted-foreground">同型別依此升冪；非數字歸 0</p>
                  </Field>
                )}
              </form.Field>
            </div>

            <form.Field name="content">
              {(field) => (
                <Field>
                  <FieldLabel for="announcement-form-content">內容</FieldLabel>
                  <textarea
                    id="announcement-form-content"
                    class="flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={field().state.value}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                    placeholder="公告內文（article 類型以此呈現圖文文字）"
                  />
                </Field>
              )}
            </form.Field>

            <div class="flex flex-wrap gap-4">
              <form.Field name="isActive">
                {(field) => (
                  <label class="flex items-center gap-2 text-sm text-foreground">
                    <input
                      type="checkbox"
                      checked={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.checked)}
                    />
                    啟用
                  </label>
                )}
              </form.Field>
              <form.Field name="deployWeb">
                {(field) => (
                  <label class="flex items-center gap-2 text-sm text-foreground">
                    <input
                      type="checkbox"
                      checked={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.checked)}
                    />
                    投放 Web
                  </label>
                )}
              </form.Field>
              <form.Field name="deployApp">
                {(field) => (
                  <label class="flex items-center gap-2 text-sm text-foreground">
                    <input
                      type="checkbox"
                      checked={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.checked)}
                    />
                    投放 App
                  </label>
                )}
              </form.Field>
            </div>

            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>取消</DialogClose>
              <Button type="submit" loading={isSubmitting()}>
                {editing() ? "儲存" : "建立"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </main>
  );
}
