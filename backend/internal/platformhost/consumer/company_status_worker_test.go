package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/sw-ssd/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/platformhost/consumer"
)

// newWorker 建 worker 並注入 fake setStatus(預設 services.SetCompanyStatus,本檔不連真庫)。
func newWorker(src consumer.EventStore, tx consumer.SystemTx, setter *fakeStatusSetter) *consumer.CompanyStatusWorker {
	w := consumer.NewCompanyStatusWorker(src, tx)
	w.WithSetStatus(setter.setStatus)
	return w
}

// fakeStatusCall 記錄一次 setStatus 呼叫。
type fakeStatusCall struct {
	companyID int
	status    company.Status
	reason    string
	actor     authz.Identity
}

// fakeStatusSetter 實作 worker 的 setStatus 簽章。failFor>0 時該公司一律失敗。
type fakeStatusSetter struct {
	calls   []fakeStatusCall
	failFor int
}

func (f *fakeStatusSetter) setStatus(_ context.Context, _ *ent.Client, companyID int, status company.Status, reason string, actor authz.Identity) error {
	if companyID == f.failFor {
		return errors.New("模擬 SetCompanyStatus 失敗")
	}
	f.calls = append(f.calls, fakeStatusCall{companyID: companyID, status: status, reason: reason, actor: actor})
	return nil
}

// ① 一筆 company.status_changed outbox → 消費並呼叫 setStatus(凍結公司 42)。
func TestWorkerProcessOnceConsumesOutbox(t *testing.T) {
	src := &fakeEvents{events: []store.Event{
		{ID: 1, EventType: "company.status_changed", Payload: []byte(`{"company_id":42,"status":"suspended","reason":"逾期未付"}`)},
	}, actorID: 7}
	tx := &fakeSystemTx{events: src, claimed: map[int64]bool{}}
	setter := &fakeStatusSetter{}
	w := newWorker(src, tx, setter)

	n, err := w.ProcessOnce(context.Background(), 10)
	if err != nil || n != 1 {
		t.Fatalf("應消費 1 筆: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 || setter.calls[0].companyID != 42 || setter.calls[0].status != company.StatusSuspended {
		t.Fatalf("應凍結公司 42,got %+v", setter.calls)
	}
	if src.actorCalls != 1 {
		t.Fatalf("同一趟應解析 actor 一次,got %d", src.actorCalls)
	}
}

// ② 第二趟:事件已認領 → 無可消費 → 不重複呼叫。
func TestWorkerProcessOnceIsIdempotent(t *testing.T) {
	src := &fakeEvents{events: []store.Event{
		{ID: 1, EventType: "company.status_changed", Payload: []byte(`{"company_id":42,"status":"active","reason":"補款復原"}`)},
	}, actorID: 7}
	tx := &fakeSystemTx{events: src, claimed: map[int64]bool{}}
	setter := &fakeStatusSetter{}
	w := newWorker(src, tx, setter)

	first, err := w.ProcessOnce(context.Background(), 10)
	if err != nil || first != 1 {
		t.Fatalf("第一趟應消費 1 筆: n=%d err=%v", first, err)
	}
	second, err := w.ProcessOnce(context.Background(), 10)
	if err != nil || second != 0 {
		t.Fatalf("第二趟應無事可做: n=%d err=%v", second, err)
	}
	if len(setter.calls) != 1 {
		t.Fatalf("第二趟不得重複呼叫 setStatus,got %d", len(setter.calls))
	}
}

// ③ setStatus 失敗 → 整筆回滾(認領還原)、報錯、下趟重試成功。
func TestWorkerProcessOnceRollsBackOnFailure(t *testing.T) {
	src := &fakeEvents{events: []store.Event{
		{ID: 1, EventType: "company.status_changed", Payload: []byte(`{"company_id":42,"status":"suspended","reason":"逾期未付"}`)},
	}, actorID: 7}
	tx := &fakeSystemTx{events: src, claimed: map[int64]bool{}}
	setter := &fakeStatusSetter{failFor: 42}
	w := newWorker(src, tx, setter)

	if _, err := w.ProcessOnce(context.Background(), 10); err == nil {
		t.Fatal("setStatus 失敗必須報錯")
	}
	if tx.claimed[1] {
		t.Fatal("失敗事件不得留成已認領")
	}

	setter.failFor = 0 // 修好
	n, err := w.ProcessOnce(context.Background(), 10)
	if err != nil || n != 1 {
		t.Fatalf("修好後應重試成功: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 || setter.calls[0].companyID != 42 {
		t.Fatalf("重試應真的呼叫 setStatus,got %+v", setter.calls)
	}
}

// ④ 壞 payload → 報錯、不算認領(留待重試)。
func TestWorkerProcessOnceRejectsBadPayload(t *testing.T) {
	src := &fakeEvents{events: []store.Event{
		{ID: 1, EventType: "company.status_changed", Payload: []byte(`{"company_id":42}`)}, // 缺 status
	}, actorID: 7}
	tx := &fakeSystemTx{events: src, claimed: map[int64]bool{}}
	setter := &fakeStatusSetter{}
	w := newWorker(src, tx, setter)

	if _, err := w.ProcessOnce(context.Background(), 10); err == nil {
		t.Fatal("壞 payload 必須報錯")
	}
	if tx.claimed[1] {
		t.Fatal("壞 payload 不得認領")
	}
	if len(setter.calls) != 0 {
		t.Fatalf("壞 payload 不得呼叫 setStatus,got %+v", setter.calls)
	}
}

// ⑤ 非 company.status_changed 型別不應被 worker 讀取(UndispatchedCompanyEvents 只回 outbox)。
func TestWorkerIgnoresNonOutboxEvents(t *testing.T) {
	src := &fakeEvents{events: []store.Event{
		{ID: 1, EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)},
		{ID: 2, EventType: "company.status_changed", Payload: []byte(`{"company_id":43,"status":"suspended","reason":"x"}`)},
	}, actorID: 7}
	tx := &fakeSystemTx{events: src, claimed: map[int64]bool{}}
	setter := &fakeStatusSetter{}
	w := newWorker(src, tx, setter)

	n, err := w.ProcessOnce(context.Background(), 10)
	if err != nil || n != 1 {
		t.Fatalf("只應消費 1 筆 outbox: n=%d err=%v", n, err)
	}
	// 確認 subscription.suspended 不在 emitted(它由 consumer 主迴圈發 outbox,不在此佇列)。
	for _, e := range src.emitted {
		if e.EventType == "subscription.suspended" {
			t.Fatalf("worker 佇列不得含非 outbox 事件: %+v", e)
		}
	}
	if len(setter.calls) != 1 || setter.calls[0].companyID != 43 {
		t.Fatalf("應處理公司 43 的 outbox,got %+v", setter.calls)
	}
}

// ensureJSON 防止誤用:驗證測試 payload 是合法 JSON(開發期自我檢查)。
func ensureJSON(t *testing.T, b []byte) {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("測試 payload 必須合法 JSON: %v", err)
	}
}

var _ = ensureJSON
