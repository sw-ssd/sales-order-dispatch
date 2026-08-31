package casl

import (
	"reflect"
	"strings"
)

// Rule 為一條已解析的 CASL 規則。
type Rule struct {
	Action     string
	Subject    string
	Conditions []FieldCondition // nil = 無條件
	Inverted   bool             // true = cannot
}

// Identity 為佔位符展開用的當前身分。
type Identity struct {
	UserID       string
	CompanyID    string
	DepartmentID string
	CustomerID   string
}

const (
	actionManage = "manage"
	subjectAll   = "all"
)

var placeholders = map[string]func(Identity) string{
	"${user.id}":            func(i Identity) string { return i.UserID },
	"${user.company_id}":    func(i Identity) string { return i.CompanyID },
	"${user.department_id}": func(i Identity) string { return i.DepartmentID },
	"${user.customer_id}":   func(i Identity) string { return i.CustomerID },
}

// Evaluator 持有以身分展開後的規則快照;規則須依 sort_order 升冪傳入。
type Evaluator struct {
	rules []Rule // disabled 規則以 disabledRule 標記,永不命中
}

// NewEvaluator 組裝 Evaluator;佔位符展開失敗(身分缺對應值)的規則停用。
func NewEvaluator(rules []Rule, id Identity) *Evaluator {
	out := make([]Rule, len(rules))
	for i, r := range rules {
		out[i] = Rule{Action: r.Action, Subject: r.Subject, Inverted: r.Inverted}
		disabled := false
		for _, c := range r.Conditions {
			nc := c
			if s, ok := c.Value.(string); ok {
				if fn, isPH := placeholders[s]; isPH {
					v := fn(id)
					if v == "" {
						disabled = true
						break
					}
					nc.Value = v
				}
			}
			out[i].Conditions = append(out[i].Conditions, nc)
		}
		if disabled {
			out[i].Subject = disabledRule
		}
	}
	return &Evaluator{rules: out}
}

const disabledRule = "\x00disabled"

// Can 由後往前掃規則,先命中者決定;無命中 = deny。
// instance 為 nil(type-level 檢查)時帶條件規則不命中。
func (e *Evaluator) Can(action, subject string, instance map[string]any) bool {
	for i := len(e.rules) - 1; i >= 0; i-- {
		r := e.rules[i]
		if r.Action != actionManage && r.Action != action {
			continue
		}
		if r.Subject != subjectAll && r.Subject != subject {
			continue
		}
		if len(r.Conditions) > 0 {
			if instance == nil || !conditionsMatch(r.Conditions, instance) {
				continue
			}
		}
		return !r.Inverted
	}
	return false
}

func conditionsMatch(conds []FieldCondition, instance map[string]any) bool {
	for _, c := range conds {
		v, ok := instance[c.Field]
		if !ok || !matchValue(c, v) {
			return false
		}
	}
	return true
}

func matchValue(c FieldCondition, v any) bool {
	switch c.Op {
	case OpEq:
		return reflect.DeepEqual(normalize(v), normalize(c.Value))
	case OpNe:
		return !reflect.DeepEqual(normalize(v), normalize(c.Value))
	case OpIn, OpNin:
		arr, _ := c.Value.([]any)
		found := false
		for _, item := range arr {
			if reflect.DeepEqual(normalize(v), normalize(item)) {
				found = true
				break
			}
		}
		if c.Op == OpIn {
			return found
		}
		return !found
	default: // $lt/$lte/$gt/$gte:數值或字串序比較
		cmp, ok := compareOrdered(v, c.Value)
		if !ok {
			return false
		}
		switch c.Op {
		case OpLt:
			return cmp < 0
		case OpLte:
			return cmp <= 0
		case OpGt:
			return cmp > 0
		default: // OpGte
			return cmp >= 0
		}
	}
}

// normalize 把 Go 常見數值型別統一為 float64,讓 JSON 數字與 int 可比。
func normalize(v any) any {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	default:
		return v
	}
}

func compareOrdered(a, b any) (int, bool) {
	an, aok := normalize(a).(float64)
	bn, bok := normalize(b).(float64)
	if aok && bok {
		switch {
		case an < bn:
			return -1, true
		case an > bn:
			return 1, true
		default:
			return 0, true
		}
	}
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return strings.Compare(as, bs), true
	}
	return 0, false
}
