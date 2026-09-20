// Package entitlements 判定「這個租戶現在可以做什麼」：方案預設 ⊕ 未過期 override，
// 並在訂閱不可用時全面 fail-closed。與 OpenFGA 的角色授權是**不同軸**：
// 這裡管「買了沒有」，那裡管「誰能做」。
package entitlements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// Feature code：方案／override／計數器三端共用同一組字串，打錯字就是算錯帳。
const (
	LimitSeats       = "limit.seats"
	LimitCustomers   = "limit.customers"
	LimitProducts    = "limit.products"
	LimitDepartments = "limit.departments"
	LimitStorageGB   = "limit.storage_gb"
	FeaturePrinting  = "feature.printing"
	FeatureDispatch  = "feature.dispatch"
	FeatureReturns   = "feature.returns"
)

// 訂閱狀態（platform.subscriptions.status）。只有 suspended／cancelled 是「合約不可用」：
// past_due 仍在寬限內（催收由 dunning job 改狀態，判定層不自行推算 grace_until）。
const (
	statusNone      = "none" // 沒有未取消的訂閱（Store 回 nil）
	statusTrialing  = "trialing"
	statusSuspended = "suspended"
	statusCancelled = "cancelled"
)

// Counter 由業務域提供（於 server.InitDomains 注入）：判定層不認得業務 schema，
// 只認 feature code → 目前用量的對照。回傳錯誤即視為系統錯誤（不得當成 0 放行）。
type Counter interface {
	Count(ctx context.Context, companyID int, feature string) (int, error)
}

// tenantState 為單一租戶的權益來源快照（快取的單位）。
type tenantState struct {
	CompanyID int    `json:"company_id"`
	PlanCode  string `json:"plan_code"`
	// PlanName 目前**沒有來源**：store.Subscription 不帶方案名（T3 的 SELECT 只取 p.code）。
	// 欄位先留著，讓租戶端投影的契約（proto 的 plan_name）不用等它。
	// ponytail: 天花板＝租戶端顯示的方案名為空；升級路徑＝store 的 Subscription 加 PlanName
	// （SELECT p.name）後，這裡改成 out.PlanName = sub.PlanName，其餘不動。
	PlanName     string                       `json:"plan_name"`
	Status       string                       `json:"status"`
	TrialEnds    *time.Time                   `json:"trial_ends_at,omitempty"`
	Features     map[string]store.Feature     `json:"features"`
	Entitlements map[string]store.Entitlement `json:"entitlements"`
	Overrides    []store.Override             `json:"overrides"`
}

// Service 為權益判定入口。
type Service struct {
	st        store.Store
	counters  Counter
	cache     Cache
	ttl       time.Duration
	now       func() time.Time
	unlimited bool
}

// New 建立判定服務；ttl <= 0 表示不快取。
func New(st store.Store, counters Counter, c Cache, ttl time.Duration) *Service {
	return &Service{st: st, counters: counters, cache: c, ttl: ttl, now: time.Now}
}

// Unlimited 回傳「全部允許、不限額」的實例：測試與 CLI 使用，不進 production 路徑。
func Unlimited() *Service { return &Service{unlimited: true, now: time.Now} }

func cacheKey(companyID int) string { return fmt.Sprintf("ent:%d", companyID) }

