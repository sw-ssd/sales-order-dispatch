import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createSignal, For, onMount, Show } from "solid-js";
import {
  RoleService,
  type Permission,
  type Role,
} from "~/lib/proto/salesorder/v1/role_pb";
import { transport } from "~/lib/transport";
import { Button, Card } from "~/components/ui";
import { cn } from "@/lib/cn";
import { queryClient } from "~/lib/query-client";
import { PermissionMatrix } from "../components/PermissionMatrix";
import { ListPagination } from "../components/ListPagination";

const roleClient = createClient(RoleService, transport);

const PAGE_SIZE = 20;

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
 */
export default function RolesPage() {
  const [roles, setRoles] = createSignal<Role[]>([]);
  const [total, setTotal] = createSignal(0);
  const [page, setPage] = createSignal(1);
  const [selectedId, setSelectedId] = createSignal<string | null>(null);
  const [permissions, setPermissions] = createSignal<Permission[]>([]);
  const [loadingRoles, setLoadingRoles] = createSignal(true);
  const [loadingPerms, setLoadingPerms] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [saving, setSaving] = createSignal(false);
  const [dirty, setDirty] = createSignal(false);
  const [savedAt, setSavedAt] = createSignal<string | null>(null);

  const selectedRole = () => roles().find((r) => r.id === selectedId()) ?? null;

  const goToPage = (p: number) => {
    setPage(p);
    setSelectedId(null); // 換頁後角色清單不同,清空選取由 loadRoles 自動選第一筆
    void loadRoles();
  };

  const loadRoles = async () => {
    setLoadingRoles(true);
    setError(null);
    try {
      const res = await roleClient.listRoles({ page: page(), pageSize: PAGE_SIZE });
      const list = res.roles;
      setRoles(list);
      const t = Number(res.pagination?.total ?? 0);
      setTotal(t);
      // 目前頁碼超出總頁數時退回最後一頁(與 CompaniesPage 一致)。
      const maxPage = Math.max(1, Math.ceil(t / PAGE_SIZE));
      if (page() > maxPage) {
        setPage(maxPage);
        return loadRoles();
      }
      if (!selectedId() && list.length > 0) {
        setSelectedId(list[0].id);
        void loadPermissions(list[0].id);
      }
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoadingRoles(false);
    }
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
      void queryClient.invalidateQueries({ queryKey: ["ability"] });
      await loadPermissions(role.id); // 回讀(sort_order 正規化後)
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  onMount(() => {
    void loadRoles();
  });

  return (
    <main>
      <header class="mb-6 border-b-2 border-border pb-4">
        <h1 class="text-2xl font-bold text-foreground">角色權限設置</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          管理角色功能權限(resource × action);內建角色權限為系統預設值
        </p>
      </header>

      <Show when={error()}>
        <p
          class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
          role="alert"
        >
          {error()}
        </p>
      </Show>

      <div class="grid gap-6 lg:grid-cols-[240px_1fr]">
        <Card class="h-fit">
          <Show when={loadingRoles()}>
            <p class="px-4 py-8 text-center text-sm text-muted-foreground">載入中…</p>
          </Show>
          <Show when={!loadingRoles() && roles().length === 0}>
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
