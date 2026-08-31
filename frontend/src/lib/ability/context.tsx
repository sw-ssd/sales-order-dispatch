import { createContext, useContext, type Accessor, type ParentComponent } from "solid-js";
import type { AppAbility } from "./service";

const AbilityContext = createContext<Accessor<AppAbility>>();

export const AbilityProvider: ParentComponent<{ ability: Accessor<AppAbility> }> = (props) => (
  <AbilityContext.Provider value={props.ability}>{props.children}</AbilityContext.Provider>
);

// 回傳 accessor 而非實例:ability 更新=整個實例替換,JSX 內引用點隨 signal 重算。
export function useAbility(): Accessor<AppAbility> {
  const acc = useContext(AbilityContext);
  if (!acc) throw new Error("useAbility must be used within <AbilityProvider>");
  return acc;
}
