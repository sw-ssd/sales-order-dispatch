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
		field.Int("customer_id").
			Optional().
			Nillable(), // D22 建檔連動:指向所屬 customers(可空,員工/管理員無)
		field.Bool("is_primary").
			Default(false), // D22:每客戶恆恰一主帳號(部分唯一索引兜底)
		field.Bool("system_generated").
			Default(false), // D22:業務子帳號由建檔自動附帶,供灰化判斷
		field.String("account_name").
			Optional(),
		field.Int("token_version").
			Default(0),
		field.String("password_hash").
			NotEmpty().
			Sensitive(),
		field.Bool("must_change_password").
			Default(false), // A3 首登強改(1.5.2):true 時僅 ChangePassword 可用
		field.Time("temp_password_expires_at").
			Optional().
			Nillable(), // A3 臨時密碼效期(now+24h, 1.5.2)
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
