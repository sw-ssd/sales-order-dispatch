import { Show, type JSX, type ParentComponent } from "solid-js";
import { useAbility } from "./context";
import { hasPermission } from "./permissions";

/**
 * 顯示控制元件:依權限集合判斷 resource 是否具 action(取代 CASL ability)。
 *   <Can I="read" a="sales_order" fallback={<Disabled/>}>
 * 物件狀態條件由 domain 處理,不在此做 instance 判斷。
 */
export const Can: ParentComponent<{
  I: string;
  a: string;
  fallback?: JSX.Element;
}> = (props) => {
  const ability = useAbility();
  const allowed = () => hasPermission(ability(), props.a, props.I);
  return <Show when={allowed()} fallback={props.fallback}>{props.children}</Show>;
};
