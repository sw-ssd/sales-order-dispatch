//go:build integration

package postgres_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationOperators 以真 PostgreSQL 驗證 operatorauth 的存取層:
// 白名單查詢(含「查無此人」不是錯誤)、登入時間、以及登入稽核**真的寫進 platform.audit_logs**。
//
// 這三條查詢是手寫 SQL(platform schema 不用 ent),欄位名寫錯、表名寫錯、jsonb 參數型別沒轉,
// 在記憶體假實作上永遠測不出來;而登入稽核寫不進去等於「誰進過平台工具」無跡可查(S9)。
func TestIntegrationOperators(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.RunContext(t.Context(), "up", db, platformMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	ctx := t.Context()
	var opID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO platform.operators (email, name, role, status)
		VALUES ('ops@example.com', 'Ops', 'admin', 'active') RETURNING id`).Scan(&opID); err != nil {
		t.Fatalf("seed operator: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO platform.operators (email, name, status)
		VALUES ('gone@example.com', 'Gone', 'disabled')`); err != nil {
		t.Fatalf("seed disabled operator: %v", err)
	}

	ops := postgres.NewOperators(db)

	// 白名單命中:每一欄都必須從資料帶出(role/status 漏掃會讓停用者被當成啟用)。
	op, err := ops.OperatorByEmail(ctx, "ops@example.com")
	if err != nil || op == nil {
		t.Fatalf("OperatorByEmail(ops): got %+v err=%v", op, err)
	}
	if op.ID != opID || op.Email != "ops@example.com" || op.Name != "Ops" ||
		op.Role != "admin" || op.Status != "active" {
		t.Fatalf("白名單欄位錯位,got %+v", *op)
	}
	// 停用者必須查得出來(status 由呼叫端判定;查不到就無法區分「已停用」與「不在名單」)。
	if gone, err := ops.OperatorByEmail(ctx, "gone@example.com"); err != nil || gone == nil || gone.Status != "disabled" {
		t.Fatalf("已停用的 operator 仍須查得到且帶出 status,got %+v err=%v", gone, err)
	}
	// 查無此人 = (nil, nil):回錯誤會讓「陌生人嘗試登入」被當成系統故障。
	if nobody, err := ops.OperatorByEmail(ctx, "nobody@example.com"); err != nil || nobody != nil {
		t.Fatalf("查無此人應回 (nil, nil),got %+v err=%v", nobody, err)
	}

	// 登入時間。
	loginAt := time.Now().Add(-time.Minute).Truncate(time.Millisecond)
	if err := ops.TouchOperatorLogin(ctx, opID, loginAt); err != nil {
		t.Fatalf("TouchOperatorLogin: %v", err)
	}
	var gotAt time.Time
	if err := db.QueryRowContext(ctx,
		`SELECT last_login_at FROM platform.operators WHERE id = $1`, opID).Scan(&gotAt); err != nil {
		t.Fatalf("讀回 last_login_at: %v", err)
	}
	if !gotAt.Equal(loginAt) {
		t.Fatalf("last_login_at 應為 %v,got %v", loginAt, gotAt)
	}

	// 登入稽核:actor 為 operator_id(S9,不 FK 租戶 users),來源 IP 與 User-Agent 必須留痕。
	if err := ops.AuditOperatorLogin(ctx, opID, "ops@example.com", "203.0.113.7", "console-test"); err != nil {
		t.Fatalf("AuditOperatorLogin: %v", err)
	}
	var (
		action, targetType, targetID, ip, email, ua string
		auditCount                                  int
	)
	if err := db.QueryRowContext(ctx, `
		SELECT action, target_type, target_id, ip_address, after->>'email', after->>'user_agent'
		  FROM platform.audit_logs WHERE operator_id = $1`, opID).
		Scan(&action, &targetType, &targetID, &ip, &email, &ua); err != nil {
		t.Fatalf("登入稽核必須寫進 platform.audit_logs: %v", err)
	}
	if action != "login" || targetType != "operator" || targetID != strconv.FormatInt(opID, 10) ||
		ip != "203.0.113.7" || email != "ops@example.com" || ua != "console-test" {
		t.Fatalf("稽核內容不對: action=%q target_type=%q target_id=%q ip=%q email=%q ua=%q",
			action, targetType, targetID, ip, email, ua)
	}
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.audit_logs WHERE action = 'login'`).Scan(&auditCount); err != nil {
		t.Fatalf("count 稽核: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("登入稽核應恰為 1 筆,got %d", auditCount)
	}
}
