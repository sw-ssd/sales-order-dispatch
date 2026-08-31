package casl

import "fmt"

// FieldType 為條件欄位值型別,供驗證與 UI 建構器。
type FieldType string

const (
	TypeString FieldType = "string"
	TypeEnum   FieldType = "enum"
	TypeID     FieldType = "id"
	TypeNumber FieldType = "number"
)

// FieldDef 為單一邏輯欄位的定義。
type FieldDef struct {
	Column string    // DB column 名
	Ops    []Op      // 允許運算子
	Type   FieldType // 值型別
	Enum   []string  // TypeEnum 的合法值
}

// SubjectDef 為單一 subject 的欄位集與實例擷取器。
type SubjectDef struct {
	Fields map[string]FieldDef
	// ToInstance 由 Ent 實體擷取條件所需欄位值(供 Evaluator.Can)。
	ToInstance func(entity any) map[string]any
}

// FieldRegistry 為 subject → 定義的註冊表;SQL 翻譯、實例擷取、UI 白名單三方共用。
type FieldRegistry struct {
	subjects map[string]SubjectDef
}

func NewFieldRegistry() *FieldRegistry {
	return &FieldRegistry{subjects: map[string]SubjectDef{}}
}

func (r *FieldRegistry) Register(subject string, def SubjectDef) {
	r.subjects[subject] = def
}

// Field 查詢 subject 的欄位定義;未知 subject/欄位回 error(fail-closed 由呼叫面處理)。
func (r *FieldRegistry) Field(subject, field string) (FieldDef, error) {
	sd, ok := r.subjects[subject]
	if !ok {
		return FieldDef{}, fmt.Errorf("casl: unknown subject %q", subject)
	}
	fd, ok := sd.Fields[field]
	if !ok {
		return FieldDef{}, fmt.Errorf("casl: unknown field %q on subject %q", field, subject)
	}
	return fd, nil
}

// Instance 由實體擷取 instance map;subject 未註冊回 nil。
func (r *FieldRegistry) Instance(subject string, entity any) map[string]any {
	sd, ok := r.subjects[subject]
	if !ok || sd.ToInstance == nil {
		return nil
	}
	return sd.ToInstance(entity)
}
