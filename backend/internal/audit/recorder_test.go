package audit_test

import (
	"context"
	"strings"
	"testing"

	"encoding/json"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
)

// newAuditDB 建立 enttest sqlite client。
func newAuditDB(t *testing.T) *ent.Client {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// mustCompany 建立測試公司。
func mustCompany(t *testing.T, db *ent.Client) *ent.Company {
	t.Helper()
	c, err := db.Company.Create().
		SetName("測試公司").
		SetIdentifier("audit-test-co").Save(context.Background())
	if err != nil {
		t.Fatalf("建立 company: %v", err)
	}
	return c
}

// mustJSON 將任意值序列化為字串（供敏感欄位洩漏檢查）。
func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// mustUser 建立測試使用者（附公司）。
func mustUser(t *testing.T, db *ent.Client, companyID int) *ent.User {
	t.Helper()
	u, err := db.User.Create().
		SetCompanyID(companyID).
		SetEmail("auditor@test.com").
		SetName("稽核員").
		SetPasswordHash("x").
		SetRole("super").
		Save(context.Background())
	if err != nil {
		t.Fatalf("建立 user: %v", err)
	}
	return u
}

func TestRecordWritesSnapshot(t *testing.T) {
	db := newAuditDB(t)
	c := mustCompany(t, db)
	u := mustUser(t, db, c.ID)

	tx, err := db.Tx(context.Background())
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	err = audit.Record(context.Background(), tx, audit.Entry{
		Action:       "create",
		ResourceType: "user",
		ResourceID:   "u-1",
		CompanyID:    c.ID,
		UserID:       u.ID,
		After:        map[string]any{"name": "新員工", "email": "e@t.com"},
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	rows, err := db.AuditLog.Query().All(context.Background())
	if err != nil {
		t.Fatalf("查 audit: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("期望 1 筆 audit,得到 %d", len(rows))
	}
	r := rows[0]
	if r.Action != auditlog.ActionCreate {
		t.Errorf("action = %q,期望 create", r.Action)
	}
	if r.ResourceType != "user" {
		t.Errorf("resource_type = %q", r.ResourceType)
	}
	if r.CompanyID != c.ID || r.UserID != u.ID {
		t.Errorf("company/user 脈絡錯誤: %d/%d", r.CompanyID, r.UserID)
	}
	if r.AfterSnapshot == nil || r.AfterSnapshot["name"] != "新員工" {
		t.Errorf("after_snapshot 缺失: %+v", r.AfterSnapshot)
	}
}

// 未結項 #4（RED）：同一次請求的 trace_id 必須落進 after_snapshot 的 _trace_id ——
// console 合併「一次請求兩筆稽核」的前提。呼叫端不填 TraceID 時由 Record 自 ctx 取。
func TestRecordStampsTraceIDFromContext(t *testing.T) {
	db := newAuditDB(t)
	c := mustCompany(t, db)
	u := mustUser(t, db, c.ID)

	tx, err := db.Tx(context.Background())
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	ctx := requestid.With(context.Background(), "trace-4-abc")
	for _, after := range []map[string]any{
		{"name": "新名"},
		nil, // after 為空時也要能攜帶（login/logout 兩者皆 null 的形狀）
	} {
		if err := audit.Record(ctx, tx, audit.Entry{
			Action: "update", ResourceType: "company", ResourceID: "c-1",
			CompanyID: c.ID, UserID: u.ID, After: after,
		}); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	rows, err := db.AuditLog.Query().All(context.Background())
	if err != nil {
		t.Fatalf("查 audit: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("期望 2 筆 audit,得到 %d", len(rows))
	}
	for _, r := range rows {
		if r.AfterSnapshot["_trace_id"] != "trace-4-abc" {
			t.Fatalf("after_snapshot 應帶 _trace_id，got %+v", r.AfterSnapshot)
		}
	}
}

func TestRecordFiltersSensitiveFields(t *testing.T) {
	db := newAuditDB(t)
	c := mustCompany(t, db)
	u := mustUser(t, db, c.ID)

	tx, err := db.Tx(context.Background())
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 快照含密碼雜湊與 token——不應落入稽核。
	err = audit.Record(context.Background(), tx, audit.Entry{
		Action:       "update",
		ResourceType: "user",
		ResourceID:   "u-2",
		CompanyID:    c.ID,
		UserID:       u.ID,
		Before:       map[string]any{"password_hash": "$argon2id$abc", "status": "active"},
		After:        map[string]any{"password_hash": "$argon2id$xyz", "token": "secret", "status": "inactive"},
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	r, err := db.AuditLog.Query().Order(ent.Desc(auditlog.FieldID)).First(context.Background())
	if err != nil {
		t.Fatalf("查 audit: %v", err)
	}
	raw := r.BeforeSnapshot
	// before 只剩 status
	if r.BeforeSnapshot["password_hash"] != nil {
		t.Errorf("before 洩漏 password_hash: %+v", raw)
	}
	if r.BeforeSnapshot["status"] != "active" {
		t.Errorf("before 缺 status: %+v", raw)
	}
	if r.AfterSnapshot["password_hash"] != nil || r.AfterSnapshot["token"] != nil {
		t.Errorf("after 洩漏敏感欄位: %+v", r.AfterSnapshot)
	}
	if r.AfterSnapshot["status"] != "inactive" {
		t.Errorf("after 缺 status: %+v", r.AfterSnapshot)
	}
	// 序列化檢查：任何欄位不得含 "argon2id"
	blob := strings.Join([]string{
		mustJSON(t, r.BeforeSnapshot),
		mustJSON(t, r.AfterSnapshot),
	}, "|")
	if strings.Contains(blob, "argon2id") {
		t.Errorf("快照洩漏 argon2id hash: %s", blob)
	}
}

// TestRecordStampsActorKindForAPIToken 01 1.6.6:機器代打(server-to-server token)的稽核歸屬標記。
//
// 機器身分沒有自己的 users 列,稽核的 user_id 是 token 綁定的**真實使用者**;若不留標記,
// 稽核調查時會誤判成「這個人親自做的」。故以 _actor_kind 區分。
// 人為操作不寫此鍵(既有快照斷言不受影響)。
func TestRecordStampsActorKindForAPIToken(t *testing.T) {
	db := newAuditDB(t)
	c := mustCompany(t, db)
	u := mustUser(t, db, c.ID)

	write := func(ctx context.Context) {
		t.Helper()
		tx, err := db.Tx(ctx)
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		if err := audit.Record(ctx, tx, audit.Entry{
			Action: "update", ResourceType: "company", ResourceID: "c-1",
			CompanyID: c.ID, UserID: u.ID, After: map[string]any{"x": 1},
		}); err != nil {
			t.Fatalf("Record: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}
	}

	// 機器代打 → 標記 token 名稱。
	write(audit.WithMeta(context.Background(), audit.Meta{ActorKind: "api-token:scheduler"}))
	// 人為操作(meta 無 ActorKind)→ 不寫該鍵。
	write(audit.WithMeta(context.Background(), audit.Meta{IP: "1.2.3.4", UserAgent: "ua"}))

	rows := db.AuditLog.Query().Order(ent.Asc("id")).AllX(context.Background())
	if len(rows) != 2 {
		t.Fatalf("期望 2 筆,got %d", len(rows))
	}
	if got := rows[0].AfterSnapshot["_actor_kind"]; got != "api-token:scheduler" {
		t.Fatalf("機器代打應標 _actor_kind,got %v", got)
	}
	if _, ok := rows[1].AfterSnapshot["_actor_kind"]; ok {
		t.Fatalf("人為操作不應寫 _actor_kind,got %v", rows[1].AfterSnapshot)
	}
}
