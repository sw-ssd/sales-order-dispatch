package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Announcement 公告(spec announcements):Banner/最新消息/圖文文章三型別,
// 三層發佈範圍(company/department 皆 NULL = 全系統)、上下架時間窗、平台投放。
// company_id/department_id 採純欄位(RLS 注入比對),不建 edge;軟刪除(D10)。
type Announcement struct {
	ent.Schema
}

// Fields of the Announcement.
func (Announcement) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id").
			Optional().
			Nillable(), // NULL = 全系統公告(僅 super 可寫,服務層守衛)
		field.Int("department_id").
			Optional().
			Nillable(), // NULL = 全系統或公司層公告
		field.String("type").
			NotEmpty(), // banner / news / article(驗證在服務層,非法值 InvalidArgument)
		field.String("title").
			NotEmpty(),
		field.String("content").
			Default(""),
		field.String("image_url").
			Optional(), // 公告圖(經 file-assets 上傳,WebP 由該域負責)
		field.String("link_url").
			Optional(), // banner 點擊導向
		field.Time("publish_at").
			Default(mutableNow), // 上架時間(前台: <= now 才顯示)
		field.Time("unpublish_at").
			Optional().
			Nillable(), // 下架時間(NULL = 不自動下架)
		field.Int("sort_order").
			Default(0), // 同型別排序
		field.Bool("is_active").
			Default(true),
		field.Bool("deploy_web").
			Default(true), // 平台投放(Web 中台)
		field.Bool("deploy_app").
			Default(true), // 平台投放(App 首頁)
		field.Int("created_by").
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

// Indexes of the Announcement.
func (Announcement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "department_id"),
		index.Fields("type", "sort_order"),
	}
}
