import 'dart:async';

import 'package:flutter/cupertino.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

/// 平台自適應 UI 入口（Android → Material、iOS → Cupertino）。
///
/// 判定用 [defaultTargetPlatform] 而非 dart:io Platform：測試可透過
/// debugDefaultTargetPlatformOverride 切換，同一份頁面碼兩平台共用。
bool isCupertinoTarget() =>
    !kIsWeb && defaultTargetPlatform == TargetPlatform.iOS;

Color _statusColor(String status) => switch (status) {
      'completed' || 'approved' => const Color(0xFF2E7D32),
      'processing' => const Color(0xFF1565C0),
      'pending' => const Color(0xFFEF6C00),
      'cancelled' || 'rejected' => const Color(0xFFC62828),
      'voided' || 'failed' => const Color(0xFF6A1B9A),
      _ => const Color(0xFF616161),
    };

const _statusLabels = <String, String>{
  'pending': '待處理',
  'processing': '處理中',
  'completed': '已完成',
  'cancelled': '已取消',
  'voided': '已作廢',
  'approved': '已核准',
  'rejected': '已拒絕',
  'read': '已讀',
  'sent': '已發送',
  'failed': '失敗',
};

/// 篩選 chip：ChoiceChip / 版型化 CupertinoButton。
Widget adaptiveFilterChip({
  required String label,
  required bool selected,
  VoidCallback? onTap,
}) =>
    isCupertinoTarget()
        ? CupertinoButton(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            borderRadius: BorderRadius.circular(16),
            color: selected ? CupertinoColors.activeBlue : CupertinoColors.systemGrey5,
            onPressed: onTap,
            child: Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: selected ? FontWeight.w600 : FontWeight.w400,
                color: selected ? CupertinoColors.white : CupertinoColors.label,
              ),
            ),
          )
        : ChoiceChip(
            label: Text(label),
            selected: selected,
            onSelected: (_) => onTap?.call(),
          );

/// 狀態徽章：兩平台同一顆視覺（狀態語意不隨平台變）。
Widget statusChip(String status) {
  final color = _statusColor(status);
  return Container(
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
    decoration: BoxDecoration(
      color: color.withValues(alpha: 0.12),
      borderRadius: BorderRadius.circular(999),
      border: Border.all(color: color.withValues(alpha: 0.5)),
    ),
    child: Text(
      _statusLabels[status] ?? status,
      style: TextStyle(fontSize: 12, color: color, fontWeight: FontWeight.w600),
    ),
  );
}

/// 頁面骨架：iOS CupertinoPageScaffold+CupertinoNavigationBar、
/// Android Scaffold+AppBar。
Widget adaptivePage(
  BuildContext context, {
  required String title,
  required Widget body,
  List<Widget>? actions,
  Widget? trailing,
  Widget? leading,
}) {
  if (isCupertinoTarget()) {
    return CupertinoPageScaffold(
      navigationBar: CupertinoNavigationBar(
        middle: Text(title),
        trailing: trailing,
        leading: leading,
        automaticBackgroundVisibility: false,
      ),
      // 透明 Material 襯底：頁內仍可能有 Material 控制項（TextFormField 等,
      // 混用是刻意的 v1 折衷——它們要求 Material 祖先,而 CupertinoPageScaffold 沒有）。
      child: Material(
        type: MaterialType.transparency,
        child: SafeArea(child: body),
      ),
    );
  }
  return Scaffold(
    appBar: AppBar(
      title: Text(title),
      actions: actions,
      leading: leading,
    ),
    body: body,
  );
}

/// 主行動按鈕：FilledButton / CupertinoButton.filled。
Widget adaptiveFilledButton({
  required VoidCallback? onPressed,
  required Widget child,
}) =>
    isCupertinoTarget()
        ? CupertinoButton(onPressed: onPressed, child: child)
        : FilledButton(onPressed: onPressed, child: child);

