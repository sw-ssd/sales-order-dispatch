// ReturnService 審核與證明(06 計畫 Task 4.7.3–4.7.4, D25)。
// 審核僅主責業務(dept 客戶 default_sales_rep)/dept_admin/company_admin;
// 樂觀鎖 version;僅 pending 可審;同交易寫稽核;全程不碰原訂單;
// 證明僅 approved,快照內容,唯讀不寫稽核。
package services

import (
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"context"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequestitem"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// ReviewReturnRequest 審核(approved/rejected + 樂觀鎖;同交易寫稽核)。
func (s *ReturnService) ReviewReturnRequest(ctx context.Context, req *connect.Request[salesorderv1.ReviewReturnRequestRequest]) (*connect.Response[salesorderv1.ReviewReturnRequestResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	// 客戶帳號不得審核。
	if id.Role == "customer" {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	cid, did, err := deptScope(id)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, invalidArgField("id")
	}
	decision := strings.TrimSpace(req.Msg.GetDecision())
	if decision != "approved" && decision != "rejected" {
		return nil, invalidArgField("decision")
	}
	reason := strings.TrimSpace(req.Msg.GetRejectReason())
	if decision == "rejected" && reason == "" {
		return nil, invalidArgField("reject_reason")
	}
	expVer, err := parseVersion(req.Msg.GetExpectedVersion())
	if err != nil {
		return nil, invalidArgField("expected_version")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	rr, err := db.ReturnRequest.Query().
		Where(returnrequest.ID(rid), returnrequest.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if rr.CompanyID != cid || (did != nil && rr.DepartmentID != *did) {
		return nil, errcode.SysNotFound.Error(nil)
	}
	if rr.Status != "pending" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "僅待審核可審"})
	}
	// 審核權:主責業務(客戶 default_sales_rep)或 dept_admin 以上。
	actor, _ := parseID(id.UserID)
	allowed, err := s.canReview(ctx, db, cid, did, actor, id.Role, rr.CustomerID)
	if err != nil {
		return nil, toConnectError(err)
	}
	if !allowed {
		return nil, errcode.SysPermissionDenied.Error(nil)
	}
	// 樂觀鎖:version 比對 + 條件更新。
	if rr.Version != expVer {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "資料已變更，請重新載入"})
	}
	now := time.Now()
	upd := db.ReturnRequest.UpdateOneID(rr.ID).
		SetStatus(decision).SetReviewedByUserID(actor).SetReviewedAt(now).
		SetVersion(rr.Version + 1)
	if decision == "rejected" {
		upd = upd.SetRejectReason(reason)
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	after := map[string]any{"decision": decision}
	if decision == "rejected" {
		after["reject_reason"] = reason
	}
	if err := recordAudit(ctx, tx, "return_request", "update", rr.ID, cid, did, actor, after); err != nil {
		return nil, toConnectError(err)
	}
	// 稽核 action 用 update(audit_logs.action 列舉無 review;decision 存 after.decision)。
	// 通知觸發(4.7.5/D23):僅推發起帳號;同交易建 pending + AfterCommit 發送。
	if err := OnReturnReviewed(ctx, db, cid, did, rr.CreatedByUserID, decision, reason, rr.ID); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&salesorderv1.ReviewReturnRequestResponse{
		Id: strconv.Itoa(updated.ID), Status: decision,
		ReviewedAt: now.Format(time.RFC3339),
	}), nil
}

// canReview 審核權:company_admin/super/dept_admin 可審;staff 僅該客戶主責業務可審。
func (s *ReturnService) canReview(ctx context.Context, db *ent.Client, cid int, did *int, actor int, role string, custID int) (bool, error) {
	switch {
	case isSuperIdentity(authz.Identity{Role: role}), role == "company_admin", role == "dept_admin":
		return true, nil
	case role == "staff":
		cust, err := db.Customer.Query().
			Where(customer.ID(custID), customer.DeletedAtIsNil()).Only(ctx)
		if err != nil {
			return false, err
		}
		return cust.DefaultSalesRepID != nil && *cust.DefaultSalesRepID == actor, nil
	default:
		return false, nil
	}
}

