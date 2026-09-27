import 'package:flutter_test/flutter_test.dart';
import 'package:sales_order_app/ui/format.dart';

// 這一組測的核心是「**印出的值不得隨輸入的時區偏移改變**」。
// 後端有 5 個服務輸出 `...Z`、15 個輸出 `+08:00`（見 format.dart 的說明），
// 先前的 `split('+').first` 會把 UTC 的牆上時間當成台北時間——同一個瞬間差八小時。
void main() {
  group('formatDateTime', () {
    test('同一個瞬間不論後端用 Z 還是 +08:00 序列化，印出來完全相同', () {
      const instant = '2026-09-24T17:43:09Z';
      const sameInstant = '2026-09-25T01:43:09+08:00';
      expect(formatDateTime(instant), formatDateTime(sameInstant));
      // 營業時區 UTC+8：17:43Z 是隔天凌晨 01:43。
      expect(formatDateTime(instant), '2026-09-25 01:43');
    });

    test('帶負偏移也正確（後端若改以其他偏移序列化不會走鐘）', () {
      expect(formatDateTime('2026-09-24T12:43:09-05:00'), '2026-09-25 01:43');
    });

    test('跨日與跨年邊界以營業時區判定', () {
      // 16:00Z = 隔天 00:00(+08)。若誤用 UTC 會顯示成前一天。
      expect(formatDateTime('2026-12-31T16:00:00Z'), '2027-01-01 00:00');
      expect(formatDateTime('2026-12-31T15:59:00Z'), '2026-12-31 23:59');
    });

    test('空值與無法解析的值顯示佔位字元，不丟錯', () {
      expect(formatDateTime(''), '—');
      expect(formatDateTime('not-a-date'), '—');
    });

    test('純日期輸入原樣回傳（不是瞬間，加八小時會推成前一天 08:00）', () {
      expect(formatDateTime('2026-09-01'), '2026-09-01');
    });
  });

  group('formatDate', () {
    test('純日曆日欄位原樣取日期，不經時區換算（轉了會平移一天）', () {
      expect(formatDate('2026-09-25'), '2026-09-25');
      expect(formatDate(''), '—');
    });
  });
}
