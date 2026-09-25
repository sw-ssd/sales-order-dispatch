//
//  Generated code. Do not modify.
//  source: salesorder/v1/announcement.proto
//

import "package:connectrpc/connect.dart" as connect;
import "announcement.pb.dart" as salesorderv1announcement;
import "announcement.connect.spec.dart" as specs;

/// AnnouncementService:公告 CMS(spec announcements)。
/// 三型別(banner/news/article)、三層發佈範圍(company/department 皆空 = 全系統)、
/// 上下架時間窗、平台投放(Web/App)。CRUD 範圍守衛:super 全範圍、
/// company_admin 所屬公司(公司層+該公司部門層)、dept_admin 本部門層。
extension type AnnouncementServiceClient (connect.Transport _transport) {
  /// ListAnnouncements:管理列表(僅列可管理範圍;含未上架/停用;預設排除已刪除)。
  Future<salesorderv1announcement.ListAnnouncementsResponse> listAnnouncements(
    salesorderv1announcement.ListAnnouncementsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AnnouncementService.listAnnouncements,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListActiveAnnouncements:前台列表(規格「上下架時間窗與啟用狀態」+「平台篩選投放」)
  /// —— 只回**當下可見**的公告(is_active=true、publish_at<=now、(unpublish_at 空或 >now)、
  /// deploy_web|deploy_app 依 platform 過濾),依 type 分組供輪播(banner)與列表(news/article)。
  /// 可見範圍仍由 RLS 兜底(全系統 + 自己公司 + 自己部門);不帶 page(前台一次全取)。
  Future<salesorderv1announcement.ListActiveAnnouncementsResponse> listActiveAnnouncements(
    salesorderv1announcement.ListActiveAnnouncementsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AnnouncementService.listActiveAnnouncements,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateAnnouncement:建立公告(範圍依身分收斂;type 非法 → invalid_argument)。
  Future<salesorderv1announcement.CreateAnnouncementResponse> createAnnouncement(
    salesorderv1announcement.CreateAnnouncementRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AnnouncementService.createAnnouncement,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateAnnouncement:全量替換(id 除外表所有欄位;布林無 present 語意,不採欄位式)。
  /// 空字串的可選欄位(image_url/link_url/unpublish_at)即清空;publish_at 空 → invalid_argument。
  Future<salesorderv1announcement.UpdateAnnouncementResponse> updateAnnouncement(
    salesorderv1announcement.UpdateAnnouncementRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AnnouncementService.updateAnnouncement,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteAnnouncement:軟刪除。
  Future<salesorderv1announcement.DeleteAnnouncementResponse> deleteAnnouncement(
    salesorderv1announcement.DeleteAnnouncementRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AnnouncementService.deleteAnnouncement,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
