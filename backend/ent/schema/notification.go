package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Notification 通知記錄(07 計畫 Task 4.3.1):pending → sent/read,pending → failed(終態)。
// 通知記錄不可刪除(規格 §5.4):無 deleted_at,不提供 Delete RPC,亦不軟刪除。
// RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批。
type Notification struct {
	ent.Schema
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id"),
		field.Int("user_id"), // 接收者
		field.Int("template_id").
			Optional().
			Nillable(), // 可空:手動/系統直發無範本
		field.String("channel").
			NotEmpty(), // fcm/in_app(D16)
		field.String("title").
			NotEmpty(),
		field.String("content").
			NotEmpty(),
		field.JSON("payload", map[string]any{}).
			Optional(), // 導頁資訊(訂單 ID 等),通知系統不解析
		field.String("status").
			NotEmpty().
			Default("pending"), // pending/sent/failed/read
		field.String("failure_reason").
			Optional(),
		field.Time("sent_at").
			Optional().
			Nillable(),
		field.Time("read_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(mutableNow),
	}
}

// Indexes of the Notification.
func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		// 通知中心:本人未讀 + 分頁。
		index.Fields("user_id", "status", "created_at"),
	}
}
