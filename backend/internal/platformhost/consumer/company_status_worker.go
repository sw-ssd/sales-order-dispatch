// Package consumer 見 consumer.go。本檔為 plan p3 的 outbox 反轉消費端:把 consumer 主迴圈
// 發出的 company.status_changed outbox 事件翻成 services.SetCompanyStatus 的呼叫。
//
// 為什麼需要這一層(而非 consumer 直接 SetCompanyStatus,舊設計):platform 獨立部署後,
// 凍結公司這一步必須跨網路/跨進程,不能用「認領 + 寫 companies」同一個交易鎖死。outbox 反轉讓
// 認領與發事件保持原子(同一平台交易),公司狀態變更由本 worker 最終一致地執行;兩者失敗皆可重試。
//
// 冪等:認領是條件式 UPDATE(一筆事件只會被認領一次),而 SetCompanyStatus 對同值為 no-op(不留稽核)
// —— 兩層加起來使重跑不產生第二個副作用。worker 與 consumer 主迴圈互斥讀取(主迴圈排除
// company.status_changed 型別),不會搶認領。
package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/platform/entitlements"
	"github.com/salesorder/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
)

// CompanyStatusWorker 消費 company.status_changed outbox → services.SetCompanyStatus。
type CompanyStatusWorker struct {
	events    EventStore
	sysTx     SystemTx
	cache     entitlements.Cache
	setStatus func(ctx context.Context, db *ent.Client, companyID int, status company.Status, reason string, actor authz.Identity) error
}

// NewCompanyStatusWorker 建立 worker:events 為平台讀取(同 consumer 的生產實作),
// sysTx 提供「認領 + 狀態變更」所在的系統範圍交易(同 consumer 的 DBSystemTx)。
func NewCompanyStatusWorker(events EventStore, sysTx SystemTx) *CompanyStatusWorker {
	return &CompanyStatusWorker{
		events:    events,
		sysTx:     sysTx,
		setStatus: services.SetCompanyStatus,
	}
}

// WithCache 接上權益快取(可選;nil＝不失效)。公司狀態變更後失效該租戶的權益快照。
func (w *CompanyStatusWorker) WithCache(cache entitlements.Cache) *CompanyStatusWorker {
	w.cache = cache
	return w
}

// WithSetStatus 替換狀態變更入口(預設 services.SetCompanyStatus;測試注入 fake 以脫離真庫)。
func (w *CompanyStatusWorker) WithSetStatus(fn func(ctx context.Context, db *ent.Client, companyID int, status company.Status, reason string, actor authz.Identity) error) *CompanyStatusWorker {
	w.setStatus = fn
	return w
}

// statusChangedPayload 為 company.status_changed outbox 的 payload 形狀。
type statusChangedPayload struct {
	CompanyID int    `json:"company_id"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
}

// ProcessOnce 依序處理未消費的 company.status_changed 事件,回傳本趟認領數。
//
// 單筆失敗不停整趟:記錯誤後續下一筆,最後 errors.Join 往外傳;失敗的那筆維持 dispatched_at
// IS NULL(下趟重試),不阻塞後續。語意與 consumer.DispatchOnce 一致。
func (w *CompanyStatusWorker) ProcessOnce(ctx context.Context, limit int) (int, error) {
	events, err := w.events.UndispatchedCompanyEvents(ctx, limit)
	if err != nil {
		return 0, wrapUncoded(err)
	}
	claimedN := 0
	var errs []error
	for _, ev := range events {
		claimed, err := w.consume(ctx, ev)
		if err != nil {
			errs = append(errs, fmt.Errorf("事件 %d(company.status_changed): %w", ev.ID, err))
			continue
		}
		if claimed {
			claimedN++
		}
	}
	return claimedN, errors.Join(errs...)
}

// consume 在交易內認領事件並呼叫 SetCompanyStatus。回傳本筆是否由這趟認領。
func (w *CompanyStatusWorker) consume(ctx context.Context, ev store.Event) (bool, error) {
	var p statusChangedPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return false, wrapUncoded(fmt.Errorf("事件 %d payload 不是合法 JSON: %w", ev.ID, err))
	}
	if p.CompanyID == 0 || p.Status == "" {
		return false, wrapUncoded(fmt.Errorf("事件 %d payload 缺 company_id 或 status", ev.ID))
	}
	actor, err := systemActor(ctx, w.events)
	if err != nil {
		return false, err
	}
	claimed := false
	err = w.sysTx.Run(ctx, func(ctx context.Context, tx Tx) error {
		ok, err := tx.Claim(ctx, ev.ID)
		claimed = ok
		if err != nil || !ok {
			return err
		}
		return w.setStatus(ctx, tx.Client(), p.CompanyID, company.Status(p.Status), p.Reason, actor)
	})
	if err != nil || !claimed {
		return claimed, err
	}
	if err := entitlements.Invalidate(ctx, w.cache, p.CompanyID); err != nil {
		log.Printf("platform status worker: 權益快取失效失敗(company=%d): %v（最長 TTL 內仍讀舊權益）", p.CompanyID, err)
	}
	return claimed, nil
}
