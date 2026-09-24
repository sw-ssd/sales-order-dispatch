import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/returns.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';
import 'return_create_page.dart';
import 'return_detail_page.dart';

const _pageSize = 20;

const _statusFilters = <(String, String)>[
  ('', '全部'),
  ('pending', '待審核'),
  ('approved', '已核准'),
  ('rejected', '已拒絕'),
];

/// 退貨清單頁（tab 2）：狀態篩選＋分頁＋下拉更新＋發起退貨入口。
class ReturnsPage extends HookWidget {
  const ReturnsPage({super.key, required this.api});

  final Api api;

  @override
  Widget build(BuildContext context) {
    final status = useState('');
    final page = useState(1);

    final query = useQuery<ListReturnRequestsResponse, Exception>(
      ['returns', status.value, page.value],
      () => api.returns.listReturnRequests(ListReturnRequestsRequest(
        status: status.value,
        page: page.value,
        pageSize: _pageSize,
      )),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );

    final data = query.data;
    final entries = data?.entries ?? const <ReturnRequestEntry>[];
    final total = data?.total ?? 0;
    final pageCount = total == 0 ? 1 : (total + _pageSize - 1) ~/ _pageSize;

    return adaptivePage(
      context,
      title: '退貨',
      trailing: CupertinoButton(
        padding: EdgeInsets.zero,
        onPressed: () async {
          final created = await pushAdaptive<bool>(
            context,
            ReturnCreatePage(api: api),
          );
          if (created == true) query.refetch();
        },
        child: const Text('新建'),
      ),
      actions: [
        IconButton(
          icon: const Icon(Icons.add),
          tooltip: '發起退貨',
          onPressed: () async {
            final created = await pushAdaptive<bool>(
              context,
              ReturnCreatePage(api: api),
            );
            if (created == true) query.refetch();
          },
        ),
      ],
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: [
                  for (final (value, label) in _statusFilters)
                    Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: adaptiveFilterChip(
                        label: label,
                        selected: status.value == value,
                        onTap: () {
                          status.value = value;
                          page.value = 1;
                        },
                      ),
                    ),
                ],
              ),
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
                    itemCount: entries.length,
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
                              child: Text('沒有退貨申請'),
                            ),
                          ),
                    itemBuilder: (context, index) {
                      final entry = entries[index];
                      return adaptiveListTile(
                        title: Row(
                          children: [
                            Expanded(
                              child: Text('退貨 #${entry.id}',
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w600)),
                            ),
                            statusChip(entry.status),
                          ],
                        ),
                        subtitle: Text(
                          [
                            '申請 ${formatDateTime(entry.createdAt)}',
                            if (entry.remark.isNotEmpty) entry.remark,
                          ].join(' · '),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        trailing: Icon(
                          isCupertinoTarget()
                              ? CupertinoIcons.chevron_forward
                              : Icons.chevron_right,
                          size: 18,
                          color: CupertinoColors.inactiveGray,
                        ),
                        onTap: () async {
                          final changed = await pushAdaptive<bool>(
                            context,
                            ReturnDetailPage(api: api, returnId: entry.id),
                          );
                          if (changed == true) query.refetch();
                        },
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
