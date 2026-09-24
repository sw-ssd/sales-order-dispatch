package services

import (
	"context"
	"slices"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// notifByUser 取某使用者在本公司未軟刪除的通知(直連 fixture db)。
func notifsOf(f *logisticsFixture, uid int) []string {
	var out []string
	for _, n := range f.db.Notification.Query().AllX(context.Background()) {
		if n.UserID == uid {
			out = append(out, n.Channel+"/"+n.Status)
		}
	}
	slices.Sort(out)
	return out
}

// TestDeliveryAssignedNotifiesDriver 10.11:指派完成推「新任務」給該車次的司機本人
// (in_app + fcm 各一筆;狀態為提交後發送的結果 —— sent)。
func TestDeliveryAssignedNotifiesDriver(t *testing.T) {
	SetTriggerSender(&FakeSender{})
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	f, driverUID, _, _ := setupAssignedDelivery(t)
	got := notifsOf(f, driverUID)
	want := []string{"fcm/sent", "in_app/sent"}
	if len(got) != len(want) {
		t.Fatalf("司機應收到 %d 筆通知(in_app+fcm),got %v", len(want), got)
	}
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("通知通道/狀態不符: got %v 缺少 %s", got, w)
		}
	}
	// 非司機的使用者不該收到(任務是個人責任,不做部門廣播)。
	if n := notifsOf(f, f.managerID); len(n) != 0 {
		t.Fatalf("管理者不應收到指派通知,got %v", n)
	}
}

// TestDeliveryCompletedNotifiesCustomerAndRep 10.11:送達後推店家(客戶子帳號)與
// 主責業務;同一客戶多筆訂單只推一次(不洗版);非該車次客戶不推。
func TestDeliveryCompletedNotifiesCustomerAndRep(t *testing.T) {
	SetTriggerSender(&FakeSender{})
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	ctx := context.Background()
	f, driverUID, _, _ := setupAssignedDelivery(t)
	rpc := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))

	custA, subA := mkCustomerWithSub(t, f, "N1")
	custB, subB := mkCustomerWithSub(t, f, "N2")
	// 主責業務:customer A 指向 manager(收件者應為該人)。
	f.db.Customer.UpdateOneID(custA).SetDefaultSalesRepID(f.managerID).ExecX(ctx)

	// 車次上:客戶 A 兩筆(應只推一組)、客戶 B 一筆。
	mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custA)
	mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custA)
	mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custB)
	// 非本車次:不應推。
	otherRoute := f.db.Route.Create().SetCode("R7").SetName("別線").
		SetCompanyID(f.coID).SetDepartmentID(f.deptID).SaveX(ctx).ID
	mkOrderFor(t, f, OrderStatusProcessing, &otherRoute, custB)

	if err := startAndComplete(t, rpc); err != nil {
		t.Fatalf("完成配送: %v", err)
	}

	// 店家端:每個客戶 in_app + fcm 各一筆(合併成一組)。
	for name, sub := range map[string]int{"N1": subA, "N2": subB} {
		got := notifsOf(f, sub)
		if len(got) != 2 {
			t.Fatalf("客戶 %s 子帳號應收 2 筆(in_app+fcm,合併),got %v", name, got)
		}
	}
	// 主責業務:客戶 A 的主責是 manager → 2 筆;客戶 B 無主責 → 退回同部門 dept_admin,
	// 而 manager 正是該部門唯一 dept_admin → 再 2 筆。合計 4 筆。
	if got := notifsOf(f, f.managerID); len(got) != 4 {
		t.Fatalf("主責業務應收 4 筆(A 直推 2 + B 退回 admin 2),got %v", got)
	}
	// 司機本人的指派通知仍在(2 筆),送達不推司機。
	if got := notifsOf(f, driverUID); len(got) != 2 {
		t.Fatalf("司機應只有指派通知 2 筆,got %v", got)
	}
}

// TestDeliveryCompletedNotifyMergesPayload 同一客戶多筆訂單合併成單一通知,
// payload 帶全部 order_ids(前端點一則即可看到整批)。
func TestDeliveryCompletedNotifyMergesPayload(t *testing.T) {
	SetTriggerSender(&FakeSender{})
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	ctx := context.Background()
	f, driverUID, _, _ := setupAssignedDelivery(t)
	rpc := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))
	custA, subA := mkCustomerWithSub(t, f, "M1")
	o1 := mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custA)
	o2 := mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custA)

	if err := startAndComplete(t, rpc); err != nil {
		t.Fatalf("完成配送: %v", err)
	}
	var inApp int
	for _, n := range f.db.Notification.Query().AllX(ctx) {
		if n.UserID != subA || n.Channel != "in_app" {
			continue
		}
		inApp++
		ids, _ := n.Payload["order_ids"].([]any)
		if len(ids) != 2 {
			t.Fatalf("payload.order_ids 應含 2 筆,got %v", n.Payload)
		}
		got := map[int]bool{}
		for _, v := range ids {
			if f, ok := v.(float64); ok {
				got[int(f)] = true
			}
		}
		if !got[o1] || !got[o2] {
			t.Fatalf("payload.order_ids 應含 %d 與 %d,got %v", o1, o2, n.Payload)
		}
	}
	if inApp != 1 {
		t.Fatalf("同一客戶應合併為 1 筆 in_app,got %d", inApp)
	}
}

// TestLogisticsNotifySendFailureDoesNotBlockDelivery 10.11 錯誤處理:
// 提交後的 FCM 發送失敗**不影響**配送完成 —— 通知只標 failed,不重試、不回滾(D16)。
// 這是「通知是旁支、不擋主流程」的守門:若有人把發送拉進交易,這條會紅。
func TestLogisticsNotifySendFailureDoesNotBlockDelivery(t *testing.T) {
	ctx := context.Background()
	f, driverUID, _, _ := setupAssignedDelivery(t)
	failing := &FakeSender{OnSend: func(*ent.Notification) SendResult {
		return SendResult{Sent: false, FailReason: "fcm_unavailable"}
	}}
	SetTriggerSender(failing)
	t.Cleanup(func() { SetTriggerSender(&FakeSender{}) })

	rpc := f.newLogisticsServer(t, logisticsIdentity("staff", driverUID, f.coID, f.deptID))
	custA, subA := mkCustomerWithSub(t, f, "F1")
	mkOrderFor(t, f, OrderStatusProcessing, &f.routeID, custA)

	if err := startAndComplete(t, rpc); err != nil {
		t.Fatalf("發送失敗不應讓配送失敗: %v", err)
	}
	// 配送照樣完成、訂單照樣送達。
	d, err := f.db.LogisticsDelivery.Query().Only(ctx)
	if err != nil {
		t.Fatalf("讀配送: %v", err)
	}
	if d.Status != DeliveryStatusCompleted {
		t.Fatalf("配送應 completed,got %s", d.Status)
	}
	// 通知存在但標 failed(不重試、不刪)。
	var failed int
	for _, n := range f.db.Notification.Query().AllX(ctx) {
		if n.UserID == subA && n.Status == "failed" {
			failed++
		}
	}
	if failed == 0 {
		t.Fatalf("發送失敗的通知應標 failed,使用者 %d 的通知: %v", subA, notifsOf(f, subA))
	}
}
