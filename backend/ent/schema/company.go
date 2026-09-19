package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Company holds the schema definition for the Company entity.
// 軟刪除(deleted_at,P2-A):刪除公司只標記不刪列 —— audit_logs.company_id 的 FK 才永遠
// 有主可依,且被刪公司的識別碼可被新公司重用。
type Company struct {
	ent.Schema
}

// Fields of the Company.
func (Company) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("tax_id").
			Optional(),
		field.Enum("status").
			Values("active", "inactive", "suspended").
			Default("active"),
		field.String("identifier").
			NotEmpty(), // 唯一性由 Indexes() 的部分唯一索引表達(表層 UNIQUE 無法加 WHERE)
		field.JSON("public_info", map[string]any{}).
			Optional(),
		field.JSON("capabilities", []string{}).
			Optional(),
		field.String("logo_url").
			Optional(),
		field.String("customer_code_prefix").
			Optional(), // D7:客戶編號前綴(大寫英數 1–4;全系統唯一由 Service/migration 把關)
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(P2-A)
	}
}

// Indexes of the Company.
// identifier 僅在未刪除的公司之間唯一(部分唯一索引,同 D10 的部門級主檔樣板),
// 故軟刪除後識別碼可重用。實際 DDL 由 migration 00019 收斂既有 DB。
func (Company) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("identifier").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

// Edges of the Company.
func (Company) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("departments", Department.Type),
		edge.To("users", User.Type),
	}
}
