//go:build integration

// 平台域 seed 的整合契約（D34／G5）：7 個 features、3 個方案與價目、方案權益、首位 operator、
// 平台自營公司與系統使用者、platform.settings 的系統 actor 與營運參數。
// 驗收要求：**連跑兩次，筆數不變**（`task seed` 會反覆執行）。
//
// 兩條連線的分工在本檔被真的驗到：
//   - platform.*（00029 刻意不套 RLS）→ admin（owner）連線；
//   - companies／users（00028 ENABLE＋FORCE RLS 的業務表）→ dbtenant.NewClient ＋ SystemScopeTx。
//     最後一支測試以 **app_rw**（NOBYPASSRLS）連線反證「少了系統範圍就 42501」—— 用 superuser
//     測不出東西（PG 的 superuser 永遠繞過 RLS，FORCE 亦然）。
package main

import (
	"database/sql"
	"log"
	"maps"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// platformTestConfig 為測試用 seed 設定：email 一律用 example.com（**不得**用版控裡的真實 email）。
func platformTestConfig() config.Platform {
	return config.Platform{
		SeedOperatorEmail:    "ops@example.com",
		SeedOperatorName:     "平台維運",
		SeedSystemActorEmail: "system@example.com",
		DefaultTrialDays:     14,
		DefaultGraceDays:     7,
		DefaultLeadDays:      14,
		SeedPriceFreeBase:    "0",
		SeedPriceFreeSeat:    "0",
		SeedPriceStdBase:     "1500",
		SeedPriceStdSeat:     "150",
		SeedPriceProBase:     "4500",
		SeedPriceProSeat:     "150",
	}
}

// newPlatformSeedFixture 起一台全新容器、套用遷移，回傳 admin 連線、業務 ent client 與 seed 設定。
// 生產的 seed 是同一條 owner 連線同時供兩域；測試刻意分成兩條 —— client.Close() 會關掉底層
// *sql.DB，共用會讓 admin 連線提前失效。
func newPlatformSeedFixture(t *testing.T) (*sql.DB, *ent.Client, config.Platform) {
	t.Helper()
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateForSeed(t, dsn)
	admin := openSeedAdmin(t, dsn)
	client := dbtenant.NewClient(openSeedAdmin(t, dsn))
	t.Cleanup(func() { _ = client.Close() })
	return admin, client, platformTestConfig()
}

// platformSeedCounts 為 seed 後應穩定的列數（全部以 admin 連線取真值）。
type platformSeedCounts struct {
	features, plans, prices, entitlements, operators, companies, users int
}

func platformSeedCountsOf(t *testing.T, db *sql.DB) platformSeedCounts {
	t.Helper()
	return platformSeedCounts{
		features:     seedCount(t, db, `SELECT count(*) FROM platform.features`),
		plans:        seedCount(t, db, `SELECT count(*) FROM platform.plans`),
		prices:       seedCount(t, db, `SELECT count(*) FROM platform.plan_prices`),
		entitlements: seedCount(t, db, `SELECT count(*) FROM platform.plan_entitlements`),
		operators:    seedCount(t, db, `SELECT count(*) FROM platform.operators`),
		companies:    seedCount(t, db, `SELECT count(*) FROM companies WHERE identifier = 'platform'`),
		users: seedCount(t, db, `SELECT count(*) FROM users u JOIN companies c ON c.id = u.company_users
			WHERE c.identifier = 'platform'`),
	}
}

// seedString 以 admin 連線取單一字串真值。
func seedString(t *testing.T, db *sql.DB, query string) string {
	t.Helper()
	var s string
	if err := db.QueryRow(query).Scan(&s); err != nil {
		t.Fatalf("查詢 %q：%v", query, err)
	}
	return s
}

// TestIntegrationSeedPlatformIdempotent 驗 v1 平台域 seed 的內容與冪等。
func TestIntegrationSeedPlatformIdempotent(t *testing.T) {
	admin, client, cfg := newPlatformSeedFixture(t)
	ctx := t.Context()

	// ① 設定壞掉 → fail-fast，且**不得**留下半套資料。價目字串來自 env（SEED_PRICE_*），
	// 壞值默默當成 0 等於把定價寫成免費。
	bad := cfg
	bad.SeedPriceStdBase = "一千五"
	if err := SeedPlatform(ctx, admin, client, bad); err == nil {
		t.Fatal("價目不是數字必須回錯誤（不得默默寫 0）")
	}
	if n := seedCount(t, admin, `SELECT count(*) FROM platform.features`); n != 0 {
		t.Fatalf("設定驗證必須在任何寫入之前，已寫入 %d 個 feature", n)
	}

	// platform.settings 由 Plan C 的 migration 建立（Plan B 內尚不存在）；本測試自行建表，
	// 才能驗證系統 actor 與營運參數真的寫得進去（沒有這張表時的行為見下一支測試）。
	createPlatformSettingsTable(t, admin)

	// ② 連跑兩次：第二次不得新增任何列。
	if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
		t.Fatalf("第 1 次 seed：%v", err)
	}
	first := platformSeedCountsOf(t, admin)
	if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
		t.Fatalf("第 2 次 seed（冪等）：%v", err)
	}
	if second := platformSeedCountsOf(t, admin); second != first {
		t.Fatalf("重跑 seed 不得改變列數：\n第一次 %+v\n第二次 %+v", first, second)
	}
	want := platformSeedCounts{features: 7, plans: 3, prices: 6, entitlements: 16, operators: 1}
	want.companies, want.users = 1, 1
	if first != want {
		t.Fatalf("seed 內容不符：got %+v want %+v", first, want)
	}

	assertPlatformCatalog(t, admin)
	assertPlatformOperator(t, admin, cfg)
	assertPlatformSystemActor(t, admin, cfg)
	assertPlatformSettings(t, admin, cfg)

	// ③ 營運調整過的值不得被重跑 seed 蓋回去：價目與營運參數都只補缺。
	tuned := cfg
	tuned.SeedPriceStdBase, tuned.SeedPriceStdSeat = "9999", "999"
	tuned.DefaultTrialDays = 99
	if err := SeedPlatform(ctx, admin, client, tuned); err != nil {
		t.Fatalf("調價後重跑 seed：%v", err)
	}
	if base := seedString(t, admin, `SELECT p.base_price::text FROM platform.plan_prices p
		JOIN platform.plans pl ON pl.id = p.plan_id
		WHERE pl.code = 'std' AND p.billing_cycle = 'monthly'`); base != "1500.00" {
		t.Fatalf("已有價目不得被重跑 seed 覆寫，got %q", base)
	}
	if days := seedString(t, admin, `SELECT value FROM platform.settings WHERE key = 'trial_days'`); days != "14" {
		t.Fatalf("營運參數不得被重跑 seed 覆寫，got %q", days)
	}
	if got := platformSeedCountsOf(t, admin); got != first {
		t.Fatalf("調價後重跑仍不得改變列數：got %+v want %+v", got, first)
	}
}

