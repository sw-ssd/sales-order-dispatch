import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/notifications.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';

const _pageSize = 50;

/// 通知中心頁（tab 3）：全部/未讀切換、點列標記已讀並顯示全文。
class NotificationsPage extends HookWidget {
  const NotificationsPage({super.key, required this.api});

  final Api api;

  @override
  Widget build(BuildContext context) {
    final unreadOnly = useState(false);
    final page = useState(1);

    final query = useQuery<ListNotificationsResponse, Exception>(
      ['notifications', unreadOnly.value, page.value],
      () => api.notifications.listNotifications(ListNotificationsRequest(
        unreadOnly: unreadOnly.value,
        page: page.value,
        pageSize: _pageSize,
      )),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );

    final data = query.data;
    final list = data?.notifications ?? const <NotificationView>[];
    final total = data?.total ?? 0;
    final pageCount = total == 0 ? 1 : (total + _pageSize - 1) ~/ _pageSize;

    Future<void> markReadAndShow(NotificationView n) async {
      final isUnread = n.readAt.isEmpty;
      await showModalAdaptive(
        context,
        title: n.title,
        message: [
          n.content,
          if (n.payload.isNotEmpty) '\n${n.payload}',
          formatDateTime(n.createdAt),
        ].join('\n'),
        confirmLabel: '關閉',
      );
      if (isUnread) {
        try {
          await api.notifications
              .markRead(MarkReadRequest(notificationIds: [n.id]));
          query.refetch();
        } catch (_) {
          // 標記失敗不干擾閱讀；下次進入重試。
        }
      }
    }

    return adaptivePage(
      context,
      title: '通知',
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Row(
              children: [
                adaptiveFilterChip(
                  label: '全部',
                  selected: !unreadOnly.value,
                  onTap: () {
                    unreadOnly.value = false;
                    page.value = 1;
                  },
                ),
                const SizedBox(width: 8),
                adaptiveFilterChip(
                  label: '未讀',
                  selected: unreadOnly.value,
                  onTap: () {
                    unreadOnly.value = true;
                    page.value = 1;
                  },
                ),
                const Spacer(),
                if (data != null && data.unreadCount > 0)
                  Text('未讀 ${data.unreadCount}'),
              ],
            ),
          ),
          Expanded(
            child: query.isError
                ? Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text('載入失敗：${query.error}'),
                        const SizedBox(height: 8),
                        AdaptiveButton(
                          onPressed: () => query.refetch(),
                          child: const Text('重試'),
                        ),
                      ],
                    ),
                  )
                : adaptiveRefreshList(
                    onRefresh: () async => query.refetch(),
                    itemCount: list.length,
                    empty: query.isLoading
                        ? const Center(
                            child: Padding(
                              padding: EdgeInsets.only(top: 48),
                              child: CircularProgressIndicator(),
                            ),
                          )
                        : const Center(
                            child: Padding(
                              padding: EdgeInsets.only(top: 48),
                              child: Text('沒有通知'),
                            ),
                          ),
                    itemBuilder: (context, index) {
                      final n = list[index];
                      final unread = n.readAt.isEmpty;
                      return adaptiveListTile(
                        title: Row(
                          children: [
                            if (unread)
                              Padding(
                                padding: const EdgeInsets.only(right: 6),
                                child: Container(
                                  width: 8,
                                  height: 8,
                                  decoration: const BoxDecoration(
                                    color: Color(0xFF1565C0),
                                    shape: BoxShape.circle,
                                  ),
                                ),
                              ),
                            Expanded(
                              child: Text(
                                n.title,
                                style: TextStyle(
                                  fontWeight: unread
                                      ? FontWeight.w700
                                      : FontWeight.w400,
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),
                        subtitle: Text(
                          '${n.content} · ${formatDateTime(n.createdAt)}',
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        onTap: () => markReadAndShow(n),
                      );
                    },
                  ),
          ),
          if (total > _pageSize)
            Container(
              padding: const EdgeInsets.symmetric(vertical: 4),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  AdaptiveButton(
                    onPressed: page.value > 1
                        ? () => page.value = page.value - 1
                        : null,
                    child: const Text('上一頁'),
                  ),
                  Text('第 ${page.value}/$pageCount 頁'),
                  AdaptiveButton(
                    onPressed: page.value < pageCount
                        ? () => page.value = page.value + 1
                        : null,
                    child: const Text('下一頁'),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}
