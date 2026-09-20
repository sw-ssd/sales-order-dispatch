package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// Admin 為平台營運工具(PlatformAdminService)的存取層:跨租戶投影查詢 ＋ 平台稽核寫入。
// 一律 admin(owner)連線:platform schema 對業務角色 app_rw 零權限(00029/S9)。
//
// 稽核一律**與資料同一個交易**(RecordAuditTx,見 admin_writes.go):本型別沒有非交易式的
// 稽核入口 —— 那會允許「稽核說改了、其實沒動」的半成品(T9 移除了 v1 唯讀時期的暫置物)。
//
// 為什麼不併進 Store:
//   - Store 服務權益**判定**(方案／權益／訂閱／例外,供四個業務服務的配額守衛),本型別
//     服務**顯示與稽核**(operator console);
//   - 更硬的理由是同名不同義:Store.PlanEntitlements 只回權益,本型別還要一併帶出完整功能
//     清單(權益矩陣要顯示未設定的格),同一型別上放不下兩個同名方法。
//
// 投影查詢會讀 companies(業務表):spec §6.4 明訂平台方跨租戶視圖走「admin 連線 ＋ 投影查詢」,
// 不開 super 的 data_scope=all 那條路。本檔不 JOIN 寫入、不碰 ent、也沒有跨域 FK。
type Admin struct{ db *sql.DB }

// NewAdmin 建立 Admin(呼叫端負責 admin DSN;業務連線沒有 platform schema 的權限)。
func NewAdmin(db *sql.DB) *Admin { return &Admin{db: db} }

// tenantCols 為租戶投影的欄位清單;tenantJoins 為其來源。列表與詳情共用同一份,欄位順序
// 不會在兩處之間漂移(掃描函式 scanTenant 也是同一份)。
//
// 三個刻意的選擇:
//   - **訂閱以 LATERAL 取「優先未取消、否則最新一筆 cancelled」**:不能把
//     `s.status <> 'cancelled'` 寫進 JOIN 條件 —— 那樣「只有一筆已取消訂閱」的公司會被投影成
//     status=none／無方案／0 席,與「從未訂閱」無法區分,而 proto 明列 cancelled 為合法值、
//     status=cancelled 的篩選也會永遠 0 筆。spec §5.6 的取消是「期末終止、資料不刪除」,
//     營運必須看得到那份合約、方案、席位與到期日(Plan C 的生命週期掃描也要用它)。
//     同一家公司可能同時有歷史 cancelled 與現行訂閱,partial unique index 保證未取消者至多一筆,
//     故此處 LIMIT 1 不會少算,也不會重複列。
//   - current_period_end 取「**已開始**的最後一期」:期別會在到期前 14/7/1 天預先建立(spec §5.4),
//     取 MAX(period_no) 會把還沒到的下一期算成本期到期日 —— 看起來像客戶已經預繳一期;
//   - overdue 為「已過期未付的 open 期別」(spec §5.4 的待收款定義)。本查詢只看 platform 表,
//     不碰業務表的帳務欄位。
const tenantCols = `SELECT c.id, c.name, COALESCE(p.code, ''), COALESCE(p.name, ''),
	       COALESCE(s.status, 'none'), COALESCE(s.seat_count, 0),
	       (SELECT pp.period_end
	          FROM platform.subscription_periods pp
	         WHERE pp.subscription_id = s.id AND pp.period_start <= now()
	         ORDER BY pp.period_no DESC LIMIT 1),
	       EXISTS (SELECT 1 FROM platform.subscription_periods op
	                WHERE op.subscription_id = s.id AND op.status = 'open' AND op.period_end < now())`

const tenantJoins = `
	  FROM companies c
	  LEFT JOIN LATERAL (
	        SELECT s.status, s.seat_count, s.plan_id, s.id
	          FROM platform.subscriptions s
	         WHERE s.company_id = c.id
	         ORDER BY (s.status = 'cancelled'), s.started_at DESC, s.id DESC
	         LIMIT 1
	       ) s ON true
	  LEFT JOIN platform.plans p ON p.id = s.plan_id`

// 排序鍵的第一項是「已取消」的布林:ASC 讓 false(未取消)排在最前面,取消的才排在後面
// —— 寫成 `status <> 'cancelled'` 會反向(未取消 = true 反而排在後面),投影就會挑到歷史合約。