// state 取得租戶權益來源：快取命中即回，未命中則從 store 組裝並寫回。
// 無訂閱（或訂閱不可用）時回傳 status 對應的狀態，由判定層決定拒絕。
func (s *Service) state(ctx context.Context, companyID int) (*tenantState, error) {
	if s.cache != nil && s.ttl > 0 {
		if raw, ok, err := s.cache.Get(ctx, cacheKey(companyID)); err != nil {
			return nil, err
		} else if ok {
			var st tenantState
			if err := json.Unmarshal(raw, &st); err == nil {
				return &st, nil
			}
			// 反序列化失敗視為未命中（不讓壞快取擋住服務）
		}
	}

	sub, err := s.st.Subscription(ctx, companyID)
	if err != nil {
		return nil, err
	}
	out := &tenantState{CompanyID: companyID, Features: map[string]store.Feature{},
		Entitlements: map[string]store.Entitlement{}}
	if sub == nil {
		out.Status = statusNone
	} else {
		// PlanName 見 tenantState 的說明（目前無來源，維持空字串）。
		out.PlanCode, out.Status, out.TrialEnds = sub.PlanCode, sub.Status, sub.TrialEnds
	}

	features, err := s.st.Features(ctx)
	if err != nil {
		return nil, err
	}
	out.Features = features

	if sub != nil {
		ents, err := s.st.PlanEntitlements(ctx, sub.PlanCode)
		if err != nil {
			return nil, err
		}
		for _, e := range ents {
			out.Entitlements[e.FeatureCode] = e
		}
		overs, err := s.st.Overrides(ctx, companyID)
		if err != nil {
			return nil, err
		}
		out.Overrides = overs
	}

	if s.cache != nil && s.ttl > 0 {
		if raw, err := json.Marshal(out); err == nil {
			if err := s.cache.Set(ctx, cacheKey(companyID), raw, s.ttl); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// unavailable 回報訂閱是否處於「合約不可用」；沒有訂閱同理（沒買就沒有權益）。
func unavailable(status string) bool {
	return status == statusNone || status == statusSuspended || status == statusCancelled
}

// resolved 為單一 feature 的最終權益。
type resolved struct {
	enabled bool
	limit   *int64
}

// resolveFeature 依 tenantState 計算單一 feature 的最終權益（純函式：可單獨測試）。
func resolveFeature(st *tenantState, feature string, now time.Time) (resolved, bool) {
	_, known := st.Features[feature]
	if !known {
		return resolved{}, false // 未定義的功能一律 denied（fail-closed）
	}
	if unavailable(st.Status) {
		return resolved{}, true
	}

	var out resolved
	if e, ok := st.Entitlements[feature]; ok {
		out.enabled, out.limit = e.Enabled, e.Limit
	}
	for _, o := range st.Overrides { // 未撤銷；到期在此過濾；最特定者勝
		if o.FeatureCode != feature {
			continue
		}
		if o.ExpiresAt != nil && !o.ExpiresAt.After(now) {
			continue
		}
		if o.Enabled != nil {
			out.enabled = *o.Enabled
		}
		if o.Limit != nil {
			out.limit = o.Limit
		}
	}
	if st.Status == statusTrialing {
		out.enabled = true // 試用期內不因方案 disabled 而擋（方案仍是上限來源）
	}
	return out, true
}

// Allows 判定 boolean 功能是否可用。
//
// **分工（RPC 守衛必須用 CheckLimit）**：CheckLimit 才是守衛——它會依情境回精確的對外碼與
// details（PLAT-3001 訂閱不可用／PLAT-5002 未含功能／PLAT-5001 額度不足），前端據以導向收款或
// 升級方案。Allows 只回布林、對政策拒絕**不回錯誤**（fail-closed 回 false），供展示與非 RPC
// 路徑使用：把「沒買」「訂閱停了」轉成錯誤會讓每個呼叫點各自發明錯誤碼。
func (s *Service) Allows(ctx context.Context, companyID int, feature string) (bool, error) {
	if s.unlimited {
		return true, nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return false, errcode.SysInternal.Wrap(err)
	}
	r, _ := resolveFeature(st, feature, s.now())
	return r.enabled, nil
}

// CheckLimit 檢查「再加 delta 是否超過上限」；額度不足或未含功能回 FailedPrecondition。
// 不得用 PermissionDenied：那是「缺權限」，前端要導向的是升級方案或收款。
func (s *Service) CheckLimit(ctx context.Context, companyID int, feature string, delta int) error {
	if s.unlimited {
		return nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	if st.Status == statusSuspended || st.Status == statusCancelled {
		// PLAT-3001：合約不可用。先於功能判定——訂閱死了，功能有沒有買都不是重點。
		return errcode.PlatformSubscriptionInactive.Error(nil)
	}
	r, known := resolveFeature(st, feature, s.now())
	if !known || !r.enabled {
		// PLAT-5002：方案／override 未含此功能（含「根本沒訂閱」）。
		return errcode.PlatformFeatureNotInPlan.Error(map[string]string{"feature": feature})
	}
	if r.limit == nil {
		return nil // 不限
	}
	if s.counters == nil {
		return errcode.SysInternal.Wrap(errors.New("entitlements: 未注入 Counter"))
	}
	cur, err := s.counters.Count(ctx, companyID, feature)
	if err != nil {
		return errcode.SysInternal.Wrap(err)
	}
	if cur+delta > int(*r.limit) {
		// PLAT-5001：已達上限；used／limit 進 details 供前端顯示用量。
		return errcode.PlatformLimitExceeded.Error(map[string]string{
			"feature": feature,
			"used":    strconv.Itoa(cur),
			"limit":   strconv.FormatInt(*r.limit, 10),
		})
	}
	return nil
}
