// NotificationService 通知中心(07 計畫 Task 4.3.3):List / MarkRead / UnreadCount。
// 僅回傳當前使用者本人的通知(user_id);記錄不可刪除;failed 不可轉 read。
package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// NotificationService 為通知中心服務。
type NotificationService struct {
	db *ent.Client
}

// NewNotificationService 建立 NotificationService。
func NewNotificationService(db *ent.Client) *NotificationService {
	return &NotificationService{db: db}
}

// RegisterNotificationService 掛到 /api/v1(租戶 session + RLS)。
func RegisterNotificationService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewNotificationServiceHandler(NewNotificationService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// ListNotifications 本人通知列表(分頁;unread_only 篩 pending/sent;channel 過濾)。
func (s *NotificationService) ListNotifications(ctx context.Context, req *connect.Request[salesorderv1.ListNotificationsRequest]) (*connect.Response[salesorderv1.ListNotificationsResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if id.Role == "customer" && strings.TrimSpace(id.CustomerID) == "" {
		// 客戶主帳號無 CustomerID 身分(僅供帳號管理,D22) → 拒絕。
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	actor, err := parseID(id.UserID)
	if err != nil {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	page, pageSize := normalizePage(req.Msg.GetPage(), req.Msg.GetPageSize())
	q := db.Notification.Query().Where(notification.UserIDEQ(actor))
	if req.Msg.GetUnreadOnly() {
		q = q.Where(notification.StatusIn("pending", "sent"))
	}
	if ch := strings.TrimSpace(req.Msg.GetChannel()); ch != "" {
		if ch != "fcm" && ch != "in_app" {
			return nil, invalidArgField("channel")
		}
		q = q.Where(notification.ChannelEQ(ch))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	unread, err := db.Notification.Query().
		Where(notification.UserIDEQ(actor), notification.StatusIn("pending", "sent")).Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	rows, err := q.Order(ent.Desc(notification.FieldCreatedAt), ent.Desc(notification.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	resp := &salesorderv1.ListNotificationsResponse{
		Page: int32(page), PageSize: int32(pageSize), Total: int32(total),
		UnreadCount: int32(unread),
	}
	for _, r := range rows {
		resp.Notifications = append(resp.Notifications, notificationView(r))
	}
	return connect.NewResponse(resp), nil
}

// MarkRead 標記已讀(僅本人;sent/pending → read;failed 不可轉;冪等)。
func (s *NotificationService) MarkRead(ctx context.Context, req *connect.Request[salesorderv1.MarkReadRequest]) (*connect.Response[salesorderv1.MarkReadResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	actor, err := parseID(id.UserID)
	if err != nil {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	ids := req.Msg.GetNotificationIds()
	if len(ids) == 0 || len(ids) > 100 {
		return nil, invalidArgField("notification_ids")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	marked := 0
	now := time.Now()
	for _, raw := range ids {
		nid, err := parseID(raw)
		if err != nil {
			return nil, invalidArgField("notification_ids")
		}
		n, err := db.Notification.Query().
			Where(notification.ID(nid), notification.UserIDEQ(actor)).Only(ctx)
		if err != nil {
			return nil, toConnectError(err) // 他人通知視同 not_found
		}
		switch n.Status {
		case "read":
			continue // 冪等略過
		case "failed":
			continue // failed 不可轉 read(略過,反映未異動)
		case "sent", "pending":
			if _, err := db.Notification.UpdateOneID(n.ID).
				SetStatus("read").SetReadAt(now).Save(ctx); err != nil {
				return nil, toConnectError(err)
			}
			marked++
		default:
			continue
		}
	}
	return connect.NewResponse(&salesorderv1.MarkReadResponse{MarkedCount: int32(marked)}), nil
}

// UnreadCount 未讀數(pending + sent)。
func (s *NotificationService) UnreadCount(ctx context.Context, req *connect.Request[salesorderv1.UnreadCountRequest]) (*connect.Response[salesorderv1.UnreadCountResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	actor, err := parseID(id.UserID)
	if err != nil {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	n, err := db.Notification.Query().
		Where(notification.UserIDEQ(actor), notification.StatusIn("pending", "sent")).Count(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.UnreadCountResponse{Count: int32(n)}), nil
}

// notificationView 轉視圖(payload 轉 JSON 字串)。
func notificationView(r *ent.Notification) *salesorderv1.NotificationView {
	v := &salesorderv1.NotificationView{
		Id: strconv.Itoa(r.ID), Channel: r.Channel, Title: r.Title,
		Content: r.Content, Status: r.Status,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
	}
	if len(r.Payload) > 0 {
		if b, err := json.Marshal(r.Payload); err == nil {
			v.Payload = string(b)
		}
	}
	if r.SentAt != nil {
		v.SentAt = r.SentAt.Format(time.RFC3339)
	}
	if r.ReadAt != nil {
		v.ReadAt = r.ReadAt.Format(time.RFC3339)
	}
	return v
}
