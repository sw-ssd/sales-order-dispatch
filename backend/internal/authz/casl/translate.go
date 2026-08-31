package casl

import (
	"fmt"
	"strings"
)

// Translate 將規則翻譯為參數化 SQL WHERE 片段(佔位符 ?,Ent sql.ExprP 依 dialect 重寫)。
// 對應 @casl/ability 的 rulesToQuery:直接規則條件 OR 併、inverted 規則 NOT(...) AND 併;
// 無任何直接規則命中 → denied=true(對應 rulesToQuery 回 null,呼叫面短路回空集)。
// 未知 subject/欄位的規則略過(fail-closed,等同不命中),由呼叫面記 security log。
func Translate(rules []Rule, action, subject string, reg *FieldRegistry) (clause string, args []any, denied bool, err error) {
	var ors, ands []string
	var orArgs, andArgs []any
	for _, r := range rules {
		if r.Action != actionManage && r.Action != action {
			continue
		}
		if r.Subject != subjectAll && r.Subject != subject {
			continue
		}
		frag, a, ferr := conditionsSQL(subject, r.Conditions, reg)
		if ferr != nil {
			continue // fail-closed:視為不命中
		}
		if r.Inverted {
			ands = append(ands, "NOT ("+frag+")")
			andArgs = append(andArgs, a...)
		} else {
			ors = append(ors, frag)
			orArgs = append(orArgs, a...)
		}
	}
	if len(ors) == 0 {
		return "", nil, true, nil
	}
	if len(ors) == 1 {
		clause = "(" + ors[0] + ")"
	} else {
		wrapped := make([]string, len(ors))
		for i, o := range ors {
			wrapped[i] = "(" + o + ")"
		}
		clause = "(" + strings.Join(wrapped, " OR ") + ")"
	}
	if len(ands) > 0 {
		clause += " AND (" + strings.Join(ands, " AND ") + ")"
	}
	args = append(orArgs, andArgs...) // 佔位符順序:直接規則在前,inverted 在後
	return clause, args, false, nil
}

// conditionsSQL 將單條規則的欄位條件 AND 併為 SQL 片段。
func conditionsSQL(subject string, conds []FieldCondition, reg *FieldRegistry) (string, []any, error) {
	if len(conds) == 0 {
		return "TRUE", nil, nil
	}
	var parts []string
	var args []any
	for _, c := range conds {
		fd, err := reg.Field(subject, c.Field)
		if err != nil {
			return "", nil, err
		}
		if !opAllowed(fd, c.Op) {
			return "", nil, fmt.Errorf("casl: op %s not allowed on %s.%s", c.Op, subject, c.Field)
		}
		col := `"` + fd.Column + `"`
		switch c.Op {
		case OpEq:
			parts = append(parts, col+" = ?")
			args = append(args, c.Value)
		case OpNe:
			parts = append(parts, col+" <> ?")
			args = append(args, c.Value)
		case OpIn, OpNin:
			arr, _ := c.Value.([]any)
			if len(arr) == 0 {
				return "", nil, fmt.Errorf("casl: %s on %s.%s requires non-empty array", c.Op, subject, c.Field)
			}
			ph := strings.TrimSuffix(strings.Repeat("?, ", len(arr)), ", ")
			kw := "IN"
			if c.Op == OpNin {
				kw = "NOT IN"
			}
			parts = append(parts, col+" "+kw+" ("+ph+")")
			args = append(args, arr...)
		default: // $lt/$lte/$gt/$gte
			sym := map[Op]string{OpLt: "<", OpLte: "<=", OpGt: ">", OpGte: ">="}[c.Op]
			parts = append(parts, col+" "+sym+" ?")
			args = append(args, c.Value)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}

func opAllowed(fd FieldDef, op Op) bool {
	for _, o := range fd.Ops {
		if o == op {
			return true
		}
	}
	return false
}
