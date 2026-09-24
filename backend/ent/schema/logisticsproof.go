package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LogisticsProof 簽收證明 POD(D32/10.6/D17):型別 photo/signature/scan,
// 實體檔案走既有 file_assets(file_asset_id 指向該列)。
// 軟刪除(D10):實體檔案保留供稽核,查詢預設排除。
type LogisticsProof struct {
	ent.Schema
}

// Fields of the LogisticsProof.
func (LogisticsProof) Fields() []ent.Field {
	return []ent.Field{
		field.Int("logistics_delivery_id"),
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		// photo / signature / scan(10.6)
		field.String("proof_type").
			NotEmpty(),
		field.Int("file_asset_id"), // file_assets.id(D17)
		field.String("remarks").
			Optional(),
		field.Int("captured_by"), // 上傳者(users.id)
		field.Time("captured_at"),
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(D10)
		field.Time("created_at").
			Default(mutableNow),
	}
}

// Indexes of the LogisticsProof.
func (LogisticsProof) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("logistics_delivery_id"),
	}
}
