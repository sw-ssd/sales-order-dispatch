// DeviceService 裝置註冊/註銷(07 計畫 Task 4.3.4)。
// 同一 token 同一時刻只屬一人(換帳轉移);註銷冪等;失效 token 清除供發送迴路呼叫。
package services

import (
	"context"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"strconv"
)

// DeviceService 為裝置服務。
type DeviceService struct {
	db *ent.Client
}

// NewDeviceService 建立 DeviceService。
func NewDeviceService(db *ent.Client) *DeviceService {
	return &DeviceService{db: db}
}

// RegisterDeviceService 掛到 /api/v1(租戶 session + RLS)。
func RegisterDeviceService(mux *http.ServeMux, db *ent.Client) {
	path, handler := salesorderv1connect.NewDeviceServiceHandler(NewDeviceService(db),
		connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(db)))
	mux.Handle(path, handler)
}

// RegisterDevice 註冊裝置(冪等;換帳轉移歸屬;同交易寫稽核)。
func (s *DeviceService) RegisterDevice(ctx context.Context, req *connect.Request[salesorderv1.RegisterDeviceRequest]) (*connect.Response[salesorderv1.RegisterDeviceResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if id.Role == "customer" && strings.TrimSpace(id.CustomerID) == "" {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	platform := strings.TrimSpace(req.Msg.GetPlatform())
	if platform != "ios" && platform != "android" && platform != "web" {
		return nil, invalidArgField("platform")
	}
	token := strings.TrimSpace(req.Msg.GetFcmToken())
	if token == "" || len(token) > 512 {
		return nil, invalidArgField("fcm_token")
	}
	actor, err := parseID(id.UserID)
	if err != nil {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	cid, err := parseID(id.CompanyID)
	if err != nil {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	now := time.Now()
	// 已存在(含軟刪除) → 同人復原 / 他人轉移。
	if dev, err := db.UserDevice.Query().Where(userdevice.FcmTokenEQ(token)).Only(ctx); err == nil {
		if dev.UserID == actor && dev.DeletedAt == nil {
			// 冪等更新。
			if _, err := db.UserDevice.UpdateOneID(dev.ID).
				SetPlatform(platform).SetLastSeenAt(now).Save(ctx); err != nil {
				return nil, toConnectError(err)
			}
			if nm := strings.TrimSpace(req.Msg.GetDeviceName()); nm != "" {
				_, _ = db.UserDevice.UpdateOneID(dev.ID).SetDeviceName(nm).Save(ctx)
			}
			return connect.NewResponse(&salesorderv1.RegisterDeviceResponse{
				DeviceId: strconv.Itoa(dev.ID),
			}), nil
		}
		// 他人或已軟刪除 → 原記錄軟刪除(若未刪),為當前使用者新建。
		if dev.DeletedAt == nil {
			if _, err := db.UserDevice.UpdateOneID(dev.ID).
				SetDeletedAt(now).Save(ctx); err != nil {
				return nil, toConnectError(err)
			}
		}
	}
	b := db.UserDevice.Create().SetUserID(actor).SetCompanyID(cid).
		SetPlatform(platform).SetFcmToken(token).SetLastSeenAt(now)
	if nm := strings.TrimSpace(req.Msg.GetDeviceName()); nm != "" {
		b = b.SetDeviceName(nm)
	}
	dev, err := b.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	var did *int
	if d, err := parseID(id.DepartmentID); err == nil {
		did = &d
	}
	if err := recordAudit(ctx, tx, "user_device", "create", dev.ID, cid, did, actor,
		map[string]any{"platform": platform}); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.RegisterDeviceResponse{
		DeviceId: strconv.Itoa(dev.ID),
	}), nil
}

// UnregisterDevice 註銷裝置(冪等;查無不報錯;他人 token 視同 not_found)。
func (s *DeviceService) UnregisterDevice(ctx context.Context, req *connect.Request[salesorderv1.UnregisterDeviceRequest]) (*connect.Response[salesorderv1.UnregisterDeviceResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	actor, err := parseID(id.UserID)
	if err != nil {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	token := strings.TrimSpace(req.Msg.GetFcmToken())
	if token == "" {
		return nil, invalidArgField("fcm_token")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	dev, err := db.UserDevice.Query().
		Where(userdevice.FcmTokenEQ(token), userdevice.UserIDEQ(actor),
			userdevice.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		// 冪等成功(含他人 token:視同不存在,不洩漏)。
		return connect.NewResponse(&salesorderv1.UnregisterDeviceResponse{}), nil
	}
	if _, err := db.UserDevice.UpdateOneID(dev.ID).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.UnregisterDeviceResponse{}), nil
}

// PurgeInvalidTokens 清除失效 token(發送迴路呼叫;不限當前使用者)。
// 系統性的清理動作不寫稽核列(呼叫端亦尚未接上;cid／actor 保留給日後補稽核的路徑)。
func PurgeInvalidTokens(ctx context.Context, db *ent.Client, tokens []string, cid int, actor int) {
	for _, tk := range tokens {
		dev, err := db.UserDevice.Query().Where(userdevice.FcmTokenEQ(tk)).Only(ctx)
		if err != nil || dev.DeletedAt != nil {
			continue
		}
		if _, err := db.UserDevice.UpdateOneID(dev.ID).SetDeletedAt(time.Now()).Save(ctx); err != nil {
			continue
		}
	}
}
