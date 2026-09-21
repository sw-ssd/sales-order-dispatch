package print

import "errors"

// errUnknownType 為未知單據類型錯誤(呼叫端轉 invalid_argument)。
func errUnknownType(t string) error {
	return errors.New("未知的單據類型: " + t)
}
