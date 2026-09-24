import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

import '../../../ui/adaptive.dart';
import '../auth_repository.dart';

/// 店家登入頁(/login/shop):customer_code + password → AuthService.Login。
class LoginPage extends StatefulWidget {
  const LoginPage({super.key, required this.authRepository});

  final AuthRepository authRepository;

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
      // 登入成功 → 進入主殼（navigate 清掉登入棧，返回鍵不會退回登入頁）。
      context.router.navigatePath('/home');
    } catch (e) {
      if (!mounted) return;
      showFeedback(context, authErrorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return adaptivePage(
      context,
      title: '店家登入',
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
