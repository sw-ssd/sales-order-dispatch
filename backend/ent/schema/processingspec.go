package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProcessingSpec 加工/處理規格(04 計畫 3.4.3,泛化自 cutting_specs 2026-09-19):
// 商品無關、普適多商品的處理要求(切/修/醃/包裝…),供商品多對多關聯(3.3)與訂單明細選用。
// kind 為開放集(metadicts processing_kind 背書);applies_to_processing/picking 為多值旗標;
// attributes 為不透明結構指令(後端不解析、列印模板渲染)。軟刪除(D10)。
type ProcessingSpec struct {
	ent.Schema
}

// Fields of the ProcessingSpec.
func (ProcessingSpec) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("kind").
			Default("other"), // 開放集,metadicts processing_kind 背書
		field.Bool("applies_to_processing").
			Default(false),
		field.Bool("applies_to_picking").
			Default(false),
		field.JSON("attributes", map[string]any{}).
			Optional(), // 自由結構指令,後端不解析
		field.Int("sort_order").
			Default(0),
		field.Bool("is_active").
			Default(true),
		field.Int("created_by").
			Optional(),
		field.Int("updated_by").
			Optional(),
		field.Time("created_at").
			Default(mutableNow),
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(D10)
	}
}

// Indexes of the ProcessingSpec.
func (ProcessingSpec) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "code"),
	}
}
