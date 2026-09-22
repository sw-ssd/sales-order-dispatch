package services

import (
	"testing"
)

// TestRenderVarsMissingKeepsSource 缺漏變數保留原文 + 多餘忽略(白盒純單元)。
func TestRenderVarsMissingKeepsSource(t *testing.T) {
	out := renderVars("訂單 {{order_no}} 共 {{item_count}} 項", map[string]string{"order_no": "W000001"})
	if out != "訂單 W000001 共 {{item_count}} 項" {
		t.Fatalf("缺漏應保留原文,got %q", out)
	}
	out2 := renderVars("{{a}}{{b}}", map[string]string{"a": "1", "b": "2", "extra": "x"})
	if out2 != "12" {
		t.Fatalf("多餘應忽略,got %q", out2)
	}
	// 非變數名(連字號)不匹配。
	if out3 := renderVars("{{a-b}}", map[string]string{"a-b": "x"}); out3 != "{{a-b}}" {
		t.Fatalf("非法變數名應保留,got %q", out3)
	}
}
