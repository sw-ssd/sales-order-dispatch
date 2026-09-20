// 平台域 seeder（D34／G5）：7 個 features、3 個方案（免費／標準／專業）與價目、方案權益、
// 首位 operator，以及 G5 的平台自營公司＋系統使用者與 platform.settings 的營運參數。
//
// 冪等：重跑 seed 不新增列，也不覆寫營運已調整的值（價目與營運參數只補缺）—— `task seed`
// 會反覆執行。
//
// **兩種連線各用在哪一段**（由呼叫端建立，main.go 已有同一條 owner 連線供兩用）：
//   - `platform.*`（features／plans／plan_prices／plan_entitlements／operators／settings）：
//     00029 刻意不套 RLS、且 app_rw 對此 schema **零權限**，故直接以 admin（owner）連線
//     `db *sql.DB` 執行 —— seedPlatformCatalog／seedPlatformSettings。
//   - `companies`／`users`（G5 的平台自營公司與系統使用者）：00028 已 ENABLE＋FORCE RLS，
//     FORCE 讓 table owner 也受 policy 約束 → **必須**用 dbtenant.NewClient 建立的 client 並在
//     dbtenant.SystemScopeTx（scope=all）內執行，否則 SQLSTATE 42501 ——
//     seedPlatformSystemActor（實測見 platform_integration_test.go）。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// platformCompanyIdentifier 為平台自營公司的識別碼（G5 的系統 actor 錨點）。
const platformCompanyIdentifier = "platform"

// platformFeatures 為 v1 定案清單（spec §4.5）：4 個數值上限 ＋ 3 個 boolean 功能。
//
// `limit.storage_gb` **刻意不在清單內**（controller 2026-09-20 裁定）：計數器
// （internal/services/counters.go）沒有檔案空間的來源可以量，而判定層的 Snapshot 對每個
// integer feature 都要用量 → 種了它只會讓租戶端權益投影對**所有租戶**失敗。檔案功能（P2-1）
// 落地並補上計數器後再加回來。
var platformFeatures = []struct{ Code, Type, Unit, Desc string }{
	{"limit.seats", "integer", "席", "帳號席位上線"},
	{"limit.customers", "integer", "客戶", "客戶筆數上線"},
	{"limit.products", "integer", "商品", "商品筆數上線"},
	{"limit.departments", "integer", "部門", "部門數上線"},
	{"feature.printing", "boolean", "", "單據列印"},
	{"feature.dispatch", "boolean", "", "派車看板"},
	{"feature.returns", "boolean", "", "退貨申請與審核"},
}

// platformPlan 為一個起始方案：Entitle 列出即 enabled；值 > 0 = 上限、-1 = enabled 但不限
// （limit_value 為 NULL）。未列出的 feature 即未含（判定層對缺席者 fail-closed）。
type platformPlan struct {
	Code, Name string
	SortOrder  int
	Entitle    map[string]int64
}

var platformPlans = []platformPlan{
	{"free", "免費", 1, map[string]int64{
		"limit.seats": 3, "limit.customers": 50, "limit.products": 100,
		"limit.departments": 1,
	}},
	{"std", "標準", 2, map[string]int64{
		"limit.seats": 10, "limit.customers": 500, "limit.products": 2000,
		"limit.departments": 5,
		"feature.printing":  -1,
	}},
	{"pro", "專業", 3, map[string]int64{
		"limit.seats": 50, "limit.customers": -1, "limit.products": -1,
		"limit.departments": 20,
		"feature.printing":  -1, "feature.dispatch": -1, "feature.returns": -1,
	}},
}

// planPrice 為方案的月繳預設價（字串原樣交給 numeric，避免浮點誤差進定價）。
type planPrice struct{ Base, Seat string }

