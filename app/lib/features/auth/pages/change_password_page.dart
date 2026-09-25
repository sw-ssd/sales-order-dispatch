import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

import '../../../core/error_info.dart';
import '../../../ui/adaptive.dart';
import '../auth_repository.dart';

/// 改密碼頁（A3 1.5.2；規格「臨時密碼首次登入強制修改」）。
///
/// 兩種進入情境，**同一個頁面**：
/// 1. **首登強制**（`mustChangePassword=true`）：後端 middleware 只放行
///    `ChangePassword`，其他一律 AUTH-3004 —— 使用者除了改密碼無事可做，
///    故此模式不顯示「取消」，只給登出。
/// 2. **主動更改**（從「我的」進入）：顯示取消，改完不必重新登入。
///
/// 為何不叫「重置密碼」：後端 `ChangePassword` 驗的是**舊密碼**（臨時密碼亦以此驗證），
/// 與 `ResetCustomerPassword`（dept_admin 代發臨時密碼）是兩件事。
///
/// 成功後本機 token 已由 `AuthRepository.changePassword` 清除（後端 tv+1 使它失效），
/// 一律導回登入頁以新密碼重登 —— 兩情境都如此，因為 tv+1 對「主動更改」同樣撤銷了
/// 現有 session。
class ChangePasswordPage extends StatefulWidget {
  const ChangePasswordPage({
    super.key,
    required this.authRepository,
    required this.mustChange,
  });

  final AuthRepository authRepository;

  /// true = 首登/臨時密碼受限態（後端只放行本 RPC）。
  final bool mustChange;

  @override
  State<ChangePasswordPage> createState() => _ChangePasswordPageState();
}

class _ChangePasswordPageState extends State<ChangePasswordPage> {
  final _formKey = GlobalKey<FormState>();
  final _oldController = TextEditingController();
  final _newController = TextEditingController();
  final _confirmController = TextEditingController();
  bool _busy = false;

  @override
  void dispose() {
    _oldController.dispose();
    _newController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _busy = true);
    try {
      await widget.authRepository.changePassword(
        oldPassword: _oldController.text,
        newPassword: _newController.text,
      );
      if (!mounted) return;
      // token 已被清除：留在原頁只會看到「未登入」。改完一律回登入頁以新密碼重登。
      context.router.navigatePath('/login');
      showFeedback(context, '密碼已更新，請以新密碼重新登入');
    } catch (e) {
      if (!mounted) return;
      showFeedback(context, localizedErrorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return adaptivePage(
      context,
      title: '修改密碼',
      // 受限態不給返回：後端只放行本 RPC，退回去每個請求都是 AUTH-3004。
      // 用 `automaticallyImplyLeading: false` 的方式（leading 覆寫）保留其他頁的預設行為。
      leading: widget.mustChange
          ? const SizedBox.shrink()
          : null,
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 320),
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    widget.mustChange
                        ? '首次登入請先設定新密碼'
                        : '請輸入目前的密碼與新密碼',
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 20),
                  TextFormField(
                    controller: _oldController,
                    enabled: !_busy,
                    obscureText: true,
                    decoration: InputDecoration(
                      labelText: widget.mustChange ? '臨時密碼' : '目前密碼',
                    ),
                    textInputAction: TextInputAction.next,
                    validator: (v) =>
                        (v == null || v.isEmpty) ? '請輸入目前的密碼' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _newController,
                    enabled: !_busy,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: '新密碼',
                      helperText: '至少 8 個字元',
                    ),
                    textInputAction: TextInputAction.next,
                    validator: (v) {
                      if (v == null || v.isEmpty) return '請輸入新密碼';
                      // 後端 minNewPasswordLen = 8，前端鏡射同一門檻以免白送一次往返。
                      if (v.length < 8) return '新密碼至少 8 個字元';
                      if (v == _oldController.text) return '新密碼不可與目前密碼相同';
                      return null;
                    },
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _confirmController,
                    enabled: !_busy,
                    obscureText: true,
                    decoration: const InputDecoration(labelText: '確認新密碼'),
                    textInputAction: TextInputAction.done,
                    onFieldSubmitted: (_) => _submit(),
                    validator: (v) =>
                        (v != _newController.text) ? '兩次輸入的新密碼不一致' : null,
                  ),
                  const SizedBox(height: 24),
                  adaptiveFilledButton(
                    onPressed: _busy ? null : _submit,
                    child: _busy
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('確認修改'),
                  ),
                  const SizedBox(height: 8),
                  AdaptiveButton(
                    onPressed: _busy
                        ? null
                        : () async {
                            await widget.authRepository.logout();
                            if (!context.mounted) return;
                            context.router.navigatePath('/login');
                          },
                    child: const Text('登出'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
