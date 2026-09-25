import 'package:connectrpc/test.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fquery/fquery.dart';
import 'package:fquery_core/fquery_core.dart';
import 'package:sales_order_app/core/api.dart';
import 'package:sales_order_app/features/announcements/announcements_page.dart';
import 'package:sales_order_app/features/auth/auth_repository.dart';
import 'package:sales_order_app/features/auth/token_storage.dart';
import 'package:sales_order_app/gen/salesorder/v1/announcement.connect.spec.dart'
    as specs;
import 'package:sales_order_app/gen/salesorder/v1/announcement.pb.dart';
import 'package:sales_order_app/gen/salesorder/v1/auth.connect.client.dart';

class _InMemoryTokenStorage implements TokenStorage {
  AuthTokenPair? stored;
  @override
  Future<void> save(AuthTokenPair tokens) async => stored = tokens;
  @override
  Future<AuthTokenPair?> read() async => stored;
  @override
  Future<void> clear() async => stored = null;
}

Announcement _ann({
  required String id,
  required String type,
  required String title,
  String content = '',
  String linkUrl = '',
}) =>
    Announcement(
      id: id,
      type: type,
      title: title,
      content: content,
      linkUrl: linkUrl,
      sortOrder: 1,
      isActive: true,
      deployApp: true,
    );

/// 以 fake transport 掛首頁公告；回傳 captured 供斷言請求參數。
Future<List<ListActiveAnnouncementsRequest>> _pump(
  WidgetTester tester, {
  List<Announcement> banners = const [],
  List<Announcement> news = const [],
  List<Announcement> articles = const [],
  bool fail = false,
}) async {
  final captured = <ListActiveAnnouncementsRequest>[];
  final transport = FakeTransportBuilder()
      .unary<ListActiveAnnouncementsRequest, ListActiveAnnouncementsResponse>(
          specs.AnnouncementService.listActiveAnnouncements, (req, _) {
    captured.add(req);
    if (fail) {
      throw Exception('boom');
    }
    return ListActiveAnnouncementsResponse(
      banners: banners,
      news: news,
      articles: articles,
    );
  }).build();
  final auth = AuthRepository(
    client: AuthServiceClient(transport),
    authed: AuthServiceClient(transport),
    tokenStorage: _InMemoryTokenStorage(),
  );

  await tester.pumpWidget(CacheProvider(
    cache: QueryCache(),
    child: MaterialApp(
      home: AnnouncementsPage(
        api: Api(baseUrl: 'http://test', auth: auth, transport: transport),
      ),
    ),
  ));
  await tester.pumpAndSettle();
  return captured;
}

/// 收掉 fquery 的 GC 計時器（見 accounts_page_test.dart 同註）。
Future<void> _drain(WidgetTester tester) async {
  await tester.pumpWidget(const SizedBox());
  await tester.pump(const Duration(minutes: 6));
}

void main() {
  testWidgets('以 platform=app 查公告 —— 平台不是後端預設，必須明確指定', (tester) async {
    final captured = await _pump(tester, news: [
      _ann(id: 'n1', type: 'news', title: '本週菜價調整'),
    ]);

    // 不傳 platform 後端會 invalid_argument；傳成 web 則會顯示只投 Web 的公告。
    expect(captured.single.platform, 'app');

    await _drain(tester);
  });

  testWidgets('news 與 article 都進列表；兩者皆可展開看內文', (tester) async {
    await _pump(
      tester,
      news: [_ann(id: 'n1', type: 'news', title: '本週菜價調整', content: '葉菜類調漲')],
      articles: [
        _ann(id: 'a1', type: 'article', title: '食材保存指南', content: '冷藏 0-7 度'),
      ],
    );

    expect(find.text('本週菜價調整'), findsOneWidget);
    expect(find.text('食材保存指南'), findsOneWidget);
    // 標籤區分型別。
    expect(find.text('消息'), findsOneWidget);
    expect(find.text('圖文'), findsOneWidget);

    // 展開 article 才看到內文（規格「article 提供圖文內容檢視」）。
    expect(find.text('冷藏 0-7 度'), findsNothing);
    await tester.tap(find.text('食材保存指南'));
    await tester.pumpAndSettle();
    expect(find.text('冷藏 0-7 度'), findsOneWidget);

    await _drain(tester);
  });

  testWidgets('banner 有多張時可滑動，且依後端回傳次序（前端不重排）', (tester) async {
    await _pump(tester, banners: [
      _ann(id: 'b1', type: 'banner', title: '第一則'),
      _ann(id: 'b2', type: 'banner', title: '第二則'),
    ]);

    expect(find.text('第一則'), findsOneWidget);
    // 滑到下一張。用 `fling`（帶速度）而非 `drag`：`drag` 的位移在同一幀內完成、
    // PageView 的捲動物理判定為未過半頁而不切頁（實測）。
    await tester.fling(find.byType(PageView), const Offset(-400, 0), 1000);
    await tester.pumpAndSettle();
    expect(find.text('第二則'), findsOneWidget);
    expect(find.text('第一則'), findsNothing);

    await _drain(tester);
  });

  testWidgets('banner 有 link_url 時點擊顯示連結（App 內不直開瀏覽器）', (tester) async {
    await _pump(tester, banners: [
      _ann(id: 'b1', type: 'banner', title: '優惠', linkUrl: 'https://example.com/p'),
    ]);

    await tester.tap(find.text('優惠'));
    await tester.pumpAndSettle();

    // 連結以對話框呈現（url_launcher 是本階段未引入的平台相依）。
    expect(find.textContaining('https://example.com/p'), findsOneWidget);

    await _drain(tester);
  });

  testWidgets('沒有公告時顯示空態，不顯示任何卡片', (tester) async {
    await _pump(tester);

    expect(find.text('目前沒有公告'), findsOneWidget);
    expect(find.byType(ExpansionTile), findsNothing);

    await _drain(tester);
  });

  testWidgets('查詢失敗顯示可讀訊息（不靜默留白）', (tester) async {
    await _pump(tester, fail: true);

    // 首頁是主畫面，失敗必須看得出來 —— 與 Web 的 Dashboard 附掛不同
    // （那裡公告失敗不該吐錯誤區塊，因為訂單概況才是主體）。
    expect(find.textContaining('載入失敗'), findsOneWidget);

    await _drain(tester);
  });
}
