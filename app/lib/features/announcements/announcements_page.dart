import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/announcement.pb.dart';
import '../../ui/adaptive.dart';

/// App 首頁（Task 49；spec announcements「前台展示與排序」）：
/// banner 輪播 ＋ 最新消息列表（news/article，article 可展開看全文）。
///
/// 可見性一律由後端決定（`ListActiveAnnouncements` 已套用 is_active、上下架時間窗、
/// 平台投放=app 與 RLS）——**前端不再自行過濾**，否則同一條規則會有兩份可能分岐的實作，
/// 而漏掉時間窗那半會在正式環境直接把未上架公告曝光。
///
/// 沒有公告時兩個區塊都不渲染（不留空輪播與「沒有公告」的噪音）。
class AnnouncementsPage extends HookWidget {
  const AnnouncementsPage({super.key, required this.api});

  final Api api;

  @override
  Widget build(BuildContext context) {
    final query = useQuery<ListActiveAnnouncementsResponse, Exception>(
      // platform 進 queryKey：web 與 app 的投放集合不同，快取不可互用。
      ['announcements', 'active', 'app'],
      () => api.announcements
          .listActiveAnnouncements(ListActiveAnnouncementsRequest(platform: 'app')),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );

    final data = query.data;
    final banners = data?.banners ?? const <Announcement>[];
    final news = data?.news ?? const <Announcement>[];
    final articles = data?.articles ?? const <Announcement>[];

    return adaptivePage(
      context,
      title: '首頁',
      body: adaptiveRefreshList(
        onRefresh: () async => query.refetch(),
        // 輪播本身是一列，其餘逐筆消息各一列。
        itemCount: (banners.isEmpty ? 0 : 1) + news.length + articles.length,
        empty: query.isLoading
            ? const Center(
                child: Padding(
                  padding: EdgeInsets.only(top: 48),
                  child: CircularProgressIndicator(),
                ),
              )
            : Center(
                child: Padding(
                  padding: const EdgeInsets.only(top: 48),
                  child: Text(
                    query.isError ? '載入失敗：${query.error}' : '目前沒有公告',
                  ),
                ),
              ),
        itemBuilder: (context, index) {
          var i = index;
          if (banners.isNotEmpty) {
            if (i == 0) return _BannerCarousel(banners: banners);
            i -= 1;
          }
          if (i < news.length) {
            return _MessageTile(item: news[i], badge: '消息');
          }
          return _MessageTile(item: articles[i - news.length], badge: '圖文');
        },
      ),
    );
  }
}

/// Banner 輪播：`PageView` ＋ 頁碼指示。
///
/// 用 `PageView` 而非引入套件：需求只有「多張、可滑動、依 sort_order」—— 原生元件
/// 已含手勢與慣性，拉一個 carousel 依賴只換來樣式覆寫。順序**照後端回傳的次序**
/// （後端已依 sort_order 排好），前端不重排。
class _BannerCarousel extends HookWidget {
  const _BannerCarousel({required this.banners});

  final List<Announcement> banners;

  @override
  Widget build(BuildContext context) {
    final controller = usePageController();
    final current = useState(0);

    return Column(
      children: [
        SizedBox(
          height: 180,
          child: PageView.builder(
            controller: controller,
            itemCount: banners.length,
            onPageChanged: (i) => current.value = i,
            itemBuilder: (context, index) => _BannerCard(banner: banners[index]),
          ),
        ),
        // 只有一張時不顯示指示（沒有東西可切）。
        if (banners.length > 1)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                for (var i = 0; i < banners.length; i++)
                  Container(
                    width: 6,
                    height: 6,
                    margin: const EdgeInsets.symmetric(horizontal: 3),
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: i == current.value
                          ? CupertinoColors.activeBlue
                          : CupertinoColors.systemGrey4,
                    ),
                  ),
              ],
            ),
          ),
      ],
    );
  }
}

/// 單張 banner。`link_url` 存在時以 `url_launcher` 開外部連結 ——
/// App 內沒有瀏覽器，沒有這條就等於「看得到標題、點不動」。
class _BannerCard extends StatelessWidget {
  const _BannerCard({required this.banner});

  final Announcement banner;

  @override
  Widget build(BuildContext context) {
    final card = Container(
      margin: const EdgeInsets.symmetric(horizontal: 12),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: CupertinoColors.separator),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (banner.imageUrl.isNotEmpty)
            Image.network(
              banner.imageUrl,
              height: 110,
              fit: BoxFit.cover,
              // 圖片載不出來不該讓整張 banner 消失（title 才是主體）。
              errorBuilder: (_, _, _) => const SizedBox.shrink(),
            ),
          Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  banner.title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                ),
                if (banner.content.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(
                      banner.content,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        fontSize: 12,
                        color: CupertinoColors.secondaryLabel.resolveFrom(context),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );

    if (banner.linkUrl.isEmpty) return card;
    // 連結以 `showModalAdaptive` 顯示網址供使用者確認/複製：App 內開外部瀏覽器需要
    // url_launcher 這個平台相依，本階段不引入（見檔頭說明與計畫殘項）。
    return GestureDetector(
      onTap: () => showModalAdaptive(
        context,
        title: banner.title,
        message: '此公告的連結：\n${banner.linkUrl}',
        confirmLabel: '知道了',
      ),
      child: card,
    );
  }
}

/// 最新消息單筆：news 直接顯示內容；article 以可展開區塊提供圖文全文。
class _MessageTile extends StatelessWidget {
  const _MessageTile({required this.item, required this.badge});

  final Announcement item;
  final String badge;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      child: Container(
        clipBehavior: Clip.antiAlias,
        decoration: BoxDecoration(
          color: Theme.of(context).cardColor,
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: CupertinoColors.separator),
        ),
        child: ExpansionTile(
          // 非 article 不需展開：內容直接顯示，展開鈕只會是空操作。
          shape: const Border(),
          title: Row(
            children: [
              Expanded(
                child: Text(
                  item.title,
                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: CupertinoColors.systemGrey5.resolveFrom(context),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(badge, style: const TextStyle(fontSize: 11)),
              ),
            ],
          ),
          childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
          children: [
            if (item.imageUrl.isNotEmpty)
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Image.network(
                  item.imageUrl,
                  fit: BoxFit.cover,
                  errorBuilder: (_, _, _) => const SizedBox.shrink(),
                ),
              ),
            Align(
              alignment: Alignment.centerLeft,
              child: Text(
                item.content,
                style: TextStyle(
                  fontSize: 13,
                  color: CupertinoColors.secondaryLabel.resolveFrom(context),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
