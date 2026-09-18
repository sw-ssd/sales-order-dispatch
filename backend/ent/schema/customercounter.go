package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// CustomerCounter 客戶編號計數器(04 計畫 3.1.3, D7)。
// 每公司一列:next_seq 為下一個序號,version 用於樂觀鎖(取號與建檔同交易,衝突回滾重試)。
// 不對外暴露 RPC,僅供 CustomerService.CreateCustomer 內部使用。
type CustomerCounter struct {
	ent.Schema
}

// Fields of the CustomerCounter.
func (CustomerCounter) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id").
			Unique(), // 每公司一列(PK 語意於 migration 00013 PRIMARY KEY)
		field.Int("next_seq").
			Default(1),
		field.Int("version").
			Default(0), // 樂觀鎖版本
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}
