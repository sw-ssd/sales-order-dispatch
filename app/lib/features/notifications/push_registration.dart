import 'dart:io';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/devices.pb.dart';

/// 推播註冊（07 計畫 Task 4.3.4／4.4.5 的 App 端）。
///
/// **可選能力，缺席時靜默跳過**——這是刻意的：
///   1. Firebase 需要原生設定檔（Android `google-services.json`、iOS
///      `GoogleService-Info.plist`）。這些檔綁定**特定 Firebase 專案**，不該進版控，
///      本 repo 也沒有（見計畫殘項）。沒有它們時 `Firebase.initializeApp()` 會擲例外。
///   2. 推播只是「有更好」的通知管道；站內通知中心已完整可用。若讓註冊失敗中斷登入，
///      等於為了附加功能擋掉主流程——那比沒有推播更糟。
///
/// 因此本函式**吞掉所有例外**並記一行 debug log；要讓它真的生效，只需補上兩個原生
/// 設定檔（程式碼無需再改）。
class PushRegistration {
  PushRegistration({required this.api, required this.token});

  final Api api;

  /// 目前登入者的 access token 供應者（由 Api 的 transport 共用同一來源）。
  final Future<String?> Function() token;

  bool _registered = false;

  /// 登入後呼叫：背景嘗試註冊裝置 token。
  ///
  /// 回傳是否註冊成功（供測試斷言；生產不依賴回傳值）。
  Future<bool> register() async {
    if (_registered) return true;
    if (kIsWeb) return false; // 1.0 的 App 推播不含 Web
    try {
      await Firebase.initializeApp();
      // iOS 需先取得通知權限才會核發 APNs token；Android 13+ 亦需 POST_NOTIFICATIONS。
      await FirebaseMessaging.instance.requestPermission();
      final fcmToken = await FirebaseMessaging.instance.getToken();
      if (fcmToken == null || fcmToken.isEmpty) return false;

      await api.devices.registerDevice(RegisterDeviceRequest(
        platform: Platform.isIOS ? 'ios' : 'android',
        fcmToken: fcmToken,
        deviceName: Platform.operatingSystem,
      ));
      _registered = true;
      // token 會輪替（重裝/還原/久未使用）→ 之後每次輪替都要回報，否則推播送到舊 token。
      FirebaseMessaging.instance.onTokenRefresh.listen((next) {
        api.devices
            .registerDevice(RegisterDeviceRequest(
          platform: Platform.isIOS ? 'ios' : 'android',
          fcmToken: next,
          deviceName: Platform.operatingSystem,
        ))
            .catchError((Object e) {
          debugPrint('push: token 輪替註冊失敗（不影響使用）: $e');
          return RegisterDeviceResponse();
        });
      });
      return true;
    } catch (e) {
      // 沒有 Firebase 原生設定檔、或平台不支援 → 不影響任何其他功能。
      debugPrint('push: 未啟用推播（缺 Firebase 原生設定檔？）: $e');
      return false;
    }
  }

  /// 登出時呼叫：註銷本機 token（盡力而為）。
  ///
  /// 不註銷的後果是**下一個人可能收到前一位使用者的推播**——同一台裝置換帳號登入時，
  /// 舊 token 仍綁在前一個 user_id 上（後端 RegisterDevice 是「同 token 轉移歸屬」，
  /// 只有新註冊才會改歸屬；沒登出的帳號期間推播會送到這台裝置）。
  Future<void> unregister() async {
    if (!_registered) return;
    try {
      final fcmToken = await FirebaseMessaging.instance.getToken();
      if (fcmToken == null || fcmToken.isEmpty) return;
      await api.devices.unregisterDevice(UnregisterDeviceRequest(fcmToken: fcmToken));
    } catch (e) {
      debugPrint('push: 登出時註銷裝置失敗（不影響登出）: $e');
    } finally {
      _registered = false;
    }
  }
}
