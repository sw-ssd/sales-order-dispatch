//go:build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

// TestIntegrationBoardValkeyCrossReplica 跨 replica 送達(08 5.2.2):
// 發佈端以**獨立連線**經 Valkey 廣播(模擬另一 replica 的 mutation) →
// 本機訂閱迴圈收到事件灌入 hub;其他部門收不到(部門隔離)。
func TestIntegrationBoardValkeyCrossReplica(t *testing.T) {
	addr := testsupport.Valkey(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 訂閱側(本機 replica):先起訂閱迴圈(內部 PSubscribe.Receive 同步等就緒),
	// 再註冊 hub 訂閱者,避免發佈搶跑。
	subClient := cache.NewClient(addr)
	t.Cleanup(func() { _ = subClient.Close() })
	stop := startBoardSubscriber(ctx, subClient)
	t.Cleanup(stop)

	const deptID, otherDept = 4242, 4243
	ch, unsub := sharedHub.subscribe(deptID)
	t.Cleanup(unsub)
	otherCh, unsubOther := sharedHub.subscribe(otherDept)
	t.Cleanup(unsubOther)

	// 發佈端(另一 replica):獨立 client、直接走 valkeyPublisher(不經本機直投)。
	pubClient := cache.NewClient(addr)
	t.Cleanup(func() { _ = pubClient.Close() })
	(&valkeyPublisher{client: pubClient}).Publish(ctx, deptID, BoardEvent{
		Type: "dispatch", SalesOrderID: 7, DepartmentID: deptID,
	})

	select {
	case ev := <-ch:
		if ev.SalesOrderID != 7 || ev.DepartmentID != deptID || ev.Type != "dispatch" {
			t.Fatalf("跨 replica 事件不符: %+v", ev)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("3 秒內未收到跨 replica 看板事件")
	}

	// 部門隔離:同一批廣播不得送進其他部門的訂閱者。
	select {
	case ev := <-otherCh:
		t.Fatalf("看板事件不得跨部門送達: %+v", ev)
	case <-time.After(200 * time.Millisecond):
	}
}

// TestIntegrationBoardValkeyDegrade 發佈端 Valkey 不可用的降級邊界:
// 複合發佈的**本機直投仍必須即時送達**,遠端失敗只記日誌(D14:不影響 mutation)。
// 這正是 mountAuth 在 Valkey ping 失敗時不呼叫 EnableBoardFanout 的同一語意,
// 此處直接以壞 client 模擬「運行中 Valkey 掛掉」。
func TestIntegrationBoardValkeyDegrade(t *testing.T) {
	bad := cache.NewClient("127.0.0.1:1") // 必然連不上
	t.Cleanup(func() { _ = bad.Close() })
	SetBoardPublisher(combinedPublisher{
		local:  localPublisher{},
		remote: &valkeyPublisher{client: bad},
	})
	t.Cleanup(func() { SetBoardPublisher(localPublisher{}) })

	const deptID = 4244
	ch, unsub := sharedHub.subscribe(deptID)
	t.Cleanup(unsub)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	boardPublisher.Publish(ctx, deptID, BoardEvent{
		Type: "route_assign", SalesOrderID: 9, DepartmentID: deptID,
	})

	select {
	case ev := <-ch:
		if ev.SalesOrderID != 9 {
			t.Fatalf("本機直投事件不符: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Valkey 故障時本機直投仍必須送達(降級失效)")
	}
}
