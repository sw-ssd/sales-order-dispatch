import { createContext, useContext, type Accessor, type ParentComponent } from "solid-js";
import type { AppAbility } from "./service";

const AbilityContext = createContext<Accessor<AppAbility>>();

export const AbilityProvider: ParentComponent<{ ability: Accessor<AppAbility> }> = (props) => (
  // Solid 2:context 本身即 provider;value 傳遞的是 accessor 參照而非讀值,
  // 下游 useContext 取得同一 reactive accessor。
  // eslint-disable-next-line solid/reactivity
  <AbilityContext value={props.ability}>{props.children}</AbilityContext>
);

// 回傳 accessor 而非實例:ability 更新=整個實例替換,JSX 內引用點隨 signal 重算。
// default-less context 在無 Provider 時由 useContext 拋 ContextNotFoundError。
export function useAbility(): Accessor<AppAbility> {
  return useContext(AbilityContext);
}
