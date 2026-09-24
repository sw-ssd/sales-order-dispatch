// Package services 的派車事件發佈(08 計畫 Task 5.2.2 的最小落地)。
// 發佈點統一規則:mutation 的 DB 交易提交成功後才發佈(交易內不發佈);rollback 無事件。
// 傳輸以 Valkey pub/sub(部門分 channel);Valkey 異常僅記日誌,不影響 mutation(D14)。
// 本檔同時承載跨 replica 訂閱迴圈(startBoardSubscriber,5.2.2);WatchBoard 串流本體在 dispatch_watch.go。
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// BoardEvent 為看板事件(僅失效提示;前端收到後全量重查)。
type BoardEvent struct {
	Type             string `json:"type"` // route_assign/dispatch/dispatch_cancel
	SalesOrderID     int    `json:"sales_order_id"`
	RouteID          *int   `json:"route_id"`
	DeliverySequence *int   `json:"delivery_sequence"`
	Version          int    `json:"version"`
	DepartmentID     int    `json:"department_id"`
	ExpectedDate     string `json:"expected_date,omitempty"`
}

// boardPublisher 為發佈器(預設程序內直投;server 組裝時可換 Valkey 跨 replica)。
var boardPublisher BoardPublisher = localPublisher{}

// BoardPublisher 為發佈介面(測試注入 fake)。
type BoardPublisher interface {
	Publish(ctx context.Context, deptID int, ev BoardEvent)
}

// valkeyPublisher 以 Valkey pub/sub 發佈(部門分 channel)。
type valkeyPublisher struct {
	client *redis.Client
}

// NewValkeyPublisher 建立 Valkey 發佈器。
func NewValkeyPublisher(c *redis.Client) BoardPublisher {
	return &valkeyPublisher{client: c}
}

// boardChannel 部門分 channel。
func boardChannel(deptID int) string {
	return fmt.Sprintf("board:dept:%d", deptID)
}

// Publish 發佈(異常僅記日誌,不回錯)。
func (p *valkeyPublisher) Publish(ctx context.Context, deptID int, ev BoardEvent) {
	b, err := json.Marshal(ev)
	if err != nil {
		log.Printf("dispatch: 看板事件序列化失敗: %v", err)
		return
	}
	if err := p.client.Publish(ctx, boardChannel(deptID), b).Err(); err != nil {
		log.Printf("dispatch: 看板事件發佈失敗(降級,mutation 不受影響): %v", err)
	}
}

// localPublisher 將發佈直投程序內訂閱(同 replica 即時;跨 replica 走 combinedPublisher)。
type localPublisher struct{}

func (localPublisher) Publish(_ context.Context, deptID int, ev BoardEvent) {
	sharedHub.fanout(deptID, ev)
}

// boardChannelPattern 訂閱用的部門 channel 模式。
const boardChannelPattern = "board:dept:*"

// combinedPublisher 同時走「本機直投 + 遠端(Valkey)」:本機路徑保證同 replica
// 即時且 Valkey 異常不影響;遠端路徑跨 replica(best-effort,失敗僅記日誌)。
//
// 同一事件會因「直投 + 訂閱回聲」對本機連線送達兩次——看板事件語意是失效提示
// (前端全量重查),重複只多一次重查,刻意不去重(不值得為它加 origin 欄位)。
type combinedPublisher struct {
	local  BoardPublisher
	remote BoardPublisher
}

func (p combinedPublisher) Publish(ctx context.Context, deptID int, ev BoardEvent) {
	p.local.Publish(ctx, deptID, ev)
	p.remote.Publish(ctx, deptID, ev)
}

// EnableBoardFanout 啟用看板跨 replica 層(08 5.2.2):複合發佈 + 訂閱迴圈。
// 由 server 組裝期呼叫一次(Valkey 已 ping 過);Valkey 不可用時組裝期不呼叫,
// 維持預設 localPublisher(降級:只看本 replica 的變更,heartbeat 不受影響)。
func EnableBoardFanout(ctx context.Context, client *redis.Client) {
	boardPublisher = combinedPublisher{
		local:  localPublisher{},
		remote: &valkeyPublisher{client: client},
	}
	_ = startBoardSubscriber(ctx, client) // stop 捨棄:與程序同壽命(測試直接呼叫 startBoardSubscriber)
}

// startBoardSubscriber 起 Valkey 訂閱迴圈:遠端看板事件 → 本機 hub。
//
// PSubscribe.Receive 同步確認訂閱建立;失敗 → log 並回 no-op stop(發佈端的
// 本機直投仍在,降級語意同上)。斷線由 go-redis 自動重連;中斷期間的跨 replica
// 事件不補發——事件本就不保證送達,狀態正確性由前端重查兜底,heartbeat(25s)
// 持續到達可當輪詢節拍。
func startBoardSubscriber(ctx context.Context, client *redis.Client) (stop func()) {
	pubsub := client.PSubscribe(ctx, boardChannelPattern)
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Printf("dispatch: 看板 Valkey 訂閱失敗(降級:僅本機直投): %v", err)
		_ = pubsub.Close()
		return func() {}
	}
	go func() {
		for msg := range pubsub.Channel() {
			var ev BoardEvent
			if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
				log.Printf("dispatch: 看板事件反序列化失敗(略過): %v", err)
				continue
			}
			sharedHub.fanout(ev.DepartmentID, ev)
		}
	}()
	return func() { _ = pubsub.Close() }
}

// SetBoardPublisher 設定發佈器(僅測試/server 組裝用)。
func SetBoardPublisher(p BoardPublisher) { boardPublisher = p }

// publishAfterCommit 註冊提交後發佈(交易內只收集,提交後才送)。
func publishAfterCommit(ctx context.Context, deptID int, ev BoardEvent) error {
	e := ev
	d := deptID
	return dbtenant.AfterCommit(ctx, func(ctx context.Context) error {
		boardPublisher.Publish(ctx, d, e)
		return nil
	})
}
