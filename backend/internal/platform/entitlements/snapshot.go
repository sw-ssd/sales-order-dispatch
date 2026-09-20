package entitlements

import (
	"context"
	"sort"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// FeatureUsage 為單一 feature 的判定結果與用量（供租戶端投影；boolean 的 Used 為 0）。
type FeatureUsage struct {
	FeatureCode string
	Enabled     bool
	Limit       *int64
	Used        int
}

// Snapshot 為租戶的權益全貌（租戶端投影；只回自己公司）。
type Snapshot struct {
	PlanCode    string
	PlanName    string
	Status      string
	TrialEndsAt *time.Time
	Usage       []FeatureUsage
}

// Snapshot 組裝租戶權益與用量；feature 清單以 store 的定義為準（未定義者不出現）。
// 前端據此 disable 按鈕與顯示用量——**前端 disable 不構成授權**，真正的擋在 CheckLimit 與 RLS。
func (s *Service) Snapshot(ctx context.Context, companyID int) (*Snapshot, error) {
	if s.unlimited {
		return &Snapshot{PlanCode: "unlimited", PlanName: "不限", Status: "active"}, nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	out := &Snapshot{PlanCode: st.PlanCode, PlanName: st.PlanName, Status: st.Status, TrialEndsAt: st.TrialEnds}
	for code, def := range st.Features {
		r, _ := resolveFeature(st, code, s.now())
		u := FeatureUsage{FeatureCode: code, Enabled: r.enabled, Limit: r.limit}
		if def.Type == "integer" && s.counters != nil {
			n, err := s.counters.Count(ctx, companyID, code)
			if err != nil {
				return nil, errcode.SysInternal.Wrap(err)
			}
			u.Used = n
		}
		out.Usage = append(out.Usage, u)
	}
	sort.Slice(out.Usage, func(i, j int) bool { return out.Usage[i].FeatureCode < out.Usage[j].FeatureCode })
	return out, nil
}
