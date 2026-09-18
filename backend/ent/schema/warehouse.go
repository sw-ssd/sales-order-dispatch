package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Warehouse 倉別(04 計畫 3.4.1):部門級主檔,供商品分庫/揀貨倉與訂單明細選用。
// company_id/department_id 由 session 租戶注入(RLS);軟刪除(deleted_at, D10)。
// code 部門內部分唯一(部分唯一索引於 migration 00016)。
type Warehouse struct {
	ent.Schema
}

// Fields of the Warehouse.
func (Warehouse) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("address").
			Optional(),
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

// Indexes of the Warehouse.
// 部分唯一索引 (department_id, code) WHERE deleted_at IS NULL 於 migration 00016 實作。
func (Warehouse) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "code"),
	}
}
