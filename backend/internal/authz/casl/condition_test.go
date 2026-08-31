package casl

import (
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestParseConditions(t *testing.T) {
	t.Run("裸值視為 $eq", func(t *testing.T) {
		got, err := ParseConditions(map[string]any{"status": "pending"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, []FieldCondition{{Field: "status", Op: OpEq, Value: "pending"}}, got)
	})
	t.Run("運算子物件展開，同欄位多運算子排序決定性", func(t *testing.T) {
		got, err := ParseConditions(map[string]any{"qty": map[string]any{"$lt": 5.0, "$gte": 1.0}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, []FieldCondition{
			{Field: "qty", Op: OpGte, Value: 1.0},
			{Field: "qty", Op: OpLt, Value: 5.0},
		}, got)
	})
	t.Run("$in 接受陣列", func(t *testing.T) {
		got, err := ParseConditions(map[string]any{"status": map[string]any{"$in": []any{"pending", "processing"}}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("want 1 condition, got %d", len(got))
		}
		assertEqual(t, OpIn, got[0].Op)
		assertEqual(t, []any{"pending", "processing"}, got[0].Value)
	})
	t.Run("$in 非陣列回錯", func(t *testing.T) {
		_, err := ParseConditions(map[string]any{"status": map[string]any{"$in": "pending"}})
		if err == nil {
			t.Fatal("want error for non-array $in value")
		}
	})
	t.Run("未知運算子回錯", func(t *testing.T) {
		_, err := ParseConditions(map[string]any{"x": map[string]any{"$regex": ".*"}})
		if err == nil {
			t.Fatal("want error for unknown operator")
		}
	})
	t.Run("nil/空 map 回 nil", func(t *testing.T) {
		got, err := ParseConditions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("want nil, got %#v", got)
		}
	})
}
