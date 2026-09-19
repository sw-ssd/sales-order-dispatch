import { Code, ConnectError } from "@connectrpc/connect";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  tableFeatures,
  type PaginationState,
} from "@tanstack/solid-table";
import { batch, createEffect, createSignal, For, Show } from "solid-js";
import type { Permission, Role } from "~/lib/proto/salesorder/v1/role_pb";
import {
  Badge,
  Button,
  Card,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui";
import { PermissionMatrix } from "../components/PermissionMatrix";
import { ListPagination } from "../components/ListPagination";
import { PAGE_SIZE, roleClient, rolesQueryOptions } from "../queries";

/**
 * 角色表格的 table 功能集：目前只有分頁（排序波次再加入 `rowSortingFeature`）。
 * features 必須是穩定的靜態值——每個元件都自己 `tableFeatures({...})` 會多一份無用的定義。
 */
const ROLE_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

/**
 * 欄位定義工具：features 已綁定，`accessor` 的值型別因此跟著功能集推導。
 * v9 的 table 是 headless 的——欄位定義只描述資料與渲染內容，實際的 `<th>`/`<td>`
 * 仍由 `~/components/ui` 的 `TableHead`/`TableCell` 負責（`ui/table.tsx` 未動）。
 */
const roleColumnHelper = createColumnHelper<typeof ROLE_TABLE_FEATURES, Role>();

/**
 * 沒有資料時的穩定空陣列：`data` 每次回傳新陣列會讓 table 的 row model 反覆重算。
 */
const NO_ROLES: Role[] = [];

const DATA_SCOPE_LABELS: Record<string, string> = {
  all: "全部",
  company: "公司",
  department: "部門",
  self: "本人",
};

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (Code[err.code]) {
      case "Unauthenticated":
        return "請先登入";
      case "PermissionDenied":
        return "無角色權限管理權限(僅 super / company_admin)";
      case "NotFound":
        return "角色不存在";
      case "InvalidArgument":
        return err.rawMessage || "請求參數錯誤";
    }
    return err.rawMessage || "請求失敗";
  }
  return "無法連線至伺服器,請確認後端服務已啟動";
}

/**
 * 角色權限設置頁(/users/roles;T19):角色清單 + 權限矩陣(resource × action)。
 * 版型照 Tailkit（Page Headings + 側欄卡片）：頁首標題區塊、左側角色表格（列內「選取」按鈕、
 * 選中列標示 `data-state="selected"`）、右側矩陣。內距由 AppShell 內容區負責。
 * 清單資料（角色列／總筆數／載入與錯誤狀態）一律來自 `../queries.ts` 的
 * `rolesQueryOptions`＋`createQuery`：頁面持有的只有「已選取角色」與 table 的 pagination state
 * （＝查詢輸入；`page` 不再是獨立 signal，見下方）。
 * 權限矩陣是受控編輯緩衝（80 格 checkbox 的草稿＋dirty），不是清單資料，
 * 仍以區域 signal 持有、由 `getRolePermissions` 讀取（載入時機與改寫前相同）。
 */
