package store

import (
	"context"
	"slices"
	"sync"
)

// Fake 為記憶體實作:供判定層的單元測試與平台 CLI 使用,不進 production 路徑。
//
// 它必須與 SQL 實作**語意一致**,否則判定層的單元測試會失真:
//   - Subscription 只回未取消者,無則 (nil, nil);
//   - Overrides 不過濾到期 —— 到期與否是判定層的職責。至於「已撤銷」,在記憶體裡無從
//     表達:撤銷是資料庫的事(revoked_at),放進來的即視為有效。
type Fake struct {
	mu       sync.Mutex
	features map[string]Feature
	plans    map[string][]Entitlement
	subs     map[int]Subscription
	overs    map[int][]Override
}

var _ Store = (*Fake)(nil)

func NewFake() *Fake {
	return &Fake{
		features: map[string]Feature{},
		plans:    map[string][]Entitlement{},
		subs:     map[int]Subscription{},
		overs:    map[int][]Override{},
	}
}

func (f *Fake) PutFeature(x Feature) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.features[x.Code] = x
}

func (f *Fake) PutPlan(planCode string, ents []Entitlement) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.plans[planCode] = ents
}

func (f *Fake) PutSubscription(s Subscription) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subs[s.CompanyID] = s
}

func (f *Fake) PutOverride(o Override) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.overs[o.CompanyID] = append(f.overs[o.CompanyID], o)
}

// Features 回傳副本:呼叫端(或另一個測試)改動不得汙染後續讀取。
func (f *Fake) Features(context.Context) (map[string]Feature, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]Feature, len(f.features))
	for k, v := range f.features {
		out[k] = v
	}
	return out, nil
}

func (f *Fake) PlanEntitlements(_ context.Context, planCode string) ([]Entitlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.plans[planCode]), nil
}

func (f *Fake) Overrides(_ context.Context, companyID int) ([]Override, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.overs[companyID]), nil
}

func (f *Fake) Subscription(_ context.Context, companyID int) (*Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.subs[companyID]
	if !ok || s.Status == "cancelled" {
		return nil, nil
	}
	return &s, nil
}
