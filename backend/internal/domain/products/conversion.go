// Package products 商品主檔(04 計畫 3.3)。本檔為 3.3.3 單位換算純邏輯:
// 十進位有理數算術(禁二進位浮點),供 05 下單端(base_qty)/揀貨/加工彙總消費。
package products

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// baseRoundScale 輸出四捨五入至小數 3 位(供彙總一致性,見 3.3.3 步驟 2)。
const baseRoundScale = 1000

// ParseRate 解析換算率文字(十進位)為有理數;非法或非正 → 錯誤。
// 換算率為「指定單位數量 × 本率 = 基本單位數量」之比率(如 1 盒=5 斤 → "5")。
func ParseRate(s string) (*big.Rat, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return nil, fmt.Errorf("換算率格式非法: %q", s)
	}
	if r.Sign() <= 0 {
		return nil, errors.New("換算率須為正數")
	}
	return r, nil
}

// ParseQty 解析十進位數量;0 合法,負數拒絕。
func ParseQty(s string) (*big.Rat, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return nil, fmt.Errorf("數量格式非法: %q", s)
	}
	if r.Sign() < 0 {
		return nil, errors.New("數量不可為負")
	}
	return r, nil
}

// ToBase 換算:指定單位數量 × 換算率 = 基本單位數量,四捨五入至小數 3 位回十進位文字。
func ToBase(rate, qty *big.Rat) string {
	return ratToDecimal3(new(big.Rat).Mul(qty, rate))
}

// FromBase 反向換算:基本單位數量 ÷ 換算率 = 指定單位數量,四捨五入至小數 3 位回十進位文字。
func FromBase(rate, baseQty *big.Rat) string {
	return ratToDecimal3(new(big.Rat).Quo(baseQty, rate))
}

// ratToDecimal3 有理數轉十進位文字:先乘 1000 四捨五入至整數千分位,再除以 1000 並去掉尾零。
// 數量非負(0/正),四捨五入採「半值進位」(round half away from zero 對非負即 floor(x+0.5))。
func ratToDecimal3(r *big.Rat) string {
	if r.Sign() == 0 {
		return "0"
	}
	// scaled = r × 1000 = (num×1000)/den
	num := new(big.Int).Mul(new(big.Int).Set(r.Num()), big.NewInt(baseRoundScale))
	den := r.Denom()
	// rounded = floor((2·num + den) / (2·den))
	twoNum := new(big.Int).Mul(num, big.NewInt(2))
	twoNum.Add(twoNum, den)
	twoDen := new(big.Int).Mul(den, big.NewInt(2))
	thousandths := new(big.Int).Quo(twoNum, twoDen)
	return formatDecimal3(thousandths)
}

// formatDecimal3 將整數千分位格式化為十進位文字(最多 3 位小數,去尾零)。
func formatDecimal3(thousandths *big.Int) string {
	neg := false
	v := new(big.Int).Set(thousandths)
	if v.Sign() < 0 {
		neg = true
		v.Abs(v)
	}
	whole := new(big.Int).Quo(v, big.NewInt(baseRoundScale))
	frac := new(big.Int).Rem(v, big.NewInt(baseRoundScale))
	if frac.Sign() == 0 {
		if neg {
			return "-" + whole.String()
		}
		return whole.String()
	}
	fracStr := strings.TrimRight(fmt.Sprintf("%03d", frac), "0")
	out := whole.String() + "." + fracStr
	if neg {
		out = "-" + out
	}
	return out
}
