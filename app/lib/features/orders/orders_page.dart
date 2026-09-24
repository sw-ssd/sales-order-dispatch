import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/salesorder.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';
import 'order_detail_page.dart';

const _pageSize = 20;

const _statusFilters = <(String, String)>[
  ('', '全部'),
  ('pending', '待處理'),
  ('processing', '處理中'),
  ('completed', '已完成'),
];

const _eventLabels = <String, String>{
  'create': '建單',
  'edit': '編輯',
  'dispatch': '派車',
  'dispatch_cancel': '取消派車',
  'cancel': '取消',
  'complete': '完成',
  'void': '作廢',
};

String eventLabel(String type) => _eventLabels[type] ?? type;

/// 訂單清單頁（tab 1）：狀態篩選＋分頁＋下拉更新。
class OrdersPage extends HookWidget {
  const OrdersPage({super.key, required this.api});

  final Api api;

  @override
  Widget build(BuildContext context) {
    final status = useState('');
    final page = useState(1);

    final query = useQuery<ListOrdersResponse, Exception>(
      ['orders', status.value, page.value],
      () => api.orders
          .listOrders(ListOrdersRequest(
            page: page.value,
            pageSize: _pageSize,
            status: status.value,
          ))
          ,
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );

    final data = query.data;
    final orders = data?.orders ?? const <SalesOrder>[];
    final total = data?.total ?? 0;
    final pageCount = total == 0 ? 1 : (total + _pageSize - 1) ~/ _pageSize;

    return adaptivePage(
      context,
      title: '訂單',
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
                    itemCount: orders.length,
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
                              child: Text('沒有訂單'),
                            ),
                          ),
                    itemBuilder: (context, index) {
                      final order = orders[index];
                      return adaptiveListTile(
                        title: Row(
                          children: [
                            Text(order.orderNo,
                                style: const TextStyle(
                                    fontWeight: FontWeight.w600)),
                            const SizedBox(width: 8),
                            statusChip(order.status),
                          ],
                        ),
                        subtitle: Text(
                          [
                            if (order.expectedDeliveryDate.isNotEmpty)
                              '交期 ${formatDate(order.expectedDeliveryDate)}',
                            '建單 ${formatDateTime(order.createdAt)}',
                            if (order.note.isNotEmpty) order.note,
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
                        onTap: () => pushAdaptive(
                          context,
                          OrderDetailPage(api: api, orderId: order.id),
                        ),
                      );
                    },
                  ),
          ),
          if (total > _pageSize)
            Container(
              padding: const EdgeInsets.symmetric(vertical: 4),
              decoration: BoxDecoration(
                border: Border(
                  top: BorderSide(
                    color: CupertinoColors.separator,
                    width: Theme.of(context).dividerTheme.thickness ?? 0.5,
                  ),
                ),
              ),
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
