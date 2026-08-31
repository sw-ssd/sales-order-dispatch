package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RolePermission holds the schema definition for the RolePermission entity
// (02 計畫 2.9.1 基底 + D30 CASL 三欄)。
type RolePermission struct {
	ent.Schema
}

// Fields of the RolePermission.
func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.Int("role_id"),
		field.String("resource").
			NotEmpty(),
		field.String("action").
			NotEmpty(),
		field.JSON("conditions", map[string]any{}).
			Optional(),
		field.Bool("inverted").
			Default(false),
		field.Int("sort_order").
			Default(0),
	}
}

// Edges of the RolePermission.
func (RolePermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Role.Type).
			Unique().
			Required().
			Field("role_id"),
	}
}

// Indexes of the RolePermission.
// 唯一鍵以四欄組表達 (role_id, resource, action, conditions);
// production migration 以 COALESCE(md5(conditions::text), '') 表達同義約束。
func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "resource", "action", "conditions").
			Unique(),
	}
}
