import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/salesorder.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';
import 'orders_page.dart' show eventLabel;

/// 訂單詳情：主檔＋明細＋事件軌跡；pending 可取消、processing 可完成。
class OrderDetailPage extends HookWidget {
  const OrderDetailPage({super.key, required this.api, required this.orderId});

  final Api api;
  final String orderId;

  @override
  Widget build(BuildContext context) {
    final orderQuery = useQuery<GetOrderResponse, Exception>(
      ['order', orderId],
      () => api.orders.getOrder(GetOrderRequest(id: orderId)),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );
    final eventsQuery = useQuery<ListOrderEventsResponse, Exception>(
      ['order-events', orderId],
      () => api.orders.listOrderEvents(
          ListOrderEventsRequest(salesOrderId: orderId, page: 1, pageSize: 100)),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );

    final busy = useState(false);

    Future<void> runAction(
      String label,
      Future<SalesOrder> Function() action,
    ) async {
      final ok = await confirmAdaptive(
        context,
        title: '$label訂單',
        message: '訂單 ${orderQuery.data?.order.orderNo ?? ''} 將被$label，確定？',
        confirmLabel: label,
        destructive: label == '取消',
      );
      if (!ok) return;
      busy.value = true;
      try {
        await action();
        if (context.mounted) {
          ScaffoldMessenger.maybeOf(context)
              ?.showSnackBar(SnackBar(content: Text('已$label')));
        }
        await orderQuery.refetch();
        await eventsQuery.refetch();
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.maybeOf(context)
              ?.showSnackBar(SnackBar(content: Text('失敗：$e')));
        }
      } finally {
        busy.value = false;
      }
    }

    final order = orderQuery.data?.order;
    final items = orderQuery.data?.items ?? const <SalesOrderItem>[];
    final events = eventsQuery.data?.events ?? const <SalesOrderEvent>[];

    return adaptivePage(
      context,
      title: '訂單詳情',
      body: orderQuery.isError
          ? Center(child: Text('載入失敗：${orderQuery.error}'))
          : order == null
              ? const Center(child: CircularProgressIndicator())
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(order.orderNo,
                              style: Theme.of(context).textTheme.titleLarge),
                        ),
                        statusChip(order.status),
                      ],
                    ),
                    const SizedBox(height: 8),
                    _kv('客戶', order.customerId.isEmpty ? '—' : '#${order.customerId}'),
                    _kv('來源', order.source),
                    if (order.expectedDeliveryDate.isNotEmpty)
                      _kv('預定交期', formatDate(order.expectedDeliveryDate)),
                    _kv('建單時間', formatDateTime(order.createdAt)),
                    if (order.note.isNotEmpty) _kv('備註', order.note),
                    _kv('版本', '${order.version}'),
                    const SizedBox(height: 16),
                    Text('明細（${items.length}）',
                        style: Theme.of(context).textTheme.titleMedium),
                    for (final item in items)
                      adaptiveListTile(
                        title: Text(item.displayName),
                        subtitle: Text(
                          [
                            '${item.qty} ${item.unit}'
                                '${item.baseQty.isEmpty ? '' : '（基數 ${item.baseQty}）'}',
                            if (item.specialCutNote.isNotEmpty)
                              '分切備註：${item.specialCutNote}',
                          ].join('\n'),
                        ),
                      ),
                    const SizedBox(height: 16),
                    Text('事件軌跡（${events.length}）',
                        style: Theme.of(context).textTheme.titleMedium),
                    for (final event in events)
                      adaptiveListTile(
                        title: Text(eventLabel(event.eventType)),
                        subtitle: Text(
                          [
                            formatDateTime(event.createdAt),
                            if (event.actorId.isNotEmpty) '操作者 #${event.actorId}',
                            if (event.reason.isNotEmpty) '原因：${event.reason}',
                          ].join(' · '),
                        ),
                      ),
                    const SizedBox(height: 24),
                    if (order.status == 'pending')
                      adaptiveFilledButton(
                        onPressed:
                            busy.value ? null : () => runAction('取消', () async {
                                  final r = await api.orders
                                      .cancelOrder(CancelOrderRequest(id: orderId));
                                  return r.order;
                                }),
                        child: const Text('取消訂單'),
                      ),
                    if (order.status == 'processing') ...[
                      const SizedBox(height: 8),
                      adaptiveFilledButton(
                        onPressed: busy.value
                            ? null
                            : () => runAction('完成', () async {
                                  final r = await api.orders
                                      .completeOrder(CompleteOrderRequest(id: orderId));
                                  return r.order;
                                }),
                        child: const Text('完成訂單'),
                      ),
                    ],
                    if (busy.value) ...[
                      const SizedBox(height: 12),
                      const Center(child: CircularProgressIndicator()),
                    ],
                  ],
                ),
    );
  }

  Widget _kv(String key, String value) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 3),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 84,
              child: Text(key,
                  style: const TextStyle(color: CupertinoColors.inactiveGray)),
            ),
            Expanded(child: Text(value)),
          ],
        ),
      );
}
