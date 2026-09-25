import { createSignal, For, Show, type Component } from "solid-js";
import { createQuery } from "@tanstack/solid-query";
import { Badge, Card, CardContent, CardHeader, CardTitle } from "~/components/ui";
import { queryData } from "~/lib/query-data";
import type { Announcement } from "~/lib/proto/salesorder/v1/announcement_pb";
import { activeAnnouncementsQueryOptions } from "../queries";

/**
 * 公告前台（spec「前台展示與排序」）：Web 中台的 banner 輪播 ＋ news/article 最新消息列表。
 *
 * 掛在 Dashboard 而非獨立路由：spec 說的是「使用者開啟 Web 首頁」，公告是首頁的一塊，
 * 另開一頁等於沒人在首頁看到它。
 *
 * 可見性一律由後端決定（`ListActiveAnnouncements` 已套用時間窗、is_active、平台投放與 RLS）；
 * 這一層**不再自行過濾** —— 前端再篩一次會產生兩份可能分岐的規則。
 *
 * 沒有公告時整塊不渲染：空輪播會佔掉首頁一大塊並顯示「沒有公告」，對使用者是噪音。
 */
const AnnouncementBoard: Component = () => {
  const query = createQuery(() => activeAnnouncementsQueryOptions("web"));

  const banners = () => queryData(query, (d) => d?.banners ?? []) ?? [];
  const news = () => queryData(query, (d) => d?.news ?? []) ?? [];
  const articles = () => queryData(query, (d) => d?.articles ?? []) ?? [];
  /** 最新消息列表＝news ＋ article（spec：兩者都以列表展示；article 另提供圖文檢視）。 */
  const messages = () => [...news(), ...articles()];

  // 兩塊各自守衛自己的內容；**沒有載入中提示是刻意的** —— 首頁掛了訂單概況等多支查詢，
  // 公告再閃一個 spinner 只是噪音，而公告本身是附加資訊。失敗同樣不顯示錯誤區塊
  // （訂單概況那幾張卡才是首頁主體，它們的失敗才有提示）。
  //
  // 也刻意不做「整塊空就不渲染」的外層守衛：資料為空時兩個 Show 都不成立，
  // section 裡沒有任何可見內容 —— 外層判斷只是多一層不會改變結果的條件。
  return (
    <section class="mt-8">
      <Show when={banners().length > 0}>
        <BannerCarousel banners={banners()} />
      </Show>

      <Show when={messages().length > 0}>
        <div class="mt-6">
          <h2 class="mb-3 text-lg font-semibold text-foreground">最新消息</h2>
          <ul class="flex flex-col gap-3">
            <For each={messages()}>
              {(item) => <MessageCard item={item} />}
            </For>
          </ul>
        </div>
      </Show>
    </section>
  );
};

/**
 * Banner 輪播（spec「以輪播展示 banner」）。
 *
 * 自製而非引入套件：需求只有「多張、可切換、依 sort_order」—— 一顆 `index` signal
 * 加兩個按鈕，比拉一個 carousel 依賴（及其樣式覆寫）小得多。
 *
 * 只有一張時不顯示切換控制：沒有東西可切，按了不動的按鈕比沒有按鈕更糟。
 * `link_url` 存在時整張可點（spec「Banner 連結導向」），以 `<a>` 呈現讓
 * 瀏覽器原生行為（新分頁、右鍵、狀態列顯示網址）都成立。
 */
const BannerCarousel: Component<{ banners: Announcement[] }> = (props) => {
  const [index, setIndex] = createSignal(0);
  // 公告集合改變（重新取得、刪除）時把索引夾回範圍內 —— 否則會停在一個空的輪播上。
  const current = () => {
    const list = props.banners;
    if (list.length === 0) return undefined;
    return list[Math.min(index(), list.length - 1)];
  };
  const step = (delta: number) => {
    const n = props.banners.length;
    setIndex((i) => (i + delta + n) % n);
  };

  return (
    <div class="relative overflow-hidden rounded-lg border border-border bg-card shadow-xs">
      <Show when={current()}>
        {(banner) => (
          <>
            <BannerContent banner={banner()} />
            <Show when={props.banners.length > 1}>
              <div class="absolute inset-x-0 bottom-0 flex items-center justify-between bg-gradient-to-t from-black/50 to-transparent px-3 py-2">
                <button
                  type="button"
                  class="rounded px-2 py-1 text-sm text-white hover:bg-white/20"
                  aria-label="上一則"
                  onClick={() => step(-1)}
                >
                  ‹
                </button>
                <span class="text-xs text-white/90 tabular-nums">
                  {index() + 1} / {props.banners.length}
                </span>
                <button
                  type="button"
                  class="rounded px-2 py-1 text-sm text-white hover:bg-white/20"
                  aria-label="下一則"
                  onClick={() => step(1)}
                >
                  ›
                </button>
              </div>
            </Show>
          </>
        )}
      </Show>
    </div>
  );
};

/** 單張 banner：有 link_url 時整塊是連結，否則只是內容。 */
const BannerContent: Component<{ banner: Announcement }> = (props) => {
  const inner = () => (
    <>
      <Show when={props.banner.imageUrl}>
        {(url) => (
          <img
            src={url()}
            alt={props.banner.title}
            class="h-40 w-full object-cover sm:h-52"
          />
        )}
      </Show>
      <div class="p-5">
        <h2 class="text-lg font-semibold text-foreground">{props.banner.title}</h2>
        <Show when={props.banner.content}>
          <p class="mt-1 line-clamp-2 text-sm text-muted-foreground">
            {props.banner.content}
          </p>
        </Show>
      </div>
    </>
  );

  return (
    <Show when={props.banner.linkUrl} fallback={inner()}>
      {(href) => (
        <a
          href={href()}
          target="_blank"
          rel="noopener noreferrer"
          class="block hover:bg-muted/40"
        >
          {inner()}
        </a>
      )}
    </Show>
  );
};

/**
 * 最新消息單筆：`article` 類型可展開看圖文全文（spec「article 提供圖文內容檢視」）。
 *
 * 用 `<details>` 而非對話框：全文檢視是就地閱讀，開對話框會蓋掉整個首頁，
 * 而原生 `<details>` 免費帶鍵盤與螢幕閱讀器語意。
 */
const MessageCard: Component<{ item: Announcement }> = (props) => {
  const isArticle = () => props.item.type === "article";

  return (
    <li>
      <Card>
        <CardHeader class="flex-row items-center justify-between gap-3">
          <CardTitle class="text-base">{props.item.title}</CardTitle>
          <Badge variant="secondary">
            {isArticle() ? "圖文" : "消息"}
          </Badge>
        </CardHeader>
        <CardContent>
          <Show when={props.item.imageUrl}>
            {(url) => (
              <img
                src={url()}
                alt={props.item.title}
                class="mb-3 max-h-56 w-full rounded object-cover"
              />
            )}
          </Show>
          <Show
            when={isArticle()}
            fallback={
              <p class="whitespace-pre-wrap text-sm text-muted-foreground">
                {props.item.content}
              </p>
            }
          >
            <details>
              <summary class="cursor-pointer text-sm font-medium text-primary">
                閱讀全文
              </summary>
              <p class="mt-2 whitespace-pre-wrap text-sm text-muted-foreground">
                {props.item.content}
              </p>
            </details>
          </Show>
        </CardContent>
      </Card>
    </li>
  );
};

export default AnnouncementBoard;
