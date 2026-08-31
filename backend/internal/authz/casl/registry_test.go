package casl

import (
	"testing"
)

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestConditionFields(t *testing.T) {
	reg := NewFieldRegistry()
	RegisterSalesOrder(reg)
	fields := reg.ConditionFields("sales_order")
	if len(fields) == 0 {
		t.Fatal("want non-empty fields")
	}
	byName := map[string]ConditionFieldInfo{}
	for _, f := range fields {
		byName[f.Field] = f
	}
	assertEqual(t, TypeEnum, byName["status"].Type)
	if !containsStr(byName["status"].Enum, "pending") {
		t.Fatalf("status enum should contain pending, got %#v", byName["status"].Enum)
	}
	assertEqual(t, TypeID, byName["department_id"].Type)
	if len(reg.ConditionFields("nonexistent")) != 0 {
		t.Fatal("want empty for unknown subject")
	}
}

func TestValidateRuleConditions(t *testing.T) {
	reg := NewFieldRegistry()
	RegisterSalesOrder(reg)
	t.Run("合法", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"status": map[string]any{"$in": []any{"pending"}}})
		assertNoError(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("未知欄位", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"nope": 1.0})
		assertError(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("運算子不在該欄位白名單", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"status": map[string]any{"$lt": "x"}})
		assertError(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("enum 值非法", func(t *testing.T) {
		conds, _ := ParseConditions(map[string]any{"status": "bogus"})
		assertError(t, reg.ValidateRuleConditions("sales_order", conds))
	})
	t.Run("無條件恆合法", func(t *testing.T) {
		assertNoError(t, reg.ValidateRuleConditions("sales_order", nil))
	})
}

func containsStr(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
