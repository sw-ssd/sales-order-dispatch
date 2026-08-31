import 'package:flutter/widgets.dart';

import 'app.dart';
import 'config.dart';

/// 共用啟動入口：main_dev.dart / main_prod.dart 注入各自 AppConfig。
void bootstrap(AppConfig config) {
  runApp(SalesOrderApp(config: config));
}
