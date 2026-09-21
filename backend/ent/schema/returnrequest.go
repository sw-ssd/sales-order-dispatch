package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ReturnRequest 退貨申請(06 計畫 Task 1, D25):pending → approved / rejected。
// 業務實體軟刪除(D10);審核全程不修改原訂單(僅參照)。
// RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00033-00038 模式)。
type ReturnRequest struct {
	ent.Schema
}

// Fields of the ReturnRequest.
func (ReturnRequest) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id"),
		field.Int("customer_id"),
		field.Int("created_by_user_id"),
		field.String("status").
			NotEmpty().
			Default("pending"), // pending/approved/rejected
		field.String("remark").
			Optional(),
		field.Int("reviewed_by_user_id").
			Optional().
			Nillable(),
		field.Time("reviewed_at").
			Optional().
			Nillable(),
		field.String("reject_reason").
			Optional(),
		field.Int("version").
			Default(0), // 樂觀鎖
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

// Indexes of the ReturnRequest.
func (ReturnRequest) Indexes() []ent.Index {
	return []ent.Index{
		// 客戶自查(客戶 + 狀態)。
		index.Fields("company_id", "customer_id", "status"),
		// 業務待審清單(部門 + 狀態)。
		index.Fields("company_id", "department_id", "status"),
	}
}
