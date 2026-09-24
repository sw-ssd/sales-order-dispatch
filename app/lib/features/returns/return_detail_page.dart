import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/returns.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';

/// 退貨詳情：明細＋狀態；pending 可審核（核准/拒絕，帶 Get 回填的 version），
/// approved 可看退貨證明。回傳 true 表示有異動（審核成功），供清單重查。
class ReturnDetailPage extends HookWidget {
  const ReturnDetailPage({super.key, required this.api, required this.returnId});

  final Api api;
  final String returnId;

  @override
  Widget build(BuildContext context) {
    final query = useQuery<GetReturnRequestResponse, Exception>(
      ['return', returnId],
      () => api.returns.getReturnRequest(GetReturnRequestRequest(id: returnId)),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );
    final data = query.data;

    final busy = useState(false);

    Future<void> review(String decision, {String rejectReason = ''}) async {
      busy.value = true;
      try {
        await api.returns.reviewReturnRequest(ReviewReturnRequestRequest(
          id: returnId,
          decision: decision,
          rejectReason: rejectReason,
          // expectedVersion 只能取自 Get 回填（後端遞增；寫死 0 版本一變就永遠審不了）。
          expectedVersion: data?.version ?? '',
        ));
        await query.refetch();
        if (!context.mounted) return;
        ScaffoldMessenger.maybeOf(context)?.showSnackBar(
          SnackBar(content: Text(decision == 'approved' ? '已核准' : '已拒絕')),
        );
        Navigator.of(context).pop(true);
      } catch (e) {
        busy.value = false;
        if (context.mounted) {
          ScaffoldMessenger.maybeOf(context)
              ?.showSnackBar(SnackBar(content: Text('審核失敗：$e')));
        }
      }
    }

    return adaptivePage(
      context,
      title: '退貨詳情',
      body: query.isError
          ? Center(child: Text('載入失敗：${query.error}'))
          : data == null
              ? const Center(child: CircularProgressIndicator())
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text('退貨 #${data.id}',
                              style: Theme.of(context).textTheme.titleLarge),
                        ),
                        statusChip(data.status),
                      ],
                    ),
                    const SizedBox(height: 8),
                    if (data.remark.isNotEmpty) _kv('備註', data.remark),
                    if (data.rejectReason.isNotEmpty)
                      _kv('拒絕原因', data.rejectReason),
                    _kv('申請時間', formatDateTime(data.createdAt)),
                    _kv('版本', data.version),
                    const SizedBox(height: 16),
                    Text('品項（${data.items.length}）',
                        style: Theme.of(context).textTheme.titleMedium),
                    for (final item in data.items)
                      adaptiveListTile(
                        title: Text(item.productName),
                        subtitle: Text(
                          [
                            '${item.quantity} ${item.unit}'
                                '${item.spec.isEmpty ? '' : '（${item.spec}）'}',
                            '退貨原因：${item.reason}',
                            if (item.photoUrls.isNotEmpty)
                              '照片 ${item.photoUrls.length} 張',
                          ].join('\n'),
                        ),
                      ),
                    if (data.status == 'pending') ...[
                      const SizedBox(height: 24),
                      adaptiveFilledButton(
                        onPressed: busy.value
                            ? null
                            : () async {
                                final ok = await confirmAdaptive(
                                  context,
                                  title: '核准退貨',
                                  message: '確定核准此退貨申請？',
                                  confirmLabel: '核准',
                                );
                                if (ok && context.mounted) {
                                  await review('approved');
                                }
                              },
                        child: const Text('核准'),
                      ),
                      const SizedBox(height: 8),
                      AdaptiveButton(
                        destructive: true,
                        onPressed: busy.value
                            ? null
                            : () async {
                                final reason = await promptAdaptive(
                                  context,
                                  title: '拒絕退貨',
                                  message: '拒絕原因（必填）',
                                  confirmLabel: '拒絕',
                                  requireText: true,
                                );
                                if (reason != null && reason.isNotEmpty &&
                                    context.mounted) {
                                  await review('rejected',
                                      rejectReason: reason);
                                }
                              },
                        child: const Text('拒絕'),
                      ),
                    ],
                    if (data.status == 'approved')
                      _CertificateSection(api: api, returnId: returnId),
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

/// 退貨證明（approved 時才查詢；快照唯讀）。
class _CertificateSection extends HookWidget {
  const _CertificateSection({required this.api, required this.returnId});

  final Api api;
  final String returnId;

  @override
  Widget build(BuildContext context) {
    final query = useQuery<GetReturnCertificateResponse, Exception>(
      ['return-cert', returnId],
      () => api.returns
          .getReturnCertificate(GetReturnCertificateRequest(id: returnId)),
      context: context,
    );

    final cert = query.data;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),
        Text('退貨證明', style: Theme.of(context).textTheme.titleMedium),
        if (query.isLoading)
          const Padding(
            padding: EdgeInsets.only(top: 12),
            child: Center(child: CircularProgressIndicator()),
          )
        else if (query.isError)
          Text('證明載入失敗：${query.error}')
        else if (cert != null) ...[
          _kv('客戶', '${cert.customerName}（${cert.customerCode}）'),
          _kv('審核人', cert.reviewerName),
          _kv('審核時間', formatDateTime(cert.reviewedAt)),
          Text('品項 ${cert.items.length} 項',
              style: Theme.of(context).textTheme.bodyMedium),
        ],
      ],
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
