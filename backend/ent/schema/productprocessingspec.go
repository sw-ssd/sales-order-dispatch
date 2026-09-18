package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProductProcessingSpec 商品 × 處理規格多對多(04 計畫 3.3.1;規格主檔見 3.4.3)。
// attributes 為可選配對層覆寫指令(JSONB,後端不解析、列印模板渲染)。
// 關聯列隨商品整組替換(3.3.2),不設軟刪除;商品編輯時過濾已刪除規格(3.4.3 步驟 5)。
type ProductProcessingSpec struct {
	ent.Schema
}

// Fields of the ProductProcessingSpec.
func (ProductProcessingSpec) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("processing_spec_id"),
		field.JSON("attributes", map[string]any{}).
			Optional(), // 配對層覆寫指令
		field.Time("created_at").
			Default(mutableNow),
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}

// Indexes of the ProductProcessingSpec.
// (product_id, processing_spec_id) 唯一(一商品對一規格僅一筆,多對多)。
func (ProductProcessingSpec) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "processing_spec_id").Unique(),
	}
}
