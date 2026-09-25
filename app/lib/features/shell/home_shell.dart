import 'package:auto_route/auto_route.dart';
import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:fquery/fquery.dart';

import '../../core/api.dart';
import '../../gen/salesorder/v1/notifications.pb.dart';
import '../../features/auth/auth_repository.dart';
import '../../features/announcements/announcements_page.dart';
import '../../features/notifications/notifications_page.dart';
import '../../features/orders/orders_page.dart';
import '../../features/returns/returns_page.dart';
import '../../ui/adaptive.dart';

/// 主殼：Android Material BottomNavigationBar、iOS CupertinoTabScaffold；
/// 通知未讀數 60s 輪詢作角標。
class HomeShell extends HookWidget {
  const HomeShell({super.key, required this.api, required this.auth});

  final Api api;
  final AuthRepository auth;

  @override
  Widget build(BuildContext context) {
    final tabIndex = useState(0);

    // 未登入（token 缺失）→ 導回登入頁。
    useEffect(() {
      auth.currentAccessToken().then((token) {
        if (token == null && context.mounted) {
          context.router.navigatePath('/login');
        }
      });
      return null;
    }, const []);

    final unread = useQuery<UnreadCountResponse, Exception>(
      ['unread'],
      () => api.notifications.unreadCount(UnreadCountRequest()),
      context: context,
      refetchInterval: const Duration(seconds: 60),
    );
    final badge = unread.data?.count ?? 0;

    final pages = <Widget>[
      AnnouncementsPage(api: api),
      OrdersPage(api: api),
      ReturnsPage(api: api),
      NotificationsPage(api: api),
      _ProfilePage(auth: auth),
    ];

    if (isCupertinoTarget()) {
      return CupertinoTabScaffold(
        tabBar: CupertinoTabBar(
          currentIndex: tabIndex.value,
          onTap: (index) => tabIndex.value = index,
          items: [
            const BottomNavigationBarItem(
              icon: Icon(CupertinoIcons.house),
              label: '首頁',
            ),
            const BottomNavigationBarItem(
              icon: Icon(CupertinoIcons.doc_text),
              label: '訂單',
            ),
            const BottomNavigationBarItem(
              icon: Icon(CupertinoIcons.arrow_uturn_left),
              label: '退貨',
            ),
            BottomNavigationBarItem(
              icon: Badge(
                isLabelVisible: badge > 0,
                label: Text('$badge'),
                child: const Icon(CupertinoIcons.bell),
              ),
              label: '通知',
            ),
            const BottomNavigationBarItem(
              icon: Icon(CupertinoIcons.person_circle),
              label: '我的',
            ),
          ],
        ),
        tabBuilder: (context, index) =>
            CupertinoTabView(builder: (_) => pages[index]),
      );
    }

    return Scaffold(
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: tabIndex.value,
        onTap: (index) => tabIndex.value = index,
        type: BottomNavigationBarType.fixed,
        items: [
          const BottomNavigationBarItem(
            icon: Icon(Icons.home_outlined),
            label: '首頁',
          ),
          const BottomNavigationBarItem(
            icon: Icon(Icons.receipt_long),
            label: '訂單',
          ),
          const BottomNavigationBarItem(
            icon: Icon(Icons.keyboard_return),
            label: '退貨',
          ),
          BottomNavigationBarItem(
            icon: Badge(
              isLabelVisible: badge > 0,
              label: Text('$badge'),
              child: const Icon(Icons.notifications),
            ),
            label: '通知',
          ),
          const BottomNavigationBarItem(
            icon: Icon(Icons.person),
            label: '我的',
          ),
        ],
      ),
      body: IndexedStack(index: tabIndex.value, children: pages),
    );
  }
}

/// 我的：身分出口（登出）。
class _ProfilePage extends StatelessWidget {
  const _ProfilePage({required this.auth});

  final AuthRepository auth;

  @override
  Widget build(BuildContext context) {
    return adaptivePage(
      context,
      title: '我的',
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('多公司訂出貨系統',
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 24),
            AdaptiveButton(
              onPressed: () =>
                  context.router.pushPath('/change-password?forced=false'),
              child: const Text('修改密碼'),
            ),
            const SizedBox(height: 12),
            AdaptiveButton(
              destructive: true,
              onPressed: () async {
                final ok = await confirmAdaptive(
                  context,
                  title: '登出',
                  message: '確定要登出？',
                  confirmLabel: '登出',
                  destructive: true,
                );
                if (!ok || !context.mounted) return;
                await auth.logout();
                if (context.mounted) {
                  context.router.navigatePath('/login');
                }
              },
              child: const Text('登出'),
            ),
          ],
        ),
      ),
    );
  }
}
