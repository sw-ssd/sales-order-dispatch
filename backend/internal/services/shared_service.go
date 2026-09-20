// 服務層共用 helper(全 domain):身分/驗證/分頁/稽核。
// 因 ent 為每實體生成不同具體型別(無法共用 Go interface),採「共用流程 + per-entity 小橋接」模式:
// 此檔載 type-agnostic 的流程,各 service 以 listSource[M] 等小橋接明示 ent 型別。
package services

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// entitlementChecker 為寫入守衛所需的最小介面（consumer 端定義 —— Go 慣例「接受介面」）：
// services 只用到 CheckLimit，故介面只宣告它，測試才能注入「記錄呼叫了哪個 feature」的假物件。
// `*entitlements.Service` 與 `entitlements.Unlimited()` 都直接滿足此介面，無需轉接。
//
// 注意：RPC 守衛一律用 CheckLimit（才會帶 PLAT-3001／5002／5001 與 details，見 T4 的 Allows
// doc 分工）；租戶端投影需要的是另一組方法（Allows／Load），由該 consumer（T10）自訂自己的
// 窄介面，不併入此處（同一個介面塞兩組用途就成了胖介面）。
type entitlementChecker interface {
	CheckLimit(ctx context.Context, companyID int, feature string, delta int) error
}

// guardCompanyID 回傳配額守衛要檢查的租戶：**以身分為準**（authz.IdentityFrom(ctx)），
// 不得採用請求帶入的 company_id —— 否則可以用別人方案的額度替自己的寫入背書。
//
// 只有 super／developer 例外：他們沒有租戶範圍（在任意公司代營運），此時目標公司就是請求
// 指定的公司 —— 它的存在與操作權限已於守衛之前驗證完畢（CreateUser 的公司存在檢查、
// CreateDepartment 的 requireScope 與存在檢查）。
func guardCompanyID(id authz.Identity, requested int) int {
	if isSuperIdentity(id) {
		return requested
	}
	if own, err := parseID(id.CompanyID); err == nil {
		return own
	}
	return requested
}

// requireAuth 取得已登入身分;未登入 → AUTH-4001(對外 Unauthenticated,所有主檔方法共用)。
//
// 「未登入」的判定沿用既有語意(無任何角色):注入工具的測試以 Roles-only 身分驗「缺租戶範圍」
// (例:rls_metadict_audit_integration_test.go 的 noScope),改判 UserID 空值會把那條路徑
// 變成未登入 —— 本任務只換錯誤的產生方式,不動判定。
func requireAuth(ctx context.Context) (authz.Identity, error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return authz.Identity{}, errcode.AuthUnauthenticated.Error(nil)
	}
	return id, nil
}

// codeName 修剪 code/name 並驗證必填;空 → invalid_argument。
func codeName(code, name string) (string, string, error) {
	c := strings.TrimSpace(code)
	n := strings.TrimSpace(name)
	if c == "" || n == "" {
		return "", "", connect.NewError(connect.CodeInvalidArgument, errors.New("code 與 name 必填"))
	}
	return c, n, nil
}

// normalizePage 收斂分頁參數:page ≥ 1、page_size 落在 [1, maxPageSize](全 domain 共用)。
func normalizePage(page, pageSize int32) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return int(page), int(pageSize)
}

// trimNonEmpty 修剪並驗證單一欄位不可為空(供 update 使用);空 → errMsg。
func trimNonEmpty(v, errMsg string) (string, error) {
	t := strings.TrimSpace(v)
	if t == "" {
		return "", connect.NewError(connect.CodeInvalidArgument, errors.New(errMsg))
	}
	return t, nil
}

// listSource 為泛型分頁查詢的最小橋接:由各 service 以該實體 ent 查詢實作。
type listSource[M any] interface {
	Count(ctx context.Context) (int, error)
	Page(ctx context.Context, offset, limit int) ([]M, error)
}

// pageList 與 pageListE 以分頁查詢取得列、轉 proto 並組分頁 meta(共用 List 尾段)。
// toProto 不 error 者用 pageList;會 error 者(如 company)用 pageListE。
func pageList[M any, P any](ctx context.Context, page, pageSize int32, src listSource[M], toProto func(M) P) ([]P, *v1.Pagination, error) {
	return pageListE(ctx, page, pageSize, src, func(m M) (P, error) { return toProto(m), nil })
}

// pageListE 為 toProto 會回 error 的變體。
func pageListE[M any, P any](ctx context.Context, page, pageSize int32, src listSource[M], toProto func(M) (P, error)) ([]P, *v1.Pagination, error) {
	p, ps := normalizePage(page, pageSize)
	total, err := src.Count(ctx)
	if err != nil {
		return nil, nil, toConnectError(err)
	}
	items, err := src.Page(ctx, (p-1)*ps, ps)
	if err != nil {
		return nil, nil, toConnectError(err)
	}
	out := make([]P, 0, len(items))
	for _, it := range items {
		pp, err := toProto(it)
		if err != nil {
			return nil, nil, toConnectError(err)
		}
		out = append(out, pp)
	}
	return out, &v1.Pagination{Page: int32(p), PageSize: int32(ps), Total: int64(total)}, nil
}

// deptRefProbe 供 validateDeptMasterRef 檢查部門級主檔參照是否存在(已套租戶範圍 + 未刪除過濾由呼叫端套用)。
// ent 對各實體生成具體查詢型別,故以最小 interface 收斂(與 listSource 同思路)。
type deptRefProbe interface {
	Count(ctx context.Context) (int, error)
}

// validateDeptMasterRef 驗證部門級主檔參照:恰一列存在(範圍內且未軟刪除)。
// 供商品主檔(04 計畫 3.3)等跨主檔參照驗證;不存在/範圍外/已刪除 → invalid_argument(引用非法,
// 非權限錯誤)。呼叫端以帶 id + company/department 範圍 + DeletedAtIsNil 的查詢作為 probe。
func validateDeptMasterRef(ctx context.Context, ref string, probe deptRefProbe) error {
	n, err := probe.Count(ctx)
	if err != nil {
		return toConnectError(err)
	}
	if n == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("參照主檔不存在或已刪除: "+ref))
	}
	return nil
}

// recordAudit 寫一筆稽核(共用;D18 同事務)。
// action 為 create/update/delete;payload 依 action 語意放 after 或 before 內容。
func recordAudit(ctx context.Context, tx *ent.Tx, resource, action string, id, cid int, did *int, actor int, payload map[string]any) error {
	before, after := map[string]any(nil), payload
	if action == "delete" {
		before, after = payload, nil
	}
	return recordAuditBA(ctx, tx, resource, action, id, cid, did, actor, before, after)
}

// recordAuditBA 寫一筆含 before+after 快照的稽核(供需完整前後對照者;D18 同事務)。
func recordAuditBA(ctx context.Context, tx *ent.Tx, resource, action string, id, cid int, did *int, actor int, before, after map[string]any) error {
	return audit.Record(ctx, tx, audit.Entry{
		Action:       action,
		ResourceType: resource,
		ResourceID:   strconv.Itoa(id),
		CompanyID:    cid,
		DepartmentID: did,
		UserID:       actor,
		Before:       before,
		After:        after,
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	})
}