/// 次要按鈕：TextButton / CupertinoButton；destructive 走錯誤色。
class AdaptiveButton extends StatelessWidget {
  const AdaptiveButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.destructive = false,
  });

  final VoidCallback? onPressed;
  final Widget child;
  final bool destructive;

  @override
  Widget build(BuildContext context) {
    if (isCupertinoTarget()) {
      return CupertinoButton(
        onPressed: onPressed,
        child: DefaultTextStyle(
          style: TextStyle(
            color: destructive
                ? CupertinoColors.destructiveRed
                : CupertinoColors.activeBlue,
            fontSize: 15,
          ),
          child: child,
        ),
      );
    }
    return TextButton(
      onPressed: onPressed,
      child: DefaultTextStyle(
        style: TextStyle(
          color: destructive ? Theme.of(context).colorScheme.error : null,
        ),
        child: child,
      ),
    );
  }
}

/// 清單列：CupertinoListTile / ListTile。
Widget adaptiveListTile({
  required Widget title,
  Widget? subtitle,
  Widget? trailing,
  VoidCallback? onTap,
  bool isSelected = false,
}) =>
    isCupertinoTarget()
        ? CupertinoListTile.notched(
            title: title,
            subtitle: subtitle,
            trailing: trailing,
            onTap: onTap,
          )
        : ListTile(
            title: title,
            subtitle: subtitle,
            trailing: trailing,
            onTap: onTap,
            selected: isSelected,
          );

/// 可下拉更新清單：iOS CupertinoSliverRefreshControl、Android RefreshIndicator。
Widget adaptiveRefreshList({
  required Future<void> Function() onRefresh,
  required IndexedWidgetBuilder itemBuilder,
  required int itemCount,
  Widget? empty,
  ScrollController? controller,
}) {
  if (isCupertinoTarget()) {
    return CustomScrollView(
      controller: controller,
      physics: const AlwaysScrollableScrollPhysics(),
      slivers: [
        CupertinoSliverRefreshControl(onRefresh: onRefresh),
        if (itemCount == 0 && empty != null)
          SliverToBoxAdapter(child: empty)
        else
          SliverList(
            delegate: SliverChildBuilderDelegate(
              (context, index) => itemBuilder(context, index),
              childCount: itemCount,
            ),
          ),
      ],
    );
  }
  return RefreshIndicator(
    onRefresh: onRefresh,
    child: itemCount == 0
        ? ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            children: [?empty],
          )
        : ListView.builder(
            controller: controller,
            physics: const AlwaysScrollableScrollPhysics(),
            itemCount: itemCount,
            itemBuilder: itemBuilder,
          ),
  );
}

/// 確認對話框：CupertinoAlertDialog / AlertDialog。
Future<bool> confirmAdaptive(
  BuildContext context, {
  required String title,
  String? message,
  String confirmLabel = '確定',
  String cancelLabel = '取消',
  bool destructive = false,
}) async {
  final result = isCupertinoTarget()
      ? await showCupertinoDialog<bool>(
          context: context,
          builder: (context) => CupertinoAlertDialog(
            title: Text(title),
            content: message == null ? null : Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(message),
            ),
            actions: [
              CupertinoDialogAction(
                child: Text(cancelLabel),
                onPressed: () => Navigator.of(context).pop(false),
              ),
              CupertinoDialogAction(
                isDestructiveAction: destructive,
                isDefaultAction: true,
                child: Text(confirmLabel),
                onPressed: () => Navigator.of(context).pop(true),
              ),
            ],
          ),
        )
      : await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text(title),
            content: message == null ? null : Text(message),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(false),
                child: Text(cancelLabel),
              ),
              TextButton(
                onPressed: () => Navigator.of(context).pop(true),
                child: Text(
                  confirmLabel,
                  style: destructive
                      ? TextStyle(color: Theme.of(context).colorScheme.error)
                      : null,
                ),
              ),
            ],
          ),
        );
  return result ?? false;
}

/// 單行輸入對話框（原因/備註）：取消回 null、確認回輸入文字（可為空字串）。
///
/// controller 由**對話框自己**持有（[_PromptDialog] 的 State），不是外層函式建立後
/// `finally` 丟掉 —— `showDialog` 的 Future 在 `pop()` 當下就 resolve，而對話框此時還在
/// 播退場動畫並持續重建；外層若在此時 dispose controller，動畫期間就會擲
/// 「A TextEditingController was used after being disposed」（Android 實機可見）。
/// 擁有權跟著 widget 生命週期，動畫結束才釋放。
Future<String?> promptAdaptive(
  BuildContext context, {
  required String title,
  String? message,
  String confirmLabel = '確定',
  String initialText = '',
  String? hint,
  bool requireText = false,
}) {
  return showAdaptiveDialog<String>(
    context: context,
    builder: (context) => _PromptDialog(
      title: title,
      message: message,
      confirmLabel: confirmLabel,
      initialText: initialText,
      hint: hint,
      requireText: requireText,
    ),
  );
}

