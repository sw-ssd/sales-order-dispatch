package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PrintPreview 預覽記錄(09 計畫 Task 5.5.1):任何訂單狀態皆可預覽,不觸碰 print_logs。
// 記錄類資料不軟刪除,由保留排程管理(90 天,§14.1)。預覽不寫 audit_logs(本表即記錄)。
type PrintPreview struct {
	ent.Schema
}

// Fields of the PrintPreview.
func (PrintPreview) Fields() []ent.Field {
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
		field.Int("previewed_by"),
		field.Time("previewed_at").
			Default(mutableNow),
		field.Int("file_asset_id"),
	}
}

// Indexes of the PrintPreview.
func (PrintPreview) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "target_date", "document_type"),
	}
}