// GetReturnCertificate 退貨證明(僅 approved;快照內容;唯讀不寫稽核)。
func (s *ReturnService) GetReturnCertificate(ctx context.Context, req *connect.Request[salesorderv1.GetReturnCertificateRequest]) (*connect.Response[salesorderv1.GetReturnCertificateResponse], error) {
	id, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rid, err := parseID(req.Msg.GetId())
	if err != nil {
		return nil, invalidArgField("id")
	}
	tx, ok := dbtenant.TxFrom(ctx)
	if !ok {
		return nil, errcode.SysInternal.Error(nil)
	}
	db := tx.Client()
	q := db.ReturnRequest.Query().Where(returnrequest.ID(rid), returnrequest.DeletedAtIsNil())
	if id.Role == "customer" {
		if strings.TrimSpace(id.CustomerID) == "" {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		custID, err := parseID(id.CustomerID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		cid, err := parseID(id.CompanyID)
		if err != nil {
			return nil, errcode.SysPermissionDenied.Error(nil)
		}
		q = q.Where(returnrequest.CompanyIDEQ(cid), returnrequest.CustomerIDEQ(custID))
	} else {
		cid, did, err := deptScope(id)
		if err != nil {
			return nil, err
		}
		q = q.Where(returnrequest.CompanyIDEQ(cid))
		if did != nil {
			q = q.Where(returnrequest.DepartmentIDEQ(*did))
		}
	}
	rr, err := q.Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	if id.Role == "customer" {
		// 子帳號 self 已由 where 限定;主帳號在上層已被拒(無 CustomerID)。
	} else if _, _, err := deptScope(id); err != nil {
		return nil, err
	}
	if rr.Status != "approved" {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"reason": "尚無證明可出示"})
	}
	items, err := db.ReturnRequestItem.Query().
		Where(returnrequestitem.ReturnRequestIDEQ(rr.ID)).All(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	cust, err := db.Customer.Query().Where(customer.ID(rr.CustomerID)).Only(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	resp := &salesorderv1.GetReturnCertificateResponse{
		Id: strconv.Itoa(rr.ID), CustomerCode: cust.CustomerCode, CustomerName: cust.Name,
		CreatedAt: rr.CreatedAt.Format(time.RFC3339), Status: rr.Status,
		ReviewedAt: formatTimePtr(rr.ReviewedAt),
	}
	if rr.ReviewedByUserID != nil {
		if u, err := db.User.Query().Where(user.ID(*rr.ReviewedByUserID)).Only(ctx); err == nil {
			resp.ReviewerName = u.Name
		}
	}
	for _, it := range items {
		v := &salesorderv1.ReturnRequestItemView{
			Id: strconv.Itoa(it.ID), SourceType: it.SourceType,
			ProductName: it.ProductName, Spec: it.Spec, Unit: it.Unit,
			Quantity: it.Quantity, Reason: it.Reason,
		}
		for _, fid := range it.PhotoFileIds {
			fidN, err := parseID(fid)
			if err != nil {
				continue
			}
			if fa, err := db.FileAsset.Query().Where(fileasset.IDEQ(fidN)).Only(ctx); err == nil {
				v.PhotoUrls = append(v.PhotoUrls, fa.URL)
			}
		}
		resp.Items = append(resp.Items, v)
	}
	return connect.NewResponse(resp), nil
}

// formatTimePtr 格式化可空時間(空回 "")。
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

var _ = authz.Identity{}

// parseVersion 解析樂觀鎖版本(0 起跳;parseID 拒 0 故另立)。
func parseVersion(s string) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v < 0 {
		return 0, errcode.SysInvalidArgument.Error(map[string]string{"field": "expected_version"})
	}
	return v, nil
}
