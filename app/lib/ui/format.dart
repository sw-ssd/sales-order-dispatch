// 日期時間顯示：RFC3339 → `YYYY-MM-DD HH:mm`（時區原樣顯示，不引 intl）。

String formatDateTime(String rfc3339) {
  if (rfc3339.isEmpty) return '—';
  final t = rfc3339.replaceFirst('T', ' ');
  final cut = t.split('.').first.split('+').first;
  return cut.length >= 16 ? cut.substring(0, 16) : cut;
}

String formatDate(String rfc3339) =>
    rfc3339.isEmpty ? '—' : rfc3339.split('T').first;
