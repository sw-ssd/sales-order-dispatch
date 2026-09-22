// Package services 的派車事件發佈(08 計畫 Task 5.2.2 的最小落地)。
// 發佈點統一規則:mutation 的 DB 交易提交成功後才發佈(交易內不發佈);rollback 無事件。
// 傳輸以 Valkey pub/sub(部門分 channel);Valkey 異常僅記日誌,不影響 mutation(D14)。
// WatchBoard 串流訂閱(5.2.1/5.2.3)另案;本檔只供 mutation 發佈。
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

// boardPublisher 為發佈器(預設空轉;server 組裝時注入 Valkey)。
var boardPublisher BoardPublisher = noopPublisher{}

// BoardPublisher 為發佈介面(測試注入 fake)。
type BoardPublisher interface {
	Publish(ctx context.Context, deptID int, ev BoardEvent)
}

// noopPublisher 無 Valkey 時的空轉(單測/未組裝)。
type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, int, BoardEvent) {}

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
