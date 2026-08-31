import { createEffect, Show, type ParentComponent } from "solid-js";
import type { JSX } from "@solidjs/web";

/**
 * 表單對話框外殼(原生實作;取代 @ark-ui/solid Dialog —
 * ark 3.0.0-7 與 solid-js 2 runtime 不相容)。
 * 語意對齊 ark Dialog:開啟時掛載、點擊背板外區域關閉、Escape 關閉、
 * 關閉按鈕 type="button";外觀 class 沿用原 ark 版。
 */
export const FormDialog: ParentComponent<{
  open: boolean;
  onClose: () => void;
  title: JSX.Element;
  description: JSX.Element;
}> = (props) => {
  // Escape 關閉:僅在開啟期間監聽,關閉即移除(cleanup)。
  createEffect(
    () => props.open,
    (open) => {
      if (!open) return;
      const onKey = (e: KeyboardEvent) => {
        if (e.key === "Escape") props.onClose();
      };
      window.addEventListener("keydown", onKey);
      return () => window.removeEventListener("keydown", onKey);
    },
  );

  return (
    <Show when={props.open}>
      <div class="fixed inset-0 z-40 bg-black/40" aria-hidden="true" />
      <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        onClick={(e) => {
          // 點擊內容外的 positioner 區域 = ark 的 interact-outside 關閉
          if (e.target === e.currentTarget) props.onClose();
        }}
      >
        <div
          role="dialog"
          aria-modal="true"
          class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl"
        >
          <div class="flex items-start justify-between">
            <div>
              <h2 class="text-lg font-bold text-gray-900">{props.title}</h2>
              <p class="mt-1 text-sm text-gray-500">{props.description}</p>
            </div>
            <button
              type="button"
              onClick={() => props.onClose()}
              class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
              aria-label="關閉"
            >
              ×
            </button>
          </div>
          {props.children}
        </div>
      </div>
    </Show>
  );
};
