// 平台營運工具的**寫入**(T9):租戶例外、方案價目與權益、operator 白名單、營運參數,
// 以及待收款清單的查詢。
//
// 與 admin.go(唯讀投影)分開的理由:這一檔的每個方法都**改資料**,而且必須與
// platform.audit_logs 落在**同一個交易**(見 store.Admin 的說明與 T9 的服務層)——分成兩個檔案
// 才看得出來「哪些方法是寫入、哪些是查詢」。平台 schema 對業務角色 app_rw 零權限(00029/S9),
// 故一律走 admin(owner)連線;這些交易**不套 RLS**、也不 SET app.current_data_scope
// (那是業務表在租戶 interceptor 交易內的慣例,抄來這裡只會誤導)。
//
// 唯一的例外是 ListReceivables:它讀 companies(業務表),而 companies 是 FORCE RLS 的,
// 故與 admin.go 的投影查詢一樣包在 withSystemScope 內。
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/money"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// zeroRows 把「INSERT…SELECT 沒有寫入任何列」轉成哨兵錯誤:這些寫入的 0 列一律代表
// **來源不存在**(方案不存在／方案已歸檔／功能不存在),不是成功 —— 靜默回 nil 會讓服務層
// 在同一交易內寫下一筆「改了一個不存在的方案」的稽核。
func zeroRows(res sql.Result, sentinel error) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sentinel
	}
	return nil
}

// WithTx 在 admin 連線上開一個交易並把 *sql.Tx 交給 fn;fn 回錯誤即回滾,否則提交。
//
// 與 *Store.WithTx 逐字同形(*Store 服務帳務,*Admin 服務 console 的寫入):兩個型別各自持有
// admin 連線,但交易語意必須一致 —— 「平台的一次寫入(資料＋稽核)是同一個 commit」不該因為
// 呼叫端拿了哪個型別而不同。交易由呼叫端擁有,store 不代開、不代 commit。
func (s *Admin) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// 提交或回滾之後再呼叫 Rollback 會回 sql.ErrTxDone,是無害的 no-op。
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// RecordAuditTx 寫入平台稽核(與資料同一個交易)。reason 必填:空字串即拒絕 —— 動到錢與權限的
// 操作必須留下「為什麼」(表上 NOT NULL,但空字串是合法的 NOT NULL)。
func (s *Admin) RecordAuditTx(ctx context.Context, tx *sql.Tx, operatorID int64,
	action, targetType, targetID, reason string, before, after []byte) error {
	return recordAuditTx(ctx, tx, operatorID, action, targetType, targetID, reason, before, after)
}

// SetTenantOverrideTx 新增一筆租戶例外,回傳新列的 id。
//
// 同一租戶同一功能**不得同時有兩筆生效中的例外**(00029 的 tenant_overrides_active_unique):
// 兩筆互相矛盾的承諾(一筆加席位、一筆減席位)沒有「哪一筆贏」的定義。衝突時回 store.ErrConflict
// 而不是讓 23505 冒上去變成 SYS-9000(那是使用者輸入問題,console 要顯示得出來)—— 以 ON CONFLICT
// 的 0 列判定,比在錯誤訊息裡撈 constraint 名可靠,而且沒有「先查再寫」的競態。
func (s *Admin) SetTenantOverrideTx(ctx context.Context, tx *sql.Tx,
	in store.TenantOverrideInput) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO platform.tenant_overrides
			(company_id, feature_code, enabled, limit_value, reason, owner, expires_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (company_id, feature_code) WHERE revoked_at IS NULL DO NOTHING
		RETURNING id`,
		in.CompanyID, in.FeatureCode, in.Enabled, in.Limit, in.Reason, in.Owner, in.ExpiresAt, in.CreatedBy).
		Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, store.ErrConflict
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

// RevokeTenantOverrideTx 撤銷一筆例外(revoked_at = now()),回帶該租戶與功能(失效快取與稽核要用)。
//
// 已撤銷／不存在 → sql.ErrNoRows:重複撤銷不是「成功」,它會讓稽核上多出一筆「撤銷了某個
// 不存在的承諾」。條件寫在 WHERE 而不是先查再寫,是因為撤銷是**一次性的狀態轉移**(只前進)。
func (s *Admin) RevokeTenantOverrideTx(ctx context.Context, tx *sql.Tx, overrideID int64) (store.OverrideRef, error) {
	var ref store.OverrideRef
	err := tx.QueryRowContext(ctx, `
		UPDATE platform.tenant_overrides
		   SET revoked_at = now()
		 WHERE id = $1 AND revoked_at IS NULL
		RETURNING company_id, feature_code`, overrideID).Scan(&ref.CompanyID, &ref.FeatureCode)
	if errors.Is(err, sql.ErrNoRows) {
		return store.OverrideRef{}, sql.ErrNoRows
	}
	if err != nil {
		return store.OverrideRef{}, err
	}
	return ref, nil
}

// UpsertPlanPriceTx 新增一筆方案價目(調價)。plan_prices 是**價格史**,故寫入一律是 INSERT:
// 改寫既有列會讓「當時的價目」消失,對帳時再也查不出某期是照哪個價開的。
//
// 現行價由 effective_from 最新者決定(plan_prices_plan_effective_idx),故新列的預設
// now() 即為「從此刻生效」。方案不存在 → store.ErrNotFound(0 列的 INSERT...SELECT)。
func (s *Admin) UpsertPlanPriceTx(ctx context.Context, tx *sql.Tx, in store.PlanPriceInput) error {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO platform.plan_prices (plan_id, billing_cycle, base_price, seat_price, currency)
		SELECT p.id, $2, $3::numeric, $4::numeric, COALESCE(NULLIF($5,''), 'TWD')
		  FROM platform.plans p
		 WHERE p.code = $1 AND p.status = 'active'`,
		in.PlanCode, in.BillingCycle, money.FormatCents(in.BaseCents), money.FormatCents(in.SeatCents),
		in.Currency)
	if err != nil {
		return err
	}
	return zeroRows(res, store.ErrNotFound)
}

