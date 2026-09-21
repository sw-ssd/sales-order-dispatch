package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PrintLog 正式列印記錄(09 計畫 Task 5.5.1):每次列印/重印一筆,PDF 全留存。
// 記錄類資料不軟刪除(同 audit_logs/sales_order_events,D10),由保留排程管理(2 年,§14.1)。
// is_reprint 由比對鍵推導(部門/document_type/route/target_date/選用 customer/warehouse),
// 重印必填 reprint_reason;file_asset_id 參照落檔 PDF。
type PrintLog struct {
	ent.Schema
}

// Fields of the PrintLog.
func (PrintLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id"),
		field.String("document_type").
			NotEmpty(),
		field.Int("route_id"),
		field.Int("customer_id").
			Optional().
			Nillable(),
		field.Int("warehouse_id").
			Optional().
			Nillable(),
		field.Time("target_date"),
		field.Bool("is_reprint").
			Default(false),
		field.String("reprint_reason").
			Optional(),
		field.Int("printed_by"),
		field.Time("printed_at").
			Default(mutableNow),
		field.Int("file_asset_id"),
	}
}

// Indexes of the PrintLog.
func (PrintLog) Indexes() []ent.Index {
	return []ent.Index{
		// 重印判斷查詢(5.5.1 步驟 4):部門/document/車次/日期/選用選擇器比對。
		index.Fields("department_id", "target_date", "document_type"),
		index.Fields("department_id", "document_type", "route_id", "target_date"),
	}
}
