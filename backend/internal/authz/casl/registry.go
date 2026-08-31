package casl

import (
	"fmt"
	"slices"
	"sort"
)

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
// ConditionFieldInfo 為 UI 條件建構器的欄位描述。
type ConditionFieldInfo struct {
	Field string
	Type  FieldType
	Ops   []Op
	Enum  []string
}

// ConditionFields 回傳 subject 的條件欄位白名單(依欄位名排序);未知 subject 回空。
func (r *FieldRegistry) ConditionFields(subject string) []ConditionFieldInfo {
	sd, ok := r.subjects[subject]
	if !ok {
		return nil
	}
	names := make([]string, 0, len(sd.Fields))
	for n := range sd.Fields {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]ConditionFieldInfo, 0, len(names))
	for _, n := range names {
		fd := sd.Fields[n]
		out = append(out, ConditionFieldInfo{Field: n, Type: fd.Type, Ops: fd.Ops, Enum: fd.Enum})
	}
	return out
}

// ValidateRuleConditions 驗證寫入的規則條件:欄位/運算子白名單 + enum 值合法。
func (r *FieldRegistry) ValidateRuleConditions(subject string, conds []FieldCondition) error {
	for _, c := range conds {
		fd, err := r.Field(subject, c.Field)
		if err != nil {
			return err
		}
		if !opAllowed(fd, c.Op) {
			return fmt.Errorf("casl: op %s not allowed on %s.%s", c.Op, subject, c.Field)
		}
		if fd.Type == TypeEnum {
			vals, _ := c.Value.([]any)
			if c.Op != OpIn && c.Op != OpNin {
				vals = []any{c.Value}
			}
			for _, v := range vals {
				s, ok := v.(string)
				if !ok || !slices.Contains(fd.Enum, s) {
					return fmt.Errorf("casl: invalid enum value %v on %s.%s", v, subject, c.Field)
				}
			}
		}
	}
	return nil
}

// RegisterSalesOrder 註冊 sales_order subject 的條件欄位。
// ToInstance 由 05-sales-orders 計畫於 Ent 產生碼就緒後補實作(消費 *ent.SalesOrder)。
func RegisterSalesOrder(reg *FieldRegistry) {
	idOps := []Op{OpEq, OpNe, OpIn, OpNin}
	reg.Register("sales_order", SubjectDef{
		Fields: map[string]FieldDef{
			"status":                 {Column: "status", Ops: idOps, Type: TypeEnum, Enum: []string{"pending", "processing", "completed", "cancelled", "voided"}},
			"department_id":          {Column: "department_id", Ops: idOps, Type: TypeID},
			"company_id":             {Column: "company_id", Ops: idOps, Type: TypeID},
			"customer_id":            {Column: "customer_id", Ops: idOps, Type: TypeID},
			"created_by":             {Column: "created_by", Ops: []Op{OpEq, OpNe}, Type: TypeID},
			"expected_delivery_date": {Column: "expected_delivery_date", Ops: []Op{OpEq, OpNe, OpLt, OpLte, OpGt, OpGte}, Type: TypeString},
		},
	})
}
