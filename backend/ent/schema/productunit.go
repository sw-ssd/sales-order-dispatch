package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProductUnit 商品單位換算(04 計畫 3.3.1):一商品多組單位,恰一個基本單位(is_base),
// 其餘單位以 conversion_rate 換算為基本單位(如 1 盒 = 5 斤 → rate=5)。
// conversion_rate 存**十進位文字**(3.3.3 要求十進位算術,禁二進位浮點;以 math/big 解析計算)。
// 「每商品恰一個 is_base」由 3.3.2 寫入驗證 + migration 00017 部分唯一索引雙重保證。
// 單位列隨商品整組替換(3.3.2),不設軟刪除。
type ProductUnit struct {
	ent.Schema
}

// Fields of the ProductUnit.
func (ProductUnit) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.String("unit_code").
			NotEmpty(), // 對應 metadicts unit 字典字值
		field.String("conversion_rate").
			NotEmpty(), // 換算為基本單位之比率(十進位文字)
		field.Bool("is_base").
			Default(false),
		field.Int("sort_order").
			Default(0),
		field.String("size_desc").
			Optional(), // 單位規格描述,如「1盒=3kg」
		field.Time("created_at").
			Default(mutableNow),
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}

// Indexes of the ProductUnit.
// (product_id, unit_code) 唯一(同商品同單位僅一筆);部分唯一 is_base 於 migration 00017。
func (ProductUnit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "unit_code").Unique(),
	}
}