// tenantFilter 為列表與計數共用的篩選($1 = keyword 原文判斷空、$2 = LIKE 樣式、$3 = 訂閱狀態)。
//
// keyword 一律先 trim 再由服務層傳入:前後空白的 keyword 不得變成「篩掉全部」的樣式 %  %。
//
// identifier <> 'platform' 排除 **G5 的平台自營公司**(`cmd/seed/platform.go` 的
// platformCompanyIdentifier):它是平台自己的系統 actor 錨點,不是租戶。不排除的話 console 會把
// 它當成一個租戶而對平台自己計費/凍結(Plan C 的生命週期掃描 `ActiveOrTrialingSubscriptions`
// 等查詢必須用同一條排除規則)。
const tenantFilter = `
	 WHERE c.deleted_at IS NULL
	   AND c.identifier <> 'platform'
	   AND ($1 = '' OR c.name ILIKE $2 OR c.identifier ILIKE $2)
	   AND ($3 = '' OR COALESCE(s.status, 'none') = $3)`

// ListTenants 分頁列出租戶(含未訂閱者),可依 keyword(公司名稱／識別碼模糊)與訂閱狀態篩選,
// 回傳符合篩選的總數(分頁用)。
//
// page／pageSize 由服務層正規化後傳入(page ≥ 1、1 ≤ pageSize ≤ maxPageSize):上下限只留在
// 服務層一處,免得兩個地方各有一套(改了一邊就會出現「第 0 頁」這種查詢)。
func (s *Admin) ListTenants(ctx context.Context, keyword, status string, page, pageSize int32) ([]store.TenantRow, int, error) {
	pattern := likeContains(keyword)
	var (
		out   []store.TenantRow
		total int
	)
	err := s.withSystemScope(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `SELECT count(*)`+tenantJoins+tenantFilter,
			keyword, pattern, status).Scan(&total); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, tenantCols+tenantJoins+tenantFilter+`
	 ORDER BY c.id
	 LIMIT $4 OFFSET $5`, keyword, pattern, status, pageSize, (page-1)*pageSize)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		out = make([]store.TenantRow, 0, pageSize)
		for rows.Next() {
			row, err := scanTenant(rows)
			if err != nil {
				return err
			}
			out = append(out, row)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// GetTenant 取單一租戶的概況與其未撤銷的例外;公司不存在(或已軟刪除)回 store.ErrNotFound。
//
// companyID 為 bigint 的文字形(proto 是 string):服務層已驗證格式,這裡的 ParseInt 失敗
// 屬程式錯誤(回原錯誤,SYS-9000),不是使用者輸入問題。
func (s *Admin) GetTenant(ctx context.Context, companyID string) (*store.TenantRow, []store.TenantOverrideRow, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(companyID), 10, 64)
	if err != nil {
		return nil, nil, err
	}
	var (
		tenant    store.TenantRow
		overrides []store.TenantOverrideRow
	)
	err = s.withSystemScope(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, tenantCols+tenantJoins+`
	 WHERE c.id = $1 AND c.deleted_at IS NULL`, id)
		tenant, err = scanTenant(row)
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrNotFound
		}
		if err != nil {
			return err
		}
		overrides, err = tenantOverrides(ctx, tx, id)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return &tenant, overrides, nil
}

