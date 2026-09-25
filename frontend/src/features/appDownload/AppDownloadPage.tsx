import { For } from "solid-js";

import { buttonVariants } from "~/components/ui";

/**
 * App 下載落頁（Universal Link / App Link 的**未安裝**分支）。
 *
 * 店家的 QR 登入與帳號管理連結都是 App Link：手機已安裝 App 時由作業系統攔截、直接開
 * App，這個頁面根本不會被載入。只有「未安裝」時瀏覽器才會落到這裡，於是它的唯一職責
 * 是把人送到商店（規格 §4.2 兩條 Requirement 的 Scenario：未安裝導向商店）。
 *
 * 商店網址尚未定案（App 未上架，repo 內沒有真實 id）。刻意**不寫死假網址**：兩顆指向
 * 不存在 App 的按鈕會讓使用者以為壞了，比誠實說明更糟。上架後設
 * `VITE_IOS_APP_URL` / `VITE_ANDROID_APP_URL` 即自動生效。
 *
 * ponytail: 上架前以文案代替死連結,上架後由 env 注入真實網址
 */
const IOS_APP_URL = import.meta.env.VITE_IOS_APP_URL ?? "";
const ANDROID_APP_URL = import.meta.env.VITE_ANDROID_APP_URL ?? "";

/** 偵測裝置以決定先推哪個商店（僅影響排序，兩者皆提供）。 */
function isIOS(): boolean {
  if (typeof navigator === "undefined") return false;
  return /iPad|iPhone|iPod/.test(navigator.userAgent);
}

/**
 * 兩種流程共用同一個版面：差別只有文案，商店按鈕與未設定時的退路完全相同 ——
 * 拆成兩個元件等於把同一段 JSX 抄兩遍。
 */
function AppDownloadPrompt(props: { title: string; description: string }) {
  const stores = [
    { label: "App Store", url: IOS_APP_URL, primary: isIOS() },
    { label: "Google Play", url: ANDROID_APP_URL, primary: !isIOS() },
  ]
    .filter((s) => s.url !== "")
    .sort((a, b) => Number(b.primary) - Number(a.primary));

  return (
    <main class="flex min-h-dvh items-center justify-center bg-card px-6 py-16 text-card-foreground">
      <div class="w-full max-w-lg space-y-6 text-center">
        <h1 class="text-2xl font-extrabold text-foreground md:text-3xl">{props.title}</h1>
        <p class="font-medium text-muted-foreground">{props.description}</p>

        {stores.length > 0 ? (
          <div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
            <For each={stores}>
              {(store) => (
                <a
                  href={store.url}
                  class={buttonVariants({
                    variant: store.primary ? "default" : "outline",
                  })}
                >
                  前往 {store.label}
                </a>
              )}
            </For>
          </div>
        ) : (
          // 未設定商店網址（App 尚未上架）：給可行的下一步，而不是兩顆死按鈕。
          <p class="text-sm text-muted-foreground">
            App 下載連結尚未開放，請聯絡您的業務人員取得安裝方式。
          </p>
        )}
      </div>
    </main>
  );
}

/** `/customer_account_qrcode/{token}` 的未安裝落頁。 */
export function QRDownloadPage() {
  return (
    <AppDownloadPrompt
      title="請先安裝 App 再掃碼登入"
      description="這組 QR Code 是店家登入用的。安裝 App 後再用手機相機掃描，或請業務重新傳送連結。"
    />
  );
}

/** `/customer_account_manage` 的未安裝落頁。 */
export function ManageDownloadPage() {
  return (
    <AppDownloadPrompt
      title="請先安裝 App 管理帳號"
      description="帳號管理在 App 內操作。安裝 App 後由業務提供的連結開啟，並以主帳號登入。"
    />
  );
}
