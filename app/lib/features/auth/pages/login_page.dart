import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

import '../../../core/api.dart';
import '../../../gen/customers/v1/customer.pb.dart';
import '../../../ui/adaptive.dart';
import '../auth_repository.dart';

/// 店家登入頁(/login/shop):帳號名稱 + 密碼 → AuthService.Login。
///
/// 登入後**依身分分流**(規格 §4.2「主帳號登入僅顯示帳號管理畫面」):
/// 主帳號落在帳號管理頁 —— 它的業務能力已被後端 OpenFGA 全數排除(`primary_account`),
/// 進主殼只會看到一連串 403。子帳號落在主殼。
///
/// 分流靠「試呼 `ListCustomerAccounts`」判斷:這是唯一能區分主/子的既有 API,而且
/// **不新增 proto 欄位**(不為了前端分流而動對外契約)。非主帳號會被後端回
/// permission_denied —— 那正是判別訊號,不是錯誤。
class LoginPage extends StatefulWidget {
  const LoginPage({
    super.key,
    required this.authRepository,
    required this.api,
    this.title = '店家登入',
  });

  final AuthRepository authRepository;
  final Api api;

  /// 導覽列標題；帳號管理深層連結進來的登入頁用不同字樣。
  final String title;

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _customerCodeController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _busy = false;

  @override
  void dispose() {
    _customerCodeController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _busy = true);
    try {
      await widget.authRepository.loginShop(
        customerCode: _customerCodeController.text.trim(),
        password: _passwordController.text,
      );
      if (!mounted) return;
      // 登入成功 → 依身分流（navigate 清掉登入棧，返回鍵不會退回登入頁）。
      final isPrimary = await _isPrimaryAccount();
      if (!mounted) return;
      context.router.navigatePath(isPrimary ? '/account' : '/home');
    } catch (e) {
      if (!mounted) return;
      showFeedback(context, authErrorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  /// 以「試呼帳號管理 API」判斷目前登入者是否為客戶主帳號。
  ///
  /// 主帳號是唯一有權呼叫 `ListCustomerAccounts` 的身分（後端 `is_primary` 檢查），
  /// 故 `permission_denied` 代表「不是主帳號」。**其他錯誤一律當作不是主帳號** ——
  /// 分流判斷失敗時走主殼（通用路徑），不讓使用者卡在登入後的白畫面。
  Future<bool> _isPrimaryAccount() async {
    try {
      await widget.api.accounts
          .listCustomerAccounts(ListCustomerAccountsRequest());
      return true;
    } catch (_) {
      return false;
    }
  }

  @override
  Widget build(BuildContext context) {
    return adaptivePage(
      context,
      title: widget.title,
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 320),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextFormField(
                  controller: _customerCodeController,
                  enabled: !_busy,
                  decoration: const InputDecoration(
                    labelText: '客戶編號',
                    hintText: 'customer_code',
                  ),
                  textInputAction: TextInputAction.next,
                  validator: (value) =>
                      (value == null || value.trim().isEmpty) ? '請輸入客戶編號' : null,
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _passwordController,
                  enabled: !_busy,
                  obscureText: true,
                  decoration: const InputDecoration(labelText: '密碼'),
                  textInputAction: TextInputAction.done,
                  onFieldSubmitted: (_) => _submit(),
                  validator: (value) =>
                      (value == null || value.isEmpty) ? '請輸入密碼' : null,
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
                      : const Text('登入'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