// TestIntegrationSeedPlatformOperatorEmailUnset 驗首位 operator 的環境變數契約：
// 未設 `PLATFORM_SEED_OPERATOR_EMAIL` → 跳過並印提示（**不得**把版控裡的真實 email 種進
// platform.operators 白名單）；補上設定後才種，且白名單只種一次。
func TestIntegrationSeedPlatformOperatorEmailUnset(t *testing.T) {
	admin, client, cfg := newPlatformSeedFixture(t)
	ctx := t.Context()
	cfg.SeedOperatorEmail = ""

	var logs strings.Builder
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	for i := range 2 {
		if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
			t.Fatalf("第 %d 次 seed（未設 operator email）：%v", i+1, err)
		}
	}
	if n := seedCount(t, admin, `SELECT count(*) FROM platform.operators`); n != 0 {
		t.Fatalf("未設 email 時不得種下任何 operator，got %d 列", n)
	}
	if !strings.Contains(logs.String(), "PLATFORM_SEED_OPERATOR_EMAIL") {
		t.Fatalf("未設 email 必須印出提示，got log=%q", logs.String())
	}
	// 本測試的庫沒有 platform.settings（Plan C 的 migration 尚未落地）→ 跳過該段但不得中斷 seed。
	if !strings.Contains(logs.String(), "platform.settings") {
		t.Fatalf("缺 platform.settings 必須印出提示並繼續，got log=%q", logs.String())
	}
	if n := seedCount(t, admin, `SELECT count(*) FROM platform.features`); n != 7 {
		t.Fatalf("跳過 operator 不得影響其他 seed，features=%d", n)
	}

	// 補上 email → 種下首位 operator（跳過不是永久的）。
	cfg.SeedOperatorEmail = "ops@example.com"
	if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
		t.Fatalf("補上 operator email 後 seed：%v", err)
	}
	if got := seedString(t, admin, `SELECT email || '/' || name || '/' || role FROM platform.operators`); got != "ops@example.com/平台維運/admin" {
		t.Fatalf("首位 operator 內容不符，got %q", got)
	}

	// 改 email 再跑 → 不得多出第二個可登入者（白名單只種首位；換人要用營運工具）。
	cfg.SeedOperatorEmail = "other@example.com"
	if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
		t.Fatalf("換 email 後重跑 seed：%v", err)
	}
	if n := seedCount(t, admin, `SELECT count(*) FROM platform.operators`); n != 1 {
		t.Fatalf("重跑不得新增 operator，got %d 列", n)
	}
}

