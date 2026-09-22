package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PromoTag 促銷分類標籤(07 計畫 Task 4.3.5, D24):部門級標籤,僅建資料層。
// CRUD RPC、客戶套用(訂閱)API、依分類選群推播皆屬 Phase 7 Task 7.4,本 Phase 不出現對應 RPC。
// 業務實體軟刪除(D10);刪除時不反向清理各宿主 promo_tag_ids(殘留 ID 選群時自然失效)。
// RLS policy 僅「定義」不 ENABLE/FORCE:ENABLE 與服務收斂同批。
type PromoTag struct {
	ent.Schema
}

// Fields of the PromoTag.
func (PromoTag) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id"), // 部門級(D24)
		field.String("code").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
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

// Indexes of the PromoTag.
func (PromoTag) Indexes() []ent.Index {
	return []ent.Index{
		// 部門內代號唯一(部分唯一,排除軟刪除由 migration 保證)。
		index.Fields("company_id", "department_id", "code").
			Unique(),
	}
}