// planPriceDefaults 依方案 code 取 env（SEED_PRICE_{FREE,STD,PRO}_{BASE,SEAT}）提供的預設價。
// 值一律先驗證再寫入：壞值默默當成 0 等於把定價寫成免費。
func planPriceDefaults(cfg config.Platform) (map[string]planPrice, error) {
	out := map[string]planPrice{
		"free": {cfg.SeedPriceFreeBase, cfg.SeedPriceFreeSeat},
		"std":  {cfg.SeedPriceStdBase, cfg.SeedPriceStdSeat},
		"pro":  {cfg.SeedPriceProBase, cfg.SeedPriceProSeat},
	}
	for _, p := range platformPlans {
		for _, v := range []string{out[p.Code].Base, out[p.Code].Seat} {
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return nil, fmt.Errorf("方案 %s 的 seed 價目不是數字（%q）: %w", p.Code, v, err)
			}
		}
	}
	return out, nil
}

// yearlyPrice 由月費推年繳價：月費 × 12 × 0.9（取整到元）。v1 的年繳折扣率固定 9 折；
// SEED_PRICE_* 只是**首次建立**的預設值，上線前務必改為真實定價。
func yearlyPrice(monthly string) string {
	f, _ := strconv.ParseFloat(monthly, 64) // 呼叫前已由 planPriceDefaults 驗證
	return strconv.FormatFloat(math.Round(f*12*0.9), 'f', 0, 64)
}

// SeedPlatform 冪等建立平台域基礎資料。
//
// 設定一律由 config.Platform 帶入（env 驅動），使「上線前改預設值」不必改程式碼：
//   - SeedOperatorEmail／SeedOperatorName：首位平台操作者（未設 email → 跳過，見下）
//   - SeedSystemActorEmail：G5 的系統 actor
//   - DefaultTrialDays／DefaultGraceDays／DefaultLeadDays：首次寫入 platform.settings
//   - SeedPrice{Free,Std,Pro}{Base,Seat}：方案價目的首次預設值（佔位數字，上線前務必改）
//
// 連線分工見檔頭：platform.* 走 db；companies／users 走 client＋SystemScopeTx。
func SeedPlatform(ctx context.Context, db *sql.DB, client *ent.Client, cfg config.Platform) error {
	prices, err := planPriceDefaults(cfg)
	if err != nil {
		return err
	}
	if err := seedPlatformCatalog(ctx, db, cfg, prices); err != nil {
		return err
	}
	// G5 的平台自營公司與系統使用者是**業務表** → 這一步自己開系統範圍交易。
	actorID, err := seedPlatformSystemActor(ctx, client, cfg)
	if err != nil {
		return err
	}
	return seedPlatformSettings(ctx, db, cfg, actorID)
}

