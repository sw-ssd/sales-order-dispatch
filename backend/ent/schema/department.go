package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Department holds the schema definition for the Department entity.
// 軟刪除(deleted_at):刪除部門只標記不刪列 —— audit_logs.department_id 的 FK 才永遠有主可依,
// 以 department_id 為 FK 的倉別/路線/加工規格/產品分類/客戶亦然(migration 00020)。
type Department struct {
	ent.Schema
}

// Fields of the Department.
func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除
	}
}

// Edges of the Department.
func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("company", Company.Type).
			Ref("departments").
			Unique().
			Required(),
		edge.To("users", User.Type),
	}
}
