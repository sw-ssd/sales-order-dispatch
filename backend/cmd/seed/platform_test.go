package main

import "testing"

// TestEntitlementValues 釘住方案權益的**值域**（審查 M-3 的 fail-open 風險）：
// plan 表填 0 時多數人的直覺是「上限 0＝不可用」，若被當成「enabled 且不限」就是靜默 fail-open；
// 打錯的 -2 同理不得變成「不限」。→ 0＝未含（enabled=false）、-1＝enabled 但不限、>0＝上限、
// 其他負值＝錯誤。
func TestEntitlementValues(t *testing.T) {
	cases := []struct {
		limit     int64
		enabled   bool
		wantLimit any
		wantErr   bool
	}{
		{limit: 50, enabled: true, wantLimit: int64(50)},
		{limit: -1, enabled: true, wantLimit: nil},
		{limit: 0, enabled: false, wantLimit: nil},
		{limit: -2, wantErr: true},
	}
	for _, c := range cases {
		enabled, limit, err := entitlementValues(c.limit)
		if c.wantErr {
			if err == nil {
				t.Errorf("limit=%d 必須回錯誤（不得當成不限）", c.limit)
			}
			continue
		}
		if err != nil {
			t.Errorf("limit=%d: %v", c.limit, err)
			continue
		}
		if enabled != c.enabled || limit != c.wantLimit {
			t.Errorf("limit=%d → (enabled=%v limit=%v)，want (enabled=%v limit=%v)",
				c.limit, enabled, limit, c.enabled, c.wantLimit)
		}
	}
}

// TestPlatformPlansAreValid 確認內建方案表本身落在合法值域內（否則 SeedPlatform 會在
// 任何寫入之前就 fail-fast，seed 整個跑不起來）。
func TestPlatformPlansAreValid(t *testing.T) {
	if err := validatePlanEntitlements(); err != nil {
		t.Fatalf("內建方案表必須合法：%v", err)
	}
}
