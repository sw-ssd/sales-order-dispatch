// 日期時間顯示：RFC3339 → `YYYY-MM-DD HH:mm`（**營業時區 UTC+8**）。
//
// 為什麼不是「時區原樣顯示」：後端各服務對同一個 `timestamptz` 的序列化偏移並不統一
// （部分 `.UTC()` 輸出 `...Z`、其餘輸出伺服器本地 `+08:00`），所以直接切字串會讓同一個
// 瞬間在不同端顯示不同時間。先前這裡 `split('+').first` 把偏移整個丟掉，等於**把 UTC
// 的牆上時間當成台北時間**——Web 端顯示 01:43 時 App 顯示 17:43，相差八小時。
//
// 因此：一律解析成瞬間，再換算到營業時區輸出。後端算營業日曆用的是
// `time.FixedZone("UTC+8", 8*3600)`（見 backend sales_order_assembly.go），前端同一組規則。
// 台北自 1979 年起無日光節約，固定偏移與具名時區恆等。
//
// `formatDate` 只給**純日曆日**欄位用（`expectedDeliveryDate` 是 `YYYY-MM-DD`，不是瞬間），
// 因此不經時區換算——轉了會平移一天。

/// 營業時區偏移：UTC+8。
const Duration _businessOffset = Duration(hours: 8);

/// 空值與無法解析時顯示的佔位字元。
const String _empty = '—';

/// RFC3339 瞬間 → `YYYY-MM-DD HH:mm`（營業時區 UTC+8）。
///
/// 形狀是純日期（`YYYY-MM-DD`，不含 `T`）時原樣回傳：那不是瞬間，加八小時只會把它推成
/// 前一天 08:00。舊版就是這個行為，保留它才不會讓既有呼叫端安靜地走鐘。
/// 其他無法解析的輸入一律給佔位字元（不把垃圾字串當日期印出去）。
String formatDateTime(String rfc3339) {
  if (rfc3339.isEmpty) return _empty;
  final t = DateTime.tryParse(rfc3339);
  if (t == null) return _empty;
  // 純日期：`DateTime.tryParse` 會補成當天 00:00（本地時區）。這種輸入沒有瞬間語意，
  // 不入時區換算，直接回原字串的日期部分。
  if (!rfc3339.contains('T') && !rfc3339.contains(' ')) {
    return rfc3339.length >= 10 ? rfc3339.substring(0, 10) : rfc3339;
  }
  final b = t.toUtc().add(_businessOffset);
  final mm = b.month.toString().padLeft(2, '0');
  final dd = b.day.toString().padLeft(2, '0');
  final hh = b.hour.toString().padLeft(2, '0');
  final mi = b.minute.toString().padLeft(2, '0');
  return '${b.year}-$mm-$dd $hh:$mi';
}

/// 只取日期。**僅供 `YYYY-MM-DD` 形狀的日曆日欄位**（`expectedDeliveryDate`）；
/// 瞬間欄位請用 [formatDateTime]，否則 UTC 的日期會在營業時區的隔天出錯。
String formatDate(String dateOnly) =>
    dateOnly.isEmpty ? _empty : dateOnly.split('T').first;
