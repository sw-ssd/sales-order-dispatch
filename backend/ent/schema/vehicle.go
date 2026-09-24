package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Vehicle 車輛(D32/10.1):部門級;plate_no 部門內唯一(migration 部分唯一索引,D10)。
// company_id/department_id 採純欄位(RLS 注入比對),不建 edge;軟刪除(D10)。
type Vehicle struct {
	ent.Schema
}

// Fields of the Vehicle.
func (Vehicle) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("plate_no").
			NotEmpty(),
		field.String("vehicle_type").
			Optional(),
		field.String("status").
			Default("idle"), // idle / in_use / maintenance(1.0 簡版)
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(D10)
		field.Time("created_at").
			Default(mutableNow),
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}

// Indexes of the Vehicle.
func (Vehicle) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "plate_no"),
	}
}
