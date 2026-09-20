// Package postgres 以 database/sql 實作平台 store(00029)。
//
// 連線一律是 admin(owner)連線:platform schema 對業務角色 app_rw 零權限(S9),平台域的
// 存取不得走業務連線。平台域不用 ent(它不是租戶資料,沒有 RLS 與 codegen 的需求),
// 故這裡是手寫 SQL;查詢的欄位名與表名與 00029 逐字對應。
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

type Store struct{ db *sql.DB }

var _ store.Store = (*Store)(nil)

func New(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Features(ctx context.Context) (map[string]store.Feature, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT code, type, unit, description FROM platform.features`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]store.Feature{}
	for rows.Next() {
		var f store.Feature
		if err := rows.Scan(&f.Code, &f.Type, &f.Unit, &f.Description); err != nil {
			return nil, err
		}
		out[f.Code] = f
	}
	return out, rows.Err()
}

func (s *Store) PlanEntitlements(ctx context.Context, planCode string) ([]store.Entitlement, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT pe.feature_code, pe.enabled, pe.limit_value
		  FROM platform.plan_entitlements pe
		  JOIN platform.plans p ON p.id = pe.plan_id
		 WHERE p.code = $1`, planCode)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []store.Entitlement
	for rows.Next() {
		var e store.Entitlement
		var limit sql.NullInt64
		if err := rows.Scan(&e.FeatureCode, &e.Enabled, &limit); err != nil {
			return nil, err
		}
		// NULL = 不限,必須維持 nil(0 代表「一個都不給」)。
		if limit.Valid {
			v := limit.Int64
			e.Limit = &v
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Overrides 只剔除已撤銷者;到期的照樣回傳(到期與否由判定層判斷)。
func (s *Store) Overrides(ctx context.Context, companyID int) ([]store.Override, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT company_id, feature_code, enabled, limit_value, expires_at
		  FROM platform.tenant_overrides
		 WHERE company_id = $1 AND revoked_at IS NULL
		 ORDER BY created_at DESC`, companyID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []store.Override
	for rows.Next() {
		var o store.Override
		var enabled sql.NullBool
		var limit sql.NullInt64
		var expires sql.NullTime
		if err := rows.Scan(&o.CompanyID, &o.FeatureCode, &enabled, &limit, &expires); err != nil {
			return nil, err
		}
		// NULL 代表「此維度不覆寫」,不是 false／0。
		if enabled.Valid {
			v := enabled.Bool
			o.Enabled = &v
		}
		if limit.Valid {
			v := limit.Int64
			o.Limit = &v
		}
		if expires.Valid {
			v := expires.Time
			o.ExpiresAt = &v
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// Subscription 回傳未取消的訂閱;沒有則回 (nil, nil)。billing_cycle 必須帶出:期別產生
// 靠它決定 +1 月或 +1 年(G1)。方案名取自 JOIN 的 plans.name(租戶端投影要顯示它)。
func (s *Store) Subscription(ctx context.Context, companyID int) (*store.Subscription, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT s.company_id, p.code, p.name, s.status, p.id, s.seat_count,
		       s.billing_cycle, s.trial_ends_at, s.grace_until
		  FROM platform.subscriptions s
		  JOIN platform.plans p ON p.id = s.plan_id
		 WHERE s.company_id = $1 AND s.status <> 'cancelled'`, companyID)
	var sub store.Subscription
	var trial, grace sql.NullTime
	err := row.Scan(&sub.CompanyID, &sub.PlanCode, &sub.PlanName, &sub.Status, &sub.PlanID, &sub.SeatCount,
		&sub.BillingCycle, &trial, &grace)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if trial.Valid {
		v := trial.Time
		sub.TrialEnds = &v
	}
	if grace.Valid {
		v := grace.Time
		sub.GraceUntil = &v
	}
	return &sub, nil
}
