package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FileAsset 檔案資產(04 計畫 Task 3.6.2):驗證通過的檔案元資料。
// filename 為系統產生(uuid + 正規化副檔名),original_filename 原樣保留;
// storage_path 為本地相對路徑(不對外暴露);url 為下載相對路徑。
// owner_type/owner_id 關聯驗證由服務層執行;company_id/department_id 採純欄位(RLS 注入比對)。
// 軟刪除(D10):實體檔案保留供稽核,下載拒絕。
type FileAsset struct {
	ent.Schema
}

// Fields of the FileAsset.
func (FileAsset) Fields() []ent.Field {
	return []ent.Field{
		field.Int("company_id"),
		field.Int("department_id").
			Optional().
			Nillable(),
		field.String("owner_type").
			NotEmpty(),
		field.Int("owner_id"),
		field.String("filename").
			NotEmpty(), // 系統產生檔名
		field.String("original_filename").
			NotEmpty(),
		field.String("mime_type").
			NotEmpty(),
		field.Int("size_bytes"),
		field.String("storage_path").
			NotEmpty(), // 本地相對路徑,不對外
		field.String("url").
			NotEmpty(), // 下載相對路徑
		field.Int("created_by").
			Optional(),
		field.Time("created_at").
			Default(mutableNow),
		field.Time("deleted_at").
			Optional().
			Nillable(), // 軟刪除(D10)
	}
}

// Indexes of the FileAsset.
func (FileAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "owner_type", "owner_id"),
	}
}
