import 'package:flutter/material.dart';

import 'config.dart';

/// 根 Widget。D29：後續在此掛 CacheProvider（fquery）+
/// 根 ProviderScope（disco）+ MaterialApp.router（auto_route）。
class SalesOrderApp extends StatelessWidget {
  const SalesOrderApp({super.key, required this.config});

  final AppConfig config;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '多公司訂出貨系統',
      home: Scaffold(
        body: Center(
          child: Text('Wave 1 骨架(env: ${config.env.name})'),
        ),
      ),
    );
  }
}
