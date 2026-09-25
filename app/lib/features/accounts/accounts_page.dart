import 'package:auto_route/auto_route.dart';
import 'package:connectrpc/connect.dart' as connect;
import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';

import '../../core/api.dart';
import '../../gen/customers/v1/customer.pb.dart';
import '../../ui/adaptive.dart';
import '../../ui/format.dart';
import '../auth/auth_repository.dart';

/// 帳號管理頁（D22／規格 4.2，Task 6.7）：主帳號自助管理自己客戶底下的子帳號。
///
/// 只列不篩：一個店家客戶的帳號數量是個位數（主帳號 1 + 業務子帳號 1 + 自建數個），
/// 加篩選或分頁只是把後端 `ListCustomerAccounts` 的單次全量查詢繞遠路。
///
/// 三類帳號在 UI 上必須**看得出差別**，因為可做的事不同（後端也各自把關）：
/// - 主帳號（`isPrimary`）：帳號體系的管理者，但不可自停／自重置（避免把店家鎖在門外）。
/// - 業務子帳號（`systemGenerated`）：建檔時自動附帶、專供所屬業務使用 —— 店家**沒有**
///   它的密碼，故不可改名／停用／重置。標為「系統預設（業務使用）」並灰化。
/// - 自建子帳號：唯一可停用／重置的對象。
class AccountsPage extends HookWidget {
  const AccountsPage({
    super.key,
    required this.api,
    required this.auth,
    this.onLoggedOut,
  });

  final Api api;

  /// 供登出用。主帳號登入後唯一的落點就是本頁，沒有登出等於無路可退。
  final AuthRepository auth;

  /// 登出後的回呼（預設回 `/login`）。
  final VoidCallback? onLoggedOut;

  @override
  Widget build(BuildContext context) {
    final query = useQuery<ListCustomerAccountsResponse, Exception>(
      ['customer_accounts'],
      () => api.accounts.listCustomerAccounts(ListCustomerAccountsRequest()),
      context: context,
      refetchOnMount: RefetchOnMount.always,
    );

    Future<void> createAccount() async {
      final name = await promptAdaptive(
        context,
        title: '新增子帳號',
        message: '子帳號以帳號名稱登入，首次登入須改密碼。',
        hint: '帳號名稱',
        confirmLabel: '建立',
        requireText: true,
      );
      if (name == null || name.isEmpty || !context.mounted) return;
      try {
        final res = await api.accounts.createCustomerAccount(
          CreateCustomerAccountRequest(accountName: name),
        );
        if (!context.mounted) return;
        await _showTempPassword(
          context,
          title: '已建立「${res.account.accountName}」',
          password: res.tempPassword,
          expiresAt: res.tempExpiresAt,
        );
        query.refetch();
      } catch (e) {
        if (!context.mounted) return;
        showFeedback(context, accountErrorMessage(e));
      }
    }

    Future<void> resetPassword(CustomerAccount account) async {
      final ok = await confirmAdaptive(
        context,
        title: '重置「${account.accountName}」的密碼？',
        message: '將核發新的臨時密碼，該帳號目前的登入狀態會立即失效。',
        confirmLabel: '重置',
      );
      if (!ok || !context.mounted) return;
      try {
        final res = await api.accounts.resetCustomerAccountPassword(
          ResetCustomerAccountPasswordRequest(accountId: account.id),
        );
        if (!context.mounted) return;
        await _showTempPassword(
          context,
          title: '「${account.accountName}」的新臨時密碼',
          password: res.tempPassword,
          expiresAt: res.tempExpiresAt,
        );
        query.refetch();
      } catch (e) {
        if (!context.mounted) return;
        showFeedback(context, accountErrorMessage(e));
      }
    }

    Future<void> deactivate(CustomerAccount account) async {
      final ok = await confirmAdaptive(
        context,
        title: '停用「${account.accountName}」？',
        message: '停用後該帳號無法登入，已登入的裝置會立即失效。',
        confirmLabel: '停用',
        destructive: true,
      );
      if (!ok || !context.mounted) return;
      try {
        await api.accounts.deactivateCustomerAccount(
          DeactivateCustomerAccountRequest(accountId: account.id),
        );
        query.refetch();
      } catch (e) {
        if (!context.mounted) return;
        showFeedback(context, accountErrorMessage(e));
      }
    }

    Future<void> logout() async {
      final ok = await confirmAdaptive(
        context,
        title: '登出？',
        confirmLabel: '登出',
        destructive: true,
      );
      if (!ok || !context.mounted) return;
      await auth.logout();
      if (!context.mounted) return;
      if (onLoggedOut != null) {
        onLoggedOut!();
      } else {
        context.router.navigatePath('/login');
      }
    }

    final accounts = query.data?.accounts ?? const <CustomerAccount>[];

    return adaptivePage(
      context,
      title: '帳號管理',
      // 導覽列：iOS 用文字鈕（新增）＋左側登出；Android 用圖示鈕。
      leading: isCupertinoTarget()
          ? CupertinoButton(
              padding: EdgeInsets.zero,
              onPressed: logout,
              child: const Text('登出'),
            )
          : null,
      trailing: isCupertinoTarget()
          ? CupertinoButton(
              padding: EdgeInsets.zero,
              onPressed: createAccount,
              child: const Text('新增'),
            )
          : null,
      actions: isCupertinoTarget()
          ? null
          : [
              IconButton(
                onPressed: createAccount,
                icon: const Icon(Icons.add),
                tooltip: '新增子帳號',
              ),
              IconButton(
                onPressed: logout,
                icon: const Icon(Icons.logout),
                tooltip: '登出',
              ),
            ],
      body: query.isError
          ? Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(accountErrorMessage(query.error!)),
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
              itemCount: accounts.length,
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
                        child: Text('沒有帳號'),
                      ),
                    ),
              itemBuilder: (context, index) {
                final account = accounts[index];
                return adaptiveListTile(
                  title: Text(
                    account.accountName,
                    style: account.manageable
                        ? null
                        : const TextStyle(color: Color(0xFF9E9E9E)),
                  ),
                  subtitle: Text(
                    _accountSubtitle(account),
                    style: account.manageable
                        ? null
                        : const TextStyle(color: Color(0xFF9E9E9E)),
                  ),
                  trailing: account.manageable
                      ? _AccountActions(
                          onReset: () => resetPassword(account),
                          onDeactivate: () => deactivate(account),
                        )
                      : null,
                );
              },
            ),
    );
  }
}

