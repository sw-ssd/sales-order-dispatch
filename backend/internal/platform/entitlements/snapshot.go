package entitlements

import (
	"context"
	"errors"
	"log"
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
//
// 未結項 #33（Plan B #13，已關）：無訂閱列（status=none）時用量列為空 ——
// 「尚未開通計費」（空列）與「方案不含」（enabled=false 的列）在投影形狀上可區分，
// 前端仍以 isUnprovisioned（status=none 或 planCode 空）顯示未開通文案。守衛面不受影響：
// Allows／CheckLimit 對 none 一律不施加限制（見 service.go 的 statusNone），與投影無關。
//
// **缺計數器的降級（不得讓整筆投影失敗）**：integer feature 若沒有對應的計數器（新的 feature
// 忘記配 Count 分支）或計數失敗，**略過該筆用量並記一行 log**——不回 0（0 會被前端讀成「用量為
// 零」）、不回 -1、也不把整個投影變成 SYS-9000。理由：一個配置錯誤不得升級成全站故障（所有租戶
// 的權益頁同時掛掉）；守衛面仍 fail-closed（CheckLimit 遇到同樣的錯誤一律拒絕，不受此影響）。
func (s *Service) Snapshot(ctx context.Context, companyID int) (*Snapshot, error) {
	if s.unlimited {
		return &Snapshot{PlanCode: "unlimited", PlanName: "不限", Status: "active"}, nil
	}
	st, err := s.state(ctx, companyID)
	if err != nil {
		return nil, errcode.SysInternal.Wrap(err)
	}
	out := &Snapshot{PlanCode: st.PlanCode, PlanName: st.PlanName, Status: st.Status, TrialEndsAt: st.TrialEnds}
	if st.Status == statusNone {
		// 尚未開通計費：用量列為空（不列 enabled=false 的假列）。空列即「未開通」，
		// 與「方案不含」的 false 列形狀不同，前端不再需要猜。
		return out, nil
	}
	for code, def := range st.Features {
		r, _ := resolveFeature(st, code, s.now())
		u := FeatureUsage{FeatureCode: code, Enabled: r.enabled, Limit: r.limit}
		if def.Type == "integer" {
			n, err := s.count(ctx, companyID, code)
			if err != nil {
				log.Printf("entitlements: 略過 feature %q 的用量（公司 %d）：%v", code, companyID, err)
				continue
			}
			u.Used = n
		}
		out.Usage = append(out.Usage, u)
	}
	sort.Slice(out.Usage, func(i, j int) bool { return out.Usage[i].FeatureCode < out.Usage[j].FeatureCode })
	return out, nil
}

// count 取 feature 的當前用量；**未注入計數器**同樣視為「沒有這個 feature 的計數器」而回錯誤
// （呼叫端據以略過該筆），不得當成 0。
func (s *Service) count(ctx context.Context, companyID int, feature string) (int, error) {
	if s.counters == nil {
		return 0, errors.New("未注入計數器")
	}
	return s.counters.Count(ctx, companyID, feature)
}
