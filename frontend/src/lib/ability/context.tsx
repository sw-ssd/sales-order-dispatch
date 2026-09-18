import { createContext, useContext, type Accessor, type ParentComponent } from "solid-js";
import type { AbilitySet } from "./service";

const AbilityContext = createContext<Accessor<AbilitySet>>();

export const AbilityProvider: ParentComponent<{ ability: Accessor<AbilitySet> }> = (props) => (
  // 權限集合為 accessor(signal/memo),此處傳遞 accessor 參照;JSX 內引用點隨 signal 重算。
  // eslint-disable-next-line solid/reactivity
  <AbilityContext.Provider value={props.ability}>{props.children}</AbilityContext.Provider>
);

// 回傳 accessor 而非集合實例:權限更新=整個集合替換,JSX 內引用點隨 signal 重算。
export function useAbility(): Accessor<AbilitySet> {
  const acc = useContext(AbilityContext);
  if (!acc) throw new Error("useAbility must be used within <AbilityProvider>");
  return acc;
}
