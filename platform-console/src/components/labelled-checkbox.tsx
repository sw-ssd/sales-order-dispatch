import type { JSX } from "solid-js";
import { Checkbox } from "@ui/checkbox";

/**
 * 帶可見文字的勾選框。
 *
 * 共用元件庫的 `Checkbox` 是一個 `<label>` 根節點且**不收 children**（字串要另放），
 * 所以文字放在它旁邊；無障礙名稱仍由 `aria-label` 給（與租戶 SPA 的權限矩陣同款做法）。
 */
export function LabelledCheckbox(props: {
  label: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  class?: string;
}): JSX.Element {
  return (
    <div
      class={props.class ? `flex items-center gap-2 ${props.class}` : "flex items-center gap-2"}
      onClick={(e) => {
        // 點在 Ark 的 `<label>`（＝勾選框本體）上時它自己會切換；點在旁邊的文字上則由這裡切換，
        // 讓整列都可按（`closest` 是為了不切兩次）。
        if ((e.target as HTMLElement).closest("label")) return;
        props.onCheckedChange(!props.checked);
      }}
    >
      <Checkbox
        checked={props.checked}
        onCheckedChange={props.onCheckedChange}
        aria-label={props.label}
      />
      <span class="text-sm text-foreground">{props.label}</span>
    </div>
  );
}
