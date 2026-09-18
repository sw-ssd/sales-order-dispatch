# 前端 OpenFGA 能力遷移 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 Web 中台的 CASL ability 遷移至「後端 OpenFGA proxy」驅動的權限判斷（移除 `@casl/ability`），選單/路由/按鈕守衛改用後端回傳的可用 (resource, action) 集合。

**Architecture:** 後端 `GetAbility` 已改為 OpenFGA 驅動，回傳目前身分可用 (resource, action) 集合。前端以輕量 `hasPermission(resource, action)`（Set 查詢）取代 CASL ability；`queryKey ["ability"]` 失效語意保留，權限異動後 `invalidateQueries` 精準重載。

**Tech Stack:** SolidJS、TanStack Router、TanStack Query、connect-es 生成 `AbilityService`、移除 `@casl/ability`。

**Spec:** `docs/superpowers/specs/2026-09-18-openfga-authz-design.md`（§4 前端遷移）

## Global Constraints
- TypeScript strict；遵循既有目錄慣例（`src/lib/ability/*`）。
- 不引入新權限依賴（本波終點為移除 `@casl/ability`）。
- 前端守衛不構成授權；後端仍為唯一授權決策者（此遷移僅為 UI 控制，不改變安全語意）。
- 提交前跑前端測試 `pnpm test` + lint。

---

### Task F1: 建 permissions 模組（取代 CASL ability）

**Files:**
- Create: `frontend/src/lib/ability/permissions.ts`
- Modify: `frontend/src/lib/ability/service.ts`
- Test: `frontend/src/lib/ability/permissions.test.ts`

**Interfaces:**
- Produces: `type Permission = { resource: string; action: string }`；`createPermissions(rules: Permission[]): Set<string>`（key = `${resource}:${action}`）；`hasPermission(perms: ReadonlySet<string>, resource, action): boolean`。
- Consumes: 後端 `GetAbility` 回傳的規則（現 proto 型別 `AbilityRule` 仍可用，取其 resource/action）。

- [ ] **Step 1: 寫失敗測試**

`permissions.test.ts`：`createPermissions` 建集合、`hasPermission` true/false、空規則 fail-closed false。

- [ ] **Step 2: 跑測試確認失敗** → Run: `cd frontend && pnpm vitest run src/lib/ability/permissions.test.ts`。Expected: FAIL（模組不存在）。

- [ ] **Step 3: 實作 permissions.ts**

```ts
export type Permission = { resource: string; action: string };
export const permKey = (resource: string, action: string) => `${resource}:${action}`;
export function createPermissions(rules: Permission[]): Set<string> {
  return new Set(rules.map((r) => permKey(r.resource, r.action)));
}
export function hasPermission(perms: ReadonlySet<string>, resource: string, action: string): boolean {
  return perms.has(permKey(resource, action));
}
```

- [ ] **Step 4: 跑測試確認通過** → Run: `pnpm vitest run src/lib/ability/permissions.test.ts`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/ability/permissions.ts frontend/src/lib/ability/permissions.test.ts
git commit -m "feat(frontend): permissions Set 模組取代 CASL ability(核心)"
```

---

### Task F2: service.ts 改載入權限集合（取代 createAppAbility）

**Files:**
- Modify: `frontend/src/lib/ability/service.ts`
- Test: `frontend/src/lib/ability/service.test.ts`（改寫）

**Interfaces:**
- Produces: `abilityQueryOptions` 的 `queryFn` 回傳 `Set<string>`（由 `client.getAbility({}).rules` → `createPermissions`）；queryKey 仍 `["ability"]`。
- Consumes: Task F1。

- [ ] **Step 1: 改寫 service.ts**

`queryFn` 改為 `createPermissions(await (await client.getAbility({})).rules)`；移除 `createMongoAbility`/`RawRuleOf`/`MongoAbility` 與 `subject`。

- [ ] **Step 2: 改寫 service.test.ts**

改為驗證 `queryFn` 產出 Set 且空規則 fail-closed。

- [ ] **Step 3: 跑測試確認通過** → Run: `pnpm vitest run src/lib/ability`。Expected: PASS。

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/ability/service.ts frontend/src/lib/ability/service.test.ts
git commit -m "refactor(frontend): service 載入權限 Set 取代 ability 規則"
```

---

### Task F3: guards / Can / context 改權限集合

**Files:**
- Modify: `frontend/src/lib/ability/guards.ts`
- Modify: `frontend/src/lib/ability/Can.tsx`
- Modify: `frontend/src/lib/ability/context.tsx`
- Modify: `frontend/src/lib/ability/guards.test.ts`、`frontend/src/lib/ability/Can.test.tsx`

**Interfaces:**
- Consumes: Task F1/F2。
- Produces: `requireAbility(action, resource)` 用 `hasPermission`；`Can` 元件 props 不變（`I`/`a`）。

- [ ] **Step 1: guards.ts 改寫**

`requireAbility`：`ensureQueryData(abilityQueryOptions)` 取得 `Set` → `hasPermission(perms, resource, action)`。移除 `createAppAbility`。

- [ ] **Step 2: Can.tsx / context.tsx 改寫**

`Can.tsx` 用 `useAbility()`（改回傳權限 Set accessor）→ `hasPermission`。`context.tsx` 型別改 `Accessor<ReadonlySet<string>>`。

- [ ] **Step 3: 改寫 guards.test.ts / Can.test.tsx**

移除 `@casl/ability` 的 `subject` 用法，改以權限集合作斷言。

- [ ] **Step 4: 跑測試確認通過** → Run: `pnpm vitest run src/lib/ability src/router src/features/users`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/ability
git commit -m "refactor(frontend): guards/Can/context 改權限 Set 判斷"
```

---

### Task F4: 移除 @casl/ability 依賴 + 清路由/頁面

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/src/router/index.tsx`、`frontend/src/features/users/pages/RolesPage.tsx`（若直接 import @casl）
- Test: 全量前端測試

**Interfaces:**
- Consumes: Task F1–F3。
- Produces: 無 `@casl/ability` 依賴；全前端編譯/測試通過。

- [ ] **Step 1: 移除依賴**

`pnpm remove @casl/ability`（在 `frontend/`）。grep 確認無殘留 import。

- [ ] **Step 2: 修殘留 import**

`RolesPage.tsx`/`router/index.tsx` 內能力判斷改 `hasPermission`/`requireAbility`（若仍引用 @casl）。

- [ ] **Step 3: 全量跑測試 + build**

Run: `cd frontend && pnpm test && pnpm build`。Expected: PASS。

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json frontend/pnpm-lock.yaml frontend/src
git commit -m "chore(frontend): 移除 @casl/ability;選單/守衛改後端 OpenFGA proxy"
```

---

## 執行驗收（DoD）
- 無 `@casl/ability` 依賴與 import。
- `requireAbility`/`Can` 以 `hasPermission`（後端 OpenFGA 結果）判斷。
- `queryKey ["ability"]` invalidate 語意保留，權限異動後精準重載。
- 前端 `pnpm test` + `pnpm build` 通過。

*最後更新：2026-09-18*
