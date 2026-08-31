package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Company holds the schema definition for the Company entity.
type Company struct {
	ent.Schema
}

// Fields of the Company.
func (Company) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("tax_id").
			Optional(),
		field.Enum("status").
			Values("active", "inactive", "suspended").
			Default("active"),
		field.String("identifier").
			NotEmpty().
			Unique(),
		field.JSON("public_info", map[string]any{}).
			Optional(),
		field.JSON("capabilities", []string{}).
			Optional(),
		field.String("logo_url").
			Optional(),
	}
}

// Edges of the Company.
func (Company) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("departments", Department.Type),
		edge.To("users", User.Type),
	}
}
