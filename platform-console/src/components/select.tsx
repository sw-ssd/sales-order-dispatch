import { splitProps, type JSX } from "solid-js";

/**
 * 原生 select（共用元件庫沒有 select 元件；租戶 SPA 也是手寫視覺，見
 * `frontend/src/features/users/pages/DepartmentsPage.tsx` 的 `SELECT_CLASS`）。
 *
 * 只做視覺，不接 Ark `Field` 的 a11y 關聯：console 目前的 select 都用於**篩選與選項**，
 * 沒有欄位級錯誤要關聯；`Field` 內真正需要 `aria-describedby` 的錯誤仍走 `FieldError`。
 */
const SELECT_CLASS =
  "block w-full rounded-lg border border-border bg-card py-2 pr-10 pl-3 text-sm text-foreground focus:border-primary focus:ring-3 focus:ring-primary/50 focus:outline-none disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground";

export function Select(
  props: JSX.SelectHTMLAttributes<HTMLSelectElement> & { class?: string },
): JSX.Element {
  const [local, rest] = splitProps(props, ["class", "children"]);
  return (
    <select class={local.class ? `${SELECT_CLASS} ${local.class}` : SELECT_CLASS} {...rest}>
      {local.children}
    </select>
  );
}