/// 帳號的副標：型別標記 + 狀態。
///
/// 業務子帳號必須明示「系統預設（業務使用）」—— 店家看到一個自己沒建過、又不能停用的
/// 帳號時，不解釋就只會像是系統壞了。
String _accountSubtitle(CustomerAccount account) {
  final parts = <String>[
    if (account.isPrimary) '主帳號',

    if (account.systemGenerated) '系統預設（業務使用）',
    if (!account.manageable && !account.isPrimary && !account.systemGenerated)
      '不可管理',
    account.status == 'active' ? '啟用中' : '已停用',
  ];
  return parts.join('・');
}

/// 列尾動作：重置密碼／停用。
class _AccountActions extends StatelessWidget {
  const _AccountActions({required this.onReset, required this.onDeactivate});

  final VoidCallback onReset;
  final VoidCallback onDeactivate;

  @override
  Widget build(BuildContext context) {
    if (isCupertinoTarget()) {
      return CupertinoButton(
        padding: EdgeInsets.zero,
        onPressed: () => _showSheet(context),
        child: const Icon(CupertinoIcons.ellipsis),
      );
    }
    return PopupMenuButton<String>(
      onSelected: (value) =>
          value == 'reset' ? onReset() : onDeactivate(),
      itemBuilder: (context) => const [
        PopupMenuItem(value: 'reset', child: Text('重置密碼')),
        PopupMenuItem(value: 'deactivate', child: Text('停用帳號')),
      ],
    );
  }

  void _showSheet(BuildContext context) {
    showCupertinoModalPopup<void>(
      context: context,
      builder: (context) => CupertinoActionSheet(
        actions: [
          CupertinoActionSheetAction(
            onPressed: () {
              Navigator.of(context).pop();
              onReset();
            },
            child: const Text('重置密碼'),
          ),
          CupertinoActionSheetAction(
            isDestructiveAction: true,
            onPressed: () {
              Navigator.of(context).pop();
              onDeactivate();
            },
            child: const Text('停用帳號'),
          ),
        ],
        cancelButton: CupertinoActionSheetAction(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('取消'),
        ),
      ),
    );
  }
}

/// 顯示臨時密碼（僅此一次；系統不留明文）。
Future<void> _showTempPassword(
  BuildContext context, {
  required String title,
  required String password,
  required String expiresAt,
}) =>
    showModalAdaptive(
      context,
      title: title,
      message: '臨時密碼：$password\n'
          '有效至：${formatDateTime(expiresAt)}\n'
          '請立即轉交，關閉後無法再次查看。',
    );

/// 帳號管理錯誤訊息（繁體中文）。
String accountErrorMessage(Object error) {
  if (error is connect.ConnectException) {
    return switch (error.code) {
      connect.Code.permissionDenied => '此帳號沒有帳號管理權限（僅店家主帳號可管理）',
      connect.Code.unauthenticated => '登入狀態已失效，請重新登入',
      connect.Code.invalidArgument => '操作不允許：此帳號不可停用或重置',
      connect.Code.alreadyExists => '已有同名帳號，請換一個名稱',
      connect.Code.unavailable => '無法連線至伺服器，請檢查網路後再試',
      _ => '操作失敗（${error.code.name}），請稍後再試',
    };
  }
  return '操作失敗，請稍後再試';
}
