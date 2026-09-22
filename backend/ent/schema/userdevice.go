package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserDevice 裝置 FCM token(07 計畫 Task 4.3.1):同一 token 同一時刻只屬一人。
// 業務實體軟刪除(D10);換帳登入原記錄軟刪除 + 為新使用者新建。
// RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批。
type UserDevice struct {
	ent.Schema
}

// Fields of the UserDevice.
func (UserDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("company_id"),
		field.String("platform").
			NotEmpty(), // ios/android/web
		field.String("fcm_token").
			NotEmpty(),
		field.String("device_name").
			Optional(),
		field.Time("last_seen_at").
			Optional().
			Nillable(),
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

// Indexes of the UserDevice.
func (UserDevice) Indexes() []ent.Index {
	return []ent.Index{
		// 同一 token 全表唯一(部分唯一,排除軟刪除)。
		index.Fields("fcm_token").
			Unique(),
		// 使用者裝置查詢。
		index.Fields("user_id"),
	}
}