// TestIntegrationSeedPlatformActorNeedsSystemScope 以 app_rw（NOBYPASSRLS）證明 G5 的業務表寫入
// 一定要在系統範圍交易內：00028 的 policy 讓沒有 scope 的連線**寫不進去**（WITH CHECK → 42501），
// 而不是「看不到就算成功」。
func TestIntegrationSeedPlatformActorNeedsSystemScope(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateForSeed(t, dsn)
	admin := openSeedAdmin(t, dsn)
	app := openSeedAppRole(t, dsn)
	client := dbtenant.NewClient(app)
	t.Cleanup(func() { _ = client.Close() })
	cfg := platformTestConfig()
	ctx := t.Context()

	// ① 不帶系統範圍 → 42501（SeedPlatform 的 G5 段若忘了包 SystemScopeTx 就是這個下場）。
	if _, err := ensurePlatformSystemActor(ctx, client, cfg); !isSeedRLSViolation(err) {
		t.Fatalf("未在系統範圍內建立平台自營公司必須是 RLS 違反(42501)，got %v", err)
	}
	if n := seedCount(t, admin, `SELECT count(*) FROM companies WHERE identifier = 'platform'`); n != 0 {
		t.Fatalf("被擋下的寫入不得留下公司列，got %d", n)
	}

	// ② SeedPlatform 自己開系統範圍交易 → 成功且冪等；platform.*（app_rw 零權限）走 admin 連線。
	for i := range 2 {
		if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
			t.Fatalf("第 %d 次 seed（app_rw 業務連線）：%v", i+1, err)
		}
	}
	counts := platformSeedCountsOf(t, admin)
	if counts.companies != 1 || counts.users != 1 {
		t.Fatalf("系統範圍內的 seed 應恰建一家自營公司與一位系統使用者，got %+v", counts)
	}
	if counts.features != 7 || counts.operators != 1 {
		t.Fatalf("平台域內容不符，got %+v", counts)
	}
}

