import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/returns.pb.dart';
import '../../gen/salesorder/v1/salesorder.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';

/// 發起退貨（僅客戶子帳號，後端把關）：選原訂單 → 勾品項填數量/原因 → 送出。
/// 回傳 true 表示已建立（清單需重查）。
///
/// v1 只走 order_item 來源；照片上傳（photo_file_ids）另案。
class ReturnCreatePage extends StatefulWidget {
  const ReturnCreatePage({super.key, required this.api});

  final Api api;

  @override
  State<ReturnCreatePage> createState() => _ReturnCreatePageState();
}

class _ReturnCreatePageState extends State<ReturnCreatePage> {
  SalesOrder? _order;
  List<SalesOrderItem> _items = const [];
  bool _loadingItems = false;
  String? _itemsError;

  final _selected = <String>{};
  final _qty = <String, TextEditingController>{};
  final _reason = <String, TextEditingController>{};
  final _remark = TextEditingController();
  bool _submitting = false;

  @override
  void dispose() {
    for (final c in [
      ..._qty.values,
      ..._reason.values,
      _remark,
    ]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _pickOrder() async {
    final order = await pushAdaptive<SalesOrder>(
      context,
      _OrderPickerPage(api: widget.api),
    );
    if (order == null || !mounted) return;
    setState(() {
      _order = order;
      _items = const [];
      _itemsError = null;
      _loadingItems = true;
      _selected.clear();
    });
    try {
      final r = await widget.api.orders
          .getOrder(GetOrderRequest(id: order.id));
      if (!mounted) return;
      setState(() => _items = r.items);
    } catch (e) {
      if (!mounted) return;
      setState(() => _itemsError = '$e');
    } finally {
      if (mounted) setState(() => _loadingItems = false);
    }
  }

  Future<void> _submit() async {
    final selected = _selected.toList(growable: false);
    if (selected.isEmpty) {
      _snack('請至少勾選一個品項');
      return;
    }
    final items = <ReturnItemInput>[];
    for (final id in selected) {
      final qty = (_qty[id]?.text ?? '').trim();
      final reason = (_reason[id]?.text ?? '').trim();
      final value = double.tryParse(qty);
      if (value == null || value <= 0) {
        _snack('品項數量須為正數');
        return;
      }
      if (reason.isEmpty) {
        _snack('每個品項都要填退貨原因');
        return;
      }
      items.add(ReturnItemInput(
        sourceType: 'order_item',
        salesOrderItemId: id,
        quantity: qty,
        reason: reason,
      ));
    }
    setState(() => _submitting = true);
    try {
      await widget.api.returns.createReturnRequest(CreateReturnRequestRequest(
        items: items,
        remark: _remark.text.trim(),
      ));
      if (mounted) Navigator.of(context).pop(true);
    } catch (e) {
      if (!mounted) return;
      setState(() => _submitting = false);
      _snack('送出失敗：$e');
    }
  }

  void _snack(String message) => showFeedback(context, message);

  @override
  Widget build(BuildContext context) {
    return adaptivePage(
      context,
      title: '發起退貨',
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text('原訂單', style: Theme.of(context).textTheme.titleMedium),
          adaptiveListTile(
            title: Text(_order?.orderNo ?? '選擇訂單'),
            subtitle: _order == null
                ? const Text('點此選擇要退貨的訂單')
                : Text('建單 ${formatDateTime(_order!.createdAt)}'),
            trailing: Icon(
              isCupertinoTarget()
                  ? CupertinoIcons.chevron_forward
                  : Icons.chevron_right,
              size: 18,
            ),
            onTap: _pickOrder,
          ),
          if (_loadingItems)
            const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator()),
            )
          else if (_itemsError != null)
            Padding(
              padding: const EdgeInsets.all(16),
              child: Text('品項載入失敗：$_itemsError'),
            )
          else if (_order != null) ...[
            const SizedBox(height: 8),
            Text('品項（勾選要退的）',
                style: Theme.of(context).textTheme.titleMedium),
            for (final item in _items)
              _ItemRow(
                item: item,
                selected: _selected.contains(item.id),
                onToggle: () {
                  setState(() {
                    if (!_selected.add(item.id)) {
                      _selected.remove(item.id);
                    }
                  });
                },
                qty: _qty.putIfAbsent(
                    item.id, () => TextEditingController(text: '1')),
                reason: _reason.putIfAbsent(
                    item.id, () => TextEditingController()),
              ),
          ],
          const SizedBox(height: 16),
          Text('備註', style: Theme.of(context).textTheme.titleMedium),
          TextField(
            controller: _remark,
            maxLines: 3,
            decoration: const InputDecoration(hintText: '整單退貨說明（可留空）'),
          ),
          const SizedBox(height: 24),
          adaptiveFilledButton(
            onPressed: _submitting ? null : _submit,
            child: const Text('送出退貨申請'),
          ),
          if (_submitting)
            const Padding(
              padding: EdgeInsets.only(top: 12),
              child: Center(child: CircularProgressIndicator()),
            ),
        ],
      ),
    );
  }
}

