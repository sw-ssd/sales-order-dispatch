import 'package:auto_route/auto_route.dart';
import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';

import '../../../core/error_info.dart';
import '../../../gen/salesorder/v1/auth.pb.dart' as pb;
import '../../../ui/adaptive.dart';
import '../auth_repository.dart';

/// QR 兌換頁（規格 §4.2 QR Code 登入；深層連結 `/customer_account_qrcode/:token`）。
///
/// 兩段式：先以 token 兌換出「這是哪家公司／哪個客戶」與**可選的子帳號清單**，
/// 使用者選一個子帳號後輸入該帳號的密碼完成登入。兌換本身不核發憑證 —— token 只是
/// 把店家帶到登入頁，登入仍須帳密。
///
/// 兌換用未認證 transport（`AuthRepository.exchangeQR`）：這是公開端點，且此時
/// 使用者本來就還沒登入。
///
/// 為何選單不讓使用者自己打帳號名稱：業務子帳號與主帳號由後端排除（規格明訂），
/// 讓使用者從清單挑才不會打進一個必然被拒的名字。
class QRLoginPage extends HookWidget {
  const QRLoginPage({
    super.key,
    required this.authRepository,
    required this.token,
    this.onLoggedIn,
  });

  final AuthRepository authRepository;
  final String token;

  /// 登入成功後的回呼（預設導向 `/home`；QR 以外的深層連結流程可覆寫）。
  final void Function(String accountName)? onLoggedIn;

  @override
  Widget build(BuildContext context) {
    final selected = useState<String>('');
    final password = useTextEditingController();
    final busy = useState(false);

    final exchange = useQuery<pb.QRLoginResponse, Exception>(
      ['qr_exchange', token],
      () => authRepository.exchangeQR(token),
      context: context,
    );

    Future<void> submit(pb.QRLoginResponse data) async {
      if (selected.value.isEmpty || password.text.isEmpty) return;
      busy.value = true;
      try {
        final result = await authRepository.loginShop(
          customerCode: selected.value,
          password: password.text,
        );
        if (!context.mounted) return;
        // 首登/臨時密碼受限態（A3）：後端只放行 ChangePassword，直接進主殼只會失敗。
        if (result.mustChangePassword) {
          context.router.navigatePath('/change-password');
          return;
        }
        if (onLoggedIn != null) {
          onLoggedIn!(selected.value);
        } else {
          context.router.navigatePath('/home');
        }
      } catch (e) {
        if (!context.mounted) return;
        showFeedback(context, localizedErrorMessage(e));
      } finally {
        busy.value = false;
      }
    }

    return adaptivePage(
      context,
      title: '店家登入',
      body: exchange.isLoading
          ? const Center(child: CircularProgressIndicator())
          : exchange.isError
              ? _QRError(
                  error: exchange.error!,
                  onRetry: () => exchange.refetch(),
                )
              : _QRForm(
                  data: exchange.data!,
                  selected: selected,
                  password: password,
                  busy: busy.value,
                  onSubmit: () => submit(exchange.data!),
                ),
    );
  }
}

/// 兌換成功後的表單：顯示公司／客戶，並讓使用者選一個子帳號。
class _QRForm extends StatelessWidget {
  const _QRForm({
    required this.data,
    required this.selected,
    required this.password,
    required this.busy,
    required this.onSubmit,
  });

  final pb.QRLoginResponse data;
  final ValueNotifier<String> selected;
  final TextEditingController password;
  final bool busy;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    final accounts = data.accounts;
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 360),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                data.companyName,
                style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 4),
              Text(
                data.customerName,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 14, color: Color(0xFF757575)),
              ),
              const SizedBox(height: 24),
              const Text('選擇登入帳號'),
              const SizedBox(height: 8),
              // 用可點選清單列而非 RadioListTile：Flutter 3.35 已棄用其
              // groupValue/onChanged（改走 RadioGroup 祖先），而這裡只要「單選清單」，
              // 勾選符號自己畫反而少一層依賴、也不必追 API 遷移。
              for (final account in accounts)
                adaptiveListTile(
                  title: Text(account.accountName),
                  trailing: selected.value == account.accountName
                      ? const Icon(CupertinoIcons.checkmark_alt, size: 20)
                      : null,
                  isSelected: selected.value == account.accountName,
                  onTap: busy ? null : () => selected.value = account.accountName,
                ),
              const SizedBox(height: 16),
              TextField(
                controller: password,
                enabled: !busy,
                obscureText: true,
                onSubmitted: (_) => onSubmit(),
                decoration: const InputDecoration(labelText: '密碼'),
              ),
              const SizedBox(height: 24),
              adaptiveFilledButton(
                onPressed: busy ? null : onSubmit,
                child: busy
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
    );
  }
}

/// 兌換失敗：token 失效／已使用／客戶不存在等，一律導回一般登入。
class _QRError extends StatelessWidget {
  const _QRError({required this.error, required this.onRetry});

  final Exception error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(CupertinoIcons.exclamationmark_triangle, size: 40),
            const SizedBox(height: 12),
            Text(
              qrErrorMessage(error),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            AdaptiveButton(onPressed: onRetry, child: const Text('重試')),
            AdaptiveButton(
              onPressed: () => context.router.navigatePath('/login/shop'),
              child: const Text('改用帳密登入'),
            ),
          ],
        ),
      ),
    );
  }
}

/// QR 兌換錯誤訊息（繁體中文）：失效／已用一律請使用者重新取得 QR。
String qrErrorMessage(Object error) {
  if (error is StateError) return error.message;
  return '此 QR Code 已失效或已使用過，請向業務重新取得。';
}