export default function RolesPage() {
  // table 的 pagination state（受控）：頁碼與每頁筆數的唯一真相，query key 一律取自這裡。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: PAGE_SIZE,
  });
  const [selectedId, setSelectedId] = createSignal<string | null>(null);
  const [permissions, setPermissions] = createSignal<Permission[]>([]);
  const [loadingPerms, setLoadingPerms] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [saving, setSaving] = createSignal(false);
  const [dirty, setDirty] = createSignal(false);
  const [savedAt, setSavedAt] = createSignal<string | null>(null);

  const client = useQueryClient();

  /**
   * 欄位定義（`roleColumnHelper` 已綁定 features）：角色代碼／名稱／系統角色／狀態／ID／操作。
   * 操作欄的「選取」是**列內按鈕**（不是整列可點）——鍵盤可達、有名稱，
   * 並以 `aria-pressed` 標示該列是否為目前已選取的角色。
   * 定義在元件內：列內按鈕要關到 `selectRole`/`selectedId`；Solid 元件只執行一次，陣列因此穩定。
   */
  const columns = roleColumnHelper.columns([
    roleColumnHelper.accessor("code", {
      header: "角色代碼",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    roleColumnHelper.accessor("name", {
      header: "名稱",
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    roleColumnHelper.accessor("isSystem", {
      header: "系統角色",
      cell: (info) => (
        <Badge variant={info.getValue() ? "info" : "outline"}>
          {info.getValue() ? "內建" : "自訂"}
        </Badge>
      ),
    }),
    roleColumnHelper.accessor("isActive", {
      header: "狀態",
      cell: (info) => (
        <Badge variant={info.getValue() ? "success" : "secondary"}>
          {info.getValue() ? "啟用" : "停用"}
        </Badge>
      ),
    }),
    roleColumnHelper.accessor("id", {
      header: "ID",
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    roleColumnHelper.display({
      id: "actions",
      // 表頭用內層 span 靠右：這樣就不必為單一欄位鋪 column meta 的樣式管線。
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="text-right">
          {/*
            `aria-label` 以角色 `name` 產生（「選取 系統管理員」）——多列都有「選取」時，
            螢幕閱讀器與測試都必須能指名道姓地選到某一列；可見文字「選取」包含在名稱內（WCAG 2.5.3）。
            條件：`roles.name`／`code` 在 DB 皆無唯一約束（00005_core_schema.sql 只 NOT NULL），
            所以可及名稱**只在資料沒有同名角色時**才唯一；若出現同名角色，會有多列共用同一個
            名稱。本頁測試亦以「角色名稱不重複」為前提（`getByRole("button", { name })` 撞名即失敗）。
          */}
          <Button
            type="button"
            variant="outline"
            size="sm"
            aria-label={`選取 ${info.row.original.name}`}
            aria-pressed={info.row.original.id === selectedId()}
            onClick={() => selectRole(info.row.original)}
          >
            選取
          </Button>
        </div>
      ),
    }),
  ]);

  // 清單唯一的資料來源：列資料、總筆數、載入與錯誤狀態全部由 query 狀態推導。
  // 頁碼與每頁筆數取自 `pagination`（= table 的 pagination state，見下方 `state`／
  // `onPaginationChange` 接線；`pageIndex` 0-based，query key 用 1-based）。
  const query = createQuery(() =>
    rolesQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
    })
  );

  const roles = () => query.data?.roles ?? [];
  const total = () => Number(query.data?.pagination?.total ?? 0);
  const selectedRole = () => roles().find((r) => r.id === selectedId()) ?? null;
  // 清單載入失敗由 query 狀態驅動；矩陣讀取／儲存失敗不屬於任何 query，走區域 signal。
  // 兩者共用一條 banner，清單錯誤優先（與改寫前的共用 `error` 同一個可見結果）。
  const banner = () => (query.error ? errorMessage(query.error) : error());

  /**
   * table 實例。分頁一律手動（`manualPagination: true`）：
   * - 資料永遠只有當前頁 → table 不得再依 `pageIndex` 切一次（否則第 2 頁會變 0 列）；
   * - 資料變更時也不得把頁碼打回第 1 頁（v9 的 `autoResetPageIndex` 預設會，manual 模式使
   *   它 `?? !manualPagination` 為 false → 不重置）。
   * 頁數來自後端的 `total`（`rowCount`），查詢結果與輸入都以 getter 暴露以維持反應性。
   *
   * 順序說明：state → query → table 是 v9 的必然順序——`createTable` 建構時就會把 options
   * （含 getter）解析一次，所以 `data` 來源（query）必須先存在；而 query 的頁碼又來自 table
   * 持有的 pagination state。兩者都以 `pagination` 這一個 signal 為唯一真相（table 由
   * `state`／`onPaginationChange` 綁定它，是全頁唯一的寫入者），因此不構成兩份頁碼。
   */
  const table = createTable({
    features: ROLE_TABLE_FEATURES,
    columns,
    get data() {
      return query.data?.roles ?? NO_ROLES;
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

  /**
   * 超頁退回 ＋ 自動選取**同一個 effect**（讀寫順序即語意）。兩者不可以拆成兩個 effect：
   * clamp 先 `setPageIndex` 之後，observer 的 `data` 還會落後一拍（實測：同一次 flush 內 clamp
   * effect 先跑並把頁碼改成第 1 頁，自動選取 effect 卻仍讀到「被放棄那一頁」的 data），
   * 於是拿舊資料判斷「頁碼是否超界」必為合法 → 會選了被放棄那一頁的第一筆，`selectedId`
   * 指向新頁清單中不存在的角色 → 面板永久停在「請選擇角色」、矩陣不再自動載入，且多打一次
   * `getRolePermissions`（BASE 不會）。合併後同一次讀取裡 `page` 與 `data` 才是一致的。
   *
   * 1. 超頁退回：回傳的 total 讓目前頁碼超界時（例：該頁資料被刪光），把頁碼夾到合法值。
   *    頁碼是 query key 的一部分 → `setPageIndex` 自己就會觸發重取，不必也不能再手動重載；
   *    夾到的頁碼必定 ≤ maxPage < 原頁碼（嚴格遞減、下界 1），所以重取次數有界。
   * 2. 自動選取：沒有選取（首屏、或剛換頁被清空）時，選取當前頁第一筆並載入其矩陣——
   *    與改寫前 `loadRoles` 的 `if (!selectedId() && list.length > 0)` 同義。
   *
   * 兩個守衛的理由不同：
   * - `!query.data`：首次載入／換頁失敗時沒有資料，`total()` 會算成 0；據以夾取會把使用者的
   *   頁碼靜默改寫並多打一次請求（「沒有資料」不等於「這一頁不存在」）。**這一半有牙**：
   *   `clamp 守衛` 測試移除它就會變紅。
   * - `query.isPlaceholderData`：placeholder 期間的 `data` 屬於前一個 key（`placeholderData`
   *   保留的舊結果）：據以退回會把剛切過去的頁碼彈回來、據以選取會把舊頁的角色留在面板上
   *   （新頁資料到了就會被當成「已有選取」而不再更正），故必須排除。**對 clamp 而言這半邊是
   *   防禦性的**——能寫入頁碼的來源只有以同一份 `total()` 產生的分頁 UI／Ark 自己的 clamp，
   *   placeholder 保留的正是剛離開那一頁的 total，`page > maxPage` 在 placeholder 期間構造不
   *   出來（3B 移交 table 後可達性沒有上升：manual 模式下 table 不自己動頁碼）；
   *   **但對選取而言它是有牙的**：本波突變實測（移除 `|| query.isPlaceholderData`）→
   *   `換頁清空已選角色`、`換頁：清空舊選取…` 兩條變紅（舊頁第一筆會被選回來）。
   */
  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / pagination().pageSize));
    if (pagination().pageIndex + 1 > maxPage) {
      table.setPageIndex(maxPage - 1);
      return;
    }
    const list = roles();
    if (selectedId() || list.length === 0) return;
    setSelectedId(list[0].id);
    void loadPermissions(list[0].id);
  });

  /**
   * 換頁（Ark 分頁 UI 的唯一入口）：清空選取（新頁的角色清單與舊頁無關）＋換頁碼，
   * 兩個寫入放在同一個 `batch`（兩者都是同一輪 flush 的輸入，同批寫入讓 effect 只看到最終狀態）。
   * 註：現行寫法即使沒有 batch 也不會把舊頁第一筆選回來——observer 已切到新 key
   * （`isPlaceholderData` 為 true），自動選取 effect 會被守衛擋下（複審實測 m8 全綠）。
   */
  const goToPage = (p: number) => {
    batch(() => {
      table.setPageIndex(p - 1);
      setSelectedId(null);
    });
  };

  const loadPermissions = async (roleId: string) => {
    setLoadingPerms(true);
    setError(null);
    setDirty(false);
    try {
      const res = await roleClient.getRolePermissions({ roleId });
      setPermissions(res.permissions);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoadingPerms(false);
    }
  };

  const selectRole = (r: Role) => {
    if (r.id === selectedId()) return;
    setSelectedId(r.id);
    void loadPermissions(r.id);
  };

  const changePermissions = (next: Permission[]) => {
    setPermissions(next);
    setDirty(true);
  };

  const save = async () => {
    const role = selectedRole();
    if (!role || saving()) return;
    setSaving(true);
    setError(null);
    try {
      await roleClient.updateRolePermissions({
        roleId: role.id,
        permissions: permissions(),
      });
      setDirty(false);
      setSavedAt(new Date().toLocaleTimeString());
      // 權限異動後失效 ability 快取(queryKey ["ability"]),守衛/Can 立即以新規則生效。
      // 用頁面所在的 client(provider 注入的單例)而非 import 單例:兩者在 app 是同一個實例,
      // 但測試注入的 client 才吃得到失效(與 CompaniesPage/DepartmentsPage 同一慣例)。
      void client.invalidateQueries({ queryKey: ["ability"] });
      await loadPermissions(role.id); // 回讀(sort_order 正規化後)
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <main>
      <header class="mb-6 border-b-2 border-border pb-4">
        <h1 class="text-2xl font-bold text-foreground">角色權限設置</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          管理角色功能權限(resource × action);內建角色權限為系統預設值
        </p>
      </header>

      <Show when={banner()}>
        <p
          class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
          role="alert"
        >
          {banner()}
        </p>
      </Show>

      {/*
        左欄從 240px 放寬為等分比例：角色清單從兩行式 `<ul>` 改成六欄表格（D7），
        240px 只能看到一欄半、得靠水平捲動才能選取；矩陣仍需要它的欄寬（80 格 checkbox）。
      */}
      <div class="grid gap-6 lg:grid-cols-2">
        <Card class="h-fit">
          {/* 表格的可及名稱：頁面上另一個表格是權限矩陣，兩者都必須能被指名（螢幕閱讀器與測試同理）。 */}
          <Table aria-label="角色清單">
            <TableHeader>
              <For each={table.getHeaderGroups()}>
                {(headerGroup) => (
                  <TableRow class="hover:bg-transparent">
                    <For each={headerGroup.headers}>
                      {(header) => (
                        <TableHead>
                          {flexRender(header.column.columnDef.header, header.getContext())}
                        </TableHead>
                      )}
                    </For>
                  </TableRow>
                )}
              </For>
            </TableHeader>
            <TableBody>
              <Show when={query.isPending}>
                <TableRow class="hover:bg-transparent">
                  <TableCell
                    colspan={table.getAllLeafColumns().length}
                    class="py-8 text-center text-muted-foreground"
                  >
                    載入中…
                  </TableCell>
                </TableRow>
              </Show>
              {/*
                空狀態必須排除 `isError`：換頁失敗時新 key 沒有 placeholder 結果
                （`data` 為 undefined、`isPlaceholderData` 為 false），只憑「非 pending 且 0 列」
                會把它當成空清單，與 `placeholderData` 「不閃空」的意圖相反。
              */}
              <Show when={!query.isPending && !query.isError && roles().length === 0}>
                <TableRow class="hover:bg-transparent">
                  <TableCell
                    colspan={table.getAllLeafColumns().length}
                    class="py-8 text-center text-muted-foreground"
                  >
                    尚無角色
                  </TableCell>
                </TableRow>
              </Show>
              {/*
                選取態的視覺沿用 `ui/table.tsx` 已為選取預留的 `TableRow data-state="selected"`
                （不另寫顏色字面值）；語意上的選取態由列內按鈕的 `aria-pressed` 承載。
              */}
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow
                    data-state={
                      row.original.id === selectedId() ? "selected" : undefined
                    }
                  >
                    <For each={row.getAllCells()}>
                      {(cell) => (
                        <TableCell>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      )}
                    </For>
                  </TableRow>
                )}
              </For>
            </TableBody>
          </Table>

          {/*
            Ark 只負責渲染：頁碼與總數都取自 table 實例（`pageIndex` 0-based → `page` 1-based），
            換頁也只回寫 table 的 pagination state（`goToPage` 另外清空選取）。
            `table.getRowCount()` 是全 repo 唯一的 rowCount 消費點；若改回直接讀 `total()`，
            表格的 `rowCount` 就沒有消費端，這層保護會靜默失效（實測全綠，無測試可抓）。
          */}
          <ListPagination
            total={table.getRowCount()}
            pageSize={table.atoms.pagination.get().pageSize}
            page={table.atoms.pagination.get().pageIndex + 1}
            onPageChange={goToPage}
          />
        </Card>

        <section class="min-w-0">
          <Show
            when={selectedRole()}
            fallback={<p class="text-sm text-muted-foreground">請選擇角色</p>}
          >
            {(role) => (
              <>
                <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <h2 class="text-lg font-bold text-foreground">{role().name}</h2>
                    <p class="mt-1 text-sm text-muted-foreground">
                      {role().code} · 資料範圍{" "}
                      {DATA_SCOPE_LABELS[role().dataScope] ?? role().dataScope}
                      <Show when={savedAt()}> · 已儲存 {savedAt()}</Show>
                    </p>
                  </div>
                  <Button
                    type="button"
                    onClick={save}
                    disabled={!dirty()}
                    loading={saving()}
                  >
                    儲存變更
                  </Button>
                </div>

                <Show when={loadingPerms()}>
                  <p class="py-8 text-center text-sm text-muted-foreground">
                    載入權限中…
                  </p>
                </Show>
                <Show when={!loadingPerms()}>
                  <PermissionMatrix
                    permissions={permissions()}
                    onChange={changePermissions}
                    isSystem={role().isSystem}
                  />
                </Show>
              </>
            )}
          </Show>
        </section>
      </div>
    </main>
  );
}
