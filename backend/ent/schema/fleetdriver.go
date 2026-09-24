package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FleetDriver 司機(D32/10.1):部門級,關聯既有 users(不另建帳號,10.1)。
// company_id/department_id 採純欄位(RLS 注入比對),不建 edge;軟刪除(D10)。
type FleetDriver struct {
	ent.Schema
}

// Fields of the Driver.
func (FleetDriver) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.Int("user_id"), // 關聯 users.id(司機以既有 JWT 登入,10.12)
		field.String("name").
			NotEmpty(),
		field.String("phone").
			Optional(),
		field.String("current_status").
			Default("offline"), // online / offline / on_task(10.1)
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

// Indexes of the FleetDriver.
func (FleetDriver) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "user_id"),
		index.Fields("user_id"),
	}
}
