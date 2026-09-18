package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CustomerContact 客戶聯絡人(04 計畫 3.2.2):每客戶多筆聯絡人,至多一筆預設(database 部分唯一索引)。
// company_id/department_id 自客戶複寫(供 RLS);軟刪除(deleted_at, D10)。
// 預設聯絡人不分類型;首筆自動為預設。
type CustomerContact struct {
	ent.Schema
}

// Fields of the CustomerContact.
func (CustomerContact) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("customer_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("name").
			NotEmpty(),
		field.String("title").
			Optional(),
		field.String("email").
			Optional(),
		field.String("phone").
			Optional(),
		field.Bool("is_default").
			Default(false),
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

// Indexes of the CustomerContact.
// 部分唯一索引 (customer_id) WHERE is_default = true AND deleted_at IS NULL 於 migration 00015 實作;
// 此處提供查詢索引(依客戶列聯絡人)。
func (CustomerContact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id"),
	}
}
