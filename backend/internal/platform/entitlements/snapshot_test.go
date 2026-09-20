package entitlements_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// TestSnapshot 驗租戶端投影：方案／狀態／試用到期 + 每個已知 feature 的判定與用量。
// feature 清單以 store 的定義為準（未定義者不得出現），順序固定（map 走訪不保證順序）。
func TestSnapshot(t *testing.T) {
	trialEnds := time.Now().Add(72 * time.Hour).Truncate(time.Second)
	tests := []struct {
		name     string
		f        *store.Fake
		counts   map[string]int
		wantMeta entitlements.Snapshot
		want     []entitlements.FeatureUsage
	}{
		{
			name: "active：integer 帶用量、boolean 不查計數器（Used 固定 0）",
			f: func() *store.Fake {
				f := store.NewFake()
				f.PutFeature(seatsDef)
				f.PutFeature(printDef)
				f.PutPlan("std", stdPlan)
				sub := store.Subscription{CompanyID: 1, PlanCode: "std", Status: "active", TrialEnds: &trialEnds}
				f.PutSubscription(sub)
				return f
			}(),
			counts:   map[string]int{seats: 3, entitlements.FeaturePrinting: 999},
			wantMeta: entitlements.Snapshot{PlanCode: "std", PlanName: "標準", Status: "active", TrialEndsAt: &trialEnds},
			want: []entitlements.FeatureUsage{
				{FeatureCode: entitlements.FeaturePrinting, Enabled: true},
				{FeatureCode: seats, Enabled: true, Limit: ptr(int64(10)), Used: 3},
			},
		},
		{
			name: "無訂閱：狀態 none，所有 feature 都是 disabled 且用量 0（fail-closed）",
			f: func() *store.Fake {
				f := store.NewFake()
				f.PutFeature(seatsDef)
				f.PutFeature(printDef)
				f.PutPlan("std", stdPlan)
				return f
			}(),
			wantMeta: entitlements.Snapshot{Status: "none"},
			want: []entitlements.FeatureUsage{
				{FeatureCode: entitlements.FeaturePrinting},
				{FeatureCode: seats},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := entitlements.New(tc.f, counting(tc.counts), entitlements.NewMemoryCache(), 0)
			got, err := svc.Snapshot(context.Background(), 1)
			if err != nil {
				t.Fatalf("Snapshot: %v", err)
			}
			if got.Status != tc.wantMeta.Status || got.PlanCode != tc.wantMeta.PlanCode ||
				got.PlanName != tc.wantMeta.PlanName {
				t.Fatalf("方案／狀態 = (%q, %q, %q)；want (%q, %q, %q)",
					got.PlanCode, got.PlanName, got.Status,
					tc.wantMeta.PlanCode, tc.wantMeta.PlanName, tc.wantMeta.Status)
			}
			if (got.TrialEndsAt == nil) != (tc.wantMeta.TrialEndsAt == nil) {
				t.Fatalf("TrialEndsAt = %v；want %v", got.TrialEndsAt, tc.wantMeta.TrialEndsAt)
			}
			if got.TrialEndsAt != nil && !got.TrialEndsAt.Equal(*tc.wantMeta.TrialEndsAt) {
				t.Fatalf("TrialEndsAt = %v；want %v", got.TrialEndsAt, tc.wantMeta.TrialEndsAt)
			}
			if len(got.Usage) != len(tc.want) {
				t.Fatalf("Usage 長度 = %d；want %d（%v）", len(got.Usage), len(tc.want), got.Usage)
			}
			for i, want := range tc.want {
				got := got.Usage[i]
				if got.FeatureCode != want.FeatureCode || got.Enabled != want.Enabled || got.Used != want.Used {
					t.Fatalf("Usage[%d] = %+v；want %+v", i, got, want)
				}
				if (got.Limit == nil) != (want.Limit == nil) {
					t.Fatalf("Usage[%d].Limit = %v；want %v", i, got.Limit, want.Limit)
				}
				if got.Limit != nil && *got.Limit != *want.Limit {
					t.Fatalf("Usage[%d].Limit = %d；want %d", i, *got.Limit, *want.Limit)
				}
			}
		})
	}
}

func TestSnapshotUnlimited(t *testing.T) {
	got, err := entitlements.Unlimited().Snapshot(context.Background(), 1)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if got.PlanCode != "unlimited" || got.Status != "active" {
		t.Fatalf("Unlimited 的投影 = %+v；want plan_code=unlimited／status=active", got)
	}
}
