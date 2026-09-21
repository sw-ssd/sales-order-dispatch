package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SalesOrderItem 訂單明細(05 計畫 Task 1)。
// product_id nullable:手打品名無商品關聯(4.2.2),display_name 存名稱快照(D10)。
// qty/base_qty 存十進位文字(禁二進位浮點,比照 product_units conversion_rate);
// base_qty 為換算基本單位後數量(手打品名直接照填 qty)。
// company_id/department_id 冗餘以利 RLS(子表傳遞性免 JOIN)。
// 不含任何金額欄位(D12)。
type SalesOrderItem struct {
	ent.Schema
}

// Fields of the SalesOrderItem.
func (SalesOrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("sales_order_id"),
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.Int("product_id").
			Optional().
			Nillable(),
		field.String("display_name").
			NotEmpty(),
		field.String("qty").
			NotEmpty(), // 十進位文字
		field.String("unit").
			NotEmpty(), // 下單時選用單位
		field.String("base_qty").
			NotEmpty(), // 十進位文字,換算基本單位後數量
		field.Int("processing_spec_id").
			Optional().
			Nillable(),
		field.String("special_cut_note").
			Optional(),
		field.Int("warehouse_id").
			Optional().
			Nillable(),
		field.Int("sort_order").
			Default(0),
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

// Indexes of the SalesOrderItem.
func (SalesOrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sales_order_id"),
	}
}
