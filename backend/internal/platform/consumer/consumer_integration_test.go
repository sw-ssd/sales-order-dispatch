//go:build integration

// Task 6 的整合驗收(真 PostgreSQL):單元測試的假交易驗不到「認領與狀態變更同一個 commit」的
// 真正語意(真交易、真回滾、真欄位落地),故這裡再看四件事:
//
//	① subscription.expired(G7)→ 公司在同一交易內被凍結,且租戶稽核的 actor 是系統 actor
//	   (platform.settings.system_actor_user_id 的**租戶 users.id**,不是平台 operator);
//	② 事件在同一交易內被認領:dispatched_at 落地、attempts 計數;
//	③ 沒有產品域動作的型別(period.opened)只被認領,不留任何稽核;
//	④ 重跑:不再認領(attempts 不變)、狀態與稽核也不變(可重跑)。
//
// 執行:`task test:integration -- -count=1 -run TestIntegrationDispatchOutboxEvents -v`
package consumer_test

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// consumerMigrationsDir 相對套件目錄(go test 以套件目錄為 cwd),與 cmd/migrate 同路徑。
const consumerMigrationsDir = "../../../database/migrations"

func TestIntegrationDispatchOutboxEvents(t *testing.T) {
	testsupport.RequiresContainer(t)
	ctx := t.Context()
	adminDB, err := sql.Open("pgx", testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = adminDB.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.RunContext(ctx, "up", adminDB, consumerMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	companyID, actorID := seedTenant(t, ctx, adminDB)
	if _, err := adminDB.ExecContext(ctx, `
		INSERT INTO platform.settings (key, value) VALUES ('system_actor_user_id', $1)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		strconv.Itoa(actorID)); err != nil {
		t.Fatalf("寫 platform.settings: %v", err)
	}

	// 排程(任務 5)在「已取消且期末已過」時寫的事件,外加一筆 consumer 沒有產品域動作的型別
	// (payload 仍照契約帶 company_id)。
	var expiredID, unmappedID int64
	if err := adminDB.QueryRowContext(ctx, `
		INSERT INTO platform.events (aggregate_type, aggregate_id, event_type, payload)
		VALUES ('subscription', 3, 'subscription.expired',
		        jsonb_build_object('company_id', $1::bigint, 'subscription_id', 3,
		                           'reason', 'cancelled_at_period_end'))
		RETURNING id`, companyID).Scan(&expiredID); err != nil {
		t.Fatalf("寫 expired 事件: %v", err)
	}
	if err := adminDB.QueryRowContext(ctx, `
		INSERT INTO platform.events (aggregate_type, aggregate_id, event_type, payload)
		VALUES ('subscription', 4, 'period.opened', jsonb_build_object('company_id', $1::bigint))
		RETURNING id`, companyID).Scan(&unmappedID); err != nil {
		t.Fatalf("寫未對應型別事件: %v", err)
	}

	c := consumer.New(postgres.New(adminDB), consumer.NewDBSystemTx(adminDB), consumer.ProductDomain{})
	n, err := c.DispatchOnce(ctx, 100)
	if err != nil {
		t.Fatalf("派送: %v", err)
	}
	if n != 2 {
		t.Fatalf("兩筆事件都應被認領(未對應型別只認領),got %d", n)
	}

	// ① 公司在同一交易內被凍結,且稽核的 actor 是系統 actor(租戶 users.id)。
	status := companyStatus(t, ctx, adminDB, companyID)
	if status != string(company.StatusSuspended) {
		t.Fatalf("expired(G7)應把公司凍結為 suspended,got %q", status)
	}
	var audits, auditUser int
	var afterStatus, afterReason string
	if err := adminDB.QueryRowContext(ctx, `
		SELECT count(*) OVER (), user_id, after_snapshot->>'status', after_snapshot->>'reason'
		  FROM audit_logs WHERE company_id = $1 AND resource_type = 'company'`, companyID).
		Scan(&audits, &auditUser, &afterStatus, &afterReason); err != nil {
		t.Fatalf("查租戶稽核: %v", err)
	}
	if audits != 1 {
		t.Fatalf("凍結應留 1 筆租戶稽核(未對應型別不得留痕),got %d", audits)
	}
	if auditUser != actorID {
		t.Fatalf("稽核主體應為系統 actor(users.id=%d),got %d", actorID, auditUser)
	}
	if afterStatus != string(company.StatusSuspended) || afterReason == "" {
		t.Fatalf("稽核的 after 快照應帶新狀態與原因,got status=%q reason=%q", afterStatus, afterReason)
	}

	// ② 兩筆事件都在同一交易內被認領(attempts 是同一張表的計數欄)。
	for _, id := range []int64{expiredID, unmappedID} {
		var dispatched time.Time
		var attempts int
		if err := adminDB.QueryRowContext(ctx,
			`SELECT dispatched_at, attempts FROM platform.events WHERE id = $1`, id).
			Scan(&dispatched, &attempts); err != nil {
			t.Fatalf("事件 %d 應已認領(dispatched_at 落地): %v", id, err)
		}
		if attempts != 1 {
			t.Fatalf("事件 %d 應正好認領一次,got attempts=%d", id, attempts)
		}
	}

	// ④ 重跑(排程每日一趟、手動補跑):不再認領、不再變更狀態、不再寫稽核。
	again, err := c.DispatchOnce(ctx, 100)
	if err != nil {
		t.Fatalf("第二趟: %v", err)
	}
	if again != 0 {
		t.Fatalf("第二趟應無事可做,got %d", again)
	}
	var attempts, reAudits int
	if err := adminDB.QueryRowContext(ctx,
		`SELECT attempts FROM platform.events WHERE id = $1`, expiredID).Scan(&attempts); err != nil {
		t.Fatalf("重跑後查事件: %v", err)
	}
	if err := adminDB.QueryRowContext(ctx,
		`SELECT count(*) FROM audit_logs WHERE company_id = $1`, companyID).Scan(&reAudits); err != nil {
		t.Fatalf("重跑後查稽核: %v", err)
	}
	if reStatus := companyStatus(t, ctx, adminDB, companyID); reStatus != status {
		t.Fatalf("重跑不得再變更狀態: got %q want %q", reStatus, status)
	}
	if reAudits != audits || attempts != 1 {
		t.Fatalf("重跑不得產生第二個副作用: audit=%d(原 %d) attempts=%d(應 1)", reAudits, audits, attempts)
	}

	// ⑤ 認領是**條件式**的:來源若硬是回了一筆「別的執行已經派送」的事件(讀取與認領之間的競態,
	// 單飛鎖失效時真的會發生),條件式 UPDATE 必須拒絕第二次 —— attempts 不變、不留第二筆稽核。
	// 這一條是「重跑不得產生第二個副作用」在真 PG 上唯一有鑑別力的斷言(SetCompanyStatus 對同值
	// 本來就是 no-op,稽核數看不出重複認領;attempts 看得出來)。
	var payload []byte
	if err := adminDB.QueryRowContext(ctx,
		`SELECT payload FROM platform.events WHERE id = $1`, expiredID).Scan(&payload); err != nil {
		t.Fatalf("讀事件 payload: %v", err)
	}
	stale := consumer.New(
		&staleEvents{events: []store.Event{{ID: expiredID, AggregateType: "subscription",
			AggregateID: 3, EventType: "subscription.expired", Payload: payload}}, actor: actorID},
		consumer.NewDBSystemTx(adminDB), consumer.ProductDomain{})
	if n, err := stale.DispatchOnce(ctx, 100); err != nil {
		t.Fatalf("被搶先認領不得算失敗: %v", err)
	} else if n != 0 {
		t.Fatalf("條件式認領必須拒絕第二次,got 派送 %d 筆", n)
	}
	var afterAttempts, afterAudits int
	if err := adminDB.QueryRowContext(ctx,
		`SELECT attempts FROM platform.events WHERE id = $1`, expiredID).Scan(&afterAttempts); err != nil {
		t.Fatalf("查事件 attempts: %v", err)
	}
	if err := adminDB.QueryRowContext(ctx,
		`SELECT count(*) FROM audit_logs WHERE company_id = $1`, companyID).Scan(&afterAudits); err != nil {
		t.Fatalf("查稽核: %v", err)
	}
	if afterAttempts != 1 || afterAudits != 1 {
		t.Fatalf("重複認領必須被擋下: attempts=%d(應 1) audit=%d(應 1)", afterAttempts, afterAudits)
	}
}

// staleEvents 只回一次指定事件:模擬「讀取之後、認領之前被別的執行搶先」的競態
// (真實的 store 不會回已派送的事件,故這條路徑只能在這裡人造)。actor 直接給定,不必查 DB。
type staleEvents struct {
	events []store.Event
	actor  int
	served bool
}

func (s *staleEvents) UndispatchedEvents(context.Context, int) ([]store.Event, error) {
	if s.served {
		return nil, nil
	}
	s.served = true
	return s.events, nil
}

func (s *staleEvents) SystemActor(context.Context) (int64, error) { return int64(s.actor), nil }

// companyStatus 讀目前狀態(每次都以真值重查,不靠上一段的區域變數)。
func companyStatus(t *testing.T, ctx context.Context, db *sql.DB, companyID int) string {
	t.Helper()
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM companies WHERE id = $1`, companyID).
		Scan(&status); err != nil {
		t.Fatalf("查公司狀態: %v", err)
	}
	return status
}

// seedTenant 建立一名租戶與其系統 actor 使用者,回傳 (companyID, actorID)。
// companies／users 是 FORCE RLS 的業務表(00028),故走 dbtenant.NewClient ＋ SystemScopeTx
// (scope=all)—— 與生產的排程同一條路;少了這一層會以 42501 失敗。
func seedTenant(t *testing.T, ctx context.Context, adminDB *sql.DB) (companyID, actorID int) {
	t.Helper()
	admin := dbtenant.NewClient(adminDB)
	err := dbtenant.SystemScopeTx(ctx, admin, func(tx *ent.Tx) error {
		co, err := tx.Client().Company.Create().
			SetName("期末凍結測試").
			SetIdentifier("T6-EXPIRED").
			SetStatus(company.StatusActive).
			Save(ctx)
		if err != nil {
			return err
		}
		companyID = co.ID
		u, err := tx.Client().User.Create().
			SetEmail("t6-system@example.invalid").
			SetName("系統排程").
			SetRole("super").
			SetStatus("active").
			SetPasswordHash("!").
			SetCompanyID(co.ID).
			Save(ctx)
		if err != nil {
			return err
		}
		actorID = u.ID
		return nil
	})
	if err != nil {
		t.Fatalf("建立租戶與系統 actor: %v", err)
	}
	return companyID, actorID
}
