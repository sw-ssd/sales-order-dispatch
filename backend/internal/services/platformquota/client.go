// Package platformquota 為 phase-2 獨立 HTTP 服務下的配額守衛客戶端:product 側經由此呼叫
// platform 的 CheckLimit RPC,取代舊有的 in-process entitlements.Service 直接依賴。
//
// 邊界鍵約定:platform 側以 company_internal_id(uuid) 為通用鍵(三鍵策略),不認識業務 companies.id;
// 本客戶端在請求交易內先把 companyID 反查為 internal_id(company 本地表,不跨網路),再帶往 platform。
package platformquota

import (
	"strconv"
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/contracts/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1"
	platformv1connect "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1/platformv1connect"
)

// Client 實作 internal/services.entitlementChecker 介面(CheckLimit + CheckLimitRPC)。
type Client struct {
	rpc     platformv1connect.TenantEntitlementServiceClient
	ent     *ent.Client
	counter entitlements.Counter
}

// New 建立配額客戶端。rpc 為 platform TenantEntitlementService 的 connect client;
// ent 為產品側 ent client(同一請求交易內反查 company_internal_id 與計數 current_used);
// counter 為產品側計數器(同請求交易內數業務表,零漂移,見 services.NewEntitlementCounter)。
func New(rpc platformv1connect.TenantEntitlementServiceClient, ent *ent.Client, counter entitlements.Counter) *Client {
	return &Client{rpc: rpc, ent: ent, counter: counter}
}

// CheckLimit 為舊介面方法(productID 預設 'sales-order',單產品場景):維持 8 個 guardQuota 呼叫點
// 簽章不變的過渡;新路徑應直接呼叫 CheckLimitRPC。
func (c *Client) CheckLimit(ctx context.Context, companyID int, feature string, delta int) error {
	return c.CheckLimitRPC(ctx, companyID, "sales-order", feature, delta)
}

// CheckLimitRPC 為 phase-2 寫路徑配額預約:反查 internal_id、同請求交易內計數 current_used 後
// 帶往 platform CheckLimit RPC。計數在 product 側(零漂移、same-tx);platform 只做運算與訂閱狀態判定。
func (c *Client) CheckLimitRPC(ctx context.Context, companyID int, productID, feature string, delta int) error {
	internalID, err := c.resolveInternalID(ctx, companyID)
	if err != nil {
		return err
	}
	cur, err := c.counter.Count(ctx, companyID, feature)
	if err != nil {
		// 計數失敗一律 fail-closed(不得當 0 放行)。
		return errcode.SysInternal.Wrap(err)
	}
	_, err = c.rpc.CheckLimit(ctx, connect.NewRequest(&platformv1.CheckLimitRequest{
		ProductId:         productID,
		CompanyInternalId: internalID.String(),
		Feature:           feature,
		Delta:             int32(delta),
		CurrentUsed:       int32(cur),
	}))
	return fromConnectError(err)
}

// resolveInternalID 在產品側同一請求交易內把 companyID 反查為 company_internal_id。
func (c *Client) resolveInternalID(ctx context.Context, companyID int) (uuid.UUID, error) {
	co, err := c.ent.Company.Query().Where(company.ID(companyID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return uuid.Nil, errcode.SysNotFound.Error(map[string]string{"company_id": itoa(companyID)})
		}
		return uuid.Nil, errcode.SysInternal.Wrap(err)
	}
	return co.InternalID, nil
}

// fromConnectError 將 RPC 錯誤轉為判定層錯誤:已是 *connect.Error(platform 端 errcode 產物)原樣透傳;
// 網路層錯誤(unavailable/timeout)收斂為 SysInternal(見 plan 的 fail-open 決策待 Task 2.2 後續確認)。
func fromConnectError(err error) error {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*connect.Error); ok {
		return ce
	}
	if errors.Is(err, context.DeadlineExceeded) || isUnavailable(err) {
		return errcode.SysInternal.Error(map[string]string{"reason": "platform_rpc_unavailable"})
	}
	return errcode.SysInternal.Wrap(err)
}

func isUnavailable(err error) bool {
	if ce, ok := err.(*connect.Error); ok {
		return ce.Code() == connect.CodeUnavailable
	}
	return false
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
