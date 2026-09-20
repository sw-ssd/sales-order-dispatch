package money_test

import (
	"math"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
)

func TestParseCentsRejectsAmbiguity(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"0", 0, false},
		{"1500", 150000, false},
		{"1500.5", 150050, false},
		{"1500.05", 150005, false},
		{"1500.005", 0, true}, // 超過兩位小數：拒絕，不得默默四捨五入（帳務不得失真）
		{"", 0, true},
		{"abc", 0, true},
		{"-10.00", 0, true}, // 金額不得為負（退款是獨立流程，不是負應收）
		// int64 邊界：整數部位 <= MaxInt64/100 仍可能一乘就回繞成負數，兩個方向都要釘住。
		{"92233720368547757.99", 9223372036854775799, false},
		{"92233720368547758.00", 9223372036854775800, false},
		{"92233720368547758.99", 0, true},
	}
	for _, tc := range cases {
		got, err := money.ParseCents(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseCents(%q) 應報錯，got %d", tc.in, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Fatalf("ParseCents(%q) = %d, err=%v；want %d", tc.in, got, err, tc.want)
		}
	}
}

func TestFormatCentsRoundTrip(t *testing.T) {
	for _, in := range []string{"0.00", "1500.00", "1500.05", "12.30"} {
		c, err := money.ParseCents(in)
		if err != nil {
			t.Fatalf("ParseCents(%q): %v", in, err)
		}
		if got := money.FormatCents(c); got != in {
			t.Fatalf("FormatCents(%d) = %q；want %q", c, got, in)
		}
	}
	// 取負對 math.MinInt64 會回繞，格式不得因此損壞（修前為 "-92233720368547758.-8"）。
	if got := money.FormatCents(math.MinInt64); got != "-92233720368547758.08" {
		t.Fatalf("FormatCents(math.MinInt64) = %q；want \"-92233720368547758.08\"", got)
	}
}

// 期別金額 = 月費（或年費）＋ 席位單價 × 席位數；整數運算，不得溢位。
func TestPeriodAmount(t *testing.T) {
	got, err := money.PeriodAmount(150000, 15000, 10) // 1500 + 150×10 = 3000
	if err != nil || got != 300000 {
		t.Fatalf("PeriodAmount = %d, err=%v；want 300000", got, err)
	}
	if _, err := money.PeriodAmount(100, 100, -1); err == nil {
		t.Fatal("負席位數應報錯")
	}
	if _, err := money.PeriodAmount(math.MaxInt64, math.MaxInt64, 2); err == nil {
		t.Fatal("溢位應報錯，不得默默回繞")
	}
	if _, err := money.PeriodAmount(math.MaxInt64, 1, 1); err == nil {
		t.Fatal("相加溢位應報錯，不得默默回繞")
	}
}

// 年繳 = 月費 × 12 × (1 - 折扣基點/10000)，四捨五入到分。
func TestYearlyFromMonthly(t *testing.T) {
	if got := money.YearlyFromMonthly(150000, 1000); got != 1620000 { // 1500×12×0.9
		t.Fatalf("年繳 = %d；want 1620000", got)
	}
	if got := money.YearlyFromMonthly(100000, 0); got != 1200000 {
		t.Fatalf("無折扣年繳 = %d；want 1200000", got)
	}
	if got := money.YearlyFromMonthly(12345, 1000); got != 133326 { // 123.45×12×0.9 = 1333.26
		t.Fatalf("四捨五入到分 = %d；want 133326", got)
	}
	// 不足一分的尾差向上取到分（帳務不得無條件捨去）。
	if got := money.YearlyFromMonthly(1, 1000); got != 11 { // 0.12×0.9 = 0.108 → 0.11
		t.Fatalf("尾差進位 = %d；want 11", got)
	}
	// 折扣基點由方案價目寫入路徑提供（營運可編輯）：>= 10000 一律視為 100% 折扣，年費 0；
	// 修前 10000-bps 轉負會讓年費靜默變成負值（退款方向），違反本套件不變式。
	for _, bps := range []int{10000, 10001, 12000, 20000} {
		if got := money.YearlyFromMonthly(1200000, bps); got != 0 {
			t.Fatalf("YearlyFromMonthly(1200000, %d) = %d；want 0（金額不得為負）", bps, got)
		}
	}
	// 負折扣基點 = 不打折（沿用既有語意）。
	if got := money.YearlyFromMonthly(1200000, -500); got != 14400000 {
		t.Fatalf("負折扣基點年繳 = %d；want 14400000", got)
	}
}
