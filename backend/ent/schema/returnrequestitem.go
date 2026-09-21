package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ReturnRequestItem 退貨申請明細(06 計畫 Task 1, D25):雙來源並存
// (order_item 歷史訂單品項 / customer_product 專屬商品),品項一律快照商品資訊。
// 業務實體軟刪除(D10);主檔軟刪除時品項於 usecase 層同事務一併軟刪除。
type ReturnRequestItem struct {
	ent.Schema
}

// Fields of the ReturnRequestItem.
func (ReturnRequestItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("return_request_id"),
		field.Int("company_id"),
		field.Int("department_id"), // 冗餘帶入供 RLS
		field.String("source_type").
			NotEmpty(), // order_item/customer_product
		field.Int("sales_order_id").
			Optional().
			Nillable(),
		field.Int("sales_order_item_id").
			Optional().
			Nillable(),
		field.Int("customer_product_id").
			Optional().
			Nillable(),
		field.Int("product_id"), // 一律快照
		field.String("product_name").
			NotEmpty(), // 快照:證明內容與審核當下一致
		field.String("spec").
			Optional(),
		field.String("unit").
			NotEmpty(),
		field.String("quantity").
			NotEmpty(), // 十進位字串(比照訂單 qty,禁浮點)
		field.String("reason").
			NotEmpty(), // 品項退貨原因
		field.Strings("photo_file_ids").
			Optional(), // file_assets.id 陣列
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

// Indexes of the ReturnRequestItem.
func (ReturnRequestItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("return_request_id"),
	}
}
