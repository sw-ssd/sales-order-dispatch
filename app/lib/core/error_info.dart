import 'dart:io';

import 'package:connectrpc/connect.dart' as connect;

import '../gen/errcode.dart';
import '../gen/salesorder/v1/common.pb.dart';

/// 由 Connect 錯誤取出後端 [ErrorInfo]（碼／訊息／details）。
///
/// 後端每個錯誤都掛一筆 `salesorder.v1.ErrorInfo` detail（見 `errcode.Code.Error`），
/// 但 Dart 端的 `ConnectException.details` 只給 `(type, value)` 原始位元組 —— 要自己
/// `ErrorInfo.fromBuffer`。少了這一步，App 只能看 connect 碼（`failed_precondition`
/// 底下 AUTH-3002/3003/3004 三碼無法區分），於是「臨時密碼過期」與「首登須改密碼」
/// 都會顯示成同一句泛用錯誤。
ErrorInfo? errorInfoOf(Object error) {
  if (error is! connect.ConnectException) return null;
  for (final d in error.details) {
    // type 是 any 的 typeUrl 尾段，可能是 `salesorder.v1.ErrorInfo` 或簡名。
    if (!d.type.endsWith('ErrorInfo')) continue;
    try {
      return ErrorInfo.fromBuffer(d.value);
    } catch (_) {
      return null; // 別的服務的同名 detail 或格式不符 → 當作沒有
    }
  }
  return null;
}

/// 錯誤碼（`ErrorInfo.code`，如 'AUTH-3004'）；無則 null。
String? errorCodeOf(Object error) => errorInfoOf(error)?.code;

/// 取錯誤訊息：優先用後端碼表（繁中、與 Web 一致），退回 detail 自帶訊息，
/// 再退回 [authErrorMessage] 的連線層對照。
///
/// `details` 的佔位符（如 AUTH-3003 的 `{until}`）以 detail 帶入的參數渲染。
String localizedErrorMessage(Object error) {
  final info = errorInfoOf(error);
  if (info == null) return authErrorMessage(error);
  final template = errCodeMessages[info.code];
  if (template == null) {
    return info.message.isNotEmpty ? info.message : authErrorMessage(error);
  }
  var out = template;
  for (final e in info.details.entries) {
    out = out.replaceAll('{${e.key}}', e.value);
  }
  return out;
}

/// 連線層錯誤訊息（無 `ErrorInfo` 時的最後一道）。
///
/// 全 App **唯一一份**：改密碼、登入、身分選擇、QR 兌換都走這裡，第二份複本會讓
/// 同一個 connect 碼在不同頁面出現不同說法。
String authErrorMessage(Object error) {
  if (error is connect.ConnectException) {
    return switch (error.code) {
      connect.Code.unimplemented => '伺服器尚未支援登入功能,請稍後再試',
      connect.Code.unavailable => '無法連線至伺服器,請檢查網路後再試',
      connect.Code.unauthenticated => '客戶編號或密碼錯誤,請重新輸入',
      connect.Code.permissionDenied => '沒有權限執行此操作',
      connect.Code.deadlineExceeded => '連線逾時,請稍後再試',
      _ => '操作失敗(${error.code.name}),請稍後再試',
    };
  }
  if (error is SocketException) return '無法連線至伺服器,請檢查網路後再試';
  if (error is StateError) return error.message;
  return '操作失敗,請稍後再試';
}
