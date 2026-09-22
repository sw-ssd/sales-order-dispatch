// Package services 的通知觸發(07 計畫 Task 4.4.3–4.4.4 + 06 掛點)。
// 觸發方在業務同一 DB 交易內建 notifications(status=pending);FCM 外部呼叫於
// 交易提交後經 AfterCommit 執行,失敗僅標 failed 不回滾(D16)。
// 路由:業務下單推客戶全部子帳號(主帳號排除);客戶自行下單不通知;
// 後台新增專屬商品推主責業務(無則退回 dept_admin);退貨審核推發起帳號。
package services

import (
	"context"
	"strconv"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// triggerSender 為觸發用的 Sender(預設 Fake 全成功;測試可換)。
var triggerSender Sender = &FakeSender{}

// SetTriggerSender 設定觸發 Sender(僅測試用)。
func SetTriggerSender(s Sender) { triggerSender = s }

// queueNotifications 建 pending 通知 + 註冊提交後發送(同一交易)。
func queueNotifications(ctx context.Context, db *ent.Client, cid int, did *int, userIDs []int, channel, title, content string, payload map[string]any, templateID *int) error {
	if len(userIDs) == 0 {
		return nil
	}
	var ids []int
	for _, uid := range userIDs {
		b := db.Notification.Create().SetCompanyID(cid).SetUserID(uid).
			SetChannel(channel).SetTitle(title).SetContent(content).SetStatus("pending")
		if did != nil {
			b = b.SetDepartmentID(*did)
		}
		if payload != nil {
			b = b.SetPayload(payload)
		}
		if templateID != nil {
			b = b.SetTemplateID(*templateID)
		}
		n, err := b.Save(ctx)
		if err != nil {
			return err
		}
		ids = append(ids, n.ID)
	}
	batch := append([]int(nil), ids...)
	if err := dbtenant.AfterCommit(ctx, func(ctx context.Context) error {
		triggerSender.Send(ctx, db, batch)
		return nil
	}); err != nil {
		// 無收集器（直呼服務的單測語境）→ 同步直發；生產 HTTP 必經 Interceptor，
		// 故此分支只在測試出現，不影響提交後發送語意。
		triggerSender.Send(ctx, db, batch)
	}
	return nil
}

// subAccountIDs 取客戶全部子帳號(is_primary=false + 啟用中)。
func subAccountIDs(ctx context.Context, db *ent.Client, custID int) ([]int, error) {
	users, err := db.User.Query().
		Where(user.CustomerIDEQ(custID), user.IsPrimaryEQ(false),
			user.StatusEQ(user.StatusActive)).All(ctx)
	if err != nil {
		return nil, err
	}
	var out []int
	for _, u := range users {
		out = append(out, u.ID)
	}
	return out, nil
}

// OnOrderCreated 下單觸發(業務代客下單才推;客戶自行下單直接結束)。
func OnOrderCreated(ctx context.Context, db *ent.Client, cid int, did *int, custID int, isCustomerOrder bool, orderID int, orderNo string, itemCount int) error {
	if isCustomerOrder {
		return nil
	}
	subs, err := subAccountIDs(ctx, db, custID)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}
	title, content, _ := RenderTemplate(ctx, db, cid, did, "order_created", "in_app", "zh-Hant",
		map[string]string{"order_no": orderNo, "item_count": strconv.Itoa(itemCount)})
	if title == "" {
		title = "新訂單 " + orderNo
		content = "業務已為您下單"
	}
	payload := map[string]any{"order_id": orderID}
	for _, ch := range []string{"in_app", "fcm"} {
		if err := queueNotifications(ctx, db, cid, did, subs, ch, title, content, payload, nil); err != nil {
			return err
		}
	}
	return nil
}

// OnCustomerProductCreated 專屬商品觸發(主責業務;無則退回 dept_admin)。
func OnCustomerProductCreated(ctx context.Context, db *ent.Client, cid int, did *int, custID, productID int, customerName, productName string) error {
	rep, err := primaryRepOrAdmins(ctx, db, cid, did, custID)
	if err != nil {
		return err
	}
	if len(rep) == 0 {
		return nil
	}
	title, content, _ := RenderTemplate(ctx, db, cid, did, "customer_product_created", "in_app", "zh-Hant",
		map[string]string{"customer_name": customerName, "product_name": productName})
	if title == "" {
		title = "專屬商品已新增"
		content = customerName + "：" + productName
	}
	payload := map[string]any{"customer_id": custID, "product_id": productID}
	for _, ch := range []string{"in_app", "fcm"} {
		if err := queueNotifications(ctx, db, cid, did, rep, ch, title, content, payload, nil); err != nil {
			return err
		}
	}
	return nil
}

// OnDispatchConfirmed 派車通知(每筆已派訂單推該客戶子帳號;fire-and-record)。
func OnDispatchConfirmed(ctx context.Context, db *ent.Client, cid int, did *int, custID, orderID int, orderNo, routeName, date string) error {
	subs, err := subAccountIDs(ctx, db, custID)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}
	title, content, _ := RenderTemplate(ctx, db, cid, did, "dispatch", "in_app", "zh-Hant",
		map[string]string{"order_no": orderNo, "route_name": routeName, "date": date})
	if title == "" {
		title = "訂單已派車 " + orderNo
		content = routeName + " " + date
	}
	payload := map[string]any{"order_id": orderID}
	for _, ch := range []string{"in_app", "fcm"} {
		if err := queueNotifications(ctx, db, cid, did, subs, ch, title, content, payload, nil); err != nil {
			return err
		}
	}
	return nil
}

// OnReturnReviewed 退貨審核觸發(僅推發起帳號)。
func OnReturnReviewed(ctx context.Context, db *ent.Client, cid int, did *int, creatorID int, decision, reason string, requestID int) error {
	title, content, _ := RenderTemplate(ctx, db, cid, did, "return_reviewed", "in_app", "zh-Hant",
		map[string]string{"decision": decision, "reject_reason": reason})
	if title == "" {
		title = "退貨審核" + decision
		if content = reason; content == "" {
			content = "您的退貨申請已" + decision
		}
	}
	payload := map[string]any{"return_request_id": requestID, "decision": decision}
	for _, ch := range []string{"in_app", "fcm"} {
		if err := queueNotifications(ctx, db, cid, did, []int{creatorID}, ch, title, content, payload, nil); err != nil {
			return err
		}
	}
	return nil
}

// primaryRepOrAdmins 解析接收者:主責業務(啟用中)否則同部門 dept_admin。
func primaryRepOrAdmins(ctx context.Context, db *ent.Client, cid int, did *int, custID int) ([]int, error) {
	cust, err := db.Customer.Query().Where(customer.IDEQ(custID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	if cust.DefaultSalesRepID != nil {
		if u, err := db.User.Query().Where(user.IDEQ(*cust.DefaultSalesRepID)).Only(ctx); err == nil &&
			u.Status == user.StatusActive {
			return []int{u.ID}, nil
		}
	}
	q := db.User.Query().Where(user.HasCompanyWith(company.IDEQ(cid)), user.RoleEQ("dept_admin"),
		user.StatusEQ(user.StatusActive))
	if did != nil {
		q = q.Where(user.HasDepartmentWith(department.IDEQ(*did)))
	}
	admins, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	var out []int
	for _, a := range admins {
		out = append(out, a.ID)
	}
	return out, nil
}
