package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// mutableNow 供 created_at 預設值（每行建立時取當前時間）。
func mutableNow() time.Time {
	return time.Now().UTC()
}

// AuditLog holds the schema definition for the AuditLog entity.
// 稽核日誌（03 計畫 2.6.1, D18）：記錄操作者、動作、資源、變更前後快照與來源資訊。
// 不可軟刪除、不可由 API 修改；寫入一律發生於業務交易內（同一 DB 交易，D18）。
// company_id / department_id / user_id 以純欄位記錄（不建 edge，避免改變既有實體 edges）。
type AuditLog struct {
	ent.Schema
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional(),
		field.Int("user_id"),
		field.Enum("action").
			Values("create", "update", "delete", "login", "logout", "print",
				"force_logout", "role_change", "dispatch_cancel", "void"), // §5.2 action 列舉
		field.String("resource_type"),
		field.String("resource_id").
			Optional(),
		field.JSON("before_snapshot", map[string]any{}).
			Optional(),
		field.JSON("after_snapshot", map[string]any{}).
			Optional(),
		field.String("ip_address").
			Optional(),
		field.String("user_agent").
			Optional(),
		field.Time("created_at").
			Immutable().
			Default(mutableNow),
	}
}
