// Package scopecond 為 role_permissions.conditions 的條件 AST、Go 評估器與欄位白名單。
// 原始設計(D30-3)以 CASL JSON 為格式並與 @casl/ability 對賭;CASL 已於 D32 移除,
// 本套件現只服務角色權限的寫入驗證、company 範圍驗證與防鎖死判斷(不再參與授權決策)。
// 條件語意在唯讀端仍與 golden fixture testdata/cases.json 對賭。
package scopecond

import (
	"fmt"
	"sort"
)

// Op 為白名單運算子。
type Op string

const (
	OpEq  Op = "$eq"
	OpNe  Op = "$ne"
	OpIn  Op = "$in"
	OpNin Op = "$nin"
	OpLt  Op = "$lt"
	OpLte Op = "$lte"
	OpGt  Op = "$gt"
	OpGte Op = "$gte"
)

var validOps = map[Op]bool{
	OpEq: true, OpNe: true, OpIn: true, OpNin: true,
	OpLt: true, OpLte: true, OpGt: true, OpGte: true,
}

// FieldCondition 為單一欄位的一個條件。
type FieldCondition struct {
	Field string
	Op    Op
	Value any // $in/$nin 時為 []any
}

// ParseConditions 將已 unmarshal 的 conditions 物件解析為欄位條件切片。
// 裸值視為 $eq;輸出依 (field, op) 排序保證決定性。
func ParseConditions(raw map[string]any) ([]FieldCondition, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	fields := make([]string, 0, len(raw))
	for f := range raw {
		fields = append(fields, f)
	}
	sort.Strings(fields)

	var out []FieldCondition
	for _, f := range fields {
		switch v := raw[f].(type) {
		case map[string]any:
			ops := make([]string, 0, len(v))
			for op := range v {
				ops = append(ops, op)
			}
			sort.Strings(ops)
			for _, op := range ops {
				o := Op(op)
				if !validOps[o] {
					return nil, fmt.Errorf("scopecond: unknown operator %q on field %q", op, f)
				}
				val := v[op]
				if o == OpIn || o == OpNin {
					arr, ok := val.([]any)
					if !ok {
						return nil, fmt.Errorf("scopecond: %s on field %q requires array value", op, f)
					}
					val = arr
				}
				out = append(out, FieldCondition{Field: f, Op: o, Value: val})
			}
		default:
			out = append(out, FieldCondition{Field: f, Op: OpEq, Value: v})
		}
	}
	return out, nil
}
