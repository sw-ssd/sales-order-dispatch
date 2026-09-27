import 'package:connectrpc/connect.dart' as connect;
import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/core/error_info.dart';
import 'package:sales_order_app/gen/salesorder/v1/common.pb.dart';

/// 造一筆後端形狀的錯誤：`ErrorInfo` 掛在 details 裡（與 backend `errcode.attachInfo` 同形）。
connect.ConnectException _coded(
  String code,
  connect.Code connectCode, {
  String message = '',
  Map<String, String> details = const {},
}) {
  final info = ErrorInfo(code: code, message: message)
    ..details.addAll(details);
  return connect.ConnectException(
    connectCode,
    message,
    details: [
      connect.ErrorDetail('salesorder.v1.ErrorInfo', info.writeToBuffer()),
    ],
  );
}

void main() {
  group('localizedErrorMessage：details.reason 不得被樣板吞掉', () {
    // 後端 `errcode.Code.Render` 只替換**樣板裡**的 {佔位符}；`SYS-1001` 的樣板是
    // 「參數驗證失敗」、沒有任何佔位符，所以帶 `reason` 的呼叫（print_service 的
    // 「無可列印資料」、dispatch_service 的「僅待派訂單可指派」、customer_account_service
    // 的 `not_manageable`…）渲染出來的 `message` 就是那句無用的樣板。
    // 可行動的原因只存在 `details['reason']`，前端不讀就等於沒說。
    test('SYS-1001 帶中文 reason → 顯示那句 reason，不是「參數驗證失敗」', () {
      final err = _coded(
        'SYS-1001',
        connect.Code.invalidArgument,
        message: '參數驗證失敗',
        details: {'reason': '無可列印資料'},
      );
      expect(localizedErrorMessage(err), '無可列印資料');
    });

    test('SYS-1001 帶符號 reason（後端內部代號）→ 不得原樣顯示給使用者', () {
      // customer_account_service 用 `primary_account`／`system_generated`／`not_manageable`
      // 這類英文代號當 reason。它們是給程式判斷的，不是文案。
      final err = _coded(
        'SYS-1001',
        connect.Code.invalidArgument,
        message: '參數驗證失敗',
        details: {'reason': 'not_manageable'},
      );
      expect(localizedErrorMessage(err), isNot(contains('not_manageable')));
    });

    test('SYS-2001 帶 reason 時顯示碼表樣板（reason 為符號時不得當文案）', () {
      final err = _coded(
        'SYS-2001',
        connect.Code.alreadyExists,
        message: '資料衝突，請確認識別碼是否已被使用',
        details: {'reason': 'already_inactive'},
      );
      expect(localizedErrorMessage(err), isNot(contains('already_inactive')));
    });
  });
}
