package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Customer 客戶主檔(04 計畫 3.1.1, D7/D10/D22)。
// customer_code 由 customer_counters 取號(公司前綴+6 位自增,D7),建立後不可改。
// 字典欄位(payment/settlement/customer_type/invoice_type)為 metadicts 參考值,不設硬 FK;
// default_sales_rep_id 指向 users,可由服務層驗證。軟刪除(deleted_at, D10)。
// company_id/department_id 採純欄位(RLS 注入比對),不建 edge。
type Customer struct {
	ent.Schema
}

// Fields of the Customer.
func (Customer) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("customer_code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("tax_id").
			Optional(),
		field.Int("payment_method_id").
			Optional().
			Nillable(),
		field.Int("settlement_method_id").
			Optional().
			Nillable(),
		field.Int("customer_type_id").
			Optional().
			Nillable(),
		field.Int("invoice_type_id").
			Optional().
			Nillable(),
		field.Int("default_sales_rep_id").
			Optional().
			Nillable(),
		field.JSON("preferred_delivery_days", []bool{}).
			Default([]bool{false, false, false, false, false, false}), // 一~六,D26
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

// Indexes of the Customer.
// 部分唯一索引 (company_id, customer_code) WHERE deleted_at IS NULL 於 migration 00013 實作(D10);
// 查詢索引:部門列表、業務反查。
func (Customer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "customer_code"),
		index.Fields("department_id"),
		index.Fields("default_sales_rep_id"),
	}
}
