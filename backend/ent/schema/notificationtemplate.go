package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationTemplate 通知範本(07 計畫 Task 4.3.1):code + channel + locale 選取。
// 業務實體軟刪除(D10);同範圍同通道同語系以部分唯一索引保證單一啟用代號。
// RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批(比照 00033-00040 模式)。
type NotificationTemplate struct {
	ent.Schema
}

// Fields of the NotificationTemplate.
func (NotificationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(), // NULL = 公司層範本(語系退回鏈用)
		field.String("code").
			NotEmpty(), // 範本代號,如 order_created
		field.String("name").
			NotEmpty(),
		field.String("channel").
			NotEmpty(), // fcm/in_app(D16)
		field.String("subject").
			Optional(),
		field.String("body").
			NotEmpty(), // 含 {{變數}} 佔位
		field.String("locale").
			NotEmpty().
			Default("zh-Hant"),
		field.Bool("is_active").
			Default(true),
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

// Indexes of the NotificationTemplate.
func (NotificationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		// 同範圍同通道同語系單一啟用代號(部分唯一)。
		index.Fields("company_id", "department_id", "code", "channel", "locale").
			Unique(),
	}
}
