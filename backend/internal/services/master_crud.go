// 部門級主檔(04 計畫 3.4)共用 CRUD 流程:身分/驗證/分頁/稽核。
// 因 ent 為每實體生成不同具體型別(無法共用 Go interface),採「共用流程 + per-entity 小橋接」模式:
// 此檔載 type-agnostic 的流程,各 service 以 masterListSource[M] 等小橋接明示 ent 型別。
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
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// masterRequireAuth 取得已登入身分;未登入 → unauthenticated(所有主檔方法共用)。
func masterRequireAuth(ctx context.Context) (authz.Identity, error) {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return authz.Identity{}, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	return id, nil
}

// masterCodeName 修剪 code/name 並驗證必填;空 → invalid_argument。
func masterCodeName(code, name string) (string, string, error) {
	c := strings.TrimSpace(code)
	n := strings.TrimSpace(name)
	if c == "" || n == "" {
		return "", "", connect.NewError(connect.CodeInvalidArgument, errors.New("code 與 name 必填"))
	}
	return c, n, nil
}

// masterTrimNonEmpty 修剪並驗證單一欄位不可為空(供 update 使用);空 → errMsg。
func masterTrimNonEmpty(v, errMsg string) (string, error) {
	t := strings.TrimSpace(v)
	if t == "" {
		return "", connect.NewError(connect.CodeInvalidArgument, errors.New(errMsg))
	}
	return t, nil
}

// masterListSource 為泛型分頁查詢的最小橋接:由各 service 以該實體 ent 查詢實作。
type masterListSource[M any] interface {
	Count(ctx context.Context) (int, error)
	Page(ctx context.Context, offset, limit int) ([]M, error)
}

// masterPage 以分頁查詢取得列、轉 proto 並組分頁 meta(共用 List 尾段)。
func masterPage[M any, P any](ctx context.Context, page, pageSize int32, src masterListSource[M], toProto func(M) P) ([]P, *v1.Pagination, error) {
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
		out = append(out, toProto(it))
	}
	return out, &v1.Pagination{Page: int32(p), PageSize: int32(ps), Total: int64(total)}, nil
}

// recordMasterAudit 寫一筆部門級主檔稽核(共用;D18 同事務)。
// action 為 create/update/delete;payload 依 action 語意放 after 或 before 內容。
func recordMasterAudit(ctx context.Context, tx *ent.Tx, resource, action string, id, cid int, did *int, actor int, payload map[string]any) error {
	e := audit.Entry{
		Action:       action,
		ResourceType: resource,
		ResourceID:   strconv.Itoa(id),
		CompanyID:    cid,
		DepartmentID: did,
		UserID:       actor,
		IPAddress:    audit.MetaFrom(ctx).IP,
		UserAgent:    audit.MetaFrom(ctx).UserAgent,
	}
	if action == "delete" {
		e.Before = payload
	} else {
		e.After = payload
	}
	return audit.Record(ctx, tx, e)
}
