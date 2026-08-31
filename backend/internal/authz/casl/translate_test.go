package casl

import (
	"encoding/json"
	"testing"
)

func orderRegistry() *FieldRegistry {
	r := NewFieldRegistry()
	r.Register("sales_order", SubjectDef{Fields: map[string]FieldDef{
		"status":        {Column: "status", Ops: []Op{OpEq, OpNe, OpIn, OpNin}, Type: TypeEnum, Enum: []string{"pending", "processing", "completed", "cancelled", "voided"}},
		"department_id": {Column: "department_id", Ops: []Op{OpEq, OpNe, OpIn, OpNin}, Type: TypeID},
		"created_by":    {Column: "created_by", Ops: []Op{OpEq, OpNe}, Type: TypeID},
	}})
	return r
}

func TestTranslate(t *testing.T) {
	reg := orderRegistry()
	read := func(raw map[string]any, inverted bool) Rule {
		c, err := ParseConditions(raw)
		if err != nil {
			t.Fatal(err)
		}
		return Rule{Action: "read", Subject: "sales_order", Conditions: c, Inverted: inverted}
	}
	t.Run("單一直接規則", func(t *testing.T) {
		clause, args, denied, err := Translate([]Rule{read(map[string]any{"status": "pending"}, false)}, "read", "sales_order", reg)
		if err != nil {
			t.Fatal(err)
		}
		assertFalse(t, denied)
		assertEqual(t, `("status" = ?)`, clause)
		assertEqual(t, []any{"pending"}, args)
	})
	t.Run("多直接規則 OR 併 + inverted AND NOT 併", func(t *testing.T) {
		clause, args, denied, err := Translate([]Rule{
			read(map[string]any{"status": map[string]any{"$in": []any{"pending", "processing"}}}, false),
			read(map[string]any{"created_by": "u1"}, false),
			read(map[string]any{"status": "voided"}, true),
		}, "read", "sales_order", reg)
		if err != nil {
			t.Fatal(err)
		}
		assertFalse(t, denied)
		assertEqual(t, `(("status" IN (?, ?)) OR ("created_by" = ?)) AND (NOT ("status" = ?))`, clause)
		assertEqual(t, []any{"pending", "processing", "u1", "voided"}, args)
	})
	t.Run("action/subject 不符的規則略過", func(t *testing.T) {
		_, _, denied, err := Translate([]Rule{
			{Action: "update", Subject: "sales_order"},
		}, "read", "sales_order", reg)
		if err != nil {
			t.Fatal(err)
		}
		assertTrue(t, denied)
	})
	t.Run("無條件直接規則 = 全放行(true)", func(t *testing.T) {
		clause, args, denied, err := Translate([]Rule{{Action: "read", Subject: "sales_order"}}, "read", "sales_order", reg)
		if err != nil {
			t.Fatal(err)
		}
		assertFalse(t, denied)
		assertEqual(t, "(TRUE)", clause)
		assertEmpty(t, args)
	})
	t.Run("未知欄位規則 fail-closed 略過", func(t *testing.T) {
		_, _, denied, err := Translate([]Rule{
			read(map[string]any{"nonexistent": "x"}, false),
		}, "read", "sales_order", reg)
		if err != nil {
			t.Fatal(err)
		}
		assertTrue(t, denied) // 唯一規則被略過 → 無允許規則
	})
	t.Run("manage/all 保留字參與聚合", func(t *testing.T) {
		clause, _, denied, err := Translate([]Rule{{Action: "manage", Subject: "all"}}, "read", "sales_order", reg)
		if err != nil {
			t.Fatal(err)
		}
		assertFalse(t, denied)
		assertEqual(t, "(TRUE)", clause)
	})
}

// TestGoldenTranslate 比對 cases.json 的 query 聚合形狀(denied / $or 數 / $and 數)。
func TestGoldenTranslate(t *testing.T) {
	reg := orderRegistry()
	for _, c := range loadGolden(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var rules []Rule
			for _, rr := range c.Rules {
				var jr struct {
					Action     string         `json:"action"`
					Subject    string         `json:"subject"`
					Conditions map[string]any `json:"conditions"`
					Inverted   bool           `json:"inverted"`
				}
				if err := json.Unmarshal(rr, &jr); err != nil {
					t.Fatal(err)
				}
				conds, err := ParseConditions(jr.Conditions)
				if err != nil {
					t.Fatal(err)
				}
				rules = append(rules, Rule{Action: jr.Action, Subject: jr.Subject, Conditions: conds, Inverted: jr.Inverted})
			}
			id := Identity{c.Identity.UserID, c.Identity.CompanyID, c.Identity.DepartmentID, c.Identity.CustomerID}
			rules = NewEvaluator(rules, id).rules // 佔位符展開後再翻譯
			for _, chk := range c.Checks {
				_, _, denied, err := Translate(rules, chk.Action, chk.Subject, reg)
				if err != nil {
					t.Fatal(err)
				}
				if denied != chk.Query.Denied {
					t.Errorf("denied for %s/%s = %v, want %v", chk.Action, chk.Subject, denied, chk.Query.Denied)
				}
			}
		})
	}
}

func assertEmpty(t *testing.T, args []any) {
	t.Helper()
	if len(args) != 0 {
		t.Fatalf("want empty, got %#v", args)
	}
}
