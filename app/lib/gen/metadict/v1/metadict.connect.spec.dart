//
//  Generated code. Do not modify.
//  source: metadict/v1/metadict.proto
//

import "package:connectrpc/connect.dart" as connect;
import "metadict.pb.dart" as metadictv1metadict;

/// MetadictService:字典管理。
abstract final class MetadictService {
  /// Fully-qualified name of the MetadictService service.
  static const name = 'metadict.v1.MetadictService';

  /// ListMetadicts:分頁查詢(系統預設 + 當前部門擴充合併),可依 type 篩選。
  static const listMetadicts = connect.Spec(
    '/$name/ListMetadicts',
    connect.StreamType.unary,
    metadictv1metadict.ListMetadictsRequest.new,
    metadictv1metadict.ListMetadictsResponse.new,
  );

  /// GetMetadict:取得單一字典。
  static const getMetadict = connect.Spec(
    '/$name/GetMetadict',
    connect.StreamType.unary,
    metadictv1metadict.GetMetadictRequest.new,
    metadictv1metadict.GetMetadictResponse.new,
  );

  /// CreateMetadict:建立字典(super 建系統級;dept_admin/staff 自動帶當前部門,不接受請求帶 department_id)。
  static const createMetadict = connect.Spec(
    '/$name/CreateMetadict',
    connect.StreamType.unary,
    metadictv1metadict.CreateMetadictRequest.new,
    metadictv1metadict.CreateMetadictResponse.new,
  );

  /// UpdateMetadict:更新 display_name / sort_order / is_active(不含 type / code)。
  static const updateMetadict = connect.Spec(
    '/$name/UpdateMetadict',
    connect.StreamType.unary,
    metadictv1metadict.UpdateMetadictRequest.new,
    metadictv1metadict.UpdateMetadictResponse.new,
  );

  /// DeleteMetadict:軟刪除(order_source 不可刪)。
  static const deleteMetadict = connect.Spec(
    '/$name/DeleteMetadict',
    connect.StreamType.unary,
    metadictv1metadict.DeleteMetadictRequest.new,
    metadictv1metadict.DeleteMetadictResponse.new,
  );

  /// ListOptions:表單下拉選項(僅可選用啟用值;客戶端身分僅回系統預設)。
  static const listOptions = connect.Spec(
    '/$name/ListOptions',
    connect.StreamType.unary,
    metadictv1metadict.ListOptionsRequest.new,
    metadictv1metadict.ListOptionsResponse.new,
  );
}
