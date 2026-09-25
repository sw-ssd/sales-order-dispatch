import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';

import '../../../core/error_info.dart';
import '../../../ui/adaptive.dart';
import '../auth_repository.dart';

/// 身分選擇頁(/login):我是店家 / 我是業務(D5 雙身分入口)。
class IdentitySelectPage extends StatefulWidget {
  const IdentitySelectPage({super.key, required this.authRepository});

  final AuthRepository authRepository;

  @override
  State<IdentitySelectPage> createState() => _IdentitySelectPageState();
}

class _IdentitySelectPageState extends State<IdentitySelectPage> {
  bool _busy = false;

  Future<void> _loginSalesWithGoogle() async {
    setState(() => _busy = true);
    try {
      final grant = await widget.authRepository.loginSalesWithGoogle();
      if (!mounted) return;
      if (grant == null) return; // 使用者取消,不提示
      showFeedback(context, '已取得 Google 授權,待後端端點上線後完成登入');
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
      title: '登入',
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 320),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                '請選擇您的身分',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 24),
              adaptiveFilledButton(
                onPressed:
                    _busy ? null : () => context.router.pushPath('/login/shop'),
                child: const Text('我是店家'),
              ),
              const SizedBox(height: 12),
              AdaptiveButton(
                onPressed: _busy ? null : _loginSalesWithGoogle,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    if (_busy)
                      const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    else
                      const Icon(Icons.login, size: 18),
                    const SizedBox(width: 8),
                    const Text('我是業務(Google 登入)'),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
