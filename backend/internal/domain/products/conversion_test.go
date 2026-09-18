package products

import "testing"

// TestConversion 3.3.3:基本單位 kg + 換算單位「條」(0.6) — 2 條得 1.2 kg;反向 1.2 kg 得 2 條。
func TestConversion(t *testing.T) {
	rate, err := ParseRate("0.6")
	if err != nil {
		t.Fatalf("rate: %v", err)
	}
	qty, err := ParseQty("2")
	if err != nil {
		t.Fatalf("qty: %v", err)
	}
	if got := ToBase(rate, qty); got != "1.2" {
		t.Fatalf("2條×0.6 應為 1.2,got %q", got)
	}
	base, _ := ParseQty("1.2")
	if got := FromBase(rate, base); got != "2" {
		t.Fatalf("1.2kg÷0.6 應為 2,got %q", got)
	}
}

// TestConversionZeroAndNegative:數量 0 合法(結果 0);負數拒絕。
func TestConversionZeroAndNegative(t *testing.T) {
	rate, _ := ParseRate("0.6")
	zero, err := ParseQty("0")
	if err != nil {
		t.Fatalf("數量 0 應合法: %v", err)
	}
	if got := ToBase(rate, zero); got != "0" {
		t.Fatalf("0×0.6 應為 0,got %q", got)
	}
	if _, err := ParseQty("-1"); err == nil {
		t.Fatalf("負數應拒絕")
	}
}

// TestConversionDecimalExact 3.3.3 驗收:0.1 kg 級連續換算不產生浮點誤差(有理數算術)。
func TestConversionDecimalExact(t *testing.T) {
	rate, _ := ParseRate("0.1")
	qty, _ := ParseQty("3")
	// 3 條 × 0.1 = 0.3(精確,非 0.30000000000000004)
	if got := ToBase(rate, qty); got != "0.3" {
		t.Fatalf("3×0.1 應為 0.3,got %q", got)
	}
}

// TestParseRateInvalid:0 / 負 / 非法格式 → 錯誤。ParseRate 供 3.3.2 寫入驗證。
func TestParseRateInvalid(t *testing.T) {
	for _, s := range []string{"0", "-1", "abc", ""} {
		if _, err := ParseRate(s); err == nil {
			t.Fatalf("ParseRate(%q) 應回錯誤", s)
		}
	}
	// 合法：整數、小數、分數。
	for _, s := range []string{"1", "5", "0.6", "3/5"} {
		if _, err := ParseRate(s); err != nil {
			t.Fatalf("ParseRate(%q) 應合法,got %v", s, err)
		}
	}
}

// TestConversionRounding 3.3.3:四捨五入至小數 3 位。
func TestConversionRounding(t *testing.T) {
	rate, _ := ParseRate("0.333") // 1/3 近似率
	qty, _ := ParseQty("1")
	if got := ToBase(rate, qty); got != "0.333" {
		t.Fatalf("1×0.333 應為 0.333,got %q", got)
	}
	// 進位:2 × 0.3335 → 0.667。
	r2, _ := ParseRate("0.3335")
	if got := ToBase(r2, qty); got != "0.334" {
		t.Fatalf("0.3335 應四捨五入為 0.334,got %q", got)
	}
}
