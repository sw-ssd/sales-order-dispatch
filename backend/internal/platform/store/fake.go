package store

import (
	"context"
	"sync"
)

// Fake 為記憶體實作:供判定層的單元測試與平台 CLI 使用,不進 production 路徑。
//
// 它必須與 SQL 實作**語意一致**,否則判定層的單元測試會失真:
//   - Subscription 只回未取消者,無則 (nil, nil);
//   - Overrides 不過濾到期 —— 到期與否是判定層的職責。至於「已撤銷」,在記憶體裡無從
//     表達:撤銷是資料庫的事(revoked_at),放進來的即視為有效;
//   - Put 與 getter 兩端都不別名(alias)呼叫端的指標:PG 每次掃描都配置新的指標,
//     呼叫端改動不到 store,共用一份假實作的測試案例之間也不得互相汙染。
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
	f.plans[planCode] = cloneEntitlements(ents)
}

func (f *Fake) PutSubscription(s Subscription) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subs[s.CompanyID] = cloneSubscription(s)
}

func (f *Fake) PutOverride(o Override) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.overs[o.CompanyID] = append(f.overs[o.CompanyID], cloneOverride(o))
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
	return cloneEntitlements(f.plans[planCode]), nil
}

func (f *Fake) Overrides(_ context.Context, companyID int) ([]Override, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return cloneOverrides(f.overs[companyID]), nil
}

// Subscription 回傳該公司的訂閱（**含已取消者**）；沒有列才回 (nil, nil)。
//
// 假實作與 SQL 實作的過濾語意必須一致，否則判定層的測試會失真：**cancelled 由判定層判定**
// （allow-list → PLAT-3001），store 不得預先濾掉 —— 濾掉會讓「只有一筆已取消合約」與「完全沒有
// 合約」不可區分，而後者的語意是「尚未開通計費 → 不施加限制」（spec §4.5），等於取消即送免費
// 方案（F-8）。與 override 的「到期由判定層判斷」同一個原則。
func (f *Fake) Subscription(_ context.Context, companyID int) (*Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.subs[companyID]
	if !ok {
		return nil, nil
	}
	s = cloneSubscription(s)
	return &s, nil
}

// 以下為深拷貝:PG 實作的每一列都是掃描出來的新配置,呼叫端與 store 之間沒有共享記憶體;
// 假實作若只複製切片,元素內的指標仍與 store 共用,`*ov[0].Limit = 1` 就會靜默改到之後
// 所有讀取。
func clonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}

func cloneEntitlements(in []Entitlement) []Entitlement {
	out := make([]Entitlement, len(in))
	for i, e := range in {
		e.Limit = clonePtr(e.Limit)
		out[i] = e
	}
	return out
}

func cloneOverride(o Override) Override {
	o.Enabled = clonePtr(o.Enabled)
	o.Limit = clonePtr(o.Limit)
	o.ExpiresAt = clonePtr(o.ExpiresAt)
	return o
}

func cloneOverrides(in []Override) []Override {
	out := make([]Override, len(in))
	for i, o := range in {
		out[i] = cloneOverride(o)
	}
	return out
}

func cloneSubscription(s Subscription) Subscription {
	s.TrialEnds = clonePtr(s.TrialEnds)
	s.GraceUntil = clonePtr(s.GraceUntil)
	return s
}
