//
//  Generated code. Do not modify.
//  source: salesorder/v1/returns.proto
//

import "package:connectrpc/connect.dart" as connect;
import "returns.pb.dart" as salesorderv1returns;

/// ReturnService:退貨申請與審核(06 計畫 Task 4.7.2–4.7.4, D25)。
/// 發起僅客戶子帳號(主帳號一律拒絕);審核僅主責業務/dept_admin 以上;
/// 全程不修改原訂單(僅參照)。
abstract final class ReturnService {
  /// Fully-qualified name of the ReturnService service.
  static const name = 'salesorder.v1.ReturnService';

  /// CreateReturnRequest:發起退貨(子帳號;雙來源並存;同交易寫申請+明細+稽核)。
  static const createReturnRequest = connect.Spec(
    '/$name/CreateReturnRequest',
    connect.StreamType.unary,
    salesorderv1returns.CreateReturnRequestRequest.new,
    salesorderv1returns.CreateReturnRequestResponse.new,
  );

  /// ListReturnRequests:申請列表(客戶 self 限自己客戶;staff 依範圍;分頁 per_page ≤ 100)。
  static const listReturnRequests = connect.Spec(
    '/$name/ListReturnRequests',
    connect.StreamType.unary,
    salesorderv1returns.ListReturnRequestsRequest.new,
    salesorderv1returns.ListReturnRequestsResponse.new,
  );

  /// GetReturnRequest:單筆含品項明細(照片轉下載 URL)。
  static const getReturnRequest = connect.Spec(
    '/$name/GetReturnRequest',
    connect.StreamType.unary,
    salesorderv1returns.GetReturnRequestRequest.new,
    salesorderv1returns.GetReturnRequestResponse.new,
  );

  /// ReviewReturnRequest:審核(approved/rejected + 樂觀鎖;同交易寫稽核;不碰原訂單)。
  static const reviewReturnRequest = connect.Spec(
    '/$name/ReviewReturnRequest',
    connect.StreamType.unary,
    salesorderv1returns.ReviewReturnRequestRequest.new,
    salesorderv1returns.ReviewReturnRequestResponse.new,
  );

  /// GetReturnCertificate:退貨證明(僅 approved;快照內容;唯讀不寫稽核)。
  static const getReturnCertificate = connect.Spec(
    '/$name/GetReturnCertificate',
    connect.StreamType.unary,
    salesorderv1returns.GetReturnCertificateRequest.new,
    salesorderv1returns.GetReturnCertificateResponse.new,
  );
}
