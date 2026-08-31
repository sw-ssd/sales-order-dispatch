package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			NotEmpty().
			Unique(),
		field.String("name").
			NotEmpty(),
		field.Enum("status").
			Values("active", "inactive", "pending").
			Default("active"),
		field.String("role").
			NotEmpty(),
		field.String("phone").
			Optional(),
		field.String("employee_no").
			Optional(),
		field.Bool("is_customer").
			Default(false),
		field.String("account_name").
			Optional(),
	field.Int("token_version").
		Default(0),
		field.String("password_hash").
			NotEmpty().
			Sensitive(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("company", Company.Type).
			Ref("users").
			Unique().
			Required(),
		edge.From("department", Department.Type).
			Ref("users").
			Unique(),
	}
}
