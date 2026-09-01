import { Show, type JSX, type ParentComponent } from "solid-js";
import { useAbility } from "./context";

/**
 * 顯示控制元件。instance 判斷時以 CASL subject() 包裝傳入(前端唯一 instance 包裝慣例):
 *   <Can I="cancel" a={subject("sales_order", order)} fallback={<Disabled/>}>
 */
export const Can: ParentComponent<{
  I: string;
  a: string | object;
  fallback?: JSX.Element;
}> = (props) => {
  const ability = useAbility();
  const allowed = () => ability().can(props.I, props.a as never);
  return <Show when={allowed()} fallback={props.fallback}>{props.children}</Show>;
};
