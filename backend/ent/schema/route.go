package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Route 車次(04 計畫 3.4.2):部門級主檔,供訂單/派車選用。
// 軟刪除(D10);code 部門內部分唯一(migration 00016)。
type Route struct {
	ent.Schema
}

// Fields of the Route.
func (Route) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
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

// Indexes of the Route.
func (Route) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "code"),
	}
}
