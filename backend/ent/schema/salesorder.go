package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SalesOrder 銷售訂單主表(05 計畫 Task 1, D7/D10/D12/D13)。
// order_no 由 order_counters 取號(來源碼+6 位自增,D7),建立後不可改。
// 不含任何金額欄位(D12,硬性邊界);狀態機由服務層 Transition 統一守衛(D13)。
// company_id/department_id 採純欄位(RLS 注入比對),不建 edge。
type SalesOrder struct {
	ent.Schema
}

// Fields of the SalesOrder.
func (SalesOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("order_no").
			NotEmpty(),
		field.Int("customer_id"),
		field.String("source").
			NotEmpty(), // 訂單來源碼(metadicts order_source,如 W/A),建立後不可改
		field.String("status").
			NotEmpty().
			Default("pending"), // pending/processing/completed/cancelled/voided(D13)
		field.Time("expected_delivery_date").
			Optional().
			Nillable(),
		field.Int("sales_rep_id").
			Optional().
			Nillable(),
		field.String("note").
			Optional(),
		field.Time("dispatched_at").
			Optional().
			Nillable(), // 派車域(08)寫入,本域僅定義欄位
		field.Int("dispatched_by").
			Optional().
			Nillable(),
		field.Int("route_id").
			Optional().
			Nillable(), // 取消派車保留看板位置(D13),故本欄不清
		field.Int("delivery_sequence").
			Optional().
			Nillable(),
		field.Int("version").
			Default(1), // 樂觀鎖(D14 看板拖放)
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

// Indexes of the SalesOrder.
// 部分唯一索引 (company_id, order_no) WHERE deleted_at IS NULL 於 migration 實作(D10)。
func (SalesOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "order_no"),
		index.Fields("company_id", "department_id", "status", "expected_delivery_date"),
		index.Fields("customer_id"),
	}
}
