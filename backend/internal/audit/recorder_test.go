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