// SetPlanEntitlementTx 設定(新增或覆寫)方案的一個功能權益。limit 為 nil = 不限額(與 0 不同:
// 0 是「上限 0」)。方案或功能不存在 → store.ErrNotFound(兩個條件都寫在同一句 SELECT 的 WHERE,
// 故 0 列即「其一不存在」;console 顯示「查無此方案／功能」即可,不需要分辨是哪一個)。
func (s *Admin) SetPlanEntitlementTx(ctx context.Context, tx *sql.Tx,
	planCode, featureCode string, enabled bool, limit *int64) error {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value)
		SELECT p.id, f.code, $3, $4
		  FROM platform.plans p, platform.features f
		 WHERE p.code = $1 AND f.code = $2
		ON CONFLICT (plan_id, feature_code) DO UPDATE
		   SET enabled = EXCLUDED.enabled, limit_value = EXCLUDED.limit_value`,
		planCode, featureCode, enabled, limit)
	if err != nil {
		return err
	}
	return zeroRows(res, store.ErrNotFound)
}

// CreateOperatorTx 新增一個 operator(白名單),回傳新列的 id。email 已存在 → store.ErrConflict
// (理由同 SetTenantOverrideTx:唯一鍵的 23505 會被收斂成 5xx,而重複的 email 是輸入問題)。
//
// role 由服務層驗證(只允許 operator／admin);這裡的 COALESCE 只把空字串收斂成表上的預設值。
func (s *Admin) CreateOperatorTx(ctx context.Context, tx *sql.Tx, email, name, role string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO platform.operators (email, name, role)
		VALUES ($1, COALESCE(NULLIF($2,''), ''), COALESCE(NULLIF($3,''), 'operator'))
		ON CONFLICT (email) DO NOTHING
		RETURNING id`, email, name, role).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, store.ErrConflict
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

// DisableOperatorTx 停用一個 operator(status='disabled')。不存在或已停用 → sql.ErrNoRows:
// 重複停用不是「成功」,它會在稽核上留下一筆沒有實際效果的紀錄。
//
// 停用是**即時生效**的:operatorauth 每次請求都查 status(見 operatorauth.Store),故不需要
// 撤銷已簽發的 token。
func (s *Admin) DisableOperatorTx(ctx context.Context, tx *sql.Tx, operatorID int64) error {
	return execOneTx(ctx, tx, `
		UPDATE platform.operators
		   SET status = 'disabled', updated_at = now()
		 WHERE id = $1 AND status <> 'disabled'`, operatorID)
}

// Settings 讀出全部營運參數(key → value)。回傳整份而不是單鍵:console 的設定頁要顯示
// 「目前的值」,逐鍵查詢會讓「改了一鍵、另一鍵顯示舊值」有機會發生。
func (s *Admin) Settings(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM platform.settings`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// UpsertSettingTx 寫入一筆營運參數(seed 亦寫同一張表,故用 upsert 保持冪等)。
func (s *Admin) UpsertSettingTx(ctx context.Context, tx *sql.Tx, key, value string) error {
	return upsertSettingTx(ctx, tx, key, value)
}

// ListReceivables 分頁列出**未付**的期別(含已逾期者),帶公司名稱與方案代碼,供 console 顯示
// 與匯出 CSV;回傳符合條件的總數。
//
// 兩個刻意的選擇:
//   - **排除 G5 的平台自營公司**(identifier = 'platform';同 tenantFilter 的謂詞):它不是租戶,
//     對它開出的期別是平台自己的帳,列進待收款會讓營運對自己催收(console 的租戶投影已排除它,
//     兩處的租戶定義必須一致);
//   - 一律 **status = 'open'**:已付款(paid)與作廢(void)都不是待收款;「已逾期」不是另一個
//     狀態,而是同一列上 period_end < now() 的事實,由呼叫端(或前端)判定 —— 排程補跑時的
//     「過期」跟著呼叫端的時鐘走,不跟著資料庫的(與 cron 的待收款計數同一個立場)。
func (s *Admin) ListReceivables(ctx context.Context, page, pageSize int32) ([]store.ReceivableRow, int, error) {
	const receivablesJoins = `
		  FROM platform.subscription_periods per
		  JOIN platform.subscriptions s ON s.id = per.subscription_id
		  JOIN companies c ON c.id = s.company_id
		       AND c.deleted_at IS NULL AND c.identifier <> 'platform'
		  LEFT JOIN platform.plans p ON p.id = s.plan_id
		 WHERE per.status = 'open'`

	var (
		out   []store.ReceivableRow
		total int
	)
	err := s.withSystemScope(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `SELECT count(*)`+receivablesJoins).Scan(&total); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `
			SELECT c.id::text, c.name, COALESCE(p.code,''), per.period_no,
			       (per.amount*100)::bigint, per.period_end, per.status`+receivablesJoins+`
			 ORDER BY per.period_end, s.company_id, per.period_no
			 LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		out = make([]store.ReceivableRow, 0, pageSize)
		for rows.Next() {
			var r store.ReceivableRow
			var cents int64
			if err := rows.Scan(&r.CompanyID, &r.CompanyName, &r.PlanCode, &r.PeriodNo,
				&cents, &r.PeriodEnd, &r.Status); err != nil {
				return err
			}
			r.Amount = money.FormatCents(cents)
			out = append(out, r)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