// TestIntegrationSeedPlatformDoesNotAnchorDeveloper 驗平台自營公司不會被當成 developer 帳號的
// 錨點。實測（真容器 + `go run ./cmd/seed` 連跑兩次）：第一次 seed 時庫裡還沒有任何公司 → 略過
// developer；第二次 seed 因為自營公司已存在 → `firstCompanyID` 挑到它 → **共用開發者帳號被建進
// 平台自營租戶**（平台公司的 users 由 1 變 2）。自營公司是系統自己的租戶，不得住任何租戶帳號。
func TestIntegrationSeedPlatformDoesNotAnchorDeveloper(t *testing.T) {
	admin, client, cfg := newPlatformSeedFixture(t)
	ctx := t.Context()
	if err := SeedPlatform(ctx, admin, client, cfg); err != nil {
		t.Fatalf("seed：%v", err)
	}
	if id := firstCompanyID(ctx, client); id != 0 {
		t.Fatalf("平台自營公司不得作為 developer 錨點，got companyID=%d", id)
	}

	// 一般公司仍是合法錨點（否則開發環境的 developer 帳號永遠建不出來）。
	var companyID int
	if err := admin.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ('種子測試公司', 'SEED-T2', 'active') RETURNING id`).
		Scan(&companyID); err != nil {
		t.Fatalf("建公司錨點：%v", err)
	}
	if id := firstCompanyID(ctx, client); id != companyID {
		t.Fatalf("一般公司仍應可作為錨點，got companyID=%d want %d", id, companyID)
	}
}

// createPlatformSettingsTable 建立 Plan C（生命週期／console 計畫）的 platform.settings。
// Plan B 內這張表還不存在（00029 沒有它），而 G5 要求 seed 把系統 actor 與營運參數寫進去 →
// 測試自行建表以驗證那條路徑；DDL 逐字對齊 Plan C 的定義（key／value／updated_at）。
//
// **Plan C 必須照這個形狀建（欄位型別與值的形狀是跨計畫契約）**：
//   - 兩欄皆 TEXT：`key text PRIMARY KEY`、`value text NOT NULL`（＋`updated_at timestamptz`）；
//   - `system_actor_user_id`：**users.id 的十進位字串**（seed 以 strconv.FormatInt 寫入；
//     讀取端 strconv.ParseInt → id）；
//   - `trial_days`／`grace_days`／`lead_days`：**十進位整數字串**（seed 以 strconv.Itoa 寫入）。
//
// 值的形狀由測試釘住：`system_actor_user_id` 必須等於系統使用者的 `id::text`。
func createPlatformSettingsTable(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS platform.settings (
			key        text PRIMARY KEY,
			value      text NOT NULL,
			updated_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		t.Fatalf("建立 platform.settings（Plan C 的 migration 產物）：%v", err)
	}
}

// assertPlatformCatalog 逐條釘住 v1 清單：8 個 feature（code／type／unit）、3 個方案、
// 月繳與年繳價目（年繳＝月費 × 12 × 0.9，取整到元）、方案的權益（列出即 enabled；-1 = 不限）。
func assertPlatformCatalog(t *testing.T, db *sql.DB) {
	t.Helper()
	if got, want := seedMap(t, db, `SELECT code, type || '/' || unit FROM platform.features`), map[string]string{
		"limit.seats":       "integer/席",
		"limit.customers":   "integer/客戶",
		"limit.products":    "integer/商品",
		"limit.departments": "integer/部門",
		"feature.printing":  "boolean/",
		"feature.dispatch":  "boolean/",
		"feature.returns":   "boolean/",
	}; !maps.Equal(got, want) {
		t.Fatalf("features 不符：\ngot  %v\nwant %v", got, want)
	}

	if got, want := seedMap(t, db, `SELECT code, name || '/' || sort_order::text FROM platform.plans`), map[string]string{
		"free": "免費/1", "std": "標準/2", "pro": "專業/3",
	}; !maps.Equal(got, want) {
		t.Fatalf("方案不符：\ngot  %v\nwant %v", got, want)
	}

	if got, want := seedMap(t, db, `SELECT pl.code || '/' || p.billing_cycle,
		p.base_price::text || '/' || p.seat_price::text || '/' || p.currency
		FROM platform.plan_prices p JOIN platform.plans pl ON pl.id = p.plan_id`), map[string]string{
		"free/monthly": "0.00/0.00/TWD",
		"free/yearly":  "0.00/0.00/TWD",
		"std/monthly":  "1500.00/150.00/TWD",
		"std/yearly":   "16200.00/1620.00/TWD",
		"pro/monthly":  "4500.00/150.00/TWD",
		"pro/yearly":   "48600.00/1620.00/TWD",
	}; !maps.Equal(got, want) {
		t.Fatalf("價目不符（年繳應為月費 × 12 × 0.9）：\ngot  %v\nwant %v", got, want)
	}

	// 權益逐列比對（含「未列出＝未含」：免費方案不得有 feature.printing 等列）。
	got := seedMap(t, db, `SELECT pl.code || '/' || e.feature_code,
		e.enabled::text || '/' || coalesce(e.limit_value::text, 'NULL')
		FROM platform.plan_entitlements e JOIN platform.plans pl ON pl.id = e.plan_id`)
	want := map[string]string{
		"free/limit.seats":       "true/3",
		"free/limit.customers":   "true/50",
		"free/limit.products":    "true/100",
		"free/limit.departments": "true/1",
		"std/limit.seats":        "true/10",
		"std/limit.customers":    "true/500",
		"std/limit.products":     "true/2000",
		"std/limit.departments":  "true/5",
		"std/feature.printing":   "true/NULL", // -1：enabled 但不限
		"pro/limit.seats":        "true/50",
		"pro/limit.customers":    "true/NULL",
		"pro/limit.products":     "true/NULL",
		"pro/limit.departments":  "true/20",
		"pro/feature.printing":   "true/NULL",
		"pro/feature.dispatch":   "true/NULL",
		"pro/feature.returns":    "true/NULL",
	}
	if !maps.Equal(got, want) {
		t.Fatalf("方案權益不符：\ngot  %v\nwant %v", got, want)
	}
}

// assertPlatformOperator 驗首位 operator 落成白名單的 admin。
func assertPlatformOperator(t *testing.T, db *sql.DB, cfg config.Platform) {
	t.Helper()
	got := seedString(t, db, `SELECT email || '/' || name || '/' || role || '/' || status FROM platform.operators`)
	want := cfg.SeedOperatorEmail + "/" + cfg.SeedOperatorName + "/admin/active"
	if got != want {
		t.Fatalf("首位 operator 不符，got %q want %q", got, want)
	}
}

// assertPlatformSystemActor 驗 G5 的平台自營公司與系統使用者（業務表）。
func assertPlatformSystemActor(t *testing.T, db *sql.DB, cfg config.Platform) {
	t.Helper()
	if got := seedString(t, db, `SELECT name || '/' || status FROM companies WHERE identifier = 'platform'`); got != "平台營運/active" {
		t.Fatalf("平台自營公司不符，got %q", got)
	}
	// password_hash='!' 是不可登入哨兵值（不是任何密碼的雜湊）—— 此帳號只作為系統 actor。
	got := seedString(t, db, `SELECT u.email || '/' || u.name || '/' || u.role || '/' || u.status || '/' || u.password_hash
		FROM users u JOIN companies c ON c.id = u.company_users WHERE c.identifier = 'platform'`)
	want := cfg.SeedSystemActorEmail + "/系統排程/super/active/!"
	if got != want {
		t.Fatalf("系統使用者不符，got %q want %q", got, want)
	}
	if n := seedCount(t, db, `SELECT count(*) FROM users WHERE password_hash <> '!'`); n != 0 {
		t.Fatalf("seed 不得建立任何可登入帳號，got %d 列", n)
	}
}

// assertPlatformSettings 驗系統 actor 的 id 與營運參數真的寫入 platform.settings。
func assertPlatformSettings(t *testing.T, db *sql.DB, cfg config.Platform) {
	t.Helper()
	actorID := seedString(t, db, `SELECT value FROM platform.settings WHERE key = 'system_actor_user_id'`)
	if userID := seedString(t, db, `SELECT u.id::text FROM users u JOIN companies c ON c.id = u.company_users
		WHERE c.identifier = 'platform'`); actorID != userID {
		t.Fatalf("system_actor_user_id 應指向系統使用者，got %q want %q", actorID, userID)
	}
	for key, want := range map[string]string{
		"trial_days": strconv.Itoa(cfg.DefaultTrialDays),
		"grace_days": strconv.Itoa(cfg.DefaultGraceDays),
		"lead_days":  strconv.Itoa(cfg.DefaultLeadDays),
	} {
		if got := seedString(t, db, `SELECT value FROM platform.settings WHERE key = '`+key+`'`); got != want {
			t.Fatalf("settings %s 不符，got %q want %q", key, got, want)
		}
	}
}

// seedMap 以 admin 連線取出「兩欄 → map」的真值。
func seedMap(t *testing.T, db *sql.DB, query string) map[string]string {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("查詢 %q：%v", query, err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			t.Fatalf("掃描 %q：%v", query, err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("讀取 %q：%v", query, err)
	}
	return out
}
