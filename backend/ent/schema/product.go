package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Product 商品主檔(04 計畫 3.3.1):承載商品主列,關聯分類(3.4.4 product_categories)、
// 產品分庫/揀貨倉(3.4.1 warehouses)與處理規格(3.4.3 processing_specs)。
// code 部門內部分唯一(部分唯一索引於 migration 00017 實作);soft 軟刪除(D10)後可重建同碼。
// 單位(product_units)與處理規格關聯(product_processing_specs)不在此表,整組替換管理(3.3.2)。
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.Int("category_id").
			Optional().
			Nillable(), // 商品分類(3.4.4),寫入路徑驗證同部門未刪除
		field.Int("inventory_warehouse_id").
			Optional().
			Nillable(), // 產品分庫(3.4.1)
		field.Int("picking_warehouse_id").
			Optional().
			Nillable(), // 揀貨倉別(3.4.1)
		field.String("description").
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

// Indexes of the Product.
// 部分唯一索引 (department_id, code) WHERE deleted_at IS NULL 於 migration 00017 實作。
func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id", "code"),
	}
}
