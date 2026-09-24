//
//  Generated code. Do not modify.
//  source: salesorder/v1/announcement.proto
//

import "package:connectrpc/connect.dart" as connect;
import "announcement.pb.dart" as salesorderv1announcement;

/// AnnouncementService:公告 CMS(spec announcements)。
/// 三型別(banner/news/article)、三層發佈範圍(company/department 皆空 = 全系統)、
/// 上下架時間窗、平台投放(Web/App)。CRUD 範圍守衛:super 全範圍、
/// company_admin 所屬公司(公司層+該公司部門層)、dept_admin 本部門層。
abstract final class AnnouncementService {
  /// Fully-qualified name of the AnnouncementService service.
  static const name = 'salesorder.v1.AnnouncementService';

  /// ListAnnouncements:管理列表(僅列可管理範圍;含未上架/停用;預設排除已刪除)。
  static const listAnnouncements = connect.Spec(
    '/$name/ListAnnouncements',
    connect.StreamType.unary,
    salesorderv1announcement.ListAnnouncementsRequest.new,
    salesorderv1announcement.ListAnnouncementsResponse.new,
  );

  /// CreateAnnouncement:建立公告(範圍依身分收斂;type 非法 → invalid_argument)。
  static const createAnnouncement = connect.Spec(
    '/$name/CreateAnnouncement',
    connect.StreamType.unary,
    salesorderv1announcement.CreateAnnouncementRequest.new,
    salesorderv1announcement.CreateAnnouncementResponse.new,
  );

  /// UpdateAnnouncement:全量替換(id 除外表所有欄位;布林無 present 語意,不採欄位式)。
  /// 空字串的可選欄位(image_url/link_url/unpublish_at)即清空;publish_at 空 → invalid_argument。
  static const updateAnnouncement = connect.Spec(
    '/$name/UpdateAnnouncement',
    connect.StreamType.unary,
    salesorderv1announcement.UpdateAnnouncementRequest.new,
    salesorderv1announcement.UpdateAnnouncementResponse.new,
  );

  /// DeleteAnnouncement:軟刪除。
  static const deleteAnnouncement = connect.Spec(
    '/$name/DeleteAnnouncement',
    connect.StreamType.unary,
    salesorderv1announcement.DeleteAnnouncementRequest.new,
    salesorderv1announcement.DeleteAnnouncementResponse.new,
  );
}
