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

// Record 於 t 所在的交易內寫入一筆 audit_logs。
// 呼叫方必須已在業務交易中（tx），寫入與業務異動共用同一 commit / rollback（D18）。
// 稽核寫入失敗 → 回傳錯誤，呼叫方應使整個交易回滾（不降級為略過）。
func Record(ctx context.Context, tx *ent.Tx, e Entry) error {
	companyID := e.CompanyID
	if companyID == 0 {
		return errors.New("audit: company_id 為 0,缺租戶脈絡")
	}
	if e.UserID == 0 {
		return errors.New("audit: user_id 為 0,缺操作者脈絡")
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
	if after := sanitize(e.After); after != nil {
		build = build.SetAfterSnapshot(after)
	}
	_, err := build.Save(ctx)
	return err
}
