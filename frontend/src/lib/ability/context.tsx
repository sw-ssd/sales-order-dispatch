import { createContext, useContext, type Accessor, type ParentComponent } from "solid-js";
import type { AppAbility } from "./service";

const AbilityContext = createContext<Accessor<AppAbility>>();

export const AbilityProvider: ParentComponent<{ ability: Accessor<AppAbility> }> = (props) => (
  // ability 本身即 accessor(signal/memo),此處傳遞的是 accessor 參照而非讀值;
  // 規則誤判為「只讀一次」,實際 context 值即為該 reactive accessor。
  // eslint-disable-next-line solid/reactivity
  <AbilityContext.Provider value={props.ability}>{props.children}</AbilityContext.Provider>
);

// 回傳 accessor 而非實例:ability 更新=整個實例替換,JSX 內引用點隨 signal 重算。
export function useAbility(): Accessor<AppAbility> {
  const acc = useContext(AbilityContext);
  if (!acc) throw new Error("useAbility must be used within <AbilityProvider>");
  return acc;
}
