package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CustomerAddress 客戶地址(04 計畫 3.2.1):每客戶多筆地址,同類型至多一筆預設(database 部分唯一索引)。
// company_id/department_id 自建檔客戶複寫(供 RLS),不信任請求攜帶值;軟刪除(deleted_at, D10)。
// type 僅接受 shipping / billing / other 三值。
type CustomerAddress struct {
	ent.Schema
}

// Fields of the CustomerAddress.
func (CustomerAddress) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("customer_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.Enum("type").
			Values("shipping", "billing", "other").
			Default("other"),
		field.String("recipient_name").
			NotEmpty(),
		field.String("phone").
			Optional(),
		field.String("address_line").
			NotEmpty(),
		field.String("city").
			Optional(),
		field.String("postal_code").
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

// Indexes of the CustomerAddress.
// 部分唯一索引 (customer_id, type) WHERE is_default = true AND deleted_at IS NULL 於 migration 00015 實作;
// 此處提供查詢索引(依客戶列地址、依客戶+類型)。
func (CustomerAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id"),
		index.Fields("customer_id", "type"),
	}
}
