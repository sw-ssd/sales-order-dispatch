import { createQuery } from "@tanstack/solid-query";
import { Truck } from "lucide-solid";
import { Show } from "solid-js";
import { meQueryOptions } from "~/lib/me";

/**
 * 側邊欄品牌圖示（規格 §8.1「側邊欄顯示當前公司 Logo」）：所屬公司有 Logo 用公司 Logo，
 * 其餘（無公司／無 Logo／未登入／查詢失敗）一律退回系統預設圖示——降級只發生在顯示層，
 * 永遠不阻斷導覽，也不對失敗狀態做重試對話框之類的介入。
 */
export default function BrandLogo() {
  const me = createQuery(() => meQueryOptions);
  const logo = () => me.data?.company?.logo_url ?? "";
  return (
    <Show
      when={logo()}
      fallback={<Truck class="size-5 flex-none text-primary transition group-hover:scale-110" />}
    >
      {/* alt＝公司名：收合成 icon rail 時標題 span 雖 sr-only 仍在無障礙樹，
          Logo 因此提供「哪一家公司」的語意，而不是重複標題文字。 */}
      <img
        src={logo()}
        alt={me.data?.company?.name ?? ""}
        class="size-5 flex-none rounded object-contain transition group-hover:scale-110"
      />
    </Show>
  );
}
