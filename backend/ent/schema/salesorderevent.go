package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SalesOrderEvent 訂單異動事件(05 計畫 Task 1, D13)。
// 僅追加:無 updated_at/deleted_at,不提供修改刪除入口(DB 層 REVOKE)。
type SalesOrderEvent struct {
	ent.Schema
}

// Fields of the SalesOrderEvent.
func (SalesOrderEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int("sales_order_id"),
		field.Int("company_id"),
		// create/edit/dispatch/dispatch_cancel/cancel/complete/void(D13)
		field.String("event_type").
			NotEmpty(),
		field.Int("actor_id"),
		field.String("reason").
			Optional(), // dispatch_cancel/void 必填語意,其餘 nullable
		field.JSON("payload", map[string]any{}).
			Optional(), // 異動摘要,如狀態前後值
		field.Time("created_at").
			Default(mutableNow),
	}
}

// Indexes of the SalesOrderEvent.
func (SalesOrderEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sales_order_id", "created_at"),
	}
}
