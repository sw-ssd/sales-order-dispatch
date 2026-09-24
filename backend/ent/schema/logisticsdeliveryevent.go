package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LogisticsDeliveryEvent 配送執行軌跡(D32/10.6)。
// 僅追加(比照 sales_order_events):無 updated_at/deleted_at,不提供修改刪除入口
// (00048 對 app_rw 只授 SELECT/INSERT)。
type LogisticsDeliveryEvent struct {
	ent.Schema
}

// Fields of the LogisticsDeliveryEvent.
func (LogisticsDeliveryEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int("logistics_delivery_id"),
		field.Int("company_id"),
		// created / started / completed / cancelled(10.6 狀態機事件)
		field.String("event_type").
			NotEmpty(),
		field.Int("actor_id"),
		field.String("reason").
			Optional(), // cancelled 必填語意,其餘 nullable
		field.JSON("payload", map[string]any{}).
			Optional(), // 異動摘要,如狀態前後值
		field.Time("created_at").
			Default(mutableNow),
	}
}

// Indexes of the LogisticsDeliveryEvent.
func (LogisticsDeliveryEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("logistics_delivery_id", "created_at"),
	}
}
