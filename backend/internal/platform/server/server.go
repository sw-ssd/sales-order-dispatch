// Package server 為 platform 域的 Connect RPC 伺服器(phase-2 獨立 HTTP 服務)。
//
// 獨立服務下,product→platform 的寫路徑(CheckLimit)與租戶端唯讀投影(GetTenantEntitlements)
// 都由本包暴露;platform 自備 entitlements.Service,不依賴產品側 services 包。
package server

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/contracts/errcode"
	"github.com/salesorder/sales-order-1.0/backend/contracts/requestid"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1/platformv1connect"
)

// TenantEntitlementService 實作 platformv1connect.TenantEntitlementServiceHandler。
// 寫路徑 CheckLimit 直接轉調 entitlements.Service.CheckLimitRPC;唯讀投影
// GetTenantEntitlements 留待 session 機制(task 2.x),本階段用 Unimplemented 嵌入。
type TenantEntitlementService struct {
	ent *entitlements.Service
	platformv1connect.UnimplementedTenantEntitlementServiceHandler
}

// NewTenantEntitlementService 建立 handler。
func NewTenantEntitlementService(ent *entitlements.Service) *TenantEntitlementService {
	return &TenantEntitlementService{ent: ent}
}

// CheckLimit 為寫路徑配額預約(phase-2 獨立服務)。邊界鍵用 company_internal_id(uuid),
// 不洩漏業務 companies.id;product_id 單產品場景預設 'sales-order'。
func (s *TenantEntitlementService) CheckLimit(
	ctx context.Context, req *connect.Request[platformv1.CheckLimitRequest],
) (*connect.Response[platformv1.CheckLimitResponse], error) {
	internalID, err := uuid.Parse(req.Msg.GetCompanyInternalId())
	if err != nil {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "company_internal_id"})
	}
	if err := s.ent.CheckLimitRPC(ctx, internalID, req.Msg.GetProductId(), req.Msg.GetFeature(), int(req.Msg.GetDelta())); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&platformv1.CheckLimitResponse{}), nil
}

// toConnectError 將判定層錯誤轉為 Connect 錯誤。已是 *connect.Error(如 errcode.Code.Error
// 的產物)原樣透傳;其餘未知錯誤收斂為 SYS-1001,避免洩漏內部堆疊。
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*connect.Error); ok {
		return ce
	}
	return errcode.SysInternal.Error(nil)
}

// Register 把 TenantEntitlementService 掛到 mux(phase-2 獨立 HTTP 服務)。平台層 RPC 只需要
// requestid interceptor(填 ErrorInfo.trace_id);不掛 dbtenant(平台 schema 不受 RLS 約束,見
// AGENTS §17)。opts 可注入額外 interceptor(如 auth)。
func Register(mux *http.ServeMux, ent *entitlements.Service, opts ...connect.HandlerOption) {
	path, handler := platformv1connect.NewTenantEntitlementServiceHandler(
		NewTenantEntitlementService(ent),
		append([]connect.HandlerOption{connect.WithInterceptors(requestid.Interceptor())}, opts...)...,
	)
	mux.Handle(path, handler)
}