// withSystemScope 在系統範圍(scope=all)的交易內執行 fn。
//
// companies 已 ENABLE ＋ FORCE RLS(00028),而 FORCE **讓 table owner 也受 policy 約束**
// (00025/00028 檔頭:生產的 owner 不是 superuser)→ 連 admin(owner)連線都必須先
// `SET LOCAL app.current_data_scope = 'all'` 才讀得到跨租戶的公司列。少了這一層的失效模式
// 不報錯、只是**靜默回 0 列** —— console 顯示「沒有任何租戶」,而 migration 檔頭點名的正是這件事。
//
// 為什麼是交易而非 session 級 SET:連線池的 session 會被下一個請求重用,殘留的 scope=all
// 會讓租戶請求讀到全庫。交易結束即失效是唯一安全的範圍。platform schema 本身不套 RLS,
// 故只有「讀業務表」的查詢包在這裡。
func (s *Admin) withSystemScope(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// commit 後 Rollback 回 ErrTxDone(no-op);失敗路徑一律回滾,沒有殘留的 scope=all 連線。
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SET LOCAL app.current_data_scope = 'all'`); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// tenantOverrides 取某公司未撤銷的例外(已到期者照樣回傳;到期與否由判定層／UI 判斷)。
func tenantOverrides(ctx context.Context, q queryer, companyID int64) ([]store.TenantOverrideRow, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, feature_code, enabled, limit_value, reason, owner, expires_at
		  FROM platform.tenant_overrides
		 WHERE company_id = $1 AND revoked_at IS NULL
		 ORDER BY feature_code`, companyID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []store.TenantOverrideRow
	for rows.Next() {
		var (
			id      int64
			row     store.TenantOverrideRow
			enabled sql.NullBool
			limit   sql.NullInt64
			expires sql.NullTime
		)
		if err := rows.Scan(&id, &row.FeatureCode, &enabled, &limit, &row.Reason, &row.Owner, &expires); err != nil {
			return nil, err
		}
		row.ID = strconv.FormatInt(id, 10)
		if enabled.Valid {
			v := enabled.Bool
			row.Enabled = &v
		}
		if limit.Valid {
			v := limit.Int64
			row.Limit = &v
		}
		if expires.Valid {
			v := expires.Time
			row.ExpiresAt = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// ListPlans 回傳全部方案(含已歸檔)與各計費週期的現行價目,依 sort_order／id 排序。
//
// 現行價目在 SQL 端以 DISTINCT ON 取每週期 effective_from 最新者:調價後 plan_prices 只增不減,
// 在 Go 端分組會讓「哪一筆才是現行價」變成散在各處的判斷。
func (s *Admin) ListPlans(ctx context.Context) ([]store.PlanRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.code, p.name, p.status, p.sort_order,
		       pr.billing_cycle, pr.base_price, pr.seat_price, pr.currency, pr.effective_from
		  FROM platform.plans p
		  LEFT JOIN LATERAL (
		        SELECT DISTINCT ON (billing_cycle)
		               billing_cycle,
		               base_price::text AS base_price,
		               seat_price::text AS seat_price,
		               currency, effective_from
		          FROM platform.plan_prices
		         WHERE plan_id = p.id
		         ORDER BY billing_cycle, effective_from DESC
		       ) pr ON true
		 ORDER BY p.sort_order, p.id, pr.billing_cycle`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []store.PlanRow
	for rows.Next() {
		var (
			id        int64
			plan      store.PlanRow
			sortOrder int32
			cycle     sql.NullString
			base      sql.NullString
			seat      sql.NullString
			currency  sql.NullString
			effective sql.NullTime
		)
		if err := rows.Scan(&id, &plan.Code, &plan.Name, &plan.Status, &sortOrder,
			&cycle, &base, &seat, &currency, &effective); err != nil {
			return nil, err
		}
		plan.ID = strconv.FormatInt(id, 10)
		plan.SortOrder = sortOrder
		// 同一方案的多列(每個計費週期一列)在 SQL 端相鄰:以 id 分組即可,不需要 map
		// (map 會打亂 ORDER BY 建立的順序,而價目的順序是 UI 的顯示順序)。
		if n := len(out); n > 0 && out[n-1].ID == plan.ID {
			out[n-1].Prices = append(out[n-1].Prices, planPriceRow(cycle, base, seat, currency, effective))
			continue
		}
		if cycle.Valid {
			plan.Prices = []store.PlanPriceRow{planPriceRow(cycle, base, seat, currency, effective)}
		}
		out = append(out, plan)
	}
	return out, rows.Err()
}

// planPriceRow 把一個 LEFT JOIN LATERAL 的價目欄位轉為 DTO(全部欄位都可能為 NULL:方案沒有價目)。
func planPriceRow(cycle, base, seat, currency sql.NullString, effective sql.NullTime) store.PlanPriceRow {
	return store.PlanPriceRow{
		BillingCycle: cycle.String, BasePrice: base.String, SeatPrice: seat.String,
		Currency: currency.String, EffectiveFrom: effective.Time,
	}
}

// PlanEntitlements 回傳某方案(以 code 指定)的權益 ＋ **完整**功能清單(權益矩陣要顯示未設定
// 的格,故清單不可只回有權益的項目)。查無此方案回 store.ErrNotFound —— 「方案沒有權益」與
// 「方案不存在」必須分得開,否則 console 會把打錯的 code 顯示成空矩陣。
func (s *Admin) PlanEntitlements(ctx context.Context, planCode string) ([]store.Entitlement, []store.Feature, error) {
	var planID int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM platform.plans WHERE code = $1`, planCode).Scan(&planID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, store.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	entRows, err := s.db.QueryContext(ctx, `
		SELECT feature_code, enabled, limit_value
		  FROM platform.plan_entitlements
		 WHERE plan_id = $1
		 ORDER BY feature_code`, planID)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = entRows.Close() }()

	var ents []store.Entitlement
	for entRows.Next() {
		var (
			e     store.Entitlement
			limit sql.NullInt64
		)
		if err := entRows.Scan(&e.FeatureCode, &e.Enabled, &limit); err != nil {
			return nil, nil, err
		}
		if limit.Valid {
			v := limit.Int64
			e.Limit = &v
		}
		ents = append(ents, e)
	}
	if err := entRows.Err(); err != nil {
		return nil, nil, err
	}

	featureRows, err := s.db.QueryContext(ctx, `
		SELECT code, type, unit, description
		  FROM platform.features
		 ORDER BY code`)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = featureRows.Close() }()

	var features []store.Feature
	for featureRows.Next() {
		var f store.Feature
		if err := featureRows.Scan(&f.Code, &f.Type, &f.Unit, &f.Description); err != nil {
			return nil, nil, err
		}
		features = append(features, f)
	}
	return ents, features, featureRows.Err()
}

// ListPlatformAudit 分頁列出平台稽核(新到舊),可依 target_type／target_id 篩選,回傳符合
// 篩選的總數。
//
// 排序為 created_at DESC, id DESC:created_at 同值(同一批寫入)時仍要有**穩定**順序,
// 否則翻頁會漏掉或重複列。
func (s *Admin) ListPlatformAudit(ctx context.Context, targetType, targetID string, page, pageSize int32) ([]store.PlatformAuditRow, int, error) {
	const filter = `
	 WHERE ($1 = '' OR a.target_type = $1)
	   AND ($2 = '' OR a.target_id = $2)`

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM platform.audit_logs a`+filter,
		targetType, targetID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, o.email, a.action, a.target_type, a.target_id, a.reason, a.created_at
		  FROM platform.audit_logs a
		  JOIN platform.operators o ON o.id = a.operator_id`+filter+`
		 ORDER BY a.created_at DESC, a.id DESC
		 LIMIT $3 OFFSET $4`, targetType, targetID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]store.PlatformAuditRow, 0, pageSize)
	for rows.Next() {
		var (
			id  int64
			row store.PlatformAuditRow
		)
		if err := rows.Scan(&id, &row.OperatorEmail, &row.Action, &row.TargetType, &row.TargetID,
			&row.Reason, &row.CreatedAt); err != nil {
			return nil, 0, err
		}
		row.ID = strconv.FormatInt(id, 10)
		out = append(out, row)
	}
	return out, total, rows.Err()
}

