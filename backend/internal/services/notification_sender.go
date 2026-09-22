// Package services 的 FCM 發送抽象(07 計畫 Task 4.4.1/4.4.5)。
// Sender 介面供觸發方呼叫;FakeSender 供測試;HTTPClient 為最小 HTTP 發送實作骨架。
// 時序邊界:FCM 呼叫屬外部 I/O,不得包在業務 DB 交易內(先建 pending 並提交,提交後才發送)。
// 失敗不重試(D16):MarkFailed 僅 pending 生效(sent 競態略過)。
package services

import (
	"context"
	"log"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
)

// SendResult 為單通知發送結果。
type SendResult struct {
	NotificationID int
	Sent           bool
	InvalidTokens  []string // FCM 回報失效的 token(供 Purge)
	FailReason     string   // 失敗原因列舉
}

// Sender 為 FCM 發送介面(測試注入 FakeSender)。
type Sender interface {
	Send(ctx context.Context, db *ent.Client, notificationIDs []int) []SendResult
}

// FakeSender 為測試用發送器(結果由 Result 函式決定)。
type FakeSender struct {
	// OnSend 決定每筆結果;nil 時全成功。
	OnSend func(n *ent.Notification) SendResult
	Calls  int
}

// Send 執行假發送(更新通知狀態 + last_seen_at;失效 token 清除由呼叫端經 Purge)。
func (f *FakeSender) Send(ctx context.Context, db *ent.Client, notificationIDs []int) []SendResult {
	var out []SendResult
	for _, nid := range notificationIDs {
		f.Calls++
		n, err := db.Notification.Query().Where(notification.ID(nid)).Only(ctx)
		if err != nil {
			out = append(out, SendResult{NotificationID: nid, FailReason: "not_found"})
			continue
		}
		var r SendResult
		r.NotificationID = nid
		if f.OnSend != nil {
			r = f.OnSend(n)
			r.NotificationID = nid
		} else {
			r.Sent = true
		}
		applySendResult(ctx, db, n, r)
		out = append(out, r)
	}
	return out
}

// applySendResult 落發送結果:成功 → sent + sent_at;失效 token → 收集;
// 其他失敗 → failed + failure_reason。僅 pending 生效(sent 競態略過)。
func applySendResult(ctx context.Context, db *ent.Client, n *ent.Notification, r SendResult) {
	if n.Status != "pending" {
		return
	}
	now := time.Now()
	if r.Sent {
		_, _ = db.Notification.UpdateOneID(n.ID).SetStatus("sent").SetSentAt(now).Save(ctx)
		// 成功發送更新裝置 last_seen_at(供久未使用清理參考)。
		_, _ = db.UserDevice.Update().
			Where(userdevice.UserIDEQ(n.UserID), userdevice.DeletedAtIsNil()).
			SetLastSeenAt(now).Save(ctx)
		return
	}
	if len(r.InvalidTokens) > 0 {
		// 失效 token 清除由 PurgeInvalidTokens 執行(此處只標記;清除需稽核)。
		log.Printf("notify: FCM 失效 token %d 個(通知 %d)", len(r.InvalidTokens), n.ID)
	}
	MarkFailed(ctx, db, []int{n.ID}, r.FailReason)
}

// MarkFailed 標失敗(共用失敗標記函式;僅 pending 生效;failed 終態)。
func MarkFailed(ctx context.Context, db *ent.Client, ids []int, reason string) {
	if reason == "" {
		reason = "unknown"
	}
	for _, nid := range ids {
		n, err := db.Notification.Query().Where(notification.ID(nid)).Only(ctx)
		if err != nil {
			continue
		}
		if n.Status != "pending" {
			continue // sent 競態略過,不覆蓋成功
		}
		_, _ = db.Notification.UpdateOneID(n.ID).
			SetStatus("failed").SetFailureReason(reason).Save(ctx)
		log.Printf("notify: 通知 %d 標 failed(%s)", n.ID, reason)
	}
}
