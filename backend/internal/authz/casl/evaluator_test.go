package casl

import (
	"testing"
)

var staffIdentity = Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1"}

func mustConds(t *testing.T, raw map[string]any) []FieldCondition {
	t.Helper()
	c, err := ParseConditions(raw)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func assertTrue(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Fatal("want true, got false")
	}
}

func assertFalse(t *testing.T, got bool) {
	t.Helper()
	if got {
		t.Fatal("want false, got true")
	}
}

func TestEvaluatorCan(t *testing.T) {
	t.Run("無規則 deny", func(t *testing.T) {
		e := NewEvaluator(nil, staffIdentity)
		assertFalse(t, e.Can("read", "sales_order", nil))
	})
	t.Run("無條件規則命中", func(t *testing.T) {
		e := NewEvaluator([]Rule{{Action: "read", Subject: "sales_order"}}, staffIdentity)
		assertTrue(t, e.Can("read", "sales_order", nil))
		assertFalse(t, e.Can("update", "sales_order", nil))
	})
	t.Run("manage/all 保留字", func(t *testing.T) {
		e := NewEvaluator([]Rule{{Action: "manage", Subject: "all"}}, staffIdentity)
		assertTrue(t, e.Can("delete", "customer", map[string]any{"x": 1.0}))
	})
	t.Run("instance 條件比對", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "cancel", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": "pending"}),
		}}, staffIdentity)
		assertTrue(t, e.Can("cancel", "sales_order", map[string]any{"status": "pending"}))
		assertFalse(t, e.Can("cancel", "sales_order", map[string]any{"status": "processing"}))
	})
	t.Run("type-level 檢查忽略帶條件規則", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "cancel", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": "pending"}),
		}}, staffIdentity)
		assertFalse(t, e.Can("cancel", "sales_order", nil))
	})
	t.Run("cannot 排後者優先", func(t *testing.T) {
		e := NewEvaluator([]Rule{
			{Action: "read", Subject: "sales_order"},
			{Action: "read", Subject: "sales_order", Inverted: true,
				Conditions: mustConds(t, map[string]any{"status": "voided"})},
		}, staffIdentity)
		assertTrue(t, e.Can("read", "sales_order", map[string]any{"status": "pending"}))
		assertFalse(t, e.Can("read", "sales_order", map[string]any{"status": "voided"}))
	})
	t.Run("佔位符以身分展開", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"department_id": "${user.department_id}"}),
		}}, staffIdentity)
		assertTrue(t, e.Can("read", "sales_order", map[string]any{"department_id": "d1"}))
		assertFalse(t, e.Can("read", "sales_order", map[string]any{"department_id": "d2"}))
	})
	t.Run("佔位符展開失敗規則停用", func(t *testing.T) {
		customer := Identity{UserID: "u9", CompanyID: "c1", CustomerID: "cu1"} // 無 department
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"department_id": "${user.department_id}"}),
		}}, customer)
		assertFalse(t, e.Can("read", "sales_order", map[string]any{"department_id": "d1"}))
	})
	t.Run("instance 缺欄位不命中", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": "pending"}),
		}}, staffIdentity)
		assertFalse(t, e.Can("read", "sales_order", map[string]any{"other": 1.0}))
	})
	t.Run("$in 比對", func(t *testing.T) {
		e := NewEvaluator([]Rule{{
			Action: "read", Subject: "sales_order",
			Conditions: mustConds(t, map[string]any{"status": map[string]any{"$in": []any{"pending", "processing"}}}),
		}}, staffIdentity)
		assertTrue(t, e.Can("read", "sales_order", map[string]any{"status": "processing"}))
		assertFalse(t, e.Can("read", "sales_order", map[string]any{"status": "completed"}))
	})
}