// scanTenant 讀取 tenantCols 的一列(*sql.Row 與 *sql.Rows 共用)。
func scanTenant(sc rowScanner) (store.TenantRow, error) {
	var (
		id        int64
		row       store.TenantRow
		periodEnd sql.NullTime
	)
	if err := sc.Scan(&id, &row.CompanyName, &row.PlanCode, &row.PlanName, &row.Status,
		&row.SeatCount, &periodEnd, &row.Overdue); err != nil {
		return store.TenantRow{}, err
	}
	row.CompanyID = strconv.FormatInt(id, 10)
	if periodEnd.Valid {
		v := periodEnd.Time
		row.CurrentPeriodEnd = &v
	}
	return row, nil
}

// rowScanner 為 *sql.Row 與 *sql.Rows 的共用面(兩者的 Scan 形狀相同)。
type rowScanner interface {
	Scan(dest ...any) error
}

// queryer 為 *sql.DB 與 *sql.Tx 的共用面:同一個查詢在交易內外都要能跑(例外查詢在
// withSystemScope 的交易內執行,因為它與租戶概況必須是同一個快照)。
type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// likeContains 把使用者輸入轉為 ILIKE 的「包含」樣式,並跳脫 LIKE 的萬用字元。
//
// 為什麼要跳脫:keyword 打 "%" 會變成「符合全部」,打 "_" 會變成「任一字元」—— 使用者以為在
// 搜一個字面字元,卻得到另一種結果。跳脫字元在 PG 的 ILIKE 預設是反斜線。
func likeContains(keyword string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(keyword) + "%"
}
