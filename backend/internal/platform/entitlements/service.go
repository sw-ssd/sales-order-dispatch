// Package entitlements 判定「這個租戶現在可以做什麼」：方案預設 ⊕ 未過期 override，
// 並在訂閱不可用時全面 fail-closed。與 OpenFGA 的角色授權是**不同軸**：
// 這裡管「買了沒有」，那裡管「誰能做」。
package entitlements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

// 訂閱狀態（platform.subscriptions.status）。判定採 **allow-list**（見 usable）：
// **新增狀態必須明確決定它是否可用 —— 未列舉＝不可用**（fail-closed）。
//
// **契約警語（operator 必讀）：刪除訂閱列＝不施加限制** —— statusNone 的語意是「尚未開通計費」，
// 判定層對它不施加任何配額／功能限制（見 Allows／CheckLimit）。要停用租戶必須把 status 設成
// cancelled／suspended，**不得刪列**；否則一次 DELETE 就等於送一個不限額方案。
const (
	// statusNone：沒有未取消的訂閱（Store 回 nil）。這是**尚未開通計費**，不是「已判定不可用」：
	// Plan C 的訂閱指派／onboarding 落地前沒有任何程式會建立 platform.subscriptions 列（今天每個
	// 真實公司都是這個狀態），所以在這裡 fail-closed 等於「員工自助註冊與首次 OIDC 登入全被硬擋，
	// 只有進得去的管理員才能補訂閱」＝上線即癱瘓。
	//
	// 因此 Allows／CheckLimit 對它**不施加限制**（等同 Unlimited），只留一行 log 讓它可見
	// （見 state）。真正的 fail-closed 針對**已知不可用**與**未列舉**的狀態：suspended／cancelled／
	// 任意未知字串一律 PLAT-3001（見 usable）。
	statusNone      = "none"
	statusTrialing  = "trialing"  // 可用：試用中
	statusActive    = "active"    // 可用
	statusPastDue   = "past_due"  // 可用：仍在寬限內（催收由 dunning job 改狀態，判定層不自行推算 grace_until）
	statusSuspended = "suspended" // 不可用
	statusCancelled = "cancelled" // 不可用（store 契約上不回，但換一個 store 實作就可能看到）
)

// usable 為訂閱狀態的 **allow-list**：只有明確可用的狀態才走正常權益路徑，其餘一律視為合約不可用。
//
// **新增狀態必須明確決定它是否可用 —— 未列舉＝不可用**。為什麼不用 deny-list（只列不可用者）：
// platform.subscriptions.status 在 00029 **沒有 CHECK 約束**，漏列一個狀態（新狀態、拼字錯誤、
// 大小寫不同、空字串）就會讓該租戶靜默恢復全部權益 —— 症狀是「訂閱停了的租戶照常寫資料」，
// 沒有任何測試或 log 會指出來。allow-list 的相反失誤（新狀態明明可用卻被擋）是吵的、立刻會被發現。
func usable(status string) bool {
	switch status {
	case statusTrialing, statusActive, statusPastDue:
		return true
	}
	return false
}

// Counter 由業務域提供（於 server.InitDomains 注入）：判定層不認得業務 schema，
// 只認 feature code → 目前用量的對照。回傳錯誤即視為系統錯誤（不得當成 0 放行）。
type Counter interface {
	Count(ctx context.Context, companyID int, feature string) (int, error)
}

// tenantState 為單一租戶的權益來源快照（快取的單位）。
type tenantState struct {
	CompanyID int    `json:"company_id"`
	PlanCode  string `json:"plan_code"`
	// PlanName 為方案名（store 由 JOIN plans.name 帶出），租戶端投影顯示用。
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
		raw, ok, err := s.cache.Get(ctx, cacheKey(companyID))
		if err != nil {
			// **快取是加速器，不是資料來源**：Valkey 掛掉（或設定錯）時回源，不得因此拒絕服務。
			// 判定是配額守衛的來源，讓它失敗等於全站業務寫入失敗 —— 那比「這陣子每個租戶都打一次
			// 平台庫」貴得多。log 要吵：效能問題必須看得見，但不可以變成可用性問題。
			log.Printf("entitlements: 權益快取讀取失敗(company=%d)，改為回源: %v", companyID, err)
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
		// 無訂閱列＝尚未開通計費 → 判定層不施加限制（見 statusNone）。留一行 log：這個狀態在
		// Plan C 的訂閱指派落地前是常態，而「無聲地不限制」正是最該被看見的事（快取命中時不重記）。
		log.Printf("entitlements: 公司 %d 無訂閱列（尚未開通計費）→ 不施加配額限制"+
			"（要停用租戶請設 status=suspended/cancelled，勿刪列）", companyID)
	} else {
		out.PlanCode, out.PlanName, out.Status, out.TrialEnds = sub.PlanCode, sub.PlanName, sub.Status, sub.TrialEnds
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
				// 寫不進去只是「下次還要回源」，同樣不得讓判定失敗（見上方 Get 的說明）。
				log.Printf("entitlements: 權益快取寫入失敗(company=%d)，本次不快取: %v", companyID, err)
			}
		}
	}
	return out, nil
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
	if !usable(st.Status) {
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
//
// 訂閱狀態的判定見 statusNone 的 doc：**沒有訂閱列＝尚未開通計費 → 不施加限制**（回 true）。
func (s *Service) Allows(ctx context.Context, companyID int, feature string) (bool, error) {
	if s.unlimited {
		return true, nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return false, errcode.SysInternal.Wrap(err)
	}
	if st.Status == statusNone {
		return true, nil // 尚未開通計費（log 已由 state 記）
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
	// 沒有訂閱列＝尚未開通計費 → 不施加限制（見 statusNone 的 doc；log 已由 state 記）。
	// 這裡**不是** fail-open：無限制是這個狀態的既定語意，而「已判定不可用」與「未列舉」的狀態
	// 一律 PLAT-3001（下一行）。
	if st.Status == statusNone {
		return nil
	}
	// 其餘不可用狀態（suspended／cancelled／**任何未列舉者**）一律 PLAT-3001，且先於功能判定：
	// 合約死了，功能有沒有買都不是重點。
	if !usable(st.Status) {
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
