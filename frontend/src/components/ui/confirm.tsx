import {
  createContext,
  createSignal,
  Show,
  useContext,
  type Component,
  type ParentProps,
} from "solid-js";
import { Button, buttonVariants } from "./button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./dialog";

/**
 * 確認對話框的**唯一**機制。
 *
 * 為什麼需要它：破壞性操作原本分兩套語言 —— 作廢走設計系統 Dialog（可填原因），
 * 其餘 16 處走原生 `window.confirm`。原生 confirm 不吃 token（是瀏覽器 chrome 的樣式）、
 * 行動裝置上樣式各異、放不下影響範圍說明，也無法在遮罩底下被看見。
 * 同一句「確定刪除？」在兩種外觀之間切換，是使用者無法預測的代價。
 * 2026-09-26 已全數收斂到這裡（原生 confirm 在 production code 歸零）。
 *
 * 用法（把原本的同步 confirm 換成 await）：
 *
 * ```tsx
 * const confirm = useConfirm();
 * const remove = async (c: Customer) => {
 *   if (!(await confirm({ title: `刪除客戶「${c.name}」`, description: "…" }))) return;
 *   …
 * };
 * ```
 *
 * 併發：**一次只會有一個確認框**。新請求會把前一個以 `false` 收尾（後者取代前者），
 * 而不是兩個 promise 一起懸著 —— 呼叫端不會拿到永不解決的 promise。正常使用下不會發生
 * （確認框是模態的，得先回答才能再觸發）；真發生時是呼叫端的 bug，以「取消」收尾比
 * 靜默洩漏 promise 安全。
 */
export interface ConfirmOptions {
  /** 標題。寫成動作本身（「刪除客戶『永和豆漿』」），不要寫「請確認」。 */
  title: string;
  /** 影響範圍說明：這一下會發生什麼、可不可以復原。 */
  description?: string;
  /** 確認鈕文字。預設「確定」；破壞性操作應寫出動詞（「刪除」「作廢」）。 */
  confirmLabel?: string;
  /** 取消鈕文字。預設「取消」。 */
  cancelLabel?: string;
  /** 確認鈕樣式。破壞性操作傳 `destructive`（紅底），其餘用預設主色。 */
  variant?: "default" | "destructive";
}

type ConfirmFn = (opts: ConfirmOptions) => Promise<boolean>;

const ConfirmContext = createContext<ConfirmFn>();

/** 取得確認函式；必須在 `ConfirmProvider` 之內（否則拋錯，不靜默變成「直接放行」）。 */
export function useConfirm(): ConfirmFn {
  const fn = useContext(ConfirmContext);
  if (!fn) {
    throw new Error("useConfirm 必須在 ConfirmProvider 之內使用");
  }
  return fn;
}

/**
 * 掛載位置：`App` 的最外層（與 `AbilityProvider` 同層），因此 shell 與 chromeless 路徑
 * 都涵蓋得到。
 */
export const ConfirmProvider: Component<ParentProps> = (props) => {
  const [opts, setOpts] = createSignal<ConfirmOptions | null>(null);
  // 用一般變數存 resolver 而非 signal：把函式傳進 signal setter 會被當成 updater
  // （Solid 的 setter 對函式參數有特殊語意），下場是「收尾前一個」靜默失效。
  let settle: ((ok: boolean) => void) | undefined;

  const confirm: ConfirmFn = (next) => {
    settle?.(false);
    setOpts(next);
    return new Promise<boolean>((resolve) => {
      settle = resolve;
    });
  };

  const done = (ok: boolean) => {
    const s = settle;
    settle = undefined;
    setOpts(null);
    s?.(ok);
  };

  return (
    <ConfirmContext.Provider value={confirm}>
      {props.children}
      <Dialog
        open={opts() !== null}
        onOpenChange={(open) => {
          // Esc、右上關閉鈕都收斂成「取消」。
          if (!open) done(false);
        }}
        // 不因誤點遮罩而關閉：誤觸應該是取消，但不該讓使用者在沒看清楚的情況下被關掉，
        // 以為自己已確認（dialog.md 對破壞性確認框的建議）。
        closeOnOutsideClick={false}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{opts()?.title}</DialogTitle>
            <Show when={opts()?.description}>
              {(desc) => <DialogDescription>{desc()}</DialogDescription>}
            </Show>
          </DialogHeader>
          <DialogFooter>
            {/* 取消排在 DOM 前面：焦點先落在安全選項，誤按 Enter 不會執行破壞性動作。 */}
            <DialogClose
              class={buttonVariants({ variant: "outline" })}
              onClick={() => done(false)}
            >
              {opts()?.cancelLabel ?? "取消"}
            </DialogClose>
            <Button
              type="button"
              variant={opts()?.variant ?? "default"}
              onClick={() => done(true)}
            >
              {opts()?.confirmLabel ?? "確定"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </ConfirmContext.Provider>
  );
};