/// 品項列：勾選＋數量＋退貨原因。
class _ItemRow extends StatelessWidget {
  const _ItemRow({
    required this.item,
    required this.selected,
    required this.onToggle,
    required this.qty,
    required this.reason,
  });

  final SalesOrderItem item;
  final bool selected;
  final VoidCallback onToggle;
  final TextEditingController qty;
  final TextEditingController reason;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        adaptiveListTile(
          title: Text(item.displayName),
          subtitle: Text(
            '${item.qty} ${item.unit}'
            '${item.specialCutNote.isEmpty ? '' : ' · ${item.specialCutNote}'}',
          ),
          onTap: onToggle,
          trailing: Icon(
            selected
                ? (isCupertinoTarget()
                    ? CupertinoIcons.checkmark_circle_fill
                    : Icons.check_circle)
                : (isCupertinoTarget()
                    ? CupertinoIcons.circle
                    : Icons.circle_outlined),
            color: selected ? const Color(0xFF2E7D32) : CupertinoColors.inactiveGray,
          ),
        ),
        if (selected)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: qty,
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                    decoration: const InputDecoration(
                      labelText: '退貨數量',
                      isDense: true,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  flex: 2,
                  child: TextField(
                    controller: reason,
                    decoration: const InputDecoration(
                      labelText: '退貨原因（必填）',
                      isDense: true,
                    ),
                  ),
                ),
              ],
            ),
          ),
        const SizedBox(height: 4),
      ],
    );
  }
}

/// 訂單挑選器：列出可見訂單，選回上一頁。
class _OrderPickerPage extends HookWidget {
  const _OrderPickerPage({required this.api});

  final Api api;

  @override
  Widget build(BuildContext context) {
    final query = useQuery<ListOrdersResponse, Exception>(
      ['return-order-picker'],
      () => api.orders
          .listOrders(ListOrdersRequest(page: 1, pageSize: 50)),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );
    final orders = query.data?.orders ?? const <SalesOrder>[];

    return adaptivePage(
      context,
      title: '選擇訂單',
      body: query.isError
          ? Center(child: Text('載入失敗：${query.error}'))
          : adaptiveRefreshList(
              onRefresh: () async => query.refetch(),
              itemCount: orders.length,
              empty: query.isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : const Center(child: Text('沒有可退貨的訂單')),
              itemBuilder: (context, index) {
                final order = orders[index];
                return adaptiveListTile(
                  title: Row(
                    children: [
                      Text(order.orderNo,
                          style:
                              const TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(width: 8),
                      statusChip(order.status),
                    ],
                  ),
                  subtitle: Text('建單 ${formatDateTime(order.createdAt)}'),
                  onTap: () => Navigator.of(context).pop(order),
                );
              },
            ),
    );
  }
}
