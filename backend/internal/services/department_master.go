// 部門級主檔(04 計畫 3.4)共用：scope 推導。
package services

import (
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

// masterScope 依身分推導部門主檔可見/操作範圍,語意等同 customerScope：
// super/company_admin 公司層(did=nil)、dept_admin/staff 限本部門、其餘拒絕。
func masterScope(id authz.Identity) (int, *int, error) {
	return customerScope(id)
}
