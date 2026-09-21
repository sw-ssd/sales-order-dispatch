package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CustomerProduct 客戶專屬商品清單(04 計畫 Task 3.5.1):一客戶一商品一筆。
// alias_name 為客戶慣用名稱(預設商品名);default_qty 十進位文字,0 合法(保留不顯示,見 3.5.2);
// 明確不含 custom_price 或任何單價/金額欄位(1.0 單據不涉價格,D12)。
// company_id/department_id 自客戶複寫,供 RLS;軟刪除(D10)。
type CustomerProduct struct {
	ent.Schema
}

// Fields of the CustomerProduct.
func (CustomerProduct) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.Int("customer_id"),
		field.Int("product_id"),
		field.String("alias_name").
			NotEmpty(),
		field.String("default_qty").
			Default("0"), // 十進位文字,0 = 保留於清單但單據不顯示
		field.String("cut_note").
			Optional(),
		field.JSON("promo_tag_ids", []int{}).
			Default([]int{}), // D24
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

// Indexes of the CustomerProduct.
// 部分唯一索引 (customer_id, product_id) WHERE deleted_at IS NULL 於 migration 實作(D10)。
func (CustomerProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id", "product_id"),
		index.Fields("company_id", "customer_id"),
	}
}
