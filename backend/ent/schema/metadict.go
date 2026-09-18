package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Metadict 字典檔單表兩層（03 計畫 2.5.1, D11）：
// department_id IS NULL = 系統預設；非 NULL = 部門擴充。
// 軟刪除(deleted_at) + 部分唯一索引(兩道, D10)以 migration 00011 原生 SQL 實作
// （避開 PostgreSQL NULL 不參與唯一比較的陷阱）；本 schema 僅宣告查詢索引。
type Metadict struct {
	ent.Schema
}

// Fields of the Metadict.
func (Metadict) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").
			NotEmpty(), // 字典類型:unit/payment_method/settlement_method/customer_type/invoice_type/order_source
		field.String("code").
			NotEmpty(), // 程式識別,同 type + 部門範圍內唯一
		field.String("display_name").
			NotEmpty(),
		field.Int("department_id").
			Optional().
			Nillable(), // NULL = 系統預設(D11)
		field.Int("sort_order").
			Default(0),
		field.Bool("is_active").
			Default(true),
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(D10)
		field.Time("created_at").
			Default(mutableNow),
		field.Time("updated_at").
			Default(mutableNow).
			UpdateDefault(mutableNow),
	}
}

// Indexes of the Metadict.
// 查詢主力 (type, department_id, is_active)；
// (type, code, department_id) 的部分唯一約束依 D10/D11 於 migration 00011 以兩道條件索引實作。
func (Metadict) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "department_id", "is_active"),
	}
}