/// 依平台選對話框外殼（Cupertino / Material）。
Future<T?> showAdaptiveDialog<T>({
  required BuildContext context,
  required WidgetBuilder builder,
}) =>
    isCupertinoTarget()
        ? showCupertinoDialog<T>(context: context, builder: builder)
        : showDialog<T>(context: context, builder: builder);

/// 輸入對話框（兩平台共用同一份 State 邏輯，只有外觀不同）。
class _PromptDialog extends StatefulWidget {
  const _PromptDialog({
    required this.title,
    required this.message,
    required this.confirmLabel,
    required this.initialText,
    required this.hint,
    required this.requireText,
  });

  final String title;
  final String? message;
  final String confirmLabel;
  final String initialText;
  final String? hint;
  final bool requireText;

  @override
  State<_PromptDialog> createState() => _PromptDialogState();
}

class _PromptDialogState extends State<_PromptDialog> {
  late final TextEditingController _controller =
      TextEditingController(text: widget.initialText);

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  /// 確認：`requireText` 且空白時**不關閉**（讓使用者繼續輸入，而不是把空字串當成答案）。
  void _confirm() {
    final text = _controller.text.trim();
    if (widget.requireText && text.isEmpty) return;
    Navigator.of(context).pop(text);
  }

  @override
  Widget build(BuildContext context) {
    if (isCupertinoTarget()) {
      return CupertinoAlertDialog(
        title: Text(widget.title),
        content: Column(
          children: [
            if (widget.message != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(widget.message!),
              ),
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: CupertinoTextField(
                controller: _controller,
                placeholder: widget.hint,
                autofocus: true,
                onSubmitted: (_) => _confirm(),
              ),
            ),
          ],
        ),
        actions: [
          CupertinoDialogAction(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('取消'),
          ),
          CupertinoDialogAction(
            isDefaultAction: true,
            onPressed: _confirm,
            child: Text(widget.confirmLabel),
          ),
        ],
      );
    }
    return AlertDialog(
      title: Text(widget.title),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (widget.message != null) Text(widget.message!),
          TextField(
            controller: _controller,
            decoration: InputDecoration(hintText: widget.hint),
            autofocus: true,
            onSubmitted: (_) => _confirm(),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('取消'),
        ),
        TextButton(onPressed: _confirm, child: Text(widget.confirmLabel)),
      ],
    );
  }
}

/// 單確認內容對話框（通知全文等）：CupertinoAlertDialog / AlertDialog。
Future<void> showModalAdaptive(
  BuildContext context, {
  required String title,
  String? message,
  String confirmLabel = '確定',
}) async {
  if (isCupertinoTarget()) {
    await showCupertinoDialog<void>(
      context: context,
      builder: (context) => CupertinoAlertDialog(
        title: Text(title),
        content: message == null
            ? null
            : Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(message, textAlign: TextAlign.start),
              ),
        actions: [
          CupertinoDialogAction(
            isDefaultAction: true,
            child: Text(confirmLabel),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
    return;
  }
  await showDialog<void>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text(title),
      content: message == null ? null : Text(message),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(confirmLabel),
        ),
      ],
    ),
  );
}

/// 輕量回饋訊息：Android → SnackBar；iOS → alert（ CupertinoScaffold 無
/// SnackBar 所需的 Scaffold,Messenger 會靜默吞掉 —— 導致失敗看似無反應）。
void showFeedback(BuildContext context, String message) {
  if (isCupertinoTarget()) {
    unawaited(showModalAdaptive(context, title: '提示', message: message));
    return;
  }
  ScaffoldMessenger.maybeOf(context)?.showSnackBar(SnackBar(content: Text(message)));
}

/// 依平台推入子頁：CupertinoPageRoute（iOS）／MaterialPageRoute（Android）。
Future<T?> pushAdaptive<T>(BuildContext context, Widget page) =>
    Navigator.of(context).push<T>(
      isCupertinoTarget()
          ? CupertinoPageRoute(builder: (_) => page)
          : MaterialPageRoute(builder: (_) => page),
    );
