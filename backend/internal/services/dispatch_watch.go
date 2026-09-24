// Package services 的看板訂閱(08 計畫 Task 5.2.1/5.2.2/5.2.3)。
// WatchBoard 為 server streaming:部門級訂閱(department_id → 連線集合),事件僅轉發不過濾日期;
// heartbeat 25 秒保活(純連線維持,不寫事件、不進 Valkey);斷線清理註冊。
// 跨 replica(5.2.2)已落地:發佈走 combinedPublisher(本機直投 + Valkey 廣播),
// 訂閱迴圈(startBoardSubscriber)把遠端事件灌回本機 hub —— Valkey 異常時維持
// 程序內直投(降級:事件不保證補發,heartbeat 持續到達可當查詢節拍)。
package services

import (
	"context"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	salesorderv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// heartbeatInterval 為保活間隔(25 秒,低於常見 ingress 30 秒閾值)。
const heartbeatInterval = 25 * time.Second

// boardHub 為程序內訂閱註冊表(department_id → 連線集合)。
type boardHub struct {
	mu   sync.Mutex
	subs map[int]map[chan BoardEvent]struct{}
}

// sharedHub 為程序內單例。
var sharedHub = &boardHub{subs: map[int]map[chan BoardEvent]struct{}{}}

// subscribe 註冊部門訂閱(回取消函式)。
func (h *boardHub) subscribe(deptID int) (<-chan BoardEvent, func()) {
	ch := make(chan BoardEvent, 16)
	h.mu.Lock()
	m, ok := h.subs[deptID]
	if !ok {
		m = map[chan BoardEvent]struct{}{}
		h.subs[deptID] = m
	}
	m[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if m, ok := h.subs[deptID]; ok {
			delete(m, ch)
			if len(m) == 0 {
				delete(h.subs, deptID)
			}
		}
	}
}

// subscriberCount 回部門訂閱數(測試握手用)。
//
//lint:ignore U1000 測試經 integration tag 呼叫，非 integration 視角視為未用。
func (h *boardHub) subscriberCount(deptID int) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs[deptID])
}

// fanout 轉發部門事件(非阻塞;滿了丟棄該連線本則,at-most-once)。
func (h *boardHub) fanout(deptID int, ev BoardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs[deptID] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// WatchBoard 看板訂閱(server streaming;部門隔離;heartbeat 保活)。
func (s *DispatchService) WatchBoard(ctx context.Context, req *connect.Request[salesorderv1.WatchBoardRequest], stream *connect.ServerStream[salesorderv1.BoardEvent]) error {
	id, err := requireAuth(ctx)
	if err != nil {
		return err
	}
	_, did, err := deptScope(id)
	if err != nil {
		return err
	}
	// 看板為部門級:無部門上下文 → failed_precondition。
	if did == nil {
		// company_admin/super 無部門 → 以請求交易看其部門?不:看板必須定位部門,無部門即拒。
		return errcode.SysInvalidArgument.Error(map[string]string{"reason": "看板需選擇部門"})
	}
	if _, err := parseBoardDate(req.Msg.GetExpectedDeliveryDate()); err != nil {
		return err
	}
	// 串流不開請求交易(dbtenant.Interceptor 僅 unary):本 handler 只訂閱程序內 hub,
	// 不讀 DB,故無需交易;RLS 由連線既有 scope 承載,事件依部門隔離。
	ch, cancel := sharedHub.subscribe(*did)
	defer cancel()
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-ch:
			if err := stream.Send(boardEventToProto(ev)); err != nil {
				return nil // 客戶端已離線 → 清理並靜默結束
			}
		case <-ticker.C:
			if err := stream.Send(&salesorderv1.BoardEvent{Type: "heartbeat",
				DepartmentId: strconv.Itoa(*did)}); err != nil {
				return nil
			}
		}
	}
}

// boardEventToProto 轉 proto。
func boardEventToProto(ev BoardEvent) *salesorderv1.BoardEvent {
	p := &salesorderv1.BoardEvent{Type: ev.Type,
		DepartmentId: strconv.Itoa(ev.DepartmentID)}
	if ev.RouteID != nil {
		p.RouteId = strconv.Itoa(*ev.RouteID)
	}
	if ev.DeliverySequence != nil {
		p.DeliverySequence = strconv.Itoa(*ev.DeliverySequence)
	}
	if ev.SalesOrderID != 0 {
		p.SalesOrderId = strconv.Itoa(ev.SalesOrderID)
	}
	if ev.Version != 0 {
		p.Version = strconv.Itoa(ev.Version)
	}
	return p
}
