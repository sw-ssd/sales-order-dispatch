package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Role holds the schema definition for the Role entity (02 計畫 2.9.1 基底)。
type Role struct {
	ent.Schema
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.Enum("data_scope").
			Values("all", "company", "department", "self").
			Default("company"),
		field.Bool("is_system").
			Default(false),
		field.Bool("is_active").
			Default(true),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}
