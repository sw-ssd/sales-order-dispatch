// Package money 以「分」為單位的整數運算處理金額：不用 float（二進位浮點無法精確表示
// 十進位金額），DB 邊界以字串轉換 numeric(12,2)。
//
// 金額一律不得為負——退款是獨立流程，不是負應收。
package money

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseCents 解析 "1500.00" 形式的金額字串為分；超過兩位小數或負值一律拒絕
// （帳務不得四捨五入掉使用者看不到的尾差）。
func ParseCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("金額為空")
	}
	if strings.HasPrefix(s, "-") {
		return 0, errors.New("金額不得為負")
	}
	intPart, fracPart, hasFrac := strings.Cut(s, ".")
	if intPart == "" {
		return 0, fmt.Errorf("金額格式錯誤: %q", s)
	}
	for _, r := range intPart + fracPart {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("金額格式錯誤: %q", s)
		}
	}
	if hasFrac {
		if len(fracPart) > 2 {
			return 0, fmt.Errorf("金額最多兩位小數: %q", s)
		}
		for len(fracPart) < 2 {
			fracPart += "0"
		}
	} else {
		fracPart = "00"
	}
	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("金額過大: %q", s)
	}
	cents, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("金額格式錯誤: %q", s)
	}
	// 須先扣掉分位再比：whole 恰好等於 MaxInt64/100（可通過整數部位的 ParseInt）時，
	// whole*100 已貼齊上限，再加小數位就回繞成負數——金額不得默默變負。
	if whole > (math.MaxInt64-cents)/100 {
		return 0, fmt.Errorf("金額過大: %q", s)
	}
	return whole*100 + cents, nil
}

// FormatCents 將分格式化為兩位小數字串（DB numeric(12,2) 的寫入形式）。
// 以 uint64 取絕對值：-c 對 math.MinInt64 會回繞，這裡不留那個洞。
func FormatCents(c int64) string {
	u := uint64(c)
	neg := c < 0
	if neg {
		u = -u
	}
	s := fmt.Sprintf("%d.%02d", u/100, u%100)
	if neg {
		return "-" + s
	}
	return s
}

// PeriodAmount 計算單期金額 = baseCents ＋ seatCents × seats（整數運算，溢位即錯）。
func PeriodAmount(baseCents, seatCents int64, seats int) (int64, error) {
	if seats < 0 {
		return 0, errors.New("席位數不得為負")
	}
	seatTotal, err := mulCheck(seatCents, int64(seats))
	if err != nil {
		return 0, err
	}
	return addCheck(baseCents, seatTotal)
}

// YearlyFromMonthly 由月費推導年費：月費 × 12 後套用折扣基點（1000 = 9 折），
// 結果四捨五入到分（月費 ×12 恆為偶數，故 num 的餘數不可能正好是 5000，無平手情形）。
//
// **本函式夾住折扣基點 0..10000**：<= 0 視為不打折；>= 10000 視為 100% 折扣（年費 0）。
// 折扣基點由方案價目寫入路徑提供（營運可編輯，屬外部輸入），夾住是為了不讓 10000-bps
// 轉負後把年費算成負數——負的年費等於靜默變成退款方向（金額一律不得為負）。
// **呼叫端（方案價目寫入路徑）仍必須自行驗證 0 <= discountBps <= 10000 並拒絕超界值**：
// 這裡的夾住只保證不回傳負數，不是把超界輸入當成合法設定。
//
// 月費與折扣基點都來自方案設定（非使用者輸入），且 DB 的 numeric(12,2) 上限遠低於 int64——
// 本函式沒有 error 回傳，溢位時一律「飽和」成上限值或無折扣年費，**不得回繞成負數**。
func YearlyFromMonthly(monthlyCents int64, discountBps int) int64 {
	if discountBps <= 0 {
		discountBps = 0
	}
	if discountBps >= 10000 {
		return 0
	}
	yearly, err := mulCheck(monthlyCents, 12)
	if err != nil {
		return math.MaxInt64
	}
	// x × (10000 - bps) / 10000，四捨五入：先乘後除以避免精度損失。
	num, err := mulCheck(yearly, int64(10000-discountBps))
	if err != nil {
		return yearly
	}
	return (num + 5000) / 10000
}

func addCheck(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, errors.New("金額溢位")
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, errors.New("金額溢位")
	}
	return a + b, nil
}

func mulCheck(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	// 須先擋 MinInt64 × -1：二補數下這個乘法回繞成 MinInt64，而 MinInt64/-1 依 Go 規格
	// 也回繞成 MinInt64，除以 b 驗不出溢位（唯一的偵測破口）。
	if a == math.MinInt64 && b == -1 {
		return 0, errors.New("金額溢位")
	}
	r := a * b
	if r/b != a {
		return 0, errors.New("金額溢位")
	}
	return r, nil
}
