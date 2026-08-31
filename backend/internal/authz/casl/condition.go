// Package casl 為 CASL JSON 規則的 Go 評估器與 SQL 翻譯器(D30-3)。
// 語意對齊 @casl/ability,由 testdata/cases.json golden fixture 對賭。
package casl

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
					return nil, fmt.Errorf("casl: unknown operator %q on field %q", op, f)
				}
				val := v[op]
				if o == OpIn || o == OpNin {
					arr, ok := val.([]any)
					if !ok {
						return nil, fmt.Errorf("casl: %s on field %q requires array value", op, f)
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
