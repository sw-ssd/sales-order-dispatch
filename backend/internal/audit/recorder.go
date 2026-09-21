// Package audit 提供稽核日誌的同事務寫入入口（03 計畫 2.6.2, D18）。
//
// 統一入口原則：各 domain usecase 不得自行組裝 audit_logs 寫入，一律經本包；
// Record 只接受交易內 client（*ent.Tx），從簽名上杜絕「業務已 commit、稽核裸寫」的
// 脫鉤寫法——稽核寫入與業務異動同一 DB 交易，同成功同失敗。
package audit

import (
	"context"
	"errors"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
)

// Entry 為一筆稽核內容（action / resource / 前後快照）。
// snapshot 採「僅收錄變更欄位摘要」，敏感欄位（密碼雜湊等）由 Record 過濾後不落盤。
type Entry struct {
	Action       string
	ResourceType string
	ResourceID   string
	CompanyID    int
	DepartmentID *int // 可空（跨部門或公司級操作）
	UserID       int
	Before       map[string]any // 可空（Create 僅 after；login/logout 兩者皆 null）
	After        map[string]any // 可空
	IPAddress    string
	UserAgent    string
	// TraceID 為發起請求的 trace_id（requestid.Interceptor 產生，未結項 #4）：
	// 同一次請求寫出的多筆稽核（含 name＋status 的兩筆）以此關聯 —— console 端
	// 自行合併的前提。呼叫端不填時由 Record 自 ctx 取（測試直呼 Record 時 ctx 無值即空）。
	TraceID string
}

// sensitiveKeys 為不得落入稽核快照的欄位鍵（密碼/雜湊/密鑰/權杖等）。
var sensitiveKeys = map[string]bool{
	"password":      true,
	"password_hash": true,
	"passwordHash":  true,
	"hash":          true,
	"secret":        true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
	"refreshToken":  true,
}

// sanitize 過濾快照中的敏感欄位（不整列複製，僅保留受允許鍵）。
func sanitize(snap map[string]any) map[string]any {
	if len(snap) == 0 {
		return nil
	}
	out := make(map[string]any, len(snap))
	for k, v := range snap {
		if sensitiveKeys[k] {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
func Record(ctx context.Context, tx *ent.Tx, e Entry) error {
	companyID := e.CompanyID
	if companyID == 0 {
		return errors.New("audit: company_id 為 0,缺租戶脈絡")
	}
	if e.UserID == 0 {
		return errors.New("audit: user_id 為 0,缺操作者脈絡")
	}
	// 未結項 #4：TraceID 未填時自 ctx 取 —— 各 domain 的 recordAuditBA 照樣只傳 ctx，
	// 不必逐點改 30+ 呼叫點；trace_id 本來就只有邊界（interceptor）知道。
	// 落點：after_snapshot 的 `_trace_id` 鍵 —— 不加 migration／不改 ent schema：
	// audit_logs 專屬欄位要動 00009＋ent＋codegen，而快照本來就是「變更摘要＋關聯鍵」
	// 的半結構欄位；console 合併只讀此鍵，不影響既有快照斷言（鍵名底線前綴防碰撞）。
	traceID := e.TraceID
	if traceID == "" {
		traceID = requestid.From(ctx)
	}
	after := sanitize(e.After)
	if traceID != "" {
		if after == nil {
			after = map[string]any{}
		}
		after["_trace_id"] = traceID
	}
	build := tx.AuditLog.Create().
		SetCompanyID(companyID).
		SetUserID(e.UserID).
		SetAction(auditlog.Action(e.Action)).
		SetResourceType(e.ResourceType).
		SetResourceID(e.ResourceID).
		SetIPAddress(e.IPAddress).
		SetUserAgent(e.UserAgent)
	if e.DepartmentID != nil {
		build = build.SetDepartmentID(*e.DepartmentID)
	}
	if before := sanitize(e.Before); before != nil {
		build = build.SetBeforeSnapshot(before)
	}
	if after != nil {
		build = build.SetAfterSnapshot(after)
	}
	_, err := build.Save(ctx)
	return err
}
