import { Code, ConnectError } from "@connectrpc/connect";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import { batch, createEffect, createSignal, For, Show } from "solid-js";
import type { Permission, Role } from "~/lib/proto/salesorder/v1/role_pb";
import { Button, Card } from "~/components/ui";
import { cn } from "@/lib/cn";
import { PermissionMatrix } from "../components/PermissionMatrix";
import { ListPagination } from "../components/ListPagination";
import { PAGE_SIZE, roleClient, rolesQueryOptions } from "../queries";

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
 * 版型照 Tailkit（Page Headings + 側欄卡片）：頁首標題區塊、左側角色卡片（選中列用 `bg-primary/10`）、
 * 右側矩陣。內距由 AppShell 內容區負責。
 * 清單資料（角色列／總筆數／載入與錯誤狀態）一律來自 `../queries.ts` 的
 * `rolesQueryOptions`＋`createQuery`：頁面持有的只有查詢輸入（頁碼）與「已選取角色」。
 * 權限矩陣是受控編輯緩衝（80 格 checkbox 的草稿＋dirty），不是清單資料，
 * 仍以區域 signal 持有、由 `getRolePermissions` 讀取（載入時機與改寫前相同）。
 */
export default function RolesPage() {
  const [page, setPage] = createSignal(1);
  const [selectedId, setSelectedId] = createSignal<string | null>(null);
  const [permissions, setPermissions] = createSignal<Permission[]>([]);
  const [loadingPerms, setLoadingPerms] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [saving, setSaving] = createSignal(false);
  const [dirty, setDirty] = createSignal(false);
  const [savedAt, setSavedAt] = createSignal<string | null>(null);

  const client = useQueryClient();

  // 清單唯一的資料來源：列資料、總筆數、載入與錯誤狀態全部由 query 狀態推導。
  const query = createQuery(() =>
    rolesQueryOptions({ page: page(), pageSize: PAGE_SIZE })
  );

  const roles = () => query.data?.roles ?? [];
  const total = () => Number(query.data?.pagination?.total ?? 0);
  const selectedRole = () => roles().find((r) => r.id === selectedId()) ?? null;
  // 清單載入失敗由 query 狀態驅動；矩陣讀取／儲存失敗不屬於任何 query，走區域 signal。
  // 兩者共用一條 banner，清單錯誤優先（與改寫前的共用 `error` 同一個可見結果）。
  const banner = () => (query.error ? errorMessage(query.error) : error());

  /**
   * 超頁退回 ＋ 自動選取**同一個 effect**（讀寫順序即語意）。兩者不可以拆成兩個 effect：
   * clamp 先 `setPage` 之後，observer 的 `data` 還會落後一拍（實測：同一次 flush 內 clamp
   * effect 先跑並把 `page` 改成 1，自動選取 effect 卻仍讀到「被放棄那一頁」的 data），
   * 於是拿舊資料判斷「頁碼是否超界」必為合法 → 會選了被放棄那一頁的第一筆，`selectedId`
   * 指向新頁清單中不存在的角色 → 面板永久停在「請選擇角色」、矩陣不再自動載入，且多打一次
   * `getRolePermissions`（BASE 不會）。合併後同一次讀取裡 `page` 與 `data` 才是一致的。
   *
   * 1. 超頁退回：回傳的 total 讓目前頁碼超界時（例：該頁資料被刪光），把頁碼夾到合法值。
   *    `page` 是 query key 的一部分 → `setPage` 自己就會觸發重取，不必也不能再手動重載；
   *    夾到的頁碼必定 ≤ maxPage < 原頁碼（嚴格遞減、下界 1），所以重取次數有界。
   * 2. 自動選取：沒有選取（首屏、或剛換頁被清空）時，選取當前頁第一筆並載入其矩陣——
   *    與改寫前 `loadRoles` 的 `if (!selectedId() && list.length > 0)` 同義。
   *
   * `isPlaceholderData` 期間的 `data` 屬於前一個 key（placeholderData 保留的舊結果）：據以退回
   * 會把剛切過去的頁碼彈回來、據以選取會把舊頁的角色留在面板上（新頁資料到了就會被當成
   * 「已有選取」而不再更正），故必須排除；這不會漏掉任何一步——新資料一到，`data` 與
   * `isPlaceholderData` 都變動，這個 effect 會再跑一次。
   *
   * 誠實標註：**對 clamp 而言**這個守衛目前是防禦性的（UI 上不可達）——頁碼唯一的來源是以同一份
   * total 產生的分頁 UI，placeholder 保留的正是剛離開那一頁的 total，`page > maxPage` 在
   * placeholder 期間不會成立；**對選取而言它是有牙的**（`超頁退回`、`換頁清空已選角色` 兩條
   * 測試在移除後會變紅）。3B 把 `page` 移交 table 的 pagination state 後可達性會上升。
   */
  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / PAGE_SIZE));
    if (page() > maxPage) {
      setPage(maxPage);
      return;
    }
    const list = roles();
    if (selectedId() || list.length === 0) return;
    setSelectedId(list[0].id);
    void loadPermissions(list[0].id);
  });

  /**
   * 換頁：清空選取（新頁的角色清單與舊頁無關）＋換頁碼，兩個 signal 放在同一個 `batch`。
   * 這裡是**順序無關性**的防禦：兩者都是同一輪 flush 的輸入，同批寫入讓 effect 只看到最終狀態。
   * 註：現行寫法（`setPage` 先）即使沒有 batch 也不會把舊頁第一筆選回來——observer 已切到
   * 新 key（`isPlaceholderData` 為 true），自動選取 effect 會被守衛擋下（複審實測 m8 全綠）。
   */
  const goToPage = (p: number) => {
    batch(() => {
      setPage(p);
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

      <div class="grid gap-6 lg:grid-cols-[240px_1fr]">
        <Card class="h-fit">
          <Show when={query.isPending}>
            <p class="px-4 py-8 text-center text-sm text-muted-foreground">載入中…</p>
          </Show>
          {/*
            空狀態必須排除 `isError`：換頁失敗時新 key 沒有 placeholder 結果
            （`data` 為 undefined、`isPlaceholderData` 為 false），只憑「非 pending 且 0 列」
            會把它當成空清單，與 `placeholderData` 「不閃空」的意圖相反。
          */}
          <Show when={!query.isPending && !query.isError && roles().length === 0}>
            <p class="px-4 py-8 text-center text-sm text-muted-foreground">尚無角色</p>
          </Show>
          <ul class="divide-y divide-border">
            <For each={roles()}>
              {(r) => (
                <li>
                  <button
                    type="button"
                    onClick={() => selectRole(r)}
                    class={cn(
                      "w-full px-4 py-3 text-left hover:bg-muted",
                      r.id === selectedId() && "bg-primary/10",
                    )}
                  >
                    <span
                      class={cn(
                        "block text-sm font-medium",
                        r.id === selectedId() ? "text-primary" : "text-foreground",
                      )}
                    >
                      {r.name}
                    </span>
                    <span class="block text-xs text-muted-foreground">
                      {r.code} · {DATA_SCOPE_LABELS[r.dataScope] ?? r.dataScope}
                      {r.isSystem ? " · 內建" : ""}
                      {!r.isActive ? " · 停用" : ""}
                    </span>
                  </button>
                </li>
              )}
            </For>
          </ul>
          <ListPagination
            total={total()}
            pageSize={PAGE_SIZE}
            page={page()}
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
