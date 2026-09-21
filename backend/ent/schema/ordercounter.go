package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OrderCounter 訂單編號計數器(05 計畫 Task 2, D7)。
// 每公司每來源一列:next_seq 為下一個序號,version 用於樂觀鎖(取號與建單同交易,衝突重試)。
// 不對外暴露 RPC,僅供建單流程內部使用。
type OrderCounter struct {
	ent.Schema
}

// Fields of the OrderCounter.
func (OrderCounter) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.String("source").
			NotEmpty(), // 訂單來源碼(metadicts order_source),取號軌道
		field.Int("next_seq").
			Default(1),
		field.Int("version").
			Default(0), // 樂觀鎖版本
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}

// Indexes of the OrderCounter.
func (OrderCounter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "source").Unique(),
	}
}
