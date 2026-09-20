// Task 5：addBillingPeriod 的月底處理（G2）—— 未匯出，故與實作同套件測試。
//
// 為何單獨在此檔：`time.AddDate(0, 1, 0)` 會把 1/31 正規化成 3/3（帳期跳過整個 2 月），
// 這條的輸入空間（月底、閏年、跨年、時分秒）用表驅動最快，而它不需要假 store。
package billing

import (
	"testing"
	"time"
)

// G2：月底不能靠 time.AddDate（1/31 + 1 月會正規化成 3/3，帳期跳過整個 2 月）。
// 這裡的 anchorDay 一律取 from.Day()（＝未被夾擠過的起租日），驗的是「夾擠本身」；
// 「夾擠之後不得把錨點帶走」由 lifecycle_test.go 的連續續開測試驗（I-1）。
func TestAddBillingPeriodHandlesMonthEnd(t *testing.T) {
	cases := []struct {
		from, cycle, want string
	}{
		{"2026-01-31T00:00:00Z", "monthly", "2026-02-28T00:00:00Z"},
		{"2026-01-31T03:30:15Z", "monthly", "2026-02-28T03:30:15Z"}, // 時刻保留（期末時刻不得漂移）
		{"2026-03-31T00:00:00Z", "monthly", "2026-04-30T00:00:00Z"},
		{"2028-01-31T00:00:00Z", "monthly", "2028-02-29T00:00:00Z"}, // 閏年
		{"2026-01-15T00:00:00Z", "monthly", "2026-02-15T00:00:00Z"},
		{"2026-12-15T00:00:00Z", "monthly", "2027-01-15T00:00:00Z"}, // 跨年
		{"2026-12-31T00:00:00Z", "monthly", "2027-01-31T00:00:00Z"}, // 跨年且月底
		{"2026-02-28T00:00:00Z", "yearly", "2027-02-28T00:00:00Z"},
		{"2028-02-29T00:00:00Z", "yearly", "2029-02-28T00:00:00Z"}, // 閏日遇平年
	}
	for _, tc := range cases {
		from, err := time.Parse(time.RFC3339, tc.from)
		if err != nil {
			t.Fatalf("解析 %s: %v", tc.from, err)
		}
		got, err := addBillingPeriod(from, from.Day(), tc.cycle)
		if err != nil {
			t.Fatalf("addBillingPeriod(%s, %s): %v", tc.from, tc.cycle, err)
		}
		want, _ := time.Parse(time.RFC3339, tc.want)
		if !got.Equal(want) {
			t.Fatalf("addBillingPeriod(%s, %s) = %s；want %s",
				tc.from, tc.cycle, got.Format(time.RFC3339), tc.want)
		}
	}
}

// I-1：錨點不隨 from 漂移 —— 被 2 月夾擠過的期別（from = 2/28）仍要用原錨（31）算下一期，
// 否則 1/31 → 2/28 → 3/28 → 4/28…，月繳每期縮成 28 天、一年開出 13 期。
func TestAddBillingPeriodKeepsAnchorAfterClamp(t *testing.T) {
	cases := []struct {
		from   string
		anchor int
		cycle  string
		want   string
	}{
		{"2026-02-28T00:00:00Z", 31, "monthly", "2026-03-31T00:00:00Z"},
		{"2026-04-30T00:00:00Z", 31, "monthly", "2026-05-31T00:00:00Z"},
		{"2026-02-28T00:00:00Z", 30, "monthly", "2026-03-30T00:00:00Z"},
		{"2029-02-28T00:00:00Z", 29, "yearly", "2030-02-28T00:00:00Z"}, // 平年再夾一次
		{"2032-02-29T00:00:00Z", 29, "yearly", "2033-02-28T00:00:00Z"}, // 閏年回到錨日
	}
	for _, tc := range cases {
		from, err := time.Parse(time.RFC3339, tc.from)
		if err != nil {
			t.Fatalf("解析 %s: %v", tc.from, err)
		}
		got, err := addBillingPeriod(from, tc.anchor, tc.cycle)
		if err != nil {
			t.Fatalf("addBillingPeriod(%s, %d, %s): %v", tc.from, tc.anchor, tc.cycle, err)
		}
		want, _ := time.Parse(time.RFC3339, tc.want)
		if !got.Equal(want) {
			t.Fatalf("addBillingPeriod(%s, anchor=%d, %s) = %s；want %s",
				tc.from, tc.anchor, tc.cycle, got.Format(time.RFC3339), tc.want)
		}
	}
}

// 未知與空字串的計費週期一律報錯（G1）：默默當成月繳會讓年繳少收 11 個月。
// 空字串單獨列出，因為它是「真 store 少帶 billing_cycle」的實際形狀。
func TestAddBillingPeriodRejectsUnknownCycle(t *testing.T) {
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, cycle := range []string{"weekly", "", "Monthly", "Yearly"} {
		if got, err := addBillingPeriod(from, from.Day(), cycle); err == nil {
			t.Fatalf("週期 %q 應報錯（不得默默當成月繳），got %s", cycle, got.Format(time.RFC3339))
		}
	}
}