// seedPlatformCatalog 建立 features、方案、價目、方案權益與首位 operator（全部走 platform schema）。
func seedPlatformCatalog(ctx context.Context, db *sql.DB, cfg config.Platform, prices map[string]planPrice) error {
	for _, f := range platformFeatures {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.features (code, type, unit, description)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (code) DO UPDATE SET type = EXCLUDED.type, unit = EXCLUDED.unit,
				description = EXCLUDED.description`, f.Code, f.Type, f.Unit, f.Desc); err != nil {
			return fmt.Errorf("seed feature %s: %w", f.Code, err)
		}
	}

	for _, p := range platformPlans {
		var planID int64
		if err := db.QueryRowContext(ctx, `
			INSERT INTO platform.plans (code, name, sort_order)
			VALUES ($1,$2,$3)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, sort_order = EXCLUDED.sort_order
			RETURNING id`, p.Code, p.Name, p.SortOrder).Scan(&planID); err != nil {
			return fmt.Errorf("seed plan %s: %w", p.Code, err)
		}

		// 價目**只補缺**：plan_prices 沒有 (plan_id, billing_cycle) 唯一鍵（同一方案可多次調價、
		// 以 effective_from 取當期價），故以 NOT EXISTS 表達「只補缺的週期」—— 重跑 seed 不會把
		// 營運調過的價格蓋回去。
		price := prices[p.Code]
		for _, row := range []struct{ cycle, base, seat string }{
			{"monthly", price.Base, price.Seat},
			{"yearly", yearlyPrice(price.Base), yearlyPrice(price.Seat)},
		} {
			if _, err := db.ExecContext(ctx, `
				INSERT INTO platform.plan_prices (plan_id, billing_cycle, base_price, seat_price, currency)
				SELECT $1,$2,$3::numeric,$4::numeric,'TWD'
				 WHERE NOT EXISTS (
					SELECT 1 FROM platform.plan_prices WHERE plan_id = $1 AND billing_cycle = $2)`,
				planID, row.cycle, row.base, row.seat); err != nil {
				return fmt.Errorf("seed price %s/%s: %w", p.Code, row.cycle, err)
			}
		}

		for code, limit := range p.Entitle {
			// 列出即 enabled；-1 = 不限（limit_value NULL）。
			var limitValue any
			if limit > 0 {
				limitValue = limit
			}
			if _, err := db.ExecContext(ctx, `
				INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value)
				VALUES ($1,$2,true,$3)
				ON CONFLICT (plan_id, feature_code)
				DO UPDATE SET enabled = EXCLUDED.enabled, limit_value = EXCLUDED.limit_value`,
				planID, code, limitValue); err != nil {
				return fmt.Errorf("seed entitlement %s/%s: %w", p.Code, code, err)
			}
		}
	}

	return seedPlatformOperator(ctx, db, cfg)
}

// seedPlatformOperator 種下首位平台操作者（`platform.operators` 白名單＝能登入平台工具的人）。
//
// email 由 env（PLATFORM_SEED_OPERATOR_EMAIL）提供且**無預設**：未設就跳過並提示 ——
// 不得把版控裡的真實 email 種成可登入帳號。
// 白名單**只種首位**（表中已有任何 operator 就不再種）：改 env 不會多出第二個可登入者，
// 換人要經營運工具（Plan C 的 operator 管理）而不是重跑 seed。
func seedPlatformOperator(ctx context.Context, db *sql.DB, cfg config.Platform) error {
	if cfg.SeedOperatorEmail == "" {
		log.Println("seed 平台域: 未設 PLATFORM_SEED_OPERATOR_EMAIL，跳過首位 operator（平台工具登入白名單請自行以營運工具設定）")
		return nil
	}
	res, err := db.ExecContext(ctx, `
		INSERT INTO platform.operators (email, name, role)
		SELECT $1,$2,'admin'
		 WHERE NOT EXISTS (SELECT 1 FROM platform.operators)`, cfg.SeedOperatorEmail, cfg.SeedOperatorName)
	if err != nil {
		return fmt.Errorf("seed operator: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("seed operator: %w", err)
	}
	if n == 0 {
		log.Printf("seed 平台域: 已有 operator，略過首位 operator（%s）", cfg.SeedOperatorEmail)
	}
	return nil
}

// seedPlatformSystemActor 在**系統範圍交易**內建立 G5 的平台自營公司與系統使用者，
// 回傳系統使用者的 users.id（供 platform.settings.system_actor_user_id）。
//
// 為何非這個交易不可：companies／users 是業務表且 00028 對它們 ENABLE＋FORCE RLS，FORCE 讓
// table owner 也受 policy 約束 —— 少了 scope=all，連 owner 連線都以 42501 失敗
// （且 SET LOCAL 是 RLS driver 裝飾器在 Tx(ctx) 內套的，未經 dbtenant.NewClient 的裸 client
// 包進 SystemScopeTx 一樣沒用）。
func seedPlatformSystemActor(ctx context.Context, client *ent.Client, cfg config.Platform) (int64, error) {
	if cfg.SeedSystemActorEmail == "" {
		return 0, fmt.Errorf("PLATFORM_SEED_SYSTEM_ACTOR_EMAIL 不可為空：排程與 consumer 的稽核 actor 需要真實的 users 列")
	}
	var actorID int64
	err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		id, err := ensurePlatformSystemActor(ctx, tx.Client(), cfg)
		actorID = id
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("seed 平台自營公司與系統使用者: %w", err)
	}
	return actorID, nil
}

// ensurePlatformSystemActor 冪等建立系統使用者；呼叫端必須已進入系統範圍交易。
//
// password_hash 固定 '!' 是刻意的**不可登入**哨兵值：此帳號只作為排程／consumer 的稽核 actor
// （developer 帳號同一慣例），'!' 不是任何密碼的雜湊，且不得有任何人以它登入。
func ensurePlatformSystemActor(ctx context.Context, client *ent.Client, cfg config.Platform) (int64, error) {
	companyID, err := ensurePlatformCompany(ctx, client)
	if err != nil {
		return 0, err
	}
	u, err := client.User.Query().Where(user.EmailEQ(cfg.SeedSystemActorEmail)).Only(ctx)
	switch {
	case err == nil:
		return int64(u.ID), nil
	case !ent.IsNotFound(err):
		return 0, fmt.Errorf("查詢系統使用者: %w", err)
	}
	u, err = client.User.Create().
		SetEmail(cfg.SeedSystemActorEmail).
		SetName("系統排程").
		SetRole("super").
		SetStatus(user.StatusActive).
		SetPasswordHash("!").
		SetCompanyID(companyID).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("建立系統使用者: %w", err)
	}
	return int64(u.ID), nil
}

// ensurePlatformCompany 冪等建立平台自營公司（identifier='platform'）。
// 唯一性由**部分**唯一索引（WHERE deleted_at IS NULL）表達，部分索引不能當衝突目標 → 先查後建。
func ensurePlatformCompany(ctx context.Context, client *ent.Client) (int, error) {
	c, err := client.Company.Query().
		Where(company.IdentifierEQ(platformCompanyIdentifier), company.DeletedAtIsNil()).
		Only(ctx)
	switch {
	case err == nil:
		return c.ID, nil
	case !ent.IsNotFound(err):
		return 0, fmt.Errorf("查詢平台自營公司: %w", err)
	}
	c, err = client.Company.Create().
		SetName("平台營運").
		SetIdentifier(platformCompanyIdentifier).
		SetStatus(company.StatusActive).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("建立平台自營公司: %w", err)
	}
	return c.ID, nil
}

// seedPlatformSettings 寫入系統 actor 與營運參數（試用／寬限／提前天數）。
//
// platform.settings 由 **Plan C（生命週期／console 計畫）的 migration** 建立，Plan B 內尚不存在
// → 沒有這張表時跳過並提示，而不是把別人的 migration 抄成 seeder 裡的 DDL（migration 歸屬要清楚）。
// 缺表不影響其餘 seed，也不影響 Plan B：讀 settings 的 cron／consumer 同樣在 Plan C。
func seedPlatformSettings(ctx context.Context, db *sql.DB, cfg config.Platform, actorID int64) error {
	ready, err := platformSettingsReady(ctx, db)
	if err != nil {
		return err
	}
	if !ready {
		log.Println("seed 平台域: platform.settings 尚未建立（Plan C 的 migration），跳過系統 actor 與營運參數的 settings 寫入")
		return nil
	}
	// 系統 actor 的 id 由 seed 決定 → 可覆寫（自癒：值被改壞時重跑 seed 會修正）。
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES ('system_actor_user_id', $1)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		strconv.FormatInt(actorID, 10)); err != nil {
		return fmt.Errorf("seed system_actor_user_id: %w", err)
	}
	// 營運參數**只補缺**：營運調過的天數不得被重跑 seed 蓋回去（之後由 UpdateBillingSettings 維護）。
	for _, kv := range []struct {
		key   string
		value int
	}{
		{"trial_days", cfg.DefaultTrialDays},
		{"grace_days", cfg.DefaultGraceDays},
		{"lead_days", cfg.DefaultLeadDays},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO platform.settings (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO NOTHING`, kv.key, strconv.Itoa(kv.value)); err != nil {
			return fmt.Errorf("seed setting %s: %w", kv.key, err)
		}
	}
	return nil
}

// platformSettingsReady 回報 platform.settings 是否已存在（Plan C 的 migration 尚未落地時為 false）。
func platformSettingsReady(ctx context.Context, db *sql.DB) (bool, error) {
	var ready bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('platform.settings') IS NOT NULL`).Scan(&ready); err != nil {
		return false, fmt.Errorf("檢查 platform.settings: %w", err)
	}
	return ready, nil
}
