// Package services 的通知觸發(07 計畫 Task 4.4.3–4.4.4 + 06 掛點)。
// 觸發方在業務同一 DB 交易內建 notifications(status=pending);FCM 外部呼叫於
// 交易提交後經 AfterCommit 執行,失敗僅標 failed 不回滾(D16)。
// 路由:業務下單推客戶全部子帳號(主帳號排除);客戶自行下單不通知;
// 後台新增專屬商品推主責業務(無則退回 dept_admin);退貨審核推發起帳號。
package services

import (
	"context"
	"strconv"
	"strings"

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
	if err := dbtenant.AfterCommit(ctx, func(hookCtx context.Context) error {
		// 掛鉤在請求交易**提交後**執行,必須用 PostCommitClient 另開交易落發送結果 ——
		// 這裡傳入的 db 是請求交易 client,已提交,對它寫入會 ErrTxDone 而通知永遠停在 pending。
		triggerSender.Send(hookCtx, dbtenant.PostCommitClient(hookCtx, db), batch)
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

// OnDeliveryAssigned logistics 指派觸發(10.11):推「新任務」給該車次被指派的司機本人。
// 收件者是司機關聯的 users.id(單人),不做部門廣播 —— 任務是個人責任。
func OnDeliveryAssigned(ctx context.Context, db *ent.Client, cid int, did *int,
	driverUserID, deliveryID int, routeName, routeCode string) error {
	if driverUserID <= 0 {
		return nil
	}
	title, content, _ := RenderTemplate(ctx, db, cid, did, "logistics_assigned", "in_app", "zh-Hant",
		map[string]string{"route_name": routeName, "route_code": routeCode})
	if title == "" {
		title = "新配送任務"
		content = routeName + "（" + routeCode + "）"
	}
	payload := map[string]any{"delivery_id": deliveryID}
	for _, ch := range []string{"in_app", "fcm"} {
		if err := queueNotifications(ctx, db, cid, did, []int{driverUserID}, ch, title, content, payload, nil); err != nil {
			return err
		}
	}
	return nil
}

// OnDeliveryCompleted logistics 送達觸發(10.11):逐筆已送達訂單推**店家**(該客戶全部
// 子帳號;主帳號排除,同 OnOrderCreated)與**主責業務**(無則同部門 dept_admin)。
//
// 一筆車次可能載多個客戶的貨,故逐客戶成一組通知;同一客戶在一車次出現多筆訂單時
// 只推一次(避免洗版),payload 帶該客戶的 order_ids。
func OnDeliveryCompleted(ctx context.Context, db *ent.Client, cid int, did *int,
	routeID int, orders []*ent.SalesOrder) error {
	if len(orders) == 0 {
		return nil
	}
	// 依客戶分組(保序,讓輸出可預期)。
	byCustomer := map[int][]*ent.SalesOrder{}
	var order2cust []int
	for _, o := range orders {
		if _, seen := byCustomer[o.CustomerID]; !seen {
			order2cust = append(order2cust, o.CustomerID)
		}
		byCustomer[o.CustomerID] = append(byCustomer[o.CustomerID], o)
	}
	for _, custID := range order2cust {
		group := byCustomer[custID]
		// 1. 店家端(客戶子帳號;無子帳號則略過該客戶)。
		subs, err := subAccountIDs(ctx, db, custID)
		if err != nil {
			return err
		}
		if len(subs) > 0 {
			orderIDs := make([]int, 0, len(group))
			nos := make([]string, 0, len(group))
			for _, o := range group {
				orderIDs = append(orderIDs, o.ID)
				nos = append(nos, o.OrderNo)
			}
			title, content, _ := RenderTemplate(ctx, db, cid, did, "logistics_delivered", "in_app", "zh-Hant",
				map[string]string{"order_nos": strings.Join(nos, "、"), "count": strconv.Itoa(len(group))})
			if title == "" {
				title = "訂單已送達"
				content = strings.Join(nos, "、")
			}
			payload := map[string]any{"order_ids": orderIDs, "route_id": routeID}
			for _, ch := range []string{"in_app", "fcm"} {
				if err := queueNotifications(ctx, db, cid, did, subs, ch, title, content, payload, nil); err != nil {
					return err
				}
			}
		}
		// 2. 主責業務(無則同部門 dept_admin)。
		rep, err := primaryRepOrAdmins(ctx, db, cid, did, custID)
		if err != nil {
			return err
		}
		if len(rep) == 0 {
			continue
		}
		custName := ""
		if len(group) > 0 {
			if c, err := db.Customer.Query().Where(customer.IDEQ(custID)).Only(ctx); err == nil {
				custName = c.Name
			}
		}
		title, content, _ := RenderTemplate(ctx, db, cid, did, "logistics_delivered_rep", "in_app", "zh-Hant",
			map[string]string{"customer_name": custName, "count": strconv.Itoa(len(group))})
		if title == "" {
			title = "配送完成"
			content = custName + " 共 " + strconv.Itoa(len(group)) + " 筆已送達"
		}
		payload := map[string]any{"customer_id": custID, "route_id": routeID}
		for _, ch := range []string{"in_app", "fcm"} {
			if err := queueNotifications(ctx, db, cid, did, rep, ch, title, content, payload, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
