// 共用 scope 推導:業務 domain(customers / 部門級主檔)依身分決定可見/操作範圍。
package services

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

// deptScope 依身分推導可見/操作範圍,回傳(companyID, departmentID *int)。
// super/company_admin → 公司層(did=nil);dept_admin/staff → 限本部門;其餘(customer/guest)→ permission_denied。
func deptScope(id authz.Identity) (int, *int, error) {
	switch {
	case isSuperIdentity(id), hasRole(id, "company_admin"):
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
		}
		return cid, nil, nil
	case hasRole(id, "dept_admin"), hasRole(id, "staff"):
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少公司範圍"))
		}
		did, err := parseID(id.DepartmentID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("缺少部門範圍"))
		}
		return cid, &did, nil
	default:
		// customer 主帳號 / guest 一律拒絕。
		return 0, nil, connect.NewError(connect.CodePermissionDenied, errors.New("無主檔權限"))
	}
}
